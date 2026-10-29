package main

import (
	"bufio"
	"context"
	"errors"
	"io"
)

// Read SSE lines without materializing lines or frames. Only data fields enter
// staging; comments, id/retry/event and unknown fields are drained. This is the
// existing non-replaying subscription: disconnect remains an explicit failure.
func readProjectedSSE(ctx context.Context, input io.Reader, dir string, consume func([]byte) error) (result error) {
	r := bufio.NewReaderSize(input, protocolChunkSize)
	if b, _ := r.Peek(3); len(b) == 3 && b[0] == 0xef && b[1] == 0xbb && b[2] == 0xbf {
		_, _ = r.Discard(3)
	}
	var frame *protocolFrame
	defer func() {
		if frame != nil {
			result = errors.Join(result, frame.Close())
		}
	}()
	buffer := make([]byte, 0, protocolChunkSize)
	flush := func() error {
		if len(buffer) == 0 {
			return nil
		}
		_, err := frame.Write(buffer)
		buffer = buffer[:0]
		return err
	}
	put := func(b byte) error {
		buffer = append(buffer, b)
		if len(buffer) == cap(buffer) {
			return flush()
		}
		return nil
	}
	var field [5]byte
	fieldSize := 0
	lineEmpty, value, data, skipSpace, hasData, afterCR := true, false, false, false, false, false
	beginData := func() error {
		if hasData {
			if err := put('\n'); err != nil {
				return err
			}
		}
		hasData = true
		if frame == nil {
			frame = &protocolFrame{ctx: ctx, dir: dir}
		}
		return nil
	}
	endLine := func() error {
		if !value && fieldSize == 4 && string(field[:4]) == "data" {
			if err := beginData(); err != nil {
				return err
			}
		}
		if err := flush(); err != nil {
			return err
		}
		if lineEmpty {
			if frame != nil {
				var projected []byte
				var err error
				if frame.size > 0 {
					projected, err = frame.project("opencode", openCodeFrameLimit)
				}
				err = errors.Join(err, frame.Close())
				frame = nil
				if err != nil {
					return err
				}
				if len(projected) > 0 {
					if err := consume(projected); err != nil {
						return err
					}
				}
			}
			hasData = false
		}
		fieldSize = 0
		lineEmpty, value, data, skipSpace = true, false, false, false
		return nil
	}
	var count uint64
	for {
		count++
		if count%protocolChunkSize == 0 {
			if err := ctx.Err(); err != nil {
				return err
			}
		}
		b, err := r.ReadByte()
		if err == io.EOF {
			return nil
		} // An unterminated SSE frame is never published.
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return errors.New("Could not read the OpenCode event stream.")
		}
		if afterCR {
			afterCR = false
			if b == '\n' {
				continue
			}
		}
		if b == '\r' || b == '\n' {
			if err := endLine(); err != nil {
				return err
			}
			afterCR = b == '\r'
			continue
		}
		lineEmpty = false
		if !value {
			if b == ':' {
				value, skipSpace = true, true
				data = fieldSize == 4 && string(field[:4]) == "data"
				if data {
					if err := beginData(); err != nil {
						return err
					}
				}
			} else if fieldSize < len(field) {
				field[fieldSize] = b
				fieldSize++
			}
			continue
		}
		if skipSpace {
			skipSpace = false
			if b == ' ' {
				continue
			}
		}
		if data {
			if err := put(b); err != nil {
				return err
			}
		}
	}
}

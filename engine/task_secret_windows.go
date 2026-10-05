//go:build windows

package main

import (
	"encoding/base64"
	"errors"
	"unsafe"

	"golang.org/x/sys/windows"
)

// DPAPI binds webhook secrets to the engine user's Windows account. Only the
// encrypted blob goes into checkpoints/journals; projections never include it.
func protectTaskSecret(secret string) (string, error) {
	in := []byte(secret)
	input := windows.DataBlob{Size: uint32(len(in)), Data: &in[0]}
	var output windows.DataBlob
	if err := windows.CryptProtectData(&input, nil, nil, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &output); err != nil {
		return "", errors.New("Cannot protect webhook secret for the engine user.")
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(output.Data)))
	return base64.StdEncoding.EncodeToString(unsafe.Slice(output.Data, int(output.Size))), nil
}

func unprotectTaskSecret(blob string) (string, error) {
	in, err := base64.StdEncoding.DecodeString(blob)
	if err != nil || len(in) == 0 {
		return "", errors.New("Invalid protected webhook secret.")
	}
	input := windows.DataBlob{Size: uint32(len(in)), Data: &in[0]}
	var output windows.DataBlob
	if err := windows.CryptUnprotectData(&input, nil, nil, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &output); err != nil {
		return "", errors.New("Cannot unlock webhook secret as the engine user.")
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(output.Data)))
	return string(unsafe.Slice(output.Data, int(output.Size))), nil
}

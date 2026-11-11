//go:build linux

package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

const mcpEnvelope = "KLM-MCP-AESGCM:1:"
const mcpKeyFile = "credentials.key"

// Open relative to verified directory descriptors. Never follow/replace a key
// link, repair an existing key's permissions, or regenerate a dependent key.
func linuxMCPKey(storage mcpCredentialStorage, create bool) ([]byte, error) {
	if !filepath.IsAbs(storage.Directory) {
		return nil, errors.New("MCP credential storage requires the engine data directory.")
	}
	root, err := unix.Open(storage.Directory, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, errors.New("Cannot open MCP engine storage securely.")
	}
	defer unix.Close(root)
	var stat unix.Stat_t
	if unix.Fstat(root, &stat) != nil || stat.Uid != uint32(os.Geteuid()) || stat.Mode&0022 != 0 {
		return nil, errors.New("MCP engine data directory has unsafe ownership or permissions.")
	}
	err = unix.Mkdirat(root, "mcp-secrets", 0700)
	if err != nil && !errors.Is(err, unix.EEXIST) {
		return nil, errors.New("Cannot create private MCP key directory.")
	}
	if err == nil {
		if err := unix.Fsync(root); err != nil {
			return nil, errors.New("Cannot sync private MCP key directory creation.")
		}
	}
	dir, err := unix.Openat(root, "mcp-secrets", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, errors.New("MCP key directory must not be a link.")
	}
	defer unix.Close(dir)
	if unix.Fstat(dir, &stat) != nil || stat.Uid != uint32(os.Geteuid()) || stat.Mode&07777 != 0700 {
		return nil, errors.New("MCP key directory requires engine-user ownership and permissions 0700.")
	}
	read := func() ([]byte, error) {
		fd, err := unix.Openat(dir, mcpKeyFile, unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
		if err != nil {
			return nil, err
		}
		f := os.NewFile(uintptr(fd), mcpKeyFile)
		defer f.Close()
		var s unix.Stat_t
		if unix.Fstat(fd, &s) != nil || s.Mode&unix.S_IFMT != unix.S_IFREG || s.Uid != uint32(os.Geteuid()) || s.Mode&07777 != 0600 || s.Nlink != 1 {
			return nil, errors.New("MCP key requires a regular single-link engine-user file with permissions 0600.")
		}
		key, err := io.ReadAll(io.LimitReader(f, 33))
		if err != nil || len(key) != 32 {
			return nil, errors.New("MCP key format is invalid; restore the original key from backup.")
		}
		return key, nil
	}
	key, err := read()
	if !errors.Is(err, unix.ENOENT) {
		return key, err
	}
	if !create || storage.Existing {
		return nil, errors.New("MCP credential key is missing. Restore the matching database/key backup; no key was regenerated.")
	}
	key = make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	name := ".key-" + newID()
	fd, err := unix.Openat(dir, name, unix.O_CREAT|unix.O_EXCL|unix.O_WRONLY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0600)
	if err != nil {
		return nil, errors.New("Cannot create private MCP key.")
	}
	defer unix.Unlinkat(dir, name, 0)
	f := os.NewFile(uintptr(fd), name)
	_, writeErr := f.Write(key)
	err = errors.Join(writeErr, f.Sync(), f.Close())
	if err != nil {
		return nil, errors.New("Cannot persist MCP key.")
	}
	// Exclusive atomic rename: readers see either no key or all 32 bytes, with
	// link count one. Concurrent creators always read the winner's key.
	if err := unix.Renameat2(dir, name, dir, mcpKeyFile, unix.RENAME_NOREPLACE); err != nil && !errors.Is(err, unix.EEXIST) {
		return nil, errors.New("Cannot publish private MCP key atomically.")
	}
	if err := unix.Fsync(dir); err != nil {
		return nil, errors.New("Cannot sync private MCP key directory.")
	}
	return read()
}

func linuxMCPCipher(storage []mcpCredentialStorage, create bool) (cipher.AEAD, error) {
	if len(storage) != 1 {
		return nil, errors.New("MCP key storage context is missing.")
	}
	key, err := linuxMCPKey(storage[0], create)
	if err != nil {
		return nil, err
	}
	defer clear(key)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func protectMCPCredential(data []byte, storage ...mcpCredentialStorage) ([]byte, error) {
	aead, err := linuxMCPCipher(storage, true)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	prefix := append([]byte(mcpEnvelope), nonce...)
	return aead.Seal(prefix, nonce, data, []byte("klm/mcp/credentials/v1")), nil
}

func unprotectMCPCredential(data []byte, storage ...mcpCredentialStorage) ([]byte, error) {
	if !bytes.HasPrefix(data, []byte(mcpEnvelope)) {
		return nil, errors.New("Unsupported MCP credential envelope (Windows DPAPI is not portable). Authenticate again on this engine.")
	}
	aead, err := linuxMCPCipher(storage, false)
	if err != nil {
		return nil, err
	}
	sealed := data[len(mcpEnvelope):]
	if len(sealed) < aead.NonceSize()+aead.Overhead() {
		return nil, errors.New("Invalid MCP credential envelope.")
	}
	plain, err := aead.Open(nil, sealed[:aead.NonceSize()], sealed[aead.NonceSize():], []byte("klm/mcp/credentials/v1"))
	if err != nil {
		return nil, errors.New("MCP credential integrity check failed. Restore the matching key; credentials were not changed.")
	}
	return plain, nil
}

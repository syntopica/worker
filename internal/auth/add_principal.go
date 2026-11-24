package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

const (
	tokenBytes  = 32
	privateFile = 0o600
	privateDir  = 0o700
)

// AddPrincipal creates or replaces a principal and returns its plain token.
// Only the token's SHA-256 is stored in principals.json; the token itself is
// written once to tokens/<name>.token, owner-only.
func AddPrincipal(dir, kind, name string) (string, error) {
	if !regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,63}$`).MatchString(name) {
		return "", ErrInvalidName
	}
	raw := make([]byte, tokenBytes)
	_, _ = rand.Read(raw) // never fails: it aborts the process instead
	token := base64.RawURLEncoding.EncodeToString(raw)
	principals, err := loadPrincipals(dir)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(token))
	principals[name] = entry{Kind: kind, SHA256: hex.EncodeToString(sum[:])}
	listing, err := json.MarshalIndent(principals, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encode principals: %w", err)
	}
	tokens := filepath.Join(dir, "tokens")
	err = os.MkdirAll(tokens, privateDir)
	if err == nil {
		err = os.Chmod(tokens, privateDir)
	}
	if err == nil {
		err = writePrivateFile(filepath.Join(tokens, name+".token"), []byte(token))
	}
	if err == nil {
		err = writePrivateFile(filepath.Join(dir, "principals.json"), listing)
	}
	if err != nil {
		return "", err
	}
	return token, nil
}

// writePrivateFile writes a sibling temporary file that is 0600 from its
// first byte and renames it into place.
func writePrivateFile(path string, data []byte) error {
	temporary := path + ".tmp"
	err := os.Remove(temporary)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("replace %s: %w", filepath.Base(path), err)
	}

	file, err := os.OpenFile(temporary, os.O_CREATE|os.O_WRONLY|os.O_EXCL, privateFile)
	if err != nil {
		return fmt.Errorf("create %s: %w", filepath.Base(path), err)
	}
	_, err = file.Write(data)
	err = errors.Join(err, file.Close())
	if err == nil {
		err = os.Rename(temporary, path)
	}
	if err != nil {
		return fmt.Errorf("write %s: %w", filepath.Base(path), err)
	}
	return nil
}

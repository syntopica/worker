package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// loadPrincipals reads the token hashes by principal name; empty when none
// exist yet.
func loadPrincipals(dir string) (map[string]entry, error) {
	principals := map[string]entry{}

	data, err := os.ReadFile(filepath.Join(dir, "principals.json"))
	if errors.Is(err, fs.ErrNotExist) {
		return principals, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read principals: %w", err)
	}
	err = json.Unmarshal(data, &principals)
	if err != nil {
		return nil, fmt.Errorf("decode principals: %w", err)
	}
	return principals, nil
}

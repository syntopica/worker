package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var tokenName = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,63}$`)

func readToken(state, name string) (string, error) {
	if !tokenName.MatchString(name) {
		return "", userError("invalid principal name")
	}
	token, err := os.ReadFile(filepath.Join(state, "tokens", name+".token"))
	if errors.Is(err, os.ErrNotExist) {
		return "", userError(fmt.Sprintf("no token for '%s' (run: worker token add --name %s)", name, name))
	}
	if err != nil {
		return "", userError("token file cannot be read")
	}
	return strings.TrimSpace(string(token)), nil
}

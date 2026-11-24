package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/syntopica/worker/internal/fault"
)

// Authenticate resolves an Authorization header to a principal, comparing
// hashes in constant time; it refuses with 401 when nothing matches.
func Authenticate(dir, header string) (Principal, error) {
	token, found := strings.CutPrefix(header, "Bearer ")
	if !found {
		return Principal{}, fault.New(http.StatusUnauthorized, "unauthorized")
	}
	principals, err := loadPrincipals(dir)
	if err != nil {
		return Principal{}, err
	}
	sum := sha256.Sum256([]byte(token))
	digest := hex.EncodeToString(sum[:])
	for name, known := range principals {
		if subtle.ConstantTimeCompare([]byte(digest), []byte(known.SHA256)) == 1 {
			return Principal{Kind: known.Kind, Name: name}, nil
		}
	}
	return Principal{}, fault.New(http.StatusUnauthorized, "unauthorized")
}

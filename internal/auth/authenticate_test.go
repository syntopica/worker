package auth_test

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/syntopica/worker/internal/auth"
	"github.com/syntopica/worker/internal/fault"
)

func TestAnythingButAKnownBearerTokenIsUnauthorized(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	listing := `{"pa": {"kind": "producer", "sha256": "` +
		`28d8d1bc0dfa4d2bb1d0cc7a4a3cb3e1a5e5a3f6bfc9a3c5cc5f7a8f2c9d7c11"}}`

	err := os.WriteFile(filepath.Join(dir, "principals.json"), []byte(listing), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	for _, header := range []string{"", "Bearer ", "Bearer wrong", "Basic abc", "bearer x"} {
		_, refused := auth.Authenticate(dir, header)

		refusal, ok := errors.AsType[*fault.Error](refused)
		if !ok || refusal.Status != http.StatusUnauthorized || refusal.Code != "unauthorized" {
			t.Fatalf("%q: err = %v", header, refused)
		}
	}

	_, err = auth.Authenticate(filepath.Join(dir, "missing"), "Bearer x")
	if _, ok := errors.AsType[*fault.Error](err); !ok {
		t.Fatalf("no listing: err = %v", err)
	}
}

func TestAListingWrittenByThePythonEngineAuthenticates(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// hashlib.sha256(b"abc").hexdigest()
	listing := `{"node-a": {"kind": "node", "sha256": "` +
		`ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"}}`

	err := os.WriteFile(filepath.Join(dir, "principals.json"), []byte(listing), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	who, err := auth.Authenticate(dir, "Bearer abc")
	if err != nil || who != (auth.Principal{Kind: "node", Name: "node-a"}) {
		t.Fatalf("principal = %+v, %v", who, err)
	}
}

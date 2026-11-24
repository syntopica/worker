package auth_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/syntopica/worker/internal/auth"
)

func TestATokenAuthenticatesItsPrincipalAndOnlyItsHashIsListed(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "state")

	token, err := auth.AddPrincipal(dir, "producer", "pa")
	if err != nil {
		t.Fatal(err)
	}

	other, err := auth.AddPrincipal(dir, "node", "node-a")
	if err != nil || other == token || len(token) != 43 {
		t.Fatalf("tokens = %d %d bytes, %v", len(token), len(other), err)
	}

	who, err := auth.Authenticate(dir, "Bearer "+token)
	if err != nil || who != (auth.Principal{Kind: "producer", Name: "pa"}) {
		t.Fatalf("principal = %+v, %v", who, err)
	}

	listing, err := os.ReadFile(filepath.Join(dir, "principals.json"))
	if err != nil || strings.Contains(string(listing), token) || !strings.Contains(string(listing), `"sha256"`) {
		t.Fatalf("listing leaks or is missing: %v", err)
	}
}

func TestTheTokenIsStoredOnceInOwnerOnlyFiles(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "state")

	token, err := auth.AddPrincipal(dir, "producer", "pa")
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{"principals.json", "tokens/pa.token", "tokens"} {
		info, statErr := os.Stat(filepath.Join(dir, path))
		if statErr != nil || info.Mode().Perm()&0o077 != 0 {
			t.Fatalf("%s mode = %v, %v", path, info.Mode(), statErr)
		}
	}

	stored, err := os.ReadFile(filepath.Join(dir, "tokens", "pa.token"))
	if err != nil || string(stored) != token {
		t.Fatalf("stored token differs: %v", err)
	}
}

func TestAReplacedPrincipalInvalidatesItsOldToken(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	old, err := auth.AddPrincipal(dir, "admin", "ops")
	if err != nil {
		t.Fatal(err)
	}

	_, err = auth.AddPrincipal(dir, "admin", "ops")
	if err != nil {
		t.Fatal(err)
	}

	_, err = auth.Authenticate(dir, "Bearer "+old)
	if err == nil {
		t.Fatal("the replaced token still authenticates")
	}
}

func TestUnsafeNamesAreRefused(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	for _, name := range []string{"", "../x", "a/b", "A", ".hidden", "-dash", strings.Repeat("a", 65), "a b", "a\n"} {
		_, err := auth.AddPrincipal(dir, "producer", name)
		if !errors.Is(err, auth.ErrInvalidName) {
			t.Fatalf("%q: err = %v", name, err)
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 0 {
		t.Fatalf("a refused name left files: %v, %v", entries, err)
	}
}

func TestACorruptListingStopsACreation(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	err := os.WriteFile(filepath.Join(dir, "principals.json"), []byte("{"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	_, err = auth.AddPrincipal(dir, "producer", "pa")
	if err == nil {
		t.Fatal("a corrupt listing was overwritten")
	}

	_, err = auth.Authenticate(dir, "Bearer x")
	if err == nil {
		t.Fatal("a corrupt listing authenticated")
	}
}

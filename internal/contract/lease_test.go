package contract_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/syntopica/worker/internal/contract"
)

// roundTrip decodes a fixture into shape, refusing unknown fields, and checks
// that encoding it again gives the same JSON document.
func roundTrip(t *testing.T, name string, shape any) {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("..", "..", "contract", "fixtures", name))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()

	if err = decoder.Decode(shape); err != nil {
		t.Fatalf("decode %s: %v", name, err)
	}

	encoded, err := json.Marshal(shape)
	if err != nil {
		t.Fatalf("encode %s: %v", name, err)
	}

	var want, got any
	if err = json.Unmarshal(raw, &want); err != nil {
		t.Fatalf("parse fixture: %v", err)
	}

	if err = json.Unmarshal(encoded, &got); err != nil {
		t.Fatalf("parse encoding: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s round trip\n got %s\nwant %s", name, encoded, raw)
	}
}

func TestLeaseMatchesFixture(t *testing.T) {
	t.Parallel()

	var lease contract.Lease

	roundTrip(t, "lease.json", &lease)

	if lease.AttemptID == "" || lease.Generation != 1 {
		t.Errorf("lease fence = %q/%d, want an attempt id and generation 1", lease.AttemptID, lease.Generation)
	}
}

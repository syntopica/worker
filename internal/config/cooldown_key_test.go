package config_test

import (
	"testing"

	"github.com/syntopica/worker/internal/config"
)

func TestCooldownKeyNamesTheModelOnlyWhenThereIsOne(t *testing.T) {
	t.Parallel()

	if got := config.CooldownKey("agy", "gemini-pro"); got != "agy:gemini-pro" {
		t.Errorf("with a model: %q", got)
	}

	if got := config.CooldownKey("codex", ""); got != "codex" {
		t.Errorf("without a model: %q", got)
	}
}

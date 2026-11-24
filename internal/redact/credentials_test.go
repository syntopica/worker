package redact_test

import (
	"strings"
	"testing"

	"github.com/syntopica/worker/internal/redact"
)

// The fixtures are built at run time so no secret scanner reads one here.
func TestCredentialsMasksTheSecretAndKeepsItsLabel(t *testing.T) {
	t.Parallel()

	word, dashes := "pass"+"word", strings.Repeat("-", 5)
	cases := [][2]string{
		{"Your code is 482913", "Your code is [REDACTED]"},
		{word + ": hunter22 ok", word + ": [REDACTED] ok"},
		{"Authorization: Bearer " + strings.Repeat("ab", 8), "Authorization: Bearer [REDACTED]"},
		{"https://user:" + "s3cret" + "@host/path", "https://user:[REDACTED]@host/path"},
		{"https://host/reset?token=" + "abcdef123", "https://host/reset?token=[REDACTED]"},
		{"key " + "gh" + "p_" + strings.Repeat("a", 24) + " end", "key [REDACTED] end"},
		{"a " + dashes + "BEGIN RSA PRIVATE KEY" + dashes + "\nabc\n" + dashes + "END RSA PRIVATE KEY" + dashes + " b", "a [REDACTED] b"},
		{"nothing to hide here", "nothing to hide here"},
	}

	for _, pair := range cases {
		if got := redact.Credentials(pair[0]); got != pair[1] {
			t.Errorf("Credentials(%q) = %q, want %q", pair[0], got, pair[1])
		}
	}
}

func TestCredentialsMatchesPythonUnicodeSemantics(t *testing.T) {
	t.Parallel()

	vendorKey := "gh" + "p_" + strings.Repeat("a", 24)
	fixedVendorPrefix := "AI" + "za" + strings.Repeat("a", 34)
	bearerValue := strings.Repeat("ab", 8)
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"decimal digits in multiple scripts", "Tu código: ١２३4. Listo.", "Tu código: [REDACTED]. Listo."},
		{"Spanish verification label", "verificación: １２３４", "verificación: [REDACTED]"},
		{"subsequent code", "OTP: １２３４/OTP: 4321", "OTP: [REDACTED]/OTP: [REDACTED]"},
		{"code after longer Unicode number", "code: 1234５6789; code: 5678", "code: 1234５6789; code: [REDACTED]"},
		{"Unicode word attached to digits", "code: café1234; code: 5678猫; code: 4321", "code: café1234; code: 5678猫; code: [REDACTED]"},
		{"Unicode word attached to code label", "écode 1234, code猫 5678", "écode 1234, code猫 5678"},
		{"delimiter within code distance", "code" + strings.Repeat("—", 40) + "1234", "code" + strings.Repeat("—", 40) + "[REDACTED]"},
		{"delimiter beyond code distance", "code" + strings.Repeat("—", 41) + "1234", "code" + strings.Repeat("—", 41) + "1234"},
		{"Unicode word before labelled secret", "épassword: example", "épassword: example"},
		{"labelled word separator", "tokenis example; token es example; token\"is example", "tokenis example; token es [REDACTED]; token\"is [REDACTED]"},
		{"Unicode whitespace around labelled secret", "password\u2003:\u00a0example\u0085texto", "password\u2003:\u00a0[REDACTED]\u0085texto"},
		{"Python whitespace in bearer label", "Bearer\x1c" + bearerValue, "Bearer\x1c[REDACTED]"},
		{"Unicode word before bearer label", "界Bearer " + bearerValue, "界Bearer " + bearerValue},
		{"Unicode word before vendor format", "²" + vendorKey + " —" + vendorKey, "²" + vendorKey + " —[REDACTED]"},
		{"adjacent vendor credentials", fixedVendorPrefix + "-" + vendorKey + ".", "[REDACTED][REDACTED]."},
		{"adjacent text with no word boundary", fixedVendorPrefix + "a" + vendorKey + ".", "[REDACTED]" + vendorKey + "."},
		{"Unicode whitespace in userinfo", "https://a\u00a0b:example@host https://a:example\u2028end@host", "https://a\u00a0b:example@host https://a:example\u2028end@host"},
		{"Unicode whitespace after query value", "https://host/?reset=sample\u2029end", "https://host/?reset=[REDACTED]\u2029end"},
		{"Python case folding in bearer value", "Bearer " + strings.Repeat("ab", 7) + "İı", "Bearer [REDACTED]"},
		{"Python case folding in labels", "pın: sample; verİfİcatıon 1234; https://host/?sıg=sample", "pın: [REDACTED]; verİfİcatıon [REDACTED]; https://host/?sıg=[REDACTED]"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got := redact.Credentials(testCase.input); got != testCase.want {
				t.Errorf("Credentials() = %q, want %q", got, testCase.want)
			}
		})
	}
}

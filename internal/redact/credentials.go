package redact

import (
	"regexp"
	"strings"
)

// mark replaces a credential.
const mark = "[REDACTED]"

// The patterns err toward redacting: a postal code after "código" is a
// smaller loss than a one-time code sent to a third party. A pattern's group
// "secret" is replaced, or the whole match when it has none.
var (
	privateKey = regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----[\s\S]*?-----END [A-Z ]*PRIVATE KEY-----`)
	vendorKey  = regexp.MustCompile(`\b(?:sk-(?:ant-|proj-)?[A-Za-z0-9_-]{20,}|gh[pousr]_[A-Za-z0-9]{20,}` +
		`|github_pat_[A-Za-z0-9_]{20,}|glpat-[A-Za-z0-9_-]{20,}|xox[abprs]-[A-Za-z0-9-]{10,}` +
		`|AKIA[0-9A-Z]{16}|AIza[0-9A-Za-z_-]{35}|sk_live_[0-9A-Za-z]{16,}` +
		`|eyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,})`)
	bearer      = regexp.MustCompile(`(?i)\bBearer\s+(?P<secret>[A-Za-z0-9._~+/=-]{16,})`)
	urlPassword = regexp.MustCompile(`(?i)://[^\s/:@]+:(?P<secret>[^\s/@]+)@`)
	labelled    = regexp.MustCompile(`(?i)\b(?:password|passwd|pwd|contrase[ñn]a|clave|passcode|pin|token|secret|` +
		`api[ _-]?key|access[ _-]?key|client[ _-]?secret)\b["']?\s*(?:[:=]|is|es)\s*["']?` +
		`(?P<secret>[^\s"',;<>&]{4,})`)
	oneTimeCode = regexp.MustCompile(`(?i)\b(?:code|c[óo]digo|verification|verificaci[óo]n|otp|2fa|one[- ]time|` +
		`security|seguridad)\b[^\d\n]{0,40}?(?P<secret>\b\d{4,8}\b)`)
	queryParam = regexp.MustCompile(`(?i)[?&](?:token|key|code|sig|signature|auth|access_token|reset|otp)=` +
		`(?P<secret>[^&\s"'<>]{6,})`)
)

// Credentials returns text with every credential pattern's secret replaced
// by a fixed mark (owner rule 2026-10-01).
func Credentials(text string) string {
	for _, pattern := range []*regexp.Regexp{
		privateKey, vendorKey, bearer, urlPassword, labelled, oneTimeCode, queryParam,
	} {
		text = mask(pattern, text)
	}

	return text
}

// mask replaces each match's secret, keeping the label around it so the text
// still reads naturally.
func mask(pattern *regexp.Regexp, text string) string {
	group := max(pattern.SubexpIndex("secret"), 0)

	var out strings.Builder

	last := 0

	for _, match := range pattern.FindAllStringSubmatchIndex(text, -1) {
		out.WriteString(text[last:match[2*group]])
		out.WriteString(mark)

		last = match[2*group+1]
	}

	out.WriteString(text[last:])

	return out.String()
}

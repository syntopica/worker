package redact

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	mark = "[REDACTED]"

	// Python's Unicode word and whitespace classes are wider than Go's Perl classes.
	wordRunes  = `\p{L}\p{N}_`
	wordStart  = `(?P<leadingDelimiter>[^` + wordRunes + `])`
	spaceRunes = `\s\v\x1c-\x1f\x85\p{Z}`
	space      = `[` + spaceRunes + `]`

	// Python's case-insensitive i also matches dotted and dotless I.
	foldI = `[iİı]`
)

// The patterns err toward redacting: a postal code after "código" is a
// smaller loss than a one-time code sent to a third party. A pattern's group
// "secret" is replaced, or the whole match when it has none.
var (
	privateKey = regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----[\s\S]*?-----END [A-Z ]*PRIVATE KEY-----`)
	vendorKey  = regexp.MustCompile(wordStart + `(?P<secret>sk-(?:ant-|proj-)?[A-Za-z0-9_-]{20,}|gh[pousr]_[A-Za-z0-9]{20,}` +
		`|github_pat_[A-Za-z0-9_]{20,}|glpat-[A-Za-z0-9_-]{20,}|xox[abprs]-[A-Za-z0-9-]{10,}` +
		`|AKIA[0-9A-Z]{16}|AIza[0-9A-Za-z_-]{35}|sk_live_[0-9A-Za-z]{16,}` +
		`|eyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,})`)
	bearer      = regexp.MustCompile(`(?i)` + wordStart + `Bearer` + space + `+(?P<secret>[A-Za-zİı0-9._~+/=-]{16,})`)
	urlPassword = regexp.MustCompile(`://[^` + spaceRunes + `/:@]+:(?P<secret>[^` + spaceRunes + `/@]+)@`)
	labelled    = regexp.MustCompile(`(?i)` + wordStart +
		`(?:password|passwd|pwd|contrase[ñn]a|clave|passcode|p` + foldI + `n|token|secret|ap` + foldI +
		`[ _-]?key|access[ _-]?key|cl` + foldI + `ent[ _-]?secret)` +
		`(?:["']?` + space + `*[:=]|(?:["']` + space + `*|` + space + `+)(?:` + foldI + `s|es))` +
		space + `*["']?(?P<secret>[^` + spaceRunes + `"',;<>&]{4,})`)
	// The gap is 1–40 runes, contains no decimal digit or newline, and starts
	// and ends with a nonword rune to bound both the label and the code.
	oneTimeCode = regexp.MustCompile(`(?i)` + wordStart +
		`(?:code|c[óo]d` + foldI + `go|ver` + foldI + `f` + foldI + `ca(?:t` + foldI + `on|c` + foldI + `[óo]n)` +
		`|otp|2fa|one[- ]t` + foldI + `me|secur` + foldI + `ty|segur` + foldI + `dad)` +
		`(?:[^` + wordRunes + `\n][^\p{Nd}\n]{0,38}?)?[^` + wordRunes + `\n]` +
		`(?P<secret>\p{Nd}{4,8})(?P<invalidSuffix>[` + wordRunes + `])?`)
	queryParam = regexp.MustCompile(`(?i)[?&](?:token|key|code|s` + foldI + `g(?:nature)?|auth|access_token|reset|otp)=` +
		`(?P<secret>[^&` + spaceRunes + `"'<>]{6,})`)
)

// Credentials returns text with every credential pattern's secret replaced
// by a fixed mark (owner rule 2026-10-01).
func Credentials(text string) string {
	// Supply the leading delimiter for patterns that start at the first rune.
	text = " " + text
	for _, pattern := range []*regexp.Regexp{
		privateKey, vendorKey, bearer, urlPassword, labelled, oneTimeCode, queryParam,
	} {
		text = mask(pattern, text)
	}

	return text[1:]
}

// mask replaces each match's secret, keeping the label around it so the text
// still reads naturally.
func mask(pattern *regexp.Regexp, text string) string {
	group := max(pattern.SubexpIndex("secret"), 0)
	invalidSuffix := pattern.SubexpIndex("invalidSuffix")
	reuseDelimiter := pattern.SubexpIndex("leadingDelimiter") >= 0

	var out strings.Builder

	last := 0

	for offset := 0; offset < len(text); {
		match := pattern.FindStringSubmatchIndex(text[offset:])
		if match == nil {
			break
		}

		start, end := offset+match[2*group], offset+match[2*group+1]
		offset += match[1]
		if reuseDelimiter {
			// A match's final rune can also be the delimiter for the next match.
			_, width := utf8.DecodeLastRuneInString(text[:offset])
			offset -= width
		}

		// A code with a word character immediately after it has no end boundary.
		if invalidSuffix >= 0 && match[2*invalidSuffix] >= 0 {
			continue
		}

		out.WriteString(text[last:start])
		out.WriteString(mark)

		last = end
	}

	out.WriteString(text[last:])

	return out.String()
}

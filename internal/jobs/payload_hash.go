package jobs

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode/utf16"
)

const (
	asciiBelow  = 0x7f
	controlLast = 0x1f
	bmpLast     = 0xffff
)

// payloadHash is the SHA-256 of a submit body's canonical JSON, for
// idempotency comparisons: sorted keys, no whitespace, ASCII only, numbers in
// the form the first coordinator wrote them, so a key admitted by it is still
// recognised.
func payloadHash(body map[string]any) string {
	var canonical strings.Builder
	writeCanonical(&canonical, body)
	sum := sha256.Sum256([]byte(canonical.String()))

	return hex.EncodeToString(sum[:])
}

func writeCanonical(out *strings.Builder, value any) {
	switch typed := value.(type) {
	case nil:
		out.WriteString("null")
	case bool:
		out.WriteString(strconv.FormatBool(typed))
	case json.Number:
		out.WriteString(canonicalNumber(typed.String()))
	case string:
		writeCanonicalString(out, typed)
	case []any:
		out.WriteByte('[')
		for i, item := range typed {
			if i > 0 {
				out.WriteByte(',')
			}
			writeCanonical(out, item)
		}
		out.WriteByte(']')
	case map[string]any:
		writeCanonicalObject(out, typed)
	}
}

func writeCanonicalObject(out *strings.Builder, fields map[string]any) {
	names := make([]string, 0, len(fields))
	for name := range fields {
		names = append(names, name)
	}
	slices.Sort(names)
	out.WriteByte('{')
	for i, name := range names {
		if i > 0 {
			out.WriteByte(',')
		}
		writeCanonicalString(out, name)
		out.WriteByte(':')
		writeCanonical(out, fields[name])
	}
	out.WriteByte('}')
}

func writeCanonicalString(out *strings.Builder, value string) {
	short := map[rune]string{'"': `\"`, '\\': `\\`, '\n': `\n`, '\r': `\r`, '\t': `\t`, '\b': `\b`, '\f': `\f`}
	out.WriteByte('"')
	for _, char := range value {
		escape, ok := short[char]
		switch {
		case ok:
			out.WriteString(escape)
		case char > controlLast && char < asciiBelow:
			out.WriteRune(char)
		case char > bmpLast:
			high, low := utf16.EncodeRune(char)
			fmt.Fprintf(out, `\u%04x\u%04x`, high, low)
		default:
			fmt.Fprintf(out, `\u%04x`, char)
		}
	}
	out.WriteByte('"')
}

// canonicalNumber keeps an integer literal as written and writes a float in
// its shortest round-trip form.
func canonicalNumber(literal string) string {
	if !strings.ContainsAny(literal, ".eE") {
		if literal == "-0" {
			return "0"
		}

		return literal
	}
	// An out-of-range literal parses to an infinity, which is what is wanted.
	number, _ := strconv.ParseFloat(literal, 64)

	return shortestFloat(number)
}

package jobs

import (
	"strings"
	"unicode"
)

// normalizeSchemaRegexp translates Python named groups while keeping the
// engine's default Unicode and anchor behavior.
func normalizeSchemaRegexp(pattern string) (string, bool) {
	if !strings.Contains(pattern, "(?P") {
		return pattern, true
	}

	scan := schemaRegexpScan{rest: pattern, closed: make(map[string]bool)}
	scan.out.Grow(len(pattern))
	for scan.rest != "" && !scan.invalid {
		scan.advance()
	}

	return scan.out.String(), !scan.invalid
}

type schemaRegexpScan struct {
	rest    string
	out     strings.Builder
	closed  map[string]bool
	scopes  []schemaRegexpScope
	verbose bool
	invalid bool
}

type schemaRegexpScope struct {
	name    string
	verbose bool
}

func (scan *schemaRegexpScan) advance() {
	switch scan.rest[0] {
	case '\\':
		scan.copy(escapeLength(scan.rest))
	case '[':
		scan.copy(classLength(scan.rest))
	case '(':
		scan.openGroup()
	case ')':
		scan.closeGroup()
	case '#':
		scan.lineComment()
	default:
		scan.copy(1)
	}
}

func (scan *schemaRegexpScan) copy(length int) {
	scan.out.WriteString(scan.rest[:length])
	scan.rest = scan.rest[length:]
}

func (scan *schemaRegexpScan) openGroup() {
	if strings.HasPrefix(scan.rest, "(?#") {
		scan.inlineComment()
		return
	}

	if scan.namedGroup() {
		return
	}

	length, verbose := schemaRegexpInlineVerboseOption(scan.rest, scan.verbose)
	if length == 0 {
		scan.pushScope("")
		scan.copy(1)
		return
	}

	if scan.rest[length-1] == ':' {
		scan.pushScope("")
	}

	scan.verbose = verbose
	scan.copy(length)
}

func (scan *schemaRegexpScan) pushScope(name string) {
	scan.scopes = append(scan.scopes, schemaRegexpScope{name: name, verbose: scan.verbose})
}

func (scan *schemaRegexpScan) closeGroup() {
	if last := len(scan.scopes) - 1; last >= 0 {
		scope := scan.scopes[last]
		if scope.name != "" {
			scan.closed[scope.name] = true
		}
		scan.verbose = scope.verbose
		scan.scopes = scan.scopes[:last]
	}

	scan.copy(1)
}

func (scan *schemaRegexpScan) inlineComment() {
	length, ended := commentLength(scan.rest, ')')
	if !ended {
		scan.invalid = true
		return
	}

	scan.out.WriteString("(?#)")
	scan.rest = scan.rest[length:]
}

func (scan *schemaRegexpScan) lineComment() {
	if !scan.verbose {
		scan.copy(1)
		return
	}

	length, ended := commentLength(scan.rest, '\n')
	scan.out.WriteString("#")
	if ended {
		scan.out.WriteString("\n")
	}
	scan.rest = scan.rest[length:]
}

func (scan *schemaRegexpScan) namedGroup() bool {
	const declaration = "(?P<"
	const reference = "(?P="

	if name, rest, ok := cutSchemaGroupName(scan.rest, declaration, '>'); ok {
		if _, exists := scan.closed[name]; exists {
			scan.invalid = true
			return true
		}
		scan.closed[name] = false
		scan.pushScope(name)
		scan.out.WriteString("(?<" + name + ">")
		scan.rest = rest
		return true
	}

	if name, rest, ok := cutSchemaGroupName(scan.rest, reference, ')'); ok {
		if !scan.closed[name] {
			scan.invalid = true
			return true
		}
		scan.out.WriteString(`\k<` + name + ">")
		scan.rest = rest
		return true
	}

	return false
}

func cutSchemaGroupName(input, prefix string, delimiter byte) (name, rest string, ok bool) {
	group, found := strings.CutPrefix(input, prefix)
	if !found {
		return "", input, false
	}

	for index, character := range group {
		if index > 0 && character == rune(delimiter) {
			return group[:index], group[index+1:], true
		}

		if !schemaGroupNameCharacter(character, index == 0) {
			break
		}
	}

	return "", input, false
}

func schemaGroupNameCharacter(character rune, first bool) bool {
	if character == '_' || unicode.In(character, unicode.Letter, unicode.Nl) {
		return true
	}

	return !first && unicode.In(character, unicode.Nd, unicode.Mn, unicode.Mc, unicode.Pc)
}

func classLength(pattern string) int {
	index := 1
	if index < len(pattern) && pattern[index] == '^' {
		index++
	}
	if index < len(pattern) && pattern[index] == ']' {
		index++
	}

	for index < len(pattern) {
		switch pattern[index] {
		case '\\':
			index += escapeLength(pattern[index:])
		case ']':
			return index + 1
		default:
			index++
		}
	}

	return len(pattern)
}

func escapeLength(pattern string) int {
	const quoteStart, quoteEnd = `\Q`, `\E`
	quoted, found := strings.CutPrefix(pattern, quoteStart)
	if !found {
		return min(len(pattern), 2)
	}

	if end := strings.Index(quoted, quoteEnd); end >= 0 {
		return len(quoteStart) + end + len(quoteEnd)
	}

	return len(pattern)
}

func commentLength(pattern string, delimiter byte) (int, bool) {
	for index := 0; index < len(pattern); index++ {
		switch pattern[index] {
		case '\\':
			index++
		case delimiter:
			return index + 1, true
		}
	}

	return len(pattern), false
}

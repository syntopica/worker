package jobs

import "strings"

func schemaRegexpInlineVerboseOption(pattern string, verbose bool) (int, bool) {
	if !strings.HasPrefix(pattern, "(?") {
		return 0, verbose
	}

	enabled := true
	for index := 2; index < len(pattern); index++ {
		switch pattern[index] {
		case 'x', 'X':
			verbose = enabled
		case 'i', 'I', 'm', 'M', 'n', 'N', 's', 'S', 'u', 'U':
		case '-':
			enabled = false
		case '+':
			enabled = true
		case ':', ')':
			return index + 1, verbose
		default:
			return 0, verbose
		}
	}

	return 0, verbose
}

package cli

import (
	"flag"
	"strings"
)

func flagsFirst(flags *flag.FlagSet, arguments []string) []string {
	options := make([]string, 0, len(arguments))
	var positionals []string
	for len(arguments) > 0 {
		arg := arguments[0]
		arguments = arguments[1:]
		if arg == "--" {
			positionals = append(positionals, arguments...)
			break
		}
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			positionals = append(positionals, arg)
			continue
		}
		options = append(options, arg)
		if flagTakesNext(flags, arg) {
			if len(arguments) == 0 {
				return options
			}
			options = append(options, arguments[0])
			arguments = arguments[1:]
		}
	}
	return append(append(options, "--"), positionals...)
}

func flagTakesNext(flags *flag.FlagSet, arg string) bool {
	if strings.Contains(arg, "=") {
		return false
	}
	field := flags.Lookup(strings.TrimLeft(arg, "-"))
	if field == nil {
		return false
	}
	value, ok := field.Value.(interface{ IsBoolFlag() bool })
	return !ok || !value.IsBoolFlag()
}

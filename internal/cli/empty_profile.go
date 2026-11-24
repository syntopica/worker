package cli

import "flag"

func emptyProfile(flags *flag.FlagSet, profile string) bool {
	if profile != "" {
		return false
	}
	set := false
	flags.Visit(func(option *flag.Flag) { set = set || option.Name == "profile" })
	return set
}

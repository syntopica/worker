package cli

import "regexp"

var refusalCode = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)

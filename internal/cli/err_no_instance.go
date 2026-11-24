package cli

import "errors"

// ErrNoInstance reports that no Syntopica instance could be resolved.
var ErrNoInstance = errors.New("no Syntopica instance (set SYNTOPICA_DATA)")

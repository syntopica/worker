package tasks

import "os"

// privateFile is the mode of a file only the worker may read.
const privateFile os.FileMode = 0o600

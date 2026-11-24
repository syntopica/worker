package tasks

import "os"

// privateDir is the mode of a directory only the worker may enter.
const privateDir os.FileMode = 0o700

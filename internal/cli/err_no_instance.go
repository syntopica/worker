package cli

// ErrNoInstance means neither the environment nor the current directory selects an instance.
const ErrNoInstance = userError("no Syntopica instance (set SYNTOPICA_DATA)")

package cli

type executorArgs struct {
	name, profile string
	slots         int
	overridePins  bool
	createdBefore *float64
}

func (a executorArgs) valid(emptyProfile bool) bool {
	return !emptyProfile && a.name != "" && a.slots >= 0
}

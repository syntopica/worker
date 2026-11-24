package jobs

import "encoding/json"

// excludedProviders returns the job's stored exclusion list; only a missing
// value excludes nothing.
func excludedProviders(stored string) ([]string, error) {
	var names []string
	if stored != "" {
		if err := json.Unmarshal([]byte(stored), &names); err != nil || names == nil {
			return nil, errInvalidExclusions
		}
	}
	return names, nil
}

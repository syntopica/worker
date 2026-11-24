package policy

import "github.com/syntopica/worker/internal/config"

// CappedQueues holds the queues that reached their own daily OpenRouter cap
// or their key's, which is compared with every queue's spend together.
func CappedQueues(cfg *config.Config, spent map[string]int) map[string]bool {
	capped := map[string]bool{}
	keySpent := 0
	for _, count := range spent {
		keySpent += count
	}
	for name := range cfg.Queues {
		route := cfg.Queues[name].OpenRouter
		if route == nil {
			continue
		}
		if route.DailyCap != nil && spent[name] >= *route.DailyCap {
			capped[name] = true
		}
		if route.DailyKeyCap != nil && keySpent >= *route.DailyKeyCap {
			capped[name] = true
		}
	}
	return capped
}

package api

const defaultDays = 7

func reportDays(q query) (float64, error) {
	return q.number("days", defaultDays)
}

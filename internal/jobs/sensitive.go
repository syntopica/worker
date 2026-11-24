package jobs

// Sensitive reports whether a privacy class loses its payloads as soon as
// the producer acknowledges the job (spec 9).
func Sensitive(privacy string) bool {
	return privacy == "personal" || privacy == "mail" || privacy == "secret"
}

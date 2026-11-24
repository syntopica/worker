package jobs

// KnownRating reports whether a producer rating is part of the contract.
func KnownRating(rating string) bool {
	return rating == "good" || rating == "edited" || rating == "discarded"
}

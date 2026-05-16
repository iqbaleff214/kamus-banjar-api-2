package domain

type ModerationStats struct {
	PendingContributions int
	FlaggedComments      int
	ApprovedThisWeek     int
	RejectedThisWeek     int
}

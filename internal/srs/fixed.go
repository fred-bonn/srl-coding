package srs

import "time"

// Fixed schedules reviews on a fixed interval ladder, resetting to one day
// after any failure.
type Fixed struct{}

var _ Scheduler = Fixed{}

var fixedLadder = [5]int{1, 3, 7, 14, 30}

func (Fixed) NextDue(reviews []Review, now time.Time) time.Time {
	if len(reviews) == 0 {
		return now
	}
	streak := 0
	for i := len(reviews) - 1; i >= 0 && reviews[i].Correct; i-- {
		streak++
	}
	if streak == 0 {
		return now.Add(24 * time.Hour)
	}
	days := fixedLadder[min(streak, len(fixedLadder))-1]
	return now.Add(time.Duration(days) * 24 * time.Hour)
}

// Package srs computes when a topic is next due for review. Strategies are
// pure: they rebuild their state from the completed review history and
// return the next due date. No algorithm state is persisted.
package srs

import "time"

type Review struct {
	Correct    bool
	ReviewedAt time.Time
}

type Scheduler interface {
	// NextDue returns when the topic is next due. reviews must be in
	// chronological order, oldest first, and contain only completed
	// attempts. A topic with no completed reviews is due immediately.
	NextDue(reviews []Review, now time.Time) time.Time
}

// Package challenge picks challenges for topic reviews: an unseen one
// from the pool, a freshly generated one if the pool is exhausted, or a
// seen one when no generator is configured.
package challenge

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/google/uuid"

	"github.com/fred-bonn/srlcode/internal/challengegen"
	"github.com/fred-bonn/srlcode/internal/database"
)

type Picker struct {
	db  database.Querier
	gen challengegen.Generator
}

func NewPicker(db database.Querier, gen challengegen.Generator) *Picker {
	return &Picker{db: db, gen: gen}
}

// ReviewChallenge returns the ID of a challenge to review topicID with.
func (p *Picker) ReviewChallenge(ctx context.Context, userID, topicID uuid.UUID) (uuid.UUID, error) {
	unseen, err := p.db.ListUnseenChallengesForTopic(ctx, database.ListUnseenChallengesForTopicParams{
		UserID:  userID,
		TopicID: topicID,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("listing unseen challenges: %w", err)
	}
	if len(unseen) > 0 {
		return unseen[rand.Intn(len(unseen))].ID, nil
	}
	if p.gen != nil {
		return p.generate(ctx, topicID)
	}

	// No generator configured: fall back to reusing a seen challenge.
	seen, err := p.db.ListChallengesForTopic(ctx, topicID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("listing challenges: %w", err)
	}
	if len(seen) == 0 {
		return uuid.Nil, fmt.Errorf("no challenges for topic and no generator configured")
	}
	return seen[rand.Intn(len(seen))].ID, nil
}

func (p *Picker) generate(ctx context.Context, topicID uuid.UUID) (uuid.UUID, error) {
	topic, err := p.db.GetTopic(ctx, topicID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("loading topic: %w", err)
	}
	generated, err := p.gen.GenerateChallenge(ctx, topic.Name)
	if err != nil {
		return uuid.Nil, fmt.Errorf("generating challenge: %w", err)
	}
	stored, err := p.db.CreateChallenge(ctx, database.CreateChallengeParams{
		Assignment:   generated.Assignment,
		StartCode:    generated.StartCode,
		TestCode:     generated.TestCode,
		SolutionCode: generated.SolutionCode,
		Difficulty:   int32(generated.Difficulty),
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("storing challenge: %w", err)
	}
	for _, name := range generated.Topics {
		linked, err := p.db.CreateTopic(ctx, name)
		if err != nil {
			return uuid.Nil, fmt.Errorf("creating topic %s: %w", name, err)
		}
		if err := p.db.EnsureChallengeTopic(ctx, database.EnsureChallengeTopicParams{
			ChallengeID: stored.ID,
			TopicID:     linked.ID,
		}); err != nil {
			return uuid.Nil, fmt.Errorf("linking challenge to topic %s: %w", name, err)
		}
	}
	return stored.ID, nil
}

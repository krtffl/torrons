package domain

import "context"

// UserDuelStat is one pairing a user voted on, alongside the crowd's own
// split on that same pairing. It's the shared shape behind both stats
// WrappedStatsRepo.DuelStats derives (most contested, most unpopular
// pick) - see that method's doc for how each is picked out of the same
// underlying dataset.
type UserDuelStat struct {
	TorroAId     string
	TorroAName   string
	TorroBId     string
	TorroBName   string
	UserPickId   string
	UserPickName string
	VotesA       int
	VotesB       int
	TotalVotes   int
}

// WrappedDuelStats holds the two duel-shaped stats featured on a user's
// Wrapped recap. Both are derived from the same per-user-pairing dataset
// (see WrappedStatsRepo.DuelStats).
type WrappedDuelStats struct {
	// HasContestedDuel is false if the user never voted on a pairing that
	// met the minimum-total-votes eligibility threshold - a legitimate
	// empty state, not an error.
	HasContestedDuel bool
	// ContestedDuel is meaningful only if HasContestedDuel.
	ContestedDuel UserDuelStat

	// HasUnpopularPick follows the same eligibility rule as
	// HasContestedDuel above.
	HasUnpopularPick bool
	// UnpopularPick is meaningful only if HasUnpopularPick.
	UnpopularPick UserDuelStat
}

// BracketPathStat summarizes a single user's participation in one Phase 2
// knockout bracket, for the Wrapped recap's "your path to the Gran Final"
// section.
type BracketPathStat struct {
	// HasVoted is false if the user never voted in this bracket at all;
	// every other field is then a zero value.
	HasVoted bool

	RoundsVoted    int
	MatchesDecided int // how many of the user's voted matches already have a WinnerId
	PicksCorrect   int // subset of MatchesDecided where the user's pick == WinnerId

	// HasChampion is whether the bracket itself is fully decided.
	HasChampion  bool
	ChampionName string
	// MatchedChampion is meaningful only if HasChampion: did the user
	// pick the champion in any of their voted matches.
	MatchedChampion bool
}

// WrappedStatsRepo provides the read-only cross-source aggregation behind
// a user's personal "Torrorèndum Wrapped" recap. Like PressStatsRepo, it
// only reads from existing Phase 1 (Results/Pairings/Torrons) and Phase 2
// (BracketMatches/BracketMatchVotes) tables - no new schema is needed.
type WrappedStatsRepo interface {
	// DuelStats returns, for one user, their most contested duel (closest
	// global vote split among pairings they voted on) and their most
	// unpopular pick (the pairing where the user's own pick had the
	// lowest global support %), among pairings with at least
	// minTotalVotes total votes cast (avoids a near-empty pairing reading
	// as maximally contested/unpopular). Both are derived from the SAME
	// underlying per-user-pairing dataset - fetch it once, derive both.
	// Returns a WrappedDuelStats with both Has* flags false (not an
	// error) if the user has no eligible pairings yet.
	DuelStats(ctx context.Context, userId string, minTotalVotes int) (*WrappedDuelStats, error)

	// BracketPath summarizes a user's participation in one bracket.
	// Returns a BracketPathStat with HasVoted false (not an error) if the
	// user never voted in this bracket.
	BracketPath(ctx context.Context, userId string, bracketId string) (*BracketPathStat, error)
}

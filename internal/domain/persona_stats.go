package domain

import "context"

// PersonaStats holds the data behind a user's "torró personality reveal"
// (GET /reveal, design prompt 14): which voting ARENA (Pairings.Class -
// see the doc comment on TopClassId below for why it's the arena and not
// Torrons.Class) they voted in the most, how that compares to everyone
// else who's cleared the same vote threshold, and which specific torró
// they crowned the most within that arena.
//
// Like domain.WrappedStatsRepo's WrappedDuelStats/BracketPathStat, every
// "not enough data yet" case is expressed as a Has* flag plus a zero
// value, never an error.
type PersonaStats struct {
	// HasEnoughVotes is always true when returned by PersonaRepo.Stats:
	// callers are expected to check this threshold themselves first (see
	// PersonaRepo.Stats's doc comment) before ever calling Stats, the same
	// way wrappedCardData checks user.VoteCount before ever calling
	// WrappedStatsRepo. Kept here (rather than dropped) so PersonaStats has
	// the same self-describing "Has* flag + zero value" shape as the rest
	// of this file, for any future caller that doesn't already have this
	// context.
	HasEnoughVotes bool
	// VotesRemaining is meaningful only when !HasEnoughVotes, which
	// PersonaRepo.Stats itself never produces (see above) - always 0 here.
	VotesRemaining int
	// TotalVotes is the user's all-time vote count (domain.User.VoteCount).
	TotalVotes int

	// HasClearFavorite is false when the user's ClassVotes has no unique
	// arg-max arena (an exact tie between two or more arenas) - the "ELS
	// EQUILIBRATS" persona (design prompt 14's approved copy table).
	// TopClassId/TopClassName/Percentile are only meaningful when true.
	HasClearFavorite bool
	// TopClassId is the Pairings.Class (1-5) the user voted in the most,
	// or "" when !HasClearFavorite. This is deliberately the voting ARENA,
	// not the winning torró's own Torrons.Class: individual Torrons rows
	// are never tagged Class="5" (Global), yet "Global" ("ELS ÀRBITRES") is
	// a real persona in the approved copy table, so only Pairings.Class
	// (which Global pairings do carry) can ever produce it.
	TopClassId   string
	TopClassName string

	// Percentile is what % (1-100) of other users who've also cleared the
	// same reveal-unlock threshold share this user's TopClassId. Only
	// meaningful when HasClearFavorite.
	Percentile int

	// TopTorroId/TopTorroName is the torró the user voted for (won) most
	// often within TopClassId's arena (or, when !HasClearFavorite, across
	// every arena - see PersonaRepo.Stats).
	TopTorroId   string
	TopTorroName string
	// TopTorroTag is always left empty by PersonaRepo: the per-class tag
	// copy ("PER ALS GOLAFRES", etc.) is presentation-layer copy owned by
	// the HTTP handler (see arenaTagForClass in
	// internal/http/reveal_handler.go), the same way wrappedCardData
	// computes vote percentages downstream of WrappedStatsRepo rather than
	// inside it.
	TopTorroTag string

	// TopTorroVotesCast is how many votes the user cast within that same
	// arena (every Results row in it, regardless of who won).
	// TopTorroWins/TopTorroLosses are narrower: among that arena's
	// Results, only the pairings where TopTorroId itself was one of the
	// two contenders.
	TopTorroVotesCast int
	TopTorroWins      int
	TopTorroLosses    int
}

// PersonaRepo provides the read-only aggregation behind a user's "torró
// personality reveal". Like WrappedStatsRepo/PressStatsRepo, it only reads
// from existing Users/Results/Pairings/Torrons/Classes tables - no new
// schema is needed.
type PersonaRepo interface {
	// Stats computes one user's PersonaStats. Callers must only invoke
	// this once they've already confirmed the user has cleared the reveal
	// unlock threshold (getMinVotesForClass(embedDefaultClassId) in
	// internal/http/user_api.go) themselves - mirroring exactly how
	// wrapped_handler.go checks user.VoteCount before ever calling
	// WrappedStatsRepo. minVotes is that same threshold, threaded through
	// again to scope the percentile cohort query to "every other user who
	// has ALSO cleared it" (see PersonaStats.Percentile's doc comment):
	// since the current user themselves is always a member of that
	// cohort, the percentile's denominator can never be 0.
	Stats(ctx context.Context, userId string, minVotes int) (*PersonaStats, error)
}

// TopClassIdFromVotes returns the arg-max key of a ClassVotesMap-shaped
// vote-count map (see domain.ClassVotesMap), and false if the map is empty
// or the maximum count is shared by more than one key (an exact tie - the
// "ELS EQUILIBRATS" persona has no single top class). Iteration order over
// a Go map is randomized, so ties are detected by tracking how many keys
// currently hold the running maximum, not by first-seen-wins - see the
// tests in persona_stats_test.go for the tie-then-new-max ordering this
// guards against.
func TopClassIdFromVotes(votes map[string]int) (string, bool) {
	topId := ""
	topCount := -1
	tiedAtTop := 0

	for id, count := range votes {
		switch {
		case count > topCount:
			topId = id
			topCount = count
			tiedAtTop = 1
		case count == topCount:
			tiedAtTop++
		}
	}

	if tiedAtTop != 1 {
		return "", false
	}
	return topId, true
}

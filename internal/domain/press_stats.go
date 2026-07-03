package domain

import "context"

// TorroStat is a single torró paired with a numeric value whose meaning
// depends on which PressStatsRepo method produced it (a total vote count,
// or a net ELO rating change).
type TorroStat struct {
	TorroId string
	Name    string
	Image   string
	Value   float64 // meaning depends on which method returned it (win count, or net rating change)
}

// ClosestDuel is the specific pairing (two torrons) with the smallest gap
// between their vote tallies against each other.
type ClosestDuel struct {
	TorroA     TorroStat // Value = vote count for this torró in the pairing
	TorroB     TorroStat // Value = vote count for this torró in the pairing
	TotalVotes int
}

// PressStatsRepo provides read-only aggregate statistics computed over the
// existing voting history (Results/Pairings/Torrons), for the public
// /premsa press page. It is deliberately not named StatsRepo: that name is
// already owned by stats_handler.go's unrelated personal-stats page.
type PressStatsRepo interface {
	// MostVotedTorro returns the torró with the most total votes cast in
	// its favor across all Results ever recorded. Returns (nil, nil) if
	// no votes have ever been cast (not an error - a legitimate empty state).
	MostVotedTorro(ctx context.Context) (*TorroStat, error)

	// BiggestRiser returns the torró with the largest net ELO rating
	// change summed over its Results in the last windowDays days. Returns
	// (nil, nil) if no votes fall in that window.
	BiggestRiser(ctx context.Context, windowDays int) (*TorroStat, error)

	// ClosestDuel returns the specific pairing (across all of history) with
	// the smallest absolute vote-count gap between its two torrons, among
	// pairings with at least minTotalVotes total votes cast (avoids a 1-0
	// "duel" reading as maximally close). Returns (nil, nil) if no pairing
	// meets the threshold yet.
	ClosestDuel(ctx context.Context, minTotalVotes int) (*ClosestDuel, error)

	// VotesForTorro returns the total number of times this specific torró
	// has been chosen as the winner across all Results ever recorded.
	// Used by the press-kit card to report "segons X vots" for a known
	// champion torró (as opposed to MostVotedTorro, which finds whichever
	// torró has the most wins, not necessarily this one).
	VotesForTorro(ctx context.Context, torroId string) (int, error)
}

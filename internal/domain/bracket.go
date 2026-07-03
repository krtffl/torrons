package domain

import (
	"context"
	"database/sql"
)

// Phase 2 - The knockout.
//
// This is a deliberately different mechanic from Phase 1's open season
// (see pairing.go / result.go): a bracket match has a single deterministic
// winner decided by vote tally within a round, each user votes at most once
// per match, and bracket play never touches Torro.Rating. Phase 1's ELO is
// only read once, at seeding time, to build the initial bracket.

// BracketStatus constants
const (
	BracketStatusInProgress = "in_progress"
	BracketStatusCompleted  = "completed"
)

// BracketMatchStatus constants
const (
	BracketMatchStatusPending   = "pending"
	BracketMatchStatusCompleted = "completed"
)

// DefaultBracketSize is used when a bracket is created without an explicit
// size. Must be a power of two.
const DefaultBracketSize = 8

// Bracket represents a single-elimination knockout tournament for one class
// within one campaign, seeded from Phase 1 ELO ratings.
type Bracket struct {
	Id           string  `db:"Id"           json:"id"`
	CampaignId   string  `db:"CampaignId"   json:"campaign_id"`
	ClassId      string  `db:"ClassId"      json:"class_id"`
	Size         int     `db:"Size"         json:"size"`
	CurrentRound int     `db:"CurrentRound" json:"current_round"`
	Status       string  `db:"Status"       json:"status"`
	ChampionId   *string `db:"ChampionId"   json:"champion_id,omitempty"`
	CreatedAt    string  `db:"CreatedAt"    json:"created_at"`
	CompletedAt  *string `db:"CompletedAt"  json:"completed_at,omitempty"`
}

// BracketEntry is one seeded participant in a bracket. Seeds are assigned
// 1..N by descending Phase 1 rating at bracket-creation time.
type BracketEntry struct {
	Id         string  `db:"Id"         json:"id"`
	BracketId  string  `db:"BracketId"  json:"bracket_id"`
	TorronId   string  `db:"TorronId"   json:"torron_id"`
	Seed       int     `db:"Seed"       json:"seed"`
	SeedRating float64 `db:"SeedRating" json:"seed_rating"`
}

// BracketMatch is a single knockout match: (Round, Slot) uniquely identify
// its position in the bracket. Torro2Id is nil for a bye (Torro1Id
// auto-advances with no vote needed). WinnerId is nil until the match is
// decided.
type BracketMatch struct {
	Id        string  `db:"Id"        json:"id"`
	BracketId string  `db:"BracketId" json:"bracket_id"`
	Round     int     `db:"Round"     json:"round"`
	Slot      int     `db:"Slot"      json:"slot"`
	Torro1Id  string  `db:"Torro1Id"  json:"torro1_id"`
	Torro2Id  *string `db:"Torro2Id"  json:"torro2_id,omitempty"`
	WinnerId  *string `db:"WinnerId"  json:"winner_id,omitempty"`
	Status    string  `db:"Status"    json:"status"`
	CreatedAt string  `db:"CreatedAt" json:"created_at"`
}

// IsBye reports whether this match has no second competitor.
func (m *BracketMatch) IsBye() bool {
	return m.Torro2Id == nil
}

// BracketMatchVote is a single user's vote for a single match. The
// (MatchId, UserId) pair is unique at the DB level, enforcing "at most one
// vote per user per match".
type BracketMatchVote struct {
	Id        string `db:"Id"        json:"id"`
	MatchId   string `db:"MatchId"   json:"match_id"`
	UserId    string `db:"UserId"    json:"user_id"`
	TorronId  string `db:"TorronId"  json:"torron_id"`
	CreatedAt string `db:"CreatedAt" json:"created_at"`
}

// BracketRepo defines data access for the whole Phase 2 knockout schema
// (brackets, their seeded entries, their matches and the per-match vote
// log). It is intentionally one interface across four tables, mirroring
// how the "bracket" concept is owned by a single domain/repository file.
type BracketRepo interface {
	// -- Brackets --

	// Create creates a new bracket.
	Create(ctx context.Context, bracket *Bracket) (*Bracket, error)

	// Get retrieves a bracket by ID.
	Get(ctx context.Context, id string) (*Bracket, error)

	// GetByCampaignAndClass retrieves the bracket for a given campaign and
	// class, if one exists.
	GetByCampaignAndClass(ctx context.Context, campaignId string, classId string) (*Bracket, error)

	// GetLatestByClass retrieves the most recently created bracket for a
	// class, regardless of campaign. Used by the bracket overview page.
	GetLatestByClass(ctx context.Context, classId string) (*Bracket, error)

	// UpdateRound advances a bracket's current round pointer.
	UpdateRound(ctx context.Context, id string, round int) error

	// Complete marks a bracket as completed with its champion.
	Complete(ctx context.Context, id string, championId string) error

	// -- Bracket entries (seeding) --

	// CreateEntry adds a seeded participant to a bracket.
	CreateEntry(ctx context.Context, entry *BracketEntry) (*BracketEntry, error)

	// ListEntries lists all seeded participants for a bracket, ordered by
	// seed ascending.
	ListEntries(ctx context.Context, bracketId string) ([]*BracketEntry, error)

	// -- Bracket matches --

	// CreateMatch creates a single match.
	CreateMatch(ctx context.Context, match *BracketMatch) (*BracketMatch, error)

	// GetMatch retrieves a match by ID.
	GetMatch(ctx context.Context, id string) (*BracketMatch, error)

	// ListMatchesByRound lists all matches for a bracket round, ordered by
	// slot ascending.
	ListMatchesByRound(ctx context.Context, bracketId string, round int) ([]*BracketMatch, error)

	// ListMatches lists every match ever played in a bracket, ordered by
	// round then slot. Used to render past-round results.
	ListMatches(ctx context.Context, bracketId string) ([]*BracketMatch, error)

	// ListOpenMatchesForUser lists the still-open (pending) matches of a
	// round that the given user has not yet voted on. Used to serve a
	// random still-open match to a viewer.
	ListOpenMatchesForUser(ctx context.Context, bracketId string, round int, userId string) ([]*BracketMatch, error)

	// SetMatchWinner marks a match completed with the given winner.
	SetMatchWinner(ctx context.Context, matchId string, winnerId string) error

	// -- Votes --

	// CreateVote records a user's vote for a match. Returns a
	// DuplicateKeyError (via handleErrors) if the user already voted on
	// this match.
	CreateVote(ctx context.Context, vote *BracketMatchVote) (*BracketMatchVote, error)

	// HasVoted reports whether the user has already voted on this match.
	HasVoted(ctx context.Context, matchId string, userId string) (bool, error)

	// CountVotesByTorron tallies votes for a match, keyed by torron ID.
	// Torrons with zero votes are simply absent from the map.
	CountVotesByTorron(ctx context.Context, matchId string) (map[string]int, error)

	// -- Transaction methods --
	// The vote+advance flow is a read-then-write that needs the same
	// consistency guarantees as Phase 1's result handler, so every method
	// above that participates in that flow has a Tx twin.

	CreateTx(tx *sql.Tx, ctx context.Context, bracket *Bracket) (*Bracket, error)
	GetTx(tx *sql.Tx, ctx context.Context, id string) (*Bracket, error)
	UpdateRoundTx(tx *sql.Tx, ctx context.Context, id string, round int) error
	CompleteTx(tx *sql.Tx, ctx context.Context, id string, championId string) error

	CreateEntryTx(tx *sql.Tx, ctx context.Context, entry *BracketEntry) (*BracketEntry, error)
	ListEntriesTx(tx *sql.Tx, ctx context.Context, bracketId string) ([]*BracketEntry, error)

	CreateMatchTx(tx *sql.Tx, ctx context.Context, match *BracketMatch) (*BracketMatch, error)
	GetMatchTx(tx *sql.Tx, ctx context.Context, id string) (*BracketMatch, error)
	ListMatchesByRoundTx(tx *sql.Tx, ctx context.Context, bracketId string, round int) ([]*BracketMatch, error)
	SetMatchWinnerTx(tx *sql.Tx, ctx context.Context, matchId string, winnerId string) error

	CreateVoteTx(tx *sql.Tx, ctx context.Context, vote *BracketMatchVote) (*BracketMatchVote, error)
	CountVotesByTorronTx(tx *sql.Tx, ctx context.Context, matchId string) (map[string]int, error)
}

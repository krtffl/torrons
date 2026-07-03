package domain

import "context"

// FriendCircle is a shareable "friend group" a user can create. Other
// (anonymous, cookie-identified) users join by following the circle's
// invite link.
type FriendCircle struct {
	Id          string `db:"Id"          json:"id"`
	OwnerUserId string `db:"OwnerUserId" json:"owner_user_id"`
	InviteCode  string `db:"InviteCode"  json:"invite_code"`
	CreatedAt   string `db:"CreatedAt"   json:"created_at"`
}

// FriendCircleMember is a join-table row linking a user to a circle they
// belong to. A user can belong to multiple circles (they might be invited
// by several friends), so membership is many-to-many.
type FriendCircleMember struct {
	CircleId string `db:"CircleId" json:"circle_id"`
	UserId   string `db:"UserId"   json:"user_id"`
	JoinedAt string `db:"JoinedAt" json:"joined_at"`
}

// FriendCircleRepo defines the interface for friend-circle data access
type FriendCircleRepo interface {
	// Create creates a new circle owned by the given user (with a freshly
	// generated, unique invite code) and adds the owner as its first member
	Create(ctx context.Context, ownerUserId string) (*FriendCircle, error)

	// Get retrieves a circle by ID
	Get(ctx context.Context, id string) (*FriendCircle, error)

	// GetByInviteCode retrieves a circle by its invite code
	GetByInviteCode(ctx context.Context, inviteCode string) (*FriendCircle, error)

	// AddMember adds a user to a circle if they're not already a member
	// (idempotent -- safe to call every time someone opens an invite link)
	AddMember(ctx context.Context, circleId string, userId string) error

	// IsMember reports whether a user belongs to a circle
	IsMember(ctx context.Context, circleId string, userId string) (bool, error)

	// ListForUser lists every circle a user belongs to (owned or joined)
	ListForUser(ctx context.Context, userId string) ([]*FriendCircle, error)

	// GetCircleLeaderboard returns a leaderboard of torrons for the given
	// class, scoped to the circle's members: each torron's rating is the
	// average of its members' personalized ELO ratings (only members who
	// have actually rated that torron contribute).
	GetCircleLeaderboard(ctx context.Context, circleId string, classId string) ([]*UserLeaderboardEntry, error)

	// GetCircleGlobalLeaderboard is the circle-scoped equivalent across all
	// classes (top 100 by averaged member rating)
	GetCircleGlobalLeaderboard(ctx context.Context, circleId string) ([]*UserLeaderboardEntry, error)
}

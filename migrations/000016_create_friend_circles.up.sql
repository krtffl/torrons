-- Create FriendCircles table: a shareable "friend group" a user can create
-- and invite others to via a shareable invite code/link.
CREATE TABLE IF NOT EXISTS "FriendCircles" (
    "Id" VARCHAR(36) NOT NULL
        CONSTRAINT pk_friend_circles PRIMARY KEY,
    "OwnerUserId" VARCHAR(36) NOT NULL
        CONSTRAINT fk_friend_circles_owner
        REFERENCES "Users"("Id") ON DELETE CASCADE,
    "InviteCode" VARCHAR(20) NOT NULL,
    "CreatedAt" TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Invite codes must be unique so they can be looked up directly
CREATE UNIQUE INDEX idx_friend_circles_invite_code ON "FriendCircles"("InviteCode");

-- Index for listing circles a user owns
CREATE INDEX idx_friend_circles_owner ON "FriendCircles"("OwnerUserId");

-- Create FriendCircleMembers join table: a user can belong to multiple
-- circles (they might be invited by several friends), so this is
-- many-to-many between Users and FriendCircles.
CREATE TABLE IF NOT EXISTS "FriendCircleMembers" (
    "CircleId" VARCHAR(36) NOT NULL
        CONSTRAINT fk_friend_circle_members_circle
        REFERENCES "FriendCircles"("Id") ON DELETE CASCADE,
    "UserId" VARCHAR(36) NOT NULL
        CONSTRAINT fk_friend_circle_members_user
        REFERENCES "Users"("Id") ON DELETE CASCADE,
    "JoinedAt" TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT pk_friend_circle_members PRIMARY KEY ("CircleId", "UserId")
);

-- Index for listing every circle a given user belongs to
CREATE INDEX idx_friend_circle_members_user ON "FriendCircleMembers"("UserId");

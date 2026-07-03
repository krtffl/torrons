-- Drop indexes first
DROP INDEX IF EXISTS idx_friend_circle_members_user;
DROP INDEX IF EXISTS idx_friend_circles_owner;
DROP INDEX IF EXISTS idx_friend_circles_invite_code;

-- Drop tables (members first: it references FriendCircles via FK)
DROP TABLE IF EXISTS "FriendCircleMembers";
DROP TABLE IF EXISTS "FriendCircles";

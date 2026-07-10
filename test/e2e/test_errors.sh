#!/usr/bin/env bash
# Validation / error-handling / status-code correctness.
# Confirmed defects are encoded as xfail_status: the suite stays green on the
# known-buggy status but shouts XPASS the moment it's fixed. Correct behaviors
# are hard assertions.
section "ERRORS — validation, bad input, status-code correctness"

C_UNLOCKED='torrons_user_id=e2e-user-unlocked'

# --- correct behaviors (hard asserts) ---
assert_status E01 "unknown route -> 404"                 404 GET /no-such-route
assert_status E02 "torró detail, unknown id -> 404"      404 GET /torro/00000000-0000-0000-0000-000000000000
assert_status E03 "vote, wrong winner id -> 400"         400 POST "/pairings/${FX_PAIRING}/vote?id=not-a-torro"
assert_status E04 "vote, missing winner param -> 400"    400 POST "/pairings/${FX_PAIRING}/vote"
assert_status E05 "vote, empty winner param -> 400"      400 POST "/pairings/${FX_PAIRING}/vote?id="
assert_status E06 "wrong method on GET route -> 405"     405 POST /leaderboard
assert_status E07 "wrong method (PUT) -> 405"            405 PUT "/torro/${FX_T1}"
assert_status E08 "api lb class, unknown class -> 404"   404 GET /api/leaderboard/class/999
assert_status E09 "path traversal is neutralized"        404 GET /leaderboard/../../etc/passwd
assert_status E10 "friends circle, unknown id -> 404"    404 GET /friends/00000000-0000-0000-0000-000000000000

# --- FIXED (Batch 2): NotFound now mapped to 404 via domain.ErrFromRepo ---
assert_status S1  "vote screen, nonexistent class -> 404"  404 GET /classes/999/vote
assert_status S1b "vote screen, non-numeric class -> 404"  404 GET /classes/abc/vote
assert_status S2  "vote, nonexistent pairing -> 404"       404 POST "/pairings/00000000-0000-0000-0000-000000000000/vote?id=x"

# --- FIXED (Batch 2): negative OFFSET clamped to 0 -> 200 ---
assert_status S7  "history negative offset clamped -> 200" 200 GET "/history?offset=-5" -b "$C_UNLOCKED"
# non-numeric / overflow offset are handled (default to 0) — hard asserts
assert_status S7b "history non-numeric offset -> 200"    200 GET "/history?offset=abc" -b "$C_UNLOCKED"
assert_status S7c "history overflow offset -> 200"       200 GET "/history?offset=99999999999999999999" -b "$C_UNLOCKED"

# --- FIXED (Batch 10): unknown class id now 404 consistently across surfaces ---
assert_status S6a "leaderboard ?category=999 -> 404"           404 GET "/leaderboard?category=999"
assert_status S6b "api/user/leaderboard/class/999 -> 404"      404 GET /api/user/leaderboard/class/999 -b "$C_UNLOCKED"
assert_status S6c "embed/leaderboard ?classId=999 -> 404"      404 GET "/embed/leaderboard?classId=999"
# valid class / default still fine
assert_status S6d "leaderboard ?category=1 -> 200"             200 GET "/leaderboard?category=1"
assert_status S6e "embed/leaderboard default -> 200"           200 GET /embed/leaderboard

# --- gated pages below threshold render locked-state 200 (design), not an error ---
assert_status E11 "wrapped, fresh user -> 200 (locked state)" \
	200 GET /wrapped -b 'torrons_user_id=e2e-user-fresh'
assert_body_contains E11b "wrapped locked state message" "desbloq" GET /wrapped -b 'torrons_user_id=e2e-user-fresh'

# --- bracket admin input validation ---
assert_status E12 "bracket create size=-5 -> 400"       400 POST "/bracket/1/create?size=-5" -H "Authorization: Bearer ${ADMIN_TOKEN}"
assert_status E13 "bracket create size=7 (not pow2) -> 400" 400 POST "/bracket/1/create?size=7" -H "Authorization: Bearer ${ADMIN_TOKEN}"
# FIXED (B3): huge power-of-two used to reach standardSeedOrder's doubling
# allocation before the "not enough torrons" check ever ran; now rejected
# immediately by a MaxBracketSize bound. size=1 (technically 2^0, a valid
# power of two) is rejected too - a single-slot bracket has no matches.
assert_status E14 "bracket create size=huge pow2 -> 400 (fast, no OOM)" \
	400 POST "/bracket/1/create?size=1073741824" -H "Authorization: Bearer ${ADMIN_TOKEN}"
assert_status E15 "bracket create size=1 -> 400"        400 POST "/bracket/1/create?size=1" -H "Authorization: Bearer ${ADMIN_TOKEN}"

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

# --- confirmed defects: NotFound wrapped as ErrInternal -> 500 (should be 404) ---
xfail_status  S1  "vote screen, nonexistent class -> 500 (want 404)" \
	500 404 GET /classes/999/vote
xfail_status  S1b "vote screen, non-numeric class -> 500 (want 404)" \
	500 404 GET /classes/abc/vote
xfail_status  S2  "vote, nonexistent pairing -> 500 (want 404)" \
	500 404 POST "/pairings/00000000-0000-0000-0000-000000000000/vote?id=x"

# --- confirmed defect: negative OFFSET reaches SQL -> 500 (should clamp/400) ---
xfail_status  S7  "history negative offset -> 500 (want 200/400)" \
	500 200 GET "/history?offset=-5" -b "$C_UNLOCKED"
# non-numeric / overflow offset are handled (default to 0) — hard asserts
assert_status S7b "history non-numeric offset -> 200"    200 GET "/history?offset=abc" -b "$C_UNLOCKED"
assert_status S7c "history overflow offset -> 200"       200 GET "/history?offset=99999999999999999999" -b "$C_UNLOCKED"

# --- confirmed inconsistency: unknown class id contract differs across surfaces ---
# api/leaderboard/class/999 -> 404 (asserted E08 above), but these return 200+empty:
xfail_status  S6a "leaderboard ?category=999 -> 200 empty (api sibling 404s)" \
	200 404 GET "/leaderboard?category=999"
xfail_status  S6b "api/user/leaderboard/class/999 -> 200 empty (sibling 404s)" \
	200 404 GET /api/user/leaderboard/class/999 -b "$C_UNLOCKED"

# --- gated pages below threshold render locked-state 200 (design), not an error ---
assert_status E11 "wrapped, fresh user -> 200 (locked state)" \
	200 GET /wrapped -b 'torrons_user_id=e2e-user-fresh'
assert_body_contains E11b "wrapped locked state message" "desbloq" GET /wrapped -b 'torrons_user_id=e2e-user-fresh'

# --- bracket admin input validation ---
assert_status E12 "bracket create size=-5 -> 400"       400 POST "/bracket/1/create?size=-5" -H "Authorization: Bearer ${ADMIN_TOKEN}"
assert_status E13 "bracket create size=7 (not pow2) -> 400" 400 POST "/bracket/1/create?size=7" -H "Authorization: Bearer ${ADMIN_TOKEN}"

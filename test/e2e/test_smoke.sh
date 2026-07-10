#!/usr/bin/env bash
# Happy-path smoke: every route answers with its expected status + a content
# sanity check. Fixtures (FX_PAIRING, FX_T1, e2e-user-unlocked) come from run.sh.
section "SMOKE — every route, happy path (R1–R43)"

C_UNLOCKED='torrons_user_id=e2e-user-unlocked'

# --- static / SEO / infra ---
assert_status  R1  "healthcheck"                200 GET /healthcheck
assert_body_contains R1b "healthcheck body"     "answer" GET /healthcheck
assert_status  R5  "home"                       200 GET /
assert_body_contains R5b "home renders shell"   "torró" GET /
assert_status  R6  "classes list"               200 GET /classes
assert_status  R31 "about page"                 200 GET /sobre
assert_status  R32 "IGP explainer"              200 GET /torro-agramunt-igp
assert_status  R33 "agramunt vs xixona"         200 GET /torro-agramunt-vs-xixona
assert_status  R34 "torró glossary"             200 GET /tipus-de-torrons
assert_status  R43 "static css asset"           200 GET /public/css/main.css
assert_header  R43b "css content-type"          "content-type" "text/css" GET /public/css/main.css

# --- voting core ---
assert_status  R7  "vote screen for class 1"    200 GET /classes/1/vote
assert_body_contains R7b "vote screen has pairing" "pairings/" GET /classes/1/vote
assert_status  R9  "torró detail"               200 GET "/torro/${FX_T1}"

# --- leaderboards ---
assert_status  R10 "leaderboard (default)"      200 GET /leaderboard
assert_status  R10c "leaderboard global view"   200 GET "/leaderboard?view=global"
assert_status  R41 "api leaderboard global"     200 GET /api/leaderboard/global
assert_header  R41b "api lb global is json"     "content-type" "application/json" GET /api/leaderboard/global
assert_status  R42 "api leaderboard class 1"    200 GET /api/leaderboard/class/1

# --- personal surfaces (unlocked user) ---
assert_status  R11 "stats (unlocked)"           200 GET /stats -b "$C_UNLOCKED"
assert_status  R12 "history (unlocked)"         200 GET /history -b "$C_UNLOCKED"
assert_status  R35 "api user stats"             200 GET /api/user/stats -b "$C_UNLOCKED"
assert_header  R35b "user stats is json"        "content-type" "application/json" GET /api/user/stats -b "$C_UNLOCKED"
assert_status  R36 "api user lb class 1"        200 GET /api/user/leaderboard/class/1 -b "$C_UNLOCKED"
assert_status  R37 "api user lb global (>=50)"  200 GET /api/user/leaderboard/global -b "$C_UNLOCKED"
assert_status  R26 "wrapped (unlocked)"         200 GET /wrapped -b "$C_UNLOCKED"
assert_status  R29 "reveal (unlocked)"          200 GET /reveal -b "$C_UNLOCKED"

# --- generated images (PNG) ---
assert_status  R13 "share card png"             200 GET /share/card.png
assert_header  R13b "share card is png"         "content-type" "image/png" GET /share/card.png
assert_status  R27 "wrapped card png"           200 GET /wrapped/card.png -b "$C_UNLOCKED"
assert_status  R30 "reveal card png"            200 GET /reveal/card.png -b "$C_UNLOCKED"
assert_status  R28 "press-kit card png"         200 GET /press-kit/card.png

# --- campaign / advent / press / embed / friends ---
assert_status  R38 "campaign countdown json"    200 GET /api/campaign/countdown
assert_status  R39 "campaign countdown widget"  200 GET /api/campaign/countdown/widget
# /api/campaign/info returns 404 "No active campaign" when none is active — the
# suite seeds no campaign, so 404 is the correct behavior here (not an error).
# run.sh seeds an active "e2e-campaign" fixture (needed for the bracket-create
# tests), so /api/campaign/info correctly returns the campaign, not a 404.
assert_status  R40 "campaign info (active campaign seeded)" 200 GET /api/campaign/info
assert_status  R19 "advent"                     200 GET /advent
assert_status  R25 "premsa (press) page"        200 GET /premsa
assert_status  R24 "embed leaderboard"          200 GET /embed/leaderboard
assert_status  R20 "friends index"              200 GET /friends -b "$C_UNLOCKED"
assert_status  R14 "bracket overview class 5"   200 GET /bracket/5

# --- embed must NOT set a user cookie (cross-origin iframe path) ---
_embed_setcookie="$(curl_header GET /embed/leaderboard 'set-cookie')"
if [[ -z "$_embed_setcookie" ]]; then _ok "[R24b] /embed/ sets no user cookie"; else _no "[R24b] /embed/ leaked Set-Cookie: $_embed_setcookie"; fi

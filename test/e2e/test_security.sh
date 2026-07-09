#!/usr/bin/env bash
# Security regression guards: rate limiting, admin auth, identity, injection safety.
section "SECURITY — rate limiting, auth, identity, injection"

# --- admin token gate (fail-closed) ---
assert_status SEC01 "bracket create, no token -> 401"    401 POST /bracket/1/create
assert_status SEC02 "bracket create, wrong token -> 401" 401 POST /bracket/1/create -H "Authorization: Bearer wrong-token"
assert_status SEC03 "bracket advance, no token -> 401"   401 POST /bracket/00000000-0000-0000-0000-000000000000/advance
assert_header  SEC04 "401 sets WWW-Authenticate"         "www-authenticate" "Bearer" POST /bracket/1/create

# --- injection safety: metacharacter ids must behave like unknown ids, never 500-by-injection ---
assert_status SEC05 "SQL-metachar torró id is inert -> 404" \
	404 GET "/torro/x%27%3B%20DROP%20TABLE%20%22Torrons%22%3B--"
# prove the table still exists afterwards by hitting a route that reads it
assert_status SEC05b "Torrons table intact after injection attempt" 200 GET /classes/1/vote

# --- global per-IP rate limit works for a FIXED ip ---
_rl_fixed() {
	local ip="203.0.113.77" n=115 got429=0 i code
	for i in $(seq 1 $n); do
		code=$(curl -s -o /dev/null -w '%{http_code}' -H "X-Forwarded-For: $ip" "${BASE_URL}/api/campaign/countdown")
		[[ "$code" == "429" ]] && got429=1
	done
	echo "$got429"
}
if [[ "$(_rl_fixed)" == "1" ]]; then
	_ok "[SEC06] global rate limit fires at >100/min for a fixed IP"
else
	_no "[SEC06] global rate limit never fired for a fixed IP (limiter broken?)"
fi

# --- CONFIRMED FINDING: rate limit is bypassable by rotating X-Forwarded-For ---
# RealIP trusts the header from any client, so a rotating XFF defeats both the
# global and the per-user vote limiter. We assert the CURRENT (vulnerable)
# behavior so this guard flips loudly once a trusted-proxy allowlist is added.
_rl_rotating() {
	local n=140 blocked=0 i code
	for i in $(seq 1 $n); do
		code=$(curl -s -o /dev/null -w '%{http_code}' -H "X-Forwarded-For: 10.10.$((i/256)).$((i%256))" "${BASE_URL}/api/campaign/countdown")
		[[ "$code" == "429" ]] && blocked=$((blocked+1))
	done
	echo "$blocked"
}
_blocked="$(_rl_rotating)"
if [[ "$_blocked" == "0" ]]; then
	XFAIL=$((XFAIL+1))
	printf '  \033[33mxfail\033[0m [SEC07] rate limit fully bypassed by rotating X-Forwarded-For (140/140 passed) — KNOWN: add trusted-proxy allowlist\n'
else
	XPASS=$((XPASS+1))
	printf '  \033[36mXPASS\033[0m [SEC07] rotating-XFF bypass now partially blocked (%s/140 got 429) — trusted-proxy fix may be in; tighten this guard\n' "$_blocked"
fi

# --- identity model: an unknown-UUID cookie is treated as anonymous (new user),
#     and /api/user/stats reflects whatever cookie is presented. This documents
#     that user identity is an unauthenticated bearer of the UUID (by design);
#     the guard exists so any future hardening (signed cookies) is noticed. ---
_forged="torrons_user_id=e2e-user-unlocked"
assert_body_contains SEC08 "user stats keyed purely off cookie UUID" \
	'"total_votes":60' GET /api/user/stats -b "$_forged"

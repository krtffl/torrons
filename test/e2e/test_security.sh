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

# --- FIXED (Batch 3): X-Forwarded-For is only honored from a trusted proxy ---
# The main suite server runs with the default trusted set (loopback trusted), so
# its XFF rotation is honored — that's why the functional tests can dodge the
# per-IP limit. To prove the bypass is CLOSED for an untrusted peer, boot a
# second instance with TRUSTED_PROXIES set to a range that EXCLUDES loopback:
# now the loopback test client is untrusted, its spoofed XFF is ignored, and all
# requests key on the real ::1 peer -> the rotation no longer bypasses the limit.
if [[ -n "${SERVER_BIN:-}" ]]; then
	_sec_port=8479
	_host="$(sed -E 's|.*@([^:/]+):.*|\1|' <<<"$DB_URL")"
	_dbport="$(sed -E 's|.*:([0-9]+)/.*|\1|' <<<"$DB_URL")"
	env -C "$ROOT" DB_HOST="$_host" DB_PORT="$_dbport" DB_USER=myUser DB_PASSWORD=myPassword \
		DB_NAME=databaseName DB_SSL_MODE=disable PORT="$_sec_port" LOGGER_LEVEL=error LOGGER_PATH= \
		ADMIN_TOKEN="$ADMIN_TOKEN" TRUSTED_PROXIES="203.0.113.0/24" \
		"$SERVER_BIN" >/dev/null 2>&1 &
	_sec_pid=$!
	disown "$_sec_pid" 2>/dev/null || true
	for _i in $(seq 1 40); do curl -s -o /dev/null "http://localhost:${_sec_port}/healthcheck" 2>/dev/null && break; sleep 0.5; done
	_blocked=0
	for _i in $(seq 1 140); do
		_c=$(curl -s -o /dev/null -w '%{http_code}' -H "X-Forwarded-For: 10.10.$((_i/256)).$((_i%256))" "http://localhost:${_sec_port}/api/campaign/countdown")
		[[ "$_c" == "429" ]] && _blocked=$((_blocked+1))
	done
	kill -9 "$_sec_pid" 2>/dev/null
	if (( _blocked > 0 )); then
		_ok "[SEC07] untrusted peer: spoofed rotating X-Forwarded-For ignored, rate limit enforced ($_blocked/140 blocked)"
	else
		_no "[SEC07] untrusted peer still bypassed the rate limit via rotating X-Forwarded-For (0/140 blocked)"
	fi
else
	printf '  \033[36minfo\033[0m  [SEC07] skipped (SERVER_BIN not exported; run via run.sh)\n'
fi

# --- identity model: an unknown-UUID cookie is treated as anonymous (new user),
#     and /api/user/stats reflects whatever cookie is presented. This documents
#     that user identity is an unauthenticated bearer of the UUID (by design);
#     the guard exists so any future hardening (signed cookies) is noticed. ---
_forged="torrons_user_id=e2e-user-unlocked"
assert_body_contains SEC08 "user stats keyed purely off cookie UUID" \
	'"total_votes":60' GET /api/user/stats -b "$_forged"

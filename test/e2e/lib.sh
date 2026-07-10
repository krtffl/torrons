#!/usr/bin/env bash
# Shared assertion + fixture helpers for the black-box e2e suite.
# Sourced by run.sh; expects BASE_URL, PSQL(), and counters to be exported.

# --- counters ---
: "${PASS:=0}" "${FAIL:=0}" "${XFAIL:=0}" "${XPASS:=0}"
FAILED_CASES=()

# Rotating X-Forwarded-For so the 100/min-per-IP global limiter never masks a
# functional assertion. (That the limiter *can* be bypassed this way is itself
# a tested finding — see test_security.sh — but for functional checks we want
# to exercise handlers, not the limiter.)
_xff() { printf 'X-Forwarded-For: 198.51.%d.%d' $((RANDOM % 255)) $((RANDOM % 255)); }

# curl_code METHOD PATH [curl-args...] -> prints numeric HTTP status
curl_code() {
	local method="$1" path="$2"
	shift 2
	curl -s -o /dev/null -w '%{http_code}' -H "$(_xff)" -X "$method" "$@" "${BASE_URL}${path}"
}

# curl_body METHOD PATH [curl-args...] -> prints response body
curl_body() {
	local method="$1" path="$2"
	shift 2
	curl -s -H "$(_xff)" -X "$method" "$@" "${BASE_URL}${path}"
}

# curl_header METHOD PATH HEADER-NAME [curl-args...] -> prints that response header's value
curl_header() {
	local method="$1" path="$2" hdr="$3"
	shift 3
	curl -s -o /dev/null -D - -H "$(_xff)" -X "$method" "$@" "${BASE_URL}${path}" \
		| grep -i "^${hdr}:" | head -1 | sed "s/^[^:]*: //I" | tr -d '\r'
}

_ok()   { PASS=$((PASS + 1));  printf '  \033[32mok\033[0m   %s\n' "$1"; }
_no()   { FAIL=$((FAIL + 1));  FAILED_CASES+=("$1"); printf '  \033[31mFAIL\033[0m %s\n' "$1"; }

# assert_status ID "desc" EXPECTED METHOD PATH [curl-args...]
assert_status() {
	local id="$1" desc="$2" expected="$3" method="$4" path="$5"
	shift 5
	local got
	got=$(curl_code "$method" "$path" "$@")
	if [[ "$got" == "$expected" ]]; then
		_ok "[$id] $desc ($method $path -> $got)"
	else
		_no "[$id] $desc ($method $path -> got $got, want $expected)"
	fi
}

# assert_body_contains ID "desc" NEEDLE METHOD PATH [curl-args...]
assert_body_contains() {
	local id="$1" desc="$2" needle="$3" method="$4" path="$5"
	shift 5
	local body
	body=$(curl_body "$method" "$path" "$@")
	if grep -qiF -- "$needle" <<<"$body"; then
		_ok "[$id] $desc (body contains '$needle')"
	else
		_no "[$id] $desc (body missing '$needle')"
	fi
}

# assert_header ID "desc" HEADER EXPECTED-SUBSTR METHOD PATH [curl-args...]
assert_header() {
	local id="$1" desc="$2" hdr="$3" expected="$4" method="$5" path="$6"
	shift 6
	local got
	got=$(curl_header "$method" "$path" "$hdr" "$@")
	if grep -qiF -- "$expected" <<<"$got"; then
		_ok "[$id] $desc ($hdr: $got)"
	else
		_no "[$id] $desc ($hdr: '$got' missing '$expected')"
	fi
}

# xfail_status: a KNOWN, currently-unfixed defect. We assert the *buggy* status
# so the suite stays green on a known bug, but LOUDLY flips to a failure the
# moment the behavior changes (i.e. once the bug is fixed, come update this to
# assert_status). This is how the suite guards against silent regressions in
# both directions.
#   xfail_status ID "desc" BUGGY_EXPECTED CORRECT METHOD PATH [curl-args...]
xfail_status() {
	local id="$1" desc="$2" buggy="$3" correct="$4" method="$5" path="$6"
	shift 6
	local got
	got=$(curl_code "$method" "$path" "$@")
	if [[ "$got" == "$buggy" ]]; then
		XFAIL=$((XFAIL + 1))
		printf '  \033[33mxfail\033[0m %s (KNOWN BUG: %s -> %s; correct=%s)\n' "$id" "$desc" "$got" "$correct"
	elif [[ "$got" == "$correct" ]]; then
		XPASS=$((XPASS + 1))
		printf '  \033[36mXPASS\033[0m %s (BUG FIXED? %s now returns %s — update this guard to assert_status)\n' "$id" "$desc" "$got"
	else
		_no "[$id] $desc ($method $path -> got $got, was buggy=$buggy, correct=$correct — new behavior)"
	fi
}

section() { printf '\n\033[1m### %s\033[0m\n' "$1"; }

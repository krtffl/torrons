#!/usr/bin/env bash
# Performance budget smoke test. NOT a load test — it flags endpoints whose
# single-request latency exceeds a budget, and specifically guards the
# /premsa aggregation hotspot (3+ full seq-scans of the Results table per hit).
#
# These budgets are generous and meant to catch ORDER-OF-MAGNITUDE regressions,
# not micro-fluctuations. On an empty DB everything is fast; the real signal
# comes from running this against a volume-seeded DB (see docs/PERFORMANCE_AUDIT.md
# for the 200k-row methodology). Budgets are overridable via env.
section "PERF — per-request latency budgets (single request)"

BUDGET_STATIC_MS="${BUDGET_STATIC_MS:-300}"
BUDGET_AGG_MS="${BUDGET_AGG_MS:-1500}"
BUDGET_PREMSA_MS="${BUDGET_PREMSA_MS:-2000}"

C_UNLOCKED='torrons_user_id=e2e-user-unlocked'

_latency_ms() {
	local path="$1"; shift
	local t
	t=$(curl -s -o /dev/null -w '%{time_total}' -H "$(_xff)" "$@" "${BASE_URL}${path}")
	python3 -c "print(int($t*1000))"
}

_budget() {
	local id="$1" desc="$2" budget="$3" path="$4"; shift 4
	local ms
	ms=$(_latency_ms "$path" "$@")
	if (( ms <= budget )); then
		_ok "[$id] $desc (${ms}ms ≤ ${budget}ms)"
	else
		_no "[$id] $desc (${ms}ms > ${budget}ms budget — perf regression)"
	fi
}

_budget PERF01 "home page"            "$BUDGET_STATIC_MS" /
_budget PERF02 "classes list"         "$BUDGET_STATIC_MS" /classes
_budget PERF03 "vote screen"          "$BUDGET_AGG_MS"    /classes/1/vote
_budget PERF04 "leaderboard"          "$BUDGET_AGG_MS"    /leaderboard
_budget PERF05 "api leaderboard glob" "$BUDGET_AGG_MS"    /api/leaderboard/global
_budget PERF06 "stats"                "$BUDGET_AGG_MS"    /stats -b "$C_UNLOCKED"
_budget PERF07 "wrapped"              "$BUDGET_AGG_MS"    /wrapped -b "$C_UNLOCKED"
_budget PERF08 "premsa (hotspot)"     "$BUDGET_PREMSA_MS" /premsa

# Concurrency signal for the /premsa hotspot: fire a small burst and report the
# max latency. Informational (does not fail the run) unless it blows past the
# WriteTimeout, which would surface as truncated/aborted responses in prod.
_burst=12
tmp="$(mktemp -d)"
for i in $(seq 1 "$_burst"); do
	( curl -s -o /dev/null -w '%{http_code} %{time_total}\n' -H "X-Forwarded-For: 192.0.2.$i" "${BASE_URL}/premsa" >"$tmp/$i" ) &
done
wait
maxms=$(cat "$tmp"/* | awk '{print $2}' | sort -rn | head -1 | python3 -c "import sys; print(int(float(sys.stdin.read())*1000))")
non200=$(cat "$tmp"/* | awk '$1!="200"' | wc -l | tr -d ' ')
rm -rf "$tmp"
printf '  \033[36minfo\033[0m  [PERF09] /premsa under %s concurrent: max=%sms, non-200=%s (single-request budget=%sms)\n' "$_burst" "$maxms" "$non200" "$BUDGET_PREMSA_MS"
if (( non200 > 0 )); then _no "[PERF09b] /premsa returned $non200 non-200 under a ${_burst}-way burst (pool/timeout exhaustion)"; else _ok "[PERF09b] /premsa served all $_burst concurrent requests with 200"; fi

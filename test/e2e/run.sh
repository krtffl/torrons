#!/usr/bin/env bash
#
# Reproducible black-box end-to-end test suite for the Torrorèndum API.
#
# It stands up a throwaway Postgres, migrates it, seeds deterministic fixtures,
# builds and boots the real server binary, exercises every endpoint and every
# error/validation/security/concurrency case, then tears everything down.
#
# Usage:
#   test/e2e/run.sh                 # full self-contained run (spins up podman PG)
#   KEEP=1 test/e2e/run.sh          # leave server + DB up after the run (for poking)
#   ONLY=security test/e2e/run.sh   # run only test_security.sh
#   TORRO_E2E_REUSE_DB_URL=postgres://... test/e2e/run.sh   # use an existing DB
#
# Exit code is non-zero if any hard assertion fails (xfail known-bug guards do
# not fail the run, but an xfail that starts passing — i.e. a fixed bug — prints
# an XPASS notice so you remember to tighten the guard).

set -uo pipefail
cd "$(dirname "$0")"
HERE="$(pwd)"
ROOT="$(cd ../.. && pwd)"

# ---- configuration (override via env) ----
PG_CONTAINER="${PG_CONTAINER:-torrons-e2e-pg}"
PG_PORT="${PG_PORT:-55440}"
PG_IMAGE="${PG_IMAGE:-postgres:16}"
APP_PORT="${APP_PORT:-8477}"
ADMIN_TOKEN="${ADMIN_TOKEN:-e2e-admin-secret}"
export BASE_URL="http://localhost:${APP_PORT}"
export ADMIN_TOKEN
WORKDIR="$(mktemp -d)"
SERVER_LOG="${WORKDIR}/server.log"
SERVER_BIN="${WORKDIR}/server"
SERVER_PID=""
OWN_DB=0

log()  { printf '\033[34m[e2e]\033[0m %s\n' "$*"; }
die()  { printf '\033[31m[e2e] FATAL:\033[0m %s\n' "$*" >&2; teardown; exit 2; }

# PSQL runs SQL against the test DB. Uses podman exec when we own the container,
# else a psql client against the reuse URL.
if [[ -n "${TORRO_E2E_REUSE_DB_URL:-}" ]]; then
	DB_URL="$TORRO_E2E_REUSE_DB_URL"
	PSQL() { psql "$DB_URL" "$@"; }
else
	DB_URL="postgres://myUser:myPassword@localhost:${PG_PORT}/databaseName?sslmode=disable"
	PSQL() { podman exec -i "$PG_CONTAINER" psql -U myUser -d databaseName "$@"; }
fi
export DB_URL
export -f PSQL 2>/dev/null || true

teardown() {
	[[ "${KEEP:-0}" == "1" ]] && { log "KEEP=1 — leaving server (pid $SERVER_PID) + DB up"; return; }
	if [[ -n "$SERVER_PID" ]]; then
		# Kill the REAL bound PID (see engineering notes: never trust $! alone).
		local real_pid
		real_pid="$(ss -tlnp 2>/dev/null | grep ":${APP_PORT} " | grep -oE 'pid=[0-9]+' | head -1 | cut -d= -f2)"
		[[ -n "$real_pid" ]] && kill -9 "$real_pid" 2>/dev/null
		kill -9 "$SERVER_PID" 2>/dev/null
	fi
	if [[ "$OWN_DB" == "1" ]]; then
		podman stop "$PG_CONTAINER" >/dev/null 2>&1
		podman rm -f "$PG_CONTAINER" >/dev/null 2>&1
	fi
	rm -rf "$WORKDIR"
}
trap teardown EXIT INT TERM

# ---- 1. database ----
setup_db() {
	if [[ -n "${TORRO_E2E_REUSE_DB_URL:-}" ]]; then
		log "reusing DB at $DB_URL"
		return
	fi
	log "starting throwaway Postgres ($PG_IMAGE) on :$PG_PORT"
	podman rm -f "$PG_CONTAINER" >/dev/null 2>&1
	podman run -d --rm --name "$PG_CONTAINER" \
		-e POSTGRES_USER=myUser -e POSTGRES_PASSWORD=myPassword -e POSTGRES_DB=databaseName \
		-p "${PG_PORT}:5432" "$PG_IMAGE" >/dev/null || die "could not start postgres container"
	OWN_DB=1
	log "waiting for postgres to accept connections..."
	local i
	for i in $(seq 1 30); do
		if podman exec "$PG_CONTAINER" pg_isready -U myUser -d databaseName >/dev/null 2>&1; then
			sleep 2  # pg_isready races the first real connection; give it a beat
			return
		fi
		sleep 1
	done
	die "postgres did not become ready in time"
}

# ---- 2. build + boot the real server ----
boot_server() {
	log "building server binary"
	( cd "$ROOT" && go build -o "$SERVER_BIN" ./cmd/server ) || die "go build failed"
	# Exported so security sub-tests can boot a second instance with a different
	# TRUSTED_PROXIES config (see test_security.sh SEC07).
	export SERVER_BIN
	log "booting server on :$APP_PORT (migrations run automatically on boot)"
	local host port
	host="$(sed -E 's|.*@([^:/]+):.*|\1|' <<<"$DB_URL")"
	port="$(sed -E 's|.*:([0-9]+)/.*|\1|' <<<"$DB_URL")"
	env -C "$ROOT" \
		DB_HOST="$host" DB_PORT="$port" DB_USER=myUser DB_PASSWORD=myPassword \
		DB_NAME=databaseName DB_SSL_MODE=disable PORT="$APP_PORT" \
		LOGGER_LEVEL=error LOGGER_PATH= ADMIN_TOKEN="$ADMIN_TOKEN" \
		"$SERVER_BIN" >"$SERVER_LOG" 2>&1 &
	SERVER_PID=$!
	# Detach the server from this shell's job table. Test modules use a bare
	# `wait` to await their fired-off curls; without this, `wait` would also
	# block on the never-exiting server job and deadlock the whole run.
	disown "$SERVER_PID" 2>/dev/null || disown 2>/dev/null || true
	# First boot against a fresh DB runs migrations AND seeds ~3138 pairings
	# one-by-one (CheckPairingsCreated/createGlobalPairings) BEFORE the HTTP
	# listener opens, which can take a while — hence the generous window.
	local i
	for i in $(seq 1 120); do
		if curl -s -o /dev/null "http://localhost:${APP_PORT}/healthcheck" 2>/dev/null; then
			log "server up (pid $SERVER_PID)"
			return
		fi
		if ! kill -0 "$SERVER_PID" 2>/dev/null; then
			cat "$SERVER_LOG" >&2; die "server process died during boot"
		fi
		sleep 1
	done
	cat "$SERVER_LOG" >&2; die "server did not answer /healthcheck in time"
}

# ---- 3. deterministic fixtures ----
# The suite needs, by stable handle, at least: a class id, a pairing id + its two
# torró ids, a user with >=50 votes (unlocks gated pages), a user with 0 votes.
seed_fixtures() {
	log "seeding deterministic fixtures"
	PSQL -q -v ON_ERROR_STOP=1 >/dev/null <<-SQL || die "fixture seed failed"
		-- unlocked user (>= 50 votes, gates all pass)
		INSERT INTO "Users"("Id","VoteCount","ClassVotes","CurrentStreak","LongestStreak","LastVoteDate")
		VALUES ('e2e-user-unlocked', 60, '{"1":60,"5":60}', 3, 9, CURRENT_DATE)
		ON CONFLICT ("Id") DO UPDATE SET "VoteCount"=60, "ClassVotes"='{"1":60,"5":60}';
		-- fresh user (0 votes, gates fail)
		INSERT INTO "Users"("Id","VoteCount","ClassVotes")
		VALUES ('e2e-user-fresh', 0, '{}')
		ON CONFLICT ("Id") DO UPDATE SET "VoteCount"=0;
		-- give the unlocked user some ELO snapshots so /reveal /wrapped have data
		-- (Id is VARCHAR(36); gen_random_uuid()::text is exactly 36 chars)
		INSERT INTO "UserEloSnapshots"("Id","UserId","TorronId","Rating","VoteCount")
		SELECT gen_random_uuid()::text, 'e2e-user-unlocked', t."Id", 1400+random()*200, 5
		FROM "Torrons" t LIMIT 20
		ON CONFLICT ("UserId","TorronId") DO NOTHING;
		-- some results by the unlocked user so /history and /stats have rows
		INSERT INTO "Results"("Id","Pairing","Torro1RatingBefore","Torro2RatingBefore","Winner","Torro1RatingAfter","Torro2RatingAfter","UserId")
		SELECT 'e2e-res-'||g, p."Id", 1500,1500, p."Torro1", 1510,1490, 'e2e-user-unlocked'
		FROM (SELECT "Id","Torro1" FROM "Pairings" WHERE "Class"='1' LIMIT 1) p, generate_series(1,5) g
		ON CONFLICT ("Id") DO NOTHING;
	SQL

	# Export handles the test files read.
	FX_PAIRING="$(PSQL -t -A -c "SELECT \"Id\" FROM \"Pairings\" WHERE \"Class\"='1' ORDER BY \"Id\" LIMIT 1" | tr -d '[:space:]')"
	FX_T1="$(PSQL -t -A -c "SELECT \"Torro1\" FROM \"Pairings\" WHERE \"Id\"='$FX_PAIRING'" | tr -d '[:space:]')"
	FX_T2="$(PSQL -t -A -c "SELECT \"Torro2\" FROM \"Pairings\" WHERE \"Id\"='$FX_PAIRING'" | tr -d '[:space:]')"
	FX_TORRO="$FX_T1"
	export FX_PAIRING FX_T1 FX_T2 FX_TORRO
	[[ -n "$FX_PAIRING" && -n "$FX_T1" ]] || die "could not resolve pairing fixtures"
	log "fixtures: pairing=$FX_PAIRING t1=$FX_T1 t2=$FX_T2"
}

# ---- run ----
main() {
	setup_db
	boot_server
	seed_fixtures

	# shellcheck disable=SC1091
	source "$HERE/lib.sh"

	local files
	if [[ -n "${ONLY:-}" ]]; then
		files=("$HERE/test_${ONLY}.sh")
	else
		files=("$HERE"/test_*.sh)
	fi

	for f in "${files[@]}"; do
		[[ -f "$f" ]] || { log "no such suite: $f"; continue; }
		# shellcheck disable=SC1090
		source "$f"
	done

	printf '\n\033[1m==================== SUMMARY ====================\033[0m\n'
	printf '  passed:        %s\n' "$PASS"
	printf '  failed:        %s\n' "$FAIL"
	printf '  known-bug xfail: %s\n' "$XFAIL"
	printf '  bug-now-fixed XPASS: %s\n' "$XPASS"
	if ((FAIL > 0)); then
		printf '\n\033[31mFAILURES:\033[0m\n'
		printf '  - %s\n' "${FAILED_CASES[@]}"
		return 1
	fi
	printf '\n\033[32mAll hard assertions passed.\033[0m\n'
	return 0
}

main
rc=$?
exit $rc

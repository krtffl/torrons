#!/usr/bin/env bash
# CONFIRMED CRITICAL FINDING: the vote path does read-modify-write on
# Torrons.Rating inside a transaction with NO "SELECT ... FOR UPDATE" and an
# absolute "UPDATE SET Rating=$2". Concurrent votes on the same torró read the
# same pre-image and clobber each other -> lost updates, corrupted ELO.
#
# This test fires N concurrent votes for the SAME winner on the SAME pairing
# from a clean 1500/1500 baseline, then compares the resulting rating against
# the mathematically-correct serialized value. If per-row locking is correct
# the two match within tolerance; today they diverge sharply.
section "CONCURRENCY — ELO lost-update race on Torrons.Rating"

N="${RACE_VOTES:-40}"
TOL="${RACE_TOLERANCE:-15}"   # ELO points

# reset the two torrons in FX_PAIRING to a clean baseline and clear their results
PSQL -q >/dev/null <<-SQL
	UPDATE "Torrons" SET "Rating"=1500 WHERE "Id" IN ('${FX_T1}','${FX_T2}');
	DELETE FROM "Results" WHERE "Pairing"='${FX_PAIRING}';
SQL

# fire N concurrent votes for FX_T1 as winner (rotating XFF so the per-IP limiter
# doesn't cap us below N; that limiter's bypassability is itself tested elsewhere)
tmp="$(mktemp -d)"
for i in $(seq 1 "$N"); do
	( curl -s -o /dev/null -w '%{http_code}\n' \
		-H "X-Forwarded-For: 172.20.$((i/256)).$((i%256))" \
		-X POST "${BASE_URL}/pairings/${FX_PAIRING}/vote?id=${FX_T1}" >"$tmp/$i" ) &
done
wait
ok200=$(grep -l '^200' "$tmp"/* 2>/dev/null | wc -l | tr -d ' ')
rm -rf "$tmp"

observed="$(PSQL -t -A -c "SELECT round(\"Rating\",2) FROM \"Torrons\" WHERE \"Id\"='${FX_T1}'" | tr -d '[:space:]')"
results_written="$(PSQL -t -A -c "SELECT count(*) FROM \"Results\" WHERE \"Pairing\"='${FX_PAIRING}'" | tr -d '[:space:]')"

expected="$(python3 -c "
K=42; r1=r2=1500.0
for _ in range($N):
    e1=1/(1+10**((r2-r1)/400)); e2=1/(1+10**((r1-r2)/400))
    r1,r2 = r1+K*(1-e1), r2+K*(0-e2)
print(round(r1,2))
")"

diff="$(python3 -c "print(round(abs($observed-$expected),2))")"
printf '  N=%s concurrent votes | 200s=%s | Results rows=%s\n' "$N" "$ok200" "$results_written"
printf '  winner rating: observed=%s  serialized-correct=%s  |Δ|=%s (tol=%s)\n' "$observed" "$expected" "$diff" "$TOL"

# Sanity: every vote must still be recorded (the race corrupts ratings, not row writes)
if [[ "$results_written" == "$N" ]]; then _ok "[RACE0] all $N votes recorded as Results rows"; else _no "[RACE0] only $results_written/$N Results rows written"; fi

# FIXED (Batch 1): the vote tx now locks both torró rows FOR UPDATE in sorted
# id order, so concurrent votes on the same torró serialize their ELO
# read-modify-write instead of clobbering each other. Hard assertion now.
within="$(python3 -c "print(1 if abs($observed-$expected)<=$TOL else 0)")"
if [[ "$within" == "1" ]]; then
	_ok "[RACE1] ELO consistent under $N concurrent votes (|Δ|=$diff ≤ $TOL, no lost update)"
else
	_no "[RACE1] lost-update race present: |Δ|=$diff > $TOL — concurrent votes clobbered each other's rating"
fi

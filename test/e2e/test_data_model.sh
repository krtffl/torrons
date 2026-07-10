#!/usr/bin/env bash
# Data-model integrity checks that aren't observable through a single HTTP
# response, so they query the DB directly (PSQL, from run.sh) the same way
# test_concurrency.sh does.
section "DATA MODEL — Pairings uniqueness"

# FIXED: a multi-instance boot race could previously double-seed the same
# matchup (in either Torro1/Torro2 order, since ListByClass has no ORDER BY).
# idx_pairings_unique_matchup (migration 000020) now enforces this at the DB
# level, order-independent via LEAST()/GREATEST().
_idx="$(PSQL -t -A -c "SELECT indexname FROM pg_indexes WHERE indexname='idx_pairings_unique_matchup'" | tr -d '[:space:]')"
if [[ "$_idx" == "idx_pairings_unique_matchup" ]]; then
	_ok "[DATA01] idx_pairings_unique_matchup exists"
else
	_no "[DATA01] idx_pairings_unique_matchup missing"
fi

# No duplicate matchups exist in the seeded data (order-independent check).
_dupes="$(PSQL -t -A -c 'SELECT count(*) FROM (SELECT 1 FROM "Pairings" GROUP BY LEAST("Torro1","Torro2"), GREATEST("Torro1","Torro2"), "Class" HAVING count(*) > 1) d' | tr -d '[:space:]')"
if [[ "$_dupes" == "0" ]]; then
	_ok "[DATA02] no duplicate (order-independent) pairing matchups in seeded data"
else
	_no "[DATA02] found $_dupes duplicate matchup group(s)"
fi

# A swapped-order duplicate of an existing pairing is rejected at the DB level.
_p="$(PSQL -t -A -F'|' -c 'SELECT "Torro1","Torro2","Class" FROM "Pairings" LIMIT 1')"
_t1="${_p%%|*}"; _rest="${_p#*|}"; _t2="${_rest%%|*}"; _class="${_rest##*|}"
_insert_err="$(PSQL -t -A -c "INSERT INTO \"Pairings\"(\"Id\",\"Torro1\",\"Torro2\",\"Class\") VALUES ('e2e-dup-check','${_t2}','${_t1}','${_class}')" 2>&1)"
if grep -qi 'duplicate key value violates unique constraint "idx_pairings_unique_matchup"' <<<"$_insert_err"; then
	_ok "[DATA03] swapped-order duplicate insert correctly rejected by the DB"
else
	_no "[DATA03] swapped-order duplicate was NOT rejected (output: ${_insert_err:0:200})"
fi

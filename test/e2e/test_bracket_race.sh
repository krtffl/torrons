#!/usr/bin/env bash
# CONFIRMED FIX (D3): concurrent bracket-match votes that together close a round
# used to race — Handler.bracketMatchVote read the bracket row with a plain
# SELECT (no lock) before deciding whether to resolve/cascade the round, so two
# (or more) concurrent votes on different still-open matches could each see the
# round as "not yet fully voted" and skip advancing (a stall), or both try to
# cascade the same next round and hit BracketMatches' (BracketId,Round,Slot)
# unique index, rolling back the loser's transaction — including its own
# just-inserted vote — behind a 500.
#
# Fix: postgresBracketRepo.GetTx now takes a FOR UPDATE row lock on the bracket,
# serializing concurrent check-and-advance attempts on the SAME bracket. This
# test creates a real bracket, then fires one concurrent vote per round-1 match
# (closing the round all at once) and asserts every vote succeeds and the round
# advances exactly once.
section "BRACKET RACE — concurrent round-closing votes (D3)"

AUTH="Authorization: Bearer ${ADMIN_TOKEN}"

# Create a fresh bracket for class 1 (uses the e2e-campaign fixture from run.sh).
create_body="$(curl -s -X POST -H "$AUTH" -H "$(_xff)" "${BASE_URL}/bracket/1/create?size=8")"
bracket_id="$(python3 -c "import json,sys; print(json.loads(sys.argv[1]).get('id',''))" "$create_body" 2>/dev/null)"

if [[ -z "$bracket_id" ]]; then
	_no "[RACE2] could not create a bracket to test against (response: ${create_body:0:200})"
else
	# All round-1 real matches (both torrons present, i.e. not a bye — a bye is
	# already decided and wouldn't be open for voting).
	mapfile -t matches < <(PSQL -t -A -F'|' -c \
		"SELECT \"Id\",\"Torro1Id\" FROM \"BracketMatches\" WHERE \"BracketId\"='${bracket_id}' AND \"Round\"=1 AND \"Torro2Id\" IS NOT NULL ORDER BY \"Slot\";")

	if [[ "${#matches[@]}" -lt 2 ]]; then
		_no "[RACE2] expected at least 2 real round-1 matches to race, got ${#matches[@]}"
	else
		tmp="$(mktemp -d)"
		i=0
		for m in "${matches[@]}"; do
			mid="${m%%|*}"; winner="${m##*|}"
			i=$((i + 1))
			( curl -s -o "$tmp/body$i" -w '%{http_code}' -H "$(_xff)" -X POST \
				"${BASE_URL}/bracket/match/${mid}/vote?winner=${winner}" > "$tmp/code$i" ) &
		done
		wait

		ok=0
		for f in "$tmp"/code*; do
			[[ "$(cat "$f")" == "200" ]] && ok=$((ok + 1))
		done

		if [[ "$ok" == "${#matches[@]}" ]]; then
			_ok "[RACE2] all ${#matches[@]} concurrent round-closing votes returned 200"
		else
			_no "[RACE2] only $ok/${#matches[@]} concurrent votes returned 200 (lost vote / 500 under the race)"
		fi

		vote_rows="$(PSQL -t -A -c "SELECT count(*) FROM \"BracketMatchVotes\" v JOIN \"BracketMatches\" m ON v.\"MatchId\"=m.\"Id\" WHERE m.\"BracketId\"='${bracket_id}'" | tr -d '[:space:]')"
		if [[ "$vote_rows" == "${#matches[@]}" ]]; then
			_ok "[RACE3] all ${#matches[@]} votes recorded in BracketMatchVotes (none lost to a rollback)"
		else
			_no "[RACE3] expected ${#matches[@]} recorded votes, found $vote_rows"
		fi

		round="$(PSQL -t -A -c "SELECT \"CurrentRound\" FROM \"Brackets\" WHERE \"Id\"='${bracket_id}'" | tr -d '[:space:]')"
		if [[ "$round" == "2" ]]; then
			_ok "[RACE4] round advanced exactly once (1 -> 2) despite the concurrent close"
		else
			_no "[RACE4] expected CurrentRound=2 after all round-1 matches voted concurrently, got '$round' (stall or double-advance)"
		fi

		rm -rf "$tmp"
	fi
fi

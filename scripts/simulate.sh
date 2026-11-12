#!/usr/bin/env bash
# ------------------------------------------------------------------
#  Draft simulation: creates users + a league, fills it, starts the
#  draft, then lets the clock run out so the auto-pick worker picks.
#
#  Usage:   make simulate                 (server must be running)
#  Options: BASE_URL=http://localhost:8080  SIZE=3  TIMEOUT=5  KEEP=1
#           KEEP=1 keeps the demo data instead of deleting it at the end.
# ------------------------------------------------------------------
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
SIZE="${SIZE:-3}"
TIMEOUT="${TIMEOUT:-5}"
KEEP="${KEEP:-0}"

# ---------- colors ----------
R=$'\033[31m'; G=$'\033[32m'; Y=$'\033[33m'; B=$'\033[34m'; M=$'\033[35m'; C=$'\033[36m'
BOLD=$'\033[1m'; DIM=$'\033[2m'; X=$'\033[0m'

step()  { printf "\n${BOLD}${B}▶ %s${X}\n" "$1"; }
ok()    { printf "  ${G}✔${X} %s\n" "$1"; }
bad()   { printf "  ${R}✘${X} %s\n" "$1"; }
info()  { printf "  ${DIM}%s${X}\n" "$1"; }

need() { command -v "$1" >/dev/null || { bad "'$1' is required"; exit 1; }; }
need curl; need jq

# api METHOD PATH [JSON]  -> prints body, sets $CODE
api() {
  local out
  out=$(curl -s -w $'\n%{http_code}' -X "$1" "$BASE_URL$2" ${3:+-H 'Content-Type: application/json' -d "$3"})
  CODE="${out##*$'\n'}"
  BODY="${out%$'\n'*}"
}

# ---------- 0. server up? ----------
printf "${BOLD}${M}"
cat <<'BANNER'
   ____             __ _     ____  _      _
  |  _ \ _ __ __ _ / _| |_  |  _ \(_) ___| | __
  | | | | '__/ _` | |_| __| | |_) | |/ __| |/ /
  | |_| | | | (_| |  _| |_  |  __/| | (__|   <
  |____/|_|  \__,_|_|  \__| |_|   |_|\___|_|\_\   auto-pick simulation
BANNER
printf "${X}"
info "server: $BASE_URL   league size: $SIZE   pick timeout: ${TIMEOUT}s"

if ! curl -s -o /dev/null "$BASE_URL/users"; then
  bad "API is not reachable at $BASE_URL. Start it with: make run"
  exit 1
fi

USERS=(); LEAGUE=""
cleanup() {
  [ "$KEEP" = "1" ] && { info "KEEP=1 -> demo data kept (league $LEAGUE)"; return; }
  [ -n "$LEAGUE" ] && curl -s -o /dev/null -X DELETE "$BASE_URL/leagues/$LEAGUE"
  for u in "${USERS[@]:-}"; do [ -n "$u" ] && curl -s -o /dev/null -X DELETE "$BASE_URL/users/$u"; done
  info "demo data cleaned up"
}
trap cleanup EXIT

# ---------- 1. users ----------
NAMES=("Aarav Mehta" "Diya Kapoor" "Kabir Singh" "Meera Iyer" "Rohan Das" "Anya Rao" "Vihaan Shah" "Isha Nair")
step "Creating $((SIZE + 1)) users (one extra to show a full league)"
for ((i = 0; i <= SIZE; i++)); do
  n="${NAMES[$((i % ${#NAMES[@]}))]}"
  api POST /users "{\"first_name\":\"${n% *}\",\"last_name\":\"${n#* }\",\"age\":$((21 + i))}"
  id=$(jq -r .id <<<"$BODY"); USERS+=("$id")
  ok "user #$id  $n"
done

# ---------- 2. league ----------
step "Creating a draft league"
api POST /leagues "{\"name\":\"Simulation Cup\",\"description\":\"auto-pick demo\",\"type\":\"draft_pick\",\"series_name\":\"IPL 2026\",\"size\":$SIZE,\"timeout\":$TIMEOUT}"
LEAGUE=$(jq -r .id <<<"$BODY")
ok "league #$LEAGUE  $(jq -r '"\(.name) · \(.type) · size \(.size) · \(.timeout)s per pick · status \(.status)"' <<<"$BODY")"

# ---------- 3. join ----------
step "Users join the league (pick_order is assigned automatically)"
for ((i = 0; i < SIZE; i++)); do
  api POST "/leagues/$LEAGUE/entries" "{\"user_id\":${USERS[$i]}}"
  id=$(jq -r .id <<<"$BODY")
  ok "entry #$id  user #${USERS[$i]}  → pick_order ${C}$(jq -r .pick_order <<<"$BODY")${X}"
done

api POST "/leagues/$LEAGUE/entries" "{\"user_id\":${USERS[$SIZE]}}"
[ "$CODE" = "409" ] && bad "user #${USERS[$SIZE]} tried to join → ${Y}$CODE${X} $(jq -r .error <<<"$BODY")  ${DIM}(expected)${X}"

# ---------- 4. start ----------
step "Starting the draft"
api POST "/leagues/$LEAGUE/start"
FIRST=$(jq -r .current_pick.entry_id <<<"$BODY")
ok "status → ${G}$(jq -r .league.status <<<"$BODY")${X}"
ok "entry #$FIRST is on the clock until $(jq -r .current_pick.expires_at <<<"$BODY")"

api POST "/leagues/$LEAGUE/start"
bad "start again → ${Y}$CODE${X} $(jq -r .error <<<"$BODY")  ${DIM}(expected)${X}"

# ---------- 5. clock ----------
step "Entry #$FIRST does nothing… ⏳"
for ((s = TIMEOUT; s > 0; s--)); do
  filled=$(( (TIMEOUT - s) * 30 / TIMEOUT )); empty=$(( 30 - filled ))
  printf "\r  ${Y}[%s%s]${X} %2ss left " "$(printf '█%.0s' $(seq 1 $filled) 2>/dev/null)" "$(printf '░%.0s' $(seq 1 $empty))" "$s"
  sleep 1
done
printf "\r  ${R}[%s]${X}  time's up!   \n" "$(printf '█%.0s' $(seq 1 30))"

# ---------- 6. worker ----------
step "Waiting for the auto-pick worker"
for ((t = 0; t < 20; t++)); do
  api GET "/entries/$FIRST/roster"
  if [ "$(jq '.players | length' <<<"$BODY")" -gt 0 ]; then
    ok "🤖 auto-picked: ${BOLD}$(jq -r '.players[0] | "\(.name) (\(.position))"' <<<"$BODY")${X} for entry #$FIRST"
    break
  fi
  printf "  ${DIM}.${X}"; sleep 1
done
[ "$(jq '.players | length' <<<"$BODY")" -gt 0 ] || { bad "worker did not pick within 20s - is it running?"; exit 1; }

# ---------- 7. rosters ----------
step "League rosters"
api GET "/leagues/$LEAGUE/rosters"
printf "  ${BOLD}%-8s %-8s %-6s %s${X}\n" "ENTRY" "USER" "ORDER" "PLAYERS"
jq -r '.[] | [.entry_id, .user_id, .pick_order, (if (.players|length)==0 then "—" else ([.players[] | "\(.name) (\(.position))"] | join(", ")) end)] | @tsv' <<<"$BODY" |
  while IFS=$'\t' read -r e u o p; do printf "  %-8s %-8s %-6s %s\n" "#$e" "#$u" "$o" "$p"; done

printf "\n${BOLD}${G}✔ Simulation complete${X}\n"

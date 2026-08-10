# HANDOFF 2026-08-09b — webdesign.uk: SITE + CHAT SERVICE both LIVE; next is Phase 5 (the input box)

**Start here cold. Supersedes `HANDOFF_2026-08-09_continue_here.md`** (the `b`
one — same day, Phase 4 went from "blocked on a key" to "built, deployed,
proven live" in one session). Read with: `PLAN_2026-08-04_webdesign_uk_vm_hosting.md`
**§4 Phase 5** (your brief now) · `RUNBOOK` "Phase 4 — the chat service" (build/
deploy/verify commands, all proven, not theoretical) · `NOTES` 08-09 (6) (the
build, the bug caught before shipping, the mutation proofs) · `box/chat-service/`
(the actual source — 7 files, ~600 lines, read it before extending it).

## 0. One paragraph

The five-page site is built, reviewed, and parked ("ok for now"). **The chat
service (Phase 4) is now ALSO built, deployed, and proven live through the real
tunnel** — a stdlib-only Go service on `127.0.0.1:8081`, all five §5.1 controls
present and the two hard gates (turn cap, daily spend ceiling) mutation-proven.
A real message went in through `https://preview.webdesign.uk/api/chat` and got
a real Haiku 4.5 reply, correctly logged at the correct cost. **The next job is
Phase 5: the input-box page section** — a pinned section in the site plan,
posting same-origin to `/api/chat`, which is what the 9 parked `unresolved_cta`
items are waiting for. After that: owner review of the whole thing, then Phase
6 cutover.

## 1. THE DRIVE LOOP — still needed for Phase 5's pipeline half

Nothing in the pipeline self-drives (queue starved, no CronJob — unchanged from
08-08/09). Per stage: **find the item → claim it → orchestrate its
`handler_agent` directly → verify by payload → complete it**.

```sql
SELECT id, item_type, handler_agent, status FROM site_work_items
WHERE site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f' AND status='triaged';
UPDATE site_work_items SET status='claimed', updated_at=now() WHERE id='<id>' AND status='triaged';
```
Envelope: `{"action":"orchestrate","config":{"agent_type":"<handler_agent>"},
"input_data":{"domain":"webdesign.uk","site_id":"1fcfa4f3-…","work_item_id":"<id>",
"item_type":"<item_type>","source":"<source>","spec":<spec>,"current_page":<spec>}}`
via kcat (`-P -c 1`, topic `system.agent.generic.requests`, fresh UUIDs,
`client_id=demo_client`). Verify by payload: `SELECT status, current_step FROM
orchestration_states WHERE correlation_id='<corr>' AND parent_orchestration_id
IS NULL;` — kcat exit 0 proves nothing, and filter `parent_orchestration_id IS
NULL` or child spawns inflate every count. Worked examples throughout NOTES
08-08/09.

## 2. What is DONE — verified 2026-08-09, do not redo

**The site**: 5 pages, all assets, 0 ban hits, capped terms stated, provenance
proven. See the 08-09(a) handoff for the full review-round account if you need
the history; it is not repeated here.

**The chat service (Phase 4)** — everything in this section is LIVE right now,
not planned:
- `box/chat-service/*.go` — `clientip.go` (CF-Connecting-IP, NOT idea.uk's
  X-Real-IP pattern — this box is tunnel-fronted, idea.uk isn't), `ratelimit.go`
  (per-IP sliding window), `store.go` (JSON-file state + two JSONL append logs),
  `claude.go` (raw HTTP call to Haiku 4.5, no SDK), `chat.go` (the five-gate
  handler), `health.go`, `main.go` (fails loudly at startup — no key, no
  contact details, no boot).
- Deployed: `systemd` unit `webdesign-chat.service` (sandboxed, runs as
  `www-data`), nginx wired (`/api/chat`, `/health` → `127.0.0.1:8081`;
  `/stripe/webhook` deliberately NOT wired — later phase).
- **Proven, not assumed**: `/health` green through nginx AND direct. A real
  chat turn through `preview.webdesign.uk` got a real, on-brief reply, logged
  at $0.000427 — exact arithmetic match to Haiku's $1/$5-per-MTok pricing. The
  logged `client_ip` matched my own externally-confirmed IPv6 exactly (proves
  the header is genuine, not a stuck constant).
- **Both hard gates (turn cap, spend ceiling) MUTATION-proven**: each
  condition was neutralized in turn, the corresponding test correctly failed,
  then reverted. Not just "tests pass" — the tests were shown capable of
  catching the regression they exist to catch.
- **OWED, not done**: the full two-network CF-Connecting-IP proof
  (`bugs_open/139` shape — `count(DISTINCT ip) > 1`). One proof point exists
  (my own IP matched); a second request from a genuinely different network
  (phone on mobile data, anything) would close it. RUNBOOK has the two-curl
  recipe. Not urgent — the single-point proof already rules out the dangerous
  failure modes (constant, loopback, tunnel's own address) — but worth closing
  when convenient.

## 3. NEXT: Phase 5 — the input box

**Brief: PLAN §4 Phase 5.** What's already decided:

- A **pinned section** in the site plan (roadmap `section_types` — the
  planner's own mechanism, already proven working for hero/CTA/etc sections
  this lane has used all week), whose markup POSTs same-origin to `/api/chat`.
- **CTS-044 generation guards**: external loader file, no inline script,
  `data-runtime-fill` shell. Read CTS-044 before writing the component if you
  haven't touched a generated-site component before — this repo's framework
  has its own conventions for how JS reaches a page.
- **The pairing "this site has `capabilities:["backend"]`, this section is
  pinned on this site" is the whole safety story** (measured 2026-08-04 in the
  original plan) — write it in the site's `site_config` notes so the next
  planner run has it in context. Don't try to express "this component needs a
  backend" in the component library itself — CTS-049's gate has no column to
  hang that on.
- **Landing it resolves the 9 parked `unresolved_cta` `needs_human_review`
  items** — their hero/CTA slots have no real destination today; the input box
  IS that destination. Check them after Phase 5 ships:
  `SELECT id, left(summary,80) FROM site_work_items WHERE site_id='1fcfa4f3-…'
  AND item_type='unresolved_cta';`
- The chat service itself needs no further work to support this — `/api/chat`
  already accepts `{"conversation_id": "...", "message": "..."}` and returns
  `{"conversation_id": "...", "reply": "..."}`; `conversation_id` empty on
  first call, echo the returned one on subsequent calls from the same page
  session (e.g. held in a JS variable — no cookie needed for a single-page
  conversation).

## 4. Traps earned in this lane — carried forward, still current

- **`evidence_base.facts[]` is bookkeeping; `writer_block` is the WIRE** for
  the SITE writer. Note: the chat service's system prompt (`chat.go`'s
  `systemPromptFacts`) is a SEPARATE, hand-copied string with NO code link to
  `evidence_base` at all — if the owner changes the price/terms/contact later,
  BOTH the site's `writer_block` AND this Go constant need updating, by hand,
  by whoever makes the change. Flagged in the file's own comment; not yet a
  LANDMINES entry because it hasn't bitten anyone — it will, eventually, if
  this note is lost. Consider filing it once there's a live incident, per the
  "landmine needs a real trap, not just a suspicion" bar.
- **A page REBUILD resets the hero binding AND regenerates the TITLE** (banned
  em dash returns). After ANY page rebuild: re-check both. Unrelated to Phase
  5 unless Phase 5's pinned section triggers a rebuild of the pages carrying
  it — check when it happens.
- **A visitor check MUST extract `url(...)` from style attributes.**
  `verify_served_site.sh` does. Re-run it after Phase 5 ships the input box —
  a new component is exactly the kind of change that check exists for.
- **`asset-deployer` is DOWN** ("storage client not available") — no favicon
  yet, platform-side fix owed, not lane work.
- idea.uk's nginx snippets (`proxy_tool.conf`) set `X-Real-IP $remote_addr` —
  **do not copy that onto this box** if extending the chat service's nginx
  config. This box is tunnel-fronted; idea.uk isn't. See `clientip.go`'s own
  comment and the nginx file's own warning block.

## 5. Owner ledger

**Owed by the owner:** correction-fee number · written terms before live
Stripe · final review + cutover approval (now covers the chat box too, once
Phase 5 lands).
**Settled, do not reopen:** £1,200, no VAT · 2 revision rounds, 14-day review
window · Sonnet 5 on the site writer · Haiku 4.5 on the chat intake ·
framework-only builds · one box per trust class · the site itself is "ok for
now" · **the scoped Anthropic key (created 08-09, workspace
`webdesign-uk-chat`, live on the box)**.

## 6. Access map

`~/.ssh/webdesign_box_ed25519` → root@box · `~/.config/cloudflare/token`
(works, expires 09-01) · `gh` = gqls, ADMIN on `vm-sites` · `b2` CLI · kubectl
→ cluster/DB (`site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f'`) ·
`/etc/webdesign-chat.env` on the box (600, root:root — the chat service's
secrets; never committed, never read into a chat transcript). **No Mythic
panel, no Stripe, no Cloudflare dashboard, no Anthropic Console access from
this session** (the owner created the key directly).

---

> **CORRECTION APPENDED 2026-08-10 by the bugfix_239 lane (not this lane's thread — nothing
> above has been edited).**
>
> **The drive-loop `kcat` recipe in this file will silently do nothing if the JSON reaches
> stdin on more than one line.** `kcat -P` publishes **one message per line**, applying the
> same `-H` headers to each, so a pretty-printed heredoc envelope arrives as four to six
> invalid-JSON fragments — and the chassis (pre-fix) ran each one as the `generic` no-op and
> reported `COMPLETED` with an empty `execution_path`. That is the whole of `bugs_open/239`,
> whose root cause was found today from `chassis_intake_events` payload bytes: 8 of 8
> single-message sends resolved correctly, 10 of 10 fragmented sends did not.
>
> **So the `source`+`spec` trigger this lane bisected, and the "omit `source`" workaround,
> are both superseded** — omitting `source` worked only because it shortened the envelope
> enough to fit on one line.
>
> **Send it on ONE line:** `kcat ... <<<'{"action":"orchestrate",...}'` (here-string, single
> quotes), or pipe through `jq -c`. Never `<<JSON … JSON`.
>
> **Verify what arrived, not what you meant to send:**
> ```sql
> SELECT left(correlation_id::text,8), count(*) AS msgs FROM chassis_intake_events
> WHERE kind='request' AND correlation_id::text LIKE '<corr>%' GROUP BY 1;
> ```
> `msgs > 1` ⇒ fragments. And keep this lane's own rule: check `owner_agent_type` equals the
> agent you asked for, never just `status`.
>
> The chassis half is fixed (fail-closed refusal, `DISPATCH_FAIL_CLOSED`) but **inert until
> the fleet rolls** — `strings /app/agent-chassis | grep -c DISPATCH_FAIL_CLOSED` says
> whether it is live yet. The send-side rule above applies either way.

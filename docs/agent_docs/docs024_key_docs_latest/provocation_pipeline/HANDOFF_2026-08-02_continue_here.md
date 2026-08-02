# HANDOFF — provocation pipeline, 2026-08-02

**Supersedes `HANDOFF_2026-07-31_continue_here.md`** (still worth reading for the
design reasoning and the owner's settled decisions; this file supersedes only its
"what to do next" section).

---

## State in one paragraph

The mechanism is **live, enabled and proven on both paths**. vonc.com's
provocation feed is now published by the platform: a `provocations` pool table, a
`render_provocation_feed` Go action, an agent and a 6-hourly schedule. There is
**no code left to write for rotation**. The site is nonetheless still making a
false claim — "Today's Provocation" over an entry dated 26 Jul — because **the
pool holds nothing newer than 26 Jul**. The remaining gap is content and only
content. Adding provocations is `INSERT`s: no code, no image, no deploy.

## What is true right now (all verified 2026-08-02, evening)

| thing | state | how it was checked |
|---|---|---|
| `render_provocation_feed` in the binary | **yes, both replicas** (v1.0.1229) | `strings /app/agent-chassis \| grep -c` = 1; positive control `deploy_image_asset` = 5; synthetic negative = 0 |
| `provocation-feed-refresh` schedule | **enabled**, 21600s | `SELECT enabled FROM scheduled_tasks` |
| skip path | **proven live** | fired 18:57Z in ~1.4s → `committed:false`, `reason:"no change since the served feed"`, 8 archive entries |
| commit path | **proven live**, induced deliberately | `a1bf37d55` in `gqls/sites` |
| pool | 9 rows, newest `publish_on` = **2026-07-26** | `SELECT … FROM provocations WHERE domain='vonc.com'` |
| served feed | `today.slug = nobody-wants-personalised-internet` | `curl https://vonc.com/data/provocations.json` |

## The one thing that is fixed but NOT live

Commit **`59f3c67dd`** — `marshalFeedFile` with `SetEscapeHTML(false)`.

The action shipped HTML-escaped markup: headlines carrying `<em>` reached the file
as the escaped six-character form. Every parser decodes it, which is why it was
invisible to the golden-fixture parity test — **that test unmarshals both sides
before comparing, so an encoding difference cannot reach it by construction.** It
was found by diffing the first artefact the action committed against the one the
Python oracle had written.

**It needs a chassis roll.** Not urgent — nothing is broken, the escaped form
renders correctly — but until it rolls, the file carries escaped markup on the
hand-edited fallback path. Any session's roll ships it.

Also shipped and **deliberately not fixed**: Go sorts map keys, so the file's
top-level order is alphabetical where the oracle wrote a human order. That cost is
paid once per *change of writer*, not once per day. Pinning it needs a second
hand-maintained copy of every object's field list — a worse trade. Recorded in
VONC-011's open-question field with the ping-pong caveat.

## What to do next — in order

### 1. Content (the whole remaining gap) — OWNER'S CALL, do not do this unasked

The site claims a daily provocation and serves a week-old one. Fixing that means
writing provocations. **They publish as the owner's opinions under his name, so a
session must not invent them.** What a session CAN do once given text:

```sql
INSERT INTO provocations (domain, slug, category, publish_on, status, title, teaser,
                          card_desc, detail_body, headline, body, source)
VALUES ('vonc.com', '<slug>', 'general', '<YYYY-MM-DD>', 'approved',
        '<title>', '<teaser>', '<card_desc>', '<detail_body>', '<headline>', '<body>', 'human');
```

Then either wait up to 6h for the schedule, or make it due:
`UPDATE scheduled_tasks SET last_triggered_at=NULL, last_completed_at=NULL WHERE name='provocation-feed-refresh';`

The partial unique index makes two approved provocations on one date
unrepresentable — you will get a constraint violation, not a silent overwrite.

### 2. Verify the escaping fix after the next roll

```bash
strings /app/agent-chassis | grep -c marshalFeedFile      # symbol may be inlined — prefer the artefact:
gh api "repos/gqls/sites/commits/<next-sha>" --jq '.files[] | "+\(.additions) -\(.deletions)"'
```

After the roll, the **next** commit should be a small diff, and the file should
carry literal `<em>`. Note the first post-roll commit will again be large (the
escaping change touches every escaped line) — that is expected, once.

### 3. The generative half (PLAN §4/§10) — unbuilt, unchanged

Grok generates candidate provocations; a gate checks safe / interesting /
current / good-provocation-properties. The pool's `source`, `source_ref`,
`gate_verdict` and `gated_at` columns exist for exactly this and are unused. The
owner has settled: **no human approval of publishes** — do not re-open that.

### 4. Categories, then paired mode

Categories break the one-`today`-per-site engine contract (`round.go` reads a
single `today`), so that is an engine change, not a feed change. Paired mode needs
identity. Both are in the PLAN; neither is started.

## Traps this lane has already paid for

- **The feed is read SERVER-SIDE.** `internal/tools-api/handlers/round.go`
  `FetchProvocation` takes the whole `today` object and persists it as the round's
  provocation. `today.headline/body/slug/date` are load-bearing for the GAME.
  Never "seal" today's provocation by emptying them — `checkFeed` refuses to emit
  a feed that does, in both directions.
- **A structural test cannot see an encoding.** See `LANDMINES.md`; it is why the
  parity test was green for a week while the artefact diverged on every line.
- **Never TYPE a backslash-u escape sequence**, in Go source or in prose about it.
  The Go compiler decodes it inside double quotes and several tool channels decode
  it in transit — an assertion that a sequence is ABSENT silently becomes a search
  for the character it decodes to. Build it by concatenation. It fired **four
  times in one session**, including inside the paragraph warning about it.
- **`cd` persists between tool calls.** A relative-path search resolved against a
  stale cwd returns a confident, wrong absence. Use absolute paths.
- **A roll is not evidence.** Grep the running binary with a positive control. Note
  this change was purely additive, so the ideal negative control (a string it
  REMOVED) does not exist — establish provenance separately with
  `git merge-base --is-ancestor <commit> HEAD`.

## Where everything lives

- Action: `platform/orchestration/actions/provocation_feed_action.go`
- Tests: `provocation_feed_action_test.go` (15, ports `verify_rotation.py`),
  `provocation_feed_parity_test.go` (parity + its can-fail control + the new
  byte-level encoding test)
- Migrations: `docs/agent_docs/sql_for_agents/282_provocation_pool.sql`, `283_provocation_feed_publisher.sql`
- Oracle / manual fallback: `provocation_pipeline/builder/build_provocations.py`
- Register: VONC-011 (this mechanism), VONC-002 (the wider pipeline), VONC-003 (the contract)
- Council: round 2 APPROVED, correlation `6612dc0b-8e03-4039-a8c8-fe4fabaaddeb`

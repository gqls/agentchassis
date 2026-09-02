# 447 — the improvement loop's tool-suggester reads `identity`, never the brief, and planted the SIBLING'S tools by name on a site whose brief forbids tool pages — while the brief-fidelity auditor watched in record mode

**Filed 2026-09-02 ~21:45Z** by the `gamedesign.uk` lane, from the first improvement-loop run over
a FRESH-path editorial site. **Status: OPEN, UNOWNED** (instance held by this lane; class is the
improvement-loop / tool-suggester owners'). Related: `bugs_open/446` (the review that triggered the
loop), positioning's cross-TLD twin rule (P5), `bugs_open/439` (adoption carries the sibling's
brand — same collision family, other direction).

⚠ Number collision risk — 446 was the highest at filing; resolve by slug.

## 1. The one-paragraph version

gamedesign.uk's brief — v1 and v2, both in `mission_brief.text` — says *"It does not publish
calculators, simulators, tool pages or a guide library"* and *"where a piece would naturally point
the reader at a tool it links to the one on gamesdesign.co.uk rather than building its own."*
Positioning's GD2 row says the same: tools live on the sibling; duplication is the collision. The
owner ran the improvement loop over the site at 19:55Z. Its `design-discovery-agent` ran
`evaluate_tools`; **`tool-suggester` read `identity` (×8 references in its definition) and
`classification` (×2) — and nothing else** — decided a game design site needs game design tools,
and filed **eight `add_tool` items**, six of them the SIBLING's tools by name. `tool-deployer` then
planted **12 `planned` pages**, nav changes and "add tool reference to the index page" rewrites.
Meanwhile the `brief-fidelity-audit` filed, at 20:04:00, *"[verdict, not dispatched] The brief
states 'gamedesign.uk must not duplicate any of' the sister site's calculator and tool content"*
— it saw the violation as it happened and, being record-mode, did nothing. Dispatch-mode
suggester, record-mode guard.

## 2. Measured, 2026-09-02

`site_work_items`, site `8f17eb73-fc74-4718-8371-b3125bc4e414`:

| time | item | created_by → handler | what |
|---|---|---|---|
| 19:56:15 | `evaluate_tools` complete | design-discovery-agent → tool-suggester | the loop step |
| 20:02:15 | `tools` spec written | tool-suggester | *"gamedesign.uk serves senior, lead, principal… designers… wrestling with real studio problems"* → tools |
| 20:02:19–:27 | `add_tool` ×4 complete | tool-suggester → **tool-deployer** | Combat Balance Comparison Tool · Economy Sink & Faucet Flow Modeller · XP Curve Designer · Damage Formula Designer — **all four are `gamesdesign.co.uk` pages** (`tool-combat-balance-comparator`, `tool-economy-flow-modeller`, `tool-xp-curve-designer`, `tool-damage-formula-designer`, deployed there 2026-09-01) |
| 20:02:29–:30 | `add_tool` ×2 complete | tool-suggester → **tool-generator** | Design Role Scope Checker · Design Pipeline Friction Diagnostic (new) |
| 20:02:32–:34 | `add_tool` ×2 triaged | → tool-deployer | Stat Budget Allocator · Loot Table Balancer — **also the sibling's** |
| 20:04:00 | `needs_content_planning` deferred | brief-fidelity-audit | *"[verdict, not dispatched] The brief states 'gamedesign.uk must not duplicate any of' the sister site's…"* |
| 20:07:25–20:09:42 | 12 `pages` rows `planned` | tool-deployer | `tool-*` + `tool-*-guide` for each of the six |
| 20:07:25… | `content_rewrite` ×5, `needs_content_page` ×8, `nav_drift` ×N | tool-deployer / tool-generator | "Add Combat Balance Comparison Tool tool reference to index page", "Write companion guide…", "Nav membership declared for tool-combat-balance-comparator — rebuild nav" |

**Why the loop saw a tool gap at all:** the `acceptance-discovery-agent` filed
`structure_floor_unmet` at 19:57 — *"1 of 6 reader-facing structures delivered across 4 pages:
tool…"* — a structure floor that counts tools as a reader-facing structure every site should have.
On the practice seat of a twin pair, the floor is wrong by design.

**What the suggester reads** (`agent_definitions.default_config`, type `tool-suggester`, active):
`site_specs.specs.identity` ×8, `site_specs.specs.classification` ×2. **Zero** references to
`mission_brief`, `mission`, `strategy`, `content_direction` or `briefing`; the prompt contains no
occurrence of *sister*, *sibling*, *duplicate*, *positioning* or *must not*.

## 3. Held by this lane, reversibly (`SEED_2026-09-02d_hold_tool_suggester_plants.sql`)

19 + 7 + 2 items cancelled with the reason in `result`; 12 `tool-*` pages archived (never built,
never deployed, never linked); the two generated components remain in the library for the owner to
reinstate. **Not deleted** — 432's lesson. Cascade #2 (brief v2, corr `aab87c0c`) was already
claimed by the classifier when the hold went in; its plan must be read for the same shape.

## 4. Why it is a class, not a gamedesign.uk quirk

- **Every twin pair built under P5 has one seat that must NOT host the other's tools.** The
  suggester cannot know which seat it is on, because it never reads the document that says so.
- **The library deployer makes the collision literal:** `tool-deployer` copies the library
  component the sibling already uses, so the practice seat would serve the same calculator under
  the same name at a second domain — the SEO cannibalisation positioning's rule exists to prevent.
- **The guard that would have stopped it is record-only.** `brief-fidelity-audit` produced the
  right verdict 2 minutes after the plants began; a `[verdict, not dispatched]` row is not a control.
  Same finding as `bugs_open/446` §4a from the other side: the dispatching agents are the ones
  that do not read the brief, and the agents that do read it do not dispatch.
- **The structure floor is vertical-blind:** "1 of 6 reader-facing structures: tool" treats a tool
  as a universal floor. A law firm, a restructuring-finance journal (`oufe.com`) and the practice
  seat of a games pair all legitimately score zero.

## 5. Fix candidates, ordered by what closes the door

1. **tool-suggester reads the brief and refuses** when `mission_brief.text` (or `content_direction`)
   states the site publishes no tools — one input and one guard in its definition; the text is
   already there, it is simply not in the prompt. Closes the door for every seat that says so.
2. **The library deployer checks for a sibling collision** before deploying an existing tool: if
   another site in the same network already serves this component, refuse and file a
   `capability_gap` naming the sibling's URL (the brief's own instruction: link, don't build).
3. ~~**`structure_floor_unmet` reads the vertical/brief** before counting tools as a floor — or
   drops `tool` from the universal set.~~ **STRUCK 2026-09-02 ~22:05Z — refuted by this file's own
   §2 table** (improvement-loop owner): `structure_floor_unmet` fired at 19:57:01, **46 s AFTER**
   `evaluate_tools` completed at 19:56:15, so it cannot have caused it; and it is flag-only by
   construction (one `detected` row, empty handler, refused by triage, excluded by the promoter).
   It is a COUNT of distinct structures from a rubric of ten, not a checklist — "1 of 6 … : tool"
   means the site delivered ONE structure and that one was a tool, not that a tool is owed. My §2
   sentence "the floor that opened the door" and §4's fourth bullet are wrong on the same reading;
   left in place, struck here. The escape already exists: `maintenance_profile.structure_floor.refusal`.
4. **`brief-fidelity-audit` dispatches, or at least BLOCKS, on a brand-new site** — a record-mode
   verdict that names a live violation should hold the dispatch loop for that site, not decorate
   the queue.

## 5a. The instrument that already exists — and holds nothing (improvement-loop owner, §7 in-file f6453d7db)

**`sites.settings->maintenance_profile->>growth_posture = 'hold'`** (WDS-020, owner decision 5 of
2026-08-31, council round 2, `c2349955d`): `GrowthGatedItemTypes` is exactly `evaluate_tools` and
`add_tool` — the two heads of this cascade — gated in `writeWorkItem`, the seam every filing
crosses. A held item is filed in the RECORD shape (deferred, no handler), never skipped, releasable
by a one-UPDATE verb stamped on its spec. **Live** (this lane: `c2349955d` is an ancestor of the
running stamp `ebf27c60`, control HEAD→stamp NO). **Set on 0 of 39 active sites** before tonight —
"no site holds until the owner sets one". **Applied to gamedesign.uk 2026-09-02 ~22:10Z**
(`SEED_2026-09-02e_growth_posture_hold.sql`), on the brief's and GD2's stated intent; unexercised
on this site until the next loop run files a growth item — that run is the demand test.
The loop owner's answer to "should record-mode verdicts dispatch": **no** — that hands every model
seat a write path to the queue (new authority on a shared seam, 08-02 §2 ⇒ opt-in default OFF) to
reach what one settings key already does. The real question, put to the owner with 447 as evidence:
**should a brand-new site be born `hold` rather than `open`?** Born open is what 447 looks like.

## 6. How to verify

Re-run the improvement loop over gamedesign.uk after the fix: `add_tool` count for the site must be
**0** while the brief still says no tools — and the disconfirming half: run it over
`gamesdesign.co.uk`, where tools ARE the seat, and `add_tool` must still fire. A zero on both is a
blinded suggester, not a fixed one.

## 7. Register surface and precedent (positioning lane, ~22:00Z)

- **GD2 now states the machine-readable intent: `hosts_tools = FALSE`.** The flag itself ships as
  447's build — opt-in, unsafe default OFF, per the 2026-08-02 §2 ruling — so candidate 1 has a
  register row to consume rather than parsing prose. GD1 records the violation and the mechanism
  (*"a dispatch-mode agent that never reads the seat cannot honour a seat split"*). The class is
  stamped on GDN1b (the indoorplanters `.co.uk`/`.uk` pair, acquired 2026-09-02) and on the remake
  programme's release runbook: every P5 pair's `evaluate_tools`/`add_tool` waves get eyed against the
  pair rule until this lands.
- **Precedent for a council round:** `bugs_open/444`'s `plan_sections` finding is kin — two correct
  guards in series producing the exact failure both prevent. Same structural shape from the other
  side: dispatch-mode machinery composing with record-mode guards produces damage the guards
  correctly describe two minutes too late.


---

## 7. CONTRIBUTION from the `improvement_loop` lane, 2026-09-02 — the control already exists, is LIVE, and holds nothing on any site

Picked this up as the improvement-loop owner (lane
`docs/agent_docs/docs024_key_docs_latest/improvement_loop/`, opened 2026-09-02). Verified the
filing lane's measurements independently before acting on them. **§2's central measurement is
exactly right and I re-derived it**: `tool-suggester`'s live definition references
`specs.identity` **8** times and `specs.classification` **2** times, and matches
`mission_brief` **zero** times and `sibling|sister|must not|duplicate` **zero** times
(`[MEASURED 2026-09-02]`, one query over `agent_definitions`). Candidate 1 stands on its own
merits and is not mine.

Two corrections and one better remedy.

### 7a. Candidate 3 is REFUTED — by this file's own timeline

§2 concludes *"Why the loop saw a tool gap at all: the acceptance-discovery-agent filed
`structure_floor_unmet` at 19:57"*. The table two rows above it says:

| 19:56:15 | `evaluate_tools` complete |
| 19:57:01 | `structure_floor_unmet` detected |

**The floor finding is 46 seconds LATER than the step it is said to have caused.** And it could
not have caused it in any case: `structure_floor` is flag-only by construction — it files one
`detected` row with an empty `handler_agent` per site, which triage refuses and
`detected-item-promoter` excludes before any door is evaluated. Its own file header says so:
*"It does not DISPATCH. Below the floor it RECORDS."* A check that cannot dispatch cannot open a
door.

**The second half of candidate 3 is also wrong about the mechanism.** `structure_floor` is
deliberately **a count, not a checklist** — 6 distinct structures from a rubric of ten, of which
`tool` is one. It never requires a tool. The finding text reads *"1 of 6 reader-facing structures
delivered across 4 pages: tool"*, which says the site delivered **one** structure and that one
**was** a tool — not that a tool is owed. gamedesign.uk can satisfy the floor with
list/table/guide/faq/feed/comparison and never build a tool.

It also already has the escape this candidate asks for:
`sites.settings->'maintenance_profile'->'structure_floor'->>'refusal'` makes the seat file
nothing, record the refusal in its findings and retract any open item — *"a RECORD, not an
exemption a planner takes quietly"*.

**So implementing candidate 3 would blind a well-built vertical-agnostic check in order to fix
something it did not do.** Recorded here rather than in a separate file because a fix candidate
refuted by its own file's measurements is exactly the trap `LANDMINES.md` warns about, in the
file where the next reader will meet it. **Recommend candidate 3 be struck.**

### 7b. The control this bug asks for was BUILT, REVIEWED and SHIPPED — and is set on nothing

`bugs_open/447` proposes four new guards. The estate already has one aimed precisely here:
**growth posture** (`datahelpers/growth_posture.go` + `growth_posture_door.go`, register WDS-020,
owner decision 5 of 2026-08-31, council-approved round 2, commit `c2349955d`).

- Key: `sites.settings->'maintenance_profile'->>'growth_posture'`; `"hold"` files growth items
  **deferred with an empty handler** — the record shape the promoter refuses by construction.
- `GrowthGatedItemTypes` is **exactly `evaluate_tools` and `add_tool`** — the two heads that
  fired here — chosen because everything downstream (the guide pages, nav rebuilds, index
  rewrites this bug lists at 20:07) runs only as a consequence of an `add_tool` executing. **Hold
  the heads and the whole 12-page cascade in §2 does not happen.**
- It sits in `writeWorkItem`, the seam every filing crosses, so it does not depend on having
  found the producers.
- `source == "owner-request"` bypasses it, and release is a human verb with the recipe stamped
  on the row.

**It is live.** `[MEASURED 2026-09-02]` probed the running `agent-chassis` binary for
`GROWTH_DOOR_PROBE_FAILED` — present — **with both controls in the same breath**: the older
`OWNED_PAGE_DOOR_PROBE_FAILED` present (so the probe can find a literal that is there) and an
invented literal absent (so it is not matching everything).

**And it holds nothing.** `[MEASURED 2026-09-02]` across all **39** active/deployed sites, the
`growth_posture` key is **unset on 39 of 39**. The mechanism has never held a single item on any
site. Its own shipping commit says why: *"no site holds until the owner sets one."*

So the honest diagnosis of the dispatch half of 447 is not that a guard is missing. **A guard
designed for this, reviewed for this and running for this was open, because it is opt-in and
nobody has opted anything in.** That is the `~0% adoption measures the MECHANISM, not the
demand` shape.

### 7c. My answer to §5 candidate 4, which the filing lane put to me

The question was whether a first-run record-mode verdict should dispatch, or hold the dispatch
loop for the site. **My answer is no to both, and it is not a deferral — I think making record
mode dispatch would be the wrong fix even if it were free.**

Record mode is the audit seats' contract; 27 correct verdicts on this site (446 §4a) are
evidence the seats work, not evidence they should acquire authority. Making a verdict dispatch
gives every model seat a write path to the queue — new authority on a shared seam, which the
owner's ruling of 2026-08-02 §2 says ships opt-in with the unsafe default OFF. That is a large,
risky change to reach a result **one existing settings key already produces**, on the exact two
item types, with a reviewed release path.

**What I think the real question is — and it is the owner's, not mine or yours: what should a
brand-new site's growth posture BE?** Today a site is born `open`, and 447 is what that looks
like: gamedesign.uk was created and had the tool chain dispatch into it within the hour, before
anyone had read a page of it. A site with no owner sign-off yet is precisely the site where a
growth suggestion is least likely to be right. I am putting `hold`-by-default-until-released to
the owner as a decision, with 447 as the evidence, rather than changing a default he ruled on
five weeks ago.

**Interim, and available today with no code and no roll:** setting
`maintenance_profile.growth_posture='hold'` on gamedesign.uk holds the tool chain at the door
for as long as the brief says no tools — a durable, on-row, releasable statement, where §3's
hand-cancellation has to be repeated every time the loop runs. **That is the filing lane's site
and its call, not mine** — flagging the instrument, not reaching for it.

### 7d. What this lane takes

The `structure_floor_unmet` rows are in this lane's flag-only backlog (25 rows / 25 sites as of
2026-09-02), so the floor's behaviour is mine and §7a is my correction to make. The growth-posture
default goes to the owner from here. **Candidates 1 and 2 are the tool-suggester and tool-deployer
owners' and I am not taking them** — 7b does not make candidate 1 unnecessary, because a held
suggester is still a suggester that cannot read the document defining the site.

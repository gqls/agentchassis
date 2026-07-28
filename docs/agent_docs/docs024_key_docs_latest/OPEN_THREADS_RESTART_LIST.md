# Open threads — restart list

**Rewritten 2026-07-27, 16:05 UTC.** (The previous version of this file was dated
2026-07-22 and its bug table had gone badly stale — it listed bugs as open that
had since closed and predated about forty newer ones. That table is not
reproduced here; see "About the bug backlog" at the bottom for why.)

> **REFRESHED 2026-07-27, 22:00 UTC (23:00 BST).** Six hours after the rewrite
> above, and the tree moved further in those six hours than the previous rewrite
> covered in five days: **~190 commits, four chassis rolls (v1.0.1175 → 1179),
> eight bugs closed and eleven filed.** The refresh touched the standing facts,
> §0 (which is a **different defect** — the old one is fixed and closed), the
> rows for the twelve threads that committed this evening, and five §C items the
> owner has since ruled on. **Rows not listed as changed were not re-checked
> tonight** — they were accurate at 16:05 UTC and their threads were quiet since.
> Where this refresh corrects the 16:05 text, the correction is marked inline
> rather than edited away.

**What this file is for.** The owner is about to start a run of fresh sessions to
clear backlog. For each thread that is still alive, this says: where it stopped,
the *one* thing to do next, and which document to open first. Threads that a
session simply cannot move — because they need the owner personally — are kept in
their own section so they are not picked up by mistake.

**How the sections work**

- **§0 Do this first** — one live, armed defect that will start breaking page
  builds across the estate this evening if nobody touches it.
- **§A Ready to resume** — a session can open the cold-start doc and start working.
- **§B Waiting on another thread or a bug** — the work is understood; something
  else has to move first.
- **§C Waiting on the owner** — a decision or a physical action only he can take.
  Each entry states the decision in one line.
- **§D Parked on purpose** — do not restart these without an explicit go.

**Three standing facts, checked today**

- ~~The fleet is on **agent-chassis `v1.0.1174`**, rolled at 15:11:15 UTC.~~
  **CORRECTED 22:00 UTC — the fleet is on `v1.0.1179`.** Four more rolls went out
  during the evening: 1176, 1177 (`403f67920`, five bugs_open fixes), 1178
  (content-creator + the island rebuild for 083), 1179. Verified:
  `kubectl -n ai-persona-system get deploy agent-chassis -o jsonpath='{...image}'`
  → `docker.io/aqls/agent-chassis:v1.0.1179`. **Everything marked "roll-unblocked"
  below is still true and then some** — but note the corollary that bit a thread
  tonight: **a roll kills in-flight council runs.** One session rolled the chassis
  and destroyed its own council round, and `EXECUTING_STEP` hid it for an hour
  (`64ae96cd2`, logged in `WRONG_CALLS.md`). On a tree rolling four times an
  evening, check for your own in-flight runs before you build.
- **Clock trap.** This machine runs BST (UTC+1). `git log` prints BST; `kubectl`
  prints UTC. Compare in UTC: `TZ=UTC git log --date=iso-local`. Several docs
  written today have their commit times an hour ahead of the image record for
  this reason.
- **Being in the binary is not the same as working.** Two threads have now been
  bitten by a clean pod-grep on a feature that still did not work, because the
  code was in the binary but not on the feature's path. Verify by making the
  thing happen, and by making the *failing* case happen.

**Before you act on any row:** re-run `git status` and `git log` (the tree is
shared and moves within minutes), and run `scripts/who-owns.py <bug>` before
routing work at any bug — several rows here name bugs that other threads already
own.

---

## §0 — Do this first

> **The 16:05 §0 is CLOSED.** It was `bugs_open/112` — spawned pods getting no
> `GEMINI_API_KEY`, predicted to take down every page build from ~20:25 UTC. It
> was fixed in **both** spawners (the second was found during the fix, not
> before), rolled, and **closed at 21:54 BST after both acceptance tests passed
> on the real path** (`60bdd15a3`); council round 2 approved and pinned to the
> committed plan (`df856872b`). **No owner action is needed and the §C item
> asking for a revert decision is struck.** The predicted casualty never
> happened. Below is a *different* defect.

### Nothing checks content-creator's copy for invented statistics, and it is inventing them

`bugs_open/123`, filed 20:43 BST by the Gemini thread, **OPEN and unowned**,
**HIGH**. This is the one to read first because it is a **publication gate**, not
a repair job.

A live `content-creator-agent` run on `gemini-pro-latest` produced, unprompted:

> *"Industry data shows that large language models experience hallucination rates
> between 3% and 10% depending on the task."*

No source, invented range, and phrased as sourced. **Nothing in the pipeline
objected, because nothing in the pipeline looks** — the claims assessor cannot be
pointed at this output at all, since its input contract cannot be satisfied. This
is not a forgotten call; it is a contract mismatch. Family: `bugs_closed/043`
(the fabricated-stats lane) and `bugs_open/102` (the claims layer is
`page_type`-blind).

**Why it is more exposed tonight than this morning:** the **house voice went
fleet-wide as the default for ALL content** this evening (`d39995125`, deployed in
content-creator `v1.0.1178`, `971ede672`). More producers now route through this
path, and the fabrication was found *by accident* while checking that the voice
had landed — **the voice check passed and the copy still was not publishable.**

**Do first:** nothing is auto-publishing blog or social copy right now, so the
cheap correct move is to **keep it that way until the owner rules** — the bug file
asks for a call before any content-creator output is published anywhere. Read the
file, then put the gap in front of him. Do **not** start writing the fix before
that: it is a contract change on the claims layer and it collides with
`bugs_open/102`.

**Open:** `bugs_open/123_HANDOFF_2026-07-27_content_creator_output_can_never_reach_the_claims_assessor.md`

### Second: links are invisible on three live public sites

`bugs_open/122`, filed 20:41 BST from the oufe.com lane after the owner reported a
link as "dark blue on the black background and not easily readable". Measured with
the WCAG relative-luminance formula on `--color-primary` against each site's own
background:

| site | link on background | on a card | state |
|---|---|---|---|
| oufe.com | 1.23 | 1.00 | **fixed by hand** |
| dartsonline.com | 1.11 | 1.06 | untouched |
| robot-hands.com | 1.14 | 1.07 | untouched |
| vonc.com | 3.71 | 3.48 | untouched (fails AA body text) |

AA needs 4.5. Three of those are effectively invisible links on live public sites.
**oufe was fixed by hand and the generator was not changed**, so the next generated
stylesheet reproduces it — and `features_open/026` phase 2b, built tonight, already
finds contrast defects on **7 of 10** live sites (`6dd8667ea`).
**Do first:** hand-fix the three, the same way oufe was; the generator fix belongs
with the 026/`platform/colour` work, which is a different thread's live workstream.
**Open:** `bugs_open/122_HANDOFF_2026-07-27_generated_css_fails_wcag_on_four_live_sites.md`

---

## §A — Ready to resume now

### Brochure component library (fundamentallyai.com) — roll-unblocked
The site is live and link-sound, and the self-refreshing `evidence-chart`
component works. Per-page chart targeting was blocked by `bugs_open/085`, which
was filed as one line and turned out to be four separate defects on one journey —
the fourth found only by testing the deploy, because a pod-grep proves the binary
and never the path. The last fix (`32a55597e`) is in v1.0.1174.
**Do first:** fire one scoped section re-render of the fundamentallyai home page
and confirm it leaves **one** chart where three stand now — then deliberately
test the case that should show nothing. A green happy path proves the deploy, not
the fix; this thread has already paid for that lesson once.
**Open:** `docs/agent_docs/docs024_key_docs_latest/brochure_component_library/HANDOFF_2026-07-26_continue_here.md`

> **ADDED 22:00 UTC — this became the evening's busiest lane.** The owner reported
> the site was "nothing like the brief" on mobile. The thread built a **render
> audit** that measures the page a visitor actually gets rather than the sources it
> was built from (`881b4b5cf`) — and that one change found what fifty source-side
> checks could not. Outcome: a readable palette shipped, the site pointed at **21
> generated images it had never referenced** (`f56e78ea5`), three rounds on the CSS
> renderer, and `bugs_open/113`'s fix **live and verified** — which then exposed a
> **mirror defect** (`6162c7cd6`). Filed **113**, **114**, **122** and
> `features_open/026`; wrote `SUMMARY_2026-07-27b` under the title *"the site was
> unreadable, not undesigned"*.
> **Two things owed here, and they are not the chart:**
> 1. **The palette repair broke one page** (`e0d5c0a4b`) — that regression is
>    recorded and not yet fixed.
> 2. `features_open/026` **phase 2** is built and enabled, and its contrast check
>    finds defects on **7 of 10 live sites** (`6dd8667ea`, `63a64db48`). The WCAG
>    maths now lives in `platform/colour` with the real regression as its corpus
>    (`e43a3bda0`). That is the generator-side fix `bugs_open/122` in §0 needs.
> **On the stylesheet publish that memory records as an owner step:**
> `[INFERRED — not verified against the live site]` it appears to have happened.
> The sequence is `1cde867d8` (18:16, "pending one publish command") → `e0d5c0a4b`
> (19:57, "the palette repair **landed**, and repairing it broke one page"), and
> "landed" reads as published. **Check the served CSS before repeating either the
> "publish is owed" line or this one** — both are claims about a live site made
> from commit subjects.

### Model directory pipeline — roll-unblocked
Phase E is live on ai-agent-orchestration.com with three citation-verified
registers, but the publisher was deliberately cut back to the model-only chain
after a step-config string resolved as a *reference* rather than a literal and
published the model register three times under three different commit messages.
The fix (`bb99df77a`) is in the binary.
**Do first:** re-extend the publish chain to all three kinds, dispatch once, and
check the result by the **committed `files` map**, not by step statuses —
identical output from two supposedly different selectors is the entire signal.
**Open:** `docs/agent_docs/docs024_key_docs_latest/model_directory_pipeline/README_where_we_are.md`

### Gripper dossier pilot — roll-unblocked
The cluster half is built, tested, committed and council-approved, and had been
sitting inert waiting for exactly this roll. The island half was designed out of
existence on 07-26 when `tools-api` turned up on the same VM.
~~**Do first:** apply seeds 207 → 209 → 210 …~~ **DONE this evening, and the lane
is PROVEN END TO END.** Seeds 204/207/209/210 are applied (**208 deliberately
NOT** — its committed `base_url` still sits outside the island Caddy allow-list).
**Fixtures 1, 2 and 3 all pass live on robot-hands.com** — a formula literal, an
honest no-match, and the **failing branch verified live** rather than assumed
(`7bf582713`, `4366d69ee`, `e74372636`). Two dossiers are live and unlinked, both
returning 200.
**The lane is now PARKED, not blocked** — it works; nobody has asked for more of
it. Next real step is the public half, which needs `/api/v1/tools/gripper` built in
`tools-api` (the gauntlet thread's), and the consolidation programme wants
`platform/mailer` and `platform/httpguard` to land first or the estate forks again.
**New landmine found here:** `scheduled_tasks.target_topic`'s **DEFAULT is a topic
nothing consumes**, so a lane can fire into a void and report success (`4834b0d50`).
**Open:** `docs/agent_docs/docs024_key_docs_latest/robot_hands_gripper_dossier/` —
the RESUME doc written with fixture 3, not the 07-24 DESIGN.
**Note:** pod-grep the *spawned* worker, not the deployment — spawned pods run
`agent_definitions.image_tag`.

### vonc.com — publish the corrected About figures
Migration 229 fixed a real defect (the About page had "Archetypes" and "Tools
Live" transposed) but edited `content_data` only, so the live page still serves
the wrong numbers. The 043 lane calls this "the smallest outstanding task in the
lane — nothing else is closer to done", and it doubles as `bugs_open/093`'s
outstanding verification bar.
**Do first:** re-render vonc.com `/about` **without a writer pass**, and confirm
the stat-audit finding is raised.
**Trap:** `spec.reason` is control flow, not a label — `check_rerender_mode`
branches on it and an unrecognised value silently takes the assemble-only path,
republishing stale HTML and reporting COMPLETED. Vary `item_key` for dedup, never
`spec.reason`.
**Open:** `docs/agent_docs/docs024_key_docs_latest/fabricated_stats_043/HANDOFF_2026-07-26_continue_here.md` §8

### Claims verification — the recorded blocker is stale
Memory says V5 is "inert until image + evidence-researcher seed". **Both halves
of that have been true since 07-20.** The seed was applied that evening; what
actually stopped the smoke run was `bugs_open/047` (the webscrape adapter
rejecting every `batch_scrape`), and that closed on 2026-07-21. V5 has been
unblocked for six days and nobody has re-run it. V0–V4 are live and the daily
freshness sweep has been genuinely running since 07-26.
**Do first:** re-run the V5 evidence-researcher smoke for gaswholesalers per
`RUNBOOK_claims_verification.md` §7.
**Open:** `docs/agent_docs/docs024_key_docs_latest/claims_verification/HANDOFF_2026-07-20_claims_verification_resume.md`
(then the 07-26 entry at the tail of `NOTES_claims_verification.md`, which is far
newer than the README).

### Review queue drain / admin dashboard — roll-unblocked
The dashboard exists and its 50-row read cap is long fixed. The drain
(`revalidate_review_queue`) was built on 07-25 and documented as inert — it is
not inert any more; `f2570e1bc` shipped in v1.0.1174.
~~**Do first:** fire the revalidator with `dry_run=true` …~~ **DONE, and then it
ran for real.** The dry run worked but its first batch "proved nothing yet" by the
thread's own account (`5d7f970ff`); the **full-queue dry run then found 57
resolvable, and the thread recorded that its earlier caution had been 28× understated**
(`8732bdae7`). **The live run drained: 382 → 325**, and `resolution_path` has its
first automated writer (`b12d66ba4`).
**Do first now:** read what the live drain actually closed before firing it again —
this is a queue of *human review* items and the bar is whether each closure was
right, not the count. `bugs_open/033` is still OPEN.
**Open:** `docs/agent_docs/docs024_key_docs_latest/review_queue_drain/README_where_we_are.md`

### Experience register (substrate + write path)
The substrate went live today (migrations 218 and 230; the table is empty by
design), P2a was council-APPROVED, and the validating write path
`write_experience_pattern` (`36bb6c992`) is in v1.0.1174 and executable in
production. Its own council verdict `2e71f640` is still owed a read.

> **ADDED 22:00 UTC — the thread caught a serious one in itself.** The contract
> shape it had been building against was **invented**: all nine harvested entries
> would have been refused (`799c0c97e`). It logged the mistake to `WRONG_CALLS`
> in the sharpest possible terms — *"I invented the shape the experience register
> exists to stop people inventing"* (`3c8396578`) — and answered the
> P2b council with a **mechanical guard on the demotion list** rather than a
> promise (`45b2a3d90`). Separately `bugs_open/105` (`EvidenceFact.Kind` declared
> and never read) got readers **without rewriting a single stored register**
> (`606f485f7`); council APPROVED, **round 1 is live on v1.0.1179 and round 2 is
> owed a roll** (`1fdfc7cd5`, `b18dd564d`).
> The lesson generalises and is worth carrying: **a fixture you wrote yourself is
> not provenance.** Load the real artefact.
**Do first:** build the **bind path** — `bind_site_experience`, the action that
attaches a base register entry to one real site's pages and URLs, running
bind-time closure checks (every blank the base entry declares is filled) and
anchor checks (every declared destination resolves).
**Open:** `docs/agent_docs/docs024_key_docs_latest/experience_register/harvest/HARVEST_01_2026-07-26_vonc_provocations.md`
(it corrects the PLAN in ten places — read it before the PLAN).
**Keep distinct from the experience *loop* in §B. Two different threads.**

### Dispatch queue serialisation
`bugs_closed/030` is fixed and live — the cron lane split took publish-to-start
from about eighteen minutes to about one second. The residual found on 07-26 is
that a **wedged head orchestration freezes the whole interactive lane until a pod
roll**, and nothing notices.
**Do first:** design and build the idle-orchestration watchdog — today nothing
detects an orchestration whose `orchestration_states.updated_at` has stopped
advancing while it holds the lane offset.
**Open:** `docs/agent_docs/docs024_key_docs_latest/dispatch_queue_serialisation/README_where_we_are.md`

### Durable write guard
`bugs_open/021` closed 07-25 with both instances live and behaviourally proven;
its descendant 077 has closed too. Nothing in flight.
**Do first:** take the next verifier candidate — `undeployed_asset` (45 items in
seven days) — and **read the handler's remit before writing the verifier**. That
lesson has now been paid for twice.
**Open:** `docs/agent_docs/docs024_key_docs_latest/durable_write_guard/RUNBOOK_durable_write_guard.md`

### VM estate — the free, read-only phase
Design record only, nothing built. Three boxes were provisioned three different
ways; the two `setup.sh` forks share 61 lines and differ on 614. The owner has
ruled the island is pull-only and fixed the merge order.
**Do first:** P1/P2 on relojistas — extract the box's actual state into DB profile
state, render the nginx conf from it, and **diff byte-for-byte against the live
conf, read-only**. It costs nothing and it is what kills the 614-line fork.
**Open:** `docs/agent_docs/docs024_key_docs_latest/traffic_probe/HANDOFF_2026-07-26_continue_here.md` §5

### Leopardess rebuild — a config-only next step
Paused 07-26, and paused by someone else: the brochure thread built the charts
work as one shared fleet component on an owner ruling, rather than per-site.
**Do first:** add a `charts` key to leopardess's own `site_specs.evidence_base`.
The 18 facts already there include two chartable pairs. No code, no registration,
no image roll.
**Open:** `docs/leopardessconsulting/RUNNING_NOTES.md` (tail).
**Warning:** the READ-FIRST box at the top of `docs/leopardessconsulting/HANDOFF.md`
is stale — it makes `bugs_open/001` the headline blocker and 001 is now closed.

### Travelling docs
Paused 07-22 at T34: migration 195 is applied, so the tool-improver's tail now
emits a `section_edit` work item routed to the section editor instead of the
forbidden generic rebuild. Proven end to end on a hand-shaped item.
**Do first:** confirm the one link never exercised — that a **full LLM
tool-improver run** creates the `section_edit` item itself. It was deliberately
not forced, to avoid a regression on a good benchmark; the next tool-auditor
improve cycle confirms it naturally.
**Open:** `docs/agent_docs/docs024_key_docs_latest/travelling_docs/HANDOFF_2026-07-10_stage5_live_and_next_fronts.md` §0 then §7
**Correction:** its recorded caveat "delivery rides the cron-starved dispatch
lane, `bugs_open/030`" is stale — 030 is closed. The live sibling to watch is
`bugs_open/096`.

### Bug-backlog clearing
The 07-20/21 session shipped four items and left a method. The backlog is now 40
open files (was about 46 at handoff); discovery still outpaces closure.
**Do first:** clear the red build guard — register a verifier for
`contact_form_undeliverable`, or add it to `itemTypesWithoutVerifiers` with a
written reason. The `discovery_checks` test package is RED at HEAD, and `go build`
still passes, which is why images keep shipping over it.
**Open:** `docs/agent_docs/docs024_key_docs_latest/bug_backlog_clearing/HANDOFF_2026-07-21_bug_backlog_clearing.md` §4–§5

### Reasoning dataset
The extractor is built and proven; all three lanes are mined (7,712 records
across 448 trajectories). The go/no-go came back a split verdict — no training
model, yes a modest eval set, volume only from the council and feed lanes.
**Do first:** make the council records self-contained by materialising the two
joins — the `fix_plan` under review (via `trajectory_id`, since `input_state` is
empty) and per-seat model provenance via `llm_call_log`.
**Open:** `docs/agent_docs/docs024_key_docs_latest/reasoning_dataset/HANDOFF_2026-07-21_reasoning_dataset_resume.md` §Next actions

### Consolidation programme
Written and owner-directed today; nothing built yet. The one real scale blocker
it found is per-site Go actions — 9 of 296 registry entries serve 2 of about
1,000 sites, and five of those shipped in a single week. There is also **no SMTP
mailer anywhere in the built code**; the only working one lives in idea.uk's VM
app, outside `go build`.
**Do first:** build `platform/mailer` (item A2), then `platform/httpguard` (A3) —
both must land before the gripper dossier's public half or the estate forks
again. Coordinate with the gauntlet thread first: A2/A3 land in or next to
`tools-api`, which it owns.
**Recorded WON'T-DO:** merging the eight `StartHealthServer` copies. A sweep
called them "8 byte-identical copies"; hashing gives **eight distinct hashes**
serving one to three endpoints each. Merging means eight behavioural migrations
on the Kubernetes liveness path for zero benefit at any domain count — health
endpoints do not scale with site count. It is recorded so it stops looking
available.
**Open:** `features_open/024_FEATURE_consolidation_programme_for_fleet_scale.md`

### Concept register — R2 is due today
Stages 1–3 are complete (1,633 concepts, 16 seats live in both councils, the
direction guard and mission lane live in observe-only). The plan was "R2 is a
week of numbers, run after ~07-27, owner grades". Today is 07-27 and it is
unstarted; R1 has been emitting findings since 07-20 with nobody reading them.
**Do first:** run
`docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/101_REPORT_mission_review_findings.sh 7`
— it is free and read-only — and put the findings in front of the owner. **His
grading of them *is* R2** (that half is in §C).
**Open:** `docs/agent_docs/docs026_concept_register/HANDOFF_2026-07-20_council_continuation.md`
**New inbound, unactioned:** `bugs_open/106` — the register froze on 07-13 and
67% of workstream directories postdate it — plus a coverage checker and a backlog
ratchet, filed *for* this thread by the oufe thread today.

### vetcomparison.uk — one free check settles the blocked half
P1 (scaling the company-number crawl) is halted even though the owner has already
said "yes, scale it", because a crawl today **cannot record where any fact came
from** and would produce data we are not allowed to publish. That half is blocked
(see §B) — but the deciding check is free.
**Do first:** run **one real `vet-practice-verifier` verification** and read the
stored scrape markdown for footer and company-registration text. That settles
whether Firecrawl strips footers, and it decides whether the `bugs_open/101` fix
is worth writing at all. The bug file's own first line says the same.
**Open:** `docs/agent_docs/docs024_key_docs_latest/vetcomparison/HANDOFF_2026-07-26_continue_here.md`
**Rail:** never reintroduce `vetcomparison.co.uk` — we do not own it.

### idea.uk VM site — the chain has run for real
**First sale, 2026-07-27 11:13 UTC**: order `ord_1785090638951163875` went
request → confirm → engine → draft → approve → pay link → **real card payment** →
signed Stripe webhook → delivery → slot auto-released. Confirmed in three places
(memory, the handoff's START HERE block with the `POST /stripe/webhook 200` from
Stripe's own IP, and the owner log). The `running`-order slot leak was also fixed
and proven by inducing it, on the fifth deploy.
**Do first:** build out the **News** section — the only unfinished surface on the
site. (Correction worth noting: "News is next" appears in the repo docs, not in
the memory file, which still carries a passive next-step about cost measurement.)
**Open:** `docs/agent_docs/docs024_key_docs_latest/idea_uk_vm_site/HANDOFF_RESUME_idea_uk_vm_site.md`
— the START HERE block. The "PENDING — next actions" section further down is
07-18 vintage and stale.
**Note:** this tool is a standalone module with its own build-and-deploy path.
Chassis rolls, including today's, are irrelevant to it.

### Gauntlet dead-CTA (vonc.com)
P4 and the scoped-down Arena are both live; 72 of 73 deployed-page checks pass
and the claim scanner finds nothing across all 49 components — **nothing on
vonc.com is fabricated now**. Two faults left.
~~**Do first:** `bugs_open/083` — add the one-line log … then rebuild the island.~~
**DONE 20:47 BST.** The island was rebuilt on **v1.0.1178** and the fix is **live
and proven by an induced fault** (`a0d275916`, `a36b3d4d0`) — not by a happy path.
On the way the thread found **seven discard sites, not two**, and a third endpoint;
it corrected its own blast-radius claim twice and logged eight missteps from the
council rounds (`f1d224007`). The owner ruled on the council's deferred item: **log
a structural fingerprint, never model text, and check it mechanically**
(`1e2762809`, `7d12d4f2b`).
**Do first now:** candidates **2–4**, which the deployment unblocked. **083 stays
OPEN** — candidate 1 is only the first of four. A sibling thread has also
contributed fleet-wide scale numbers for unroutable findings, with a caveat
attached (`ae1dacc8b`) — read that before sizing the rest.
**Open:** `docs/agent_docs/docs024_key_docs_latest/gauntlet_dead_cta/HANDOFF_2026-07-26b_after_p4_live.md`
**Explicitly not helped by today's roll** — the thread checked: `browserrunner`
has had no commits since 07-25, and 083's engine runs on the island VM, outside
the cluster. This thread owns `tools-api` and the island; coordinate before
touching either.

### robot-hands.com site fixes — close it out
Declared done on 07-24 with R1–R9 closed; untouched since.
**Do first:** re-verify the deployed stat blocks still read index 10/6/39, about
10/6, gripper-detail 10/6/4/39, catalog 10 — then formally close the workstream.
**Open:** `docs/agent_docs/docs024_key_docs_latest/robot_hands/HANDOFF_2026-07-20_robot_hands_start_here.md`
(read its 07-24 UPDATE block first).
`[INFERRED]` — no doc names a remaining action; this is the check its own handoff
used.

### Site maturity ladder — requested, and *never started*
This is the one to look at hardest, because it is invisible. `features_open/015`
records it as the **stated methodology for scaling the fleet** — the owner's
words are that "just asking a new domain to become as developed as idea.uk in one
step is too much", and that he wants the actual planning done in a **separate
thread**, keeping the idea.uk thread for idea.uk. And yet: `find -iname
"*maturity*"` returns exactly one file, that feature request, with a single
commit. **There is no workstream directory, no PLAN, no standing five and no
owner.** The consolidation programme independently noticed the same gap today.
**Do first:** open the thread properly — create the standing five under
`docs/agent_docs/docs024_key_docs_latest/site_maturity_ladder/` and draft the rung
definitions. Design only, no build.
**Open:** `features_open/015_FEATURE_staged_site_maturity_ladder.md`
**Sequencing note:** the ladder sits behind the pilots, so it will interlock with
per-site AI operation (§C) — but nothing stops the design work starting now.

### Chassis replica scaling
Design only, nothing built. §5A is settled by a code read, and the ownership
discard that made a shared consumer group unsafe has already shipped from the
other side, leaving exactly one known blocker.
**Do first:** build **P1** — thin ingest: persist an intake row, commit the
offset, and claim work with `FOR UPDATE SKIP LOCKED`. It needs no topology
change. Before P2, guard `processResponseClaimWithRetry`'s claim recovery on
staleness for `status='processing'` rows.
**Open:** `docs/agent_docs/docs024_key_docs_latest/chassis_replica_scaling/NOTES_chassis_replica_scaling.md`
(newest — the PLAN's `SetExecutingStep` claim is corrected there).
**Owner questions outstanding but not blocking the build:** does P1 go first, what
is the target per-domain daily volume, and how long are job records kept.

---

## §B — Waiting on another thread or a bug

### Gemini content provider — blocked on `bugs_open/112`, and on 029
P1–P6 are done and this is a genuine success: both `content-creator` and
`page-content-writer` are configured for Gemini, and the 07-24 verdict that
"Gemini writes badly" turned out to be **our own starved token budget**, not the
model. P7 — rebuild one real page and read its copy — has never happened.
~~**Blocked by:** `bugs_open/112` … and `bugs_open/029` … Also needs an owner call
on which live page may be mutated.~~
**UNBLOCKED AND DONE — move this row to §A.** `112` closed at 21:54 BST. **P7
happened: the Gemini writer wrote a live page**, and `bugs_open/107` closed with it
(`ebe4ac313`) — with the tart footnote that *the handoff's own next step could
never have worked*. The owner call on which page to mutate was overtaken by events
and is struck from §C.
**Its actual state now:** the lane works end to end. What it produced is the
problem — see **§0 / `bugs_open/123`**: the copy it writes **can never reach the
claims assessor**, and a live run invented a hallucination-rate statistic. This
thread filed that bug against its own success, which is the right instinct.
**Also from this lane tonight:** a four-model bake-off (`a51914ecf`) — Grok
amplifies the dryness complaint, **Fable fixes it**, and one prompt clause nearly
closes the gap for free. That is a live content-quality decision with data behind
it, sitting unmade.
**Open:** `docs/agent_docs/docs024_key_docs_latest/gemini_content_provider/HANDOFF_2026-07-27_continue_here.md`
**Landmine:** `jsonb_set` with a literal object is a REPLACE — it would have
silently deleted the writer's `max_tokens: 8000` sibling, a fourfold budget cut,
invisible.

### CTA / link integrity — the fix has shipped and never fired
`bugs_open/023` and `049` are both closed, both halves live in v1.0.1171, and the
fleet's broken links went from 312 to 118. The real mechanism turned out to be
that **chrome was writing links to pages that were never built**. The gap: the
fix has never fired in production, because the two remediated sites no longer
have qualifying nav items.
**Blocked by:** the three sites that still qualify — dartsonline.com,
leopardessconsulting.co.uk and oufe.com — are outside this thread's remit and a
chrome re-render is an outward-facing content change. Their owning threads should
fire it, then grep the pod logs for "dropped nav items whose target page has
never been deployed".
**Open:** `docs/agent_docs/docs024_key_docs_latest/cta_link_integrity/HANDOFF_2026-07-26_049_closed_continue_here.md`

### Experience loop — its build round was taken over
CP2 closed on 07-19. Its T4 MVP build has been **superseded**: the gauntlet
dead-CTA thread now owns vonc's build cycle end to end and shipped that work
itself. What remains genuinely unbuilt here is the journey-acceptance runner.
**Next when unblocked:** T5.1, the browser-runner Tier-4 journey runner —
additive persistent-context path in `internal/adapters/browserrunner/`, never a
rework of `Execute`. Separate image from the chassis.
**Blocked by:** the gauntlet thread, whose own next steps overlap T5.1's browser
half.
**Open:** `docs/agent_docs/docs024_key_docs_latest/experience_loop/RUNBOOK_experience_loop.md` §T5.1

### vetcomparison.uk P1 — blocked on `bugs_open/100` and `101`
Provenance is **structurally impossible**, not intermittently broken. The write
path reads `source_type` / `source_name` / `source_url` out of the LLM's own
returned object, the verifier prompt never asks for them, and `store_results`
excludes `scraped_data` from its inputs, so the fetched URL can never reach the
writer. Zero of 2,970 observation rows carry a `source_url` key at all. Separately
(101), four `scrape_web` config keys — `follow_links`, `max_pages`,
`extract_mode`, `fallback_url_field` — are read by no Go code, so a step
configured as a six-page crawl fetches the home page once.
**The Go change is not written.** Verified against HEAD today: the offending reads
are still at `platform/orchestration/actions/business_intel_actions.go:267-268,
301-302, 322-324`. Both bugs are OPEN and unowned. Both are `platform/` changes,
so: council gate → build → roll.
**The free check that decides whether 101 is even worth writing is in §A.**

### Imagery best-in-class
Paused 07-25 having closed `bugs_open/011` on a fleet-wide zero, with a round-9
unanimous approval. Its two non-routing residuals were moved out to
`features_open/022` and `023` rather than held open.
**What is left:** a gibberish SDXL hero still serving as the header on
`leopardessconsulting.co.uk/how-it-works.html`. Regenerate it, rebuild the page,
no code change.
**Blocked by:** it is a live change to a client site owned by the leopardess
thread, and the imagery thread deliberately declined to make it. Also an owner
nod (§C).
**Open:** `docs/agent_docs/docs024_key_docs_latest/imagery/HANDOFF_imagery_best_in_class.md`
**Roll note:** its round-9 truncation-accounting change is now in v1.0.1174.
Verify by log-message literal in the pod, **not** by symbol grep — the alpine
build does not retain `case` values and a symbol grep gives a false negative that
looks exactly like a stale deploy.

### webdesign.co.uk — the news feed ingested nothing
Phase 1 is live: 98 pages, all returning 200. Phase 2 was redirected the same day
— "Hire" was rejected, and the section is now "Buying design" for the
£100k-plus commissioner. The news feed was armed properly (creating
`content_sources` rows does *not* arm a feed — the classification flag is the
switch), the 13:49 UTC tick dispatched all five sources correctly, and **ingested
zero**: everything died at `spawn_ingester`, which is `bugs_open/029`,
roll-adjacent at 218 seconds after a chassis start. Sources were re-armed for the
19:49 UTC tick.
~~**Do first (free):** confirm items landed … If the 19:49 tick also failed, this
is blocked on 029.~~ **ANSWERED: the feed INGESTED — 50 items** (`231e56bd3`), and
the thread posted two corrections to its own earlier claims in the same commit.
**Not blocked on 029 after all.** The queries were then retuned at the industry and
real developments rather than the generic terms (`d00c65099`).
**Three things this lane found tonight, all worse than the feed problem:**
- **All 63 tool pages were dead ends** — no link to their guide or siblings. Fixed,
  and the deploy bug that hid it was filed (`a5731de0f`).
- **The Cloudflare beacon could never have rendered.** The owner supplied the token
  and it was armed (`6dca0f664`) — then the thread found **the chrome renderer does
  not read `site_components.content_data`** (`9f019c9f3`), so the beacon was inert
  by construction. §C item 10(a) is therefore **done as an owner action** but the
  analytics still are not live.
- That chrome-renderer finding is the same mechanism as `bugs_open/117`/`118` from
  the relojistas lane — **site chrome is a stored artefact that no page re-render
  rebuilds.** Two independent threads hit it within an hour.
**Order matters:** feed → build the page → **then** chrome. Publishing the News
nav early puts a 404 in the header of all 98 pages.
**Open:** `docs/agent_docs/docs024_key_docs_latest/webdesign_couk/HANDOFF_2026-07-27_phase2_uk_authority.md`
(the live plan is `PLAN_2026-07-27b_buying_design.md`; the earlier buyer-track
plan is superseded). Two owner items on this thread are in §C.

---

## §C — Waiting on the owner

*A session cannot clear any of these. Each line is the decision, not the
background.*

> **REFRESHED 22:00 UTC. Five of the seventeen are gone — four ruled, one overtaken.**
>
> - **Struck 1 (Gemini writer revert)** — the defect was fixed properly instead;
>   no revert needed. See §0.
> - **Struck 2 (which page for Gemini P7)** — overtaken; the writer wrote a live
>   page and `bugs_open/107` closed.
> - **Item 3 (relojistas box session) — HE RAN IT.** 3 of 4 items closed; the 4th
>   *was never buildable* (`3668904ca`). Read that before re-asking for anything on
>   that box. The contact route is separately **verified gone from 18 of 18 pages**
>   and the site is link-clean (`daae6e593`).
> - **Item 9 (oufe audience) — RULED, then REVERSED the same evening.** First
>   decided as "anyone learning how this works" (`0f137e1e8`), then **reversed to
>   the mid-market professional, with students getting their own domain**
>   (`00600f89b`). The mission brief and house voice were realigned to the
>   *reversed* answer (`19f0f5eb0`) — so if you read only the first commit you have
>   the wrong audience.
> - **Item 10(a) (Cloudflare token) — SUPPLIED and armed.** But the beacon still
>   does not work for an unrelated reason (see the webdesign row in §B), so do not
>   record analytics as live.
>
> **Two new ones arrived tonight and belong at the top of his queue:**
>
> - **A: may any content-creator output be published at all?** `bugs_open/123`,
>   §0 — a confirmed live fabrication with no checker that can see it. Until he
>   rules, nothing blog- or social-shaped should ship.
> - **B: `bugs_open/096` needs a maintenance window that the bug itself prevents.**
>   The thread measured this rather than trying it (`684ec9ca5`, `d97a7527c`), and
>   two hazards for applying the lane fix are written up. It is explicitly his call.
>
> Items **4, 5, 6, 7, 8, 10(b), 11, 12, 13, 14, 15, 16, 17 were not re-checked
> tonight** and stand as written at 16:05 UTC — except item 4, updated below.

1. **Gemini writer revert (§0).** Un-arm tonight's fleet-wide build failure by
   reverting `page-content-writer` to Claude in the DB — accepting that it undoes
   the Gemini thread's P6 — or leave it armed and ship the ten-line Go fix plus a
   roll first. The cost data argues for the revert regardless: ~10x billable
   output per section.
2. **Gemini P7.** Which live page may be rebuilt to prove the writer's output.
   The thread suggests `fundamentallyai/about` after a word with the brochure
   thread.
3. **relojistas.com — the one box session.** He must run one root SSH session on
   `167.233.33.159`. Four things in order: (a) `scp` the reconciled generator up
   and run it in `MODE=full` — the current hand-edit is case-**sensitive** so
   lowercase `?type=rss2` still 404s; (b) append `WEBROOT_DIR` and `RESULTS_PATH`
   to `/etc/site-engine/site-engine.env` and restart, which switches on
   search-that-answers (binary already deployed); (c) confirm real client IPs are
   reaching the access log rather than Cloudflare edge addresses — **this is a
   measurement prerequisite, not housekeeping**: today every visitor IP is a
   Cloudflare address, so subscriber counts are impossible; then (d) a
   cluster-side scheduled-task flip that anyone can do once (a) is done.
   **Correction to the memory index:** the homepage hero CTA 404 was fixed at
   15:30 BST today on his ruling that relojistas gets no contact route
   (`0a959e784`). The underlying platform default still belongs to
   `bugs_open/071`.
   **~~Second item~~ RETRACTED 07-28:** I claimed the Afternic listing had Minimum
   Offer = 0 and that the anti-lowball floor was absent. **Both false** — the floor is
   **$12,000** (owner). I misread a column-aligned dashboard paste. Nothing to decide.
4. **Architecture review seat.** Approve one new `capability_gap` spec — set both
   `owner_approval` and `code_pointers` — and give the go for a single
   `feature-designer` run on it. This is the *only* action that produces a
   review. The seat is fully live and config-only, but `review_architecture` has
   **zero reviews** and waiting will not change that: it exists only on
   `feature-designer`, which refuses to run without an approved spec, and the two
   approved specs both belong to other threads. Candidates named:
   `forced-text-color-fixer`, `site-metadata-fixer`.
   `[INFERRED]` — the trigger the docs promise ("the colour thread's round 4")
   has already been spent: round 4 ran and was approved 4–1 at 12:32 UTC, about
   an hour before the seat's cutover.
   Decisions D7(b) and D9 are **closed** — he ruled on both this evening. The
   open one is **D10** (landmines as a footprinted corpus), drafted by the
   bugfix-061 thread, awaiting his read. The file is still named
   `PROPOSAL_D9_...` after a numbering collision.

   > **UPDATED 22:00 UTC — the ask has changed shape, and the seat argued its own
   > case.** The thread's D11 submission was that **seats must be able to look
   > things up**, and it landed: `code_checks` now reach the code index and
   > `feature-designer` gained `code_lookup` (`9360f2997`), with **symbol bodies
   > finally in the index** so a `content` check can match what its contract
   > promised (`37f7deff9` — that is `bugs_open/108`'s core defect). Council
   > **APPROVED layer 1 on round 3** (`c1e014c1b`), and layer 1 is **built but not
   > live**; the thread pre-measured body slicing so the post-roll numbers have a
   > baseline (`8c7a2064f`).
   > **Three honest findings from the same thread, worth more than the approval:**
   > `prior_art` caught that the "shared slicer" it proposed **already existed**,
   > and two of its numbers were borrowed rather than measured (`dc3873c64`); the
   > **approved plan's central claim was false** and it said so after approval
   > (`53ff8c80a`); and it filed `bugs_open/121` against itself — the house voice
   > was duplicated and *"the architecture seat that should have caught this has
   > never reviewed anything"* (`e96923b1f`). A **commit hook** produced the
   > architectural observation the seat could not (`3bf0cf1ee`).
   > **So the decision is unchanged but sharper:** the seat still has zero reviews,
   > and tonight produced direct evidence of the cost. Cold-start is the newest
   > handoff (`c036cd1f6`/`1c712931d`), not the "b" one.
5. **Feature-builder / work-item completion integrity.** Give the go to fire
   `feature-implementer-orchestrator` on plan `b5097ade` — round 4 approved 4–1
   today, objection trend 3→3→2→1, and the recurrence risk is now pinned by a
   build-enforced test. He has approved the *spec*, not the *plan*, and this
   thread spends no run without a per-run go; its handoff says "STOP THERE" after
   reading a verdict. If the answer is not yet, the parked alternative is
   Submission A, three rounds in and "two small answers away".
   Pre-flight when it fires: `git fetch` — PR #3's merged code is **not in this
   working tree**.
6. **Diagnosis→fix loop.** Which bug the loop is aimed at next. Every tier is now
   proven end to end; each run spends credits, so the target is his pick.
7. **Per-site AI operation — which un-pause bar applies?** He paused it on 07-24
   until "the team already working that site finishes". Both written conditions
   are now **met**: the robot-hands thread closed 07-24, and bug 043 closed
   07-26. But the *043 lane* has not gone quiet — it committed again today, and
   still owes `bugs_open/093`, `bugs_open/102`, a council round 7 and a ruling
   from him. **The load-bearing detail is that none of that residual is on
   robot-hands.com**, which is the pilot site, so the stated reason for the pause
   — building on ground that is still moving — no longer holds. One line from him
   unblocks the Tier-3 pilot.
8. **The aao agent figure (043 lane).** The site says both "70+ agents in 8
   departments" and "30+ agent types" while live is ~175. "Agents", "agent types"
   and `agent_definitions` rows are three different units, and the database has
   no notion of a department. This is a statement about what the business *is*,
   not a query.
9. **oufe.com audience.** Mechanism-led wide professional audience (the thread's
   recommendation), students proper, or mid-market professional. The briefs need
   revising either way and pages are being written now. Disclaimer wording is a
   second, smaller call.
10. **webdesign.co.uk, two items.** (a) One Cloudflare dashboard step — Web
    Analytics, add domain, Automatic Setup; the token cannot be minted from here.
    (b) How exposing does the "Buying design" copy get: generic failure classes,
    anonymised cases, or **our own named failures with evidence**? The thread's
    standing instruction is "write nothing at (c) until he rules". A third, softer
    one: do the designers stay — it decides whether part of that section is a real
    rewrite or a holding action.
11. **About-page commercial block — Spanish wording.** The approved English
    register ("available to acquire — register your interest", never "for sale",
    no on-page price) does not survive machine translation, which is explicitly
    ruled out. relojistas is the first non-English site to want the block, and it
    sits configured and deliberately uninserted. Related: should
    `for_sale_requested` imply a non-zero Afternic minimum-offer floor.
12. **Empty sections — seven hollow stubs on three live sites** (finetuning.uk
    `ai-guides` and `insights`, gaswholesalers.com `fuel-industry-insights`).
    These are pre-guard *data damage* on live pages, not a mechanical fix:
    stripping them degrades the pages, and rebuilding them risks the
    `bugs_open/029` fabrication path. **Do not force-rebuild those sites to clear
    them.**
13. **Concept register R2.** The 101 report is free and read-only; his grading of
    its findings *is* R2.
14. **Council gate — one stalled item.** Correlation `bd12762a` has sat at REVISE
    since 07-20 with no runs. Spend a round-3 credit, or record it closed
    unapproved. (Separately: PR-mode enforcement stays deferred, and scheduling
    the `098 --persist` coverage report needs an approved runner with both git and
    kubectl.)
15. **Multi-session coordination.** Approve the narrow `commit-msg` hook that
    rejects a broad multi-area commit unless the message is labelled `sweep:`. The
    machinery this thread built is otherwise live and working; this is the last
    offer, left open since 07-18. (His 07-17 ruling was against a *staging*
    enforcement hook — this is the narrower variant.)
16. **Reasoning dataset item 2.** Generate data at volume by replaying the ~30
    already-solved bugs? It would also refill the dead revise lane. Item 1 needs
    no decision.
17. **Imagery.** Say the word and the leopardess hero is a few minutes' work
    (§B).

---

## §D — Parked on purpose

- **News feed pooling** — his hold of 07-20: "not quite ready to onboard more
  domains". Designed and partly built: 17 inert pool sites, 17 pools covering
  1,037 of 1,625 domains. Ingestion is structurally inert and safe to leave
  indefinitely. **Do not onboard a pilot domain, arm a pool, or write a
  `classification` spec with `news_feed.recommended` to a pool site — that write
  is the arming act.** Two technical blockers should clear first anyway:
  `bugs_open/026` (English hardcoded, and the cohort is multilingual on day one)
  and `027` (news renders nothing without JavaScript, which defeats the point).
  Anchor: `features_open/005`.
- **Med pipeline** — the three scheduled tasks are disabled rows, blanked
  fail-closed after the fabrication remediation. `med-export-json`'s empty domain
  is deliberate, not a gap. The defect underneath closed as `bugs_closed/061`.
  Enabling is his call; when it happens, re-arm the retailer arm
  provenance-first. Never restore `vetcomparison.co.uk` as the export domain.
- **UK-sovereign stack exploration** — deferred to a dedicated future chat; do not
  start unprompted. The 07-10 baseline still stands: compute is UK (Rackspace),
  storage is **not** (Backblaze B2, us-east-005), models are US plus in-cluster
  Ollama. First real question when it opens is UK object storage. No repo doc
  exists — only the memory file.
- **vonc.com / Spark** as a standalone thread — the memory entry is stale (~07-17)
  and the live vonc work now lives in two other threads: gauntlet dead-CTA owns
  the experience, and the 043 lane owns the claims. Both are listed above. Nothing
  to restart under this name.

---

## About the bug backlog

~~**40 open files in `/bugs_open/`, 90 closed, as of 16:05 UTC today.**~~
**CORRECTED 22:00 UTC: 44 open, 98 closed** (`ls bugs_open/*.md | wc -l`). Net +4
open, but the gross movement is the story — **eight closed and eleven filed in six
hours**, and **20 of the 44 open files were filed today**. Closed this evening:
**010, 034, 070, 086, 095, 103, 107, 112**. Filed this evening: **113–123**.
Discovery is still outpacing closure, and the reason is visible in the filings —
several came from threads measuring *rendered output* for the first time
(`881b4b5cf`, `6dd8667ea`), which is a new instrument finding a standing backlog,
not a new rate of breakage.

The old version of this file carried a bug table; it is
deliberately not reproduced, because bugs move between the two directories
several times a day and a copied table is wrong within hours. `/bugs_open/` *is*
the index — read it directly, and read §10 of
`016b_debugging_guide_8_consolidated.md` for the mechanism-level cross-reference.

Two things worth knowing before picking one up:

- **Numbering is one sequence shared across both directories and is never
  reassigned**, so a stale `bugs_open/NNN` pointer resolves by number in the other
  directory. `016` and `017` are each used by two different cases — resolve those
  by slug, not by number.
  **Updated 22:00 UTC:** `083` is **currently duplicated inside `bugs_open/`** —
  two different cases, one number, both open. (`107` and `112` were also duplicated
  earlier today; both resolved when the closures moved files out.) And **112 was a
  live collision this afternoon**: a bug was filed as `110`, renumbered to `112`
  because another session had taken `110` concurrently, and collided with the
  Gemini `112` — it was refiled as **`113`**. So a doc written this afternoon may
  say "112" and mean the CSS palette bug. **Resolve by slug, always.** This is the
  second time in one day that concurrent filing produced a collision; it is a
  property of the workflow, not an accident.
- **Run `scripts/who-owns.py <number|slug>` first.** Most "open" items are already
  owned by a workstream. It is advisory, takes about a third of a second, and
  makes no cluster calls. Its blind spot: it reads *commits*, so a session
  mid-fix with nothing committed is invisible to it — re-run `git log` at
  implementation start.

**The diagnosis queue is still empty, but it is now failing rather than idle.**
Re-checked 22:00 UTC: no `awaiting_diagnosis` rows — but the tallies moved to **20
complete, 4 cancelled, 6 failed**, i.e. **both diagnosis attempts this evening
died** rather than queueing. That is `bugs_open/029` (see the note below), and it
means "the queue is empty" currently reads as "nothing can get *into* it", which is
the opposite of reassuring. **Do not read an empty queue as spare capacity until
029 is understood.**

### `bugs_open/029` — the fleet blocker underneath several of these rows

Hung spawns have halted builds fleet-wide since 19 July, and tonight it took out
two diagnosis runs and is what blocks `bugs_open/097`'s mechanism question
(`ee015698e`) and is cited by `085`. A fresh instance was captured with a clean
timeline (`a5ebcce5d`), and the thread ruled out the obvious explanations in the
bug file itself: **not** the ~300s-after-restart rule (the spawn was 22 minutes
clear), **not** a stale image (the pod ran `v1.0.1179`, so `bugs_open/066`'s
spawn-image path is working), and **not** a general spawn failure — 17 page
re-renders went through spawned pods in the same window and all completed.
Whatever it is, it is specific to *the request reaching the spawned worker*.

> **The live specimen is GONE.** The 029 thread deliberately left the hung pod
> running — *"so that whoever owns this bug gets a live specimen rather than my
> description of one"* — and the bug file gives two `kubectl` commands to inspect
> it. **I checked at 22:00 UTC and the pod no longer exists**
> (`Error from server (NotFound): pods "agent-diagnose-orchestrator-f26bf2fb-g2sz6"
> not found`). Do not spend time on those two commands. Whoever picks 029 up needs
> to capture a **fresh** specimen, and should assume the window is well under two
> hours — which is itself worth knowing, because the bug file's own advice
> ("while it lasts") turned out to be shorter-lived than the handoff describing it.

---

## What was verified for this rewrite, and what was not

**Checked live, today:** the running image tag and pod start time; the open and
closed bug file counts; the diagnosis queue; `git log` for the last three days;
`git status`; the contents of `bugs_open/112`.

**Read from the workstream docs and memory, not independently re-verified against
the cluster:** every per-thread STATE line, and every claim about what is live on
a given site. Those come from each thread's own newest handoff or owner log,
which is the best available source but is in some cases hours old. Where a thread
records that something is "live" or "proven", that is its claim, not a
re-measurement made here. `[UNVERIFIED]` markers appear inline above where a
specific claim could not be grounded.

**Ask for this file to be refreshed at any time** — it is a snapshot, and on a
tree this busy the half-life is roughly a day.

---

## What the 22:00 UTC refresh checked, and what it did not

**Checked live, at 22:00 UTC:** the running pod image (`v1.0.1179`); the makefile
tag; open and closed bug file counts (44/98) and the duplicate-number scan; the
`needs_diagnosis` queue tallies straight from `site_work_items`; the full
`git status` and `git log` for the six hours since the rewrite; **the existence of
029's specimen pod (it is gone)**; and the first 30 lines of `bugs_open/123` and
`bugs_open/122`, which is why §0 changed.

**Read from commit messages, not independently verified:** every per-thread update
in this refresh. Commit subjects in this repo are unusually descriptive and often
self-correcting, which makes them good evidence of *what a thread believes about
its own work* — but a thread saying "live and proven" is that thread's claim, not a
measurement made here. **No site was fetched and no page was rendered for this
refresh.** Where a claim is an inference from commit ordering it is marked
`[INFERRED]` inline — there is one, on the fundamentallyai stylesheet publish.

**Deliberately not touched:** §D, and the §A/§B rows whose threads were quiet this
evening (model directory pipeline, claims verification, dispatch queue
serialisation, durable write guard, VM estate, leopardess, travelling docs,
bug-backlog clearing, reasoning dataset, consolidation, concept register,
vetcomparison, site maturity ladder, chassis replica scaling, experience loop, CTA/
link integrity, imagery, vonc About figures). Those stand as written at 16:05 UTC.
**"Not re-checked" is not the same as "still true"** — a quiet thread's row ages at
the same rate as a busy one's, it just has nobody correcting it.

**A note on this file's own half-life.** The 16:05 rewrite said "roughly a day".
Six hours later its §0 was fixed and closed, four rolls had shipped, and a third of
the owner queue had been ruled on. **On an evening like this one the half-life is
closer to three hours than a day.** Re-run `git log` before trusting any row.

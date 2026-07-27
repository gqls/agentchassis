# NOTES — fabricated_stats_043 (bug 043 fixing lane)

Running record, append-only, newest at the bottom. **The account of record is
`bugs_open/043_HANDOFF_2026-07-20_generated_page_copy_invents_quantitative_claims.md`**
— this file holds only what doesn't belong there (working detail, missteps).
The robot-hands-specific half of the story lives in
`../robot_hands/NOTES_robot_hands_site_fixes.md` (Turns 10–13).

---

## 2026-07-24 — sweep run, wave-1/2 fixes, migration 201, evidence_base

Session artefacts, in order:
- `SQL_2026-07-24_vonc_index_stats_de_fabricated.sql` — wave 1, applied.
- `SQL_2026-07-24_gamesdesign_index_stats_traced.sql` — wave 1, applied.
- `SQL_2026-07-24_aiagent_casestudy_stats_grounded.sql` — wave 1, applied.
- `SQL_2026-07-24_wave2_sweep_findings_fixed.sql` — sweep findings (robot-hands
  index REGRESSION, vonc about/index extras, aiagent index/about, leopardess
  clone fallbacks), applied. NOTE the two `UPDATE 0`s in its first apply: the
  vonc index components had been REPLACED WITH NEW IDs by the in-flight wave-1
  re-render — **page_components.id is not stable across re-renders; key content
  edits on (page, component, label), never on pc.id.** The residual fix went in
  by label after the render completed (wave-2b, in-session SQL, recorded here).
- `../sql_for_agents/201_content_writers_never_invent_numbers.sql` — candidate 3,
  applied + ledger-recorded (`--record-only`).
- `SQL_2026-07-24_evidence_base_four_sites.sql` — writer_blocks for the four
  fixed sites, applied; **carries an in-file CORRECTION**: the vonc seed
  superseded migration 166's structured row (experience-loop `banned_claims`)
  unread — caught on the `UPDATE 1` count, fixed with a MERGE row. WRONG_CALLS'd.

Verifications were against rendered pages, not statuses (all pass; the one
residual "70%" grep hit on aiagent was a CSS gradient stop, not a stat).

**Open when picking this lane back up** (also in 043 itself):
1. Candidate 1 — bind stat fields to provenance in component schemas; the
   `stat_N_value` llm_guidance examples ('2.4M') are the invention's seed shape.
2. Candidate 2 — post-generation numeric audit → needs_human_review.
3. finetuning.uk/about — fabricated "Clients Served 11+ / Satisfaction 100%";
   needs the owner's real story, do NOT invent a replacement.
4. Prose numbers outside stat fields were not audited (the robot-hands 9-vs-42
   ratio says stat blocks are the tip); 201 guards new writes only.
5. Behavioural proof of 201: the next routine `needs_page` full-writer render on
   any of the four seeded sites should keep the corrected stats (evidence_base
   supplies them). Probe: compare stat values before/after the next
   `render_news_section` item on robot-hands/index.

## 2026-07-24 (later) — wave-2c: the prose tail on vonc

The last grep residual ("4h 12m") led past the stat fields into ordinary prose:
fabricated countdowns, "real time"/"clock is live" theatre, and an entire
component (archetype-combinations) built on six archetypes the site does not
have. Fixed in `SQL_2026-07-24_wave2c_…` (5 UPDATEs, verify query returned 0
remaining, 3 re-renders queued). Scope line drawn and recorded in 043: the
present-tense product-VISION copy (arena guide article, conceptual
differentiators) belongs to the experience-loop/vonc-spark thread — their 166
banned_claims routes it to review by design; not touched from this lane.

## 2026-07-24 (evening) — wave-2d: the poisoned given, caught live

The final drain-verify showed aiagent/index CLOBBERED at 15:10 — a full-writer
re-render wrote "70+/8/30+/1000s" back over the grounded values, WITH migration
201 and the evidence_base live. Mechanism (fourth for the 043 file): the site's
own specs hardcode the figure list into the writer's "follow this closely"
content-direction — a spec number is a GIVEN and outranks every writer-side
rule. Truth audit flipped part of my wave-1 read: "over 70 agents" and the
"8 departments" taxonomy were TRUE (registry 171; departments = the platform's
own named structure — my evidence_base ban was over-broad, now narrowed to the
"departments served" misframing). The one untrue clause ("thousands of
concurrent agent instances") became the measured truth ("over a thousand
orchestrations a day" — 1,699 in the last 24h). All four aspects patched by
versioned supersede+insert; one differently-worded variant in briefing needed a
second pass (verify query caught it). Index stats restored and re-queued —
values recomputed live came back HIGHER than wave-1 (171 agents, 14 sites,
1,284 work items): the platform grew during the session, which is the whole
point of computed values.

Lessons: (1) verify EVERY page the writer touched after a full-writer pass, not
just the one you edited — about/case-study held (lightweight path), index did
not (writer path); (2) sweep the spec aspects for numeric claims — added to
043's candidate-1/2 scope.

## 2026-07-24 (close) — BEHAVIOURAL PROOF: the writer regenerated a stat block and did not fabricate

The wave-2d re-render (16:08) went through the FULL WRITER again — provably: the
persisted labels are fresh rewordings ("Agent Definitions", "Work Items
Completed"), not the stored ones. And it wrote **170 / 13 / 17 / 1,267** — the
exact dated snapshots listed in the evidence_base writer_block, rendered live.
So the first live exercise of the complete stack (de-poisoned spec → repointed
content_direction → Verified Facts → rule 14 v2) produced a freshly-generated,
fully-true stat block. Item 5 of the open list (behavioural proof of 201) is
DONE — earlier and more convincingly than the planned probe: same page, same
writer path, same afternoon as the 15:10 fabrication, opposite outcome.

Note the writer chose from the evidence list rather than echoing stored
content_data (it replaced my 171/8/14/1,284 restore with the LISTED 170/13/17/
1,267 snapshots) — exactly what the block licenses ("dated snapshots up to a
listed live count are fine"). On writer-path pages the evidence_base IS the
source of truth; keep IT current, not the content_data.

## 2026-07-26 — candidate 1 shipped (mig 217), candidate 2+4 built (mig 218 + Go), and three things I got wrong on the way

Session artefacts:
- `../../sql_for_agents/217_stat_values_optional_and_template_gated.sql` — candidate 1,
  applied + ledger-recorded. 80 fields / 46 template gates / 10 components, plus the
  `component-creator` NUMERIC FIELDS RULE and the writer-prompt optional marker.
- `../../sql_for_agents/218_evidence_facts_for_043_sites.sql` — real `facts[]` for
  robot-hands, gamesdesign, ai-agent-orchestration. Applied.
- `platform/orchestration/datahelpers/claims_stats.go` +
  `platform/orchestration/actions/validate_page_content_stats.go` (+ tests) — candidates
  2 and 4. Committed, **INERT until the next image roll**.
- Council gate submission `569241fb-dd8d-4bcf-b382-234dfca1365c`.

**MISSTEPS, which are the point of this file.**

1. **I nearly shipped a check that could never fire.** I wrote a
   `stat_partial_blanking` lint (043's point (c) — a block where the unsourceable stats
   were blanked to an em dash and the rest left reads as *checked* while carrying a
   surviving invention) into `LintStatUnits`. It takes `[]StatClaim`, and
   `ExtractStatClaims` drops blank sentinels by design, so the blanked stats never reach
   it: the check was structurally incapable of firing. Caught only by re-reading my own
   code before running it. Deleted rather than shipped, with the reason written where the
   function would have been. **A detector that cannot fire is worse than none — it reads
   as coverage.** That is the same failure shape as the finding below, which is why it is
   worth the paragraph.

2. **Two of my assertions were caught by the migration's own guards, not by me.** The
   dry-run failed twice before it passed: first on a miscounted needle total (I wrote 41,
   it is 46), then on a post-condition regex `_(value|description)$` that flagged
   `archetype_description` — a prose field the migration deliberately leaves required.
   The second is exactly the over-broad-predicate mistake the migration's own header
   warns about for the `WHERE` clause, made in the assertion instead. Both are arguments
   for writing the guard before the change, not after.

3. **My test premise was wrong, not the code.** `TestStatFieldPairing`'s "ambiguous
   anchor" case used `row2_note` as the second candidate, expecting `unpaired`. It
   resolved to `anchor` — correctly, because `note` is a detail-role token and role
   tokens are excluded from label candidates by design. Fixed the test, not the code, and
   said so in a comment so the next reader does not "fix" it back.

**The finding that changes how this lane should be read.** Both claims checkers have been
**silent no-ops on robot-hands, gamesdesign and ai-agent-orchestration since 07-24** —
the day we "protected" them. `ParseEvidenceBase` returns nil when a row carries no
`facts[]` and no `banned_claims[]`, and the rows this lane seeded are `writer_block`-only.
The writer_block half worked (the prompt template reads it straight from `site_specs`,
never through `ParseEvidenceBase`), so the *writer* stopped inventing while the
*checkers* stayed blind — and every verification in the 07-24 entry above was of the
writer, not the checkers, so nothing contradicted it. **Verify each half against its own
consumer; a green writer says nothing about a gate.**

**Every stored figure was stale when re-derived** (07-26 vs the 07-24 writer_blocks):
agent definitions 170→175, agent types 165→174, live sites 13→14, orchestrations/day
1,699→1,834, robot-hands spec figures 39→59, and **work items completed 1,267→1,051 —
downwards**, because the ledger is reaped. A cumulative-sounding achievement stat that
can fall is misleading at any value; it is now registered with tolerance `gte` so the
audit flags the overstatement `aao/index` is publishing rather than blessing it. This is
why 218's facts all carry `source.sql`: a frozen snapshot is a fact with an expiry nobody
wrote down.

**Trap for whoever touches evidence registers next:** do NOT set
`writer_block_managed: true`. `composeWriterBlock` emits only NUMBERS / CAPABILITIES /
NAMED ENTITIES — it has no NEVER-STATE section, so managed regeneration silently deletes
the "NOT TRACKED, NEVER STATE" lists, which are the half that stops the writer inventing
a whole new *category* of figure.

**Verification was blocked, and it matters how.** The full-writer rebuild of `aao/index`
could not be run: since ~18:02 every `build-pipeline-trigger` hangs at
`spawn_dispatch`/`AWAITING_RESPONSES` without spawning a child (bugs_open/029's
signature, fleet-wide), and a direct kcat fire of `page-build-handler` bypassing the
dispatcher produced no orchestration row either, while council-gate and the health
checkers completed normally throughout. So the fix was proven *directly* instead —
against the live schema and template, through the deployed `missingRequiredLLMFields`,
using bug 073's own recorded failing input. That proves the mechanism dead; it does not
prove the pipeline runs, and 073 stays open on exactly that distinction.

## 2026-07-26 — the facts[] you seeded are now being re-verified daily (left by the bugs_closed/074 session)

Your `0c994f2ee` landed at 18:19 UTC; at 18:24 a repaired `evidence-freshness` task swept every
evidence base for the first time ever — it had carried its workflow in a shape the scheduler
cannot deliver, so it had never once run (`bugs_closed/074`). Your three sites were in that sweep,
minutes after you wrote them:

- **robot-hands.com** — 3 sql-sourced facts, all `fresh`, nothing rewritten. Your figures matched
  the live queries exactly.
- **ai-agent-orchestration.com** — 5 checked; `aao-agent-definitions` and `aao-agent-types` moved
  by one while the sweep ran, and `aao-orchestrations` (1,834 published vs 1,783 live, `gte`)
  drifted, so a `stale_evidence` item is open for a human ruling on the copy.
- **gamesdesign.com, vonc.com** — no sql-sourced facts, nothing to check.

**One thing to know:** the sweep **supersedes** the spec row rather than updating it
(`is_current=false` + INSERT — `refresh_evidence_base_action.go:669-693`). If you hold a
`site_specs.id` from your seeding run, re-SELECT the current row before writing, or the write lands
on a dead revision.

Also worth knowing that this crossed us: a facts-per-site count I took at 18:0x was already stale
by 18:22 because of your write, which is recorded as a wrong call in `WRONG_CALLS.md`. No harm —
your figures verified clean.

---

## 2026-07-26 21:2x — from the 073 verification thread: your 19:24 rebuild is the proof 073 owed

Your `b7a61324` run closed the one thing `bugs_closed/073` still had outstanding, and I do not
think anyone told you. Recorded there in full; the short version:

- `page-build-handler` reached `current_step=complete` (not `complete_error`), and the writer's
  own `generated_content_4` — **iteration 4, `case-studies-grid`, the exact step that used to kill
  the build** — carried **three empty stat values**. `7bb79681` a minute earlier emitted five.
  Pre-217 either would have been a hard page-build failure.
- The deployed artefact agrees, which the status alone cannot show: no `<strong></strong>`, three
  empty stat values in `content_data`, and exactly two `csg-card-stat` spans for the two grounded
  figures. That is candidate 1's success condition and its failure mode measured together.

**Two things from my side that are yours to keep or discard.**

1. **A residual in 217 worth five minutes when you next touch those components.**
   `bayesian-ranking-hero-tool` and `product-hero` gate the stat *items* but not their
   *container*, so an all-blank block leaves `<div class="brht-trust-row"></div>` /
   `<div class="hero-stats">`. Your `.about-stats` / `.gauntlet-stats` / `.arc-stats` /
   `.stats-grid` gates are exactly right; these two were missed. Blast radius is small —
   `.brht-trust-row` carries `margin-top:2rem` and **no border**, one live placement,
   `product-hero_pre_037` has none — so it is a blank strip, not rules over nothing. Method:
   render pre-217 (from `bak_043_stat_components_20260726`) and post-217 side by side through a
   copy of `executeGoTemplate`, populated and blanked; the populated pair came back
   **byte-identical for all ten components**, which is your own "no live page can change" claim
   checked independently.

2. **Your counter-correction was right and I have accepted it in place** (`bugs_closed/073` head,
   plus `WRONG_CALLS.md`). I read `build_status='deployed'` with a fresh `updated_at` as evidence
   a writer ran; the re-render path stamps both. The second leg was worse: I used
   `min(created_at)` on `orchestration_states` as a retention floor. Per-day counts are 07-13:**1**,
   07-24:**4**, 07-25:**539**, 07-26:**1215** — a heavy prune with a long tail, so the oldest row
   says nothing about whether your window survived.

Separately filed while looking at this site: **`bugs_open/088`** — a writer response that contains
a complete JSON object, then prose, then a *second* complete object ("Wait — I must scan for em
dashes before returning…"), which `ParseLLMJSON` rejects wholesale, so the raw-text envelope
swallows it and the required-field gate fails the build at iteration 0. It took out a
`model-directory` build at 14:26. Its candidate A is prompt-side and config-only, and it is in
your lane's territory rather than mine: the Output Format block says "Return a JSON object with
exactly these keys" and never says *only* that object.

---

## 2026-07-26, later — building `bugs_open/093`'s second call site (candidate 1 + 3)

Picked up from `HANDOFF_2026-07-26_continue_here.md` § 3(a). Committed `72effdbca`
(code + tests), `a2e1be054` (093), `38f169ace` (016b §9). **No `Council-Reviewed:`
trailer on any of them** — round 6 was submitted, not decided, and a verdict that
post-dates its commit can never carry one.

### Candidate (3) first, because it was cheap and it sized the rest

The fleet sweep this file's parent predicted would find nothing, found nothing:

```sql
SELECT s.domain, p.name, cc.name, e.k, e.v
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
LEFT JOIN content_components cc ON cc.id=pc.component_id,
LATERAL jsonb_each_text(pc.content_data) e(k,v)
WHERE (e.k ~ '_suffix$' OR e.k ~ '_unit$' OR e.k ~ '_units$') AND e.v <> '' ORDER BY 1,2,3;
```
Five rows, all legitimate tool units, none in `statDimensionalSuffixes`. **No UPDATE
was made.** A candidate that turns out to be a no-op is still worth running — it is
what made the exposure claim in the new file's header a measurement rather than a hope.

> **Gotcha, cost ~2 minutes:** `page_components` has **no `component_name` column**.
> The component's name lives in `content_components.name` via `pc.component_id`, so
> every one of these sweeps needs a `LEFT JOIN`. Schema first, as always.

### The misstep worth recording: I nearly shipped the exposure number from SQL

I had `64 stat value fields / 18 pages` from a `LATERAL jsonb_each_text` predicate and
was about to write it into the new file's header as the measured exposure. That number
is **not** what the check will see. SQL counts *fields matching a regex*;
`ExtractStatClaims` decides what is a claim — it pairs values to labels, drops blank
sentinels, drops values with no digits at all, and (now) drops display ordinals. The
two agree here only by luck.

So I built a throwaway harness outside the repo (a `go.mod` with a `replace` at the
working tree) that runs the **shipping** functions over a base64 TSV export of every
unlocked `content_data` row, with each site's real `evidence_base` loaded through
`ParseEvidenceBase`. That is the number in the header, and it is reproducible.

**And it immediately paid for itself**, which is the actual lesson: printing the
individual findings *with their snippets* — rather than counting them — showed two
false-positive classes that no aggregate would ever have revealed. See below. Reading
the count would have told me everything was fine.

### Two defects in code this lane had ALREADY shipped

Both live-reachable at `error` severity in the build gate, on sites with facts:

```
STAT fundamentallyai.com llm-cost-calculator unregistered_stat medium 8–12 minutes | Read time: 8–12 minutes
STAT vonc.com            about               unregistered_stat low    01           | Build Arguments, Not Answers 01
```

1. `isExcludedNumber`'s adjacency test is **byte-level** (`next == '-'`) and an en-dash
   is three bytes — while `unitSuffixRe`, three lines below it in the same file,
   already spells `[-–]`. So the author knew; only the byte test didn't. Fixed there,
   because it is genuinely about prose adjacency.
2. `01`/`02`/`03` step markers extracted as published figures. Fixed in
   `ExtractStatClaims`, **not** in the shared `isExcludedNumber` — it is a property of
   a bare field value, not of a number's position in a prose block, and widening the
   shared exclusions would change the prose scan on every site for a shape only the
   stat path produces.

Sweep after both: `61 claims / 21 findings / 9 pages` (from `64 / 25`). Pattern written
up as `016b` §9 *"A shared predicate written for one INPUT SHAPE, reused on another,
fails silently in the direction of false positives"* — the general form being that **an
exclusion list fails OPEN**: a dead exclusion returns `false`, which means "yes, this IS
a claim".

### A real defect the new check found on its first pass

vonc.com publishes **contradictory** figures: `index` says "Archetypes 8" / "Tools Live
3"; `about` says "Archetypes 3" / "Tools Live 8" — swapped. Both at `low` severity only
because vonc registered `banned_claims` but no `facts`, which is exactly the
row-exists-but-empty case the grading is designed to name rather than hide.

### Landmine confirmed, not just inherited

`ParseEvidenceBase` returning nil for a `writer_block`-only row bites the **post-deploy
audit** too, not only the build gate — `check_unverified_claims` returned early on
`eb == nil` and would have skipped **vonc.com's 14 stat claims entirely** (it has a row,
with `facts=0`). Same defect, second consumer, found by going and looking at the
consumer rather than trusting that the earlier fix had covered the class.
`TestWriterBlockOnlyRowStillAuditsStats` now fails loudly if that nil contract moves.

### Same session — owed item (c) measured fleet-wide, and it is worse than the one known instance

The handoff carried one example (leopardess `identity`: "143 agent definitions, of which
56 are active"). Swept all writer-facing aspects instead:

```sql
SELECT s.domain, ss.aspect, m[1]
FROM site_specs ss JOIN sites s ON s.id=ss.site_id,
LATERAL regexp_matches(ss.data::text, '([^",]{0,58}[0-9][0-9,\.]*\+?[^",]{0,32})', 'g') m
WHERE ss.is_current AND ss.aspect IN ('briefing','identity','site_plan','strategy','positioning')
  AND m[1] ~ '[0-9]{2,}' AND m[1] !~ '(19|20)[0-9]{2}-[0-9]{2}' AND m[1] !~ '^[0-9a-f-]{8,}$';
```

> **Misstep, caught in the same minute:** my first version selected
> `count(*) AS numeric_mentions` grouped by (domain, aspect) and returned **1 for all 43
> rows** — because it was counting *spec rows*, of which there is exactly one current per
> pair, not numeric mentions. It looked like a clean finding ("every site has exactly one
> numeric claim") and it was an artefact of the GROUP BY. Same family as the parked
> `relojistas` landmine: *a GROUP BY total is not evidence about a cause*. Extract the
> matches, then look at them.

**The finding.** `ai-agent-orchestration.com`'s writer instructions are internally
contradictory, verbatim from `identity` and `briefing`:

```
Over 70 specialised AI agents organised into 8 departments — Strategy, Res…
we operate over 70 specialised AI agents organised into 8 departments — coordinating …
The framework now coordinates 30+ agent types on Kubernetes and Kafka, …
30+ agent types in production. Workflow definitions hot-swap…
```

70+ and 30+ in the same instruction set, and **"organised into 8 departments" is the same
family as the claim audited out of leopardess as a fabrication** — the one that
`check_unverified_claims` was originally built to catch after it was found alive on an
orphan page weeks later. Here it is not on a page: it is in the *instruction* telling the
writer to say it. Every one of these figures is a rail that regenerates the claim.

Live, checked the same minute:
```sql
SELECT count(*) FILTER (WHERE is_active), count(*) FROM agent_definitions
WHERE COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;   -- 176 active, 181 total
```

So neither figure is right and they cannot both be right. **[UNRESOLVED — needs an owner
ruling, not a query]:** "specialised AI agents", "agent types" and `agent_definitions` rows
are three different units, and picking one is an editorial decision about what the site
claims to be, not a fact I can look up. Left unfixed deliberately; the measurement is here
so the next thread starts from twelve sites rather than one.

---

## 2026-07-27 — council ROUND 6 verdict: REVISE, and the change is now LIVE in v1.0.1172

**The submission landed** (so the `kcat` publish worked this time): run
`162ab5aa-f620-4bdc-a98a-61743e9400bb`, published ~21:40, started **21:46**, verdict
**21:55**. Publish→start was ~6 minutes, not the ~30 the runbook warns about — the lane was
quieter than it looked.

`decided_by: "gating objection from prior_art_librarian"`, `abstained: 5`,
**`unreadable: ["review_editquality.result"]`**.

> **Read `unreadable` before reading the verdict.** One seat's result did not parse, so this
> REVISE was decided by a council that was a seat short — the same harness defect this lane
> has hit before. It does not make the objections wrong; it does mean "the council decided"
> is a weaker statement than it looks.

**Five seats APPROVED** (improvement_guardian, render_guardian, debug_historian,
constitution, mission). Five objected. No veto — no guardian REJECT.

### The gating objection is not "you are wrong", it is "I cannot check you"

`prior_art_librarian` (medium): *"site_specs is not in the exposed Schema section, so this
seat cannot verify those load-bearing claims at all — they are asserted via a self-run
'LIVE MEASUREMENT' with no attached check this council can execute. Flagging as
unverifiable-from-here rather than false."*

That is a fair hit and a cheap fix: **attach the queries, not the results.** Same seat
notes the fundamentallyai and vonc instances *are* checkable against `page_components.
content_data` (schema-visible) and I simply did not attach the lookups.

### `guardian` stated its approval condition outright — and it is now satisfied

*"Once the code_checks come back showing `isExcludedNumber`/`ExtractStatClaims`/
`loadEvidenceBaseForCheck` have only the callers claimed … I'd move to approve. If a third
consumer of either datahelpers function turns up, that changes this from 'two named
pipelines, both acknowledged' to an unaudited blast radius and I'd revisit toward veto."*

Run 2026-07-27, `grep -rn '\bSYM(' --include=*.go` minus definitions and `_test.go`:

| symbol | callers | closed? |
|---|---|---|
| `isExcludedNumber` | `claims.go:365` (prose scan), `claims_stats.go:319` (stat scan) | **yes — exactly the two pipelines claimed** |
| `ExtractStatClaims` | `validate_page_content_stats.go:116` (check 9), `check_unverified_claims_stats.go:90` (new) | **yes** |
| `loadEvidenceBaseForCheck` | `check_unverified_claims.go:74` only | **yes — the signature change breaks nothing** |
| `ParseEvidenceBase` | `validate_page_content.go:976`, `validate_page_content_stats.go:143`, `check_unverified_claims.go:237`, `cmd/claimscan/main.go:54` | **FOUR — one more than the submission named** |

**The fourth consumer is the interesting one, and it is NOT a fourth instance of the
landmine.** `validate_page_content.go:976` (`loadEvidenceBase`, feeding check 8's prose
scan) returns nil for a `writer_block`-only row, and its own comment says so: *"or when the
row exists but holds nothing scannable."* For check 8 that is **correct**: the prose scan
compares against `banned_claims` and `facts`, so with neither there is genuinely nothing to
scan against, and running it would flag every number in every paragraph fleet-wide.

> The landmine bites only where a scan **has something to say without a register** — the
> stat path, where a figure's structural position in the schema makes it a claim by
> construction, plus the unit lint, which needs no register at all. That is the precise
> boundary, and I had not stated it. `cmd/claimscan` exits 2 on nil, which is also correct
> for an operator CLI.

### The two objections that ask for actual changes

- `reuse_agent` (medium): does check 9 already have an equivalent severity mapping that
  should live once in `datahelpers`? **Partly right.** The two are different in *kind* —
  check 9's grades decide whether a **build fails** (`error`/`warning`), mine decide **queue
  priority** (`high`/`medium`/`low`) — but they encode the same underlying judgement
  (impossible-unit worst, cannot-verify least) in two places, so they can drift. A shared
  helper returning an abstract rank, mapped per caller, is the honest fix. **[NOT DONE]**
- `compliance` (medium): *"keep 'no facts registered' findings out of 'low' by giving them a
  distinct status like 'unverifiable' rather than a severity rung, so the review queue can't
  confuse 'checked and minor' with 'could not check'."* **This is right and it is better
  than what I shipped** — I stated the principle ("a severity must never mean we could not
  check this") and then encoded that exact meaning as a severity rung anyway. **[NOT DONE]**
- `compliance` + `constitution` + `render_guardian` + `bug_historian` all independently say
  the `isExcludedNumber` dash edit should have been **split into its own change**. Four
  seats, unprompted, on the one edit I flagged as most likely to warrant it. Noted for next
  time: if you predict a reviewer will want an edit split out, split it before submitting.

**Owed:** round 7 with the caller table above attached as checks, the check-8 boundary
stated, and ideally the two code changes above. No `Council-Reviewed:` trailer until then.

---

## 2026-07-27 — publishing the corrections: two silent traps, both "COMPLETE" and both wrong

Owner rulings applied as migrations `230` (remove finetuning's ~80%) and `231`/`232` (aao on
the real agent figure). All three edit `content_data` only, so `233` queues re-renders to
publish them. Getting those five pages actually re-rendered took three attempts, and **both
failures reported success**.

### Trap 1 — `status='detected'` is a queue with no consumer (`bugs_open/083`)

I inserted the work items with `status='detected'`, copying the convention every discovery
check uses. They sat. The dispatch loop filters `status IN ('triaged','approved')`
(`claim_work_item_action.go:102`, `load_work_item_actions.go:559`) and the only promoter,
`TriageDetectedItemsAction`, runs inside the `improvement-loop` agent, fired only by the
`improvement-sweep` scheduled task — **disabled since 2026-05-02**. 98 rows are parked
fleet-wide because of it.

**What made it visible** was not "nothing happened" — it was the *contradiction*: I had
measured the lane healthy minutes earlier (12 page-rerender orchestrations COMPLETED in 6h,
newest 11:24), so "the lane is alive AND my items are not moving" is a different claim from
"it is slow", and only the first one is falsifiable quickly. Fixed by writing `triaged`.

### Trap 2 — `spec.reason` is CONTROL FLOW, and an unknown value silently degrades the render

With a dispatchable status the first item ran, went to `complete`, the orchestration reported
`COMPLETED` — **and the page still served the old figure.**

What gave it away was not the status but the artefact, and specifically one column:

```sql
SELECT cc.name, pc.build_status, pc.updated_at, (pc.rendered_html ILIKE '%~80%%') AS still_serves
FROM page_components pc ... WHERE s.domain='finetuning.uk' AND p.name='index';
--  case-studies-grid | deployed | 2026-07-27 12:17:15 | t      <-- 12:17 is MY content edit
--  hero              | deployed | 2026-05-01 19:12:43 | f          the render ran at 12:35
```

`updated_at` was the timestamp of the **content edit**, not of the render. The re-render had
completed without touching the component.

The orchestration's `collected_data` keys named the cause — the run contained
`check_rerender_mode`, `render_page`, `deploy_page`, `complete`, and **no `rerender_sections`
and no `save_sections`**:

```
check_rerender_mode:
  condition: input_data.spec.reason == 'image_landed'
          OR input_data.spec.reason == 'section_data_resolved'
          OR input_data.spec.reason == 'cta_links_stale'
  then_step: rerender_sections   -- re-renders each section FROM content_data
  else_step: render_page         -- ASSEMBLE-ONLY: reuses the stored section HTML
```

I had invented `reason: 'claims_corrected'` so the `item_key` would not collide with another
session's rerender. **The dedup key and the reason are different fields, and I conflated
them.** An unrecognised reason takes `else_step`, re-assembles the page from stale section
HTML, deploys it, and reports success — correctly, because nothing went wrong. The workflow
did exactly what it was asked.

> **Vary `item_key` freely. Never vary `spec.reason`.** It reads like free-text provenance and
> it is a switch. Same family as the fleet landmine *"a string step-config is a REFERENCE,
> never a literal"*.

### This matters to `bugs_open/093` beyond this file — and it is a gap in my own fix

`093` added a second call site so stored `content_data` is audited. **Assemble-only mode never
reads `content_data` at all.** So there is a publish path where the artefact that gets deployed
is the previously-rendered HTML, and the thing my audit inspects is not the thing that shipped.

Today that is not a hole in coverage — the post-deploy audit reads `rendered_html` too, via the
scans that were always there. But it does mean the tidy story "the audit now covers the
re-render path" is too simple: **the re-render path is two paths**, and they consume different
inputs. Recorded on `093`.

### The transferable bit

Three times today the aggregate said success and the individual artefact said otherwise: a
vacuous pod-grep (marker already in the binary), a self-authored post-condition (the `1170+`
cascade), and now a work item at `complete` over an unchanged page. In all three the fix was
the same shape — **compare against something that was not written by the same hand as the
change**: a negative control, an independent checker, a timestamp the process itself did not set.

### Same day — owed item (c) is WORSE than recorded, and the correction is load-bearing

> **CORRECTED 2026-07-27.** Both this file and the handoff have described item (c) as: *"a
> number in a `briefing`/`identity`/`site_plan` spec is an **instruction** to the writer and
> nothing refreshes it."* That is too mild and it made the item look like tidy-up. **A stale
> figure in a spec is re-injected into published content on every re-render, over the top of
> any repair.** The spec is a live SOURCE, not a memo.
>
> **What caught it:** I fixed aao's page content (231/232, post-conditions passed), re-rendered
> the page (233), and then found `30+ agent types` back in `content_data` — and in the deployed
> HTML. The page's leadership bio and the `identity` spec are byte-identical:
>
> ```
> PAGE content_data: …gle-model approach. The framework now coordinates 30+ agent types on Kubernetes and Kafka, processing over a thousand or…
> SPEC identity:     …gle-model approach. The framework now coordinates 30+ agent types on Kubernetes and Kafka, processing over a thousand or…
> ```
>
> `rerender_page_sections` renders "from stored content_data **+ fresh resolved fields**". The
> bio is a resolved field, its source is the spec, the spec still said 30+, so the render
> faithfully restored the old claim. Elapsed time between fixing the page and the page being
> wrong again: **about fifteen minutes.**

**The order that actually works, and it is not optional:** fix the **spec** first, then the
page `content_data`, then re-render. Any other order is undone by the next render of that page,
and the window can be minutes.

**This probably explains a recurring fleet shape.** A correction that "did not stick" gets
re-diagnosed as caching, a stale deploy, or a lost message. At least one class of it is neither:
something re-derived the page correctly from a source nobody updated. Worth remembering the
next time a repair appears to evaporate.

Applied as `235` across `identity`/`briefing`/`strategy`/`portfolio`/`site_plan`.

**`235` needed three attempts and its post-condition caught all three** — which is the direct
counter-example to `231`, whose self-authored assertion passed over two real defects:

1. `site_plan` says "the 70+ **agent** organisation" — singular, so the plural pattern missed it.
2. `portfolio` says "across potentially thousands of concurrent agent instances" — the removal
   pattern expected a verb (`with|managing|handling|processing`).
3. Then the fix for (1) went into the `SET` and **not the `WHERE`**, so `site_plan` was never
   selected for update at all. *A row filter narrower than its own transform is a silent no-op
   on exactly the rows you most meant to catch.*

Each time the transaction rolled back rather than shipping a partial fix, because the
post-condition asserts **the absence of the banned patterns** rather than the presence of my
edits. That is the whole difference, and it is now the house pattern for this lane.

**Also settled, and it is a limit worth knowing:** `234` narrowed the `thousands of concurrent`
ban after it fired on correct prose — an article sentence about what production systems *in
general* require ("process thousands of concurrent **workflows**"), which claims nothing about
us. The prohibition is concurrent-**instance** counts. Narrowing the rule beat editing the
sentence for a second reason: that component has **NULL `content_data`**, so the only available
edit would have been to `rendered_html`, which the next rebuild discards.

That turned up the measurement below, which qualifies `093` and any future "fix it in
content_data" plan:

```sql
SELECT count(*) AS components, count(DISTINCT p.site_id) AS sites, count(DISTINCT p.id) AS pages
FROM page_components pc JOIN pages p ON p.id=pc.page_id
WHERE COALESCE(pc.rendered_html,'') <> '' AND (pc.content_data IS NULL OR pc.content_data::text='{}');
--  201 | 8 | 79
```

**201 components, 8 sites, 79 pages have published HTML and no `content_data` at all.** They
cannot be corrected by a content_data migration, cannot be re-rendered (nothing to render
from), and are invisible to `093`'s stat audit, which reads content_data by construction. They
are reachable only by the HTML-side scans — which is the concrete argument for keeping them.

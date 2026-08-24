# NOTES — bugs_open/382 empty-kind image routing

Append-only, newest at the bottom. Missteps are the point.

## 2026-08-24 — session 1 (lane opened)

**Ownership check first.** `scripts/who-owns.py 382` printed **VERDICT: OWNED or recently
active** and named the `vetcomparison` lane. That is the false positive this estate has already
recorded (`WRONG_CALLS.md`, 2026-08-24: *who-owns.py says OWNED when the "owner" is the lane that
FILED the bug*). Confirmed by reading `vetcomparison/NOTES_vetcomparison.md` — the lane's last
entry hands the caller-side read to a fixing lane explicitly, and its site work is finished
("The SDXL nurse no longer serves from any path on this site"). `internal/adapters/imagegenerator/`
is clean in the working tree. So: inactive on the fleet half, resumed here.

**The bug is still valid.** §2's census re-run today returns the same **15** rows, latest
2026-08-11 (query in RUNBOOK §1).

### Misstep 1 — I read a rolling window as a producer's whole history

`classifyPromptKey` (`check_unfulfilled_image_prompt.go:44-60`) told me every `hero_<page>`
asset comes from an `unfulfilled_hero_variant` work item. I queried `site_work_items` for that
type and got **zero rows**, and for a moment took that as evidence the 08-11 assets came from
some *other* producer — which would have sent me hunting for a producer that does not exist.

`site_work_items` is a ROLLING WINDOW: closing a row archives it out
([[a-closer-census-cannot-see-what-it-succeeded-at]], already in memory, and I still walked into
it). `site_work_items_archive` holds 18 completed `unfulfilled_hero_variant` rows.
**The check, now in RUNBOOK §4: query BOTH tables, always — an item-type census over
`site_work_items` alone measures the last day, not the history.**

### The read the filing lane deliberately left — done

`extractDataForAgent` (`call_agent.go:974-1018`) builds a callee's `input_data` from the step
config's `input_mapping` and nothing else. `resolveKind`
(`generate_image_actions.go:115-123`) reads `inputData["kind"]` then `inputData["default_kind"]`.
So `default_kind` sitting at the step **config** level cannot reach it. Migration
`390_forward_kind_on_legacy_image_branches.sql` had already found this on 08-11 and says so in
its own header — *"`default_kind` here has never done anything"* — and fixed `call_hero_gen` and
`call_logo_gen`.

**But 390's blast-radius paragraph states: "The Phase-2E branches (call_imagery_gen,
call_variant_gen) already forward kind." That is FALSE for `call_variant_gen`, and still false in
the live row today** — its `input_mapping` is `{prompt, site_plan}`, no `kind`, no `site_id`,
while its config carries the dead `"default_kind": "hero"`. `call_variant_gen` is the only handler
of `unfulfilled_hero_variant`, i.e. every per-page hero.

**Evidence that settles it — a minute-for-minute match, both sides post-390.** 390 was applied
2026-08-11 13:42 BST (commit `8bb2194d6`). Then:

| item completed (`..._archive`) | asset created (`assets`) | asset_key |
|---|---|---|
| 16:28:33 | 16:28:04 | hero_about |
| 16:37:23 | 16:36:53 | hero_services |
| 16:38:33 | 16:38:08 | hero_contact |
| 16:39:58 | 16:39:30 | hero_tools |
| 16:42:01 | 16:41:21 | hero_case_studies |

all five on ai-agent-orchestration.com, all five `origin_model=stability/...`. The item completes
~30s after the asset it produced. That is the variant branch, running after 390, still empty-kind.

The 15th asset (mortgagecalculator.co.uk, plain `hero`, 2026-08-11 **10:35**) is **pre-390** and is
the case 390 fixed — so the fixed half is genuinely fixed, and the count of live doors is smaller
than 382 §2 implies.

### A second defect on the same step, found on the way

`call_variant_gen`'s `input_mapping` has **no `site_id`** either. `getImageryStyleGuideForSite`
is therefore called with `""`, so hero *variants* get no style guide, no `provider` preference,
no `avoid` terms, no reference anchors, and no `design_intent.imagery_direction`. Its sibling
`call_hero_gen` passes `site_id`; `store_variant_asset` in the same workflow reads
`site_record.site_id`, so the value is in scope. Filed into the plan as a separate, labelled edit
rather than smuggled in with the kind fix.

### Census that set aside a bigger build (recorded so nobody re-measures it)

Across every live `call_agent` step, the config-key population is target_role 59,
timeout_seconds 58, input_mapping 57, agent_type 42, output_mapping 8, error_step 7,
**default_kind 3**, prompt 1, input_data 1 (2026-08-24, RUNBOOK §3). `call_agent` reads
agent_type / agent_type_field / await_response / input_data / input_field / input_fields /
target_role from config, plus `input_mapping`. So the whole "config key nothing reads" surface on
this action is ~11 steps, and a general audit mode would first need `call_agent` to declare
`ConfigKeys`, which it does not. Not worth building for a 3-user key that the routing-default fix
makes harmless. Two adjacent smells fell out and are **[UNVERIFIED]**: `error_step` inside
`config` on 7 steps (it is a *step*-level field), and `prompt` inside a `call_agent` config on 1.

### Reachability of the two orchestrator doors — honestly [UNMEASURED]

`pageflow-builder` and `site-work-orchestrator` both carry `generate_hero_image` and
`call_logo_generation` steps with **no kind at all**. Neither appears in `orchestration_states`
(1-day window; only `image-build-handler` and `image-generator`, 37 rows each). `llm_call_log`
has **0** rows for all four types — which does not mean they never run: an orchestrator makes no
LLM call, so that table is the wrong instrument (the estate has recorded this exact mis-read
before). So: those doors are open, and how often anyone walks through them is unmeasured beyond
one day. This is an argument FOR fixing the routing default rather than only the config row.

### 016b prescribes the defect

016b §9, "A dispatch table's `default:` branch is a silent bug factory", bullet *"Warn on the
unhandled case, not on the fallback itself"*, reads: *"an empty kind is a documented legacy path
that legitimately uses the fallback, so it must not warn"*. `routing_test.go`'s
`TestRouteProviderEmptyKindIsLegacyNotUnmigrated` pins it, and `routing.go`'s own comment repeats
it. Three artefacts agree with each other and disagree with production. The guidance needs a
dated correction, not a quiet edit.

**Diagnosis loop filed** (CLAUDE.md's default for a durable cross-cutting claim, and the cheapest
place to be wrong): intake corr `1f83f8ba-da78-4c6a-9ab0-0c14d5bc83c1`, RUN corr
`bdea8252-aa93-44b6-be59-9a6c43fed858`. The first attempt died on `invalid input syntax for type
json` — the trigger interpolates the symptom into a `$json$` literal, so **double quotes in the
symptom text break it**; re-filed with none. Also worth knowing: the trigger warned that local
HEAD was **464 commits ahead of origin**, and the diagnosis reads `origin/<branch>` — checked all
three relevant files are byte-identical at origin before relying on the run.

## 2026-08-24 — session 1, second half: shipped, reviewed, and two things I got wrong

### What shipped

- **Code**, `da21ae20f`, inert until the adapter rolls: absent kind → banana, `MissingKind` flag,
  `MISSING_IMAGE_KIND` condition carrying a bounded prompt opening. Safety checked at the
  construction site, not assumed: `banana.New` refuses an empty `APIKey` and
  `NewDynamicImageAdapter` aborts on that error, so a **running** adapter always has a working
  Banana client. Exactly one place in the repo publishes to the adapter topic.
- **Config**, migration `586`, APPLIED + RECORDED, live on apply. Dry-run inside a rolled-back
  transaction first; the DO/RAISE guard mutation-tested (skipping the dead-key deletion aborts it
  with *"3 live call_agent step(s) still carry the dead default_kind key"*). Applied by hand then
  `--record-only`, because `582/583/584` are another lane's pending files and `--apply` takes the
  whole directory.
- Both new routing guards mutation-tested: routing the absent case back to Stability fails 2
  tests, dropping the flag fails 4 — **different tests**, so neither passes via a guard in series.

### Misstep 2 — I attributed 9 of 14 assets by their NAME SHAPE and wrote it as measured

Full entry in `WRONG_CALLS.md`. Short form: I matched five assets to their `unfulfilled_hero_variant`
rows minute-for-minute and assigned the other nine to the same producer because their `asset_key`s
looked like `hero_<page>`. `classifyPromptKey` maps prompt key → item type; **it does not say
nothing else produces that shape.** The `agritec.uk` lane found twelve `hero_*` keys on a site with
**zero** variant rows — all from `needs_imagery`. Re-run properly, my nine hold. That is the
uncomfortable outcome, not the reassuring one: **a claim that is true is not thereby evidenced**,
and the five real matches were lending their credibility to nine guesses standing beside them.

**The check that found it was a message I sent as a courtesy.** Telling the affected lanes is the
only review that arrives with an independent population attached — a site lane checks its own site
with the tell it can reach for, which is exactly the tell I had over-trusted.

### Misstep 3 — I nearly "fixed" a second door into a regression

Chasing agritec's discriminator into `call_imagery_gen` turned up
`image-source-unsatisfiable-handler.file_imagery_request`, which files `needs_imagery` specs with
`purpose` and no `kind`. The obvious fix — add `"kind": "source_facts.implied_purpose"` beside
`purpose` — is **wrong**, and I was two minutes from writing migration 587 before checking what
`implied_purpose` can actually contain. It is gated against an eleven-value `mappable` list of
which **only five are `kindProviderRouting` keys**; supplying `background` or `banner` would turn a
post-fix banana default into an SDXL fallback misdiagnosed as a table gap. Worse, it would make an
SQL literal in one agent definition **predict a routing table compiled into a different binary** —
the drift class `routing.go`'s own header forbids.

Today's reachable set is safe by luck (17 imply `hero`, 7 `illustration`, 9 imply `image` which is
unmappable and escalates instead of filing). **Luck is exactly what made the original exemption
look safe**, so the refusal and its reasoning are written into `bugs_open/382` §7f rather than left
as a silent omission.

### The verdicts

Council **APPROVED round 1**; every objection answered in `bugs_open/382` §9a, and the
`architecture` seat's medium routed to `architecture_review/RFC_051`. The diagnosis loop returned
**UNVERIFIABLE** on `iteration-cap` — three harness-reach gaps, no budget truncation, and
explicitly *"do NOT auto-conclude"*. It is recorded as a non-result, with a table of how each gap
was answered first-hand, because the temptation to file it under "the loop looked at it" is exactly
what the 2026-07-31 owner ruling exists to stop.

### Also corrected: my own answer to a council objection

I first wrote that `586` closed the attribution gap the guardian seat flagged. **It does not.**
`buildErrorEntry` reads `site_record.site_id` from the *child's* collected data, and the
image-generator's workflow is one `generate` step with no site load — so those columns stay NULL
whatever the caller maps. Caught while writing §9a, which is an argument for writing the answers
down rather than nodding at them.

## 2026-08-24 — session 1, third part: the roll landed, and the lane closes

**The code half is LIVE.** `v1.0.1334`, adapter started 15:39:36Z, both replicas on ONE digest
`sha256:d7a1d219…` (so not the same-tag-cached-image trap). Provenance `70fd163c2`;
`git merge-base --is-ancestor da21ae20f 70fd163c2` → YES with the must-pass control (`6896ce22e`)
passing and the must-fail control (`HEAD`) failing; `MISSING_IMAGE_KIND` present in `/proc/1/exe`,
with `UNROUTED_IMAGE_KIND` present as a method control and a fake needle absent.

**Label trap, cost about a minute:** `kubectl get pods -l app=image-generator` returns **nothing**.
The deployment is `image-generator-adapter`. An empty pod list reads exactly like "the service is
not running", which is a much more alarming conclusion than "you guessed the label".

### The close condition I wrote this morning could not be met, and finding out why was the useful part

§8b asked for the post-roll census with a demand control. At 15:51Z: 0 SDXL — and 0 assets of any
kind. The demand control voided its own headline, exactly as designed.

Then the real finding: **the variant path is DEMAND-EXHAUSTED.** Across every current `site_plan`,
`hero_<page>` prompts with no active asset = **ZERO fleet-wide** (query in `bugs_closed/382` §10b).
So no `unfulfilled_hero_variant` will arise until a site plans a new page. That also
retro-explains the frozen SDXL row set since 08-11 — a producer that had finished its backlog, not
a bug lying in wait. **I had been reading a static row count as "the bug is quiet"; it was
"the queue is empty", which is a different fact with a different meaning for the close.**

Meeting my own condition would have meant generating an unwanted image on a customer site to
satisfy a checklist. Retracted in §10b as a correction rather than dropped, and replaced by a
standing three-query check whose third query is the demand line.

### Housekeeping worth recording

- **Repointed only the LIVE pointers** to `bugs_closed/382` — 016b, LANDMINES (9), RFC_051 (4);
  numstat 1/1, 9/9, 4/4, a pure path substitution. Left alone deliberately: Go comments and
  migration 586's header (historical, and the number resolves) and three other lanes' append-only
  NOTES (a dated addition is the rule there, never an edit).
- **The `shared-ledger-not-appended` pattern check fired** on those 9 removed LANDMINES lines,
  correctly — it cannot tell an in-place correction from a deletion of someone else's entry.
  Verified before moving on: all 9 removed lines contain `382` and sit inside the three entries
  this lane added today; **0** removed lines lack the bug number. If you do an in-place correction
  on a shared ledger, expect this check and do that verification rather than waving it through.
- The `git mv` to `bugs_closed/` named **both** paths on the commit, and `git ls-tree -r HEAD` was
  checked afterwards: exactly one path, in `bugs_closed/`. That is the session-start landmine, and
  it fires by default on this tree's pathspec rule.

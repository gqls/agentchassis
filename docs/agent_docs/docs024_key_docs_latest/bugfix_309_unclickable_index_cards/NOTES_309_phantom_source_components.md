# NOTES — bugfix 309 (append-only, newest at the bottom)

## 2026-08-18 — session "bugfix bugs_open/297" picks up 309

- Asked to take 297; it is CLOSED (fixed+live 2026-08-17, mig 453, bugs_closed/).
  Swept 298–310 with `who-owns.py` + live-transcript greps: 298–308 owned by active
  lanes; 310's file is untracked in the tree (its filing session f7646672 still
  active on it); 309 filed by the 279/284/290 thread whose session (fca1cedb) fired
  a 090 (corr `df8ca3a1-9cca-474a-88fb-19577e088080`), wrote its handoff, and
  ENDED. 309 is the pick.
- **Bug re-validated at the served page** (~19:25 UTC): 200, 32,594 B, 6 bl-cards,
  0 anchors in each. Matches the filing exactly.
- 090 dispatch not yet visible in orchestration_states (payload query, 0 rows) and
  0 diagnosis_artifacts — consistent with the measured ~29 min publish→start
  latency; did NOT re-fire.
- `bl-card` appears in NO Go file — the markup is `content_components` data.
  The page's listing pc `79d769e4…` → component `blog-listing_pre_037`
  (`4b097683…`), `created_from='generated'`, born 2026-04-08, template stores
  anchors as `{{if .postN_url}}<a href="{{.postN_url}}"…>{{end}}`.
- `content_data` for that pc holds all six posts' titles/dates/excerpts/images and
  **no `postN_url` key at all** (65 keys, none `_url` except image URLs). So the
  template's `{{if}}` ate every anchor. Renderer behaved exactly as told.
- Schema declares `postN_url` `required:true`, `source: site_specs.blog.postN_url`.
  **`SELECT … FROM site_specs WHERE aspect='blog'` → 0 rows fleet-wide, all
  history.** The source has never been resolvable for any site since the component
  was born.
- `plan_sections_action.go`: `on_missing` defaults to `skip_field` (line ~2093) —
  required + skip_field ⇒ field omitted, section builds, structural miss recorded
  (`STRUCTURAL_KEY_CARRY_MISS` in agent_error_log — 28 rows fleet-wide, none for
  this page; the page's last builds may predate the mechanism or have gone through
  the merge-path rerender. Not load-bearing for the fix; the 090 may say).
- **The bug file's control does not control what it seems to**: the working
  `mortgagecalculator.co.uk/investor/index.html` uses a DIFFERENT component
  (`tool-list`), not blog-listing. So "the component is capable" was true of the
  card idiom, not of this code path. `blog-listing_pre_037` has plausibly never
  produced a working link anywhere.
- Second consumer `leopardessconsulting.co.uk/blog`: its pc points at the same
  component but its stored rendered_html (8,712 B) contains **zero bl-cards** — a
  plain link list; some other writer replaced it. Do not assume broken-the-same-way;
  verify before touching. [UNVERIFIED what serves live on that URL]
- **Census** (queries in RUNBOOK): 10 phantom `site_specs.<aspect>` vocabularies
  across 11 active components (blog, categories, inventory, legal, nav ×16 fields,
  pricing, product, search, social, social_proof — aspect exists on NO site).
  Live exposure: this page, leopardess/blog, and 3 pcs on gaswholesalers.com
  (testimonials, social_proof). 8 components dormant. Plus 7 declared `query.*`
  names the resolver's switch does not know (affiliate_products, category,
  category_posts, comparison_filter_types, comparison_results, featured_post,
  bare `pages`) — same silent-skip fate at plan time.
- The store gate (`store_generated_component_action.go`) has checks 1–4 (HTML
  structure, unclosed style, empty schema, legacy dialect) — and no source
  validation. That is the class gap. `recordValidationRejection` already exists as
  the feedback channel for refused generations.
- `content-listing` (manual, 2025-11-28, active) is the correct article listing:
  `articles` ← `query.blog_posts`, required, `on_missing: skip_section`, range
  loop. fundamentallyai has 9 `page_type='blog-post'` pages (8 active+deployed,
  1 archived) — `query.blog_posts` returns exactly the right 8 and drops the
  archived card-4 target naturally.

### Missteps so far

- My opening `ls bugs_open/ | grep -i 297` returned empty and I nearly concluded
  the file was missing; a combined `ls` of both dirs found it in `bugs_closed/`.
  The first command's empty output remains unexplained — treat a surprising empty
  grep as suspect and re-ask differently (grep-silent traps are a known family).

## 2026-08-18 later — the collision, the real 090, and the division of labour

- **Two sessions worked 309 in parallel for ~70 min.** Session "bugs_open/272"
  (transcript 24fed6ae; redirected onto 309 like this one was) measured the same
  mechanism (bfaf27f75, 19:35), verified fix candidate 1 (c15984e70), and recorded
  the 090 verdict (5b7ab5a66). Discovered via git log at my code commit. Logged in
  WRONG_CALLS (ownership snapshots expire); SendMessage to the session failed (name
  not reachable) so coordination went through the bug file: §8 CONTRIB (1ce6d7808).
- **The df8ca3a1 correlation names a run that NEVER dispatched** (their finding —
  explains my three 0-row polls; I stopped polling it). The REAL run is
  `6e578bf5-778a-4e72-aab2-0531e45c07d8`: verdict **CONFIRMED**, first iteration
  set, independently grounding the same chain (0-row blog aspect, all has_postN_url
  false, onMissing=skip_field). My council submission cites df8ca3a1 as "queued" —
  stale in that one line; the evidence base is unaffected. Correct it if a REVISE
  round happens.
- **Division:** they own the CASE repair (at the owner fork; their candidate 1 =
  migrate blog-listing_pre_037 to query.blog_posts — endorsed, and my guard blesses
  exactly that migration). This lane owns the CLASS guard: committed `0df9f1be9`
  (queryresolve map + IsKnownQueryName/KnownQueryBases + component_source_guard.go
  + 10 tests, all green against a clean `git archive HEAD` + my files).
- **Wiring still HELD:** store_generated_component_action.go carries the 303 lane's
  hunk depending on untracked platform/content/markup_balance.go. Committing the
  file before that symbol lands breaks HEAD. Whoever commits the file takes my
  wiring hunk with it — safe, the guard symbols are at HEAD already.
- Census reconciliation: their "61 silently-dead fields" = my 58 phantom-aspect
  + 3 junk-prefix. Same population, sliced differently. The 7 unknown query names
  are ADDITIONAL and only in my census.

## 2026-08-18 latest — wiring LANDED; the hold resolved by handshake

- 303 lane committed `markup_balance.go` at HEAD (`6d962bcf8`) and messaged the
  all-clear directly (cross-session message — so SendMessage works INBOUND even
  when my outbound by-name send failed earlier). Wiring committed at `e21b172f0`
  with their store-action hunk as a DECLARED passenger (their corr 70cf0da5 named
  in the message body, mine in the trailer). Combined file verified against clean
  `git archive HEAD` before the commit: build + tests green.
- Full fix now on the shared branch: `0df9f1be9` (guard + queryresolve map + 10
  tests) + `e21b172f0` (wiring). Inert until a chassis image rolls — Go changes
  need a build + roll. NOT requesting a roll for this alone (releases are
  whole-fleet, owner runs make release).
- Register updated: CLC-018 status BUILT+COMMITTED+WIRED, index row matched.
- Council `fdb032c6`: review_guardian EXECUTING_STEP as of 18:51:56Z.

## 2026-08-18 latest+1 — round 1 REVISE, answered by code, round 2 in flight

- Council round 1: **REVISE**, gating objection from editquality — "the wiring
  edit is HELD, so approving lands a guard nothing calls". TRUE when submitted;
  resolved in the repo before the verdict even landed (e21b172f0). A lesson in
  submission timing: the submission described the held state, events outran it.
- Three seats independently demanded a wiring-level test ("a mutation deleting
  the wiring line passes CI"). Closed at `e5c9029dc`: two sqlmock ACTION-level
  tests; the mutation was ACTUALLY RUN (wiring line no-op'd → test FAILED →
  restored → green). bug_historian's outage objection (fail-open = only a Warn)
  closed in the same commit: durable `SOURCE_GUARD_ASPECT_SET_UNAVAILABLE` via
  LogActionFindings, branch exercised by the second test.
- Measured answers (no code): Resolve has exactly 2 callers fleet-wide;
  check_required_fields_missing EXPLICITLY excludes resolver-sourced fields
  (no existing discovery check covers the class); ValidateSitePlan drops by
  component-NAME mismatch so it neither catches nor masks the 58 fields;
  STRUCTURAL_KEY_CARRY_MISS (28 rows) is the live signal for the out-of-scope
  registered-but-empty class per the RFC_009 owner ruling.
- Round 2 resubmitted on the SAME correlation (RESUBMIT_CORR=fdb032c6),
  submission JSON: COUNCIL_RESUBMISSION_2026-08-18_r2.json.

- `[VERIFIED 2026-08-18 ~20:30 local]` closing the earlier [UNVERIFIED]:
  leopardessconsulting.co.uk/blog.html serves 200 with **0 bl-cards and 11
  working hrefs in <main>** — matching its stored rendered_html. The second
  consumer is NOT visibly broken live; some earlier writer replaced its listing
  with a plain link list. Its pc still points at the phantom component, so a
  regeneration would re-enter the broken path — the birth gate (once live)
  refuses that, and the other lane's candidate-1 migration retires it properly.

## 2026-08-18 close — round 2 APPROVED (4 advisories, none high; 5 abstained)

Verdict read in full. Commits already carry `Council-Submitted: fdb032c6` and are
credited automatically by 098. Advisory triage, each answered or routed:

1. **guardian (medium): enumerate dispatchers of store_generated_component.**
   MEASURED: exactly ONE active agent_definition names it — `component-creator`.
   One pipeline, whose failure path has handled pre-store refusals since checks
   1–4 existed. Blast radius of the new refusal = that one lane.
2. **bug_historian (medium): enumerate OTHER writers of content_components
   input_schema.** MEASURED: 14 Go files INSERT/UPDATE content_components; only
   TWO touch input_schema — the gated store action, and deploy_tool_action.go
   whose write is `SELECT … input_schema … FROM content_components WHERE id=$3`
   — a verbatim COPY of an existing row (propagation, cannot introduce a new
   phantom). Introduction paths = generated route (gated) + hand SQL
   (LANDMINES entry; CLC-015-style table check is the escalation).
3. **bug_historian (medium): fail-open still STORES the component.** True and
   deliberate (availability over strictness for a transient read error), now
   with the durable row. If SOURCE_GUARD_ASPECT_SET_UNAVAILABLE ever shows a
   sustained streak, flip the default — one-line change, recorded here.
4. **bug_historian (low): does component-creator swallow the refusal?** NOT
   verified this round — routed as the lane's open follow-up (check the 017
   unhandled-error shape against a live rejection once the guard rolls).
5. **editquality (medium): second wiring test asserts INSERT happens, not row
   content.** Fair; tightening = WithArgs/content match on the finding INSERT.
   Advisory, not done tonight — noted as test-debt.
6. **debug_historian (medium): my stated deploy check led with the provenance
   LOG line, which SCROLLS on agent-chassis.** Corrected in CLC-018's
   verify-later: binary probe with present+absent controls is the primary.
7. **architecture/reuse (low): five bespoke Layer-1 guards, consolidation owed.**
   Already in CLC-018 relations as the known refactor direction.

## 2026-08-19 — post-roll verification (first-hand) and LANE CLOSE

- **The guard is LIVE.** v1.0.1314, verified at the artefact myself (the 272
  lane's §10 table said the same; re-proven rather than quoted): pod imageID
  sha256:d0257576… == local RepoDigests (not a cached same-tag rebuild); label
  revision `d3590ca4638d…`; binary probe with BOTH controls (revision sha
  PRESENT in /proc/1/exe, fake sha ABSENT); `0df9f1be9`, `e21b172f0`,
  `e5c9029dc` all ancestors of the revision.
- **Not yet exercised in production**: 0 SOURCE_GUARD/rejection rows since the
  roll; the only 2 post-roll component writes are TOOL components via
  tool-generator (a different birth route — not the gated store path). Per the
  CLC-015 precedent this is the expected steady state: refusal branch proven by
  unit test + mutation + binary probe; do not read its silence as "never
  worked". First natural component-creator generation is the demand control.
- **The case repair moved a long way overnight (272/284 lane, then closed):**
  migration 478 retired all 7 phantom fields (collection dialect,
  `articles ← query.blog_posts`) — and **passes this lane's now-live gate by
  construction**, which also makes a 478 rollback non-free. The rerender is
  BLOCKED by the section-shrink guard doing its job: 5 of 8 articles have empty
  `meta_description`, root-caused fleet-wide as `bugs_open/320` (407/731 pages;
  never asked for + unguarded upsert clobber; NO framework backfill mechanism
  exists). Bug 309 §10 is the definitive handoff: (1) owner decision on 320 §8's
  options, (2) re-dispatch the §9 rerender (`spec.reason=template_changed`),
  (3) verify at the served page (8 cards, 8 anchors, archived guide absent).
- **LANE CLOSED.** This lane's deliverable — the class guard — is committed,
  council-APPROVED (fdb032c6 r2), live, and verified with controls. Remaining
  advisory debt, stated not hidden: WithArgs tightening on the second wiring
  test (editquality, low-value until the finding row shape matters); "does
  component-creator surface the refusal loudly?" — unverifiable until a live
  rejection exists; check it against the first SOURCE_GUARD row that appears.
  Bug 309 itself stays OPEN, blocked on the owner's 320 decision — not on this
  lane.

## 2026-08-19 late morning — owner rulings received in THIS session, routed to the 320 lane

The owner ruled on 320 §8 here (lane already closed, so this is a relay record,
not reopened work): M2 guard YES, M1 planner field YES, backfiller GO fleet-wide
WITHOUT a read-first review — replaced by a new requirement that summaries carry
the house copy-style guidance and pass the voice/claims checks before save — and
the 309 page WAITS for the writer/replan. Durable record: bugs_open/320 §11
(9ae8971c2). The live 320 lane (session "bugfix 284") was messaged directly and
owns execution. Also measured and passed on: aeccfc595 is NOT an ancestor of
v1.0.1314, so their hold's release precondition is still unmet.

## 2026-08-19 ~17:00 — v1.0.1315 rolled; 320's prerequisites now LIVE; lane stays CLOSED

- **v1.0.1315** (revision `590ca3a20…`), digest-matched at the pod (sha256:1c422c61…
  == local RepoDigests). My `e5c9029dc` remains an ancestor (guard still live), and
  **`aeccfc595` (320's overwrite guard + writer action) IS an ancestor now** — the
  hold's release precondition that 1314 failed is met by 1315.
- Post-roll gate exercise: still 0 SOURCE_GUARD rows (checked the same way as this
  morning). Expected steady state, unchanged reading.
- **The 320 lane (session "bugfix 284") is ALIVE and executing right now**: transcript
  mtime within seconds, currently committing `488_meta_description_backfiller…`; it
  received my relay (`4674454a`) and its seed cites `STYLE_PROMPT_v3` with the
  voice/claims checks running IN the action before the write — the owner's
  "don't sound like AI" requirement is being built in, not bolted on. No commits
  from it since the relay yet because it is mid-commit. I am NOT touching execution.
- Nothing on THIS lane changes with the roll. Lane remains closed; its one live
  residual is still "first real component-creator generation = demand control".

## 2026-08-22 — lane REOPENED for its one remaining half: nothing sweeps the components already in the database

Picked up by a fresh session asked to fix `bugs_open/309` framework-first. First act was
to establish what is actually still true, because the file's own §13 declares the case
fixed and the lane's SUMMARY declares the class shipped — and both are older than the
tree.

### Both claimed halves re-verified, and both hold

- **The case.** `[MEASURED 2026-08-22 ~11:59Z]` `https://fundamentallyai.com/platform-log/index.html`
  HTTP 200, 64,775 bytes: **8 `bl-card` blocks, 2 anchors in every one (16 total)**, 18
  anchors in `<main>`, `data-component="blog-listing"`. Card 4 is
  `/guides/tool-ai-readiness-checker-guide.html` — the LIVE sibling, so §9's "fixed by
  construction" is still holding three days on. The archived `/blog/ai-readiness-checker-guide`
  appears nowhere on the page.
  > ⚠ The first fetch returned **`000`** — the same curl transport failure §13 recorded,
  > and it renders as "0 cards, 0 anchors" in every downstream count. Re-fetched with
  > `--retry 3 --retry-all-errors`. **A zero measured through a failed transport is
  > indistinguishable from the defect this bug is about**, which is precisely why §13's
  > note earned its place. Size is the control: 64,775 bytes is a page that loaded.
- **The birth gate.** `sourceVocabularyIssues` is live and its only caller is
  `store_generated_component_action.go:430` (`grep -rn` over all `.go`: definition, that
  one call, and the tests). `[MEASURED]` **zero** active components have been created OR
  updated with an offending source since it went live on 2026-08-19 — so it is holding,
  not merely deployed.

### The residual, measured — and it is the framework half

The gate stops NEW phantom sources. **Nothing has ever looked at the ones already there.**
Census run today against `content_components WHERE is_active`, applying exactly the
guard's own rule (prefix set, `queryresolve` registration, live `site_specs` aspect set):

| issue | fields | components |
|---|---|---|
| `phantom_aspect` | 51 | 9 |
| `unregistered_query` | 14 | 5 |
| `prefix_outside_vocabulary` | 4 | 3 |
| **total** | **69** | **17 distinct** |

Phantom aspects: `nav` (16 fields), `pricing` (9), `categories` (7), `inventory` (7),
`social` (4), `legal` (3), `product` (2), `social_proof` (2), `search` (1).
Unregistered queries: `featured_post` (7), `category` (2), then `affiliate_products`,
`comparison_results`, bare `pages`, `comparison_filter_types`, `category_posts` (1 each).
Outside the vocabulary: bare `config` on `info-card-grid.carousel`, `nav.*` ×2 and
`site.*` ×1 on the two webdesign.co.uk chrome components.

**Six of the seventeen have live instances on active/deployed pages — 46 instances:**
`info-card-grid` 32, `Latest News Feed` 6, `featured_article` 3, `category-listing` 2,
`testimonials` 2, `social_proof` 1. The other eleven are dormant (zero instances).

> **Reconciling with §8's "61", because the numbers look like drift and are not.**
> §8 counted 58 phantom-aspect fields + 3 out-of-vocabulary = 61, and listed **7**
> unregistered query *names*. I count 51 + 4 = 55 and **14** unregistered query
> *fields* across those same **7 names**. The phantom drop of exactly 7 is migration
> 478 retiring the seven `site_specs.blog.*` fields — the arithmetic closes. The
> query figure never moved: **§8 counted names, I counted fields.** Two units, one
> population. This is the "your measurement answers the question you ENCODED" trap
> and it very nearly became a finding that the gate was leaking.

### Why the census is calibrated to production and not to the guard's opinion of it

The risk here is reporting fields as dead that actually resolve — `fixing a checker to
agree with a broken site` in reverse. Checked at the resolver, not assumed:
`plan_sections_action.go:623 resolve()` returns `(nil, false)` for a source with no dot
(so bare `config` IS dropped, on all 32 `info-card-grid` instances), and `site_specs`
falls through to `resolveSpecAlias`, whose step 1 needs `identityContainerAspects[aspect]`
to be populated and whose step 2 is `if aspect != "identity" { return nil, false }`.
**No phantom aspect can be rescued by the alias.** So every one of the 69 takes
`on_missing` → `skip_field` → key omitted → `{{if}}` swallows the markup. Same mechanism,
same silence, as the six orphaned articles.

### Why the birth gate can never close this door

A component is routinely inserted or altered by a hand-written migration or by hand SQL,
which never passes through `store_generated_component_action` at all. That is already a
recorded LANDMINE (85dbf889d). An at-rest sweep is the only shape that sees every write
path, which makes it the *robust* half rather than merely the *remaining* half.

### Ownership checked before starting

`scripts/who-owns.py 309` names the `meta_description_never_backfilled` (320) and
`bugfix_284_flag_only_items_promoted` lanes; both closed themselves on 2026-08-19. Live
transcript sweep of all sessions active today: three cite 309 (the 238 lane, the 337
lane, a remortgagecalculator CSS lane) and every hit is prior-art citation — none edits
`component_source_guard.go` or anything in this lane. **No competing owner.**

## 2026-08-22 (afternoon) — the at-rest audit BUILT, and everything that had to be proven rather than asserted

### What shipped

`config-key-audit --component-source-vocabulary` + daily CronJob at 07:20 UTC.
`747e717a1` (mode, rule refactor, frozen baseline, tests), `effd08fff` (image, manifests,
makefile), `62f187442` (`bugs_open/362` + a correction), register `CLC-025`.

**The rule is CALLED, not mirrored.** `component_source_guard.go` now returns
`[]SourceIssue{Field,Source,Class,Message}` from `SourceVocabularyFindings`, and
`SourceVocabularyIssues` is a thin projection over it. The birth gate's pre-existing
tests pass **untouched**, which is the evidence its behaviour is unchanged — not a
claim about the refactor, an observation about the tests that were already there.

### The controls, run against the REAL live library

| control | expected | got |
|---|---|---|
| day-one green over live state | exit 0, 69 grandfathered | **exit 0**, 69 = 51/14/4, 17 components, 6 live, 46 instances |
| the **real** pre-478 `blog-listing_pre_037` schema re-enters the library | RED | **exit 1**, phantom_aspect ×7 |
| a SECOND bad field on already-baselined `info-card-grid` (32 live instances) | RED | **exit 1** |
| a baselined field changes to a DIFFERENT dead source | RED twice (new + stale) | **exit 1**, both |
| a baselined component is repaired/removed | RED (stale) | **exit 1** |
| a DORMANT baselined component gains a live instance | RED | **exit 1** |
| zero aspects (the FLOOD direction) | exit 2 | **exit 2** |
| zero components (the SILENCE direction) | exit 2 | **exit 2** |

> **The demand control is the real pre-478 schema, not a synthetic one** — pulled from
> `content_components_bak_20260818_309_blog_listing`, the backup migration 478 took of
> itself. A synthetic phantom schema would have proven the regex works; this proves the
> check catches **the thing this bug is about**.

> ⚠ **And the demand control alone does NOT prove the baseline key is narrow enough**,
> which I nearly recorded as if it did. `blog-listing_pre_037` has no baseline entry at
> all (478 repaired it), so a component-keyed baseline would have gone red too. The test
> that actually discriminates is the **second bad field on `info-card-grid`** — a
> component that IS baselined, on 32 live instances. Under the 4-tuple key it is red;
> under a component-keyed one it is silently grandfathered. **The control that proves a
> detector fires is not automatically the control that proves it fires for the right
> reason.**

### Six mutation proofs — every one KILLED its test

Each applied to the working tree, test run, file restored byte-identical
(`diff -q` confirmed). `make build-*` builds from `git archive HEAD`, so a working-tree
mutation structurally cannot reach an image — which is what makes this safe on a shared
tree.

| mutation | test that must die | result |
|---|---|---|
| `baselineKey` returns `componentID` only | `TestBaselineKeyIsNarrow` | ✔ killed |
| the closure-date refusal removed | `TestBaselineIsClosed` | ✔ killed |
| `WokeUp` hard-coded false | `TestDormantWakingIsRed` | ✔ killed |
| `staleBaselineEntries` returns nil | `TestRepairedEntryIsStale` | ✔ killed |
| audit drops a class the birth gate reports | `TestAuditRunsTheBirthGatesOwnRule` | ✔ killed |
| a baseline route points at a non-existent file | `TestRepoBaselineMatchesItsRecordedCensus` | ✔ killed |

### The council round, and the part of it that was my fault

**Round 1: REVISE**, gated by `editquality`. **The gating objection was right about the
SUBMISSION and wrong about the WORK**, and the fault is mine, not the seat's: squeezing
the edit list to the ≤8 cap left one entry named `base/cronjob.yaml` carrying makefile
and docs content. Four objections across two seats then reported `base/kustomization.yaml`,
the production overlay and the `RELEASE_IMAGES` entry as **MISSING** — all three shipped
in the same commit. **A reviewer can only review what the submission says.** Logged in
`WRONG_CALLS.md`, and it is the second time this exact failure is on that file.

Objections that were owed real work, and got it:

- **debug_historian — no deploy-verification step.** Owed, and now in the RUNBOOK: image
  before manifest; the cronjob's image by jsonpath; a binary probe with a
  **must-be-ABSENT control** alongside the must-be-present one; a manual Job read at the
  **pod's** `terminated.exitCode`; then the `doc_notes` row as the positive control on
  the report path — `writeDocNote` is best-effort, so a silently refused insert otherwise
  looks exactly like a healthy quiet run.
- **guardian — is `doc_notes.subject_type` CHECK-constrained?** It is:
  `{tool, pipeline, experience, action, experience-pattern, landmine, component, decision}`.
  `writeDocNote` sends `'pipeline'`, which is in the set (`[MEASURED]` 1,895 rows). Checked
  rather than assumed, because the write is best-effort and a violation would be swallowed.
- **guardian — enumerate the call sites rather than asserting the refactor is safe.**
  Done: exactly **one** production caller, `store_generated_component_action.go:413` and
  `:430`. The seat guessed `create_tool_component_action.go` might be another; it contains
  **zero** references. Its instinct was right and its specimen was wrong, which is
  precisely why "enumerate, don't assert" is the correct ask.
- **reuse_agent — why not reuse `optional_key_budget_acks.json`'s mechanism?** Fair, and
  answered on **lifecycle** rather than format: the existing ack files are
  acknowledgement registers keyed by one subject with a scalar level, and they **grow** as
  a human acknowledges each new subject. This is a frozen census, machine-generated, keyed
  per finding, and it may only **shrink**. One abstraction would force one lifecycle on
  both and make this file appendable — the single property it exists to deny.
- **bug_historian — a detector whose only output is a `doc_notes` row is the
  `bugs_open/083` shape.** The sharpest objection of the round. Answer: the row is the
  DETAIL, the **failed Job is the alarm** — a red exits 1. And filing into work items
  would BE the 083 shape, since detection→dispatch is the broken half.

**Round 2 resubmitted** under the same trail correlation `a092d7d8` with an accurate edit
list. `RESUBMIT_CORR` keeps the trail in one place.

### Two missteps recorded at the point they were made

- **"147 components"** went into the PLAN having never been measured — it is **285
  active**. Caught re-reading my own prose for the one number with no query behind it.
  Dangerous precisely because every other figure in that paragraph *was* measured: an
  invented number inherits its neighbours' credibility, and no marker rule catches that.
- **"`STRUCTURAL_KEY_CARRY_MISS` is one of 358's unread codes"** — false when written.
  `cmd/content-loss-check` consumed it as of `cba51ad1d` *that morning*; 8 of its 28 rows
  are now `resolved`. Caught by the 358 lane, **re-verified here at the source and the
  table before accepting it**. The decision it justified (don't route findings into
  `agent_error_log`) survives on a better reason the correction does not touch: that
  writer only fires when a page is **BUILT**, so the eleven dormant components are outside
  its reach permanently, however faithfully its rows are read.

### One shared-tree incident, and what it teaches

The `358` session committed `cmd/config-key-audit/main.go` by pathspec while my dispatch
arm was in the working tree and its symbol was still untracked — so HEAD carried
`undefined: emitComponentSourceVocabulary` and **would not build fleet-wide**. They
removed the arm (`8664a7f96`) rather than committing my files under their message, which
was the right call twice over. Restored here in the commit that supplies the symbol.

**The durable lesson is the check, not the incident:** a green working-tree build says
nothing about HEAD when an untracked file supplies a symbol. Every commit in this phase
was verified by extracting `git archive HEAD` into a scratch tree and building **there**.
That also caught a second instance of the same class one layer down — my
`TestRepoBaselineMatchesItsRecordedCensus` reads the baseline JSON from a repo-relative
path, so the JSON had to be in the same commit or HEAD's tests would fail while my tree
stayed green.

## 2026-08-22 (late) — council ROUND 2, one real defect found, and the decision to stop at two rounds

**Verdict: REVISE again**, gated by `editquality`. **8 of 12 seats approve**, and three
moved **object → approve** on this round: `reuse_agent`, `debug_historian`,
`tooling_provenance`. The round earned its cost, and here is the accounting.

### The one objection that found a REAL defect — and it was a good one

`debug_historian`, edit 3: *"The 69-entry baseline must be a mechanically-derived count
from actually running the shipped `SourceVocabularyFindings` against the live DB, not
reconstructed from separate ad-hoc reasoning."*

**Correct, and it caught a second predicate I had let in through the back door.** The
baseline was generated by a **psql query that re-stated the classification** — inside a
change whose entire argument is that the audit must CALL the rule rather than mirror it.
The SQL and the Go agreed (the audit graded all 69 grandfathered, zero stale, which is
what proves it) — but **"happened to agree" is not the property; "cannot disagree" is.**

Fixed: `--emit-baseline` writes the file from `SourceVocabularyFindings` itself. Run
against live state and compared tuple-for-tuple with the committed SQL-derived file:
**69 vs 69, IDENTICAL.** So the shipped file was right and is now *regenerable* rather
than merely lucky.

### A mutation proof that SURVIVED, which is the most useful thing in this session

Corrupting the emitter's date stamp **did not kill** the new round-trip test. Because the
test **built its own entries** with the same field assignments instead of calling the
emitter — so the mutation changed code the test never touched.

> **A test that duplicates the logic it is meant to pin cannot fail for the right reason.**
> This is the estate's own *"a mock's own bookkeeping cannot assert a NEGATIVE"* landmine,
> met head on, by a session that had already cited it — and missed on the first attempt.

Split `componentSourceBaselineEntriesFrom` out of the I/O so the test exercises the
decision; the date mutation then died. **Then the same mistake one level down:** the
tightened test asserted the route **constant** was non-empty rather than what the emitter
**stamped**, so dropping the route stamp still survived. Now it asserts on the emitted
entries — route non-empty, route **resolves to a real file**, date equals the closure
date. Both mutations killed.

**The generalisable shape: when a mutation survives, suspect the TEST before the code.**
Two rounds of this in ten minutes, both the same error wearing different clothes.

### Why I am stopping at two rounds rather than chasing a third

The gate is **advisory** and CLAUDE.md says one run per coherent task, not per iteration.
Of what remains:

- **`editquality`'s two format objections are mine and are real** — I claimed
  `base/kustomization.yaml`, the overlay and the RUNBOOK in *rationale prose* rather than
  listing them as edits. Same shape as round 1. With 11+ real files and a **≤8-edit cap**,
  a third round could only satisfy this by dropping genuine edits to list plumbing. That is
  chasing the form of the submission, not improving the change.
- **`editquality`'s third objection asks for something the owner has ruled against.** It
  objects that the plan *"describes work already applied outside this review, not a set of
  edits awaiting review"*. But the owner ruling of **2026-07-29 §2** is explicit: *"review
  here is after the fact, by design … Do not claim an ordering constraint you do not have;
  do not pretend you could have waited"* — and `Council-Submitted:` exists precisely so a
  commit can precede its verdict. **A seat gating on the workflow the owner mandated is a
  defect in the gate, not in the submission, and resubmitting to satisfy it would be
  complying with the opposite of the standing rule.** Recorded here rather than acted on;
  it belongs to whoever owns the gate roster, not to this lane.
- **`guardian`'s and `prior_art_librarian`'s remaining objections ask for evidence that now
  exists** — call sites enumerated (one production caller; the seat's guessed caller has
  zero references), `main.go` verified against `git archive HEAD`, mutation proofs run
  rather than promised.
- **`bug_historian`'s 083-shape objection is answered on the merits** and I still think the
  answer is right: the `doc_notes` row is the detail, the **failed Job is the alarm**, and
  filing into work items would BE the 083 shape.

### Coverage report, and a filter artefact of my own

`098` lists `747e717a1` as `[submitted: a092d7d8…]`, **MISMATCH: 0**. The trailer works.

> ⚠ **I nearly recorded the opposite.** My first run piped the report through a
> `grep … | head -20`, and my commits were below the cut — so I read "not in the report"
> off an absence **my own filter had created**, and said so out loud before checking. The
> other two commits are correctly absent for a real reason (build/, deployments/, cmd/ and
> docs are outside council scope — only `747e717a1` touches `platform/`). Third instance
> today of the same class: **an absence is only evidence if you know what your filter could
> have shown.**

## 2026-08-22 (after the fleet build) — LIVE and PROVEN, and the defect the verification recipe caught

### Proven in-cluster

`[MEASURED 2026-08-22 17:55Z]` image `v1.0.1326`, CronJob applied (`configured`, not
`unchanged`), manual Job run, **pod `state.terminated.exitCode` = 0**, `doc_notes` row
written at 17:55:14Z carrying **69** fields / 51/14/4 / **17** components / **6** live on
**46** instances. The row is the positive control on the report path: `writeDocNote` is
best-effort, so a pod exiting 0 having written nothing reads exactly like a clean estate.

### The check shipped in a state where it could not run

Everything said it was fine — release built it, overlay applied, `get cronjob` listed it,
and the **binary probe found the mode compiled in with its must-be-absent control
passing**. Every scheduled run would nonetheless have exited 2, because the image is
`FROM alpine` + `COPY` of the binary alone while the mode loads its baseline from a
repo-relative path. The builder stage `COPY . .`'d the repo, so the path existed at build
time and nothing failed during the build.

> **A binary probe answers "did my CODE ship". It says nothing about "can my code RUN."**
> The moment a mode needs *data*, the deployment has a second dependency that no image
> tag, provenance stamp or `/proc/1/exe` grep will mention.

**What caught it:** running the container with its real arguments and no environment and
reading the **first** refusal — it named the baseline, not `PG_CLIENTS_HOST`, so it had
not reached the DB check. Two seconds, and the only step in the chain that could have
failed. That step exists because the council's `debug_historian` seat asked for a
deploy-verification recipe; **the objection I nearly dismissed as procedure is the one
that caught a live defect.**

### Why the fix is a mount, not a `COPY`

The baseline is **designed to shrink** as `bugs_open/362`'s repairs land. Baked in, it
goes stale on the first repair and the check reports the repaired entry as STALE until
someone rebuilds an image for a *data* change. Mounted from a `configMapGenerator`, a
repair is `apply -k`. That forces the real file into the kustomize base with a symlink
from the docs path (kustomize refuses a generator source outside its root) — the
`component-fallback-check` arrangement, same constraint, and its kustomization.yaml says
so for the same reason.

**No rebuild was needed** — `--baseline` had been in the binary since `747e717a1`. Worth
stating because the instinct on finding a broken image is to bump a tag, and a fleet tag
is not this lane's to bump.

### Two small things worth not re-deriving

- **`kubectl kustomize` rewires the volume to the content-hashed ConfigMap name.** Checked
  rather than assumed: a volume left pointing at the generator's *base* name mounts
  nothing, and the pod would start clean and read an empty directory.
- **There are TWO `bugs_open/362`s** — mine (`…sixty_nine_grandfathered_phantom_source_fields`)
  and another lane's (`…two_tool_writers_persist_rendered_html_without_link_repair`). The
  new CLAUDE.md ruling on dated counts cites "`bugs_open/362` §6a", which is **the other
  one**. Exactly the ambiguity CLAUDE.md warns about; this lane's baseline routes by full
  slug path, so nothing of ours is ambiguous. **Resolve 362 by slug, never by number.**

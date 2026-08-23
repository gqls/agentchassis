# NOTES — bugfix_342_absent_required (append-only, newest at the bottom)

## 2026-08-22 — state verification before touching anything

- who-owns: "OWNED or recently active" — but the activity is the `bugfix_260_render_fallback`
  lane's own filing + fix work, finished 2026-08-21 evening (their NOTES: APPROVED round 6,
  live-verified on v1.0.1322). The bug file says UNOWNED; RFC_041 §5 says "what it needs now
  is a name". This lane is the name. Tree checked: nothing dirty touches the 342 surfaces.
- **The bug file's warning banner is STALE.** It says the editor-route escalation "is NOT in
  v1.0.1322 and is inert until the next roll". The next roll happened: fleet is on v1.0.1323
  (both replicas, pod images read). Ancestry: `cd90e8b27` (wire editor routes), `65f1b0b95`
  (single emitter), `af4743464` (the seam) are ALL ancestors of the v1.0.1323 stamp
  `70e7b4f9c` (found via another lane's commit `49d90b280`, then verified at the artefact:
  `grep -aq "70e7b4f9c" /proc/1/exe` present on BOTH replicas, nonsense control absent).
- `record_absent_required_fields`: **0 agent_definitions rows name it.** Chrome escalation is
  dormant. **7** live steps as of 2026-08-22 use `render_site_components` (nav-link-fixer, nav-updater,
  pageflow-builder, rerender-chrome, rerender-pages, rerender-site, site-work-orchestrator).
- `required_fields_missing` items: 45 complete / 30 needs_human_review, ALL from the
  post-deploy check (latest 08-18, all naming deployed `ported-page` rows). Zero render-time
  items yet — expected: editor escalation live only since the 1323 roll this morning, chrome
  half unarmed, and the write is per-edit.

### Misstep (mine, caught in-flight): a VACUOUS census read as a finding

I ran the chrome census (site_components rows missing required source:llm fields), got zero
rows, and stated "no chrome row is missing a required field" — then the vacuity check showed
**candidate_pairs_tested = 0**: no site_components row references ANY component declaring a
required llm field, so the query could never have returned a hit. The zero was true but
meant something different: "the chrome store uses only 0-required components", not "all
required fields are supplied". The census was re-read on the corrected join and the
conclusion for arming purposes survives (0 rows fire either way), but the first phrasing
would have been a WRONG_CALLS row if it had reached a doc. Cheap check, now in RUNBOOK:
**count the candidate pairs your census tested in the same query; a zero over zero
candidates is not a finding.** Logged in WRONG_CALLS.md per the tally rule.

### Demand-control failure worth recording

Pod-log grep for render activity (`RenderTemplate: fields rendered empty|Re-rendered
component|render_site_components|RenderComponentAction`) returned 0 on both replicas while
the DB showed 38 page_components writes since pod start — the greps don't match what these
paths actually log at INFO. So "0 absent-required reports in the logs" could not be given a
log-side demand control; the DB-side census is the evidence that matters and is what the
plan uses.

### Populations re-measured (the bug file's figures had moved)

- Active components: **283 as of 2026-08-22** (was 253 at filing, 2026-08-20). No-schema: **100 as of 2026-08-22** — **95 are tools** (schema-less
  by design; `isSelfContainedSection` codifies it). Non-tool no-schema: **5 as of 2026-08-22**, one
  page_components usage each; 2 with `{{.field}}` placeholders (`report-request-form`,
  `audience-check-form`). §5's "expect the 75-of-253 to be the hard part" has dissolved into
  small data work.
- Chrome components WITH required llm fields exist in the library (footer-with-disclaimer 17,
  header-with-categories 16, header-with-search 5, header-with-cart-or-nav 4,
  header-minimal-tool 2) but **no site_components row references any of them** — the store
  uses site-header/site-footer/head (0 required). Hence arming the record = 0 items today.
- page_components writer-side census: **131 rows as of 2026-08-22** missing a required llm field top-level;
  breakdown: ported-page/body 77 deployed + 23 removed (the post-deploy check's existing,
  already-itemised population), hero/headline 14 deployed, scattered singles after that.
  NOTE this is the WRITER question (pre-gate); the seam's render-time answer is a strict
  subset after `contextToInterfaceMap` defaults.

### Code reads that shaped the plan

- Editor routes: emit item, then **persist the blank anyway** — refusal is the missing half
  (`section_editor_actions.go` applyContentEdit ~:1042, applyComponentSwap ~:1169).
- `ApplySectionEditAction` already has the ONE-persist-switch idiom (link repair :~445,
  envelope refusal :~455) — the refusal gate belongs there, not in the two branches.
- `section-editor.apply_edit` has **no error_step** (live row) → a refusal meets
  `bugs_open/344`'s completion-trample on the DRIVING item. Live page still protected; the
  item status is 344's bug, not this lane's. Interaction stated in the code comment.
- Six unwired sites re-verified: GateConvertedTemplate + tool_birth_instance_scope (raw
  candidate templates, no component row), legacy head render (template load only),
  RenderTemplateWithMap (contact-info block, callers have no component schema),
  offline audit ×2 (probes with fields REMOVED by design). No change owed.
- `assemble_from_library` renders a stitched TEMPLATE (content arrives later) — report-only
  is correct there; no escalation owed.
- RFC_022 budget: `apply_section_edit` counts 7 optional keys; ConfigKeys are not counted by
  the audit; the new key changes nothing. `render_site_components` has NO registered input
  spec at all (pre-existing; not widened by this lane).

## 2026-08-22 late morning — built, tested, submitted, committed

- Refusal half implemented: key + deciding arm in `mistyped_llm_fields_gate.go`
  (`refuse_absent_required_fields`, `refusePersistForAbsentRequired`), outcome field +
  branch copies + ONE gate at the persist switch in `section_editor_actions.go`, ConfigKeys
  declared. Tests: two table tests + a seam→outcome→gate chain test; **mutation-proven**
  (inverting the deciding arm fails TestEditorRefusalNeedsBothArmingAndAFinding AND
  TestSeamFindingSurvivesOntoTheEditOutcome; reverted, green). Honest limit written into the
  chain test's comment: the in-branch copy needs a DB — the post-roll canary covers it.
- Migrations: 550 (chrome record arm, appliable) + 551_HOLD (editor refusal arm, after the
  roll), both with rollback sidecars, DO/RAISE verifies, double-apply refusals; 550 also
  refuses if a render_site_components step has moved into a sub_workflow (the write only
  reaches top level).
- Council: corr `3626629a-f2bc-4089-9118-c1d6dd007807`, submitted 09:32Z, dispatched almost
  immediately (no queue wait this time). Two client-side schema rejects first: `create` is
  not a valid operation (use `add`), and `risks` must be a STRING not an array.
- Committed `0ee442cfb` with `Council-Submitted:` trailer, 13 files, scope report all mine.
  Clean `git archive HEAD` build verified after commit (platform/internal/cmd all compile).
- **My WRONG_CALLS entry (vacuous census) was swept into the 337 lane's commit `9e23fb852`
  as a same-file passenger** between my append and my commit — the exact CLAUDE.md case;
  stated in my commit message, nothing lost.
- MEMORY_workstreams: lane registered with the owed follow-ups (verdict read, 550 apply,
  post-roll 551 + canary, the 5 no-schema components decision).
- 550 deliberately NOT applied pre-verdict: migrations are council scope (314 — live the
  moment they apply), and the arm fires on 0 rows, so waiting costs nothing.

### The 9-of-15 arithmetic re-derived first-hand (not inherited from the bug file)

`grep -rn "RenderTemplate(" --include=*.go platform/ internal/ cmd/ | grep -v _test | grep -v
"func RenderTemplate"` → **14**, plus `RenderTemplateWithMap` = **15**. Schema-wiring sites,
counted with a grep that NAMES THE RECEIVER (the bug file's own landmine: a bare
`grep -c 'InputSchema = '` also matches `ci.InputSchema`, a different struct) → **9**:
assemble_from_library:302, v3_site_actions:2464, section_editor:1069 + :1206,
render_site_components:1051, component_library:1730/1805/2092, rerender_page_sections:655.
**9 + 6 = 15 closes.**

The six unwired, each read at its own call site rather than taken on trust:
- `GateConvertedTemplate(function, converted string, …)` — signature takes a raw string; no
  component row is in scope, so there is no schema to pass.
- `ScopeToolBirthTemplate(html, function string, …)` — same shape, raw candidate template.
- the legacy head render (`rerenderSinglePage` :538) — its loader
  `rerenderLoadHeadTemplate` selects `defaults->>'head'` then `html_template` only, never
  `input_schema`; and **`RerenderSitePagesAction` is in no `GlobalActionRegistry` entry**
  (grep: 0 hits), matching RFC_041 §4 — a dead path, so wiring it would be inert anyway.
- `RenderTemplateWithMap` — a different executor (contact-info block), callers hold no schema.
- the two `cmd/component-render-check` probes — they render with fields REMOVED on purpose, so
  a report there fires on every probe by design.

## 2026-08-22 midday — council round 1: REVISE, and the objection that changed the design

**Gating (`prior_art_librarian`, HIGH):** 551 UPDATEs by `type='section-editor'` while a landmine
documents four agent types carrying TWO active rows, of which only the higher version loads — so
an UPDATE-by-type can silently arm the row nobody reads. **The seat was right that I asserted it
rather than checking it.** Checked: section-editor has exactly one live row (`2ed3b581`, v1); the
four are `chief-strategist`, `content-creator`, `content-creator-contact`,
`site-component-architect`. My precondition was already the correct guard — it now NAMES the four
and says why aborting (not writing both) is the right response.

**The one that changed the code (`bug_historian`, medium):** arming DETECTION on chrome while only
the section editor got PROTECTION reproduces **this bug's own shape** on the sibling call site —
016b §9's *"one call site of a shared judgement gets the rigorous fix, the sibling stays
heuristic"* — and my "0 rows fire today" is a population snapshot, not a structural guarantee.
Acted on rather than argued: the chrome store now has the SAME refusal, through the SAME decision
function, **default OFF with no migration arming it**. That is the honest middle: arming today
would arm an unexercisable refusal, while leaving the capability out would put a code change, a
review and a roll between the first adopting site and its protection. The flip trigger is named —
the first `capability_gap` item whose `spec->>'finding_type'` is `required_fields_missing` (⚠ CORRECTED 2026-08-23 — this trigger used to name a `required_fields_missing` item with `surface='site_component'`, which the capability_gap rework means will now NEVER be filed, so the old trigger could not fire). A new test asserts the
two paths agree on all four arming×finding combinations, so a forked chrome implementation fails.

**`guidelines`, medium:** 550's post-verify walked only the top level — i.e. it certified exactly
the coverage the write can reach and called it full. It now walks `config.sub_workflow.steps` too
and RAISEs on any unarmed nested occurrence. **This is the sharpest kind of objection: the check
that would have passed was the one I wrote to prove the write was complete.**

**`guardian`, medium:** name the owning pipelines, don't say "chrome". Enumerated in the header.
Its two low objections taken: both migrations relabelled `config_change`.

Everything else was answered with a query and attached rather than re-asserted: one consumer of
`apply_section_edit` (top-level 1, nested 0), both keys armed 0 **re-measured this round**,
`render_site_components` **7** top-level / **0** nested as of 2026-08-22, optional-key budget **7** of 10 as of 2026-08-22 (ConfigKeys are not
counted).

**What I got right and want to keep doing:** I re-read the submission against the actual diff
before re-firing (`git diff` + a grep per claimed symbol), which is the 260 lane's own hard-won
rule — three of their six rounds were the submission describing code that had moved.

## 2026-08-22 afternoon — APPROVED at round 2, and the advisories acted on rather than banked

Verdict: **approved with 3 advisory objections, none high** (trail `3626629a`, round 2). An
approval is not a reason to stop reading the objections — one of them named a real property of
my own change that I had not stated.

**`guardian`, medium — where does the new gate sit relative to the `!force` idempotence exit?**
The honest answer is *after it, so it never runs when the exit fires*, and I had not said so.
Checked at the code (`:850`): the exit returns when the slot already holds non-empty,
non-`pending` HTML. **For the REFUSAL that is correct and not a gap** — the exit fires precisely
when nothing is about to be written, and a gate that prevents writes has nothing to prevent.
**For DETECTION it is a real blind spot and a pre-existing one**: an already-populated slot whose
`content_data` lacks a required field is never inspected by a non-forced refresh, because no
render happens to inspect it. Now written into the code at the gate, so the next reader does not
have to rediscover it: *what is checked is every slot this function is about to write*, not every
chrome slot continuously.

**`bug_historian`, medium — is `chromeSlotHasStoredHTML` a fresh ad-hoc "does this hold real
content?" discriminator, the class this estate has had to re-harden twice (016b §9 items 3, 5)?**
No, on both counts, and the objection was worth answering rather than waving off. It is
**pre-existing** (`:1706`), and it is already the discriminator the execution-failure branch
~90 lines above uses for exactly this fatal-vs-degraded split (and the caller at `:333`) — so
this is reuse, not a second lineage. And it is not that class of judgement: it asks whether any
non-empty `rendered_html` is STORED, never whether the HTML looks meaningful, so there is no
sparse-but-real content for it to misread.

**`prior_art_librarian`, medium — the chrome zero was quoted, not re-run, while being the
load-bearing argument for shipping the refusal unarmed.** Fair. Re-run at approval, one statement
with its own vacuity guard: `candidate_pairs 0 | rows_missing 0 |
components_with_required_fields 813 | chrome_rows_total 72`. **Read in that order it says
something the bare zero does not**: neither side is empty — 813 component×field pairs in the
library do declare a required field, and 72 chrome rows exist — what is empty is the JOIN. The
figure is now in 550's header in that form.

**`editquality` + `guardian`, low — enumerate every caller of the signature you changed.** Done:
one production caller (`:312`) and three test call sites (`site_component_lock_guard_test.go`
:292/:327/:353, all updated). "It compiles" was already evidence, but the enumeration is the
thing that could have come out otherwise.

**`debug_historian`, low — bake the row-count assertion into 551's UPDATE.** Done, with
`GET DIAGNOSTICS`. The point is sharper than it first reads: a precondition that ran *beforehand*
cannot see a second active row inserted between it and the write, and on this tree that is not
hypothetical.

**`bug_historian`, low — this is still per-call-site patching of a generic root cause
(`missingkey=zero`), and the historian's index says narrow per-site patches on this class have
not historically been the fix that stuck.** True, and deliberately so: refusing at the SEAM is
new authority over content that renders successfully today at sites that never asked (owner
2026-08-02 §2). Recorded here as the standing argument rather than settled — if a future owner
ruling licenses a seam-level default, this is the note that says why we did not take it.

**Applied `550` after the verdict**: 7/7 steps armed, verified by reading the live rows rather
than trusting the migration's own post-check, with a negative control (the REFUSE key must still
be armed nowhere — it is). Ledgered via `--record-only`.

## 2026-08-22 evening — v1.0.1326 rolled, 551 APPLIED, and the canary PASSED on both arms

### The build check, and one probe arm that was invalid

Fleet on **v1.0.1326** (both replicas, started 15:10Z). Probed at the binary, both replicas:

| literal | novel? | want | got |
|---|---|---|---|
| `refuse_absent_required_fields` (editor key) | **yes**, 0 at `0ee442cfb^` | present | **present** |
| `REQUIRED content field(s) absent — refusing to store` (chrome) | **yes**, 0 at `a39a16555^` | present | **present** |
| ~~`refusing to persist`~~ | **NO — 3 pre-existing files** | — | **INVALID ARM**, see WRONG_CALLS |
| nonsense literal | — | absent | absent |
| `Go template execution failed, using regex fallback` (deleted) | — | absent | absent |

The startup provenance line had scrolled (`--tail=3000`, absent), and probing the binary for my
own commit shas returns OUT for all three — **a binary carries the ONE sha it was built from, not
its ancestors**, so that is uninformative by construction, not evidence of absence. The two novel
literals are the evidence, and they are capability-level, which is the stronger form.

### 551 applied, and verified independently of its own post-check

Preconditions re-checked live first (8 hours had passed): 1 live section-editor row, `apply_edit`
still `apply_section_edit`, key absent, and this morning's chrome arm still 7/7. Applied clean.
Read back from the live row: key `true`, **both sibling keys survived** (`strip_literal_markdown`,
`allow_rendered_html_transform` — a wrong `jsonb_set` path replaces the whole config branch, which
is what that check exists to catch), version 1→2. Two negative controls: no other agent gained the
key (0), and the **chrome refusal is still armed nowhere** (0) — its deliberate state.
`_HOLD` suffix dropped in the same commit; ledgered via `--record-only`.

⚠ Another session held `.git/index.lock` during the rename. **Waited, did not remove it** — that
file is another session's commit in progress.

### The canary — BOTH ARMS, at the artefact

Targets deliberately NOT `deployed`, so a failing canary could harm no live page.

**Refusal arm** (`0a1498b3`, tool-cta, 3 required fields absent), corr `63f2eab4`:
- step **FAILED** with the exact text: *"refusing to persist — 2 schema-required field(s)
  rendered empty (headline, trust_note); the live section is left unchanged and a
  required_fields_missing item has been filed"*;
- **artefact byte-identical**: md5 still `69a2f28c…`, and `updated_at` still **2026-07-17** —
  the row was not touched at all;
- **item filed**: `required_fields_missing`, status `detected`, naming both fields,
  `surface=page_component`, `route=content_edit`. Refusing did not cost the record.

**Positive control** (`9737d0d9`, use-cases-list, required field present), corr `3eb5318a`:
all orchestrations **COMPLETED** through `deploy_page`. The artefact's bytes are identical
(expected — I set the field to its own value), but **`updated_at` moved to 18:05:04**, which is
the discriminator that separates *"persisted"* from *"skipped"*. **An arm that only stops edits
would have failed here.**

### A finding the canary produced that no test could: the subset behaviour, observed

The census predicted three absent fields (`headline`, `description`, `trust_note`); the refusal
named **two**. `description` was filled by a default in the merged map the template actually sees.
That is the documented seam-vs-gate divergence — *"did the WRITER supply it?"* vs *"will it RENDER
EMPTY?"* — **confirmed in production for the first time**, having been pinned only by
`TestSeamReportsASubsetOfThePreRenderGateAndSaysWhy` until now. The seam's report really is a
strict subset. Do not "fix" the count.

### HONEST LIMIT: check (c) of the canary did NOT fire

551's header says the canary reads the DRIVING work item's terminal status, expecting `complete`
until `bugs_open/344` lands. **It could not: a CLI-dispatched canary has no driving work item.**
Measured — 0 trampled rows in the window, and the only `required_fields_missing` item is the one
the refusal filed. So the 344 interaction remains **[UNEXERCISED]**, not "verified benign", and
closing that needs a work-item-driven edit (the `section-editor` route reached from a queue item),
not another CLI run.

### Two measurement missteps of mine, both in WRONG_CALLS

1. The invalid probe arm above — a positive control on a literal that already existed.
2. My canary monitor waited for `count(*) = 2` across BOTH correlations, and the refusal arm
   alone writes two rows (child + parent), so it fired **BOTH ARMS TERMINAL** while the control
   was still running. A count aggregated over two populations can be satisfied by one; the fix is
   `count(DISTINCT correlation_id)` or asserting each separately.

## 2026-08-22 evening — residual #2 CLOSED: the five no-schema components owe NOTHING

The bug file's §5 told a fixing lane to *"expect the 75-of-253 no-schema components to be the
hard part, not the seam"*. That claim has now shrunk three times, and the third time takes it
to zero:

1. **75 of 253 → 100 of 283** (as of 2026-08-22) — the figure simply moved.
2. **100 → 5** — **95 are `component_level='tool'`**, schema-less BY DESIGN, which the rerender
   gate's own `isSelfContainedSection` already codifies. (Now a LANDMINES entry; the verifier
   returned STILL_VALID.)
3. **5 → 0** — measured this evening, per component rather than in aggregate:

| component | placeholders? | verdict |
|---|---|---|
| `lobby-grid` (vonc.com) | **none** | `missingkey=zero` cannot bite: no field to render empty |
| `provocation-card` (vonc.com) | **none** | same |
| `gauntlet-round-record` (vonc.com) | **none** | same |
| `report-request-form` (idea.uk) | heading, subtitle, button_text, footnote | **all four author-gated** |
| `audience-check-form` (idea.uk) | heading, subtitle, button_text, footnote | **all four author-gated** |

The two templated ones looked like the worst case in the estate — both live instances supply
**none** of the four fields, and one of them is a submit BUTTON's label. They are safe anyway,
because every reference is wrapped `{{if .field}}{{.field}}{{else}}Run the free check{{end}}`.
That is precisely the shape `missingBareFields` is scope-aware about — an author-gated field's
empty case is deliberately handled, so reporting it would be the false-positive noise the
council flagged. **Confirmed at the artefact, not just in the template**: neither rendered row
contains an empty `<button>` or empty `<h1>/<h2>` tag.

**So no schema is owed anywhere in this class, and "author minimal schemas" would have been
work that made nothing safer.** The honest general lesson, which is why this is written up rather
than just deleted from the list: **"declares no contract" and "has no contract to declare" are
different things, and the query cannot tell them apart.** Three of these have no fields; two
handle their own absence. An aggregate count of empty `input_schema` sees five problems where
there are none — the same misreading as the tool case one level down.

### Check (c) will resolve ITSELF — measured, so it is a wait, not a gap

The one canary check that could not fire needed a queue-driven edit rather than a CLI one. Rather
than fabricate a work item to exercise another lane's bug, I measured how `apply_section_edit` is
actually driven — and queue-driven is the NORM:

- **15 `section-editor` orchestrations in the week of 2026-08-17, 12 of them carrying a
  `work_item_id`** (80%).
- Live queue: **4 open items** handled by `section-editor` as of 2026-08-22; all-time
  **74 complete + 4 failed + 3 in_progress + 2 cancelled `section_edit`**, plus 8 `literal_markdown`.

So the refusal will meet a real queue-driven edit in ordinary traffic, probably within days. **What
to watch, and it is one query** — a refused edit whose driving item nevertheless reads terminal:

```sql
SELECT wi.id, wi.item_type, wi.status, wi.completed_at, wi.retry_after,
       wi.retry_after > wi.completed_at AS trampled_344, left(wi.error, 200)
  FROM site_work_items wi
 WHERE wi.handler_agent = 'section-editor'
   AND (wi.error ILIKE '%refusing to persist%' OR wi.updated_at > '2026-08-22')
 ORDER BY wi.updated_at DESC LIMIT 10;
```

A row with `status='complete'` AND an error naming the refusal is `bugs_open/344` confirmed on
this route; a row parked `failed`/`needs_human_review` means 344 does not reach here and check (c)
passes outright. **Either result closes it — which is what makes it worth waiting for rather than
staging.** ⚠ Do NOT file a synthetic work item to force this: it would put fabricated work on a
live queue to test a defect another lane already owns and has measured.

## 2026-08-23 — the escalation was filing items its own router could not route (2 of 2), and the answer to "can we close it?" is NO

Came back to run the one outstanding check and found something worse than an unexercised check.

### What I found, and it was not what I was looking for

The check-(c) query returned seven queue-driven section edits since arming, **all `complete`, none
refused** — a good demand control (the arm is not breaking healthy traffic) but no verdict. What
it also surfaced was **one `required_fields_missing` item at status `failed`**. Pulling it apart:

- item `a31da7f3` (my canary's), handler `required-fields-missing-handler`, **attempt_count 3**,
  terminal `failed`, `error` column NULL;
- its three servicing orchestrations all read `current_step=done, status=COMPLETED`;
- the handler's own triage recorded **`route: "malformed"`**, `component_id: ""`, `html_len: 0`,
  `page_type: ""`.

And then the part that matters: **there are TWO such items, not one.** `a6e00dcf` was filed at
**13:32 on 2026-08-22 by real production traffic** on `loans-application-tracker`
(loanandmortgagecalculator.co.uk), hours before my canary, and failed identically. **So the
producer's items were unroutable 2 times out of 2 — a 100% failure rate — and the first was not
a test.**

### The cause: two false claims, both mine, both in four places

`required-fields-missing-handler`'s `classify` step resolves the page by `spec->>'page_name'` and
the component by `spec->>'slot_name'` — I read the predicate out of the live agent row. **The
producer supplied neither.** And `check_required_fields_missing.go:180` keys on
`<page_id>:<slot_name>` while this producer keyed on `<site_id>:<component function>` — so
**"the item_key matches the check's so the two producers co-dedup" was also false**, and the two
would have filed two items for one defect.

Both are one assumption: **that reusing the item TYPE meant inheriting its router.** An
`item_type` is a string; a router is the fields it reads and the key it dedupes on. The
`reuse_agent` seat approved the reuse reasoning and was right to — the reasoning was sound. What
nobody checked, me least of all, was whether the artefact matched it. **And the failure is
invisible from the producing side**: the insert succeeds, the emitter logs `item filed`, every
orchestration reads `COMPLETED`. Only the item's own terminal status says otherwise, and only if
you look a day later.

### Fixed at source (`eb918bd58`), and the chrome case forced a real decision

The emitter now carries a `pageContext`, writes `page_name`/`slot_name` into the spec, takes the
check's key shape exactly when a page is known, and sets `page_id` on the row — a field that
existed on the shared `workItem` struct and was simply never set.

**Chrome has no page**, so the page-resolving router cannot classify it by construction. Three
options, and the middle one is what shipped: hand it over anyway (buys three failed attempts —
the defect I was fixing), invent a chrome handler (`bugs_closed/291`: an unregistered handler is
born `blocked` and never claimed, which reads as a queue bug), or **file it for a human at
`needs_human_review` — the router's own `park_*` vocabulary, so no new status is invented.**

The decision is factored into `requiredFieldsMissingRouting` so the test runs what production
runs. Mutation-proven: routing chrome to the page-router trips two assertions, and a
page-*without*-slot taking the routed path trips a third — the classifier needs both, so half the
context must not be allowed to look like all of it.

### A shared-tree note worth keeping

Three `UpdateWorkItemStatus` tests fail in the working tree and are **not mine** — another session
has uncommitted failure-ladder work in `load_work_item_actions.go`. Proved rather than assumed by
copying my four changed files onto a clean `git archive HEAD` and running the whole suite there:
**ok**. On this tree, "the suite is red" is not evidence about your own change until you isolate it.

### Also today

`/tmp` (a 16G tmpfs shared with every session) was **93% full** and Go could not write its build
output. **Did not clear it** — other sessions' work lives there. Pointed `TMPDIR` at this
session's scratchpad on the 139G root volume instead, which is the sanctioned place. If you hit
`no space left on device` on a build, that is the fix, not `rm`.

### So: can 342 close? NO — and this is why the question was worth asking

The refusal half is live, armed and proven. But the ESCALATION half — reviewed and approved on an
earlier trail — has been filing unactionable items since it shipped, and the fix for that is
**committed and INERT until the next roll**. A bug whose fix is committed but not live stays OPEN
by this estate's own bar. Submitted as council trail `a0ef0b07`.

## 2026-08-23 later — the "fresh build" does NOT carry the routability fix, and the lane is now purely waiting

Asked to carry on after a fresh build. **It is not fresh enough**, and the check that settles that
is worth writing down because the tag says otherwise.

| signal | reading |
|---|---|
| pod image, both replicas | **v1.0.1328**, started **11:51Z** |
| `deploy/agent-chassis` desired image | **v1.0.1328** (rollout reports "successfully rolled out") |
| makefile `IMAGE_TAG` | **v1.0.1329** — bumped, but nothing is running it |
| my routability commits | `eb918bd58` **13:08**, `23d2a577d` **13:34** — *after* the pods started |

So the arithmetic already says no. **Confirmed at the artefact anyway, with the control that makes
the negative mean something:** the new literal `capability_gap:required_fields_missing` is
**ABSENT** from `/proc/1/exe`, while yesterday's `refuse_absent_required_fields` (which IS live) is
**PRESENT** in the same probe. Without that second arm an ABSENT reading is indistinguishable from
a broken probe — which is the mistake I made in the other direction on 08-22 (a positive control on
a literal that already existed, `WRONG_CALLS.md`).

⚠ **Probe-literal novelty checked BEFORE probing this time**, per that same entry:
`capability_gap:required_fields_missing` → 0 occurrences at `23d2a577d^`, valid; `handler_remit` →
**1 occurrence, NOT a valid control**, and it was dropped rather than used.

### What production did in the meantime — three readings, all reassuring

- **Accrual is ZERO.** Still exactly **2** render-time items, most recent `2026-08-22 18:03`. The
  broken producer has filed nothing new in ~24h, so the cost of it staying inert another roll is
  measured, not guessed.
- **The armed refusal is not breaking healthy traffic.** **12 `section-editor` items completed**
  since arming (18:00 on 08-22), **0** carrying a refusal error. That is the demand control the
  348/`bug_historian` worry asks for — an arm that only stops edits would show up here as failures,
  and it does not.
- **Both armings survived the roll**: editor refusal `true`, chrome record 7/7. (Worth checking
  every time — a roll re-applies overlays, and config that "was armed" is a claim with a shelf
  life.)

### Canary check (c): still unexercised, and now with a measured reason

0 queue-driven refusals. 12 queue-driven edits completed cleanly in the window, so the traffic
exists — what has not occurred is an edit that *leaves a required field empty*. That is a rarer
event than "an edit", and the honest statement is that the sample has not contained one yet, not
that the interaction is benign.

### So: still cannot close, and the blocker is now ONE thing

Everything this lane owns is written, reviewed and approved. **The only blocker is a roll** — the
fix is committed and inert, and this estate's bar is *fixed AND live*. Nothing here needs a
decision or more work; it needs `make release` (owner-run, whole-fleet) and then the two checks in
the handoff's §4.

## 2026-08-23 ~17:00Z — v1.0.1330 CARRIES THE ROUTABILITY FIX, and the closure question narrows to one word: exercised

The roll happened for real this time. **v1.0.1330 on both replicas, started 16:03Z** — after the
13:34 commit — desired image and makefile agree, and the artefact confirms it:

| needle | novel? | want | got |
|---|---|---|---|
| `capability_gap:required_fields_missing` (new key prefix) | yes, 0 at `23d2a577d^` | present | **present, both replicas** |
| `a required_fields_missing router that can service` (builder_needed) | yes, 0 at parent | present | **present** |
| `refuse_absent_required_fields` (known-live) | — | present | **present** — *the probe works* |
| nonsense literal | — | absent | absent |

That third row is the arm that matters on a POSITIVE probe as much as a negative one: it proves
the probe is capable of reading the binary at all, so PRESENT is a reading rather than an artefact
of a working grep on the wrong thing.

**Both armings survived the roll**: editor refusal `true`, chrome record 7/7. Re-checked because a
roll re-applies overlays.

### The state that decides closure: LIVE but UNEXERCISED

- render-time items: **still exactly 2**, both the old `required_fields_missing`/`failed` rows,
  newest `2026-08-22 18:03` — i.e. **nothing has been filed since the fix went live**;
- `capability_gap` rows from this producer: **0**;
- queue-driven refusals (check c): **0**.

So the fix is live and has never run. A zero here means "no demand yet", not "it works" — the
distinction this estate keeps re-learning. **Firing the canary's refuse arm is the demand**, and
it is the same safe target (`0a1498b3`, `pending`, nothing live at risk), re-baselined first:
md5 `69a2f28c…`, `updated_at` still `2026-07-17`.

⚠ Note the old item does NOT block the new one: the key shape changed from
`<site_id>:<function>` to `<page_id>:<slot_name>`, so there is no dedup collision — which is
itself a small confirmation that the two shapes really were different.

### The canary rerun on v1.0.1330 — the item is now SHAPED correctly

Refuse arm, corr `ceff1346`. The refusal fired again with the same message (`headline,
trust_note`) and the artefact is **byte-identical, `updated_at` still 2026-07-17** — so the
protection is unchanged by the rework, which matters because the rework touched the emitter that
runs immediately before the gate.

And the item it filed, `562788c3`, is the whole point:

| field | before (a31da7f3) | now (562788c3) |
|---|---|---|
| `item_key` | `…:<site_id>:tool-cta` | **`…:f438eca6…:tool-cta`** — page-scoped, the check's shape |
| `page_id` | NULL | **set** |
| `spec.page_name` | absent | **`ai-agent-roi-estimator`** |
| `spec.slot_name` | absent | **`tool-cta`** |
| handler / status | handler / `detected` → `failed` ×3 | handler / `detected` (pending pickup) |

**Both fields the router resolves by are now present, and the key matches the sibling producer's**
— the two defects are fixed at the artefact, not just in the source. What is NOT yet proven is the
router's *disposition*: the item is `detected, attempts=0`, so it has not been picked up. That is
the last piece of closure evidence and it is a wait, not a task.

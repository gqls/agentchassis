# NOTES — bugs_open/404, a reason vocabulary whose readers disagree

Append-only, newest at the bottom. Technical log: evidence, commands, what the system actually
said, and every misstep.

---

## 2026-08-26 — lane opened; ownership checked, defect re-verified at both readers

### Why this bug, and what the sweep ruled out

Third bug of this session, after `bugs_open/359` and `bugs_open/407`. The ownership sweep ran
over 39 candidate numbers and found only **three** with no ACTIVE owning workstream — 338, 356
and 404. Recorded so the next sweep does not re-walk them:

| candidate | why not |
|---|---|
| `338` | voice-gate density rules on a single sentence — genuinely open and unowned; **a real candidate, left for next** |
| `356` | fixed in the tree and awaiting a roll; its remaining work is **17 separate routing gaps, each needing its own judgement** — a programme, not a bug fix |
| `404` | taken |

`scripts/who-owns.py 404` → no ACTIVE lane. Filed 2026-08-25 by the `bugs_open/384` lane, which
found it while ANSWERING a council objection — the `prior_art_librarian` seat asked why migration
615 hand-rolled a fan-out instead of reusing the shared per-page re-render creator, and reading
the shared creator to answer showed that reusing it would have shipped 40 assemble-only items and
no visible change. **The hand-rolled version was right for the wrong reason.**

### The defect, verified at BOTH readers today rather than taken from the file

The live gate, read from `agent_definitions` `[MEASURED 2026-08-26]`:

```
check_rerender_mode.condition:
  input_data.spec.reason == 'image_landed'
   OR ... == 'section_data_resolved'
   OR ... == 'cta_links_stale'
   OR ... == 'template_changed'
   OR ... == 'literal_markdown'
then_step: rerender_sections     else_step: render_page     <- assemble
```

The Go reader, `create_rerender_items_action.go:~219-235`, verbatim:

```go
scoped := (reason == "section_data_resolved" || reason == "image_landed") && componentIDStr != ""
stampReason := scoped || reason == "cta_links_stale"
keyReason := ""
if stampReason { keyReason = reason }     // otherwise the item is ASSEMBLE-ONLY
```

**Five values in the gate, three in Go.** `template_changed` and `literal_markdown` were added on
the SAME DAY — 2026-08-18, migrations 460 and 473, by different lanes — and neither touched Go.
So "the next vocabulary addition will repeat this" is not a prediction; it has already happened
twice in parallel within one day.

### The property that makes this dangerous rather than untidy

**Every reader that does not know a value fails toward `assemble`** — which re-ships the stored
HTML verbatim, completes green, and changes nothing. Checked at each reader: the gate's
`else_step` is `render_page`; the Go reader's unknown reason leaves `keyReason` empty so the item
carries no reason at all. A vocabulary whose readers failed toward *re-resolve* would announce
itself — you would get too many re-renders and notice. Failing toward assemble means the estate's
own preferred, safe, cheap mode is also its silent-failure mode.

### ⚠ THE EXPOSURE IS LATENT, AND THE BUG FILE PROVES IT RATHER THAN ASSUMING IT

The file carries **three dated self-corrections**, and the second and third are the instructive
part: the filer's own inference that 471 reason-bearing items were shipping silently was **wrong**,
and they went and checked. Every live producer stamps `spec.reason` in its own INSERT, so the gate
sees it and routes correctly; **not one item ever reached the stale Go reader.** Verified across
live AND archive: of 17,285 `page_rerender` items from that path, 203 carry a reason and all 203
are `section_data_resolved` — which Go knows.

**So this is a trap for the next author, not live damage.** Anyone planning this must quote the
471 as an URGENCY argument (the reason is heavily used via paths that bypass the shared action, so
a future author routing through it is likely) and never as a damage claim.

The third correction is worth reading twice on its own account: the filer's control —
*"6,428 items, 3 carry a reason"* — was a LIVE-WINDOW undercount, because closing a row archives
it out of `site_work_items`. The real figure is 203 of 17,285. And the `bugs_open/410` lane had
already relayed the wrong number into their own rationale, then re-ran the ORIGINAL query
independently, got it to the digit, and recorded that as first-hand confirmation — **they verified
the number by making the same population error, and the exact agreement made it more convincing.**
The lesson recorded there is the one to carry: *re-derive the POPULATION — which tables, which
window — not only the arithmetic over someone else's choice of table.*

### The machinery that already exists for this class

`platform/livespec` is `bugs_open/363`'s answer to exactly this problem: a Go guard that asserts a
property of a live DB object by reading the MIGRATION FILE cannot work, because a migration is
append-only history frozen by its checksum while the live object keeps moving. So livespec is the
**declaration of what a live object should contain, in a file that is allowed to change**, with
both legs live since 2026-08-23 — Go guards compare Go against the declaration, and a daily
auditor (`config-key-audit --live-declaration-drift`, 07:00 UTC) compares the declaration against
the live object through each entry's `ProbeSQL`. `Kind` already includes `workflow`, and
`ClaimedItemTimeoutExclusions` is the worked precedent for a Go list generating the fragment a
declaration asserts.

That is where this fix belongs, and it is why fix candidate 0 (a parity test) and candidate 1 (one
definition) are the same change here rather than two.

### ⚠ Adjacent, not ours

`platform/livespec` is **RED at HEAD** on `TestNoNewMigrationFileReadersOutsideTheAllowList`,
failing on the 405 lane's committed `write_audit_findings_origin_test.go` (`ffa1707b3`). Clean in
the working tree, so it is a committed-HEAD failure. Run this lane's tests by name so that
does not mask the result.

---

## 2026-08-26 (later) — the vocabulary is SIXTEEN values, and the drift is already realised at the OTHER reader

### The census, taken over live AND archive because that is what the file's own §c is about

`spec->>'reason'` on `item_type='page_rerender'`, `site_work_items` UNION `site_work_items_archive`,
`[MEASURED 2026-08-26]`:

```
<none>                          18165   17844 of them via the shared creator — CORRECT; a
                                        site-wide refresh IS supposed to be assemble-only
cta_links_stale                  1905 ✓ in gate
section_data_resolved            1428 ✓ in gate
template_changed                  390 ✓ in gate
verbatim_adoption_deploy           86 ✗ NOT IN GATE
light_palette_chrome_replaced      13 ✗ NOT IN GATE   first seen 2026-08-25 — ONGOING
"migration 415 repointed .article-body__content a…" 11 ✗ free prose
image_landed                        6 ✓ in gate
meta_description_corrected          4 ✗ NOT IN GATE
"the 20:2x rewrite deployed these pages before…"     4 ✗ free prose
"bugs_open/238: the £149 rewrite dropped…"           4 ✗ free prose
legal_page_publish                  3 ✗ NOT IN GATE
"section_edit a007f0ff complete + tool-list removed" 1 ✗ free prose
listing_stale                       1 ✗ NOT IN GATE   first seen 2026-08-24 — ONGOING
m2_rebuild_safety_proof             1 ✗ NOT IN GATE
claims_corrected                    1 ✗ NOT IN GATE
```

### The drift is REALISED at the gate, not only latent at the Go reader

**129 `page_rerender` items carry a reason the live gate does not know. All 129 were handled by
`page-rerender` — the gate's own agent — and 96 COMPLETED.** By the gate's own structure
(`else_step: render_page`) every one took the assemble branch.

The bug file's corrections establish, correctly and by measurement, that **zero** items ever
reached the stale **Go** reader, so that arm is latent. But it bounds the *gate*-side instance at
"7 historical `literal_markdown` items… not chased here". **It is 129, across eleven distinct
reason values, and two of those values first appeared in the last two days.** Same asymmetry,
same silent direction, live and ongoing.

### ⚠ WHAT I COULD NOT ESTABLISH — the discriminator FAILED and I am recording that, not hiding it

I tried to convert *"took the assemble branch"* into *"therefore shipped nothing"*. The
`migration 415` cohort was the strongest candidate because its own reason text says *"this page
still serves the raw rule"* — a checkable claim. Result: 1 of 3 components on **every one of the
11 pages** now carries `--color-primary-ink`, **including the pages whose items were CANCELLED**.

The control and the treatment agree, so the marker arrived by some other route and its presence
is not attributable to these items. **The cohort proves nothing either way.**

So the honest split, and it must travel with the number:

| claim | status |
|---|---|
| 129 items carry a reason the gate does not know | **MEASURED** |
| all 129 went to `page-rerender`; 96 completed | **MEASURED** |
| every one took the assemble branch | **MEASURED**, by the gate's own structure |
| therefore they shipped nothing | **NOT ESTABLISHED** |

⚠ The temptation at this point is to try the next cohort, and the next, until one agrees. That is
how the estate's worst measurements get made. **A cohort is only evidence if a marker can be
ATTRIBUTED to it** — which needs pages that did NOT get the item as a control, and here the
control had the marker too.

### The structural finding, which is the real design input

**`spec.reason` is TWO FIELDS WEARING ONE NAME.** Four of the sixteen values are free prose —
whole sentences, a `£` sign, a bug reference, an operator's note to themselves. Humans are using
`reason` as an ANNOTATION while the gate uses it as a ROUTING KEY.

Three consequences, and I think they decide the fix's shape:

1. **A parity test over "the five" is not enough.** It keeps Go and the gate in step and leaves
   the sixteenth free-text value silently assembling, for ever.
2. **The single definition must also answer "what happens to a reason nobody declared?"** Given
   the fail-toward-assemble asymmetry, the safe answer is probably not "assemble silently". An
   unknown routing key that completes green is this bug in one sentence.
3. **The vocabulary spans at least THREE item types**: `template_changed` also appears on **65
   `section_edit`** items, and `literal_markdown` appears ONLY on `item_type='literal_markdown'`
   items — **never on a `page_rerender` item at all**. So the gate's fifth value may not be
   exercised through this path; check before assuming it is.

---

## 2026-08-26 (evening) — BUILT and committed; council said REVISE and was right twice

Commit `ef4236b4d`. Council `f2e4ac2a-2bfc-4c82-ac99-d5fd7616edef` — **round 1 REVISE**
(gating HIGH from `debug_historian`), round 2 resubmitted on the same trail.

### What shipped

The vocabulary is one list in `platform/livespec/rerender_reasons.go`, carrying per value whether
it scopes the fan-out by component and whether it stamps without one — because **the two gates in
`create_rerender_items` are DIFFERENT TESTS** and putting a reason in the wrong one is a new
defect. `scoped` narrows WHICH PAGES; `stampReason` decides whether the item carries a reason at
all, which is what the live gate branches on.

Three declarations: the gate's condition (`PhaseGoSide`), a **paired count** (`PhaseLiveAudit`),
and the fixer's raw INSERT (`PhaseLiveAudit`). The count is not decoration — the whole-clause
fragment is **not self-bounding** here, because unlike the exclusions clause this condition has no
terminator, so a sixth reason appended live leaves the declared prefix present and `Min:1/Max:1`
green. A fragment sees loss; only a count sees addition, which is the direction this bug drifted.

Migration **656** gives the fixer the page-status filter it never had (`bugs_open/098`'s mechanism
at a seam its sweep did not reach).

### The two objections that were worth more than the verdict

1. **`debug_historian` [HIGH]: the page-status literal was CHOSEN, not measured.** Not that it
   was wrong — that nobody had run the enumeration the estate requires before any status-scoped
   query. True. `[MEASURED 2026-08-26]` `pages.status` holds exactly **active 948 / archived 68**;
   `PageWantedLivePredicateFor("p")` renders **byte-identically** to the literal the migration
   writes; and the containment across ALL components is **31 archived pages / 111 non-owned
   component instances** — the bug file's 16-of-60 tool-cta figure is one component's view of it.
2. **`bug_historian` and `reuse_agent`, independently: the deferral was the bug.** I had left the
   two per-value gates in `rerender_page_sections_action.go` as bare literals on an edit-budget
   argument, and both seats named it as 016b §9's *"one call site of a shared judgement gets the
   rigorous fix; the sibling stays heuristic"* — **this bug's own thesis turned on its own plan.**
   Now all four readers name a constant, the list is built from the constants, and a test pins
   both directions, so retiring a value breaks COMPILATION everywhere rather than silently
   disarming one gate.

### Misstep 6 — the lint that would have caught 2026-08-18 matched only PROSE COMMENTS

It reported **12 reason literals checked** and passed. Every one was a comment: a migration writes
the gate condition **inside a SQL string literal**, so its quotes are DOUBLED
(`== ''template_changed''`), while the explanation eight lines above writes them singly. My regex
accepted single quotes only, so it could not see either executable line it exists to catch, in
either file it was written for.

**Third distinct failure mode in one day for the same family**, and the differences are the
useful part:

| # | lane | why the guard was green and inert |
|---|---|---|
| 1 | 359 | the tests were passing on a **guard in series** — a second failure downstream |
| 2 | 407 | the **fixtures could not produce the failure** (order already alphabetical; utility group empty) |
| 3 | 404 | the **discriminator could not see the real shape of its input** — right corpus, wrong dialect |

Fixed with both quotings AND **positive controls**: named values that must be found in named files
(460 → `template_changed`, 473 → `literal_markdown`). Mutation-proved — narrowing the regex back
now fails with `POSITIVE CONTROL FAILED` instead of passing with a smaller number. **A scan must
be pinned to a KNOWN POSITIVE, never to a non-zero count**: "it found 12 things" answers *did it
run*, never *can it see what it is for*. Full entry in `WRONG_CALLS.md`.

### Still open, and it is the honest residual

`spec.reason` is two fields wearing one name, and this change does not fix that — it makes an
undeclared reason LOUD while still assembling. Splitting annotation from routing key changes what
the field GUARANTEES and is RFC-scope; the census and the warning are the evidence pack that RFC
would need. Also deferred: the 7 pre-473 `literal_markdown` items (a different mechanism, one
layer up), and any retro-repair of the 129 (86 are assemble-by-design and no discriminating
marker exists).

---

## 2026-09-02 — ROUND 3 SUBMITTED, and re-measuring corrected one of our own answers

Picked up from `bugsweep_2026_08_26/HANDOFF_2026-08-26_continue_here.md` §5, which recorded
round 3 as owed with the gating fix already in the tree. Confirmed before acting: exactly
**two** `council_report` artifacts exist on `f2e4ac2a`, both `revise`, both 2026-08-26 —
so no third round had been submitted in the intervening week.

⚠ **`orchestration_states` is the wrong table to ask.** It returned ZERO rows for this
correlation, which reads like "never submitted" and actually means the runs have aged out.
The verdict is an artefact: query `diagnosis_artifacts`. (This is the exact trap recorded in
`WRONG_CALLS.md` on 2026-09-02 by the lane that reported a verdict as "still running".)

### What each r2 objection's state actually was

| objection | state at r3 |
|---|---|
| `debug_historian` [HIGH] edit 7 — 656's `SELECT INTO` had no `STRICT`, no row-count guard | **FIXED in the shipped file**, verified by reading it: an explicit count guard that RAISEs unless exactly 1 active row, and `INTO STRICT` on BOTH the pre-image anchor and the post-UPDATE verify |
| `editquality` [medium] ×2 — sketches did not show the constants or the pin test | **SUBMISSION-ACCURACY failure, not a code one.** Both exist in the tree; the r2 sketches were written from the design. Corrected in r3 with verbatim tree excerpts |
| `debug_historian` [medium] — no pod-grep | **DONE** 08-26, both pods, with a negative control in the same exec |
| `bug_historian` [medium] — WARN lives only in `create_rerender_items` | **REAL AND UNFIXED**, now stated explicitly as the residual rather than left implicit. Gate-side loudness is new authority on a shared seam and ships opt-in default-OFF under the owner ruling 2026-08-02 §2 |
| `guardian` [medium] — enumerate callers seeing a behaviour change | **RE-MEASURED, AND OUR PREVIOUS ANSWER WAS TOO STRONG** — see below |
| `guardian` [low] — `platform/livespec` RED at HEAD | **STILL RED, seven days on.** Reported with its current state, not its r2 state |

### The correction: "MEASURED: ZERO" answered a different question than the seat asked

The 08-26 handoff recorded the guardian objection as **"MEASURED: ZERO, over full live+archive
history"**. That figure is true of realised **ITEMS**. The seat asked about **CALLERS**. Those
are different questions, and the item answer cannot settle the caller one — because **pre-fix
the action DISCARDED the reason**, so a caller that passed `template_changed` produced an item
carrying no reason at all and is invisible to any query keyed on `spec->>'reason'`.

Answered config-side instead `[MEASURED 2026-09-02]` — three live agents call the action:

- `nav-updater` — passes NO reason. Mechanically unaffected.
- `tool-generator` (step `enqueue_rerender`) — passes NO reason. Mechanically unaffected.
- **`rerender-pages` — passes `input_data.spec.reason` THROUGH**, with `component_id`.

So there **is** one caller whose behaviour changes, and the honest submission concedes it
rather than repeating the zero.

### The discriminator, stated because it is disconfirmable

`[MEASURED 2026-09-02]` of 78 live `page_rerender` items on `template_changed`: **64 keyless**
(the fixer's raw INSERT signature) and **14 keyed** — and all 14 carry a **SITE** id in the key,
which is migration 615's fan-out shape
(`'page_rerender_'||page_name||'_'||site_id||'_template_changed'`). **ZERO carry a PAGE id**,
which is the action's shape. No `literal_markdown` items exist at all.

The tell that makes it checkable rather than asserted: two `gaswholesalers.com` rows share ONE
uuid across DIFFERENT pages, so that uuid is a site, not a page. `[MEASURED 2026-09-02]` 3 of
the keyed items are dated 08-25 (615's apply) and 11 are dated 08-31, so 615's fan-out ran
again — worth knowing, not a defect of this fix.

⚠ **The keyless/keyed heuristic recorded earlier ("334 of 338 keyless") is NOT the same
discriminator** and would have given the wrong answer here: it counts 615's hand-rolled INSERT
as if it were the action, because 615 deliberately composes an action-shaped key. Resolve by
whether the uuid is a site or a page.

### Adjacent, flagged not fixed

- **`platform/livespec` is RED at committed HEAD** — `TestNoNewMigrationFileReadersOutsideTheAllowList`
  fails on `platform/orchestration/actions/write_audit_findings_origin_test.go` (405 lane,
  `ffa1707b3`). Both paths clean in the tree, so it is committed breakage, unchanged for seven
  days, on the same allow-list seam this plan touches. The test's own text names the two
  sanctioned remedies. Not touched — it is the 405 lane's file.
- The 097 trigger now WARNs that `_RELOCK` is an unclassified migration suffix in
  `sql_for_agents`, treated as IN scope by the safe default. Someone should classify it in
  `COUNCIL_SCOPE_NOT_THE_CHANGE_RE` if it is not the change (`bugs_open/314`).

**Submitted** on the same correlation, `RESUBMIT_CORR=f2e4ac2a-2bfc-4c82-ac99-d5fd7616edef`,
so the trail accumulates. Draft: `scratchpad/submission_404_r3.json` (session scratch).
Verdict not yet read at time of writing — the weaker true statement, not "still running".

### 2026-09-02 (later) — r3 verdict: REVISE, and the gating objection was MY repeat of r3's own lesson

`[MEASURED 2026-09-02]` verdict landed 13:53:03Z: **revise**, `decided_by: gating objection from
editquality`, 3 abstained. **Nine seats approved** (`reuse_agent`, `guidelines`,
`tooling_provenance`, `adoption_guardian`, `improvement_guardian`, `render_guardian`,
`debug_historian`, `constitution`, `mission`, `architecture`). The design drew no objection in
either round — every gating objection so far has been about SUBMISSION ACCURACY.

**The gate: `editquality` [HIGH] on edit 7, seconded independently by `prior_art_librarian`
[HIGH].** r3 existed to fix "the narrative claims what the sketches do not show", and I corrected
edits 1, 4 and 8 while leaving **r2's pre-fix sketch on edit 7 — the one edit the r2 gating HIGH
was about.** Full entry in `WRONG_CALLS.md`; the transferable half is that I *had* verified the
guard at the artefact, and that real check made the claim feel safe while the thing I was actually
publishing went unchecked.

Second self-inflicted fault the same round: correcting edit 4 I **replaced** its sketch instead of
adding to it, deleting the corpus-lint test that edit 3's rationale still referenced.

### r4, submitted with the check that would have caught it

Before submitting r4 I asserted on the SUBMISSION rather than the tree — a loop over `plan.edits`
checking each sketch contains the strings its rationale claims, plus a negative control that the
superseded text is gone. Nine PASS lines, under a second. Any round claiming "sketches corrected"
can afford this.

| r3 objection | r4 |
|---|---|
| `editquality`/`prior_art_librarian` [HIGH] edit 7 — sketch showed neither guard | Sketch is now the shipped file verbatim (lines 80–121): count guard, the RAISE, `INTO STRICT` on both reads, idempotency-before-snapshot |
| `prior_art_librarian` [medium] — "exactly 1 active row" asserted with no query | Query shown and **re-run today**: component-template-fixer 1, page-rerender 1, rerender-pages 1. States it proves the guard INERT, not unnecessary |
| `editquality` [medium] edit 4 — corpus lint no longer shown | Both tests shown, including the positive controls naming migrations 460 and 473 |
| `reuse_agent` [medium] edit 1 — `operation:add` vs "already in the tree" | Framing restored: this code is COMMITTED AND LIVE, review here is after-the-fact by the owner ruling 2026-07-29 §2 |
| `debug_historian` [medium] edit 7 — `_ROLLBACK` not in the edits list | It exists; `_ROLLBACK` sidecars are refused client-side by `council-scope.sh`. r2 said so and r3 dropped the sentence |
| `guardian` [low] edit 5 — census blind to hand-dispatched callers | **Conceded and named.** A config scan cannot see a kcat/kubectl dispatch — something this lane did today for a nav rebuild |
| `guardian` [medium] edit 3 — allow-list RED at HEAD | Still red, still the 405 lane's file, named with no mitigation claimed |
| `tooling_provenance` [low] edit 7 — "every RAISE cites '655:'" | ⚠ **FALSE, and checked rather than obeyed.** `'656:` appears 9 times, `'655:` zero. Complying would have introduced the defect it warned of |

Submitted on the same correlation. `RUN_ORCH_ID=40639f27-fdca-4059-92bd-1a01d9f55f57`.
Draft: `scratchpad/submission_404_r4.json`. **Verdict not read at time of writing.**

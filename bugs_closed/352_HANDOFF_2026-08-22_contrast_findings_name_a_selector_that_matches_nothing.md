# 352 — contrast findings name a selector that matches NOTHING, so the fix is authored, deployed and inert: a class-less element is filed with `Class` = its TAG NAME

> # ✅ CLOSED 2026-08-25 — fixed, live, and proven at the artefact. Arm 2 was SPLIT OUT, not dropped.
>
> **Owner ruling, 2026-08-25: split arm 2 into its own bug and close 352 against arm 1.** The two arms
> shared a symptom and nothing else — different mechanism, different remedy — and keeping them stapled
> made a finished fix read as unfinished.
>
> **Arm 2 now lives at `bugs_open/390_HANDOFF_2026-08-25_a_correct_contrast_selector_still_loses_the_cascade_so_the_repair_is_authored_and_inert.md`**
> — still open, still live, and its mechanism is now verified first-hand rather than sketched (the
> verification **corrected** the remedy this file proposed for it: see 390 §2, the offending *value*
> is reachable even though the *declaration* is not, so this file's "refuse and park" precondition
> would have parked repairable findings).
>
> ⚠ **Anything deferred "until 352 is fixed" should now read 390, not this file.** Closing a bug does
> not retract the deferrals pointing at it.
>
> ## What closing asserts, with its evidence
>
> | | |
> |---|---|
> | fix | `ffa6e1c3d`, council **APPROVED** round 1 (`acadbe8b-f131-4d4b-b4de-5b61f0898f93`) |
> | live | `v1.0.1334` 2026-08-24 15:39 UTC; still carried on `v1.0.1337` (`4c996e1b5`), ancestry checked with a dated control |
> | proven | two scheduled audits straddle the roll — 15:31:50 filed 47 rows with **3** invented; 17:33:16 filed 10 with **0**, all `verified/v1`. Then settled in the page: `.ported-page-content A` counted **15** and **8** independently, matching the producer exactly; two pre-roll rows' `SPAN.SPAN`/`LABEL.LABEL` counted **0** against 22 real `<span>`s and 6 real `<label>`s |
> | held | [MEASURED 2026-08-25 09:40 UTC] **15** rows filed since the roll, **0** invented, **15/15** carrying `selector_scheme` and `matches` |
> | legacy rows | migration **587** applied by hand 2026-08-24 19:11:22 UTC, `UPDATE 73` — withdrawn, **not** resolved |
>
> ⚠ **One dated check is still owed and is NOT a reason to reopen this file: from 2026-08-28**, any of
> the 13 sites whose rows 587 withdrew that has been re-audited and has a visible contrast fault with
> no re-filed `contrast_failure` is a defect in 587's promise. All 13 are due by ~2026-08-27 21:30 UTC
> (the rotation window is **3 days**). Query and context: the lane's `HANDOFF_2026-08-25_continue_here.md` §4(2).
>
> ⚠ **The permanently-quotable damage figure is 111** [MEASURED 2026-08-25 09:40 UTC] — `complete`
> rows carrying an invented selector, i.e. repairs recorded that could never have applied. 587 never
> touches those. It read 108 earlier on 2026-08-24 and grew by three that the last pre-roll audit
> filed minutes before the fix rolled.
>
> ---
>
> ## The 2026-08-25 09:40 status, kept for the trail — ARM 1 **DONE**; ARM 2 WAS THE ONLY REASON THIS FILE WAS OPEN
>
> **Cold-start for this lane:**
> `docs/agent_docs/docs024_key_docs_latest/bugfix_352_invented_selector/HANDOFF_2026-08-25_continue_here.md`
>
> **Arm 1 has now held for a day of ordinary fleet traffic** [MEASURED 2026-08-25 09:40 UTC]: **15**
> `contrast_failure` rows filed since the roll, **0** invented, **15/15** carrying `selector_scheme`
> and `matches`. Still live after two further fleet rolls (`v1.0.1337`, all three services stamping
> `4c996e1b5`, ancestry checked with a dated control). **Nothing about arm 1 is outstanding.**
>
> **To close this file you need arm 2 fixed and live — or arm 2 split into its own bug and 352 closed
> against arm 1.** The second is a legitimate shape (the arms share a symptom and nothing else) and
> is an **owner call**, not a session's.
>
> ⚠ **One dated check comes due 2026-08-28**: 587 freed 73 dedup slots on the promise that
> still-failing pairings return under verified selectors. 0 of 56 have returned so far — expected,
> because **0 of the 13 sites have been re-audited since 587 applied**. The rotation's live window is
> **3 days**; all 13 are due by ~2026-08-27 21:30 UTC. From 2026-08-28 a non-return is a defect.
>
> ### The 2026-08-24 19:15 status, still accurate for what it covers
>
> **This file stays OPEN, and the reason is arm 2.** Read this banner before quoting
> anything below it.
>
> | arm | state |
> |---|---|
> | **Arm 1 — the producer invents the selector** | **FIXED, SHIPPED, AND PROVEN ON A LIVE PAGE.** `ffa6e1c3d`, council-approved round 1 (`acadbe8b-f131-4d4b-b4de-5b61f0898f93`), live on both images since `v1.0.1334` (15:39 UTC) and re-confirmed on `v1.0.1335` (18:32 UTC). Proof below. |
> | **Arm 2 — a correct rule still loses on source order** | **LIVE, reproducible, untouched.** Nothing in this change addresses it. |
> | the 73 legacy rows | **WITHDRAWN.** Migration `587` applied by hand **2026-08-24 19:11:22 UTC**, `UPDATE 73`. |
>
> ### The proof, at the artefact — a before/after pair from the SAME producer, 2 hours apart
>
> Two scheduled render audits straddle the 15:39 UTC roll. Nothing was staged for this; the
> rotation simply ran, which is why it is worth more than the driven canary it replaced.
>
> | | pre-roll run, **15:31:50 UTC** | post-roll run, **17:33:16 UTC** |
> |---|---|---|
> | rows filed | 47 | 10 |
> | invented (`TAG.TAG`) | **3** | **0** |
> | `spec.selector_scheme` | absent on all 47 | `verified/v1` on all 10 |
> | `spec.matches` | absent | present on all 10 |
>
> **Then settled in the page itself, not in the database.** Fetched the live pages over HTTPS
> (invented-path control on each domain returned 404, so a 200 means a real page) and counted
> matches with an independent stdlib parser that walks the open-element stack:
>
> - post-roll `.ported-page-content A` on `/guides/index.html` → **15 matches**, and on
>   `/guides/jargon-buster.html` → **8**. Both agree **exactly** with the `matches` the producer
>   recorded, and **all 23 matched elements are class-less `<a>`s** — precisely the population the
>   old code filed as `A.A`. Controls: a non-existent class → 0; the same selector on the 404 body
>   → 0; `class="A"` and `class="H3"` appear **nowhere** on the page.
> - pre-roll `SPAN.SPAN` on `loanzy.uk/tools/loan-repayment-calculator/` → **0 matches**, while
>   **22** `<span>`s genuinely exist. `LABEL.LABEL` → **0**, against **6** real `<label>`s. Both
>   rows are already `complete`: two more false repairs, recorded today, eight minutes before
>   the fix rolled.
>
> The measurement could have come out otherwise, which is the point: the pre-roll arm *did*
> produce invented selectors, on the same day, from the same code path.
>
> ### ⚠ CORRECTION TO THIS FILE'S OWN CANDIDATE (1) — the naive fix is a REGRESSION
>
> Candidate (1) below says to omit the class component so the selector reads `h3`
> rather than `H3.H3`, and names only the dedup interaction as the risk. That is right
> about the producer and **wrong about the consequence.**
>
> Today `p.P { color:#fff }` matches **nothing**, so it is inert and harmless.
> Corrected to `p`, it matches — and css-patch-agent's own live prompt says *"The
> platform APPENDS your rules to the END of the stylesheet"*, one stylesheet per site.
> So the "fix" would recolour **every paragraph on the site**. [MEASURED 2026-08-24]
> `P.P` (77) and `A.A` (44) are **121 of the 181** invented selectors: the two
> commonest cases are the two worst possible bare selectors, across 13 sites.
>
> The shipped fix therefore composes the selector **in the page** — class → own id →
> nearest ancestor carrying an id or class → bare tag — and **asserts it selects the
> very element that was measured**. A bare tag is refused and counted. The invariant
> is not "stop lying", it is "prove it": any *future* composition defect is now
> self-disclosing, not just this one. (The `bugs_open/198` lane, which wrote the
> original candidate, has accepted this correction against its own text.)
>
> ### What the numbers actually were [MEASURED 2026-08-24]
>
> 181 of 452 `contrast_failure` rows carried an invented selector; **108 of them were
> already `complete`** — repairs recorded against rules that could never apply. 92
> filed in the last 7 days, so it was actively producing, not history. Of the **171
> open, still-failing** rows (the `brochure_component_library` lane's "durable 185"),
> **73 (43%) across 13 of 15 sites** were the invented kind — that lane had an open
> owner decision to release them to the fixer and has been told.
>
> ### Two hazards the obvious fix creates, both designed for
>
> 1. **False retraction.** `item_key` embeds the selector and `retractResolvedContrastFindings`
>    builds its still-failing set the same way, so a key-shape change would close **73 open
>    rows** stamped *"no longer below its contrast threshold"* — false, on a path that has
>    **already closed 79 rows** for real. Ordering cannot fix it (the window is symmetric), so
>    the legacy composition is inserted as an **alias key** and a **scheme guard** stops an old
>    adapter grading a new-shape row.
> 2. **Blast radius**, above.
>
> ### ⚠ TWO IMAGES — a chassis roll ships only half of this
>
> `internal/adapters/browserrunner` compiles into `cmd/browser-runner-adapter` and
> nothing else; `render-audit-adapter` runs that same image (makefile:107). So the
> producer half needs `make build-browser-runner-adapter` **and** the overlay `newTag`,
> and the consumer half rides `agent-chassis`. **"The chassis rolled" is not evidence
> this is live.**
>
> ### The 73 legacy rows — **WITHDRAWN 2026-08-24 19:11:22 UTC**
>
> `docs/agent_docs/sql_for_agents/587_retire_invented_contrast_selectors_HOLD.sql`
> **withdrew** them as `cancelled` — withdrawal, **not** resolution — freeing the dedup
> slot so still-failing pairings return under verified selectors. Its ordering gate (both
> images confirmed at the artefact) was met, and it was applied by hand: `UPDATE 73`.
> Post-checks: `open_invented = 0`, `withdrawn = 73`, every row carrying its prior status,
> `falsely_completed = 0`.
>
> ⚠ **So every census in this lane's docs now returns ZERO for the open population — by
> design, by SUBTRACTION, which reads as "this never happened".** The figure that keeps
> returning 73 for ever is the recovery query in the lane's RUNBOOK §10; use that one.
> The **`complete` + invented** count is untouched by 587 and is the permanently-quotable
> damage figure: **111 as of 2026-08-24 19:10 UTC** — it was 108 earlier the same day and
> grew by the three false repairs the 15:31 audit filed while the fix was minutes from rolling. Re-detection window is **up to a fortnight** (measured:
> all 13 sites audited within 14 days, only 3 within 7) — an earlier draft said "weekly"
> and that was an overstatement, caught by the council's prior-art seat.
>
> ### Arm 2 — what it would take, so it is not lost
>
> Before planning a patch, css-patch-agent's workflow needs a **measurable precondition**:
> grep the editable stylesheet (`css_themes`) for a declaration governing the filed
> selector's property; if the offending declaration lives in page-level component CSS
> emitted *after* it, **refuse and park** with a `parked_by` marker (198's
> `mark_base_unsafe` shape) rather than append a rule that cannot win. And completion
> should consult the spec's own `acceptance_test` at the `checks.GetVerifier` /
> `verifyBeforeComplete` choke point — which `write_audit_findings_verifier_join_test.go:85`
> confirms **nothing reads today**. Not designed further; that is the next session's work.
>
> **Working record:** `docs/agent_docs/docs024_key_docs_latest/bugfix_352_invented_selector/`
> (PLAN, RUNBOOK, NOTES, README_where_we_are, council submission). Register: **VIZ-016**,
> and **WII-016** for the key-shape statement.


Filed 2026-08-22 by the bugfix-198 lane, spun out of `bugs_open/198` candidate (6) at close-out.
198's own defect (the stylesheet clobber) is fixed, live and closed; **this is the other thing
that lane's evidence turned up, and it is a different defect with a different cause.**

**090 substitution stated plainly (owner ruling 2026-07-31):** not filed through the diagnosis
loop, because the entire causal chain was read first-hand in the production path this session —
the two lines below in `render_audit_action.go`, the live `site_work_items.spec` shape, and
three sites' worth of deployed rules. Nothing here is inferred. A 090 run would re-read the same
two lines.

## The mechanism, at the producer — and the agent is FAITHFUL, not confused

`internal/adapters/browserrunner/render_audit_action.go` (this is the production audit; the
identical code in `scripts/render_audit.py:139` is the local probe, not the filer):

```js
// :221 — inside the in-page contrast sweep
var cls = (typeof el.className === 'string' ? el.className : '') || el.tagName;
```

```go
// :329 — and that value is filed under a field whose NAME asserts it is a class
URL: url, Tag: c.Tag, Class: c.Cls, Text: c.Text, FG: c.FG, BG: c.BG,
```

**For an element with no class, `cls` falls back to `el.tagName`** — uppercase, per the HTML DOM
— and is then recorded as `Class`. Downstream the finding is labelled `TAG.Class`, which for a
class-less `<h3>` composes to **`H3.H3`**. As CSS that selects elements carrying
`class="H3"`, of which there are none.

**Nothing downstream can undo this**, because the two cases are indistinguishable by shape:
`SPAN.calc-eyebrow` (a real class) and `H3.H3` (a fallback) arrive in the same field, in the same
format. css-patch-agent turning `H3.H3` into `H3.H3 { color:#ffffff }` is the correct reading of
what it was told. **The fault is a producer that emits a tag name in a field called `Class`.**

A sibling in the same package does it differently and better —
`run_checks_action.go:1123` takes the FIRST class and falls back to `tagName.toLowerCase()` —
which shows the fallback is a choice, not an accident. (Note lowercasing alone does NOT fix it:
`h3.h3` still matches nothing. The fix is to omit the class component entirely when there is no
class, so the selector is `h3`.)

## Evidence — three sites, and one item that proves the damage is silent

| site | rule the agent deployed | element actually carries | filed item |
|---|---|---|---|
| dartsonline.com | `H3.H3 { color:#ffffff }` | no `class` at all | `H3.H3 on /contact.html`, **status `complete`** |
| remortgagecalculator.uk | `p.P { … }` ×2 (68 chars each) | no `class` at all | 2 × `contrast_failure`, both **`complete`** |

**The dartsonline `H3` row is the one that matters:** an item marked `complete` by
css-patch-agent whose text was **still invisible when measured two days later**. That is
"processed, correctly fixed, and never applied" with a row id behind it. Recorded in `198`
§"CORRECTED 2026-08-20" with the measurement.

Both sites' rules were deliberately **not carried forward** into their stylesheet restores,
precisely because they match nothing — so the evidence is historical rather than currently
present in `css_themes`. `198`'s restore procedure records that decision.

## The SECOND arm — a correct rule can still lose, and this one is not the producer's fault

Even when the selector is right, css-patch-agent's appended rule can be inert: for the
`~1.0x:1` family the offending declaration lives in **page-level component CSS emitted AFTER
the stylesheet the agent edits**, so an equal-specificity rule loses on source order however
correct it is. The dartsonline lane worked around it with `body`-prefixed overrides rather than
`!important`. `bugs_open/296` §10.5 states the same finding from the other end and notes it may
explain a subset of its **durable 185** parked findings directly: *processed, the fix was
correct, and it never applied.*

So there are two independent ways for a `contrast_failure` to complete without repairing
anything, and they need different remedies.

## Why this survives every existing check

- The work item completes **honestly** by the workflow's own lights — a rule WAS authored,
  appended and deployed. `bugs_open/198`'s migrations 542/546 made refusals and failures stop
  minting `complete`; they do not and cannot cover this case, where the write genuinely happened.
- **Each spec already carries an `acceptance_test`** naming the exact single-selector
  re-measurement (`"computed contrast for elements matching X on Y is at least 4.5:1 — a
  single-selector, single-page measurement, not a site re-audit"`). Confirmed present on live
  items 2026-08-22. **It is written by the audit and read by nothing.**
- The next render audit re-measures the same pairing, files it again, and the promoter routes it
  back to the same agent — so the symptom is a finding that keeps returning, not one that fails.

## Fix candidates, ordered by what closes the door

1. **Stop emitting a tag name in a `Class` field** (`render_audit_action.go:221`/`:329`). Emit
   the class when there is one and **omit the class component entirely when there is not**, so
   the finding says `h3`, not `H3.H3`. This makes the bad selector unrepresentable at source and
   is a few lines. ⚠ It changes the `item_key` shape for class-less findings, so check the dedup
   interaction before applying — existing open items keyed `TAG.TAG` will not match new ones.
2. **Refuse and re-file rather than append a rule that cannot win** (198's candidate 6 as
   originally stated). Measurable precondition, and it covers BOTH arms: before planning, grep
   the theme for the selector; if the offending declaration is not in the file the agent can
   edit, refuse. `bugs_open/198`'s `mark_base_unsafe` step is the shape to copy — it is already
   wired to park with a `parked_by` marker rather than mint `complete`.
3. **Use the spec's own `acceptance_test` post-deploy**: re-measure the one pairing at the served
   page and complete the item only on measured improvement. This is the general guard — it
   catches every "authored but inert" cause, including ones nobody has thought of yet.
4. NOT a candidate: teaching the agent to recognise `TAG.TAG` as suspicious. It is guessing at
   the producer's intent from a lossy string, and it would misfire on any site that genuinely
   uses `class="H3"`.

## How to verify a fix

For (1): a class-less element's finding arrives with a selector that a browser actually matches
— check `document.querySelectorAll(<selector>).length > 0` on the affected page for a fresh
finding. For (2)/(3): an item that cannot be fixed must NOT read `complete` — it should park or
re-file, and the served page must be re-measured rather than trusted.

## Related

- `bugs_open/198` (CLOSED 2026-08-22) — where this evidence was gathered; its §"CORRECTED
  2026-08-20" and THIRD WAVE §2 carry the three sites.
- `bugs_open/296` §10.4/§10.5 — the parked-findings backlog this may partly explain; **any
  census of its durable 185 taken before this is fixed cannot distinguish "declined" from
  "fixed but inert"**.
- `bugs_open/211` — a different reason a contrast fix does not take (the alias `:root` block is
  absent), worth reading alongside so the two are not confused.
- 016b §9 — the transferable pattern is filed there as "a fix aimed at a selector the producer
  invented".

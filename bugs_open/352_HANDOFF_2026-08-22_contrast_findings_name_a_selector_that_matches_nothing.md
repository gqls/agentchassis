# 352 — contrast findings name a selector that matches NOTHING, so the fix is authored, deployed and inert: a class-less element is filed with `Class` = its TAG NAME

> ## STATUS 2026-08-24 — ARM 1 IS FIXED AND COMMITTED; ARM 2 IS STILL LIVE
>
> **This file stays OPEN, and the reason is arm 2.** Read this banner before quoting
> anything below it.
>
> | arm | state |
> |---|---|
> | **Arm 1 — the producer invents the selector** | **FIXED at source**, `ffa6e1c3d`, council-approved round 1 (`acadbe8b-f131-4d4b-b4de-5b61f0898f93`). **INERT until BOTH images roll** — see "two images" below. |
> | **Arm 2 — a correct rule still loses on source order** | **LIVE, reproducible, untouched.** Nothing in this change addresses it. |
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
> ### The 73 legacy rows
>
> `docs/agent_docs/sql_for_agents/587_retire_invented_contrast_selectors_HOLD.sql`
> **withdraws** them as `cancelled` — withdrawal, **not** resolution — freeing the dedup
> slot so still-failing pairings return under verified selectors. Held until both images
> are confirmed at the artefact. Re-detection window is **up to a fortnight** (measured:
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

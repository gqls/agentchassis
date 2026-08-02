# RFC 009 — the renderer should enforce `input_schema.on_missing`, instead of trusting every template author to

**Filed** 2026-08-02 by `bugfix_140_contact_info_fabrication`, at the direction of
the council gate's **architecture** and **reuse_agent** seats, which reached this
independently in the same round (`40de12b0-36fa-4c06-82b4-995dc9098593`,
APPROVED with 7 advisory objections).
**Status** **DECIDED 2026-08-03 by the owner: "C now, B next".** C is DONE and live (see the decision record at the foot). B is the next piece of work. A is NOT taken, and the sizing below says why.

> The seat's own verdict on the change that prompted this was **`point_fix`** —
> *"no new namespace, wire shape, or shared runtime contract is introduced … does
> not meet the needs_rfc trigger test on its own terms"* — while recording
> `ARCHITECTURE_SIGNAL: insufficient`. So this RFC is not an appeal against that
> plan and does not ask for it to be revisited. It exists because a **scope
> observation is answered by routing it, not by resubmitting the same plan with
> better measurements** (CLAUDE.md, owner ruling 2026-07-28).

## The claim

`content_components.input_schema` declares, per field, what to do when the datum
is absent:

```json
"hours":   { "source": "site_specs.identity.hours", "on_missing": "skip_field" },
"section_title": { "source": "llm", "fallback": "Contact Us", "on_missing": "use_fallback" }
```

**Nothing reads `on_missing` at render time.** It is documentation. Obedience is
hand-implemented by whoever writes the Go template, in `{{if}}` branches, per
component, per field, for ever.

## The evidence that this does not generalise

It has now failed **twice**, and both times the repair was a hand-rewritten
template rather than a mechanism:

| | what happened | fix |
|---|---|---|
| `bugs_open/111` (2026-07-28) | the fallback footer rendered an `<h4>Contact</h4>` over an empty mailto | `RenderFallbackFooter` gated by hand (`d4731109d`) |
| `bugs_open/140` (2026-08-02) | `contact-info` substituted `+1234567890`, `Monday – Friday, 9am – 6pm`, `info@example.com` for absent data — **8 live sites served the invented hours**, one served the invented phone as a `tel:` link | migration 287 gated each card by hand |

In the second case the schema said `skip_field` for exactly the three fields that
fabricated. The contract was correct and the template ignored it, and **no layer
between them could tell**.

Two costs are measured rather than asserted:

- `hours` is supplied by **0 of 1,089** `page_components` fleet-wide, so every
  Hours card the platform has ever served was fabricated.
- The desync in that same row was detected by `compute_component_quality.go` on
  **2026-05-18** (`schema_template_synced = false`) and consumed by nothing for
  eleven weeks.

## Why the current defences do not cover it

- **`scoreComponent`** (`compute_component_quality.go`) is absolute and
  single-artifact, and does not inspect `{{else}}` branch CONTENTS at all. It
  scored this component 80 while it was fabricating.
- **`component_write_guard.go`** is COMPARATIVE by design — "is this replacement
  worse than the row it replaces" — so a birth write carrying a fabricated
  fallback passes cleanly. Its header says as much and says why.
- **`check_placeholder_contact`** matches a roster of literals against rendered
  HTML. It contained not one literal our own library ships; across every unlocked
  row its nine patterns matched **1** while missing **9** fabrications.
- **`scripts/check_placeholder_fallbacks.py`** (CGV-029, shipped with 140) reads
  the live library and separates a fact default from a label default. **This is
  the closest thing we have, and it is advisory, by-hand and post-hoc.** It
  catches the third instance after it exists, not before.

Every one of those is downstream of the template author's decision. None is the
enforcement point.

## SIZING, measured 2026-08-02 after the RFC was first written — and it changes the picture

The RFC originally argued from two incidents. Here is the corpus, which is a
better basis for the decision and which nobody had measured.

**1. The contract is mostly ABSENT, not mostly disobeyed.** Across every active
component's `input_schema.fields`:

| `on_missing` | fields | components |
|---|---|---|
| **(not declared at all)** | **1,938** | **116** |
| `skip_field` | 181 | 56 |
| `use_fallback` | 21 | 9 |
| `skip_section` | 15 | 14 |
| `needs_human_review` | 8 | 6 |

**~90% of fields (1,938 of 2,163) declare no `on_missing` whatsoever.** So a
render-time gate driven by `on_missing` would be **inert for nine fields in ten**
on day one. Its reach is not a function of how well it is built; it is a function
of a declaration most component authors do not make. That is the single most
important number here and it cuts against option A being a general answer.

**2. The live violation surface is 68 fields across 20 components** — declared
`skip_field`, referenced by their template, with no `{{if .field}}` gate anywhere:

```
platform-comparison   15   (row1_platform1_value … row5_spark_value)
product-specs          8   (spec_1_value … spec_8_value)
system-stats           8   (stat1_label … stat4_description)
featured_article       6   hero-tool 5   case-studies-grid 5   Pricing Tiers 4
archetype-result-card 3   bayesian-ranking-hero-tool_pre_037 3
about-hero / case-studies-hero / contact-hero / hero / services-hero  1 each (subheadline)
content-listing 1   portfolio-showcase 1   social_proof 1   about-commercial-block 1
```

*[APPROXIMATE — this tests for a gate anywhere in the template, not a block-stack
scope check, so it over-counts a field gated in one place and used bare in
another. It does not under-count.]*

**3. And these are a DIFFERENT, milder class than `140`.** This is the
distinction the decision turns on:

- **`140`'s class — ungated AND substituting a literal.** The platform *asserts a
  false fact*: an invented phone number, invented opening hours. **Fleet-wide
  count today: 0.** `check_placeholder_fallbacks.py` (CGV-029) covers exactly this
  and reports clean across 173 active components.
- **The 68 above — ungated with NO fallback.** The platform renders a *blank*: an
  empty table cell in `platform-comparison`, an empty spec row in
  `product-specs`, a missing subheadline. Bad, sometimes visibly so on a spec
  sheet, but it **asserts nothing untrue**. Nothing currently detects it.

So the harm this RFC would prevent is mostly **blanks, not fabrications** — the
fabrication class is already closed by a lint that needs no roll and no schema
declaration. That materially weakens the urgency argument in the section above,
and it is why the recommendation below has changed to prefer the cheap options.

## The shape being proposed (not yet designed)

A render-time gate driven by `input_schema.on_missing`, applied **once** at the
`executeGoTemplate` / `RenderContext` layer: a field declared `skip_field` that
is absent yields nothing regardless of what the template says, and a field
declared `use_fallback` yields its declared `fallback` value.

## The hard questions, stated so nobody thinks this is easy

1. **Where does the schema reach the renderer?** `executeGoTemplate` receives a
   `map[string]interface{}` from `contextToInterfaceMap`, not the component row.
   The schema is not currently in scope at the point of execution, so this is a
   plumbing change before it is a policy change.
2. **`RenderContext` can legitimately supply what `content_data` lacks.** It
   carries top-level `Email`/`Phone` whose json tags reach the template contract,
   so "absent" is not simply "missing from `content_data`" — `idea.uk` renders a
   phone from site identity. Any enforcement must define absence across every
   path a value can arrive by. This is the same unsoundness that killed the
   roster-free version of the 140 detector.
3. **A template can already contradict the schema in the other direction**, and
   some of those are correct today. An enforcement layer that is stricter than
   the fleet's live templates would break working pages on the roll — the classic
   over-fire that gets a guard switched off.
4. **Who owns the migration path?** 172 active components exist; the ones whose
   `{{else}}` branches are legitimate labels must keep working unchanged.
5. **Is the right layer the RENDERER or the WRITE PATH?** Refusing a fact-shaped
   `{{else}}` at `store_generated_component` closes the door at birth and cannot
   break an existing page, but does nothing about the components already there.
   These are different trades, not the same fix at different depths.

## What is NOT being claimed

- Not that the 140 fix was wrong or should be redone. It is the correct point
  fix, it is live, and it was approved.
- Not that a third instance exists. `check_placeholder_fallbacks.py` reports
  **clean across all 172 active components** as of 2026-08-02. The exposure is
  prospective.
- Not that this is urgent. Two instances in eleven weeks, one now closed at
  source, with a standing lint that would surface a third.

## THE DECISION — four options, costed

Not "should we do this" but "at which layer, and is it worth it". The sizing above
should be read first; it moved my own recommendation.

**A — Render-time gate.** `executeGoTemplate` reads `on_missing` and enforces it:
`skip_field` + absent ⇒ nothing, whatever the template says.
*Reach:* every render, all 173 components. *Fixes:* all 68 blank-field violations
at once, and any future one, without touching a template.
*Cost:* the schema is not in scope at execution — plumbing first. Can BREAK LIVE
PAGES: any of the 173 templates whose current output depends on a bare
`{{.field}}` rendering empty changes behaviour on the roll, and the ones most
likely to are the 20 already in violation. *And it is inert for ~90% of fields*,
which do not declare `on_missing` at all. Highest power, highest risk, worst
reach-per-unit-effort.

**B — Write-path refusal.** `store_generated_component` refuses a new/updated
template that leaves a declared `skip_field` ungated, or that gives a fact-shaped
`{{else}}` literal.
*Reach:* new and rewritten components only. *Fixes:* nothing that exists today.
*Cost:* low, contained, cannot break a live page — it only ever refuses a write.
Needs calibration against the real corpus or it will refuse legitimate work (the
lesson `component_write_guard.go`'s header is built around).
Closes the door at birth; leaves the 68 where they are.

**C — Promote the lint (CGV-029) from advisory to routine.** It already exists,
needs no roll, no schema declaration, and is the only defence that does not depend
on `on_missing` being declared. Extend it to also report the 68 ungated
`skip_field` fields, and run it on a schedule or from the existing sweep rather
than by hand.
*Reach:* the whole live library, both classes. *Fixes:* nothing automatically —
it reports. *Cost:* hours, not days. Cannot break anything.

**D — Do nothing further.** The fabrication class is closed and measured at 0
fleet-wide. The remaining 68 are blanks, not false claims. Accept that a third
incident, if it comes, is caught by the lint and fixed per-component as the first
two were.

### Recommended: **C now, B next, A only on evidence**

The sizing inverted my original instinct. **A is the architecturally satisfying
answer and the poorest value**: it is the only option that can break live pages, it
requires plumbing the schema into the renderer, and after all that it is inert for
nine fields in ten because the declaration it depends on is usually missing. **The
`on_missing` contract is too sparsely populated to be load-bearing**, and making it
load-bearing is a bigger cultural change (get 1,938 fields declared) than a
technical one.

C is cheap, safe, already built, and covers the class A cannot (undeclared
fields). B closes the door for new components at low risk. A becomes worth
revisiting **if** the declaration rate rises, or **if** a third *fabrication*
incident occurs despite the lint — which would be evidence the lint is the wrong
layer. Neither is true today.

**What I would want from you:** just the C/B/A/D call. If C, I would also want to
know whether "routine" means scheduled or wired into the existing discovery sweep
— the second is more useful and slightly more invasive.

## Sources

`bugs_open/140` · `bugs_open/111` / `d4731109d` ·
`docs024_key_docs_latest/bugfix_140_contact_info_fabrication/` (PLAN, NOTES —
the council dispositions are recorded there) ·
council correlation `40de12b0-36fa-4c06-82b4-995dc9098593`, seats `architecture`
and `reuse_agent` · CGV-029 in `docs026_concept_register/register/content-governance.md`


---

## DECISION RECORD — owner, 2026-08-03: "C now, B next"

**A (render-time gate) is not taken**, and the sizing above is the reason: ~90% of
fields declare no `on_missing`, so the gate would be inert for nine fields in ten
while being the only option that can break a live page.

### C — DONE, live, and proven on the real path

- The lint gained the **second finding class**: a field declaring `skip_field`
  that the template references but never gates. **68 fields across 20 components**
  today. It is reported but deliberately does **not** fail the exit code — a
  blank asserts nothing untrue, 68 of them predate the check, and a permanently
  red gate is one everybody learns to ignore. Only a FABRICATED FACT exits 1.
- It now runs **daily at 06:40 UTC** as the `component-fallback-check` CronJob,
  and writes a `doc_notes` row on **every** run, clean or not.
- **Deliberately a CronJob and NOT a discovery check.** The obvious home is
  `discovery_checks/` beside `check_placeholder_contact` — and that is exactly
  how the defect this lint exists to catch survived from the library's birth:
  that check's host, `quality-discovery-agent`, has raised **0** items in all
  history (`bugs_open/149` Group B/C). Measured 2026-08-02, the CronJob and
  `scheduled_tasks` substrates ARE driven; the discovery-agent one is not. Wiring
  a working check to a broken carrier is how you get inert-by-omission.
- `scheduled_tasks` was the other healthy substrate and was rejected for C: a row
  there fires a Kafka message at an agent type, so it needs a Go action and a
  chassis roll to carry a rule that already exists and works. That is the right
  shape for **B**, which has to live in Go anyway.
- **Proven, not assumed:** a manual `--from=cronjob` run succeeded, printed both
  classes, and wrote its own `doc_notes` row on the direct-Postgres path (the job
  cannot use `kubectl` — `ai-persona-app` has no pods/exec RBAC, a constraint that
  would otherwise have failed in a way that looks like a clean run).

### B — DONE 2026-08-03 (`87ea0a5e7`), inert until a chassis roll

`store_generated_component` now refuses a template that substitutes a business
fact for an absent datum. One more predicate in the existing `blockingIssues`
list, beside the five structural birth checks — so it adds no namespace, no wire
shape and no shared runtime contract, and it **cannot break a live page**: it only
ever refuses a write.

**Calibrated before it was written**, per `component_write_guard.go`'s standing
instruction, in two halves because one is not enough:

1. **0 findings across the FULL recorded write history** — every
   `component_versions` row plus every `content_components` row, **347 writes**,
   exported in batches and count-verified 347/347. It would have refused nothing
   the platform has ever written.
2. **That number cannot tell "correct" from "inert"** — and worse, by the time it
   was measured the corpus no longer held the motivating defect, because migration
   287 had already repaired it. **Your own fix silences your own detector.** The
   pre-fix template was recovered from the before-image that migration wrote to
   `migration_backups` and re-tested: **5 findings**, all four literals. Both
   halves are pinned as tests.

**The parity problem is real and is stated, not hidden.** The rule now has two
implementations — this Go gate and C's Python lint — which is the drift class the
rule itself detects. They are pinned to **one shared fixture**
(`platform/orchestration/actions/testdata/component_fallback_fixtures.json`, 24
cases, 10 must-refuse / 14 must-allow). The Go test reads it;
`check_placeholder_fallbacks.py --selftest` reads it; each refuses to pass if the
fixture stops exercising both arms, because a fixture of only must-allow cases
would pass against an implementation that never fires. The must-allow arm is the
larger half on purpose: refusing legitimate work is what gets a guard switched
off. The one-implementation alternatives were a Go CLI in the CronJob (rejected —
needs a clone and a compile in a job with uncertain module-proxy egress) or no
write-path gate at all.

**Wiring is proven, not assumed.** A helper with no callers looks exactly like a
finished refactor, so a test asserts the call site exists AND that its result
reaches `blockingIssues` rather than being computed and dropped. Mutation-proven:
replacing the append with `_ = issue` fails it; restoring passes.

**Known consequence, documented at the call site:** a component that already
fabricates cannot be repaired by regenerating it into another fabrication. That is
deliberate — the escape is a gated regeneration or a migration, which is how 140
itself was fixed. Currently vacuous: 0 of 173 active components fabricate.

**Owed:** pod-grep at the next roll (positive and negative control); read the
verdict for `19bee790-ea55-46eb-9f39-c985ecf8bd56` and act on a REVISE.

#### B — council APPROVED (11 seats), and the two objections answered with work

`19bee790-ea55-46eb-9f39-c985ecf8bd56` → **approved, 2 advisory objections, none
high-severity**, 4 abstained.

**bug_historian + guardian, both medium, and they asked the SAME checkable
question:** is `store_generated_component` really the only writer of
`html_template`? **It is not, and they were right to ask.** Census taken rather
than argued (`grep` for INSERT/UPDATE on `content_components`):

| writer | gate |
|---|---|
| `store_generated_component` (INSERT + UPDATE) | the **absolute** gate — birth |
| `update_component_html` (UPDATE) | the **comparative** gate — added answering this objection |
| `create_tool_component`, `deploy_tool_action` (INSERT) | ungated — tool components |
| `fix_component_template` ×2, `fix_harcoded_colours`, `fix_forced_text_colours`, `fix_nav_link_templates` (UPDATE) | ungated — narrow, non-LLM style repairs |
| core-manager admin handler (UPDATE) | ungated — human-driven, deliberately |

So **the write path alone does not close the door, and claiming otherwise would be
false.** What was added is the sound part: `update_component_html` shares
`componentRegressionIssues`, so the fabrication check went in there in its
**COMPARATIVE** form — refusing only a fabrication the replacement INTRODUCES.
That is not a stylistic choice: an absolute gate on a repair path would refuse
repairs to exactly the components most likely to need them, trapping an
already-fabricating component for ever. Pinned by a test that a repair to an
already-broken component passes, while a NEW fabrication in the same write is
refused and the message names only the new one.

**Re-calibrated, because `component_write_guard.go`'s header requires it of any
new check there:** the transition simulation over every consecutive
`component_versions` pair — **49 transitions, count-verified 49/49** — blocks
**0**. The remaining ungated writers are covered by the daily lint, which reads
the LIVE library and therefore sees a fabrication whichever writer introduced it.
**Gate where it is sound, report everywhere.**

**bug_historian, low: "the parity fixture's protection is only as good as both
selftests actually running."** Correct, and it was half-true when raised — the Go
side runs under `go test ./...`, the Python side ran nowhere. Fixed: the fixture
now ships in the CronJob's ConfigMap and the job runs `--selftest` **before**
`--report`, so a drifted rule is loud and no result is published measured by a
rule the write path no longer agrees with. Proven in-cluster with a manual run.

**guardian, low: confirm the CronJob's daily behaviour is unchanged.** It is —
the edit adds a flag and a preceding selftest line; the reporting path, schedule,
doc_notes row and exit-code semantics are untouched. Re-proved by a manual
`--from=cronjob` run after the change.

**A consistency fix made while answering these:** the shared fixture had become a
second copy in the deployment directory — the exact drift risk deliberately
avoided for the script. It is now the same arrangement: one real file in the
kustomize base, symlinked from `platform/orchestration/actions/testdata/`.

### B — the original notes, kept because they set the constraints


Refuse, at `store_generated_component`, a template that ships a **fact-shaped
`{{else}}` literal**. Notes for whoever picks it up:

- **It shares the classifier with C.** The fact-vs-label rule is the whole value
  and it currently exists only in Python. B must either port it to Go or call it;
  **two implementations of this rule is the drift class the rule itself detects**,
  so decide that deliberately rather than by default.
- **Calibrate before shipping**, against the live corpus, the way
  `component_write_guard.go`'s header describes. The lint's own history is the
  warning: it had a false positive (a builder attribution rendered link-or-text)
  and a false negative ("Weekdays 8am to 5pm") before it was fit, and only running
  four controls found the second.
- **It cannot break a live page** — it only ever refuses a write — which is what
  makes it the cheap half of the pair. It also fixes nothing that exists: the 68
  stay until someone gates them, and that is a separate, larger piece of work
  nobody has costed.

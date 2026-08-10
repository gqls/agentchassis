# NOTES — bug 122 contrast / ink slots

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-08-06 — picking the bug, and the three checks that nearly sent me elsewhere

**How 122 was chosen.** Ranked all 38 open bugs by reference-heat over the 42 session
transcripts touched in the last 4 hours, counting `bugs_open/NNN`-shaped references
(never the `NNN_HANDOFF_….md` filename — every session runs `ls bugs_open/`, so
filenames measure the floor). Coldest were `085` (28), `113` (38), `093` (39),
`146` (39), `114` (42), `203` (43), `122` (45). Hottest: `201` (506), `178` (331),
`149` (274).

I did **not** simply take the coldest:

- **085** — read it; both paths verified live, only a placement restoration and an
  induced empty case still owed, and it belongs to the brochure lane that owns the
  site. Not a fix task.
- **093** — read it; its own triage says *"093 is not a code task any more. It is
  blocked on `bugs_open/083`."* Correct to leave.
- **113** — fixed in code, awaiting a roll plus a fleet sweep decision the filing
  lane explicitly reserved.
- **122** — a live, user-visible, high-severity accessibility defect on public
  commercial sites, whose candidate 1 was explicitly released by its only active
  citer (`dartsonline_traffic`: *"candidate 1 still belongs to whoever takes it"*).

Then the symbol grep for a session already inside it —
`palette_specialised_slots|render_audit\.py|header-cta|fix_forced_text_colours|wcagContrastRatio|AuditPalette`
over the last 30 transcripts. One session at **45 hits**, which reads as ownership.
**I read the hits instead of tallying them**, per the recorded trap, and every one was
from **2026-07-28** — that session's own concept-register and memory writes about
building `cmd/contrastscan` and then deleting it. A closed historical session, not a
live competitor. `ls -d *122*` in `docs024_key_docs_latest/` returned nothing, so no
lane existed.

**MISSTEP 1 — I ran `who-owns.py` across 23 bugs and read the VERDICT line first.**
It said `OWNED or recently active` for essentially everything, exactly as the recorded
trap says it will at ~1,500 commits/week. Wasted output. The section that
discriminates is `=== likely OWNING workstream(s) ===`. Cost: nothing, because the
trap was already written down — which is the whole argument for writing them down.

**MISSTEP 2 — I wrote SQL against `site_components.is_active` without running `\d`
first.** `ERROR: column sc.is_active does not exist`. Then did the same thing again
minutes later with `content_components.css_styles`, and a third time with
`site_specs.resolved_composition`. Three round trips to a live cluster for a rule
CLAUDE.md states in four words ("Schema first"). The tell is that each query *reads*
perfectly — nothing about it looks like a guess. Logged in `WRONG_CALLS.md`; the
check is in the RUNBOOK.

**MISSTEP 3 — a census that returned 0 and would have returned 0 whatever was true.**
I counted `css_snippets` rows using `--color-primary` as an ink: **0**. I nearly
recorded "no component CSS uses primary as an ink". Induced a non-zero before
trusting the zero, per the marker discipline: of 21 `css_snippets` rows, **0 mention
`--color-primary` at all**. So the zero was real *and* the measurement was blind —
component CSS lives inside `content_components.html_template`, a completely different
surface, where the real answer is 17 of 18 layouts plus two shared components. Had I
not induced, I would have written a clean bill of health for the exact mechanism this
bug is about.

**MISSTEP 4 — a 19-site stylesheet census returned `403 Forbidden` on every row** and
I briefly read it as an origin/routing problem worth investigating. It is a
user-agent rejection; a browser UA fetches all 19. A failure that is uniform across
every subject is a property of the *method*, not of the subjects.

## 2026-08-06 — what the re-measurement found, and why the file needed correcting

The load-bearing result: **two of 122's three findings are fixed and its first fix
candidate has shipped.** The live `header-theme-chrome` template is
`color: var(--color-cta-text, var(--color-primary-text))`, and 0 of 19 stored header
chrome rows carry a hardcoded white CTA ink. robot-hands' white-on-white primary CTA
is gone. Only the vonc Gauntlet finding survives as filed, and that surface is owned
elsewhere.

If I had taken the file at face value I would have spent a roll re-fixing a shipped
fix. **The general form: a bug file's findings section is a dated measurement, and
this one is nine days old on a tree taking ~1,500 commits a week.**

What survives is a different class, in three sub-shapes (full working in
`PLAN_2026-08-06`):

- **A** — `--color-primary` used as an ink where the palette makes it a near-background
  dark. 17 of 18 layouts do this. `warnUnusablePrimary` already detects it and only
  warns; no derived slot offers "primary made legible as an ink".
- **B** — ai-agent-orchestration.com serves six `.H3` at **1.00:1**, heading equal to
  its own background. Worst instance on the fleet, in **no** bug file. Cause
  `[UNMEASURED]` — going to the diagnosis loop rather than into a guess.
- **C** — components hard-coding white over a themed mid-tone fill. `accent_text` was
  derived on 07-27 for exactly this and has **zero consumers** across all five
  surfaces that could name it. The dead-config LANDMINE, unfired, in the list it was
  written about.

**A correction to my own framing, caught while writing the plan.** I first sketched
the fix as an addition to `darkSchemeDerivations`. That would have been dead config —
the LANDMINE from this bug's own dartsonline round says a palette slot no layout
declares is never emitted, and I had already read it in the SessionStart output. Two
of the three failing sub-shape-C sites (gaswholesalers `#F4F1EB`, finetuning
`#F5F3EF`) are also **light**, so a dark-only derivation cannot reach them at all.
The renderer-owned `:root` block (`buildSectionDefaults` / `buildTokenAliases`
pattern) is the shape that survives both objections.

## 2026-08-06 (later) — the fresh build changed candidate 2 from a build task to one row

Chassis rolled to **v1.0.1257**. Pod-grepped `write_render_audit_findings` on
`agent-chassis-5b9fd84984-hqc5d`: **11**, invented control **0**, positive controls
`scanStoredStatClaims` **2** and `fillDarkSchemeSpecialisedSlots` **4**. So the
work-item drain (register VIZ-013), recorded as *"inert until an image roll AND the
config tail step lands"*, has its image half satisfied — and the config half is there
too: the live `render-audit-agent` row's steps are
`site → audit → write_findings → complete`.

**And nothing dispatches it.** 28 enabled `scheduled_tasks`, none targeting
`render-audit-agent`; `contrast_failure` items ever raised: **4**, all relojistas.com,
all 2026-08-04, all `complete` — one hand-run. 122's candidate 2 is therefore no
longer "wire the tool up"; it is a single `scheduled_tasks` insert. Same shape as
`083`/`093`/`115`: a mechanism made correct, then guarded behind something that never
runs.

**A check I nearly got wrong here.** I queried `orchestration_states` for
`owner_agent_type='render-audit-agent'`, got 0 rows, and started to write "has never
run". Terminal rows are reaped at ~24h, so 0 rows means *not in the last day*. The
question is answered by `scheduled_tasks` (no reaper) and by the work-item counts,
which is what I used instead.

## 2026-08-06 (later still) — both reviews fired; the sharpest measurement of the day

**`accent_text` is declared by 0 of 18 layouts** — measured directly, and it is the
finding that reshaped the fix:

```sql
SELECT count(*) FROM layouts WHERE css_template LIKE '%palette "accent_text"%';   -- 0
SELECT count(*) FROM layouts WHERE css_template LIKE '%palette "primary_text"%';  -- 18
SELECT count(*) FROM site_components WHERE rendered_html LIKE '%--color-accent-text:%'; -- 0
```

So the platform has derived a correct answer for sub-shape C since 2026-07-27 and it
has **never reached one stylesheet**, while its sibling `primary_text` (18 of 18) lands
everywhere. That halves the architecture surface of the fix: `--color-accent-text` is
not a new name to be argued over, it is an existing derived slot being made reachable.
Only `--color-primary-ink` is genuinely new, and I measured it unused on all five
surfaces before choosing it (0/0/0/0/0) rather than leaving that for a reviewer — the
2026-07-28 ruling is explicit that "no collision is possible" is a query, not an
argument.

**A design constraint I nearly missed.** dartsonline places the *same* ink on two
different grounds — the eyebrow on `background` (1.04) and the card link on the derived
`card_bg` (1.11). One variable cannot be right for two grounds unless it is right for
the worse of them, so `legibleInkFor` takes a *list* of grounds and requires the
candidate to clear AA against every one. My first sketch took a single ground and would
have shipped a value that fixed the eyebrow and left the card link failing — which
would have read as a working fix on the page I happened to test.

**Fired, both waiting:**

| what | correlation | for |
|---|---|---|
| council gate | `c4d9c841-3658-4742-85b5-961e062ecad2` | the fix plan (sub-shapes A + C) |
| 090 diagnosis | `5853ee07-a49c-4571-8ea0-3eb660e43dfd` (run) / `2f3d2cc0-197c-46ff-aac7-bd5e77ea782e` (intake) | sub-shape B, the six invisible headings |

Queue at time of writing: council at `review_editquality / EXECUTING_STEP`; diagnosis
at `diagnosing` with two bundles written.

> **A trap that briefly confused me reading those timestamps.** The diagnosis bundles
> are stamped `10:10:57` and `10:12:00` — *earlier* than when I fired the trigger, by
> my clock. They are UTC; this machine is BST. Comparing a BST wall-clock against a
> UTC DB timestamp makes a completed thing look like it never started, which is the
> same trap 122's own sibling files record in the other direction (it makes a live fix
> look un-shipped). **State the zone or convert.**

**Why the diagnosis loop for sub-shape B rather than just reading the code.** I could
have followed the alias chain myself. The reason not to is that this repo's diagnosis
section was *rewritten after* a thread with full context filed a confident structural
claim built from greps whose functions it had never opened, and the loop refuted it in
9.5 minutes. Sub-shape B has the exact profile that section names: two independent
mechanisms that both *appear* to be in place, and a resolved value that contradicts
both. That is a cause living somewhere other than the symptom, and "obvious" is
explicitly not the gate.

## 2026-08-06 (evening) — the council REVISEd me, and it was right: my sub-shape C was wrong

Round 1 verdict on `c4d9c841-3658-4742-85b5-961e062ecad2`: **REVISE**, gated by the
`editquality` seat. Its objection, verbatim in substance: the plan "makes accent_text
reachable but ships no edit that actually consumes it for any of the cited failures".

**It was right, and answering it properly refuted my own framing.** I had described
sub-shape C as *"a component hard-coding an ink over a themed fill"*. I then read the
three rules — which I had never opened, having inferred the shape from the audit's
output — and **not one of them hard-codes anything**:

```
finetuning .csg-cta-btn   (case-studies-grid)
    background: var(--color-accent, var(--color-secondary))  -> #C8873A
    color:      var(--color-primary-text, #fff)              -> #ffffff   = 3.01:1
```

The `#fff` fallback never fires: `--color-primary-text` **is** defined, as `#ffffff`,
and it is *correct for its own slot* (primary is `#1A1A2E`). The fill is **accent**.
The component names the wrong ink slot. A grep for hard-coded whites finds nothing here.

And the other two are the **opposite direction** again:

```
gaswholesalers .A   -> the LAYOUT's base rule  a { color: var(--color-accent); }  2.22:1 x6
gamesdesign .stats-eyebrow (system-stats) -> color: var(--color-accent, #7dd3fc)  1.44:1
```

Those use accent **as** an ink. `--color-accent-text` — the ink that goes ON an accent
fill — is the wrong repair and would have changed nothing on either site. So round 1's
plan could not have fixed 7 of the 9 failures it claimed. **The seat caught a real
hole, not a presentation problem.**

The correction is that a palette colour needs **both** directions named:
`--color-<x>-text` (ink ON an x fill) and `--color-<x>-ink` (x made legible AS an ink).
Round 2 shipped three variables instead of two and was **APPROVED**.

> **MISSTEP 5, and it is the same one as MISSTEP 3.** I described three CSS rules
> from the audit's rendered output without opening the templates that produce them.
> The audit tells you the computed colour; it cannot tell you which declaration chose
> it. I wrote "hard-coding" into a plan, a bug file and a council submission on that
> inference. The check is one query per selector and it is in the RUNBOOK now.

## 2026-08-06 (evening) — a fourth shape, found by accident, deliberately not fixed here

Chasing gamesdesign's other 8 failures found `rgba(255,255,255,0.7)` in **no**
stylesheet. It comes from `system-stats`'s own inline `<style>`:

```css
.system-stats-section { --section-text-muted: rgba(255,255,255,0.7); }
```

— a component redefining a token the renderer emits under a comment reading
*"Themes MUST NOT declare --section-* defaults; the renderer owns this."* The
component's scoped selector beats the renderer's `body` block, so the
contrast-checked value loses to a literal.

**47 of 173** active unforked components do this; **32** with a raw rgb/rgba literal.
~24 of the fleet's 109 failures. Filed as `bugs_open/212` rather than folded in — it
is an unenforced contract, not a missing variable, and bundling it is what the
guardian seat vetoes.

## 2026-08-06 (evening) — sub-shape B: two diagnosis runs, no verdict, and run 1 was unanswerable

Run `5853ee07` came back **UNVERIFIABLE** — *"Diagnosis NOT confirmed (stopped:
iteration-cap)"*, five bundles, no cause.

Then I went and measured, and found **the symptom I gave it was built on a false
premise**. `PLAN_2026-08-06` §2B says ai-agent-orchestration.com's stylesheet
*"does define `--color-heading: var(--color-text)`"*. It defines `--color-heading`
**zero times**. Headings never consult it — the rule is
`h1..h6 { color: var(--section-heading, var(--color-primary)) }`. The loop spent five
iterations reading `tokenAliases` and `darkSchemeDerivations`, two mechanisms that do
not participate in this failure, because I sent it there.

What is actually true, measured:

- the served stylesheet is **missing the renderer's step-11 compatibility-alias block
  entirely**, and ends at the close of step 10's output. **4 of 4** other sampled sites
  have it.
- so `--hero-ink` is undefined; `hero`'s `--section-heading: var(--hero-ink)` is
  therefore **guaranteed-invalid**; so the fallback `var(--color-primary)` applies;
  and `--color-primary` `#0D1117` is **byte-identical to `--color-surface`**.
- ruled out: staleness. `buildTokenAliases` landed 2026-07-06, the pages deployed
  2026-08-06.

Refiled as run `750e162e` with the corrected symptom — **also capped, also no verdict.**
Filed as `bugs_open/211` with the mechanism marked MEASURED and the cause marked
UNMEASURED, because I have not established *why* the block is missing and will not guess.

> **A `090` symptom naming the wrong mechanism returns UNVERIFIABLE, not REFUTED.**
> An iteration-cap stop reads like "hard bug" when it can mean "wrong question". That
> is a genuinely new failure mode for me — the corrected section in CLAUDE.md is about
> the loop catching *my* wrong claim, and this is the reverse: my wrong claim wasting
> the loop. Run 2 capped too, so the loop also has a real gap here.

## 2026-08-06 (evening) — the baseline I banked could not have verified my own fix

`BASELINE_2026-08-06_render_audit.txt` as first written covered **10 sites, 82
failures**. The plan quotes 15 sites and 109. The five missing were dartsonline,
robot-hands, vonc, relojistas, vetcomparison — **and dartsonline and robot-hands are
the two sub-shape A sites the entire fix targets.**

So the artefact banked specifically to make the fix measurable omitted every site the
fix was for, and nothing about the file said so — it looks complete, it is headed
"BASELINE", and the total line reads `10 page(s): 82 contrast failure(s)`. Completed
and appended; 15 sites / 109 now, which reconciles with the plan.

> **The check: a baseline is only a baseline for the rows it contains.** Assert the
> subject list against the thing you are about to change, not against the total. I
> would have "verified" a fix for dartsonline against a file with no dartsonline row.

## 2026-08-06 (evening) — writing the tests, and a fixture that could not have passed

Three mutations, each producing a DISTINCT failure, so none is a guard in series:

| mutation | fails | and only |
|---|---|---|
| move the ink call before `buildTokenAliases` | `InkCompanionsComeAfterTokenAliases` | that one |
| `for _, g := range grounds[:1]` | `LegibleInkFor_TwoGroundsDisagree` | that one |
| delete the `source:unchanged` branch | `LegibleInkFor_AlreadyLegibleIsLeftExactlyAlone` | that one |

> **MISSTEP 6 — my first two-grounds fixture was arithmetically unsatisfiable.** I used
> grounds `#101010` and `#E9E9E9`. AA against both is impossible: the darker demands
> relative luminance ≥ 0.200, the lighter ≤ 0.140. Every candidate correctly fell
> through to the achromatic fallback, so **the test failed while the code was right**.
> A trap no value can escape does not test preference — it tests the fallback. Rebuilt
> with two dark grounds, like dartsonline's real ones, where a satisfying colour
> exists and the CHOICE is what is under test.

Also: the package would not build in the working tree — another session's
`diagnose_persist_fix_plan_action.go` edit is missing an `agenterrors` import. Built and
tested against `git archive HEAD` + my four files instead, per the recorded practice.
The pre-commit pattern check then caught my test file not being gofmt-clean, which
would have failed the build gate. Both are the shared-tree tax, and both were caught by
machinery rather than by me.

## 2026-08-06 (evening) — two council objections that were measurement errors of mine

Round 2 was APPROVED with three advisory objections. Two were fair and cheap:

- **`reuse_agent`:** my "0 of 28 enabled `scheduled_tasks` target render-audit-agent"
  **filtered on `enabled=true`**, and a DISABLED row would have been invisible. Re-asked
  without the filter: **46 rows total, 29 enabled, 0 targeting render-audit-agent either
  way.** The claim survived — but it was luck, not method, and the seat was pointing at
  a recorded landmine. (Note "28" had already drifted to 29 in two hours.)
- **`guardian`:** blast radius stated only for the failing sites. Real fleet count:
  `tool-list` 6 placements / 4 sites, `system-stats` 5 / 4, `case-studies-grid` 4 / 3,
  `image-hover-card-grid` 1 / 1 — **16 placements**. Modest, and now stated.

## 2026-08-07 — the engine is live; `212`'s premise and its fix ranking are both wrong

**Deploy proven at the pod, not at the tag.** v1.0.1262 was rolled by another lane
(the `201` thread) and carries VIZ-014. Both replicas, one exec each:

| symbol | count | role |
|---|---|---|
| `buildLegibleInkDefaults` | 4 | the new emitter |
| `legibleInkFor` | 3 | " |
| `worstRatioAgainst` | 2 | " |
| `fillDarkSchemeSpecialisedSlots` | 4 | positive control — proves the pipeline |
| `zzzInventedControlXyz` | 0 | negative control — proves the grep discriminates |

So step 1 of the handoff is done, and it was **not** done by us. The image carries no
provenance; the symbols are the only evidence, which is why both controls are in the
table. Migrations 324/325 are still unwritten — and **`324` is now TAKEN** by another
session (`docs/agent_docs/sql_for_agents/324_asset_deployer_passes_asset_id.sql`,
untracked in the tree). The handoff's own warning — *"a number is not yours because you
named a file"* — fired within 24 hours. Pick a fresh number at write time, not now.

**Then `bugs_open/212`, which I picked up next, and which I filed yesterday.** Three
things in it are wrong, and all three were checkable when I wrote it.

> **MISSTEP 7 — I ranked four fix candidates for 212 without putting a number on any of
> them, and the two I ranked highest make the motivating case no better or worse.**
> The renderer's own contrast-checked value on gamesdesign's `.system-stats-section`
> ground (`rgb(13,191,214)`) is **`#e2e2e2` = 1.71:1**, against the component literal's
> **1.72:1**; the muted slot **regresses 1.72 → 1.46**. So candidate 2 ("emit at a
> specificity components cannot beat") is a very slightly worse repaint, and candidate 3
> (`var(--section-text, <literal>)`) resolves to the same value, so it is candidate 2
> with extra steps. I had written *"candidate 2 is the only class fix and also the only
> one that can break something that currently works"* — it breaks the very case the file
> is about. **What would have caught it: five minutes of arithmetic on a value that was
> already sitting in the served stylesheet I had open.** The 1.72:1 row of my own script
> reproduces the browser-measured 1.72:1 from the day before, which is what licenses the
> other four rows — they are counterfactual and no browser can measure them.

> **MISSTEP 8 — "an unenforced contract" was wrong; the contract is enforced, and the
> enforcement closed the item.** gamesdesign's defect was *detected* by the design
> audit on 2026-08-03, described correctly ("--color-primary as #00bcd4 (cyan)… making
> white text nearly illegible"), given a correct `acceptance_test`, routed to a live
> fixer, and stamped **`complete` 3m17s later** with nothing written. I filed a bug
> asking which of four repairs to build, when the repair already exists
> (`fix_forced_text_colours_action.go` classifies what a template paints and rewrites
> `--section-*` to the on-colour family — `system-stats` matches its `paintPaletteBand`
> regex) and the real question was why it never ran. **What would have caught it: asking
> the work-item queue what it already knew about the site.** CLAUDE.md tells you to check
> the queue *before dispatching*; I read that as being about collision with other
> sessions, and it is also a source of diagnosis. Now `bugs_open/213`.

> **MISSTEP 9 — I nearly filed 213 as an instance of RFC_017's fail-open policy.**
> RFC_017 names `hardcoded_section_colors` explicitly as one of its seven inheriting
> verifiers, so the fit looked exact. It is not that bug: the verifier **did not error**
> and **was not wrong**. It answered producer A's question correctly on an item filed by
> producer B. Reading its source looking for a defect finds a well-written function with
> an honest doc comment. A named prior bug that matches on symbol and on symptom can
> still be the wrong mechanism — the discriminator was one column, `spec->>'audit_source'`.

**The measurement that reframed 213.** Splitting the route by producer rather than by
status: **7 of the 9 `complete` items are design-audit (producer B), all seven carrying
an `acceptance_test` nothing reads — and every item that ever failed to close or is still
open is producer A's, 6 of 6.** A producer whose items never fail is the finding; the
individual false-complete is just the instance. Disconfirmable, and it could easily have
come out mixed.

**`buildSectionDefaults` emits nothing at all unless something is dark**
(`color_util.go:185-187`), and its surface variant covers a **hardcoded five-class list**
that `.system-stats-section` is not in and cannot join. Served stylesheets agree:
gamesdesign 1 block, vonc 1, **idea.uk 0**. That confirms 212's trap 4, which was a guess
at filing time, at the source rather than from an absence.

**090, run 3 for this lane: `b6ab22d6-e49c-4b55-a9d9-dd026532a595` — UNVERIFIABLE again**,
and again by iteration cap: three `bundle` artifacts, no verdict artifact, no
`metadata->>'decision'`. **This time it was not the stale code index** —
`symbols_unreadable` was 1 on iteration 1 and **0** on iterations 2 and 3, and the bundle
shrank to 1,943 chars by the last pass. It read the code fine and ran out of iterations.
Its hypothesis, though, independently reached my §8.1/§8.2 conclusion before it stopped:

> *"adding .system-stats-section to buildSectionDefaults' enumeration would not change
> what ships… The low-contrast risk is not a coverage gap in the enumeration; it is that
> the component's hardcoded near-white text assumes --color-primary is always dark enough
> to read against, an assumption nothing in the render path (warnUnusablePrimary checks
> primary-vs-background contrast, not primary-as-a-section-background-vs-hardcoded-white)
> validates."*

Corroboration, not proof — it is a hypothesis the loop never got to test. But it names
`warnUnusablePrimary`'s exact blind spot, which is the remedy edit the council's
`editquality` seat told us closes no failure. Consistent from two directions.

**Three consecutive UNVERIFIABLE verdicts in this lane** (`5853ee07`, `750e162e`,
`b6ab22d6`) — the first from a wrong question, the last two iteration-capped on
questions that were sound enough for the loop to form the right hypothesis. That pattern
is worth someone's attention on the loop itself, not on our symptoms. Run 4
(`84c3da66-06c0-41a5-94dc-21fbf71260f0`, the 213 mechanism) was still `diagnosing` at
handoff; **record its verdict here and in `bugs_open/213` §9 when it lands, including if
it is REFUTED.**

**Unrelated, noticed in the diagnosis bundle and not chased:** `agent_error_log` is
carrying a steady flood of *"workflow completed but its result could not be delivered to
the parent (failed_transient): message validation failed"* across `page-rerender`,
`feed-ingester`, `page-build-handler`, `build-dispatch-loop` and others — roughly one
every few seconds on 2026-08-07 morning. Not ours, not investigated, possibly adjacent to
`bugs_open/207`. Flagging it because it is the sort of thing that makes an unrelated
canary look broken.

## 2026-08-08 — run 4's verdict landed, and three of yesterday's figures had already gone stale

**`090` run 4 (`84c3da66-06c0-41a5-94dc-21fbf71260f0`, the 213 mechanism): `complete` at
08:48:02Z, five `bundle` artifacts, no `decision` on any — UNVERIFIABLE, iteration-capped.**
Same ending as run 3. It did **not** refute `bugs_open/213`: it never reached a verdict, so
213's root cause still rests on the timestamp evidence in its §3, which never depended on
this run. Recorded in `bugs_open/213` §9 and in the handoff §5.

That makes **four consecutive UNVERIFIABLE runs in this lane**. Worth stating precisely,
because the standing lesson ("UNVERIFIABLE means the question was wrong") does not fit the
last three: run 3's final bundle contained a *correct* hypothesis about
`warnUnusablePrimary`'s blind spot, and run 4's last bundle was still echoing my symptom
back unrefined. One wrong question and three that ran out of iterations is a statement
about the loop's budget, not about our symptoms. **Not filed** — it is a claim about the
diagnosis loop and filing it honestly needs the cross-lane run history, not just ours.

**Three figures in the handoff I wrote yesterday were stale within 24 hours**, which is the
more useful finding:

| figure | 08-07 | 08-08 |
|---|---|---|
| chassis image | v1.0.1262 | **v1.0.1264** (two more rolls, neither ours) |
| next free migration number | 324 | **335** — 324 *and* 325 both taken, 334 exists untracked |
| 090 run 4 | `diagnosing` | `complete`, UNVERIFIABLE |

I re-proved VIZ-014 on v1.0.1264 rather than carrying yesterday's table forward: identical
counts on both replicas, both controls holding. **The table surviving two rolls it was not
written for is luck, not evidence** — a rebuild from a HEAD that happened to contain our
commit. The handoff now says so, because the next roll could just as easily predate
something.

> **MISSTEP 10 — I wrote a migration number into a handoff as though reserving it, twice.**
> 08-06's handoff said 324; by 08-07 it was taken. 08-07's said "re-check at write time" but
> still left 324 as the anchor; by 08-08 both 324 and 325 were gone and the ledger was at
> 334. **A number in a handoff reads as a reservation no matter how it is caveated.** The
> fix that stuck: give the *command* that finds the number, and state that the file has been
> wrong about it twice — a reader will run a command and will not re-derive a caveat.

**One thing I have flagged but not measured: the 08-06 baseline is ageing.** Other lanes
ship to these same 15 sites continuously — the 08-07/08-08 log alone carries a 23-page
voice rollout and a placeholder-scan validator. A page re-rendered for someone else's
reason carries every change since it last rendered. So the banked before-state may no
longer be the before-state. Marked `[UNMEASURED]` in the handoff §6 with the instruction to
re-run and diff the baseline *before* grading the migrations, rather than quietly trusting
a file that was complete two days ago.

## 2026-08-08 (later) — writing migration 338, and what measuring the live rows did to the approved plan

Chassis rolled again to **v1.0.1266**; VIZ-014 re-proved on both replicas, identical
counts, controls holding. Third roll in three days, none of them ours.

**Wrote `338_components_opt_into_legible_ink_slots.sql` — committed `5e77607cf`, NOT
applied.** The pre-flight against the live rows changed it substantially, and the change
is the interesting part.

The approved sketch (submission edit 7) used a bare global `replace()` per component,
gated on a **row** count. But `replace()` is global **within the string**, and a row-count
gate is structurally blind to that. Occurrences vs intended targets, measured today:

| component | needle occurrences | intended | the extras |
|---|---|---|---|
| `case-studies-grid` | 2 | 2 | — sketch correct |
| `system-stats` | 5 | 1 | 2 backgrounds, 1 outline, 1 other ink |
| `image-hover-card-grid` | 2 | 1 | 1 outline |
| `tool-list` | 6 | 2 | **2 backgrounds**, 2 hover-fallbacks |

> **MISSTEP 11 — the approved plan would have repainted two backgrounds with an ink
> colour, and the gate it specified could not have caught it.** `tool-list` has
> `.tl-card-icon { background: var(--color-primary); }` and
> `.tl-cta-btn { background: var(--color-primary); }`. Repointing those to
> `--color-primary-ink` is the *exact inversion* this bug is about — an ink used as a fill.
> The council approved this, and so did I when I wrote it: five seats read the sketch and
> none of us asked how many times the needle occurs. **The check that would have caught it
> is one query and it is not about the fix at all — it is about the DATA the fix runs
> against.** A needle gate proves your string is present; it says nothing about how many
> other rules contain it. Every needle in 338 is now rule-scoped and each block carries a
> negative control asserting the non-targets survived.

**The layouts half did not exist as specified either.** The sketch's
`a { color: var(--color-accent); }` one-liner matches **0 of 18** layouts. Four of the five
named layouts use a multi-line rule (`a {\n  color: var(--color-accent);`) and
`tool-portal-light` uses a different single-line form — two separate blocks now. Worth
noting *how* this was missed: the five layout names in the plan are correct, and they are
the five with the most bare accent uses (16/13/12/9/8), so the plan's *conclusion* was
right and only its *needle* was invented. A correct list of targets makes a wrong needle
look verified.

**Deliberately not done, and it is the next thing:** none of the `RAISE`s have been
induced. A verify block that cannot fail is the trap 338's own header warns about, and
writing careful-looking assertions is not evidence they fire. That, plus the apply,
propagation enqueue and served-page verification, is §2 of the handoff.

## 2026-08-08 (late) — 338 APPLIED. And the near-miss on the way there was worse than the migration.

Chassis rolled again to **v1.0.1269**; VIZ-014 re-proved, both replicas, controls holding.
Fourth roll in three days, none of them ours.

**Migration 338 is applied and recorded** — 2026-08-08 22:12:55Z. Post-state matches the
dry run exactly, including both negative controls: `tool-list` gained 2 `--color-primary-ink`
and **kept both `background: var(--color-primary)` rules**; `system-stats` gained 1
`--color-accent-ink` and **kept both accent backgrounds**. All 5 layouts carry the new token.

**The dry run earned its keep before that, by failing.** It caught two bugs in my own file:
the `system-stats` negative control counted the bare needle and expected it to drop, but the
replacement NESTS the needle, so the count is unchanged by design — the replace was right and
my assertion was wrong. That same nesting made all six blocks non-idempotent. Then three
induced RAISEs proved the guards fire, the important one being the control that stops an ink
replacement reaching `.tl-card-icon`/`.tl-cta-btn`.

> **MISSTEP 12 — I audited the payload and shipped the delivery unexamined, and it applied
> four other threads' migrations.** The scoped apply was handed over as
> `MIGRATIONS_DIR=… ./run-migrations.sh --apply` on a line long enough to wrap. Entered as
> two lines, `VAR=value` is an unexported shell assignment the child never sees, so the
> runner used its default directory — **98 pending files** — and applied 198, 203, 204 and
> 207 before halting on a syntax error in 208. `204` wrote 10 live `products.content_data`
> rows on robot-hands; `198` created `gauntlet_rounds` in `clients_db` when migration 276's
> own guard says that table belongs on the ISLAND. **338 itself never ran** — the run died
> before reaching it.
>
> Everything I had verified was real and thorough and protected nothing, because the risk was
> never in the SQL. **The blast radius of what I checked was four rows; the blast radius of
> what I did not check was ~100 other threads' migrations, and the only reason it stopped at
> four is that an unrelated file happened to have a syntax error.** That is luck.
>
> Two checks, both four characters or one glance: **`env VAR=x cmd`** (a single command word,
> so splitting it errors instead of silently changing behaviour), and **read the runner's own
> `Pending (N)` line before `--apply`** — it asserts its scope on stdout before touching
> anything. I had seen `Pending (1)` in my own shell and treated it as a property of the
> command *text* rather than of the *invocation*, which is exactly what came apart.
>
> Sharpest part: I had read the landmine that says "`--apply` takes EVERY pending file —
> scope the dir", and obeyed it correctly in my own shell. Knowing a trap does not protect a
> **different surface** — here the handover rather than the command. Recorded in
> `LANDMINES.md` (with the `Pending (N)` check) and `WRONG_CALLS.md`.

The four applied migrations are **left in place**: forward-only, they are other threads'
files, their own guards passed, and whether they stand is the owner's call rather than this
lane's. Flagged to the owner explicitly rather than quietly reverted.

**Where the work stands.** The source is fixed; **no visitor sees anything yet**. The
propagation re-render is the whole remaining job and it is §2b of the handoff: 16 component
placements across 8 sites, plus a wider stylesheet re-render for 14 `css_themes` on the five
changed layouts. Two things I did NOT resolve and have marked as such: the enqueue mechanism
(ran out of budget on schema archaeology, not on the decision), and **whether the layouts
change even reaches gaswholesalers.com — the site with 6 of the 12 expected closures, and
the one site absent from the theme join.** If it does not, §6's expected-12 is wrong.

## 2026-08-09 — propagation is BLOCKED, and the blocker is bigger than this bug

Chassis at **v1.0.1270** (fifth roll, none ours); VIZ-014 re-proved, controls holding.
`enforceLayoutScheme` also greps 2, which matters below.

**Resolved yesterday's [UNMEASURED] flag, positively.** The layouts half *does* reach
gaswholesalers.com. I asked the served stylesheet instead of the schema — it carries
`a {\n  color: var(--color-accent);` exactly once. Cheaper and more direct than the
`css_themes.layout_id` join I had been grinding through, and it answers the question the
schema only implies. `--color-accent-ink` appears zero times there, which is expected: the
file predates both the engine and 338. §6's expected-12 stands.

**Then the real finding, which stops the lane.**

A CSS re-render is **unavoidable** — the new `--color-*-ink` / `--color-accent-text`
variables are defined only by the engine's step 12, which runs during a stylesheet render.
Until `styles.css` regenerates, every reference 338 added resolves to its fallback and
renders exactly as today. So the component half is inert too, not just the layouts half.

And the only live agent holding `render_css_from_spec` is `webdesign-agent`, whose step graph
forces `analyze_design` (an `execute_llm_prompt`) before `generate_css`. There is no entry
that skips it. Which runs straight into the recorded colour-churn landmine: `analyze_design`
invents a fresh palette per run unless the site carries a structured
`design_intent.palette.reference_values` block.

**0 of the 12 affected sites carry that pin** [MEASURED]. Eight have
`content_data ? 'color_scheme'`, which the check-side fix accepts — but that only suppresses
*spurious dispatch*; it does nothing about the LLM re-inventing on a real run.
`enforceLayoutScheme` is live and catches a background contradicting `layouts.scheme`, so the
worst 2026-07-17 outcome (light bg onto a dark site) cannot recur — but it does not constrain
accent, text or heading, which is most of what this lane measures.

> **The thing worth carrying forward: a fix can be complete, verified, reviewed and applied,
> and still have no delivery path.** 338 is correct — dry-run, induced guards, negative
> controls, applied, verified at the rows. None of that was ever the hard part. The hard part
> is that making a visitor see it requires running an LLM design pass across 12 live sites
> that are not this lane's to repaint. **I checked the payload, the delivery command, and the
> migration ledger, and never once asked "what actually re-renders a stylesheet?" until after
> the change was live.** That question was answerable on day one and would have reordered the
> whole lane — the pins should have been the *first* migration, not a discovery made after
> the source was already changed.
>
> Generalised: **"how does this reach production?" is a question about the LAST mile, and I
> keep answering it about the first.** Pod-proving the engine, gating the migration and
> scoping the apply are all upstream checks. The delivery path had a live landmine filed
> against it 23 days ago, in this repo, and I read that landmine only when I got to the step
> that needed it.

**Left explicitly undone rather than fired.** Three options are costed in handoff §2b — pin
then dispatch (recommended for this lane), build a non-LLM CSS render path (architecture
scope, recommended as a separate item), or leave it inert. The churn risk lands on sites this
lane does not own, so it wants an explicit yes rather than a session's judgement.

## 2026-08-09 (re-check, owner-prompted) — one of the four blocker claims was measured at the wrong store

Re-verified all four claims in the morning's BLOCKED write-up. Three survive, and are now
*proven* rather than inferred: the LLM pass is unavoidable (read the conditionals' branch
configs, not just `next_step` — every entry routes through `analyze_design`);
`webdesign-agent` is the sole holder of `render_css_from_spec` (text-wide over live
definitions, not step-shaped); `buildLegibleInkDefaults` has exactly one call site.

> **MISSTEP 13 — "0 of 12 pinned" was FALSE; it is 9 of 12, and the falsifier was in the
> memory's own citation.** I measured `sites.content_data->design_intent…` — the wrong
> store. The proven pin (`robot_hands/SQL_2026-07-17_r1b_…`) writes a **`site_specs`** row,
> `aspect='design_intent'`, superseding the previous one; `webdesign-agent`'s own
> `read_site_specs` step is what consumes it. **The contradiction was visible in my own
> output and I did not stop for it: robot-hands showed `pinned = false`, when the entire
> reason I knew a pin pattern existed is that robot-hands was pinned, provenly, on
> 2026-07-17.** A census whose one known-true row reads false is not a census — it is a
> wrong query answering confidently. What caught it: the owner asking for a re-check; the
> cheap check that would have: reading the cited pattern SQL before writing the measurement
> query — the file names the table in its first UPDATE.

Consequences, all pointing the same way: the blocker shrinks from "12 pins before anything
moves" to **3** (`ai-agent-orchestration.com`, `finetuning.uk`, `gaswholesalers.com`), and
the 9 pins were mostly written by `domain-research-classifier` in the normal pipeline, so a
pin is the platform's default posture, not an intervention. Handoff §2b corrected in place,
visibly. Recommendation unchanged in shape — pin the 3, then dispatch all 12 — but the cost
argument I gave the owner overstated the work fourfold.

## 2026-08-09 (canary review, owner-prompted) — the canary WORKED and the forecast was WRONG

Owner asked to look over the canary again before the batch. The review held on drift and
delivery, and caught two things the completion statuses could not show.

**What held.** Full-file diff of gamesdesign's stylesheet (not just colour slots): exactly
the intended 12 lines — the `a{}` repoint and the `:root` ink block. Zero palette drift,
typography untouched. Both re-rendered pages verified at their SERVED URLs: `index` carries
1 accent-ink + 2 primary-ink references, and the surviving bare `.stat-suffix` form is 338's
negative control, now visible in production. The advisory pin held on this run.

> **MISSTEP 14 — the expected-close figure was wrong BEFORE any of this ran, and I only
> found out by measuring the canary's "success".** The approved plan counted `.stats-eyebrow`
> closures on gamesdesign and vonc. `buildLegibleInkDefaults` computes the ink companions
> against `pageGrounds := {background, surface}` — and the eyebrow sits on the
> component-painted PRIMARY fill. On gamesdesign accent is 12.46:1 on the page ground, so
> `legibleInkFor` correctly returned it unchanged (`--color-accent-ink: #00e5ff` = accent),
> and the served, re-rendered eyebrow measures **1.44:1 on its real ground — byte-identical
> to the baseline failure**. The fix I shipped, reviewed and verified cannot close the
> failure it was partly named for. Expected close: **10, not 12** (vonc same mechanism,
> [PREDICTED] until its render lands). This is VIZ-014 inheriting `bugs_open/212` §8's
> two-ground blindness — the instance is recorded there, and it is the second time this
> lane has found "the renderer's answer is computed on a ground the element is not standing
> on". The check that would have caught it at plan time: for every expected closure, name
> the GROUND the failing element sits on and confirm the mechanism actually measures that
> ground. Five council seats and four sessions did not ask it.

**Two smaller catches, both cheap, both now in the recipes:**
- My `page_rerender` spec set `filename: 'tools-index.html'`; the page is served at
  `/tools/index.html`. Harmless — the handler resolves the deploy path from `pages.url` —
  but I only know that because the 2,856-byte fetch smelled like a 404 stub and was one.
  Verify at `pages.url`, never at a filename you constructed.
- My batch guard blocked 7 of 10 sites by counting ANY dispatchable build item as a
  collision. The collision surface is `styles.css`, so the guard belongs on
  `handler_agent='webdesign-agent'` in imminent/in-flight states; parked (`unresolved`) and
  untriaged (`detected`) rows act on nothing. dartsonline's 69 "collisions" were asset and
  page items. Corrected and re-dispatched: **all 11 live sites queued** (canary complete;
  lendzy.co.uk excluded — Cloudflare 522, origin down, no before-state to grade against).

Verification protocol for the batch, already banked: before-state sha+accent for all 12 in
`scratchpad/before_css/`; per site when its item completes — full diff (drift), the three
ink slots present, then per-selector grading against `BASELINE_2026-08-06` for the 10
reachable closures.

## 2026-08-10 (afternoon) — pages verified, re-audit graded 10/10, a same-day recurrence fixed, edit 8 live

**All 12 page items completed** (10:53–11:23Z, faster than the morning's hours-not-minutes
forecast) and **all 12 verified at the served artefact**: every page carries the new ink
tokens (`--color-accent-ink` / `--color-primary-ink` / `--color-accent-text`), fetched at
`pages.url` with a browser UA. Three pages (aao index, finetuning index + tools) have
Last-Modified stamps from 08-09, BEFORE our items ran — they already carried the tokens
because other lanes re-rendered them after mig 338 applied (08-08 22:12Z), so our rerender
was byte-identical and the upload was skipped. Not a false complete; a no-op re-proof.
vonc index was re-rendered AGAIN at 13:54Z by another lane after our 10:54Z run — tokens
still present.

**Re-audit run and banked**: `AFTER_2026-08-10_render_audit.txt`, 15 pages, 112 total vs
baseline 109 — but graded per selector as required, and the totals mislead exactly as
predicted. **All 10 expected closures delivered**:

| site | expected | result |
|---|---|---|
| gaswholesalers | 6 (.A orange-on-white) | **all 6 gone**, contrast=0 |
| robot-hands | 2 (.tl-card-link 1.07, .tl-eyebrow 1.14) | **both gone** (only the over-image approximate .H2 remains, discounted per runbook) |
| finetuning | 2 (.csg-filter-btn active, .csg-cta-btn 3.01) | **both gone** — plus `.cta-btn cta-btn-primary` 1.00 white-on-white closed (unattributed bonus; likely a layout accent-ink or another lane) and all 5 broken case-study images fixed (another lane) |
| dartsonline | 1 (.image-hover-card-grid__eyebrow 1.04) | **gone by removal** — but see the recurrence below |

**§6's [PREDICTED] for vonc is now [MEASURED]: confirmed.** `.stats-eyebrow` 1.63:1 and the
whole gauntlet/stats band byte-identical to baseline after TWO same-day re-renders — the
component-painted ground is unreachable by the shipped engine, exactly as the 08-09
correction said. gamesdesign's eyebrow likewise still 1.44:1. Both remain open under
`bugs_open/212` §8.

**The three drifted advisory-pin slots introduced no new failures** (idea.uk's growth is
more instances of its standing muted-on-cream class, incl. the new `.tl-intro`; dartsonline's
eyebrow ground now measures the drifted #0F1219 and the ratio moved 1.04→1.14 — still
hopeless, closed by 368 below). New failures NOT ours, for whoever owns those lanes:
aao `.H2` 1.12:1 light-on-light on a new 'Eight departments' section (aao total 30→33);
vetcomparison 0→3 (grey #6B7C85 at 4.10–4.14, marginal); idea.uk 14→21 same class.

**RECURRENCE, found and closed same-day: dartsonline swapped `image-hover-card-grid` for
`info-card-grid` and reintroduced sub-shape A ×6** — `.info-card-grid__card-link` 1.06:1 ×5
on card-bg, `.info-card-grid__eyebrow` 1.14:1 on background. Classified at the template:
info-card-grid hardcodes NOTHING — it inks `color: var(--color-primary)` on the PAGE grounds
(`--color-background` / `--color-card-bg, --color-surface`), so this is engine-reachable
(NOT 212's class), literally 338 §4 tool-list's shape (eyebrow + card-link). dartsonline's
served CSS already carries `--color-primary-ink: #F0F2F7`; predicted 1.14→16.73:1 and
1.06→13.78:1 (the current values reproduce the audit's failing ratios exactly — independent
corroboration). **Migration 368 written, dry-run rolled back, both RAISE guards induced,
applied + recorded 14:47Z**: 2 foreground uses wrapped, 4 non-targets asserted intact.
27 placements / 14 sites measured first: no-op where primary is legible (ink ≡ primary),
var() fallback where slots are absent. Re-render enqueued for dartsonline index
(`page_rerender_index_dartsonline_com_viz014icg_20260810`, 14:50Z, page_id set in spec AND
column per the LANDMINES entry). **brands NOT enqueued: its only page_component has NULL
`content_data` — fails the batch's own pre-check; left for its owning lane.**

> **MISSTEP 15 — my first placements query returned 0 rows for a component the served page
> demonstrably renders.** I encoded `sec->>'component_name'` / `sec->>'name'` — an OBJECT
> shape; `pages.sections` is an array of PLAIN STRINGS. The a-jsonb-path-read-cannot-see-
> the-shape landmine, hit while measuring. Caught because a zero against a visible
> counter-example cannot stand; `jsonb_typeof` + a 300-char peek settled it in one query.
> The 27-placement figure above is from the corrected query.

**Edit 8 DONE and LIVE-PROVEN: migration 369** (`site-render-audit-rotation`). Clones the
site-discovery-rotation mechanism: hourly tick, 7-day due window = weekly per site, one due
site per fire, skip mid-build sites, stamped no-op when none due; own concurrency group
`render-audit` cap 1, timeout 1800s < interval. Dry-run clean, both guards induced,
pre_query tested in a rolled-back txn (picked robot-hands). Applied + recorded ~14:53Z —
and **the scheduler fired it within a minute**: rotation stamped robot-hands.com 14:54:23Z,
orchestration `b30943e4` at step `audit` 14:54:24Z. That is the artefact-level proof, not
"enabled + a fresh tick" (the thunder-reaper trap). The dartsonline recurrence above is the
standing demonstration of why this row exists: the defect class returned within 2 days of a
component swap, and only a hand-run audit caught it.

Commits: mig 368 `0d9e555ec`, mig 369 `4b924895f`, both pathspec-scoped, both carrying the
lane's standing `Council-Reviewed: c4d9c841`.

Still open at session end: dartsonline item queued (verify at served URL + targeted re-audit
when it lands — inside the deploy window use the poll-with-cache-buster check from
LANDMINES, assert an added AND a removed string plus a control); robot-hands weekly audit
orchestration in flight (its write_findings will file any firm findings as `contrast_failure`
items — note `bugs_open/213`'s false-complete risk sits on THAT repair route, unfixed).

**Addendum, ~15:00Z: the first weekly sweep COMPLETED (~3 min) and filed 34 firm
`contrast_failure` items on robot-hands' INTERIOR pages** — the homepage-only baseline
never looked there. Among them: `info-card-grid__card-link` + `__eyebrow` on
`/selection-guide.html` (mig 368 closes these when that page next re-renders — a live
cross-check of both of today's changes), plus `A.cta-btn` on 10 pages, tool-page
buttons/legends/H2s, and eyebrow variants. All born `detected`, deduped on
`contrast_failure:<page>#<selector>`, routed to css-patch-agent — where `bugs_open/213`'s
false-complete defect is now load-bearing at scale: 34 real findings are about to flow
through a verifier that has stamped work complete without doing it. 213's priority just
went up.

**Addendum, 15:12Z: dartsonline re-rendered (complete 15:09:36Z, deployed 15:09:34Z) and
VERIFIED — the page measures CLEAN.** `python3 scripts/render_audit.py
https://dartsonline.com/` → `contrast=0 broken-img=0`. All six 368-target failures closed
(card-link 1.06 ×5, eyebrow 1.14). Served-page checks, cache-busted, per the LANDMINES
protocol: **added** `var(--color-primary-ink, var(--color-primary))` ×2; **removed** bare
`color: var(--color-primary);` inside the info-card-grid block → 0; **controls** the
`.info-card-grid__card-link:focus-visible` outline intact, and 3 bare
`color: var(--color-primary);` still present elsewhere on the page
(`.pricing-tier__price`, `.stat-highlight__number`, +1) — proof the edit was surgical and
not a global replace.

> Note on control B: I asserted "≥2 outline rules intact" and got 1, which read like a
> miss for a minute. The second is `.info-card-grid__arrow:focus-visible`, inside a
> carousel block that does not render on this page (`grep -c info-card-grid__arrow` on the
> served HTML = 0). **A template-level count is not a served-page count** — the non-target
> assertions belong where I actually put them (in the migration, against the row), and a
> served-page control has to be chosen from what the page actually renders.

**This is the inflection §8 was waiting for: the first page measuring clean.** SUMMARY
written (`SUMMARY_2026-08-10_contrast_ink_slots.md`).

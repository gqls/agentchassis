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

## 2026-08-10 (close of session) — two corrections against my own commits

> **MISSTEP 16 — my index commit `01d64d564` carried a SAME-FILE PASSENGER, and its
> message asserts something now demonstrably false.** Another lane's **IMP-053** row plus
> its rewrite of the header's count paragraph were sitting uncommitted in
> `000_concept_index.md` when I edited it; a pathspec commit takes the FILE, so both
> shipped under my message. Nothing is lost (forward-only; their work is committed and
> attributed here), but the message says *"1,811 immediately before, 1,812 after — VIZ-015
> added ONE and **nothing arrived concurrently in that window**"*. **The arithmetic is
> right and the conclusion is wrong: IMP-053 had already arrived — it was IN the 1,811 I
> counted, as untracked WIP.** A working-tree count cannot distinguish "committed" from
> "another session's uncommitted edit", so it can never support a claim about concurrent
> arrivals; only a count against `HEAD` can (`git show HEAD:<file> | grep -c …`). The
> same-file passenger is a known landmine and I checked for it on the FIRST docs commit
> (clean, verified line by line) and then did not repeat the check on the second.
>
> **Corrected claim:** VIZ-015 added one row; at least one other row (IMP-053) arrived in
> the working tree during this session and travelled in my commit.

**Number collision, flagged not fixed: there are now TWO migration 368s** — mine
(`368_info_card_grid_opts_into_legible_ink_slots.sql`, applied + recorded 14:47Z) and
another lane's `368_site_availability_driver_HOLD.sql` (uncommitted, `_HOLD`, unapplied).
**Not a functional break:** the runner lists by filename and `schema_migrations.filename`
is the primary key, so both are tracked independently, and `SIDECAR_RE` excludes an
UPPERCASE-suffixed file from `--apply` regardless. It is a convention break and a
readability trap. **I have not renumbered either file:** mine is applied and recorded (a
rename would orphan its `schema_migrations` row), and theirs is another session's WIP,
which is not mine to rename. Whoever picks up the availability driver should renumber it
before applying — it is the unapplied half, so it is the cheap one to move.

This is systemic today, not a one-off: **two different lanes have each written a `370`**
(`370_experience_planner_escalation_descriptions_catch_up_with_363.sql` and
`370_retire_update_page_status_notes_and_validation_issues_fields.sql`), both untracked.
The handoff's standing advice to "re-check the free migration number" is necessary and
**not sufficient** — I did check (`ls | grep -E '^3[3-9][0-9]'` showed 367 as the highest)
and 368 was genuinely free at that moment. The collision arrived in the window between
the check and the write. On this tree the number is not reservable; only the filename is
unique, which is the property the runner actually relies on.

## 2026-08-10 (post-handoff) — the cadence's own second run found a defect in the cadence

Checked the second rotation fire before closing out, because "0 findings" has more than one
cause. It picked loancalculator.co.uk, `COMPLETED`, `error` NULL, **0 items filed** — and
the reason turned out to be a *third* cause I had not listed: `skipped_locked: 2`, both firm
findings sitting in locked components, which VIZ-013 deliberately declines to file. **Zero
was correct on that run.**

But the same check surfaced a real one. The site has **27 deployed pages and the sweep
measured 25** — and nothing in the stored run says so. `truncated` is computed
(`request_render_audit_action.go:157`), warned (`:160`) and returned in `Metadata`
(`:251-259`), yet `collected_data->'render_audit'` holds exactly three keys —
`response`, `response_status`, `response_received_at` — enumerated with
`jsonb_object_keys`, **not** probed by path, precisely because a path read cannot see a
shape change underneath it.

Re-checked fire 1 rather than leaving it for the next reader (the bug file's own §6 told
that reader to do it; doing it myself was one query): **robot-hands.com is 31 pages, also
swept 25 — 6 never rendered.** So **both rotation runs in existence were truncated and
neither said so.**

**The instructive contrast, and the reason the bug is filed the way it is:** the drain one
step downstream hit *its* cap on the same run and was honest about it —
`findings_capped: true`, `findings_dropped: 111` of 171 firm. Same class of decision, same
run, opposite outcome. The fix candidate is not "invent a mechanism", it is "give the sweep
the parity the drain already has".

> **MISSTEP 17 — I wrote "filed 34 real findings" in the handoff, the register index row
> and this file, and the denominator was 171.** Every instance was literally true and all
> three read as "the sweep found 34 things". A filed count is a count that survived two
> caps and a dedup; it is not a measurement of the site. Corrected in all three places, and
> the bug file carries the correction too. The check that would have caught it at the time:
> **I read `inserted` and stopped, without reading the sibling keys in the same object** —
> `findings_capped` and `findings_dropped` were sitting right there in the JSON I had
> already fetched.

Filed as **`bugs_open/242`**. It asserts only what I measured; the mechanism (why
`Metadata` does not reach `collected_data`) is marked `[UNVERIFIED]` with the three checks
that would settle it, and §6 states plainly that I did not run `090` and why the trigger
does not fire for an observation-only file.

## 2026-08-11 — the owner's decision executed, and a verification method retired under me

**Spend baseline before enabling anything** (llm_call_log, calls/in/out per hour):
08:00 `8/40,076/13,228` · 09:00 `28/175,532/87,506` · 10:00 `134/514,812/361,053` ·
11:00 `57/397,350/80,514` · 12:00 `31/113,502/60,002`. Banked so the sweep can be judged
against a no-sweep busy hour rather than a feeling.

**Migration 389** — park 226 `contrast_failure` → `deferred`, then re-enable
`improvement-sweep` at 900s (from 180s). Dry-run rolled back clean, **both RAISE guards
induced** (double-run → "already parked"; pre-enabled row → "someone else moved it"),
applied 12:31Z, recorded. The sweep fired at 12:31:38.

**The park was measured into existence, not reasoned into it.** I nearly enabled the sweep
on its own. Reading its `pre_query` first showed a `< 50` guard on
`(triaged,detected)`+`pipeline='build'` — and our own overnight findings had pushed **five**
sites over it, including the two with the most re-renders (40 and 22). Enabling alone would
have skipped exactly the sites the owner wanted drained. Post-park, sites over the guard:
**5 → 1**, asserted inside the migration rather than claimed afterwards.

> **MISSTEP 19 — my engine proof yesterday used a method that was retired today.**
> CLAUDE.md now forbids `strings /app/agent-chassis | grep -c` ("three confidently wrong
> readings in one day") in favour of the binary's own provenance stamp, and records that
> `v1.0.1284` — the tag I proved against — **shipped three revisions** (`bugs_open/249`).
> My yesterday reading was not wrong (re-proven today on v1.0.1286 by the sanctioned
> `grep -aq … /proc/1/exe` with both controls) but **the method was, and the handoff I
> wrote enshrines it in §1 as the loop to re-run.** Corrected in the new handoff.
> The general lesson is not about `strings`: **a verification recipe copied forward in a
> handoff is a claim about the present that nobody re-checks**, because its whole purpose
> is to look like the settled part of the document. When the standing docs change under a
> lane, the lane's own recipes are what go stale first.
> Second-order: the sanctioned method did not work either — the provenance line had
> already rotated out of the chassis logs (retention is brutally short), so the documented
> primary path failed and the documented fallback carried it. Worth knowing before you
> budget on the log line.

**Contributed into `bugs_open/213` rather than working it** — `who-owns.py` returns OWNED
and its fix is already live and pod-proven, so competing would have been the wrong move.
What that lane lacked was behavioural proof, and I could supply the shape of it: all 26
`_verification` rows predate the roll (latest 08-09), so their `out_of_scope = 0` means
**idle, not blind**; the verifier can refuse (4 `defect_persists`) and reports its own
errors (2 `error`), neither of which is the false-complete shape. And enabling the sweep is
precisely the traffic that will exercise their gate — so their proof arrives as a
side-effect of this lane's decision, which they should know before reading today's 0.

## 2026-08-11 (evening) — the sweep ran 5h29m and was stopped; its own promotions re-locked the fleet

`improvement-sweep` disabled **18:00:39Z**. Ran 12:31:40Z → 18:00:39Z = **5h 29m**, ~22 fires
at 900s. One `UPDATE`, reversible.

**Cost — inside the band the handoff set, which turned out to be the wrong band.**

| window | calls/h | in_tok/h | out_tok/h |
|---|---|---|---|
| pre-sweep 08:00–12:00 | ~52 | ~248k | ~120k |
| sweep 13:00–17:00 | ~132 | ~806k | ~223k |

Calls stayed at the fleet's own busy-hour shape (93–184/h against the 10:00 baseline of 134),
so by the handoff's stated criterion — "a few hundred calls/hour is the mechanism working; a
jump into the thousands is not" — it **never tripped**. Input tokens were **3.2x** the
pre-sweep average, which that criterion does not look at. **A call count is not a spend:** each
sweep call is a full site pass, so the per-call cost is what moved.

**Progress, as the handoff framed it:** `page_rerender` `detected` **193 → 25**; `complete`
2,017 → **2,258** (+241). Read alone, that is the drain working.

> **MISSTEP 20 — two claims inherited from this lane's own handoff were wrong, and the second
> inverted the decision.**
>
> **(a) "the re-render drain is a *downstream* effect of it" — REFUTED.** Completions ran
> **49/h and 48/h at 10:00 and 11:00 on 08-11**, before the 12:31Z re-enable, and **85 / 110 /
> 41** across 08-10 13:00–15:00 while the sweep was disabled. Sweep-era completions were
> 35 / 32 / 63 / 53 / 58 — *the same rate*. The drain is a **separate, always-on path at ~50/h**;
> the sweep's marginal contribution is discovery and promotion, **not execution**. The whole
> cost/benefit was therefore mis-stated: stopping the sweep does not slow the drain at all.
> The check that settled it is one `GROUP BY` over **a window when the thing was OFF** — a
> downstream-dependency claim is cheaply falsifiable, and I carried it a day instead.
>
> **(b) The stop criterion watched the wrong axis.** `detected` 193 → 25 looks like a queue
> emptying; it is a queue **moving**. The sweep **files** work as well as promoting it:
> **526 new `page_rerender` rows created during the sweep hours (~105/h)** against ~48/h
> completed. Open work (`detected+triaged+unresolved`) went **~273 → 544**. The one number
> anyone was watching is the one that fell.

**The finding that actually justifies the stop: the guard counts `triaged`, so the sweep
re-locks the fleet with its own output.** `pre_query` skips a site with **≥50**
`(triaged,detected)` rows on `pipeline='build'` — promotion moves a row from one side of that
`IN` list to the other and **does not reduce the count**. Sites over the guard:

- **5** before migration 389's park
- **1** immediately after it (the migration's own post-check)
- **8** now (17:58Z) — *worse than the state the park was performed to fix*

The bulk of each locked site's backlog is now triaged re-renders the sweep itself promoted
(leopardessconsulting 55 of 95, finetuning 56 of 69, gamesdesign 41 of 79). **The park bought
eligibility and the sweep spent it**, then overdrew — 226 contrast findings are still held in
`deferred` funding a guard allowance that has since been consumed.

**The park itself held**, and that was the risk worth controlling: all 226 `contrast_failure`
rows still `deferred`, every one stamped `parked_by='migration_389'`, and **zero**
`contrast_failure` rows outside the park — so no new contrast finding was filed and promoted
into `bugs_open/213`'s path while the sweep ran.

**Where this leaves the lane:** 544 open re-render items draining at ~50/h with arrivals now
stopped ⇒ **~11h to clear**, after which the guard releases the locked sites by itself. That
outcome needs *not* doing, not doing.

## 2026-08-12 — the overnight prediction held, and the park's REASON changed

Measured 12:39Z on **v1.0.1290** (both chassis replicas, started 2026-08-11 21:53Z; the
`build provenance` startup line had already rotated out of `--tail=3000`, as CLAUDE.md warns).

**The stop was the right call and the prediction was disconfirmable, which is why it is worth
recording.** `page_rerender`: `triaged` **446 → 0**, `complete` 2,261 → **2,803 (+542)**,
`failed` 66 → **15**, `detected` 12. Guard census: **0 of 22 sites locked out**, from 8 at the
moment of stopping. Sweep still `enabled=false`, untouched by any other session; 226
`contrast_failure` rows still `deferred`. Had the drain in fact depended on the sweep — the
claim MISSTEP 20(a) refuted — `triaged` would have sat at 446 all night. It went to zero.

**The 213 re-run I promised them.** `_verification` population **26 → 44**, and rows now
**postdate** the roll (latest `completed_at` 2026-08-11 18:57Z vs 08-09 the day before), so
the "nothing gradeable has completed" explanation for `out_of_scope = 0` is dead. But
`out_of_scope` is **still 0**, and the reason is not their bug:

- **14 `dark_section_audit` items**, all `audit_source='design-audit'`, all created
  12:49–17:56Z (the sweep's own output) and completed 12:56–21:35Z — **all post-roll** — and
  **0 of 14 carry a `_verification` key at all.** Same window, `hardcoded_section_colors` is
  **9 of 9**.
- Mechanism read, not inferred: `verifyBeforeComplete` resolves via
  `checks.GetVerifier(itemType)` (`complete_work_item_verification.go:70`) and an
  **unregistered** type completes untouched, documented at `:16`. The `out_of_scope` branch
  (`:112`) needs a *registered* verifier that declines. `dark_section_audit` is at no
  `RegisterVerifier*` call site (12 types are; it is not one).

> **This is the "a gate's 0 findings has TWO causes" trap, and I recorded the reassuring
> cause yesterday.** Yesterday's contribution read `out_of_scope = 0` as **idle, not blind**
> and said so in their bug file. That was right *for the evidence then* — every row predated
> the roll — but I stated it as a property of the gate rather than of the traffic, and the
> caveat I attached ("re-run tomorrow") is the only reason it did not harden into a false
> reassurance. **A verdict about a silent mechanism expires when the traffic changes**, and
> the check is the one I nearly skipped because the answer was already written down.

**The park STAYS, and my previous reason for it was wrong.** I had told 213's lane "they
unpark when you close this". Withdrawn: **`contrast_failure` has no registered verifier
either** — filed at `write_render_audit_findings_action.go:258`, absent from every
`RegisterVerifier*` site. So unparking the 226 mints rows that complete **ungraded by
construction**, which 213 closing does not change. The trigger is now "`contrast_failure` has
a verifier, or someone rules it needs none" — and 213 is **no longer blocking this lane**,
which they have been told so they do not hold their closure for my 226.

## 2026-08-12 — migration 395: the quality rotation is ENABLED and will do NOTHING until 2026-08-16

Owner decision, after reading the cost and lock-out evidence: **"enable the discovery
rotations, slowly"** — not raise the sweep's guard, not make its discovery step conditional.
Executed as migration `395`, scoped to **one** rotation (`quality`; `design` and
`completeness` stay off), `enabled=true`, cadence `3600s → 10800s`.

**Read the pre_query before reasoning about the rate — the interval is NOT the cost dial.**
`site-discovery-rotation-quality` is `LIMIT 1` per fire against a **7-day** due window, and it
**stamps `last_selected_at = now()` in the same statement that selects the site**. So the work
rate is bounded by the rotation period: 22 active sites / 7 days ≈ **3.1 site passes per day**
however often it polls. The interval only sizes the initial ramp (3600s → whole fleet in ~22h;
10800s → ~8/day, ~3-day ramp). That stamp-on-selection is also the fix for the sweep's
`ORDER BY sites.updated_at ASC` starvation (IMP-010) — nothing the sweep does advances its own
sort key, whereas this cannot re-pick a site it just examined.

> **The post-check earned its place: it printed `0 site(s) due for the initial ramp`.**
> **All 22 sites are ALREADY stamped** for `quality-discovery-agent`, from 2026-08-09 09:49Z →
> 2026-08-10 16:39Z — the window when this rotation was briefly enabled before the owner
> switched it off on 08-10. Stamps are per `agent_type`, so the still-enabled
> `availability` rotation's stamps are irrelevant. **First site comes due 2026-08-16 09:49Z**
> (robot-hands.com), the rest over roughly the following 31 hours at the cadence they were
> originally stamped.
>
> So the honest state is: **enabled, correct, and inert for four more days.** Had I not put a
> due-count in the post-check, `enabled=true` plus an advancing `last_triggered_at` would have
> read as "running" for four days — a fire with no due site dispatches nothing, costs nothing,
> and logs like success. This is the `enabled` + fresh tick ≠ ever RUN shape, and the thing
> that caught it was asserting a **row count I could be wrong about** rather than asserting
> the UPDATE succeeded.

**I considered backdating one stamp to price a single pass now, and decided against it.** The
reasoning is worth recording because the first instinct was wrong: the 08-16 ramp delivers the
*same* single-pass measurement for free, four days later, so a probe buys lead time only — and
buying lead time with an unrequested LLM spend, four days after a 3.2x cost surprise, is not a
call to make on the owner's behalf. What waiting genuinely costs is **visibility**, not data:
the risk is that on 08-16 the ramp starts with nobody watching and 22 passes run over ~3 days.
That is a docs problem, so it is written down here, in the handoff, and at the foot of `395`
rather than solved by spending.

**📅 2026-08-16 — the dated action this lane now owns.** When the ramp starts, take the spend
per hour (calls AND tokens — calls/hour did not discriminate the sweep's 3.2x) against the
~248k input tokens/h no-rotation baseline, and confirm the rotation is actually fair
(every site acquires a stamp; none re-selected inside 7 days). Both queries, plus the
one-UPDATE stop and the statement to add the other two rotations, are at the foot of
`docs/agent_docs/sql_for_agents/395_enable_quality_discovery_rotation_slow_ramp.sql`.

**⚠ This restores DETECTION, not repair — deliberately.** SCH-025 findings are born
`detected` and `improvement-sweep` (the only triage carrier) stays disabled, so findings will
accumulate unconsumed. Two consequences, written down now rather than discovered later:
`detected` rows **count toward the sweep's ≥50 guard**, so anyone re-enabling it must re-read
the guard census knowing this rotation has been raising the numbers it reads; and this lane's
226 parked `contrast_failure` rows are unaffected, because `deferred` is not `detected`.

**MISSTEP 21 — I nearly ran `run-migrations.sh --apply`.** The dry run listed **12 pending
files belonging to other threads**, including `324_asset_deployer_passes_asset_id.sql`, whose
own guard refuses without a marker because *"on an older binary this config deploys the WRONG
asset bytes (bugs_open/155)"*, plus two probe-inconclusive files and one containing its own
`ROLLBACK`. `--apply` takes **every** pending file in order — my one-line config change would
have shipped all of them. Correct path, which the runner supports directly: apply the single
file by hand (`psql -v ON_ERROR_STOP=1 -f -`), then `--record-only <file> --note "<why>"`.
The dry run is not a formality on a tree this many sessions share; it is the only thing
standing between a scoped intent and an unscoped action.

---

## 2026-08-12 (afternoon) — the verifier fork was costed, and it collapses differently than §3 expected

Task taken from `HANDOFF_2026-08-12_continue_here.md` §3: *"cost (1) against that standing
objection first … Do not start by writing the verifier."* Done. The objection holds, it is
wider than the handoff knew, and it kills **all three** listed options — but a **fourth**
option exists that the handoff did not list, and the estate has already built its plumbing.

### (a) The standing objection is at `:199–201`, not `:171` — and there are THREE of them

`verifier_coverage_test.go:171` today is `needs_component_regeneration`, an unrelated entry.
The objection actually lives in three consecutive gap entries [MEASURED — read the file]:

- `:199` `image_url_404` — *"deliberately NOT a verifier candidate: verification would add an
  outbound HTTP call to the completion path"*
- `:200` `backend_entry_orphaned` — *"same reason as image_url_404"*
- `:201` `asset_reference_404` — same, **plus the constructive half**: *"It does not NEED one:
  the check retracts its own findings through `CheckResult.Resolved` on a positive 2xx/3xx
  re-observation, which is the same information a verifier would fetch, taken on the discovery
  path where the probe is already precedented."*

⚠ **The stale `:171` citation is also in `LANDMINES.md`** (the `assets`-table entry), which is
the system of record. Both point at the right file and the wrong line. Fixing that is a
separate small task; noted here so the next reader does not conclude the objection was moved
or withdrawn.

### (b) The objection kills options (1), (2) AND (3)

The handoff expected (3) to survive. It does not: **every option that computes contrast needs
an outbound fetch on the completion path**, because a source-level read cannot settle contrast
(our own `var(--x,#fallback)` landmine). (3) is narrower in *predicate*, not in *mechanism*, so
it draws the identical objection. The fork as written has no survivor.

### (c) `contrast_failure` is ALREADY an on-record decision — and its reason is unsound

Not a gap awaiting a ruling. `verifier_coverage_test.go:156` classifies it, and the guard
`TestEveryItemTypeIsVerifiedOrAnAcknowledgedGap` refuses to let a type be both verified and
listed — so its presence there is machine-checked proof it has no verifier [MEASURED: `go test`
passes; 12 verified types, `contrast_failure` absent]. Its stated reason:

> *"verification needs a browser — the dedup key `contrast_failure:<page>#<selector>` plus the
> NEXT render audit is the verifier, and the two-strike rule catches a persistent pairing"*

Three independent findings say that reason does not hold:

1. **It is the argument the owner overturned.** `complete_work_item_verification.go:26–34`:
   the old fail-open policy was justified by *"discovery re-detection + two-strike is the
   backstop"*, and on 2026-08-08 that was **measured and failed** — the error path had fired
   twice, both items completed at `attempt_count` 0, and *"five days later no re-detection had
   followed although the detector's own predicate still matched. The backstop did not catch the
   only two cases it was ever asked to catch."* That measurement is what produced RFC_017's
   fail-closed flip. **Dates** [MEASURED, `git log -S`]: the `contrast_failure` reason was
   authored `f2a222964`, **2026-08-02**; the ruling that refuted its argument landed
   `1c5d9ceb5`, **2026-08-08**. The entry was never revisited. `[INFERRED]` — and marked as
   such — is the *transfer*: RFC_017 measured the argument on a different item type's error
   path, not on `contrast_failure`. The form is identical; the measurement is not ours.
2. **The producer has no retraction at all.** `grep -n "Resolved\|retract"
   write_render_audit_findings_action.go` → **zero hits** [MEASURED]. So what the entry calls
   "the de-facto verifier" is the **absence of a re-file**, never a positive observation. That
   is the precise inference the estate's own contract forbids, stated in the file the entry is
   leaning on — `check_asset_reference_404.go:~150`: *"inferring 'resolved' from absence is
   exactly what `CheckResult.Resolved`'s contract forbids."* `asset_reference_404` earns its
   exemption **because it retracts**; `contrast_failure` copies the exemption without the thing
   that justifies it.
3. **The mechanism has never run once.** [MEASURED, live DB]: 226 rows, **all** `deferred`,
   **all** `parked_by='migration_389'`, `max(attempt_count)=0`, created 08-10→08-11 across 16
   sites / 203 distinct keys. **Zero `contrast_failure` items have ever been dispatched,
   completed, or re-detected in the platform's whole history.** The two-strike backstop has
   never fired for this type; it is an untested design claim. Compare
   `hardcoded_section_colors` — the type that HAS a verifier — 9 complete / **9 carrying
   `_verification`**, plus 5 `unresolved` the verifier actually blocked.

**Disconfirmation check, because a zero is cheap:** the same query returns non-zero for three
sibling types in the same breath (`dark_section_audit` 14 complete, `hardcoded_section_colors`
9 complete, `asset_reference_404` 1 cancelled), so the instrument is not blind — the zero is
about `contrast_failure`, not about the query.

### (d) Option (4), which the handoff did not list: retract on the DISCOVERY path

This is `asset_reference_404`'s posture, done properly, and it is what the gap entry was
reaching for and got wrong. The render audit **already** runs a browser over every page weekly,
already outbound, already precedented. What is missing is only the closing half.

The plumbing exists and is shared, not new: **`resolveWorkItems`**
(`work_items_common.go:249–300`) is the runner-side half of `checks.CheckResult.Resolved`
(RFC_010, owner ruling 2026-08-02) — *"One implementation, called from one place, deliberately
… It never infers. It closes what it is told to close, and only that."* It validates as a
**refusal** (empty `ItemType`/`Reason`, or neither `ItemKey` nor `AllOfType` → error, not a
guess) and it will not close what the current run just raised (`batch_id IS DISTINCT FROM`).
It is **unexported but in package `actions`** — the same package as
`write_render_audit_findings_action.go` — so it is directly callable, and that file already
opens a `tx` at `:336` and loops filings at `:343`.

**The one real blocker, and it is small and precedented.** Retraction must be scoped to pages
the run actually measured. The payload carries `Summary{Pages, PagesTotal, Truncated}` —
**counts, not identities** (`:143-147`). The audited set cannot be derived from
`payload.Contrast`, because **a page that is now clean produces zero findings and is therefore
indistinguishable from a page never audited** — and that is exactly the repaired case we need
to retract. So option (4) needs `pages_audited: []string` (URL identities) in the adapter's
summary: the same one-step-downstream parity fix `bugs_open/242` just made for the counts,
extended from *how many* to *which*.

⚠ **`bugs_open/242` is NOT "open, unstarted"** as `HANDOFF_2026-08-12` §5 row 2 and
`HANDOFF_2026-08-11` NEXT item 2 both state. It was **done 2026-08-11, live on v1.0.1288,
council APPROVED, behaviourally proven** by the `bugfix_242_render_audit_truncation` lane
(forced-truncation run on loancalculator, cap 5 vs 26 pages). Our handoff was written today and
is wrong about it. **And the relationship is the opposite of what §5 implies:** 242 is not a
lower-priority sibling, it is the **enabling precondition** — before it, a capped run was
indistinguishable from a complete one, so no retraction could ever have been scoped safely.
That half is already paid for.

### What this changes

The park's trigger, as `HANDOFF_2026-08-12` §4 states it, is *"a `contrast_failure` verifier
exists, or someone rules the type needs none"*. Both branches are now better specified:

- "Needs none" **cannot be ruled on the current reason** — that reason is refuted above. It
  could still be ruled on the strength of option (4), but only once (4) exists.
- The cheapest sound path is **(4), not a completion-time verifier**: no completion-path probe,
  no second implementation of a measurement we already have, reuses the shared closer, and it
  converts "absence of a re-file" into a positive observation — which is the only version the
  `Resolved` contract permits.

Unparking still comes last, and its order is now determined: **build the retraction, then
unpark.** With retraction live, a `contrast_failure` that completes ungraded is *corrected* by
the next weekly audit rather than merely unnoticed — the backstop the gap entry claimed we
already had.

**MISSTEP 22 — I nearly took the handoff's `:171` citation as read.** It names a real objection
and I could have quoted it from the handoff without opening the file. Opening it changed the
answer twice over: the objection is wider than described (three entries, not one) and its
newest member does not merely object — it states the alternative, which is the recommendation
above. A `file:line` in a handoff is a pointer, not a quotation, and line numbers drift.

---

## 2026-08-12 (after the v1.0.1291 roll) — the retraction design, and four hazards found while specifying it

Carried on from the costing above. Design is now **fully settled**; implementation not started.
Handoff: `HANDOFF_2026-08-12b_continue_here.md` §3. Recording the evidence here so the design is
auditable rather than merely asserted.

**State re-verified post-roll** [MEASURED]: chassis `v1.0.1291`, both replicas, started 14:55Z.
226 items still `deferred`, `max(attempt_count)=0`, zero completions all-history. The roll
changed nothing in this lane — worth stating, because "a fresh build" invites the assumption
that it did.

### The two constraints that fix the design's shape

**(A) The requester's own result is DESTROYED by the await, so the data must travel in the
adapter's response.** `request_render_audit_action.go` already computes `urls_audited` metadata.
It never survives: *"an awaiting step's own result never survives the park
(persistAwaitingStateWithRetry loads fresh state and keeps only the awaited-request entries —
RFC_012 addendum 2, owner-ruled option B)"*. This is why `bugs_open/242` echoed
`pages_total`/`truncated` **through the request into the response** instead of fixing it at the
requester — a detour that looks redundant until you know this. Anyone who tries the "obvious"
route will lose a session to it.

**(B) The adapter already models the failure side, and it is a warning.**
`RenderAuditResult.Unreachable []string` exists because *"it would let a dead page pass as
clean"*. So "audited" must mean **successfully measured**, never "requested". Retraction scoped
to the requested list would commit precisely the error that field was added to prevent.

### HAZARD 1 — retraction will close the PARKED items, and that changes what the park means

`workItemClosedStatuses` = `complete, verified, rejected, wont_fix, cancelled`
(`work_items_common.go:83-89`). **`deferred` is not in it** — nor are `unresolved`/`failed`,
which are retractable on purpose (owner ruling, Decision 2). So `resolveWorkItems` will close a
parked row. Shipping retraction therefore starts closing the 226 as each site's weekly audit
confirms them fixed, **with nobody unparking anything**.

I think that is the RIGHT behaviour — a parked ticket whose defect is gone should close, and it
leaves only the genuinely-still-broken remainder needing a fixer, which is a better outcome than
the unpark-everything plan. But it is a **change in the park's meaning** and must be ruled on
deliberately, not arrive as a side effect. Flagged for the council submission.

### HAZARD 2 — the closer's self-protection guard is INOPERATIVE for our rows

`resolveWorkItems` guards with `batch_id IS DISTINCT FROM $6` so a run cannot close what it just
raised. `NULL IS DISTINCT FROM <uuid>` is TRUE, so a NULL-batch row is **not** protected.

[MEASURED 2026-08-12, and the controls came out the other way, so the zero is real]:

| item_type | rows | with batch_id |
|---|---|---|
| `contrast_failure` | 226 | **0** |
| `empty_section` | 61 | 61 |
| `hardcoded_section_colors` | 15 | 15 |
| `asset_reference_404` | 1 | 1 |

The three controls are filed through the discovery-check runner, which sets it;
`contrast_failure` comes from the render-audit producer, and `insertWorkItem` never writes the
column at all. **Preferred remedy: populate `batch_id` on rows this producer files** — it
restores a guard the estate deliberately built rather than routing around it, and makes these
rows consistent with every other type. Leaving it NULL rests a *destructive* operation entirely
on one loop's set logic with no backstop.

> **Method note, because this nearly went out as an inference.** I first concluded "no batch_id"
> from `grep batch_id work_items_common.go` returning only the closer's own lines — a grep of one
> file, which is an argument about that file, not about the column. The query above is what makes
> it a finding, and the control types are what make the zero disconfirmable. A grep proves absence
> only for the spelling it searches, in the file it searches.

### HAZARD 3 — the audited set is NOT derivable from the findings

A repaired page contributes **zero** findings, so it is indistinguishable from a page never
visited — and that is exactly the case retraction exists to catch. This is the entire
justification for adding `pages_audited` rather than deriving it from `payload.Contrast` URLs.
Written down because "just derive it" is the obvious reviewer objection and it is wrong.

### HAZARD 4 — do not compare colours as strings

`fg`/`bg` are inconsistently formatted **within one row** (`"rgb(26, 31, 46)"` vs
`"rgb(15,18,24)"`). The design keys on selector+page and never compares colour strings, so it
sidesteps this. Any later refinement that compares recorded to computed colour must parse to
numeric triples first.

### One doc change the fix MUST carry

`verifier_coverage_test.go:156`'s exemption becomes **true for the first time** once retraction
ships — but its current wording describes the unsound absence-inference. It has to be rewritten
to the `asset_reference_404` formulation in the same commit. Shipping the mechanism and keeping
the false reason would leave the next reader inheriting the thing this session spent its time
refuting.

### MISSTEP 23 — my LANDMINES correction was swept, exactly as I predicted, and that was fine

I left `LANDMINES.md` uncommitted because another session had an in-flight entry in the same
file and I did not want their work under my message. Their commit `dbf74bc71` took my line with
it. **Verified at HEAD, not at the tree** (`git show HEAD:… | grep -c`) → present. Nothing lost;
forward-only held. Recording it because the lane's standing-trap list already carries this trap
and this is now its second occurrence in two days — the correct response is to expect it and
verify at HEAD, not to hold work back.

---

## 2026-08-12 (evening) — the retraction is BUILT and COMMITTED (`5639a1103`), and implementing it found a defect in this lane's OWN spec

Session task: the three edits of `HANDOFF_2026-08-12b` §3.2. All three done, plus tests, register
entry `WII-016`, and council submission `a43b63d6-da35-4136-9471-88ec6ace799a`. Committed by
pathspec, 8 files, scope block clean (no passengers this time).

### CORRECTION TO THIS LANE'S OWN HANDOFF — §3.2 step 1 would have closed live defects

`HANDOFF_2026-08-12b` §3.2 specified the still-failing set as:

> Build the set of pairings this run **observed still failing**: `workItemKey(...)` for every firm
> (`over_image=false`) contrast finding — i.e. exactly the keys already computed at `:266`.

**The clause after the dash contradicts the clause before it, and the dash is where the defect is.**
"Every firm contrast finding" is right. "Exactly the keys already computed at `:266`" is NOT the
same set, because `:266` is reached only by findings that survive TWO filters that sit above it in
the same loop:

- `if htmlCorpusContainsClass(lockedHTML, c.Class) { skippedLocked++; continue }` — the culprit
  class lives in a LOCKED component, so we deliberately do not file it;
- and after the loop, `if len(toFile) > maxItems { ... toFile = toFile[:maxItems] }` — the
  `max_items` cap (default 60), worst-ratio-first, dropping the remainder.

A finding removed by either was **measured, and is still failing**. It is simply one we cannot act
on. Building the retraction set from the filed items reads "not filed" as "fixed" and closes those
tickets — a false completion, which is the precise outcome the park of 226 exists to prevent. The
`undeployed_asset` items appended to `toFile` after the cap are a third reason the filed list is
the wrong source.

So the set is built from `payload.Contrast` directly, **before every filter that decides what to
FILE**. `over_image` approximations are counted as still-observed for the same reason: the
adapter's own header calls that backdrop unknown, so "I could not tell" is not a positive
observation of health, and a pairing that has gone from firm to approximate has not been shown
fixed. Erring here costs one ticket staying open a week; erring the other way closes a live defect.

**How it was caught:** by reading the function before editing it, per the lane's own standing trap
("a `file:line` in a handoff is a pointer, not a quotation" — the trap that fired on `:171`
yesterday, now fired on `:266` today). The citation was accurate about the LINE and wrong about
the SET. **Cheap check that would have caught it at authoring time:** ask what is between the top
of the loop and the cited line. Two `continue`s were.

### The four guards are proven by MUTATION, not by a green suite

A passing sqlmock test proves nothing about a negative — the estate's own rule. Each guard was
reverted to its plausible-but-wrong form, the suite re-run, and the corresponding test confirmed
RED; the file was then restored and `diff`ed byte-identical.

| mutation | test that must go RED | result |
|---|---|---|
| still-failing set built from the FILED items (the handoff's spec) | `MeasuredButUnfiledFindingsAreNotRetracted` | RED |
| audited-page scope removed (retract anything absent) | `UnreachablePageRetractsNothing` | RED |
| `over_image` treated as evidence of health | `OverImageReadingDoesNotRetract` | RED |
| parked counter made indiscriminate (`n > 0` alone) | `RetractsOnlyAbsentPairingsOnAuditedPages` | RED |

The first row is the load-bearing one: **the handoff's spec fails this lane's own test.**

### Figures re-measured rather than carried forward

Ran against `clients_db` before quoting the handoff's numbers, with sibling types as controls so
the check could have come out otherwise:

```
item_type                | rows | deferred | with_batch_id | max_attempts | completed
contrast_failure         |  226 |      226 |             0 |            0 |         0
empty_section            |   61 |        0 |            61 |            3 |        10
hardcoded_section_colors |   15 |        1 |            15 |            0 |         9
asset_reference_404      |    1 |        0 |             1 |            0 |         0
```

All of `HANDOFF_2026-08-12b` §1's figures hold. The `batch_id` zero is real and the controls are
what make it evidence rather than a broken query — hence the change also populates `batch_id`, so
`resolveWorkItems`' `batch_id IS DISTINCT FROM $6` guard can fire here for the first time.

### Consumers enumerated, not asserted (owner ruling 2026-07-29 §3)

Live agents whose config references `render_audit`: exactly **one**, `render-audit-agent`, steps
`ensure_site_record` / `request_render_audit` / `write_render_audit_findings` / `complete_workflow`
— all in this lane. The only other Go reader of the same collected-data field is
`execute_vision_prompt_action.go`, which resolves **only** `.renders` via `resolveVisionImageRefs`
to feed the design critic and never touches `.summary`, so a new summary key cannot affect it.

### Verified at HEAD, not at the working tree

The shared tree carries four other sessions' modified Go files, so a green build in it is evidence
of nothing. Extracted `git archive HEAD` into a clean directory: `go build ./...` clean, both
suites green. This matters more than usual today — a fresh chassis is being built from committed
HEAD while this was written, so the commit ships on someone else's roll.

### What is NOT done, stated so nobody reads the above as more than it is

- **Nothing has been retracted in production.** The mechanism is INERT until BOTH `agent-chassis`
  and `browser-runner-adapter` carry it. A chassis-only roll does nothing, by design.
- The council verdict is **unread** (`Council-Submitted:` trailer, not `Council-Reviewed:`).
- The 226 stay parked. Do NOT unpark them: the order is ship → watch one weekly audit → then
  unpark, and hazard (1) means a shipped retraction closes the genuinely-fixed subset on its own.
- **226 is still a FLOOR, not a census** — the audit was capped at 25 pages until `v1.0.1288`, so
  this closes tickets without discovering the ones that were never filed.

### Same evening — the roll killed the council round, and the two facts about that roll are independent

Round 1 (`a43b63d6`) stalled at `review_tooling_provenance`. Diagnosed rather than waited out, per
the written remedy in `LANDMINES.md` (*"A chassis roll KILLS an in-flight council"*): last
`updated_at` **19:13:18Z** against chassis pod `startTime` **19:13:54Z** / **19:14:16Z** — frozen
36 seconds before the first new pod. Twelve days after that entry was written, same arithmetic.
Resubmitted with `RESUBMIT_CORR=`, ~300s window already past.

**Why the trail id matters more than usual here, and it is a general point:** the commit was
ALREADY PUSHED carrying `Council-Submitted: a43b63d6`. A fresh correlation would have stranded
that trailer on a run that can never produce a verdict, and forward-only forbids the amend that
would fix it. `RESUBMIT_CORR` reuses the same `fix_correlation_id`
(`FIX_CORR="${2:-${RESUBMIT_CORR:-}}"` in `097`), so the trailer stays honest and `098` will credit
the commit when round 2 lands. **If you have already committed, resubmit on the trail, never
fresh.**

**The second fact, which I nearly conflated with the first: that roll did NOT ship this change.**
Both services came up on `7a1887e31`; my commit `5639a1103` sits **3 commits after** that build
point. Checked, with a control, rather than inferred:

```
git merge-base --is-ancestor 5639a1103 7a1887e31   -> false   (not in the build)
git merge-base --is-ancestor 7a1887e31  HEAD       -> true    (control: the check works)
```

Watching a roll happen tells you nothing about whether your commit is in it — the build was cut
before I committed, and the roll that killed my review is the same roll that missed my code. Two
independent facts about one event.

**One piece of good news, measured:** `agent-chassis` and `browser-runner-adapter` stamped the
SAME commit 29 seconds apart, so a fleet release rolls the adapter with the chassis. The handoff's
"roll both services" is **one release, not two** — which also means the version-skew branch I built
is unlikely to be exercised in practice, and stays as the guard for the case where it is.

### Council round 2: APPROVED — and the best objection found a gap in my evidence

Round 2 completed 20:01:45Z. **APPROVED, 5 advisory objections, none high-severity.** The health
fields matter as much as the decision and were checked rather than assumed: **13 reviewers, 4
abstained (relevance-gated), `unreadable: 0`, `gated_by_truncation: false`.** An approval where
seats were truncated or unreadable is not an approval on substance — this one is (relevant given
this lane's own note that the architecture seat's first three reviews were 2/3 truncated).

**`guardian`'s objection was right and I had missed it.** I enumerated CONSUMERS of the render-audit
response with care — and never enumerated PRODUCERS of `contrast_failure`. The seam's own landmine
is explicit that adopting retraction on a co-filed `item_type` silently closes the other producer's
findings; that is exactly why `check_undeployed_assets` was rejected as WII-009's first adopter,
and `undeployed_asset` is co-filed *by this very action*. So the risk was live and adjacent, not
theoretical. Measured after the fact, with the controls the census can give:

```
audit_source | source       | created_by         | rows | distinct_key_sets
(none)       | render-audit | render-audit-agent |  226 |                 1
```

One producer, one shape — plus a Go grep showing one file files the type. **The limit, stated:
this census sees producers that have FILED. A producer that has never fired is invisible to it**
(WII-015's own "not claimed", inherited here). Re-run before anything else files this type.

**`bug_historian`'s URL-shape objection produced the one code change:** the retraction compares a
key built from today's URL shape against one a PAST filing wrote, and nothing pinned which way a
mismatch fails. Added `ShorterPageDoesNotPrefixMatchALongerOne`, pinning the dangerous direction —
`/pricing` must never prefix-match a key belonging to `/pricing.html`. That is what the `#` inside
the prefix buys; removing it makes the test RED (fifth mutation proof this session).

**`reuse_agent` asked the Step-Zero question I had not shown:** why not extend
`revalidate_review_queue` / `reviewRevalidators`, which already closes parked items on fresh
positive evidence? Checked rather than argued, and decisive twice over: that action runs
**in-chassis, which has no Chromium**, so it cannot take a contrast measurement at all — the very
constraint that put this on the audit path; and its selector is `workItemRevalidatableStatuses` =
{`needs_human_review`, `unresolved`}, which excludes `deferred`, and widening a shared selector is
what that file's own comment warns against (it records `failed` being left out after measuring the
blast radius). Right call, undocumented reasoning — now documented.

Two accepted and NOT fixed: it is a symptom fix for one item_type while `GetVerifier` still
completes any unregistered type untouched (`bugs_open/213`'s call), and `architecture` flagged that
this is the **second** inline copy of "still-failing set before filters, retract via
resolveWorkItems" — **a third must extract a helper.** Both written into WII-016.

**Residual, unclosed:** concurrent render-audit runs on one site. `batch_id` guards a run against
its own rows only, so run B can retract a row run A filed seconds earlier if B did not re-measure
that pairing. Defensible, untested, recorded rather than solved.

**Not amending `5639a1103` to say `Council-Reviewed:`.** Forward-only forbids it, and `098` credits
the commit automatically from its `Council-Submitted:` trailer now the correlation is approved —
which is precisely the hole that trailer was invented to close.

---

## 2026-08-13 — LIVE on both services, and the first exercise is four days away by design

The 08-13 roll carries it. **Verified at the artefact with controls, and the verification itself hit
two traps worth more than the result.**

Both services are stamped `69612d692a4a07d61eea3f648e1152e0fd36fd0a`, and
`git merge-base --is-ancestor 5639a1103 69612d692` → **true** (controls: yesterday's build IS an
ancestor of it; HEAD is NOT). The chassis startup line had scrolled, so binary probe with both
controls: `69612d692` PRESENT, `7a1887e31` ABSENT.

**Trap 1 — my own documentation poisoned my own detector.** CLAUDE.md's recipe
(`logs | grep -m1 'build provenance'`) returned **1.1MB** on the chassis. Landmine entries ABOUT
build provenance are synced into `doc_notes`, injected into prompts, and logged — so the phrase now
occurs in ordinary log lines as documentation text. `grep -m1` stopped on one of those and handed me
something that was not the stamp. The CLAUDE.md fallback rule only covers the false NEGATIVE
("empty means not in range, never unstamped"); this is a false POSITIVE, which it does not cover.
Fix: match the log line's structure (`"caller":"…/main.go:N","msg":"build provenance","git_commit":"<40 hex>"`),
never the phrase. Filed as a landmine, since CLAUDE.md prescribes the command that misfires.

**Trap 2 — grepping the binary for MY OWN commit returns `absent`, and that is CORRECT.** Only the
BUILD's commit is stamped, never every ancestor. For a moment that read as "my change did not ship".
It is an ancestry question, which is a git query, not a grep. Also filed.

### It cannot fire until 2026-08-17, and that is the mechanism working

`site-render-audit-rotation`: enabled, hourly, `last_triggered_at` 13:27Z — and **0 sites due.** Its
`pre_query` window is **7 days** per site and the active fleet was stamped 08-10 by the very audits
that produced our 226. So the hourly tick dispatches nothing and advances `last_triggered_at`
exactly like a working fire — `enabled` + a fresh tick ≠ ever ran, and there are **no**
`render-audit-agent` orchestrations in 10 hours to confirm it. Same shape as the
`site-discovery-rotation-quality` trap this lane already documented, arrived at from a different
direction. **Do not force a run to watch it work.**

### The gate debug_historian asked for turned out to be free, and I made it falsifiable

Its `pre_query` is `LIMIT 1`, so the first exercise IS a one-site canary by construction — no
engineering needed. Baseline taken BEFORE any pass could run: `retracted_so_far` **0**, 226
`deferred`, 0 with `batch_id`.

Then, rather than record a bound and wait, I named the outcome. First site up is `robot-hands.com`
(due 08-17 14:54:23Z), ceiling **34 rows / 21 pages**. And `/selection-guide.html` there holds
exactly three open rows, which is the sharpest test available and it came with its own control:

| key | required outcome | why |
|---|---|---|
| `…#A.info-card-grid__card-link` | **RETRACT** | migration `368` fixed it |
| `…#SPAN.info-card-grid__eyebrow` | **RETRACT** | same |
| `…#A.cta-btn` | **STAY OPEN** | `368` did not touch it |

Same page, same run, opposite required outcomes — so the run distinguishes "retraction works" from
"retraction closes everything on a page it looked at", which no closure COUNT can do. This is the
lane's own standing rule applied to my own verification: a measurement that could not come out
otherwise is not evidence. If all three close, the scope is wrong.

`loancalculator.co.uk` is second up with **0** open rows — a pass that retracts anything there is
also a scope bug.

---

## 2026-08-13 (evening) — CONTRIBUTED FROM OUTSIDE THIS LANE: `legibleInkFor`'s preference walk cannot return anything but `text`, so `--color-<x>-ink` is `--color-text` renamed

Not this lane's session. I took the `article-body` consumer named in `bugs_open/122`'s 08-12
contribution (which deliberately applied nothing and handed the renderer-level repair here),
intending to repoint it plus the other 167 components carrying the same shape. **Nothing was
applied.** Recording here because the finding falsifies a claim in this lane's own `PLAN` and
`VIZ-014`, and because it changes what this lane should do next. Full write-up: `bugs_open/122`
contribution 2026-08-13 §§1–10. Landmine filed same date. `WRONG_CALLS.md` entry same date.

### The finding

`palette_specialised_slots.go:350` — `for _, key := range []string{"text", "accent", "text_muted",
"secondary", "primary"}`. First match wins; `text` is first; `text` is by construction the slot
chosen to be legible on `background`, so it clears the grounds whenever any candidate does.
**`accent`, `text_muted`, `secondary`, `primary` are unreachable in production.**

`[MEASURED 2026-08-13, served stylesheets, all 22 live domains — 18 palette-driven]`: every
divergence between an ink companion and its source slot equals that site's own `--color-text`.
**16 of 16, zero exceptions.** Table in the bug file. Any one row returning a third colour would
have falsified it, and 14 palettes × 2 slots is a wide enough net that this is evidence rather than
coincidence.

So `PLAN_2026-08-06:189-195`'s *"the first palette colour that does (the existing `pickInkOn` walk,
which prefers a palette colour so the site keeps its character)"* is **false as built**, and so is
the slot description at `palette_specialised_slots.go:401`. VIZ-014 corrected in place.

**What is NOT wrong:** the four shipped repoints (338, 368). Those elements were at 1.06–1.14:1 —
invisible — and now measure 13–16:1. That was measured at the artefact and it stands. The error is
confined to the belief that the repoint preserves the brand colour.

### Why it stopped the sweep this lane was heading toward

Corpus `[MEASURED 2026-08-13]`: **168 components (135 active)**, **453 declarations** = 423 in-block
+ 30 in inline `style="…"` attributes (16 tool components, invisible to any block-walking
transform). Block-scoped: 259 primary blocks (51 self-painting) + 164 accent (25) → ~347 eligible.
The defect also lives on four further surfaces — `layouts.css_template` **17 of 18**,
`css_snippets` 2/21, `site_components.rendered_html` 33/66, `page_components.rendered_html`
461/1485 — so a `content_components`-only transform is canonical for about a fifth of it.

Automating the repoint on that eligibility rule would have written the equivalent of
`color: var(--color-text)` across those components' placements on the 14 diverging sites. **And
`render_audit.py` scores that a clean pass** — it measures contrast, and near-black on a light
ground has excellent contrast. This lane's own grading discipline ("grade per selector, at the
artefact") would not have caught it either: the selector would genuinely have improved.

### A second defect, in the eligibility rule, refuted by this lane's own ground truth

The rule "skip any declaration block that paints its own background" **refuses `system-stats
.stats-eyebrow`** — the only hand-made `--color-accent-ink` repoint in the corpus, made from the
owner's evidence — because its background is `var(--section-surface)` = `rgba(255,255,255,0.05)`
(`color_util.go:220`), a translucent overlay whose element sits on the page ground. 1 of 6
ground-truth rules, 1 of 1 on accent. **"Sets a background" ≠ "the ink lands on a different
ground."**

Reuse rather than re-derive: `fix_forced_text_colours_action.go:164-188` already carries a
calibrated **four-way** `paintClass` classifier over this same corpus, council-reviewed, and it
classifies from the whole template rather than per block — i.e. the prior art already decided that
per-block is the wrong granularity. `bugs_open/122:205-206` names that file as worth reading first.
It was right and I found it late.

### Proposed next step for this lane (not taken — it is yours)

Fix the derivation, not the consumers: in `legibleInkFor`, try stepping the source colour's
lightness toward whichever achromatic extreme gains contrast until it clears `inkMinContrast`
against **every** ground, *before* falling through to the palette walk. Retroactively improves all
four shipped repoints with no further config edits. Constraints already recorded here that must
survive: `grounds` stays a slice (R8), no `isDarkHex` narrowing (R7), the already-clears-AA branch
still returns the source unchanged. Go, so it needs a roll; architecture-scope on the 2026-07-29 §1
test (it changes what the shared mechanism guarantees), so council-gate it.

**Control D, the disconfirming test:** after the change, dartsonline's `--color-primary-ink` must be
a **navy** — a lightened relative of `#1A1F2E`. If it is still `#F0F2F7`, the fix did not land and no
repoint should run.

### Missteps this session, in full

1. **I marked a subagent's figures `[MEASURED]`.** I ran an adversarial reviewer against my own
   plan; its structural argument was sound and produced the finding above, which I then verified
   myself. But I wrote its *supporting* counts into the bug file as though I had measured them. Of
   the two I re-checked, one was wrong by 20× ("102 components carry `background-color:
   var(--color-…)`" — actual **5**) and one was right-for-the-wrong-reason (the 11-component
   anchored/unanchored gap is compound colour properties generally, not `background-color`). Three
   more are now marked `[UNVERIFIED — inherited]` with their queries in the bug file §9, because the
   kubeconfig token expired before I reached them. **Trust an adversary's judgement; re-measure its
   arithmetic.** `WRONG_CALLS.md`.
2. **A control that could not fail.** Probing the chassis binary for provenance I used 40 zeros as
   the must-be-ABSENT sha; it **matched** (Go's internal tables carry zero runs). Re-ran with
   yesterday's real build sha, correctly absent. Same family as CLAUDE.md's warning against a
   *discovery* grep for "some 40-hex string" — here wearing a control's clothing.
3. **I expected my own anchor to undercount and it does not.** `[;{ ]color:` omits `\n` from the
   class, which I assumed would miss pretty-printed declarations. Measured: `[;{ ]` and
   `(^|[;{[:space:]])` return the **identical** 168 components / 453 declarations. Worth recording
   because the check took one query and I nearly rewrote the census on the assumption.

### Owed, blocked on the cluster login

- `./scripts/landmines-sync.py --apply` — the file entry is committed (D10: the file is the system
  of record) but the `doc_notes` rows did not write; `--check` will report drift until re-run.
- The three `[UNVERIFIED — inherited]` figures in `bugs_open/122` §9, queries written out there.

---

## 2026-08-13 (evening) — lane picked up, state re-verified, and the third canary row's root cause found

Session goal was to pick the lane up from `HANDOFF_2026-08-12c` and check other threads first. The
lane itself is date-gated and nothing was owed today, so the work is verification plus one finding.

### 1. Nothing has drifted — the §1c prediction still holds exactly

```
status   |  n  | retracted | max_att
deferred | 226 |         0 |       0        -- unchanged since 08-11
```
`site-render-audit-rotation` enabled, `last_triggered_at` 16:28Z, and **0 sites due**
(`count(*) WHERE last_selected_at < now() - interval '7 days'` → `0`). Due order is unchanged:
`robot-hands.com` **2026-08-17 14:54:23Z**, `loancalculator.co.uk` 15:54:31Z, `cookly.uk` 16:55:01Z.

Live-ness re-proved at the artefact rather than assumed, with both controls, on the **v1.0.1295**
pods (started 13:53Z — a newer release than the one §1c verified, so this needed redoing):
`grep -aq` on `/proc/1/exe` → `69612d692…` **PRESENT**, `7a1887e31` **absent** (so the probe
discriminates); `git merge-base --is-ancestor 5639a1103 69612d692…` → **true**.
`browser-runner-adapter` and `render-audit-adapter` both print `69612d692…` in their own startup
provenance line. **A newer fleet release did not displace the retraction** — that was the thing
worth checking, and it is checked, not inferred.

### 2. kubectl expired mid-session — one item is half-done and owed

Fleet-wide `Unauthorized` at ~16:52Z, i.e. the 3-day token expiry (owner refreshes). It landed
between the queries above and the producer census, so **§1b(1)'s standing caveat is HALF re-run**:

- **done, at HEAD:** exactly one Go file mints the type — `write_render_audit_findings_action.go`
  (`:296`, `:304`, `:535`, `:547`, `:604`). No second producer has appeared in code.
- **owed, needs a live token:** the row-side half (`source` / `created_by` / `spec ? 'audit_source'`
  / distinct spec key-sets over the 226). Re-run it before anything else starts filing
  `contrast_failure` — the census cannot see a producer that has never fired, which is the whole
  point of the caveat.

### 3. Another thread is live inside this bug, and its finding is real — I checked three of its sixteen

Session `581eb30a` (autonomous "take the next unworked bug", started 08-12 20:48Z, **still running**)
has **181 uncommitted lines** in `bugs_open/122`. I did not touch that file. Its claim:
`--color-<x>-ink` is `--color-text` under another name on every site, so the "keeps its character"
promise in this lane's own `PLAN_2026-08-06:189-195` is false as built.

**Verified independently before recording it**, because an uncommitted claim in another session's
working tree is not evidence:

- the deciding line, read at HEAD — `palette_specialised_slots.go:350` walks
  `{"text","accent","text_muted","secondary","primary"}` and returns the **first** candidate that
  clears every ground. `text` is by construction the slot picked to be read on the page ground.
- three of their sixteen sites, re-fetched at the served artefact:

| site | `--color-text` | `--color-primary-ink` | `--color-accent-ink` |
|---|---|---|---|
| robot-hands.com | `#E2E8F0` | `#E2E8F0` | `#E2E8F0` |
| dartsonline.com | `#F0F2F7` | `#F0F2F7` | `#F0F2F7` |
| cookly.uk | `#2C2C27` | `#2C2C27` | `#2C2C27` |

Ink == text, exactly, 3 for 3. One caveat on their §2 wording, recorded because this lane cares
about the difference: *"`accent`/`text_muted`/`secondary`/`primary` are unreachable"* is a
**[MEASURED]** fleet fact, not a logical necessity — the walk would return `accent` on a site where
`text` failed one of the grounds and `accent` cleared them all. Nothing in the fleet is such a site.
The measurement is the evidence; the mechanism only explains why it came out that way.

### 4. ⚠ That thread's proposed fix would land on top of this lane's 08-17 canary

Not a conflict of edits — a conflict of **experiments**, which is harder to see and was worth the
check. Two of the three canary rows are `--color-primary-ink` consumers. Verified in the **served
page**, not in the template:

```
robot-hands.com/selection-guide.html
  :554  .info-card-grid__eyebrow    color: var(--color-primary-ink, var(--color-primary));
  :648  .info-card-grid__card-link  color: var(--color-primary-ink, var(--color-primary));
  :805  .cta-btn-primary            color: var(--color-cta-bg,     var(--color-primary));   <- NOT an ink consumer
```

So if a derivation fix lands **and** robot-hands re-renders before 08-17 14:54:23Z, the two
"must RETRACT" rows are no longer testing retraction — they are testing whatever the new ink
computes to. The run could then read as "retraction is broken" when the derivation simply moved.
**The `A.cta-btn` control is immune** (different variable), so the over-closure half of the test
survives intact either way — which is a piece of luck worth knowing rather than discovering on the
day. Mitigation is free: either nobody re-renders robot-hands before Monday, or whoever does
re-states the two predictions against the served page first.

### 5. NEW FINDING — why `A.cta-btn` fails, and it is a defect class this bug has not recorded

§1c predicted this row must stay open on the sound but shallow ground that "`368` did not touch it".
It now has a mechanism, and the mechanism is **not** an ink-slot problem at all:

```css
:root (styles.css:43-44)   --color-cta-bg:   linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
                           --color-cta-text: #ffffff;

.cta-section    (page:765) background: var(--color-cta-bg,   …);   /* gradient — VALID for background */
                    (:766) color:      var(--color-cta-text, …);   /* #ffffff  — correct, intentional */

.cta-btn-primary(page:804) background: var(--color-cta-text, …);   /* #ffffff  — a white pill, intentional */
                    (:805) color:      var(--color-cta-bg,   …);   /* A GRADIENT IN A COLOUR SLOT */
```

The intent is obvious and sane: a white button whose label is the CTA's brand blue. But
`--color-cta-bg` holds a **gradient**, which is not a valid `<color>`. The variable **is** defined,
so `var()`'s fallback `var(--color-primary)` — which is `#1A1F2E` and would have been fine — is
**never reached**. The declaration is invalid at computed-value time, `color` therefore **inherits**
from `.cta-section` (`#ffffff`), and the result is `#ffffff` on `#ffffff`. That is the invisible
text this bug is about, arriving by a route no ink slot and no repoint can fix.

**CONFIRMED AT THE INSTRUMENT — this is no longer inference.** The token came back mid-session, so
the disconfirming check I had written down for someone else got run here. The audit's own filed row
carries what it measured, and it is exactly what the mechanism predicts:

```
contrast_failure:/selection-guide.html#A.cta-btn
  fg: "rgb(255, 255, 255)"   bg: "rgb(255,255,255)"   ratio: 1   need: 4.5
  text_sample: "Run MatchMatrix"
```

`text_sample` settles which of the two buttons it is: "Run MatchMatrix" is the `href` on
`<a class="cta-btn cta-btn-primary">`, so the failing element is the **primary** button, the one
carrying the gradient-in-a-colour-slot. Any other `fg`/`bg` pair would have refuted the mechanism
outright; `#ffffff` on `#ffffff` at exactly 1.00 is the only reading consistent with "the
declaration was discarded and `color` inherited from `.cta-section`".

**Blast radius, from the filed rows rather than a curl sample** — 17 rows across 4 sites match
`spec->>'selector' LIKE '%cta-btn%'`:

| site | rows | fg == bg at 1.00:1 |
|---|---|---|
| robot-hands.com | 10 | **10** |
| finetuning.uk | 4 | **4** |
| ai-agent-orchestration.com | 2 | **2** |
| leopardessconsulting.co.uk | 1 | 0 — min ratio **2.27** |

**16 of 17 are the type error. The 17th is the control and it is a different defect on a different
selector** (`A.tool-cta-btn-primary`, `fg` `rgb(200,169,81)` on white = the site's `--color-cta-bg`
`#C8A951` used as a label *validly*, and simply too pale). Same token, valid value, declaration
applies, ordinary low contrast. That is the contrast between the two failure modes in one query.

### 5a. MISSTEP — I measured the token at the wrong layer and briefly refuted myself

Worth recording because the wrong answer looked exactly like a right one, and the *check* was what
was broken, not the claim.

First pass sampled `--color-cta-bg` from `https://<domain>/assets/css/styles.css` only, and reported
**"gradient on 3 of 8 sites"**. Then `ai-agent-orchestration.com` turned up with 2 white-on-white
rows and a site-level `--color-cta-bg: #0D1117` — a perfectly valid colour. That is a genuine
refutation of the mechanism as I had stated it, and I treated it as one.

It was the measurement. **The page's own `<style>` block redefines the token** — `aao/about.html:76`
sets `--color-cta-bg: linear-gradient(135deg, #1e40af 0%, #1e3a8a 100%)`, overriding the stylesheet.
Reading the site stylesheet alone yields a **false negative** on exactly the sites where a page-level
override introduces the defect.

Re-measured at the **effective** layer (page `<style>` if present, else the stylesheet), 10 sites:

| effective `--color-cta-bg` | sites |
|---|---|
| **GRADIENT** | `robot-hands.com` (site), `finetuning.uk` (page), `gaswholesalers.com` (page), `ai-agent-orchestration.com` (page), `leopardessconsulting.co.uk` (page) — **5** |
| plain colour | `dartsonline.com`, `cookly.uk`, `vetcomparison.uk`, `oufe.com`, `lendzy.co.uk` — **5** |

So **5 of 10, not 3 of 8** — and every white-on-white row in the table above sits on a gradient site
once the right layer is read. The corrected figure is also the more alarming one, which is the
direction a measurement error is least likely to be caught in.

**And the page-level gradient is the SAME LITERAL on 4 of the 5** —
`linear-gradient(135deg, #1e40af 0%, #1e3a8a 100%)`, a blue that belongs to no site's palette,
appearing verbatim on four unrelated sites. That points at a shared page/layout template default
rather than anything the palette generator computed, and it is the same shape as `bugs_open/113`
(generated palettes inheriting the layout's light literals). `[UNVERIFIED]` as to which template
emits it — I did not chase it, and it is 113's territory rather than this lane's.

**The transferable check: resolve a custom property at the layer that actually wins, never at the
one that is convenient to `curl`.** A page `<style>` block beats the site stylesheet, and this
estate uses page-level palette blocks routinely.

The general shape is one this lane keeps meeting from new directions and should name:
**a `var(--x, fallback)` whose `--x` is DEFINED BUT OF THE WRONG TYPE is strictly worse than an
undefined one** — the sane-looking fallback in the source is dead code, and an inherited-value
failure looks in the template exactly like a working declaration.

### 6. The selector is lossy — checked, and it does NOT threaten the canary

Chased this as a possible false-closure route and it came out clean, so recording the negative.
`contrastSelector` (`write_render_audit_findings_action.go:777-786`) keeps **`tokens[0]` only**, so
`class="cta-btn cta-btn-primary"` and `class="cta-btn cta-btn-secondary"` **both** key to
`A.cta-btn` — and the served page carries exactly one of each. The browser-side dedup
(`render_audit_action.go:222`) keys on the *full* class string, so the two are distinct there and
collapse only at the filing layer.

**The retraction is nevertheless sound here, because it derives its still-failing keys with the
same function on both sides** (`:548` uses `contrastSelector` exactly as `:252` does). The row
retracts only when no element sharing that key still fails — conservative, i.e. it errs toward
leaving tickets open. The residual is cosmetic rather than dangerous: one row can stand for two
elements, so the 226 is a floor for the *element* count as well as for the reasons already recorded.

### 7. Two corrections to my own commit message, and the cross-thread loop closing

**`5baecdfe1`'s last paragraph is wrong in one respect and cannot be amended** (forward-only), so it
is corrected here. It says the commit "unavoidably takes two passengers … in that append-only file".
**It does not — `LANDMINES.md` is not in that commit at all.** Between the `git diff --numstat` I
based that sentence on and the `commit` itself, session `581eb30a` committed `2009b9243`, which
carried my landmine correction as *its* passenger. My pathspec therefore found nothing left to take,
and the two entries I named (`site_work_items.item_key` re-type; the `kubectl patch` /
`047-base-configs` trap) remain uncommitted in the tree, still their authors' to commit. Nothing was
lost and nothing of mine is missing — the claim was simply stale by the time it was written.

**The general lesson, which is the reason this is worth a paragraph rather than a shrug:** a
same-file passenger census goes stale between the measurement and the commit, on exactly the file
most likely to be written concurrently. I described the tree as it had been a minute earlier and
stated it in the present tense. `git show --stat <sha>` after the fact is the only account of what a
commit contains that cannot be overtaken.

**And the cross-thread loop closed within the session, which is the part worth keeping.** The
message I sent to `581eb30a` came back as commit `2009b9243`: it took the "unreachable" wording
correction and propagated it to four places I could not see from here (the bug file and its heading,
`LANDMINES`, register entry **VIZ-014**, and the Go doc comment), and it carried the gradient-type
finding into `bugs_open/122` under its own name — which is what I asked for, and avoided the
same-file collision that filing it myself would have caused. Contributing into another thread's open
file by *message* rather than by edit cost one tool call and produced a better result than either
competing for the file or waiting for it.

## 2026-08-14 — the sibling fix is COMMITTED, and the canary's exposure is computed rather than feared

`581eb30a` committed the `legibleInkFor` repair (`12cf55015`): `colour.LegibleVariant` gets first
refusal ahead of the palette walk, moving the source in HSL lightness only and returning the
smallest sufficient change. It is not rolled, but `make build-*` builds from committed HEAD, so
**the next fleet roll carries it, whoever triggers it and for whatever reason.** The hazard recorded
yesterday therefore stopped being conditional on that thread's own actions.

### 1. I nearly filed a regression against their fix, and the check that stopped me was arithmetic

They quoted the new robot-hands ink as `#7785b2`. Measured against the grounds **the audit itself
recorded** (`spec->>'bg'`, so the composited reality rather than a template's claim):

| ink | vs card-link ground `rgb(30,37,53)` | vs eyebrow ground `rgb(15,18,24)` |
|---|---|---|
| today `#E2E8F0` | 12.42:1 PASS | 15.21:1 PASS |
| their quoted `#7785B2` | **4.22:1 FAIL** | 5.16:1 PASS |
| original defect `#1A1F2E` | 1.07:1 FAIL | 1.14:1 FAIL |

On that reading their fix **regresses** `A.info-card-grid__card-link` — an element migration `368`
repaired — back into a filed failure, and my "must RETRACT" prediction flips. That is a serious
claim about someone else's committed code, so it went through the code before it went anywhere near
a message.

**It does not survive the code.** `LegibleVariant` (`platform/colour/contrast.go:301`) requires
`clearsAll` — *every* ground at or above `minRatio` — before returning, and the caller
(`palette_specialised_slots.go:480`) passes `pageGrounds = {bg, surface}` with
`inkMinContrast = 4.5`. It is structurally incapable of emitting a colour that fails a ground it was
handed. So `#7785B2` cannot be what it produces for this palette.

Replicating the algorithm faithfully (HSL, step 0.01, both directions per distance, first hit wins)
against the served palette gives **`#7D8BB6`** — **5.57:1** and **4.55:1**. Both clear. **The
predictions hold.** `[MEASURED]` by replication, not by running their binary — my HSL round-trip may
round differently from `platform/colour`'s, which is exactly the residual and is now their test to
write.

**The lesson: a number quoted in prose is not the number the code emits, and the gap between the two
was the whole finding.** Had I trusted their hex I would have filed a false regression against a
sound fix; had I trusted my alarm I would have raised it without the replication that dissolved it.
Both hexes are ~0.02 lightness apart — two steps of their search — and land on opposite sides of the
threshold.

### 2. The structural finding, which is worth more than either hex

"Smallest sufficient change" means the walk returns the **first** lightness that clears, so **every
output sits by construction within one step of the failure boundary.** `[MEASURED 2026-08-14]`
across 10 served palettes, both ink slots, 12 emitted values:

| margin over 4.5 | sites |
|---|---|
| **within 0.10** — 7 of 12 | webdesign.co.uk accent **+0.02** · vonc primary +0.04 · robot-hands accent +0.04 · robot-hands primary +0.05 · dartsonline accent +0.05 · finetuning accent +0.06 · dartsonline primary +0.09 |
| 0.10–0.20 — 5 of 12 | oufe primary +0.11 · cookly accent +0.12 · lendzy accent +0.14 · vetcomparison accent +0.17 · relojistas accent +0.18 |

Now set that against the premise this entire lane rests on: **the derivation reasons about the
palette's DECLARED `surface`; the render audit measures the COMPOSITED ground actually painted.**
Those differ routinely — `--section-surface` is `rgba(255,255,255,0.05)` over the page ground, which
that thread's own §5 established. **At +0.02 headroom, a composited ground marginally off the
declared one flips the element back under 4.5 and the audit files a fresh `contrast_failure` for an
element the fix just repaired.**

Predicted consequence, recorded now so it is falsifiable later: **expect some new `contrast_failure`
volume after the roll on elements the ink fix "fixed", and expect it to look like a retraction
defect.** It would not be one. The remedy is theirs and cheap — walk to a target above the threshold
(4.6, 5.0) so there is absorption, or feed the derivation the composited ground.

### 3. Canary procedure, now three-branched

The discriminator is free and is one `curl` before 14:54:23Z on 08-17: `#E2E8F0` → my retraction is
under test · `#7D8BB6` → theirs is underneath and both rows still retract · `#7785B2` → the
regression branch, `card-link` stays open and it is not a retraction bug. Read the **page** block
first, not just the site stylesheet (§5a). Full table in the handoff.

### 4. Housekeeping done for both threads

`landmines-sync.py --apply` run on my live token (theirs expired): `--check` now reports **in sync**,
so both entries reached `doc_notes`. They restored the two continuation lines their stale write had
dropped from my landmine entry (`eb07f1fc9`), and logged it in `WRONG_CALLS`; the two-source `pv=`/
`sv=` form is at HEAD and is better than what was lost. Their correction of my `[INFERRED]` marker is
now itself out of date — the token came back and §5 records the instrument reading, so it is
confirmed; they have been asked to drop the marker in `bugs_open/122` §11 next time they touch it.

### 2026-08-14 — the derivation fix, ROUND 2: certifying against the composited ground, and a mutation that PASSED

Round 1 committed as `12cf55015`; round 2 as `8ad05d01a`. Council trail
`afcec886-f84c-4fb4-8876-43502e70965b` (submitted on round 1's design, in flight at `gate_guidelines`
when round 2 landed — deliberately not resubmitted mid-round; read the verdict, then resubmit on the
trail with `RESUBMIT_CORR` if it wants the newer design).

**This lane reviewed round 1 within the hour and both objections were right.**

1. **My quoted hex was wrong; the code was not.** I reported robot-hands emitting `#7785b2`. The real
   emission for its real palette is **`#7d8bb6`** — identical to dartsonline's, as the reviewer
   deduced. Mine came from a throwaway probe whose **grounds I had invented** (`#0F1319`/`#1A1F2E`
   instead of the served `#0F1218`/`#1E2535`). Nothing could contradict me because no test named an
   output. `TestLegibleVariant_EmittedHexIsPinnedForRealPalettes` now pins four, with inputs
   transcribed from the artefact. *A probe with fabricated inputs produces a figure indistinguishable
   from a measured one* — and I published it to a peer.
2. **The structural objection, and it was worse than they could see from outside.** "Smallest
   sufficient change" puts every output within ~0.1 of the floor **by construction**, while the
   derivation reasoned about the palette's *declared* surface and a component painting with
   `--section-surface` puts the ink on a ground 5% lighter. `[MEASURED]` that overlay costs **0.62**
   of contrast ratio on the dark palettes — ~10× the headroom — and round 1's own emissions measured
   **3.93–4.03:1** on the composited ground. **Round 1 would have filed a fresh `contrast_failure`
   for `A.info-card-grid__card-link`, an element migration 368 repaired**, and this lane's Monday
   canary would have read it as a retraction regression.

**Fixed at the cause, not padded.** `buildLegibleInkDefaults` composites the overlay onto both page
grounds and requires all four to clear; `colour.CompositeOverGround` does it deliberately, because
`platform/colour` refuses to composite alpha implicitly (its header explains why) and that left this
exact gap for a caller that *knows* its overlay. Re-measured worst-of-four: robot-hands primary
`#8a97bd` 4.56, dartsonline primary `#8a97bd` 4.60, vonc primary `#9b6aff` 4.62, cookly accent
`#c04d28` 4.53, webdesign.co.uk accent `#9d6630` 4.52 — all still recognisably the brand.

**I rejected the margin the reviewer offered as an alternative** (walk to 5.0 instead of 4.5), and
the reasoning is worth keeping: it buys absorption without fixing the wrong-ground error, and it
would *imply* that unmodelled grounds are handled. They are not — component-painted grounds stay
`bugs_open/212` §8, over-image stays excluded from firm findings. Better a narrow guarantee stated
than a wide one implied.

**⚠ THE MOST USEFUL THING IN THIS ENTRY: a mutation that PASSED.** Deleting the compositing loop from
`pageGrounds` left **every test in the actions package green**. The composited-grounds test lived in
`platform/colour` and exercised `LegibleVariant` with hand-built grounds — so the **unit was pinned
and the WIRING was not**, and the mutation walked straight through the gap. A test proving a function
honours an argument says nothing about whether its caller passes that argument.
`TestBuildLegibleInkDefaults_EmittedInkClearsTheCompositedGround` now reads the emitted `:root` block
and measures the value it *actually contains*; under that mutation it goes RED naming **3.93:1**, the
figure this lane predicted. Five mutation proofs now stand on this change (M1, M3, M4 on round 1; M5,
M6 on round 2), each a distinct failure, files restored byte-identical each time.

`sectionSurfaceOverlayAlpha` is now a named constant with **two** readers — the emitted CSS literal
and the derivation — joined by `TestSectionSurfaceOverlayAlphaMatchesTheEmittedCSS`, which fails if
either moves. The literal stays inside the format string so emitted bytes cannot change.

**Still inert until `agent-chassis` is rebuilt and rolled.** Control D unchanged: dartsonline's
`--color-primary-ink` must read as a navy, now `#8a97bd`. If it reads `#F0F2F7` the change did not
ship; if it reads `#7d8bb6` **round 2 did not ship** and the composited grounds are absent.

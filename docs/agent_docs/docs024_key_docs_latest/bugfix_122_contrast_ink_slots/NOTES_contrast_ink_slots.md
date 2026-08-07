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

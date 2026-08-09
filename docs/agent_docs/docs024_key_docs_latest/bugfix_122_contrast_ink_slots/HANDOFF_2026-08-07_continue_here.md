# HANDOFF — bug 122, contrast / ink slots. START HERE.

**Written 2026-08-07, last updated 2026-08-08 (§0).** Supersedes
`HANDOFF_2026-08-06b_continue_here.md`. **That file's "next action 1" is now DONE** and
its §"What NOT to do" list still stands in full — it is not repeated here, so read it
there before touching code. Everything else below is new or changed.

## The one-paragraph state

The **engine half is live and has stayed live across three rolls**: v1.0.1266 carries
VIZ-014, pod-proven on both replicas (§1). No lane in this workstream rolled any of them. The
**config half is APPLIED** (migration `338`, 2026-08-08 22:12:55Z, verified at the rows)
— but **nothing a visitor sees has changed yet**: the propagation re-render is the whole
remaining job (§2b), and a second migration is still unwritten (§2b foot). Separately, picking up
`bugs_open/212` reframed it twice over: its fix ranking is refuted by arithmetic, its
"unenforced contract" premise is wrong, and the real blocker turned out to be a work-item
contract defect now filed as **`bugs_open/213`** (§3, §4).

## If you are starting cold, do these five things first

They take about ten minutes and every one of them has caught a stale figure in this
lane's own history.

1. **Re-read `CLAUDE.md`** from disk — it is co-edited and it changed twice this week.
2. **Re-prove the engine at the pod** (§1 has the exact loop). The image tag in the
   makefile is not evidence; another lane rolls it every few hours.
3. **Re-check the free migration number** (§2). Wrong twice in two days.
4. **`git log --oneline -15`** and `git status` — this tree has ~30 concurrent writers,
   and your session-start snapshot is already stale.
5. **Read `bugs_open/212` §8 before its §1–§7**, and `bugs_open/213` whole. §1–§7 of 212
   are the version of the story that was wrong.

Then pick up at **§2b** — migration 338 is applied; the propagation re-render is what is
left, and until it runs no visitor sees any of this.

## 0. What changed on 2026-08-08, before you trust anything below

Re-measured at the start of the 08-08 session. **Four of this file's own figures went stale
within a day**, one of them twice — that rate is the point, not the individual numbers.

| what | was (08-07) | is (08-08) | so |
|---|---|---|---|
| chassis image | v1.0.1262 | **v1.0.1270** (re-proved; five rolls, none ours) | §1 — a new image is a new fact |
| migration 338 | not written | **APPLIED + verified 22:12:55Z** | §2 |
| `090` run 4 | still `diagnosing` | **`complete`, UNVERIFIABLE** | §5 |
| what is left | the migrations | **propagation — blocked on an owner yes; 3 pins then dispatch** | §2b |

Nothing has touched `bugs_open/212`, `bugs_open/213` or this directory since the 08-07
commits (`git log b938b54d8..HEAD -- …` is empty), so §3–§4 stand as written.

## 1. DONE — the engine is live, pod-proven [re-proved 2026-08-08]

Both replicas of `agent-chassis`, image `docker.io/aqls/agent-chassis:v1.0.1270`:

| symbol | count | role |
|---|---|---|
| `buildLegibleInkDefaults` | 4 | the new emitter |
| `legibleInkFor` | 3 | " |
| `worstRatioAgainst` | 2 | " |
| `fillDarkSchemeSpecialisedSlots` | 4 | positive control |
| `zzzInventedControlXyz` | 0 | negative control |

Identical counts on v1.0.1262, 1264, 1266, 1269 and 1270 — five rolls, none of them ours. Nothing downstream of a
build records which commit it came from, so the symbols are the only evidence — which is
why both controls are in the table and why every replica was checked. **Do not re-derive
this from the tag, and do not carry this table forward across another roll without
re-running it** — it has already survived three rolls it was not written for, which is
luck, not evidence.

```bash
for POD in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[*].metadata.name}'); do
  echo "=== $POD $(kubectl -n ai-persona-system get pod $POD -o jsonpath='{.spec.containers[0].image}') ==="
  for s in buildLegibleInkDefaults legibleInkFor worstRatioAgainst fillDarkSchemeSpecialisedSlots zzzInventedControlXyz; do
    printf "  %-32s " "$s"; kubectl -n ai-persona-system exec $POD -- sh -c "strings /app/agent-chassis | grep -c '$s'"
  done
done
```

## 2. DONE — migration 338 is APPLIED and verified at the rows [2026-08-08 22:12:55Z]

`338_components_opt_into_legible_ink_slots.sql`, recorded in `schema_migrations`.
The run printed `Pending (1)`, six `DO` blocks, `COMMIT`, `recorded`, and nothing else.

Post-state, and it matches the dry run exactly:

| row | new token | negative control |
|---|---|---|
| `case-studies-grid` | 2 × `--color-accent-text` | — |
| `system-stats` | 1 × `--color-accent-ink` | **2 accent backgrounds intact** |
| `image-hover-card-grid` | 1 × `--color-primary-ink` | focus outline intact |
| `tool-list` | 2 × `--color-primary-ink` | **2 primary backgrounds intact** |
| 5 layouts | 1 × `--color-accent-ink` each | — |

Before applying: the whole file was dry-run in a rolled-back transaction (which found two
bugs in the migration itself — see NOTES), three RAISE paths were induced to prove they
fire, and both `\copy` backups were taken to
`…/scratchpad/backups/backup_338_{content_components,layouts}.tsv`. Rollback is at the foot
of 338.

> **READ THIS BEFORE ANY FUTURE `--apply`.** The first attempt at this apply was made with
> `MIGRATIONS_DIR=… ./run-migrations.sh --apply` entered as **two lines**. `VAR=value` on
> its own line is an unexported shell assignment, so the runner used its default directory
> — **98 pending files** — and applied four other threads' migrations (198, 203, 204, 207)
> before halting on a syntax error in 208. `204` changed 10 live `products.content_data`
> rows on robot-hands; `198` created `gauntlet_rounds` in `clients_db` when migration 276's
> own guard says it belongs on the ISLAND. **Those four are still applied and recorded —
> forward-only, and they are other threads' files, so whether they stand is the owner's
> call, not this lane's.** Full account: `LANDMINES.md` and `WRONG_CALLS.md`, 2026-08-08.
> **Always `env VAR=x cmd` on one line, and always read the runner's `Pending (N)` line
> before `--apply` — on a scoped run it must say 1.**

## 2b. NEXT — PROPAGATION. Nothing a visitor sees has changed yet

**This is now the whole remaining job.** The migration changed the SOURCE; every live page
still carries its old `rendered_html` and every site still serves its old stylesheet. This
is the council's still-open `editquality` MEDIUM objection, and §4 is why "the queue says
complete" will not be evidence it happened.

**Component placements — 16 across 8 sites [MEASURED 2026-08-08]:**

| site | component | placements |
|---|---|---|
| ai-agent-orchestration.com | case-studies-grid 2, system-stats 2, tool-list 1 | 5 |
| gamesdesign.co.uk | system-stats 1, tool-list 2 | 3 |
| idea.uk | tool-list 2 | 2 |
| robot-hands.com | system-stats 1, tool-list 1 | 2 |
| dartsonline.com | image-hover-card-grid 1 | 1 |
| finetuning.uk | case-studies-grid 1 | 1 |
| leopardessconsulting.co.uk | case-studies-grid 1 | 1 |
| vonc.com | system-stats 1 | 1 |

```sql
SELECT s.domain, cc.name, count(*) FROM page_components pc
  JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
  JOIN content_components cc ON cc.id=pc.component_id
 WHERE cc.name IN ('case-studies-grid','system-stats','image-hover-card-grid','tool-list')
 GROUP BY 1,2 ORDER BY 1,2;
```

**The layouts half is WIDER and is a stylesheet re-render, not a page one.** The site↔layout
join runs through `css_themes.layout_id` (there is no `sites.layout_id` — that cost several
wrong guesses). 14 active themes carry the five changed layouts:

| layout | themes |
|---|---|
| brochure-formal | 5 — default, standard-brochure, professional-dark, fundamentallyai.com, leopardessconsulting.co.uk |
| tool-portal-light | 4 — idea.uk, lendzy.co.uk, mortgagecalculator.co.uk, webdesign.co.uk |
| technical-precise | 2 — modern-engineering-clean, premium-elegant |
| tool-portal-dark | 2 — gamesdesign.co.uk, robot-hands.com |
| high-energy | 1 — boxing |

> **RESOLVED 2026-08-09 — the layouts change DOES reach gaswholesalers.com.** Asked its
> served stylesheet rather than the schema: it carries the multi-line needle
> `a {\n  color: var(--color-accent);` exactly once, so it runs one of the four multi-line
> layouts. `--color-accent-ink` appears **zero** times in that file, which is expected and is
> the point — the served CSS predates both the engine and 338. §6's expected-12 stands.

### BLOCKED — the only CSS re-render path runs an LLM design pass on an unpinned site

**Determined 2026-08-09, and this is a decision for the owner, not a detail.**

**A CSS re-render is unavoidable.** The components now say
`var(--color-primary-ink, var(--color-primary))`. Those variables are *defined* only by the
engine's step 12, which runs during a stylesheet render. Until a site's `styles.css` is
regenerated, every new reference resolves to its fallback — i.e. renders exactly as today.
**So without a CSS re-render the entire fix, component half included, is inert.** Confirmed
at the artefact: gaswholesalers' served stylesheet carries the base-link needle once and
`--color-accent-ink` **zero** times.

**The only path to `render_css_from_spec` is `webdesign-agent`**, and its step graph forces
the LLM through first:

```
check_site_context → … → load_decisions → analyze_design (execute_llm_prompt)
                       → check_update_db → update_site → generate_css (render_css_from_spec)
                       → deploy_css
```

There is no entry that reaches `generate_css` without `analyze_design`. It is the only
active agent holding that action (checked across all live, non-snapshot definitions).

**And `analyze_design` re-invents the palette per run.** That is the recorded fleet-wide
mechanism from 2026-07-17 on robot-hands — four CSS rewrites in one day, one of which shipped
a LIGHT background onto a dark site. The prompt only renders the structured
`design_intent.palette` block, so a site without it gets the "invent a palette" branch every
time.

> **CORRECTED 2026-08-09 (same day, on a re-check):** this section first said **0 of 12
> pinned**, measured at `sites.content_data` — the WRONG STORE. The proven pin pattern
> (`robot_hands/SQL_2026-07-17_r1b_…`) writes a `site_specs` row, `aspect='design_intent'`,
> and `webdesign-agent`'s own `read_site_specs` step is what consumes it. Measured there:

**9 of the 12 affected sites ARE pinned** [MEASURED 2026-08-09, at `site_specs`]. Most pins
were written by `domain-research-classifier` in normal operation (06-05 → 08-02), so a pin is
the platform's normal state. **Unpinned: `ai-agent-orchestration.com`, `finetuning.uk`,
`gaswholesalers.com`** — and gaswholesalers is the site with 6 of the 12 expected closures.

```sql
SELECT s.domain, bool_or(ss.aspect='design_intent' AND ss.is_current
       AND ss.data->'palette'->'reference_values' IS NOT NULL) AS pinned
FROM sites s LEFT JOIN site_specs ss ON ss.site_id=s.id
WHERE s.domain IN (…the 12…) GROUP BY 1;
```

**The one guard that exists is live but narrow.** `enforceLayoutScheme` (`bugs_closed/022`)
pod-greps 2 on v1.0.1270. It rejects a merged background that contradicts `layouts.scheme` —
so the worst case (light background onto a dark site) is caught. It does **not** constrain
accent, text or heading drift, which is most of what this lane measures.

### The options, costed. Do not pick one silently — this is wider than bug 122.

1. **Pin the 3 unpinned sites, then dispatch all 12.** The proven pattern
   (`robot_hands/SQL_2026-07-17_r1b_design_intent_palette_pin.sql` — the next run reproduced
   the pinned values exactly; supersede the current `design_intent` spec row, do not UPDATE
   it). Cost: 3 pins derived from each site's *current live* palette — measure the served
   stylesheet, not a DB copy. The 9 pinned sites can be dispatched as-is.
2. **Give `render_css_from_spec` a caller that is not the design LLM** — a minimal
   render-and-deploy path. Cleanest long-term and it makes this class of propagation cheap
   for ever. But it is a **new shared mechanism, so architecture-scope** under the 2026-07-29
   ruling: register it in the same commit, and expect the guardian seat to want it on its own
   merits rather than folded into a bug fix.
3. **Do nothing and leave the fix inert.** Honest, reversible, and currently the de-facto
   state — but then 338 is a source change nobody sees, which is precisely the shape
   `bugs_open/213` is about.

**My recommendation is 1 for this lane and 2 as a separate item** — and after the re-check,
1 is much cheaper than first stated: 3 pins, not 12. The churn risk on the 9 pinned sites is
the proven-contained case. It still wants an explicit yes, because a fleet-wide CSS
re-dispatch is not this lane's surface alone.

## 3. `bugs_open/212` — reframed. Read its §8 before acting on §1–§7

I filed 212 on 2026-08-06 and corrected it on 08-07. Three things changed:

- **The fix ranking is refuted.** On the motivating case the renderer's own
  contrast-checked value measures **1.71:1** against the component literal's **1.72:1**,
  and the muted slot **regresses 1.72 → 1.46**. §5's candidates 2 and 3 are a no-op and a
  slight worsening. `--color-primary-text` (`#0f0f0f`, already in gamesdesign's served
  CSS) gives **8.65:1** — that is candidate 5, and it is new to the file.
- **The class repair already exists.** `fix_forced_text_colours_action.go`'s
  `classifySectionPainting` derives what a template paints from its own CSS and
  `rewriteSectionDeclarationsInHTML` repoints `--section-*` to the on-colour family.
  `system-stats` matches its `paintPaletteBand` regex. **Nothing needs building** — the
  question is why it never ran, and §4 below answers that.
- **`buildSectionDefaults` emits nothing unless something is dark**
  (`color_util.go:185-187`) and its surface variant covers a hardcoded five-class list
  `.system-stats-section` is not in. Served blocks: gamesdesign 1, vonc 1, **idea.uk 0**.
  212's trap 4 was a guess; it is now confirmed at the source.

**Still open in 212, and still wanting a human:** the renderer knows exactly two grounds.
A component that paints its own has no correct token to ask for. Repointing components
one class at a time works; whether the renderer should learn about component-painted
grounds is an architecture question, not a bug fix.

## 4. NEW — `bugs_open/213`, and it is the actual blocker

gamesdesign's defect was **detected, described correctly, given a mechanical
`acceptance_test`, routed to the live fixer, and stamped `complete` in 3m17s with nothing
written** (proof: `content_components.updated_at` is 10.5 hours *earlier* than the item's
`created_at`). Not RFC_017's fail-open — nothing errored. Two producers file under one
`item_type` and the verifier implements only one of their predicates.

**7 of 9 `complete` items on that route are from the audit producer, all carrying an
`acceptance_test` nothing reads; every item that ever failed to close is from the other
producer, 6 of 6.** Do not read that as seven confirmed false-completes — read §4 of the
bug for what it does and does not establish. gamesdesign is the one confirmed at the
artefact.

**Where this bites this lane:** it is why 212 sat unrepaired while the queue reported the
site clean, and it will do the same to any repair we enqueue. Worth reading before
writing the propagation step in §2.

## 5. `090` — FOUR UNVERIFIABLE in a row, and how to read one

Run 3, `b6ab22d6-e49c-4b55-a9d9-dd026532a595` (the renderer's grounds): **UNVERIFIABLE,
iteration-capped** — three `bundle` artifacts, no verdict artifact, no
`metadata->>'decision'`. **Not the stale-index cause this time**: `symbols_unreadable`
was 0 on iterations 2 and 3. Its final hypothesis independently reached §3's conclusion
and named `warnUnusablePrimary`'s blind spot; corroboration, not proof.

Run 4, **`84c3da66-06c0-41a5-94dc-21fbf71260f0`** (the 213 mechanism): **terminated
`complete` at 08:48:02Z with five `bundle` artifacts and no `decision` on any of them —
UNVERIFIABLE, iteration-capped.** [CONFIRMED 2026-08-08.] On 08-07 this file recorded it
as "still diagnosing, not yet callable"; that was the right call at the time and the
answer is now in. **It did not refute anything in `bugs_open/213`** — it never reached a
verdict to refute with, so 213's root cause still rests on the first-hand evidence in its
§3, which is timestamp-based and does not depend on this run.

**Run 4 straddles the code-index fix, and that explains a number in its artifacts that
otherwise looks like progress.** Its `symbols_unreadable` fell **3 → 1 → 0 → 0 → 0** across
its five bundles. That is not the loop reading better each pass: **migration 332 repointed
the code index to the live working branch mid-run** (`code_symbols`: 5,754 symbols at
`087_towards_multiple_domains`, indexed 08:31Z, verified independently — the single-`ref`
row proves the old `086_experience_loop` pin is gone). **A `090` filed before ~08:30Z on
2026-08-07 was reading a stale index and one filed after was not** — worth knowing before
concluding anything from an UNVERIFIABLE dated that day, and it retires the standing
"confirm from code, not the index" workaround.

**So: four consecutive UNVERIFIABLE `090` runs in this lane** (`5853ee07`, `750e162e`,
`b6ab22d6`, `84c3da66`) — the first from a demonstrably wrong question, the last three
iteration-capped. **Do not read a fifth UNVERIFIABLE as evidence about your symptom.** The
standing lesson says UNVERIFIABLE means the question was wrong; four in a row — one of
them (run 3) with a *correct* hypothesis visible in its final bundle, and run 4 reaching
its cap with the index problem fixed under it — is evidence about the loop's iteration
budget instead. If you file another, read the last bundle's `## Hypothesis under test`
regardless of the verdict: on run 3 that section was the useful output, on run 4 it was
still my own symptom echoed back unrefined.

**This is worth someone's attention as its own item, and I have not filed it** — it is a
claim about the diagnosis loop, not about 122, and filing it properly needs the run
history across lanes, not just ours.

There is no `verdict` artifact kind and the outcome is not in `doc_notes`,
`site_work_items.spec` or `orchestration_states` — the query and how to read it are now
in `RUNBOOK_contrast_ink_slots.md` § "Reading a 090 verdict".

## 6. Verification target — and the baseline is now ageing

`BASELINE_2026-08-06_render_audit.txt` — 15 sites, 109 failures, complete. Grade **per
selector at the named ratio**, never by total; a falling count is content-dependent.
Expected close from the §2 migrations: 12 failures, plus `.stats-eyebrow` on vonc.
212's class is a **further ~24** and is not in that 12.

**[UNMEASURED as of 2026-08-08] The baseline is two days old and other lanes ship to
these same sites continuously** — the 08-07/08-08 log alone carries a voice rollout
across 23 pages, a placeholder-scan validator and several site lanes. A page re-rendered
since 08-06 for someone else's reason carries **every** change since it last rendered,
not just theirs, so a selector may have moved or a failure may have gone without us
touching it. **Re-run the baseline immediately before you grade the migrations, and diff
it against the banked one first** — if the before-state has drifted, an unchanged total
after your change is not evidence of anything. Do not skip this because the file says
"complete": it was complete on 08-06.

## 7. Advisory objections still open on the APPROVED verdict

Unchanged from `HANDOFF_2026-08-06b` §5 — six of them, `editquality` medium (the
propagation enqueue) being the one to act on. Approval means you may proceed; none were
withdrawn.

`Council-Reviewed: c4d9c841-3658-4742-85b5-961e062ecad2` still applies to further commits
in this lane.

## 8. Still no SUMMARY, still deliberately

The pages are unchanged, so the five headings would read "we measured, planned and built"
for the third time. The inflection to write one at is the first page measuring clean.

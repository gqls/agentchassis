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

## 2b. PROPAGATION — CSS half DONE all 11 live sites; page half enqueued [2026-08-10]

**Stylesheets: delivered and verified per site against the banked before-state.** All 11
live sites re-rendered via `webdesign-agent` (canary gamesdesign 08-09, batch overnight),
all verified at the SERVED file: changed, ink slots present, palette diffed slot-by-slot.
**lendzy.co.uk excluded** — origin down (Cloudflare 522), no before-state to grade against.

**The advisory-pin drift materialised, small: 3 of 11 runs adjusted exactly one core slot**
[MEASURED 2026-08-10]:

| site | slot | before → after |
|---|---|---|
| idea.uk | border | #C8BFA8 → #1A1816 (light→dark — visible; keep an eye) |
| dartsonline.com | background | #111520 → #0F1219 (imperceptible) |
| mortgagecalculator.co.uk | secondary | #334155 → #1e3a5f (moderate) |

No scheme flips. 8 of 11 held byte-identical on all nine core slots. This is the accepted
risk, quantified — and the measured drift rate for anyone weighing a future dispatch.

**Pages: 12 `page_rerender` items enqueued 2026-08-10** (`…_viz014_20260810`, reason
`section_data_resolved`, all pages pre-checked zero NULL `content_data`) across aao,
finetuning, idea.uk, leopardess, robot-hands, vonc. Gamesdesign's two were done with the
canary and verified at the served URLs.

**dartsonline needs NO page re-render and its expected closure may already be moot:**
`image-hover-card-grid` now has **zero placements fleet-wide** — the dartsonline lane
removed it since 08-08. If the failing element is gone, the baseline row closes by removal;
only the re-audit can say. **A placement table is a snapshot — re-run it before acting on
it**, this one lost a row in two days.

### NEXT, in order

1. **Confirm the 12 page items completed AND verify at served URLs** — `pages.url`, never a
   constructed filename (tools-index taught that: served at `/tools/index.html`).
2. **Re-audit and grade per selector**: `python3 scripts/render_audit.py <the baseline's 15
   urls> > after_$(date +%F).txt`, diff against `BASELINE_2026-08-06_render_audit.txt`
   selector-by-selector. Expected: the reachable closures gone (robot-hands 2, finetuning 2,
   gaswholesalers 6), dartsonline 1 gone-or-removed, **both `.stats-eyebrow` rows still
   failing** (§6 correction — unreachable by the shipped engine), 212's ~24 unchanged, and
   the three drifted slots checked for NEW failures they might have introduced.
3. **Write the render-audit cadence migration** (edit 8, the second half of the approved
   plan — still unwritten; `\d scheduled_tasks` first, weekly not daily).
4. Then 122 is closable to the extent the engine allows, and what remains is 212's
   architecture question (two mechanisms now measured blind to component-painted grounds).

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

> **CORRECTED 2026-08-09 (canary review): expected close is 10, NOT 12.** The two
> `.stats-eyebrow` closures (gamesdesign, vonc) are **unreachable by the shipped engine**:
> `buildLegibleInkDefaults` computes the ink companions against `{background, surface}`
> only (`palette_specialised_slots.go` — `pageGrounds`), and the eyebrow sits on the
> component-painted **primary section fill**. On gamesdesign, accent is 12.46:1 on the page
> ground, so `legibleInkFor` correctly returned it unchanged — and the served post-fix
> eyebrow measures **1.44:1 on its real ground, byte-identical to the baseline failure**
> [MEASURED]. vonc is the same mechanism [PREDICTED — verify when its re-render lands].
> This is the second concrete instance of `bugs_open/212` §8's open architecture question
> (component-painted grounds are invisible to the renderer); recorded there.

Expected close from the §2 migrations: **10** — dartsonline 1, robot-hands 2, finetuning 2,
gaswholesalers 6 (minus 1 if any single audit row spans both halves — grade per selector).
212's class is a **further ~24** and is not in that 10.

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

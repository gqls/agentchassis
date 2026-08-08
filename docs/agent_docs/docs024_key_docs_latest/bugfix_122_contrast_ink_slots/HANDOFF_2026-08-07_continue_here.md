# HANDOFF — bug 122, contrast / ink slots. START HERE.

**Written 2026-08-07, last updated 2026-08-08 (§0).** Supersedes
`HANDOFF_2026-08-06b_continue_here.md`. **That file's "next action 1" is now DONE** and
its §"What NOT to do" list still stands in full — it is not repeated here, so read it
there before touching code. Everything else below is new or changed.

## The one-paragraph state

The **engine half is live and has stayed live across three rolls**: v1.0.1266 carries
VIZ-014, pod-proven on both replicas (§1). No lane in this workstream rolled any of them. The
**config half — the part that changes a visitor's page — is WRITTEN BUT NOT APPLIED**:
migration `338`, committed `5e77607cf`, plus a second migration still unwritten (§2). Separately, picking up
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

Then pick up at **§2** — migration `338` is written and committed but **not applied**,
and that plus one more migration is the entire remaining job for bug 122.

## 0. What changed on 2026-08-08, before you trust anything below

Re-measured at the start of the 08-08 session. **Four of this file's own figures went stale
within a day**, one of them twice — that rate is the point, not the individual numbers.

| what | was (08-07) | is (08-08) | so |
|---|---|---|---|
| chassis image | v1.0.1262 | **v1.0.1266** (08-08 pm, re-proved again) | §1 — a new image is a new fact |
| next free migration number | 324 | **338, now used** — 324/325/334/336/337 all taken | §2 |
| `090` run 4 | still `diagnosing` | **`complete`, UNVERIFIABLE** | §5 |
| the config half | not written | **338 written + committed, NOT applied** | §2 |

Nothing has touched `bugs_open/212`, `bugs_open/213` or this directory since the 08-07
commits (`git log b938b54d8..HEAD -- …` is empty), so §3–§4 stand as written.

## 1. DONE — the engine is live, pod-proven [re-proved 2026-08-08]

Both replicas of `agent-chassis`, image `docker.io/aqls/agent-chassis:v1.0.1266`:

| symbol | count | role |
|---|---|---|
| `buildLegibleInkDefaults` | 4 | the new emitter |
| `legibleInkFor` | 3 | " |
| `worstRatioAgainst` | 2 | " |
| `fillDarkSchemeSpecialisedSlots` | 4 | positive control |
| `zzzInventedControlXyz` | 0 | negative control |

Identical counts on v1.0.1262, v1.0.1264 and v1.0.1266 — three rolls, none of them ours. Nothing downstream of a
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

## 2. NEXT — migration 338 is WRITTEN and COMMITTED but NOT APPLIED

**`docs/agent_docs/sql_for_agents/338_components_opt_into_legible_ink_slots.sql`**
(`5e77607cf`, 2026-08-08). This is the config half of 122 — the only thing left that
changes a visitor's page. **Read its header block before running it**: it explains, with
the measurements, why it does not match the approved sketch.

### Why 338 diverges from the approved plan, and why that was necessary

The sketch used a bare global `replace()` per component, gated on a **row** count.
`replace()` is global **within the string**, so a row-count gate cannot see
over-application. Measured against the live rows on 08-08:

| component | needle occurrences | intended targets |
|---|---|---|
| `case-studies-grid` | 2 | 2 — sketch correct |
| `system-stats` | **5** | 1 (`.stats-eyebrow`) |
| `image-hover-card-grid` | **2** | 1 (`__eyebrow`) |
| `tool-list` | **6** | 2 (`.tl-eyebrow`, `.tl-card-link`) |

The extras are not harmless. `tool-list`'s four include
`.tl-card-icon { background: var(--color-primary); }` and
`.tl-cta-btn { background: var(--color-primary); }` — the approved form would have
**repointed two backgrounds to an ink colour**, the exact inversion this bug is about.
`system-stats` has two backgrounds and an outline among its five; `image-hover-card-grid`'s
second is an outline. So every needle in 338 is rule-scoped, every block asserts its
expected occurrence count, and the three risky ones carry a **negative control** asserting
the non-targets survived.

**The layouts half did not match either.** The sketch's `a { color: var(--color-accent); }`
one-liner exists in **0 of 18** layouts. Four of the five named layouts use a multi-line
rule; `tool-portal-light` uses a different single-line form. Two separate blocks in 338.

### What is left to do, in order

1. **Induce every `RAISE` against a scratch copy before applying.** Not done. A verify
   block that cannot fail is precisely the trap 338's own comments warn about, and
   asserting that it is careful is not the same as proving it fires. Clone the four
   component rows and five layout rows into a scratch table, break each needle in turn,
   confirm each `RAISE EXCEPTION` actually aborts.
2. **Re-check the number.** `338` was free at write time (`337` highest on disk and in
   `schema_migrations`). Re-check immediately before `--apply` — this file has been wrong
   about a migration number twice already.
   ```bash
   ls docs/agent_docs/sql_for_agents/ | grep -oE '^[0-9]{3}' | sort -n | tail -3
   kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
     -t -A -c "SELECT filename FROM schema_migrations ORDER BY filename DESC LIMIT 3;"
   ```
3. **Take the two `\copy` backups** named in 338's header, outside the transaction.
4. **Apply**, scoped to this file. `328` and `330` are `_HOLD`-renamed so a blanket
   `--apply` cannot ship them — do not sweep the directory.
5. **Enqueue the propagation.** The UPDATE changes the SOURCE; a visitor sees nothing
   until the pages re-render. This is the still-open `editquality` MEDIUM objection, and
   **listing the rows is not enough — the approved plan only listed them and the seat was
   right that this is insufficient.** The query is at the foot of 338. §4 is why a
   `complete` work item is not evidence this happened.
6. **Verify at the served page** — never row counts, never a work-item status. And re-run
   the baseline first: see §6, it has drifted risk.

### Still unwritten: the second migration

The **render-audit cadence** (`scheduled_tasks` row, submission edit 8) is not written.
One row; `\d scheduled_tasks` first — it is `interval_seconds`, not a cron string, plus
`target_topic`, `input_data`, `pre_query`, `fire_message`. Weekly, not daily: findings
dedupe on `contrast_failure:<page-path>#<selector>` and a falling count is
content-dependent, so daily multiplies noise without signal. The whole detection chain has
been live since v1.0.1257 with **0 `scheduled_tasks` rows targeting `render-audit-agent`**,
enabled or disabled — it has run exactly once, by hand.

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

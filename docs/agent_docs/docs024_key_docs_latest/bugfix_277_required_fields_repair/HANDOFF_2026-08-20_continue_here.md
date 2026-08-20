# HANDOFF — 2026-08-20, morning. `497`+`498` applied, `277`'s clause-1 blocker has CHANGED SHAPE, and the 7-row population this lane tracked for two days is gone

**Supersedes `HANDOFF_2026-08-19b_continue_here.md`.** Read this from disk, then `NOTES_required_fields_repair.md`
**from the bottom** (two new entries today, `~08:05Z` and `~09:00Z`). 08-19b stays authoritative for
`301`'s council trail and the `v1.0.1316` probe; **§1 below names what is dead in it**, and one of
those is its recommended first action.

---

## 0. STATE TABLE

| bug | state | what blocks the close |
|---|---|---|
| **`bugs_open/277`** | router live, approved | **clause 1 — one worked example REPAIRED. Blocker has CHANGED SHAPE (§4): it was "owned pages have no route at all", it is now "a route exists at 36/37, and one CODE question remains".** `no_content_data` (27 of 30 parked) untouched by any of this |
| **`bugs_open/083`** | fix live + artefact-proven | the door soak, **~2026-08-25**. ⚠ `479`'s reclaim arm **has still never fired** — the close must SAY that. **New knock-on in §5: the escalation clock is NOT a backstop for owned-page rows** |
| `bugs_open/300` | fix live, council APPROVED | demand, not a defect. Leave it. **NB it is no longer the "unassigned" item type — `497` fixed that** |
| `bugs_open/314` | filed, unfixed | owner's call. **Corroborated again today**: `497`/`498` are config migrations under a `docs/` path, so the gate refused them client-side. Third lane |
| `bugs_open/333` | filed 08-19 22:02 by the 301 lane | theirs. **CONTRIB filed today** (§5) — their loop ran end to end at n=7 |
| `301` | **CLOSED** | done — the mid-move 08-19b warned about completed |

---

## 1. WHAT IS DEAD IN THE PREVIOUS HANDOFF — and one of them is its §8 item 1

1. **§6's clock table is VOID for `literal_markdown`, and §8's first action no longer exists.**
   08-19b said *"the 7 `literal_markdown` rows before the 08-21 12:57 tick"* was the only item with
   a clock. **Those 7 rows are terminal.** [MEASURED 08-20 08:11Z] between **07:20:42Z and
   07:23:58Z** every one was dispatched, refused by `OWNED_PAGE_GUARD`, and closed **`wont_fix`**.
   **The 08-21 escalation will not fire.** Why they moved: the pair rose **8.1% → 44%** overnight
   (16 completions across 3 sites in the 07:00Z hour), crossed the floor, and the promoter fed its
   owned-page rows straight into the guard.
2. **§5's remaining general form needs one more narrowing.** *"Re-routing a producer strands its
   backlog"* still holds. But the backlog here was not stranded for ever — it was **drained into a
   terminal status by a door nobody was watching**. See §5.
3. **§0's "nothing repairs `no_content_data`" is unchanged and still true.** Today's finding does
   **not** touch it. Do not let §4's good news bleed into that column.

---

## 2. WHAT WAS APPLIED TODAY — two migrations, both live, both verified at the live column

| mig | what | live md5 after |
|---|---|---|
| `497` | the `owners` map pointed at **three dead or wrong destinations** | `406bd757…` len 13106 |
| `498` | de-volatilises `497`'s own `literal_markdown` string, whose figures went stale in 12h | **`ddd0c894…` len 13746** ← current |

Commits `a888224f0`, `8b4def77c`. Both use `479`'s surgical-anchor pattern: anchors asserted unique,
md5-guarded, forward and reverse from the **same variables**, and a **reverse-replacement control**
that returns the body to its exact pre-image — **mutation-proven** in both files (damage one byte of
unrelated `what_to_do` prose and it fails). Both round-tripped with their `_ROLLBACK.sql`, and the
corrected query was **EXECUTED** rolled back (`CREATE TEMP TABLE t AS <pre_query>`), because the
bytes being right does not prove the SQL still runs.

**Owner approved applying `497` ahead of today's tick.** The argument that made it a same-day change:
the `owner` value is **stamped into the row ONCE**, at escalation time, and never revisited — so a
later fix cannot repair rows already escalated — while the change itself cannot move a row.

> ⚠ **Do NOT re-derive `497`'s guard md5 from these files if you are writing `499`.** The live text
> is `ddd0c894…` (post-`498`). Read the column.

---

## 3. TODAY'S TICK — what it will do, previewed rolled-back against live data

**`escalated=4, reclaimed=0, watching=6`** at **~12:58Z** (task `enabled`, `interval_seconds 86400`,
`last_triggered_at 2026-08-19 12:58:16Z`). All 4 are `placeholder_contact → page-build-handler`,
each stamped `days_waiting=4` and the corrected owner string naming **`083` + this lane**.

⚠ **Four, not the three 08-19b predicted** — a 4th row joined the pair overnight. Re-derive at the
tick; the held set is not stable between a query and a tick.

**Remaining clocks** [MEASURED 08-20 08:11Z, re-derived from the live `pre_query`'s own `classified` CTE]:

| pair | kind | rows | oldest | escalates |
|---|---|---|---|---|
| `placeholder_contact → page-build-handler` | canary | **4** | 08-16 19:17:45 | **today ~12:58Z** |
| `missing_conversion_path → content-gap-planner` | canary | 1 | 08-17 22:21:46 | 08-21 ⚠ `bugs_open/255` owns it — do not canary |
| `dead_fragment_link → page-build-handler` | canary | 1 | 08-18 01:38:47 | 08-21 |

`literal_markdown` is **absent** — no longer in the held set at all.

---

## 4. THE FINDING — "an owned page has NO route at all" is REFUTED, and the route was in this task's own prose

> # ⚠⚠ CORRECTED ~10:15Z, SAME DAY — THE ROUTE BELOW DOES NOT APPLY TO THIS POPULATION
>
> **The 36/1 measurement stands. The target does not.** Answering §7's own item 1 ("can a producer
> file a `section_edit`?") refuted this section within two hours of writing it.
>
> **All 7 findings are `source: rendered_html`** (`pattern: code_span`, `slot: ported-page` —
> backticked code tokens like `` `fetch()` `` in ported prose). The `ported-page` component's
> `content_data` is 215 bytes of provenance metadata; its template's only field is `{{.body}}`,
> which is **not a key**. So `section_edit` + `strip_literal_markdown` strips a map with no prose,
> and `473`'s rerender regenerates from nothing — **both inapplicable BY CONSTRUCTION.**
> Measured against production's own engine: the owned payload renders **0 visible characters**,
> `err=<nil>`; a generic control on the same template renders 6,568 bytes of prose.
>
> **Fleet census, and it reframes three days of this lane:** the component has 115 instances, 100
> missing `body` — **all 100 are `owned`, all 15 that have it are `generic`.** Ownership and
> un-regenerability *coincide*, so the ownership guard takes the blame while the operative property
> is whether the content is reachable from `content_data` at all.
>
> **NOT "100 pages one edit from blanked"** — `enforceSingleSlotFloors` measures VISIBLE text and
> refuses at zero, leaving the component standing. A third refusal mode, not damage.
>
> **Clause 1's blocker is RETRACTED to where it was**: no route, and repairing these needs an
> HTML-level transform on `rendered_html` that nothing does. Full evidence: `bugs_open/277` §5,
> NOTES `~10:15Z`, `LANDMINES.md` *"A component whose `content_data` CANNOT REPRODUCE…"*, and
> migration `499`, which replaces the target below with a TEST.

**This is the day's result and it changes `277`'s blocker rather than closing it.**

`466`'s `what_to_do` — the prose our own escalation hands a human — already names the route, quoting
`bugs_closed/295` fix candidate 3 at a conservative *"18 completes"*. **[MEASURED 08-20, live +
archive] it is better than that:**

| `section_edit → section-editor` | complete | failed | total |
|---|---|---|---|
| **`rebuild_policy='owned'`** | **36** | **1** | **39** |
| `generic` | 53 | 4 | 57 |

Against the generic repair on the same axis — `literal_markdown → page-rerender`: **8 complete on
`generic`, 1 failed on the single `owned` page it tried.** One route is refused by design; the other
is how this estate already edits owned pages.

**The severe landmine was checked and its precondition is ABSENT.** `LANDMINES.md` warns a
`section_edit` on a per-site **tool fork** with `{{.field}}` template copy and `content_data='{}'`
re-renders every text node to **EMPTY** while every floor passes — and 6 of our 7 pages are `tool-*`.
It needs **both** halves: the tool forks have `content_data='{}'` and **zero** template fields, so
nothing can fail to fill. **And the defect is not in the fork** — it is in the `ported-page` slot, a
`component_level='section'` component whose `content_data` **is** populated, i.e. the 36/1 target.
`466`'s own caveat fits too: `section_edit` **REWRITES** and cannot **ADD**; literal asterisks in
existing prose are a rewrite.

### WHAT IS LEFT IS ONE CODE QUESTION — this is the next job

**Can a producer file a `section_edit` item for a `literal_markdown`-shaped finding at all?** What
`spec` / `field_updates` would it carry, and which `page_component_id` does it target (here:
`ported-page`, not the tool slot)? Read `section_editor_actions.go` and the producers.
**Three landmines already written down apply to whoever answers:**
- the `field_updates` merge is **per-field and reverts intervening edits** (a stale proposal restores
  older text and reports success);
- `apply_section_edit` writes `rendered_html` with **no content validation**;
- it **cannot ADD** a component.

⚠ **And read the two WRONG_CALLS entries filed today before you start** — both are about checks I ran
*because* I was being careful, and neither could have returned bad news.

---

## 5. CONTRIB filed into `bugs_open/333` — their loop ran end to end, and it is invisible in the ledger

Both properties were **already known** in `bugs_closed/301` and `333` (prior art checked first, and
it changed what I wrote) — so this is the **measured instance**, not a new claim:

released → refused → **`wont_fix`** → re-filed by the detector. `wont_fix` is excluded from **both**
sides of the promoter rule *and* from `idx_swi_dedup`. So **a healthy ratio on a MIXED-POLICY pair is
not evidence its owned-page rows are being repaired — it is evidence their failures stopped
counting.** That argues for 333's `writeWorkItem`-**door** check over anything downstream.

**Knock-on for `083`, and it belongs in the close-out:** the escalation this residual was heading for
**will not fire**. **A row that can only be refused reaches `wont_fix` faster than it reaches a
human, and the faster path is the silent one.** So the escalation clock is not a backstop for the
owned-page class.

---

## 6. STILL OWED, unchanged

- **§7a: the `copy_edit_proposed` exclusion in the promoter's `pre_query`** (owner decision D2,
  2026-08-12). Still not done, still deliberately: unlike `497`/`498` it **changes which rows are
  dispatched**, so it is not the behaviourally-nil class and it goes past the owner.
- **Tell the `diagnosis_guardian` seat its `error_step` discipline is INVERTED** (08-19b §3). Not
  done. It will mis-fire on every correct submission until someone tells them.
- **~2026-08-25: close `083`**, both paths on the commit, verified at HEAD with `git ls-tree`, and
  **stating that `479`'s reclaim arm has never fired.**
- **`277`'s `no_content_data` half** — different agent; `473`'s deterministic route does not cover it.

---

## 7. WHAT I WOULD DO NEXT

1. ~~**§4's code question.**~~ **DONE — and it refuted §4.** See §4's correction box and
   `bugs_open/277` §5. **The live successor question is for the OWNER, not a session:** repairing
   these 7 needs an **HTML-level transform on `rendered_html`** (`` `x` `` → `<code>x</code>`) that
   no route performs. They are 7 findings of the mildest pattern — backticks in developer-tool
   prose, not broken pages — so "build a new repair shape" is not an obvious yes. **Put the choice
   to the owner rather than starting it.**
   ⚠ **And do not re-derive the route from the 36/1 figure.** It is real, it is kept in the config
   string on purpose, and it is aimed at a different property. The gate is
   *"can `content_data` reproduce `rendered_html` for the target component?"* — two lines of SQL,
   in the `LANDMINES.md` entry.
2. **Watch today's 12:58Z tick land** and confirm the 4 rows carry the corrected owner
   (`result->'held_pair_escalation'->>'owner'`). First real test of `497` in production.
3. **The `diagnosis_guardian` message** (§6) — cheap, and it degrades every future submission.
4. **`083` on ~08-25.**

---

## 8. Session-start checklist
`git log --oneline -10` · re-read this from disk · `scripts/who-owns.py` **by slug** for `277`,
`083`, `300`, `314` · **`distinct digests`** on the chassis image (and note: if the digest matches
the one a probe already ran against, the probe transfers — you do not need to re-run it) · re-derive
§3's clocks from the live `pre_query` **and expect the population to have moved** · then §7.

# HANDOFF — 2026-08-20. `497`/`498`/`499` applied and `497` PROVEN at the artefact; `277`'s clause 1 is back where it was, for a better-understood reason; the day's real finding is that OWNERSHIP WAS THE WRONG DIAGNOSIS

**Supersedes `HANDOFF_2026-08-19b_continue_here.md`.** Read this from disk, then
`NOTES_required_fields_repair.md` **from the bottom** (three new entries today, `~08:05Z`, `~09:00Z`
and `~10:15Z`). 08-19b stays authoritative for `301`'s council trail; **§1 names what is dead in it**,
including its recommended first action.

> ### ⚠ READ THIS BEFORE ANY SECTION BELOW — the day corrected itself twice
> Sections **2, 3 and 4 each carry a correction box**, and in every case the box is the current
> state and the prose beneath it is what I believed earlier. **§4's headline conclusion is
> RETRACTED.** If you read only one thing: *ownership was never why these pages could not be
> repaired* — their content is not regenerable, and on this component those are the same 100 pages.
> **Nothing in this lane is repaired yet.**
>
> **Written at 14:30Z:** chassis is **`v1.0.1319`** (`sha256:9be1aa50…`, rolled 10:18Z, digest
> genuinely changed). Live escalation `pre_query` md5 is **`0d72b423…` len 13938** (post-`499`).
> Both are deploy facts with a shelf life of hours — **re-read, do not quote.**

---

## 0. STATE TABLE

| bug | state | what blocks the close |
|---|---|---|
| **`bugs_open/277`** | router live, approved | **clause 1 — one worked example REPAIRED, and it has NOT narrowed.** ~~"a route exists at 36/37, one CODE question remains"~~ **RETRACTED (§4 box, `277` §5)**: the findings are `source: rendered_html`, and no `content_data` route reaches them. Needs an **owner decision** (§7). `no_content_data` (27 of 30 parked) untouched by all of this |
| **`bugs_open/083`** | fix live + artefact-proven | the door soak, **~2026-08-25**. ⚠ `479`'s reclaim arm **has still never fired** — the close must SAY that. **New knock-on in §5: the escalation clock is NOT a backstop for owned-page rows** |
| `bugs_open/300` | fix live, council APPROVED | demand, not a defect. Leave it. **NB it is no longer the "unassigned" item type — `497` fixed that** |
| `bugs_open/314` | filed, unfixed | owner's call. **Corroborated again today**: `497`/`498`/`499` are config migrations under a `docs/` path, so the gate refused all three client-side. Three live-config changes, zero review available. Third lane to hit it |
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

> ⚠ **Do NOT re-derive a guard md5 from these files.** **`499` was applied later the same morning**
> (§4's correction box) — the live text is **`0d72b423…` len 13938**. Read the column, always.

### 2b. UPDATED 2026-08-20 ~14:30Z — a THIRD migration, and the chassis rolled again

| mig | what | live md5 after |
|---|---|---|
| `499` | `498`'s candidate route is **inapplicable to this population** — replaces a TARGET with a TEST | **`0d72b423…` len 13938** ← **current** |

Commit `3a41595c3`. Same controls, mutation-proven, round-tripped.
**All three migrations are DB config with NO binary dependency** — none of them needed a roll, and
none is affected by one.

**⚠ The chassis rolled at 10:18Z: `v1.0.1319`, `distinct digests: 1`, `sha256:9be1aa50…`.**
The digest **CHANGED** from `v1.0.1316`'s `sha256:2d0d3def…`, so this is genuinely new bytes and
**not** the same-tag-cached-image trap. §2's 5-needle probe describes `2d0d3def…` and **no longer
describes what is running** — do not quote it.

**Re-probed on `v1.0.1319` [MEASURED 14:35Z] — and note this probes the CAPABILITY, not the commit,
which is the more useful question:** every symbol §4/§5's conclusions rest on is in the running
binary.

| needle | hits | role |
|---|---|---|
| `executeGoTemplate` | 12 | the real render path (`call_agent.go:1171`) |
| `missingkey=zero` | 1 | the option that makes a missing `{{.body}}` render EMPTY |
| `StripLiteralMarkdownFromContentData` | 4 | the strip `473`/`474` gate |
| `enforceSingleSlotFloors` | 2 | the floor that refuses a hollowing |
| `OWNED_PAGE_GUARD` | 3 | **positive control** — the probe works |
| `ZZQQ_NEEDLE_THAT_MUST_NOT_EXIST` | **absent** | **negative control** — it discriminates |

⚠ `missingkey=zero` hits **once** though it appears twice in source — Go dedupes identical string
constants. That is consistent, not contradictory; **do not read a binary hit count as a call-site
count.** ⚠ And digest identity proves every pod runs these bytes; it does **not** prove any pod
executes this code. ⚠ And do not quote a pod count: it read **8** at
14:29Z against **94** at 08:00Z, because the `Job`-owned per-work-item pods spawn and age out
continuously (4 `ReplicaSet` both times).

---

## 3. TODAY'S TICK — **IT FIRED, AND `497` IS PROVEN IN PRODUCTION**

> ## ✅ VERIFIED 2026-08-20 14:29Z — this is §7 item 2, done
>
> **`last_triggered_at = 2026-08-20 12:58:33.590523+00`.** It escalated **exactly the 4 rows the
> rolled-back preview predicted**, each stamped `days_waiting=4` and each carrying the **corrected**
> owner string naming `083` + this lane — *not* the dead `bugs_open/201` pointer it would have
> carried before `497`:
>
> | page | item_type | escalated_at |
> |---|---|---|
> | `tool-seo-injector` | `placeholder_contact` | `2026-08-20T12:58:33Z` |
> | `guide-creating-ideas` | `placeholder_contact` | `2026-08-20T12:58:33Z` |
> | `tool-insight-injector` | `placeholder_contact` | `2026-08-20T12:58:33Z` |
> | `tool-privacy-redactor` | `placeholder_contact` | `2026-08-20T12:58:33Z` |
>
> **This is the artefact, not the status** — the test was never "the migration applied" (its own
> NOTICE said that at 08:00Z) but "the mechanism wrote the right sentence into a real row". It did.
> Re-check any future claim the same way: `result->'held_pair_escalation'->>'owner'`, not the md5.
>
> ⚠ **These 4 are now a live human-queue item and they are `083`'s canary decision, not this lane's
> repair question.** Do not act on them without reading `bugs_open/083` — and note three of the four
> are `tool-*` pages, so the owned-page question may apply to them too. **Unmeasured; do not assume
> it does or does not.**

*(Written before the tick, and it held:)* **`escalated=4, reclaimed=0, watching=6`**, previewed
rolled-back against live data. All 4 `placeholder_contact → page-build-handler`.

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
2. ~~**Watch today's 12:58Z tick land.**~~ **DONE — `497` is PROVEN in production**, see §3's box:
   fired 12:58:33Z, 4 rows, each carrying the corrected pointer. Nothing further owed on `497`.
3. **The `diagnosis_guardian` message** (§6) — cheap, and it degrades every future submission.
   **This is the largest un-started item that a session can do alone.**
4. **`083` on ~08-25**, and see §3's note: its canary decision now has **4 escalated rows** sitting
   in the human queue as of today, three of them `tool-*`.

### THE ONE THING WAITING ON THE OWNER

**Do these 7 `code_span`-in-`rendered_html` findings get a repair route at all?** Building one means
a transform that edits `rendered_html` directly — a shape this estate does not have. The cost is
real; the defect is 7 instances of backticks around code tokens (`` `fetch()` ``) on developer-tool
pages, which many readers would take as deliberate. **Neither answer is obviously right, which is
why it is his and not a session's.** Everything needed to decide is in `bugs_open/277` §5.

---

## 8. Session-start checklist
`git log --oneline -10` · re-read this from disk · `scripts/who-owns.py` **by slug** for `277`,
`083`, `300`, `314` · **`distinct digests`** on the chassis image (and note: if the digest matches
the one a probe already ran against, the probe transfers — you do not need to re-run it) · re-derive
§3's clocks from the live `pre_query` **and expect the population to have moved** · then §7.

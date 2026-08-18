# HANDOFF — 2026-08-18c, fresh chat starts here: six owner decisions taken, and the path to each

**Supersedes `HANDOFF_2026-08-18b_continue_here.md`** for the 083 half; that file's account of the
archiver diagnosis (migrations `465`/`466`) stands and is not repeated here. Everything below was
measured 2026-08-18 evening against `kafka-scheduler:v1.0.1309` unless dated otherwise. **Read this
from disk, then `NOTES_required_fields_repair.md` from the bottom.**

---

## 0. State: 083 is DONE except a soak; the work has moved downstream of it

| thing | state | re-verify with |
|---|---|---|
| `cmd/scheduler` pre_query logging | **LIVE + VERIFIED** on `v1.0.1309`, pod's own `build provenance` `git_commit=f0117fb8b`, `03012d862` an ancestor | `kubectl -n ai-persona-system logs -l app=kafka-scheduler --tail=3000 \| grep detected-item-promoter` — the line must carry `pre_query_result` |
| `detected-item-promoter` (SCH-026) | live, 900s, reads live+archive (`465`), suppresses idle ticks | `SELECT last_triggered_at, pre_query LIKE '%archive%' FROM scheduled_tasks WHERE name='detected-item-promoter'` |
| `held-pair-canary-escalation` (`453`+`466`) | live, daily, now also escalates floor-held pairs | same query, that name |
| held pile | **15 rows / 4 pairs**, all correctly held | the log line's `held` field |
| council `8dc58e2a` (held visibility) | **APPROVED** round 1, 4 advisories all answered | `bugs_open/083` §7 |
| `bugs_open/083` | **fix complete and artefact-proven**; open only pending the door soak | §5 below |

**Verification habit this lane keeps getting value from:** a control that cannot come out false is
not a control. Two instances in two days — a git ancestry "control" using a commit that predated the
build, and a pair-health query keyed on a column `failed` rows never carry. Both are in
`WRONG_CALLS.md`.

---

## 1. OWNER DECISIONS, 2026-08-18 — all six taken

| # | decision | ruling |
|---|---|---|
| 1 | what an owned-page refusal records | **do not switch the handler off for this — write something other than `failed`**; and *"address why the tool fix isn't working properly"* |
| 2 | the one-way escalation door | **fix the door** |
| 3 | canary `dead_fragment_link` / `missing_conversion_path` | **leave both** |
| 4 | `bugs_open/300` fix candidate | as recommended — resolve by `(page_id, slot_name)`, id as tiebreak |
| 5 | closing `083` | as recommended — close after the doors soak (~2026-08-25) |
| 6 | council gate cannot review config | as recommended — raise as its own item |

---

## 2. DECISION 1 — the diagnosis the owner asked for, then the two-tier fix

### 2a. Why the tool repair path "isn't working", answered

**It is not missing and it is not broken.** `section-editor` / `apply_section_edit` is the estate's
best-performing repair path on owned pages — **220 complete / 5 failed = 98%**. On owned pages
generally, `component-template-fixer` is 150/0, `page-rerender` 3754/89, `tool-generator` 14/0.

**The boundary is the KIND OF REPAIR, not the handler and not the page.** `page-build-handler`
succeeds on owned pages 74 times — for types that ADD content (`content_rewrite` 30/52,
`needs_content_page` 18/15, `empty_section` 16/2, `link_resolution_rebuild` 6/0, `needs_page` 4/0).
It is refused **0 for 39** for the five types that MODIFY existing content — `literal_markdown` 0/16,
`phantom_internal_link` 0/14, `placeholder_contact` 0/4, `empty_internal_href` 0/4, `tone_shift` 0/1
— because modifying is exactly what *"a generic section save would clobber it"* means.

**And the reason it cannot be fixed by re-pointing `handler_agent`** — this is the real finding.
`apply_section_edit` is content-first: its input is `{edit_type, page_component_id, field_updates}`
and **`field_updates` carries the corrected values**. The five detectors report that a defect exists,
never what the corrected content is: `literal_markdown` reports that asterisks reached the page, not
the de-asterisked text. **Nothing converts a detector's finding into an editor's edit.** That gap —
not a missing mechanism — is why tool/widget pages do not get repaired.

### 2b. Tier 1 — the status change (small, reversible, protects a fleet-wide gate)

Refuse to **`wont_fix`** with the reason, not to `failed`. Checked against the live gate:

- the promoter's floor counts `failed` and **never mentions `wont_fix`**, so the refusal leaves both
  numerator and denominator — the pair reads *never tested here*, which is the truth;
- `idx_swi_dedup` **excludes** `wont_fix`, so the dedup key is released and the finding re-raises
  naturally once routing exists.

No new status vocabulary, no change to the gate. It protects `phantom_internal_link`'s
**69%-on-generic-pages** repair path from being switched off by refusals it was never responsible
for. Site: the ownership guard in `save_page_sections` (`platform/orchestration/actions/
save_page_sections_action.go`). **Coordinate with `bugs_open/301`, whose author is already
reordering that guard — same function, same week.**

### 2c. Tier 2 — the routing, which is a design piece and not a patch

Someone must build the step that turns a finding into a `field_updates` payload, per defect type,
then route the five types by `pages.rebuild_policy`. Cheapest credible shape: a small
`compute_section_edit` action that takes the finding + the current `content_data` and returns the
corrected fields — deterministic for `literal_markdown` (strip markdown) and `empty_internal_href`,
LLM-assisted for `tone_shift`. **Architecture-scope** (a new shared repair route), so RFC before code.
~134 non-terminal findings are queued behind it on owned pages today.

---

## 3. DECISION 2 — fix the one-way door

`453`/`466` move a held row `detected → needs_human_review`; the promoter only ever selects
`detected` and nothing moves it back, so a pair that later becomes known-good never reclaims its
findings. `465` proved that "later becomes known-good" is real, not hypothetical — it released a pair
holding **9 lifetime successes** that had been held as "never completed".

**Shape:** a new numbered migration, on `458`'s pattern (guarded on a verbatim pre-image, separate
`_ROLLBACK.sql`), adding a reclaim arm to the daily escalation task — return rows to `detected`
where `resolution_path='auto:held_pair_escalated'` **and** their pair now passes the known-good rule
(reading live+archive, per `465`). Reversal SQL already exists in `466`'s rollback header; this makes
it automatic and conditional rather than manual.

**Controls it needs:** a positive one (a pair that has since qualified must be reclaimed) and a
negative one (a pair still held must NOT be), or the arm cannot be shown to discriminate. Today's
population makes the negative control easy and the positive one may need waiting for a real
qualification — say so rather than asserting it fires.

**Clock:** nothing is escalated yet (`result ? 'held_pair_escalation'` = 0 across all 15). The
3 `placeholder_contact` rows (created 08-16) cross the 3-day limit on **2026-08-19**; the 10
`literal_markdown` rows on **2026-08-20**. Building the reclaim arm before then avoids a manual
un-escalation later, but nothing breaks if it lands after.

---

## 4. DECISION 3 — both canaries stay unrun, and why that is written down

- `missing_conversion_path → content-gap-planner`: **`bugs_open/255`** (diagnosis loop CONFIRMED)
  records that the type is routed at a handler that cannot read its spec. A canary would fail for a
  documented routing defect and record it as handler incompetence.
- `dead_fragment_link → page-build-handler`: the defect is **real and verified at the served page**
  (`vetcomparison.uk/tools/pet-treatment-cost-estimator/index.html` carries `href="#directory"` with
  no `id="directory"`; negative control absent). Not run because the handler regenerates the page,
  the page is a live tool, and — the trap — **its `rebuild_policy` is `generic`, not `owned`**, so
  the guard that refuses tool rebuilds would NOT protect it. Under Decision 1's Tier 2 this finding
  is one of the five types that needs the editor route anyway.

Both are `LANDMINES.md` sub-bullets now: grep `/bugs_open/` for the item type **and** its handler
before promoting, and read what the handler will *do* for this `fix_type` — "the last canary was
fine" does not carry across.

> ⚠ **Follow-on worth its own look:** a page whose name and URL are `tool-…` carrying
> `rebuild_policy='generic'` looks like a data defect. If tool pages are inconsistently marked, the
> ownership guard protects some and not others, and Decision 1's Tier 2 routing would inherit that.
> Not filed — nobody has counted how many.

---

## 5. DECISIONS 4-6 — the smaller three

**4 — `bugs_open/300`.** Fix candidate 1: `fixPageComponentStatus` resolves the component by
`(page_id, slot_name)` and uses `spec.page_component_id` only as a tiebreak. `016b` already
prescribes that key ("`page_components.id` is not stable across re-renders"). ⚠ `page_id` is **not**
in the dispatch loop's `call_handler` input_mapping (verified live) — so this needs either a mapping
addition or `page_name` from the spec. Go change; inert until a chassis roll.

**5 — close `083`** at ~2026-08-25 once `444`/`458`'s doors have held a week. Move with **both paths
on the commit** (`git mv` landmine) and verify at HEAD with `git ls-tree`.

**6 — the council gate cannot review config.** `097` scopes to `platform/`/`internal/`/`pkg/`, so a
mechanism shipping as `scheduled_tasks` / `agent_definitions` rows cannot be submitted; both of this
lane's rounds needed `FORCE=1`. A large share of this estate's behaviour is config. File as its own
item against the gate.

---

## 6. THE PATH — sequenced, with what each unblocks

1. **Decision 2's reclaim arm** (½ day). Self-contained, one migration, no roll. Do first: it has a
   clock and it protects the findings the later work will want back.
2. **Decision 1 Tier 1**, the `wont_fix` refusal (½ day + a roll). Coordinate with `301`'s author
   first — same function. Protects `phantom_internal_link`'s 69% path immediately.
3. **Decision 4**, `bugs_open/300`'s handler fix (½ day + the same roll). Bundle the roll with (2).
4. **Decision 5**, close `083` (~08-25, minutes).
5. **Decision 6**, file the gate item (minutes; do any time).
6. **Decision 1 Tier 2**, the detector→editor routing. **RFC first**, then build. This is the large
   one and it is the only thing that actually repairs the ~134 queued findings on owned pages.

Steps 1-3 are independent of each other and none blocks step 6's RFC being written in parallel.

## 7. Session-start checklist
`git log --oneline -10` · re-read this file from disk · `scripts/who-owns.py 083` and `301` **by
slug** (083 is an ambiguous number) · grep live `.jsonl` for `301|section-editor|held_pair` ·
re-measure §0 · then §6 step 1.

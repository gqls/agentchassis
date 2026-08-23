# HANDOFF — `bugfix_337_token_cap`, 2026-08-23 (evening). Read this first.

> ### ⚠ CORRECTED 2026-08-23 ~18:00Z — the previous version of this file (17:15Z, commit `ba255074e`) was WRONG on both of its live items, and re-reading the artefact is what caught it.
>
> It said **337 was blocked on one page** and that the live thread was **"determine what empties
> `hero-tool`'s `content_data` values"**. Neither survives contact with the data:
>
> 1. **The blocked page had already repaired itself** — refused 14:03:06Z, **saved cleanly on
>    retry at 14:23:29Z**, i.e. 20 minutes *before* that handoff was written, and ~2h50m before
>    it was saved. **337 is CLOSED.**
> 2. **Nothing empties `hero-tool`'s values.** All 40 empty values are `stat_*` keys and the
>    count moves **both ways** across writes. There is no writer to census; the suggested
>    investigation is **cancelled**. Refutation filed to `bugs_open/253`.
>
> The previous version is intact in git (`git show ba255074e:<this path>`). Its §4 traps and §5
> commands were good and are kept below, extended.

---

## 1. CURRENT STATE IN ONE SCREEN

| | state |
|---|---|
| **`bugs_open/337`** | **CLOSED 2026-08-23.** Moved to `bugs_closed/`. All three pages serve their tools |
| **337 fix — Go** | **LIVE**, and re-proven *after* the 16:03Z roll to v1.0.1330 — by **capability**, not by sha (see §4) |
| **337 fix — prompt** | **APPLIED.** Migration **565** |
| **Council** | **APPROVED** 08-23 12:25Z, corr `9efde776-a210-42bc-aa99-899d0d301c67`, 11 of 12 seats |
| **§3c "does Arm A raise the orphan rate?"** | **ANSWERED as undecidable at this n**, with the required sample size named. Do not spend a session on it before ~1 week of volume |
| **`bugs_open/253`** | **OPEN.** This lane's two contributions to it are now **corrected**; one open question cancelled, none added |
| **Parked backlog** | Closed. 9 items cancelled as superseded |

**Nothing is in flight. No background job is running. Nothing uncommitted is mine.**

---

## 2. WHAT 337 TURNED OUT TO BE (re-scoped twice; both re-scopes were wrong, and so was the close-out)

Filed as a token-cap bug. **The cap was never the binding constraint** (73 of 82 generations
succeeded under the old 16,000 cap). Re-scoped to `bugs_open/309`'s unresolvable-source class —
**also wrong**: 101 `component_validation_rejected` rows split **97 field-contract / 3
source-vocabulary**.

**The actual defect: the birth gate enforced two contracts the writer was never shown.**

- The pre-generation advisory keyed on `section_type` (+ `is_active` + `component_level`) while
  the gate resolved the row it would overwrite by `function` with **neither** filter. A miss left
  the preservation block dormant behind `{{if .existing_component.field_names}}`, taking the
  function pin with it — so field preservation was **chance**. An 18-field component failed 70
  times; a 4-field sibling escaped on its second attempt by luck.
- TIER D enumerated every valid `query.*` name; **TIER C named no `site_specs` aspect at all**.
  The live blocker was a one-character invention: `site_specs.ctas` where the aspect is `cta`.

**The fix: state the gate's contract to the writer, computed by the gate's own functions, so the
two cannot drift.** Three reused (`resolveStorageIdentity`, `LoadKnownSpecAspects`,
`KnownQueryBases`), one extracted (`KnownAspectsSorted`), plus an `is_active`-gated
`section_type` heal on the rejection path. Register **CLC-027**.

**Proven end to end 08-23:** writer shown the vocabulary → stopped inventing a source → one
orphan-field refusal → `bugs_open/345`'s typed feedback named the field → retry stored the
component (12:31, 17,163 chars, 43 fields, zero rejections) → re-render attached it → the page
serves the tool.

**And then the part the previous handoff missed:** the third page did the same thing on its own,
20 minutes after a floor refusal, with no intervention.

---

## 3. WHAT IS LEFT

### 3a. On 337 — nothing. Verify it if you like; here is the predicate

```bash
curl -s -o /tmp/p -w '%{http_code} %{size_download}\n' https://loanzy.uk/tools/credit-health-check/index.html
grep -oc 'loans-credit-health-check' /tmp/p          # expect 18
grep -o 'id="c-loans-credit-health-check-[a-z0-9-]*"' /tmp/p | sort -u | wc -l   # expect 9
```
[MEASURED 2026-08-23 ~17:50Z] 200, 30,756 B, 18 refs, 9 instance ids, 13 `<button>`, 3,974 script
chars — the same repaired profile as `tool-eligibility-checker`, with page-correct content.

### 3b. On 253 — the investigation this lane proposed is CANCELLED, and that is the finding

The previous handoff sent you to census writers for whatever empties `hero-tool`'s
`content_data`. **There is no such writer.** [MEASURED 2026-08-23 ~17:45Z, 11 loanzy pages]

- Of **40** empty values, **40 are `stat_*` keys**. No other key is ever empty. Every page
  stores exactly **11** keys.
- They empty in label/value **pairs**: non-empty 11/9/7/5 = **3/2/1/0** of three optional stat
  slots filled; rendered classes 15/12/9/5, i.e. ~3 classes per unfilled stat (`skip_field` gates).
- Across successive writes the count moves **both ways**: comparison **0→3**, repayment
  **0→0→3**, settlement **0→2**, overpayment **1→0**. *A mechanism that empties cannot fill.*

So the floor refusals are **per-run generation variance** and they are **retryable** — which the
`bugs_open/305` lane had already established on 08-22, in the same file, directly above where
this lane wrote the opposite the next day. Full correction, with the disconfirming condition
stated, at the foot of `bugs_open/253`.

**If you want to do something for 253**: its own subject — a framework rewrite stripping layout
components — is untouched by any of the above and is where the bug actually lives.

### 3c. Genuinely optional

- **Orphan rate under Arm A** — [MEASURED 2026-08-23] `08-16..08-22: 0 / 48 stored (0%)` vs
  `08-23 post-fix: 1 / 10 (10%)`. **One event decides nothing** (95% interval ~0.3%–45%).
  Re-run in ~a week; you need ≥16 advised generations to *see* a 10% rate at all and hundreds to
  distinguish a raised one. The whole test is the two queries in §5.
- `bug_historian`'s conceded objection (real leaf paths make a plausible-but-wrong leaf key more
  tempting) — a residual in `bugs_open/362`. Guidance, not enforcement.

---

## 4. TRAPS THIS LANE PAID FOR — the most reusable thing here, now FIVE

**Four wrong success-predicates and one blind instrument, all of which would have reported a
working thing as broken.**

| predicate | why wrong |
|---|---|
| `grep -c '<input'` | the component is a **button-driven quiz** — 0 inputs while fully working |
| a name-derived URL | URL shape is **per site**; loanzy serves `/tools/<n>/index.html`, loancalculator `/tools/<n>.html`. A wrong guess returns a **1,201-byte custom 404 with a stable md5** that survives a two-reads stability check |
| `<section class="…-section">` | these render as **instance-scoped `<div>`s** — `id="c-<function>-…"` |
| **a handoff's own status table** | **added 08-23 evening.** It recorded a page as blocked ~2h50m after that page repaired itself on retry. **A snapshot of a RETRYABLE failure reads exactly like a permanent blocker** |

**The reliable predicate is the component's `function` name plus the `id="c-<function>-` prefix**
— properties of the artefact, not of a rendering you assumed. **And re-read the artefact before
repeating any blocked/failed status you did not personally re-measure this hour.**

**⚠ The blind instrument, added 08-23 evening — this one is the dangerous species.**
`grep -ac <sha> /proc/1/exe` on the chassis pod returned **0 for the fix commit and 0 for the
positive control**. A failed positive control means **the METHOD is blind — it is NOT evidence
the fix is absent.** Read as a bare result it says "your fix is not deployed", which is false and
would have re-opened a closed bug. The `build provenance` log line had already scrolled past
`--tail=3000` (as CLAUDE.md warns for this service). **What worked: probe the CAPABILITY** — see §5.

**Other traps, carried forward from the previous version:**

- **`git checkout <file>` destroyed another session's uncommitted work.** It restores the WHOLE
  file from the index. `git stash` is hook-blocked for this blast radius; the one-path form is
  **not**. To mutation-test: `cp` to scratch immediately before, `cp` back after.
- **`who-owns.py` reads COMMITS and is blind to uncommitted sessions.** Check `git status` on the
  actual code files, and check **mtimes** — 253's guard files are dirty with gofmt whitespace
  from 08-21, which reads as active work and is not.
- **Before filing a stall/starvation claim, get the self-heal interval of the nearest known
  mechanism.** `bugs_closed/029` self-heals at ~40 minutes.
- **`find_dispatchable_site` no longer exists.** `RUNBOOK_311_fix.md` and the
  `single-page-deploy` memory both still name it. Only a stale comment at
  `load_work_item_actions.go:829` references it.
- **Queue order is `priority ASC, created_at ASC`** — *lower first*. The runbook's re-render
  recipe uses priority **99**, so a hand-filed re-render goes behind everything.
- **Pick a migration number when you WRITE the file.** 561→564 all went to other lanes within
  the hour; this lane landed on **565**.

---

## 5. COMMANDS YOU WILL WANT

```bash
# "Did my Go change ship?" — try the log line first, but EXPECT IT TO HAVE SCROLLED on agent-chassis
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=3000 | grep -m1 'build provenance'
# Empty means NOT IN RANGE, not unstamped. And if you probe the binary, ALWAYS run both controls:
#   a sha that must be PRESENT and one that must be ABSENT. If the present-control fails, the
#   probe is blind — discard it, do not report the absence.
```
```sql
-- ✅ THE ONE THAT ACTUALLY WORKS: probe the CAPABILITY, not the commit.
-- aspect_paths is the field the 337 fix adds; non-zero on a RECENT row = live NOW.
SELECT created_at,
       length(collected_data->'existing_component'->>'field_names')  AS fn_chars,
       length(collected_data->'existing_component'->>'aspect_paths') AS ap_chars
FROM orchestration_states WHERE collected_data ? 'existing_component'
ORDER BY created_at DESC LIMIT 3;      -- ap_chars ~10,292. fn_chars 0 is NORMAL when no existing row.

-- Classify rejections BEFORE naming a class (the query whose absence cost two re-scopes)
SELECT CASE WHEN error_message ILIKE '%no site carries a site_specs aspect named%' THEN 'phantom_aspect'
            WHEN error_message ILIKE '%removes/renames%'                           THEN 'stranded_fields'
            WHEN error_message ILIKE '%template variables and schema fields%'      THEN 'orphan_field'
            ELSE 'other' END AS class, count(*)
FROM agent_error_log WHERE error_code LIKE 'component_validation%' GROUP BY 1 ORDER BY 2 DESC;
-- gotcha: the timestamp column is occurred_at, NOT created_at

-- DEMAND CONTROL — a post-fix zero is worth nothing without it. Note component_level:
-- the birth gate governs 'section'; counting all levels inflates the denominator (16 vs 9).
SELECT component_level, count(*) FILTER (WHERE updated_at >= '2026-08-23 12:00+00') AS post_fix,
       count(*) FILTER (WHERE updated_at < '2026-08-23 12:00+00') AS prior
FROM content_components WHERE updated_at >= '2026-08-22 12:00+00' GROUP BY 1;

-- 253: non-empty values vs rendered class count, and WHICH keys are empty (the discriminating half)
SELECT p.name, kv.key FROM pages p JOIN sites s ON s.id=p.site_id
JOIN page_components pc ON pc.page_id=p.id
JOIN content_components c ON c.id=pc.component_id AND c.function='hero-tool'
CROSS JOIN LATERAL jsonb_each(pc.content_data) kv
WHERE s.domain='loanzy.uk' AND kv.value::text IN ('""','null') ORDER BY 1,2;   -- all 40 are stat_*

-- Direction of change across writes (kills any "something empties it" theory)
-- page_component_history rows are the PRE-WRITE state, source='artefact_archive_trigger', op='delete'.
```
```bash
# Verify a repaired tool page — HTTP 200 AND the function name, never <input or <section class
curl -s -o /tmp/p -w '%{http_code} %{size_download}\n' https://loanzy.uk/tools/<name>/index.html
grep -oc '<function-name>' /tmp/p ; grep -o 'id="c-<function-name>-[a-z0-9-]*"' /tmp/p | sort -u | wc -l
```

---

## 6. THE FILES

- **`bugs_closed/337_HANDOFF_2026-08-20_one_section_type…md`** — moved 08-23 evening. Read the
  **last** section for the close-out; the header and several middle sections are superseded.
- `bugs_open/253_HANDOFF_2026-08-11_framework_rewrite_of_a_prose_block…md` — **resolve 253 by
  SLUG; the number names two unrelated bugs.** This lane's contributions and their correction
  are at the foot.
- This lane: `PLAN` (D4/D5 carry the design decisions and the owner's rulings), `NOTES`
  (every misstep, newest at the bottom), `RUNBOOK`, `README_where_we_are.md`
  (**the owner's document — append, never rewrite**).
- Code: `load_existing_component_action.go`, `component_source_guard.go` (`KnownAspectsSorted`),
  `store_generated_component_action.go` (`healRejectedComponentSectionType`), + two test files.
- Migration `565_component_creator_prompt_names_the_source_vocabulary.sql` (+ `_ROLLBACK`).

## 7. OWNER RULINGS IN FORCE

- **Re-drive all parked items** — *given 08-22, stale by 08-23.* Checked before spending: 11
  became 9, most damage was already repaired, and a re-drive would have **regenerated**
  components other pages depend on. Closed as superseded instead. **The principle: re-check an
  instruction's premises before executing it on a tree this fast.** *(08-23 evening: the same
  principle applies to a handoff's own status table — see §4.)*
- **Vocabulary richness** — leaf `aspect.key` paths *with site coverage*, not bare aspect names.
- **Take 253 on** — given 08-23. Done; the contribution is filed and now corrected.

## 8. CROSS-LANE — a message from the `bugs_open/345` lane, 08-23 evening

The 345 lane reports that the **truncation remedy text this lane drafted into 345 on 08-22 is
still unwired and cannot be wired by a prompt migration alone**: migration 561's typed channel
(`site_work_items.retry_feedback`) has exactly one writer (`recordRetryFeedback` at
`store_generated_component_action.go:1549`, called from one site, `:477`), and 563 gates the
prompt block on `last_error_code`, rendering nothing otherwise. **A truncation branch needs a
WRITER for the truncation class first** — added to the prompt alone it renders silence and looks
like a working fix. They also corrected a near-miss in this lane's write-up: the two
truncation-shaped checks at `:176-180` and `:186-193` return ~300 lines *before* `:477`, so they
write neither `agent_error_log` nor the channel — the path this lane implied does not exist.
**Unstarted, spans both lanes, and they have asked to agree it rather than build it
unilaterally.** Seat WII-026 in `work-item-integrity.md`.

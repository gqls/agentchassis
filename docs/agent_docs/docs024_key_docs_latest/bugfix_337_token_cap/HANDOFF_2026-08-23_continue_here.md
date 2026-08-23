# HANDOFF — `bugfix_337_token_cap`, 2026-08-23. Read this first.

Two things live here now: **`bugs_open/337`, which is one page from done**, and
**`bugs_open/253`, newly taken on**, because 253's guard is what blocks that page.

---

## 1. CURRENT STATE IN ONE SCREEN

| | state |
|---|---|
| **337 fix — Go** | **LIVE.** Build `2dbe12f1d` (11:51Z 08-23); `e1951c24b` confirmed an ancestor with a discriminating control |
| **337 fix — prompt** | **APPLIED.** Migration **565** (not 561/562/563/564 — all taken by other lanes) |
| **Council** | **APPROVED** 08-23 12:25Z, corr `9efde776-a210-42bc-aa99-899d0d301c67`, 11 of 12 seats |
| **Demand-proven?** | **Yes, at the mechanism AND at the artefact** — see §2 |
| **Pages repaired** | **2 of 3.** `is-a-loan-right-for-me` ✔, `eligibility-checker` ✔, **`credit-health-check` ✘** |
| **Parked backlog** | **Closed.** 9 items cancelled as superseded, reason on each row |
| **337 status** | **OPEN, for one page only**, blocked by 253 |
| **253 status** | **OPEN, TAKEN ON by this lane 08-23.** First finding recorded; no code written |

**Nothing is in flight. No background job is running. Nothing is uncommitted that is mine.**

---

## 2. WHAT 337 TURNED OUT TO BE (it was re-scoped twice, and the second re-scope was also wrong)

Filed as a token-cap bug. **The cap was never the binding constraint** (73 of 82 generations
succeeded under the old 16,000 cap). The previous session re-scoped it to `bugs_open/309`'s
unresolvable-source class — **also wrong**: measured at the call level, 101
`component_validation_rejected` rows split **97 field-contract / 3 source-vocabulary**.

**The actual defect: the birth gate enforces two contracts the writer is never shown.**

- The pre-generation advisory keyed on `section_type` (+ `is_active` + `component_level`) while
  the gate resolves the row it will overwrite by `function` with **neither** filter. A miss left
  the whole preservation block dormant behind `{{if .existing_component.field_names}}`, taking
  the function pin with it — so field preservation was **chance**. An 18-field component failed
  70 times; a 4-field sibling escaped on its second attempt by reproducing generic names by luck.
- TIER D of the prompt enumerated every valid `query.*` name; **TIER C named no `site_specs`
  aspect at all**, and the prompt rendered no part of `site_specs`. The live blocker was a
  one-character invention: `site_specs.ctas` when the aspect is `cta`, which carries **exactly**
  the two keys the writer wanted.

**The fix, in one line: state the gate's contract to the writer, computed by the gate's own
functions, so the two cannot drift.** Three reused (`resolveStorageIdentity`,
`LoadKnownSpecAspects`, `KnownQueryBases`), one extracted to be shared (`KnownAspectsSorted`).
Plus an `is_active`-gated `section_type` heal on the rejection path. Register **CLC-027**.

**The chain, proven end to end 08-23:** writer shown the vocabulary → stopped inventing a source
→ one orphan-field refusal → `bugs_open/345`'s typed feedback named the field → retry stored the
component (12:31, 17,163 chars, 43 fields, **zero rejections on the retry**) → re-render attached
it → **the page serves the tool.** Neither half would have done it alone.

---

## 3. WHAT IS LEFT — start here

### 3a. The one blocked page, and why you must not force it

`loanzy.uk/tools/credit-health-check` — the page 337 is named after. Its re-render is refused:

> `SECTION COMPONENT FLOOR REFUSED … hero-tool 12→5 class attributes (42% kept, floor 50%) …
> Nothing was written (bugs_open/253)`

**Do not set `section_component_floor` to push it through.** The floor is correctly refusing a
save that would leave the page thinner. Overriding a damage-prevention guard to land your own
repair is "fixing the checker to agree with a broken save".

### 3b. 253, and the finding that reframes it — this is the live thread

The five refusals all hit **one slot** (`hero-tool`) and looked like a renderer regression.
They are not. `hero-tool`'s template carries **18** `class=` attributes behind **11 `{{if}}`
gates**, with **11 of 13 fields `on_missing: skip_field`**. Every loanzy page stores **exactly
11** `content_data` keys for the slot; only the number of **non-empty values** differs, and it
tracks the class count almost exactly:

| non-empty values (of 11) | class attributes |
|---|---|
| 11 | 15 |
| 9 | 12 |
| 7 | 9 |
| 5 | 5 |

**So the class count is a proxy for CONTENT FULLNESS, not layout integrity.** The refusal's own
message (*"may not lose more than 50% of the elements carrying layout classes"*) sends you
hunting a layout defect. The real question is **why four of eleven field values emptied between
saves** — the `bugs_open/238` / `355` content-loss family.

**The next step, not started:** determine what empties `hero-tool`'s `content_data` values.
Suggested order — (1) find a page where it happened and diff `content_data` across
`component_versions` / the page's history; (2) identify the writer that emptied it (census the
write history, don't infer from the bug file); (3) only then decide whether the floor should
measure non-empty field values instead of class attributes — **that is a shared-guard contract
change affecting nine writers and needs the cause first.**

### 3c. Smaller, genuinely optional

- **Does Arm A raise the orphan rate?** Open and honestly unresolved. Advising a 43-name
  contract *might* make dropping one more likely. Evidence leans against (the class is 58 rows /
  24 items since 2026-08-03, three weeks older than the change) but it is unproven. The test:
  orphan rate on advised-contract generations vs blind ones, post-roll.
- `bug_historian`'s conceded objection: showing real leaf paths makes a plausible-but-wrong leaf
  key more tempting. Filed as a residual in `bugs_open/362`. Guidance, not enforcement.

---

## 4. TRAPS THIS LANE PAID FOR — read before verifying anything

**THREE wrong success-predicates in one lane, all of which would have reported a working page as
broken.** This is the single most useful thing here:

| predicate | why wrong |
|---|---|
| `grep -c '<input'` | the component is a **button-driven quiz** — 0 inputs while fully working |
| a name-derived URL | URL shape is **per site**; loanzy serves `/tools/<n>/index.html`, loancalculator `/tools/<n>.html`. A wrong guess returns a **1,201-byte custom 404 with a stable md5** that survives a two-reads stability check |
| `<section class="…-section">` | these components render as **instance-scoped `<div>`s** — `id="c-<function>-…"` |

**The reliable predicate is the component's `function` name plus the `id="c-<function>-` prefix**
— properties of the artefact itself, not of a rendering you assumed. And when a cheaper
measurement contradicts you (byte count grew, `page_components` says attached), **go and look;
do not invent a mechanism to reconcile them.**

**Other traps:**

- **`git checkout <file>` destroyed another session's uncommitted work.** It restores the WHOLE
  file from the index. `git stash` is hook-blocked for this blast radius; the one-path form is
  **not**. To mutation-test: `cp` to scratch immediately before, `cp` back after. Never revert a
  shared file to undo your own edit. (LANDMINES entry + `WRONG_CALLS.md`.)
- **`who-owns.py` reads COMMITS and is blind to uncommitted sessions.** It said 253 was OWNED;
  both named lanes only *cite* it. Check `git status` on the actual code files, and check file
  **mtimes** — 253's guard files are dirty with **gofmt whitespace from 08-21**, which reads as
  active work and is not.
- **Before filing a stall/starvation claim, get the self-heal interval of the nearest known
  mechanism.** loanzy went 40 minutes with zero dispatches while 12 other sites cycled 15–59
  times. Filing-grade evidence; entirely normal — `bugs_closed/029` self-heals at ~40 minutes.
- **`find_dispatchable_site` no longer exists.** `RUNBOOK_311_fix.md` and the
  `single-page-deploy` memory both name it as the dispatcher's selector. Gone — only a stale
  comment in `load_work_item_actions.go:829` references it.
- **Queue order is `priority ASC, created_at ASC`** — *lower number first*. The runbook's
  re-render recipe uses priority **99**, so a hand-filed re-render goes behind everything.
- **Pick a migration number when you WRITE the file.** 561→562→563→564 all went to other lanes
  within the hour; this lane landed on **565**.

---

## 5. COMMANDS YOU WILL WANT

```bash
# Did my Go change actually ship? (per SERVICE, and USE A CONTROL)
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=3000 | grep -m1 'build provenance'
git merge-base --is-ancestor <your-commit> <stamped-sha>   # must be YES
git merge-base --is-ancestor HEAD <stamped-sha>            # must be NO, or the test proves nothing
```
```sql
-- Did the advisory carry both contracts? (the demand bar for 337)
SELECT length(collected_data->'existing_component'->>'field_names')  AS fn_chars,
       length(collected_data->'existing_component'->>'aspect_paths') AS ap_chars
FROM orchestration_states WHERE collected_data ? 'existing_component'
ORDER BY created_at DESC LIMIT 3;                 -- expect ~697 and ~10,292, not 0/absent

-- Classify rejections BEFORE naming a class (the query whose absence cost two re-scopes)
SELECT CASE WHEN error_message ILIKE '%no site carries a site_specs aspect named%' THEN 'phantom_aspect'
            WHEN error_message ILIKE '%removes/renames%'                           THEN 'stranded_fields'
            WHEN error_message ILIKE '%template variables and schema fields%'      THEN 'orphan_field'
            ELSE 'other' END AS class, count(*)
FROM agent_error_log WHERE error_code LIKE 'component_validation%' GROUP BY 1 ORDER BY 2 DESC;
-- gotcha: the timestamp column is occurred_at, NOT created_at

-- 253's live evidence: non-empty values vs rendered class count
SELECT p.name, count(*) FILTER (WHERE kv.value::text NOT IN ('""','null')) AS non_empty,
       (length(pc.rendered_html)-length(replace(pc.rendered_html,'class=','')))/length('class=') AS classes
FROM pages p JOIN sites s ON s.id=p.site_id
JOIN page_components pc ON pc.page_id=p.id
JOIN content_components c ON c.id=pc.component_id AND c.function='hero-tool'
CROSS JOIN LATERAL jsonb_each(pc.content_data) kv
WHERE s.domain='loanzy.uk' GROUP BY p.name, pc.rendered_html ORDER BY 2 DESC;
```
```bash
# Verify a repaired tool page — HTTP 200 AND the function name, never <input or <section class
curl -s -o /tmp/p -w '%{http_code} %{size_download}\n' https://loanzy.uk/tools/<name>/index.html
grep -oc '<function-name>' /tmp/p ; grep -oc 'id="c-<function-name>' /tmp/p
```

---

## 6. THE FILES

- `bugs_open/337_HANDOFF_2026-08-20_one_section_type…md` — the full account, two corrections
  and all measurements. **Read the last three sections; the header is superseded.**
- `bugs_open/253_HANDOFF_2026-08-11_framework_rewrite_of_a_prose_block…md` — **resolve 253 by
  SLUG; the number names two unrelated bugs.** My take-on and the reframing finding are at the foot.
- This lane: `PLAN` (D4/D5 carry the design decisions and the owner's rulings), `NOTES`
  (every misstep, newest at the bottom), `RUNBOOK` (the queries above with their gotchas),
  `README_where_we_are.md` (**the owner's document — append, never rewrite**).
- Code: `load_existing_component_action.go`, `component_source_guard.go`
  (`KnownAspectsSorted`), `store_generated_component_action.go`
  (`healRejectedComponentSectionType`), + two new test files.
- Migration `565_component_creator_prompt_names_the_source_vocabulary.sql` (+ `_ROLLBACK`).

## 7. OWNER RULINGS IN FORCE

- **Re-drive all parked items** — *given 08-22, and it went stale by 08-23.* I checked before
  spending: 11 became 9, most damage was already repaired, and every section type had gained a
  component, so a re-drive would have **regenerated** components other pages depend on. Closed
  them as superseded instead, and said so. **The principle: re-check an instruction's premises
  before executing it on a tree this fast.**
- **Vocabulary richness** — leaf `aspect.key` paths *with site coverage*, not bare aspect names.
  That choice is what survives `345` landing mid-build: the leaf paths are the half no refusal
  message carries.
- **Take 253 on** — given 08-23, after an ownership check. Done; §3b is where it stands.

# 215 — two emitted pages canonicalise to ONE name and the whole replan write dies on the unique index; which plans fail is decided by LLM emission variance

**Filed 2026-08-08** by the brochure_component_library lane, from a live failure
the same morning (fundamentallyai replan, corr
`1cb17b11-fd10-4d52-836c-36e2fa246ff6`):

```
step write_site_plan failed: ... insert site_plan_pages for "tool-llm-cost-calculator":
ERROR: duplicate key value violates unique constraint "idx_site_plan_pages_name" (SQLSTATE 23505)
```

**Verification route, declared per the 2026-07-31 owner ruling:** not through the
090 loop; substituted first-hand evidence from the failing run's own
`collected_data` plus the code path read at HEAD — the raw emission contains
both colliding names (query below), the canonicalisation site and the
unguarded insert are cited by line, and the day-before run on the SAME site
shows the differential (one variant emitted → no collision).

## Mechanism

1. The planner LLM may emit a page under its stem name AND its canonical name
   in one plan. This run: `llm-cost-calculator` (page_type `tool`, 3 sections)
   **and** `tool-llm-cost-calculator` (page_type `tool`, 0 sections — a stub),
   plus the same pattern in waiting: `tools` (2 sections) and `tool-tools`
   (0 sections). Read them from the failed run:
   ```sql
   SELECT p->>'name', p->>'page_type', jsonb_array_length(p->'sections')
   FROM (SELECT jsonb_array_elements(collected_data->'llm_plan'->'result'->'pages') p
         FROM orchestration_states
         WHERE correlation_id='1cb17b11-fd10-4d52-836c-36e2fa246ff6') x
   WHERE p->>'name' LIKE '%llm-cost-calculator%' OR p->>'name' LIKE '%tools%';
   ```
2. `WriteSitePlanAction` canonicalises each page independently
   (`datahelpers.CanonicalisePage`, called at
   `write_site_plan_action.go:277`): a `tool`-typed `llm-cost-calculator`
   becomes `tool-llm-cost-calculator`. **Nothing dedups the page list after
   canonicalisation**, so two entries now carry one name.
3. The per-page insert (`write_site_plan_action.go:379`) hits
   `idx_site_plan_pages_name` on the second one and the action errors. The
   write is transactional — verified in the same incident: no new `site_plans`
   row, previous plan still `is_current`, zero orphan rows — so there is **no
   data damage**, but the entire replan is lost.
4. Whether any given replan fails is therefore decided by **whether the LLM
   happened to emit both spellings that run**. The previous morning's replan
   of the same site emitted only the stem variant and succeeded; this one
   emitted both and died. A retry may pass. That makes this a low-frequency,
   zero-signature reliability hole in every `build-site-planner` run
   fleet-wide — the failure names neither the canonicaliser nor the LLM, and
   an operator's likeliest wrong conclusion is the one the error suggests
   (a stale unique index or concurrent write).

## Why the stub pages exist at all

The prompt's context shows the site's existing pages (which include the
canonical `tool-*` names) while the planning instructions talk about tool
pages by stem. The model, asked to enumerate every page (rule 17's
every-page requirement arrived the same morning — seed 333 — and plausibly
raised the odds of exhaustive enumeration), lists both spellings. Emission
variance, not a prompt defect as such — the write path must be safe against
it regardless.

## Fix candidates, ordered by what closes the door

1. **Dedup after canonicalisation, inside `WriteSitePlanAction`** (the only
   door): group the validated+canonicalised page list by final name; merge
   duplicates keeping the entry with sections (a stub loses to a composed
   entry; two composed entries = keep first, log loudly with both section
   lists). ~20 lines before the insert loop; makes the collision
   unrepresentable regardless of what the LLM emits.
2. **Planner prompt: name pages canonically** (tell it tool pages are always
   `tool-<stem>` and never to emit both). Reduces the odds; closes nothing —
   the write path would still be one emission away from dying.
3. Retry-on-23505 at the orchestration layer — treats the symptom, hides the
   defect, explicitly NOT recommended.

## How to verify a fix

Feed `WriteSitePlanAction` a page list containing a stem + its canonical
variant (unit: the failed run's raw pages array is a ready fixture) — the
write must succeed with ONE `site_plan_pages` row for the name, the composed
sections must win over the stub, and a log line must name the merge. Then
re-run the census: no plan write failure with `SQLSTATE 23505` on
`idx_site_plan_pages_name` in `orchestration_states.error` after the fix
ships (query `error LIKE '%idx_site_plan_pages_name%'`).

## Relations

Found while executing `bugs_open/151` candidate 1 Slice B (RFC_016 §3a option
(a) compliance replan — the failure cost that observation run);
`bugs_open/204`/`214` (the same wire's other positional/naming traps);
`datahelpers.CanonicalisePage` (the canonicaliser itself is correct — the gap
is the absent post-canonicalisation dedup).

---

## CORRECTED + STRENGTHENED 2026-08-09 — the collision is proven, the PAIRING was read from the wrong key, and the same defect has a second, quieter damage mode

**What stands, re-verified at HEAD today:**

- The error is a quoted fact: `insert site_plan_pages for "tool-llm-cost-calculator":
  duplicate key ... idx_site_plan_pages_name`. The insert names `r.Name`, so **two
  rows in `planRows` carried that canonical name** — that much is certain.
- **There is no dedup anywhere on the path**, re-read at HEAD 2026-08-09:
  the canonicalise loop appends unconditionally
  (`write_site_plan_action.go:274-315`) and the insert loop executes one statement
  per row (`:355-381`). `idx_site_plan_pages_name` is `UNIQUE(plan_id, name)`.
  So a post-canonicalisation duplicate ALWAYS aborts the write.

> **CORRECTED: the claim that the colliding pair was the emitted
> `llm-cost-calculator` + a `tool-llm-cost-calculator` stub is an INFERENCE, not a
> measurement — and it was read from the wrong stage.** I took it from
> `llm_plan.result` / `validate_plan`, but `WriteSitePlanAction` reads neither: it
> calls `extractPagesFromPlan`, which reads **`page_plan` then `site_plan`**
> (`site_db_actions.go:749-782`). I never inspected `site_plan` for that run, and
> the row has since expired (~24h; verified gone 2026-08-09), so **which two
> entries collided is now permanently [UNVERIFIABLE]** for this incident. This is
> the same error class as `WRONG_CALLS` 2026-08-08, committed one day after
> writing that entry — see the 08-09 entry there.
> **A reproduction must read `site_plan`, not `validate_plan`.**

**Second damage mode, measured today — the same dual-identity problem that does
NOT crash, and it reached production.** The 2026-08-07 replan of fundamentallyai
(plan `8ee5807b`) wrote page rows for canonical/stem twins of pages that were
already live under the other spelling. Three rows, all created 08-07 08:24:22,
all `planned` + `deployed_at IS NULL` + zero components — i.e. permanent 404s:

| phantom row (archived 08-08) | live twin, serving 200 |
|---|---|
| `tool-llm-cost-calculator` → `/tools/llm-cost-calculator/index.html` | `llm-cost-calculator` → `/tools/llm-cost-calculator.html` |
| `tool-tools` → `/tools/tools/index.html` | `tools` → `/tools.html` |
| `ai-readiness-checker-guide` → `/blog/ai-readiness-checker-guide.html` | `tool-ai-readiness-checker-guide` → `/guides/…` |

Note the direction flips (phantom is the canonical form twice, the stem form
once) — the invariant is **two identities for one page**, not a fixed prefix.

They were found and hand-archived by the fundamentallyai sweep front on
2026-08-08 (`HANDOFF_2026-08-09_sweep_front_continue_here.md` §2b), which also
had to cancel four `needs_human_review` work items pointing at them. Worse,
while they existed they were valid internal-link targets — a `pages` row is
`active` from creation — which is the ammunition behind that front's own
linkability fix (`1c2e25c8f`): a served page linked to
`/platform-log/index.html` for 18 days while it 404'd.

**So the severity is higher than filed, and the cost is already paid twice:**
one lost replan (crash mode) and three phantom 404s plus four dangling work
items (quiet mode), from two consecutive replans of one site.

**Fix candidate 1 covers both modes and needs one addition:** dedup by canonical
name inside `WriteSitePlanAction` closes the crash; the quiet mode also needs the
plan's page identities reconciled against **realised pages under either
spelling** (a plan row whose canonical name differs from a live page's name but
resolves to the same page must not create a second identity). Verify the quiet
mode with the census:

```sql
SELECT s.domain, p.name, p.url, p.created_at
FROM pages p JOIN sites s ON s.id=p.site_id
WHERE p.status NOT IN ('deleted','archived')
  AND p.deployed_at IS NULL AND COALESCE(p.build_status,'')='planned'
ORDER BY 1,2;   -- fleet-wide phantom candidates; HTTP-test before acting
```

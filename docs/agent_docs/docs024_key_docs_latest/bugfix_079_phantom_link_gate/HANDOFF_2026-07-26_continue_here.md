# HANDOFF — 079 phantom link gate — **CLOSED 2026-07-27**

> **This case is CLOSED.** Fix live in `v1.0.1171` and PROVEN end to end on that binary
> (`checked_links 3 · rewritten 1 · unlinked 1`; the invented target unlinked with its text
> intact, `/contact` rewritten to the stored `/contact.html`, valid and external links
> untouched; `agent_error_log` row `CONTENT_LINK_REPAIR_DETAIL` carries both hrefs). File moved
> to `bugs_closed/`. Kept for the method and the landmines — **the live work is now
> `bugs_open/092`**, the upstream cause.

**Written 2026-07-26 ~21:30Z.** Read this first, then `NOTES_phantom_link_gate.md` (missteps),
then `RUNBOOK_phantom_link_gate.md` (the commands).

---

## One paragraph

The deploy gate detected every dead in-body link accurately and published the page anyway,
because a miss is filed at `warning` and `valid` counts only blockers and errors. **Fixed and
live in `v1.0.1171`**: the gate now repairs the link in `clean_html` — rewrite to the stored
`pages.url` if the target really exists, otherwise drop the `<a>` and keep the text. Proven end
to end on the deployed binary 2026-07-27; case closed.

## State

| thing | state |
|---|---|
| Code | `43f254be5` (fix + tests + docs), `31d8ac7dc` (gofmt) |
| Live | **YES — `v1.0.1171`**, pod-grep verified against a pre-roll baseline |
| Unit tests | 13 cases, green, **induced-fault probed** (8/8 fail when stubbed) |
| End-to-end induction | **DONE 2026-07-27** — both arms exercised live, owner-approved |
| Council | `97904892-…` **DROPPED, not queued** — zero rows after 13h while 676 orchestrations ran. No trailer, none possible |
| `079` | **CLOSED**, in `bugs_closed/` |
| `bugs_open/092` | NEW, filed by this thread — the upstream cause |

## HOW THE INDUCTION WAS DONE (kept — it took four attempts)

A crafted dispatch only materialises if the agent type has a **live pod**. An inline
`config.workflow` and a freshly-seeded agent type both vanished with no orchestration row and no
error. What worked: point `content-reviewer`'s `validate_content` step `html_field` at
`input_data.page_html` (owner-approved, ~2 min, restored and **byte-compared** against a
pre-patch snapshot), then dispatch the crafted HTML. No natural induction was ever available —
5 real builds ran the new code with `checked_links: 0` because the writer emits no anchors at
all (`bugs_open/092`).

## SUPERSEDED — the original "one thing owed"

A `page-build-handler` run was dispatched and had not started:

```
corr       a1dfbf68-a312-4009-8bb4-5375224e87c9
work item  560d50cd-c7b2-48e1-bc17-fb7c215d31b1   (status 'triaged')
target     webdesign.co.uk / learn-design-digital-grain
```

**First check whether it ran while you were away:**

```sql
SELECT status, current_step, left(COALESCE(error,''),150)
FROM orchestration_states WHERE correlation_id='a1dfbf68-a312-4009-8bb4-5375224e87c9'::uuid;

-- the fix's own fingerprint — these keys exist ONLY in the new binary
SELECT collected_data->'validation_result'->'links_rewritten' AS rewritten,
       collected_data->'validation_result'->'links_unlinked'  AS unlinked,
       collected_data->'validation_result'->'link_repairs'    AS repairs
FROM orchestration_states WHERE correlation_id='a1dfbf68-a312-4009-8bb4-5375224e87c9'::uuid;
```

`links_rewritten` being **present at all** (even as 0) proves the new code path executed. Repairs
> 0 gives the full chain. Then the durable record, and then the artefact:

```sql
-- NOTE: the column is occurred_at, NOT created_at
SELECT occurred_at, domain, context FROM agent_error_log
WHERE error_code='CONTENT_LINK_REPAIR_DETAIL' ORDER BY occurred_at DESC LIMIT 5;
```

**The assertion that actually matters** — assert the OLD state is ABSENT from the saved artefact,
not that new machinery ran (a log row alone is a vacuous marker):

```sql
SELECT p.url, pc.slot_name FROM page_components pc
  JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE s.domain='webdesign.co.uk' AND p.name='learn-design-digital-grain'
  AND pc.rendered_html ~ 'href="/[a-z/-]*"'   -- then eyeball: every href must resolve in pages
;
```

**If it never ran**, re-dispatch — the envelope is in `NOTES`/`RUNBOOK`; the key point is that
`page-build-handler` must be dispatched **directly by kcat**, because the work-item dispatcher is
per-site (`WHERE wi.site_id = $1`) and a `triaged` row sits forever until that site's build
pipeline is triggered. Do not dispatch within ~300s of a chassis restart.

**Clean up when done:** work item `560d50cd-…` is mine and exists only for this test —
`UPDATE site_work_items SET status='cancelled' WHERE id='560d50cd-c7b2-48e1-bc17-fb7c215d31b1';`

## Then close 079

Move `bugs_open/079_…` → `bugs_closed/` once the induction lands, replacing the "NOT PROVEN"
paragraph in its banner with the evidence. If the induction shows a defect instead, that is a
better outcome than a quiet close — record it in `NOTES` and `WRONG_CALLS`.

## What NOT to redo

- **Do not re-measure the census.** 3 of 16 builds, 17 instances, 15 unique targets, all pure
  inventions, both pages homepages. Queries in RUNBOOK R1/R2.
- **On a missing council orchestration row: check whether the LANE is moving before concluding
  anything.** > **CORRECTED 2026-07-27:** this entry originally said the missing row was latency
  and "not a dropped submission". That was wrong — 13h later there was still no row while **676
  orchestrations ran fleet-wide**, so it was a drop. The standing "do not resubmit on a missing
  row" rule is still right as a first reading, but it needs its falsifier attached:
  `SELECT count(*) FROM orchestration_states WHERE created_at > '<dispatch time>'`. Large number
  = your message is not queued behind anything. Logged in `WRONG_CALLS`.
- **Do not promote the severity** "to be safe". It stops homepages shipping. The measurement is
  the whole argument and it is in the bug file.
- **Do not make `NormalizePagePath` tolerant of a missing `.html`.** `049` recorded why: `/contact`
  genuinely 404s, so tolerance silences the detector on a live defect. The tolerance belongs in
  the repair, where it now is.

## Landmines this thread hit

1. **`orchestration_states` retains 13 days, not ~24h.** `071` and other docs assert the shorter
   figure. The census they call impossible is possible.
2. **A comment is not in the binary.** "the old policy comment is gone" is a *vacuous* pod-grep
   marker. Only compiled strings — error codes, config keys, log messages — discriminate. Always
   pair with a positive control that predates your change.
3. **`agent_error_log` has `occurred_at`, not `created_at`.**
4. **A panic in one Go test disables every test declared after it**, and `go test` without `-v`
   makes a third of a run look like all of it. Read `=== RUN` lines. Full write-up in
   `016b` §9 and `WRONG_CALLS`.
5. **`page-rebuild` / `page-rerender` do NOT call `validate_page_content`.** Only
   `page-build-handler`, `content-reviewer`, `tool-recreation-handler` and `report-builder` do.
   Dispatching a rerender to test this gate proves nothing.
6. **Checking a bug number is free is not reserving it.** I checked 090, wrote the file, and lost
   it by 67 seconds. Renumbered to 092.
7. **My `016b` §9 append was swept into another session's commit** (`d5988a8ed`). Nothing lost;
   the only defence is committing sooner.

## Next work, in priority order

1. ~~Finish the induction, close 079~~ — **DONE 2026-07-27.**
2. **`bugs_open/092`** — the writer gets no link constraints, 20 of 20 runs. Prevention, and
   worth more than the backstop. Two traps recorded in the file: do **not** wire
   `InjectLinkConstraints` (dead duplicate of the step that already runs), and fix the
   `"/"+name+".html"` synthesis to read `pages.url`.
3. `bugs_open/071` — the fragment blind spot (24 of 25 anchored links fleet-wide point at an `id`
   that does not exist) is untouched by this fix. `RepairPageLinks` repairs the *path* and carries
   the fragment across, which turns those from 404 into inert — an improvement, not a fix.

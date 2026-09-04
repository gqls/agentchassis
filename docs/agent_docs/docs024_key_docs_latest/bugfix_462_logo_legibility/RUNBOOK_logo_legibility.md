# RUNBOOK — 462, logo legibility

**Shared logo recipes live in the 417 lane and are NOT duplicated here** —
`docs/agent_docs/docs024_key_docs_latest/bugfix_417_logo_text_policy/RUNBOOK_logo_text_policy.md`
has: fetching a generated asset's bytes safely, regenerating one site's logo by resetting the work
item, telling a guard refusal from a roll-kill, and why replacing a logo needs no page re-assembly.
This file holds only what is 462's own.

## Run the check

```bash
./scripts/audit-logo-legibility.py                     # the fleet (~90s)
./scripts/audit-logo-legibility.py --site a.com --site b.com
./scripts/audit-logo-legibility.py --json out.json     # one record per site, shaped for a filer
./scripts/audit-logo-legibility.py --self-test         # offline: no cluster, no network
```

`--self-test` is the fastest way to confirm the arms still fire after any change. Six cases,
including both **preserved** websitepromotion marks (pre- and post-regeneration), so it exercises the
motivating artefacts and not only synthetics.

**Exit codes: 0 = every logo measured and legible; 1 = anything else; 2 = usage.** A BLIND,
NOT-DISPLAYED or baked-background row always prints and never counts as a pass.

> ⚠ **Do not read the exit code through a pipe.** `… | tail -40; echo $?` prints **`tail`'s** status,
> which is 0 whatever the script did. I read that as the script exiting 0 on two findings and nearly
> filed a false defect against it. Check it bare:
> `./scripts/audit-logo-legibility.py --site X >/dev/null 2>&1; echo $?` → `1`.
> Or use `${PIPESTATUS[0]}`.

## The two arms, and why both are needed

- **Arm A** — *no pixel anywhere in the mark reaches 3:1 against its header.* Catches
  pre-regeneration websitepromotion (max 2.55:1) and mortgagecalculator (max 2.39:1).
- **Arm B** — *less than 15% of the mark's ink clears 3:1.* Catches post-regeneration
  websitepromotion, which has **max 20.75:1** and is *less* visible: the high reading is a magenta
  despill fringe while 86% of the mark is white on a white header.

> **A max-only rule passes the artefact 462 was filed about.** The 417 RUNBOOK's original
> "read max, not median" advice was written against the pre-regeneration mark and is refuted by the
> post-regeneration one. **Both arms or neither.**

## Verify a claim about a logo, by hand

The script is the check; this is the cross-check. §1b of `bugs_open/462` has the luminance snippet.
Two operands people get wrong:

- **The backdrop is per site and sometimes dark.** `dartsonline.com`'s header is `#111520`. A check
  hardcoding white inverts the verdict. `seotools.co.uk` is `#faf8f3`, **not** white — §8c corrects a
  control figure that was measured against the wrong operand.
- **The mark is what the PAGE loads, not what `assets.url` says.** `fundamentallyai.com`'s row holds
  a presigned B2 link minted 2026-08-10 with `X-Amz-Expires=604800`; it 401s, while the page's
  `/assets/images/logo.png` is fine. Resolve from the served markup.

⚠ **And not the first `<header>` in the document.** Three sites open with
`<header class="info-card-grid__header">` — a content heading — long before the chrome.

## Queries this lane had to get right

```sql
-- Provenance. Anything that FILES must read this: a generated mark and a human's
-- upload are different findings (462 §9c). The sweep does not read it today.
SELECT domain, origin_type, coalesce(nullif(origin_model,''),'(empty)'), origin_prompt IS NULL
FROM assets a JOIN sites s ON s.id=a.site_id
WHERE a.status='active' AND a.purpose='logo' ORDER BY origin_type, domain;

-- The live logo-repair unit. NOT needs_logo (2 rows, both cancelled).
SELECT item_key, status, created_by, source, count(*) FROM site_work_items
WHERE item_type='needs_imagery' GROUP BY 1,2,3,4 ORDER BY 5 DESC;

-- Is a routing destination alive? needs_human_review was 370 on 2026-07-25 and is
-- 1,439 on 2026-09-04. Re-measure before routing anything there.
SELECT status, count(*), min(created_at)::date, max(created_at)::date
FROM site_work_items WHERE status='needs_human_review' GROUP BY 1;
```

> ⚠ **`site_work_items` is a rolling window.** Closing a row archives it into
> `site_work_items_archive`, so a count of the live table cannot see what a type SUCCEEDED at.
> `needs_logo` reads as "2 rows, both cancelled" live and **13 complete** in the archive. Query both
> before calling a mechanism unused.

## Traps handed over by the 417 lane (2026-09-04)

- **Date a regeneration by the STORAGE KEY's date prefix, never `assets.updated_at`.** 8 of 34 rows
  moved overnight with zero regenerations; two of those are on 417's licence list, which would have
  taken a sample from n=1 to a claimed n=3.
- **`assets.file_size` is not a change signal.** `mortgagecalculator` records 12,325 while the file
  its page serves is 70,156 bytes, and most logo rows leave the column NULL.

## Guards that exist but have NEVER fired — do not cite them as proven

`[MEASURED 2026-09-04, all 32 sites carrying `--color-header-bg`]`

- the **cascade-disagreement BLIND** branch: inline `<style>` and the linked stylesheet **agreed** on
  every site;
- the `var(--color-header-bg, var(--color-surface))` **fallback** path: resolved for **none** of them;
- **no site has a gradient or image header token** — 13 distinct values, all flat hex.

So option (b)'s documented weakness ("cannot see a logo on an image/gradient header") is currently
hypothetical. That does **not** weaken the staleness argument for option (a), which is the real one.

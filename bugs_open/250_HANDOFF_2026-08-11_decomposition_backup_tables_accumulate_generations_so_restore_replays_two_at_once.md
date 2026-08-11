# 250 — the decomposition loaders' backup table poisons its own rollback

**Filed 2026-08-11** by the LMC Track A session, after `load_lmc.py --apply` died
on its own backup step. **BOTH HALVES NOW FIXED (2026-08-11).** LMC: `load_lmc.py` (`1d99731ab`),
proven by round-trip. loancalculator: `decompose/load_decomposition.py`, ported
the same afternoon at the owner's direction, and its backup table repaired from
**28 rows over 27 pages** to one generation per page with column parity restored
(28 == live), stray preserved to `page_components_bak_strays_20260811_loancalc`,
`DO`/`RAISE` verify block asserting 1 row/page, 1 stray and column parity.

**What remains open:** the loancalculator fix is **not round-trip proven** — that
lane's `--restore` has not been exercised against a live page since the port. LMC's
was. Do that before relying on it (apply → restore → assert the stored md5 equals
that lane's recorded baseline → re-apply).

> **On the 2026-07-31 owner ruling (a `bugs_open/` file asserting a structural
> cause must go through `090`, or say why it substituted first-hand
> verification): substituted, and here is the why.** This is not an inferred
> cause. I hit the failure, read the two statements that cause it, measured the
> damage in *both* lanes' live backup tables, fixed one, and proved the fix by
> round-tripping a live page and comparing to a recorded baseline md5. The
> diagnosis loop's value is finding a cause that is not where the symptom is;
> here the symptom *is* the cause, in five lines of SQL I can quote. What the
> loop would add is a second opinion on a defect that is already reproducible on
> demand, and it cannot check the thing that actually matters (whether the
> sibling's owner wants my fix in their tool).

---

## The defect

Both decomposition loaders back a page up before replacing it, and both build the
backup the same way:

```sql
INSERT INTO <bak> SELECT pc.*
FROM page_components pc JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = '<site>'
  AND NOT EXISTS (SELECT 1 FROM <bak> b WHERE b.id = pc.id);   -- ← per ROW
```

The guard is **per row**. So once a page has been decomposed, the *next*
`--apply` of **any** page finds that page's new `prose-0` row is "not yet backed
up" and copies it in **beside** the original `ported-page` row.

`--restore` then does:

```sql
DELETE FROM page_components WHERE page_id = '<p>';
INSERT INTO page_components SELECT * FROM <bak> WHERE page_id = '<p>';
```

— which replays **both generations onto one page**: a whole verbatim document
*and* a prose section. That is precisely the nested-`<html>` corruption
`load_lmc.py`'s own docstring warns about ("ADDING a row beside a verbatim one
silently flips the page to assembly with the whole stored document as one
section"), arriving through the mechanism that exists to be the safety net.

**Nothing reports it.** The backup grows quietly, `--restore` exits 0, and the
page is wrong only when rendered.

### Second, independent defect in the same function: schema drift

The table is created `LIKE page_components INCLUDING ALL` on the day the lane
starts. `page_components` has since gained **`rendered_html_digest`**, so
`SELECT pc.*` yields 28 expressions for 27 target columns and the backup cannot
run at all — `ERROR: INSERT has more expressions than target columns`.

**The asymmetry is the trap.** The *restore* direction is unaffected: fewer
expressions than target columns is legal SQL and the trailing ones take their
defaults, so 27-into-28 succeeds. **Drift breaks the backup loudly and the
restore not at all** — so a session that tests only `--restore` finds nothing
wrong, and a lane that has not run `--apply` since the column landed does not
know its backup is inoperative.

## Measured, 2026-08-11

| | `..._bak_20260805_lmc` (LMC) | `..._bak_20260802_decomp` (loancalculator) |
|---|---|---|
| rows / distinct pages | 42 / 41 → **repaired to 41 / 41** | **28 / 27 — still poisoned** |
| pages holding two generations | 1 (`guides/how-loans-affect-mortgage-affordability`) | **1** |
| backup cols vs live cols | 27 vs 28 → **fixed** | **27 vs 28 — backup inoperative** |

```sql
-- the poisoned-rollback census, run per lane
SELECT count(*) FROM (
  SELECT page_id FROM <bak> GROUP BY page_id
  HAVING count(*) FILTER (WHERE slot_name =  'ported-page') > 0
     AND count(*) FILTER (WHERE slot_name <> 'ported-page') > 0) x;
```

**Blast radius if unfixed.** It scales with the work: converting a lane's pages
one at a time poisons one rollback per page after the first. Track A would have
poisoned ~16 of 17. Track B is 22 pages **carrying live consumer-finance
calculators**, and is exactly where someone will reach for `--restore` in a hurry.

## The fix (already applied to `load_lmc.py`)

1. **Guard per PAGE, not per row** — `NOT EXISTS (… b.page_id = pc.page_id)`.
   The table then means "the pre-decomposition snapshot, one generation per
   page", which is what a rollback source has to be, and it is idempotent.
2. **Name the columns** the backup table actually has, `pc.`-qualified (the
   source joins `pages`, so a bare `id` is ambiguous — that costs a second
   failed run if you miss it). A future added column then degrades to "not
   captured" rather than "nothing can be backed up".
3. `ALTER TABLE <bak> ADD COLUMN IF NOT EXISTS rendered_html_digest text;`

## To close this bug

- [x] Port 1–3 into `loancalculator_couk/decompose/load_decomposition.py`. **Done 2026-08-11.**
- [x] Repair that lane's backup, stray preserved before deletion, `DO`/`RAISE`
      verify block. **Done 2026-08-11** — 27 rows over 27 pages, 28 cols == live.
- [ ] **Round-trip one page in the loancalculator lane to prove it** — apply →
      restore → assert the stored md5 equals that lane's recorded baseline →
      re-apply. **This is the only thing still owed, and it is the whole point:**
      believing a rollback you have not watched work is how this went unnoticed
      since 08-05, and porting a fix is not the same as proving it.

**Ported at the owner's explicit direction (2026-08-11)**, having flagged that it
is another lane's tool. `scripts/who-owns.py` resolves no owner for it. The change
is confined to `backup_everything()` and is additive to the docstring otherwise;
nothing about that lane's apply/restore transaction shape was touched.

## How it was found

`load_lmc.py --apply legal` refused to run — it calls `backup_everything()`
first, so it **failed before writing anything**, which is the only reason this
was an inconvenience rather than a half-converted live page. Fixing defect 2 to
get moving is what put the backup table on screen, and the row count (42 over 41
pages) is what exposed defect 1. **A tool that fails safe told me something a
tool that succeeded would not have.**

## See also

- `docs024_key_docs_latest/loanandmortgagecalculator_couk/NOTES_…md`, 2026-08-11
  — full working, both inductions, and the round-trip proof.
- `016b` §9 — the transferable pattern: *a backup that re-captures on every run
  is a log, not a snapshot; a rollback needs exactly one generation per subject.*

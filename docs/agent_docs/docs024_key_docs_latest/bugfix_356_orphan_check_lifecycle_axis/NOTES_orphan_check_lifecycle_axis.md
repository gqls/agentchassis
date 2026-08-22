# NOTES — bugfix 356 (orphan check lifecycle axis)

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-08-22 — opening the lane

Session was named `bugs_open/298`. **298 is not in `bugs_open/`** — it is
`bugs_closed/298_HANDOFF_2026-08-17_internal_linker_candidate_pages_capped_at_15_…md`,
moved there by `140286123` on 2026-08-20. Verified at HEAD, not on disk, per the
`git mv` landmine:

```bash
git ls-tree -r --name-only HEAD -- bugs_open/ bugs_closed/ | grep 298
# -> exactly one line, in bugs_closed/
```

**Re-validated 298's closure before doing anything else** (the brief asked whether the
bug is still valid). It is genuinely closed, and the check is live config, not a claim:

| 298's fix | live reading 2026-08-22 |
|---|---|
| `LIMIT 15` dropped from `load_candidate_pages` | `has_limit15 = f` |
| `output_format` → `object` (313's half, same migration) | `object` |
| migration 490 recorded | `applied_at 2026-08-19 16:06:15Z` |
| `plan_links` has run | 1 row in `llm_call_log`, 21:18:53Z, `success=t` |

No binary probe needed for 298 specifically: its fix is pure DB config (migration 490),
live on apply. The `fail_on_non_numeric` tripwire is 313's Go half, not 298's, and 313's
own file already re-verified it aboard v1.0.1319–1322.

**So the live residual in 298 is its adjacent finding**, which 298 twice marks as
"unaddressed and unclaimed … a candidate for its own ticket". That is what this lane took.

### Ownership check before starting

`who-owns.py 298` → OWNED/active: `bugfix_275_silent_row_caps` (closed) and
`bugfix_313_internal_linker`. Both lanes' own docs say the linker work is closed out;
neither claims the adjacent finding — 298 and 313 both explicitly release it. Filed as a
NEW bug (356) rather than reopening 298, because it is a different defect from 298's cap.

---

## MISSTEP 1 — I read `result->'target_page'` and got 34 unreadable rows

First census of 298's adjacent finding returned:

```
complete_total 34 | complete_no_target 0 | complete_with_target 0 | complete_unreadable 34
```

I nearly recorded "the adjacent finding can no longer be measured — every result is the
`bugs_closed/287` spawn record". **That would have been a false finding built on a wrong
JSON path.** The outcome is nested one level deeper, under `response`:

```
result->'response'->'target_page'->>'count'     -> 17 no-target / 9 with-target / 8 genuinely unreadable
```

**What caught it:** dumping one row with `jsonb_pretty(result)` instead of trusting the
path I had inferred from 298's prose. **The cheap check that would have:** never write a
JSON path from a doc's description of the data — read one row first. Cost ~5 minutes.
The dangerous part is that the wrong path *fails silently as a plausible finding*: "all
unreadable" is a known real phenomenon on this table, so the wrong answer looked like a
corroboration of `bugs_closed/287` rather than like an error. Logged in `WRONG_CALLS.md`.

## MISSTEP 2 — I grep-counted lifecycle predicates and got the answer backwards

To size the class I ran a per-file `grep -c` for `status = 'active'` across
`discovery_checks/`. It scored **`check_orphan_pages.go` at 2** — the file whose entire
defect is that it has no lifecycle arm. The two hits are on `site_nav_items`
(`sni.status='active'`, lines 220 and 226), not on the page row.

Had I trusted it I would have declared the very file under investigation "clean" and gone
looking for the defect somewhere else. **What caught it:** noticing the count disagreed
with the SQL I had just read in the same file. **The cheap check that would have:** a
predicate is only a page-row predicate if it binds to the alias in the `FROM pages`
clause — grep cannot see that, so the audit has to read SQL. Recorded as a ⚠ in the
RUNBOOK so the next person does not repeat it. Logged in `WRONG_CALLS.md`.

## MISSTEP 3 (avoided, and worth recording because I nearly made it)

My first framework fix candidate was **a blanket archived-page guard at the work-item
filing seam** — attractive because `bugs_open/266` fixed its four-producer problem exactly
that way, at the one seam every producer passes through.

Reading 266 to borrow its argument is what stopped me. Its 2026-08-14 note from the
`bugfix_168_deployed_asset_path` lane says, in terms:

> **Do NOT let anyone "fix" the audits by filtering on `pages.status`.** … an archived page
> can be live, so filtering it out would stop auditing a page that really is asserting
> unsupported claims to the public. The audit is right to look.

A seam-level filter would have broken `check_unverified_claims`, which deliberately looks
at archived pages and has a **mutation test** protecting exactly that
(`check_unverified_claims_archivedskip_test.go`). So the fix has to be per-check with a
declared posture, not one gate. **The lesson is not "read 266"** — it is that a precedent
worth copying is also worth reading for what its own consumers later told it, which is the
half that sits below the fix.

---

## 2026-08-22 — the measurements that decided the shape

All `[MEASURED 2026-08-22]`, queries in the RUNBOOK.

1. **All 17 no-target `internal-linker` completions name an `archived` page.** `site_match`
   true on every row, so not a site-id mixup.
2. **Fleet census, running the producer's own predicate and grouping by `pages.status`:**
   34 archived pages are being filed as work right now — 15 `needs_internal_links`,
   3 `nav_drift`, 16 blog. **This measurement is disconfirmable and that is why it counts:**
   grouped by `status`, it could have returned zero archived rows.
3. **All three remedy paths already carry a lifecycle arm** — the live `internal-linker`
   `load_target_page` step, `navPageScopeSQL`, `blogPostsQuery`. The producer is the sole
   outlier. This is the finding that turned "three handlers are broken" into "one producer
   disagrees with itself".
4. **Recurrence:** the same archived keys filed 3–4× between 2026-04-24 and 2026-08-17.
5. **`pages.status` has exactly two live values** — `active` 759, `archived` 65. So
   `check_sectionless_pages`' `COALESCE(p.status,'') <> 'deleted'` excludes **nothing**, and
   `status IN ('active','deployed')` elsewhere works only by accident. Both *read* as
   lifecycle filters. I would not have caught either without querying the vocabulary rather
   than trusting the SQL's apparent intent.
6. **Only 3 files in `discovery_checks/` call `PageWantedLivePredicateFor`.** The class is an
   adoption gap, not eighteen independent mistakes.

### On the subagent audit — grounded, not quoted

The 18-check blast radius came from a delegated read of all 71 files. Per the standing rule
that *a subagent's report is another doc*, I re-verified a sample by hand **in both
directions** before it entered the bug file: three claimed-unarmed
(`check_sectionless_pages`, `check_empty_sections`, `check_literal_markdown`) and two
claimed-clean (`check_componentless_pages`, `check_tool_recreation_needed`). All five agreed.
Checking only the unarmed side would have tested one direction and missed a
systematically over-reporting audit. I also **did not carry forward its `557 active / 25
archived` figure** — that came from a doc comment dated 2026-08-03; the live numbers today
are 759 / 65.

### Postures are not binary — the design constraint that came out of the reading

`check_unverified_claims.go:458` uses a third posture:
`NOT (p.status = 'archived' AND <never deployed>)` — keep archived pages that are still
serving, drop the ones that never shipped, mutation-tested. Any registry that only offers
armed/unarmed cannot express it and would push authors back to hand-spelling, which is the
disease. Recorded in the bug file as a hard requirement on fix candidate B.

## `[UNMEASURED]` / open at time of writing

- **090 verdict.** Run `7bac4520-651d-41f9-aa98-f4721c49902f`; artifacts had not landed when
  the bug file was written. Recorded as `[VERDICT PENDING]` in the file, including the
  instruction to record it even if it REFUTES this analysis.
- Whether the 17 flag-only checks that select archived pages cause any harm at all. They
  route to `handler_agent = ""`, so they cannot dispatch — assumed harmless, **not measured**.
- Whether `check_unlinked_components.go:57`'s `UPDATE` on archived pages' components has done
  observable damage. Noted in the bug file, not chased.

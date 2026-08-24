# 386 — refreshing a counting fact turns every page that already rendered the OLD value into an "unregistered claim"

**Filed** 2026-08-24 by the `bugs_open/364` lane, found by a fleet claims census run for that
bug's residual. **Not** 364's mechanism — filed separately so the two are not conflated.
**Severity** medium. Nothing false is published; honest pages are convicted of dishonesty,
and the conviction is at `error` severity, which refuses a rebuild.
**Class** evidence-register drift / detector false positive with a moving reference.
**Status** OPEN, unowned, not fixed. Diagnosis below is first-hand and cited; no `090` run.

## 1. The symptom

`cmd/claimscan` over live `rendered_html`, fundamentallyai.com against its own current
register, 2026-08-24 — five findings, all in one `evidence-chart` component on `capabilities`:

```
NUMBER  capabilities  evidence-chart  "11513"  …Feed items collected 11513 verified 2026-08-23…
NUMBER  capabilities  evidence-chart  "10194"  …Feed items with a credibility assessment 10194 verified 2026-08-23…
NUMBER  capabilities  evidence-chart  "428"    …Council review rounds sent back for revision 428 verified 2026-08-23…
NUMBER  capabilities  evidence-chart  "483"    …Council review rounds approved 483 verified 2026-08-23…
NUMBER  capabilities  evidence-chart  "23"     …Council review rounds rejected outright by a guardian veto 23 verified 2026-08-23…
```

Every one carries its own `verified <date>` stamp. These are not invented figures — they are
this platform's own metrics, rendered from the register by the component that exists to render
the register.

## 2. Root cause

The register has moved and the page has not. Live `site_specs` (`aspect='evidence_base'`,
`is_current`) for `199733a8-ac9c-4c30-b2ce-65ecdac6f3bd`, read 2026-08-24:

| page renders | register now holds |
|---|---|
| 11513 | **11646** |
| 10194 | **10416** |
| 428 | **437** |
| 483 | **503** |
| 23 | 25 |

These are **counting facts** — `SELECT count(*) …` metrics that increase every day. Each refresh
of the register silently invalidates every deployed page that rendered the previous value.
`numberSupported` (`claims.go`) compares the page's number against the fact's **current** value,
with `tolerance` deciding the slack; an `exact` counting fact is wrong the moment the counter ticks.

**The blast radius is every site with a counting fact, not this one page.** The tighter the fact
hygiene (`exact` rather than `gte`), the sooner an honest page is convicted — the same perverse
gradient `bugs_open/364` §2 records for its own mechanism.

## 3. Why it matters more than five findings

- The finding is filed at `error` severity by `validate_page_content`, which **refuses the
  rebuild** — so a page whose only fault is being a day old cannot be rebuilt to fix itself
  without the writer happening to regenerate a different number.
- It is **self-inflicted and periodic**: the register refresh is a scheduled job, so this
  re-arms on its own cadence. Nobody has to do anything wrong for it to recur.
- It is **indistinguishable, in the queue, from a real fabrication.** A reviewer reading
  `unregistered_number "11513"` has no way to tell "the counter moved" from "the writer invented
  a figure" without going to the register and diffing. That is the honest cost: it spends the
  credibility of the finding class.

## 4. Fix candidates, ordered by what closes the door

1. **Make a counting fact's support a range anchored on its verification time**, not a point.
   A fact that is `count(*)` and monotonic is supported by any value between its previous and
   current reading. Needs the register to record that a fact is monotonic — it currently cannot
   say so. This makes the bad state unrepresentable rather than merely rarer.
2. **Re-render on fact refresh.** Whatever writes a new fact value queues a `page_rerender` for
   every page whose stored copy cites the old one. Closes the door on the *page* rather than on
   the *detector*, and is the direction that keeps published copy true.
   ⚠ This is the expensive one and it races the sweep — see the `LANDMINES.md` entry
   "A data repair RACES the sweep that publishes it".
3. **Widen tolerance on counting facts** (`gte`, or `approx_pct`). Cheapest, and it is what the
   fleet has drifted into doing by accident — but a broad `gte` silently vouches for unrelated
   numbers in the same window, which is the accidental-support mechanism `bugs_open/364` §2 warns
   about. **This one trades a false positive for a false negative; do not take it without measuring
   what it starts vouching for.**
4. Suppress the check on components that render the register. Rejected as written: it hides the
   drift instead of resolving it, and an `evidence-chart` is exactly where a wrong figure matters.

## 5. How to verify

```bash
# the drift, per site — any row where a page's rendered figure is not the register's current one
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At -c "
SELECT f->>'value', f->>'tolerance', f->>'kind'
FROM site_specs ss, LATERAL jsonb_array_elements(ss.data->'facts') f
WHERE ss.site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd'
  AND ss.aspect='evidence_base' AND ss.is_current;"
```
Then `cmd/claimscan` over that site's live `rendered_html` — **assert the exported row count
against the DB before trusting the scan**, it truncates silently (`WRONG_CALLS.md`, 2026-08-24).

## 6. Relations

- `bugs_open/364` (the census that found this; different mechanism — that one is about *whose*
  number it is, this one is about *when* the number was true).
- CLM-016 / CLM-014 (the surface gate and how this layer is measured), `numberSupported` and
  `EvidenceFact.Tolerance` in `platform/orchestration/datahelpers/claims.go`.
- `refresh_evidence_base_action.go` / `evidence_citations.go` — the two live writers of the register.

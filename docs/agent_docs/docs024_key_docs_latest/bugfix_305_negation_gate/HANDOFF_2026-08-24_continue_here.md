# HANDOFF 2026-08-24 — continue here (`bugfix_305_negation_gate`)

**Supersedes `docs/agent_docs/docs024_key_docs_latest/bugfix_305_negation_gate/HANDOFF_2026-08-23_continue_here.md`.**
Read that one for the council rounds and how the ceiling was found; read this one for state.

> ## ▶ ONE-LINE STATE
> **Every defect this lane owns is fixed, LIVE and demand-proven on chassis `v1.0.1332`.** The
> recommendation is **CLOSE `bugs_open/305`** with the residual filed — the only repairable copy left
> is on ONE page, blocked behind ANOTHER lane's claims work, with its rerender already queued.
> **Nothing here is inert and nothing is waiting on a build.**

---

## 1. Verified today `[MEASURED 2026-08-24 ~10:35Z]`

| thing | state | how it was checked |
|---|---|---|
| §26 accounting fix | **LIVE** | binary probe BOTH replicas: `no_answer_for_target` **1**, control `no_answer_for_targez` **0**, known-present `rewrite_negations` **8** |
| §27 ceiling fix (`569`) | **LIVE + DEMAND-PROVEN** | **124 calls at `max_tokens=16000`, `cut = 0`** (was 4/34 = 11.8%). Zero `repair_unavailable` since the apply. 6-, 7- and 8-target pages now `repaired` |
| post-roll reconciliation | **22/22 markers reconcile** | 0 not-reconciling, 0 account-for-none, 0 over-counted — see §2 for the honest limit |
| the three pages | **repairable 6 → 2** | shipped-scanner canary, §3 |
| council `f3046f0c` / `4829bd48` | both **APPROVED** | `unreadable=0`, all seats voted, on both |

## 2. ⚠ THE ONE THING THAT IS *NOT* PROVEN, AND IT IS A DEMAND-CONTROL GAP

The three-era split (bug file §28a) shows something neither bug file predicted — **the two defects
were causally linked**:

| era | markers | non-reconciling |
|---|---|---|
| A — old ceiling, old accounting | 37 | **40.5%** |
| B — NEW ceiling only | 137 | **15.3%** |
| C — new ceiling + new accounting | 22 | **0.0%** |

Raising the ceiling **alone** took 40.5% → 15.3% with the accounting code untouched: most of "the model
ignored a target" was **the model running out of room**.

**But era C's 0.0% does not yet prove the accounting fix in production.** Post-roll rejections are all
*judged* (`still_rather_than` ×4, `still_x_not_y` ×1) — **`no_answer_for_target` has never fired.** So
there was nothing for it to record. The mechanism is proven by three mutation-proven properties plus a
council round; the production sighting is outstanding.

**The check to run:**
```sql
SELECT r->>'reason', count(*) FROM orchestration_states os,
  LATERAL jsonb_each(os.collected_data) AS e(key,val),
  LATERAL jsonb_array_elements(e.val->'rejected') r
WHERE e.key LIKE 'copy\_gate%' AND os.updated_at > '2026-08-24 09:40:00+00' GROUP BY 1;
```
⚠ **If `no_answer_for_target` never appears at all, investigate rather than celebrate** — era B says
omissions ran at 15.3% under the same ceiling, so permanent absence would mean something else changed.

## 3. The damage half — canary re-run today

Shipped scanner over real `content_data`, 679-string brief corpus. Baseline 2026-08-20 was
`TOTAL 7 | exempt 1 | repairable 6`:

| page | total | exempt | repairable |
|---|---|---|---|
| `model-directory` | **0** | — | **0** ← carried BOTH sentences the owner quoted |
| `adoption-tracker` | 1 | 1 (`brief_supplied_sentence` = **D2**) | **0** |
| `protocol-tracker` | 2 | 0 | **2** |

**`protocol-tracker` is blocked and it is NOT this lane's.** It already carries a filed `needs_page`
rerender (`needs_human_review`, plus one `failed`) **and** a `claims_unverified` item naming 3
unregistered numbers. The site has 30+ open `needs_human_review` items. **Do not fire a rerender** —
it would duplicate a queued item and fail at the claims gate.

To re-run the canary: `RUNBOOK` §7 (scratch tree; `cmd/gatecanary` must never become a real command).
⚠ Two things that cost time today: the import is
`platform/orchestration/datahelpers`, **not** `platform/datahelpers`; and the brief lives in
`site_specs` keyed on **`aspect`** (not `spec_type`), with `content_direction` on `pages`, not `sites`.

## 4. What is left — none of it a defect, none of it blocking

1. `protocol-tracker`'s 2 hits — one rerender, already queued, another lane's claims work first.
2. **`D2`** — the exempt tagline: an owner decision about a brief.
3. **`D3` — still must NOT be decided.** The rejection log only became trustworthy today; 5 judged
   rejections so far. Give it a week. ⚠ `rather_than` is **71% of all rewrites**, so this is the
   decision with real reach.
4. `D4` (`negation_density` threshold), `D5` (`brief_supplies_negation` routing) — unchanged.
5. Not ours: the accounting-loop **sibling audit** a council seat asked for
   (`evidence_citations.go`, `revalidate_unverified_claims.go`) — open and unowned.

## 5. Standing cautions (carried + new)

- **The reconciliation census MUST be segmented by `status`**, and must **not** be tuned until it reads
  zero — five early returns (454/458/540/559/570) precede the accounting call at 665, so a
  `repair_unavailable` marker accounts for none of its targets **by design**. `RUNBOOK` §9.
- **Split any marker census at the change points.** A window spanning a roll is a MIXED batch and reads
  as one rate; splitting it is what revealed §28a.
- `llm_call_log.step_name` is the **loop-expanded** name — filed fleet-wide in `LANDMINES.md`.
- `/tmp` is a shared 16G tmpfs that was at 100%; point `TMPDIR` at your own scratchpad, don't clear it.
- Everything in the 08-22 handoff §6 still stands (`\y` not `\b`; brief-supplied is exempt BY DESIGN).

**Migrations owned:** `509`, `517`, `548`, `569` — all applied and recorded.
**Council:** `c48b7612`, `a696e2a3`, `f3046f0c`, `4829bd48` — all APPROVED.

# HANDOFF 2026-09-02 — improvement loop, continue here

> **⚠ SUPERSEDED 2026-09-03 by `HANDOFF_2026-09-03_continue_here.md`. Read that first.**
> Kept because §1's four traps are all still true and still catch people. What is STALE here:
> §3a's "re-submit migration 722" (done — APPROVED r5 and APPLIED, and it became a TRIGGER,
> not the column default this file describes) and every backlog figure (978 → 704 as the
> skip-link drain runs).

**COLD-START: read this file, then `SUMMARY_2026-09-02_improvement_loop_ownership.md`.**
Everything else in this directory is depth: `PLAN` §5 for the ordered work, `NOTES` for the
missteps (§(w), §(z), §(gg) and §(ii) are the four worth your time), `RUNBOOK` for the
commands with their traps attached, `README_where_we_are` for the owner's plain-prose log.

Lane opened 2026-09-02 on the owner's instruction. Before that the pipeline ran ~50×/day
with no owner and no standing account.

---

## 1. The four things that will mislead you in the first ten minutes

1. **The written record says the loop is switched off. It is not.** There is a standing
   owner ruling of 2026-07-29 — *"the improvement loop is stopped DELIBERATELY … do not
   re-enable it"* — still quoted as current by at least two documents. Migration `389`
   re-enabled it in August. **Read `scheduled_tasks`, never the ruling.**
2. **`complete_clean` does not mean the site is clean.** It is also the terminus for a
   skipped audit (mig 291's fingerprint gate) and for a site whose entire finding pile was
   held back as unroutable. `collected_data->'audit_state'` separates them.
3. **`execution_path` is empty on every improvement-loop row.** It looks like the way to
   ask which steps ran and will tell you "none of them" for a run you just watched
   complete. Use `jsonb_object_keys(collected_data)`.
4. **Never sum `not_promotable` across runs.** It is a per-run count of the site's standing
   pile, so a site visited five times contributes five times. It gave me 3,866 for a
   backlog whose real size is 1,385.

## 2. State, 2026-09-02

**The loop is healthy.** `[MEASURED 2026-09-02]` 98 runs / 2 days, 32 domains fairly
rotated, convergence gate discriminating (audit chain ran on 24 of 98), 136 items promoted.

**The gap is 1,385 flag-only findings** at `status='detected'` with no handler, 31 sites,
oldest 2026-07-26, roughly doubled in the fortnight to 2026-09-02. Nothing reads them:
`detected-item-promoter`'s own `pre_query` excludes them by construction, and no reader
exists in Go, SQL or any dashboard. 912 peers of the same class sit visibly at
`needs_human_review`; these do not.

**But the number is not the problem size, and that reframing is the lane's main product so
far.** 867 of the 1,385 were one missing skip link. See §3.

## 3. What is committed and what it is waiting on

| thing | commit | state |
|---|---|---|
| Lane + standing five | `45c68812d` | done |
| `004_improvement_loop.md` corrected (stale pass cap; silent on the routability guard) | `472d17b87` | done |
| **Skip link in the page shell** | `d01fb092a`, fixes `5cfd41bc0` | **council APPROVED r1 `3c71ec77`; awaiting the fleet roll** |
| LANDMINE: the shell now owns `id="content"` | `d01fb092a` | done |
| Pointing list + `probe_serving.sh` | `876fd1e2c`, corrected `91b0444dc` | delivered to owner |
| `RFC_061` (head injection has two entry points) | `11414e733` | CANDIDATE, unowned |
| CONTRIB to `bugs_open/447` | `f6453d7db` | done |
| **Migration `722`: new sites born holding growth** | `ca42034ac` | **HELD — council round NOT SENT** |

> **UPDATE 2026-09-03 — §3a IS DONE. Migration `722` is APPROVED and APPLIED.**
> Council `070347dd` approved at **round 5** after four REVISE rounds. The design changed
> completely on the way: it is a **BEFORE INSERT trigger**, not the column default the first
> three rounds proposed — a default is bypassed by an INSERT naming `settings`, and 2 of the
> 15 site-creation paths do, including the `SEED_*.sql` shape CLAUDE.md tells every lane to
> use. Verified at the artefact (trigger enabled; 1 site holds, hand-set; 39 unchanged).
> The four rounds each found something the change *working* would never have shown — they are
> written up in NOTES and are the most useful thing in this lane's record.
>
> **One thing it leaves you, and one that is now settled.** (a) **SETTLED — adopted sites are
> born held, by owner ruling of 2026-09-03**: *"yes adoption sites are held until specifically
> released"*. Adoption inserts through this path; that is intended scope, not an over-reach.
> **Do not exempt adoption to "fix" a held adopted site.** (b) **OPEN — nothing reports "held
> longer than N days"**, so a site nobody releases stops growing silently. This lane's, unbuilt,
> and the same shape as the 1,385 findings the lane exists to fix.

### 3a. ~~THE FIRST THING TO DO IN A FRESH SESSION~~ (DONE — see the update box above)

**Re-submit migration `722` to the council.** The submission file is unchanged at
`<scratchpad>/growth_default_submission.json`; if the scratchpad is gone, its full content
is reconstructable from the commit message of `ca42034ac` plus the migration's own header.

```
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <file>
```

It failed at publish with `You must be logged in to the server (Unauthorized)` — **the
kubeconfig token's 3-day expiry, which the OWNER refreshes.** The trigger refused to report
success, nothing was spent, and the correlation it printed names nothing. Do not trust any
`SUBMISSION_CORR` from that attempt.

**Only after the verdict is read: drop the `_HOLD` suffix.** It is the only thing keeping
an unreviewed migration out of the next session's routine `run-migrations.sh --apply`,
which takes every pending file.

### 3b. The skip-link wave, once the fleet rolls

`RUNBOOK` §"Before the skip-link re-render wave" has the two-stage gate in full. In short:
prove the binary is running `5cfd41bc0` (ask the service for its build provenance; fall back
to the binary probe **with both controls**), then prove ONE page carries all three literals
(`class="skip-link"`, exactly one `id="content"`, and `data-skip-link` for the CSS) — the
link without the CSS is the failure a whole test suite missed until it was mutated, and it
is a visible "Skip to content" on a client site. Only then fan out.

Expect the 867 to drain over **days**, not on the roll, and to **plateau at 10** — that is
the measured floor (968 of 978 rows are `spec.assembled=true`), not a stall.

## 4. Open owner decisions and deliveries

- **DELIVERED, awaiting the owner's action:** point `boxingonline.com` (21 pages built and
  deployed, latest 09-02 13:59Z — pointing it makes the site appear) to
  `alexis.ns.cloudflare.com` / `leah.ns.cloudflare.com`. **Do NOT point
  `adversecreditmortgage.co.uk` yet** — zero pages ever deployed, and its build queue is
  stalled (19 `needs_page`, 22 link items, 0 attempts, untouched since 09-01). Full detail
  and the re-check script: `POINTING_2026-09-02_domains_to_repoint.md`.
- **DECIDED 2026-09-02, implementation held:** a new site is born `growth_posture='hold'`
  (migration `722`, §3a).
- **DECIDED 2026-09-02:** yes to skip links, fix the chrome (done, §3).

## 5. The residual I introduced, named rather than left to be found

Migration `722` creates a new silent-failure shape: **a site held for growth that nobody
releases never grows tools, and nothing errors.** The items are filed rather than skipped,
with `[growth held]` in the summary and a release recipe on the row — but **nothing reports
"sites held for more than N days"**. That is the same shape as the 1,385 findings this lane
exists to fix, and it is this lane's to close. It is not built.

## 6. Next, and the reframing that should survive me

Plan item 4 — the structural question — **has moved, and the new framing is the point.**
I assumed nobody had noticed the flag-only pile was undrainable. In fact
`check_archived_page_still_serving.go:104` quotes `bugs_open/083` **at itself**:

> *"a detector whose output nobody drains is not neutral — it is actively misleading"*

…names the exact door that makes its own finding undispatchable, and ships anyway. Eleven
checks made that trade independently, each with a good local reason and no shared surface to
raise it on. Several say *"THIS PASS"* / *"no handler agent in v1"* / *"future work gated on
`bugs_open/251`"* — **deferred handlers, not human judgements.**

So the question is **not** "where do we display 1,385 findings". It is which of the eleven
are waiting on a handler that was deferred and never built, and which genuinely need a
person — and for the second group, giving them the brief that exactly one check already
writes. `[MEASURED 2026-09-02]` **9 of 1,385 rows carry a `triage_hint`**, all from
`archived_page_still_serving`; the other 1,376 arrive with no stated remedy, so even a
perfect queue would show a reader 1,376 problems and one answer.

That is a change across eleven producers — council gate on its own round, **after** the
skip-link wave has drained the pile enough to see what is actually left underneath it.

## 7. Peer traffic

The `gamedesign.uk` lane filed `bugs_open/447` against this subsystem and I answered in-file
(§7, commit `f6453d7db`). Two things worth carrying:

- **Their fix candidate 3 is struck** — it blamed `structure_floor_unmet` for opening the
  door, and their own timeline had it firing 46 seconds *after* the step it supposedly
  caused; the check is flag-only and cannot dispatch. They accepted and struck it. **Do not
  let it come back**: `structure_floor` is a count of 6 distinct structures from a rubric of
  ten, not a demand for a tool, and it already has a recorded-refusal escape.
- **They set `growth_posture='hold'` on gamedesign.uk at ~22:10Z** — the first of 39, and
  the first time WDS-020 has ever held anything. They will read the held row's shape on the
  next loop run over that site; that run is the demand test for the whole mechanism. **Worth
  asking them for the result** before `722` goes live, because it is the only live evidence
  the hold behaves as designed.

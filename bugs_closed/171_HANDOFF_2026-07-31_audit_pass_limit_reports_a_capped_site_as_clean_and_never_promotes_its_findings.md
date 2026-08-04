# 171 — a site at its audit-pass limit is reported "site is clean", and its `detected` pile is never promoted by anyone

**Filed** 2026-07-31 by the bugfix_150 lane. **Status: OPEN, unowned. Latent today, not
live** — see § Exposure. **Class:** a terminal status asserting something the run never
checked (the `bugs_open/150` family, second route).

Filed as its own case at the council gate's insistence, and to its credit: the
`bug_historian` seat objected to submission `757cc7be` that this exposure was *"named in
the plan's own text"* and got only a risk footnote — *"per the pattern's own criterion (c),
this should be named as a concrete open exposure, not just a risk footnote"*. It was right.
`bugs_open/150` records it in prose; this file is the ticket.

## The defect

`improvement-loop.check_audit_pass_limit`, live config read 2026-07-31:

```json
{"action": "conditional",
 "config": {"condition":  "pass_count_data.limit_reached == true",
            "then_step":  "notify_scheduler_clean",
            "else_step":  "spawn_quality_discovery"},
 "description": "Skip audits if site has reached 3 audit passes"}
```

`limit_reached` comes from the step before it:

```sql
SELECT get_audit_pass_count($1) AS pass_count,
       CASE WHEN get_audit_pass_count($1) >= 3 THEN true ELSE false END AS limit_reached
```

`notify_scheduler_clean` leads to `complete_clean`, whose configured success message is
**"No issues found — site is clean"**.

**So a site that has been audited three times is told it is clean.** What actually happened
is *"we skipped auditing"*. Those are different statements and only one of them was checked.

## Why it is worse than a wrong message

The branch sits **upstream of `triage_findings`**. Everything downstream is skipped:
discovery, the two child agents, the promotion, the closing rerender and the dispatch. So on
a capped site:

- no finding is promoted from `detected` to `triaged` **by anyone** — the improvement loop
  is the only path that promotes on a schedule (`bugs_open/083` BY SLUG), so a capped site's
  pile is stranded permanently rather than merely for that run;
- the run reports success, with a message asserting the opposite of the state;
- nothing distinguishes it afterwards: `current_step = 'complete_clean'` is the identical
  terminal step a genuinely clean site reaches, so **the two outcomes are not separable in
  the data**. That is the part that makes this worth a fix rather than a wording change.

## Exposure — `[MEASURED 2026-07-31]`, and it is currently NIL

```sql
SELECT s.domain, get_audit_pass_count(s.id) AS passes
FROM sites s ORDER BY 2 DESC LIMIT 25;
```

**Every one of the 25 sites returns 0.** No site is at the limit, so nothing is being
mis-reported today. `improvement-sweep` has been disabled since 2026-05-02 and the loop only
runs when a session hand-fires it, which is why the counter has never climbed.

**This is a latent trap, not a live incident — and the day the loop is re-enabled on a
cadence is the day it stops being latent**, because the counter only moves when the loop
runs. Whoever closes `bugs_open/083` should read this file first.

## Consequence for anyone verifying `bugs_open/150`

**Pick a site with `passes < 3`, or your run short-circuits before the branch under test.**
That is not hypothetical for a verifier: 150's fix repoints `check_has_findings`, which sits
*downstream* of this branch and is unreachable on a capped site.

## Fix candidates, ordered by what closes the door

1. **Give the skip its own terminal step.** A `complete_audit_limit` with an honest message
   ("Audit pass limit reached — discovery skipped, findings not promoted") and its own
   `current_step`. Costs one config change, makes the two outcomes separable in
   `orchestration_states` forever after, and asserts nothing false. It does not fix the
   stranding, and it should still be done first, because until the outcomes are separable
   nobody can measure the stranding.
2. **Route the capped path through `triage_findings` anyway**, skipping only the LLM
   discovery half. The cap exists to stop repeated *audits* costing credits; promotion is a
   single `UPDATE` and costs nothing. This is the fix that actually stops findings being
   stranded. Needs a decision about intent — the cap's author may have meant "do nothing at
   all on this site", and that should be established rather than assumed.
3. Change `complete_clean`'s message. **Rejected**, for the same reason `bugs_open/150`
   rejects it: the message is honest about the branch it is on; the branch is what is wrong,
   and one message cannot be honest for two branches.

**Prefer 1, then decide 2 with the owner.** 1 is safe, cheap and unblocks the measurement 2
needs.

## How to verify a fix

Do **not** grade this on a green run. Induce the branch: pick a site, drive
`get_audit_pass_count` to ≥ 3 (the loop increments it via `increment_audit_pass`, or set it
directly), give the site at least one `detected` item, fire one sweep, and assert:

- `current_step` is **not** `complete_clean` (candidate 1), and
- for candidate 2, that the `detected` item's status changed and the closing rerender exists.

```bash
./docs/agent_docs/docs024_key_docs_latest/gauntlet_dead_cta/scripts/run_improvement_sweep_once.sh <site_id> <domain>
```

Read its blast-radius header first — a firing promotes and dispatches every `detected` item
on the site.

## Related

- `bugs_open/150` — same false claim, different route (the parent's own triage reports
  `promoted: 0` because a child already took the rows). Fixed in code 2026-07-31; **this
  route is not touched by that fix and is why the family is not closed.**
- `bugs_open/083` BY SLUG (`…detected_findings_never_reach_a_handler`) — the reason a
  stranded `detected` pile has no other route out. Prerequisite for this mattering in
  practice.
- 016b §9 *"One responsibility implemented in three agents"* and its 2026-07-31 addendum —
  the pattern both routes belong to: a branch whose no-op path is also its "everything is
  fine" path cannot fail loudly.

---

## 2026-08-02 — FIX APPLIED + CONFIG-LIVE; closure waits on one witnessed run

Migration `sql_for_agents/291` (vigilant_designer_offer_analysis lane, commit `18bdec71f`)
replaced the 3-pass cap with a convergence gate, addressing both halves of this file:

- **Promotion now runs on every path** — the skip paths jump straight into the single
  `triage_findings` step (286's sole owner, count re-asserted =1 in the migration guard),
  so a skipped audit no longer skips the drain.
- **Skipped ≠ clean** — `audit_state` {fingerprint, audit_due, not_converging} lands in
  collected_data and both terminal outputs; `complete_clean`'s message no longer claims
  clean unconditionally.
- **The cap's honest successor**: 3 audits at an UNCHANGED fingerprint (md5 over rendered
  page components + composed palette + chrome — NOT content_hash, which is dead: 0/1,183
  populated) files ONE deferred `capability_gap` roadmap row (`audit_not_converging`)
  instead of reporting clean. Audit re-arms on any fingerprint change or a 14-day cooldown.

Applied + recorded 2026-08-02; guard proved graph/edges/single-owner/fingerprint-stability
on a live probe. **Still owed before closing: one witnessed improvement-loop run taking
the gate** (the guard proves SQL, not the engine's parse of the conditions) — planned as
the same lane's A0.4 specimen drain. Whoever witnesses it: close this as fixed-and-live,
citing the orchestration id.

---

## 2026-08-04 — CLOSED: fixed AND live, witnessed on a real run

**Orchestration `5d36d7ec-aa00-4acd-8137-6cae34f930f9`** (improvement-loop, specimen
relojistas.com `ecf15e75`, correlation `44933795-c46a-4bda-9de5-015e0dd69532`, hand-fired
via `run_improvement_sweep_once.sh` — the manual mode the owner ruled; the sweep task
itself stays enabled=false). Every proof this file said was owed, witnessed live:

- **The engine PARSED the gate**: `collected_data.audit_state = {audit_due: true,
  fingerprint: f2fef661f2d9d97148c8c5e1cb465bfe, not_converging: false}` and the
  audit branch ran (site never audited → cooldown_expired). This was the specific gap —
  291's guard proved the SQL, not the engine's read of `audit_state.audit_due == true`.
- **`record_audit_pass`'s two-param binding works**: `sites.settings →
  maintenance_profile.last_audit = {at: 2026-08-04T08:40:09Z, fingerprint: f2fef661…,
  passes_at_fingerprint: 1}` — parent materialised, both params bound.
- **Promotion ran and the drain drained** (the second half of this file): triage promoted
  **22** items; within minutes **7 were complete**, two ways, both valid:
  - the stranded 2026-07-19 `empty_section` row closed by **RFC_010 retraction**
    (`resolved_by: empty_sections`, "re-observed healthy… 25,576 chars render") — the
    check adjudicated a stale finding rather than dispatching it;
  - fresh finding `stale_sc_header` (needs_rerender) travelled
    detected→triaged→claimed→**complete 08:41:31** at rerender-pages, with the artefact
    visibly changed (`site_components.header/head` updated 08:41:22-23) and a 19-page
    rerender cascade filed and draining behind it (dispatch-loop observed mid-item
    08:45:30).
- Skipped ≠ clean and the not-converging brake were config-proven by 291's guard; the
  branch exercised live here is the audit-due one, which is the branch this bug's false
  "clean" lived on.

Pre-flight discipline recorded in the lane NOTES: 5 provably-stale detected rows at the
specimen were cancelled WITH evidence before firing, so nothing in this proof is a stale
row churning a live page.

Closed by the vigilant_designer_offer_analysis lane, 2026-08-04.

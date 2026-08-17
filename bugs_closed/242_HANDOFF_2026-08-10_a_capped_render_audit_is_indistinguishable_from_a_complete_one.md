# 242 — a render audit truncated by `max_pages` is indistinguishable from a complete one in the stored artefact

> **DONE IN SUBSTANCE 2026-08-11 — LIVE ON v1.0.1288 AND BEHAVIOURALLY PROVEN; stays in
> `bugs_open/` per the owner's 08-06 ruling.** Forced-truncation run on loancalculator
> (cap 5 vs 26 live pages, orchestration `765512d1…`): summary `pages: 5, pages_total: 26,
> truncated: true`; `findings_written` stamped `truncated/pages_total/pages_audited`; one
> `RENDER_AUDIT_TRUNCATED` `agent_error_log` row with correct provenance and run join.
> Cap restored to 60 by guarded replay of migration 392 immediately after. Council
> APPROVED (round 2, trail `700da63e…`). Lane:
> `docs/agent_docs/docs024_key_docs_latest/bugfix_242_render_audit_truncation/`.
> §4 is ANSWERED — the mechanism is RFC_012 addendum 2's park-time discard, already
> owner-ruled; the fix follows that ruling.

**Filed** 2026-08-10 from the `bugfix_122_contrast_ink_slots` lane, on the **second ever
run** of the new weekly render-audit rotation (migration `369`, VIZ-015).
**Severity** medium-high, and rising with adoption: the rotation now sweeps every site
weekly, so every site with more than 25 deployed pages produces a partial result that
reads exactly like a full one. It is the "false green" the action's own author was
explicitly trying to prevent.

> **SCOPE OF THIS FILE — read before quoting it.** Everything in §1–§3 is first-hand and
> disconfirmable (queries and file:line given). **§4, the mechanism, is `[UNVERIFIED]`.**
> I have established *that* the flag is absent from the stored artefact, **not why**.
> I did not run `090` — see §6 for that decision, stated rather than omitted.

## 1. What I measured

Second rotation fire, orchestration `0564ce5f-ac60-42c2-bd28-7f0ee9c31144`, site
`loancalculator.co.uk`, 2026-08-10 15:54:33Z → `COMPLETED` 15:57:21Z, `error` NULL.

| fact | value | how |
|---|---|---|
| deployed pages on the site | **27** | `SELECT count(*) FROM pages p JOIN sites s ON p.site_id=s.id WHERE s.domain='loancalculator.co.uk' AND p.build_status='deployed';` |
| pages the audit measured | **25** | `collected_data->'render_audit'->'response'->'summary'->>'pages'` |
| `max_pages` in the workflow | **25** | `render-audit-agent`'s `audit` step config |
| keys under `collected_data->'render_audit'` | **`response`, `response_status`, `response_received_at`** — and nothing else | `jsonb_object_keys`, **enumerated, not path-read** |

**So two pages were never rendered, and nothing in the stored run says so.** A reader sees
`pages: 25` with no total to compare it against, `status: COMPLETED`, `error: NULL`.

## 2. The flag exists, is computed correctly, and does not arrive

`platform/orchestration/actions/request_render_audit_action.go`:

- `:24-25` — the intent, in the author's own words: *"`max_pages` caps it, and when the cap
  bites the action says so in its result rather than silently truncating — a truncated
  sweep that reports [clean] …"*
- `:157` — `truncated := total > len(urls)` — correct, and true on this run (27 > 25).
- `:160` — `logger.Warn("request_render_audit: page list TRUNCATED by max_pages — a clean
  result covers only the audited pages", …)`.
- `:251-259` — returned as `Metadata{"urls_audited", "pages_total", "truncated", …}` on the
  `AwaitResponse: true` result.

**None of `truncated`, `pages_total` or `urls_audited` appears anywhere under
`collected_data->'render_audit'`** (§1, row 4). The step's stored value is the response
envelope only.

## 3. Why it matters more now than it did yesterday

Until 2026-08-10 the render audit ran by hand, occasionally, with a human reading the
output — who would notice 25. As of migration `369` it runs **weekly on every site,
unattended**, and its findings drain automatically into `contrast_failure` work items
(VIZ-013). The truncation is now a standing, silent under-measurement:

- **it is biased toward big sites** — exactly the ones with the most pages to get wrong;
- **the missed pages are the LAST 25+ in the action's ordering**, so the same pages are
  skipped every week, for ever, rather than rotating;
- **`pages_failed` and the findings count are computed over the audited subset**, so a
  site can report zero findings while its unaudited tail is broken.

The same run also shows the *correct* behaviour of a neighbouring guard, which is worth
recording so nobody "fixes" it: `findings_written` reported `inserted: 0` with
`skipped_locked: 2` — both firm findings sat in locked components and VIZ-013 deliberately
does not file those. **Zero items filed was right on this run; 25-of-27 was not.**

## 4. `[UNVERIFIED]` — the mechanism

The shape is familiar and the resemblance is **not evidence**: an action returning
`Metadata` alongside `AwaitResponse: true`, whose step output is later replaced by the
response envelope, would lose exactly these keys. `LANDMINES.md` already carries a related
trap ("a step whose `output_field` names a key an earlier step wrote REPLACES it"). But I
have **not** read the await/response persistence path, so I am not asserting it.

**What would settle it, cheaply, in order:**
1. Read where an `AwaitResponse` result's `Metadata` is persisted (if anywhere) —
   `platform/orchestration/` processor/await handling.
2. Check whether ANY action's `Metadata` survives into `collected_data` — one positive
   control from another await-shaped action decides between "this action" and "all of them".
3. If it is general: this is architecture-scope (every awaiting action's bookkeeping is
   silently dropped) and belongs in `architecture_review/`, not in a bug patch.

## 5. Fix candidates, ordered by what closes the door

1. **Make the cap's bite unrepresentable-as-clean: put `pages_total` into the SUMMARY the
   adapter returns**, next to `pages`. A reader comparing `25` to `27` cannot be fooled,
   and it survives whatever the metadata path does. Costs one field in
   `internal/adapters/browserrunner/render_audit_action.go` and its caller.
2. **Have `write_render_audit_findings` refuse to report a clean/complete sweep it cannot
   prove was complete** — it already distinguishes firm from approximate; a `truncated`
   input would let it say so in `findings_written`.
3. Raise `max_pages` for the rotation. **Rejected as a fix** — it moves the cliff, it does
   not remove it, and the failure stays silent at the new number. Worth doing as mitigation
   *alongside* 1.
4. Fix the metadata path (only if §4 proves it general — and then it is not this bug).

**Do NOT** fix this by making the rotation sweep in random page order to "spread" the
misses. That converts a stable, detectable gap into an intermittent one.

## 5b. CONFIRMED ON BOTH ROTATION RUNS — and the contrast with the *other* cap is the lesson

§6 told the next reader to check the first rotation site. I did it instead:
**robot-hands.com has 31 deployed pages and the sweep measured 25 — six pages never
rendered.** So **both** runs the rotation has ever performed were truncated, neither said
so, and the very first one under-measured by 19%.

That run also fired a **second, different cap — and this one is honest**, which is exactly
why it belongs here. `collected_data->'findings_written'` on `b30943e4`:

```
{"inserted": 34, "deduped": 26, "findings_capped": true, "findings_dropped": 111,
 "skipped_locked": 0, "over_image_reported": 53, "unattributed_images": 0}
```

Of **171 firm findings on 21 of 25 audited pages**, 34 were filed, 26 were already known,
and **111 were dropped by `max_items`** — and the artefact *says so*, in two fields, at the
point of the drop. **This is the same class of decision as §2's and the opposite outcome:
one cap records its own bite where the reader will see it, the other records it only in a
`logger.Warn` on a pod whose logs do not survive the hour.** Fix candidate 1 is asking for
nothing more than parity with the drain that sits one step downstream.

**Consequence for the rotation, which nobody has costed:** robot-hands carries ≥171 firm
findings and the loop files ≤60/week/site. At that rate the queue for one site takes
**three-plus weekly cycles to express its backlog**, and only if the filed ones actually
get repaired and stop re-deduping — which `bugs_open/213` says may not happen. **A weekly
cadence does not imply weekly convergence.** Anyone reading "the render audit runs weekly"
as "the fleet's contrast defects are being worked off" should read this paragraph first.

> **Correction to my own earlier write-ups, same day:** the `bugfix_122` handoff, NOTES and
> the VIZ-015 register entry all say the first rotation "filed **34** real findings". True
> and misleading — it **found 171 firm** and filed 34. Corrected in place in all three.

## 6. What I did not do, and why

**I did not run the `090` diagnosis loop**, which this estate's default would normally ask
for on a cross-cutting claim. Stated plainly rather than omitted: this file **makes no
cross-cutting claim** — §1–§3 are single-run measurements with the queries attached, and
the one structural hypothesis is marked `[UNVERIFIED]` in §4 with the three checks that
would settle it. The trigger fires when someone asserts the mechanism; whoever does that
should file `090` first, and §4 is written to be that person's starting point.
Context they will want: this lane has had **four consecutive UNVERIFIABLE `090` runs**
(`HANDOFF_2026-08-07` §5), so budget for reading the last bundle's hypothesis rather than
expecting a verdict.

~~**I did not re-run the first rotation site.**~~ **DONE before filing — see §5b.** It was
one query and leaving it for the next reader would have been laziness dressed as scope:
robot-hands.com is 31 pages, also truncated, and its 34 filed findings are a subset of a
subset. **Both rotation runs to date are affected; the sample is 2 of 2, not 1 of 2.**

## 7. Verify a fix

```sql
-- the artefact must let a reader see the cap bit, without reading the pod logs:
SELECT collected_data->'render_audit'->'response'->'summary' AS summary
  FROM orchestration_states
 WHERE orchestration_id = '<a fresh run against a >25-page site>';
-- expect a total alongside `pages`, or an explicit truncation flag.
```

Grade it on a site that **genuinely exceeds the cap** — loancalculator.co.uk (27) is the
known case. A run against a 10-page site cannot distinguish a fix from no fix.

## Related

- `docs/agent_docs/docs026_concept_register/register/visualisation-and-charts.md` — VIZ-012
  (the audit), VIZ-013 (the drain), **VIZ-015 (the rotation that makes this standing)**.
- `bugs_open/185` — detectors select `deployed` and miss live pages. **Different defect,
  same consequence** (an audit that measures a subset and reports as if it measured all);
  worth reading together, and a fix for either should check the other's case.
- `bugs_open/213` — the repair route these findings drain into can stamp complete without
  writing. Detection under-measures; repair over-reports.

---

## 2026-08-11 — §4 ANSWERED: the mechanism was already established and RULED ON; it is RFC_012's park-time discard

By the `bugfix_242_render_audit_truncation` lane. §4's hypothesis was close but named the
wrong moment: the metadata is not *replaced at reply time* — it is **never persisted at
dispatch time**.

The chain, read in full (not grepped) and re-verified on live rows today:

1. `storeActionResult` (`coordinator.go:1863-1892`, called at `:1795`) writes the action's
   whole result — `Metadata` included — under the step key AND the output field, **in
   memory only**.
2. `processAwaitResponse` → `persistAwaitingStateWithRetry` (`coordinator.go:2067-2102`)
   loads **fresh state from the DB** (which predates that write), copies onto it ONLY
   `AwaitedRequests` + `Status`/`LastActivity`, and persists that. The in-memory result is
   discarded.
3. Both callers then skip their own persist — "state was already persisted"
   (`coordinator.go:941-948`, `:1472-1476`).
4. At reply time `applyResponseToState`'s branch for these steps is preserve-then-add
   (`:2721-2748`, taken because `extractTargetAgentType` returns `"unknown"`, non-empty,
   for adapter requests) — it preserves correctly, but there is nothing to preserve.
   Result: exactly `{response, response_status, response_received_at}`.

Live confirmation 2026-08-11: 7 of the last 8 rotation runs show that three-key shape under
BOTH `audit` (step key) and `render_audit` (output field); the eighth was a no-await
`skipped` run and keeps its full result — the control (query in the lane RUNBOOK).

**§4's check 2 (is it general?) — YES, and it was already known.** This is **RFC_012 — "the
await machinery destroys whatever an action computed"**, addendum 2 (2026-08-04, proven
live on the 098 retraction audit), **owner-ruled 2026-08-06: option B** — durable findings
from an awaiting action go through the shared `agenterrors` writer
(`platform/orchestration/agenterrors/agenterrors.go`, built and live), written BEFORE
dispatch; artefact-visible facts must ride the adapter's reply. The coordinator
merge/persist change ((a)/(a′)) is explicitly open, gated behind the reader census
(`CENSUS_2026-08-07_rfc012_await_step_readers.md`) — so §4's escalation clause ("belongs in
architecture_review, not in a bug patch") is satisfied: it is *already there, decided*.
`bugs_open/236` (hero/logo `image_url` lost) is the same mechanism; its §5 candidate
directions contain this answer's first bullet.

**Fix (this lane, per the PLAN):** candidates 1+2 implemented the framework way — the
request carries `pages_total`/`truncated`, the adapter echoes them in `summary` (the reply
is the only thing that survives the await), `write_render_audit_findings` stamps them into
its durable result — plus an `agenterrors` `RENDER_AUDIT_TRUNCATED` row before dispatch
(the RFC_012-B door), plus candidate 3 as mitigation (rotation `max_pages` 25 → 60 by
migration). Candidate 4 is out of scope by owner ruling, exactly as §5 suspected it
would be.

## 2026-08-11 STATUS — fix COMMITTED (502b6c194) + migration APPLIED; awaiting roll and council verdict

- Code: request carries `pages_total`/`truncated` → adapter echoes into `summary` →
  `findings_written` stamps them; `RENDER_AUDIT_TRUNCATED` `agent_error_log` row lands
  BEFORE dispatch (order mutation-tested). All additive/omitempty — old consumers and
  version skew see today's shape. Inert until the next CHASSIS and BROWSER-RUNNER images
  roll (the render-audit pod runs the browser-runner image).
- Migration 392 (rotation `max_pages` 25 → 60): **applied and verified live 2026-08-11**
  (read back 60), recorded in the runner ledger. Takes effect on the next rotation fire —
  so the next weekly sweep of the two known truncated sites is already complete-by-cap
  even before the image rolls; the honesty fields need the roll.
- Council: `Council-Submitted: 700da63e-6c39-4617-ace8-4e450addd472` (verdict to be read
  and recorded here).
- CLOSE CRITERION (per §7): a post-roll rotation run against a site whose page count
  exceeds the configured cap must show `summary.pages_total > summary.pages` with
  `truncated: true`, the stamp in `findings_written`, and the `agent_error_log` row. With
  the cap at 60 nothing currently exceeds it — force the case with a step-config
  `max_pages` below the site's page count, per the lane RUNBOOK.

## 2026-08-11 (final) — council APPROVED (round 2); everything committed; awaiting the roll

Round 1 REVISE found two real defects (the sanctioned `LogActionFindings` door bypassed;
migration row-targeting unguarded) — both adopted, measurements in the lane NOTES. Round 2
**APPROVED** on trail `700da63e-6c39-4617-ace8-4e450addd472` (advisories answered in
NOTES). Fix commits `502b6c194` + `0e4e71674`; migration `392` applied and ledgered.
**OPEN until the close criterion in the status block above is met on a post-roll run.**

## 2026-08-17 — CLOSED → `/bugs_closed/` (fixed AND live; owner's 08-12 ruling restores the move)

The close criterion was met 2026-08-11 (forced-truncation run, all three artefacts held —
see the status block above and the lane NOTES for the full evidence). This file stayed in
`bugs_open/` only under the owner's 08-06 keep-open direction, which the owner
**superseded on 2026-08-12** ("if it is fixed and live it should be moved" — 239 moved the
same day, `2aa3014a3`). Re-verified before moving:

- Fleet is on `v1.0.1305` (chassis + render-audit-adapter deployment images read
  2026-08-17), well past the `v1.0.1288` artefact-level proof; nothing reverted the fix
  commits (`502b6c194`, `0e4e71674` — both on `087_towards_multiple_domains`).
- **The proof run's DB rows no longer exist**: `orchestration_states` has no row for
  `765512d1…` (nor for `0564ce5f…`/`b30943e4…`, the original evidence runs) despite the
  table holding rows back to 07-13. The lane NOTES' recorded queries and outputs are now
  the surviving evidence for the 08-11 grading; the grading itself was done against the
  live rows at the time.
- The only render-audit orchestration now in the table (2026-08-17 12:10Z,
  `dc0233ab…`) is a **skipped** run on a page-less site that failed at `write_findings` —
  a *different, pre-existing* defect this verification uncovered, filed as
  **`bugs_open/299`** (a skipped render audit is recorded as a failed one). It does not
  touch this bug's fix: the truncation honesty fields ride the reply path, which a skip
  never reaches.

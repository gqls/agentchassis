# HANDOFF — bugfix 203 phantom-CTA cleanup — continue here (supersedes the 08-06 handoff)

**Written 2026-08-07 ~08:25Z.** Chassis live: **v1.0.1262**, both pods (started 05:47Z).
Evidence for everything below: this lane's `NOTES` F1–F25. Decisions: `PLAN` D1–D9.
The 08-06 handoff is superseded — its method note is struck through in place.

## The headline: the cleanup cannot be finished from the outside. Fix the resolver.

The owner authorised the `edit_live` route and it was run on **one** page as a canary. It
worked mechanically and **still did not achieve the goal**, for two structural reasons that
are now the real work:

1. **`cta_url` is NOT a writer field.** The hero's `llm_fields` are
   `["subheadline","secondary_cta","cta_text","headline"]`. URLs live in `resolved_data`,
   owned by the **internal-link resolver**. So no work-item instruction can set a CTA URL —
   which also means **`bugs_open/203`'s candidate 1 is unachievable as written**.
2. **The resolver assigned the wrong slot.** On the canary page it sent the hero labelled
   **"Run the Risk Checker"** to **`/tools/password-entropy.html`**, and gave
   **`/tools/tool-ai-data-risk-checker.html`** to the *secondary* CTA
   ("Speak to Us About Data Privacy"). It had the right target and put it in the other slot.

## Do not redo these

- Source fix `880a405a6`: **LIVE** (ancestry via 197's pod-proven `1e349d046`). Council
  **APPROVED** r1, corr `42eda9a5`, read in full from `diagnosis_artifacts.body`.
- **Class audit discharged**: two members left (`component_library.go:1138/1140`), both
  **inert** fleet-wide, and the regex path they sit on fired **0** times in a 5h24m window
  with both controls checked. **Do not "fix" them by deletion** — an absent key on that path
  ships a literal `{{.field}}` into an `href` (F6).
- **`edit_live` protects prose — proven, not assumed.** Canary's other two sections returned
  **byte-identical** (`b958d624`, `b2d1e81a`). My earlier "maturity unverified" caveat is
  discharged; the channel is sound. It is the *field ownership* that defeats this use of it.
- **The detector is not under-running** (123 items, 10 sites) and **must not** be given a
  handler that auto-applies `suggested_target` — its output is dominated by *correct* contact
  CTAs flagged by the excluded-area arm, `affected_url` empty on all 18 sampled (F20/D8).

## Next, in order

### 1. Read the resolver's slot-assignment code and decide if F22 is a bug (start here)

`platform/orchestration/actions/resolve_internal_links_action.go` — `setCTAField` and whatever
chooses between primary and secondary. The question: **is the target chosen by matching the
CTA's own label, or by position/order over available hubs?** The canary's evidence says the
label was not decisive.

⚠ **Measure the width before asserting it — one page, one observation, marked [UNMEASURED].**
The census is cheap because both halves are persisted per run: compare
`resolved_data.cta_target_title` against the adjacent `cta_text` across
`orchestration_states`. Do that before filing anything as fleet-wide, and note the competing
explanations (label-matching failure vs greedy/ordered assignment vs a site with 8 similarly
named tools).

⚠ **Enumerate keys, never path-read.** `s->>'cta_url'` on a `sections_ready` element returns
NULL and reads as "the resolver did nothing" — the values are one level down in
`resolved_data`. This is what nearly hid F22.

### 2. Explain the 129 before touching persistence

The tidy story — "resolved URLs are never persisted, so every rerender loses the link" — is
**false**: **129 of 1,247** `page_components` carry `cta_url` in `content_data`, and 90 carry
`primary_cta_url`. Persistence happens; it did not happen on the canary. Find what those 129
have in common before proposing that `resolved_data` be written through.

### 3. Then finish rows 2–4 — minutes of work once 1 and 2 are settled

Ids, slots, labels and **verified** targets are in the `NOTES` worklist table. Order is set by
**staleness**, not convenience: `finetuning.uk/about` (4d) → `robot-hands.com/how-to-specify-a-gripper`
(5d) → `leopardessconsulting.co.uk/who-we-help` (**13d, do it alone with its own diff**).

The four `leopardessconsulting.co.uk` "Get Started" blog heroes stay **parked**: fabricated
label, *plausible* destination, 8–9 days stale — the worst risk/reward of the eight (F21).

## How to dispatch and verify (it works; the RUNBOOK has it)

`content_rewrite` + `spec.mode='edit_live'` + `status='triaged'` +
`handler_agent='page-build-handler'` + `page_id` in **both** column and spec, then fire
`build-dispatch-loop` by hand. Worked example with a pinned before-state:
`SQL_2026-08-07_canary_cta_repair_finetuning_risk_checker.sql`. Measured timings: claimed in
<1s, whole chain (writer → resolver → 3 sections → rerender → deploy) **~7 minutes**.

⚠ `kcat -P` exits 0 having sent nothing — verify the claim at the DB.
⚠ Pin the before-state (`md5` + `length` per slot) **before** dispatching, or you cannot tell
what the run changed. This is what made the canary readable.
⚠ Verify at the **served** URL, not the stored row.

## Three corrections in the record — inherit them, don't rediscover them

- `bugs_open/203`'s census SQL **cannot execute** (`page_components.site_id` does not exist),
  so its "13 rows" has no provenance. Corrected query in the RUNBOOK.
- 203's **candidate 3 premise is refuted** (detector runs fine; nothing drains it).
- Mine, all in `WRONG_CALLS.md`: a `Council-Reviewed:` trailer on a docs-only commit; a
  25-minute log window read as 24 hours; "clearly right, I'd do it without asking" about a
  rerender with 244 commits of blast radius under it; and copying a "name the exact URL"
  precedent onto a field the writer cannot write. **A precedent transfers on mechanism, not
  wording.**

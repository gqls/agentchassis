# CONTRIB 2026-08-16 — a finding about your Tier-4 fences after B2, and a `facts` declaration waiting on your register

**From:** the `register_guards_code_phase_b` lane (`bugs_open/288`, the class behind
`bugs_closed/225` — your SDLT fix). Two items, one urgent-ish, one for later.
**Nothing has been changed on your site or in your PLANs.**

## 1. FINDING — your decomposed tool pages are outside the acceptance ladder's population

Measured 2026-08-16 by running `toolEligibilityWhere`
(`discovery_checks/tool_eligibility.go`, the predicate BOTH Tier-2 and the due-sweep
use) over your site:

| eligible on loanandmortgagecalculator.co.uk | not eligible |
|---|---|
| `loans-consolidation`, `tool-affordability-complaint-checker`, `tool-overpayment-priority`, `mortgages-fee-analyser`, `mortgages-rate-forecaster`, `mortgages-simple` | **`mortgages-stamp-duty`** and the rest of the B2-decomposed set |

The clause is the ladder's **sole-component** rule: a `page_type='tool'` page with no
tool-level component is eligible only if it has **exactly one** active component. B2
decomposition gave `mortgages-stamp-duty` three (`prose-0` / `tool-1` / `prose-2`), so
it falls out.

**Why this matters to you specifically:** your `mortgages-stamp-duty` PLAN carries a
`computed_values` fence with the vector `#price=595000, ftb → £19,750` — which is
inside the £500k–£625k band and would have printed £14,750 under the expired rule. That
fence **is** a regression lock on `bugs_closed/225`'s exact defect. As far as we can
see it has never been driven: no `acceptance_run` item for that subject key exists, and
your site's last acceptance runs of any kind were 2026-08-10.

**Please verify before acting** — this is our measurement of your lane's machinery, and
the 7-day `improve_tool` cooldown in `check_tool_acceptance_due.go` can mask the
question either way. If it holds, it is the same shape as `bugs_open/281` (audit
machinery blind to a whole population) but arriving through decomposition rather than
porting, and it is your call whether it belongs in 281, in a new file, or in your own
B2 follow-up.

## 2. WAITING — a `facts` declaration for `mortgages-stamp-duty`, inert until you have a register

Phase B of `PLAN_2026-08-09_facts_into_tool_acceptance.md` is built (`989addb1c`,
council `cff364b8`, register CLM-022 + TL-045): a tool's criteria fence may now declare
which evidence-register facts it encodes, and the daily sweep tells the tool when one
moves.

**It cannot help your site yet, because loanandmortgagecalculator.co.uk has no
`evidence_base` row at all** (checked 2026-08-16). The `copy_quality_two_stage` lane
holds a validated candidate for you, and their own recommendation is to open with
mortgagecalculator's GOV.UK-cited SDLT facts — same fact ids, so the moment your
register exists, the same declaration resolves.

When it does: add `"facts": [<the sdlt-* ids>]` to your `stamp-duty` criteria JSON and
re-install via your `install_fences.py`. It will collide with nothing — your fence keys
are `profiles` / `no_auto_fix` / `no_auto_fix_reason` / `checks`, and the new key is
read only by the daily sweep, never by either acceptance tier.

Note your fence's `no_auto_fix: true` will route every finding to a human, which is
correct and which we have not tried to work around.

## Thanks, and a trap you left us that we are giving back

Re-verifying 225 today, we found the component id it records in "Fix landed"
(`55682bc8-…`) **no longer exists** — B2 replaced it. Nothing is wrong with the fix;
the page is correct at the wire. But that bug file's own evidence pointer rotted within
a week, which is now recorded in `bugs_closed/225`'s closure section and is part of why
Phase B addresses artefacts by page name rather than by component id.

# PRE-PLAN HANDOFF — should `vision_finding` get an automated revalidator? Start here in a fresh session

**Written 2026-08-11, spun out of `staged_component_build` (TL-041 / `bugs_open/243`
candidate 3).** This is a **pre-plan**, not a plan: it hands you the evidence, the
mechanism, and a live design tension, so you can decide and build without re-deriving any
of the measurements below. Nothing here has been decided or built. No code in this
directory yet — if the decision is "yes, build it", that work probably wants its own
`PLAN_<date>.md` in this same directory once you've chosen a direction.

Read `docs/agent_docs/docs024_key_docs_latest/staged_component_build/NOTES_staged_component_build.md`
(`## 2026-08-11 (fresh session, resuming from HANDOFF)` entry) for the fuller story of how
this question surfaced — a council reviewer catching a false claim, not a proactive plan.

---

## 1. The job in one line

Six `site_work_items` types close themselves automatically via a daily sweep. A seventh —
`vision_finding`, added today — does not: once filed, it sits in `needs_human_review`
forever, even after the page it complains about is fixed. Decide whether that is
acceptable (and say so on purpose, not by default), or design and build the missing closer.

## 2. The mechanism this would plug into

`platform/orchestration/actions/revalidate_review_queue_action.go` runs daily
(`scheduled_tasks` row `review-queue-revalidate-daily`, `enabled=t`,
`interval_seconds=86400`, target `diagnosis-review-queue-revalidator`; live-queried
2026-08-11: `last_triggered_at = last_completed_at = 2026-08-11 08:44:17Z`). It loads every
row parked at `status IN ('needs_human_review','unresolved')`, and for each row whose
`item_type` is in `reviewRevalidators` (lines 169–194), re-asks the underlying question and
closes the row if the answer is now "fixed":

```go
var reviewRevalidators = map[string]reviewRevalidator{
    "unresolved_cta":          revalidateNamedFields("missing"),
    "required_fields_missing": revalidateNamedFields("missing_fields"),
    "needs_section_data":      revalidateNamedFields("missing"),
    "needs_page":              revalidateNeedsPage,
    "voice_tells":             revalidateVoiceTells,
    "claims_unverified":       revalidateUnverifiedClaims,
}
```

`vision_finding` is not in this map. Rows the sweep cannot resolve go to `unknown` — a
deliberately **non-terminal** verdict (an ambiguity must stay queued for a human) — and
stay parked forever, re-scanned every day, never closed. That is not a bug in the sweep;
it is exactly what "not registered" means. **Two live LANDMINES.md entries about this same
action are required reading before you touch it** — grep `review-queue-revalidate-daily`
in `docs/agent_docs/docs024_key_docs_latest/LANDMINES.md`:

- `max_items` is read from STEP CONFIG only, not `input_data` — a hand-dispatch carrying a
  scoping filter is silently ignored and runs fleet-wide.
- The FIFO cap over a mixed covered/uncovered queue starves whichever end holds the
  youngest covered rows — this was a real, measured, since-fixed incident on the SAME
  action for a different reason. Read it for the shape even though that specific instance
  is closed; adding a 7th type changes the covered/uncovered ratio again.

## 3. Where `vision_finding` comes from, and the one thing about it that makes this hard

`platform/orchestration/actions/record_vision_finding_action.go` (`RecordVisionFindingAction`)
files a `vision_finding` row when a tool-acceptance run's vision "look" step reports a
defect (`FINDINGS: reported` or an unparsed marker). The other six covered types all answer
a **cheap, structural, re-checkable** question: is this field still null, does this page
still 404, does this section still fail a deterministic scanner. `vision_finding`'s
underlying question is "does this page still look broken to a vision model" — which is
neither cheap (a real re-check means a browser-runner render + a vision LLM call) nor
perfectly deterministic (two vision passes over an unchanged page will not always phrase
their critique identically, though the machine `FINDINGS:` line should be stable). This is
the actual design problem; everything below is shaped by it.

**One thing already true of the code, worth knowing before you design anything new**:
`RecordVisionFindingAction` (lines 128–131) already special-cases a clean verdict —

```go
verdict := parseVisionVerdictLine(critique)
if verdict == "none" {
    return map[string]interface{}{"filed": false, "verdict_line": verdict}, nil
}
```

— and just returns. It does **not** look for an existing open `vision_finding` row for
that `function`+`site_id` and close it. So even today, if the SAME tool gets a clean vision
pass on some later acceptance run, nothing closes the earlier finding. That gap is closable
without touching the daily sweep at all — see option (d) below.

## 4. Four shapes, not a recommendation — weigh them

**(a) Do nothing; document it as deliberate.** Cheapest. The honest framing for round 3's
council submission ("no cheap re-checkable predicate for a subjective judgement") already
argued this, and it got APPROVED on that basis. Cost: `vision_finding` remains a silent-sink
shape once a human forgets about it — the same shape `bugs_open/033`/`083` document for
`needs_human_review` generally, just narrower. If you choose this, say so explicitly
somewhere durable (this directory, or update TL-041 in the register) so it reads as a
decision, not an oversight.

**(b) Register a real revalidator in `reviewRevalidators`.** The consistent answer — every
other type gets closed the same way, one queue, no bespoke mechanism (this is literally the
position this lane argued to the council to win round 2/3). Predicate: re-run the
tool-acceptance `look` step (or a lighter-weight version of it — just the screenshot +
vision call, not the full mechanical suite) for the row's `function`/`site_id`, parse the
same `FINDINGS:` line, close on `none`. Cost: a real LLM+browser-runner call **per parked
vision_finding row, per daily sweep pass**, for however long the row stays open. Budget and
rate-limit implications need sizing — check current vision-call volume/cost
(`llm_call_log` filtered to the vision provider/model used by `tool-acceptance-agent`,
noting the three silent traps on that table — see the memory entry
`llm-call-log-agent-type-relabelled.md`) before committing to this shape.

**(c) A cheap staleness proxy instead of a real re-check.** Instead of re-running vision,
compare the page's last-modified/rerendered timestamp (or a content hash, if one exists —
check `pages`/`page_components`/`content_components` for what's actually tracked) against
the finding's `created_at`. If the page changed since the finding was filed, don't
auto-close (you don't know the change fixed it) — but you COULD downgrade it, annotate it
"page has changed since this was filed, please re-verify" for the human, or resurface it in
some review-priority ordering. Cheap, no new LLM cost, but it never actually closes
anything on its own — it only helps a human triage faster. Worth asking whether that is
enough to satisfy the spirit of "one queue, closable."

**(d) Piggyback on the EXISTING acceptance cadence instead of the daily review-queue sweep
at all.** `check_tool_acceptance_due` (`platform/orchestration/actions/discovery_checks/
check_tool_acceptance_due.go`) already re-verifies every active tool at least every 7 days
(cooldown default, line ~10) by queuing a fresh `acceptance_run` work item — which runs the
SAME vision `look` step that originally filed the finding. If `RecordVisionFindingAction`
were extended to look up and close an existing open `vision_finding` row for its own
`item_key` whenever the current verdict is `none` (the gap named in §3), a stale finding
would self-close within roughly a week of the underlying page being fixed, **for free**,
riding a cadence that already exists and already runs the exact check needed — no entry in
`reviewRevalidators`, no new scheduled task, no extra LLM spend beyond what already happens.
This is the option that looks cheapest and most elegant from the code alone. It has two
open costs, both unverified, both flagged in §6 — check them before treating this as decided.

## 5. A design tension worth naming, not resolving here

Option (d) closes the item from a DIFFERENT mechanism than the other six types (which all
close via the shared daily sweep). Round 2/3 of the council review for THIS SAME feature
argued hard, and won, against inventing bespoke side-channels instead of using the one
existing shared mechanism (`agenterrors.Write` over a bespoke doc_note — see the
`staged_component_build` NOTES entry cited above). Is closing `vision_finding` via its own
producer action, rather than via `reviewRevalidators`, the same mistake in a different
shape — or is it a genuinely different case, because the "shared mechanism" here
(re-verification cadence) already exists and belongs to tool-acceptance, not to the review
queue? Worth putting to the council either way once you have a direction, precisely because
the same lane just won an argument that could be read as pointing the other way.

## 6. Open questions to answer before you build anything

1. **Is `check_tool_acceptance_due`'s cadence actually live and reaching every site?** It's
   registered in the generic `discovery_checks` registry, and — unverified, flagged
   honestly rather than checked to save time in this pre-plan — the `scheduled_tasks` rows
   that plausibly drive it (`site-discovery-rotation-completeness`, `enabled=f` as of
   2026-08-11) look disabled, yet the SAME DAY's `bugs_open/243` candidate-1 proof
   explicitly relied on "the due-sweep" producing a real, unforced acceptance run
   (`ae33ed59…`). Reconcile which scheduled task actually drives `tool_acceptance_due`
   before assuming option (d)'s cadence exists in practice — do not repeat this round's
   own mistake (an asserted-absence claim that turned out false) in the opposite direction
   by assuming presence without checking.
2. **What is the real LLM/browser-runner cost of re-checking a parked vision_finding row
   daily** (option b) vs. weekly-at-most (option d)? Size it against current spend before
   picking a cadence.
3. **Does closing on a clean re-check risk false negatives?** A vision model's phrasing is
   not perfectly deterministic; is there a risk of a finding flapping open/closed across
   consecutive runs, and if so does that matter (a flapping item is still better than a
   permanently-stuck one, but worth naming).
4. **Is there a content-hash or last-modified column** on the page/component this finding's
   `spec.page_id` points at, cheap enough to support option (c) as a pre-filter before any
   of the more expensive options (e.g. skip a live re-check entirely if the page provably
   has not changed since the finding was filed)?
5. **Should this go to architecture review or just the normal council gate?** Per the
   2026-08-02 owner ruling (RFC_010 §1), converging a new closer onto an existing shared
   type is council-gate scope, not RFC scope, PROVIDED the producer set and mechanism are
   named in the concept register in the same commit (TL-041 already exists — extend it,
   don't fork a new entry).

## 7. State of the world when this was written

- `vision_finding` (TL-041) is **live, council-approved** (correlation
  `310dee45-ab34-4246-a69b-ab2df818a80f`, round 3 APPROVED 2026-08-11 ~18:22Z). Nothing
  about it is broken; this pre-plan is about a gap the review process surfaced, not a
  defect in what shipped.
- Zero `vision_finding` rows exist in production as of this writing — the negative arm
  (clean pages produce none) is proven three times independently; the positive arm has
  never fired for real. Whatever you build, you will not have a live row to test against
  until one is filed — either wait for a genuine finding or file one by hand for testing
  (the row shape is in TL-041 / `record_vision_finding_action.go`).
- No work has been claimed on this pre-plan. Check `who-owns.py` and recent lane
  transcripts before starting, per the fleet's usual multi-session-coordination practice —
  this document does not itself constitute a claim.

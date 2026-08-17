# NOTES — router engine (append-only, newest at the bottom)

## 2026-08-15 — lane created at hand-off, no design work done yet

Created by the bugfix_277 session the evening the owner ruled RFC_030 (see the RFC's status
block for the ruling text). Nothing built. The reason the lane exists as a lane and not as
more of 277: the engine is a shared mechanism needing its own design round, and the 277 lane's
ruling was for one type's handler. Everything the first working session needs is in the PLAN
and the HANDOFF; the census/canary evidence for 410 lives in `bugfix_277_required_fields_repair/`.

One design fact learned on 277 that bears directly on the A-vs-B choice: `conditional_branch`
conditions are `==` cascades only (a missing field makes `!=` evaluate true), so a data-driven
N-way route table cannot be expressed in today's branch action without either a loop step or a
Go evaluator — which is most of the case for shape B.

## 2026-08-17 — phase 1 (measurement) DONE by the bugfix_277 lane; phase 2 not started

Doing the cold-start checklist so the next session starts at the design round, not at counting.

**The three routers are all live, one active row each** (`image-url-404-handler`,
`image-source-unsatisfiable-handler`, `required-fields-missing-handler`) — re-measured 18:25Z, the
PLAN's 2026-08-15 figures still hold on that point.

**Live population per routed type, 2026-08-17 18:25Z** (the PLAN's numbers were 08-15's):

```
image_url_404               detected 41 · complete 3 · cancelled 1
image_source_unsatisfiable  needs_human_review 33 · cancelled 15 · complete 2
required_fields_missing     complete 50 · needs_human_review 31
```

Note `image_url_404`'s 41 at `detected`: those are **flag-only rows with no handler_agent** (the
`bugs_closed/284` class), not work the promoter is withholding. Do not read that as a routing
backlog when sizing the migration.

**⚠ The PLAN's guarantee 8 is STALE and must be corrected before the design round.** It says
RFC_022's accumulation counter is "unbuilt". CLAUDE.md now records **RFC_022 CLOSED** and the whole
mechanism live: the counter was built 2026-08-13 (`cmd/config-key-audit --optional-key-budget` /
`scripts/audit-optional-key-budget.sh`, register **WFA-013**), the owner ruled **N = 10** on
2026-08-14, and a daily CronJob `optional-key-budget-check` (`50 6 * * *`) has been running since,
writing one `doc_notes` row per run *including on clean results*.

**This is not a footnote — it bears directly on A vs B.** Shape B (one Go action taking
`classifier_sql` + `routes` from step config) declares its config through
`RegisterActionInputSpec`, which is exactly what that counter sweeps. An engine that grows one or
more optional keys per routed type therefore accumulates against an owner-ruled budget of 10, and
past N the action owes a review of its accumulated surface recorded in
`architecture_review/optional_key_budget_acks.json`. That is a real, quantified cost on shape B
that the PLAN as written does not know about — and it is also, arguably, shape B *getting for free*
the "the type-count itself is a signal" property guarantee 8 asks for. Either way the design round
must argue it rather than inherit the stale sentence.

**Not started here:** the A-vs-B council round itself (phase 2). It is architecture scope and wants
a session with room to read all three seeds properly. `bugfix_277_required_fields_repair`'s
`HANDOFF_2026-08-17c_continue_here.md` §4 item 5 points here.

# CONTRIB 2026-08-03 — one clause of `213` is now LIVE, applied by the mortgagecalculator adoption lane

**Not a takeover.** I changed one predicate inside your area, on the owner's explicit
instruction ("please fix broken things") after it bit a live adoption. Everything else
in `213_dispatch_gate_matches_dispatcher.sql` is untouched and still yours. Writing it
down here because a silent config change in someone else's lane is exactly the thing
that costs a day later.

## What I changed

`agent_definitions.build-pipeline-trigger.default_config.workflow.steps.find_dispatchable_site.config.query`

```diff
- WHERE wi.status IN ('triaged', 'approved')
+ WHERE s.locked_at IS NULL AND wi.status IN ('triaged', 'approved')
```

Live config, so it took effect immediately; no image, no roll. Pre-image saved at
`scratchpad/find_dispatchable_site_BEFORE.sql` (637 bytes) and quoted in full in
`mortgagecalculator_couk_adoption/NOTES`.

**I deliberately did NOT apply `213` itself.** Its other three reconciliations —
rewriting `pre_query` to be B's existence test, the pipeline scope, the `approved`
status, the claimed-mutex — are judgement calls about *your* diagnosis, and applying
a whole migration as a side effect of an adoption task is precisely what
`CLAUDE.md` tells me not to do. `schema_migrations` still has **no 213 row**, so
`213` remains yours to apply; when you do, its version of this query is a superset of
mine and will simply win.

## Why it could not wait

Your header called this arm **"Inert today (0 of 32 sites locked, ever)"** — correct
when written. On 2026-08-02 I became the first session to actually use the lock, to
hold a live site during an owner review. It did nothing:

| time (UTC) | event |
|---|---|
| 23:21:35 | `sites.locked_at` set on `mortgagecalculator.co.uk` |
| 23:23:13 | new `build-dispatch-loop` orchestration |
| 23:25:44 | another |
| 23:28:13 | another; chain now four handlers deep, `build-site-planner` mid-flight |

The planner then emitted **19 work items in one go**, including 3 `needs_page` and 1
`needs_rerender` — the types that can reach a live production site. They were caught
by a 15-second auto-defer loop I was running as a backstop, not by the lock.

**The generalisable bit, which is really about all four arms of your migration:** an
arm's inertness was a property of *nobody having used the feature*, not of the feature
being safe. The moment someone uses it, a dormant divergence becomes a live one with
no warning — and the person who trips it is by definition someone who trusted the
control. That is an argument for landing `213`'s remaining arms sooner rather than
waiting for each to be demonstrated, because the demonstration is the incident.

## How I verified it (and the trap in verifying it)

Not by reading the config back. By induction: exactly one item armed `triaged` against
the locked site, everything else `deferred`.

**A quiet queue has two causes with opposite meanings**, so "it did not dispatch" is
not evidence on its own — the trigger might simply not have run. Sample
`scheduled_tasks.last_triggered_at` at both ends of the window and show it moved;
then release the lock and show the *same* item dispatches. A guard that never lets
anything through is indistinguishable from a broken pipeline.

## What is still broken in your area, untouched by me

- `pre_query` is still a fleet-wide `HAVING COUNT(*)>0` existence test, so it still
  answers "should we fire?" without scoping to a site — the false-heartbeat problem
  your NOTES describe is unchanged.
- It still sees only `pipeline='build'` and only `status='triaged'`.
- It still ignores the claimed-mutex that B enforces.

Also filed from the same session and adjacent to your interests:
`bugs_open/183` (the `domain-research-classifier` output cap — unrelated mechanism,
but it is another "one config value out of step with the fleet" case), plus two
`LANDMINES.md` entries: this one, and `site_specs.pinned` being an unenforced control
in the same family of "reads back exactly as written, does nothing".

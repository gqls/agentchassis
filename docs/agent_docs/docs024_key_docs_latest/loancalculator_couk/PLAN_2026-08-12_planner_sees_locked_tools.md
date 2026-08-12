# PLAN 2026-08-12 — teaching the planner to see a site's own calculators

Owner decision, 2026-08-12: *"keep the 12 calculator locks for the first pass as you
suggest and fix the planner so it looks at them and doesn't delete them."*

That is TWO changes, and only the second one is small.

## Half 1 — "doesn't delete them" — DONE, committed `f4820a877`

`matchLockedRow` now pairs a locked row by component identity, not slot name alone
(council `a625c326`). Full reasoning in the commit and the bug file. **Inert until
half 2 lands**, because nothing can name a tool component today.

## Half 2 — "looks at them" — DESIGNED, NOT APPLIED

The planner's component menu is one shared step:

```
build-site-planner → workflow.steps.load_components (action: query_database)
  SELECT name, display_name, "function", category, description
  FROM content_components
  WHERE component_level IN ('section', 'element') AND is_active = true
  ORDER BY category, name
```

### Three traps, all measured 2026-08-12 — this is why it is not a one-line edit

1. **`content_components` has NO `site_id` column.** The library is GLOBAL. Adding
   `'tool'` to that `IN` list does not show a planner "this site's calculators" — it
   shows it **all 81 active tool components on the platform**, including other sites'.
   A gaswholesalers plan could place loancalculator's settlement calculator. The
   widening MUST be scoped to tools already placed on the planning site's own pages.
2. **21 sites already place tool components** (78 placements: finetuning.uk 9,
   ai-agent-orchestration 7, gamesdesign 7, robot-hands 6, gaswholesalers 6, …).
   So even a site-scoped widening is NOT inert — it changes the planner's vocabulary
   for 21 sites the next time each replans. That is a fleet behaviour change, and
   under the 2026-08-02 §2 ruling it wants an opt-in flag with the unsafe default OFF.
3. **A `params` path that resolves to nil is a HARD STEP FAILURE.**
   `QueryDatabaseAction` returns `fmt.Errorf("query param path '%s' resolved to nil")`
   (`database_actions.go`). Binding `$1` to a site id means picking a path that
   resolves on **every** planner run on **every** site — because if it ever does not,
   `load_components` errors and NO site can plan. That is an outage-class failure mode
   on a single shared row, and it is the reason this half must not be rushed.

### The shape to build (self-gating, so it is provably inert where it should be)

```sql
SELECT name, display_name, "function", category, description
FROM content_components
WHERE is_active = true
  AND ( component_level IN ('section','element')
        OR ( component_level = 'tool'
             AND EXISTS (SELECT 1 FROM site_specs ss
                         WHERE ss.site_id = $1 AND ss.aspect = 'structure'
                           AND ss.is_current
                           AND ss.data->>'plan_includes_tools' = 'true')
             AND id IN (SELECT pc.component_id FROM page_components pc
                        JOIN pages p ON pc.page_id = p.id
                        WHERE p.site_id = $1) ) )
ORDER BY category, name
```

Two gates in series: the site must opt in **and** the tool must already be on one of
its own pages. For every site without the flag the result set is byte-identical to
today's, which is a property a reviewer can check rather than take on trust.

**Before applying, three things must be settled:**
- **The param path.** Read a real recent `build-site-planner` orchestration's
  `collected_data` and find a site-id key present in *every* one. Do not guess.
  Verify against several runs, including a fresh-build run and a replan.
- **A control run.** Confirm the new query returns an identical row set to the old
  one for a site without the flag, executed against the live DB, before it goes in
  the agent definition.
- **Fifth flag in `structure`.** This would be `url_shape`, the three PLAN-048
  identity gates, and now `plan_includes_tools`. That is exactly the accumulation
  `RFC_022` says nothing currently counts. Say so in the submission rather than
  letting it pass unremarked.

### Sequencing, and why it matters

Half 2 without half 1 is **worse than doing nothing**: the planner names the tool, the
composition inserts a fresh copy in place, and the locked original is exiled to the
page foot — two calculators per page. Half 1 is committed but **not rolled**. So the
order is: roll half 1 → build half 2 → council → roll → verify on one site → then a
rebuild that keeps its calculators in place.

## What this means for the rebuild, today

With half 1 unrolled and half 2 unwritten, a rebuild fired now behaves exactly as the
decomposition lane's test predicts: the 12 locked calculators **survive** (locks hold;
they are never deleted) but each is **repositioned to the foot of its page**, with a
`lock_blocked_change` item filed per row saying so. Recoverable — position is one
UPDATE per row and the pre-state is recorded three ways — but it is 12 visibly wrong
pages on a live indexed site in the meantime.

# Handoff — a page flagged `needs_rebuild` still loses its composition to a re-plan

**Filed 2026-07-20**, measured while verifying the `/bugs_open/001` fix live. This is a **boundary
decision left open by that fix**, not a regression it introduced — the behaviour is the same as
before 001; 001 simply protected `deployed` pages and left this case exactly where it was.

**It may well be the wanted behaviour.** It is filed so the choice is made deliberately rather than
inherited from an implementation detail.

## What happens

`/bugs_open/001`'s guard preserves a realised page's composition when it is adoption-locked **or
built**, where "built" is `build_status = 'deployed'`:

```go
// v3_site_actions.go
func realisedPageIsBuilt(rm map[string]interface{}) bool {
	status, _ := rm["build_status"].(string)
	return status == "deployed"
}
```

`needs_rebuild` is therefore outside the preserved set, so a re-plan takes the LLM's proposed
composition for such a page — even though the page may have a full, previously-deployed composition
sitting in `pages.sections`.

## Evidence (dartsonline.com, 2026-07-20, plan `fba367c9` → `5d438145`)

`index` was `needs_rebuild` with seven sections. The re-plan replaced them:

| | sections |
|---|---|
| before (realised) | `hero, category-listing, product-grid, differentiators, call-to-action, testimonials, content-listing` |
| LLM proposed | `hero, product-grid, category-listing, features, call-to-action, testimonials` |
| plan took | the LLM's — **unchanged from the proposal** |

Net: **lost `differentiators` and `content-listing`, gained `features`**, and two slots reordered.
These are distinct components, not naming variants — checked against `content_components`, unlike
the `about` case on the same run (see `/bugs_open/039` Part 1 for that trap):

```sql
SELECT name, function FROM content_components
WHERE function IN ('features','content-listing','differentiators');
--  features                | features
--  content-listing         | content-listing
--  differentiators-section | differentiators
```

In the same run, `contact` was also `needs_rebuild` and came through unchanged — but only because
the LLM happened to re-propose its exact composition, **not** because anything protected it. So the
exposure is real and its effect is a coin-flip on what the LLM writes that run.

## The argument each way

**For leaving it as-is.** `/bugs_open/001`'s own fix step 4 proposed precisely this as the escape
hatch: *"Consider gating a deliberate rebuild behind explicit intent (a per-page `rebuild:true` in
the `needs_site_plan` spec, or **a page whose `build_status` was set to `needs_rebuild`**), so a
genuine redesign is still possible — just never the silent default."* On that reading `needs_rebuild`
IS the explicit intent, and the current behaviour is the design working. It also keeps a way to
redesign a page at all, which a blanket guard would remove.

**Against.** `needs_rebuild` is set by machinery, not only by a human asking for a redesign —
`v3_site_actions.go:644` sets `build_status='needs_rebuild', built_from_plan_version=NULL` as part of
ordinary status handling. So a page can acquire the flag for reasons that have nothing to do with
wanting a new design, and then silently lose its composition at the next re-plan. "Rebuild this
page" and "recompose this page from scratch" are different intents sharing one flag. Fleet-wide this
is not a corner case:

```sql
SELECT rebuild_policy, build_status, count(*) FROM pages GROUP BY 1,2;
--  generic | needs_rebuild | 27
--  owned   | needs_rebuild |  7
```

34 pages currently sit in this state.

## Fix candidates (if the decision is to change it)

1. **Separate the two intents.** Keep `needs_rebuild` meaning "re-render this page as planned"
   (composition preserved, like `deployed`), and introduce an explicit `needs_replan` /
   per-page `rebuild: true` in the `needs_site_plan` spec for "recompose it". This is fix step 4
   done properly and makes the silent case impossible.
2. **Preserve unless the page's composition is empty.** Widen `realisedPageIsBuilt` to
   `deployed OR needs_rebuild`, relying on Pass B2's existing non-empty gate to let a genuinely
   uncomposed page still be composed. One-line change, but it removes the only current route to a
   deliberate redesign — do not take it without (1) or some other way back in.
3. **Surface rather than decide.** When a re-plan would change a `needs_rebuild` page's composition,
   emit the diff as a review item instead of applying it silently. Correct in spirit, but
   `/bugs_open/033` says that queue currently has no working surface, so it would rot — fix that
   first or this is a no-op.

Recommend **(1)**, and until it exists, treat `needs_rebuild` as "this page's composition is
unprotected" when planning any re-plan.

## How to verify a fix

1. Set a deployed page to `needs_rebuild` without asking for a redesign, re-plan, and assert its
   `pages.sections` is unchanged.
2. Assert a page with genuine redesign intent (whatever (1) settles on) IS still re-composed — do
   not fix this by making redesign impossible.
3. Check the artefact: the rebuilt page's `page_components` should match the preserved list, matching
   on `function` per `/bugs_open/039` Part 1.

## Related

- `/bugs_open/001` — the fix this sits at the edge of; its "VERIFIED LIVE" section records this as
  limit 1 and points here.
- `/bugs_open/038` — the other half: even a *protected* page is still rebuilt and its content
  regenerated.
- `/bugs_open/039` — why the `differentiators` / `differentiators-section` distinction above needed
  checking before this could be called a real loss.
- `/bugs_open/033` — why fix candidate 3 would currently rot.

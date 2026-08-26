# NOTES — bugs_open/407, a site cannot put its own page in its own header

Append-only, newest at the bottom. Technical log: evidence, commands, what the system
actually said, and every misstep.

---

## 2026-08-26 — lane opened; ownership checked, bug re-validated

### Why this bug

Picked after `bugs_open/359` closed out for the session. The sweep for the next genuinely
open, genuinely unowned, framework-shaped bug ruled out the nearer candidates on evidence
rather than on feel, and the reasons are recorded so the next sweep does not re-walk them:

| candidate | why not |
|---|---|
| `348` | superseded by its own banner — *"Do not fix from this file, `bugs_open/344` OWNS THIS"*, and 344 is closed |
| `361` | `bugfix_260_render_fallback`, ACTIVE 24 commits/14d |
| `376` | `loanzy_uk_example_site`, ACTIVE 119 commits/14d, 44 mentions, and a handoff literally named *"the route is gated on 376"* |
| `386` | `bugfix_386_counting_fact_drift`, ACTIVE, fix already live |

`407`: `scripts/who-owns.py 407` → **`likely OWNING workstream(s): (none identified)`**, with
two cross-reference mentions from the filing lane only. Filed today at the **owner's explicit
direction**, with his preferred direction quoted in the file.

### The bug is STILL VALID — re-measured with its own §6 query

`[MEASURED 2026-08-26]` **6 pages across 4 sites** declare `in_header`, are `active` +
`deployed`, are not child pages, and are absent from their site's primary nav:

```
ai-agent-orchestration.com  adoption-tracker      /adoption-tracker.html       nav_order 100
ai-agent-orchestration.com  news-index            /news/index.html             nav_order 100
ai-agent-orchestration.com  protocol-tracker      /protocol-tracker.html       nav_order 100
finetuning.uk               approach              /approach.html               nav_order 4
gaswholesalers.com          pricing-transparency  /pricing-transparency.html   nav_order 100
idea.uk                     report                /report.html                 nav_order 3
```

The filing recorded **8 across 5 sites**. **`loanzy.uk`'s two are gone and `idea.uk/report` is
new.** That movement is itself evidence for the filer's own §4 ⚠: a SECOND cause — a merely
stale nav — is mixed into this count, and loanzy.uk was their named example of it. So the
number is a true upper bound on this defect and not a measurement of it, exactly as filed.

**Building the discriminator is part of this job**, not a nicety: without it a post-fix count
cannot be compared with a pre-fix one, and the fix would be verified against a number that
moves on its own. The filer's attempt —
`pages.updated_at > max(site_nav_items.updated_at)` — does not work, because
`pages.updated_at` is bumped by any re-render and so does not tell you the FLAGS changed.

### §3 is the argument, and it is measurement rather than opinion

The existing workaround lives inside the broken thing. Three site-specific page names
(`model-directory`, `adoption-tracker`, `protocol-tracker`) were hardcoded into the
fleet-wide tier-2 map (`populate_nav_tables_action.go:461-471`) with a comment stating that
as tier 3 they would be *"the first thing dropped when max_header_items truncates"*.

`[MEASURED 2026-08-26]` **two of those three names are in the absent list above**, on the very
site the comment names. The documented remedy is in the source today and the pages it was
written for are still missing. That is what refuses candidate 3 (extend the list again) —
not a preference for structure, an observation that the last extension did not hold.

### Architecture read (first-hand, today)

- `classifyPagesForNav` (`:297`) sorts **tier ASC, then `nav_order` ASC**, then the caller
  truncates to `maxHeaderItems` (`:131`) and sends the overflow to `utility` so it survives in
  the footer. So `nav_order` genuinely cannot cross a tier boundary.
- `navPriorityTier` (`:442`) is two hardcoded maps plus `isSectionIndexType`.
- `max_header_items` is `PopulateNavTablesInputSpec.Optional`, default 8, read from the STEP
  config — **fleet-wide**.
- The action **`DELETE`s every `site_nav_items` and `site_nav_groups` row for the site**
  (`:160`, `:163`) and re-derives from `pages`. A per-site declaration therefore cannot live
  in `site_nav_items`: it is a derived table.
- `nav_prune_floor.go` guards that delete with two cohorts, and its header records that
  **seven other actions write `pages.in_header/in_footer`** while this one only reads them —
  which is what makes its "pages seen" cohort an independent second opinion.
- `classifyPagesForNav` also pins the `bugs_open/149` A2 invariant, enforced by
  `nav_membership_test.go`: *`pages.in_header`/`in_footer` declares nav MEMBERSHIP; a page's
  URL shape may decide WHERE it appears, never WHETHER it appears.* Any fix must keep it.
- `sites.settings` is the established per-site config home. `[MEASURED 2026-08-26]` keys in
  use fleet-wide: `maintenance_profile` 31, `pool` 17, `skip_build` 1, `skip_deploy` 1 — and
  `maintenance_profile` already carries Go-read sub-objects (`content_feed.enabled`,
  `content_feed.max_age_hours`, `structure_floor.n`/`.refusal`). There is **no** `site_specs`
  table.

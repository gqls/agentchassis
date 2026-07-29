# NOTES — dartsonline.com traffic & affiliate-readiness

Append-only, newest at the bottom. Technical log: evidence, commands, what the system
actually said, and every misstep.

---

## 2026-07-29 — session 1 (workstream opened)

### Exploration corrections (things the brief and my own first pass got wrong)

Six claims I formed early and then had to correct. Recorded because each was locally
plausible and none survived a query.

1. **"The nav 404s mean three landing pages need building."** WRONG. `shop`, `brands`,
   `guides` are orphan `pages` rows from SUPERSEDED plans. The current plan
   (`site_plans.is_current`) contains `shop-index`, `brands-index`, `guides-index`
   (role `section-index`) and no landing rows at all. The hubs replaced them; the old
   rows kept `in_header=true` and `in_footer=true`. What caught it: querying
   `site_plan_pages` for the current plan instead of trusting the `pages` table alone.

2. **"`detected` work items can never dispatch, so `check_undeployed_assets` needs a
   2-line fix."** WRONG, and this one nearly became a code change. `triage_detected_items`
   (`platform/orchestration/actions/triage_detect_items_action.go`) promotes EVERY
   detected item to `triaged`/`build`; three live agents call it (site-review-agent,
   design-audit-agent, improvement-loop). The items are stranded because
   `scheduled_tasks.improvement-sweep` is `enabled=false` (last triggered **2026-05-02**),
   not because the check writes the wrong literal. And `bugs_open/083`'s owner has
   already gone further: *"routing is NOT the bottleneck … there is no reader for ANY of
   the queues"* — 325 items sit in `needs_human_review`, the queue humans CAN see, oldest
   2026-03-15. That file ends **"Decision pending — do not act on this section until it
   is recorded here."** So the planned generic fix is cancelled: it would have been a
   routing fix that its own bug file argues against, on someone else's open decision.
   Site-scoped SQL promotion for dartsonline is data, not mechanism, and stays in scope.

3. **"The nav-404 class is a generic framework defect."** WRONG — the framework already
   fixes it. `GetNavItems` (`nav_tables.go:215-240`) drops nav items whose target has
   never been deployed. The fleet measurement is decisive:
   ```sql
   SELECT s.domain, count(*) FROM pages p JOIN sites s ON s.id=p.site_id
   WHERE p.in_header AND p.deployed_at IS NULL AND p.status IN ('active','deployed','pending')
   GROUP BY 1;
   -- dartsonline.com | 4      <- one site, nothing else fleet-wide
   ```
   dartsonline serves STALE STORED CHROME (`bugs_open/117`). Data fix + chrome rebuild,
   no platform change. **The lesson is the order:** I measured the fleet before writing
   the fix, and the measurement removed the fix.

4. **"The fabricated identity is on the About page."** Partly wrong, and the important
   half was invisible from the page. The live page has NO Portland, NO brand names, NO
   darts.com contact (all greps 0). It DOES claim "We stock across the full range" ×2 and
   "We carry the manufacturers…" ×1. The Portland address, `sales@darts.com`,
   `(800) 526-1920`, the seven brands and the AUSTRALIAN company's Facebook URL
   (`facebook.com/dartsonlineau`) all live in the `identity` and `briefing` SPECS — i.e.
   in the source every future rebuild would draw from. Fixing only what was visible would
   have let the next rebuild reintroduce all of it.

5. **"S3 credentials block the imagery work."** WRONG for this site — that blocker is
   about writing hand-made exact bytes for three OTHER sites. All 33 dartsonline assets
   already serve HTTP 200.

6. **"The tool-imagery hold gates the setup-builder."** STALE — `bugs_closed/020` closed
   2026-07-23 and the owner lifted the hold 2026-07-24.

### The `gaps`-array trap (my own detector poisoned by my own text)

After the briefing reset my verification query still reported `portland: t` and
`stock_words: t` on the NEW row. Two different causes, and only one was a problem:
- `honesty_rails` (a key I had just added) contains *"Never claim to stock, carry…"* —
  the prohibition matches the pattern for the thing it prohibits. This is exactly
  [[prompt-text-poisons-its-own-detector]] and it bit inside five minutes of my writing it.
- `gaps` still read *"number sourced from associated Portland operation"*. That WAS a
  real leak risk: the phone had been replaced, so the sentence was untrue-of-now, and any
  writer reading the whole briefing blob into a prompt could re-contaminate copy from it.
  Rewrote `gaps` to current reality; provenance kept in the row's `notes` column and the
  backup table, neither of which is fed to a prompt.

**Transferable:** when a fix ADDS prohibition text, a substring check over the whole blob
can no longer distinguish "claim removed" from "claim named in order to forbid it". Check
per-key (`jsonb_each`), not over `data::text`.

### The gap that actually shipped the fabrication

`briefing.gaps` recorded, on 2026-07-06, that the phone was *"sourced from associated
Portland operation; not confirmed"* and that there were no *"confirmed brand partnership
or stock relationship details beyond signal-level inference"*. The research was honest
about its uncertainty **and the build rendered the claims as fact anyway.** The defect is
not bad research — it is that nothing between "recorded as unverified" and "rendered as
prose" reads the uncertainty. Worth a bug file of its own if a second site shows it.

### Applied this session (all live, all reversible via the bak_ tables)

| what | file | result |
|---|---|---|
| identity truth reset | `SQL_2026-07-29_identity_truth_reset.sql` | 1 current row; portland/stocking/AU-facebook all false; old row preserved |
| briefing truth reset | `SQL_2026-07-29b_briefing_truth_reset.sql` | about_us + services rewritten; `headquarters`/`location` keys dropped; `honesty_rails` added |
| stale `gaps` rewrite | inline (recorded above) | 0 keys mention Portland |
| nav reconciliation | `SQL_2026-07-29c_nav_reconciliation.sql` | 3 orphans archived; setup-builder out of nav; 3 hubs in with clean labels; 13 polluted nav_labels trimmed |

Backups: `bak_darts_identity_20260729`, `bak_darts_briefing_20260729`,
`bak_darts_pages_nav_20260729`.

### Verified facts worth not re-deriving

- Work items open: 17 `undeployed_asset` (detected/design), 8 `needs_page` (NHR),
  4 `needs_rerender`, 3 `page_rerender`, 3 `deactivated_component`, 3 `unresolved_cta`,
  1 each `capability_gap` / `empty_section` / `evaluate_tools` / `needs_section_data` /
  `owned_page_review`. (Exploration said 5 `unresolved_cta`; the live count is 3.)
- `bugs_closed/141` is CLOSED and proven live on v1.0.1198 (`9db57e426`) — news-index
  can enter nav. No pre-check needed before creating the news page.
- `check_missing_tools` (`discovery_checks/check_missing_tools.go`) already exists and is
  the natural home for the owner's 1-tool-per-6-articles ratio (D5). Today it is purely
  TIME-based: 7-day cooldown at 0 tools, 30-day at 1+. It counts deployed tools via
  `content_components.component_level='tool'`. It emits `evaluate_tools` →
  `tool-suggester` at `Status:"detected"` — same stranding as everything else.
- `loadSiteContactEmail` (`validate_page_content.go:1274-1298`) reads FLAT
  `identity.data->>'email'` / `->>'contact_email'`; `sync_site_identity_action.go:103-110`
  writes NESTED `identity.contact.email`. Both shapes now written (bug-072 class).
- `populate_nav_tables` DELETEs and rebuilds nav from `pages` — hand-editing
  `site_nav_items` is pointless, `pages.in_header` is the control surface.

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

---

## 2026-07-29 — session 1 continued: the third source of the same lie

### The nine guides were unblocked by a data fix, and it worked first time

`site_plan_sections` had zero rows for all nine blog-posts, and `pages.sections` was
`'[]'`. Inserted the canonical article layout — `["hero","article-body","call-to-action"]`,
taken from `create_blog_posts_action.go:183` rather than invented — into the PLAN table
(the authority; `pages.sections` is documented in `load_page_sections_from_spec_action.go`
as the materialised cache, and the loader syncs down to it itself).

Promoted two items only. **barrel-weight became the first guide page ever built on this
site** (13:32Z, 3 components, 9,039 bytes), then beginners. Both `complete`.

**Misstep worth recording:** while the first build was running I read the work item's
`error` column and saw *"page-build-handler no-op: no sections ready to build"* — the
exact failure the backfill was meant to cure — and briefly concluded the fix hadn't
worked. It had. `error` holds the text of the LAST failure and is not cleared when a row
is re-claimed; `beginners` still showed a `claimed_at` of 2026-07-20 next to it. The
authority on "is it working" was `orchestration_states`, which showed the run stepping
through `deploy_page` to `complete`. **A stale error column reads exactly like a live one.**

### No blog-index page was created — a deliberate departure from the plan

The plan said to hand-create one. `guides-index` (`/guides/index.html`) is already
deployed and already carries `content-listing`, whose `input_schema` sources
`query.blog_posts` — the same resolver a blog index would use. It IS the guides index;
it lists nothing only because no guide had ever been built. A second listing page would
have duplicated it and split internal links.

### THE FINDING: fixing the identity specs was necessary and NOT sufficient

barrel-weight came back well-written, on-voice, spec-accurate — and its call-to-action
read **"Filter our ranges by weight and tungsten percentage."** There are no ranges.

This was not the writer hallucinating. `content_direction` — untouched by the morning's
truth reset — instructs it to write shop copy:

| key | text |
|---|---|
| `writing_rules[0]` | "…on **product listings**" |
| `writing_rules[4]` | "Keep CTAs action-first: **'Add to Bag'**, 'Pick Your Weight'" |
| `writing_rules[7]` | "**Price copy** should be direct and confident: show savings…" |
| `writing_rules[8]` | "**Brand pages** should… not just list SKUs" |
| `persuasion_approach.method` | "Position the **store** as a trusted guide" |
| `content_depth.thoroughness` | "**Product pages** go deep on specs…" |

So the false premise had **three** homes — `identity` (who we are), `briefing` (what the
About page says), and `content_direction` (how every page is written) — and the third is
the one the writer reads most directly. Fixing the two obvious ones and building eight
more guides would have produced eight more invitations to browse a catalogue that does
not exist.

**How it was caught: by building one page and READING it.** No spec inspection would have
found it, because each aspect is internally coherent — it is only wrong relative to a
business that no longer exists. This is the case for building ONE and looking, rather than
promoting all nine and discovering it nine pages later. Two, then look, was the right size.

Fixed in `SQL_2026-07-29f_content_direction_editorial.sql`: replaced the four
shop-assuming keys, kept the voice keys untouched (the voice was never the problem), added
`editorial` (D2) and `honesty_rails` (D4). Both guides queued for rebuild.

**`formatted` is the load-bearing field and must be regenerated by hand here.**
`page-content-writer`'s prompt reads exactly one thing:
`{{.site_specs.specs.content_direction.formatted}}`. It is normally produced by the
`write_site_spec` action (`site_spec_actions.go:206-216` → `FormatContentDirection`). A
hand-authored spec that forgets it is **invisible to the writer** — the edit would look
applied and change nothing. The SQL reproduces the formatter exactly (string → `Key: v`;
array → `Key:\n- item`; object → `Key:\n` + per-subkey, joined `\n`; blocks joined
`\n\n`; `HumaniseKey` = underscores→spaces + capitalise first char). Go map order is
random, so block order carries no meaning.

`feed-triage` needed nothing: its prompt iterates every top-level `content_direction`
key, so `editorial` reaches news scoring the moment it exists.

### News feed armed

`classification` now carries `content_features.news_feed` (recommended, separate_page,
source_types rss+news_search+api_news, darts-specific keywords). Gate query verified
passing for this site. Also corrected `reasoning` and `detected_signals`, which were the
research provenance built on the Australian/Portland conflation — not rendered copy, but
what a future planner would read to re-derive the site's purpose, so leaving them would
re-seed the same error. `category`/`site_type`/`recommended_builder` untouched (they
drive builder selection).

Keywords are deliberately darts-specific: if `industry_tags` were ever fed to
`matchVerticalNews`, "sports-equipment" would token-match the generic `sports` entry
whose keywords are "sports news / match results / tournament / league standings" — true
of darts and useless for it.

### growth_config

Inserted (no prior row): `weekly_blog_posts_max` 5 (D3), plus `content_tools_ratio: 6`
(D5) carrying an explicit note that **nothing reads it yet** — the intended reader is a
change to `discovery_checks/check_missing_tools.go`, which today decides tool need on a
7-day/30-day timer with no reference to how much content a site has. Recorded as config
so the policy and its future reader sit together, marked UNREAD so it is not mistaken
for live behaviour.

### Truth reset verified on the artefacts, not the status

`about` and `shipping-returns` rebuilt and checked by counting phrases in the stored
`rendered_html`, not by trusting `status='complete'`:

| page | before | after |
|---|---|---|
| about | "we stock" x2, "we carry" x1, title "Specialist Darts **Retailer**" | all 0; now *"We don't sell darts, hold stock or ship products… an independent guide"* |
| shipping-returns | dispatch / tracking / courier / working days / checkout / cut-off / 30 days | all 0; now *"This site holds no stock and ships no orders… you buy your gear from your preferred retailer"*, and an FAQ that answers "Do you ship darts directly to me?" with "We hold no stock and ship nothing" |

The FAQ is the part I did not expect to come out well: asked "Who do I contact if I have
a problem with an order?" it answers *"the specific store where you completed your
purchase. We can't track shipments or process refunds"*. Setting
`pages.page_spec->>'purpose'` with an explicit FORBIDDEN list was what did it — the same
writer, one hour earlier, wrote the courier promises.

**Follow-up (small, not worth its own build cycle):** shipping-returns contains
"analyzing" — US spelling, against the platform's British-English convention. Sweep it
with the next rebuild of that page rather than churning a build for one word.

### Council + queue notes

- Tool-ratio change submitted to the council gate, correlation
  **`f5fc3014-973c-49a2-8d42-4bf9b401eaeb`**. Commit `f8190a7de` carries NO trailer yet;
  add `Council-Reviewed:` on a later commit only if the verdict is APPROVED.
- **The build dispatcher serves ONE SITE PER TICK, fleet-wide.**
  `build-pipeline-trigger.find_dispatchable_site` is
  `SELECT DISTINCT ON (wi.site_id) … ORDER BY wi.site_id, wi.priority ASC LIMIT 1`, so
  the winner is chosen by **site_id UUID order**, not by priority or age, and a site with
  any `claimed` item is excluded. I briefly diagnosed this as starvation when four
  dartsonline items sat `triaged` while `gaswholesalers` (UUID `5fe15466…`, one sort
  position ahead of `5fe8785b…`) held the slot with an `amend_asset:logo_failtest` row.
  **It was contention, not starvation** — that item completed, dartsonline became the
  selected site on the next tick, and the four items drained in sequence. Worth knowing
  for pacing: queue several items for one site and they run one at a time, a few minutes
  apart, not in parallel. Not filed as a defect; the throughput ceiling is real but no
  site was actually starved.
- Timing trap: I twice concluded "this has been stuck for minutes" from my own sense of
  elapsed time. `SELECT now()` said 56 seconds. Read the DB clock before calling
  something stalled.

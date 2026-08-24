# HANDOFF 2026-08-24 — continue here

**Lane:** `apis_uk_bees_homepage`. Start here, then `NOTES_apis_uk_bees_homepage.md` (newest at
the bottom) for the evidence behind any line below.

> ## ▶ ONE-LINE STATE
> The bees page is **finished and live** — illustrated, footer-free, locked. What is open is all
> **mechanism** work the owner has approved: a fleet GTM roll-out, a per-site tag field, the
> section-subject parser change, and two-round image accuracy. **Nothing is mid-flight; nothing
> is waiting on the owner except where marked DECISION.**

## 1. The owner's immediate blocker — the jq command was run on the wrong box

He ran it on `webdesignbox1`. **The probe log is on the ISLAND box.**

```bash
ssh root@toolsapisuk.vs.mythic-beasts.com          # Mythic Beasts vds:toolsapisuk, key-only
cd /opt/island/logs/probe                           # compose bind-mounts ./logs/probe -> /var/log/caddy
ls -la                                              # expect probe_access.log + rolled .gz
# current + rolled, ranked hostname x path:
{ cat probe_access.log 2>/dev/null; zcat -f probe_access*.log.gz 2>/dev/null; } \
  | jq -r '.request | .host + " " + .uri' | sort | uniq -c | sort -rn | head -30
```

⚠ **Two things to say when reporting the result.** The **apex arm has been dark** since the
`apis.uk/*` worker route appeared, so **missing apex rows mean "not observable", not "nobody
asked"** — the wildcard arm is the only live data. And Caddy rolls at 10MiB × 10 keep × 720h, so
the un-rolled file alone is a partial window.

## 2. Live state of apis.uk — verified, not assumed

| | |
|---|---|
| page | ~67.5KB, `<h1>A closer look at bees</h1>` |
| sections | 6, all `data-component="illustrated-text-block"` (**CLC-030**) |
| images | 7 (6 section + logo), every URL 200, solitary bees anatomically corrected |
| footer / email / company name | **removed** — `site_components.footer` emptied |
| protection | **all 7 `page_components` `lock_type='permanent'`** |
| `tools.apis.uk` | 200 throughout; **DNS never touched** |
| open review items | 2 — `brief_supplies_negation` (owner's call), `save_refused_incomplete` (historical) |

**Do not "tidy" any of the following**, each is deliberate and load-bearing:
- the empty `site_components.footer` row — emptying it is what removes the footer
  (`rerender_single_page_action.go:710` writes it only when non-empty);
- the permanent locks — they are what stops a sweep destroying the illustrated sections;
- `sites.email` / `company_name` / `content_data->'email'` all NULL, **and** an any-email ban in
  `evidence_base` — the ban went in FIRST because `bugs_open/063`'s check fails open without an
  address. ⚠ `multipage_actions.go:417` and `section_editor_actions.go` **synthesise
  `info@<domain>`** when a site has none; the rerender path does not. A multipage build here
  would invent an address.

## 3. GTM — DONE in the database, NOT yet on served pages

`GTM-PQ3WCTBD` inserted after `<meta charset>` in **all 27** `site_components.head` rows
(14 already had it; 13 backfilled), **exactly once each, 0 duplicated** — asserted by counting
occurrences inside the transaction.

**Reaching visitors needs a re-render: 695 deployed pages across 28 sites.**

**Use the site-level fan-out, not 695 direct dispatches:**
```bash
# per site — creates one page_rerender work item per page, which the dispatch loop picks up
{"action":"orchestrate","config":{"agent_type":"rerender-pages"},
 "input_data":{"site_id":"<id>","domain":"<domain>","refresh_site_components":false}}
```
⚠ **`refresh_site_components` MUST stay `false`.** True re-renders the chrome from
`RenderFallbackHead`, which has no GTM — **it would erase the very thing you are rolling out.**

⚠ **`rerender-pages` does not render.** Its steps are `get_pages_for_rerender` →
`create_rerender_items`; `page-rerender` does the work per page, asynchronously. A COMPLETED
`rerender-pages` therefore proves **queuing**, not deployment. Verify at the served page.

**Owner's constraint — do not re-render anything mid-change.** The fan-out is itself the safest
route (the framework's own claim machinery serialises), but gate the per-site trigger on:
```sql
-- skip a domain with a non-terminal orchestration, or components touched in the last 10 minutes
SELECT s.domain FROM sites s WHERE s.status IN ('active','deployed')
  AND NOT EXISTS (SELECT 1 FROM orchestration_states o
                  WHERE o.collected_data->'input_data'->>'domain' = s.domain
                    AND o.status NOT IN ('COMPLETED','FAILED'))
  AND NOT EXISTS (SELECT 1 FROM page_components pc JOIN pages p ON p.id=pc.page_id
                  WHERE p.site_id = s.id AND pc.updated_at > now() - interval '10 minutes');
```
⚠ **Do NOT gate on `build_status <> 'deployed'`** — 25 of 28 sites have unsettled pages at any
moment because `needs_rebuild` is a standing queue, so that filter skips almost everything.

## 4. DECISION + BUILD — per-site tag, because third-party sites need their own or none

Owner: *"When we build third party sites, they will need different tags or none at all. Go change
is fine if necessary."*

**Where it has to go.** Every site's head comes from **`RenderFallbackHead`** (a Go function),
because `ChromeSlotFunction("head")` asks the library for `function='head'` and **both** such
components are inactive — `Document Head` is deliberately ineligible (`component_level='section'`;
using it would render a page section as `<head>`).

**Design to build (council-scope, needs a roll):**
1. Read a per-site id — `sites.settings->>'analytics_container_id'` is the natural home (jsonb,
   already present, no migration).
2. In `RenderFallbackHead`, emit the GTM snippet **only when that value is non-empty**, straight
   after `<meta charset>`.
3. **Default empty ⇒ no tag.** A third-party site gets nothing unless someone sets it; ours get
   `GTM-PQ3WCTBD`. **Never hardcode our container** — that is the whole point of the request.
4. Backfill `settings` for the sites we own, so the Go path and today's hand-inserted heads agree.

**Falsifier before claiming it works:** a site with the key unset must render a head with **zero**
`googletagmanager` occurrences, and one with it set must render exactly **one**.

## 5. BUILD — per-section subjects (owner approved)

**The defect, measured four times:** `pages.sections` is a flat array of component names and is
parsed as `[]string` (`PlannedSections`), so every slot gets an identical brief and the writer
produces variations on one subject. One `content_rewrite` for the waggle dance rewrote **all six
sections about the waggle dance**. That is why apis.uk has four solitary-bee sections and why
`illustration_waggle_dance`, `_swarm` and `_pollination` are generated, live, and **unused**.

**Shape:** let an entry be either a string (today) or `{"component": "...", "subject": "..."}`,
and thread `subject` to the per-section writer prompt. Backward-compatible by construction.

**Scope:** touches the parser every page render passes through ⇒ **architecture-scope, council
gate, ships with a chassis roll.** Two `content_rewrite` items (swarm, pollination) are
**`deferred` with this as their written unblock condition** — un-defer them once it lands.

## 6. BUILD — image accuracy (owner approved: D done, then A + C)

**D is DONE** on apis.uk. ⚠ **`imageryStyleGuide` is a TYPED STRUCT** — `palette`, `medium`,
`mood`, `avoid`, `provider`, `reference_asset_keys`, `kinds`. **An added key such as
`subject_accuracy` is dropped on unmarshal with no error**, so accuracy had to go in fields that
are read:
- **negatives → `avoid`**, which the adapter routes to the **negative prompt** — a *separate
  channel*, so ~1KB of "no corbicula on a mining bee, wrong species for the subject" costs the
  length-limited main prompt **nothing**. This is the answer to the owner's context-length point.
- **positives → the FRONT of `kinds.illustration.medium`**, because medium heads the composed
  prompt, so vital detail survives truncation.

**A (next):** instruct whatever composes `spec.prompt` (`build-site-planner`'s imagery block,
`design-discovery-agent`) to name the subject's distinguishing features **and what it must not be
mistaken for**, front-loaded. Cheap, fleet-wide — but it is an instruction with no check, so it
does not stand alone.

**C (the one that holds):** generate → **vision critique** → regenerate. **Reachable as agent
CONFIG, not new code:** `execute_vision_prompt` is a registered live action
(`registry.go:1208`), backed by `aiservice.VisionCapable` / `GenerateWithImages` on Anthropic and
Gemini, and **`tool-acceptance-agent` already uses it in production**. Inputs: `images_field`,
`max_images`, `prompt_template`, `output_type`. `visual-design-auditor` is currently **text-only**
(`query_database` → `execute_llm_prompt` → `write_audit_findings`) — giving it eyes is adding a
step, not writing a provider. **Canary on one apis.uk image before proposing it fleet-wide.**

## 7. Traps this lane paid for — read before touching a page

- **`build_status='needs_rebuild'` is queue membership, not a note.** A sweep regenerated this
  page and discarded hand-edits **four minutes** after a green verification. Settle → edit →
  render → **settle again** (a render re-queues the page).
- **The renderer reads stored `rendered_html`.** Cleaning `content_data` alone re-renders
  byte-identical output. Assert both columns.
- **`COMPLETED` is the commit, not the deploy.** The Action + B2 sync run after; a cache-buster
  cannot help because it is a pipeline stage, not a cache. Compare served bytes to
  `git -C ~/projects/sites show <sha>:<domain>/index.html | wc -c`.
- **An image embedded as markup in `content_data.content` is prose the writer will overwrite** —
  25 `page_divergence_overwritten` rows proved it. Use CLC-030's fields, and lock.
- **A prohibition in a brief has no detector.** `roadmap_brief` forbade the exact string
  `A page about bees`; it served as the `<h1>` for two days. Write the `banned_claims` pattern in
  the **same edit** as the rule.
- **Never add a wildcard worker route `*.apis.uk/*`** — it would swallow the live `tools.apis.uk`
  API with no DNS change and the sweep reporting success.

## 8. Open decisions for the owner

1. **Fleet GTM re-render** — approved; sequence it (all 28 at once, or batched) — see §3 gate.
2. **Per-site tag field** — §4 design; confirm `sites.settings->>'analytics_container_id'` as the key.
3. **`brief_supplies_negation`** — recommend closing "reviewed, no change": the contrast list took
   the flagged constructions from **12 → 0**, and stripping it trades a measured benefit for a
   better detector score.

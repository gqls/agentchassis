# RUNBOOK — Imagery: best-in-class sites (operator tasks)

**What this project is about (read this first).** agentchassis is an autonomous
agent platform that plans, builds, and operates a fleet of content websites.
Its imagery pipeline generates logos, heroes, illustrations, icons, and
infographics via LLM-planned prompts and two image providers (Stability SDXL,
Google Banana), committing optimised assets to each site's git repo. The
current workstream (`PLAN_imagery_best_in_class.md`) upgrades the fleet's
visual quality to best in class: data-accurate infographics, consistent brand
language, content-linked card images, sprite-sheet bullets, copyright-safe
product sketches, news imagery, and performance budgets. Most of the work is
autonomous, but some steps need a human: credentials, key rotation, database
backups before migrations, visual approvals (logos, sprite sheets), budget
sign-off, and final eyeballing of rendered pages. **This document is the list
of those human tasks, with the commands to run.** robot-hands.com is the
testbed domain.

Conventions used below:
- `psql` means: `kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db`
- Tasks are grouped: **standing rituals** (repeat every time they apply) and
  **one-off tasks** (checked off once, tracked in the running notes).
- In a Claude Code session you can run any of these yourself by prefixing the
  command with `!` at the prompt, which puts the output straight into the
  conversation.

---

## A. Standing rituals

### A1. Refresh cluster credentials at session start
The DB-backed steps fail with `You must be logged in to the server` when your
kubeconfig auth expires (this is why the seeded bundle's schema sections are
empty). At the start of a working session run your usual cluster login, then
verify:
```
kubectl -n ai-persona-system get pods | head -5
```
If that lists pods, the agent can proceed with DB verification and migrations.

### A2. Backup before any agent-definition or schema migration
Per doc 009 convention, every migration session starts with a timestamped
backup of the rows being touched, e.g.:
```
psql -c "CREATE TABLE agent_definitions_backup_YYYYMMDD_<label> AS
         SELECT * FROM agent_definitions WHERE type = '<agent>' AND is_active = true;"
```
The agent will provide the exact statement per migration; your job is to run it
(or approve it) *before* the migration and confirm the row counts match.

### A3. Visual approval gates
Generated art the agent cannot judge alone. When asked, open the referenced
image/page and answer approve / reject / adjust:
- **Logo approval (per site, once):** the logo is generated, you approve it,
  then it is LOCKED for the life of the site (favicon and social cards derive
  from it). Rejecting triggers regeneration; approving is meant to be final.
- **Sprite sheet check (per site):** confirm each grid cell reads as a clean
  glyph at bullet size and note what each cell should mean (cell meanings are
  assigned after generation, by design).
- **Sampled page eyeball (per phase acceptance):** load the listed
  robot-hands.com pages and confirm the acceptance line in the plan (e.g.
  "cards' images match their articles", "chart axes are real data").

### A4. Refresh the context bundle when starting a fresh agent thread
```
z_bundles/imagery_seed_docs/imagery_bundle.sh
```
Run it *after* A1 so the schema/runtime sections populate. Output lands at
`z_bundles/imagery_bundle.md`. Override site/page with `SITE=… PAGE=…` env vars.

### A5. Discovery/dispatch nudges
Some verification steps need a discovery pass or a stuck item unstuck. The
agent will hand you exact SQL/kubectl one-liners; typical shapes you'll see:
```
-- re-run a discovery pass result check
psql -c "SELECT item_type, status, count(*) FROM site_work_items
         WHERE site_id = '<site>' GROUP BY 1,2 ORDER BY 1;"
```
Nothing here is destructive; run as provided and paste the output back.

---

## A6. Commands that were hard to get right (agent-facing; each with its gotcha)

*Added 2026-07-18/19. These are the ones that cost a cycle when done the obvious
way. The gotcha is the point — the command alone is not the lesson.*

### A6.1 Fire a discovery pass by hand — the reliable form
**Gotcha: `kubectl run -i --rm` silently produces NOTHING, perhaps half the
time.** The stdin attach races the container start, so the message is never
written to the topic. It looks exactly like a dead consumer, and I lost two
cycles to it. Use the detached form, passing the payload as an env var:
```bash
PAYLOAD=$(printf '{"headers":{...},"config":{"agent_type":"design-discovery-agent"},"input_data":{"site_id":"<site>","domain":"<domain>"}}')
kubectl -n kafka run kcat-fire --image=edenhill/kcat:1.7.1 --restart=Never \
  --attach=false --env="PAYLOAD=$PAYLOAD" --command -- \
  sh -c 'printf "%s" "$PAYLOAD" | kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
     -t system.agent.generic.requests -H action=orchestrate -H message_type=request ...'
kubectl -n kafka delete pod kcat-fire     # --rm is REJECTED with --attach=false
```
**Verify it was actually produced before diagnosing anything downstream:**
```bash
kubectl -n kafka run kcat-read -i --rm --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -C -b <broker> -t system.agent.generic.requests -o -15 -e -q | grep -o '"agent_type":"[a-z-]*"'
```

### A6.2 Make a COMPONENT TEMPLATE change actually appear on a page
**Gotcha: an ordinary page re-render will not do it, and will report success.**
`page-rerender` branches on `spec.reason`:
`image_landed` / `section_data_resolved` / `cta_links_stale` re-render from the
template with freshly resolved fields; **anything else reassembles the stored
`page_components.rendered_html`**, which is a rendered artifact that may be days
old. A reason-less item completes green, redeploys the page, and changes nothing.
```sql
INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity, summary,
                             spec, priority, handler_agent, status, created_by, item_key, triaged_at)
SELECT p.site_id, 'discovery', 'build', 'page_rerender', 'medium', '<why>',
       jsonb_build_object('domain','<domain>','page_id',p.id::text,'filename','<file>.html',
                          'page_name',p.name,'reason','image_landed'),   -- ← load-bearing
       5, 'page-rerender', 'triaged', '<who>', 'page_rerender_<slug>_'||p.site_id::text, now()
  FROM pages p WHERE p.site_id='<site>' AND p.name='<page>';
```

### A6.3 Verify a deploy — and validate the marker FIRST
**Gotcha: a pod-grep is a POSITIVE test only.** Some string literals (notably
switch-`case` values like `content_hero`, `sprite_sheet`) are **not retained** by
the Dockerfile build (`CGO_ENABLED=0 go build -a -installsuffix cgo`, alpine),
though a plain host `go build` keeps them. A miss therefore proves nothing, and
reads exactly like a stale deploy — this produced a false "the adapter shipped
stale" alarm on 2026-07-18 (retracted; 016b §9).
```bash
# Control test: does this marker survive a known-good build of the same pipeline?
CID=$(docker create <registry>/<svc>:<tag>); docker cp $CID:/app/<svc> ./shipped; docker rm $CID
go build -o ./control ./cmd/<svc>
for m in "<marker>" "<a log line you know is in it>"; do
  echo "$m shipped=$(strings ./shipped | grep -c -- "$m") control=$(strings ./control | grep -c -- "$m")"
done
```
`shipped=0 control=0` ⇒ **the marker is useless**, not the image. Prefer a
**log-message string** as your marker; those always survive. Best evidence of all
is runtime behaviour (e.g. the adapter logging `kind=content_hero →
provider=banana` on a real request).

### A6.4 Unstick imagery work items
**Gotcha: discovery items strand in `detected` AND in `unresolved`** — promote
both, not just `detected` (the derive items came back `unresolved`).
```sql
UPDATE site_work_items SET status='triaged', triaged_at=now(), priority=5
 WHERE site_id='<site>' AND item_type IN ('needs_imagery','needs_content_image')
   AND status IN ('detected','unresolved') AND created_at > now() - interval '15 minutes';
```
Zombie claims block the queue; clear ONLY when no long page build is in flight
(page builds legitimately run 20+ minutes — do not blanket-clear on a timer):
```sql
UPDATE site_work_items SET status='triaged', claimed_at=NULL, claimed_by=NULL, priority=5
 WHERE id='<id>' AND status='claimed' AND claimed_at < now() - interval '15 minutes';
```
**Dispatch is one site at a time against a fleet-wide pool** — priority 5 orders
*within* a site, it does not jump the queue ahead of other sites. Items sitting
`triaged` for ten minutes are usually waiting, not broken.

### A6.5 Force-regenerate a content hero (and let its card follow)
The check only GENERATES when no active content hero exists, so supersede first;
the card re-derives on the next pass by origin-staleness.
```sql
UPDATE assets SET status='superseded', updated_at=now()
 WHERE site_id='<site>' AND asset_key='content_hero_<page_underscored>' AND status='active';
```
Then A6.1 to sweep, A6.4 to promote.

> **CORRECTED 2026-07-19 — this said `avoid` fixes ground drift. It does not, and
> cannot.** The original text read: *"Style drift in the ground colour is fixed via
> the style guide's `avoid`, not its `medium`"*. **`avoid` reaches only the negative
> prompt, and Banana discards negative prompts entirely**
> (`banana/provider.go:18`), while every declared kind now routes to Banana. The
> 2026-07-18 improvement was a re-roll that happened to come out darker; the `avoid`
> edit made alongside it took the credit. Proven at n=9 on 2026-07-19: 4 of 9
> gamesdesign heroes violated their own `avoid` list (white grounds, numerals).
> Full evidence: **`/bugs_open/028`**. Caught by asking why the pilot images still
> had text in them after the guide was fixed.

---

## B. One-off tasks (current queue)

### B1. ~~Confirm the API-key scrub and rotation happened~~ ✅ DONE 2026-07-08
Keys rotated (user confirmed). Kept for the record: the 2026-05-20 finding was
plaintext logging of `STABILITY_API_KEY`/`BANANA_API_KEY` + B2 keys.

### B2. ~~Confirm robot-hands.com rebuild status~~ ✅ ANSWERED 2026-07-08
Not rebuilt; site is in poor shape. **Decision: rebuild from scratch and add
news at the same time; full latitude on this site.**
**Update (same day, evening):** the re-plan has RUN and completed — new
33-page plan with news in the header nav; content build is working through
the queue unattended. See PLAN Phase I0 status block and running notes Turn 4.

**Update 2026-07-09:** build essentially done — 14 hero images generated and
deployed to git overnight, content layer healthy. Known issue (agent-side, no
action needed from you): heroes don't render on 4 pages because some preset
components ask for image keys the generator doesn't produce
(`site_assets.background`/`product_screenshot`/etc.). Diagnosed; fix approach
in PLAN "I0 finding". You may still see empty image areas on product-detail /
about / gripper-detail / learning-center-index until that fix lands.

**Update 2026-07-10 — RESOLVED.** Two deploys later the imagery rendering is
verified end-to-end: every re-rendered page shows its own committed-path hero
(`/assets/images/hero-<page>.jpg`), zero expiring S3 URLs, and the three
corrupted component templates (content-block-about, product-specs,
info-card-grid) were regenerated via the component-creator path. One empty
image slot remains site-wide (learning-center-index listing card — Phase I3
scope). Full story: running notes Turns 8–14.

### B3. ~~Answer the seven open questions~~ ✅ DONE 2026-07-08
All answers recorded in the running notes (Turn 2) and folded into the plan.

### B4. Data-source API keys (needed before Phase I4)
Decision 2026-07-08: data sources are chosen **per domain** (each domain gets
the source appropriate to its vertical); **free-tier only for now**, paid
sources may be added later. When I4 starts, the agent proposes the source for
the domain in hand (for robot-hands.com: robotics/automation trade data;
generic fallbacks: FRED https://fred.stlouisfed.org/docs/api/, EIA
https://www.eia.gov/opendata/). Your task: register and add the API key(s) as
cluster secrets.

### B5. Budget sign-off for increased generation volume — STILL OPEN; spend to date recorded
Best-in-class imagery multiplies per-site image counts (cards, news items,
product sketches). Banana runs on a paid Google Cloud tier. Before Phases
I3/I5/I6 scale beyond the testbed, review expected volumes with the agent and
set per-pass caps you're comfortable with.

**Actual caps in force:** `contentImageMaxPerPass = 10` per site per discovery
pass, and that cap now spans BOTH page types (articles + tools), so a site with
both cannot spend double in one pass.

**Spent so far (all robot-hands, all Banana):** 4 on the D14 article pilot
(3 heroes + 1 re-roll after the white-ground drift), 3 on tool pages = **7**.

**Authorised but unspent:** you funded the tool-page rollout on 2026-07-18 at
~33 deployed tool pages fleet-wide. gamesdesign.co.uk (9) and idea.uk (1) will
draw on it automatically at their next discovery passes. The other 7 sites with
tool pages carry no `tool-list` component, so the per-surface consumer gate
spends nothing on them.

**Still unsigned:** the formal per-site/per-phase volume agreement for I5 (news)
and I6 (products), which are the phases that actually multiply counts.

### B6. ~~Logo approval for robot-hands.com~~ ✅ DONE 2026-07-11
You approved the existing (May-8) logo as-is. It is now **locked**
(`assets.locked_at` set, lock_type=permanent) — the store guard refuses any
future overwrite, making it the site's permanent mark. It already serves at
`/assets/images/logo.jpg`. **It will appear IN THE HEADER after the next
chassis deploy** (the header-resolves-logo fix, commit b00c150b, then a
site-component re-render). Remaining logo work is agent-side: favicon + OG
card derived from this locked logo (Phase I1 tail).

### B7. ~~Layout-gap decision for robot-hands.com~~ ✅ COMPLETED 2026-07-10
Your decision: no brochure fallback. Root cause was missing `industry_tags`
in the old-format classification (fixed, supersede-row); the library already
held `tool-portal-dark`. The re-composition path turned out to be unsupported
by design ("re-resolve not supported"), so the switch was executed as a
targeted `css_themes.layout_id` migration (025 pattern,
`SQL_2026-07-10_b7_layout_swap.sql`) + webdesign-agent CSS re-render/deploy.
**👀 YOUR A3 EYEBALL IS NOW DUE: open robot-hands.com — it should read as a
dark, engineered tool-portal (charcoal/electric-blue), not a formal
brochure.** HTML-side re-renders are still draining through dispatch; minor
header/footer refresh may land after your first look.

### B8. ~~Run the check-registration SQL after your NEXT chassis deploy~~ ✅ DONE 2026-07-10
Chassis deployed; `component_template_corrupted` registered on
design-discovery-agent (snapshot taken, verified in live config). The 8
remaining corrupted components fleet-wide (lobby-grid, archetype-*,
game-master-explanation, platform-comparison, provocation-card, tool-cta,
tool-guide-intro — mostly games sites) now regenerate automatically as those
sites' discovery passes run; the agent set a watch for the first emissions.
`image_source_unsatisfiable` was already registered earlier.

### B9. Infra priority nudge (your call, already scoped in your TODO)
The zombie-claim dispatch stall was the single biggest time cost of the
2026-07-09/10 verification — every batch needed manual claim-clearing
(15-min-stuck items block the whole site from dispatch). Your existing TODO
items 6/10/11 (reaper cadence bump, per-item-type circuit breaker) are now the
highest-leverage infra fix for every future build. Recommended before the next
site build.

### B10. ~~Sprite-sheet deploy (jpg switch)~~ ✅ DONE 2026-07-13
The jpg switch deployed; the sheet is live at `/assets/images/sprite-sheet-main.jpg`
(768², 75,745 bytes, under the ≤80KB budget).

### B11. ~~Sprite-sheet EYEBALL GATE~~ ✅ PASSED 2026-07-13/15
You confirmed the nine glyphs (check, gauge, gripper, cog, chart, download,
arrow, info, warning) and later confirmed the themed list bullets render
correctly on a live page. Cell map written back to the plan; **Phase I2 is
COMPLETE** (see B12).

### B12. Sprite bullets — Phase I2 CLOSED (2026-07-15). One pending gate.
I2 is complete and live: the sheet, per-page glyph bullets, a self-healing
fulfilment check (fires once, then quiet), and a "house style" so ordinary
article lists theme themselves. **You chose the neutral ARROW as the default
bullet** (check reserved for explicit use). That change is committed but rides
the next chassis deploy — it self-heals on the next discovery pass.
**👀 AFTER THE NEXT DEPLOY (your A3):** open the friction-calculator guide and
confirm the plain content-list bullets flipped from check to arrow. Fast check:
`curl -s https://robot-hands.com/assets/css/sprites.css | grep -o 'li::before[^}]*background-position:[0-9px -]*}' | head -1`
— the default should read `background-position:0px -40px` (arrow), not `0px 0px`
(check). No action if it hasn't flipped yet; it lands with the format-3 binary.

### B13. Content-loss fixes — NOT imagery; guard now LIVE, recovery still open
While doing I2 the work surfaced pre-existing platform bugs that silently lose
live content (empty product pages; article bodies stored as unparsed JSON; an
image landing that can blank an article). These are handed off to dedicated
threads — you are already driving the article-body / image-landing pair.
Written up in `../aaa_fails_to_mend/004_HANDOFF_image_landing_blanks_article_body.md`,
`../HANDOFF_2026-07-14_article_body_json_envelope.md`, and
`../HANDOFF_2026-07-14_empty_product_sections.md`.
**Update 2026-07-16: the escalate-not-blank guard IS live in prod** (verified in
the running pod) — the standing image-landing hazard is lifted. Still open in
your separate thread: 4 pages remain blanked (finetuning ×3, gamesdesign ×1;
the 004 handoff's §5 correction has the exact list and the fixed detection
query — the old `length=1326` test under-reports).

### B14. ~~Phase I3 mechanism acceptance~~ ✅ RAN AND PASSED 2026-07-16 (v1.0.1125)
Both checks fired in one pass; 9 cards derived + entity-linked; the arrow
default landed (B12 closed); `query.blog_posts` resolved; the served
learning-center-hub shows all 9 card refs. Two live findings fixed same-day:
dispatch priority is ASC (lower = sooner) and q82 cards busted the 60KB budget
(→ q78, rides next deploy). Superseded by B15.

### B15. ~~Phase I3/D13 convergence + gate~~ ⚠️ FAILED 2026-07-17 → **IMAGERY HALF CLOSED 2026-07-18/19 (D14 + F3)**

> **UPDATE 2026-07-19.** The imagery failures F1 (style consistency) and F2
> (click-through) are **fixed and verified live**; F3 (rollout) has its first two
> surfaces done. D14 moved content heroes/cards to flat duotone illustration
> under a new `content_hero` kind routed to Banana, with per-kind style-guide
> overrides; the listed-article eligibility rule removed the 404 links. Live on
> robot-hands: learning-center-hub shows 3 articles / 3 on-style cards / all
> click-throughs matching, and the front-page tool directory shows 3 tool cards.
> Card sizes 22–36KB against the ≤60KB budget. Live on **v1.0.1136**.
> **Nothing is needed from you for B15 itself.** What remains needs your
> decisions and is listed in B16.
> Detail: `SUMMARY_2026-07-19_imagery_i3_card_imagery.md`, notes Turns 48–52.

**Original failure record, kept for the trail:**
The pipeline converged exactly as designed (9 heroes → 9 auto-re-derived
distinct cards → live), but the gate failed on style consistency (mixed
photo/line-art/colour), click-through integrity (6 of 9 listed pages 404; 1
mismatch), rollout coverage, and site-level damage that is not imagery.
**Superseded by two fix handoffs** (start new chats there):
`HANDOFF_2026-07-17_i3_imagery_gate_fixes.md` and
`../robot_hands/HANDOFF_2026-07-17_robot_hands_site_fixes.md`.
Original sequence kept below for the record.

### B16. Two imagery decisions waiting on you (2026-07-19)

**B16.1 — Should the category card grid carry imagery at all?**
`info-card-grid` is the most-deployed listing we have: **15 live pages across 7
sites**. Unlike the article and tool listings it is *not* query-fed and has **no
image slot in its template at all** — it was designed as text cards. So this is a
design decision, not a build task:
- do these cards want pictures, or is the current text treatment right for
  category/navigation cards?
- if yes, where does the picture come from — the card's *linked page* (reuse the
  existing card asset, no new spend), or a generated icon/sprite per card (new
  spend, but visually lighter)?
Nothing happens here until you answer; a wrong answer is a fleet-wide visual
change across 7 sites.

**B16.2 — The I5/I6 volume agreement** (see B5 "still unsigned").

> **CORRECTED 2026-07-19 (Turn 53):** the pending tool-surface volume this was
> sized against was wrong. It is **19 generations across 3 sites**, not 10 across
> 2 — `finetuning.uk` and `leopardessconsulting.co.uk` both pass the tool-list
> consumer gate (5 deployed tool pages each) and were missed; `idea.uk` passes the
> gate but its one tool page has `deployed_at IS NULL`, so it spends nothing.
> Caught by running the check's own gate query instead of trusting the survey.

**B16.3 — Hold, style, or accept the 19 tool-header generations?** (NEW 2026-07-19)
Those three sites have **no `imagery_style_guide` at all** — robot-hands is the only
site in the fleet with one, so it is the only site where D14's `kinds.content_hero`
override exists. Everywhere else `content_hero` falls back to the free-text
`design_intent.imagery_direction` (written for photography) — the F1 style-consistency
class that failed the D13 gate. Full evidence: **`/bugs_open/027`**. Options:
- **hold** the rollout until the three sites have a style guide (recommended);
- **write the three style guides now**, then let it run — config only, live
  immediately, no image roll needed;
- **let it run as-is** and accept per-site inconsistency.
There is also a fleet-wide code half (give `content_hero` a defined default rather
than a per-site lottery) — 027 §5(a), wants the council gate, inert until an image roll.

*Not waiting on you:* `featured_article` and `product-card-with-cta` are on
**zero live pages fleet-wide**, so they need no decision and no work — the
earlier handoff's suggestion to start there was wrong and is corrected. The news
feed is Phase I5's own scope.

### B15-orig. Phase I3/D13 — per-article heroes: post-deploy convergence + gate
You chose D13: articles with no hero of their own get a GENERATED content hero
(prompt from the article's own title/description + your imagery style guide).
Built; rides your next deploy. **Expect ~9 SDXL generations on robot-hands when
it first fires (real API spend — the B5 budget question is now live, not
hypothetical).** After the deploy:
1. Trigger (or let cycle) a discovery pass → ~9 `needs_imagery` content-hero
   generations complete; each landing re-renders its article page.
2. Next pass: 9 cards RE-derive automatically at q78 (origin-staleness — the
   card's lineage no longer matches its new per-article source). Third pass:
   silent.
3. Re-render `learning-center-hub` (needs_page, priority **5** — ASC!).
4. **👀 A3 gate:** `/learning-center-hub.html` (hard-refresh) — **9 visually
   DISTINCT article cards**, each ≤60KB; click through — the article page
   carries the same image family (its own content hero).

---

## C. What you should expect the agent to do (so you don't have to)
- All Go/SQL/prompt authoring, migrations drafted with backups + verification
  blocks, deploys through the standard git → GitHub Actions path.
- DB ground-truth checks each session (after your A1).
- Updating `RUNNING_NOTES_imagery_best_in_class.md` every turn and keeping
  `PLAN_imagery_best_in_class.md` in sync with decisions.
- Flagging exactly which of your tasks (A/B numbers) are needed next, rather
  than assuming you've read the whole queue.

## The `seal_declared_field_contract` before/after (added 2026-09-04, for the 114 lane)

**What it is for.** The 114 lane is adding a declared-field carry-forward to
`save_page_sections` (opt-in key `seal_declared_field_contract`, default OFF, **0** live
`agent_definitions` rows as of 2026-09-04; consumers are exactly three pipelines, one step
each — `page-build-handler`, `page-rerender`, `tool-recreation-handler`). Arming it changes
whether `wire_page_hero_on_landing`'s value-gate refuses a row. This measures that.

### ⚠ READ THIS BEFORE RE-RUNNING ANYTHING — the obvious method is wrong

**Do NOT re-run the population filter after arming and diff the counts.** The filter is
defined by `hero_url` being page-specific, and that is *the very property the seal
changes*. A row leaving the set is then indistinguishable from a row that was never in it,
and from a row someone edited in between. **Key the comparison on pinned `component_id`s
and compare PER ROW.** This holds whoever runs it, and it is why the baseline below stores
ids rather than a number.

### ⚠ The gate has TWO conjuncts — a census of one arm answers nothing

`wire_page_hero_on_landing.go:147-148` requires **both** `hero_url` **and**
`background_image` to be in `('', $3, $5)`. A census filtering on `hero_url` alone
overstates the affected set: on 2026-09-04 that mistake produced a set of 20 of which
**8 were already refused today** (leopardess, blocked by the `background_image` arm) and
**12 were already empty in both keys** (so the carry-forward carries emptiness forward and
they stay wireable). Always select both arms.

**And the principle the whole exercise turns on: the carry-forward PREVENTS FUTURE LOSS,
it does not RESTORE PAST LOSS.** Rows already emptied by an earlier rebuild are past the
event and unaffected. The population that moves is the one that *still holds* a value.

### The baseline (pinned 2026-09-04, before arming)

- `baselines/BASELINE_2026-09-04_seal_at_risk_312_pinned.jsonl` — **312 rows**, the
  at-risk set: hero-family components whose `hero_url` is page-specific today, so the value
  is destroyed on their next rebuild now and preserved after arming.
- `baselines/BASELINE_2026-09-04_hero_gate_at_risk_20.jsonl` — the **superseded** 20-row
  set, kept because it is the worked example of the one-arm mistake above.

Each row carries `component_id`, `page_id`, `domain`, `page`, both key values, the site
fallback, `build_status` and `updated_at`.

### Re-pin the baseline (before arming)

```sql
COPY (
  WITH h AS (
    SELECT pc.id AS component_id, p.id AS page_id, s.domain, p.name AS page,
           COALESCE(pc.content_data->>'hero_url','')         AS hero_url_now,
           COALESCE(pc.content_data->>'background_image','') AS bg_image_now,
           COALESCE(s.content_data->>'hero_url','')          AS site_fallback,
           pc.build_status, pc.updated_at
      FROM page_components pc
      JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
      JOIN content_components cc ON cc.id=pc.component_id
     WHERE (cc.function='hero' OR cc.function LIKE 'hero-%'
            OR cc.function LIKE '%-hero' OR cc.category='hero')
  )
  SELECT row_to_json(h) FROM h
   WHERE (hero_url_now <> '' AND hero_url_now NOT IN ('/assets/images/hero.jpg', site_fallback))
) TO STDOUT;
```

### Read the same rows back (after arming) — per row, by id

```sql
-- $1 = the pinned component_ids, e.g. from: jq -r .component_id BASELINE_*.jsonl | paste -sd,
SELECT pc.id, s.domain, p.name,
       COALESCE(pc.content_data->>'hero_url','')         AS hero_url_after,
       COALESCE(pc.content_data->>'background_image','') AS bg_image_after,
       pc.updated_at
  FROM page_components pc
  JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
 WHERE pc.id = ANY($1::uuid[])
 ORDER BY s.domain, p.name;
```

**Reading it.** A pinned row still holding its page-specific value **after a rebuild that
went through an armed pipeline** = the seal worked. The same row gone empty = it was
rebuilt through an un-armed pipeline, or the seal did not fire. **`updated_at` is what
tells you a rebuild happened at all** — without a rebuild in between, an unchanged value
proves nothing, and that is the control this comparison needs most.

⚠ **Arming may be partial.** There are three consumer pipelines and the 114 lane spoke of
arming `page-rerender` first. A row only tests the seal if it was rebuilt through an
**armed** pipeline, so record which are armed at the time you read:

```sql
SELECT a.type, s.key
  FROM agent_definitions a, jsonb_each(a.default_config->'workflow'->'steps') s
 WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
   AND s.value->>'action'='save_page_sections';
```
Use `jsonb_each` as above — a naive `default_config::text LIKE '%save_page_sections%'`
returns **ten** types because it also catches prompt text that merely mentions the action.

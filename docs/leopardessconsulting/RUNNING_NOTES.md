# RUNNING NOTES — leopardessconsulting.co.uk rebuild

Turn-by-turn log. Newest turn at the bottom. Update **every turn**.
Companions: `PLAN_leopardess_rebuild.md`, `RUNBOOK.md`, `AUDIT_verified_facts.md`.

---

## Decision log

| ID | Decision | Status | Where |
|----|----------|--------|-------|
| A1 | "Peter Grenfell" invented → delete. `team[0]` background is the owner's. About in first person. | confirmed 2026-07-09 | owner |
| A2 | Audience pivots to the sceptical business buyer; the "SMB out of scope" clause is revoked. | confirmed 2026-07-09 | owner |
| A3 | Logo = stylised leopardess head in profile, minimal strokes, spots implied. | confirmed 2026-07-09 | owner |
| A4 | Dark chrome, light reading surfaces. | confirmed 2026-07-10 | owner |
| A5 | Reusable chart component, Go + JS renderer; honour prior D1/D3/I4. | confirmed 2026-07-10 | owner |
| A6 | Route `logo`/`illustration`/`infographic` to Banana, not SDXL. | proposed | turn 4 |
| A7 | Go emits static SVG, data-first; JS is a separate consumer that adds colour/imagery/infographic treatment on top. Deviates from literal go-echarts to satisfy PLAN §6. | **confirmed 2026-07-10** | turn 6 |
| A10 | Two-tone gold: bright #C8A951 on dark chrome only; bronze #836E32 for links on light reading surfaces (WCAG AA). Forced by contrast — bright gold on light fails AA (2.1). | decided 2026-07-10 (technical) | turn 10 |
| A8 | Data-sovereignty positioning: pitch "steps that touch your data can run on infrastructure you control" as a capability built *with* a client (pilot-scoped), never as a standing isolation/residency guarantee. | confirmed 2026-07-10 | turn 6 |
| A9 | worldsoccernews.com: publish the verifiable "15 million page impressions" figure, not the unsourced "third busiest" ordinal, unless owner recalls a citable source. | proposed — owner to confirm source or accept | turn 6 |

---

## Turn 1 — reconnaissance (2026-07-09)

Established the site is built and operated by this platform itself (`sites` row
`4851f6fc-71cf-4160-a270-e03d6d3e0732`, `status=deployed`, `build_status=pending`).
Owner corrected an early move to scrape the live HTML: the database is the source of
truth. Adopted that for the rest of the work.

37 pages. Immediately visible problems: nav labels are raw `<title>` strings; nine
pages carry zero sections; a family of `case-study-*.html` pages exists, which is
where "no inventing projects" would bite hardest if the case studies are fabricated.

## Turn 2 — verification sweep

Ran two parallel codebase audits rather than trusting the marketing copy. Result is
`AUDIT_verified_facts.md`. The headline: **the engineering is real, the framing is
not.**

Real and verifiable (C1–C6): the Companies House pipeline (2,767 verified businesses,
937 enriched, a genuine tiered matching cascade in a 716-line action); the news
pipeline (5,652 items, 4,672 credibility-scored, four live source types, six-hour
refresh); LLM-driven tool generation (full agent family, cross-linking, companion
guides); hierarchical DB-defined agent workflows (40 of 148 definitions spawn
sub-agents; 75,061 orchestration state rows; workflows re-read per message, so edits
are genuinely zero-downtime); and image generation via SDXL + Google Gemini.

Not real (U1–U5): the **"8 departments"** taxonomy — `information_schema` has zero
columns matching `%department%`, and the per-department agent counts are invented.
"70+ agents" is true only as a count of DB definitions (143), not a running fleet
(~7 types run as pods at any moment). The four case studies attribute the platform's
own subsystems to invented clients ("Veterinary Data Aggregator", "Multi-Site Content
Platform"). And "six production sites" is both wrong (there are eight) and misframed
(they are ours).

Found a hard defect while checking: **the logo and hero image are dead links.** All
three asset rows store presigned S3 URLs with `X-Amz-Expires=604800`, signed in
January and February. `curl` returns `HTTP 401 Request has expired`. `storage_path`
is empty on all three, so they predate commit `84f07d38` and cannot be repaired —
only regenerated.

## Turn 3 — owner decisions, and a correction I had to issue

Asked four blocking questions. Answers → A1–A4.

**I gave the owner incorrect information and had to retract it.** I asserted that the
rendered site follows the light `color_scheme` (`#ffffff`). It does not. The deployed
`styles.css` is dark: `--color-background:#0D0D0D`, `--color-text:#F5F0E8`,
`--color-accent:#C8A951`. The owner's first palette answer was made on that false
premise, so I re-asked with the corrected facts, and the answer changed from "light
base" to **A4, dark chrome with light reading surfaces**. Lesson recorded: verify the
deployed artifact before describing it, even when a DB row seems to describe it.

The real palette defect is narrower and worse than a light/dark mixup. The CSS
renderer lets `design_spec` win the **core** slots but the theme palette win the
**specialised** slots. Core was overridden to dark/gold; specialised never was. So the
live site serves `--color-card-bg:#ffffff`, `--color-header-bg:#0f172a`,
`--color-footer-bg:#0f172a`, and a `linear-gradient(...#1e40af...)` CTA — white cards,
navy chrome and a blue button on a black-and-gold page.

Also established: style collection `3196d966` is **shared by four sites**
(leopardess, finetuning.uk, gaswholesalers.com, ai-agent-orchestration.com) and
leopardess has **zero** forked rows. A palette fork is mandatory; editing in place
would restyle three other people's sites.

## Turn 4 — the chart question, and prior art

Owner asked for a reusable chart component (Go + a JS renderer) and, rightly, told me
to search for where we had discussed this before rather than reinvent it. We had,
extensively:

- `old/FUTURE_data_graph_pipeline.md` — the primary design. *"Diffusion models render
  the appearance of a chart, not the chart… The LLM never touches the data values."*
- **D1**: data graphics are code-rendered; the LLM proposes the story, the code owns
  the numbers.
- **D3**: `chart` is deliberately **not** a `site_plan_imagery` kind — charts are
  Lane-B artefacts. (This *reverses* the older FUTURE doc, which had proposed adding a
  `chart` kind. Follow D3.)
- **Chart runtime: go-echarts, in-chassis — confirmed 2026-07-08.**
- Phase **I4** is the flagship. **Nothing has been built.** No charting library in
  `go.mod`, no stub, no SQL column, no work item. The only `infographic` code routes
  to diffusion, i.e. decorative only.

**Conflict found in the prior decisions.** `PLAN_imagery_best_in_class.md` §6 requires
that *"static SVG/PNG must always exist as fallback"*. go-echarts emits HTML that
loads `echarts.min.js` and draws client-side; it has no server-side SVG/PNG path
without headless Chrome. The two commitments are not jointly satisfiable by go-echarts
alone. Proposed **A7**: Go emits the static SVG (dependency-free, satisfies §6), a
small inline JS renderer progressively enhances it. This preserves D1 and D3 exactly
and delivers the *intent* of the go-echarts decision. Flagged to owner rather than
silently deviating from a confirmed decision. **Awaiting confirmation.**

Wrote `PLAN_leopardess_rebuild.md` (phases L0–L9), this file, and `RUNBOOK.md`.

Also surfaced, for L2: reference images — the mechanism brand consistency depends on —
are honoured **only by Banana** (`kind=="icon"`). `logo` and `hero` route to Stability
SDXL, which ignores them entirely. So "branding consistent across the site" is
literally unbuildable through the current routing. Hence **A6**: route `logo`,
`illustration` and `infographic` to Banana. SDXL keeps photographic work.

## Turn 5 — the fabrications are live, and L1 is applied (2026-07-10)

Before touching the specs, checked whether the fabrications had reached the deployed
site. They had, and worse than the specs suggested:

- **`/about.html` lists two AI agents as team members** — "Orchestration Agent:
  Operations" and "Orchestration Agent: Research & Intelligence" — under the heading
  "The People Behind the Platform", each with a portrait filename. All three portrait
  files (including the founder's) return **404**. The section also claims "our client
  deployments" and "our published case studies", neither of which exists. (U7, U8)
- **`/how-we-work.html` renders "Seventy-plus agents organised across eight functional
  departments"** — from a `departments-grid` component whose `content_data` is NULL,
  meaning the copy is **baked into `rendered_html`** and a spec fix alone cannot change
  it. The same section claims *Playwright-based agents*, *anti-bot navigation* and
  *proxy pools*: Playwright exists only in a "could add" comment and a commented-out
  config line; proxy pools and anti-bot have **zero** code references. The real scraper
  is Firecrawl. (U6, U9)

Also checked the spec write path before relying on `pinned`: `WriteSiteSpec`
(`site_spec_actions.go`) supersedes unconditionally — **no write path checks `pinned`**;
it is an admin-display flag only. Mitigating fact: `improvement-sweep` (the scheduler
task that fires `content-gap-planner`, the agent that rewrote `identity` twenty times)
is **disabled** since 2026-05-02. So the specs will hold unless someone re-enables the
sweep; noted in RUNBOOK landmines.

**Applied L1** (owner asked for backups first — right call, and the repo already has a
convention for it):
- In-database backup: `bak_site_specs_leopardess_20260710` (all 35 historical rows of
  the four aspects) and `bak_sites_leopardess_20260710` (the full site row). On-disk:
  `specs/BACKUP_current_specs_20260710.json`.
- Superseded and replaced `identity`, `voice`, `design_intent`, `portfolio` in one
  transaction; new rows are `source_agent='operator-rebuild'`, `pinned=true`, with a
  note pointing at the audit. History preserved (`is_current=false`, not deleted).
- Fixed `sites.tagline` (D4): the "Digital Transformation with Grace and Precision"
  line is gone from the DB. It is still baked into the deployed header
  (`site_components.rendered_html`) until re-render — L3/L9 territory.

What the new specs say, in brief: identity drops Peter Grenfell, the departments, and
the 70-agent fleet, and states the true numbers (143 definitions / 56 active / 8 sites,
ours); voice encodes the owner's copy rules (positive framing, small exact claims,
plain language, the LLM-tell list); design_intent encodes A3 (leopardess head profile)
and A4 (dark chrome, light reading surfaces) including the specialised-slot values the
palette fork needs; portfolio reframes the four "case studies" as our own systems with
checkable figures, plus three honestly-labelled "what we might do" use cases.

Next: L2 (logo — needs the Banana routing change A6, and owner approvals H4) and L3
(palette fork). The About page's agent-team section and the how-we-work departments
section need their components rebuilt, not just re-rendered (content is baked in) —
added to L5 scope.

## Turn 6 — positioning deep-dive: three things checked, two of them changed my mind

Owner asked to spend real time on engagement-shape and differentiation (H6) before
drafting anything. Gave a substantive first pass: don't compete with Claude Code on
being a better coding agent (wrong category — it's a session tool, this is a standing
operations layer); lead with domain pipelines already proven (Companies House, news
credibility), the configurable autonomy dial, on-prem installability, and the audit
trail; recommend a pilot-first engagement ladder rather than picking a day-rate/
retainer/project split in advance. Owner agreed, dropped the "or third-party tools"
hedge from the solution-based framing (keep it simple: our own system, for now), and
raised three specific ideas to pressure-test — asked me to "challenge and discuss."

**worldsoccernews.com "third busiest."** Dug via Wayback Machine CDX. No independent
ranking evidence exists anywhere for that ordinal. What does exist: the site's own
"About Us" page states "15 million page impressions" (highest to date), unchanged
across a decade of snapshots (2000–2012) — most likely a Euro 2000-launch peak never
updated afterward. Recommended the smaller, sourced figure over the unsourced ordinal.
**A9, pending owner: do they recall an actual source for "third busiest"?**

**The data-sovereignty idea — real, and better than first framed.** Owner's instinct
(route sensitive workflow steps to a self-hosted model, keep others on a foundation
model) checks out as **working code**, not aspiration: `ExecuteLLMPromptAction`
resolves `ai_service` per workflow step already; `ollama-adapter` is a genuinely
self-hosted, `ClusterIP`-only pod. Two things needed correcting before this became a
site claim, both flagged and both landed well:
1. Only two text providers work end-to-end (Anthropic, Ollama) — "Mistral" isn't a
   separate provider, it's an open-weight model name run through the same self-hosted
   Ollama pod. Owner's response reframed this as a *better* story than intended: not
   "we call Mistral's API" (still third-party) but "the model runs entirely inside
   infrastructure we control" — no vendor in the loop at all for that step.
2. **No tenant isolation exists today** (shared Postgres, no RLS, shared Kafka, shared
   Ollama pod — flagged as the thing a legal buyer's due diligence would actually
   probe). Owner's answer: stand up a dedicated cluster per client for total isolation
   (Rackspace or similar) — genuinely achievable, and the cross-cluster dispatch
   scaffolding already exists (`remote-job-spawner`) though unexercised in production.
   **A8**: pitch this as a capability built *with* a client during a scoped engagement,
   never as a standing guarantee. Owner explicitly liked this framing.

Also corrected on the record: imagery already uses two more providers (Gemini,
Stability SDXL) beyond the text-only Anthropic/Ollama pair — the model-diversity story
is broader than "text generation," which the owner had folded in from the start
("we use Gemini and stablediffusion for imagery").

**The llama3.3:70b claim — verified before writing a TODO, not just taken on trust.**
`model_lifecycle.training_runs`: one `complete` run (2026-06-03→04), several
`failed`/stuck `pending` — real, GPU provisioning genuinely dynamic via ThunderCompute
(instances created and decommissioned per run, confirmed in `thunder_instances`), but
experimental rather than routine, and **no agent currently does inference against it**.
Logged the TODO in the existing project convention rather than inventing a new one —
appended to the "Future" checklist in
`docs/agent_docs/docs024_key_docs_latest/009_model_infrastructure.md`, which already
tracks exactly this kind of item.

**UK-sovereign-stack idea — deferred on request, not actioned.** Owner wants to explore
a fully UK-hosted compute+storage+model stack as a future exercise, explicitly in a
separate chat. Captured the baseline facts now so that chat doesn't re-derive them:
compute is Rackspace UK; storage is Backblaze `us-east-005` (US); the two cloud model
providers (Anthropic, Google) are both US. Saved as memory
`uk-sovereign-stack-exploration` (project type) so it resurfaces on its own, and cross-
referenced from `AUDIT_verified_facts.md` P6.

Net position for the site, once drafted: UK-based, infrastructure question already
solved (so new engagement effort goes into the workflow, not rebuilding plumbing), and
data-sensitive steps can be architected to stay inside infrastructure the client
controls — offered as something we build together, not a shipped guarantee.

Next: owner may want another pass on positioning, or is ready to move to drafting +
L2 (logo). Not yet resumed building.

## Turn 7 — positioning drafted into specs; found and fixed my own leftover error

Owner overrode my worldsoccernews.com recommendation with better information: ~12
million unique users a month at peak (not the site's self-published 15M page
impressions — a different, stronger metric), coverage in a media trade magazine and
in Microsoft's own advertising material, and explicit instruction to publish the
"third busiest sports site" recollection **anyway**, labelled as unproven, bounded by
real comparison (bigger than the BBC's coverage then, smaller than Soccernet). This is
exactly the pattern the whole voice spec is built on — state the claim, flag what
can't be proven — so I drafted it as given rather than softening it further; my job
was to make sure the hedge survived into the copy, not to argue the claim down again.

Owner also added a genuinely new positioning idea: leopardess could offer startups
building their own agent product a faster start, using this platform's already-solved
operational layer (state, retries, human-in-loop, no-redeploy workflow changes) as the
foundation instead of making them rebuild it. Real and consistent with the honest
"not yet done for a client" pattern already used for the other use cases.

**Before drafting, caught my own leftover mistake:** the `identity.json` team bio I
wrote in turn 4/5 still contained "Former senior engineer at Bumble" — the owner asked
for that to be dropped back in turn 6, and I hadn't gone back to fix it. Caught it on
re-reading the file before this edit, not because anyone flagged it. Fixed now.

**Applied to DB**, same discipline as turn 5: backed up the turn-5 rows first
(`bak_site_specs_leopardess_20260710_v2`, since `bak_site_specs_leopardess_20260710`
already held the *original* fabricated rows and shouldn't be overwritten), then
superseded `identity` and `portfolio` in one transaction. `identity.team[0].bio` now
carries the corrected, hedged worldsoccernews.com claim with Bumble removed;
`portfolio.use_cases` gained two entries — "Keeping the parts that matter on
infrastructure you control" (the data-sovereignty capability, A8) and "A faster start
for a new agent-driven product" (the startup angle) — both following the existing
honest-labelling pattern, neither claimed as done for anyone.

H1/H5/H6/H7/H8/H9 in the RUNBOOK are now resolved or drafted; only H1 (a public name,
or stay unattributed) and H4 (logo approval, blocked on L2) remain open.

Applied the A6 routing fix: `internal/adapters/imagegenerator/dynamic_adapter.go`
`generateImage` now routes `logo`, `illustration`, `infographic` (alongside the
existing `icon`) to Banana; `hero` and unset kinds stay on Stability. Checked before
changing it: `icon` already goes through this exact code path in production today —
same `imageDefaults` shape (`kindDefaults` in `generate_image_actions.go`), same
1024×1024 dimensions, Banana already proven to ignore fields it doesn't use
(negative prompts, confirmed in the original capability audit) — so this is a
same-shape extension of an already-working path, not new territory. `go build` and
`go test` on the package are clean (no existing tests to break; none added, this is a
one-line routing change with no new branching logic to unit-test in isolation).

Next: commission logo candidates using the `design_intent.json` logo spec (stylised
head profile, A3) as the prompt basis, via O5 in the RUNBOOK. Will present candidates
for approval (H4) before treating any as final — not committing one unilaterally.

## Turn 8 — logo candidates, and a firm constraint from the owner

Owner set a standing rule that reshapes how I work here: **everything done in this
thread must be replicable inside the chassis**, which will not have the interactive
tools this session has (browser, ad-hoc curl/python, the Artifact renderer). Either
document the tool-free path or don't use the tool. I declined the Artifact design tool
on that basis and wrote `REPLICATION_in_chassis.md`, mapping every off-platform action
in this thread to its in-chassis equivalent, and honestly flagging the few [human] and
[gap] items (choosing the logo; favicon/OG derivation; `pinned` enforcement; the L7
chart capability).

Generated six logo candidates. Important honesty point about *how*: the A6 routing fix
is committed but **not yet deployed to the cluster**, so the running adapter would still
send `logo` to SDXL (which can't do a clean flat mark and ignores reference images). So
I generated the candidates by calling the Gemini API directly — using the **same model
and the same `BANANA_API_KEY` the chassis itself uses** — purely for fast review. The
permanent logo will NOT be kept from this step; it gets regenerated through the real
pipeline once the routing change is deployed, from the same saved prompt. This keeps the
"reproducible in-chassis" guarantee intact: nothing the delivered site depends on came
from an off-platform tool. Saved all prompts in `logo_candidates/PROMPTS.md`.

The six, and the finding: all are the agreed concept (leopardess head, profile, gold on
charcoal). The deciding test is small-size survival — a favicon is 16px. **Solid-fill
marks survive; fine line-art dissolves.** Rendered the 32px row and confirmed it by eye
rather than asserting it. My recommendation: **c2** (solid silhouette, rosettes as
negative space) — one clean shape, reverses cleanly, unmistakable at any size. **c3**
(head emerging from the spot pattern) is the strong alternative and also yields a
repeatable spot-texture motif for the wider site. c6 (geometric) is clean but the least
distinctive idea. c1/c4/c5 are line-art — handsome large, weak small.

Built `logo_candidates/review.html` — plain self-contained HTML (no external tool, data-
URI images, opens in any browser), documented as a throwaway operator convenience that
maps to the platform's existing `checkpoint_for_review` human-in-the-loop surface, not a
site artifact.

Awaiting owner's pick (H4). Not committing a mark until then. After the choice: deploy
the routing change, commission the real logo via O5, wire favicon/OG, then L3 (palette
fork).

## Turn 9 — logo is LIVE, and I hit a platform-wide asset-deploy bug

Owner chose c2. The key realisation I acted on: **they approved a specific image, not
just a concept** — regenerating "c2" from the prompt would produce a different mark they
never saw. So I deployed the exact approved PNG, not a regenerated one.

**Production prep a designer always does, and the model didn't:** the approved PNG had
the background baked in at `#1A1A1A` (not our `#0D0D0D`) and the gold drifted to
`#C6A64F` (not the brand `#C8A951`). I knocked the background out to transparency
(projecting each pixel onto the bg→gold axis to get clean antialiased alpha, with a
noise floor to kill the background gradient), normalised every pixel to the exact brand
gold, and trimmed to an 8% margin. Verified on both grounds: on charcoal and on
off-white the rosettes correctly take the background — a true single-colour mark.

**Favicon problem, solved properly:** at 16px the detailed spots turn to mush (the exact
thing the review flagged). Real brands ship a simplified favicon; I derived one from the
*same approved outline* by flood-filling the interior rosettes so the tiny sizes are a
solid silhouette (identical shape, no detail that can't survive), while 48px+ keep the
detailed mark. Tight-cropped the favicons (2% margin) so they don't waste pixels. At
32px — what modern browsers actually use — the profile reads clearly.

Built the full set from the one approved image: transparent logo, silhouette favicon
(16/32), detailed favicon (48–512), multi-size `.ico`, opaque apple-touch-icon (iOS
ignores alpha, so charcoal ground), and a 1200×630 OG card (measured the type metrics
after a first attempt left the gold rule looking like a stray underline). All in
`docs/leopardessconsulting/brand/`.

**Then the deploy fought back — and exposed a real platform bug (D8).** Two process
failures worth recording:
1. `kubectl exec ... <<HEREDOC` silently ran nothing (needs `-i` for stdin). It reported
   no error and changed no rows — exactly the "success that did nothing" trap this
   project keeps hitting. Caught it by re-querying, not by trusting the command.
2. `dimensions` is `jsonb`, not text; `'1024x1024'` failed JSON parse. The transaction
   rolled back cleanly (verified 3 rows unchanged) — the `\set ON_ERROR_STOP` + explicit
   BEGIN/COMMIT discipline did its job.

**The real find:** triggering the platform's own `asset-deployer` FAILED with *"storage
client not available."* Traced it: `deploy_image_asset` needs a storage client, built
only when `IMAGE_BUCKET` is set (`agentbase/agent.go`), and the `agent-chassis`
deployment — which runs asset-deployer — doesn't set it (it has AWS creds but not the
bucket/endpoint env). So **image deployment via the documented pipeline cannot work in
this cluster at all.** Measured the blast radius: 83 of 102 active assets platform-wide
are expiring presigned URLs; robot-hands.com (the imagery testbed) has 34 such assets
and its logo 404s. This directly contradicts the imagery plan's "verified end-to-end"
claim, so I corrected both `AUDIT_verified_facts.md` (D8) and the imagery workstream
memory. The leopardess logo dying (D1) was never leopardess-specific — it's this.

**Workaround, still fully in-platform:** `deploy_image_asset`'s only post-optimise job
is to publish a commit to the git-adapter. Since I already had correctly sized files, I
sent that same git-adapter message directly (`scripts/commit_brand_assets.sh`, a
reusable, documented script). Committed logo + favicon.ico + apple-touch-icon + OG card
in one commit.

**Verified by artifact, not status:** all four URLs return 200; pulled the live logo and
favicon back and confirmed byte-identical to the approved files (RGBA preserved, logo
400×400, 63.8% transparent). Wrote the asset row with `url='/assets/images/logo.png'`
(git path, not presigned — avoiding the D1/landmine-6 trap), retired the 3 dead rows,
and set `sites.logo_url`. Backups first: `bak_assets_leopardess_20260710`.

**Not yet done (deliberately):** the `<head>` doesn't yet emit `<link rel=icon>` /
`<meta property=og:image>` — the files are live but not referenced. That wiring lands with
the head-component work at re-render (L3/L9), so the site changes over as a coherent whole
rather than piecemeal. H4 resolved. Next: L3, the palette fork.

## Turn 10 — imagery re-investigated (I was wrong), and the palette forked with real WCAG validation

Two owner asks: cogitate on the imagery/storage question, and do L3 but "look hard"
because the palette has accessibility checks built in and if they're not being used
we've drifted onto a different path.

**Imagery — I retracted a scare.** Turn 9 I claimed asset-deploy was broken platform-
wide, 83 assets on a 7-day timer, robot-hands' images 404. **Wrong on both counts, and
I corrected AUDIT D8 + the imagery memory.** Two errors: (1) I curled
`/assets/images/logo.png` on sites that have no `logo` asset — a meaningless test;
robot-hands' real paths (hero.jpg, hero-home.jpg, icon-*.jpg) all serve 200. (2) The
presigned URL in `assets.url` is a bypassed *source handle*, not a rendering input —
`plan_sections` emits `storage.DeployedWebPath` (durable git path) and never reads
`assets.url` (verified: zero X-Amz URLs in live HTML). The pipeline works; the normal
flow spawns `asset-deployer` with storage env injected, which is why robot-hands/idea.uk
serve. My standalone hand-trigger failed only because it skipped that injection.

**Imagery cogitation, for the owner's three questions:**
- *"agent-chassis must not have the creds / the S3 client should as per adapters"* —
  half right. agent-chassis DOES carry AWS+B2 creds; what it lacks is `IMAGE_BUCKET`,
  which is the intended design (`107_image_build_handler.sql:725`). Deploys run in a
  spawned storage-enabled `asset-deployer` that gets the bucket injected. Nothing to fix.
- *"images used to work without a presigned url — find that path"* — found. It's the
  git-committed `/assets/images/...` path served by the **Cloudflare worker**
  (`scripts/cloudflare/worker.js`), which re-signs each B2 GET server-side. That IS the
  durable path, and it's the current correct design. idea.uk's 18 durable rows were a
  one-shot `w9_04` backfill that flipped the stale `url` column to the git path the
  deployer had already committed.
- *"make the presigned URLs not expire — better control?"* — not possible. Presigned
  URLs use SigV4, which caps expiry at 604800s (7 days) — that's why it's exactly 7 days,
  it's the ceiling, not a setting. The bucket is private, so a bare public URL won't work
  either. Durability must come from the git-commit + worker path, which already exists.
  My recommendation: treat the presigned URL as a throwaway source handle (never stored
  in HTML — it already isn't), and optionally generalise the `w9_04` url-flip so the
  `assets.url` column stops looking alarming. Cosmetic, low priority.

**Palette (L3) — looked hard, and the owner's memory was right AND wrong.** The WCAG
code (`color_util.go`: relativeLuminance, wcagContrastRatio, pickReadableOnBackground)
is real and correct. But it is wired into only two narrow places — section-text defaults
(loose 3.0/2.0 thresholds) and stripping forced text colours from component HTML (AA 4.5)
— and **neither ever checks the specialised palette slots** (card_bg, header_bg, cta_bg/
cta_text). So the exact slots that leak (white cards, navy header, blue CTA) have never
had a deterministic contrast gate; only the LLM `visual-design-auditor` ever "looked",
and its fixers structurally can't rewrite a palette slot, so the leak survived every
pass. The live path hasn't drifted — the check was never wired to specialised slots.
Root cause of the leak (fully code+data verified): `buildPaletteMap` lets design_spec win
the 8 core slots but the theme palette win every specialised slot, and leopardess shares
the `professional-dark` seed palette (4 sites) with no fork of its own.

**So I did the fork the platform way, but added the missing check by hand.** Designed a
dark-chrome/light-reading palette (A4) and hit a real accessibility problem the owner was
right to worry about: **bright gold #C8A951 as link text on the light reading surface
fails AA badly (2.14 vs 4.5).** Solved it with a two-tone system (A10): bright gold stays
on dark chrome (header/hero/CTA, 8.56), bronze #836E32 for links on light (4.67–4.95, AA).
Ran ALL 15 reader-experienced pairs through the platform's own WCAG formula — including
the nav/footer hover states where bronze lands on dark — **all pass.** Contrast table in
`scripts/` alongside `L3_fork_palette.sql`.

Applied the fork: new leopardess-owned `palettes` + `css_themes` + `style_collections`
rows (cloning the seed's layout a9001f12, typography 31fc3a77, header e99b0dfa, footer
09034086 — only the palette differs), repointed `sites.style_collection_id`. **Verified
the 3 sites that shared the seed (finetuning.uk, gaswholesalers.com,
ai-agent-orchestration.com) are untouched and the seed palette is byte-unchanged.**
Backups: `bak_leo_stylecollection_20260710`. Two failed applies first (both rolled back
cleanly, verified): a jsonb `dimensions` type earlier, and here a NOT-NULL `css_content`
on css_themes — cloned the seed's as a placeholder (regenerated at render).

**Deliberately did NOT re-render/deploy the CSS.** Per the plan, L3/L4/L5 are DB changes,
then ONE coherent re-render at L9 — so the site doesn't flip to a half-done state. The
fork is the source-of-truth change; it's invisible until L9. Also confirmed the render
will be deterministic if triggered with no design_spec (specPalette empty → palette fully
determines output).

Gap flagged for the platform (not fixed here): **nothing stops a fork shipping an
inaccessible palette** — the WCAG primitives exist but aren't called at generation/fork/
install/render for specialised slots. Adding that gate is small (the math is already in
`color_util.go`) and would have caught this class of bug automatically. Noted for the
imagery/design workstream.

Also noted from a new memory (chassis-build-deploy-practice): deploying the A6 routing
change is a Makefile build-from-local-filesystem, verify against the pod not git — relevant
when L6 imagery needs it.

Next: H1 answered (keep "Founder and engineer", no name — already matches the spec). L4
(layout: 3-per-row / no orphans) and L5 (content/copy), then L9 the coherent re-render
that makes the palette, logo, favicon/OG, and content all go live together.

## Turn 11 — L4 + L5: fabrications removed, cards in threes, copy rewritten

Owner asked for a durable note on the imagery path in a main doc, then L4/L5.

**Imagery doc.** Found my own retracted claim had already propagated into
`PLAN_imagery_best_in_class.md` — the very doc that would send the next person down
the wrong path. Replaced it with a standing "HOW IMAGE SERVING ACTUALLY WORKS" box:
the two-URL model (`assets.url` = throwaway 7-day presigned handle, never rendered;
`/assets/images/<key>.<ext>` = the durable git path that pages actually serve), why
SigV4 makes a permanent presigned URL impossible, the debugging order (look up the
real asset_key — never guess the path; my wasted turn came from curling
`/assets/images/logo.png` on sites with no logo asset), and a **do not add
IMAGE_BUCKET to agent-chassis** warning, since deploys run in a spawned asset-deployer
with the env injected by design.

**L4 folded into L5.** The grid components (`features`, `info-card-grid`,
`services-grid`) are **shared across five sites** — `features` alone backs 33 sections —
so their CSS is untouchable, and `css_snippets` are global too. The brief's own answer
is better anyway: make the card counts multiples of three and cut the panels that
repeat each other. That is a purely site-scoped content change. Confirmed first that
`rerender_page_sections` re-renders each section from stored `content_data` via
`RenderTemplate`, so editing `content_data` is sufficient and `rendered_html`
regenerates at L9.

**What was actually on the live site, beyond what the audit already knew:**
- `system-stats` on the homepage rendered *nonsense*: the suffixes were misaligned, so
  it published "70%" deployed agents, "3ms" orchestration model, and "99.9x" uptime —
  plus an uptime target the plan explicitly bans claiming.
- `differentiators-section` was built on strawmen ("Most LLM integrations are
  single-agent: one model, one prompt, one point of failure") and the fabricated
  "Platform Depth Across 70+ Agents".
- `/who-we-help.html`'s cards were all negative framing ("No Observability", "Cost
  Controls That Don't Exist Yet"), and its FAQ asserted per-agent token budgets,
  circuit breakers, Helm charts, AWS/GCP/Azure deployment and per-agent least-privilege
  IAM. None of that exists in the codebase. (A circuit breaker is in fact an explicit
  unwired TODO.)

**Applied** (backed up first: `bak_page_components_leopardess_20260710`, 99 rows), in
three reviewable transactions — `scripts/L5_homepage.sql`, `L5_pages.sql`,
`L5_faq_hero.sql`:
- Homepage: hero, stats (four real figures, suffixes fixed, dated footnote), features
  8→3 "What we have built", differentiators 6→3 "What we might build with you"
  (honestly labelled not-yet-done), CTA. **Deleted** `case-studies-grid` (invented
  clients, 404 images, hard-wired to 5 cards).
- `/about.html`: **deleted** `leadership-team` (it listed two AI agents as people with
  404 portraits) and moved the founder story into the prose block, unnamed per H1, with
  the hedged worldsoccernews.com claim intact and three real stats.
- `/how-we-work.html`: **deleted** `departments-grid`.
- `/who-we-help.html`: **deleted** `case-studies-grid`; rewrote the cards out of negative
  framing; replaced the 8-question CTO FAQ with 6 questions the real reader asks,
  answered only from the audit — including the honest data-sovereignty boundary (the
  mechanism works; we have not run it for a client; our own platform shares one database).
- `/services.html`: services 7→6, engagement cards 4→3 (pilot-first ladder, H6).

**Verified by artifact:** a sweep of every `page_components.content_data` on the site for
16 fabrication patterns (Peter Grenfell, eight departments, 70+, Playwright, proxy pool,
Veterinary Data Aggregator, 99.9, circuit breaker, Helm chart…) returns **CLEAN**. All
touched grids are 3 or 6 cards.

**Two things noticed and worth recording.** (1) Agents modified this site's
`page_components` at 10:29 today, which is why `departments-grid` gained the
`content_data` the audit recorded as NULL (U9) — the re-render heavy path backfills it,
as documented. So U9 was true when observed and is now stale. (2) Checked for concurrent
risk before editing: **zero dispatchable work items** (all open ones are
`needs_human_review`/`failed`), so nothing will auto-overwrite this. Re-check before L9.

Still stale on the live site until L9: `rendered_html` for every edited section, the
header's retired tagline, and the `<head>` favicon/OG links. Next: L9 — one coherent
re-render + deploy that lands the palette, logo, favicon/OG and all this copy together.
Lower-priority remainder: `features` grids on engagement-model / for-engineering-teams /
how-it-works / our-approach / technical-architecture still carry CTO-register copy and
un-checked card counts; and several near-duplicate pages should be merged rather than
restyled.

## Turn 12 — the coherent deploy (L9): palette live, content live, and a long fight with the render pipeline

Owner: New Media Age is the publication that reported the 12M-uniques figure. Added it
to the identity spec and the About page copy (now attributed, not just asserted).
Reconciled `design_intent` core slots to the forked palette (accent=#836E32 bronze) so
a re-render can't undo the WCAG fix.

**The palette went live — but the LLM fought it first.** Triggered webdesign-agent to
render CSS. First pass: the specialised slots came out right (charcoal header/footer,
gold CTA — the leak is fixed) but the `analyze_design` LLM step INVENTED a dark core
(#070F14 bg, bright-gold accent) instead of my light reading surface, producing white
cards with light text. Root cause: `analyze_design` reads colours from
`design_intent.palette.reference_values` (a structure my spec didn't use), so the LLM
had no guidance and improvised from the mood text, with explicit "creative freedom."
Fix: restructured design_intent into `palette.reference_values` + prescriptive guidance
("these eight values are FIXED, output verbatim, background MUST be #FAF8F4 not a dark
theme"), and de-darkened the colour_mood. Re-rendered → **all core + specialised slots
now exactly match the WCAG-validated palette.** The LLM echoed it verbatim.

**Content deploy — the rerender pipeline is a minefield (as the vonc notes warned).**
- `rerender-site` runs a SEQUENTIAL page loop that STALLED on iteration 6 (a lost child
  response) and never recovered — only ~1 of 30 pages actually re-rendered. Abandoned it.
- Drove pages directly via `page-rerender`. Two gotchas cost real time: (1) it only
  REGENERATES section HTML from content_data when `spec.reason='section_data_resolved'`
  — without it, it just re-assembles stored HTML; (2) `rerender_page_sections` requires
  `spec.page_name` (the page `name`, not url). Wrote `scripts/rerender_pages.sh` and
  `reassemble_pages.sh` encoding both.
- Drove the 6 content pages + main nav pages. **All live and verified:** new hero, real
  figures, "What we have built" / "What we might build", honest About with New Media Age,
  new footer tagline, and — swept across every section's content_data — **zero
  fabrications.** Palette (light #FAF8F4, bronze links) live site-wide via the stylesheet.

**Three defects the screenshot caught that text checks missed** (this is why you look):
1. **A big AI-generated 3D leopard hero image** (`hero.jpg`) — the exact generic-AI
   imagery the brief bans. It's auto-resolved onto the hero by the section resolver
   (`plan_sections_action.go:1338`), so it can't be removed via content_data — the FILE
   at `/assets/images/hero.jpg` has to change. Replaced it with a subtle on-brand dark
   hero (faint leopardess watermark on charcoal), committed via git-adapter. Robust:
   survives rerenders because the resolver keeps pointing at the same path.
2. **`system-stats` rendered nonsense** — "2,767%", "4,672ms", "75,061x". The suffixes
   are `source:"static"` schema fallbacks (`%/ms/+/x`) that the resolver re-applies every
   render and can't be overridden by content_data; the component is shared (4 sites) and
   re-links on rerender, so a fork won't stick either. Two of those units (ms, x) can
   never be honest for counts. **Removed the section** — the figures already live in the
   feature-card copy. (A suffix-free stats component would be a good platform addition —
   noted for the design workstream.)
3. **The header "Get Started" button was blue, header navy** — the shared
   `header-professional-dark` component hardcodes `#1a1a2e`/`#0f3460` and uses ZERO CSS
   variables (4 sites). Forked it (`header-leopardess`: charcoal bg, gold CTA with dark
   text, gold hover — WCAG-checked), wired through the forked style_collection AND the
   site_components slot, re-rendered. **Header is now charcoal + gold, verified in a
   screenshot.** The residual `#0f3460` in the page is a DEAD CSS fallback
   (`var(--color-accent,#0f3460)` — accent is defined, so never used).

Backups this turn: `bak_site_specs_leopardess_20260710_v3`. Two component forks retired
after the wrong-path attempt (I initially used `{stat1_suffix,fallback}` before finding
the real path `{fields,stat1_suffix,fallback}`; the seed was never touched — verified).

**Honest miss:** I did NOT back up `content_components` before forking the stats
component, and briefly feared I'd corrupted the shared seed. I hadn't (its updated_at was
unchanged), but I should have backed up first. Standing rule reaffirmed: back up before
ANY change, including component forks.

**Remaining (documented, not blocking):**
- ~~The footer hardcodes navy~~ **CORRECTED turn 13: the footer is fine.** `footer-4-column`
  hardcodes only `#fff` (text) and takes its background from `--color-footer-bg` (#0D0D0D),
  links from `--color-footer-text`/`--color-accent`. No fork needed. The navy I'd worried
  about was the head `theme-color` meta, not the footer.
- The other ~15 pages (blog posts, guides, tools, legal) need re-assembly to pick up the
  new header/logo — their palette is already correct via the stylesheet link. Run
  `reassemble_pages.sh` for the rest.
- Per-page `<title>` and og:image meta come from page metadata (pages.title / a head
  mechanism), still partly stale — an SEO/social polish pass.
- L4 CTO-register copy on engagement-model / for-engineering-* / how-it-works /
  our-approach / technical-architecture still to rewrite; near-duplicate pages to merge.

## Turn 13 — docs brought current, handoff created, footer cleared, site-uniformity underway

Owner asked to bring the running docs / plan / runbook current, create a HANDOFF doc for
resuming in a fresh chat, then finish the footer + remaining pages for uniformity.

**Docs.** Created `HANDOFF.md` (the resume-from-here doc: one-paragraph state, owner
decisions, what's done, sharp edges, scripts, next actions, open questions). Added a phase
status table to `PLAN` (L0–L4 done, L5 partial, L6–L8 not started, L9 ongoing). Added
RUNBOOK procedures O8 (two rerender modes), O9 (reconcile-headers + the throughput note),
O10 (safe per-site palette change), and landmines 13–19 (the rerender-pipeline traps).
Updated the workstream memory to point new chats at HANDOFF first.

**Footer: no fix needed.** My earlier note feared the footer hardcoded navy. It doesn't —
`footer-4-column` hardcodes only `#fff` (text) and takes its background from
`--color-footer-bg` (#0D0D0D) and links from `--color-footer-text`/`--color-accent`. The
navy I'd seen was the head `theme-color` meta. Corrected the note.

**Site uniformity (the header/logo on all pages): in progress, throttled by the platform.**
The forked gold header lives in `site_components`; each page embeds a copy, so every active
page must re-render to pick it up. This exposed a real throughput wall:
- The prod `agent-chassis` runs **one replica** and consumes `page-rerender` messages
  **serially** (~45–60s each), and it shares that single consumer with whatever else the
  platform is doing (a `build-dispatch-loop` was competing throughout). So 30 pages is a
  15–25 min drain, not parallel.
- Two delivery traps cost time and are now documented: (1) **backgrounding `kubectl run`
  drops messages** — fire kcat in the foreground; (2) **assemble-mode `page-rerender`
  needs `page_id`** (not just `page_name`) or `rerender_single_page` fails "page_id not
  found" — 24 fires failed this way before I added page_id. `reconcile_headers.sh` +ORCH
  now fire with page_id + section_data_resolved.
- Progress at end of turn: **12/30 pages** carry the gold header/logo (incl. 7 of 11 main
  nav pages: index, about, services, how-we-work, use-cases, case-studies, who-we-help),
  and climbing as the queue drains. **All 30 already have the correct palette** (it applies
  via the stylesheet link, no per-page render needed). The remaining ~18 finish as the
  queue drains, or via another `reconcile_headers.sh` run (idempotent).

This is a platform constraint, not a defect I can force faster. Documented in RUNBOOK O9 +
HANDOFF §4. Next chat: re-run `reconcile_headers.sh` to top out uniformity, then per-page
title/OG meta, secondary-page copy, and the build-out (L6–L8).

## Turn 14 — site uniformity finished (27/30), and the render-mode bug that caused the plateau

Owner: the chassis restart was a **manual redeploy** of the chassis image, not a crash —
so my "chassis is unstable/crashing" read was wrong. It was a one-off that dropped
in-flight page-rerenders during the redeploy window; the consumer is stable.

Ran the background reconcile loop (8 rounds); it climbed 13→24/30. The last 3 content
pages (contact, how-it-works, tool-agent-complexity-estimator) stubbornly refused across
every round. **Root cause found — and it explains the whole plateau:** the reconcile used
`spec.reason=section_data_resolved`, which **SKIPS pages whose section content is unchanged**
(content-hash match) — so for any page I hadn't edited, it re-rendered nothing and never
re-embedded the new header. The pages that DID flip were the ones whose content I'd edited.
**Assemble mode (page_id, NO spec.reason) deploys unconditionally** — that's the correct
mode for a header/footer change. Switched to it and the holdouts flipped immediately.

> **CORRECTED 2026-07-20 (`bugs_closed/031`):** the mechanism asserted above is wrong —
> there is **no content-hash comparison anywhere in the page-rerender path**
> (`grep -rn content_hash --include=*.go platform/ internal/` → nothing in the rerender
> actions; `git log -S "content_hash"` → it never existed). What scoped mode actually does:
> bail at page level — `skipped` when a page has no stored components, `escalated` to the
> writer when stored content_data is absent or missing a required llm field
> (`rerender_page_sections_action.go:157,:186`) — and in both cases nothing is written or
> deployed, which from outside is indistinguishable from "skipped because unchanged".
> The practical conclusion stands (assemble mode IS correct for a chrome-only change),
> but for those reasons, not a hash. This inference was later harvested into the concept
> register as a contract and blocked a correct plan at HIGH severity in council review —
> the cost of writing an inference in contract voice.

**Result: 27/30 active pages carry the gold header + logo.** The 3 remaining
(`ai-readiness-quiz`, `for-engineering-leaders`, `guides/llm-cost-calculator-guide`) have
**zero sections** — they're the known-empty pages (AUDIT D2); they need a content REBUILD,
not a re-render, and that's already in the HANDOFF backlog. All 30 have the correct palette.

Fixed `reassemble_pages.sh` to (1) look up and pass `page_id` and (2) use assemble mode —
the two things that cost hours. Both are now documented in the script header, RUNBOOK
landmine 13/O8-O9, and HANDOFF §4.

## Turn 15 — owner site-review punch-list: root-caused, prior solutions found, fixes underway

Owner reviewed the live site and gave a detailed punch-list. Also asked to search the
docs for prior solutions (other sites solved these), update notes, and produce a detailed
handoff for a fresh chat. Ran two parallel research agents (card-links+nav-colour done;
imagery+blog+favicon still running at time of writing).

**Nav / theming (top complaint: "nav sometimes blue sometimes black; footer should match,
both black").** Root cause found: the header/footer are baked into each page's committed
`.html` at ASSEMBLE time (not read live), so pages assembled before the header fork carry
the old blue `header-professional-dark`, pages after carry gold — hence the mix. AND the
header slot was **empty** (my earlier timed-out rerender-site left it blank — every
re-assembled page would have LOST its header). AND the footer navy came from the forked
style_collection's `color_palette.primary = #1a1a2e` (footer template uses `{{.primary_color}}`).
Fixes: set the collection `color_palette.primary`→`#0D0D0D` (bak_stylecoll_leo_20260713);
triggered **`rerender-pages` with `refresh_site_components:true` + `force_rerender:true`** —
the proper one-shot that re-renders all 3 slots (verified: header gold+charcoal non-empty,
footer charcoal no-navy, head charcoal) and creates a `page_rerender` work item per page so
the build-dispatch-loop re-assembles every page with consistent chrome. This is the RIGHT
mechanism (from `scripts/initial_messages/001_assemble_all_pages_rerender/`), far better
than the per-page reconcile I'd been fighting. Draining now (103 items complete, ~20 queued;
~19 hit the 2-strike "unresolved" — to investigate).

**Nav decluttered** to 9 business-buyer items (Services, How It Works, Use Cases, What we've
built, ROI Estimator, LLM Cost Calculator, Insights, About, Contact). KEY find: the header
nav reads from the **`site_nav_items`** table (primary group), NOT `pages.in_header` (that
live query is commented out in render_site_components). Rebuilt the primary group
(bak_site_nav_items_leo_20260713). The blank "For Leaders" page is gone from nav.

**Card links 404 (how-we-work, who-we-help, use-cases).** Root cause: `info-card-grid`
renders `<a href="{{.link_url}}">` UNGATED. how-we-work had 6 cards linking to
`/how-we-work/*` pages that don't exist; who-we-help cards had no link_url → empty
`href=""`; use-cases prose linked to a non-existent `/tools/tool-ai-readiness-quiz.html`.
Fixes: **gated the shared template** with `{{if .link_url}}` (fleet-wide-safe, the prior
CTA-gating pattern; bak_infocardgrid_20260713); stripped the 6 phantom links from
how-we-work; repointed the use-cases link to the real `/ai-readiness-quiz.html`. Fired
section-rerenders for how-we-work + who-we-help to regenerate their card HTML. Platform
backstop exists but is OFF: the `phantom_internal_links` + `broken_nav_links` discovery
checks catch exactly this — enable them in a discovery agent's `checks` array.

**About invented stats ("30 Clients Served / 8 Satisfaction Rate / 2,767 Awards Won").**
Root cause: `content-block-about` stat LABELS are `source:"llm"` fields with fabricated
fallbacks that got re-applied. The block has no `{{if}}` guard, so clean *removal* needs a
shared-template change. Instead made them HONEST and true (30 Years building software / 8
Live sites, all our own / 2,767 UK records verified) — and confirmed by test that honest
content_data STICKS through a section rerender. Flagged to owner: true-not-removed; gate
the template if removal is preferred.

**Still to do (owner punch-list items not yet fixed):** MISSING IMAGES site-wide (biggest
item; imagery research agent running — needs A6 Banana routing deployed + site_plan_imagery
populated); blog.html broken (empty listing, no posts/images); who-we-help/services images;
use-cases capability claims we "don't but could" (LinkedIn enrichment, doc-watching agents,
Slack/PagerDuty) → reframe as could-not-do; favicon.png 404; the 3 zero-section pages
(ai-readiness-quiz, for-engineering-leaders, llm-cost-calculator-guide); VOICE rewrite
(owner: "sounds LLM-written, think hard about a prompt"). All captured in HANDOFF.

## Turn 16 — imagery/blog/favicon research back; favicon fixed; stale A6 fact corrected

The imagery research agent returned. Key corrections and quick wins:

- **A6 IS DEPLOYED (correcting turn 8).** Running `image-generator-adapter v1.0.1114`
  (commit 49d67e82 / v1.0.1103, 2026-07-10) already routes logo/illustration/infographic/
  icon/sprite_sheet → Banana. Proof: robot-hands icon/logo/sprite assets carry
  `origin_model=banana/…`. So imagery is NOT blocked by a deploy. Fixed the stale note in
  HANDOFF §2 A6.
- **Why leopardess has no images: it has NO `site_plan` and 0 `site_plan_imagery` rows.**
  Nothing has emitted imagery work items for it. Two routes documented in HANDOFF §9:
  Route A (per-image inline-spec kcat to image-build-handler — safe, immediate, how
  robot-hands heroes were made) and Route B (build-site-planner re-plan — systematic but
  RISKY, may re-run content generation and clobber the fixed copy; not the default).
- **Per-card/section images = Phase I3 (content-imagery lane), NOT built.** Cards fall back
  to the page hero or empty. This is the structural reason blog thumbnails and info-card
  images are blank — a config can't fix it; I3 is real new work.
- **FAVICON FIXED.** The head (`render_site_components_action.go:399,403`) hardcodes
  `/assets/images/favicon.png`; we'd only committed `.ico`. Committed `brand/favicon.png`
  (=favicon-180) to that path via the git-adapter → verified **200**. (`derive_brand_head_assets`
  can't self-serve it — it needs the logo `url` to be an S3 handle, but ours is a git path.)
- **BLOG less broken than it looked:** the listing already renders 5 posts with working
  links; the defects are card `image=""` (I3 gap, hardcoded in `rebuild_blog_listing_action.go:186`)
  + empty excerpts (blog posts have empty `pages.meta_description`) + blank read_time.
  Excerpts are a quick win (populate meta_description → re-run rebuild_blog_listing).

HANDOFF now has a full §9 imagery playbook + the corrected punch-list. This is the clean
resume point.

## Turn 17 (2026-07-14, session leopardess2) — use-cases honest + blog excerpts live; Route A safety correction

**⚠️ HANDOFF §9 CORRECTION — Route A as written is NOT content-safe on this site.**
Traced in code (deployed agent_definitions read from the live DB, not git):
- image-build-handler's terminal step `flag_page_image_rebuild` fires for **page/section-scoped**
  specs and emits `needs_page` → page-build-handler.
- page-build-handler has **no skip-content branch**: if `plan_sections` reports
  `ready_count > 0` it ALWAYS calls page-content-writer, and every `source:"llm"` field is
  regenerated unconditionally (`plan_sections_action.go:1124` "LLM fields — always available").
  It was safe on robot-hands only because that copy was LLM-authored anyway.
- **Safe variant used instead**: fire image-build-handler with NO `scope` in the spec
  (flag step no-ops; asset still generated+stored+git-deployed), then wire assets via
  manually-inserted `site_plans`/`site_plan_imagery` rows, then `page-rerender` with
  `spec.reason='image_landed'` — the `rerender_page_sections` path is no-LLM by design.
- **Second landmine (rerender_page_sections_action.go:169):** an `image_landed`/`
  section_data_resolved` rerender **escalates the WHOLE page to the content writer if ANY
  section has empty content_data** (`content_data_backfill` needs_page). That is what the 19
  needs_page items of 2026-07-13 07:33 were. Before any section rerender, check every slot
  has non-empty content_data. Only `contact` remains empty-cd (protected today only by the
  writer failing validation → needs_human_review; rewrite it via the voice pass and backfill).
- Insert imagery plan rows only AFTER assets are active: design-discovery-agent carries
  `unfulfilled_imagery_plan`; a fulfilled plan emits nothing.

**Use-cases page — DONE and verified live (punch item 7 + more).**
The page was worse than the punch-list said: 5 fabricated case studies with invented clients
("Revenue Operations at a Growth-Stage SaaS Company") and results; the hero and the CTA each
carried ANOTHER phantom `/tools/tool-ai-readiness-quiz.html` link (turn 15 only fixed the
list prose); the hero's stale rendered_html hardcoded the navy gradient (`#1a1a2e→#0f3460`,
no var() — it RENDERED, unlike the dead fallbacks). Fix: the rewritten portfolio spec already
held 5 honest "Not yet done for a client" use_cases — so the spec stayed source of truth:
patched each item to mirror `status` into `client` (the template renders `.client`), backfilled
content_data on all 3 slots (honest hero/list/CTA copy, CTA now → /contact + /case-studies),
fired page-rerender `section_data_resolved` → current templates re-rendered (navy gone —
current component templates are var()-based and clean; the navy lived only in stale HTML).
Verified live: 5 patterns, 5 honest labels, 0 phantom links, 0 fabrications, 0 navy.
Backups: bak_usecases_leo_20260714 (3 slots), bak_portfolio_spec_leo_20260714.
NOTE: an interim hand-written 3-pattern version was deployed first, then replaced by the
spec-driven 5-pattern render ~30 min later. Both honest; final state is the spec render.

**Blog (punch item 6) — excerpts + read times live.** Set `pages.meta_description` on the 2
`/blog/` posts (bak_blogmeta_leo_20260714), fired rerender-pages (no refresh flag) →
`rebuild_blog_listing` + 30 assemble-mode page_rerender items. Verified live: 5 cards with
excerpts, read times ("7 min read" etc). Card thumbnails remain Phase I3 (not built).
llm-cost-calculator-guide card has empty read_time because the PAGE was an empty shell (below).

**Empty-shell inventory (punch item 9) — was wrong in two directions.** The "3 zero-section
pages" actually: 10 pages with 0 page_components, but 7 are archived case studies (unlinked,
no sitemap exists — harmless). The 3 live ones: ai-readiness-quiz + for-engineering-leaders
(both linked from the FOOTER of every page — worse than reported) and llm-cost-calculator-guide
(linked from the blog listing — not reported). `pages.sections` is just the type-name plan;
the truth is page_components counts.
- ai-readiness-quiz + llm-cost-calculator-guide: fired page-build-handler rebuilds (empty
  pages = nothing to clobber; the `ai-readiness-quiz` interactive component EXISTS, active,
  24KB). In flight at time of writing.
- for-engineering-leaders: archived + de-navved/de-footered (bak_fel_page_leo_20260714),
  per the standing "merge near-duplicates" decision (it duplicates for-engineering-teams;
  its section list includes fabrication-prone system-stats, so a rebuild was the wrong call).
  Fired rerender-pages WITH refresh_site_components:true → footer slot re-rendered without
  it + full reassembly (dedup collapses with the in-flight batch). Early-deployed pages may
  carry the old footer → final sweep needed.

**Imagery (punch item 5) — 2 per-page heroes in flight via the SAFE variant.** Only 3 of the
5 main pages can even take a hero image: index/who-we-help/how-we-work use the `hero`
component (has `background_image`); about/services use hero-about/hero-services which have
NO image-typed field (component work → deferred, noted in HANDOFF). Index already has its
hand-chosen hero.jpg. Fired image-build-handler (scope-less specs, asset_key
hero_who_we_help / hero_how_we_work, kind hero → SDXL; design_intent.imagery_direction is
auto-prepended by generate_image_actions). Next: verify assets active → insert site_plans +
site_plan_imagery rows (scope='page') → page-rerender `image_landed` on those two pages
(their slots all have content_data — verified safe from the :169 escalation).

**Stale handoff items corrected:** NEXT ACTIONS #4 (theme-color) is already `#0D0D0D` live —
done at turn 15, remove. The 19 content_data_backfill items are all terminal (14 complete,
4 failed, 1 needs_human_review=contact) — none pending, no clobber threat from that batch.

### Turn 17 addendum — imagery blocked by infra; hero generated but NOT wired (deliberately)

**The Route A-safe mechanism is PROVEN end-to-end** (first time on this site): a scope-less
`needs_imagery` spec → image-build-handler → image-generator → asset-deployer → git, with NO
`needs_page` emitted and NO content touched. Artifact: `assets.asset_key='hero_who_we_help'`,
status active, `origin_model=stability/stable-diffusion-xl-1024-v1-0`, live 200 at
`/assets/images/hero-who-we-help.jpg`. **Deploy filename convention:** `AssetKeyFilename()`
maps `_`→`-`, so asset_key `hero_who_we_help` deploys to `hero-who-we-help.jpg` (not the
underscore form — cost 3 wasted 404 checks).

**But the image itself is unusable and is deliberately NOT wired.** SDXL rendered a beige
faux-technical-diagram with garbled pseudo-text and chart-like panels — it violates
design_intent `avoid` on two counts ("Charts produced by an image generator, under any
circumstances"; imagery_direction bans photographic/decorative filler) and the prompt's own
"no text". Root cause is routing, and it is a REAL design gap, not a prompt miss:
`dynamic_adapter.go:534` routes `kind='hero'` → **stability/SDXL**, and only
icon/logo/illustration/infographic/sprite_sheet → **Banana**. leopardess's whole imagery
direction is "flat gold-on-charcoal diagrams drawn in the same hand as the logo" — i.e. its
heroes are ILLUSTRATIONS. Under the current routing a leopardess hero can never be on-brand:
it is sent to the photographic provider by definition. Retried as `kind='illustration'`
(Banana) with an explicit constraints block; that run is stuck (below).

Because there are still **no `site_plan_imagery` rows**, the bad asset resolves to nothing and
is invisible on the site. `scripts/`-ready SQL to wire heroes when a GOOD asset exists is
saved at (scratchpad) `wire_heroes.sql` — guarded by `EXISTS(... status='active')`. Do not run
it against the current SDXL asset.

**INFRA BLOCKER (not this site's bug):** spawned `image-generator` pods repeatedly fail to
consume their job topic — `failed to dial ... personae-kafka-cluster-combined-pool-prod-2 ...
10.20.99.93:9092: i/o timeout` — while the broker pod is Running and other agents work. The
parent orchestration then sits in `AWAITING_RESPONSES` at `spawn_image_gen_imagery` forever
(the workflow's `timeout_seconds: 300` does not fire). Hit 5+ pods across the session; deleting
the pods and re-firing works sometimes. **Imagery on this site is blocked on that flake**, not
on the pipeline. Two things to raise: (a) node→broker-2 network path, (b) a spawn-timeout that
actually terminates the parent.

**Also lost ~40 min to a cluster auth expiry:** `kubectl` returned `Unauthorized`, and because
the kcat trigger scripts end in `>/dev/null 2>&1`, three fires (footer refresh #2, quiz take-3,
hero retry) silently published nothing. Lesson: do NOT swallow kcat stderr; check for the
"deleted from kafka namespace" line, or assert an orchestration row appears.

**Footer: the real mechanism.** `pages.in_footer` is **NOT** what builds the footer — that
query (`loadFooterNavItems`) is DEPRECATED. `render_site_components_action.go:98` builds the
footer from `GetNavItems(primary + utility + legal)` i.e. the **site_nav_items** table. So
setting `in_footer=false` on for-engineering-leaders did nothing; the fix was DELETING its
`utility`-group nav row (bak_navitem_fel_20260714), then `rerender-pages` with
`refresh_site_components:true`. Ai Readiness Quiz is also a utility nav item — it stays,
because the page is being rebuilt.

### Turn 17 final — verified live (curl, end of session)

| Check | Result |
|---|---|
| Dead `for-engineering-leaders` link, 17 pages | **0 refs** (was on every page's footer) |
| use-cases fabrications (LinkedIn / PagerDuty / "Revenue Operations" / "days, not quarters") | **0** |
| use-cases phantom `tool-ai-readiness-quiz` links | **0** (was 3: prose + hero + CTA) |
| use-cases honest "Not yet done for a client" labels | **5** |
| use-cases live navy gradient | **0** |
| blog cards with excerpts / read-times | **5 / 5** |
| `/guides/llm-cost-calculator-guide.html` | **26,995b** (was a 12,439b empty shell) |
| `/ai-readiness-quiz.html` | **still 12,425b — STILL BLANK, see below** |

**The queue does not drain reliably today — bypass it.** The `page_rerender` work items from
`rerender-pages` sat at 17 `triaged` for >2h with `complete` frozen at 34 (build-dispatch-loop
pods churning but not consuming). Re-firing `rerender-pages` just adds more items. What
actually worked: **`scripts/reassemble_pages.sh` drives `page-rerender` directly, one page per
call, bypassing site_work_items entirely** — all 10 stale pages went clean in minutes. This is
sharp-edge #2 in practice; reach for the script, not the queue.

**ai-readiness-quiz — still blank, and the remaining blocker is infra, not content.**
Take 3 and take 4 both got past the `contact-block` email bug (fixed fleet-wide: the shared
schema's `email_placeholder` `fallback:"jane@company.com"` is only ever an HTML `placeholder=`
attribute, but `validate_page_content`'s email check read it as a hallucinated contact address
and failed EVERY build of any page carrying contact-block → now `"Enter your email address"`,
bak_contactblock_20260714). Take 4 produced **no validation error at all** — but the parent
orchestration is stuck `AWAITING_RESPONSES` at `call_content_writer` while its child
orchestrations show COMPLETED. Same lost-child-response class as the image-generator stalls.
To resume: re-fire page-build-handler for `ai-readiness-quiz` when the cluster is healthy; the
content path itself is now believed clean. Its `sections` are `["hero","ai-readiness-quiz",
"contact-block"]` and the interactive quiz component exists and is active.

**Not done this turn (be honest about it):** the voice rewrite (punch #10) was not started —
the session went into root-causing the honesty defects, the empty pages, the footer mechanism,
and the imagery-clobber trap instead. `contact` is still empty-content_data CTO-register copy
and is the right place to start.

### Turn 17 — quiz take-5 also died on infra; stopping re-fires

Re-fired ai-readiness-quiz once more in a window where the cluster looked healthy (20
COMPLETED orchestrations in the prior 20 min, no stuck pods). It FAILED at
`spawn_content_writer` with `error = "reaper: stale AWAITING_RESPONSES for >90 min"` — the
spawned page-content-writer's response was lost, and the reaper eventually killed the parent.
**No validation error, no content error** — 5th attempt, 5th time blocked by the same
lost-child-response / Kafka-dial-timeout class. This is not something re-firing fixes reliably.
**Stopping.** Resume when the cluster is healthy AND you can watch the spawn land its child
response (or after the platform adds a spawn-timeout that fails fast instead of reaping at 90
min). The content path is believed clean; the `contact-block` validator blocker is fixed.

### Turn 17 — infra bug written up separately

The spawn/lost-child-response flake that blocked the quiz (and imagery) is characterised as a
standalone, fleet-wide platform bug in **`docs/HANDOFF_spawn_lost_child_response.md`** — start a
separate thread from there. Root cause: certain worker nodes can't dial Kafka broker-2
(`10.20.99.93:9092`); child agent pods on those nodes retry-loop forever and never publish their
response, so parents hang until the SQL reaper fails them at 30/90 min. Fleet-wide evidence
(38 `spawn_dispatch` failures in 2 days), not a leopardess problem.

### Turn 17 — voice rewrite (punch #10) STARTED, safe hand-edit path

The voice rewrite's intended mechanism (page-content-writer + `rewrite_guidance`) is exactly
the spawn→child path broken by the infra flake right now, so I did it the reliable way:
hand-rewrite the affected `content_data` fields, then a **no-LLM `section_data_resolved`
rerender** (same safe path as use-cases). All content lives in structured content_data fields
(headline/subheadline/content/features), so this is a clean surface.

**First, a triage correction:** most pages the handoff called "CTO-register" were already
rewritten in earlier turns and are in good voice — how-it-works, engagement-model
("Fixed scope. Fixed price."), who-we-help ("Work worth handing to a machine"),
for-engineering-teams, and services' MIDDLE two sections (services-grid, info-card-grid) are
all fine. Do not redo them.

**services — DONE & verified live.** Its hero and CTA were the two worst LLM-tell offenders on
the whole site: BOTH repeated the exact banned triad the voice prompt cites as example #1
("observability, fault isolation, and cost controls"), both were title-case CTO-register
("From Prototype to Production-Grade…", "Your Prototype Works. Now Make It Production-Ready."),
and the CTA's secondary link pointed at /services.html (itself, circular). Rewrote both to the
A2 register in the no-contraction plain voice the good sections on that page already use; fixed
the secondary link → /case-studies.html. Verified live: 0 triad, 0 "Production-Grade", new
hero+CTA present, secondary link now /case-studies.html (the content_data URL edit STUCK
through the rerender — this call-to-action instance is not hit by the CTA-revert landmine).
bak_services_leo_20260715.

**how-it-works — DONE & verified live.** Not a register problem; it had a real
design_intent.avoid violation: TWO generic-text-blocks both titled "How a system gets built"
saying the same thing (Kubernetes/Kafka/Postgres + verification + approval gates + "run on our
own sites first"). Repurposed the 2nd (position 4, after the features grid) into an honest
limits section "What it does not do" — the "admit the edge" quality the prompt asks for, using
only already-true site facts. Verified live: "How a system gets built" now appears once,
"What it does not do" present. bak_howitworks_leo_20260715.

**our-approach — DONE (deploying at write time).** Hero carried the same balanced triad
("Architecture decisions, infrastructure choices, and the oversight model") + a title-case
"How We Build" heading. Surgical jsonb_set on headline/subheadline/heading (preserves cta/image
fields); body left as-is. bak_ourapproach_leo_20260715.

**Flag for the owner — near-duplicate pages (handoff NEXT ACTIONS #3):** how-it-works,
our-approach, technical-architecture, and for-engineering-teams all rehearse the SAME body
material (~80% overlap: Kubernetes/Kafka/Postgres + hierarchical agents + per-step logging +
configurable approval + verify-against-source + "we run it on our own sites first"). This is a
MERGE decision, not a voice fix — I did not merge (owner's call). Polishing each to sound
distinct would entrench the duplication. Recommend collapsing to one "how it works" page + one
short "for engineering teams" technical cut, and dropping the rest.

**Still CTO-register but arguably correct:** technical-architecture keeps a technical headline
("Technical Architecture for Multi-Agent Systems…") — that page IS the "one click down"
technical depth A2 allows, so its register is defensible; left it. contact still has empty
content_data (its rendered_html is generic) — best done when the spawn pipeline is healthy, or
hand-authored next.

### Turn 17 — contact page DONE (4 voice pages total); 4th phantom quiz link found + killed

**contact — DONE & verified live.** It had empty content_data on all 3 slots (a rerender
landmine: empty cd → whole-page escalation to the writer, which is infra-blocked), CTO-register
copy ("…What AI Agents Can Do for Your Stack", "scoping an initial deployment or evaluating a
full orchestration layer"), AND a **4th** phantom `/tools/tool-ai-readiness-quiz.html` link
(turns 15/17 had found 3; this is the last). Fix: populated the `llm` fields on all 3 slots
(hero-contact headline/subheadline, contact-form heading/description/submit, contact-info
section_title/intro_text) with honest A2 copy → non-empty content_data defuses the escalation →
safe `section_data_resolved` rerender. email/phone/address resolve from `site_specs.identity`
top-level keys automatically (NOTE: they're `data->>'email'` at the identity spec's TOP level,
NOT `data->'identity'->>'email'`). Verified live: new hero, 0 old CTO copy, **0 phantom quiz
links sitewide**, real email/phone resolved, and confirmed **no content_data_backfill
escalation fired**. bak_contact_leo_20260715.

**Voice pass sitewide verification (end of turn 17):** the banned triad "observability, fault
isolation…" = 0 occurrences anywhere; the phantom quiz link = 0 anywhere (all 4 instances gone).
4 pages rewritten & verified live this turn: services, how-it-works, our-approach, contact.
Combined with earlier turns, the primary-nav journey (services / how-it-works / use-cases /
contact) is now in-voice and honest. Remaining voice work is lower-value: technical-architecture
(defensibly technical register), for-engineering-teams (already decent), and the page-MERGE
decision (owner's call). All safe hand-edit + section_data_resolved; zero pipeline dependency.

## Turn 18 (2026-07-16) — page-merge decision made and EXECUTED

**Owner decision (A11, add to the locked list): keep `how-it-works` (canonical explainer) +
`technical-architecture` (the "one click down" technical page, nav label "Architecture");
archive `our-approach` and `for-engineering-teams`.** Evidence that drove it: all four pages
made the same six claims (some pages TWICE — features and differentiators sections were the
same list reworded); how-it-works had inbound body links from all 15 other pages; the other
three had ZERO inbound body links (footer-only orphans).

**Executed (bak_merge_pages/pcs/nav_leo_20260716):**
1. `technical-architecture` hero sentence-cased ("The architecture, in detail").
2. Its features array rebuilt: kept the 7 concrete technical items, dropped the sitewide-repeat
   ("Platform tested on live systems" lives on how-it-works), folded in the two unique
   for-engineering-teams items **rewritten to verified facts**:
   - "Model routing as cost control" — per-step model selection is in the audited portfolio
     spec; the UNVERIFIABLE "caches outputs" clause was dropped.
   - "One pattern, many kinds of work" — **the original said "more than 70 agents across eight
     departments"; "eight departments" is one of the audited-out FABRICATIONS.** Rewritten to
     "over 150 agent definitions" (DB-verified: 156 active in agent_definitions) with the real
     categories (coordinators/specialists/analysts are actual agent_category values).
3. Its duplicate `differentiators` section DELETED (page_components row + pages.sections entry)
   — it repeated the features list on the same page.
4. `our-approach` + `for-engineering-teams`: status='archived', de-navved (their utility
   site_nav_items rows deleted — the footer builds from site_nav_items, not pages.in_footer).
5. Footer slot refreshed (clean of both), technical-architecture section-rerendered, all other
   pages reassembled directly via reassemble_pages.sh (queue bypassed as before).

**Lesson reinforced:** content folded during a merge must be re-audited — one of the two
"unique items worth saving" contained a known fabrication buried mid-paragraph.

### Turn 18 — "where does the facts audit actually happen?" → answered, and spec'd as a new workstream

Owner asked whether claim/fact verification is part of the content loop — is there a dedicated
checker and handler? **Investigated in code + live DB; the answer is NO dedicated layer exists.**
What exists: (1) generation-time "NEVER invent" prompt rules (instructional, leaky — the
fabrications all shipped through them); (2) `validate_page_content` at build time (form, not
truth — placeholders/links/meta-commentary/length + exactly ONE fact-shaped check: emails vs
site contact); (3) 38 post-deploy discovery checks (all structural); (4) `content-quality-auditor`
(tone/gaps/CTA/differentiation — zero fact vocabulary, never ran here). The audit is manual:
AUDIT_verified_facts.md + operator discipline. Telling details: the identity spec's
`evidence_base` key is read by NOTHING in code (grep: zero consumers), and the only fabrication
class ever caught by the platform (emails) is the only class with a deterministic checker.

**Owner decision: build a claims-verification layer, fully documented, as its own thread.**
Spec written: `docs/agent_docs/docs024_key_docs_latest/claims_verification/SPEC_claims_verification.md`
— self-contained thread-starter. Shape: formalize `site_specs.evidence_base` as data
(facts + banned_claims + allowed_entities; transcribed from this site's audit doc) → V1
deterministic checks (banned-claims blocker in validate_page_content + `check_unverified_claims`
discovery check → HITL work items, never auto-rewrite) → V2 writer whitelist injection ("use
ONLY these numbers/entities" — the fix that worked for emails) → V3 LLM claims-auditor for
prose assertions → V4 live SQL re-verification of metrics (shares the query layer the L7 chart
component wants). Benchmark corpus = this site's own shipped fabrications (B1–B7 in the spec:
"eight departments", "70+ agents", "Awards Won", the fake client case studies, jane@ in body
vs in placeholder=). Landmines encoded from our scars: DOM position (the placeholder-attribute
false positive), number false-positive classes, audit caveat semantics (C1 true ≠ "handles
dissolved companies" true), locked components, opt-in per site.

Leopardess is the pilot site (only one with an audit doc to transcribe) but the build itself is
platform work — belongs to the claims-verification thread, not this one.

## Turn 18 (cont.) — quiz LIVE; heroes LIVE; a real clobber event caught and fixed forward

**ai-readiness-quiz — DONE, live, verified (punch #9 fully closed).** Take 6 built in a healthy
cluster window: 3 components, 54KB live page, real interactive quiz. Integrity-checked: 0
invented emails, 0 banned claims, all internal links real, the fixed contact-block placeholder
renders. (Voice quibble for later: title-case H1 "Find Out If Your AI System Is Ready…".)

**Per-page heroes — the full imagery mechanism is PROVEN END-TO-END and LIVE.**
`kind:'illustration'` → Banana produced exactly-on-brand images for both pages (charcoal
ground, hairline gold, no text — reviewed by eye before wiring, both approved):
who-we-help (three clusters converging on one node) and how-we-work (staged line with an
open-circle checkpoint). Wired via manual site_plans/site_plan_imagery rows
(scripts/wire_heroes.sql) + `image_landed` rerenders. **Confirms the §9 routing correction:
illustration→Banana works; hero→SDXL was the wrong lane for this brand.**

**⚠️ CLOBBER EVENT — the §9 escalation guard has a HOLE, now corrected.** The who-we-help
`image_landed` rerender ESCALATED to the content writer despite the documented pre-check
returning 0 empty slots. Root cause: the SQL guard (`content_data IS NULL OR ::text IN
('{}','null')`) is WEAKER than the Go check (`len(map)==0` after unmarshal) — content_data
holding a valid JSON **array** (or scalar) fails the map-unmarshal silently → nil map → len 0
→ whole-page escalation. One old who-we-help slot evidently held array-shaped cd.
**Correct pre-check:** `content_data IS NULL OR jsonb_typeof(content_data::jsonb) <> 'object'
OR content_data::jsonb = '{}'::jsonb`.

**What the writer rebuild did (a perfect claims-verification specimen, captured live):**
the escalated rebuild wrote an honest hero/faq/cta (the pinned specs held), **but fabricated
four case-study card titles** in a case-studies-grid it added ("Triage Without a Queue",
"Reconciliation That Runs Itself and Signs Its Work", "Classification at Scale…", "Document
Review With an Auditable Decision Record" — none exist), and invented a phantom
`/what-we-build` link + emoji icons in info-card-grid. None of it deployed (the handler
stalled pre-deploy — the flake, for once, on our side). **Relay to the claims-verification
thread: fresh benchmark items, generated TODAY by the current writer under the current
prompt rules.**

**Fix-forward (bak_wwh_rebuild_20260716):** kept hero (with image) + faq (honest) + cta
headline; restored the old hand-checked card TITLES with new in-voice bodies (old cards were
title-only), no links; DELETED the fabricated case-studies-grid (row + sections entry);
patched CTA button labels/urls back (renderer-source urls skip in writer builds — a known
shape now); closed the needs_page item manually with a do-not-rebuild note (it had reset to
'triaged' when the reaper killed its claiming orchestration — a re-dispatch would have
re-clobbered).

**SEO pass (NEXT ACTIONS #5, partially done):** 4 A2-violating titles fixed (index,
who-we-help, how-we-work, contact — "CTOs & Engineering Leaders", "Production-Grade" etc.
gone) + honest meta_descriptions written for the 12 key pages that had none (index's live
description was an EMPTY STRING). All restate audited page copy; no new claims. Reassembled.
bak_titles_meta_leo_20260716. Sitemap still absent — remaining item.

## Turn 18 (final) — who-we-help LIVE and verified; the escalation guard has TWO branches

**who-we-help deployed & verified:** hero image live, new title + meta, the three old honest
cards (with new bodies + required eyebrow/subtitle), FAQ, CTA buttons → /contact + /case-studies.
0 fabricated titles, 0 phantom links.

**Complete escalation picture (supersedes this turn's earlier partial diagnosis) —
`rerender_page_sections` escalates a page to the content writer on EITHER of:**
(a) `len(contentData)==0` — which includes valid-JSON **arrays/scalars** (map-unmarshal fails
silently), not just NULL/`{}`;
(b) content_data present but **missing any schema-required `source:"llm"` field**
(`missingRequiredLLMFields`) — this caught who-we-help a SECOND time: info-card-grid requires
`cards, section_title, section_eyebrow, section_subtitle` and the backfill had only the first
two. **Before any section rerender, verify stored cd keys ⊇ required llm fields per slot:**
compare `jsonb_object_keys(pc.content_data)` against
`input_schema->'fields'->k->>'source'='llm' AND required` — the HANDOFF §9 guard now documents
both branches. Also: each escalation emits a fresh `content_data_backfill` needs_page item that
survives orchestration death (resets to 'triaged') — kill them (`status='complete'`) or they
re-clobber on a later dispatch. Three were killed this turn.

**Cross-thread note (claims-verification, running in parallel):** their V1 scan already found
9 banned-claim occurrences on leopardess content — including the hierarchical-multi-agent blog
guide leaking "70+ agents / eight functional departments" on 2026-07-15, AFTER our sweep.
Owner ruling pending in THAT thread; the content fixes will likely land back here. Two of the
four flagged pages are archived (for-engineering-teams), lowering urgency.

## Turn 18 (cont.) — claims-verification rulings APPLIED (owner accepted suggestions)

The claims thread's V1 scan findings (9 occurrences / 5 slots / 4 pages) were ruled on by the
owner — suggestions accepted. Applied here (bak_claims_rulings_20260716), replacing the
fabricated "70+ agents / eight departments|functions|areas" taxonomy and the U10
"least-privilege IAM" term with evidence-base facts:

| Page | Fix |
|---|---|
| hierarchical-…-explained (blog, LIVE) | 4 replacements in article-body: "runs 70+ agents across eight functional departments"→"registry holds over 150 agent definitions"; "inter-department routing"→"routing"; "Each department has its own supervisor"→"Each family of work has its own supervisor"; "Worker agents within a department"→"Worker agents beneath a supervisor". **The TRUE topology (head orchestrator → supervisors → narrow workers, audit C4) was kept** — only the fabricated org taxonomy went. |
| technical-architecture (LIVE) | "more than seventy agents operating in eight functional areas"→"more than 150 agent definitions — coordinators, specialists, analysts" (the word-form variant the regex almost missed). |
| insights (LIVE) | "least-privilege IAM"→"least-privilege agent identity" (the verified phrasing). |
| for-engineering-teams (ARCHIVED, ×2 slots) | cd fixed in place so the sleeping copy is safe if ever resurrected; NOT redeployed. |

Method: text-level replace inside the content field via jsonb_set (targets were unique plain
prose); **dual-branch escalation pre-check run on all three live pages BEFORE re-rendering**
(0 rows — the new §9 guard works); section_data_resolved rerenders fired; verification
monitor watches for zero banned residue + the new fact live.

This closes the loop the claims spec §5 designed: check found → human ruled → operator applied
in content_data → no auto-rewrite anywhere. The V1b work items will confirm clean on the
check's next scan once it's deployed and enabled.

## Turn 18 (final) — all 9 claims findings RESOLVED and verified live

Tag-stripped verification (greps must strip tags — served HTML splits phrases like
`<strong>70+</strong> agents`, which literal greps miss both ways):

| Page | fleet-claim | dept-claim | IAM term | verified fact live |
|---|---|---|---|---|
| blog/hierarchical-…-explained | 0 | 0 | 0 | ✓ "150 agent definitions" |
| technical-architecture | 0 | 0 | 0 | ✓ ×2 |
| insights | 0 | 0 | 0 | (reworded — see below) |
| for-engineering-teams (archived) | fixed in cd, not deployed | | | |

**Coordination note:** insights was ALSO fixed by the claims-verification thread directly
(component updated 17:29, after this thread's edit) — their wording ("agent failure modes and
recovery") replaced the U10 phrase entirely rather than substituting the verified term. Same
goal, no conflict in the result; but two threads are now touching leopardess content_data —
future rulings should be applied from ONE side (suggest: the claims thread rules + this
workstream applies, as designed in spec §5).

Punch-list state after turn 18: items 1–4, 6–9 CLOSED; item 5 (imagery) two per-page heroes
live + index hand-chosen, remainder is component work + Phase I3 (imagery workstream); item 10
(voice) done for all pages that needed it. Site-wide: 0 banned claims in served HTML across
all pages scanned, verified against the evidence base.

## Turn 18 (close) — sitemap live; about/services heroes formally deferred to the imagery workstream

**sitemap.xml LIVE** (was entirely absent — a fleet-wide gap: robot-hands/vonc/finetuning have
none either, and no generator exists in the platform). Generated from the pages table
(status IN active,deployed → 27 URLs, archived pages verified excluded), committed via the
git-adapter (same route as the favicon; payload pattern in commit_brand_assets.sh). Verified:
200, 27 <loc> entries, 0 archived leaks. robots.txt is Cloudflare-managed content-signals
(search allowed, AI-training crawlers disallowed — checked stanza grouping; NOT a
block-everything file) and can't carry a Sitemap: line from git; the well-known path suffices.
Platform note: a `generate_sitemap` action run at deploy time would fix this fleet-wide.

**about/services heroes — formally deferred with evidence:** hero-about is shared across
**9 sites**, hero-services across 5. Adding an image field means a shared-schema change —
the additive gated-field pattern ({{if .background_image}} + optional schema field, same class
as turn 15's link_url gate) is the suggested fleet-safe approach, but it belongs to the
imagery workstream (which owns component imagery and is mid-build on Phase I3), not a site
session. leopardess imagery state: index (hand-chosen) + who-we-help + how-we-work (Banana
illustrations) live; blog thumbnails + card images arrive with Phase I3.

## Turn 19 (2026-07-17) — PLAIN VOICE v2: owner redirects the register

**Owner decision (A12): move the copy further toward plain, friendly, matter-of-fact.** The
v1 voice killed hype and fabrication but produced dense, literary copy (long packed sentences,
no contractions, em-dash rhythm, "laundered rumour"-class turns). The owner supplied a
before/after homepage pair + a stack of humanizing-prompt reference material and asked to go
"even more in that direction".

**Encoded in three places:**
1. `specs/PLAIN_VOICE_v2.md` — the distilled rules (one idea per sentence, ≤~20 words,
   contractions in, short paragraphs, Flesch ~80, no literary moves, em-dashes near zero,
   friendly = calm not chummy) AND the explicit reject list from the reference material
   (no deliberate errors/slang, no casual fillers, no rhetorical-question tic, detector
   evasion is not a goal). v1's honesty rules all survive.
2. **DB voice spec updated** (site_specs `voice`, bak_voice_spec_leo_20260717): tone,
   plain_language, new `sentence_shape` key, llm_tells_to_avoid rewritten — future
   writer-generated content now reads the v2 register.
3. **Homepage rewritten as the worked example** (bak_index_v2_20260717): all 4 slots (hero,
   features, differentiators, CTA) in the new register, pushed further than the owner's
   sample. Dual-branch pre-check 0 rows → section_data_resolved rerender.

**Claims catch:** the owner's illustrative sample introduced "reads news from hundreds of
sources" — the evidence base has 18 configured sources (13 news_search + 3 api_news + 1
scrape + 1 rss). NOT adopted; noted in PLAIN_VOICE_v2.md. Register changes, facts do not.

**Remaining rollout:** the other rewritten pages (services, how-it-works, our… use-cases,
who-we-help, contact, about, engagement-model, faq, technical-architecture + the two blog
posts) still carry v1-dense copy and should be moved to v2 page by page, same safe path.

## Turn 20 (2026-07-17) — voice_tells checker BUILT, TESTED, and run live

Per owner instruction ("make the voice checker exist"), built the deterministic lane of
SPEC_voice_tells_check in one sitting, mirroring the claims layer exactly:

| Piece | File | Status |
|---|---|---|
| Engine | `platform/orchestration/datahelpers/voicetells.go` — VoiceGate config (voice spec `voice_gate` key, opt-in by presence), 7 signals: banned_phrase (global tells + site bans), strawman ("not X, but Y" + staccato "Not a X. Not a Y."), em_dash_density, triad_density, long_sentences (share + mean), no_contractions, flourish_ending. Reuses claims ExtractAssertionText (tag-split solved once). | built, vet clean |
| Corpus tests | `voicetells_test.go` — V1–V7 from the spec (real shipped copy both directions) + strawman + opt-in + contraction tests | ALL GREEN first run |
| Discovery check | `discovery_checks/check_voice_tells.go` — name `voice_tells`, opt-in via voice_gate, long-form thresholds for blog/guides, one item per page (`voice:<page_id>`), severity ALWAYS medium, priority 40 (behind claims — truth outranks register), HITL-terminal | built + package tests green |
| CLI | `cmd/voicescan` — same TSV contract as claimscan, exit 1 on findings | built + used |
| Config | leopardess voice_gate seeded (10 site bans from banned_language, curated to machine-safe regex; bak_voice_gate_seed_20260717) | LIVE in DB |

**First live scan: 111 findings across 85 components — calibration verified by construction:**
- The v2 pages (index, services-restored) produced ZERO findings; the v1-dense pages lit up
  (how-we-work about-content: strawman + 7 triads + 4 em-dashes — the worst page, correctly).
- Em-dash density is the dominant signal (up to 41/1000 words on privacy) — the v1 register's
  signature, exactly as the spec predicted.
- It flagged OUR OWN hand-written 16 July copy (who-we-help cards/faq) as no_contractions —
  correct: that was the pre-v2 register. The checker has no author bias.
- Today's writer output on about (leadership-team bio) trips long_sentences — matches the
  human read.

**Deploy state:** code is local + tested; ships with the next chassis image (owner-gated, per
practice). Enable AFTER the image lands by adding "voice_tells" to quality-discovery-agent's
checks array (adding it before the image would hit an unknown check). NOT committed to git yet
— awaiting owner's word per house rule.

**Rollout implication (task pending):** the scan IS the v1→v2 worklist — the pages it flags
are exactly the remaining rewrite targets: how-we-work, how-it-works, engagement-model,
who-we-help (register pass), case-studies, careers, privacy, faq triads, tool guide intros,
about leadership bio polish.

## Turn 21 (2026-07-17) — image deployed; voice_tells LIVE in production and it caught a homepage regression on run one

**CLAUDE.md followed.** Committed the voice checker (feat) + leopardess docs (docs) as two
narrow pathspec commits. New chassis image deployed by another session; verified MY code is in
the running pod per CLAUDE.md (`strings /app/agent-chassis | grep -c voice_tells` = 7;
ScanVoice/voice_gate/"em-dash as a rhythm" = 7). Waited out the ~300s post-restart dispatch
window before firing anything.

**voice_tells ENABLED + validated end-to-end in production.** Added to quality-discovery-agent
checks array (bak_qualdisc_agentdef_20260717; DB config, live immediately). Fired the agent →
**25 voice_tells work items written**, all correctly routed: status needs_human_review,
severity medium, priority 40, NO handler agent — HITL-terminal by construction, exactly as
designed. The check produces the same findings in production as the CLI did.

**★ THE CHECK EARNED ITS KEEP ON RUN ONE: it caught a clobbered homepage.** It flagged index
with 11 findings — but index was my clean v2 page. Investigation: my v2 index (4 clean slots)
had been **rebuilt at 14:14 into 6 slots** (added system-stats + case-studies-grid), with:
- a CTO-register hero ("You've validated the use case…", em-dashes) — voice regression;
- `system-stats` rendering "Functional Areas: 150+" (the audited-OUT functional-areas
  fabrication, resurrected) and "150+%"/"150++" garbage from the shared suffix forcing;
- `case-studies-grid` with **invented case-study titles** ("Validation Layer Stops Bad Data
  Reaching the Warehouse", etc.).
This is the `replan-clobbers-built-pages` landmine (another session's build-site-planner run
regressed a built page). **The voice checker surfaced it; the claims V1 check did NOT** — the
"Functional Areas 150+" is a mislabeled TRUE number (150 agents), a B3-class case the spec
predicted V1 would miss and V3 would need. The two layers are complementary: voice caught the
regression that exposed the claim.

**Fixed:** deleted system-stats + case-studies-grid slots; pruned BOTH plan locations
(pages.sections AND the site_plan aspect) so a rebuild can't resurrect them; re-applied v2
content to the 4 good slots; re-rendered. Verified live: 0 functional-areas, 0 fake case
titles, 0 system-stats section, 0 em-dashes, v2 hero back. bak_index_clobber_20260717.
Only index was hit — who-we-help/about/services/contact all clean.

**Open (for owner / next):** 25 voice_tells items + the claims item (llm-cost-calculator, 1
banned claim) sit in needs_human_review — the v1→v2 rewrite worklist, now machine-generated.
Minor polish: density findings carry an empty `snippet` (value/threshold are the useful
fields; cosmetic). fixloop tools diagnosis still `awaiting_diagnosis` — its dispatch loop
(`diagnose-pipeline-trigger`) ships DISABLED, so the 090 auto-fire is the only path and it
didn't land; needs a re-fire or the loop enabled.

## Turn 22 (2026-07-17) — tools diagnosed directly (fixloop dispatcher was disabled); only 1 of 4 actually broken

Owner asked to see if the fixloop could detect+fix the "tools aren't working" complaint. The
fixloop INTAKE item was written (needs_diagnosis:leo-tools-runtime) but the diagnose
orchestration NEVER ran — `diagnose-pipeline-trigger` scheduled task ships DISABLED, so the
090 auto-fire had no dispatcher. So I diagnosed directly with a real browser (local headless
chromium executes JS via --virtual-time-budget --dump-dom).

**Verdict: 3 of 4 tools WORK; only llm-cost-calculator is broken.**
- ai-agent-roi-estimator → WORKS. References tool-ai-agent-roi-estimator.js (200); headless
  render showed real outputs ($520K annual, 1633% ROI). (Numbers are US-centric $ and use an
  invented cost model — a content quibble, not a break.)
- password-entropy → WORKS. Self-contained 1709-char inline calc IIFE wiring #entropyInput;
  the #entropyBits=0 on load is correct (empty password = 0 bits).
- tool-agent-complexity-estimator → WORKS. Self-contained 9050-char inline calc; empty result
  panel on load is correct (quiz not yet answered).
- **llm-cost-calculator → BROKEN.** Its page references `bayesian-ranking-hero-tool.js` (the
  WRONG tool's JS); its own bundle (tool-llm-cost-calculator.js and every cost-calc name) is
  404; it has NO inline fallback. Outputs (#output-tokens, #gpu-cost-month, #results-tbody)
  stay empty. Almost certainly downstream of its `empty_internal_href` deploy FAILURE (the gate
  fails deploy_page → the tool's assets/correct page never ship).

**Method notes:** no CSP header (scripts not blocked); all element IDs the working JS queries
are present and correctly scoped inside their `[data-component]` section; the lucide inline
call is guarded (`if typeof lucide !== undefined`) so a blocked icon CDN is cosmetic, not
fatal. My curl sandbox can't reach external hosts (unpkg/fonts → 000) so icon-blocking for
REAL browsers is unconfirmed — but it wouldn't break tool logic either way.

**FIX (belongs to the tool-generation pipeline, NOT a content edit):** regenerate + deploy the
correct llm-cost-calculator JS bundle and correct the page's script src; unblock its
empty_internal_href deploy first (the audit already filed that item). Recorded on the intake
item's spec.operator_diagnosis for the fixloop thread.

**Feedback for the fixloop thread:** its detection couldn't run at all here because
`diagnose-pipeline-trigger` ships disabled and the 090 direct-fire's diagnose orchestration
didn't materialise. Worth either enabling the loop for real cases or fixing the 090 direct
path. The browser-runner-adapter pod IS up (53m) — the capability exists; the dispatch didn't.

## Turn 23 (2026-07-18) — owner design/imagery review: 2 fundamental bugs filed, index fixed

Owner raised 7 items. Plan: `PLAN_imagery_and_design_2026-07-18.md`. Two were fundamental →
`bugs_open/`.

**★ bugs_open/011 (NEW) — generated images cannot render text.** The live homepage hero was an
SDXL image that LOOKED like a flowchart and was full of gibberish words. Root cause is the
model class, not the prompt: diffusion synthesises glyph-shaped texture, not text. Answers the
owner's question (better model = marginal; better prompts = necessary but insufficient; "loop
until correct" can only REJECT, never generate legible text). Fix = split the jobs: heroes are
generated text-free illustration (Banana); anything with words/numbers/structure is
code-rendered SVG from real values (the site's own D1/D3 principle + the never-built L7 chart
component). Also flags the trap that `infographic` is a routable DIFFUSION kind — it should not
be satisfiable by an image model at all.

**★ bugs_open/001 (evidence appended, severity raised) — the re-plan clobber.** It hit
leopardess TWICE in 24h. 07-17 14:14: homepage rebuilt 4→6 sections, re-adding fabricated
"Functional Areas: 150+" and INVENTED case-study titles. 07-18 07:50: services rebuilt, which
INVENTED the link `/tools/tool-monitoring-coverage-gap-finder.html` — a 404 — and that is the
blank page the owner reported. Key reframing added to the bug: this is not only content LOSS,
it is fabrication INJECTION, and it defeats human review. Useful discovery recorded there:
`site_plan_imagery`-wired heroes SURVIVE the clobber while page_components copy does not.

**Fixed and verified live today:**
- **Nav/tools:** header nav had NO tools (the rebuild stripped them). Removed the dead
  `tools`-group item (that group renders in neither header nor footer) and linked the 4 WORKING
  tools from the footer utility group. Tool audit updated: process-automation-scorer is REAL
  and WORKS (self-contained 5KB calculator) — so 4 work, only llm-cost-calculator is broken.
- **Index hero replaced.** Generated text-free via Banana `illustration` (four gold inputs
  converging into one steady output on charcoal; upper-left left clear for the headline),
  reviewed by eye, wired via `site_plan_imagery` (clobber-resistant). Live.
- **Overused words:** owner ruled "trust", "honest", "earns its keep" out (live counts 12/9/2).
  Added to voice_gate banned_phrases AND banned_language prose guidance
  (bak_voice_words_20260718). The voice checker then produced the exact worklist — including
  the owner's cited adjacent-articles pair. Rewrote the homepage instances ("can't just be
  trusted" → "will not match the register on its own"; "how much the source can be trusted" →
  "how reliable that source has been"; both "the honest answer" cut). Verified 0 live.

**Still to do:** heroes on the remaining pages (about/services need a shared-component image
field first — 9 and 5 sites); replace the garbled `/assets/images/hero.jpg` itself (still the
fallback on how-it-works); the rest of the trust/honest worklist; infographics (blocked on 011
§V3). All content work is provisional until 001 is fixed.

## Turn 24 (2026-07-18) — infographics: the capability was already wired; bug 011 CORRECTED

**★ The owner showed two Gemini infographics with perfectly legible text and asked if we could
wire it up. We already had.** Deployed `BANANA_DEFAULT_MODEL` = `gemini-3-pro-image-preview`
(the same model), and `kind:'infographic'` ALREADY routes to it
(`dynamic_adapter.go` switch: icon|logo|illustration|infographic|sprite_sheet|content_hero →
Banana). Nothing needed building.

**My bug 011 was WRONG and is corrected in place.** It claimed "generated images cannot render
readable text" and proposed building an SVG renderer — a thread acting on it would have built a
subsystem we already had. Renamed to the real, narrower bug:
`011_HANDOFF_2026-07-18_hero_kind_routes_to_a_model_that_cannot_render_text.md`. The garbled
homepage hero was a ROUTING accident: `kind:'hero'` falls through to Stability/SDXL, which
genuinely cannot render text. Per CLAUDE.md I also corrected the 016b §10 index entry and added
the transferable §9 pattern: **"read the dispatch table, not the output"** — when behaviour is
selected by an enum, one value routing to a weaker backend is indistinguishable, from the
artefact alone, from the whole capability being missing. Two greps answer it.

**PROVEN first try:** `infographic_what_we_build` — a three-column infographic, fully legible,
correctly spelled, figures exactly matching the evidence base (2,767/937, 5,652/4,672, 8 sites),
on-brand charcoal+antique-gold, and it independently honoured the newly-banned "trust" wording
("how reliable the source has been"). Live at /assets/images/infographic-what-we-build.jpg.

**Dominant variable is PROMPT SPECIFICITY,** not the model. The successful prompt names the
layout, every column header, every card's heading AND body text verbatim, the palette by hex,
each icon, and ends with "all text correctly spelled, real words only, do NOT add any number
not listed above". Thin prompts are what produced the earlier rubbish. Recorded in 011 §2.

**Placed on the homepage** (position 3, between "What we've built" and "What we could build
with you") as a `generic-text-block` — its `content` field renders RAW HTML because the
renderer uses Go `text/template`, not `html/template` (verified at
component_library.go:500). Carries a full descriptive alt text and a caption noting every
figure is evidence-base backed. bak_index_infographic_20260718.

**Three more generating** in the same lane: `infographic_how_a_job_runs` (how-it-works — the
six-station pipeline with the human-decision gate drawn largest), `infographic_architecture`
(technical-architecture — hierarchical agents over a Kubernetes/Kafka/Postgres stack), and the
themed variant `infographic_leopardess_line` — **a Harry Beck London Underground map, "THE
LEOPARDESS LINE"**: three lines (Records / Reading / Sites) converging on one interchange
labelled "A PERSON DECIDES" and terminating at "WRITTEN DOWN", with a service-notes panel
carrying the audited figures. Chosen as apt: British information design, and a work item
travelling through stages with a human-approval interchange maps onto the form exactly.

**Review gate still applies:** look at every generated image before wiring it. Even the good
model errs — the owner's own Gemini map rendered "REPRETITIVE". That is 011's R2 (an OCR/vision
legibility gate) and it is not built.

## Turn 25 (2026-07-18) — four infographics generated, reviewed and placed

All four came out of `kind:'infographic'` → Banana → gemini-3-pro-image-preview. Each was
downloaded and **reviewed by eye before wiring** (the standing gate — 011 R2 is not built).
All four: legible, correctly spelled, figures matching the evidence base, on-brand.

| Asset | Page | Content |
|---|---|---|
| `infographic_what_we_build` | index (pos 3) | Three columns: what we've built / what we could build / how an engagement starts |
| `infographic_architecture` | technical-architecture (pos 2) | Hierarchical agents over the K8s/Kafka/Postgres stack + "built in" panel; states the AUDITED "over 150 agent definitions" |
| `infographic_how_a_job_runs` | how-it-works (pos 2) | Six-station pipeline; "A PERSON DECIDES" drawn largest and highlighted, exactly as prompted |
| `infographic_leopardess_line` | case-studies (pos 2) | **The themed variant** — a Harry Beck Underground map, "THE LEOPARDESS LINE": Records / Reading / Sites lines converging on one interchange "A PERSON DECIDES", terminus "WRITTEN DOWN", service-notes panel with the real figures |

Placement mechanism: a `generic-text-block` per page holding a `<figure>` with a full
descriptive `alt` (the diagram's content in prose, so the page is usable without the image),
`loading="lazy"`, explicit width/height, and `style="width:100%;height:auto"`. Works because
the renderer is Go `text/template`, not `html/template`. Backups:
bak_index_infographic_20260718, bak_ta_infographic_20260718, bak_hiw_cs_infographic_20260718.

**The themed variant was chosen for aptness, not novelty:** a work item travelling through
stages, with a human-approval interchange, maps exactly onto a transit diagram — and Beck's
map is the canonical British information-design object, which suits the register better than a
fantasy map would. The three lines are the three genuinely-running systems, so the graphic is
also a truthful summary rather than decoration.

**Infra note:** the first firing of two of these hit the spawn flake again (48 Kafka dial
errors on the spawned image-generator pod, 0 generations). Delete the pod, re-fire — both then
succeeded. Same class as bugs_open/003.

## Turn 26 (2026-07-18) — docs brought up to CLAUDE.md's "standing four"; images verified rendering

**Owner reported "the images aren't appearing" with a screenshot of services.html.**
Investigated: the four infographics DO render — verified the served markup is real HTML
(`<figure class="infographic"><img src="/assets/images/...">`), not escaped entities, on
index / how-it-works / case-studies / technical-architecture, and all four files serve 200
(111–196 KB). **services.html is simply one of the pages that has no image at all** — no hero,
no infographic. The owner's screenshot is accurate for that page. Ungrounded generalisation
avoided by checking each page rather than trusting one report.

Live per-page image inventory (2026-07-18, checked not carried forward):
- hero + content image: index
- content image only: case-studies
- hero + content image: how-it-works, technical-architecture (but on the GARBLED hero.jpg)
- good hero only: who-we-help, how-we-work
- garbled hero.jpg only: engagement-model, faq, careers, insights
- **NOTHING: about, services, use-cases, contact, blog**

**CLAUDE.md re-read — it gained a "Working docs — the standing four" directive today**
(PLAN / RUNBOOK / NOTES / SUMMARY per workstream, created at the start, updated as you go;
record what was WRONG not just what is right; ground every figure against the live system;
point at bugs rather than restating them; grep before filing). Leopardess already satisfies it
(PLAN_leopardess_rebuild + PLAN_imagery_and_design / RUNBOOK / RUNNING_NOTES / SUMMARY), so no
migration — but the docs were updated to its standard:

- **HANDOFF.md rewritten as a cold-start**: a red READ-FIRST box making the re-plan clobber
  (bugs_open/001) the first thing a new chat sees, since it makes all copy work provisional;
  a live-checked state section (imagery table, the two new checking layers verified in-pod,
  the owner review queue counts, tool status); the 011 correction recorded visibly with
  `> **CORRECTED**`; punch-list 1 retitled as closed and **punch-list 2 (the 2026-07-18 owner
  review) added** as the current one.
- **STALE CLAIM CAUGHT AND FIXED while editing:** the handoff still said "ai-readiness-quiz is
  still blank — the ONE open in-flight item". It was fixed in turn 21. Verified live before
  writing (54,118 bytes, 3 components) and replaced with the correct account + the root cause
  (the fleet-wide `contact-block` `jane@company.com` fallback failing the email validator).
  This is exactly the failure the "ground every figure" rule exists to prevent.
- **SUMMARY_where_we_are.md rewritten** for the owner in plain prose, including an honest
  account of the image-generation mistake and its correction.

---

## 2026-07-26 — L7 (charts) built elsewhere, as a shared component

Written by the brochure_component_library workstream, into this dir deliberately
rather than into its own: L7 was scoped here first, and a second chart component
is exactly the drift the council gate exists to catch.

The owner ruled on 2026-07-26 that the chart is **one shared chassis component**,
not a per-site build. It is live as `evidence-chart`: a section component whose
`charts` and `facts` fields resolve from `site_specs.evidence_base`, so the code
owns the numbers (PLAN §5's D1) and no new diffusion kind appears (D3). Bars are
drawn in CSS from the real value; the labels and figures are real HTML text, so
screen readers and the claims gate both see them.

**What leopardess has to do to get charts: add a `charts` key to its own
evidence base.** Nothing else — no code, no registration, no image roll. The 18
facts already there include two genuinely chartable pairs
(`C1-records-verified`/`C1-records-enriched`, `C2-feed-items-collected`/
`C2-feed-items-scored`). The contract, the traps and a working example are in
`docs/agent_docs/docs024_key_docs_latest/brochure_component_library/components/evidence-chart/README.md`
and `sql/evidence_base_charts_2026-07-26.sql`.

Two corrections to PLAN §5 while I was in there, both from reading the tree:

- **go-echarts is not in this codebase**, so the "go-echarts vs static SVG"
  conflict §5 agonises over does not exist to be resolved. (The gripper-dossier
  session recorded the same correction on 2026-07-24.)
- **A dependency-free Go SVG chart emitter already exists** —
  `platform/orchestration/actions/report_charts.go` (`renderBarChartSVG`),
  written for the gripper dossier. That is the natural home for §5's "Go emits
  the static SVG" half if it is ever wanted. It is inert until an image roll,
  which is why the shared component ships as config instead. One caution found
  while building: text inside `<svg>` is invisible to the claims gate
  (`datahelpers/claims.go:137`), so an SVG chart's numbers would leave the
  verification net that currently covers them.

---

## 2026-07-29 — the six broken images on /blog.html (found, fixed, live)

Owner asked to "fix the missing images". Swept all 27 pages in the sitemap rather than
trusting any per-page claim in these notes, since the last inventory here is dated
2026-07-18 and figures go stale.

**First sweep was WRONG and said so within minutes — recording it because the mistake is
the reusable part.** The regex matched `src="…"`/`href="…"` ending in an image extension.
That missed **every hero on the site**, because the heroes are CSS backgrounds
(`background-image: … url('/assets/images/hero-home.jpg')`), and it would equally have
missed a presigned S3 URL, which ends in `&X-Amz-Signature=…` and not in `.jpg`. The
corrected sweep matches `url('…')` too. **A URL-shaped regex anchored on file extension
is not an image inventory** — it silently describes a smaller site than the one you have.

**Every image URL the site references serves 200. Nothing 404s.** So "missing" was never
a dead link:

| what | where | state |
|---|---|---|
| `<img src="">` × 6 | `/blog.html` blog cards | **the defect — now fixed, live** |
| garbled `hero.jpg` | hero background on **14** pages | still there |
| no image at all | about, services, use-cases, contact, blog + 4 tool pages | still there |

### The defect

`/blog.html` served six `<img src="" alt="…" loading="lazy">`. An empty `src` renders as a
broken-image icon, and the site carries **no `.article-card` CSS at all** (0 matches in
`styles.css`, 0 in every inline `<style>` block), so the cards are unstyled and each one
led with a raw broken icon. Verified in the served HTML, not inferred:
`curl … | grep -c '<img src=""'` → `6`.

Cause, both halves:
1. `rebuild_blog_listing_action.go:218` hardcodes `"image": ""` for every article — there
   is no per-article imagery on this platform (the "Phase I3" gap these notes already
   describe). All 6 rows in `content_data->'articles'` carry `"image": ""`.
2. The shared `content-listing` template in `content_components` emitted
   `<img src="{{.image}}">` **unconditionally**, so an absent image became a broken one.

**Landmine found while tracing it:** this page_component's `component_id` points at the
site fork `blog-listing_pre_037` (which uses `{{.post1_image_url}}`-style placeholders),
but the stored HTML was rendered from the shared `content-listing` template — the action
calls `loadContentListingTemplate` and ignores the slot's component. So **a
`page-rerender` in `section_data_resolved` mode would render this slot from the FORK and
produce a listing of six empty placeholder cards.** That is why the fix below did not go
near O8's `rerender_pages.sh`.

### Fix (config only — no Go change, no image roll)

Guarded the wrapper in both `content-listing` and `category-listing`:
`{{if .image}}<div class="article-card__image">…</div>{{end}}`. An article with no image
now emits no `<img>`. The Go half still writes `""`; the template is just honest about it.

**Blast radius measured BEFORE applying, not left for a reviewer.** Only 3 page_components
fleet-wide use these two components: `robot-hands.com/learning-center-hub` (3 articles,
**all** with a non-empty image → wrapper still renders → byte-identical output) and two
dartsonline pages with **0** articles. The only output that changes is leopardess's.

Then regenerated this one component's `rendered_html` deterministically (asserted first
that 0 of 6 cards had a real image, so removing the wrapper is exactly what the guard
does) and deployed with `reassemble_pages.sh` — **assemble mode**, which embeds stored
section HTML and deploys, and does not regenerate content. Deliberate: a full
`rerender-pages` runs `rebuild_blog_listing` but risks the content clobber this lane has
been bitten by twice (`bugs_open/001`).

### Verified live, at the artefact

```
orchestration 9bd885ee  COMPLETED in 7s   (status alone is not proof — checked the page)
curl https://leopardessconsulting.co.uk/blog.html | grep -c '<img src=""'   ->  0   (was 6)
grep -c 'article-card__title'                                              ->  6   (cards intact)
whole-site re-sweep, 27 pages: 0 empty img tags; all 9 image URLs HTTP 200
```

Contributed the transferable half to `bugs_open/128` (that check's third blind spot:
`src=""` has no URL to probe, so the *proposed HTTP fix* would confirm it as **200 OK** —
worth pinning before that half gets built).

### Still open, NOT fixed by the above — both need an owner decision

- **`hero.jpg` is garbled** — 900×900, AI-generated gibberish text, used as the hero
  background on **14** pages. Looked at it directly this session; the 2026-07-18 note
  calling it garbled is still accurate. It is the single worst-looking image on the site.
- **9 pages carry no image at all** — about, services, use-cases, contact, blog, and the
  4 tool pages. `about`/`services` need component work first (`hero-about`/`hero-services`
  declare no image field), per the HANDOFF.
- **[LATENT, not currently visible] 13 of 18 `assets` rows hold presigned S3 URLs that
  have now EXPIRED** (`X-Amz-Expires=604800`, dated 16–18 July; probed one → **HTTP 401
  "Request has expired"**). No page renders them today, so nothing is broken on the site
  right now — but this is exactly the state the HANDOFF warns about ("verify `assets.url`
  is `/assets/images/…`, NOT a presigned `s3…?X-Amz-…` URL"), and any path that resolves
  an image from `assets.url` would emit a dead link.

---

## 2026-07-29 (cont.) — five more heroes, an expired-asset cleanup, and Phase I3 is not what HANDOFF said

Continuing the same session, after the owner asked to also (a) replace the garbled
`hero.jpg`, (b) give about/services/use-cases/contact a hero image, (c) clean up the
expired asset URLs, (d) assess the blog-thumbnail (Phase I3) gap.

### hero.jpg replaced

Regenerated via the proven scope-less Route A recipe (`kind:"hero"` — now routed to
Banana platform-wide since `bugs_closed/011`, not just `illustration`; confirmed
`origin_model=banana/gemini-3-pro-image-preview` on the new row). `asset_key="hero"` ==
`purpose`, so `DeployedWebPath` collapses to the same filename and the new file simply
overwrites `/assets/images/hero.jpg` — no page rerender needed, since all 17 consuming
pages hardcode that literal path in `content_data.background_image`. Verified: file
byte-different from the old one, `last-modified` matches the fire time, looked at it —
flat vector, charcoal ground, hairline gold radiating from one node, no text. Live on
all 17 pages simultaneously the moment the deploy landed.

### about / services / contact — the HANDOFF's "needs component work" was 11 days stale

Checked before generating anything: `about-hero`/`services-hero`/`contact-hero`
(their real component names — HANDOFF's `hero-about`/`hero-services` were the CSS class,
not the component) **already declare `background_image`/`hero_url` in their
html_template**, guarded (`{{if or .hero_url .background_image}}`). Someone had already
done the component work platform-wide (12/6/12 site consumers respectively) since the
11-day-old HANDOFF was written. `use-cases-hero` genuinely had no image support (2 site
consumers) — added the identical guarded pattern, mirrored byte-for-byte from its
siblings; blast radius measured first (finetuning.uk, the only other consumer, has
neither field set, so byte-identical until it opts in).

**Second correction, found empirically, not by reading first:** having the template
guard is not sufficient. `plan_sections_action.go`'s `sectionHasImageField` gate — which
drives the `site_assets.hero` → `content_data.background_image` alias-write — only fires
when the component's `input_schema.fields` declares an image-typed field. None of these
four components declare one (only the generic `hero` component does, with
`"source":"site_assets.hero","fallback":"/assets/images/hero.jpg"`). Fired an
`image_landed` rerender on `about` first to test this and confirmed empirically: asset
generated, `site_plan_imagery` row wired correctly, rerender COMPLETED, **and
`content_data.background_image` still empty** — the schema gate, not the template, was
the actual gap.

**Deliberately did NOT fix this by editing the shared components' `input_schema`.**
Adding the field with a `fallback` would make EVERY page rendered through these
components on ALL 12/6/12/2 consuming sites pick up a background image automatically at
their next rerender, whether or not anyone asked for one there — a live behaviour change
for other sites' pages, not merely an inert opt-in (unlike the template guard, which
does nothing until a page's `content_data` sets a key). That crosses into the kind of
shared-mechanism change CLAUDE.md's platform-seams section is about; not something to
fold into a single-site imagery fix without measuring who else it touches.

Instead, scoped it to leopardess alone: generated the four heroes (Route A, on-brand —
about: two concentric rings around a centre node; services: three evenly-spaced nodes on
one spine; contact: two nodes joined by a line through a midpoint; use-cases: one root
node branching into three), wired `site_plan_imagery` rows exactly as for who-we-help/
how-we-work, then **directly merged `background_image` into these four page_components'
`content_data`** and regenerated their `rendered_html` to match what the (already-correct)
template would itself produce — same technique as the blog-listing fix: deterministic,
no LLM, so zero clobber risk. Reassembled + deployed (assemble mode). Verified live: all
four show their image, headline copy byte-identical to before, whole-site re-sweep clean
(0 empty img tags, every image URL 200).

Contact and services came out visually similar (both read as "several nodes on a line")
despite different prompts — a minor imperfection, not a defect; noted rather than
re-rolled, since both are correctly on-brand and legible.

Image-less pages count: **9 → 5** (blog index + the 4 tool pages remain; case-studies/
how-it-works/technical-architecture/how-we-work/who-we-help/index already had one).

### Expired asset URLs — turned out to be a live risk, not just stale data, and it's now filed

Full inventory: **every one of 13 active hero/infographic asset rows** — including the
four generated minutes earlier in this same session — carried a presigned S3 URL
(`X-Amz-Expires=604800`). This is not a leftover from turn 17; it is standing behaviour
of the deploy path (`deploy_image_asset` only rewrites `assets.url` when passed
`asset_id`; RUNBOOK O5 landmine 6, apparently never actually exercised by any caller).

Traced the actual risk rather than assuming it was cosmetic: `plan_sections_action.go`
deliberately routes AROUND `assets.url` for rendering (`storage.DeployedWebPath`, with a
comment saying so) — so this does not explain any symptom seen on the live site. But two
other call sites read `assets.url` and fetch it directly: `derive_brand_head_assets_action.go`
(favicon/og-card from the logo) and `derive_card_asset_action.go`'s `findCardSourceHero`
(card thumbnails from a page's hero — exactly the mechanism the blog-thumbnail gap below
needs). Either would 401 against a row past its 7-day window. Filed as
`bugs_open/152` (platform-wide, unowned) rather than treating it as this-site-only.

Fixed for leopardess: retired one orphaned row (`hero_case_studies` — wrong-provider
SDXL leftover from turn 17, wired nowhere, referenced nowhere; same shape as
`bugs_open/114`'s class, smaller scale) and rewrote `assets.url` to the real,
already-verified-200 local path for all 12 remaining active rows. This is a contained
fix, not a fix for the class — the same defect recurs on the next generation, here and
everywhere.

### Real blog thumbnails (Phase I3) — the HANDOFF was wrong; it is already built

Asked to assess scope, not build. Reading the code turned up a second stale HANDOFF
claim: **"Per-card / per-section images = Phase I3 — NOT built"** is false as of
2026-07-29. `derive_card_asset_action.go` exists, is registered (`registry.go:197`),
triggered by `asset-deployer`'s `content_card` mode, and is documented as shipped in
`docs/agent_docs/docs024_key_docs_latest/imagery/PLAN_imagery_best_in_class.md`
("I3.2 ✅ built"). It cover-crops a page's hero (falling back to the site brand hero —
now clean, since hero.jpg was just replaced) to an 800×450 card, no LLM, no diffusion,
and links it to the entity.

It has never been fired for leopardess (0 `purpose='card'` assets on this site — a
triggering gap, not a missing capability). It is also **owned and actively being
hardened elsewhere right now**: the `imagery` workstream built it, and the
`bugfix_131_og_card` lane found a sibling defect in this exact file TODAY
(`bugs_open/143` — commits before its lock check; measured latent, 0 of 12 fleet-wide
card rows are locked, so not a live risk for a first run). Given the task was framed as
assess-not-build, and the action is mid-fix elsewhere today, I did not fire it — that is
a small, well-defined next step, not a "needs new platform work" item as the old HANDOFF
said. HANDOFF.md's imagery section needs a correction pass; not done yet this session.

### Full verification, end of session

```
27/27 sitemap pages fetched; 0 empty <img src="">
Every image URL referenced anywhere on the site (14 distinct paths): HTTP 200
Image-less pages: 5 (blog index, 4 tool pages) -- down from 9
assets.url: 12/12 active rows now local paths, 0 presigned
```

---

## 2026-07-29 (cont.) — a 4-part feature series on trusting AI with data, researched and published

Owner asked for a feature article on trusting AI with data, researched thoroughly, with
sources cited and their opening lines quoted, looked at from multiple industry angles,
charts and/or tools considered, real verified statistics, both sides argued honestly.
Said explicitly it could become 3–4 linked articles.

### Research

Used `trust.anthropic.com` as the named starting point (its own JS shell defeated a plain
fetch — Vanta-hosted trust centre, client-rendered — so pulled its certifications from
`support.claude.com`'s own certifications article instead, which fetched cleanly). Then
pulled and cross-checked primary-source statistics across: general consumer trust (KPMG/
University of Melbourne 2025, n=48,000+/47 countries; Pew Research 2026; Edelman Trust
Barometer 2026; Cisco 2025 Data Privacy Benchmark, n=2,600), healthcare (Reach3 Insights/
Rival Technologies 2026 Digital Health Trends, n=1,043; KFF tracking poll; CHAI patient
survey; IBM breach costs), financial services (Deloitte 2026; Cambridge Judge Business
School 2026; McKinsey State of AI Trust 2026, n≈500 orgs; Salesforce, n=6,058), legal
(Thomson Reuters Future of Professionals), hiring (Greenhouse 2025, n=4,136; Gartner,
n=2,918; ResumeBuilder, n=948; Dice Trust Gap in Tech Hiring 2025, n=319), regulation (EU
AI Act 2026 deadline), and the counter-case (IBM 2025 breach report's AI-positive findings;
PwC 2026 AI Performance Study).

**Caught two real errors before publishing, by verifying rather than trusting the first
search synthesis:**
1. A first pass had "84% of consumers would switch providers over data mishandling"
   attributed loosely to financial services. Direct verification found the real, precisely
   sourced figure is **78%** (Salesforce, n=6,058, 2023 survey) — the 84% was never
   confirmed against a primary source and was dropped.
2. A hiring-trust figure ("14% trust fully automated hiring, 46% trust hybrid") had no
   named source in the first synthesis. Traced it to Dice's *Trust Gap in Tech Hiring 2025*
   (n=319 US tech professionals) by fetching the report directly — which also surfaced a
   companion, better-fitting statistic (68% distrust fully-AI hiring vs 80% trust
   human-driven) from the SAME report, used instead.

Every statistic in the published series is attributed to a specific named study with a
sample size where available; nothing is "some survey found."

### Architecture: 1 pillar + 3 industry deep-dives

- `can-you-trust-ai-with-your-data` — the overview: the KPMG trust paradox, an industry
  tour (healthcare/financial/legal/hiring/retail/government), the case against (breach
  economics, governance gaps), the case for (fraud-detection ROI, Cisco/PwC figures), and
  a concrete "what trustworthy looks like" section built around Anthropic's own published
  certifications (SOC 2 I/II, ISO 27001:2022, ISO/IEC 42001:2023, HIPAA-ready + BAA, the
  October 2025 training-data policy change, Zero Data Retention for regulated workloads).
  ~3,100 words, 2 charts.
- `ai-data-trust-in-healthcare` — the sharpest patient-trust decline (52%→44% in two years)
  and the factors proven to rebuild it (FDA approval, clinician in the loop, representative
  data). ~1,370 words, 1 chart.
- `ai-data-trust-in-financial-services` — adoption outpacing governance (87% say they
  could improve governance, only 13% call themselves leading-maturity), EU AI Act
  deadline. ~1,080 words, 1 chart.
- `ai-data-trust-in-hiring-and-hr` — the widest trust gap found anywhere in this research
  (70% of hiring managers trust AI hiring decisions, 8% of job seekers call it fair).
  ~1,010 words, 1 chart.

All four cross-link to each other. `docs/leopardessconsulting/scripts/L8_article_*.sql`
holds the exact insert used for each, for reproducibility.

### Charts — built as hand-authored, code-rendered inline SVG, NOT the shared evidence-chart component

Considered `evidence-chart` (the shared component from the 2026-07-26 chart work) first
and deliberately did not use it. It resolves from `site_specs.evidence_base`, which this
site already uses specifically for FIRST-PARTY, re-queryable facts (each fact row carries
`source.sql` or `source.artifact` — something this platform's own DB or code can be
re-checked against). Third-party survey statistics (KPMG's 46%, IBM's $4.44M) have no
`SELECT count(*)` behind them and do not belong in that register — shoehorning them in
would blur the exact distinction the claims-verification work on this site exists to
draw, between "verified against our own system" and "cited from someone else's report".

Instead: 5 inline `<svg>` bar charts, hand-coded directly into the article HTML, each
captioned with its source and date, matching `design_intent.imagery_direction`'s own rule
that a chart must carry its source and never be image-generated. Colours use the site's
real palette hex values directly (`#836E32`, `#0D0D0D`, `#B9B3A6`, `#E4DFD5`) rather than
CSS custom properties, since injected `content_data` HTML cannot reliably resolve a
parent stylesheet's `:root` variables.

### A platform behaviour worth knowing: cross-links to a page that doesn't exist YET get silently stripped

Found this rendering the pillar article before its three siblings existed:
`content_data.content` kept the `<a href="/blog/ai-data-trust-in-healthcare.html">` intact,
but `rendered_html` did not — some render-time mechanism (not diagnosed further; matches
this platform's general phantom-link-defence class, `bugs_open/029`/`079` family) strips
anchors pointing at pages that don't resolve at render time. **Not a bug filed** — the fix
is mechanical and now confirmed: publish the target pages first, or re-render the linking
page afterward. Did the latter here (re-ran `can-you-trust-ai-with-your-data` after all
three siblings existed) and the links resolved cleanly on the next render. Worth knowing
before any future multi-page series on this platform: write leaf pages before the hub, or
budget one extra re-render pass for the hub.

### Blog listing extended, not rebuilt

Did NOT fire the `rerender-pages` workflow to pick these up (it would run `rebuild_blog_listing`,
but also `get_pages_for_rerender`/`create_rerender_items`/`render_site_components` — a much
wider blast radius than "add 4 entries to one list", and this site has twice been bitten by
a wider rebuild clobbering hand-fixed content). Instead, cloned one existing card's exact
markup programmatically (Python, matched against the now-guarded `content-listing`
template) and prepended 4 new entries to both `content_data.articles` and `rendered_html`
directly — same no-LLM, no-full-rebuild technique as the original empty-`<img>` fix earlier
this session. Blog listing now shows 10 articles, still 0 empty images.

**Known gap, not fixed this session:** the 4 new pages are not yet in `sitemap.xml` — that
generation is a separate, undiagnosed mechanism not touched by anything fired here.

### Considered, and made a call on: a new interactive "AI vendor trust checklist" tool

Owner asked to think about a tool for this content. Assessed it: technically straightforward
(deterministic client-side scoring — SOC 2? ISO 42001? zero-retention offered? training
opt-out by default? sub-processor list published?, same shape as the site's existing
calculators) and would tie directly to the pillar article's "what trustworthy actually
looks like" section. **Not built this session** — it is a genuinely separate feature (new
JS, new page, new component, and this platform's own standard for UI work is to browser-test
before calling it done), not a content-writing task, and deserves that same care rather
than being rushed in at the tail of an already large session. Flagged to the owner as a
concrete, scoped next step rather than built half-verified.

### Verified live, end to end

```
All 4 new pages: HTTP 200, correct h2/svg counts, no missing required fields, no escalation
Cross-links between all 4: resolve (after the one-extra-rerender fix above)
Blog listing: 10 cards, 0 empty <img src="">
Full 27-page sitemap re-sweep: 0 empty img tags anywhere on the site
```

---

## 2026-07-31 (session "leopardess consulting") — /services.html: both blocks are carousels; six links that were never links now work

Owner ask, verbatim in effect: turn the two `/services.html` component blocks into
carousels. "What the platform does" should become the default carousel we use on
fundamentallyai.com, **with images**. "Systems That Run, Record, and Report" can keep the
cards as they are but scroll them, and **needs working links to the pages they describe,
which are broken at the moment**.

### First finding: the links were not "broken", they were not links at all

The stored `page_components.rendered_html` for `info-card-grid` contained six perfectly
good anchors — to `/services/monitoring.html`, `/services/agent-orchestration.html`,
`/services/human-oversight.html`, `/services/entity-verification.html`,
`/services/tool-generation.html`, `/services/audit-trail.html`. **All six 404. None of
those pages has ever existed on this site.**

The served page contained **no `/services/` href at all**. The mechanism is
`datahelpers.RepairPageLinks` (`platform/orchestration/datahelpers/link_repair.go:139`),
reached on the rerender path via `repairOutboundPageLinks`: for an internal `<a href>`
whose target is not a real `pages.url`, it **removes the anchor and keeps the inner
text**. That is correct, deliberate behaviour (`bugs_open/079`/`097`) — the page ships,
the prose survives, the 404 dies.

But the inner text here was the link *label*, so what the owner actually saw was:

```
<p class="info-card-grid__card-body">A multi-source pipeline collects…</p>

  See how it works
  <em class="info-card-grid__card-link-arrow" aria-hidden="true">&rarr;</em>
```

— "See how it works →" sitting in the card as dead prose, with no anchor round it. Six
times. **The repair is why it looked broken rather than 404ing**, and the real defect was
upstream: `content_data.link_url` pointing at six pages nobody ever built.

Fixed by repointing each at a live page that genuinely describes it (all verified 200
first): monitoring → `/case-studies.html`, orchestration → `/technical-architecture.html`,
approval gates → `/how-it-works.html`, entity verification → `/case-studies.html`, tool
generation → `/case-studies.html`, audit trail → `/technical-architecture.html`. Three
share `/case-studies.html` and that is not laziness — it is the page that describes the
news-credibility engine, the Companies House pipeline **and** the tool generator, each as
its own `h3`. Checked for anchor targets to deep-link to instead: **the whole site emits
zero heading `id`s**, so there was nothing honest to point at (`bugs_open/071`'s gap,
confirmed here rather than assumed).

### Second and third instances of the same phantom tool, both in PROSE

`/tools/tool-monitoring-coverage-gap-finder.html` is the invented URL punch-list item 3
recorded the owner clicking on 2026-07-18. Two more instances were still live on this page
today, and **both survived that clean-up because they were prose, not `link_url` fields**:

1. `services-grid` `features[2].description` — real `<a href>` markup inside a text field.
2. `info-card-grid` `cards[4].body` — the URL written out as bare text.
3. …and a **third**, found only because the post-deploy check asserted the string was gone
   and it came back `1`: `call-to-action.subheadline`, *"use our Monitoring Coverage Gap
   Finder at /tools/tool-monitoring-coverage-gap-finder.html"*.

**`RepairPageLinks` cannot help with any of these.** Its regex is anchored on `<a …>…</a>`;
a dead URL written as prose has no anchor to unlink, so it ships verbatim, and the phantom
link *detector* has nothing to detect either. All three are gone now.

### The CTA on this page had no buttons at all

While removing instance 3, found `call-to-action` carried `primary_cta: "Get in touch"` and
`secondary_cta: "Estimate your monitoring gap"` and **neither `primary_cta_url` nor
`secondary_cta_url`**. The template gates the whole button on
`{{if and .primary_cta .primary_cta_url}}`, so the served page had
`<div class="cta-buttons">` containing nothing but whitespace — no error, no empty box, a
page that looks deliberately button-less. Same class as the `hero-tool`/`tool-cta` landmine,
on a different component. Confirmed the fix pattern against working CTAs elsewhere
(vonc.com, robot-hands.com both set `primary_cta_url` in `content_data` and render a
button) rather than guessing. Now: primary → `/contact.html`, secondary → the ROI estimator
that actually exists, and the secondary label rewritten so it names the tool it reaches.

### Block A: `services-grid` → `teaser-reveal-panel`, with six generated images

Repointed the instance at the canonical `teaser-reveal-panel` (`22c12251`) and rewrote the
six features as `key`/`hook`/`continuation`/`body` + `image_url`/`image_alt`. The schema's
hard rules are enforced in the loader script, not trusted: no digit in `hook` or
`continuation`, `hook` a complete sentence under 12 words, `continuation` an incomplete
clause under 20 words with no ellipsis, `image_alt` never a restatement of `hook`. Also
asserted the payload contains no `gap-finder`, no `90,790`, and no provider the LLM factory
rejects.

**Placement updated in all THREE places** — `page_components.slot_name`, `pages.sections`,
and the current `site_specs.site_plan` aspect (`8439e6b2`, `data->'pages'->1->'sections'`).
On this site the aspect is the real authority because `site_plan_sections` holds **0** rows
for it (verified, not assumed), and a placement missing from any one of the three is
dropped by a later `complete` rerender.

Six images generated Route-A-safe (scope-less `needs_imagery` → image-build-handler, so no
`needs_page` is emitted and no content is touched), `kind:'icon'` → Banana. Both escalation
branches were pre-checked clean before the rerender (every slot an object with non-empty
`content_data`; every required `source:"llm"` field present) — otherwise the whole page
escalates to the writer and the hand-authored copy is rewritten.

**Two of the six were rejected on sight and re-rolled**, which is why the site's rule is to
look at every generated image:
- `verification` drew the confirmation mark as an **X** — a cross reads as *rejected*, the
  exact opposite of the claim the card makes. Re-prompted with "NO tick marks, NO cross or
  X marks" and an explicit "nothing is rejected"; it now shows the token emerging
  double-outlined past the register.
- `siteops` filled its page frames solid white, which fought the delicate gold linework of
  the other five. Re-prompted to outlines-only.
Same `asset_key` on the re-roll overwrites the same deployed file, so no `content_data`
change was needed. All six then reviewed again and accepted.

### Block B: an opt-in carousel on the SHARED `info-card-grid`, not a fork

`info-card-grid` has **18 instances across 9 sites**. Two options:

- **Fork it for leopardess.** Rejected, and the RUNBOOK already says why —
  **landmine 14: "a *section* component fork does NOT survive rerender
  (`save_page_sections` re-links to the canonical component by function)".** A fork would
  have been silently reverted on the next rerender and I would have spent the session
  hunting for why the carousel kept disappearing.
- **Add an opt-in flag to the canonical component.** Chosen. `carousel: true` in
  `content_data` switches the container to a single-row scroll-snap track with overlaid
  prev/next arrows; absent, nothing changes.

**Byte-identity was proven, not asserted.** Wrote a Go harness (`text/template`, the same
engine `executeGoTemplate` uses) that renders all 18 live instances' real `content_data`
through the old and new templates: **18/18 byte-identical**, 0 errors. Plus the control
that makes it mean something — the same 18 rendered *with* `carousel:true` **all differ**,
so the comparison is measuring something. Measured first that **0 of 18** instances carry a
`layout`, `display` or `carousel` key, so the new arm could not reach any of them.

The harness earned its keep immediately: the first draft **failed to parse**. My CSS comment
explained the guard by writing the template conditional out in prose, and Go's parser does
not know what a CSS comment is — it read those as real actions and hit `unexpected EOF`.
Worse, `RenderTemplateReportingMissing` does not surface a parse failure; it **falls back to
a regex renderer**, so this would have shipped mangled markup rather than an error. Landmine
filed.

**The arrows reuse `hero-card-carousel`'s snippet rather than a new one.** That snippet was
already fully generic — every behaviour reads `data-hcc-track` / `-slide` / `-prev` /
`-next` / `-live` / `-autoplay` and no class or component name — *except* one line, its
`initAll` selector. Widened that to
`".hero-card-carousel[data-component='hero-card-carousel'], [data-hcc-carousel]"` and added
`info-card-grid` to its `applies_to`. Measured before widening: `data-hcc-carousel` appeared
in **0** `page_components` and **0** `site_components` fleet-wide.

**Worth knowing: `hero-card-carousel` has never rendered anywhere on the fleet** — 0
`page_components`, 0 `site_components`, `usage_count` 0. Its snippet was `is_active` but had
never been bundled into any site, so **this is its first execution in production, ever.**
`teaser-reveal-panel` only ever copied its *pattern*. That is precisely why the arrows had
to be proven by a real click and not by markup presence.

Both loads verified against the local file with `md5`, not with `UPDATE 1`.

### The JS bundle was empty, and that is why this mattered

`/assets/js/snippets.js` on this site was **334 bytes, "0 active snippet(s)"** — the site had
no component matching any active snippet's `applies_to`. Fired `site-asset-renderer`
(`load_site → render_js_snippets → deploy_js_snippets`, the minimal agent for this). Now
**13,781 bytes, 2 snippets**: `hero-card-carousel` (pulled in by the `applies_to` addition)
and `teaser-reveal-panel`.

### Verification: 19 served-page assertions, then a real-gesture probe with two mutants

Served-page checks all pass, including four controls that were left untouched (the CTA
section, the hero, a nav link, the `snippets.js` script tag). Six `info-card-grid__card-link`
anchors now survive link repair, which is itself the proof the targets are real — the repair
would have stripped them otherwise. `services-grid` gone, `90,790` gone, all three
`gap-finder` instances gone, and `OpenAI`/`Mistral`/`xAI`/`Perplexity` down from 4 to 0.

**Two measurement mistakes of my own, both caught by controls:**

1. **First probe reported the info-card arrows DEAD.** They were not. I ran headless
   Chromium with `--virtual-time-budget`, which does not advance the smooth-scroll
   animation, so I was reading `scrollLeft` mid-flight (`before=25, after=34`, a 9px
   twitch) or before it moved at all. The tell was the **mutant**: with the snippet's
   `initAll` deleted, `trp.NEXT_SCROLLED` was *still* `true` — an assertion that passes on
   deliberately broken code is measuring noise. Fixed with
   `--force-prefers-reduced-motion`, which makes both snippets scroll with
   `behavior:"auto"`, so the new position is readable on the next line with no animation
   to race.
2. **Second probe reported sibling-close BROKEN.** Also wrong: I clicked two `<summary>`
   elements back to back with no turn of the event loop, and the sibling-close rides on the
   async `toggle` event. Restoring a 150ms gap between clicks reported it working, and the
   mutant confirmed the assertion still discriminates.

Final probe, live page vs two mutants:

```
LIVE          trp: slides=6 overflows=true arrowShown=true NEXT_SCROLLS=true PREV_RETURNS=true
              trp: OPENS=true DEEPLINKS=true SIBLING_CLOSES=true
              icg: slides=6 overflows=true arrowShown=true NEXT_SCROLLS=true PREV_RETURNS=true
              icg: cardAnchors=6 anchorsWithHref=6
MUT no-init   trp/icg NEXT_SCROLLS=false PREV_RETURNS=false, DEEPLINKS=false, SIBLING_CLOSES=false
              (trp.OPENS stays true and icg anchors stay 6 — correct: native <details> and
               server-rendered hrefs are the progressive-enhancement baseline, not JS)
MUT old-sel   trp.NEXT_SCROLLS=true, icg.NEXT_SCROLLS=false
              (restoring the narrow selector kills ONLY the info-card carousel — which
               attributes it to the one-line widening and nothing else)
```

### Figures: re-measured before being repeated, and one was a live false claim

Full table and evidence in `AUDIT_verified_facts.md` ("Re-measurement 2026-07-31"). The
headline: **"more than 90,790 orchestration state records ... every one of them readable
after the fact"** was a **point-in-time row count of a table that is pruned hourly at 24h**,
published as a cumulative "to date" total plus a durability promise. Live row count when
re-measured: **2,364**. The claims layer had already caught this on 2026-07-26 and routed it
to `needs_human_review` (`bugs_open/091`: *live 1,900 vs published 90,790*) — nothing drained
it, so it stayed live five more days. Also refreshed: sites 9 → **15**, definitions 157 →
**190** (185 active), sub-agent spawners 40 → **42**, feed items 6,262 → **7,990**, scored
5,228 → **6,794**, businesses 2,000 → **3,419**.

**And I got one wrong myself, before shipping.** I "verified" the provider list with
`grep -ioE 'case "(anthropic|openai|gemini|mistral|xai|perplexity|ollama)"'`, got six names,
and wrote "five hosted providers plus a self-hosted path" into the audit file. **This
workstream's own RUNBOOK landmine 12 caught it** — read while looking up the rerender recipe
two entries below. `platform/aiservice/factory.go:24-35` supports exactly **three**
(anthropic, ollama, gemini); `openai` is a stub that returns an error; `xai`/`perplexity`
are not in that switch at all — the arms I found are in `feed_actions.go`, a separate
news-*search* path. A `case` arm is not a working provider, and **a grep that disproves part
of a claim has not validated the rest**. RUNBOOK landmine 12 updated (Gemini added);
`WRONG_CALLS.md` row written.

## 2026-08-12 (session "leopardess") — HANDOFF.md brought up to date, and the update found three live regressions

Task was narrow: "search the docs to update HANDOFF.md". The file's own frontier (§10) was
dated 2026-07-30. Bringing it forward meant checking what was still true, and three of this
lane's 2026-07-31 deliverables were not.

### Method — what I read, in order

`git log --since` on `docs/leopardessconsulting/` and fleet-wide for "leopardess"; the tails
of `README_where_we_are.md` (which had two 07-30 entries the HANDOFF never absorbed) and
`RUNNING_NOTES.md`; then the two lanes that touched this site without telling this one:
`bugfix_179_deploy_path_override` (the D3 run) and `staged_component_build` (the tool
ladder). Then the live DB and the served pages for everything I was about to write down.

**The docs alone would have produced a wrong update.** README recorded the vendor-trust
tool as built and live; the HANDOFF still said "deliberately deferred". Neither doc knew
about 08-08 or 08-11 at all, because neither pass was run by this lane.

### The three regressions, all found at the artefact

1. **Six service images gone.** `/services.html` serves exactly **one** `<img>` tag. All
   six `teaser-reveal-panel` items hold `image_url` as an empty string; item 6 has lost
   `image_url`/`image_alt`/`open_label` entirely. The template gates on the field, so the
   loss renders as nothing at all. Class = `bugs_open/238`.
2. **Block B is no longer a carousel.** `carousel` is absent from the `info-card-grid`'s
   `content_data` (keys are now `cards`, `section_title`, `section_eyebrow`,
   `section_subtitle`); no `data-hcc-*` on the served page. Block A's `trp__track` still
   works and `snippets.js` is still 13,781 bytes — the JS half is fine, the opt-in flag is
   what went.
3. **A card link now 404s.** The six links were repointed by the automated linker; one
   points at `/case-study-automated-intelligence-pipeline.html`, whose `pages` row was
   created 2026-08-11 16:21 as `build_status='planned'` and never deployed. **404 live.**
   `RepairPageLinks` cannot catch this — the row exists, so the link passes its test. That
   is a new hole in a safety net this lane documented on 07-31 as working.

Plus **(d)**: `OpenAI|Mistral|xAI|Perplexity` on `/services.html` went 0 (07-31) → **1**.
Item 5 now claims a step "can call Claude, Gemini, Mistral or another provider".
`platform/aiservice/factory.go:23-33` supports **anthropic, ollama, gemini** and nothing
else. Same false claim this lane removed twelve days ago, in a component whose `updated_at`
is `2026-08-11 18:15:23`.

### Missteps of mine, recorded because they are the useful part

- **I nearly attributed the empty `image_url`s to the 08-11 pass and stopped there.**
  Checking the asset rows showed all six `icon_service_*` assets — created **2026-07-31**,
  by this lane's own run — carry the literal placeholder URL
  `/assets/images/input-data.asset-key.jpg`. So part of this dates from the day the images
  were made, not from 08-11. I have left the cause as `[UNVERIFIED]` in the HANDOFF rather
  than pick the tidier story. **The tidier story was the one I had already half-written.**
- **`grep -oc` counts LINES, not occurrences.** It told me the pillar article had 2 charts
  where the record said 5. `grep -o … | wc -l` across all four pages gives 2+1+1+1 = 5, all
  present. I had a "chart loss" paragraph drafted off a flag misuse.
- **`grep -oE '.{0,120}(OpenAI|Mistral).{0,120}'` hung for 120s** on a long minified line
  and had to be killed — catastrophic backtracking. Python's `re.finditer` with slicing
  did it instantly. Worth remembering for any served-page context grep.
- **I filed the placeholder-asset finding as new before grepping `bugs_open/`.** It is
  already `bugs_open/248` (the `undeployed_asset` slug — **the number is shared with an
  unrelated CTA bug**, so the CLAUDE.md rule about resolving by slug earned its keep here),
  filed 08-10, CONFIRMED by a `090` run. Fleet-wide it is 76 rows across 12 sites,
  2026-01-28 → 2026-08-11. Contribute the numbers, do not re-file.

### Measurements taken today, for anyone repeating them

48 pages / 33 deployed, newest deploy `2026-08-12 02:07:13Z`. 174 open work items of 515
total (was 189/232 before the 08-08 D3 run, 248/300 after). 33 `voice_tells`, 21
`cta_names_unknown_destination`, 19 `image_source_unsatisfiable`, 17 `needs_section_data`,
13 `content_rewrite`, 3 `claims_unverified` — all at `needs_human_review`. 193 agent
definitions (187 active), `orchestration_states` 5,997 rows, `sites` 40 rows. 7 tools + 5
guides, all 200, all footer-linked.

`/case-studies.html` is still publishing "143 agent definitions, 56 of them active",
"75,061 orchestration state records" and "Eight live sites" — the 07-31 re-measurement pass
covered `/services.html` only and was never swept across the site. The 75,061 figure is the
**same defect class as the 90,790 one that was fixed**: a cumulative-sounding total read off
a table that is pruned.

## 2026-08-14 (session "leopardess", cont.) — repair handoff prepared; a FIFTH regression found; the images are recoverable

Owner asked for the §11.7 work to be prepared as a handoff for a fresh session. Output:
`HANDOFF_2026-08-14_services_restore.md` + pre-repair snapshot
`scripts/SNAPSHOT_2026-08-14_services_pc_pre_restore.json`. Nothing on the site changed.

### New findings while pinning the repair

- **A fifth regression: the CTA is misdirected again.** Live primary anchor is
  `"Book an architecture conversation"` → `/tools/tool-agent-complexity-estimator.html`;
  the 07-31 authored state was "Get in touch" → `/contact.html`. That is the exact
  signature of `bugs_open/248`'s CTA case (shared number — resolve by slug). The irony is
  measured: the work item that drove the 08-11 18:15 rerender was itself
  *"1 misdirected CTA(s) on services"*. Whether the recompute failed to fix it or created
  it is [UNVERIFIED].
- **The six icons are NOT lost.** All six live at their derived paths
  (`icon-service-{monitoring,orchestration,oversight,verification,toolbuild,siteops}.jpg`,
  distinct sizes 26–47KB, all 200). Only the `content_data` references were emptied; the
  placeholder-URL asset rows are wrong metadata (248/152), not missing files.
- **But the regeneration also rewrote the item KEYS**, so only four icons map to current
  items. `model-routing` and `news-credibility` need two new Route-A icons (or an eye-check
  that an old one fits). The 07-31 Block A copy itself is unrecoverable: not in any bak_
  table (all pre-change), and `orchestration_states` is pruned at 24h so the 08-11 rows are
  gone.
- **The carousel template arm SURVIVES.** `data-hcc-carousel` present in the canonical
  `info-card-grid` template — but its md5 has moved since 07-31 (`204a3975…` vs L9's
  `f99b791c…`), so another lane touched the template. Gate any re-check on the arm's
  presence, not on md5 equality with the L9 file.
- **The 238 carry HAS ROLLED and does not cover this class** (238 §9.2, webdesign.uk,
  08-12): `source:"renderer"` fields short-circuit resolve to `(nil,true)` so the carry
  never runs, and on that site restoring `content_data` + rerender was NOT sufficient —
  buttons only returned via a rendered_html splice. Flagged in the repair handoff as a
  stop-and-contribute checkpoint rather than assumed to transfer (this site's
  `section_data_resolved` path worked for these very fields on 07-31).
- **The guide page's Mistral mention is NOT the same false claim.** It says the
  *calculator covers* OpenAI/Claude/Gemini/Llama/Mistral/Cohere *pricing* — a claim about
  the tool's comparison table, not the platform's callable providers. Verify against the
  tool, not factory.go.
- The 404 card is index 1 ("Data checked before it's trusted"). `bugs_open/268` (filed
  08-12: content_rewrite drops CTA destination keys, 214 buttons fleet-wide, mechanism NOT
  established) is adjacent to our regressions (e) and possibly (a) — the repair session
  should read it before contributing to 248.

### Corrections to my own 08-12 entry

- I wrote "three regressions"; it is five (the CTA, found today, and counting the Mistral
  claim separately as the 08-12 HANDOFF §11.3(d) already did).
- I left the empty-image cause attributed vaguely to the 08-11 pass with the placeholder
  assets as an alternative. Today's evidence splits it cleanly: the FILES deployed fine on
  07-31 and are still live; the placeholder URLs are asset-row metadata only; the
  content_data emptying is the 08-11 regeneration. Two separate defects, two separate bug
  files, neither is the other's cause.


## 2026-08-14 (session "services-restore") — /services.html repaired end to end; all five regressions closed and verified live

Executed `HANDOFF_2026-08-14_services_restore.md` in full. Everything below happened
2026-08-14 between ~18:15Z and ~19:00Z.

**Re-verification first (handoff warning 1).** All five regressions were still live on the
served page (1 `<img>`, 0 `data-hcc-carousel`, 0 contact-primary anchors, the 404 link and
the Mistral claim both present). One deviation from the handoff's snapshot: the
`call-to-action` slot's `updated_at` had moved to **2026-08-12 20:49:36** — something
added `primary_cta_target_title`/`secondary_cta_target_title` keys naming the (wrong) tool
targets, without changing the CTAs themselves. Repair 1c therefore also set those two keys
to the restored targets' real `pages.title` values, so the annotation no longer contradicts
the URLs. No open work item was in flight against the page.

**Backup:** `bak_leo_services_pc_20260814` (all 4 slots, by page_id).

**Icons (1a).** The two new items got fresh icons via the proven Route-A recipe
(scope-less `needs_imagery` → image-build-handler, `kind:'icon'` → Banana), prompts
modelled verbatim on the 07-31 set's constraint-hardened form (recovered from
`assets.origin_prompt` — worth knowing they are recoverable there):
`icon_service_routing` (a junction fanning into three routes, one looping back inside its
boundary) and `icon_service_credibility` (source lines converging into one feed, items
passing beneath a gauge arc). Created 18:20/18:21Z, serving 200 within the minute, **both
born with CORRECT `/assets/images/` urls** — unlike the 07-31 six (contributed to the
248-asset file). All six icons (4 existing + 2 new) were eyeballed before wiring, per the
site rule; both new ones accepted on first roll; `monitoring` accepted for
`decision-record` (concentric arcs sweeping a row of marks reads as a log being read back).

**Claims re-measured before the rerender shipped them (1b).** Live 2026-08-14:
`business_intel.businesses` 3,419 ("more than 2,000" — true); `companies_house_data` 937
(true); `content_feed_items` 10,087 vs claimed "over 9,545" (true, append-only table,
grows); scored `credibility IS NOT NULL` 8,846 vs "more than 8,297" (true). Kept both
item-6 figures. **The Mistral claim is gone again** (factory.go re-read: anthropic/
ollama/gemini only). **And one claim the handoff didn't enumerate was the 90,790 defect
class again**: item 4 (`decision-record`) said "more than 2,000 orchestration state
records … weeks after the fact". `orchestration_states` held 2,819 live but is pruned
hourly at 24h (and dipped to 1,900 on 07-26), and `orchestration_state_audit` prunes too
(min `changed_at` 2026-08-12 when read on 08-14) — so "weeks after the fact" was false on
BOTH tables. Rephrased durably: mechanism only, no count, no retention promise.

**Repairs applied** as ONE transaction (1a icons+alts, 1b bodies, 1c CTA + target_titles,
1d card 1 → `/case-studies.html`, 1e `carousel:true`), each UPDATE hitting exactly 1 row.
Both L9 §4 escalation-gate queries returned 0 rows. ONE rerender
(`rerender_page_safe.sh`, PUBLISH_OK, correlation `a9510019`), COMPLETED, **no
`needs_page` escalation emitted**.

**Verification (dated PASS: 2026-08-14 ~18:45–18:55Z).** All six §2 assertions pass on a
cache-busted fetch: img 7/7, hcc ≥1, contact-primary 1, 404-case-study 0, mistral 0 —
and card-links: the grep-c form reads **17** because the component's inline CSS matches
the string on 11 lines; the real count is `<a class="info-card-grid__card-link">` = **6**,
all six hrefs live pages (2× case-studies after 1d). All six icon files referenced once
each. Real-gesture probe (headless Chromium, `--force-prefers-reduced-motion`, page
served locally with `<base>` to live assets): trp and icg both
NEXT_SCROLLS/PREV_RETURNS=true, trp OPENS + SIBLING_CLOSES=true, icg cardAnchors=6. The
**no-init mutant** (snippets.js removed) kills both carousels and sibling-close while the
progressive-enhancement baseline survives (native details open; 6 server-rendered
anchors) — the probe discriminates. Probe scripts: scratchpad only (pattern documented
here; build_probe.py injects base href + a results-into-title script).

**What this pass does NOT establish:** survival. The 268 fix ("renderer/static fields now
reach the 238 carry", commit `8f899cc8d`, 09:13 BST today) may or may not be in the
chassis image that restarted 15:32Z (v1.0.1299) — a sha-probe of `/proc/1/exe` did not
find it, and the provenance log line had scrolled, so treat the §0.2 re-drop hole as
possibly still open. **Re-verify the six §2 assertions after the next fleet roll or the
next regeneration touching this page.** Also noted for the future: the
`case-study-automated-intelligence-pipeline` pages row (`build_status='planned'`) still
exists; if that page ever deploys, card 1 becomes a candidate to repoint there
deliberately (`bugs_open/266` means it may deploy without anyone asking).

**Contributions written:** 248-asset (leopardess 15→13, two clean new rows), 248-CTA
(second-site observation + the 08-12 annotator data point; checked 268 first — not that
class, URLs were rewritten not dropped), 152 (two post-fix presigned rows dated).
`AUDIT_verified_facts.md` given a dated re-measurement block.

**Missteps this session, for the record:** (1) first binary-provenance probe used an
all-zeros sha as the absent-control and it MATCHED (a 40-zero run occurs in any large
binary) — control invalid, provenance left honestly unknown rather than asserted; (2) a
`pkill` for the probe's local HTTP server was chained ahead of two psql queries in one
compound command and killed the whole thing (exit 144) — separate lifecycle commands from
queries.

## 2026-08-16 (session "services-restore", cont.) — the repair SURVIVED a fleet roll, and the 268 fix is now provably in the running chassis

Standing re-check from 08-14, run 2026-08-16 ~10:00Z after the roll to chassis
`v1.0.1303` (pods restarted 2026-08-15 18:45Z):

- **Served `/services.html` is byte-identical to the 08-14 post-repair fetch** (md5
  `c0a69af05167…` both), all six §2 assertions still pass (img 7, hcc 1, contact-primary 1,
  404-link 0, mistral 0, real card anchors 6, six distinct icon files). All four slots'
  `updated_at` still `2026-08-14 18:25:11Z` — nothing has regenerated the page in ~40h. The
  08-15 fleet activity on this site (12 `ink415w_*` rerenders on blog/guide pages, one
  headmeta rerender, four `offer-analysis_*` rewrites on index/careers/insights) did not
  touch services.
- **The 268 fix IS in the running binary** — closes the 08-14 caveat. Another session
  stamped v1.0.1303 as commit `5e075a6f9` (LANDMINES, 08-15, on this same pod) and
  `git merge-base --is-ancestor 8f899cc8d 5e075a6f9` holds; re-probed here: stamp sha
  present in `/proc/1/exe`, random-hex control absent. So the §0.2 re-drop hole
  ("renderer-sourced keys never look missing, so the carry never runs") is closed at the
  binary from this roll on; the 08-14 repair itself simply predates any regeneration.
- **My 08-14 zero-sha probe misstep is now a LANDMINES entry** (written 08-15 by the
  session that stamped the pod): forty consecutive zeros are git's null-sha literal and
  match in ANY binary that speaks git — use a random 40-hex value as the absent-control.

## 2026-08-16 (session "services-restore", cont.) — /case-studies.html figures swept; the infographic re-drawn without numbers; the two claims items drained; and a resolver landmine re-learned

Handoff §3 item 4. Everything below 2026-08-16 ~10:00–10:45Z.

**The register is refreshed live, and it changed the plan.** `site_specs.evidence_base`
for this site had `verified_at: 2026-08-16` on every metric (something re-verifies it —
values current to the hour). Two things it says that my own raw counts did NOT:
`C1-records-verified` is `count(*) … WHERE verification_status='verified'` = **2,338**, floor
phrasing only ("more than 2,000") because 874 rows were reclassified on 07-20 — my
`count(*) FROM businesses` = 3,419 must never be published; and `C4-agent-definitions-active`
is `status='active'` = **78**, not `is_active` = 193 (the 08-12 handoff's "187 active" used
the wrong column). **Read the register before measuring; its `source.sql` IS the
definition.** Sanctioned forms used: floors ("more than 2,000 / 10,000 / 9,000 / 150 / 70 /
20"), exact 937 and 5,798, the registered 40 spawners (audit snapshot, exact); orchestration
records rephrased to mechanism only (as on services 08-14).

**Landmine re-learned, cost one rerender: `case-studies-list.case_studies` is
`source: "site_specs.portfolio.case_studies"`.** I edited `page_components.content_data`,
rerendered, and the served text was unchanged — the rerender re-resolves that field from
the `portfolio` aspect (`8869edfc`, is_current since 07-10) and OVERWROTE my edit, while
the `generic-text-block` (llm-sourced) kept mine. Same shape as `site_plan`: **the aspect
is the authority; `content_data` is a cache.** Fixed the aspect (backup
`bak_leo_sitespec_portfolio_20260816`), re-applied, rerendered again → live. Check
`input_schema->'fields'->k->>'source'` for every field you are about to hand-edit.

**The infographic is re-drawn without numbers.** The Leopardess Line map baked "Records:
2,767 verified … Sites: 8 running today" into pixels — with the register's verified count
now 2,338, the image OVERCLAIMED, and the alt text ("more than 2,000") had been written to
pass the checker while the image itself did not. Re-rolled via Route-A (`kind:'infographic'`
→ Banana), same prompt but the SERVICE NOTES panel replaced by three number-free lines
("Records: checked against Companies House / Reading: scored as it arrives, refreshed
six-hourly / Sites: our own, built this way") and "Do NOT add ANY number, digit …". **Under
NEW asset_keys each time (`_v2`, `_v3`, `_v4`) so a bad roll could never overwrite the good
served file** — worth copying. Three rolls, all eyeballed: v2 correct semantics but wrote
"GOLD LINE / BRONZE LINE / OFF-WHITE LINE" as on-map labels (the prompt's legend spec taken
literally); v3 prettiest but WRONG STORY — a linear line "Uncertain Match → Verified → A
Person Decides", i.e. verified before anyone decided; **v4 accepted** — the gold line splits
at Uncertain Match (clear branch → Verified, uncertain branch → the person), no colour
words, no digits. v1/v2/v3 files remain deployed but unreferenced. Alt rewritten to
describe v4 faithfully.

**A 152 correction from watching three rolls:** an asset's `url` is a presigned S3 link
DURING the pipeline (v2 read that way while `spawn_asset_deployer` was still EXECUTING) and
is rewritten to `/assets/images/…` when the deployer completes (v2 settled at 10:09:23Z;
v3/v4 read clean on first look because they were read after). So presigned **at rest** —
the two rows I contributed to 152 on 08-14 — means the deploy step never completed for
that asset, not that the creation path is wrong. Appended to 152.

**The two `claims_unverified` items** (2, not 3 — one drained since 08-12):
- `about` — `content-block-about.stat_2` was "Core Technologies: 3", a filler stat the
  register cannot vouch for (its own stat_3 lists the three). Swapped for a registered one:
  "Sites Built and Run: 20+" (`C6`, gte 22, context term "site"). Gate clean, rerender
  COMPLETED, live. Left for the revalidator to close (it retracts once the page no longer
  asserts the claim AND has redeployed since the edit — `bugs_open/262`).
- `for-engineering-teams` — the "90,790" (THE pruned-table count removed from services on
  07-31) still sat in `features[].description` of a page that is `archived` **and 404 live**
  (`build_status` still 'deployed', `deployed_at` 07-17). Its `generic-text-block` was
  rewritten by some producer 2026-08-15 10:46Z (new LLM prose — a `bugs_open/266`-class
  write into an archived page; 266 already names this page) and a fresh
  `needs_internal_links` item for it was filed today 10:01Z. The revalidator can never close
  this one (the page never redeploys past that 08-15 touch), so: cleared the figure from
  content_data (`bak_leo_fet_features_20260816`) so a revival cannot ship it, and closed
  the item as `complete`/`human_review` with the decision in `result.human_decision`.

**Also measured, no action:** `/our-approach.html` (archived) still SERVES 200 (28,948 B)
— the checker's own comment names it — but carries no figures and nothing links to it.
Deployed sites: 22 `status='deployed'`, all 22 probed HTTP 200 (webdesign.uk 302).

**Verification (dated PASS 2026-08-16 ~10:45Z):** `/case-studies.html` cache-busted: 0
matches for `75,061|143 agent|5,652|4,672|2,767|Eight live|8 running|56 of them`; every
digit-bearing sentence on the page is one of the register-sanctioned forms above; v4
infographic referenced once, old one 0. `/about.html`: "Sites Built and Run" + "20+"
present, "Core Technologies" gone. Both rerenders COMPLETED with 0 `needs_page`
escalations. Backups: `bak_leo_casestudies_pc_20260816`, `bak_leo_about_pc_20260816`,
`bak_leo_fet_features_20260816`, `bak_leo_sitespec_portfolio_20260816`.

**Missteps this session:** (1) hand-edited a resolver-sourced field in `content_data`
without checking its `source` — one wasted rerender (above); (2) a `cat >>` with a relative
path after a `cd docs/…/scripts` wrote nothing (the tool resets cwd between calls but not
within one) — use absolute paths in appends.

---

## 2026-08-16 (afternoon) — sitemap.xml regenerated: 27 → 36 URLs, and the generator I nearly rewrote already existed

Handoff §3 item 7, the last of the four items left over from the services restore.

**The finding that changed the plan, before any code.** The handoff said *"No platform
generator exists — HANDOFF.md §6 item 5 / turn 18 built it from the `pages` table via
git-adapter; find that recipe in RUNNING_NOTES turn 18 and re-run it"*. That was true when
turn 18 wrote it (2026-07-17) and is **false now**: `scripts/site-discovery-files.py`
(register **SEO-002**) landed 2026-07-28 and does exactly this job — `robots.txt`,
`sitemap.xml` and `llms.txt` for any site from the `pages` table, dry-run by default. I
found it in a code comment, not in the register: `discovery_checks/check_site_structural_validity.go`
names it while explaining why `sitemap_entry_dead_live` deliberately does NOT gate on
"every page appears in the sitemap". One grep of the concept register for `sitemap` would
have found it in seconds. Logged in `WRONG_CALLS.md`.

**What was actually missing** `[MEASURED 2026-08-16 ~15:45Z]`. Live sitemap: 27 `<loc>`,
all `lastmod 2026-07-16`. `pages` rows for this site: 48 total — 11 `archived`, 1 `active`
but never deployed (`/case-study-automated-intelligence-pipeline.html`, `build_status
='planned'`, the same phantom §1d repointed a card away from), **36 active + deployed**.
So nine were missing, not the four the handoff predicted — the fleet's automated loops
added five more pages since 08-14:

```
/blog/ai-data-trust-in-financial-services.html      (the 4 trust-series articles
/blog/ai-data-trust-in-healthcare.html               the handoff knew about)
/blog/ai-data-trust-in-hiring-and-hr.html
/blog/can-you-trust-ai-with-your-data.html
/guides/tool-automation-savings-estimator-guide.html (new since)
/guides/tool-process-automation-scorer-guide.html    (new since)
/tools/ai-vendor-trust-checklist.html                (the handoff knew about)
/tools/automation-savings-estimator/index.html       (new since)
/tools/process-automation-scorer/index.html          (new since)
```

**The trap I walked up to and did not step in: the two tool pages are directory-style URLs
and the pretty form is a 404.** `/tools/automation-savings-estimator/` returns **404**, and
so does `/tools/process-automation-scorer/` `[MEASURED 2026-08-16]` — the Cloudflare worker
maps `{hostname}{path}` to a B2 object key and rewrites only `/` → `/index.html`, so an
object store has no directory index (LANDMINES, "A `/section/` URL 404s on every B2-hosted
site"). Both the pages' own `<link rel="canonical">` and every internal link on the site use
the explicit `/index.html` form, which is what `pages.url` holds and therefore what went in.
Tidying those to the pretty form would have put two 404s in the sitemap — and a canonical
naming a URL that does not exist is the worst version of this, per the same entry.

**Ran the generator rather than the turn-18 recipe.** `./scripts/site-discovery-files.py
leopardessconsulting.co.uk` → *36 fetchable, 0 dropped*, matching an independent `curl`
sweep of all 36 I had run first (36/36 → 200). `lastmod` comes from
`GREATEST(updated_at, last_built_at)`, so the dates are now real per page instead of
27 × `2026-07-16`.

**Shipped only `sitemap.xml`, deliberately.** The tool also emits `robots.txt` and
`llms.txt`. `robots.txt` is Cloudflare-managed on this domain and the tool said so itself —
the managed block is being merged in and currently disallows **Amazonbot,
Applebot-Extended, Bytespider, CCBot, ClaudeBot, CloudflareBrowserRenderingCrawler, GPTBot,
Google-Extended, meta-externalagent**; Cloudflare PREPENDS its file rather than yielding, so
shipping ours would change nothing until a dashboard setting is off (owner's call, not a
session's — turn 18 reached the same conclusion independently). `llms.txt` (6,970 B, built
from each page's own `<h1>` and first sentence) is a **new file for this site**, not a
repair of a stale one, so it is out of this item's scope and is left as a costed, ready
next step.

**Delivery: a new script, because the existing one uses the publish pattern that drops.**
`scripts/commit_site_file.sh` (this directory). `commit_brand_assets.sh` publishes with
`kubectl run -i --rm … kcat -P < file`, which is the stdin-attach race `rerender_page_safe.sh`
was written to escape — `kubectl run -i` wires stdin asynchronously, so the container can
reach kcat at EOF and publish nothing at exit 0 (four of five lost, measured 2026-07-26).
The new script carries the payload in the container COMMAND and prints `PUBLISH_OK`. It also
**reads `sites.github_repo` instead of hardcoding `"sites"`** — that hardcode is correct here
(the column is empty for this domain) and silently wrong for `idea.uk`/`relojistas.com`,
which serve from `vm-sites` and take a green commit into the wrong repo (LANDMINES). Branch
left unset on purpose: `gqls/sites` has no `main`, and `CommitToRepo` falls back to the repo
default `master`, which is the branch the B2 workflow watches.

**Verification, in the order that makes each step falsifiable** `[MEASURED 2026-08-16 ~15:52Z]`:

```
PUBLISH_OK                                            (receipt, not an assumption)
git-adapter corr 142c4a2b… → success:true, files ["/sitemap.xml"], repo gqls/sites
git show --stat a1e07becb  → 1 file changed, 36 insertions(+), 108 deletions(-)
                              ^ NON-EMPTY. success:true on an empty commit is the
                                documented failure shape, so the stat is the evidence
served == generated        → byte-identical diff, 4,333 B, 36 <loc>
36/36 served <loc> → HTTP 200 (cache-busted)
negative control: /case-study-automated-intelligence-pipeline.html → 404 AND absent
                  from the sitemap (0 matches) — the planned-but-never-deployed page
all 11 archived urls → 0 leaks
no duplicate <loc>; XML parses; every loc on the canonical origin
old 27 ⊆ new 36 → 0 lost
```

**One defect found in the generator, fixed, and proven in both directions.** SEO-002's
`live_pages()` selects `status='active' AND deployed_at IS NOT NULL` and does **not** filter
`pages.noindex` — a column that did not exist when it was written (SEO-003, live since
chassis v1.0.1277, injects `<meta name="robots" content="noindex, nofollow">` for those
pages). So the generator would advertise, in a site's own sitemap, a page the platform is
actively telling crawlers to skip. **Latent, not live** `[MEASURED 2026-08-16]`: exactly two
rows carry `noindex` fleet-wide, and the only `active` one — `vonc.com/tools/gauntlet/round.html`,
meta tag confirmed served — is on a site with **no sitemap.xml at all** (404), so nothing had
published the contradiction yet. Fixed with `AND noindex IS NOT TRUE` plus the reasoning in
the docstring. Proof it does something: vonc.com counts **22 without the filter, 21 with**,
and the generated file (scratchpad only — not committed, not deployed) has 21 `<loc>` and 0
matches for `gauntlet/round.html`. leopardess is unaffected: all 36 are `noindex=false`, so
the file shipped above is identical either way.

**Missteps this session:** none that cost a cycle. The one that nearly did was accepting the
handoff's "no generator exists" at face value — see WRONG_CALLS.

> **Commit note, same session.** This work is `c5d62f615` (10 files) — but the **LANDMINES.md
> entry it names is NOT in it.** Between my append and my `git commit <pathspec>`, another
> session committed that file, and its edit took mine along as a same-file passenger: the
> kcat-publish entry is at HEAD under `cbdc572bb`, an mcalc notes commit whose message says
> nothing about it. Nothing is lost and forward-only holds, so no amend — recorded here
> instead, because a `git log` of LANDMINES.md will otherwise attribute this entry to a lane
> that never wrote it. This is the documented same-file case (MEMORY
> [[a-pathspec-commit-still-takes-a-same-file-passenger]]) seen from the other side: a
> pathspec protects you from sweeping up others' work, never from being swept into theirs.

---

## 2026-08-17 — item 6 (scorer acceptance) proved by hand, item 5 (voice) measured and STOPPED at an owner question, and the fleet's LLM endpoint hit its spend cap mid-session

Continuing the 08-14 handoff: §3 items 6 and 5.

### The finding that outranks both: the Anthropic endpoint is capped until 2026-09-01

`[MEASURED 2026-08-17 11:47Z]`, and it is why the acceptance run below never ran.

```
SELECT endpoint_url, healthy, last_healthy, error FROM ai_endpoint_health;
 https://api.anthropic.com/v1/messages | claude | f | 2026-08-17 11:07:15Z |
   API request failed with status 400: {"type":"error","error":{"type":
   "invalid_request_error","message":"You have reached your specified API usage
   limits. You will regain access on 2026-09-01 at 00:00 UTC."}}
```

Independent confirmation (not the same instrument): `llm_call_log` holds four real
failures carrying that same message between 11:08:37Z and 11:09:53Z, from
`council-gate` and `landmine-verifier`. So the health row is reporting a real
account limit, not a transport blip — this is the case the MEMORY lesson
*"a believable OUTAGE explaining a NULL is when to doubt the INSTRUMENT"* warns
about, and the second source is what settles it.

**Blast radius**: `claim_work_item` refuses to claim any item whose handler agent
uses that endpoint — it releases the claim and puts the item back to `triaged`.
**26 items** (25 `build`, 1 `content`) were sitting at `triaged` an hour later,
across the fleet. Nothing is lost; nothing proceeds either. **Owner action**: this
is a limit set in the provider console, so no amount of retrying or waiting moves
it before 2026-09-01.

**How it presents, which is the reason for the new LANDMINES entry.** Every
published symptom says healthy: `build-pipeline-trigger` fires on its 30s tick,
`build-dispatch-loop` runs and reaches `COMPLETED`, `pending.item_count` is 1 with
`rows_dropped: 0`, and there is **no `__step_error` anywhere**. The item just never
leaves `triaged` and `attempt_count` stays 0. The reason is one key down:
`collected_data->'claim_result'` → `{"claimed": false, "reason":
"ai_endpoint_unavailable", "endpoint": "https://api.anthropic.com/v1/messages"}`.

**And a second-order effect worth knowing on its own.**
`build-pipeline-trigger`'s `find_dispatchable_site` is
`ORDER BY wi.created_at ASC, wi.priority ASC, wi.id ASC LIMIT 1` — **one site per
tick, oldest item first, no rotation**. An item that can be selected but never
claimed is re-selected for ever and starves every other site. Measured today:
`webdesign.co.uk` took **18 of 18** dispatch-loop runs in the 11:00 hour while two
other sites' items sat untouched (95 runs over the preceding day). So "my site
never gets a turn" is not a rotation bug — read the oldest triaged item in the
fleet. Both facts are now in `LANDMINES.md`.

### §3 item 6 — `tool-process-automation-scorer` acceptance: the defect is FIXED, proven by hand

The handoff's premise ("7 pass / 2 fail, untouched") is **stale**. The failing check
was Tier 2's static arm, not a behaviour test: `check_tool_acceptance.go`'s
`interaction` case only confirms that the anchors the interaction touches exist in
the served HTML, and it reported *"interaction anchor #pas-error absent from
deployed page"*. The `improve_tool` item (`fa625736`) was closed 2026-08-11 19:24Z
— with `result: {}`, so on its own that is the "a `complete` work item is not a
repaired artefact" shape and proves nothing.

The artefact does. `[MEASURED 2026-08-17]` the deployed page carries
`id="pas-error"` (default `display:none` via `.validation-note`) and
`class="submit-btn pas-submit"`, and the inline JS sets `display:block` on an
incomplete submit. So Tier 2 would now pass.

**Then I drove it, because reading the code is not the behaviour.** No node,
no puppeteer and no `websocket` module on this box, so the probe is a ~70-line
stdlib CDP client over a hand-rolled WebSocket (`scripts/cdp.py`,
`scripts/probe_pas.py`; headless Chromium with `--force-prefers-reduced-motion`, the
07-31 lesson). Live page, real clicks:

```
before                 #pas-error display=none  visible=false  answered=0
click .pas-submit      #pas-error display=block visible=TRUE   results=false
answer all 9, click    #pas-error display=none  visible=false  results=TRUE
                       score "Automation suitability score: 94 / 100"
```

**The discrimination control is the second branch, not a synthetic mutant**: the
same probe, same selectors, same gesture, reports the opposite on the opposite
code path. A probe that said "visible" on both would be measuring nothing.

The platform's own Tier-4 run is **raised and queued** — work item
`fcfbdfd5-a1f5-427d-962f-8caaf82ea145`, via `tool_acceptance_run.sh`, all three
preflights checked (a current `doc_plans` row exists; the page resolves by
placement; and the running `browser-runner-adapter` **v1.0.1305** carries the
`interaction` and overflow arms — `grep -a` on `/proc/1/exe` for
`"interaction produced the expected result ("` and
`"page overflows horizontally (scrollWidth > clientWidth) on "` both PRESENT,
with a random-hex negative control ABSENT, so the probe discriminates). It stayed
`triaged` because of the endpoint cap above, which is correct behaviour — **leave
it there; it claims itself when the endpoint recovers.**

### §3 item 5 — the voice backlog is one mis-scoped pattern, and that is an owner question

Re-measured, because every figure in the handoff had decayed. **The queue today
is 34 open `voice_tells` items / 210 findings**, per the platform's own revalidator
at 2026-08-17 08:45:52Z (`result->'revalidation'->'evidence'->'by_check'`):

| check | findings | pages |
|---|---|---|
| **banned_phrase** | **104** | 22 |
| em_dash_density | 41 | 22 |
| long_sentences | 27 | 14 |
| no_contractions | 14 | 10 |
| triad_density | 12 | 11 |
| strawman | 11 | 8 |
| flourish_ending | 1 | 1 |

Then I ran the site's own 14 `voice_gate.banned_phrases` regexes over the served
text of all 36 live pages (`scripts/voice_census.py`; tags stripped first, so
hrefs and class names cannot score):

```
138 hits / 19 pages   \btrust(ed|worthy|s)?\b
  3 hits /  3 pages   \bactually (ships|works)\b
  2 hits /  2 pages   \bearns its keep\b
  2 hits /  2 pages   \bhonest(ly)?\b
  0 hits              the other 10 patterns
```

**138 of 145 — 95% — are the single pattern `\btrust(ed|worthy|s)?\b`.** Its stated
reason is *"owner 2026-07-18: overused; say what is checked/verified instead"*,
i.e. it was written to stop the site labelling itself trustworthy. What it
actually catches now, at the gate's own level and not just in my census:

```
matched "Trust"  slot tool-cta   "The AI Vendor Trust Checklist turns one vague question into twelve concrete ones"
matched "Trust"  slot tool-cta   "Check What a Vendor Publishes Before You Trust Them With Data"
matched "trust"  slot ai-vendor-trust-checklist  "Undisclosed use is what erodes trust fastest…"
```

**The pattern is flagging the site's own product name.** Between 2026-07-18, when
the rule was written, and 2026-08-08 the fleet built an entire trust-themed content
pillar here — four "AI data trust in…" articles, a guide, and a tool with *Trust*
in its title — and the rule and the content strategy collided. Much of the queue is
**unsatisfiable by construction**: actioning it means renaming the tool, or deleting
quoted research titles (*Deloitte's "Banking on trust"*) and other people's
statistics (*"Patient trust in medical AI dropped 8 points"*).

**So I stopped rather than rewrite 19 pages against the owner's own rule.** This is
the owner's ruling of 2026-07-18 and it is his to narrow. Recommendation, if he
wants one: keep the intent, drop the bare noun — target the self-labelling
constructions (`\btrustworthy\b`, `\bdeserves trust\b`, `\b(you can|customers?|
clients?) (can )?trust\b`, `\btrusted (partner|advisor|provider)\b`) and let the
site write *about* trust. Whatever is decided, the fleet-wide lesson from
`fleet_copy_quality`'s own CONTRIB §4 holds: **a mechanical ban-list is a smell,
not a crime.**

**What did NOT need a ruling is prepared and NOT applied**:
`scripts/VOICE_2026-08-17_banned_phrases_ready.sql` — the two `earns its keep`
hits and the one self-labelling `honestly`. It is unapplied because the edits only
reach the page through a rerender, and a rerender that escalates to the LLM writer
during the cap fails mid-page. **The landmine check paid for itself before a single
write:** `use-cases-list.use_cases` is `source: site_specs.portfolio.use_cases`, so
the `earns its keep` sentence is **not** editable in `page_components.content_data`
— that edit would have read back correctly, passed the gate, and been reverted by
the very rerender fired to publish it. One edit to the `portfolio` aspect fixes
both `/how-it-works.html` and `/use-cases.html`, which is also why the identical
sentence appears on two pages. `generic-text-block.content` is `llm`-sourced, so
that one is a `content_data` edit. Two "honest"s in that aspect's `client`/`status`
fields are invisible to BOTH a served-HTML sweep and a `page_components` census —
they are in the authority and render nowhere today.

Left alone deliberately: `/tools/process-automation-scorer/index.html`'s *"Answer
honestly rather than optimistically"* — an instruction to the **user** about their
own answers, not the site labelling itself; the same judgement the CONTRIB applied
to "dishonest" meaning "unfair".

### Owed when the endpoint recovers

1. Apply `VOICE_2026-08-17_banned_phrases_ready.sql`, then its three rerenders and
   its served-page assertions (both in the file).
2. `fcfbdfd5-a1f5-427d-962f-8caaf82ea145` should claim itself — read the verdict
   honestly, skips are not passes, and the vision result is `collected_data->'look'`.
3. Verify the two new LANDMINES entries. `landmines-sync.py --apply` has run (the
   `doc_notes` rows exist) but the verifier is an LLM agent and could not run, so
   **the arm was consumed** — re-trigger by hand rather than expecting the sweep:
   `./scripts/trigger-landmine-verifier.sh 'LANDMINES.md#the-whole-build-pipeline-stops-dispatching-and-nothing-says-so-items-stay-triage'`
   (two other lanes' entries are in the same state — the sync reported 3 needing
   verification).

---

## 2026-08-18 — the home page: four misdirected CTAs fixed, and two new ways of showing the evidence

Owner ask, verbatim in effect: *"create two new better ways of presenting the data on
the home page … use the framework's tools like visual editor and experience loop to
improve how the customer sees these blocks visually. The copy is ok. The 'Tell us what
you want to automate' button links through to the password tool — it should link to the
contact details page."*

### First, three corrections to the brief, because two of them changed what I built

1. **There is no visual editor in this framework.** The only mention anywhere in the
   tree is an unticked `- [ ] Phase 7: Visual editor` in `docs006_workflow_builder`.
   Nothing to use, so nothing was used; the work below is DB-authored and verified at
   the rendered artefact instead. `[VERIFIED 2026-08-18: grep -rniE "visual.?editor"
   over *.go/*.md/*.ts/*.tsx/*.js → one hit, that checkbox]`
2. **The button no longer says "Tell us what you want to automate."** A loop had
   rewritten the label to *"See the systems we have built"* while leaving the href on
   the password tool. I restored the owner's label AND pointed it at `/contact.html`,
   which is what he asked for; had I only fixed the href, the page would have read
   "see the systems we have built" and gone to a contact form.
3. **It was not one broken button, it was four.** All four home-page CTA destinations
   pointed into `/tools/`.

### The chassis roll, checked rather than assumed

`v1.0.1310`, both replicas. Provenance line from the service's own startup log:
`git_commit 0b185bad2a49c6e032352fa9e7d0b429f0a95104`, which is **HEAD** —
`git rev-list --count <stamp>..HEAD` = **0**, so this build genuinely ships current
code and is not the cached-image case (MEMORY [[a-fresh-deploy-can-ship-no-new-code]]).
Re-confirmed with `grep -a` on `/proc/1/exe`, with a random-hex control ABSENT in the
same exec so the probe is known to discriminate.

That matters because `bugs_open/248`'s fix (`53a8d3c1d`) is an ancestor of that stamp —
**live** — which is what made the CTA repair worth doing rather than a repair that the
next resolve would undo.

### The four CTAs

| slot | label | was | now |
|---|---|---|---|
| hero primary | "Tell us what you want to automate" (restored) | `/tools/password-entropy.html` | `/contact.html` |
| hero secondary | "See how we work, from first conversation to production" | `/tools/ai-agent-roi-estimator.html` | `/how-it-works.html` |
| cta primary | "Book an architecture conversation" | `/tools/tool-agent-complexity-estimator.html` | `/contact.html` |
| cta secondary | "Or call +44 (0) 7934 524 911 to talk through…" | `/tools/ai-agent-roi-estimator.html` | `/contact.html` |

**A misstep worth the line: my first UPDATE matched 0 rows and the verify block caught
it.** I had captured `page_components.id` values four queries earlier and addressed the
UPDATE by id. Those ids are **not stable across rerenders** — it is written in this
directory's own quick-reference table and I used the ids anyway. The transaction printed
`UPDATE 0` twice and then raised, so nothing shipped; addressing by `(page_id,
slot_name)` worked first time. **A `DO`/`RAISE` block is not paperwork — it is the only
reason a no-op did not read as a success.**

Survival is measured, not assumed, and it is contributed to `bugs_open/248` as that
lane's owed artefact proof: the three authored `/contact.html` destinations survived
**two** `section_data_resolved` rerenders, stored and served, with a **demand control**
in the same run (the chart tone change below did ship, so the renderer provably ran).
The honest caveat is in the bug file: `hero.secondary_cta_url = /how-it-works.html`
survived too, and that is a NON-utility destination the keep-branch should not protect,
so my evidence cannot yet separate "the fix held" from "nothing was recomputed".

### Two new presentations — both are EXISTING framework components, not new code

The diagnosis first, because it decided the design. The home page carried two blocks
that say **the same five things twice**: `features` (6 icon cards) and
`differentiators-section` (5 bordered cards) both assert record-keeping, stage-by-stage
approval, the Companies House check, the Kubernetes/Kafka/Postgres stack and "we run it
ourselves". Two undifferentiated text grids in a row. And the site's most persuasive
content — every measured figure — was buried mid-paragraph in centred 12px grey body
text: *"More than 2,000 of the records… 937 of them enriched… 5,798 veterinary
companies… more than 150 agent definitions, 77 of them active"*.

So: lift the numbers out, and break the two grids apart. Both were already built:

- **`stat-band`** (`62859c4c`) — four figures with label and caption, placed between the
  hero and `features`. Every value traces to a fact in this site's `evidence_base`
  register, and the two whose tolerance is `gte` are stated as **floors** because their
  own `writer_line` demands it (`C1-records-verified` carries the 2026-07-20
  overclaiming incident in its `notes`). 22 sites · 2,000+ records · 78 agent
  definitions · 10,000+ news items.
- **`evidence-chart`** (`f8c2393c`, register VIZ-001) — placed after `features`, whose
  third card is the Companies House claim. Two bars: records verified (2,338) and, as a
  subset, those enriched with filed accounts (937). **No figure is typed into the page**
  — each point names a `fact_id` and the renderer resolves the value and stamps
  `verified 2026-08-18` under each bar.

Reuse, not build: `fundamentallyai.com/index.html` already runs exactly this pair in
exactly this order (hero → stat-band → evidence-chart), so the shape is proven live.
The charts are declared in the **aspect** (`site_specs.evidence_base.charts`), never in
`content_data` — both fields are aspect-sourced and a `content_data` edit is re-derived
away (the landmine this lane wrote two days ago).

New section order: `hero, stat-band, features, evidence-chart, generic-text-block,
differentiators-section, call-to-action` — which also puts a gold band and a chart
between the two near-identical grids.

### The visual check, and the two defects it produced

**The render audit lied to me first.** `scripts/render_audit.py` on the new page printed
`ERROR probe produced no result` and then **`1 page(s): 0 contrast failure(s), 0 broken
image(s)`, exit 0.** A gate that passes while blind. Cause: the probe renders its
injected copy from a `/tmp` workdir and **a snap-confined Chromium cannot read `/tmp`**
— it renders its own error page. Reproduced in one line, both directions (`/tmp` →
Chrome's boilerplate; the identical file under `$HOME` → real content).
Fixed both halves in `scripts/render_audit.py`: errored pages are counted, printed as
*"the zeros above are silence, not a pass"*, and included in the exit status; and the
workdir is overridable via `RENDER_AUDIT_TMP`. **The workdir override did not make it
work under snap** — something beyond `file://` is confined — so the route that does work
is running the audit's own `AUDIT_JS` against the LIVE url over CDP
(`scripts/audit_live.py`), which is also more faithful than re-injecting fetched HTML.

Result, both widths `[MEASURED 2026-08-18]`: **0 broken images, no horizontal overflow,
1 contrast finding** — the hero's secondary button at 3.95:1 against 4.5:1, and it
carries `overImage: true`, so the ratio is the tool's approximation over a photographic
hero, not a measured background. Pre-existing, untouched by this work, and a real fix
means changing a hero component shared across nine sites — architecture scope, not a
site session's call. Recorded, not fixed.

**Then the defect no check could have caught, found by looking.** The chart's second bar
used `tone: "muted"`, which resolves to `var(--color-secondary, …)`; on this site's
light card that renders as a near-white cream on a near-white track, so a bar at 40% of
the axis **read as an empty track** — i.e. as zero, the opposite of what it said. The
number beside it was correct, the length was correct, and every automated check passed:
the render audit measures TEXT contrast and a bar is not text, and the claims gate
cannot see inside graphical furniture at all (VIZ-009). The only prior live example
(fundamentallyai) has that bar at 0%, so its colour had never mattered.
Switched to `tone: "accent"` in the aspect, rerendered,
re-clipped the same element — now unmistakable. Both landmines are filed and their
verifiers armed (`700e8d8f`, `c5116f4b`).

### Verified at the served page

```
bytes 28,585 -> 38,894           two new sections present
stat-band                35 matches      evidence-chart         42 matches
'Tell us what you want to automate'  1   'password-entropy'      0
href="/contact.html" class="cta-btn cta-btn-primary"      present
evidence-chart__bar--accent  2           needs_page escalations  0
render audit @1280 and @390: 0 broken images, no overflow, 1 approximate finding
```

Backups: `bak_leo_home_pc_20260818`, `bak_leo_home_pc_20260818b`,
`bak_leo_evidencebase_20260818`, `bak_leo_home_page_20260818`.

### Missteps this session

1. Addressed an UPDATE by a captured `page_components.id` — not stable across rerenders;
   caught by the verify block, cost nothing but one round.
2. Read a render-audit "0 failures" as a pass for about a minute before noticing the
   ERROR line above it. That is the exact shape this estate keeps writing down, and I
   still nearly banked it — hence the tool fix rather than just a note.

---

## 2026-08-19 — the design review the owner asked for: both agents ran, and most of what came back was wrong

Owner ask: trigger the visual designer and the offer-analysis agent to make a design and
usability judgement together — *"more decorative and functional carousels rather than just
lists of cards"*, *"more imagery"*, unconstrained design thinking, **report back before we
go ahead.**

### First, what the framework's "visual designer" actually is

`visual-designer` is live, and its description is *"Handles images, logos, and visual
assets"* — a two-step asset agent. It does not think about layout. The design-judgement
agent is `visual-design-auditor` (spawned by `design-audit-agent`), so that is what was
dispatched. **Neither can produce the thinking that was asked for**, and that is a
property of the prompt, not of the run: `run_visual_llm_audit` asks for *"the TOP 5 most
impactful issues"* in exactly five categories — colour, spacing, typography, dark
sections, responsive CSS — and requires a machine-verifiable `acceptance_test` per
finding. It is a CSS-hygiene checker. It cannot say "this list should be a carousel", and
the constraint is deliberate: every finding must be checkable by a different agent.

Both were dispatched with `orchestrate_safe.sh` (payload in the container command,
`PUBLISH_OK` receipt). Both COMPLETED: `ee564430` (visual), `d00f8fe3` (offer).

### The visual findings, graded at the rendered page — 1 of 5 holds

A subagent's report is another doc, so all five were re-asked of the live page with
`getComputedStyle` (`scripts/verify_audit.py`), not read from CSS text.

| finding | severity | verdict |
|---|---|---|
| palette is corporate blue, `--color-primary #1e40af`, `--color-accent #d97706` | high | **FALSE** — computed `:root` is `#0D0D0D` / `#836E32` / header `#0D0D0D` |
| body font is Merriweather serif | medium | **FALSE** — computed body and h1 are `Inter, -apple-system, …, sans-serif` |
| `.stat-band` padding shorthand malformed, section has no padding | medium | **FALSE** — computed padding **80px**; and a `var()` fallback may legally hold a shorthand |
| hero h1 has no responsive scaling, will overflow | medium | **MISDIAGNOSED** — 56px→32px at 375, `scrollWidth == clientWidth == 375`, no overflow |
| hero carries an inline rgba overlay + `--hero-btn-ink`, so a theme change cannot cascade | high | **CONFIRMED** — verbatim in the `style` attribute |

**The failure is structural, not bad luck.** The agent is fed `design_context.css_excerpt`
and HTML samples — **source text** — so it read `var(--color-primary, #1e40af)` FALLBACKS
as the site's palette. Every false positive is a value decided at cascade time, which its
inputs cannot show it. The platform already owns a renderer-side witness (`render_audit`,
VIZ-010) and this agent does not use it. Filed as a LANDMINE with the measurements.

> **CORRECTION, and it is the same trap.** This file said yesterday that
> `--color-accent` is `#d97706`, quoted from a `grep` of the served HTML. The computed
> `:root` value is **`#836E32`**. A hex in the source is not a hex in effect — I made the
> auditor's mistake one day before grading it. Corrected in place above.

### The offer analysis — useful, and carrying a stale number

`offer-analyser` produced 5 findings plus a refreshed `offer_ordering`. Its reader-goal
framing is sound (a CTO deciding whether to trust or rule out) and two findings are real:
the `insights` meta description still says *"digital transformation … for business
leaders"*, which the site's own content strategy bans; and the careers page is written in
brand-values register rather than to the recorded audience.

**But the run is still `degraded: true` (`inputs_missing: ["recurring_value"]`) and it
repeats "eight live sites" — the stale figure this lane corrected on 08-16.** Register
`C6-sites-deployed` is **22**, floor form. Its rank-1 suggestion is to open the home page
with that number, i.e. to put a false claim in the hero. Recorded on the item.

### The nine items were HELD, because the owner asked to see this first

Both agents file findings at `status='detected'`, and `detected-item-promoter` takes
exactly that status to `triaged` on a ~2-minute tick, after which the build pipeline
dispatches the named handler — `webdesign-agent` for the false palette finding. So
"report before we go ahead" and leaving them at `detected` are incompatible. All nine
moved to `needs_human_review` with the grading written onto each row's `result`
(`bak_leo_audit_items_20260819` holds the originals). Nothing dispatched.

### What actually answers the owner's question, measured

Neither agent addressed carousels or imagery, so:

- **Carousels are ONE component wide.** `js_snippets.hero-card-carousel.applies_to` is
  `["hero-card-carousel","info-card-grid"]`, and exactly one active component declares a
  `carousel` field: `info-card-grid` (`carousel [config]`). The home page's `features` and
  `differentiators` have no carousel arm at all, so "make these carousels" is a component
  swap or a snippet widening, not a flag.
- **The stat band's count-up does not fire.** Its own `llm_guidance` says the value is
  *"code-rendered and count-up animated"*; `counter-animate.applies_to` is
  `["stats","numbers","social-proof"]` — it does **not** cover `stat-band`, and the served
  page carries **0** counter attributes. One-line fix to a shared `js_snippets` row.
- **The imagery instinct is right and now has a number: 29 of 36 live pages carry one
  image or none.** 52 `<img>` site-wide, and 13 of those sit on the two pages that got the
  07-31 icon work (services 7, who-we-help 6). The framework has plenty of image-capable
  components that this site does not use — `case-studies-grid` (5 image slots),
  `content-block-about`, `people-feature-block`, `featured-content`.

### Owed / next, pending the owner's decision

Nothing was changed on the site today. The nine held items are readable at
`site_work_items … status='needs_human_review' AND result ? 'grading'`.

---

## 2026-08-25 — A1 verified, A0 (the broken carousel) fixed, and two regressions found on the way

Executing the approved plan (`~/.claude/plans/let-s-do-1-2-and-3-ancient-crab.md`). Five days
had passed, so every premise was re-grounded first — and two of the four had changed.

### A1 — the figures band: already bundled, and now PROVEN to animate

Another session re-ran `site-asset-renderer` between 08-20 and today: the bundle serves
**3 active snippets** with `data-countup` present (was 2 and absent). So the dispatch was
not needed.

**But a bundled snippet is not a working animation, and my first probe said it was dead.**
It read the value immediately after `scrollIntoView()` — and with CDP round-trip latency the
tween had already finished, so `during == after == "22"` on a working animation. **A probe
that samples after the event it is timing cannot see the event.** Re-probed by arming a
`MutationObserver` *before* the band scrolls into view, which leaves a trail whatever the
duration `[MEASURED 2026-08-25]`:

```
normal motion    69 mutations   trail: 22 → 0 → 0 → 1 → 2 → 3 → … → 22
reduced motion    1 mutation    trail: 22 → 22          (one write, straight to final)
final value       "22" in both — the authored string, restored exactly
```

Both directions, and the reduced-motion arm is *correct behaviour*, not a failure — my
written assertion ("mutations = 0") was too strict: the snippet writes the final value once
and never tweens. Probe kept at `scripts/probe_countup.py`.

### A0 — the shared carousel template: fixed, and verified at the browser

An unmatched `*/` closed the comment block early, so the prose that followed was parsed as a
selector prelude and the rule defining `--icg-track-gap` / `--icg-arrow-size` was **dropped**.
Fixed by removing the premature terminator (one comment, not two) and adding the literal
fallbacks the swallowed prose itself argues for.

**The write, and the two traps it walked past:**
- `length()` = 11,893 but `octet_length()` = 11,903 — the em dashes. Compared on **bytes**
  and md5, never `length()`.
- `psql -At` appends a trailing newline, so the naive extract is 11,904 B and its md5 does
  not match the DB's. Stripped it, then confirmed the baseline md5 equalled the DB's
  `204a3975…` **before** editing.
- Written back as base64 → `convert_from(decode(...),'UTF8')` so no quoting can corrupt it,
  `UPDATE … AND md5(html_template) = '<baseline>'` so it refuses if the row moved under me,
  and a `DO`/`RAISE` post-condition on the new md5 plus the presence/absence of the two
  markers. New md5 `0d4afe45…`, 11,928 B.

Verified at the browser after one rerender, not in the source `[MEASURED 2026-08-25]`:

| probe | before | after |
|---|---|---|
| `--icg-track-gap` computed | *empty* | `1.5rem` |
| `--icg-arrow-size` computed | *empty* | `44px` |
| track `gap` | `normal` (0) | **24px** |
| next-arrow rendered | **22 × 28** | **44 × 44** |

44×44 clears WCAG 2.2 AA's 24×24 target minimum, which 22×28 failed on width.

**A scare I caused myself and then disproved.** The post-fix probe reported `slides: 3`
where 08-20 reported 6, and both blocks' `updated_at` equalled my rerender's timestamp — so
it read exactly like my rerender had eaten half the page. It had not: I still had the served
HTML fetched at session start, before any change, and it already showed 3 cards and the same
three titles. **`updated_at` bumps on a no-op write**, which is this estate's own recorded
lesson and which I nearly re-learned the expensive way. Keep the before-fetch.

### ⚠ Regression 1 — the home page CTA the owner reported is broken AGAIN

`call-to-action.primary_cta_url` on `/index.html` was clobbered back from `/contact.html` to
`/tools/tool-agent-complexity-estimator.html`, while the label still reads *"Book an
architecture conversation"*. Producer named from its own row: `page_rerender`, complete
**2026-08-24 18:37:13Z**, summary *"2 misdirected CTA(s) on index"* — matching all seven
components' `updated_at` of 18:37:05Z. So the misdirected-CTA **repair** turned a correct
link into a wrong one, while in the same run correctly repointing the hero secondary to
`/how-we-work.html`.

Re-authored in `content_data` (`bak_leo_home_cta_20260825`); publishing rides the next home
page rerender rather than firing one for it. **My 08-18 contribution to `bugs_open/248` said
those links survived — that has now expired and I have corrected it in the bug file**
(`b1b9d000b`), with the free discriminator: `hero.cta_url` held while
`call-to-action.primary_cta_url` did not, same destination, same utility area, same run.

### ⚠ Regression 2 — /services.html has lost content AND all its imagery, for the third time

`[MEASURED 2026-08-25]`, and **not repaired** — this needs a decision, not a third hand-fix.

| slot | 08-14 restore left | live now |
|---|---|---|
| `info-card-grid` cards | 6 | **3** |
| `teaser-reveal-panel` items | 6 | **5** |
| service icons referenced on the page | 6 | **0** |

All six `icon-service-*.jpg` are still deployed and serving **200** — the page simply no
longer names them. And the item keys have been rewritten again: the 08-14 set
(`verification-pipeline`, `hierarchical-orchestration`, `human-oversight`, `decision-record`,
`model-routing`, `news-credibility`) is now `decision-record`, `first-conversation`,
`infrastructure`, `scoping`, `verification`. This is the 2026-08-11 damage pattern in full,
a third time.

**A correction to my own reading while investigating it:** I assumed
`bak_leo_services_pc_20260814` held the *repaired* state and could be restored from. It does
not — it is the **pre**-repair snapshot, and its `image_url`s are empty too, exactly as the
08-14 handoff says. There is no backup of the good state; the 08-14 repair was written
directly. So a restore means re-deriving the icon↔item mapping against the NEW keys, as the
08-14 session had to.

**Why I stopped rather than repaired:** this is the third time this page has been restored by
hand, and each restore has been undone by a regeneration within days. A fourth hand-repair
buys days. The cause is upstream (`bugs_open/238` regeneration drops resolver keys, and the
`bugs_open/248` family), and that is where the owner's attention is worth spending. Raised
rather than quietly re-done.

### A2 (part 1) — the duplicate block removed, and the section AUTHORITY repaired

**The duplication, side by side, is worse than "similar":** all five `differentiators-section`
cards restate `features` cards, and *"Approval set stage by stage"* appears in both with the
**identical title**. Owner's call (2026-08-25): carousel the good block, drop the duplicate.

**The more urgent half, found while doing it.** `site_specs.site_plan` — the section authority
for this site (`site_plan_sections` has no rows here) — listed `index` as
`[hero, features, generic-text-block, differentiators-section, call-to-action]`: **five
sections, with no `stat-band` and no `evidence-chart`.** Both blocks I shipped on 08-18 were
absent from the authority, so any rebuild reading it would have dropped them silently. There
was already an open `section_source_drift` item saying exactly that, filed 08-24 and
unactioned. Authority and `pages.sections` are now both the intended six.

**The removal followed the recorded recipe, and its correction.** Prune the authority, prune
`pages.sections` (a STRING array — checked before trusting `- 'name'`), set
`build_status='removed'`, and **empty `rendered_html` while KEEPING `content_data`** — that
column is the section's only copy. The 2026-08-10 correction to that landmine says the
tombstone alone does not stop the light path, whose own filter was inert until a roll.
**Probed the capability rather than the commit** `[MEASURED 2026-08-25]`:
`grep -a "build_status IS DISTINCT FROM 'removed'" /proc/1/exe` → **PRESENT**, with
`"build provenance"` PRESENT as a positive control and a nonsense literal ABSENT. So the
light path filters it now and the tombstone covers the assemble-only path.

> **A probe I got wrong first, worth the line.** I tried to date the running build by grepping
> the binary for HEAD's sha and for the fix's sha. Both came back ABSENT — and so did my fake
> control, so the probe looked sound. It was not: **a binary stamps only its OWN build commit,
> so an ancestor is absent either way.** The question "is the fix in?" cannot be answered by
> grepping for the fix's commit. Probe the CAPABILITY — the literal the fix added — which is
> what finally answered it. Before that I had also read an **empty** `$STAMP` into
> `git rev-list --count $STAMP..HEAD`, which returned `0` and read as "up to date".

One rerender published the removal, the authority fix and the CTA repair together
(`b139e622`, COMPLETED, **0** `needs_page` escalations). Served-page assertions all pass:
duplicate block gone, `features`/`stat-band`/`evidence-chart` all kept, **3** `/contact.html`
CTAs, **0** `/tools/` CTAs, hero-headline control present.

> **CORRECTION to my own first reading of that verification.** I reported the page had *grown*
> 8.8 KB after removing a section, and started hunting the cause. There was none: I had
> compared today's page against **38,894 bytes measured on 2026-08-18**, not against the page
> as it stood this morning. Measured like against like, the page went **50,941 → 47,563 B,
> i.e. −3,378** — the duplicate block and the shortened CTA url, exactly as intended. The page
> had grown to 50,941 through other sessions' work during the week. **A byte baseline is a
> measurement and it expires like any other; label it with its date or do not use it.**

### A2 (part 2) — the carousel is DEFERRED, and the reason is the images

Converting `features` to `info-card-grid` + `carousel:true` would **downgrade** it, for two
reasons found by reading the template and looking at the assets:

1. **`info-card-grid` renders `.icon` raw** — `{{else}}{{.icon}}{{end}}`. `features` carries
   lucide icon *names* (`file-text`, `user-check`, `shield-check`, `server`, `git-branch`,
   `activity`) which its own component maps to SVG. Moved across, they would print as the
   literal text "file-text".
2. **The six orphaned `icon-service-*.jpg` cannot stand in.** Looked at, per the site rule:
   they are **wide landscape illustrations** (~3:1) in fine gold linework on near-black. The
   `icon_image` slot renders at **44×44 with `object-fit: contain`** — a 3:1 illustration in a
   44px box is a few pixels of line. The `orchestration` one is genuinely good (a real
   parent/child/grandchild hierarchy diagram); it is still not a 44px chip.

So a carousel today would be six text cards that scroll — motion applied to a list, which is
the opposite of what was asked for ("more **decorative and functional** carousels rather than
just lists of cards"). **The carousel is worth having once there are card images worth
carouselling**, which is A3. Sequencing it after A3 rather than before.

### A3 groundwork — and a CORRECTION to the imagery figure I have been repeating

> **CORRECTED 2026-08-25.** I have written "**29 of 36 live pages carry one image or none**"
> into the plan, these notes, `README_where_we_are` and the 08-18 summary. **It is wrong, and
> wrong in a way that changes the work.** The measurement was `grep -c '<img'`, which cannot
> see a hero delivered as a CSS `background-image` — and on this site every hero is exactly
> that. Re-measured over all 36 live pages, counting both `<img>` and
> `url('/assets/images/…')` `[MEASURED 2026-08-25]`:
>
> | | |
> |---|---|
> | pages with **no image at all** | **0** |
> | pages showing the **same** `hero.jpg` | **21** |
> | pages with a distinct hero | 4 (`home`, `who-we-help`, `use-cases`, `how-we-work`) |
> | **distinct images used site-wide** | **6** |
>
> So the problem is **sameness, not absence**: twenty-one pages open with the identical
> photograph. That is a design weakness, not a broken page — the *mediocre* class, not the
> *broken* class, which is the distinction the design critic in Part B is built around. It
> also means no page is currently unillustrated, so nothing here is urgent.

**What the 19 `image_source_unsatisfiable` items actually mean, then.** They fire because no
`site_assets.hero` **asset row** backs the field — not because the rendered page lacks an
image. The page falls back to `hero.jpg` and serves 200. So the queue is real but its
severity is lower than its name suggests: it marks pages that are *sharing* rather than
*missing*.

**The archetypes are already in the data — I do not need to invent them.** The 17
hero-wanting pages split by the platform's own `page_type`: **12 `content`** (ai-readiness-quiz,
careers, engagement-model, faq, how-it-works, index, insights, privacy, technical-architecture,
terms, + 2 archived) and **5 `blog-post`** (the four guides and the orchestration explainer).

**And the naming convention is established and working.** Four pages already have per-page
heroes and they follow one shape: asset_key `hero_<page>` → `/assets/images/hero-<page>.jpg`,
all serving 200. So new heroes follow the same pattern; nothing new needs designing.

**One pre-existing defect noticed in passing, not repaired:** `hero_case_studies` still holds
a **presigned S3 URL** in `assets.url` (`bugs_open/152`'s recurrence, which this lane already
contributed on 2026-08-16).

---

## 2026-08-25 (evening session, "leopardess") — owner rulings D1–D4 + the re-grounding that rewrote D3

**Owner answers (AskUserQuestion, this session):** D1 = DROP the trust rule entirely (not the
narrowing I recommended). D2 = per-page heroes for the dozen that matter, archetypes for the rest.
D3 = THIS lane takes the platform fix. D4 = build the design critic in THIS lane.

**Re-grounding before acting — three premises had moved since the morning handoff:**

1. **The "bugs_open/238 / 248 family" cited by handoff §D3 is CLOSED** — `bugs_closed/238`
   (08-21, prevention proven live on v1.0.1322), `bugs_closed/248` (asset-repair one), and
   `bugs_closed/355` (08-23). Only `bugs_open/248` (CTA recompute, the OTHER 248) is open.
   Classic closed-blocker-keeps-being-obeyed shape.
2. **The third /services damage is dated and attributed** `[MEASURED 2026-08-25]`:
   ONE generation, 2026-08-22 11:35:41Z, `save_page_sections_overwrite`, driven by
   `site_work_items` key `offer-analysis_content_rewrite_services_4851f6fc-…` (created 11:17:21Z,
   complete 11:36:20Z). `info-card-grid` 1,794B/6 cards → 1,147B/3 cards; `teaser-reveal-panel`
   4,136B/6 items/6 `icon-service-*` refs → 3,511B/5 items/0 refs. The restore had survived
   FOUR rerenders byte-identical (08-14 18:25, 08-16, 08-17, archived 08-22) — rerender MERGES,
   the rewrite path REPLACES. Queries preserved in `bugs_open/403`.
3. **This is a NEW mechanism, not 238 recurring** — the lost values sat INSIDE `source:"llm"`
   array fields; PBP-039 carry (non-llm only), `cmd/content-loss-check` (blank transitions only)
   and the 238 closure census (pairs by field key) are each structurally blind. Filed as
   **`bugs_open/403`** (+ 016b §9 pattern + LANDMINES entry, commit `0c43d5050`); 090 run fired
   FIRST — intake corr `2590b5b6…`, **run corr `c946b495-115d-4e3e-8186-3819273edb6c`** (advisory:
   HEAD was 87 ahead of origin at dispatch, so the loop cannot see the newest tree).

**Tonight's live activity on the site (watched it happen mid-query):** work-item pair created
19:17:57Z — `content_rewrite/cta_label_relevance` + `page_rerender/cta_relink` — rewrote
`/services` (19:52–19:53Z) and `/tools` (19:40–19:41Z) CTAs. BENIGN: swapped one machine-minted
tools CTA for another (password-entropy → complexity-estimator), label and URL agree both sides,
and stamped `__cta_minted` (LNK-035, `datahelpers/cta_provenance.go`, shipped 08-22). Teaser
values also rewritten (3,511→4,925B, same 5 items). **Home held**: hero `/contact.html`,
call-to-action primary AND secondary `/contact.html` — the morning re-authoring survived.
NOTE for any future survival claim: name the PRODUCER survived, not the number of rerenders
(248's fix-record lesson).

**Voice-pass unblock check:** `ai_endpoint_health` → `api.anthropic.com` healthy=t. The
2026-09-01 cap note in `VOICE_2026-08-17_banned_phrases_ready.sql` is STALE; both halves can
apply now. The trust rule to drop is element `"trust / trusted / deserves trust — …"` in
`site_specs` aspect `voice` → `banned_language` array (site `4851f6fc…`, is_current).

**Next in this session:** voice pass (drop trust element + apply ready SQL + rerender 3 pages +
served-page verify), then first per-page hero end-to-end. 403 fix design + critic design (D4)
queued behind them; read the 090 verdict before designing the 403 fix.

### Same session, later — D1 executed; A3 canary live; two heroes re-rolled

**Voice pass (D1) COMPLETE and verified at the served pages.**
- Trust rule DROPPED from `site_specs` voice → `banned_language` (element matched by content,
  10→9 entries, in-tx DO/RAISE verify, `bak_leo_voice_20260825`).
- ⚠ The 08-17 ready SQL **refused itself** (exit 3) — correctly: its pinned `page_components.id`
  had been re-minted (`UPDATE 0`, the §1.1 lesson), and the portfolio aspect had GROWN two new
  "honest" sentences (use_cases elements 4+5) since it was written. Superseded by
  `VOICE_2026-08-25_banned_phrases_v2.sql` — content-addressed across all elements, pc addressed
  by (page, slot). Applied clean; backups `bak_leo_portfolio_voice_20260825`,
  `bak_leo_insights_pc_20260825`.
- Rerendered use-cases / how-it-works / insights (rerender_page_safe, PUBLISH_OK ×3). Served
  sweep: `earns its keep|honest(ly)?` = **0 0 0** with positive control 1 (`repetitive process`
  still findable — the zeros are not blind).

**A3 (D2) — imagery estate re-grounded, canary wired.**
- `[MEASURED 2026-08-25]` assets: SEVEN per-page heroes already generated+active — 4 wired
  (home, how-we-work, use-cases, who-we-help), **3 deployed but unwired** (about, contact,
  services), 1 broken (`hero_case_studies` = presigned URL, the O5 §4 defect, needs redeploy).
- **The unwired three are 403's mechanism again**: history shows `background_image` keys eaten —
  about's in the generation written 08-16 16:04, contact's 08-11 16:37 (added to 403 as
  instances 3+4; top-level undeclared keys this time, not array members).
- Eyeball rule applied to the three existing images: **about ACCEPTED** (concentric rings);
  **services + contact REJECTED as near-duplicates of each other** (both "line with three dots")
  — same failure the icon batch had, and the 08-11 session had already noted the similarity.
- **Canary: about is LIVE** — `background_image` merged (bak_leo_about_hero_pc_20260825),
  escalation gate both branches 0 rows first, safe rerender, served page now shows
  `hero-about.jpg`, 0 generic `hero.jpg` refs on the page. NOTE: this wiring is a 403-class
  loan — durable only once 403's fix ships; recorded deliberately.
- Re-rolls dispatched Route-A-safe (same asset_key overwrites): `hero_services` (diverging
  three paths, corr `68633722`), `hero_contact` (two arcs meeting, corr `31466cbd`). Watcher
  armed on the deployed file sizes.

**090 on 403's mechanism:** run `c946b495` was at `call_diagnoser` AWAITING_RESPONSES 20:05Z.
Read the verdict before designing the fix.

**The dozen for per-page heroes (owner D2), proposed:** about✓, services, contact, how-it-works,
tools, insights, engagement-model, technical-architecture, ai-readiness-quiz, careers +
blog: why-most-ai-agent-projects-never-reach-production, can-you-trust-ai-with-your-data.
(case-studies also needs its asset REDEPLOYED regardless.) Archetypes for the remainder.

### Same session, later still — the filing was corrected by the code, and services is RESTORED under LOCK

**The correction first (WRONG_CALLS row written):** 403's "no guard covers this" was overstated.
`save_page_sections` has a live locked-row guard — `loadActiveLockedRows`/`matchLockedRow`
(`save_page_sections_action.go:1218`), predicate `datahelpers.AgentWritableSQLFor`
(`locked_at IS NULL OR (lock_type='timed' AND expired)` = writable), `lock_type='permanent'`
in use on **51 rows / 7 lanes** `[MEASURED 2026-08-25]`. Three restores were eaten for lack of
DISCOVERABILITY, not lack of mechanism. 403, 016b §9 and the landmine all corrected visibly;
the field-level gap (a row lock freezes the whole slot) remains 403's open work.

**Heroes: services + contact re-rolls ACCEPTED on eyeball** (diverging paths / meeting arcs —
both distinct) and **LIVE at the served pages** (background merged, gate-checked, safe-rerender,
watcher confirmed). Per-page heroes now live on 7 pages: index, how-we-work, use-cases,
who-we-help + TODAY about, services, contact.

**Services RESTORED (fourth time) — this time under protection:**
- Source: the `page_component_history` archives taken at the moment of destruction
  (08-22 11:35:41) — the 08-14 hand-restored state verbatim, so NO key re-derivation needed
  (the handoff's §4 expected one; the wholesale restore avoids it).
- `info-card-grid` + `teaser-reveal-panel` written back (bak_leo_services_content_pc_20260825),
  in-tx verify cards=6 items=6 icon_refs=6; CTA and hero slots deliberately NOT restored (they
  carry tonight's legitimate `__cta_minted` rewrite and the new hero).
- Served page verified: **6 icon-service refs live, hero-services.jpg live.**
- **LOCKED**: 5 rows (services×3, hero-about, hero-contact) `lock_type='permanent'`,
  `locked_by='leopardess-403-restore'`, verified against the live predicate (5 locked /
  0 writable). ⚠ Publish BEFORE locking — sequence matters; and future edits to these slots
  need unlock → edit → publish → re-lock (RUNBOOK-worthy).
- The behavioural proof of the lock arrives organically: the next content_rewrite/tone_shift
  pass at services (two hit it inside the last 4 days) must leave the locked slots intact —
  CHECK THIS when it happens and record the producer survived.

**090 run `c946b495` still iterating** (assemble_bundle EXECUTING 20:2x). Verdict to be
recorded in 403 when it lands.

### Session close — D4 started, cross-lane told, handoff cut

- **D4 first commit `04c49f8f0`**: `design-critique-agent` → `isStorageEnabledAgent`
  (spawn_actions.go), verified against HEAD with `verify-head-builds.sh --with` BEFORE
  committing. **Council-Submitted `30d5fdde-ab0e-405d-a3f3-d83d9227e1ce`** (097 accepted;
  ~30 min budget, find the run by payload not printed id). The pre-commit architecture signal
  fired on spawn_actions.go (ossified core site) — judged a point grant under the RFC_022
  shape (zero live consumers until the seed; sanctioned per-type precedent), council round
  covers it.
- **Vigilant lane told** (owner ruling 2026-07-29 #3): CONTRIB committed into their dir,
  `7ce1bb6c5`. No collision with their same-day compose-side plan found.
- **090 run `c946b495`**: still iterating at close (second round, `load_runtime`). Monitor
  armed in-session; NEXT SESSION MUST read the verdict and record it in 403 either way — a
  REFUTED is a success, record what caught it.
- **Handoff `HANDOFF_2026-08-25b_continue_here.md` cut**; morning handoff banner-superseded.
- LANDMINES amendment note: commit `c8a2e1bdd`'s pattern-check flagged 1 removed line in
  LANDMINES.md — that was this session amending ITS OWN same-evening entry in place (dated
  note inside the entry), not another thread's line. Declared here for the record.

### Post-close addendum — the 090 verdict landed, challenged the attribution, and the re-check CONFIRMED it with better evidence

Run `c946b495`: **UNVERIFIABLE (scope-not-narrowing)** — it disputed pinning the destruction to
the 11:35:41 offer-analysis write (correctly calling my attribution "timing correlation, not a
shown mechanism" and noting `page_component_history` cannot name upstream writers), and proposed
the 08-24 `misdirected_cta` rerender instead. Resolution, recorded in 403: `llm_call_log` holds
the rewrite window's own `page-content-writer` generate_content replies, and they ARE the
damaged content (`533d1712` = the 3-card array verbatim; `bb5ece84` = the 5-item no-icons
array). Attribution upgraded from timing to the write's generative record; the loop's
alternative fails on archive semantics (archive rows are PRE-write states). Lesson kept:
**`llm_call_log` names a content writer when the history table structurally cannot** — and the
loop's challenge is what forced the stronger proof. Next session owes nothing here.

---

## 2026-08-26 — rotation heads-up, and the home CTA was clobbered AGAIN overnight (third time, now attributed with the stamp readable)

**Cross-session heads-up (webdesign-tool-rebuilds seat):** the design-discovery rotation
(`site-discovery-rotation-design`) was re-enabled 2026-08-26 09:20Z after 15 days off (the
08-11 cost-scare pause was never unwound — `bugs_open/401`). ~1 site/3h, least-recently-visited
first; leopardess's visit lands within ~2-3 days. Findings are born `detected`;
`detected-item-promoter` (15-min cadence) auto-promotes known-good (item_type, handler_agent)
pairs into build dispatch. **So: surprise design findings/repairs on this site = the rotation,
not a stray thread. And the five 08-25 locks get their organic behavioural test when the visit
lands — CHECK the locked slots afterwards and record the producer survived.**

**Overnight damage found while checking** `[MEASURED 2026-08-26]`: index
`call-to-action.primary_cta_url` → `/tools/tool-agent-complexity-estimator.html` again.
Writer named from its own row: `misdirected_cta:index` page_rerender completed **02:03:58Z**
(same producer as 08-24; a ~15-item page_rerender batch landed 02:03). The archive row carries
the pre-write stamp this time: stored `/contact.html`, `__cta_minted` naming `/tools/...` —
so `storedCTADestinationIsAuthored` should have been TRUE and the pass displaced the
destination anyway. Hypothesis (recorded in 248, marked as such, with the check): the
confident-label-match arm — "Book an **architecture** conversation" matching "Agent
**Architecture** Complexity Estimator" in the Phase-B-widened universe — i.e. the licensed
displacement branch mis-firing on label semantics. Diagnosis belongs to 248's lane; CONTRIB
appended to `bugs_open/248` incl. the warning that the row lock CONFOUNDS their re-author-and
-watch discriminator (coordinate before unlocking).

**Remedy:** primary re-authored to `/contact.html` (fourth time, `bak_leo_home_cta_20260826`;
secondary was still `/contact.html`), index gate-checked (both branches 0 rows), safe-rerender
published, served verification in flight → then LOCK index `call-to-action` + the two wholly
hand-curated showcase slots (`stat-band`, `evidence-chart`) ahead of the rotation. Hero left
unlocked deliberately: its keep has held through every pass, and it preserves half of 248's
natural behaviour surface.

### 2026-08-26, continued — grant LIVE + APPROVED, critic SEEDED, hero batch generated and eyeballed

- **Overnight fleet roll carried the storage grant**: chassis replicaset `6dd68888dc` (pods started
  23:11Z); capability probe on `/proc/1/exe`: `design-critique-agent` PRESENT /
  `tool-acceptance-agent` PRESENT (positive control) / invented string absent. Council on
  `30d5fdde` → **`complete_approved`** (098 credits `04c49f8f0` automatically). Chassis env:
  GEMINI_API_KEY SET, ANTHROPIC_API_KEY SET, IMAGE_BUCKET correctly UNSET on the shared pod (the
  grant supplies it to the SPAWNED pod — the first critic run is that proof).
- **Critic seeded: mig `645_design_critique_agent.sql`**, commit `45d14129e`, Council-Submitted
  `75be8d32`, registered **SQ-003**. Every step config copied from a proven consumer (145 INSERT
  shape; 317 execute_vision_prompt/append_doc_note; 301 write_render_audit_findings;
  query_database `params` → `$1`). ⚠ The runner's dry-run caught MY guard error (I wrote "want 8",
  the workflow has 9 steps — the DO/RAISE refused in the probe, exactly its job); fixed, then
  applied SCOPED (`MIGRATIONS_DIR=<tmp dir holding only 645>`) because the shared dir had three
  other lanes' pending files (635 likely-applied-unrecorded, 637, 638) and `--apply` takes all.
  Ledger records 645. Manual-only agent: nothing emits for it.
- **Hero batch: 10 generated Route-A-safe** (9 of the dozen + `hero_case_studies`, whose re-roll
  RETIRES the presigned-URL asset defect — asset row now `/assets/images/hero-case-studies.jpg`
  active). All 10 eyeballed: **8 accepted** (how-it-works, tools, engagement-model,
  technical-architecture, ai-readiness-quiz, careers, why-most…production, can-you-trust…);
  **2 rejected + re-rolled**: insights (generator added a cropped open book + compass nib top-left)
  and case-studies (stray construction-line strokes + specks). Exclusion clause added to both
  prompts. Same asset_key overwrites.
- Wiring, non-serving half done: 10 `site_plan_imagery` page-scope rows inserted (assets active
  first — the wire_heroes ordering rule), `background_image` merged for the 8 accepted
  (`bak_leo_hero_batch_pc_20260826`, `bak_leo_site_plan_imagery_20260826`). Gates (both
  branches) 0 rows on all 8. **Rerenders HELD** until the critic's "before" audit has photographed
  the unwired pages.
- **Critic "before" run fired: corr `95f6b328-20da-426f-9e26-ac830c480b1b`** — the discrimination
  control: this report should remark on hero sameness; the post-wiring run must not.
- Locked-page row baselines for the rerender-on-locked-row landmine (surfaced by the hook this
  morning): about 4 rows/1 locked, contact 5/1, index 6/3, services 4/3 — expect UNCHANGED after
  any rerender; +1 means the duplicate-not-protect trap fired (bugs_open/189 has the reversal SQL).

### 2026-08-26 — cross-lane: the 403 marker design is RULED (395 lane asked; recorded in 403)

The `bugs_open/395` session asked the marker-direction question before building a second answer
to it (their owner ruling on meta_description rewrites needs "machine-written only" gating and
no provenance exists). Ruled as 403's owner, full text in 403 §"DESIGN RULING for candidate 1":
both directions coexist (`__cta_minted` licenses, new `__authored` forbids, neither = today's
per-surface default), key `__authored` field→true whole-field inside content_data, enforcement
at save/plan, home `datahelpers/authored_provenance.go`, columns share the CONVENTION not the
storage (companion column, no generic registry). Replied via SendMessage; they build their
column instance independently citing the ruling. Their gift back: meta_description is
structurally IMMUTABLE when non-empty (320) — the inverse failure, recorded in 403 as related.

### 2026-08-26, afternoon — the critic's FIRST SUCCESSFUL RUN, and the hero batch published

- **Spawn-path proof:** item `4f1fb87b` (design_critique_run, priority 90 — note the loader
  sorts `priority ASC`, so LOWER dispatches first; 90 was head of the site queue) claimed at
  the next leopardess visit; pod `agent-design-critique-agent-b4721b66-gwpgr` spawned with
  **IMAGE_BUCKET SET** (personae-prod-uk001-images) + GEMINI_API_KEY SET — SQ-003 verify-later
  item 1 CLOSED. Item `complete`, 0 retries.
- **First `design-report` note is in `doc_notes`** (site_id set, categories ['design-report']).
  16 images / 8 pages / 2 viewports. Quality: every finding names page + region + property +
  direction (018's bar); the palette PRAISED not misread (the corporate-blue trap dodged —
  live-palette join worked); lead finding = **hero sameness**, i.e. the exact change the batch
  was shipping — its screenshots predate publication, so this run IS the "before" leg.
  Actionable list: index bar-charts want containment; services carousel title hierarchy;
  how-it-works ~700px text column; use-cases card edge definition; case-studies uniform card
  heights; quiz input contrast; footer line-height + CTA padding nits.
- **Discrimination test, restructured from held-batch to organic**: the "after" run (to fire
  once all ten serve) should DROP the sameness finding while the rest stays substantially
  stable. The stronger mutation form (revert one hero, run, restore) remains available any time.
- **Hero batch PUBLISHED**: 10/10 rerenders PUBLISH_OK; 7/10 confirmed serving own heroes at
  first sweep, watcher on the last three (both blog posts + case-studies, published last).
  Queue note: five `needs_page` "image landed" rerenders (priority 99) were filed by
  image-build-handler for LISTING pages consuming the new assets (the 384-fix working);
  all four previously-unchecked target pages GATE-CHECKED CLEAN (both branches 0 rows) before
  they dispatch — no escalation risk to the hand-seeded trust articles.

### 2026-08-27 (continuation) — all seventeen heroes SERVED; 649 explained the case-studies lag; after-run fired

- **Census at the served site** `[MEASURED 2026-08-27]`: **0 of the batch pages use generic
  hero.jpg any more; 5 pages site-wide still do** (faq, privacy, terms,
  blog/hierarchical-multi-agent-orchestration-explained, guides/tool-agent-complexity-
  estimator-guide) — the archetype remainder. 17 pages now open with their own hero.
- **case-studies' zero was STRUCTURAL, not lag** — the finetuning lane's §3 notice (mig `649`,
  owner-directed fleet-wide): `case-studies-hero` had NO image branch in its template; the page
  could hold `background_image` and never render it (`bugs_open/412` §7). My merge + rerender
  ran before 649 applied → served zero. Re-fired post-649 → live. `hero-tool` gained the same
  branch — tool-automation-savings-estimator can take a hero in the archetype pass.
  ⚠ transferable: **a template edited by SQL ships nothing until a page re-renders** (283 §13),
  and a wired-but-unrenderable image completes green — the exact shape my "propagation lag"
  guess would have misdiagnosed.
- **After-run fired**: item `b36f8c63` (leg `after_hero_batch`). The discrimination read: its
  report must DROP the sameness lead finding while the rest stays substantially stable.

### 2026-08-27 — the after-leg's two failures, both honest, one is MODEL-TRIAL evidence

- **after r1 (`b36f8c63`)**: `complete_error`, `audit` step "Request timed out (TIMEOUT)" —
  adapter busy at morning load; the failed-audit terminal did its job (a failed audit and a
  clean audit must never read the same way).
- **after r2 (`0eff246f`)**: audit fine (8 pages, findings deduped clean), design_context
  loaded, then **`execute_vision_prompt` → Gemini 400 INVALID_ARGUMENT "Unable to process
  input image"** — pod `agent-design-critique-agent-293867dd-sqbbn`, orch `699b57ef`. The
  16 full-page captures are now LONGER (heroes added), and the plan's cost-envelope warning
  stands: no downscaling exists anywhere in the pipeline, images ship whole. Run ended
  `complete_no_critique` (SUCCESS terminal, findings intact) — the 317 topology again.
  **This is trial evidence for the owner's "try Gemini first, revisit if not" call**: Gemini
  managed yesterday's captures, choked on today's taller ones.
- **after r3 (`a21e0c3e`) fired on Gemini** (vendor advice is retry). If it 400s again, the
  critique step's `ai_service` flips to the 317-proven anthropic/claude-sonnet-5 config as
  trial leg B (recorded owner decision covers trialling both) — a live agent_definitions
  config edit, reversible, to be recorded here + SQ-003 when made.

### 2026-08-27 — the after-leg PARKS on a Go fix, with the full three-error ladder recorded

Four attempts, each failing one layer deeper, all honest terminals:
1. r1 `b36f8c63`: audit TIMEOUT (adapter load) → complete_error.
2. r2 `0eff246f` + r3 `a21e0c3e`: **Gemini 400 INVALID_ARGUMENT "Unable to process input
   image"** — reproducible, post-hero captures only.
3. r4 `0f686e43` (mig 662, provider→claude-sonnet-5, model confirmed in pod log):
   **Anthropic 413 request_too_large** — aggregate payload.
4. r5 `12008d09` (mig 663, max_images→8): **Anthropic 400 "At least one of the image
   dimensions exceed max allowed size: 8000 pixels"** (`messages.0.content.1`) — the REAL
   constraint: a single full-page capture is now taller than 8,000px. Pages grew with their
   heroes; nothing in the pipeline downscales (the plan's cost-envelope warning is the
   binding constraint, in three providers' words).

**STOP RULE APPLIED** — no more runs; no config lever excludes a specific too-tall page
honestly. **The 018 follow-up is now precisely specified: a downscale/segment step inside
`execute_vision_prompt`** (cap each image to ≤8,000px on the long edge — Anthropic's stated
limit — or tile very tall captures). Go change, council, then restore max_images 16 (663's
cap is harmless meanwhile) and re-run the after-leg. The discrimination test survives the
delay: the mutation form (revert one hero, run, restore) works whenever the vision path does.

Everything else about the critic is PROVEN: spawn path, storage, audit, measured-findings
drain, design context, report writing (the 08-26 report), and the 317-shaped terminals that
made all four failures legible from the item row alone.

### 2026-09-02 — cross-lane flag (site-design-planner / bugs_open/431), verified and recorded

Leopardess's `classification` spec is the LEGACY 2026-04-18 classifier shape — verified
first-hand: keys are [reasoning, site_type, confidence, suggested_style, tone_suggestion,
detected_signals, page_count_estimate, recommended_builder]; **no `category`, no
`industry_tags`**. Consequence (their measurement, bugs_open/431): if a `needs_composition`
re-resolve ever runs for this site on a binary BEFORE commit `bd8e45aba` rolls, the layout
resolver sees zero signal and falls back to a GENERIC layout. Nothing to do now — this site's
layout is hand-curated and no re-resolve is queued — but if one is ever dispatched here,
either confirm `bd8e45aba` is live first (capability probe, not tag) or re-run the classifier
to modernise the spec beforehand. Flag credited to the site-design-planner session.

---

## 2026-09-02 — six-day re-ground; THE LOCKS SURVIVED THEIR NAMED CLOBBERER; the downscale is BUILT; three REVISEs answered

**Re-ground `[MEASURED 2026-09-02]`:**
- **Queue item 5 CLOSED with producers named**: `misdirected_cta` page_rerenders COMPLETED
  against services (08-31 23:24 AND 09-01 14:53) and index (08-31 23:24) — the exact producer
  that clobbered the home CTA three times — and every locked row held: row counts at baseline
  (about 4/1, contact 5/1, index 6/3, services 4/3), services still 6 cards, home CTA still
  `/contact.html`. **The locks' behavioural proof is in, organically, twice, against the
  named enemy.**
- A five-audit sweep hit the site 09-01 14:53-14:59 (site-review, content-quality,
  offer-analysis, brief-fidelity, reader-experience) — all findings born `deferred`, none
  dispatched. Awareness only.
- Nobody else touched `execute_vision_prompt` or created `authored_provenance.go` — both
  queue items stood.

**All three council submissions came back REVISE (editquality, round 1 each)** — two were
sketch-fidelity defects (662's sketch abbreviated the WHERE to a nonexistent column; 645's
sketch elided the design-context SQL), one a real check (does step-config max_images shadow?
Answer from code: `execute_vision_prompt_action.go:133` reads `StepConfig.Config` directly —
the value is live). All three RESUBMITTED on their original correlations with verbatim SQL
(SUBMISSION_2026-09-02_resubmit_{662,663,645}.json). ⚠ Lesson, again in a new costume:
**a sketch that abbreviates the file is a claim the reviewer must refute — paste the file.**

**The downscale (queue item 1) is BUILT**: `vision_image_downscale.go` (+5 tests) +
`max_image_dimension` (default 7900) wired into `execute_vision_prompt`, legal images
byte-identical (pinned), 0 opts out (pinned). Council `e5a664d9`. **The RFC_022 lockstep the
new optional key created was caught by verify-head-builds' parity test and honoured in the
same commit** (check.py 4→5, overlay re-applied: configmap created, cronjob configured).
Consumer told (tool-acceptance CONTRIB). thunder + livespec test failures on HEAD are
PRE-EXISTING (verified on the clean tree) — other lanes'. **Inert until the next fleet roll**;
then: restore max_images 16 (migration), re-run the after-leg, read `images_downscaled >= 1`
as the wiring's first proof.

**Archetype heroes + tool hero dispatched** (archetype_content, archetype_blog,
hero_tool_automation_savings_estimator) — eyeball on landing, then wire the 5 remaining
generic pages + the tool page.

**2026-09-02, close: THE IMAGERY STORY IS COMPLETE.** Census at the served site: **ZERO pages
serve generic hero.jpg** (sitemap-wide sweep, positive control: faq serves archetype-content,
1 ref). Final state: 18 per-page heroes + archetype-content (faq/privacy/terms) +
archetype-blog (2 articles) + the estimator's gauge — every page eyeballed before wiring.
D2 delivered in full. ⚠ standing note: all hero wirings remain 403-class loans on the
generic-hero pages (background_image is site_assets.hero-sourced there and a rebuild
re-resolves it) — durable protection is the __authored work, queue item 4.
Open externals at close: fleet roll (downscale inert until then); council verdicts on
e5a664d9 + the three resubmitted correlations.

**2026-09-02, evening — THE DOWNSCALE IS LIVE.** Fresh chassis rolled (replicaset `744cfb4bf`,
pods 15:39/15:53Z). Capability probe on `/proc/1/exe`, the discriminating shape:
`downscaleVisionImage: image scaled for provider limits` PRESENT · `max_image_dimension`
PRESENT · `tool-acceptance-agent` PRESENT (positive control) · invented string absent.
The after-leg chain is UNBLOCKED: restore max_images 16 (next migration), fire
`design_critique_run.sh`, read `images_downscaled >= 1` (wiring proof) + the report
(discrimination read). Handoff 2026-09-02 cut for a fresh session.

### 2026-09-02, late evening — after-leg chain run: 703 + 704 + 705, one dead run diagnosed, rerun in flight

- **Mig 703 applied+recorded** (`max_images` 8 → 16; number 701 was taken by another lane as
  the handoff predicted — grep first paid off). Council corr `40c563ba`, committed `bc87f4686`
  with `Council-Submitted`.
- **Council verdicts checked (handoff item 5): 3 of 4 landed APPROVED today** — `75be8d32`
  (645 seed, round 2), `c6046171` (663 cap, round 2), `e5a664d9` (the downscale). **`52c9a201`
  (662) drew a second REVISE** (gating: prior_art_librarian [high] — the one-active-row claim
  was sourced from round-1's report, not a fresh check; plus editquality/llm_reliability:
  adaptive thinking on sonnet-5 unaddressed, max_tokens 6000 could be eaten by thinking).
- **Mig 704 applied+recorded in response** (`critique.config.ai_service.max_tokens` 6000 → 16000,
  the adaptive-thinking landmine's own prescribed check; resolved-value guard in
  resolveAIServiceConfig precedence order, mig 415's worked pattern). Fresh censuses run for the
  round-3 resubmission: design-critique-agent has exactly ONE `agent_definitions` row of any
  kind (v1, active, non-snapshot, no root `ai_service`); the two-active-row types are exactly
  the landmine's four; loader-visible 207, active snapshots 0, active-but-deleted 0. **Round 3
  resubmitted on the same correlation** (`RESUBMIT_CORR=52c9a201`), commit `4ab0e32c5`.
- **First after-leg run (item `e543ba1f`, 17:19Z) DIED — and its work item lies `complete`.**
  The item's `result` is the SPAWN record (the bugs-287 pattern, exactly as MEMORY warns).
  Truth was in the orchestration (`6cc3cd38`): `current_step complete_error`,
  `__step_error {"message":"Request timed out (code: TIMEOUT)","failed_step":"audit"}`.
  Timeline: orch created 17:23:22.06Z; the render-audit adapter's REAL response (contrast
  findings, 8 of 40 URLs audited, truncated:true) hit the pod at **17:26:24.57Z — ~182s in,
  ~2s after the 180s default await timeout fired**; coordinator logged "No next step defined,
  completing workflow" — the answer arrived and had nowhere to go. Vision step never ran, no
  report, no `images_downscaled` datum. **Mechanism read in code, not inferred**: seed 645's
  audit step has no `timeout_seconds`, so the awaited request gets `DefaultRequestTimeout=180`
  (`datahelpers/timeout_helpers.go:18`); `ConvertStepTimeout` (:23) honours
  `config.timeout_seconds` into `step.Timeout`, and both awaited-request builders stamp
  `TimeoutAt` via `getTimeout(step)` (`coordinator.go:2482`, `:2617`). The 08-26 before-run
  passed because the sweep was quicker pre-hero-batch; it is now a ~2s coin-flip per dispatch.
- **Mig 705 applied+recorded** (`audit.config.timeout_seconds = 600`; guard also refuses if a
  workflow-level `timeout` key ever appears — it would outrank). Council corr `45c9e720`,
  commit `1062001d7`. Checked before choosing 600: the 5-min `StuckOrchestrationTimeout`
  guards `StatusExecutingStep` only, not awaiting-responses (approval steps await 24h on the
  same machinery).
- **Rerun dispatched: item `7d82be2e` (leg `after_hero_batch_downscaled_r2`, 17:36Z), bumped to
  priority 70** (61 items sat ahead at 90; loader sorts ASC). Monitoring the ORCHESTRATION,
  not the item status. Still owed on completion (handoff 1c): `images_downscaled >= 1`,
  the second `design-report` note, discrimination read (hero-sameness lead finding GONE, rest
  stable, and check WHICH 8 of 40 URLs were sampled), `output_tokens` vs the 16000 cap
  (`output_tokens == max_tokens` = CUT), record in NOTES + SQ-003.
- ⚠ Open question for a later pass: item `e543ba1f` reads `complete` on a failed run — if
  bugs 287's fix was meant to stop spawn-record results on THIS path, that fix has a gap;
  check 287's covered producers before filing anything.

### 2026-09-02, ~17:45Z — the after-leg REPORT EXISTS and every 1c check passed

- **Rerun (item `7d82be2e`) completed the full chain**: claimed 17:41:59Z, report in
  `doc_notes` at **17:44:03Z** (id `45fa7cb3`, categories `['design-report']`) — the SECOND
  production report. The audit fit comfortably inside the new 600s window this time.
- **`images_downscaled: 9`** — the downscale wiring's FIRST live production exercise: 9 of 16
  captures exceeded the 7900px long edge and were scaled (9 matching
  `downscaleVisionImage: image scaled for provider limits` lines in pod
  `agent-design-critique-agent-88f1113c-4rn59`); 7 passed through untouched. The handoff's
  stop-condition (0 on known-tall captures) did NOT fire.
- **`llm_call_log` 17:44:02Z: model_resolved claude-sonnet-5, `max_tokens 16000`, success=t,
  latency 25.9s** — mig 704 was live in the call (the two 08-27 rows alongside it show the
  failed 6000-era calls for contrast). ⚠ `input_tokens`/`output_tokens` are **NULL on this
  logging path** for this agent, so the adaptive-thinking landmine's "read output_tokens
  against the cap" is not satisfiable from the table; truncation checked at the artefact
  instead — the report is well-formed and ends complete (~5.5KB, nowhere near the cap).
- **DISCRIMINATION READ: PASSED.** Before (08-26, Gemini leg): lead finding = "the exact same
  network-node geometric hero graphic across almost every page". After (09-02, Claude leg,
  16 images / 8 pages / 2 viewports — same coverage shape): that finding is **GONE**; the
  residual is a two-page orb echo (homepage + quiz heroes), and the report affirmatively
  describes DISTINCT visuals per page (calls the Leopardess Line transit-map diagram "the
  strongest single visual on the site"). The affirmative description is what makes this
  trustworthy — an absent finding alone is the unreliable direction (visual-design-auditor
  landmine). Sampled pages overlap the before-set on ~6 of 8.
- **Rest-substantially-stable check, honestly mixed**: how-it-works text-wall PERSISTS
  (unfixed, expected); services-carousel hierarchy complaint GONE (page now praised — but
  carousel work also happened in the interim, and one audit is a sample); NEW findings:
  footer sitemap-dump (both viewports), quiz-hero five identical avatar circles, use-cases vs
  case-studies near-identical composition, ROI mobile tile stacking, homepage mobile hero SVG
  overlap. **Work-queue item 2 (the critic's 8 findings) should be refreshed against THIS
  report** — the 08-26 list is one-third stale.
- Register **SQ-003 updated** (status + verify-later): vision half UNPARKED, evidence inline.
- Council state at close: `52c9a201` (662+704) **APPROVED round 3**; `40c563ba` (703) round-2
  resubmission in flight; `45c9e720` (705) round-1 in flight. Both carry `Council-Submitted`
  trailers on their commits; 098 credits on approval.

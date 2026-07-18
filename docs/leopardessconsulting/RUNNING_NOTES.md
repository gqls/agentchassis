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

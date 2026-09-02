# NOTES — analytics / GTM

Append-only, newest at the bottom. Technical log: what was tried, what the system
actually said, and every misstep.

---

## 2026-07-30 — Phase A: GTM on idea.uk's static site

**Task.** Owner supplied container `GTM-PQ3WCTBD` and asked for the script as high in
`<head>` as possible and the noscript immediately after `<body>`, on every page.

### What I got wrong on the way

- **Guessed the schema instead of reading it — twice.** Wrote `pages.rendered_html`
  (does not exist; it is `rendered_head` and pages store no full HTML) and
  `site_components.function` (it is `slot_name`). CLAUDE.md says `\d <table>` first;
  I didn't, and both errors were caught only by the DB refusing the query. Cheap here
  because SQL fails loudly — but the same habit against a `jsonb` path fails *silently*.
- **`go build ./...` at the repo root "passed" and proved nothing.** The idea.uk tool
  is a **separate Go module** (`golang_files/go.mod`, `module idea`), so the root build
  never compiled the file I had just edited. I nearly recorded a green build as
  evidence. Caught by checking for a nested `go.mod`. → `go build -C <dir>`.
  **Generalises: "the build passed" is only evidence if you know which module it built.**

### Findings that shaped the design

- `assemblePage:574-593` writes `<body>` as a **Go string literal**, so "immediately
  after `<body>`" is reachable only as the top of the `header` slot — no chassis
  change needed, but no other option either.
- `Document Head` is shared by **9 sites** (measured). Hardcoding the container would
  have tagged eight other domains with idea.uk's id.
- `Document Head`'s `input_schema` holds **flat scalars**, which the gap-fill loop
  skips as "not a field descriptor" (`:612-615`) — so its `title`/`description`
  entries have never resolved and never could. A map-valued key was required.
- No CSP header on idea.uk (`curl -sI`), so nothing blocks googletagmanager.com.

### Result

`p4_34_gtm_container.sql` applied: site_specs key + both templates gated + both stored
artefacts written, all inside one transaction with pre-guards and post-assertions.
20 pages re-assembled via assemble mode; **20/20** rendered artefacts carried both tags
with the noscript first after `<body>`; **19/19** fetchable live URLs verified.

`tool-audience-check` was excluded — it is a stub row (`/tools.html#audience-check`,
0 sections, `deployed_at` NULL) whose URL would derive the junk filename
`tools.html#audience-check.html`. It has never been deployed.

---

## 2026-07-30 (later) — the finding that mattered more than the task

**`/privacy.html` was the one live URL that failed verification.** It returns **301**
to `/privacy` — and `/privacy` had **zero** GTM hits.

Chasing that down: **idea.uk is two applications behind one domain.** nginx proxies a
**16-route reserved set** to a Go binary on the VM; everything else is the static site.
Eleven of those routes render HTML through a single wrapper, `App.page()`, and none of
them exist in the static build:

- **"Payment received"** (`/order/success`) — **the conversion page**
- **"Request received"** — the £29 order submission
- `/privacy`, `/terms`, `/refund-policy` — and the static `.html` copies **301 to
  them**, so the tag I had just shipped on those three static pages can never fire
- `/order/cancel`, subscribe confirmation, audience-check result, operator pages

**So "GTM is live on every page of idea.uk" would have been a false claim** — and
specifically false about the only two pages that can evidence a sale. Google would
have shown traffic and zero conversions, which reads as "the site doesn't convert"
rather than "the tag isn't there".

> **The check that caught it was `curl` WITHOUT `-L`, reported per URL.** A summary
> line ("19/20 pass") nearly buried it; following the redirect blindly would have
> scored `/privacy`'s content as `/privacy.html`'s. Neither `-L` nor no-`-L` is right
> on its own — what worked was recording the **status code** alongside the hit count
> and looking at the one row that differed.

**Phase B written, not deployed.** `App.page()` now injects both snippets from
`GTM_CONTAINER_ID` (env), sanitised to `[A-Za-z0-9_-]{1,32}` because the value lands
in both a JS string literal and a URL. Landing page got `<!--GTM_HEAD-->` /
`<!--GTM_BODY-->` placeholders wired through `NewApp`'s Replacer. Module builds, `go
vet` clean, full suite green, 5 new tests in `gtm_test.go` asserting **placement**
(not mere presence), inertness when unset, and rejection of malformed ids.

**Not deployed:** it is the live Stripe payment service and `capacity` reported
`{"active":1}`. Restart is an owner call. Rollback is the existing
`/opt/idea/idea.bak.*` binary-swap pattern.

---

## 2026-07-31 — a competing analytics seam already existed, and I only found it by accident

**Phase A re-verified after the chassis roll.** Still live: 5/5 spot-checked URLs return
2 GTM hits, and `site_components.updated_at` for both idea.uk slots is unchanged at
`2026-07-30 19:37:05`. The roll neither reverted nor rebuilt the chrome, as expected —
Phase A is DB-only and touched no Go.

### The finding

While checking charset anchors across the three head components (for Phase C, because
an anchor mismatch would make the insertion silently miss), the same query returned
`has_gtm` — and **`head-seo-standard`, which I have never touched, came back true.**

It has carried a gated analytics block since **2026-05-13**:

```
{{if .analytics_id}}
<script async src="https://www.googletagmanager.com/gtag/js?id={{.analytics_id}}"></script>
… gtag('config', '{{.analytics_id}}'); …
{{end}}
```

declared properly — `{"type":"text","source":"config.analytics_id","required":false,
"on_missing":"skip_field"}`. That is **the identical mechanism I designed for Phase A**,
two months older, on a different head component covering 4 domains
(ai-agent-orchestration.com, finetuning.uk, gaswholesalers.com, leopardessconsulting.co.uk).

> **CORRECTED:** yesterday's commit message and STY-050 both claimed this was "the first
> real consumer of `bugs_open/018`'s schema-driven fill". **That is false.** Withdrawn in
> the register with the correction inline.

**It is dormant, not broken:** `SELECT count(*) FROM site_specs WHERE is_current AND
data::text ILIKE '%analytics_id%'` → **0**. No site sets it, so the gate never opens and
no artefact carries gtag. A correctly-built seam that nothing populates — which looks
exactly like a broken one until you check the schema rather than the output.

### Two mistakes inside the same investigation

1. **I never grepped for prior art on the thing I was building.** One query would have
   done it before I designed anything:
   `SELECT name FROM content_components WHERE html_template ILIKE '%googletagmanager%'
    OR html_template ILIKE '%gtag%';`
   I grepped `/bugs_open/` and the workstream dirs as CLAUDE.md requires, but never the
   **live component corpus** — which is where a rendering mechanism actually lives.
2. **I then mis-read its schema and nearly recorded a second false claim.**
   `input_schema->'analytics_id'` returned NULL and I wrote "not declared". The shape is
   **wrapped** — `input_schema->'fields'->'analytics_id'` is the live path, and
   `render_site_components_action.go:604-607` handles both shapes precisely because both
   exist. A working seam looked undeclared because I queried the wrong path. Caught only
   because I re-checked before asserting.

### What it changes for Phase C

There are now **two competing analytics seams** in the fleet:

| seam | component | domains | mechanism | state |
|---|---|---|---|---|
| `config.analytics.gtm_container_id` | `Document Head` | 9 | **GTM** | live on idea.uk only |
| `config.analytics_id` | `head-seo-standard` | 4 | **gtag.js / GA4 direct** | dormant, 0 sites |
| — | `webdesign.co.uk Document Head` | 1 | neither | — |

**A site that ever carries both would load GA4 directly AND through GTM and
double-count every pageview.** So Phase C is no longer "repeat the recipe 13 times" —
it needs a decision on which seam survives. Recommendation: **GTM only**; retire
`analytics_id` (or make the two mutually exclusive in the template) rather than leave a
dormant mechanism that will look reasonable to the next session that finds it.

**Third trap for Phase C, unrelated:** `webdesign.co.uk Document Head` uses a
**lowercase** `<meta charset="utf-8">`. The anchor is case-sensitive in `replace()`, so
the idea.uk migration applied verbatim would update 0 rows on that site and report
success. Guard on the anchor count, per site, as p4_34 does.

---

## 2026-07-31 (afternoon) — Phase B deployed, Phase C applied

### Phase B — the payment box

Owner authorised the deploy. `/capacity` had reported `{"active":1}` on three separate
checks, so before restarting I read `orders.json` rather than trusting the count: the
active order was **`ord_1785236456008987049`, status `awaiting_payment`** since 07-28 —
a persisted state waiting on Stripe, **not** an engine run mid-flight. The rest were
stale `requested` rows from 07-13 onward (the known spam class) and three `declined`.
That distinction is what made the restart safe rather than merely survivable: the
recover-interrupted path exists for the mid-run case, and this wasn't one.

Sequence: backed up the binary (`idea.prev-2026-07-31-123615-pre-gtm`) **and**
`orders.json`; built `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 -trimpath` (9,909,585 bytes
vs the live 9,999,070 — same shape, stdlib-only module so no glibc coupling); appended
`GTM_CONTAINER_ID=GTM-PQ3WCTBD` **on its own line**; restarted.

Post-restart checks that actually prove something:
- `systemctl is-active` → active
- `/health` → `"provider":"*main.StripeProvider"` — **the useful one.** The service falls
  back to `FakeProvider` if either Stripe var is missing, so this single field proves the
  EnvironmentFile still parsed correctly after my append. A bare "it's running" would not
  have caught a mangled env file.
- `/capacity` → `{"active":1}` unchanged, so `orders.json` survived intact.

All five tool-served HTML pages: `script=1 noscript=1 body_adjacent=yes`. And
`curl -sL /privacy.html` — the 301 that started the whole investigation — now returns 2.

### Phase C — DB applied fleet-wide

`sql/c1_gtm_fleet_rollout.sql`, one transaction. Retired the gtag seam, added the gated
GTM block + `gtm_container_id` descriptor to the 2 remaining head components and 5
remaining header components, set `site_config` on the 13 other domains, patched 26
artefacts. Post-conditions asserted and passed: **14/14** head artefacts tagged, **14/14**
headers with the noscript FIRST, **14/14** specs, **9/9** templates gated, and
`analytics_id` gone from every template and schema (0 matches).

Three things the guards earned:

- **`header-theme-chrome` has a NULL `input_schema`.** Its descriptor had to go at top
  level, not under `fields` — with no `fields` key the gap-fill treats the whole object
  as flat (`:604-607`). A blind `jsonb_set(…,'{fields,gtm_container_id}')` on NULL would
  have produced a schema the resolver never reads.
- **`webdesign.co.uk Document Head` needed the lowercase anchor**, exactly as predicted:
  the uppercase UPDATE touched 12 rows, the lowercase one touched 1. Had I written only
  the uppercase form it would have reported success and left that site untagged.
- **`webdesign.co.uk Document Head` emits no `<head>` open tag at all** — it begins at
  `<meta charset>`, so those 99 pages serve an *implicit* head. GTM lands correctly
  inside it, so this is not a blocker, but it is a real pre-existing defect and is
  deliberately **not** fixed here: it would touch every page of that site for a reason
  unrelated to analytics.

### Page re-assembly

Proven on the smallest site first (vetcomparison.uk, 6 pages) before touching the other
12 — 6/6 rendered with both tags, 6/6 live. One page read `gtm=0` on first check and was
simply **fetched mid-deploy**: `last-modified` was 43 seconds *after* the `deployed_at`
I had just read. Re-checked and it was fine. Worth remembering — an immediate live check
after a deploy can catch the old bytes and look like a failure.

> Incidental: `pages.deployed_at` is **not** reliably refreshed. `contact` still shows
> 2026-07-18 yet serves the new chrome. Do not use `deployed_at` to decide what shipped;
> use the artefact or the live page.

### Page re-assembly fleet-wide — and the deploy-queue mistake I made

Rolled smallest-first. **377/377 pages published (0 publish failures), 377/377
orchestrations COMPLETED, 0 FAILED, 377/377 rendered artefacts carry both tags.**

Then the live sweep came back **9/14 PASS** — and the 5 failures were not what they
looked like.

`ai-agent-orchestration.com`, `finetuning.uk`, `gamesdesign.co.uk`,
`leopardessconsulting.co.uk`, `webdesign.co.uk` served untagged pages with
**`last-modified: Mon, 27 Jul 2026`** — bytes four days old. The render was fine; the
**deploy had not landed**. Textbook "COMPLETED is not proof": `deploy_page` even reported
`"success": true` with a real commit to `github.com/gqls/sites` at 12:54:59Z. The commit
was true. The *deploy* was not, yet.

**The cause is mine.** These sites deploy by GitHub Actions ("Deploy to B2") on push, on a
**self-hosted runner with ~2 concurrent slots**. One page-rerender = one commit = one
workflow run. 377 pages therefore queued **~230 runs**. And they are all redundant: the
workflow does `b2 sync --delete --skip-newer "$domain" "b2://portfolio-sites/$domain"` —
a **whole-directory sync**, so the last run per domain would have deployed everything.

Measured: 77 completed / 221 queued / 2 in progress, draining at roughly 2/min ⇒ **~100
minutes**. Left to drain rather than cancelled: `--skip-newer` makes out-of-order syncs
safe, and mass-cancelling runs on a repo other sessions also push to is a worse risk than
waiting. But the backlog blocks *their* deploys too, which is the part that was
inconsiderate rather than merely slow.

**Do it differently next time.** For a chrome change affecting a whole site, the
re-render fan-out should be decoupled from the deploy fan-out — either batch the commits
(one commit per site, not per page) or let the pages render and then trigger **one**
sync. The workflow even has the hook for it: `CHANGED` is derived from
`git diff HEAD~1 HEAD | grep -E '^[^/]+\.[^/]+/'`, and **when it comes back empty the
workflow syncs every domain**. So a single commit touching only a root-level file
triggers one full-estate deploy. That is one run instead of 230.

**Two deploy runs failed, and it is a race, not a defect.** Both died with
`FileNotPresent … ERROR: Incomplete sync` while deleting an old version — two concurrent
runners doing `b2 sync --delete` against the same bucket prefix, one removing a version
the other had already removed. Self-healing here, because later queued runs re-sync the
same domains. Worth knowing: **`b2 sync --delete` is not safe to run concurrently against
one prefix**, and its failure mode is a red workflow rather than corrupted output.

> Also confirmed the earlier live-check caveat is real and generalises: one vetcomparison
> page read `gtm=0` because I fetched it 43 seconds *before* its `last-modified`. An
> immediate post-deploy check can read the old bytes and look exactly like a failure.

### The regression that proved `--skip-newer` does not protect you

On the final sweep **`dartsonline.com` went BACKWARDS** — it read `gtm=2` at 13:00 and
`gtm=0` twenty minutes later, cache-buster confirmed, not a CDN artefact. Its
`last-modified` was `12:52:18Z`: *rewritten during my own rollout*, with pre-GTM bytes.

Checked the source of truth before theorising: **git master is correct** — `index.html`
for dartsonline and all four other pending domains returns `grep -c googletagmanager` = 2
straight from the GitHub API. So the render, the commit and the repo are all right; only
the bucket was wrong.

**Mechanism.** Queued runs check out **their own** (older) commit, and if that commit's
diff touches no domain-shaped directory the workflow falls through to syncing **every**
domain — from an older tree. I had assumed `--skip-newer` made out-of-order execution
safe. **It does not:** `git checkout` stamps every file with a *fresh* mtime, so the
source always looks newer than the bucket and nothing is ever skipped. An older run can
therefore overwrite a newer one, and with ~230 runs draining out of order the estate
churns rather than converging monotonically.

> **CORRECTED:** the entry above says "left to drain … `--skip-newer` makes out-of-order
> syncs safe". **That was wrong**, and dartsonline is the counter-example. Draining is
> still safe in the sense that nothing is lost — the repo is authoritative — but the
> *final* state depends on whichever run happens to touch a domain last, which is not
> something to leave to chance.

**Fix applied:** pushed one root-level `.full-sync-stamp` (commit `2582e69f5`). A
root-level file matches neither the domain-dir grep nor `paths-ignore`, so it triggers the
workflow with an **empty `CHANGED`** → the fallback branch syncs **every** domain from the
current tree. Queued at 13:14:00Z, i.e. **behind** the whole backlog, so it runs last and
converges the estate regardless of the order everything before it ran in. One run instead
of 230 — which is also the shape the rollout should have used from the start.

---

## 2026-08-25 (session "google") — picked up from the apis.uk handoff; the container is empty and the estate's tag is half-durable

**Entry point.** Owner opened a session named "google" with
`apis_uk_bees_homepage/HANDOFF_2026-08-25_continue_here.md` — §4a there is his verbatim request
("walk me through setting up the google tags in baby steps … under agent chassis and not idea.uk"),
spun out "to a new lane". This lane already exists and is that lane; work lands here from now on.
Nothing was touched on the cluster except reads and one dry run that rolled back.

**1. Container state, measured, not remembered.** `curl https://www.googletagmanager.com/gtm.js?id=GTM-PQ3WCTBD`
at 16:06 BST → HTTP 200, 322,091 B, `"version":"2"`, `"tags":[]`, zero `G-` ids, zero `__googtag`,
zero `__gaawe`. **Nothing is published; nothing is recorded.** Same as the owner's 08-24 screenshots.
apis.uk: 200, 67,877 B, 1 × `googletagmanager`, 0 `Set-Cookie`.

> **CORRECTED 2026-08-25 — my own 07-31 entries and README told the owner "analytics cookies now
> fire on 14 domains" and "they're all running analytics". Both were INFERRED from the snippet being
> on the page; neither was measured at the container, and the container had no tag.** So no cookie
> was ever set by it, and the consent gap I flagged (still real once a tag publishes) was raised
> about cookies that did not exist. `WRONG_CALLS.md` 2026-08-25. The check I should have run is one
> curl of `gtm.js`; it is now `scripts/check_gtm_state.sh`.

**2. No Google credential exists anywhere** — `~/.config` (only chrome-for-testing), no `gcloud`,
`GOOGLE_APPLICATION_CREDENTIALS` unset, no google/gcp secret in `ai-persona-system`. So Search
Console (039 §5) stays an owner action; nothing here can be automated yet.

**3. The census moved, and then split.** Handoff said 27/27 heads. Today (`sites.status IN
('deployed','active')`, head slot): **26** carry `GTM-PQ3WCTBD` in `rendered_html`, **4** deployed
sites do not — `agritec.uk cv1.co.uk homegarden.uk lampenkap.com`, all born 08-24/25, all
`Document Head`, all with NO `site_config` spec row. Then the split: spec key
(`site_config.analytics.gtm_container_id`, the thing the template's `{{if .gtm_container_id}}`
reads) vs artefact → **A 14 durable / B 12 artefact-only / D 4 neither / E 1 no head**. Bucket B is
exactly the 08-24 backfill (minus agritec): `updated_at = 2026-08-24 13:13:52` on all 12.

**4. The disconfirming check confirmed it.** I asked for the case that would refute "the next chrome
render drops it": a head re-rendered after 13:13:52 that still carries the tag. Instead:
`agritec.uk | 2026-08-24 19:20:53.178 | app - 10.20.31.31:32834` — had it before, lacks it now — and
the platform's own `chrome_divergence_overwritten:site_component:head:920676eab287` row at
`needs_human_review`, created 19:20:53.239. Trigger chain on that site: `add_tool` 19:01 →
`nav_drift`/`nav_rebuild` 19:17–19:21 → chrome regenerated. Two instruments, one second apart, the
only rows of their kind. **Filed `bugs_open/397`** with the 090 substitute declared (template,
fingerprint SQL and check read; census; this refutation attempt).

**5. The fix is a wave, and I did not fire it.** `ChromeRenderInputsSQL` md5s `site_config` into
`render_inputs`; `StaleSiteComponentsCheck` files `needs_rerender`/`stale_chrome` on drift (20 ever,
all `complete` — it dispatches); `rerender-pages` force-rerenders every slot and page. 12 sites =
**241 pages**, 16 = 280, one commit + one Actions run each, 130 still queued from 08-24. Wrote
`sql/c2_gtm_spec_key_for_artefact_only_sites.sql`: `-v GO=yes` gate (refuses at exit 3 without it),
`-v DRY=1` dry run (ran: 12 targets, 241 pages, merged keys `analytics,locale[,chrome]`, `ROLLBACK`,
0 rows written), `-v UNTAGGED=1` adds bucket D, network-scoped to `…0002`, post-conditions incl.
"no key lost". **Merge, not replace** — found the 07-31 rollout (mine) replaced wholesale and dropped
`relojistas.com`'s `intent_probe` key; `locale` was later re-merged by migration 508/530. Recorded
against myself.

**6. The apis.uk handoff's §3 design is superseded, not wrong in intent.** "Per-site analytics id:
read `sites.settings->>'analytics_container_id'` in `RenderFallbackHead`" — the per-site seam
already exists (STY-050, this lane, 07-31), and `RenderFallbackHead` is the path taken only when the
head component fails to render. Told them (CONTRIB in their dir). The real gap is that **no Go
writer touches `site_config`** (every current row is `created_by` a migration or a session), so
"standard for new builds" has no mechanism — 397 §6.2.

**7. Contamination note for the record.** This session curled 25 domains once each (`--sites`) plus
apis.uk ~5×. Cloudflare counts all of it. 039 §1.

**Commits this session:** lane docs + script + sql; `bugs_open/397`; ledgers (LANDMINES correction +
entry, WRONG_CALLS); CONTRIB to apis.uk. Nothing applied to the cluster.

**8. 17:16 BST — owner ruling relayed by the apis.uk session (cross-session message + their CONTRIB
`443066755`).** Owner, verbatim: *"section 4 has google in it which is taken by another lane, please
communicate to that lane that that is what they take and we will take the rest here."* So this lane
now owns GA4 publication + consent, 397 in full, Search Console, the fleet dashboard script (never
started anywhere), and `039_REFERENCE`. They accepted the CONTRIB §2 and dropped the
`RenderFallbackHead` build (their handoff corrected visibly at the top; `sites.settings->>
'analytics_container_id'` has 0 rows and stays unused). Two apis.uk facts they passed over, one of
which I measured rather than relayed: (a) apis.uk's index refuses page re-renders —
`[MEASURED 17:20]` `page_rerender_index_…_template_changed` `failed` at 11:19 with `result={}`, no
error recorded (099 pattern) — so the c2 wave's page item on apis.uk will fail, harmlessly; (b) add
them to 397 §9 and tell them after c2 runs. Both recorded (397 §9, c2 banner, handoff). No reply
sent — none was asked for.

---

## 2026-08-26 — the strip happened overnight; c2 applied on the owner's go

**9.** Second apis.uk relay warned the rotation (back on per `bugs_open/401`) "may promote repairs"
but claimed nothing promotable today. Measured instead: their own 00:40Z batch promoted
`rerender-pages` twice, and the loss query returned **10 new strips overnight** — 7/7 keyed kept,
10/10 unkeyed lost (397 §10 has the table). Their favicon/og-card "404" was also stale by
measurement time (asset-deployer had completed; both 200). **Both directions of staleness in one
message, hours old — measure, never relay.**

**10.** Owner: "please carry on" (after my read-out naming notify+apply as the next act). Notified 9
of 10 lanes by cross-session message (loanzy blocked by the local classifier — record stands in 397
§9/§10); rotation lane banked the natural experiment and relayed the "nothing promotable" refutation
to apis.uk itself. **Applied c2 (10:12:11Z by the rows' stamp — I first wrote 10:50Z here without measuring; apis.uk lane caught it), UNTAGGED=1: 17 sites, UPDATE 11 / INSERT 17, keys preserved,
census A 16 / C 15 / B 0 / D 0, 17 rows `created_by='claude-session-google-2026-08-25'`.** Dry-run
pages 323 → apply 334 in ~1 h — the estate builds under you; the count carries its date.

**11.** Baselines banked from acks: webdesign.uk was 5/7 pages tagged pre-strip, expect 7/7 post-
rebuild, and their hand-placed "Not active yet" index label will be wiped by the rerender (theirs,
expected — a diff must not read it as my change adding/removing content). homegarden: pages *built*
(not re-rendered) by the wave run the post-627/628 writer — copy deltas across that boundary are the
copy lane's migrations, not GTM. apis.uk: page item expected to fail; they settle `build_status`.

**12.** Open tail: watch C drain to A (`check_gtm_state.sh --db`); the §6.2 structural half is still
owed (adversecreditmortgage.co.uk arrived overnight unkeyed and proved it again — c2 caught it only
because it ran today); GA4 still unpublished (container version 2, 0 tags at 10:55Z).

**13.** `[12:00–12:30 BST]` Disposition of the 11 `chrome_divergence_overwritten` head items (one per
stripped site, matched to strip events ≤130 ms): **the batch UPDATE was blocked by this session's
auto-mode permission layer. Not retried, not routed around, no peer asked** — the items stay open;
397 §10-addendum is their written answer and the UPDATE goes to the owner. noted's 08-18 header
sibling (`ab9afa54…`): predates the backfill, no GTM link found — not ours, told them.

**14.** **apis.uk correction adopted:** "index refuses page re-renders" DISPROVEN — three completed
overnight (that is how its tag went); the 11:19 failure was 383's re-walk. My adoption of their
inference after measuring one failed row is the misstep half that is mine: one failed item is an
event, not a property. Corrected in 397 §9, c2 banner, handoff. Their WRONG_CALLS carries the
inference half. Also theirs: my c2 apply and their "no site_config row" read raced within the hour —
unexplained on their side, flagged not smoothed.

**15.** Banked coordination facts (397 §10-addendum has the full versions): cv1 — rerender safe,
**REBUILD crashes the pod (408); never `needs_rebuild` cv1's index/tool-example**; remortgage —
interim tag-drop window from six pre-queued rerenders, rotation ETA ~18 h; agritec — their own
13-page wave + 17 imagery items interleave, favicon 404 pre-existing; loancalculator — post-wave
`toolgolden.py` 11/11 expected, and mid-wave "one page differs" = not-yet-reached, not skipped;
webdesign.uk — their "Not active yet" index label is wiped by the rerender, theirs, expected;
homegarden — pages BUILT by the wave run the post-627/628 writer (copy deltas are the copy lane's).

**16.** Owner question now open: cv1.co.uk and lampenkap.com in the estate's GA4, or keys retracted?
Both lanes ruled it portfolio, not technical. Cost either way is one supersede.

**17.** `[2026-08-26, 384-lane close]` lampenkap ruling: put to the owner twice, no answer yet; the
384 lane will write it here as a dated CONTRIB when it comes, and **"no news from them = no
retraction requested"** — the key stays, the decision is retract-or-not (one supersede), nothing
blocks on it. Confirmed: their sweep's `consumer_pages: 1, current: 0, unknown: 1` reading on
lampenkap is the empty `tool-list` (no `tool` pages), unrelated to GTM, and will read the same
after the re-render — nobody should expect the wave to change it.

**18.** Same-file passenger, owned: my WRONG_CALLS commit `b6fd59944` swept in the `bugs_open/384`
lane's then-uncommitted entry ("USAGE count read as DAMAGE count") alongside mine — the scope report
cannot see a same-file passenger, and the numstat (58 lines for a ~16-line entry) was the tell.
Nothing lost, forward-only; their entry is intact and now committed under my message; they were told.

**19.** `[2026-08-26 ~11:55 BST]` **OWNER RULING: "leave lampenkap google tag"** — filed by the 384
lane into this dir as agreed (`CONTRIB_2026-08-26_from_bugs_open_384_owner_ruling_leave_lampenkap_tagged.md`,
`d3f04b95a`). cv1 remains open. The ruling arrived with an owner question — *"is that what GA4 is?"* —
answered with a fresh measurement, not memory: container re-read 10:50Z, **still version 2 / 0 tags /
no `G-` id**, so tagged ≠ reporting; nothing reports until he publishes. Plain-English version added
to README for him. The 384 lane deliberately left the wiring fact to this lane — correct split.

**20.** `[2026-08-26 pm]` webdesign.uk relayed the customer-default ruling (their
`DECISION_2026-08-26_default_tag_hosted_copy_only.md`): owner default on hosted customer copies
only, per-site override, ZIP clean. Told them their work-package §1 field ALREADY EXISTS (STY-050 —
their "no per-site tag field" was measured at `sites.settings`+Go, the 07-31 false-absence class
again) and adopted their `analytics.mode` idea into 397 §6.2 (a seeder makes "none" need a stored
representation). **Flagged the collision that is mine to own: GA4 publication into `GTM-PQ3WCTBD`
makes hosted customer sites set `_ga` with no banner on day one** — second cookie-light container /
Consent Mode / re-ruling are the options; goes to the owner with the consent decision, not after it.

**21.** `[2026-08-26 ~13:10 BST]` Wave progress, measured for the webdesign.uk launch snapshot:
webdesign.uk's `stale_chrome → needs_rerender` created 12:06:23Z, COMPLETE; head regenerated
12:17:57Z **with the tag** — first C→A conversion observed end-to-end. Census at the same read:
see the bucket line committed below. Two loose ends flagged to their lane, not mine: their vm-sites
census path can't be `gqls/sites/webdesign.uk/` (holds only `assets/`, no HTML — their 5/7 figure
read something else), and `page_rerender misdirected_cta:what-you-get` unresolved since 12:06:51Z.
Their question "does the second-container ruling gate your publish?" answered NO with reasons
(estate publish gated only by the standing consent decision; the ruling gates the FIRST hosted
customer build) — consistent with 397 §6.2 as banked.

**22.** `[2026-08-26, webdesign.uk close]` Repo-path reconciliation, banked so nobody re-derives my
404: **webdesign.uk does not serve from `gqls/sites`** (hence its dir holding only `assets/`) — it
serves from **`gqls/vm-sites`** via the box's 5-min sitesync to `/var/www`. Their 5/7 census was the
right instrument. My 12:17:57Z regenerated head reaches it as a `Rerender:` commit when the deploy
runner drains; label still present at their last fetch, wipe watched and re-placed by their lane.
`misdirected_cta` theirs; no traffic figures in their launch compile. Both flags closed.

**23.** `[2026-08-26 night]` **OWNER RULING via webdesign.uk lane: second cookie-light GTM container
for customer sites** ("please go ahead with a second GTM analytics container") — their DECISION doc
§5 updated same night; estate GA4 publication into GTM-PQ3WCTBD is thereby unblocked. Execution
mine; **blocked on ACCESS, not decision** — re-checked: still no gcloud/GAC/cluster credential, and
a container lives in the owner's Google account. Handed the owner the 2-minute dashboard walkthrough
(README); alternative is the Tag Manager API via the same service account Search Console needs (one
grant, two unblocks). Design note banked in 397 §6.2 before it becomes a repeat: an EMPTY second
container is cookie-light and RECORDS NOTHING (the 0-tags lesson, by construction this time) — the
decision's visit-visibility purpose needs a Consent-Mode-defaults-DENIED GA4 tag speced at
publication. The container id, once it exists, is the one-place fleet default handed to the
webdesign.uk + delivery lanes.

**24.** `[2026-09-02 ~21:00 BST]` Owner created the customer container from the All-accounts screen
(my 08-26 walkthrough said "Admin", which only exists INSIDE an account — corrected in chat to the
⋮ → Create Container path off the account row; README below carries the right steps for next time).
**`GTM-TH5XGNQ4`, verified live 20:01Z: 200 / v1 / 0 tags / no `G-` id.** Banked in 397 §6.2 +
handoff + RUNBOOK (with the inverted-verdict warning: 0 tags is CORRECT for this container, and a
tag appearing IS the §5 re-ruling trigger). Id handed to the webdesign.uk lane. Estate GTM-PQ3WCTBD
re-read in the same hour: still 0 tags — the owner's publish click is now the only outstanding
Google action from the whole 08-25/26 arc.

**25.** `[2026-09-02]` webdesign.uk confirmed: `GTM-TH5XGNQ4` recorded in their DECISION doc §5 as a
dated addendum (`1539c1651`) with the inverted-verdict warning and the co-canonical pointers
(397 §6.2 + this lane's handoff) — the three records agree. Handover complete; seeder +
`analytics.mode` build remain this lane's, their intake/ZIP packages now cite the concrete id.

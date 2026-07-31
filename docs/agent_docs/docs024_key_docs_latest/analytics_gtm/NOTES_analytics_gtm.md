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

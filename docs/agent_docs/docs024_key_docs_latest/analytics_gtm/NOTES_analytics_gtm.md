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

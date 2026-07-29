# 131 — every page advertises a social preview image that does not exist, on 11 of 14 live sites

**Found:** 2026-07-28, from relojistas.com, while auditing what crawlers and social unfurlers
actually receive.
**Severity:** medium and **entirely outward-facing**. Every share of every page on 11 sites —
WhatsApp, Slack, X, Facebook, LinkedIn, iMessage — renders with no preview image. This is the
first impression the estate makes when anyone passes a link on, and it has never worked.
**Status:** OPEN — measured, not fixed.

## The defect

`render_site_components_action.go:417-448` builds the head and writes, unconditionally:

```go
b.WriteString("  <meta property=\"og:image\" content=\"" + origin + "/assets/images/og-card.png\">\n")
```

**Nothing generates `og-card.png`.** The tag is emitted whether or not the file exists.

## Measured across all 14 sites with deployed pages, 2026-07-28

| | sites |
|---|---|
| `og:image` present, **card 404s** | **11** |
| `og:image` present, card 200 | 2 — `leopardessconsulting.co.uk`, `robot-hands.com` |
| no `og:image` at all | 1 — `webdesign.co.uk` |

The 11: `ai-agent-orchestration.com`, `dartsonline.com`, `finetuning.uk`,
`fundamentallyai.com`, `gamesdesign.co.uk`, `gaswholesalers.com`, `idea.uk`, `oufe.com`,
`relojistas.com`, `vetcomparison.uk`, `vonc.com`.

```bash
# reproduce — one line per site
while read -r d; do
  img=$(curl -s "https://$d/" | grep -o 'property="og:image" content="[^"]*"' | sed 's/.*content="//;s/"//')
  printf "%-26s %s %s\n" "$d" "${img:-none}" "$([ -n "$img" ] && curl -s -o /dev/null -w '%{http_code}' "$img")"
done < domains.txt
```

The two that pass are the interesting control: **something did generate a card for them**, so
the emitter is not universally broken — the asset pipeline is. ~~**[UNDIAGNOSED]** which path
produced those two; find it before building a new generator, because it may only need wiring.~~

> **ANSWERED 2026-07-28 (evening) — and the instinct was right: nothing needs building.**
> The generator exists, is registered, and has been live since 2026-07-11:
> `platform/orchestration/actions/derive_brand_head_assets_action.go` (`registry.go:185`).
> It reads the site's active `logo` asset from S3, resizes it to a 64px favicon, composes it
> centred on the brand colour as a 1200×630 card, and commits both to the site repo.
> **Reachable in production today** — `asset-deployer` carries a live `check_mode` branch
> (verified on the live agent row, not the seed):
> `input_data.spec.mode == "brand_head" OR input_data.mode == "brand_head"`.
> So "nothing generates og-card.png" is true of what has *run* and false of what *exists*.
> **robot-hands got its card from this path on 2026-07-11. leopardess's was hand-committed**
> from its owner-approved logo (`docs/leopardessconsulting/RUNBOOK.md` H4) — which is why
> leopardess serves a 200 card while having **no `og_card` row at all**.

## Second defect in the same head block: `og:title` is often just the domain

```
dartsonline.com     og:title = "dartsonline.com"
fundamentallyai.com og:title = "fundamentallyai.com"
relojistas.com      og:title = "relojistas.com"
idea.uk             og:title = "idea.uk"          … and 4 more
```

versus sites that get it right — `"AI Agent Orchestration"`, `"Leopardess Consulting"`,
`"Gas Wholesalers"`, `"FineTuning"`, `"OUFE"`. So the value falls back to the domain when
whatever supplies the display name is empty. **`og:description` is absent on relojistas
entirely**, though the emitter has a branch for it (`:446`) — so that too is conditional on a
field that is not always populated.

Net effect on a share of `https://relojistas.com/glosario/tourbillon.html`: no image, the title
"relojistas.com", and no description — for a page whose actual `<h1>` is
*"Tourbillon: qué es y cómo funciona esta complicación"*.

## Why it has never been noticed

Nothing in the platform renders a social card, and no check fetches one. The tag is *present*
and well-formed, so any test asserting "the page has og:image" passes. **It is the same
check-with-no-failing-branch shape this fleet keeps hitting** — presence was asserted, the
target was never fetched.

## Fix candidates, ordered by what closes the door

1. **Do not emit `og:image` unless the asset exists.** Makes the bad state unrepresentable and
   is strictly better than today: a missing tag lets platforms fall back to their own heuristics
   (often the first in-page image), whereas a 404 tag yields nothing at all. Cheapest, and it
   should land regardless of whether (2) is ever built.
2. **Generate the card.** The imagery pipeline already produces per-site assets; a 1200×630 card
   from the site's logo, palette and title is the same class of derivation as the favicon built
   for relojistas on 2026-07-27. Find the path that produced leopardess's and robot-hands' cards
   first.
3. **Fix the title/description fallback** so `og:title` uses the page `<h1>`/title and falls back
   to the site display name, never to the bare domain, and `og:description` uses the page's meta
   description.

Do (1) even if (2) and (3) wait — an absent tag is better than a broken one.

> **CORRECTED 2026-07-28 (evening) — this ordering is wrong for THIS estate, and measuring is
> what showed it.** The ranking above is sound in the abstract ("make the bad state
> unrepresentable first"), but it was written without checking whether (2) was actually
> available. It is — on every site:
>
> **All 14 live sites have an active `logo` asset**, which is the generator's only
> precondition. So (2) needs *no code, no council, no build, no roll, and no chrome re-render*
> — the tag already points at the right path; the file is simply absent. Whereas (1) needs all
> of those **plus** a head re-render on 14 sites (head is a stored artefact — `bugs_open/117`)
> and a page redeploy, and its outcome is *no* preview rather than a *working* one.
>
> **Measured, not predicted:** the derivation was run for relojistas on 2026-07-28 and the card
> went from 404 to a live 1200×630 PNG in **18 seconds**, with no deploy of anything else.
>
> **(1) still belongs — as the guard, second.** It is what protects a future site whose logo is
> missing or whose derivation failed.
>
> **And a landmine for whoever implements (1):** do **not** gate the tag on "an `og_card` asset
> row exists". That is the obvious design and it would suppress the tag on
> **leopardessconsulting.co.uk — the one site whose preview actually works** — because its card
> was hand-committed and has no row. Whatever (1) keys on must not have that false negative.
> The precedent to follow is in the same function: `render_site_components_action.go:704-712`
> already gates `sprites.css` on an active asset "otherwise the `<link>` would 404 on sites
> without one", while the comment at `:700` waves the og card away as "harmless if they 404
> until derivation runs". Same question, adjacent lines, opposite answers.

## ⚠ Running the generator is NOT sufficient — look at what it produces

Found the hard way on the relojistas pilot, 2026-07-28. The derivation succeeded on every
signal — item `complete`, URL 200, valid 1200×630 `image/png`, provenance rows written — and
**the card is a picture of a brand SPECIFICATION SHEET**: the wordmark twice, side by side, on
a light swatch and a dark swatch. The generator is faultless; relojistas' stored `logo` asset
is simply not a logo.

**For an image artefact, dimensions and MIME type are not "structure" — what the picture shows
is, and the only way to know is to look at it.** `bugs_open/012`'s rule ("check the artefact
after a rewrite, not just the status") applies here in a medium where every mechanical check
passes.

Knock-on, and bigger than this bug: **the same asset is relojistas' live header logo** —
`<img src="/assets/images/logo.jpg" class="logo-img">`, `.logo-img { max-height: 44px; width:
auto; }`, no crop anywhere in the only stylesheet, source 1408×768. So every page renders the
two-up board at roughly 81×44px, both wordmarks illegible. Weeks live. **Filed separately —
see `docs024_key_docs_latest/bugfix_131_og_card/`.**

So the rollout of (2) is **per-site with an eyeball on each result**, not a batch. Spot-checked
`ai-agent-orchestration.com` and `finetuning.uk` — both proper 400×400 marks, so this is not a
fleet-wide input problem, which is exactly why it needs checking one at a time rather than
being assumed either way.

## How to verify

**Not by asserting the tag exists** — that is the bug. Fetch the URL the tag names:

```bash
img=$(curl -s https://<site>/ | grep -o 'property="og:image" content="[^"]*"' | sed 's/.*content="//;s/"//')
[ -z "$img" ] && echo "no tag (acceptable under fix 1)" || curl -s -o /dev/null -w "%{http_code}\n" "$img"
```

Pass = `200`, or no tag at all. Fail = a tag whose target is not 200. Today: 11 of 14 fail.

## Related

- `bugs_open/117` / `118` — same family: machinery that is present and reachable but produces
  nothing, and nothing checks the output.
- The `process_html` → `AddStructuredData` path is a third instance: registered
  (`registry.go:1042`), referenced by 2 agent definitions, and **zero of 14 sites emit any
  `application/ld+json`**, because it only fires when `business_name` is populated.

---

## CONTRIBUTION 2026-07-29 (gauntlet_dead_cta lane, vonc6) — the head block has NO page context, so `og:url` is wrong on every inner page fleet-wide

Not filing a competing bug: `who-owns.py` and this file both say the head block
is yours. This is evidence for the same emitter, found while shipping vonc's
opinion ledger.

**`og:url` is hard-coded to the site root**, `render_site_components_action.go:449`:

```go
origin := "https://" + ctx.Domain
b.WriteString("  <meta property=\"og:url\" content=\"" + origin + "/\">\n")
```

**Measured live 2026-07-29 — 7 pages, 4 sites, no exceptions:**

```
vonc.com/tools/gauntlet/index.html   og:url = https://vonc.com/
vonc.com/tools/arena/index.html      og:url = https://vonc.com/
vonc.com/                            og:url = https://vonc.com/
dartsonline.com/about.html           og:url = https://dartsonline.com/
dartsonline.com/guides/index.html    og:url = https://dartsonline.com/
finetuning.uk/about.html             og:url = https://finetuning.uk/
leopardessconsulting.co.uk/          og:url = https://leopardessconsulting.co.uk/
```

`rel="canonical"` is **absent on all seven** (the emitter never writes one).

**This sharpens fix candidate 3, and candidate 3 as written would not fix it.**
Candidate 3 treats title/description as a *fallback* problem ("falls back to the
bare domain when the display name is empty"). `og:url` shows the defect is
structural and one level up: **`injectBrandHeadTags` takes a site context
(`ctx.Domain`, `ctx.CompanyName`, `ctx.Tagline`) and no page context at all**,
then that single block is injected into every page's head. So a site with a
perfect display name still advertises the homepage's identity on every inner
page — a share of any deep page unfurls as the site front door. Fixing the
fallback makes `og:title` say the right *site*; it cannot make it say the right
*page*.

**Consequence for a lane you may not be tracking:** the owner's 2026-07-29 H
ruling made distribution the next move for vonc.com — the share card and the
daily provocation travel to where people argue. Today a shared Gauntlet link
unfurls as "vonc.com", the site tagline, a 404 image, and the root URL: nothing
that identifies the thing being shared. So this emitter is on the critical path
for that experiment, alongside the missing card itself.

**Suggested addition to your candidate list (yours to accept or reject):**
give the injector the page it is rendering (url, title/`<h1>`, meta description)
and emit per-page `og:url` + `og:title` + `og:description` + `rel="canonical"`;
keep the site values as the fallback, not the value. That is a signature change
on a shared renderer — architecture-shaped, not a bug patch — which is probably
why it wants to be one deliberate change rather than three.

**Verify (same shape as your og:image recipe — fetch, don't assert presence):**

```bash
u=$(curl -s "https://<site>/<inner-page>" | grep -o 'og:url" content="[^"]*"' | sed 's/.*content="//;s/"//')
# FAIL today: $u is the site root for every inner page.
```

*Unrelated small fact from the same session, in case it bites your card work:*
vonc's gauntlet serves **only** at `/tools/gauntlet/index.html` — both
`/tools/gauntlet` and `/tools/gauntlet/` are 404. A share of the "tidy" URL is
already dead before any unfurl.

---

## STATUS 2026-07-29 (session relojistas-5) — mostly fixed; what remains is named below

Workstream: `docs/agent_docs/docs024_key_docs_latest/bugfix_131_og_card/`. Read
`SUMMARY_2026-07-29b_og_card.md` first, then NOTES (4)–(10).

**DONE and verified by eye (the only verification that works here):**
- 12 of 13 tagged sites serve a real 1200×630 card (was 2).
- **relojistas is fully repaired at source**: owner-approved crop written to S3 through the
  cluster, `logo` row repointed (path-style HTTPS) and **locked**, header live on the site,
  card + favicon re-derived post-fix. Card is clean and **letterbox-free** — confirming the
  letterbox was an ASSET defect (opaque logo), not a code one.
- **Code fix LIVE on v1.0.1199+, council-APPROVED** (trail `bfd73f71`): favicon derivation
  preserves aspect (`composeFavicon`), and `assets.locked_at` is honoured **before** the git
  commit, with **no `status` filter** — a lock fails closed whatever status the row carries.
  (Round 1's HIGH was right: `assets.status` is unconstrained free text.)
- **leopardess protected deliberately**: locked `og_card` + `favicon` rows backfilled and now
  armed. Its malformed `logo` row is no longer the only guard — but still do not tidy it.
- gaswholesalers + idea.uk were reassigned to another session and are being fixed there.

**STILL OPEN, in the order that matters:**
1. **A wide logo cannot make a legible favicon, and 5 of 14 sites have one.** The distortion is
   fixed; the illegibility is not. relojistas' repaired favicon puts 19px of ink in a 64px
   canvas and renders as a grey smudge at true 16px tab size (measured, not guessed). The real
   fix is a **square favicon source** — the mark, not the wordmark — which `derive_brand_head_assets`
   cannot express: it always reads `asset_key='logo'`. Affects relojistas, fundamentallyai,
   oufe, robot-hands, vetcomparison. **Deliberately did NOT re-derive the other four**: it
   would buy undistorted-but-equally-illegible icons at the cost of four slots in a busy queue.
2. **Fix 1, the tag gate** — unchanged and still worth landing; the original constraint holds
   (**do not key it on an `og_card` row**: leopardess had a working card and no row). Follow
   the `sprites.css` precedent at `render_site_components_action.go:704-712`.
3. **`og:title` is the bare domain on ~8 sites; `og:description` absent on relojistas** — the
   case file's "second defect", untouched.
4. **webdesign.co.uk emits no `og:image`** — `injectBrandHeadTags` skips a head that already
   has `rel="icon"`/`og:image`. Never investigated.
5. Detector blindness → **`bugs_open/142`**. Sibling lock gap → **`bugs_open/143`**.

**Landmine added by this session:** `sites.github_repo` selects which deploy repo serves a site
(`vm-sites` = nginx on a VM; empty = B2 + Cloudflare Worker). **Both repos contain a
`<domain>/` folder for some VM sites, so publishing to the wrong one succeeds with a green
workflow and changes nothing.** relojistas + idea.uk are `vm-sites`. It cost this lane an hour
and a wrong inference; see `WRONG_CALLS.md` 2026-07-29.

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
the emitter is not universally broken — the asset pipeline is. **[UNDIAGNOSED]** which path
produced those two; find it before building a new generator, because it may only need wiring.

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

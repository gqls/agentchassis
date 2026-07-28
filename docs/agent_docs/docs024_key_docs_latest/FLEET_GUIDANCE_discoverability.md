# FLEET GUIDANCE — discoverability: being found, quoted and cited

**Written 2026-07-28 from the relojistas.com investigation. Applies to every site in the
estate.** Evidence: `traffic_probe/EVIDENCE_2026-07-28_crawl_budget_and_the_dead_forum.md`.

The finding that prompted it: a site publishing curated content **every day** was very nearly
invisible — real search and AI crawlers spent **78% of their requests on URLs that 404**, and
only **seven requests in 33 hours** reached any live article. Nothing about that was specific to
relojistas. It was measured there first because it was the first site whose real client IPs we
could see.

---

## 1. The three files every site should serve — and the tool that makes them

```bash
./scripts/site-discovery-files.py <domain>            # dry run, prints what it would do
./scripts/site-discovery-files.py <domain> --write    # emit; committing is yours
```

Generates `robots.txt`, `sitemap.xml` and `llms.txt` from the `pages` table. Registered as
**SEO-002** in the concept register.

**Three rules it enforces as behaviour, not advice** — each cost real time to learn:

1. **Probe before listing.** Every URL is fetched; only 200s are emitted. *A sitemap advertising
   a 404 is worse than no sitemap.*
   **But the probe is POINT-IN-TIME.** On its first fleet run it dropped
   `oufe.com/cases/thames-water.html` as 404 — correct at that moment, and the page was
   **deployed 1.5 hours later** (`deployed_at 07-28 16:38` vs a ~15:0x probe). *I first wrote
   this up as the tool "catching a dead page", which implied a defect on oufe's side. It was a
   site mid-build.* Run it when the target is not building, and read a dropped URL as **"not
   fetchable right now"**, never as "broken".
2. **`llms.txt` is built FROM the pages, not written ABOUT them.** Each entry is the page's own
   `<h1>` and its own first sentence. Nothing invents a description of a site — that is how
   unsupported claims get published, and this estate has a whole verification layer devoted to
   stopping exactly that.
3. **It checks who is actually serving `robots.txt`** (see §2), and names the agents currently
   disallowed.

**Status across the fleet, 2026-07-28:** relojistas.com has all three, live. Every other site
has none. Running the tool is cheap; the judgement is in the `--signal` and `--summary` values.

---

## 2. Cloudflare's managed robots.txt — the trap, and both settings

**Cloudflare PREPENDS its managed `robots.txt` to the origin's. It does not yield to it.**
Shipping your own file changes nothing about blocking until the dashboard is changed. Measured:
after shipping ours, the served file was **82 lines with two `User-agent: *` groups and two
contradictory `Content-Signal` values**, and ClaudeBot/GPTBot/CCBot were still `Disallow: /`.

**It is two settings, under `Manage AI bot access`, and it is PER ZONE:**

| setting | what it does | set to |
|---|---|---|
| **Block AI training bots** | a managed WAF rule that **actually blocks requests at the edge** — enforcement, not advice | *Do not block (allow crawlers)* |
| **Set your preference to block training in robots.txt** | injects the managed block that gets prepended | *no preference / do not manage* |

Only the second stops the merge; only the first stops the enforcement. **Both, or the effect is
partial — and the failure is silent, because a blocked crawler never appears in your log at all.**

**Verify PER AGENT. Never with a single `curl`** — the file is served conditionally, so one
fetch tells you nothing about any particular crawler:

```bash
for ua in ClaudeBot GPTBot OAI-SearchBot PerplexityBot CCBot Googlebot; do
  printf "  %-18s " "$ua"
  curl -s -A "Mozilla/5.0 (compatible; $ua/1.0)" "https://<domain>/robots.txt?x=$RANDOM" \
   | awk -v n="$ua" 'BEGIN{IGNORECASE=1} $0~"^User-agent:[ ]*"n"[ ]*$"{f=1;next}
       f&&/^Disallow:[ ]*\/[ ]*$/{print "DISALLOWED";x=1;exit} f&&/^User-agent:/{f=0}
       END{if(!x)print "allowed"}'
done
```

**Owner ruling 2026-07-28 (relojistas): allow everything, including training** —
`Content-Signal: search=yes, ai-input=yes, ai-train=yes`. Reasoning below in §4. **Applied to
the relojistas zone only; every other zone still carries the block.**

---

## 3. What is actually worth having, ranked — and one claim withdrawn

> **WITHDRAWN: "the dead URLs are eating the crawl budget".** Crawl budget is a large-site
> concept; Google's own guidance is that it is not the limiting factor below a few thousand
> URLs. relojistas has 18 pages and will be crawled regardless. The 2,942-vs-38 measurement is
> real; the *causation* was not established. Stated here because it was repeated in three
> documents before being checked.

What actually matters, in order:

1. **A sitemap.** Discovery was the binding constraint — 18 pages reachable only by walking the
   homepage. This is the cheap, high-value one.
2. **Not being blocked.** §2. Free, and observed to matter: ClaudeBot made 24 requests to
   relojistas, **all of them `robots.txt`, zero pages**. It asked, was refused, and left.
3. **`410 Gone` on a genuinely dead URL surface** — housekeeping, not an unlock. 404 means "try
   later"; only 410 makes a crawler drop it. Worth doing for server load and for not looking
   abandoned. **Do not `Disallow` those paths instead** — blocking them stops crawlers ever
   *seeing* the 410, so they stay indexed forever.
4. **Structured data and a working `og:image`** — §5.

---

## 4. Training vs citation — the distinction to decide per site

Two different transactions, routinely conflated:

| | what you give | what you get back |
|---|---|---|
| **ai-train** | content absorbed into weights | no link, no attribution, no referral. Slow, and it makes you *known* rather than *visited* |
| **ai-input / reference** | content fetched at answer time | **quoted and cited with a clickable link** — a real referral channel, working today, no training cycle |

**Both are legitimate positions.** Refusing training and allowing citation is coherent and is
exactly what `search=yes, ai-input=yes, ai-train=no` expresses.

**The argument that decided relojistas:** a `robots.txt` block is *voluntary compliance* — it
binds only crawlers that read it. On relojistas, a 184,000/day scraper fleet ignored it entirely
while ClaudeBot obeyed it and left. **So a blanket block stops the crawlers that would attribute
and cite, and does nothing to the ones that take without credit.** It filters for good behaviour
and penalises it.

**Do not expect training to send readers.** The mechanism the owner asked about — *"if we're the
only ones with it, we'll be mentioned"* — is real but weak: models rarely attribute from
parametric memory, and the dominant input to a domain being *known* is how often **other** sites
mention it, not how much it publishes about itself. Allow it if you like; invest in being
retrievable.

---

## 5. Two fleet defects found while doing this

- **`bugs_open/131` — `og:image` points at a card that was never generated.** 11 of 14 live
  sites emit an `og:image` whose target 404s. Every WhatsApp/Slack/X share of those sites shows
  no preview. `og:title` also falls back to the bare domain on 8 sites.
- **Structured data is dormant, not missing.** `process_html` is registered
  (`registry.go:1042`) and calls `datahelpers.AddStructuredData`, but that only emits when
  `businessInfo["business_name"]` is populated — and **zero of 14 sites carry any
  `application/ld+json`**. Registered, reachable, silently producing nothing. Fixing that is the
  fleet-wide answer; it needs a Go change and a roll.

> **RETRACTED 2026-07-28, hours after writing it — the DB-only pattern CANNOT WORK. Do not
> attempt it.** I built a dedicated `structured-data-block` component, attached it to
> relojistas' glossary index with a valid `DefinedTermSet` of 8 terms, and it rendered
> correctly — `page_components.rendered_html` was **2,646 bytes containing valid JSON-LD**.
> **It never reached a page.**
>
> The page assembler drops it, correctly and by design. `getPageSections`
> (`rerender_single_page_action.go`) calls `sectionHasVisibleContent(html)`, which **strips
> `<style>` and `<script>` blocks, then requires more than 10 characters of remaining text**.
> A JSON-LD block is `<script type="application/ld+json">` and nothing else — it leaves
> **zero**. So a metadata-only section is structurally incompatible with this assembler and
> always will be. The guard is a good one: it exists because nine blanked article bodies once
> vanished unnoticed.
>
> **How it hid:** every DB-side check said success — component rendered, payload intact,
> JSON-LD present in the stored HTML — and the deployed file simply never contained it. The
> tell was in the *deploy* history, not the render: `glosario/index.html` had not been written
> since the previous day while other pages deployed normally that morning. **When a render
> "completes" but nothing changes on the page, check whether a deploy artefact was written at
> all before investigating the render.**
>
> There is an escape hatch — `reRuntimeFill`, which keeps sections that carry a runtime-fill
> marker. **Do not use it for this.** That marker means "a browser-side loader fills this",
> and borrowing it to smuggle metadata past a content check would make the section lie to
> every other consumer of that signal.
>
> Component left in the library but `is_active = false`, with the reason in its description so
> nobody rediscovers it and repeats the experiment.
>
> **JSON-LD belongs in `<head>`** — where the `og:` tags already are, emitted by
> `render_site_components_action.go`. That is a Go change and a roll, and it is the only route.
> It is the same fix as the dormant `AddStructuredData` above: one place, whole fleet.

---

## 6. Reviving a dead URL surface — when it is content strategy and when it is a doorway

relojistas inherited a dead vBulletin forum whose URLs are still crawled hard. The owner's
question — *populate those URLs with great content* — is right in part, and the distinction
matters more than the tactic:

**The test: does the content genuinely serve what the request was for?**

| old URL | what the requester wanted | honest answer |
|---|---|---|
| `/faq.php` (480 real-crawler hits) | a FAQ about watches | **Publish a real FAQ.** It is our domain and a genuinely useful Spanish watch FAQ *is* what that URL promises. Legitimate, and it captures live demand. |
| `/search.php?q=X` | a search for X | **301 to `/buscar?q=X`.** Same function, honestly mapped. |
| `/showthread.php?t=<id>` | one specific dead thread | **No.** We cannot know what thread 41,233 was about. Serving arbitrary articles to capture the hit is a **doorway page** — the pattern search engines penalise — unless the topic is genuinely recovered (the Wayback board map in `EVIDENCE_2026-07-25` is the only honest route). |
| `/attachment.php?attachmentid=<id>` (184k/day) | one specific image we do not have | **410.** Optionally serve a small branded placeholder *only when a genuine third-party referer is present* (`foroderelojes.es` still embeds these in live threads) and 410 everything else. |

**The line:** publishing a real FAQ at `/faq.php` is content strategy. Generating 25,000 pages to
absorb `attachmentid` hits is a doorway scheme that would put the domain at risk. The volume of
requests is not the justification — **whether the answer is true to the question** is.

---

## 7. Measurement discipline this whole investigation kept teaching

- **Per-source analysis needs real client IPs.** Behind Cloudflare every address is an edge
  address; none of §1–§6 was measurable on relojistas before real-ip landed on 2026-07-27.
- **A check that names the defect it expects cannot fail.** Three instances here: a phantom-link
  grep searching for the two links already fixed; `grep -c 'mailto:'` on a CF-proxied site (CF
  rewrites them, so it is always 0); and `grep -c buscar-item` to prove search works, which
  matches the CSS that *defines* the class, so a nonsense query "passes".
- **Assert the target, not the tag.** `og:image` is present and well-formed on 11 sites and
  points at a 404. Fetch what the markup names.
- **When the answer is "nothing happens", ask whether the input was ever non-empty.** The first
  conclusion here — "crawlers do nothing with our content" — was drawn while every crawler was
  receiving a 404.

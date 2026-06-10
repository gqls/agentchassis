# TASK BRIEF — Domain traffic-probe (intent discovery on parked domains)

**This is a SEPARATE project from idea.uk.** It reuses idea.uk's VM/nginx/Go deployment
model, but it is not the AI-chat/ideas work and should run in its own chat. This brief plus
the bundled files are meant to be enough to start that chat cold.

---

## 1. Objective

Several domains still receive residual visitor traffic even though they no longer serve any
content (they sit on a parking lander). The question is: **what are those visitors actually
looking for?** Some of these domains previously had real sites — possibly with content behind
a login or paywall — and people still arrive expecting it.

The plan: host the domains on the idea.uk VM/nginx/Go model, replace the parking lander with a
small page that (a) plausibly reflects what the domain was about, and (b) invites the visitor
to do one thing — a search, a category click, or a short "what were you after?" prompt — so we
capture their **stated intent**. We then use that to decide which domains are worth building out
(idea.uk-style) or otherwise using.

To be precise about scope: we are capturing what visitors *say they want* via an action they
take on our own page. We are not trying to recover anyone's old gated content.

---

## 2. Method — archive.org first, then probe

**Step 1, per domain: look up what it used to be (Wayback).** This tells us how to make the
probe page plausible and what action to invite. The path list is the most useful part — if the
old site had `/login`, `/members`, `/subscribe`, `/forum`, `/account`, that signals what was
gated and what visitors may still want.

Exact Internet Archive endpoints to use (no key needed):
- **CDX (historical URL/path list):**
  `https://web.archive.org/cdx/search/cdx?url=DOMAIN&output=json&from=2016&to=2024&collapse=urlkey&limit=200`
  (drop the `collapse`/`limit` to see everything; add `&matchType=domain` to include subpaths)
- **Availability (closest snapshot):**
  `https://archive.org/wayback/available?url=DOMAIN&timestamp=20190101`
- **View a snapshot:** `https://web.archive.org/web/TIMESTAMP/http://DOMAIN/`
- Reading one or two snapshots tells you the vertical, language, and what the site offered.

(Tooling note for the new chat: in this assistant, `web_fetch` only opens URLs that came back
from a `web_search`/`web_fetch` result, so search the domain first, or run the CDX calls from
the box/your laptop with `curl` and paste results in.)

**Step 2: build a minimal probe page per domain.** It should look intentional, not like a
parking page: a one-line description matching the old vertical, and ONE invited action:
- a search box ("What are you looking for on DOMAIN?"), or
- 3–5 category links derived from the old site's sections, or
- a short free-text "what brought you here today?" field.
Log the action server-side as a structured intent event.

**Step 3: measure.** Per domain, track intent events per 1,000 visits, the actual search
terms / categories chosen, referer (coarse), and country. After a set run (say 2–4 weeks),
rank domains by how clearly visitors expressed a want we could serve.

---

## 3. Reuse plan (the idea.uk model)

Fork idea.uk's Go service rather than starting fresh. From `golang_files/`:
- **Keep / adapt:** `service.go` (HTTP server, handlers, the `a.page()` HTML wrapper, `writeHTML`),
  `store.go` (file-based persistence — repurpose for intent events), `page.html` (becomes a
  per-domain template), `main.go`, `go.mod`, `deploy/` (env + setup), the nginx + systemd +
  Let's Encrypt setup, and the B2 checkpoint-upload pattern (for shipping captured data off-box).
- **Drop (idea.uk-specific):** `engine.go`, `prompts.go`, `audience_check.go`, `billing.go`.
  (They're bundled only as reference for the Go + Anthropic-API + prompt-caching pattern, in case
  you want the probe to auto-summarise a domain's Wayback content or auto-classify intents.)

**Multi-domain on one box (cheaper path):** one binary, nginx with a `server_name` block per
domain all proxying to it, and the page chosen by the `Host` header. store.go keys events by
host. The alternative — one systemd unit per domain — is heavier; only worth it if domains need
isolation. Start with the single-binary-multi-vhost approach.

**Prerequisite:** you must control DNS for each domain (point an A record at the box). The
marketplace eligibility statuses in the domain list (Pending Review / Opt-in Required / TLD Not
Eligible / Registrar Not Eligible) concern the *parking program's* monetisation, not DNS — but
repointing DNS will stop any parking revenue while a domain is on the probe, so choose
deliberately.

---

## 4. Privacy / legal (UK, low risk appetite)

- Real human visitors → UK GDPR + PECR apply. Keep it minimal: log server-side only, no
  third-party trackers, and **no non-essential cookies** (if you store nothing on the device,
  you avoid needing a consent banner). Don't collect names/emails; a search box is fine, but
  treat free-text as potentially personal and don't retain longer than needed.
- Put a plain privacy line on the page (who runs it, what's logged, retention).
- Check each parking program's terms before repointing (the list looks like a Dan/Afternic/
  Sedo/Bodis-style export). You control the domains, but read the opt-out terms.
- Start with **3–5 high-traffic, clearly generic domains you fully control** and expand once the
  capture + privacy posture is proven.

---

## 5. Prioritised shortlist (by estimated visits — verify each in the new chat)

The figures are the marketplace's own estimates (relative ranking, not audited). `relojistas.com`
at ~1.2M is a big outlier — treat with scepticism until Wayback + your own logs confirm it.

| est. visits | domain | likely vertical (name-based; confirm via Wayback) | probe idea |
|---|---|---|---|
| 1,201,799 | relojistas.com | ES "watchmakers / watch repair" — likely old directory/forum or heavy type-in | watch brand / repair search |
| 28,822 | wayfaringlondoner.com | London travel blog | destination / trip-guide search |
| 3,285 | traderboltai.com | AI trading bot (registered Dec 2025, "Active") — recent, traffic may be launch/bot | market / feature interest |
| 3,104 | surgerylight.com | surgical / operating lights (B2B medical) | product / spec enquiry |
| 1,489 | kitchensep.com | unclear — confirm via Wayback | — |
| 880 | zdec.com | short/brandable, unclear | — |
| 523 | komunikatif.com | ID/MS "communicative" — comms/PR? | service search |
| 335 | fallingwaterdesignbuild.com | architecture / design-build firm | project / portfolio enquiry |
| 234 | hoeinvestereninvastgoed.com | NL "how to invest in real estate" | investing-guide search |
| 224 | makeitaquote.com | quote-image generator OR insurance quote — disambiguate | depends on Wayback |
| 150 | vinrose.com | wine (rosé) or a personal name | — |
| 141 | buysportskit.com | sports kit / teamwear retail | product search |
| 130 | thecentralbanker.com | finance / central-banking news | topic search |
| 122 | monitorizare.com | RO "monitoring" — CCTV / surveillance? | service search |
| 114 | lodgeswithhottubs.club | UK holiday lodges with hot tubs | location / date search |
| 109 | equitycalculator.co.uk | UK home-equity / release calculator | run the calculator |
| 106 | designs.co.uk | generic high-value | what kind of designs |

Health-adjacent names in the wider list (healthscare.*, surgeryhealthcare.com, plantarproblems.com,
overpronation.com) need careful, non-clinical framing if used.

---

## 6. What's in this package

- This brief.
- `traffic_probe_domains.tsv` — the full domain list, ranked by estimated visits + the verbatim export.
- The reusable Go service (`golang_files/`: service.go, store.go, page.html, main.go, deploy/, etc.).
- Deploy / persistence / VM docs: idea.uk architecture & deployment, VM launch plan, box setup,
  persistence design, and the B2 checkpoint-upload plan.

## 7. Open questions for the new chat

- Single binary + Host-header multiplex vs one unit per domain (lean towards the former).
- Capture style: search box vs category links vs free-text — possibly A/B per domain.
- Run length and the success metric (intent events per 1k visits; clarity of expressed want).
- Data retention period and where captured events live (on-box store.go + periodic B2 push).
- Which 3–5 domains to start with (suggest relojistas.com, wayfaringlondoner.com, surgerylight.com,
  plus one finance tool like equitycalculator.co.uk and one clear retail like buysportskit.com).
- Whether to verify the marketplace visit figures against a second source before committing effort.

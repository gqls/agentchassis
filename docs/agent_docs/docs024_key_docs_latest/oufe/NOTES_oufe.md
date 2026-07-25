# NOTES — oufe.com / oxenunity.com

Append-only. Newest at the bottom. Missteps and wrong turns belong here as much as
successes — more, in fact.

---

## 2026-07-25 — workstream opened

**Origin.** The owner developed the proposition with Gemini, then asked it four
times to export the conversation as a running-notes markdown file. Three attempts
returned bare errors ("I'm having a hard time fulfilling your request", "I seem to
be encountering an error"). The fourth returned a document that had lost the
earlier strategic reasoning — the audience analysis, the "what can I offer that
terminals don't" argument, the phasing rationale — and leaked its own Python
`with open(...) as f: f.write(md_content)` source into the visible answer. The
conversation was pasted into a Claude Code session instead. **This directory is
that record**, reconstructed. PLAN §1 holds the decisions, §2 the challenges.

**Prior-art sweep first.** Case-insensitive, binary-safe grep for `oufe`,
`oxenunity`, `oxen unity`, `financial engineer` across the repo (including
`bugs_open/`, `bugs_closed/`, `features_open/`, all of `docs/`), the auto-memory
directory (93 files), and ~200 session transcripts. **Zero real hits.** The ~20
transcript matches were case-variant fragments (`OUFe`, `OUfe`) inside base64
blobs — verified by extracting them. Greenfield, no other session has this.

**Three corrections to the capabilities inventory the owner had pasted into
Gemini**, all of which changed the plan:

1. **V5 is not "DESIGNED".** It was built, seeded and activated on v1.0.1140
   (2026-07-20); `evidence-researcher` is an active agent; its blocker
   (`bugs_open/047`) closed 2026-07-21. But the end-to-end smoke run was never
   repeated after that fix, so it is *live and never successfully exercised*.
   oufe becomes its first real test.
2. **go-echarts does not exist.** Gemini's whole charting architecture assumed it,
   because the inventory said so; that line was corrected on 2026-07-24 (not in
   `go.mod`, no chart action, register marks data-charts "aspirational — not
   started"). Only `report_charts.go` exists, purpose-built inline SVG for one
   report page. Doctrine intact, renderer absent.
3. **The deterministic number-scan is near-inert for finance prose.**
   `businessClaimContextRe` (`datahelpers/claims.go:334`) carries no debt,
   creditor, recovery or covenant vocabulary, and `isExcludedNumber` excludes
   currency amounts outright. So "£16bn of Class A debt" is never scanned. Already
   recorded for relojistas.com; it lands much harder on a site whose entire
   subject is money figures. **Do not read a clean claims report on this site as
   "no invented numbers".**

**One challenge to an owner decision, recorded because it inverts his stated
ordering.** He wrote *"first direction 3 as that is lowest risk"* (the automated
distress radar). It is the highest-risk first move available: no market-data feed
exists anywhere in the platform, UK dockets have no API, and a wrong distress
signal is a factual assertion about a named real company — the exact class as the
vetcomparison incident. The genuinely low-risk start is the thing he separately
named as the primary magnet: one flagship dossier plus one excellent tool. Argued
in PLAN §C1; the owner has not yet responded to this specific inversion.

**Decisions taken with the owner this session:** first slice = docs + oxenunity
live + oufe P1 skeleton; oxenunity presents a wordmark and a link with **no entity
claims at all** (rather than an explicit "not a company" statement) — a page that
claims nothing cannot be untrue about a company that doesn't exist.

**Deliberate deferral recorded so it isn't mistaken for an oversight:** no news
feed at launch. The classifier will read this site as `finance` and the vertical
map would seed generic financial-markets / interest-rates keywords with a separate
news page — the opposite of a specialist restructuring publication, and it spends
credits per fetch. There is no `restructuring` / `insolvency` / `corporate-finance`
vertical in the map; adding one is a fleet-wide Go change.

---

## 2026-07-25 — oxenunity.com shipped to B2; the two domains are in OPPOSITE infra states

Page authored (`sites/oxenunity.com/index.html`, 111 lines, hand-written), pushed
to `gqls/sites` master, **Deploy to B2 action completed success in 21s**, object
confirmed present:
```
$ b2 ls b2://portfolio-sites/oxenunity.com/
oxenunity.com/index.html
```
The local `sites` checkout was **1,532 commits behind origin** — fast-forwarded
before committing. Worth checking every time; this repo takes rerender commits
from every session in the fleet all day.

**CORRECTION to my own plan (PLAN §6 item 2 / the owner checklist).** I had
written the Cloudflare wiring as one blocking step covering both domains, on the
fundamentallyai precedent. Measured, it is not one step and it does not cover both:

| | oufe.com | oxenunity.com |
|---|---|---|
| NS | `leah` + `alexis.ns.cloudflare.com` (our fleet pair) | `*.ns.porkbun.com` (registrar) |
| A | 104.21.85.181 / 172.67.208.225 (Cloudflare) | 207.207.210.36 / .50 |
| Serving | **our Worker, already bound** | openresty 302 → `oxenunity-com.l.ink` (parking) |
| Needs | **nothing — content only** | full zone move + Worker route |

The proof that oufe.com's Worker route is already bound is its 404 **body**, which
is our own Worker's error JSON, not a Cloudflare 404 page:
```json
{"error":"B2 returned error","objectKey":"oufe.com/index.html","status":404,
 "body":"…<Code>NoSuchKey</Code>…"}
```
It is looking in exactly the right bucket prefix and finding nothing there, which
is correct — `b2 ls b2://portfolio-sites/oufe.com/` is empty. **So the failure mode
that left fundamentallyai.com dark after a successful build cannot happen to
oufe.com: the moment content lands in B2, it serves.** That removes the only infra
item from oufe's critical path.

oxenunity.com is the reverse: the page is built and in the bucket, and it is
unreachable at its own domain until the zone moves to Cloudflare and the
portfolio-sites Worker route is bound to `oxenunity.com/*` and `*.oxenunity.com/*`.
Owner step, and now the *only* infra step in this workstream.

**Misstep worth logging:** I tried to prove the deployed page by sending
`-H "Host: oxenunity.com"` at oufe.com's Cloudflare edge, expecting the Worker to
key off the Host header the way its code does. Cloudflare returned **403 at the
edge** — host/SNI mismatch is rejected before any Worker runs. The check was
never going to work, and a 403 there says nothing about the Worker. `b2 ls` was
the honest check and I should have gone there first.

---

## 2026-07-25 — oufe.com seeded and submitted; a measured instance of bugs_open/030

**Seeded before submission** (`SEED_2026-07-25_oufe_site_and_specs.sql`, applied
out of band, not through the migration runner — it is per-site setup, not a schema
change). Site id `a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39`. Three rows: the `sites`
row with a contact email (bug 063 fails open without one), `evidence_base` with
**18 banned patterns and zero facts**, and `imagery_style_guide` (bug 027).

**The banned list was tested before it went in, not after.** 18 patterns against 20
sentences: 10 fabrication shapes I want blocked, and 10 pieces of legitimate
mechanism prose the site must be able to publish. Result 10/10 and 0/10 — every
fabrication caught, no false positives. This mattered because a banned-claim hit
is a **blocker**, and a blocker fails the whole page build: on the last fresh
domain five of nine pages died at `validate_content`. A too-eager regex here would
have looked exactly like that failure and cost hours to attribute.

The patterns target **shapes**, not numbers — recovery percentages asserted as
fact, trading levels ("cents on the dollar" — we have no market data, so no level
is knowable to us), predictions stated as fact, promotion language, fabricated
sourcing ("people familiar with the matter"), audience-scale and tenure claims.
Five literal figures from the Gemini transcript are banned by value; if one turns
out to be real, the ban forces a conscious return here after registering the fact
with a citation. That friction is deliberate.

**Wrote `TRIGGER_submit_tier3.sh`** because `082_submit_domain_unified.sh` has no
roadmap flag, and without a roadmap brief the planner falls back to
`ensure_pages ["index","contact"]` plus whatever it invents. The shape that
matters and is easy to get wrong: **both briefs are objects with a `text` key.**
The prompt renders `{{.site_specs.specs.roadmap_brief.text}}`, so a bare string
renders empty and the roadmap is silently ignored — you would get a plausible
twenty-page site and no error anywhere.

Submitted 16:56, correlation `e916f41b-a534-4b12-883f-411312ee7ad8`.

### The message did not get consumed, and chasing that was worth it

Ten minutes later: no orchestration row, no work items, no specs beyond my seeds —
while *other* orchestrations created after mine had already completed. Per the
runbook that is the signal that says it is not simply queue latency.

Walking it back, in order, because each step ruled out a different explanation:

1. **Was it published?** `kcat -C -o -6` — yes, offset 104104, all headers
   correct. **Misstep:** my first look used `grep -o '"domain":"oufe.com"'` and
   returned two hits, which I briefly read as a double-publish. It was one grep
   matching twice inside a single 6,767-byte message. Count messages by offset,
   never by string occurrences.
2. **Was the payload valid?** Dumped the message and parsed it — valid JSON,
   `agent_type: domain-submitter`, both briefs objects with text of 3,696 and
   2,828 chars. (The dump file was double the message size: kcat's `-c 1` did not
   limit as expected, so the parse failed on "extra data" until I split on the
   first newline. A dump artefact, not a topic artefact.)
3. **Was it consumed and failed?** `processed_messages` — no row. But neither had
   the scheduler messages immediately after it, and those schedulers were
   demonstrably running, so **`processed_messages` is not a reliable
   was-it-consumed oracle** and I nearly drew a false conclusion from it.
4. **The decisive check — consumer group lag:**
   ```
   generic-requests-group  system.agent.generic.requests  0
     current-offset 104102   log-end 104115   LAG 13   agent-chassis-…zjh4t
   ```
   The consumer is **wedged at 104102**. My message at 104104 is ahead of the
   commit point: never consumed, not dropped, not malformed.
5. **What is at 104102?** A **20,178-byte `council-gate` submission from another
   session** (robot-hands gripper dossier, round 2). A 16-seat council run takes
   minutes, and `system.agent.generic.requests` has a single consumer, so
   everything behind it waits.

**So this is a measured instance of `bugs_open/030` (dispatch queue
serialisation), and specifically its head-of-line blocking mode**: one long
council run stalled 13 unrelated messages for ~9 minutes and counting. Nothing is
wrong with the submission and **resubmitting would be the wrong move** — it would
put a duplicate behind the same blockage. Waiting.

Worth adding to the 030 picture: the memory note for that workstream records the
lane as "93% cron" and the remaining symptom as "~8min stalls". This is a
non-cron instance with a named cause (a large council submission), which is a
sharper statement of the mechanism than "stalls" — the queue is not slow, it is
strictly serial behind whatever is longest.

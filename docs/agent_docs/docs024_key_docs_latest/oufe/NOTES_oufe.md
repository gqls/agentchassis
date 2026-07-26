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

### Resolved, and the diagnosis held exactly

The blocking council reached `complete_revise | COMPLETED`. Within seconds the
committed offset moved 104102 → 104105 and the submission at 104104 was consumed.
**Total wait ~28 minutes, entirely behind one unrelated council run.** No
intervention, no resubmission, nothing wrong with the message. Had I resubmitted
on the "no orchestration row" evidence, the site would have been submitted twice.

Cascade confirmed started:
```
aspect              source_agent      text_chars
submission          domain-submitter
mission_brief       domain-submitter  3696
roadmap_brief       domain-submitter  2828

item_type              status   handler_agent                priority
needs_domain_research  triaged  domain-research-classifier   5
```
`roadmap_brief` persisting with 2,828 characters is the part worth noting: this is
a **Tier-3 submission carrying an authoritative roadmap, which the shipped trigger
script cannot produce.** The item lands at `triaged`, which is what
build-pipeline-trigger dispatches on, so the chain now self-advances:
classifier → strategist → briefing → planner → pages.

**One lesson to carry:** `processed_messages` looked like the natural
"was it consumed?" oracle and would have sent me the wrong way — neither my
message nor the scheduler messages around it had rows, while those schedulers were
demonstrably running. **Consumer-group lag is the oracle**; `processed_messages`
records a narrower path, and logs `DEDUPE_SKIPPED_NO_REQUEST_ID` when it records
nothing at all.

Also proven end to end while waiting: the oxenunity object in B2 is
**byte-identical** to the source file (2,767 bytes, `diff -q` clean). So authoring
→ commit → Action → B2 works for this workstream; what remains for that domain is
purely DNS.

---

## 2026-07-26 — both sites live; the site made a promise we had just retired

**oxenunity.com is live.** The owner moved the nameservers; the `.com` delegation
now points at the same `alexis`/`leah.ns.cloudflare.com` pair as oufe.com and the
page serves HTTP 200. **Misstep:** my first check said the NS were still porkbun
and I nearly reported it as not done — that was my own resolver holding a cached
answer. Query the TLD (`dig +norec @a.gtld-servers.net <domain> NS`) and the
target nameservers directly; a local `dig +short NS` tells you what your resolver
remembers, not what the internet is delegating.

**oufe.com is live.** The cascade completed overnight: identity → classification →
content_direction → design_intent → vertical_landscape → strategy → briefing →
resolved_composition → site plan → pages → imagery. `index` and `about` are
deployed and serving; `cases-index` and `thames-water` are parked at
`needs_human_review` ("not_built"), which is the 149 fix working as designed
rather than a silent no-op.

**The Tier-3 roadmap held exactly.** Five pages planned, and they are the five I
specified — index, about, cases-index, thames-water, contact. No invented extras.
That is the roadmap-brief authority doing its job, and it is worth recording
because the default outcome is a plausible twenty-page site.

**The figure rail held completely.** Scanned all current generated specs (~50KB of
identity/content_direction/strategy/briefing/classification) and both live pages:
**zero currency amounts, zero percentages, zero thousands-separated numbers.**
Keeping every figure out of the briefs prevented the 043 spec-poisoning class
outright. The about page even says, unprompted, "We're new. We have no readership
figures to cite and no track record to invoke" — the mission brief's honesty
instruction propagated into the voice.

### But: the copy adopted the exact promise the owner had just struck

The owner struck "every figure here links to the document it came from" that
morning. The site, written the night before, had independently arrived at a
*stronger* version of it and put it in six places:

> "Every factual claim is sourced to a named, dated primary document."
> "A claim without a named, dated source does not appear here."
> "If we can't trace a number to a document we hold, it doesn't appear here."
> "This discipline is not a disclaimer. It's the method."

That is a promise of infallibility of process, and we cannot keep it. Fixed in
`FIX_2026-07-26_fallibility_copy.sql` across six blocks on the two live pages,
plus the tools paragraph, which described the tools as "scenario illustrators"
that "do not produce valuations or predictions" but **never said they can be
wrong** — which was the owner's specific instruction.

**The finding worth keeping: the claims machinery could never have caught this,
and it is not a bug that it didn't.** Every layer we have — banned patterns, the
writer whitelist, the number scan, V5 citations — polices *claims about third
parties*, principally numbers. This was a claim about **us**, and a qualitative
one. **A promise of infallibility is a different failure class from an invented
figure, and nothing in the estate looks for it.** On a site whose whole positioning
is epistemic honesty, that is the class most worth watching, and it is the class
with no automation behind it. Human reading is the only control.

Method note for the fix: `content_data` edits alone do not change the page —
`rerenderLoadSections` stitches the STORED `rendered_html`. The
`section_data_resolved` reason on `049b_deploy_single_page.sh` re-renders every
section from stored content_data through the current template with no LLM call,
which is what makes an authored copy edit stick. Its documented gotcha was checked
first: **any section with NULL content_data escalates the whole page to the
content writer and regenerates the copy** — all seven sections here were non-NULL.

### Live-site defects found, not yet fixed

Every content link on the homepage is broken. `/cases`, `/cases/index.html`,
`/cases/thames-water`, `/tools`, `/framework`, `/restructuring-plan`,
`/creditor-waterfall` — all 404, including the header's own **Cases** nav item.
The homepage advertises six sections that do not exist, which is the
`bugs_open/052` class (listings re-advertising pages never built).

The platform caught part of this itself: two `unresolved_cta` items are already
parked at `needs_human_review` for the hero and call-to-action secondary CTAs. It
did **not** flag the six info-card links or the Cases nav item, which is a real
coverage gap in CTA integrity worth reporting to that workstream — the cards are
anchors with resolvable-looking hrefs to pages that were planned but never built.

Also live: a **"Get Started"** header CTA pointing at `/contact.html`, and a
footer heading **"Our Services"**, on a site with nothing to sell and a roadmap
brief that explicitly said nothing may offer or imply a purchase. Both are
commercial-shaped furniture that the component templates supply by default.
Neither is dishonest exactly, but both are wrong for this site.

---

## 2026-07-26 (later) — the honesty rule made generic; homepage and chrome repaired

**The compliance council seat now looks for this class** (owner direction).
Migration 223 adds two contracts to `review_compliance`:

- **OVERCLAIMED RELIABILITY** — a promise about our own accuracy is a claim like
  any other, and it is the one class every scanner misses. Seat prompt names the
  four phrases oufe actually shipped, and requires the honest posture instead.
- **ILLUSTRATION, NOT AUTHORITY** — lead with mechanism; a real named case is
  clearly-marked illustration, "a possibly inaccurate case study", not a
  definitive account; a tool must say it can be wrong, and the caveat belongs
  inside the result so it survives a screenshot.

Judge clauses (e) and (f) added, and the seat is explicitly asked to **suggest**
the honest wording rather than only name the fault — the owner's ask.

Seated on `fix-proposer` and mirrored with `099_SYNC_gate_roster.py --apply`, per
CLAUDE.md (never hand-patch the gate). The mirror reported drift ==
`['review_compliance']` exactly, 16 seats both sides, routing OK, snapshot taken.
**Verified afterwards that each roster kept its OWN context block** (fix-proposer:
diagnosis; council-gate: rationale) — that swap is precisely what a hand-patch
would have destroyed.

A matching `STRICT RULE — NEVER PROMISE ACCURACY YOU CANNOT GUARANTEE` went to
both content writers, so the rule shapes copy at generation and not only at
review. **Trap:** the two writers keep their prompt at *different paths* —
`page-content-writer` under `process_sections_loop.sub_workflow.steps
.generate_content`, `content-writer` at the shallower `steps.generate_content`.
The first UPDATE silently matched only one. Verify BOTH, by type.

### Homepage and chrome repaired

Six broken links down to one. Rather than invent destinations, the fix uses the
templates' own fail-closed behaviour: `info-card-grid` wraps its anchor in
`{{if .link_url}}`, so a card with no url renders as plain text. Cards for
unwritten pages now say they are being written. Removed `Get Started` (header)
and `Our Services` (footer) — default commercial shapes on a site that sells
nothing.

The two `unresolved_cta` items turned out to be labels with no `*_url`: hero and
call-to-action both gate on `{{if and .cta_text .cta_url}}`, so those buttons had
been rendering as **nothing at all**. Now pointed at pages that exist.

### Two rendering landmines, both paid for

1. **`049b_deploy_single_page.sh`'s `section_data_resolved` branch cannot work.**
   It sends `{page_id, site_id, domain}`; `rerender_page_sections` requires
   `target_site_id` and `page_name`. Both oufe re-renders FAILED with
   `missing required fields: [page_name]`. Its assemble-only branch never touches
   that action, which is why the gap survived — it only bites on the branch you
   need after editing content_data. Replaced locally by `TRIGGER_rerender_page.sh`.
   **Belongs to the cta_link_integrity workstream to fix at source.**
2. **`slot_name` must equal the component's function name, not `'main'`.** On
   every working page `slot_name` matches the entry in `pages.sections`. My
   hand-inserted cases-index row used `'main'`, matched no section, rendered
   nothing — and the run reported **`COMPLETED | complete_skipped`**. A
   success-shaped non-event. The known trap is a NULL slot_name; a *wrong* one
   fails identically and just as quietly. Also: a page still at
   `build_status='planned'` is skipped regardless.

### bugs_open/030 head-of-line blocking, second instance in two days

The re-render queued behind offset 105214 — **another session's council-gate
submission**, exactly as on 07-25 (that one was offset 104102). Consumer wedged,
lag climbing, while 24 orchestrations from other paths completed normally in the
same window.

Two instances, two days, same named cause: **a large council-gate run stalls
every unrelated message behind it on `system.agent.generic.requests`.** That is a
sharper statement than the workstream's current "~8min stalls", and it is not
cron-related. The practical consequence for anyone doing live-site work: a page
fix can be authored, correct, and committed, and still not be visible for half an
hour, with no failure anywhere to look at.

### Resolved — and a correction to my own reporting

The cases index went live. Full crawl of every live page, every internal link
followed:

```
internal links: /about.html  /cases/index.html  /contact.html  /index.html
  200  /about.html      200  /cases/index.html
  200  /contact.html    200  /index.html
```
**Zero broken links.** From six 404s (including the header's own nav item) to none.

> **CORRECTION to what I reported mid-flight.** I read a run of `404`s on
> `/cases/index.html` as the page not being deployed, and told the owner the
> header CTA was dead. Several of those were **`000`, not `404`** — transient
> curl failures, which I had been collapsing into "not live". The giveaway was a
> `000` on `/` in the same sweep, a page I already knew was serving. The object
> was in B2 (`b2 ls b2://portfolio-sites/oufe.com/cases/index.html`) before I
> said otherwise.
>
> Cheap check that would have caught it: `curl --retry 3 --retry-all-errors`, and
> **treat `000` as "no answer", never as "not found"** — they mean opposite
> things. A status code of 000 is the absence of a response, so it carries no
> information about the resource at all.

Also worth noting for the deploy model: `orchestration_states` sat at
`AWAITING_RESPONSES | deploy_page` while the page was **already live**. The
git-adapter commits and the GitHub Action syncs to B2 independently of the
orchestration closing its loop, so **the orchestration status is not the
liveness oracle** — the object in B2, or a retried curl, is. `pages.deployed_at`
was still NULL at that point too.

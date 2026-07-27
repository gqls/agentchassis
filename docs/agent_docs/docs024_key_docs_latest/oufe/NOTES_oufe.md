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

---

## 2026-07-26 (evening) — the generic high-attention lane, and the bugs it grew out of

**Four platform defects filed** rather than left in this workstream's notes,
because none of them is oufe-specific and all four will bite the next site:

| bug | what |
|---|---|
| `bugs_open/094` | `049b_deploy_single_page.sh`'s `section_data_resolved` branch cannot work — omits `page_name`, required at `rerender_page_sections_action.go:80` |
| `bugs_open/095` | a **wrong** `slot_name` renders nothing and reports `COMPLETED \| complete_skipped` |
| `bugs_open/096` | residual of the closed 030: cron got its own lane, the generic lane is still strictly serial, so a long council run blocks everything |
| `bugs_open/097` | CTA integrity misses `content_data.cards[*].link_url` — six broken links incl. a nav item shipped while two were correctly caught |

On 096: 030 is genuinely fixed and its close is sound — cron latency went from
~18 min to ~1 s. What I saw is the part that fix did not cover, and filing it as
a residual rather than reopening 030 keeps that distinction honest.

On 097 the sharpest detail is that **two of the six broken links had targets that
existed** — `/cases/thames-water` and `/cases`, whose real urls are
`/blog/thames-water.html` and `/cases/index.html`. The cards were written from the
plan's intent rather than from the page record. So "does a page with this name
exist" is not a sufficient check; it has to resolve against `pages.url`.

### The grounded-explainer lane (migration 224)

The owner asked for the mechanism explainers to be research-grounded and truthful,
and for that to be **generic across domains** via a workflow that deliberately
calls for increased attention.

The design decision worth recording: **"be careful" in a prompt is not a control.**
Every step removes a way to be careless instead —

- facts arrive by search, never by recall;
- each candidate must carry a verbatim quote from one named source;
- `verify_and_register_citations` re-fetches every url and discards any claim
  whose quote is not literally present (**the model proposes, the fetcher
  disposes**);
- the composer sees only the survivors, and is told explicitly what it may not
  assert — legal conditions, thresholds, definitions, facts about real named
  organisations — while being left free to explain mechanism and to use openly
  hypothetical figures, which cannot be wrong about anybody;
- an **independent** audit re-reads the draft against the same verified set and
  lists every sentence it cannot trace, and separately flags any claim that the
  page is accurate/verified/authoritative;
- the run terminates at `needs_human_review`. **There is no config flag that
  makes it publish.**

That last one is load-bearing. An automated content lane that *can* publish will
eventually publish something wrong unattended; one that cannot, cannot. It is the
only property in the list that does not depend on a model behaving well.

Steps 2–6 are the proven V5 acquisition chain copied wholesale rather than
reinvented, so facts it registers stay V2-usable by writers and V4-refreshed
afterwards.

### Missteps this session

- **I reported the header CTA as dead off a run of curl `000`s**, having collapsed
  them into "404". `000` is the absence of a response and says nothing about the
  resource; the giveaway I walked past was a `000` on `/`, a page I already knew
  was serving. Corrected to the owner in the same turn. Check:
  `curl --retry 3 --retry-all-errors`, and never treat `000` as `404`.
- **`agent_definitions` has no `name` column** — it is `display_name`. The insert
  failed on it. I had written the column list from the shape of the
  evidence-researcher seed without reading the actual schema, which is precisely
  the "schema first: `\d <table>` before writing SQL" rule in CLAUDE.md.
- **The two content writers keep their prompts at different paths**, so my first
  fleet-rule UPDATE silently covered only one of them. `UPDATE 1` looked like
  success. Verify by type, not by rowcount.
- Earlier in the day: patching `content_data` without re-rendering (the assemble
  stitches stored `rendered_html`), and a `slot_name` of `'main'` that matched no
  section. Both now in the RUNBOOK.

The pattern across all four is the same and worth naming: **each was a case where
something returned a success-shaped result** — `UPDATE 1`, `COMPLETED`, a status
code that was not a status code — and I read the shape rather than the substance.

### First live run of the grounded lane — and what it caught about itself

Corr `8896cc75`. The acquisition half worked on the first attempt:
**14 citations machine-verified**, mostly from **legislation.gov.uk** itself —
s901F(1)'s 75%-in-value majority, s901G, both cram-down conditions, the statutory
definition of the *relevant alternative*, Part 26A's insertion by CIGA 2020, plus
the Adler Court of Appeal outcome and Re Nasmyth. Exactly the load-bearing
conditions a reader could get wrong, quoted from the instrument rather than from
commentary about it.

**That is also V5 completing end to end for the first time.** It was activated
2026-07-20 and had never finished a successful run; the workstream inherited it as
"live but never exercised". It is now exercised: 14 candidates proposed,
14 re-fetched and confirmed, 0 discarded.

Then two defects, and the second one is the more interesting.

**The composer never saw the facts.** `verify_and_register_citations` returns a
receipt — `{"registered": ["CIT-a7f91f88754d560", …]}` — and my prompt
interpolated that, so the writer received a list of opaque identifiers with no
content. **And it behaved correctly**: `sources_used: []`, and a `gaps` list naming
as unverifiable precisely the facts that had just been registered — "the statutory
majority required for a class to be treated as approving a plan", "the precise
legal definition of the no worse off test", "whether cross-class cramdown is
available at all".

> **This is the design working, and it is worth dwelling on.** A writer starved of
> facts produced honest gaps instead of confident law. The bug cost one run. The
> alternative failure mode — a writer that fills the gap from memory — costs a
> wrong statement of statute on a live page, and would have read completely
> plausibly. **Making the unsafe path impossible matters more than making the
> happy path smooth**, and here the unsafe path simply was not available.

**The audit destroyed the work it was auditing.** `audit_grounding` hit its
6000-token cap (`stop_reason=max_tokens … 0 chars recovered`) and, having no
`error_step`, failed the whole orchestration — taking a 6,397-character draft with
it, recoverable only by hand out of `collected_data`.

That is CLAUDE.md's `output_tokens == max_tokens means the completion was CUT`
rule meeting a new surface: not an artifact truncated into the database, but a
**verification step whose truncation took a good draft down with it**. The
durable principle, now in migration 225: **a check that fails must not destroy the
thing it was checking.** The audit now routes its own failure to the review item,
flagged unaudited so a human reads it *more* carefully — the same discipline as
tool-generator's doc steps, which carry `error_step: complete` so a docs failure
can never fail tool creation.

Migration 225 fixes all three: a `load_evidence` step reads the register back so
the composer and the auditor both see the claims themselves; the audit cap goes to
12000 and its prompt asks only for problems rather than an enumeration; and the
audit can no longer fail the run.

### Run 2 died to a chassis roll — a clean instance of bugs_open/003

Corr `2363dbb7` froze at `extract_claims` and never moved again. The timing
settles it without ambiguity:

```
orchestration last updated : 2026-07-26 21:02:53.988+00
new chassis pod started    : 2026-07-26 21:02:56Z   (5b4456686c, replacing 76745d8f45)
```
Another session rolled the image ~3 seconds after my run entered its first LLM
step. The in-flight awaited response died with the old pod — **bugs_open/003
spawn loss**, not a defect in the lane.

Worth recording as an operational fact rather than a grievance: on a shared
cluster, any multi-minute orchestration is exposed to any other session's deploy,
and there is no signal at the time. The symptom is a step whose `updated_at`
simply stops. **`now() - updated_at` against pod `startTime` is the two-line
diagnosis**, and it is much faster than reading logs:

```sql
SELECT current_step, now()-updated_at AS since_update
  FROM orchestration_states WHERE correlation_id='<corr>'::uuid;
```
```bash
kubectl -n ai-persona-system get pod -l app=agent-chassis \
  -o jsonpath='{.items[*].status.startTime}'
```
A `since_update` that exceeds the pod's age means the work predates the pod that
would have to finish it. Re-fire; nothing is recoverable.

Note this also means **the earlier successful run's 14 verified citations
survived** — they were written to `evidence_base` at the time, and a lost
orchestration does not roll them back. The register is the durable artefact, not
the run.

### Run 3 — the lane worked end to end, and the audit earned its place

Corr `cade11a4`. `COMPLETED | complete`. The full chain:

- **19 facts** now in the register (14 from run 1 + 5 more this run), all
  quote-verified against their live source;
- draft **7,546 chars**, citing **3 sources**, declaring **3 gaps** honestly;
- audit verdict **`needs_revision`** with **2 ungrounded sentences** and
  **0 reliability overclaims**;
- work item `grounded_draft_review` sitting at `needs_human_review`. The gate
  fired. Nothing published itself.

**The two catches are the point, so record them verbatim:**

> "Creditors are divided into classes, generally grouped by the similarity of
> their legal rights against the company."
> — *no verified fact states this class-grouping principle; it is an unsourced
> legal generalization presented as fact.*

> "commonly this will be some form of insolvency process, but the statute does not
> name one"
> — *an added generalization about typical relevant-alternative outcomes that is
> not supported by any verified fact — only the bare statutory definition is
> verified.*

Both are the kind of sentence that is probably true, reads as settled law, and
would never have been questioned by a reader. **That is exactly the class the lane
exists to catch**, and a `needs_revision` verdict with two items is the check
doing its job — not a failure. The auditor also correctly *declined* to flag the
closing disclaimer, calling it "appropriately modest and not a reliability
overclaim", so it is discriminating rather than merely strict.

Minor defect noted, not yet fixed: the work item's `summary` rendered as the
literal string `grounded_draft_review` rather than my `summary_template`, so that
is not the right config key for `create_work_item`. Cosmetic — the spec payload is
intact.

---

## 2026-07-27 — the owner refused the answer, and the answer was wrong

He read the O4 recommendation — build a live sweep, open a promise register — and
pushed back: *"we have existing functionality that double checks claims, and we
have the council. Please look hard at our existing documentation and solutions."*

He was right, and the research took ten minutes.

### The misstep, stated plainly

**I claimed the overclaim class was invisible to every scanner we own. It was
never invisible.** `ScanBannedClaims` (`datahelpers/claims.go:284-325`) is a bare
case-insensitive regex over prose blocks. No number extraction, no
`businessClaimContextRe`, no `isExcludedNumber` — those gate only
`ScanUnregisteredNumbers` (`claims.go:365,369`). It matches whatever patterns a
site is given, about anyone, numeric or not. Live registers already carried purely
qualitative patterns: `leaderboard`, `live now`, `price target`, `years of
experience`.

Arming oufe took **one UPDATE and no image roll** (mig 226) and bought a
build-time **blocker** plus a **high**-severity post-deploy finding. Ten patterns,
tested both ways: 10/10 fabrication shapes blocked including all four phrases the
site actually shipped, 13/13 legitimate sentences pass — including the honest
replacement copy and the approved disclaimer's own wording.

### How the error got in, which is the useful part

**A limitation about one component bled into its neighbour.** Earlier the same day
I established, correctly and with evidence, that `ScanUnregisteredNumbers` is inert
on finance prose. That is true and it is written up accurately. Then
*"the number scanner cannot see this"* quietly became *"the scanner cannot see
this"* — and I never opened the sibling function twelve lines away. Thirty seconds
of reading would have stopped an afternoon of building on it.

**I wrote a universal negative from local evidence.** "Nothing in the estate looks
for this" is a claim about the whole estate. My sentence named four mechanisms; I
had read the source of exactly one. The tell was in the sentence itself.

**The answer was already filed.** `SPEC_claims_verification.md:250-252` poses this
precise question — should `banned_claims` be fleet-shareable, *"some patterns are
universal"* — and defers it: *"per-site only until two sites have evidence bases"*.
Written when n=1. There are now 8. **The decision was due, not new**, and I never
grepped the spec for my own problem before declaring the platform had no answer to
it.

### Where it did damage

Not just a note. It went into **the standing instructions of a live reviewing
agent**: migration 223 told the compliance seat *"no scanner will catch this, so
this seat is the only control."* A reviewer told it is the only control will
substitute for a mechanism instead of asking for one. Corrected in mig 227 and
mirrored to both rosters; the false premise is also corrected inline in 223's
header, in the SUMMARY, and in O4, with the original reasoning kept visible.

**The content of the error is the failure the seat exists to catch** — a
confident, unverified claim about what our system guarantees, written by the person
building the overclaim detector. That is worth sitting with rather than tidying
away.

### What was actually missing

Reach, not capability. **5 of 15 live sites carry a single banned pattern.** The
ten without include **vetcomparison.uk** — the site of the fabricated-prices
incident — and **idea.uk**, which takes real money. There is no way to define a
pattern set once for the fleet, and `globalTellPhrases()`
(`voicetells.go:121-137`, unioned with the per-site list at `:109`) shows exactly
how the sibling engine already solves that.

And a second reach problem, separately owned: `check_unverified_claims` runs only
under `quality-discovery-agent` → `improvement-loop` → `improvement-sweep`, which
has been **disabled since 2026-05-02** (`bugs_open/083`). The post-deploy check
exists and effectively never runs.

### The promise question, corrected too

`evidence_base` already declares `kind: metric | capability | entity | attestation`
and `source: sql | artifact | attested_by`, and V4 already re-runs sql-sourced
facts daily across the fleet. **`Kind` is declared in the struct and read
nowhere.** A promise is a `capability` fact whose source is the mechanism that
keeps it — the slot was cut and never used. Meanwhile EXPERIENCE_PLAN §2 is
*literally called a promise ledger*, and the experience-register harvest already
recorded the sharpest line on it: **"A promise ledger the platform cannot
mechanically check is prose."**

A new register would have been the third thing in this estate to model the same
idea.

---

## 2026-07-27 — the owner read the site and it fails on four counts

He put it plainly: the text is AI-sounding, the case isn't written, there are no
tools, and the design is the same site the platform has produced before. All four
are right. Recorded here in his order.

### The voice

He asked me to find the earlier discussion about voice rather than invent a new
one. It's `travelling_docs/pitch_pdf_source/REVERSE_ENGINEERED_STYLE_PROMPT_v3.md`,
built 2026-07-17 by comparing AI copy against a hand-edited rewrite he judged
better, then refined across three rounds of him critiquing the prompt's own
output. Rule 3 is the one he named this morning: say what a thing is before you
say what it isn't.

Measured before touching anything. The homepage carried 3 em dashes and 7
negative-frame constructions. The about page carried 3, 19, and 6 sentences
opening "It is…" or "It does…".

**`page-content-writer` already had rule 3.** It didn't work. The rule was one
block inside a 12.5 KB prompt that also carries the schema, the section spec, the
research findings and four other rule sets, and a rule competing with that much
context loses. `content-writer` had no style guidance at all.

Two things worth separating. Some of the worst copy on the about page was **mine**,
hand-written during the 07-26 fallibility rewrite: "That does not make us right."
opens on a negative, and "A citation shows you our source. It does not prove that
we read it properly." is a contrastive pair saying one thing twice. I wrote the
guidance about plain prose into the mission brief and then broke it myself.

Fixed three ways, on his directive that the writer should default to this voice:
migration 228 puts the house voice into both writers as the default, deferring to
a site's own `voice` spec where one exists; the oufe copy is rewritten across six
blocks; and `check_voice_tells` is armed on oufe, which was enabled on **1 of 15
sites**. Writing a rule and checking it are different jobs, and this week keeps
teaching that.

### The tools

There were none, and that was my error rather than the framework's. I wrote
`PREPARED_tool_insert.sql` on 07-25, deliberately did not apply it because the
site plan didn't exist yet, and then never went back. The file even says "NOT YET
APPLIED" at the top. Applied today.

Two traps on the way in, both mine:
- `\set tmpl \`cat …\`` inside `kubectl exec` runs **in the pod**, where the repo
  isn't mounted. The template has to be inlined host-side.
- My inlining replaced **both** occurrences of the placeholder, including one
  inside a header comment, so the 20 KB template broke out of the comment block
  and psql tried to execute CSS. Replace the value line, not the token.

The page then refused to render because its section had NULL `content_data` — my
own guard in `TRIGGER_rerender_page.sh`, doing exactly what it was written to do.
A standalone tool has no template fields, so `{}` is right and NULL is not.

### The spec

Against the roadmap brief: index, about, cases-index and contact are deployed.
**`thames-water` is still `planned` with zero sections** — the case he says isn't
written, and it isn't. The tool page now exists and is rendering.

So the site matches the spec on page *inventory* and fails it on *content*: the
flagship case is an empty shell, and the cases index honestly says so, which is
the least bad version of being wrong.

### The design

Filed as `bugs_open/107`. The palette is not the problem — oufe has its own style
collection. The sameness is in the section composition:

```
ai-agent-orchestration  hero › system-stats › features › differentiators › case-studies-grid › departments-grid › latest-news › call-to-action
finetuning.uk           hero › features › differentiators › case-studies-grid › departments-grid › call-to-action
fundamentallyai.com     hero › stat-band › evidence-chart › differentiators › features › info-card-grid › portfolio-showcase › call-to-action
robot-hands.com         hero › features › brief-explanation › tool-list › latest-news › call-to-action
oufe.com                hero › brief-explanation › info-card-grid › call-to-action
```

Hero first, call-to-action last, interchangeable card furniture between. A gripper
manufacturer, a consultancy, a fine-tuning service and a restructuring publication
all land on the same page. Nothing in the planning loop represents *what kind of
publication this is* as a constraint on shape, so the shape defaults to the
commonest arrangement in the component library, which is a brochure.

oufe wants a reading order and got a conversion funnel. The "Get Started" button
and "Our Services" footer group I removed earlier in the week were the same cause
showing up in the chrome.

### The tool install reproduced a bug I had filed the day before

`PREPARED_tool_insert.sql` was written 2026-07-25 with `slot_name = 'main'` and no
`pages.sections` entry. I filed `bugs_open/095` on 07-26, describing exactly that
failure: a wrong slot name matches no section, renders nothing, and the run
reports `COMPLETED | complete_skipped`. On 07-27 I applied the file unchanged and
hit it again.

The file predates the bug report, so nothing was ignored. What was missing is the
step between: **filing a bug does not fix the artefacts that already contain it.**
Nothing swept the repo for other files carrying the same shape, and the prepared
file sat there for two days with the defect written into it in advance.

Both corrected in the file itself, so the next reader gets the fixed version
rather than the fixed *description*.

Worth noting alongside the humanised-copy work in the same session, because it is
the same shape twice in one day: the rule existed and the artefact didn't follow
it. Rule 3 was in the writer prompt and the copy still broke it; bug 095 was in
the bug list and the SQL still broke it. **Writing a rule down changes nothing
that already exists** — the rule needs either a sweep over what's already there,
or a check that runs against output rather than intent.

### Copy measured after the rewrite

| page | em dashes | negative frames | "It is/does" openers |
|---|---|---|---|
| `/` before | 3 | 7 | 0 |
| `/` after | 1 | 6 | 0 |
| `/about` before | 3 | 19 | 6 |
| `/about` after | **0** | **6** | **1** |

The remaining six on `/about` are largely legitimate and should stay: the
required "not investment advice, not a financial promotion" disclaimer, a
conditional ("where a figure is not yet verified"), and two trailing-clause
contrasts, which rule 3 explicitly permits once the fact has been stated first.
The counter is crude — it matches "is not" anywhere — so read the sentences
rather than the number. The paragraph-opening negative frames, which are what
rule 3 is actually about, are gone.

Three highlight blocks on `/about` are still generator copy; only the first was
rewritten. Owed.

### The tool is live, and the research lane met a real judgment

`https://oufe.com/tools/tool-recovery-waterfall.html` serves. Verified against the
live page rather than the row: the condition-of-use gate, the EV input, the
verdict target, the caveat inside the output block, the inline JS, the site
chrome, and no unresolved template placeholders.

Getting there needed the assemble-only route, because `save_page_sections`
refuses an owned page and every tool page is owned. Runbook §8b now carries it.

### Thames Water: the lane found the right document and choked on it

Corr `2c5bbf90` ended `complete_no_sources`. The acquisition half worked
**better than I expected** — it went to the primary instrument, not commentary:

```
judiciary.uk/…/Kington-S.A.R.L.-Thames-Water-…-judgment.pdf
  → "Neutral Citation Number: [2025] EWCA Civ …"
```

The scrape succeeded. `extract_claims` then failed and took the error branch.

Cause: size. The prompt interpolated every scraped source whole.

| run | scraped chars | outcome |
|---|---|---|
| mechanism explainer | 320,692 | 19 citations registered |
| Thames | **584,152** | nothing |

A Court of Appeal judgment is not a web page.

**And the fix already existed.** `format_research_content`
(`research_actions.go:204-320`) takes `max_content_per_source` and truncates each
source into one LLM-ready block. **I listed that action by name in my own research
notes while designing this lane** — "format_research_content — Format scraped
content for LLM context" — and then wired scrape straight into the extractor.

Third time this week: the capability was present, unused, and invisible because
nothing pointed at it. The other two were `ScanBannedClaims` and
`EvidenceFact.Kind`. In all three I reasoned from what I could see running rather
than from what was available.

Migration 229 inserts the step at a 24,000-char cap. **Honest limitation, stated
rather than papered over:** a long judgment is now searched across roughly its
first fifth. Judgments front-load the citation, parties, issues and often a
summary, so that is the richest part for quotable claims — but a claim at
paragraph 180 will not be found and the run will not say so. Chunking is the real
answer and needs a loop this workflow does not have.

Re-run fired as corr `a07edd25`.

### The sweep I kept saying was missing, finally written

After wiring the house voice into `page-content-writer` and `content-writer`
(mig 228), then `grounded-explainer` (mig 230), I counted:

```
with_voice | total
         3 |    26
```

Three of twenty-six prose producers. Each of the three was patched by hand
because it was the one in front of me at the time — and I wrote "the rule
existed and the artefact carried on regardless" into 230's own commit message
while doing exactly that for the third time.

`SWEEP_house_voice_coverage.py` finds every reader-facing prose producer, locates
its prompt wherever it lives (the paths genuinely differ: `generate_content`,
`generate_hero_content`, `generate_about_content`, `generate_draft`, some nested
under `process_sections_loop.sub_workflow`), and appends the block once.
Idempotent, so the count is the report and a re-run is a no-op. It deliberately
skips classifiers, planners, reviewers and extractors: a voice instruction on a
JSON-emitting prompt is noise at best.

Result: **7 of 15** genuine candidates now carry it, and re-running finds nothing.

**A finding the sweep surfaced that I would not have looked for.** Seven
`content-creator-*` agents (cta, features, testimonials, contact, and the bare
`content-creator`) have exactly one step, `execute_llm_prompt`, whose config
contains **only `input_fields` and no `prompt_template` at all**. So either the
action supplies a default prompt, or these agents cannot produce anything. Left
alone here — there is nothing to append to — but it belongs with the
dormant-agents work (`bugs_open/044`). Also note some types have **two active
rows**, which the sweep's `LIMIT 1` papers over and which nothing else seems to
police.

This is the fourth instance of the week's recurring shape, and the first time the
answer has been a sweep rather than another hand-patch:

| the rule | what already existed and ignored it |
|---|---|
| extraction froze 07-13 | 51 workstreams shipped after it, nothing noticed |
| "revisit at two sites" | reached eight, nothing re-read it |
| `bugs_open/095` filed | prepared SQL still carried the defect next day |
| house voice in the writer | live copy still broke it; 23 writers never had it |

**Writing a rule is the cheap half. The sweep is the half that changes anything.**

# NOTES — farmerinsurance.uk (append-only, newest at the bottom)

Site id `99cae989-2413-430d-b026-59dfeeb638c0`. Lane opened 2026-09-04.

## 2026-09-04 — session 1: the lane opens with a census against the artefact

### Ownership check before claiming anything
`scripts/who-owns.py farmerinsurance` and `… farmer` both return "no bug file matches" (it
resolves BUG numbers/slugs, not domains — noted so the next session does not read that as an
all-clear). Real check: `ListAgents` at 12:1x BST — 104 peer sessions, none named for farmer;
`loanzy.uk`, `lendzy`, `copy quality two stage` all live and each owns a different aspect (see
PLAN's table). Fleet-wide grep of `docs024_key_docs_latest/` for "farmerinsurance" hits 30
lanes and **no dedicated farmer directory** — the site had no lane of its own until now.

### The served site, measured (not inferred)
`[MEASURED 2026-09-04 ~12:20 BST]` All 18 active pages fetched with curl:

- 18/18 return **200**; `/this-path-does-not-exist-9f3a.html` returns **404** — the control
  that makes every other reading on this domain informative (parked-domain landmine cleared).
- 18/18 carry `<title>`, `<meta name=description>` non-empty, `<footer>`, and a `rel=canonical`
  whose href matches the page's own URL.
- 27 distinct internal link/`src` targets across those pages; **all 27 return 200**.
- Homepage: **0** carousels, **102** elements with a `card` class (was 72 on 08-31 — the
  owner's §5 card-stack complaint has got numerically worse, not better; the fix is the offer/
  component lanes', not this one's).
- The skip link (`Skip to content`) IS present — the improvement_loop lane's `assemblePage`
  fix has reached farmer's rendered pages.

### The queue reads far worse than the site is — three buckets, three different reasons
Open (non-terminal) rows: **~274** across 27 item types `[MEASURED 2026-09-04]`. The three
biggest are all stale or moot, and each fails differently:

1. **`unbuilt_internal_link` — 62 unresolved + 42 failed = 104 rows, and EVERY ONE names the
   same target: `/claims.html`.** ("href /claims.html points at a page that has never been
   deployed", filed 09-02/09-03.) `/claims.html` now has `pages.build_status='deployed'`,
   `deployed_at = 2026-09-04 01:18:56Z`, and serves **200 / 84,739 bytes**. So all 104 are
   FALSE as of this morning's deploy. The claim was true when written — `/claims.html` was the
   site's `bugs_open/437` victim (`needs_rebuild` since 08-27, recorded by the loanzy lane
   09-02) — which is exactly the "claim of UNKNOWN AGE" shape: `insertWorkItem` is
   dropOnConflict, so nothing refreshes a row whose premise expired.
2. **`canonical_mismatch` — 20 rows, all 20 on ARCHIVED pages** (the tool cull's `/tools/…`,
   `/guides/tool-…`, guide pages). MOOT, not false: the pages they judge are retired and 404.
3. **`head_essentials_missing` — 34 rows: 20 on archived pages (MOOT) + 14 on ACTIVE pages
   that today serve title, description, canonical AND footer (FALSE).** This confirms, at a
   second site, the improvement_loop lane's standing warning that farmer's "no title/footer"
   rows are false at the served page — and gives it a clean split (moot vs false) it did not
   have.

**Why this matters beyond farmer:** 158 of ~274 open rows (58%) are stale or moot, and the
site's real state is BETTER than its queue by a wide margin. Any reader who sizes farmer's
health from `site_work_items` gets the wrong answer. Disconfirmable: had a page been missing a
title, or had `/claims.html` 404'd, the same commands would have said so.

### The two defects a visitor CAN see, both on the homepage
1. **A UK private-medical-insurance directory on a farm insurance hub.** `/health-insurers.html`
   — title "UK Health Insurer Directory | Farmer Insurance UK", h1 "A plain list of UK health
   insurers." — lists Bupa, AXA Health, VitalityHealth, WPA, Saga, The Exeter, Freedom,
   National Friendly, General & Medical, Aviva, Drewberry. The homepage links it under
   "A directory of UK health insurers". The site's own identity spec says industry
   "Insurance Information & Education", services = farm buildings / livestock / machinery /
   crop / liability. Root cause found this session — see below. **Cosmetic second defect on the
   same page:** the directory headings render the provider name twice ("Drewberry Drewberry",
   "WPA WPA") or name+underwriter unseparated ("Saga Health Insurance Bupa",
   "VitalityHealth Discovery Holdings").
2. **The news section is still American**, four days after the owner flagged it (08-31 finding
   §3). Live headlines on the homepage today: "Aon to acquire USI to establish the premier U.S.
   middle-market platform", "Aon buys mid-market insurance specialist in $17B deal",
   "Governor Abbott Directs the Texas Department of Insurance…", plus CNBC, insurancejournal,
   programbusiness, chicagobusiness, jdsupra, prnewswire. Two readings worth keeping apart:
   (a) the REGION defect is unfixed — the news lane's CONTRIB measured the capability as
   absent fleet-wide (zero region keys in 48 `news_search` configs), and nothing has shipped
   since; (b) the SECOND defect the council seats filed against the same section — "malformed
   Google redirect URLs instead of real content" — **appears FIXED**: every news link on the
   homepage today points at a real publisher host, none at a google redirect. Recorded as a
   measurement, not as a claim about who fixed it.
   Also worth naming: the stories are not merely US, they are corporate M&A trade press —
   irrelevant to a farmer whether or not they are British. The site's `vertical_keywords` are
   "insurance market / insurance regulation / claims / premiums": generic, regionless, and
   audience-less.

### Root cause of the health-insurer directory — one map entry, refuted by its own file
`platform/orchestration/actions/feed_directory_recommendation_action.go`:

```go
"insurance": {
    Recommended:  true,
    Reason:       "Insurance sites gain authority from a cited, verified directory of UK health
                   insurers (the one insurer kind built so far; more kinds follow)",
    Kind:         "health-insurer",
    SpecKey:      "health_insurer_directory",
    SeparatePage: true,
},
```

Two entries below it, the same map states the rule that forbids exactly this:

> `"finance"` alone is deliberately NOT recommended: it is too generic to pick a single
> provider class, and **a wrong directory on a site is worse than none.**

`matchVerticalDirectory` lowercases `industry`, `site_type`, `category` and appends one
domain-derived signal, then takes the LONGEST map key contained in each. Farmer's industry is
"Insurance Information & Education" → contains "insurance"; the domain
"farmerinsurance.uk" → also contains "insurance". Either path alone reaches the health-insurer
kind. The reason string in the shipped spec row admits the substitution in writing ("the one
insurer kind built so far"), and that string is what a reviewer of the site's spec would read
as a justification.

Fleet census `[MEASURED 2026-09-04]`: exactly **1** site carries
`content_features.health_insurer_directory` today — farmer. So the damage is site-local; the
MECHANISM is not, and it fires on any insurance site that is not health insurance (pet, car,
travel, farm). Filed as `bugs_open/481` (renumbered off a same-day collision on 479) and put through the 090 diagnosis loop rather than
asserted: intake corr `bc8d399f-da3b-4408-ad0e-985bf2f7cd7a`, RUN corr
`c705263c-9b07-40fb-800f-6ebe7e1ce4a8` (the run correlation is the key the artifacts carry).

### Parked for the owner, and ageing
14 `copy_edit_proposed` rows sit at `needs_human_review`, filed 2026-09-02 17:37–17:48Z by the
copy lane (13 PASS + 1 annotated FAIL on `/blog/farm-buildings-insurance` — the proposal empties
a dead-tool secondary CTA). They are waiting on the owner's own ruling of 09-02 ("I'll review
them as a batch, please present them on the admin page and let me know when to look"). Two days
old at this session's start; nobody has told him they are ready.

### Site state facts worth not re-deriving
- `sites.status='deployed'`, `last_deployed_at` 2026-09-04 01:10:33Z.
- `settings->maintenance_profile->>'growth_posture'` is **NULL** — farmer is neither held nor
  released; it is one of the UNSET sites the improvement_loop lane counts, because migration
  722's born-held trigger only fires on INSERT and farmer predates it.
- `/contact.html` is the one active page at `build_status='needs_rebuild'` (all 17 others
  `deployed`). It still serves 200 — so this is a build-state row, not a live outage.

### ⚠ THE FLEET'S ANTHROPIC CREDIT RAN OUT AT 11:21Z TODAY — my 090 run was one of the first casualties
The diagnosis run fired for the directory defect (run corr `c705263c-9b07-40fb-800f-6ebe7e1ce4a8`)
**FAILED**, and not on its merits. Step `verdict` → `call_diagnoser` → `call_handler` all carry:

> `AI endpoint unavailable: provider=anthropic model=claude-sonnet-5 … status 400:`
> `{"type":"invalid_request_error","message":"Your credit balance is too low to access the`
> `Anthropic API. Please go to Plans & Billing to upgrade or purchase credits."}`

`[MEASURED 2026-09-04 11:25Z]` in `llm_call_log`, this is fleet-wide and total, not my run's bad luck:

| window | successes | failures |
|---|---|---|
| 11:09:26 → **11:20:50** | 32 | 0 |
| **11:21:12** → 11:24:43 | **0** | **18** |

Failing agents in that window: `council-gate` (14), `diagnose-agent`, `landmine-verifier`.
Every LLM call in the estate goes to Anthropic — `llm_call_log` for the last 45 minutes shows
provider=anthropic on 100% of calls (claude-sonnet-5 38, claude-opus-4-6 9, claude-sonnet-4-6 4),
so there is no second provider to fall back to. First sighting of the message anywhere in
`orchestration_states`: **11:21:12Z**, i.e. it began today, minutes before I looked.

Consequence for this lane: the 481 diagnosis must be **re-fired after the account is topped up**,
and the failure must NOT be read as "the loop looked and found nothing". Consequence for the
estate: every council verdict, diagnosis, landmine verification, writer and planner run started
after 11:21Z fails the same way, and the failures look like ordinary step errors.

⚠ When topping up, the known trap (MEMORY): **capped while billing reads 0% used ⇒ WRONG ACCOUNT**
— the fleet key is not on the default console org; check the keys' `Last used` column to find the
org that is actually serving these calls.

### 15:50Z — the credit outage CLEARED, the fleet is rolling, and my bug number moved
- **Credit restored.** `[MEASURED 2026-09-04 15:50Z]` `llm_call_log` for the last 90 minutes:
  **129 calls, 129 successes, zero failures**, 14:20:27Z → 15:45:09Z. So the window was roughly
  11:21 → 14:20Z. The 090 for the directory defect can be re-fired — but AFTER the roll settles
  (a chassis restart silently drops a dispatch for ~300s).
- **v1.0.1361 is rolling** (peer "inter thread comms", cut `06c0b18f2`, 14 images built
  15:29–15:36Z). Nothing of this lane's is in it and nothing needed holding: farmer's lane is
  docs-only so far, and the 481 fix is not written.
- **NUMBER COLLISION, and I moved rather than stood on being first.** This lane's bug was
  created at **12:20:07Z as 479**; another lane created its own 479 (Layer-2 tool orphans) at
  **12:23:57Z**. Being first settled nothing useful: by mid-afternoon every inbound "479" in the
  estate — `bugs_open/385`, `LANDMINES.md`, portfolio_positioning, bugfix_450, and Go commits in
  this roll — meant THEIRS, and mine had no inbound references outside this lane. So this file is
  now **`bugs_open/481`**, renamed with `git mv` (forward-only), with the collision recorded in
  its own header. **The check that would have caught it:** claiming max+1 by *looking* is not
  claiming it — the number is only yours once the file is committed, and two lanes looked inside
  the same four minutes. Cheap discipline: after committing a new bug file, re-run the max query
  and grep for your own number; if a second file appeared, move the one with no inbound pointers.

### 16:0xZ — ⚠ CORRECTION to my own headline number, and a sharper finding underneath it

> **CORRECTED 2026-09-04 (same session, ~3½ hours later): "158 of ~274 open rows (58%)" is WRONG.
> The true figure is 117 of 269 — 43.5%.** What I did: the denominator EXCLUDED terminal
> statuses (my open-rows query filters out `complete/cancelled/rejected/failed`), and then the
> numerator INCLUDED the 42 `unbuilt_internal_link` rows sitting at `failed`. A terminal row is
> not part of an open backlog, so those 42 could never have been in the 274 to begin with. The
> claim is repeated in `README_where_we_are.md`, in commit `d0a70596f`'s message and in the
> workstream memory line; corrected in the first two, and the commit stands as written because
> forward-only forbids an amend.
>
> **What caught it:** reading `workItemRevalidatableStatuses` to answer a different question —
> whether these rows can ever close — and noticing `failed` is terminal, therefore outside the
> population I had just quoted. Not a re-check of the arithmetic; a re-check of the DEFINITION.
> The cheap check I skipped: **run the numerator and the denominator in ONE query with the same
> status filter.** Two queries with two filters is how a percentage acquires a mixed population,
> and it reads exactly as plausible either way.

`[MEASURED 2026-09-04 15:5xZ, one query, one status filter]` Open non-terminal rows: **269**.
Stale or moot within them: **117 (43.5%)** —

| bucket | rows | why it is not a live defect |
|---|---|---|
| `unbuilt_internal_link` unresolved | 62 | all name `/claims.html`, deployed 01:18Z today, serves 200 |
| `dead_internal_link_live` detected | 1 | same target, same reason (from `/about.html`, 404 on 09-02) |
| `head_essentials_missing` detected | 34 | 20 judge archived pages (moot); 14 judge active pages that serve title+desc+canonical+footer (false) |
| `canonical_mismatch` detected | 20 | all 20 judge archived pages (moot) |

**And the sharper half — 63 of those 117 can drain themselves; 54 cannot.**
`workItemRevalidatableStatuses` (`platform/orchestration/actions/work_items_common.go:143`) is
exactly `{needs_human_review, unresolved}`. So:

- the **62 `unresolved`** rows are in scope for the review-queue sweep, and
  `unbuilt_internal_link` has a registered drain (`revalidate_unbuilt_link.go`) that delegates to
  the same `VerifyUnbuiltInternalLinkResolved` the completion gate uses — whose second disjunct
  is "the target has shipped". `/claims.html` now carries a `deployed_at` stamp, so these should
  close as `resolution_path='auto:revalidated'`, status **`complete`**. ⚠ **They land at
  `complete`, NOT `retracted`** — a query looking for `retracted` reads zero and concludes the
  drain is dead (the improvement_loop lane's landmine; it applies here verbatim).
- the **54 `detected`** rows (34 head + 20 canonical) are **NOT revalidatable** — `detected` is
  absent from that list. Nothing drains them, and 40 of the 54 judge pages that no longer exist.
  They are permanent furniture in this site's queue unless their producing check re-runs and
  retracts, which for an archived page it will not.

Check to run in a few days rather than assume: the 62 should have become `complete` /
`auto:revalidated`; if they have not, the drain is not reaching this site and that is a finding.

### 16:1xZ — ⚠ SECOND CORRECTION: the outage was 36 minutes, not 3 hours, and the roll has landed

> **CORRECTED 2026-09-04: "the window was roughly 11:21 → 14:20Z" is WRONG. The credit outage
> ran 11:21:12 → 11:56:48Z — about 36 minutes.** I checked recovery with a
> `now() - interval '90 minutes'` query, got 129 calls / 129 successes / earliest **14:20:27Z**,
> and wrote that earliest timestamp down as the moment the outage ended. It was the left edge of
> my own sampling window. The "inter thread comms" session contradicted it; I verified rather
> than adopting, and they are right — `[MEASURED 2026-09-04 16:1xZ]` credit failures number
> **117 rows, 11:21:12 → 11:56:48Z**, and the gap I had never queried (11:56:48 → 14:20:27)
> holds **416 successes and exactly one failure**, that one a `stop_reason=max_tokens`
> truncation at 13:31:31, not credit. All three non-credit failures today are truncations
> (08:42 tool-generator, 10:32 build-site-planner, 13:31 tool-generator).
>
> **The check, which is the general form of it:** an interval query can tell you a thing IS
> over; it can NEVER tell you when it ended. For the end, query the gap between the last failure
> and your window's start, and require the successes in that gap to be non-zero.
>
> Cost: the owner and at least one peer lane were told a 3-hour outage. Anyone writing off a run
> that failed "during the outage" at, say, 12:30 would have been writing off a run that actually
> succeeded — the opposite mistake to the one I was warning them about.

**Roll landed.** All 20 backend deployments on v1.0.1361, chassis pods ready 16:01:26Z /
16:01:53Z stamped `06c0b18f2`, `rollout status` clean at 16:02:15Z, settle window closed ~16:07Z
(peer measurement, and their caveat is worth keeping: at 16:00:08 one new pod was not ready while
two old ones still served, so a probe then reads the OLD binary and its clean pass is about the
wrong thing — wait for `rollout status`, and match `service_binary_capabilities` rows against
`kubectl get pods`, since four rows can exist for two live pods).

**The 090 for 481 has been re-fired** post-settle — RUN corr `cdcb2981-36ce-4d02-8d37-6ab302aede12` (the first attempt, run corr `c705263c`, is the outage casualty; intake for this one is a fresh row). Also recorded from the same peer, because it
bears on this lane's future council submissions: **a killed correlation is REUSABLE** — one
correlation carried four council runs across the outage (revise 11:15, complete_invalid 11:29,
revise 12:08, approved 12:23). Use `RESUBMIT_CORR=<corr>`; minting a fresh one splits the trail
and leaves a `Council-Submitted:` trailer pointing at a correlation that never produces a verdict.

### 16:3xZ — the worst thing on the site is invisible to every link checker: 41 buttons advertising deleted tools

`[MEASURED 2026-09-04, served HTML of all 18 active pages]` **41 anchors across 15 of the 18
pages carry a LABEL inviting the reader to use one of the seven tools the owner had deleted on
08-31, and every one of them points somewhere else that returns 200.** Examples, verbatim:

| page | label | href |
|---|---|---|
| `/contact.html` | "Check what cover your farm might need with the Farm Insurance Needs Checker" | `/blog/farm-machinery-insurance.html` |
| `/crop-insurance` | "Try the Farm Insurance Needs Checker" | `/guides/index.html` |
| `/legal.html` | "Try the Livestock Value Estimator" | `/blog/livestock-insurance.html` |
| `/glossary.html` | "Try the Farm Insurance Needs Checker to see which areas of your own cover…" | `/blog/farm-machinery-insurance.html` |

Two named tools account for nearly all of it — Farm Insurance Needs Checker and Livestock Value
Estimator. Plus **41 prose mentions across 9 pages** (outside anchors) naming the culled tools,
including 5 of "Farm Building Rebuild"/"Rebuild Cost Estimator" on `/blog/farm-buildings-insurance`
and 4 of "Farm Insurance Needs Checker" on `/guides/index.html`.

**Why this is the sharpest thing found today and why nothing caught it.** The CTA recompute
during the cull did its job — it re-pointed every href at a page that exists — so *every link
returns 200*. My own crawl this morning proved 27/27 internal targets healthy and was blind to
this by construction: a link checker asks "does this go somewhere?", and the defect is "it goes
somewhere that is not what the button says". A reader clicking "Try the Farm Insurance Needs
Checker" lands on a blog post about machinery insurance. On 15 of 18 pages.

**Ownership:** this is `copy_quality_two_stage`'s parcel — they named the class first
("misdirected-CTA labels, 52 fields; live-URL-dead-label class first") and hold the ruling.
Their 52 is a count of stored spec/CTA FIELDS; my 41 + 41 is a count of what the SERVED pages
show today. Different populations, both true, and the served count is the one that says what a
visitor meets. Handed over as a CONTRIB rather than touched here.

### 16:3xZ — the contact page is a dead end, three ways, and one of them is fleet-wide
`[MEASURED 2026-09-04 at the served page]` `/contact.html`:

1. **The form silently swallows messages.** `<form class="contact-form" id="cf-contact-form"
   action="#contact" method="POST">` — it POSTs to a fragment of itself on a static host.
   Nothing receives it. The work item `contact_form_undeliverable` said exactly this on
   **2026-08-28** and has sat at `needs_human_review` for a week.
2. **There is no other way to reach anyone.** No `mailto:`, no `tel:` anywhere in the page, and
   the identity spec's contact block is `{email: null, phone: null, address: null}`.
3. **The page contradicts itself.** Its own opening line is "Farmer Insurance UK does not arrange
   insurance or *take your contact details*, so this page will not get you a quote" — and then it
   presents "Send us a message · Name · Email · Message · Send Message". Given (1), the honest
   version of this page may be one with no form at all; that is a content decision, not a bug fix.

**Fleet-wide, the form defect is 7 sites** `[MEASURED 2026-09-04]`: ai-agent-orchestration.com,
**boxingonline.com** (the first paid build), cv1.co.uk, farmerinsurance.uk, garden-tools.uk,
relojistas.com, vonc.com. That is the `static_site_form_endpoint` lane's mechanism — routed to
them, not fixed here.

### 16:5xZ — the copy lane generalises the CTA finding, and their framing is better than mine
Reply from `copy_quality_two_stage`, recorded here because it changes what the finding IS:

> "A page title promises something the page body does not deliver. An anchor label promises
> something the destination does not deliver. **Same failure, same reason no existing check sees
> it:** every instrument we have asks whether the target exists, and the defect is that the target
> is not what the promise said."

So my label-vs-destination class and their **title-promise** class (which they already own) are
one defect at two scopes. Their consequence, which is the useful part: **if a detector is built
for either, it must be specified for both, because the expensive half — deciding what counts as a
broken promise — is shared.** My "27 of 27 internal targets healthy" and their "pages that pass
every structural gate" are the same blindness measured from two directions.

They also accepted the 52-vs-41 framing (stored CTA fields vs clickable anchors) as two
populations rather than a correction — their words: someone would have quoted one at the other
within a day. Worth keeping as a habit, not just this once.

Their two returns, neither an ask: (1) if the 14 proposals come back approved, `/contact.html` is
the one to HOLD — the copy fix alone makes that page more coherent and no more honest while the
form still swallows messages; (2) boxingonline (the shared victim of the form defect) has an owner
rejection open on its article copy, so anyone touching that site should expect company.

Also recorded from reading their lane rather than from their message: the
`static_site_form_endpoint` lane already knows farmer and boxingonline are victims, is mid-build
(migration 756, council corr `3aff429e`), and has found that `check_contact_form_undeliverable`'s
predicate is an ENUMERATION of known-bad form actions — so its coverage is a list, not a rule.
Farmer's `action="#contact"` is evidently on that list (the item exists); a form posting to some
other useless target might not be. Not this lane's to fix; no message sent, because sending one
would have told them what they already have written down.

### 17:0xZ — the feed lane accepts the confounding argument, refutes (A) by reading, and (B) stands
Their reply (their file: `news_feed_ingestion/RESPONSE_2026-09-04_to_the_farmer_lane_candidate_A_refuted_by_reading_and_my_proof_site_was_confounded.md`, commit `44e7177e7`):

- **(A) refuted by reading all nine hops** — dispatcher selects the whole config blob and publishes
  it as `source_config`; `call_agent` maps it to feed-ingester; `findSourceConfig` reads
  `input_data.source_config`; `FetchNewsSearchAction` copies `region` into `StepConfig.Config`;
  `WebSearchAction` puts it in the adapter request; the adapter assigns `SearchOptions.Region`;
  `firecrawl.go` sets `payload["country"]`. No hop filters keys. **Stated with its limit, which is
  the right way to state it:** reading proves the path EXISTS, not that the value arrived on
  farmer's 15:14Z fetch — and both log windows that would have shown it are gone (adapter pod
  restarted ~15:45Z, chassis rolled 16:01Z). The A/B remains the only positive proof.
- **(B) stands, and farmer's own result is the evidence:** `cbo.gov` returning beside `bbc` and
  `ft` is what a parameter that BIASES looks like; one that CONSTRAINS cannot return a US federal
  site.
- **The remedy is in their code, not the flag.** Farmer's four queries come from
  `verticalNewsMap`'s insurance entry, and *every vertical in that table is country-neutral by
  construction*. So a .uk site gets a country-neutral query plus a geo hint, and the query is the
  stronger signal. 691 wires the hint correctly end to end and that is all a hint can do.
- Their own harder admission, in their WRONG_CALLS: their plan's step 4 was the controlled test,
  it sat blocked on 691, another session applied 691, and step 4 was never run — so the fix
  reached production never verified end to end at the provider. **"Pending 691" aged into a false
  all-clear the moment someone else cleared the blocker.** Worth carrying: this lane will
  accumulate its own "blocked on X" lines.
- **The discriminator (my A/B) is theirs to fire and they are taking it to the owner** rather than
  spending credits and dispatching live on their own authority. Recorded in their RUNBOOK,
  credited here. This lane does not fire it.

They also supplied the uniqueness constraint that reorders `bugs_open/483` — appended there, §7.

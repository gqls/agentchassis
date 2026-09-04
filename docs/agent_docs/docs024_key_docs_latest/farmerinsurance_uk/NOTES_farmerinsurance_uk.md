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

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

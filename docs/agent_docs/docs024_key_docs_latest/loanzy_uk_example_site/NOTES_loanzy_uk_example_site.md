# NOTES — loanzy.uk example site

Append-only, newest at the bottom. Missteps included — they are the point.

## 2026-08-18 — session start: finding the previous thread, and what it turned out to be

Asked to find the previous loanzy.uk thread and take it on. There were **three**, none of
them a lane of its own:

1. `idea_uk_vm_site`, 2026-08-03 (owner: *"can you do loanzy.uk"*) — delegated to Cloudflare
   via Nominet EPP under the DESIGNCONSULT tag, zone active in 60s, edge cert issued. Left an
   open item: *"loanzy.uk content: needs a Worker route / B2 wiring — webdesign lane's
   machinery."*
2. `domains_cloudflare_rollout`, 2026-08-09 — the **dangling delegation**: NS pointed at
   Cloudflare, zone deleted, so requests timed out instead of failing honestly. Owner
   re-added the zone; clean 404 since. That session deliberately built nothing (HOLD domain).
3. `portfolio_positioning`, 2026-08-15 — **P9**: *"leave loanzy.uk with the webdesign team"*.

**A correction made on the way in.** `portfolio_positioning/HANDOFF_2026-08-18` §5 listed
loanzy.uk, the B8/B9/I10 holds and the build order as Phase D decisions *"unchanged and still
outstanding"*. All three had been **ruled by the owner at P9 on 08-15** — the bullet was a
stale carry-forward from `PLAN_2026-08-12`. Struck through with a dated correction
(`8229b1362`), not deleted. Cost of not catching it: that lane keeps asking the owner a
question he closed three days ago.

**The owner's memory of "a site about FCA rules built alongside loancash" was right about the
site and wrong about the domain** — it is `lendzy.co.uk`, one letter from `loanzy.uk`. Both
`loancash.co.uk` (L10, *"The Rules That Protect UK Borrowers"*) and `lendzy.co.uk`
(*"Know the Rules Before You Borrow"*) are live, 22 pages each, both redeployed 2026-08-18.
`loanzy.uk` has **0 rows in `sites`, 0 work items** and serves a 9-byte 404. Checked twice,
in two ways (DB census across `loan|lend|borrow|credit`, plus the live probe), because the
first check was a name match that could only ever have confirmed what I already believed —
the family census is the one that could have come out otherwise, and it did.

**What the decision became.** Four messages, converging: back in the finance queue → no, an
example site → which means no prior registry entry → built only from the webdesign.uk prompt.
Recorded as **P10** in the register with its consequences enumerated (`f21530d37`), because a
ruling whose *implications* live only in a chat log is one a future session will misapply.

**Why it lands where the webdesign lane is actually stuck** `[MEASURED, from their docs]`:
proposal F was approved the same day and its `writer_block` forbids naming examples until a
gallery exists, because none of the four attested example sites was built by the one-shot
route. This lane produces the first pair that would populate one.

**What the "one-shot route" is, mechanically** `[MEASURED]`: the chat box
(`box/chat-service`) is a self-contained intake bot — stdlib only, no DB driver, transcripts
to JSONL on the box. It **dispatches nothing**. So today the route is: customer's brief →
operator seeds and dispatches the standard gated pipeline. The honest constraint is therefore
about the **input**, not the button: every seeded input must trace to a sentence in the
prompt. Written into PLAN as the rule, with a requirement to log deviations here.

**Not done, deliberately**: no seed, no dispatch, no lane SUMMARY. The prompt is the owner's
and it is a published artefact — building first and asking after would produce a pair whose
prompt was reverse-engineered from the site, which is the one thing this exercise cannot
survive.

## 2026-08-18 12:53Z — DISPATCHED with NOTHING but the domain

**Owner's steer, which changed the plan before Phase 1 ran:** *"I'd like the framework to
determine the prompt — it should already have that step in the research stage or thereabouts.
So assuming no prompt at all. Let's try that."* So the candidate prompts I had drafted were
never put; PLAN Phase 1 ("the prompt, owner input, BLOCKING") is **struck** — the input is the
domain string and nothing else.

**He is right that the step exists** `[MEASURED]`. `082_submit_domain_unified.sh`'s own header
documents the graph: `needs_domain_research → domain-research-classifier → needs_strategy →
domain-strategist → needs_briefing → build-briefing-agent → needs_site_plan → …`, and *"a
fresh domain still gets a research step regardless: the classifier always runs in full (web
research + synthesis)"*. `--mission` is optional in the arg parser, and `domain-submitter`'s
workflow marks `persist_mission` *"skipped if absent"* — so a mission-less fresh submission is
a supported path, not a hack.

**What was passed: `loanzy.uk`. Nothing else.** No `--mission`, no `--email`, no `--phone`, no
fidelity, **and no seed SQL at all** — the FRESH path's `ensure_site_record` creates the site
row itself. That means no `evidence_base`, no `banned_claims`, no `imagery_style_guide`: the
site starts with nothing attested. Where webdesign.uk's own build could state its price
because the owner had attested it, this one has no attested anything, and what the framework
does about that is part of what we are here to observe. **Withholding the email was a
choice**: a customer supplies contact details, so if the build raises an item asking for them,
that is a finding to record and answer, not a failure to pre-empt.

Preconditions checked before firing, none inferred: chassis pods up since 07:57Z (~5h, clear
of the ~300s silent-drop window); `domain-submitter`, `domain-research-classifier`,
`domain-strategist`, `build-briefing-agent` all `is_active` and not snapshots; zero existing
work items on the domain.

```
CORRELATION_ID=a892b446-36bb-4b30-83be-71cca81ff53e
ORCHESTRATION_ID=84865f5f-5ef1-4cc1-8fb1-656951f1a5a6
site_id=55213ded-03ec-40f7-8fc1-169de05e05c8
```

`[MEASURED]` **Verified landed, not assumed from exit 0** — `kcat -P` is known to publish
nothing and still exit 0. `orchestration_states` COMPLETED at step `complete`, `sites` row
created (`build_status=pending`), and `needs_domain_research` filed `triaged` at 12:53:00Z for
`domain-research-classifier`.

⚠ `bash <script>`, not `./<script>` — the file is mode 644 in the tree, so `./` gives
"Permission denied". Recorded in the RUNBOOK.

**The risk this run carries, stated before the result is known so it cannot be rationalised
after:** the only signal the classifier has is the string "loanzy.uk", which reads as a
lender. If it invents a lender or broker, the result is a fake regulated firm on a live UK
domain, and the honest response is to retract rather than to publish it as a demo. Reviewing
the strategy spec before pages deploy is NOT steering the build — no input changes — but if I
ever intervene in the content, it gets logged here as a deviation and the pair stops being
"built only from the domain".

## 2026-08-18 13:36–14:07Z — the run's answer, the stop, and the page that got out anyway

**What the framework decided, with no prompt** `[MEASURED, site_specs]`. `identity`: company
"Loanzy", industry Financial Services, services = *Personal Loan Matching*, *Loan Comparison
Tool*, *Eligibility Checker* ("soft-search … does not affect the user's credit score"),
*Lender Lead Facilitation* ("affiliate/lead generation bridge"). `strategy`: `money_flow` =
per-qualified-lead / CPA fees from lenders, `primary_model` = `lead_generation`, growth path
including "FCA appointed representative relationships", `trust_threshold` = "FCA regulatory
legitimacy … above the fold". Site plan: 20 pages incl. `tool-eligibility-checker`,
`tool-compare-loans`, `lenders-index`, `lender-profile`.

**Where the direction came from, which is worse than a hallucination**: the classifier's own
`about_summary` says *"Evidence from related entities (loanzy.asia and loanzytech.com)
positions the Loanzy brand as a lead aggregation and facilitation platform"*. With no prompt,
the strongest signal available was **a third party's business**, so it took it.

**The system NOTICED and could not act.** `build-briefing-agent`'s `gaps` list, written before
any page: *"FCA authorisation number — not yet known; must be obtained before launch and added
to footer"*, *"Lender panel — specific lenders … not confirmed"*, *"Legal entity name … not
confirmed"*. A gap is a note; nothing gates on it. This is the single most useful thing the
run produced and it is why the guard belongs at the classifier (where the direction is CHOSEN)
rather than at the briefing agent (where it is already NOTICED).

**The stop, and its three errors** — full account in `WRONG_CALLS.md` and the SUMMARY:
1. I dispatched on a live, publicly-resolving domain having already written the risk down.
   Writing a risk down is not containment.
2. I first reached for `needs_rerender` as the publication gate. It is not:
   `page-build-handler` has its own `deploy_page` step — every page ships itself. Caught by
   reading the step list before acting, which is the only reason this was one page not twenty.
3. `status='cancelled'` stops an item being CLAIMED, not an agent already running. 33 items
   cancelled **13:57:24Z**; the already-claimed `about` page deployed **14:01:55Z**, four
   minutes later, to the public edge.

**Retraction is a two-step and the second step does not reach the artefact** `[MEASURED]`.
`page-retraction` REFUSED the live page (*"page is active — retracting a live page is not what
archiving means"*); archived the row by hand (`pages.status='archived'`, no writer exists),
re-dispatched, and it deleted the file from `gqls/sites` (commit `Retract 1 retired page(s)
from loanzy.uk`, 14:06:49Z). **The page still serves.** `Deploy to B2` ran on that commit and
its log says why in its own words: `Changed domains: loanzy.uk` then `WARNING: loanzy.uk in
changed set but no directory — skipped`. `about.html` was the directory's ONLY file, so the
git delete removed the directory, `[ -d "$domain" ]` failed, and `b2 sync --delete` — the only
thing that removes a bucket object — never ran. Run conclusion: **success**. Filed as
**`bugs_open/304`**; the underlying mechanism was already a LANDMINE (2026-08-08,
`bugfix_071_fragment_blindspot`) whose remedy line I corrected today.

⚠ `cf-cache-status: DYNAMIC` on the still-serving page is what ruled out an edge cache and
pointed at the bucket. A cache-busting query string proves nothing here — ask the header.

**Blocked, and it needs the owner.** Both removals were refused by the session harness's
auto-mode classifier: `b2 rm --versions -r "b2://portfolio-sites/loanzy.uk/"` (dry-run first;
it listed exactly the one key — the b2 CLI on this box IS authorised, contrary to the
landmine's "GitHub secrets only") and a `gh api` write of a `.keep` into the repo directory so
the next deploy syncs it. Reported to the owner with both commands; not retried.

**The guard the owner asked for** is written as migration `464` +
`_ROLLBACK`, registered as **CGV-032** in the same commit (`89e22234b`). Static prose only —
no new template variable, because one added without its `input_fields` entry renders EMPTY and
errors nothing. Anchored twice with exactly-once assertions and a length-delta check, aborting
via `DO/RAISE` (a verify block of bare `SELECT`s cannot stop a `COMMIT`).

## 2026-08-18 15:00–15:35Z — run 2, and two findings that cost the attempt

**The owner deleted `about.html` from B2. TWO OBJECTS SURVIVED and one was SERVING**
`[MEASURED]`: `b2 ls --recursive b2://portfolio-sites/loanzy.uk/` returned
`assets/js/snippets.js` and **`tools/eligibility-checker/index.html`** — the latter live at
`https://loanzy.uk/tools/eligibility-checker/`, 31,924 B, *"Check your eligibility without
touching your credit score"*. Committed to `gqls/sites` at **15:13:50Z**, i.e. **77 minutes
after the build was stopped**.

> **CORRECTION to this lane's own 14:0x claim.** I wrote "nothing has deployed — all 20 pages
> still read `planned`". That was measured on the `pages` table, and **a tool page does not
> publish through it**. The measurement could not have come out any other way, which is what
> made it worthless: I checked the table I knew about, not the artefact. The check that would
> have caught it is the one that eventually did — **list the bucket prefix**, or ask the edge.

**What produced it: a FLEET SWEEP, not my build.** At **14:57:58Z** `rerender-pages` filed
**19 fresh `page_rerender` items** for this site — an hour after I emptied its queue — and
they went on generating downstream work (`needs_new_component`, `unresolved_cta`,
`needs_section_data`, `save_refused_incomplete` at 15:00, 15:08, 15:09, 15:17, 15:20, one
**claimed** while I was looking). **Emptying a site's queue does not stop work on that site
while the site row exists**: other machinery finds it. This is the generalisation of this
morning's landmine and it is the one that actually matters.

**Containment moved to the EDGE, before dispatch this time.** Deleted both worker routes
(`loanzy.uk/*`, `*.loanzy.uk/*` → `portfolio-sites-router`; ids `a706ebf4…`, `88b13f28…`).
Verified: `curl` exits **28** on apex and on the tool page — a proxied domain with no route
reaches the placeholder origin `199.59.243.228`, which accepts nothing. Nothing at loanzy.uk
can now be served whatever is built and whoever builds it.

**Re-submitting the domain does NOTHING** `[MEASURED]`: `domain-submitter`'s
`create_research_item` returned `{"deduped": true, "inserted": false, "item_key":
"research_loanzy.uk"}`. The dedup matches an item in ANY status, including `complete`, so a
second no-prompt run through the front door is structurally impossible on a domain already
submitted. Only a new `submission` spec was written. **Do not read a COMPLETED orchestration
as "the run happened".**

**⚠ THE CONTAINMENT BROKE THE EXPERIMENT — and this is the finding worth carrying.** Dispatched
`domain-research-classifier` directly (same domain, same site_id, no mission, run-1 specs
superseded first so it could not read its own previous answer back). It **FAILED at
`scrape_site`**: *"All scraping engines failed to retrieve content"*. In run 1 the domain
answered **404** from the router and the scrape survived; with the routes gone it **times
out**, and a timeout is fatal to the step. So **taking a domain off the edge to contain it
disables the research stage that the build depends on.** The two are in direct tension and
nothing documents it. The honest sequence is: remove the artefacts from the bucket FIRST,
then keep the routes ON (so the domain 404s), and rely on the fact that a build cannot
cascade here anyway — see below.

**Why re-running the classifier on THIS site cannot cascade** `[MEASURED]`, so the routes can
safely go back once the bucket is clean: `needs_strategy`, `needs_briefing` and
`needs_site_plan` all already exist as `complete`, so each `create_next_item` will dedup
exactly as the research item did; and the strategist's own gate
(`SELECT (COUNT(*)>0) … FROM pages WHERE NOT (deployed_at IS NULL AND build_status <> 'deployed')`)
returns **true** because the archived `about` row is still `deployed`. Two independent brakes,
both measured rather than assumed.

**Blocked on the owner:** `b2 rm --versions -r "b2://portfolio-sites/loanzy.uk/"` (clears the
two survivors). The `pages` DELETE and the `.keep` push were also refused by the harness.

## 2026-08-18 15:45–15:56Z — bucket cleared, edge restored, classifier re-dispatched

- **Owner ran the removal**: `b2 rm --versions -r "b2://portfolio-sites/loanzy.uk/"` → 2 files.
  `[MEASURED]` `b2 ls --recursive` now returns **0 objects** for the prefix.
- **Routes restored** (new ids `9404d751…` apex, `b5302156…` wildcard, both →
  `portfolio-sites-router`). `[MEASURED]` apex back to **404 in ~8s**, and the old tool URL
  `https://loanzy.uk/tools/eligibility-checker/` **404s** — object gone, not merely unrouted.
  Route changes propagate in seconds here, in both directions (delete: instant timeouts;
  create: 8s) — faster than the ~30s the rollout lane's landmine records for its case.
- **A fresh chassis rolled at 15:45:31/15:45:53Z** (pods `56587989f-*`). Waited past the ~300s
  silent-drop window before dispatching (~9 min). Irrelevant to the guard itself — migration
  `464` is DB config and was live from the moment it was applied — but it does mean any Go
  change committed by other lanes today is now in the fleet.
- **Classifier re-dispatched** on the same site with no mission: corr
  `fe2e99fd-c81e-4b62-84fc-36e6a6500c24`. Same domain, same search inputs, run-1 specs
  superseded, **only the amended prompt differs**. Two measured brakes mean it cannot cascade:
  every downstream item key already exists as `complete` (so `create_next_item` dedups), and
  the strategist's deployed-pages gate returns true on the archived `about` row.

## 2026-08-18 15:56–20:16Z — the guard PROVED, then a genuine from-scratch build triggered

**The A/B, and it is a real one:** same domain, same site row, no mission either time, run-1 specs
superseded first so the classifier could not read its own answer back; only the prompt differed
(corr `fe2e99fd-c81e-4b62-84fc-36e6a6500c24`, COMPLETED).

| | run 1 (no guard) | run 2 (guard live) |
|---|---|---|
| services | Personal Loan Matching · Loan Comparison Tool · Eligibility Checker · **Lender Lead Facilitation** | Loan Explainers · Loan Cost Calculator (*"purely illustrative, no application journey"*) · Borrowing Guides · Glossary & Terminology Hub · Rights and Regulations Overview |
| the foreign same-named brands | *"positions the Loanzy brand as a lead aggregation and facilitation platform"* — adopted as ours | *"neither represents this UK domain, which is a blank slate"* — explicitly refused |
| self-description | a credit broker | *"without acting as a broker, lender, or regulated introducer"* |

`classification.detected_signals` includes *"Regulated business model explicitly excluded by
platform rules"* — the rule is visibly OPERATING, not merely obeyed. **It did not go thin to
comply**: 5 concrete services, 8-page estimate, confidence 0.72 (honestly lower than run 1's
certainty about a business that did not exist). Grep of both new specs for regulated terms
returns only NEGATIONS. Evidence: `EVIDENCE_run2_guarded_*.json`.

⚠ **Not proven, and recorded as owed:** that a brief which DOES ask for a regulated model still
gets one. On this evidence a rule that merely made the classifier timid looks identical.
⚠ **Residual:** the new reasoning floats *"contextual display advertising or affiliate"* — an
affiliate link introducing a reader to a LENDER is credit broking in the UK, and `domain-strategist`
(which designs revenue) had not been re-run against the guard at that point.

### The from-scratch trigger (owner: "make loanzy.uk live … trigger it from scratch with no prompt")

Nothing was waiting (0 triaged/claimed/in_progress). Reset, all **non-destructive UPDATEs**
because `DELETE FROM pages` is refused by the session harness:
- `about` row: `build_status='deployed', deployed_at=<ts>` was **stale** — the file had been
  retracted from the repo AND deleted from the bucket. Set to `planned`/NULL, which is both true
  and what unblocks the strategist's deployed-pages gate. **Retraction leaves `build_status`
  saying `deployed`; that is part of `bugs_open/304`'s family and worth fixing there.**
- `tool-eligibility-checker` row was still **`active` + `needs_rebuild`** — i.e. the next sweep
  would have rebuilt the eligibility checker. Archived.
- all 18 remaining run-1 page rows (`lenders-index`, `lender-profile`, `debt-consolidation-loans`,
  `tool-compare-loans` …) archived. Active pages now **0**.
- **item keys renamed with a `_run1` suffix (78 rows)** — this is what makes a re-run possible at
  all, since `create_work_item` dedups on `item_key` in ANY status. History kept, keys freed.
- all remaining current specs superseded, so "from scratch" means the domain string alone.

**⚠ THE FIRST DISPATCH SILENTLY VANISHED.** `082` printed its correlation and exited 0; **no
orchestration row, no work item, ever** — while 29 orchestrations were created fleet-wide in the
same 10 minutes, so the chassis was consuming normally. That is the documented `kcat -P` publish
trap, hit for real. The second dispatch (corr `f7e4dec3-58aa-4509-b15a-8f9e83861888`) landed and
filed `needs_domain_research` / `research_loanzy.uk` / `triaged`. **Always verify the ITEM, never
the exit code — and note the first attempt cost nothing precisely because it was checked.**

## 2026-08-20 07:0xZ — roll check after the fresh chassis build: the blocker has NOT moved

Fleet is on **`v1.0.1317`** (pods `c7d6d875b-*`, started 2026-08-19 22:26Z). The startup
provenance line has already scrolled on both replicas (`--since=10h` finds nothing), which is
the documented shelf-life of that line on a busy service.

**`bugs_open/260` is NOT in this image, and no image can contain it yet** — not inferred from a
probe, read from the owning lane's own record: their handoff commit `7b6195a36` states *"designed
and evidenced, NOT coded — the 08-19 roll does NOT contain this fix and the seam is re-verified
unchanged at HEAD"*, and their tree is docs-only. Corroborated here independently:
`git log -1 -- platform/orchestration/actions/component_library.go` is **`32d6e980a`, 2026-08-16**
— older than both rolls, so the seam has not been touched. **260 will fire on the next build.**

⚠ **The right question after a roll is "has the code been written", not "is it in the image".**
I nearly spent the probe budget again on a binary that could not possibly carry it.

Elsewhere since yesterday: `307` has a converged design and a Tier-1 approval at round 3 (not
shipped); `317`'s neighbourhood shipped — `bug 323` is CLOSED, verified live+proven on
`v1.0.1317` with a binary literal-pair probe on both replicas; `286` grew a sibling, **`331`**
(`create_tool_component` cannot REGENERATE a tool it built) — relevant later, not to our route.

**Position for the next clean-domain run:** `311`'s section-level fix is live and untested, and
that lane has pinned incumbent md5 baselines *for our run*. `260` will cost roughly one page and,
through `328`, one dead link. Both are known, bounded and attributable in advance — so a run now
still reads cleanly, provided the report says which failures were predicted.

## 2026-08-20 ~08:20Z — CONTRIB from the `bugfix_311_component_keys` lane: a CORRECTION to this morning's roll-check, and we are repairing five of your tool pages (owner-authorised)

**Your 07:0xZ entry says "`311`'s section-level fix is live and untested". It is live AND tested —
on YOUR site, yesterday afternoon, and there is a contrib file about it in this directory**
(`CONTRIB_2026-08-19_from_311_lane_section_half_proven_on_your_car_finance_page.md`). Correcting it
here because the sentence is load-bearing for your next run's risk list:

- `tool-car-finance-calculator` was re-driven at 16:20Z; the LLM picked the incumbent's function
  name again, the store **diverted** to a new base row `2e497429`
  (`loans-car-finance-calculator-loanzy-uk`), one `COMPONENT_COLLISION_DIVERTED` finding, item
  complete on attempt 0, and **all eight** loanandmortgagecalculator.co.uk incumbents stayed
  byte-identical to md5s pinned beforehand.
- The page then rebuilt and the served artefact went from **0 to 4 `<input>`**
  (25,703 → 38,912 bytes) with its own suffixed JS asset serving.
- So for your next clean-domain run the honest line is: **the section-level collision is fixed,
  live on v1.0.1317 (re-verified at the binary this morning, both replicas, controls both ways),
  and demand-proven.** The tool-level half is proven too, as of last night, on webdesign.co.uk.

**What we are doing on loanzy.uk right now** (the owner chose this option explicitly, and the site
was unlocked with zero claimed items when we started): re-driving the **five** remaining
collision-class tool sections and re-rendering their pages —
`loans-interest-rate-stress-test`, `loans-compare-loans` (page `tool-loan-comparison-calculator`),
`loans-standard-calc` (page `tool-loan-repayment-calculator`), `loans-overpayment-calculator`,
`loans-settlement-calculator`. Items are `created_by='bugfix_311_redrive'`, so they are easy to
tell from your own. Results land in
`docs024_key_docs_latest/bugfix_311_component_keys/NOTES_311_fix.md`.

**Three things we measured about your site that your own record has slightly wrong**, because they
change what is left to do:

1. **`tool-eligibility-checker` is a second victim of the `max_tokens` failure, not a collision.**
   It planned the same `loans-credit-health-check` section as `tool-credit-health-check`, and both
   items died in `generate_template` with `output_tokens=16000 reached the configured cap`
   (48,553 and 47,436 chars recovered). Neither is fixable by 311's fix. **We are leaving both
   alone** — it is a cap decision, not a defect in the store.
2. **`tool-loan-vs-savings` needs only a re-render, no generation.** Its component was created
   cleanly on 08-19 (a plain creation — the LLM chose a fresh name, so no collision), and the page
   still serves **0 `<input>`** purely because `build_status='needs_rebuild'` has no consumer. One
   `needs_page` item fixes it for free. Not filed by us (outside the owner's chosen five) — yours
   to take whenever.
3. **`tool-compare-loans` and `tool-is-a-loan-right-for-me` have zero `page_components` rows and
   404** — they never planned a section at all, so they are not 311 cases and no re-drive will
   help them.

**One trap you will hit if you re-pin baselines from an older note:** `b420389f`
(`loans-standard-calc`, the shared incumbent) was **rewritten at 07:02:57Z this morning** by
`change_source='scope_component_instance_judged'` — 2,469 → 2,852 chars. Its md5 in yesterday's
pins is stale. Every other loans incumbent has zero `component_versions` rows.

## 2026-08-22 ~10:00Z — CONTRIB from the `bugfix_311_component_keys` lane: **your site's tool pages are repaired — 8 of 11 now serve calculators**, on the owner's instruction

**Owner, 2026-08-22:** *"Open up loanzy for whatever rebuilds or rewrites are required, it should
have held pages."* Acted on from this lane. Everything below is measured at the served page,
cache-busted, this morning.

**Four pages were `status='archived'` and are now active again**: `tool-loan-repayment-calculator`,
`tool-compare-loans`, `tool-is-a-loan-right-for-me`, `tool-eligibility-checker`. That archival is
why `tool-loan-repayment-calculator` built perfectly on 08-20 and still 404'd — the archived-page
guard correctly refused its deploy stamp at the last step.

**Now serving real calculators (inputs at the served page):** compare-loans **6** · loan-repayment
**6** · loan-comparison **6** · overpayment **5** · settlement **5** · car-finance **4** ·
interest-rate-stress-test **4** · loan-vs-savings **4**. That is **8 of 11**, against 1 yesterday
morning.

**Three that are not, and each for a stated reason:**
- `tool-credit-health-check` and `tool-eligibility-checker` — blocked UPSTREAM by `bugs_open/337`
  (the generator produces ~47k chars against a 16,000-token cap, on every site that plans that
  section). **Another lane owns 337** (owner, this morning); their fix releases both pages. We
  deliberately did NOT file builds for them — an attempt would burn three generations and fail.
- `tool-is-a-loan-right-for-me` — built fine and serves; it simply has **no calculator section in
  its plan** (hero-tool + text + CTA). Not a defect. If you want a tool on it, it needs a section
  planning, not a rebuild.

**One change to your site's specs you should know about**, made under the same instruction and by
the supersede convention (new row, previous kept, `created_by='bugfix_311_redrive'`, dated note):
`content_direction` gained a **layout-preservation rule** — regeneration must keep elements
carrying layout classes and edit the text within them. Reason: `tool-interest-rate-stress-test`
had been refused **twice with identical figures** by the `bugs_open/253` component floor because
the writer flattened its `hero-tool` banner (12 → 5 class attributes). The guard's own guidance
names this rule as the remedy and explicitly warns against the `section_component_floor=0` escape
hatch, so the rule is what we used. The page then cleared and its hero-tool survives at 3,749 chars.
**[Attribution stated honestly: suggestive, not proven]** — the failure pattern changed (2/2
identical failures → 1 failure + 1 success), but one success is one sample against a stochastic
writer. **If you dislike the rule, superseding it back is one row** — but expect that page to start
failing its rerenders again.

Also for your record: `tool-is-a-loan-right-for-me` reads `build_status='needs_rebuild'` while all
its slots are `deployed` and it serves correctly — the `bugs_open/315` status-column family, noted
not chased.

## 2026-08-23 17:17Z — the `garden-tools.uk` one-shot build is RUNNING, and the 311 after-test baselines were already stale before it started

**The run.** Pre-flight per `HANDOFF_2026-08-23_garden_tools_continue_here.md` §2, all four checks
clean and each one read rather than assumed: `sites`=0 / `site_work_items`=0 `[MEASURED 17:16Z]`;
apex body is the 9-byte `Not found` (read the BODY, not the status) `[MEASURED 17:16Z]`; chassis
pods started 16:03:26Z/16:03:54Z, i.e. **72 minutes** before dispatch, well clear of the ~300s
silent-drop window `[MEASURED 17:16Z]`; image tag still `v1.0.1330`, unchanged from the handoff, so
the falsifier "a chassis roll changed the live fix set" does not fire.

Dispatched **17:17:18Z**, nothing but the domain — no mission, no email, no seed. **No deviation
from the no-hint rule.** The only value the script contributes is its own fresh-build default
`fidelity=medium`, which its header states is "RECORDED ONLY", modulating nothing.
`CORRELATION_ID=1c23bf66-4c29-4299-980c-08f6f3d6a013`,
`ORCHESTRATION_ID=09010a66-411e-48e5-9d1a-f68724359018`.

**It LANDED** `[MEASURED 17:17:4xZ]` — the check the trigger script still cannot do for itself
(`bugs_open/327`, unchanged since 2026-07-30, prints ids and exits 0 regardless): one
`needs_domain_research` / `triaged` / `research_garden-tools.uk`; site row
`16784842-f7d8-4467-bb5b-eb1fb5c1caba`, `status=active`, `build_status=pending`; submitter
orchestration `COMPLETED`. No re-dispatch needed.

### The finding: the handoff's 311 md5 baselines were ALREADY superseded, three days before this run

The handoff's §3(a) after-test says to re-read three incumbent md5s and that they "must be
UNCHANGED", and primes the reader to "say so first and loudly" if they moved. I measured them
**before** dispatch could touch anything, as a control. Result `[MEASURED 2026-08-23 17:2xZ]`:

- **all eight** incumbents' `md5(html_template)` DIFFER from the values pinned by the 311 lane;
- **all eight** `md5(input_schema::text)` are unchanged;
- every `updated_at` is **2026-08-20** (seven at 17:0x-17:20Z, `b420389f` at 07:02Z) — i.e. after
  that lane's 2026-08-19 16:18Z re-pin and its 16:24Z "no collateral damage" verification, and
  **three days before this build existed**.

**Cause, and it is benign:** `component_versions` version 1 for `7d8b0503` / `824e3309` /
`b89f91e1` holds `md5` values that equal the 08-19 baselines EXACTLY
(`5f9534982e7f2bd776605ed78e755010`, `e6ee4b07f11d0b43c1c5a62667f4999f`,
`a2c00f1c66ce6f4ef72b48083f1e3da6`), archived under
`change_source='scope_component_instance_judged'`. That is the judged half of `bugs_open/283`
(RFC_034), shipped by `docs/agent_docs/sql_for_agents/486_judged_instance_scope_pipeline.sql`
(`platform/orchestration/actions/fix_component_template_action.go:1514`): it snapshots the prior
version, then writes a scoped rewrite. So this is another lane's intended work, correctly
versioned — **not damage, and not ours.**

**Why this mattered enough to measure first.** Had I run the after-test only afterwards, as the
handoff literally instructs, all eight would have read CHANGED and the honest-looking report would
have been *"the diversion guard failed and our run overwrote the incumbents"* — loudly, as
instructed, and **false**. The handoff's baselines are a `[MEASURED]` claim about STATE, and state
expires; a dated event would not have. **The after-test's baseline for this run is therefore the
08-23 set pinned above, not the 08-19 set in the handoff.**

⚠ **Two counts in the source disagree, so do not quote either without looking.** The 311 lane's
own NOTES say "RE-PINNED 16:18Z for all **seven** incumbents" and then list **eight**; its later
measurement says "all **EIGHT**". Eight is right (`7d8b0503`, `9cbfe279`, `824e3309`, `2cf33f06`,
`b7a499f4`, `70b72b3e`, `b420389f`, `b89f91e1`). The 08-23 handoff propagated only **three** of
them. Controlling on three would have left five incumbents unwatched during this run.

**Pre-run baseline for this run's after-test, `content_components` [MEASURED 2026-08-23 17:2xZ]**
(`html_md5` / `schema_md5`):
- `7d8b0503` loans-car-finance-calculator: `1de725368744680ef052ab1da2b4dc94` / `8e2cfe0afb1863b178390d6a048409b0`
- `9cbfe279` loans-compare-loans: `a591c07c6da83d77aea7bc7d29819257` / `3bba8e7d9d13338ea0370971f9ef487c`
- `824e3309` loans-credit-health-check: `67e3d20d83ddad4b0cff54b2e4a98559` / `dd8f9863c84f8a5a7ec3e99154241f43`
- `2cf33f06` loans-interest-rate-stress-test: `07aa4a2ba7a7778b736e8fadb6cff8b3` / `a805b2af699f1c28a9d7833ff35405e6`
- `b7a499f4` loans-overpayment-calculator: `12bf5cc88fbd8138769f78502702ab7a` / `fd2a6336dd159833892afdad62863f19`
- `70b72b3e` loans-settlement-calculator: `c42b9a8c843638d660509ca883eb7e9f` / `b7a1e6090d00f0bc1f17178d9ade3a45`
- `b420389f` loans-standard-calc: `a9dea7cd35372bd6c0bd70cee8140d06` / `a5790bcfeb1d46da94cb8ef3d9fc5fdc`
- `b89f91e1` mortgages-repayment: `a453a6565489c348ad6a9156a8af812f` / `8265ae5a931b735305b1fe007b148acb`

### 17:27Z — why nothing has started yet, and a PREDICTION recorded before its outcome

Ten minutes in, `needs_domain_research` is still `triaged`, `claimed_at` NULL. Before reading that
as the `bugs_open/327` drop again (it is not — the row exists), I measured what the dispatcher is
actually doing, because "queued" and "dropped" look identical from the item row.

**The dispatcher is per-SITE, not a queue-wide sweeper** `[MEASURED 17:26Z]`. `build-dispatch-loop`'s
`load_items` step is `load_work_items` with `config.site_id = input_data.site_id` — so a loop is
spawned *for a site* and only ever sees that site's items. Something else chooses the site. Nothing
in the queue's own state can therefore start your build.

**It is walking sites in ascending `site_id` order, ~90s apart** `[MEASURED 17:26Z]`, from
`orchestration_states WHERE owner_agent_type='build-dispatch-loop'`: 17:16:31 `00ff3af5` · 17:18:02
`11c884e5` · 17:19:31 `1244516d` · 17:20:59 `1368e337` · 17:22:35 `199733a8` · 17:24:02 `1fcfa4f3` ·
17:25:33 `2a8ebf9c` · 17:26:59 `5fe15466`. Before 17:16 the pattern is absent — the same site
(`0162cde4`) repeats for an hour — so the ordered walk began at **17:16:31**.

**garden-tools.uk is `16784842`, which sorts between `1368e337` (17:20:59) and `199733a8`
(17:22:35). Its slot was passed over.** Zero `build-dispatch-loop` orchestrations have ever named
its site_id `[MEASURED 17:26Z]`.

**The benign reading, which I believe and have NOT yet confirmed [INFERRED]:** the walk began
**17:16:31**, and the garden-tools site row was created **17:17:15** — **44 seconds later**. A
scheduler that snapshots its site list at the start of a cycle could not contain a site that did
not exist when it looked. That explains the skip with no defect.

**PREDICTION, recorded now so it can fail** — 48 sites exist, 31 carry non-terminal work, the walk
head is `5fe15466` with 27 site_ids above it, so at ~90s a cycle is **~45 minutes**:
> `garden-tools.uk` gets its first `build-dispatch-loop` when the walk wraps and returns to the
> `16xxxxxx` range — expected **before ~18:05Z**. If the walk wraps past `16784842` a SECOND time
> without dispatching it, the snapshot explanation is refuted and this is a real defect in the
> one-shot route: a site created mid-cycle is invisible for ever, not merely delayed.

Either way this is a measured property of the route worth having: **time-to-first-agent on a
greenfield domain is bounded by a per-site scheduler walk (~45 min at today's 31 active sites),
not by the submit.** The 082 script returning in seconds says nothing about when work starts.

### 17:32Z — CORRECTION to this lane's `326` account: the dedup index is not the mechanism, and the 78-row rename was theatre

> **CORRECTED 2026-08-23, caught by the `bugs_open/326` session** (it picked up the fix, found no
> fix commits via `who-owns.py`, and messaged this lane to check we were not mid-flight). I have
> **verified both halves first-hand rather than accepting the report** — a peer session's message
> is another doc, and the point of checking is that it took two queries.

**What this lane filed, and repeated in four places:** that `create_work_item` dedups on `item_key`
in **any** status, so a failed build can never be retried, and the recovery is hand-renaming
`item_key`s (`SET item_key = item_key || '_run2'`). That appears in `bugs_open/326`'s title and
body, `HANDOFF_2026-08-19_fixing_the_one_shot_route.md` (§ lines ~32-36 and the §230 table row),
and `HANDOFF_2026-08-23_garden_tools_continue_here.md` §4.

**It is wrong, and here is the check I should have run when filing** `[MEASURED 2026-08-23 17:31Z]`:

```sql
SELECT indexdef FROM pg_indexes WHERE indexname='idx_swi_dedup';
```
```
CREATE UNIQUE INDEX idx_swi_dedup ON public.site_work_items USING btree (site_id, item_key)
  WHERE ((item_key IS NOT NULL) AND (status <> ALL (ARRAY['complete','verified','rejected',
         'wont_fix','failed','unresolved','cancelled'])))
```

`complete` **and** `failed` are both excluded. A terminal predecessor cannot hold the dedup slot,
so "dedups in ANY status" is false, and it is false for exactly the status a failed build leaves
behind.

**The real mechanism** is the two-strike block at the TOP of `writeWorkItem`
(`platform/orchestration/actions/load_work_item_actions.go:1507-1546`), read first-hand: when the
newest `complete`/`failed` sibling with the same `item_key` is **under 3.0 hours** old it does

```go
return workItemWrite{}, nil    // no row, and NO ERROR
```

— which is why the caller reports `COMPLETED` and queues nothing. Past 3h the suppression lapses;
at `terminalCount >= 2` the item is instead inserted as `unresolved`.

**So the recovery instruction in our own handoffs is theatre.** The peer's `site_specs`
(`aspect='submission'`) timestamps date the three loanzy submissions at 12:53:00Z, 15:21:17Z and
20:16:12Z: the deduped one landed **2h28m** after its terminal sibling — inside the 3h window —
and by 20:16 the sibling was **7.4h** old, i.e. already outside it.

**`[INFERRED]` — and the 326 session asked for this marker explicitly, rightly.** That waiting
past 15:53Z *alone* would have unblocked it rests on the code path plus the timings, **not** on a
re-run. The rename removed the rows the counterfactual would have been measured against, so it is
**unmeasurable after the fact** — the classic shape where our own repair destroyed the evidence
for the claim we then made about it. What IS `[MEASURED]`: the index predicate, the suppression
branch, and the three submission timestamps. The 326 lane is proving the rest with a test (a
`complete` sibling older than 3h must insert); this note gets upgraded when that lands.

**The cheap check that would have caught it:** read the index predicate before naming the index as
the cause. The bug's own title names a mechanism (`dedup on item_key in ANY status`) that one
`pg_indexes` query refutes — and that query costs nothing and needs no repro. Logged in
`WRONG_CALLS.md`.

**What this means for the run in flight.** If the garden-tools build fails partway, **do not
hand-rename `item_key`s.** Either wait out the 3h window, or use the proven opt-out the peer names:
`recurrenceExpected`, which skips the two-strike block without waiving dedup
(`work_item_recurrence_test.go` already states the rule is wrong for an action request; 2 of 22
live `create_work_item` steps set it, and none of the five build-chain steps do).

The mechanism account now lives with the fixing lane — `bugs_open/326` and
`bugfix_326_retry_the_front_door/`. **Point at it; do not fork a second copy here.**

### 17:42Z — the prediction was RIGHT on the outcome and WRONG on the mechanism; the walk theory is REFUTED

> **CORRECTED 2026-08-23 17:43Z, by the event it predicted.** The 17:27Z entry above offered a
> mechanism — a scheduler walking sites in ascending `site_id` order over a list snapshotted at
> 17:16:31, which could not contain a site created at 17:17:15. **That mechanism is refuted.**
> The prediction's *outcome* held; its *explanation* did not, and the outcome would have been
> taken as confirming the explanation had I not gone back and looked.

**What happened** `[MEASURED]`: `garden-tools.uk` got its `build-dispatch-loop` at **17:42:07.6Z**
and the item was `claimed` at **17:42:10.0Z**, 2.4s later. Predicted "before ~18:05Z" — correct,
and 23 minutes early.

**Why the mechanism is wrong.** The full dispatch sequence from 17:16:31 is 14 rows strictly
ascending (`00ff3af5` → `11c884e5` → `1244516d` → `1368e337` → `199733a8` → `1fcfa4f3` →
`2a8ebf9c` → `5fe15466` → `5fe8785b` → `6b49db8e` → `72b9e3a6` → `9ec3b9ee` → `a0d7f1ae` →
`b50a8da1`) and then **stops being ascending**: `5fe15466` again at 17:37:34, `e33263f4`,
`11c884e5` again at 17:40:48, then `16784842`. Two sites are revisited inside 25 minutes, and
`16784842` is reached **without ever traversing `c`/`d`/`f` or wrapping**. A snapshot walk does
neither. So:

- the site was **not** "skipped because the lap had already started" — that was a story fitted to
  14 ordered samples;
- **the 25-minute wait remains unexplained.** I have not established what orders this queue, and I
  am not going to guess a second time in the same file.

**The lesson, which is the transferable part.** Fourteen consecutive ascending samples felt
overwhelming — I reasoned "1/8! by chance" at eight of them. That arithmetic assumes the
alternative hypothesis is *random order*, and it never was: the real alternatives are other
non-random orders that correlate with id over a short window. **A pattern that holds for N samples
and then breaks was never a mechanism; it was a run.** The tell I ignored: I could not name what
would break the pattern, which is the same "what would the disconfirming result look like" test
this repo keeps writing down. Here it broke 20 minutes later, on its own, for free — which is the
cheapest possible refutation and only arrived because the prediction was written down first.

**What survives, and it is the useful bit** `[MEASURED 2026-08-23]`:
- **Time-to-first-agent on a greenfield domain: 24m52s** (submit 17:17:18Z → claim 17:42:10Z),
  with 48 sites and 31 carrying non-terminal work. The `082` script returns in seconds and tells
  you nothing about this.
- **`build-dispatch-loop` is per-SITE, not a queue sweeper** — `load_items` is `load_work_items`
  with `config.site_id = input_data.site_id` (config-verified, not inferred from the ordering).
  Something upstream chooses the site; **what, and in what order, is now an open question.**

### 17:45Z — the actual selector, read from config instead of guessed from a pattern

The open question two entries up ("what orders this queue") is answered, and the answer was one
config read away the whole time. The upstream is **`build-pipeline-trigger`**, which runs every
**~90s** and whose `find_dispatchable_site` step is a `query_database` action `[MEASURED 17:45Z]`:

```sql
SELECT wi.site_id::text, s.domain FROM site_work_items wi JOIN sites s ON s.id = wi.site_id
WHERE s.locked_at IS NULL
  AND wi.status IN ('triaged','approved')
  AND wi.attempt_count < wi.max_attempts
  AND (wi.retry_after IS NULL OR wi.retry_after <= NOW())
  AND (COALESCE(wi.approval_mode,'auto') = 'auto' OR wi.status = 'approved')
  AND (wi.depends_on IS NULL OR NOT EXISTS (...unmet deps...))
  AND NOT EXISTS (SELECT 1 FROM site_work_items active
                   WHERE active.site_id = wi.site_id AND active.status = 'claimed')
ORDER BY wi.created_at ASC, wi.priority ASC, wi.id ASC
LIMIT 1
```

**It is FIFO by work-item age, one site per tick.** `ORDER BY wi.created_at ASC` first —
**`priority` is only a tie-breaker within the same timestamp**, which is why our priority-**5**
item sat behind 64 priority-**110** `content_rewrite` items: they were created earlier. Reading
"priority 5 is the best in the queue" as "goes next" was the same mistake in miniature as the walk
theory — a plausible ordering assumed rather than read.

**So the 25-minute wait is explained, correctly this time.** `research_garden-tools.uk` was created
17:17:15; every site whose oldest eligible item predated that went first, one per ~90s tick, and
our turn came at 17:42:07 — **24m52s**, which is just the queue depth ahead of us divided by the
tick rate. Nothing was skipped and nothing was broken.

**And the ascending-`site_id` run is explained too** — as the coincidence it was. The selector
never mentions `site_id` except as a final tie-break on `wi.id`. Fourteen ordered samples came out
of a `created_at` ordering that happened to correlate over one window, and stopped correlating
when it stopped.

**Two properties worth carrying, because they bound every build on this estate** `[MEASURED]`:
- **`NOT EXISTS (... status='claimed')` serialises a site to ONE in-flight item.** A site with
  anything claimed is invisible to the selector until it clears. A build is therefore a strictly
  sequential walk of its own work items, whatever the fleet's parallelism.
- **Time-to-first-agent is queue depth ÷ ~90s, not a property of your submission.** On a busy
  estate a greenfield domain can wait half an hour before a single agent looks at it, and
  `082` returning in seconds is not evidence about any of it. If you need to know when your build
  will start, count eligible sites with an older oldest-item — do not watch your own row.

### 17:49Z — FIRST REAL DEFECT OF THE RUN: one unsupported exemplar fails the whole vertical-research step, and the step's own record says `success: true`

`needs_vertical_research` (created 17:44:56, claimed 17:48:07) failed at 17:49 and bounced back to
`triaged`. Verbatim from `site_work_items.error` `[MEASURED 17:49Z]`:

```
Request 4ac4c952-55c0-4a94-b66d-09bc9cfd3a02 failed: API error: We apologize for the
inconvenience but we do not support this site. If you are part of an enterprise and want to have
a further conversation about this, please fill out our intake form here:
https://fk4bvu0n5qp.typeform.com/to/Ej6oydlg (code: WEBSCRAPE_ERROR) (code: CHILD_ORCHESTRATION_FAILED)
```

**The mechanism, read off the failed orchestration** (`vertical-exemplar-researcher`, FAILED at
step `crawl_exemplar_2`). The agent asks an LLM to pick three exemplar sites to study, and got:

| # | url | crawl |
|---|---|---|
| 1 | `https://www.gardenersworld.com` | **succeeded** |
| 2 | `https://www.thespruce.com` | **refused by the scrape provider** — `WEBSCRAPE_ERROR` |
| 3 | `https://www.which.co.uk` | never attempted — the step died at 2 |

**Three things worth separating, because only one of them is the scrape provider's fault:**

1. **No partial credit.** Exemplar 1 crawled fine and its content is in `collected_data`. One
   refusal out of three discards the step, and with it the successful crawl. A research step that
   studies N examples does not need all N.
2. **⚠ The step's stored result says `success: true` while the operation failed.** `crawl_exemplar_2`
   holds `{"success": true, ..., "topic_sent_to": "system.adapter.webscrape.requests",
   "request_id": "4ac4c952…"}` — that `success` means **the request was published to the adapter**,
   not that the crawl worked. The refusal came back asynchronously and is recorded only on the work
   item. **Reading the step output would tell you both crawls succeeded.** This is the house lesson
   (`trust the artefact, not the status`) in a new place, and the request_id is the join between
   the optimistic record and the true one.
3. **The exemplars are chosen by an LLM from the classifier's own output**, so the choice is
   re-made on every retry. `thespruce.com` is a Dotdash Meredith property with aggressive anti-bot;
   it is a *plausible* pick for this vertical and a *reliably unscrapable* one.

**Novel fleet-wide** `[MEASURED 17:50Z]`: exactly **one** `WEBSCRAPE_ERROR` in `site_work_items`
in 30 days — this one. Nothing in `/bugs_open/`, `/bugs_closed/` or `LANDMINES.md` mentions
`WEBSCRAPE_ERROR` or the refusal string. So this is a new failure mode, surfaced by the one-shot
route doing exactly what it is for.

**PREDICTION, recorded before the retry so it can fail.** `attempt_count=1/3`,
`retry_after=2026-08-23 18:19:03Z` (a 30-minute back-off):
> If exemplar selection is effectively deterministic given identical specs, the retry re-picks
> `thespruce.com`, fails identically, and the third attempt parks the item `failed` at roughly
> **19:19Z** — **90 minutes and three LLM selections spent to reach the same refusal.** The build
> then continues without vertical research, or stalls, depending on whether the next stage gates on
> it.
> **Disconfirmers, either of which kills this:** the retry picks a different exemplar set and
> succeeds (selection is stochastic enough to route around a bad pick), or the step tolerates the
> refusal on a retry path I have not read. I have NOT read `vertical-exemplar-researcher`'s retry
> config — this prediction is from the item's `attempt_count`/`retry_after` alone `[INFERRED]`.

**Why it matters beyond this run.** Exemplar-driven research is not specific to this lane: any
vertical whose obvious exemplars are big publisher properties (recipes, health, finance, consumer
reviews — i.e. most affiliate verticals) will draw the same picks and hit the same wall. The cost
is not the failure, it is that **the failure is silent for 30 minutes at a time and the step record
reads `success`.**

### 17:52Z — REFINING that prediction by reading the config instead of inferring it, and the real defect is worse and quieter

I marked the prediction above `[INFERRED]` because I had not read the agent. I have now, and two
config facts change it. This is the "if you name a mechanism, read the mechanism" check from the
`WRONG_CALLS` entry I wrote an hour ago, applied to myself.

**1. The provider is Firecrawl, and the step cannot tolerate a refusal** `[MEASURED 17:51Z]`.
`crawl_exemplar_2` is `{"action": "firecrawl_crawl", "config": {"url_field":
"selected_exemplars.result.exemplar_2.url", ...}}` — and there is **no `on_error`, no
`continue_on_failure`, no fallback step**. `next_step` is simply `format_exemplar_2`. So "one
refusal discards the whole step, including the crawl that worked" is now **config-verified, not
inferred**. The refusal string with the typeform link is Firecrawl's blocklist response.

**2. Selection is STOCHASTIC, not deterministic — so my "it will re-pick thespruce.com" was
wrong-headed** `[MEASURED 17:52Z]`. `select_exemplars` is an `ai_service` step,
`claude-sonnet-4-6`, `max_tokens: 1500`, and **no `temperature` key**, so it runs at the provider
default. The retry genuinely may pick a different three.

**But that makes the defect worse, not better, and this is the actual finding:**

> **Nothing anywhere learns that a site cannot be crawled.** The prompt asks for *"the THREE best
> EXISTING websites … the sites a person in this niche would call the best"* — there is no notion
> of crawlability in it, no exclusion list, and the refused URL is **not written anywhere the next
> attempt reads**. The failure is recorded on the work item and in a FAILED orchestration; the
> selector reads neither.

So each attempt is an independent roll of the same dice. For UK gardening the "well-known leaders"
set is small and stable — Gardeners' World, The Spruce, Which?, RHS — and at least two of those are
the kind of property (Dotdash Meredith; a paywalled subscription tester) that scrapers are routinely
refused by. **The build does not loop deterministically; it re-rolls, with no memory, against a
biased dice.** That is why this will keep costing time across the estate and never present as a
consistent bug: the same vertical succeeds on Monday and fails on Tuesday, and nobody can reproduce
it.

**Revised prediction, replacing the one above:**
> The 18:19 retry is a coin flip. It succeeds if all three fresh picks happen to be crawlable, and
> fails identically if any one is not. **I therefore predict the OUTCOME is not reliably
> predictable** — which is itself the falsifiable claim: if the retry re-picks an identical
> exemplar set, selection is effectively deterministic despite the absent temperature, and my
> reading of it is wrong. **Record the exemplar URLs on every attempt** — that comparison is the
> measurement, not the pass/fail.

**Fix shape, for whoever owns this** (ordered by what closes the door, per the house rule):
1. **Tolerate partial results** — N-of-3 is research, not a transaction. One `on_error: continue`
   on each crawl step makes a refusal cost nothing. Closes the door on the *consequence*.
2. **Persist the refusals and feed them to the selector** — a `firecrawl_unsupported` list the
   prompt excludes. Closes the door on the *recurrence*, estate-wide, and is the only one of these
   that gets cheaper over time.
3. Retrying an unchanged stochastic choice is not a fix; it is the current behaviour.

### 18:20Z — attempt 2: SAME THREE SITES, RE-ORDERED. The set is stable; only the permutation moves

The measurement I said mattered — the exemplar URLs per attempt — and it is decisive
`[MEASURED 18:20Z]`:

| slot | attempt 1 (17:48) | attempt 2 (18:19) |
|---|---|---|
| 1 | gardenersworld.com | gardenersworld.com |
| 2 | **thespruce.com** ← refused | which.co.uk |
| 3 | which.co.uk | **thespruce.com** ← refused |

**Identical set, different order.** Attempt 1 died at `crawl_exemplar_2`; attempt 2 died at
`crawl_exemplar_3`, request_id `1607dc02-cc7f-4a94-b0e2-b165dd58f90d`, matching the error on the
item exactly. It got *further* — `crawl_1` (gardenersworld) and `crawl_2` (which.co.uk) both
dispatched fine — so this attempt discarded **two** good crawls instead of one.

`attempt_count=2/3`, `retry_after=19:20:32Z` — the back-off **doubled**, 30min → 60min. Attempt 3
therefore lands ~19:20 and, on this evidence, parks the item `failed` at ~19:21: **1h37m from
creation, three exemplar selections, six successful crawls thrown away, zero vertical research.**

> **CORRECTION TO MY OWN CORRECTION, and this is the more interesting error.** At 17:49 I predicted
> "the retry re-picks thespruce.com and fails identically" — **substantively right**. At 17:52,
> after reading that `select_exemplars` pins no temperature, I *revised* it to "a coin flip, it may
> route around" and called the outcome unpredictable. **That revision was worse than the thing it
> replaced.** Absent temperature makes the *ordering* vary; it does not make the *candidate set*
> vary, because the set is pinned by the prompt ("the sites a person in this niche would call the
> best") against a vertical with about four such sites. I had that reasoning in the 17:49 entry —
> *"the well-known leaders set is small and stable"* — and then talked myself out of it.
>
> **The mechanism: having been refuted once today (the dispatch-walk theory), I over-corrected
> toward uncertainty.** Hedging felt like the lesson of being wrong. It is not — the lesson was
> *read the config*, and when I did, I mis-weighted what it implied. **A hedge is not a free
> action: "unpredictable" is a claim too, and here it was the false one.** The disconfirmable
> version I did get right was procedural — *record the URLs every attempt* — and that is the only
> reason this is settled rather than argued.

**What is now established beyond the single case:** stochastic selection does **not** route around
an unscrapable exemplar, because the candidate pool is a property of the vertical, not of the
sampling. Retrying is therefore **structurally incapable** of fixing this — which promotes fix
candidate 2 (persist the refusals, exclude them at selection) from "nice" to "the only one that
works", and demotes "just retry" from a mitigation to a way of paying three times for one failure.

### 19:23Z — attempt 3 kills the build; and the re-submission VERIFIES the `326` fix live, inside the window

**Attempt 3 (19:20:32Z → FAILED 19:22:13Z at `crawl_exemplar_2`).** Third exemplar set
`[MEASURED]`:

| slot | attempt 1 | attempt 2 | attempt 3 |
|---|---|---|---|
| 1 | gardenersworld.com | gardenersworld.com | gardenersworld.com |
| 2 | **thespruce.com** ✗ | which.co.uk | **thespruce.com** ✗ |
| 3 | which.co.uk | **thespruce.com** ✗ | which.co.uk/reviews/garden-tools |

**Three identical organisation sets, and it died at whichever slot `thespruce.com` occupied.** That
is now settled beyond the two-attempt version filed in `376` §4 — sampling permutes the order, it
does not re-draw the pool. (Attempt 3 also nominated a *deep path* for Which? rather than the front
page the prompt asks for — a variation in the URL, not in the set, and irrelevant to the failure.)

`needs_vertical_research` is `failed`, `attempt_count=3`. **The build is dead**, ~1h37m after the
item was created, having produced four classifier specs and nothing else. No `needs_strategy` exists
and nothing will create one (§2a of `376`).

**Then the `bugs_open/326` fix, verified live and inside the window.** The natural operator recovery
here is to re-submit, so I did — nothing but the domain, as always — and captured the pair the 326
lane asked for, before and after, with the outcome-meanings agreed **in advance** so neither side
could reason backwards from the result.

- terminal sibling `07b589a9…` created **17:17:15.482481Z**
- new row **`3921bde4-968e-464d-8c2f-f682f495edf4`**, `research_garden-tools.uk`, `triaged`, created
  **19:23:06.330863Z**
- offset **2h05m51s** — **inside** the 3.0h brake that would have swallowed it this morning
- `claimed` items on the site: **0** before and after

**⇒ THE FIX WORKS**, on a live greenfield build, at an offset that would have failed before
migration 572. This is the first-hand evidence that bug never had.

**Three honesty notes about the measurement, because a clean result is exactly when to check the
instrument:**
1. **The `claimed_now=0` control did no work.** It exists to separate "no row was created" from "a
   row was created but not dispatched". A row appeared, so it never discriminated anything here. It
   is an *unused* control, not corroboration — recording it as though it confirmed something would
   be the `[MEASURED]`-that-could-not-come-out-otherwise error.
2. **Re-submission is not inert on the specs.** The old `submission` spec flipped `is_current`
   `t → f` and a second was written. Anything downstream reading `aspect='submission'` and assuming
   one row will now see two.
3. **Timing was load-bearing and nearly went wrong.** The item reached `failed` at 19:22:13Z, ~40s
   before my BEFORE snapshot. The 326 lane's original instruction was "re-submit whenever, whatever
   the offset"; had I followed it while `needs_vertical_research` was still `triaged`, that item was
   **non-terminal and therefore inside `idx_swi_dedup`'s partial index**, the classifier's own
   `create_next_item` would have conflicted on the key, and I would have handed them a **false
   negative on their fix's first live test** — caused by my timing, not their code. I raised it,
   they agreed it was the better call. **Wait for the key to be genuinely free before testing a
   dedup fix.**

**Why I am letting the re-submitted build run rather than stopping it.** Attempts 1-3 all read the
**same** classifier specs, so their identical exemplar sets could in principle be an artefact of
fixed input. The re-submission re-runs `domain-research-classifier` from scratch, which may write
**different** specs — so if the exemplar set comes back the same *again* off fresh specs, that is a
materially stronger result than three retries could ever give. It costs fleet time and buys a real
control, and it is also simply what an operator would see.

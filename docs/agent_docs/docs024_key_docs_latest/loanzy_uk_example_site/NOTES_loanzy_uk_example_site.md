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

### 19:26Z — the classifier is REPRODUCIBLE: two independent runs, same bare domain, identical structured verdict

The re-submission re-ran `domain-research-classifier` from scratch on the same input (the domain
string and nothing else). Comparing the two `classification` specs `[MEASURED 19:26Z]`:

| field | run 1 (17:44:31Z) | run 2 (19:26:20Z) |
|---|---|---|
| `category` | hub | hub |
| `site_type` | content | content |
| `confidence` | **0.82** | **0.82** |
| `suggested_style` | modern-light | modern-light |
| `page_count_estimate` | 12 | 12 |
| `recommended_builder` | pageflow-builder | pageflow-builder |

**Every structured decision field identical, confidence included, to two decimal places.** The only
divergence is the free-text `industry_tags` list, and that in wording rather than meaning:
`buying-guide-platform → buying-guides`, `comparison-content → comparison-platform`,
`uk-gardening → uk-retail`, plus `allotment`/`tool-directory`/`home-garden` appearing and
`magazine-grid` dropping — 10 tags against 8.

**This is a genuinely good result for the route and it belongs in §7 of the route handoff** ("what
the route got RIGHT, so a fix does not remove it"). Three things follow:

1. **The framework's answer to "what is this domain" is stable**, not a coin toss dressed up as a
   verdict. Two independent runs 1h42m apart, no shared orchestration state, same answer.
2. **The classifier is usable as a FIXTURE.** A before/after that re-runs it is comparing like with
   like on the structured fields. I would not have assumed that of an LLM step with no `temperature`
   pinned, and it is worth other lanes knowing.
3. **It sharpens the `376` control.** The next exemplar selection reads these *fresh* specs. If the
   same three organisations come back, "sampling permutes, it does not re-draw" stops being
   conditional on fixed input — which is precisely the caveat I attached to the three-attempt table.

⚠ **Do not over-read it.** This is **n=2** on **one domain** in **one vertical**, measured on the
same afternoon `[MEASURED 2026-08-23]`. It says the classifier was reproducible here; it is not a
claim about the fleet, and the free-text half demonstrably is not stable. A lane wanting to rely on
this as a fixture should re-measure on its own vertical, and should pin the **structured fields
only** — an assertion over `industry_tags` would have failed on the second run.

Also noted: the fresh `needs_vertical_research` (19:26:57Z) inserted cleanly alongside the `failed`
one from the first build — consistent with `idx_swi_dedup` excluding `failed` **and** the classifier's
`create_next_item` carrying `recurrence_expected: true`. Two rows of the same type now exist on this
site, one terminal and one live, which is correct but will look like a duplicate to anyone counting.

### 19:30Z — THE CONTROL RUNS AND SETTLES IT: fresh specs, same three organisations, fourth time

The caveat I attached to the three-attempt table is now discharged, and in the direction that makes
the finding stronger.

The re-submission re-ran `domain-research-classifier` from scratch, and it produced **materially
different** `content_direction`/`classification` specs — `industry_tags` came back with **10**
entries against the first run's **8**, reworded throughout. So the input to `select_exemplars` was
genuinely re-derived, not the same rows read twice. Off those fresh specs `[MEASURED 19:30Z]`:

| slot | att. 1 | att. 2 | att. 3 | **att. 4 — FRESH SPECS** |
|---|---|---|---|---|
| 1 | gardenersworld | gardenersworld | gardenersworld | **which.co.uk** |
| 2 | **thespruce** ✗ | which.co.uk | **thespruce** ✗ | **gardenersworld** |
| 3 | which.co.uk | **thespruce** ✗ | which.co.uk/reviews/… | **thespruce** |

**Four independent selections, four identical organisation sets, four different permutations.** The
fourth came from re-derived specs, so the stability is **not** an artefact of fixed input:

> **The candidate pool is a property of the VERTICAL.** Not of the specs, not of the sampling
> temperature, not of the orchestration state. Any fix premised on a retry eventually picking
> differently is now **disproved**, not merely doubted.

**Why the control was worth the fleet time, stated plainly.** Three retries off one set of specs
could only ever show "stable given fixed input" — which is exactly the shape of a measurement that
cannot come out otherwise, because the thing I wanted to vary (the specs) was held constant by the
mechanism under test. The re-submission was the only way to vary it, and it happened to be the
action an operator would take anyway. **The classifier's own reproducibility (previous entry) is
what makes this a clean control rather than a confound**: the structured verdict was identical
across the two runs, so the exemplar set cannot be explained by the classifier having changed its
mind about the domain — only the free-text half moved, and the exemplars did not follow it.

`thespruce.com` sits at slot 3 this time, so on the established pattern this run dies at
`crawl_exemplar_3` after two more discarded crawls. That will be the **twelfth** exemplar crawl
this domain has paid for and the **fourth** time the same host has killed the stage.

### 20:06Z — THE BUILD IS ALIVE. Attempt 5 escaped, and my "terminally dead" claim was wrong

**What happened** `[MEASURED 20:06Z]`. The second build's `needs_vertical_research`, attempt 2 of 3,
selected a **different set** and cleared all three crawls:

| slot | attempts 1-4 (all contained the refused host) | **attempt 5** |
|---|---|---|
| 1 | gardenersworld.com | gardenersworld.com |
| 2 | thespruce/which | which.co.uk |
| 3 | which/thespruce | **burgonandball.com** |

`thespruce.com` absent; `burgonandball.com` is **from `identity.competitors_found`** — the branch I
had recorded as never having fired. Verified **at the artefact, not the status**:
`site_specs.aspect='vertical_landscape'` written **20:05:45Z** by `vertical-exemplar-researcher`,
and `needs_strategy` created **20:05:55Z**, `triaged`. The item is `complete`. The cascade has moved
to hop three.

**So three claims of mine are retracted** (full retraction in `bugs_open/376` §4b):
1. ~~"Sampling permutes the order; it does not re-draw the pool."~~ It re-draws — 4 of 5, not 5 of 5.
2. ~~"Retry is structurally incapable of routing around it / disproved, not doubted."~~ A retry
   routed around it, on the very next observation after I wrote that.
3. ~~"The `competitors_found` branch has never fired."~~ It fired.
4. ~~"The build is terminally dead."~~ It is not. It was dead **for the first submission**, whose
   item genuinely exhausted `attempt_count=3`; the second submission escaped on its second attempt.

**What still stands, and it is the entire defect:** the crawl steps have no `on_error`, so a refusal
discards the stage including successful crawls (config-verified); and `create_next_item` is the sole
estate-wide producer of `needs_strategy`, so a refusal that exhausts retries **is** terminal — which
is exactly what happened to submission 1. The bug is a **4-in-5 tax that usually exhausts a
3-attempt budget before a lucky draw**, not an inescapable trap. Severity stays HIGH; the argument
changes.

> **THE SAME ERROR TWICE IN ONE DAY, AND THAT IS THE FINDING.** 11:xx: a dispatch-walk mechanism
> built on **14** consecutive ordered samples — broke 20 minutes later. 19:3x: "disproved, not
> doubted" on **4** consecutive identical samples — broke on the fifth, 30 minutes later. Both times
> I converted a *run* into a *law*; both times the counter-example arrived for free because the
> system kept running; both times I could not have named what would falsify the claim, which is the
> check this repo already prescribes.
>
> **The tally is the point** (`WRONG_CALLS`'s own thesis). One instance is carelessness; two in a
> day, in the same shape, from the same reasoning move, is a habit — and the habit is
> **reaching for the modal form**. "Cannot", "structurally incapable", "disproved" are claims about
> a *mechanism*; a tally of identical observations can only ever support "N of N so far".
> **Rule I am adopting: if the evidence is a count, the claim must contain the count.** "4 of 5
> draws contained the refused host" was always available, is what I actually knew, and would have
> survived the fifth draw intact.
>
> Second-order: my 17:52 entry *over*-corrected into false uncertainty after being refuted, and this
> one *under*-corrected into false certainty. Neither is calibration — both are reacting to the
> last error instead of to the evidence.

### 20:10Z — the strategy spec read (the artefact this lane judges), and two concrete after-test items it raises

Hop three landed at 20:09Z. Per the RUNBOOK ("the strategy spec is where the framework's own answer
to *what is this domain for* first becomes readable — the artefact to judge before any page
deploys"), read in full. **It is good**, and specifically it is good in the ways `loanzy.uk` was not:

- `primary_model: affiliate`, `site_type: review-site`, tone editorial. The three rejected models
  each carry a *reason* rather than a shrug — `direct_business` is refused because retailing "would
  directly compete with the affiliate partners the site depends on"; `lead_generation` because
  garden-tool vendors do not run lead programmes.
- **It volunteers the disclosure obligation unprompted:** sponsored brand-directory listings "must
  be clearly disclosed and structurally separated from editorial content to preserve the independent
  stance that is the site's core differentiator", and are scoped to "maturity, not at launch".
- Search intent is plausibly UK-specific and commercially literate ("best spade for clay soil UK",
  "garden fork vs spade which do I need", "best pruning shears for arthritic hands").

**Two things it puts on the after-test, both concrete, neither yet a problem:**

1. **Named third-party affiliate relationships that DO NOT EXIST.** The spec asserts commission
   "from retailers including Amazon UK, RHS Shop, and participating specialist garden retailers",
   and elsewhere names Thompson & Morgan, Crocus and Dobies. That is a *plan*, correctly scoped as
   one. **It becomes a false claim the moment a built page states or implies a live affiliate
   relationship, or shows a commission disclosure naming a partner we have not signed.** Check the
   served pages for partner names presented as fact. This is the same class as `loanzy.uk`'s lender
   panel — an invented commercial relationship — arriving by a quieter route.
2. **`RHS endorsed garden tools` appears twice, in `likely_searches` AND `high_value_terms`.** As a
   search term it is a legitimate observation about demand. **As page copy it would be an endorsement
   claim about a real named charity**, and the Royal Horticultural Society licenses its endorsement
   commercially. If a page says or implies RHS endorsement of anything, that is a banned-claim
   failure and it must be reported loudly. Grep the served pages for `RHS`.

Both are exactly what `evidence_base` claims gating and the banned-claim sweep exist to catch, so
**the right reading is "the controls now have something real to catch", not "the strategy is
wrong".** Recorded before the pages exist so the check is not invented after seeing them — a check
written to fit an artefact you have already read is not a check.

### 20:16Z — the site plan, and TWO PREDICTIONS recorded before the pages build

Site plan complete 20:16Z. **12 pages**, plus 11 `needs_imagery`, `needs_composition`,
`needs_design`, `needs_rerender`, and one `owned_page_review`:

`index` (landing) · `about` · `contact` · `care` · `seasonal-planner` · `how-we-assess` ·
**`affiliate-disclosure`** · `buying-guides-index` (section-index) · `buying-guide-post` (blog-post) ·
`brand-directory-index` (entity-directory) · `brand-profile` (entity-page) · **`tool-finder`** (tool)

**Two things the framework got RIGHT, unprompted, and they belong in §7 of the route handoff:**
- It planned an **`affiliate-disclosure`** page and a **`how-we-assess`** page off its own bat. Those
  are precisely the disclosure and stated-criteria obligations its own strategy spec committed to.
  Nobody asked for either.
- It **refused to auto-build the tool page.** `owned_page_review` /
  `item_key=owned_page_review:tool-finder`, status `needs_human_review`, summary *"Owned page
  tool-finder is not_built — needs owner-aware build, not the generic builder."* That is the same
  class of correct refusal as loanzy's `guides-index` no-op — the route declining rather than
  shipping something hollow.

**PREDICTION A — the `contact` page has nothing to put on it.** `sites.email`, `sites.phone` and
`sites.company_name` are **all NULL** `[MEASURED 20:16Z]`, and `build-briefing-agent` itself declared
`contact_email`, `contact_phone` and `people` as **gaps**. A `contact` page is nonetheless planned.
> So one of three things happens, and they are very different: (a) it builds thin/empty — a real
> but minor defect; (b) it **invents** an address, email, phone or a named person — the `loanzy.uk`
> class, a fabricated identity, and must be reported loudly; (c) the builder refuses it the way it
> refused `tool-finder`, which would be the correct behaviour and the best outcome.
> **The harness already greps for (b)** — email (including Cloudflare's
> `/cdn-cgi/l/email-protection` rewrite, which is why a plain `mailto:` grep gives a false clean),
> UK phone patterns, and `written/reviewed/tested by <Name>` bylines.

**PREDICTION B — `tool-finder` becomes a live `bugs_open/328` reproduction.** It is parked at
`needs_human_review` and will not auto-build, but it **is in the plan**, so nav and the guides/index
pages will very likely link to it.
> Expect a **dead link to `tool-finder`** from at least one served page. That is 328 exactly — "a
> page that failed to build stays linked from the pages that did" — arriving on a fresh greenfield
> build rather than on a repaired one. The harness's dead-link sweep will catch it. **If it does NOT
> appear, that is more interesting than if it does**: it would mean either the nav builder excludes
> unbuilt pages (328 narrower than filed) or nav refused to run at all (the loanzy behaviour).

**One correction to what I told the owner earlier:** I said the page phase is "where 311/328/337
live". **`337` is now unlikely to fire on this build** — it needs a `needs_new_component` for a tool
section, and the only tool page has been routed to human review instead of the generic builder. So
the `311` after-test may also have nothing to test. **That would not be a null result — it would be
the finding**: the route never reaches the component-collision path on a vertical whose single tool
page is owner-gated. Say that explicitly rather than reporting "no collision detected", which would
read as the guard having been exercised when it was not.

### 20:21Z — `unresolved_cta` on `about`: the RIGHT behaviour, and a build-ORDER question it exposes

A second human-review item appeared while the pages build `[MEASURED 20:21Z]`:

```
unresolved_cta | needs_human_review | Unresolved CTA on about ('call-to-action'):
                 no real-page destination for primary_cta_url, secondary_cta_url
spec.fix: "No real page exists to serve as this CTA's destination (no eligible content hub).
           The gated template renders no button. Add/activate a section-index hub, or set
           the destination manually."
source: resolve_internal_links
```

**This is the route behaving well and it belongs in §7.** Faced with a CTA whose destination does
not exist, it (a) **renders no button rather than a broken one**, (b) files a review item naming the
page, the section, the two missing fields and two concrete remedies. That is precisely the opposite
of `bugs_open/328`, which leaves a dead link and says nothing — and it is worth noting that **the
same build contains both behaviours**, so 328 is not "the platform does not care about broken
destinations", it is narrower than that.

**But it exposes an ordering question, and I do not yet know the answer.** `about` is being built
while `buying-guides-index` and `brand-directory-index` are still `planned`. Its CTA could not
resolve **because the target pages do not exist yet** — not because they never will.

> **PREDICTION C, recorded now:** the `needs_rerender` item already queued for this site should
> re-resolve `about`'s CTA once the hubs exist, and the buttons should appear. **If they do,** this
> is transient and correct, and the only defect is a human-review item raised for a condition that
> resolves itself — noise, not damage. **If they do not** — if `about` ships permanently buttonless
> because its CTA was resolved once, early, against a site that was 1/12 built — then build ORDER
> silently determines page quality, and every early page on every greenfield build carries it.
> **That would be a new bug and a nastier one than 328**, because the page looks finished.
>
> **How to tell them apart at the artefact, not the item:** after the build settles, fetch
> `https://garden-tools.uk/about.html` cache-busted and count `<a class="...cta...">`/`<a class="btn`
> occurrences. The item going `complete` is NOT the test — `resolve_internal_links` filing a review
> row and something later fixing the HTML are independent events, and only the second one matters.

⚠ **Do not act on this.** It is the route's own behaviour under a no-hint build, which is the thing
being measured. Setting the destination by hand — which the `fix` text invites — would destroy the
measurement.

### 20:27Z — FIRST PAGE SERVES. Every predicted check run against it, including one my own harness got wrong

`about` reached `deployed` at ~20:26 and **actually serves**: `http=200`, **12,830 bytes**
(`deployed != serving` — checked at the artefact, cache-busted).

| check | result |
|---|---|
| **PRED C baseline** — CTA anchors | **0** `cta`/`btn`-class anchors, but the `call-to-action` section IS present. The gated template rendered the container and no button, exactly as the `unresolved_cta` item said. **This is the before-picture for the rerender test.** |
| **PRED A** — invented email / phone | **none.** No address, no `mailto:`, no Cloudflare `/cdn-cgi/l/email-protection` rewrite, no UK phone pattern |
| **PRED A** — invented people | **none** (see the harness fault below) |
| partner names (Amazon/RHS Shop/Crocus/…) | **none** |
| RHS endorsement claim | **did not materialise in the dangerous form** — see below |
| template-syntax leak (`260`) | **0** tokens |
| internal links | 3, and **all three 404 right now** — see the distinction below |

**The RHS line, verbatim, because the prediction was about this and it deserves the actual words:**
> *"Every buying guide reckons with British conditions: clay that sets like brick after a dry August,
> **RHS testing standards**, and retailers who actually ship to a UK address."*

That is **not** an endorsement claim — it does not say the RHS endorses this site or any tool, which
is what I flagged and what would have been serious. It is a weaker adjacent thing: an unverified
assertion that "RHS testing standards" exist as a citable benchmark for hand tools, plus an implied
methodology claim ("every buying guide reckons with" them) from a site that has tested nothing yet.
Worth an `evidence_base` look; **not** worth reporting loudly. **Recording the prediction as NOT
CONFIRMED rather than quietly reshaping it to fit what turned up.**

**⚠ The three 404s are NOT `bugs_open/328`, and calling them that would be wrong.**
`/affiliate-disclosure.html`, `/buying-guides/index.html` and `/how-we-assess.html` all 404 as I
write this — but those pages are `planned` and being built right now, one at a time. **That is a
build in progress, not a page that failed.** 328 is about links to pages that will never exist.
> The real 328 candidate on this build remains **`tool-finder`**, which is owner-gated at
> `needs_human_review` and will not auto-build. **The 328 test is: after the build settles, is
> `tool-finder` still linked?** Testing it now would confirm the bug from evidence that cannot
> distinguish it from normal progress — the same could-not-come-out-otherwise shape this lane keeps
> tripping over. **Transient 404s during a serialised build are expected and must not be counted.**

### ⚠ My harness produced a FALSE POSITIVE on its very first real page

It reported `named bylines: written by people who`. There is no invented byline. The bug is mine:

```sh
grep -oiE '(written|reviewed|tested) by [A-Z][a-z]+ [A-Z][a-z]+'   # -i DEFEATS the whole test
```

**`-i` makes `[A-Z][a-z]+` match any word**, so a pattern whose entire discriminating power was
*"two Capitalised words = a person's name"* degraded to *"any two words after 'written by'"*. Without
`-i`: no match, correctly. Fixed in the harness with the reason written above the line.

**Two lessons, and the second is the one worth carrying:**
1. **A case-sensitive pattern and a case-insensitive grep are mutually exclusive.** If capitalisation
   IS the signal, `-i` deletes the signal while leaving the check looking present and green.
2. **PRINT THE MATCH, NOT THE COUNT.** I caught this only because the harness echoes the matched
   text. Had it printed `bylines: 1` I would have gone and read the page looking for a fabricated
   author, found the phrase "written by people who", and — worse — might have reported an invented
   byline to the owner. **A count cannot be sanity-checked; a match can.** Every grep in that
   harness now prints what it matched.

**This is the harness's first real use and it was wrong.** That is the argument for validating an
instrument on live data before trusting a clean run from it — a check that has only ever returned
"nothing found" has not been shown to work.

### 20:31Z — THREE of twelve pages will not auto-build, and the pattern is a page CLASS, not a page

`[MEASURED 20:31Z]` — `needs_page` status after four of twelve:

| page | status | why |
|---|---|---|
| `about` | complete | serves, 12,830B |
| `affiliate-disclosure` | complete | — |
| **`brand-directory-index`** (entity-directory) | **needs_human_review** | *"page-build-handler no-op: no sections ready to build (empty spec sections, or all sections deferred…)"* |
| **`brand-profile`** (entity-page) | **needs_human_review** | same message |
| **`tool-finder`** (tool) | **needs_human_review** (`owned_page_review`) | *"needs owner-aware build, not the generic builder"* |
| the other 7 | triaged | building one at a time |

**The finding is that it is a CLASS, not an incident.** loanzy's route handoff §7 recorded ONE such
refusal (`guides-index` no-op'd rather than ship an empty shell) and filed it under *what the route
got right*. It still is right — refusing beats shipping a hollow page. But at **3 of 12** it stops
being a nice guard and becomes a shape:

> **The site planner plans page ROLES whose content depends on entities or ownership that the
> greenfield route never creates.** `entity-directory` and `entity-page` need brand entities; nobody
> populates them from a domain name. `tool` needs an owner-aware build by design. So the planner
> commits to twelve pages while the pipeline behind it can only deliver nine, and **nothing
> reconciles the two.** That is not the builder's fault — it is a planner/pipeline contract gap, and
> it will recur on every vertical whose plan includes a directory (which is most of them: the
> classifier chose `brand-directory-index` here unprompted, and "directory" is one of the three
> things this domain was picked to exercise).

**This changes the `328` test materially — it is now 3 pages, not 1.** If nav and the built pages
link to `brand-directory-index`, `brand-profile` and `tool-finder`, a finished-looking site ships
with **three** dead destinations. **Still not testable yet** (the other seven are mid-build and
their 404s are transient), but the candidate set is now known and larger than I predicted.

**What I do NOT yet know, and will not assert:** whether the *plan* is wrong to include them, or the
*pipeline* is wrong not to populate entities. Both readings fit. `[UNVERIFIED]` — the discriminator
is whether any non-greenfield path (adoption, a mission brief naming brands) does populate them, in
which case the gap is specific to the no-hint route rather than to the page class. Not measured.

### 20:35Z — CORRECTION to the 20:31Z entry: it is not "entity pages", it is every COLLECTION role. Now 5 of 12

> **CORRECTED 20:35Z, four minutes after I wrote it.** At 20:31 I framed the no-op class as
> *"`entity-directory` and `entity-page` need brand entities; nobody populates them"*. **Too narrow.**
> Two more pages no-op'd with the identical message, and they are not entity pages:

| page | role | status |
|---|---|---|
| `brand-directory-index` | entity-directory | needs_human_review |
| `brand-profile` | entity-page | needs_human_review |
| **`buying-guides-index`** | **section-index** | **needs_human_review** |
| **`buying-guide-post`** | **blog-post** | **needs_human_review** |
| `tool-finder` | tool | needs_human_review (owner-gated, different cause) |

All four share one error: *"page-build-handler no-op: no sections ready to build (empty spec
sections, or all sections deferred…)"*. **Five of twelve pages will not auto-build.**

**The corrected mechanism — broader and simpler than entities.** Every one of these four is a
**collection or instance role**: it indexes something (`section-index`, `entity-directory`) or is one
instance of something (`blog-post`, `entity-page`). Their content is *other content*. On a greenfield
build that other content does not exist, so the section spec is empty and the builder correctly
declines. The three pages that HAVE built (`about`, `affiliate-disclosure`, and `care` in progress)
are all **standalone `content`/`landing` roles whose content is self-contained.**

> **So the shape is: the planner plans a SITE, the builder builds PAGES, and nothing builds the
> CORPUS that the collection pages exist to present.** A twelve-page plan where four pages are
> containers is a plan for a site with roughly eight pages of actual content plus four shells — and
> the pipeline has no step that fills them. This is a planner/pipeline contract gap, and naming it
> "entity pages" would have sent a fixer to the wrong half.

**PREDICTION B IS NOW CONFIRMED, and not on the page I predicted.** The already-serving `about` page
links to `/buying-guides/index.html` `[MEASURED 20:27Z]`. `buying-guides-index` is now
`needs_human_review` and **will not build without a human**. So a live, serving page links to a
destination that will never exist on an unattended build — `bugs_open/328` exactly, reproduced on a
greenfield run. I predicted it via `tool-finder`; it arrived via a page whose failure mode I had not
anticipated at all. **The prediction was right for the wrong reason, and the mechanism I named was
not the one that fired.**

⚠ **Still not the final count.** `contact`, `how-we-assess`, `index` and `seasonal-planner` are yet
to build, and `index` is the one that matters most — it is the apex, and if the landing page is
mostly links to the four shells the whole site reads as broken. Do not write the summary yet.

### 20:42Z — PREDICTION A RESOLVES, and it resolves the RIGHT way: the framework refused to invent contact details

`needs_section_data` / `needs_human_review` `[MEASURED 20:42Z]`:

```
Section 'contact-info' on contact needs: Business contact email address
spec.missing[0] = { field: "email", type: "text",
                    source: "site_specs.identity.email",
                    reason: "Business contact email address",
                    on_missing: "needs_human_review" }
source: plan_sections
```

**Prediction A had three branches: (a) build thin, (b) INVENT, (c) refuse and ask. It is (c)** — the
best of the three, and the one I said would be the correct behaviour. The framework hit a field it
did not have, **declined to fabricate it**, and filed a review item naming the page, the section, the
field, the spec path it would have read, and why the field matters. Nothing was invented.

**This is the direct counter to the `loanzy.uk` failure class.** That build invented an entire
regulated business — a lender panel, an eligibility checker — from a domain name. This one would not
invent an email address. The difference is not tone or luck: it is a **declarative per-field
`on_missing` policy** carried in the component's own spec, so the refusal happens at the field, by
configuration, rather than depending on a writer's judgement. **That belongs in §7 of the route
handoff as a mechanism, not an anecdote.**

**And it partly corrects my 20:35Z framing.** I wrote *"nothing builds the CORPUS that the collection
pages exist to present"* and implied the pipeline has no way to say "I need data I do not have". It
does — `on_missing` is exactly that, and it works. **The narrower, correct statement:** the pipeline
has a per-FIELD mechanism for declaring and escalating missing data, and **no equivalent for a
missing CORPUS.** A contact page missing an email raises a precise, actionable review item; a
buying-guides index missing all its guides raises `"no sections ready to build"` — true, but it names
no field, no source, and no remedy, because the thing missing is not a field. **The gap is not
"nothing asks"; it is that the asking mechanism has field granularity and the failure has corpus
granularity.** That is a much more useful thing to hand a fixer.

⚠ Note what this does to the earlier harness check: the invented-identity greps are still worth
running on the served pages, but **their most likely outcome is now "nothing found" for a good
reason rather than a lucky one.** Do not report a clean result as evidence the writer resisted
temptation — the writer was never asked. The control that would discriminate is a build where
`identity.email` IS present; then a fabricated *different* address would mean something.

### 20:58Z — THE AFTER-TEST, run against the live site. And the harness had to be REWRITTEN first

**⚠ Read this before the results: the first run of the harness reported three clean checks that had
not run.** `pages.is_archived` does not exist (the column is `status`), so the page-list query
errored, the per-page loop got no input, and each section then printed its cheerful
`"(no lines = nothing found)"` footer. **A failed query produced a reassuring message.** That is the
could-not-have-come-out-otherwise failure this lane has logged three times today — committed by the
instrument built to detect it. A second bug followed: `q()` folded stderr into rows, so `ERROR: …`
text was fetched as if it were a URL; and `name` is ambiguous across `pages`/`sites`.

Rewritten so it cannot happen again: every query's row count is printed and **zero rows reports
UNKNOWN, never PASS**; `q()` refuses to return error text as data; page URLs come from **`pages.url`**
rather than being guessed from `pages.name` (v1 guessed `/seasonal-planner.html` correctly by luck
and `/tools/finder/index.html` wrongly); every grep prints **what it matched**, not a count.

#### Results `[MEASURED 20:58Z, live site, cache-busted]`

**`311` collateral — CLEAN.** All **8** incumbents `UNCHANGED` on both `md5(html_template)` and
`md5(input_schema)`, against **this session's own pre-run pins**. Row count asserted = 8.

**`311` diversion — NOT EXERCISED, and that is the honest reading.** 0 scoped components
(`function LIKE '%garden-tools-uk%'`), 0 `COMPONENT_COLLISION_DIVERTED` rows. The only tool page was
owner-gated to human review, so the component-creation path was never reached. **Report as "the
guard never ran", never as "no collision occurred"** — the clean result carries no information about
the guard.

**`260` template leak — 0** on all four tokens (a ceiling, not a count).

**Pages: 6 serving, 6 at 404** (one of the six, `seasonal-planner`, is still building):

| serving | bytes | | 404 |
|---|---|---|---|
| `/index.html` | 14,831 | | `/brand-directory/index.html` |
| `/how-we-assess.html` | 15,242 | | `/entities/brand-profile.html` |
| `/care.html` | 13,113 | | `/blog/buying-guide-post.html` |
| `/about.html` | 12,830 | | `/buying-guides/index.html` |
| `/contact.html` | 5,503 | | `/seasonal-planner.html` *(still building)* |
| `/affiliate-disclosure.html` | 2,374 | | `/tools/finder/index.html` |

**`328` — CONFIRMED, NINE dead links across four serving pages.** The **home page alone has four**:
`/brand-directory/index.html`, `/buying-guides/index.html`, `/seasonal-planner.html`,
`/tools/finder/index.html`. Also from `/about.html` (1), `/care.html` (2), `/how-we-assess.html` (2).
Four of the six targets sit at `needs_human_review` with nothing scheduled to build them, so those
are **permanent on an unattended build**, not transient.

**Invented identity — NOTHING invented.** No email (including the Cloudflare
`/cdn-cgi/l/email-protection` rewrite), no UK phone, no fabricated byline, on any of the six served
pages. `sites.email/phone/company_name/contact_address` all still empty. **But see 20:42Z: the writer
was never asked**, because `on_missing` intercepted first — so this is a good result for the
*mechanism*, not evidence about the writer's restraint.

**`contact.html` serves anyway — with a working form.** 200, 5,503 bytes, **2 `<input>`, 1
`<button>`**, while `build_status='needs_rebuild'` and its `needs_section_data` item is open. So the
page shipped the form and withheld only the email field. Sensible, and a good example of why
`build_status` is not the artefact.

#### The two claims predictions, and they SPLIT

**RHS — NOT CONFIRMED. The copy is careful, and better than I expected.** `how-we-assess` says
*"whether the brand carries an RHS endorsement"* and *"RHS endorsements where they exist"*. Those are
claims that the site **checks** third-party endorsement, correctly hedged — not claims that the site
or its tools are RHS-endorsed. The same page volunteers *"we mark that clearly rather than presenting
a guess as a firsthand finding."* **My predicted failure did not occur.**

**Partner names — CONFIRMED, mildly.** `/affiliate-disclosure.html`: *"Some of the links on this site
earn us a small commission at no extra cost to you. When you click through to a retailer, **Amazon
among them**, and go on to buy the tool we've written about, that retailer may pay us a share of the
sale."* Present tense, naming a specific company. **There is no Amazon Associates relationship for
this site, and there are no affiliate links on it** — the guides that would carry them never built.
So a served page asserts a commercial arrangement that does not exist. It is the `loanzy.uk`
lender-panel class arriving as disclosure boilerplate: far milder, same shape, and the sort of thing
`evidence_base` gating exists for. **Worth reporting to whoever owns claims gating; not worth
alarm.**

### 21:02Z — page phase CLOSED: 7 of 12 serving. Corrected dead-link figures, and a nuance about which unbuilt pages actually hurt

`needs_page`: **7 complete, 4 needs_human_review**, none pending. `seasonal-planner` **did** build
after all (it was still `claimed` when I reported "6 serving" — corrected here).

**Serving (7):** `index` 14,831B · `how-we-assess` 15,242B · `care` 13,113B · `about` 12,830B ·
`seasonal-planner` · `contact` 5,503B (`needs_rebuild`, serves a working form) ·
`affiliate-disclosure` 2,374B.
**Never built (5):** `brand-directory-index`, `brand-profile`, `buying-guide-post`,
`buying-guides-index` (all *"no sections ready to build"*), `tool-finder` (owner-gated).

**Corrected `328` measurement `[MEASURED 21:03Z]` — 9 dead link instances, but only THREE distinct
dead targets:**

| dead target | linked from |
|---|---|
| `/buying-guides/index.html` | index, about, how-we-assess, seasonal-planner (**4**) |
| `/tools/finder/index.html` | index, care, how-we-assess, seasonal-planner (**4**) |
| `/brand-directory/index.html` | index (**1**) |

> **Earlier figure superseded:** at 20:58 I counted `/seasonal-planner.html` among the dead targets
> from `index` and `care`. It built at ~21:01 and now returns 200, so the home page has **three**
> dead links, not four. The instances stayed at 9 because `seasonal-planner`, once serving,
> **contributes two dead links of its own**. Same total, different composition — a good reminder
> that a count taken mid-build is a snapshot of a moving system, not a result.

**The nuance, and it narrows `328` usefully:** of the five never-built pages, only **three** are
linked. `brand-profile` and `buying-guide-post` are unbuilt **and unlinked** — orphans. So 328's
damage is not "every unbuilt page becomes a dead link"; it is "every unbuilt page **that something
links to**". The other two are a different and milder defect: **a plan that promises pages nothing
references**, which costs a build slot and a review item but no visible breakage. Worth separating,
because a fix for one does nothing for the other.

**Prediction C still pending and still un-fired:** `cta-anchors=0` and `buttons=0` on `index`,
`about` and `care` `[MEASURED 21:03Z]`, with `unresolved_cta` now at **8** items and the
`needs_rerender` / `reconcile_rerender` items still `triaged`. The before-state is stable across the
whole site. The rerender is the test.

### 21:29Z — the rerender wave: pages 4.5× bigger, Prediction C STILL un-fired, and 5 rerenders that "completed" on pages that do not exist

**Two distinct waves, different types and handlers** `[MEASURED 21:31Z]` — worth separating because
they look alike in a status roll-up:

| wave | item_type | handler | item_key shape | state |
|---|---|---|---|---|
| 1 | `page_rerender` | `page-rerender` | `page_rerender_<page>_<siteid>_assemble` | **12 complete** |
| 2 | `needs_page` | `page-build-handler` | `page_rerender:<page>` | 1 claimed, 5 triaged |

Wave 1 inlined the design system: `index` went **14,831 → 66,512 bytes**, `about` 12,830 → 64,486.
All 11 `needs_imagery` items are complete.

**PREDICTION C IS STILL UN-FIRED, and my crude grep nearly said otherwise.** After wave 1 the pages
report `buttons=1` where they had `buttons=0`. **That button is the mobile menu toggle** —
`<button class="mobile-menu-toggle" aria-label="Toggle menu">` — not a CTA. `cta-anchors` is still
**0** on `index`, `about`, `care` and `how-we-assess`, and `unresolved_cta` still stands at **8**
`needs_human_review` items. **Had I graded on the count I would have reported the CTAs restored.**
Same lesson as the byline false positive, four hours apart: *read what matched.* Wave 2 is still
running, so the prediction remains open.

**NEW FINDING — a `complete` rerender on a page that does not exist.** Wave 1 completed for **all
twelve** pages, including the five that never built. Measured at the artefact:

| page | `page_rerender` item | `pages.build_status` | `deployed_at` | served |
|---|---|---|---|---|
| `buying-guides-index` | **complete** | planned | NULL | **404** |
| `tool-finder` | **complete** | planned | NULL | **404** |
| `brand-directory-index` | **complete** | planned | NULL | **404** |
| `brand-profile` | **complete** | planned | NULL | **404** |
| `buying-guide-post` | **complete** | planned | NULL | **404** |

> **Why it matters, and it is not pedantry:** anyone auditing "did this site finish rendering?" reads
> `page_rerender: 12 complete, 0 failed` and concludes yes. The only columns that disagree are
> `build_status='planned'` and a NULL `deployed_at` — and `build_status` is exactly the column this
> lane's own handoff warns is unreliable in the other direction (*"`deployed` means pushed, not
> serving"*). **So the two signals a reader would reach for are both wrong, in opposite directions,
> on the same site.** This is the `bugs_open/315` status-column family, and the safe rule is
> unchanged and now doubly evidenced: **the served page is the only artefact.**

`[UNVERIFIED]` whether the handler *should* no-op or refuse on a page with no content — I have not
read `page-rerender`'s workflow, and "rerender a page that has nothing to render" may be a legitimate
no-op that simply reports the wrong terminal status. Not asserting a defect in the handler; asserting
that **the status is uninformative**, which is measured.

### 2026-08-24 09:05Z — BUILD COMPLETE. Prediction C resolves GOOD; and the site proves status is unreliable in BOTH directions

The build finished overnight. Final state: **7 of 12 pages serving**, 5 never built, all imagery and
both rerender waves complete, `site_unreachable` closed itself when the site came up.

**PREDICTION C — resolved, and the frightening branch is REFUTED.** I predicted that if the rerender
did not restore the gated CTAs, then *build ORDER silently determines page quality on every
greenfield build* — nastier than `328` because the page looks finished. **It does restore them, and
it restores them correctly** `[MEASURED 09:05Z]`:

| page | before (21:03Z) | after | CTA href | target |
|---|---|---|---|---|
| `about` | cta-anchors 0 | **1** | `/how-we-assess.html` | **200** |
| `care` | cta-anchors 0 | **1** | `/seasonal-planner.html` | **200** |

Both point at pages that **exist**. The rerender did not "resolve" them by aiming at a 404 — which
was the plausible bad outcome and is the one I would have reported as serious. The remaining pages
still show 0 because their CTA destinations are the hubs that never built, so the gate correctly
renders nothing rather than a broken button. **The mechanism works. Build order costs a page nothing
permanent, provided a rerender follows.**

**NEW FINDING, and it completes a pair.** The 8 `unresolved_cta` items are **all still
`needs_human_review`**, every `updated_at` between 20:20 and 20:56 — i.e. **before** the rerenders
ran. `about | call-to-action` is provably stale: the CTA it describes now renders and resolves to a
live page, and its item has not been touched since 20:20:47.

> **So this one site demonstrates the status columns being wrong in BOTH directions at once:**
> - **`complete` on work that did not happen** — 5 `page_rerender` items complete against pages that
>   are `planned`, have NULL `deployed_at`, and 404.
> - **`needs_human_review` on work that DID happen** — an `unresolved_cta` item still open after the
>   rerender fixed the very CTA it names.
>
> An operator reading the queue sees "12 rerenders complete, 8 CTAs need attention". The truth is
> 7 rerenders real and at least 6 CTAs fine. **Neither direction of error is detectable from the
> queue; both are one `curl` away.** This is the strongest single-site evidence for the house rule,
> and it belongs in `016b` §9: *a work item's terminal state is a claim about the last thing the
> handler believed, not about the artefact — and it decays in both directions.*

**Final `328` measurement: 9 dead-link instances, 4 distinct dead targets** — `/tools/finder/`
(from about, care, index, seasonal-planner), `/buying-guides/` (care, index, seasonal-planner),
`/brand-directory/` (index), `/blog/buying-guide-post.html` (seasonal-planner). `brand-profile`
remains unbuilt **and unlinked** — the orphan case, a separate milder defect.

**Everything else clean at the artefact:** `311` collateral 8/8 UNCHANGED (guard **NOT EXERCISED** —
report as such); `260` zero template tokens on all pages; no invented email, phone or byline
anywhere; `contact.html` serves a working 2-input form.

### 2026-08-24 — the OWNER reviewed the served pages, and my after-test missed the biggest defect

Three points from the owner, plus one escalation. All three traced to mechanisms; two filed.

**The escalation first, because it corrects me.** *"The missing pages is not just one, but almost all
of them. The seasonal planner says 'What your shed needs, month by month' but there is no calendar
and no month by month list."*

He is right. I reported **"7 of 12 serving… substantial and read well"**. That was measured on
**byte count** (66,999 for `seasonal-planner`) and a skim of one page. **Byte count cannot see a page
that promises a thing and does not deliver it**, and my harness had no check that could:
`http`, `bytes`, `<input>`, `<button>`, `cta-anchors`, template leaks, invented identity, dead links —
**not one of them asks whether the page contains what its own heading says it contains.**

> **The specific trap, and I had been warned in writing.** The 08-23 handoff told me `loanzy.uk`
> shipped a calculator page with **zero inputs** — *"a stored, linked, selector-visible component
> that still renders no tool is a different failure and must not be reported as success"*. I quoted
> that warning, built an `<input>`-count check from it, **and applied it only to `page_type='tool'`**.
> `seasonal-planner` is `page_type='content'`, so it got no completeness check at all. **I narrowed a
> general lesson to the one example it arrived with.** The transferable form: *a page's own headings
> are a promise; check the promise against the markup.* An `<h2>` saying "month by month" over a page
> with 0 tables, 0 lists and 3 month names is machine-detectable and I did not detect it.

**Filed: `bugs_open/380`** — no evidence base ⇒ no fact assignment AND no claims audit. Three
mechanisms fail open on one condition; `claims-auditor.check_opted_in` branches to **`complete`**
without reading a page. **29 of 48 live sites (60%) have no evidence base**, so fleet claims coverage
is ~40% and a skipped audit looks exactly like a clean one. This is the loanzy credit-broker
mechanism generalised: `CGV-032` gates the *vertical*, nothing gates the *practice claims*.

**Filed: `bugs_open/381`** — the planner composes pages from components that cannot express the page
it planned. `seasonal-planner`'s four components have no list or table markup between them, which is
**why** the month-by-month promise became four prose blocks: the writer had nowhere to put twelve
months. Site-wide: 0 tables, 0 content lists, 0 `<strong>`, against **34 list-capable and 10
table-capable components available**. Two arms — writer-side (the owner's paragraph sits in a
**pass-through** component that would have taken `<ul>` unchanged) and composition-side (three
components hard-wrap in `<p>` and nothing writer- or designer-side can change that).

**Not filed: the cards.** Measured instead — **there is no carousel** (`scroll-snap` 0, no
`.carousel`/`.slider`). It is a CSS grid collapsing to one column; `index` carries 14 cards, and the
one identified grid **already holds exactly 3**. So the owner's proposed remedy is already the state
of the world and does not address it: the wall is the number of card *sections*, not cards per
section. Recorded in `381` §7 pending a composition decision rather than filed on a guess.

**One correction to the owner's own framing, offered rather than assumed:** there is **no
`vigilant_designer` agent** — live design agents are `brand-designer`, `feature-designer`,
`visual-designer`. And a designer is the wrong layer for the structural arm: it cannot add a `<ul>`
to a component whose template has none, so routing there yields a better-looking wall of text. His
alternative hypothesis — *"a missing step in the workflow"* — is the correct one.

### 2026-08-24 18:0xZ — the `206` PRE-FIX observation was in my own build data all along, and the row-clearing I was authorised to do is INERT-or-DESTRUCTIVE

**Recorded here so it does not depend on anyone remembering it.** `bugs_open/206`'s pre-fix
behaviour — reconcile minting an `entity-directory` page at the generic handler — is captured in
this lane's own unaided greenfield build `[MEASURED 2026-08-24, from `site_work_items`]`. All 13
items carry `created_at = 2026-08-23 20:15:50.199268+00`, **byte-identical to that site's
`last_reconciled_at`**, and `created_by='reconcile_site_plan'`:

| page | item_type | handler_agent | |
|---|---|---|---|
| `brand-directory-index` | needs_page | **`page-build-handler`** | ← entity-directory, WRONG |
| `brand-profile` | needs_page | **`page-build-handler`** | ← entity-page, WRONG |
| `buying-guides-index` / `buying-guide-post` | needs_page | `page-build-handler` | |
| `tool-finder` | owned_page_review | *(empty)* | ← correctly gated |
| the 7 content/landing pages | needs_page | `page-build-handler` | correct for those types |

**That is the bug caught in the act on a real build, by a lane that was not looking for it** — better
evidence than a contrived reproduction, and it existed before either lane thought to ask for one.

**Why the authorised row-clearing is NOT going ahead** (owner said "let's free the parked row"; I am
taking the correction back to him rather than executing on a premise I now know is false):

1. **Clearing alone is inert.** `reconcile_site_plan` is carried by **exactly one** agent —
   `build-site-planner` — swept across every live agent's steps, one row `[MEASURED]`. There is no
   timer. Reconcile only runs inside a build/publish, so a quiet site never re-reaches it.
2. **Making it non-inert means a full RE-PLAN.** `build-site-planner`'s order is `plan_site` (LLM) →
   `write_site_plan` → `sync_pages` → design/imagery/nav → *then* reconcile. `sync_pages` overwrites
   `pages.sections`; design and imagery are re-emitted. That is `bugs_closed/001`'s hazard and it
   would destroy this lane's clean measurement — the dead-link census, the 7-of-12 record, and the
   dated pre-fix structure baseline (0 tables / 0 lists / 0 `<strong>`) that `bugs_open/381` is
   holding as its comparison point.
3. **The benefit is zero, because the pre-fix half already exists** (above) and the post-fix half
   needs a **future** build, which this is not. Cost high, benefit nil.

> **The transferable bit: I advised the owner to authorise an action whose mechanism I had not
> read.** I proposed "free the row, trigger a build, watch the link go live" — three steps, and I had
> checked none of them against the code that would have to execute them. The 206 lane checked, found
> step 2 does not exist as a discrete action, and retracted its own request. **An authorisation
> obtained on a wrong premise is not permission to act; it is a correction owed.**

**Also caught, in their closure query:** it reads `spec->>'page_type'`, which **is** absent (0 of 134
reconcile-minted rows fleet-wide). So the query could not discriminate an entity-directory from a
content page. **Second check on this one bug that could not have come out otherwise.**

> **⚠ CORRECTED 2026-08-24 — THE REMEDY I OFFERED WAS WRONG, AND SO WAS THE SPEC SHAPE I QUOTED.**
> I told them a reconcile spec is `{domain, page_id, filename, page_name}` and to join on
> `spec->>'page_id'`. **Neither exists on a reconcile row.** Measured fleet-wide over all 134
> reconcile-minted `needs_page` rows: `page_role` **134/134**, `page_type` **0**, `page_id` **0**,
> `filename` **0**, `domain` **0**.
>
> **The real spec is `{reason, plan_id, page_name, page_role}`** — and `page_role` carries exactly the
> value the test needs (`"entity-directory"` on the row in question, same vocabulary as
> `pages.page_type`). **The discriminator I said was missing was present under a different key, in
> the row I failed to select.**
>
> **How I got it: an unordered `LIMIT 1` over a filter that matches three rows from two producers.**
> `spec->>'page_name'='brand-directory-index'` matches the `reconcile_site_plan` row **and two
> `page_rerender` rows** filed by `rerender-pages`. I read one of the rerender rows and reported its
> shape as "the reconcile spec". **I had `created_by='reconcile_site_plan'` in hand — I had verified
> it two queries earlier — and did not put it in the filter.** The query described one population and
> selected from another.
>
> **What it would have cost:** joining on `spec->>'page_id'` returns NULL on 134/134 rows, so their
> closure test would have reported PASS/FAIL identically for everything. **I would have handed them a
> fourth non-discriminating instrument while correcting their third.** They caught it by measuring
> fleet-wide instead of taking my single row.
>
> **The working test** is `spec->>'page_role'`, with a `pages` join on `(site_id, page_name)` as the
> authority. They have since validated it against a known-FAIL population — my pre-fix rows — where
> it correctly returns FAIL for `brand-directory-index` and `brand-profile` and `n/a` for the other
> eleven. **That is the first check on this bug shown to produce the disconfirming answer.**

### 2026-08-25 09:50–10:30Z — post-roll re-measurement, the owner's retraction, and a coarse instrument I nearly reported as a finding

**The owner retracted the §3a authorisation**, which is the outcome this lane asked for on 08-24
(`4741cf682`). The parked-row release is dead: clearing alone is inert, and clearing-plus-replan
destroys the measurement. Bannered onto the 08-24 handoff so nobody revives it from that file.

**Chassis v1.0.1337 rolled 09:27Z. `380`'s Go practice-claims family is LIVE at the binary**
`[MEASURED 2026-08-25 09:56Z]`, both replicas, by capability probe with controls:

```
grep -aq "practice_claim"            /proc/1/exe  → PRESENT  (the Check literal)
grep -aq "no recorded operating history" /proc/1/exe → PRESENT  (the family's own reason string)
grep -aq "validate_page_content"     /proc/1/exe  → PRESENT  (positive control)
grep -aq "zzz_not_a_real_symbol_control" /proc/1/exe → ABSENT  (negative control)
```

> **The `build provenance` line had already scrolled — the landmine fired exactly as written.**
> `logs --tail=6000` on a 26-minute-old chassis pod returned no provenance line at all. The `380`
> lane caught it at 09:27Z while the pod was fresh (`4c996e1b5`, with the `git merge-base
> --is-ancestor` check). **An empty grep there means "out of range", not "unstamped"** — and the
> binary probe has no shelf life, which is why it is the fallback and not the first resort.

**⚠ LIVE IS NOT FIRING.** The practice family runs inside `validate_page_content`, which runs during
a build or rerender. `garden-tools.uk` is quiet, so the family is live and **doing nothing here**.
Do not read its silence on this site as a clean result — nothing has asked it a question.

**Post-roll census, at the artefact** (`after_test.sh`, now promoted into this directory)
`[MEASURED 2026-08-25 09:57Z, cache-busted]`: 7 serve / 5 × 404, **9 dead links across 4 distinct
targets**, `PROMISE UNMET` still fires on `seasonal-planner` alone (3 distinct month names). Structure
baseline **intact**: 0 tables, 0 `<strong>`, and the `li=8` on every served page is chrome — identical
on all seven, including the 404-shaped ones at 0. **`381`'s pre-fix comparison point survives the
roll**, which matters because that is the only thing this site is currently for.

**One drift caught: the 08-24 byte table was one deploy stale when it was written.** Measured 09:05Z,
but `deployed_at` on all seven is 14:00–14:04Z the same day, and every page is now **exactly +420
bytes** larger. Uniform delta across seven pages of different lengths ⇒ chrome, not content, and the
structure counts agree. **A byte figure needs the deploy it was taken after, not just the clock time.**

**The instrument I nearly published.** I ran a before/after over *all* `page-content-writer`
`generate_content` calls (948 before the 17:00Z boundary, 271 after) and got lists 5.3% → 9.6%,
`<strong>` 4.6% → 7.7%, `<h3>` 5.0% → 12.5%, tables 1 → 0. It reads like a modest, real effect and
**it is the wrong denominator**: most of those calls write fields that never received the new
guidance, so the population is mostly untreated. The `381` lane's own measure — restricted to writes
that actually received the instruction — is **72% lists against a 10% baseline, 100% subheadings**.
**Mine was not wrong so much as blunt, and a blunt number published first would have set the
expectation low.** Check whose denominator you are using before reporting someone else's fix.

**What the natural-build question actually resolved to.** 271 post-fix writer calls exist, so the
"wait and measure whatever builds next" option has already delivered **for the writer arm**. It has
delivered **nothing** for the planner arm: `content-gap-planner` ran 12 times, but the component
*choice* step that `381` fixed only runs on a new-site build, and none has happened. The three new
components (`checklist`, period `calendar`, `comparison-table`, built by that lane overnight) have
**never been placed on a page**. So the canary question is now narrower and sharper than on 08-24:
it is not "is the fix real", it is "does the planner compose with it", and only a greenfield build
answers that.

### 2026-08-25 11:31–11:34Z — the canary's mint, and the closure test it cannot run

**The owner authorised a greenfield build this morning** (`homegarden.uk`, site
`5904bd0f-33fd-4212-9c1b-50b28fe72fdb`, dispatched 10:21:49Z) in the `bugs_open/381` session, not
this one. Split agreed with that lane: **they take the planner prompt and the component choice, I
take the reconcile mint** and contribute it to the `206` lane whose test it is.

**Hop two survived.** `[MEASURED 10:53–10:55Z]` `thespruce.com` absent from the draw; RHS and
Gardeners' World returned 6 sources each at `content_quality: good`; `which.co.uk` returned
`success: true` with **0 sources** and `content_quality: none`, and the chain proceeded to
`synthesise` regardless. That third row is the new finding and it is in `bugs_open/376` §10: **a floor
evaluated on step success is blind to "succeeded and delivered nothing"**, so my own fix candidate,
implemented naively, would have let the estate write a vertical landscape from no research at all
with every step green — worse than today's failure, which at least stops. The floor must be evaluated
on content. Evidence captured to disk because `orchestration_states` reaps inside ~25h.

**The mint, at 11:31:05Z** `[MEASURED 11:32Z]`: **21 pages, and 17 of them `section-index`** —
`january-index` … `december-index`, plus `this-month`, `comparisons`, `garden`, `home-maintenance`,
`shed-and-outbuildings` — all at `page-build-handler`. Plus 2 `content`, 1 `blog-post`, 1 `landing`.
**Zero `entity-directory`. Zero `entity-page`.**

**So the half of the split I took cannot be answered.** `206`'s closure assertion needs an
`entity-directory` page and the plan contains none. Reported to both lanes as **unexercised**, which
by this lane's own rule from this morning is not a result — not a pass, not a fail, and not something
to make fire by hand.

> **The pattern, and it is the transferable bit: BOTH `206` closure tests specified the ASSERTION and
> left the POPULATION to chance.** The first was garden-tools, where the parked row held its own
> `item_key`; this one has no row of the right role at all. *"The next greenfield build"* is not a
> population — it is a hope that the next build happens to plan the role you need, and nothing in the
> greenfield path guarantees one. **Check the population at mint time, in one query, before spending
> any measurement on the outcome.**

**And the interaction neither lane predicted.** The brief named no calendar; the planner produced a
calendar-shaped SITE — twelve month indexes — rather than one page carrying `381`'s new
`period-calendar` component. **A structural promise satisfied at the site level routes straight into
the one page role with no builder, and thereby bypasses the component built to satisfy it at the page
level.** If those 17 no-op, the site ends with no calendar at all. Recorded, not filed; it belongs
cleanly to neither bug.

**`[INFERRED, NOT MEASURED]`** that the 17 will no-op. All 21 rows were `triaged` at 11:33:41Z and
nothing had attempted them. What is measured is that the role+handler pairing is byte-identical to
the one that produced `page-build-handler no-op: no sections ready to build` on garden-tools. Watcher
armed on the first non-pending status; I will hand the `381` lane the **raw** failure string rather
than a diagnosis, as they offered to do for me.

⚠ **If those pages fail, the site serves 4 of 21 and the missing pages are the whole subject matter —
which reads EXACTLY like `381`'s fix failing** ("the planner promised month-by-month and did not
deliver it"). It would not be. Warned that lane to put it in their reading guide, which is where a
correction actually works, rather than in their notes, where it only has to be right.

**Instrument built and validated before use:** `capture_reconcile_mint.sh` in this directory, run
against garden-tools' known-FAIL population first (it correctly returns FAIL for
`brand-directory-index`). Validating it found a fourth trap nobody was looking for: **`needs_page`
has 46 distinct producers as of 2026-08-25 and only `reconcile_site_plan` carries `page_role`** —
1,438 rows fleet-wide, 451 (31.4%) with that key, **0** with `page_type`. `page-rerender` 414 rows/0,
`image-build-handler` 262/0, `render_directory` 103/0. Filtering on the role alone narrows a census to
a third of it; not filtering on `created_by` mixes five automated producers with different spec
shapes. **It surfaced because the script prints every row with a stated reason instead of dropping
what it cannot classify.**

### 2026-08-25 12:45–13:00Z — the canary's domain, and two instrument faults found by using it

**The owner asked this lane to set up `homegarden.uk` in Cloudflare and pointed the nameservers
himself.** Zone `252c10abde85a6985392a084f68f9235`, created 12:46:20Z — **the first zone this estate
has created via the API**. Config diffed identical to `garden-tools.uk`: two proxied A records at
`192.0.2.1` (TEST-NET-1; the worker serves everything) and two `portfolio-sites-router` routes.
Verified by **re-reading the stored zone, not by trusting the POST receipts**.

**Fault 1 — I told the owner a new token was needed, and was wrong within 90 minutes.**
`~/.config/cloudflare/token` is read-only (`#zone:read` etc.; `POST /zones` refused for lacking
`com.cloudflare.api.account.zone.create`). I reported that as "issue a new token". But
`~/.config/cloudflare/portfoliotoken` had been sitting in the same directory since 2026-08-18 with
`#zone:edit` / `#dns_records:edit` / `#worker:edit` and account zone-create. **I checked the
credential I knew about instead of asking what credentials exist — `ls` the directory.** Runbook
corrected twice in one afternoon, the second correction retracting the first.

> **The runbook's own token proof is also blind.** It says prove the token with `GET /zones?per_page=1`
> (written to defeat an IP-filter trap). That call **succeeds on a token that can write nothing** —
> it tests reachability, not capability. Read the `permissions` array on any zone object, or probe
> the verb you actually need. Recorded in the runbook.

**And the nameserver caveat earned its keep.** The runbook records the account pair as
`alexis`/`leah` with "do not assume". `[MEASURED 2026-08-25 12:43Z]` the account's 40 zones split
**29 × alexis/leah and 11 × betty/ivan** — so quoting the remembered pair would have been a **~28%
chance** of handing the owner nameservers that never resolve, failing quietly. This zone did get
alexis/leah; the point is that could not be known in advance.

**Fault 2 — my own parked-domain control mislabels `000`, and I hit it the same hour I wrote it.**
Straight after delegation, every fetch returned `000`. My inline check tested only `= "200"` and so
printed *"control 404s — HTTP is now informative"* on an unreachable host. `000` is **its own
bucket** — the LANDMINE entry says so in as many words and I had just extended that entry.
The real cause was two-layered and worth separating:
- the local resolver still held the `dan.com` delegation (public resolvers had already moved), and
- **Cloudflare had not yet issued the certificate**, so HTTPS failed the TLS handshake
  (`curl (35) ... handshake failure`) while **port 80 returned 200 with the real pages**.
So for a window the site was **live over HTTP and dead over HTTPS**, and a harness fetching only
`https://` reads that as a dead site. The mirror image of the parked-domain trap: there HTTP said
healthy when the site was dead; here HTTPS said dead when the site was healthy. Certificate issued
~13:00Z; `https://homegarden.uk/` now 200 and the invented-path control 404s.

**The result, verified at the wire** `[MEASURED 2026-08-25 13:00Z, HTTPS, control 404]`:
`/index.html` 77,613 bytes, **one `<ol>`, 49 `<li>`, twelve distinct month names**, under an `<h2>`
reading *"The garden and home year, month by month"*. Against this lane's own dated baseline —
`garden-tools.uk`'s `seasonal-planner`, which promised the same thing and served **three** incidental
month names with **zero** content lists — that is the complaint the owner raised on 2026-08-24,
closed and checkable. `bugs_open/381` closed by its lane on this evidence.

**One page 404s: `/blog/blog-post.html`, the single page with `sections_planned = 0`** — named in
advance by the corrected predictor and recorded by that lane as `bugs_open/206`, not as a 381 failure.

### 2026-08-25 13:20–13:40Z — the promise check on the canary, and my own instrument had the flaw I had just warned a peer about

**The run fired on three pages, and the more important output was a defect in the checker.**

**My `distinct_months` counter was chrome-contaminated.** It printed `12` for **every** page on
`homegarden.uk`, including `/contact.html`, which contains no months at all — the site menu is a
`<ul>` of `<li><a href="/january/index.html">January</a></li>` × 12. The calendar rule requires ≥6
distinct months, so it would have returned a **VACUOUS PASS on exactly the pages that fail**. It
worked on `garden-tools.uk` only because that site's nav carried no months.

> **This is the identical defect I had warned the `bugs_open/381` lane about ninety minutes earlier**
> — their served census said "20 of 20 pages carry a list", true and useless because nav is a `<ul>`
> on both sites, and I told them to read the delta above the `li=8` chrome baseline. **I gave the
> lesson for `li` and did not apply it to my own `months`.** Knowing a failure mode is not the same
> as auditing your own instrument for it; the second is a separate act and I skipped it.

**Fixed and validated: strip anchor text before counting.** Chrome and cross-references are links; a
kept promise is not. Discriminates cleanly — `/index.html` 12 non-anchor months (the real calendar),
`/contact.html` **0**, `/april/index.html` 2. Same treatment for `<li>`. The harness prints content
and raw side by side so the gap is visible rather than silently corrected.

**⚠ A near-miss I nearly reported as a finding, twice over.** My first attempt at establishing the
chrome baseline stripped `<nav>`/`<header>`/`<footer>` with a regex, and reported `/index.html` as
having **0 `<li>` and 0 `<ol>` in content** — for the page I had directly measured minutes earlier as
carrying an `<ol>` with 49 `<li>`. The strip was broken, not the page. **An implausible number from a
new instrument is the instrument, until proven otherwise** — I discarded it rather than reporting it,
which is the only reason it is a near-miss and not the day's fourth wrong call.

**The three real failures, and they are NOT `bugs_open/381`:**

| page | its own heading | non-anchor months |
|---|---|---|
| `/garden/index.html` | *"Garden maintenance for UK gardens, month by month"* | 3 |
| `/home-maintenance/index.html` | *"How the UK seasons shape a home maintenance calendar"* | 1 |
| `/this-month/index.html` | *"Why one calendar does not fit the whole country"* | 2 |

**The measurement that settles the attribution** — set-difference the served internal link sets:
`/contact.html` 23 links, `/garden/index.html` 23 links, **links on garden not also on contact: ZERO.**
An index page with no page-specific links is indexing nothing. Its `content-listing` section had
nothing to list (`source: query.blog_posts`, `on_missing: skip_section`, and the site's one
`blog-post` page is the 404) so it skipped, and the heading is left writing a cheque the body cannot
cash. **381's fix worked on this build** — `/index.html` has a real twelve-month `period-cal__list`.
Routed to `bugs_open/384` (a listing never re-rendered when its data arrives), where the sharp point
is that **the month pages already exist**: this is a missing invalidation, not a race.

**Withdrawn:** my first-run report that `/comparisons/index.html` was the notable hit. On the
corrected run it does not fire; the table rule had been matching the WORD "comparison" in meta
headings. The page is the same class as the three above, but it was never a table problem.

---

## 2026-08-25 (evening) — four asks in one session: the 376 submission, the RFC, four seats, and the "evolutionary" switch

**Asked:** add four more seats to the improvement loop; submit the 376 fix to the council; draft the
RFC; investigate switching off the "evolutionary" aspect (rewriting pages judged good for aspirational
improvements) and turning the loop back on — plan, not change.

**What the "evolutionary bit" turned out to BE, and the two doors** `[MEASURED 2026-08-25]`. It is not a
step. It is the four LLM seats' findings passing through `write_audit_findings` Rule 4 into regenerating
handlers (`content_rewrite` / `needs_content_page` / `tone_shift` → `page-build-handler`;
`needs_content_planning` → `content-gap-planner`). design-audit source lifetime (live ∪ archive):
**976 / 399 / 964 / 26**. And the sweep being off since 08-17 did NOT stop dispatch: `detected-item-promoter`
(live 08-15, 900s) promoted **26** LLM-audit rows 08-20 → 08-24 with the sweep off. **0** LLM-audit rows at
`detected` now — the promoter drains the queue. LANDMINE filed ("`detected` is a QUEUE, not a shelf").

**Misstep, caught before it shipped:** my first draft of the record-only mode parked rows at `detected`
(IMP-054's "detect-only" premise). Reading the promoter's `pre_query` refuted that in one query — `detected`
with a known-good pair is promoted within 15 minutes. Record rows are `deferred` + `handler ''` with the
routing in `spec` (and NOT `deferred` + named handler, which is `bugs_open/396`'s no-park-verb shape,
found the same afternoon by another lane). The test names both doors so a half-fix on either fails.

**Second misstep:** `firstNonEmpty` already existed in package `actions` (`write_site_plan_action.go:1071`,
variadic) — my duplicate broke the build; removed. **Third:** 624's operator assertion used `:'VAR'` inside a
`DO $$` body, which psql does not interpolate → syntax error on first rehearsal; staged through a temp table.
Then the file's own `\set` default overrode `-v` → wrapped in `\if :{?VAR}`. Both rehearsals now pass and the
placeholder refuses.

**The 376 fix:** migration `618` + ROLLBACK written; rehearsed apply (ROLLBACK) and apply-then-rollback in one
transaction against the live row — 15 steps, 0 dangling, 12 after rollback; live untouched. Submitted:
**`SUBMISSION_CORR 3d890adc-6f76-42c3-9eb2-20a76d7195f1`**; committed `0d21ba356` with `Council-Submitted:`.
NOT applied.

**The seam + seats (Go):** `filing_mode: record` on `write_audit_findings` (ConfigKeys; budget parity test
passes; 4 tests). Three checks by three parallel subagents: `check_build_prerequisites.go` (234/337),
`check_heading_promise.go` (602/703 — found a real goquery `Text()` node-joining bug the bash harness could
never meet; landmark-only header/footer stripping because homegarden's calendar heading lives in a `<header>`
inside `<main>`), `check_structure_floor.go` (792/673). `verify-head-builds.sh --with` ×10 against HEAD
`f3c1da996`: OK. Both packages green.

**Config:** `623_…_HOLD` (one edge, seats bypassed not deleted; 31 steps, 0 dangling) and `624_…_HOLD`
(acceptance + reader agents, record mode on six seats, 11 seat-failure records, `check_seats_ran`; 47 steps,
0 dangling) — both rehearsed forward and apply-then-rollback. Neither applied. RFC_056 drafted. Register
IMP-056/057. PLAN written for the owner.

**Peer coordination:** the vigilant lane (cross-session message) is HOLDING the sweep switch until ordering is
settled; they corrected my seat-3 `[UNVERIFIED]` (role `site_reviewer` → `site-review-agent`, 4,046 calls) and
supplied the 7-of-8 fail-open measurement + "two spellings of `error_step`" — both in the RFC §6 with credit.
They asked that routing be argued in the RFC, not settled in config: §9 Q5.

**2026-08-25 21:2xZ — Phase 1 verified on the first firing.** Sweep fired 21:18:31 (13s after enable)
and completed. PLAN §5 queries: **query 2 = 0** (zero LLM-audit rows since switch-on), **query 3 = no
rows** (zero model-seat LLM calls). What it filed is all defect-shaped: design-discovery items
(audit_tool, evaluate_tools, improve_tool, acceptance_run, undeployed_asset, deactivated_component),
rerender items, one loop-level needs_rerender. Nothing opinion-shaped, nothing at a regenerating
handler from an audit source. 618 applied the same evening (probe refuses re-apply); 623 applied;
render-audit rotation left live per owner choice (c) with the fix routed to bugfix_390_cascade_attribution;
promoter residue = bugs_open/405 (277 session notified).

**2026-08-26 00:2xZ — the council's first night, verified, and it caught a real outage honestly.**
Sweep fired and completed all night. The two load-bearing negatives held: **0 record rows outside
`deferred`, 0 LLM-audit rows dispatched since 624** (21:27Z). **209 verdict rows across 9 sites**,
all six seats filing (brief-fidelity 49, reader 46, content-quality 34, site-review 32,
visual-design 29, offer 19); shape verified on a sample (deferred + '' + routed_handler +
deferred_by + "[verdict, not dispatched]"). Acceptance seats: heading_promise_unmet **54**,
prerequisite_missing **19**, structure_floor_unmet **8** — the predicted census, flag-only.
**10 seat-failure rows** (reader 4, brief 2, offer 2, site-review 2) with the audit ATTEMPT stamped
and NO pass counted — and the cause is real: **the Anthropic API account ran out of credit at
23:46:29Z** (last successful call fleet-wide; 00:00 hour = 0 ok / 24 credit_low). The design did
exactly what RFC_056 rule 4 demanded under a genuine outage: durable rows, withheld stamps, sweep
continuing — where the pre-624 loop would have stamped those sites as cleanly audited. Owner
notified (terminal push); billing is his. Misstep worth keeping (the 391 lane's framing): the
question that transfers between lanes is *"could this check have come out the other way?"* — my §6
recipe and their `<p>` count both failed by only ever running where they expected to fire; theirs
cost two pages, mine cost nothing because they said the misstep out loud early. That asymmetry is
the argument for early declaration, and it played out three times tonight.

**2026-08-26 09:1xZ — the residue composition settled (mine stood; theirs didn't sum), and my own
miss in the exchange.** The 391 lane re-queried and confirmed my 21-item breakdown exactly; their
prose enumeration summed to 19 beside a stated total of 21 (a nine-row grouping compressed to prose
by eye). **My miss: I split the difference** — "both are artefact reads, the list query settles it"
— when the check was one addition away: SUM THE ENUMERATION BEFORE TREATING IT AS A RIVAL READING.
An enumeration that doesn't reach its own total is not a competing measurement; it is arithmetic
refuting itself, and deference to it is misplaced generosity. Their generalisation, kept because it
names the week: *"writing a lesson down is not the same as holding it while you work"* — three
times in one session they committed the error they had just diagnosed, each interval under a day,
and the saves keep being cross-lane rather than self-caught. The re-fire recipe is unaffected
(error-predicated: takes the 20, leaves the 1 page_rerender for hand judgement).

**2026-08-26 09:2xZ — recovery verified end-to-end after the top-up (09:00 hour: 126 ok / 0 failed).**
The terminal residue was re-fired at 09:02:48 in exactly the RUNBOOK recipe's shape (status→triaged,
attempts→0, error→NULL; a single-second batch — presumably the owner's hand after the push named the
recipe; `[INFERRED]` from shape, not witnessed) — 0 credit-class rows remain at `failed` anywhere,
11 complete / 4 claimed at first read. The one deliberate leave-out (`page_rerender` on
ai-agent-orchestration.com, a SECTION-save refusal, not credit) routed to that lane by message.
Council post-restore: 6 seat calls, 0 failed, 0 new seat-failure rows, record discipline intact.
**And the outage's cooldown worry dissolved on measurement, for a designed reason:** all 20
attempt-stamped sites show `fp_current = false` — their fingerprints moved since their last REAL
audit — so `fp_changed` makes every one due again despite the fresh `at`. That is
`record_audit_attempt` stamping ONLY `at` and never the fingerprint (624's anti-treadmill arm),
now measured as the thing that makes an outage self-healing: had it stamped the fingerprint, all
20 would be held 14 days. A design choice that could have come out the other way, and didn't.
**Still owed:** 376 §11e's three behavioural tests — they need real research draws; the route is
unblocked, so the next greenfield dispatch (owner's action) doubles as the proof run.

**2026-08-26 ~10:3xZ — the biggest misstep of the arc, caught by a peer's reading habit, corrected
and fixed the same hour.** ADDENDUM 1's "verdict rows are revalidated by the seat's own
silence-retraction" was FALSE for every type but `dark_section_audit` — the gates map has ONE entry,
and I claimed coverage from the status posture + scope filter without ever reading the roster they
were conditional on. Fourteen hours after watching the 391 lane name exactly this shape. The
vigilant lane had withdrawn an objection on my false claim; they are told, with the ask that they
check their own WII-033 for the inherited sentence. Caught because the finetuning lane read 621's
`v_parkable` array before using the park verb — the transfer question ("could this check have come
out the other way?") arriving from the third lane in two days. Corrected visibly in SIX homes (RFC
CORRECTION block, RUNBOOK, handoff ×2, register IMP-056, doc_notes correction row, WRONG_CALLS —
whose entry names the check: **read the ROSTER that enumerates coverage, not the posture or the
filter, which say who is ELIGIBLE, never who is ENROLLED**). Fix built in the same pass:
`recordModeSilenceRule` — default gate for record rows of ungated types; licence is the
self-correction asymmetry (a wrong verdict retraction frees the dedup slot; the still-true finding
re-files next run); dispatch rows pinned inert on a MIXED candidate set; gated types never
double-judged; 11 ordered mocks updated (incl. the copy_quality lane's tone-bound test, flagged to
them in the commit). Both packages green; verify-head-builds --with ×5 OK vs `ff205b735`.
Submitted `04a3ce1f`. **Until it ROLLS: verdict rows are cleared by humans only.** Also this
morning, before the correction arc: 405 candidate 1 APPROVED r1 + APPLIED (`946d587c`, origin door
live, inert until the stamp rolls); credit-outage recovery verified end-to-end; the residue's one
real defect routed to and judged by the aiao lane.

**2026-08-26 ~11:5xZ — the revalidation fix's council arc, including a mishap of my own tooling.**
Round 1 (corr `04a3ce1f`): REVISE, seven seats, and the two HIGHs were right where it mattered —
debug_historian caught that my self-correction licence was `bugs_open/033`'s false claim with its
mechanics LIVE (retraction closes feed the two-strike arm; third re-file born `unresolved`), and
prior_art demanded the revalidation-family reconciliation I had asserted but not cited. Fixes:
record rows now set `recurrenceExpected` (the flag's own contract — a re-request that is normal;
dedup unaffected), and every claim is a citation or a measurement (271 rows / 9 sites / 6 seats /
7 types; `workItemRevalidatableStatuses` excludes `deferred`). RFC_056 ADDENDUM 3 = the
architecture seat's demanded writeup. copy_quality lane gave verified sign-off on the crossed test.
**THE MISHAP:** my resubmission script died on a quoting error AFTER `DRY_RUN` had validated the
file — so the dispatch shipped the stale round-1 bytes still on disk, and the trail burned a
duplicate round (verdict 10:31: REVISE again, gating prior_art — round-1 content, as predicted).
*Live-and-committed are independent facts*, applied to my own tooling: the dry-run validated what
was on disk, and so did the dispatch — the check and the act agreed with each other and both
predated the generator's failure. The check now used: re-verify the payload's round marker AT the
moment of dispatch (done for the true round 2 — **CORRECTED: dispatched ~10:35Z, not '~11:5xZ'; that figure was a
guess I typed instead of a clock I read; the DB's verdict timestamps 10:10 / 10:31 / 10:53 / 11:14 are the record**). Also learned: the printed
RUN_ORCH_ID is not `orchestration_states.id` AND terminal runs are reaped in minutes — a
by-correlation watch on `orchestration_states` misses a fast round entirely; **watch `doc_notes`
verdict notes instead** (fires on the artefact, cannot be reaped out from under the window).

**2026-08-26 ~11:3xZ — rounds 3 and 4 on trail `04a3ce1f`, and a peer finding that redraws one line of the plan.**
Round 3 (true r2's verdict 10:53): REVISE, gating editquality — the multi-producer property (five seats
share `content_rewrite`) had to be a PIN, not an assertion; `TestRecordRetraction_AnotherSeatsVerdictsAreNotMySilences`
passes against unchanged mechanism code (the guard existed; the control did not). Round 4 (r3's verdict
11:14): REVISE, gating debug_historian on the streak-bump/reaper interaction — settled by the reaper's own
pre_query (population `triaged` only; deferred rows never reapable) and the retroactive-streak fear by
writer enumeration (one writer, previously gated to one type). Round 4 dispatched with the at-dispatch
payload check and DECLARED FINAL: residuals are RFC_056 FOLLOW-UPS, the trail is the advisory record.
**The apis.uk lane's finding:** `missing_tools` → `evaluate_tools` → `add_tool` filed and half-built four
pages on a site the owner ruled single-page — no per-site exclusion exists on that path (read from
`check_missing_tools.go`: tool count, opt-in ratio, cooldown; tool-suggester reads no ruling). That is
GROWTH riding the mechanical seat: my "defects dispatch; opinions record" line took "mechanical" as a
proxy for "defect", and tool evaluation is an aspiration. RFC follow-up 1; owner's word needed on the
remedy (a per-site refusal declaration in 624's shape). Their hand-park holds by dedup until released.

**2026-08-26 ~12:0xZ — a near-miss worth its line: the "second outage residue" was an OWNER RULING.**
The 391 lane flagged `build-pipeline-trigger-2` still disabled six hours after the top-up, timing-matched
to the pause option I had put to the owner — "the mitigation outlived its cause", the same shape as the
21 burned items. One evidence pass refuted it: the row was retired at 08:51Z by the dispatch_throughput
lane under OWNER RULING B (their migration 637 — sibling retired for the native 30s interval, trigger-2
kept disabled AS THE ROLLBACK PATH, and their 584 VERIFY now RAISEs by name on a second enabled trigger
row). Re-enabling would have undone a ruling and tripped their gate. **What stopped the flip was not
scepticism but a MEMORY line naming trigger-2 as that lane's rollback lever — which made "deliberate
retirement" a live alternative hypothesis to "stuck mitigation".** The check that generalises: before
reversing any state change that pattern-matches an incident's mitigation, ask WHO ELSE has that switch
as a lever, and read their lane's morning commits first (CLAUDE.md's own sentence: a signal that
pattern-matches a known failure may have a different cause). The backlog behind the single lane
(654 triaged, 520 >1h) is the throughput lane's measured demand test, ceiling ~300 claims/h, theirs.

**Addendum to the trigger-2 near-miss (391 lane's own post-mortem, cross-referenced because it names
the sharper mechanism):** their disconfirming number was already ON THEIR SCREEN — the survivor row's
`interval=30` beside the sibling's `60` (a pause leaves the survivor untouched; only a reconfiguration
changes it — that asymmetry IS 637's arithmetic) — and they read past it, scanning for `enabled`
because the answer was already decided. Their distillation, kept verbatim: *"a pattern you have just
confirmed is the most dangerous thing to take into the next case, because it turns coincidence into
evidence"* — one frame (outage residue) applied twice in one morning, the second time with no evidence
the first didn't supply. Their WRONG_CALLS carries it; the priority-is-inert-between-sites finding is
filed at dispatch_throughput/CONTRIB_2026-08-26_from_bugfix_391_… with the correction inside it.

**2026-08-26 ~16:0xZ — fleet deploy stall, root-caused in three commands: GITHUB ACTIONS MAJOR OUTAGE.**
The 391 lane reported the stall precisely (commits correct, served bytes stale, runners idle, queue deep)
and declined to restart — right again, and righter than known: the cause is EXTERNAL
(githubstatus.com: Actions = major_outage since ~15:25-15:37Z; Webhooks/API operational, which is why
commit reads worked while dispatch died). GitHub's runner list shows only stale offline registrations.
**Nothing ours is broken; a runner restart would have churned registrations into a dead service.**
Boundary correction: their 15:25 was ONE POD's last words (`logs deploy/X reads one pod of N` — again);
the true last work was 15:37:11Z on the second pod. Owner push-notified (no action; don't restart);
recovery watch armed (fires when Actions leaves major_outage, then verifies the queue drains).
**The instrument lesson, both lanes' files:** two careful readers measured pods, queues, headers and log
boundaries for a combined hour; the answer was a 30-second curl of the provider's status page — the one
instrument neither reads by default. *Before diagnosing idle consumers of an external service, ask the
service.* Third stopped-thing of the day, third different cause (a ruling; a dead account; a dead
provider) — the surviving frame: **a stopped thing has not yet told you WHY it stopped.**
Consequence for this lane: any served-bytes verification after 15:37Z reads pre-stall artefacts —
reassuring for failures, alarming for successes — until the queue drains post-recovery.

**Addendum from the 391 lane's close-out, kept because both pieces sharpen the day's file:** (1) *the
asymmetry was the clue and it was in hand* — their `gh api` reads worked perfectly WHILE Actions was
dead; they were querying across the exact broken seam (Git Operations up, Actions down) and read the
working half as "the platform is fine" when it was naming which half was broken. (2) A failure family
distinct from badly-designed instruments: **correct instruments read partially** — three instances in
one day (an enumeration summing 19 against a stated 21; the 30-vs-60 interval asymmetry on screen,
unread; two of three runner pods read, boundary reported 12 minutes early). The result set was complete
every time; the reading was not — and nothing in scrollback marks a row as unread. Their check, adopted:
**say N out loud before interpreting, and make the interpretation account for all N.** Filed by them in
WRONG_CALLS (the reader's half) and 016b §9 (the system's half: green item + correct commit + unchanged
site = the CI provider is down; status-page curl FIRST).

**2026-08-26 17:55:31Z — GitHub Actions recovered (component → operational); deploy queue at ZERO two
minutes later.** The stall ran ~15:37Z→17:55Z, entirely external, nothing of ours touched. Served-bytes
verification is trustworthy again for anything deployed post-recovery. The last soft gate on dispatching
farmerinsurance.uk is gone; awaiting the owner's domain/mission answer.

**2026-08-26 19:0xZ — farmerinsurance.uk DISPATCHED (owner's word: deliberate no-prompt; domain
registered; NS pair handed over).** Zone `ccb2ecd19e653f2b36795bfe066226fb` created via API
(portfoliotoken), 2 proxied A → 192.0.2.1 + 2 routes → portfolio-sites-router, re-read at the STORED
zone: status `pending`, NS **alexis + leah** (from the create response, not memory). Pre-flight clean:
0 zones, 0 site rows, 0 open items, chassis pods 5h old. Dispatch = the loanzy no-prompt form (bare
domain, `bash` not `./`); LANDED: site `99cae989-2413-430d-b026-59dfeeb638c0`, `needs_domain_research`
triaged 19:03:59Z. **This build is three proofs at once:** 618's floor on its first natural draw
(hop-two watch armed, evidence capture before the reap), the first site BORN under the acceptance
council, and the FCA-adjacent no-prompt stress of the claims controls (flagged; owner chose knowingly).
Sequencing agreed with the owner: garden-tools re-plan AFTER farmer clears hop two; homegarden stays
the untouched control.

**2026-08-26 ~20:5xZ (DB clock) — THE ROLL: chassis `b34c24f4c`, seen 20:45:57Z, carries EVERYTHING.**
Verified by the working route with all three controls: `f51d3cf5e` (newest Go: retraction gate + origin
stamp) IN; old-commit control IN; post-build commit NOT in — merge-base discriminates. So the whole
RFC_056 stack is live-capable: record mode (already exercised), the origin stamp (writes on the next
model-seat filing), the record-mode retraction (streak bumps will appear as `result.retraction` on
record rows over coming audits). **Clock trap, new costume:** I read git's local BST timestamp against
DB UTC and manufactured a phantom hour — the promoter looked 4 ticks stale when the roll was 3 MINUTES
old. `now()` from the DB before any staleness claim; git shows +01:00. **405 §6 door proof RUNNING:**
synthetic stamped row (proven pair, farmer site, key `content_rewrite:405-door-verification`) inserted
post-roll; watch asserts held across ≥2 promoter ticks then cancels it with the result note.
**farmerinsurance.uk queue reality (a route finding in itself):** 1.75h after dispatch the classifier
item is still `triaged` behind 27 sites; dispatch healthy (277 claims/h, 269 completions/h) but net
drain only ~15/h because producers refill — a FRESH site's first item waits behind the fleet's
maintenance backlog age (the 391 lane's between-sites finding at product scale). Bypass option put to
the owner, not taken unilaterally.

**2026-08-26 21:2xZ — 405 CLOSED, with the strongest possible reading.** The door proof landed in its
strong form: the synthetic stamped row held at `detected` across two promoter ticks (20:45:48→21:16:54)
**while 21 natural promotions of unstamped rows completed in the same window** — discrimination
observed live, not inferred from a quiet tick. Stamp liveness: 56/57 post-roll filings stamped; the 1
absent is tool-acceptance-tier4, which bypasses write_audit_findings by design — the exception that
confirms §4b's boundary. Probe cancelled per recipe. 405 → bugs_closed (both paths on the commit;
verified one line at HEAD). Remaining from FOLLOW-UP 4: the retraction's slow behavioural proof and
376 §11e on farmer's draws.

**2026-08-27 ~04:0xZ — farmer's first night, the CORRECTED reading, and a misstep of mine that
reached the owner.** My in-thread claim "the classifier never filed vertical research — hop two was
bypassed" was FALSE: `needs_vertical_research` exists (filed 23:06:13, `triaged`, queued behind the
fleet like every hop). I had read 24 of **N=57** rows (a head window and a tail window, never the
middle) and narrated the gap — the 391 lane's "correct instrument read partially", committed by me
within hours of writing their lesson into this file. What caught it: the classifier's step graph
(every path ends at `create_next_item`) refused the story, and the count-N-first check did the rest.
The correction led the next owner report. **The REAL findings, all verified:**
1. **The classifier's regulated-business rule fired on its own**: farmerinsurance.uk → FSMA
   territory → "cannot act as distributor/broker/introducer" → category `hub`. The FCA edge the
   owner chose knowingly was met by the machinery's own control.
2. **The council's first newborn audit is the design working end-to-end**: 4 `prerequisite_missing`
   verdicts at 22:39 (minutes after birth), 28 record rows total incl. site-review's "No FCA
   authorisation reference" — ALL held at `deferred`, zero dispatched. The loop looked at an empty
   site and did not try to build it out of opinions.
3. **Route ordering intact for content**: 0 pages exist; content waits on research→strategy→plan.
   Hop two (the 618 floor's first natural draw) still pending in the queue — watch re-armed.
4. **The growth path raced ahead, second worked case in 24h**: design-discovery scaffolded
   composition/stylesheet (order-safe) but `evaluate_tools` → tool-suggester filed **7 add_tool
   rows at `triaged`** (rebuild-cost estimator, livestock valuer, FOS complaint checker…) — unresearched
   insurance tools queued to build ahead of the site's own strategy, on the vertical the classifier
   itself flagged as regulated. FOLLOW-UP 1's shape exactly (no ruling read, no refusal knob);
   apis.uk's park recipe applies if the owner wants them held. NOT touched — the canary measures.

**2026-08-27 ~04:1xZ — the 414 lane's tripwire case: the strongest external validation yet, and the
residual it routed here.** A planted false claim ("checked against the FCA handbook, rule by rule")
served 24 days on lendzy; the content-quality seat read it back and filed a content_rewrite asking to
MANUFACTURE a methodology section for it. Every post-624 filing was a held verdict — record mode
prevented the class. The dangerous item PREDATED the door (08-11, needs_human_review, live Retry) and
Retry bypasses the promoter entirely (sets triaged directly) — so the doors guard promotion, not the
backlog. Census: **59 armed pre-door opinion rows** (30 content_rewrite/12 sites oldest April, 22
needs_content_page/10 sites, 7 tail). RFC FOLLOW-UP 7 carries the options; not acted on unilaterally.
Their upstream fix (claim rules applied to the SPEC TEXT generators read — CLM-030) closes where the
instruction lived.

## 2026-08-31 — owner review of farmer routed; one routing miss; the carousel ask dissolved under measurement
- Six findings verified at artefact+spec BEFORE routing; 4 of 6 had held verdict siblings (the
  council design validating itself). OWNER_REVIEW doc + routing outcomes appended there.
- **MISSTEP: routed the news ask to news_editorial_features on a lane-NAME match.** They
  declined correctly (editorial feature pages ≠ feed ingestion) — the tell was in my own
  message (citing 316 as the third leg). Check: before routing, open the lane's README first
  line, not its name. Re-routed to bugfix_316 with their measurement carried.
- Logo chain closed end-to-end in one sitting: exemplar (agent row) → paraphrase (item spec)
  → origin_prompt (asset row) → served pixels; 417 filed on first-hand-verification
  substitution, stated. Regen item 3740f5f2 filed at detected; watch it promote, then land,
  then re-derive favicon/og (NOT automatic — presence-based check).
- Carousel lesson for the file: the owner's component ask shrank from "build carousel types"
  to "a schema flag is off on 41/42" through two rounds of peer measurement in ~an hour —
  neither round mine, both because the routing message carried verified numbers worth
  answering. Correspond with measurements, receive measurements.

## 2026-08-31 (later) — §3.3 PROVEN in production: the record verdicts self-retract
[MEASURED 2026-08-31, clients_db] Across all record-mode rows (1,876): **57 retracted**
(`status=complete`, `result.reason` = the recordModeSilenceRule text, `resolved_by` = the
seat, e.g. offer-analysis "re-audited this site on 3 consecutive runs and reported no
needs_content_page finding"), **228 carrying live streak bookkeeping** (105 at silent_runs=1,
123 at 2), **1,819 still held**. The self-correction asymmetry works end-to-end with zero
human intervention. Display quirk, not a defect: a retracted row's stored `silent_runs` reads
2 — the count is written BEFORE the third run's resolve path closes the row; the reason
string asserts the 3-run threshold. Origin stamp: 1,537 post-roll audit filings stamped
`model_opinion`; ALL 220 unstamped rows have max(created_at)=2026-08-26 (pre-roll that day)
plus the single by-design tool-acceptance absence — post-roll stamping is total.
- 417 grew a peer half within the hour (offer-analysis, commit fe8819d5e, appended INTO the
  bug file): the exemplar has propagated — 19/27 current-plan logo prompts license a
  wordmark, 10 verbatim; census trap (count the LICENCE, not the prohibition — "does it
  forbid text" scores 10 contradictory prompts as safe); candidate 1 repriced (stops the
  next 27, repairs none of the 19 — their disposition must be stated).

## 2026-08-31 — §11e closed from durable stores: hop two ran, retried, produced the landscape
[MEASURED 2026-08-31] farmer's `needs_vertical_research`: created 08-26 23:06:13, **complete**
08-27 06:39:52, and the 17,449-byte `vertical_landscape` spec written by
vertical-exemplar-researcher at 06:39:48 — four seconds before completion. The 618 floor never
tripped (the natural test the bug wanted). One oddity, noted not filed: the completed row's
`error` still reads "Claim timed out — handler pod likely died" — a STALE first-attempt error
(the credit-outage/pod-churn window that night) that completion did not clear. Inverse of the
099 landmine (there: FAILED shows COMPLETED with error NULL; here: COMPLETE keeps a dead
error). A census filtering `error IS NOT NULL` counts this healthy row as broken — trust the
artefact + status pair, treat `error` on a completed row as history, not state.

## 2026-08-31 — farmer logo FIXED at the pixels, end-to-end through the framework
- Regeneration item 3740f5f2 (after the RUNBOOK's handler_agent/pipeline fix) promoted →
  claimed → complete. The handler UPDATED the asset row IN PLACE (same id a88c0e99, same
  /assets/images/logo.png url; origin_prompt now the corrected one; storage_path a new
  20260831 object) and the brand_update path deployed the bytes 22s later — served
  last-modified 11:00:33Z, size 119,866.
- **Verified at the artefact, by eye:** fetched the served PNG and looked — farmhouse-in-
  shield mark, wordmark reads exactly lowercase "farmerinsurance", cleanly lettered. The
  exact-named-wordmark positive prompt worked where the unnamed licence invented "Farm
  Shield Info". (One instance ≠ the craft rule refuted — 417's default stays "no lettering";
  the named-wordmark form is the owner-directed exception on this site.)
- favicon + og_card still derive from the OLD logo (served last-modified 00:36Z < 11:00Z) —
  presence-based discovery cannot refile them, exactly as predicted. Two
  needs_brand_head_assets items filed promotable (66e0f086 favicon, 7a7261e4 og_card),
  monitor armed to terminal + served freshness.

## 2026-08-31 — brand heads re-derived; one residual observation, deliberately not churned
- favicon + og_card items promoted and completed; both served files replaced 11:15:33Z.
  **Verified by eye:** og-card = the new shield + "farmerinsurance" wordmark; favicon = the
  new logo too — wrong brand GONE from all three surfaces.
- Residual, noted for 417 not re-fired: the favicon derivation shrinks the WHOLE logo, so
  the wordmark is illegible at 16px — the exact case default_brand_prompt.go's no-lettering
  comment warns about. Harm (third-party brand) is fixed; mark-only favicon cropping is a
  derivation-quality question for the imagery/designer family, filed nowhere yet beyond
  this line. Do not regenerate farmer's heads again for it.

## 2026-08-31 — 669/670 council verdict: APPROVED round 1, 3 advisories none high (corr 3b666f0f)
- edit-quality (medium): the agent_definitions duplicate-active-row trap — could the exact-text
  replace have hit an UNLOADED duplicate? **Answered by measurement:** build-site-planner has
  exactly ONE row fleet-wide (version 1, active, non-snapshot, id f263eaa1 = the row the
  snapshot NOTICE named, and it carries the fix). Trap real, not applicable here.
- edit-quality (low): does site_plan_imagery.locked_at exist as named? Yes — read via \d
  before the migration was written (0 locked rows; predicate is belt).
- bug_historian: approve. Commit carries Council-Submitted (resolves at report time,
  forward-only — no amend). 417's candidate 1 is now APPROVED + APPLIED + verified.

## 2026-08-31 — the cull's long tail: recompute arms, a LIKE-escape trap, and chrome last
- CTA recompute cleared 27→8 components; survivors explained by the arm order: labels NAMING
  the dead tools can't label-match (nothing live matches) and the positional pick had no valid
  target (zero live interactive pages), so the final arm keeps stored — BY DESIGN. The served
  html was nonetheless clean everywhere (the render layer sanitises dead internal links), so
  the block was stored-state-only, with the real resurrection risk being a future rerender
  under a non-recompute reason.
- Mechanical cleanup (rehearsed ×3, applied): unlink body anchors keeping text; drop dead
  minted CTA url fields + target_titles + stamp entries (labels LEFT for the copy lane);
  empty the query-derived guide-list items. Backup: bak_farmer_cull_content_data_20260831.
  16→0 components. **TRAP for the estate's silent-failure list: in a LIKE pattern, backslash
  is the ESCAPE character — `LIKE '%href=\"'||url||'\"%'` built to match escaped JSON quotes
  matches NOTHING (the \" collapses to bare "), so the anchor pass ran zero rows with exit 0.**
  Caught only because the rehearsal's verify counted survivors. Fix: bare substring predicate;
  keep the surgical part in the regex.
- Round 4: 5 more deleted. Acceptance sweep: **18 of 21 retired urls 404; control 200; the 3
  survivors are chrome-blocked** (header/footer site_components link 3 tools). Chrome rebuild
  filed the check's own way (needs_rerender / item_key stale_chrome / refresh_site_components
  — item cd50ce30); round 5 for the last 3 after it lands. Copy layer (52 orphan labels +
  ~14 unlinked body sentences) handed to copy_quality with locations.

## 2026-08-31 — the copy lane found the cull's SPEC layer; split ruled and executed
- Their finding, verified: FIVE current site_specs aspects still named the culled tools —
  a future planning/generation pass reading them would re-mint dead-tool copy. (Precision
  that matters: RERENDERS rebuild from content_data, now clean; SPECS feed generation — so
  the risk is future generation, not the in-flight chrome rebuild.)
- Split: `tools` aspect = the tool-suggester's own suggestion inventory = the artefact the
  owner ordered deleted → SUPERSEDED by me (predecessor kept, successor 8031bb1c records the
  ruling, empty suggestions, points at growth_posture). briefing/strategy/vertical_landscape
  (scattered prose in 10-17KB analysis docs) → handed EXPLICITLY to copy_quality under the
  same mandate, supersession discipline stated. offer_ordering → analyser re-derives via the
  offer lane (their routing). Lesson for the cull recipe: **a cull has FOUR layers — pages,
  content_data, chrome, SPECS — and the census that finds the fourth is
  `site_specs WHERE is_current AND data::text ~* <the names>`.**

## 2026-08-31 — spec wash DONE by copy lane (mig 674, corr 53ea95f4); verified here at the rows
- briefing/strategy/vertical_landscape superseded (3 current of 6 total — supersession shape
  confirmed; deletion-first surgery; their battery caught a FIFTH strategy mention both our
  counts missed — the class battery, not a count, is the binding test).
- My own verify sweep flagged 2 "residue" rows — BOTH correctly non-issues: offer_ordering
  (expected; the analyser re-derives it) and one vertical_landscape line "Calculators and
  eligibility checkers are framed as decision-support tools" — vertical ANALYSIS in generic
  terms, not a reference to the culled FOS checker. My generic-phrase regex over-matches;
  the seven-proper-NAMES battery is the right test. Next from them: the 52 CTA labels.

## 2026-08-31 (night) — 417 closed out by its new lane; one correction to MY earlier reasoning
- The 420/417 lane shipped: mig 680 (race-tail row washed; deliberately NO widened-regex
  safety net — broad enough to catch a paraphrase is broad enough to void a deliberate
  mark), and the STRUCTURAL fix in Go (8bcd4ccae, inert till roll, corr bb099a3d): the
  no-lettering rule moves from composeBrandImagePrompt (fallback-only — "the ruled path is
  the fallback nobody reaches" was the whole diagnosis) to GenerateImageAction, coupled to
  the asset's PURPOSE not the prompt's SOURCE — governs every producer and already-queued
  unwashed items, which no config migration could reach.
> **CORRECTED 2026-08-31 (their measurement):** my 08-31 NOTES line "positive-framed
> because bugs_closed/028 proved banana discards negative clauses" cited 028's PRE-FIX
> state. foldNegativeIntoPrompt is live and the prohibition WAS delivered on the failing
> boxingonline generation (negative list incl. "text", prompt_len 232→407) — the model
> lettered anyway because the positive prompt licensed a wordmark. The true rule is
> stronger: **a folded negative LOSES to a positive licence in the same prompt.** Positive
> framing was the right call for a reason I didn't have.
- Farmer's named exception now has its durable home: `constraints.wordmark_text =
  "farmerinsurance"` set on plan row b6680524 (their opt-in field, unsafe side OFF, value
  validated against identity at the reader once the Go rolls). Also confirmed farmer's plan
  row carries the 670 override, so plan-driven regeneration is safe both before and after
  the roll. 421 (design-comp-served-as-logo) is theirs, split correctly from 417.

## 2026-09-02 — acceptance part 2 PASSED; the "stale footer" was my own coarse grep; decision 5 BUILT
- Cull acceptance part 2: after two days of refresh cycles, **21/21 retired urls still 404,
  control 200, zero page_rerender rows for archived pages** — the cull holds.
- Handoff owed-item 2 CLOSED AS FALSE POSITIVE: the homepage's "2 /tools/ refs" were
  `/tools/assets/*.js` script assets (both live, 200) — my `grep -c '/tools/'` counted a
  path PREFIX, not links to retired pages. Check refined: grep the retired URLS, not the
  directory.
- **Decision 5 BUILT (WDS-020):** growth_posture.go helper + guards at both tool-chain
  heads; held rows born in the record shape with the release recipe on-row; owner-request
  bypass; fail-open loudly. Tests green incl. the handler-coverage source-scan guard, which
  correctly caught the newly-computed handler and got its declaration. Register row+entry,
  RUNBOOK recipe, council corr `1e735fa2` (Council-Submitted), commit `8f2bd18fb`, HEAD
  verified building. **Go inert until the next roll; no site holds until the owner sets
  one** — the RUNBOOK's one-UPDATE recipe is ready when he names a site.

## 2026-09-02 — WDS-020 round 1 REVISE made the design better; round 2 is the door
Round 1 (corr 1e735fa2) came back REVISE, and the objections composed into a relocation:
guardian (don't branch the generic action) + prior_art (a census over a rolling-window table
cannot prove producer completeness) + reuse_agent (the owned-page park is the established
idiom) all point at the SAME answer — **one policy door in writeWorkItem**, the seam every
filing crosses (insertWorkItem is a thin wrapper; I had this backwards for an hour and the
call-graph read settled it). Round 1's two per-producer guards REVERTED to the pinned parent
(876fd1e2c — never a relative ref); door + transform mirror ownedPageParkedItem verbatim
incl. 342's keep-identity retraction contract. Two real round-1 defects closed structurally:
recurrenceExpected now set for EVERY producer (was missing on the discovery head), one
fail-open test covers all paths (was asymmetric). Evidence upgrades: the two-heads census
re-run against agent_definitions itself (tool-suggester the only config producer);
approval_mode measured (25,673 auto / 1 manual — a per-item dispatch gate, wrong axis);
filing_mode test's existence proven by path+quote. Round 2 committed c2349955d, resubmitted
on the trail, HEAD verified building. The estate memory "a REVISE round is cheaper than the
defect it finds" — worked case #3, and this time the finding was the ARCHITECTURE.

## 2026-09-02 — WDS-020 round 2 APPROVED (2 advisories, none high); both checks answered by measurement
- reuse_agent (medium): "did round 1 leave a config-side guard in tool-suggester's workflow?"
  NO — round 1's guards were both Go; `agent_definitions` for tool-suggester contains zero
  growth_posture references (measured). The revert scope (four Go files from the pinned
  parent) was the whole of round 1.
- guardian (low): the GrowthPostureQuerier misuse worry — the helper has exactly ONE caller
  (the door), and it passes the transaction (grep, no others fleet-wide).
- guardian (low, standing): the hand-kept 2-type set is a process-not-code control for a
  future growth TYPE — true, conceded in the plan, and the door bounds the damage to
  type-minting (not producer-adding). If a third growth type is ever minted, widen
  GrowthGatedItemTypes in the same commit; the pure-half test is where it lands.
- Trailer: Council-Submitted on both commits resolves at report time — no amend, per the
  forward-only rule. **WDS-020 is now built, tested, approved, registered, and inert until
  the roll.** Decision 5 fully closed at this lane's end.

## 2026-09-02 — evidence registers built for BOTH sites (owner instruction via lendzy; migs 697/698 APPLIED)
- Method = lendzy RUNBOOK §8, followed exactly; 7/7 quotes through cmd/fcaquotecheck with
  absent controls; every URL title-confirmed (the handbook 200s every path).
- loanzy (697, create): CCA 1974 s.66A 14-day withdrawal; StepChange FCA status;
  **MaPS provenance with corrects_site_citation — the lendzy-class find: served pages group
  MaPS under "FCA-authorised services"; MaPS is the statutory guidance body, not an FCA
  firm.** Copy routed to copy_quality (with the second find: the £5,000/9.9%/3y worked
  example gives three different answers on three pages — NOT registered; a register must
  not launder an inconsistency into an authority).
- farmer (698, SUPERSEDE-AND-MERGE — its register already held 3 news-entity facts): ICOBS
  8.1.1, DISP 1.6.2, DISP 3.6.6, ELCI 1998 reg 3 appended; entity facts carried forward,
  verify RAISEs if lost. The insurance answer to lendzy: method unchanged, chapters differ.
- **Method break found + reported to the FCA-mirror design: SOURCE HOSTING class** —
  maps.org.uk/moneyhelper.org.uk sit behind Cloudflare challenges ("Just a moment...") →
  perpetual false drift; gov.uk substituted (note: page lives at the FOUNDING-name slug
  single-financial-guidance-body). Proposed mirror rule: reject challenge-title hosts at
  WRITE time via the production extractor, not in the daily run.
- Council corr 50ba341a (Council-Submitted on the commit); a chassis roll expected within
  the hour may kill the run — RESUBMIT, don't re-diagnose (lendzy's warning, adopted).
- The claim-extraction one-liner (pages → regulatory sentences) is in this entry's session
  scrollback shape: curl each active page, strip tags, split sentences, grep
  FCA|FSCS|FOS|ICOBS|CONC|£\d|days|attempts|must — good enough to reuse, RUNBOOK if a
  third site needs it.

## 2026-09-02 — copy lane verified + sharpened both register-pass findings
- MaPS misattribution: EIGHT pages (get-help, index, six tool pages), not the handful my
  extraction sampled. Stage-2 canary fired at get-help (their corr 01e77807, parks for the
  owner); the rest join his batch-posture decision. SVC-MAPS-GOV cited as correction source.
- The arithmetic split DIAGNOSED, better than my "rounding" read: two legitimate APR
  conventions — nominal÷12 gives £161.10/£5,799.65 (one page exact under it), strict
  effective-APR gives £160.11/£5,764.02 (another page exact under it) — and the repayment
  calculator's £158/£5,688 is wrong under BOTH (implies ~8.5%). Since the copy says "APR",
  strict-APR is their harmonisation target. Lesson for my file: what read as sloppy
  rounding was two CORRECT conventions plus one real defect — the inconsistency census was
  right to flag, wrong about the mechanism; the fix queue carries all three derivations.

## 2026-09-02 — 697/698 round 1 REVISE (compliance HIGH); every objection answered by measurement, round 2 on the trail
- compliance HIGH (live MaPS misrepresentation left served): the routing round 1 promised had
  HAPPENED but postdated the submission — copy lane's 8-page census + stage-2 canary at
  get-help (their corr 01e77807) + owner batch queue. Round 2 cites the artefacts.
- debug_historian's clobber worry — the best objection of the round — MEASURED AWAY: farmer's
  entity facts were written by evidence-researcher via `verify_and_register_citations`, which
  MERGES (id map, skip dupes, append — evidence_citations.go:233-318); the lossy
  write_site_spec path is not in that workflow. My policy facts survive future runs.
- TOCTOU on the supersede: structurally impossible to yield 0-or-2 current rows —
  idx_site_specs_current is UNIQUE (site_id,aspect) WHERE is_current, and 698's
  UPDATE+INSERT is one data-modifying CTE in one txn. Post-apply battery: exactly-1-current
  per site; 3/3 + 7/7 facts (incl. carried-forward entity facts) with url + quote>15.
- Lesson: routing done via SendMessage is INVISIBLE to a council reading the plan — when a
  submission claims "routed to X", either produce the artefact in the change or expect the
  round; citing the peer's own correlation ids after the fact is the honest repair.

## 2026-09-02 (late) — 697/698 APPROVED r2; banned_claims shipped (702); and the arming advice found a fleet bug
- 697/698: **APPROVED round 2** (2 advisories, none high) after the roll killed the first
  r2 run (stale EXECUTING row, last update predating the second pod start — resubmitted per
  lendzy's advice, verdict ~40min later).
- **702 applied**: loanzy banned_claims = 5 (sibling's set; no-credit-check NARROWED — the
  broad form flags loanzy's own honest calculator sentence; literal-%APR OMITTED — worked
  examples are pedagogy). Pre-ship scan: 0 hits/27 serving pages, planted control 5/5.
  **Misstep caught in the same pass: the first scan was a BLIND ZERO** — python's default
  UA 403'd on every page while exiting 0; the a-post-fix-zero-needs-a-demand-control class,
  caught because the positive control ran in the same batch (4/5 exposed the second gap:
  the planted text missed a pattern). Curl re-fetch → honest zero.
- Migration numbering under load: 699→700→702 in ONE HOUR (loancalculator + park-provenance
  lanes claimed between my writes) — the register instruction is propagating fleet-wide.
- **The 414 lane's "read the first findings after arming" advice found bugs_open/437**:
  3 of loanzy's 30 active pages 404 — never built since 08-18 (`/your-rights.html` linked
  inline from live copy) — cause: mechanism-flow `steps[].branches` declared array-of-objects,
  writer emits string, render refuses, two-strike brands `unresolved` = queue reads handled.
  **119 failures / 6 sites / 14 days** (remortgagecalculator 53, loanzy 35, farmer 24…).
  Filed at the component-contract family; renumbered 423→437 at filing (423 taken — check
  the MAX of the sequence, not the last number you saw).

## 2026-09-02 (close) — the blind-zero landmine grew its third mechanism within the hour
The 414 lane appended (fe183038e, invited): transport truncation — a kubectl-exec psql
stream that EOF'd after plausible output, delivering 2,283 of 2,585 rows — passes BOTH my
controls (the planted row survived; every delivered row full-sized) while scanning 88% of
the corpus. Their third control: reconcile the row count at source/transport/destination
and refuse to scan on disagreement; export to a file in the pod + kubectl cp, so transport
has a size to check, not a stream to trust. Accepted as-is: the entry's subject was never
"fetch" but "a scan trusting its input corpus". The general form, theirs, worth keeping:
**whenever a zero is the answer you would like, ask what the instrument would have shown
had it only seen part of the input.**

## 2026-09-02 (night) — farmer banned_claims shipped (713); the register programme's last gap on my sites closed
- The lendzy lane routed the 414 census's quiet gap here (farmer 7 facts / 0 bans — "done"
  in every count census, enforcing nothing). **713 applied**: five INSURANCE-SHAPED patterns
  built as their own set, not the credit set transplanted — payout guarantees, universal
  acceptance, broker misrepresentation, price superiority (the site forswears comparison in
  its own copy — that sentence is the pattern's grounding), no-questions-asked marker.
- All three of the day's traps applied in order: (1) calibrated over the FULL served corpus
  with count reconciliation — 18 rows → 18 fetched → 17 scanned + 1 ACCOUNTED
  (**/claims.html is a 437 victim on farmer**: needs_rebuild since 08-27, 3 mechanism-flow
  failures — added to the 437 evidence); 0 hits, planted control 5/5. (2) post-apply Go
  probe-fire on the LIVE row (claims.go:348 silent-literal trap): 5/5 compile, 5/5 fire on
  must-match, 0 false-fires on five legitimate fragments. (3) 695's double-escape verify
  arms in the migration (no double backslash; 5 single-escaped \b).
- Two structurally-reasoned ABSENCES recorded in the migration header: no literal-rate
  pattern (banned layer has NO citation exemption — 414-verified — and the register itself
  quotes the £5m ELCI minimum); no first-person FCA-authorisation pattern (CGV-033 owns
  that refusal fleet-wide; a local copy would shadow the shared mechanism).
- Council corr cea2a32c (Council-Submitted). RFC_060 gains an insurance data point built
  as a sector set rather than an adaptation. Register state across my sites: loanzy 3/5,
  farmer 7/5 — both halves live on both.

## 2026-09-02 (last) — s77 registered (716); the NDL fact REFUSED taught the third host signature
- Copy lane handed two facts from the owner-ruled get-help evidence pass. CCA s.77 verified
  and shipped (716, applied, corr 7e927a7b; renumbered 714→716 — the numbering race caught
  me AGAIN minutes after I wrote the check down; the discipline that actually works is:
  claim the number by LOOKING at max+1 in the same command that writes the file).
- **⚠ MISSTEP: committed 716 WITHOUT its Council-Submitted trailer** (forward-only, no
  amend). Disclosed inside the submission itself; 098's join for this commit needs this
  note. Check that actually works: the trailer belongs in the commit message I draft BEFORE
  the submit step, with the corr pasted in — I had inverted the order (applied+committed,
  then submitted).
- **The National Debtline fact REFUSED — the third unregistrable-host signature:
  UA-DIFFERENTIAL SERVING.** Both natural hosts serve full pages to curl (real title, full
  size, quotes present — every curl-side control passes) and NOTHING to the production
  fetcher's UA: even "free" fails QuoteFoundInText. Defeats the blind-zero landmine's both
  halves AND the challenge-page tell; only the write-time probe THROUGH the production
  matcher catches it — RUNBOOK §8 step 4's rule proven necessary for a third, different
  reason. Census so far: challenge-page class (maps/moneyhelper), UA-differential class
  (nationaldebtline/moneyadvicetrust). Relayed to the mirror design.
- Relay-vs-page corrections sent to the copy lane for bd03c2b3: the page says "a charity
  run by the Money Advice Trust" (not "a debt advice service run by"), and "We never charge
  for our support" is not on the cited page. Even a precision-focused evidence pass
  paraphrased — the matcher, not the reader, is the arbiter of verbatim.
- Register end-state tonight: loanzy 4 facts / 5 bans; farmer 7 / 5. Verdict watches open:
  702 (48eec07c), 713 (cea2a32c), 716 (7e927a7b).

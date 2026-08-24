# 328 — a page that failed to build is still linked from the pages that did, so one blocked page becomes a visibly broken site

**Filed 2026-08-19** by the `loanzy_uk_example_site` lane, from a greenfield one-shot build.
**Status: OPEN, UNOWNED. Live on `loanzy.uk` today.**

> **On the 090 loop:** not run. This is not a hypothesis about a mechanism — it is two URLs,
> one serving and one 404, and the anchor between them, all read from the live site.

## The one-paragraph version

When a page fails to build, the build knows. The item is `needs_human_review`, the `pages` row
never reaches `deployed`, and the failure is recorded. **Nothing tells the pages that link to
it.** They are built, deployed and served with anchors pointing at a page that does not exist,
so a single blocked page turns into a live site with dead navigation — and the reader gets a
404 with no indication that anything is unfinished.

## Evidence (measured 2026-08-19, live)

`https://loanzy.uk/` serves 200 and contains:

| link on the live home page | target status |
|---|---|
| `/your-rights.html` | **404** — build refused at `validate_content` (`bugs_open/260`, 20 blockers) |
| `/guides/index.html` | **404** — builder no-op'd honestly: *"no sections ready to build"* |
| `/tools/loan-comparison-calculator/index.html` | 200, **but empty of its tool** (`bugs_open/311`) |

Two of the three defects behind those targets are already filed and owned. **This bug is about
the third fact: the links shipped anyway.** The home page was built AFTER `your-rights` had
already failed, so the information was available at build time and simply not used.

## Why it is its own defect and not part of 260/311

Fixing 260 removes this instance; it does not remove the class. **Any** page can fail to build —
truncation, a validation blocker, a missing component, an exhausted retry budget, an
infrastructure burst (`bugs_open/307`) — and every one of them produces the same shape: a live
site advertising a page that is not there. The route needs one rule, not one rule per cause.

It is also the difference between the two possible readings of a build. A site with four good
pages and no links to a fifth reads as *small*. The same site with two dead links reads as
*broken* — which is the judgement a customer makes about the product, not about the page.

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **Resolve links against `build_status` at deploy time.** A link whose target is not
   `deployed` is dropped (or rendered as plain text) by the deploy step. The bad state stops
   being representable because the anchor cannot outlive the missing page. ⚠ must consult
   `pages.build_status` **and** the artefact, given `bugs_open/315` (`deployed_at` is stamped
   whether or not the object was written).
2. **Hold the first publish until the corpus is complete**, then publish atomically. Strongest,
   and the most disruptive: it changes the route from incremental to all-or-nothing, and it
   would also have prevented the stale-nav window this build served for ~1 hour.
3. **File a visible item per dangling link and re-render the referring pages** once the target
   builds. Weakest — it is repair after publication, and it depends on the repair path
   actually running (see `bugs_open/313`, where `plan_links` has never run at all).
4. **At minimum, a post-deploy assertion**: after a build completes, fetch every internal link
   on every deployed page and fail the run if any 404s. This does not prevent the defect, but it
   converts "nobody noticed" into "the build reported it", which is what happened here only
   because a human went looking.

## How to verify a fix

Build a site in which one page is *forced* to fail (a deliberately invalid section is enough),
and require that no deployed page links to it — asserted **at the served HTML**, not at the
plan, the sections, or `page_components`. Positive control in the same run: a page that DID
build must still be linked, or the fix has simply stopped emitting internal links (which
`bugs_open/313` shows is a state this platform can reach and not notice).

---

## CONTRIB 2026-08-21 — from the `mortgagecalculator_couk_adoption` lane: the detector already exists, so candidate 3 is much cheaper than it looks (and candidate 4 already half-runs)

**Second site, same shape, 25 days older.** `mortgagecalculator.co.uk` has carried this defect
since adoption, so this is not a greenfield-build artefact — it is what the shape looks like once
it has had a month to settle. Measured live 2026-08-21 ~10:20Z.

### The instance

One page cannot build (`scorecard-simulator`, refused by `bugs_closed/260`'s template leak).
**Six deployed pages each serve exactly one anchor to it.** Full audit, not a sample: all 29
`build_status='deployed'` pages fetched, every non-absolute `href` extracted, deduped to 33
distinct targets, each resolved once cache-busted.

| measure | value |
|---|---|
| deployed pages fetched | 29 |
| raw internal hrefs | 1,030 |
| distinct internal targets | 33 |
| targets returning 200 | **32** |
| targets returning 404 | **1** — `/scorecard-simulator.html`, 404 on 3 cache-busted repeats |
| deployed pages serving an anchor to it | **6** — `/disclaimer.html`, `/guides/first-time-buyer/`, `/guides/how-banks-decide/`, `/guides/lender-restrictions/`, `/guides/mortgage-scorecard/`, `/tools/affordability/` |

### The correction that matters: **the platform is NOT blind to this**

§"The one-paragraph version" says *"Nothing tells the pages that link to it."* On this site that
is **false in a way that changes the fix.** There are **seven `unbuilt_internal_link` work items**
on the blocked page's `page_id`, one per linking component, each naming the linking page, the
component function, and quoting the href verbatim:

```
unbuilt_internal_link in page_component (guide-mortgage-scorecard:generic-text-block):
  href "/scorecard-simulator.html" points at a page that has never been deployed
```

`page_id` on the item is the **target** page; the linking page is named inside the summary text.

**So detection exists, is per-linking-page, and is accurate.** What is missing is the route: all
seven sit at `status='needs_human_review'` with no handler, which is `bugs_open/033`'s "no working
surface" queue. The information is produced correctly and then parked unread.

That reprices the candidate list in §"Fix candidates":

- **Candidate 3 is not the weakest option on this evidence — it is the one that is already 80%
  built.** It does not need a new detector; it needs the existing item type wired to a handler
  that re-renders the referring page (or drops the anchor) once the target reaches `deployed`.
- **Candidate 4 is also already half-running**, for the same reason — the finding exists, it just
  never reaches anyone. "Convert nobody-noticed into the-build-reported-it" is a delivery problem
  here, not a detection one.
- Candidates 1 and 2 stand unchanged and remain the only ones that make the state
  **unrepresentable**; this contrib is about cost, not about preferring repair to prevention.

### A second finding: a parked item is not evidence its condition survives

**Six of the seven items match the live site. The seventh is stale.** `8a230338` names
`contact-index`, and the href is in neither the served page nor the stored content:

| check | result |
|---|---|
| `curl /contact/index.html \| grep -c scorecard-simulator` | **0** |
| `page_components.content_data LIKE '%scorecard-simulator.html%'`, this site | **6 pages — `contact-index` absent** |

Something repaired that link and nothing closed the item. Whoever builds the handler needs this:
**it must re-check the condition at handling time, not trust the summary**, or it will re-render a
page to fix a link that is not there. [UNVERIFIED] whether this staleness is general — measured on
this site only, n=1 of 7.

⚠ **And the stored/served split bites here.** `contact-index` is `build_status='needs_rebuild'`
with `deployed_at` NULL, yet serves a healthy 1,267-word page from its previous build. A `pages`
row that does not say `deployed` is **not** evidence the URL is dead — relevant to candidate 1,
which proposes gating anchors on `build_status`: on this site that predicate alone would have
dropped every anchor pointing at a perfectly healthy served page. Candidate 1's own ⚠ about
`bugs_open/315` is the same hazard from the other direction; the honest version of candidate 1
consults the artefact, and `build_status` is at best a cheap pre-filter.

### The self-fuelling property, which is what makes this worse than a static count

Every page the framework writes here **adds another anchor to the page it cannot build**, because
the site's `design_intent` names `Scorecard Simulator` as a delivered feature. Measured over this
lane's own work: while three unrelated dead links were being fixed on 08-16 by building the pages
they pointed at, live instances of *this* href went **4 → 6** — because the two pages the framework
had just built each linked to it unprompted. The count is not stable; it grows with productivity.
Three separate work items on this site (`2dca1a09`, `d5a9ae7d`, `0c65f9fa`) exist purely because
different detectors each noticed the brief promises a page that is not there.

**That is the argument for candidate 1 or 2 over 3 or 4**: repair-after-publication has to keep
pace with a producer that never stops, and on this site the producer is the brief itself.

### Cross-reference

`bugs_closed/260` closed 2026-08-20 and its fix is proven aboard `agent-chassis` v1.0.1321
(binary probe, four controls). **It does not close this instance** — 260's own closure note says
so explicitly: the fix makes a mistyped field fail early and named, it does not make the parked
content build. This lane re-fired the build on 08-21 to find out which. Whatever the outcome, the
six anchors were served for 25 days regardless, which is this bug's point and not 260's.

Lane record: `docs/agent_docs/docs024_key_docs_latest/mortgagecalculator_couk_adoption/NOTES_mortgagecalculator_couk.md`,
`## 2026-08-21`.

---

## FIX BUILT 2026-08-23 — and three corrections to the account above, all measured

**Status: fix committed, Go INERT until the next fleet roll; the config half held as
`sql_for_agents/575_enable_suppress_unshipped_links_HOLD.sql`. Council SUBMITTED, corr
`21c19c1f-e614-49bd-82ac-0bb5b58082e0`, verdict not yet read. Register entry LNK-038.
Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_328_links_to_unbuilt_pages/`.**

**Still live when the work started**, re-measured cache-busted 2026-08-23: `https://loanzy.uk/`
serves 200 carrying `href="/your-rights.html"` ×2 and `href="/guides/index.html"` ×1, both 404.
And it **reproduces on a build that ran the same day** — loanzy redeployed all 17 pages
13:13–14:24Z, the home page at 13:28:27Z, while both targets had sat `deployed_at IS NULL` and
untouched since 08-18 with no open work item. So this is not a stale artefact from the filing date;
the binary and config in production that afternoon still did it.

### CORRECTION 1 — the handler ran, and its one remedy cannot work here

The 08-21 contrib says the seven items "sit at `status='needs_human_review'` with no handler",
i.e. that detection works and **delivery** is the gap. The rows say otherwise. All **58** parked
items carry `triaged_at`, `handler_agent='page-build-handler'` and `attempt_count ≥ 1`:

| status | error | count |
|---|---|---|
| needs_human_review | `page-build-handler no-op: no sections ready to build` | **48** |
| needs_human_review | `content validation failed: N blockers` | **10** |

They were dispatched, the handler ran, and it failed — because the item type implements exactly
**one** remedy, *build the target page*, and the target is parked precisely because it cannot be
built. `bugs_open/220`'s routing is not the gap; it works. **The missing remedy is the other one
the item's own `fix` text names: stop the referrer advertising it.** That reprices the candidate
list again — candidate 3 is not "80% built", it is *built and inert*, and the 08-21 contrib's cost
argument rested on the wrong half.

### CORRECTION 2 — detection did NOT cover this bug's own headline instance

There is **no `unbuilt_internal_link` item anywhere** for loanzy's `/your-rights.html` or
`/guides/index.html` — open, cancelled or otherwise. The site's only three name a different page's
component (`about:about-content`), all filed 08-18. The audit is periodic and reads deployed HTML;
loanzy's last discovery pass predates its redeploy. So "detection exists, is per-linking-page, and
is accurate" is true of mortgagecalculator and **silent** about loanzy — and any fix routed through
the item queue would miss the youngest and worst case. This is the third argument for candidate 1
over 3 and 4, on top of the two the file already makes.

### CORRECTION 3 — the open-item count overstates live harm by about 3×

**42 of the 63 open items name targets that serve HTTP 200 today**: all 40 lendzy items (3 tool
pages), gaswholesalers' `/fuel-pricing-framework.html` — the canonical 404 quoted in `links.go`'s
own comment, which has since shipped — and mortgagecalculator's `/contact/index.html`. They are
stale records held open by a missing `deployed_at` stamp (the `bugs_open/315` family), not live
damage. **A work item records what a detector saw; it is not evidence about the wire.**

The honest measure is at the artefact. Every `<a href>` in every stored
`page_components.rendered_html` fleet-wide, resolved against the fix's predicate:

| target class | anchor hits | referring pages | distinct targets |
|---|---|---|---|
| servable — untouched | **3,193** | 577 | 557 |
| unservable — suppressed | **36** | 24 | **14** |

**1.1% of internal anchors, every one of them a live 404** — and it names damage this file's census
never reached: `remortgagecalculator.uk` (2 targets, 6 hits), `webdesign.co.uk`, both
`pool-*.internal`. 14 targets on 9 sites against the queue's 13 on 7.

### The fix — candidate 1, at the seam, with the predicate the estate did not yet have

`datahelpers.RepairPageLinks` already unlinks dead in-body links at four seams, and all four load
their index from ONE function, `loadValidPagePaths` (`validate_page_content.go:1515`), whose query
has **no build-axis arm at all**. A `pages` row that has never existed on the web is a perfectly
good link target to every one of them. The same omission was fixed on three sibling loaders on
2026-08-09; this loader was missed.

⚠ **Candidate 1's own warning was right, and both existing predicates are wrong in opposite
directions.** Measured fleet-wide against live HTTP on 2026-08-23, cache-busted, **with a
per-domain 404 control**: `NeverDeployedPagePredicate` selects **9 pages that return 200**;
`PageMayBeLinkedPredicateFor`'s `planned`-only floor still holds (17/17 404) but **misses the 3
`needs_rebuild` rows never built at all, 3/3 returning 404** — one of which is `/your-rights.html`.
The discriminator is the rendered-component count: **20 never-shipped pages with zero components →
20/20 404; 9 with ≥1 → 9/9 200.** The conjunction is load-bearing — 8 pages have `deployed_at` set
and zero components (tool/blog-index pages served by another subsystem) and a component test alone
would delist all eight.

⚠ **The control is the finding.** Uncontrolled, the same census read *"19 `planned` pages serve
200"* — a refutation of the whole approach. All 19 were one parked domain returning 200 with a
114-byte registrar redirect for every path, including a URL invented at the prompt. Logged in
`WRONG_CALLS.md`; the control is now a LANDMINE.

**What was built** (`LNK-038` carries the full account):

- `PageLinkRefusedPredicateFor` — a fourth member of the predicate family, never a second spelling.
- Suppression **at the two OUTBOUND seams only** — `repairOutboundPageLinks` (both rerender paths
  *and* the initial build, since `deploy_page` calls the page-rerender agent by role) and
  `AssemblePageAction` (the loop paths, which called **neither** repair function). Nothing writes
  to `content_data`: **the authored href survives, so the anchor returns by itself when the target
  ships** — no cascade, no repair queue. It also means this **silences no detector**: the stored
  `rendered_html` still holds the anchor for `check_phantom_internal_links` to find.
- **Two arms**, from reading all 36 anchors rather than assuming: 28 classless prose anchors unlink
  (keep the words); 8 classed template controls are dropped whole (owner's decision, 2026-08-23),
  because unlinking those would leave "Read your rights →" as bare text inside the card — a
  standing landmine this would otherwise have multiplied by eight.
- Opt-in field `suppress_unshipped_links`, **default OFF** (RFC_010 §2), with the chrome policy's
  two degrade escapes: failed lookup, and **zero shipped pages** (the first-build case).
- `unbuilt_internal_link` registered in `reviewRevalidators` — the type has **72 items in its whole
  history and ZERO ever closed**, because it is born `needs_human_review`, which
  `CompleteWorkItemAction` refuses to leave, so its registered verifier could never run. It
  delegates to that same verifier, and will honestly report `still_holds` on the ~42 items above
  rather than closing them on a stamping gap it did not fix.

**This does not close the bug.** A renderer fix is inert until something re-renders, and 36 anchors
on 24 pages are serving now. Owed after the roll: apply 575, fire `page_rerender` for those 24
pages, then verify **at the served bytes** — `href="/your-rights.html"` gone **while
`href="/calculators.html"` is still there**, because without the positive control "stopped emitting
internal links" passes the test.

---

## 2026-08-24 — LIVE. Go proven on `v1.0.1334`, migration 575 applied, and the `049` reconciliation the council asked for

**The fix is live.** Chassis `v1.0.1334` (pods 15:39Z), binary-probed on **both** replicas with a
control pair in the same run — `suppressUnshippedOutboundLinks`, `PageLinkRefusedPredicateFor` and
`CONTENT_LINK_SUPPRESSED_UNSHIPPED` all PRESENT; `repairOutboundPageLinks` PRESENT as the positive
control; an invented symbol absent as the negative. Migration `575` applied by hand at 16:07Z (5
snapshots taken first, `$post$` passed, keys read back with an independent query), recorded
`record-only`. Council **APPROVED at round 4**, corr `21c19c1f-e614-49bd-82ac-0bb5b58082e0`.

### ⚠ THE CENSUS IN THIS FILE IS ALREADY STALE — it grew overnight

| | 2026-08-23 | 2026-08-24 |
|---|---|---|
| dead anchors | 36 | **48** |
| referring pages | 24 | **28** |
| unservable targets | 14 | **16** |

**+12 in a day**, including a site that did not exist in the first census — `garden-tools.uk`,
arriving with **9 dead anchors across 4 pages**. This is §"The self-fuelling property" measured on
its own terms, and it is the reason the fix is at the renderer rather than in a repair queue:
**repair has to keep pace with a producer that never stops.**

### The `049` reconciliation (council round 4, `bug_historian`, advisory)

The seat noted that `bugs_closed/049` — *"live 404 links from stale chrome and unbuilt pages"* — is
titled almost identically to this bug's premise, and that this lane cited it without ever saying
whether its closed fix covers or conflicts with this one. Read now:

**They are the same harm on two different surfaces, and they do not overlap.** 049 is the
**chrome/nav** spelling: its fix is `applyNavVisibility` + `loadFetchablePageSet`, it deactivated
nav items and cleared `in_header`/`in_footer` flags pointing at never-built pages, and it went live
in `v1.0.1171` on 2026-07-26. 328 is the **page-body** spelling — in-body anchors written into
`content_data` by the writer, which no nav predicate has ever governed. Neither fix touches the
other's surface, and 049's closure is precisely why chrome is out of scope here (see LNK-030, its
successor).

Worth carrying across: **049's own correction records that its first pod-grep marker was VACUOUS** —
`NavFetchableOnly`, a typed constant Go resolves at compile time, so the grep returns 0 whether or
not the fix shipped. Today's probe avoided that class by using function names and string literals
with both controls. The trap 049 paid for is still live, and still worth the ten seconds.

### What is owed before this can close

The bug is **not closed**: a renderer fix is inert until something re-renders, and 48 anchors on 28
pages are serving now. **26 of those 28 pages re-rendered today**, all of them *before* the flag went
live at 16:07Z — so the next natural render is the test, and the fleet's own cadence should carry
almost all of it within a day. A 28-page dispatch was deliberately NOT fired: a re-render carries
every platform change since a page last rendered, not just this one, which is real risk on customer
sites for no gain where the page is about to re-render anyway.

Canary queued instead — loanzy.uk `index`, this bug's headline instance (item `b18a0287`). The
acceptance test is at the **served bytes with a positive control in the same fetch**: `2 ×
href="/your-rights.html"` and `1 × href="/guides/index.html"` must be **gone**, and `5 ×
href="/calculators.html"` must **remain**. Without that second half, a fix that stopped emitting
internal links altogether would pass.

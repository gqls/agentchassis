# NOTES — improvement loop

Append-only, newest at the bottom. Technical log: evidence, commands, what the system
actually said, and every misstep.

---

## 2026-09-02 — lane opened, first state assessment

**(a) Started from the wrong premise and caught it in one query.** Auto-memory and
`bugfix_136`'s file both carry the owner ruling of 2026-07-29: *the improvement loop is
stopped DELIBERATELY … do not re-enable it*, evidenced by `improvement-sweep`
`enabled=f` since 2026-05-02. I nearly wrote a plan around a stopped pipeline. The live
row says `enabled = t`, last triggered 2026-09-02 11:59:27Z. Migration
`389_park_contrast_failures_and_reenable_improvement_sweep.sql` turned it back on.
**The cheap check that caught it was the first query I ran** — one `SELECT` against
`scheduled_tasks`. Recorded because the ruling is five weeks old and is still being
cited as current by at least two documents.

**(b) `execution_path` is empty on every improvement-loop row.** My first attempt to
separate "audit ran" from "audit skipped" was
`execution_path::text LIKE '%discovery%'`, which returned `f` for all 98 rows — i.e. it
told me the loop had never run a discovery step, on a pipeline I had just watched
complete. The column is `[]`. `collected_data`'s keys are the real record. A false
negative that looks like a finding: exactly the shape to distrust.

**(c) The live workflow does not match `004_improvement_loop.md`.** 004 documents a
3-pass audit cap; the live `collected_data` carries `check_audit_due`,
`check_not_converging`, `load_audit_state` and a site fingerprint. Migration `291`
replaced the cap with a convergence gate on 2026-08-02 and 004 was never updated.

**(d) The measurement.** `[MEASURED 2026-09-02]`, 2-day window (orchestration retention
is ~1 day, so this is the whole record): 98 runs, 32 domains, fair rotation.
80 `complete_clean` / 15 `complete` / 2 failed. `audit_due=true` on 24 — the gate
discriminates. 136 items promoted. **Every single run reported `not_promotable > 0`.**

**(e) My first backlog figure was wrong, by 2.8×, and the marker would not have caught
it.** I summed `not_promotable` across runs and got 3,866. That number is meaningless:
`not_promotable` is a per-run count of the site's *standing* pile, so a site visited
five times contributes its pile five times. The standing figure, counted as rows, is
**1,385**. Both would have carried a `[MEASURED 2026-09-02]` marker honestly. The
marker rule says a figure must be disconfirmable — a sum over a rolling re-count could
not have come out any other way.

**(f) The backlog is real and it is growing.** 1,385 `detected` rows with no handler,
31 sites, oldest 2026-07-26. The `bugfix_284` lane recorded 722 of this class on
2026-08-19. Near enough doubled in a fortnight.

**(g) Confirmed there is no consumer, three ways.** (1) `detected-item-promoter`'s live
`pre_query` states it: *"Flag-only rows (no handler_agent) are NOT here"*. (2) A grep
for `handler_agent IS NULL` / `handler_agent = ''` across `*.go`, `*.sql`, `*.tsx`,
`*.py` returns only migrations, one-off repair SQL and table DDL — no agent, no report,
no dashboard. (3) The peer state exists and is populated: **912** flag-only rows sit at
`needs_human_review`, which IS looked at. So the class splits across two states and only
one is visible.

**(h) Then I probed the rows instead of counting them, and the picture changed.**

- 978 of the 1,385 are `head_essentials_missing`. Broken down by what is actually
  missing: **867 are a skip link alone**, 55 skip-link + footer, 56 all three. So the
  dominant finding is ONE fleet-wide template omission filed 867 times per-page.

- The 56 "all three" rows are on two domains. I curled them, with an invented-URL 404
  control on the same domain (control: 9 bytes, no title — the probe discriminates):

  - **farmerinsurance.uk (36 rows): the claim is two-thirds FALSE.** `/about.html`
    returns 200, 66,108 bytes, `<title>About | Farmer Insurance UK</title>`, one
    `<footer>`. `/blog/crop-insurance.html` likewise. Only the skip link is absent.

  - **boxingonline.com (20 rows): true, but not about our pages.** `/`, `/about.html`
    and `/index.html` all return 200 with the same 114 bytes:
    `<!DOCTYPE html><html><head><script>window.onload=function(){window.location.href="/lander"}</script></head></html>`.
    The domain is parked. The finding is a true statement about a lander and a
    misleading one about a page of ours.

**(i) The staleness mechanism, read in code rather than inferred.** `insertWorkItem`
(`platform/orchestration/actions/load_work_item_actions.go:1787`) writes with
`dropOnConflict`, so a re-run of the check drops the fresh row and the original
`spec.missing` stands. `HeadEssentialsMissingCheck` only emits a `ResolvedFinding` when
`len(missing) == 0` (`check_site_structural_validity.go:1116`). Since the skip link is
never present, a farmerinsurance row can never be retracted and never be refreshed — it
carries its first-ever missing-list indefinitely. **Consequence for anyone building a
consumer over this pile: `spec.missing` is a claim of unknown age.**

**(j) What I have NOT established, marked as such.** `[UNMEASURED]` whether the skip
link is genuinely absent from the chrome template fleet-wide, or absent only from these
26 sites' rendered output. `[UNMEASURED]` whether the other ten item types in the pile
carry the same staleness — I checked the mechanism, which is shared, but I have probed
served pages only for `head_essentials_missing`. `[ASSUMED]` that boxingonline.com's
parking is deliberate; that is decision D2 for the owner, not a fact I hold.

---

## 2026-09-02, later — the skip link has ONE cause, and it is provable

Followed plan item 3's question — is the skip link absent from the chrome fleet-wide, or
only from these sites' rendered output? `[MEASURED 2026-09-02]`

**(k) The estate shares one header component, and it has no skip link.** Of 34 active
sites, 33 carry a `header` slot in `site_components`. **32 of them point at the same
`component_id` `58fde68f-9190-4e5e-b6a5-ea21cf27a9af`** (three of those are forks:
`idea.uk` `f420f3fa…`, `leopardessconsulting` `990b7162…`, `webdesign.co.uk`
`ad6033ae…`). **Not one of them renders a skip link.**

**(l) The single exception is the one with no shared component at all.**
`loanandmortgagecalculator.co.uk` has `component_id = NULL` — a hand-owned header,
updated 2026-08-05 — and it is the only site on the estate whose served page carries
`<a class="skip-link" href="#content">Skip to content</a>`. So this is **not a missing
capability**: the platform renders a correct skip link today, on exactly one site, and
that site is the one not using the shared component. ⚠ That header belongs to the LMC
lane; this lane must not edit it.

**(m) The fix has a prerequisite, and skipping it would manufacture a new finding on
every page of the estate.** A skip link needs a target. LMC's points at `#content` and
its pages carry `id="content"` (2 occurrences on the front page). The shared-header
sites do NOT: probed finetuning.uk, webdesign.co.uk and cookly.uk — all three render a
`<main>` element and **zero** `id="content"` and **zero** `id="main"`.

  So adding the link to the shared header alone would produce ~1,000 dangling fragment
  links — and `check_phantom_internal_links_fragments.go` exists and would file every
  one of them. **The fix is two components, not one**: the skip link in the shared
  header AND an id on the page shell's `<main>`. That check's own header comment already
  names the contract — *"header skip-link targets id='content', which its pages carry"* —
  so the shape is settled; it is the shared sites that are missing half of it.

**(n) A second parked domain, found by accident, and NOT flagged.**
`adversecreditmortgage.co.uk` serves the same 114-byte lander stub as boxingonline.com
on `/`. It has **19 pages recorded active** and is one of only two active sites with
**no** `head_essentials_missing` finding at all. So the check that would have caught it
did not fire there. `[UNMEASURED]` why — I have not read the check's page-eligibility
gate. Worth doing before anyone treats "no finding" as "no problem": on this evidence,
the two sites with the cleanest record are a parked lander and the site nobody flagged.

**(o) Checked the fan-out mechanism before assuming the fix could ship, and it is
sound.** `bugs_open/404` says `template_changed` was in the live re-render vocabulary
while `create_rerender_items_action.go` knew neither it nor `literal_markdown` — which
would have made a chrome fix complete green and ship nothing. That is **repaired in the
tree**: the file now derives its vocabulary from `livespec.RerenderSectionReasons`, the
single definition, asserted daily against the live gate by `config-key-audit
--live-declaration-drift`. `[UNMEASURED]` whether that repair is live in the running
image — to check at the time of the fan-out, not now.

**(p) Where this leaves the backlog arithmetic.** Of the 1,385:
- **867** are the shared-header skip link — one two-component fix.
- **~110** more (skip_link+footer, and the 56 all-three) are the same fix plus something
  else; the farmerinsurance 36 are the same fix plus a stale spec.
- **20** are boxingonline being parked, and belong to whoever owns that domain, not here.
- **~390** are the other ten item types, unexamined by me so far.

So the honest size of "findings that need a human decision" is **nowhere near 1,385**,
and a screen showing that number would have been the wrong instrument. This is why plan
item 4 sits behind items 2 and 3.

---

## 2026-09-02, evening — the pointing census (owner asked for the list)

**(q) Probed all 34 active domains rather than the 2 I had stumbled on.** Script kept at
`improvement_loop/probe_serving.sh`. `[MEASURED 2026-09-02]` **31 SERVING, 2 PARKED,
1 soft-404.** The two parked are `boxingonline.com` and `adversecreditmortgage.co.uk` —
exactly the two the head-essentials evidence had pointed at, now established by census
rather than by anecdote.

**(r) The cause is DELEGATION, and I nearly reported the wrong fix.** My first instinct
was "the A record is wrong". `dig NS` says otherwise:

| domain | NS | A |
|---|---|---|
| boxingonline.com | `ns1/ns2.afternic.com` | 13.248.169.48, 76.223.54.146 |
| adversecreditmortgage.co.uk | `ns1/ns2.dan.com` | 76.223.54.146, 13.248.169.48 |
| cookly.uk, farmerinsurance.uk, agritec.uk, webdesign.uk (all serving) | `alexis/leah.ns.cloudflare.com` | 104.21.x / 172.67.x |

Both parked domains are still authoritative at the **marketplace** that sold them
(Afternic and Dan.com are the same operator, hence identical parking IPs). **An A record
set at the registrar would have done nothing**, because the domain is not delegated to
where that record lives. Recorded because "point the domain" and "change the A record"
are not the same instruction and I would have given the second one.

**(s) noted.co.uk is a soft 404, not a pointing problem.** Root and an invented path
returned byte-identical 75,546 bytes, which is the same signature as a parked domain.
It is not one: `/privacy.html` (67,008 b) and `/about.html` (66,659 b) return their own
distinct titles. Only *unknown* paths fall back to the home page. **The signature that
identifies a parked domain also identifies a soft 404** — a second fetch of a page that
should exist is what separates them, and the first version of my script would have
called this one parked. Handed to the `noted_rebuild` lane, not fixed here.

**(t) The count of active pages behind the two parked domains: 40** (21 + 19). Built,
deployed, reaching nobody.

---

## 2026-09-02, evening — D1 executed: the skip link ships in the page shell

Owner answered D1 "yes, fix the chrome" and D2 "not deliberate, investigate". This is D1.
Committed `d01fb092a`, council `3c71ec77-fd15-4aa1-a762-cc36116caca5` (submitted, verdict
not yet read).

**(u) Where the fix went, and why not the header.** `assemblePage`
(`rerender_single_page_action.go`) is the one function that builds every served page
shell — 2 callers, `rerender_single_page_action.go:163` and
`section_editor_actions.go:655`. The header component was the obvious home and is the
wrong one: it would need the same edit in four places today (the shared component plus
three forks) and in every future fork, and it would separate the link from the target it
depends on. In the shell they are emitted three lines apart and a fork cannot lose one
without the other.

**(v) The CSS travels in the head. This was the decision with the real consequences.**
The alternative — put `.skip-link` rules in `styles.css`, which `webdesign-agent` renders
per site — makes correctness depend on the CSS wave landing before the page wave, and
until it did, **31 live client sites would show a stray "Skip to content" above every
page**. An ordering an operator must remember is a defect. `injectSkipLinkCSS` follows
the existing `injectComponentCSS` shape, so there is no new pattern to learn and no
ordering to get wrong.

**(w) MUTATION FOUND A REAL HOLE IN MY OWN TESTS — this is the entry worth reading.**
I wrote five tests, all passing, then mutated the production code three ways to check
they could fail:

| mutation | result |
|---|---|
| drop the skip-link anchor | FAIL ✓ |
| drop the target span | FAIL ✓ |
| **drop `head = injectSkipLinkCSS(head)`** | **PASS ✗** |

Nothing asserted the CSS reached the assembled page. With that line deleted every test in
the file stayed green **and every page on the estate would have rendered a visible "Skip
to content" above its header** — the exact outcome the design was chosen to prevent,
undetectable by the suite written to protect it. Added the assertion, re-mutated, it now
fails. **The test I would not have written is the one guarding the risk I had already
identified and written three paragraphs about.** Identifying a hazard in prose is not the
same as asserting on it, and only mutation told the two apart.

**(x) `[UNMEASURED]`, and it is stated in the submission's risks rather than discovered
later:** `AssemblePageAction` (`multipage_actions.go`) is a second build path that does
not read this function's head injections — the same gap already documented for
`injectRobotsNoindex`, `injectCanonicalLink` and `injectPageJSONLD`. Pages built only
through that path will not carry the skip link. The 14-page sample says the estate does
not serve such pages (13 of 14 carry exactly one `<main>`; the 14th is advertise.co.uk's
hand-built index, outside the flagged set and already carrying its own skip link), but a
sample is not a census and I have not run one.

**(y) Two things this does NOT do, so nobody reads the commit as the end of it.** A Go
change is inert until the next fleet image rolls, and then inert on each page until that
page re-renders. **The fan-out is the expensive half and it is not started.** Nothing in
the 867 rows retracts itself either: `head_essentials_missing` only emits a
`ResolvedFinding` when all three essentials are present, so the rows clear as each page
re-renders and is re-probed — which is the right behaviour and also means **the backlog
will not visibly move until the wave runs.** Do not read a flat count as the fix failing.

---

## 2026-09-02, evening — I got the pointing claim wrong, and caught it before the owner acted

**(z) WRONG CALL, corrected within the hour.** I told the owner "40 pages of finished work
are built and waiting behind a delegation nobody changed". `[MEASURED 2026-09-02]`:

| domain | pages | build_status | ever deployed |
|---|---|---|---|
| boxingonline.com | 21 | 20 `deployed` + 1 `planned` | **yes — latest 2026-09-02 13:59:46Z** |
| adversecreditmortgage.co.uk | 19 | 18 `planned` + 1 `needs_rebuild` | **no — zero, ever** |

So it is 21 built pages and 19 that do not exist. **Pointing adversecreditmortgage today
would show nothing**, and the natural reading of that would be that the pointing failed.

**What I did wrong, precisely:** I took the page counts from `pages.status='active'` and
read them as "pages that exist". `status='active'` says a page is *wanted*; `deployed_at`
and `build_status` say whether one was ever *made*. I had both columns available in the
first query and grouped on neither. **A count of rows is not a count of artefacts** — the
same shape as this lane's other lesson, that a finding is not a defect.

**(aa) And it RETRACTS my own §(n).** I recorded adversecreditmortgage's zero findings as
"a detector gap, not a clean bill of health". It is neither. `loadStructuralPopulation`
gates the population on `PageHasShippedPredicateFor` (`links.go:293`), and no page on that
site has ever shipped — so a check that probes served pages had nothing to probe. **The
check was right and my suspicion was wrong.** Recorded rather than quietly deleted,
because §(n) was written in the confident voice and would have sent the next reader
hunting a defect that does not exist.

What survives from §(n) is the weaker and still-true half: *"no finding" is not "no
problem"* — here the absence of findings meant an unbuilt site, which is a bigger problem
than the findings would have been.

---

## 2026-09-02, evening — plan item 4's evidence, and it is the code's own words

Read the routing rationale of all eleven checks that file into the flag-only pile, rather
than reasoning about the pile from its shape. `[MEASURED 2026-09-02]`

**(bb) Every one is flag-only ON PURPOSE, and several say "not yet" rather than "never".**
The comments are explicit: `heading_promise_unmet` — *"the repair is a planner/writer
judgement"*; `structure_floor_unmet` — *"the seat RECORDS; the refusal is a planner/human
verdict"*; `page_content_divergence` — *"per D5 (no handler agent in v1)"*;
`check_site_structural_validity`'s five — *"FLAG-ONLY, **THIS PASS** … No auto-repair
agent is wired in this pass"*, with `canonical_mismatch`'s repair named as future work
gated on `bugs_open/251`. So the pile is not one settled category. It is at least two:
findings that genuinely need a human judgement, and findings whose handler was deferred
to a later ship that has not come.

**(cc) The authors KNEW nobody drains it, and quoted the bug that says so.**
`check_archived_page_still_serving.go:104`, verbatim:

> *The cost of flag-only is real and named in bugs_open/083: "a detector whose output
> nobody drains is not neutral — it is actively misleading". The item lands at `detected`
> where image_url_404, asset_reference_404 and stylesheet_gutted land, and is deliberately
> not dispatchable — detected-item-promoter's `handler_ok` door requires an
> agent_definitions row matching handler_agent, which `''` cannot satisfy.*

This is the strongest thing I have found on this lane, and it is not my inference: the
platform's own code states the defect, cites the bug that names it, and ships into it
anyway. **That reframes plan item 4.** It is not "nobody noticed". It is a known cost,
accepted eleven times, by authors who each had a good local reason and no shared surface
to file against.

**(dd) And exactly one of the eleven supplied what a person would need.**

| | rows | carrying a `triage_hint` |
|---|---|---|
| `archived_page_still_serving` | 9 | **9** |
| the other ten types | 1,376 | **0** |

The one hint is excellent — it names the remedy action, the `agent_error_log` codes to
read first, the opposite-direction case (un-archive rather than delete) and the ruling
that forbids auto-deleting. That is a person's brief. The other 1,376 findings arrive with
a summary and no stated remedy, so even a perfect queue would show a reader 1,376 problems
and one answer.

**(ee) What this means for the design, recorded before I build anything.** "Show the pile
to a human" is the wrong first move — it was my instinct this morning and the evidence has
moved me off it. The pile's largest member turned out to be one template fix (867, now
committed), its second-largest a set of deferred handlers, and only a residue is genuinely
a judgement call. **The question to put to the council is not "where do we show these" but
"which of these eleven idioms is a deferred handler and which is a human decision" — and
the second group needs the `triage_hint` discipline the ninth check already demonstrates.**
That is a shared-seam change across eleven producers, so it goes through the gate on its
own, after the skip-link wave has drained the population enough to see what is left.

**(ff) Verified my own prediction rather than leaving it as one.** I claimed in §(y) that
the 867 rows would retract themselves once pages re-render. Checked it in code instead of
hoping: `HeadEssentialsMissingCheck` re-probes the LIVE page each run and emits a
`ResolvedFinding` keyed on `structuralItemKey("head_essentials_missing", page.ID)` — the
same key the finding was filed under — whenever `len(missing) == 0`.
`resolveWorkItems` (`work_items_common.go:548`) then closes by
`site_id + item_type + item_key` with `status NOT IN (closed statuses)`, and **does not
filter on `handler_agent`** — so a flag-only `detected` row IS retractable, which was the
one thing that could have made the prediction false. The stale `spec.missing` does not
block it either: retraction is computed from the fresh probe, not from the stored spec.

So the stale-spec problem (§(i)) is narrower than I first wrote. It is a **reporting**
defect — anyone reading `spec.missing` gets a claim of unknown age — not a blocker on the
row ever clearing. Downgraded accordingly in the plan's item 2.

---

## 2026-09-02, late — the council APPROVED and was right twice

Round 1, `3c71ec77-fd15-4aa1-a762-cc36116caca5`: **approved, 5 advisory objections, none
high-severity**, 12 seats. Recording all five with what I did, because two of them found
real defects and the approval is the less interesting half.

**(gg) `render_guardian` (medium) + `editquality` (low): MY CSS VARIABLES WERE INVENTED.**
The block referenced `var(--brand-accent,#000)` / `var(--brand-primary,#fff)`. No such
custom properties exist anywhere on this estate. `[MEASURED 2026-09-02]` on four live
stylesheets (cookly.uk, finetuning.uk, agritec.uk, webdesign.co.uk): 51 custom properties
defined between them, `--brand-*` **0 occurrences**, `--color-primary` 12–19 each.

  So the fallback would have fired on every page of every site: **hard-coded
  black-on-white, while LOOKING like it inherited the brand.** Valid CSS, visible link,
  nothing broken enough to notice — and I would have been quoting a fix that half-worked
  for months. This is the seat's own remit and it earned its place.

  Fixed to `var(--color-primary,#1a1a1a)` / `var(--color-primary-text,#ffffff)`, chosen
  over the `--color-cta-*` and `--color-accent-*` pairs because it is the one that is a
  contrasting pair *by construction* on all four sites (dark ground, `#ffffff` text),
  where `--color-cta-bg` is a pale beige on three and a `linear-gradient` on the fourth.
  New test pins every `var()` to a measured-real name **and** requires a fallback; proven
  by mutation — reintroducing `--brand-accent` fails it.

  **The lesson is the same one as this morning's mutation, one level up.** I wrote the CSS
  from the shape of the LMC exemplar without checking that the *names* in it generalised —
  and LMC is precisely the site that does not share the estate's stylesheet. **I copied an
  exemplar's structure and assumed its vocabulary.**

**(hh) `guardian` (medium): CSP.** An inline `<style>` is dropped silently on any site
enforcing `style-src` without `unsafe-inline` — which would leave the link visible.
Answered by measurement, not argument: **no site on this estate sends a
Content-Security-Policy header at all** (probed cookly, finetuning, webdesign.co.uk,
agritec, noted, loancalculator). And the estate is already built on inline style —
`injectComponentCSS` puts one on every page — so this block adds no new exposure. If a CSP
is ever introduced it breaks far more than this.

**(ii) `bug_historian` (medium): is "retires 867" actually true?** The seat's point was
sharp — my claim only holds for pages built through `assemblePage`, and a 14-page sample
is not a census. Answered from the check's OWN recorded field rather than more curling:
`spec.assembled` is written at filing time by `pageIsAssembled` (does the page have
`page_components` rows). `[MEASURED 2026-09-02]` **968 of 978 rows are `assembled=true`
across 31 sites — 99.0%.** The residual is 10 rows on 5 sites, and they are named, not
described: 4 on ai-agent-orchestration.com, 3 on gaswholesalers.com, 1 each on
finetuning.uk, idea.uk (`/tools.html#audience-check`) and robot-hands.com. All
`["skip_link"]` except idea.uk's, which is `["skip_link","footer"]`.

  **The evidence was already in the rows and I had not looked.** I answered a census
  question by sampling served pages this morning when the check had recorded the answer
  per-row at filing time.

**(jj) `debug_historian` (medium): no deploy-verification gate before the fan-out.** Fair
and now in the RUNBOOK as a two-stage gate. A fan-out against a pod still running the old
binary re-renders the estate to identical bytes — expensive, green, indistinguishable from
success.

**(kk) `guardian` (medium) + `bug_historian` + `reuse_agent` + `architecture`: the
`AssemblePageAction` second-path gap needs TRACKING, not a risk note.** The seat's words:
*"this council has punished exactly this omission before on the same landmine family"* —
`injectRobotsNoindex`, `injectCanonicalLink` and `injectPageJSONLD` have the identical
split. The `architecture` seat called the fourth recurrence *"a mild signal that
head-injection logic wants consolidating behind one entry point eventually — a future RFC
candidate, not this patch's job"*, and agreed folding it in here would be scope creep.

  **What I am NOT doing:** filing a `bugs_open/` entry asserting AssemblePageAction serves
  live pages. I have not established that. What I measured is that 10 flagged pages have
  no `page_components` rows, which is consistent with hand-built or adopted pages and does
  **not** prove they came from that action. Filing a root-cause claim on that evidence is
  the exact thing this estate's rules forbid. Tracked instead as an RFC candidate with the
  claim stated at the strength the evidence supports.

**(ll) `improvement_guardian`'s "missing" — how do the findings get retracted?** Already
answered in §(ff), before the verdict came back: `resolveWorkItems` does not filter on
`handler_agent`, so flag-only rows retract on a clean re-probe. Good question, and the one
place I was ahead of the council rather than behind it.

---

## 2026-09-02, night — the skip link is LIVE and proven at the artefact; the backlog movement is NOT mine

**(mm) Stage 1 of the RUNBOOK gate PASSES.** The fleet rolled at **20:56:43Z** (pod
`agent-chassis-8ddbf8958-cd2h9`). Probed the running binary for `data-skip-link` — present —
**with both controls in the same breath**: the older `GROWTH_DOOR_PROBE_FAILED` present, an
invented `ZZ_IMPOSSIBLE_LITERAL_ZZ` absent. ⚠ The absent-control TIMED OUT on the first
attempt (a full-binary grep for a non-matching string is slow) and I re-ran it alone rather
than reporting the two that had returned. **An incomplete control is not a control** — the
timeout case is `rc=124` and reads nothing like `rc=1`.

**(nn) Stage 1b PASSES on three real served pages**, each re-rendered after the roll:
`robot-hands.com/product-detail.html`, `gamesdesign.co.uk/games/economy-simulator/`,
`gaswholesalers.com/tools/fuel-margin-calculator/`. All three carry `class="skip-link"` ×1,
`id="content"` **×1 — not 2**, and the CSS marker ×1. Read the emitted markup rather than
trusting the counts:

```
<style data-skip-link="1">.skip-link{position:absolute;left:-9999px;…
   background:var(--color-primary,#1a1a1a);color:var(--color-primary-text,#ffffff)}
   .skip-link:focus{left:0}</style>
</head>
<body>
<a class="skip-link" href="#content">Skip to content</a>
…</header> … <span id="content" tabindex="-1"></span>
<main>
```

Order verified positionally: body 62136 < anchor 62143 < header 62444 < target 65085 <
main 65126.

**And the council's objection is resolved AT THE ARTEFACT, not just in the source.**
robot-hands.com defines `--color-primary: #1A1F2E` and `--color-primary-text: #ffffff`, so
the link renders in the site's own brand rather than the fallback. That is the thing
`render_guardian` said the `--brand-*` version could never do, now checked on a live page.

**(oo) THE BACKLOG MOVED AND IT IS NOT MY FIX. Checked before claiming it.**
`head_essentials_missing` open fell 978 → **968**, and 10 rows completed today — which reads
exactly like the fix draining the pile on its first night. It is not. Every one of the 10 is
lendzy.co.uk, completed at **19:05:17** — **one hour and fifty-one minutes BEFORE the roll**.
They gained a skip link by some other route (another lane's design run, unidentified). The
retraction reasons are the check's own *"re-probed …: title, skip-link and footer all
present"*, so the mechanism works — but the cause is not mine.

**Zero retractions are attributable to this change yet**, and that is expected, not a
disappointment: a row clears only after its page re-renders AND the structural check next
runs over that site, whose rotation is hours. **`[MEASURED 2026-09-02 21:2xZ]` open =
968.** Anyone reading a lower number later must check the completion TIMESTAMPS against the
roll at 20:56:43Z before crediting it here.

  The trap in one line: **a number that moves the way your fix predicts, in the window your
  fix landed, is not evidence your fix moved it.** The timestamps were one column away and
  the wrong reading was the comfortable one.

---

## 2026-09-03, evening — session 3: the ledger row that was not there, the clock that did not exist, and the report built

Picked up from `HANDOFF_2026-09-03_continue_here.md`. Re-measured before believing any of it.
Clock note first: the first measurements in this session were taken at **12:48Z** and the
session then sat idle; everything from **17:03Z** onward is a second reading, and the two are
not interchangeable — every figure below carries which one it is.

**(pp) THE PREVIOUS SESSION APPLIED 722 AND DID NOT RECORD IT.** `[MEASURED 2026-09-03 17:0xZ]`
`schema_migrations` carried rows for 718, 719, 720, 721, 723, 724, 726, 727 and both 728s —
and none for 722. The auto-memory note for the runner (`migration-runner-practice.md`) says
in as many words *"Recording is part of applying — do both in one motion"* and names this
exact file shape: a fully idempotent body (`CREATE OR REPLACE` + `DROP TRIGGER IF EXISTS`)
has no guard that RAISEs `already`, so it probes `ok` and *"reads as innocently pending for
ever after a hand-apply"*. It was worse than pending here: 722's ARM 4 hard-codes *at most 1
site carries a posture*, and 2 do since the loanzy lane held apis.uk this morning — so a
blanket `--apply` by any lane would have **halted on 722 and attempted nothing after it**.
Recorded at 17:1xZ with `--record-only` after re-verifying the trigger live in `pg_trigger`.
The cheap check was one line — `SELECT filename FROM schema_migrations WHERE filename LIKE
'722%'` — and it is now in the RUNBOOK and in `WRONG_CALLS.md`. Migration 752 (below)
carries an already-applied guard so the same omission would at least read as LIKELY ALREADY
APPLIED to the runner.

**(qq) copyonline.co.uk MISSED 722 BY ABOUT TWO MINUTES, AND I CANNOT PROVE THE TWO MINUTES.**
The handoff said *"no site has been created since 722 applied"*. `[MEASURED 12:48Z]`
copyonline.co.uk was created **09:27:25Z** by the portfolio_positioning lane, `settings = {}`,
`updated_at == created_at`, status `test`, locked. Their RUNBOOK recipe names no `settings`
column, so if the trigger had existed at that insert the row would read `hold` — a `BEFORE
INSERT FOR EACH ROW` trigger has no path that leaves a settings-less insert unstamped. So the
trigger was created after 09:27:25Z and before the commit that recorded the application
(`d8409d7e9`, 09:29:06Z). `[INFERRED from mechanism, not from a timestamp]` — and the
timestamp that would have settled it is the ledger row that was missing (pp). Consequence:
the handoff's claim was true by a margin of roughly a hundred seconds, and the demand test
(§3 there) is **still unrun**. `[MEASURED 17:04Z]` copyonline is now `active`, unlocked, and
**open** — the one live site on the estate born after the owner's ruling and not held by it.
**Not touched**: that lane is building tools on it today (their commits `7d1aa86a2`,
`599f54bc8`), and a hand-hold now would file their `add_tool` work as deferred mid-stream.
Put to the owner in the README; theirs and his to decide.

**(rr) THE CENSUS (handoff item 2).** `[MEASURED 2026-09-03 17:04Z]`:

| | 09-02 midday | 09-03 12:3xZ (handoff) | 09-03 17:04Z |
|---|---|---|---|
| `head_essentials_missing` at `detected` | 978 | 704 | **626** |
| retractions since the roll (20:56:43Z, status `complete`) | — | 265 | **343** |
| flag-only pile, all types | 1,385 | 1,162 | **1,084** |
| of which not skip-link | — | "~390" | **458** |

**The handoff's "~390 rows of other types" was wrong arithmetic**: 1,162 − 704 = 458, and it
was 458 then too. The other types have also crept up, not down — image_url_404 71→87,
heading_promise_unmet 139→148, prerequisite_missing 63→71, dead_internal_link_live 22→30,
page_content_divergence 19→23, plus two new types (`site_unreachable` 2, `needs_content_planning`
1). So the non-skip-link pile is **growing ~10%/day** while the skip-link half drains. That
pile is plan item 4's real input. Per site, `webdesign.co.uk` carries **151 of the 626**
skip-link rows — a quarter of the pile on one domain — which is where the drain will be
slowest to read.

Note the retractions land as `status='complete'`, not `retracted` — the handoff's word was
loose and my first query (`status='retracted'`) returned **0**, which for a minute read as
"the drain stopped". Vocabulary is in the RUNBOOK now.

**(ss) THE CLOCK DID NOT EXIST.** Handoff item 1 — the "held longer than N days" report — needs
a *when*. `[MEASURED 17:0xZ]` 722's trigger stamps only `growth_posture`. Of the two hand-held
sites, gamedesign.uk carries `growth_posture_reason` + `growth_posture_set_by` (with the date
inside the set_by **string**); apis.uk carries `growth_posture` alone. Two lanes, two shapes,
zero timestamps. `sites.updated_at` is useless for it: `ensure_site_record` upserts every site
on every loop pass (~50/day), so it dates the last visit.

So the report came with a migration. **752** (`_HOLD` while the verdict is pending)
`CREATE OR REPLACE`s the 722 function to also stamp `growth_posture_set_at` (`to_jsonb(now())`,
the shape `last_audit.at` already uses on the same object), `growth_posture_set_by`
(`'trg_sites_born_holding_growth'`) and `growth_posture_reason` — only beside a posture IT
stamped; a seed's explicit `open` gets no record, because `set_by = trigger` on a row the
trigger did not stamp would be a lie. **No backfill** of the two hand-held rows: other lanes'
rows, and a guessed timestamp is worse than none. Five verify arms; arm 4 **measures**
before/after on existing postures instead of hard-coding a threshold, because 722's "at most 1"
was true for a few hours and read 2 by the afternoon.

**(tt) THE REPORT, and the design choices that will look odd to the next reader.**
`deployments/kustomize/services/growth-posture-hold-check/` (the postgres:16-alpine Python
family, modelled on `site-discovery-staleness-check`), daily `0 8 * * *` UTC — the first free
slot after the live window (`kubectl get cronjobs` 2026-09-03: 06:20–07:55, last occupant
07:55).
- **A CronJob, not a discovery check.** A per-site check files a handler-less `detected` row —
  precisely the class this lane opened on. The watchdog for a silent state has to report where
  a person reads, driven from outside the mechanism.
- **Every held site is listed every run; N only decides which are findings.** N = 7
  `[ASSUMED]`, owner to rule.
- **LIVE = `status IN ('active','deployed') AND locked_at IS NULL`.** The 17 `pool` rows and
  locked/`test` rows are listed but never findings: not growing for a different reason.
- **An unstamped hold reads age UNKNOWN, bounded below by the check's own first sighting** —
  it reads its earlier `doc_notes` rows, anchored on the line `HELD <domain> `. No write to
  anybody's site; self-corrects the day the lane stamps the row. Stated caveat: a
  release-and-re-hold between runs overstates the bound.
- **The 722 trigger is watched too** (`trigger_missing`): dropped or disabled, no new site is
  born held and a clean report would lie.
- **Hand runs use the same file** (`scripts/audit-growth-posture-hold.sh` → `check.py --local`,
  kubectl route, no write unless `--write`). One file, nothing to drift.

Verified before submitting: `--self-test` 16/16 (threshold by mutation — 9 days is a finding
at N=7 and not at N=10; the SQL anchor and the renderer's line shape pinned to each other; the
dollar-quote tag never in the body); live dry run via kubectl: **60 site rows, 39 live, trigger
present+enabled, 2 held (both live, both age unknown), 0 findings, exit 0**; 752 probed alone
through the runner's doomed-transaction dry run — *ran to its own COMMIT without error,
everything rolled back*.

**(uu) THREE THINGS THE BUILD CAUGHT, all mine.**
1. I numbered the migration 750; **three other sessions took 750 within two minutes** of my
   writing it (boxingonline, copy_editor, pages_cta_rank). Renumbered 752; re-check the max
   at commit time, not at write time.
2. Self-test case 9 failed on my own escaping: the f-string's `E'\\n'` is `E'\n'` in the SQL,
   and my assertion spelled it with four backslashes. A test of a string constant that fails
   on first run is the point of asserting the anchor shape instead of trusting it.
3. I wrote **"2 of 41 sites"** into the register index row. I had not measured 41. Caught on
   re-read before commit; live is 60 rows / 40 deployed-or-active. Typing `[MEASURED]` next to
   it is what made me go and run the query — the marker rule working as designed, one line
   late.

**(vv) Loop health, last 24h** `[MEASURED 12:48Z]`: 56 `complete_clean`, 31 `complete`, 1
executing, **2 FAILED** — homegarden.uk `spawn_dispatch` on a Kafka dial (bugfix 040 family)
and gaswholesalers.com `spawn_site_review` reaped after >4h stale. Both known classes; noted,
not chased.

**(ww) Council: submitted corr `a9ac3bd5-43dc-40fa-8d05-db2113fa2a26` at ~18:2xZ.** Mechanism
committed with `Council-Submitted:`; the migration is `_HOLD` and the CronJob is NOT
kubectl-applied until the verdict, as 722 was. On APPROVED: apply by hand, `--record-only` in
the same motion, `git mv` the `_HOLD` off naming BOTH paths on the commit, `kubectl apply -k`,
then `scripts/audit-growth-posture-hold.sh --write` for the first row (that starts the
first-seen clock for the two unstamped holds).

> **CORRECTED 2026-09-03 17:2xZ (same session), two clock-times above are wrong — I read a BST
> wall clock as UTC.** §(pp) "Recorded at 17:1xZ" → the ledger row says **17:03:25Z**. §(ww)
> "submitted ~18:2xZ" → the submission file's mtime is **17:13:44Z** and the council run for
> `a9ac3bd5` was created at **17:13:50Z** (six seconds; no queue this time). Caught by reading
> the artefacts (`schema_migrations.applied_at`, `orchestration_states.created_at`) for the
> handoff rather than my own prose. Same family as the 12:48Z/17:03Z split at the top of this
> section: **a time I did not read off a clock is a time I made up.**

**(xx) PLAN ITEM 4'S INPUT, MEASURED — the eleven producers, read.** A read-only census of the
code that files each of the 13 flag-only `item_type`s (subagent sweep, then I re-verified the
three claims below that would change what we build). `[MEASURED 2026-09-03 17:3xZ]`:

| group | item_types (rows) | what the code says, verbatim |
|---|---|---|
| **DEFERRED HANDLER** (5 types, **~813 rows**) | head_essentials_missing 703, canonical_mismatch 49, dead_internal_link_live 30, page_content_divergence 23, sitemap_entry_dead_live 8 | *"FLAG-ONLY, THIS PASS … No auto-repair agent is wired in this pass"* (structural_validity header); *"future work GATED on bugs_open/251's own fix"* (canonical); *"per D5 ('no handler agent in v1') … accepted here for one release"* (divergence) |
| **HUMAN JUDGEMENT** (7 types, **~270 rows**) | heading_promise_unmet 148, image_url_404 87, prerequisite_missing 71, structure_floor_unmet 27, archived_page_still_serving 9, asset_reference_404 4, site_unreachable 2 | *"the repair is a planner/writer judgement"*; *"removing or repointing it, which no image generator can decide"*; *"the difference is intent, which lives in a human"*; *"nothing can repair Cloudflare routing today"* |
| **ORPHAN** (1 row) | needs_content_planning 1 | no producer files this type handler-less at `detected`; the row was a **hand INSERT** (`created_by = gripper_dossier_ai_page_3_315_reopen`, `source = manual`, 09-02 17:17Z) — another lane's, not the loop's |

Three facts that change the design, each re-verified by me in the file:

1. **`image_url_404` CAN NEVER CLOSE.** `check_image_url_404.go` contains **zero**
   `ResolvedFinding` emissions (the one "resolved" match is a comment at :68); the control,
   `check_site_structural_validity.go`, has 5. So a repaired image leaves its row at `detected`
   for ever. **87 rows across 34 sites, oldest week 2026-07-20; 8 `complete` + 3 `cancelled` ever,
   live+archive, none by the check.** `asset_reference_404` has the same hole for its
   `empty_src` key (one retraction site at :339, keyed on `unresolvable_reference` only). This is
   a defect distinct from "nobody drains the pile": these rows are wrong even after the page is
   fixed. Candidate for its own `bugs_open/` file — **prior-art grep first** (printed above this
   entry in the session log; not yet done at the time of writing).
2. **There is no shared flag-only constructor.** The mechanism is `HandlerAgent` left
   zero-valued on `WorkItemSpec` (`registry.go:66`) and a copy-pasted comment — *"HandlerAgent
   intentionally empty — flag-only, see the header"* — at **11 sites** (`grep -c` confirms).
   Every one of these rows crosses ONE seam (`discovery_checks.go:249` → `insertWorkItem` →
   `writeWorkItem`), which neither validates nor annotates an empty handler on a `detected`
   row. That seam is where a change lands once, not eleven times — the same "door, not the
   producers" argument WDS-020's council round made.
3. **Exactly one consumer, and it only counts.** `countUnroutableDetected`
   (`work_items_common.go:359`) feeds `not_promotable` in the triage output; nothing reads it
   beyond a log line. Every other reader of handler-less rows filters on `deferred` or
   `capability_gap`. And `livespec/unarmed_completers.go:85` declares a live
   `image-url-404-handler` that the detector deliberately never routes to — *"3 rows ever
   reached this agent, by hand, and it escalated all three back to needs_human_review"* — so
   the human-judgement classification has been tested once, by the agent itself.

`triage_hint`: only `archived_page_still_serving` writes the literal key; three others write a
`not_dispatchable` sentence into `spec`; the remaining nine write nothing a human could act on.
The handoff's "9 of 1,385 carry a triage_hint" was the first group.

**What this does to item 4.** It is now two changes, not one, and neither is "a queue":
(a) a **retraction contract** — a flag-only finding must be able to close itself, and
`image_url_404` / `asset_reference_404.empty_src` violate it today; (b) at the shared seam, a
handler-less `detected` row from a HUMAN_JUDGEMENT producer files at `needs_human_review` with
its `not_dispatchable`/`triage_hint` text as the brief (912 peers already live there visibly),
while a DEFERRED_HANDLER producer's rows stay `detected` and are counted per type by the
daily check family, so "this pass" has a number and a date. Both are shared-seam changes →
`090` for the root-cause claim, then the council. **Not started; the skip-link drain is not at
its floor and the 458 still need reading at the row level, not just the producer level.**

> **§(xx) item 1, follow-through (17:4xZ):** prior-art grep done — `image_url_404` appears in 8
> `bugs_open/` files and none of them, nor 016b, nor LANDMINES, says the check cannot retract.
> The neighbouring lane (`bugfix_149`) is quiet 14 days and nobody is mid-edit on the file.
> **Filed as `bugs_open/470`** with the 090 substitution stated (local, one file,
> self-evidencing), fix candidates ordered by what closes the door, and the §9 pattern *"a check
> that files a finding and has no path to retract it"* added to 016b. Not fixed — that is item
> 4a of the revised plan, after the 458 are read at row level.

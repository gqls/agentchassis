# 391 — the primary CTA destination is chosen by `nav_order` alone, with no relevance input at all, so an off-topic tool wins the primary button — and the wrong pick then LOCKS ITSELF IN through the button's own copy

> # ⚠ THIS FILE WAS 389 FOR ~40 MINUTES ON 2026-08-25. It is now **391**.
> It collided with `bugs_open/389_HANDOFF_2026-08-25_repair_completion_is_unverified_three_classes_complete_unchanged.md`
> (the `bugfix_308` lane's re-file of 308 Phase C), filed **2m25s earlier** (`10:51:25` vs
> `10:53:50`) — the documented `ls`-then-`add` race. **That lane's 389 keeps the number**: it is
> cited from `bugs_closed/308`'s closing section, so it was the expensive one to move. I moved, and
> **390 was taken by a third session in the interval**, hence 391.
> **Commits between 10:53 and 11:40 saying "389" about CTA SELECTION mean this file.**
>
> Complementary, not duplicates: 389 is *why FIXING a CTA can report success without changing
> anything*; this is *why the CTA points at the wrong page to begin with*. **`git log` the FILE
> PATH, never the number.**

**Filed 2026-08-25, on an owner report.** The owner saw `/tools/password-entropy.html` offered as
the call-to-action on an AI-orchestration consultancy and said plainly: *"not deliberate and
actually wrong."*

**Status: OPEN. Root cause confirmed at the code, the data and the served bytes, and independently
re-verified by adversarial review on 2026-08-25 — every load-bearing claim reproduced. Nothing
fixed.**

> ⚠ **READ §"THE FEEDBACK LOOP THIS FILE ORIGINALLY MISSED" (near the end) BEFORE ACTING ON
> ANYTHING ABOVE IT.** That review found a mechanism I had missed, and it **changes both the fix
> and the sizing**: the label match runs *ahead* of the positional pick, and the framework
> instructs the writer to produce button copy naming whatever destination it picked — so a wrong
> pick converts itself into a **label-locked** one that a `nav_order` correction can no longer
> reach. **All three buttons the owner saw are in the label-locked set.** The decisions list and
> fix candidates below are superseded there.

**Escape hatch, per the 2026-07-31 owner ruling:** `090` was not run. Substituting stated
first-hand verification: I read `chooseCTATargets` and `loadInteractivePages` in full, simulated
their exact ORDER BY against the live `pages` table, confirmed the predicted winner matches the
observed stored value on all three affected sites, dated the writes by the resolver's own
provenance stamp, and confirmed the result at the served bytes on all three domains. The cause is
a 30-line deterministic sort, not something needing DB-side induction.

> ⚠ **CORRECTED 2026-08-25: that substitution covered the DIAGNOSIS half and demonstrably NOT the
> COVERAGE half.** `090`'s trigger also performs an open-work check against the target, and the
> failure that actually happened — missing a **commissioned** lane holding the same population,
> measured 10 days earlier (see §CORRECTION) — is exactly the failure that check exists to prevent.
> The paragraph above was written before I knew that. First-hand verification substituted for the
> mechanism reasoning; it did not substitute for coverage, and I should not have implied it did.

## The mechanism, plainly

When the framework needs a destination for a call-to-action button, it asks
`chooseCTATargets` (`platform/orchestration/actions/resolve_internal_links_action.go:651`). That
function:

1. takes **every** `page_type IN ('tool','game')` page on the site
   (`loadInteractivePages`, same file:918) — there is no filter on subject, quality or fit;
2. drops only two things: pages in excluded areas (contact, legal, about) and a page pointing at
   itself;
3. sorts what is left by **`nav_order` ascending, then by `name` alphabetically**;
4. returns `ordered[0]` as the primary CTA and `ordered[1]` as the secondary.

**That is the whole selection.** No topic, no tags, no semantic match, no LLM judgement, no
relation to the page the button sits on. The site's primary call to action is whichever tool page
happens to sort first.

## Why it produced this particular embarrassment [MEASURED 2026-08-25]

`password-entropy` — a password-strength toy — carries **`nav_order = 1`** on exactly the three
affected sites, while every genuinely relevant tool on those sites sits at 6–204:

| site | password-entropy | the relevant tools it beats | tools on site |
|---|---|---|---|
| `ai-agent-orchestration.com` | **1** | ROI estimator, LLM cost calculator, build-vs-buy analyzer (200–202) | 6 |
| `finetuning.uk` | **1** | AI readiness quiz, GDPR risk assessment, model-approach selector (200–204) | 9 |
| `leopardessconsulting.co.uk` | **1** | ROI estimator (6), LLM cost calculator (7), vendor trust checklist (8) | 7 |

So it wins `ordered[0]` on every page of all three sites, deterministically, every time the
resolver runs. All three rows were created `2026-03-13` — the `nav_order = 1` is a fossil of an
early deploy, and nothing since has had any reason to look at it.

**Why those sites have the tool at all** is a separate, already-documented story:
`docs/agent_docs/sql_for_tables/005_content_components.sql:8942` records that the password
checker was deployed to four sites *"because the library only had 2 tools with templates, giving
the LLM no real choice"*. That is history. **This bug is about the ranking, which is live.**

## It is live, not residue — this is the part that matters

The resolver stamps what it mints (`__cta_minted`, LNK-035, live 2026-08-22). Splitting the 80
CTA url fields that point at this tool by that stamp [MEASURED 2026-08-25]:

| provenance | fields | dates |
|---|---|---|
| **stamped as minted by the resolver** | **17** | **2026-08-23 → 2026-08-25** |
| ~~stamp present but naming a different url (reads authored)~~ → **stamp on the component but NO entry for THIS field** (it names a sibling slot) | 24 | 2026-08-24 |
| no stamp at all | 39 | 2026-08-15 → 2026-08-24 |

> ⚠ **CORRECTED 2026-08-25 (adversarial review).** I described the 24 as *"stamp present but naming
> a different url, so the value reads authored"*. Both halves are wrong. **Zero** rows carry a stamp
> entry naming a different url for the field; the 24 have **no entry for that field at all**, the
> stamp covering a sibling slot (e.g. `{"secondary_cta_url":"/contact.html"}` while `cta_url` holds
> password-entropy). And *"reads authored"* is wrong in the code's terms:
> `storedCTADestinationIsAuthored` returns true only for **utility-area** urls, which a `/tools/`
> url is not. **Evidentially the 24 belong with the 39 as unattributable** — the honest split is
> **17 attributable to the resolver, 63 unattributable**, not 17/24/39.

> ⚠ **AND "minted TODAY" is stronger than the stamp can prove.** The stamp is value-bound with **no
> timestamp of its own** — the dates above are the row's `updated_at`, and a `SeedCTAMinted`
> carry-forward is indistinguishable from a fresh mint. **What does prove it is live** is the
> ranking simulation plus one positional mint inside the stamp era whose copy cannot have
> label-matched: `ai-agent-orchestration.com/containment-first-architecture` hero, copy **"Book a
> Technical Discovery Call"**, `cta_url=/tools/password-entropy.html`, row written **2026-08-24**.
> No label match produces that pairing. **Cite that row, not the word "today".**

Any repair that only rewrites the stored values will be undone.

## Confirmed at the served bytes, 2026-08-25

```
ai-agent-orchestration.com/index.html    "Try the Password Strength Physics Tool"   (4 refs, minted today)
finetuning.uk/services.html              "Explore Password Strength Physics"
leopardessconsulting.co.uk/services.html "See a working example first: try the Password Strength
                                          Physics tool, built and run on the same platform."
```
Note the third: a consultancy pitching its build capability by pointing at a password toy. **The
button copy is generated to match whatever href it was given**, so a wrong destination does not
look wrong — it reads as a deliberate recommendation. That is why this survived to the owner's eye
rather than looking like breakage.

## The design fault under the specific bug

**One column, `pages.nav_order`, is doing two unrelated jobs**: ordering the navigation menu, and
ranking CTA candidates. Nothing says so at either site, and the two readers disagree about what
the column means.

The consequence is sharp and is documented in this repo by accident. On
`docs/leopardessconsulting/scripts/L5_nav_and_ctas.sql:29` a previous session hid this exact page
from the nav with the comment *"a password tool doesn't belong in the primary nav"* — and set
`in_header = false`. **It changed nothing.** `chooseCTATargets` never reads `in_header`; the
linkability predicate it does use (`PageMayBeLinkedPredicateFor`, `datahelpers/links.go:333`) is
purely about deployment state. So the one signal a human gave — *this page should not be
prominent* — was invisible to the code that decides the most prominent link on the page, and the
page kept the `nav_order = 1` that makes it the top CTA choice.

> ⚠ **A stronger version of that claim is FALSE and I nearly filed it.** I measured 13 sites whose
> rank-1 CTA target has `in_header = false` and was about to call them 13 deliberate contradictions.
> **62.7% of all tool/game pages fleet-wide are `in_header = false`** (143 of 228), so that column
> does not mean "a human judged this inappropriate" — for most tools it is simply the normal state.
> The leopardess case is a real deliberate hiding *because its SQL comment says so*, not because of
> the flag. Cited as one documented instance, not a population.

## Blast radius [MEASURED 2026-08-25]

Pointing a CTA at a tool is **normal and wanted** — **105** such fields on `webdesign.co.uk` and
**78** on `dartsonline.com` [MEASURED 2026-08-25 ~11:00Z; webdesign re-measured **111** two hours
later, 44 of those rows written that same day — the figure moves hourly, so quote it with its
timestamp or re-run it]. The defect is not "CTA points at a tool", it is "the tool is chosen by a sort
that cannot know what the site is about". Reviewing the current rank-1 winner for all 26 sites
with tool pages, the winner is plausibly on-topic almost everywhere (`dartsonline` →
checkout-calculator, `loancalculator.co.uk` → car-finance-calculator, `relojistas` →
watch-service-interval). **The three password-entropy sites are the only clearly wrong ones**;
`webdesign.co.uk` → `tool-ab-test-calculator` (chosen from **66** tools) is borderline and worth a
human glance. So: a fleet-wide mechanism, currently doing visible damage on three sites.

**The natural control, which I should have run first and did not** [MEASURED 2026-08-25]:
`gaswholesalers.com` is the **fourth** site named in my own citation from
`005_content_components.sql` (*"deployed to 4 sites (including gaswholesalers)"*) and the prior
lane's third-worst offender at 28 rows on one modal target. **It has no `password-entropy` page at
all, and all six of its tools sit at `nav_order = 200`.** So it escaped by lacking the fossil while
still showing the *repetition* the commissioned content pass is for, its modal target
(`tool-breakeven-volume-calculator`) being on-topic. **That cleanly separates the two problems** —
repetition is fleet-wide and is the content pass's business; off-topic is the fossil and is this
bug's. A pointer sitting inside evidence I had already quoted.

**The secondary slot is clean** [MEASURED 2026-08-25]: `chooseCTATargets` also returns `ordered[1]`
as the secondary CTA, and password-entropy appears at rank 2 on **no** site — it is rank 1 on the
three, and absent from the top two everywhere else. Measured with the code's own deployment
predicate included (see the ⚠ on the verify query).

## The decisions this needs — FIVE, not four, and they are not the same kind

> ⚠ Revised 2026-08-25 after review. Decision 5 was missing; decision 3's claim was overstated;
> decision 2's reach is smaller than stated — see §THE FEEDBACK LOOP.

1. **Content — should the tool be on those three sites at all?** Removing the page removes the
   candidate outright. Cheap, and it is a decision about what the sites offer, not a platform fix.
   The context that informs it: the tool was pushed to four sites in March *"because the library
   only had 2 tools with templates, giving the LLM no real choice"*
   (`005_content_components.sql:8942`). That reason has expired — these sites now carry 6–9 tools
   each.
2. **Data — should `nav_order` be corrected?** One `UPDATE` per site. ⚠ Two limits: on
   `ai-agent-orchestration.com` the page is `in_header = true`, so `nav_order` is *also* its visible
   menu position and the fix moves the menu item; and **it only reaches the ~60 label-less fields**,
   not the ~20 whose copy already names the tool (§THE FEEDBACK LOOP).
3. **Platform — should the chooser get a relevance input, or an explicit opt-out?** ~~The only
   option that stops the class.~~ **CORRECTED: an opt-out does NOT stop the class** — it is
   reactive, and the next off-topic tool with a low `nav_order` still wins by default until a human
   notices and sets the flag. Under this file's own ordering principle it makes the *good* state
   sayable; it does not make the *bad* state unrepresentable. Only a relevance input changes the
   default (candidate 3), and only a detector catches recurrence (candidate 4). **The pairing that
   earns the claim is candidate 1 + candidate 4** — lever plus alarm — and that is what I would
   put to a council, not candidate 1 alone.
4. **Repair — the 80 stored values.** ⚠ Constrained twice over: by `bugs_open/389`
   (`repair_completion_is_unverified` slug — a `cta_links_stale` rerender reports `complete`
   whether or not any CTA moved, so **status is not evidence**; verify at the served bytes or the
   stored field), and by the label lock, which means a re-run alone re-selects the same tool for
   the ~20.
5. **The standing commission — honour, re-scope, or withdraw it?** ⚠ **NEW, and it is a decision
   about the owner's own prior instruction**, so it cannot be folded into 4. On 2026-08-15 the
   owner accepted this as a floor and commissioned `cta_target_content_pass` to vary CTA targets
   page by page; it was never run. The 08-25 report withdraws the floor. The pass is **not**
   redundant — it is precisely what the ~20 label-locked fields need — but its **scope should be
   set after** decisions 2/3 land and the population is re-measured, not before.

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **An explicit `pages.eligible_as_cta_target` (or equivalent) opt-out, default ON.** Makes "this
   page must never be a CTA destination" *sayable* — today it is not sayable at all, which is why
   hiding it from the nav was the only move available and did nothing.
   > ⚠ **TWO CORRECTIONS, both from review, both changing the specification.**
   > **(a) Do NOT implement it in `loadInteractivePages`.** The loaders have a third consumer no CTA
   > document named until 2026-08-22: `render_site_components_action.go:182-190`, the **site header
   > CTA fallback**, which calls `loadContentHubs`/`loadInteractivePages` directly and takes
   > `ordered[0]` — and its output is **never persisted** (`site_components` carries **0** `cta_url`
   > keys across its header rows). A change at the loaders therefore re-picks **every site's header
   > button**, and no `content_data` diff will show it. **Change the RANKING, not the loaders.**
   > (Told to this lane by the `bugfix_308` lane; the header path's independence from
   > `chooseCTATargets` is recorded in `datahelpers/cta_label_universe.go`'s header, LNK-036, and
   > was confirmed by this file's own caller enumeration.)
   > **(b) A flag on the ranking alone does not bind `LoadCTALabelUniverse`.** The label match runs
   > *ahead* of the positional pick, so an "ineligible" page would still be selected whenever the
   > button's copy names it. To make "never a CTA destination" actually true, the flag must be read
   > by the **label universe too** — otherwise it is an opt-out with a hole exactly the shape of the
   > damage on these three sites.
2. **Stop overloading `nav_order`**: give the chooser its own ordering input, or have it read
   `in_header` as a demotion signal. Cheaper, but keeps two meanings in one column.
3. **A relevance signal** (site tags × `content_components.semantic_tags`). Most work, most
   judgement, most false positives — and note the existing tags are already misleading: the
   "narrow password-entropy tool affinity" migration in `005_content_components.sql`
   **added** `tech`, `cybersecurity`, `developer` to them.
4. **A detector**, not a fix: file a work item when a site's rank-1 CTA target is a tool whose
   `nav_order` is anomalous against its siblings. Catches the fossil-value shape specifically.
5. ~~Fix the three `nav_order` values and move on~~ — leaves the mechanism intact, and the ranking
   will pick something else arbitrary the next time a tool is added with a low `nav_order`.

## Verify (and the control that stops a false pass)

```sql
-- the ranking the resolver will actually use, per site
SELECT s.domain, p.name, COALESCE(p.nav_order,100) AS nav_order,
       row_number() OVER (PARTITION BY s.id ORDER BY COALESCE(p.nav_order,100), p.name) AS rank
FROM pages p JOIN sites s ON s.id=p.site_id
WHERE p.page_type IN ('tool','game') AND p.status IN ('active','deployed')
  AND NOT (p.deployed_at IS NULL AND COALESCE(p.build_status,'') = 'planned')  -- ⚠ ADDED 08-25
ORDER BY s.domain, rank;
```
⚠ **The predicate on the third line was MISSING from the first version of this query**, which is
the one the 26-site blast-radius review above was run with. `loadInteractivePages` applies
`PageMayBeLinkedPredicateFor` (`datahelpers/links.go:328` — **not :333**, corrected), so without it
the simulation can name a rank-1 "winner" the code would skip: harmless on the three sites here
(password-entropy is `deployed`), but the fleet review should be re-run with it. This file's own
RUNBOOK says *"mirror the code exactly or the simulation proves nothing"* and the first version did
not. Note also that `rank()` excludes **the page itself**, so a tool page's own CTA takes the next
candidate down, and hub pages rank behind all interactive ones.
**Control:** these sites link `password-entropy` legitimately from `/tools.html` and from the nav,
so grepping a page for the URL **passes today while the wrong button is untouched**. Assert on the
anchor whose text names the tool AND sits in a `hero`/`call-to-action` slot — or read the stored
`cta_url`/`primary_cta_url` fields directly, which is unambiguous.

**Re-minting check** (the one that proves a fix holds): after any repair, confirm no NEW
`__cta_minted` entry names this URL —
`... WHERE (content_data->'__cta_minted') ? key AND value LIKE '%password-entropy%' AND updated_at > <fix time>`.

## Relations

`bugs_closed/299` (`home_page_cta_names_the_brief_starter_tool` slug) — same family, and its close
notes the detectors it left: they check CTA **validity** (does the destination exist, is it a page,
does the copy name it). **None checks relevance**, which is why a valid, existing, correctly-named
but topically alien destination sails through. · `bugs_open/312` (the resolver/writer seam, still
open) · `bugs_open/203` (an unconditional default, different mechanism) · `bugs_closed/268` (the
fleet CTA-destination repair machinery to reuse) · `bugs_closed/191` (header CTA validated by a
looser predicate than the nav beside it — the same nav-vs-CTA divergence, one layer up) ·
`bugs_open/248` (recompute clobbering authored links — constrains any repair run).

---

# ⚠ CORRECTION 2026-08-25, same day, before anything was acted on — THIS SYMPTOM WAS ALREADY MEASURED AND COMMISSIONED, and it changes the recommendation

**I filed the sections above without finding the prior lane, and I should have.** I grepped
`bugs_open/` and `bugs_closed/` for the mechanism (which is what "grep before you file" literally
says) and found the adjacent bugs. **The prior art was not in a bug file — it was in a lane**, and
one line of `MEMORY_workstreams.md` names it. The cheap check I skipped: **grep the workstreams
index, not just the bug directories.**

## What was already known, 10 days before this file [`cta_target_content_pass`, 2026-08-15]

`docs/agent_docs/docs024_key_docs_latest/cta_target_content_pass/PLAN_2026-08-15_cta_target_content_pass.md`
measured the same population and named the same shape:

> *"Where the button's wording named a real page, label-match chose it (good). Where it did not,
> the positional fallback chose **the site's top-ranked interactive page** — the same one for every
> such page on the site. Measured 2026-08-15: **16 sites have ≥6 rows on their modal target; worst
> are finetuning.uk (39 rows on /tools/password-entropy.html), ai-agent-orchestration.com (36, same
> tool) and gaswholesalers.com (28)** — and `/tools/password-entropy.html` is the modal target on
> THREE sites, sometimes topically absurd (an AI-services page's main button pointing at a password
> checker)."*

**The owner accepted that as a floor on 2026-08-15 and commissioned a content pass** to vary the
targets page by page. **Nothing has been run.** So this file's symptom is not new, and the owner's
2026-08-25 report — *"not deliberate and actually wrong"* — is best read as **the floor being
withdrawn**, not as a fresh discovery.

## What this file actually adds, stated honestly

1. **WHY that particular tool wins**, which the 08-15 plan did not have: not "it is top-ranked" as
   a brute fact but **`nav_order = 1`, a fossil set at page creation on 2026-03-13**, beating every
   relevant tool at 6–204 on exactly the three affected sites. That is the difference between a
   condition and a cause, and it is what makes a cheap fix possible.
2. **`in_header` is not read by the chooser**, so hiding the page from the nav — already tried,
   with a comment explaining why — cannot help.
3. **It is still minting today** (17 stamped 08-23 → 08-25). The stamp did not exist on 08-15
   (LNK-035 shipped 08-22), so the earlier lane could not have known whether this was live or
   inherited. It is live.

## ⚠ The recommendation in §"The decisions this needs" is REVISED accordingly

The commissioned content pass is an **LLM rewrite over every affected page of 16 sites**, and its
own plan flags three caveats it must design around. The `nav_order` finding means **the three
worst sites may not need it at all**: correcting one number per site changes which tool the
positional fallback picks for *every* page on that site at once.

**So the ordering matters more than the choice:**

1. **First**, correct the ranking input (decision 2) or add the opt-out (decision 3) — cheap, and
   it moves the modal target on the worst three sites in one step.
2. **Then re-measure** the 08-15 population. The content pass is sized against a number that is
   now 10 days stale and that step 1 will change substantially.
3. **Only then** scope the content pass over whatever repetition remains, which is a real problem
   (repetitive CTAs across a site) but a *different* one from the off-topic CTA the owner reported.

Running the content pass first would spend an LLM rewrite of 16 sites to work around a fossil
integer — and would leave the fossil in place to mis-rank the next tool added.

## One more consumer this file did not name

The `bugfix_308` lane recorded (2026-08-22) a **third** consumer of the CTA candidate loaders that
no CTA document names: `render_site_components_action.go`'s **site header fallback**. Its output is
never persisted — *"`site_components` carries 0 `cta_url` keys across all 24 header rows"* — so a
`content_data` before/after diff **reads clean while all 24 headers move**. Any fix here must be
verified at the rendered header, not in `content_data`.

**Relations, added:** `cta_target_content_pass` (the commissioned pass this is the root cause of —
tell that lane before acting), `bugs_open/308` (the third consumer, and the provenance work).

## ⚠ DECISION 4 (repair) IS CONSTRAINED BY THE OTHER 389, and that is not a coincidence

The `repair_completion_is_unverified` slug (filed 2½ minutes before this one, from the
`bugfix_308` lane) proves that **a `cta_links_stale` rerender reports `complete` whether or not any
CTA moved** — `suggested_target` is written by the detector and read by nothing, so "repaired" is
asserted by the handler and never checked against the page. It names three measured classes where
`complete` provably meant *unchanged*, including **124 of 135 live findings sitting in components
absent from `ctaFieldNames`**, and its class 3 is `ai-agent-orchestration.com/blog`'s frozen
hero + call-to-action — **the same site as this bug**.

**So the repair step here must not be scheduled as "run the 268 re-run and read the status."** Any
repair of the 80 stored values must be verified **at the served bytes or at the stored
`cta_url`/`primary_cta_url` field**, and a green work item is not evidence. That lane's fix
candidate 1 (`VerifyMisdirectedCTAResolved` — re-run the detector's predicate post-render and
refuse completion if it still fires) is the thing that would make decision 4 safely automatable;
until it lands, treat every repair run here as unverified by default.

---

# THE FEEDBACK LOOP THIS FILE ORIGINALLY MISSED — the wrong pick writes the copy that locks it in

**Added 2026-08-25 after an adversarial review of this file.** The review confirmed every claim
above and then found the thing they all sit inside. This section supersedes the sizing in
§CORRECTION and the fix specification in §Fix candidates.

## 1. The positional pick is not the only selector, and it runs LAST

`setCTAField` (`resolve_internal_links_action.go:418-434`) tries **`BestLabelMatchForPage` on the
slot's existing label FIRST**, then the authored/utility/non-page keeps, and only then the
positional pick. Three callers, three different orders [VERIFIED 2026-08-25, caller enumeration]:

| caller | order | note |
|---|---|---|
| build — `resolve_internal_links_action.go:162` | label-match → keeps → **positional last** | no "keep any valid stored value" branch, so a slot with generic copy takes the positional pick **on every build** |
| rerender — `rerender_page_sections_action.go:969` (`applyCTARecompute`) | label-match → keeps *any* valid stored destination → positional | positional is genuinely last-resort here |
| **site header** — `render_site_components_action.go:190` | **pure positional, no label match at all** | output **never persisted**; changing the loaders moves all 24+ header buttons invisibly |

**So the positional path is common, not rare** — which is what makes this bug real rather than
theoretical. The proof is a row whose copy cannot have produced its destination:
`ai-agent-orchestration.com/containment-first-architecture` hero, copy **"Book a Technical
Discovery Call"**, `cta_url = /tools/password-entropy.html`, stamped minted, written **2026-08-24**.

## 2. The framework then writes copy that names whatever it picked

`stampCTADestinationGuidance` (`resolve_internal_links_action.go:362`) *"appends `Destination
(fixed): <title>…` to the `llm_field_specs` entry for the CTA's **LABEL** field, once the URL field
has a resolved companion title"*, and those specs pipe `plan_sections → llm_field_specs[].description
→ the writer`. The config key `stamp_cta_destination_guidance` is **live and true** in
`agent_definitions` [VERIFIED 2026-08-25].

**So the sequence is:** positional pick chooses a page → resolver writes `*_target_title` → guidance
tells the writer the destination is *fixed* → the writer produces a button that **names** it.

Measured on the 80 fields [MEASURED 2026-08-25]:

| provenance | fields | carry a `*_target_title` | title names the tool | **copy names the tool** |
|---|---|---|---|---|
| stamped minted | 17 | **17** | **17** | **16** |
| stamp on a sibling only | 24 | 21 | 21 | 1 |
| unstamped | 39 | 37 | 37 | 3 |

This settles a claim the review rightly flagged as unevidenced. I had written *"the button copy is
generated to match whatever href it was given"* with no citation; it is now `:362`, the specs pipe,
and the 17/17 correlation.

## 3. Why that changes the fix — and it is the part to act on

A `nav_order` correction only reaches a field whose copy does **not** already name the tool,
because on the next resolve the label match runs first and re-selects password-entropy on the
strength of the wording:

- **~60 of 80** fields have generic copy → the positional pick governs → **`nav_order` fixes them**;
- **~20 of 80** have copy naming the tool → **label-locked**; `nav_order` cannot reach them;
- ⚠ **all three buttons the owner actually saw are in the label-locked 20.**

**The wrong pick is therefore self-reinforcing, and the locked set grows with every mint.** Each
positional mistake gets copy written for it, and that copy converts a ranking accident into a
content fact that outlives the ranking fix.

## 4. What this does to the recommendation

My §CORRECTION said the commissioned content pass might be unnecessary for the worst sites. **That
was wrong in exactly the wrong direction:** the content pass is *precisely* what the label-locked
20 need — including every button the owner reported. The two remedies are **complementary and
separable by measurement**, not alternatives:

1. **`nav_order` / ranking fix** → the ~60 label-less fields, in one step per site, no LLM;
2. **the commissioned content pass** → the ~20 label-locked fields, scoped by the query below
   rather than by site, which is far cheaper than the 16-site sweep its plan assumed;
3. **and the ordering still holds** — do 1 first, because until the ranking stops producing new
   wrong picks, step 2 is rewriting copy that the next build re-locks.

```sql
-- which fields are label-locked (need the content pass) vs positional (need only the ranking fix)
SELECT (pc.content_data->'__cta_minted') ? kv.key AS minted,
       count(*) FILTER (WHERE COALESCE(
         pc.content_data->>(replace(kv.key,'_url','_text')),
         pc.content_data->>(replace(kv.key,'_url','_label')),
         pc.content_data->>(replace(kv.key,'_url',''))) ILIKE '%password%') AS label_locked,
       count(*) AS total
FROM page_components pc CROSS JOIN LATERAL jsonb_each(pc.content_data) kv
WHERE kv.key LIKE '%\_url' AND kv.value #>> '{}' LIKE '%password-entropy%'
GROUP BY 1;
```

## 5. Two RFC/registration consequences

- **RFC_022 (2026-08-11) applies and I had not engaged it.** I asserted decision 3 is
  "architecture-scope" citing the 2026-08-02 §2 ruling, while skipping the narrowing that qualifies
  it: an opt-in field whose unsafe default is OFF and which **no live consumer names** is *not*
  architecture-scope. Candidate 1 may well meet all three conditions — but the ruling also requires
  **enumerating the consumers, and asserting it without the query is itself the objection**. That
  enumeration is owed before anyone books a council round; over-caution here costs a round the
  owner is being asked to approve.
- **The code already reserves the hook for candidate 3.** `chooseCTATargets` carries `pageType` as
  a parameter it never reads, documented as *"v2 does not branch on it"* and carried "for a future
  intent-aware (LLM) upgrade". Relevant to sizing the relevance option, and I had not mentioned it.

---

# OWNER DECISIONS 2026-08-25 — all five answered. One is DONE; the rest are sequenced below

> The owner answered all five in chat. Recorded verbatim in substance, with what each licenses and
> what it does **not**.

| # | decision | owner's answer | state |
|---|---|---|---|
| 1 | should the tool be on those sites | **"the password tool can disappear everywhere"** | **planned, not done** — see §RETIREMENT |
| 2 | correct the `nav_order` fossil | **"yes change the menu-order numbers"** | ✅ **DONE + VERIFIED 2026-08-25** |
| 3 | build the platform lever | **"yes go ahead"** | not started — code + council |
| 4 | repair the 80 stored values | **"whatever you suggest"** | sequenced last |
| 5 | the standing commission | **"rescope it as you suggest"** | ~20 locked fields, scoped by query |

## Decision 2 — DONE. `SQL_2026-08-25_demote_password_entropy_nav_order.sql`, applied

Three rows moved `nav_order` 1 → **900**, guarded to abort unless exactly three moved. The new
rank-1 CTA target on each site, measured after the change:

| site | was | **now** |
|---|---|---|
| `ai-agent-orchestration.com` | password-entropy | **tool-ai-agent-roi-estimator** |
| `finetuning.uk` | password-entropy | **tool-ai-data-risk-checker** |
| `leopardessconsulting.co.uk` | password-entropy | **ai-agent-roi-estimator** |

⚠ **Why 900 and not 200** — this is the trap in this fix and it nearly bit: 200 is those sites'
ordinary tool value, so it would **tie**, and the tiebreak is alphabetical on `name`.
`password-entropy` sorts ahead of every `tool-*`, so **at 200 it would still have won** on two of
the three sites. A demotion that merely joins the pack is not a demotion here.

⚠ **What this does NOT do:** it changes what *future* resolutions pick. **The 80 already-stored
values are untouched**, and the ~20 label-locked ones would survive even a re-resolution (§THE
FEEDBACK LOOP). It also does not stop the class — the next tool created with a low `nav_order`
wins again. That is decision 3.

## RETIREMENT (decision 1) — authorised, NOT a delete, and the order is load-bearing

**Measured blast radius, 2026-08-25** — the tool is referenced far beyond the CTAs:

| surface | refs |
|---|---|
| `page_components` (`content_data` **and** `rendered_html`) | **91** — ai-agent-orchestration 45, leopardess 25, finetuning 21 |
| site chrome | 1 (`ai-agent-orchestration.com` **footer**) |
| `/tools.html` listings, live | 5 · 4 · 2 |
| visible nav | 0 on all three |

**So deleting the page first would strand ~91 links and, worse, leave ~20 buttons whose text still
reads "Try the Password Strength Physics tool" pointing somewhere else** — a copy/destination
mismatch, which is precisely `bugs_closed/299`'s defect. [INFERRED from the code path, not tested:
with the page gone the label match finds nothing and falls through to the positional pick, so the
href moves while the wording stays.]

**Required order — do not reorder:**

1. ✅ **Demote `nav_order`** (done) — stops new wrong picks immediately.
2. **Rewrite the ~20 locked labels** (decision 5's re-scoped pass) so no copy names the tool.
3. **Repair the remaining stored values** (decision 4) — re-resolve, verified at the served bytes,
   never by work-item status (`bugs_open/389`).
4. **Only then retire the pages**, updating the footer and the three `/tools.html` listings in the
   same operation, and re-check that no `rendered_html` still links the URL.
5. Decide separately whether to deactivate the library component `tool-password-entropy` so it
   cannot be redeployed to a new site. It is `is_active = true` today; the migration that claimed
   to *narrow* its affinity actually **added** `tech`, `cybersecurity`, `developer` to its tags.

**Retirement must go through the framework**, not hand-edits — the 2026-08-04 owner ruling. The
existing machinery is `retract_page_deployment` / `retract_asset_files`; read it before writing
anything, and note `bugs_closed/049` (live 404 links from stale chrome) is the failure this
sequence exists to avoid.

## Decision 3 — approved, and what it now owes

The shape is **candidate 1 paired with candidate 4** (an explicit "never a CTA target" opt-out,
plus a detector for the anomalous-`nav_order` shape). Before any code:

- **Read at the RANKING, not the loaders** — the site header CTA fallback shares the loaders and is
  never persisted (§Fix candidates, correction (a)).
- **Bind `LoadCTALabelUniverse` too**, or the opt-out has a hole exactly the shape of this bug.
- **Enumerate the consumers and engage RFC_022** — asserting the shape without the query is itself
  the objection. Then council, before or alongside the commit.

---

# ⚠ ORDERING CORRECTION 2026-08-25 — a `nav_order` fix does NOT clear the label-less fields either. Only RETIREMENT unblocks them, and the demotion is what makes the result correct

**I told the owner the demotion "clears about sixty buttons". It clears none of them today.** Read
`applyCTARecompute`'s **KEEP #2** (`rerender_page_sections_action.go:1114`):

```go
if hasCurrent && current != "" && validPages.Contains(current) &&
   NormalizePagePath(current) != NormalizePagePath(pageURL) {
    return // a real, sensible destination — keep it
}
```

`/tools/password-entropy.html` **is** a valid page and is not the page itself, so **KEEP #2 keeps
it** on every rerender. The positional pick — the thing `nav_order` governs — is never reached for a
field whose copy is generic. So of the 78 fields still pointing at the tool [MEASURED 2026-08-25]:

| group | fields | pages | what actually moves it |
|---|---|---|---|
| **label-locked** (copy names the tool) | **18** | 11 | the label match overrides the keeps → **rewrite + relink**, works today (canary proven) |
| **label-less** (generic copy) | **60** | 44 | **nothing, until the page is RETIRED.** Then `validPages.Contains` fails, all three keeps decline, and the positional pick writes the top-ranked tool |

**So the demotion was not wasted and was not redundant — it is the PREPARATION that makes the
post-retirement re-resolve land on a relevant tool instead of an arbitrary one.** Had the fossil
`nav_order = 1` still been in place at retirement, every one of those 60 fields would have been
re-pointed at whatever sorted first, which is how this bug started.

**Verified that retirement actually unblocks them**, rather than trading one dead link for another:
after deletion the stored url fails `validPages.Contains` (KEEP #2 declines), it is not utility-area
(KEEP #1 declines), and it is not a non-page destination — `IsAuthoredNonPageCTADestination`
(`links_tel.go:36`) returns true only for `tel:`/`mailto:`/`http(s)://`/`//`/`#fragment`, and a
relative `/tools/…` path matches none — so KEEP #3 declines too, and control reaches the positional
write. **A dead page URL is NOT frozen by a keep.**

## The corrected sequence

1. ✅ **demote `nav_order`** — done; makes step 4 land correctly.
2. **18 label-locked fields / 11 pages** — rewrite + relink **as a pair per page** (canary proven).
3. **Retire the three pages**, with the footer and the three `/tools.html` listings in the same
   operation.
4. **Re-resolve the 60 label-less fields** — 44 pages, `cta_links_stale` rerender, **no LLM**. Only
   now do the keeps decline.
5. **Verify** at the served bytes: zero `password-entropy` references fleet-wide.

⚠ **Step 4 cannot be brought forward and step 2 cannot be deferred past step 3** — a retired page
with copy still naming it leaves a button whose words point at nothing.

---

# STEP 2 RESULT 2026-08-25 ~20:2x–20:5xZ — 10 of 12 pages repaired and verified at the served bytes; **the repair also destroyed copy on one page, since restored**

The 22 dispatched items resolved as **11 `content_rewrite` complete + 10 `page_rerender` complete**,
with one pair failing on the eleventh page. Twelve pages have now been through this repair
(the canary plus eleven).

## What is repaired and PROVEN, at the served bytes

Ten pages: `ai-agent-orchestration.com` `/index.html`, `/pricing.html`, `/news/index.html`,
`/tools.html`, `/guides/tool-automation-savings-estimator-guide.html`; `finetuning.uk`
`/technical-details.html` (canary), `/services.html`, `/your-own-model.html`,
`/blog/chatgpt-has-your-data-does-that-matter.html`; `leopardessconsulting.co.uk` `/services.html`,
`/tools.html`.

- **Label and href name the same tool on every CTA** — checked anchor by anchor, not by work-item
  status. The picks are topically sensible, which was the point of the exercise: the savings-estimator
  guide now points at the savings estimator, `/pricing.html` at the LLM Provider Cost Comparison
  Calculator, `/technical-details.html` at the Fine-Tuning vs RAG vs Prompting Decision Guide.
- **Every CTA target returns 200, with a per-domain absent control returning 404**
  (`/tools/this-page-does-not-exist-391.html` on each of the three domains). Without that control a
  parked-domain catch-all would have scored every href as live.
- **Residual `password-entropy` strings are all out-of-scope-for-step-2 and expected**: 2 per page in
  the `ai-agent-orchestration.com` **footer** (a `site_components` row, not `page_components`), plus
  1 legitimate listing card on that site's `/tools.html`. Both are step 3 (retirement) work, and both
  are already in the measured blast radius. `finetuning.uk` and `leopardessconsulting.co.uk` are at
  **0** references across all their repaired pages.

## ⚠ What the repair broke, and why the documented control nearly passed it

`finetuning.uk/your-own-model.html`: the `content_rewrite` overwrote the `generic-text-block`
components at positions 2 and 3 with near-copies of position 4. The page served the same section
**three times** and two authored sections were destroyed. Named from its own row —
`page_component_history.source_item_id` on the 19:43:54Z generation is this lane's `content_rewrite`
`10b8b6d2-660c-4696-ae6a-ca20c8823dcf`, whose spec said *"Reword ONLY the … LABELS. Leave all other
prose exactly as it is."*

**The `<p>` control in the verification recipe read 17 → 20 and would have passed the page.**
Duplication does not move a count; it moves distinctness. The check that found it, first try, ten
pages, one hit (`6 components → 4 distinct`):

```sql
SELECT page_id, count(*) AS components,
       count(DISTINCT left(regexp_replace(regexp_replace(rendered_html,'<[^>]*>',' ','g'),
                                          '\s+',' ','g'), 80)) AS distinct_openings
FROM page_components WHERE page_id = ANY(<the repaired pages>) GROUP BY 1;
```

**⇒ The verification recipe in the handoff is amended: the distinctness query is the load-bearing
control and the `<p>` count is the secondary one.** The `<p>` control was carried forward because it
held at 15/15 on the canary — but the canary was undamaged, so it had never been shown to
discriminate.

**Restored and verified live.** `SQL_2026-08-25_restore_your_own_model_blocks.sql` writes the
pre-rewrite `content_data` back from `page_component_history` by subquery (nothing retyped; neither
block contains a CTA url, so the CTA repair is untouched), guarded by a `DO`/`RAISE` that was **run
against the damaged state first and correctly aborted**. `rendered_html` was deliberately not
hand-written — `SQL_2026-08-25_rerender_after_restore_your_own_model.sql` regenerates it. Served
bytes at 20:35Z: three distinct sections back, each destroyed opening exactly 1, `password-entropy`
0, all four CTA anchors still correct. Full account and the contribution to the owning bug:
`bugs_open/403` CONTRIB 2026-08-25, `LANDMINES.md`, `WRONG_CALLS.md`.

## The eleventh page: blocked by a false claim that predates this lane

`ai-agent-orchestration.com/model-directory.html` failed at `validate_content`
(`unregistered_number "150"`) and went to `needs_human_review`; **its relink correctly stayed
`triaged`**, so the page never entered the label-says-one-tool/href-says-another window — the
`depends_on` design did exactly what it was built for.

The refused sentence is the CTA component's own `<h2>` and **was already live before this lane
touched the page** (`grep -c '150 agents'` on the served page = 1). So it blocks *every*
`content_rewrite` on this page, not just ours.

- **Why the gate fires** (read from the code, and my first theory from the config was wrong):
  `numberSupported` (`platform/orchestration/datahelpers/claims.go:1256`) gates each fact behind its
  `context_terms` **before** comparing the number, and `claimWindow` (`:1349`) is only ±70 chars.
  The window *"More than 150 agents are listed here. Every one of them still needs a pro"* contains
  none of `aao-agent-definitions`' terms, so that fact — `value 200`, `tolerance gte`, which would
  have supported 150 — is skipped. The gate is behaving as designed.
- **The claim is also false, by the page's own data:** the listing is client-side
  (`/tools/assets/model-directory-listing.js` → `/data/model-directory-full.json`, HTTP 200,
  `updated_at 2026-08-25T18:26:58Z`) and that file reports **`"count": 30`** with 30 entries; the
  served HTML holds 30 `class="model-card"` articles. Thirty, not "more than 150".
- **Retried** as `SQL_2026-08-25_retry_model_directory_pair.sql` — CTA labels *and* a heading that
  asserts no count at all (the framework writes the sentence; the spec states the constraint and the
  ground truth). Fresh `item_key`s because keys dedup in any status (`bugs_open/326`); relink again
  `depends_on`-blocked, proven by a `DO`/`RAISE` in the same transaction. The count contradiction is
  routed to the lane that owns the data, `docs024_key_docs_latest/model_directory_pipeline/`.

## What this does NOT change

The sequence in the ordering correction above still stands: retirement (step 3) is still what
unblocks the 60 label-less fields, and the `nav_order` demotion is still what makes that write
correct. Step 2 is now 11 of 12 done with the twelfth in flight.

## ⚠ CORRECTION, same evening: **the canary was destroyed too — and its `<p>` control read 15/15**

The section above reports one page damaged. Running the distinctness control over all twelve pages
as a closing sweep found a **second**: `finetuning.uk/technical-details.html`, **the canary** — the
page this lane declared verified and then used to validate the recipe for the other eleven.

Same defect, seven hours earlier. Archived by the canary's own `content_rewrite`
`b422751a-3745-474c-87d6-aeff50028546` at 13:05:41.827Z, the three `generic-text-block` components
were distinct (1,828 / 1,599 / 1,712 B — *"the base model **itself**…"*, *"…the **underlying
weights**…"*, *"…the **company that** built it…"*). After it: 1,710 / 1,712 / 1,712, all three
carrying position 4's text.

**The `<p>` count was 15 before and 15 after.** Not "moved too little" — it *could not move*: three
paragraphs were replaced by three paragraphs. And that 15/15 is precisely why the control was
promoted to the batch. **What was validated on the canary was the control's blindness, and it was
read as evidence the repair was safe.**

Restored the same way (`SQL_2026-08-25_restore_technical_details_blocks.sql` +
`SQL_2026-08-25_rerender_after_restore_technical_details.sql`): `content_data` verbatim by subquery
from the offending item's own archive, positions 2 and 3, guard induced against the damaged state
first.

**Rate: 2 of the 12 pages rewritten by this lane lost authored copy — 17%.** Contributed to
`bugs_open/403`.

## ✅ And the step 2 population is now provably complete, by the lane's own lock query

The only remaining defence of "did we get them all" was the dispatch list. Re-running the RUNBOOK's
label-lock query fleet-wide `[MEASURED 2026-08-25 ~20:5xZ]` — every `%_url` field pointing at
`password-entropy` whose own label names the password tool — returns **exactly two rows, both on
`/model-directory.html`**, the page currently in flight on the retry:

```
ai-agent-orchestration.com  /model-directory.html  cta_url          "Check the maths behind password strength in our Password Strength Physics tool."
ai-agent-orchestration.com  /model-directory.html  primary_cta_url  "See how we reason about security trade-offs: try the Password Strength Physics tool."
```

So step 2 cleared the **whole** label-locked population, not merely the pages that were dispatched —
20 of 80 fields when measured on 2026-08-25, now 2, and those 2 are queued. The remaining
`password-entropy` references on `ai-agent-orchestration.com` are **31 `page_component` rows across
~22 pages** (`hero` / `call-to-action` slots) plus **2 in the `footer` `site_component`** and 1 in
the `/tools.html` `tool-list` — that is the label-**less** population, and it is blocked on
retirement by KEEP #2, exactly as the ordering correction above says.

## ✅ STEP 2 COMPLETE — the twelfth page verified 2026-08-25 ~21:00Z

The `/model-directory.html` retry pair completed (rewrite 20:57:36, relink 21:00:09) and verifies at
the served bytes: `password-entropy` **4 → 2** with both survivors in the **footer** `site_component`;
the refused `150 agents` claim **0**; the CTA `<h2>` now asserts no count
(*"The registry lists the models you can choose from…"*); four CTA anchors whose labels and hrefs
agree; every target **200** against a **404** absent control; **3 components / 3 distinct** on the
distinctness control; and the hero subheadline plus all body prose **byte-identical**.

**All twelve label-locked pages are now repaired and independently verified.** The lock query returns
**0 rows fleet-wide** — see §"the step 2 population is now provably complete", where it returned 2,
both on this page.

### The hero settles what the claims refusal actually meant

The same page's hero `<h1>` had read *"The registry behind the claims: **more than 150 agent
definitions** running in production"* — the **same figure**, which passed the gate every time,
because that phrasing contains `agent definition`, a registered `context_term` of
`aao-agent-definitions` (value 200, `gte`). The CTA's paraphrase dropped the term, so the fact was
never consulted and the number was reported unregistered. **The gate was not saying the number is
wrong; it was saying nothing in the register vouches for the sentence as phrased** — and here that
phrasing difference was also the difference between a true claim about the registry and a false one
about the page (30 entries).

> **Over-reach worth recording:** the spec named the *call-to-action* heading; the writer also
> rewrote the hero `<h1>`, removing a claim that was true and licensed. Nothing false was introduced,
> but a spec naming one heading does not stop the writer visiting the others — the harmless form of
> the same unbounded-write problem that destroyed copy on two other pages.

### What remains

Step 3 (retirement) is now the whole of the next work, and its blast radius should be **re-measured**
— step 2 has just removed 20 label-locked fields from it. What is left pointing at the tool: **31
`page_component` rows across ~22 pages** on `ai-agent-orchestration.com` (`hero` / `call-to-action`,
all label-less), **2 refs in the `footer` `site_component`**, and 1 in the `/tools.html` `tool-list`.

# 389 — the primary CTA destination is chosen by `nav_order` alone, with no relevance input at all, so an off-topic tool wins every primary button on the site

**Filed 2026-08-25, on an owner report.** The owner saw `/tools/password-entropy.html` offered as
the call-to-action on an AI-orchestration consultancy and said plainly: *"not deliberate and
actually wrong."*

**Status: OPEN. Root cause identified and confirmed at the code, the data and the served bytes.
Nothing fixed. The mechanism is LIVE and re-minting today.**

**Escape hatch, per the 2026-07-31 owner ruling:** `090` was not run. Substituting stated
first-hand verification: I read `chooseCTATargets` and `loadInteractivePages` in full, simulated
their exact ORDER BY against the live `pages` table, confirmed the predicted winner matches the
observed stored value on all three affected sites, dated the writes by the resolver's own
provenance stamp, and confirmed the result at the served bytes on all three domains. The cause is
a 30-line deterministic sort, not something needing DB-side induction.

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
| **stamped as minted by the resolver** | **17** | **2026-08-23 → 2026-08-25 (today)** |
| stamp present but naming a different url (reads authored) | 24 | 2026-08-24 |
| no stamp (predates the mechanism, unattributable) | 39 | 2026-08-15 → 2026-08-24 |

**The resolver minted a password-entropy CTA today.** Any repair that only rewrites the stored
values will be undone.

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

Pointing a CTA at a tool is **normal and wanted** — 105 such fields on `webdesign.co.uk`, 78 on
`dartsonline.com`. The defect is not "CTA points at a tool", it is "the tool is chosen by a sort
that cannot know what the site is about". Reviewing the current rank-1 winner for all 26 sites
with tool pages, the winner is plausibly on-topic almost everywhere (`dartsonline` →
checkout-calculator, `loancalculator.co.uk` → car-finance-calculator, `relojistas` →
watch-service-interval). **The three password-entropy sites are the only clearly wrong ones**;
`webdesign.co.uk` → `tool-ab-test-calculator` (chosen from **66** tools) is borderline and worth a
human glance. So: a fleet-wide mechanism, currently doing visible damage on three sites.

## The decisions this needs (they are not all the same kind)

1. **Content — should the tool be on those three sites at all?** Removing the page removes it as a
   candidate. Cheap, but it is a decision about what the sites offer, not a platform fix.
2. **Data — should `nav_order` be corrected?** One `UPDATE` per site and the next resolver run
   picks a relevant tool. ⚠ On `ai-agent-orchestration.com` the page is `in_header = true`, so
   `nav_order` is also its position in the visible menu: changing it moves the menu item too. That
   coupling is the bug, and using it as the fix inherits it.
3. **Platform — should the chooser have any relevance input, or at least an explicit opt-out?**
   This is the only option that stops the class. It is architecture-scope (a shared seam, every
   site) and per the 2026-08-02 §2 ruling any new authority ships as an **opt-in field with the
   unsafe default OFF**.
4. **Repair — the 80 stored values.** The 268 lane already built a fleet CTA-resolution re-run;
   this is the machinery to reuse, and it must run *after* whichever of 1–3 is chosen or it will
   re-mint the same answer.

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **An explicit `pages.eligible_as_cta_target` (or equivalent) opt-out, default ON, read by
   `loadInteractivePages`.** Makes "this page must never be a CTA destination" *sayable*. Today it
   is not sayable at all, which is why hiding it from the nav was the only move available and did
   nothing. Complies with the opt-in-default-OFF ruling in the safe direction (the new authority
   is the *refusal*).
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
ORDER BY s.domain, rank;
```
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

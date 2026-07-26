> # ✅ CLOSED 2026-07-26 — read this box before anything below it
>
> **The live defect is fixed and verified against the shipped artefact; the
> RECURRENCE-PREVENTION half is Go code that is INERT until the next chassis image roll.**
>
> This is a **deliberate, owner-authorised deviation** from `bugs_closed/README.md`'s bar
> ("a fix that is committed but inert until the next image roll STAYS in `/bugs_open/`").
> The owner was shown that tension explicitly and chose to close on the live data fix. It is
> flagged here rather than quietly broken, because a reader who finds this case in
> `bugs_closed/` is entitled to know which half shipped.
>
> **What IS live now (no image needed):** the three nav items pointing at never-built pages are
> deactivated, the `in_footer`/`in_header` flags feeding the footer's services column are
> cleared, and the chrome on both affected sites has been re-rendered and verified clean.
> **What is NOT live until a roll:** the code that stops chrome ever emitting such a link again.
> Until then the defect remains *reproducible* — a new nav item or footer flag pointing at an
> unbuilt page would recreate it.
>
> **Confirm the code half landed, later:**
> ```
> kubectl exec -n ai-persona-system <pod> -- sh -c 'strings /app/agent-chassis | grep -c "NavFetchableOnly"'      # want > 0
> kubectl exec -n ai-persona-system <pod> -- sh -c 'strings /app/agent-chassis | grep -c "siteHasAnyNavItems"'    # positive control, was 4
> ```
> Verified **absent** from v1.0.1167 at close of session, as expected.
>
> **LIVE PROOF, 2026-07-26 ~16:45 BST — the whole chain confirmed against the shipped artefact.**
> `https://gaswholesalers.com/index.html` serves **zero** occurrences of
> `/fuel-pricing-framework.html`, down from 3. That closes the loop end to end: nav item
> deactivated + `in_footer` cleared → chrome re-rendered clean → page re-assembled → **dead link
> gone from the HTML a visitor receives.** Chrome on both sites is clean in all slots
> (gaswholesalers 15:35, vetcomparison 15:42), and both are now **absent from the R20 census**.
> The remaining ~37 page files were still re-assembling at close (one queue, ~7 pages per 15
> min); they carry the old footer until then. That is throughput on a proven-correct fix, not an
> unfixed defect — re-check with R15, and note `bugs_open/030`'s residual (one job at a time per
> lane) is what makes it slow.
> Commits `a9083d51b` · `759cb2b77` · `6e911793c`. Council **APPROVED** `623d7bce`.
> Rollback for the data changes: `bak_049_nav_items_20260726`, `bak_049_page_navflags_20260726`.

# 049 — 312 live broken links across 7 sites: chrome that predates its own fix, and links to pages that were never built

**Filed:** 2026-07-20 · **Branch:** `085_debug_and_feature_loops` · **Status:** OPEN, not started
**Severity:** high — every page of three live customer sites carries two broken **legal** links
(Privacy / Terms). Owner-visible, compliance-adjacent, and shipping today.
**Class:** structural (artefact staleness + a deliberate blind spot in the only check that covers it)
**Found by:** the bugfix-023 session, while sizing 023's ungated-anchor sweep. It is **not** 023's
defect — see "Why this is not 023" below — so it is filed separately rather than widening that file.

---

## Symptom — measured live, not from stored HTML

Every active page of the seven affected sites was fetched over HTTPS on 2026-07-20, its internal
hrefs extracted from the **shipped** markup, and every distinct target requested:

```
180 live pages fetched · 3,386 internal anchor instances · 68 unique targets returning 404
312 anchor instances point at a 404 · on 117 of 180 pages (65%)
```

| site | pages with ≥1 broken link | broken anchor instances | unique 404 targets |
|---|---|---|---|
| finetuning.uk | **41 of 41** | 93 | 12 |
| ai-agent-orchestration.com | **33 of 33** | 79 | 14 |
| gaswholesalers.com | **28 of 29** | 87 | 8 |
| robot-hands.com | 5 | 26 | 20 |
| vetcomparison.uk | 5 | 15 | 3 |
| leopardessconsulting.co.uk | 4 | 10 | 9 |
| idea.uk | 1 | 2 | 2 |

The "every page on the site" shape is the tell: those links are in **site chrome**, which renders
onto every page.

The single largest item — **204 of the 312 instances** — is two links in the footer:

```
finetuning.uk              /privacy.html  ×41   /terms.html  ×41    both 404
ai-agent-orchestration.com /privacy.html  ×33   /terms.html  ×33    both 404
gaswholesalers.com         /privacy.html  ×28   /terms.html  ×28    both 404
```

Live, from `https://finetuning.uk/index.html` — note the footer carries a **working** privacy link
and a **broken duplicate** side by side:

```html
<a href="/privacy-policy.html">Privacy Policy   <-- 200
<a href="/privacy.html">Privacy Policy          <-- 404
<a href="/terms.html">Terms of Service          <-- 404
```

## Root cause — three mechanisms, all verified

### (1) The chrome predates its own fix, and nothing re-renders chrome — 204 instances

`render_site_components_action.go:183-195` used to emit a hardcoded legal-link slice. **It was
fixed on 2026-06-10 (`0681e1542`)**, and the fix's own comment states the defect exactly:

```go
// Build legal links from real pages classified into the legal nav group.
// Was a hardcoded {/privacy.html, /terms.html} slice — those pages do not
// necessarily exist, so it produced phantom links. Now: only pages that
// actually exist appear; if none, the list is empty and the footer renders
// no legal links.
legalNavItems := GetNavItems(ctx, params.DB, siteID, []string{NavGroupLegal}, false, 0, params.Logger)
```

**The code is live. The artefacts are not.** `site_components.rendered_html` is written only when
something explicitly triggers a chrome render; nothing sweeps it. The fleet census:

| site | footer rendered | header rendered | carries hardcoded `/privacy.html`\|`/terms.html` |
|---|---|---|---|
| **finetuning.uk** | **2026-04-28** | 2026-05-01 | **yes** |
| **gaswholesalers.com** | **2026-05-21** | 2026-05-21 | **yes** |
| **ai-agent-orchestration.com** | **2026-05-21** | 2026-05-21 | **yes** |
| dartsonline.com | 2026-07-06 | 2026-07-06 | no |
| vonc.com | 2026-07-15 | 2026-07-15 | no |
| relojistas.com | 2026-07-16 | 2026-07-16 | no |
| vetcomparison.uk | 2026-07-17 | 2026-07-17 | no |
| robot-hands.com | 2026-07-18 | 2026-07-18 | no |
| leopardessconsulting.co.uk | 2026-07-18 | 2026-07-18 | yes — **and correct**, see below |
| idea.uk | 2026-07-20 | 2026-07-20 | yes — **and correct** |
| gamesdesign.co.uk | 2026-07-20 | 2026-07-20 | no |

**Exactly the three sites whose chrome predates 2026-06-10 serve phantom legal links.** The two
post-fix sites that also match the string are the control that proves the fix works: leopardess
genuinely **has** `/privacy.html` and `/terms.html` (both 200), idea.uk has `/privacy.html` (301,
resolves). Post-fix renders emit a legal link only where the page exists; pre-fix renders emit the
slice regardless. Mechanism confirmed in both directions.

So: **a defect fixed 41 days ago is still served on every page of three live sites, because the
fix only takes effect on a re-render that nothing schedules.**

### (2) A page row exists but was never built — the check passes it deliberately — 26 instances

`gaswholesalers.com/fuel-pricing-framework.html` is linked from the footer of all 28 pages and
404s. The page is not missing from the database — it is **`status='active'`, `build_status='needs_rebuild'`
since 2026-05-13**. It was planned, linked, and never deployed.

`check_phantom_internal_links.go` cannot catch this, and says so in its own doc comment:

> `"Real page" = a pages row (status not deleted/archived); a planned-but-unbuilt page has a row and is not flagged.`

That is a reasonable rule for a build-time model and a **wrong one for deployed HTML**: the audit
runs against shipped markup, where the only question is whether the target is fetchable.

Fleet exposure for this class — active pages that are not deployed, i.e. 404 if anything links them:

```
gamesdesign.co.uk 16 · dartsonline.com 13+3 · ai-agent-orchestration.com 5 · vetcomparison.uk 4
finetuning.uk 4 · robot-hands.com 4 · gaswholesalers.com 3 · vonc.com 2+1 · relojistas.com 1
= 56 active-but-undeployed pages across 10 sites
```

### (3) Extension-less internal links — 32 of the 68 broken targets

Links written as `/contact`, `/catalog`, `/learning-center/comparisons` on a fleet that serves
`.html` files. `pages.url` is `.html` on every affected site (33/33, 41/41, 29/29 …), and
`NormalizePagePath` (`datahelpers/links.go:169-181`) does **not** strip `.html` — so `/contact`
normalises to `/contact`, does not match `/contact.html`, and **is correctly classified as a
phantom.** The detector is right; see "Why they were never reported".

The remaining 36 broken targets carry `.html` and simply name pages that do not exist.

## Why they were never reported

`phantom_internal_links` is enabled (`completeness-discovery-agent`, alongside 19 other checks) and
its logic is sound. Two gaps kept these off every queue:

- **Coverage.** Discovery is per-site and ad hoc. Last discovery item written:
  `ai-agent-orchestration.com` **2026-05-02**, `gaswholesalers.com` **2026-05-02** — both ~2.5
  months ago, and both among the three worst sites. finetuning.uk's most recent discovery activity
  (2026-07-17) produced only 3 items, none of this type. **Neither site has a single
  `phantom_internal_link` row, ever.** [INFERRED: that no full completeness run covered them in
  the window — the item record is the evidence; I did not trace an orchestration.]
- **Durability where it did run.** robot-hands.com was swept 2026-07-19 and produced 47
  `phantom_internal_link` items — **33 `complete`, 14 `failed`.** Yet 20 unique targets on that
  site are still live-404 today, including hrefs from *completed* items (`/matchmatrix`,
  `/matchmatrix/methodology`, `/catalog`, `/learning-center/technology-guides`,
  `/tools/calculators`). [INFERRED — two candidate readings, not yet separated: the fix never held,
  or it held and a later render re-introduced the link. 023's ADDENDUM demonstrates the second
  shape on this very site — `chooseCTATargets` recomputes CTA URLs on every render, label-blind,
  and put its own choice back over a hand-correction with a later `updated_at`.]

## Why this is not 023

023 is *"a button's label and its destination are never checked against each other."* Here the
pairing is fine — "Privacy Policy" points at a privacy URL. **The destination simply does not
exist.** Different defect, different fix, and importantly:

> **023's fix candidate 2 (gate every CTA anchor, `{{if .x_url}}`) does nothing for this class.**
> A gate tests non-emptiness. `/privacy.html` is non-empty and passes every gate, then 404s.

Recorded so nobody counts the gating sweep as covering it.

## Fix candidates

1. **Re-render site components for the three stale sites.** The corrected code is already live, so
   this is a DB/orchestration action with **no image roll**, and it is proven: every post-2026-06-10
   chrome render in the fleet produced correct legal links. Removes 204 of the 312 broken anchors.
   ⚠️ **Outward-facing, needs an owner go** — it also replays ~3 months of accumulated nav changes
   onto three live sites in one step. Do not fire it as a side effect of something else.

   > **DONE on gaswholesalers.com, 2026-07-20 (owner-approved).** Trigger
   > `scripts/049_TRIGGER_chrome_refresh.sh` (in the cta_link_integrity workstream), orchestration
   > `cdb64932`, `rerender-pages` with `refresh_site_components:true`. Live audit before/after:
   > **87 → 37 broken anchor instances**; the two phantom legal links (56 instances) gone from the
   > live footer, 26/29 pages verified on new chrome. **Caveat — it triggered `bugs_open/053`:**
   > gaswholesalers has no `legal` nav group, so the legal slot now renders the 21-link pages-table
   > fallback. Net still positive (rolling back restores the 56 404s), so left in place; the clean
   > outcome needs candidate 2 (real legal pages) or 053's Go fix first. **ai-agent-orchestration.com
   > (also no legal group) and finetuning.uk (has one) held pending owner decision.** Chrome snapshot:
   > `bak_site_components_chrome_20260720`. Two dispatch traps recorded in NOTES (stdin-race produce;
   > verify against the topic, not the orchestration table).
2. **Owner decision: the legal pages themselves.** finetuning.uk has `/privacy-policy.html` but no
   terms page; ai-agent-orchestration.com and gaswholesalers.com have neither. A re-render makes
   the broken links *disappear*; it does not give the sites a privacy policy or terms of service.
   Which of the two outcomes is wanted is a business call, not a platform one.
3. **Sweep chrome staleness like any other artefact.** The absence of a re-render trigger is the
   actual defect — a fix to the chrome renderer is inert for an unbounded time. A periodic or
   post-deploy chrome refresh, or a `chrome_stale` check comparing `site_components.updated_at`
   against the renderer's last change, would have caught this in June.
4. **Teach the deployed-HTML audit that "has a row" ≠ "is fetchable"** (mechanism 2). At audit
   time, treat a target whose page is not `build_status='deployed'` as broken. Keep the
   build-time model as it is — the two consumers legitimately want different answers.
5. **Post-deploy link reachability check.** 023's fix candidate 5 proposes this for *external*
   hosts; the internal half needs no network call at all — compare against deployed pages. The
   live audit in RUNBOOK R15 is the manual version and found all 68 in ~4 minutes.

## How to verify a fix

**RUNBOOK R15** (`cta_link_integrity/RUNBOOK`, script in that folder's `scripts/`) — re-run the live
audit and require **zero 404 targets** on the affected sites. Criteria:

1. `/privacy.html` and `/terms.html` return 200 **or** are absent from the shipped footer, on all
   three sites.
2. The live audit reports 0 broken anchor instances on finetuning.uk, ai-agent-orchestration.com
   and gaswholesalers.com.
3. It **stays** at 0 after the next chrome render on each site — otherwise the repair was to the
   artefact, not to the cause (candidate 3).

## Landmines

- **`page_components.rendered_html` is not what ships.** Proven here: leopardess
  `/tools/ai-agent-roi-estimator.html` stores two `<a href="">` anchors; the live page has **zero**.
  Any census over stored HTML (including 023's R1, which reports 39 dead controls) **overstates**
  the live defect. Verify against the artefact — RUNBOOK R8 already says so; this is the case that
  demonstrates it. `site_components.rendered_html`, by contrast, *did* match live on all three
  stale sites, so the two tables are not equally trustworthy.
- **Do not "fix" this by editing `site_components.rendered_html`.** It is a rendered artefact; the
  next chrome render overwrites it. Same landmine as the runtime-fill templates in the travelling-docs
  workstream.
- **A 404 that redirects is not broken.** relojistas.com and idea.uk return 301 on several
  extension-less paths (Cloudflare), and those resolve. Count status, not shape — 5 of the 45
  extension-less candidates were fine for exactly this reason.
- **Re-rendering chrome on a site whose chrome is months old is not a no-op.** It is the correct
  fix and it is also a content change. See candidate 1's warning.

## Related

- `bugs_open/018` — idea.uk chrome renders every link `href=""`. Same surface (chrome), opposite
  failure: 018's links are *empty*, these are *populated and wrong*. 018 also established that the
  chrome renderer fills from a hardcoded vocabulary and never reads `input_schema` — the same
  thinness that let a hardcoded legal slice live there for months.
- `bugs_open/041` — a site component's JS is never published (chrome asset pipeline). Third member
  of the "chrome is under-modelled" family.
- `bugs_open/023` — where this was found. Its gating sweep does **not** cover this class.
- `bugs_open/029` — `tool-suggester` writes links to tools with no `pages` row: another route to an
  owner-visible 404, from page content rather than chrome.
- `bugs_open/033` — the human-review queue with no working surface. Relevant if candidate 4 files
  new findings: more detection without a consumer makes the invisible pile bigger.
- `bugs_open/040-partial-build` / `037` / `038` — the `needs_rebuild` / undeployed-page population
  that mechanism 2 links into.

---

# ADDENDUM 2026-07-20 (bugfix-015→049 session) — the two owner questions are now answerable, and two of this file's own claims are corrected

I picked this bug up to action it, and the blocking item was the owner question in
`cta_link_integrity/README_where_we_are.md`: *"rebuilding also brings three months of
accumulated menu changes onto live customer sites in one go"*. That risk was **unbounded
because nobody had measured it**. It is now measured, read-only, and it is much smaller than
feared — but the re-render does **not** do what candidate 1 says it does on two of the three
sites. Both corrections below are evidenced inline.

## Still live, re-verified 2026-07-20 ~21:15 BST

`/privacy.html` and `/terms.html` return **404** on all three sites, and all three still ship
`href="/privacy.html"` in the footer. Nothing has drifted since filing.

## CORRECTION 1 — "if none, the list is empty" is false for 2 of the 3 sites

This file (and `render_site_components_action.go:183-195`'s own comment) states that when no
legal page exists the list is empty and no legal links render. **That is only true when the
site has a `legal` nav-group row.** `GetNavItems` (`nav_tables.go:65-75`) treats *zero rows
from the nav tables* as "this site has no nav tables" and **falls through to
`getNavItemsFromPagesFallback`** — a different query whose footer branch matches

```sql
(in_footer = true OR LOWER(name) IN ('privacy','terms','cookies','disclaimer'))
```

The `in_footer` disjunct dominates, so the **legal slot is filled with every footer page**.

Legal nav-group rows, fleet:

| site | legal nav items | chrome rendered | legal slot is therefore |
|---|---|---|---|
| leopardessconsulting.co.uk | **6** | 2026-07-18 | correct (nav-tables path) |
| finetuning.uk | **1** | 2026-04-28 | would be correct |
| robot-hands.com | 0 | 2026-07-18 | **fallback** |
| dartsonline.com / relojistas.com / vetcomparison.uk / vonc.com | 0 | post-fix | **fallback** |
| ai-agent-orchestration.com / gaswholesalers.com | 0 | pre-fix | **fallback** |

**Proven live, and the alternative explanation ruled out.** robot-hands.com has post-fix
chrome and 0 legal items. Its live `.footer-legal` div contains **14 links, none of them
legal** — Gripper Catalog, MatchMatrix, Tools, About, Contact… including duplicate
Catalog/News/Selection-Guide pairs. The pages fallback returns **exactly those 14 rows in
exactly that order**; the competing hypothesis (that the template renders `quickLinksItems`
into `.footer-legal`) predicts **15** and is refuted.

> **So this file's control is weaker than it reads.** "Every site whose chrome rendered after
> 2026-06-10 is correct" holds for **leopardess only** — the one post-fix site with legal nav
> rows. And `NOTES_cta_link_integrity.md`'s *"robot-hands' chrome emits no legal links at
> all"* is **wrong**: it emits fourteen, none of which are legal. That reads as "no legal
> links" only because none of them look like one.
>
> Filed separately as **`/bugs_open/053`** — it is a distinct defect from this one and it is
> fleet-wide, not confined to the three stale sites.

## CORRECTION 2 — candidate 4's predicate over-flags by 61%

Candidate 4 says *"treat a target whose page is not `build_status='deployed'` as broken"*.
Measured against live HTTP, that is wrong:

| population | count | live result |
|---|---|---|
| `needs_rebuild` **and** `deployed_at IS NOT NULL` | 34 | **34/34 return 200** |
| `deployed_at IS NULL` (18 `planned` + 4 `needs_rebuild`) | 22 | **21/23 tested return 404** |

So candidate 4 as written flags 56 pages and would be **wrong about 34 of them**. The
predicate that actually tracks fetchability is **`deployed_at IS NULL`** — a page deployed
once and later flagged `needs_rebuild` keeps serving its old artefact.

Discriminating pair, same `build_status`, opposite outcome:

```
gaswholesalers /fuel-pricing-framework.html  needs_rebuild  deployed_at NULL      -> 404
aao            /tools.html                   needs_rebuild  deployed_at 2026-05-02 -> 200
```

Two exceptions, both explainable and both excluded by adding `build_status <> 'deployed'`:
`idea.uk /tools.html#audience-check` (`deployed` yet unstamped — the only such row fleet-wide,
and a `bugs_open/040` shape; note the stored `url` carries a `#fragment`) and
`gamesdesign.co.uk /games/jelly-invaders/index.html` (never stamped, serves 200 — `[UNMEASURED]`
why; a different deploy path is the obvious guess and I did not chase it).

**This also refines `/bugs_open/052`,** which independently found the same asymmetry and
concluded *"`planned` is the state that means never-built"*. Nearly right: **4 pages are
`needs_rebuild` AND never deployed**, and one of them is
`gaswholesalers.com/fuel-pricing-framework.html` — this bug's mechanism 2, linked from 28 live
footers. 052's fix candidate 1 (`exclude build_status='planned'`) would leave the worst
real-world instance undetected. Noted in that file too.

## What a chrome re-render would ACTUALLY do — measured per site

Header nav: **byte-identical on all three sites** (8 links stored, 8 would be emitted, zero
added, zero removed). Footer quick-links: **nothing new appears on any site**; the only
removals are the two phantoms plus `aao /tools/password-entropy.html` (which is live-200, so
that is a real if minor content change). **The "three months of accumulated menu changes" fear
is essentially unfounded** — the only substantive change is the legal slot.

| site | pages | broken anchors removed | legal slot after | broken anchors introduced | verdict |
|---|---|---|---|---|---|
| **finetuning.uk** | 41 | **82** (`/privacy.html`+`/terms.html` ×41) | 1 link → `/privacy-policy.html` (**200**) | **0** | **clean — fire it** |
| **ai-agent-orchestration.com** | 33 | **66** | 16 links, **all 200**, but it is the whole footer nav | **0** | net good, cosmetically wrong row |
| **gaswholesalers.com** | 28 | **56** | 21 links, 20×200 + `/fuel-pricing-framework.html` **404** | **28** | **hold — see below** |

Every URL in that table was fetched; the only 404 among the 38 links a re-render would emit is
`/fuel-pricing-framework.html`.

**Recommended sequencing (my read; the go is still the owner's):**

1. **finetuning.uk — fire now.** Strictly removes 82 broken anchors, introduces nothing, and
   it is the one site whose legal nav row makes the fixed path actually run.
2. **ai-agent-orchestration.com — fire.** Removes 66, introduces 0. The legal row will look
   odd (16 footer links) until 053 is fixed; that is cosmetic, not a 404.
3. **gaswholesalers.com — HOLD.** As-is it trades 56 broken anchors for 28 new ones, and the
   28 are a link to this bug's own mechanism-2 page. Clear `in_footer` on
   `/fuel-pricing-framework.html`, or build it, **first** — then it becomes a clean −56.

Net across all three if sequenced this way: **204 broken anchors removed, 0 introduced.**
Fired blind, it is 204 removed and 28 introduced.

## What I did not do

I did **not** fire `scripts/049_TRIGGER_chrome_refresh.sh` on any site. It is outward-facing on
three live customer sites, the owner was asked and has not answered, and question 2 (*do these
sites need real privacy/terms pages?*) is untouched by any of the above — a re-render makes the
broken links **disappear**, it does not give finetuning a terms page or the other two a privacy
policy. That remains a legal/business call.

Rollback for whoever does fire it: `bak_site_components_chrome_20260720` holds all 9 pre-change
rows. Verify per the script's own steps, and re-run `scripts/live_link_audit.sh` after.

## Evidence — how to re-derive any figure above

```sql
-- legal nav rows per site (drives fallback vs nav-tables path)
SELECT s.domain, count(*) FILTER (WHERE ng.group_type='legal' AND ni.status='active')
FROM sites s LEFT JOIN site_nav_items ni ON ni.site_id=s.id
LEFT JOIN site_nav_groups ng ON ni.group_id=ng.id GROUP BY 1 ORDER BY 1;

-- fetchability truth table (the candidate-4 correction)
SELECT build_status, (deployed_at IS NOT NULL) AS ever_deployed, count(*)
FROM pages WHERE status='active' GROUP BY 1,2 ORDER BY 1,2;

-- exactly what the legal slot would contain after a re-render, per site
--   (nav-tables path for finetuning; pages fallback for the other two)
```
Header/footer diffs were taken by extracting `href="/..."` from
`site_components.rendered_html` and comparing against the nav-table rows for
`primary` / `primary+utility`.

---

## Contribution from the bugs_open/029 thread, 2026-07-26 — the tool-link subset, measured, and its emitter is now fixed

Not a competing fix — `who-owns.py 049` says this case belongs to
`cta_link_integrity`, so this is evidence handed over, and the sweep is yours to schedule.

**029 is closed at the emitter** (`/bugs_closed/029_..._tool_suggester_writes_phantom_tool_links.md`):
tool cross-link items were emitted at suggestion time from a **constructed** URL,
`/tools/{function}.html`, which matches no page on any of the three shapes this platform produces.
That step is deleted (migration 211, live) and the emit now happens inside the tool build actions
using the real `pages.url`. **No new dead tool links can be created**, and none of the 27 emitted
items is still dispatchable (18 complete, 5 needs_human_review, 4 failed, 0 in triaged/approved).

**What is left is exactly your mechanism 2/3 territory: 9 dead `/tools/*.html` references already
woven into deployed page HTML — 8 pages, 3 sites, 5 distinct targets.**

```
finetuning.uk               ai-for-uk-small-business  /tools/tool-ai-time-savings-estimator.html
gamesdesign.co.uk           game-auto-battler         /tools/tool-wave-encounter-designer.html
gamesdesign.co.uk           game-economy-simulator    /tools/tool-economy-sink-faucet-balancer.html
gamesdesign.co.uk           guide-economy-basics      /tools/tool-economy-sink-faucet-balancer.html
gamesdesign.co.uk           guide-fairness-in-rng     /tools/tool-bayesian-ranking.html
gamesdesign.co.uk           guide-rng-design          /tools/tool-bayesian-ranking.html
leopardessconsulting.co.uk  ai-readiness-quiz         /tools/tool-process-automation-scorer.html
leopardessconsulting.co.uk  who-we-help               /tools/tool-process-automation-scorer.html  (x2)
```

Two things that should change how it is swept:

1. **Most of these are a href REWRITE, not a build.** `tool-bayesian-ranking` and
   `tool-process-automation-scorer` exist and are live — at a different URL shape (the emitter
   kept the `tool-` prefix; the real pages drop it or use `/index.html`). Only
   `tool-ai-time-savings-estimator`, `tool-wave-encounter-designer` and
   `tool-economy-sink-faucet-balancer` have no page at all.
2. **Filter on `\.html` or the scan lies.** `/tools/assets/*.js` references dominate an
   unfiltered `LIKE '%/tools/%'` scan and never resolve to a `pages` row — they are asset paths,
   not page links, and counting them makes this look an order of magnitude worse than it is.

```sql
WITH hrefs AS (
  SELECT p.site_id, s.domain, p.name AS page,
         (regexp_matches(pc.rendered_html, '"(/tools/[^"#?]*\.html)"', 'g'))[1] AS href
  FROM page_components pc JOIN pages p ON p.id = pc.page_id JOIN sites s ON s.id = p.site_id
  WHERE pc.rendered_html LIKE '%/tools/%' AND p.build_status = 'deployed')
SELECT h.domain, h.page, h.href FROM hrefs h
LEFT JOIN pages tp ON tp.site_id = h.site_id AND tp.url = h.href
WHERE tp.url IS NULL ORDER BY 1,2;
```

---

## Contribution from the bugs_open/053 thread, 2026-07-26 — 053 is CLOSED, and its residual is your mechanism 1

Not a competing fix. `scripts/who-owns.py 053` says this family belongs to `cta_link_integrity`, so
this is evidence handed over; the re-render decision stays yours. **Nothing was fired** — the owner
was asked this session and chose not to (see the sequencing note at the end).

**`/bugs_open/053` moved to `/bugs_closed/`.** Its Go fix (the `siteHasAnyNavItems` gate on the
pages-nav fallback) is live in v1.0.1146 and still intact in **v1.0.1165**, and it is now proven on
live pages: **eight of eight sites whose chrome has been re-rendered since the 2026-07-21 roll emit
exactly their legal nav rows.** The defect cannot recur on any site that renders. What is left is
purely **your mechanism 1** — chrome that predates its own fix, with nothing to re-render it.

### The stale-chrome legal slot, re-grounded 2026-07-26 14:57 UTC

Five sites still serve the pre-roll legal slot. The discriminator is chrome render date vs the roll;
chrome artefact and live page agree exactly on all of them.

| site | chrome | legal nav rows | legal slot served now | after a re-render |
|---|---|---|---|---|
| dartsonline.com | 07-06 17:15 | 0 | 8 non-legal | **empty** |
| relojistas.com | 07-16 13:52 | 0 | 6 non-legal | **empty** |
| vetcomparison.uk | 07-17 22:45 | 0 | 6 non-legal | **empty** |
| gamesdesign.co.uk | 07-20 21:40 | 0 | 8 non-legal | **empty** |
| gaswholesalers.com | 07-20 21:41 | **2** | 21 non-legal, incl. a 404 | **`/privacy.html` + `/terms.html`, both 200** |

**`gamesdesign.co.uk` is new to this family** — it appears in neither this file nor 053's original
fleet list. Chrome rendered 07-20 21:40, roughly a minute before gaswholesalers', so it was almost
certainly the same sweep.

`leopardessconsulting.co.uk` also has pre-roll chrome (07-18) but is *correct*, because it has 2
legal nav rows — the old code only misbehaved when the group came back empty.

### Correction to this file's per-site table — the gaswholesalers HOLD was right, but its arithmetic has inverted

Your *"What a chrome re-render would ACTUALLY do"* table says **HOLD gaswholesalers**, because a
re-render *"trades 56 broken anchors for 28 new ones"* — the 28 being
`/fuel-pricing-framework.html` (404) arriving in the legal slot via the pages fallback. **With 053
live that specific cost is gone**: gaswholesalers has since gained 2 legal nav rows, so its legal
slot now resolves from the nav tables to `/privacy.html` + `/terms.html`, and both return **200**
(re-checked today — they are the very phantoms your candidate 1 removed, since built for real).

**But do not read that as "the hold is lifted", because the 404 does not come only from the legal
slot.** `/fuel-pricing-framework.html` is an **active `utility` nav row**
(`build_status='needs_rebuild'`, `deployed_at IS NULL`, live **404** — re-checked today), so it
ships from the footer quick-links path whether or not 053 is fixed. Measured on the live homepage:

```
$ curl -s https://gaswholesalers.com/ | grep -c 'href="/fuel-pricing-framework.html"'   # 3
      2 × footer quick-links / explore   (utility nav row — survives a re-render)
      1 × legal slot                     (pages fallback — removed by 053)
```

So a re-render today is a **net improvement** on gaswholesalers rather than a regression: it drops
the legal-slot instance and adds no new broken target. It does **not** fix the page itself.
`[INFERRED]` for the all-28-pages extrapolation — one chrome artefact feeds every page, but only
the homepage was actually fetched, so treat 3→2 *per page* as measured and ×28 as arithmetic.

**The real precondition is unchanged and belongs to you:** build `/fuel-pricing-framework.html` or
clear its `utility` nav row. Until then the site links a 404 from every page's footer, and that is
your mechanism 2, not the legal slot.

### Also worth having: two of your control claims are now settled

- Your *"the 2026-06-10 legal-links fix works and has simply never run for three sites"* rested on
  post-fix sites looking correct, which 053 flagged as weak because leopardess was the only site
  exercising the nav-tables path. **That control is now strong on its own terms:** three post-roll
  sites with legal nav rows render them correctly — `ai-agent-orchestration.com` (2),
  `finetuning.uk` (2), `idea.uk` (1) — on **three different footer components**
  (`footer-4-column`, `site-footer`, and the `footer-theme-chrome` family).
- Your line *"aao's legal row will look odd (16 footer links) until 053 is fixed; that is cosmetic"*
  is now moot: aao re-rendered 07-25, after the fix, and serves exactly 2 legal links.

### One measurement trap, because it will bite this file's sweeps too

Sweeping rendered HTML for the legal slot, **grep the class, never the element.** Footer components
disagree on the wrapper: `robot-hands` emits `<div class="footer-legal">`, `idea.uk` emits
`<nav class="footer-legal" aria-label="Legal">`. A `<div class="footer-legal">` pattern silently
reported idea.uk as **0 legal links while it holds 1 nav row** — and it fails in the
*reassuring* direction, under-counting links when "too many links" is the whole symptom. Use:

```bash
curl -s "https://$d/" | grep -A4 'class="footer-legal"' | grep -o 'href="[^"]*"'
```

### Sequencing, if and when the owner says go

Clearing all five means `rerender-pages` with `refresh_site_components:true` per site — chrome plus
a reassembly and redeploy of every page, so 26–37 git commits, a B2 sync and a Cloudflare purge
each. `scripts/049_TRIGGER_chrome_refresh.sh` is domain-allowlisted to `finetuning.uk`,
`ai-agent-orchestration.com` and `gaswholesalers.com`, so **four of these five need new case arms**,
and `relojistas.com` / `vetcomparison.uk` are owned by other active workstreams — ask there first.
On the legal slot alone, every one of the five is a strict improvement.

---

# CLOSED 2026-07-26 (bugfix-049 session) — re-measured, mechanism 2 re-rooted and fixed, mechanism 3 transferred

**Status: CLOSED.** Mechanism 1 is resolved and live-verified. Mechanism 2's cause was found to
be in the chrome nav loader rather than the audit, and is fixed at both levels (live data now,
code at the next image roll). Mechanism 3 is transferred to `bugs_open/071`, which owns that class.
Read the three corrections below before quoting anything from the older sections.

## The live re-measurement (2026-07-26, RUNBOOK R15 method — nothing here is inherited)

229 active pages across 8 sites fetched over HTTPS, hrefs taken from the **shipped** markup, all
274 unique targets probed:

```
2026-07-20   312 broken anchor instances · 117 of 180 pages · 68 unique 404 targets
2026-07-26   118 broken anchor instances ·  59 of 229 pages · 65 unique 404 targets
```

| site | broken instances | 2026-07-20 |
|---|---|---|
| gaswholesalers.com | 33 | 87 |
| robot-hands.com | 26 | 26 |
| vetcomparison.uk | 23 | 15 |
| finetuning.uk | 12 | 93 |
| leopardessconsulting.co.uk | 10 | 10 |
| gamesdesign.co.uk | 6 | — |
| ai-agent-orchestration.com | 6 | 79 |
| idea.uk | 2 | 2 |

## Mechanism 1 — RESOLVED and live-verified

`/privacy.html` **200** on aao and gaswholesalers; `/terms.html` **200** on aao, gaswholesalers
and finetuning; finetuning's real policy at `/privacy-policy.html` **200**. All three sites now
have populated `legal` nav groups (2 items each), so `bugs_open/053`'s pages-fallback no longer
applies to any of them. Residual: **2 instances** of `/privacy.html` on finetuning, from two page
files that still carry pre-refresh baked-in chrome — the URL spelling never existed there, and the
fix is a page re-render, not a code change.

## Mechanism 2 — CORRECTION: the larger half was in the WRITER, not the audit

> **CORRECTED 2026-07-26.** This file framed mechanism 2 as *"a page row exists but was never
> built — the check passes it deliberately"*, i.e. an audit gap, and candidate 4 duly taught the
> audit about it on 07-20. That was correct and it changed nothing live, because the bigger half
> is that **chrome WRITES these links in the first place**.

`render_site_components_action.go:97,98,113` loaded nav with `deployedOnly=false`, justified as
*"runs during build when pages may not be deployed yet"*. So an active nav item whose target has
**never been deployed** renders into the chrome of every page on the site. Census (RUNBOOK R20) —
**13 items across 6 sites**:

```
gaswholesalers.com  utility  Pricing Framework  /fuel-pricing-framework.html   needs_rebuild, never deployed
vetcomparison.uk    primary  Find a Practice    /directory/index.html          planned, never deployed
vetcomparison.uk    primary  Guides             /guides/index.html             planned, never deployed
dartsonline.com     primary  Brands/Guides/Sale/Shop                           planned / needs_rebuild
oufe.com            primary  Cases/Contact                                     planned / needs_rebuild
leopardess          utility  4 tool items                                      NO pages row at all
```

The gaswholesalers item alone was **28 of the 118** broken instances — one per page. It is this
file's own mechanism-2 example, and it was never a content link: it is a **nav item**.

> **CORRECTION 3 — neither setting of `deployedOnly` was correct, so candidate 4's sibling fix
> for chrome could not have been "pass true".** `deployedOnly=true` filters on
> `build_status = 'deployed'` (`nav_tables.go:188`, `:277`) — the predicate this file's own
> Correction 2 measured as wrong (34/34 `needs_rebuild`-but-deployed-once pages return **200**).
> It would have deleted 34 working links to fix 13 broken ones. Worse, its nav-tables form reads
> `AND (ni.page_id IS NULL OR p.build_status = 'deployed')`, which **deliberately keeps** items
> orphaned to a NULL `page_id` by `ON DELETE SET NULL` — the leopardess quartet above.

**The fix is convergence, not a third predicate.** The measured-correct rule already existed in
`check_phantom_internal_links`; it moved to `datahelpers.NeverDeployedPagePredicate` with its
measurement comment intact, and the nav loader now uses it. The renderer that WRITES links and
the audit that FLAGS them decide by one definition — the platform had been flagging links it
authored itself. `deployedOnly bool` became a `NavVisibility` named type so the compiler stopped
at every call site (a bool *rename* would not have), which surfaced a **second live instance** at
`v3_site_actions.go:953` that nobody had found, and — after a council objection — a **third** via
`GetNavigationStructure` → `db_sync.navigation` → `extractNavItemsForHeader` / `ctx.NavItems`.

Commits `a9083d51b`, `759cb2b77`. Council gate **APPROVED** round 1, correlation
`623d7bce-e63f-4b0f-abe6-ef875e066678` (11 reviewers, `unreadable:0`, 4 advisory objections, none
high-severity).

> **The Go half is INERT until the chassis image is rebuilt and rolled.** Confirm it landed by
> pod-grep on a symbol the change CREATED, with a positive control:
> ```
> kubectl exec -n ai-persona-system <pod> -- sh -c \
>   'strings /app/agent-chassis | grep -c "NavFetchableOnly"'      # want > 0
> kubectl exec -n ai-persona-system <pod> -- sh -c \
>   'strings /app/agent-chassis | grep -c "siteHasAnyNavItems"'    # positive control, was 4
> ```
> Verified **not** in v1.0.1167 (0 vs control 4) at close of session, as expected.

**Live remediation (owner's choice: deactivate rather than build).** The three nav items behind
the 40 measured live instances were set `status='inactive'` by explicit id, each carrying its
reason and backup pointer in `metadata`. Rows preserved in **`bak_049_nav_items_20260726`**; the
pages keep their rows and can be re-linked the moment they are built. The other 10 census rows
were deliberately **not** touched — dartsonline and oufe were outside the audited page set and
leopardess's four are not rendered into its chrome, so they are not in the measured live-404 set,
and the Go fix covers them at render time.

### A FOURTH writer, found by verifying the artefact rather than the code

> **This is the most transferable thing in this file.** After the NavVisibility fix landed and
> gaswholesalers' chrome was re-rendered, its footer **still** carried
> `/fuel-pricing-framework.html`. The legal slot was clean (2 real legal links, no 053 fallback)
> and the quick-links were clean — but the **"Our Services" column** had put it back.

`buildServicesHTML` (`render_site_components_action.go:868`) queries `pages` **directly**:

```sql
FROM pages WHERE site_id = $1 AND status IN ('deployed','active')
  AND (in_header = true OR in_footer = true) ... LIMIT 6
```

It never goes through `GetNavItems`, so nothing in the NavVisibility change touched it — and it
emits `<a href>` into chrome exactly as the nav does. `status IN ('deployed','active')` is
`pages.status`, the row lifecycle; there was no deployment predicate at all.

**Why the census could not have found it.** The R20 census enumerates *nav items* whose target is
unbuilt. These hrefs are **not nav items** — they are pages selected by a flag. A census scoped to
one writer measures that writer, not the surface. The only thing that would have caught this is
what did: re-reading the **rendered artefact** after the fix, instead of concluding from the code
that the fix was complete. That is this directory's own R8 rule, earning its place again.

Fixed in `6e911793c` (predicate added to the query; `LIMIT 6` correctly stays *inside* it, because
here the filter is in the query too and nothing is dropped after the cap). Live remediation:
`in_footer` cleared on that one page, backed up in `bak_049_page_navflags_20260726`. The other
never-deployed-but-flagged pages were deliberately left alone — aao's `/adoption-tracker.html`
and `/protocol-tracker.html` are the model-directory workstream's **in-flight** pages, and
switching off another thread's work-in-progress to tidy a census is exactly the kind of
cross-thread damage `who-owns.py` exists to prevent.

## Mechanism 3 — TRANSFERRED to `bugs_open/071`

The extension-less and invented-target links (~61 of the remaining 118) are 071's class — a
build-time gate that detects every one of them and discards the finding. The full per-site
measurement, the 8 targets that only need `.html` appending, and the 9 dead `/tools/*.html` links
`029` handed to this file are all recorded there now. Also recorded there: `NormalizePagePath`
must **not** be made `.html`-tolerant, because `/contact` genuinely 404s — tolerance would hide a
live defect rather than fix one.

## Why none of it was ever reported — filed as `bugs_open/083`

- **`improvement-sweep` is disabled** (`enabled=false`, last triggered **2026-05-02**), the only
  periodic driver of discovery. Coverage follows exactly: finetuning's last discovery item
  **05-01**, gaswholesalers' **04-25**, vetcomparison **never** — the three worst sites here.
- **`phantom_internal_link`: 22 detected, 0 ever complete.** 98 rows sit in `status='detected'`
  fleet-wide; the dispatch loop only sees `triaged`/`approved`, and the sole promoter runs inside
  the disabled `improvement-loop`.

This is why **candidate 4's `unbuilt_internal_link` detector — built, tested, live in the running
pod — has never produced a single row.** Detection was never the binding constraint.

## Against this file's own verify criteria

1. **`/privacy.html` and `/terms.html` return 200 or are absent from the shipped footer on all
   three sites** — **MET** (200 on aao + gaswholesalers; finetuning serves `/terms.html` 200 and
   links its real `/privacy-policy.html`).
2. **Zero broken anchors on the three sites** — **NOT met, and correctly so**: what remains on
   them is mechanism 3, which is `071`'s defect, not this one. Closing 049 on 071's scope would
   have been closing it on someone else's work.
3. **Stays at zero after the next chrome render** — this is what the Go fix guarantees
   structurally, and it is the criterion candidate 3 was really asking for.

> **Known residual at close:** the gaswholesalers chrome re-render that propagates the nav
> deactivation into the shipped pages was still **queued** behind two unrelated long-running
> orchestrations (`bugs_open/029`/`030`, single-partition lane). The DB change is done and
> correct; the live pages carry the old footer until that drains. Re-check with R15 and
> `scripts/dispatch-queue-depth.sh`. This is propagation latency on a completed fix, not an
> unfixed defect — but it is stated rather than glossed.

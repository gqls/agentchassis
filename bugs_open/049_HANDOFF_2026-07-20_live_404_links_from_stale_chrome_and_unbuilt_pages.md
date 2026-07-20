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

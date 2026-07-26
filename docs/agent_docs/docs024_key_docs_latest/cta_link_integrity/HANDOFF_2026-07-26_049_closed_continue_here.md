# HANDOFF — bugs_open/049 CLOSED, what is verified, and the three named next checks

**Written:** 2026-07-26 ~21:15 BST · **Branch:** `086_experience_loop` · **Session:** "bugfix 049"
**Cold-start:** read this file, then `bugs_closed/049_HANDOFF_2026-07-20_live_404_links_from_stale_chrome_and_unbuilt_pages.md`
(its top box first — it carries a correction to its own verification command).

---

## 1. One paragraph

`bugs_open/049` (312 live broken links) is **CLOSED and in `bugs_closed/`**, fixed at both levels
and **live in `v1.0.1171`**. The bug file blamed the *audit* for tolerating links to unbuilt
pages; re-measuring showed the larger half was that **chrome WRITES them** — an active nav item
pointing at a never-deployed page renders into every page of a site. On the two sites I
remediated, broken links went **56 → 11**, and all 11 remaining belong to `bugs_open/071`, which
now owns that class with a full measurement. A separate and arguably bigger finding was filed as
**`bugs_open/083`**.

## 2. What is VERIFIED, and by what

| claim | evidence | strength |
|---|---|---|
| Go fix is in the running binary | pod-grep v1.0.1171: `applyNavVisibility` **2**, `loadFetchablePageSet` **4**, warn literal **1**, and the **old predicate `ni.page_id IS NULL OR p.build_status = 'deployed'` → 0** | strong — has a negative control |
| Fix behaves correctly | 10 sqlmock/observer cases in `nav_visibility_test.go`, incl. the regression pin that a `needs_rebuild`-but-deployed-once page STAYS | strong |
| Live 404s cleared | full re-audit of all 41 pages on both sites: **56 → 11**, all 11 are 071's class | strong — probed over HTTPS |
| First re-rendered page fully clean | `gaswholesalers.com/index.html`: **19 internal links, 19 × 200**, clean 2-link legal footer replacing the 21-link `053` fallback | strong |
| **Fix has FIRED in production** | **NOT YET.** 0 emissions of the new warn since the roll | ⚠️ **quiet window, NOT proof** — see §4 |

## 3. What shipped

- **Code** (`platform/`), council **APPROVED** round 1, corr `623d7bce`:
  `a9083d51b` (core) · `759cb2b77` (GetNavigationStructure, council objection) · `6e911793c`
  (buildServicesHTML, the fourth writer) · `3030a6053`/`9bc96db39` (gofmt).
  `datahelpers.NeverDeployedPagePredicate` shared by renderer and audit; `deployedOnly bool` →
  `NavVisibility` named type; `maxItems` applied after filtering.
- **Live data** (no roll needed): 3 nav items → `status='inactive'`; `in_header`/`in_footer`
  cleared on 4 never-built pages. Backups: `bak_049_nav_items_20260726`,
  `bak_049_page_navflags_20260726`.
- **Docs:** `bugs_closed/049` (full record + corrections), `bugs_open/083` (new), handover into
  `bugs_open/071`, NOTES, RUNBOOK **R19/R20/R21**, `README_where_we_are`, `016b §9`,
  two `WRONG_CALLS` rows.

## 4. THE THREE NAMED NEXT CHECKS

**(a) Prove the fix FIRES, not just that it shipped.** Nothing has emitted the new warn since the
roll, because the two sites I cleaned no longer have qualifying nav items. Three sites still do —
`dartsonline.com` (4), `leopardessconsulting.co.uk` (4), `oufe.com` (1), per RUNBOOK **R20**. A
chrome re-render on any of them should **drop those items and log**
`dropped nav items whose target page has never been deployed`. I did **not** fire it: those sites
are outside 049's remit and it is an outward-facing content change. Whoever owns them should do
it, and it is the cleanest available proof of the failing branch.
```
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system logs $POD --since=30m | grep "dropped nav items whose target page has never been deployed"
```

**(b) One gaswholesalers page may still be stale.** `wholesale-pricing-explained.html` missed the
site-wide refresh *while reporting its work item `complete`*, and still served the entire pre-fix
footer. Re-fired via `049b_deploy_single_page.sh` (assemble-only) at 21:09, corr
`1248320f-8061-473a-ba65-474ca687960d`. **Check it:**
`curl -s https://gaswholesalers.com/wholesale-pricing-explained.html | grep -c fuel-pricing-framework` → want **0**.
If still 2, re-fire once (check `scripts/dispatch-queue-depth.sh` FIRST) — the build dispatcher
was observed stalled earlier (0 completions fleet-wide in 20 min with 2 items queued).

**(c) `bugs_open/083` is the bigger finding and needs an OWNER decision.** Broken-link findings
have been detected correctly for months and fixed **zero times ever** (22 detected, 0 complete;
98 items parked in `status='detected'`). The only promoter `detected`→`triaged` runs inside
`improvement-loop`, fired only by the `improvement-sweep` scheduled task, **disabled since
2026-05-02**. Re-enabling it sets fixing agents running fleet-wide and spends credits, and
something turned it off deliberately — so it is not a platform call. **`bugs_open/077`/`078`
must be cleared first**, or the pile just moves from `detected` to `failed`.

## 5. Landmines this session paid for (all now in RUNBOOK/WRONG_CALLS)

- ⚠️ **`NavFetchableOnly` is a VACUOUS pod-grep marker** — a typed int constant, resolved at
  compile time, never in the binary; the grep returns 0 whether or not the fix shipped. I
  published it as *the* verification step in the closed bug file before executing it. **Grep
  function symbols and string literals, and always pair with a NEGATIVE control** (old line gone)
  — that half is load-bearing. "Grep what your change created" is necessary, **not sufficient**.
- ⚠️ **A census scoped to one writer measures that writer, not the surface.** After the nav fix
  landed and chrome re-rendered, the footer *still* carried the dead link: `buildServicesHTML`
  queries `pages` directly, bypassing `GetNavItems` entirely. The R20 census structurally could
  not see it. Only re-reading the **rendered artefact** caught it.
- ⚠️ **"The queue drained" is not per-page proof.** One page reported `complete` while serving the
  old footer. Check the artefact per page.
- ⚠️ **A fixed-width context grep gives false negatives.** `grep -oE '.{120}X.{60}'` reported a
  page clean when `grep -c` said 2. Count first, contextualise second.
- ⚠️ **An absent `orchestration_states` row means QUEUED, not dropped** — run
  `scripts/dispatch-queue-depth.sh` **before** re-firing. I re-fired twice with the rule in my own
  memory, because a *real* competing mechanism (chassis roll + postgres probe restart loop,
  `bugs_open/082`) predicted the identical observation. **Two known mechanisms, one observation →
  run the discriminator, never pick by plausibility.**
- ⚠️ **Dispatch within ~300s of a chassis pod start is silently dropped.** Cost one fire; the pod
  was 221s old. Check `.status.startTime` before dispatching after any roll.
- The council's `bug_historian` seat caught a claim I had **inherited and not personally
  verified** (that `GetNavigationStructure` only fed prompt context — it feeds real page headers).
  Where a caller legitimately picks the permissive enum value, **the compiler protects nothing** —
  pin it with a test.

## 6. Not mine, observed in passing

- **Build dispatcher stalled** ~17:30 BST: 0 `page_rerender` completions fleet-wide in 20 min with
  only 2 items queued. Belongs to `bugs_open/029`/`030`. Not touched.
- **`postgres-clients-0` liveness-probe restart loop** (8 restarts) — already filed by another
  session as `bugs_open/082`.
- **`platform/kafka` did not compile** in the shared tree mid-session (`undefined: SharedDialer`)
  — another session's uncommitted WIP. I verified my packages against `git archive HEAD` instead
  of touching their files.

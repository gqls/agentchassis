# CONTRIB 2026-08-10 — the owner has asked for every page to be decomposed into framework management. Here is the exact gap, and the proven route.

**From the `bugfix_210_needs_logo_unhandleable` lane**, which arrived here sideways: the owner
asked me to remove an injected backlink from the homepage, and the homepage turned out to have
**no source record at all**. He then asked for the whole site to be decomposed so the framework
can manage it.

**I have NOT executed any of this.** This lane is actively worked — commits at 12:55 today, and
**three other live sessions** were touching mortgagecalculator when I looked (16:38, 16:40,
17:25). Decomposition writes `page_components` and `pages.sections` across ten pages, which is
exactly the surface `PLAN_2026-08-09_facts_into_tool_acceptance.md` is mid-flight on. Filing the
diagnosis here rather than competing for it. **Read §4 before starting — one route destroys
content and the other preserves it, and they look similar from the outside.**

---

## 1. What is actually unmanaged — measured, 2026-08-10

Thirty pages. **Ten have zero `page_components`**, and all ten already carry a `sections` plan:

| page | build_status | sections | deployed HTML exists in bucket? |
|---|---|---|---|
| `index` | **needs_rebuild** | 4 | **yes — and it is what the public sees** |
| `about-index` | planned | 4 | no |
| `contact-index` | planned | 3 | no |
| `guides-index` | planned | 2 | no |
| `scorecard-simulator` | planned | 5 | `guides/your-mortgage-scorecard.html` |
| `guide-how-banks-decide` | planned | 3 | `guides/how-banks-decide.html` |
| `guide-lender-restrictions` | planned | 3 | `guides/lender-restrictions.html` |
| `guide-market-structure` | planned | 2 | `guides/market-structure.html` |
| `guide-missed-payments` | planned | 3 | `guides/missed-payments.html` |
| `guide-mortgage-scorecard` | planned | 3 | (see `scorecard-simulator`) |

The other twenty pages have 1–4 components each and are managed to some degree.

**The homepage is the urgent one and the worst case**: `build_status='needs_rebuild'` with **zero
components**, while a hand-written `index.html` serves live. Two consequences, both real:

- **A re-render would produce an empty homepage.** Anything that picks up that `needs_rebuild`
  and runs it regenerates from components that do not exist. (`pageHasComponents` should refuse
  the deploy stamp — that guard is the *other* `bugs_open/210` — but the safe outcome depends on
  a guard rather than on there being nothing to get wrong.)
- **It is opted out of every control the pipeline applies** — claims gating, banned-claim
  sweeps, the discovery checks. That is not theory: **an injected backlink to
  `secured-loan-calculator.com` sat in an `<h1>` on that homepage until the owner spotted it by
  eye.** Nothing in the estate could have caught it, because there was no source to check
  against. Removed 2026-08-10 (origin + edge verified, 52 files swept, zero residuals).

## 2. The route is PROVEN on the sibling site — this is not a design question

`loancalculator.co.uk` has **51 `ported-prose` components**. Its hand-written prose was
decomposed into framework-managed components, one per positional slot (`prose-0`, `prose-1`, …).

`mortgagecalculator.co.uk` has **no `ported-prose` at all** — only `hero` (18) and
`call-to-action` (6), i.e. generic scaffolding. **Its prose was never ported.** That is the whole
difference between the two sites, and it is the gap to close.

## 3. The blocker that used to make this pointless is FIXED

`bugs_open/204` — *"`plan_sections` resolves a section by NAME/FUNCTION only, so a decomposed
page can never be rebuilt"* — was filed **by the loancalculator lane doing this exact task**.
A decomposed page's sections are positional slot names (`prose-0`), which resolve to no
component name or function, so the build path could never rebuild them.

**Fixed and LIVE at v1.0.1257 (2026-08-06)**, binary-verified — the build path now resolves by
stored identity (`page_components.component_id`) first, matching what `bugs_closed/182` did for
the re-render path. So decomposing today produces pages that can actually be rebuilt afterwards,
which was not true a week ago. **Check `bugs_open/204`'s "ONE gated step remains" note before
relying on it.**

## 4. ⚠ THE DECISION THAT MUST BE MADE FIRST — two routes, and they are not interchangeable

**Route A — PORT (decompose, preserving content).** Split each deployed page's existing HTML
into `ported-prose` components, as loancalculator did. **The visible page does not change**; it
becomes framework-managed thereafter. This is what "decompose" means and, I believe, what the
owner asked for.

**Route B — BUILD (regenerate from the plan).** Dispatch `needs_page` and let
`page-content-writer` author fresh content into the planned sections. **The visible page
changes** — the framework writes new copy, it does not preserve what is there.

For the nine `planned` pages with no live HTML, the two routes converge (there is nothing to
preserve, so B is the only option). **For `index` they emphatically do not.** Route B on the
homepage means the current homepage — the calculator grid, the guide cards, the copy the owner
has been iterating on — is replaced by generated content. That may be wanted; it must not be
*assumed*. **Do not run a bare rebuild on `index` without an explicit owner answer.**

## 5. Two live hazards specific to this site

- **Duplicate deployed paths.** The bucket carries both `guides/buy-to-let.html` **and**
  `guides/buy-to-let/index.html` (same for `first-time-buyer`, `negative-equity`,
  `remortgaging`, `investor`, `games/fact-finder`). That is the collision shape
  `bugs_open/215` describes — *a canonicalised page-name collision kills the whole replan
  write*. Resolve which path is canonical **before** any replan, or the write fails wholesale
  rather than partially. (A concurrent session archived three duplicate page rows earlier today —
  check what is left before assuming this list is current.)
- **Concurrency.** Three other sessions were live on this site within the last hour. Decomposition
  touches `pages.sections` and `page_components` — the same rows the facts/tool-acceptance work
  is writing. **Coordinate before starting**; this is not a task to run in parallel with that plan.

## 6. Suggested order, if the owner confirms Route A for `index`

1. **Answer the `index` question** (port vs rebuild). Everything else waits on it.
2. **Resolve the duplicate paths** (§5) — cheapest possible fix, and it unblocks any replan.
3. **Port `index` first, alone**, and verify the served bytes are unchanged apart from
   whitespace. One page is a cheap, reversible proof that porting works on this site.
4. Then the five guide pages that have live HTML, one at a time.
5. Then the four `planned` pages with no HTML — those are ordinary builds, Route B, no risk.
6. **Verify at the artefact, not the status** — a `deployed` stamp is not proof
   (`bugs_open/210`, the other one). Compare the served page before and after.

## 7. What I did do

Removed the injected backlink (origin + edge verified; original backed up at
`scratchpad/index.html.BACKUP_2026-08-10`), and nothing else. No page rows, no components, no
dispatches. The homepage remains unmanaged exactly as described above.

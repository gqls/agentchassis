# HANDOFF — 2026-08-25. **START HERE.** `bugs_open/389`: the primary CTA is picked by `nav_order` alone, and it is minting wrong buttons today

> **Read this from disk, then `bugs_open/389_HANDOFF_2026-08-25_cta_destination_is_ranked_by_nav_order_alone_so_an_off_topic_tool_wins_every_primary_button.md`.**
> Nothing is fixed. The root cause is confirmed at the code, the data and the served bytes.
> **Four decisions are with the owner** (§3) and no code should be written before they land.

> **Deploy facts have a shelf life of hours — re-probe, do not quote.** Chassis at handoff:
> **`4c996e1b5cb9b2513d88ec9fe2bae220c38fb6c2`**, pods up `2026-08-25 09:27Z`, capability
> re-probed on the running binary rather than inferred (`rendered_html_transform` 8,
> `code_span_to_code_tag` 5, control 0).

---

## 0. STATE

| thread | state |
|---|---|
| **`bugs_open/389`** | **FILED TODAY, OPEN, nothing fixed.** Root cause confirmed three ways. Live and re-minting: the resolver stamped a new password-entropy CTA **today**. Awaiting four owner decisions |
| `bugs_closed/277` | **CLOSED** and re-verified post-roll (§4). No work left |
| `bugs_open/357` | **NOT OURS — actively owned by another session** (commits 08-24). Do not compete |
| `bugs_open/312` | open, adjacent: the resolver/writer seam. 389's fix touches the same action |

## 1. The finding in one paragraph

`chooseCTATargets` (`platform/orchestration/actions/resolve_internal_links_action.go:651`) picks
the site's primary call-to-action destination by taking **every** `tool`/`game` page on the site,
dropping only excluded areas and self-links, sorting by **`nav_order` then alphabetically**, and
returning `ordered[0]`. There is no topic, tag, or relevance input of any kind. On three sites a
password-strength toy carries the fossil value `nav_order = 1` (set at creation, `2026-03-13`)
while every relevant tool sits at 6–204 — so it wins the primary button on every page, every run.

**The design fault under it:** `pages.nav_order` is doing two unrelated jobs — ordering the nav
menu and ranking CTA candidates — and `in_header` is not read by the chooser at all. A previous
session hid this exact page from the nav with the comment *"a password tool doesn't belong in the
primary nav"* and it changed nothing.

## 2. What is proven, and how (so nobody re-derives it)

| claim | how it was established |
|---|---|
| the chooser has no relevance input | read `chooseCTATargets` + `loadInteractivePages` in full |
| the ranking predicts the observed winner | simulated the exact `ORDER BY` against live `pages`; matches on all 3 sites |
| `in_header` is not consulted | `PageMayBeLinkedPredicateFor` (`datahelpers/links.go:333`) is deployment-state only |
| **it is live, not residue** | `__cta_minted` stamp: **17 fields minted 08-23 → 08-25 (today)** |
| the damage is real | served bytes on all 3 domains, incl. `ai-agent-orchestration.com/index.html` |
| it is NOT "CTAs shouldn't link tools" | 105 such fields on `webdesign.co.uk`, 78 on `dartsonline.com` — normal and wanted |
| only 3 sites are clearly wrong | reviewed the rank-1 winner for all **26** sites with tool pages |

⚠ **One claim I nearly filed and it is FALSE**: that 13 sites show a deliberate hide-vs-rank
contradiction. **62.7% of tool/game pages are `in_header=false`** (143 of 228) — it is the normal
state, not a human judgement. Only leopardess is documented as deliberate, by its SQL comment.

## 3. THE FOUR DECISIONS — with the owner, do not pre-empt

1. **Content:** should `password-entropy` be on those three sites at all? (Removing the page
   removes the candidate. It is a decision about what the sites offer.)
2. **Data:** correct the three `nav_order` values? ⚠ On `ai-agent-orchestration.com` the page is
   `in_header=true`, so `nav_order` is also its visible menu position — **the fix inherits the
   coupling that is the bug**.
3. **Platform:** give the chooser a relevance input or an explicit opt-out? This is the only
   option that closes the class. **Architecture-scope** (shared seam, every site) → 2026-08-02 §2:
   new authority ships **opt-in, unsafe default OFF**.
4. **Repair:** the 80 stored values. Reuse `bugs_closed/268`'s fleet CTA-resolution re-run —
   **after** 1–3, or it re-mints the same answer.

**Recommendation if asked:** candidate 1 in the bug file — an explicit "never a CTA target" flag
read by `loadInteractivePages`. It makes the intent *sayable*, which today it is not; that is
precisely why hiding the page from the nav was the only move available and accomplished nothing.

## 4. `bugs_closed/277` — re-verified on the new build, still closed, no work left

Post-roll at `4c996e1b5` [MEASURED 2026-08-25]: clause 1 holds at the served bytes (cubic-bezier
`<code>ease-in-out</code>` = 1, gas-unit-converter `<tr` = 6, control: prose backticks = 0).
The residual is stable at **21 parked** (12 drift-blocked + 9 in `357`), and **genuinely unrouted
= 0** across all **92** rows and both producers — the `partial` route grew 29 → 57 since 08-24, so
the router is actively working, not merely quiet. §10 of the bug file has the full account,
including the amended verify query. **Do not run the pre-08-24 verify query: it cannot fail.**

## 5. Session-start checklist
`git log --oneline -10` · re-read this from disk · `scripts/who-owns.py 389` and `357` ·
chassis stamp + capability probe (never infer from ancestry) · then §3. **If the decisions have
not landed, do not write code** — measure instead, and the useful next measurement is §6.

## 6. The next useful measurement, if you are waiting on decisions

Whether the *borderline* case is real: `webdesign.co.uk` picks `tool-ab-test-calculator` out of
**66** tools. On-topic enough for a web-design site, or the same fossil-nav_order shape? Its
`nav_order` is 100 — the COALESCE default, not a deliberate 1 — so it is a weaker instance and
worth one human glance before anyone widens the bug's scope to it.

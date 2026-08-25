# HANDOFF — 2026-08-25. **START HERE.** The primary CTA is picked by `nav_order` alone, and it is minting wrong buttons today

> ⚠ **THIS LANE'S BUG IS NOW `bugs_open/391`, NOT 389.** Full path:
> `bugs_open/391_HANDOFF_2026-08-25_cta_destination_is_ranked_by_nav_order_alone_so_an_off_topic_tool_wins_every_primary_button.md`
> It was 389 for ~40 minutes; that number belongs to the `bugfix_308` lane's
> `389_…_repair_completion_is_unverified_…` (filed 2m25s earlier, cited from `bugs_closed/308`), and
> **390 was taken by a third session while I was renaming.** Commits before ~11:40 saying "389"
> about CTA *selection* mean this file. `git log` the FILE PATH, never the number.
>
> The two are complementary — and **389 constrains decision 4 here**: a `cta_links_stale` rerender
> completes green whether or not any CTA moved, so no repair in this lane may be judged by its
> work-item status.
>
> ⚠⚠ **REVIEWED 2026-08-25 (two adversarial passes). Diagnosis CONFIRMED; my RECOMMENDATION was
> WRONG.** See `bugs_open/391` §THE FEEDBACK LOOP. Summary: the label match runs *ahead* of the
> positional pick, and the framework writes button copy naming whatever it picked — so a wrong pick
> becomes **label-locked** and a `nav_order` fix cannot reach it. **20 of 80 fields are locked,
> including all three buttons the owner saw.** The commissioned content pass is therefore NOT
> redundant; it is exactly what those 20 need, re-scoped to ~20 fields instead of 16 sites.

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

## 3. THE FIVE DECISIONS — ⚠ ALL ANSWERED BY THE OWNER 2026-08-25. Decision 2 is DONE.

| # | answer | state |
|---|---|---|
| 1 | *"the password tool can disappear everywhere"* | **planned, deliberately NOT done first** — 91 refs + footer + 3 listings; deleting ahead of the copy rewrite strands them and leaves ~20 buttons naming a tool they no longer point at |
| 2 | *"yes change the menu-order numbers"* | ✅ **DONE + VERIFIED** — `SQL_2026-08-25_demote_password_entropy_nav_order.sql`, 1 → **900** on three rows, guarded. New rank-1: ROI estimator / AI data-risk checker / ROI estimator |
| 3 | *"yes go ahead"* | **approved, not started.** Candidate 1 + candidate 4; read at the RANKING not the loaders; must bind `LoadCTALabelUniverse`; RFC_022 enumeration owed before council |
| 4 | *"whatever you suggest"* | sequenced last; verify at served bytes, never by work-item status |
| 5 | *"rescope it as you suggest"* | ~20 label-locked fields by query |

⚠ **Why 900 and not 200:** at 200 it ties with the sites' other tools and the tiebreak is
alphabetical on `name` — `password-entropy` precedes every `tool-*`, so **it would still have won**
on two of three sites. A demotion that joins the pack is not a demotion.

**THE NEXT ACTION IS STEP 2 OF THE RETIREMENT SEQUENCE** (`bugs_open/391` §RETIREMENT): rewrite the
~20 label-locked labels via the re-scoped content pass. Nothing else may run before it — not the
repair, and above all not the page deletion.

**One question left with the owner, raised not assumed:** the library component
`tool-password-entropy` is still `is_active = true`, so it can be handed to a *new* site.
"Everywhere" may or may not cover that switch; untouched.

### The original framing, kept for the record
> ⚠ Was four. The fifth (the standing commission) is a decision about the owner's own 08-15
> instruction and cannot be folded into the repair. Decision 3's "only option that stops the class"
> was **overstated** — an opt-out is reactive; pair it with a detector to earn that claim.

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

> ⚠ **REVISED the same day — read `bugs_open/389`'s CORRECTION section before acting.** This
> symptom was already measured on **2026-08-15** by the `cta_target_content_pass` lane (16 sites
> with ≥6 rows on one modal target; finetuning 39, ai-agent-orchestration 36; password-entropy the
> modal target on three sites, called *"topically absurd"* there). **The owner accepted it as a
> floor and commissioned a content pass; nothing was run.** So decision 4 is not "reuse 268" alone
> — there is a **standing commission** to honour or withdraw. What this lane adds is the CAUSE
> (`nav_order = 1`, a fossil from 2026-03-13) and proof it is still minting. **The ordering now
> matters more than the choice: correct the ranking input FIRST, re-measure, and only then size
> the content pass** — otherwise an LLM rewrite across 16 sites is spent working around a fossil
> integer that stays in place to mis-rank the next tool added.
>
> ⚠ Also: `render_site_components_action.go`'s **site header fallback** is a third consumer of the
> same loaders, and its output is never persisted (`site_components` holds 0 `cta_url` keys across
> 24 header rows) — so a `content_data` diff **reads clean while all 24 headers move**. Verify any
> fix at the rendered header.

**Recommendation if asked — REVISED 2026-08-25, twice over:**

1. **Ranking fix first** (decision 2 or 3) — it clears the ~60 label-less fields in one step per
   site and stops new wrong picks being minted. The locked set **grows** until this lands.
2. **Then the commissioned content pass, re-scoped** to the ~20 label-locked fields (query in
   `bugs_open/391` §4) — not the 16-site sweep its own plan assumed.
3. **The platform option is candidate 1 PAIRED WITH candidate 4** (opt-out flag + a detector for
   the anomalous-`nav_order` shape). Candidate 1 alone is reactive and does not close the class.

⚠ **Two specification constraints on candidate 1, both from review — get these wrong and the fix is
worse than nothing:**
- **Change the RANKING, not the loaders.** `render_site_components_action.go:182-190` (the site
  **header** CTA fallback) calls the loaders directly and takes `ordered[0]`, and its output is
  **never persisted** — so a loader change re-picks every site's header button with no
  `content_data` diff to show it.
- **A flag on the ranking alone does not bind `LoadCTALabelUniverse`.** The label match runs first,
  so an "ineligible" page is still selected whenever the copy names it — a hole exactly the shape
  of this bug.
- **Engage RFC_022 before booking a council round**, including its required consumer enumeration.

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

## 6. The next useful measurements, if you are waiting on decisions

1. **Re-run the 26-site blast-radius review with the code's real predicate.** The original omitted
   `NOT (deployed_at IS NULL AND build_status='planned')`, so it can name a rank-1 winner the code
   would skip. Corrected query is in the RUNBOOK and in `bugs_open/391` §Verify.
2. **Whether the borderline case is real:** `webdesign.co.uk` picks `tool-ab-test-calculator` out of
   **66** tools. Its `nav_order` is 100 — the COALESCE default, not a deliberate 1 — so it is a
   weaker instance. One human glance before widening scope.
3. **Watch the locked set grow.** Re-run the label-lock query (RUNBOOK) in a few days: if the
   locked count climbs above 20, that is the feedback loop measured over time, and it is the
   strongest argument for doing the ranking fix first.

## 7. What the two review passes confirmed, so nobody re-derives it

Every load-bearing claim reproduced independently: the sort keys and the absence of any relevance
input; `nav_order=1` on exactly three sites against 6–204; `in_header` absent from the CTA path
(and the L5 script that set it also renumbered that site's nav 2–10 while leaving the tool at 1);
the 17/24/39 provenance split; the served bytes on all three domains; and the 62.7% base rate that
refuted the 13-site reading. **What they overturned was the fix and the sizing, not the diagnosis.**

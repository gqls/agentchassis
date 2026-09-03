# 436 — CTA destinations are ranked by `nav_order` alone, so every NEW site inherits an off-topic primary button

**Filed 2026-09-02**, spun out of **`bugs_closed/391`** on its closure (owner instruction). 391 fixed
the *damage* on three sites; this is the *cause*, which is untouched and fires on every site the
framework builds. **Status: OPEN. Severity: medium** (no live damage today — 391 cleared it — but the
next build re-creates it).

## The mechanism, read from the live code

`chooseCTATargets` (`platform/orchestration/actions/resolve_internal_links_action.go:651`) picks a
site's primary CTA by sorting every `tool`/`game` page on `COALESCE(nav_order,100)`, then `name`, and
taking `[0]`. **There is no topic, tag, vertical or semantic input at all.** Whatever sorts first wins
the primary button on every page of the site.

**And the wrong pick locks itself in.** `stampCTADestinationGuidance` (`:362`) feeds the chosen
destination's title into the writer's spec for the label field, so the framework writes button copy
*naming* whatever it picked. The next resolve label-matches that copy back to the same page —
`LoadCTALabelUniverse` runs **ahead** of the positional pick — so the row becomes unreachable by any
`nav_order` change. **Measured on 391's population: 20 of 80 fields had reached that state, including
all three the owner reported.**

## Why 391's fix does not close this

391 corrected the *data* (`nav_order` 1 → 900 on three sites) and repaired the *copy* (20 label-locked
fields rewritten, 21 contact-intent fields routed to `/contact.html`). None of that touches the
ranking. A new site with a fossil or unlucky `nav_order` — or simply an alphabetically-early tool —
gets the same off-topic primary button, and the same lock-in on the next content pass.

`[MEASURED 2026-08-25]` the fossil that caused 391 was set **at page creation, 2026-03-13**, on three
sites at once. Nothing prevents the next one.

## Owner decision 3 (2026-08-25) — approved in principle, not built

**Candidate 1 (an explicit `eligible_as_cta_target` opt-out) paired with candidate 4 (a detector for
the anomalous-`nav_order` shape).** Three constraints, all from review and all still binding:

1. **Read at the RANKING, not the loaders.** `render_site_components_action.go:182-190` — the site
   **header** CTA fallback — calls the loaders directly, takes `ordered[0]`, and its output is
   **never persisted** (`site_components` holds 0 `cta_url` keys). A loader-level change moves every
   site's header button with **no `content_data` diff to show it**.
2. **It must also bind `LoadCTALabelUniverse`**, or the opt-out has a hole exactly the shape of this
   bug — the label match runs first, so an opted-out page still wins through its own copy.
3. **Engage RFC_022 and ENUMERATE the consumers before booking a council round.** Asserting the
   opt-in shape without the query is itself the objection (owner ruling 2026-07-29 §3: a shared
   mechanism's other consumers must be *told*, not merely measured).

## Why this is architecture-scope

It adds a field to a shared seam that every site's CTA resolution reads, and it changes what the
ranking **guarantees** — the 2026-07-29 §1 trigger. Expect `needs_rfc`, and note RFC_022's narrowing:
an opt-in field whose unsafe default is OFF and which no live consumer names is **not** architecture-
scope on shape alone, but this one changes the guarantee, so it is.

## How to verify a fix

**Induce, do not wait.** Seed a site with two tools where the alphabetically/`nav_order`-first one is
off-topic for a given page, run the resolve, and assert the primary CTA is **not** it — then flip the
opt-out and assert it **is** eligible again. Both directions in one run. And assert the **header**
fallback separately (constraint 1): it is a different call site whose output is never persisted, so a
`content_data` check cannot see it.

## IN PROGRESS 2026-09-02 (evening) — the lever + alarm are BUILT and council-APPROVED (round 2, corr 9faa2a23); Go inert until the roll

Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_436_cta_eligibility/` (PLAN, NOTES, RUNBOOK,
README). Register entry **LNK-041** carries the full design. Re-verified at HEAD before building:
mechanism unchanged, and the open queue independently corroborates it `[MEASURED 2026-09-02]` (two
deferred verdicts describing a 63-tool site's "two arbitrary tools … as default fallback CTAs").

What shipped, honouring all three constraints:
1. **`pages.eligible_as_cta_target boolean NOT NULL DEFAULT true`** (migration `714`; default true =
   today's behaviour byte for byte). Read at the **RANKING** — supply SQL + ordering moved to
   `datahelpers/cta_positional.go` (shared with the detector; constraint 1), `RankCTAPositionalCandidates`
   drops opted-out pages for all three `chooseCTATargets` callers including the header fallback,
   which has its own named test (`TestRankHeaderFormRefusesAnOptedOutPage`).
2. **`LoadCTALabelUniverse` bound** (constraint 2) — by **refusal, not removal**:
   `BestLabelMatchForPage` declines an opted-out best match while the page stays in the pool
   (dropping it lets a one-token runner-up win — the measured self-link failure). `JudgeCTALabel`
   gains silence reason `names_ineligible_page` so 399's judge keeps a first-class signal.
3. **Consumers enumerated** (constraint 3) — in the PLAN, the LNK-041 entry, and the council
   submission: 3 ranking callers, 4 universe callers, 3 `BestLabelMatchForPage` callers. No new
   action-input key, so the RFC_022 optional-key budget is untouched.
4. **Candidate 4** = discovery check `cta_rank_anomaly`: fires when the site-level rank-1 CTA
   target holds a unique-minimum `nav_order` below the default with a ≥50 lead, among ≥3 eligible
   interactive candidates; review-only; positively retracts when healthy. Enabled by
   `715_…_HOLD.sql`, ⛔ **hand-applied only after the roll** (the discovery runner FAILS the step
   on an unregistered check name).

**Stated limits (not regressions):** the lever does not unlock already label-locked fields on the
recompute path (KEEP #2 holds any valid stored page — 391's own finding); keeps/nav/listings are
untouched; the all-default alphabetical-winner shape (webdesign) is candidate 3, deliberately not
built. No page is opted out yet — owner decisions.

**Remaining to close (updated 2026-09-03, mid-morning):** ~~council verdict~~ (APPROVED r2) →
~~roll~~ (proven at `service_binary_capabilities`, both controls) → ~~apply `714`~~ (2026-09-02) →
~~hand-apply `715`~~ (2026-09-03, snapshot verified) → ~~observe the check's first pass~~ → ~~induced
canary, ranking, both directions~~ → **⛔ the header button at the SERVED bytes — BLOCKED** → owner's
opt-out decision (usage, not a close blocker).

**Verified live 2026-09-03** (evidence + queries: lane NOTES/RUNBOOK):

1. **The check runs, and its silence is readable.** `collected_data.run_checks` records
   `checks_run` / `checks_unregistered` / `checks_failed` per run — structured, no shelf life, names
   the check individually. First post-enablement pass (idea.uk, 09:26:06Z): ran, 46/46, none
   unregistered, none failed.
2. **The zero was rotation coverage, not a healthy fleet.** A hand-mirror of
   `datahelpers/cta_positional.go` over every site `[MEASURED 2026-09-03 10:05Z]` predicts **4** sites
   fossil-shaped: `cv1.co.uk` (`tool-example`, nav 2 vs 200), `boxingonline.com` (3 vs 200),
   `vetcomparison.uk` (4 vs 200), `gamesdesign.co.uk` (20 vs 100). Induced runs on two of them filed
   `needs_human_review` items quoting the census's own pages and numbers. **Disconfirming control:**
   idea.uk's rank-1 is also below the default and is correctly ABSENT (lead 7 — the curated ladder),
   and the check stayed silent there for that stated reason.
3. **Corroborated off the database.** cv1.co.uk serves
   `<a href="/tools/example/index.html" class="header-cta">` today (controls: target 200, invented URL
   404), and **7 of 10** stored CTA destinations on that site point at `tool-example` — including both
   *other* tools' guide pages. The detector's claim holds against the bytes a visitor gets.
4. **The lever binds the ranking, both directions, in the deployed binary against the live column.**
   cv1.co.uk: opt out → the check retracts (`items_resolved: 1`) with reason *"only 2 eligible
   interactive candidate(s)"*; opt back in → detail returns to *"among 3 candidates"*,
   `items_inserted: 1`. **The count 3 → 2 → 3 is the assertion** — only
   `RankCTAPositionalCandidates` filtering on `IneligibleAsCTATarget` produces it, and that is the
   function all three callers share. Control: vetcomparison.uk's item untouched across all four runs.
   Site restored to `eligible=true`; fleet-wide opted-out back to **0** — no data decision taken.

**5. The HEADER caller — VERIFIED at the artefact 2026-09-03 11:25Z** (owner dispatched the render
this session could not). After `rerender-pages` with `refresh_site_components:true` and
`tool-example` opted out, cv1.co.uk's stored chrome carries
`<a href="/tools/job-search-readiness-checker/index.html" class="header-cta">` (`updated_at`
11:25:20.499Z), where the previously-deployed chrome served `/tools/example/index.html`. That is the
third and last `chooseCTATargets` caller, moving with the lever.

> **CORRECTION — this file, the handoff and the round-2 disposition all said the header's pick is
> "never persisted — no DB check can see it". Too strong, and unverified when accepted.** Not
> persisted: the decision *as a field* — `site_components` holds **0** `cta_url`/`header_cta_url`
> keys `[MEASURED 2026-09-03 11:32Z]`, so `cta_positional.go`'s package comment and its
> bind-at-the-ranking argument are accurate and unchanged. Persisted: the *rendered anchor* — **36**
> rows fleet-wide carry a `header-cta` href in `rendered_html`. The outcome was one query away the
> whole time.

**What that correction bought:** the header census the lane thought impossible. Of the four fossil
sites the header points at the fossil on **two** — cv1.co.uk and boxingonline.com. On
gamesdesign.co.uk and vetcomparison.uk a footer-group nav item labelled `contact` wins before the
ranking is consulted, so the fossil reaches only their STORED page CTAs.

**Still not on the wire, and it is NOT a lever question.** 20 cache-busted fetches over 13 minutes
showed no change. `rerender-pages` re-renders chrome synchronously but only QUEUES page reassembly —
7 `page_rerender` items filed, all `triaged`, behind **170** fleet-wide. Demand control:
**21** such items completed fleet-wide in the same window, so the handler is draining and mine are
merely not at the front.

**Open loose end:** cv1.co.uk's `tool-example` is restored to `eligible=true` (fleet opted-out = 0)
but its stored chrome still holds the opted-out render's pick, so chrome and data disagree until one
more render runs. That render also completes the two-way at the same artefact.

**A wrong prediction of mine, recorded so it is not re-derived:** I expected the flip-back to file
nothing, because `bugs_open/326` dedups item keys in any status and the resolve had just left that key
on a row. It filed a fresh item. A *retraction* does not poison the key the way a human *dismissal*
does; the check's header comment describes the latter.

## Relations

- `bugs_closed/391` — the damage this caused, fixed; its lane docs carry the full evidence.
- `bugs_open/248` (`cta_recompute_clobbers_authored_contact_links`) — KEEP #1, the mechanism that
  makes an authored `/contact.html` durable; 391's contact-intent fix rests on it.
- `bugs_open/399` — records label/destination disagreement at write time. ⚠ **Structurally blind to
  this bug**: when the framework picks the destination *and* names it in the copy, the two agree, so
  the judge says nothing. Their own `TestJudgeCTALabelIsBlindToTheLabelLockedDefect` pins it.
- `bugs_open/384` — the stale-listing family; holds 391's retraction residue.
- Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_389_cta_relevance/`.

# HANDOFF — finance-domain build-out: register, design diversity, enforcement gaps — 2026-08-11, continue here

Cold-start for a fresh chat picking up this exact thread (not superseding
`HANDOFF_2026-08-03_continue_here.md`, which is this workstream's earlier seam-backlog
history and still correct on its own topic). **Read `MEMORY_workstreams.md`'s
portfolio-positioning line first for older context, then this file.**

## 0. Owner rulings in force (this thread, all 2026-08-11)

1. **Twin-pair residuals: CLOSED.** P8 — `besthealthinsurancerate.co.uk` splits from
   its plural sibling by regional scope (England vs whole-UK); `bestlandlordinsurancerate.co.uk`
   splits by portfolio size (1–10 properties vs 3+/HMO). No more per-pair owner calls
   needed anywhere in the register.
2. **vigilant_designer_offer_analysis: A-track next, not B4.** Driven by the design-
   diversity ask below. Recorded in that lane's own PLAN decision log.
3. **Bug 252's locale mechanism: option 3** — `lang` lives in the head component, not
   a new `sites.language` column.
4. **Enforcement-gaps priority for THIS session: structural-validity gate → bug 161
   (fact discipline) → fidelity dial, in that order, next.**
5. **THEN — after those three — mortgagecalculator.co.uk's copy tone/voice/style**
   (found this session, §6 below) goes to the owner for live review and approval.
   Explicitly queued *after* the enforcement-gap work, not instead of it.

## 1. What happened this session, in one paragraph

Resolved the two remaining twin-pair decisions in the register (P8), researched the
site-design pipeline and recommended a Gemini-design-variant + critique-loop experiment
rather than a fleet-wide model flip, recorded the resulting A-track-vs-B4 call for
`vigilant_designer_offer_analysis`, and worked the enforcement-gap backlog — only to find
bug 251 already fixed and bug 252 already mid-flight in another concurrent session
tonight, so contributed the one missing decision into the shared bug file rather than
writing competing code. Then found, on request, that a *third* concurrent session had
spent the same evening on mortgagecalculator.co.uk's copy voice — a live, evidenced
"AI slop" diagnosis directly relevant to the design-diversity question this thread opened
with.

## 2. The register — decision-complete

`REGISTER_positioning.md`: 152 domains, 43 propositions, P8 applied, `check_register.py`
passes. Nothing outstanding on classification. Committed `f93339a40`.

## 3. Design diversity — recommendation, and a finding that should adjust it

Owner didn't like lendzy.co.uk's design (generic/"AI-designed"). Researched: the design
step (`webdesign-agent`'s `analyze_design`) is a separate LLM call from content, currently
Claude on every site, no lane has ever attempted visual diversity, and the palette-pin
mechanism is advisory not a lock. Recommendation stands: a Gemini-driven *variant* of the
design-agent routed to a handful of sites (not a fleet-wide flip), paired with the design
critique agent (Phase 2 of `vigilant_designer_offer_analysis`'s Programme A, unbuilt) for
a reject-and-retry loop — see §4.

**Adjust expectations before building it**, per §6: the same evening, a different site's
"AI slop" diagnosis concluded **the model was not the cause** — reverting the writer to
Gemini was tried and explicitly withdrawn once the real cause (a commissioned brief plus
an unwritten voice spec) was found. It may be that a sharper *design brief* moves the
needle more than which model executes it, the same way it did for copy. Worth testing
both — model variant AND brief specificity — rather than assuming the model swap is where
the value is.

## 4. vigilant_designer_offer_analysis — the call made

A-track (not B4) is next, specifically Phase 2, the `design-critique-agent` ("018
critic"). Recorded in `vigilant_designer_offer_analysis/PLAN_2026-08-02_..._analysis.md`'s
decision log, committed `11d56fdee`. **Scope note for whoever builds it**: as specced,
Phase 2 critiques a design against the site's own `design_intent` + a fleet homepage-
skeleton summary — NOT against external "well-designed site" references. If the
reject-and-retry loop from §3 is meant to judge against outside taste, that's a scope
addition (a reference corpus in `load_design_context`), not existing plumbing — decide
before building.

## 5. Enforcement gaps — status, and what's actually next

- **Bug 251** (canonical named `/index.html`) — **DONE.** Fixed, tested, mutation-
  verified, council-submitted (`61abbdbd0`, corr `33fb41cb…`) by the
  `loanandmortgagecalculator_couk` lane, ~17:00 tonight. Inert until the next chassis
  roll — not yet live.
- **Bug 252** (dropped `og:` tags, hardcoded `lang="en"`) — **owned by the same lane,
  queued, not started.** Their plan of record: og: half after 251 rolls, lang half was
  blocked on exactly the mechanism decision in §0.3, now unblocked. Recorded in
  `bugs_open/252_HANDOFF_..._assembly_drops_...og_tags_and_hardcodes_html_lang_en...md`,
  committed `f666408ed`. **Do not write code for this — check
  `python3 scripts/who-owns.py 252` before touching it; two unrelated bugs share the
  number 252, use the slug.**
- **Structural-validity gate, bug 161 (fact discipline), fidelity dial — UNTOUCHED,
  unowned as of this session's last check.** This is the explicit next work:
  1. **Structural-validity gate.** No standing check exists for sitemap completeness,
     structured-data parse-validity, or live-vs-repo byte drift (existing discovery
     checks in `platform/orchestration/actions/discovery_checks/` cover neither).
     `loanandmortgagecalculator_couk/verify_site.py` is the prior art to generalise —
     it already caught bug 251. Shape: a new `check_site_structural_validity.go`
     following the `Register()`/`RegisterVerifier()` pattern.
  2. **Bug 161** (fact discipline / banned-claims) — the generalisable fix is
     architecture/RFC-scope, not a quick patch: 72% of facts fleet-wide are
     prose-sourced and unverifiable by construction. Route to architecture review
     rather than arming site-by-site.
  3. **Fidelity dial** — lowest priority; only `locked` is wired, `high/medium/low`
     remain inert. Confirmed, not urgent.
  - **Before starting any of these**: this tree had at least three sessions active
    simultaneously tonight. Re-run `who-owns.py` and `git log --oneline -10` — do not
    trust this handoff's ownership snapshot past the moment you read it.

## 6. NEW — mortgagecalculator.co.uk's copy tone/voice/style (queued for AFTER §5)

Found on request in `mortgagecalculator_couk_adoption/HANDOFF_2026-08-11_continue_here.md`
(and its `REFERENCE_2026-08-11_learned_by_correction_house_voice.html`, also published as
an artifact for the owner). Same evening, unrelated lane, same general shape of problem as
§3: the owner rejected the homepage as "AI slop," and the actual cause — traced live, not
assumed — was a **commissioned brief plus a recorded voice spec** the writer was
faithfully executing, not a model failure. The owner asked to revert the writer to Gemini,
then withdrew that once the brief/spec was shown to be the cause: **"the model is NOT the
lever"** is now a standing ruling on that site.

The voice spec went through four owner corrections in one evening (documented with the
owner's actual words, not paraphrase, in the REFERENCE file) — worth reading in full
before the review, since the corrections themselves are the transferable part: staccato
sentence-length rules borrowed from a safety-critical readability standard that doesn't
fit ordinary prose; an outright ban on a device that's fine at low frequency and bad only
as a barrage; a "smell" reported as a defect when the owner had already accepted it;
and a contraction rule that broke on formal words. The closing rule stated there —
*"do not write a sentence no one would say out loud"* — is doing more work than any of
the specific bans.

**What's actually live vs. not, before you present this for approval:**
- Homepage copy: live and verified (new H1, new section headings).
- 31 page titles were rewritten in `pages.title`, but **only the homepage's has reached
  served HTML** — the other 30 pages still serve their OLD `<title>` tag until each is
  individually re-assembled. That lane's own next action #1 is "finish the titles" and
  is described as purely mechanical but **not confirmed done as of this handoff** — check
  their NOTES tail before telling the owner the whole site reflects the new voice.
- The site is otherwise unlocked and framework-managed; nothing here needs a chassis roll.

**Action for next session**: once §5's three items are through, bring the live homepage
(and the finished title pass, if it's landed by then) to the owner for review against
`REFERENCE_2026-08-11_learned_by_correction_house_voice.html`, and get explicit approval
or further correction — the owner has said plainly they have not yet approved this.

## 7. Next actions, in order

1. Structural-validity gate (§5.1) — new work, unowned, start here.
2. Bug 161 fact-discipline — scope for architecture review (§5.2), don't patch piecemeal.
3. Fidelity dial (§5.3) — lowest priority, quick to confirm still inert and move on.
4. Re-check `mortgagecalculator_couk_adoption`'s title-finishing progress, then bring the
   live site + voice spec to the owner for review/approval (§6).
5. Still open, not scheduled: the Gemini-for-design experiment (§3) — its own decision on
   scope (which sites, whether to add an external reference corpus to the critic) once
   1–4 clear.

## 8. Files of record

This dir: `REGISTER_positioning.md`, `PLAN_2026-07-31_differentiation_axes.md` (P8),
`SUMMARY_2026-08-11_where_things_stand.md`, `README_where_we_are.md` (append-only, read
the tail).
Other lanes: `vigilant_designer_offer_analysis/PLAN_2026-08-02_..._analysis.md` (decision
log) + `HANDOFF_2026-08-10_continue_here.md`; `mortgagecalculator_couk_adoption/
HANDOFF_2026-08-11_continue_here.md` + `REFERENCE_2026-08-11_learned_by_correction_house_voice.html`;
`loanandmortgagecalculator_couk/` (owns bugs 251/252 execution).
Bugs: `bugs_open/251_...` (fixed, awaiting roll), `bugs_open/252_...og_tags_and_...lang_en...`
(locale decision recorded, og:/lang fixes queued by the owning lane — NOT the other bug
sharing the number 252).
Commits this session: `f8df69eab`, `f93339a40`, `11d56fdee`, `f666408ed`.

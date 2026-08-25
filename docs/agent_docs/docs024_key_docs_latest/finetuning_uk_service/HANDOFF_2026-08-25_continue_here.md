# HANDOFF 2026-08-25 — both offer pages LIVE (facts approved, REGISTER rejected + escalated); the lane is parked at the OWNER-DECISION frontier. Start here.

**COLD-START for the merged finetuning.uk lane.** Supersedes `HANDOFF_2026-08-24b_continue_here.md`
(kept for the build/377/nav detail). Technical log: `NOTES_finetuning_uk_service.md` 08-24 a–d +
08-25 a–b. Owner prose: `README_where_we_are.md` (his document — append only). The register
experiment series + owner escalation live in `copy_quality_two_stage/`:
`CONTRIB_2026-08-24_from_the_finetuning_lane_the_exemplar_seed_outcome_and_the_brief_that_taught_the_tell.md`
(3 addenda) and `CONTRIB_2026-08-25_OWNER_ESCALATION_finetuning_pages_fail_the_would_a_person_say_this_test_after_a_maximal_seed.md`.

> ## ▶ DELTA 2026-08-25 (later, same day) — the escalation is ANSWERED, the holds STAND, and one owner option has gone
> **Read this before the table below; it supersedes three of its rows.** Full evidence: NOTES 08-25c.
> 1. **`copy_quality_two_stage` answered.** Our escalation is **item 0 of their next work**, and a
>    **SECOND owner escalation arrived the same day** (his homegarden.uk review — canonical
>    `loanzy_uk_example_site/OWNER_REVIEW_2026-08-25_homegarden_and_what_it_says_about_every_site.md`)
>    carrying instructions: up their game a lot · refresh context before proposing fixes · audit
>    EVERY prompt in DB and code for AI-style writing. Two are done (their `REFRESH_2026-08-25…`
>    and `PLAN_2026-08-25_prompt_audit.md`, phase 1 next). **Nothing has shipped that changes the
>    register.** So: **every hold in the table below stands** — no rebuilds, no cross-link runs, no
>    site-wide rewrite. They also reached our instrument finding independently.
> 2. **Owner decision 1 has LOST its "apply" option.** The parked copy-editor proposal `8003c51a`
>    **fails their gate on STRUCTURE**, re-graded first-hand here, not taken on report: edit 1
>    `h3 2→1, li 3→0, ol 1→0` (an entire ordered list deleted), edit 2 `h3 2→0, p 4→2`. The
>    `/contact.html` noise their CONTRIB told us to discount is now **fixed and gone**; edit 1 is
>    credited with ADDING that link. This is the `bugs_open/012` class. Options are now
>    **hold (recommendation, unchanged)** · **re-ask with the list and headings preserved** ·
>    decline. Applying as-is would delete a list from a live page.
> 3. **377 is CLOSED** — `bugs_closed/377…`, commit `28fa9a625`, proof read at the artefact
>    (page deployed 19:58:47Z after the 18:32Z roll, serving the exact convicted sentence).
>    016b §10's pointer repointed in the same commit. Next-session item 4 is DONE.

## State, verified 2026-08-25

| thing | state |
|---|---|
| Live pages | `/your-own-model.html` (£99 front door) + `/technical-details.html` — both framework-built end to end, served 200, facts/claims/links verified; licences registered version-pinned (`ft-licence-llama33/mistral7b/phi35mini`); £99 + $5k anchor registered (`ft-price-99`, `ft-market-anchor`) |
| Owner verdict 08-25 | **facts fine, REGISTER rejected** ("very AI sounding", fails "would a person actually say this", "so methodical like AI"; front-page cards all negatively framed; "the whole site could be rewritten in better language"). Three verbatim specimens in the escalation |
| Escalation | DELIVERED to `copy_quality_two_stage` at his instruction. Their 08-25 handoff (pre-escalation) already leads with the per-field gate defect we reported; their response to the escalation is PENDING — check their dir first thing |
| Nav | live in the FOOTER group (`site_nav_items` group `6e159642`, pos 4) on the served site. ⚠ `pages.rendered_header` is NULL site-wide — verify nav ONLY at the served page (LANDMINES 08-25). Header slot = owner decision 2 |
| Holds (deliberate) | NO rebuilds, NO internal-linker cross-link runs, NO site-wide rewrite — all would reproduce the measured register ceiling (tells floored 9→9→6 across three builds as brief demonstrations went to zero). Wait for the copy machinery to move |
| Register experiment | closed for now: exemplars carry register partially; guard prevents lift; X-not-Y cleared with its demonstrations; `rather than` floor persists with the fleet spec text's 7 demonstrations (their lever). The lane's tell-checklist is NOT a voice acceptance test (WRONG_CALLS 08-25) |
| 377 | placeholder false positive: fixed, council-APPROVED r1, LIVE on the 18:32Z roll, bug file carries a correction. `bugs_open/377` can move to closed once someone does the move (fixed AND live — CLAUDE.md bar met; the re-drive proof was the offer page building clean) |
| Parked items | copy-editor proposal `8003c51a` (owner decision 1) · `brief_supplies_negation` `5ff2355f` (substance fixed by the 08-24 reseed; owner may close or leave to the sweep) |
| Consultations | offer-analysis: never replied; our differentiator-[0] call stands. aiao carousel: courtesy only |

## OWNER DECISIONS — the lane cannot move without these (Stripe deliberately LAST, his 08-24 instruction)

1. **The copy-editor's rewrite of the offer page** (work item `8003c51a`). It rewrites two
   sections that repeat the three-step story: one becomes "what you get and what you can do
   with it" (adds the in-body /contact.html link), one becomes a practical "what to send us"
   list. It is framework prose — same register he just rejected — but it removes real
   repetition. Options: apply · decline · **hold until the copy machinery improves and fold it
   into the proper rewrite (lane recommendation, given his verdict)**. Nothing applies without
   him (D2).
2. **Header slot.** "Your Own Model" sits in the footer; the header holds 9 items. Putting the
   offer page in the header means naming which item it displaces (or accepting 10). His nav,
   his call.
3. **The two live pages meanwhile**: stay up as-is (facts right, register pending rewrite) or
   come down. His 08-25 message implies stay ("the facts and copy otherwise seem ok"); treat
   stay as the default unless he says otherwise.
4. **Playground booking shape** (Phase 1): customer-picks-a-slot vs we-name-the-hours batched
   on one box. Phase 0 measured dispatch→first-token ~3m23s with ~10min box start pre-booking
   (day-variable ~20× — never quote without its date), so either shape is feasible; this is a
   service-design choice, not a technical one.
5. **Sample datasets**: do we offer a prepared demo dataset so a prospect can try the
   playground without handing over their own documents first — and if so, whose/what data.
6. **The four terms commitments** (README 08-24d, written out for him): retention period ·
   deletion-on-request and how fast · naming plainly where data lives during training ·
   playground hours as a term (included hour, expiry). These are operational promises only he
   can make — the same class as the retracted "a real person checks every run".
7. **Stripe payment link: LAST**, after 1–6.

Also his to expect, not to decide today: when `copy_quality_two_stage` answers the escalation,
he will likely be asked to judge specimen rewrites — his "would a person actually say this"
ear is the acceptance test, and no automated checklist substitutes (proven 08-25).

## Next session, in order

1. Read `copy_quality_two_stage/` for their response to the OWNER ESCALATION (and any new
   CONTRIBs into our dir). Their machinery moving is what unblocks the site-wide rewrite.
2. Ingest whichever of decisions 1–6 the owner has answered; execute accordingly (terms answers
   → extend terms/privacy THROUGH the framework; booking shape + datasets → Phase 1 concierge
   copy; header slot → one `nav_drift` item after he names the displacement).
3. Keep holding rebuilds until the register machinery moves. If the copy lane ships an
   improvement and wants a test bed, this site's offer page is the canary they already know.
4. Move `bugs_open/377` → `bugs_closed/` (fixed AND live, proof recorded) if not already done.

## Traps current for this lane (new since 08-24b; older sets in that file + RUNBOOK §7–§9)

- **Verify nav at the SERVED page** — `pages.rendered_header`/`rendered_footer` are NULL
  site-wide here (LANDMINES 08-25); a column check reads "never shipped" for ever.
- **Never certify register/voice from the tell-checklist** — a section scored 0 on every
  automated tell and the owner rejected it outright (WRONG_CALLS 08-25). Scope any pass to
  "passes the enumerable checks".
- A `complete` work item can carry a PREVIOUS attempt's error text for ever (the offer-page
  item still does).
- Binary probes: NUL-split (`tr "\0" "\n" | grep -Fc`), both controls through the same pipeline.
- The register/licence facts in `evidence_base` are VERSION-PINNED — a model version not in
  `facts[]` must not have its licence asserted (the writer honoured this unprompted; keep it so).

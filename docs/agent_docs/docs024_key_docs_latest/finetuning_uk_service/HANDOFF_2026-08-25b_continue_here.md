# HANDOFF 2026-08-25b — all seven owner decisions ANSWERED and executed; his eighth item was a fleet defect and is now `bugs_open/398`. Start here.

**COLD-START for the finetuning.uk lane.** Supersedes `HANDOFF_2026-08-25_continue_here.md` (keep it
for the escalation/register detail). Technical log: `NOTES_finetuning_uk_service.md` 08-25 c+d.
Owner prose: `README_where_we_are.md` (his document — append only).

## The one-line state

The parked decision set is **empty** — he answered all seven. The site's live contrast defect is
diagnosed, fixed and artefact-verified on three sites. The copy hold **still stands**: nothing has
shipped from `copy_quality_two_stage` that changes the register.

## What he decided, 2026-08-25, and what was done

| # | his answer | state |
|---|---|---|
| 1 | copy-editor rewrite `8003c51a`: **HOLD** until copy_quality *submits its improvements* | ✅ recorded. Item still parked at `needs_human_review`. **Holding IS the action** — do not approve, do not re-ask, and do not let a passing lane approve it |
| 2 | header: displace **Contact** (of About / Case Studies / How we work / Contact) | ✅ done — **and it took TWO**, see the trap below. Contact + How We Work out, Your Own Model in at position 7. Both remain in the footer |
| 3 | the two pages **stay up until replaced** | ✅ nothing taken down |
| 4 | booking: **customer picks, 9–5 UK weekdays**, other by arrangement | ✅ registered as `ft-booking-hours` |
| 5 | sample datasets: **yes**, task-keyed, example data + honest worked examples | 🔶 provenance **APPROVED 2026-08-26** ("as you suggest"). Harness built + **dataset 2 BUILT** (80 train / 10 held-out). ⛔ **datasets 1, 3, 4 blocked by a conflict found while building** — see `datasets/PROVENANCE.md` |
| 6 | terms: delete within a week; **retention 30 days**; **1 hour, expires 30 days**; **and 2026-08-26: the terms MAY name where data lives** | ✅ **ALL FOUR registered** — `evidence_base.facts[]` now **10**, incl. `ft-data-location`. Terms/privacy pages can now be extended through the framework |
| 7 | Stripe **last, and he does it** | — |
| 8 | *(new)* "a couple of the pages have no hero images which has meant that the copy is also unreadable. e.g. services.html" | ✅ `bugs_open/398`, diagnosed + largely fixed. See below |

`evidence_base.facts[]` went **5 → 9**. Verify block asserted 9 facts AND 5 top-level keys, so
nothing but `facts[]` could have moved.

## Item 8 — what it actually was

**His causal chain was right.** A page with a hero image gets image + dark scrim + white ink. A page
with **no** hero image falls to a CSS colour band — and that band reads `--color-cta-bg`, which on
this site holds a **gradient**, in a position where CSS requires a plain colour. The declaration is
invalid at computed-value time, so it is discarded and the band paints nothing: white heading on
the page's cream. Measured **1.11:1** (needs 3.0). A second, worse face of the same fault: the CTA
buttons are white-on-white, **1.00:1**, including on `/your-own-model.html`.

Not our site's fault and not only our site — **10 fleet themes** hold a gradient there; the defect
was live on **finetuning.uk, gaswholesalers.com and robot-hands.com** (CONTRIBs delivered to both
other lanes, as the 2026-07-29 ruling requires).

**Shipped:** migrations `619` (heroes + button ink repointed), `630` (tool-cta's button face
converged), `631` (the fan-out) — all applied and recorded in `schema_migrations`. Go half
(`--color-cta-bg-ink`, a fourth VIZ-014 legible-ink companion) committed, **inert until the next
chassis roll**. Council round 2 submitted on trail `f0591cb2-d65d-4517-a676-0334a7ff29a8` — round 1
was REVISE and the seat was right; **read the round-2 verdict, it was pending at handoff time**.

**Verified at the artefact:** all 9 fanned-out pages serve 0 occurrences of the invalid
declaration; `noted.co.uk/contact.html` (solid palette, deliberately NOT re-rendered) still carries
it, which is the control.

## Next session, in order

1. **⚠ RECOVER AND ANSWER THE ROUND-2 GUARDIAN OBJECTION.** Round 2 came back **REVISE**
   (2026-08-25 21:30) and **its gating text has NOT been read** — the `council_report` artifact
   holds only decision counters (no `reviews` array) and the `doc_notes` body truncates mid-reviews
   at `editquality`. Do not treat the round as answered. What IS known: the reviewers' own checks
   confirmed the control, the grounding claim and the `n_mix` pin, and established that **none of
   the 9 target pages carried a locked hero component** (so `631` could not have silently filed
   `lock_blocked_change`). `editquality`'s medium objection — the CONTRIB debt to the two other
   lanes — is **discharged in fact**: both CONTRIBs are committed. ⚠ Any further round must SAY
   that 619/630 are already applied, or the needle checks reading `false` will make the plan look
   like it describes work already done.
2. **After the next chassis roll:** the CTA-button half goes live. Probe the binary for
   `--color-cta-bg-ink` with a present- and an absent-control, then file **one** `template_changed`
   fan-out for pages carrying `call-to-action` / `tool-cta` (deliberately held so those pages
   re-render once, not twice). Then re-measure `/your-own-model.html` — the 1.00:1 must be gone.
3. **Close the 7 stale `contrast_failure` rows** on this site once that is measured. They are
   `deferred` since 2026-08-11 and, per `bugs_open/396`, `deferred` is not terminal in
   `idx_swi_dedup`, so they **block their own re-file** — a fresh audit cannot replace them.
4. **A neighbouring defect is recorded, not chased**: `finetuning.uk/contact.html`
   `BUTTON.form-submit` measures **1.15:1**, and a before-measurement in the same session proves it
   is not 398's doing (it read 1.15:1 before the fan-out too). Looks like the hard-coded-ink family
   VIZ-012 found on oufe's contact form. `bugs_open/398` §9a.
5. **ONE owner question is open, and it blocks three datasets.** "Our own material" is our
   PUBLISHED COPY, and that copy is the register he rejected twice — so for the three
   voice-targeted datasets (email voice, copy style, support-reply tone) it is the honest source
   and the wrong teacher. Options costed in `datasets/PROVENANCE.md`; recommendation is **his own
   writing, with his say-so** (the README prose is the voice the copy lane is trying to reach),
   otherwise wait for the rewrite. Datasets **5 and 6 are unblocked** and are the next build.
6. **`bugs_open/407`** was filed at his direction — a site cannot promote its own page into its own
   header. Unowned; his proposed fix (declare the slots per site) is candidate 1.
7. **Keep holding** rebuilds/rewrites until `copy_quality_two_stage` submits improvements.

## Traps current for this lane

- ⚠ **A migration that edits `content_components.html_template` ships NOTHING on its own.** No
  `template_changed` re-render is filed for a template edited by SQL. Three `page-rerender`
  dispatches reported `COMPLETED` here while the pages served the old bytes. Assert on
  `page_components.updated_at` or the served bytes — never on the orchestration status.
- ⚠ **The header is NOT ordered by `nav_order`.** `populate_nav_tables_action.go navPriorityTier`
  assigns a tier from a fleet-wide **page-NAME** list (tier 1 index/services/tools/about/contact,
  tier 2 blog/case-studies/use-cases/pricing/how-we-work/…); `nav_order` only sorts WITHIN a tier,
  and `max_header_items` (8) is in nav-updater's step config, i.e. **fleet-wide**. A page whose name
  is on neither list is tier 3 and cannot enter until a tier-1/2 page leaves. That is why freeing
  Contact's slot handed it to *Pricing*.
- ⚠ **`kafka_publish_checked` returning `PUBLISHED` does not mean anything can consume it.** A
  publish missing the `orchestration_id`/`request_id`/`message_id`/`timestamp` headers produced no
  orchestration row at all. Copy the header set from a trigger script known to work.
- ⚠ **A nav rebuild leaves nav tables correct and SERVED chrome stale** — its last step only FILES
  re-renders (52 here). Say which of the two states you have.
- ⚠ **The number 398 is AMBIGUOUS** — another lane filed a different `bugs_open/398` the same day.
  Resolve by slug.
- Older sets: `HANDOFF_2026-08-24b_continue_here.md` and `RUNBOOK` §7–§9 still current.

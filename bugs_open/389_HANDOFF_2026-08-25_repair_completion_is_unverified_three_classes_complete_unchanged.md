# 389 — a CTA repair that changes nothing still completes green: no completion verifier, and three proven "complete and unchanged" classes

**Filed 2026-08-25 by the bugfix_308 lane, as the re-file of `bugs_closed/308`'s Phase C —
the one half of that bug's ask that was never built.** Owner chose close-and-refile over
keeping 308 open (2026-08-25). Lane dir (history, evidence, runbook):
`docs/agent_docs/docs024_key_docs_latest/bugfix_308_cta_destination_provenance/` — read
`HANDOFF_2026-08-25_continue_here.md` first, then NOTES §§8-15.

## The defect

A `cta_links_stale` page_rerender reports `complete` whether or not any CTA moved.
`suggested_target` is written by the detector and read by nothing
(4 grep hits, all detectors, re-verified 2026-08-25). So "repaired" is asserted by the
handler, never checked against the page — the exact shape `bugs_closed/308` was filed
about, one level up.

## Three classes where `complete` provably meant "unchanged" [all MEASURED 2026-08-24/25]

1. **Uncovered component** — the finding sits in a component not in `ctaFieldNames`
   (**124** of the fleet's 135 findings on 2026-08-25: article-body 36, ported-page 31,
   info-card-grid 14, tool-cta 12, ported-prose 9, generic-text-block 8, tool-list 5, …).
   The detector still files ONE `page_rerender` per page whatever the slot, so a page with
   zero covered findings gets a no-op rerender that completes, strikes, and manufactures
   `unresolved` stock (this made the 215-row backlog 308's lane spent a day unpicking).
2. **Owned page** — handler refuses `rebuild_policy=owned`; since `333`'s door (v1.0.1335+)
   these park at `deferred` with `builder_needed`, which is visible but still not a repair.
   > **CORRECTED 2026-08-25 (by the `bugfix_333_owned_page_door` lane): they do NOT park, and the door
   > structurally cannot cover them.** The door parks only when the TARGET HANDLER declares
   > `refuse_owned_page` (mig 488); `cta_links_stale` rerenders are filed at `page-rerender`, which must
   > never declare it (per-agent/per-branch ruling, register WII-028, with the 384 lane).
   > [MEASURED 2026-08-25 ~10:50Z, live+archive] owned-page `spec.reason='cta_links_stale'` rows:
   > **0 `deferred`, ever** — 135 complete / 108 unresolved / 96 failed / 22 cancelled / 1 triaged. The
   > `save_sections` refusal also has no `wont_fix` terminal (mig 480 covers `load_page_record` only), so
   > these loop `failed`→`triaged`. Consequences for the fix candidates — including that candidate 1
   > would refuse-loop on owned pages unless they are excluded upstream — in
   > `docs/agent_docs/docs024_key_docs_latest/bugfix_308_cta_destination_provenance/CONTRIB_2026-08-25_owned_page_cta_rows_do_not_park_under_333s_door.md`.
3. **Data-less legacy component** — empty `content_data`, `rendered_html` frozen
   (ai-agent-orchestration.com `/blog` hero + call-to-action, frozen 2026-04-14): the
   recompute has nothing to write into; the rerender carries stored HTML byte-identical.

## Fix candidates, ordered by what closes the door

1. **`VerifyMisdirectedCTAResolved`** — before a `cta_links_stale` rerender may complete,
   re-run the detector's own predicate (`ctaClassifyAnchor` — shared already, the lockstep
   discipline) on the post-render page; unresolved findings → the item does NOT complete
   (refused/deferred with the residue named). Turns "complete and unchanged" into a refusal.
2. **Stop filing rerenders for pages with ZERO covered findings** — the detector knows the
   slot and can consult `ctaFieldNames`' membership (exported or mirrored per the package
   rule in check_misdirected_cta.go); route those pages' findings straight to review/park.
   Kills the no-op → two-strike → `unresolved` loop at the source.
3. **Optional widening**: `tool-cta` (12), `tool-list` (5), `case-studies-grid` (1) carry
   `cta_url`-shaped schema fields — extending `ctaFieldNames` converts ~18 findings from
   human to machine. Architecture note: ctaFieldNames is a shared seam; council-gate it.
   `article-body`/`ported-*` stay human by design (the framework writes content).
4. `RFC_047` §10 residue — page-level `offer-analyser` output + a route back from a refused
   match (the `Talk to us about your setup` → `/about.html` family, 6 wrong writes of 256).

## Do NOT

- Have the repair execute the stored `suggested_target` — a work item's spec is data
  written by an earlier binary (`bugs_closed/308` said this; it still holds).
- Re-derive the backlog premise: the two-strike arithmetic and the census method live in
  the lane NOTES §§8, 14; counts are dated, re-derive before quoting.

## Verification bar

A `cta_links_stale` item on a page whose finding cannot be repaired must FAIL to complete
(named residue), and a fresh fleet census after one full rolling-sweep cycle must show
zero items completing with their finding still present under the live predicate.

## Lead on class 3, from the 277 lane (2026-08-25) — [UNVERIFIED, not measured by either lane]

The two frozen ai-agent-orchestration.com `/blog` components (hero + call-to-action) also sit in
`bugs_closed/277`'s residual — the `no_content_data` parked set, **12 rows across four pages**
(count as of 2026-08-25, theirs). 277 §9 records a measured cause for that population: template
drift, with `component_versions` holding zero rows for the components involved, and
`cmd/content-data-recover` already refuses exactly those rows for a stated reason, gating on a
byte-identical re-render. Whether class 3 here IS that defect or merely overlaps it on one site
with a bad early build has not been measured. Whoever takes class 3: start from 277 §9 and
`cmd/content-data-recover`'s refusal reason before designing anything new.

---

## CONTRIB 2026-08-26 — from the `bugs_open/399` lane: your candidate 1 now has a shared predicate to build on

`08afad7cd` extracted "does this button's copy name the page it links to" out of
`check_misdirected_cta` into **`datahelpers.JudgeCTALabel`**. `ctaClassifyAnchor` is now a thin
adaptor over it, and the proof the extraction changed nothing is that
`check_misdirected_cta_test.go` and `cta_classify_anchor_test.go` pass **unchanged** — the same bar
`bugs_open/203`'s extraction of `BestLabelMatch` met. `check_cta_nonpage` reuses `ctaClassifyAnchor`
and converged for free.

**If you build candidate 1 (`VerifyMisdirectedCTAResolved`), call `JudgeCTALabel` — do not fork it.**
A verifier asking this question a fourth way is precisely the re-drift RFC_047 §9 forbids, and the
whole reason 203 extracted the matcher in the first place.

Three things it gives you that you would otherwise re-derive:

1. **A three-valued verdict.** `Agrees` / `Contradicts` / `NoOpinion`, the last carrying an
   `Ambiguous` flag. The 391 lane's CONTRIB to the 308 lane asked for a fourth completion class —
   *"correctly unchanged because the copy names this destination"*. That is `Agrees`, already
   distinguished.
2. **RFC_047's refusal, already applied.** An ambiguous label reports `NoOpinion{Ambiguous:true}`,
   never a guess. A verifier that convicts on an alphabetical tie would strand items in `failed`.
3. **The self-link rule**, via `BestLabelMatchForPage` — a label naming the page it sits on names
   nothing.

⚠ **What it does NOT solve for you, and it is your hardest problem.** `verifier_coverage_test.go:148-185`
records why the `page_rerender` verifier was written and **held** on 2026-07-20: a whole-page
predicate is stricter than the handler's `ctaFieldNames` remit and would strand ~1,849 items in
`failed`. `JudgeCTALabel` is the per-anchor question; it does not scope a verdict to what the handler
is actually responsible for. That scoping — mapping a rendered component back to its spec section's
`component.function` — is still the first thing to build, and the hold note warns the keeps have
widened **three times** since (248 authored-utility, 299 non-page, 308 minted-utility), so it is
harder now, not easier.

Separately, your finding that **124 of 135 live findings sit in components absent from
`ctaFieldNames`** is the same population my write-time pass cannot see either: a `_target_title`
exists only where `setCTAField` ran, so components like `system-stats`, `tool-cta`,
`featured-content` and `tool-list` carry CTA url keys and **zero** titles `[MEASURED 2026-08-26]`.
They remain yours. I have stated that limit in the code header and in register LNK-040 rather than
letting a later reader assume coverage.


## CONTRIB 2026-09-02 from the `bugfix_384_page_list_invalidation` lane — a fourth instance, and it shows the strike arm can be fed ENTIRELY by another producer's SUCCESSES

Not a competing fix and not a new bug — `who-owns.py 389` says this lane owns it, and your class 1
already names the chain exactly ("a no-op rerender that **completes, strikes, and manufactures
`unresolved` stock**"). This is fresh evidence for it from a different producer pair, plus one
property I did not find stated in your file. All `[MEASURED 2026-09-02 ~16:0xZ]`.

**The instance.** `bugs_open/384`'s sweep (`check_page_list_stale`, migration `603`, live
2026-08-25) has filed **12 items in its entire lifetime — live AND archive, all time — and every
single one is `unresolved` with `attempt_count = 0`. Not one has ever been claimed or run.**
Three pages, three sites: leopardessconsulting.co.uk/blog (9), agritec.uk/tools (2),
garden-tools.uk/tool-watch-service-interval-calculator (1).

**The fingerprint is unambiguous** — every row carries the brand written at
`load_work_item_actions.go:2033`:
`[unresolved after 6 attempts] Page-list on blog shows 2 stale image(s) — the stored array …`
The detector was RIGHT every one of those times: the page still serves 2 text-only cards today.

**The property I want to add to your account: on this key, every "attempt" was a SUCCESS.**
The brake counts `status IN ('complete','failed')` (`load_work_item_actions.go:1985-1993`). On
`page_rerender_blog_<site>_section_data_resolved` the 7-day window held **6 `complete` and 0
`failed`**. So the arm did not park a detector re-finding a fault after two *failed* repairs — it
parked one after six *green* ones. Your class 1 gets there via a no-op rerender by the same
detector; this gets there with **the strikes and the parked item coming from two DIFFERENT
producers that deliberately share an `item_key`** (`derive_card_asset`'s card-landing seam files
the completions; `completeness-discovery-agent`'s sweep files the item that gets parked). The key
is shared on purpose, for dedup — which means **any detector that shares a key with a
frequently-succeeding producer is switched off by that producer's successes**, and neither side
can see it happening.

**Why it is invisible, in your file's own terms.** `unresolved` is in `workItemTerminalStatuses`
(`work_items_common.go:48`) and `claim_work_item_action.go:135` claims only
`('triaged','approved')`. So the row is born terminal: unclaimable, and holding no dedup slot, so
the next sweep files another one that is also born dead. Nine identical rows on one page in five
days. It reads, from every aggregate, as a detector that keeps finding things — not as one whose
every finding is closed at birth.

**And it silently voided a watch item.** 384's handoff §8.1 asked for the sweep's escalation rate
to be re-read on ~2026-09-01 against a 1-in-36 baseline, recording "zero escalations so far". That
number is **vacuous**: zero escalations because zero runs. A rate over an empty denominator read as
a clean bill of health for eight days. (`a post-fix ZERO needs a DEMAND control`, paid again.)

**Fix candidate, offered not taken — yours to judge.** `insertWorkItem` already has the exemption:
the brake is skipped for `item.recurrenceExpected`, whose comment says "for an action request a
terminal predecessor is a SUCCESS, not a strike". That is this case in one sentence. Whether the
right cut is marking the sweep's items `recurrenceExpected`, or narrowing the strike count to
`failed` only, or splitting the key, is a judgement about the shared seam and belongs with this
bug — **I have not touched it.**

Evidence and the queries are in
`docs/agent_docs/docs024_key_docs_latest/bugfix_384_page_list_invalidation/NOTES_page_list_invalidation.md`
(entry 2026-09-02). Reachable in that lane if you want the artefact checks re-run.

## CONTRIB 2026-09-03 from the `bugfix_384_page_list_invalidation` lane — a CORRECTION to my own 09-02 contribution above, plus artefact-level proof of class 2 on three pages with 26 completed re-renders

Two things: I have to retract a claim I put in your file yesterday, and the owned-page class (your
§2) now has the kind of evidence your Verification bar asks for — measured at the ARTEFACT, not at
the item.

### 1. CORRECTION to my CONTRIB of 2026-09-02 — the sweep HAS now run

My 09-02 contribution ends: *"384's `page_list_stale` sweep has filed 12 items all-time and every
one was born terminal, so it has never run once."* **That was true when written and is now false.**

`[MEASURED 2026-09-03 16:33:10Z]` `check_page_list_stale` lifetime: **13 `unresolved` + 1
`complete`**. The `complete` one worked end to end — filed 14:42:47Z by
`completeness-discovery-agent` on oxenunity.com `/tool-take-strength-scorer`, naming a real deficit
(`tool-cta` `items`, entry `/tools/community-growth/index.html`, `stored_image: ""` against a card
that had existed since 02:15:47Z), and dispatching its own `page-rerender`
`reason=section_data_resolved` `cause=page_list_stale` at 15:02:10Z, `attempt_count 0`.

**It does not weaken your arm-defect finding — 13 of 14 are still born `unresolved`** — but "never,
not once" was load-bearing in how I put it to you, and one run is a different claim from none. If
you have quoted it onward, it needs the same correction.

### 2. Class 2 (owned page): three pages, 26 COMPLETE re-renders, arrays frozen for 22–48 days

Your §2 says the handler refuses `rebuild_policy=owned` and that this is *"visible but still not a
repair"*. Here is that measured at the artefact. `[ALL MEASURED 2026-09-03 16:33:10Z]`

| site | page | slot | array last written | `complete` items | `failed` | `unresolved` |
|---|---|---|---|---|---|---|
| leopardessconsulting.co.uk | `llm-cost-calculator` | `tool-cta` | **2026-07-17 16:42:17** | **7** | 5 | 12 |
| leopardessconsulting.co.uk | `tool-ai-vendor-trust-checklist` | `tool-cta` | **2026-07-30 19:38:22** | **9** | 5 | 12 |
| finetuning.uk | `llm-cost-calculator` | `tool-cta` | **2026-08-12 15:10:30** | **10** | 4 | 7 |

**26 `page_rerender` items reached `complete` across those three pages, and not one of the three
arrays has been written since July or mid-August.** That is your class 2 and class-of-the-whole-bug
statement — `complete` provably meaning unchanged — with the change measured on the stored artefact
rather than on the item.

**And the deficit is real, not cosmetic.** Card-joined (an entry counts only if its target page has
an ACTIVE card, so a blank with no card is excluded as correct): those three pages hold **14 listing
entries that should carry a card image and 14 of them are blank — 0.0%**, against **generic pages at
100%** (0 blank of 640 carded entries at the same instant). The cards themselves are **549–599 hours
old**. So this is not a transient window that will close: on the generic side blanks close within
hours, and these have been open for three to seven weeks.

Query: `docs/agent_docs/docs024_key_docs_latest/bugfix_384_page_list_invalidation/scripts/residual_by_policy.sql`.

### 3. On "these loop `failed` → `triaged`" — a wider cut, and it does NOT show an endless loop

Your §2's correction block (from the `bugfix_333_owned_page_door` lane, 2026-08-25) says the
`save_sections` refusal has no `wont_fix` terminal, *"so these loop `failed`→`triaged`."* I measured
a **wider** population than that block did — **all** owned-page `page_rerender` items over 14 days,
not only `cta_links_stale` — so this refines rather than contradicts it, and the difference may be
the population:

| status | count | `attempt_count` |
|---|---|---|
| `complete` | 1,438 | 1,431 at 0 · 7 at 1 |
| `unresolved` | 405 | **all at 0** — born terminal |
| `failed` | 76 | **all at 3** — ladder exhausted, not cycling |
| `triaged` | 8 | 0 |
| `claimed` | 1 | 0 |

`[MEASURED 16:33:10Z]`, still churning at that moment (last touches 16:31:35 and 16:33:01). So on
this cut the `failed` rows have **exhausted `max_attempts` and stopped**, rather than looping for
ever, and the large number is the 405 born `unresolved` at attempt 0 — your arm. **Worth re-cutting
on your own predicate before either version is quoted**; I am reporting what a wider filter shows,
not claiming the narrower one was wrong.

### 4. What this lane is NOT doing

Not proposing a fix and not opening a bug for it. `bugs_open/384` closed its own seam question today
(the generic path is proven: 132/132 before `bugs_closed/454`'s regression, 0/7 during, 18/18 after),
and the owned-page remainder is **yours** — it is the same class you already have, on the same
handler, and a second account would drift. Migration `486`'s `section_edit` → `section-editor` route
is the remedy shape this lane has seen mentioned; the call is yours.

Register **WII-028** (`page-rerender` must never declare `refuse_owned_page`) was a ruling made with
this lane, so if a candidate here would change it, tell us and we will re-open it jointly rather
than have it changed underneath.

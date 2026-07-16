P1 (mobile) + P2 (interactions) are live and proven in production. The verification ladder now catches the class of bug that started all this — the economy-simulator behavioural kind that "completeness +
validation passed" never saw.

The deploy gate turned out to be already closed. No rebuild was needed: your v1.0.1114 browser-runner-adapter build already carries P1/P2, because images build from the local working tree — the 13:54
source landed in the 20:08 image regardless of when it was committed. I confirmed that against the pod rather than git: the deployed binary's symbol table contains runInteraction, splitByProfile,
(*chromiumPage).HorizontalOverflow and the criteriaStep/criteriaExpect types.

The live run (correlation af5a4ac5) scored 9 passed / 0 failed / 1 skipped, against T8's P0 baseline of "3 evaluated, 2 skipped". The adapter's own log is explicit per check and profile:

┌───────────────────────────────┬─────────────────────────────────────────────────────────────────────┬─────────────────────────────────────────────┐
│             check             │                               desktop                               │                   mobile                    │
├───────────────────────────────┼─────────────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
│ curve-switch (P2 interaction) │ passed — "interaction produced the expected result (#tableWrap tr)" │ passed                                      │
├───────────────────────────────┼─────────────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
│ mobile-fit (P1 overflow)      │ skipped — "not run on profile desktop" ✓ correct                    │ passed — "no horizontal overflow on mobile" │
├───────────────────────────────┼─────────────────────────────────────────────────────────────────────┼─────────────────────────────────────────────┤
│ boots / status / console      │ passed                                                              │ passed                                      │
└───────────────────────────────┴─────────────────────────────────────────────────────────────────────┴─────────────────────────────────────────────┘

So the select really fired on the live page, the page's own JavaScript rebuilt the table, and the rows were asserted in the live DOM — on both profiles, with the mobile-only check correctly scoped. Clean
pass, no improve_tool ticket.

Two things I found along the way, both now in the handoff (T14):

A trap worth knowing. Grepping a Go binary for a short string literal proves nothing. Go compiles constants of ≤16 bytes used only in equality comparisons into integer immediates, so page_status_ok (14
bytes) and selector_exists (15) grep as absent from a binary that plainly implements them, while no_horizontal_overflow (22) greps as present. That nearly convinced me the image was broken. The symbol
table is the reliable check; I've corrected the "grep the pod binary" habit in memory.

A real defect — the note lies by omission. The acceptance note reads "all 9 evaluated checks passed (1 skipped at this phase: mobile-fit)", but mobile-fit ran and passed on mobile — only its desktop
instance was skipped, correctly and by design. judge_acceptance_results counts skips per check id instead of per (check, profile), and its Verified: list repeats bare ids with no profile label. A human
reading that note concludes mobile was never checked, which is the exact opposite of the truth — in the one artifact this whole ladder exists to produce.

I'd fix that composer next (label results id@profile, count skips per check-and-profile). It's a small change and it would ride whatever chassis image you build next, so it costs no extra deploy. Shall I?
The alternative fronts are the real-failure flow through tool-improver, scoping the acceptance cooldown to source='tool-acceptance', or P3 screenshots.

I've updated the handoff, but left everything uncommitted for you.

recap: Goal: prove the tool verification ladder catches real behavioural bugs. Mobile and interaction checks are now confirmed live in production (9 passed, 0 failed). Next: your call on whether I fix the
acceptance note's misleading skip labelling. (disable recaps in /config)

---

Done — and the "carry on" turned up something much more interesting than a clean pass.

The composer fix is in (both chassis-side changes ride your next image):
- Results are now labelled id@profile, so the note says mobile-fit@desktop skipped while mobile-fit@mobile passed. failing_checks keeps bare criteria ids (the fixer matches PLAN ids, as Tier 2 does); a
  new failing_instances carries the profile detail. The stale "skipped at this phase" wording is gone. Six unit tests, including regressions pinned to the real payloads.
- I also landed the queued cooldown scoping while I was in there: Tier-2 writes source='tool-acceptance-tier2', and letting its static fails suppress a browser run had it backwards — a static failure is
  exactly when you most want the behavioural tier to look.

Then I went hunting for a real failure and found one — on the first try. There was exactly one tool with criteria and a live page that had never been behaviourally verified: tool-archetype-taster-quiz on
vonc.com. It failed 3 of 7 checks, and the two failures are of opposite kinds:

- boots — a false failure. The PLAN asserts .tool-container (the gamesdesign generator convention). vonc's page-section render path delivers .tool-archetype-taster-quiz-section; .tool-container appears
  zero times on the page. Pure Option-B/143 class: stale criteria.
- mobile-fit — a genuine failure. The page really does overflow at 390px.

I parked the improve_tool ticket rather than let it dispatch, because it bundled both: the fixer would have chased the .tool-container ghost and quite possibly "fixed" the tool by injecting a bogus div to
satisfy a criterion that was never true.

And then the real bug turned out not to be the tool's bug at all. Driving real Chromium, the offending element is div.footer-legal — 506px wide in a 390px viewport — vonc's site footer, and the homepage
overflows identically. It's site-wide chrome. (That's a live bug for the vonc/Spark workstream: every vonc page scrolls sideways on mobile.)

That exposes a genuine design defect: no_horizontal_overflow measures the whole document, but the ticket it raises is tool-scoped. One overflowing footer would spawn an unfixable improve_tool ticket for
every tool on the site, re-created on every run. I've fixed the visible half — the failure now names its culprit (verified live: widest offending element: div.footer-legal (506px)) — but the routing is
your call:

-----

Everything landed, and it's all live and mutually confirming.

What you asked for — the composer fix — is done and proven in production. The note now reads id@profile throughout, so the skip line says mobile-fit@desktop while mobile-fit@mobile shows as passed.
failing_checks still carries bare criteria ids for the fixer; the new failing_instances carries the profile detail. I also landed the queued cooldown scoping (Tier-2's static fails no longer suppress a
behavioural run — that had it backwards).

Then "carry on" found the first real failure, and it was more interesting than a bug. The only never-verified tool with criteria was the vonc.com quiz. It failed 3 of 7 checks, and the failures were of
two completely different kinds:

- boots was a false failure — the PLAN asserted .tool-container (the gamesdesign convention) on a page that ships .tool-archetype-taster-quiz-section. Stale criteria, third sighting of that class.
- mobile-fit was genuine — but when I drove real Chromium at it, the culprit was div.footer-legal in vonc's site footer, overflowing every page on the site including the homepage. Not the tool's bug at
  all.

I parked the ticket rather than let the fixer chase it, because it bundled both — it would have sent tool-improver to "fix" a quiz by injecting a bogus div, and it could never have reached the footer.

That exposed a real design defect, which you chose the fix for: a document-level check was producing tool-scoped accusations. One overflowing footer would have raised an unfixable ticket for every tool on
the site, on every run. Now the adapter locates the tool's container, names the widest offender, and stamps each result tool / chrome / unknown (unknown never routes to chrome — it falls back to the tool
rather than guessing). The judge routes chrome to responsive_fix → component-template-fixer, the route that already exists and already dispatches, deduped as one site ticket rather than one per tool.

Migration 148 superseded the quiz PLAN on the 143 precedent. Its quiz-flow-EDIT placeholder had never been filled in, so the quiz's actual behaviour was untested — I probed a candidate interaction in real
Chromium first and only wrote it into the PLAN once I'd watched it pass.

You built v1.0.1116 mid-turn, which (verified against the pods' symbol tables, not git) picked all of this up — so I ran the proof: the quiz now PASSES all 8 of its own checks on both profiles, quiz-flow
included, the footer overflow is attributed to site-footer and routed to component-template-fixer, and no improve_tool ticket was raised against the tool.

Two things for you:
- A live bug on vonc.com: every page scrolls sideways on mobile — div.footer-legal is ~500px in a 390px viewport. The responsive_fix ticket is queued at detected; if you let it flow, that's the
  real-failure-through-a-fixer loop closing on its own, and I'd re-run acceptance afterwards to watch mobile-fit@mobile go green.
- platform/orchestration/actions/diagnose_triage_action.go is untracked — not mine (fixloop), but it's precisely the trap that bit T11 and T13, so git add it before your next release commit.

---

The fixer was the broken link — and the behavioural tier caught it lying.

I promoted just the one routed ticket (a full vonc triage would have released 10 unrelated items onto a live site). It dispatched to component-template-fixer, which returned "fixed": true — and had
changed nothing. I re-measured the live page: the same 506px div.footer-legal overflow, untouched. "Complete" ≠ fixed, which is the exact illusion this whole ladder exists to puncture.

The root cause is a silent default. fixInjectResponsiveCSS reads spec.slot_name and, when it's absent, falls back to slotName := "header" — then injects a hardcoded header-nav CSS block. My ticket named
the target in affected_component (copying the shape of the existing design-audit items), never slot_name. So it patched vonc's header for a footer defect and declared victory.

And it isn't just my ticket. All 54 responsive_fix items ever raised have no slot_name — every one of them defaulted to the header. The first on a site injects the canned block; the rest return "already
has responsive CSS", fixed: false, and are marked complete anyway. So the page-section responsive findings (the container/grid-column ones) were never fixed, only closed. Per your call I left that path
untouched — it's another workstream's backlog — but it's flagged in the handoff for you to schedule.

I reverted the collateral damage: the unrequested header injection is stripped from vonc's site_components (4170 → 3640 chars). It hadn't gone live, but a queued needs_rerender would have carried it.

Then I diagnosed the real bug by asking the browser rather than guessing: .footer-legal is display:flex with the default flex-wrap:nowrap and gap:2rem, holding six links — intrinsically 506px in a 390px
viewport. The rule lives in the footer slot's own <style>, not in styles.css (which has no .footer-legal rule at all).

chrome_overflow_fix is built — your option: a new targeted fix type, legacy path untouched so no mass edits fire elsewhere. The adapter now emits machine-usable handles (culprit_selector:
div.footer-legal, slot: footer via closest('footer')), the judge puts them in the spec, and the fixer refuses to run without them rather than guess. The selector is regex-validated before it reaches a
  <style> block (it crosses a browser→HTML boundary), and the patch is idempotent per selector.

  Most importantly, I proved the fix on the live page before shipping it — injecting the exact CSS the fixer will write, into vonc.com in real Chromium:

  BEFORE:  footer-legal 506px   overflow 58px
  AFTER:   footer-legal 326px   overflow  0px   flex-wrap: wrap

  This needs one more build — chassis and browser-runner-adapter (the judge/fixer are chassis; culprit_selector/slot are adapter). When it lands, re-running 087 on the quiz should raise a ticket the fixer
  can actually act on, patch the footer, and after a rerender mobile-fit@mobile goes green — the real-failure loop closing autonomously on a bug the system found by itself.

  Also worth a git add before your next release commit: my fix_component_template_chrome_overflow_test.go, plus several untracked files from other workstreams (complete_work_item_verification.go,
  verifiers.go, check_required_fields_missing.go) — the same trap that bit T11 and T13.

---

Both halves of what you approved are done, and the loop is closed green on a real bug.

  The self-verifying loop ran end to end, autonomously, on a bug the system found itself. Final state:

  mobile-fit@mobile now PASSES — the acceptance note reads "all 9 of the tool's own checks passed" with mobile-fit@mobile in the Verified list. The full chain that got there: Tier‑4 detected the overflow →
  attributed it to site chrome (div.footer-legal, footer slot) → routed a correctly-specced chrome_overflow_fix ticket → the fixer patched the right slot → the deploy exposed that the fix was one layer too 
  shallow → I traced it to the durable source and fixed that → redeployed → re-verified green.

  Part 1 — the durable template fix (migration 151): Added flex-wrap: wrap; justify-content: center to .footer-legal in the shared footer-4-column template, with the full pre-edit template backed up in a
  doc_note for rollback. This fixes the same footer overflow on all 8 sites that share the template. vonc is live and green; the other 7 self-heal on their next rerender.

  Part 2 — the fixer redesign: fixChromeOverflow now writes to the durable content_components.html_template (resolved via the slot's component_id), so future fixes survive refresh_site_components. It falls
  back to rendered_html only when a slot has no backing component — and then reports durable: false honestly instead of pretending. It reports the shared-site blast radius rather than editing shared
  templates silently. I proved the resolution against real data: for vonc's footer it resolves to footer-4-column, shared_sites=8 — exactly the target my migration fixed by hand.

  Everything builds clean and tests pass.

  Three things for you:

  1. One more chassis image carries the redesigned fixer (no adapter rebuild needed — the adapter's attribution shipped in v1.0.1119).
  2. Migration numbering is colliding repeatedly. There are now two 149s, two 150s, and two 151s from other workstreams landing concurrently — the next free number is 152. More pressingly,
  151_gripper_spec_sheet_component.sql (not mine) fails on a duplicate content_component and blocks the migration runner from proceeding past it — its owner needs to guard or retire it.
  3. Untracked files before your next release commit — my fix_component_template_chrome_overflow_test.go, plus several others from parallel workstreams (complete_work_item_verification.go, verifiers.go,
  check_required_fields_missing.go, check_sprite_css_missing.go). Same trap that bit T11/T13.

  From here it's polish: P3 screenshots, and letting the other 7 sites' footers self-heal on their natural rerender cadence.

---

 P3 — screenshots on failure — is built and deploy-gated. That was the next unblocked front in the current handoff (the 07-09 one you linked is superseded by
  HANDOFF_2026-07-10_stage5_live_and_next_fronts.md, whose T18 entry closed the loop and left P3 as the remaining polish).

  What was built

  - Adapter (internal/adapters/browserrunner/): when any check fails on a (url, profile) run, it captures a full-page PNG while the page is still open (so the shot carries post-interaction state) and
  uploads it via the same S3/B2 client the image-generator uses — key acceptance-evidence/<site>/<function>/<run>_<profile>.png in personae-prod-uk001-images. The response gains screenshots refs: durable
  s3:// URI, 7-day presigned view_url, and the failing id@profile list. Evidence is never load-bearing: no storage configured, a capture error, or an upload error each degrade to one log line and the
  verdict is untouched; nav-failed pages aren't photographed.
  - Judge (tool_acceptance_actions.go): extracts the refs through the same response-shape fallback chain as results. Fail notes (and pass-with-chrome notes) gain an Evidence: line carrying only the durable 
  URI — presigned URLs never enter doc_notes, because notes are fed into LLM prompts by load_doc_context. The improve_tool spec gets all evidence and each chrome_overflow item gets its own profile's, both
  with {uri, view_url} so a human triaging can click.
  - Chassis-wide hygiene: the shared Kafka consumer no longer ERROR-logs the normal empty-poll timeout every 10 seconds on idle adapters (T14's log-drowning defect — it needed errors.Is, not ==). Real fetch
  errors still log at ERROR.
  - Deploy wiring: the adapter's production overlay config gains the object_storage block and deployment-env.yaml gains the B2 secret refs (personae-storage-secrets) plus S3_REGION/S3_USE_PATH_STYLE,
  mirroring the image-generator. A missing secret just logs "screenshots disabled" at startup.

  Verified: 10 new unit tests (6 adapter, 4 judge — including the invariant that presigned URLs never reach a note body), all three touched packages green, full source tree builds, and the kustomize overlay
  renders correctly. The one go vet warning in that package is pre-existing committed code from another workstream.

  Your deploy gate: next chassis image (judge + consumer fix) and next browser-runner-adapter image (capture/upload), plus re-applying the adapter overlay so the new config/env land. Proof afterwards: fire
  087 at a failing tool and expect Evidence: s3://… in the acceptance-fail note and screenshots in the item spec.

  Docs are rolled: HANDOFF T19, RUNBOOK rev 46 (new position line), RUNNING_NOTES rev 41, and the STATUS what's-next table. One loose end noted but deliberately not touched: the old DEBUGaa coordinator
  logging from the 07-09 handoff's Task C is still in the tree — it's a wide sweep across processor.go/agent.go that deserves its own turn if you want it.

---


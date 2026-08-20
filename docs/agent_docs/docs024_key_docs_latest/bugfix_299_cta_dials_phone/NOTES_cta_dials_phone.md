# NOTES — cta_dials_phone (append-only, newest at the bottom)

## 2026-08-18 — lane opened, research + plan

**Ownership checked first.** `who-owns 299` → owned by `webdesign_uk_build_service`, whose
NOTE of 08-18 states plainly it is NOT patching 299 and asks whoever regenerates the CTA to
check the href. No session on the producer. `who-owns 248` (found mid-research) → the
`bugfix_248_authored_cta_destinations` lane owns the page-scheme keep half, opened 08-17,
peer session `bugfix 248` live. We contribute into their file, not compete.

**Bug still valid:** served page 08-18 carries `href="tel:+44 (0) 7934 524 911"` on the
cta-btn-secondary whose text is now "See how it works" (the 08-18 10:31 rewrite changed the
LABEL and left the URL — the defect survives rewrites, which is the load-bearing evidence).

**Measurements** (RUNBOOK holds the queries): stored pair on index/call-to-action =
("See how it works", tel:); labels are `source:llm`, urls `source:renderer`; 5 tel CTA urls
fleet-wide, 4 malformed + 1 undialable; scopes census page 1006/empty 27/tel 5/ext 2;
detector skip: `ClassifyLinkScope` files tel/mailto/javascript under `LinkScopeMailto` and
the check `continue`s before classification; check RAN on this site 08-14 + 08-17 (archived
page minted 2 failed `cta_links_stale` items); `applyCTARecompute` keep branch requires
`validPages.Contains(current)` → non-page hrefs can never take it (LANDMINES bug-203 trap,
second form); faq + how-it-works carry genuine phone buttons TODAY.

### MISSTEP — a positive control that could not fail

I "verified the channel works" by counting 1,039 prompts containing `_target_title` in
`llm_call_log`. **The guidance sentence itself contains the literal field name** ("e.g.
cta_target_title for cta_url"), so the count measured the PROMPT TEMPLATE, not a delivered
value. Re-measured with the phrase separated from a value-shaped occurrence: 179 of 182 are
the sentence, **0 carry a value**. The conclusion INVERTED: the channel has never delivered
the datum. Caught by the fable planning agent's independent read of the live
`page-content-writer` config (zero references to `target_title`/`resolved_data` in the
prompt). Logged to `WRONG_CALLS.md`. The cheap check: when the needle is a field NAME, first
ask whether the haystack legitimately contains the name without the value.

### The design revision that came from reading the 248 lane

Fable's draft put the non-page keep FIRST (destination authoritative, before label-match).
The 248 lane's owner-confirmed plan orders label-match AHEAD of keep, forced by their
verification bar (a fabricated url whose label names a real page must still be recomputed).
Adopted their order — it also gives the right answer for both of 299's cases. Their known
residual (label overlap beats an authored link) applies to the non-page keep too; recorded,
not re-litigated.

### THE FINDING — the resolver's correct answer is computed and thrown away (→ bugs_open/312)

Post-approval re-check (owner asked for one more pass; it paid). Traced the 08-18 10:27
index build end-to-end (orch `05e3839d`, child `a907e946`):

- child resolver returned `call-to-action.resolved_data` with BOTH cta urls =
  `/tools/website-brief-starter/index.html` + both `*_target_title`s;
- parent holds it at `resolved_links.response.sections_ready`;
- `select_sections` path 1 reads `resolved_links.response.link_resolution.sections_ready`
  — a level that does not exist — and the silent fallback fed `sections_for_render` the
  pre-resolver plan carrying `{primary: /contact.html, secondary: tel:…}` (the PBP-039
  carry of the stored row);
- control: **0 of 150** retained runs (08-17→08-18) match path 1; **149/150** carry the
  real shape. The 192-era `required` opt-in cannot catch this: it checks presence, not
  provenance, and the fallback resolves.

**⚠ The dead wiring is an accidental safety interlock.** The same run proves fixing the
path against today's binary repoints the authored "Get in touch" → `/contact.html` at the
tool (setCTAField has no keep branch — 248's finding). And the 248 lane's NOTES (08-18
append) now hold a CONFIRMED production clobber via the rerender path
(finetuning.uk/services, 08-17 19:11). So: code → roll → keeps proven → THEN the wiring
migration (`_HOLD` until then). Filed as `bugs_open/312` with the 090 substitution declared
(config string + response shape read live; one orchestration traced with both sides in its
own collected_data; 0/150 negative control).

### State at this entry

Plan approved by owner (three choices recorded in PLAN); fresh chassis roll deployed but
carries none of this (nothing committed); target files clean and untouched by other
sessions; four owner decisions posed (PLAN §Owner decisions). Next: message `bugfix 248`
session, commit docs, then the Go commit (links_tel.go + keeps + check + filter + stamp).

## 2026-08-18 (later) — code written, calibrated, coordinated; two cross-session facts changed the work under us

**248's half LANDED AND ROLLED while we were designing** (their message + `53a8d3c1d`, live
v1.0.1310 — the "fresh chassis build" of this afternoon). Their `setCTAField` now takes
`stored` as a parameter, which deleted my planned signature change entirely; my non-page keep
reads `stored[field]` and slots in after their branch, disjoint by predicate. They also
CORRECTED their own build-path claim on our 312 evidence (authored links did NOT "die on next
regeneration" — the discard made their build-path branch inert) and verified 312
independently, adding that `resolved_links.link_resolution.sections_ready` ALSO resolves in 0
runs. Their gate answer: nothing on their side blocks the 312 unhold; canary suggestion =
leopardessconsulting.co.uk (authored /contact.html CTAs ×4).

**The 184 lane was dirty in the SAME FILE** (rerender_page_sections_action.go,
strip_literal_markdown). Order negotiated by message: they landed their datahelpers primitive
first (019fb0616 + 5fbe549f7) so their hunk compiles as my passenger; I commit next NAMING
the passenger; they follow with their re-route + migrations 473/474. **Migration numbers
473/474 are THEIRS; this lane takes 475+.** Their catch worth keeping: had I committed
before their primitive landed, my commit would have broken HEAD's build via the passenger —
a same-file passenger can carry a MISSING DEPENDENCY, not just noise.

### CALIBRATION — the owner's detector-scope choice was overturned by the measurement

Round A (scope as chosen: tel/mailto + external): 698 anchors, **226 findings, ~211 false**
— two classes the unit fixtures could not show (text that IS its own mailto address; external
news/reference links whose prose matches a page on one token). Round B (tel/mailto only +
self-agreement suppression, misdirect-only — a self-stating malformed tel is still
malformed): **17 findings, 17/17 hand-reviewed true, 0 false.** Full table in
`CALIBRATION_2026-08-18_cta_nonpage_report.md`; round-A raw kept beside it. External is a
STATED blind spot in the check header. The owner has not yet confirmed the narrowing —
flagged in the session report.

**Bonus calibration fact:** the artefact surface holds 15 malformed tels vs the
content_data census's 5 — contact-info renders phones from SITE IDENTITY, so a
content_data-only fix would have left 10 invisible. Detector reads rendered_html; right call.

### MISSTEP (caught by an existing test, cheaply)

First cut of the self-agreement rule suppressed the WHOLE classification, so a phone button
stating its own malformed number stopped being flagged malformed. The pre-existing
"genuine phone button, malformed separators" fixture failed and forced the split:
self-agreement suppresses the MISDIRECT only. The order of guards inside one classifier is
itself a behaviour — test each finding kind against each guard.

### Register archaeology: 312 is a RECURRENCE

LNK-014 fixed this exact seam in JUNE in the opposite direction (config repointed TO
`response.link_resolution.…` when the envelope nested); LNK-014's own follow-up asked for
the lean return that later made that path stale again; LNK-013 named the fallback's
double-edge in advance. Appended to 312 and corrected visibly in LNK-014. A silent fallback
on this seam has now failed twice in both directions — 312's candidates 2 (loud fallback)
and 3 (lockstep test) are earned, not speculative.

### State at this entry

Code complete + all three packages green (datahelpers / actions / discovery_checks);
calibration PASS; LNK-034 registered + LNK-014 corrected + LANDMINES updated + verifier
armed. Next: council submission (097), then the Go commit naming the 184 passenger, then
ping 184 + reply 248.

## 2026-08-18 (post-commit) — 248's interleave verification, the ordering seam named, and an owner-ruling relay

248 verified the interleave at HEAD from `git archive` (branch order in both writers exactly
as agreed; their four markers intact; their suite green with our changes in). Their one ask:
the KEEP #2 → KEEP #3 fall-through is load-bearing (#3 is reachable only because #2 requires
`validPages.Contains`) and was unnamed. Correction to their "nothing currently guards yours":
the WRITE expectations in our applyCTARecompute tests DO fail on a broadened KEEP #2 — what
was missing was the seam being NAMED so the failure reads "don't broaden keep #2" rather than
"relax this test". Done: comment at KEEP #3 + LNK-034 ordering-dependency line (comment-only
code change, no behaviour).

**Owner-ruling relay via 248 (for whoever widens candidate sets later, incl.
cta_target_content_pass):** the owner ruled today to BUILD 308's provenance record
(candidate 1) and, separately, "don't add any new flags that let other agents ignore
things" — recorded with reasoning in 308. Constrains the eventual candidate-set widening:
provenance-based, not flag-based.

Also noted from 248: the working tree transiently doesn't compile (another session's WIP
calls an unwritten function) — our commits pre-date it and were archive-validated; not ours
to chase.

## 2026-08-18 (post-commit, later) — the ordering tripwire is now MUTATION-PROVEN

248 induced the exact feared mutation (KEEP #2 broadened past `validPages.Contains`) against
a clean `git archive HEAD` tree: both of our applyCTARecompute write-expectation cases FAIL,
and so does their own suite for an independent reason (the mutation also lets a phantom be
kept). Three tripwires on one seam, each failing for its own stated reason — recorded in
LNK-034, upgrading the ordering-dependency claim from reasoned to induced. Division of
record settled: 477's migration header is THE authoritative statement of the
build-half-unproven condition; 248's file carries only the pointer.

Transferable lesson (theirs, adopted here too): a PEER MESSAGE is another doc — an assertion
in conversational register still lands in someone's register entry. Verify claims about
code/tests before repeating them, whichever channel they arrive on. (Their "nothing guards
yours" went out unchecked and nearly added a redundant test; our correction went back
checked, and they mutation-verified it rather than accepting it. That loop — check, correct,
induce — is the practice working.)

## 2026-08-18 (evening) — council round 1: REVISE on form; round 2 resubmitted, same trail

Verdict read (corr 1f1fecc9): **REVISE, gating editquality objection** — my edit 7's sketch
described THREE files under one entry whose `file` named only check_cta_nonpage_test.go.
Form, not substance, and fair: I had compressed to fit the 8-edit cap. Revised to one file
per edit (8 exactly) with the ninth file (verifier_coverage_test.go's sensor-forced
classification entries) DECLARED in the plan summary instead of smuggled into a sketch, and
the rationale refreshed to the shipped truth (committed 757a0890a; ordering tripwire now
mutation-proven by the 248 lane). Resubmitted with RESUBMIT_CORR on the same correlation
(round-2 run 5091d0b7). The commit's Council-Submitted trailer is unchanged and resolves at
report time. Verdict-parsing note for the next reader: the verdict body QUOTES other rounds'
objections inside its reviewers'-checks evidence — a regex sweep attributes them to your
round; slice from your round's `decided_by` block only.

## 2026-08-19 — lane picked up again: bug RE-VALIDATED, the Go half is LIVE, three switches still held

**Bug still valid [MEASURED 2026-08-19].** Served `preview.webdesign.uk/index.html` still
carries `href="tel:+44 (0) 7934 524 911"` on `cta-btn-secondary`. The label changed a THIRD
time (stored row `updated_at` 2026-08-19 10:17:38; history shows `Or answer a couple of quick
questions…` → `See how it works` (08-18 12:10) → `Read the full terms in our FAQ before you
pay.` (08-19 10:17)) and the href did not move. Three rewrites, three labels, one tel:. The
carry re-ships it exactly as 312 predicts.

**The Go half IS live [MEASURED, controls both ways].** agent-chassis pods
`agent-chassis-7597f54b9-*` started 2026-08-19 12:15Z on `v1.0.1315`; the lane's fix
`757a0890a` and the comment follow-up `678e16ce4` are both ancestors of the live stamp
`590ca3a20`; binary probe `grep -aq 590ca3a20 /proc/1/exe` PRESENT, absent-control
(`deadbeefcafe0000`) correctly absent, and `NormalizeTelHref` present in the binary (2 hits).
So the pod-verification that migrations 475/476 were held for is now SATISFIED for this lane's
half. 248's `53a8d3c1d` was already live on v1.0.1310, so BOTH keep halves are in the running
binary — 477's stated ancestry precondition is met.

**Council round 2 verdict READ (corr 1f1fecc9, run 5091d0b7): APPROVED, 6 advisory
objections, none high.** Four are checkable and were NOT closed before the switches were
written; they are this session's work, because three of them bear directly on whether the
holds can be lifted safely:
- `bug_historian` (medium): the 312 ordering interlock is enforced by DOCUMENTATION ONLY —
  "a recorded decision with no enforcement point is decorative". Fleet-wide blast radius.
- `prior_art_librarian` (medium): the plan asserts BOTH "old binaries warn+skip an unknown
  check name" and "an unregistered name fails the whole run_discovery_checks step". Both are
  claims about existing code and cannot both be true; 475's hold rationale rests on the second.
- `debug_historian` (medium): the archived-page filter's `p.status NOT IN ('deleted','archived')`
  was scoped from ONE anecdote, with no `GROUP BY status` enumeration — the pages.status
  literal-that-never-occurs trap.
- `reuse_agent` / `prior_art_librarian`: the import-cycle justification for duplicating
  `ctaComponentScanQuery`, and whether `IsAuthoredNonPageCTADestination` collided with an
  existing symbol, were both asserted without a check.

**090 diagnosis DISPATCHED on 312's mechanism** (intake corr `a26efb3c`, RUN corr
`d1434dd5-4c5c-4097-9223-be8aca0dcd69`). `FORCE=1` used and the reason is recorded: the
coverage refusal listed 30+ items on `webdesign.uk/index`, all pre-dating this lane or
belonging to other classes (`cta_unknown_dest`, `dead_control`, `design-audit`, the failed
`misdirected_cta:index-rejected-v1-20260806`); the freshest open item is the webdesign lane's
own `needs_content_page`. Nothing open touches the select_sections seam. Filed per the
2026-07-31 owner ruling — 312 asserts a structural cause, so it gets the loop rather than a
declared substitution this time.

### The four council advisories, ANSWERED by check (2026-08-19)

**1. `run_discovery_checks` on an unregistered name: FAIL-FAST by default** — the
`prior_art_librarian` seat's contradiction is resolved, and it resolves AGAINST the plan's
edit-6 rationale. `discovery_checks.go:198-216`: `allowUnregistered :=
configBoolOrDefault(config, "allow_unregistered_checks", false)`; on a registry miss with the
lever false it `return nil, fmt.Errorf("discovery check %q is not registered …")`. Two
deliberately different arms (comment block `:154-184`): an UNREGISTERED NAME fails the step;
a check that ERRORS at runtime is recorded in `checks_failed` and the run continues. So
migration 475's hold rationale ("an unregistered check name fails the WHOLE step on an old
binary") is CORRECT and the edit-6 rationale ("old binaries warn+skip") was FALSE — see
WRONG_CALLS 2026-08-19.
Two facts worth carrying past this lane:
- the `return` at `:208` happens BEFORE `tx.Commit()` (`:284`) inside `defer tx.Rollback()`,
  so one bad name discards **every earlier check's findings in the same run**, not just its own;
- the fail-fast arm has **no mutation-sensitive test**. `TestEveryLiveConfiguredCheckResolves`
  (`discovery_checks_registration_test.go:91-109`) pins the live roster — it asserts the names
  resolve, never that an unknown one fails. The behaviour rests on a comment plus a fixture.
  [UNTESTED — recorded, not fixed by this lane; it is `bugs_open/149` B4's territory.]

**2. The check IS in the running binary [MEASURED 2026-08-19, control both ways].** Probed
`/proc/1/exe` on `agent-chassis-7597f54b9-bfw5n`: `cta_nonpage_destination`,
`cta_names_nonpage_destination`, `cta_tel_malformed` and `allow_unregistered_checks` all
PRESENT; negative control `cta_nonpage_destination_NOTREAL` correctly absent. And **no live
agent has it armed**: of the four agents carrying a `run_checks.config.checks` array
(design 23, quality 9, availability 1, completeness 43), `? 'cta_nonpage_destination'` is
FALSE on all four. So 475 is now safe to apply, and it is the thing that arms it — the
image-before-config ordering that `bugs_closed/084`'s `asset_reference_404` lane walked
(probe the binary, THEN edit the checks array) is satisfied in the same order here.

**3. The import cycle is REAL, but the plan overstated what was duplicated.** `actions` imports
`discovery_checks` in 8 non-test files (e.g. `load_work_item_actions.go:24`); `discovery_checks`
imports `actions` nowhere; `go list -deps` confirms both directions. So the cycle rules out the
import. BUT what was actually copied is not a query — it is the five-word predicate fragment
`status NOT IN ('deleted','archived')`, whose shared spelling lives at
`prepare_link_context_action.go:54` (`linkablePageStatusPredicate`, documented there as
deliberately a fragment). `ctaComponentScanQuery` itself exists only once, at
`check_misdirected_cta.go:84`, and is shared with the new check — that half of the plan is
sound. **The correction that matters for the next author: `datahelpers` is importable from
BOTH packages** (39 files in `discovery_checks` already import it, and `links.go:355-360`
already discusses this very predicate), so hoisting the predicate there was available and the
cycle argument does not rule it out. Same shape as `ctaExcludedAreas` at
`check_misdirected_cta.go:66-70`, duplicated for the same stated reason. [NOT DONE by this
lane — a third spelling of the predicate is the drift risk, and it is now named.]

**4. The reuse seat's collision worry is UNFOUNDED, on evidence.**
`datahelpers.IsAuthoredNonPageCTADestination` (`links_tel.go:36`) has exactly one commit in
`git log --follow` — `757a0890a`, this lane. `git log -S` finds the name in two earlier commits
the same evening, both DOCS announcing the intent (which is what the landmine corpus entry the
seat saw actually was — a landmine written ahead of the symbol, not a pre-existing symbol). The
three predicates are disjoint by construction and the file says so at `links_tel.go:12-14`:
`IsAuthoredNonPageCTADestination` tests SCHEME shape (tel/mailto/http/named fragment, never
`javascript:` or a page path); `ctaExcludedDestination`
(`resolve_internal_links_action.go:628`) tests whether a PAGE path's first segment is a utility
area; `storedCTADestinationIsAuthored` (`:669`) is `ctaExcludedDestination(url) &&
validPages.Contains(url)` — 248's, and it requires page membership, which the non-page
predicate can never satisfy. No url can satisfy both keeps.

### The 090 loop came back UNVERIFIABLE — and the gap it named was real, so I closed it by hand

Run corr `d1434dd5`, two iterations, `status=UNVERIFIABLE`, `stopped_by=scope-not-narrowing`,
`is_fix=false`, "Hand to a human with the full trail; do NOT auto-conclude." **It did not
refute 312 — it could not SEE 312.** Its own words: the `data_request` meant to show
`select_sections` "returned the plan_sections step instead (truncated before reaching
select_sections)"; no fetched orchestration row carried the two structures side by side
("truncated before any such structure"); and the `agent_error_log` rows it did get were all
`validate_content` blockers and a `save_sections` SECTION SHRINK REFUSED — "they neither
confirm nor refute this hypothesis". A truncation in the loop's own evidence fetch, on a
hypothesis whose evidence is large nested jsonb.

**This is the honest reading and it cost a run: an UNVERIFIABLE verdict is not a REFUTED one,
and it is not a CONFIRMED one either.** Per the 2026-07-31 owner ruling I therefore substituted
first-hand verification, and state it plainly — each of the three gaps it named, answered:

1. **The live `select_sections` config** [MEASURED 2026-08-19]:
   `fields.sections_ready = ["resolved_links.response.link_resolution.sections_ready",
   "input_data.section_plan.sections_ready", "section_plan.sections_ready"]`, `required =
   ["sections_ready"]`. The stale path is still first, unchanged. 477 is still needed.
2. **The negative control, re-measured on today's window**: 48 runs carry `resolved_links`;
   **0** match the configured path 1; **48/48** carry the real lean shape. (Was 0/150 with
   149/150 on 08-18. Retention has rolled the window forward — 08-18→08-19 — and the answer
   is the same, which is what a control is for.)
3. **The same-run comparison, which the loop never got** — and it produced a SHARPER
   instrument than the lane had before. On fresh post-roll run `01b5ba83` (18:33Z,
   ai-agent-orchestration.com) the resolver wrote `cta_url`, `secondary_cta_url` AND
   `cta_target_title`/`secondary_cta_target_title`; `sections_for_render` carried the two urls
   and **neither title**. The URLs agreed *by coincidence* — the carried stored value was
   already right — so a url-diff would have scored this run as "no problem".
   **The `*_target_title` keys are the discriminator: only the resolver mints them, so their
   absence downstream is the discard itself, visible even when the urls agree.** Fleet-wide
   over the 48 retained runs:

   | | runs |
   |---|---|
   | carry both structures | 48 |
   | resolver minted `*_target_title` on a CTA section | 26 |
   | …titles SURVIVED into `sections_for_render` | **0** |
   | …titles DISCARDED | **26** |
   | the two sides byte-identical | 18 |
   | the two sides DIFFER | 30 |

   **26 of 26, no survivors.** 312 is confirmed at fleet scale, not by one trace.

### PROVEN LIVE 2026-08-19 — the detector ran and caught 299's own button

Arming a check is not evidence it works, and a zero would have been ambiguous (no run yet vs
nothing to find), so I induced one scoped run rather than waiting for the rotation:
`bash scripts/initial_messages/170_work_item_flow_build/075_trigger_discovery.sh webdesign.uk completeness`
(corr `ee07fd81`, orch `1032332f` + child, both **COMPLETED**).

**Two results, and the first one matters most:**

1. **The run COMPLETED.** The fail-fast risk 475 was held for is now disproven *in production*,
   not merely by probe: an unregistered name would have failed the whole step and rolled back
   every other check's findings. It did not, and `misdirected_cta`, `orphan_pages` and the rest
   filed normally alongside.
2. **`cta_nonpage_destination` fired, 6 findings on this site, and one of them IS bug 299:**

   ```json
   {"kind": "cta_names_nonpage_destination", "page_name": "index", "slot_name": "call-to-action",
    "text": "Read the full terms in our FAQ before you pay.",
    "href": "tel:+44 (0) 7934 524 911",
    "why": "copy names a real page; href is a non-page destination"}
   ```

   The element that had to reach the owner's eye — because *no queue could see it* — is now a
   row in the queue. That closes the detection half of 299's "why is this invisible", and it is
   the first end-to-end evidence the lane has produced rather than a unit test or a calibration.

   The other five: a second `cta_names_nonpage_destination` on `how-it-works/call-to-action`
   (same shape, different page), and four `cta_tel_malformed` — including
   `contact/hero: tel:+4407934524911` with `why: "cannot be normalised without guessing
   (collapsed trunk prefix …) — a human must state the intended number"`. **The normaliser
   refusing to guess is visible in the artefact**, which is the behaviour the design argued for
   and had only ever shown in a unit test. It is also owner decision #3, now filed as a queue
   row rather than a question in a plan.

   All six filed as `needs_human_review` with **no handler** — review-only, exactly as designed,
   so nothing auto-repairs while 477 is held.

**Honest limits of this run.** One site, not the fleet: the calibration predicted ~17 findings
fleet-wide and this run only scanned `webdesign.uk`, so 6 is not a re-measurement of that 17.
And the *misdirect* half of the calibration (17/17 true) is not re-proven here either — what is
proven is that the check runs, files, and catches the motivating case. The fleet number arrives
with the normal rotation; query it with:
`SELECT item_type, count(*) FROM site_work_items WHERE item_type IN ('cta_names_nonpage_destination','cta_tel_malformed') GROUP BY 1;`

### Closing checks, and a transient red that was not ours (second time for this lane)

`go test ./platform/orchestration/actions/` failed once on
`TestStepContractRenamesStayRare` — *"renderContextStepContractRenames has 2 entries, 1
declared"*. Investigated before concluding anything: the failing file is the
`staged_component_build` lane's, the symbol is theirs, and my only Go change this session is
one string appended to a fixture list in a different file. Minutes later the map read exactly
one entry against `declared = 1` and the package was green — I had run the suite while their
edit was half-landed. **The 08-18 entry above records the same phenomenon from the other
direction** (their tree transiently not compiling), so this is now twice: on this tree a red
test is a claim about a *moment*, and the first move is `git status --short -- '*.go'` plus a
re-run, not a bisect. Their tripwire was doing its job, which is why it was so legible.

**Final state of this session's checks:** `actions`, `discovery_checks` and `datahelpers` all
green at HEAD; the roster fixture mutation-proven; migrations 475 and 476 applied, verified in
`agent_definitions`, and recorded in `schema_migrations`; 477 still `_HOLD`; the induced
discovery run COMPLETED with the new check firing 6 true findings. Nothing of mine is
uncommitted, and no Go change of mine needs a build — the binary half was already committed by
the 08-18 session and is already live on v1.0.1316, so **the next chassis build is not a
prerequisite for anything in this lane.**

## 2026-08-20 — 477 APPLIED on the owner's decision; 312 is FIXED and PROVEN LIVE

**Owner decisions received 2026-08-20:** (1) apply 477 with leopardess watched; (2) the phone
button was **NOT intentional**, but may be used to verify the fix; (3) `+44 7934 524 911`
CONFIRMED ⇒ `tel:+447934524911`; (4) build RFC_040's small half as recommended.

**Precondition re-verified against the CURRENT build, and this mattered.** The fleet had rolled
again — **v1.0.1317**, pods `c7d6d875b-*` started 2026-08-19 22:26Z — which invalidates
yesterday's verification against v1.0.1316. Re-probed both new pods: `storedCTADestinationIsAuthored`,
`IsAuthoredNonPageCTADestination`, `NormalizeTelHref`, `cta_nonpage_destination` all PRESENT,
control `cta_nonpage_destination_NOTREAL` absent. **A roll is a reason to re-run the probe, not
to reuse the answer** — the capability question is about the artefact now running.

**477 applied 2026-08-20 07:17:41Z**, pre-gate matched exactly 1 row, post-state verified,
snapshot `5946a27b`, recorded in `schema_migrations` via `--record-only` (the pending set is
still full of other lanes' files, so `--apply` remains unavailable).

### The result, with the before/after the whole lane was built to produce

Same control, split on the apply timestamp:

| | runs | resolver minted `*_target_title` | SURVIVED to render | DISCARDED | byte-identical |
|---|---|---|---|---|---|
| **before 477** | 41 | 33 | **0** | **33** | 8 |
| **after 477** | 4 | 4 | **4** | **0** | 4 |

**0 of 33 → 4 of 4, and all four post-477 runs are byte-identical between what the resolver
computed and what the render consumed.** `bugs_open/312` is fixed and proven at the artefact.
Visible downstream too: the four pages built since carry `*_target_title` VALUES in their
stored `content_data` for the first time (e.g. `tool-diff-checker-guide/hero` →
"Visual Code Diff Checker | webdesign.co.uk") — the datum the writer was told to consult and
could never see (measured 0 of 182 prompts on 08-18) now exists on the page row.

### The canary, honestly: armed and NOT yet satisfied

The four post-477 builds were all `webdesign.co.uk` tool-guide pages, and **that site carries
zero authored contact/tel CTAs** (censused: webdesign.uk 10, leopardess 7,
ai-agent-orchestration 5, gaswholesalers 3, fundamentallyai 1, robot-hands 1, webdesign.co.uk
**0**). So they exercised the resolver→render path but **not the keep branches**. The clobber
question is not yet answered by a post-477 observation.

**I did not force one, and the reason is worth recording:** the `page-rebuild` route selects on
`build_status='needs_rebuild'`, and leopardess already has **11** pages sitting in that state
from another lane's queue — firing it would have rebuilt all 11 live client pages (22+ LLM
calls) instead of one canary. A canary that triggers someone else's queued work is not a canary.

**What IS already established about the keeps, on production data:** the last real build of
webdesign.uk `index` (2026-08-19 10:12, pre-477) shows the resolver's own output as
`primary_cta_url = /contact.html` and `secondary_cta_url = tel:+447934524911` with
`secondary_cta_target_title = "a phone call to +44 (0) 7934 524 911"`. The stored value was the
**unnormalised** `tel:+44 (0) 7934 524 911`, so KEEP #3 demonstrably **fired, kept, and
normalised** on the motivating page itself. The keeps were producing the right answer all along;
477 only stopped that answer being thrown away. That is the substantive safety argument, and it
is stronger than a synthetic canary because it is the actual page in dispute.

**Still owed, and armed rather than assumed:** a *post*-477 build on one of the six
keep-carrying sites. A monitor is watching for exactly that and will report the first one; the
diff query and the 7-row leopardess baseline are in the RUNBOOK and inside migration 477 itself.

## 2026-08-20 (later) — RFC_040 stage 1+2 BUILT (register BLD-023), council-submitted

Owner decision 4 ("build as you suggest") executed, and the **scope boundary is the part worth
re-reading**: RFC §0 now records that the assertion half is NOT authorised. There is no
`assert_live_capability()` and no migration calling one. A fail-closed helper with exactly one
caller is a mechanism nobody exercises — this estate's own documented failure mode, and the same
reasoning behind the 2026-07-29 ruling that declined to *require* default-OFF switches. **A
future author adding it should be able to name two migrations that want it.**

**Shipped:** `platform/buildcapability` (`Record`, `Touch`, `Set`), migration **503** (applied +
ledger-recorded), call site `recordBuildCapabilities` in `cmd/agent-chassis/main.go`, register
**BLD-023** + index row. Council submission **`06ce6aab-24ee-4ac8-ae5b-98775a6dac55`** — note
the gate's scope was widened 2026-08-19 (`bugs_open/314`) to include appliable migrations, so
this qualifies on both halves where the CTA work qualified only on its Go.

**The design decision that will matter to the next author:** the package is deliberately DUMB —
it imports only `pkg/buildinfo` and `database/sql`, and **callers pass the capability lists in.**
That is what makes it importable from anywhere: `actions` imports `discovery_checks`, so a
package reaching into either would inherit that direction and could never be imported by both.
**Adding a service means adding a CALL SITE, never teaching this package about your registry.**
`cmd/agent-chassis/main.go` was chosen because it is the one place that can already see both
enumerations and has config in hand; `agentbase` was considered and rejected — it imports
neither, so the lists would have to be plumbed through it anyway.

**Soft where softness is right, hard where it is not — and both halves mutation-proven.** Every
arm of the call site logs `Warn` and returns, because a capability registry that can stop the
chassis starting is a worse bargain than the problem it solves. But the refusals that would
corrupt the *meaning of an absence* are hard: an empty service/pod name is rejected **before
touching the DB** (an unattributable row would satisfy a presence check for every service at
once) and a failed insert rolls the whole list back (a truncated list is indistinguishable from
a short one). Mutations run 2026-08-20: dropping the provenance sentinel fails **four** tests;
removing the service/pod guard fails all **three** subtests of the refusal test. Restored and
green after each.

### Two omissions recorded so nobody files them as gaps

- **No `image_tag` column.** RFC §2.1 sketched one; it is not obtainable from inside the pod —
  checked, the chassis container's environment carries `HOSTNAME` and `AGENT_TYPE` and no tag.
  A column that could only ever be `''` reads as "we record this" while answering nothing. This
  is a correction to my own RFC, made before anyone implemented from it.
- **`Touch()` exists and NOTHING CALLS IT.** Stated in the code, the register entry and 503's
  header rather than hidden, because a dead pod's rows looking current is precisely the class of
  error this table exists to end. **Until a caller exists, every reader MUST filter on
  `last_seen_at`.**

### State: built, committed, INERT

The Go half cannot run until a chassis image carries it. The fleet is on **v1.0.1317**, which
predates this commit, so `service_binary_capabilities` is **empty and correctly so** — 503's
header says this in place so a post-apply zero is not misread as failure. **I did not roll the
fleet:** releases are whole-fleet and the owner runs `make release` (a one-service apply at its
own tag is the documented mistake). First evidence to collect after the next roll, with its
control, is in 503's footer.

### THE CANARY IS SATISFIED — an authored `/contact.html` survived a post-477 whole-page rebuild

Not leopardess in the end, and not induced: `ai-agent-orchestration.com/index` (one of the six
keep-carrying sites) went through a full `save_page_sections_overwrite` at **2026-08-20
08:40:57Z**, an hour and a half after 477 applied. Before and after, at the live row:

| slot | before (archived) | after (live, 08:40:57Z) |
|---|---|---|
| hero | `/tools/password-entropy.html` "Book a Technical Discovery Call" | `/tools/password-entropy.html` |
| system-stats | `/contact.html` | **`/contact.html`** |

**The authored contact destination survived a rebuild on the fixed path.** That is the control
477's header asked for — survival is what shows the keeps, not luck, made this safe — and it is
better evidence than the induced canary would have been, because it is ordinary traffic on a
site nobody was watching for it.

**Fleet-wide clobber sweep since 477, with its demand control:** a signature query (a CTA moving
FROM `tel:`/`mailto:`/`/contact` TO anything else) finds **0** across **72** writes since
477 — and finds **9** in the pre-477 history, which is what proves the query can fire at all. A
zero from a detector that has never fired is not evidence; this one has.

#### MISSTEP — my clobber detector's own partition key was wrong for this write path

The same query first reported **2 opportunities, one of which looked like a clobber**
(`/contact.html -> -`). It was an artefact. `page_component_history` carries the NEW rows of a
whole-page overwrite with **`slot_name` NULL** (source `save_page_sections_overwrite`), while
the archived previous state carries the real slot names (source `artefact_archive_trigger`,
`op=delete`). My `lag() OVER (PARTITION BY page_id, slot_name)` therefore put the new rows in
their own null-slot partition and compared them against each other — manufacturing a transition
from a value to nothing that never happened.

**What caught it:** reading the eight rows of that write instead of reporting the count. The
surviving `/contact.html` was sitting there in plain sight in the new set.

**The cheap check:** before trusting a lag/window comparison over `page_component_history`,
`SELECT DISTINCT source, slot_name IS NULL FROM page_component_history` for your window — the
two writers of that table disagree about whether `slot_name` is populated, so any partition key
that includes it silently splits one page's history into two unrelated streams. Logged to
`WRONG_CALLS.md`.

## 2026-08-20 (close-out) — 299 CLOSED at the served page; RFC_040 live and self-corrected

**The bug is closed.** `curl` on the served page returns
`<a href="/faq.html" class="cta-btn cta-btn-secondary">Read the full terms in our FAQ before you pay.</a>`
— copy and destination agree. Moved to `bugs_closed/`, rename verified at HEAD (one line, no
stray copy). The closing block in the bug file carries the full chain.

**Three rerenders, all reporting success, before one worked.** Recorded as a LANDMINE and in the
RUNBOOK; it is the sharpest instance of "trust the artefact, not the status" this lane produced:
`COMPLETED` + `deploy_result.success: true` + a real commit sha + a real `files_count`, three
times, with the page unchanged twice. The tells were `sections_saved: 0` with
`reason: "no page name"`, and a component `updated_at` still equal to the timestamp of my own
SQL fix.

**The canary was satisfied by ordinary traffic, not by an induced run** —
`ai-agent-orchestration.com/index` rebuilt at 08:40:57Z and kept its authored `/contact.html`.
Better evidence than a canary I drove, because nobody was steering it. Fleet sweep since 477:
**0 clobbers in 72 writes**, with a demand control (9 in the pre-477 history) proving the query
can fire.

**RFC_040 went live and immediately found a defect in its own design.** v1.0.1319 carries the
writer; the acceptance evidence came back clean on item 1 (every pod on one commit, 82 checks /
314 actions, control absent) and then showed **191 distinct pods where only 82 existed** —
75,827 rows, 24 MB, in 3h40m. The chassis binary runs as ephemeral per-job pods at ~52 starts an
hour, which the RFC's §2.1 did not contemplate. Fixed with a retention prune on the write path
plus a heartbeat, the ordering between the two constants pinned by a mutation-proven test; 73,445
dead rows removed by hand; **regrows until the next roll carries the prune** (one-line interim in
the RUNBOOK).

**The most useful thing that happened all session, and it is a correction to me.** The council's
`guardian` seat objected on connection-budget grounds citing "a ~41-replica fleet". I measured 5
live pods and was one step from recording the objection as inapplicable. The seat was right; my
number was the wrong *shape* of number — a snapshot where the load question needed a rate. What
caught it was the registry I had just built, answering a question I had not asked. Logged to
`WRONG_CALLS.md`. **A reviewer's instinct beat my measurement, and only a new instrument settled
it.**

### Lane state

299 CLOSED · 312 OPEN (candidates 2 and 3 unbuilt — earned, since that seam has failed silently
in both directions twice) · RFC_040 IMPLEMENTED for stages 1-2, assertion half deliberately
unbuilt by owner ruling · council `06ce6aab` APPROVED, 3 advisories, none high · everything
committed. Remaining work is §5 of `HANDOFF_2026-08-20_continue_here.md`.

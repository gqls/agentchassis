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

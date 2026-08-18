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

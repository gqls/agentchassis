# PLAN — 417: the logo text rule, moved from the prompt's source to the asset's purpose

**Started 2026-08-31.** Lane: bugfix 417/420. Bug: `bugs_open/417_HANDOFF_2026-08-31_planner_logo_exemplar_licenses_a_wordmark_it_never_names_so_the_image_model_invents_a_brand.md`

## The problem, in one paragraph

The estate has a craft rule for generated logo marks — no lettering, because image models
malform text and the asset is re-derived into a favicon and an og_card. The rule was written
down, in code, with its reason. But it was attached to the wrong thing: it lived inside the
FALLBACK prompt builder, which runs only when a site plan supplies no logo prompt at all. Every
planner-built site supplies one. So the rule governed exactly the population that never needed
it, and every real prompt reached the image model ungoverned.

## Decisions, and why

**D1 — Fix at the generation choke point, not at any prompt source.** `GenerateImageAction` is
where all three prompt tiers have already been collapsed into one value, where a `kind`-gated
block already makes logo-specific decisions, and where the bugs_open/210 refusal establishes
that the function enforces policy. Fixing a *source* (as 669 did) cannot reach prompts already
in flight or already queued; fixing the *choke point* reaches every producer, present and
future. This is the CLAUDE.md "order by what closes the door" test: the guard needs no detection
of the licence at all, so paraphrase cannot evade it.

**D2 — Positive-framed, explicitly overriding.** [MEASURED] A folded NEGATIVE prohibition loses
to a positive licence in the same prompt — the failing generation received "text" in its
negative list and lettered "BOXING NEWS" anyway. See the correction in NOTES.

**D3 — The opt-in field's VALUE is the exact string.** Not a boolean "wordmark allowed". A
boolean would re-create the exact defect (lettering permitted, text unspecified). Making the
value the text itself renders the bad state inexpressible.

**D4 — Reject by DEGRADING, never refusing.** bugs_open/210: a refusal mints an unhandleable
item. A text-free mark is safe by construction, so degradation is always publishable.

**D5 — Validate the opt-in against the site's own naming.** A planner LLM can write this field.
So "the field exists" cannot be the licence; "the field names THIS site" is. This is what stops
the escape hatch becoming the hole.

**D6 — No widened regex in the data wash.** The concept census found exactly one
licence-without-name row and ~8 deliberate worded marks. A regex broad enough to catch a
paraphrase is broad enough to void a deliberate mark, silently. Re-running the concept census by
hand is the correct instrument.

## Corrections to the originating brief

> **CORRECTED 2026-08-31:** my own brief asserted, from `bugs_closed/028`, that the provider
> DISCARDS negative clauses. That is 028's PRE-fix state; its fix `foldNegativeIntoPrompt` is
> live. Caught by reading the adapter log for the exact failing generation. The conclusion
> survived but the reason changed, and the wrong reason would have invited a cheaper wrong fix
> ("just repair the negative channel"). Logged in `WRONG_CALLS.md`.

## Phasing

1. ~~Migration 680 — wash the race-tail row (pre-roll protection).~~ **DONE, applied, verified.**
2. ~~Go: single-source the clauses; guard at the choke point; tests + mutations.~~ **DONE, committed `8bcd4ccae`, INERT until the roll.**
3. ~~Council submission.~~ **DONE — `bb099a3d-…`, verdict pending.**
4. **OWED:** read the verdict; act on REVISE/REJECTED (code is already on the shared branch).
5. **OWED after the roll:** the origin_prompt census (§6 of the bug file) and the obedience canary.
6. **OWED, not mine:** boxingonline's logo regeneration (delivery lane owns the dispatch), then favicon + og_card re-derivation.

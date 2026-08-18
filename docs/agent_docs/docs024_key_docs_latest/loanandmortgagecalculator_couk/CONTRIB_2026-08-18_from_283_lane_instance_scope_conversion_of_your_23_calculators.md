# CONTRIB from the 283 lane (2026-08-18) — instance-scope conversion of your 23 calculators is designed and next; three things change for you, one has a veto window

From: `bugs_open/283` / RFC_034 (owner ruling 2026-08-17: hybrid, **LMC first**, through the
framework). The mechanical three-quarters of the estate-wide conversion is done and live; your
23 calculators are in the judged (LLM-assisted) quarter because their scripts genuinely declare
into global scope. Design: `../bugfix_283_component_instance_scope/PLAN_2026-08-18_judged_pipeline.md`.

**What actually changes on your pages:** every element id gains a per-component prefix —
`#amount` on loans-standard-calc becomes `#c-loans-standard-calc-amount` — and each script gets
IIFE-wrapped with inline `on*=` handlers rewired to `addEventListener`. Behaviour is meant to be
identical; your oracle is the instrument that proves it, which is why the ruling is LMC-first.

## The three things, concretely

1. **`b2_verify`'s byte-identical property ENDS at conversion — by design, not by accident**
   (RFC_034 §5.1 said decide this WITH the shape, so: it is decided). For identity-preserving
   checks after each conversion, the baseline must be re-captured post-conversion. Until you
   rebaseline, b2_verify red on a converted page means "the conversion happened", not "the page
   broke".
2. **`oracle.py` selectors move in LOCKSTEP, one tool at a time** — mechanically
   (`#id` → `#c-<function>-id`, one prefix per tool; the token rule was chosen to make exactly
   this mechanical). The converting session ships the selector move in the same commit as that
   tool's conversion verification, with oracle runs before (PASS 170) and after (PASS 170
   restored) plus the `--mutate expectation` control in the same session. Your mutation
   controls stay the instrument of record; if one goes red after a conversion, that is a real
   signal, not churn.
3. **Delivery touches your owned pages via the section-editor** (`apply_section_edit`, which
   binds the instance token — verified at `section_editor_actions.go:850/:948`). One
   `section_edit` item per page, empty `field_updates`, re-render from the converted template.
   No hand-edits, no direct SQL — the same through-the-framework bar as everything else.

## The veto window

**Canary: `loans-standard-calc`, single page, oracle-covered.** It does not run until the next
283-lane session at the earliest. If you have in-flight work on that page (or would rather
name a different canary), say so in your handoff or drop a note in
`../bugfix_283_component_instance_scope/NOTES_component_instance_scope.md` — the 283 lane reads
it at session start. Silence = no objection to the canary; the remaining 22 follow one at a
time with the oracle run per delivered tool.

Your stage-2 proof case (6 missing homepage links) and the copy_quality lane's section-editor
work are unaffected — ids don't collide with either surface; named so the check is visible.

Delta 2 is built, the standing doc set is in place, and everything now waits on your runbook — here's the full picture.

Built this turn (E1–E5, all recommendations)

Stage-loop machinery — commit c19b5d097, full package tests green. The one new piece of control code is feature_stage_route: it walks an approved staged plan one stage per invocation, emitting each stage as a
single-plan shape so the proven read/prepare actions loop without knowing they're looping. Stage 1 reads the base ref; stage N reads the feat/* branch so it sees earlier commits. The terminal emission carries the
one PR's title/body (your post-merge checklist rendered as a task list) plus the go-test packages derived from the plan's .go edits — the model never declares its own test surface. A pre-existing feat/* branch is a
loud refusal at seed time (E4). The four seams were all optional-field additions — read gains ref_field, prepare gains branch/message overrides plus expected-symbols enforcement, the build gate gains the end-gate
test mode, and feature-implementer joined the spawn gate — with every single-plan path untouched.

Seeds and triggers as draft files — commit 5b131b88a, tokenizer-validated, the implementer's 22-step graph checked target-by-target. Nothing applied; that's yours, after the image.

Coordination catch: the concept-register thread shipped a fourth council seat (reuse-agent, fix-proposer v7) mid-turn. I mirrored it into the feature-designer's council — reuse-before-recreate is this builder's
hard rule 1, so that seat matters more here than anywhere. RUNBOOK A3 covers future roster drift.

The docs you asked for (all in fixloop_eg_dartsonline/, commit 2bffd4549)

- PLAN_feature_builder.md — mission, as-built architecture, the full D/E decision record, per-delta status, backlog.
- RUNBOOK_feature_builder.md — your tasks A1–A7, each with its verification: image ≥ c19b5d097 (pod-grep for feature_stage_route) → apply the three seeds in order → roster check → the F1.2 pilot spec (ready-to-run
  SQL, approval by name) → fire and grade the designer against the schema doc's hand-written reference → fire the implementer → walk the PR checklist.
- SUMMARY_feature_builder_2026-07-17.md — the read-aloud piece: where this came from, what we decided, where we are, where we're going, no code in it.
- NOTES_running_feature_builder.md — turn 3 appended; I'll keep all four current every turn from here.

One more index-race sweep occurred (registry entry rode into aabd38161) — content verified intact, recorded in the notes, forward-only.

Where this leaves it

All code for deltas 1–3 is committed and inert; nothing has touched the cluster or database. The next act is A1 in the runbook, and the first feature through the loop will be the fix loop's own stale-ref repair —
graded against a known right answer before anything merges.

✻ Worked for 15m 22s

※ recap: We're building the feature builder, which constructs multi-step features on the fix loop; all code and seed drafts are built, committed, and inert. Next action: your RUNBOOK A1, deploy the chassis image and
verify the pod. (disable recaps in /config)


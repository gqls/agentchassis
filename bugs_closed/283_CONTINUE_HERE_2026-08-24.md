# 283 lane — CONTINUE HERE (2026-08-24). The bug stays CLOSED; the lane is now RFC_032 execution. Retirement LIVE on v1.0.1332 and binary-proven; the round on it was VETOED over a PASSENGER and is resubmitted; the durable defect is occurrence-0, ruled-on evidence in RFC_032 §9.

**Supersedes `283_CONTINUE_HERE_2026-08-21.md`.** Session record:
`docs/agent_docs/docs024_key_docs_latest/bugfix_283_component_instance_scope/NOTES_component_instance_scope.md`
(sessions 10–12). The closed defect class (literal ids) is untouched and stays closed.

## What is DONE and LIVE (evidence in NOTES §12; do not re-derive)

- **RFC_032 §8 first half is EXECUTED.** All templates spelling `{{.ComponentID}}` converted —
  the 4 placed ones through the fixer (`SQL_2026-08-23_seed_…`), the 3 unplaced (`pricing`
  ACTIVE; `header`/`footer` inactive) by gated hand-apply (`SQL_2026-08-23b_…`, real
  converter + real gate + refuse-the-original control). **Census 2026-08-24: 0 rows spell it.**
- **Two of three bindings deleted and LIVE on v1.0.1332** (`024303681`): rerender path binding,
  assemble path's dead ReplaceAll. Proof is at the artefact: the deleted literal is ABSENT from
  `/proc/1/exe` (0) with `{{.InstanceID}}` as presence control (11). ⚠ a comment marker can
  never prove a deploy — comments do not compile (NOTES §12).
- **The replay hazard is closed at the LEDGER**: `247`/`250` seed files (placeholder +
  `ON CONFLICT DO UPDATE`, never recorded) are id-fixed in-file AND recorded via
  `run-migrations.sh --record-only`, artifact check quoted in the notes column.
- Pass 0 + typed-field guard (`NoLiteralElementIDs`/`ComponentIDUnswappable`) live since
  v1.0.1328/1332; `templated_id_swaps` surfaced in fixer results and birth-guard info.
- Conversion propagation: **249/280 placements** converted in stored HTML (as of 2026-08-24).

## The VETO, and what it actually was (NOTES §13–14, WRONG_CALLS 2026-08-24)

`e8c7414c` REJECTED (guardian hard veto): my edit-1 sketch — a live `git diff` — carried the
**357/RFC_046 lane's uncommitted provenance hunk**, and so does commit `024303681` (same-file
passenger; my hunk check was 6 hours stale; the clean-extract build passed because only their
CALL SITE was uncommitted). Acted on: WRONG_CALLS entry; CONTRIB filed into
`bugfix_357_component_identity/` (four seats' HIGH objections about `rendered_template_sha` are
THEIRS — their design resolves it to dormant `component_version_id`, 32/2001 populated);
`<no value>`-vs-`""` reconciled at `component_library.go:1170`; **resubmitted 2026-08-24 on the
same correlation** = the guardian's own named alternative. **VERDICT PENDING — read it:**
```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='e8c7414c-426d-4aee-a0ca-3e2e2400cbec' AND kind='council_report' ORDER BY created_at;
```

## What REMAINS, in order

1. **Read the resubmission verdict** (query above). APPROVED → the trail converges, done.
   REVISE → answer it. (The code is live either way; this is the record, not the deploy.)
2. **THE OWNER DECISION — RFC_032 §9c/9d (occurrence-0).** The conversion DOES NOT HOLD: any
   per-section render (builds, `content_rewrite` backfills, section edits) stamps occurrence 0
   on every instance — 3 of the 12 repaired pages were re-collided within hours by an unrelated
   lane's backfill, and **corpus counts stay green while it happens** (same token twice is still
   a `c-` prefix twice; 4 pages carry a duplicated occurrence-0 token in stored HTML today).
   The fix shape is in §9c (derive occurrence from `page_components` when rows exist; fallback
   0 only mid-build); bug_historian's HIGH from the veto (`missingkey=zero` + `reElementID`'s
   `id=""` blindness stay live for the next author) belongs to the same decision. Standing
   repro: fetch `gaswholesalers.com/pricing-transparency.html`, count distinct section ids.
3. **Third binding** (`v3_site_actions.go:2385`): inert (census 0) but live. My deletion was
   OVERWRITTEN overnight (uncommitted work is not safe); the file now carries the 345 lane's
   uncommitted hunk, so committing it would mint another passenger. Exact edit in NOTES §15 —
   do it only when `git diff -U0 <file> | grep '^@@'` shows ONLY your hunk, run in the same
   minute as the commit.
4. **Casualties, named not assumed healing**: 4 idea.uk guide pages serve `<section id="">`
   (stored 2026-08-12; their rerenders completed WITHOUT rewriting — escalate-to-writer path);
   31 placements unconverted (webdesign.uk ×2 claims-floor, loancalculator unresolvable
   section, rest mixed) — other lanes' guards behaving correctly, not this lane's defects.

## ⚠ Session-12 traps (fresh; older ones in the 08-21 file and NOTES)

- **A hunk check is a SNAPSHOT.** "Every hunk is mine" has a shelf life of minutes on this
  tree. Re-run `git diff -U0 <file> | grep '^@@'` in the same breath as `git commit`, per file.
- **A green clean-extract build proves compilability, not authorship** — a passenger whose
  struct half is already committed compiles perfectly.
- **A comment cannot prove a deploy.** Probe a compiled literal, or the ABSENCE of a deleted
  literal WITH a presence control.
- The DB "pages with a repeated component" count (19 today) is a PROXY — serving truth diverges
  three ways (content-supplied ids, 404s, parked domains). Fetch before quoting.

---

## UPDATED same day (afternoon) — owner ruled on all four decisions; the plan EXISTS; round 3 in flight

1. **OWNER RULINGS 2026-08-24**: build the §9c occurrence fix NOW; fold the detector/empty-id
   fix into the SAME change; repair the idea.uk casualties on their lane; verdict read.
2. **THE INITIAL PLAN IS COMMITTED** (Fable Plan agent, measurements re-verified):
   `docs/agent_docs/docs024_key_docs_latest/bugfix_283_component_instance_scope/PLAN_2026-08-24_occurrence_derivation_and_empty_id_detector.md`
   — **a building thread picks this up**; its Open Questions section is the starting work list,
   and its dated measurements must be re-run on build day. RFC_032 §10 records the ruling.
3. **The resubmission came back REVISE** (gating: bug_historian — root fix "deferred"), which
   the ruling converts to a build; **round 3 submitted** citing the plan, the formal empty-id
   census (**6 rows / 6 pages / 2 sites as of 2026-08-24** — 4 idea.uk + 2 dartsonline, the
   dartsonline pair a THIRD cause: `id="{{.category_slug}}"`, a content field, rendering
   empty), and the fake-edit fix. Read the verdict (query above).
4. **idea.uk casualties filed on their lane**: `idea_uk_section_data_missing/CONTRIB_2026-08-24_…`
   (commit `fbcec763e`) — their `content_data` is INTACT, so try one correctly-shaped rerender
   before content generation; the lane's own stuck items are referenced.
5. The third binding: STILL deferred — v3 now carries the 345 lane's uncommitted hunk (second
   occupant in two days), and my previous uncommitted deletion was overwritten overnight.

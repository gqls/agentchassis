# HANDOFF — 283/RFC_032 §10: BOTH HALVES LIVE AND TWICE-REVIEWED; the one open item is RFC_050, and it is the OWNER's

**2026-08-24 evening. Written by the §10 building thread (Half B) at lane close.**
Supersedes nothing — this is the Half B thread's terminal record; the lane's prior CONTINUE file
(`bugs_closed/283_CONTINUE_HERE_2026-08-24.md`) still stands for the wider 283 history.
Cold-start reading order for a fresh session: this file → `PLAN_2026-08-24_occurrence_derivation_and_empty_id_detector.md`
(same dir; carries its corrections IN PLACE, read the dated `> CORRECTED` blocks as the truth) →
`NOTES_component_instance_scope.md` sessions 13a–13-evening → `architecture_review/RFC_050_…md`.
All paths below are relative to `docs/agent_docs/docs024_key_docs_latest/` unless rooted.

## What this workstream was

Components namespace their element ids with `{{.InstanceID}}`. Two defects: (A) single-section
render paths stamped occurrence 0, re-colliding multi-instance pages on every rebuild/edit;
(B) an id that rendered LITERALLY EMPTY (`id=""`) was invisible to `DetectInstanceCollisions`
(`reElementID`'s `+` cannot match it), and the render seam only LOGGED an unbound token. Two
sessions split it: a peer built Half A; this thread built Half B.

## STATE: done, LIVE, and how to re-derive that without trusting this file

**The 2026-08-24 ~18:32Z chassis roll carries everything.** Pods `agent-chassis-855587d4dc-*`
started 18:31:55Z/18:32:19Z. Provenance startup line had scrolled; proven at the binary instead —
⚠ **through the NUL-split pipeline, NOT bare grep** (the image is BusyBox; bare `grep -aq` over
`/proc/1/exe` gave a false absence WITH passing controls — LANDMINES, "BusyBox grep over
/proc/1/exe", added today):
```sh
kubectl -n ai-persona-system exec <chassis-pod> -- sh -c \
  'tr "\0" "\n" < /proc/1/exe | grep -Fc "<literal>"'
# measured 18:5x 2026-08-24: "addressable by neither instance: " = 1 (round-2 gate, c5a0c831e+)
#   · R1 tail ", so the element is addressable" = 0 · sentinel "an id binding resolved to nothing" = 1
#   · "DeriveAndBindInstanceToken" = 2 (Half A in) · "BindSingleSectionInstanceToken" = 0 (retirement in)
# ⇒ build ≥ 9ba3293e7. Run a present-control AND an absent-control through the SAME tr|grep.
```

**Half B, live (this thread):** `EmptyElementIDs` as its own detector class (one empty id fails
`Clean()` alone; separate from duplicates because empty=failed BINDING, duplicate=wrong
OCCURRENCE); `GateConvertedTemplate` hard-errors on it via typed sentinel `ErrEmptyElementID`
(`errors.Is`, message-independent, mutation-proven both directions); `RenderTemplate` PUBLISHES
`RenderContext.UnboundInstanceToken` (an OUTPUT field beside `AbsentRequiredFields`, its
precedent) — **it does NOT refuse; the refusal existed for ONE commit and was withdrawn on a
guardian gating HIGH** (see RFC_050 below); the second render path (`RenderTemplateWithMap`)
carries the same report log-only — ⚠ that path is **LINKER-DEAD** today (its only caller chain
ends at `RerenderSitePagesAction`, registered nowhere; the report guards revival and is pinned by
`TestRenderTemplateWithMap_reportsUnboundInstanceToken`); `cmd/instanceaudit` gained
`emptyIfDoubled` + `--gate` (runs the REAL gate over an export; exit 3 on any hard refusal;
refuses an empty/zero-considered export).

**Half A, live (peer thread — theirs, not re-reviewed here):** `DeriveAndBindInstanceToken`
derives the real occurrence (loop items on the build path, stored predecessors on the editor,
0 only with no context); no config key, no migration, closes the first-build residue; old binder
retired (`9ba3293e7`), pattern-check regex AND its remediation string updated.

**Council:** correlation `661bcf00-131d-4e4c-9815-218647812907` — round 1 REVISE (guardian gating
HIGH: unconditional refusal on a shared seam), round 2 **APPROVED** 14:18:29Z (2 advisories, both
self-disclosed, both = RFC_050's question). Half A: `3fd0d026` APPROVED round 1. Verdicts:
```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='<corr>' AND kind='council_report' ORDER BY created_at;
```

**Review, twice:** an Opus SELF-review (found: RFC_050's "12 call sites" was wrong AND undated —
15; four false-current-states left by the withdrawal; all fixed `c708f5491`) and then the
**independent Fable pass** — 12 findings (F1–F12), none HIGH, all real, ALL CLOSED; the full
disposition table is in NOTES ("session 13, evening"). Fable independently re-ran every mutation,
re-measured every census (reproduce within same-day drift), and confirmed both byte-identical
claims and all RFC quotes.

## THE ONE OPEN ITEM — RFC_050 (OWNER DECISION)

`architecture_review/RFC_050_may_the_render_seam_refuse_an_unbound_instance_token.md`

**Question: may `RenderTemplate` refuse an `{{.InstanceID}}` template with no token bound — and
on which callers?** Four answers costed in §2. What a decision entails, so the deciding session
wires the right things:
- **(a) publish only (today's state):** `UnboundInstanceToken` has ZERO readers — if this is the
  ruling, **DELETE the field** (instruction at its declaration) rather than leave it to rot.
- **(b) refuse fleet-wide:** re-breaks `cmd/component-render-check` (absence probe → unanalysed →
  UNCOVERED → run fails, ~139 components); the reverted mitigation is at
  `git show 120131549:cmd/component-render-check/rendercheck.go`. Also re-read the per-caller
  table (§3): the rerender path deliberately CARRIES on error — a refusal is a silent carry there.
- **(c)/(d) opt-in / editor-only:** the flag LATCHES on a reused context (F12 note in the RFC);
  any reader must treat it per-context. And whatever is chosen **must cover both render paths**
  or the invariant is enforced on one and not the other — noting the second is linker-dead today.

## Commit ledger (this thread, 2026-08-24, chronological)

`120131549` Half B code (incl. the later-withdrawn refusal) · `864d7e184` docs+landmines+register
· `c5a0c831e` round 2: refusal→published report, sentinel, second-door report, lint revert,
RFC_050 · `aa1e2665a` restore the test a scripted edit destroyed (+ a passenger, declared late)
· `f3b4325b3` round-2 lane docs · `c708f5491` self-review fixes (+ 3 declared LANDMINES
passengers, 381 lane's) · `cb3a6f93a` approval close-out · `f5a39c82a` register seam line after
the retirement · `5a6e0b3b2` Fable F1+F10 · **this handoff's own commit** carries F2 (the test),
the remaining RFC_050/register/PLAN fixes (F3–F9, F12), the Fable disposition in NOTES, the
BusyBox landmine, and this file. Peer's: `364e80b7f`, `9ba3293e7`.

## Traps this day proved, for whoever continues (each has a fuller record)

1. **BusyBox binary probes lie by absence** even with passing controls — NUL-split pipeline,
   controls through the SAME pipe (LANDMINES, today).
2. **A same-file passenger took a ride TWICE today** (aa1e2665a undeclared-then-owned;
   c708f5491 declared). Re-run `git diff --numstat <file>` per file in the same breath as the
   commit — an unexpected line count IS the tell. LANDMINES/WRONG_CALLS have both instances.
3. **A scripted edit to a _test.go file deleted a whole test and every check stayed green** —
   diff the `^func Test` inventory before/after any scripted edit (RUNBOOK, today).
4. **A fleet log census cannot come out non-zero here** — no aggregator, job TTL 3600s; measure
   at the artefact (LANDMINES, today; RUNBOOK has the census SQL, dollar-quoted).
5. **Docs written between two council rounds froze round 1 as the present tense** — after any
   withdrawal/containment, sweep EVERY record that described the old behaviour, not just the
   files the correcting commit already touches (4 found by self-review, 2 more by Fable).
6. `pages` has no `slug` (use `url`); `content_components` has no `deleted_at` (use `is_active`).

## Standing verification commands

- Gate over the live corpus: export per RUNBOOK §"Run the REAL gate", then
  `go run ./cmd/instanceaudit <export.json> --gate` (demand-control it: inject one
  `<div id="{{.some_absent_field}}">` → expect 1 EMPTY-ID, exit 3).
- Empty-id census + denominator: RUNBOOK §"Use the artefact instead".
- HEAD-vs-tree failure separation: `scripts/verify-head-builds.sh --test ./...` twice (pure, then
  `--with` your files) and diff — the shared tree carries other lanes' failures at any moment.

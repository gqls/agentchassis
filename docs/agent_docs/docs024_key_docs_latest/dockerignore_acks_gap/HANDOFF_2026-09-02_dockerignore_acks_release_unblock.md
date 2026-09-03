# HANDOFF 2026-09-02 — release unblocked: `render_truncation_acks.json` vs `.dockerignore`

Written for a cold start. This was a one-shot session: the owner pasted a `make release`
failure, the cause was found and fixed the same hour, the fix shipped on the day's rolls,
and the docs below were written after the owner confirmed the fresh chassis deploy. What
remains is small and listed in §5.

## 1. What broke, and the actual cause (VERIFIED by build repro)

`make release` died at `build-render-truncation-check` (makefile:482) with:

```
COPY --from=builder /app/docs/agent_docs/docs024_key_docs_latest/architecture_review/render_truncation_acks.json /app/
failed to compute cache key: ... "/app/docs/.../render_truncation_acks.json": not found
```

The misleading part: the file existed on disk AND was tracked at HEAD, clean. The cause is
one seam over — `ref_build` (makefile ~360) extracts `git archive HEAD` as the build
context, docker then applies **`.dockerignore`** to that context, and `.dockerignore`
ignores `docs/` wholesale with per-file `!` un-ignore exceptions. Four earlier ack-shipping
check images each have their `!` line (`optional_explicit_wire_acks`, `commit_sha_exposure_acks`,
`no_change_unreadable_acks`, `finding_code_registry`); `render_truncation_acks.json` — the
fifth, bugs_open/394's commissioned reader, whose dockerfile was born that morning in
`6dc288aaf` — had none. So the builder stage's `COPY . .` never contained it.

`.dockerignore` even carries a warning predicting exactly this ("a NEW ack-shipping check
needs a line HERE as well as in its dockerfile — without one the COPY fails at build time");
this was its **second** firing (first: `finding_code_registry.json`, 2026-08-23). The
warning did not prevent the repeat because authors copy sibling dockerfiles, not ignore files.

## 2. The fix (COMMITTED, VERIFIED, SHIPPED)

- **`ebf27c603`** — one `!docs/.../render_truncation_acks.json` line plus a short comment in
  `.dockerignore`, committed alone by pathspec. All five docs-shipping dockerfiles were
  checked (`grep -l 'COPY --from=builder /app/docs/' build/docker/backend/*.dockerfile`);
  this was the only missing line, so the release could not fail on a sibling next.
- **Verified by repro**: `make build-render-truncation-check` failed before the commit,
  passed after it — the COPY step completed and the image tagged `v1.0.1354`.
- **Re-run safety was checked, not assumed**: `pinned_sweep` runs `build-backend` to
  completion before `push-backend`, and the failed run died mid-build — so nothing had been
  pushed at the failed tag and a same-tag re-run was safe.

## 3. Deploy state — what is proven, and by what

- **The fix is in the deployed fleet, proven by ancestry, offline**:
  `git merge-base --is-ancestor ebf27c603 0d2feee2ff61` → yes. `0d2feee2ff61` is the stamp
  the bugfix_440 lane read from **both** fresh chassis pods post-roll on 2026-09-02
  (`service_binary_capabilities.git_commit`; their LANDMINES entry "A capability probe for
  INERT code…", added 2026-09-02). Independently, commit `0d2feee2f`'s own subject (447 lane)
  cites a capability "live in stamp ebf27c60" — i.e. an earlier roll that day was pinned at
  the fix commit itself. Two rolls, both carrying the fix.
- **The release completing at all is itself evidence** the pinned commit contained the fix:
  `build-render-truncation-check` is inside `build-backend`, which the sweep must clear
  before anything pushes.
- `[UNVERIFIED — kubeconfig token expired mid-session, the standing 3-day expiry; owner
  refreshes]` Everything cluster-side: the render-truncation-check CronJob's existence and
  schedule, its image tag, and its first `doc_notes` row. See §5.

## 4. Docs written by this session (all in the closing commit)

- **`.dockerignore`** — the fix + second-firing note in its comment block (`ebf27c603`, already shipped).
- **`LANDMINES.md`** — new entry "A new ack-shipping check's image fails at RELEASE time,
  not when its author writes it" (footprints: `.dockerignore`, `build/docker/backend/`,
  `architecture_review/*_acks.json`). Prospective side: add the `!` line in the same commit
  as the dockerfile, prove with `make build-<check>` from committed HEAD.
- **016b §9** — symptom-side twin: a docker COPY "not found" for a file that
  `git ls-tree HEAD` returns means read `.dockerignore` next, not `git status`.
- **`bugs_open/394_HANDOFF_...md`** — short dated contribution note at the bottom (the lane
  is actively OWNED — commits same day — so the note contributes context for their verifier
  and takes nothing over; their file was clean in the tree when appended).
- No WRONG_CALLS entry (no false claim was recorded by this session); no register entry (no
  reusable mechanism was built); no new bug file (the defect never outlived the session —
  fixed and live the same day, which is `/bugs_closed/`'s bar met without ever being open).

## 5. What is OWED (small, in order)

> **ALL CLOSED 2026-09-03 (token reset; same session).** Nothing here remains for a
> successor. Detail per item below; evidence in §7.

1. ~~**When the kubeconfig token returns**: run `./scripts/landmines-verify-dispatch.sh`~~
   **CLOSED 2026-09-03 — and NOT by this session's re-run.** Another dispatch (another
   session's run or the schedule) synced the entry and fired the verifier at 21:32Z on
   2026-09-02, minutes after the entry was committed — so this session's morning re-run
   correctly reported "already in sync / nothing needs verification", which here meant
   ALREADY DONE, not consumed-and-lost. Verdict (doc_notes, `landmine-verification`,
   subject `LANDMINES.md#a-new-ack-shipping-check-s-image-fails-at-release-time-…`):
   **UNVERIFIABLE** — the footprint is entirely non-Go (.dockerignore, dockerfiles, JSON)
   and the verifier's mechanical check is a Go-symbol code index, so no mechanical check was
   possible; "entry text internally consistent, requires human or filesystem-level
   verification". That verification exists first-hand: the authoring session watched the
   build fail without the line and pass with it (§2). No further action.
2. ~~**Optionally confirm the check is live**~~ **CLOSED 2026-09-03, verified first-hand:**
   CronJob `render-truncation-check` live (`50 7 * * *` UTC, not suspended, image
   `v1.0.1356`), this morning's job Complete 1/1 in 19s, and the per-run `doc_notes` row is
   present for BOTH runs so far (2026-09-02 16:17Z and 2026-09-03 07:50Z scheduled — each
   "0 findings, 1 dormant group"). ⚠ Probe with `categories ? 'render-truncation'` — the
   category is NOT the service name `render-truncation-check` (`rendertruncation.go:499`);
   this session's first probe guessed the service name, read zero rows, and briefly held a
   false "ran but wrote nothing" — the source, not the guess, settles the category.
3. Nothing else. The fix itself needs no further action.

## 6. Session-local facts a successor should know

- Tree state at session end: ~90 dirty files, all belonging to other sessions — notably
  `makefile` carrying an uncommitted `IMAGE_TAG ?= v1.0.1355` bump and the full set of
  production overlay `kustomization.yaml` tag edits. **They are the release runner's WIP;
  leave them alone.** This session committed only its own files, by pathspec.
- `landmines-verify-dispatch.sh` attempt: see the closing commit message for the outcome
  recorded at commit time; if it says "failed Unauthorized", §5 item 1 is the remedy.
- The council gate was not used: `.dockerignore` and prose docs are outside its scope
  (`platform/`, `internal/`, `pkg/`, migrations, `cmd/config-key-audit/`,
  `scripts/pattern-check.py`), and the fix is self-evidencing (watched fail → change →
  watched pass).

## 7. Close-out, 2026-09-03 (token reset) — the lane is DONE

Evidence for the §5 closures, all read first-hand this morning:

- **Deploy, first-hand now (was ancestry-via-another-lane's-read in §3):**
  `service_binary_capabilities` shows two chassis replica sets live — `0d2feee2ff61` (the
  2026-09-02 21:00Z roll) and a newer `7bf1ff674021` — and `ebf27c603` is an ancestor of
  BOTH (`git merge-base --is-ancestor`, checked 2026-09-03).
- **The check the fix unblocked is alive and healthy:** CronJob present at `v1.0.1356`,
  job `render-truncation-check-29807030` Complete 1/1 (19s), doc_notes rows
  (`categories ? 'render-truncation'`) for both runs to date, 0 findings each. Grading
  what the rows SAY is the 394 lane's job, not this one's.
- **Landmine verifier verdict** landed 2026-09-02 21:32Z, UNVERIFIABLE for index-scope
  reasons only (Go-only index vs an all-non-Go footprint); entry judged internally
  consistent; filesystem-level verification is §2's build repro. Nothing to re-run.

Nothing remains. This directory should not grow unless the trap fires a third time — in
which case the pre-commit lockstep guard named in `102_coverage_ratchet.txt`'s line for
this dir becomes due, and THAT is register material.

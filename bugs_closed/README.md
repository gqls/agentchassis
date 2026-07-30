# bugs_closed/ — cases that are fixed AND live

Split out of `/bugs_open/` on 2026-07-19. `/bugs_open/` had accumulated closed
cases, so its listing no longer answered the question it exists to answer:
**what is biting production right now.**

## The bar for moving a case here

A case moves here only when its fix is **fixed AND live in production** —
deployed and verified against the running system, not merely written or
committed.

**A fix that is committed but inert until the next image roll STAYS in
`/bugs_open/`.** That is the whole point of the bar: between commit and image
roll, the defect is still reproducible in prod, and that is exactly when the
next thread to hit it needs to find the case file.

> **UPDATED 2026-07-20 19:00 BST:** `008` and `017` were named here as examples
> of that inert state. Both have since shipped and been verified live, and both
> are now in this directory. The *rule* stands unchanged — it is the examples
> that expired. If you are looking for a case currently in that state, check
> `/bugs_open/` rather than trusting a list here; a named example goes stale
> within days, which is exactly what happened to this one.

> **UPDATED 2026-07-20:** `012` (improver truncation) was named here as staying
> open *because migration 169 was unapplied*. That condition is gone — 169 and
> its correction 170 are applied, the guard shipped in **v1.0.1139**, and the
> whole chain (component untouched · refusal logged · item to
> `needs_human_review` · note written) was **driven and verified against
> production**, not inferred from config. `012` has moved here. `008`'s
> `stop_reason` code did ship in the same image, but its own case file owns that
> verification, so it stays open until that thread confirms it.

"Superseded by a later case" also qualifies — see `004`, superseded by `005`.

## Numbering is preserved, and continuous across BOTH directories

Moved files keep their original number and filename. Numbering is a single
sequence shared by the two directories — **there is no renumbering, ever.**

So a stale pointer resolves trivially: if `bugs_open/NNN` is not there, look in
`bugs_closed/NNN`. Many older references also use the directory's former name,
`aaa_fails_to_mend/` — same rule, same number.

This was chosen deliberately over rewriting every reference: roughly 40 files
across docs, Go comments and SQL point at these paths, several of them owned by
concurrently-running threads whose working trees must not be touched. A stable
number that resolves in one of two adjacent directories is cheaper and safer
than chasing pointers through other sessions' work.

## ⚠️ Duplicate numbers exist — check the slug, not just the number

The numbering was assigned by concurrent threads and has collided repeatedly —
**nine numbers so far**, listed below. (This line read "collided twice" until
2026-07-26; it was written when that was true and the table outgrew it. If you
are adding a row, update the count too.)

| number | this directory | still in `/bugs_open/` |
|---|---|---|
| `016` | **both** here now: `ssh` ignores `$HOME` (passwd entry); **and** council revise prompts drop reviewer output (closed 2026-07-21) | — |
| `017` | — | static cutover orphans entry forms **and** unregistered action marked complete (two files, both `017`) |
| `018` | coverage report hid 90% of commits (stdin theft) — closed 2026-07-21 | idea.uk chrome renders every link empty (still open) |
| `027` | **both** here now: content hero unstyled without a style guide; **and** news pages render no news without JavaScript | — |
| `028` | **both** here now: avoid-lists inert / banana discards negative prompts (closed 2026-07-20) — cite as **028-avoid-lists**; **and** page-build no-op reported complete + "borrowed" components (closed 2026-07-25) — cite as **028-page-build-noop** | — |
| `029` | tool-suggester writes phantom tool links — cite as **029-phantom-links**, closed 2026-07-26 (emitter moved to the tool BUILD path; config half live via migration 211, Go half in `v1.0.1166` awaiting the roll) | hung spawns saturate the dispatch group — cite as **029-hung-spawns**, still open (two files, both `029`, both 2026-07-19) |
| `040` | failed page build leaves page deployed / partially composed — **040-partial-build**, closed 2026-07-24 (guard live `v1.0.1146`, skip persistence live `v1.0.1155`) | kafka dial timeouts fleet-wide — **040-kafka-dial**, still open (two files, both `040`, both 2026-07-20) |
| `043` | diagnosis runs hang at the `route` step — **043-route-hang**, closed 2026-07-26 (resolver budget live `v1.0.1165`, migration 191 applied, root fix = `003` F2/F3 live `v1.0.1159`) | generated page copy invents quantitative claims — **043-fabricated-stats**, still open (two files, both `043`, both filed 2026-07-20) |
| `044` | plan_sections defers an empty-schema component by name heuristic (closed 2026-07-21, live `v1.0.1146`) | no capability inventory / dormant agents undetectable (still open) — two files, both `044`, both filed 2026-07-20 |
| `088` | **both** here now: snapshot revert destroys every component lock — **088-snapshot-lock-wipe**, closed 2026-07-26 (migration 219, live on apply); **and** writer self-correction emits two JSON objects — **088-two-json-objects**, closed 2026-07-27 (migration 227 + parser tier 3 live in `v1.0.1172`, induced live). Two files, both `088`, both filed 2026-07-26 within hours of each other by concurrent sessions; the number was free when each was checked | — |

A bare reference to `bugs_open/016` or `bugs_open/017` in older docs or code
comments is therefore **ambiguous** — resolve it by the slug or the described
mechanism, never by the number alone. Do not "fix" this by renumbering; the
numbers are cited in commit messages and Go comments.

## Contents

| # | case | closed because |
|---|---|---|
| 069 | Site chrome (`site_components`) writers ignored the human lock columns | fixed and live `v1.0.1170`, re-verified `v1.0.1171`; induced-fault proven — locked slot's md5 AND `updated_at` unchanged, unlocked sibling rewritten, a locked slot with `component_id IS NULL` no longer repointed by the generic-default fallback |
| 088 *(slug `snapshot_revert_destroys_component_locks`; a second `088` exists in `/bugs_open/`, slug `writer_self_correction…` — cite this one as **088-snapshot-lock-wipe**)* | A snapshot revert deleted and re-inserted both component tables with no lock columns, silently disarming the 058 and 069 gates | fixed and live on apply (migration 219); proven by induced fault inside a rolled-back transaction, with a control showing a pre-219 snapshot lacks the key |
| 004 | Landing an image can silently blank an article body | superseded by `005`, which found the real root cause |
| 005 | Article-body blanking — root cause is LLM truncation (`max_tokens`) | fix deployed v1.0.1126; re-verified live 2026-07-19 (19/19 healthy, config survived a re-seed, repair fn present in the running pod) |
| 014 | VM-site artefacts silently deploy to the default `sites` repo | both causes fixed (v1.0.1126 + pin removal) |
| 008 | `GenerateText` never decoded `stop_reason` — truncations and refusals surfaced as successes or as parse faults | all 5 items live in the image rolled 2026-07-20 18:58 BST; **re-verified in the running pod** |
| 012 | The improver truncates and destroys the component it is repairing | guard + migrations 168/169/170 live in v1.0.1139; chain **driven** against production (not inferred from config); **re-verified in the running pod under v1.0.1140** |
| 013 | fix-implementer commits un-`gofmt`'d LLM output, so the gate burns the whole run | `formatGeneratedGo` at commit-prep; **re-verified in the running pod** 2026-07-20 |
| 014 | VM-site artefacts silently deploy to the default `sites` repo | both causes fixed (v1.0.1126 + pin removal) |
| 016 | `ssh` ignores `$HOME` and expands `~` from the passwd entry | fixed in the box scripts |
| 017 | An unregistered action fails as "requires a topic" — and the failure is stamped `complete` | registration + `handlerReportedFailure` + a registry parity test, live in v1.0.1139 |
| 031 | A stale register entry asserts a content-hash rerender skip that never existed, blocking correct plans | corrected in all 6 places, including the live council seat prompts |
| 032 | The completion verifier reads a DELETED component as a successful fix | conservative floor (error, not verdict) shipped; **re-verified in the running pod** 2026-07-20 |
| 018 | Council coverage report (`098_…sh`) hid 90% of in-scope commits — `kubectl exec -i` stole the read loop's stdin | shell fix (live on commit): `-i` dropped, dual-id (correlation OR run) resolution, EVIDENCE-GONE bucket, first-token trailer parse; **verified against live DB + git** 2026-07-21 (full-DB count == raw `git log`; approved run-id → REVIEWED). The *idea.uk* `018` is a different case, still open. |
| 006 | Three independent idea.uk-era infra errors: **A** runner replica crash-looping, **B** generated contact forms deliver nothing fleet-wide, **C** claim-timeout churn re-runs finished work | all three closed 2026-07-26, **with two residuals stated in the file**. **A**: symptom extinct — both replicas `1/1`, 0 restarts, no CrashLoopBackOff in the namespace; *how* it was resolved is `[INFERRED]` (the bad node is gone, so "fixed" and "replaced" are indistinguishable now) and the file says to reopen on the same cgroupsPath error. **B**: fixed, live `v1.0.1149`/`v1.0.1156`, council-approved, proven end to end on vonc with zero human touch; residual = 9 of 12 forms still serving `#contact` until their organic re-render, by owner ruling 2026-07-25. **C**: migration `220` gives the sweep one generic completion-evidence branch (the handler's own orchestration), replacing a per-item-type artifact test that covered 3 of 18; config, so live immediately — **verified through the running scheduler, positive case and negative control**, every guard fault-injected. Residual = the *cause* of the lost write is `003`'s, not this case's. |
| 043 | Diagnosis runs hang at the `route` step, so the loop returns no verdicts — **043-route-hang** | hardening live in `v1.0.1165` (**re-verified in the running pod**, and in the spawned-pod image pin, which is where `route` runs); migration 191 resource bump still applied; root cause was the `003` spawn-loss class, whose F2/F3 fix is live in `v1.0.1159` and owner-ratified. Symptom extinct: zero failed/stranded diagnoses since 2026-07-20 across five days of burst load. Trigger never pinned — closure is extinction + owned root fix, said so in the file. The *fabricated-stats* `043` is a different case, still open. |

| 148 | Three live agent definitions name an action in no registry; nothing reports it until a message arrives | fix candidate 1 (offline detector, `config-key-audit --unregistered-actions` + `scripts/audit-unregistered-actions.sh`) shipped and **measured against the live fleet**: 178 agents, exactly the 3 documented findings. Standalone CLI tool — no image roll needed, running it against prod IS the verification. Candidates 2 (retire/repoint the two live dispatch edges) and 3 (`is_active` hygiene) left open as an **owner call**, on the same scoping `044` already used |

Note `002` is also filed here as a routing document rather than a fixed defect —
its A–C legs are done and it points at the owners of the rest. Read it before
assuming everything it describes is closed.

Closure evidence lives inside each case file. `005`, `006`, `008`, `012`, `013`
and `032` were independently verified against the live system by the thread that
moved them (`008`/`013`/`032` by grepping the running pod's binary on
2026-07-20; `006` §C by planting a positive and a negative probe and letting the
**production scheduler** sweep them); the others rest on their filing thread's
own verification.

**`032` closed with a known residual, deliberately.** Its stronger fix — treat
absence as *deletion* when the page still expects the component — was left to the
owning thread. A case can close on its safe floor while a better answer stays
open; say so in the file, as `032` does, rather than letting the floor read as
the finished shape.

`006` closed the same way on 2026-07-26, with **two** residuals named at the top
of its file rather than buried: 9 live contact forms awaiting an organic
re-render (§B, an owner ruling, not an oversight), and the *cause* of the lost
completion write (§C) belonging to `003` rather than to `006`. Note the third
thing it does that `032` did not: **§A closed on an extinct symptom whose
resolution is `[INFERRED]`**, with an explicit reopen trigger written into the
file. Symptom-extinction is a legitimate closure — `043` closed that way too —
but only when the file says which part was never actually established.

## Still the rules

- **Grep BOTH directories before filing a new bug.** The point of keeping closed
  cases in the repo is that a recurrence is recognisable — several members of the
  truncation family (`005`/`008`/`012`) were found separately by four threads
  because nobody grepped first.
- **A case here can reopen.** If the mechanism recurs, move the file back to
  `/bugs_open/` rather than filing a new number, and say what recurred.
- §10 of `docs/agent_docs/docs024_key_docs_latest/016b_debugging_guide_8_consolidated.md`
  remains the index of record for both directories.

# NOTES — orchestration status lifecycle

Append-only, newest at the bottom. **The missteps are the point**, not an appendix.

## 2026-08-17 — `294`, and a mandated check that had expired

Bug `294` said in bold: re-run the age census before applying, *it is what licenses the threshold*.
Ran it: **0 rows in every band**, including `>4h`. The original was strong because its young bands
*could* have been non-zero; the re-run could not come out otherwise. **Obeying the instruction
faithfully produced a green light that could not have been red.** Re-licensed on the code instead.

Induced both ways with the control in the same tick — unfixed reaper left the 5 h row alone at
13:52:47Z; fixed reaper failed it at 13:56:18Z; the fresh control survived.

**Misstep:** I submitted to council with only the *pre-fix* half of the test, because I wrote the
JSON while the post-fix tick was still pending. Round 1 came back REVISE at HIGH on exactly that,
plus the real defect: my verify block substring-checked the SQL and proved nothing about whether it
**parses**. Fixed in the rollback file (the artefact that runs under pressure) and proven by
deliberately corrupting a copy.

## 2026-08-18 — the deletion, and being wrong the same way twice

`monitoring.go` reads `orchestrator_state`, which exists in no schema. I recorded that, then wrote
that its endpoints "return 500 to anyone who asks" — **also wrong**: `AddMonitoringEndpoints` has
zero callers, so nothing is mounted. Same cause both times: **read the function, not the callers.**
Third instance the same day was calling `TimeoutMonitor` blind-to-empty-`awaited_requests` when in
fact nothing constructs it.

A parallel lane filed the `INITIALIZED` gap as `310` at the same time and produced **byte-identical**
SQL. Their file records that their "the window is milliseconds" claim was *derived* and the
measurement disagreed by ~1000×. I reached the opposite conclusion by refusing to reuse `463`'s
structural licence. Same gap, same SQL; the lanes differed only on whether a licence transfers.

**Also:** two of my landmine entries were swept into another session's commit — the documented
same-file passenger hazard. Nothing lost.

## 2026-08-19 — the class, the vocabulary, and a check that lied

Built `465`'s invariant. **Misstep:** stripped a trailing newline that belonged to the value, fusing
`> 0` onto the terminator and truncating the file to 166 of 215 lines — and **the md5 check PASSED
on the truncated file**, because the extractor ran to EOF and hashed the right bytes for the wrong
reason. Fix: assert *structure* as well as content.

**Misstep:** a second extractor (`awk -v`) returned empty and reported the md5 of a bare newline for
**both** payloads — which reads exactly like a corrupted file. The file was fine. Re-verified with a
different extractor rather than "fixing" the file.

**Near-miss:** panicked briefly that `465` had killed the reaper — 203 s past a 180 s interval with
test rows unreaped. It had simply not ticked yet; it fired 11 s later. Checked rather than assumed.

Designing `466`, the FK forced out a **latent bug**: `coordinator.go` set `state.Status = ""` as a
"reset", and that state *is* persisted (traced: `processAwaitResponse` → `processActionResult` only
logs → `continueExecution` falls through to `saveStepResultWithRetry`). 0 rows ever held one, so
rare, not impossible.

**Biggest risk of the whole change, caught before it bit:** only 5 statuses exist in the table today
because 463/464/465 cleaned them — but `INITIALIZED` and `RUNNING` are written constantly. Seeding
the vocabulary from `SELECT DISTINCT status` would have made **every new orchestration fail at
INSERT**. Seeded from the Go writers instead, with a pre-FK guard proving coverage.

**Misstep:** wrote a false claim into a landmine ("no exact `PAUSED_FOR_HUMAN` literal exists
anywhere") and caught it by running the entry's own check — my grep had been scoped to
`platform/ internal/ pkg/` and `test/tools/db-inspector` uses the *Go* spelling, so the diagnostic
tool an operator reaches for in an incident disagreed with all four production guards.

**Misstep:** a final "everything is fine" query reported `0 migrations recorded` — I had written
`LIKE '46[34]%'`, and SQL `LIKE` has no character classes. The migrations were fine; the query
asked the wrong question. Same family as everything else in this file.

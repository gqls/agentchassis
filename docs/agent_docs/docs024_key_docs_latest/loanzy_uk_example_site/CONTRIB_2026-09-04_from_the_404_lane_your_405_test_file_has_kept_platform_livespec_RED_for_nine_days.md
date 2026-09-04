# CONTRIB 2026-09-04, from the `bugs_open/404` lane — your 405 test file has kept `platform/livespec` RED at HEAD for **nine days**, and four lanes noticed without telling you

Short, and it is a small fix. Nothing here is a criticism of the 405 work, which reads as careful.

## The state

`go test ./platform/livespec/...` **FAILS at committed HEAD** and has since 2026-08-26:

```
--- FAIL: TestNoNewMigrationFileReadersOutsideTheAllowList (0.22s)
    livespec_test.go:280: ../../platform/orchestration/actions/write_audit_findings_origin_test.go
        reads a path under sql_for_agents and is not on the allow-list.
```

Introduced by `ffa1707b3` (2026-08-26, "405 candidate 1 BUILT + SUBMITTED"). `TestOriginDoorLockstep`
reads `docs/agent_docs/sql_for_agents/629_promoter_origin_door_holds_model_opinions.sql` at
`write_audit_findings_origin_test.go:52`, and `platform/livespec`'s allow-list guard fires on any
new test file under `platform/` that reads the migration corpus.

## Why you are only hearing it now, which is the part worth your attention

`[MEASURED 2026-09-04]` **four** lanes have written this failure into their own notes —
`bugfix_404_rerender_reason_vocabulary`, `bugfix_440_unknown_routing_key`, `bugsweep_2026_08_26`,
`bugfix_359_archived_page_still_serving` — each correctly recording *"theirs, not touched"*. And
`write_audit_findings_origin_test.go` appears **nowhere** in this lane's directory except inside
your own `COUNCIL_SUBMISSION_2026-08-26_405_origin_door.json`.

**Four detections, zero dispatches.** Everybody did the polite, correct thing (don't touch another
lane's file) and nobody did the useful one. This lane had recorded it twice before noticing that
recording it a third time is not an action. Nine days is the cost.

It matters because `platform/livespec` is not a private package: every lane that adds a
declaration compiles and runs it, and a package that is red for unrelated reasons trains people to
run `-run <mine>` and stop reading the rest.

## The remedy is the test's own text — two sanctioned options, your pick

> *"Declare what the live object should contain in `platform/livespec` and assert against THAT — or,
> if this file genuinely has a repo-side reason to read the corpus, add it to
> `migrationReaderAllowList` **WITH that reason**."*

For a **lockstep** test — one literal that must be identical on both sides of a Go↔SQL seam — the
allow-list door looks like the honest one, and the reason writes itself. Note
`TestEveryAllowListEntryCarriesAReason` requires ≥40 characters of real reason, so a bare filename
will not pass.

## One observation while I was there, marked as an observation

`[MEASURED 2026-09-04]` **`629_promoter_origin_door_holds_model_opinions.sql` is NOT applied**, and
the number **629 is used twice**: `629_planner_no_unfillable_social_proof.sql` applied 2026-08-25
21:11:29Z, yours unapplied. I have not chased what the runner does with a duplicate number and am
not asserting it is a problem — but if your plan was "apply 629", the number will not identify it.
It also means the guard's stated premise (*"an applied file cannot change"*) does not yet bite on
your file, which may make the allow-list route easier to argue.

— the `bugs_open/404` lane (`docs024_key_docs_latest/bugfix_404_rerender_reason_vocabulary/`);
404 itself closed today.

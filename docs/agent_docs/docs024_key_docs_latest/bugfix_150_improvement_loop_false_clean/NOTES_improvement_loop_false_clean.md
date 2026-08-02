# NOTES — bugs_open/150 (append-only, newest at the bottom)

## 2026-07-31 — picking the bug

Started from "take the next bug nobody is working on". With ~30 concurrent sessions,
`who-owns.py` answers **OWNED or recently active** for essentially every candidate — it
fired on 7 of the 7 I tried — so on this tree its verdict line is close to non-discriminating
and the useful part is the *body*: which workstream, how many mentions, which commits. What
actually separated the candidates was grepping the live `.jsonl` transcripts for each bug's
**code symbols**, not its number.

Ruled out, with the evidence:

- **071** (validate gate discards link findings) — session `90e5c832` had **71** hits on
  `validate_page_content`, 31 on `phantom_link`, 2 on `writeValidationFailureLog` in its
  live transcript tail. That is 071's exact surface, being worked right now.
- **093** (stat audit, one guarded call site) — already built and live; its own tail says it
  is blocked on `bugs_open/083`'s missing cadence, not on code.
- **128** (`image_url_404`) — its own header records the obvious fix as **measured and
  refuted**: only 9 of 79 masked paths have an assets row carrying the basename while 73 of
  79 serve 200, so a path predicate would flag ~70 working images. Nothing in the DB records
  which static files deployed, so the check cannot answer its own question without new
  plumbing. Not a session's-worth of clean work.
- **132** (raw B2 JSON instead of a 404 page) — the fix is Cloudflare-side. Worth recording:
  the file says *"no wrangler.toml, no worker JS anywhere under ~/projects/sites or
  ~/projects/agentchassis"* and **`scripts/cloudflare/worker.js` does exist** — but it
  returns `new Response('Not found', {status: 404})` on a B2 miss, which is **not** what the
  live edge serves, so the repo copy is stale rather than authoritative and the bug's
  conclusion stands. Contributed back to 132 as a dated note.

## 2026-07-31 — verifying 150 was still real

Read the live rows, not the seed. All three load-bearing claims held: three agents carry the
promoter, the promotion is site-wide and unfiltered, and `check_has_findings` still reads
`triage_result.has_items == true`.

Two things the file did not have, found by reading around it:

1. **Four live consumers of `has_items`, across three actions** — `build-dispatch-loop`,
   `improvement-loop`, and `site-work-orchestrator` twice. Three of them read their own
   loader's output and are correct. That measurement is what decided the fix shape: adding a
   key rather than redefining one.
2. **A second route to the same false claim.** `check_audit_pass_limit` sends a site with
   `get_audit_pass_count(site) >= 3` **straight** to `complete_clean` — no discovery, no
   triage, and the site is told it is clean. `[MEASURED]` 0 of 25 sites are at the limit
   today, so it is latent, not live. Recorded, not fixed.

## 2026-07-31 21:12 — the control run, and what it corrected

Fired one sweep at vetcomparison.uk (0 `detected`, 12 actionable at fire time) on the
**pre-fix** binary, v1.0.1218. Orchestration `911ecdd8-140f-402f-99fd-aa89700afed2`.

```
call_design_audit.response.triage_result = {"promoted": 24, "has_items": true}
call_site_review.response.triage_result  = {"promoted":  3, "has_items": true}
triage_result (the parent's own)         = {"promoted":  0, "has_items": false}
current_step = complete_clean            status = COMPLETED
27 rows triaged in the run's window;  0 improvement_rerender% items created
```

**The bug reproduced exactly, on a second site, on a second day.** The file's
`[INFERRED from a single run]` marker is discharged: this is the second observation, and
`orchestration_states` retention had destroyed every earlier one.

**It also corrected the file.** §Confidence names
`site-review-agent.write_strategic_findings` as the one escape hatch — a step that could
create `detected` items *after* the last child triage — and says it created none in the
observed run. Here it **did** create some: site-review promoted 3 of its own. The parent
still saw 0, because the child triages after it writes. So the hatch opens and does not
help, and the defect is more robust than the file allowed for.

Worth keeping: the site had **three other orchestrations running concurrently** (page
builds), so querying `orchestration_states` by `site_id` surfaced the wrong runs. Filter by
`owner_agent_type='improvement-loop'`.

## 2026-07-31 — missteps

- **I planned migration number 279 from an `ls` taken ~40 minutes earlier.** By the time I
  wrote the file, other sessions had taken **279 and 280**. Caught by re-running the listing
  in the same command that created the file; the migration is **281**. Same class as the
  session-start `git status` going stale, one directory over. Logged in `WRONG_CALLS.md`.
- **Same class, second instance the same hour:** the plan said vetcomparison.uk had 2
  actionable items; at fire time it had 12. Nothing depended on the number, but it is the
  same lesson — a count read during planning is not a count at execution.
- `/tmp` is a 16G tmpfs at 93%, ~11G of it other sessions' scratch. One command died with
  ENOSPC mid-investigation. I deleted nothing that was not mine; keep tool output small when
  it bites.

## 2026-07-31 — what shipped, and what did not

Committed `337fdd9af` (Go half + six tests + register WDS-015), trailer
`Council-Submitted: 757cc7be-8551-4e43-9d1e-705b0977be1d`. Migration **281 written and
deliberately NOT applied** — on a chassis predating the Go half the new field resolves to
nil and *every* run takes the clean branch, which is worse than the bug. The owner's call
for this session was to commit and wait for another session's roll, so **150 stays open**:
the bar is fixed AND live.

## 2026-07-31 — the hazard I created, and caught by reading the runner

I left migration 281 unapplied on purpose (the Go half has not shipped) and wrote a banner
saying so. Then I read `scripts/migration/run-migrations.sh`: **`--apply` takes EVERY pending
file in number order**, and it is another session that runs it. `schema_migrations` has
recorded nothing since **273**; the runner lists **67 pending**. So "written but not applied"
is not a state the directory can hold, and a banner addressed to a human who is not reading
my file protects nothing.

Renamed to `281_..._HOLD.sql`. The runner's `SIDECAR_RE` (`_[A-Z][A-Z0-9_]*\.sql$`) excludes
an UPPERCASE-suffixed file from `--apply` while still **listing** it under *"Sidecars
(hand-run only, NOT applied by this runner)"* — held back visibly rather than hidden.
Verified: `--no-probe` shows 281 under Sidecars and not in the Pending 67.

The general form is now a landmine: **a migration's guard checks for DRIFT, never for
ORDER.** A `WHERE` clause that refuses a changed row still applies happily at the wrong
moment, and the applying session sees an ordinary successful run.

`sql_for_agents/278` — `bugs_open/154`'s config half, same two-part shape, same banner — is
exposed identically and sits in the same pending 67. Told that lane in their bug file rather
than renaming their file for them.

## 2026-07-31 — my LANDMINES append was swept into another lane's commit

Committed my rename with a pathspec naming LANDMINES.md and git reported only 2 files. The
entry was already at HEAD: the bugs_sweep/111 lane's `f076f4bd1` ("close(111): footer Contact
heading…", 22:28:32) had taken it — 33 LANDMINES lines and 55 WRONG_CALLS lines in a commit
about a footer.

Nothing was lost and forward-only holds, so this is recorded rather than repaired. It is the
exact scenario CLAUDE.md describes: **a pathspec commit stops me sweeping up others' work; it
cannot stop a session running `git add -A` from sweeping up mine.** The practical lesson for
this lane's remaining work is the one already in that file — commit each coherent piece
immediately and narrowly, and expect append-only fleet docs (`LANDMINES.md`, `WRONG_CALLS.md`)
to be the most contended files in the tree, because every lane appends to them.

## 2026-07-31 — the verdict, and the objection that would have been fatal

**APPROVED**, `757cc7be`, 8 advisory objections, none high, 12 seats (4 abstained).

The one worth recording is `editquality`'s: *"if `pipeline` differs per promoting agent, the
count silently under-counts exactly the items the fix exists to catch."* That is not an
opinion, it is a query, and if the answer had gone the other way the fix would have been
wrong in the specific way that is hardest to see — a count that looks right and is low.
**Checked: all three callers leave `target_pipeline` unset, so all three take the Go default
`build`.** Corroborated independently by `TriageDetectedItemsInputSpec`'s own comment, which
already recorded that `target_pipeline` is set by zero definitions fleet-wide. I had relied
on that comment without re-deriving it; the seat was right to make me.

Three more were checkable and all resolved in the change's favour: the jsonb path resolves
at the asserted depth; `snapshot_agent`'s 2-arg overload writes a real retrievable row into
`agent_definitions_backup`; and **0** live definitions declare `has_items` in an
`output_contract`, so the DECLARED CONTRACTS objection exposes an unenforced convention
rather than a defect in this plan — which is what that seat itself said it would mean.

**The lesson I want to keep: four of eight objections were answerable by a query I could
have run before submitting.** Each cost a round of council attention that a `SELECT` would
have saved, and the submission's own `grounded_in` block is where they belonged. The gate is
cheap-ish but it is not free, and "the reviewer can check that" is the thing CLAUDE.md
already tells us not to do — *measure the blast-radius claim before you submit; do not ask
the reviewer to.* I did that for the `has_items` consumer census and not for these four.

Acted on rather than filed away: a lockstep drift alarm for the status literals
(`guardian`), `bugs_open/171` for the second false-clean route and `RFC_006` for the class
(`bug_historian` + `architecture`), and a LANDMINES entry for the shared step
(`tooling_provenance`).

**Second sweep of the day:** this LANDMINES append also reached HEAD inside another lane's
commit (`4c6387139`, the 167 lane) before my own pathspec commit ran. Same as the earlier
one — nothing lost, recorded rather than repaired.

## 2026-08-01 — live, verified, closed

`v1.0.1225` carries it. Gated in the right order and it mattered: **pod-grep BOTH replicas
first** (`site_dispatchable`=3, control=1, same exec), **then** apply migration 281. Had the
grep come back 0 — as it did on `v1.0.1223` the night before, which rolled without the commit
— applying would have sent every run down the clean branch.

Migration applied 08:03Z: pre-flight `1`, `snapshot_agent` captured `source_version=1` into
`agent_definitions_backup`, `UPDATE 1`, verify-before-commit printed the new condition,
`COMMIT`. Then `--record-only` — and the runner **refuses to record a `_HOLD` sidecar**, which
is the right refusal (a held file is not an applied one), so the rename came off first. That
ordering is worth remembering: hold with the suffix, apply, drop the suffix, record.

The verification run (`21669589`, same site, same script) is the after half of a matched pair:

```
before  911ecdd8  v1.0.1218  parent promoted 0, has_items false, no such key -> complete_clean, 0 rerenders
after   21669589  v1.0.1225  parent promoted 0, has_items false, site_dispatchable true / 42 -> complete, 1 rerender
```

**The input to the branch is identical in both runs.** I did not have to construct the failing
condition — the platform produces it every time, which is the whole bug — so the pair is an
induction, not a demonstration. The artefact: `needs_rerender` /
`improvement_rerender_vetcomparison.uk` / priority 99 / handler `rerender-pages`, created
`08:07:31.78Z`, the same second the branch fired. Checked beforehand that no non-terminal
`improvement_rerender%` row existed on that site, so dedup could not have suppressed it and
its presence is a clean signal rather than a survivor.

One thing I nearly got wrong: mid-run I read the clock as ~7 minutes elapsed and started
wondering whether the LLM audit had hung (`bugs_open/029`'s shape). It had been 2.5 minutes. I
had compared a `date` from one command against an `updated_at` remembered from another.
**Read both clocks in the same breath** — `SELECT now()` beside the row, or `date -u` in the
same command — which is already in `WRONG_CALLS.md` from another lane in a different costume.
Caught before it cost anything, which is the only reason it is here and not there.

---

## 2026-08-02 — RFC 006 decided (option a), implemented, applied; one self-inflicted defect on the way

**Owner ruled option (a): one promoter, one owner.** Asked to explain the RFC for a
decision, I first ran the census the RFC itself named as the missing fact — see
`README_where_we_are.md` and RFC_006 §5 for the numbers. Summary: no other parent
calls either child (definition scan returns 2 rows, both `improvement-loop`), and
`agent_run_stats` reads 3/3/3, so neither child has ever run standalone. That was
option (a)'s entire stated cost, and it evaporated.

**What shipped, in the order it shipped.**

1. **Go half first, committed `49ecdf4fd`, council `60f4b425` submitted alongside.**
   `ActionInputSpec.SingleOwner` (runtime-inert) + `ListSingleOwnerActions()` +
   `config-key-audit --single-owner-actions` + `scripts/audit-single-owner-actions.sh`.
   Reused the existing audit binary's decode and `validation.WalkSteps` traversal
   rather than writing a second tool — WFA-004's own precedent, and the reuse seats'
   standing objection to a second binary.
2. **Detector armed against the LIVE fleet BEFORE changing anything**: `181 agents
   decoded, 1 declared single-owner action, 1 finding` naming all three carriers.
   That is the pre-state, measured rather than assumed.
3. **Migration 286 applied by hand** — pre-flight `rows_to_change_expect_2 = 2` and
   `owner_still_has_the_step_expect_1 = 1`, snapshots of both children into
   `agent_definitions_backup`, two `UPDATE 1`s, verify-before-commit.
4. **Detector re-run: `0 findings`.** Same fleet, same command. A matched pair.

**THE MISSTEP — 286 shipped a dangling `error_step`, and my own check told me so.**

`design-audit-agent.call_content_auditor` carried BOTH `next_step: triage` and
`error_step: triage`. 286 repointed the success edge and deleted the step; the error
edge was left pointing at nothing. Check (iii) of 286's own verify block — which I
wrote precisely to catch this — returned exactly the right row naming exactly the
right step. **And the transaction committed anyway, because the check was a
`SELECT`.** psql prints a result set and carries on to `COMMIT`; a non-empty result
is not an error, so `-v ON_ERROR_STOP=1` never fires. The header said "expect ZERO
rows … ROLLBACK", which is an instruction to a reader, not a mechanism.

Fixed forward by **288**, whose equivalent check is a `DO` block that `RAISE`s. I
did not trust it: I induced the exact defect inside a transaction, confirmed the
block aborts, and rolled back — then re-read the live row to confirm it was still
`complete`. Then widened the same query fleet-wide with the agent filter dropped:
**0 rows**, so 286 was the only instance, and nothing in the platform had ever asked
that question before. The widened form is recorded at the foot of 288 and as a
LANDMINE.

Both are logged in `WRONG_CALLS.md` (2026-08-02) and both became LANDMINES entries,
because the second one is prospective: it fires when you TOUCH a step-deleting
migration, with no symptom.

**One thing that went right and is worth keeping.** I re-checked the migration
number immediately before writing the file rather than trusting the one I had
planned — **285 had been taken by two other sessions** in the interim. That habit
exists because of the 2026-07-31 WRONG_CALLS entry and this is the second time it
has paid.

**Concurrency, handled rather than swept.** `cmd/config-key-audit/main.go` picked up
another session's in-flight `--relay-gaps` work while I was editing it, and their
`relaygaps.go` was still untracked. Committing main.go would have broken HEAD's
build fleet-wide (`undefined: emitRelayGaps`) — verified by extracting `git archive
HEAD`, overlaying my file and building. So I moved my detector into its own
`singleowner.go`, committed everything self-contained, and held back main.go's
4-line dispatch and the script (which without the dispatch would silently fall
through to the default mode and print the wrong thing). Both follow once that
session commits.

---

## 2026-08-02, later — the detector now runs on a clock (owner directive)

The council `architecture` seat's low-severity objection turned out to be the most
useful thing in the round: the detector exited 1 on findings and **nothing ran it**.
Confirmed before acting — `grep` over `.githooks/`, `scripts/` and the Makefile found
no gate invoking *any* audit script; `audit-unregistered-actions.sh` (WFA-004) and
`audit-config-keys.sh` are on-demand too. The owner directed it wired.

**Chose a CronJob over a pre-commit hook, and the first reason is the decisive one:**
at commit time a migration has not been applied, so live `agent_definitions` still look
correct — a pre-commit hook would cheerfully pass the very change that re-creates the
fan-out. And config on this platform is routinely changed straight in the database with
no commit at all, which a repo-side hook cannot see even in principle. The latency
argument (a cluster round-trip in ~30 sessions' commit path) is real but secondary; had
it been the only objection I would have path-scoped a hook instead.

`deployments/kustomize/services/single-owner-carriers-check`, modelled on
`bugs-open-staleness-sweep` — same image, same secret, same `doc_notes` convention,
same "connect to Postgres directly, not via kubectl exec" constraint. Daily rather than
weekly: live agent config changes many times a day here. Needs no GitHub token.

**The deliberate trade, stated because it is the weak point.** The job cannot run the
Go binary: that needs a clone of a 262M repo plus `go mod download` plus a compile with
uncertain egress, and **a gate that fails for infrastructure reasons gets ignored** —
which would leave us exactly where we started, with a mechanism that exists and does
not protect. So the job carries a Python mirror, buying two drift risks:
(a) the declared-action literal, (b) a second `WalkSteps`. **(b) is the `bugs_open/144`
shape and I am not pretending otherwise.**

Both are pinned by `cmd/config-key-audit/cron_parity_test.go`, and — the part that
matters — **both guards were proven able to fail**:
- swapping the literal → `TestCronCheckDeclaredListMatchesTheRegistry` fails;
- deleting the nested-descent line → `TestCronCheckAgreesWithTheGoDetector` fails on
  exactly the `nested` and `both shapes` fixtures, and on nothing else.

**Parity on the live fleet, and specifically on a NON-EMPTY result.** Both agreed on
`[]` over the real export, which proves almost nothing — two blind checks agree. So I
injected a second carrier into the real 1.09MB export (one top-level, one nested inside
a loop's `substeps`) and got byte-identical findings and exit 1 from both.

**The gate itself proven in both directions.** A real run walked 179 live definitions,
reported 0 violations, completed, wrote its row. Then a fail-proof run — mutating only
the *check's own* declared list to include `complete_workflow`, never touching a live
agent definition — came back **`Failed`** with the full violation report. I deleted
that run's `doc_notes` row afterwards, since it was a deliberately false finding and
leaving it would have misled the next reader.

Note the count moved: my local run saw 181 live definitions, the job saw 179, minutes
apart. Not an error — other sessions change the fleet continuously. Worth remembering
before treating any such count as a fixture.

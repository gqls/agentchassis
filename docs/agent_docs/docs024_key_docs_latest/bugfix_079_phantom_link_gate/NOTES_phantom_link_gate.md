# NOTES — bugs_open/079 phantom link gate

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-07-26 — coverage check first

079 was filed by `tool_suggester_phantom_links` while closing 029, and it overlaps almost
exactly with `bugs_open/071` (same file, same guard, same `valid :=` line). So the first
question was not "how do I fix it" but "is someone already on it".

All three adjacent workstreams disclaimed it **in writing**:

- `tool_suggester_phantom_links`, `README_where_we_are.md:99` — *"I have not quietly folded
  it into this change, because making that check block publication is a fleet-wide decision
  that could stop pages shipping for a reason nobody has measured yet."*
- `brochure_component_library` (files 071), `HANDOFF_2026-07-26_continue_here.md` §6 routes
  it to *"platform thread; candidate 1 (persist warnings) is small and worth doing alone"*.
  071 is not on their next-action list.
- `cta_link_integrity` closed 049 and wrote into 071: *"This is evidence, not a competing
  fix."*

No commit has ever touched `validateInternalLinks`, the severity, or the `valid :=`
predicate. Working tree clean of edits to it.

**Trap worth recording:** `scripts/who-owns.py 071` looks alarmingly busy — CLOSED, council
APPROVED, tombstone verified live. All of that belongs to
`bugs_closed/071_…agent_job_cleanup…`, a **different** bug sharing the number. The doubled
numbers are documented in `bugs_closed/README.md`; resolve by slug. I nearly read that
activity as "someone is fixing my bug".

## 2026-07-26 — the measurement, and what it ruled out

079 said explicitly: do not promote the severity without measuring first. Two things I had
wrong going in:

1. **I expected `collected_data` to be pruned at ~24h** (071 gap 3 says so, and it is true
   of the retention *policy*). In fact `orchestration_states` reaches back to 2026-07-13 —
   13 days. The census 079 asked for was possible after all. Had I trusted the doc I would
   have chosen a fix candidate with no data at all.
2. **I expected a mix of repairable and invented targets**, because 049's post-deploy census
   found 8 extension-less targets that resolve at `.html`. In the pre-deploy gate's own
   sample, **15 of 15 were pure inventions** — nothing resolved in any form. The repairable
   class is real but it is not what the gate is mostly seeing.

The numbers: 16 validated builds in the window, 3 carried phantom links (19%), 17 instances
/ 15 unique targets, and all 3 deployed with `valid: true`. Both affected pages were
**homepages** — oufe.com and webdesign.co.uk. That is what killed the "promote to error"
candidate: on `!valid` the action returns `(nil, error)` and routes to `mark_needs_review`,
so the page never saves and never deploys. Two homepages would not have shipped.

`bugs_open/083` killed the work-item candidate independently: `phantom_internal_link`
detected 22 times, fixed **zero** times ever, 98 rows fleet-wide stuck at `detected`,
because the only promoter lives in the disabled `improvement-sweep`.

So the repair has to happen in-band at the gate or it does not happen.

## 2026-07-26 — the upstream finding (out of scope, filed separately)

While tracing where the invented links come from: `prepare_link_context` IS wired into
`page-content-writer` and is supposed to hand the model the site's real page list. It finds
nothing. `db_sync.pages` and all three fallbacks are absent from that workflow's
`collected_data`, so `page_count: 0`, so `link_constraint_text` is empty, so the prompt's
`{{if .link_context.link_constraint_text}}` guard elides the whole "ONLY link to these
pages" block. **20 of 20 recent writer runs: zero pages.** The writer is unconstrained by
construction.

**Misstep avoided by checking, not assuming:** my first instinct was that
`InjectLinkConstraints` (defined, never called) was the missing piece and wiring it was the
fix. It is not — it is dead code duplicating `prepare_link_context`, which already runs.
Wiring it would have added a second, competing implementation of the same prompt block.

Filed separately rather than folded in: different mechanism, different file, different
agent's path, and it is a content-generation behaviour change nobody has measured. Noted
into `bugs_open/071` under its candidate 4, which owns the writer-side class.

## 2026-07-26 — MISSTEP: I wrote a test that could not fail, and nearly shipped it

After the code was green I ran the induced-fault probe (disable `RepairPageLinks`, expect
every repair assertion to fail). It reported **4 failures out of 8 repair tests**. I read
that as "4 tests are vacuous" and started looking for what was wrong with them.

That reading was wrong, and the truth was worse. The four "passing" tests **never ran**.
`TestRepairPageLinks_RewriteEmitsTheStoredURLNotAConstructedOne` indexed `repairs[0]`
without a length check; with the fault induced, `repairs` was empty, `repairs[0]` panicked,
and the panic took down the whole test binary — so every test declared after it was silently
skipped. `go test` without `-v` prints only failures, so a run that executed 4 of 12 tests
looked identical to a run that executed all 12.

Fixed by guarding the index with `t.Fatalf`. Re-probed: **8 of 8 repair tests now fail**
under the induced fault, while the four "must not change anything" tests (byte-identical
clean input, non-page scopes, runtime-fill exemption, empty index) correctly still pass —
which is the discrimination that makes the probe worth running.

Two transferable lessons, and the second is the one I did not know:

- A green suite proves nothing until you have watched it go red. Already in `WRONG_CALLS`
  in other forms; this is another instance.
- **A panic in one test masks every test after it, and the default output makes that look
  like success.** Read `=== RUN` lines, not the FAIL count. Any `slice[i]` in a test needs a
  `len()` guard with `Fatalf`, or it is a hidden kill switch for the rest of the file.

## 2026-07-26 — a second fail-open found while making the first one safe

`loadValidPagePaths` swallowed a query error and returned an **empty** page set. While
findings were only warnings that was survivable — it produced a spurious phantom warning for
every link on the page, noise and nothing more. Once the findings drive a rewrite it is not
survivable: an empty set means *every link is a phantom* and the repair pass would strip the
lot from a page whose links were all fine.

It also never checked `rows.Err()`, so a mid-iteration failure silently **truncates** the
page list. That is the same hazard in disguise and slightly worse, because it would unlink
only *some* links — much harder to notice than losing all of them.

Now returns `(index, ok)`; both detection and repair are skipped when not ok. A NULL
`pages.url` is skipped rather than treated as a load failure — checked the live data first
(0 NULLs across 408 active pages today), but the column is nullable, and treating one
malformed row as "list untrustworthy" would disable link checking for that whole site.

## 2026-07-26 — committed, submitted, NOT live; and two multi-session collisions

Committed `43f254be5` (code + tests + docs; scope report clean, 7 files, no passengers) and
`31d8ac7dc` (gofmt — the pre-commit pattern check caught a struct-tag misalignment that the
build gate would have rejected in CI).

Council submitted: **`SUBMISSION_CORR = 97904892-5c09-4782-aeda-37dd944abdfc`**. All six
`grounded_in` quotes machine-verified byte-identical against the pre-fix file (`git show
f804b84ed:…`) before submitting — a trimmed quote manufactured a false MEDIUM objection on a
previous run, and reviewers cannot open the file to check. No orchestration row after 15
minutes; that is the documented queue latency, **not** a dropped dispatch. Not resubmitting.

**Live state, measured not assumed.** The chassis pod runs `v1.0.1170`, built by another
session while I worked, and it does **not** carry this fix:

```
strings /app/agent-chassis | grep -c "CONTENT_LINK_REPAIR_DETAIL"     -> 0   (my new string)
strings /app/agent-chassis | grep -c "CONTENT_VALIDATION_BLOCKER_DETAIL" -> 2   (positive control)
strings /app/agent-chassis | grep -c "repair_internal_links"          -> 0   (my new config key)
```

That is a *discriminating* pre-state, which is the point of running it before the roll: the
control proves the grep works, and the two zeros prove the fix is absent. Post-roll the same
three commands must read ≥1, ≥1, ≥1. **079 therefore stays OPEN** — the bar is fixed AND
live, and the defect is still reproducible in production until an image ships.

Build held: the owner reported another thread mid-deploy, and racing on `IMAGE_TAG` is the
multi-session hazard this repo has a whole handoff about.

### Two collisions, both while this task was in flight

1. **My `016b` §9 append was swept into another session's commit** (`d5988a8ed`, a
   `bugs_open/006` closure) before I could commit it. Nothing lost, forward-only holds. This
   is precisely the case CLAUDE.md documents — commit-per-task stops *me* sweeping *others*,
   and cannot stop a session running `git add -A` from sweeping *me*. The only real defence
   is committing sooner, and I had a ~4-minute window open.
2. **Number collision on a fresh bug file.** I checked, found 090 free, wrote the file — and
   another session filed *their* 090 sixty-seven seconds before my commit landed. Renumbered
   mine to `092` (`cf2cafcdd`) rather than leave a sixth doubled number in a scheme where
   `bugs_closed/README.md` already lists doubled numbers as a standing trap. **Checking a
   number is free is not the same as reserving it**; on a busy day the check is stale before
   you finish writing the file. Cheap to undo at 67 seconds old, permanent if left.

## 2026-07-26 21:1x–21:3xZ — LIVE in v1.0.1171; first induction attempt died before reaching the code

Another session's build rolled `v1.0.1171` and it carries this fix. Pod-grep against the
baseline recorded above — this is the discriminating pair, not a bare grep:

| string | pre-roll (v1.0.1170) | post-roll (v1.0.1171) |
|---|---|---|
| `CONTENT_LINK_REPAIR_DETAIL` (new) | 0 | **1** |
| `repair_internal_links` (new) | 0 | **1** |
| `link check and repair SKIPPED` (new) | 0 | **3** |
| `CONTENT_VALIDATION_BLOCKER_DETAIL` (control) | 2 | 2 |

**A marker I nearly used and did not:** "the old policy comment `improvement loop resolves it` is
gone" reads 0 on both binaries — **comments are not compiled into the binary**, so it can never be
anything but 0 and proves nothing whatsoever. It is the same vacuous-marker shape already logged
against `052`. Only compiled strings — error codes, config keys, log messages — discriminate.

### The induction, attempt 1: FAILED before it reached the gate

Dispatched `page-build-handler` for `webdesign.co.uk / learn-design-digital-grain`
(corr `a1dfbf68`). Result:

```
FAILED | spawn_content_writer | Request 295ff5da-… timed out after 3 retries
```

It never got as far as `validate_content`, so **it proves nothing about the fix either way** —
neither that it works nor that it does not. Recording it because a failed run that never reached
your code is the easiest thing in the world to quietly discount, and the temptation is to treat
"no contradiction" as "confirmation".

Checked whether it was systemic before retrying rather than assuming: **1** spawn timeout
fleet-wide in 2 hours (mine), 8 orchestrations in flight. So not saturation. Not within 300s of a
pod restart either (pod up 21:02:56Z, dispatch ~21:14:55Z). Cause unresolved — plausibly the
`bugs_open/003` spawn-loss class. Retried as corr `df7437f2`.

### Two route findings worth keeping

1. **The work-item dispatcher is PER-SITE** — `load_work_item_actions.go:559`,
   `WHERE wi.site_id = $1 AND wi.status IN ('triaged','approved')`. A `triaged` row therefore sits
   untouched indefinitely until something triggers *that site's* build pipeline. Inserting a work
   item is not dispatching it. To exercise one page, publish to `page-build-handler` directly by
   kcat (envelope in the RUNBOOK).
2. **`page-rebuild` and `page-rerender` do NOT call `validate_page_content`.** Only
   `page-build-handler`, `content-reviewer`, `tool-recreation-handler` and `report-builder` do.
   The obvious-looking `TRIGGER_rerender_page.sh` would have run green and tested nothing —
   the "verify the failing branch" trap wearing a convenient disguise.

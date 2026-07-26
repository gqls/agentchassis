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

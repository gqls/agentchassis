# HANDOFF — bug 122 lane. START HERE. Written 2026-08-12 (afternoon), chassis `v1.0.1291`.

Supersedes `HANDOFF_2026-08-12_continue_here.md` for **state**. That file stays the reference
for the two owner decisions of 2026-08-12 (§2, the 08-16 rotation date) — **§2 is unchanged and
still owed.** Its §3 fork is superseded by this file. The 08-11 file remains the sweep-episode
reference; the 08-10 file the original delivery evidence.

**Nothing in this lane is on fire.** Queue empty, no site locked out, 226 items parked and
stable. The next task is a **code change with a fully settled design** — §3 is written to be
implementable cold, and the design questions that would have cost the next session a day are
answered with queries below.

---

## 1. State, measured 2026-08-12 ~15:00Z

| thing | state |
|---|---|
| chassis | **`v1.0.1291`**, both replicas, started 14:55Z. Provenance startup line already rotated out of `--tail=300` — normal; that is "not in range", never "unstamped" |
| 226 `contrast_failure` items | still **parked** `deferred`, all `parked_by='migration_389'`, **`max(attempt_count)=0`** — unchanged by the roll |
| `contrast_failure` completions, all history | **ZERO.** Never dispatched, never completed, never re-detected |
| `site-render-audit-rotation` | enabled, weekly per site, zero LLM spend — this is what finds our failures |
| `improvement-sweep` | still disabled. Do not re-enable without reading the 08-11 file |
| `site-discovery-rotation-quality` | enabled, **inert until 2026-08-16 09:49Z** — correct and waiting, do not "fix" it |

**📅 2026-08-16 is still this lane's dated action** — price the rotation ramp (calls AND tokens;
baseline ~248k input tok/h idle, driven sweep was ~806k/h). Queries at the foot of
`sql_for_agents/395_enable_quality_discovery_rotation_slow_ramp.sql`. Unchanged by this file.

---

## 2. What was settled this session (do not re-derive)

The §3 fork of the previous handoff — three ways to write a completion-time `contrast_failure`
verifier — **was costed and has no survivor.** Full evidence in `NOTES_contrast_ink_slots.md`
→ *"the verifier fork was costed"*; committed `b2fca2f8f`. Headlines only:

- The standing objection is at **`verifier_coverage_test.go:199–201`, not `:171`**, and there
  are **three** instances. It kills options (1), (2) **and** (3) — every option that computes
  contrast fetches the page, so (3) was narrower in *predicate*, not in *mechanism*.
- **`contrast_failure` is already an on-record exemption whose reason is unsound**
  (`verifier_coverage_test.go:156`): it is the argument RFC_017 refuted on 08-08 (six days
  after the line was authored, `f2a222964`, and never revisited); it infers resolution from
  **absence**, which `CheckResult.Resolved`'s contract forbids in writing; and it has never
  once been exercised.
- **The answer is option (4): retract on the DISCOVERY path** — `asset_reference_404`'s
  posture, using the shared closer that already exists.

---

## 3. THE TASK: teach the render audit to close its own tickets

### 3.1 Why it is shaped this way (read before editing — two constraints are non-obvious)

**Constraint A — it MUST travel in the adapter's response. There is no shortcut through the
requester.** `request_render_audit_action.go` already computes a `urls_audited` metadata value,
and it is **destroyed**: an awaiting step's own result never survives the park
(`persistAwaitingStateWithRetry` keeps only awaited-request entries — RFC_012 addendum 2,
owner-ruled option B). That is exactly why `bugs_open/242` had to echo `pages_total`/`truncated`
*through the request into the response* rather than fixing it at the requester. Do not
re-discover this by trying the easy route.

**Constraint B — "audited" means SUCCESSFULLY MEASURED, never "requested".** The adapter
already tracks the failure side: `RenderAuditResult.Unreachable []string`, whose own comment
says it exists because otherwise *"it would let a dead page pass as clean"*. A retraction scoped
to requested-URLs would close tickets on pages that failed to load — the precise error that
field was added to prevent, and the "positive control is the load-bearing half" rule from
`cmd/component-render-check`.

### 3.2 The three edits

**Edit 1 — `internal/adapters/browserrunner/render_audit_action.go`.** Add the audited-page
identities to the summary, alongside 242's counts:

- `Summary` struct (`:135-150`): add
  `PagesAudited []string \`json:"pages_audited,omitempty"\``.
  `omitempty` matches the 242 precedent — an old-shape reply degrades to today's behaviour,
  never to a wrong number.
- `Execute` (`:266-317`): append `url` to `res.Summary.PagesAudited` **after** the
  `if err != nil { … continue }` guard at `:272-278`, so an unreachable page is excluded by
  construction. A page WITH findings still belongs in the list — it was measured, and its
  findings are filed; that is what makes per-selector retraction possible.

**Edit 2 — `platform/orchestration/actions/write_render_audit_findings_action.go`.** Decode it
and retract. The local mirror (`renderAuditPayload`, `:134-150`) omits fields deliberately, so
add `PagesAudited []string \`json:"pages_audited"\`` to the mirrored `Summary`.

The retraction, inside the existing `tx` (opened `:336`, filing loop `:343`):

1. Build the set of pairings this run **observed still failing**:
   `workItemKey("contrast_failure", pagePath+"#"+selector)` for every firm
   (`over_image=false`) contrast finding — i.e. exactly the keys already computed at `:266`.
2. `SELECT id, item_key FROM site_work_items WHERE site_id=$1 AND item_type='contrast_failure'
   AND status NOT IN (<closed>)` — the currently-open pairings.
3. Retract a row **only if** its `item_key` names a page in `PagesAudited` **and** the key is
   absent from set (1). Anything on a page not audited this run is left alone.
4. Call `resolveWorkItems(ctx, tx, siteID, "render_audit", batchID, ResolvedFinding{…}, logger)`
   per retracted key. It is unexported but in **package `actions`** — same package, directly
   callable (`work_items_common.go:249`). It validates as a **refusal**: empty `ItemType` or
   `Reason`, or neither `ItemKey` nor `AllOfType`, returns an error rather than guessing.
   **Never set `AllOfType` here** — the wide branch would close every open contrast ticket for
   the site, including pages this run never looked at.

**Edit 3 — `verifier_coverage_test.go:156`.** Rewrite the exemption's reason. Once retraction
ships, the entry becomes **true for the first time** — but its current wording ("the NEXT render
audit is the verifier") describes the unsound absence-inference. Replace it with the
`asset_reference_404` formulation: the check retracts its own findings on a positive
re-observation, on the discovery path where the probe is already precedented. **Leaving the old
wording would mean shipping the fix and keeping the false reason.**

### 3.3 ⚠ FOUR HAZARDS, each measured — these are the ones that bite silently

1. **`deferred` is NOT in `workItemClosedStatuses`** (`work_items_common.go:83-89` — the list is
   `complete, verified, rejected, wont_fix, cancelled`). So **retraction WILL close parked
   items.** Shipping this begins closing the 226 as each site's weekly audit confirms them
   fixed, *without anyone unparking them*. That is arguably the right behaviour — a parked
   ticket whose defect is genuinely gone should close, and it means only the still-broken
   remainder ever needs a fixer — **but it is a change in what the park MEANS and must be a
   stated decision, not a side effect.** Decide it explicitly and say so in the council
   submission. (`unresolved` and `failed` are also retractable, deliberately — owner ruling,
   Decision 2.)

2. **`contrast_failure` rows carry NO `batch_id`, so `resolveWorkItems`' self-protection guard
   is INOPERATIVE.** The closer's `batch_id IS DISTINCT FROM $6` exists so a run cannot close
   what it just raised; `NULL IS DISTINCT FROM <uuid>` is TRUE, so it does not fire.
   **[MEASURED, and disconfirmable — the controls came out the other way]:**
   `contrast_failure` **0 of 226** carry one, against `empty_section` **61/61**,
   `hardcoded_section_colors` **15/15**, `asset_reference_404` **1/1**, all filed through the
   discovery-check path which sets it. `insertWorkItem` never writes the column.
   **Preferred fix: populate `batch_id` on rows this producer files**, which restores a guard
   the estate deliberately built and makes these rows consistent with every other type. The
   alternative — relying purely on the set logic in 3.2 — leaves the correctness of a
   *destructive* operation resting on one loop with no backstop.

3. **A repaired page reports NOTHING, which is why `PagesAudited` cannot be derived from the
   findings.** The audited set is *not* recoverable from `payload.Contrast` URLs: a page that is
   now clean contributes zero findings and is indistinguishable from a page never visited — and
   that is precisely the case retraction exists to catch. This is the whole reason Edit 1 is
   needed; if a reviewer proposes deriving it, this is the answer.

4. **`fg`/`bg` are formatted inconsistently in the same row** — `"rgb(26, 31, 46)"` and
   `"rgb(15,18,24)"`. Retraction as specified keys on selector+page and never compares colour
   strings, so it dodges this — **keep it that way.** Any future refinement that compares a
   recorded colour to a computed one must parse to numeric triples first.

### 3.4 Tests the estate will expect

The culture here is induced-fault proof, not hope (`check_asset_reference_404`'s header:
*"every branch below is proven by an induced fault … rather than by hope"*).

- **The load-bearing negative:** an unreachable page must NOT retract its tickets. Induce a page
  in `Unreachable`, assert the ticket stays open. This is the positive control — a retraction
  that only ever confirms "the bad pairing is gone" will also pass a page it failed to load.
- A pairing still failing on an audited page → **not** retracted.
- A pairing absent from an audited page → retracted, `result.resolved_by`/`reason` stamped.
- A pairing on a page **not** in `PagesAudited` → untouched.
- Old-shape reply (no `pages_audited`) → **zero retractions**, today's behaviour exactly.
  Version skew must degrade to inert, never to a wrong closure.

### 3.5 Process obligations

- **Council gate — required.** This is a shared seam: it adds a new field to an adapter
  contract and gives a producer authority to CLOSE work items it did not file. Submit before or
  alongside the commit; use `Council-Submitted: <corr>` if committing first. **Never** write
  `Council-Reviewed:` on a verdict you have not read. Budget ~30 min, not ~2.
  Name the other consumers of the render-audit response and tell them (owner ruling 2026-07-29
  §3) — measuring that nothing breaks is not the same as their owners agreeing.
- **Concept register** — a new callable retraction path clears the bar. Drop any matching line
  from `102_coverage_ratchet.txt`.
- **Image before behaviour** — Go is inert until a chassis rebuild and roll. `make build-*`
  builds from committed HEAD. Two services change here: `agent-chassis` (the producer) and
  `browser-runner-adapter` (the adapter). **Roll both, and check the stamp PER SERVICE** — a
  release can straddle sessions' commits (`bugs_open/249`).
- **Migrations:** none needed. **Do NOT run `run-migrations.sh --apply` on this tree** — the
  08-12 dry run listed 12 pending files from other threads, one of which deploys the wrong
  asset bytes on an older binary.

### 3.6 Order, and what NOT to do

**Edit 1 → Edit 2 → roll both services → watch one weekly audit → then unpark.** Do not unpark
the 226 first: unparked-and-ungraded is the state the park exists to prevent, and hazard (1)
means a shipped retraction starts closing the genuinely-fixed subset on its own.

**Do not write a completion-time verifier.** Three recorded objections refuse it, and §2 above
is the costing. If a future session is tempted, the fork and its refutation are in NOTES.

---

## 4. Also still open in this lane

| | item | status |
|---|---|---|
| 1 | `bugs_open/212` §8 — component-painted grounds (~24 failures) | **Owner's.** Architecture, not a bug patch. Unchanged |
| 2 | `bugs_open/242` — the silent 25-page cap | **DONE, live `v1.0.1288`, council APPROVED, behaviourally proven** by the `bugfix_242_render_audit_truncation` lane on 08-11. The previous handoff called it "open, unstarted" — wrong. It is the **precondition** for §3, and it shipped the counts; §3 adds the identities |
| 3 | Free cross-check | if a lane re-renders robot-hands `/selection-guide.html`, the audit filed `info-card-grid__card-link` + `__eyebrow` failures and migration `368` should close both. **Grade at the next audit, never at the item status** |

---

## 5. Standing traps this lane has paid for

- **Grade per selector, never by fleet total.** It rose 109 → 112 while every targeted failure closed.
- **A filed count is not a found count.** "34 findings" was 171 firm — 111 dropped by a cap.
- **Read the selection before asserting it excludes your rows.**
- **A pathspec commit still takes a same-file passenger** — and it happened AGAIN on 08-12:
  my `LANDMINES.md` correction went out inside another session's commit `dbf74bc71`. Nothing
  was lost (verified at HEAD, not at the tree), and forward-only holds. Expect it; verify at HEAD.
- **`pages.sections` is an array of plain strings**; an object-shaped census returns 0 rows silently.
- **Never run `run-migrations.sh --apply` on this tree.**
- **A call count does not price an LLM loop.** The sweep held 93–184 calls/h while running 3.2x
  the input tokens. Threshold the expensive unit.
- **Put a row COUNT you could be wrong about in your post-check.**
- **NEW — a `file:line` in a handoff is a pointer, not a quotation.** The `:171` citation this
  lane inherited (and `LANDMINES.md` carried) was stale; opening the file changed the answer
  twice over. Cite by entry name or symbol, never by line alone.

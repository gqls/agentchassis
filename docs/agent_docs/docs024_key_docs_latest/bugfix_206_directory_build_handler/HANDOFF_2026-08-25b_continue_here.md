# HANDOFF 2026-08-25b — continue here (`bugs_open/206`)

**Supersedes `HANDOFF_2026-08-25_continue_here.md`** (keep it; accurate for its own day, and its
§2 correction about `garden-tools.uk` still binds — repeated below).
Read `bugs_open/206` **bottom-up**: its last four sections are 2026-08-25, in order.

---

## State in one paragraph

**The code is finished, reviewed and LIVE.** Both page-build producers now call one routing
authority (`builderForPageType`), `section-index` routes to `directory-build-handler`, and the fleet
carries it — chassis **`v1.0.1339`**, stamp `a7459a44b`, ancestry-verified with a working control.
A second change shipped tonight (`1887a116b`) that makes the closure proof permanent rather than
perishable; it is **committed, submitted, and NOT yet rolled**. **The bug stays OPEN for one reason
only: nobody has watched a mint happen on a page type where the bug and the fix differ.** That needs
a greenfield build of a site carrying an `entity-directory` or `entity-page` page. One arrived today
(`homegarden.uk`) and could not settle it — see §3.

---

## 1. What is DONE — do not redo

| commit | what |
|---|---|
| `d1aa231aa`, `200d54bdf` | 08-24: `builderForPageType` created; `reconcile_site_plan` calls it. Council **APPROVED** r6 (`52dbd067`). Live since `v1.0.1334`. |
| `efec862f4` | The swap: `WriteBuildItemsAction`'s inline maps **deleted**; `section-index` added to the shared map in the same commit; `capability_gap` `handler_agent` → EMPTY. Council **APPROVED r1** (`b92e624d`), 13 reviewers, no vetoes. **Live on `v1.0.1339`.** |
| `0777eb297` | Closed two coverage gaps an adversarial review found — including one the swap itself created (see §5). |
| `1887a116b` | **Routing provenance**: both doors stamp `spec.handler` = the handler they chose. Council **submitted `9ff151d6`, verdict PENDING — read it** (§2). **Not rolled.** |

**Docs done**: `bugs_open/206` (4 sections today), RUNBOOK §7/§7a/§7b/§7c/§7d, NOTES (missteps 1–6),
README (3 owner entries), **BLD-027** + index row de-staled, `LANDMINES` (1 entry + 2 corrections),
`WRONG_CALLS` (3 entries today), 2 memory lessons.

---

## 2. What is LEFT — in priority order

### (a) Read the pending council verdict — `9ff151d6-c521-4bff-9ee8-e5c9ab747a52`

`1887a116b` is on the shared branch already (forward-only; that is correct and expected). If the
verdict is REVISE or REJECTED, **act on it in a follow-up commit** — do not amend.

```sql
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '9ff151d6-c521-4bff-9ee8-e5c9ab747a52';
SELECT metadata->>'decision', body FROM diagnosis_artifacts
WHERE correlation_id='9ff151d6-c521-4bff-9ee8-e5c9ab747a52' AND kind='council_report'
ORDER BY created_at DESC LIMIT 1;
```
`098` credits the commit automatically once approved — **never** hand-write `Council-Reviewed:` on
a verdict you have not read.

### (b) THE closure proof — still the only thing between this bug and `bugs_closed/`

**It is free and needs no site touched. It is also not schedulable by this lane.** It arrives when
any lane builds a greenfield site **carrying an `entity-directory` or `entity-page` page**.

⚠ **`section-index` is NOT sufficient** — that is today's hardest-won lesson (§3). Check the plan
*before* spending a build on the question:

```sql
SELECT page_type, count(*) FROM pages
WHERE site_id=(SELECT id FROM sites WHERE domain='<domain>') GROUP BY 1 ORDER BY 2 DESC;
-- no entity-directory and no entity-page ⇒ this build CANNOT close 206, whatever else it shows
```

Then use **RUNBOOK §7d** (permanent gate) if `1887a116b` has rolled, else **§7** (fingerprint +
`updated_at`, read promptly). §7c is how to tell which.

### (c) Follow-ups — named, evidenced, NOT this lane's to sneak in

- **`emitted++` without `RowsAffected`** (`reconcile_site_plan_action.go:484`), while the gap arm
  four lines above reads it and cites `bugs_open/091`. One line. Causes phantom `pages_emitted` and a
  repeated `needs_rerender` whenever an open `needs_page:<name>` row is invisible to
  `loadOpenPageItems`.
- **The `unresolved` divergence**: `loadOpenPageItems` (`:713`) treats `unresolved` as OPEN so
  reconcile skips the page and new routing never reaches it, while `idx_swi_dedup` does not cover it
  and both claim gates exclude it. Nothing can free such a row. One live casualty:
  **`adversecreditmortgage.co.uk` `blog-index`**. Changes a dedup contract → its own round.
- **`needs_directory` is a write-only item_type**: 0 rows ever minted, 0 Go readers outside
  `builder_routing.go`, 0 live agent configs. Retiring it touches
  `create_tool_cross_link_items.go:263`'s gate.
- **Residual (b), the larger and better fix**: `page-build-handler` cannot fill a *missing* layout
  for **any** type, because `ensure_page_section_layout` lives only in `directory-build-handler`'s
  workflow. `blog-post`/`blog-index` casualties measured on four sites. The right shape is making the
  layout-ensuring step reachable from the generic path — **not** routing more types to
  `directory-build-handler`. Its own submission.

---

## 3. What today's greenfield build proved, and why it was NOT enough

The `bugs_open/381` lane built `homegarden.uk` and snapshotted the mint before dispatch touched it —
their file:
`docs/agent_docs/docs024_key_docs_latest/bugfix_381_inexpressive_composition/evidence/MINT_SNAPSHOT_homegarden_20260825T113422Z.txt`

`[MEASURED]` 21 pages: **17 `section-index`, 2 `content`, 1 `landing`, 1 `blog-post`. Zero
`entity-directory`, zero `entity-page`.**

**Every one of those types routes to `page-build-handler` under BOTH the old hardcoded literal and
the new map.** So the mint was stamped, untouched, correctly routed — and *identical to what the bug
would have produced*. **It confirms the new code RAN; it cannot show it ROUTES DIFFERENTLY.**

Two caveats settled by it, free:
- **RETIRED**: "which door mints on a greenfield build?" — `reconcile_site_plan` minted all 22;
  `WriteBuildItemsAction` did not appear as a `created_by` at all.
- **CORRECTED**: `spec ? 'page_type'` is no longer a zero-population fingerprint (21 stamped rows
  now exist), and it only ever proved the **08-24** commit minted the row, never today's swap.

---

## 4. Cross-lane state — read before touching any site

- ⛔ **`garden-tools.uk`: NOTHING IS TO BE CLEARED (owner ruling, 2026-08-25).** The authorisation was
  **retracted**. See `CONTRIB_2026-08-25_from_loanzy_lane_the_owner_retracted_the_parked_row_authorisation.md`
  in this directory. Its value is that it is an unassisted greenfield build four lanes measure
  against. Three of the six parked rows are there.
- `dartsonline.com` brand-detail and `loanzy.uk` guides-index are **not** covered by that ruling.
- ⚠ **A hand re-triage proves NOTHING about this fix** and is now a known **false-PASS path** — it
  sets `handler_agent`, which is exactly what the closure test reads.
- The `bugs_open/381` lane is a good correspondent: they captured evidence on request, and they
  **caught a stale caveat of mine**. I also sent them a **false warning** which they had already
  adopted (§5). Their bug and this one overlap at the symptom and not at the cause.

---

## 5. The five traps this lane has actually fallen into — read before measuring anything

Every one produced a confident, wrong, plausible answer. Full accounts in `WRONG_CALLS.md` and NOTES.

1. **`handler_agent` has two causes.** The fix writes it; so does the documented operator repair. All
   three rows fleet-wide that would have passed the closure test were **hand repairs**. → §7d's stamp.
2. **A test can pass under mutation for more than one reason.** My first test file passed while
   deliberately broken, for two independent reasons — and fixing the first did not make it fail.
   *A mutation that still passes after you fix the reason it passed means there was another reason.*
3. **A fixture producing zero rows passes every assertion about those rows** — `nav_order: nil` failed
   a scan, emptied the page list, and the action logged "no per-page builds needed" and returned
   success.
4. **Three broken deploy probes in one command**, with a 40-zeros "negative control" that came back
   **PRESENT**. → §7c: two timestamps, then ancestry against the stamp, always with a control.
5. **I warned a peer lane about a mechanism without checking its precondition on their data**, and
   they had already put it in their acceptance guide. 206's no-op needs a **layout-less** page; all
   17 of theirs had layouts and one had already deployed through the handler I warned about.
   *A check that stops an investigation is more dangerous than one that starts a wrong one.*

**And the meta-lesson, which cost most:** a correction is not exempt from the discipline it is
enforcing. The measurement that caught trap 1 computed the very discriminator its own fix omitted.

---

## 6. Closure test — when may this move to `bugs_closed/`?

**Precondition (outranks everything):** the build carries an `entity-directory` or `entity-page` page.

1. A `reconcile_site_plan`-minted row for that page at the right handler — `entity-directory` →
   `directory-build-handler`; `entity-page` → a deferred `capability_gap` with an **empty**
   `handler_agent` — **and provably minted rather than repaired** (RUNBOOK §7d, or §7's two gates
   read promptly if `1887a116b` has not rolled).
2. That page **built and serving**, verified by `curl`, not by `build_status`.
3. The parked `entity-directory` / `entity-page` rows resolved to their designed outcomes — **except
   `garden-tools.uk`, which must not be touched** (§4), so its rows are recorded as a stated gap
   rather than resolved.

`section-index` staying parked is **no longer** part of any deliberate narrowing — that ended
2026-08-25. But an `unresolved` row still cannot move, for the separate reason in §2(c).

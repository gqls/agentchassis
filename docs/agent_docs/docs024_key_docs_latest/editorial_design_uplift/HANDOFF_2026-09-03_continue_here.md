# HANDOFF 2026-09-03 — editorial_design_uplift, continue here

**Supersedes `HANDOFF_2026-09-02_continue_here.md`.** That file is still worth reading ONCE, for its
§2–§4 (the migration 686 rollback, the real imagery finding, the planner-prompt answer) — none of
which changed today. This file replaces its §0, §5, §7 and §9.

**Branch:** `087_towards_multiple_domains`. Evidence is in `NOTES_editorial_design_uplift.md`,
2026-09-03 entry. Plain-prose account for the owner: `README_where_we_are.md`, same date.

---

## 0. Environment as of 2026-09-03 ~13:00 BST

1. **Kubeconfig is ALIVE.** The 09-02 expiry is resolved; `kubectl` works.
2. **Chassis rolled to `v1.0.1356`** (pods up 08:57–08:58Z). The post-roll migration dry-run is
   **DONE**: `Pending (177)`, and **36** files flagged `LIKELY ALREADY APPLIED; its own guard raised`
   (that second figure is `bugs_open/426`'s and was 34 on 09-02 — it is climbing).
   ⚠ The dry-run takes **over five minutes** and prints nothing until it finishes. Run it unpiped, in
   background.
3. **The owner announced a FURTHER chassis build mid-session**, so another roll is likely imminent —
   which means **another migration dry-run is owed after it**, and no orchestration dispatch within
   ~300s of the restart. Nothing in this handoff depends on which image lands: today's change is
   inert by measurement.
4. **`708_enable_unrendered_page_imagery.sql` is now APPLIED**, so IMG-077 is no longer inert. The
   114 lane discharged it (`21e18a504`) — the 09-02 handoff's §5 and §9 say otherwise and are stale
   on that line only.

---

## 1. What happened today, in one paragraph

The imagery half of this lane is parked exactly where 09-02 left it (see that handoff's §2–§4 —
**do not restart imagery work at the component layer**). Today was the structure half: **035 P1
direction 2 is wired**. `recomposeAncestors` — written 08-31, reviewed over three council rounds,
committed, and never called — now runs, is guarded, is stamped, and has tests. The reason it was
never called turned out to be the finding.

---

## 2. THE FINDING: a parameter no caller could supply, justified by a paragraph that measured nothing

`recomposeAncestors` took `tx *sql.Tx`. Its own header explained why in capitals — *"THE db/tx SPLIT
IS FORCED, not stylistic"* — reasoning about which reads must see *"the uncommitted edit"* inside
`apply_section_edit`'s transaction.

`[MEASURED 2026-09-03]` **There is no transaction:**

```
grep -nE 'BeginTx|\.Begin\(|Commit\(\)|Rollback\(\)' platform/orchestration/actions/section_editor_actions.go
→ (no matches)
```

Every persist there runs through `updatePageComponentAfterEdit(ctx, params.DB, …)` on the autocommit
connection. **No call could compile.** The function sat uncalled for three days, Go's linker dropped
it, and the 09-02 binary probe reported the symbol ABSENT — which reads exactly like a missing
commit. (§9 of that handoff got the interpretation right; it just could not see the cause.)

**Three council rounds reviewed the design and none reached it. One grep did.** A comment can assert
a fact about its caller and nothing type-checks it. The claim was specific and technical — it
correctly named two functions that cannot take a `*sql.Tx` — and inferred a transaction from that,
which does not follow.

**Attempting the wiring then found three more defects, none of them visible in a design review**,
because they live one level below the seam the rounds argued about. The ancestor write:
carried **neither the tombstone nor the lock predicate** its sibling writes on that path carry;
read **zero rows affected as success**; and was **unstamped** though it writes a column the 357/552
triggers archive (`bugs_open/355` A1). All three fixed.

⚠ **The half with no detector, worth inheriting:** an ABSENT guard is invisible to this estate's
existing checks. `TestNoHandSpelledTombstonePredicate` catches a *wrong spelling* of the tombstone
predicate and never a *missing* one; `page_component_writer_coverage_test.go` asks only about the
floors. So a green suite says nothing about whether a NEW writer of `page_components.rendered_html`
is guarded — list the predicates on a sibling write and match them deliberately.

Recorded in `LANDMINES.md` (footprinted, verifier dispatched), `WRONG_CALLS.md`, and
`features_open/035` §5.

---

## 3. State of 035 P1 — the audit, so nobody re-derives it

| deliverable | state |
|---|---|
| `deriveRenderMode` third value (`composite`) | **DONE** `1f745e730` |
| membership helpers | **DONE** `bc8167100`, reached in production |
| direction 1 — refuse to render a composition parent alone | **DONE** `028c3e112` |
| flat-pass extraction | **DONE** `2a0bdb001`, `94f81cc60`, `22ed53ee7` |
| direction 2 — `recomposeAncestors`, **called** | **DONE today** `1007be27d` |
| `check_render_mode` routing arm | **REFUTED, not deferred** (`5542a76d6`) — nothing reads `render_mode`; P1's routing story cannot work as 035 wrote it |
| the walk in both render paths, §6.9's filter, register entry, live canary | **NOT DONE — this is what remains** |

**Still inert, re-measured today:** **0 of 3,229** `page_components` rows carry a
`parent_instance_id` `[MEASURED 2026-09-03]` (0 of 2,249 on 08-31; 0 of 2,005 on 08-24 — the table
grew ~1,000 rows in ten days and the parented count stayed at zero). Direction 2 adds **one indexed
SELECT per edit** to the live edit path and nothing else.

---

## 4. What to do next, in order

1. **Read the council verdict** on `cab931b1-8b45-461e-8a37-0dbdfa6aa928` and act on it. The code is
   already on the shared branch, so a REVISE is a follow-up commit, not a hold.
   ```sql
   SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
    WHERE correlation_id='cab931b1-8b45-461e-8a37-0dbdfa6aa928' AND kind='council_report'
    ORDER BY created_at;
   ```
2. **THE READ PATH — the remaining core of P1.** `walkComponentHierarchy` still has no production
   caller, so a row that opted in today would render flat. Its own council round.
3. **Hazard 6.9's filter MUST land inside that change, not after it.** `loadStoredSections` selects
   `COALESCE(parent_instance_id::text,'')` but its WHERE is only `page_id = $1 AND <not removed>`, so
   every row comes back flat. The moment the walk renders children in a nested pass, children are in
   BOTH lists, every later section's `NextOccurrence` index shifts, and per-section figures attach to
   the wrong sections — rendering, deploying and looking correct. Read 035 §6.9 in full first: it also
   carries the `MergeLockedPageSlots` inverted-polarity trap for any plan-vs-live guard.
4. **Then the register entry and the live canary**, which are P1's actual acceptance.
5. **`news-listing`** — same defect `article-body` had, still unwritten, still deliberately behind the
   09-02 handoff's §3 question.
6. **Do NOT restart imagery work at the component layer.** Unchanged from 09-02 §7.4; the live
   question is the planner's page composition and it belongs to `bugs_open/114`.

---

## 5. ⚠ HEAD is RED in a neighbouring package, and it is not this lane's

`go test ./platform/orchestration/...` fails
`discovery_checks/TestStylesheetGutted_TokenSetMatchesCanonicalCSSTokens`:
*"canonicalCSSTokens declares 4 token(s) this check does not police: [--color-accent-ink
--color-accent-text --color-cta-bg-ink --color-primary-ink]"*. Neither file is dirty, so it fails at
committed HEAD; `git log` names `0325ddebb` (2026-09-03 12:10, the 458 lane), which added the tokens
without extending `rendererGuaranteedTokens`. **Left alone deliberately** — the test message tells
that lane exactly what to do — but know it before you read your own `go test` output, and scope your
runs to `./platform/orchestration/actions/` if you want an unambiguous green.

---

## 6. Identifiers

- commit `1007be27d` (direction 2 wiring); council corr `cab931b1-8b45-461e-8a37-0dbdfa6aa928`
- submission JSON: `editorial_design_uplift/COUNCIL_SUBMISSION_2026-09-03_035_p1_direction2_wiring.json`
- new writer stamp `action:recompose_ancestors`; new test file
  `platform/orchestration/actions/component_hierarchy_recompose_test.go` (M1–M4 mutations recorded in
  its header, all killed)
- everything from the 09-02 handoff's §8 still stands: boxingonline site `d2aa5206-73bc-4707-a69c-2702c1eb9152`
  serving at `boxingonline.ugg2.com`; `article-body` `5835b2e1-50d7-4f20-8a9c-8da4d270ae3d` at md5
  `002cbcd9cada6a37bf4a5158fd1e5f22`; planner definition `f263eaa1-61e1-446e-9410-648e12b7875b`

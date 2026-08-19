# HANDOFF — portfolio_positioning — 2026-08-19. **START HERE IN A FRESH SESSION.**

Supersedes `HANDOFF_2026-08-18c_continue_here.md` (accurate on its own history; everything a
fresh chat needs is carried forward here). Milestone read-out for the owner:
`SUMMARY_2026-08-19_first_sites_live_and_the_wall_the_fleet_would_have_hit.md`.

**Verified at the start of this file, 2026-08-19** — do not re-derive, but DO re-check if hours
have passed, because this tree moves fast (≈120 commits landed under us overnight):
- **Chassis `v1.0.1314`, built from `d3590ca46`**, both pods. Proved at the binary with a
  three-way control (the build sha PRESENT · yesterday's `0b185bad2` ABSENT · a nonsense sha
  ABSENT). The startup `build provenance` line had already scrolled past `--tail=20000` on both
  pods, so the log route does not work here.
- **Both our sites still LOCKED** (`locked_by` names this lane and the halt).
- **Directory: 25 active mortgage lenders, 17 savings providers.** Unchanged overnight.
- **`www` → apex still 301ing** (spot-checked 5 zones incl. the two that were resolver-stale).
- **`bugs_open/311` is NOT fixed** — `store_generated_component_action.go`,
  `create_tool_component_action.go` and `component_selector.go` are all unchanged since it was
  filed.

---

## 1. ⛔ THE HALT — still in force, and it is the owner's gate

**Owner, 2026-08-18:** *"Stop the builds until we sort out the classifier and which builder flow
we are using."* Implemented with `sites.locked_at`, which is exactly what
`build-pipeline-trigger.find_dispatchable_site` excludes on. **Queued work is preserved, not
cancelled.**

| site | locked | queued HELD | state |
|---|---|---|---|
| `adversecreditmortgage.co.uk` (build #1) | ✅ | 41 | dispatched, halted mid-build |
| `remortgagecalculator.uk` (pilot) | ✅ | 0 | **LIVE** at its own domain, but incomplete |

To resume (**only after §2**):
```sql
UPDATE sites SET locked_at = NULL, locked_by = NULL
WHERE domain IN ('adversecreditmortgage.co.uk','remortgagecalculator.uk');
```

## 2. THE TWO OWNER DECISIONS — everything waits on these

Write-up: `DECISION_2026-08-18_two_builder_flows_side_by_side.md`. RFC: `RFC_037`.

**(a) Which builder flow.** Flow A (seeded + hand-written mission, 45–60 min/domain ≈ 100 h
across the fleet) vs Flow B (prompt only, ~2 min, but the site gets no `evidence_base`, so
`loadEvidenceBase` returns nil and **every claims lane silently no-ops** — nothing checks a
single assertion, and the missing email makes the hallucinated-email check fail open,
`bugs_open/063`). The standing recommendation is **flow B + an automatic seed**.
> **⚠ NARROWED by owner ruling P11 (2026-08-18, in `REGISTER_positioning.md`).** The first
> `loanzy.uk` build was **cleared**: *"we shouldn't create accredited finance broker sites
> unless asked."* An auto-seed supplies the guards and **does not** stop the classifier or
> strategist adopting a regulated identity in the first place. Flow B therefore also needs a
> **prohibition on regulated-intermediary positioning** (or a check that refuses the plan).
> That is new work and it is **not costed** in the decision doc.

**(b) RFC_037 — the classifier reads the register.** Owner chose option 2. Filed, not built.
Measured case: 7 finance sites → 2 distinct classifications, `industry` null on all 7.

**Settled, so it need not be re-litigated:** the "two flows" are ONE flow. Measured at the
handler, not asserted from the script — all three sites produced identical item types handled by
identical agents (`build-site-planner`, `site-design-planner`, `page-build-handler`,
`image-build-handler`) in the same order. `pageflow-builder` is a live agent and the classifier
still writes `recommended_builder: "pageflow-builder"`, but **no Go code reads that field**
outside doc-comment examples of the generic dispatch helper. It is a leftover from the older
intake route.

## 3. 🧱 THE WALL IN FRONT OF THE FLEET — `bugs_open/311`, and it grew overnight

**A site cannot get a tool whose `function` name another site created first.** Selection keys on
`section_type`; storage keys on `function`. A component with one and not the other is invisible
to the selector ("build a new one") and matched by the writer ("this is a regeneration"), whose
field-contract guard then correctly refuses to strand the incumbent site's `content_data`. The
requesting site's section is left with `component_id=''` and **the page builds, deploys and
serves one section short, `status='active'`, with no tell in the artefact.**

- `090` verdict **CONFIRMED, first iteration** (`8aa2e283-129f-41d1-93a0-6dcacbbabeae`).
- **Three sites**: `remortgagecalculator.uk`, `loancalculator.co.uk`, `loanzy.uk`.
- **Overnight contribution from the `loanzy_uk_example_site` lane: 7 of 7 tool sections lost on
  a greenfield build**, and the consequence measured at the served page —
  `https://loanzy.uk/tools/loan-comparison-calculator/index.html` returns 200 with **zero
  `<input>` elements**. A calculator page with no calculator, live on the public web.
- **Class is 26 wide** (active base `section` components with no `section_type`), plus 79
  `tool`-level rows the section selector cannot see at all.
- The incumbent is `loanandmortgagecalculator.co.uk` in every case so far.

**⚠ THE CROSS-LANE FACT, established this morning — read before anyone "fixes tools".**
`architecture_review/RFC_036` is the SAME design fact (function is fleet-wide, identity is
per-site) at the **tool** level, and it already carries a written path (§9.3) plus an owner
direction. **They are different defects with the same remedy shape**, measured:

| | 311 | RFC_036 |
|---|---|---|
| level | `section` | `tool` |
| what refuses | field-contract guard, `store_generated_component_action.go:397-412` | unique index `idx_cc_tool_function_unique` |
| writer | `store_generated_component` | `create_tool_component` |
| presents as | work item **`failed`** | work item **`complete`** |

RFC_036 §9.3's change (set `forked_from` to the incumbent's id before the INSERT) **is 311's own
fix candidate 1, written out for the neighbouring writer** — but as scoped it fires only when a
**library (`tool`-level)** row claims the function, so **it would leave 311 entirely live.** Both
are shared-seam changes on the component write path: **council gate + chassis roll**, and one
submission covering both writers is cheaper than two rounds. Cross-references are now in both
files (they cited each other nowhere before today).

**Nothing here is built. This is a precondition for the fifty, not an improvement to them.**

## 3b. 🔗 THE PILOT IS LIVE WITH TWO DEAD LINKS ON EVERY PAGE — `bugs_open/260`, and it is ONE defect not two

Answered by an inbound note from the `loanzy_uk_example_site` lane
(`NOTE_2026-08-18_from_loanzy_lane_your_end_leak_todo_is_already_filed.md`, no reply needed) and
then **measured on our own live site 2026-08-19**.

The pilot's four `needs_page` items stuck in `needs_human_review` all failed `validate_content`
with **`20 blockers, 0 errors`** — that is `bugs_open/260`, the `{{end}}` template leak, already
filed with a proven root cause. **Do not re-file it.** Two pages were never built
(`six-month-checklist`, `what-your-number-means`): both have `pages` rows with `status='active'`
and **zero `page_components`**.

**The live cost, read from the served HTML:** `/`, `/about.html` and `/next-steps.html` **each
link to both blocked pages**, and both targets return **404**. They are nav links, so the damage
is site-wide rather than one stray reference, and nothing on any page shows a reader that
anything failed.

**This collapses two TODOs into one.** `HANDOFF_2026-08-18` carried the `{{end}}` blockers and
the dead-internal-links finding as separate items; on this site the links are dead *because* the
pages were blocked, so **repairing 260 removes both** and no link-level work is needed.
Contributed back as §12 of the bug file, which still says "no live damage" in its headline — now
wrong on two live sites, though we did not edit another lane's headline.

## 4. ✅ DONE — do not redo

- **`www` → apex, fleet-wide** (owner deployed the worker 2026-08-18 20:02:37Z; fan-out ran).
  **28 DNS records + 7 routes, 0 failed; 36/36 applicable zones verified 301ing**, path and query
  preserved. 3 deliberately skipped: `idea.uk` and `relojistas.com` (**no route to the worker —
  a proxied A there is a 522 black hole**), `webdesign.uk` (deliberate 302 elsewhere).
  Scripts: `scripts/cloudflare/worker.js`, `scripts/cloudflare/add_www_redirect.sh`.
- **Directory widened: 2 → 25 lenders** — `471_widen_finance_directory_discovery.sql`
  (**cite the FILENAME: a different `471_floor_held_remedy_partitions_failures_first.sql` was
  applied by another session two minutes earlier; the ledger keys on filename, so both are live
  and a bare "471" resolves to neither**). Rollback sidecar exists.
- **Owner ruling P11** recorded in `REGISTER_positioning.md`.
- **The pilot is live** at its own domain (Nominet delegation landed).

## 5. THE PATH FROM HERE — in order, with the reason for the order

1. **Answer §2 (a) and (b).** Owner decisions; nothing else is safely startable. (a) is now a
   three-part choice — flow, auto-seed, *and* the regulated-identity prohibition.
2. **Decide who builds the 311/RFC_036 fix, and as one submission or two.** It needs the gate
   and a roll. Until it lands, **every site built on a shared calculator name ships that tool
   hollow** — so building the fifty first means fifty sites to repair afterwards.
3. **Refresh the pilot's lender page** (see §6 — it is the one thing waiting on a single word),
   and note that the pilot ALSO needs `bugs_open/260` before it is presentable: two of its six
   pages do not exist and every serving page links to them (§3b).
4. **Lift the halt** and let build #1 continue from its 41 held items.
5. **Then wave 1 of the fifty**, one at a time and supervised, per
   `PLAN_2026-08-18_first_50_build_order_FOR_APPROVAL.md` (approved).

## 6. THE ONE OPEN ITEM WAITING ON A SINGLE WORD

**The pilot's lender page serves 2 lenders; the register holds 25.** Confirmed still stale at
the artefact this morning — `https://remortgagecalculator.uk/mortgage-lenders.html` names only
Family Building Society and Mansfield Building Society. It refreshes on the directory publish
path, but the site is locked under §1 so nothing will do it. A single page refresh is not a
build, but it is a change to a live site under the owner's halt — **his call.**

## 7. TRAPS — each one cost a cycle; do not re-pay them

- **A parked domain returns 200 on EVERY path.** Verify DNS by reading the **body**.
- **A newly created worker route 522s for the first few requests**, and 522 is exactly the
  signature of "no worker, dead origin". Retry three times before believing it.
- **Your own resolver's negative DNS cache outlives a record you just created** — `Could not
  resolve host` is indistinguishable from "it was never created". Ask authoritative DNS
  (`https://1.1.1.1/dns-query?name=…&type=A`) and prove it with `curl --resolve`.
- **A worker deploy's PUT response returns `bindings: []` even on success** — identical to the
  credential-stripping outage. Verify by fetching real pages, never from the response.
- **Never verify a fix by grepping the binary for its commit sha** without controls; and the
  `build provenance` startup line **scrolls** — on a busy service it is gone within hours.
- **Seeded `banned_claims` fail silently when double-escaped** (valid regex, matches nothing,
  count-based verify passes). Probe in Go.
- **Cost measured mid-build reads ~70% low** (`collected_data` fills in as runs progress).
- **Discovery queries must stay < 200 bytes** — `web_search` drops a ≥200-char query and the
  failure names config keys, not length.
- **A migration NUMBER identifies nothing** — two different `471`s exist. Ask the ledger by
  exact filename.

## 8. Cost baseline

**Text: $3.81/domain today · $4.83 from 2026-09-01** (Sonnet 5 intro rate ends). 73 calls,
663,759 in, 184,596 out, three agreeing runs. **Images unmeasured** — no cost column exists.
At 30 images/site, **$5.01–$11.31/domain**; fleet of 140 ≈ **$700–$1,600**.
⚠ *An account-wide Anthropic usage limit was hit fleet-wide at 10:24Z on 2026-08-19 (another
lane's council run returned no verdict because of it). If runs fail in a way that looks like a
defect, check that first.*

## 9. Files of record

**Cold start:** this file → `SUMMARY_2026-08-19_…` → `README_where_we_are.md` (owner's log,
plain prose, newest at bottom) → `NOTES_portfolio_positioning.md` (evidence, newest at bottom).
**Decisions:** `DECISION_2026-08-18_two_builder_flows_side_by_side.md` · `RFC_037` ·
`REGISTER_positioning.md` (P11, B8/B9/I10).
**The wall:** `bugs_open/311` · `architecture_review/RFC_036`.
**Build order:** `PLAN_2026-08-18_first_50_build_order_FOR_APPROVAL.md` (approved) ·
`PLAN_2026-08-12_fleet_buildout.md`.
**Ops:** `RUNBOOK_dns_pointing_a_domain_at_the_serving_worker.md` (the ✅ DONE section) ·
`scripts/cloudflare/` · `sql_for_agents/471_widen_finance_directory_discovery.sql`.
**Register:** `docs026_concept_register/register/directory-pipeline.md` (DIR-001, intake sizing).

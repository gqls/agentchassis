# HARVEST 01 — the vonc provocations pilot (2026-07-26)

The first harvest. Owner ruling of 2026-07-24 gated it on the vonc pilot being live end to
end through tools-api (the static-cut alternative was rejected). **That gate is now lifted**,
and this file records what came out of the live thing, what it changed in the design, and
what P2 must therefore build differently.

Four candidate entries: `harvest/entries/`. They are drafts — nothing is applied, no table
exists yet.

## 1. The gate — verified first-hand this session, not taken from another session's log

| check | result |
|---|---|
| `GET https://vonc.com/provocations/index.html` | **200**, 23,670 bytes |
| `GET https://vonc.com/data/provocations.json` | **200**, 9,797 bytes, **byte-identical** to `gauntlet_dead_cta/p4_sources/provocations.json` (`diff -q`) |
| archive DOM contract on the live page | `data-component="provocations-archive-list"` present; `data-archive-template` row present; 3 × `__item-title` / `__item-teaser` / `__item-date` slots |
| loader shipped in the live bundle `/assets/js/snippets.js` | present by four **discriminating** strings — ones the loader itself creates, not ones it merely uses: `provocations-archive-detail-style`, `__item--linked`, `__item--static`, `entriesBySlug` |
| the journeys themselves | the owning session's browser verification, desktop + mobile, **72 of 73** — `gauntlet_dead_cta/p4_sources/verify_live_2026-07-26.txt`. The single failure is upstream 503s, filed as `bugs_open/083`; it does not touch any behaviour harvested here |

The pilot's own state, for the record: `tools.apis.uk` live on the island VM, the debate
engine returning real AI content in 8–23 s, the whole round-trip working for a real visitor.
That belongs to the `gauntlet_dead_cta` workstream — this workstream **consumes** it and
drives nothing there (PLAN §6).

## 2. What was harvested

| entry | kind | from | the clause that makes it worth a register row |
|---|---|---|---|
| `CC-001 feed-driven-teaser-list` | component-contract | `provocations-archive-list` (`70d6662a`) + its loader | whether a row is a **control** is decided by the DATA, not the template — a row with nothing behind it is stripped of the template's `href="#"` and left out of the tab order |
| `MJ-001 teaser-detail-deeplink` | micro-journey | the same, live Journey B | the full text opens **in place** while the address bar gains a parameter that reproduces the state on a cold load |
| `CC-002 feed-promised-cta` | component-contract | `provocation-card` (`6163ff14`) + `lobby-grid` (`9304f14d`) | label and destination are **one record**; a record with no destination renders no control |
| `MJ-002 timed-remote-challenge-loop` | micro-journey | `gauntlet-interface` (`5da50747`) against `tools.apis.uk`, live Journey C | every visible state change is a consequence of a **real response** — never of the click that requested it |

All four components carry `usage_count = 0` in `content_components` (queried live). They are
bespoke, single-site builds. **That is the register's case in one number**: the components do
not repeat, but the clauses above already do — `CC-002` was harvested from two different
components on the same site, and `MJ-001`'s open/close/deep-link shape is the archive
behaviour every news, blog and directory site in the estate re-invents.

## 3. What the live implementation changed in the design

Ten corrections. Each is a place where the design as written on 2026-07-24 could not express
something the working code does. This is what harvest is *for* — the corrections are the
return on it, not a cost of it.

### 3.1 A contract needs to say what is NOT a control — `states`
The design shaped `contract` as a list of triggers. The live loader's most valuable clause is
the opposite: *a row whose record has no case gets no href, no tabindex, no handler.* A
trigger list cannot hold that, so a pattern built from the design as written would have been
silent on exactly the thing (`bugs_open/023`, the 2026-07-22 gauntlet correction) this
workstream exists to prevent. → new `states` array on both entry kinds.

### 3.2 The criteria vocabulary cannot assert an attribute
`selector_exists`, `selector_count`, `interaction`, `asset_loads`, `page_status_ok`
(`platform/orchestration/actions/discovery_checks/check_tool_acceptance.go:343-380`) plus
`no_horizontal_overflow` at Tier 4
(`internal/adapters/browserrunner/run_checks_action.go:374-467`). None of them can say
*"this element has no `href`"* or *"this `href` is not `#`"*. So the anti-dead-control clause
— the one in §3.1 — **cannot be checked at all** by the platform's own criteria today. It is
checkable only by the separate dead-control sweep, after the page is built, on a different
schedule, by a different owner.

### 3.3 There is no navigation step, and `page_status_ok` takes no URL
Interaction steps are `fill | click | select` (`run_checks_action.go:772-776`). A check runs
against one URL, decided by the run. Two consequences: **(a)** the deep link — the entire
point of `MJ-001` — cannot be asserted, because asserting it means loading a *different* URL
and looking; **(b)** `CC-002` cannot check that the promise is kept, because that means
fetching the destination. A promise ledger the platform cannot mechanically check is prose.

### 3.4 Nothing can assert an EMPTY region
`interaction`'s `expect` is `{selector, text_matches}` and fails when the selector is absent
(`runInteraction`, `run_checks_action.go:493-519`) — there is no "must be empty / must be
gone". The live loader *empties* the detail region on close precisely because `innerText` on
a `display:none` element falls back to `textContent`, so a hidden-but-populated region reads
as content and an acceptance check passes without the interaction having happened. That trap
bit a real check on 2026-07-26. The countermeasure is in the code and unassertable by the
platform.

### 3.5 No waits, no ordering, no retries — async patterns are unassertable
`stepDelay = 300 ms` (`run_checks_action.go:199`) and `runInteraction` asserts immediately
after the last step. The live AI calls take **8–23 s**. So the approved EXPERIENCE_PLAN's own
`gauntlet_position_flow` and `gauntlet_defend_flow` checks **would fail a correct page** — and
the tempting repair (make the page paint placeholder text) would make them pass *with the
engine switched off*. Fix the harness, never the page. Ordering is missing too: "exactly one
progress marker after the first response" needs a check that runs *after* another check, and
there is no such thing. That clause is the whole honesty rule of `MJ-002`.

### 3.6 Where a shared clause lives — the composition rule
The openable/inert split could sit in the component contract or in the journey. Two accounts
of one clause is the drift class this workstream keeps filing bugs about, so: **a render-time
property belongs to the `component-contract`; the journey references it** and asserts only
what activation *does*. → new `requires_component_contract` on micro-journeys; `MJ-001`
inherits `CC-001`'s checks rather than restating them.

### 3.7 Degraded states are part of the contract — and are untestable today
Both live components declare what happens when their dependency is absent (feed missing →
shell and empty state stand; engine down → honest offline message, clock does not start, no
marker set, *and the visitor's typed text is preserved*). The design had no field for this,
yet it is where dishonesty enters a page: a failure path is exactly where a build is tempted
to fake success. → new `degraded_states`. **Note the second-order finding**: no check can
induce a failing dependency, so this clause can only be observed opportunistically — as it
was on 2026-07-26, when the engine flaked mid-verification and the page told the truth.

### 3.8 A destination is not always another page — `self-state`
Destination roles were specified as page roles. `MJ-001`'s destination is *a state of the
same page*, addressable by a URL parameter. Binding that to a page id would be a lie. → role
kind `self-state`, with `addressable: true` and the parameter as a binding.

### 3.9 A pattern's honesty depends on data it does not own — `data_contract`
`MJ-001` is only honest if the feed carries a per-entry body and key; `CC-002` is only honest
if label and URL travel together; `MJ-002` renders engine output verbatim. The 2026-07-24
design had `binding_schema` for selectors and pages and **nothing for the shape of the data**.
The approved EXPERIENCE_PLAN, written by the council, has a whole section for it (§3 Data
contracts) — the register would have been the poorer of the two. → new `data_contract`.

### 3.10 Pattern #1's name was wrong
`design/taxonomy_seed.md` called it `teaser-detail-related` and described a third leg —
"related links and tools" from the detail. **The live implementation has no such leg.** The
onward links live in the feed's `today`/`arena` CTAs (which is `CC-002`), not in the detail
region. Harvested honestly: the entry is `teaser-detail-deeplink`, and "related" is a
separate, unbuilt continuation. This is the first time the bottom-up rule paid: authored
top-down, the register's first entry would have carried a leg no implementation has.

### 3.11 Write-time validation must ask "can this be executed?", not only "does it parse?"
`design/criteria_template_schema_v1.md` specified write-time validation as **parse +
placeholder closure**. The gauntlet's criteria parse perfectly and are still unexecutable
(§3.5). A council approved them. → validation must also check each check against the
**declared capabilities of the tier that will run it**, and mark the unexecutable ones
`deferred` at write time rather than letting them read as green-capable. The register is the
right place for this because it holds the template once, not once per site.

## 4. The first fork — proof the base/binding separation works

`MJ-001` bound to vonc (`9ec3b9ee-5b08-461b-b4f8-9e1e03579c74`), page `provocations-index`.
A draft `site_experiences` row; every concrete value lives here, none in the entry:

```json
{
  "site_id": "9ec3b9ee-5b08-461b-b4f8-9e1e03579c74",
  "pattern_name": "teaser-detail-deeplink",
  "instance_key": "provocations-archive",
  "bindings": {
    "list_section": "[data-component=\"provocations-archive-list\"]",
    "list_container": ".provocations-archive__list",
    "item_template": "[data-archive-template]",
    "item_openable_selector": ".provocations-archive__item--linked",
    "item_static_selector": ".provocations-archive__item--static",
    "detail_selector": ".provocations-archive__detail",
    "detail_close_selector": ".provocations-archive__detail-close",
    "empty_state_selector": ".provocations-archive__empty",
    "list_page": { "url": "/provocations/index.html" },
    "list_section_type": "provocations-archive-list",
    "feed_path": "/data/provocations.json",
    "detail_param": "entry",
    "sample_item_key": "ai-never-funny-on-purpose"
  },
  "status": "proposed"
}
```

Two things this makes visible. **First**, the entry is genuinely site-agnostic — every vonc
string is in the fork, and a news site would bind the same pattern to entirely different
selectors. **Second**, `sample_item_key` is a binding, not a constant: the criteria need an
item that is openable *today*, and the feed's openable set changes. A base entry carrying
`ai-never-funny-on-purpose` would be the `bugs_open/045` static-fallback mistake in a new
place.

## 5. What P2 must build differently

1. `experience_patterns` gains `states`, `data_contract`, `degraded_states`,
   `entry_points`, `requires_component_contract` (jsonb, defaulting to empty) — §§3.1, 3.6,
   3.7, 3.9, and `MJ-001`'s deep-link entry point.
2. Destination roles admit `self-state` with an `addressable` flag and a parameter binding
   (§3.8).
3. Write-time validation = parse + placeholder closure **+ tier-capability check**, with
   unexecutable checks recorded `deferred` (§3.11).
4. **Criteria vocabulary extensions are a real dependency, not a nice-to-have.** Four of the
   clauses that make these patterns worth having cannot be asserted today: attribute
   presence/shape (§3.2), navigation and cross-page status (§3.3), empty-region (§3.4),
   waits/ordering/retries (§3.5). They belong to `check_tool_acceptance.go` and
   `run_checks_action.go` — a **separate change-set, separate owners** (the browser-runner
   half overlaps `gauntlet_dead_cta` P5, which needs the wait/poll for its own reasons).
   P2 should ship the register with these marked `_unsupported` in the templates and NOT
   silently drop them: a template that quietly omits its unassertable clauses is how a
   pattern comes to look fully checked when its most important clause is not checked at all.
5. **064 is CLOSED and live** — `bugs_closed/064…` (commit `eb81de7b5`): the single-sourced
   `validDocSubjectTypes` shipped in v1.0.1156 on 2026-07-25 with both branches proven by live
   orchestration runs; I re-confirmed the strings in today's v1.0.1167 binary. P2's
   subject-type work is therefore **one Go list entry plus the migration**, and the
   migration-lockstep test already fails the build if a migration lands without the Go entry.

## 6. What the register would have been worth here — stated honestly

**[RETROSPECTIVE / INFERRED — the register did not exist, so this is not a measurement.]**
The first ever approved EXPERIENCE_PLAN (`doc_plans`, subject `experience`/`vonc-spark-game`,
current version ~14 KB, approved 2026-07-25 after five REVISE rounds, a designed escalation,
and a re-fire that converged in one) contains **four journeys**. All four map onto the four
entries harvested above: A and D onto `CC-002`, B onto `CC-001` + `MJ-001`, C onto `MJ-002`.

What that supports is a modest, checkable claim — *the first plan the council ever approved
contained no journey shape that was not reusable* — and one design consequence: the register's
job is to arrive as the plan's **starting draft**, so the council argues about the site-specific
binding and glue rather than re-deriving that a dead-end row must not be a link. What it does
**not** support is any claim about rounds saved. That number can only come from the first plan
written *with* the register, and it should be measured then, not asserted now.

## 7. Coordination, and what was deliberately not done

- The pilot, `tools-api`, the island and `bugs_open/083` belong to `gauntlet_dead_cta`.
  Everything above is **read-only** consumption of their live output plus my own re-checks.
  Nothing was changed on vonc, no dispatch was fired, no work item was created.
- The criteria-vocabulary extensions (§5.4) are **not** started and are not this workstream's
  to start alone — the browser-runner half is already on `gauntlet_dead_cta`'s P5 list.
  Filed here as a named dependency so P2 does not quietly assume it away.
- No table was created, no migration written, no council submitted. P2 remains gated on the
  owner's go.

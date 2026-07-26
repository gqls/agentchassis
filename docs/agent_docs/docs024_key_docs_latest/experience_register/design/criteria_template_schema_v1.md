# Criteria template schema v1 — design artifact (P2)

> **CORRECTED 2026-07-26 by HARVEST 01 — read this before the body below.**
> Writing criteria for two real live patterns falsified two things in this file.
>
> 1. **The v0 inventory below is incomplete**: Tier 4 also supports
>    `no_horizontal_overflow` (`internal/adapters/browserrunner/run_checks_action.go:422`),
>    and interaction steps are limited to `fill | click | select` (`:772-776`).
> 2. **"Non-goal: no new check-type invention beyond `journey`" is wrong.** Four extensions
>    are load-bearing, not optional — without them the clauses that make a pattern worth
>    registering cannot be asserted at all:
>    - **attribute assertions** (`attribute_absent`, `attribute_matches`) — the
>      anti-dead-control clause (*this row has no `href`, no `tabindex`*) is unassertable
>      today, which is precisely `bugs_open/023`'s class;
>    - **a navigation step + cross-page status** (`goto`, `page_status_ok` with a URL) —
>      without them a deep link cannot be checked and a promised destination cannot be
>      followed;
>    - **an empty/absent assertion** — `expect` fails when a selector is *absent* and has no
>      "must be empty"; the `innerText`-on-`display:none` trap makes this a real false pass;
>    - **waits, ordering and retries** (`expect_within_ms`, `after`, `retries`) —
>      `stepDelay = 300ms` (`:199`) versus live AI calls at **8–23 s**.
> 3. **Write-time validation needs a third rule**: parse + placeholder closure **+ a
>    tier-capability check**. The gauntlet's criteria parse perfectly, were council-approved,
>    and still cannot be executed correctly. Unexecutable checks must be marked `deferred`
>    when written, not discovered at run time.
>
> Evidence and the full ten corrections: `harvest/HARVEST_01_2026-07-26_vonc_provocations.md`
> §§3.2–3.5, 3.11. The extensions belong to `check_tool_acceptance.go` /
> `run_checks_action.go` — a **separate change-set with separate owners**; the browser-runner
> half overlaps `gauntlet_dead_cta` P5. P2 ships templates that mark those checks
> `_unsupported` rather than quietly omitting them.

Formalises the acceptance-test side of register entries (owner ruling 2026-07-24).
Extends the existing criteria schema v0 (`check_tool_acceptance.go:276-298` structs;
check types `selector_exists`, `selector_count`, `interaction`, `asset_loads`,
`page_status_ok` statically checkable; others Tier-4-only) with **placeholders** and a
**journey check type** (deferred until browser-runner T5.1 exists).

## Why formalise (the v0 failure classes this closes)

v0 criteria are parsed at read time only. Live consequences, all sighted in
`travelling_docs/README_where_we_are.md`:
- **Stale criteria** (3 sightings): selectors assert markup the page no longer ships.
- **`-EDIT` placeholders**: silently skipped by Tier-2, so a plan looks green while its
  real interaction was never asserted.
- **Unclosed fence** → extractor returns "" silently; `criteria_unparseable` surfaces only
  when a sweep runs.

## Template shape

Same document shape as v0 (`{"profiles":[…],"checks":[…]}`), stored as jsonb in
`experience_patterns.criteria_template` (not a markdown fence — the fence stays the
convention for travelling-doc PLANs; the register's copy is machine-validated at write).

Placeholder syntax: `{{binding.<name>}}` and `{{binding.<name>.<field>}}` — e.g.
`{{binding.card_selector}}`, `{{binding.detail_page.url}}`. Every `<name>` MUST be declared
in the entry's `binding_schema`.

```json
{
  "profiles": ["desktop", "mobile"],
  "checks": [
    {"id": "card-present", "type": "selector_exists",
     "selector": "{{binding.card_selector}}"},
    {"id": "read-more-expands", "type": "interaction",
     "steps": [{"action": "click", "selector": "{{binding.read_more_selector}}"}],
     "expect": {"selector": "{{binding.expanded_selector}}"}},
    {"id": "card-to-detail", "type": "journey",
     "steps": [{"action": "click", "selector": "{{binding.card_selector}}"},
                {"action": "expect_url", "value": "{{binding.detail_page.url}}"}],
     "tier": 4}
  ]
}
```

## Binding schema shape

```json
{
  "card_selector":     {"type": "selector"},
  "read_more_selector":{"type": "selector"},
  "expanded_selector": {"type": "selector"},
  "detail_page":       {"type": "page", "role": "entity-page", "of": "card.entity"}
}
```

`type: page` placeholders bind to a real `pages` row (page_id + url snapshot);
`type: selector` placeholders bind to concrete selectors on the site's components.

## Validation moments

1. **Write time** (new discipline — the register's write path, unlike `write_doc_plan`,
   validates): document parses against this schema; every placeholder used in
   `criteria_template` OR `contract` is declared in `binding_schema` (closure); every
   declared placeholder is used (or carries `"unused_ok": true` with a reason); no
   literal URLs or site-specific selectors in the template (base entries are invariant).
2. **Bind time** (fork creation → `site_experiences.status='bound'`): every placeholder
   bound; `type: page` bindings resolve to live `pages` rows whose `page_type` matches the
   declared role; `type: selector` bindings pass the v0 **anchor rule** against the site's
   actual components (validate the leftmost id/class/tag token — confirm, never refute).
3. **Run time**: Tier-2 executes the statically-checkable subset against deployed HTML
   (existing `evaluateStaticCriteria` machinery, bound values substituted). `type: journey`
   checks are **deferred** — they require the browser-runner T5.1 navigate/persistent-
   context extension (not started; G9: each URL is a fresh browser today; additive-path
   ruling of 2026-07-18 applies). Until then journey checks are recorded, reported as
   `deferred`, and never counted as passes. First green full run flips the entry
   `approved → proven` and the fork `bound → verified`.

## Non-goals (v1)

- No new check-type invention beyond `journey` — reuse v0's vocabulary.
- No write-time validation retrofit for existing tool-PLAN fences (separate concern;
  belongs to travelling_docs if taken up).
- No LLM-judged outcomes: `outcome` prose in the contract is for reviewers and harvest;
  checks assert only mechanically-observable effects.

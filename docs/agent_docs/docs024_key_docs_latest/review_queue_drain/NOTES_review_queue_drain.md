# NOTES — review queue drain (`bugs_open/033`)

Append-only, newest at the bottom. Technical log: evidence, commands, what the
system actually said, and every misstep.

---

## Turn 1 — 2026-07-25, opening the drain half of 033

### Coverage check first (nobody else on it)

- `scripts/who-owns.py 033` → no OWNED verdict; last commits on the file are the
  2026-07-20 grounding pass and a 2026-07-22 cross-reference from `bugs_open/054`.
  Named as *citing* it: `reasoning_dataset`, `cta_link_integrity`, `robot_hands`.
- `site_work_items` open items matching review-queue work: **0 rows**.
- `needs_diagnosis` queue: 5 open rows, none about this.
- `OPEN_THREADS_RESTART_LIST.md:58` — "Admin dashboard | 033 fixed+live; **open:
  the drain (split it)**". Consistent: the drain is unclaimed.

### Where the queue actually is (live, 2026-07-25)

```
status='needs_human_review'                 : 370   (292 filed, 303 grounded)
  build / content / maintenance             : 224 / 145 / 1
approved_by IS NOT NULL, all 5,600+ rows    : 0
result->>'resolved_by' = 'admin'            : 0
```

Five days after the surface was fixed and made reachable (v1.0.1141, VPN): **+67
in, 0 out.**

### The measurement that decided the design

```sql
-- items whose page has been redeployed since the item was filed
SELECT w.item_type, count(*) AS n,
  count(*) FILTER (WHERE EXISTS (
    SELECT 1 FROM pages p
    WHERE p.site_id=w.site_id AND p.name = w.spec->>'page_name'
      AND p.deployed_at IS NOT NULL AND p.deployed_at > w.created_at)) AS page_deployed_since
FROM site_work_items w
WHERE w.status='needs_human_review' AND w.spec->>'page_name' IS NOT NULL
GROUP BY 1 ORDER BY 2 DESC;
```

**321 of 370.** The queue is overwhelmingly findings about page states that have
since been rebuilt — and nothing re-checks any of them.

### Ghosts proven, not inferred

Two `unresolved_cta` items on `leopardessconsulting.co.uk/how-we-work`, parked
2026-07-10, say the hero and call-to-action have no destination for
`cta_url`/`secondary_cta_url`. The page redeployed 2026-07-18:

```
hero            cta_url=/tools/password-entropy.html  secondary_cta_url=/services.html
call-to-action  primary_cta_url=/tools/password-entropy.html  secondary_cta_url=/services.html
```

Every field both items call missing is populated. Both are ghosts.

Scaled across the class (naive `slot_name` join, so a floor not a ceiling):

| class | parked | provably resolved | still holds | not determinable |
|---|---|---|---|---|
| `unresolved_cta` | 68 | 39 | 29 | 0 |
| `required_fields_missing` | 45 | 10 | 4 | 31 |
| `needs_section_data` | 45 | 2 | 2 | 41 |

### MISSTEP 1 — I keyed the first re-validation queries on `spec.component_id`

First pass reported "**30 of 30** `needs_section_data` components gone, **11 of
45** `required_fields_missing` gone" and I nearly wrote that up as "the queue
points at deleted components". It does not. `page_components.id` is **not stable
across re-renders** — a fact already in my own memory notes from the robot-hands
workstream, which I did not apply until the number came out absurd. Re-keyed on
`(page_name, slot_name)` and the sections are all still there.

Caught by: the number being *too* clean. 30/30 is not a defect signature, it is a
join bug. **The cheap check that would have caught it earlier:** print the
page's actual slot list next to the wanted slot, which is what I eventually did:

```sql
SELECT w.spec->>'page_name', w.spec->>'slot_name',
       (SELECT string_agg(pc.slot_name, ', ' ORDER BY pc.position)
        FROM page_components pc JOIN pages p ON p.id=pc.page_id
        WHERE p.site_id=w.site_id AND p.name = w.spec->>'page_name') AS actual_slots
FROM site_work_items w WHERE w.status='needs_human_review' AND w.item_type='required_fields_missing';
-- want hero => actual "hero, article-body, call-to-action"  — it is right there
```

### MISSTEP 2 — my own "slot_gone" count was also wrong, differently

Second pass counted `content_data IS NULL` as "slot gone". It is not: the
component **row exists** and carries no `content_data` because it renders from a
template / DERIVED source / static fallback. 31 of the 45
`required_fields_missing` items are in that state. That distinction became a
design decision rather than a footnote: those return `unknown`, not `resolved`.
Judging them on `content_data` would be judging them on evidence that is not the
rendering source.

### CORRECTION to 033's own fix candidate A

033 says: *"The one genuine automated consumer, `reconcile_section_data_action.go`
(re-opens `needs_section_data` when query-sourced data later resolves — **48
items of the queue**), is registered as an action but wired to 0 live agents"*,
and candidate A is "wire it". Measured live:

```sql
WITH s AS (SELECT (SELECT bool_and(m->>'source' LIKE 'query.%')
                   FROM jsonb_array_elements(spec->'missing') m) AS all_query
           FROM site_work_items
           WHERE status='needs_human_review' AND item_type='needs_section_data'
             AND jsonb_typeof(spec->'missing')='array')
SELECT all_query, count(*) FROM s GROUP BY 1;   -- f | 30   (no 't' row at all)
```

**0 of 45** parked `needs_section_data` items have all-`query.*` sources; 30
carry `site_specs.*`/`site_assets.*` and 15 carry `missing: null`.
`ReconcileSectionDataAction` requires `strings.HasPrefix(m.Source, "query.")` for
*every* missing field or it skips the item. Wiring it today re-triggers **zero**
pages. The action is not wrong — the population it was built for is not the
population in the queue. Recorded in PLAN as a correction; `WRONG_CALLS.md` row
added.

### The structural claim, filed for diagnosis before being asserted

`insertWorkItem` (`load_work_item_actions.go:1111`) inserts findings with
`ON CONFLICT (site_id, item_key) … DO NOTHING`, and `RunDiscoveryChecksAction`
(`discovery_checks.go:166`) counts the suppressed insert into a local `skipped`
tally and nothing else. So a **re-confirmed** finding and an **abandoned** one
are byte-identical on the row.

That is durable and structural, and it is exactly the shape CLAUDE.md says to
file before asserting. Filed 2026-07-25, corr
`c19ed5b2-6d53-492a-af91-e78e175591d5`. **Verdict pending — the claim is marked
[FILED, UNCONFIRMED] wherever it appears until it lands.** The fix does not
depend on it: the revalidator re-derives the truth from deployed state rather
than trusting any row-level signal, which is why it works whether or not the
re-confirmation gap is confirmed.

### What shipped this turn

- `platform/orchestration/actions/revalidate_review_queue_action.go` — the sweep.
- `platform/orchestration/actions/revalidate_review_queue_test.go` — tests built
  from real live specs, including the 15-row `missing: null` case.
- `platform/orchestration/actions/section_editor_actions.go` —
  `loadPageComponentBySlot` split into `loadPageComponentBySlotRO` + the
  backfill. **Reason: the existing function WRITES** (it backfills
  `page_components.slot_name` on a fallback match), and a `dry_run` sweep whose
  contract is "change nothing" cannot call that.
- `platform/orchestration/actions/registry.go` — registration.
- `seed_review_queue_revalidator.sql` + `TRIGGER_revalidate_review_queue_v1.sh`.

`go test ./platform/orchestration/actions/` → **ok**. `go build ./platform/...
./internal/...` → **ok**. (`go build ./...` fails in `cmd/reasoningset` on
another session's uncommitted WIP — `declared and not used: planJoined` — not
this change; `go vet` also reports a pre-existing `unreachable code` in
`load_component_library_actions.go:207`.)

Council submitted: corr `ccba9c51-9bd5-4f1f-840c-ddd9e84a7bbe`.

### LANDMINE recorded during design, not after

`reconcile_superseded_reviews_action.go:98` computes "parked since" as
`GREATEST(wi.created_at, COALESCE(wi.updated_at, wi.created_at))`. If this sweep
bumped `updated_at` on every item it stamps, it would push that boundary forward
on each run and **hide genuinely superseded pairs from the other sweep**. So the
non-closing path deliberately does not touch `updated_at`; the timestamp lives in
`result.revalidation.at`. Two sweeps over one table, and the second one's write
would have silently blinded the first.

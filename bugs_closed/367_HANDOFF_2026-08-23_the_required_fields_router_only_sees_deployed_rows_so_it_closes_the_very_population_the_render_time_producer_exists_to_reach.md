# 367 — the `required_fields_missing` router only sees DEPLOYED rows, so it closes as "stale" the very population the render-time producer exists to reach

**Filed 2026-08-23** by the `bugfix_342_absent_required` lane, found while proving 342's own
routability fix. **Status: OPEN, UNOWNED.** The mechanism belongs to `bugs_closed/277`'s router,
which is closed — hence a new file rather than a contribution.

> # ✅✅ CLOSED 2026-08-24 — FIXED, LIVE, AND OBSERVED ON A REAL ITEM IN PRODUCTION.
> The item this bug was filed about — `562788c3`, closed `complete` by the defect on
> 2026-08-23 17:09Z — was re-opened on 2026-08-24 (the finding was still true: `headline`
> and `trust_note` still absent, component still `pending`, still 9,220 chars, untouched
> since 2026-07-17) and the live router picked it up and **parked** it.
>
> **The two orchestration rows sit side by side and are each other's control** — same item,
> same component, before and after:
>
> | orchestration | when | `route` | `target_state` | `component_id` | `html_len` |
> |---|---|---|---|---|---|
> | `ab2cf74e` | 2026-08-23 17:09Z | **`stale`** → closed `complete` | *(none)* | *(empty)* | 0 |
> | `e2a6bb94` | 2026-08-24 16:08Z | **`target_not_dispatchable`** → parked | `pending` | `0a1498b3…` | **9220** |
>
> Item now `needs_human_review`, `attempt_count 1`, route `target_not_dispatchable`,
> `triaged_by` the router — **and it HOLDS its dedup key** (1 non-terminal row on the key,
> where the wrong close had released it). That is the anti-churn property working, and it is
> the difference between the two rows above stated as data.
>
> **Survived the v1.0.1334 roll** (2026-08-24 15:39Z) — the fix is DB config, and the
> `_VERIFY` sidecar passes on the live row after it.
>
> # ✅ FIXED AND LIVE 2026-08-23 — migration `574`, config only, applied and verified at the route.
> `docs/agent_docs/sql_for_agents/574_required_fields_router_stops_closing_what_it_cannot_resolve.sql`
> (+ `_ROLLBACK`). Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_367_router_remit/`.
> Council `d48c0a89-9ff8-4286-bfe9-2690dc13d5bc`. ~~Kept OPEN until the parked disposition is
> observed on a real item in production.~~ **That criterion was met 2026-08-24 — see the banner
> above.**
>
> **§5's ordering was right about candidate 1 being first, and wrong about what it buys.**
> Widening the `comp` CTE alone does NOT repair anything, for two measured reasons this file did
> not have:
> 1. `file_rewrite` reads four spec fields the render-time producer never writes
>    (`component_id`, `page_id`, `component_function`, `reason` — 62 of 62 post-deploy items carry
>    them, 0 of 3 render-time ones do), and both `spec_paths` and `item_key_suffix_field` are
>    deliberate HARD ERRORS when unresolved (`create_work_item_action.go:281,294` and `:252-256`).
>    So widening trades a silent wrong close for a loud `mark_failed`.
> 2. The `partial` arm's destination is not a targeted edit: `save_page_sections_action.go:823`
>    DELETEs every agent-writable row on the page. **28 of 31** `from_rfm` conversions are already
>    failing there on the owned-page refusal (`bugs_open/333`, owned by the 277 lane) — and this
>    bug's own page is `rebuild_policy=owned`.
>
> **So the fix is neither candidate 1 nor candidate 2 as written. It changes the RULE:**
> a disposer may close only on **positive evidence of absence**. `stale` now requires the page row
> to be gone, or the component to be locked (277's accept-as-is), or a `build_status='removed'` row
> actually sitting at the slot. A lookup that finds nothing — or finds a real but non-deployed
> component — routes to a new **fifth park**, `park_not_dispatchable`, at `needs_human_review`,
> holding its dedup key. `triage.target_state` names which leg fired. Resolution moved to the
> lifecycle axis (`COALESCE(build_status,'pending') <> 'removed'`) so the component RESOLVES and the
> human sees its real state. The estate already states this rule at
> `revalidate_review_queue_action.go:684`.
>
> **⚠ The population is now VISIBLE AND HONEST, not REPAIRED.** Do not let anyone write that 367
> made render-time findings repairable. It did not. That needs `bugs_open/333` plus a producer that
> writes the convert arm's read-set; both named, neither taken.
>
> **Verified before applying, inside a transaction that was then rolled back** (so production was
> never the test rig), and again afterwards by reading the query back out of the live row:
> * the real item `562788c3` → `target_not_dispatchable`, `target_state=pending`, component
>   **RESOLVES** (`0a1498b3`, `html_len=9220`, `n_still_empty=2`)
> * **positive control** — retired slot (`tool-clip-path`/`ported-page`) → **still `stale`**,
>   `target_state=component_retired`
> * **positive control** — page that does not exist → **still `stale`**, `target_state=page_missing`
> * all **65** items of this type re-classified old-vs-new: **exactly one route changes**
> * apply-then-rollback returns `default_config` **byte-identical**
>
> **Council `d48c0a89` APPROVED at round 1**, 14 advisory objections, none high. Two drew code:
> migration **`576`** guards the `tomb` CTE against an empty-slot match (measured NOT reachable —
> 0 of 38 removed rows and 0 `page_components` rows anywhere have an empty `slot_name` — so it
> makes the bad state unrepresentable rather than merely unpopulated), and the behavioural checks
> became a re-runnable sidecar,
> `574_required_fields_router_stops_closing_what_it_cannot_resolve_VERIFY.sql`, **proven
> non-vacuous** by applying the 574 rollback inside a transaction and requiring it to fail.
> **Run that file before trusting this router's dispositions** — its two POSITIVE CONTROLS are
> the point, not the first check.
>
> **A precedent this file should have cited and did not** (a seat found it): `bugs_closed/032`,
> *"the completion verifier reads a DELETED component as a successful fix"* (2026-07-19), is the
> same defect one layer over, and its fix was *"return an error, never a verdict, so the gate's
> fail-OPEN policy turns a false success into a visible unknown"* — this bug's remedy in a
> verifier's vocabulary. Two lanes, five weeks apart, same shape, same answer. The lesson:
> **grep the closed bugs for the SHAPE, not just the mechanism** — a search for `build_status`
> or `required_fields_missing` finds neither of 032's words.
>
> **Also corrected here:** §4's *"Render-time items filed to date: 3"* and the §2 table remain true
> as dated, but `orchestration_states` retains only **~2 days**, so the route history they imply is
> not re-derivable — see the lane's NOTES and `LANDMINES.md`.

## 1. The defect in one paragraph

`required-fields-missing-handler`'s `classify` step resolves the offending component with
`... FROM page_components pc JOIN pg ON pc.page_id = pg.id ... WHERE pc.build_status = 'deployed'
AND COALESCE(pc.slot_name,'') = COALESCE(item.spec->>'slot_name','')`. A finding about a component
that is **not deployed** therefore resolves no component, falls to `route: "stale"` and is closed
by `close_stale` — *"the finding cannot be located on the live site"*. **That is exactly backwards
for the render-time producer `bugs_open/342` added:** its whole stated justification is that it
reaches the population the post-deploy check cannot, i.e. sections that render empty, get dropped
by assembly, and **never become a deployed row**. The producer reaches them; the router then
discards them as stale.

> # ✅ CONFIRMED BY THE LIVE HANDLER, 2026-08-23 17:09Z — the hand-run prediction held exactly.
> §2 below predicted this from the router's SQL run by hand. The router then picked the item up
> for real and did precisely that: `route: "stale"`, `component_id: ""`, `page_type: tool`,
> `rebuild_policy: owned` — **identical to the predicted row** — and closed item `562788c3` via
> `close_stale` at **`status='complete'`, attempt_count 1, `error` NULL**.
>
> **So the quietness warned about in §4 is not a projection, it is observed.** A true finding —
> `headline` and `trust_note` really are absent, the refusal really did fire on them minutes
> earlier — is now recorded in the queue as **`complete`, with no error and no trace of the
> disagreement**. Any census of "did we action our required-fields findings?" counts this as a
> success. The pre-fix behaviour (three failed attempts, parked) was ugly and *legible*; this is
> tidy and wrong.

## 2. Evidence — the router's own query, run by hand against three real items `[MEASURED 2026-08-23]`

The `classify` SQL was read out of the live `agent_definitions` row and executed with its own two
parameters bound (`$1` = site, `$2` = item), so this is the router's judgement, not a model of it.

| item | target `build_status` | page resolves? | component resolves? | `route` |
|---|---|---|---|---|
| `a31da7f3` (342's producer, pre-fix — no page context) | pending | ✗ | ✗ | `malformed` → failed ×3 |
| `562788c3` (342's producer, post-fix — page context present) | **pending** | **✓** `page_type=tool`, `rebuild_policy=owned` | **✗** | **`stale`** |
| `b2f1c7d4` (the post-deploy check's own item) | **deployed** | ✓ | **✓** `component_id`, `html_len=8711` | `no_plan_unbuildable` |

Read the middle row against the third: **the only difference that matters is `build_status`.** With
a deployed target the classifier resolves the component and produces a real route; with a
non-deployed one it resolves the page, fails to find the component, and calls the finding stale.

## 3. Why this is not `bugs_open/342` and not a regression in it

342's producer was, until 2026-08-23, filing items with no `page_name`/`slot_name` at all — those
were classified `malformed` and failed. That is fixed (`eb918bd58`, `23d2a577d`, live on v1.0.1330)
and the middle row above proves it: **the page resolves now where nothing did before.** This file
is the *next* obstacle, one layer down, and it lives in the router rather than the producer.

⚠ **It does mean a claim in `342` and in concept register `STY-057` is too strong** and both have
been corrected to point here: *"it reaches a population the existing producer structurally
cannot"* is true, but **reaching a population and being able to act on it are different things.**

## 4. Blast radius — small today, and the reason is itself the finding

- Render-time items filed to date: **3** (`a31da7f3`, `a6e00dcf` — both pre-fix and `failed` —
  and `562788c3`, post-fix, awaiting pickup). Two of the three were canary-induced; **`a6e00dcf`
  was real production traffic** (`loans-application-tracker`, 2026-08-22 13:32).
- The population is small *because the producer is young*, not because the shape is rare. Do not
  size this from three rows; re-measure with
  `SELECT status, count(*) FROM site_work_items WHERE item_type='required_fields_missing' AND spec->>'detected_at'='render' GROUP BY 1;`
- ⚠ **A `close_stale` disposition is quieter than the failure it replaces.** The pre-fix behaviour
  was three failed attempts and a parked item — visible. The post-fix behaviour for a non-deployed
  target is a clean close with the note *"cannot be located on the live site"*, which reads as
  correct triage. **This defect will not announce itself.**

## 5. Fix candidates, ordered by what makes the bad state unrepresentable

1. **Widen the `comp` CTE to the finding's own scope.** The item now carries `page_id`,
   `slot_name` and (where known) `component_id` — resolve on those and drop `build_status =
   'deployed'`, or accept any non-terminal build status. Makes the router's remit match the
   producer set that files into it, which is the property that was missing. Needs care: `stale`
   must still be reachable for a genuinely deleted page/slot, or the router loses a real route.
2. **Partition rather than widen** — keep the deployed-only path and file the non-deployed residue
   as a `capability_gap` with `gap_kind='handler_remit'`, per `bugs_closed/077`. Cheaper, honest,
   and it puts the residue where the roadmap sweep reads it
   (`diagnose_triage_action`: `item_type='capability_gap' OR status='deferred'`). Does not repair
   anything, but stops the silent close.
3. **Do nothing** — stated so the choice is a choice: the render-time producer keeps filing
   findings that are closed as stale, and 342's escalation is decorative for the population it was
   built to reach.

## 6. How to verify a fix

Run the router's own `classify` by hand (the recipe is §2 — read the SQL from the live row, bind
`$1`/`$2`) against a **non-deployed** component with a genuinely absent required field, and require
a route that is not `stale`. **Positive control in the same run:** a genuinely deleted page/slot
must STILL classify `stale`, or the fix has merely disabled a real route. Assert at the route, not
at the item's status — the status is downstream of the thing under test.

## 7. Where the record lives

`docs/agent_docs/docs024_key_docs_latest/bugfix_342_absent_required/NOTES_342_absent_required.md`
(2026-08-23 entries). Related: `bugs_closed/277` (the router, closed 2026-08-22), `bugs_open/342`
(the producer, and the corrected claim), `bugs_closed/077` (the `capability_gap` convention that
candidate 2 reuses), concept register `STY-057`.

**No `090` diagnosis run.** Stated plainly per the owner ruling of 2026-07-31: the substitute is
first-hand verification, and it is unusually direct here — the router's own SQL was read from the
live agent row and executed against three real items, with the deployed/non-deployed pair acting
as each other's control. What is NOT established is how often the shape occurs in ordinary traffic,
because the producer is four days old; §4 says so and gives the query.

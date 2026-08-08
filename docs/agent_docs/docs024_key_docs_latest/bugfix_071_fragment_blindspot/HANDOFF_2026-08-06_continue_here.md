# HANDOFF — 071 fragment arm, written for a cold start

> **SUPERSEDED IN PART, 2026-08-06 (post-roll) — the roll happened and every step
> below in §"Do this after the next roll" is DONE.** Do not re-run them looking for
> something to do. Evidence in `NOTES_fragment_blindspot.md` §"(post-roll)":
> `v1.0.1259` both replicas (`dead_fragment_link` 10, was 0, with both controls);
> a real `completeness-discovery-agent` dispatch against a four-case fixture (both
> dead fragments filed, both resolving ones silent, correct severity/routing);
> retraction proven by repairing one case and watching 2 findings become 1; fleet
> re-measured at 67 hrefs / 0 findings; fixture deleted and the pool site proven
> restored. **The only step still owed is the verifier's Go function** — see
> §"Still owed" below, rewritten.

**State: LIVE on v1.0.1259 and induction-proven. APPROVED round 1.** Nothing is
blocked. What remains is one small verification and three deliberately-deferred
pieces.

## What exists now

| thing | where | state |
|---|---|---|
| `dead_fragment_link` arm | `discovery_checks/check_phantom_internal_links_fragments.go` | `af2667453`, **LIVE v1.0.1259, proven** |
| wiring (second pass, `p.url` in the query, severity/routing) | `check_phantom_internal_links.go` | same commit, **LIVE** |
| `SplitFragment` | `datahelpers/links.go` | same commit, **LIVE** |
| `DocumentIDs` (extracted from `OrphanElementRefs`, which now runs on it) | `datahelpers/element_refs.go` | same commit, **LIVE**; refactor proven identical over 4,036 docs |
| writer constraint (no invented `#` anchors) | `prepare_link_context_action.go` `buildLinkConstraintText` | same commit, **LIVE** (effect unmeasured — needs a writer run) |
| claim-timeout exclusion | `sql_for_agents/322` + `220`'s declared list | **APPLIED AND RECORDED** 2026-08-06 10:20Z |
| register | LNK-031 new; **LNK-009 status corrected** | same commit |

Commits: `af2667453` (code+register+migration+docs) · `06d0e7695` (071
contribution, RUNBOOK, WRONG_CALLS ×2) · `72df8913e` (2 LANDMINES + sync) ·
`a13a60938` (council dispositions) · `eb35a0e13` (the looseness limit, in code) ·
`a865dc897` (owner log).

Council **APPROVED round 1**, `bbbb4132-4abe-4db1-a1ba-755377dab009`, 3 advisory
objections, none high. All dispositioned in NOTES §"2026-08-06 (evening)".
`af2667453` carries `Council-Submitted:`, so `098` credits it automatically —
**do not try to amend it** (forward-only).

## READ THIS FIRST if you are picking the lane up on 2026-08-07 or later

Re-verified on **`v1.0.1262`**: arm intact on both replicas (`dead_fragment_link`
10, verifier 2, `SplitFragment` 2, positive control 10, negative 0). No regression
across two rolls.

**Re-verified again 2026-08-07 08:35Z, still `v1.0.1262`** (image read off the pod
spec, not the makefile): 10 / 2 / 10 / 0 on both replicas.

**Re-verified 2026-08-08 on `v1.0.1264`** (fresh pods, started 13:08Z), five
strings, both replicas: `dead_fragment_link` 10 · verifier 2 · `SplitFragment` 2 ·
positive control 10 · negative control 0. **Three rolls, no regression.** Cadence
unchanged — still no `completeness-discovery-agent` dispatch since 08-05, still 0
`dead_fragment_link` rows.

**The arm has still never run on a real site, and that is a CADENCE fact, not a
fault.** `completeness-discovery-agent` — the agent whose `checks` array carries
this check — is dispatched **by hand**: 9 days out of the last 21, 1–6 sites each,
most recently 2026-08-05. `improvement-sweep` would drive it fleet-wide and is
`enabled=f` (`bugs_open/083`/`116`; the owner ruled staged supervised
re-enablement on 08-06).

> **CORRECTED 2026-08-08 — "since 2026-05-02" quotes ONE of two columns that
> disagree by three months.** The live row: `last_triggered_at = 2026-05-02`,
> **`last_completed_at = 2026-08-05 12:24:20+00`** — three days before this was
> written, and the same day `completeness-discovery-agent` last filed. `enabled=f`
> means the *scheduler* has not triggered it; it does not establish that nothing
> ran. **[UNRESOLVED — I did not establish what wrote the later stamp.]** Quote
> both columns or check what actually ran; do not repeat the May date alone. My own
> attempt to corroborate a run via a `workflow_plan::text ~ 'improvement-loop'`
> match produced **7 false positives** (council-gate runs mentioning the string) —
> see NOTES.

So, two things not to conclude:

1. **`SELECT … WHERE item_type='dead_fragment_link'` returning 0 does NOT mean the
   fleet is clean.** It means the check has not run. The fleet claim — 67
   fragment-bearing hrefs, 0 dead — comes from the **offline harness**
   (`RUNBOOK_fragment_blindspot.md`), and that is the figure to quote.
2. **Do not "fix" this by dispatching discovery at sites to make the queue
   non-empty.** Nothing is broken. The arm runs the next time any lane dispatches
   that agent for its own reasons, which happens every few days.

My own overclaim on this is corrected in four places and logged in `WRONG_CALLS.md`
(2026-08-07): I wrote "it rides an already-enabled check, so it cannot land inert",
which is true about **enablement** and silent about **cadence**. *Enabled* and
*driven* are two questions. The query that answers the second:

```sql
SELECT created_at::date, count(*) FROM site_work_items
WHERE created_by='completeness-discovery-agent' GROUP BY 1 ORDER BY 1 DESC;
```

## Still owed — two things, both small

`VerifyDeadFragmentLinkResolved` has not executed. Its three SQL shapes were
validated in both directions against the live fixture (href-presence returns `t`
for a rendered href and `f` for an absent one; the path normalisation resolves to
the target page's document and discriminates the live id from the dead one), but
the Go function is reachable only through `CompleteWorkItemAction`. **The first
real completion of a `dead_fragment_link` item exercises it.**

> **CORRECTED 2026-08-08 — the reason given here until today was FALSE.** It read:
> *"whose live callers are the dispatch loops — and `build-dispatch-loop` takes
> `item_domain='build'` while these items are `content`."* **There is no such
> filter.** That loop's only `load_work_items` step carries no `item_pipeline`, no
> `item_domain` and no `handler_agent`, and the Go filter is applied only
> `if pipelineFilter != ""` (`load_work_item_actions.go:635-673`). It loads **any**
> pipeline.
>
> **The real gate is `status`:** `load_work_items` selects
> `wi.status IN ('triaged','approved')` (`:653`), and discovery files every item as
> `detected` (`check_phantom_internal_links.go:175`). So a `dead_fragment_link`
> item sits at `detected` until something triages it.
>
> Two more corrections in the same pass: **`dead_fragment_link` is not always
> `content`** — `:145`'s routing override belongs to `unbuilt_internal_link`, not
> to us; ours falls through to `routeBySurface`, so a **chrome/`site_component`
> fragment routes to pipeline `build`** (page surfaces go to `content`). And the
> call path is **`build-dispatch-loop` → `process_item` (a `loop`) → sub_workflow
> step `mark_complete` → `complete_work_item`** — nested, so a top-level
> `jsonb_each` over `workflow.steps` cannot see it and reports the wrong callers.
> Full working in NOTES §2026-08-08.

**The registry itself is proven, including the direction that matters.** Fleet-wide
`result ? '_verification'` (underscore — `verification` returns 0 and reads like
"never ran") = **11 items, all pipeline `build`**, one of them `literal_markdown`
**`failed`** on 08-07: a completion the verifier *refused* in production, proven by
the `bugfix_201` lane (`45e0020af`). **But 0 of 321 `content` items have ever gone
through that path**, and content holds 0 `triaged`/0 `approved` — its 29 completions
came from `revalidate_review_queue` (a `revalidation` block, not `_verification`).

**So the cheapest real exercise is a chrome-surface fragment**, which lands in
`build` where the path demonstrably works — not the page-surface case that would
need a `page-build-handler` run against a pool site.

When it happens, the thing to check is that a completion is REFUSED while the href
is still rendered and the fragment still misses — a verifier that only ever agrees
is the failure mode the whole registry exists to prevent.

**Second, smaller: the arm's first production run.** Watch the next real
`completeness-discovery-agent` dispatch that includes a site carrying fragment
links (idea.uk, loancash.co.uk, loanandmortgagecalculator.co.uk,
fundamentallyai.com, vonc.com all have some). Expect **silence** — every one of
those resolves today — and read that silence as corroboration only if the run
actually covered the site. `created_by` + `created_at` on any item type from that
run is how you tell it ran at all.

## `bugs_open/213` lands on this lane's registry — checked 2026-08-07, we are clear

Filed today by `bugfix_122`: when two producers file under one `item_type` and the
registered verifier implements only one producer's predicate, the other's items
close `complete` untouched. **This lane owns the registry's newest verifier**, so
it needed checking. **`dead_fragment_link` is not exposed today** — one producer in
code (grep over `platform/ internal/ pkg/` finds only our own check pair), no
`designItemTypes` route, and zero config-driven producers of any verified item_type
(a disconfirmable zero: the same query minus the filter returns 11).

**But it is safe by circumstance, not by construction**, and that is the bit to
carry forward. `createdBy` is `config["source"]` falling back to `params.AgentType`
(`create_work_item_action.go:129-131, 283`), so **any agent definition can file
`dead_fragment_link` with any producer label, from DB config, with no code change**
— and `VerifyDeadFragmentLinkResolved` would then grade it against the fragment
predicate whatever it actually described. Two knock-ons, both measured, both in
NOTES §"2026-08-07 (later)":

- `count(DISTINCT created_by)` **cannot** enumerate producers — `params.AgentType`
  bottoms out at the literal `"generic"`, which carries 20+ item_types including
  `phantom_internal_link` (45 rows).
- 213's fix candidate 3 (verifier declares its producers) **cannot be satisfied
  from a code-side list**, because the producers live in config Go cannot see.

**Not written into `bugs_open/213` — deliberate.** That file is untracked and its
author was mid-session when I looked; appending would have risked losing their work
or mine. **If you pick this lane up and 213 is now committed, contribute the two
bullets above into it** (do not open a competing file — `who-owns.py 213` says
OWNED/ACTIVE by `bugfix_122`).

## DONE 2026-08-06 (kept for the method, not as work to repeat)

1. **Prove it shipped.** One exec, every replica, three strings:
   ```bash
   POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
   kubectl -n ai-persona-system exec "$POD" -- sh -c "
     strings /app/agent-chassis | grep -c dead_fragment_link;         # >0 = shipped
     strings /app/agent-chassis | grep -c phantom_internal_link;      # POSITIVE control, 9 pre-roll
     strings /app/agent-chassis | grep -c zzz_no_such_string_control" # NEGATIVE control, 0
   ```
   **Pre-roll baseline measured on `v1.0.1257`, both replicas: `0 / 9 / 0`.** A
   roll is not evidence your fix shipped — the image may predate the commit.

2. **Induce a live finding.** The arm only speaks when a fragment misses, and
   today's estate has none, so nothing will appear on its own — **an empty
   `dead_fragment_link` queue after the roll is the EXPECTED result and proves
   nothing.** Plant one: add `<a href="#zzz-induced-control">x</a>` to a scratch
   page component on a low-traffic site, run `completeness-discovery-agent`
   against that site, assert exactly one `dead_fragment_link` item naming that
   page and href, then remove it and confirm the verifier closes the item.
   (Item type is **`dead_fragment_link`**, singular-style like its siblings —
   the check NAME `phantom_internal_links` returns 0 rows and reads like "never
   fired". That trap is now a LANDMINE.)

3. **Re-run the no-op case in the same window** (memory: check the no-op case,
   not only the damage case): a resolving fragment such as loancash's `#content`
   must file nothing.

4. **Re-run the fleet harness** post-roll and compare with today's 67/0:
   the dump query and the planted-control recipe are in
   `RUNBOOK_fragment_blindspot.md`.

## Open, and deliberately not mine to close

1. **The deploy gate still cannot judge fragments.** `validate_page_content`
   receives the writer's `page_html` **without chrome**, so gate-side fragment
   validation would false-positive on every chrome-satisfied anchor. Needs a
   chrome-aware id load at the gate. The `bug_historian` seat objected here at
   medium and it is a fair objection, not a solved problem.
2. **Nothing REPAIRS a dead fragment.** Unlinking a label-bearing anchor leaves
   the label as bare text (recorded landmine), so: detection first, volume
   second, repair third.
3. **No section component emits a stable `id`** — the capability half. Until it
   does, a fragment link can only be avoided, never made to work on purpose.
   Changes every page's rendered HTML fleet-wide ⇒ architecture round, not a bug
   patch.
4. **Three unaligned consumers now reason about link-target resolution** (the
   gate, this arm, `link_repair.go`). Both `bug_historian` and `architecture`
   named it; architecture still returned `point_fix` and noted `DocumentIDs` is
   positioned for (3). Whether that split is itself the thing to fix is an
   architecture-track question.

## Traps this lane paid for (all now in LANDMINES/WRONG_CALLS)

- **A check's NAME is not its `item_type`.** `phantom_internal_links` → 0 rows;
  `phantom_internal_link` → 119. The zero reads as "this check is inert", which
  is the premise that justifies building a second, separate check — i.e.
  `bugs_open/093`'s shape.
- **`RegisterVerifier` obliges two more edits** (220's declared list AND the live
  `scheduled_tasks.pre_query`), and the build names only one at a time.
- **A clean measurement that has never been shown to fail is not a measurement.**
  Bit three times today: the 0-findings harness (fixed by planting controls), the
  differential test (fixed by id-stripped variants — the first run agreed only
  about nil), and the register's row count.
- **`site_components` has no `component_type` column** — it is `slot_name`.
- One regex for `href="(/[^"]*#[^"]+)"` finds **5** of the estate's **66**
  fragment links; you need the bare-`#` shape too.

## Not this lane's, recorded so nobody re-finds it

`component_library.go:1136-1147` still holds the `primary_cta_url` →
`/contact.html` / `secondary_cta_url` → `/about.html` defaults map. `bugs_open/203`'s
08-05 fix (`880a405a6`) removed the `cta_url` **scalar** defaults only. That lane
is active; contribute there, do not fix it here.

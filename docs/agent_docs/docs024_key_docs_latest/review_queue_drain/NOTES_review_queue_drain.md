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

---

## Turn 2 — 2026-07-25 evening: council APPROVED, and the objection that was right

**Verdict: APPROVED**, corr `ccba9c51-9bd5-4f1f-840c-ddd9e84a7bbe`. 13 reviewers,
3 abstained, **`unreadable: 0`** (the check that matters — "abstained" on the
16-seat gate counts relevance-filtered seats, so it is not the signal). Two
`object` verdicts, both medium, neither architecture-blocking.

### Objection 1 (bug_historian, medium) — RIGHT, and it hit the load-bearing claim

> *"The entire safety case for auto-closing … rests on an unverified assumption:
> that the checks which originally produced [these item types] actually re-run …
> The plan's own risks section admits this … but ships without confirming it."*

I had written the re-raise property as fact in the file header, the commit
message and the bug file. It was reasoned from the `idx_swi_dedup` predicate, not
measured. Checked properly afterwards, and **it does not hold unconditionally**:

```sql
SELECT item_type, count(*) FROM (
  SELECT site_id, item_key, item_type FROM site_work_items WHERE item_key IS NOT NULL
  GROUP BY 1,2,3 HAVING count(*) > 1) t
WHERE item_type IN ('unresolved_cta','required_fields_missing','needs_section_data')
GROUP BY 1;
--  (0 rows)
```

Zero recurrence for all three. **That result is itself ambiguous** and I nearly
misread it a second time: almost every row of these types is still OPEN, so the
dedup index would have blocked a second row regardless — absence of duplicates
proves nothing either way. The discriminating query is how many have ever gone
terminal at all:

```
unresolved_cta          : 70 rows, ALL needs_human_review — not one, ever
required_fields_missing : 45 parked + 1 complete
needs_section_data      : 45 parked + 7 complete
```

**8 items in the platform's entire history.** The re-raise path has essentially
never been exercised for these types.

What I could verify by reading the producers: all three insert with
`ON CONFLICT DO NOTHING` on a deterministic `item_key`
(`resolve_internal_links_action.go:257`; `plan_sections`' `createDeferredItems`;
`RunDiscoveryChecks` → `insertWorkItem`), so a terminal row genuinely does not
block a re-raise. What is NOT true is "the check will run again": all three fire
on a **page build** or a discovery pass over that site, never on a timer. **A
page that is never rebuilt again never re-raises, and a wrong close on such a
page is a silent loss.**

Corrected in the action's header (`> QUALIFIED 2026-07-25 …`), in the PLAN, and
in `bugs_open/033`. The mitigation that holds unconditionally — and which the
reviewer named — is the audit trail: every close records the exact fields it
judged populated in `result.revalidation` plus
`resolution_path='auto:revalidated'`, so a wrong close is individually
identifiable and reversible whether or not anything re-raises.

**Why this is worth the space.** The claim was written in the same confident
voice as the measured ones — 321-of-370, the leopardess ghost proof, 0-of-45 —
all of which came with a query. This one did not, and nothing in the prose
distinguished them. That asymmetry is precisely what the CLAUDE.md marker rule
(`[INFERRED]`/`[UNMEASURED]`) exists for, and I did not apply it to my own
strongest claim. The council caught what I would have shipped.

### Objection 2 (guardian, medium) — checked, and it clears

> *"loadPageComponentBySlot is described as shared machinery, but the plan only
> names one existing caller … confirm there are no other callers across other
> pipelines that also depend on the backfill side-effect."*

```
$ grep -rn "loadPageComponentBySlot(" --include=*.go platform/ internal/ cmd/ | grep -v "func loadPage"
platform/orchestration/actions/section_editor_actions.go:131:  pcRow, err = loadPageComponentBySlot(...)
```

**Exactly one caller** (`LoadEditContextAction`), and it still calls the
backfilling version — the split is behaviour-preserving for it. The guardian said
this would escalate to a blast-radius problem if another caller relied on the
backfill happening on every read; there is no other caller.

### Advisory (guardian low + guidelines) — invocation path

Both asked how this actually runs. Answer: `seed_review_queue_revalidator.sql`
seeds `diagnosis-review-queue-revalidator` with `dry_run=true`, fired manually by
`TRIGGER_revalidate_review_queue_v1.sh` — same shape as
`diagnosis-superseded-reviews`. It was invisible to the council because the gate
refuses `docs/` paths client-side, so the seed could not be in the submission.
Worth knowing for future submissions: **a code+seed change is only ever half
visible to the gate.**

### Advisory (editquality low) — `required_fields_missing` spec shape

Fair on the evidence as submitted; the shape WAS measured, I just did not quote
it. `spec.missing_fields` is an array of strings (`["headline"]`), confirmed
against live rows on ai-agent-orchestration.com and pinned in
`liveRequiredFieldsSpec` in the test file.

### Note taken, not actioned (reuse_agent)

> *"this is now the second sweep-with-verdict pattern against site_work_items
> (alongside reconcile_superseded_reviews) and a third partially-related one
> (insertWorkItem's dedup ON CONFLICT) … nobody has yet unified 'sweep produces
> evidence' vs 'sweep closes' vs 'insert dedups' into one documented family."*

Correct, and deliberately not forced here. Recorded so it does not silently
become three inconsistent things.

### OWNER RULING, 2026-07-25 — D1 and D2 answered together

Asked to decide what to do with the 78 machine failures parked in the queue, the
owner rejected the framing of all four options offered:

> *"they all should be able to be answered by the framework. the email is correct
> but if it isn't there shouldn't be a placeholder on the site, the content
> should be rewritten, the data should be collected via search etc etc"*

This is stronger than D2 as filed and it answers D1 as well: **`needs_human_review`
should not be where machine work goes to die, and the framework — not a person —
should resolve every one of these classes.** It also reframes 033: the queue's
real fix is that it should not fill, not that it should be drainable.

Seams identified (all **DB config — live immediately, no image build**):

| agent.step | mechanism | parks |
|---|---|---|
| `page-build-handler.mark_needs_review` | `fail_work_item` + `status_override: needs_human_review` | 51 (content failed validation) |
| `page-build-handler.mark_no_ready_sections` | `update_work_item_status` + `status: needs_human_review` | 27 (nothing ready to build) |
| `page-build-handler.mark_writer_skipped` | same | (same family) |
| `tool-improver` | own `status_override` | — |

`validate_content` routes its `error_step` to `mark_needs_review`, and there is
**no rewrite-retry path in the workflow at all** — the step list goes
`validate_content` → `mark_needs_review` → `complete_error`. So honouring the
ruling for the 51 means building one, not re-pointing a step.

**NOT changed this session, deliberately.** Re-pointing those steps live would
change every page build fleet-wide, immediately, on my own reading of a one-line
ruling — and the obvious naive version (send failures back to `triaged`) risks a
build loop, rebuilding and re-failing the same page on the same blocker while
burning credits. That is a design piece with its own diagnosis and council round.
Recorded here and in `bugs_open/033` as the next work, with the seams named so
nobody has to find them again.

### Diagnosis run c19ed5b2 — completed, produced NO verdict

Filed at 17:06 to grade the structural claim about `ON CONFLICT DO NOTHING`
leaving no record of a re-confirmed finding. The intake item went `complete` at
17:48 with no error, and the run wrote **5 bundles and zero verdict artifacts**:

```sql
SELECT kind, count(*) FROM diagnosis_artifacts
WHERE correlation_id='c19ed5b2-6d53-492a-af91-e78e175591d5' GROUP BY 1;
--  bundle | 5
```

Last bundle: `symbol_count: 2, symbols_in_scope: 2, truncated: false` — so it
resolved both symbols I pointed it at and had the code in hand; it simply never
graded. Same family as `bugs_open/043` (diagnosis runs that complete without a
verdict). **No verdict is not a refutation and not a confirmation** — the claim
was ungraded, and saying "the loop agreed" would be exactly the kind of
unearned-confidence move this thread has already logged once today.

What settled it instead was reading the producers directly, which I had to do
anyway to answer the council's `bug_historian` objection: `insertWorkItem`
returns `ok=false` on a dedup conflict and `RunDiscoveryChecksAction` adds that
to a local `skipped` counter and nothing else (`discovery_checks.go:166`) — no
column, no jsonb key, no log row on the item. So the claim holds, **verified by
reading, not by the loop.** Cost of the run: one queue slot and ~40 minutes, for
nothing. Worth knowing before reaching for it on a claim you can settle by
opening two files.

---

## 2026-07-28 (dashboard session) — the 50 were never invisible; the journey is verified; §4.2 of today's handoff corrected

Picked up the 2026-07-28 handoff to build its §4 step 2 ("make the 50 findable").
The step's premise failed its first measurement, so what shipped is a correction
and a verified access route, not code.

### The misstep, first: today's handoff asserted a status nobody had grouped by

The handoff says the 50 human-answer items are *"`status='detected'` with no
handler, so the dashboard cannot see them at all"*. Measured before building:

```sql
SELECT item_type, status, count(*), min(created_at)::date FROM site_work_items
WHERE item_type IN ('needs_section_data','owned_page_review','incomplete_page_group')
  AND status NOT IN ('complete','verified','rejected','wont_fix','cancelled')
GROUP BY 1,2;
--  incomplete_page_group | needs_human_review |  2 | 2026-07-15
--  needs_section_data    | needs_human_review | 42 | 2026-03-15
--  owned_page_review     | needs_human_review |  6 | 2026-07-17
```

**All 50, at `needs_human_review`, all `pipeline='build'`** (grouped by pipeline
separately — 50/50 build). That is the dashboard's default pipeline and its
best-supported status. They have been on the owner's screen the whole time.
The `detected` population (157 rows, 24 types) is a different set: all but
`image_url_404` (5) carry a live `handler_agent` — those are the dead-promoter
rows of `bugs_open/083_…_detected_findings_never_reach_a_handler.md`, which is
accurate about its own subject and contains no such misclaim; the error was
introduced in the handoff's paraphrase. Corrected in place there (§1 and §4.2);
row added to `WRONG_CALLS.md`. Queue totals at the same instant:
`needs_human_review` 327, `detected` 157 (both drifting daily — re-measure).

### What the dashboard actually supports (read, not assumed)

- `HandleListWorkItems` (`internal/core-manager/admin/site_admin_handlers.go:507`)
  accepts **any** `status` value as a query param (:609); default view is
  `status != 'complete'` (:614); `status_counts`/`type_counts` are scoped by
  pipeline+site only (:523 — the 033 fix, deliberate).
- The item-type dropdown is server-built from `type_counts`
  (`frontends/admin-dashboard/src/App.tsx:892,1010-1012`), so all three types are
  selectable with live counts. The status dropdown (:1000-1008) has no `detected`
  option — the only residual UI gap, and it belongs to 083-detected's
  observability, not to the 50. Not built: the fix for that queue is the
  promoter, not a viewer.
- `needs_section_data` has an auto-built form (`buildSectionDataForm`, :585) and
  a complete submit path (:749-826): each answered field merges into its source
  `site_specs` aspect (parsed from `spec.missing[].source`), a `content_rewrite`
  item is POSTed at priority 10 for `page-build-handler`, and the review item is
  resolved. **The owner's answer genuinely lands** — this is the surface for the
  42 oldest blocked questions (2026-03-15).
- `owned_page_review` / `incomplete_page_group` get the generic path: editable
  view + Resolve/Reject/Confirm (confirm creates a triaged follow-up item,
  `confirm_work_item_handler.go`).

### The journey, verified through one port-forward (2026-07-28)

`kubectl port-forward -n ai-persona-system svc/admin-dashboard 18080:8080`
(owner form: `make dashboard-port-forward`, then http://localhost:8080):

```
GET  /                                   -> 200 (dashboard HTML)
GET  /assets/index-CPAMeW9R.js           -> contains 'work-items?pipeline='  (the v1.0.1141 server-side-filter code is IN the served bundle)
POST /api/v1/auth/login  (bad creds)     -> 401 {"error":"Invalid credentials"}   (nginx→auth-service leg works, credential lookup completed)
GET  /api/v1/admin/work-items (no token) -> 401 {"error":"Authorization header required"}  (nginx→core-manager leg works)
```

Pods: admin-dashboard / auth-service / core-manager all `v1.0.1180`, started
2026-07-27 22:06 — post-dates the 1141 fixes, and the bundle grep above verifies
against the served artefact, not the tag. Auth users live in an **external**
MySQL (`rs17.uk-noc.com:3306`, db `catalogu_vectordb_chassis` — configmap-prod);
deliberately not queried — the 401-on-bad-creds already proves the lookup path,
and counting accounts in a production auth DB was not worth the intrusion.
**The one unverified step is the owner's own login** — `[UNVERIFIED]` by
construction; it is his credential. That is §4 step 1's residue, and it is now
a five-minute sit-down, not an engineering task.

### Where §4 stands after this session

1. Access route: verified to the credential check; his login is the last step.
   Port-forward works today; WireGuard (NodePort, since 07-20) is the standing
   alternative; an ingress remains unbuilt and is only worth discussing if he
   wants a permanent URL.
2. ~~Make the 50 findable~~ — **already findable**; nothing to build. Corrected.
3. Decision B (186 → report): untouched, his call, still the "do not build a
   second unread queue" warning.
4. Decision D (write-path rule): untouched, still the only door-closer.

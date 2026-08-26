# HANDOFF — `bugs_open/357`, component identity — 2026-08-26

**Cold-start. Read this, then `HANDOFF_2026-08-25b_continue_here.md` (superseded but it holds
the full evidence trail for the prune-floor finding, which is not repeated here in full), then
the bug file `bugs_open/357_HANDOFF_2026-08-22_a_whole_tool_page_is_stored_in_a_slot_that_claims_to_be_a_hero_component.md`.**

---

## THE ANSWER TO "CAN WE CLOSE IT?" — NO, AND THE REASON IS ONE NUMBER

**`population = 22`, re-measured 2026-08-26 08:17Z with the bug's own predicate.**

357's complaint is *"twenty-two live rows declare themselves the shared `hero` component while
storing a whole interactive tool."* That is still true of **all twenty-two**. The closing bar
on this estate is **fixed AND live** — a fix that is committed, approved and even proven stays
OPEN while the defect is still reproducible in production. Every one of those rows is still
reproducible today.

**What HAS finished is the machinery, not the repair.** That distinction is the whole state of
the lane:

| | state |
|---|---|
| Phase 0 — the provenance stamp | **DONE**, proven at volume |
| Phase 1 / F2 guard | **DONE**, proven with demand |
| Phase 2 — stop the mislabelling at birth | **DONE AND PROVEN IN PRODUCTION** (2026-08-25 12:24Z) |
| Phase 3 — repair the 22 | **NOT RUN.** One precondition untested, and currently untestable |
| **The bug itself** | **OPEN. 22 rows.** |
| NEW: the prune-floor contradiction | **DIAGNOSED, UNFIXED, needs its own bug file** |

So: **three of four phases are done, and the lane cannot close until the fourth runs.**

---

## ⚠ BLOCKER: the fleet has ZERO working LLM calls

```
success | calls | with_output | window
   f    |  126  |      0      | last 2 hours, both models, 2026-08-26 08:17Z
```

`"Your credit balance is too low to access the Anthropic API"` — **602** failed orchestrations
in the retained window, **63 in the last hour**, first seen 2026-08-25 23:46:18Z. It is getting
worse, not recovering. Owner notified 2026-08-26 00:0xZ.

⚠ **`llm_call_log` DOES NOT TELL YOU THE FLEET IS WORKING.** It has
`success boolean NOT NULL DEFAULT true` and logs the ATTEMPT. I read "35 calls in the last 30
minutes, latest seconds ago" as partial recovery; every one of them had failed. **Always
`GROUP BY success` with `count(*) FILTER (WHERE output_tokens > 0)` beside it.**

**This blocks the entire remaining lane**, because everything left needs a page rebuilt and
every rebuild runs a content writer first.

⚠ When checking the account: MEMORY [[the-fleet-key-is-not-on-the-default-console-org]] —
**capped while billing reads 0% used means the WRONG console org is in view**; check the keys'
`Last used`. Per the owner's 2026-08-23 ruling, **never read a key into a session** — probe
from the pod.

---

## What is left, exactly — four items, in order

### 1. Precondition 4 of migration 578 — the only thing between here and the repair

578 enforces three preconditions in code and **states a fourth in prose without enforcing it**:
*a rebuild has run on a page adopted at birth and the conservation loop preserved it — bytes
identical, component still `adopted-fragment`, row count unchanged.*

| precondition | enforced? | state |
|---|---|---|
| 1. phases 0/2 built, rolled, verified | no | **MET** |
| 2. an organically adopted row carrying a stamp exists | **YES, RAISEs** | **MET** — two of them |
| 3. 577 applied, carriers armed | **YES, RAISEs** | **MET** — 6 armed, re-checked post-roll |
| 4. **an adopted row survives a rebuild** | **NO — prose only** | **UNTESTED, and blocked** |
| 5. re-census on the day | **YES, by predicate** | automatic |

**Three attempts, none of which tested it:**

| run | correlation | outcome |
|---|---|---|
| canary #1 | `e0c2d505-9875-4347-a718-a852f32ec6b7` | FAILED — reaped >4h at `assemble_page`, `save_ran=f` |
| canary #2 | `5a0cad41-fe0c-4636-9b2d-9c942486019c` | FAILED identically, **on a fresh chassis** |
| control (non-adopted page) | `8d002375-1524-4abd-b04c-91a2e6a74277` | FAILED earlier still, at `write_page_content`, on the credit exhaustion — **did not discriminate** |

**When credit returns, run the CONTROL first, not the canary.** It is already set up: flag
`request-index` (2 rows, components `hero` + `contact-form`, **no adopted fragment**) as
`needs_rebuild`, clear any leftover flag on `index` so the result is attributable, and fire
`scratchpad/canary_rebuild.sh cv1.co.uk` (receipt-asserting; the repo's
`110_page_rebuild/072_page_rebuild` uses the silent-drop `kubectl run -i` + `kcat -P` shape —
do not use it).

- control also stalls at `assemble_page` → the stall is the **rebuild path**, or the
  `git_commit` immediately after it. Nothing to do with 357; precondition 4 needs a different
  vehicle.
- control reaches `save_result` → the stall is specific to a page carrying a ~17.5KB adopted
  fragment. **That IS a 357 finding** and it bears directly on whether phase 3 is safe.

⚠ **READ ANY REBUILD WITH ITS DEMAND CONTROL.** The real chain, traced from each step's own
`next_step` (not inferred — I got this wrong once and corrected it):

```
plan_sections -> write_page_content -> review_page_content -> check_review_approved
   -> assemble_page -> deploy_page (git_commit) -> save_sections -> update_page_status
   -> complete_page
```

`save_sections` runs **after** `assemble_page` **and after a git commit**. The only thing
separating a passing canary from one that never started is `collected_data ? 'save_result'`.
I read a pre-save state as a clean pass once — 1 row, right md5, right component, still
stamped, a perfect match against the pinned baseline, and **meaningless**.

Pinned baseline (`scratchpad/canary_before.txt`):
```
index         pos 1  slot hero                md5 26f484f2744ab3e9cd19e50f600a52b8  17,595 B
tool-example  pos 1  slot generic-text-block  md5 291b88d876e182a32a4a538c514878d2  20,076 B
both: component 9d4b922b-a548-4ca2-987c-ecacc7904b1f  version 3301ef65-4d83-4ea5-aa7c-65cb38e83653
```
**Pass = `save_result` present AND rows still 1 AND md5 unchanged AND still `adopted-fragment`.**
A row count of **2** is the carry-forward landmine firing: **STOP, do not run 578.**

### 2. Then run 578 — and not before

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/578_retype_mislabelled_tool_rows_HOLD.sql
```
By hand, never the migration runner. It re-censuses by predicate on the day, backs every row
into `page_components_backup_357_20260823`, prints the six `owned` pages by name, and RAISEs if
bytes / `slot_name` / `position` moved or a row is not reproducible from its own `content_data`.
Rollback: `578_..._ROLLBACK.sql`.

**Safety check already done:** 578's predicate keys on `cc.name='hero'`; the adopted-fragment
component's `name` is **"Adopted Fragment"**, so it structurally cannot sweep in the two new
adopted rows. Verified by running the predicate against them — **0**.

**Owner's standing instruction (2026-08-25): continue to phase 3 on a clean verification.** It
stands. It is not blocked on a decision — only on precondition 4, which is blocked on credit.

### 3. Verify at the artefact, then and only then close

Per 578's own afterword: curl each repaired page and assert its own markup (`class="tool-page"`,
its controls, its `<script>`); re-run the population query — the repaired rows must be **gone**
from it; let ONE rebuild run on a repaired page and compare row counts and per-row md5; confirm
the false `required_fields_missing` items about `hero` on tool pages stop being re-filed.

**357 closes when `population = 0` and that is verified at the served pages — not at the
migration's COMMIT.** A committed UPDATE is not a repaired page.

### 4. ~~File~~ **FILED 2026-08-26 as `bugs_open/406`** — the prune-floor contradiction, and it is NOT 357

> **DONE.** `bugs_open/406_HANDOFF_2026-08-26_adoption_routes_a_page_to_a_one_section_builder_and_plans_it_multiple_sections_so_the_save_is_refused_whole.md`
> — full evidence, closed-form arithmetic, four fix candidates ordered by what makes the bad
> state unrepresentable, and the verification queries. Pattern added to **016b §9**
> ("one action emits two statements about the same artefact that cannot both hold"), indexed
> in **016b §10**. **What remains is the FIX, which is a shared-seam change and wants the
> council gate — not a bug patch.** Re-measured at filing: **34** parked items across **16**
> domains, up from 32/~14 fourteen hours earlier.

`apply_adoption_plan_action.go:719` routes a page to `tool-recreation-handler` (whose save can
only ever emit ONE section) **and the same action, in the same transaction, writes that page a
multi-entry `pages.sections` plan.** `measurePageSectionCompleteness`
(`save_sections_prune_floor.go:148`) divides 1 by that count and `prune_floor_ratio=0.50`
refuses the whole save. Its own analysis spec for one cv1 page reads `"self_contained": true`
on a page it planned with three sections.

- Any adopted interactive page planned with **≥3** sections is unsaveable (1/3, 1/4 < 0.50).
- **21 of the 22** 357 rows sit on pages planned with **≤2** sections — the only ones whose
  one-section save could ever clear the floor. **This is why 357's population looks the way it
  does.**
- **32** `save_refused_incomplete` items are parked in `needs_human_review` from 2026-07-31
  across **~14 domains**, several named tool pages.
- Diagnosis loop: intake `f2fa4b9e-…`, **run `fbdaca97-a97e-41e6-b422-2475521e6a6c`** —
  returned **UNVERIFIABLE** (`scope-not-narrowing`), **not REFUTED**, naming two gaps, **both
  since closed first-hand** (see NOTES).

They share a cause and are different defects: **357 is the mislabelling; this is the refusal.**
Separate files. The durable fix — write a one-entry plan for a page routed to tool recreation —
is a change to a shared seam and wants the council gate, not a bug patch.

---

## What is PROVEN and must not be re-litigated

- **Phase 2 fired in production**, 2026-08-25 12:24Z, twice, and survived two chassis rolls:

```
cv1.co.uk/index         slot hero                adopted-fragment  regenerable t  stamped t  17,595 B
cv1.co.uk/tool-example  slot generic-text-block  adopted-fragment  regenerable t  stamped t  20,076 B
```

  `cv1.co.uk/index` **is** 357 — a whole tool in a slot named `hero` — and the row now states
  what it actually holds. Verified at the served pages (200, `tool-page` present, **zero**
  `data-component="hero"`), not just in the database.
- Both adopted rows point at **the same** `component_versions` row — the "1.00 versions per
  component, not a log" property, held across two independent adoptions.
- **Arming survives a roll.** It lives in `agent_definitions`, so it is config, not code. Six
  carriers armed, re-checked after both rolls.
- All three STOP conditions clear. `population_stamped` still **0**.
- **The current build is `v1.0.1341` and it genuinely carries phase 2.** Deployment spec and
  BOTH pods agree on the tag (`agent-chassis-6dd68888dc-*`, started 2026-08-25 23:11:5xZ), so
  this is not the same-tag-rebuild trap. Probed at the running binary with both controls —
  `adopt fragment: bound an unidentified fragment` **PRESENT**, positive control **PRESENT**,
  negative control **ABSENT**: the probe discriminates, so the PRESENT means something.

---

## Traps recorded from this work (all committed)

`LANDMINES.md`: **`--fidelity locked` on an adoption skips the entire build cascade** (the run
succeeds, a site appears, and the code under test is never reached) · **an adopted page's
interactive classification does not track its HTML** (verifier: **STILL_VALID**, 7/7).

`WRONG_CALLS.md`: I predicted both candidate sites' routing from markup and was **wrong about
both, in opposite directions** — the site with a working calculator went to the static builder,
the site with zero `<script>` tags went to tool recreation.

In NOTES, not yet landmines: spawned agent pods are **ephemeral**, so a run's logs are gone
within minutes — capture live or use the DB · `-l app=agent-chassis` is the **wrong pod set**
for spawned agents, and its silence means nothing · **`HEAD~1` is not your commit's parent** on
this tree · **`llm_call_log` logs attempts, not successes** (§ blocker above).

---

## Keys

| what | value |
|---|---|
| cv1.co.uk | site `8c3e9118-2455-4f0d-b01a-5dcde13dcf99` · adoption corr `468cb727-…` |
| lampenkap.com | corr `a3e1a948-…` (rebuilt + deployed; its index went to the STATIC builder) |
| adopted-fragment component | `9d4b922b-a548-4ca2-987c-ecacc7904b1f` ("Adopted Fragment") |
| its version row | `3301ef65-4d83-4ea5-aa7c-65cb38e83653` |
| diagnosis of the new defect | run `fbdaca97-a97e-41e6-b422-2475521e6a6c` |

```bash
# lane state, one line
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -A -F'|' -c "
SELECT (SELECT count(*) FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id WHERE cc.function='adopted-fragment') AS adopted,
       (SELECT count(*) FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id WHERE cc.name='hero' AND position(left(cc.html_template, position('{{' in cc.html_template)-1) in pc.rendered_html)=0) AS population;"

# is the fleet actually able to call an LLM? (attempts are NOT successes)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -A -F'|' -c "
SELECT success, count(*), count(*) FILTER (WHERE output_tokens > 0) AS with_output
FROM llm_call_log WHERE created_at > now() - interval '30 minutes' GROUP BY 1;"
```

---

## In one sentence

**The machinery is finished and proven; the repair is one blocked precondition away; the bug
stays OPEN at 22 rows until that repair runs and is verified at the served pages — and the
most valuable thing this lane produced is a separate, unfixed, estate-wide defect with 32
parked victims.**

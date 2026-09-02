# 434 — a live process placed a DEGRADED tool regeneration alongside the working one, and the page served the calculator twice

**Filed** 2026-09-02 by the `site_ai_agent_orchestration` lane, found while re-checking a page it
had fixed a week earlier.
**Severity** medium. Not a crash: a live shopfront tool page served **two copies** of the same
calculator, one of them a fragment, for **7 days**.
**Class** regeneration persists less than it replaced (`bugs_open/012`'s family) + placement not
de-duplicated.
**Status** DAMAGE REPAIRED on this page (migration `692`, applied + verified). **PRODUCER NOT
IDENTIFIED — that is the open half.**

## 1. What was observed

`https://ai-agent-orchestration.com/tools/agent-complexity-estimator.html`, live, HTTP 200:

- **two** `<h2>Agent Architecture Complexity Estimator` headings
- **two** estimator UIs — "Estimate complexity" and "Generate Architecture Estimate →"

`page_components` held two rows in the **same slot at the same position**, different
`component_id`, both unlocked:

| row | created | component | bytes | fieldsets | legends | **inputs** |
|---|---|---|---|---|---|---|
| `b2b7acbd` | 2026-04-09 | `tool-agent-complexity-estimator-leopardessconsulting-co-uk-ai-agent-orchestration-com` | 22,732 | 4 | 4 | **12** |
| `9aa63fc0` | **2026-08-26 14:48:27** | `tool-agent-complexity-estimator-ai-agent-orchestration-com` | 19,964 | 1 | 1 | **1** |

**An estimator with one input where twelve existed is a fragment, not a fork.**

⚠ **The byte counts would have waved it through.** −12% is well inside any plausible size floor;
the **input count** is what shows the loss. Any shrink guard on this class needs a structural axis,
not a length one.

## 2. Two second-order effects, both of which hid it

- ⚠ **It silently re-opened a defect that had been closed.** The page measured **0** firm contrast
  failures on 2026-08-25 after migration `625`. It measured **1** on 2026-09-02 —
  `#080B10` on `#0D1117` = 1.04:1 on the new component's button, which never received `625`'s
  repoint. **A fix verified at the artefact was undone by a component that arrived afterwards.**
- ⚠ **The standing de-duplication rule is VACUOUS on these rows.** The estate's guidance for
  `page_components` duplicates is "act only where `count(DISTINCT md5(content_data)) = 1`". Both
  rows carry `content_data = '{}'`, so both md5s are `99914b93…` and the rule reports agreement it
  never established — the same shape the LANDMINE warns about for NULL. **`content_hash` is empty
  and cannot stand in.** The usable discriminator here was the rendered artefact's structure.

## 3. What is NOT known, and it is the important half

**Nothing in the repo names `tool-agent-complexity-estimator-ai-agent-orchestration-com`** — no
migration, no commit, no script. So the component was created and placed by a **live agent
process**, and I could not identify which. Candidates worth checking first: the tool-fork path, and
whatever regenerates per-site tool components.

**It is NOT a class today.** Censused fleet-wide: exactly ONE page has two placements in one slot
with different `component_id`s created since 2026-08-20 — this one. So this is an incident, not a
pattern **yet**.

### ⚠ It is NOT `bugs_open/430`, and the test that separates them is one column

`430` (filed the same day) is *"forking a tool component drops `js_content`"* —
`deploy_tool_action.go`'s fork-on-deploy `INSERT`. That is the obvious suspect for anything that
creates a per-site tool component, and it is **the wrong one here**:

```sql
SELECT name, forked_from IS NOT NULL AS is_fork, js_content, html_template … 
--  …-leopardessconsulting-co-uk-ai-agent-orchestration-com   is_fork = TRUE   (the incumbent)
--  …-ai-agent-orchestration-com                              is_fork = FALSE  (the new one)
```

**`forked_from IS NULL` on the new component — it was never forked, it was GENERATED.** So 430's
INSERT cannot be its producer, and 430's mechanism (a column omitted from a copy) cannot explain a
structural reduction from 4 fieldsets and 12 inputs to 1 and 1 — copying fewer columns does not
rewrite markup.

⚠ **Both components have `js_content` length 0**, so the `js_content` signature does NOT
discriminate between them and must not be used to link these two bugs. Use `forked_from`.

**Filed before running this check** — I wrote this file and only then grepped `/bugs_open/` and
found `430`, which is the estate's stated order reversed. The check took one query and changed the
filing from "probably the same thing as 430" to a separate producer. It is recorded here rather
than quietly corrected because the near-miss is the point: a same-day sibling bug in an adjacent
mechanism is exactly what a duplicate filing looks like from the outside.

## 4. Fix candidates, ordered by what closes the door

1. **Find the producer and make fork-and-place REPLACE rather than ADD.** A path that places a new
   component into an occupied slot without removing or superseding the incumbent will do this again
   on any page it touches. This is the only candidate that makes the bad state unrepresentable.
2. **A structural floor on tool regeneration** — refuse to persist a tool component with
   dramatically fewer inputs/fieldsets than the one it replaces. `bugs_open/012`'s family already
   argues for this; the length-based floors cannot see it.
3. **A duplicate-placement detector** keyed on `(page_id, slot_name)` with >1 distinct
   `component_id`. ⚠ Must NOT use `content_data` md5 as the discriminator (§2).

## 5. What was done here

Migration `692_aiao_remove_degraded_duplicate_estimator_placement.sql` — removes the degraded
**placement** (the component row is left intact as evidence), guarded so it aborts if the survivor
has <12 inputs or has lost `625`'s repoint. Then re-assembled via the owned-page route.
**Verified: 1 placement, 12 inputs, `rebuild_policy='owned'`, 0 contrast failures, one heading.**

⚠ **If the duplicate returns, that is the finding** — it means the producer is still running, and
this file should be reopened against it rather than the damage repaired again.

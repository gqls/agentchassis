# 115 — the brief-fidelity audit predicted the owner's complaints three days early, and every finding died in `detected`

**Filed:** 2026-07-27 by the brochure_component_library workstream.
**Severity:** Medium as a defect, high as a signal — nothing is broken, and a correct
automated judgement about a live commercial site sat unread for three days until the
owner made the same judgement by eye.
**Class:** detection wired to nothing (the `bugs_open/083` / `bugs_closed/077` family).
**Status:** OPEN, unowned.

> **This file is deliberately narrow.** It was first written as a full account of the
> palette and imagery defects, and that was a **duplicate**: a sibling thread in this
> same workstream had already filed both, better, from the same owner report —
> **`bugs_open/113`** (palette merge: the layout's light literals fill the slots the
> spec never supplies; measured in headless Chromium, fixed in code, inert until the
> roll) and **`bugs_open/114`** (21 generated images live, 3 referenced). Read those
> for the mechanisms. I filed before grepping `/bugs_open/`, which is the rule that
> exists for exactly this; logged in `WRONG_CALLS.md`. Everything below is what those
> two do **not** cover.

---

## The finding

`site_work_items` for `fundamentallyai.com` holds three `audit_finding_brief_fidelity`
rows created **2026-07-24**, all still `status='detected'` on 2026-07-27:

1. *"The brief requires 'Any numbers or statistics shown must be rendered as real,
   code-generated charts from true figures…'. No chart component exists anywhere in the
   built inventory — zero of 27 components are chart components, and the metrics/data
   visualisation layer is entirely absent."*
2. *"The brief requires 'Page templates differ meaningfully across linked pages — not
   one stamp repeated.' The model-fine-tuning and multi-agent-review-council pages both
   share the identical component pattern… suggesting template repetition rather than
   meaningful differentiation."*
3. *"The brief requires 'Line illustration for any people or human figures…' with 'one
   consistent tint/treatment across illustrations'. Only 2 of 27 components contain
   images — raising serious doubt that the illustration system is meaningfully present
   or that the tint treatment is applied coherently across the site."*

On **2026-07-27** the owner looked at the site on mobile and reported, unprompted:
*"There is not enough imagery on the site"* (= finding 3) and *"It is really not at all
exciting or professional"* (finding 2 is one concrete cause). Finding 1 was acted on —
by a thread building the chart by hand on 07-26 **after the owner asked for it**, not
because anything read this row.

**The audit was right about all three, three days early, and no human or agent ever saw
it.** That is the defect: not the judgement, the routing.

## Why it went nowhere

`audit_finding_brief_fidelity` appears in the entire Go tree **only in a coverage
test**:

```
$ grep -rn "audit_finding_brief_fidelity" platform/ internal/ --include=*.go
platform/orchestration/actions/discovery_checks/verifier_coverage_test.go:285
platform/orchestration/actions/discovery_checks/verifier_coverage_test.go:564
```

No `agent_definitions` row consumes it either:

```sql
SELECT type FROM agent_definitions
 WHERE default_config::text LIKE '%audit_finding_brief_fidelity%'
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;   -- 0 rows
```

So `detected` is terminal in practice. The writer (`write_audit_findings_action.go`)
produces a well-formed, specific, correct item and hands it to nobody.

**Scale, and it is the uncomfortable part:** fleet-wide there have only ever been
**3 rows of this item type — all three are the ones above.** The brief-fidelity audit
(`features_open/016`) has run **once, on one site, and was right about everything**.
It is not an unreliable check being ignored; it is a reliable check nobody is running.

## Fix candidates, ordered by what closes the door

1. **Route the item type, or refuse to write it.** Give it a handler, or file it as
   `capability_gap` the way `bugs_closed/077` resolved the identical shape — the
   platform's durable "found work I have no handler for", which is read as a roadmap
   and cannot silently rot. An item type with no consumer should not be writable.
2. **A terminal-state audit.** `detected` reads like a live state and behaves like a
   grave. Any item that has sat in a non-terminal status past a threshold with no
   claimant is a routing defect by definition, and one query would list them fleet-wide
   for every item type at once. This generalises past this bug — it is the detector for
   the whole `083`/`077` family rather than a third instance of it.
3. **Run the brief-fidelity audit on every site, on a cadence.** One site, once, is not
   a programme. Cheap relative to what it caught.
4. **Surface open findings in the workstream handoff.** Every one of these would have
   been read on any day this workstream started a session — the standing five are read
   at cold-start and `site_work_items` is not.

## How to verify a fix

Not "the handler exists" — **take a finding through to a change**. Pick finding 2
(template repetition, still true today: `model-fine-tuning` and
`multi-agent-review-council` still differ by one component) and confirm the item leaves
`detected`, reaches a handler, and results in a page whose section plan differs. A
handler that claims the item and closes it without changing anything reproduces the
defect with better paperwork.

## What this is NOT

- Not `113` (palette merge) or `114` (imagery never referenced) — those are the
  *mechanisms* behind two of the complaints. This is about a correct finding that named
  them first and was never read.
- Not a duplicate of `083` (detected findings never reach a handler), but the same
  family and probably the same fix. **Contribute there before starting work here.**

---

## 2026-07-27 (evening) — one row drained, and the mechanism measured where it is owned

**Finding 1 is closed, with evidence rather than by fiat.** It read *"No chart component
exists anywhere in the built inventory — zero of 27 components are chart components"*.
That was true when written on 07-24 and is now false: `evidence-chart` was built,
registered and placed on 07-26, and is live on the index with its values sourced from
`site_specs.evidence_base`, so a chart structurally cannot display an unverified figure.
The resolution is written into the row's `suggested_action`, so the closure carries its
own justification and does not have to be taken on trust.

**Findings 2 and 3 stay `detected`, because they stay true.** `model-fine-tuning` and
`multi-agent-review-council` still differ by one component. Imagery is repaired
(`bugs_open/114`) but still thin. Closing them would be the failure this file is about,
one level up — a row marked done by someone who wanted the queue shorter.

**The mechanism is bigger than this file and is measured where it is owned.** Contributed
into `bugs_open/083` rather than here, since `who-owns.py` says OWNED: fleet-wide,
**298 non-terminal items across 13 item types have no handler named**, plus 9 more
naming `human-review`, which is not a registered agent. The standout is
`needs_section_data` — 44 items, 10 sites, **oldest 135 days**.

> **Caveat carried across from that contribution, because a number this size gets
> quoted:** `capability_gap` is *supposed* to have no handler. "No handler" is not
> automatically a defect, and 298 must not be read as 298 broken things. Which of the
> other twelve types are deliberate is undecided work.

The candidate I would put first is still the one in this file's fix list and now has
fleet evidence: **a routing audit on a cadence, not a handler per type.** One query
finds every type whose rows sit non-terminal with no live handler, independent of what
the type means. `bugs_closed/077`, this file and `083` are three sightings of one thing;
making the *next* one visible is cheaper than draining this batch.

---

## 2026-08-15 — the WRITER-side mechanism is now filed as `bugs_open/279`

This file's fix candidate 1 ("route the item type, or refuse to write it") got its
mechanism file: **`bugs_open/279`** establishes, with code cites and a live census,
that `classifyFinding`'s fallback mints `audit_finding_<category>` for any category
outside its hardcoded sets (silently, deprioritised), that the `work_item_type` field
two prompts demand is discarded unparsed (the struct has no such field; zero readers
at any layer), and that **brief-fidelity-auditor's hardcoded `category:"brief_fidelity"`
makes 100% of its output take that fallback** — which is exactly why this file's three
findings died. 278 owns the writer fix; candidates 2–4 here (cadence, terminal-state
audit, handoff surfacing) stay with this file and `083`. Census update: the fleet now
holds **6** `audit_finding_%` rows (4 detected on mortgagecalculator, 2026-08-13).
Related: `bugs_closed/272` fixed the same action's other silent-zero (object-shape
parse) on 2026-08-15, v1.0.1301.


---

## STATUS UPDATE 2026-08-15 (279 owner-decision session) — the auditor is being promoted; this file's evidence rows are cancelled by owner decision

The owner ruled on both halves today (decisions recorded in `bugs_open/279`):

1. **The four 2026-08-13 `detected` rows (this file's evidence) are CANCELLED** —
   owner decision, reason stamped in each row's `error` column. The audit will be
   **re-run once the routing fix (commit `d6d56e540`) is live**; the re-run doubles
   as 279's live verification. The three 07-24 findings this file was written about
   remain as history in the row archive and in this file's own quotes.
2. **The auditor becomes a real check**: it now speaks the router's category
   vocabulary (migration `417`, applied — category chosen by repair shape,
   `audit_source='brief-fidelity-audit'` keeps the identity) and joins the
   improvement loop's audit chain (migration `418_HOLD`, applied post-roll). This
   file's candidate 1 ("route the item type, or refuse to write it") is done via
   279 (unknown categories → `capability_gap`); its cadence candidate is done via
   the wiring; the terminal-state-audit family candidate stays with `083`.

Residual for this file: after the post-roll re-run, verify the findings land
routed (not `capability_gap`), then this bug is closable — fixed AND live.

# 172 — the diagnosis bundle's agent-state cap silently truncates to a NON-DETERMINISTIC five, under a heading claiming it covers every agent type the symptom named

**Filed 2026-07-31 by the `bugfix_164` lane, AT THE COUNCIL'S EXPLICIT REQUEST.** OPEN, UNOWNED.
**LATENT, not live — and one short of firing.** See § Measured before quoting this file.

## Why this exists

Fixing `bugs_open/164` (the body cap that `break`s), my submission claimed a blast radius of
"three char-budget cap sites repo-wide, all in one file". The council **APPROVED** (corr
`75f3cd52`) but the `bug_historian` seat attached an architecture note asking for exactly one
more thing:

> "this is the third piecemeal pass over caps in this one file (145, bd003f67a, now this) — the
> plan's own grep shows this closes the set, but a human should confirm no fourth cap-shaped
> site exists outside the grep's pattern (**e.g. count-based rather than char-length-based
> caps**) in this same loop family."

It was right to ask. My grep keyed on `+ len(x) > cap`, which structurally cannot see a
**count-based** cap or a **slice reslice**. Re-run for those shapes, the loop family has a
fourth site, and it is a genuine instance of the same 016b §9 pattern.

**Recording the chain plainly, because it is the point:** `bd003f67a` audited this file for
this shape and missed 164; 164's own audit missed this. Each pass narrowed by the shape it
happened to grep for. The seat that caught both is the same seat.

## The defect

`platform/orchestration/actions/diagnose_load_runtime_action.go:941-949`, in
`gatherAgentState` — which inlines live agent config/state for every agent type named in the
symptom, so the verdicter can cite it:

```go
matched := matchAgentTypes(symptomText, allTypes)
if len(matched) == 0 {
    return
}
if len(matched) > typeCap {
    matched = matched[:typeCap]          // ← silent. No marker, no count, no log of the loss.
}

b.WriteString("\n### agent state (auto-gathered: agent types named in the symptom/hypothesis)\n\n")
```

`typeCap` is `agent_state_cap`, **default 5** (`:409`). `matched` then drives all three
evidence queries (`:960` config, `:992` state, `:1017` `llm_call_log`), so a truncated
`matched` truncates every one of them.

## Why it is worse than 164 was

1. **The heading asserts the coverage the code just discarded.** It says "agent types named in
   the symptom/hypothesis" — not "up to 5 of them". A verdicter reading it has no way to know
   an agent it named is missing, and **will draw conclusions from that absence**: the whole
   purpose of this section is to let it reason about which agent misbehaved.
2. **Which five survive is NON-DETERMINISTIC.** The source list is
   `SELECT DISTINCT type FROM agent_definitions WHERE deleted_at IS NULL` (`:926-927`) —
   **no `ORDER BY`**. Postgres is free to return a `DISTINCT` in any order (hash-aggregate
   order in practice, which varies with plan, statistics and concurrency). 164's casualties
   were at least an alphabetical tail and therefore reproducible; **these are not.** Two
   identical diagnosis runs on the same symptom can gather state for different agent types,
   both reporting success, and nothing distinguishes them.
3. **The log cannot reveal it either.** `:1039` logs
   `zap.Strings("matched_types", matched)` — the ALREADY-TRUNCATED slice. There is no
   pre-truncation count anywhere, so the loss is invisible in the artefact *and* in the logs.
   (164 at least set a `truncated` boolean.)
4. **`bd003f67a` explicitly cleared this file.** Its commit message records: *"Confirmed NOT
   instances in the same audit: diagnose_run_checks and **diagnose_load_runtime** already
   report their caps."* That is true of `maxCodeChecks` at `:454` — a **different** cap in the
   same file. The clearance was written at FILE granularity over an instance-level check, and
   this site has been sitting behind that sentence since 2026-07-20.

## Measured — and the answer is LATENT, so do not quote this as having fired

**It has never fired in the retained window, and it came within one.** Instrument chosen
deliberately: `orchestration_states` retains only **one day** here (16 symptom-bearing rows,
all 2026-07-31) and cannot bound anything historically, so the 30-day `diagnosis_artifacts`
bundle corpus is the instrument instead.

The section renders one `agent_definitions[<type>]: root ai_service` line per matched type, so
the types it actually covered are countable from the artefact:

```sql
WITH b AS (
  SELECT (SELECT count(*) FROM regexp_matches(body, 'agent_definitions\[[^\]]+\]: root ai_service', 'g')) AS types_listed
    FROM diagnosis_artifacts WHERE kind='bundle' AND body LIKE '%### agent state (auto-gathered%'
) SELECT types_listed, count(*) FROM b GROUP BY 1 ORDER BY 1 DESC;
```

| types listed | bundles |
|---|---|
| **4** | 10 |
| 3 | 11 |
| 2 | 15 |
| 1 | 36 |

- **The path is live and well exercised: 72 of 254 bundles (28%)** carry the section, window
  2026-07-09 → 2026-07-31.
- **Maximum ever listed: 4. The cap is 5.** So it is latent by a single agent type — not
  theoretical, and not currently damaging anything.
- The population it caps against is **185 active agent types** (`agent_definitions`, active,
  non-snapshot, not deleted), so a symptom naming six is unremarkable. Note the platform's own
  symptom-authoring guidance pushes the other way: 016b tells authors to *state the mechanism
  and point at the symbols where the evidence lives*, which makes richer, multi-agent symptoms
  more likely over time, not less.

**`[UNMEASURED]`:** whether any *historical* run (outside the 22-day bundle retention) hit it.
Unknowable — the artefacts are gone.

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **Report the truncation in the bundle text, and make the survivors deterministic.** Two
   small halves of one fix: add `ORDER BY type` to the `DISTINCT` query so the kept set is
   reproducible, and write the count where the dropped types would have been — e.g.
   `_(N further agent type(s) named in the symptom were NOT gathered — cap M. Absence here is
   not evidence an agent was uninvolved.)_`. **This is the shape `164` and `bd003f67a` both
   settled on in the sibling file**, so it is a reuse, not a new convention; the wording above
   is deliberately near-verbatim from `diagnose_assemble_bundle_action.go:252`.
2. **Name the kept types in the heading** rather than asserting the general claim: "agent
   state (auto-gathered for N of M agent types named)". Closes the door on the heading being
   read as complete, independently of whether anyone reads the marker.
3. Raise `agent_state_cap`. **Weakest, and it is a knob** — it moves the cliff without making
   it visible, which is the candidate `145` was refused for. Worth noting only because the cap
   is one short of biting, so a raise looks tempting and would hide the defect for longer.

## How to verify a fix

- **Induce it — every assertion here is vacuous below the cap**, and the measurement above says
  production has never reached it, so a live run will NOT exercise this. Unit-test
  `gatherAgentState` with `typeCap=2` and three matching types: assert the marker text is
  present, that it names the count dropped, and that the two kept types still render.
- **Determinism branch:** with `ORDER BY type`, the same input must produce the same kept set
  across runs. A test that runs once cannot see this — assert the ordering explicitly.
- **Negative control:** a symptom naming fewer types than the cap must produce a
  byte-identical section to today, or every existing diagnosis's baseline moves.

## Related

- **`bugs_open/164`** — the sibling cap in the sibling file, same loop family, fixed
  2026-07-31 (`906fc4323`, council `75f3cd52` APPROVED). Its `VERIFY` and its lane docs
  (`docs024_key_docs_latest/bugfix_164_bundle_body_cap/`) carry the shape this should reuse.
- **`bd003f67a`** — the 2026-07-20 audit whose file-level clearance covered this site.
- **016b §9** *"A hard cap that silently discards its input's tail rewrites meaning"* — the
  indexed pattern. This instance adds a wrinkle worth folding in: **a cap whose discarded set
  is not even deterministic**, so the loss is unreproducible as well as unreported.
- Family: `bugs_open/012`, `bugs_open/171` (an audit-pass limit reporting a capped site clean),
  and MEMORY's *"a `complete` work item is not a repaired artefact"* — every member is a cap
  that reported success.

---

# 2026-08-02 — FIXED (both halves), council APPROVED, awaiting a roll

**Worked by the `bugfix_172_agent_state_cap` lane.** Full account:
`docs/agent_docs/docs024_key_docs_latest/bugfix_172_agent_state_cap/`.
Commits `3761a04ca` (fix + tests) and `c8031e284` (a council objection, below).
Council `d47b826e-6fc6-42ad-a2ef-62b1f1ba0b88` — **APPROVED round 1**, 4 advisory
objections, none high.

## Everything above is correct. It is also HALF THE DEFECT.

The re-measurement this file asks for holds: **still max 4 types listed against a
cap of 5** — the count-based cap it names has never fired.

But the same function's **llm_call_log gather has been firing since at least
2026-07-20**, and this file does not mention it. It issued ONE query for ALL matched
types under a single shared `ORDER BY created_at DESC LIMIT n`, so rows are allocated
by **global recency across the whole named set** and the chattiest agent takes the
entire budget:

| bundles naming | with rows | showing ONE type in the log lines |
|---|---|---|
| 4 types | 10 | **10** |
| 3 types | 10 | **10** |
| 2 types | 3 | **3** |

**23 of 23.** Reproduced against the live table: `{page-content-writer, council-gate,
diagnose-agent}` → 10 rows, **all council-gate**; `page-content-writer`'s 18,286 rows
and `diagnose-agent`'s 324 render nothing, with no marker distinguishing that from an
agent that made no LLM calls.

**Why this file missed it, recorded because the file itself makes this point:** its
grep keyed on the count-based shape after 164's grep keyed on the char-based one.
A shared-budget `LIMIT` is neither. That is the fourth consecutive pass over this
family narrowing by the shape it happened to search for — and this file's own §"Why
it is worse than 164 was" §4 is about exactly that failure at file granularity.

**The tell that broke it open:** `count(*)` of the rendered log lines reads HEALTHY —
a full 10, the cap, in 47 of 72 bundles. Only `count(DISTINCT agent_type)` *within*
those lines shows all 10 belong to one agent. **A count cannot see a distribution.**

## What shipped

Candidate 1 as written, plus the second half:

- `ORDER BY type` on the DISTINCT listing → the kept set is reproducible.
- Dropped types are **named** (not just counted — they are known here, unlike the
  sibling at `diagnose_assemble_bundle_action.go:328` whose wording is otherwise
  reused near-verbatim); the heading counts kept-vs-named **only when the cap fires**,
  so the negative control this file asks for stays byte-identical.
- `ROW_NUMBER() OVER (PARTITION BY agent_type ...)` → a per-type budget, which is what
  `agent_call_log_limit` and the heading always implied.
- A gathered type with no rows is **stated**; kept distinct from a type that filled its
  budget (`bd003f67a`'s coverage-vs-defect distinction).
- The log gains the **pre-truncation** count; it previously held only the truncated
  slice, exactly as §3 above says.

Candidate 3 (raise the cap) declined for the reason this file gives.

**Verification.** Seven sqlmock tests induce the cap (it cannot fire in production).
Four mutations run: three caught; **deleting `ORDER BY type` was NOT** — sqlmock
replays rows in the test's own order and structurally cannot observe the database's,
so this file's warning ("a test that runs once cannot see this") applies to the test
itself. Closed with a strict query-text assertion, plus the live ordering check in the
lane's RUNBOOK §3. Logged in `WRONG_CALLS.md`.

## Two things the council added, both acted on

- **`llm_reliability` (medium), a real defect in the fix:** my new "no rows" marker
  said "the table holds nothing for them (this is an answer, not a cap)". `agent_type`
  was **relabelled 2026-07-26** — `experience-planner`, `feature-implementer` and
  `fix-implementer` all stop before 07-27 — so a symptom spanning that boundary would
  be handed a confident "no calls" for an agent with plenty under a former name: the
  false negative this fix exists to remove, reintroduced by its own wording. Now
  states what is true and names the boundary (`c8031e284`).
- **`bug_historian` (medium):** nothing audits for other instances of this shape. The
  inventory it asked for found a fourth site → **`bugs_open/181`** (three unreported
  row caps in `diagnose_code_lookup_action.go`, whose sibling cap eight lines above
  DOES report itself).

## STATUS: still OPEN — fixed and reviewed, NOT yet live

The bar is *fixed AND live*. The chassis has not rolled since the commits, so the
defect is still reproducible in production. **Every bundle written before it rolls
carries the starvation, and they are retained ~30 days** — a `LANDMINES.md` entry now
warns against reading a missing agent in a retained bundle as an uninvolved one.

To finish: roll, then verify at the pod with BOTH controls (positive: a string the
change added; negative: `ORDER BY created_at DESC$`, expect 0) on **every replica** —
`bugfix_172_agent_state_cap/RUNBOOK_agent_state_cap.md` §7 has the exact commands.

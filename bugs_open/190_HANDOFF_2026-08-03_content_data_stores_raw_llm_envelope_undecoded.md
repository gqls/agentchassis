# 190 — `content_data` stores the raw LLM transport envelope, undecoded; the render above it is good, so the poison is invisible until a rerender

Filed 2026-08-03 by the bugfix_140 lane (handoff plan item 7). Every figure below was
measured live this session; queries inline.

## One-line

Two live, `deployed` `page_components` rows store the action-output envelope
`{"type":"text","result":"<raw text>"}` **verbatim** as `content_data`. Their
`rendered_html` is currently real content, so nothing looks wrong — but `content_data`
is the source of truth every reasoned rerender regenerates from, so each row is a
armed regression: the next rerender finds envelope keys instead of declared fields.

## The rows (measured 2026-08-03 ~21:30Z)

```sql
SELECT pc.id, cc.function, p.name, s.domain, pc.created_at, pc.build_status
FROM page_components pc
JOIN pages p ON p.id = pc.page_id
JOIN sites s ON s.id = p.site_id
LEFT JOIN content_components cc ON cc.id = pc.component_id
WHERE pc.content_data ? 'type' AND pc.content_data ? 'result'
  AND pc.content_data->>'type' = 'text';
```

| row | component | page | created | state today |
|---|---|---|---|---|
| `d2e9644b-c409-4f4c-ab81-265fc20bf31b` | `article-body` | finetuning.uk `tool-ai-data-risk-checker-guide` | 2026-07-15 | `rendered_html` 12,441 chars of real article, `deployed` |
| `25c73a1c-b3af-48af-978b-95f7e500e8fa` | `Pricing Tiers` | gaswholesalers.com `how-pricing-works` | 2026-05-02 | `rendered_html` 1,759 chars incl. a dead `<a></a>` (140-lane residual table), `deployed` |

Denominator: 1,041 `page_components` rows with non-`NULL` `content_data` (my filter —
the 140 lane's NOTES said "2 of 1,145" counting a different population; both agree the
defect count is 2). Neither row has any `page_component_history` entries — both predate
it, so `history.source` cannot name the writer.

## Mechanism and prior art — the CAUSE is already diagnosed in-tree; this file is about the residue

`platform/orchestration/actions/json_envelope.go`'s header IS the diagnosis:
`ExecuteLLMPromptAction` used to treat any unparseable response as "the model returned
prose", produce `{"type":"text","result":"<raw>"}`, and "that envelope then flowed into
`page_components.content_data` verbatim". The finetuning row here is **the exact page
that header names** as the 2026-07-16 manual one-off repair — the repair evidently fixed
`rendered_html` and left `content_data` as the envelope. The parse side has since been
fixed in three tiers (escaping repair 2026-07-16; complete-value recovery 2026-07-26,
`bugs_open/088`; corrective re-ask, `bugs_open/119`), and the text path that still
exists by design (`ai_actions.go:858-871`) now carries `json_contract_unmet` +
truncation markers so downstream fails loud.

Prior art: `bugs_closed/008` (stop_reason undecoded — the envelope-decode class),
`bugs_closed/054` (double-wrapped envelope in scheduled_task input_data).

**Why no 090 run** (per the 2026-07-31 owner ruling, stated plainly): the structural
mechanism is not a new claim — it is already diagnosed, fixed and documented in-tree by
`json_envelope.go` and the three bug files above. What this file asserts is residue
(2 rows, queried live above) and a missing repair path (verified below by running the
live parser on both payloads). That is first-hand verification of the only claims made.

## Measured: what today's parser makes of the two trapped payloads

Ran `actions.ParseLLMJSONWithProvenance` (the live code, via a scratch module with a
`replace` to this tree) on each row's `content_data->>'result'`:

| row | verdict | meaning |
|---|---|---|
| `d2e9644b` (finetuning) | **recovers via `repaired`** — one `content` key, 11,141 chars | fully mechanically repairable: decode → store the object |
| `25c73a1c` (gaswholesalers) | recovers via `prose_around` — 7 `tier_*` keys, **131 chars total** | NOT repairable mechanically: the first small JSON object parses, but the real section content is in the markdown tail after it, which recovery correctly refuses to guess at |

Observation, not a claim: `25c73a1c`'s markdown tail ends
`**...:** gas@contactforsales.com | +44 (0) 7934 524 911` — contact details inside
LLM-generated copy. Whether they are real is the 140 lane's fabrication lens; flagged
there, not asserted here.

## Fix candidates, ordered by what closes the door

1. **Guard the write seam**: any writer about to store `content_data` whose top-level
   shape is exactly `{type, result}` with `type='text'` is storing a transport
   envelope, not content — decode it (through `ParseLLMJSONWithProvenance`) or refuse.
   Requires a census of `content_data` writers first (`section_editor_actions.go`
   `applyContentEdit`, the page-build save path, others — [UNMEASURED], the census is
   the first task for whoever picks this up). This is the class fix; it makes the bad
   state unrepresentable.
2. **One-off decode of `d2e9644b`** through the live parser (proven full recovery
   above) — only alongside (1); a one-off deletion is not a class fix (RFC_006
   precedent). Verify by re-rendering to a diff-clean article body.
3. **`25c73a1c` stays with its human**: it already has a `save_refused_incomplete` /
   `needs_human_review` work item (140 lane handoff, residual-rows table). No rerender
   can populate it — every declared field is absent from the recoverable fragment.
   Do not fire automated work at it.

## How to verify a fix

```sql
-- expect 0 after (1)+(2) land and (3) is resolved by a human:
SELECT count(*) FROM page_components
WHERE content_data ? 'type' AND content_data ? 'result'
  AND content_data->>'type' = 'text';
```

Plus the mutation test for (1): feed the guard a synthetic envelope row and confirm it
refuses/decodes; feed it a legitimate `content_data` that happens to contain a `type`
field among others and confirm it passes (the guard must key on the exact two-key
shape, not on the presence of `type`).

---

## CLAIMED 2026-08-04 — re-validated live, and TWO of this file's own statements are wrong

Picked up by a bug-sweep session. Ownership checked first: `who-owns.py`, `git log` on the
file path, and a grep of all 35 live session transcripts for `bugs_open/190` — **no session
and no workstream is working it.**

**The defect is still live**, numerator and denominator in one query, per this file's own
guard:

```sql
SELECT count(*) FILTER (WHERE content_data ? 'type' AND content_data ? 'result'
                          AND content_data->>'type'='text') AS envelope_rows,
       count(*) FILTER (WHERE content_data IS NOT NULL) AS denom_nonnull,
       count(*) AS denom_all
FROM page_components;
--  2 | 1054 | 1207      (was 2 / 1041 at filing — population grew, defect count did not)
```

`site_components` is **clean**: `0 | 54 | 54`. This file never checked the sibling table;
it is in scope for the guard but has no residue to repair.

### CORRECTION 1 — this is NOT inert residue. It RECURRED, one hour after this file was filed.

This file names `25c73a1c-b3af-48af-978b-95f7e500e8fa` as the gaswholesalers row. **That id
no longer exists.** The row serving that page today is
`17e7739e-68c0-4242-9804-a234a476795e`, with

```
created_at = updated_at = 2026-08-03 22:35:17.349534+00
```

— i.e. **created after this file was written** (~21:30Z), carrying the same envelope shape.
So the framing "two rows of historical residue, cause already fixed upstream" is wrong: a
write seam produced a *fresh* envelope row yesterday. The class fix (candidate 1) is not
tidy-up; it is the fix.

**The writer is named by the data**, not inferred — every envelope-shaped row in
`page_component_history` comes from one source:

```sql
SELECT source, count(*), min(created_at), max(created_at)
FROM page_component_history
WHERE content_data ? 'type' AND content_data ? 'result' AND content_data->>'type'='text'
GROUP BY source;
--  save_page_sections_overwrite | 65 | 2026-04-23 21:27:01 | 2026-08-03 22:35:17
```

That constant is written at `save_page_sections_action.go:595`.

**RESOLVED by reading the code, same session — and the answer narrows the bug.** The history
INSERT (`save_page_sections_action.go:586-601`) is
`SELECT pc.content_data FROM page_components WHERE pc.page_id = $1`, executed *before* the
DELETE. So it archives the state being **replaced**. The 65 are therefore **overwrite events
on a page that already carried an envelope — NOT 65 envelope writes.** Anyone quoting 65 as
a write count is wrong; I nearly was.

What the 65 *do* size is the historical blast radius, which is far larger than the 2 rows
this file reports:

```sql
SELECT count(*), count(DISTINCT page_id) AS pages, count(DISTINCT site_id) AS sites
FROM page_component_history
WHERE content_data ? 'type' AND content_data ? 'result' AND content_data->>'type'='text';
--  65 | 25 | 6
```

**25 distinct pages across 6 sites** have carried this poison at some point.
⚠ `count(DISTINCT component_id)` on that same query returns **0**, which is a NULL trap, not
an absence: `page_component_history_component_id_fkey` is `ON DELETE SET NULL`, so every
archived row whose component was later deleted has `component_id = NULL`. Same shape as the
`distinct_content = 0` trap already in `LANDMINES.md`. Use `page_id`.

**And the timeline says generation has STOPPED; what survives is PROPAGATION.** Every
envelope event since the parse fixes landed:

```sql
SELECT h.created_at, s.domain, p.name FROM page_component_history h
JOIN pages p ON p.id=h.page_id JOIN sites s ON s.id=h.site_id
WHERE h.content_data ? 'type' AND h.content_data ? 'result'
  AND h.content_data->>'type'='text' AND h.created_at > '2026-07-18';
--  exactly ONE row: 2026-08-03 22:35:17 | gaswholesalers.com | how-pricing-works
```

So: the three-tier parse fix did its job — no *new* envelope has been minted since mid-July.
But on 2026-08-03 the save seam archived an envelope and, 84ms later, **created a new row
carrying that same envelope forward** (`17e7739e`, `created_at 22:35:17.349`). The defect
that remains is that **`save_page_sections` will re-persist a transport envelope it is
handed, indefinitely**, so a poisoned row survives every rebuild rather than being cleaned
or refused by it. That is precisely what a guard at the write seam closes, and it is why
candidate (1) is still the fix — but the urgency is "stop propagating", not "stop minting".

> **CORRECTION to CORRECTION 1, three paragraphs up, same session:** I first wrote that "a
> write seam produced a *fresh* envelope row yesterday". Literally true — the row is new —
> but it overstates the mechanism, because the content was carried forward, not generated.
> Left visible rather than edited away, per the working-docs rule.

### CORRECTION 2 — the verification recipe above would MISS one of the two live rows

This file's § "How to verify a fix" instructs that the guard "must key on the exact two-key
shape". Measured — the two rows do **not** share a key set:

| row | top-level keys | note |
|---|---|---|
| `d2e9644b` (finetuning) | **`{content, result, type}`** — THREE keys | carries a real `content` key *alongside* the envelope keys |
| `17e7739e` (gaswholesalers) | `{result, type}` — two keys | |

An "exactly two keys" predicate is silent on `d2e9644b`, which is the row this file also
says is fully repairable. The discriminator has to be the envelope *signature*
(`type` == `'text'` **and** a string `result`), not the key count — and `d2e9644b` is the
subtle case for decode-vs-refuse, because content and envelope coexist in one map.

Logged in `WRONG_CALLS.md` as an inherited claim I re-checked rather than repeated.

---

## FIXED IN CODE 2026-08-04 — and STAYING OPEN, because the bar is fixed AND live

Candidate (1), the class fix, is built and at `HEAD`. **It is inert until the next chassis
roll**, so by CLAUDE.md's own bar — *"a fix committed but inert until the next roll stays
OPEN, because the defect is still reproducible until it ships"* — this file does not move to
`bugs_closed/` yet. Both live rows are still reproducible today.

### What was built

`platform/orchestration/actions/content_data_envelope_guard.go` (new), called at **both**
automated `content_data` write seams. Registered as **PBP-032**. Council correlation
`09bc4b3d-6721-4479-85b8-b5b56bf9b5d7`.

- **Discriminator:** `type == "text"` AND a **string** `result` — the producer's signature.
  **Not** the key count, for the reason in CORRECTION 2 above.
- **Decode** when `ParseLLMJSONWithProvenance` returns an object via `clean` or `repaired`
  (the only tiers that discard zero bytes). **Refuse** on every lossy tier, on a parse
  failure, on a non-object payload, and on a superset whose decoded and sibling values for
  one key disagree. Double-wrap defence capped at depth 3 (`bugs_closed/054` precedent).
- **Superset handling** (`d2e9644b`'s shape): real siblings preserved, `type`/`result`/`__`
  markers dropped, overlapping keys stored only when deep-equal.
- **No opt-in field**, and the RFC_010 argument is in the guard's header: that ruling fires
  when a branch is licensed by *caller identity*; this one is licensed by the *payload*.
- Refusals are countable: `agent_error_log.error_code = 'CONTENT_DATA_ENVELOPE'`.

### Commits — and note the second one is not mine

| commit | what |
|---|---|
| `ce675f019` | the guard, 14 tests, the truncation-coverage exemption, and the `ApplySectionEditAction` seam |
| `84b7d561c` | **the `save_page_sections` call site — swept into the `bugs_open/156` lane's commit** |
| `0e7b1c9e1` | concept register PBP-032 + index row (should have been in `ce675f019`; see `WRONG_CALLS`) |

I held the `save_page_sections` hunk back deliberately, because that file simultaneously
carried another session's uncommitted `156` dedup wiring which called an untracked file —
committing it would have broken `HEAD` fleet-wide. They then committed with a broad `add` and
took my hunk with them. Nothing is lost and forward-only holds (CLAUDE.md anticipates exactly
this), but **the wiring's provenance is a commit whose message is about a different bug**, so
`git log --grep` will not find it. Verified at `HEAD` rather than trusted:
`git archive HEAD` into a clean directory → both seams present, full `actions` suite green.

### Testing

14 tests, each naming the mutation that must break it, and **four mutations were actually run
against the shipped code**, each going red for the right reason and green again on restore:
predicate weakened to `type`-only; provenance rule dropped; predicate keyed on the exact
two-key shape (**this file's own recommended predicate**); seam ranging by value instead of
mutating in place. The no-op control asserts **byte identity** through `json.Marshal`, because
this guard's dominant behaviour is "change nothing" — and so is a completely broken one.

### What is still owed

1. **The roll.** Then pod-grep with a discriminating marker and a positive control:
   `strings /app/agent-chassis | grep -c sanitizeSectionsContentData` and
   `grep -c CONTENT_DATA_ENVELOPE` — both 0 before, ≥1 after. Do not roll to verify this
   alone; releases here are whole-fleet and the owner runs them.
2. **Repair `d2e9644b` (finetuning) through the framework, not by hand.** Its stored map has
   a real `content` key, so a scoped rerender should hit the guard's superset branch and come
   out clean through ordinary machinery. Back the row up first
   (`CREATE TABLE bak_pc_190_d2e9644b_<date> AS SELECT * …`). Only if the guard refuses on a
   conflict does this become a considered one-off `UPDATE`.
3. **`17e7739e` (gaswholesalers) is NOT ours to automate.** The guard will refuse every
   automated rebuild of it, permanently and by design, until a human repairs the page. It
   already carries a `needs_human_review` item. See the LANDMINE — the one-line "fix" for
   that recurring noise stores a 131-character fragment over a live page.
4. **Read the council verdict** on `09bc4b3d` and act on a REVISE/REJECTED; the code is
   already on the shared branch.
5. **Then** the count query in § "How to verify a fix" goes 2 → 1 → 0, the last step only
   when the human closes item 3. **A count of 1 is the expected post-repair state**, not a
   failed fix.

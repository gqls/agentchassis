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

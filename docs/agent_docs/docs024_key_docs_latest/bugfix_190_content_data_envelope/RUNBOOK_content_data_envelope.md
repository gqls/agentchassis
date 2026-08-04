# RUNBOOK — bugs_open/190, content_data transport envelope

Every command here had a gotcha attached when I first ran it. The gotcha is the reason the
line exists — change it HERE when it changes, not in your scrollback.

DB: `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`

---

## 1. Is the defect still live? (numerator AND denominator, one query)

```sql
SELECT count(*) FILTER (WHERE content_data ? 'type' AND content_data ? 'result'
                          AND content_data->>'type'='text') AS envelope_rows,
       count(*) FILTER (WHERE content_data IS NOT NULL) AS denom_nonnull,
       count(*) AS denom_all
FROM page_components;
--  2026-08-04:  2 | 1054 | 1207
```

**Gotcha — do not split this into two queries.** The filer of this bug was burned by a
count that returned 0 from an empty population, and left the guard in the file: a zero
numerator must never be readable without its denominator beside it. Same for the sibling
table, which nothing had checked before 2026-08-04:

```sql
SELECT count(*) FILTER (WHERE content_data ? 'type' AND content_data ? 'result'
                          AND content_data->>'type'='text') AS envelope_rows,
       count(*) FILTER (WHERE content_data IS NOT NULL) AS denom_nonnull, count(*) AS denom_all
FROM site_components;
--  2026-08-04:  0 | 54 | 54
```

## 2. What shape are the offending rows, exactly?

```sql
SELECT id, (SELECT array_agg(k ORDER BY k) FROM jsonb_object_keys(content_data) k) AS top_keys,
       jsonb_typeof(content_data->'result') AS result_type
FROM page_components
WHERE content_data ? 'type' AND content_data ? 'result' AND content_data->>'type'='text';
```

**Gotcha — the two rows do NOT share a key set**, and the bug file's own verification recipe
("the exact two-key shape") is wrong because of it. `d2e9644b` is
`{content, result, type}`; `17e7739e` is `{result, type}`. Any predicate keyed on key
*count* is silent on the first. Key on the signature: `type == 'text'` plus a **string**
`result`.

**Second gotcha:** `jsonb_object_keys` is set-returning, so selecting it bare gives you one
ROW PER KEY and it looks like duplicate rows. Wrap it in the scalar sub-select above.

## 3. Who wrote it? — and the two traps in this one query

```sql
SELECT source, count(*), count(DISTINCT page_id) AS pages, count(DISTINCT site_id) AS sites,
       min(created_at), max(created_at)
FROM page_component_history
WHERE content_data ? 'type' AND content_data ? 'result' AND content_data->>'type'='text'
GROUP BY source;
--  save_page_sections_overwrite | 65 | 25 | 6 | 2026-04-23 | 2026-08-03 22:35:17
```

**Trap 1 — 65 is NOT a write count, and reading it as one is the mistake this runbook exists
to stop.** The history INSERT (`save_page_sections_action.go:586-601`) selects
`pc.content_data` *before* the DELETE, so it archives the state being **replaced**. 65 =
overwrite events on pages that already carried an envelope. The question is answered by
reading the action's SQL, not by querying harder.

**Trap 2 — do not use `count(DISTINCT component_id)` here. It returns 0, and 0 means NULL.**
`page_component_history_component_id_fkey` is `ON DELETE SET NULL`, so every archived row
whose component was later deleted has a NULL `component_id`. Group by `page_id`. (Same shape
as the `distinct_content = 0` trap already in `LANDMINES.md`.)

## 4. Is anything still MINTING envelopes, or only propagating them?

```sql
SELECT h.created_at, s.domain, p.name
FROM page_component_history h JOIN pages p ON p.id=h.page_id JOIN sites s ON s.id=h.site_id
WHERE h.content_data ? 'type' AND h.content_data ? 'result' AND h.content_data->>'type'='text'
  AND h.created_at > '2026-07-18'          -- the date the three-tier parse fix landed
ORDER BY h.created_at DESC;
--  2026-08-04: exactly ONE row (2026-08-03 22:35, gaswholesalers how-pricing-works)
```

**Why the date filter is the whole point:** without it you cannot tell a defect that is
still being created from one that is merely being carried forward. Generation stopped in
mid-July; the single recent event is the save seam re-persisting an envelope it was handed.
Re-run this before claiming the fix's urgency either way.

## 5. Ownership — check ALL of these, they fail differently

```bash
python3 scripts/who-owns.py 190          # lagging: reads COMMITS, blind to a session mid-fix
git log --oneline -- bugs_open/190_*.md  # resolve by PATH, never by bare number
```

```bash
# live sessions that actually OPENED the file (not merely ran `ls bugs_open/`)
cd ~/.claude/projects/-home-ant-projects-agentchassis/
grep -o "\"file_path\":\"[^\"]*bugs_open/190_[^\"]*\"" $(ls -t *.jsonl | head -35) \
  | cut -d: -f1 | sort -u
```

**Gotcha — grepping transcripts for the bare string `bugs_open/190` is worthless.** Every
session that has ever run `ls bugs_open/` carries the whole bug list in its transcript, so
every bug scores several "live sessions". Match on the `file_path` of a tool call instead.

**And do not trust a bug file's own "OPEN, UNOWNED" header** — it is a snapshot of the day it
was typed. `bugs_open/181` says exactly that and is being worked by two live sessions.

## 6. Is the tree red because of ME? — the only check that settles it

The shared tree is frequently red on another session's uncommitted work, and "it is not my
change" is a claim. Build HEAD plus *only* your files, in a clean directory:

```bash
SP=<your scratchpad>
rm -rf $SP/headtest && mkdir -p $SP/headtest
git archive HEAD | tar -x -C $SP/headtest
cp platform/orchestration/actions/content_data_envelope_guard.go \
   platform/orchestration/actions/content_data_envelope_guard_test.go \
   platform/orchestration/actions/truncation_guard_test.go \
   platform/orchestration/actions/section_editor_actions.go \
   $SP/headtest/platform/orchestration/actions/
cd $SP/headtest && go test ./platform/orchestration/actions/
```

**Gotcha:** `go build` in the working tree is NOT this check — it compiles everyone's WIP
together, so it can be green when HEAD+yours is broken, and red when HEAD+yours is fine. Both
happened in one session on 2026-08-04.

**Second gotcha:** before committing a file, check it does not carry a passenger:
`git diff --numstat <file>` then read the diff. On 2026-08-04
`save_page_sections_action.go` held my guard call AND another session's `bugs_open/156` dedup
wiring, which called an **untracked** file — committing it would have broken HEAD fleet-wide.

## 7. Proving the guard can actually fail (mutation testing)

A green suite proves nothing here: this guard's happy path is "change nothing", and so is a
completely broken one. Back the file up, mutate, run the named test, restore, and diff.

```bash
SP=<your scratchpad>; G=platform/orchestration/actions/content_data_envelope_guard.go
cp $G $SP/guard.orig.go
# e.g. drop the provenance rule:
python3 - <<'EOF'
p='platform/orchestration/actions/content_data_envelope_guard.go'
s=open(p).read()
s=s.replace('if provenance != ProvenanceClean && provenance != ProvenanceRepaired {','if false {')
open(p,'w').write(s)
EOF
go test ./platform/orchestration/actions/ -run TestLossyProvenanceIsRefusedNotDecoded   # MUST fail
cp $SP/guard.orig.go $G && diff $SP/guard.orig.go $G && echo RESTORED
```

**Gotcha:** a mutation run can be defeated by an unrelated build break from another session —
mine was, once. If the mutated run reports a compile error in a file you did not touch, that
is not your test passing or failing; re-run when the tree compiles.

The four mutations that must each go red: predicate keyed on `type` alone; provenance rule
dropped; predicate keyed on `len(m)==2` (**this is the bug file's own recommended predicate**);
seam ranging by value instead of mutating in place.

## 8. Before committing platform code: the register check nobody's hook performs

The platform-seams ruling requires a shared mechanism's concept-register entry **in the same
commit that ships it**. No hook checks this — the `commit-msg` nudge and the `098` report both
check the *council* trailer, which is a different thing, and having that one makes you feel
compliant. One command, run before the commit:

```bash
git status --short docs/agent_docs/docs026_concept_register/register/
```

Non-empty, and in the same pathspec as the code, or you are shipping folklore. I missed this
on 2026-08-04 by exactly one commit.

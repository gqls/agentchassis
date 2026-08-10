# NOTES — bugs_open/223, the code index's blind spots are narrated as absence

Append-only, newest at the bottom. Technical log: what was tried, what the system
actually said, and every misstep.

Lane opened **2026-08-10** by a fresh session, on the owner's instruction to take the
next unowned bug in `bugs_open/`. 223 was named as the candidate.

---

## 2026-08-10 — ownership check before touching anything

`scripts/who-owns.py 223` reports no owning lane: the four commits on the bug file are
the *filing* and three *contributions* (a correction, a third failure mode from the
`bugfix_209` lane, a recurrence from the `bugfix_201` lane). The three lanes it names
(`bugfix_201`, `bugfix_209`, `bugfix_221`) each merely **cite** it — none has code in
flight against the mechanism. The bug file itself says **OPEN, UNOWNED** and names
`architecture_review` as the owning lane *for routing*, not as a thread already on it.

`who-owns.py` reads commits, so it is blind to a session mid-fix. Checked the live
transcripts too (`grep -c` over `~/.claude/projects/*/**.jsonl` for
`landmine.verifier|bugs_open/223`): seven other sessions mention it, and every hit is
incidental — one was reading the `scripts/landmines-*` toolchain while finishing
`bugs_open/232`, one hit is a `bugs_open/` directory listing, one is a pod list from
08-08. **No session is working the mechanism.** Taking it.

## 2026-08-10 — the bug is still valid, re-measured rather than believed

The bug's central claim is a census, so it is cheap to re-run and it is the claim's
single point of failure (the file says so itself). Both halves still hold today:

```sql
SELECT count(*) FILTER (WHERE path LIKE '%.go')     AS go,      -- 5837
       count(*) FILTER (WHERE path NOT LIKE '%.go') AS non_go,  --    0
       count(*)                                     AS total    -- 5837
  FROM code_symbols;

SELECT kind, count(*) FROM code_symbols GROUP BY 1 ORDER BY 2 DESC;
-- func 3653 | method 1119 | struct 987 | alias 42 | interface 36
```

682 distinct paths, all `.go`. **No `var`, no `const`, no `type`, no `doc` row exists.**
The corpus has grown since the bug was filed (5755 → 5837) and its composition has not
moved at all, which is the point: the blindness is structural, not a lag.

**A fact the bug file does not record, and it changes the fix.** The table's own CHECK
constraint already permits both missing kinds:

```
"code_symbols_kind_check" CHECK (kind = ANY (ARRAY['func','method','struct',
                                 'interface','alias','type','var','const']))
```

So `var`/`const` are **schema-legal and simply never written**. `\d code_symbols` is how
that was found — the schema-first rule paying for itself.

## 2026-08-10 — where the blindness actually becomes a false verdict

Read the whole chain rather than the symptom. The verifier is DB config on
`agent_definitions.type='landmine-verifier'`, five steps:

`load_entry` (query_database → `doc_notes`) → `derive_checks` (an LLM turns the entry's
`footprint` line into `code_checks` of kind `symbol|ls|content`) → `run_checks`
(`diagnose_code_lookup`, `max_checks:10`) → `verify` (an LLM returns
`{status: STILL_VALID|STALE|NEEDS_HUMAN_REVIEW, rationale, body}`) → `persist_verdict`
(`append_doc_note`, categories `['landmine-verification']`).

**Nothing parses the status mechanically.** The verdict is prose in `doc_notes`, read by
humans, by the next session, and — per RFC_005 §1 — fed into council seats'
`schema_hint`. Which is exactly why a false `STALE` is expensive: it argues for deleting
a correct entry, and it argues to a machine.

The damaging sentence is in the shared lookup action, not in the agent. From
`platform/orchestration/actions/diagnose_code_lookup_action.go`, `emptyAnswer("ls")`:

> `answered: 0 rows — no indexed path has that prefix, out of N indexed symbols. The
> query was RUN; this is not an unanswered question.`

For `scripts/landmines-sync.py` that sentence is **true and misleading**: the query did
run, and it could never have matched. The action's zero-row wording was written (for
`bugs_closed/108` defect B) to stop empty answers reading as silence — it succeeded, and
in doing so it now asserts the strongest available reading of a structurally empty
answer. That is the defect: not a missing guard, an **over-confident guard**.

The same file already contains the correctly-shaped precedent, one census along —
`bodyCoverageNote()`: *"source BODIES ARE NOT INDEXED (0 of N) … read those zeros as
UNKNOWN, not absent."* The fix is to generalise that, not to invent anything.

## 2026-08-10 — the seam is shared by four consumers, measured

```sql
SELECT type FROM agent_definitions
 WHERE default_config::text LIKE '%diagnose_code_lookup%'
   AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- feature-designer | fix-proposer | landmine-verifier
```

Plus a fourth, in Go and invisible to that query:
`diagnose_load_runtime_action.go:479` calls `answerCodeCheck` directly (the runtime
lane). **So any change to the answer wording is seen by the council's fix-proposer and
feature-designer seats too** — they must be told, not merely measured (owner ruling
2026-07-29 §3).

## 2026-08-10 — the write side, and why `var` is missing

`internal/analysis/analyse.go:140` walks the AST and takes `*ast.FuncDecl` plus
`*ast.GenDecl` **only when `d.Tok == token.TYPE`**. `token.VAR` and `token.CONST` are
dropped there, `internal/analysis/types.go:Output` has no field for them, and
`code_symbols_actions.go` therefore has nothing to write. Meanwhile the READER
(`codeKindList` in `diagnose_code_lookup_action.go`) already lists `var` and `const` as
code kinds. Reader and writer disagree, and the writer is the half that is behind.

Rough volume, for sizing: `grep -rhE "^(var|const) " --include=*.go platform/ internal/
pkg/ cmd/ | wc -l` → **930** declaration lines (a lower bound on specs — a `const (…)`
block holds many).

[UNMEASURED] whether adding those kinds would trip the per-kind prune floor on its first
run — `prune_floor.go` cohorts by `kind=`, and a brand-new kind arrives with
stored=0. To be measured before any such change ships, not argued.

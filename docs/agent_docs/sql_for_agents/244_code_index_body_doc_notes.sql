-- ============================================================================
-- 244_code_index_body_doc_notes.sql
--
-- The travelling-docs trail for the two actions 243 changed. tooling_provenance
-- objected during the council round that there was NO history to consult for
-- this work — measured: zero doc_notes and zero doc_plans rows for any
-- subject_key matching code_symbol / code_lookup / diagnose_code. That was true,
-- and the answer to "there is no trail" is to leave one.
--
-- EXTENDS the trail opened 2026-07-27 19:38 (rows c0098a43 on index_code_symbols
-- and d89f31b6 on diagnose_code_lookup, which record the DEFECT). These rows
-- record the FIX and the landmines it exposed. doc_notes is append-only, so the
-- earlier rows stay and the pair reads in order — do not update them in place,
-- and do not start a third trail under a new subject_key.
--
-- doc_notes is one of the runner's declared idempotent sinks (an unguarded
-- INSERT here is harmless on replay), so this file carries a guard for
-- readability rather than for safety: a replay would add duplicate notes, which
-- is untidy, not damaging.
-- ============================================================================

BEGIN;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM doc_notes
        WHERE subject_key = 'index_code_symbols'
          AND body LIKE '%out.Root%'
    ) THEN
        RAISE EXCEPTION '244: the body-column notes are already present — nothing to add. Re-read the trail rather than forcing this.';
    END IF;
END $$;

INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES (
    'action',
    'index_code_symbols',
    $note$2026-07-27, later the same day (D11 layer 1, council 18fe4035, migration 243): the indexer now writes SYMBOL BODIES as well as declarations. code_symbols.body holds the source text for lines [line_start,line_end], sliced by internal/analysis.SliceLines — the function that already existed, exported rather than re-extracted, so the body a reader slices on demand and the body the indexer stored cannot silently differ.

THE LANDMINE, and it nearly sank the design. Bodies are read from out.Root, which is a LOCAL checkout. That is only true because the LIVE code-indexer workflow's first step is analyse_repo_local, which fetches the tarball into this pod's own temp dir and deliberately does not clean it up. Under the ORIGINAL wiring in seed 118 (request_repo_analysis), the analyser adapter parses in its OWN pod and returns line spans whose root does not exist here: every read would fail, every body would be NULL, and the change would be inert while looking shipped. THE SEED FILE STILL SHOWS THE OLD WIRING. The council-approved plan asserted "the indexer is already walking the file" — it is not; it walks a JSON-decoded analysis.Output that carries spans and no source text. Reading the LIVE agent_definitions row is what caught it. If this action ever reports with_body=0 while symbols>0, check the workflow's analyse step FIRST.

body is a *string in Go and NULL in SQL when it could not be sliced — never "". "Could not read it" and "genuinely empty" must stay distinguishable, which is the same empty-vs-absent confusion bugs_open/108 is about, one layer down.

body is assigned plainly in the upsert, never COALESCEd onto the previous value the way embedding is. content_hash covers DECLARATION text only (kind + symbol + signature + doc + path), so a function whose body changed while its signature did not has an UNCHANGED hash: there is no cheap test for "this stored body is still current". line_start/line_end are overwritten from EXCLUDED regardless, so a preserved body would end up contradicting the very span it claims to be.

body is a SEPARATE COLUMN and is absent from content_hash on purpose. content does three jobs at once — trigram search text, the EMBEDDED text, and the hash input that triggers re-embedding — so folding bodies in would have re-embedded and re-skewed all 4,535 vectors as an invisible side effect of a search fix. 243_code_symbols_body_column_VERIFY.sql asserts that invariant as code.

with_body is reported in the action result beside upserted, so a run that indexed everything and sliced nothing is legible from the orchestration record without anyone thinking to query the column.$note$,
    '["code-index", "architecture-review", "bugs_open-108", "council-gate", "landmine", "d11-layer-1"]'::jsonb,
    'architecture_review workstream — implementation of council-approved plan 18fe4035',
    'architecture council 2'
);

INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES (
    'action',
    'diagnose_code_lookup',
    $note$2026-07-27, later the same day (D11 layer 1, council 18fe4035, migration 243): the reader side is fixed, and the package doc at :29-31 is now TRUE.

WHAT CHANGED. kind=content searches (COALESCE(body,'') ILIKE ... OR content ILIKE ...) — body OR declaration, not body alone, because declaration matches were the kind's only WORKING use until now and dropping them silently would have broken checks that currently succeed. COALESCE rather than a bare body predicate so rows indexed before the column existed still match on content instead of vanishing. Each hit is marked [body] or [decl]: a reviewer who cannot tell a doc-comment mention from an implementation will read the first as the second. The excerpt is taken AROUND the match (guardian's low objection) — against a 200-line body, truncating from the head returns a prefix the reviewer cannot read.

AN EMPTY RESULT IS NOW AN ANSWER. "answered: 0 rows — searched 4,535 indexed symbols" instead of an empty section. This is the root cause bugs_open/108 names: the prompt that tells a reviewer to treat an empty answer as UNKNOWN is NOT in front of them when they read the result, so the distinction has to travel with the data. codeIndexScope carries it, is read once per action run, and is used by BOTH lanes that answer code checks — diagnose_code_lookup (the council's verify tier) and diagnose_load_runtime (the diagnosis lane, whose verdicter's cite-or-abstain acts directly on absence). One judgement, not a sibling copy that drifts (016b section 9).

THE DEGRADE IS NAMED, NOT SILENT. Between migration 243 applying and the chassis image rolling, the column exists and every row is NULL. Content checks fall back to declaration-only behaviour — not broken, but a reviewer must be told or they will read the degrade as absence. The coverage note says "source BODIES ARE NOT INDEXED (0 of N) ... read those zeros as UNKNOWN, not absent" for exactly that window, and says so again for any partial coverage.

STILL NOT FIXED HERE: council-gate has no code_lookup step at all, so its authors get no code answers and its verdict note cannot distinguish "searched, found nothing" from "nobody ran the query". That asymmetry is DELIBERATE (099_SYNC_gate_roster.py:24-29 — code_lookup/repropose serve the blind reproposer, which the gate has no equivalent of), so the remedy is to surface code results into the verdict note, not to bolt a reproposer-shaped step onto a lane with no reproposer. Recorded in bugs_open/108 as owed. Also still unreachable: markdown. 0 of 4,535 index rows are markdown, so WRONG_CALLS.md, /bugs_open/ and the concept register remain invisible to every seat — that needs the kind CHECK constraint relaxed, which is its own migration.$note$,
    '["code-index", "architecture-review", "bugs_open-108", "council-gate", "landmine", "d11-layer-1"]'::jsonb,
    'architecture_review workstream — implementation of council-approved plan 18fe4035',
    'architecture council 2'
);

COMMIT;

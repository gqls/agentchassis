-- 285_diagnose_dispatch_loop_forwards_seed_scope.sql
--
-- bugs_open/174 — the diagnosis dispatch loop drops `seed_scope` (and
-- `runtime_page`), so a diagnosis aimed at chosen symbols silently runs against
-- whatever the code search happened to return.
--
-- THE DEFECT
-- ----------
-- `090_TRIGGER_needs_diagnosis_v1.sh` accepts `SEED_SCOPE`, documents it, writes
-- it correctly into the work item spec as a JSON array, and keys its own
-- pre-dispatch coverage probe on it. `diagnose-orchestrator`'s `call_diagnoser`
-- forwards it. The work item still holds it afterwards. It simply never reaches
-- the agent, because the hop in front of the orchestrator does not carry it.
--
-- There are TWO hand-maintained allow-lists in series on that hop, not one, and
-- the ticket's fix candidate 1 named only the second:
--
--   1. `claim_item`'s SQL RETURNING clause — a projection of nine spec keys.
--      `seed_scope` and `runtime_page` are not among them, so `claimed.seed_scope`
--      does not exist at all.
--   2. `call_handler`'s `input_mapping` — a ten-key allow-list. `input_mapping` is
--      an allow-list, not a passthrough (see 274, which fixed the same mechanism
--      one hop away for `fidelity`), so an unlisted key is dropped in silence.
--
-- Adding the key to (2) alone would map from a source path that resolves to
-- nothing, and — because the key is optional — would be dropped again without a
-- word. Both lists have to change together, which is the point: they agree with
-- each other today and disagree with the callee's declared `input_contract`,
-- which lists `seed_scope` and `runtime_page` among the ten keys
-- `diagnose-orchestrator` accepts.
--
-- WHY THE FAILURE IS INVISIBLE RATHER THAN LOUD
-- ---------------------------------------------
-- `diagnose_assemble_bundle`'s scope resolution is a fallback chain BY DESIGN:
-- loop scope -> `input_data.seed_scope` -> `lookup_code_symbols`' code_results.
-- With the seed confiscated, step 2 finds nothing and step 3 quietly supplies a
-- different, plausible scope. The action cannot tell "the caller gave no seed"
-- from "the seed was taken in transit", so it correctly does not complain. A
-- fallback chain converts a lost parameter into a successful run with different
-- inputs.
--
-- Measured on the intake table, 2026-07-28 -> 2026-08-01: of four intakes that
-- ever carried a non-empty `seed_scope`, THREE were claimed by this loop and lost
-- it; the only survivor was a manual `DISPATCH=1` publish fired to work around
-- the bug. Two of the three losses are other lanes' real work (`bugs_open/155`,
-- and a scheduler diagnosis from 07-28) — both aimed at chosen symbols, both
-- silently re-aimed, and neither author had any way to know.
--
-- THE CHANGE
-- ----------
-- (1) Two more projections on `claim_item`'s RETURNING. `seed_scope` uses `->`
--     (jsonb) rather than `->>` deliberately — see the TYPE note below.
-- (2) Two more optional keys on `call_handler`'s `input_mapping`, in the
--     `claimed.*` form the other ten already use.
--
-- Both are optional/additive: an intake with no `seed_scope` projects SQL NULL,
-- the optional mapping key resolves to null, `ExtractStringListHelper` returns
-- nil for it, and the code_results fallback runs exactly as it does today. That
-- negative control is pinned in Go by
-- `TestScopeSource_NoSeedFallsThroughToCodeResults`.
--
-- TYPE — AND WHY THIS MIGRATION IS INERT WITHOUT THE CODE HALF
-- -----------------------------------------------------------
-- `QueryDatabaseAction` scans every column into `interface{}` and stringifies any
-- `[]byte` it gets back (`database_actions.go`, "if b, ok := values[i].([]byte)").
-- A jsonb column therefore arrives in collected_data as the STRING `["a","b"]`,
-- not as a list, and `ResolveInputMapping` passes values through unchanged, so it
-- reaches the action still a string. Before the accompanying code change,
-- `ExtractStringListHelper` returned nil for a string — so this migration ALONE
-- would have dropped the seed a third time, in a new place, just as silently.
--
-- ORDERING: this migration is inert but harmless until the chassis carrying the
-- `ExtractStringListHelper` change is live. Applied first, a seeded diagnosis
-- behaves exactly as it does today (falls through to code_results); it cannot
-- half-enable anything, and it cannot break an unseeded one. Per CLAUDE.md's
-- 2026-07-29 ruling I claim NO ordering constraint — HEAD is shared and any
-- session's roll ships the code half regardless of what I do here.
--
-- VERIFY AFTER APPLYING (the mapping, then the BEHAVIOUR — they come apart):
--   -- (a) both lists now agree with the callee's contract:
--   SELECT default_config #>> '{workflow,steps,call_handler,config,input_mapping}',
--          default_config #>> '{workflow,steps,claim_item,config,query}'
--     FROM agent_definitions WHERE type='diagnose-dispatch-loop' AND is_active
--       AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
--   -- (b) fire 090 WITHOUT DISPATCH=1 with a SEED_SCOPE, let the loop claim it,
--   --     then assert the agent actually RECEIVED it. Use the run correlation
--   --     stamped onto the item as spec.dispatch_correlation_id, NOT the intake
--   --     correlation — they differ:
--   SELECT collected_data->'input_data'->'seed_scope'
--     FROM orchestration_states WHERE owner_agent_type='diagnose-agent'
--      AND correlation_id='<run corr>';       -- must be the value, not absent
--   -- (c) assert the EFFECT, not the field. Field-present is a weaker claim than
--   --     scope-used, and the fallback chain is exactly what pulls them apart:
--   SELECT collected_data->'assembled'->>'scope_source'
--     FROM orchestration_states WHERE correlation_id='<run corr>';  -- 'seed'
--   -- and the bundle's "## In-scope code" must contain the symbols you named.
--
-- ROLLBACK: the snapshot below is taken first and holds the pre-change config;
--   UPDATE agent_definitions a SET default_config = b.default_config
--   FROM agent_definitions_backup b
--   WHERE a.type='diagnose-dispatch-loop' AND b.type=a.type
--     AND a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
--     AND b.snapshot_taken_at = (SELECT max(snapshot_taken_at) FROM agent_definitions_backup
--                                WHERE type='diagnose-dispatch-loop');
--
-- LANDMINES honoured here (docs024_key_docs_latest/LANDMINES.md):
--   * `snapshot_agent(text,text)` writes to `agent_definitions_backup`; the
--     one-arg form writes an `is_snapshot=true` row into `agent_definitions`.
--     Checking the wrong table makes a good snapshot look like a no-op — so the
--     assertion below reads `agent_definitions_backup` and checks it holds the
--     PRE-change value, not merely that a row exists.
--   * `jsonb_set(..., create_if_missing := false)` on a wrong path is a SILENT
--     no-op. The mapping keys are new by design so `true` is correct, which means
--     a typo in the PATH would silently create a useless key instead of erroring —
--     hence the post-write assertions that read the keys back FROM the resolved
--     path and compare their values.
--   * The RETURNING edit is a `replace()` on a query string, which is a silent
--     no-op if the needle does not match byte-for-byte. The assertion after it
--     therefore checks the two new projections are PRESENT rather than trusting
--     the UPDATE's row count, which counts rows touched, not text changed.

\set ON_ERROR_STOP on

BEGIN;

-- 1. Snapshot first (two-arg form -> agent_definitions_backup).
SELECT snapshot_agent(
    'diagnose-dispatch-loop',
    'pre 285: forward seed_scope + runtime_page from the claimed spec to the handler (bugs_open/174)'
) AS snapshot_id;

-- 2. Assert the snapshot holds the PRE-change config. If it already carries
--    seed_scope, an earlier run applied this and the file is a no-op re-run.
DO $$
DECLARE
    snap_has_seed boolean;
BEGIN
    SELECT (default_config #>> '{workflow,steps,call_handler,config,input_mapping}') LIKE '%seed_scope%'
      INTO snap_has_seed
    FROM agent_definitions_backup
    WHERE type = 'diagnose-dispatch-loop'
    ORDER BY snapshot_taken_at DESC
    LIMIT 1;

    IF snap_has_seed IS NULL THEN
        RAISE EXCEPTION '285: no snapshot row found for diagnose-dispatch-loop — refusing to change a live agent with no rollback point';
    END IF;
    IF snap_has_seed THEN
        RAISE NOTICE '285: snapshot already contains seed_scope — this appears to be a re-run; continuing (idempotent)';
    END IF;
END $$;

-- 3a. ALLOW-LIST ONE: project the two spec keys the claim never returned.
--
--     `spec->'seed_scope'` uses -> (jsonb), not ->> (text). Both reach Go as the
--     same JSON text through QueryDatabaseAction's []byte stringification, so
--     this is about honesty of type at the SQL boundary rather than behaviour:
--     the value IS a JSON array, and a later reader who changes the scanner
--     should find a jsonb here, not a text cast that hides it.
--     `runtime_page` is a scalar, so ->> is correct for it.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,claim_item,config,query}',
        to_jsonb(
            replace(
                default_config #>> '{workflow,steps,claim_item,config,query}',
                'spec->>''correlation_id'' AS correlation_id',
                'spec->>''correlation_id'' AS correlation_id, spec->''seed_scope'' AS seed_scope, spec->>''runtime_page'' AS runtime_page'
            )
        ),
        false   -- the key exists; create_if_missing=false makes a wrong path ERROR rather than invent one
    ),
    updated_at = NOW()
WHERE type = 'diagnose-dispatch-loop'
  AND is_active
  AND COALESCE(is_snapshot, false) = false
  AND deleted_at IS NULL
  -- Idempotent: skip if a previous run already added the projection.
  AND (default_config #>> '{workflow,steps,claim_item,config,query}') NOT LIKE '%AS seed_scope%';

-- 3b. ALLOW-LIST TWO: forward them, in the claimed.* form the other ten use.
UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
            default_config,
            '{workflow,steps,call_handler,config,input_mapping,seed_scope?}',
            '"claimed.seed_scope"'::jsonb,
            true   -- new key by design
        ),
        '{workflow,steps,call_handler,config,input_mapping,runtime_page?}',
        '"claimed.runtime_page"'::jsonb,
        true       -- new key by design
    ),
    updated_at = NOW()
WHERE type = 'diagnose-dispatch-loop'
  AND is_active
  AND COALESCE(is_snapshot, false) = false
  AND deleted_at IS NULL;

-- 4. Assert BOTH lists landed, at the intended paths, with the intended values —
--    and that they now agree with the callee's declared input_contract, which is
--    the invariant this bug was actually about.
DO $$
DECLARE
    rows_live       integer;
    q               text;
    im              jsonb;
    callee_declared text[];
    missing         text[];
BEGIN
    -- Counted separately from the values: there is no min(jsonb) aggregate, and
    -- reaching for one is how this file failed its first run.
    SELECT count(*) INTO rows_live
    FROM agent_definitions
    WHERE type = 'diagnose-dispatch-loop'
      AND is_active
      AND COALESCE(is_snapshot, false) = false
      AND deleted_at IS NULL;

    IF rows_live <> 1 THEN
        RAISE EXCEPTION '285: expected exactly 1 live diagnose-dispatch-loop row, found %', rows_live;
    END IF;

    SELECT default_config #>> '{workflow,steps,claim_item,config,query}',
           default_config #>  '{workflow,steps,call_handler,config,input_mapping}'
      INTO q, im
    FROM agent_definitions
    WHERE type = 'diagnose-dispatch-loop'
      AND is_active
      AND COALESCE(is_snapshot, false) = false
      AND deleted_at IS NULL;

    -- (a) the RETURNING projection — replace() is a silent no-op on a miss
    IF q NOT LIKE '%AS seed_scope%' THEN
        RAISE EXCEPTION '285: claim_item RETURNING does not project seed_scope — the replace() needle did not match';
    END IF;
    IF q NOT LIKE '%AS runtime_page%' THEN
        RAISE EXCEPTION '285: claim_item RETURNING does not project runtime_page — the replace() needle did not match';
    END IF;

    -- (b) the input_mapping keys, read back from the resolved path
    IF (im ->> 'seed_scope?') IS DISTINCT FROM 'claimed.seed_scope' THEN
        RAISE EXCEPTION '285: call_handler.input_mapping seed_scope? = %, expected claimed.seed_scope', im ->> 'seed_scope?';
    END IF;
    IF (im ->> 'runtime_page?') IS DISTINCT FROM 'claimed.runtime_page' THEN
        RAISE EXCEPTION '285: call_handler.input_mapping runtime_page? = %, expected claimed.runtime_page', im ->> 'runtime_page?';
    END IF;

    -- (c) THE INVARIANT, not just the two keys: every key the handler declares it
    --     accepts must now be forwardable by the loop. This is what makes the fix
    --     a closed door rather than two more entries on a list that drifted once
    --     and will drift again.
    SELECT array_agg(k) INTO callee_declared
    FROM agent_definitions,
         LATERAL jsonb_array_elements_text(
             COALESCE(input_contract->'required','[]'::jsonb) || COALESCE(input_contract->'optional','[]'::jsonb)
         ) AS k
    WHERE type = 'diagnose-orchestrator'
      AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

    SELECT array_agg(d) INTO missing
    FROM unnest(callee_declared) AS d
    WHERE NOT (im ? d OR im ? (d || '?'));

    IF missing IS NOT NULL THEN
        RAISE EXCEPTION '285: diagnose-orchestrator declares %, which call_handler still cannot forward', missing;
    END IF;

    RAISE NOTICE '285: OK — claim_item projects seed_scope + runtime_page, call_handler forwards them, and the mapping now covers every key diagnose-orchestrator declares';
END $$;

COMMIT;

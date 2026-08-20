-- 503 — RFC_040 stage 1: the table a running binary writes its capabilities into.
--
-- OWNER-RATIFIED 2026-08-20, and SCOPED: RFC_040's recording half only. There is
-- deliberately NO assert_live_capability() function here and no migration calls
-- one. The owner's reasoning, recorded in RFC_040 §0: a fail-closed assertion
-- helper with exactly ONE caller is a mechanism nobody exercises, which is this
-- estate's own documented failure mode. Make the fact durable first; let the
-- second real demand shape the assertion. **A future author adding the helper
-- should be able to name two migrations that want it.**
--
-- ── WHY ──────────────────────────────────────────────────────────────────────
-- Go changes are inert until an image is rolled; DB config is live immediately.
-- So a config change that names a behaviour in its code half must be able to ask
-- "is the new code live yet?" BEFORE it applies. Measured 2026-08-19 over this
-- directory: 32 migrations assert a binary precondition in prose, ZERO can
-- verify one — SQL in postgres-clients-0 cannot reach a pod.
--
-- The documented human check frequently cannot be performed at all: the
-- `build provenance` line is a STARTUP line (measured gone from a FULL
-- `kubectl logs` three hours after a roll) and buildinfo.GitCommit is ONE string
-- rather than an ancestry, so grepping the binary for your own commit returns
-- ABSENT for a binary that certainly contains it. Two lanes have been burned by
-- exactly that (bugs_open/215 on v1.0.1288; bugs_open/299 on v1.0.1316).
--
-- This table holds the answer that HAS no shelf life: which named capabilities
-- are in the artefact that is running.
--
-- ── READING IT (the two rules that keep it honest) ───────────────────────────
-- 1. **ALWAYS filter on last_seen_at.** A dead pod's rows sit here looking
--    current until something prunes them. A reader that ignores staleness will
--    conclude a binary is running that is not — the SAME class of error this
--    table exists to end, reintroduced by the table itself.
-- 2. **EVERY pod, not ANY pod.** During a partial roll two replicas differ, and
--    the run that fails is the one that lands on the old one ("logs deploy/X
--    reads one pod of N", LANDMINES.md). A presence check that stops at the
--    first matching row is wrong.
--
-- Worked read — is a capability live on every fresh pod of a service?
--   SELECT count(*) FILTER (WHERE kind='discovery_check' AND name='cta_nonpage_destination') AS have,
--          count(DISTINCT pod_name) AS pods
--     FROM service_binary_capabilities
--    WHERE service='agent-chassis' AND last_seen_at > now() - interval '10 minutes';
--   -- `have` must equal `pods`, and `pods` must equal the live replica count.
--
-- ── WHAT IS NOT HERE, AND WHY ────────────────────────────────────────────────
-- **image_tag.** RFC_040 §2.1 sketched it; it is omitted because it is NOT
-- obtainable from inside the pod — checked 2026-08-20, the chassis container's
-- environment carries HOSTNAME and AGENT_TYPE and no image tag at all. A column
-- that could only ever be '' is worse than no column: it reads as "we record
-- this" and answers nothing. A future author with a real source can add it.
--
-- ROLLBACK: DROP TABLE service_binary_capabilities;  -- nothing reads it, by design.

BEGIN;

CREATE TABLE IF NOT EXISTS service_binary_capabilities (
  service      text        NOT NULL,   -- 'agent-chassis', … the DEPLOYMENT's name, not AGENT_TYPE
  pod_name     text        NOT NULL,   -- HOSTNAME; unique per roll, so successors never collide
  git_commit   text        NOT NULL,   -- buildinfo.GitCommit verbatim ('unknown' if built off-makefile)
  kind         text        NOT NULL,   -- 'build' (the sentinel) | 'discovery_check' | 'action'
  name         text        NOT NULL,
  started_at   timestamptz NOT NULL DEFAULT now(),
  last_seen_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (service, pod_name, kind, name)
);

-- The read above is (service, last_seen_at) filtered then grouped; without this
-- it is a seq scan over every pod that ever reported.
CREATE INDEX IF NOT EXISTS idx_sbc_service_last_seen
  ON service_binary_capabilities (service, last_seen_at DESC);

-- Induced verification (a DO/RAISE, because ON_ERROR_STOP ignores a non-empty
-- SELECT result and a verify block of SELECTs cannot stop the COMMIT):
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n
    FROM information_schema.columns
   WHERE table_name = 'service_binary_capabilities'
     AND column_name IN ('service','pod_name','git_commit','kind','name','started_at','last_seen_at');
  IF n <> 7 THEN
    RAISE EXCEPTION '503: expected 7 columns on service_binary_capabilities, found % — the table pre-existed in a different shape; investigate before assuming this applied', n;
  END IF;

  SELECT count(*) INTO n
    FROM pg_indexes
   WHERE tablename = 'service_binary_capabilities' AND indexname = 'idx_sbc_service_last_seen';
  IF n <> 1 THEN
    RAISE EXCEPTION '503: the (service, last_seen_at) index is missing — the documented read would seq-scan';
  END IF;
END $$;

COMMIT;

-- Post-apply expectation: ZERO rows until a chassis carrying the writer rolls.
-- An empty table is the correct state on the day this applies, and must NOT be
-- read as the mechanism failing — the Go half ships in the same commit but is
-- inert until the image is rebuilt and rolled. First real evidence to look for
-- after a roll (with the negative control, which is the point):
--   SELECT service, pod_name, git_commit, count(*) FILTER (WHERE kind='discovery_check') AS checks
--     FROM service_binary_capabilities GROUP BY 1,2,3;
--   -- expect one row per live chassis pod; `checks` > 0; and a name you know is
--   -- NOT in the binary must be absent from the name column.

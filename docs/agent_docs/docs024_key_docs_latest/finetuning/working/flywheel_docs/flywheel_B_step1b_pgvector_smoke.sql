-- ============================================================================
-- Flywheel B — Step 1b: Prove pgvector retrieval end-to-end
-- ============================================================================
-- Before we touch agents, confirm: given embeddings in knowledge_base, we can
-- retrieve by cosine similarity. We use synthetic vectors so this test is
-- independent of Ollama — it only tests the pgvector path.
--
-- Run inside a transaction and rollback at the end so nothing persists.
--
-- Run: kubectl -n ai-persona-system exec -it deploy/postgres-clients -- \
--        psql -U clients_user -d clients_db
-- ============================================================================

BEGIN;

-- Insert three rows into a temp collection 'flywheel_b_smoke' with crafted
-- embeddings. In 3D-representative space, vectors [1,0,0], [0,1,0], [0,0,1]
-- are orthogonal. We pad to 768 dimensions.
--
-- Row A: "cat" — embedding peaks on dim 0
-- Row B: "dog" — embedding peaks on dim 1
-- Row C: "piano" — embedding peaks on dim 2

-- Helper: build a 768-float vector with 1.0 at position N and 0.0 elsewhere
-- vector_cosine_ops needs non-zero vectors, so we add a tiny baseline.
WITH inserts AS (
    SELECT
        'cat' as title,
        'Cats are small domesticated carnivorous mammals.' as content,
        (array_fill(0.01::real, ARRAY[768])::vector
         + (array_fill(0.0::real, ARRAY[1]) || array_fill(0.99::real, ARRAY[1]) || array_fill(0.0::real, ARRAY[766]))::real[]::vector
        ) as emb
    UNION ALL
    SELECT
        'dog',
        'Dogs are domesticated descendants of wolves.',
        (array_fill(0.01::real, ARRAY[768])::vector
         + (array_fill(0.0::real, ARRAY[2]) || array_fill(0.99::real, ARRAY[1]) || array_fill(0.0::real, ARRAY[765]))::real[]::vector
        )
    UNION ALL
    SELECT
        'piano',
        'A piano is a large keyboard musical instrument.',
        (array_fill(0.01::real, ARRAY[768])::vector
         + (array_fill(0.0::real, ARRAY[3]) || array_fill(0.99::real, ARRAY[1]) || array_fill(0.0::real, ARRAY[764]))::real[]::vector
        )
)
INSERT INTO knowledge_base (collection, title, content, embedding, embedding_model, metadata)
SELECT 'flywheel_b_smoke', title, content, emb, 'synthetic-test', '{"test":true}'::jsonb
FROM inserts;

-- Confirm they're in
SELECT title, LEFT(content, 40) as preview,
       embedding IS NOT NULL as has_embedding
FROM knowledge_base
WHERE collection = 'flywheel_b_smoke'
ORDER BY title;

-- Query: use a vector close to "cat" (peak on dim 1). Expected: cat wins.
WITH query_vec AS (
    SELECT (array_fill(0.01::real, ARRAY[768])::vector
            + (array_fill(0.0::real, ARRAY[1]) || array_fill(0.99::real, ARRAY[1]) || array_fill(0.0::real, ARRAY[766]))::real[]::vector
           ) as v
)
SELECT title,
       LEFT(content, 40) as preview,
       ROUND((1 - (embedding <=> (SELECT v FROM query_vec)))::numeric, 4) as similarity
FROM knowledge_base
WHERE collection = 'flywheel_b_smoke'
ORDER BY embedding <=> (SELECT v FROM query_vec)
LIMIT 3;

-- Expected output order: cat > (dog, piano ~equal distant)

ROLLBACK;

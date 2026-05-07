# PATCH — 2026-05-06 — Corrections to HANDOFF_2026-04-23 § "Alternative to `kubectl exec`"

The credential-free dataset-pull path documented in the original handoff was
not exercised before being written down. When run on 2026-05-06 it produced
two concurrent failures that masked each other; both are described below
along with the corrected command that has now been validated end-to-end
(1,958 rows, 21 MB, head and tail both parse as valid JSON).

## Bug 1 — `COPY ... TO STDOUT` mangles JSONB output

`COPY` in the default TEXT format applies its own escape layer (backslash
doubling, control-character escaping) on top of whatever the column already
contains. For a `jsonb` column whose values are already escape-bearing JSON,
the two layers collide: the resulting bytes look correct on casual inspection
but fail JSON parsing the moment the parser hits the first internal `\n`
or `\"`. The original symptom was `JSONDecodeError: Expecting ',' delimiter`
at character 2310 of line 1 — i.e. the first internal escape.

**Fix:** drop `COPY` and use `psql -tAXc` with a plain `SELECT`. In `-tA`
mode psql writes each tuple's text representation verbatim, separated by
single newlines — which is exactly JSONL. `-X` additionally suppresses
`~/.psqlrc` so a local psql config can't leak formatting into the stream.

## Bug 2 — `kubectl exec -i` truncates the output stream

The original used `kubectl exec -i …`. The `-i` flag opens an interactive
stdin we never write to. Under load this produced sporadic
`next reader: unexpected EOF` errors and a truncated stream — 1,716 lines
arrived instead of the expected 1,958, with no other error signal.

**Fix:** drop `-i` for non-interactive commands. The two errors are
correlated (the EOF appears to come from the stdin half of the duplex
closing before stdout is fully drained), and removing `-i` resolves both.

## Corrected command (validated 2026-05-06)

```bash
# On laptop, in any working directory
kubectl -n ai-persona-system exec postgres-clients-0 -- \
    psql -U clients_user -d clients_db -tAXc "
SELECT jsonb_build_object('messages', messages, 'metadata', metadata)
FROM training_exports.rows
WHERE export_id = '146a9a12-c953-48eb-bf1f-c1856e5f13b7'::uuid
ORDER BY row_index
" > training_iter0.jsonl

# Local checks BEFORE transfer
wc -l training_iter0.jsonl    # expect 1958
ls -lh training_iter0.jsonl   # expect ~21M
head -1 training_iter0.jsonl | python3 -c \
    "import json,sys; d=json.loads(sys.stdin.read()); print('keys:', list(d.keys()), 'msgs:', len(d['messages']))"
tail -1 training_iter0.jsonl | python3 -c \
    "import json,sys; json.loads(sys.stdin.read()); print('tail OK')"

# Transfer to GPU VM
tnr scp training_iter0.jsonl 0:/home/ubuntu/training_iter0.jsonl
```

Expected output of the head/tail parse:

```
keys: ['messages', 'metadata'] msgs: 2
tail OK
```

## Two lessons worth folding into FOCUS § 14

These are general enough that they don't belong only in the handoff:

- **`COPY TO STDOUT` is not a JSON-safe transport for `jsonb` columns.**
  Use `psql -tAXc` with a plain `SELECT` when you want JSONL output. The
  TEXT-format escape layer applied by COPY collides with the JSON value's
  own escapes and the resulting stream looks valid on inspection but
  fails parsers at the first internal control character.
- **`kubectl exec -i` should only be used when stdin is actually
  consumed.** Adding `-i` to one-shot commands occasionally produces
  truncated stdout streams with `next reader: unexpected EOF`. Default
  to no `-i`; add it only for actual interactive or piped-stdin uses.

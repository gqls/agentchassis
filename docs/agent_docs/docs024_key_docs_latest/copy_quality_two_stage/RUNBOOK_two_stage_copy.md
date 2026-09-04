# RUNBOOK — copy_quality_two_stage

Created 2026-09-04, late: the lane ran three weeks without one and the first thing this session
needed — how the 08-31 model arms were actually called — was nowhere. Commands go here the moment
they were hard to get right, with their gotcha attached. Newest sections at the bottom.

## Offline model replay — one stored writer prompt, any vendor, key never leaves the pod

The subject prompt for every arm of `AUDIT_prompts/EXPERIMENT_2026-08-31_model_trials.md`:

```bash
S=<scratch dir>
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At \
  -c "SELECT prompt_rendered FROM llm_call_log WHERE id='79257fb4-fcfa-4ff6-9923-dc4e7fcd2b6a';" \
  > $S/benchmark_prompt.txt          # 38,002 chars; -At keeps newlines
```

Gotchas: `-At` output has psql's trailing newline (`rstrip` it once, not `strip` — the prompt ends
in a colon-less line and the body is what was sent). The row is in `llm_call_log`, which the owner
wants kept verbatim (MEMORY: it is the training corpus) — read it, never trim it.

**xAI (Grok).** The pods carry `XAI_API_KEY`/`GROK_API_KEY` from `personae-default-secrets`, have
BusyBox `wget` and no `curl`/`python3`/`jq`. Build the body locally, pipe it in, POST from the pod,
cat the reply back. OWNER RULE (08-23): never read a key into the session — `$XAI_API_KEY` is only
ever expanded inside the pod's shell.

```bash
python3 - "$S" <<'PY'
import json,sys; S=sys.argv[1]
p=open(f"{S}/benchmark_prompt.txt").read().rstrip("\n")
open(f"{S}/req.json","w").write(json.dumps({"model":"grok-4.6",
  "input":[{"role":"user","content":p}],"max_output_tokens":16000}))
PY
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | head -1)
kubectl -n ai-persona-system exec -i $POD -- sh -c '
  cat > /tmp/req.json
  wget -S -qO /tmp/resp.json -T 500 --header="Content-Type: application/json" \
       --header="Authorization: Bearer $XAI_API_KEY" --post-file=/tmp/req.json \
       https://api.x.ai/v1/responses 2>/tmp/hdr.txt
  echo "wget exit $?"; head -1 /tmp/hdr.txt; cat /tmp/resp.json' < $S/req.json > $S/resp.raw
```

> ~~**⚠ For Anthropic, do NOT use wget at all** — six of six BusyBox `wget --post-file` calls to
> `api.anthropic.com` returned 400 while xAI returned 200, `[INFERRED]` BusyBox's own
> `Content-Type` header colliding with the `--header` one.~~
> **REFUTED 2026-09-04 13:24Z, by the control, ~70 minutes after it was written.** The identical
> wget recipe returns **200** now. Those six 400s all fell inside **11:21–11:57Z**, when the
> Anthropic account ran out of credit and every LLM call on the fleet 400'd (the council was dead
> in the same window — six lanes' submissions died `complete_invalid`; the `dispatch_throughput`
> lane has it as an incident with a 120s detector live). **wget was never the cause**, and the
> tool difference was pure coincidence of timing: the Go poster's first call went out at ~12:00Z,
> three minutes after the outage ended. What made the wrong story so easy to believe is that
> BusyBox DISCARDS the 4xx body, so the outage's own error text — which names credit — was the one
> thing I could not see; the tool that hid the evidence got blamed for the failure it hid.
> **The check, and it is one command:** re-run the failing call after the suspected window, and
> before theorising about an HTTP client, send it a body small enough to fail fast
> (`{"model":...,"max_tokens":16,"messages":[{"role":"user","content":"Say OK."}]}`).

The poster is still the better tool for this job — it prints the response body on a non-200, which
is what turns "400" into a diagnosis:

```bash
# $S/antpost/main.go — ~50 lines: reads <request.json>, POSTs to anthropic|xai with the key from
# the POD's env (x-api-key + anthropic-version: 2023-06-01, or Bearer $XAI_API_KEY), writes the
# full body to <out.json> and prints status + latency; prints the body on non-200.
(cd $S/antpost && go mod init antpost 2>/dev/null; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o antpost .)
kubectl -n ai-persona-system cp $S/antpost/antpost ai-persona-system/<pod>:/tmp/antpost
kubectl -n ai-persona-system exec -i <pod> -- sh -c 'cat > /tmp/req.json; chmod +x /tmp/antpost;
  /tmp/antpost anthropic /tmp/req.json /tmp/resp.json; echo =====; cat /tmp/resp.json' < $S/req.json > $S/resp.raw
```

Source is reproduced in `AUDIT_prompts/EXPERIMENT_2026-08-31_model_trials.md` §"Arm 5" so it is
not lost with the scratch dir. (BusyBox `wget --post-file` works too — see the struck block above.) Anthropic request gotchas for `claude-sonnet-5`: `thinking` must be
`{"type":"adaptive"}` (add `"display":"summarized"` to SEE it) or `{"type":"disabled"}` —
`budget_tokens` 400s; effort goes in `output_config.effort`; **thinking counts against
`max_tokens`** — at effort `max` the model spent all 16,000 on thinking and returned NO text
(`stop_reason: max_tokens`), which scores as a vacuous zero on every battery. Assert
`stop_reason == end_turn` and a non-zero word count BEFORE reading any count.

Gotchas: BusyBox `wget` has no `--version` and DROPS the body of a 4xx — a refusal is `wget exit 1`
plus the status line in `hdr.txt` and an empty `resp.json`; the reason (credits, bad model id) is
unreadable from the pod, so read `orchestration_states`' captured error for the news arm's copy of
it (`bugs_open/418`) or the console. Reasoning models take **4–5 minutes** on this prompt — use a
Bash timeout of 600s and `-T 500`. Response envelope: `output[]` holds a `reasoning` item (with a
`summary`) and a `message` item whose `content[].type == "output_text"` is the text; `usage`
carries `reasoning_tokens` and `cost_in_usd_ticks` (1 tick = 1e-10 USD — verified against the list
prices). Model ids: `wget -qO- --header="Authorization: Bearer $XAI_API_KEY" https://api.x.ai/v1/models`
from the pod; `grok-4-1-fast` (the news arm's doc-comment recommendation) is NOT listed as of
2026-09-04. Run two calls on two different chassis pods in parallel — they are independent.

## Score a replay with the PRODUCTION scanner, without touching the tree

`count_negation_tells.py` (this dir) is the lexical battery the canaries used; it is NOT the
production gate's instrument, and the two disagree (sonnet's stored section: 10 vs 8). For a
number comparable with what the gate would do, run `datahelpers.ScanDefineByNegation` +
`ScanContrastNeighbours` over the flattened output. A `go run -overlay` makes a package that exists
only in scratch — nothing untracked lands in the shared tree for another session's `git add -A`:

```bash
# main.go in $S/goscore/ — reads each arg, JSON-parses it if it can, joins every string value,
# strips tags, prints len(ScanDefineByNegation) with shapes and len(ScanContrastNeighbours).
printf '{"Replace":{"%s/cmd/zz_scratch_negscore/main.go":"%s/goscore/main.go"}}' "$PWD" "$S" > $S/overlay.json
go run -overlay $S/overlay.json ./cmd/zz_scratch_negscore $S/sonnet_response.txt $S/out1.json $S/out2.json
```

**`cmd/zz_scratch_negscore/` DOES NOT EXIST and must not be created** — that is the point of the
overlay, and the pre-commit pattern check flags the name (correctly) as a proposed new capability.
The package is materialised only inside `go run`'s build, so nothing untracked is left in a tree
that other sessions run `git add -A` on. The real detector already exists twice over
(`datahelpers` for the library, `cmd/brief-negation-check` for the scheduled fleet check); this is
neither — it is a throwaway main that calls the library on FILES, which neither of those does.

Gotchas: score the BASELINE in the same run as the arms — a NEG figure quoted from a doc was made
by whatever the scanner was on that date (LANDMINE: two rates over one corpus). Store every arm's
output verbatim (`AUDIT_prompts/TRIAL_OUTPUTS_<date>_*.md`); the 08-31 Fable/Gemini outputs were
not stored and cannot be re-scored.

## Does the production writer reason? — the column CANNOT tell you

```sql
SELECT count(*) FILTER (WHERE thinking_tokens > 0) AS with_thinking, count(*) AS total
FROM llm_call_log WHERE agent_type='page-content-writer' AND created_at > now() - interval '7 days';
-- 0 | 6724 on 2026-09-04 — and it is NOT evidence of no reasoning.
```

Gotcha (WRONG_CALLS 2026-09-04): `thinking_tokens` is filled from parsed thinking-block TEXT, and
`claude-sonnet-5` returns those blocks empty by default (`display: "omitted"`), while running
adaptive thinking whenever `thinking` is omitted — which `anthropic.go` does. So the writer DOES
reason and the column says 0. The honest indicator is `output_tokens` against the visible text
(2,713 vs ~1,100 for the benchmark row). To see the reasoning, replay with
`"thinking":{"type":"adaptive","display":"summarized"}`; never send `budget_tokens` to a 4.7+
model — it 400s, and BusyBox wget will not show you the body that says so.

## The planner canary needle (unchanged from the handoff, here so it is findable)

```sql
SELECT agent_type, created_at,
       position('BUILD STANDARD (applies to every site' IN prompt_rendered) > 0 AS has_standard
FROM llm_call_log WHERE agent_type IN ('build-site-planner','content-gap-planner')
ORDER BY created_at DESC LIMIT 5;
```

Case-sensitive on purpose (`position`/`LIKE`, never `ILIKE`) — the sentence-case form belongs to
`domain-research-classifier`'s hard-coded copy and would count a non-consumer.

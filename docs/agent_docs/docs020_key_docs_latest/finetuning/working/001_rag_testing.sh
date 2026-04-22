missing test cases ...

---
Safer verifier, paste this:
echo "Lines in file: $(wc -l < /tmp/flywheel_d/test_cases.jsonl)"
echo "Lines that parse as JSON: $(jq -c . /tmp/flywheel_d/test_cases.jsonl 2>/dev/null | wc -l)"
echo ""
echo "Prompt/response size check on first 5 rows:"
jq -c '{id, prompt_len: (.prompt_rendered | length), response_len: (.response_text | length)}' /tmp/flywheel_d/test_cases.jsonl | head -5

jq -c . < file reads all lines itself and only emits ones that parse. If it emits 20, all 20 are valid.

Check tunnel
curl -s -m 3 -o /dev/null -w "tunnel: %{http_code}\n" http://localhost:21434/api/tags
# If 000: restart
# If 200: proceed

If that's 000, restart the tunnel:
pkill -f "port-forward" 2>/dev/null; sleep 2
kubectl -n ai-persona-system port-forward svc/ollama-adapter 21434:11434 > /tmp/pf.log 2>&1 &
sleep 5
curl -s -m 3 -o /dev/null -w "tunnel: %{http_code}\n" http://localhost:21434/api/tags

See if ollama responds to a single test generate
curl -s -m 60 -X POST http://localhost:21434/api/generate \
  -H 'Content-Type: application/json' \
  -d '{"model":"mistral-small3.1","prompt":"Reply with a single word: hello","stream":false,"options":{"num_predict":10}}' \
  | jq -r '.response // "(no response field)"'

The new safer runner
Rewritten to fail fast, read from JSONL (not TSV), and not silently continue past errors:
IN=/tmp/flywheel_d/test_cases.jsonl
OUT=/tmp/flywheel_d/results.jsonl
rm -f "$OUT"

TOTAL=$(jq -c . "$IN" 2>/dev/null | wc -l)
echo "Processing $TOTAL test cases..."

# Pre-flight — one quick sanity call to make sure Ollama is responsive
preflight=$(curl -s -m 30 -X POST http://localhost:21434/api/generate \
    -H 'Content-Type: application/json' \
    -d '{"model":"mistral-small3.1","prompt":"hello","stream":false,"options":{"num_predict":5}}' \
    | jq -r '.response // ""')
if [ -z "$preflight" ]; then
    echo "PREFLIGHT FAILED — Ollama not responding. Aborting."
    exit 1
fi
echo "Preflight ok, Ollama responsive"
echo ""

ROW_NUM=0
while IFS= read -r row_json; do
    ROW_NUM=$((ROW_NUM + 1))
    log_id=$(echo "$row_json" | jq -r '.id')
    prompt=$(echo "$row_json" | jq -r '.prompt_rendered')
    claude=$(echo "$row_json" | jq -r '.response_text')

    echo "[$ROW_NUM/$TOTAL] log_id=$log_id prompt_len=${#prompt} ..."

    payload=$(jq -nc --arg m "mistral-small3.1" --arg p "$prompt" \
        '{model:$m, prompt:$p, stream:false, options:{temperature:0.3, num_predict:600}}')

    start=$(date +%s)
    ollama_raw=$(curl -s -m 300 -X POST http://localhost:21434/api/generate \
        -H 'Content-Type: application/json' -d "$payload")
    elapsed=$(( $(date +%s) - start ))
    ollama_response=$(echo "$ollama_raw" | jq -r '.response // ""')

    if [ -z "$ollama_response" ]; then
        echo "  EMPTY — raw response was: $(echo "$ollama_raw" | head -c 200)"
        echo "  Aborting so we can diagnose."
        break
    fi

    echo "  ok ${elapsed}s, ${#ollama_response} chars"

    jq -nc \
        --arg log_id "$log_id" \
        --arg prompt "$prompt" \
        --arg claude "$claude" \
        --arg ollama "$ollama_response" \
        --argjson elapsed "$elapsed" \
        '{log_id:$log_id, prompt:$prompt, claude_response:$claude, ollama_response:$ollama, ollama_latency_s:$elapsed}' \
        >> "$OUT"
done < <(jq -c . "$IN")

echo ""
echo "Done. $(wc -l < "$OUT") results written"

Key differences from before:

Preflight check — if Ollama doesn't answer "hello", abort immediately
Reads from JSONL via jq -c . — safe for multi-line fields
Breaks the loop the first time Ollama returns empty — surfaces the problem instead of producing 20 bad results

Steps to run in order:

Run the safer verifier from "Problem 2" — confirm 20/20 valid JSON.
Run the tunnel check and restart if needed.
Run the single Ollama smoke test.
Only if all three pass, run the new runner.



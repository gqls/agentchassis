TEST
Running the test (this is separate from the engine)
The test checks the plumbing — the request → confirm → pay → deliver state
machine, the auth gates, idempotency, the capacity limit. It uses a fake
payment provider and a stubbed engine, so it makes no API calls, costs
nothing, and needs no keys:
bashcd idea-go
GOPROXY=off GOTOOLCHAIN=local go test ./...
# expect: ok  idea  (PASS)
Add -v to see each check:
bashGOPROXY=off GOTOOLCHAIN=local go test -v ./...
Key point that resolves the confusion: the test does not touch OpenAI or
Anthropic at all. It deliberately swaps the real engine for a stub so it can
test the money/flow logic fast and free. OpenAI only ever comes in when you run
the real engine (go run . internal …, or a real paid order through the
service). The two are independent.
GOPROXY=off GOTOOLCHAIN=local just tells Go "don't go to the internet" — the
code is stdlib-only so it builds offline. On a normal networked machine you can
drop those and plain go test ./... / go run . work too.


export OPENAI_API_KEY=sk-...
export OPENAI_CRITIQUE_MODEL=gpt-4o      # or whatever current model you prefer
go run . internal "agritec.uk" "UK small farmers" "curate scheme docs"
# watch for: [cut] cross-vendor: OpenAI (gpt-4o)

cd idea-go
export ANTHROPIC_API_KEY=...           # real — the engine will spend
export INTERNAL_API_KEY=$(openssl rand -hex 16)
export AUTO_DELIVER=false
export OPENAI_API_KEY=sk-...
export OPENAI_CRITIQUE_MODEL=gpt-4o      # or whatever current model you prefer
GOPROXY=off GOTOOLCHAIN=local go run .   # starts the service on :8080
# Or run the engine directly against one of our own domains (no billing, no server):
go run . internal "agritec.uk" "UK small farmers" "curate scheme docs"
# Internal HTTP endpoint:
curl -s localhost:8080/internal/run -H "X-Internal-Key: $INTERNAL_API_KEY" \
  -H 'content-type: application/json' \
  -d '{"domain":"agritec.uk","audience":"UK small farmers","assets":"curate scheme docs"}'
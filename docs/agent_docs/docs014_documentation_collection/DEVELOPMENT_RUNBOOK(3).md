# Development runbook — idea.uk + SFI26 tool

What we're building, in what order, with explicit acceptance criteria so we know
when each step is done. Designed to be checked off as we go. Each task says
*what's done* (the output), *how to test it*, and *what unblocks*.

References:
- Architecture: `idea_uk_architecture_and_deployment.md`
- Method (v3, with Risk column added 2026-05-28): `idea_uk_method_v0.md`
- Single-shot prompt (Risk-aware): `idea_method_prompt.md`
- Liability + T&Cs: `LIABILITY_AND_TERMS.md`
- Method test runs: `idea_uk_testrun_v0.md`, `idea_uk_testrun_v2.md`
- Plan: `PLAN_idea_uk.md`
- Debugging guide §0 item 23 (the Risk-column lesson): `016_debugging_guide_v2_30.md`
- Ongoing journal: `running_notes.md`

---

## Phase A — Pre-launch quality and liability hardening for idea.uk

Goal: bring the existing standalone Go service from "works in test" to "fit
to take real money from real people without exposing us."

### A1 — Engine upgrade: latest LLM features  **— DONE 2026-05-28**

**Output:** engine.go uses extended thinking on the cut step, the newer
`web_search_20260209` tool with code execution on verify, prompt caching on
the static system prompts (capability menu + system base), and Anthropic file
attachments where useful.

**Landed:**
- `callClaude` refactored to `callClaudeOpts` (options struct) so thinking +
  caching could be added without rewriting the existing call sites.
- Default models bumped to **Opus 4.8** for generate + verify (frontier);
  Sonnet 4.6 stays on cut + score for cross-model diversity. All four
  models overridable via env vars.
- **Cut step**: extended thinking enabled (budget 4000 tok) on the Anthropic
  branch — sharper critique was the whole point. OpenAI branch unchanged.
- **Verify step**: upgraded from `web_search_20250305` to
  **`web_search_20260209` + `code_execution_20250522`** — dynamic filtering
  via the code execution tool keeps only the relevant chunks of search
  results in context. Anthropic's published benchmarks: ~11% accuracy lift on
  BrowseComp/DeepsearchQA, ~24% fewer input tokens. Code execution is **free**
  when paired with web_search. Extended thinking enabled (budget 8000 tok)
  for multi-step inference over search hits. Max tokens raised to 16000.
- **Score step**: extended thinking enabled (budget 2000 tok) — supports
  careful Risk-column application.
- **Audience + generate steps**: thinking deliberately off (we want breadth at
  the brainstorm step, not deeper reasoning). Caching on.
- **Prompt caching**: system prompt wrapped as a cached content block on every
  step. Within a single run (5 LLM calls in <10 min), steps 2–5 land cache
  reads on the (identical) system. Cache-hit telemetry logged to stderr.
- File attachments deferred — not needed by idea.uk MVP; useful later for the
  SFI tool's PDF corpus reads.

**Acceptance:**
- ✅ `go build ./...` clean, `go vet ./...` clean.
- ✅ `go test ./...` — 19 original flow checks + 5 new audience-check tests, all pass.
- ✅ **End-to-end live run validated 2026-06-04** (agritec.uk). Confirmed: cut
  step "with extended thinking"; `[cache] claude-opus-4-8: created=13798
  read=77940` (caching working — ~78k tokens served from cache); verify ran on
  Opus 4.8 at xhigh with the web-search loop (took a few minutes); the report
  carried an "Operator risk: N/5" line per candidate with two candidates at
  Risk 2 auto-flagged "⚠ needs liability work" and their cheapest_test rewritten
  to require PII + T&Cs first. Three API bugs found and fixed during validation
  (see debugging guide v2_32 items 24–26).
- Note: the run reported "search quota exhausted" with 6 searches across 4
  candidates, leaving several premises "provisional." **Fixed** — web-search
  budget is now `WEB_SEARCH_MAX_USES` (default 12, was 6).

### A2 — Free audience-check taster endpoint  **— DONE 2026-05-28**

**Output:** new `/audience-check` endpoint on the service (public, no auth, no
billing). Accepts `business` + `audience` form fields. Returns the rendered
HTML of Step 1 output (carried audience + reasoning + 3 alternatives) in
~10 seconds. ~£0.02 per call.

**Landed:**
- New file `idea-go/audience_check.go` with the handler + a per-IP sliding-
  window rate limiter (3/hour, 20/day) + HTML rendering that uses the page's
  existing `.taster-result` CSS.
- `runAudience` extracted from `RunMethod` so step 1 is independently
  callable; injected onto App as a function field so tests can stub it.
- HTML output escapes all user-supplied input (XSS-safe — explicit test).
- Kill switch via `TASTER_ENABLED=false` env var (returns 503 with a polite
  message redirecting visitors to the request form).
- Rate-limit response includes a `Retry-After` header and a plain-English
  "have another go in N minute(s)" body.
- Tests: GET rejected; missing field rejected; happy-path content shape
  asserted; XSS escaping asserted; rate-limit cap-and-window asserted.

**Acceptance:**
- ✅ Route lives at `/audience-check`, only accepts POST.
- ✅ Output is HTML suitable for direct innerHTML insertion by the page widget.
- ✅ Rate-limited per IP (3/hour, 20/day).
- ✅ Cost per call bounded — same one Opus call as Step 1 of the full method.

### A3 — Streaming progress page after Stripe redirect ("real door")

**Output:** post-payment success page polls a new `/status/{order_id}`
endpoint and renders progress in real time. Engine writes status updates to
the store at each step. Report renders in-browser when done; copy also
emailed.

**Acceptance:**
- After completing a test Stripe payment, browser shows "preparing… generating
  candidates… cutting… verifying claim 1 of N… complete" with each step
  updating live.
- Report renders inline in the browser on completion.
- Customer also receives the report by email.
- If the engine fails mid-run, the page shows a clear "we hit a problem,
  you've been refunded" state.

**Unblocks:** decent launch UX (the 72h email model is honest but feels
flat compared to real-time).

### A4 — Programmatic refund endpoint

**Output:** new `/refund` endpoint on the service, operator-gated (same
X-Internal-Key as confirm/decline). Body: `{order_id, reason}`. Calls Stripe's
`POST /v1/refunds` with the payment_intent we already store; sets order
status to `refunded`; emails the customer to confirm.

**Acceptance:**
- Refunding a test payment via this endpoint refunds it in the Stripe test
  dashboard, customer gets confirmation email, order shows `refunded`.
- Unauthenticated calls return 401.
- Already-refunded orders return 409.

**Unblocks:** smooth ops on day 1; saves time vs the Stripe dashboard at
volume.

### A5 — Page rewrite: simpler language + integrated taster + disclaimers

**Output:** `idea_uk_fakedoor.html` rewritten with plain English copy, the
audience-check taster integrated as a working widget (calls `/audience-check`),
the T&Cs and refund policy clearly linked at checkout and at the footer, and a
short "what this is and isn't" plain-English honesty paragraph in the body.
Editorial design retained.

**Acceptance:**
- A non-technical reader can answer in 30 seconds: what is this, what do I
  get, what does it cost, what's free.
- Taster widget on the page works end-to-end (filled fields → audience-check
  result rendered → CTA to pay).
- T&Cs link visible at the page footer and at the order checkbox.
- Reading age (a Hemingway-app check, say grade 8 or below) for the hero +
  value-prop sections.

**Unblocks:** launch.

### A6 — Terms of service and refund policy pages

**Output:** `/terms` and `/refund-policy` pages on idea.uk. T&Cs based on the
draft in `LIABILITY_AND_TERMS.md` §3, *after solicitor review*. Refund policy
matches what the page promises ("14 days, no quibble").

**Acceptance:**
- Pages live, linked from footer + checkout.
- Solicitor's revisions incorporated.
- Checkout flow has a "I accept the terms" checkbox; orders without it are
  rejected.

**Unblocks:** legally cleaner launch. Also reduces real-world dispute risk.

### A7 — Insurance: PII in force

**Output:** confirmed PII policy covering "AI-assisted analysis services,
operator-reviewed," £100k+ cover, ~£200–500/year. Insurer told honestly what
we do. Policy documents on file.

**Acceptance:**
- Active policy, in date, with the cover summary on file.
- Insurer aware that scope will widen to SFI advice when that product
  launches (so they can quote on the extension).

**Unblocks:** launching to real customers responsibly. The £29 idea.uk
product is arguably borderline-OK without PII, but it's so cheap not to be
uninsured that there's no good reason to skip it.

### A8 — Stripe live mode

**Output:** Stripe account verified, live keys in service env, live webhook
endpoint pointed at production URL, signing secret rotated to live. One real
£0.30 test transaction (your own card → your own account) end-to-end to prove
live mode works.

**Acceptance:**
- Live mode test: real card, real payment, real webhook, real fulfilment,
  real refund via the new endpoint.
- AUTO_DELIVER stays false.

**Unblocks:** launch.

### Phase A acceptance summary

Phase A is done when all of A1–A8 are done. At that point idea.uk can take
real orders responsibly. Estimated effort: roughly 2–4 days of focused work
plus solicitor-review wait (~1 week elapsed for A6) and insurance turnaround
(~1–2 days for A7).

---

## Phase B — Deploy idea.uk

Goal: idea.uk live, taking orders, operator-reviewing each report manually,
producing real evidence of demand.

### B1 — Deploy the service

**Output:** container running on a small box (Hetzner / Fly / Railway / your
ai-persona-system K8s pod), behind HTTPS, with persistent storage for the
JSON store. Health endpoint reachable.

**Acceptance:**
- `https://idea.uk/health` returns OK from outside.
- Restart-safe: container restart preserves state.
- Stripe webhook endpoint reachable from Stripe's IPs.

### B2 — Deploy the page

**Output:** rewritten page (from A5) deployed through the normal git → GitHub
Actions → Backblaze pipeline. Page can call the service (same origin via path
proxy, or CORS configured).

**Acceptance:**
- Page loads cleanly at https://idea.uk.
- Taster widget works against live service.
- Order form posts to live service and creates a real order.

### B3 — First operator-reviewed orders

**Output:** the first 10 paid orders go through with operator review on
every one. Each review takes ~15–30 minutes (read the draft, check sources,
edit anything wrong, send). Refund anything we can't deliver well.

**Acceptance:**
- 10 orders delivered.
- Pattern data captured: which reports were good first time, which needed
  edits, which were refunded and why.
- At least one of: the method is sharpened based on real customer reactions;
  or the report template improved; or the prompt set adjusted.

### B4 — Throughput decision

**Output:** based on B3 data, decide:
- Stay operator-review (default) and keep going.
- Switch AUTO_DELIVER on (only if no errors in B3).
- Hire a part-time reviewer.
- Pause new orders and improve the engine.

**Acceptance:** decision documented in running notes.

### Phase B acceptance summary

Phase B is done when 10 real reports have been delivered and reviewed,
patterns documented, and B4 decision made.

---

## Phase C — SFI26 Diff Alerts (the first vertical tool)

**Tool swapped 2026-05-28.** Originally Phase C was the SFI26 single-farm
assessment (£49–99 one-off, "tell us about your farm and we'll recommend an
action stack"). That product was paused on liability grounds: the Risk column
(added to the method this same session — see `idea_uk_method_v0.md` and the
debugging guide §0 item 23) scored it as Risk 2 — a wrong number could cost a
farmer £5–50k of lost grant money. Genuine opportunity, but not the right
first vertical to test the operator capability on.

**SFI26 Diff Alerts** replaces it. Same vertical (UK farm advisors, the audience
the method now confidently surfaces), much lower Risk (≈4 — we summarise what
changed in Defra/RPA guidance and cite the source; the advisor still does their
own application thinking). Re-scored under the new rubric:

- Defensibility 4 (versioned corpus + meaningful-change detection is the asset)
- Willingness 4 (advisors bill £400–800/day; £30–100/month/seat clears easily)
- Buildability 4 (mostly assembling existing actions: scrape, store, summarise)
- Reuse 3 (same shape works for CSHT, CS, BPS-transition, agritech regulation)
- Durability 4 (rules keep changing; the freshness is the moat)
- **Risk 4** (we report what changed; advisor decides what to do about it)

Sum 19. **Test-now.** The single-farm assessment goes into the backlog for
later, once we have PII in force, named-advisor on retainer, and operational
credibility from running Diff Alerts.

### C1 — Method specification for the Diff Alerts tool

**Output:** `SFI_DIFF_METHOD.md` — same shape as `idea_uk_method_v0.md` but
domain-specific. Inputs: nothing per-customer at MVP (every subscriber gets the
same digest); later, an optional client portfolio for personalised alerts.
Steps: ingest latest scrape, diff against previous version, classify changes
by impact type (rate change, action removed, eligibility change, deadline
shift), summarise with citations to specific Defra/RPA pages, produce digest.
Output template: weekly email + dashboard view, with each change linked to the
primary source and dated.

**Acceptance:**
- A human can follow the method by hand against a real diff and produce the
  digest.
- Every change in the digest cites the specific page and version it came from.
- The digest distinguishes formatting/typo changes (suppress) from rule
  changes (surface).

### C2 — Corpus curation

**Output:** scraped + versioned set of Defra/RPA SFI26 pages and PDFs. Each
document stored with a fetch date and version hash. Re-scraped weekly. Stored
in a way the engine can retrieve and diff.

**Acceptance:**
- ≥30 source documents in the corpus.
- Re-scrape job runs weekly without intervention.
- Engine can retrieve a specific document and produce a diff against the prior
  version.

### C3 — Diff engine

**Output:** new endpoint or binary running the diff method. Reuses the engine
framework (LLM calls, structured outputs, citations) but the workflow is:
ingest → diff → classify → summarise. No web_search at MVP (the corpus is the
ground truth).

**Acceptance:**
- `go run . sfi-diff` produces a draft weekly digest covering all material
  changes since the last run.
- Every claim is cited to a specific document + version.
- Operator can review and edit the draft before sending.

### C4 — Operational shape

**Output:**
- Static landing page on agritec.uk (built through your normal pipeline)
  describing the product.
- Sign-up flow (email + Stripe subscription, £49/month per advisor seat to
  start, adjusted after first 10 subscribers).
- Operator-review the first 8 weekly digests; switch to auto-send only after
  no material errors.
- PII scope confirmed in force (lower threshold than the single-farm
  assessment would have needed, but still warranted).

**Acceptance:**
- Page live, sign-up works, payment recurring works, operator-review process
  documented.
- First digest produced and delivered without error.

### C5 — First 8 weekly digests

**Output:** weekly cadence for ≥2 months. Operator-reviewed every week.
Subscriber feedback collected (was this useful? did you act on anything?
anything missing?). Method/corpus sharpened from feedback.

**Acceptance:**
- 8 digests delivered.
- No material errors (claimed change misrepresented, missed change later
  noticed).
- ≥5 subscribers report having acted on at least one alert.

### Phase C acceptance summary

C is done when 8 weekly digests have been delivered with operator review and
no material errors. Decision then: layer the single-farm assessment on top
(now backed by domain credibility + PII + advisor relationships), or build out
the diff product further (per-subscriber portfolio overlays, cross-scheme
diff for BNG/CSHT/Woodland Carbon).

---

## Phase D — Chassis-native version (deferred)

Goal: express the idea-generation method as an internal chassis agent +
workflow reusing existing actions, for running the method on our own domains
on demand or on a schedule.

**Not started yet.** Per the architecture doc, this needs a schema pass first.

### D1 — Schema and action contract pass

**Output:** brief document mapping each method step onto specific existing
actions (`execute_llm_prompt`, `web_search`, `request_human_input`, etc.) with
the exact action contracts (inputs/outputs) we need to honour. Confirm what
new actions, if any, need writing.

**Acceptance:**
- Mapping is complete and consistent with the action registry.
- Any new actions identified are listed with their proposed contracts.

### D2 — Agent definition + workflow SQL

**Output:** new agent definition `idea-orchestrator` + a workflow that
sequences the steps using the actions from D1. Spawns sub-agents per
convention (every agent is an orchestrator). Children reply on the parent's
responses topic.

**Acceptance:**
- Workflow runs end-to-end via the chassis orchestrator.
- Produces the same shape of output as the standalone Go service.
- Logs are clean.

### D3 — Wire into site_specs feasibility model

**Output:** a domain's site spec can declare it needs the idea-generation
capability; while blocked, `feasibility-recheck` waits; once the agent is
deployed, the page builds and is linked to the central service.

**Acceptance:**
- A new domain in the build pipeline can include "ideation" as a feature in
  its site_plan, status flips from blocked to planned to built once the agent
  exists.

### Phase D priority

Low for now. Phase A → B → C is the demand-validation path; D is the
"productionise inside the platform" path that pays off only once the standalone
versions have proven demand.

---

## Right-now-this-week shortlist

Concrete tasks ordered for the next session:

1. ~~**A1** — engine upgrade to latest LLM features.~~ **DONE 2026-05-28**
2. ~~**A2** — audience-check taster endpoint.~~ **DONE 2026-05-28**
3. **A3** — streaming progress page after Stripe redirect (real-door UX).
4. **A4** — programmatic refund endpoint.
5. **A5** — page rewrite (already done earlier — verify the taster widget
   wires to the new live `/audience-check` cleanly when both sides deploy).
6. **A6** kickoff — send draft T&Cs from `LIABILITY_AND_TERMS.md` §3 to a
   solicitor.
7. **A7** kickoff — get a PII quote.

A6 and A7 run in the background while we build the rest. Phase B (deploy)
starts after A1–A8 are all complete.

### One open validation step before going further

Run the upgraded engine once against a real domain end-to-end to confirm
extended thinking + web_search v2 + caching all behave as expected in prod.
Suggested:

```
cd idea-go
GOPROXY=off GOTOOLCHAIN=local go run . internal "agritec.uk" "UK small farmers" "curate scheme docs"
```

Watch stderr for: `[cut] ... with extended thinking`, `[cache] ... read=N` on
steps 2–5, and the report should now include the Operator risk line per
candidate (Risk column landed earlier in the same session).

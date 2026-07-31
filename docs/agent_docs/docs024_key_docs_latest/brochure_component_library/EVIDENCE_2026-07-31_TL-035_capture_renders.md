# EVIDENCE — TL-035 `capture_renders`, and the council's two advisory objections answered

Council gate, submission `ab21beac-b5cd-43a8-a66f-c73ef33b6d49`, **APPROVED round 1**
("approved with 2 advisory objection(s) — none high-severity"). Two seats objected; both
objections were fair and both were answerable with a check rather than an argument. This
file is that check, written because the objections were specifically that my claims rested
on **unattached** evidence — so answering them in prose would have reproduced the fault.

---

## 1. editquality, medium — "mutation-proven" was a claim, not an artifact

> *"Rationale claims tests were 'mutation-proven' and lists specific mutants caught, but
> mutation-testing claims of this kind are exactly the landmine the fleet has been bitten
> by before ('A mutation that never happened is indistinguishable from a guard that
> works'). The plan gives no artifact."*

**Conceded, and the seat is right that the shape of the claim is the problem.** The mutants
did run. Here is the run, reproducibly, with the actual output.

Each mutant is a one-line edit to `internal/adapters/browserrunner/run_checks_action.go`,
followed by `go test ./internal/adapters/browserrunner/ -run 'TestRenders|TestFailingRunGoes|TestP3' -count=1`,
then restore. Reproduce with:

```bash
cp internal/adapters/browserrunner/run_checks_action.go /tmp/orig_rc.go
# M1 — ignore the opt-in, capture always (breaks default-off)
#   replace: if len(failed) == 0 && !req.CaptureRenders {   ->   if false {
# M2 — invert the routing (failures file as renders)
#   replace: if failing {   ->   if !failing {      (in the Execute call site)
# M3 — ignore the opt-in so renders never fire
#   replace: if len(failed) == 0 && !req.CaptureRenders {   ->   if len(failed) == 0 {
cp /tmp/orig_rc.go internal/adapters/browserrunner/run_checks_action.go   # restore
```

Observed output, verbatim:

```
--- MUTANT: always capture (default-off broken) ---
--- FAIL: TestP3NoScreenshotWhenAllPass (0.00s)
--- FAIL: TestRendersOffByDefault (0.00s)
FAIL	github.com/gqls/agentchassis/internal/adapters/browserrunner	0.008s

--- MUTANT: opt-in ignored (capability inert) ---
--- FAIL: TestRendersCapturedOnPassWhenRequested (0.00s)
FAIL	github.com/gqls/agentchassis/internal/adapters/browserrunner	0.008s

--- MUTANT M2 (routing inverted: failures -> Renders) ---
--- FAIL: TestP3ScreenshotCapturedOnFailure (0.00s)
--- FAIL: TestRendersCapturedOnPassWhenRequested (0.00s)
--- FAIL: TestFailingRunGoesToScreenshotsEvenWithRendersOn (0.00s)
FAIL	github.com/gqls/agentchassis/internal/adapters/browserrunner	0.008s

=== restored, full package ===
ok  	github.com/gqls/agentchassis/internal/adapters/browserrunner	6.013s
```

**Note which tests fire.** M1 and M2 are each caught by a **pre-existing** test
(`TestP3NoScreenshotWhenAllPass`, `TestP3ScreenshotCapturedOnFailure`) as well as by a new
one. That is the strongest available evidence that the change did not move the old
behaviour: the tests that guarded it before still guard it, and they still bite.

**And the failure that proves the method rather than the code:** my first attempt at M2 was
`if failing {` → `if false {`. It did not fail the tests — it failed to **compile**
(`failing` became unused), so it tested nothing. Recorded because it is precisely the
landmine the seat cited, caught in the act: *a mutant must compile before its result means
anything*, and "the suite went red" is not the same as "the guard fired".

## 2. prior_art_librarian, medium — a THIRD render/audit surface the rationale never mentioned

> *"The rationale's central absence claim — 'the platform's only screenshot path fires
> exclusively on failure' — is evidenced only against two Go files … The landmine list flags
> a THIRD render/audit surface this rationale never mentions: `scripts/render_audit.py`."*

**A real gap in my evidence, and checking it STRENGTHENS the claim rather than qualifying
it.** `scripts/render_audit.py` exists, renders pages, and:

```
$ grep -n -i "screenshot\|--screenshot\|dump-dom" scripts/render_audit.py
199:         "--dump-dom", "file://" + path],
```

**One hit, and it is `--dump-dom`, not `--screenshot`.** That surface renders a page and
never photographs it at all — so it is a third place that looks at a page's *structure*
while producing no image for anyone to look at. It also renders a **local `file://` copy**,
which the lane runbook already records as measuring the wrong thing (a query string never
reaches `window.location`, and cross-origin scripts do not execute).

So the corrected absence claim, stated with its full surface:

| surface | renders? | photographs? | when |
|---|---|---|---|
| `internal/adapters/browserrunner/run_checks_action.go` | yes, live | yes | **only on failure** (was `if len(failing) == 0 { return }`) |
| `internal/adapters/browserrunner/render_audit_action.go` | yes | **no** — `grep Screenshot` returns two comment lines only | n/a |
| `scripts/render_audit.py` | yes, but a **local copy** | **no** — `--dump-dom` only | n/a |

Three surfaces, one camera, and it was wired to fire only after something had already
failed. The objection improves the entry; TL-035's sources now name all three.

## 3. prior_art_librarian, low — the consumer grep was unattached

> *"'CONSUMERS NAMED AND TOLD' asserts the three tool_acceptance_actions.go call sites are
> 'the only consumers … (verified by grep)' — an absence claim resting on an unattached
> grep."*

Attached. Command and full output:

```
$ grep -rn "Screenshots\b\|ScreenshotRef" --include="*.go" platform/ internal/ \
    | grep -v "_test.go" | grep -v "browserrunner/screenshots.go"
internal/adapters/browserrunner/run_checks_action.go:121:// ScreenshotRef points at one captured full-page screenshot — the P3 evidence
internal/adapters/browserrunner/run_checks_action.go:125:type ScreenshotRef struct {
internal/adapters/browserrunner/run_checks_action.go:138:	// Screenshots is present only when a (url, profile) run had failures AND
internal/adapters/browserrunner/run_checks_action.go:140:	Screenshots []ScreenshotRef `json:"screenshots,omitempty"`
internal/adapters/browserrunner/run_checks_action.go:297:				out.Screenshots = append(out.Screenshots, ref)
internal/adapters/browserrunner/run_checks_action.go:327:	req RunChecksRequest, results []CheckResult, profile, url string, urlIdx int) (ScreenshotRef, bool)
internal/adapters/browserrunner/run_checks_action.go:330,339,346,353,358: (inside captureFailureEvidence)
internal/adapters/webscrape/truncation.go:231:	// Screenshots go entirely; the URI is the useful form and the base64 is the

$ grep -rn '"screenshots"' --include="*.go" platform/ internal/ | grep -v _test
platform/orchestration/actions/tool_acceptance_actions.go:650:				spec["screenshots"] = shots
platform/orchestration/actions/tool_acceptance_actions.go:704:				spec["screenshots"] = shots // P3 evidence: what the page looked like when it failed
platform/orchestration/actions/tool_acceptance_actions.go:970:			spec["screenshots"] = evidence
```

Every hit outside the adapter itself is one of the three named call sites, plus one
**comment** in `internal/adapters/webscrape/truncation.go:231` which is about dropping
screenshot base64 from scrape payloads and touches neither type.

**The limitation of this evidence, stated rather than glossed:** a Go grep cannot see a
consumer that reads `data.screenshots` out of the JSON response in a workflow step, a
prompt template, or an SQL expression. I checked those too and found none, but the honest
form of the claim is *"no Go consumer, and no workflow/seed reference I could find"*, not
*"no consumer exists"*. This does not affect the change's safety either way: `Renders` is a
**new** key, so a consumer that reads `screenshots` cannot be affected by a field it has
never seen — which is the property that made the two-list design worth choosing.

## 4. editquality, low — the register entry changes no runtime behaviour

Accepted as characterised: TL-035 is a compliance artifact required by the platform-seams
convention (register the seam in the same commit that ships it), not a functional edit. It
was listed as an edit because the convention makes it part of the change rather than
optional documentation. No action.

---

**Verdict trailer:** the commit that shipped this (`6c1531f61`) carries
`Council-Submitted: ab21beac-b5cd-43a8-a66f-c73ef33b6d49`, written before the verdict
landed. It is now APPROVED, and per CLAUDE.md the `098` report resolves the correlation at
report time and credits the commit automatically — no amend, which forward-only forbids
anyway. **`Council-Reviewed:` was deliberately not written on that commit**, because at the
time of writing I had not read an approved verdict, and asserting one I had not read is the
coverage report's dishonesty surface.

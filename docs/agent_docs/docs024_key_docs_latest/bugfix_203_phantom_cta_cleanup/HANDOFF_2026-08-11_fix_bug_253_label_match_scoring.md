# HANDOFF — FIX `bugs_open/253`: `BestLabelMatch` scoring — continue here in a fresh session

**Written 2026-08-11.** Single-purpose handoff: fix bug 253. Everything you need is here or
named here; you should not have to re-derive any measurement below.

Sibling docs, read in this order if you want the arc: `bugs_open/253` (the case),
`bugs_open/248` (the other blocker, separate fix), `NOTES_phantom_cta_cleanup.md`
2026-08-10/11 entries (the trail), `HANDOFF_2026-08-10_continue_here.md` (the lane's other
open decisions — **do not action those; this handoff supersedes only the 253 thread**).

---

## 1. The job in one line

`datahelpers.BestLabelMatch` scores a candidate page by **how many of the LABEL's tokens it
contains**, and never by how much of the **CANDIDATE** that represents — so a page with a
long descriptive `nav_label` ties with (or beats) the page a label actually names, and an
alphabetical tie-break then picks arbitrarily. Fix the scoring; recalibrate; council-review;
ship.

## 2. Start by reproducing it — 2 minutes, no cluster needed

Drop this in `platform/orchestration/datahelpers/` as `zz_repro_test.go`, run, then **delete
it** (do not commit a test that asserts wrong behaviour — that is how a bug becomes a spec):

```go
package datahelpers

import "testing"

func TestREPRO(t *testing.T) {
	mk := func(name, title, nav, url string, interactive bool) LabelMatchCandidate {
		c, ok := NewLabelMatchCandidate(name, name, title, url, interactive, name, title, nav)
		if !ok { t.Fatalf("no tokens for %q", name) }
		return c
	}
	pages := []LabelMatchCandidate{
		mk("gripper-payload-calculator",
			"Gripper Payload Calculator — Overview | Robot-Hands.com",
			"Gripper Payload Calculator — Calculate Required Grip Force with Safety Factor | Robot-Hands.com",
			"/gripper-payload-calculator.html", false),
		mk("tool-gripper-payload-calculator",
			"Gripper Payload Calculator | Robot-Hands.com",
			"Gripper Payload Calculator — Validate Capacity with Safety Factor | Robot-Hands.com",
			"/tools/gripper-payload-calculator/index.html", true),
		mk("tool-gripper-safety-factor-calculator",
			"Gripper Safety Factor Calculator | Tools", "",
			"/tools/gripper-safety-factor-calculator/index.html", true),
	}
	best, ok := BestLabelMatch("Gripper Safety Factor Calculator", pages)
	t.Logf("-> %q %s ok=%v", best.Name, best.URL, ok)
	if best.URL != "/tools/gripper-safety-factor-calculator/index.html" {
		t.Errorf("WRONG: %s", best.URL)
	}
}
```

Those rows are **verbatim from live `pages`** for robot-hands.com (2026-08-11). Current
output:

```
label tokens: [gripper safety factor calculator]
  gripper-payload-calculator              interactive=false overlap=4
  tool-gripper-payload-calculator         interactive=true  overlap=4
  tool-gripper-safety-factor-calculator   interactive=true  overlap=4
-> "tool-gripper-payload-calculator" /tools/gripper-payload-calculator/index.html
```

All three tie at 4. `c.Name < bestPtr.Name` gives it to `payload` ("p" < "s"). The payload
page earns its 4 from a nav_label reading *"…Validate Capacity with **Safety Factor**…"*.

## 3. The one thing that will catch you out

**There are TWO different candidate pools, and your change hits both — differently.**

| caller | pool | file |
|---|---|---|
| **detector** `check_misdirected_cta` | **every** non-deleted/archived page with a url (index/home excluded) | `discovery_checks/check_misdirected_cta.go` → `loadCTAMatchIndex` |
| **resolver** `setCTAField` / `applyCTARecompute` | **only** `page_type='tool'\|'game'` + `page_type='section-index'` hubs | `resolve_internal_links_action.go` → `candidatesFromHubs` |

The 253 repro is a **detector-pool** case (`gripper-payload-calculator`, a `content` page, is
only a candidate there). A fix validated solely against the detector pool is half-tested.
Both pools feed `BestLabelMatch`; there are exactly four non-test files touching it:

```
platform/orchestration/datahelpers/label_match.go          (the function)
platform/orchestration/actions/resolve_internal_links_action.go   (write path)
platform/orchestration/actions/rerender_page_sections_action.go   (repair path)
platform/orchestration/actions/discovery_checks/check_misdirected_cta.go (detector)
```

## 4. Fix candidates — analysis already done, pick with your eyes open

`bugs_open/253` lists four. Worked numbers for the repro case (tokens = union of name, title,
nav_label; `recall` = overlap÷|label|, `precision` = overlap÷|candidate|):

| candidate | tokens | overlap | recall | precision | F1 |
|---|---|---|---|---|---|
| `tool-gripper-safety-factor-calculator` | ~6 | 4 | 1.00 | 0.67 | **0.80** |
| `tool-gripper-payload-calculator` | ~11 | 4 | 1.00 | 0.36 | 0.53 |

**Recommendation: candidate 2 (weight the token SOURCES) as the primary sort, not candidate 1
alone.** Reasoning:

- It is semantically the real distinction: `name`/`title` are the page's **identity**;
  `nav_label` is **description**. "Does this text name that page?" should be answered on
  identity. Under source-weighting the case is not even close — safety-factor matches 4 on
  name+title, payload matches 2 on name+title (`gripper`, `calculator`) and only reaches 4
  via nav_label.
- **Pure precision/F1 (candidate 1) has its own failure mode**: it advantages candidates with
  very FEW tokens. A page named `tools` (1 token) scores precision 1.0 against the label
  "Tools Guide" and would beat a richly-named page that is the better answer. Do not ship
  candidate 1 on its own without testing that shape — construct it deliberately, it will not
  appear in calibration by luck.
- A clean lexicographic ordering that subsumes both: **name/title overlap → total overlap →
  interactive → shortest candidate token set → name.** Note the fourth key replaces the
  alphabet with something meaningful and, on its own, fixes this exact transcript
  (candidate 4 in the bug file) — worth keeping even beside a real fix.

**Cost of candidate 2**: `LabelMatchCandidate.tokens` is a single unexported
`map[string]bool` built by unioning all `tokenSources`. You need per-source sets (or at
minimum an `identityTokens` set alongside `tokens`). `NewLabelMatchCandidate`'s variadic
`tokenSources ...string` signature has **three call sites with different arities** — the
detector passes `(name, title, navLabel)`, `candidatesFromHubs` passes `(h.Name, h.Title)`.
Changing the signature is the honest move; keep the unexported-tokens invariant (the comment
at the struct says why: a caller must not be able to construct a candidate whose tokens
disagree with its own name/title).

**Explicitly rejected — do not do this:** adding `safety`, `factor`, `calculator` etc. to
`LabelStopwords`. They are genuinely distinctive here, and the standing landmine (*narrowing a
detector past an invented false positive can make the rule inert*) applies exactly. The
2026-08-08 interrogative fix was legitimate because `what`/`how` are grammatical; these are not.

**Worth considering, not yet costed:** IDF-style down-weighting over the candidate set —
`calculator` appears in 6 of 7 robot-hands tool names and carries almost no information,
while `safety`/`factor` are rare. Principled, but it does NOT fix this case alone (payload
holds `safety`+`factor` too, via nav_label), so it is a complement to source-weighting, not a
substitute.

## 5. Recalibrate before shipping — this is not optional and the method matters

A scoring change moves **every** match on the fleet. The lane already learned the wrong way
to measure this (NOTES 2026-08-08): an early pass calibrated against *all active pages*
rather than the pool the shipped code actually uses and produced a materially larger,
unfaithful number.

**Method that works** (from that entry, and it is the one to repeat):

- A throwaway `cmd/ctacalibrate` importing the **real** `datahelpers` package — not a SQL
  re-implementation of the scoring. **It does not exist in the tree and is not meant to** —
  `pattern-check.py`'s `new-capability-surface` rule fires on this path, correctly, and the
  answer is: it existed on 2026-08-08, produced
  `CALIBRATION_2026-08-08_label_match_report.txt`, and was **deliberately deleted before
  commit** per its own header comment. It is a measuring instrument, not a service — it
  imports the package under test and reads live data, so a committed copy would rot against
  the very code it exists to measure. Recreate it, use it, delete it again; keep the report.
- `kubectl port-forward` to `postgres-clients`, run against live data.
- **Calibrate against the SAME candidate pool the caller uses** — and now there are two
  pools (§3), so report them separately.
- Delete the harness before commit (the previous one carried a header comment saying so;
  `CALIBRATION_2026-08-08_label_match_report.txt` in this directory is the retained
  evidence, 898 lines — read it for the output shape).

**Baseline figures from 2026-08-08, resolver pool** (what the last calibration measured, so
you can compare like with like): 1,251 labelled CTA fields examined; 634 matched; 315 newly
resolved; **162 would override an existing different valid URL** — that last number is the
risk-bearing one and is the one to watch move.

**Fleet scale today (measured 2026-08-11):** 607 active candidate pages; **37** have a
`nav_label` longer than 40 chars and **22** have a `nav_label` longer than their own `title`
— that second figure is roughly the population with the 253 shape. Small enough that you
should inspect the affected matches individually rather than trusting an aggregate.

## 6. Tests

Must keep passing (they encode the two already-shipped fixes in this same function — do not
regress them to make yours pass):

- `TestBestLabelMatch`, `TestLabelTokens`, `TestNewLabelMatchCandidateRejectsAllGenericSources`
- `TestBestLabelMatchOverlapBeatsCategory` — overlap beats category (shipped `3bc0486d7`)
- `TestBestLabelMatchInteractiveTiesBreakToInteractive` — interactive still wins a true tie
- In `actions`: `TestApplyCTARecompute*`, `TestSetCTAField*`, `TestChooseCTATargets*`
- In `discovery_checks`: `TestCTAClassifyAnchor` (its `interactive_page_preferred_as_suggested_target`
  subtest has candidates of **equal** overlap, so it does not conflict with a scoring change —
  verified 2026-08-10)

Add: the 253 case as a proper regression test (assert the CORRECT answer), plus the
short-candidate shape from §4 if you ship anything precision-based. **Mutation-prove the new
test** — `git stash` the fix and confirm the test fails without it. A test that passes both
ways is not evidence.

## 7. Ship it — the process bits, including one I got wrong

- **Council gate: yes.** Shared mechanism, three live consumers, changes what every CTA
  resolution is allowed to depend on. Submit via
  `docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <file.json>`.
  Budget ~30 min. Name the three consumers in the submission (owner ruling 2026-07-29 §3:
  other consumers must be **told**, not merely measured).
- **⚠ PUT THE TRAILER IN THE COMMIT.** I committed `3bc0486d7` before building the
  submission, so it carries no trailer, and forward-only means it can never be credited —
  it will read UNREVIEWED in `098` for ever. Logged in `WRONG_CALLS.md` 2026-08-10. Submit
  first (or same breath), then commit with `Council-Submitted: <SUBMISSION_CORR>`. Never write
  `Council-Reviewed:` on a verdict you have not read.
- **Register/landmines**: if the scoring becomes a named, reusable concept, it wants a
  concept-register entry. At minimum, update `bugs_open/253` in place with the outcome.
- **Deploy**: Go changes are inert until built and rolled. Releases here are **whole-fleet and
  owner-run** (`make release redeploy-agents ENVIRONMENT=production REGION=uk001`) — do not
  build+deploy this service alone at its own tag; that fragments the fleet. Ask the owner.
- **Verify at the artefact.** Binaries now carry build provenance (BLD-019, since
  `v1.0.1283`), so *"did my fix ship?"* is a query:
  `kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'`
  then `git merge-base --is-ancestor <your-commit> <the stamp>`. **Do not use `strings`** —
  it is absent from these images and CLAUDE.md now forbids the old recipe. Per SERVICE, not
  per fleet (`bugs_open/249`: one tag shipped three revisions).
  **Best verification is still behavioural**: §8 below.

## 8. How you will know it worked

1. Repro (§2) returns the safety-factor URL.
2. Full test suites green, new test mutation-proven.
3. Calibration diff reviewed — specifically the count of "would override an existing valid
   URL", against the 162 baseline.
4. **Live control, once rolled**: re-run detection on robot-hands.com and confirm
   `how-to-specify-a-gripper` drops from **3 findings to 0**. Those three anchors are
   known-correct (text names the safety-factor calculator, href points at it), so they are
   the cleanest available signal. Dispatch it directly — the scheduled rotation cannot be
   targeted (its `pre_query` picks one site per 7-day cycle):
   ```bash
   kubectl -n kafka run -i --rm kcat-disc-$(date +%s) --image=edenhill/kcat:1.7.1 \
     --restart=Never -- kcat -P -c 1 \
     -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
     -t system.agent.generic.requests \
     -H correlation_id=$CORR -H request_id=$REQ -H message_id=$MSG \
     -H orchestration_id=$ORCH -H orchestration_name=verify_253 -H step_name=start \
     -H client_id=demo_client -H message_type=request -H action=orchestrate \
     -H from_agent_type=user -H from_agent_id=cli \
     -H responses_topic=system.agent.generic.responses <<JSON
   {"action":"orchestrate","config":{"agent_type":"completeness-discovery-agent"},"input_data":{"site_id":"00ff3af5-dad8-4770-9f70-3edc267a3c92","domain":"robot-hands.com"}}
   JSON
   ```
   Read the findings out of the orchestration, **not** out of `site_work_items` — see the
   trap in §9.

## 9. Traps already paid for — do not rediscover these

- **Detection files NOTHING while the old rows are open.** Fresh findings dedupe against the
  192 existing `detected` rows via `ON CONFLICT DO NOTHING` (`insertPageRerenderItem`). So
  after a verification run, `site_work_items` will show zero new rows and that is NOT a
  failure. Read the result from
  `orchestration_states.collected_data->'discovery_result'->'findings'` instead. This cost me
  a confused detour on 2026-08-11.
- **`misdirected_cta` items are created `detected`, and only `triaged`/`approved` dispatch.**
  The bridge is `TriageDetectedItemsAction`, which lives in the **improvement-loop** — and
  `improvement-sweep` plus all three `site-discovery-rotation-*` tasks are currently
  `enabled=f`. So nothing self-heals; equally, nothing runs away with your change.
- **Never run `TriageDetectedItemsAction` over a site to "let it heal".** It promotes **every**
  `detected` row for that site with no type filter (its own header says so) — and with
  `bugs_open/248` unfixed that would overwrite 24 working contact CTAs fleet-wide.
- **`bugs_open/248` is a SEPARATE fix and still open.** 253 alone does not make the repair
  path safe. Both are prerequisites for draining the `misdirected_cta` queue.
- A `090` diagnosis run on a symbol in a file over ~60KB returns bundles and no verdict —
  `resolve_internal_links_action.go` is large. Prefer the local repro; it is decisive.

## 10. State of the world when this was written

- Both earlier fixes are **approved, live and proven**: `bd6e3320c` (label-aware resolution)
  and `3bc0486d7` (overlap before category). Pods on `v1.0.1286`, started 2026-08-11 12:03Z.
- `bugs_open/248` — open, not started. `bugs_open/253` — open, this handoff.
- The lane's page-repair rollout is **halted by design** pending both.
- Nothing is in flight: no dispatched work, no pending council submission for this lane.
- `HANDOFF_2026-08-10_continue_here.md` §"Open decisions" still holds D1 (four
  leopardess "Get Started" heroes) and D3 (two scoped-out follow-ons); its D2 is superseded.

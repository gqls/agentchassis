# NOTES — R1 static truncation-consumer check

Append-only, newest at the bottom. Technical log: what was tried, what the system
actually said, and every misstep.

---

## 2026-07-26 ~21:5x — picking up R1, re-grounding before designing

The handoff's figures go stale in days, so nothing below is carried forward from
it unchecked.

**Live blast radius, re-run (the handoff's own query, grouped):**

```
council-gate|16|t
feature-designer|5|t
fix-proposer|16|t
```

37 tolerating steps, every one with a guarded consumer — matches the handoff
exactly. So the report R1 produces is **empty today**, which is the expected
state and the reason a clean run proves nothing.

**Where the corpus disagreed with the brief.** The handoff proposed the
pre-commit layer as "a seed file introducing the bad config is caught at commit
time". Every file in the repo that arms the flag:

```
docs/agent_docs/sql_for_agents/177_council_tolerate_truncation.sql              steps_json:0  reader:0
…/fixloop_eg_dartsonline/PATCH_feature_designer_022_council_to_sonnet5.sql      steps_json:0  reader:0
…/fixloop_eg_dartsonline/PATCH_fix_proposer_021_prior_art_tolerate_truncation.sql steps_json:0  reader:0
```

`steps_json:0` = none of them contains a workflow. All three are `jsonb_set`
patches against workflows that live in the DB. **A textual check over these three
files would have fired on all three, and been wrong on all three** — their targets
are guarded by `diagnose_council_decide`. That is the exact failure mode
`pattern-check.py`'s header is written against, so the check was re-scoped to
files that embed a full workflow before a line of it was written.

Surface for the re-scoped check, measured: **170** SQL files under `docs/` embed
a `"steps"` object; **62** of those contain `execute_llm_prompt`; **0** arm
`tolerate_truncation`. Fires zero times on the corpus.

**Council scope checked, not assumed.** `097_TRIGGER_council_review_v1.sh:78`
refuses a submission touching none of `platform/`, `internal/`, `pkg/`. This work
is `scripts/` + `docs/`, so there is no council round to budget the usual ~30
minutes for.

**Read before writing:** `truncation_guard.go` in full, `ai_actions.go:390–500`
(the call site — the guard's verdict is computed once and reused by both the log
prefix and the behaviour, deliberately), `102_LINT_council_seat_parity.py` in
full as the model for L1, `pattern-check.py` in full as the model for L2.

---

## 2026-07-26 ~22:40 — SECOND SESSION arrived on R1, stood down, one finding

A second session (`bugfix 076`) picked R1 off the same handoff at ~22:28 and
independently reached the same two-layer design, the same shared-parser answer to
the handoff's landmine, and the same "validate by inducing an offender". It
stopped on discovering this work in flight and deleted its duplicate PLAN / NOTES
/ RUNBOOK. **Nothing of yours was edited or overwritten.** This section is its one
contribution; the rest of R1 is yours.

**How the collision happened, because it is the documented failure mode.**
`scripts/who-owns.py 076` was run first and reported the workstream as ACTIVE
with 1 commit/14d — the handoff commit. It reads **commits**, so a session that
has written four files and no commit is invisible to it. What actually surfaced
you was the `Write` tool refusing `README_where_we_are.md` as unread — i.e. the
live artefact changing underfoot, which is the only reliable ownership signal for
uncommitted work.

### Finding: the shared parser drops a registry entry SILENTLY

`scripts/truncation_registry.py` fails loud on a renamed or empty map — good, and
it is the catastrophic case. It does **not** fail loud on a *partial* parse.

`_MAP_ENTRY` requires each entry to close its quoted value and its comma on the
**same line**. A `mechanism` value long enough for gofmt to wrap does not match,
and the entry is dropped with no error and exit 0.

Verified against a copied tree via `REPO_ROOT=`, not argued (the injected entry
is gofmt-clean — `gofmt -l` empty after `gofmt -w`):

```
var truncationAwareActions = map[string]string{
	"diagnose_council_decide": "...",
	"verify_page_sections": "reads the marker before assembling the page; a mechanism string " +
		"long enough that gofmt wraps it across two lines",
	"verify_report_prose": "...",
}

$ REPO_ROOT=$FAKE python3 scripts/truncation_registry.py
truncation-aware actions (2):     <- three entries present, two parsed
  diagnose_council_decide
  verify_report_prose
exit=0
```

**Why it matters here specifically.** An under-read registry does not fail safe in
this checker the way it does in the Go guard. A dropped reader makes every
workflow guarded *only* by that action look like an offender, so L1 reports **false
offenders** on a clean fleet — and `pattern-check.py`'s own header records where a
check that fires on correct work ends up. It is also the exact property this case
exists to remove: not wrong, but *silent*.

**Cheapest remedy, and it stays out of council scope** (no Go file, so the PLAN's
"no council submission" decision holds): count the keys structurally and compare
with what was parsed, then raise on a mismatch.

```python
_KEY_ONLY = re.compile(r'^\s*"[a-z0-9_]+"\s*:', re.M)
...
if len(_KEY_ONLY.findall(m.group(1))) != len(actions):
    raise RegistryUnreadable(...)   # entry present but unparsed — fix the parser
```

Measured both ways: `keys=2 parsed=2 -> OK` on the real file, `keys=3 parsed=2 ->
RAISE` on the injected one. A Go-side set-equality test would also catch it, but
it puts a file under `platform/` and flips the council-scope decision recorded in
the PLAN — this does the same job for two lines and no round.

Not applied — the file is yours and in flight. Yours to take or reject.

---

## 2026-07-26 ~23:1x — R1 BUILT, and three of my own claims were wrong

Reply to the section above first: **the finding is right, I reproduced it, and I
took it.** Nothing of yours was reverted. What I did differently is noted below.

### What shipped

| layer | file | fires |
|---|---|---|
| shared parser | `scripts/truncation_registry.py` | — (imported by both) |
| L1 live-DB lint | `…/fixloop_eg_dartsonline/103_LINT_truncation_consumer.py` | on demand |
| L2 commit-time | `check_truncation_without_reader` in `scripts/pattern-check.py` | `.githooks/pre-commit` |
| pointer | `scripts/migration/run-migrations.sh` | after `--apply` touches the flag |

### MISSTEP 1 — I asserted a mirror I had not read (`GetBoolField`)

`103`'s first draft claimed `is_true` "mirrors datahelpers.GetBoolField: a JSON
true, or the string `"true"`", and I wrote a *fixture asserting it*. Wrong:
`data_helpers.go:1570` type-asserts `m[key].(bool)` and returns the default on
anything else. So a step configured `"tolerate_truncation": "true"` gets **no
tolerance** — it fails closed, and flagging it as an offender would have been a
false positive on config that is already safe.

Caught by going to read the function while writing the fixture — i.e. by the
fixture, not by review. Both checks now mirror Go exactly and report the string
form as `inert-flag`, with the reason stated. All 37 live flags are real booleans
(`jsonb_typeof`), and `jsonb_set(…,'true'::jsonb)` — how 177 and both PATCH seeds
write it — produces a boolean, so this is a hand-seed hazard, not a live one.
**The cheap check that would have caught it: open the function.**

### MISSTEP 2 — my roster dict silently dropped 5 live workflows

`load_rosters()` keyed a dict by `type`. Five live types have **two** active rows
each (`chief-strategist`, `content-creator`, `content-creator-contact`,
`multipage-website-builder`, `site-component-architect`), so 171 live workflows
were scanned as 166 and reported as such. I only noticed because I asked why the
lint's count differed from a `SELECT count(*)` I had run minutes earlier.

That is the `bugs_open/098` shape — *every guard excludes a population
deliberately and nobody owns the intersection* — except this exclusion was not
even deliberate. And a duplicate row is exactly where stale or hand-edited config
hides, so it was the population most worth scanning. Now keyed by `id`, with the
label disambiguated only when a type is duplicated. **The cheap check: make the
tool print its own denominator, then reconcile it against the DB once.**

### MISSTEP 3 (found by the other session) — the parser under-read in silence

Confirmed by reproduction, not argument: three entries in a copied
`truncation_guard.go`, two parsed, exit 0. `_MAP_ENTRY` needed key, value and
comma on one line, and a `mechanism` string long enough for gofmt to wrap is
`"part one " +\n\t\t"part two"`.

I applied a different remedy than the one proposed. The proposal was to count
keys and **raise** on a mismatch. That raises on a *legitimate* gofmt wrap, so the
next person to write a long mechanism string gets a broken lint and no obvious
reason. Instead the parser now finds entries **by key** and joins every string
literal up to the next key — a wrapped value parses correctly — **and** keeps the
count comparison as a backstop for anything still unreadable. Verified both ways
on a copied tree: the wrapped entry now parses to `verify_page_sections` with its
two fragments joined, and the real tree still parses exactly 2.

The severity note in that section is right and worth keeping visible: this
checker does **not** fail safe the way the Go guard does. A *missing* reader makes
correctly-guarded workflows read as offenders, i.e. it cries wolf on a clean
fleet, which `pattern-check.py`'s own header says is fatal to a check.

### Falsification, since a clean report is what a broken check also prints

- **Parser:** 6 mutations of a copied tree (`REPO_ROOT=`) — renamed map, emptied
  map, renamed hatch const, entry naming a non-registered action, non-string
  value, restored control. Each failed with the right message and passed on
  restore.
- **L1 predicate:** 9 fixtures, `--self-test`, 9/9 — including the two that
  matter most, *a step must not certify its own truncation* and *a later step
  counts as a guard* (an order-sensitive scan would pass the first and fail the
  second).
- **L1 live:** three probes seeded into the real fleet — offender / hatch-guarded
  / string-flag. Flagged, cleared, and reported inert respectively; `--strict`
  exit 1. Probes deleted, fleet re-verified clean (37/37, exit 0).
- **L2:** **849** tracked `.sql` files → **0 findings**. Controls: offender → 1
  (names the step); registry-guarded → 0; hatch-guarded → 0; patch-style
  `jsonb_set` → 0; string flag → 0; self-certifying → 1; nested inner workflow →
  1, attributed to `inner_ask` while the guarded outer workflow stayed silent.

### One thing I did NOT build, and why

The handoff's third shape — a chassis startup scan. It needs a Go change, a
council round and an image roll to add a layer that catches, at the next roll, a
subset of what L1 catches on demand today. Revisit only if L1 turns out to be run
too rarely; that is a usage fact nobody has yet.

---

## 2026-07-27 ~11:0x — post-roll re-verification (chassis v1.0.1172)

A fresh chassis build was deployed (pod `agent-chassis-7f88c4bd7f-bhhbf`, image
`v1.0.1172`, started `2026-07-27T10:55:44Z`). **R1 itself is unaffected by a roll**
— it is Python and docs, nothing inert — but 076's guard lives in the binary, so
its liveness is a claim about the NEW pod, not about git.

**Guard still live on the new binary.** `guard:1 REFUSED:1 degrade:1`, negative
control `0`. And the source is byte-identical since `511670fc8`:
`git diff 511670fc8 HEAD -- ai_actions.go truncation_guard.go
diagnose_council_decide_action.go` is empty.

**Config unaffected by the roll:** the lint reports 171 definitions, 37 tolerating
steps, 37 guarded, exit 0. No new tolerating config arrived with the roll.

**Post-roll traffic:** 6 LLM calls, 4 of them council `review_*` steps — the
guarded path has run on the new binary. **0 tolerated, 0 refused**: no response
has been cut since the roll, so this is evidence of DEPLOYMENT and of the path
running, not of the guard's behaviour. The behaviour was proven by induction on
v1.0.1169 and the code has not changed since.

**Unattended evidence the consumer half keeps working:** `TRUNCATION_DEGRADED_REVIEW`
rows are now **6**, up from the 3 the handoff recorded, latest `2026-07-26
21:55:10Z`. Three real degradations were captured after that file was written,
with nobody watching.

### FINDING — the handoff's positive control asserted a number that is not a property of the code

Its §6 recipe says `grep -c "tolerate_truncation"` → **3**. On v1.0.1172 it
returns **4**, with the 076 source unchanged.

`strings` does not print one Go string literal per line. The linker packs string
data into blobs; `strings` prints each blob; so `grep -c` on a common substring
counts **blobs that happen to contain it**, and where those blobs split moves
between builds for reasons unrelated to your change. (Visible directly: piping
that grep through `cut -c1-95` shows merged fragments — `stop_reason=refusal`,
`heap_released_bytes`, an SQL `DO NOTHING`, an OCI runtime description — all on
one "line".)

**So: a pod-grep control must assert PRESENCE (`>=1`) and a negative control at 0,
never equality on a count.** An equality assertion fails on an innocent rebuild,
and a control that fails on a correct binary is worse than none — it sends the
next thread hunting a regression that does not exist. The handoff is corrected in
place with the evidence; `WRONG_CALLS.md` has the entry, because the pod-grep
family already has a tally there.

**R5 re-checked, unchanged:** `platform/orchestration/orchestration_test.go:171`
still fails `go vet` (`NewSagaCoordinator` called with 3 args, needs 4). Still not
ours; `platform/orchestration/actions` — where all of 076 lives — still compiles
and passes.

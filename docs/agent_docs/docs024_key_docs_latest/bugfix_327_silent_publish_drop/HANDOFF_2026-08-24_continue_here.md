# HANDOFF — bug 327, the silent publish drop — 2026-08-24

**Read this first, then `NOTES_silent_publish_drop.md` (evidence + every misstep) and
`RUNBOOK_silent_publish_drop.md` (the commands, with their gotchas).**

**For the read-aloud version — what this is and why it matters — take the SUMMARY SERIES in
order:** `SUMMARY_2026-08-23_…` then `SUMMARY_2026-08-24_…`. The second exists because two of
the first's headline claims **inverted** (the fix became reviewable; the remaining-work figure
was corrected and re-scoped). **Read both — the 08-23 file is deliberately not amended**, and
the pair is the record of how the understanding moved.

> ⚠ **TWO BUGS SHARE THE NUMBER 327. Resolve by SLUG.** This lane is
> `bugs_closed/327b_HANDOFF_2026-08-19_the_build_trigger_can_publish_nothing_and_exit_zero.md`.
> The other — `bugs_closed/327_..._a_partial_spec_write_silently_shrinks_the_brief...` — was
> closed 2026-08-23 by `copy_quality_two_stage` and is unrelated. **Almost every commit in
> `git log --grep 327` is the OTHER one.** A live handoff already conflated them; corrected in
> `loanzy_uk_example_site/HANDOFF_2026-08-23_garden_tools_continue_here.md`.

## 0. STATE AS OF 2026-08-25 — the work is DONE; what remains is three decisions

`[MEASURED 2026-08-25, re-derived not assumed]` **live racing queue: 0 · adopters: 21 · library
self-test: pass · my files untouched by other lanes.**

**A fresh chassis build was deployed and is IRRELEVANT to this lane** — re-checked today across
every commit this lane made: **zero** touch `platform/`, `internal/`, `pkg/` or `cmd/`. Everything
here is shell and Python, live the moment it was committed. No roll was ever pending, and none can
change anything here. (Say this explicitly to anyone who asks "has it shipped?" — the answer is
that the question does not apply.)

**The landmine-verifier verdicts came back NEEDS_HUMAN_REVIEW on both 327 entries and that means
"cannot check", not "doubtful"** — its code index is `.go`-only and every footprint is a `.sh`.
Answered by hand on 2026-08-25 and written into both entries.

**⚠ One finding from doing that, and it is the third of its kind here:** checking that `097` prints
its receipt *after* publishing reported **WRONG** — because the file now contains a COMMENT quoting
the landmine's own title, and the grep found the comment before the `echo`. The claim was right.
**In this lane, your own writing about the trap contaminates every grep you build to detect it.**
Three occurrences in four days; `WRONG_CALLS.md` 2026-08-25 has the tally and the habit that fixes
it (`sed 's/#.*//'` on every hazard grep, including throwaways).

## 0a. CLOSING AUDIT, 2026-08-25 — what is actually outstanding

Audited rather than asserted. `[MEASURED 2026-08-25]`

| check | state |
|---|---|
| live racing queue | **0** |
| adopters of the library | **22** — three of them lanes this one never spoke to (`380`, `384`, `140_tool_suggester`) |
| library self-test | ALL PASS |
| uncommitted work | none |
| bug renamed to disambiguate | done — `327b` |
| standing five + closing summary | complete (3-summary series) |
| open decisions | both DECIDED and recorded (`HANDOFF_2026-08-25_open_decisions.md`) |
| concept register | **OPP-009** |
| `LANDMINES` entries | 2, both human-verified after the verifier returned NEEDS_HUMAN_REVIEW |
| **`016b` §9 entry** | **was MISSING — added 2026-08-25.** The case file and landmine both existed, but neither serves a session arriving *with the symptom*, which is what §9 is for |

**ONE thing is outstanding, and it is a single `git mv`:** `bugs_closed/327b_…` has **not** been
moved to `bugs_closed/`. Everything else is done. See §0b for the two traps in doing it.

## 0b. WHEN YOU CLOSE IT — two things that bite on the way out

**(1) `[MEASURED 2026-08-25]` 66 files reference the string `bugs_open/327`** — 35 under
`docs/agent_docs`, 11 under `scripts/initial_messages`, 4 under `platform/orchestration`, plus the
library, the detector and several migrated scripts. **The moment the file moves, all 66 point at a
path that does not exist.** There is a standing landmine on exactly this shape (*closing a bug does
not retract the deferrals pointing at it* — 20 days were lost to one). Most of these are
*explanatory citations* ("why this code looks like this"), not deferrals, so they are far less
dangerous than that case — but they are dead paths.
**Cheapest honest fix: leave the citations, and do NOT rewrite 66 files.** A reader who greps
`327` finds it in `bugs_closed/`. What matters is (2).

**(2) BOTH 327s will then live in `bugs_closed/`, so the ambiguity gets WORSE, not better.**
Today you can disambiguate by directory — open = the dispatch drop, closed = the partial spec
write. After the move you cannot. `bugs_closed/` will hold:

```
327_HANDOFF_2026-08-19_a_partial_spec_write_silently_shrinks_the_brief_the_writer_reads.md
327b_HANDOFF_2026-08-19_the_build_trigger_can_publish_nothing_and_exit_zero.md
```

**So: resolve by SLUG, always** — `scripts/who-owns.py 327` already prints the ambiguity warning,
and a live handoff has already conflated the two once (corrected in
`loanzy_uk_example_site/HANDOFF_2026-08-23_garden_tools_continue_here.md`). Worth saying in the
closing commit message so the next reader meets it there.

## 1. State in one paragraph

A shared, receipt-asserting Kafka publisher (`scripts/kafka-publish-lib.sh`, register
**OPP-009**) is **built, live and adopted by 10 callers** — 8 by this lane, **2 by lanes nobody
asked**. Everything here is shell/Python, so **it was live the moment it was committed; no image
roll is involved and the fresh chassis build changes nothing for this lane** (it shipped zero Go
— verified by checking every `327*` commit for files under `platform|internal|pkg|cmd`). The
customer path the bug names is fixed. The bug **stays OPEN and tracks the class** (owner ruling
2026-08-24).

## 2. What "migrating the remaining publishers" actually means — and why the number is NOT the goal

**One migration = replacing a ~20-line block** of the form

```bash
kubectl -n kafka run -i --rm "kcat-x-$(date +%s)" --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -b <broker> -t <topic> -H ... < "$MSG_F" >/dev/null 2>&1
```

with a call to `kafka_publish_checked`, plus two things that are part of the fix and easy to
skip: **move any `SAVE:`/`CORRELATION_ID` print to AFTER the receipt**, and give the failure path
a message naming what did not happen. ~15 minutes each, and each needs its own induced-failure
test. The recipe is in §5.

**`[MEASURED 2026-08-24]` the class is 155 runnable racing publishers — but the real queue is 11.**

| slice | count | verdict |
|---|---|---|
| runnable racing publishers | **155** | the headline |
| …under `docs/` (lane one-offs) | 86 | **do not touch** |
| …under `scripts/` | 66 | the shipped surface |
| …`scripts/` **and** touched in 30d | **11** | **this is the work** |
| literal duplicates `(1)`,`(2)`,`(4)` | 6 | download artefacts, junk |
| files matching only in COMMENTS | 18 | not publishers at all |

**Why "migrate all 155" is the wrong goal:**

1. **86 live in `docs/` lane directories** — one-shot scripts recording what a lane did once.
   Rewriting them **falsifies the record**, and there is a documented landmine that a
   `*_TRIGGER_*`/`*_SUBMIT_*` script **publishes on every invocation** — so editing and testing
   one risks firing a live dispatch.
2. **102 of the 155 have not been touched in 30 days.** Dormant.
3. **The detector already stops the class growing** — `check_kcat_stdin_race` fires on any commit
   that *adds* the racing form (1.7% over 300 commits, 5/5 true positives).

**So the class is bounded by a detector, not by finishing a sweep.** What is left is the
*shipped, still-used* surface.

## 3. THE QUEUE IS DONE (2026-08-24, later)

> **All 11 migrated. `[MEASURED 2026-08-24]` the live queue re-derives to ZERO**, and there are
> **21 adopters** of the library, all parsing, self-test green. The three closing conditions in
> §4 are therefore MET — the decision to close is the owner's, and it is the one open item.
>
> Migrated in four batches: the five `fire-*` operator tools; the file/herestring group
> (`backfill-`/`regen-meta-descriptions`, the two `140_tool_suggester` creators, `074b`); the
> heredoc group (`081b`, `081_reconcile_plan_noted`, `081f`, `074_cta`); and the last two
> (`082_trigger_rerender_site_noted`, `074c` email sweep).
>
> Three judgements made per file rather than by rule, worth knowing if you extend this:
> **(a)** in-function publishers `return` instead of `exit`, so one undispatched item does not
> tear down a sweep that may already have done real work (`081b`, `074_cta`, `074c`);
> **(b)** heredoc delimiters were left **unquoted** when lifting the payload into a variable —
> quoting one would silently ship literal `${VAR}` placeholders; **(c)** every `SAVE:` /
> `CORRELATION_ID` print was moved below the receipt, and verified absent on the failing path.

## 4. (historical) The queue as it stood — 11 files

```
scripts/backfill-meta-descriptions.sh
scripts/regen-meta-descriptions.sh
scripts/initial_messages/140_tool_suggester/075_create_noted_legacy_rescue_tool.sh
scripts/initial_messages/140_tool_suggester/076_create_noted_write_tool.sh
scripts/initial_messages/170_work_item_flow_build/081b_bind_noted_experiences.sh
scripts/initial_messages/250_site_design_planner/081_reconcile_plan_noted.sh
scripts/initial_messages/130_section_editor/074_section_editor_noted_cta_urls.sh
scripts/initial_messages/130_section_editor/074b_section_editor_noted_privacy_copy.sh
scripts/initial_messages/130_section_editor/074c_section_editor_noted_email_sweep.sh
scripts/initial_messages/001_assemble_all_pages_rerender/082_trigger_rerender_site_noted.sh
scripts/initial_messages/001_assemble_all_pages_rerender/081f_rerender_pages_for_ai-agent-orchestration.com.sh
```

Re-derive rather than trusting this list — it goes stale by addition:

```bash
grep -rl "kcat -P" --include="*.sh" . \
  | while read f; do sed 's/#.*//' "$f" | grep -q "run -i" && bash -n "$f" 2>/dev/null && echo "$f"; done \
  | grep '^scripts/' \
  | while read f; do d=$(git log -1 --format=%at -- "$f"); [ "$d" -gt "$(date -d '30 days ago' +%s)" ] && echo "$f"; done
```

## 5. Can the lane close? — recommendation

**Yes, and the bar is the 11 above, not 155.** Suggested closing conditions:

1. The 11 live `scripts/` publishers migrated and induced-failure tested.
2. The detector left in place (it is; advisory, in `.githooks/pre-commit`).
3. `docs/` one-offs **explicitly declared out of scope** in the bug file, so the next reader does
   not treat 86 dormant files as outstanding work.

**Then close `bugs_open/327` → `bugs_closed/`.** The residual (86 historical + dormant files) is
DATA, not an open defect — the same disposition the other 327 took.

**What would NOT justify keeping it open:** the raw 155. That number is dominated by files nobody
runs, and treating it as a backlog will keep this bug open for ever while nothing improves.

## 6. How to migrate one (the recipe, proven 19×)

```bash
REPO_ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel 2>/dev/null || true)"
if [ -z "$REPO_ROOT" ] || [ ! -f "$REPO_ROOT/scripts/kafka-publish-lib.sh" ]; then
  echo "ERROR: scripts/kafka-publish-lib.sh not found — refusing to publish unverified (bugs_open/327)." >&2
  exit 1
fi
. "$REPO_ROOT/scripts/kafka-publish-lib.sh"

PUBLISH_RC=0
kafka_publish_checked \
  --topic <topic> --correlation "$CORR" --payload "$(cat "$MSG_F")" \
  --header "orchestration_id=$ORCH" --header "..." || PUBLISH_RC=$?

if [ "$PUBLISH_RC" -ne 0 ]; then
  echo "NOT DISPATCHED — <what will not happen> (bugs_open/327)." >&2
  exit "$PUBLISH_RC"
fi
```

Then: **move the correlation print below this block**, and test with the broker unreachable:

```bash
KAFKA_PUBLISH_BROKER=nonexistent-broker.invalid:9092 bash <script> <args>   # expect exit 10
```

**Exit codes.** `kafka_publish_checked`: 0 ok · **10 not published (retry now)** · 11
indeterminate (verify first) · 2 usage. `kafka_verify_landing`: 0 landed · **12 consumed and
REFUSED** · **13 published, not landed (wait)**. They start at 10 because callers own 1 and 2
(`097`'s documented contract).

## 7. Traps specific to THIS work

- **⚠ A census of this trap counts the WARNING COMMENTS about it — including the ones you add
  when you migrate a file.** 18 files match on comments alone. **Migrating a file does not move a
  naive count**, which is how I published "178 remaining" while the real figure was 160. Always
  `sed 's/#.*//'` first. (`pattern-check.py`'s detector strips comments and was never fooled.)
- **⚠ A `.sh` extension is not a script.** 23 of the racing files do not parse — pasted SQL, no
  shebang. `076_trigger_build_pipeline.sh` (1,101 lines, no shebang) and `077_submit_domain.sh`
  (fails `bash -n`, and hardcodes `DOMAIN="idea.uk"` *after* the argument parse) are both
  scrapbooks. **Do not "fix" one — it makes a dead file look runnable.**
- **⚠ Verify at the SOURCE, not at `orchestration_states`.** That table retains **~2 days**, and
  `min(created_at)` reads a month back because the purge exempts stuck rows. A zero-row lookup on
  anything older than two days is the retention window, not a drop.
- **⚠ Rule out a REFUSAL before blaming the publisher.** A validation refusal produces identical
  absences and IS recorded in `agent_error_log` (~30 days). `kafka_verify_landing` now does this
  automatically (code 12).
- **⚠ Induce the SILENT arm.** An unreachable broker exits **1 loudly on the unfixed form**, so
  that induction cannot come out false. The real arm is empty stdin against a *healthy* broker:
  zero messages, **exit 0**, no output.

## 8. Open decisions for the owner — these are ALL that is left

> **These two now have their own file, deliberately outside the bug so closing it cannot bury
> them: `HANDOFF_2026-08-25_open_decisions.md`.** It carries the measurement behind each and the
> trigger that would make B urgent.

**Nothing is blocked and nothing is in flight.** The lane is complete to its stated bar; these
three are judgement calls, not work items.

1. **Close `bugs_open/327`?** All three closing conditions written into the bug file are MET
   (11 live publishers migrated + tested · detector in place · scope statement standing). It is
   still open only because "keep it open and track the class" was an explicit instruction and
   reversing that is the owner's call. **Recommendation: close it** (owner agreed 2026-08-25; see §0b for the two things that bite on the way out). What remains after closing is
   not work — ~55 dormant `scripts/` files (untouched 30d+) and the `docs/` one-offs that must not
   be rewritten — and the commit-time detector catches any of them the moment someone picks one up.
2. **Spend credits on a council round?** It is now **legitimately in scope** (the 2026-08-24
   widening admitted `scripts/pattern-check.py`), so this is no longer a `FORCE` question — purely
   whether the review is worth the credits. Submission written and validated:
   `COUNCIL_SUBMISSION_2026-08-23_publish_receipt.json`, `DRY_RUN=1` → admitted, exit 0.
   **No recommendation either way** — the load-bearing artefact (`kafka-publish-lib.sh`) is still
   out of scope, so a round would review the detector and not the publisher.
3. **Build Layer 3 — the in-cluster Go submit path?** `platform/kafka.Producer` is already
   `RequiredAcks: RequireAll, Async: false`, so it would make a silent drop **unrepresentable**
   rather than detected, and close the ~2-day forensic window. Filed proposed-and-unbuilt in
   **OPP-009**. **Recommendation: not now** — the receipt closes the live exposure, and this needs
   Go, a council round, an image and a fleet roll. Revisit if a drop is ever observed *through* the
   library, which would be the evidence that the `kubectl attach` receipt is not enough.

## 9. Verify the lane in 60 seconds

```bash
bash scripts/kafka-publish-lib.sh --self-test              # expect ALL PASS (11/11)
DRY_RUN=1 ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh \
  docs/agent_docs/docs024_key_docs_latest/bugfix_327_silent_publish_drop/COUNCIL_SUBMISSION_2026-08-23_publish_receipt.json  # expect 0
python3 scripts/pattern-check.py --commit HEAD             # detector loads, 22 checks
```

## 10. What this lane is really about (for whoever picks it up)

The safe publish form had been documented in `LANDMINES.md` **for a month**. On 2026-08-23,
**25** scripts printed the `PUBLISH_OK` receipt it prescribes and **2** asserted on it. Twenty-three
authors read the guidance, followed it, and still exited 0 when the receipt was absent — and the
`fire-*.sh` family literally carried *"⚠ kcat -P EXITS 0 HAVING SENT NOTHING"* in its own header
while using the racing form underneath.

It became **callable** on the 23rd. Within a day, **two lanes nobody spoke to adopted it**
(`caa55f04f`, and `bugfix_380`'s trigger). **The missing thing was never the knowledge.** If you
are tempted to fix a class like this by writing the warning down more clearly, that is the
evidence against.

# NOTES — bug 327, the build trigger can publish nothing and exit 0

Append-only, newest at the bottom. Technical log: evidence, commands, what the system
actually said, and every misstep.

---

## 2026-08-23 — session start, evidence gathering

**Disambiguation first.** There are TWO bug 327s and `scripts/who-owns.py 327` says so
explicitly. Resolving by slug, as CLAUDE.md requires:

- `bugs_open/327_..._the_build_trigger_can_publish_nothing_and_exit_zero.md` — THIS one.
- `bugs_closed/327_..._a_partial_spec_write_silently_shrinks_the_brief_the_writer_reads.md` — closed 08-23 by the `copy_quality_two_stage` lane, unrelated.

`who-owns.py` conflates them: it reports 24 mentions in `copy_quality_two_stage` and a
"bug 327 CLOSED" commit, **all of which belong to the other 327.** `git log` on the FILE
PATH separates them cleanly:

```
git log --format="%ad %h %s" --date=short -- "bugs_open/327_HANDOFF_2026-08-19_the_build_trigger_can_publish_nothing_and_exit_zero.md"
# 2026-08-19 db375212c  bug 327: the build trigger can publish nothing and exit 0 ...
```

**One commit, ever — the filing.** The open 327 is UNOWNED and untouched. Confirmed no
dirty files under `scripts/` in the working tree, and no open `site_work_items` naming
this mechanism.

### The bug is still valid — established at the SOURCE, not at the DB

`scripts/initial_messages/020_build_pipeline/082_submit_domain_unified.sh` is unchanged
since **2026-07-30** (`95639d4f6`), i.e. before the bug was filed on 08-19. Lines 161-178
still carry every element of the defect:

```
161: kubectl -n kafka run -i --rm kcat-submit-$(date +%s) \
164:   kcat -P \
178:   -H responses_topic=system.agent.generic.responses <<JSON
```

`run -i` + payload on stdin, no `--command`, no `PUBLISH_OK` receipt, no landing check.

### MISSTEP 1 — my first re-verification query could not have come out otherwise

I ran this to re-confirm the drop:

```sql
SELECT count(*) FROM orchestration_states
WHERE correlation_id = '8fa2a4a6-2af1-4675-bae9-bbd59b702160';
-- 0
```

Zero. It would have been easy to write "re-verified 2026-08-23, still zero rows". **That
would have been worthless**, and I nearly had it. The control that killed it:

```sql
SELECT created_at::date, count(*) FROM orchestration_states GROUP BY 1 ORDER BY 1;
-- 2026-07-19 |    6
-- 2026-07-20 |    8
-- 2026-07-21 |    6
-- 2026-07-24 |    4
-- 2026-08-22 | 1417
-- 2026-08-23 | 3140
```

**`orchestration_states` holds ~2 days.** There are ZERO rows for the whole of 2026-08-18.
The query returns 0 whether the message landed or not. This is exactly the shape CLAUDE.md
warns about — *"a `[MEASURED]` figure is only evidence if the measurement could have come
out otherwise"* — and the cheap check is one `GROUP BY` on the date. Logged in
`WRONG_CALLS.md`. **The validity of the bug rests on the SOURCE (above), not on this query.**

### The competing explanation, ruled out properly

On 2026-08-20 another lane (`41c06f1d1`) blamed kcat for what turned out to be a
**validation REFUSAL recorded in `agent_error_log` all along**. A refusal produces the same
three absences as a drop, so it must be excluded, not assumed away. It was not checked in
the original bug file.

```sql
SELECT occurred_at, agent_type, error_code, LEFT(error_message,90)
FROM agent_error_log
WHERE context::text ILIKE '%8fa2a4a6%' OR orchestration_id ILIKE '%8fa2a4a6%'
   OR error_message ILIKE '%8fa2a4a6%';
-- (0 rows)
```

An absence, so it needs a positive control — is the instrument even alive on that date?

```sql
SELECT count(*) AS rows_on_18th,
       count(*) FILTER (WHERE error_code='VALIDATION_ERROR_DROPPED') AS refusals_on_18th
FROM agent_error_log WHERE occurred_at::date='2026-08-18';
-- 3761 | 1
```

**3,761 rows including one real `VALIDATION_ERROR_DROPPED` on the very day.** The recorder
was live and the refusal path demonstrably fires. Its silence about `8fa2a4a6` is therefore
evidence. **The refusal explanation is ruled out; the drop stands.**

### Retention asymmetry — and it is an argument, not a footnote

| table | retained from | rows |
|---|---|---|
| `orchestration_states` | ~2 days (08-22 onward) | 4,580 |
| `agent_error_log` | 2026-07-24 (~30 days) | 46,141 |

So **a dropped submission is forensically unrecoverable after ~48h.** You can always ask
later "was it refused?"; you can never ask later "did it land?". That asymmetry is why the
fix has to be a receipt at publish time and not better retrospective diagnosis.

### Framework-wide census `[MEASURED 2026-08-23]`

| | count |
|---|---|
| shell scripts publishing via `kcat -P` | **218** |
| using the racing `kubectl run -i` + stdin form | **200** |
| printing a `PUBLISH_OK` receipt | **25** |
| **actually asserting on that receipt** (grep/if → non-zero exit) | **2** |

The two that assert: `analytics_gtm/scripts/fire_reassemble_site.sh`,
`idea_uk_vm_site/scripts/fire_reassemble_idea_uk.sh`.

**The finding that matters most: the documented remedy is itself only half-applied.** 23 of
the 25 "fixed" scripts print a receipt for a human to read and still exit 0 when it is
absent. A receipt nobody asserts on is not a control — it is a log line.

### The class is live, not historical

2026-08-22 — one day before this session — a landmine was filed for a `097` council
dispatch that exited 0 having published nothing: 90 minutes with no row while other lanes'
runs executed; re-dispatch produced a row in 3 seconds (`7760963cf`). Same class, a
different trigger, current.

---

## 2026-08-23 — the failure INDUCED, and a correction to the bug file's own verification

CLAUDE.md and the bug file both insist the failure must be induced, not assumed. I ran two
inductions. Both publish **zero messages, so both are side-effect free.**

### Induction A — unreachable broker (this is the recipe the bug file proposes)

```bash
( set -euo pipefail
kubectl -n kafka run -i --rm kcat-induce-a-$(date +%s) --image=edenhill/kcat:1.7.1 \
  --restart=Never -- kcat -P -b nonexistent-broker.invalid:9092 \
  -t system.agent.generic.requests -H correlation_id=INDUCTION-TEST-A <<'JSON'
{"action":"orchestrate","config":{"agent_type":"NONEXISTENT-INDUCTION-TEST"},"input_data":{}}
JSON
)
```

```
% ERROR: Local: Host resolution failure: ... Name does not resolve
% ERROR: Local: All broker connections are down: 1/1 brokers are down: terminating
pod kafka/kcat-induce-a-... terminated (Error)
>>> EXIT CODE: 1
```

**Exit 1, loudly — on the UNFIXED script.**

### Induction B — real broker, empty stdin (what a LOST RACE actually looks like)

```bash
( set -euo pipefail
kubectl -n kafka run -i --rm kcat-induce-b-$(date +%s) --image=edenhill/kcat:1.7.1 \
  --restart=Never -- kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests -H correlation_id=INDUCTION-TEST-B-NEVER-SENT < /dev/null )
```

```
pod "kcat-induce-b-..." deleted from kafka namespace
>>> EXIT CODE: 0
```

**Exit 0. Zero messages. No error output of any kind.** The pod is reported as deleted
normally — which is precisely the sentence the 327 filing quotes as reassuring.

### > **CORRECTION to `bugs_open/327` § "How to verify a fix"**

The bug file says: *"Run the trigger with the broker deliberately unreachable (or the topic
renamed) and require a non-zero exit."*

**That recipe is already satisfied by the UNFIXED script** — induction A exits 1 today,
because `set -euo pipefail` propagates `kubectl`'s non-zero status when the pod terminates
in Error. A fix verified that way would be verified against a control that **cannot come
out false**, which is the exact trap CLAUDE.md's measurement-discipline rule names.

The two failure modes are **opposite in observability**, and the file's recipe tests the
wrong one:

| induced condition | messages published | exit | operator-visible |
|---|---|---|---|
| broker unreachable | 0 | **1** | loud broker errors |
| **stdin lost / empty** | 0 | **0** | **nothing at all** |

**So the verification must induce the SILENT arm**: an empty or unattached stdin against a
*healthy* broker. `< /dev/null` reproduces the post-race state deterministically and with no
side effects, and is the control any candidate fix has to fail on before it is believed.
"Topic renamed" is also weak — brokers with auto-create will accept the publish and exit 0,
so it tests neither arm.

**What this means for the fix:** the receipt is not belt-and-braces on top of an exit code
that mostly works. For the failure mode that actually bites, **the exit code carries no
information at all**, and the receipt is the only signal that exists.

### The receipt mechanism tested against all three known traps

Before designing around the receipt, I checked it actually discriminates. Both tests
publish nothing (unreachable broker), so both are side-effect free.

**Test 1 — safe form, unreachable broker.** `--command -- sh -c "… | kcat -P … && echo
PUBLISH_OK"`. Result: broker errors, pod terminated (Error), **receipt ABSENT**. The `&&`
chain does suppress the marker on a failed publish, as intended.

**Test 2 — the `--command` omission trap.** Same command with `--command` dropped, so
`sh -c …` arrives as *arguments to kcat* (the image ENTRYPOINT). Result: kcat's usage text,
pod terminated (Error), **receipt ABSENT**.

`[MEASURED 2026-08-23]` So an **asserted** receipt is a single control that catches all
three known ways this fails silently:

| trap | exit code today | receipt catches it |
|---|---|---|
| stdin race / empty stdin (the 327 case) | **0 — no signal** | yes |
| broker unreachable | 1 | yes |
| `--command` omitted (ENTRYPOINT trap) | 1 | yes |

The middle and bottom rows are already loud; **the top row is the one with no other
signal**, and it is the one that produced 327. That is the argument for the receipt being
*asserted* rather than printed: it is the only instrument that reads the silent arm, and 23
of the 25 scripts that emit it never look at it.

### Where the operator's false confidence is manufactured — it is an ORDERING defect too

`082_submit_domain_unified.sh`:

```
158: echo "SAVE: CORRELATION_ID=${CORRELATION_ID}  ORCHESTRATION_ID=${ORCHESTRATION_ID}"
161: kubectl -n kafka run -i --rm kcat-submit-$(date +%s) \
```

**The script tells the operator to SAVE an identifier three lines before it first attempts
to send it.** Both ids are generated locally by the script itself (lines 148-149,
`/proc/sys/kernel/random/uuid`) — they are not issued by the broker and they do not depend
on anything having happened. So the most reassuring line of output, the one phrased as an
instruction to record the id for later, is printed unconditionally and is exactly as
confident on the failing path as on the succeeding one.

This matters for the fix beyond tidiness: even after a receipt is added, **if the summary
block still prints before the publish, the operator's transcript still reads as a success**
above whatever error appears below it. The ids must be printed *after* a confirmed publish,
or explicitly marked as unconfirmed until one arrives.

---

## 2026-08-23 — V2 run: the race did NOT reproduce today, and the old form duplicated instead

The decisive experiment the plan called for: 10 publishes by the old racing form and 10 by
the new library, into `system.agent.scratch-327-publishrace.requests` (a scratch topic
nothing consumes; naming follows the existing `system.agent.scratch-069-chromelock`
precedent). Then count what actually landed with `kcat -C -e`.

```
old: kubectl reported success on 10/10 invocations
new: receipt asserted on          10/10 invocations
OLD landed: 11        NEW landed: 10
```

Broken down by message id:

```
OLD:  n=1 -> TWICE;  n=2..10 -> once each     (10 distinct, 11 delivered)
NEW:  n=1..10 -> once each                    (10 distinct, 10 delivered)
```

### > **This is NOT the result I expected, and it must not be written up as a success.**

**The drop did not reproduce. Zero of 10 old-form publishes were lost.** The library form
is structurally immune (the payload rides in the container command; stdin is not involved
at all) and delivered cleanly with an asserted receipt on every attempt — but **I have not
demonstrated that the new form beats the race, because the race did not fire.**

**What the sample can and cannot say** (the estate's rule: name the failure rate your
sample could DETECT):

- The historical rate — 4 in 5 lost, measured 2026-07-26 — is **decisively excluded for
  today's conditions**: P(0 drops in 10 | p=0.8) = 0.2^10 ≈ 1×10⁻⁷.
- With 0 drops in 10, the **95% upper bound on today's drop rate is ~26%**
  ((1−0.26)^10 ≈ 0.05). Anything below that is entirely consistent with what I saw.
- So: the race is **timing- and load-dependent**, and today the cluster is not producing
  it at anything like the rate that was originally measured. `[MEASURED 2026-08-23]`

**This does not weaken the bug, and here is why.** 327 is not a claim about frequency; it
is a claim that when the publish is lost the loss is **silent and indistinguishable from
success**. That property is unchanged and was verified directly earlier today (empty
stdin, healthy broker → zero messages, exit 0, no output). And the class is demonstrably
still live: the 2026-08-22 `097` incident is five days after the 08-18 one.

### The unexpected half: the racing form published a DUPLICATE

`n=1` landed **twice** from a single invocation. Distinct pod name, distinct payload, a
run id unique to this run — so it is not contamination from an earlier test.

That is a second, independent pathology in the same mechanism, and it is arguably worse
than the drop for the customer path: **a duplicate submission on the real topic is two
builds, or two orchestrations racing one site.** It also compounds with `bugs_open/326`
from the opposite direction — 326 is about a retry doing nothing, and this is the trigger
doing something twice without being asked.

`[MEASURED 2026-08-23 — ONE observation, n=1 of 10.]` One duplicate in ten is not a rate,
and I am not claiming one. Recorded because it was observed under controlled conditions
and because nothing in the bug file or the landmine anticipates it. The new form showed no
duplicate, but with n=10 that is not evidence of immunity either.

**What would settle both questions** and is deliberately NOT claimed here: a run of
several hundred publishes each way, under load, on a day the cluster is busy. Neither the
drop rate nor the duplicate rate is established by 10 samples, and this note should not be
quoted as though either were.

---

## 2026-08-23 — Phase 1: 082 migrated, and every arm verified against live data

### The publish path

```
$ KAFKA_PUBLISH_BROKER=nonexistent-broker.invalid:9092 \
    bash scripts/initial_messages/020_build_pipeline/082_submit_domain_unified.sh \
    induction-327-notreal.example --email test@example.com

NOT PUBLISHED  topic=system.agent.generic.requests correlation=9cbd57b6-…
  Nothing landed. RETRY NOW — a retry collides with nothing.
  | % ERROR: Local: Host resolution failure: … Name does not resolve
SUBMISSION DID NOT GO OUT — nothing has been queued for induction-327-notreal.example.
Re-run: scripts/…/082_submit_domain_unified.sh induction-327-notreal.example --email test@example.com
>>> EXIT CODE: 10
>>> correct: no 'SAVE:' line on a failed publish
```

Four properties, each of which was false before: **non-zero exit**, a message naming what
did not land, **the `SAVE:` line suppressed** on the failing path, and a re-run line that
actually works.

**A defect found while writing that last one.** The re-run hint was first written as
`$0 ${DOMAIN} ${*}` — and `$*` is **empty** by then, because the option loop has shifted
every argument away. A retry hint that silently drops `--mission-file` or `--from` is worse
than no hint at all: the operator re-runs a *different* submission and believes it is the
same one. Fixed by capturing `ORIGINAL_ARGS=("$@")` before the loop
(`082_submit_domain_unified.sh:96`). Nothing would have caught this — the line only prints
on a path nobody exercises.

### `kafka_verify_landing` — all three arms, against the live database

| arm | input | result |
|---|---|---|
| **0 LANDED** | a real recent correlation (`daf6bbe8…`) | `LANDED … COMPLETED|complete` |
| **12 CONSUMED AND REFUSED** | `c66c480d…`, a real `VALIDATION_ERROR_DROPPED` with **no** orchestration row | quoted the code and the message |
| **13 PUBLISHED, NOT LANDED** | a freshly generated UUID that never existed | latency guidance, no retry advice |

The **12** arm is the one worth dwelling on: it is exactly the case a lane got wrong on
2026-08-20, when it blamed kcat for a drop while `agent_error_log` held the delivery record
the whole time. That misdiagnosis is now mechanical rather than a matter of remembering to
check — and it is discriminated on live rows, not fixtures.

### What is NOT verified end-to-end, and why

**V5 (exit 11, indeterminate receipt) is verified at the classifier, not end-to-end.** With
the payload in the container command, the remote is `… && echo PUBLISH_OK`: if kcat
succeeds the marker prints, so an exit-0-without-receipt cannot be induced from outside the
library — it requires a genuinely lost output stream. `_kafka_classify 0 ""` → 11 is covered
in the self-test. Stated rather than glossed, because "all verified" would be a false claim.

**The healthy end-to-end path (V6) is deliberately NOT run.** Submitting a real domain
creates a site row and consumes every stage `item_key` for ever (`bugs_open/326`), so it is
not a test you can run twice. The publish half was exercised 10/10 against the scratch topic
in V2; the landing half was exercised against real landed correlations above.

---

## 2026-08-23 — the two "Phase 1 siblings" are NOT scripts, and my census over-counted

The plan named `077_submit_domain.sh` and `076_trigger_build_pipeline.sh` as the siblings to
migrate alongside 082. **I named them from their filenames without reading them.** Reading
them says do not touch either:

**`076_trigger_build_pipeline.sh`** — 1,101 lines, mode 664, **no shebang**, opening with a
bare `INSERT INTO build_queue …` and pasted `psql` session output. It contains two copies of
the racing publish as *quoted command text inside a scrapbook*. It is a notes file with a
`.sh` extension.

**`077_submit_domain.sh`** — has a shebang and looks like a script, and **does not parse**:

```
$ bash -n scripts/initial_messages/020_build_pipeline/077_submit_domain.sh
line 117: syntax error near unexpected token `wi.summary,'
line 117: `       LEFT(wi.summary, 70) as summary'
```

Pasted SQL in the body, a bare `---` separator at line 26, and hardcoded `DOMAIN="idea.uk"`
assignments *after* the argument parse that would override whatever the operator passed.
Mode 664. **It cannot be executed at all.**

**So neither is migrated**, and that is the right call twice over: it would be churn, and
"fixing" a file that cannot run would make it *look* runnable — manufacturing a new trap
rather than closing one.

### > **CORRECTION to my own census, and it halves the headline**

I reported **200 scripts using the racing form**. That counts *files containing the
pattern*, which is not the same as *publishers that can run* — and the difference is large:

```
racing-form files:                 201     (201 today, not 200 — the set grows)
  ...that PARSE (bash -n):         178
  ...that are EXECUTABLE (+x):     106
  ...both parse AND +x:            105
```

`[MEASURED 2026-08-23]` **23 of the 201 cannot run at all** — scrapbooks like the two above.
A file that parses but lacks `+x` can still be run as `bash <file>`, so the honest exposure
is **178 runnable racing publishers, of which 105 are marked executable** — not 200.

Both numbers were true; they answer different questions, and I quoted the one that flatters
the case. The number that matters for exposure is 178. The earlier figures in this file and
in the RUNBOOK are corrected accordingly, and the detector below should count the same way.

---

## 2026-08-23 — the landmine VERIFIER publishes with the racing form, and says "0 failed to publish"

Found by using it: arming my own LANDMINES contribution meant running
`./scripts/landmines-verify-dispatch.sh`, which printed

```
Dispatched 2, 0 failed to publish.
```

**That sentence is the bug, printed by the landmine system about itself.**

- `scripts/trigger-landmine-verifier.sh:84` publishes with `kubectl -n kafka run -i --rm …
  kcat -P` — the racing form, payload on stdin, no `--command`, no receipt.
- `scripts/landmines-verify-dispatch.sh:45-62` increments `FAILED` **only if that script
  returns non-zero**. On the silent arm it returns **0**. So "0 failed to publish" is
  computed from the one signal that is absent precisely when a publish is lost.

The tool that verifies landmines is subject to the landmine it verifies, and reports success
in the words most likely to be false. Nobody would notice: a verdict that never arrives looks
like the async wait the script's own closing message tells you to expect.

**A second gap, found the same way:** appending a CONTRIBUTED bullet to an *existing* entry
does **not** mark it changed for dispatch. The sweep picked up two unrelated entries and not
mine. CLAUDE.md's escape hatch is the per-entry trigger, and it works:

```bash
./scripts/trigger-landmine-verifier.sh 'LANDMINES.md#kubectl-run-i-rm-kcat-p-file-drops-roughly-4-publishes-in-5-at-exit-0-and-with-b'
```

### The library earned its keep on someone else's dispatch

Rather than trusting that trigger's `exit 0`, I asked:

```
$ kafka_verify_landing 80404f33-1520-4632-9ebd-0d2684f34670 45
LANDED  correlation=80404f33-…  EXECUTING_STEP|spawn_verifier
>>> returned 0
```

Positive evidence that the verifier run actually started, on a real dispatch this lane did
not author. Before today the available evidence was "the trigger exited 0", which is exactly
what it also does when nothing was sent.

**`scripts/trigger-landmine-verifier.sh` is therefore the top Phase 2 migration candidate** —
ahead of `097` — because it is a *shipped* tool whose whole purpose is verification, and its
failure is invisible by construction. **Not migrated in this lane**, which was scoped to the
filed case; recorded here so the next session does not have to rediscover it.

---

## 2026-08-23 — the council was asked, and REFUSED on scope (exit 2)

Submission written in full and put through the gate's admission test, which is free:

```
$ DRY_RUN=1 ./…/097_TRIGGER_council_review_v1.sh COUNCIL_SUBMISSION_2026-08-23_publish_receipt.json
REFUSED: no edit touches the review scope.
  In scope: platform/, internal/, pkg/ … cmd/config-key-audit/ … sql_for_agents/NNN_name.sql
>>> exit: 2
```

It cleared every schema check — rationale, ≤8 edits with file/operation/rationale/sketch,
`grounded_in` strings, `risks` as prose, no comment-only sketch, single repo-relative paths —
and was refused **only** because all three edits live under `scripts/`.

**This is a real answer, not a failure, and it is recorded rather than forced.** `FORCE=1`
would override it; I have not used it, because the scope boundary is an owner ruling and
spending fleet credits against it is the owner's call, not a session's. The submission file is
kept here so a forced round costs nothing to start.

**So: the load-bearing artefact of this fix ships UNREVIEWED.** Compensations actually
delivered: the library's offline self-test (11/11), every runtime arm proven against live data,
the detector measured for precision before wiring in, and both directions controlled.

### > A scope observation worth the owner's attention

**`scripts/pattern-check.py` is detector logic the gate cannot see — the same shape as the gap
the owner closed TODAY.** The 2026-08-23 widening admitted `cmd/config-key-audit/` on the
stated reasoning that *"the detector logic for the daily check fleet has accumulated in one
binary the gate cannot see"*. `pattern-check.py` now carries **22 advisory checks** that run on
**every commit in every session** via `.githooks/pre-commit` — a larger and more frequently
executed detector surface than the nightly CronJob fleet — and it is entirely out of scope. A
false positive there is, in the pre-commit hook's own words, *"a fleet-wide commit outage"*.

Not proposed as a change from this lane: widening council scope is an architecture-scope
decision and the owner has just made a deliberate, measured, targeted one. Recorded because
the parallel is exact and nobody appears to have noticed it while making that ruling.

---

## 2026-08-23 — closing checks, and prior art I should have grepped for first

**Final state:** library self-test 11/11 · `082` parses · `pattern-check.py` loads with 22
checks including `check_kcat_stdin_race` · `landmines-sync.py --check` → **in sync, exit 0** ·
the kcat entry's `doc_notes` rows are current (written 18:01 and 18:09 today).

**Two false alarms of mine at the end, both logged as `WRONG_CALLS` 9:** `to insert/refresh:
747` is a **total, not a delta**, and landmine `doc_notes` rows are **verifier verdicts, not
entry bodies** (841 chars average) — so a `LIKE` against my own prose was never going to match.
Both caught by running a control instead of writing the claim.

### Prior art this lane failed to grep for — and one piece corrects me

Surfaced by accident, in the sync's own output. **CLAUDE.md says to grep LANDMINES for the
symbol you are about to trust; I did not, in a session about landmines.**

1. **`orchestration_states` retains about TWO DAYS — but `min(created_at)` reads over a month
   back, because the purge EXEMPTS stuck rows.** Filed 2026-08-23 by another lane, hours from
   my own measurement, and it **corrects the check I published**: my `WRONG_CALLS` 8 recommended
   `min(created_at)` alongside the `GROUP BY`, and `min()` is precisely the misleading reading —
   the survivors are 24 stuck `CANCELLED` rows the purge skips. Their day-plot
   (`08-23: 3,299 · 08-22: 1,324 · then nothing until four July dates`) matches mine
   (`3,140 · 1,417`) taken hours apart. Entry 8 now carries the correction.

2. **`097_TRIGGER_council_review` prints its `SAVE: SUBMISSION_CORR=` receipt BEFORE it
   publishes.** The ordering defect I found at `082:158` is therefore a **second instance of a
   documented class**, not a discovery. That strengthens the case rather than weakening it: two
   independent triggers, written years apart, both print the operator's reassurance before
   attempting the send — which is what a *class* looks like, and what the shared library fixes
   by construction (ids print only after a receipt).

**The transferable bit:** grepping LANDMINES costs seconds and I skipped it because I already
knew the landmine I was working on. The entries that would have helped were about
`orchestration_states` and `097` — neither of which is the symbol I thought I was studying.
**Grep for the symbols you TOUCH, not the ones your task is named after.**

---

## 2026-08-24 — owner rulings applied: council scope widened, verifier trigger migrated

### 1. `scripts/pattern-check.py` enters council scope (OWNER RULING 2026-08-24)

The 2026-08-23 widening admitted `cmd/config-key-audit` because *the detector logic for the
check fleet had accumulated where the gate could not see it*. The owner accepted that the same
argument applies to `pattern-check.py`.

`[MEASURED 2026-08-24]` **2,058 lines carrying 22 checks**, against **2,220 lines for every
other `audit-*`/`check-*` script under `scripts/` combined**. So one file is roughly half the
non-Go detector surface — and the half that runs **on every commit in every session** via
`.githooks/pre-commit`, rather than once a night.

**Targeted, anchored with `$` to the one file.** Verified across eight paths:
`scripts/kafka-publish-lib.sh` stays **out**, and so does `scripts/audit-advisory-findings.py`
(429 lines) even though it **imports the `CHECKS` tuple** — it reports on findings rather than
deciding them, and a reporting bug cannot block a commit. That coupling is written down at the
definition site so it can be revisited rather than rediscovered.

### > The trap I would have walked into, caught by a warning placed there one day earlier

**Widening the regex is NOT enough.** `098_REPORT` must *enumerate* candidate commits before
`in_council_scope` can judge them, and `git log` takes pathspecs, not regexes — so it carries
`SCOPE_PATHS`, a hand-kept array, as a **pre-filter**. A path added to the regex and not to
that array is **invisible** to the coverage report: not listed as unreviewed, *absent*, which
reads as nothing to report.

That failed on 2026-08-23 — the day `cmd/` was added — hiding **22 in-scope commits across four
lanes**. The lane that found it wrote the warning **beside the regex**, on the reasoning that
*"a warning only in 098 would be read by whoever edits 098, who is not the person with the
problem."* I was the person with the problem, one day later, and it worked exactly as intended.

**Both halves changed in one commit, and proven end to end:**

| check | 2026-08-23 | 2026-08-24 |
|---|---|---|
| `DRY_RUN=1 097` on this lane's submission | **REFUSED, exit 2** | **admitted, exit 0**, nothing dispatched |
| `d000f07c5` (a `pattern-check.py` commit) in the `098` report | **no bucket at all** | listed under **UNREVIEWED** |

`CLAUDE.md` was corrected too: it did not mention the 08-23 widening at all, and its line
*"097, the commit-msg nudge and the 098 report all read it, so do not re-hardcode it"* is
exactly what let 098's second copy stay invisible. It now says single-sourced for the
**decision**, and *widen both, in one commit*.

**The council was NOT run** (owner: do not force). Note the widening makes this lane's own
submission legitimately admissible now rather than forced — spending the credits remains the
owner's call, and `DRY_RUN` proved admission for free.

### 2. `scripts/trigger-landmine-verifier.sh` migrated (Phase 1b)

Migrated ahead of `097` because its failure is invisible **by construction**: the caller counts
a dispatch failure only on this script's non-zero exit, and the old form exited 0 on the silent
arm — so `Dispatched N, 0 failed to publish` was computed from the one signal that is always
absent when a publish is lost.

**The caller needed no change.** `kafka_publish_checked` returns 10 or 11, both non-zero, so
`FAILED` now counts what it claims to.

Proven both ways:

```
unreachable broker → exit 10
  VERIFICATION NOT DISPATCHED for <slug> — no verdict will ever arrive for this run.

healthy → exit 0, PUBLISHED
  kafka_verify_landing → LANDED  EXECUTING_STEP|spawn_verifier
```

The healthy run was a real dispatch, and it arms the verifier for the kcat entry this lane
edited — the thing that silently did not happen yesterday.

**Adoption: 2 of ~178.** `097` is next; its exit codes **1 and 2 are reserved** by a documented
contract, so a publish failure must take a distinct code there.

---

## 2026-08-24 — Phase 1c: `097` migrated, and its landmine closed in all three halves

`097_TRIGGER_council_review_v1.sh` was the last of the three named hot dispatchers. Its own
LANDMINES entry listed three defects; the migration closes each:

1. **`SAVE: SUBMISSION_CORR=` printed before the publish** — so the operator recorded a trail
   id for a submission that might never have been sent, and any error appeared **below** the
   summary and the "if APPROVED, commit with…" advice, where the eye has already read success
   and stopped. It now prints only after an asserted receipt.
2. **The payload rode on `kubectl run -i` stdin** — the race itself.
3. **The pod name carried only second resolution** (`kcat-cgate-$(date +%s)`), so two sessions
   submitting in the same second collided with `AlreadyExists` and nothing published. The
   library's names carry `$RANDOM` too.

**The exit-code contract is why the library's codes start at 10, and it is now verified rather
than assumed:**

| arm | expected | got |
|---|---|---|
| `DRY_RUN=1`, in scope | 0 | **0** |
| missing submission file | 1 | **1** |
| out-of-scope submission | 2 REFUSED | **2** |
| publish failure | 10 | **10**, and **no `SAVE:` line** |

The failure message names the fact that decides whether re-running is safe — *"Nothing was
spent. Re-run the same command; the submission file is unchanged."*

**The landmine entry is updated, not retired.** The reading habit it taught — read the tail of
the output, verify at the row — is still correct on the ~175 unmigrated triggers, where an error
still prints below the summary and a silent drop still exits 0. Generalising this fix to a
trigger that has not had it is the mistake the entry now warns against.

**End-to-end proof through the migrated chain:** arming the verifier for that very entry went
through the migrated `trigger-landmine-verifier.sh` → `PUBLISHED` with a receipt → landing
confirmed at `EXECUTING_STEP|spawn_verifier`.

**Adoption: 3 of ~178.**

---

## 2026-08-24 — my census was wrong in a way my own fix made worse, and a stranger adopted the library

### > **CORRECTED: the exposure figure is 160, not 178 — and migrating a file made the naive count go UP**

`[MEASURED 2026-08-24]`, comments stripped:

| | count |
|---|---|
| `kcat -P` publishers | **219** |
| racing form, **counting comments** (what I quoted) | 201 |
| racing form **in code** | **183** |
| racing in code **and** parses — *the exposure* | **160** |
| callers on the checked library | **4** |

**Two independent inflations, both upward, and I only caught the first yesterday.** 23 files are
scrapbooks that cannot run (found 2026-08-23). A further **18 match only inside comments** —
warnings *about* this very hazard, written by people trying to help.

**The part that is my own doing:** every file I migrate gets a comment block explaining what the
racing form was and why it went. **That comment keeps matching a naive census for ever.** So the
three migrations I made today moved the naive count by **zero** — the number stops moving exactly
when the work starts working, and it does so in the direction that flatters nothing and confuses
everything.

This is `check_stdin_eater`'s documented lesson one level up: *a detector that flags its own
warning text teaches people to ignore it.* I honoured that rule inside the **detector** —
`pattern-check.py` strips comments and was never fooled — and then broke it in the **census
one-liner** I published in the RUNBOOK for everyone else to use. The rule was in my hands and I
applied it to one of the two instruments.

**Corrected command** (`RUNBOOK`, now authoritative):

```bash
grep -rl "kcat -P" --include="*.sh" . \
  | while read f; do sed 's/#.*//' "$f" | grep -q "run -i" && bash -n "$f" 2>/dev/null && echo "$f"; done \
  | wc -l
```

Figures corrected in the RUNBOOK, OPP-009 and the bug file. Logged in `WRONG_CALLS`.

### A different lane adopted the library within a day, unprompted

`scripts/initial_messages/140_tool_suggester/077_update_noted_write_tool.sh`, commit
`caa55f04f` (2026-08-24): *"077: publish via kafka_publish_checked instead of the racing
kubectl-run stdin form."* Nobody asked them, and this lane never spoke to them.

**That is the evidence for the central decision of this whole piece of work.** The safe form had
been documented in `LANDMINES.md` for a month and had **2 assertions** to show for it. It became
*callable* on the 23rd and a stranger reached for it on the 24th. The thing that was missing was
never the knowledge — it was the absence of something to call.

**Adoption: 4 of 160, one of them not mine.**

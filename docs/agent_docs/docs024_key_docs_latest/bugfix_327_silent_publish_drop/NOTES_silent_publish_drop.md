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

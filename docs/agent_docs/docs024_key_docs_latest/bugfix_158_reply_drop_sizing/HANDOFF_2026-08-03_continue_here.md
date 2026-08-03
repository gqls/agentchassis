# HANDOFF — bugfix_158_reply_drop_sizing — 2026-08-03

**Cold-start doc.** Read this, then `README_where_we_are.md` (owner-facing prose)
and `RUNBOOK_reply_drop_sizing.md` (the commands, each with the gotcha that nearly
falsified it). The ticket itself is `bugs_open/158`, which now carries three dated
sections at the foot — read those before anything above them.

## State in one paragraph

`bugs_open/158` is a container ticket with four findings. **Item 4 was already fixed
by someone else** (07-31). **Item 2 needs nothing** — its `[UNMEASURED]` is now
measured and no consumer of `storage.pages` exists. **Item 0 was surveyed and handed
over**, not actioned, because it is `bugs_open/100`'s pipeline. **Item 1 was ruled on
by the owner on 2026-08-03 and the ruled scope is SHIPPED.** What remains open is
**item 3** (an owner decision) and the four deliberately-deferred adapter sites.

## What shipped today

| commit | what |
|---|---|
| `2091903ab` | `check_silent_reply_drop` in `scripts/pattern-check.py` — closes the class prospectively, no behaviour change |
| `2d976c026` | **option (b)**: `DeliverReply` adopted at the three plumbing sites + coordinator tests + a detector false-positive fix |
| `fca47d869` | agentbase adoption tests — the council's one real gap, mutation proved |
| `b759bb072` | ticket updated with the ruling and the census correction |
| `300c04bfd`, `c34fbaadd` | 016b correction, WRONG_CALLS |

**Council `f13212d4-2787-448f-bf49-b57506ded74e`: APPROVED round 1**, 3 advisory
objections, none high. Acted on: the `editquality` seat's medium — my submission
invoked 158's landmine about tests living in the caller's package and then tested
only the coordinator — closed by `fca47d869` (agentbase adoption tests, mutation
proved). Still standing, and worth a glance rather than action:

- `guardian` (medium x2): the owner-ruling claim is checkable and should be checked
  — it is recorded in `bugs_open/158`'s 2026-08-03 section and in the ruling summary
  above; the seat is right that a plan asserting an owner ruling should be verified,
  not taken on trust.
- `guardian`/`debug_historian` (medium): the timeout-monitor site still has no test
  (no repository fake in the package), and the plan named no post-deploy
  verification — the commands ARE below, they simply were not in the submission.
- `editquality`/`guardian` (low): the `pattern-check.py` false-positive fix rode in
  on a core-plumbing change and diluted the plan's minimality. Fair; it was a
  correction of my own error and I did not want it sitting unfixed.

```sql
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = 'f13212d4-2787-448f-bf49-b57506ded74e';
SELECT body FROM diagnosis_artifacts WHERE kind='council_report'
 AND correlation_id='f13212d4-2787-448f-bf49-b57506ded74e' ORDER BY created_at DESC LIMIT 1;
```

`2d976c026` carries `Council-Submitted:` and `fca47d869` carries `Council-Reviewed:`
for the same correlation, so `098`'s join resolves both.

## The owner's ruling, so it is not re-litigated

Presented with four options and the sizing, the owner chose **(b): the plumbing
sites only** — `agentbase`, the saga coordinator, the timeout monitor — because
every agent inherits them, and **explicitly not** the four adapter/agent sites.
Those four stay as they are and are held by the detector. Adopting them later is a
**fresh decision**, not a continuation of this one.

## NOT YET LIVE

`2d976c026` is committed but the chassis running at the time of writing is
**v1.0.1238**, which predates it. The fix is inert until a roll.

**Verify at the pod, both controls, every replica** (`bugs_open/153`):

```bash
for POD in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name); do
  kubectl -n ai-persona-system exec $POD -- sh -c '
    echo -n "  notifying parent of FAILURE (want >=1): "; strings /app/agent-chassis | grep -c "notifying parent of FAILURE instead"
    echo -n "  ERROR_FORWARD_TOO_LARGE   (want >=1): "; strings /app/agent-chassis | grep -c "ERROR_FORWARD_TOO_LARGE"
    echo -n "  OLD line, want 0:                     "; strings /app/agent-chassis | grep -c "Failed to notify parent of success"'
done
```

The third is the negative control — a string this change **removed**. Differential
the old and new images before pushing if you can; a control that does not MOVE
between them is not a control (see `WRONG_CALLS`, 2026-08-02).

## Traps this lane hit, so you do not

1. **There is no Kafka broker in `ai-persona-system`.** It is
   `personae-kafka-cluster-combined-pool-prod-{0,1,2}` in namespace **`kafka`**. The
   ticket's suggested `kafka-configs.sh` command therefore returns *empty*, which
   reads exactly like "no override is set". Never `2>/dev/null` a probe whose empty
   result would be the finding.
2. **`llm_call_log`'s `agent_type` labels are not the service names.** Filtering by
   `'%reasoning%'` etc. returns 0 rows, which reads as "no data". Enumerate the
   column before filtering on it.
3. **Scrape action names are not the obvious ones** — the live vocabulary is
   `scrape_web`, `batch_webscrape`, `firecrawl_scrape`, `firecrawl_extract`,
   `fetch_scrape`, `med_scrape_prices`, `format_crawl_for_analysis` (13 distinct).
4. **`coordinator.go`'s site moved `:3663 → :3721` mid-session** because another
   session committed to it. **Find these sites by symbol, not by line number.**
5. **My own detector produced a false positive that I propagated into two
   documents.** `processor.go:2000` returns its error two lines below the block and
   its function is dead. A detector's output is a list of *candidates*.

## The numbers, so they are not re-derived

- Broker `message.max.bytes` = **1048588** (DEFAULT_CONFIG; no override on the reply
  topic). **97%** of `*.responses` topics DO carry a 5MB override; the **91** that do
  not are the long-lived per-agent ones, including `system.agent.generic.responses`.
- Largest payload the fleet has ever produced: **48,327 bytes** across **47,577**
  `llm_call_log` calls. **Zero** above 512KB. So the size-refusal path is **latent**.
- ⚠ That sizing bounds the **SIZE** failure only. All these sites swallow **transient**
  broker failures identically, and reply sizes say nothing about how often those occur.
  Do not quote "latent" as though it covered both.
- Census: the rule holds at **2 of 12** (the ticket's own "2 of 9" keyed on one log
  string; my first correction said 13 and was one too many).
- Detector sweep: **4** findings now, all of them the deferred adapter/agent sites.

## What is owed, in order

1. ~~Read the council verdict~~ **DONE — APPROVED, and its one real objection is
   closed** (`fca47d869`). Nothing further owed to the gate.
2. **Roll and pod-verify** — the three adoptions are inert until then. This is now
   the FIRST thing owed.
3. **Item 3 needs the owner**: the uploader stores 2 per-page fields, the truncator
   cuts 6, only `markdown` overlaps — so a `firecrawl_crawl` result loses page
   content over 50KB with no recoverable copy, on 4 live steps. Options: upload what
   is truncated, truncate only what is uploaded, or leave it (honest marker, content
   still lost).
4. **Decision 2 from the owner's list is still open**: align the 91 reply topics onto
   the 5MB override, or deliberately keep replies at 1MB? It is config, reversible,
   and it is a prerequisite for item 1b(a).
5. Then `158` can close, or be split — it can never close as a unit while item 3
   waits on a human.

## Related lanes — do not collide

- `bugs_open/100` owns the vet scrape pipeline (item 0's target). Coordinate.
- Someone is actively editing `platform/orchestration/coordinator.go` (they added
  `extractWorkflowResultWithSizeLimit`). My change sits next to theirs and agrees
  with it; re-read before editing.
- `bugs_open/181` (filed by this session's earlier work on `172`) is the fourth
  instance of the silent-cap family. Its first question is not the fix — it is why
  **0 of 276** retained diagnosis bundles contain a rendered `code_check` block.

---

## 2026-08-03 (later) — LIVE, pod-verified on both replicas

Chassis rolled to **v1.0.1243**. Verified on both running pods:

| check | pod `mxjt7` | pod `wxbbg` |
|---|---|---|
| `notifying parent of FAILURE instead` (added) | 2 | 2 |
| `ERROR_FORWARD_TOO_LARGE` (added) | 1 | 1 |
| `Failed to notify parent of success` (removed, negative control) | 0 | 0 |
| `PARTITION BY agent_type` (172, added) | 1 | 1 |
| `\t\tLIMIT $2` whole-line (172, removed, negative control) | 4 | 4 |

Option (b) is now genuinely live, not merely committed and approved. The "roll and
pod-verify" step from this handoff's "what is owed" list is **done** — the only
items still owed are item 3 (owner decision) and the reply-topic alignment decision,
both listed above.

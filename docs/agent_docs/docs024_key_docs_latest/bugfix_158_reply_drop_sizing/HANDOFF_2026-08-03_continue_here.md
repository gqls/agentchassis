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


---

# 2026-08-03 (evening) — decisions 1, 2, 5 actioned; 3 and 4 explained, awaiting the owner

## Decision 1 (item 3: "upload everything that gets truncated") — SHIPPED, council PENDING

Commit `a92a32fba`. `internal/adapters/webscrape/{adapter,truncation,truncation_test}.go`.

**FIRST THING TO CHECK in the new thread:**
```sql
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = 'ee9f6210-3bda-4efa-a25c-92ce4a7666a1';
```
It was stuck at `review_debug_historian` for 13+ minutes when this was written — an
outlier against every other round this lane ran (2-6 min). May have landed by now,
may need a nudge, may be worth asking why that seat specifically stalled if it
recurs. If APPROVED: commit trailer needed on a FOLLOW-UP commit (any tiny doc
touch will do) reading `Council-Reviewed: ee9f6210-3bda-4efa-a25c-92ce4a7666a1` —
or just let `098` resolve it automatically from `Council-Submitted:`, already on
`a92a32fba`. If REVISE/REJECTED: the code is already on the shared branch, revise
forward.

**What it does:** two layers. `uploadScrapingResults` (adapter.go) now uploads all
6 per-page truncatable fields instead of 2 — and fixes a real key-name bug found
while extending it: the old code checked `pageMap["html"]`, which never matched
`batch_handler.go`'s own key `html_content` for the same concept, so per-page HTML
was silently never archived for that path even with `upload_results:true`.
`truncateResultForTransport`/`stripResultForRetry` (truncation.go) gain a
`FieldUploader` fallback: the moment a field is about to be cut and has no static
URI yet, the FULL pre-truncation content is uploaded on the spot — closing the gap
for the 18 of 22 live scrape/crawl steps that never set `upload_results` at all.

**NOT YET LIVE.** Needs: council approval (or a decision to proceed on REVISE
notes) → chassis roll → pod-verify. Positive control: grep the running binary for
`"could not preserve"` (the fallback's own failure-path log line) or
`fieldUploaderFor` — either string only exists post-fix. No natural negative
control string was retired here (unlike 172/158b), since this ADDS a capability
rather than replacing one; verify via the artefact instead once live: a
`diagnosis_artifacts`-style check isn't applicable, but a live scrape with
`upload_results:false` and content >50KB should now show a `storage.<field>_uri`
key that pre-fix would have been entirely absent.

**Stated residual, not a bug:** `stripResultForRetry`'s re-cut fields
(markdown_content, html_content, per-page content/markdown/markdown_content) do
NOT get a new fallback call — only raw_html does, because it's the one field
deleted outright with no size check. The gap (content between
`oversizeStripContentCap` and 50,000 chars that was never over the FIRST cap) is
left open on purpose: this whole function has never fired on measured fleet
traffic (max reply ever recorded: 48KB).

**A landmine for whoever touches this file next:** per-page S3 keys now put the
field name in a SUBDIRECTORY (`pages/<field>/page_%d.<ext>`), not the filename.
`pageStorageURIFor` resolves a page by searching for the literal needle
`/page_<i>.` immediately before the extension — putting the field name AFTER
`page_` (e.g. `page_1_html.html`) breaks that match silently. Verified against
`TestPageURIIndexMatchIsNotAPrefixMatch` before shipping; keep it that way if you
touch the key format again.

## Decision 2 (align under-provisioned reply topics) — DONE, live immediately

Kafka topic config only, no code, no roll. 74 `*.responses` topics were on the 1MB
broker default (re-measured fresh at execution time — NOT the earlier "91", topics
churn continuously); all 74 now carry `max.message.bytes=5242880`, plus one
(`system.thunder.smoke.responses`) that appeared mid-batch and needed a second
pass. **Final live verification: 0 of 3,828 `.responses` topics missing the
override**, re-checked after the fact, not assumed from the batch script's own
count.

## Decision 5 (unblock bugs_open/100) — DONE

`UPDATE agent_definitions ... jsonb_set(..., '{workflow,steps,scrape_website,config,upload_results}', 'true')`
for `vet-practice-verifier`, live, DB config only. Verified via `RETURNING`. A note
was ALSO added directly to `bugs_open/100` (`e4de1a7d0`), since the owner is
starting that bug in a separate thread and the pointer needs to live where that
thread will look, not just here.

## Decisions 3 and 4 — explained in chat, not actioned

**3** (widen the reply-delivery fix to the 4 remaining adapter sites): explained in
more depth; still not a live decision — nothing forces it, the detector just
prevents a fifth silent site. **4** ("should replies carry full content up to the
bus limit") was clarified (a "reply" here means the Kafka message sent back on the
`*.responses` topic, not an HTTP response) and left for the owner to confirm he
still wants it, now that decision 1 removes the main reason it used to be risky
(nothing is destroyed either way now — decision 4 would ONLY change how large the
INLINE reply itself gets, not whether content survives). Not implemented.

## Everything from this pass's earlier sections (option-b plumbing fix, verified
live on `v1.0.1243`) is unaffected and stays as documented above — this section is
additive.

---

# 2026-08-03 (night) — all five decisions closed out; decision 1 LIVE on the adapter

Owner delegated the remaining decisions ("do what you think best"). Final state:

| decision | outcome |
|---|---|
| 1 — upload what's truncated | **SHIPPED + LIVE on `web-scrape-adapter:v1.0.1245`**, all 3 replicas pod-verified (positives 1/1, negative control 0). Council re-round pending, see below |
| 2 — align reply topics | **DONE earlier** — 0 of 3,828 `.responses` topics missing the 5MB override |
| 3 — the four adapter reply sites | **DECIDED: not adopting.** Latent twice over, class capped by the detector. Reopen trigger: one observed refusal → adopt all four in one round |
| 4 — full-content replies | **DECIDED: no.** The 50KB inline cap stays; decision 1 removed the correctness argument; readers (LLM prompts, `collected_data`) want URI + head. Reopen trigger: a consumer needing >50KB *inline* |
| 5 — unblock 100 | **DONE earlier** — `upload_results` live on the vet verifier step |

## THE DEPLOY TARGET LESSON — read this if you touch item-3 code again

The truncation code (`internal/adapters/webscrape/`) ships in the
**`web-scrape-adapter`** image, NOT agent-chassis. A chassis roll does nothing for
it. `make build-web-scrape-adapter IMAGE_TAG=<tag>` + `make deploy-web-scrape-adapter`.
Pod-verify greps `/app/web-scrape-adapter` (label `app=web-scrape-adapter`,
three replicas). Verified controls: `could not preserve content that is about to be
truncated` (added, 1), `/pages/%s/page_%d.%s` (added, 1),
`%s/pages/page_%d.html` (removed, 0) — all three MOVED between 1243 and 1245.

## The council saga for corr `ee9f6210` — two disruptions, neither the submission's fault

1. First run (`e5f7f2f0`): **killed by a chassis roll** at 20:10:22Z — the
   documented "a roll KILLS an in-flight council" landmine, observed precisely
   (row's last touch 20:10:02, new pods 20:10:22). The dead row still sits at
   `review_debug_historian|EXECUTING_STEP`; left alone deliberately.
2. Resubmitted with `RESUBMIT_CORR=ee9f6210…` → run `7ec39ca1`. Progressed
   normally, then `review_architecture` stalled ~11+ min with **failed LLM calls**
   visible in `llm_call_log` (`success=f` for architecture and prior_art at
   21:14) — retry backoff, not a dead row.
3. A Monitor is armed on the terminal state. **Whoever reads this: check
   the verdict** — `SELECT ... WHERE left(correlation_id::text,8)='7ec39ca1'` —
   and act on REVISE/REJECTED (code is live; revise forward). `a92a32fba` carries
   `Council-Submitted:` so 098 auto-credits on approval.

## What would close bugs_open/158 entirely

Item 1: closed (option b live + decision 3 decided). Item 2: no consumer, nothing
owed. Item 4: fixed 07-31. Item 3: decision 1 is live; **the only thing owed is
the council verdict read-and-act**. After that, 158 moves to `bugs_closed/` —
every item has an outcome and none is silently dropped.


---

# ⛔ LANE CLOSED 2026-08-04 — `bugs_open/158` → `bugs_closed/158`. Nothing owed.

Everything above is history. Final state:

- **Council `ee9f6210` APPROVED** (round 4 — rounds 1-3 killed by chassis rolls /
  infra). All 4 advisory objections answered; one **corrected a real error of mine**
  (residual justified with `llm_call_log`, a model-completion statistic, for a claim
  about scrape replies — see `WRONG_CALLS.md` 2026-08-04).
- **LIVE and pod-verified on the CURRENT images**, both rolled by other sessions
  after my deploys: `agent-chassis:v1.0.1247` (option b + `bugs_closed/172`) and
  `web-scrape-adapter:v1.0.1246` (item 3). Negative controls 0 on every replica.
- All six ticket items discharged; two decided-no with written reopen triggers.
- §10 index row added; §9 pattern added; `LANDMINES` entry for the deploy-target
  trap; four `WRONG_CALLS` entries.

**If you are a fresh session that opened this file looking for work: there is none
here.** Read `bugs_closed/158_HANDOFF_2026-07-30_*.md` for the account. Live
follow-ons that came OUT of this lane and are still open:

- **`bugs_open/181`** — the fourth silent-cap instance (filed by this lane's earlier
  `172` work). First question is not the fix: why do 0 of 276 retained diagnosis
  bundles contain a rendered `code_check` block?
- **The four adapter reply sites** — decided-no, held by `check_silent_reply_drop`.
  Reopen only on one observed refusal, then all four in one round.
- **Roll-killed council rows** — two now sit in `EXECUTING_STEP`/`FAILED` forever
  with no reaper, from four rolls in one evening. Not filed as a bug by this lane;
  adjacent to `bugs_open/173`/`029` if someone wants it.

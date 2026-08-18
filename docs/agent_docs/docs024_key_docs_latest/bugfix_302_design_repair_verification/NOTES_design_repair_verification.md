# NOTES — `bugs_open/302` design-repair completion verification

Append-only, newest at the bottom. Evidence, commands, and every misstep.

---

## 2026-08-18 — session opens: scoping, ownership, and what the filing got wrong

### Ownership checks, before touching anything

- `scripts/who-owns.py 302` → OWNED/recently active, `finetuning_uk_service`. **Checked the live
  transcript rather than trusting the commit log** (per the landmine "an ownership check that
  reads COMMITS cannot see the session that is fixing it right now"): that session filed 302,
  appended its SCOPING UPDATE (`b6f869676`), and wrapped up with a docs summary. Its current
  work is GPU/training Phase 0, not this bug. **Not competing — it is the FILING lane, not a
  fixing lane.**
- `scripts/who-owns.py 201` → `bugfix_201_page_content_writer_dispatch`, last commit 2026-08-09.
  Its `HANDOFF_2026-08-09` says in terms: *"there is nothing left to do on 201 itself"*.
- Grepped all 40+ live `.jsonl` session transcripts for `bugs_open/302` and `bugs_open/201`.
  One session had substantive hits (the filing lane, above); every other hit was an incidental
  `ls bugs_open/` listing.
- **The seam's own lanes:** `bugfix_213_verifier_producer_join` built gate 1b and is FINISHED
  (its handoff: "There is no queued work… Do not pick this up here", and it explicitly hands the
  `NO_CHANGE_GATE_UNREADABLE_RESULT` stream to the `RFC_029` lane as their before/after).
  `work_item_completion_integrity` is the historical registry lane, dormant since 08-08.

### The machinery, read first-hand

Two completion gates, both consulted from `verifyBeforeComplete`
(`complete_work_item_verification.go`), called at `load_work_item_actions.go:982`:

| gate | file | keyed on | opt-in roster | on "cannot read/run" |
|---|---|---|---|---|
| 1b no-change | `complete_work_item_no_change.go` | handler's returned payload counters | `noChangeGates` — **1 type**: `dark_section_audit` | **abstain and COMPLETE** (+ `agent_error_log` row) |
| 2 verifier | `discovery_checks/verifiers.go` | `item_type` | 11 registered types, all discovery shapes | **fail CLOSED** since RFC_017 (owner ruling 08-08) |

The two sibling gates hold **opposite policies on the same question**. Gate 2's own comment
refuses to exempt an unparseable spec because doing so "would leave a second silent completion
path behind the one RFC_017 closed" — and gate 1b, written five days AFTER that ruling, is
exactly such a path for its own unreadable case.

### [MEASURED 2026-08-18] the arm is live and it fired — with a demand control

```sql
SELECT error_code, count(*), min(occurred_at), max(occurred_at)
FROM agent_error_log WHERE error_code='NO_CHANGE_GATE_UNREADABLE_RESULT' GROUP BY 1;
--  11 rows, 2026-08-14 14:24Z → 2026-08-17 12:44Z
```

Not a theoretical arm, and **not a broken gate**: in the same window gate 1b correctly BLOCKED
4 rows (`status='failed'`, `_verification.status='handler_reported_no_change'`). The gate works
when it can read the payload. Both directions observed, so the measurement could have come out
otherwise.

### [MEASURED] the filing's account of the 11 payloads is WRONG, and this matters

302 attributes the population to the handlers returning analysis blobs. The actual shapes:

| result top-level keys | rows | what it is |
|---|---|---|
| `agent_id,agent_type,role,topics` | **7** | a **SPAWN RECORD** — `bugs_closed/287`'s defect |
| `color_scheme,design_notes,spacing,typography` | 3 | the design-token blob 302 names |
| `add_to_page,approach,new_page,not_actionable,reasoning,retype_existing,update_spec` | 1 | an unrelated child-page triage decision |

`bugs_closed/287` is **fixed, live and proven** on chassis `v1.0.1307` (roll 2026-08-17 17:05Z;
fleet now `v1.0.1309`): its own close-out measures `field=result` resolver rows at **0**, down
from ~455/day, with 11/11 loop completions carrying the handler's reply. So the majority cause
of unreadable payloads was removed at source yesterday.

### [MEASURED] and therefore: the arm has had ZERO demand since that roll

```sql
SELECT count(*) FROM site_work_items
WHERE item_type='dark_section_audit' AND updated_at > '2026-08-17 17:05:00+00';   -- 0
SELECT count(*) FROM site_work_items
WHERE status='complete' AND updated_at > '2026-08-17 17:05:00+00';                -- 1862
```

The fleet is busy; this item type specifically is not. **All 11 abstentions predate the roll.**
So the honest position, and the one the fix must be argued on:

- the **structural** hole is real and untouched by 287 — an opted-in type whose payload cannot be
  read is silently exempted from its own opt-in, by construction, for ever;
- the **observed rate** post-roll is **unmeasured, with zero demand**, and I must not claim a
  continuing flood. 7 of 11 are attributable to a bug that is now closed;
- whether the 3 blob rows and the 1 triage row share 287's cause (the recursive `$.**` search
  binding the wrong value into `result`, `RFC_029`'s subject) is **[INFERRED], not established** —
  `bugs_closed/213` §D records it as NOT ESTABLISHED and `RFC_029` owns it. Do not assert it.

### [MEASURED] 302's own working candidate is trap-laden — the producer-split check refutes it

302's SCOPING UPDATE names "Gate-2 artefact verifiers for the design-repair family" as the
working candidate. `LANDMINES.md` mandates a producer split before registering any verifier
(`spec->>'audit_source'` is the only thing that names a producer; `created_by` bottoms out at
`generic`):

```sql
SELECT item_type, count(DISTINCT COALESCE(spec->>'audit_source','<none>')) AS producers,
       string_agg(DISTINCT COALESCE(spec->>'audit_source','<none>'), ' | ') AS which, count(*) n
FROM site_work_items WHERE item_type IN (...) GROUP BY 1 ORDER BY producers DESC;
```

| item_type | producers | which | rows |
|---|---|---|---|
| `needs_design_review` | **4** | brief-fidelity-audit, design-audit, visual-design-audit, `<none>` | 75 |
| `responsive_fix` | 3 | design-audit, visual-design-audit, tool-acceptance-tier4 | 19 |
| `dark_section_audit` | 2 | design-audit, visual-design-audit | 30 |
| `spacing_fix` | 2 | design-audit, visual-design-audit | 30 |
| `contrast_failure` | 1 | `<none>` | 284 |
| `hardcoded_section_colors` | 1 | `<none>` | 9 |

One verifier per `item_type` over a 4-producer population **is** `bugs_closed/213`'s defect: the
verifier is correct about the wrong question, returns `Resolved:true`, and the item closes
untouched. So any gate-2 route for this family owes a `VerifierPolicy.Grades` remit function per
type (contract WII-013) — a materially bigger job than the filing implies, and not the cheapest
thing that closes the door.

### `bugs_open/201` — checked for validity, and it holds up

Not a fix task: both symptoms were fixed, live and proven before 08-08, and RFC_017 (the flip
this lane surfaced) is decided, built and proven. Re-verified today at the artefact rather than
from the lane's own account:

```sql
SELECT status, result->'_verification'->>'status', count(*), max(updated_at)
FROM site_work_items WHERE item_type='literal_markdown' GROUP BY 1,2;
```

- **15** rows `failed` + `_verification.status='defect_persists'` — the verifier REFUSING a
  completion, most recent 2026-08-17;
- **1** row `complete` + `verified` — the verifier CERTIFYING a real repair (08-15);
- both directions present, so this is a discriminating check, not a one-sided pass.

Two `complete` rows carry no `_verification` at all (08-17 13:20Z). **Chased rather than
assumed** — they closed through the discovery-check retraction seam (WII-009), not a handler:
`result.resolved_by='literal_markdown'`, `reason='literal_markdown re-scan: page's unlocked
components carry no markdown syntax on either surface'`. A retraction is the detector's own
measurement, so it legitimately does not run the completion gate. **No residual hole.**

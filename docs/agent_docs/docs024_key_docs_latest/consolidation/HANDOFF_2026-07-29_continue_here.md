# RESUME HERE — consolidation programme (supersedes the 07-28 door)

**Written 2026-07-29 ~00:00Z.** Anchor: `features_open/024`. Sibling cold-start:
`robot_hands_gripper_dossier/HANDOFF_RESUME_gripper_dossier.md`.

**Read the 07-28 handoff too, but read its CORRECTION BLOCKS** — its §4 named the
wrong council objection as gating and its remediation was built on that. Both
corrections are inline there now.

---

## 1. State in one line

The gripper pilot is proven, live and now **council-APPROVED**; the only open
item on this lane is **adopting `platform/mailer` + `platform/httpguard` into
`tools-api`**, which is another thread's code — and as of today that adoption has
a **filed, evidenced bug behind it (`bugs_open/139`)** instead of just a
recommendation.

> **CORRECTED 2026-07-29 (later the same day) — read §4 before acting on that
> sentence.** The probe was run. **`139`'s headline claim is REFUTED**, a
> different and live defect was proven in its place, and **`httpguard` does not
> fix it** — so "adopt httpguard into tools-api" is no longer the next action as
> written. §4 below is corrected in full.

## 2. What is LIVE (verified this session, not inherited)

**Chassis `v1.0.1196`**, two replicas, started 2026-07-28 22:37/22:38Z — a roll I
did not do. Re-grepped both replicas after it:

| marker | count |
|---|---|
| `carries no prose_sections` (created by the fix) | 1 |
| `carries no no_match_sentence` (created by the fix) | 1 |
| `No gripper in this index` (positive control) | 3 |
| `nonexistent_marker_xyz` (negative control) | 0 |

Identical on both pods. **Re-run this after any roll you did not do** — a retag is
not a rebuild.

## 3. CLOSED THIS SESSION — council `721ac4f7` is **APPROVED**

Round 2, 2026-07-28 21:43Z, `decided_by`: *"approved with 1 advisory objection(s)
— none high-severity"*. Trailer is on **`c0dac11ef`**; the platform code shipped
earlier in `7f87c0afa`, which cannot gain a trailer under forward-only, and the
commit message says so.

**The plan did not change — same six edits, same sketches. Only its evidence
changed.** Four seats moved object → clean. Full record:
`robot_hands_gripper_dossier/NOTES_…` §"Round 2 of council 721ac4f7".

**Why round 1 was misread, because it will happen again.** `decided_by` names the
**seat**, not the objection. The 07-28 handoff recorded *"gating objection from
`bug_historian`"* and then attached a **different seat's** content to it (the
DECLARED CONTRACTS point, which was medium, from three other seats).
`bug_historian`'s actual HIGH asked whether anything **blocks** a report the gate
rejects. Answered by reading, not arguing:

- `verify_report_prose_action.go:135-139` returns `(nil, error)` — the violations
  *are* the error text.
- `coordinator.go:3350-3363` `routeToErrorStepOrFail`: step-level `error_step`,
  then config-level, **then `failWorkflow` if neither**. No branch continues to
  `next_step`. Fail-closed by default.
- `verify_prose` wires `config.error_step=handle_failure`,
  `next_step=compose_page`, and `compose_page` is the fleet's only
  `create_report_page` step.

**LANDMINE that nearly reversed this.** Querying `s.value->>'error_step'` returns
NULL for **every** step in report-builder — it reads exactly like "no error
routing anywhere". It lives at `s.value->'config'->>'error_step'`. Already in
`016b` (§~663; census at ~4890: *0 of 14,209 persisted plan steps carry the
step-level twin vs 1,828 carrying the config one*). A result that is uniformly
NULL across every row is almost never a fact about the world.

**OWED, and it is the right ask:** `prior_art` approved but asked that the
contracts-gap claim — *no mechanism declares action-to-action `collected_data`
fields; `input_contract` only fires at the `call_agent` boundary;
`output_contract` has zero readers* — be **independently verified before anyone
cites it as precedent**. I am the only reader so far. Do not cite it as settled.

## 4. NEXT — item 1: the `httpguard`/`mailer` adoption, now with a bug behind it

**`bugs_open/139`** filed today: *tools-api: a visitor can still choose the IP
they are rate-limited as (and the IP we store)*. Read that file first; it has the
evidence, three ordered fix candidates and the verification shape.

What is established:

- `tools-api` keys on gin's `c.ClientIP()` at **two** sites —
  `middleware/ratelimit.go:30` (limiter) and `handlers/round.go:109` (`hashIP`,
  which is **persisted**).
- `internal/tools-api/api/server.go:14` is `gin.New()` with **no
  `SetTrustedProxies`**.
- `platform/httpguard` + `platform/mailer` still have **zero importers**
  (`grep -rl 'agentchassis/platform/httpguard'` → nothing), re-measured today.
- `bugs_closed/090` is the same mechanism, **proven against production** on
  idea.uk. 139 is a second service, not a reopening.

> **CORRECTED 2026-07-29 — measured, and it went the other way.**
>
> Both probes were fired (forged `X-Forwarded-For`, then forged `X-Real-IP`) and
> gin's source was read. **A visitor CANNOT choose the IP.** Both requests
> returned 200 and both stored the same hash as every other row.
>
> **What is real instead: the identity is a CONSTANT.** `client_ip_hash` is
> `sha256("172.18.0.1")` — the docker bridge gateway — in **83 of 83 rows** since
> 2026-07-25 (one distinct value, whole table). So the "per-IP" limiter is a
> single global bucket shared by every visitor, and the stored identity column has
> never distinguished anybody. No attacker required.
>
> **Why:** Caddy overwrites `X-Forwarded-For` with its own peer before the app
> sees it, and Cloudflare strips `X-Real-IP` at the edge. tools-api is exactly as
> trusting as filed (its log carries gin's `[WARNING] You trusted all proxies`);
> it is simply never handed anything to be fooled by. **The protection is two
> other components' defaults, none of it the service's own.**
>
> **This changes the next action.** `httpguard.ClientIP` **does not fix it** — its
> peer gate passes, `X-Real-IP` is absent, and it falls to the rightmost XFF
> entry, which is the same `172.18.0.1`. It would read as a fix while changing
> nothing. The real client address reaches the app only in `CF-Connecting-IP`
> (unforgeable — the edge 403s a supplied one), which nothing currently reads.
> Revised, evidenced fix ordering is in `bugs_open/139`.
>
> **Still true and unchanged:** `platform/mailer` and `platform/httpguard` remain
> approved, correct and with zero importers; the `mailer` half of the adoption is
> untouched by this. What changed is only the *rationale* offered for the
> `httpguard` half — and A3 now has a real design input: the package's docstring
> justifies preferring `X-Real-IP` on a property of **nginx on idea.uk** that
> **Caddy on the island does not provide**. A second adopter inherits the
> reassurance without the mechanism.

**This stays a conversation.** `tools-api` belongs to the **gauntlet_dead_cta**
thread and `bugs_open/083` (slug `gauntlet_engine_503_discards_the_error` — the
number is ambiguous, cite the slug) is open against it. Contribute into 139 and
083; do not fix their service under them.

## 5. NEXT — the rest, unchanged from 07-28

2. **Finish the pilot's public half** — `/api/v1/tools/gripper` **inside**
   tools-api. Do **not** write `cmd/gripper-intake/`; that would be the estate's
   fourth VM fork. Re-seed 208's `base_url` to
   `https://tools.apis.uk/api/v1/tools/gripper`.
3. The two live fixture pages on robot-hands.com await the owner's read; cleanup
   (`source='manual-test'` rows + 2 pages) is owed once seen. They were scored by
   the **pre-fix** code, so do not use them as a reference for current behaviour.
4. **A1 remains a WON'T-DO** — see the 07-28 handoff §5 and `features_open/024`.
   Reopen only if a second site wants a *physics* scorer.

## 6. Open question nobody has measured

`bug_historian`'s surviving advisory asked whether bare `.(string)` assertions on
LLM-parsed maps recur elsewhere unaudited. I showed its premise was wrong
(`bugs_closed/076` is a *truncation-tolerance* mechanism, has **no** shared
safe-extraction helper, and its headline "113 call sites" is a figure that file
itself retracts to *"37 of 118"*) — but the underlying question is real and
**[UNMEASURED]**. A count is needed before it is a claim; do not file it as a bug
on a hunch.

## 7. Landmines this lane paid for

- **A retag is not a rebuild.** Verify by a string your change CREATED, plus a
  positive and a negative control, against the pod running *now*.
- **`decided_by` names the SEAT, not the objection.** Never inherit a council
  verdict through prose — print the severity table, quote the HIGH verbatim.
  Full entry in `WRONG_CALLS.md` (2026-07-28).
- **`error_step` persists under `config`, not at step level.**
- **`scheduled_tasks.target_topic`'s column DEFAULT is a topic NOTHING
  consumes.** Fails silently and looks healthy; only downstream evidence
  discriminates.
- **`create_report_page` requires `request_id` to be a real UUID**; an invalid
  one also silently disables the failure sidecar.
- **`complete`/`deployed_at` is not fetchability.** Poll the URL.
- **A duplication audit sees SHAPE, not USAGE.** Open both files and query live
  usage before calling anything a duplicate.
- **Council submission types:** `operation` ∈ `modify|add|remove|config_change`;
  `grounded_in` is `[]string`; `risks` is one `string`. `097` type-checks all
  three client-side now.
- **Build the submission from the persisted round-1 plan with `jq`**, not by
  retyping it — `SELECT collected_data->'input_data'->'plan'` — so the edits
  carry across byte-exact and only the evidence changes. And put the jq program
  in a **file**: an apostrophe in the program breaks a single-quoted shell string.

# Council trail — 2026-08-21 session (`343` (silent post-abandonment freeze; lane `docs024_key_docs_latest/bugfix_029_retry_kills_live_child/`) + bugs_open/040)

Five coherent tasks, five submissions, **all four live correlations APPROVED**. Recorded here because
one of them cannot be joined mechanically and the reason is my own mistake.

| task | correlation | rounds | verdict | commit(s) |
|---|---|---|---|---|
| 343 park identity + typed outcome + table cross-check | `RESUBMIT_CORR=cc782778…` **(malformed — see below)**, then `a9b9d4b8-c30f-432a-b35d-5e5864733fc5` | REVISE → **APPROVED** | `ca5e41122` (no trailer), `7f3875d3c` (`Council-Submitted: a9b9d4b8…`) |
| 343 P2 loop-skip persistence | `bcea1e9f-ab14-4398-b8b2-16bfe2ca1f64` | **APPROVED** round 1, 2 advisory | `bf1fbc5b7` |
| 040 produce counter + empty-host guard | `a414d81b-30c7-4661-9856-d7f2d669b1a6` | REVISE → **APPROVED** (3 advisory) | `e4ce7073b`, `9b93af8a0` |
| 040 opt-in produce retry + needles | `921b1b1f-818b-4e79-8236-e65070073c33` | **APPROVED** round 1, 4 advisory | `<retry commit>` |

## The one that cannot be joined, and why it is recorded rather than papered over

`ca5e41122` — the main 343 code commit — **carries no council trailer at all**, and the 098 coverage
report will list it as un-reviewed for ever. That is not a missing review: the plan it ships was
reviewed twice and **APPROVED** at round 2 under `a9b9d4b8-c30f-432a-b35d-5e5864733fc5`, whose verdict
I have read.

What happened: `097` prints `SAVE: SUBMISSION_CORR=<uuid>` **before** it publishes, and my publish
failed (`AlreadyExists` on the epoch-second `kcat` pod name — another session submitted in the same
second). Re-running, I passed the previous correlation **positionally** as
`RESUBMIT_CORR=<uuid>` instead of as an environment variable, so the script took the literal string as
the trail id. The round dispatched and produced a real verdict under that key, but the key is not a
UUID, so the `commit-msg` **trailer gate refused it** — correctly, since a non-UUID join key resolves
to nothing.

I chose not to spend a duplicate paid round buying a bookkeeping key for a review that had already
happened. **`Council-Reviewed:` was never written on an unread verdict**, which is the line that
actually matters.

**Two things for the next session.** `RESUBMIT_CORR` is an ENV VAR (`RESUBMIT_CORR=<uuid> ./097…`) or
a **bare** uuid as arg 2 — never `NAME=value` positionally. And read the **last** line of a `097` run,
not the loudest one: the summary block prints before the publish, so a failed dispatch looks exactly
like a successful one until you check.

## What the reviews actually bought, which is the reason to keep using the gate

Two of the five rounds were REVISE and **both found real defects**:

- **040 round 1, HIGH:** `topicClass`'s `system.*` arm returned its raw input — the exact case the
  plan's own cardinality rule forbids. I had stated the rule and then flagged the exception in the
  *risks* for a reviewer to confirm, rather than closing it. Measured afterwards: **78** distinct label
  values today, growing per agent type. Now a closed compile-time family set.
- **040 round 1, MEDIUM:** `no_leader` collapsed the client-side and broker-side errors, which behave
  **oppositely** inside kafka-go — destroying the one distinction the whole retry analysis rests on.
  Split into `client_no_leader` / `broker_no_leader`.
- **343 round 1, gating:** read the submission's *sketch* as dropping two guards. The guards are in the
  shipped code; the sketch omitted them — so the objection was right about what it was shown. And it
  exposed a **real test gap** I would not have looked for: the two guards cover each other *in series*,
  so either could be deleted with every test still passing. Closed by a new test.

One objection was accepted as fair and **not** acted on: the PodMonitor kustomize wiring is orthogonal
to the kafka mechanism (minimality). It stays because it was an explicit owner decision this session
and forward-only forbids an amend — recorded so a future reader takes it as owner-directed scope.

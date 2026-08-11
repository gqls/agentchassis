# The 090 diagnosis loop — what it is, when to spend one, and how to read what comes back

**A standing explainer, not a milestone summary.** It has no date in its name on
purpose: it describes a mechanism rather than a moment, so it should be *edited*
when the mechanism changes, unlike the dated `SUMMARY_<date>_<slug>.md` series,
where each file is a new one and the series is the record.

Written 2026-08-11 for the owner, after the question "what is a 090 run?".
Everything with a number in it was measured on the live system that day.

---

## 1. What it is, in one paragraph

The platform can diagnose its own bugs. You hand it a **symptom** — one sentence,
no theory — and it dispatches an agent that goes and reads the actual Go code and
queries the live database, gathers evidence, forms a theory, tests the theory
against more evidence, and comes back with a verdict: **CONFIRMED**, **REFUTED**,
or **UNVERIFIABLE**, with file:line citations and real query results attached. It
is not a chatbot answering from memory about the codebase; it is a process that
goes and looks. It takes minutes and it costs credits.

It is named for the script that starts it:

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/090_TRIGGER_needs_diagnosis_v1.sh "<symptom>"
```

## 2. The problem it exists to solve — and the day that proved it

The failure mode it targets is not ignorance. It is **confident wrongness**.

An agent (or a person) investigating a bug forms a theory early, finds evidence
consistent with it, and stops looking. The theory gets written into a handoff in
the same confident voice as the facts around it, and every later thread inherits
it as established. Those are the expensive mistakes, precisely because they arrive
stated with confidence and nobody re-checks them.

The repo's guidance on this used to say the opposite of what it says now. It read:
*"It is NOT a gate and NOT a default … you have full context and will out-diagnose
the loop faster and for free."* That was tested and it failed. On 2026-07-19 a
session with full context filed a confident structural claim about two rerender
paths, built from grep hits whose functions it had never opened. **The loop refuted
it in 9.5 minutes**, by reading the one function the session had skipped, and the
refutation held on re-check.

The lesson written into `CLAUDE.md` from that day is the one worth keeping:

> **Confidence is not a signal.** The wrong claim felt obvious; that is exactly why
> "obvious" cannot be the gate. Full context is no protection, because the failure
> mode is not missing information — it is not looking.

So a **REFUTED verdict is a success, not a waste.** It is the cheapest possible
place to be wrong. One run, versus a wrong root cause that reaches a handoff and
costs every thread that then believes it.

## 3. When to spend one

The test is **not how hard the bug feels**. It is **what you are about to assert**.

**File one before committing to a root cause when the claim is durable:** a
mechanism, a structural property of the platform, a cause that lives outside the
symptom, or a fix that changes behaviour beyond the one site in front of you.
Specifically, always when any of these hold:

- the cause is still non-obvious after a quick look (grep, then read the function);
- you suspect it is cross-cutting, or that the cause is **not** where the symptom
  is — a local fix would then paper over a shared defect;
- you want a cited, auditable diagnosis, because the fix will change behaviour
  fleet-wide or you will be asked to justify it later.

**Debug directly instead when the fix is local and self-evidencing** — you can
watch it fail, change it, watch it pass, and nothing outside that file depends on
your being right.

There is a standing owner ruling (2026-07-31) that a `bugs_open/` file asserting a
**cross-cutting or structural** root cause is not properly "filed" until it has
been through the loop, **or** the filing session says plainly why it substituted
equivalent first-hand verification. The escape hatch is real, but it must be
declared, not silently taken. That ruling came from `bugs_open/155`, which was
filed on genuinely rigorous first-hand work — reproduced, read the code path,
confirmed the fix — and *still* should have run the loop by the section's own
criteria. It was run afterwards: **CONFIRMED on the first iteration**, citing the
same lines.

## 4. How to use it

**Check first whether it is already filed.** The platform's own immune system
sweeps recorded failures fleet-wide and routes genuine platform-wide code bugs into
the same queue automatically, so part of this class arrives without anyone asking:

```sql
SELECT summary, status FROM site_work_items
 WHERE item_type='needs_diagnosis' AND status='awaiting_diagnosis';
```

This is a de-duplication check, not a reason to skip filing — the sweep only sees
failures the platform already *recorded*. A wrong belief you are about to write
into a handoff has no failure row, and no sweep will ever catch it. Also grep
`/bugs_open/` and `/bugs_closed/` for the mechanism first.

**Then write the symptom.** This is the part that decides whether the verdict is
worth anything:

- **State the MECHANISM**, then **POINT** at the tables and symbols where the
  evidence lives.
- **Assert no rows and no counts** — the loop fetches and cites them itself. If you
  assert a number, you have told it what to conclude.
- **No downstream-consequence clauses** ("…which means every site is broken"). They
  go stale and they bias the run.
- **One coherent bug per run.**

The trigger refuses if another thread already has open work on the target
(`FORCE=1` overrides once you have read their findings).

## 5. What comes back, and where it lands

Everything joins on **one correlation id** — the work item, the evidence bundles,
and the verdict:

| where | what it holds | live count (2026-08-11) |
|---|---|---|
| `site_work_items` (`item_type='needs_diagnosis'`) | the durable intake record, queryable long after the pods are reaped | 32 all-time: 25 complete, 5 cancelled, 2 failed |
| `diagnosis_artifacts` (`kind='bundle'`) | the evidence the run actually gathered — code and query results | 435 |
| `doc_notes` (`categories ? 'diagnosis'`) | the terminal verdict, human-readable | 11 `diagnosis`, 5 `unconfirmed-diagnosis`, 1 `corrected-diagnosis` |

**The verdict is a `doc_notes` row, not a `diagnosis_artifacts` row** — there is no
`kind='diagnosis_report'`, and looking for one returns a confident zero. Read it by
correlation, never by `ORDER BY created_at DESC LIMIT 1`: this is a shared table
with several lanes writing to it, so the newest row is very often somebody else's.

Note the three verdict categories are not two. **`unconfirmed-diagnosis` is a third
of everything the loop has ever concluded** — that is the mechanism working, not
failing.

**UNVERIFIABLE does not mean "hard bug".** It usually means the *question* was
wrong — the loop looked where you pointed it and the evidence was not there. The
most common concrete reason on this system is that the rows had already been
deleted: `orchestration_states` keeps `AWAITING_RESPONSES` rows for **4 hours** and
`COMPLETED`/`FAILED` for 24, so a run fired the morning after the failure is
searching an empty table however well it searches.

## 6. What it cannot do, stated plainly

- **It reads the repo at the last PUSHED commit, not your working tree.** It fetches
  a tarball of the remote ref. This tree commits far more often than it pushes —
  measured 2026-08-11, HEAD was **91 commits** ahead of the pushed tip. So it
  cannot see the file you just wrote, and a symbol you added an hour ago may be
  invisible to it in a way that looks like the symbol not existing.
- **It sees the code and the clients database. It does not see the cluster.**
  Kubelet configuration, node filesystem usage, pod scheduling — all invisible.
  This is why `bugs_open/252` (the disk-pressure one) was filed with the escape
  hatch declared: half its evidence came from `kubectl get --raw .../configz` and
  `.../stats/summary`, which the loop cannot reach.
- **It reviews nothing.** It tells you where the cause is; it does not judge a fix
  you have written. That is the **council gate** (`097`), a different mechanism with
  a different trigger, and the two are easy to confuse because they share the
  artifact table. The council reviews the fix you wrote; only the diagnosis loop
  tells you the cause is not where you are looking.
- **It is not a blocking gate.** Nothing stops you committing without one.

## 7. The one-line version

Spend a run **before** you assert a durable cause, not after somebody contradicts
you — and treat a refutation as the run doing its job.

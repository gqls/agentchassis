# NOTES — bugs_open/172 (append-only, newest at the bottom)

## 2026-08-02 — picking it up

Ownership check before starting, because `who-owns.py` reads COMMITS and is blind
to a session mid-fix:

- `who-owns.py 172` says **OWNED or recently active** — but the "owner" is the
  `bugfix_164` lane that *filed* it, and that lane closed 164 on 08-01
  (`SUMMARY_2026-08-01_bundle_body_cap_closed.md`). Not a live claim.
- Grepped all 30 live `.jsonl` transcripts (`-mmin -400`) for `gatherAgentState`,
  `agent_state_cap`, `diagnose_load_runtime`, `bugs_open/172`. Every hit was
  another session's own **ownership survey output** (who-owns dumps, bug-count
  tables) — nobody is in the code.
- `git status`: `diagnose_load_runtime_action.go` is **not dirty**, so no session
  has uncommitted edits to it.
- `site_work_items`: no open item touching this target.

**Trap avoided:** several sessions' transcripts mention `bugs_open/172` dozens of
times, which reads as contention. It is not — it is the *index*, which every
session auto-loads, plus survey output. Counting mentions without reading them
would have made me skip a genuinely unowned bug.

## 2026-08-02 — the measurement that changed the job

The ticket says the cap is latent and asks, first thing, to measure it. Re-measured
today: still **max 4 types listed against a cap of 5**. Latent confirmed.

Then I counted what the section actually renders, which is not the same question.
`log_lines` looked healthy — a solid 10 rows, the cap, in 47 of 72 bundles. It was
`count(DISTINCT agent_type)` inside those lines that broke it open: **23 bundles
named more than one agent type and had rows, and in every single one all the rows
belong to ONE type.**

Root cause is a *second* cap in the same function, three lines of SQL below the one
the ticket names: `WHERE agent_type = ANY(...) ORDER BY created_at DESC LIMIT 10`.
One budget, allocated by global recency, so the chattiest agent takes all of it.

Reproduced against the live table directly rather than trusting the artefact:
`{page-content-writer, council-gate, diagnose-agent}` → 10 rows, **all
council-gate**. `page-content-writer` has 18,286 rows and `diagnose-agent` 324;
both render nothing, and nothing in the bundle says so.

**So the ticket's "latent, one short of firing" is true of the cap it names and
false of the function it lives in.** Recording that plainly because the ticket is
otherwise unusually careful — it was written by a lane that had just been caught by
the same narrowing, and it *still* narrowed. That is the fourth pass over this loop
family narrowing by the shape it happened to grep for.

**A row count cannot see a distribution.** That is the transferable line.

## 2026-08-02 — filed the claim before asserting it

Per CLAUDE.md's default: the starvation claim is structural and durable, so it went
to the diagnosis loop (`090`) *before* I wrote it down as a cause.
`RUN_CORRELATION_ID=13e95253-45aa-440a-826e-7bb6f9e0e5b3`. Fired it in the
background and built the fix while it ran, rather than blocking on it.

## 2026-08-02 — missteps, in the order I made them

1. **Invented a test helper that does not exist.** Wrote `newAgentStateDB` returning
   `*mockDBCloser` — a type I assumed the package had for sqlmock. It does not;
   `sqlmock.New()` returns a plain `*sql.DB`. Caught by the compiler in seconds, so
   it cost nothing, but it is the same reflex that writes a plausible symbol into a
   *document*, where no compiler is watching.

2. **Got the council submission schema wrong.** Sent `plan` as an array of edits;
   the schema is an object `{summary, edits[], grounded_in[], risks}`. The trigger
   refused it client-side (`ERROR: .plan missing`) — the refusal is cheap and
   spends no credits, which is clearly deliberate.

3. **THE REAL ONE — I assumed a passing determinism test meant determinism was
   tested.** I mutated all four halves of the fix to prove the tests bite. Three
   mutations failed correctly. Deleting `ORDER BY type` **passed unchanged**.

   Why: sqlmock replays rows in whatever order the *test* supplies, so a mock-driven
   test structurally **cannot observe the database's ordering**. My test asserted on
   the returned slice, which the mock had already ordered for me. It was asserting
   my own fixture back at myself.

   Had I not run the mutation, I would have shipped a test suite that *looked* like
   it covered the determinism half of the fix and covered none of it — and the
   ticket specifically warns "a test that runs once cannot see this".

   Fixed two ways, because neither alone is honest: a strict `ExpectQuery` on the
   query TEXT (a mock *can* catch the clause being deleted — verified by re-running
   the mutation, which now fails), and the live SQL check in RUNBOOK §3 for the
   ordering guarantee itself, which belongs to Postgres and not to the test.
   The test says this in its own comment rather than reading as full coverage.

   Logged in `WRONG_CALLS.md`.

## 2026-08-02 — shipped to the gate and committed

- Council: `d47b826e-6fc6-42ad-a2ef-62b1f1ba0b88`, submitted before committing.
- Commit `3761a04ca`, pathspec, two files, `Council-Submitted:` trailer (the verdict
  had not landed, and holding code back on a shared HEAD is not available here).
- Deliberately **no shared "cap and report" helper**, though this family now has six
  marker sites that rhyme. A shared mechanism arriving inside a bug patch is what
  `bugs_closed/124` was vetoed for. The two new helpers are file-local and pure.

## 2026-08-02 — the diagnosis loop came back UNVERIFIABLE, and it was right to

`090` run `13e95253-45aa-440a-826e-7bb6f9e0e5b3` finished: **UNVERIFIABLE** — not
CONFIRMED and not REFUTED. Recording it as it reads, because a verdict that does
not go your way is the one worth writing down:

> "this bundle contains neither symbol's body to verify that mechanism, nor any
> instance where a matched type WITH real, non-zero call history was pushed out of
> the render by another type's more recent volume (the only kind of row that would
> demonstrate the mechanism actually firing) … a genuine crowding instance, which
> has not yet been found (only a genuine-zero instance has)."

Two things follow, and they point in opposite directions:

1. **It did not verify my claim, so it is not evidence for it.** What carries the
   claim is first-hand work: the artefact census (23 of 23) and the live
   differential in RUNBOOK §3, where `page-content-writer` holds 18,286 rows and
   renders nothing. That differential **is** the "genuine crowding instance" the
   verdicter says was not found — it had only the run's own matched pair to work
   from (`build-dispatch-loop`, which has zero rows ever, a genuine zero) and
   correctly declined to generalise from it.
2. **The run is itself an instance of the problem.** Its verdicter spent a
   `data_request` on "total calls for build-dispatch-loop" purely to decide whether
   an absence in the section was real. That is exactly the question the fix answers
   in the artefact: post-fix the section states `no llm_call_log rows exist for:
   build-dispatch-loop … (this is an answer, not a cap)`. The loop paid a round trip
   for a distinction the bundle should have handed it.

**Near-miss, caught while reading the run's own bundle.** My census regex
`agent_definitions\[([^\]]+)\]: root ai_service` also matches the **source code of
the function being fixed**, which the bundle quotes verbatim as `agent_definitions
[%s]`. A bundle about *this file* therefore counts its own code excerpt as rendered
evidence. Re-ran the census partitioned on whether the body quotes the source: the
72 historical bundles do **not**, and the 23-of-23 result is unchanged; the 4 that
do are all from today's own run. **The headline survived, but only because I
checked — a measurement whose pattern can match the code it measures will
eventually be run against a corpus containing that code.**

**Two shell missteps in one command, neither costly, both worth the line.** I left a
stray `cat >> <another lane's file>` with no stdin at the head of a compound
command; it blocked on stdin for two minutes and the heredoc behind it never ran.
Then I reached for `pkill -f <pattern>` to clear it and the pattern matched **my own
shell**, killing the session's command with exit 144 — the same "pkill variant"
another lane logged today. Nothing was damaged: `>>` cannot truncate and no bytes
ever arrived, so the other lane's file is byte-for-byte what it was (0 bytes, as it
already was in the session-start `git status`). **Check what the FIRST command in a
compound line reads from, and never pattern-kill from inside the process you are
matching.**

# HANDOFF 2026-08-12 — 040-kafka-dial re-opened a residual, fix council-approved; here's the next pickup

Continues this series (`HANDOFF_2026-08-11_247_closed_next_pickup.md` and
earlier). Standing brief unchanged: `HANDOFF_2026-08-05_next_bug_pickup.md`.

## `bugs_open/040-kafka-dial`: not closed, but advanced — full account in §11 of the bug file

Picked this up because the **entire recent range (`029`–`255`) came back OWNED**
via `who-owns.py` — every single bug from the previous handoff's "check the older
backlog" list was claimed by an active workstream, which the previous handoff's
own point 1 had already warned would happen ("this will already be stale — 35
peer sessions were active concurrently"). Confirmed with `ListAgents`: 35 peer
sessions, mostly idle rather than busy, so the "OWNED" signal is real workstream
attribution (mentions in HANDOFF docs), not live busy-ness — still correctly
read as "don't compete here."

Found one exception by reading the who-owns output carefully rather than
trusting the verdict line alone: **040 is one of this repo's known number
collisions** (two unrelated bugs share it). `who-owns.py 040` matched the
*other*, closed 040 (`040-partial-build`) — a false positive. The real
`040-kafka-dial` workstream directory had sat untouched for 17 days, past its
own 7-day close-condition checkpoint, genuinely unowned.

**What happened:** re-ran the week-old close condition live. It fails (32
`timeout`/7d, non-zero) and a second residual — 71,832 `refused`/17d, two
orders of magnitude bigger than everything else combined, entirely invisible to
`agent_error_log` — was found, time-bisected, and had broker restarts ruled out
before any code theory. Filed to the `090` diagnosis loop first (verdict
**UNVERIFIABLE** — its tools can't reach Prometheus or vendored kafka-go
internals, not a refutation). Shipped one narrow, unambiguously-safe hardening
regardless (`getController` now rejects an empty controller `Host` instead of
building a dial target that resolves to loopback) rather than wait on a
confirmation the loop structurally cannot supply. **Council: APPROVED**, round 1,
2 advisory objections, none high-severity (`af5f74bc-5e6c-4a6c-a3fc-7ac27eab4b6f`).
One of those objections (`editquality`, surfacing `bugs_open/240`'s
blank-`MetadataTopics` landmine as an uncovered mechanism) led to a timing
cross-check written into both bug files: it explains at most the smaller of the
burst's two episodes, not the majority.

**Commits:** `e1f960ac2` (the fix + test), `8eef55748` (bug file + workstream
docs), `4945da743` (council verdict + cross-bug contribution to `bugs_open/240`),
`84ce1dc66` (NOTES).

**Stays OPEN.** Root cause not established — this is explicit in every doc, not
a hedge. The `refused` burst has not recurred in the ~24h since; if it does, the
new `Warn` log this session added will say whether `getController` is the site.

Full account: `bugs_open/040_HANDOFF_2026-07-20_kafka_dial_timeouts_fleetwide_intermittent.md`
§11, workstream docs at
`docs/agent_docs/docs024_key_docs_latest/bugfix_040_kafka_dial/` (all five
current, plus a new `SUMMARY_2026-08-12`).

## Next pickup, mechanically

1. **Do not re-check the 029–255 range** — it was fully saturated as of
   2026-08-12 and will only have gotten more so. Check the **older backlog**
   this session did not have time for: anything below 029 (there is a header
   `HANDOFF_2026-07-31_checker_layer_remaining_items.md` file with no number,
   worth reading), and re-verify any number already reported OPEN in an earlier
   handoff in this series in case it was quietly picked up and finished without
   moving to `bugs_closed/` (per owner ruling, finished bugs stay in
   `bugs_open/` with closure evidence inline — check the file's own text for a
   closure/OUTCOME section, don't infer from directory alone).
2. **Number collisions are a real seam for finding unowned work** — `040` was
   found exactly this way: `who-owns.py` matched the wrong twin. The known
   collision list as of this session: `016`, `017`, `040`, `083`, `112`, `131`,
   `146` (grows). If a `who-owns.py` OWNED verdict looks weak (one workstream,
   few mentions, or the matched workstream's name doesn't semantically fit the
   bug's own title), open the bug file itself and check whether it names a
   *different* slug for a number collision before accepting the verdict.
3. Standing four before touching anything: `scripts/who-owns.py <n|slug>`, `git
   log` the file the bug's own text names, grep live `.jsonl` transcripts,
   `site_work_items` queue check. This session also found `ListAgents` useful
   as a live cross-check against `who-owns.py`'s commit-based (and therefore
   laggy) signal.
4. Re-verify the defect against the live system before planning — this session's
   whole pickup was exactly that: a bug marked "awaiting 7 days of metric" whose
   7 days had actually already passed, unread.
5. Then the brief as written: fable for the plan, opus to implement, build+test,
   council gate for `platform/`/`internal/`/`pkg/`, commit per task with an
   explicit pathspec, keep the five standing docs current, missteps to
   `WRONG_CALLS.md`.

## Method lessons from this arc

1. **A `090` UNVERIFIABLE is a statement about the loop's REACH, not a
   refutation.** The loop's evidence tiers are `agent_error_log` (SQL) and the
   static code index — it cannot query Prometheus and cannot see vendored
   third-party library internals. A finding that lives entirely in a metric
   (this one did) is structurally outside what it can independently confirm.
   Don't read UNVERIFIABLE as "go find a different bug" when the underlying
   runtime evidence (gathered first-hand, with controls) is solid; do read it as
   "you must now decide whether to ship on partial confirmation, and say so
   plainly if you do."
2. **A council submission is itself a second-opinion mechanism, not just a
   gate.** The `editquality` seat surfaced a directly relevant landmine
   (`bugs_open/240`'s `MetadataTopics`) that this session's own reading had
   missed entirely, despite reading `dialer.go` closely. Read every advisory
   objection even after APPROVED — round-1 approval does not mean the
   objections were wrong, only that none was severity-gating.
3. **`count by (...)` on a Prometheus counter counts SERIES, not events —
   `sum by (...)` is the total.** Cost 340× on the first pass this session
   (212 vs. the real 71,832). Cross-check any metric breakdown with the other
   aggregator once before citing it. Logged in `WRONG_CALLS.md`.
4. **`max_over_time` bisection finds WHEN a burst happened on ephemeral-pod
   counters, where `increase()`/`rate()` cannot** (an already-known trap in this
   same bug file, §10) — widen the window from now in fixed steps and read off
   where the cumulative total jumps. `offset` past the series' actual history
   returns EMPTY, not `0` — cross-check against a control label before trusting
   an empty result as "no events".
5. **Rule out the boring explanation with a live check before reading code.**
   Confirming broker pods had zero recent restarts, in one `kubectl get pods`
   call, closed off "the broker bounced" before any time was spent on dialer
   internals — cheap, and it kept the code-reading honest about what it was
   and wasn't explaining.

# HANDOFF 2026-08-12c — 040-kafka-dial: both follow-ups closed by hand; the only
# remaining step is time-gated. Pick up backlog work, or re-read the metric later.

Continues `HANDOFF_2026-08-12_040_council_approved_next_pickup.md` →
`HANDOFF_2026-08-12b_040_build_verified_diagnosis_rerunning.md` → this file.
Read those two first for the full arc; this is the narrow "what changed since,
and what to do next" continuation. Written at a genuine pause point, not a
crisis: nothing is broken or blocked-on-a-decision, the next useful action on
040 just isn't available for ~24h.

## State, precisely

1. **The fix is committed, council-APPROVED, and PROVEN LIVE across two
   further chassis rolls** since the last handoff in this series
   (`v1.0.1291` → `v1.0.1293`, most recently git commit `7a1887e31`, both
   pods restarted `2026-08-12T19:13Z`). `e1f960ac2` (the fix) is confirmed an
   ancestor of both. Nothing to do here; just keep confirming it stays an
   ancestor on future rolls (`git merge-base --is-ancestor e1f960ac2 <new
   commit>`).

2. **The re-fired follow-up diagnosis run (`58a0390c-33ec-4580-9697-3320b280475d`)
   COMPLETED — with outcome `UNVERIFIABLE`, not a hit.** Its own
   `needed_evidence` named the exact gap (a capped `fmt.Sprintf` content
   search that never reached `spawn_actions.go`). Rather than fire a third
   `090` run for what turned out to be two greps and two function reads,
   answered it first-hand: `getController`/`controllerAddress` is the only
   fleet-wide site building a dial target from a live kafka-go
   `Controller`/`Brokers` response; the `:9092` filters elsewhere
   (`CreateJobTopic`, `spawn_actions.go`'s `getConfiguredKafkaBrokers`) guard
   statically-configured strings, not live metadata, so they're an unrelated
   mechanism, not a sibling instance. **This closes both council objections**
   (`guardian`'s caller enumeration confirmed exactly: `CreateTopic`,
   `DeleteTopic`, `ListTopics`, `TopicExists`, no fifth; `prior_art_librarian`'s
   reuse-vs-duplicate question dissolves — nothing of the same kind to reuse).
   Written up in `bugs_open/040` §11.7. Commit `2dedcf1f7`.

3. **First post-fix read of the `refused` counter is silent, but — checked
   carefully — NOT YET INFORMATIVE.** Bisecting
   `sum(max_over_time(ai_persona_kafka_dial_total{outcome="refused"}[Xh]))`
   shows genuinely zero `refused` events for ~26.5h (cross-checked against
   live `ok`/`timeout` series so the zero isn't a blind query). **But** the
   `v1.0.1291` replicaset's creation timestamp
   (`2026-08-12T14:55:10Z`, `kubectl get rs -l app=agent-chassis`, old RS
   still present scaled to 0) shows the fix has only actually been *running*
   for the last ~4h24m of that 26.5h silent window — the other ~22h of quiet
   happened on the **unfixed** binary, before the fix ever shipped. **Do not
   read this as the fix working.** The metric was already capable of a
   day-long quiet stretch with the bug present (episodes 1→2 were ~12h apart,
   themselves too small a sample to call a rate). Written up in full,
   including the explicit disconfirming-shape reasoning, in `bugs_open/040`
   §11.8. Commit `e1d905bb7`.

## What to do next

**Two independent threads. Pick whichever fits the session.**

### A. Re-read the `refused` counter once it means something (time-gated, not urgent)

Not before **2026-08-13 ~15:00Z** (24h into the fix's actual runtime) — and
even then, treat it as one more data point, not a verdict, since two prior
episodes don't establish a rate. Query:
```
sum(max_over_time(ai_persona_kafka_dial_total{outcome="refused"}[Xh]))
```
bisected as in §11.3/§11.8. **Before trusting a zero, cross-check
`outcome="ok"`/`"timeout"` are nonzero over the same window** (the runbook's
own documented trap — an empty result can mean "no samples" or "broken
query" and looks identical). If it's still 71,832 flat (no increase) across a
span meaningfully longer than the ~12h inter-episode gap already observed,
that starts to be real evidence, not just quiet timing — write it into
`bugs_open/040` as §11.9, and reconsider whether 040 can move toward closure.
Re-verify the fix is still live first (`git merge-base --is-ancestor
e1f960ac2 <current chassis commit>` — check `kubectl -n ai-persona-system
logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'`, or the
`get rs` timestamp trick above if the log's scrolled out of range).

### B. Move to the next item in the bug backlog

040's own next action is time-gated (above), so this is the natural place to
resume general backlog clearing. **Standing four before touching anything**
(as in every prior handoff in this series):
- `scripts/who-owns.py <number|slug>` on whatever you're about to pick up.
- `git log` on the bug's own named file (not just the bare number — several
  numbers are shared between two unrelated cases).
- Live `.jsonl` transcript grep for another session already mid-fix
  (ownership checks lag uncommitted work).
- `site_work_items` queue check for existing open work on the same target.

The `029`–`255` range was reported fully saturated as of
`HANDOFF_2026-08-12b`; that is only more true now. Check `/bugs_open/` fresh
rather than trusting that claim unverified — it will have moved.

## Files touched this session (all committed, pathspec per task)

- `bugs_open/040_HANDOFF_2026-07-20_kafka_dial_timeouts_fleetwide_intermittent.md`
  — §11.7 (isolation question closed by hand), §11.8 (first post-fix
  `refused` read, with the time-gating caveat).
- `docs/agent_docs/docs024_key_docs_latest/bugfix_040_kafka_dial/NOTES_040_kafka_dial.md`,
  `README_where_we_are.md` — both updated in step with the bug file, twice
  this session.
- `docs/agent_docs/docs024_key_docs_latest/bug_backlog_clearing/` — this
  file.

No bugs_closed move — 040 stays open (closure evidence lives inline in
`bugs_open/`, per CLAUDE.md's 2026-08-06 override of the fixed-AND-live bar,
and here the fix isn't even confirmed-as-cause yet, only proven live and
isolated).

The council gate is live and commissioned. Applied to clients_db on postgres-clients-0 in ai-persona-system with your named go, verified, and proven end to end on a real submission — and your warning about the
growing roster was load-bearing.

The apply. Before touching anything I re-read the live fix-proposer row rather than trusting the v11 files — and found the council had already grown to seven seats (an adoption-pipeline guardian had landed since).
I mirrored the seventh seat and its gate/footprint verbatim from the live row, went to seven-way review and check arrays, and recorded two standing rules in the seed header: always diff against the live row before
any re-apply, and code_lookup (the fix loop's code-question answerer for its blind reproposer) is a deliberate non-mirror, since gate authors read objections themselves. Applied cleanly; post-apply verification all
green: 27 steps, seven prompts intact, five gated footprints.

The smoke — a pass, with a real catch. I submitted a genuine change (the digest gate-verdicts section the handoff names as the channel to extend). In ~100 seconds: five seats reviewed, guidelines and adoption were
correctly skipped by the relevance filter with no LLM spend, and the abstention path held in the deterministic decision. Verdict: REVISE, on four correct, fixable objections from the edit-quality seat. Best of all,
the reviewers' verification checks — running against live data on their first real outing — refuted the plan's core assumption: every council_report row in the database has source_agent='generic', so my proposed
filter would have produced a permanently empty digest section with no error, exactly the silent-blank family the bug-historian watches for. That change now sits on its revise trail (bd12762a) with the corrected
partition key as revision item one, and the source_agent fact is cross-flagged to the fixloop thread.

How threads use it from now: write a submission JSON, run 097_TRIGGER_council_review_v1.sh, and on approval commit with the Council-Reviewed: <correlation> trailer; 098_REPORT_unreviewed_commits_v1.sh measures
fleet coverage against the 28-commits-in-3-days baseline. Everything is committed narrowly (9049ec53a, 6191fdbf1, e289ed9af), notes and memory are current, and PR-mode enforcement remains untouched behind your
explicit go.

✻ Worked for 10m 45s · 1 shell still running

● Background command "Poll gate orchestration until terminal" completed (exit code 0)

● That's the original background watcher winding down — its polling loop observed the same terminal state I already reported (complete_revise / COMPLETED at 20:37:11). Nothing new in it; the smoke result, verdict,
and all records stand as summarised: the gate is live, commissioned, and its first real verdict (REVISE, with the source_agent='generic' catch) is on the trail at correlation bd12762a.

---

Both documents are written, and the gate now runs the full council. Three things done:

The council had indeed grown — to nine seats. Two new guardians had landed overnight (one watching the diagnosis machinery's honesty gates, one watching the improvement loop's termination limits — the guards that
stop the audit loop running unbounded, as it once did for ten days). I read them from the live database rather than trusting our files, mirrored both verbatim into the gate with their relevance footprints,
re-applied, and verified: 31 steps, nine reviewers, seven gated footprints, with the previous version snapshotted for rollback.

The how-to-use note went into CLAUDE.md — which is the right home, and worth saying why: every session loads that file at startup, so a note there reaches threads that will never open the fixloop directory, whereas
a note in the runbook only reaches people already looking for it. I kept it to the essentials (how to submit, what each verdict means, the Council-Reviewed trailer, the fact that cost is relevance-gated so
submitting is cheaper than it looks, the coverage report, and the patch-both-councils rule) and pointed at the runbook for the submission schema and detail.

The read-aloud summary is SUMMARY_council_gate_2026-07-18.md, superseding yesterday's, which was written before the gate went live and now reads as stale. It covers what we set out to achieve, what we did, where we
are, and where we're going — including the first real catch, where the reviewers' own database queries disproved an assumption in my plan that would have shipped a permanently empty digest section with no error.

One honest note for the read-out: the coverage number is still zero out of twenty-eight. The gate works, but nothing has been through it except our own smoke test — adoption by other threads is the real test, and
the CLAUDE.md note is the main lever we have on that. Everything is committed narrowly (0351e193e, b2eb452ea), and enforcement mode remains untouched behind your decision.

---

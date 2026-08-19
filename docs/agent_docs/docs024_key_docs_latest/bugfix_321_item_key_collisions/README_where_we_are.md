# Where we are — bug 321 (most tool suggestions silently thrown away)

Append-only, newest at the bottom. Plain prose for the owner.

## 2026-08-19 — what the bug is, and what happened today

When the system asks an AI "what interactive tools would help this website?", it gets
back several good suggestions — and then quietly threw most of them away. The
bookkeeping label used to record each suggestion as a work item was built from the
website's name only, and the work queue refuses two open items with the same label. So
the first suggestion got a work item and the rest bounced off it, invisibly: the run
still reported success. Measured before today's fix: 40 suggestions across eight runs
became 11 work items — about 72% lost. This morning's run alone: 7 suggestions, 1 item.

The fix was already sitting in the code, unused: an optional setting that adds the
tool's own name into the label, so every suggestion gets its own. I checked the
system's whole history first — every one of the 239 suggestions ever made carries a
usable name, and no answer ever repeats one — then switched that setting on for the
two suggestion steps, plus two other agents with the identical flaw that simply
haven't fired yet (one of them is about to be brought to life by another session's
fix, so it would have started losing two-thirds of its output on day one). I also made
the suggestion loop tolerant of a single bad suggestion, so one malformed answer
costs one item, loudly recorded, rather than the whole batch.

Because this same "label too coarse, work silently vanishes" mistake is one we have
written up before and it still happened again, I also built the check that makes it
mechanical: a small daily job that scans every agent definition for this exact shape
and reports every morning — including an explicit "all clean" note, so silence means
the job is broken, never that all is well. It found the fleet clean after the fix; when
I deliberately fed it yesterday's configuration it flagged all the right steps and
nothing else. Your evening release picked the new job up automatically.

Decisions you made today: no throttle on how many suggested tools get built per run
(the caps already in place bound it at 8); fix all four affected steps, not just the
two named in the bug; and add both the daily check and a small runtime warning.

Still to come: the runtime warning (a one-line note in the logs whenever a work item
gets deduplicated away inside a loop — the net under the daily check's one blind
spot), which needs a code review round and rides the next release; and the proof at
the artefact — the next suggestion run must produce one work item per suggestion,
where yesterday it produced one in seven. Worth knowing before you look at the first
post-fix run: fixing this means suggested tools actually get built now, so build
volume per run will genuinely rise — that is the pipeline working as intended, and
I'll report the real numbers so you can throttle on evidence if you want to.

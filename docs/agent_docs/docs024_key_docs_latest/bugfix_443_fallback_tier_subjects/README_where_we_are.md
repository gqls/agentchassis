# Where we are — bug 443 (repeated sections on plan-less sites)

*Owner's plain-prose log. Append-only, newest at the bottom.*

## 2026-09-02 — picked up, measured, plan written

The bug: on six of our real sites, pages sometimes show the same section heading two or three
times. The finetuning.uk booking page was the example — "What you do in the hour" three times.
It is not the writer having a bad day. Those six sites are the ones that never got an entry in
the newer site-planning tables, and the mechanism that gives each section its own one-line
topic ("subject") only works for sites that are in those tables. So on these sites, when a
page's layout uses the same building block twice, the writer is handed the identical
instructions twice, and writes the same thing twice.

What we checked today: every one of the eleven pages our database flagged as "layout repeats a
block" really does show repeated headings on the live site — eleven out of eleven, most of
them word-for-word identical. So the damage is real but contained: eleven pages, three sites.

The plan, in plain terms: teach the fallback path (the one these six sites use) to carry
per-section topics too, by storing the topics right next to the page's own section list — same
row, same shape — so they can never drift out of step with it. Nothing changes for any page
until someone actually writes topics for it; a page without them behaves exactly as today. We
also add a quiet alarm: whenever a page gets built with a repeated block and no topics, a note
is recorded, so this stops being invisible.

One dependency worth knowing: the final step that makes the writer actually USE the topic is a
prompt change that is waiting on your read (seed 641). Until that lands, topics travel all the
way to the writer and stop there — for the planned sites exactly as for these. So the full
"headings become distinct" proof comes after that lands.

Bigger question we are NOT deciding here, filed for you as RFC_063: more and more capability
(topics, fact-scoping, hero imagery) only works for sites in the planning tables, and six real
sites aren't in them. Do we bring those sites into the tables properly (bigger job, fixes the
class), or keep adding fallback support capability by capability? The fix above is correct
under either answer.

## 2026-09-02, end of day — fix approved, committed, columns live; the writer step waits on you

The review council approved the plan first time (eleven advisory notes, none blocking — each
answered with a measurement, recorded in NOTES). The code is committed and will ride the next
chassis build; the two new database columns are already live (safe first — the code that reads
them tolerates their absence and vice versa).

Concretely, after the next roll: a page on one of the six table-less sites can carry one line
of "what THIS section is about" per section, stored right beside its section list so the two
cannot drift apart; and whenever any page anywhere builds a repeated section type with no such
line, a note is recorded instead of the problem passing silently.

What still stands between the finetuning pages and visibly distinct headings: seed 641, the
writer-prompt change that is waiting on your read. Until it lands, the topics reach the writer
and stop there — true for every site, planned or not.

The bigger question is now written up for you as RFC_063: do we bring the six older sites into
the planning tables properly (fixes hero imagery for them too; suggest trying it on the
smallest site first), or keep adding per-capability fallbacks like today's? Today's fix is
correct either way, so there is no urgency — but the imagery lane is blocked on exactly this
for those sites.

## 2026-09-03 — the new build is live and the fix is in it; the quiet alarm caught its first four pages within two hours

Verified directly against the running service (not the deploy log): the fix shipped. The
per-section topic plumbing now works end to end for the six older sites, and the recorded-note
alarm is already doing its job — within two hours it flagged four pages on three OTHER sites
(leopardess, seotools, vetcomparison) building repeated sections with nothing to tell them
apart. Those are sites that DO have planning-table entries, so the problem is broader than the
six sites we started with; the next session will check whether those pages actually look wrong
to a visitor before treating them as new damage.

What's left before we can call this bug closed, in order: the finetuning lane writes the
topics for their four pages and rebuilds (they have the green light as of this morning); your
rewritten writer-prompt (the positive-prompting redraft you asked for) lands via the apis.uk
lane once you pick a framing; then we rebuild the test page and check the headings really
differ; then the same treatment for the seven pages on gaswholesalers and
ai-agent-orchestration. The full checklist is in HANDOFF_2026-09-03_continue_here.md.

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

## 2026-09-03 afternoon — the alarm's first catches are all explained, and they widen the clean-up list, not the bug

The quiet alarm we switched on with the fix caught seven events in its first two hours, and
they were a surprise at first: none of them were on the six unplanned sites. They were on
sites that DO have a plan. Today I chased all of them down, and the answer is reassuring —
there is no new defect anywhere.

What happened is simply that those sites' plans were written before Tuesday's new planning
rule (the one that makes the planner give each repeated block its own topic) existed. The
closest case was seotools.co.uk, whose plan was written 34 minutes before the rule went in —
bad luck, nothing more. Old plans have no topics, so any page rebuilt from one still gets
the repeated-heading problem, and the alarm rightly says so. The rule itself is working:
the only plans written since it landed (gamedesign.uk, two of them) carry topics properly.

I checked the live pages too. Four more pages really do show the repeated headings to
visitors today (three on seotools, one on vetcomparison) — same disease, so they join the
clean-up list for after your prompt read lands. One page that looked alarming
(leopardessconsulting) turns out never to have been published at all: its rebuild keeps
failing for an unrelated reason, which that site's own lane should look at.

I also sized how much of this is still out there: only six pages fleet-wide sit in old plans
with the dangerous repeated-block shape (one of them is apis.uk's own home page — I've left
their team a note), and six more on planned sites fall through to the fallback path we just
fixed. Another lane (dartsonline) independently found that most article-type pages across
the estate use that same fallback path — which means the fix we shipped covers far more
pages than the eleven we started with. Good news dressed as bad.

Where this leaves us, unchanged from yesterday: everything now waits on your read of the
writer prompt (seed 641 — it was redrafted to your wording and approved by the council
today; your read of the exact text is the one remaining gate). After that: the
before/after proof on one page, then the clean-up of the wider list above.

## 2026-09-03 evening — the writer prompt landed, and it half-worked

Another lane test-ran the fixed writer on the technical-details page tonight. Good news first:
the six section headings are now genuinely different from each other — that part of the fix
works exactly as intended. But when I read the actual page and the actual text the writer
produced, the three body paragraphs still open by saying almost the same thing: all three
start by explaining that the model is "small and open-weight" and what that means, even
though only one of them was supposed to be about that. Different headline, same paragraph
underneath.

I dug into why, rather than taking the other lane's word for it, and found the actual cause.
It's a second problem, separate from the one we fixed, and it was hiding behind the first one
the whole time. Every section's writing instructions include, in full, the entire page's brief
— all six sections' worth of guidance, not just the one the writer is meant to be writing right
now. Nothing in the instructions tells the writer "ignore the other five, you're only doing
this one." So when three sections are the same type, given the same full brief, the writer
naturally gravitates to the same part of it — usually whichever bit comes first. This is not
specific to this one page: the same shape would show up on any page whose brief is written
with that much per-section detail.

So the picture is: the fix we built genuinely works — the topics are travelling all the way to
the writer correctly. It's just that a second, older habit in how we write to the writer was
masking that the topics weren't the only thing steering the words. Fixing it is a change to
the shared writer instructions, not to anything we built — the lane that owns writer prompts
has the evidence and a proposed fix already. Until that lands, I'm treating this bug as not yet
closeable: distinct headings alone isn't the bar any more, matching content underneath counts
as the same bug.

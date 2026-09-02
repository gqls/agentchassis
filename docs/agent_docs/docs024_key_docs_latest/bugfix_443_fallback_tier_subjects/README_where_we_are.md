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

# SUMMARY — 2026-08-20b — we were blaming the wrong thing for three days, and the correction is worth more than the fix we thought we had

**Written because this morning's summary (`SUMMARY_2026-08-20_required_fields_router.md`) had its
headline claim retracted four hours later.** That file stays as written, with a correction box — it
is the record of what we believed at the time. This one is the corrected read-out, and it is short
because the position is simpler than the day was.

## What we are trying to do

When the system inspects a site it has built, it files a note about anything wrong — a missing
field, a placeholder left in, formatting characters showing through. Those notes must reach
something that can fix them. This lane exists because for a whole family of them, nothing could:
the note was filed, correctly labelled, and then sat there for ever. The aim is that a page which
is wrong gets **repaired**, and that we can point at one and say so.

## Where we have come from

We built the missing piece — a router that reads each finding and sends it to the right repairer.
It went live, drained its backlog and passed review. Then it exposed the problem underneath: a large
group of findings sit on pages we do not own, which our general repairer refuses. For three days —
across this lane, the neighbouring one, and the closed bug before it — the explanation everyone
worked from was **"we cannot repair these because we do not own them."** Every fix anyone proposed,
including two of mine this week, was a scheme to get around that refusal.

## What we have done

**Established that the explanation was wrong.** These pages cannot be repaired, but ownership is not
why. Their words exist only in the *finished* page; the source data our repairs work from holds just
a fingerprint and a note of where the page was copied from. So there is nothing for any repair to
rewrite. I proved it rather than arguing it: I took the real page template and the real source data
and rendered it the way the live system does — **it produces a page with zero words, and reports no
error** — then ran the same test on a page whose source data *does* hold its words, which came out
complete.

**And the two things turn out to be the same 100 pages.** Of 115 pages using this component, 100
cannot be regenerated, and those 100 are exactly the ones we do not own. That coincidence is what
made a wrong explanation survive three days and three separate attempts: the guard was taking the
blame for a problem it was not causing, and getting around it would have repaired nothing.

**We also fixed something that was quietly misdirecting people.** When a job escalates to a person,
the system attaches a note saying which team should pick it up. All three entries in that list were
wrong — two named bug files by folders they had since moved out of, the third said "nobody owns
this" about a problem with an owner. That note is stamped on once and never revisited, so each wrong
entry is a wrong instruction handed to one person, permanently. **Fixed, and confirmed working in
production at lunchtime**: the day's escalation went out carrying the correct pointer.

## Where we are now

The main bug is where it was, for a reason we now understand properly. Nothing is repaired. The
larger hole — 27 jobs whose pages have no content at all — is untouched by any of this and remains
the bigger piece of work.

The seven jobs this lane tracked closed themselves this morning: their group's success rate crossed
a threshold, they were released, tried, refused, and marked "will not fix" within three minutes.
That is worth knowing about the design — **a job that can only be refused reaches its dead end
faster than it reaches a person, and the quick path is the silent one.**

I corrected myself three times today and each correction is recorded where the claim was made. The
most useful is the last: the first two fixes replaced a wrong answer with a better answer, and the
third replaced a *right* answer that I had applied to the wrong situation. So instead of an answer,
the system now carries the **question** — the next person is told to check whether the repair can
reach their particular page. An answer is only right for the case its author was looking at.

## Where we are going

**One decision is yours, and it is the only thing blocking this lane.** Giving these pages a repair
route means building something that edits finished HTML directly — a shape this estate does not
have. The defect it would fix is seven instances of the mildest kind: backticks around code words
like `fetch()` on developer-tool pages, which many readers would take as deliberate. **It is a real
defect and a small one, and the build is not small.** Neither answer is obviously right, which is
why I have not started it.

After that, the honest remaining work is the one we have been circling: the 27 jobs whose pages have
no content, which needs a different repairer entirely.

Two smaller debts stay open, both named in the handoff: telling a review seat that one of its
standing checks is inverted, and a change to the dispatch rules that genuinely alters behaviour and
so goes past you rather than in on our own judgement.

Full state, with every correction box: `HANDOFF_2026-08-20_continue_here.md` beside this file.

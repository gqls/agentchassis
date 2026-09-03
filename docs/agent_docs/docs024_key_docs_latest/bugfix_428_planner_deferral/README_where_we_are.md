README — where we are, bugfix_428_planner_deferral (plain prose, append-only, newest at the bottom)

2026-09-02. You asked me to look at bug 428 after I'd finished the fight-calendar
bug. Someone else (another session) had already worked out what was going on: the
AI that plans a new website's pages sees the client's own strategy document say
"you should build a page for each fighter" and "a page for each event", and about
three times out of four it just... doesn't. Not a glitch — it reads the suggestion,
decides to keep the launch site simpler, and writes down a one-line explanation
saying so. It's doing exactly what its instructions allow it to do.

That other session also found that the platform already independently noticed
this was happening — thirteen times, across different client sites — and filed it
each time as a flagged item. But nothing ever acted on those flags. Their proposed
fix was: build something that automatically acts on those thirteen flags.

Before building that, I checked why nothing already acts on them. And it turns
out that's deliberate, and quite recent. About a week and a half ago, a similar
"automatically act on the AI's flagged opinions" mechanism accidentally destroyed
real content on a live client site — it rewrote a homepage section and left five
broken images on it, and the original text was gone for good. You (or whoever
holds that role) ruled at the time that the AI's opinions should still be
recorded, but nothing should act on them without a person looking first. That's
exactly the switch sitting in front of these thirteen flagged items.

So building an auto-fix here would have quietly undone that safety decision, for
the exact kind of finding it was put in to stop. I didn't think that was mine (or
the other session's) call to make alone, so I asked you. You agreed: fix the two
harmless parts now, and instead of an automatic fix, build a proper "let a human
look and then release it" button — which is what the safety mechanism was always
supposed to have, and never got.

Done: I tidied up the messy raw data the AI planner was shown (it was reading it
fine either way, but no human could easily audit it before); I added a rule
requiring the planner to name, specifically, any recommended page type it decides
to skip, rather than one vague sentence covering everything it dropped; and I
built the review-and-release button in the admin tool, so a person can now
actually find these flagged items (there was no way to even filter for them
before) and choose, one at a time, to act on one.

All three changes are live or on the shared codebase, and sitting in the
platform's automatic review queue — I'll know within the hour if anything comes
back with concerns.

---

2026-09-03 update. The review queue came back clean on both the backend release
tool and the prompt fix — approved, no concerns. The backend has been live for a
while. The missing piece was the admin webpage itself: it hadn't been rebuilt and
redeployed since before this ticket started, so the button existed in the code
but nobody could actually see or click it.

A later session (continuing this same thread) built and checked that webpage
update, then had to stop and ask you to approve actually publishing it, since
that's a real production action. It's live now — I independently confirmed the
running page really does contain the new button, not just that the deploy
reported success.

So: the release surface is fully live, front and back. Nobody has actually used
it on a real flagged item yet — that's still a genuine human decision, not a
code task, same as before.

---

2026-09-03. A new session picked this back up. The short version: the fix we shipped last time
worked, and working is exactly how we found the next problem.

Last time we told the site planner that if it decides not to build a page type the strategy asked
for, it has to name that type and give a real reason. It now does that, every time. But nobody had
ever checked whether the reasons were TRUE — and it turned out our system had no way to. The
planner writes its reasoning into a notes field that literally nothing in the code ever reads. So a
good reason and a made-up one were the same thing to us.

The made-up one duly turned up. On one site the planner said the articles would be "satisfied by the
blog infrastructure". There is a piece of software by that description, it is real, and it stopped
running on 24 April and nobody noticed for four months. So the site shipped with no articles and a
tidy explanation.

What I have built is a check that does the reading. After the planner finishes, the system now
compares what the strategy asked for against what the plan actually contains, and writes down
anything missing along with which step lost it.

The part I did not expect, and the reason this took the shape it did: another session found that
sometimes the planner does its job and WE throw the work away. On one site the planner produced nine
pages including five real articles, and our own validation step silently deleted five of them and
kept four. No error, no warning, nothing recorded. It had been doing that since May, and it was
invisible because on an established site the deleted pages get restored a moment later — only a
brand-new site shows the loss.

That changed my design. The check now looks at the pages three times — what the planner proposed,
what survived our processing, and what came out the end — so "the planner didn't want it", "we
deleted it" and "we deliberately held it back" are three different, separately recorded things
instead of one shared silence. Before this, all three looked identical from the outside, which is
why nobody caught any of them.

One thing I was careful about: this does not tell the planner off for every decision. If it declines
a page type and points at some other mechanism, the system now checks whether that mechanism is
actually running. If it is, that is a sound decision and nothing is flagged. Only a promise pointing
at something that has stopped gets recorded. A warning system that cries wolf is one people learn to
ignore.

There is also a small honesty fix to the admin page. The "Review & Release" button was being offered
on every flagged item, but for the new kind of item there is nothing to release — they are notes for
a person, not repairs waiting for a yes. Clicking would have failed with a confusing error. It now
says "record only — no automatic route" instead. That was only visible because I built the page and
looked at it rather than reasoning about the code.

Nothing here is live yet: it needs the next routine rebuild of the system, which happens on its own
schedule. The database part is written but deliberately NOT switched on, because switching it on
early would ask the planner to fill in a field that nothing reads — which is the exact mistake this
whole ticket is about.

One decision I need from you when you have a moment. Somebody passed on your ruling that guides
should be their own page type. I have written it down and NOT started it, because there are really
two jobs in there: adding the type (cheap, changes nothing on its own) and re-labelling the 167
guide pages already live across 20 sites (not cheap — it changes what every blog and guide listing
on those sites resolves to). The second one wants a proper review. Tell me whether you want the
cheap half now or both together, and I will sequence it.

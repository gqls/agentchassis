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

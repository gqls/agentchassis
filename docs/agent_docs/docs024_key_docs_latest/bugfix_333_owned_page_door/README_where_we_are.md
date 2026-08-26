# Where we are — 333, the owned-page door

## 2026-08-24, opening

Some of our pages are marked "owned". That means the page belongs to a tool or a widget, and the general-purpose
page writer is not allowed to rewrite it — if it did, it would replace a working calculator with an essay about
calculators. We enforce that: the page writer checks the mark and refuses.

The problem is what happens upstream of the refusal. About two dozen places in the code notice a defect on a page
— a broken link, some raw markdown showing through, a placeholder phone number — and hand it to the general page
writer. **None of them checks whether the page is owned first.** So a genuine defect on an owned page gets sent to
the one worker who is forbidden to touch it. He refuses, and the item is filed as "we decided not to fix this".
Nobody decided that. The defect is still there, and the next sweep finds it again and sends it to the same dead end.

It has been happening for over four months, and it is happening today: 83 findings have ended that way since
19 August, five of them today, and the most recent arrived at 11:48 this morning.

**What we are doing about it.** There is one place nearly all of these findings pass through on their way into the
queue — a shared function that actually writes the row. We are putting the check there, once, instead of in
twenty-six places. It asks two questions: does this worker refuse owned pages, and is this page owned? If both,
the finding is not sent to the dead end. It is parked, visibly, with its reason attached and a note saying which
route actually does work on owned pages — and, crucially, it is parked **per finding**, so a second problem on the
same page is still recorded. Today it would not be.

There is an important thing this does NOT do, and I want to be plain about it: it does not repair the owned pages.
It converts "silently refused and forgotten" into "visibly waiting for a route". That is deliberate — the question
of what *does* repair an owned page is a different bug (277), and the owner already ruled that the two should not
be merged. But it is the precondition for answering it, because after this you can count the waiting work instead
of grepping for it.

**One thing I got wrong, and caught before writing any code.** My first design re-labelled the parked finding as a
generic "capability gap", following a convention we already use elsewhere. It had precedent and it was still wrong:
our detectors withdraw their own findings when the problem goes away, and they do that by matching the finding's
label. Re-labelling it would have meant nothing could ever withdraw it — so once a page was finally fixed, the
parked note would have sat there for ever. A review pass over the plan caught it. The fix keeps the finding's own
label and only changes where it sits.

**Something I noticed that is not mine to touch.** While measuring, one of my queries took sixteen minutes to come
back. The query itself took under three milliseconds. The rest was waiting: somebody had a large export running
against the page-components table for the best part of an hour, and somebody else's `ALTER TABLE pages ADD COLUMN
noindex` was queued behind it — which meant every read of the pages table across the whole fleet was stuck behind
that queue, including the page builds. It had cleared by the time I finished. Flagging it because "the database
felt slow this afternoon" has a specific cause and it is worth knowing the shape of it.

## 2026-08-24, later — it is built and committed

The fix is in (`6ab0b3434`), along with its documentation (`68734b771`). It will not do anything until the
chassis is next rolled out, because Go changes only take effect when a new image ships — the owner runs those,
and they go out fleet-wide.

**What it does, in one line:** when a defect is found on a page that belongs to a tool, and the worker it would
be sent to has declared that it refuses such pages, the finding is now set aside visibly instead of being sent
to that worker and thrown away.

Three things I want to record plainly, because they are the interesting parts rather than the code.

**The design I started with was wrong, and the review caught it, not the tests.** My first plan re-labelled the
set-aside finding as a generic "capability gap", copying a convention we already use. It looked right and had
precedent. But our detectors withdraw their own findings when a problem goes away, and they find them again by
the original label — so re-labelling would have meant nothing could ever withdraw it. Every one of those notes
would have sat in the queue for ever, and worse, would have blocked the detector from re-reporting the problem
if it came back. I only found this because I had the plan adversarially reviewed before writing code. That is
the cheapest possible place to be wrong.

**Adding the check silently broke other people's tests without failing any of them.** This is the one worth
knowing about. The new check is deliberately forgiving: if it cannot reach the database it shrugs and carries on,
so a finding is never lost. But our test framework reports an unexpected database query as an *error on that
query* — which is exactly the thing the check shrugs off. So twenty-one existing tests carried on passing while
quietly no longer testing anything. Nothing went red. I found them by temporarily making the check shout every
time it was in that state, running the whole suite, and collecting the names. All twenty-one are fixed, and I
have left two ready-made helpers and a written warning so the next person does not spend the hour I did.

**Two other sessions were working in the same file.** One messaged me before touching it, which is the reason
this went smoothly — we agreed I would land first, and I sent them the exact shape of my changes so they can
write theirs on top. While testing I also found that a third piece of someone's unfinished work in the shared
tree breaks one of its own tests; I checked it was not caused by my change (by turning my change off and seeing
it still fail), left it alone, and told them.

**What this does not do, and I want to be honest about it:** the owned pages still are not being repaired. The
defects on them are real. What has changed is that they are now visibly waiting with a note saying what would
work, instead of being recorded as "we decided not to fix this". That was the whole point of splitting this from
the repair question, which is a different bug and someone else's.

## 2026-08-25 — it is live, it is working, and it is not finished

The build went out and the check is now doing its job on real traffic. Since it went live at about half past
seven on Sunday evening, **32 findings on tool-owned pages have been set aside properly instead of being thrown
away**. Before this, every one of those would have been marked "we decided not to fix this" on a defect nobody
had decided anything about.

Two things I checked rather than assumed, because both could have gone the other way:

**The routes that were working still work.** The main way we legitimately rebuild owned pages completed 244
times in the same window, untouched. That was the whole point of having the check ask each worker whether it
refuses owned pages, rather than hardcoding a list — get that wrong and we would have quietly broken the busiest
repair path on the estate.

**A page can now hold more than one set-aside finding.** There is a page carrying two, separately recorded. The
old mechanism kept one note per page, so a second problem on an already-noted page vanished without trace. That
was the specific hole this was filed to close, and it is closed.

**Why I am not closing the bug.** Three findings were still thrown away *after* the fix went live — at ten past
eleven on Sunday night, three hours in. They came from one report-writing component that inserts its work
straight into the database and never passes the point where the new check sits. So the sentence at the top of
the bug report is still true of that one producer. Closing it now would be recording a success that the evidence
does not support, which is the exact failure this bug is about.

**What is left is small and well-defined.** One more change closes the gap — and the cheaper version of it is a
single check at the point where work is promoted for dispatch, rather than editing nine separate places. Beyond
that, there are 111 old records still holding the wrong worker. None of them are doing anything — they are not
consuming resources, they are just wrong on paper — so what happens to them is a decision rather than an
emergency, and it is yours.

The handoff for picking this up in a fresh session is at
`docs/agent_docs/docs024_key_docs_latest/bugfix_333_owned_page_door/HANDOFF_2026-08-25_continue_here.md`.

**2026-08-25, later — the second look, and your three decisions.** You asked me to run the morning's
analysis past a second, different reviewer. Good call: it found I had got three things wrong. The seven
failures I said belonged to another team's bug actually belong to the CTA work — their own notes even said
so, and I hadn't read them. The thirteen "bypass" rows aren't a bypass at all: that producer goes through
our new check, but drops the page's ID on the way in even though it looked the ID up moments earlier — so
the check never fires. One missing field, not a missing capability. And my "53 all-time" count for the
name-only class was really 272, because I'd only counted the current table and not the archive.

The leftover work is therefore three small, separate things, and you've now decided all three: the
report-writer that inserts straight into the database will be routed through the shared write path; the
rerender escalation will check whether a page is owned before raising the alarm, so those false alarms stop
at source; and the old records get left alone — most of them turn out not to be ownership refusals at all,
and the rest age out on their own within days.

One more thing worth knowing: two other teams' notes said their stuck items "now park under our new check".
They never did and never could — their items point at a different worker, one our check deliberately leaves
alone. I've corrected both documents and left them a note explaining why, with the numbers.

**2026-08-26 — closed.** The last two gaps are not just deployed but seen working on real traffic:
twelve times since yesterday evening a rerender hit an owned page's empty widget slot — the exact
false alarm that used to mint doomed tickets, on four of the exact pages it used to burn — and all
twelve times it stepped aside, wrote nothing wrong, and the page's standing review record covers it.
No finding on an owned page has been thrown away since the fixes went in. One honest footnote in the
close: the report-writing seats are now also stopped one layer earlier by a newer recording mechanism
another thread built, so our check there is the safety net rather than the thing firing daily — and a
daily auditor now watches the setting our check depends on. The bug file has moved to the closed
shelf with the numbers in it; the summary alongside it is written to be read aloud.

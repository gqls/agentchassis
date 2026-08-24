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

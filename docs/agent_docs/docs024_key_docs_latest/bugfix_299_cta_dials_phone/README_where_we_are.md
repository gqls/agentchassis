# Where we are — the phone-dialling button (append-only, newest at the bottom)

## 2026-08-18 — what we found, in plain terms

You reported a button on the webdesign.uk home page whose text talks about answering a few
questions with the Brief Starter tool, but clicking it opens the phone dialler. We took that
bug on, and the short version is: the system already computes the right link for that button
— and then throws its own answer away.

There is a piece of the pipeline whose whole job is to work out where each button should
point (it is called the internal link resolver). On the very rebuild that regenerated your
home page, it decided — correctly — that the button should point at the Brief Starter tool.
But the step that hands the finished sections to the page writer looks for the resolver's
answer under a name the answer is not stored under. It never finds it, and it silently falls
back to the old stored values — which is where the phone link lives. So every rebuild
recomputes the right answer, discards it, and re-ships the phone link. That is why the button
survived a complete rewrite of the page.

Three more things compound it. The checker that is supposed to catch "the words promise one
thing, the link does another" deliberately ignores phone and email links, so this exact
button is invisible to it. The repair machinery, if it ever did touch a phone link, would
replace it with a link to some tool page — which would destroy the genuine "call us" buttons
on your FAQ and how-it-works pages (this exact accident happened in production on another
site on the 17th — a "Start a Conversation" button pointing at the contact page was rewritten
to point at a password checker). And the phone numbers themselves are written in a format
phones cannot actually dial.

The fix is at the framework level, not a hand-edit of your page: teach the system that a
phone/email/external link can be deliberate and must be kept (and tidied into a dialable
format); teach the checker to see this class; tell the writer what the button's destination
actually is, so it writes words that match; and fix the one wrong name in the hand-off so the
resolver's answer is actually used. The order matters — the hand-off fix must go last,
after the protective parts are live, or it would trigger the button-destroying behaviour
above on other sites.

Four things need your call (also in the plan): whether the hand-off fix waits for a
neighbouring fix to land too (recommended — a second team is fixing the contact-button half
of the same trap); whether the home-page button should end up as a phone button with honest
words, or a link to the Brief Starter tool; confirming the right phone number for the
contact page (we think +44 7934 524 911); and a look at the numbers before we switch the new
checker on, since its findings go to the human review queue.

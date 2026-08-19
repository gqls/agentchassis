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

## 2026-08-18, evening — the fix is written, tested, calibrated and committed; three switches wait, on purpose

Everything described this morning is now built and on the shared branch (commits 757a0890a
and 0f483c8ab), reviewed by the council gate (submission 1f1fecc9, verdict pending — the
code ships regardless, that is how review works here). Before writing the detector we ran
it over the whole fleet as a dry run: the first version — the scope we'd originally chosen —
would have raised 226 alarms of which about 211 were wrong (mostly news headlines and email
links that were perfectly fine). We narrowed it and re-ran: 17 alarms, every one checked by
hand, every one real. That narrowing means "copy names our page but the link goes to some
other website" goes undetected for now — written down as a known gap, not swept under.

We also coordinated with two neighbouring efforts working the same files at the same time:
the contact-button fix (their half is already live) and a markdown-cleanup fix (their edit
rides inside one of our commits, by agreement, labelled).

Three switches are written and deliberately NOT thrown (files 475/476/477, all marked HOLD):
turning the new detector on, telling the writer what each button's destination is, and — the
big one — fixing the hand-off so the link resolver's answers are actually used. The last one
is only safe after the next build of the system is rolled out and verified, because using
the resolver's answers today would overwrite good contact buttons on other sites. After the
roll: verify, throw the switches in order, canary one site, then rebuild the webdesign.uk
home page through the normal pipeline and check the button by its text.

## 2026-08-19 — the detector is on, the button is not fixed yet, and one thing needs your decision

Short version: the protective machinery we built on the 18th is now running in production,
two of the three switches are thrown, and the third — the one that actually makes your home
page button behave — is waiting on you.

**Your button is still wrong, and it told us something useful by staying wrong.** The label
changed again yesterday morning; it now reads "Read the full terms in our FAQ before you pay."
and it still dials the phone. That is the fourth different sentence over the same unchanged
phone link. Whatever writes the words and whatever writes the destination are two different
things, and only one of them is being updated. That is exactly the fault we diagnosed, so
seeing it happen a fourth time is confirmation rather than bad news.

**We proved the diagnosis properly this time.** On the 18th we had one build traced in detail.
Now we have the whole fleet: across every recent page build we could examine, the link
resolver worked out the right destinations 26 times and **not one of those answers survived**
to the part of the system that writes the page. Along the way we found a much better way to
measure it. The resolver also writes a little note recording what each button points AT, and
nothing else in the system produces those notes — so if the note is missing downstream, the
answer was thrown away, full stop. That test works even on the builds where the link happens
to be right already, which the old test could not distinguish from a healthy one.

**What went live today.** The code shipped with the fleet's overnight roll, and we confirmed
it is genuinely running by asking the running program directly (with a deliberate wrong
answer included as a control, to prove the question was being answered honestly rather than
always saying yes). We then switched on the detector that finds this class of fault — it was
completely blind to phone and email links before, which is why your button never appeared in
any queue and had to reach your eye instead. We also switched on the part that tells the copy
writer what a button actually points at, so it can write words that match. That second one is
correctly doing nothing at the moment, and will keep doing nothing until the third switch is
thrown — we have written that into the file itself, because otherwise the next person to check
will see zero and think it is broken.

**What needs your decision.** The third switch is the repair to the hand-off — the one that
makes the system actually use the link resolver's answers. It is safe now in the sense that
both protective pieces are live and verified. But it changes how buttons are written on
**every** site at the moment it is applied, and the sensible thing is to watch one site
closely as it happens. So: may we apply it and use leopardessconsulting.co.uk as the canary
(it has four hand-written "contact us" buttons that must survive untouched)? Two smaller
questions from the 18th are also still open: whether your home page button should end up as a
phone button with honest wording or as a link to the Brief Starter tool, and confirmation that
the intended phone number is +44 7934 524 911.

**One broader thing we found, and wrote up for you separately.** While doing the safety check
that was supposed to gate all this — "is the new code definitely running before we switch
anything on?" — we discovered the standard way of doing that check does not actually work most
of the time. The program announces which version it is when it starts up, but that
announcement scrolls out of the log within a couple of hours, and the alternative check gives
a confidently wrong answer. We got a "no, your fix is not there" for code that was certainly
there. Nobody has been harmed by this yet, as far as we can tell, but it means everyone has
been improvising this check privately. There are 32 places in the system that promise to do
this check and none that can enforce it. We have written a proposal (RFC_040) to fix it
properly, and in the meantime written down the reliable method so the next person does not
have to rediscover it.

## 2026-08-19, later — the four decisions, explained, and one of them changed shape

You asked what you actually need to decide. Four things, and the second one is not what it was
this morning, because we found the evidence that settles it.

**1. Do we apply the wiring fix?** It is one line of configuration. The page writer looks for
the link resolver's answers under a name that does not exist, finds nothing, and quietly uses
the last build's values instead. Applying it corrects the name. The reason it has been held is
that configuration goes live everywhere the instant it is applied, and until yesterday using
the resolver's answers would have overwritten hand-written "contact us" buttons on other
sites. The protective code is now live and verified, so that danger is closed — but it has
never actually run, because its output has always been thrown away, so this is the first time
it does real work. Two things make the risk small: we can see exactly what it will release
(below), and the fleet only rebuilds **one to seven pages an hour across five sites**, so a
mistake shows on the first build and reverts in one line, having touched a handful of pages.
Recommendation: apply it, and watch leopardessconsulting.co.uk, which has four hand-written
contact buttons that must survive untouched.

**2. What should your home page's second button be?** This changed. We found the last build of
that page still on record, and it shows the new code already computing the right answer and
having it discarded: it kept the phone link, **tidied it into a dialable form**
(`tel:+447934524911`), and wrote a plain note saying the destination is "a phone call to
+44 (0) 7934 524 911" for the copy writer to work from. So after the wiring fix, that button
becomes a working phone button and the words should be rewritten to match it — the mismatch
resolves itself, in favour of the phone.

The thing worth your attention: that phone link looks like it arrived by accident. On 13
August the section read "Prefer to talk it through first? Call +44 (0) 7934 524 911 or
email…", which is a real phone button. The words have been rewritten four times since and the
link never moved. **The system cannot tell a deliberate phone link from an inherited one** — it
now treats any phone or email link as intentional and defends it, which is exactly what stops
your FAQ and how-it-works "call us" buttons being destroyed. The cost of that protection is
that it will also faithfully protect a leftover. So the question is simply: should that button
be a phone call, or the Brief Starter tool? Phone means do nothing. Tool means say so, and we
change the stored value, because otherwise the protection keeps the phone link indefinitely.

**3. The contact page's number.** A separate, genuinely broken one: the "(0)" has been
swallowed into the digits, leaving something no phone can dial. The code repairs only what is
unambiguous and **refuses** this one rather than inventing digits, and has filed it for a human.
We think the intended number is +44 7934 524 911. Confirm and it is a one-line fix.

**4. Is RFC_040 worth building?** Plainly: every service knows what it can do, but that
knowledge only exists inside the running program and in one line it prints when it starts,
which scrolls out of the log within hours. This matters because code needs a rebuild and a roll
while configuration is live immediately, so many changes arrive in two halves that must land in
order. Thirty-two of our configuration changes promise to check that the code half is live
first. **None of them can**, because a database cannot ask a running program what it can do.
Worse, the check we tell everyone to perform often cannot be performed at all — yesterday the
startup line was gone three hours after the roll and the documented fallback told me a fix was
missing when it was certainly there. Another thread hit exactly the same wall on 11 August.

The honest argument against building it: nothing has actually gone wrong yet. It works because
people are careful. That is a near-miss, not damage, and you should weigh it as one.
Recommendation: agree the problem is real, and build only the small half — have each service
write down what it can do when it starts. That alone ends "I cannot find out what is running",
costs little, and changes no behaviour. The part where a configuration change refuses to apply
itself can wait until something else wants it; building a contract for a single user is how
mechanisms rot unused.

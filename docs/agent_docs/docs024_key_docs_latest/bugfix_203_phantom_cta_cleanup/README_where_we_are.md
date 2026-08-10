# README — where we are (plain prose, append-only, newest at the bottom)

## 2026-08-06 — picked up the loose ends of the phantom-link bug

Yesterday a session found why several sites had buttons pointing at the wrong page:
when the system couldn't work out where a button should lead, a shared helper quietly
filled in "/contact.html" rather than leaving the button out. So a real button label —
"Read the tungsten percentage guide", say — got glued to the contact page. That helper
is now fixed, the review council approved the fix, and we've confirmed the fix is in
the software actually running in production.

What's left, and what this thread is doing: thirteen wrong buttons are still live on
seven sites, because fixing the machine doesn't retouch what it already shipped. We'll
get those pages rebuilt through the normal pipeline — first letting the link resolver
have another go at finding each button's real target, so buttons come back pointing at
the right place rather than simply vanishing.

Two things the review council asked for alongside, which we're taking on: check the
same file for OTHER quietly-invented defaults (we already found two more — a
"primary" and "secondary" button link that still default to /contact.html and
/about.html on a backup rendering path), and check why the automatic detector only
caught 2 of the 13 wrong buttons on its own.

## 2026-08-06 (evening) — what we actually found, and why we didn't change any code

Three things worth knowing, in plain terms.

**The fix is genuinely in production.** We didn't take anyone's word for it, and we didn't
rely on the version number either — a version number here tells you nothing about what's
inside it. What settles it is that our fix sits underneath another fix that a different
thread proved, at the running process, against real traffic earlier today. Since every
build is taken from committed history, anything underneath a proven change is necessarily
in the same binary.

**The wrong buttons are fewer, and less broken, than the bug file says.** Yesterday's count
was thirteen. The honest count today is four genuinely misleading buttons — a "Run the Risk
Checker" and a "Run MatchMatrix" and two similar — plus four more on a blog that say
"Get Started" and were never really written by anyone; the system invented the words as
well as the destination. The remaining ones we'd been counting say things like "Talk to us"
and "book a discovery call" and point at the contact page, which is exactly where they
should point. And the contact page exists on all seven sites, so none of this was ever a
broken link — it was buttons pointing somewhere other than what they promised.

Worth adding: the page that started the whole bug, the darts news page, quietly repaired
itself this evening. It got rebuilt through the normal pipeline on the fixed software and
the wrong link simply isn't there any more. That's the clean-up route working, on the
original casualty, without anyone hand-editing anything.

**We deliberately made no code change, and that's the interesting part.** The review
council had asked us to check whether the same "invent a plausible default" habit was
hiding elsewhere in that file. It is — in two more places. But when we looked at what
would happen if we simply deleted them, we found we'd trade a quiet wrong link for
something worse: raw template gibberish printed into the page where the address should be.
One page on another site is already doing exactly that, so this isn't theoretical. The
council reviewer who warned about this was right. Fixing it properly means changing how
that older rendering path behaves when a value is missing, which is a bigger change to
shared machinery and deserves its own review round rather than being tacked onto a tidy-up.
Both of the two remaining cases are, in any event, currently harmless: we checked every
page on the fleet and neither has produced a single bad link.

**One correction to the bug file, and one to ourselves.** The bug file blamed the automatic
detector for not running often enough. That's not true — it has run across ten sites and
filed 123 findings. The problem is that nothing ever picks those findings up: they're all
sitting in a queue waiting for a human, with no handler assigned. That's a different and
more serious problem than the one we thought we had. And a mistake of my own: I tagged a
documentation-only commit as council-reviewed, which quietly pollutes the report that
tracks which code changes went unreviewed. Logged it rather than let it pass.

## 2026-08-07 — the tidy-up turns out to need a decision from you

I picked this back up expecting to fix the eight wrong buttons and found the plan I'd left
myself was wrong, so here's the honest position.

**Three of those buttons have a correct destination sitting right there.** "Run the Risk
Checker" on finetuning.uk, "Run MatchMatrix" on robot-hands, "Score your process" on
leopardess — each of those tools genuinely exists as a page on its own site. That matters
because the fix we shipped makes the system leave a button out when it can't work out where
it goes. So if I simply rebuild those three pages, the wrong link disappears and so does the
button — on pages whose entire job is to send the reader to that tool. Honest, but worse for
the visitor.

**And the obvious remedy isn't available.** I'd assumed we could just ask the system to
re-work out the links for one page and then rebuild it. It can't: the part that works out
where links should point only runs while a page's words are being written, not as a repair
you can run afterwards. I checked the other repair machinery too — it only mends links that
point at pages which don't exist, and ours point at the contact page, which does exist on
every one of these sites. So there's nothing in the system today that re-aims a link that's
live but wrong.

That leaves three ways forward, and the middle one needs your say-so:

1. **Rebuild the four blog buttons on leopardess and let them vanish.** These say "Get
   Started" — words the system invented, on buttons nobody ever wrote. Nothing worth keeping.
   ~~I'd do this one without asking, it's clearly right.~~ **Corrected an hour later, before
   I did it:** those four pages were last built 8–9 days ago, and there have been 244 changes
   to the page-building code since. Rebuilding doesn't apply just our button fix — it applies
   all 244, to a live customer page, and I can't tell you in advance what else would move.
   Meanwhile these four are the *least* harmful of the eight, because "Get Started" pointing
   at the contact page isn't really a broken promise. So it isn't worth doing casually after
   all, and I've left it.
2. **Let the content writer edit those three tool pages.** This is the system's own proper
   route and needs no new machinery — the link-resolving step is already part of writing, so
   the buttons would come back pointing at the real tools. The catch: it puts the AI writer
   over copy that's already published on customer sites. It edits rather than rewrites from
   scratch, but it is still changing live pages, so I'm not doing it on my own initiative.
3. **Build the missing piece properly** — a way to re-aim an existing page's links without
   touching its words. That's the right long-term answer and it would fix this whole class
   rather than these eight, but it's new shared machinery and wants a review round.

**Also worth flagging, because it changes something I told you yesterday.** I said the
detector's findings just need someone to act on them. Looking at the actual findings, most of
them are wrong: it's flagging perfectly good "Get in Touch" buttons that point at the contact
page as if they were misdirected. If we'd built something to auto-apply its suggestions, it
would have broken a lot of working buttons. So that queue needs its accuracy sorted out
before anything drains it — I've corrected my own note.

One good bit of news: the code side is now measured and it's a non-issue. The risky old
rendering path I was worried about didn't run once in five and a half hours of live traffic,
and neither of the two remaining invented-default cases can fire on a path that never runs.

## 2026-08-07 (later) — we tried it on one page, and it taught us the real problem

You said go ahead with the careful content-writer route, so I did it on one page first — the
finetuning.uk risk-checker guide, chosen because it was the freshest of the four and so the
least likely to spring surprises. Here's what happened, and it was worth doing on one.

**The good part.** The page has stopped lying: the "Run the Risk Checker" button no longer
points at the contact page, verified on the live site, not just in the database. And the
cautious editing mode did exactly what it promised — the article body and the closing section
came back **byte-for-byte identical**. Nothing was quietly rewritten. That's a real
endorsement of that mechanism, which I'd flagged as unproven.

**The bad part, and it's the useful bit.** The button wasn't re-aimed — it was removed. So we
spent an AI writing pass to achieve exactly what a plain rebuild would have done for free.

**And the reason why is a genuine defect we didn't know about.** Watching the run mid-flight, I
could see what the link-resolving step decided: it pointed "Run the Risk Checker" at the
**password strength calculator**, and gave the actual risk checker to the *other* button on the
page, the one that says "Speak to Us About Data Privacy". It swapped them. That wrong link
never reached the page, but only by luck — not by any safeguard.

Underneath that there's a structural reason my instruction could never have worked: the button
label is something the AI writer controls, but the button's **destination is not** — that
belongs to the automatic resolver. So telling the writer "point this at the risk checker" was
asking it to change a field it isn't allowed to touch. I'd copied that instruction pattern
from working code elsewhere without noticing that the other case puts its link inside
sentences, where the writer *is* in charge.

**So I've stopped, and not done the other three pages.** Repeating this would spend three more
AI passes to delete three more buttons, one of them on a page that hasn't been rebuilt in
nearly two weeks. The thing actually worth fixing is the resolver picking the wrong slot. I've
left the other three fully worked out — page ids, correct destinations, all verified — so
whoever fixes the resolver can finish them in minutes.

One small thing I should own: the hero's second button text, "Speak to Us About Data Privacy",
was emptied during the edit. No prose was lost, but that phrase is gone.

## 2026-08-10 — the review council said yes, the fix is live, and testing it on one real page found a second bug I've now also fixed

Picked this back up cold. First thing: the review council's verdict on the resolver fix
came back approved, and I checked the actual software running in production, not just
the record — both copies of it have the fix in them. Good news, and also a small
surprise along the way: one of the two buttons we'd been holding back because it
seemed too risky to touch got repaired automatically by the site's own maintenance
routine, on its own, half an hour after the fix went live. It didn't fix it correctly —
it pointed the button at an unrelated tool — but I worked out why, and it's not
something either version of the fix could ever have got right: the correct destination
for that particular button is an ordinary page, not a tool or an index page, and the
fix only ever looks at tools and index pages. That's a real limit, not a bug, and I've
written it down so nobody spends time trying to make it work.

The other held-back button was different — its correct destination genuinely is a
tool, so I tested it for real: promoted the button's existing repair ticket and let the
system run it. **The main result is good news** — the button now points at the right
tool, live, the first time this fix has actually proven itself on a real customer page
rather than just in test data. But watching the same page rebuild, I noticed a second
button on it went wrong in a new way: it should have pointed at a "browse our catalog"
page, and instead it pointed at an unrelated calculator tool, even though the catalog
page was a much better match by any reasonable reading of the button's own words.

I dug into why, and it's a real defect in the matching logic itself, not a one-off
fluke: the code was written to always prefer sending people to an interactive tool over
a plain page, and it applies that preference too bluntly — it doesn't check whether the
plain page is actually a much better match first. So a weak, barely-related tool match
can beat a strong, obviously-correct page match. I found the exact line responsible,
fixed it so it compares match quality first and only falls back to "prefer the tool" as
a tie-breaker, added tests that prove the fix actually changes the outcome (not just
tests that would pass either way), and submitted that fix to the same review council
for its own check before it goes live. I've held off dispatching the four remaining
buttons on the other site until that verdict comes back, since the same defect could
quietly make any of them worse rather than better.

One thing I should flag plainly: I wrote the fix and committed it before I'd finished
preparing the submission to the review council, which means that commit is missing the
little marker that normally links a commit to its review automatically. It's not a
real problem — the review is still happening, and I've written down where to find it —
but the automatic bookkeeping that's supposed to show "this was reviewed" won't show it
for this one commit. I've made a note of that too, so it doesn't get missed.

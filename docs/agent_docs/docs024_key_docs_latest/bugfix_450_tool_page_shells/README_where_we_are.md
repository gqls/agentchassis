# Where we are — the tool pages that are not tools (bugs_open/450)

*(Plain-prose log for the owner. Append only, newest at the bottom.)*

---

**2026-09-03, morning — what this is, and why it is a separate piece of work from the seotools clean-up.**

When seotools went live it had seven pages at addresses like `/tools/robots-txt-tester/`. All seven
answer, all seven look finished, all seven carry the right headline — and none of them contains a
tool. They are articles *about* tools. The site whose entire product is tools shipped seven pages
promising tools that are not there.

You have already decided what to do about those seven: build the tools they promise. That is the
portfolio lane's job and it is in hand. This lane is the other half — making sure the next site
does not do the same thing, because nothing about seotools was special.

Here is the sequence, because the shape of it is the whole problem. The planner writes the site
plan and names the tool pages. At that moment the tools do not exist: they are invented later by a
separate process that visits one site every three hours or so, and when it arrives it invents its
own names — on seotools, not one of the seven names it chose matched a name the planner had used.
So the planner is naming pages that nothing will ever fill. Meanwhile the plan validator does spot
the problem and files a note saying "this page needs a proper build, not the generic one". Nobody
reads that note. It goes into a queue and sits there.

Then the ordinary machinery does exactly what it is built to do. Other pages link to
`/tools/robots-txt-tester/`. A checker notices the link points at a page that has not been built,
and files a repair: "build the target page". The builder picks it up, reads the plan, and the plan
says the page consists of a headline and a text block. So it writes a headline and a text block,
deploys it, and the link is no longer broken. Everything reports success. There is a guard meant to
stop exactly this — but it asks a question that does not apply here, so it waves the page through.

The blunt version: **a tool page is judged by whether it has a form on it, and nothing in the
system was asking that question.**

**What I am changing.** Two things, independent of each other.

The first is a rule the builders consult: *a page typed "tool" that contains no tool is not
available for generic building.* Note what that rule is not — it is not a flag someone sets and
someone else must remember to unset. It is a question asked fresh each time, so the moment the real
tool arrives the page stops being protected and everything proceeds normally. That mattered more
than it sounds: the obvious version of this fix was to mark the page somehow, and when I looked, I
found the field people would use for that has never once been changed by any code in this system,
in either direction. A mark nobody clears is a page that stays stuck for ever.

The second is upstream: stop the planner planning tool pages whose tools do not exist. That one is
a sibling of a gate another session shipped yesterday for a near-identical problem with listing
pages, and I am deliberately reusing its shape and giving it its own on/off switch, so if mine
misbehaves theirs is unaffected.

**One thing I checked before writing any of it,** because it could have broken your seotools
clean-up: when the tool pipeline builds a tool onto one of those seven pages, does my new rule get
in its way? No — the tool component is inserted before the pipeline asks for the surrounding
content, so by the time anything consults the rule, the page has a tool in it and the rule stands
down. Your wave is safe either side of this landing.

**Timing.** A chassis build is going out now; my work is not in it and will ride the next one after.
I will say plainly when it is live rather than implying it from the fact that it was committed —
those are different things here.

---

**2026-09-03, late morning — the guard is written and committed. Not live yet, and I will not say it is until I can show you.**

The rule I described is in. A page typed as a tool that has no tool on it now refuses to be built
by the generic builder, and it refuses at every door I could find that leads to one — six of
them. It is committed and it compiles and its tests pass, but it is **not running anywhere yet**:
this system only picks up code changes when a new chassis image is built and rolled out, and the
one that went out this morning was built before my change existed. So the seven pages on seotools
are still exactly as they were. When it does go live I will check it by looking at what the
machine actually did with a real piece of work, not by looking at the fact that I committed it.

**Two things went wrong along the way and both are worth telling you about.**

The first was mine. While I was still designing, my half-finished edits were sitting in the shared
working folder that every session on this machine uses. Another session committed a one-line fix
of its own to a file I happened to have open, and the way this repository works, its commit
correctly took my unfinished work with it. The result was that the main branch briefly did not
compile, which blocks *everyone's* next build. That session noticed, worked out exactly which of
my files would fix it, and — rather than committing my unfinished work under its own name and
guessing whether it was ready — messaged me and asked. That was the right call and I have said so
to them. The underlying mistake was not theirs, though: it was mine, for leaving a half-finished
rename lying around for hours. A rename is the worst possible thing to leave unfinished here,
because it breaks the code for everyone from the first keystroke until the last piece lands. I
have written that up in the fleet's log of wrong calls under my own name.

The second was a claim I made. I had cited an old database script as evidence that pages
sometimes get mislabelled as tools. It does not say that — it fixes a different thing entirely. I
caught it before it went anywhere important, but it had already reached my design notes, and it
would have gone to the reviewers as evidence. It is withdrawn in the review submission itself
rather than quietly dropped, because a withdrawn citation is information and a disappeared one is
not.

**One genuinely useful thing came out of talking to the other sessions rather than working alone.**
I mentioned in passing what my change did to a function another session owns. They replied with a
warning I had no way of knowing: since yesterday, asking this system to "re-render" a page makes
it hand the page its own old content back and report success. If I had verified my fix by
re-rendering one of these tool pages and looking at the result — which is the obvious thing to do
— I would have been looking in a mirror and drawing conclusions from it. Their lane lost a day to
exactly that. My verification now deliberately reads the work queue and the live page instead, and
the reason is written into the runbook so the next person does not reach for the obvious thing.

**Where this sits now:** the fix is with the automated reviewers. The other half — stopping the
planner from inventing these pages in the first place — is designed but not written, and I have
offered it to the session whose framework I would be borrowing, since it is their design and they
may prefer to own it. Meanwhile the seven seotools tools have all been built by the other session
and are queued to publish, so the pages you were told about are being fixed on their own track.

---

**2026-09-03, afternoon — both halves are done and approved. Still nothing live, and I want to be precise about that.**

The second piece is written too: the planner will stop inventing tool pages whose tools do not
exist, rather than us catching them downstream. Both pieces went through the platform's automated
review council and both were approved.

**The approvals are the least interesting part, and I would rather tell you what the reviewers
caught.** On the first piece, all four of their substantive objections pointed at something I had
asserted rather than checked — and one of them was simply false: I had written that a particular
database field had never been changed by anything, and six older maintenance scripts change it.
The conclusion I had drawn from it still holds for a narrower and better reason, but the claim was
wrong and it was sitting in the code where the next person would have believed it. On the second
piece, a reviewer asked for the riskiest behaviour to be pinned by a test — and the comment in my
code already claimed that test existed. It did not. I had written the description of the test and
not the test. Both are now written up in the fleet's log of wrong calls under my name, and the
missing test exists, along with a second one proving the first is not vacuous.

**What is live: nothing.** The server image running today was built before any of this. The
database change is written, rehearsed against the real system inside a transaction I rolled back,
and deliberately **not applied** — because the sentence it installs tells the planner that a
validation is running, and that will not be true until the code ships. I would rather hold it than
ship a prompt that asserts something false.

**Three other sessions were involved and all three made the work better.** One caught that my
count of affected pages was a floor rather than a total — the query I had repeated five times
could not see the very case I had been calling my best evidence. One asked me to protect a
sentence in a shared configuration row that their live detector depends on; without it, a change
of mine could have made their detector look like it was reporting a different problem entirely.
One confirmed a detail from its own source and warned me not to record any fixed position or
length — which caught a figure in my own file that was already out of date within the hour,
because three sessions edit that row.

**The honest weak point, written down rather than buried.** The second fix rests on a measurement
that nothing in the system reads these planned pages. That is a statement about absence, and
statements about absence quietly stop being true when someone adds something. If a future process
starts reading them, my gate would starve it silently. Turning that assumption into a standing
check is the most valuable next thing to build here, and it is named as such.

**Next:** when the next server roll happens, prove the fix by watching a real piece of queued work
be refused with a receipt — not by observing that no new bad pages appeared, which would prove
nothing unless something had tried. Then apply the database change. The seven seotools pages are
being fixed on the other lane's track and are queued to publish.

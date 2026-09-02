# README — where we are (plain prose, append-only, newest at the bottom)

## 2026-09-02

Picked up `bugs_open/426`. In plain terms: there's a database migration tool on
this platform, and it has a safety check built into it already — every time you run
it without any arguments, it tells you which database changes are waiting to be
applied, and it double-checks each one by actually trying it in a way that can't do
any harm, just to see whether it looks like it's secretly already been done by hand
and never written down. That "written down" step matters, because if it's skipped,
the tool will confidently redo something that's already been done, and that can
break in a confusing way — it happened for real once and stopped the whole migration
tool from working for three days.

The check that would have caught this already exists. It's free. It's even written
into our own house rules that every session should run it. The problem this bug is
about is simply: nobody was running it, because it's a manual step and manual steps
get forgotten. One team noticed seven cases of exactly this problem sitting quietly
in the system for about a month, only because someone happened to look.

Before doing anything else, I ran that check myself today to see how bad it
actually is right now. The answer: worse than the seven. Out of everything currently
waiting to be applied, **34 separate cases** show the "this looks like it's already
been done" warning, live, today. So this isn't a rare edge case — it's a steady drip
that nobody is watching.

The fix I'm building is simple in spirit: put that same free check on a timer, once
a day, and have it leave a note behind every single day — even on a day when
everything's fine — so that a missing note becomes its own warning sign ("did the
check even run today?"), rather than a clean report and a broken check looking
identical. This is the same pattern already used successfully for a couple of other
quiet-drift problems on this platform, so it's not a new idea, just applying it here.

I've told the team that found the original seven cases what I found, so they're not
duplicating effort or caught by surprise. I've also checked — this kind of change
(a new scheduled housekeeping job, no changes to the core platform code) doesn't
need to go through the formal reviewer council; the same kind of change was built
before and explicitly didn't need it, for the same reason.

Next: I'm having a planning pass (using a different, non-default model, at the
owner's request) work out the exact technical shape of the fix — mainly, how to get
the real content of each pending database-change file safely into the check's own
little container without duplicating any of the existing tool's logic, since
duplicating that logic is precisely the kind of mistake this platform has been bitten
by before. I'll build from that plan once it's back.

## 2026-09-02, later — it's done, and there were two real snags worth knowing about

The planning pass I mentioned didn't finish — the model I was asked to use ran out
of its own quota partway through. Rather than wait around for it, I did the design
work myself, using everything already gathered, and then — instead of just handing
over a design on paper — I actually built it and tested it for real against a
throwaway copy first, before touching anything that matters. That's what caught the
two real problems below; a design that only ever exists on paper wouldn't have.

The first problem was small and technical: the little container this check runs in
doesn't have quite the same tools as a normal computer does, and one part of the
existing tool it reuses didn't work properly inside it. Easy fix, but it would have
made the daily report an unreadable mess if I hadn't caught it.

The second was more interesting. I built it, switched it on for real, and ran it
once to see what happened — and it failed. Not because the checking itself was
broken; when I looked at what it had actually written down, it had done its job
correctly, twice over. It failed because of how I'd told it to react when it found
something worth flagging: I'd copied the pattern from a similar check elsewhere on
this platform, where finding something is supposed to be rare and worth a loud
failure. For THIS check, finding something is the normal, everyday state — today it
found 72 things worth a look — so treating that as a failure meant it would try
the whole expensive job twice, every single day, and still end up looking "broken"
in the system's own records, forever, even though it was doing exactly its job.
I've fixed that: it now only reports itself as failed if the check genuinely
couldn't run at all, not if it simply found things.

It's genuinely running now, not just written and waiting for a release — I checked
by running it twice and reading back what it actually wrote to the database both
times. First real scheduled run (rather than one I triggered by hand) is tomorrow
morning, UK time-zone-adjusted, at 07:45 UTC.

One thing I deliberately did NOT fix, and said so plainly in the bug's own file: the
underlying check sometimes phrases its "this looks like it was already done" warning
in words that don't get picked up properly, so a few cases can slip through wearing
a different message instead. I found a safe way to catch those too in the daily
report without touching the more sensitive, older piece of code where that wording
problem actually lives — fixing that properly is a bigger, separate piece of work,
and I've left it clearly marked for whoever picks it up next.

# Where we are — the "as an AI" false alarm

Plain prose, append-only, newest at the bottom.

---

## 2026-08-08 — what this is, in one paragraph

We have a safety check that reads every page before it is saved and refuses the
page if it finds the sort of text a language model writes when it gives up:
"As an AI, I cannot generate this listing." That check is right to exist — it has
caught a real apology shipping as page copy before. But it looks for the phrase
"as an ai" as plain letters, anywhere in the page's prose, and English does not
cooperate. On webdesign.co.uk's tools index there is a tool called the JSON-LD
SEO Injector, described honestly as *"LocalBusiness schema, as an AI-builder
prompt"*. The checker sees "as an AI-builder", finds "as an ai" inside it, and
refuses the page.

**The consequence is worse than a warning, and this is the part worth
understanding:** the refusal is a *blocker*, which means the build stops with an
error and the page is never written. So that tools index cannot be rebuilt at
all — not the sentence, the whole page — for as long as that description is on
it. Nobody has noticed because nobody has asked to rebuild that page since the
copy landed. It is a trap sitting armed, not a fire.

The irony is exact: the check is most likely to fire on the sites where the copy
is most likely to be *correct*. Any page about AI tooling will reach for "as an
AI assistant", "as an AI-builder prompt", "as an AI agent". We build several of
those.

## What I did before touching anything

Two things, because a bug filed eleven hours ago on a shared tree is not
automatically still true, and because other people are working here.

First, I checked nobody else has this one. The lane that found it wrote in their
own handoff that it belongs to someone else and they will not be fixing it, and
I grepped the live transcripts of the fourteen most recently active sessions for
the name of the function — the one that came up is doing something else
entirely. I had been pointed at a different bug (116) to start with; that one
turned out to have two active teams on it and a run fired today, and the owner
has already ruled it is a decision rather than a coding job, so I left it alone.

Second, I proved the bug is still real *by running the actual code*, not by
searching the database for the phrase. Those are different questions: a database
search sees the whole page including the bits inside `<script>` tags, and the
checker deliberately ignores those since a fix last week. I pulled the five
pages that contain any of the suspicious phrases, ran the real checker over their
real stored bytes, and got exactly what the bug describes — one refusal, on the
webdesign tools index, quoting that sentence. The other four came back clean,
which matters more than it sounds: those four are the pages that *used* to be
wrongly refused before last week's fix, so their coming back clean proves my test
rig can tell a refusal from a pass. A test that only ever says "fine" is not a
test.

## Where we are going

The fix is to make the check look for what it was always for — a model talking
about itself in the first person — rather than for three letters that happen to
spell "ai". I have asked for a design and I will put it through the review
council before it ships. The one thing I will guard hardest is the reverse
mistake: after the change, "As an AI, I cannot generate this listing." must still
be refused. Loosening a safety check until it stops complaining is the easy
version of this job and the wrong one.

I am deliberately *not* widening this into "how should all our text-scanning
safety checks work", even though there is a second, near-identical bug open
against a different checker. That is a bigger question with an owner, and the
bug file itself says so.

## 2026-08-08 (evening) — done and reviewed, but not yet live

The fix is written, tested and approved. The check now looks for the *sentence
shape* a model uses when it apologises — "as an AI, I …" — instead of the three
letters "ai". The tools-index page it was blocking is clear, and the apology it
was built to catch is still caught: I took the real page, injected "As an AI, I
cannot generate this listing." into it, and the checker still refused it. That
was the test I cared most about, because the easy way to fix a check that
complains too much is to quieten it until it stops, and that would have been
worse than the bug.

The review council approved it first time round, with four advisory comments.
Two of them caught me out, and both are worth repeating because they are the
same kind of mistake:

**I said the change affected one place. It affects four.** I had counted the
places in the *code* that call this function — one — and reported that. The
reviewers asked the question I had not: how many of the running agents are
configured to use this check at all? The answer is four. Nothing I wrote was
untrue; I had just measured the half that was easy to look up and let it stand
for the whole.

**I claimed a verification step I could not actually perform.** The house rule
for proving a change reached production is to look for something the change
added *and* something it removed — the second half is what proves you are
looking at your own change and not somebody else's. When I went to write those
two commands down, it turned out my change removes nothing: every phrase it
touches is still there afterwards, by design. So the usual proof is not
available here. I have said so plainly rather than nominate some string as
"removed" and print a reassuring zero, which would have looked exactly like
evidence and been worth nothing.

**What is left.** The fix is Go code, so it does nothing until the chassis is
rebuilt and rolled — which happens on the fleet's schedule, not mine. Until
then the bug is still real in production, and the file stays open. When it does
roll, there is one command in the runbook to confirm it landed, with a note on
how to check that command can actually tell the difference.

One thing I have deliberately *not* done: there is a second, near-identical bug
in a different checker (a comment saying "no fabricated data" gets convicted for
containing the words "fabricated data"). One reviewer pushed on whether fixing
mine and leaving that one is really a fix at all. That is a fair challenge and I
have recorded it rather than argued it away — but that checker belongs to
another team who are actively working it, and the question of how *all* our
text-scanning checks should behave is a bigger design decision than this bug
should be allowed to settle on its own.

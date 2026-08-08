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

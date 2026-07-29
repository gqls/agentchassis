# Where we are — the scraper that lied, and the records that couldn't say where they came from

Plain prose, append-only, newest at the bottom.

---

## 2026-07-28, evening

I picked up two bugs nobody was working on: **100** and **101**. They're on the same
step of the same agent, and both bug files said, independently, "fix these together" —
so I did.

**What 101 was.** One of our agents is configured to visit a vet practice's website,
follow up to three pages — fees, prices, about, team, contact, services — and read
them. That's what the configuration says in plain text, and anyone reading it would
believe it. It has never done any of that. It fetches the home page, once. The four
settings that describe the crawl are read by no code anywhere. They sit there looking
implemented.

The reason this matters more than "a feature is missing" is that the configuration
*reads as evidence*. Someone looked at it two days ago, concluded that widening the
list of pages was a free win, and wrote that into four documents and a commit message
before catching it. The config was lying, convincingly, to careful readers.

**What 100 was.** We have a table whose entire job is to record where a piece of data
came from. It has 2,970 rows in it. Every single one has an empty source. Not "mostly
empty" — empty, all of them, since the table was created.

The cause is that the code asks the **AI model** where the data came from. And the
prompt never asks the model for that, so the answer was always blank. The tempting fix
is obvious and it is a trap: if we ask the model to tell us its source, then the same
call that produced the facts also produces the claim about where the facts came from,
and nothing checks it. We spent July cleaning up exactly that kind of self-asserted
claim across the estate. So the answer had to come from somewhere else.

It turned out the answer was already being written down. The component that actually
does the fetching records the URL and the time, right next to the HTTP call. Nobody was
reading it. The writer sat there asking the model a question that the fetcher had
already answered correctly. So the fix is small and it is the right shape: stop asking
the model, read the fetch record. I deleted the three lines that read the model's
answer rather than leaving them as a fallback — a fallback would have quietly come back
to life the first time a model volunteered a plausible-looking URL.

**The thing I didn't expect.** Bug 101 had a warning box in it: don't implement this
until you've settled whether our scraper is throwing away page footers, because company
registration numbers live in footers, and if we're discarding them then fetching more
pages achieves nothing. Nobody had settled it.

It's settled now, and the answer was a third bug, in shared code that every scrape on
the estate goes through. There's a setting called "only main content". If you turn it
**on**, you get just the article and no navigation or footer. Our code could send
"on". Our code could not send "off" — it only put the setting in the request when the
answer was "on", and when it was "off" it sent nothing at all. Firecrawl's default,
when we send nothing, is **on**. So every caller that explicitly asked for the full
page has been getting the opposite of what it asked for, silently, forever.

Three live steps ask for the full page. One of them is the step that fetches a site's
stylesheet. The very same file gets it right on the other code path, twenty lines away.

**What I built beyond the three fixes.** The pattern underneath all of this is that a
setting nobody reads looks exactly like a setting that works. So I made that
detectable: an action can now declare which settings it actually reads, and anything
else in its configuration gets reported instead of silently ignored. There's also a
report that scans the whole estate and asks the same question offline.

I deliberately did not make this compulsory. There are 228 different actions and 811
distinct settings across them; a rule that strict, applied on a guess, would start
rejecting configurations that work — which would be a much worse bug than the one I'm
fixing. So actions opt in, and the report shows how many haven't (currently 208 of
them), so the gap is a number someone can chip away at rather than an invisible
absence.

**It earned its keep immediately.** The first time I ran the report against the live
system it found a *fifth* dead setting that the bug file never mentioned — and it's a
typo. The configuration says `add_protocol`; the code that would do that job reads
`add_protocol_if_missing`, and it belongs to a different action entirely. That one
matters, because without it a bare domain name goes to the scraper and the fetch just
fails. No amount of re-reading the bug file would have found it. A machine comparing
what the config claims against what the code declares found it in one run.

**Where it stands.** All of it is committed and the tests pass, including one I proved
would fail against the old code before I trusted it. It's gone to the review council
for a verdict. **None of it is live yet** — it needs two images rebuilt and rolled (the
scraper adapter, and the main chassis), and there's one database constraint that has to
be applied *after* the code ships, not before, or it would start refusing writes the
running code can't yet satisfy.

**One judgement call I'd flag.** I did not switch the two affected agents over to
actually crawling multiple pages. That's a real behaviour change to two agents I don't
own — one of which has no owner at all — under a data collection lane that's been
switched off since March. Getting that wrong at restart would be worse than the current
state. What they do now is *say so*: instead of silently fetching one page and
reporting success, they log that the setting cannot take effect and why. The honest
half is done; the behaviour change is somebody's deliberate decision to make, not a
side effect of my bug fix.

---

## 2026-07-28, evening (new session, picked this up from the handoff)

**First thing: the file I was handed was already an hour out of date, and the work
had moved on without it.** Worth saying because it is the normal condition here, not
a mishap — several sessions work this tree at once. I checked the live system rather
than believing the document, and found the review council had already been round
three times and **approved** it. So the main outstanding job on that handoff was
finished before I read the sentence saying it wasn't.

**You made a call and I recorded it.** The two agents that claim to read three pages
but only read one — you said leave them warning, don't switch them to crawling. I've
written that into the bug file and the handoff as a *decision*, not an omission, with
an explicit note telling the next person not to "finish the job" by flipping it. That
matters more than it sounds: in six weeks a deliberate won't-do and an overlooked gap
look identical, and someone helpfully fixes the one you chose.

**I checked the two things that were blocked. Both still are.** The vet data hasn't
restarted — the newest record is still 18 March — so the test that would close the
remaining bug still cannot run. And the payload alarm that came back "clean" after
the last deploy was clean because **nothing had been scraped at all**. Zero errors out
of zero attempts. That is not reassurance, it is an empty measurement.

**So I made something actually run — and it found a new bug.** I fired one real scrape
at our own site to give that alarm something to measure. No payload error, which is
good. But the reply only got through because the adapter had quietly thrown away the
last 3,800 characters of the page and stamped it *"full version in S3"* — when nothing
had been uploaded to S3 at all. The note telling you where to find the missing data
points at a file that was never written. **Four of the six live scrapers are set up
that way.** One of them is the vet verifier, and the thing it exists to extract — a
company registration number — is normally in a page footer, which is exactly the part
that gets cut. I've marked that last bit clearly as a suspicion rather than a finding,
because no vet page has been scraped since March so nobody can know yet. It is cheap
to settle the moment collection restarts.

**Underneath that was a measuring fault worth more than the bug.** The command in our
own runbook for checking that alarm reads the logs of *one* of three copies of the
service — and because of how the queue works, two of the three never do any work at
all. So it can report a perfectly clean log while the only copy doing anything is
failing. It picked an idle one today. Fixed everywhere we'd written it down.

**Then the config audit, which turned out not to be the job I was given.** The task
was "keep declaring config keys, 208 to go". I'd have spent the evening grinding
through them. Instead the first one I opened had already declared all its keys — which
made no sense — and the reason turned out to be that **the opt-in mechanism was
blocking the very adoption it was built to encourage.** Signing up required filling in
a field that means "settings, not references", so any component whose settings are all
references literally could not sign up without lying about its own configuration. One
component in 152 had managed it. That is not people being slow; that is a door that
doesn't open. I separated the two things, and the number went from 208 outstanding to
152 in one change — with no new claims about behaviour, because I only signed up the
ones where the answer was already proven. It's gone to the council.

**The honest bit I want on the record.** That is the second time in one day this piece
of work has shipped a fix that hid its own problem — first a filter that silenced the
report meant to catch it, now a gate that blocked its own adoption. I've logged both.
The pattern is that whoever builds a mechanism is the worst-placed person to notice it
is too expensive to use, because for them it wasn't.

---

**2026-07-28, later that evening.** Two things to report: the detector we shipped has
finally had something to do, and looking through the backlog it was built to expose turned
up a proper bug.

The detector first. Last night's note said it was live but had never run against anything —
which is a fair description of "we don't know if it works". That's changed: 48 workflow runs
since the roll, nine of them using an action we'd signed up, and every one of those carrying
the sort of configuration the check reads. It found nothing wrong. I want to be careful about
what that's worth, because the actions we signed up were deliberately the ones we'd already
proven clean, so "found nothing" was very nearly guaranteed. What makes it meaningful is that
the tests prove the thing does shout when there is something to shout about. So: it runs, it
would speak up, and it has nothing to say yet. There is still one gap I'm not papering over —
nobody has yet watched a real warning travel from a real run into the log. I've written that
down as owed rather than ticked off.

Now the bug, which I think is a good one. The remaining work on this thread is a list of 152
actions whose configuration nobody has checked. I picked at the most promising slice and
found that three of them had been **renamed in the code and not in the configuration**. The
code now looks for a setting called "pipeline"; the live configuration still says "domain".
Nothing reads "domain" any more, and — this is the part I'd want someone to notice — *nothing
anywhere sets "pipeline" either*. So those settings do nothing at all.

Here's why it has gone unnoticed for so long, and why I don't think anyone was careless. When
the setting is ignored, the code falls back to a built-in default. And in both cases the
default happens to be exactly what the configuration was asking for anyway. So the system
behaves correctly, every test passes, and the configuration file confidently describes a
choice that is not actually being made. I checked the real data rather than assuming, and
genuinely nothing is wrong today. But it's correct by luck. The moment somebody edits one of
those settings expecting a change, nothing will happen, and there will be no error to explain
why.

One related thing *is* wrong right now. There's an agent that writes drafts into a
human-review queue, and it sets a nicely worded caption for each one — "Grounded explainer
draft ready for review", plus the topic. It's using the wrong setting name, so that caption
is thrown away, and the code falls back to using the internal type name instead. Two items
are sitting in the review queue today labelled `grounded_draft_review`. Someone reviewing
them sees a database-ish label where a sentence should be. It's small, but it's the kind of
small that quietly makes a queue harder to work through.

I've written all of it up as bug 136 with the evidence and a fix order, and left it alone
otherwise — these are other people's agents, and changing a setting that currently does
nothing is a behaviour change, not a tidy-up. Worth flagging one trap in the fix, because
this workstream already fell into it once: the tempting way to make the warning go away is to
tell the system "yes, that key is fine". That silences the alarm and leaves the behaviour
broken. The fix is to make the config say what the code reads, not to teach the code to
tolerate a word it ignores.

I also nearly filed a false one. I had a fourth case lined up and was about to call it dead,
then found there's a legitimate back door where a step can pass values the code doesn't
formally declare. It didn't apply here — I checked — but if it had, I'd have reported a bug
that wasn't. Noted in the technical log, because the near-miss is the useful part.

---

**2026-07-29, morning.** The new build went out overnight, so the first job was checking our
work is still in it — it is, and the safety checks on the neighbouring bug are fine too. Then
I had a look at how much the checker has actually had to do. Overnight it saw 397 workflow
runs, 32 of them using something we'd signed up, across 16 different actions. It reported
nothing wrong.

Yesterday I said that "nothing wrong" was worth less than it looked, and having sat with it
I think I understated the problem. It isn't that the evidence was thin. It's that **we had
signed up only the things we'd already proved were clean**, so the checker could never have
found anything, no matter how long we waited or how much traffic went through it. We were
waiting for proof that was not capable of arriving.

So I stopped waiting and gave it something real to find. Yesterday's bug turned up two
actions with settings that genuinely do nothing. I've now told the system what those two
actions actually read — which is a small, careful change: it declares the truth, it doesn't
alter any behaviour. I checked that last part rather than assuming it, because it's the
whole reason the change is safe: the declaration is only ever used for reporting, and never
by the part of the code that decides what a step does.

The effect was immediate in the offline report. There's a section in it headed "unknown
keys" that has printed the word "none" every single time since we built it. It now names two
real problems. That section going from empty to populated is, I think, the moment this piece
of work started paying for itself — up to now it has been measuring its own coverage rather
than finding anything.

One honest caveat, and it's the same one as yesterday: the change is committed but it isn't
running yet. Code changes here only take effect when a new build goes out. When the next one
does, the first run of any of those six agents should produce the warning in the log — and
that will finally be the thing we've been owed for two days: not a test proving the checker
*can* speak, but a real run where it *did*. I'm not claiming that until I've seen it.

I've also sent the change for review by the automated reviewer panel, and I'm deliberately
not triggering a new build while that's running, because a build would kill the review
halfway through. So the order is: let the review finish, then get it live, then watch for
the warning.

What I've left alone on purpose: the six agent configurations that carry the wrong setting
name. Fixing those means changing what other people's agents do, and a setting that currently
does nothing is exactly the kind of thing where "just rename it" can have consequences
nobody expects. That's written up with a suggested order in the bug, for whoever owns them.

---

**2026-07-29, mid-morning.** You've answered the question I left open, and it turned out to
matter more than a footnote. The improvement loop being switched off is deliberate — you
stopped it while we're in a heavy development stage. I've recorded that as a *decision*
rather than a defect in three places, so that no future sweep files it as a dead scheduled
task and nobody helpfully turns it back on.

The reason it mattered: I'd been planning to prove our new warning works by triggering one of
the discovery agents. Those are exactly the agents you've quiesced. Firing one would have
pushed real work items into a queue you deliberately stilled, to buy myself a single line in
a log. That's a bad trade, and I'd have made it if you hadn't said.

So I went looking for a different route, and it turned out to be the better piece of work
anyway. Instead of asking "which agent can I poke", I asked "which of the actions we haven't
checked yet are actually *running*". Fifty-two of them had real traffic in the last day. Of
those, exactly one was carrying a setting its own code never mentions: the page builder
passes a `domain` value that nothing reads, on an action that runs about fourteen times a
day. It's the same broken rename as the other two — third site, same story, and this one is
on a live path.

That means the proof can now come from work that's already happening, on its own, without
touching anything you've turned off. I've made that change and sent it for review.

Where that leaves the honest scorecard: the checker is live and now watches three real
faults; the offline report names all three by name; and the thing I still haven't got is a
real log line from a real run. That last one is no longer blocked on a mystery — it's just
waiting for the next build to go out, and then for the page builder to run, which it does
several times a day by itself.

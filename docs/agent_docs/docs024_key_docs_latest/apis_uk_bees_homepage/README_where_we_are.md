# apis.uk — where we are

Plain-prose running log for the owner. Append only; newest at the bottom.

---

**2026-08-22 — the bees page, and the good news about the API.**

You asked for a page about bees on the apis.uk home page, without disturbing the
tools API that runs on the same domain. I took the second half first, because
that is the half that can break something that is currently working.

The short version: **nothing needs to change in DNS at all, and the API was never
really at risk from the DNS anyway.** But there *is* a real way to kill it by
accident, it is not the one everyone has been writing down, and I have now written
the real one down.

Three of our own documents say that when the bees page exists we will need to
"repoint the apex at its hosting — one record swap". That was true when it was
written and it is not true now. The apex of apis.uk is *already* being served by
the same Cloudflare worker that serves all the portfolio sites. What happened is
that a worker "route" was added for the bare domain at some point, and a route
overrides where DNS points — Cloudflare intercepts the request at its own edge
before it ever goes looking for the origin server. So the DNS entry for the bare
domain still says "send this to the island machine", and that instruction is now
simply never consulted. The upshot is that putting the bees page in the usual
place publishes it, with no zone changes whatsoever. The safest possible change
to the API's DNS is the one we are making: none.

I did not take that from the configuration files, because the configuration files
are exactly what misled the earlier documents. I asked the four hostnames
directly, and they give themselves away by *how* they fail. All four return "not
found", but the bare domain and the www one return it in the worker's words,
while the API hostname and a made-up subdomain return it with an empty body,
which is the island machine's way of saying it. Same status code, four names,
three different servers, all distinguishable.

**The thing that could actually kill the API.** It is not the DNS records — those
are per-name and independent, so the bare domain and the API subdomain were never
going to interfere with each other. It is the *worker route pattern*. Our routes
today name the bare domain and www specifically. If anyone ever adds a wildcard
route — the "everything under this domain" form — it would swallow the API
hostname too, hand it to the portfolio worker, which would look for a web page
that does not exist and return "not found". The API would be dead, no DNS record
would have changed, nothing would look wrong, and our own tidy-up script would
report success. Twenty-four other domains already have that wildcard route, quite
correctly, because they do not have an API living on a subdomain. This one does.
I have filed that as a landmine so the next person to run a "give every domain the
standard treatment" sweep is warned before they run it, not after.

Before touching anything I also checked the API was alive, properly rather than
lazily. Worth knowing: the API's front door returns "404 not found" and always
has — that is just how it is configured, and someone has previously read that as
proof it was dead and been wrong. The honest test is to call a real endpoint, so I
did, and it answered correctly. That is now the before-and-after control for
anything done to this domain.

**On the page itself.** You chose a home page only, as a personal enthusiast page
rather than a beekeeping guide or a conservation campaign, and I have set it up
that way and handed it to the build pipeline. It writes the content, not me —
that is the rule and it is a good one. I gave it a brief explaining that the
domain name is the joke (Apis is the genus the honey bee belongs to, so a domain
called apis.uk ought to be about bees), that the visitor is someone curious who
knows nothing much about bees, and that the page sells nothing, collects nothing
and asks nothing of anybody.

I also constrained it to build *only* the home page. Left alone the pipeline
would plan a whole site — an about page, a contact page, guides and so on — so
there is a specific instruction in place that says build one page and do not
invent others.

**One thing I spent real care on, which I think is the right call but you should
know about.** Bees are a subject made almost entirely of famous numbers: the
share of our food that depends on pollinators, the two million flowers in a jar
of honey, the miles flown, the tens of thousands of bees in a hive, the percentage
declines. Every one of them is repeated everywhere and none of them has been
checked by us. There is also a very well-known quote about having four years left
to live that is confidently attributed to Einstein and that he did not say.

So I have set the page up to assert **no quantities at all** — no counts,
percentages, distances, weights, temperatures or lifespans, whether in digits or
spelled out. That sounds like a severe restriction and I do not think it is,
because none of what makes bees genuinely remarkable needs a number: a returning
forager telling the others where she has been by dancing on the comb, direction
read against gravity; wax worked into a shape that wastes nothing; a colony
dividing its work by age so a bee's job changes as she ages; the fact that most
bees are not honey bees at all but solitary insects nobody notices. If we later
want a specific figure on the page, the way to get it is to look it up properly,
record where it came from, and then it is allowed. The friction is deliberate.

Worth saying plainly: I tested that restriction rather than assuming it worked,
and **it failed three times before it passed.** One rule was written in a way that
looked correct, was technically valid, and would never have matched anything — the
kind of error that reports "all clean" for ever. Another missed "two million
flowers" entirely, which is the most repeated bee statistic there is, because it
was only looking for digits sitting directly against the word. Those are now
fixed and there is a test holding them in place.

**Where it stands.** The site is seeded and the build was submitted at lunchtime.
These take a while to come through — anything from a quarter of an hour to half
an hour just to start under normal load — so the next step is simply watching it
build, then checking the page that comes out and confirming, separately, that the
API still answers. Those are two independent facts and I will check them as two.

**One thing you may want to decide.** The bare domain used to be part of a traffic
probe — a passive listener recording who was still asking for apis.uk and what
they wanted, which was due a read on the 8th of August and has not had one. The
apex arm of that probe has in fact been silently inactive since the worker route
was added, so the bees page is not taking anything away that was still working.
But the wildcard arm is still listening on every other subdomain, and the log has
never been read. It might be worth reading before we lose interest in it — that is
a separate small job and I have not done it.

---

**2026-08-23 — I put your infrastructure on a public page. That was my mistake, it is off, and here is what I have changed so it cannot happen again.**

You found four sentences live on apis.uk telling the world that the domain also runs an
unrelated technical service on another hostname. You called it a serious misstep. It is,
and it was mine.

It is worth being exact about whose mistake it was, because the answer decides what gets
fixed. The framework did not invent that sentence. **I asked for it.** The brief I wrote
on the 22nd said, in as many words, that a short line acknowledging the other service was
welcome so a developer arriving by mistake would not be confused. The framework did what
it was told, and it did it four times.

The reasoning behind that instruction was wrong in a way I want to name properly. You
asked me to protect the API. I turned that into telling everyone the API exists. Those
are close to opposite things. **A constraint about protecting something is not permission
to describe it**, and the confused developer I was writing for does not exist in numbers
that would justify putting infrastructure on a public page.

Three things made it worse than a single bad sentence.

First, I said "somewhere unobtrusive, once". That names no particular place, so the writer
put one in every section. An instruction that cannot say which part of the page carries it
is not a light touch. It is an instruction repeated at every opportunity.

Second, it spread. By the time I went looking, that instruction had propagated into seven
different planning documents the system keeps for this site: it had become a stated fact
about what the site is, an item on its constraints list, a footer element in the strategy,
and — worst — an **acceptance criterion**, meaning a check could in principle have failed
the page for *leaving it out*. Deleting it from where I typed it would have left six
copies and a test demanding it back.

Third, and this is the one I am least happy about: I checked the wrong things. After the
build I confirmed the page count was one, and I confirmed the API still answered. Both
were true. Neither was relevant. **I never read the page.** The check that would have
caught this in ten seconds is the one any person would have done first.

**Where it stands now.** The sentences are gone from the live site, verified by fetching
the page rather than by trusting a status. The API still answers correctly — I checked
that separately, because "the page changed" and "the API still works" are two facts and I
am treating them as two from now on.

Removing my instruction only stops the system being *asked* for that sentence. So I have
also made it refused: the site now carries an explicit ban on mentioning anything else
running on this domain, in any phrasing, plus a flat prohibition in the writing rules. A
deleted instruction is invisible to whoever reads this next. A ban is not. I have logged
the whole thing in the two fleet-wide records we keep for exactly this, so it is not just
buried in this lane's notes.

Two things surprised me while fixing it, both worth telling you because they are the sort
of thing that bites again.

The first: I removed the sentences from the page's stored content, confirmed it was clean,
re-rendered — and got back a file identical to the byte. The renderer had not used the
content I cleaned. It used a *cached* copy of the rendered version sitting in another
column. So I had cleaned the source and the machine was reading the cache. Fixed both.

The second: my clean-up left an orphan. It removed the sentence naming the service, but
left the next sentence, which said "state it once, style it quietly, never present it as
documentation". That fragment contains none of the words I was searching for, so every
check I ran said the file was clean. I only caught it because I happened to rebuild that
document from scratch and compared it against the stored one for an unrelated reason. That
is luck, and I would rather say so than dress it up.

**On the copy sounding like AI wrote it — you are right, and it has the same root.**

You pointed at "worth sitting with" and the "not just" framing. Another lane here has
already done the work on why our copy comes out negatively framed, and their finding is
that it comes from the site's own stated identity, not from the model. That reproduced
here exactly. Four of the five things recorded as this site's distinguishing features were
written as "X, **not** Y": reads like a friend *not an institution*, covers things deeply
*rather than* skimming, *no* agenda, *nothing* to sell. Those are a faithful summary of my
brief, which was mostly a list of things the page must not be.

But the sharper cause is worse and simpler. The system keeps a short list of example
sentences for the writer to imitate. Four of the five examples were written in exactly the
style you objected to — "A returning forager does not simply arrive back at the hive — she
announces where she has been." The writer copied its examples. **Showing a writer examples
in the style you do not want beats any rule telling it not to.**

The good news is that the framework already has proper controls for all of this — voice,
sentence style, writing rules, things to avoid, example phrases — and I had simply left
them at whatever the classifier first guessed. We also already have a house
"de-AI-ify this copy" prompt, built by reverse-engineering an edit you yourself judged
more readable. I have now put that to work: the example sentences are rewritten to state
things plainly and lead with the fact, the house rules are in the writing rules, and the
specific tells you named are on the banned list.

Then I handed it back to the framework to rewrite the prose. I am not writing the copy
myself — that is the rule, and it is the right one. That rebuild is running now, and I
will check the result by reading it.

## 2026-09-03 (evening) — the subjects feature is switched on, and the home page survived a near miss

A big day, told plainly. The per-section subjects work — the thing that stops a page with six
similar sections writing the same section six times — is now fully switched on. It took three
tries to get the wording right: you rejected the first drafts as sounding like AI, picked one,
then when we showed you what its fixed opening sentence did to real data ("You'll want to know
Brief description of the sister-site relationship…") you dropped the frame and chose the simple
version: whatever line we write for a section is printed exactly as written, and every other
section on the page is told about it. You read the final text and said "yes", and it went live
this evening. The planner's instructions were updated to match: from now on it writes each
section's brief as the actual opening line a reader would see, in the site's voice.

Separately, the home page had a near miss. The site's internal plan still described the page as
it was BEFORE the illustrated redesign, and an ordinary rebuild would have quietly put the old
version back — two other sites' pages have already lost pieces this exact way. Another session
spotted it, we double-checked everything, and corrected the plan to match the real page. The
illustrations and locked sections were never at risk of being edited; the risk was the page's
own table of contents flipping back, and that door is now closed.

What's left on this: writing the six actual opening lines for the home page's six illustrated
sections (they're your page's future first sentences, so we'll take care over them and you may
want to see them), and your two paused rewrite requests (the swarm and pollination sections) can
now go ahead whenever you decide how to add sections to a locked page. Nothing needs your
attention urgently; the decisions list from this afternoon still stands.

# Consolidation and divergence — where we are, 2026-07-27

*First summary in this series. Anchor: `features_open/024`. Mechanism findings:
`bugs_open/108` and `architecture_review/…architecture_seat.md` §8.*

## What we're trying to do

We intend to run thousands of domains — the working list is about 1,625 — from
one platform, and we are at roughly fifteen live sites. Getting from here to
there is not mainly a hosting problem or a cluster problem. It is the problem of
not building the same thing repeatedly in slightly different ways.

The aim of this work is to answer two questions the owner put directly. Which of
our duplications actually block that scale, as opposed to merely being untidy?
And what would notice the next duplication happening, early enough to matter?

## Where we've come from

The question arrived from a real near-miss rather than a worry. On 24 July a
design here specified a new public service on the island machine — its own
database, its own AI key, its own rate limiter, its own web route. On 25 July a
different thread shipped `tools-api` to that same machine, built from the start
to serve many tools across many sites, with all four of those things already in
it. The design would have been a second copy of the lot, on a box with one
processor and 2GB of memory.

Nothing caught it. Not a hook, not the council, not a lint. The owner caught it
on 26 July by asking how the dossier would fit with the other tools.

That was not the first warning. The VM estate thread had already measured the
estate as two forks of one script sharing 61 lines and differing on 614, and had
written down that a third divergence was being born and *"left alone this becomes
four."* This was that fourth, and it was being born in a different thread that
had not read theirs.

## What we've done

We chased the cause rather than just fixing the instance, and the cause turned
out to be worth more.

The finding that reframes it: the prior-art search **was done, and was correct**.
On 24 July `tools-api` genuinely did not exist. The failure is that every review
mechanism we own judges a proposal against a snapshot, and nothing ever looks
again. This repository moves fast enough — around 1,500 commits a week — that a
design document outlives the world it was written in. Being more careful does not
fix that; only something free and repeatable does.

Three structural findings came out of it, all checked against the live system.
The decision lived in a document, and our review machinery refuses to read
documents — a correct decision on cost grounds, which means a document that
decides to build a new service is refused by the only mechanism that would
object. The council seat whose entire job is asking "does this already exist?"
asks its code questions into thin air, because the step that answers them was
deliberately left off that council. And the search index behind that seat records
what functions are *called* but never what they *do*, so a search for any web
address, table name or piece of text comes back empty — and the seat policing
claims of absence is handed manufactured absence. Its own documented example
cannot work. That is filed as `bugs_open/108`, along with a second defect: the
index reports itself fresh while describing code 667 commits old, because it
measures when it was last written rather than how far behind it is.

We also measured the fix rather than proposing it. A check that tells a document
proposing a new program which programs already exist fires on 0.67% of the last
1,500 commits — normal for our checks — and fires on the exact commit that opened
the workstream that made the mistake, two days before the human noticed.

And we produced a consolidation programme with an opinion, including an item we
recommend *not* doing.

## Where we are now

The answer to "should this be a council member, the diagnosis loop, or the
architecture council" is none of the three, and each for its own reason. A
council member is the wrong instrument because "does this exist" is a question a
search settles, and we already have two seats holding that job — they need their
instrument repaired, not company. The diagnosis loop is built to narrow onto one
bug and halts when scope widens, so it can never run a survey. And the
architecture council already exists as a live thread that the owner was ruling on
the same day; our measurements went into it as evidence rather than becoming a
fourth overlapping mechanism.

On the consolidation itself, the number that matters is that nine of our 296
pieces of pipeline machinery exist for two sites out of a thousand, and five of
those nine shipped in a single week. That is the thing that does not reach 1,600
domains. We already do it correctly elsewhere — for company data a new industry
is a row in a table, not new code.

Two capability gaps sit in front of the next piece of work. There is no email
sender anywhere in the code we build and deploy; the only working one lives in
idea.uk's box outside the build. And the rate limiter guarding our only public
API is the weakest of the three we have, while the best one is unreachable in the
documentation tree.

One item is a recommended won't-do, recorded so it stops looking available: eight
copies of a small health-check server look like the tidiest possible win and turn
out not to be identical at all, so merging them means eight risky changes to what
tells Kubernetes our services are alive, for no benefit at any number of domains.

There is also a live change since this work began. The chassis rolled to
v1.0.1173 at 13:45 today, and all six of the gripper dossier's new actions are
now in the running binary, verified by name against the pod with a negative
control. **Everything described in our own notes as "inert until the next roll"
is no longer inert.** That is a small, pointed illustration of this summary's own
finding: a fact recorded accurately in the morning was false by the afternoon.

## Where we're going

Three things, in order.

Apply the gripper seeds and run the pilot's three end-to-end tests, including the
deliberately induced failure, now that the image carrying the code is live. The
owner has ruled: finish the pilot as it stands, then generalise the scoring
engine into something config-driven, rather than abstracting from a single
example before it has worked once.

Build the divergence check, since it is measured and cheap, and feed the result
back to the architecture thread that owns the wider question of a review seat.

And take the two capability gaps — the mailer and the request guard — into shared
code *before* the dossier's public half is built, because that is precisely the
moment the estate would otherwise fork again. That half now belongs inside
`tools-api`, which another thread owns and which has an open bug against its
error handling, so it is a conversation before it is a commit.

One thing is unowned and should not stay that way. The staged site maturity
ladder is our stated method for lifting the whole portfolio — named rungs, worked
examples, a site climbing one step at a time instead of facing a cliff. It has no
directory, no plan and no owner. It is flagged here rather than adopted, because
quietly starting it inside this work would be the exact behaviour this programme
exists to stop.

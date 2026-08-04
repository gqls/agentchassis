# SUMMARY — webdesign.uk build service, 2026-08-04

*The first summary in this lane's series. Current state only. The chronology lives
in `README_where_we_are.md` and the technical log in `NOTES_...`.*

---

## What we are trying to do

Sell complete websites to small British businesses at a fixed price, built by the
system we already own and checked by us before anyone pays.

The offer is one price, twelve hundred pounds, paid once. No VAT, because we are not
registered for it. The customer gives us a domain and a sentence or two about what
they do. We build the whole site, the pages, the words, the design, and put it on
their own domain. They see it finished on a private link before they decide, and up
until the moment they accept it they can walk away and get everything back. After
they accept it, changes they ask for are charged as work, and anything that is our
own mistake we fix for nothing.

The reason for building it this way is that we already have the machinery. The
framework builds sites now, for our own portfolio. The question this project answers
is not whether we can build them, it is whether anyone will pay us to.

## Where we have come from

This started as a plan on the twenty-eighth of July and spent its first week being
argued with rather than built, which was the right order.

The price moved from a cost-plus calculation to a quality-based one. We measured
what a build actually costs us in model time, and then deliberately set that aside,
because pricing a website off its compute cost would be like pricing a house off the
cement. Twelve hundred is what the work is worth, not what it costs.

The guarantee took a while to get right. A money-back promise and a charge-for-changes
policy pull against each other unless you pick a single moment where one stops and
the other starts. That moment is acceptance. Before it, everything is refundable and
nothing is chargeable. After it, the site is theirs and changes are work. One hinge,
and both halves hang off it.

We also settled that the previews would live on a short domain, ugg2.com, and spent
some time planning a server to host them on, complete with certificates and renewal
timers. That turned out to be unnecessary. The sites we build are static files, so a
preview needs no server at all. A single wildcard setting on the domain does the
whole job, and we proved it works. That deleted a sizeable piece of planned work.

## What we have done

The plumbing is finished and proven. Both domains are set up, previews work end to
end, and the whole thing was tested with made-up addresses that had never existed to
confirm it works for any customer we might get.

We found and closed a genuine hazard along the way. For a period, webdesign.uk was
quietly serving the idea.uk shop, and because that software does not check which
domain it is answering for, somebody could have placed a real order through the
wrong address. It is now behind a holding redirect pointing at a dead address, so if
anything ever fails it fails safely rather than selling something.

The shopfront itself has been written, and then written again properly. The first
version I built by hand, and that was a mistake worth recording. We sell
framework-built websites. A hand-built shopfront demonstrates nothing, and it also
skipped every safety check the framework applies, including the one that stops
invented claims reaching a page. It looked fine. Everything I checked about it
passed. That is exactly the problem: I was carefully verifying the wrong thing. It
is now a standing rule that every site goes through the framework, no exceptions,
and the reason is written down where the next person will find it.

So the site is now being built the right way. The instructions the framework works
from carry the corrections you asked for, and they are enforced rather than merely
requested. The em dash is banned outright. The line about a person checking the site
before you see it is banned, along with the whole family of phrasings that make the
work sound like a template getting a glance. The three-to-four day timescale and the
price are recorded as facts you have attested, which is what allows the writer to
state them at all. Fourteen things are banned outright, mostly the things a new
business does not have and should not pretend to: client numbers, testimonials,
awards, years of experience.

## Where we are now

The build is running and has got most of the way. The research, the strategy, the
brief, the site plan, the design composition and most of the images are done.

It cannot finish, because of a fault in the platform itself that appeared on Sunday
evening and is affecting other sites too, not just ours. Another team filed it this
morning; our failure is a third example, and a useful one, because ours is a brand
new site rather than an existing one, which rules out some of the possible
explanations. I have added that evidence to their file rather than starting a
competing investigation.

I have deliberately not worked around it. The obvious workaround is to hand-build
the page, and that is precisely what we have just agreed we do not do. So we are
properly blocked rather than quietly cheating, which is the right place to be.

Nothing is publicly broken. The holding redirect means the half-finished site is not
visible to anyone.

One other thing is in the way: Cloudflare has locked me out of making changes,
because the access token is tied to particular internet addresses and ours moved.
That does not block the site build, and the two changes needed to put the site live
can be done by hand in the dashboard in about a minute.

## Where we are going

Four things, in order, and only the first is stuck.

Finish the site once the platform fault is fixed, then read the finished page
against the rules to confirm the em dash and the phrasing really did stay out. Then
put it live, which is two clicks.

Then the box. It goes on Mythic Beasts, and we already run one machine on exactly
the right pattern, so we are copying something that works rather than inventing.
Two processor cores, four gigabytes of memory, and notably no need to pay for a
public address at all, because nothing will connect to it from outside. It will be
reached through a tunnel instead.

Then the chat box, which is the actual point of the exercise. It is a real
conversation, not a form, and that is what makes it expensive: every visitor costs
us money whether or not they buy. So the limits on spending come first and the chat
comes second, not the other way round. There is a specific trap here that we have
been bitten by before, where a per-person limit sitting behind Cloudflare silently
becomes one shared limit for the entire internet while still appearing to work.

And then the two halves get joined: the site stays as static files, which is free
and cannot really be attacked, and only the chat lives on the machine. That way the
thing facing strangers is as small as possible.

What we still need from you is the hosting access for that machine, and eventually a
number for what a change costs after someone has accepted their site. Neither is
urgent this week. The price and the VAT question are both settled and need nothing
further.

# Where we are — the contact page that invented a phone number

Plain prose, append-only, newest at the bottom.

---

## 2026-08-02, morning — what was wrong

Eight of our live sites had a contact page telling visitors the business is open
"Monday – Friday, 9am – 6pm". Nobody ever said that. No record anywhere holds
those hours. The component that draws the contact page simply made them up
whenever the site had not supplied any, and printed them in exactly the same
style as the real details, so there is no way for a visitor — or for us — to tell
the invented line from the true one by looking.

On vetcomparison.uk it went further. That page was built two days ago, and it
publishes a phone number: `+1234567890`. It is not a phone number at all, it is
the placeholder someone typed into the template years ago, and the site has been
serving it as a way to contact the business.

This was already written up as bug 140 back on 29 July, when it affected six
sites. Nobody had picked it up. By this morning it was eight, which is really the
whole argument for fixing the machine rather than the eight pages: two more sites
walked into it in the four days the ticket sat there.

## What turned out to be the real story

I expected this to need a decision from you. "Should a component invent a
plausible default, or show nothing?" is a judgement call, and it affects eight
sites that other work-streams own, so it did not feel like mine to make.

It is not a judgement call, and that was the useful discovery of the morning.
Every component carries a little contract describing where its data comes from
and what to do when it is missing. This component's contract **already says the
right thing** — for the phone, the hours and the address it says, in as many
words, *skip this field if you have not got it*. The template just ignored its
own contract. So the change is not a new policy; it is making the thing obey the
rule it already publishes. That needs no ruling from you.

The same contract, incidentally, has a proper way to say "here is a sensible
default" — the section heading uses it. So the system already knows the
difference between inventing a **fact** about a business and providing a
**label** for a button. Only the template had lost track of it.

## A second thing, which is arguably worse

While reading the template against its contract I found that it was looking for
the heading and the introduction under the wrong names. The sites are all
supplying them properly — "Get in Touch", "Contact Darts Online", "Reach us
directly", "Contact VetComparison.uk" — and the template was looking for
something else, finding nothing, and falling back to a generic "Contact
Information" every time. Every one of those eight bespoke headings, and every
intro paragraph, has been written and then silently thrown away.

Our own quality checker noticed this and recorded it on **18 May**. The finding
sat in a database column and nothing ever read it. That is a pattern we already
have open tickets about, so I have noted it and not gone chasing it here.

## What I changed

Three things, in order of how much they matter.

**The template now obeys its contract.** Each contact card appears only if the
site actually supplied that piece of information, and the four invented values
are deleted outright. A detail nobody supplied can no longer appear on a page.
While I was in there I pointed it at the right names, so those eight real
headings and intros will start appearing, and turned on the address field, which
was defined but never drawn.

**The detector that was supposed to catch this has been taught our own
placeholders.** We have a checker whose actual job is finding fabricated contact
details on pages — it raises alerts literally titled "Fabricated contact info".
It was looking for the textbook fakes: numbers beginning 555, `example.com`, "123
Main Street", "Lorem ipsum". It did not know a single one of the placeholders our
own component library ships. Across every page in the fleet its nine patterns
found one thing; the fabrications it could not see numbered nine. It was blind to
its own platform's output.

**And a new standing check reads the library itself.** The problem with a list of
known fakes is that the list is only ever as good as somebody's memory — a new
component with a newly invented default is invisible until a human thinks to add
it, which is exactly how this survived from the beginning. So the new check does
not use a list. It reads every live component and asks whether its fallback is
inventing a *fact* — a phone number, an address, a price, opening hours — as
opposed to supplying a *label* like "Read more", which is perfectly fine and
which the library is full of.

## Two things that were wrong in my own work, both caught before they shipped

The new check's first run accused a component of hard-coding a domain name. It
looked bad. It was wrong: that component prints "designed and built by
fundamentallyai.com", as a link when it has a URL and as plain text when it does
not — the same words either way, inventing nothing. I fixed the rule generally
rather than adding an exception.

Then, having tightened the rule, I ran it against half a dozen made-up examples
to check it still caught real problems. It missed one: a component inventing
"Weekdays 8am to 5pm" sailed through, because I had written the rule to look for
day names like "Monday". Had I not run that check, I would have shipped a
safeguard that misses the most natural way of writing invented opening hours.
Both are written up properly in the notes.

## Where this leaves the eight sites

The template is fixed and live now — it is configuration, not code, so it took
effect immediately. But the eight contact pages already exist as finished pages,
and they keep serving what they were built with until each one is rebuilt. So
today the fabricated hours are still on the live sites even though the thing that
caused them is gone.

That is why I have left the ticket open rather than closing it. The rule here is
that a bug is only closed when it is no longer reproducible on the live fleet,
and it still is. The pages correct themselves as they rebuild, and I would rather
prove that on a real page than assert it.

**One thing worth your attention, which I have deliberately not touched:** the
same number, `+44 (0) 7934 524 911`, is published as the contact number on six
different businesses. I traced it before assuming the worst — it comes from the
site records, it is your own number, and these are your own portfolio sites, so
it is real and correctly propagated, not another fabrication. But six businesses
sharing one number is a thing you may or may not want, and that is a call for you
rather than a defect for me.

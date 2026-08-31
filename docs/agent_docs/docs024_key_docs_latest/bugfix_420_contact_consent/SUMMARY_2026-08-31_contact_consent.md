# SUMMARY — 2026-08-31 — a billing address is not consent to publish

## What we're trying to do

Make sure the email address a customer pays with can never become the address their website shows
to the public. Those are two different facts with two different permissions, and the platform had
been treating them as one.

## Where we've come from

On the first paid build, the owner ordered as customer zero — and his own personal address was
published in the footer of every page on the site. He saw it and had it taken off the same day.

It was cheap only because it was his. On order two it is a stranger's, and it would be published
the same way, for the same reason.

The delivery lane cleaned the site, filed the class defect, and left two things for whoever took
it: the owner's ruling on what the framework should do in general, and a condition — find every
piece of code that reads that column before moving the contract.

## What we've done

We got the ruling and built to it: **when nobody has been asked what contact details a site
should show, the site shows none.** The address a customer pays with is now used only to deliver
their site and never reaches anything that renders a page; a published contact has to come from
someone explicitly answering the question.

The second half of the defect is the one worth understanding, because it explains why deleting
the address was never going to be enough. The same intake step had also written "Enquiries reach
«the address»" into the site's register of **facts it is permitted to state** — the register that
exists to stop the system inventing things. The address wasn't merely stored; it was *licensed*.
A later rebuild could have republished it and our own honesty checks would have approved it.

That wasn't theoretical. We could show it had already spread: twelve minutes after the order, the
address appeared in a second document nothing had copied it into directly, and by ruling out
every other possible source we established the register was the only one it could have come from.

Two things in the original account turned out to be wrong, and both made the work cheaper. The
delivery system does not actually read that column in code — the coupling was a convention, not a
dependency, so separating the two uses changes a recipe rather than a chain. And the footer was
already guarded against showing an empty address; the fault was never a missing guard, it was the
value the guard was letting through.

We also found the platform inventing contact details of its own. Where a site had no email, two
places quietly synthesised one — "info@" plus the domain — and published it as though the business
had chosen it. That is now removed, because otherwise "this site publishes no contact" would have
been quietly false.

And one safeguard arrived free. The check that hunts for made-up email addresses on pages had
been treating whatever sat in that column as automatically legitimate — which is precisely why the
leak was invisible to it. Now that the column no longer holds the payer's address, that same check
will **flag** it if it ever reappears.

## Where we are now

Committed, with the council, and it takes effect at the next chassis build. No database change and
nothing to migrate — the delivery address already had a proper home in the original order record,
and adding another contact-shaped column to the table every page-render touches would have been
the kind of thing that gets misused within a week.

The half that asks the customer the question lives in the intake chat, on the box, not in this
system. Our side accepts the answer whenever that starts arriving; until then every new site
publishes no contact details, which is the safe direction.

One live hazard until it ships: the address is still in the original order record, and the normal
way to retry a build reads from there. Because the site's email field is now empty, a retry would
refill it — the guard that protects a human correction works backwards when the correct value is
"nothing". Written up where someone about to do it will see it, because every check you would
think to run first reads clean.

## Where we're going

Read the council verdict and act on it. After the roll, run the order-two rehearsal: an order
whose payer address differs from anything the brief asks to publish, then check the **served
pages** — not the database — because during the incident four separate database sweeps each read
clean while a fifth store was still serving the address. Every false "all clear" that night was an
incomplete list reading as a result.

Then the intake chat needs to start asking the question, which is a conversation with the lane
that owns it.

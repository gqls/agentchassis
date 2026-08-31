# Where we are — your email address on a customer's website

*(Plain-prose log for the owner, append-only, newest at the bottom.)*

## 2026-08-31, evening

What went wrong is simpler than it looks. We had one box in the database labelled "email", and we
were putting two different things in it: the address we use to reach the person who paid, and the
address the finished website shows to the world. Nobody had ever separated them, so when you paid
for boxingonline as your own first customer, your personal address went into the footer of every
page on the site.

You've already had that cleaned off. This was about making sure it can't happen to customer two,
who won't be you.

There was a second half, and it's the one I'd want you to know about, because it's the reason
deleting the address wasn't enough. When the order came in, the system also wrote down "Enquiries
reach «your address»" in the site's register of **facts it's allowed to state** — the same
register that stops it inventing things. So the address wasn't just sitting in a field; it was
*licensed*. Any future rebuild could have put it back, and our own honesty checks would have
approved it, because as far as they were concerned it was a verified fact about the business.

And that wasn't theoretical. I could show it had already spread: twelve minutes after the order,
the address turned up in a second document that nothing had copied it into directly. The only
place it could have come from was that register.

So the fix is two things. The address you pay with is now used **only** to deliver your site — it
never reaches anything that renders a page. And the address a site *publishes* now has to come
from someone actually being asked "what contact details should the site show?". You ruled tonight
that if nobody asks, the site shows none, and that's what I built.

The question itself lives in the chat on the box, not in this system, so that half is somebody
else's to add. I've built our side to accept the answer whenever it starts arriving, and until it
does, every new site simply publishes no contact details. That's the safe direction, and it
matches your ruling.

Two smaller things worth telling you. First, I found the system was inventing addresses of its
own: where a site had no email, two places quietly made one up — `info@` plus the domain — and
published it as though the business had chosen it. That's now gone; a business with no contact
details shows none, rather than a fake one. Second, we got a safety net for free: the check that
looks for made-up email addresses on pages used to treat whatever was in that box as automatically
legitimate. Now that the box no longer holds the payer's address, that same check will **flag** it
if it ever reappears, instead of waving it through.

One live warning while this waits to go out. The address is still sitting in the original order
record, and the normal way to retry a build reads from there. Because the site's email field is now
empty, a retry would refill it — the guard that normally protects a correction works backwards
when the correct value is "nothing". So boxingonline mustn't be re-run until the fix ships. I've
written that up where someone about to do it will see it, because every check you'd think to run
first reads clean.

Where it stands: committed, with the review council, and it takes effect at the next chassis
build. No database change and nothing to migrate.

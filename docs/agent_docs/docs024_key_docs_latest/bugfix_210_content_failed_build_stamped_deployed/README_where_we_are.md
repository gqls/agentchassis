# Where we are — bugfix 210 (plain prose, append-only, newest at the bottom)

**2026-08-08.** Picked up bug 210: when a page rebuild's content generation fails, the system
skips the assembly and the deploy (correctly — the live page is untouched), but then stamps the
page "deployed" anyway, because the checks only look at the *previous* build's output, which is
still there. So a page that asked to be rebuilt, and wasn't, now claims it was — and nothing
will ever try again. The page quietly serves old content forever.

Checked nobody else is on it (the thread that found it closed its own bug yesterday and
explicitly left this one for a future session). Confirmed the hole is still in the code today.

The plan, in one breath: refuse the false "deployed" stamp whenever the assembly was skipped,
put the page honestly back in the "needs rebuilding" queue, and let the system retry — but only
three times. After the third failure, park the page behind a visible "needs a human" ticket so
the retries stop costing money and a person decides. A successful rebuild clears the ticket
automatically. Every refusal is also written to the error log, which finally gives us a count of
how often this actually happens (today nobody knows).

Off to the review council next, then implementation.

**2026-08-08, later.** The fix is built, tested and committed. Three details worth saying out
loud. First: the retry cap is three — after the third failed content generation for a page, we
stop paying for retries and raise a ticket a human can see; a later successful rebuild clears
the ticket automatically. Second: the "park" ticket deliberately sits on the same queue slot the
build system uses for that page, which is what actually stops the retries — but it also means a
sibling team's tool-recreation requests for that page are held while the ticket is open; they
have been told, in their own notes. Third: we found and fixed a small inconsistency next door —
a cancelled ticket used to free the slot for one producer but not the other; now both agree.
Awaiting the council verdict and the next build roll.

**2026-08-08, evening.** The council approved the change (first round, four advisory
objections, none serious; each is answered with evidence in the notes). One of them earned
its keep: checking a reviewer's worry, we found 25 old "cancelled" build requests sitting in
the queue, and our tidy-up means eight pages on dartsonline.com — muted by someone on 20 July
— will be rebuilt (at LLM cost) the next time that site is replanned. If that mute was
deliberate and should stay, someone needs to re-mark those eight items as "won't fix" (which
stays muted under both the old and new rules) before dartsonline is next replanned. That is an
owner/operator decision — the eight are listed in the notes; the query to find them is in the
runbook.

**2026-08-08, after the roll.** The fix is live on the whole fleet and verified on the running
binaries. Nothing has fired yet (correct — the counter starts at zero and now measures how often
this bug really happens). One correction to my last note: it is SEVEN dartsonline pages that the
mute-release affects, not eight — I miscounted when summarising and the re-check after the roll
caught it. The seven are: brands, brands-index, grip-styles, guides, product-detail, shop and
shop-index, all muted on 20 July. They rebuild only when dartsonline is next replanned, so
there is time to re-mark them "won't fix" if the mute should stand. One more page, vonc.com's
"provocation", is owner-flagged rather than rebuilt — it raises a review ticket, costs nothing.

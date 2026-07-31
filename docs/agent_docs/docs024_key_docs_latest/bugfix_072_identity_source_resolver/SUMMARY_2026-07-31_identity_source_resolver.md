# SUMMARY — 2026-07-31 — the contact details were never lost, nothing could see them

## What we're trying to do

Make the block that shows a site's contact details — phone, email, address —
actually appear on the contact pages of our sites. It has been missing from most
of them for months, silently, and nothing reported it. The trigger was concrete:
the owner supplied a phone number for fundamentallyai.com on 24 July and it never
showed up.

## Where we've come from

Someone diagnosed this on 25 July and filed `bugs_open/072`. Their explanation was
crisp and well-evidenced: the component that displays contact details asks the
system for `identity.email`, while the agent that researches a site writes
`identity.contact.email` — a **flat** address read against a **nested** store. And
because that field is marked "needs human review" when missing, the whole block is
withheld rather than partially rendered. They proved it with a fleet-wide
discriminator: the block appeared on exactly the five sites that had the flat key,
and on none of the eight that didn't. No exceptions in either direction.

Two people contributed to the file after that, one correcting the other about
whether a sibling bug shared the same cause. The proposed fixes were to repoint
the component at the nested address, or to teach the resolver a nested fallback.

Nobody built either. The bug sat for six days.

## What we've done

**We measured the proposed fix before building it, and it fixes nothing.** Putting
all three candidate stores side by side, per site, shows the nested `contact`
object exists on 14 of our 15 sites and its **values are empty on exactly the 8
that fail**. So repointing the reader from one empty shelf to another resolves on
**0 of 8 sites**. The discriminator was real and its causal reading was inverted:
the sites that worked were simply the sites that had the fact at all — somebody
had typed it in by hand.

**We found where the data actually is, and the bug file already said so.** It is in
the site's own record: `sites.email` is populated on 12 of 15 sites, including 5 of
the 8 that fail. The original file records this in a single sentence about the
owner's phone number — *"written only to `sites.phone`, which no component reads"* —
and treats it as a footnote to a workaround. That sentence was the root cause.

**We established it was the third path to need the same fix.** Two other code
paths that render pages already read those columns, one of them changed
specifically to do so with a comment saying it now makes "both render paths agree".
The page *planner* was missed both times. That turned the change from "invent a
resolver fallback" into "bring the last path into line with the two that already
agree" — better founded, and it told us exactly which columns to read instead of
guessing.

**We put the diagnosis through the diagnosis loop before writing code**, because we
were contradicting a filed explanation. CONFIRMED on the first iteration, in about
four minutes, having fetched its own evidence — a live vetcomparison row with the
email present in the site record and absent everywhere else.

**We shipped a bounded fix**: after a declared path misses, try the writer's nested
shape, then the site's own record. The literal path is tried first and always wins,
so nothing that resolves today changes value — and that property is a test, not a
promise. It is registered as a platform seam (PBP-026) with its landmine and its
open question, in the same commit as the code.

**The council approved it at round one** with four advisory objections. Two were
actionable and both improved the change. One seat argued the nested-shape half was
unnecessary scope — correct from the data, and we answered it with a measurement
rather than an argument: the action designed to copy nested contact details into
the site record is wired into **zero** live agents, so nothing else reads that
shape and a new site would resolve nothing without it. The other asked for a
queryable record of which fields changed provenance rather than just a log line;
we added one.

We also caught ourselves three times, all logged: believing the filed remedy for
ten minutes; counting 12 of 29 sites because `sites` holds 14 empty internal pool
rows; and nearly filing a fresh bug for 74 dead source paths that `bugs_closed/018`
had already diagnosed in July.

## Where we are now

The fix is committed, tested, diagnosed-confirmed and council-approved. It is Go,
so **it changes nothing on any live site until the chassis is rebuilt and rolled**.
The ticket therefore stays in `bugs_open/` — this repo's bar for closing is *fixed
and live*, not *fixed and committed* — with a banner at the top so nobody
re-diagnoses it. We did not roll it ourselves because another session had a council
run in flight, and a roll destroys one.

When it rolls, five sites (oufe, robot-hands, vetcomparison, vonc, webdesign) can
find their contact email and will render the block on their next page rebuild.
Three cannot: gamesdesign, loancalculator and relojistas have no contact detail
anywhere. That last one is deliberate — the owner ruled it has no contact route —
and it is now the negative control: if it *starts* showing a contact block, the
fallback is fabricating and the fix is wrong.

## Where we're going

1. **Roll and verify.** Three checks, in the runbook: pod-grep with a positive
   control in the same exec, induce the failing case on vonc.com, then the
   negative control on gamesdesign. Then the ticket moves to `bugs_closed/`.
2. **Rebuild five contact pages** so the block actually appears.
3. **Two things handed to other lanes, deliberately not done here.** The component
   library owns the 74 declared data addresses that point at categories no site
   has, the near-duplicate vocabulary (`nav` vs `navigation`, `cta` vs `ctas`), and
   the paths using array syntax the resolver cannot parse at all. They have been
   told in their cold-start doc, not just in a directory.
4. **One open question for the owner**, written into the register rather than
   quietly decided: a resolver fallback fixes every component at once, but it
   means the address a component declares is no longer the whole truth about
   where its value came from. The alternative — make the research agent write
   where the components read — is cleaner and needs a migration. And separately:
   `sync_site_identity` exists, does exactly that copying job, and has never been
   wired into anything. Someone should decide whether it should be.

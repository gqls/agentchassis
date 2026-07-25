# Summary — 2026-07-25: the site became reproducible, and the platform told on itself

*Read-out for the owner. Previous: `SUMMARY_2026-07-24_components_live_across_site.md`
(components live but hand-placed). This one marks a different position: the site
is now rebuildable by the platform without losing anything, its links are sound,
and the process of getting there exposed three platform defects that affect every
site we run.*

## What we're trying to do

Two things, still. Build genuinely good consultancy-brochure components — the
interactive kind that Bain and McKinsey sites use — **through the framework**, so
any site we run can have them. And stand up fundamentallyai.com as a site that
markets this platform's real, verified capabilities, with every claim sourced.

## Where we've come from

The site was onboarded and built overnight on 20 July. Two blockers held it back:
a content check that mistook our own approved reference to a sibling site for
contamination, and hosting. Both cleared by 22 July. On 24 July the five
interactive components went live, one per page — but **placed by hand**. That was
the honest weakness of the last read-out: the site looked right, and a routine
rebuild would have wiped it.

## What we've done

**Made the site reproducible.** All five main pages have now been rebuilt by the
platform itself, and every one automatically restored its interactive component
from the site's own plan — the carousel on capabilities, the people block on
about, the swipeable strip on the council page, the hover cards on fine-tuning,
the counting stat band on the homepage. Nothing was re-placed by hand. That
closes the whole class of problem where a rebuild silently undoes design work.

**Built the trust story its own page.** The self-correction account — naming
leopardessconsulting.co.uk directly as the worked example, without repeating the
invented details as if they were facts — is live with five sections. The brief
called this the differentiating story; until today it had no page.

**Finished the contact page.** The phone number is live and tappable. It took a
real investigation to get there (below).

**Repaired every internal link.** Twenty-one of twenty-two were broken.

**Filed three platform defects and one feature spec**, each from friction we hit
rather than from speculation.

## Where we are now

Ten pages deployed. Every internal link resolves. Contact details display. The
five components are in the site plan, so they survive rebuilds. The copy carries
the new plainer voice, with one known residue: em dashes persist after two prompt
tightenings.

The more valuable position is what we learned about the platform, because it
applies to every site:

**The gate finds broken links and throws the answer away.** The content check
detected all eight broken links on the newest page, by name, at build time — then
deployed it, because they are classified as warnings. On the success path the
findings are never saved; they live about a day in run state and expire. The code
comment justifies this by saying the improvement loop will repair them. That loop
is switched off. So the platform knew, said so internally, kept no record, and
shipped. Nothing needs *building* to detect this class — it is already detected
perfectly. It needs enforcing and keeping. (`bugs_open/071`)

**The contact-details block cannot render on most of our sites.** It reads the
business phone and email as top-level fields; the thing that writes that record
files them one level down. The lookup fails, and because a missing email is
configured to withhold the whole block, the section vanishes from the build with
no error and no skip record. Across our thirteen live sites the correlation is
exact: the five with the flat shape are the five with a contact block; the eight
without have none. The default way a site is created produces the broken shape,
so **every new site is born without contact details**. (`bugs_open/072`)

**Deliberate rebuilds are treated as runaway loops.** A housekeeping job parks
build requests idle for 48 hours, but measures the age of the *request row* rather
than how long it has waited. Reusing an old row — the only way to ask for a
rebuild today — means the request is stale on arrival. It parked one of ours
twelve minutes after we made it, stamped "48h+". The workaround is to create a
fresh row, and with fresh rows batches are safe. (`bugs_open/070`) The paved road
for this already exists in the codebase, wired end to end, with a
"requested_by" field and everything — and no trigger, two rows ever, last used in
February. The feature spec is to revive it, not rebuild it. (`features_open/021`)

**Two of the pages the brief asked for were never in the site's plan.** Not a
failure — an absence. The decision-record index still has no sections, which is
why it has never built. The builder handles this well: it looks, finds nothing,
and stops in under a minute without spending anything on writing.

## Where we're going

**Owner decisions needed on two things.** The em dashes: a third prompt attempt
using a worked before-and-after example, or a mechanical pass that strips them
after writing. The second definitely works and is free; the first is cleaner if
it lands. And the decision-record page: the self-correction page currently
promises a record "you can ask to see it", which is an overclaim while that page
doesn't exist — either we build it or we soften the sentence.

**Then the automation set, in the order already agreed.** The brief-fidelity
auditor (016) is built and seeded and needs a post-wave run to read properly.
Next is the component-adoption check (017) — does the planner ever *choose* the
new components, or do they only appear where we placed them? Then the specialist
design critic (018), using Gemini on rendered screenshots. Sweep enrolment (019)
stays last until the improvement loop is back on — though `071` sharpens why it
matters: enrolment is the *other* route to a durable record of findings, and it is
also off for this site.

**Three contract fixes are queued behind owner/thread capacity**, all small and
all fleet-wide: persist the gate's warnings, correct the contact-field paths, and
make a silently-withheld section loud.

## One thing worth saying about how today went

Twice I reported something verified when it wasn't, and the same mistake caused
both. I found the broken links with a search pattern, fixed them with the same
pattern, and confirmed with the same pattern — all three shared a blind spot for
links that jump to a section, so all three agreed and all three were wrong. What
caught it was visiting the live pages. Earlier, my database queries were trimming
a text field for readability, and the trimmed-off part named the exact cause I
then spent two wrong theories guessing at.

The rule now written into the workstream's command book: **a check that shares its
logic with the fix cannot test the fix, only agree with it.** Verify against the
served artefact, never the source you just edited. Both incidents are in
`WRONG_CALLS.md`, because the tally is the useful part.

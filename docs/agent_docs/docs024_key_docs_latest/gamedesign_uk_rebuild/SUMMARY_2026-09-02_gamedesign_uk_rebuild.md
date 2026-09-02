# SUMMARY — gamedesign.uk rebuild — 2026-09-02 (first milestone: the site is live)

Written to be read aloud. Current state only; the chronology is in NOTES and README_where_we_are.

## What we're trying to do

Take gamedesign.uk — a domain we own that had been serving broken, empty pages to the public
since April — and turn it into a real site in its own right: the professional-practice side of
game design, written for leads, producers and senior designers, deliberately distinct from its
sibling gamesdesign.co.uk, which does free calculators and guides for people learning the craft.
Two seats, one domain pair. And along the way, close the gap that let a broken site sit unnoticed
for four and a half months.

## Where we've come from

In April an adoption was run on gamedesign.uk with the site as both the source and the
destination. It wiped the pages, recreated them as empty placeholders, and the rerender published
the placeholders over the good HTML before any content was written. The thread at the time saw the
empty pages and called them temporary. The content never came, the site row was later deleted, and
with it went every handle a repair could have used. Thirteen of forty-seven files served an empty
body until this afternoon. Nothing in our monitoring could see it: every detector starts from a
database row, and this site had none.

Three guards landed between May and July that stop that publish path today. A second investigation
on a different model confirmed the mechanism exactly and corrected seven details of mine, and found
that one of those guards only protects new publishes — an already-served empty page is never
repaired, which reopened bug 315 on another site.

## What we've done

The diagnosis, verified twice. Bug 432 filed for the monitoring gap, and its fix built and run: a
check that lists what the storage bucket actually serves and asks the database whether it knows
each domain. First run found ten domains serving with no record. The owner ruled on all of it:
rebuild through the framework, different direction, agreed with the positioning lane; a bespoke
warm-paper, serif, earth-accent look from the theme-kits lane, seeded where the pipeline reads it;
the old files cleared from the deploy repo so the domain 404d honestly for the half hour it took;
oxenunity.com given the row it should always have had; 315 reopened and handed to its site's lane,
where it was fixed the same day; the name clash routed to a new gamesdesign.co.uk session and the
class behind it filed as 439.

Then the build. Dispatched at 17:07. Homepage live at about 17:56, styled by 18:00.

## Where we are now

Four pages serve to the public with real content, a sitemap, a real contact address, and every
internal link resolving. The homepage opens "Game design, examined as a practice, not a pitch."
The copy contains none of the things the brief forbade and links to the sibling where tools are
mentioned. The classifier's own classification of the site was "editorial — not a tool platform;
those live on the sister domain."

The look is warm off-white paper, dark warm ink, a rust-brown accent, Playfair headings over Libre
Baskerville body, magazine-grid layout — everything the owner asked for in direction. The exact
hex values are not the ones seeded: the render step took the classifier's near-identical values
and re-derived the rest as warmer browns. Under today's ruling that a seeded palette is a starting
point and not a lock, that is the system exercising the authority it was given, and it is recorded
as values, not as a complaint. It did produce one finding for the design lanes: the composed
palette record and the served stylesheet disagree on every slot, so the composition record does not
describe what the public sees. Theirs to judge.

Parked, deliberately: an article slot the planner created with nothing to fill it, and three
call-to-action buttons that have nothing real to point at until articles exist. Both are the
system refusing to ship a dead page or a dead link. Missing: a favicon, and privacy and terms
pages — neither planned nor linked, but a public site probably wants a privacy notice.

## Where we're going

Owner decisions outstanding: whether to leave the article slot parked or cancel it; whether the
site needs legal pages; the retract-or-rebuild call on ai-agent-orchestration's archived empty
page. The eight remaining rowless domains are the adoption backlog, after this lane, with oversight.
Bug 432 stays open until its check is scheduled rather than run by hand. This lane closes when the
owner has looked at the site and the two parked shapes have a ruling.

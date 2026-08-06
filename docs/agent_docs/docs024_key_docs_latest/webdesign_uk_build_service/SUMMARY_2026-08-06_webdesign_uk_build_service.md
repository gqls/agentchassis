# SUMMARY — webdesign.uk build service, 2026-08-06

*Second in the series. Current state only; the chronology is in
`README_where_we_are.md` and the technical log in `NOTES_...`. The previous
summary (2026-08-04) stands as what we believed then.*

---

## What we are trying to do

Unchanged: sell complete websites to small British businesses at twelve hundred
pounds, total, no VAT, built by our own framework and shown to the customer on a
private link before they pay a penny. The customer types their domain into a box
on webdesign.uk, we build the site, they accept it or walk away with a full
refund.

## Where we have come from

Since the last summary the infrastructure went from planned to real. The server
was ordered and set up in an afternoon: a small machine in Cambridge that
nothing on the internet can connect to directly. It pulls its pages from our
repository every five minutes, its firewall refuses everything except our own
login, and the one genuinely valuable thing on it, which will be the
conversations customers have with us, gets a backup design rather than an
afterthought.

The tunnel deserves a plain explanation, since it took four attempts to
authorise and is worth understanding. Normally a web server sits on the internet
with its doors open and a firewall deciding who comes in. Ours does the
opposite: it has no open doors at all. Instead, the machine itself dials out to
Cloudflare and holds that line open, and visitors' requests travel down it.
Nobody can attack a door that does not exist, we never manage certificates or
firewall rules for the web side, and because every visitor arrives through
Cloudflare, we can trust what Cloudflare tells us about who they are, which is
what makes a fair per-person limit on the chat possible later. The four failed
authorisations were a timing problem, not a fault: the authorisation link only
lives for about eight minutes, and the click kept landing after the deadline.
It is done now and never needs doing again.

The build pipeline also ran end to end for the first time: research, strategy,
planning, design, writing, checking, deployment, all through the framework, with
the model swapped from Gemini to Claude mid-week when Google's daily allowance
ran dry. That swap was meant to be temporary; the owner has kept it.

## What we have done

Built and shipped a first version of the site through the framework, and then
had it correctly rejected by the owner. That rejection is the most useful event
of the week, because tracing its three complaints led somewhere important.

The site came out as one page when the fleet's planner normally produces twenty
or thirty. The styling and images were missing entirely. And the copy read as
dense and unfriendly, nothing like a person talking. Each complaint traced to
its own cause. The missing styling is a delivery gap: the pages travel to the
server through the repository, but the stylesheets and images travel a
different route that never reaches it, and my verification failed to notice
because I checked the text of the page rather than looking at it as a visitor
would. That failure of checking is written up in the missteps ledger in some
detail, because it is the second time this lane has carefully verified the
wrong thing.

The one-page shape has a more instructive cause. Back on the third of August I
hand-built a stand-in page, against what is now a standing rule. We deleted it
days ago, but the damage kept compounding: the research step had already read
that page, concluded the site was meant to be a single-page brochure, and
classified it that way with near-total confidence. Everything downstream
honoured that classification. So the hand-built error has now cost us three
separate ways: banned phrases laundered into the build instructions, a broken
expectation about assets, and the very shape of the site. It is the best
argument for the framework-only rule anyone could have manufactured.

The copy tone had three compounding causes: build instructions derived from
that same contaminated source, my own writing rules which optimise hard against
overclaiming but not for warmth, and a writer prompt that has since been
improved. The owner has ruled: rewrite everything under the improved prompt.

## Where we are now

The machinery is right and the product on it is wrong, which is the correct
order to have things wrong in.

Working and proven: the box, the tunnel, the pull deployment, the build
pipeline on Claude, the price and guarantee copy rules enforced mechanically,
and a private preview address where the owner can see whatever is currently
built. Not working: the site itself, rejected on shape, styling and tone. The
public domain currently shows nothing at all, because the holding redirect was
removed in the dashboard; restoring it is the first job for the new Cloudflare
token, which the owner is part way through creating.

## Where we are going

A full rebuild, through the framework's own submission triggers, constrained
this time rather than left to inference. The resubmission carries an explicit
list of pages so the planner builds a proper site instead of guessing from a
contaminated signal. The build instructions that derived from the hand-built
page get regenerated from scratch rather than patched a third time. All copy is
rewritten under the improved, friendlier writer prompt. The asset delivery gap
gets diagnosed properly, by finding how the two older server-hosted sites
actually get their styling across, and using that mechanism rather than
inventing one.

After that, in order: the domain-input box backed by a small hand-written chat
service on the box, with spending limits built in from the first line; the
owner reviews the rebuilt site on the preview address; and only then does the
public domain point at any of it. The holding redirect covers the gap.

Still with the owner: finishing the Cloudflare token, the correction-fee
number, terms before any real payment, and an Anthropic key for the chat.

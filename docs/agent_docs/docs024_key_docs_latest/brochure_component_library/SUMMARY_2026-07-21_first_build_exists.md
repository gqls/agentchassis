# SUMMARY — brochure component library (2026-07-21): the first build exists

**What we're trying to do.** Build genuinely good-looking, interactive brochure
components — Bain/BCG/McKinsey-style hero carousels, hover-zoom cards, swipeable
mobile carousels, code-rendered stat bands — reusable through the site-generation
framework, and stand up a new brand, fundamentallyai.com, that markets this
platform's own real, verified capabilities as consultancy service lines and case
studies.

**Where we've come from.** A day ago this was a research question. We studied the
reference sites and pulled concrete, implementable recipes for each effect;
mapped how our own framework turns a mission brief into a live site; and built a
fact-checked inventory of what this platform genuinely does well, separating the
real from the aspirational. The owner made every positioning call: line
illustration for people, the embeddings offer framed honestly as buildable
rather than already-delivered, the "we catch our own AI's mistakes" story told
openly with the sibling site named. All of that went into a mission brief the
owner approved before we fired it.

**What we've done.** We onboarded fundamentallyai.com through the real pipeline
and it built almost the whole site overnight, unattended. This is itself the
proof of the "rapid delivery" pillar we wanted to sell. The mission brief carried
all the way through: the site came out with a dark navy-and-amber consultancy
palette, an explicit instruction that people are drawn as line illustrations with
one consistent tint and never photographed, charts rendered from real data rather
than generated as pictures, and — most tellingly — pages named for exactly the
capabilities we chose to lead with, including a multi-agent-review-council page
and a self-correction page that names leopardessconsulting directly. The research,
the decisions, and the machine's output are all in agreement, which is the moment
this stopped being a plan and became a thing that exists.

**Where we are now.** The site is built but not yet live, and honestly, the two
things standing in the way are the useful part of this milestone. First, five of
the content pages — including the homepage — are held back by the platform's own
content-validation gate, one blocker each, while two other pages passed cleanly.
We can't yet see what the blocker is, because the logs that would show it were
rotated when the fresh v1.0.1144 build restarted the system; the honest next step
is to rebuild one page and watch it fail in real time rather than guess. Second,
nothing is actually being served yet: the domain is on Cloudflare, but the path
from our rendered pages to what a visitor's browser receives isn't delivering,
and that's not yet diagnosed — part of it is probably an infrastructure step at
the owner's end, the same shape as when idea.uk was cut over. Neither of these is
a nasty surprise; they're the normal last mile between "the pipeline ran" and
"the public can see it," and the platform's own rule — trust the rendered page,
never the status field — is exactly what caught them, since the database cheerfully
calls two pages "deployed" that serve nothing.

**Where we're going.** The next thread picks up from a written handoff with two
clear jobs. The immediate one is to get this built site unblocked and genuinely
live: diagnose the validation blocker by watching a real rebuild, sort out the
serving path, and feed the couple of sections that are correctly asking for real
data rather than inventing it. The larger one is the original ask — the fancy
components themselves, which still don't exist in the framework and are a genuine
build, starting with the hero card carousel that exercises the carousel motion,
the hover-zoom image treatment and the imagery question all at once. The site as
it stands today uses the standard existing components; making it look like the
reference sites is the work that remains after it's live.

# SUMMARY — brochure component library (2026-07-24): the components are live across the site

**What we're trying to do.** Build genuinely best-in-class interactive brochure
components — Bain/BCG/McKinsey-style card carousels, hover-zoom and hover-reveal
imagery, swipeable mobile strips, code-rendered stat bands — reusable through the
site-generation framework, and prove them on fundamentallyai.com, the new brand
that markets this platform's own real, verified capabilities.

**Where we've come from.** The last summary marked the site's central blocker
resolved: the content gate had been silently stripping the owner-approved
self-correction story, and the fix turned out to be a data switch for code that
was already live. Since then the owner cut the site over to hosting, and the last
mile was cleared: dead menu links removed, phone captured, the homepage's "our
work" showcase fed with three grounded projects.

**What we've done.** The original ask, delivered. All five interactive components
now exist in the framework's component library — an auto-advancing (opt-in),
swipeable, hover-zoom hero card carousel; a hover-reveal image card grid; a
swipeable insight strip; a code-rendered stat band with an honest count-up; and a
line-illustration people/approach block. Each is accessible by construction
(pausable rotation, keyboard-safe, reduced-motion and touch aware, screen-reader
friendly) and version-controlled alongside the exact content placed on each page.
The owner reviewed the first component live and shaped it through two rounds of
design feedback — overlaid arrows over the imagery, whole-card click-through,
tighter cards, and stillness by default with movement as a per-carousel opt-in.
Then the rollout: one new component per page across the live site, every line of
copy grounded in verified facts, and every page checked live after deploy — all
up, rendering cleanly, with the count-up wired through the site's script bundle.
A real platform gap surfaced and was documented on the way: per-component
scripts were published but never loaded on any page, so component behaviour now
ships through the site-wide bundle, which actually works.

**Where we are now.** fundamentallyai.com is a live, varied, interactive
brochure site: real numbers counting up on the homepage, a still-until-clicked
card carousel on capabilities, an illustration-led approach block on about, a
swipeable evidence strip on the review-council page, and hover-reveal cards on
fine-tuning. The pages differ from one another the way the reference sites do.
Every figure shown is real; every claim traceable. The component library is
proven end-to-end and ready for reuse on any other site the platform builds.

**Where we're going.** The remaining items are all smaller than what just
shipped: the three empty pages (the dedicated self-correction page first), a
contact-details block so the phone number displays, one stray link on the
council page, and — the next structural step — letting the site planner select
these components automatically for future pages, which registration has made
possible but which hasn't yet been exercised end-to-end.

# oufe.com — where we are, 2026-07-29

Fifth in the series (2026-07-26 / 27 / 28 / 28b). Written after the render-audit
loop closed and the first time-series chart shipped, both this morning.

## What we're trying to do

Build oufe.com into a UK corporate-finance and restructuring research site whose
distinguishing property is that it cannot publish an unsourced figure — every
number on every page traces to a registered fact with a named, dated source.
The flagship is a living Thames Water dossier plus interactive tools, with the
mechanism explained exactly and the live facts about real companies kept
separate and separately verified.

## Where we've come from

The site went live on the 26th with five pages and zero broken links. Over the
two days after, it gained the case write-up, the recovery-waterfall tool, legal
pages, a mechanism diagram, and its first two charts — and the owner found one
of those charts unreadable in a screenshot, which exposed that all of our
checking tools existed but nothing ran them. Yesterday evening the render audit
was rebuilt as a proper platform citizen: its own measurement code, its own
dedicated browser pod, a seeded agent to drive it. The last hop — actually
dispatching it — failed silently, and diagnosing that was this morning's first
job.

## What we've done

The silent failure turned out to be one wrong word in the agent seed
(`initial_step` for `start_step`): the platform had rejected every dispatch and
sent the error to a reply channel nobody reads. One query found it, one key
fixed it. The audit then ran end to end and immediately earned its keep: it
exposed that the contact page had been invisible to every check (its database
row said "not deployed" while the page served fine — a planned section that was
never built kept it permanently unstampable), and once that page became visible,
the audit found a real accessibility failure on it — white text on a gold
button. We fixed the button, re-ran the audit, and got a clean site: eight
pages, no failures. Along the way we found that the standard "contact
information" component invents a phone number and office hours when a site
doesn't supply them — six other live sites are publishing those invented hours
today. That's filed as a fleet bug with the evidence; we didn't touch the other
sites.

Then the owner's standing ask — more charts — met the one component still
unused: the time-series renderer. Thames Water's own annual report charts five
years of leakage performance, and that became the register's first series fact:
five dated observations, each carrying its own citation, plotted so that a bar
reaching the top of the chart would have exactly met the company's final-year
commitment. The first two years beat their targets; the last three fell short,
and the final miss carried a £19.178 million penalty — all in the company's and
regulator's own figures. The near-miss worth remembering: the regulator's
contemporaneous reports disagree with the company's restated series, because
Thames changed its measurement methodology in 2023-24. We caught that by
cross-checking one overlapping year before committing, and the chart's footnote
discloses the restatement in the company's own words.

## Where we are now

Eight pages live. Thirty-six registered facts, including the first series.
Three charts and a mechanism diagram on the Thames page. The claims scanner
reports zero findings across all eighteen components, and the render audit —
now dispatchable on demand, proven four runs deep — reports zero readability
failures across all eight pages. Every check that exists now actually runs, and
the site is clean under all of them.

## Where we're going

More charts, tools and guides, per the owner's standing ask — the machinery for
all three now exists and has been exercised. The premise-branching and deepthink
design note holds the next big design decision. The render audit should be
pointed at the five other sites with known failures, but those belong to other
workstreams. And somebody — the platform, not this site — needs to decide what
to do about the contact component that invents office hours.

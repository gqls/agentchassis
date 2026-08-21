# agritec.uk — where we are

Plain-prose log for the owner. Append only, newest at the bottom. No jargon where a plain word
will do.

---

## 2026-08-21 — first session: what we found before building anything

You asked to bring agritec.uk into the framework — an adoption, but rebuilt from scratch so the
tools and guides stop being hand-written files and become things the system can maintain and
improve on its own.

Before planning any of that I went and looked at the site as it actually stands. Three things
came out of that, and two of them changed the plan.

**First: the copy in our domains folder is badly out of date.** It has six calculators and six
guides. The live site has **thirteen calculators, six guides, and a seven-part engineering series
about a camera-based crop monitoring system** that isn't in the folder at all. If I had planned
from the folder, I'd have quietly left out more than half the site while believing I had all of
it. So everything below is measured from the live site.

**Second: you were right about the missing links, and it's worse than you remembered.** One of
the guides — the one on vapour pressure deficit — isn't linked from *any* index page. Not the
home page, not the guides index, not the tools index. The only way to reach it is from the VPD
calculator. Three more calculators are on the home page but missing from the tools index, and one
is on the tools index but missing from the home page. So five pages are in some way stranded.

**Third, and this one you hadn't flagged: the market ticker on the old homepage was made up.** The
little strip of wheat, oil, carbon and fertiliser prices. There's a small Go program in the folder
that was supposed to feed it, and that program generates the numbers with a random number
generator — it even labels its own output "Simulated Exchange". The homepage didn't use it
anyway; the prices were typed straight into the page. Someone has already commented the ticker out
on the live site, and it isn't coming back.

Related: the site has six data files of reference numbers — energy prices, LED efficiencies,
fertiliser solubilities, crop light requirements — and **not one page reads them**. Every number
is typed into each calculator by hand instead. The data files are still sitting there publicly
readable, doing nothing.

### The decisions you made, and one measurement that settled a fourth

You chose: agri content first with the IoT series to follow; a fresh submission rather than
letting the system crawl the old site; **every figure properly sourced before any page is
written**; and clean new URLs with a redirect map built afterwards so old links still work. Then
you added that everything should eventually migrate, and that the new pages must be written afresh
but to the same detail **and greater**.

That last instruction needed a number attached to it, or it's just a wish. So I measured. The
existing guides run 315 to 453 words each. The deep dives run 434 to 605. And across all six
guides there is exactly **one** diagram — which is precisely the gap your request for more imagery
and infographics is aimed at.

Then I measured what our own framework produces, because "greater" needs somewhere to go. It turns
out the answer depends almost entirely on one setting. Pages built as **articles** average around
1,600 words across four of our sites. Pages built as **guides** average around 500.

That matters more than it sounds. The site calls them guides, they live at `/guides/`, and the
obvious thing to do is build them as guides — which would have landed them at about 500 words and
reproduced exactly the depth you asked me to exceed, while looking like the natural choice. So
they'll be built as articles, hubbed under a guides index. Target is around 1,600 words each, with
a sourced figure behind every number and at least one proper infographic per page.

### On the cannabis question

I raised it because the crop light table includes cannabis, which is licensed-only to grow in the
UK, and you said drop it. Done. Worth knowing where it actually lives: one line of help text under
an input on the energy calculator, and two rows in one of those publicly-readable data files. Both
disappear at cutover, because the deploy wipes anything not in the new build rather than just
overwriting.

One subtlety I've handled rather than left to chance: the framework writes the content itself, and
it knows perfectly well that cannabis is a controlled-environment crop. It could reintroduce it
without ever having seen the old page. So rather than relying on an instruction in a brief, it
goes in as a hard banned pattern that the publishing checks enforce.

### Where we are right now

Nothing has been sent to the cluster yet. No site record, no build, nothing running. What exists
is the plan, the commands, and a **subject ledger** — a list of all 26 live pages, each with what
it teaches and how deep it currently is, so that "nothing gets dropped" and "nothing comes out
thinner" are things we can check rather than hope for.

The next step is the seed: creating the site record and, importantly, the evidence register
*before* any page is written. That ordering isn't fussiness — if the evidence register isn't there
first, the checks that stop invented numbers reaching the page silently do nothing at all, and
they do it without complaining.

After that comes the part that will take the longest, and it's the part you chose deliberately:
sourcing every number properly — the SFI payment rates, the energy prices, the LED efficiencies,
the bioconversion rates — before a single page asserts any of them. Some of them may turn out to
have no citable source at all, and where that happens the honest answer is that the number becomes
something you type into the calculator yourself rather than something we assert. Saying "we
haven't verified this" is always publishable. A plausible guess never is.

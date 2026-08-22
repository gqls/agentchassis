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

---

## 2026-08-21 — later the same day: the foundations are in, and one thing worth your attention

The site record and its guard rails are now in the live system. Concretely that means agritec.uk
exists as far as the framework is concerned, with a contact address, an evidence register (empty
of facts, deliberately, because nothing is verified yet), and a house style for imagery.

**Nothing is running.** I checked rather than assumed: zero queued jobs, zero pages. The seed
creates the guard rails and starts nothing. That ordering is on purpose — if the evidence register
isn't in place before the first page is written, the checks that stop invented numbers silently do
nothing at all, and they don't complain while doing it.

I also tested the cannabis ban rather than trusting that I'd written it correctly. It fires on a
cannabis line and stays quiet on a leafy-greens line — worth doing, because a badly written rule
can match everything and look like it's working.

### The thing worth your attention

While checking that the honesty rules actually reach the writers, I found something I think you
should know about, because it bears directly on what you asked for.

The framework has a set of instructions that tell a writer "you may not assert a number unless
it's in the register". I checked which agents actually receive that instruction. The two that
write **prose** do. The two that write **calculators** — the one that builds a new tool, and the
one that later *improves* it — do not. The register is handed to them along with everything else,
but nothing in their instructions points at it or tells them it's binding.

That matters here more than on most sites, for two reasons. This site is mostly calculators, and
their numbers — payment rates, lighting efficiencies, carbon fractions — live inside the
calculator's code. And there's a known gap at the other end too: the checks that scan a published
page read the visible words, not the code, so a figure baked into a calculator is checked by
nothing afterwards either.

Put together: for prose we have both an instruction going in and a check coming out. For
calculators we currently have neither. And the agent with the thinnest coverage is the one that
would *evolve* the tools over time, which is precisely the thing you asked for.

I've handled it for this build — the honesty rules go into the tool brief itself, in its own
words, rather than relying on the register being noticed. So agritec is covered.

**What I haven't done is file it as a platform problem**, because that's a claim about the whole
estate rather than about this site, and the house rule is that a claim like that goes through the
diagnosis loop first rather than being asserted from one afternoon's reading. Say the word and
I'll do that properly. It would affect every site with calculators, which is most of them.

### What happens next

The long part, and it's the part you chose: sourcing every number properly before any page
asserts one. Six areas — SFI payment rates, energy prices and carbon intensity, LED efficiencies,
crop light requirements, insect bioconversion rates, and seaweed carbon fractions. Some of those
will have a clean official source. Some, I suspect, won't have anything citable, and where that
happens the honest answer is that the number becomes something you type into the calculator rather
than something we tell you.

---

## 2026-08-22 — the first evidence run found something you should probably act on today

I started the sourcing work with the SFI payment rates, on the reasoning that they're the numbers
on the site most likely to have a clean official source and the ones where being wrong costs a
reader real money. That turned out to be the right place to start, for a reason I didn't expect.

**The SFI management payment has been abolished, and the calculator on your live site still pays
it.**

The calculator leads with it. There's a green box at the top of the SFI Revenue Stacker saying you
receive £20 per hectare for your first 50 hectares, and that this means "the first £1,000 of your
SFI income is effectively guaranteed". It's a line item in the results panel. And it sits directly
above a link to the official GOV.UK guidance, which makes the whole thing look checked.

The government's own SFI26 scheme rules say, in as many words: *"the SFI management payment has
been removed for SFI26 agreements"*, and *"You will not be paid: an SFI management payment for
your SFI26 agreement"*. DEFRA's farming blog explained the reasoning back in February — it was
always meant to be a temporary payment to help people move into the scheme, and dropping it frees
money to fund more agreements.

I didn't take the research agent's word for this. It's built to re-fetch each source page and
throw the claim away unless the quoted sentence is still there word for word, which is good, but
that only proves the words exist — not that they were read properly. So I opened the GOV.UK page
myself and checked all of it in context. It's unambiguous.

So a farmer using your calculator today is being told they'll receive up to £1,000 they will not
receive, and being told it's guaranteed.

Three other things came out of the same run that the current calculator doesn't know about either:
there's now a £100,000 annual cap on an SFI26 agreement, a three-hectare minimum to be eligible at
all, and a limit of 25% of your farm's area on certain action types.

**A question for you, and it's the only urgent thing here.** The rebuild will take a while, and the
old site is live and wrong in the meantime. Do you want to leave it as is until the new site
replaces it, take that calculator down now, or put a correction notice on it? I haven't touched it
either way — that's your call, not mine, and all three are defensible.

For the rebuild itself it's already handled. I've added rules to the site's evidence register that
make it impossible to state the management payment as something a reader will receive. I was
careful about how: the new site absolutely must still be able to *explain* that the payment was
removed, and what it used to be, because right now that's genuinely the most useful thing this
site could tell an SFI reader. So the rules block "you will receive" and leave "was removed" and
"was available under the 2023 offer" alone. I tested both halves of that rather than assuming —
two of the five rules were wrong on the first attempt and let through a sentence they should have
caught, or caught one they should have let through.

This is the strongest argument yet for the thing you asked for. The framework has a machinery that
re-checks sourced figures on a schedule and raises a flag when the underlying source moves. A
hand-built calculator has nothing of the kind, which is why this sat there quietly getting more
wrong.

---

## 2026-08-22, later — the SFI calculator, taken apart

You asked me to make it use correct figures and to deconstruct it as needed. Both done, as far as
they can go before the new site exists. The short version is that patching it was never an option.

I went and got the real SFI26 rates. They're published on GOV.UK in twenty-one tables, and I read
them directly rather than sending the research agent — because the rates sit in table cells, and
the verification step needs a quotable sentence, so a research run would have rejected them as
unverifiable. That's a trap I'd hit earlier the same day and written up, so this time I went
straight to the reliable route.

Seventy-one actions, and I have a decent check that I got them all: the page separately says in
its own prose that there are 71 actions, down from 102, and that sentence had already been
verified independently. My count agrees with it. If I'd missed a table the two numbers wouldn't
have matched.

### What the audit found

Your calculator has nine revenue lines. **Two of them are right.**

The management payment is abolished — you know that one. But four more of its actions **no longer
exist at all**: the soil assessment, the pest management plan, the nutrient management review, and
the hedgerow assessment. Those were the paid planning actions, and DEFRA dropped them for the same
reason they dropped the management payment — to free money for more agreements.

Of the four actions that do survive, two have moved. **Herbal leys is £224 per hectare, not £382.**
The calculator overstates it by £158 a hectare, which on fifty hectares is nearly eight thousand
pounds of income that isn't there. And hedgerow management went the other way — £13 rather than
£10 — but with a catch: it's now **per 100 metres for one side**, where the old tool just said per
100 metres. On a boundary that's a factor-of-two difference, and it flatters the figure on screen
before going against the farmer at audit.

### What the new one has to do that the old one didn't

Three things, and none of them is cosmetic.

**Handle seven different payment units.** The old tool assumed per-hectare or a flat fee. SFI26
pays per hectare, per 100 metres one side, per 100 metres both sides, per square metre, per pond,
per plot, and per tonne. That "one side / both sides" distinction has to be visible on screen.

**Apply the caps.** There's a £100,000 annual ceiling on an agreement, a three-hectare minimum
before you can apply at all, one agreement per business, and a rule that certain action types can't
cover more than a quarter of your land. The old calculator models none of these, so it will happily
total up a number the scheme would never pay.

**Respect the constraints buried in individual rates.** Ponds are capped at three per hectare.
Skylark plots have a *minimum* of two. And supplementary bird food is capped at one tonne for every
two hectares of a *different* action — so one action's ceiling depends on another's area.

I've written all of that up as a build specification, including test criteria. Two of the tests
exist purely to stop this recurring: one checks that £382 never comes back, and one checks that no
line item called "management payment" ever appears again.

### The one thing standing in the way

The tool gets built by the framework, and the framework needs the site to exist first — it reads
the site's design and voice before generating anything. agritec.uk has no pages yet, because I
haven't submitted it.

That submission is the next step, and it's yours to authorise: it starts the whole build cascade,
not just this one tool. The mission and roadmap briefs are written and waiting for you in this
directory if you'd like to read them first — I'd suggest you do, since the roadmap brief is what
the planner treats as binding.

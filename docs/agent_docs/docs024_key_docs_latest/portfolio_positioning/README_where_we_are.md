# Where we are — portfolio positioning

The owner's running log. Plain prose, append-only, newest at the bottom.

---

**2026-07-31, late evening.** You gave me the full list of domains — about 150 once the
duplicates between the two lists are merged — and said the important thing: you want thick
sites, in genuinely different directions, and the differentiation has to be there from the
start rather than patched in later.

**The first thing I did was count what we actually have.** Those 150 domains collapse to
**42 distinct promises** — a "proposition" being what the name tells a visitor they'll get.
Some names have already chosen their audience for you: an equity release calculator is for
people over 55, an adverse credit mortgage site is for people who've been declined, SMB
mortgages is for businesses. Nothing to decide there except to obey the name. But about a
third of the portfolio — all the savings-rates variants, the generic loan names, the rate
comparison names — promise nothing but a topic, and for those the direction has to be
chosen deliberately, because the visitors arriving have not self-selected in any way.

**And I have to be straight about one group.** About 40 domains are spelling twins —
savingsrates with and without a hyphen, .co.uk and .uk of the same phrase. There is no
direction that separates a hyphen from its absence. Two thick sites for the same phrase are
the same site competing with itself. For each twin pair the honest choices are: point the
spare at the built one, or knowingly accept a near-duplicate. I've listed every pair and
recommended pointing the spares, but each one is your call and I haven't decided any.

**What I built tonight.** Three things, all committed:

First, the **thinking**: a catalogue of every angle I could find for telling two finance
sites apart — who the visitor is (age, credit history, sophistication, business or person,
wealth), what they're buying (size, type, structure), when in their journey they are
(dreaming, preparing, shopping, applying, owning, struggling), what job the site does
(calculator, live data, ranked verdicts, decision trees, forecasts), its stance, its
emotional register, and which side of the market it faces. The most useful discovery: the
*grammar* of a domain name assigns its job. "Rates" names want live tables. "Quote" names
want estimators. "Best" names want ranked verdicts with the method shown. "Which" names
want decision trees. That single observation is what carves apart clusters that looked
identical — the twelve savings-rates domains yield eight genuinely different sites once you
apply it.

Second, the **register**: one entry per proposition, every domain placed exactly once, each
entry saying who it's for, what job it does, what it will NOT cover, and — most importantly —
naming its nearest neighbours and the sentence that separates them. That last part is the
discipline that stops two sites drifting into each other.

Third, the **check**: a small program that refuses the register if any domain is claimed
twice, any domain is unclaimed, any entry fails to say where its ground ends, or two
entries claim the same audience doing the same job. I built this because I confirmed
earlier today that the platform itself has nothing that detects two of your sites
converging — so the register is the only guard, and a guard needs teeth. I tested it by
deliberately breaking the register: it caught the break and named both entries.

**Two things worth knowing from tonight's work.** The check caught me twice while I was
building it — once because my register used shorthand the check couldn't read (I fixed the
register to be checkable rather than making the check guess — the same mistake I made this
morning on the link checker, avoided the second time), and once for a real omission. And a
few flags surfaced early while they're cheap: one of your new domains overlaps a guide
already live on the combined site (resolved: the domain gets the depth, the guide keeps the
overview and links to it); the "banking equipment" pair isn't consumer finance at all; and
I'd advise against building loancash.co.uk — the name attracts exactly the audience the
regulator protects hardest.

**What I need from you, when you're ready:** the twin-pair decisions (I've defaulted every
one to "point at the built sibling" — overrule any), whether loancash gets built at all,
and which propositions to build first — that's a commercial call and you have data I don't.

---

**2026-08-02 — your question: did the trigger build loancash, and what would it take to
build the rest of the portfolio "from the trigger" at full quality?**

Straight answer: the trigger built none of it. loancash.co.uk was hand-written by this
thread — every guide, the three rule-checking tools and their code, the styling, the
structured data. The trigger was used exactly once, *after* the site was live, in its
"locked adoption" mode — whose whole design purpose is to generate nothing (we counted
zero AI writing tasks as the success condition). The platform's generative build
pipeline was bypassed entirely. So "16 minutes" measures my typing, not the machine.

The good news: a fresh-build path through the trigger exists end to end (research →
strategy → briefing → plan → design → a written page per planned page → deploy), and it
does emit sitemaps, canonical links and structured data. What stands between that path
and "the best these sites can possibly be" is six gaps, in order of importance:

1. **Positioning goes in as a suggestion, not a setting.** The only input is a mission
   brief the research step "weights". Our register entry needs to be written into the
   site's specification fields *before* any page is written, and protected from being
   overwritten by the pipeline's own research. We already have the script that writes
   those fields; the change is sequencing. And nobody has yet PROVEN a positioning spec
   changes what the writer writes — that acceptance test (planted marker sentence) is
   already queued and is the gate for the whole programme.
2. **No structural validity check.** Three fleet sites 404 today on links their own
   pages carry. The live-origin checks we built (every link resolves, sitemap honest,
   canonicals correct, structured data parses) exist only as our script — they need to
   run after every pipeline deploy, or the worker needs the directory-index fix.
3. **Tool correctness is unaudited.** For calculator/rates/quote domains the tools ARE
   the product, and nothing checks a generated tool computes the right answer. The
   real-browser audit exists as an instrument; it needs correctness fixtures per
   vertical (assert the £15 and the 0.8%, not just "it responds").
4. **Fact discipline.** loancash's quality is every figure carrying its rule name and no
   market rates anywhere. The writer has no citation mechanism, and a wrong fact in the
   specs gets written into pages and then defended (bug 161). Constraint text can reach
   the writer's prompt; enforcement needs the banned-claims detector armed per vertical.
5. **Truncation guard** — a cut-off AI write can save half a page and report success;
   the post-write structure check is doctrine but not a pipeline gate.
6. **The fidelity dial** (high/medium/low) still modulates nothing — acceptable for now.

Recommended first move: pick one cheap domain and run the experiment — fresh build via
the trigger with the register entry both as mission brief AND pre-seeded specs, marker
sentence planted. One run tells us whether the remaining gaps are polish or foundations.
My honest expectation: content and structure can get most of the way; hand-built tools
with exact regulatory constants are the piece the pipeline will do worst, so a hybrid
(pipeline writes the site, we hand the tools in via the locked-component route) may be
the realistic ceiling for the tool-led domains. Worth testing rather than assuming.

---

**2 August 2026, late night (fresh session picking up the handoff).** The seam
backlog is real work now, and the first two seams are done as designs, reviewed
and shipped.

The big one first: pages can now carry a line that genuinely appears on every
page. Until tonight nothing owned "on every page" — the writers put the
independence line on 3 pages of 15 because each page's writer made its own
choice. The fix rides the footer, which is stamped onto every page when it is
assembled: the shared footer template now has a slot for "compliance lines"
that reads from a site's own configuration. A site that sets nothing renders
exactly as before — we proved that byte-for-byte with an automated test before
touching anything, because fourteen live sites share this footer. Lendzy's two
mission lines (does not lend; independent of the FCA) are seeded, and as I
write the pages are re-rendering one by one and landing in the sites repo with
the lines in place. The reviewer council approved it, first round, and one of
its advisory notes was good enough that we hardened the mechanism further the
same evening: a badly-typed config value now gets refused loudly instead of
quietly degrading the whole footer.

Second: every page will get a canonical link (the tag that tells search
engines "this is my real address"), and pages with no description will stop
shipping an empty description tag. Measured tonight: no page on any of our
sites emits a canonical at all — the machinery simply never wrote one. The fix
is code, so unlike the footer change it waits for the next chassis deploy;
it is committed, tested and council-approved (also first round), and the
verification steps for after that deploy are written down.

Third, a nice surprise: the "favicon never generated" item on the backlog is
not a missing platform feature. The machinery to derive a favicon from a
site's logo exists and runs — lendzy just has no logo yet, so there is nothing
to derive from. That moves the item from "build something" to "give lendzy a
logo", which is imagery work, not platform work.

One process note worth saying out loud: I nearly registered the footer
mechanism as brand new. Reading the register first showed it already existed —
built on 30 July for the analytics tags — and my work is properly its second
user, not an invention. The register did exactly what it is for.

---

**2026-08-11.** You said you want to start building all of them out, but not yet, and
raised four things: you don't like lendzy's design (too obviously AI-made), you don't want
to rule on the twin-pair domains one by one, you asked what's left on the enforcement gaps
and the vigilant-designer work, and asked for this file to be brought up to date. Here's
where each of those actually stands.

**The twin-pair thing is already done — you just haven't seen it summarised.** Back on
1 August you made two rulings (cross-TLD pairs both get built and split by depth;
same-TLD spelling twins split by *seat* — buyer, setter, broker, analyst, compliance
reader all read the same phrase differently) and the second one explicitly retired the
need to decide each pair by hand. Of the ~40-odd twin pairs in the portfolio, only **two**
genuinely can't be separated even with five seats available and still need a plain
hold-or-build-it-anyway call from you: `besthealthinsurancerate.co.uk` and
`bestlandlordinsurancerate.co.uk`. Everything else runs on the policy you already set.
The register's prose still has old flags scattered through it from before that ruling —
that's a tidy-up owed, not a decision owed.

**On the design — you're right, and nobody's tried to fix it yet.** The site's visual
design is a separate AI step from the one that writes the words, and every site so far has
run that step on the same model (Claude). Nothing in any of these lanes has ever attempted
visual variety across the estate — the "pin the palette" mechanism that exists is closer
to a suggestion than a lock, so even re-running the same model can drift, let alone give
you genuinely different design languages. My recommendation: build a second version of the
design step pointed at Gemini and try it on a small number of sites as a real side-by-side,
rather than flipping every site over to a new model and getting one uniform look instead of
another. I haven't built this yet — just researched that the seam exists and is roughly the
same shape as the writer's earlier model swap. Say the word and I'll set it up as an
experiment on a couple of sites, not the whole portfolio.

**Enforcement gaps: one bit of bad-good news.** Two brand new bugs were filed today,
fleet-wide, in exactly the piece of this that had shipped — every assembled homepage's
canonical link points at the wrong address, and page assembly is quietly dropping social
preview tags and forcing every page to claim it's in English. Neither was there because of
anything from this lane; they were just found today. The other gaps from the 2 August list
are mostly still open: positioning only reaches specs for one hardcoded site, not
generalised yet; there's still no standing check that a deployed site's links/sitemap/tags
are actually correct (only a script you have to remember to run); nothing checks a
generated calculator computes the right number; the "don't repeat unverified facts" guard
is switched on for exactly one site; and the fidelity dial still does nothing. None of this
blocks starting a build, but it does mean a build run today needs a human checking its
output, not the pipeline policing itself yet.

**Vigilant-designer / offer-analysis lane has moved further than last reported here.** It
shipped the two premise-awareness fixes (the site reviewer now judges a site against its
actual revenue model; a strategy refresh no longer silently triggers a full rebuild on a
live site), and then went further than planned and built the two offer-completeness
checks too — which have already swept the whole estate and found four real things needing
attention (three sites need a strategy written, one 30-page site has no way to contact
anyone). That's stuck behind a council-review technicality that needs your call, not more
building. The piece that's genuinely not started is the analyser agent itself and the
design-critique half of the visual-diversity work on that lane's side — which, worth
noting, is a different design effort from the Gemini idea above: that one critiques a
site's design against screenshots, this one is about generating more than one style in the
first place.

**Full detail:** `SUMMARY_2026-08-11_where_things_stand.md` in this folder — five short
sections, written to hand to someone else if you want to talk it through.

---

**2026-08-11, later the same evening.** You made both twin-pair calls yourself (health
by England-vs-UK, landlord by portfolio size) rather than going pair-by-pair — both are
now in the register. You picked A-track for the design lane over the offer-analyser, and
picked how mortgagecalculator.co.uk's language bug should be fixed. Then you asked me to
look at what a different session had been doing on mortgagecalculator.co.uk's copy voice
tonight — it's a real, evidenced case of the same "looks AI-made" problem you raised about
lendzy's design, except there the cause turned out to be the brief, not the model. You
haven't seen or approved that site's new copy yet. You asked for the enforcement-gap work
(the structural-checks/fact-discipline/fidelity-dial backlog) to happen first, and the
mortgagecalculator copy review to happen after that, then asked for a handoff to carry all
of this into a fresh chat. That handoff is `HANDOFF_2026-08-11_continue_here.md` in this
folder.

---

**2026-08-12.** A fresh session picked up that handoff, and you asked it to go further:
move toward all ~152 domains live, full-featured — 20-plus pages each, images, good copy,
tools, guides, infographics/graphs where they fit, newsfeeds, and directory listings built
from real web searches, not invented.

Before building anything, it checked what's actually true today rather than trusting the
handoff's own framing, and the picture is more basic than "148 domains left to build." Only
four of the 152 domains have ever had a site built at all — and all four (the mortgage/loan
calculators you already know) were either hand-written or copied from an existing site, not
built by the automated pipeline. The pipeline itself has only been run for real once, on a
throwaway domain deliberately kept outside the portfolio and never made public, specifically
so nobody would mistake a test for a launch. That one run proved the machine can write copy
that follows a brief, but it also came out capped at under 20 pages (a hard limit in the
code, since raised), with several rough edges — wrong canonical links, missing page previews,
tools it refused to build without a person unblocking them — most of which are still unfixed.

You made five calls when this was put to you: don't start any new domain until both the
broken-link/wrong-tag checking (which doesn't exist as a standing check yet, only a script
someone has to remember to run) and the "wrong facts get written into pages and then
defended" problem (bug 161 — genuinely needs a proper review, not a quick patch) are fixed.
Newsfeeds and directory listings should be part of the very first sites, not bolted on later.
Leave the generated calculators on manual review for now rather than trying to automate that
away too. Use the one working chart type (bar charts built from facts a site has actually
verified) rather than commissioning a general graphics tool the platform doesn't have. And
keep any provider directories to facts that don't go stale fast — no rates or premiums
attached to a named regulated firm, which is a compliance risk if it's even slightly wrong.

One more thing worth recording: when asked to double-check the "graphs where appropriate"
claim against a real example, the session went and looked at `oufe.com` directly rather than
trusting its own research — found the claim held (a real chart is live on its Thames Water
page, using exactly the mechanism described), but also noticed in passing that oufe's own
"recovery waterfall" tool doesn't actually draw a waterfall chart despite the name — a small,
separate thing worth fixing on that site sometime, not part of this plan.

Full phased plan: `PLAN_2026-08-12_fleet_buildout.md` in this folder. Work is starting on
the two blocking pieces (the link/tag checker, and bug 161's proper fix) — nothing else
begins until both are live.

---

**2026-08-13.** Both blocking pieces are now built, tested, and approved by the review
council — the gate you asked for before any new domain gets built is satisfied.

Neither went through on the first try, and that's worth saying plainly because the review
process earned its keep. The fact-checking fix went through two rounds: the reviewers
caught, among other things, that the new checker could be fooled by a number appearing
inside a bigger number (exactly the mistake that caused the original bug it was built to
fix), and that a fact could accidentally be "verified" against a different website's page.
Both were real, both got fixed with tests proving the fix. The link-and-tag checker took
three rounds: the reviewers caught that we'd promised a sitemap check in the description
but hadn't built one (now built), and then caught that a comparison I'd written against an
older check rested on a false assumption about how that older check works.

Chasing down that last point, at your "check it once more" prompt, uncovered something
nobody was looking for: that older check — which is switched on and has been running since
April — can never say "all clear". The columns it reads have been empty on every page of
every site for months, so it fires every time, and each time it fires it orders a full
rebuild of the site's pages. Roughly twenty-five of those pointless rebuilds have actually
run. Nothing visibly broke, which is why nobody noticed for four months. It's now written
up as bug 270 with the evidence and two candidate fixes; it isn't part of this build-out's
critical path, but whoever picks it up will save the fleet a steady drip of wasted work —
and it muddies the water for anyone investigating why pages rebuild when they shouldn't.

What's next, in order: the new checks are approved but not yet switched on (deliberate —
switching on is its own small, separate step), and the fact-checking code needs the next
routine deployment to reach the live system. Then Phase B: building the machinery that
gives each finance site a directory of real providers, sourced from live web searches with
every claim verified against its source — the last new capability needed before the pilot
site and then the fleet build-out proper.

---

**2026-08-15, later.** The directory work passed its review on the fourth round — the
round you asked for — and everything it was holding back is now switched on: the page
components are in place, the three weekly research sweeps are armed (first runs will
happen under supervision, not unattended), and the two stale review entries from the old
scheme are retired. You then settled all six open decisions in one go:
remortgagecalculator.uk pilots first; build order goes mortgages, then savings, then
insurance; loanzy.uk stays with the webdesign team; the three oddball domains (the
savings app, the banking equipment site, the unnamed insurance brandables) are held out
of the first waves; reviews of this kind proceed on the advisory record rather than
looping on paperwork; and the mortgagecalculator copy review is now a live repair job —
you've said plainly the homepage copy is bad right now, so the next session on that site
starts from "diagnose it against the correction reference", not "approve it". Where to
look is written in the handoff. The register carries all of this as ruling P9, and its
own checker still passes. Fresh cold-start for the next chat:
HANDOFF_2026-08-15_continue_here.md.

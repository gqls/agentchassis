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

---

**2026-08-15, evening.** The supervised first runs are done for all three directory
kinds, and the whole research chain is now proven honest end to end. The savings sweep
worked first time: fifteen facts registered from fifteen candidates, thirteen real firms
— Nationwide, Coventry, Skipton, Monzo, NS&I and eight more — anchored by the
government's own list of approved ISA managers, which conveniently states each firm's
regulator reference right on the page. The health insurance sweep needed one round of
correction, which is exactly what supervision is for: its first run registered real
insurers (Bupa, Aviva, AXA Health, Vitality and others) but let three bad habits
through — a pound amount smuggled inside a "what do they cover" fact, a duplicate that
overwrote a good answer with a worse one, and marketing fluff ("no shareholders,
reinvests profits") where an underwriter's name should be. All three got rules written
against them, plus one more source (Forbes) blocked because it refuses our
verification checks. The re-run came back clean on every count and added three more
insurers. The directory now holds twenty-five named firms and thirty-three cited facts
across the three kinds, every fact traceable to the page it came from, no prices
anywhere. One gap worth knowing about: the health facts all currently come from a
single broker's comparison guide rather than the insurers' own pages — honest and
verifiable, but thin, and it's flagged for the manual review that happens before
anything goes on a site. Next up is the plumbing that actually publishes these
directories onto pages, then the wiring that lets the improvement loop and the site
planner know directories exist.

**2026-08-15, later — the publishing plumbing is fixed and proven.** The machinery that
takes a finished directory and actually puts its data file onto a website had two old
problems: it only ever looked for the original AI-model directory (so our three new
finance directories could never be published at all), and when it did publish, it sent
every directory to every opted-in site whether the site wanted it or not. Both are now
fixed, with a database-only change — no new software build needed. The publisher now
works through a simple list of "this site wants this directory" pairs, checks each pair
really is wanted, really has a page for it, and really has facts to publish, and refuses
loudly if anyone ever asks it to publish without saying which directory (that exact
silent mix-up put the wrong data out under the wrong name back in July, so the refusal
matters). We watched the first run live: the existing AI site got its three directories
published correctly — three separate runs, three different entry counts, the live files
on the site refreshed within seconds of each run — and the finance directories correctly
published nowhere, because no finance site exists yet. The moment one does (the
remortgage pilot is next), its directory will flow without any further plumbing work.
One honest cost: there's now a list inside the publishing query that has to be kept in
step with the list of directory kinds in the code — add a kind to one and not the other
and it quietly never publishes. That trap is written down in the places people actually
check. The change went to the advisory review council as usual; verdict not yet back at
time of writing.

**2026-08-15, evening — publishing fix approved; the "should this site have a
directory?" question is now asked automatically.** The review council approved the
morning's publishing fix first time, with a few low-level suggestions that turned out to
already be handled in the applied version. Then the second piece of wiring went in: the
small decision-maker that looks at what a site is about and, if it's one of our three
finance verticals, marks it as wanting the matching directory. It now runs in two
places — every time the routine site-improvement loop visits a site, and at the moment
a brand-new site is first classified, so a new build knows it wants a directory before
its pages are planned rather than finding out afterwards. It writes nothing at all for
sites outside those verticals, so the rest of the fleet is untouched. One thing to
know: the routine improvement sweep is currently switched off fleet-wide (it has been
since yesterday, unrelated to this work), so the "existing sites" half of this wiring
sits ready but idle until that sweep is switched back on; the "new sites" half will get
its first real exercise when the remortgage pilot site is built. Also submitted for
advisory review.

### Saturday 15 August, late evening — Phase B's build work is finished

That advisory review came back approved, on its second round. The reviewers had asked
for two things and both were fair. The first was a genuine gap: the "undo" script for
that last change protected the step it was reversing but not the neighbouring steps it
would have written over, so if a third piece of work had been slotted in beside ours in
the meantime, undoing ours would quietly have disconnected theirs. That is fixed — the
undo script now refuses to run at all if anything around it has moved, rather than
guessing. The second was a request to write down, somewhere permanent, a pattern we had
only noted in passing: we have now hand-wired this same kind of small decision-maker
into a site's workflow twice, and a third would be the point to build one proper
mechanism instead of a third hand-copy. That is now a numbered item in the architecture
review track rather than a line in a working file. Writing it down was worth it on its
own, because doing so surfaced something nobody had decided: the equivalent news
decision-maker runs for existing sites but not for brand-new ones, while ours now runs
for both. Nobody chose that inconsistency — it is just what you get when each piece is
wired in by hand, which is precisely the argument for building the shared version.

Then the last two pieces of Phase B went in.

The first teaches the site planner about directories. Until tonight, a new finance site
would be marked as wanting a lender directory, and the planner would then plan the site
without one — because nothing had ever told the planner what that marking meant. The
site would eventually get its directory, but only after a later inspection pass noticed
it was missing, by which point the site had already been built and published once
without it. The planner now knows: if a site is marked for a directory, put the
directory panel on the home page, and give it a dedicated directory page built the same
way our three existing directory pages are built. I took the exact wording for all of
that from the code that defines those pages and from the three real pages already live,
rather than inventing a description of them.

The second switches on eleven automatic checks: six that watch for a site that wants a
directory but hasn't got one, and five site-health checks (broken internal links, bad
canonical tags, invalid structured data, missing page-head essentials, dead sitemap
entries) that were built earlier in this project and had been sitting switched off. One
correction to what I said earlier: I had described the six directory checks as new, but
three of the six pairs were already switched on months ago for the AI-model
directories — the six that were actually missing are the finance ones. Same number,
different six.

Worth knowing about that switch-on: these checks will report nothing at all for now, and
that is intentional, not a failure. They only speak up for a site that has opted into a
directory, and no site has yet. Switching them on before the pilot is built is the whole
point — they are the safety net that would catch the pilot being built wrong. The five
site-health checks are different: they will run on any site we inspect from now on, and
we genuinely do not know how much they will find, because they have never been let loose
on real sites. So the first few inspections after tonight want watching rather than
assuming they will be quiet.

Both changes are submitted for advisory review; verdicts will land shortly and are
credited automatically. That is Phase B complete. Next is the pilot itself — building
remortgagecalculator.uk end to end, which is the first time all of this machinery gets
exercised by a real site rather than by us checking it piece by piece.

### Sunday 17 August — the pilot is away, and two bugs turned up on the way to it

All four reviews came back approved, so the review side of Phase B is closed as well as the
build side. The last one raised something genuinely useful: the checks I write at the end of
each database change to confirm it worked could not actually have failed. They read a value
and then did arithmetic on it, and if the value had been missing the arithmetic quietly
produces "nothing to see" rather than an error — so the check would have reported success
having looked at nothing. That is fixed in the undo scripts. It is the kind of finding worth
paying a review round for: nothing was broken, but a safety net had a hole in it that only
shows up on the day you need it.

Then, before dispatching the pilot, I asked a boring question: what does the system actually
write down about a finance site when it classifies one? The answer was that the fields I
expected to be filled in are empty, and the only thing identifying these sites as
mortgage-related is the domain name itself. Following that turned up a real bug. The code
that decides "this site should carry a lender directory" reads the domain for keywords, and
where a domain contains two keywords that point opposite ways it was picking between them at
random — genuinely at random, differently each time it ran. One of our own domains does
exactly that: mortgage-refinance.co.uk contains "mortgage", which says yes, and — hidden
inside "refinance" — "finance", which says no. So that site would have got a lender directory
or not depending on the toss of a coin, with nothing recorded either way.

Two things about that are worth saying. First, it had done no damage: no site has been
through this yet, so I caught it before it cost anything rather than after. Second, the file
already contained a comment explaining this exact hazard, twenty lines below, guarding a
different part of the same function. Someone (us) fixed one half and left the other, and
because searching for the problem finds the comment, the search that should have caught it
would have stopped at the fix. I have written that up as a general lesson, not just this bug.
It is fixed and tested; it needs the next fleet release to take effect. The pilot site is not
affected either way — its domain contains only one keyword.

The pilot itself is now seeded and dispatched, and it immediately embarrassed me. Part of
setting up a new site is a list of banned phrases — things the site must never say, like
"guaranteed" or a specific saving figure. I wrote six of them, and my own check confirmed all
six were in place. They were all dead. A punctuation subtlety two layers deep meant every
pattern was technically valid but matched nothing, for ever, silently. My check had counted
them rather than tried them, and six broken patterns count exactly the same as six working
ones. I only found it because I tested them against sentences they were supposed to catch.

Correcting it took three goes, and each failure was a version of the same mistake, which is
why I have written all three down rather than just the fix. That whole episode is now a
warning in the shared file so the next person seeding a site tests their guards instead of
counting them.

One deliberate decision worth flagging: this site starts with no verified facts registered at
all. A remortgage site is made of numbers, and I could not check a single rate or threshold
against a live source in this session. Writing plausible numbers with invented sources is
precisely what this whole verification layer exists to prevent, so the site starts with none
and is instructed to say "this depends on your lender" rather than guess. The real, cited
material comes from the lender directory, which is what the pilot is meant to exercise.

The build is now queued. Three things need checking as it moves, in order: that the site gets
marked as wanting a lender directory when it is classified; that the planner then puts a
lender directory page into the plan; and only if those failed, that the automatic checks
catch it. The third is the safety net, so a clean run has to be confirmed by looking at the
plan directly — a quiet safety net is not the same as a good result.

### Tuesday 18 August — it works, and there is a real page to look at

The lender directory page is live and I have read the actual HTML rather than trusting a
status field. It says "UK mortgage lenders, listed", it names Mansfield Building Society and
Family Building Society, each fact carries a link to the source it came from, and it opens by
saying plainly that it does not list rates, fees or APRs because those change daily. That last
sentence matters more than it looks: it is the compliance rule you set weeks ago, having
survived every stage between a researcher reading a lender's website and a page being
published, without anyone re-typing it along the way.

So the question this pilot existed to answer is answered. A site can be marked as wanting a
lender directory, plan itself around one, build it, and publish a page of cited, non-price
facts about regulated firms. Nothing about that was true a week ago.

The site itself is half-published: three of its six pages are live, three still need
rebuilding, and there is a queue of things waiting for a human — mostly calls-to-action the
system could not resolve on its own. That is ordinary new-site work rather than anything
broken. The genuinely broken items are a handful of component failures and one where a
component is leaking raw template code into the finished page; that is a platform bug and I
have written it up rather than patched around it.

Two corrections to what I told you yesterday, because both would have misled you. I said the
retried image jobs were succeeding when what I was actually looking at was a job that had
finished the day before — I read the starting number as a result. And I had told you the way
to tell whether the retry worked was to watch the image count go up. It did not go up, and
that turns out to be correct behaviour: the images had already been made, and what had failed
was publishing them, so a successful retry publishes an existing image rather than making a
second one. I had picked a measurement that watched the wrong half of the process. The retry
did work — eight of the eleven jobs completed.

The costs, for pacing the rest: this site took 43 model calls, about 390,000 tokens in and
121,000 out, and 11 images. Read it as a floor rather than an exact figure.

What I need from you before the fleet goes out is sign-off on that cost and on the pilot
itself — the plan puts that decision before any wave of sites, and it is the right place for
it.

### Tuesday 18 August (later) — correcting the cost figure, upward

The cost I gave you yesterday was wrong, and wrong in the direction that matters. I said one
site cost about 43 model calls and 390,000 tokens in. The real figure is **73 calls and
664,000 tokens in** — about 70% higher. I have now measured it three times and got the same
answer each time, and the site has been quiet since early this morning, so this number is
stable in a way yesterday's was not.

The mistake was mine and it is worth understanding, because it will recur otherwise. I
measured while the site was still building. The way I identify "this site's work" depends on a
field the system fills in as each job progresses, so jobs still in flight were invisible to my
count. I labelled the number a floor, which was the right instinct, but a floor 70% below the
answer is not much of a floor.

**In money: about £3.00 per site today, rising to about £3.80 on 1 September** when an
introductory rate on one of the models ends. Across 140 domains that is roughly £420 now
versus £530 after. I have quoted those in dollars in the working notes — $3.81 and $4.83 per
site, $534 and $677 for the fleet — since that is how they are billed.

Three things that number does not cover, so it is not an all-in figure. The 11 images are not
in it — they come from a different supplier and are not recorded alongside the text costs, so
they are genuinely unmeasured rather than estimated as small. This particular build ran
through the outage and repeated some work, so a clean run should come in lower. And where the
system reuses cached text it is billed at a reduced rate that my arithmetic does not know
about, which pushes the real figure down again.

Your plan of running the next domains one at a time fixes this properly: each one gives a
clean measurement from a quiet start to a quiet finish, and three of those are worth far more
than one figure taken from a build that had an outage in the middle of it.

### Tuesday 18 August (evening) — the pilot is live on its own domain; www is a separate job; and what the "citation queue" actually is

**Your nameserver change worked.** `remortgagecalculator.uk` now serves the pilot site at its
own address — I fetched the page and read the words on it, not just the status code, because a
parked domain returns a cheerful 200 on every path. It is the real thing: the title is
"Remortgage Calculator UK — Your Number, in Seconds" and it is 40kB of the framework's own
markup. That is the first of these sites to reach its own domain.

**But `www` still does not resolve, and the nameserver change was never going to fix it.**
Those are two different jobs. Changing the nameservers at Nominet decides *who answers*
questions about the domain — it hands that to Cloudflare. It does not decide *what the answer
is*. Each of our zones at Cloudflare contains exactly one entry, for the bare domain, and
nothing at all for `www`; and the little program that serves the pages is attached only to the
bare name. So `www.remortgagecalculator.uk` and `www.ai-agent-orchestration.com` both come back
"no such name" — I checked both just now. Fixing it is two small additions per domain, times
36, and it needs one decision from you first: whether `www` should show the same pages or
simply bounce visitors to the bare domain. I would bounce them — one page, one address, no
chance of the two drifting apart.

**Noted on loanzy — and I have written the rule down.** The cleared build is recorded in the
positioning register as P11, along with what I take to be the general instruction: we do not
create sites that present themselves as accredited finance brokers unless you have asked for
one. I want to flag that this makes the flow decision in front of you slightly bigger than the
paper says. My recommendation was "the cheap flow plus an automatic seed", on the grounds that
the missing safety machinery was a seeding problem. That is still true, but it is no longer the
whole problem: an automatic seed would have given that site a claims-checker and a contact
address, and it would still have invented a regulated broker, because nothing in the cheap flow
stops the strategist choosing that identity. So there is a second piece of work — a
prohibition, or a check that refuses the plan — and I have not costed it.

**The citation queue, in plain terms.** We keep one shared directory of facts about outside
companies — mortgage lenders, savings providers, health insurers, AI models — and the rule for
that directory is that every fact must carry a quotation from a public web page that says it.
When the research agent proposes a fact, the system re-fetches the page itself and checks the
exact words are there. If they are, the fact is registered. If they are not, the fact is
**refused** — never quietly kept, never quietly dropped either. The refusals go into the
citation queue, which is simply a list of "the machine would not accept these, a person needs
to rule on them". Nothing else in the system touches it; it waits for a human.

There are four rulings waiting, one per kind of thing we catalogue. The number in the handoff —
"4 items" — is four *lists*, not four facts: between them they hold 27 refused facts, of which
**one** concerns a mortgage lender. So working that queue will not produce more lenders, and I
should not have implied it would.

The single mortgage-lender refusal is a good illustration of the system behaving well rather
than badly. It is a fact about Family Building Society's product range, and the page it quotes
has been reworded since we last looked, so the quotation no longer appears on it word for word.
Two facts about that lender that we verified on Friday are still standing and untouched. The
refusal is the check doing its job on a re-run, not something breaking.

**Where the directory genuinely is thin is the research, not the review.** We have four
mortgage-lender records; two of them were struck out on Friday because they were not lenders at
all but categories — one was literally "FCA-regulated mortgage lenders (general)". So two real
firms: Family Building Society and Mansfield Building Society. The queue cannot fix that. What
fixes it is running the lender researcher more, or aiming it better.

**On your other point — the two builds are not two different machines.** You said one looked
like a page-flow builder and the other like the work-item system. I checked what actually did
the work rather than what the write-up claims, and all three sites — the pilot, the new build,
and loanzy — produced the identical sequence of jobs handled by the identical agents: research,
strategy, briefing, site plan, design, then one job per page and one per image. loanzy's jobs
are marked cancelled because you cleared the site, not because something else built it. There
*is* an agent called `pageflow-builder`, and the classifier still writes its name into a field
called "recommended builder", which is almost certainly where the impression comes from — but
nothing in the current build path reads that field. It is a leftover from an older route in.
So the difference between the two flows really is only what we prepare before pressing go.

### Tuesday 18 August (evening, later) — why the calculator was missing, and the lender directory going from 2 to 25

**The missing calculator has a real cause, and it is not the site's fault.** You
were right to flag it, and right to wonder whether the tools were still being
built — they were not. Nothing was pending; the attempt had already failed three
times and stopped.

Here is what happened, in plain terms. The system keeps one shared library of page
sections. Every component in it has two names: a *type* name, used when the system
goes looking for something to reuse, and a *function* name, used when it goes to
save one. For most components those two names are the same, so nobody notices. For
the mortgage repayment calculator they are not: it has a function name and no type
name at all.

So the pilot asked the library, "have you got a mortgage repayment calculator?" —
and the library, searching by type name, said no. The site therefore set about
generating its own. When it went to save the new one, the save searched by
*function* name, found the existing calculator, and concluded it was being asked to
replace it. It refused — correctly, because that calculator belongs to
loanandmortgagecalculator.co.uk and overwriting it would have blanked that site's
live pages. Three attempts, three identical refusals, and the page was then built,
deployed and served with the calculator simply absent.

Two things make this worse than one missing section. The first is that a page with
one fewer section looks completely finished — there is no visible tell, which is
why it took you noticing the site rather than any alarm. The second is that it is
not a one-off: I found the same collision live on two other sites this evening
(loancalculator.co.uk and loanzy.uk, both blocked by another component of
loanandmortgagecalculator.co.uk's, retrying every few minutes), and 26 components
in the library are in the same state waiting to catch the next site. On a plan to
build 50 finance sites that will naturally want the same calculators, this would
have bitten repeatedly.

The particularly annoying part: a perfectly good mortgage repayment calculator has
been sitting in the library since May, properly typed and ready to be reused. The
site never needed a new one.

I have written the whole thing up as bug 311, with the fix options ranked. I put it
through the independent diagnosis loop before asserting any of it, and that came
back confirmed on the first pass. **The fix itself is a change to shared machinery
that every site's build goes through, so I have not made it while the builds are
halted — it wants your say-so and a review round.** The one that closes the door
properly is to make the save *fork* a copy for the new site instead of colliding,
which is exactly what the tool-deployment path already does elsewhere.

**The directory: from 2 lenders to 25, in about a quarter of an hour.** You asked
whether we should widen the search. We should have, and the reason it was so thin
turned out to be something I would not have guessed, so I measured it rather than
assuming.

I asked the register which *sources* have actually produced facts, across its whole
history. The answer was lopsided: twelve of our thirteen savings providers came
from a single page on gov.uk, and all ten health insurers came from two specialist
broker round-ups. Every source that ever produced more than one firm was a page
that *lists many firms and says something about each*. The mortgage search had
never once landed on a page like that — its four slots went to two market-overview
pages that name firms without saying anything quotable about them, and two
individual building societies' own pages. Two candidates, one registered. That is
the whole story of "only two lenders".

Worse, the instructions we give the extractor actually said third-party round-up
pages were weak sources and to prefer the regulator or the firm — the exact
opposite of what our own history shows.

So I made three changes: each search now reads ten pages instead of four (I checked
the size against a known failure where an oversized response gets silently dropped —
we are running at about a fifth of that limit); I added four new mortgage searches
aimed at the kind of page that actually works, including one specifically for
adverse-credit lenders; and I corrected that instruction.

The four searches ran this evening and registered 42 new facts. **We now have 25
named lenders** — Skipton, Leeds, Yorkshire, Principality, Newcastle, Suffolk,
Teachers, Ecology and other societies, plus the specialist end: Bluestone,
Kensington, Pepper Money, Vida, Foundation Home Loans, Aldermore, The Mortgage
Lender. That specialist list is precisely what adversecreditmortgage.co.uk had
nothing to put on its directory page.

Eight of the new facts failed the verbatim check and went to the review queue,
which is the normal rate and the gate doing its job.

**One thing needs you.** The pilot's lender page is live and still shows two
lenders; it will show all 25 once it re-renders, but the site is locked under the
build halt, so I have not lifted that. Say the word and I will refresh just that
page.

**On www: your nameserver change did work**, and the pilot is now serving at its
own address because of it. But www is a separate job, and I want to be straight
that I got the earlier description of it wrong. It is not that no domain has a www
— eight of the 39 do, in four different states. Two (cookly.uk, dartsonline.com)
already redirect exactly the way you want. One (webdesign.uk) deliberately points
somewhere else and I have left it. Two (robot-hands.com,
leopardessconsulting.co.uk) have a www that simply hangs, and this fixes them. Two
zones are not served by our worker at all, so adding a record there would create a
new broken address rather than a working one, and I have excluded them.

The redirect itself is written and committed. **It needs one command run by you** —
the deploy step is blocked for me by a permission rule, which is reasonable given
that a bad deploy of that one file would take all 39 sites down at once. I have
checked the file against the live version and syntax-checked it (including proving
the checker actually fails on a deliberately broken copy, because it silently
passes broken files otherwise). Once you have run it, I will add the DNS entries
across the 32 zones that need them and verify each one.

### Tuesday 18 August (night) — www is done, on every site

Your deploy went through at 20:02, and the rest followed. **Typing `www.` in front of any
of our addresses now takes you to the site**, rather than to a browser error — on 36 of the
36 domains where that makes sense, verified one at a time by actually following the
redirect rather than trusting what the API told me. It keeps the rest of the address too,
so a link to `www.<site>/some-page` lands on the right page, not the front door.

Three domains I left alone on purpose. Two of them (`idea.uk`, `relojistas.com`) are not
served by our worker at all, so adding the record would have created a new broken address
rather than a working one. The third is `webdesign.uk`, which deliberately sends `www`
traffic to webdesign.co.uk — that looked intentional, so I did not overwrite it.

Two domains were quietly broken before tonight and are fixed as a by-product:
`robot-hands.com` and `leopardessconsulting.co.uk` had a `www` entry with nothing behind
it, so those addresses simply hung.

I also swept every domain's front page after your deploy, because that one file serves all
of them and a bad deploy takes the lot down at once. All fine. Three do not answer, and
none of them is new: `apis.uk` and `ugg2.com` have no site behind them at all — they are
parked — and `loanzy.uk` is the one you cleared. Worth knowing that our records still list
loanzy as deployed even though its pages are gone; that is the other thread's to tidy, but
it will read as "live" to anything that asks.

**Two things nearly made me undo a change that was working**, and I want them on record
because they will catch the next person. First, a freshly created route returns a
particular error — 522 — for the first few seconds, and that error means, in every other
circumstance, "there is nothing behind this address". It settled within a minute. Second,
and stranger: two domains reported "cannot find this address" from my machine for about
four minutes *after* the entry existed and the real DNS was serving it perfectly — my own
machine had remembered the earlier "no such address" answer and kept giving it back. In
both cases the honest-looking reading was the wrong one, and the tempting fix — delete it
and try again — would have destroyed a correct change. Both are now written down.

**The only thing still waiting on you** is whether I refresh the pilot's lender page, which
still shows 2 lenders instead of the 25 we now have. The site is locked under your build
halt, so nothing will do it by itself.

### Wednesday 19 August — what moved underneath us overnight, and a read-out you can hand to someone

About a hundred and twenty commits landed on the shared tree overnight from other threads, so
before doing anything I re-checked the things I had told you were true. They still are: the two
sites are still locked as you asked, the directory still holds 25 lenders, `www` still redirects
on every domain I sampled, and the missing-calculator bug is still unfixed — none of the three
files that would have to change has been touched.

The machine everything runs on was rebuilt overnight and is now several hundred commits further
on than when we spoke. I confirmed that from the running program itself rather than from the
version label, using a check that would have failed if I were wrong: the new build's fingerprint
is present, yesterday's is absent, and a made-up one is absent too.

**One thing got materially worse, and it is worth knowing.** Another thread hit the same
missing-tools bug on a site built from scratch and lost **seven of seven** calculators. They then
did the thing I had not: they looked at the finished page a visitor sees. It is live, it loads
fine, it is 22,000 bytes of prose about a calculator — and it contains no calculator at all. Not
a broken one. None. Nothing on the page tells a reader anything failed.

**And one thing got better.** A different thread had independently walked into the same wall from
the other side and written out exactly how to fix it. Neither of us knew about the other — I
checked, and each document mentioned the other precisely zero times. I have now linked them, and
found something that matters: their fix and ours are the same idea applied to two different parts
of the code, and **their version, as written, would fix their half and leave ours untouched.** If
someone had built it, we would have announced that tools were fixed and been wrong. Both
documents now say so.

**On your question about a summary — yes, and I have written one.** The last one, from Monday,
says the pilot is "built but not published". That is no longer where we are, so the series had a
genuine gap. The new one is
`SUMMARY_2026-08-19_first_sites_live_and_the_wall_the_fleet_would_have_hit.md`, and it is written
to be read aloud: what we are trying to do, where we came from, what we have done, where we are
now, where we are going. The short version of "where we are now" is: the machinery is proved, the
first sites are live, and the thing that would have quietly degraded the next fifty is understood
but not yet fixed.

**To start again in a fresh session, point it at:**
`docs/agent_docs/docs024_key_docs_latest/portfolio_positioning/HANDOFF_2026-08-19_continue_here.md`

That file re-verifies the state at the top, carries the halt and your two open decisions, explains
the wall, and sets out the path in five steps in the order they have to happen. The one thing
still waiting on a word from you is whether I refresh the pilot's lender page — it still shows 2
lenders instead of 25, and it will stay that way while the site is locked.

### Thursday 20 August — the regulated guard is live, and your domain list answered three questions at once

**The guard is running.** Since last night's build, every site the platform builds is checked for
claiming to be an authorised firm about itself, and refused unless there is a record saying
otherwise. That record is the part you asked for: a client who emails proof gets a proper entry —
firm name, registration number, who checked it, what they saw — and then may say it, and the
number itself becomes a fact the system can check their pages against. Both tools are written:
one to record it, one to tell you whether a site is covered and, if not, exactly which detail is
missing.

Three honest limits. The version that shipped covers the main page-building path; the fix for one
other editing route was written today and needs the next build. Site headers and footers are not
covered at all — and the footer is exactly where a firm's registration line normally sits, so
that is a real gap rather than a technicality. And nothing has actually been refused yet, because
both our sites are locked under your halt — so I can tell you it is installed, not that it works.

**The reviewers made this change better twice, and I want to say so plainly.** They caught two
things I had got wrong. The first: I had claimed a particular safety check covered every route
into a page. It does not — one editing route ran no checks whatsoever, which is precisely the
loophole the whole change exists to close. The second was subtler: I said the guard handled
sentences like "we are *not* regulated". It turned out those are safe for a completely different
reason than I thought, and knowing which reason matters, because the one I assumed has a known
fault in it. Neither would have surfaced without the review.

**Your domain list answered three questions in one go.** Of the 1,567 domains: about 1,250 are
parked at marketplaces, 207 have never had nameservers set at all, 14 are the sites we have
built, and **62 are on real hosting and genuinely serving** — those are the live ones, and some
carry substantial sites. Five domains are family names. And I have picked 50 candidates for your
test set, spread deliberately across quite different subjects — audio, insurance, chocolate,
farming, gardens, jewellery, vets, web design — so they exercise different kinds of brief rather
than fifty variations of one. They are candidates until you say yes; reserving a domain is a
decision about an asset, not something a script should settle.

**Three things I had told you were wrong, and you corrected two of them.** The planner does know
what the framework can build — it reads the component library every run, and I simply had not
looked. Games are buildable, as interactive tools; there is a 22,000-character one running on the
game design site. And "no nameserver" means you never set one, not that a registration has
lapsed — my version would have had someone re-buying domains you already own. I have corrected
all three where they were read, not just noted them.

One more, which nobody caught but me: I wrote in the new handoff that a review round had been
submitted when it had not. Counting them is a one-line check. It is submitted now.

**To pick this up in a fresh session:**
`docs/agent_docs/docs024_key_docs_latest/portfolio_positioning/HANDOFF_2026-08-20_continue_here.md`


---

**2026-08-24 (later) — the sitemap generator now has something that actually calls it, except for
the last step, which needs you.**

Back on 20 August you ruled that all future sites should have sitemaps. A sitemap is the file a
site publishes listing its own pages, so a search engine can find them all without guessing. We
built the generator on the 21st. The review panel then failed it, and was right to: the generator
worked, and nothing in the system ever called it. A tool nobody runs is exactly the situation your
ruling was meant to end.

So today's job was to give it a caller. It now has one: a scheduled sweep that picks one site
every half hour, regenerates that site's sitemap, and commits it. A site comes round again after
three days.

**Why a sweep and not "regenerate whenever a page changes".** Both are worth having eventually.
The sweep goes first for a plain reason: it is the only one that helps the sites that have no
sitemap **today**. Regenerating on page-change only ever helps a site that gets edited. There is
also a cost: before listing a page, the generator fetches it to check it really is there, so a
sitemap for our biggest site means 135 fetches. On a sweep that happens once every three days. On
the page-change version it would happen every time anyone touched a page.

**A second problem turned up while I was measuring the first, and it is the more interesting one.**
To prove the work later, I first needed the "before" number — how many sites serve a sitemap now.
The answer is 8 out of 28. Getting that number right was harder than it sounds, in two ways.

First, three sites *look* like they have a sitemap and do not. One returns the file left behind by
the company the domain was parked with. One returns its own homepage for any address you ask for,
so asking for a sitemap gets you a web page with no page list in it at all. A third quietly
forwards you somewhere else. All three would count as successes if you only checked whether the
request succeeded rather than reading what came back.

Second — and this is the one I nearly got wrong — I found that our generator was listing the wrong
address for every site's front page. It listed `/index.html`, while each site itself declares that
its front page is plain `/`. Same page, two addresses; search engines want to be told which one is
official.

I checked one site's homepage, saw the mismatch, and was ready to strip `index.html` everywhere.
That would have been a mistake, and a bigger one than the bug. I checked a *section* page next —
and the sites use the opposite convention there, keeping the `index.html`. Stripping it everywhere
would have fixed 27 addresses and broken 228. What makes that worth writing down is that nothing
would have complained: every broken address still works when you fetch it, so the generator's own
checking step would have passed them all. I have written the trap down in the shared traps file
and logged the near-miss in the mistakes log, because the lesson is general — if you are making a
rule about a pattern, check the pattern in two places before you trust it.

**Where this stands, and the one thing I could not finish.** The fix and the sweep are both written,
tested and committed, and I have sent the whole thing back to the review panel. But the final step —
switching the sweep on, which means one write to the live database — was refused by this session's
own safety permissions. Everything short of it is done and rehearsed: I ran the change against the
live database inside a transaction that throws itself away, and I deliberately broke each of its
four safety checks in turn to confirm they each stop it.

So, plainly: **no site has gained a sitemap yet.** The machinery is ready and one approval away. If
you are happy for me to apply it, say so and it is about a minute's work, after which sites start
picking up sitemaps at one every half hour.

One thing I have deliberately left alone: `adversecreditmortgage.co.uk` is still under your halt
from the 18th, and the sweep skips it. Publishing a sitemap means committing a file to the site,
which is a deployment, and I did not want a background job quietly deploying to a site you have
stopped.


---

**2026-08-25 — it ran all night on its own. Sitemaps went from 8 sites to 26.**

You switched it on yesterday afternoon. It then worked unattended for just over thirteen hours,
doing one site every half hour, and finished at quarter to five this morning.

**Twenty-seven sites, twenty-seven successes, nothing dropped.** I had warned you that a job firing
during the chassis restart could be silently lost — the system marks a site as "done" just before
sending the message, so a lost message leaves a site looking finished when nothing happened. I
checked every one against the record of what actually ran. All twenty-seven line up. The restart
landed at half past nine this morning, well after the last one, so the risk never arose.

**Where we are: 26 of the 28 live sites now publish a proper sitemap, against 8 yesterday.** And
they are all complete — every address in every sitemap is a real page on that site, no
approximations.

The two that don't:

- `adversecreditmortgage.co.uk` — still under your halt from the 18th, and the sweep skips it
  deliberately. It shows the file left behind by the company the domain was parked with.
- `webdesign.uk` — this one turned out to be interesting, see below.

**The homepage fix proved itself by accident, which is the nicest kind of proof.** Your chassis
rebuild went out between the first and second site of the sweep. So the very first site,
`robot-hands.com`, was done on the old code and lists its front page the wrong way; all
twenty-six after it were done on the new code and list it the right way. Same job, same night,
one change in between. I've told the system to redo `robot-hands.com`, which it will within the
half hour.

**`webdesign.uk` was the loose thread, and pulling it found a real bug.** That domain forwards
every address to `webdesign.co.uk`. It doesn't serve anything itself. But our sweep had reported
it as a complete success — seven pages found, none rejected.

The reason is worth explaining, because it is the same mistake in a new place. Before listing a
page, the generator fetches it to check it's really there. The rule was: if the page forwards you
somewhere else, don't list it — a sitemap should name the real address, not one that bounces. That
rule was written down clearly in the code. It was never actually working: the standard way of
fetching a page follows the forwarding automatically, so the checker only ever saw the destination
and reported success every time.

So it looked perfect precisely where it was blindest. The count of rejected pages — the number
whose whole job is to catch this — read zero.

I've fixed it and, before doing so, checked what turning the rule on would break: across every one
of the 27 sites, exactly one page anywhere forwards, and it's on the parked domain we already skip.
So nothing you have loses anything. `webdesign.uk` will correctly end up with no sitemap, which is
right for a domain that serves no pages of its own.

**One more thing, in the interest of not overstating yesterday.** I told you the action was
already live in the running system, and justified it by saying my check's controls had behaved
correctly. I've since found a warning — written the same day by another thread — saying that
particular check can give a false "not there" *while the controls still look right*. My conclusion
happened to be correct, and twenty-seven successful runs have since settled it beyond doubt. But
the reason I gave you for believing it was not a good one, and I'd rather say so than let it
stand. Both of the two things that went wrong this session were in my measuring tools, not in the
work itself.

**To pick this up in a fresh session:**
`docs/agent_docs/docs024_key_docs_latest/portfolio_positioning/HANDOFF_2026-08-25_continue_here.md`


---

**2026-08-25 (evening) — the sweep is now picking up new sites by itself, and the thing I told you
we'd deferred turned out to be live.**

Your new chassis went out at ten past seven. It carries the forwarding fix from this morning.

**First, the good news, and it's the point of the whole exercise.** Three new sites appeared on the
estate since yesterday — `homegarden.uk`, `lampenkap.com`, `cv1.co.uk`. Nobody told the sitemap
sweep about them. It found all three and did them anyway. That is the difference between a tool and
a mechanism, and it is now working the way you asked for on the 20th.

We're at **27 of 31 live sites** publishing a proper sitemap, all complete.

**Now the part worth your attention.** When the review panel approved the sweep, it raised one
objection I told you we were accepting and not fixing: when the generator finds nothing to list, it
goes quiet, and it goes quiet for two opposite reasons — the site opted out, or something is wrong.
I judged that a low-priority gap.

It was live, on two sites, today. `homegarden.uk` and `cv1.co.uk` were both visited by the sweep
**before their pages had finished publishing** — two hours before, in one case. The generator found
nothing, correctly refused to publish an empty file, and the system then marked both sites as
"done". They'd have sat with no sitemap until Friday, and nothing anywhere would have said so.

**The panel's objection was right, but it named the wrong cost, and that mattered.** It said the
problem was a missing warning. It isn't — a warning wouldn't have helped either site. The problem
is that a visit which achieved nothing still *used up the site's turn*. So the fix belongs in the
part that chooses which site to visit, not in the reporting. That's now in place: a site with
nothing published yet simply isn't chosen, so it keeps its turn until it's ready.

I've put both sites back in the queue by hand.

**Two of my own checks were lying to me today, and one of them was already in a handoff.**

The first was a query I wrote *yesterday* to detect exactly the kind of silent failure above. Run
today it reported six sites as failures. All six were fine. The evidence it checks against is
deleted after 24 hours, so anything older than a day looks like a failure. The dangerous part is
that the fix I'd written next to it was "put the site back in the queue" — which against six
healthy sites means redoing work for nothing and concluding the system is broken when it isn't.

The second I caught before it got anywhere: hunting for other affected sites, I used a "published
since" date that turns out to be rewritten every time a page is republished. It told me 24 of 24
sites were affected. The real number was two.

Both are now written down where the next person will hit them. I'd rather flag that three of the
problems this week were in my measuring tools than let the tools keep their reputation.

**Where it stands.** Two fixes went live in the last hour — the forwarding one and the turn-taking
one — and neither has yet been confirmed on a real site, because the sweep only runs every half
hour and three sites are queued ahead. That check is running now and the recipe is written down, so
it survives this session either way.

**To pick this up in a fresh session:**
`docs/agent_docs/docs024_key_docs_latest/portfolio_positioning/HANDOFF_2026-08-25b_continue_here.md`


---

**2026-08-25 (late) — finished. Every site that should have a sitemap has one.**

All three checks came back over the last hour and all three were what I predicted.

`webdesign.uk` — the domain that forwards everything elsewhere — now correctly produces **no**
sitemap. It found the same seven pages as before, rejected all seven because they forward, and
published nothing. That is the right answer, not a gap: it stays on the uncovered list permanently
and deliberately.

`homegarden.uk` and `cv1.co.uk` — the two that had been quietly skipped because the sweep visited
them before their pages went up — are both live: 20 pages and 3 pages respectively, every address
matching a real page.

**We're at 29 of 31.** The two that aren't covered are `adversecreditmortgage.co.uk`, which is under
your halt and deliberately skipped, and `webdesign.uk` above. So there is no site left that ought to
have a sitemap and doesn't. On Monday morning it was 8 out of 28.

**One thing I want to be straight about, because it would be easy to let it slide past.** The fix I
made this evening — the one that stops a site being visited before it's ready — **has not actually
been exercised yet.** Both sites recovered because I put them back in the queue by hand, and both
now have pages, so the new check isn't even consulted for them. It will get its first real test when
the next brand-new site appears. I've written down exactly what to look for so nobody has to take my
word for it: a site with no pages published that has still been marked as visited means the fix
didn't work.

I mention it because "29 of 31, all green" is exactly the sort of result that lets an untested piece
of work ride along unnoticed.

**Also worth knowing:** the review panel approved the fix but flagged that my database edit used a
find-and-replace which would have hit *every* match if the phrase appeared twice. It appeared once,
so nothing went wrong — I checked. But the criticism was correct and I've written down the cheaper
check that would have removed the doubt entirely.

**To pick this up in a fresh session:**
`docs/agent_docs/docs024_key_docs_latest/portfolio_positioning/HANDOFF_2026-08-25b_continue_here.md`



---

**2026-08-26 — a changed page now reaches its sitemap in about an hour, not three days.**

Yesterday's sweep visited each site every three days. That was fine for getting every site a sitemap
in the first place, but it meant a page you published — or took down — could sit unlisted for up to
three days. The original plan had put this off with a costing: "if we regenerate on every edit,
that's 136 checks on the biggest site, every time anyone touches anything."

That costing was right about the wrong design. We didn't need to regenerate on every edit. The sweep
already has a hard ceiling — one site every half hour, so at most 48 a day no matter what — and all
we had to do was tell it *which* sites to pick. So: a site whose pages have changed since its last
visit now goes to the front of the queue, once it has been quiet for half an hour (so a big batch of
edits settles into one visit instead of several). The three-day rule stays as a floor underneath, so
a site that is edited constantly can never be starved.

**It was proven the same day, and the proof was built in.** Every site had been visited in the last
two days, so under the old rule nothing was due before tomorrow afternoon. Anything picked before
then could only be the new rule. By mid-afternoon eight sites had been picked, one per half hour,
all regenerated cleanly and published — and the first one's live sitemap carries today's dates on
exactly the two pages whose edits triggered it.

The review panel approved it in ten minutes with two minor notes. One said I should check there was
only one copy of the scheduling row before editing it — I already had, in the file, but I'd left it
out of the summary the panel reads, so they couldn't see it. Fair; I've written down that the
summary must list every check, not just the interesting one.

**Also today:** the old hand-run generator listed every site's homepage by the wrong address
(`/index.html` rather than plain `/`). The automatic version was fixed on Sunday; the hand-run one is
now fixed too, so nothing left in the estate writes the wrong form.

**One thing that isn't ours but you should know:** running that tool against `cv1.co.uk` showed
Cloudflare's managed robots file is being merged into ours and is currently turning away the AI
crawlers (ClaudeBot, GPTBot, Google's, and others). Whether we want that is your call, not mine; I
haven't touched it.

**Still unproven, same as yesterday:** the check that stops a brand-new site being visited before
its pages exist. Nothing new has appeared to test it against.

**To pick this up in a fresh session:**
`docs/agent_docs/docs024_key_docs_latest/portfolio_positioning/HANDOFF_2026-08-26_continue_here.md`

**2026-08-26, evening — the remaking of your hosted sites has started. There is a brief waiting for you.**

First, the morning's work is holding: the new "follow the deploy" sitemap rule kept running all day
— sixteen sites picked, one every half hour, every one regenerated and published cleanly. The one
oddity in the numbers turned out to be deliberate: idea.uk's privacy page is served at a different
address than our records say, by a decision that lane made in July, so the sitemap correctly leaves
it out. Nothing to fix.

Then the bigger step. You ruled last week that 22 of your hosted sites can be completely remade, and
the defect that was blocking building at scale was fixed two days ago. So tonight the machine wrote
the first real brief: **advertise.co.uk**, the first of the single-pagers you cleared, picked exactly
in the order we agreed (small, strong generic names first; the insurance one last).

The brief proposes making it the plain-English reference on advertising itself — what it is, how it
works, who pays for it — for UK business owners and the curious, explicitly *not* an agency and not
a site that sells advertising. It stays off the ground of your other domains by name (website
promotion, SEO tools, web design are their territory, not this one's). And it asks you five good
questions, the two that matter most being: is that broad "explain advertising" reading what you
want for this name, and is advertise.co.uk meant to be the parent of your marketing cluster? Your
answer to the second one changes how the neighbouring sites' briefs should be written — which is
why I stopped at one brief tonight instead of firing all five.

**Nothing will be built until you read the brief and release it.** It is held the same way the two
test briefs from last week are (those are still waiting too, if you want to see how the machine
handled a houseplant domain and a mortgage one). Before anything can overwrite it, I saved a copy of
what advertise.co.uk serves today — it turned out to be an old feed-aggregator page with no original
writing on it, so nothing of value is at risk.

**2026-08-26, late evening — a leftover test sentence is live on lendzy, and I've pulled it out at the root.**

One more thing tonight, found by accident and worth telling you straight. Back on 2 August, when
lendzy was an experiment, we planted a test sentence in its writing instructions — "checked against
the FCA handbook, rule by rule" — purely to verify the machine follows instructions. The note to
remove it before the site ever went live was written down, and then the site was built and went
live weeks later without anyone re-reading that note. So today lendzy's About page and one guide
say, in the site's own voice, that its content has been checked against the FCA handbook rule by
rule — which nobody has done. On a site about borrowers' rights, that is a claim we should not be
making.

I've removed the instruction at its source tonight, so no future rewrite can plant the sentence
again, and filed the clean-up of the three passages that already carry it (bug 414). One caution
for your review queue: the maintenance system read that sentence on the live page and concluded it
was the site's core selling point — there is a held item asking to lean into it further. When you
work through lendzy's queue, that one should be rejected, not released.

**2026-08-27 — a correction to last night's lendzy note: I had not, in fact, pulled it out at the root.**

Last night I wrote that I'd removed the planted test sentence at its source so no future rewrite
could plant it again. That was wrong, and another session caught it this morning. Between the
experiment and now, one of the automatic planners had read the planted instruction and restated it
in its own words in a second planning document — one I checked, but cleared after reading only the
first thing that matched in it. So for the last ten days the system still held an instruction to
include the sentence, just written differently, somewhere I'd looked and not seen.

The other session has now cleaned every planning document fleet-wide (I've re-checked that myself
against the live system: nothing anywhere still mandates the phrase), rejected the held queue item I
warned you about last night (so that's off your plate), and shipped a proper fix so the claims
machinery recognises this whole family of "everything is checked/verified" sentences rather than
just this one. The two pages still carry the sentence at this moment; the rewrite that removes it
is dispatched and they'll verify it on the live pages, not on a job status. My mistake and its
lesson are written up where we log wrong calls.

**2026-08-31 — quiet few days, one hiccup fixed, and the next move is yours: three briefs are waiting.**

Checked everything over after the new platform build went out. The sitemap machinery came through
the upgrade fine — still picking a site every half hour when one needs it, still publishing
cleanly. One hiccup: oufe.com's sitemap refresh failed in the small hours (a timeout talking to
the git service, a one-off — every run after it worked) and that left oufe serving "not found"
for its sitemap. The failed attempt also used up oufe's turn in the queue, which would have meant
a three-day wait; I put it back at the front of the queue and watched the next sweep restore it — oufe's sitemap is
serving correctly again, all 19 pages listed, checked on the live site before writing this.

The lendzy test-sentence saga is essentially over: the live pages no longer carry the phrase
anywhere (I checked the pages themselves this morning), and the session handling bug 414 just
needed this platform build to land their permanent fix — closing it is their call now.

**The thing only you can move: the remake of your hosted sites is waiting on you.** The brief for
advertise.co.uk — the first of the 22 — has been sitting ready since Tuesday evening, alongside
the two test briefs from the week before. Nothing builds until you read it and release it. The
one question in it worth answering even if you change nothing else: should advertise.co.uk be the
flagship of your marketing-related domains, or stand alone? Your answer decides how the next few
briefs get written.

**2026-08-31, later — your brief comments are recorded, and the advertising-marketplace future is safe.**

Your four points are written up as a decision file in the lane. The short answers: the
copy-but-improve instinct is genuinely in the advertise brief already (its UK rules guide, channel
matcher and directory all exist explicitly because the competition's versions are worse or
commercial), but nothing guarantees that for every future brief yet — the machinery for that is
the "best in class propagation" plan another thread has ready and waiting on your go, so I've not
built a duplicate; until it ships I'll carry the instruction by hand into every brief request.

On advertise as a future marketplace for your network's ad space: nothing in the current brief
forecloses it. Three sentences will need deliberate amendment when that day comes (the ones saying
it doesn't sell advertising), the directory takes your own offerings as ordinary entries whose
prominence we control, and the one thing I'll watch at build time is that "we don't sell anything"
never gets written into the site's permanent page furniture — that's the lendzy lesson: identity
claims baked into every page are expensive to walk back.

On fullness: the brief is bigger than it reads — three of its line items fan out (six platform
guides, a 15–20 entry directory, a news feed), so it builds to thirty-to-forty-plus pages, upper-
middle of the current fleet. But the two most "full-feeling" features — the news digest and the
glossary — are sitting in the lowest-priority tier. I've prepared the edit that promotes them and
spells out the fan-outs; one word from you on release and it's applied.

Still waiting on the one question only you can answer: is advertise.co.uk the flagship of the
marketing cluster? That's what the next three briefs are queued behind.

**2026-09-02 — your three answers are in effect: the brief is edited, and the cluster briefs are being written.**

The advertise brief now carries your changes: the glossary and the news digest are promoted into
the plan proper, the platform guides are spelled out as six pages, and — per your ruling — the
brief now explicitly forbids "we don't sell…" style statements in the site's copy, with the old
"it does not sell advertising" framing removed. Your original machine-written brief is preserved
underneath, so the difference is auditable. **The build is still held: you've edited, you haven't
yet said "go build". One word does it.**

Your ruling about negative-identity claims is recorded as fleet-wide and handed to the thread
that owns copy standards (the same one whose best-in-class plan is waiting on your go — these two
rulings will travel together). The legal/finance exception is noted; lendzy's "not a lender" line
stays.

And with your flagship answer, the three marketing-cluster briefs are being written now —
website promotion, SEO tools, and design blog — each told advertise.co.uk is the flagship, each
told to own its own patch, match and better what its competitors offer, and plan a full site.
They'll hold for your review exactly like the first one. (Small note: you wrote "advertise.uk" —
we don't own that domain; I've taken it as advertise.co.uk throughout.)

**2026-09-02, later — you said "go build advertise", and the build is running.**

Thanks for confirming the domain — advertise.co.uk it is (advertise.uk isn't in the portfolio,
so everything was already pointed at the right one).

The release went exactly as the machinery was designed to work. Your review item is closed with
your approval recorded on it. The build queue was handed the "research and classify this domain"
task that starts every framework build, and the site row was unlocked as the last, deliberate
step — so nothing could start until everything was in place. The pipeline picked it up within
thirty seconds of the unlock, and as I write this the classifier is reading the site's specs —
including your edited brief, with your changes and the original preserved underneath.

This is the first of the 22 remakes to go through the framework end to end, so we're watching it
closely: research first, then strategy, the content plan, the pages themselves, and finally
deployment to the live domain. The old single-page site (the Drupal feed aggregator) is safely
snapshotted and will be overwritten when the new site deploys. The guard we agreed on stays in
force at build review: no "we don't sell advertising" style copy anywhere on the served site, so
the door stays open for selling space on the network later.

The five other briefs (three real, two test) are still in your queue, holding for your review.

**2026-09-02, build afternoon — the advertise build is deep into pages; one small thing waits on you.**

The build got through research, strategy, planning and all seven interactive tools, and is now
writing pages and imagery — roughly forty pages' worth queued and draining. Two hiccups so far,
both handled: a data gap from when the site row was first created (fixed, and fixed for the five
other briefed sites so their releases won't trip on it), and one page draft that said "more tools
coming soon" — the platform's own checks blocked that as placeholder text, and it's simply being
rewritten.

One item is yours when you next look: the negation checker flagged the advertise brief for one
phrase — "tell the reader what to do with the information, not just what the information is".
That's a writing-style instruction, not a "we don't sell anything" claim, so my recommendation is
to leave it exactly as it is; but the checker routes that decision to you by design. Say the word
and I'll either clear it as accepted or edit it out.

**2026-09-02, late afternoon — the build is paused on Anthropic credits, not on anything wrong.**

Fourteen of the eighteen planned pages were built when the Anthropic account ran out of credits
(14:43 UTC) — every AI call across the whole fleet now fails with "credit balance too low", so
this pauses everything, not just advertise. The build queue is designed for this: the failed
pages re-queue themselves and will resume without any hand-holding once credits are back.

The one thing only you can do: top up the credits. One gotcha from last time — the fleet's API
key is not on the org the console shows you by default; pick the right account by checking which
key shows recent "Last used" activity. I'm watching for the first successful call and will
confirm the build resumes.

**2026-09-02, after the top-up — your three cluster sites are building.**

You said the three briefs are good, so all three are released and building: website promotion,
SEO tools, and the design blog. Same machinery as advertise, plus the lessons from this morning
already applied — their site rows went in complete this time, including contact emails, so they
won't stall where advertise did. The build queue works them in alongside advertise's remaining
pages. Your review queue is now down to the two test briefs and that one style-phrase question
on advertise.

**2026-09-02, evening — advertise is built and deployed; one step remains before the world sees it, and it's yours.**

The first framework remake went end to end today: brief to research to plan to seven working
interactive tools to twenty-four pages, through one credit outage and a handful of self-healing
retries, to a deployed site — verified by reading the actual built pages, not the statuses. The
homepage reads "Advertise.co.uk — The UK Guide to Advertising", the glossary and news feed from
your fullness edit are in, and there isn't a "we don't sell advertising" anywhere on it, so the
door to selling your own network's space stays open.

The catch: advertise.co.uk the domain still points at the old hosting, so visitors still see the
old Drupal page. The new site sits ready behind the platform's serving worker. To cut over:
the domain needs a Cloudflare zone, then the nameserver change at Nominet — that second step
only you can do. The same will be true for the three cluster sites when they finish, so it may
be worth doing all four domains as one Nominet batch, which is the bulk approach you ruled for
anyway. If you create the DNS-scoped Cloudflare token the runbook describes (the correction from
2026-08-18 has the exact permission recipe), I can script everything except the Nominet step.

**2026-09-02, night — all four sites are built and deployed. One Cloudflare step stands between them and the world.**

Advertise, SEO tools, website promotion and the design blog all went through the full pipeline
today — the whole first wave of remakes, built in a single day. The nameservers you set are the
right ones; the missing half is that the four domains need adding as zones in the Cloudflare
dashboard (Add site → each domain, free plan). Until then all four domains answer nothing at
all — old sites included. The moment the zones exist I add the worker routes; each zone also
needs one proxied A record (dashboard, or give me the DNS token and I do it all).

**2026-09-02, late — they're live. All four.**

advertise.co.uk, websitepromotion.co.uk, seotools.co.uk and designblog.co.uk now serve their
new sites to the public — I've read the live pages themselves: the right titles, the tools
serving, the old Drupal pages gone, and nothing anywhere saying "we don't sell advertising".
Brief to live site in one day, four times over, through every check the platform owns.

Small things that finish themselves over the next day: sitemaps appear on the next rotation
pass, seotools' seven tools arrive via the discovery sweep, and a few final rerenders drain.
I've written today up as a milestone summary you can read aloud
(SUMMARY_2026-09-02_first_four_remakes_live.md). Eighteen remakes to go.

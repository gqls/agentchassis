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

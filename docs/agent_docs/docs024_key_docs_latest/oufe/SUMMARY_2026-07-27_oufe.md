# SUMMARY — oufe.com and oxenunity.com, 2026-07-27

Second in the series. The previous one (`SUMMARY_2026-07-26_oufe.md`) is still
accurate about the sites; this one exists because **a central conclusion in it
turned out to be wrong**, and correcting it changed what we should build. That is
an inflection rather than a day's progress, which is the bar for writing one.

---

## What we are trying to do

Unchanged. A publication about how corporate finance works when a company is under
strain — restructuring, distressed debt, liability management — focused on the UK,
where the courts and regulators publish. Mechanism explained clearly, with tools
that let a reader move the assumptions and see who wins and who loses. Alongside
it, oxenunity.com: one page carrying the name, claiming nothing.

## Where we have come from since yesterday

Yesterday's summary said the site had made a mistake worth the whole exercise: its
copy wrote a promise of its own infallibility, and I concluded that **no check we
own could ever have caught it** — that every verification layer polices claims
about other people, mostly numeric, and a qualitative claim about ourselves was
invisible to all of it.

On that premise I proposed building two new things: a sweep over live content, and
a promise register.

**The owner refused the answer.** He said we already have functionality that
double-checks claims, and we have the council, and told me to look hard at what
exists before building anything. He was right, and the research took ten minutes.

**The banned-claim scanner is an ordinary text search over prose.** It has no
numeric restriction — the restriction I had found applies only to the *number*
scanner, a few lines away in the same file. It catches whatever patterns a site is
given. It would have caught all four phrases. Other sites already use it for
purely qualitative things: "leaderboard", "live now", "price target".

Nobody had ever written a pattern for this class. On any site. That is the whole
of it: **the capability was never missing, the reach was.**

## What we have done

**Armed the scanner on oufe.** Ten patterns, tested in both directions before
applying: every fabrication shape blocked, including all four phrases the site
shipped, and thirteen legitimate sentences pass untouched — among them the honest
replacement copy and the approved disclaimer's own wording. That mattered, because
a false positive here fails an entire page build. One database update, no
deployment. It buys a build-time blocker and a high-severity finding on live pages.

The line those patterns draw is worth keeping, because it generalises: **a site
may describe what it does; it may not claim what that guarantees.** "We cite every
figure and date it" is a process we control and can keep. "A claim without a source
does not appear here" asserts a completeness nobody can verify, including us.

**Corrected the error where it had spread**, which included a live agent. My wrong
conclusion had gone into the standing instructions of the compliance reviewer,
telling it *"no scanner will catch this, so this seat is the only control"* — which
invites the seat to stand in for a mechanism instead of asking for one. Fixed and
mirrored to both rosters, and corrected in place in the original migration, the
previous summary and the decisions register, with the original reasoning left
visible. The content of the error was itself a confident unverified claim about
what our system guarantees, written by the person building the overclaim detector.

**Filed the real defects**, which are about reach rather than capability:

- Only **5 of 15 live sites** carry a single banned-claim pattern. The ten without
  include **vetcomparison.uk**, where we published fabricated prices for three
  thousand named real businesses, and **idea.uk**, the only site taking money. A
  lesson learned on one site cannot reach any other, and every new site is born
  unarmed. This decision was **already filed and deferred** — "per-site only until
  two sites have evidence bases" — when there was one site. There are eight. A
  deferral with a numeric trigger and nobody watching it is indistinguishable from
  a decision.
- The register defines a field for what kind of claim each fact is —
  metric, capability, entity, attestation — and **nothing in the platform ever
  reads it**. That empty slot is exactly where a promise belongs: a capability
  claim whose source is the mechanism that keeps it. Its invisibility is why I
  nearly proposed building a third thing that models the same idea.
- The post-deploy sweep that would catch drift on published pages is reachable
  only through a scheduled task **disabled since early May**. It is not broken and
  not unwired; it has no cadence. Contributed as evidence to the existing bug
  rather than filed as a competing one.

**Finished the grounded content lane**, which was the other half of the week's
work. It searches for sources rather than recalling them, requires a verbatim
quote for every claim, re-fetches each source and discards any quote that is not
really there, writes only from what survived, audits its own draft, and stops at
human review with no setting that lets it publish.

Its first real run acquired **nineteen verified citations, most from the statute
itself**, and the audit then caught two sentences the draft could not support —
both plausible legal generalisations that read as settled law and would never have
been questioned by a reader. That is the lane doing exactly its job, and it is
also the first time the citation-verification machinery has completed end to end
since it was switched on.

## Where we are now

Both sites serve, with no broken links and no unverified figure anywhere on them.
oufe's register holds nineteen quote-verified facts and twenty-eight banned
patterns. The disclaimer is approved through section F, with the liability cap
drafted and awaiting a read.

The honest position on the machinery: **generation and review are covered
fleet-wide; detection is covered in principle and armed on a third of the estate;
and nothing yet checks that a promise we make has a mechanism behind it.** The
last of those has a designed home that was never wired up, not an absence.

## Where we are going

The immediate content work is the remaining mechanism explainers through the
grounded lane, then the Thames Water dossier once its evidence is gathered.

The immediate platform decision is whether to make the universal patterns
fleet-wide — the change is small and the precedent for it sits one directory away
in the sibling engine — and whether to give the live-page sweep a cadence, which
is a larger call because it re-arms other automatic behaviour with it.

And the decisions register carries ten open items, of which the audience question
still blocks the most: whether this is written for students, for the mid-market
professional, or for anyone learning how this works. Pages are being written now,
so it is the cheapest it will ever be to answer.

## The thing worth carrying out of this week

The most useful finding was not the site's mistake. It was that **I diagnosed a
missing capability where there was an unused one**, and would have built a
redundant subsystem if the owner had accepted the answer. The estate had the
mechanism, a filed decision about extending it, a precedent for how, and an empty
slot waiting for the exact case — and none of that was visible from where I stood,
because a thing that does nothing points at nothing in the code.

"Nothing in the estate does X" is an enormous claim. It needs a search, not an
inference, and the tell that I had not done one was in my own sentence: it named
four mechanisms and I had read the source of one.

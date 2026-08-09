# Evidence freshness across the framework — what it is, what it catches, where it stops (2026-08-09)

*Written by the mortgagecalculator adoption lane at the owner's request, as a
new file in this workstream's series. Figures measured live 2026-08-09; the
design history is `SPEC_claims_verification.md` (V4 §5) and
`SPEC_V5_researched_citations.md`; the programme story to V2 is
`SUMMARY_where_we_are_claims_verification.md`.*

## What it is

Evidence freshness is the arm of the claims-verification layer that keeps a
site's registered facts true over time. Each participating site carries an
evidence base — a machine-readable register of verified facts, each with its
value, its source, a verified-on date and a tolerance — and a scheduled task
(`evidence-freshness`, daily, enabled, last completed today) sweeps every site
that has one. Participation is opt-in by the register's existence: a site
without an evidence base is untouched.

## How a fact stays fresh

What the sweep can do depends on what kind of source a fact declares.
Measured today across the fleet: **115 facts on 11 sites** — 39
citation-sourced, 29 SQL-sourced, 28 artifact-sourced, 18 human attestations,
1 unsourced.

- **SQL-sourced facts** are re-run mechanically. The query is the truth: the
  sweep updates the stored value and date, and raises a review item when the
  live number moves outside tolerance — including when the site now
  *under*-claims, since copy saying 2,767 while the database says 3,104 is
  also worth a look.
- **Citation-sourced facts** — quotes from an external page, including
  legislation — are re-verified by re-fetching the cited URL and requiring the
  stored verbatim quote to still appear in the page's visible text (matching
  survives cosmetic changes: case, whitespace, curly punctuation, thousands
  separators). If the wording moves or vanishes, that is a `citation_lost`
  finding. Each citation also carries a staleness clock that forces periodic
  human re-attestation even when the quote still matches.
- **Artifact and attested facts** age by the staleness clock only — there is
  nothing live to re-run.

Two design rules hold throughout. Truth decisions stay human: the sweep
updates numbers from their own declared sources, but every drift finding parks
as a review item; it never rewrites published copy. And the writer whitelist
stays current: for sites that opt in, the sweep regenerates the block of
facts the page writer is permitted to assert — humans own the words, the
machine owns the numbers.

## What it has actually caught

The sweep has raised real findings on five sites since late July: twelve
facts drifted on oufe.com in one pass (27 July), single-fact drifts on
leopardessconsulting and vonc.com, two on ai-agent-orchestration, five on
fundamentallyai — all parked for human ruling, none auto-rewritten. Its
build-time siblings (the banned-claims gate and post-deploy scans) have a
longer record, including catching a stranger thread's invented service within
three hours of going live. The one infrastructure bug that could have made
the schedule a silent no-op (074, a scheduled task's inline workflow being
ignored) was found, fixed and closed in July — the shape can no longer be
authored.

## New this week: watching legislation

mortgagecalculator.co.uk became the first site to register **tax law** as
citation facts: the SDLT bands, the first-time-buyer thresholds, the £500,000
relief cliff, and the additional-property surcharge, each quoting GOV.UK
verbatim. The motivating incident is instructive: the site's original
stamp-duty calculator still grants first-time-buyer relief between £500,000
and £625,000 — rules that ended in April 2025 — under-quoting a £595,000
purchase by £5,000. A sister site's calculator was separately found running a
tax rule sixteen months expired. Calculators encode legislation, legislation
moves, and until now nothing watched. From its next daily pass, the sweep
re-checks those GOV.UK pages and raises a finding when the Treasury moves a
threshold.

## Where it stops — the honest limits

- **It guards copy, not code.** A registered fact constrains what writers may
  assert in prose; nothing yet connects a fact to the constants inside a
  calculator's JavaScript. A tool encoding a stale threshold passes every
  current check. The designed fix — an acceptance check that computes expected
  answers from the fact register itself — is the most valuable unbuilt piece,
  and touches a shared platform seam, so it goes through the council gate.
- **It detects moved wording, not new law.** Quote-matching notices when a
  cited page changes; a new act published on a page nobody cited will not trip
  it. The staleness clock is the backstop: it forces a human re-attestation on
  schedule regardless.
- **Coverage is thin where it is newest.** Citation facts exist on only a
  handful of sites; most registers are SQL and attestation facts about the
  owner's own business. Legislation-watching is proven in mechanism but one
  site old.
- **Publishing is the natural next step, not yet taken.** A "current rules"
  page whose numbers can only come from the register — correct by
  construction, refreshed as the facts refresh — is agreed in principle for
  mortgagecalculator and applies to any site whose subject is regulated.

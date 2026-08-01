# SUMMARY — the portfolio has a positioning framework, a register, and a guard

**2026-08-01.** First summary in this workstream, written the morning after the framework
shipped. Current state only — the chronology is in `README_where_we_are.md`, the evidence
in `NOTES_portfolio_positioning.md`.

## What we are trying to do

Turn roughly 150 finance and insurance domains into substantial sites that genuinely
differ from each other — different target markets, ages, stages, purchase sizes and types,
levels of experience — with the differences **enforced by configuration from the start**,
not remembered by whoever writes the next page. The reason this needs machinery rather
than good intentions: sites this close in subject will drift back together unless
something structural stops them, and we have confirmed the platform contains nothing that
would notice if they did.

## Where we have come from

Two sites are already live and framework-managed: `loancalculator.co.uk` (another thread's)
and `loanandmortgagecalculator.co.uk` (this thread's — built, verified, adopted, and its
positioning written into the fields the content pipeline actually reads). A third,
`mortgagecalculator.co.uk`, is handed off for adoption with its own cold-start document,
which leads with a live-outage hazard specific to that domain. The positioning work on the
live pair produced the pattern everything else now copies: a site's direction as a spec
the writer reads, with an explicit rule naming its neighbours and what separates them.

Then the owner supplied the full portfolio and the requirement that shaped this
workstream: thick sites, deliberately divergent, insurance included, one site per real
proposition.

## What we have done

**Counted before designing.** The 152 distinct domains collapse to **42 propositions** —
42 different promises a name can make to a visitor. Only two domains have framework
records, so roughly 150 are greenfield and the direction is being set at the cheapest
possible moment. The decisive measurement: **a third of the finance names encode no
audience whatsoever**, so differentiation has to be assigned and recorded, because for
those domains the name assigns nothing.

**Built the axis catalogue** — every angle we could find for telling two finance sites
apart: who the visitor is (age, credit history, sophistication, person or company,
wealth), what they are buying (size, type, structure, risk posture), when in their journey
they are (with "preparing" and "owning" identified as the near-empty ground), what job the
site does, its stance and emotional register, and which side of the market it faces, plus
a compliance axis for the insurance family. The single most useful finding: **the grammar
of a domain name assigns its job** — *rates* names want live tables, *quote* names want
estimators, *best* names want ranked verdicts with the method shown, *which* names want
decision trees, and *forecast* is predictive. That observation carves the twelve
savings-rates domains, which looked like one query, into eight genuinely different sites —
the worked proof in the plan.

**Wrote the register**: one entry per proposition, all 152 domains claimed exactly once,
each entry carrying its audience, mode, stance, what it will not cover, and its named
neighbours with the separating sentence. A machine-readable claims table is the contract;
the prose is for humans.

**Built the guard the platform lacks**: a check that refuses the register if any domain is
claimed twice or not at all, if any entry fails to say where its ground ends, or if two
entries claim the same audience doing the same job. It runs green on the real register and
was deliberately broken to prove it goes red — and while being built it caught its own
author twice, which is the reason to trust it.

## Where we are now

The framework, register, portfolio list and check are committed, the workstream's standing
documents are open, and the rule every future site lane must follow is recorded in the
register, the runbooks and the memory index: **register row before build or adoption** —
because adoption auto-writes a generic identity that was measured near-identical between
two live siblings, which is precisely the convergence this exists to prevent.

Honestly stated limits: the clusters were assigned by reading the names, not from
search-volume data, so keyword numbers may refine build order but should rarely move a
domain between propositions. And the register is the *only* overlap mechanism — nothing
downstream will warn anyone if discipline lapses; the check has to be run.

## Where we are going

**Three decisions wait on the owner**, all rolled up in the register: the ~40 twin-pair
calls (spelling variants with no differentiating axis available — each defaulted to "301
to the built sibling", any can be overruled); whether `loancash.co.uk` is built at all
(recommended not — the name attracts the audience the regulator protects hardest); and
build order across the 42 propositions, which is a commercial call.

**Then the machine starts**: for each proposition in the chosen order — register row
first, then build or adopt through the proven path (deploy repo before adoption, the byte
gate, the rerender hold), then the positioning written into the live specs through the
gated script. The two open items on the crossing-point site continue in parallel:
decomposing its 13 guides so the framework can genuinely evolve them, and then proving
end-to-end that a positioning spec changes what gets written — the acceptance test the
whole divergence design ultimately rests on.

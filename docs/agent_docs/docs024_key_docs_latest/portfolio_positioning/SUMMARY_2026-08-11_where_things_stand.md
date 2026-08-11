# SUMMARY — portfolio positioning, 2026-08-11: where things stand before the big build

## What we're trying to do

Build out roughly 150 finance and insurance domains as thick, differentiated sites through
the platform's own generative pipeline — not hand-built, not thin exact-match pages — so
that visitors arriving at each one meet a genuinely different site, not the same content
and the same look wearing a different domain name.

## Where we've come from

On 2026-07-31 the owner handed over the full domain list and said the thing that mattered:
different directions, decided from the start, not patched in later. That produced a
34-axis classification framework and a register — one entry per proposition, naming who
each site is for, what job it does, what it will not cover, and its nearest neighbours —
checked mechanically by a script that refuses the register if any domain collides or goes
unclaimed. By 2026-08-02 the pipeline had proven, on a shadow build (lendzy.co.uk), that a
positioning entry actually changes what the writer writes. That run also surfaced six gaps
between "the pipeline runs" and "the pipeline produces the best site it can" — structural
link-checking, tool correctness, fact discipline, truncation guards, the fidelity dial, and
positioning itself only being a suggestion rather than a setting.

## What we've done

**The register is decision-complete.** 152 domains, 43 propositions, 27 assigned
coordinate sets, and — the piece worth saying plainly today — **the twin-pair question the
owner didn't want to rule on domain-by-domain is already answered as policy, not as 40
individual calls.** Two owner rulings in the first days of August did that: cross-TLD pairs
(`.co.uk`/`.uk` of the same phrase) both get built, split by depth — `.co.uk` is the
authority, `.uk` is the instrument — and same-TLD spelling twins are split by *seat*
(buyer/setter/intermediary/observer/referee: the same phrase means something different to
the saver, the bank pricing the product, the broker, the analyst, and the compliance
reader). That seat-map ruling explicitly **retired the per-pair owner-call default** for
every twin it could reach. Only **two** domains resist it entirely
(`besthealthinsurancerate.co.uk`, `bestlandlordinsurancerate.co.uk` — even five seats can't
separate seven spellings of one phrase) and still need a hold-or-301 call. Everything else
in the register runs on the policy, not on a decision queue. (Housekeeping owed, not
blocking: ~33 older inline `⚑OWNER` markers scattered through the register's prose predate
that ruling and should be swept to match the rollup table, so the prose doesn't keep
flagging a decision that's already made.)

**Content diversity is proven; visual diversity has not been attempted.** The design step
(`webdesign-agent`'s `analyze_design`) is a separate LLM call from the content writer,
currently running Claude Sonnet on every site — the same model, every time, which is
almost certainly why lendzy reads as "typical AI design." No prior work in this or any
other lane has touched visual diversity across the estate; the only related mechanism
(`design_intent.palette.reference_values`, meant to pin a site's colours) is advisory —
the prompt itself tells the model these are "starting points... you have creative
freedom" — so even a same-model rebuild can drift, and a different model doesn't
automatically fix that on its own.

**The two adjacent workstreams that make positioning stick have both moved.** The
`vigilant_designer_offer_analysis` lane shipped B1 (the site-review agent now judges a
site against its own actual revenue model, not blind) and B2 (a strategy refresh on an
already-deployed site no longer silently re-triggers a full rebuild) — both live, no
image roll needed. It went further than planned: B3, the two premise-completeness checks
originally scoped as "later," is already built and has swept the whole estate, surfacing
four real findings (three sites need a strategy row, one 30-page site has no contact
form) — currently blocked only on a council-review technicality, not on missing work.

## Where we are now

Two fresh, fleet-wide bugs were filed **today** (251, 252) in the one enforcement gap that
had actually shipped — the canonical-link and meta-tag work from early August. 251: every
assembled homepage's canonical points at `/index.html` instead of its real address. 252:
page assembly silently drops each page's own `og:` tags and hardcodes the page language to
English, and this used to be a two-page rounding error and is now a fleet-wide default
affecting 503 already-assembled pages. Both are live, both are diagnosed to a specific line
of Go, neither is fixed yet. Of the original six enforcement gaps: positioning-as-a-setting
is still one-domain-hardcoded, not generalised; structural validity has no standing gate
and just got fresh evidence of why it needs one; tool correctness and the fidelity dial are
untouched; fact discipline is armed for exactly one site (bug 161, still open); truncation
guard status is unverified either way. None of the six is a blocker to *starting* a build —
all of them are reasons a build run today would need a human to check its own homework.

## Where we're going

Before any fleet-wide build order gets decided, three things are worth doing on a small
number of sites first, not the whole portfolio: (1) fix 251/252, since they'd otherwise
ship broken on every new site from day one; (2) run a real visual-diversity experiment —
the cleanest seam is a second `webdesign-agent`-type row pointed at Gemini, routed to a
chosen handful of sites via their work items, rather than a fleet-wide model flip (which
would just replace one uniform look with another); (3) get the owner's call on B3's four
findings and its stuck council round, and a decision on whether B4 (the analyser agent
itself) or the design-critique/page-recompose half of Programme A goes next — both are
genuinely unstarted, not blocked. Build order across the 43 propositions remains the
owner's commercial call and is still not made.

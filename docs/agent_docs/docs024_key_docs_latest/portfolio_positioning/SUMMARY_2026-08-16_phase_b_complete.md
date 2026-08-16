# Summary — Phase B is complete: the machinery for finance directories is finished, and what remains is a site

2026-08-16. The previous summary (`SUMMARY_2026-08-15b_first_supervised_runs.md`) ended with the research pipeline proven — real firms, cited, verified. This one exists because everything downstream of that research is now built too, which is a different claim and the last one Phase B had to make.

## What we're trying to do

Build a fleet of finance and insurance sites where each one carries a directory of the relevant UK providers — mortgage lenders, savings providers, health insurers — with every fact non-price, cited to a named source, and machine-verified against that source before publication. One directory per provider class, identical on every site that opts in. The point is authority a competitor cannot cheaply copy: not opinions about lenders, but a checked, re-checkable record.

## Where we've come from

Yesterday morning the pipeline existed but had never run. Yesterday afternoon we ran it under supervision, five times, and fixed what the watching revealed until a run registered real firms with all claims verified. That proved the *research* half: we can produce trustworthy directory data.

What it did not prove is that a site would ever *show* it. Between verified data and a published directory page sit five separate pieces of machinery, and a kind that is missing from any one of them fails in a way that looks like success everywhere else. Yesterday evening and today closed the last three.

## What we've done

**The recommender is wired in, and reviewed.** The small deterministic component that looks at what a site is about and marks it as wanting a particular directory now runs in two places: every time the routine improvement loop visits an existing site, and at the moment a brand-new site is first classified. That second one matters — it means a new finance site knows it wants a directory *before* its pages are planned, rather than discovering it afterwards. It writes nothing at all for sites outside the three finance verticals, so the rest of the fleet is untouched by construction. Its advisory review came back approved on the second round, and both objections were worth having: one found a real gap in the undo script, which protected the step it reversed but not the neighbouring steps it would have written over.

**The planner now understands what that marking means.** Until yesterday evening it did not. A site would be marked as wanting a lender directory, and the planner would plan the site without one — because nothing had ever taught the planner to read the mark. The site would still get its directory eventually, but only after a later inspection pass noticed the absence, by which point it had been built and published once without it. That gap existed for all six directory kinds, including the three that have been live since July. It is closed: the planner now puts the directory panel on the home page and plans a dedicated directory page, built exactly the way our three existing directory pages are built. We took the wording from the code that defines those pages and from the live pages themselves, rather than describing them from memory.

**Eleven automatic checks are switched on.** Six watch for a site that wants a directory but has not got one; five are site-health checks — broken internal links, bad canonical tags, invalid structured data, missing page-head essentials, dead sitemap entries — that were built earlier in this project and had been sitting inert. One correction to what we said before: three of those six directory pairs had already been switched on back in July for the AI-model directories. The six actually missing were the finance ones. Same number, different six.

**Two things worth carrying beyond this project.** The first is a trap we avoided rather than hit: switching on a check is a configuration change, so it applies cleanly and looks complete even if the running software does not contain that check — in which case it never runs, for ever, with no error anywhere. Worse, when a check is designed to stay quiet until a site opts in, "silent because it never ran" and "silent because there is nothing to report" are indistinguishable. We checked the running software directly before switching anything on, and wrote the trap down where the next person meets it. The second is smaller: writing down *why* we keep hand-wiring these components one at a time surfaced an inconsistency nobody had decided — the equivalent news component runs for existing sites but not for new ones, while ours now runs for both. Nobody chose that; it is just what hand-wiring produces. It is now a numbered item in the architecture track, with the rule that a third one gets built properly instead.

## Where we are now

Phase B is complete. All seven places a directory kind has to be registered are filled, for all six kinds. The three finance registers hold verified, cited entries. Every piece between "we have the data" and "a site publishes it" is built, applied, and confirmed live.

What is not yet true is that any of it has been *exercised end to end*. No site has yet opted into a finance directory, so the planner rule is live and has never fired, and the six directory checks are armed and correctly silent. That is the expected state, not a problem — but it does mean the honest description of Phase B is "finished and unproven in the round", and we should not describe it otherwise until a site has been through it.

One thing we do not control is also worth stating: the routine improvement sweep is switched off fleet-wide, and has been since 14 August for reasons belonging to another workstream. That leaves the existing-sites half of the recommender ready but idle. The new-sites half does not depend on it.

## Where we're going

Straight to the pilot: build remortgagecalculator.uk end to end. It is the first time this machinery is driven by a real site rather than by us checking each piece in isolation, and it tests the three new pieces in a specific order — the site should be marked as wanting a lender directory when it is classified, that mark should produce a lender-directory page in its plan, and if it does not, the checks should raise it. The third is the safety net, which means a clean run has to be verified directly rather than inferred from nothing going wrong.

After the pilot: a cost baseline from what the build actually consumed, owner sign-off, then the outstanding Phase D decisions before any fleet-wide wave.

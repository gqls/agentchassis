# SUMMARY — loanandmortgagecalculator.co.uk is live, adopted, and its divergence is configuration

**2026-07-31.** First summary in this workstream. Written because the milestone happened:
the site went from files-in-a-bucket to live, verified, framework-managed, with its
positioning recorded as config. Current state only — the chronology is in
`README_where_we_are.md`, the evidence in `NOTES`.

## What we are trying to do

Own the ground between two kinds of borrowing. `mortgagecalculator.co.uk` answers mortgage
questions and `loancalculator.co.uk` answers loan questions, and neither can answer the
questions real households actually have: how does my car finance change what a mortgage
lender will offer, should I consolidate this debt into a remortgage, does the next £1,000
go on the deposit or on clearing the loan.

So `loanandmortgagecalculator.co.uk` is the whole-borrowing-picture site: both toolsets on
one domain, and guides that exist only at the crossing points. The requirement that shaped
everything is that the three sites must **evolve apart** — the divergence has to be
something the framework enforces, not something a person remembers.

## Where we have come from

The site was built by porting 23 working calculators — 12 mortgage, 11 loan — with their
arithmetic copied byte-for-byte and the build refusing to run if any of it changed, plus 13
wholly new guides. It reached the storage bucket on 30 July but had no serving path: the
domain was still parked at its registrar, and there are no Cloudflare credentials on this
machine, so the last step was the owner's.

The work up to that point was mostly about **not trusting instruments**. The tool that
tests calculators in a browser reported all 14 mortgage pages broken; all 14 worked, and
the tell was the number — 14 identical failures describes a broken instrument, not a
website. Three faults in that shared harness were found and fixed, each with the fix proven
not to move any verdict it was already giving.

## What we have done

The owner added the Cloudflare zone and the Workers Route, and the site came up.

**It was verified against the live origin, and that immediately exposed three defects that
every pre-launch check had passed.** All three were ours, and all three had the same shape:
**the checks asserted presence where they needed to assert validity.**

1. **The three section-hub URLs were live 404s** — linked from all 42 pages, listed in the
   sitemap, and named by three canonicals. An object store has no directory index and the
   worker rewrites only `/`, so `/loans/` asks for an object called `loans/`. Every site in
   the fleet has this property; ours was the only one that linked that way.
2. **All 13 guides' structured data was invalid JSON** — `html.escape()` escaped the quotes
   JSON needs. Google discards invalid structured data silently.
3. **The copy claimed 24 calculators and 12 loan calculators; there are 23 and 11** —
   dropping `credit-roadmap` never reached the prose.

**The most useful finding is how defect 1 stayed hidden.** Verification ran against
`python3 -m http.server`, which resolves directory indexes that production cannot — a more
forgiving server than production. Then, when our own link checker flagged `/loans/`, the
checker was "fixed" to resolve it, **turning a true positive into permanent silence.** We
were right about 57 of 60 hits, and being mostly right is what stopped us looking.

All three are fixed structurally: one helper defines the hub URL shape for all 13 emission
sites, counts are derived from the tool tables, and `write()` — the single function both
builders funnel through — now refuses any reference naming a directory and any `ld+json`
that will not parse. **All four build assertions were mutation-tested red.**

**Then the site was adopted into the framework**, and the byte gate earned its place. A
locked adoption's only possible byte source is a post-JavaScript crawl, and **41 of 41
components came back changed.** Two mattered: the skip link on every page became an
absolute URL that reloads the page instead of jumping down it — an accessibility regression
— and the amortisation calculator came back **+11,432 bytes** because the crawler captured
the 24-row table the page builds on load. The 41 rerender jobs were held automatically
within one second of being created, all 41 components replaced with the repo bytes, and one
released as a test: it republished **byte-identically**, which is the rebuild-safety
property working.

**Finally, the divergence was recorded as configuration** in the three fields the content
pipeline actually reads: `identity.target_audience`, `identity.key_differentiators`, and a
positioning block inside `content_direction` — including an explicit rule that a subject
answerable without reference to the other kind of borrowing belongs on a single-subject
site, and that the crossing-point framing must always win.

## Where we are now

Live and verified: **23/23 calculators respond in a real browser against the live domain**,
0 dead references of 48, 0 non-200 of 41 sitemap URLs, 42/42 canonicals resolve and
self-name, 14/14 structured-data blocks parse, 51 of 52 files byte-identical (the 52nd is
`robots.txt`, inflated by Cloudflare's own Managed robots.txt on every zone in the fleet).

Framework-managed: 41 pages, all `rebuild_policy='owned'` and verbatim, **zero** LLM work
items — nothing was handed to a model to rewrite. Components 41/41 byte-exact against the
repo, with a rerender proven content-neutral.

**Three limitations, stated rather than buried:**

- **`audience` is a dead end.** The aspect literally named for this purpose is populated on
  29 of 33 sites and read by nothing. Our own earlier plan named it as one of three targets
  for the divergence; that third would have looked done and done nothing.
- **Nothing detects the two sites converging again.** There is no cross-site
  duplicate-content or topical-overlap machinery in the platform at all — what exists
  compares a site against itself. The spec steers new writing; no guard will raise a hand.
- **Two writers now own the same 41 files** — the builders via the repo, and
  `page_components` via rerender. They agree because we just made them agree; a builder
  change without re-running the gate will be silently reverted by the next rerender.

**[UNVERIFIED]** that each ported calculator returns the same answer for the same input as
its original. Byte-identical logic with dependencies present is strong, but it is not
output equivalence.

## Where we are going

**The guides need handing over.** They are registered but frozen, so the framework cannot
yet improve them — the owner chose framework authority for the guides, and delivering that
means decomposing each guide from one whole-document component into sections before
flipping it to `generic`. A straight flip is a known one-way door: the page assembler
concatenates chrome and sections, so a complete `<!DOCTYPE html>` blob yields nested
`<html>`. This is a proper phase, not a flag.

**Then prove the divergence spec reaches a prompt** rather than assuming it, using the
`bug025` acceptance-test pattern — plant a greppable marker, drive one guide through
generation, and pass only when the marker appears in a saved component. That depends on the
guides being generatable, so it follows the phase above.

**Two owner actions in Cloudflare**, both fleet-wide rather than caused by this site:
enable "Always Use HTTPS" (the site currently answers on plain `http://` too, so the same
content sits at two addresses — precisely the duplicate-content problem this project
exists to avoid), and decide whether `www` should resolve, which it does not on any site.

**And an open question worth a decision:** teaching the worker to rewrite `/dir/` →
`/dir/index.html` would restore clean URLs for all 14 sites at once and close
`bugs_open/116`'s root cause. It needs dashboard access and, as a change to a shared
mechanism, the council gate and a register entry. Nothing here depends on it.

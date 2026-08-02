# SUMMARY — the first two seams are shipped, and the carrier is proven on the artefact

**2026-08-02, late night.** Fourth summary today, marking a different thing again:
the seam backlog stopped being a list tonight. Its two top items are through
review and shipped, and the biggest quality gap the lendzy experiment measured
is closed at the mechanism level.

## What we are trying to do

Unchanged: ~150 finance and insurance domains as substantial, deliberately
different sites — built by the pipeline, differentiated by configuration the
register controls. The standing directive is "fix the pipeline": every gap the
lendzy experiment measured becomes a platform seam, reviewed and shipped one
coherent task at a time.

## Where we have come from

The evening summary closed the experiment: the pipeline built lendzy end to end,
and the distance to the hand-built benchmark was seven evidenced seams. The
handoff ordered them; a fresh session picked the list up tonight.

## What we have done since

**Seam 1 — every-page invariants — is live and verified on the built artefact.**
A mission line that must appear on every page ("independent of, and not
affiliated with, the FCA") was structurally unproducible by per-page writers —
measured 3 of 15 pages. It now rides the footer chrome: the shared footer
template has a gated compliance-lines block reading per-site configuration
(`config.chrome.compliance_lines`, the second consumer of the registered
STY-050 mechanism — found as prior art by reading the register before
inventing an entry, which is what the register is for). Fourteen live sites
share that template, so the safety case was proven, not asserted: a Go test
pins byte-identical rendering for any site that sets nothing, and the
old-template constant in the test is md5-identical to the live row it
replaced. The council approved it first round. Lendzy's pages re-rendered
through the normal queue and the census on the sites repo shows the lines
landing on every page.

**Seam 2 — canonicals and honest meta descriptions — is approved, committed,
and waiting on the next chassis deploy.** Nothing on any platform path emitted
a canonical link, and pages without a description shipped an empty description
tag. Both are fixed at the single live assembly path, mirroring the JSON-LD
injector's reviewed discipline: emit the page's one identity or emit nothing,
never guess. Also approved first round. The code is inert until an image
rolls; the pod-grep and census obligations for that moment are written down.

**A council objection made the mechanism better the same night.** The
historian seat pointed out that a badly-typed config value would silently
degrade an entire chrome slot — and that this seam exposes that fallback to
operator-written config for the first time. The gap is now guarded: a
mismatched value is refused loudly and the gated block simply stays absent,
measured safe against all 69 array-declared fields fleet-wide. The trap is in
LANDMINES with its check.

**Seam 3 dissolved under measurement.** The favicon machinery exists and runs
(discovery files the item; asset-deployer derives favicon and OG card from the
logo). Lendzy has no logo — the item is imagery work for the lendzy lane, not
a platform seam.

## Where we are now

The carrier for "on every page" exists, is registered (STY-051), reviewed, and
demonstrated end to end on lendzy: register-controlled configuration → shared
chrome → every built page. The seam backlog stands at: canonicals awaiting the
next roll (verification steps recorded), then the tool handler's missing
rerender, links to unbuilt pages, and the planner's imposed shape — plus the
in-browser tool fixtures, which need serving.

## Where we are going

Next session in this lane: verify seam 2 at the pod after the next chassis
roll, then the tool-handler rerender seam (backlog #4). On the owner's queue,
unchanged: build order across the 43 propositions, the two residual insurance
twins, the fleet-wide www/HTTPS decision — which now moves the canonical and
the JSON-LD identity together, by construction — and the FCA citation pass
both loancash and lendzy still owe. vetcomparison is the named candidate to be
the compliance-lines carrier's third consumer: its content direction already
asks for exactly this in prose.

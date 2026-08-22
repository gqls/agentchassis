# SUMMARY — editorial design uplift, 2026-08-22: the composition plan is written

First summary of this lane (opened 2026-08-20). Written because the read-out
genuinely changed: the one blocked item is unblocked, and the lane's measurement
phase has been acted on, not just completed.

**What we're trying to do.** Make the editorial feature pages — the cited,
chart-backed news analysis pages — look a great deal better: imagery placed
against the parts of the text it belongs to, richer typography and graphic
treatment, more varied charts, and timelines. And do it without losing what the
family already gets right: every plotted number resolves through a registered,
cited fact, so an unsourced chart is unrepresentable.

**Where we've come from.** The lane opened on the owner's ask two days ago.
Rather than design from taste, we asked the platform's own design auditor to look
at the live site, and hand-ran a render audit when the dispatched one timed out.
Those found real, measurable defects — mostly contrast failures in the chart
furniture. Separately, the owner had twice asked that the deeper design — 
components composed of smaller components — be planned by Fable specifically,
and four attempts to dispatch that had died on Fable capacity limits.

**What we've done.** The audit findings were fixed and measured: three selectors
repointed onto the per-palette ink tokens (migration 496), robot-hands 10 → 4
findings, dartsonline 8 → 1, with the de-branding disconfirmation test run and
passed — the brand survived the fix. And the blocked plan is now written:
`features_open/035_FEATURE_component_hierarchy.md`, authored by Fable in this
session (the owner's own session runs Fable 5 — no substitution, fifth attempt).
It absorbed a new owner steer given with the go-ahead: interleaved content and
imagery must not be one LLM call — decompose it, for control, consistency,
versions, and design variations of the same content.

**Where we are now.** The design says: every piece of an interleaved page
becomes its own addressable component row — its own generation call against one
shared brief, its own lock, its own history, its own pinnable design version.
The schema has carried the parent-child column for this since the original
architecture, unused by all 1,903 instance rows; the plan adopts it rather than
inventing a rival. A half-built version seam was found in the same pass: 363
template snapshots exist and nothing reads them back — the plan makes the render
walk read them, which is most of the version-control ask for free. Nothing is
built; the plan is staged with a falsifier per phase, starting on our own locked
insights pages where the blast radius is one page.

**Where we're going.** Next: the local render-walk proof (no cluster writes),
then the first council-gated code phase — the read path plus one recomposed live
page, with the rewrite-survival test as its acceptance. After that: decomposed
generation for one editorial family, then versions and variants, then
agent-proposed re-arrangements through human review. The guide articles across
the other seventeen sites come last, jointly with the inline-imagery lane whose
plan remains the guides' near-term mechanism. In parallel, the flat-mechanism
phases (typography, hero imagery, chart variety, timeline substrate) continue —
none of them wait on composition.

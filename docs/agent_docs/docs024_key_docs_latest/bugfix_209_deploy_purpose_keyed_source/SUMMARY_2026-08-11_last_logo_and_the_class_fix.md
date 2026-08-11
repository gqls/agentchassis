# SUMMARY — 2026-08-11 — the last logo, the class fix that replaced a refuted plan, and the review that sharpened it

**What we're trying to do.** Every site the platform builds should carry a
proper logo — a transparent PNG at logo dimensions — produced and deployed by
the framework, with the mechanism that once mislabelled them fixed at the
class level, not site by site.

**Where we've come from.** Eleven live sites were serving their logo as a
hero-processed JPEG. The store-side cause (a static "hero" on a branch
handling both logos and heroes) was fixed at source on 08-09 (migration 360),
and ten of the eleven sites were repaired and re-rendered over 08-09/10. The
eleventh, relojistas.com, kept failing for a subtler reason: the dispatch
path delivers a work item's fields nested one level down, the deploy step's
purpose binding looked one level up, and the spec's built-in default "hero"
silently won — twice, against a correct item spec AND a correct asset row.

**What we've done.** The plan we inherited proposed bridging the deploy step
via a deprecated config alias; the bug file's own mechanism notes — and a
committed test — prove that bridge structurally dead for any defaulted field,
so it was refuted before a line was written (logged in WRONG_CALLS). The fix
that shipped is migration 380: the dispatch loop now also maps the item's
purpose to the top level, the exact idiom its sister dispatcher already used.
It was applied with an induced verify, proven the same hour at the artefact
(the re-run deploy committed "Deploy logo image" where two pre-fix runs said
"Deploy hero image"; the site serves a 400×170 RGBA PNG), and put through the
council: round 1 said REVISE with a high-severity question about a
multi-active-row hazard; every objection's check was run — one active row, a
snapshot taken, a both-arms characterisation test written — and round 2
APPROVED it. The estate-wide audit then found only three sites still
referencing any logo.jpg; all are flipped and verified at their served pages,
except fundamentallyai's index, which another lane is actively rebuilding and
which carries our patches in both its data and its html. Separately: the
kafka scheduler's memory fix (C3) is confirmed live, and the topic-sweep
cron's first-ever firing exposed and fixed a missing KUBECONFIG in the cron
environment.

**Where we are now.** 11 of 11 logos done. The class fix is live, tested,
council-approved, and every future undeployed-asset deploy resolves its real
purpose. Zero renderable logo.jpg references remain except the one page
another lane owns mid-rebuild. The stale .jpg files still exist and still
serve, deliberately.

**Where we're going.** Deleting the stale logo.jpg files needs the owner's
word plus fundamentallyai's index verified after its rebuild. Bug 231 stays
open for the fleet census (now three arms) and its two structural remedies.
Bug 240 keeps its scheduler-scoped transport work and the C1 question. 209
Phase 3 and 236 remain unowned by this thread.

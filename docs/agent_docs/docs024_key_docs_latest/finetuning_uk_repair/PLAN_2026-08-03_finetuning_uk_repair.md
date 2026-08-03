# PLAN — finetuning.uk repair, and what its failure says about the framework

**Started:** 2026-08-03 · **Trigger:** owner — "the finetuning site is looking
terrible. Fix it using the framework (not locally). Run the audit checks, see
what you can see wrong including visually. Make sure the handlers are all
automatically picking up the items properly and fixing them. Check the framework
catches everything."

## The brief, read carefully

Four instructions, and the third and fourth are the substantial ones. "Fix the
site" is a day's work on one site. "Make sure the handlers are picking items up"
and "check the framework catches everything" are questions about the machine, and
the site is the test case that exposes it. This plan is written around that
reading: **the site is the symptom, the framework is the subject.**

"Not locally — I want the framework to do it" is a constraint on HOW, and it is
load-bearing. Hand-editing `page_components.rendered_html` would have made the
site look right in about ten minutes and taught us nothing, and the next
re-render would have reverted it. Every repair here is made at the thing that
GENERATES the artefact.

## What was actually wrong (established 2026-08-03, evidence in NOTES)

Three findings, in descending order of how much they matter.

**1. A component renders an icon name into an image slot.** `departments-grid`
emitted `<img src="cpu" class="member-photo">` for every department. `/cpu`,
`/network`, `/database` all return 404 on the live host, so the browser painted a
broken-image icon 120px across — eight down the homepage, eleven more on
/about.html. This is the "looking terrible". The component was forked from a
team/staff-photo component (the markup still says `team-member`, `member-photo`,
`member-bio`) and repurposed for departments, whose schema declares
`icon: string`. The `<img src>` came along for the ride.

**2. The framework could not see any of it.** `check_image_url_404` is the check
that owns broken images, and it has two predicates: one needs
`/assets/images/<name>.<ext>`, the other needs an empty or `#` src. A bare word
matches neither. In the same discovery run of 2026-07-26 it raised five findings
for `case-study-*.jpg` and stayed silent on nineteen broken images on the same
two pages.

**3. Nothing was dispatching, and that is fleet-wide, not site-specific.** 61
open items on this site, every one at `detected` or `unresolved`, with
`attempt_count = 0` — never even attempted. The dispatcher claims on
`status IN ('triaged','approved')`; the only promoter of `detected → triaged` in
the platform is the `triage_findings` step INSIDE the improvement-loop; the
improvement-loop's only scheduled route, `improvement-sweep`, has been
`enabled=f` since 2026-05-02. Fleet-wide that day: **204 detected across 10
sites, 2 triaged.** Detection works. Dispatch does not.

## Phasing, and why in this order

| # | Step | Why here |
|---|---|---|
| 1 | Fix `departments-grid` at the template | Root cause. DB config, live immediately. Fixes all 31 occurrences on both sites at once. |
| 2 | Fix `check_image_url_404` to see the shape | So the framework catches the CLASS next time. Go — inert until a build. |
| 3 | Fire the improvement-loop at the site | The framework does the repair: discovery → triage → dispatch → handlers. |
| 4 | Report the dispatch finding | It is bigger than this site and is not mine to decide. |

Step 1 before step 3 is deliberate: the loop's rerender is what puts the fixed
template onto the live pages, so the template must already be right when it runs.

Step 2 cannot help step 3 — the running binary predates it (pod-grepped and
confirmed, see NOTES). That is stated rather than hidden: **the checker fix is
committed and inert.** It changes what the framework catches on the next build,
not on this run.

## Decisions, with reasons

**The template fix targets `<i data-lucide>`, and that is not a guess.** The
`features` component already renders `{{if .icon}}<div class="feature-icon"><i
data-lucide="{{.icon}}"></i></div>{{end}}`, on the same page as the broken one,
working. All four affected pages load `lucide.min.js` and call
`lucide.createIcons()` — verified live before applying, because a fix that
renders into a page with no icon library is a different kind of broken.

**The new check is a separate KIND, not an extension of the empty-src tally.**
"This img has no source" is repaired by supplying one. "This img has a source
that is not a source" is repaired by changing the template that wrote it.
Different repairs, so a shared work item would send a reader to the wrong place.

**Severity high, where the other path findings are medium.** A template defect
repeats on every page that mounts the component and every site that mounts it.
The 31/2/1 census is the argument: one bad template, two sites, thirty-one broken
images.

**I did NOT re-enable `improvement-sweep`.** It has been off since 2026-05-02 and
`IMP-016` records the pause as deliberate. Turning a fleet-wide sweep back on to
fix one site would be a fleet decision taken as a side effect of a site task.
Instead the loop was fired at ONE site, which is the same machinery with the
blast radius the task actually has. Whether the sweep should come back on is in
"the open question" below, for the owner.

## The open question, for the owner

**204 findings across 10 sites are parked where nothing can reach them, and the
one component that promotes them is switched off.** Three ways forward, and the
choice is a judgement about risk appetite, not a technical one:

1. **Re-enable `improvement-sweep`.** Cheapest. Also the least controlled: it
   would begin promoting and dispatching against ten sites at once, and the
   reason it was paused (IMP-016: "deliberately off during core build") may or
   may not still hold.
2. **Fire the loop per site, deliberately**, with `294_TRIGGER_…` — what this
   session did. Controlled, and it does not scale to ten sites by hand.
3. **Promote by hand and let `build-pipeline-trigger` do the rest** —
   `UPDATE site_work_items SET status='triaged' WHERE status='detected'`, which
   004:305 explicitly offers. Fastest to a dispatching queue, and it skips triage,
   which exists to decide what SHOULD be worked; 235 items are already
   `unresolved` after failed attempts, and this would not distinguish them.

My recommendation is (2) for now and a decision on (1) separately, because the
finding that matters — that detection and dispatch have been disconnected for
three months — deserves its own answer rather than being closed by whichever
option unblocks this one site.

## Corrections to this plan

*(none yet — corrections land here, marked and dated, never silently edited away)*

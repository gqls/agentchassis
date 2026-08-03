# SUMMARY — 2026-08-03 — the site was the symptom; the framework was the subject

## What we're trying to do

The owner said finetuning.uk was looking terrible and asked for it to be fixed
**by the framework rather than by hand** — then asked three harder things: run the
audit checks and see what is wrong including visually, make sure the handlers are
picking work up and fixing it automatically, and check that the framework catches
everything.

The first is a site task. The other three are questions about the machine, with
the site as the test case. That is how this was taken.

## Where we've come from

The site has been live and deployed for months and has a long history of work
items filed against it — 61 open when this began, some dating to April. It was
last touched by a discovery run on 26 July, and nothing had moved since 27 July.
No one had established why.

The wider context matters: `improvement-sweep`, the scheduled task that drives
the whole detect-triage-repair loop, has been switched off since 2 May. Several
sessions have tripped over consequences of that without the scale of it being
written down.

## What we've done

**Found the visual defect, and it was not subtle once located.** Nineteen broken
images — eight on the homepage, eleven on the About page — each a 120px
broken-image icon in a grey circle. The cause: the `departments-grid` component
was forked from a staff-photo component and repurposed for departments. Its
schema correctly declares `icon` as a name; its template still rendered that name
into `<img src="…">`. So the page asked the server for a picture called "cpu",
and got a 404.

**Fixed it at the source, not at the artefact.** The template now emits
`<i data-lucide="{{.icon}}">` inside a badge of the same geometry — which is not
an invention, it is what the `features` component on the same page already does,
successfully. Verified beforehand that all four affected pages load the icon
library. The census said 31 occurrences across 2 sites from 1 component, so one
template fix closes the whole class.

**Found why the framework never reported any of it.** `check_image_url_404` owns
broken images and has exactly two predicates: a path under `/assets/images/` that
resolves to nothing, and an empty or `#` source. A bare word is neither. In the
same run, on the same two pages, it filed five findings for missing case-study
photos and stood next to nineteen worse ones without seeing them. A third
predicate now closes that, with six negative controls guarding against the
classic failure of widening a regex until it reports nothing.

**Found why nothing was being repaired, and it is fleet-wide.** Every open item
had a handler named, and `attempt_count = 0` — nothing had ever picked one up.
The dispatcher claims only `triaged` or `approved`; every item was `detected`;
and the only thing in the platform that promotes `detected → triaged` is a step
inside the improvement-loop, whose only schedule has been off since 2 May.
Fleet-wide: **204 detected across 10 sites, 2 triaged.** Detection has been
running into a wall for three months.

**Ran the framework, and it worked.** Fired the improvement-loop at this one
site. It promoted **128 items to triaged** — the first dispatchable work on this
site in months — and the handlers began claiming immediately.

**Caught the trap that would have made all of that cosmetic.** The 68 queued
rerenders would have drained to zero and repaired nothing: the rerender agent
routes on `spec.reason`, and without one it re-staples the page from HTML
*already stored* — the very HTML being replaced. 42 of the queued items were
reason-less, including the only one for the About page. An explicit
`reason='section_data_resolved'` item completed in 46 seconds, and the live page
now serves ten proper icons and zero broken images.

## Where we are now

**`/about.html` is fixed and proven live** — checked by fetching the served page,
not by trusting a work item's status. Eleven broken images gone.

**`/index.html` is queued and will land without intervention** — it carries a
rerender that does take the regenerating branch, sitting near the front of the
queue.

**The queue is draining.** 128 promoted, and completions are arriving steadily.

**The checker fix is committed and deliberately NOT live.** Pod-grepped both
replicas with a positive control: the new symbol is absent, the control present.
It ships on the next chassis build. Said plainly rather than left to be assumed.

**The council said REVISE and it was right.** Six of thirteen seats flagged that
a standing landmine pairs this file with a sibling check and my submission never
showed I had read it — I had not. Having now run the check that landmine
prescribes, the answer is structural: the sibling matches two literal paths, both
containing characters the new pattern excludes, so no input can fire both. That
analysis went into the file, not just the submission, and round 2 is in flight.

## Where we're going

**One decision belongs to the owner**, and it is deliberately not being taken as a
side effect of a one-site task: what to do about the 204 findings parked
fleet-wide. Re-enabling the sweep fixes everything at once but was switched off
deliberately and would start changing ten sites immediately. Running site by site
is safe and does not scale. Promoting everything by hand is fastest and skips the
step that decides what is worth doing, on a queue where 235 items have already
failed at least once. The recommendation is to keep running it deliberately and
give the three-month disconnection its own answer.

**Two things are recorded and not acted on.** On ai-agent-orchestration.com,
eight department icons are named "strategy", "research", "operations" and the
like — not icon names at all — so they will render as empty badges rather than
broken images. That is a content defect on that site, not a template one. And a
council seat made the broader point that this closes one shape of a general
problem: schema-declared strings rendered into structural HTML sinks with no
validation that the string resolves. Both are follow-ups, both written down.

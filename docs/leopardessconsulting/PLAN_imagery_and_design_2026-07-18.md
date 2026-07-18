# PLAN — imagery, graphics and design (owner review 2026-07-18)

Owner's review raised seven things. This is what each one is, whether it is a leopardess fix
or a platform bug, and what happens next. Two are fundamental and now have handoffs in
`bugs_open/`.

## The seven items

| # | Owner's report | Verdict | Where it goes |
|---|---|---|---|
| 1 | Tools not linked from the nav | **True — fixed today.** The rebuild had stripped tool links; a `tools` nav group existed but renders in neither header nor footer. 4 working tools now linked from the footer (utility group). | done |
| 2 | Blank page behind "Monitoring Coverage Gap Finder" on services | **True — 404.** Not a missing page: a *phantom link the re-plan invented* when it clobbered services at 07:50 today. | `bugs_open/001` (evidence appended) |
| 3 | Hero image has unreadable text | **True, and it is a model-class limitation, not a prompt miss.** | `bugs_open/011` (new) |
| 4 | "trust", "honest", "earns its keep" overused | **True — 12 / 9 / 2 live occurrences.** Now banned in the voice gate + prose guidance; owner-cited homepage instances rewritten. | done + rolling |
| 5 | Want infographics showing system strengths | **Needs a build that does not exist.** Must be code-rendered SVG, never generated. | `bugs_open/011` §V3 |
| 6 | Want more imagery / graphics / better design | Partly available now (per-page heroes), partly needs #5. | this plan |
| 7 | Want more hero images — only index had one, and it was bad | **True.** index replaced today; who-we-help + how-we-work already good; the rest have none. | this plan |

## The two fundamental problems

**A. Generated images cannot render text — `bugs_open/011`.**
The old homepage hero was an SDXL image that *looked* like a flowchart and was full of
gibberish words. That is what diffusion models do: they synthesise glyph-shaped texture, not
text. Answering the owner's question directly — a better model helps a little, better prompts
help (they are why the two good heroes are clean), and a "loop until correct" can only ever
*reject* a bad image, never make a model render text it structurally cannot render.

The fix is to split the two jobs:
- **atmosphere / heroes →** generated, Banana `illustration`, hard no-text constraint;
- **anything with words, labels, numbers or structure →** rendered in code as SVG from real
  values. This is already the site's own stated principle (decisions D1/D3: code renders data,
  the LLM never touches the values) and the planned L7 chart component. It was simply never
  built, while `infographic` remained a diffusion target — which is the trap.

**B. The re-plan keeps clobbering this site — `bugs_open/001`.**
Twice in 24 hours a rebuild replaced hand-verified pages: the homepage (re-adding fabricated
stats and invented case-study titles) and services (inventing the dead tool link the owner
clicked). Content fixes on this site currently have an undefined shelf life. Until 001 is
fixed, **every content improvement below is provisional** — worth doing, but expect to redo it.

One useful discovery: per-page heroes wired through `site_plans`/`site_plan_imagery` **survive
the clobber**, while `page_components` copy does not. So imagery work is durable in a way copy
work currently is not — a good reason to do the imagery now.

## What is being done, in order

**Now (done today)**
- 4 working tools linked from the footer; the dead `tools`-group entry removed.
- Homepage hero replaced: text-free Banana illustration (four inputs converging into one
  steady output), wired via `site_plan_imagery` so a rebuild cannot drop it.
- "trust" / "honest" / "earns its keep" added to the voice gate and to `banned_language`;
  the owner's cited homepage sentences rewritten ("…can't just be trusted" → "…will not match
  the register on its own"; "how much the source can be trusted" → "how reliable that source
  has been"; both "the honest answer" instances cut).

**Next — heroes on the remaining pages (durable, do regardless of 001)**
about, services, use-cases, contact, case-studies, engagement-model, faq,
technical-architecture. Same proven route: `kind:"illustration"` → Banana, explicit no-text
constraints, **look at every image before wiring it**, then a `site_plan_imagery` row + an
`image_landed` re-render. Note `hero-about` and `hero-services` components have no image field
— they need the additive gated-field change first (a shared-component edit, 9 and 5 sites
respectively), which belongs to the imagery workstream.
Also: `/assets/images/hero.jpg` — the garbled image — is still the site-wide fallback and is
live on how-it-works. Replace the file itself so every fallback page improves at once.

**Then — the words pass**
The voice checker now lists every remaining "trust"/"honest" instance per page (engagement-model
FAQ, technical-architecture, the ROI guide, and others). Work that list once the clobber is
fixed, so it is not done twice.

**Blocked on a build — infographics (`bugs_open/011` §V3)**
The owner wants graphics showing the system's strengths. The honest position: we can produce
those properly only as code-rendered SVG driven by evidence-base values, so the numbers are
real and the text is legible, selectable and accessible. First candidates: the verification
pipeline flow (scrape → match → human check → record) and a real figures panel (2,767 verified
/ 937 enriched / 5,652 collected / 4,672 scored / 8 sites). Until that exists, adding more
*generated* imagery in place of infographics would repeat the exact defect the owner reported.

## Standing rule this review establishes

An image on this site is allowed to be a picture, or a real diagram. It is never allowed to be
a picture pretending to be a diagram. Anything carrying words or numbers is code-rendered.

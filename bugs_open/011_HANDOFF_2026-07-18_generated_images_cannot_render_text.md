# HANDOFF — Generated images render unreadable text; the site has no way to make a real infographic

**Filed:** 2026-07-18, from the leopardessconsulting.co.uk rebuild (owner review).
**Severity:** High for any site that wants explanatory graphics. It is currently *impossible*
to put a correct diagram, chart, or infographic on a chassis site: the only image path is a
diffusion model, and diffusion models cannot render legible text.
**Status:** OPEN — architecture decision needed. Nothing here is a config toggle.

---

## 1. The evidence (look at this first)

`https://leopardessconsulting.co.uk/assets/images/hero.jpg` — the live homepage hero until
2026-07-18. It is an SDXL-generated image that *looks* like a professional flowchart
infographic: callout boxes, bullet lists, connector lines, section headings, a caption strip
along the bottom. **Every word in it is gibberish.** Convincing layout, zero legible content.

Owner's report, verbatim: *"The hero image has unreadable text in it, it is a sort of plan
image, if so it should be a fully working plan or infographic or have no text at all."*

That sentence is the whole specification for this bug. An image is allowed to be a picture, or
a real diagram. It is not allowed to be a picture *pretending* to be a diagram.

## 2. Root cause (not a prompt problem)

Diffusion image models (Stability SDXL, and Banana/Gemini-image to a lesser degree) synthesise
glyph-shaped texture, not text. They cannot reliably render words, and they *especially* cannot
render a labelled diagram, because a diagram is mostly text. This is a property of the model
class, not of our prompt.

Two compounding platform facts:

1. **`kind:"hero"` routes to SDXL** (`internal/adapters/imagegenerator/dynamic_adapter.go`,
   the provider switch: `icon|logo|illustration|infographic|sprite_sheet` → Banana, everything
   else → Stability). So a hero prompt that mentions anything structural gets the model least
   able to do it.
2. **`infographic` is a routable image `kind`.** The system therefore *offers* infographics as
   a generated-image type. That is the trap: an "infographic" produced by any image model is
   guaranteed to contain fake text. The kind should not exist as a diffusion target.

Corroborating success: on the same site, `kind:"illustration"` → Banana, with an explicit
"absolutely no text, no letters, no numbers, no labels" constraint, produced two genuinely
good, on-brand, text-free heroes (`hero-who-we-help.jpg`, `hero-how-we-work.jpg`). Text-free
generation works. Text-bearing generation does not.

## 3. Answering the owner's question directly

> *"Do we need to change the image llm, or do we need a pipeline to describe the images better,
> or loop until it is correct?"*

- **Change the model — helps a little, does not solve it.** Banana is better than SDXL and is
  already the right lane for flat illustration. No current image model renders a correct
  labelled diagram to production standard. Do route heroes for illustration-style brands to
  Banana; do not expect it to fix text.
- **Better prompt descriptions — necessary, not sufficient.** Explicit no-text constraints
  measurably improved output (the two good heroes carry them). They reduce the chance of
  glyph-texture appearing; they cannot make glyphs legible.
- **Loop until correct — useful only as a *reject* gate.** A verify loop cannot make a model
  render text it structurally cannot render. It *can* detect text and reject the image. That is
  worth building (see V2 below), but as a guard, not a generator.

**The actual fix is to split the two jobs apart:**

| Job | Correct mechanism |
|---|---|
| Atmosphere, brand texture, page heroes | Generated image (Banana, `illustration`), **hard constraint: contains no text at all** |
| Anything with words, labels, numbers, structure, data — diagrams, infographics, charts | **Rendered in code** (SVG/HTML from real values), never generated |

This is not a new principle for this platform — it is already written down. Leopardess design
decisions D1 and D3 say *code renders data, the LLM never touches the values*, and the planned
L7 chart component (Go emits static SVG, JS progressively enhances) is exactly this shape. The
bug is that the principle exists on paper while `infographic` remains a diffusion kind and no
code-rendering path was ever built.

## 4. Proposed work

**V1 — stop producing fake diagrams (small, do first).**
- Remove `infographic` from the diffusion routing map, or hard-fail it with a message pointing
  at the code-rendered path. It must not be silently satisfiable by an image model.
- Inject the no-text constraint into *every* generated-image prompt at the action layer
  (`generate_image_actions.go` already folds `constraints.no_text` into the negative prompt —
  make it a default for all kinds, not an opt-in the caller may forget).
- Route hero generation by the site's own `design_intent.imagery_direction`: a site whose
  house style is flat illustration should get Banana for heroes, not SDXL. Today the kind
  string alone decides, and `hero` is hardcoded to the photographic provider.

**V2 — a reject gate (the honest form of "loop until correct").**
After generation, before deploy: run OCR (or a vision call) over the candidate image. If any
text-like glyph region is detected above a small threshold, reject and regenerate with a
strengthened negative prompt, up to N attempts, then fail loudly to a human. Never deploy an
image with detected text. This is a natural discovery-check/gate shape and mirrors what the
claims and voice gates do for copy: deterministic detection, human terminal.

**V3 — the code-rendered graphics lane (the thing the owner actually asked for).**
The owner wants *"infographics showing the strengths of the system"* and *"more graphics to
make it look better designed."* Build a component that emits **SVG from real values**:
- Inputs come from the evidence base / live SQL (the claims-verification layer already
  formalises facts with `sql` sources — reuse it, so an infographic cannot state an unverified
  number).
- Go renders static SVG server-side; JS enhances only if present; text is real `<text>`, so it
  is legible, selectable, accessible, and translatable.
- Candidate first pieces for leopardess: the verification pipeline flow (scrape → match →
  human check → record), and a real figures panel (2,767 records verified / 937 enriched /
  5,652 items collected / 4,672 scored / 8 sites) — all evidence-base-backed.
This is the L7 chart component generalised, and it is the only way to satisfy "should be a
fully working plan or infographic."

## 5. Blast radius

Fleet-wide. Every chassis site uses the same image path. Any site that has ever generated an
`infographic`, or a hero whose prompt implied structure, is likely carrying an image full of
fake text right now. Worth a sweep once V2's detector exists — it doubles as an audit tool for
already-deployed images.

## 6. Interim mitigation (already applied on leopardess)

- Per-page heroes regenerated as `kind:"illustration"` via Banana with explicit no-text
  constraints, wired through `site_plans`/`site_plan_imagery` rows so a page rebuild does not
  drop them (`docs/leopardessconsulting/scripts/wire_heroes.sql`).
- **Review every generated image by eye before wiring it.** Two of four generated on this site
  were unusable; the failure is not detectable from the pipeline's own success signals.

## 7. Key files

- `internal/adapters/imagegenerator/dynamic_adapter.go` — the provider switch (kind → Banana/Stability)
- `platform/orchestration/actions/generate_image_actions.go` — prompt assembly, `constraints` → negative prompt, per-kind defaults
- `platform/orchestration/actions/imagery_style_guide.go` — style-guide/reference plumbing
- `docs/leopardessconsulting/HANDOFF.md` §9 — the safe per-image route and its landmines
- `docs/leopardessconsulting/PLAN_leopardess_rebuild.md` §5 (L7) — the code-rendered chart design, decisions D1/D3

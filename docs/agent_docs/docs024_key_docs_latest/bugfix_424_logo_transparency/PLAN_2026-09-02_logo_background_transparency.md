# PLAN — bugs_open/424, logo background transparency

Drafted by a fable pass, revised after review from the `bugfix 420 and 417` session
(peer feedback folded in below, marked). Status: implementing.

## Problem

The logo pipeline asks the image model for a transparent background. Gemini's image
family (the `banana` provider) cannot emit an alpha channel — it always produces flat
RGB pixels — so a prompt asking for transparency gets painted as its own UI symbol: a
grey-and-white checkerboard, served as opaque paint. Verified structurally (colour
type 2, no `tRNS` chunk). No prompt wording fixes this — the property does not exist
in the model's output space. Owner ruling: the background behind a logo is not part of
the logo.

## Constraints that bind

- Go, stdlib where possible. `image`, `image/draw`, `image/png`, `nfnt/resize` already
  in use (`platform/storage/image_processing.go`); no segmentation/ML library exists
  anywhere in the repo (`background_remov|rembg|alpha_channel|make_transparent` → 0
  hits, `platform/`+`internal/`, 2026-09-02). Do not add one.
- No asset version history — `assets` UPSERTs in place, no revert. The fix must fail
  **closed**: refuse to store rather than store something worse than the interim.
- `composedPaletteDirection` deliberately excludes `logo` from the brand palette
  (2026-05-20 lesson) — the key colour must not be brand-derived.
- Council gate applies (`platform/`, `internal/`).
- Interim already live (webdesign lane, boxingonline.uk / served at
  boxingonline.ugg2.com): logo on a solid near-black ground matching the header — not
  transparent, acceptable only against that one header colour. 424 stays the real fix.
- **[peer, bugfix 420/417]** `kindDefaults["logo"]` also carries `"…, complex
  background"` in its negative prompt, which `foldNegativeIntoPrompt` (bugs_open/028)
  folds into the *positive* prompt as a prohibition. A flat key colour is the opposite
  of complex, so this should be harmless — but read the composed `origin_prompt` after
  the first real run rather than trusting the template (bugs_closed/390: co-present
  instructions are adjudicated by the model, not by precedence wording).

## Options considered

- **(a) Provider-side alpha** — REFUTED. `ImageConfig` (banana/api/types.go) has one
  field, `AspectRatio`. Structural model limitation, not a missing knob.
- **(b) Two-backdrop difference matting** — rejected. Needs the foreground pixel-
  identical across two independent generations; the model only reproduces subjects
  approximately.
- **(c) ML background removal** — rejected. New dependency, non-deterministic, and
  segmentation clips thin strokes — exactly what a mark is made of.
- **(d) Keyed ground + border flood-fill matting** — RECOMMENDED. Ask for a flat,
  deterministic, saturated colour the model *can* paint; remove it mathematically in
  pure Go; verify the removal at the artefact, the same way the defect was caught.

## Recommended design

### Where the clause lives — the SAME choke point as bugs_open/417's text policy

**[peer's correction, adopted]** The fable draft originally had the adapter
(`dynamic_adapter.go`) build the ground clause text from a raw hex value carried over
Kafka. Moved to `generate_image_actions.go`'s existing `if kind == "logo" { … }` block
instead, alongside `applyLogoTextPolicy`, for a reason specific to this codebase: the
banana provider's own header comment records that `assets.origin_prompt` is written by
the **action layer**, before the adapter's own negative-prompt fold ever runs — "a
thread checking there would wrongly conclude the avoid list is still inert" is 424's
own file quoting that exact trap. Building the clause in the adapter would make it
invisible to `origin_prompt`, which is precisely the shape of blindness 417 was fixed
to avoid. Building it in the action layer means the clause is greppable in
`origin_prompt` by its sentinel, exactly like `LogoTextFreeSentinel` — the census
mechanism 417 already proved out.

The adapter still needs the colour as **structured data**, not just prose, because it
has to matte the exact hex — so a `KeyGround` field is threaded through Kafka
alongside the prompt text. Both halves (prose clause, structured value) are derived
from one Go constant, so they cannot drift apart.

### The constants (`discovery_checks/default_brand_prompt.go`, beside `LogoTextFreeClause`)

```go
const LogoBackgroundKeyHex = "#FF00FF" // pure magenta — never brand-derived

const LogoBackgroundKeySentinel = "single flat, uniform, edge-to-edge field of pure magenta"

const LogoBackgroundKeyClause = "The entire background is a " + LogoBackgroundKeySentinel +
	" (" + LogoBackgroundKeyHex + "), with no gradient, vignette, shadow, glow, texture, " +
	"panel or border, and the artwork must not touch the image edges. This instruction " +
	"overrides any earlier wording in this prompt about transparency, a plain background, " +
	"or any other ground colour."
```

Magenta: maximally distant from greys/blacks/whites/metallics/most brand hues, and
unlike black or white it never reads as a "natural" ground a model might blend a mark
into. Same override-previous-wording shape as `LogoTextFreeClause` (417's proven
pattern — a folded negative demonstrably loses to a positive licence sitting in the
same prompt; the converse should hold here too, so state it positively).

### `applyLogoBackgroundPolicy` (`generate_image_actions.go`, pure function, mirrors `applyLogoTextPolicy`)

- Idempotent on `LogoBackgroundKeySentinel` (already-washed prompt gets no second copy
  — same shape as 417's idempotence guard).
- Appends `LogoBackgroundKeyClause` to the prompt.
- Adds `magenta, #ff00ff, checkerboard, transparency pattern` to the negative prompt —
  belt, and the checkerboard/transparency-pattern terms name this bug's own failure
  mode directly.
- Called in the existing `if kind == "logo" { … }` block, beside
  `applyLogoTextPolicy`.
- Sets `imageData["key_ground"] = checks.LogoBackgroundKeyHex` alongside the prompt
  fields already forwarded — the adapter's structured signal.

### The matte: `KeyOutBackground` (new file, `internal/adapters/imagegenerator/keyground.go`)

Pure function, no I/O, unit-testable on synthetic images:

```go
type MatteStats struct {
    BorderKeyed float64 // fraction of border-ring pixels made transparent, 0..1
    Keyed       int
}

func KeyOutBackground(img image.Image, key color.NRGBA, inner, outer float64) (*image.NRGBA, MatteStats)
```

1. **Border flood-fill**: seed every edge pixel within `outer` of `key`; BFS 4-connected
   neighbours within `outer`. Only background *connected to the outside* is erased —
   the safety property that stops an interior element merely resembling the key from
   getting punched through.
2. **Enclosed holes** (a ring, a counter): a second, tighter pass — any pixel anywhere
   within `inner` of `key` is also keyed. Its safety rests entirely on the negative
   prompt forbidding near-exact magenta inside the mark; this is the one place the
   design can misfire (see Risks).
3. **Graded edge + despill**: for `inner < d < outer`, alpha ramps linearly; unmix the
   foreground colour against the key (`fg = (p − (1−a)·key) / a`, clamped) so
   anti-aliased edges don't carry a magenta fringe.
4. Encoding a resulting `*image.NRGBA` with any non-opaque pixel via `png.Encode`
   already writes colour type 6 — no format hacking.

Must operate through `image.Image`'s generic `At(x,y).RGBA()` (16-bit-normalised
regardless of source bit depth) rather than assume a concrete 8-bit type — **[peer,
boxingonline session]** the live boxingonline logo asset is confirmed 16-bit depth
(`colour_type=2`, `400×218`, measured 2026-09-02), so an implementation that assumes
8-bit NRGBA would silently mishandle exactly the asset this bug is about.

Thresholds `inner=48, outer=110` **[UNMEASURED]** — the interim ground-colour run
showed the model approximates a requested hex by roughly 17 Euclidean RGB units
(`#0a0a0a` asked, `(10,16,16)`–`(12,20,18)` measured); 48 is ~3× that drift as a safety
margin. Set precisely from the first real magenta-keyed output, dated.

### The guard — fail closed (`dynamic_adapter.go`, `generateImage`)

Immediately after `result, err := p.GenerateImage(...)` succeeds and
`data.KeyGround != ""`, before upload:

1. Decode the returned bytes into `image.Image`.
2. Run `KeyOutBackground`.
3. **If `stats.BorderKeyed < 0.95`, return an error and do not upload.** The model
   ignored the key-colour instruction (a checkerboard, a solid ground, a vignette) —
   this is the fail-closed half that makes the whole design safe against there being
   no revert seam: a refused generation leaves the existing asset in place; a silently
   half-matted logo stored over the only copy would not.
4. Re-encode as PNG; hand bytes + `image/png` to the existing upload path unchanged.

### Does NOT fix — recorded so it isn't mistaken for closing a sibling bug

**[peer, bugfix 420/417]** bugs_open/421 (multi-panel design comp — mark on two
different grounds in one image) is a **different mechanism**. A border flood-fill
keys whichever ground touches the border; a two-panel comp could come back
half-matted, which reads as a subtler bug than it is. The `BorderKeyed` guard above
gives some protection (a comp is likely to fail the ≥95% threshold and get refused
outright), but this fix must not be read as having verified single-composition — 421
stays open and gets checked independently.

## Verification — must be symmetric with how the defect was caught

**[peer, boxingonline session]** A viewer cannot tell painted-checkerboard from real
alpha by looking — it draws a checkerboard *for* real transparency too. Check BOTH
signals, never one alone (a palette-transparent PNG can carry `tRNS` at colour type 3,
which a colour-type-6-only check would misreport as a fail): colour type 6 **or**
`tRNS` chunk present. And key any before/after comparison on prompt content or the
work item, never the asset row's age — `assets` UPSERTs in place and `created_at`
keeps the original generation's timestamp after every regeneration (LANDMINES).

## Gating — why a `KeyGround` field, not a bare kind check

**[peer's caution, addressed]** A bare `kind=="logo"` check at the adapter has a real
history of going blind on legacy callers that map no kind at all (417 was rejected
twice on this exact shape before `stepNameKindHint` closed it). Traced end-to-end
here: `generate_image_actions.go` resolves `kind` (including the step-name fallback)
**before** dispatch and writes it into `imageData["kind"]` only when non-empty — so
the adapter's `data.Kind` already reflects the same fully-resolved value 417's fix
made comprehensive. Gating `applyLogoBackgroundPolicy` on that same resolved
`kind == "logo"` inherits 417's coverage rather than re-deriving it. The `KeyGround`
field carried to the adapter is not a second gate — it's the same decision, expressed
as the value the matting step needs rather than re-inspected as a string kind.

This is not shipped opt-in-default-OFF per-site. The owner's ruling defines what a
correct logo asset *is*, for every site — a per-site opt-in would re-implement the
bug as the default. Every kind other than `logo` sees an empty `KeyGround` and
identical behaviour to today, so no new authority reaches any consumer that did not
already carry this defect. Register the seam in the concept register in the same
commit, naming logo generation as the sole consumer today.

## Edits (feeds the council submission `plan` array)

1. `discovery_checks/default_brand_prompt.go` — add `LogoBackgroundKeyHex`,
   `LogoBackgroundKeySentinel`, `LogoBackgroundKeyClause` beside `LogoTextFreeClause`.
2. `generate_image_actions.go` — add `applyLogoBackgroundPolicy`; call it beside
   `applyLogoTextPolicy` in the `kind == "logo"` block; set
   `imageData["key_ground"]`.
3. `internal/adapters/imagegenerator/dynamic_adapter.go` — add `KeyGround` to
   `ImageRequestData`; in `generateImage`, run the matte + fail-closed guard after the
   provider call, before upload.
4. `internal/adapters/imagegenerator/keyground.go` — NEW: `KeyOutBackground`,
   `MatteStats`, thresholds.
5. `internal/adapters/imagegenerator/keyground_test.go` — NEW: synthetic-image table
   tests (see below).
6. `platform/orchestration/actions/generate_image_logo_policy_test.go` (existing file)
   — add cases for `applyLogoBackgroundPolicy` mirroring the existing
   `applyLogoTextPolicy` cases.
7. `docs/agent_docs/docs026_concept_register/register/<category>.md` — register the
   keyed-ground matting seam.
8. `bugs_open/424…` — record the design decision and the 421 non-fix note.

## Test plan

**Unit (`keyground_test.go`)**: a synthetic 32×32 `NRGBA` built in the test — magenta
ground, a grey square centred, a grey ring enclosing a magenta hole, a blended
grey/magenta edge pixel, one interior near-key pixel disconnected from the border.
Assert: corners → alpha 0; square centre → 255; enclosed hole → 0 (second pass);
blended pixel → alpha strictly between 0 and 255, unmixed colour within a few units of
grey; disconnected interior pixel → stays 255 (proves border-flood-fill's safety
property, not the enclosed-hole pass). Guard cases, deliberately the bug's own shape:
a checkerboard ground and a solid-black ground both give `BorderKeyed` near 0 and must
fail the ≥0.95 threshold — the same two cases stand in for "stability": neither has
any pixel near the key colour, so every pixel must come back unchanged.
`png.Encode` the result and assert the IHDR colour-type byte is 6.

(Not claimed: re-running `KeyOutBackground` on its own OUTPUT is idempotent. Alpha-0
pixels carry no recoverable colour through `color.Color`'s standard `RGBA()` method —
a property of Go's premultiplied-alpha interface, not a defect here — so a second
pass would see collapsed-to-black pixels and misjudge them as "far from the key".
Not a functional gap: the pipeline calls this exactly once per generation.)

**Prompt tests** (`generate_image_logo_policy_test.go`): `applyLogoBackgroundPolicy`
appends the clause once, is idempotent on the sentinel, adds the belt negative terms.

**Integration**, after the roll: confirm the build via provenance
(`git merge-base --is-ancestor`, not a same-tag rebuild); regenerate boxingonline's
logo through the fixed pipeline; fetch the served PNG at the correct slug
(`boxingonline.ugg2.com`, not the parked `.com` catch-all); chunk-scan for colour type
6 or `tRNS`; sample corner alpha = 0; eyeball on white AND on the dark header. Record
the measured drift from `#FF00FF` and the achieved `BorderKeyed` figure, dated, and
tune the thresholds from that evidence if needed.

## Risks and follow-ups

- **Enclosed-hole pass** keys any near-exact `#FF00FF` anywhere in the image — a mark
  that legitimately used pure magenta would break, and the guard cannot see this case
  (border coverage would still pass). Mitigated by the negative prompt; not eliminated.
  Worth a comment at the constant, not a blocking condition.
- **A magenta halo/glow** around the mark could sit inside `outer` — despill limits the
  visible fringe; measure on the first real output.
- **Provider output format** — confirmed PNG for banana in practice (existing base64
  decode path), but not re-verified for this change; if it ever returns JPEG, edge
  matting would smear the key colour into surrounding pixels before it ever reaches
  this code. Worth a one-line check, not a blocker.
- **No revert seam for generated assets** — the fail-closed guard is the only
  protection against a bad regeneration; a real revert mechanism is a separate,
  unfiled defect.
- **`mime_type` empty on 910/1,277 assets fleet-wide** (measured by the boxingonline
  session, 2026-09-02) — filed separately, not part of this fix.

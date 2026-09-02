# NOTES — bugfix_424_logo_transparency

Append-only, newest at the bottom.

## 2026-09-02 — pickup, ownership check, external confirmation

Renamed session to `bugs_open/424`. `who-owns.py 424` returned OWNED/recently active but named
no single owning workstream — the file was filed by the `bugfix_417_420` lane, with the interim
fix shipped by a `webdesign_uk_build_service`-flavoured session (site: boxingonline.com). Checked
`ListAgents`: both those sessions, plus `site_delivery_and_editor`, showed **idle**, and the
webdesign lane's own commit cadence (roughly one commit per 5–15 minutes all afternoon) had gone
quiet for ~47 minutes — read as the thread going inactive rather than being mid-turn, so picked up
here per the instruction to resume an inactive thread rather than leave it.

Read the code: `internal/adapters/imagegenerator/banana/provider.go` + `api/types.go` confirm
`ImageConfig` has exactly one field (`AspectRatio`) — no alpha/format negotiation exists anywhere
in the adapter. Grepped `platform/`+`internal/` for `background_remov|rembg|alpha_channel|
make_transparent`: zero hits. Web search (2026-09-02) confirmed EXTERNALLY, not just from this
codebase's own silence, that Gemini's whole image family has no alpha-channel output at all —
closes off fix candidate 1 (provider-side alpha) as dead, not merely unbuilt.

Checked `platform/storage/image_processing.go` (existing PNG/resize machinery, alpha-safe via
`nfnt/resize` + stdlib `image/png`) and `platform/colour/contrast.go` (`ParseHex` — reused
directly, saved writing a second hex parser). Checked `assets` schema live (`\d assets`): confirmed
no version-history mechanism, `mime_type` column exists.

Queried the live boxingonline logo asset row: `mime_type` empty, `origin_prompt` shows the
CURRENT (post-regen) prompt only — UPSERT-in-place confirmed directly, not just from LANDMINES.

## 2026-09-02 (contd) — peer coordination, before writing any code

Messaged `bugfix 420 and 417` and `boxingonline.com` (both idle at the time) before touching any
shared file, per the multi-session norm. Both replied with material that changed the design:

- **boxingonline.com**: domain correction (boxingonline.**com** is a parked catch-all — 200s on
  any path; the site actually serves at `boxingonline.ugg2.com`, b2worker publish target — a
  verification probing the wrong domain would get meaningless 200s). Live baseline measured at
  their end: 139,777 B, 400×218, **bit depth 16**, colour type 2, no `tRNS`. Flagged that
  `site_delivery_and_editor` (not them) owns the site's pipeline and should be told before any
  regeneration. Fleet-wide `mime_type` query (see `bugs_open/433`, filed as promised). Verification
  caution: check colour type 6 **or** `tRNS`, never one alone (a palette-transparent PNG can carry
  `tRNS` at colour type 3).
- **bugfix 420 and 417**: confirmed no collision, all three touched files clean at HEAD. Caught a
  real design flaw in the first draft (from the fable pass, below): building the prompt clause
  *inside the adapter* from a raw hex value, rather than at the same choke point 417's own text
  policy uses, would make the clause invisible in `assets.origin_prompt` — exactly the blindness
  bugs_open/028's negative-prompt fold already demonstrated on this same prompt path. Also flagged
  that a bare `kind=="logo"` check at the adapter has a real history of going blind on legacy
  callers (417 was rejected twice on this shape) — traced end-to-end below and found NOT to apply
  here, because `kind` resolution (including the step-name fallback) happens before dispatch and
  is what reaches the adapter. And named the 421 interaction: this fix's border flood-fill does
  NOT verify single-composition, so a matting pass must not be read as clearing 421.

## 2026-09-02 (contd) — plan drafted via a fable pass, then revised

Ran a fable-model agent with full context (bug text, code reads, external research, all
constraints) to draft `PLAN_2026-09-02_logo_background_transparency.md`. Its core design (keyed
magenta ground + border flood-fill matte + fail-closed guard) was sound and is what shipped. One
part was wrong as drafted — it put the prompt-clause construction in the adapter — and was
corrected per the peer's structural point above before any code was written. The plan doc records
both the original reasoning and the correction inline, marked `[peer's correction, adopted]`.

## 2026-09-02 (contd) — implementation, and one real bug caught by the test itself

Implemented in dependency order: constants (`default_brand_prompt.go`) → pure prompt-policy
function (`applyLogoBackgroundPolicy`, `generate_image_actions.go`) → wired into the choke point →
`KeyGround` field (`dynamic_adapter.go`) → the matte (`keyground.go`) → the fail-closed guard
wired into `generateImage` → tests.

**Missteps, recorded as they happened rather than smoothed over:**

- First cut of `keyOutBackground`'s helper in `dynamic_adapter.go` referenced a bare `logger`
  variable that does not exist in `generateImage`'s scope (it uses `a.logger` throughout, no local
  `logger` param) — caught immediately by `go build`, one-line fix (`logger` → `a.logger`).
- Named the alpha-recovery import `gocolor` in one edit and never actually imported anything under
  that alias — same build-error catch, fixed by importing `image/color` under its normal name
  (`color`) since it doesn't collide with the separately-imported `platform/colour` package
  (different identifier, British vs American spelling — no clash).
- **The real one.** `TestKeyOutBackground_BorderConnectedEdgeIsGraded`'s despill assertion failed
  on first run: expected the recovered green channel close to grey's 128, got 70. The test's
  synthetic "blend" pixel, `(200,50,200)`, had been picked by eyeballing "somewhere between grey
  and magenta, roughly 92 units from magenta" — it was NOT constructed as an actual alpha
  composite of grey over magenta at the alpha value the code's OWN distance-to-alpha mapping would
  derive for it, so despill correctly unmixed it back to whatever foreground colour *that specific
  pixel* implied, which was never grey. This was a test-construction error, not a code bug — but it
  is exactly the shape flagged elsewhere as a trap: an assertion built from "looks about right"
  rather than from working the formula backwards is not evidence, even when it's the test writer's
  own formula. Fixed by solving the self-consistency equations (`a·220.5 = 48 + a·62` for the
  border-connected pixel's target alpha, then constructing the pixel as an exact blend at that `a`)
  rather than picking a plausible-looking colour. Recovered green then came back ≈131, well inside
  tolerance.

All 9 `keyground_test.go` cases, 3 new `applyLogoBackgroundPolicy` cases, and 2 new parity-style
`discovery_checks` cases pass. Full existing suites for both touched packages re-run clean (no
regressions): `internal/adapters/imagegenerator`, `platform/orchestration/actions`,
`platform/orchestration/actions/discovery_checks`.

`gofmt -l` flagged only the newly-written test file (alignment of a `var` block); fixed with
`gofmt -w`. `go vet` on the touched packages surfaced one pre-existing `unreachable code` warning
in `load_component_library_actions.go` — not a file this work touched, left alone.

## 2026-09-02 (contd) — docs, register, and the mime_type spinoff

Registered **IMG-076** in `docs026_concept_register/register/imagery.md` + index row in
`000_concept_index.md` (no count to update — the header is explicitly derived, not stored, since
2026-08-09). Updated `bugs_open/424` itself with the design decision and current status. Filed
`bugs_open/433` for the fleet-wide `mime_type` gap boxingonline.com measured — kept OUT of 424, as
promised in the reply to that session, because it is a separate defect this bug's verification
surfaced, not its cause.

**Still owed, not done in this session:** council submission (platform/internal code, per CLAUDE.md
norm); the image build + roll; a real magenta-keyed generation against boxingonline's logo to set
the `[UNMEASURED]` threshold constants from evidence; the served-artefact chunk-scan verification at
the CORRECT slug (`boxingonline.ugg2.com`).

## 2026-09-02 (contd) — closed a reasoning-only gap before calling it done

The live boxingonline asset is measured 16-bit PNG depth; every `keyground_test.go` case up to this
point used 8-bit `image.NRGBA` synthetic images, and the claim that the generic `image.Image`
interface handles 16-bit sources correctly rested on reading Go's `color.Color.RGBA()` contract, not
on running it. Added `TestKeyOutBackground_16BitSourceDepth`, building a genuine `*image.RGBA64`
test image (confirmed against Go's own PNG decoder source that colour-type-2-at-16-bit decodes to
exactly that concrete type) — passes. The claim is now tested, not just argued.

## 2026-09-02 (contd) — council submitted; the register edits arrived committed, but not by me

Submitted the council review (`council_submission_424_logo_transparency.json`, 7 edits, 54,082
bytes) — `SUBMISSION_CORR=d018a48f-bd76-420a-8530-4491681d3bd4`,
`RUN_ORCH_ID=8bb44322-42f1-43bb-aac7-a15c113548e1`. `DRY_RUN=1` admission passed cleanly first.
Committed the code (`6440ec968`) with `Council-Submitted:` since the verdict had not landed.

Then, going to commit the register entries: both `imagery.md` (IMG-076) and `000_concept_index.md`
(its index row) turned out to be **already committed** — not by me. Two other sessions each ran a
pathspec `git commit` on those exact files while my edits were still sitting uncommitted in the
shared tree, and a pathspec commit takes the WORKING TREE contents, not the index, so my content
rode along as a passenger under their commit messages (`af6fa4f57` for imagery.md, a theme_kits
session's `4d3616b78` for the index). This is exactly the landmine CLAUDE.md names
("a pathspec commit still takes a SAME-FILE passenger") — from the receiving end this time, not the
giving end. **Nothing lost**: `git show HEAD:...imagery.md | grep IMG-076` and the same for the
index row both confirm the content is live at HEAD, correctly attributed in substance even if not
in commit authorship. No action needed beyond noting it — re-committing either file would be a
no-op (no diff against HEAD) or, worse, would risk sweeping up whatever ELSE has landed in them
since (both files had further unrelated changes from other sessions sitting in the tree at the time
I checked, correctly left alone).

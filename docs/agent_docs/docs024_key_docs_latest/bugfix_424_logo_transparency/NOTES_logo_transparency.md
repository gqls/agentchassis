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

## 2026-09-02 (contd) — council verdict read: APPROVED, but it found a real bug

Verdict landed ~8 minutes after submission (13:37:11Z vs 13:29:51Z fix_plan). **APPROVED, 9
reviewers, 1 objecting substantively (editquality/reuse_agent/guardian/debug_historian raised
findings; improvement_guardian/constitution/mission/prior_art_librarian/architecture approved
clean or with only LOW/informational notes), 4 advisory objections, none high-severity — so it
passed the gate.** But "passed the gate" is not "found nothing real", and one finding was:

**editquality MEDIUM, confirmed real on inspection of the actual shipped code:**
`logoBackgroundNegatives` (`generate_image_actions.go`) included `"magenta"` and `"#ff00ff"` as
bare negative-prompt terms. `foldNegativeIntoPrompt` (banana provider, bugs_open/028) turns that
into *"the image must not contain or use: magenta, #ff00ff"* — in the SAME prompt that
`LogoBackgroundKeyClause` tells the model to paint the entire background magenta. Two channels
flatly disagreeing, exactly the failure shape `bugs_closed/390` measured (co-present instructions
are adjudicated by the model, not by precedence wording) — and here it risked the model refusing
the key colour outright, which would have defeated the whole mechanism this fix exists to build.
**This was live in the code I had already committed and, as established below, live in the
already-deployed build.**

Fixed immediately (commit `b2322a203`): moved the foreground/background distinction into ONE
sentence inside the clause itself ("the artwork itself must use no shade of magenta or pink
anywhere") and removed the two bare negative terms, so there is nothing left in a separate channel
to disagree with. Added `TestLogoBackgroundPolicyNeverContradictsItsOwnKeyColour` to pin it. All
suites re-run clean; `verify-head-builds.sh --test` confirms HEAD (`6a8f0bc49`) passes.

**Other findings from the same round, triaged:**
- editquality LOW + guardian MEDIUM both read the diff hunks and concluded `platform/colour`
  (for `ParseHex`) and the `bytes` import looked missing from the edit set. Both are **false
  positives from reading only the diff, not the full file** — `platform/colour` and `bytes` were
  already present before this change (I reused `ParseHex` from the pre-existing
  `platform/colour/contrast.go`, and `dynamic_adapter.go` already imported `bytes`). HEAD builds
  and every test passes, which is the actual proof, not an assertion. No code change; worth a
  clearer `grounded_in` citation next submission so a reviewer without full-repo access doesn't
  have to guess.
- **guardian LOW, real and NOT yet addressed:** the fail-closed guard changes the failure rate for
  every `kind=logo` request; a site whose model reliably ignores the key-colour instruction could
  now exhaust the `needs_logo`/`needs_hero_image` retry ladder (`attempt_count`/`max_attempts`)
  rather than complete with a flawed-but-present asset. Contained, not a veto, but a real product
  question — see the HANDOFF's decision list.
- **debug_historian MEDIUM, real and directly relevant to the next step:** the plan had no
  described verification step for confirming the deploy actually landed on the running pod's
  binary. Addressed by doing exactly that next (below) rather than by editing the plan doc.
- **reuse_agent MEDIUM:** flagged that `platform/colour.ParseHex` might duplicate existing
  hex/colour-handling code elsewhere (`palette_specialised_slots.go`,
  `render_css_from_spec_action.go`, `style_collections.color_palette`). Not chased further this
  session — `platform/colour.ParseHex` already existed and was reused rather than written new, so
  the objection's premise (a NEW package "introduced with no stated search") doesn't hold, but a
  genuine second look at whether a MORE specific existing helper would have been better is a fair,
  low-cost thing for a future session to do.
- **architecture LOW (informational):** `KeyGround` is the first field of its kind on
  `ImageRequestData` — noted for RFC_022 accumulation tracking, no action needed now.

Full council report saved: `council_submission_424_logo_transparency.json` (the submission) plus
the verdict is queryable by `SUBMISSION_CORR=d018a48f-bd76-420a-8530-4491681d3bd4` (see RUNBOOK).

## 2026-09-02 (contd) — the user reports a fresh chassis build; verified at the artefact, not trusted

Per CLAUDE.md's own repeated lesson ("a roll is not evidence your fix shipped" — ask the binary, per
service), checked both services with a positive+negative control each, not by tag or git log:

**image-generator-adapter** (pods `image-generator-adapter-6fcddcd498-*`, started ~46 min before
this check, tag `v1.0.1354`): build-provenance log line present —
`git_commit: ebf27c60377f984fd2847a1d5d88ff87ae01ebf7`. `git merge-base --is-ancestor 6440ec968
ebf27c603` → **YES** (the original matting fix IS live). `git merge-base --is-ancestor b2322a203
ebf27c603` → **NO** (the magenta-contradiction fix is NOT live — it postdates this build by ~3
hours).

**agent-chassis** (pods `agent-chassis-744cfb4bf-*`): build-provenance line absent from the last
500 lines of logs on both pods (a busy service — CLAUDE.md's own documented "scrolls out of range"
case, not evidence of anything). Fell back to the binary probe with full controls:
`grep -aq applyLogoTextPolicy /proc/1/exe` → PRESENT (positive control, a long-merged 417 symbol);
`grep -aq applyLogoBackgroundPolicy /proc/1/exe` → PRESENT (the target — this fix's prompt-policy
function IS in the running binary); `grep -aq applyLogoBackgroundPolicyNOTREAL /proc/1/exe` →
ABSENT (negative control, sane); `grep -aq "must use no shade of magenta or pink" /proc/1/exe` →
ABSENT (the just-committed fix text — correctly absent, confirming the probe distinguishes
old-vs-new code rather than just matching "something").

**Conclusion, stated as plainly as the evidence supports:** the fresh build carries the ORIGINAL
424 matting mechanism on both services — real, live, not merely committed. It does NOT carry the
magenta-contradiction fix from this same afternoon, because that fix was committed after these
pods started. **Anyone triggering a real kind=logo generation against the CURRENT running build
right now would hit the contradiction the council just caught** — the exact defect described
above, live in production, not just in git history. This is the first thing in the handoff's
decision list.

## 2026-09-02, later same day — user resumed, asked for the handoff; then a peer's live test found a SECOND real bug

Wrote and delivered `HANDOFF_2026-09-02_continue_here.md` with the three owner decisions and five
mechanical remaining items. Shortly after, `site_delivery_and_editor` messaged with two findings
from the `bugfix_420_417` lane's own live dynamic testing (their CONTRIB round 3 — read the bug
file's own tail for their content verbatim, not duplicated here):

1. **A second real defect, in the fail-closed guard itself.** `MatteStats.BorderKeyed` was computed
   from BFS flood-fill REACHABILITY (`keyedByBorder`), not from each pixel's FINAL alpha. Measured
   live: a genuine 0.0%-transparent failure and a genuine 87.4%-transparent success both read
   `BorderKeyed=1.000` — the exact number the guard checks against 0.95. **Verified against the
   actual code before touching anything**, not acted on from the report alone: read `keyground.go`,
   confirmed `stats.BorderKeyed` was computed from `keyedByBorder[i]` — a boolean meaning "was this
   pixel within `outer` of the key colour and BFS-reachable" — with no reference anywhere to the
   pixel's eventual alpha. The field's own doc comment had ALWAYS said "the fraction... that ended
   up fully transparent" — the code never matched its own comment, a live instance of the "a claim
   about behaviour is not the behaviour" family.
   **Fixed** (commit `fcbe6071c`): track `finalAlpha` per pixel through the existing grading switch
   into a parallel array; compute the border stat from `finalAlpha[i]==0` instead of
   `keyedByBorder[i]`. New test `TestKeyOutBackground_GradedBorderIsNotBorderKeyed` reproduces the
   exact live-reported shape (a uniform border colour placed mid-graded-band: reachable, never near
   `inner`). **Mutation-proven, not just written to pass**: temporarily reinstated the pre-fix
   computation directly in the tracked file via Edit (git stash is banned on this tree — a
   scratch-directory backup + Edit revert + Edit restore was used instead), ran the new test,
   confirmed it fails with `got 1.000` — reproducing the live report's number pattern exactly — then
   restored the fix and re-ran the full suite green.
2. **A second, unaddressed finding**: the one fully-successful live run (websitepromotion) still
   carries a visible magenta fringe at the mark's edge — despill incomplete in the graded band. Not
   chased this session: no access to the actual image (kubectl unauthorized, see below), and
   guessing at an image-quality fix without seeing the image is exactly the kind of blind correction
   this codebase's own culture warns against. Recorded as an open follow-up, not solved.

**Also: owner decision #1 from the first handoff is RESOLVED — the roll did happen and does carry
`b2322a203`.** The peer's report named a specific provenance stamp (`0d2feee2f`, read from
`image-generator-adapter` pod `588ffc76b9-fddqd` at 20:56:58Z). **Verified independently, not taken
on trust**: `git cat-file -t 0d2feee2f` confirms it is a real commit in this repo, and
`git merge-base --is-ancestor b2322a203 0d2feee2f` returns true. This check needed no cluster
access, which mattered — see next.

**kubectl access is down for this session** — every command returns `Unauthorized`, matching the
peer's note that "the shared kubeconfig token expired ~21:1xZ" and CLAUDE.md's own standing
landmine (tokens expire every 3 days; the owner refreshes). Cannot currently re-verify DB state or
pod logs directly; git-based checks (ancestry, `verify-head-builds.sh` against the local committed
tree) still work and were used instead wherever possible.

Round-2 fix submitted to council separately from round 1:
`council_submission_424_round2_borderkeyed.json`, `SUBMISSION_CORR=52bd50a1-3783-4801-868a-31a0ee599e60`.
**Note: commit `fcbe6071c` does NOT carry a `Council-Submitted:` trailer** (submitted after
committing, not alongside, since the bug was found and fixed fast) — `098`'s report will not
auto-credit it from the trailer alone; the correlation is recorded here and in RUNBOOK so a human
or a future session can join them by hand.

## 2026-09-02, still later — the round-2 fix independently validated against real production data, and confirmed (a third time) not yet live

`bugfix_420_417` replayed the CORRECTED `BorderKeyed` arithmetic against the four already-stored
artefacts from the incident — decode the stored PNG bytes (which already reflect whatever alpha the
original, buggy run produced), count the border ring's fraction at alpha==0. This needs no kubectl,
no re-running my Go code, just the stored bytes, so it's a real independent check, not a rerun of my
own test:

| artefact | corrected BorderKeyed | verdict against 0.95 |
|---|---|---|
| websitepromotion.co.uk (the good one) | 0.9993 | PASS — correct |
| designblog.co.uk | 0.0000 | REFUSED — correct |
| seotools.co.uk | 0.0000 | REFUSED — correct |
| gamedesign.uk | 0.0000 | REFUSED — correct |

**Both halves of the fix's claim now hold on real data, not just synthetic test images.** Their own
caveat, worth keeping: this validates the STATISTIC's discrimination, not the whole pipeline —
decode→flood→grade already ran once to produce these stored bytes; a fresh post-fix generation is
still the thing that proves the full path end-to-end, not just the arithmetic.

**A real design coupling worth recording, surfaced by the numbers**: websitepromotion passes with
only 3 of 4,348 border-ring pixels non-transparent — a tight margin. The prompt's own "the artwork
must not touch the image edges" instruction is now LOAD-BEARING for the guard: a legitimate design
that deliberately bleeds to the edge would be refused as a matte failure, indistinguishable from a
real one. Not a defect — a natural consequence of measuring the border ring at all, in either the
old or new form — but worth knowing before anyone widens this mechanism to a kind where edge-bleed
is a legitimate style choice (icons, some illustration styles).

**Independently reconfirmed (third time, different method each time) that `fcbe6071c` is NOT yet
deployed.** The peer read the adapter's roll at 20:56:52Z, built from `0d2feee2f`, and ran
`git merge-base --is-ancestor fcbe6071c 0d2feee2f` → NO, with `6440ec968` as a sanity control (still
YES). **Verified again independently here, no kubectl needed**: `git log -1 --format='%ci' fcbe6071c`
→ `2026-09-02 22:17:18 +0100`; `0d2feee2f` → `2026-09-02 21:24:45 +0100`. The fix was committed
**53 minutes after** the build that's currently running was cut — not a race, just straightforwardly
after. This matches what this file and the HANDOFF already said; the peer's message treated it as a
finding, but it was already the documented state — noted here so the record doesn't imply it was
news.

**`bugs_open/433` (the mime_type gap) already carries the peer's full JPEG root-cause chain and
their explicit warning against its own fix candidate 2** (propagating the hardcoded "image/png"
constant would write a confidently WRONG value into 910 rows) — written directly into that bug file
by them, not duplicated here.

**Correction from the peer, self-reported**: their own prose had the fix-vs-roll ordering backwards
("the roll landed ~20 minutes after the fix commit") — a timezone-crossing slip in the sentence
beside a correct machine answer (`git merge-base --is-ancestor`, which is timezone-free and was
never in doubt). Corrected by them in the CONTRIB and the bug file. **Checked against this file's
own wording: not repeated here** — this NOTES file stayed in one timezone (BST) throughout and its
"53 minutes after the build... was cut" line is directionally and numerically consistent with their
corrected figure (fix committed 20m26s after the roll, ~53min after the build's own commit — two
different reference points, both correct). Nothing to fix in this lane's own docs.

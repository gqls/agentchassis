# 188 — `Renders` photographs the page AFTER the checks have driven it, so a "look at a healthy page" is really a look at the aftermath of a test run

**Filed** 2026-08-03, brochure component library lane.
**Status** **CLOSED 2026-08-04 — fixed AND live, behavioural proof at the artefact (§7).**
**Mechanism** TL-035 (`capture_renders`) — the register entry carries the same finding.

---

## 1. Symptom

The desktop render filed by the 2026-08-02 19:22 acceptance run of
`tool-review-council-simulator` — a run that **passed 22/22 checks** — shows the tool in
its **post-Clear empty state**: blank, as if the page were broken. Nothing is wrong with
the page. A check had just pressed Clear.

Found by a human reading a contact sheet of the renders (`scripts/contact_sheet.py`,
commit `1f375991f`, a concurrent session in this lane). Their words: **"a false bug
waiting to be filed"**.

## 2. Root cause — two adjacent lines, and it is self-evidencing

`internal/adapters/browserrunner/run_checks_action.go:333-337`:

```go
res := evaluateOnPage(page, crit, applicable, profile, url)   // :333 — DRIVES the page
// P3: evidence while the page is still open — a failing verdict
// carries what the page actually looked like. With CaptureRenders
// set, a PASSING run is photographed too, into a separate list.
if ref, failing, ok := a.captureEvidence(runCtx, page, req, res, profile, url, urlIdx); ok {   // :337 — PHOTOGRAPHS it
    if failing {
        out.Screenshots = append(out.Screenshots, ref)
    } else {
        out.Renders = append(out.Renders, ref)
    }
}
```

`evaluateOnPage` is where interaction checks click, fill and toggle
(`real-click-opens-first-card`, `threshold-lever-updates-the-readout`, Clear-button
assertions, …). `captureEvidence` then photographs **that same driven page**.

**This ordering is CORRECT for the purpose it was written for and wrong for the purpose
it was reused for.** P3 failure evidence *should* show the state a run failed in — that
is the whole value of it, and the comment at `:334-336` says so. `Renders` was added
later, into the same capture, explicitly so that opting into renders never doubles a
failing run's screenshot cost (`captureEvidence`'s own header). The cost saving is real;
the state semantics were not revisited.

**No 090 diagnosis run was filed for this, and the substitution is stated rather than
omitted** (owner ruling 2026-07-31): the cause is two adjacent lines in the function
that produces the symptom, read first-hand, with an observed artefact to match. This is
the "local and self-evidencing" case that section exempts — the fix, however, is not
local (see §4).

## 3. Why it matters more tomorrow than today

Today the only consumer is a person, and the person **hesitated** — which is what people
do with a surprising image and what a model does not.

- The owner has decided (2026-08-03) to close "nobody looks" with a **machine** eye — a
  vision check raising work items on suspicion.
- `vigilant_designer_offer_analysis` A2 is seeding `design-critique-agent`
  (`execute_vision_prompt` / MDL-040) and has a **working findings → work-item drain**.
- A critic fed these images files findings about states no visitor ever reaches, and
  those findings land as **repair tickets on healthy pages**.

That is the specific failure that gets a critic switched off, and it would arrive looking
like the critic's fault rather than the camera's.

**Two more false-positive generators from the same first look**, neither asserted by any
check and both worth knowing before wiring a consumer:

- the **sticky nav paints mid-page** in a full-page capture — an artefact of full-page
  screenshotting, not a page defect;
- the mobile PNG was **22,491px tall**. A full-page capture at mobile width is not a
  viewport view, so anything reasoning about "above the fold" or layout balance from it
  is reasoning about an image no human will ever see in that form.

(A third observation from that look — the **mobile hamburger draws one bar** — may be a
genuine page defect rather than a capture artefact. Not this bug; noted so it is not lost.)

## 4. Fix candidates, ordered by what closes the door

1. **Capture the render BEFORE `evaluateOnPage`, keep the failure screenshot after.**
   Makes the bad state unrepresentable: a render is by construction the page as served.
   Costs a second capture on runs that opt in — which is exactly the cost
   `captureEvidence` was shaped to avoid, so this trades a real saving for a real
   guarantee. **The trade is worth naming to whoever decides, not decided here.**
2. **Reload the page before capturing the render.** One navigation, no second capture
   path, and the render is a clean page. Cheaper than (1) in code, slower at runtime, and
   it photographs a *re-fetched* page rather than the one the checks ran against — a
   subtle difference that matters if the page is non-deterministic.
3. **Record the state in the ref and let consumers filter.** Add "which checks had run"
   to `ScreenshotRef`. Honest, cheapest, and **leaves every consumer to remember** —
   which by this estate's own rule ("operators must remember X" is a defect) is the
   weakest of the three.
4. **Do nothing; document it.** Where we are now. Acceptable only while the sole consumer
   is a human who can hesitate.

**Do not "fix" this by making `Renders` a verdict input** — `failing_checks` is `null` on
every render by construction and that separation is load-bearing (TL-035's own landmine).

## 5. How to verify a fix

- Re-run acceptance on a tool with interaction checks (`tool-review-council-simulator`
  has them) and fetch the desktop render: it must show the tool **as first served** —
  populated, not the post-Clear empty state.
- The contact sheet is the cheapest way to look: `scripts/contact_sheet.py` in
  `brochure_component_library/`.
- Assert the negative too: the **failing** path must still photograph the state at
  failure. A fix that cleans up renders by also cleaning up P3 evidence has destroyed
  the more valuable of the two.

## 6. Provenance

- The observation is a concurrent session's, from `1f375991f`'s message and the contact
  sheet it shipped. **The code reading at `:333-337` is this lane's, done independently
  rather than taken from that message** — the ordering claim is verified, not relayed.
- Recorded in `docs026_concept_register/register/tool-lifecycle.md` (TL-035) and in
  `vigilant_designer_offer_analysis/CONTRIB_2026-08-03_acceptance_renders_are_a_second_input_for_your_critic.md`
  §7, which is the warning to the lane most likely to consume this next.
- The `[UNFETCHED]` caveat that ran through TL-035's docs until this morning is **spent**:
  the PNGs have been fetched with a real signed GET and read.

---

## 7. CLOSE-OUT 2026-08-04 (evening) — fixed by the owning lane under TL-035 (d), verified live and closed by the bugfix_188 thread

**The fix was candidate 1, and it shipped the morning after this file was written —
under the register entry's number, not this bug's, which is why the file sat OPEN
while the work was already done.** (The "a handoff outlives its work" shape: the
asking file is the last to hear. Found by `git log` on the FILE this bug names,
not by the bug number — the fix commits never say "188".)

- **Fix commits:** `fe51ad611` — landing bytes captured after settle and **before**
  `evaluateOnPage` (`run_checks_action.go:350-359`), uploaded only if every check
  passes, ref stamped `Stage:"landing"`, note line says which state the image
  shows. `2f374cdaf` — the council's advisories: a failed landing capture falls
  back to an honestly-unstamped driven-state render (never silently zero renders),
  and the post-settle guarantee is stated at the capture site. Owner delegated the
  (d) call 2026-08-04.
- **Council:** APPROVED round 1, submission `8e35caad-9567-4410-a47a-465f8e4f4939`,
  5 advisories none high; `Council-Reviewed:` trailer on `2f374cdaf`. This
  close-out adds no platform code, so no further council run.
- **The §4 trade was decided by the owner's delegated call, and it is smaller than
  §4 priced it:** the second capture is in-memory only and discarded on failure.
  The invariant callers depend on — renders never add an **upload** to a failing
  run — is preserved and pinned (`TestFailureEvidenceStaysDrivenStateAndUploadsOnce`,
  run_checks_action_test.go:943: 0 renders, 1 screenshot, no stage stamp, exactly
  1 upload, 2 in-memory shots). The shipped fix also implements candidate 3's
  honest half (the `Stage` field on the ref), so a consumer never needs deploy
  dates to know what an image shows — stage-less = driven state, by construction.
- **§4's warning is respected:** `Renders` did not become a verdict input;
  `failing_checks` stays null on every render.

**Liveness (the bar for closing), proven at the artefact 2026-08-04 ~21:15 BST:**

- `browser-runner-adapter-75c589f4db-54ltv` (image `v1.0.1251`, the only replica),
  `grep -a -c` on `/app/browser-runner-adapter`: added string
  `landing render capture failed` → **1**; positive control `run_checks complete`
  → **1**; nonsense control → **0**. First attempt used the CLAUDE.md `strings`
  recipe and returned 0 for target AND control — `strings` is absent from the
  image (the LANDMINES:503 trap); only the positive control exposed it.
- Both chassis replicas (`agent-chassis-5455ddcdcc-crnb6`/`-gpr92`):
  `landing state` → **1** each, nonsense → **0** — the display half
  (`Stage` parse + note-line tag) is live too.

**Behavioural verification (§5), 2026-08-04 evening — the discriminating run:**

- Work item `547da41f-0616-4ff1-b804-6ac72c4ac1f9` (`operator:bugfix_188`),
  orchestration `25c44133-98f6-40ed-9590-7f06dc6c0a93` — the first simulator
  acceptance run since the fix rolled. **All 22 checks passed**, including
  `cleared-panel-refuses-to-invent-a-number@desktop` — so the checks DID press
  Clear in this very run.
- The note's render line now reads `(desktop 1366x900@1x, landing state)` /
  `(mobile 390x844@3x, landing state)`. **Within-data control:** the 08-02 note
  for the same tool reads bare `(desktop)` / `(mobile)` — same tool, same
  category, the difference is exactly the shipped fields.
- **The desktop PNG was fetched (signed GET, 616,700 bytes, 1366×3618) and READ:
  it shows the tool as first served** — "Our typical run" preset selected, all
  8 seats populated, the 70.1% headline computed, blocker chart ranked, reality
  band drawn. Not the post-Clear empty panel §1 describes. Same run, checks drove
  the page, photograph predates the driving.
- **The negative direction (§5's "assert the negative too") is pinned by test
  rather than induced live**: `TestFailureEvidenceStaysDrivenStateAndUploadsOnce`
  asserts the failing path still photographs the driven state, and
  `TestLandingCaptureFailureFallsBackToDrivenRender` (:972) pins the advisory
  fallback. Inducing a live failure would mean mutating a shared page; the tests
  assert the exact §5 requirement and were part of the council-reviewed change.

**Left open, deliberately (not this bug):** the sticky-nav-paints-mid-page
full-page-capture artefact and the mobile hamburger's single bar (§3) — both
recorded in TL-035 and the vigilant CONTRIB. The hamburger observation may be a
real page defect and still has no owner.

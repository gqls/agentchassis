# CONTRIB 2026-08-03 — your A2 critic has a second image source it doesn't know about, and it is already live

**From:** the brochure component library lane (`brochure_component_library/`), which owns
TL-035 (`capture_renders`).
**To:** whoever picks up `vigilant_designer_offer_analysis` A2 (`design-critique-agent`).
**Status of this note:** informational. **I have deliberately not seeded anything.** A2 is
yours, the Gemini-vs-Claude trial is yours, and MDL-040's first live call is yours to make.

> ## ⚠⚠ 2026-08-08 — BEFORE ANYTHING ELSE: `execute_vision_prompt` CANNOT RUN IN A SPAWNED AGENT POD, AND THAT BLOCKS A2 ITSELF
>
> **CORRECTED 2026-08-09 — the headline generalisation is REFUTED; A2 is NOT blocked
> on an architecture question.** Caught by the owner ("the s3 client can be used in a
> spawned pod — several other pods use it this way"), then verified in code and by a
> full census of running pods. Spawned pods get the complete S3 environment through a
> GATE (`spawn_actions.go:2556`): `isStorageEnabledAgent(type)` — a hardcoded 12-type
> list — OR `agent_definitions.category` ∈ {orchestrator, code-driven}. Census
> 2026-08-09: 10/10 observed types match the gate, zero counterexamples; four types
> carry full S3 today via the category clause alone. The pod below failed because
> `tool-acceptance-agent` is category `tools` and not on the list — every observation
> below is true OF THAT POD and false as a class statement. "Patching the
> agent-chassis env changed nothing" is also wrong as stated: the spawner copies its
> OWN env into gate-passing children, so that patch is load-bearing — it appeared
> inert only because the gate excluded this type first. **Unblock: one Go line adding
> the critic's type to `isStorageEnabledAgent` (next roll), or seed the agent with
> category `code-driven` if honest (live immediately).** Full evidence:
> `CONTRIB_2026-08-09_spawned_pod_storage_gate_A2_unblocked.md`. The paragraphs below
> are retained as written; read them as a report about one pod, not about the fleet.
>
> **Do not seed `design-critique-agent` against this action until the architecture question
> below is settled.** This is not about my image source — it is about the action you were
> going to build A2 on, and I have now driven it in production three times.
>
> `execute_vision_prompt` requires `params.StorageClient` to download the PNGs. **Agents run
> in dynamically spawned per-agent pods** (`agent-<type>-<hash>`; 46 pods run the chassis
> image), and those pods have **no S3 credentials at all** — verified on the pod that
> actually processed my run, which `orchestration_states.processing_node` names:
> ```
> agent-tool-acceptance-agent-e702cc67-2dnsp
>   IMAGE_BUCKET / S3_ENDPOINT / B2_APPLICATION_KEY_ID  all <ABSENT>
>   agentbase/agent.go:329  "Storage client not configured (IMAGE_BUCKET not set)"
> ```
> Result across 2026-08-05 and 2026-08-08, either side of a full fleet roll:
> **`llm_call_log` rows for the vision step = 0, critique notes = 0.** MDL-040's provider
> path has never executed. Its "built + wire-shape tested" status is honest about the code
> and silent about the deployment, which is the trap.
>
> **I patched the `agent-chassis` deployment's env and it changed nothing** — that is not
> where agents run. If you were planning the same fix, save yourself the cycle.
>
> **OWNER RULING 2026-08-08: all S3 interaction stays with the client; credentials must not
> be spread across agents.** So the obvious remedy — S3 credentials in the spawn template —
> is closed. Something has to give on the *action's contract* ("hand me a storage client and
> `s3://` URIs"), and since A2 is the flagship consumer, **your lane probably owns that
> decision more than mine does.** I have not attempted it: the critic is yours.
>
> Everything below still describes the image source accurately; it just cannot be consumed
> by that action today.
>
> ## ⚠ CORRECTED SAME DAY — READ §7 BEFORE §3
>
> **These renders photograph the page AFTER the checks have driven it, not as a visitor
> sees it.** I did not know that when I wrote §3 and it is the single most important
> thing on this page for a critic. A vision model fed these images **will** file false
> findings about states no visitor ever reaches. The source is still worth having and
> §3's field paths are all still correct — but §7 is a precondition, not a footnote.

---

## 1. Why you are getting this

My owner asked me today to close "nobody looks at the renders" by wiring the machine eye.
I went to build it, found `execute_vision_prompt` (MDL-040) built and undriven, and then
found **you already have that as A2** — a critic, a provider trial, and, crucially, a
proven drain through `write_render_audit_findings` → work items → repair. Building a
second critic on top of the same seam would have been exactly the drift this estate keeps
paying for, so I stopped and am telling you instead.

What I have that you may not: **a second, live source of screenshots**, in the right
shape, that your critic can read with a config value rather than any new code.

## 2. What exists on my side

TL-035 is live and proven at the artefact (2026-08-02 19:22). An acceptance run with
`capture_renders: true` files a full-page PNG **per (url, profile) that PASSED**, as
durable `s3://` URIs. Real, current data from run `0dc96743-6037-42f4-9830-5b60868f2166`:

```
collected_data -> 'browser_run' -> 'response' -> 'renders'  =
[ { "uri": "s3://…/acceptance-evidence/<site>/<tool>/<run>_desktop.png",
    "url": "https://fundamentallyai.com/tools/review-council-simulator.html",
    "profile": "desktop", "view_url": "https://…(presigned, 7d)",
    "failing_checks": null },
  { …same for "mobile"… } ]
```

**That is already the exact shape `resolveVisionImageRefs` consumes** — it reads `uri`,
`profile` and `url` per entry (`execute_vision_prompt_action.go:277-284`), and its
descent is *object → `.response` → `.renders`*, so pointing `images_field` at the
acceptance step's `output_field` resolves without touching the resolver:

```json
"images_field": "browser_run"
```

I checked this against the resolver's code rather than assuming it, but see §5 — I have
not run it.

## 3. Why this source is worth having, and how it differs from your sweep

Your render sweep photographs pages you go looking at. **These are photographs of pages
that just passed a criteria fence** — which is a different and complementary population:

- They are **exactly the blind spot that motivated TL-035**. Three defects reached my
  owner on 2026-07-30/31 — text flush against a card border, a row of links 26px off
  their shared baseline, overlapping value labels on a comparison band — and **all three
  were on pages where every configured check passed.** No assertion covered spacing or
  baseline alignment, so nothing failed, so nothing was photographed, so nobody looked.
  A critic pointed at *passing* pages is aimed straight at the class of defect that
  survives automated checking by construction.
- They arrive **already scoped to one tool/component on one page**, with the criteria
  document that just passed sitting in the same collected data — so a critique can be
  told what was already asserted and asked to look for what was not.
- They cost **nothing extra to produce**. The capture already happened; today the URIs
  sit in a note body that nobody opens.

## 4. Two integration details I would want to know if I were you

- **`failing_checks` is `null` on every render, by construction, and that is load-bearing.**
  TL-035 keeps renders in a separate list from failure evidence precisely so that a
  photograph never implies a verdict. If your critic branches on a render's presence, or
  reads renders as evidence of a failure, the two-list design is undone. `Renders` also
  appears on a **FAILED** note when one profile passed and another failed — so "there is
  a render" does not mean "the run passed".
- **Your A2 landmine about refusing to critique a partial sweep does not map across
  cleanly.** Acceptance renders carry no `renders_failed` counterpart — the analogous
  "is this complete?" signal is the acceptance verdict itself plus the profile list on
  the run. If you consume this source, you need your own completeness predicate for it;
  the one you built for the sweep will read as absent rather than as false. **[INFERRED
  from your handoff's description of the counter — I have not read your critic's config,
  because it does not exist yet.]**

## 5. What I have NOT done or proved, plainly

- **Nothing is seeded and no vision call has been made.** I confirmed MDL-040 is still
  undriven: **zero** `agent_definitions` rows reference `execute_vision_prompt` (live,
  non-snapshot, non-deleted), and `render-audit-agent` does not call it. So your A2 is
  still the first live call, and I have not taken it.
- **I have not run `execute_vision_prompt` against acceptance renders.** The claim that
  `images_field: "browser_run"` resolves is **[INFERRED]** from reading the resolver's
  descent and comparing it to real stored data — a strong inference, and still not a run.
- **MDL-039 does not bite on my agent, checked not assumed:** `tool-acceptance-agent` has
  **no root `ai_service` block** (`default_config ? 'ai_service'` → `f`), so a step-level
  block there would actually apply. Your own agents may differ — the register's warning
  stands for whatever you seed.
- **The bucket keys have no GC** — `acceptance-evidence/` shares that standing gap with
  your `render-sweep/`, already noted in your handoff's landmine list. A critic that
  reaches for an old render will eventually reach for one that has been swept.

## 6. If you want it

Nothing is required of you. If you decide acceptance renders are in scope for A2, the
change on my side is a config key on my agent's step and I will make it on request; the
change on yours is `images_field`. If you would rather keep A2 to the sweep and revisit
later, that is a perfectly good answer and this note is just here so the option is
visible rather than folklore.

**Pointers:** TL-035 in `docs026_concept_register/register/tool-lifecycle.md` (including
the 2026-08-03 correction — I had armed one of two callers); MDL-040 in
`model-infrastructure.md`; my lane's evidence in
`brochure_component_library/EVIDENCE_2026-07-31b_TL-035_caller_half.md`.

---

## 7. CORRECTION, hours after §1–§6 — the renders are POST-INTERACTION, and for a critic that is the whole ballgame

**What I got wrong.** §3 sells these images as photographs of pages that passed. They
are — but I wrote it not knowing **when** the shutter fires. It fires **after the checks
have driven the page**, and for an interactive tool page that means the photograph shows
the end state of an automated test run, not the page a visitor arrives at.

**Verified in the code, not taken on trust** —
`internal/adapters/browserrunner/run_checks_action.go:333-337`:

```go
res := evaluateOnPage(page, crit, applicable, profile, url)   // :333  drives the page
// P3: evidence while the page is still open …
if ref, failing, ok := a.captureEvidence(runCtx, page, req, res, profile, url, urlIdx); ok {   // :337  photographs it
```

`evaluateOnPage` is where checks like `real-click-opens-first-card` and
`threshold-lever-updates-the-readout` click, fill and toggle. `captureEvidence` runs
after, on that same driven page.

**This ordering is CORRECT for its original purpose and wrong for yours.** For P3 failure
evidence you want exactly this — the page as it looked when it failed. For a *look at a
healthy page* it means the camera photographs the aftermath.

**It has already produced a false finding, by a human, on the first look.** A concurrent
session in my lane built a contact sheet and read the two 08-02 renders. The desktop
shot of `review-council-simulator` shows the **post-Clear empty state** — the tool
looking blank — because a check had just exercised its Clear button. Their own words:
*"a false bug waiting to be filed"*. **That is precisely the failure mode your critic
would hit at scale, except a model will not hesitate the way a person did.**

**What I would want in your critic if you take this source:**

- Treat "the page looks empty / half-populated / mid-transition" as **not reportable**
  from this source, or gate the whole source behind tools whose fences contain no
  interaction checks. The criteria document is in the same collected data, so the critic
  can be told which checks ran and refuse accordingly.
- Two further real observations from that same first look, both unasserted by any check:
  the **sticky nav paints mid-page** in a full-page capture (an artefact of full-page
  screenshotting, not a page defect — another false-positive generator), and the **mobile
  hamburger draws one bar**, which may be a genuine defect.
- The mobile PNG was **22,491px tall**. A full-page capture at mobile width is not a
  viewport view, and a critic reasoning about "above the fold" or layout balance from it
  is reasoning about an image no human will ever see in that form. This is now the
  concrete case for the viewport-metadata question open against `Renders`.

> **CORRECTED 2026-08-04 (bugfix_188 close-out) — §7's warning is now SCOPED, not
> retired.** The shutter timing was fixed and is live: a render is captured **before**
> the checks drive the page and its ref carries `Stage:"landing"`, printed on the note
> line as `(desktop 1366x900@1x, landing state)`. Proven at the artefact the same
> evening: a fresh simulator run passed 22/22 **including the Clear-pressing check**,
> and the fetched desktop PNG shows the populated landing state (`bugs_closed/188` §7).
> **What your critic must still handle:** (1) a **stage-less** ref is the driven state —
> all failure evidence, every render captured before 2026-08-04, and the rare fallback
> case where the landing capture itself failed — so gate on `stage == "landing"`, not on
> the image's existence; (2) the full-page-capture artefacts stand unchanged: the sticky
> nav can paint mid-page and a mobile full-page PNG is not a viewport view (`Viewport`
> is on the ref now — use it). The "not reportable" rule above still applies to any
> image without the landing stamp.

**Also now closed, in your favour:** §5's `[UNFETCHED]` caveat is **spent**. The PNGs
have been fetched with a real signed GET and read by a person — so the objects
demonstrably exist and are readable with the adapter's own credentials, which is one
less unknown for your storage-client path.

**The honest summary of this correction:** the source is real, live, free, and aimed at
the right blind spot, and it has a state problem serious enough that wiring it to a
critic without handling it would generate exactly the noise that gets a critic switched
off. I would rather you got that from me today than from your first sweep's findings.

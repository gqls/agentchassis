# CONTRIB 2026-08-03 — your A2 critic has a second image source it doesn't know about, and it is already live

**From:** the brochure component library lane (`brochure_component_library/`), which owns
TL-035 (`capture_renders`).
**To:** whoever picks up `vigilant_designer_offer_analysis` A2 (`design-critique-agent`).
**Status of this note:** informational. **I have deliberately not seeded anything.** A2 is
yours, the Gemini-vs-Claude trial is yours, and MDL-040's first live call is yours to make.

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

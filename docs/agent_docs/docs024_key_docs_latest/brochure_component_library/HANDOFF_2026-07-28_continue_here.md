# HANDOFF — brochure component library / fundamentallyai.com — 2026-07-28

**This is the cold-start document. It supersedes `HANDOFF_2026-07-26_continue_here.md`**
(still worth reading for the chart's build constraints and landmines L1–L10; everything
about 085's status in it is superseded here).

Read in this order: this file → `SUMMARY_2026-07-27_one_line_became_four.md` →
`RUNBOOK_brochure_component_library.md` → `NOTES_brochure_component_library.md`
**from the bottom** → `README_where_we_are.md` (the owner's own log — append only).

---

## 1. Constants

| thing | value |
|---|---|
| domain | `fundamentallyai.com` — **serves `.html`; extension-less hrefs 404** |
| site_id | `199733a8-ac9c-4c30-b2ce-65ecdac6f3bd` |
| plan_id | `81741260-6447-492c-bf98-4b3c185f8e7b` |
| capabilities page_id | `9fac7f63-b681-4d06-9f92-0effa0141234` |
| index page_id | `17b355e1-4e5a-4bac-8683-6d6a0825c657` |
| DB | `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db` |

**Chassis: read it from the pod, never from this line or the tag.**
```bash
kubectl -n ai-persona-system get pods -l app=agent-chassis \
  -o custom-columns='NAME:.metadata.name,IMAGE:.spec.containers[0].image,START:.status.startTime'
```
At writing: **`v1.0.1182`, deployed 2026-07-28 09:55Z.** It rolled **four times** on
07-27 (1173, 1174, 1175, 1180) — see landmine L14.

## 2. State

**The site reads properly now.** The palette was repaired on 07-27 and verified in
headless Chromium: **0 contrast failures and 0 broken images across 8 of 9 deployed
pages**, against ~101 failures that morning. Card titles went 1.21:1 → 13.19:1.

**`bugs_open/085` is FIXED and VERIFIED on both render paths** (build and scoped
re-render). Index carries exactly the one chart assigned to it.

**Live and working:** `evidence-chart` on the index (3 charts' worth of data, 1 rendered
— the other two belong to capabilities); the `palette_contrast` discovery check
(`features_open/026` Phase 2), proven end to end.

### In flight at handoff — check these first

| item | id | what to do |
|---|---|---|
| capabilities rebuild | `8f366ce5-cffc-49f6-8cb3-23957c8a9918` | fired 10:05Z as the spend-cap canary. See §3. |
| 2 × `needs_imagery` | `62b7918f-…`, `539893ae-…` | at `failed`, attempts remain. **Re-fire only once the canary proves the cap is lifted.** |

## 3. THE THING THAT BLOCKED EVERYTHING — and how to tell if it is back

On 2026-07-28 the Google project hit its **monthly spend cap**. Both modalities refused:

```
image: banana / gemini-3-pro-image-preview  -> 429 RESOURCE_EXHAUSTED
text:  provider=gemini / gemini-pro-latest  -> 429 "exceeded its monthly spending cap"
```

The owner raised it at ~10:00Z. **The cap is Google's, not ours** — grepped the tree for
any spend/budget/quota config of our own: none. `banana` is our name for the Gemini
image API (`DefaultBaseURL = https://generativelanguage.googleapis.com/v1beta`).
Owner-only lever: <https://ai.studio/spend>. Alternative: the banana client already
supports a Vertex base URL (`us-central1-aiplatform.googleapis.com/v1/projects/...`),
which bills through GCP budgets instead of a hard ceiling.

> **L15 — THE FLEET LOOKS HEALTHY WHILE IT CANNOT WRITE A SENTENCE.** During the cap,
> `page-rerender` completed **36 times**, plus the feed ingester, health checker and
> build triggers. All **non-LLM** paths. A status count showed a busy, green platform
> that could not generate a word or an image. **Never infer LLM health from
> orchestration status counts.** Detect it directly:
> ```sql
> SELECT count(*) FROM orchestration_states WHERE created_at > now() - interval '2 hours'
>   AND COALESCE(error, collected_data->>'__step_error','') ILIKE '%spending cap%';
> ```

> **L16 — a 429 naming one model is a statement about the PROJECT, not the model.** I
> read the imagery 429 as an imagery problem and told the owner so; the same cap was
> thirteen minutes from stopping every page build on the fleet. When a quota error names
> a billing scope, measure at that scope before reporting.

## 4. Next actions, in order

1. **Confirm the canary** (`8f366ce5`). If `complete`: the cap is lifted and the
   capabilities chart should be on the page. If it 429s again, stop and tell the owner —
   do not burn the remaining attempts.
2. **Verify the capabilities chart properly.** It must carry **two** charts
   (`news-pipeline-credibility`, `council-review-outcomes`) and **not three** — that is
   085's fix on the build path, which has never been exercised on this page:
   ```sql
   SELECT p.name, (SELECT string_agg(DISTINCT m[1],',') FROM regexp_matches(pc.rendered_html,'data-chart="([a-z-]+)"','g') m)
     FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id
     JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
    WHERE s.domain='fundamentallyai.com' AND cc.function='evidence-chart';
   ```
3. **Run the link crawl against the baseline.** A pre-build baseline of **9 internal
   targets** is at `scratchpad/cap_links_before.txt`; if that scratchpad is gone, the
   count is the number to beat. **This site's rebuilds have AUTHORED broken links twice**
   (6 on index, 10 on capabilities) and the gate shipped them as warnings both times.
   Capture `href="(/[^"]*)"` and strip the fragment **afterwards** — never `[^"#?]`.
4. **Re-fire the two `needs_imagery` items** for the calculator and selector hero images.
5. **`bugs_open/128`** — filed today. `image_url_404` flagged an image serving **200**
   and missed two that **404**. Its own header admits it never makes an HTTP request.
   Contains one hypothesis marked UNVERIFIED; read the code before acting on it.

## 4b. REGRESSION I CAUSED — the carousel images, and the rule the owner drew from it

**Owner reported 2026-07-28: the images no longer render in the carousel.** Confirmed,
and it was the capabilities rebuild I fired at 10:05Z.

| | before the rebuild | after |
|---|---|---|
| carousel `<img src>` | `/assets/images/hero-home.jpg`, `hero-fine-tuning.jpg`, `hero-review-council.jpg`, `brand-illustration.jpg` | `/assets/illustrations/review-council.svg`, `rapid-delivery.svg`, `verification-audit.svg`, `vector-search.svg` |
| do they resolve? | **all 200** | **404** (different directory: `/assets/illustrations/`, which has never existed) |

The writer **replaced four working image references with four invented ones**, in a
directory the site does not use, and the build passed the gate. Same mechanism as the 9
invented link targets the same rebuild authored (`bugs_open/071`, third instance) and
plausibly the same upstream cause — `bugs_open/092`, the writer never receiving its
constraints. Four `<img>` tags are still emitted, so nothing looks structurally broken;
the boxes are simply empty.

> ### OWNER RULE, 2026-07-28 — **replace before deleting**
> *"if that's because we're changing them then it should replace before deleting them."*
>
> A regeneration that swaps an asset reference must not leave the page pointing at
> something that does not exist yet. Either the new asset is created first and the
> reference swapped after, or the old reference stands. **An empty box is worse than a
> stale picture** — the stale one still communicates, and the empty one reads as
> brokenness to a visitor who has no idea a change was intended.
>
> This generalises past imagery: it is the same shape as `bugs_open/098` (archiving
> retracts a page from every derivation while the frozen artefact keeps serving) and as
> the link authoring above. **The platform is willing to write a reference to something
> that does not exist, in three separate subsystems.**

**Fix here is not a hand-repair of four `src` values** — that is what failed twice for
links (`071`) and would fail the same way at the next rebuild. Establish first whether
`/assets/illustrations/*` was ever a real convention on any site or is pure invention;
then it belongs with `092`/`071` as one authoring defect, not three patches.

## 4c. OWNER DESIGN DIRECTION, 2026-07-28 — panels as carousels, with a deliberate cliffhanger

Recorded verbatim in intent, not yet designed or built:

> *"as a whole almost all the panels could be carousels of one sort or another,
> especially on the home page, and have a really short first sentence and a small
> potentially incomplete second sentence to be completed when they click through — that
> may be one of the styles but not necessarily all of them."*

Reading it back for whoever picks this up:

- **Most panels become carousels**, home page first. Not necessarily the same carousel —
  *"one sort or another"*, and *"one of the styles but not necessarily all of them"*.
  So this is a **family** of panel treatments, not a single component applied everywhere.
- **The copy pattern is the interesting part**: a very short first sentence, then a
  second that is deliberately **incomplete** and resolves on click-through. The panel
  poses; the destination answers.
- **This is a content contract as much as a layout one.** It changes what the writer is
  asked for per panel — two sentences with a specific asymmetry — so it will need the
  writer prompt and the component `input_schema` to agree, not just CSS.

**Do not treat this as a ticket to start building.** The owner also said *"We may already
be addressing the design aspect"* — `features_open/018` (the screenshot design critic) is
the existing lane, and three carousel components already exist
(`hero-card-carousel`, `swipeable-insight-carousel`, `image-hover-card-grid`).
**Check what those already do before proposing anything new** — this workstream's
recurring mistake is building beside existing machinery rather than on it.

Two constraints already established that this direction must respect: the owner ruled on
2026-07-24 that **autoplay is opt-in, not default** (*"movement not for all carousels"*),
and every figure that appears must come from the evidence base, never from the writer.

## 5. Open, not blocked, and worth a thread

- **Template sameness** — `model-fine-tuning` and `multi-agent-review-council` still
  share a component pattern. This is `audit_finding_brief_fidelity` finding 2, still
  `detected`, and it is one concrete cause of the owner's *"not exciting or
  professional"*.
- **Imagery is repaired but thin** — finding 3, also still `detected`.
- **`features_open/018`** — the screenshot design critic. The taste layer. Worth doing
  **after** the mechanics, not before.
- **`features_open/026` Phase 3** — the rendered check on the deploy path. Phase 2
  (palette) is live; Phase 3 is what catches a component hard-coding an ink over a themed
  fill, which is the family that regressed the calculator page on 07-27.
- **3 × `deactivated_component`** — the site's `head`, `header` and `footer` chrome point
  at deactivated components. Renders fine; stale reference.

## 6. Landmines added this session (L1–L10 are in the 07-26 handoff)

- **L11 — `--color-primary` is used as BOTH a foreground and a fill.** Repairing a
  palette can therefore *break* a page: on 07-27 primary flipped near-black → light blue,
  fixing five pairings and breaking a sixth (white ink on the new light fill, 17:1 →
  2.32:1). **Run `scripts/render_audit.py` BEFORE and AFTER any palette change** — every
  "after" number improved and one page still regressed.
- **L12 — `styles.css` is written ONLY by a `webdesign-agent` run.** A Go fix to the CSS
  renderer sits live in the binary and never reaches the site. Regenerate by inserting a
  fresh `needs_design` work item (handler `webdesign-agent`). CSS is a linked asset, so
  the change is site-wide and instant — no page re-render needed.
- **L13 — probe pages by `pages.url`, never `pages.name`.** Blog, guide and tool pages
  live under `/blog/`, `/guides/`, `/tools/`. I built a probe list from names and briefly
  concluded four pages were 404.
- **L14 — an image tag observed NOW does not tell you what was running THEN.** The
  chassis rolled four times on 07-27. I compared two runs against a binary I had observed
  at a third time and nearly published a false contradiction. Establish the image for
  each run from its own timestamp.
- **L15, L16** — see §3.
- **L17 — `image_url_404` does not mean 404.** It is a DB registry check; its own header
  defers the HTTP half. Its silence is not evidence of working images (`bugs_open/128`).

## 7. Practices this session paid for

- **A green result from a check you just enabled proves nothing until you have seen it go
  red.** `palette_contrast` filed nothing on fundamentallyai (correct — repaired). That
  is indistinguishable from never running. The proof was running it on **dartsonline**,
  predicted at 1.11:1, which filed exactly that.
- **Deployment is not correctness.** A textbook pod-grep passed while the feature did not
  work, because the fix was not on the path the feature took.
- **Grep `/bugs_open/` before filing.** I filed a duplicate of `113`/`114` within forty
  minutes of the owner's report, because urgency suppresses the dedup check — and his
  report had gone to more than one session.
- **`ls a b` exits non-zero if EITHER glob misses**, so it is not a per-argument
  existence test. It told me a bug number was free when it was not.

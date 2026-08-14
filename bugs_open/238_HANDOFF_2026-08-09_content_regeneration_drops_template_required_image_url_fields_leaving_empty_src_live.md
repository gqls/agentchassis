# 238 — a content regeneration drops the template's required image-URL fields, shipping `src=""` to the live page

> **⚠ READ THE ADDENDUM AT THE FOOT BEFORE ACTING ON THIS FILE (2026-08-11).**
> **The §Mechanism below is WRONG about the cause** — the LLM never saw those
> keys and cannot emit them; they are resolver-sourced, and the loss is a
> `skip_field` default plus a wholesale replace. The damage was also **wider than
> five images** (11 `_url` keys; the five card links and the section CTA vanished
> silently because the template gates them). The **"cheapest real win"** in
> §Immediate state is **refuted** — `/case-studies.html` never referenced these
> images. The site is repaired and both fix halves are committed (inert until the
> next roll). Everything below is kept unedited as the original account.

**Status: OPEN** — fixes committed 2026-08-11 and **inert until the chassis image
next rolls**; the live symptom on finetuning.uk is gone (repaired by SQL 378/379,
verified at the served page). Four other rows across three sites remain in the
damaged state, deliberately (addendum §7).

~~**Live on finetuning.uk's homepage right now** — five
`<img class="csg-card-image" src="">` on `/index.html`, verified at the served
page 2026-08-09 ~16:10Z.~~ **Repaired 2026-08-11: 0 empty `src`, 5 images, 5 card
links, 1 CTA anchor at the served page.**

**I caused this run to happen** (see §Provenance) — the item was promoted and
dispatched by an improvement-loop firing I made for an unrelated reason. The
defect is in the handler, not in the firing, but the honest ordering is that the
homepage was intact at 14:47Z and is broken now.

## Mechanism

A `tone_shift` work item (`design-audit_tone_shift_index_…`, handler
**`page-build-handler`**, claimed by `build-dispatch-loop`, completed
**2026-08-09 15:18:10Z**) regenerated the `case-studies-grid` section data on
`/index.html`. The page's components were rewritten at 15:17:19Z.

The regenerated `content_data` contains, for all five cards:

- `card1_image_alt` … `card5_image_alt` — **present**
- `card1_image_url` … `card5_image_url` — **absent entirely**

The component template still requires them. `content_components` row
`3f946437-1dc7-4164-987d-620933589076` (`case-studies-grid`) emits, five times:

```html
<img class="csg-card-image" src="{{.card1_image_url}}" alt="{{.card1_image_alt}}" loading="lazy" />
```

A missing key renders empty (`missingkey=zero`), so the live page serves
`src=""` five times with the alt text fully intact — which is what makes it read
as a working image block in every check that looks at markup shape.

**The regeneration also rewrote the case studies themselves**, not just their
tone: card 1 went from "Cutting Quote Turnaround for a Facilities Management
Firm" to "Cutting the Manual Work Out of Quote Requests", and card 4 changed
subject from a private-AI deployment to "coordinating agent processes". So the
five images that *do* exist no longer map 1:1 onto the five cards — this cannot
be repaired by pasting the old URLs back.

## Why this is a class, not one site

The failure needs only two things, both common: a component whose template
sources an image from `content_data`, and any handler that regenerates that
section's data. The generator is not told that certain keys are structural
rather than editorial, so it reproduces the ones that look like copy
(`*_image_alt`) and drops the ones that look like plumbing (`*_image_url`).

Worth measuring before sizing a fix (NOT done here — stated as the open
question, not as a finding): how many content-driven components fleet-wide have
a `{{.*_image_url}}` in their template, and how many of their current
`content_data` rows are missing at least one of those keys.

## Evidence

```sql
-- the template requires the keys
SELECT substring(html_template from '<img class="csg-card-image"[^>]*') FROM content_components
WHERE id='3f946437-1dc7-4164-987d-620933589076';

-- the content no longer has them (returns 0)
SELECT count(*) FROM page_components pc JOIN pages p ON p.id=pc.page_id
WHERE p.site_id='1368e337-dd1d-4799-bbb3-8221a1b79bcc'
  AND p.url='/index.html' AND pc.slot_name='case-studies-grid'
  AND pc.content_data ? 'card1_image_url';

-- the item that did it
SELECT item_type, handler_agent, completed_at FROM site_work_items
WHERE site_id='1368e337-dd1d-4799-bbb3-8221a1b79bcc'
  AND item_key='design-audit_tone_shift_index_1368e337-dd1d-4799-bbb3-8221a1b79bcc';
```

At the artefact:
`curl -s https://finetuning.uk/index.html | grep -c 'csg-card-image" src=""'` → 5.

## Why no detector caught it

The site's own `empty_src` checker exists and is live (shipped 2026-08-03,
pod-verified). It has not fired because **site discovery has no recurring
driver** — the last discovery pass on this site ran at 13:51Z, roughly 90
minutes *before* the regression. That is `bugs_open/230`, and this is a concrete
instance of its cost: a defect introduced by the repair pipeline itself sits
unseen because nothing re-examines the site afterwards.

**The transferable shape:** a repair loop that regenerates content and a
discovery pass that only runs on demand means the loop can introduce exactly the
class of defect the loop exists to remove, and report success.

## Fix candidates, ordered by what closes the door

1. **Make structural keys non-regenerable.** Have the content generator receive
   the existing `content_data` and merge rather than replace, or mark
   `*_image_url` (and siblings) as preserved fields the LLM cannot emit or drop.
   Only this makes the bad state unrepresentable.
2. **Fail the write.** Validate regenerated section data against the template's
   required keys before persisting; refuse and raise a work item on a missing
   key. Cheaper, and turns a silent live regression into a queued item.
3. **Detect after the fact** — give discovery a recurring driver (`bugs_open/230`)
   so `empty_src` catches this within a cycle. Necessary regardless, but it is
   detection, not prevention.

## Immediate state of the affected site

- **The five case-study images now exist and serve** (`/assets/images/case-study-*.jpg`,
  all HTTP 200, `image/jpeg`, 52–94KB, distinct). That work is done and is not
  what is broken.
- **`/index.html`** — five empty `src`. Needs new content that carries image
  URLs matching the *rewritten* case studies. Not a paste-back.
- **`/case-studies.html`** — unaffected by the rewrite (its `case-studies-list`
  template **hardcodes** the five paths, unchanged since March). But its
  `rendered_html` dates from 2026-04-23 and predates the assets, so the page
  does not yet show them. **A rerender of that page alone should light up all
  five images with no content decisions required** — the cheapest real win here,
  and the one to do first.

## Provenance

Found while verifying `bugs_open/233`-adjacent work: the finetuning imagery
Phase 1 task (`docs/agent_docs/docs024_key_docs_latest/finetuning_uk_repair/PLAN_2026-08-04_imagery_then_visual_designer.md`).
The images were generated successfully; checking whether the *pages* referenced
them is what exposed this. Verifying at the asset URL alone would have reported
Phase 1 complete with the homepage newly broken.

---

# ADDENDUM 2026-08-11 — root cause CORRECTED, damage wider than filed, both halves fixed, site repaired

Worked by the `bugfix_238_regeneration_key_loss` lane (session "bugfix 238").
Standing five: `docs/agent_docs/docs024_key_docs_latest/bugfix_238_regeneration_key_loss/`.

## 1. ⚠ The mechanism above is WRONG, and the difference decided the fix

§Mechanism says the generator "reproduces the ones that look like copy
(`*_image_alt`) and drops the ones that look like plumbing (`*_image_url`)", and
§"Why this is a class" says "the generator is not told that certain keys are
structural rather than editorial".

**The LLM never saw those keys and cannot emit them.** `plan_sections` splits a
component's schema fields by `source` and puts only `source:"llm"` fields into
`llm_field_specs`; the writer prompt closes with *"Return a JSON object with
exactly the keys listed in 'What To Write'. Do not add any keys not in that
list."* Every lost key is **resolver-sourced**:

| field | declared source | required | on_missing |
|---|---|---|---|
| `card1..5_image_url` | `site_assets.image` | true | (absent → `skip_field`) |
| `card1..5_link_url` | `site_specs.case_studies.cardN_url` | true | (absent → `skip_field`) |
| `cta_link_url` | `site_specs.pages.contact_url` | true | (absent → `skip_field`) |

The real chain, four faults:

1. **The sources resolve nothing on this site.** `site_assets.image` aliases to
   role `hero` (`imageryplan.ImageRoleForPath`), and `r.assets` is populated only
   from `site_plan_imagery` joins — finetuning has 0 imagery rows and no current
   plan. And `SELECT aspect, count(*) FROM site_specs WHERE aspect IN
   ('case_studies','pages') GROUP BY aspect` returned **zero rows fleet-wide**:
   those six link fields have never resolved on any site, on any build, ever.
2. **A required field's absence is silent by default.** `on_missing` defaults to
   `skip_field` (`plan_sections_action.go:1844-1846`) and the required branch
   honours it, on the stated premise *"templates gate on the field"* — false for
   this template, whose `<img>` is at root scope with a bare `src=`.
3. **The render gate is exempt by construction.** `missingRequiredLLMFields`
   refuses to render a section missing a required field — but
   `if source != "llm" || !required { continue }` (`json_envelope.go:463-467`).
4. **`save_page_sections` replaces wholesale** (DELETE `:756` + INSERT `:904`),
   having just snapshotted the old row to `page_component_history` — i.e. it
   holds the values in hand at the moment it destroys them.

**Why it had never happened before, which the file could not have known:** the
values were byte-identical from 2026-05-01 to 2026-08-03 because every
intervening run was a *re-render*, and `rerender_page_sections` **merges**
`stored ⊕ fresh resolved_data` (`:498-506`). The first true regeneration killed
them. **The asymmetry between those two write paths is the bug.**

## 2. The damage is wider than five images

The dropped set is exactly **every key ending `_url`** — 11 of the template's 58
fields. Because the template guards the links (`{{if .cardN_link_url}}`,
`{{if .cta_link_url}}`), the five "Read case study" links and the section CTA
**vanished without leaving anything behind**: 0 `<a class="csg-card-link">` and 0
`<a class="csg-cta-btn">` anchors on the served page, while a grep for the bare
class names returned 4 and 3 (all CSS rules in the component's own `<style>`).

**A gated field fails more quietly than an ungated one** — the opposite of the
intuition that gating is the safe pattern, and worth carrying out of this bug.

## 3. The open question is answered

> "how many content-driven components fleet-wide have a `{{.*_image_url}}` in
> their template, and how many of their current `content_data` rows are missing
> at least one of those keys."

**10 active components** carry a `{{.*_image_url}}`; **5 deployed rows across 4
sites** are missing at least one — finetuning.uk `/index.html`,
ai-agent-orchestration.com `/index.html`, leopardessconsulting.co.uk `/blog.html`
(post1-6) and its automation-savings-estimator tool page, oufe.com's
recovery-waterfall tool page. Separately, **26 fields across 8 components**
declare a `site_assets.*` source with type `"url"` — invisible to all three
existing image checks, which key on `type ∈ (image, image_url)`.

## 4. ⚠ The "cheapest real win" in §Immediate state is REFUTED

> "**A rerender of that page alone should light up all five images with no
> content decisions required** — the cheapest real win here, and the one to do
> first."

`/case-studies.html` **never referenced these images.** Its `case-studies-list`
template has no `<img>`, no `.jpg` and no `/assets/images/`. The claim came from
`html_template LIKE '%case-study-%'`, which matched the CSS class
`case-study-item`. Caught by the finetuning lane on 2026-08-10 and recorded in
`WRONG_CALLS.md`; the `needs_page` rebuild ran anyway on 08-09 and the page still
shows 0 image paths, as it always would have.

## 5. What shipped

**Prevention — PBP-039, commit `d26c26a9a`, council `bd38df2e`.** `planSection`
now falls back to the page's own deployed `content_data` when a non-llm source
resolves nothing (order: literal → aliases → stored row → `on_missing`). Placed
at `handleMissingField`, the one closure both the build and re-render paths pass
through, so there is no third merge path to drift. Live resolution always wins;
llm fields are never carried; an empty stored value is never carried; the preload
is lazy. A required field resolving NOWHERE writes a durable
`STRUCTURAL_KEY_CARRY_MISS` row, because a plan that lost a key and a plan for a
component that never declared one are otherwise byte-identical. **Fix candidate 1,
as ordered in this file.**

**Detection — PBP-040, commit `51f56d0c9`, council `98852baa`.**
`RenderComponentAction` stops discarding the dead-URL report `missingBareFields`
already computes — it returned `[card1_image_url … card5_image_url]` at the very
render that shipped this bug — and, when armed, files a page+slot-keyed
`dead_url_control` item and refuses. Opt-in with the unsafe default OFF (owner
ruling 2026-08-02); the config half is HELD until the binary is confirmed.
Record-only on the re-render path. Two blind discovery predicates widened from
type to source. **Fix candidate 2, scoped to what template authority can prove.**

Both are **inert until the chassis image next rolls**.

**Fix candidate 3 (detection via `bugs_open/230`) remains correct and remains
unavailable**: the discovery rotations were switched off on 2026-08-10 on the
owner's cost instruction (`9a9070ab7`).

## 6. The site is repaired — SQL `378` / `379`, applied 2026-08-11

The 11 keys were restored as data and the page re-rendered with **no LLM**
(`reason: section_data_resolved`). Verified at the served page, not the row:

```
grep -c 'csg-card-image" src=""'          → 0   (was 5)
grep -c 'src="/assets/images/case-study-' → 5
grep -c '<a class="csg-card-link" href="' → 5   (was 0)
grep -c '<a class="csg-cta-btn" href="'   → 1   (was 0)
```
Both link targets serve 200; a fabricated sixth image filename 404s (the control).

**Not a paste-back, as this file correctly insisted.** The same run rewrote which
case studies the cards describe, so mapping is by **subject**, corroborated
independently by the regenerated alt texts (card2's "indexed documents connected
by thin linking lines" → legal-rag; card5's "secure enclosed network" →
private-ai), each asset used exactly once. Cards 2 and 4 were judgement calls and
went to the owner as such. The five historical card links **404 today**, so they
were re-pointed at `/case-studies.html` on the owner's decision rather than
restored. `378` also seeds the `case_studies` and `pages` spec aspects — the
first of either on any site — so six of the eleven fields now resolve live.

## 7. The other four damaged rows are NOT repaired, deliberately

- **ai-agent-orchestration.com `/index.html`** — a genuine regression (its history
  had the keys), but the historical URLs 404 and the site has no case-study
  assets, so restoring would trade an empty `src` for a 404 one. Needs imagery
  first. ~~This is also the honest acceptance case for PBP-039.~~
  > **⚠ CORRECTED 2026-08-11, by the council's `editquality` seat.** It is NOT.
  > **The carry reads the CURRENT deployed row, not `page_component_history`** —
  > and this row no longer holds the keys, so there is nothing to carry from. It
  > would produce a `STRUCTURAL_KEY_CARRY_MISS` and no repair. **PBP-039 is
  > prospective only: it protects a row that still HAS its keys and remediates
  > none of the rows already damaged**, including the one in this bug's title.
  > That is a real limitation and neither the submission nor this file said so.
  > The genuine acceptance population, measured the same day — deployed
  > `case-studies-grid` rows still holding `card1_image_url`, i.e. with something
  > to lose — is **3 rows**: aao `/enterprise-reference-deployment.html`,
  > finetuning.uk `/index.html` (repaired), leopardessconsulting.co.uk
  > `/who-we-help.html`.
- **leopardessconsulting.co.uk `/blog.html` + its tool page, oufe.com's tool
  page** — the "never had the key" class, not regressions; no candidate assets
  measured. The carry has nothing to carry for these; they get a
  `STRUCTURAL_KEY_CARRY_MISS` row and, once armed, a `dead_url_control` item.

## 8. Status

**OPEN.** The live symptom on finetuning.uk is gone and both mechanism halves are
committed — but they are Go, so they are **inert until the fleet next rolls**, and
the guard's config half is held behind that roll too. Per the standing bar a fix
that has not shipped leaves the defect reproducible, and per the owner's 08-06
ruling this file stays in `bugs_open/` regardless.


---

## 9. CONTRIBUTION 2026-08-12 (ai_site_selling_automation lane, not the 238 lane) — the fix HAS rolled, and a live case walked straight past it

Two things this file does not yet know. Neither is a re-diagnosis: the mechanism
claim in the second half is **filed for independent diagnosis, not asserted**
(`090` run `97ef39f0-19df-4935-834d-c80514fbc43e`), because my first hypothesis
about it was refuted within the hour.

### 9.1 §8 is STALE: the carry is live, checked at the artefact

§8 says both halves are "inert until the fleet next rolls". They have rolled.
`agent-chassis` is on `v1.0.1291`, whose image label
`org.opencontainers.image.revision` is `da5a7eb8f`, and
`git merge-base --is-ancestor d26c26a9a da5a7eb8f` passes — with controls in
both directions (the stamp is trivially its own ancestor; a commit made twenty
minutes before the check is correctly NOT an ancestor). The provenance log line
had already scrolled out of `--tail=3000` on that service, so the image label
is the instrument that still works hours later.

**This does not close the bug** — see 9.2 — but the reason it stays open has
changed, and "inert until the roll" now reads as a reason it cannot have been
tested, which is no longer true.

### 9.2 A live regeneration lost resolver-sourced keys ANYWAY, on 2026-08-12, with the carry aboard

**The case.** A `content_rewrite` (mode `edit_live`) across four pages of
webdesign.uk dropped `cta_url`, `primary_cta_url` and `secondary_cta_url` from
every `hero` and `call-to-action` component — **14 anchors over 7 components**.
Both templates gate the anchor on the URL rather than the label
(`{{if and .cta_text .cta_url}}`), so each button rendered as **nothing at
all**: no error, no truncated prose, healthy byte counts (bodies GREW), and a
clean `claimscan`. This is exactly §the-existing-landmine's "a gated field fails
more quietly than an ungated one", one component family further on.

**Why the carry did not fire — the part for the diagnosis loop to confirm or
refute.** These fields are declared in `content_components.input_schema` with
`source: "renderer"`. In `plan_sections_action.go`, `sourceResolver.resolve`
short-circuits that source —
`if source == "" || source == "llm" || source == "renderer" || source == "static" { return nil, true }`
— returning value nil and **found true**. A renderer-sourced field is therefore
never *missing*, so `handleMissingField` never runs, so `carryStored` never runs:
the carry guards fields that FAIL to resolve, and this class always "succeeds"
with nothing. The field loop's own renderer/static branch (~:2362) `continue`s
after writing only a declared `fallback`, and these declare none.

**Refuted on the way, recorded because it was the obvious reading:** I first
concluded the URL keys were simply **not declared** in the schema, so the carry's
loop never saw them. The schema query returned all four, declared, with
`source: renderer`. A field being outside the carry's reach and a field being
absent from the schema look identical from the symptom.

**Repair applied on the affected site, and its shape is evidence too.** Restoring
the URLs into `content_data` was **necessary and not sufficient**: a
`page_rerender` dispatched afterwards still rendered no buttons, which is the
observable that most directly contradicts "the stored value gets used". The
buttons only came back once the anchors were spliced into `rendered_html`
(`SQL_2026-08-12e`) and the deployed files patched (`gqls/vm-sites` `b538295`).
That hand patch re-arms `bugs_open/229` on those components by construction — the
next rebuild will overwrite it and file a divergence item, exactly as this
afternoon's rebuild did to the sibling lane's own hand repair of the same
buttons.

**What this suggests for the fix, without prejudging the diagnosis:** if the
carry is meant to be the backstop for "a key the live page already carries", the
`renderer`/`static` short-circuit is a hole in it that no amount of `on_missing`
handling can see, because those fields never reach the missing path at all.

- **filed by:** `ai_site_selling_automation` lane, 2026-08-12, during the £149
  copy migration of webdesign.uk (`87eebf7d5` … `1ee940968`)
- **full account:** `docs/agent_docs/docs024_key_docs_latest/ai_site_selling_automation/NOTES_ai_site_selling_automation.md`
  (2026-08-12 "later" entry) and `WRONG_CALLS.md` 2026-08-12 (why five green
  checks could not see it)

### 9.3 CORRECTION 2026-08-12, same day — the mechanism in 9.2 was REFUTED, and 9.2 mis-frames this bug

`090` run `97ef39f0-19df-4935-834d-c80514fbc43e`: **REFUTED**.

**Retract from 9.2:** the claim that `source: "renderer"` fields fall outside
`carryStored` because `sourceResolver.resolve` returns `(nil, true)` for them.
It is a plausible reading of the code and it is not established. Do not build a
fix on it.

**The run is not decisive either, and the reason is my error.** I repaired
`content_data` at 17:23 and fired the run at 17:39, so its citations are the
values I had just restored (`"cta_url": "/contact.html"`, sampled fresh). It
measured a repaired system and correctly found nothing missing. The evidence of
the loss had moved to `page_component_history`, and my symptom text pointed at
the live table. **Owed: a re-run against that history for the 16:37–17:23
window, whose symptom says the live rows were repaired at 17:23.**

**What 9.2 gets wrong about THIS FILE, which matters more than my hypothesis:**
the run's revised hypothesis is that **238 as tracked in the codebase is the
dead-URL-CONTROL defect** — a section that RENDERS while leaving a URL attribute
empty, recorded non-fatally by `recordDeadURLControls` / `emitSectionDeadControlItem`
on the rerender path — **not** a case of a renderer-sourced key being dropped
from `content_data`. Its cited tell is the code's own comment:
*"DeadURLSlots names sections that RENDERED, but with a URL attribute left
empty (bugs_open/238)"*. `next_scope`: `dead_url_guard.go`,
`emitSectionDeadControlItem`, `recordDeadURLControls`.

So 9.2 filed a real, measured incident under this number on a reading of the
number that the loop disputes. **The damage stands; its home may not.** Whoever
picks this up should decide whether the 08-12 CTA loss belongs here at all, or
under a new number — and 9.1 (the carry has rolled, §8 is stale) is unaffected
by any of this.

---

## §10 — 2026-08-14, from the bugfix_268 lane: the carry has been EXTENDED to renderer/static fields

The 08-12 CTA loss got its own number (`bugs_open/268`) and its answer: the
carry this file shipped never covered `renderer`/`static`-sourced fields,
because the field loop's early "resolved at render time, not now" branch
`continue`s before `handleMissingField`/`carryStored` — original design,
predating the carry. The 268 fix calls `carryStored()` inside that branch
(stored beats declared fallback; early continue preserved). Registered as an
extension on **PBP-039** (whose stale "INERT until roll" status is corrected
in the same edit — 9.1 here already said §8 was stale). No change to this
file's own arms; the 16:37–17:23 history re-run this file says is owed was
fired by the 268 lane (run correlation `38e53a03-…`).

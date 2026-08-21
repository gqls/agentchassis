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

---

## §11 — 2026-08-20, the 238 lane resumed: the detection half is ARMED at last, the RFC is filed, and the remediation population is MEASURED (and it is not what §7 or any plan assumed)

Worked by the original lane (`docs/agent_docs/docs024_key_docs_latest/bugfix_238_regeneration_key_loss/`),
resumed after eight days dormant. Ownership re-checked before starting: `scripts/who-owns.py 238`
names this lane; no open work item or diagnosis run targeted the bug.

### 11.1 §8 and the register were BOTH stale, and the staleness was load-bearing

§9.1 corrected §8 on 2026-08-12 ("the carry has rolled"). **The concept register was never
corrected**, and it is what council seats and later sessions read. PBP-040's status line still said
*"Go half built and committed 2026-08-11, INERT until the chassis image rolls; the config half is
HELD"* — both clauses false, for nine days.

The consequence was not just misinformation. **The config half stayed unarmed BECAUSE arming it
looked premature.** Measured 2026-08-19/20: ZERO active `agent_definitions` rows carried
`record_dead_url_controls` or `refuse_dead_url_controls`, and ZERO `dead_url_control` items existed
in all platform history. A filed bug's detection half sat switched off after its blocker had
cleared, because a snapshot outlived its truth. Corrected visibly in the register (strike-through +
date + cost), not quietly overwritten.

Provenance re-verified rather than assumed: image `v1.0.1317`, revision label `2d13d530d`,
`git merge-base --is-ancestor 51f56d0c9 2d13d530d` PASSES, with controls both ways (the stamp is
trivially its own ancestor; that day's HEAD is correctly NOT an ancestor, so the probe
discriminates). Not `strings` — see `bugs_open/249`.

### 11.2 ARMED: the record-only half (migration `504`, applied 2026-08-20)

`sql_for_agents/504_bugfix_238_arm_dead_url_record_on_rerender.sql` sets
`record_dead_url_controls: true` on `page-rerender`'s `rerender_sections` step. Applied; verify
block reported **1 of 1** `rerender_page_sections` steps armed fleet-wide; re-checked independently
after COMMIT; recorded in `schema_migrations`. Council `8a2aab7c`.

Three things worth carrying out of it:

- **The step path is NOT 380's.** `page-content-writer` nests its render steps inside a
  `process_sections_loop` sub_workflow; `page-rerender` does not. Copying 380's nested path finds
  nothing and reads as "no such step".
- **The key was UNDECLARED on `RerenderPageSectionsInputSpec`** — the same omission
  `refuse_dead_url_controls` had on `RenderComponentInputSpec` until 2026-08-19, and for the same
  reason: both are read through a helper taking `config`, so no literal `config["..."]` appears in
  either function body and a grep-based census cannot see them. Declared in `bb6600e48`, with a
  test whose mutation is exactly the omission. Arming an undeclared key here is warn-only
  (`CheckConfig`, not `StrictConfig` — read at the deciding arm before touching anything, because
  the same mistake on a StrictConfig spec took the fleet's page-publishing path down for 33 minutes
  on 2026-08-19).
- **The refusal half (`380_*_HOLD`) is still held, but on a NEW ground.** Its original hold
  condition (the binary carries the Go half) has been satisfied since 08-12. It is now held on the
  owner's 2026-08-20 sequencing decision: the refusal blocks the rebuild of a page whose data is
  still damaged, so it follows the drain. `504` carries a **negative control** asserting the
  refusal is armed nowhere, so a later edit cannot make that owner-facing decision silently.

### 11.3 FILED: `RFC_042`, discharging the seat instruction PBP-039 collected eighteen days ago

`architecture_review/RFC_042_content_data_has_nine_writers_two_write_disciplines_and_one_carried_funnel.md`.
The `architecture` seat returned `needs_rfc` on PBP-039 and said the asymmetry itself was the thing
an RFC would be about; it was "routed for a human call" and no file was ever written.

**The delay is itself evidence and the RFC opens with it:** in the interval the divergence was
re-reached — `bugs_open/268`, 214 CTA anchors, 19 sites, through a branch older than the carry.

Census in the RFC: **nine writers** of `page_components.content_data`, one carried funnel, most of
the rest REPLACE. Recorded with how it was obtained, because a grep for the SQL literal finds
**eight** and reads as complete — the admin handler builds its SET clause dynamically. Four options
costed; (d) "make the save merge" stays rejected but with one earlier ground corrected rather than
inherited ("merge of two versions of prose is ill-defined" is true of LLM fields, which nobody
proposed merging). Recommendation: contract-plus-`planSection`-completion now, the unified detector
next, the shared write seam only if the detector shows real non-funnel losses — **build the detector
before the guard, because the guard's population is currently an inference.** Flagged to be answered
jointly with `RFC_008`, the same question about the sibling column of the same table.

### 11.4 ⚠ MEASURED: the remediation population, and it refutes the plan it was meant to serve

The plan for this session was a route-and-classify remediation: send carry-misses whose source now
resolves to a no-LLM rebuild, and those with a recoverable history value to a validated restore.
**Both routes have ZERO population. Measured before building, which is the only reason the work was
not wasted.**

| measurement | result |
|---|---|
| open `required_fields_missing` items whose still-empty fields are resolver-sourced | **0** (29 are llm-sourced, 1 asset-sourced) |
| of the 25 field slots the 28 carry-miss findings name, how many have a declared `site_specs.*` source that resolves today | **0** — *not one aspect row exists* for any of them |
| how many hold a recoverable value in history, unambiguously | **11** (aao `/index` `case-studies-grid`; exactly 1 deployed component declares each field, so the field name attributes it) |
| ambiguous — history holds the field but several components on the page declare it | **1** (gamesdesign `/index` `system-stats` `cta_url`; **5** components declare `cta_url`, and the rows that hold it have `slot_name` NULL) |
| never held a value at all | **13** (fundamentallyai `platform-log-index` `post1..6_url` — whose current schema does not even declare them, the `bugs_open/309` class; `contact_email` ×6; `result_cta_primary_url`) |

**Why the router cannot help as designed: the detector never files these items.**
`check_required_fields_missing` deliberately skips `site_specs.*` / `pages.*` / `query.*` sources
("a separate change with its own census"), so the population the router would classify does not
exist as work items. **The gap is in the producer, not the router** — and widening the router first
would have been a change that ships and moves nothing (`WRONG_CALLS` 2026-07-25).

**What the population actually is:** almost entirely *"the declared source has never existed on this
site"*. The remedy for that is to populate the source or amend the schema — content and owner work,
not a rebuild. `378` seeding finetuning.uk's `case_studies`/`pages` aspects remains the only
instance of either aspect on any site.

So §3's open question, which §2 of the addendum answered by component count, now has its answer by
*field slot and by cause*, and the causes have opposite remedies. **Do not plan a restore path off
the aggregate.**

### 11.5 Two probe traps recorded, because I hit both, in opposite directions, in ten minutes

Both are in `WRONG_CALLS.md` 2026-08-20 in full. Short form, because anyone re-measuring this will
meet them:

- a history probe joined on `page_id` alone **over-counts by slot** — it credited `system-stats`
  with 103 `cta_url` values belonging to `tool-list`/`game-list`/`hero`, and I nearly filed that as
  a new regression;
- the obvious tightening (`slot_name` match + `source='artefact_archive_trigger'`) **under-counts by
  writer** and returned 0 for all 25 — including the page §7 correctly calls a genuine regression.
  `page_component_history` has five writers and only the trigger populates `slot_name`, so
  **selecting on provenance silently selects on schema completeness**: aao has 1,184 app-written
  rows with NULL slot (42 holding the value) against 211 trigger rows (0 holding it).

The working discriminator is **content identity** — how many deployed components on that page
declare the field — which is what produced the three-way split above.

### 11.6 Status

**STILL OPEN**, and the reason has changed again:

- prevention: LIVE (carry + the 268 renderer/static extension), verified in the running binary;
- detection, ungated class: **ARMED** on the re-render path as of today; refusal sequenced behind
  the drain;
- detection, **gated class**: still only the carry-miss findings, which still have no automated
  consumer. This is the honest residual and it is now precisely bounded: the producer skips the
  source families, and the discovery rotations that would drive a widened producer are paused on
  the owner's cost instruction (`bugs_open/230`);
- remediation: the population is measured and parked-or-parkable, but **10 (page, slot) pairs are
  not yet visible as work items at all** — the backfill is the next concrete step, and it must land
  with a park route for resolver-sourced fields or those items will misroute to the prose writer,
  which structurally cannot fill them (proven by the `410` `asset_sourced` canary refusal).

### 11.7 The `090` came back UNVERIFIABLE — and chasing what it asked for produced the behavioural proof PBP-039's "verify-later" has been owing since 2026-08-11

Run `68b3f9b6-1674-41a0-bc9e-c251192daaa1` (intake `5477195a`), fired on the two carry gaps §11.8
names. Verdict: **UNVERIFIABLE, stopped at iteration-cap** — not confirmed, not refuted. Its own
account of why is worth keeping, because two of the four gaps are tooling and two are substantive:

- it could not read `planSection`'s field loop, `storedFieldValue` or `carryStored` at all — the
  symbol search for `handlemissingfield` returned 0 rows and `carrystored` matched only
  `carryStoredSection` in the *rerender* file. Per its own rule 10(b) that is UNKNOWN, not absent,
  and it said so rather than concluding;
- **substantively**, it observed that every `plan_sections` record it could find says *"no
  previously-built row held a value, so there was nothing to carry"* — i.e. in every observed
  occurrence there was no stored value to destroy, **which is the opposite of the hypothesised
  scenario.** That independently corroborates §11.4's census from the work-item side.

**What it asked for to settle it:** *"a state/runtime instance where a field HAD a non-empty
previously-stored `content_data` value and a subsequent run overwrote it with an empty value without
a carry — e.g. `page_component_history` rows for the same page/slot across two consecutive builds
showing a populated field becoming blank."* That is a query, so it was run.

**The result, and it is the useful output of this whole thread:**

| measurement | result |
|---|---|
| consecutive archived generations of the same (page, slot) where ANY field went non-empty → blank | 348 events / 127 slots / 58 fields |
| …restricted to **non-LLM-sourced** fields (an LLM field changing is the writer working) | **66 events**, 11 sites — **all `renderer` (48) or `static` (18)** |
| …the window those 66 occupy | **2026-08-11 → 2026-08-14 18:36 UTC**, and nothing since |
| …events with a `site_specs.*` / `site_assets.*` / `query.*` source, at any time | **0** |
| **demand control** — archived generation-pairs since the last loss event | **3,033**, 2026-08-14 → 2026-08-20 |

**`renderer`/`static` is exactly the class `bugs_open/268` fixed** (`8f899cc8d`, committed
2026-08-14 09:13 BST). The losses stop on the day it landed and have not recurred across 3,033
subsequent archived generations. **That is the behavioural acceptance test PBP-039's `verify-later`
asked for and nobody ever ran** — "dispatch a regeneration at a page whose component declares a
non-llm URL field, and assert the persisted `content_data` still holds that key afterwards" — except
it is better than the single dispatch it specifies, because it is the whole fleet's ordinary traffic
rather than one induced case.

⚠ **State the window, because it bounds the claim.** The `artefact_archive_trigger` archive begins
**2026-08-09** (migration 357), so this can see eleven days, not the bug's whole life. There are two
clear days inside the window before the first loss event, so the window is not the only reason the
series starts on 08-11 — but a longer history would be a stronger claim, and this is not one.
The `save_page_sections_overwrite` rows reach back to 2026-03-16 and **cannot** be used the same
way: they carry no `slot_name` (§11.5), so consecutive generations of one slot cannot be paired.

### 11.8 The two carry gaps: REAL IN THE CODE, UNOBSERVED IN PRODUCTION — and therefore NOT shipped

Both were read directly in `plan_sections_action.go` and both are reachable:

1. **A blank resolved value beats a good stored one.** The generic branch stores whenever
   `found && value != nil`; `resolveSpecPath` / `resolveSpecAlias` / the `site_assets` lookup can
   each return a present-but-empty string, which is non-nil, so it lands in `resolvedData` and
   `carryStored` never runs. `storedFieldValue` refuses empties on the *stored* side; nothing
   refuses them on the *resolved* side.
2. **A `query.*` resolver ERROR drops the key.** That branch deliberately does not route into
   `on_missing` (`bugs_open/054`: an error must not be masked as "no data"), applies a declared
   fallback if there is one, and otherwise `continue`s with the field **absent from
   `resolvedData`** — so the wholesale save drops it, while the re-render merge would have kept it.
   One unregistered query name in a schema destroys that key on every regeneration, for ever, with
   a `Warn` as the only trace.

**Neither has a single observed instance.** 0 loss events for those source families across the whole
archive window; the fleet-wide census found only **2** empty-string spec values behind declared
sources; and the `090` reached the same conclusion from the work-item side.

**So they are recorded, not shipped.** Writing 3 lines of Go and a mutation-proved test would have
been easy and would have looked like diligence — but it would be a fix sized from a code reading
rather than from evidence, and this estate has a name for that (`WRONG_CALLS` 2026-07-25: counting
the population a fix is ABOUT instead of the population it would ACT on). They belong to `RFC_042`
option (e) — "declare `planSection` the sole complete producer" — as its concrete content, with the
honest label: **reachable by reading, unobserved in production.**

**What would justify shipping them**, stated so the next session need not re-derive it: one loss
event with a `site_specs.*`/`site_assets.*`/`query.*` source in the pairing query above, or a
`STRUCTURAL_KEY_CARRY_MISS` row whose page/slot DID hold the value in the prior generation. Re-run
the query; it is in the lane RUNBOOK.

### 11.9 — 2026-08-20 later: council APPROVED at round 2, the declaration shipped on v1.0.1319, and the emit still has not fired (because nothing has run)

**Council `8a2aab7c` — APPROVED at round 2** (round 1 REVISE, verdict read before claiming it).
The round-2 advisory objections and what was done about each:

- **`editquality` (medium) — a `grounded_in` line contradicted the plan's own arming claim.** It
  was right, and the cause is worth naming because it is a generic resubmission hazard: the line
  read *"0 rows … the guard is armed nowhere"*, which was TRUE at round 1 and was made false **by
  migration 504 applying between the rounds**. Round-2 evidence carried forward unchanged becomes
  a claim about the present. Corrected visibly in the submission rather than deleted, so the trail
  shows both states.
- **`tooling_provenance` (medium) — does `subject_type='decision'` / `categories ? 'decision'`
  collide with `decision_guard.go`?** Checked rather than argued. **It does not, and the reason is
  that someone already split the vocabulary for exactly this:** the ENFORCEABLE tag is
  `'decision-record'`, separated from bare `'decision'` on 2026-08-10 *because three pre-existing
  rows already used `'decision'` to mean "a note ABOUT a decision" — prose, not an enforceable
  record*. The row written here carries `'decision'` and **not** `'decision-record'`, has no
  `site_id` and no fences, so it is invisible to enforcement by design and matches the prose
  meaning exactly. The seat's "missing" item is also answered: `doc_notes_subject_type_check`
  permits `'decision'` (one of eight allowed values).
- **`guardian` (medium) — arming a new work-item emitter into a pipeline that is itself paused and
  backlogged, with no handler and no drain path.** Accurate, and stated in the plan rather than
  discovered: `dead_url_control` items are born `needs_human_review` with **no handler, by design**
  — nothing automated can invent a missing image or destination. The drain path is a human, and
  the volume bound is `insertWorkItem`'s dedup plus the two-strike label. It is a real gap in the
  sense that items will accumulate unattended; it is not a new one — see §11.4, where the
  *existing* `image_url_404:empty-src` item for aao `/index.html` has sat at `detected` unworked
  since before this change.

**The declaration shipped.** `bb6600e48` is an ancestor of **v1.0.1319** (revision `447f3a8a8`,
controls both ways). So the interim window §11.2 flagged — where the config report named a live,
working setting as unrecognised, whose stated remedy is to delete it — **is now closed**. That was
the one cosmetic cost of arming config ahead of a roll, and it lasted one build.

**⚠ The emit has still never fired, and the reason is measured, not assumed.** Since arming:
**0 `dead_url_control` items, 0 `page_rerender` items, 0 archived generations.** The fleet has not
re-rendered anything at all. That is the "sustained zero has two readings" case in its benign
form — the archive count independently establishes "no traffic" rather than leaving it ambiguous —
but it means the arming is verified **at the config and at the code path, not at the artefact**.

The code-path verification, done because this exact file has been bitten by it before (council
round 1 on migration `473` objected that the action might read `ExecutionContext.Config` while the
migration wrote step-level config — *"two distinct maps"*):
`recordDeadURLControls(params.StepConfig.Config)` reads the **same map** as
`shouldStripLiteralMarkdown(params.StepConfig.Config, reason)`, and `strip_literal_markdown` is
already `true` at the identical path — visible in 504's own BEFORE output, shipped by `473`, and
documented live. **The key is in the map the code reads, proven by a sibling flag already working
there**, with no production dispatch.

**The one experiment that would close it is named and NOT run** (§11.10 in the lane NOTES): a
379-shape `page_rerender` at aao `/index.html`. It is an outward-facing action on a live customer
site, and `bugs_open/229` says a rebuild silently discards hand-patched `rendered_html` with no
divergence warning. The merge cannot lose a KEY — that is structural — but that is not the same
claim as "cannot overwrite hand-edited markup". **Owner's call.** Check first whether ordinary
traffic has already done it: `SELECT count(*), max(created_at) FROM site_work_items WHERE
item_type='dead_url_control';`

### 11.10 — 2026-08-20, owner-authorised: the demand control RAN, the emit FIRED, and its first output disproved the justification I gave the council

**Owner authorisation** (2026-08-20): *"Re run ai-agent-orchestration home page if you like because
there shouldn't be any hand patches so it is ok to overwrite them."* The hand-patch check was run
anyway, because it is a twenty-second read and it either confirms the expectation or surfaces
something worth knowing: **all 8 deployed sections stamp machine-made** (`rendered_html_digest =
md5(rendered_html)`). Expectation confirmed. Dispatch pre-checks also clear: 6 open items on the
page, all parked at `needs_human_review`, none dispatchable; chassis pods 70 minutes old (past the
~300s post-restart window in which a spawn is silently dropped).

**THE EMIT FIRED. This is the first `dead_url_control` item in the platform's history** (the count
was 0 for all time before it):

```
key      dead_url_control:index:case-studies-grid:card1_image_url,card2_image_url,card3_image_url,card4_image_url,card5_image_url
status   needs_human_review     handler  (none — by design)
refused  false                  (record-only on the re-render path, as designed)
summary  Dead URL control on index/case-studies-grid: no destination for card1..card5_image_url — recorded (the section still rendered)
```

So PBP-040 is now verified **at the artefact**, not merely at the config and the code path. Item
`page_rerender:238-demand-control:index` reached `complete` on its first attempt.

**The page was not harmed, and the one change is instructive.** Served bytes 64,139 → 64,184;
empty `src` **5 → 5**; `csg-card-link` anchors **0 → 0** — exactly as predicted, because the
declared sources still resolve to nothing (§11.4). The entire 45-byte delta is **another lane's
Open Graph improvement** arriving for free: two empty `og:title`/`og:description` tags and a
duplicate set were replaced by one correct per-page set (SEO-005). A live instance of *"a stale page
holds every improvement since it rendered"* — this page had not been re-rendered since before that
mechanism landed.

**⚠ AND THE ITEM DISPROVED MY OWN COUNCIL ARGUMENT.** It names **five** fields — the ungated `src=`
ones — and **not one** of the six gated `*_link_url` fields. At round 2 I had told the
`reuse_agent` seat that *"the vanished-ANCHOR class has NO other detector on this estate — that is
the non-overlapping half and it is the larger one"*. Wrong twice:

- **`href=""` is already covered** by `empty_internal_href` (`validate_page_content.go:909`,
  `discovery_checks/check_phantom_internal_links.go`; `datahelpers/links.go:133` states the division
  explicitly). I grepped the one file the seat named and generalised to the estate. The live proof
  is on this very component: `empty_internal_href:page_component:index:case-studies-grid:`, dated
  **2026-07-24**, parked ever since.
- **The gated class is not covered by `dead_url_control` either** — and I had already written that
  down correctly in the migration header, in §11.2 and in the `doc_notes` row, before offering the
  same class to the council as the change's unique contribution.

**What actually justifies the arming** (three reasons, all measured, recorded so the file does not
rest on the false one): it names the **schema field**, which an HTML-scanning detector structurally
cannot — this item says `card3_image_url` where `image_url_404` can only say "an img has no src";
it fires **synchronously at render** while the others are async discovery checks whose rotations are
paused on cost; and its key carries **page+slot**, where `image_url_404:empty-src` is site-wide and
one `blocked` row jams a site. Correction filed in `WRONG_CALLS.md` and appended to the `doc_notes`
decision row (2 rows now under that `subject_key`, the second being the correction).

**Consequence for the detector picture, corrected:** this class now has **three** producers on the
ungated symptom (`image_url_404` for `<img>`, `empty_internal_href` for `<a>`, `dead_url_control`
for both, field-named and synchronous) and still **none** for the gated/vanished class. That is the
honest residual and it is unchanged by today's work — with the added point, which the
`reuse_agent` seat's instinct was right about even though my answer was wrong, that a fourth
producer in this space would need consolidating rather than adding.

### 11.11 — 2026-08-21: the residual is NOT nine content decisions. Six of nine are one aspect-spelling mismatch, and the value already exists

Re-verified on **v1.0.1321** (revision `0483e7f4e`): all four fix commits ancestors with controls
both ways; record arm still `true` on `page-rerender.rerender_sections`; refusal still armed
**nowhere**; **1** `dead_url_control` item (yesterday's, unchanged); **28** carry-miss findings,
newest still 08-17 — so nothing new has been damaged.

Population re-measured, comparable with §11.4: **UNGATED (ships an empty attribute) 5 field slots,
1 page, 1 site — unchanged.** Gated 477 → **506** across 156 pages, which is other lanes building
new pages whose gated fields have no resolvable source, not new regression. ⚠ Treat the gated
number as a container count, not damage: a gated field absent is a template degrading *as designed*.

**Of the 10 damaged (page, slot) pairs, exactly 1 is visible as a work item** — aao `/index`
`case-studies-grid`, by yesterday's `dead_url_control` plus a pre-existing
`required_fields_missing`. **Nine exist only as `agent_error_log` findings that nothing reads.**

**And the nine are not nine problems.** Grouped by what they actually need:

| what they ask for | pairs | state of the fact |
|---|---|---|
| `site_specs.contact.email` | **6** (leopardess ×4, robot-hands, gamesdesign) | **the value EXISTS** — see below |
| `site_specs.contact.cta_url` / `site_specs.cta.primary_url` | 2 | no `contact`/`cta` value; gamesdesign has no email at all |
| 6 fields the component's schema no longer declares | 1 (fundamentallyai `platform-log-index`) | `bugs_open/309`'s class — a schema question, not a data one |

**⚠ THE FINDING: the sites-row alias is hard-gated to ONE aspect, so the resolver refuses to read a
value it is sitting on.** `resolveSpecAlias` step 2 (`plan_sections_action.go`):

```go
// 2. The canonical sites row.
if aspect != "identity" { return nil, false }
```

So `site_specs.identity.email` resolves from the `sites.email` column, and
`site_specs.contact.email` — **the same fact, different spelling** — returns not-found. Measured:
exactly **one** active component declares `contact.email` (`contact-block`) and exactly **one**
declares `identity.email`. The first resolves nowhere on any site; the second works. And
`sites.email` is populated for leopardess, robot-hands and fundamentallyai.

**So §11.4's conclusion needs narrowing, in my own words rather than someone else's later.** It said
the class is *"the declared source has never existed on this site"*. True of the **declared path**;
**false of the fact** for these six. The value is one column away and the resolver will not cross an
aspect name to reach it. Blast radius of the mismatch is contained — `contact_email` on
`contact-block`: **6 deployed rows, 3 sites, all 6 missing it** — which is exactly the six pairs.

**⚠ BUT DO NOT JUST WIRE IT UP.** Every one of those addresses is
`<site>@contactforsales.com`, and that pattern holds on **15 of 44 live sites** (26 empty, 3 other
domains). That is systematic, which makes it look like a platform-assigned lead-capture or parking
address rather than a client's inbox. Publishing it because it happens to be reachable is
`bugs_open/140`'s exact defect — a contact-info component serving an address nobody chose to
publish. **Whether that address should appear on a customer's page is an owner question, not an
engineering one**, and it is the single decision that settles 6 of the 9.

**Two candidate fixes, once that is answered, and the cheap one is also the better one:** point
`contact-block`'s field at `site_specs.identity.email` — one component row, no Go, no roll,
reversible, and it aligns with the sibling spelling that already works — rather than widening the
alias, which is a shared-resolver seam change and would make every aspect able to read the sites
row (bigger blast radius, council/RFC-shaped, and it would take the placeholder question fleet-wide
in one step).

**Nothing was minted into the review queue for these nine, deliberately.** Filing six items that a
single decision would close is the churn the two-strike rule exists to punish, and the estate has a
standing warning about a queue that drains so nicely it looks like the whole thing worked. They are
one decision away from being either fixed or correctly-parked; the queue is the right home for them
*after* the answer, not before.

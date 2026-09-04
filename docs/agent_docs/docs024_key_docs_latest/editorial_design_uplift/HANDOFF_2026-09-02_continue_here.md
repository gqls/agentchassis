# HANDOFF 2026-09-02 — editorial_design_uplift, continue here

**Supersedes** the joint cold-start pointer (`news_editorial_features/HANDOFF_2026-08-25_continue_here.md`)
**for the design half only**. The 035 composition work is still described there and in
`features_open/035_FEATURE_component_hierarchy.md`; this file is the design/imagery half plus the
state of everything this lane touched on 2026-09-02.

**Branch:** `087_towards_multiple_domains`. **Read `NOTES_editorial_design_uplift.md` from the
2026-09-02 entries down** — this file is the summary, the NOTES are the evidence.

---

## 0. STOP — three things about the environment before you act

1. ~~**The kubeconfig token EXPIRED at ~22:08 BST 2026-09-02.**~~ **RESOLVED — the owner refreshed it the same evening, and every §0/§7 item below was then verified; results in §9.** The original text is kept because the expiry produced the session's cleanest demonstration of why a probe needs a must-be-present control (see §9).
   ~~ Every `kubectl` call returns
   `You must be logged in to the server (Unauthorized)` — `get`, `exec`, psql-through-exec, all of
   it. **The owner refreshes it.** Nothing in this handoff can be re-verified at the cluster until
   he does. Every figure below carries the date it was measured, all before expiry.
2. **A fresh chassis rolled at 2026-09-02 20:56/20:57Z — `v1.0.1355`.** Per CLAUDE.md you owe a
   **migration dry-run after every roll**: `./scripts/migration/run-migrations.sh` (no args,
   read-only). ⚠ It takes **over five minutes** and, piped through `tail`, prints NOTHING until it
   finishes — which reads exactly like a hang. Run it unpiped, in background, with a generous
   timeout. Last run (pre-roll, 2026-09-02): **`Pending (164)`**.
3. **No orchestration dispatch within ~300s of a chassis restart** — long past now, but relevant if
   another roll lands while you work.

---

## 1. What this lane is, in one paragraph

The design/imagery half of the editorial work: how editorial pages LOOK, and whether they carry
imagery at all. Its companion is `news_editorial_features` (the content half), and the two share the
035 composition feature. Phase A (contrast/de-branding, migration 496) is done. The 2026-09-02 work
was entirely about **imagery density** — why editorial pages serve no pictures — triggered by the
owner's review of boxingonline.com, the first paid customer site.

---

## 2. THE HEADLINE: migration 686 was built, approved, applied, and ROLLED BACK. Read this before touching imagery.

**What it was.** `article-body` — the component every article page is built from — has one field
(`content`) and a template whose only interpolation is `{{.content}}`: no `<img>`, no `<figure>`, no
`background-image`. It cannot display an image by construction. Meanwhile six generated, deployed,
HTTP-200 `content_hero` images on boxingonline rendered nowhere. 686 gave it an optional image field
plus guarded template markup.

**It passed everything.** A 14/14 local harness with two discriminating controls, a rehearsed
apply-then-reverse round-trip, and **two council rounds** — round 1 REVISE (8 approve / 3 object),
round 2 **APPROVED** (`4bf6c48f-9cd6-440f-9257-a5668b6635fc`), eleven seats.

**It was wrong.** `[MEASURED 2026-09-02]` **292 of the 301 pages carrying `article-body`, across 31
sites, ALSO carry a `hero` component whose `background_image` reads the same `site_assets.hero`
key.** The new field would have rendered **the same image twice on 97% of the population**. The six
boxingonline pages that motivated it are in the **nine-page minority with no hero component at all**.

**Applied 13:56Z, rolled back 15:05Z. Nothing was ever rendered with it** — 0 of 301 instances
acquired the field. That is a short window plus luck, not a control anyone had built.

**What caught it:** not the harness, not the council. A peer lane's *unrelated* bug report about a
page showing one image twice. I recognised the shape and went to count.

**State now:** template at its exact pre-686 md5 `002cbcd9cada6a37bf4a5158fd1e5f22`, len 1378. The
file `docs/agent_docs/sql_for_agents/686_article_body_hero_image_capability.sql` carries a
**DO-NOT-APPLY header**. Its `schema_migrations` row is **deliberately left in place** so `--apply`
cannot replay it, with `notes` recording the rollback and the live md5. **Any superseding fix takes
a NEW migration number** (forward-only).

> **⚠ DO NOT re-propose a component-level image field for `article-body`.** Read §3 first. If you
> conclude it is right anyway, you must first explain the 292.

---

## 3. THE REAL FINDING — and it is a QUESTION, deliberately, not a diagnosis

Healthy pages were never missing a capability. They show imagery through the **`hero` component fed
by a page-scope `site_plan_imagery` row**. Verified at the artefact 2026-09-02:
`agritec.uk/blog/insect-bioconversion.html` serves
`background-image: linear-gradient(…), url('/assets/images/hero-bsf.jpg')`, has exactly **1**
page-scope plan hero row, and its only `<img>` is the logo.

Three lanes converged on the same place from three directions on 2026-09-02:

| route | measurement |
|---|---|
| this lane, orphaned heroes | **189 pages / 21 sites** hold an ACTIVE `content_hero` asset with no component that could render it |
| this lane, prose grain | across **462** `blog-post`+`guide` pages: **max prose sections on any page = 1; pages with more than one = 0** |
| `inline_guide_imagery`, selection | illustration-capable sections ARE selected since migration 644 — but on `landing` pages, one per page, and **zero** on `blog-post`/`guide` |

**The one sentence they compose into:** *the estate's article-shaped output is one prose slab plus
chrome, everywhere, regardless of what is in the component menu — and the menu demonstrably works,
because the same planner reaches for an illustrated section on landing pages.*

**So the open question is:** *why does the planner compose a blog/guide page with no hero when 330 of
432 peers have one, and never compose an article out of parts?* **Nobody has read the planner's
composition logic.** State it as a question. `ContentHeroKey`'s own doc comment says it exists for
*"a page the planner gave no hero of its own"* — so the system detects the gap, generates an image
to fill it, and has nowhere to put the result.

**Where it lives:** `bugs_open/114_HANDOFF_2026-07-27_generated_imagery_is_deployed_and_never_referenced.md`
— **OWNED and active; contribute into it, do not file a competing bug.** This lane's CONTRIB is
§4–5 of that file (the 189 census, the DO-NOT-repeat warning, the prose-grain addendum).

---

## 4. ~~ANSWERED on 2026-09-02~~ **REFUTED 2026-09-04 — READ THIS BEFORE QUOTING ANY OF §4**

> **⚠ THE ANSWER BELOW IS WRONG, AND IT WAS ALREADY WRONG WHEN IT WAS WRITTEN.** §4 says the planner
> produces no infographics because the prompt tells it *"Use sparingly in v1 — most plans will have
> zero section-scope entries."* `[MEASURED 2026-09-04]` that sentence is **not in the live prompt**:
> migration **718** replaced it with *"Content-carrying imagery is EXPECTED here, not exceptional"* on
> **2026-09-02 — the same day this §4 was written**, and `infographic` now occurs **8** times, not 3.
> The behaviour did not follow the instruction either way: **since 718 there have been 111 planned
> imagery entries, 12 illustrations and 0 infographics**, so the prompt was never the binding
> constraint.
>
> **The observation survives; the cause does not.** The estate really does hold **1** infographic in
> all history. Two live candidates, both measured 2026-09-04 and neither confirmed: **(A)** rule 13 is
> a DISJUNCTION (`illustration … or … infographic`, illustration named first) and it has won 12–0
> since 718; **(B)** of the 7 sites that planned imagery since 718, only **2** hold an
> `evidence_base` — so on 5 of 7 there is nothing for a figure to be *about*, and the zero may not be
> a defect at all. Full account and the free test that separates them:
> `framework_prompts_positive_voice/CONTRIB_2026-09-04_…_my_section2_was_ALREADY_WRONG_…` (`c44f2b613`).
> Also dead: the "no infographic-typed component" theory — `plan_sections_action.go:563` takes all
> three kinds in one query and never branches on kind.
>
> **The lesson, because it cost an owner decision:** a verbatim quotation is a measurement of a
> mutable string and decays exactly like a count. I re-measured every number that evening and quoted
> this one from my own handoff. `WRONG_CALLS.md` 2026-09-04.

## 4. (original, retained) ANSWERED on 2026-09-02: "what would ever WRITE an infographic row?"

This lane asked it on 08-31 and could not answer it. **The planner is obeying an instruction.**
Verified by me at the live prompt (`agent_definitions` id `f263eaa1-61e1-446e-9410-648e12b7875b`,
34,781-byte config, read 2026-09-02 before the token expired):

- **The vocabulary is COMPLETE** — `kind` is *"one of: `logo`, `hero`, `illustration`, `icon`,
  `infographic`. No other values permitted"*, repeated by rule 15. Nothing needed teaching.
- **Section-scope imagery is told to be rare**, verbatim: *"Use sparingly in v1 — **most plans will
  have zero section-scope entries.**"*
- **The stated minimum is chrome only**: one site `logo`, one `hero` on index, one `hero` per page
  with a hero-class component. **No floor for illustration or infographic anywhere.**
- **`infographic` appears exactly 3 times in the whole config, all three in rule/schema text, and
  NEVER in the worked example** — while the other four kinds do appear there.

**So the census is the prompt working as written:** hero 399 / icon 211 / logo 50 / illustration 25 /
**infographic 1**, fleet-wide, ever `[MEASURED 2026-09-02]`. *"Most plans will have zero"* produced
almost exactly zero.

**NOT TOUCHED, deliberately** — the prompt is read by the build path for every new site, the cost is
real generated images per section, and 18 site remakes are queued behind it. **Planner owners' call.**
If it is edited, rule 16's *"each entry produces exactly ONE image"* discipline must ride in the same
edit (under-decomposition produces unusable multi-panel images).

> **⚠ TWO ASKS, NOT ONE — keep them apart in any owner report.** A prompt change lands pictures where
> there is structure to hold them. On article pages there is none (§3). It would improve **landing
> pages**; it would put **zero** images inside article text. Do not report progress on the first as
> progress on the second.

---

## 5. State of everything this lane touched

| thing | state |
|---|---|
| migration 686 | **rolled back**, file marked DO-NOT-APPLY, ledger row kept on purpose |
| `article-body` | unchanged, md5 `002cbcd9…`, len 1378 |
| `news-listing` | **same defect, no change written** — deliberately out of scope until the §3 question is settled |
| `bugs_open/114` | this lane's CONTRIB + addendum committed; bug OWNED by `bugfix_114_imagery_wiring` |
| IMG-077 `check_unrendered_page_imagery` | built by the 114 lane from this lane's 189 census. Needed *roll + migration 708*. **Roll happened; 708 NOT applied as of 22:07Z** — so still inert |
| 035 P1 (composition) | separate track. Last commit `2a0bdb001` (08-31). **Whether it is in `v1.0.1355` is UNVERIFIED** — the binary probe was blocked by token expiry |
| harness | `harness/articlehero/` (14/14, durable) and `harness/composewalk/` (8/8, 035 P0) |

---

## 6. Traps this lane paid for on 2026-09-02 — all in `WRONG_CALLS.md` / `LANDMINES.md`

1. **A remedy is fitted to a POPULATION, not to a defect.** I measured the component, the assets,
   all three resolver arms, the llm-dispatch mechanism, the alias map and the precedents — **every
   layer except the neighbours.** Before adding a field to a shared component, ask what the OTHER
   instances already do. One query.
2. **A council cannot catch this class.** Eleven seats reviewed 686 and approved it. A submission can
   be entirely accurate about the change and entirely silent about the estate around it, and the
   estate is the one input reviewers cannot fetch. **State what the healthy instances already do.**
3. **The unit of the query must be the unit of the claim.** `blog-post` averages 2.7 section rows,
   which reads as "not one slab" until you ask what the rows are (hero + CTA). Counting rows answers
   a different question from counting prose sections.
4. **Two agreeing measurements that share an encoding are ONE measurement.** Two lanes independently
   reported designblog serving zero images; both had encoded *image* as `<img>`, and every hero here
   renders as a CSS `background-image`. Use
   `grep -o "background-image:[^;}]*url([^)]*)"` alongside any `<img>` count.
5. **A positive finding survives a blind predicate; an absence does not.** Run the control whenever
   the claim is an absence. My probe of `content_data->>'eyebrow'` returned a false "0 of 15" — the
   key is `eyebrow_text`. List `jsonb_object_keys` before probing.
6. **Probe the PUBLISH TARGET, not the customer domain.** `boxingonline.com` has no DNS and returns
   000; the site serves at `boxingonline.ugg2.com`. `sites.publish_target` / `publish_project` name
   it. I probed ~40 times without ever stating which host, which made a peer's correct observation
   look like a contradiction.
7. **Backticks inside `git commit -m "…"` EXECUTE.** They ate three identifiers out of a commit
   message. **Use `git commit -F <file>`.**
8. **Rehearse every migration under `BEGIN`/`ROLLBACK` before submitting it** — mine contained
   invalid PL/pgSQL (`IF <expr> FROM …`) that only a rehearsal could catch.

---

## 7. What to do next, in order

1. **Ask the owner to refresh the kubeconfig token** — nothing below is verifiable until then.
2. **Run the post-roll migration dry-run** (§0.2) and compare against `Pending (164)`.
3. **Verify whether 035 P1 shipped in `v1.0.1355`**: probe the binary for `rerenderFlatSections`
   **with a must-be-present control** (`PlanSectionsAction`) **and a must-be-absent control**. When
   the token expired, all five probes returned "absent" including the control — without it, the
   result read as "P1 did not ship". Then `git merge-base --is-ancestor <commit> <stamp>`.
4. **Do NOT restart imagery work at the component layer.** The live question is §3, it belongs to
   whoever owns the planner, and it is already routed into `bugs_open/114`.
5. **`news-listing`** — same defect as `article-body` had, still unwritten. It carries the real news
   surface. It should follow the §3 answer, not precede it.
6. **035 P1** — resume from `features_open/035_FEATURE_component_hierarchy.md` §5, and read the new
   **hazard 6.9** first (`loadStoredSections` has no `parent_instance_id` filter, so a nested walk
   silently mis-attaches per-section figures; plus the `MergeLockedPageSlots` polarity trap).

---

## 8. Identifiers you will want

- boxingonline site id `d2aa5206-73bc-4707-a69c-2702c1eb9152`, serves at `boxingonline.ugg2.com`
- `article-body` component id `5835b2e1-50d7-4f20-8a9c-8da4d270ae3d`, pre-686 md5 `002cbcd9cada6a37bf4a5158fd1e5f22`
- council correlation (686, APPROVED r2) `4bf6c48f-9cd6-440f-9257-a5668b6635fc`
- live planner definition `f263eaa1-61e1-446e-9410-648e12b7875b` (`build-site-planner`)
- peer lanes: `inline_guide_imagery` (in-body imagery, IMG-075/644), `bugfix_114_imagery_wiring`
  (owns 114, built IMG-077), `designblog_couk` (owner critique, 18 remakes queued),
  `site_delivery_and_editor` (owns the boxingonline pipeline — this lane dispatches nothing there)

---

## 9. POST-ROLL VERIFICATION — done 2026-09-02 after the token refresh

**All of §7's items 1–3 are discharged. Nothing below changes §2–§4.**

**035 P1 SHIPPED in `v1.0.1355`.** Fully-controlled binary probe, controls on OPPOSITE sides
(`LANDMINES.md:492` — never ship a probe whose controls are all on the same side):

| symbol | result | role |
|---|---|---|
| `PlanSectionsAction` | **PRESENT** | must-be-present control ✓ |
| `zzzInventedControl_NotInAnyBinary` | absent | must-be-absent control ✓ |
| `rerenderFlatSections` | **PRESENT** | the P1 extraction (`2a0bdb001`) |
| `hierarchyChildrenOf` | **PRESENT** | membership helpers (`bc8167100`) |
| `recomposeAncestors` | absent | **NOT a missing commit** — see below |

⚠ **`recomposeAncestors` reads absent because it has NO CALLER** — every source hit is inside its own
definition, so Go's linker eliminates it. `3fd617ef6` is in the tree. **An absent capability literal
means "not reachable in this build", never "the commit did not ship."** P1's wiring is still
incomplete, which is the true statement that absence encodes.

⚠ **Two probe hazards hit in one session, both worth inheriting.** (a) When the token expired, ALL
FIVE symbols read `absent` **including `PlanSectionsAction`** — without that control, "P1 did not
ship" is what you would have written down. (b) `LANDMINES.md:6277`: BusyBox `grep` over `/proc/1/exe`
needs **100–120s PER GREP**, and a grep killed by the command timeout is **indistinguishable from a
negative**. Do not chain several greps in one 2-minute call — run them singly.

**Migration dry-run (mandated after every roll): `Pending (176)`**, up from **164** pre-roll — 12
files accumulated in a day. **And the number that actually matters for `bugs_open/426`: the runner
flags 34 files `LIKELY ALREADY APPLIED; its own guard raised`** — i.e. applied by hand and never
recorded, which is the replay hazard that blocked the runner for three days once. 426's §8 left the
fleet-wide figure `[UNMEASURED]`; **34** is it. Contributed to that bug.

**686 rollback survived the roll**: `article-body` still md5 `002cbcd9cada6a37bf4a5158fd1e5f22`,
len 1378, `hero_image_url` absent. **Migration 708 still NOT applied**, so IMG-077 remains inert
despite the roll it was waiting for — that is the 114 lane's to apply, not this one's.

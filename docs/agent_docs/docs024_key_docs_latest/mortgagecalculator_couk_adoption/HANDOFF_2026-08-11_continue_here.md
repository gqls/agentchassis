# HANDOFF — mortgagecalculator.co.uk — cold start, read this first (2026-08-11)

**Supersedes `HANDOFF_2026-08-10c_continue_here.md`** (read it second: its §0 owner
rulings, §2 corrections and §6 landmines all stand). Site: UNLOCKED. Everything
this session changed is **config or content, live immediately** — no image
dependency. Chassis rolled 20:38Z; nothing here needs it.

## 0. Owner rulings in force

1. Correctness beats fidelity — never copy a wrong original; improve past it.
2. The checker proves results don't differ on identical inputs, and catches wrong
   results. 3. Site stays unlocked. 4. Everything runs from the framework.
5. Both-right-differently → supply BOTH, the alternative on its own signposted page.
6. **NEW — the framework writes the content, but the VOICE is ours to specify**
   (2026-08-11, throughout). Owner rewrote the brief four times in one evening;
   the resulting rules are in `site_specs.content_direction` and the reasoning is
   in `REFERENCE_2026-08-11_learned_by_correction_house_voice.html`.
7. **NEW — the model is NOT the lever.** Owner asked to revert the writer to
   Gemini, then withdrew it ("ok, don't change the model then") once the cause was
   shown to be the brief and the voice spec. **Do not change the writer model.**

## 1. What this session did, in one paragraph

Executed the equity-release decision (rebuild + fence, PASSED); diagnosed and
fixed three **ghost Calculate buttons** (migration 393, 1.05:1 → 15.39:1, live and
independently re-observed); then diagnosed the homepage "AI slop" as **a
commissioned brief plus a recorded voice spec, not a model failure**, rewrote the
voice spec four times under owner correction, rewrote 31 page titles and 15
homepage copy fields, and verified the layout survived every edit. Also stopped
the churn that was rewriting the homepage unattended.

## 2. THE FINDING THAT MATTERS BEYOND THIS SITE — an audit that reads the DB, not the site

`content-quality-auditor.load_page_content` selects from `page_components`. An
**adopted** page that serves HTTP 200 but has no components reads as EMPTY. On
2026-08-11 17:41Z that produced a `content_rewrite` whose brief said *"With no
retrievable content, there is no evidence of any differentiator explaining why a
user should use this calculator over MoneySavingExpert, Which…"*, with a pass mark
of *"a benefit not shared by all generic mortgage calculators"*. Ten minutes later
the framework built and deployed a homepage that passed that test — which is the
copy the owner called slop. **The competitive voice was commissioned, and a
different model would have written differently-worded competitive copy.**

Two further blindnesses in the same agent: `LEFT(...,1000)` judges differentiation
from 1,000 characters of raw HTML including tags; and `load_brief` reads
`site_specs` aspect **`site_plan`**, which this site does not have, so it ran with
`tone`, `target_audience` and `key_messages` all empty.

**OWED: a `090` diagnosis run on this class.** It is durable, cross-cutting (every
adopted site) and the cause sits outside the symptom — CLAUDE.md's own criteria say
file it rather than assert it. **I did not run it; do not repeat my claim as
established without it.** Distinct from `bugs_open/253`, which is about markup loss
*during* a rewrite; this is about the brief that *orders* the rewrite.

## 3. The voice spec — four corrections, and the reasoning is the valuable part

`site_specs.content_direction` (site-scoped, superseded rows keep the history).
**The writer reads ONE field, `formatted`** — RUNBOOK §8; raw SQL that updates the
arrays and not `formatted` looks applied and steers nothing. Mine was regenerated
each time and **verified line-for-line against `datahelpers.FormatContentDirection`**
in a scratch Go module, with the check proven able to fail.

What the original spec (extracted from the adopted original, 08-02) actually said —
worth reading, because it explains weeks of output: *"galvanising rather than
reassuring"*, *"use the lender's voice ('we')"*, coined labels (*"The Inheritance
Destroyer"*), emoji on cards, and *"do not write in a reassuring or apologetic
tone"*. The owner's request had been forbidden in writing since adoption.

**The four corrections, each of which was a defect in a rule I wrote:**

1. **Staccato.** I set `sentence_style` from the readability rail's ASD-STE100
   ceilings (20 max / 15 avg / one idea per sentence). Those are for
   safety-critical technical instructions read by non-native speakers; they forbid
   the subordinate clause, which is what makes English read as considered. Now:
   considered sentences of 25–40 words, at most one short sentence per section.
   `WRONG_CALLS.md` — **provenance is not fitness; name the corpus a borrowed
   threshold was measured on.**
2. **Presumption is a DENSITY property.** An outright ban produced flatness. The
   owner: the presumptive heading *"reads fine as a one-off … it was the barrage"*.
   Now a frequency rule: at most one per page, rarely.
3. **The filler list is a smell, not a crime.** I reported four sound card
   descriptions as defects purely because they matched my own ban-list. **The owner
   reviewed them and accepted them as they are — do not "fix" them.** The spec now
   says explicitly: do not hunt existing pages for listed words.
4. **Register must hold within a clause.** My rule "contractions preferred in
   ordinary sentences" was too blanket: a formal word takes the full form
   (*are not just arithmetic*, never *aren't just arithmetic*).

**The rule worth more than the lists:** *do not write a sentence no one would say
out loud.* Both phrases the owner rejected were grammatical and on-message
("the questions a single number can't settle", "nothing here is selling you a
deal") and no vocabulary rule would have caught either.

## 4. Live state of the site

- **Homepage copy, live and verified.** H1 *"There's usually more to a mortgage
  than the rate"*; sections *"Calculators for the parts that are hard to picture"*,
  *"Reading round the subject"*, *"If you'd like somewhere to start"*.
- **31 page titles rewritten**, benefit-led (*"What stamp duty will cost you"*,
  *"If your home is worth less than your loan"*). **`pages.title` does TWO jobs** —
  the `<title>` tag AND the homepage card heading, because `tool-list`/`guide-list`
  render `items[].title` verbatim. The card items hold a FROZEN copy (both
  components have `data_sources` NULL), so the same transaction re-pointed them at
  `pages.title` by SQL join with a card-label = page-title guard.
- **⚠ OPEN: the other 30 pages still SERVE their old `<title>`.** The new value is
  in `pages`; it only reaches the HTML when each page is next assembled. The
  homepage got one because it was the page being re-rendered. A per-page §10b
  assemble pass finishes it. **Cards are already correct** — those read `pages.title`.
- **Ghost buttons fixed** (migration `393` + ROLLBACK, recorded `record-only`):
  the generator wrote `--x: var(--x, #lit)`, a self-cycle, so the fallback can never
  apply and the subtree is poisoned even though `:root` defines the token.
  equity-release, stamp-duty, rate-forecaster: transparent button, white label,
  **1.05:1** → **15.39:1**. Fleet-wide `LANDMINES.md` entry + CONTRIB to
  `staged_component_build` (one regex catches the class:
  `(--[A-Za-z-]+): *var\( *\1[,)]`).
- **Acceptance: 11 fences installed, all green.** equity-release, stamp-duty,
  rate-forecaster re-ran 19:30–19:33Z after 393 — **PASSED 4/4 each, zero
  `acceptance-fail` notes**, and their vision passes report the layout clean, which
  is the ghost-button fix confirmed by the mechanism that found it.
- **Layout survived every edit.** Same measurement both sides, live page before and
  after: **48 distinct classes, 88 class attributes, 33 links, zero classes
  decreased.** This is the `bugs_open/253` check. ⚠ My first attempt counted
  `class="card` / `tool-grid` and got 0/0, which looks exactly like flattening —
  **wrong needles for this template** (it uses `tl-card`, `guide-card`). Baseline,
  do not read raw zeros.

## 5. The churn, and why the page is now stable

`improvement-sweep` (re-enabled that morning by migration `389`, cost-watched) swept
at 17:41Z and filed three items on a false premise. It was **disabled again ~17:54Z**
and is `enabled=f`. Two of its three items are now closed by this lane
(`d1cd9757` content_rewrite → `wont_fix`, premise false and it conflicts with the
live spec; the `needs_content_planning` lineage → `wont_fix`). **Nothing queued
rewrites copy.** If the sweep is re-enabled, the same differentiation brief can
recur — the spec now contradicts it explicitly, but **which wins is UNMEASURED.**

## 6. Dispatch — read this before concluding a message was dropped

Extends RUNBOOK §15. Two things starve this site:

1. `build-pipeline-trigger.find_dispatchable_site` picks **one site per 120s tick,
   ordered by the fleet's globally OLDEST dispatchable item** (`LIMIT 1`). Measured
   2026-08-11: ~273 items older than a fresh one, 81 of them 18 days old.
2. **The same query skips any site holding a `claimed` item.** Two by-hand nudges
   appeared to do nothing because a planning job held the single-flight slot for 80
   minutes. **Check for a `claimed` row before re-firing** — a quiet nudge usually
   means busy, not dropped.

Bypass (precedent `scripts/initial_messages/180_adoption/081b_…`): one `orchestrate`
message pinning `build-dispatch-loop` to the site. Worked four times tonight.

## 7. Editing page copy — the ONLY path that works

- `content_data` edits are **invisible** until a re-render; the page is served from
  `page_components.rendered_html`.
- **An assemble-only `page_rerender` will NOT do it.** One reported `complete` with
  `rendered_html` untouched: assemble concatenates existing component HTML and
  re-reads `pages.title`, so **the `<title>` changed and the body did not.** A
  complete work item is not a repaired artefact.
- **`apply_section_edit` is the action that writes `rendered_html`**
  (`section_editor_actions.go:229`, context built by `buildRenderContextFromDB`).
  Drive it per slot: `scratchpad/fire_section_edits.py` pattern, `edit_type =
  content_edit`, slots `hero` / `tool-list` / `guide-list` / `call-to-action`.
  Do NOT pass `items[]` — the render context reads it from the DB.
- **Do NOT fire a `content_rewrite` to fix copy.** `bugs_open/253` (same day,
  sibling site): kept 84% of the words and **0% of the layout classes**, and the
  shrink guard passed it because it measures text volume and is blind to markup.

## 8. Next actions, in order

1. **Finish the titles**: assemble the other 30 pages so their `<title>` matches
   `pages.title` (§10b, one per page). Purely mechanical; nothing else depends on it.
2. **File the `090`** on the audit-blindness class (§2). Until then that root cause
   is *my claim*, not an established finding.
3. **Tell the port lane their premise moved**: `index` is now framework-managed
   (components created 17:51Z, `build_status='deployed'`), which their
   `PLAN_2026-08-11_decompose_into_framework.md` says must never happen outside
   their port. The improvement loop did it, unasked. Their pages and ours are
   otherwise still disjoint.
4. **`portfolio` still has no fence** and needs hand-written vectors (10c §5.2 —
   toolgolden drove its term to 1000 years and every emitted assertion is the
   validation message).
5. **A3**: `tool-generator` / `tool-improver` still need the `read_site_spec` step
   (10c §5.3). Read `CLM-021` first — these agents delete what the register omits.
6. Phase B, then the Phase C RFC (10c §5.8).

## 9. Landmines live on this work

- All of 08-08b §4, 08-08c §3, 08-10 §5, 10b §4 and **10c §6** still stand.
- **NEW — a CSS `--x: var(--x, #lit)` self-cycle**: fallback inoperative, subtree
  poisoned even where `:root` defines the token. Fleet-wide in `LANDMINES.md`.
- **NEW — a `triaged` item can starve for hours** behind the fleet's oldest
  backlog, or behind its own site's `claimed` row. Fleet-wide.
- **NEW — an assemble-only rerender updates `<title>` and not the body**, and
  reports `complete` either way (§7).
- **NEW — `pages.title` is both the `<title>` and the card heading** (§4).
- **A style rule is a prompt for judgement, not a substitute.** Four defects
  tonight were in rules I wrote, and every one was caught by the owner reading the
  live page — never by the rule, the check or the spec. The spec cannot test itself.

## 10. Files of record

This dir: `PLAN_2026-08-09_facts_into_tool_acceptance.md` (the acceptance design) ·
**`REFERENCE_2026-08-11_learned_by_correction_house_voice.html`** (the voice
reasoning; also published as an artifact for the owner) · `NOTES` (08-11 entries =
the whole evening, including four missteps) · `README_where_we_are` (owner's log) ·
**RUNBOOK §12 (recreation items), §14 (tool PLANs), §15 (dispatch starvation)** ·
`acceptance/verify_criteria.py` + `install_fences.py` + `criteria/*.json`.
Fleet: `LANDMINES.md` (+2 tonight), `WRONG_CALLS.md` (+2 tonight).
CONTRIBs out: `staged_component_build/…ghost_buttons_self_cycle.md`,
`vigilant_designer_offer_analysis/…the_offer_question_arrived_as_a_copy_complaint.md`.
Migrations: `393_fix_self_referential_css_vars_three_tool_components.sql` (+ROLLBACK).
Backups: `migration_backups` under `titles_2026-08-11b_benefit_led_titles`,
`homepage_copy_2026-08-11_benefit_led`, `393_…`.
Bugs: `bugs_open/218`, `222`, `225`, `178`, `253` (sibling lane's, governs §7).

## 11. OWNER OBSERVATIONS 2026-08-12, and why the FRAMEWORK didn't do it

Owner, looking at the live homepage: **the hero image is no longer there · the top
nav says just "Home" · the cards have no imagery but should have some.** Plus one
copy line: **"and you don't need to sign up for any of it."** is bad copy (§3's
rule — do not write a sentence no one would say out loud).

The owner's instruction was explicit: **do not fix these by hand — find out why the
framework didn't do it.** Every finding below is first-hand, this session.

### 11.1 The hero image — GENERATED, DEPLOYED, and committed to the WRONG PATH

> **CORRECTED 2026-08-12, hours after this section was first written and committed
> (`f1f0d30b2`): this is `bugs_open/248_…undeployed_asset_repair_deploys_every_asset_as_a_hero_under_a_placeholder_name.md`
> — ALREADY FILED on 08-10, already `090`-CONFIRMED, and its 08-11 contribution
> ALREADY NAMES THIS SITE** (an earlier instance the same morning, 10:36Z, asset
> `477838e3`). I wrote the section below as a fresh finding and only grepped
> `/bugs_open/` afterwards, which is the wrong order and the rule's own example.
> Logged in `WRONG_CALLS.md`. **Refer to that bug by SLUG — `248` is an ambiguous
> number** (the other 248 is a CTA/contact-link case).
> **What survives as this lane's own contribution** (now written into 248): the
> fleet census `assets.filename LIKE '%asset-key%'` **cannot see this site at all**
> — all five of its hero rows have `filename = ''` — so the 118/150-row figure is a
> **floor, not a count**; the count re-measured on `updated_at` is **150 rows / 16
> sites**, up from 118/10; and **rung 2 is still live today**
> (`asset-deployer.deploy_asset config.asset_key = "input_data.asset_key"`), so the
> `asset_key?` marker I flagged below as a possible fix is **not** one.

Not a detection failure and not a dispatch failure. The framework did the whole job
and filed the bytes under a template placeholder.

- `check_placeholder_image_in_use` filed `needs_hero_image` **five times**
  (`placeholder_image_in_use:hero`, priority 65/95): 3 `cancelled`, 2 `complete`.
- The `complete` one at **2026-08-11 12:46:30Z** generated the image
  (`banana/gemini-3-pro-image-preview`, 12:45:53Z), stored it, and git-committed it
  to `gqls/sites`. Its own `result` records both halves of the defect:
  `"hero_url": "/assets/images/hero.jpg"` alongside
  `"file_path": "/assets/images/input-data.asset-key.jpg"`.
- **`input-data.asset-key` is the slugified form of the literal string
  `input_data.asset_key`** — a dotted-path INPUT that was never resolved, shipped as
  a value.
- **Proven on the wire, 2026-08-12:**
  `/assets/images/input-data.asset-key.jpg` → **200, 68,984 bytes, `image/jpeg`**.
  `/assets/images/hero.jpg` → **404**. The hero's inline
  `background-image: …url('/assets/images/hero.jpg')` is still in the served HTML,
  so the section renders its dark gradient over nothing.
- `storage.DeployedAssetPath(assetKey, purpose)` (`platform/storage/url_helpers.go:317`)
  returns the correct `hero.jpg` **only** when `assetKey == "" || assetKey == purpose`;
  handed a literal it faithfully builds a filename from it. The shared derivation
  (`bugs_open/168`) is not at fault — its input was a template that never resolved.
- The item closed `complete` because every step returned success: generate ✓, store ✓,
  commit ✓. **A complete work item is not a repaired artefact** — again.
- ⚠ ~~The live config now reads `"asset_key?": "input_data.spec.asset_key"` … **[UNVERIFIED]**
  whether that change is the fix~~ — **RESOLVED 2026-08-12: it is NOT the fix.** That
  optional marker sits on the **caller's** `input_mapping`
  (`image-build-handler.call_asset_deployer`), while the literal that becomes the
  filename comes from the deployer's own config, and
  **`asset-deployer.deploy_asset config.asset_key = "input_data.asset_key"` is still
  there today** — 248's rung 2, unchanged. Both rows were updated 2026-08-11
  21:52:40Z (nine hours after the bad deploy) but not in the part that matters. The
  trap that nearly fooled me: **two spellings of the same intent on one path** —
  `input_data.asset_key` (the one that lands in filenames) and
  `input_data.spec.asset_key` (everywhere newer) — so grepping for the newer spelling
  finds it "fixed". `schema_migrations` still has no row after 20:00Z on 08-11, so who
  changed it is **[UNVERIFIED]**.
- **Also still true:** `image_url_404:hero.jpg` has been **`blocked`** since
  2026-08-05, and `image_source_unsatisfiable:…:hero` sits at
  `needs_human_review` priority **150** for 17 pages — every tool and guide hero.

### 11.2 The top nav — the data is FINE; chrome can only THIN, never THICKEN

- `site_nav_items` holds **16 items** for this site: 5 `primary`, 11 `utility`, all
  `active`, all with a `page_id`. Nothing is missing from the nav tables.
- The header renders **`primary` only** (`InjectHeader`, `component_library.go:2106`);
  the footer renders primary+utility. Hence the served page carries **16 footer links
  and 1 header link** — same site, same tables, same render pool. That contrast is
  the whole tell.
- Of the 5 primary items, **3 target pages have never deployed** —
  `/guides/index.html` (`planned`), `/investor/index.html` (`needs_rebuild`,
  `deployed_at` NULL), `/scorecard-simulator.html` (`planned`). `ChromeLinkPolicy`
  drops them so chrome cannot ship a site-wide 404. **That is correct behaviour.**
- The 4th, `/about/index.html`, deployed **19:38:18Z**. The stored header chrome was
  last written **18:06:39Z** — it predates it by 92 minutes.
- `loadFetchablePageSet` **always injects the site root**, so a "Home" item survives
  every filter. That is why the failure looks like "just Home" rather than an empty bar.
- **THE DURABLE, CROSS-CUTTING FINDING.** Chrome is written **once** behind an
  idempotence gate (`render_site_components_action.go:656`, the EXISTS probe), and the
  only repair channel — `markStaleChromeLinkSlot` (`chrome_link_policy.go:142`) —
  fires when stored chrome contains a link the policy now **REFUSES**. A nav that is
  **MISSING** an item contains no offending href, so **nothing ever marks the slot
  stale**. The repair is one-directional: a nav thins when a target stops being
  fetchable and never thickens when one starts. `chrome_link_policy.go:15-18` already
  names THIS SITE for the mirror-image case (a dead CTA beside a filtered nav); the
  omission direction has no channel at all.
- So the header's five slots are spent on hub pages the framework planned and never
  built, and the one that did build arrived 92 minutes too late to be seen.

### 11.3 The cards — the right work items exist and are in a status DISPATCH CANNOT SEE

- Every card item carries **`"image": ""`** — the field exists, the template supports
  it, the value was never filled. **Zero `<img>` in any card** on the page; all cards
  render a generic inline SVG glyph.
- The framework filed exactly the right work:
  `needs_imagery:section:index:1:icon_stamp_duty`, `…:icon_affordability`,
  `…:icon_repayment`, `…:icon_scorecard` (+ `…:infographic_decision_engine`),
  **priority 98**, handler `image-build-handler`, created by `build-site-planner` at
  **2026-08-02 23:30:20Z**.
- **72 seconds later all 13 `needs_imagery` rows were set to `deferred`**
  (23:31:32.884181Z), in one batch with 3 `needs_page` + 1 `needs_rerender`.
  `handled_by` NULL, `attempt_count` **0**, `error` empty — **never attempted**.
- Dispatch claims only `status IN ('triaged','approved')`
  (`claim_work_item_action.go:102`). `TriageDetectedItemsAction` promotes
  **`detected` → `triaged`** only. **Nothing promotes `deferred`.** So these rows are
  structurally invisible to dispatch and have sat for **10 days**.
- **`deferred` is a black hole**: not terminal (so it still holds the dedup slot in
  `idx_swi_dedup`, blocking a re-file of the same `item_key`), yet unreachable by
  dispatch. Four bulk-deferral batches exist on this site: 07-31 23:24, 08-02 23:31,
  08-03 11:02, 08-05 13:35.
- **[UNVERIFIED]** what set them. No Go path writes `deferred` for these item types
  (only migration `389` does, for `contrast_failure`, on 08-11). Most likely a session
  parked the adoption-time build queue by hand — **I did not establish it, do not
  repeat it as fact.**

### 11.4 Two more defects on the same page, NOT yet named by the owner

- **The hero CTA is bare text with no anchor.** The served `hero-content` holds
  `Work out your payments` as a raw text node — no `<a>`, no `href`, no class. The
  hero has no working call to action at all. (Distinct from the §4 ghost-button
  contrast fix, which was about a poisoned CSS var; this element has no link element.)
- **One tool card renders an empty description.** `<p class="tl-card-desc"></p>` on
  stamp-duty, because its `pages.meta_description` is empty; **9 of 31 pages** have
  none. The other 5 tool cards and all 4 guide cards are populated.
  > **CORRECTED in-session:** my first read said *every* card description was empty —
  > I generalised from the first card in the HTML. It is **1 of 6**. The check that
  > caught it was counting all matches instead of reading one
  > (`empty=1` of `tl-card-desc: 6`), i.e. the "a count you kept is not a census" trap.

### 11.5 What these three share, and what is OWED

All three symptoms are the **same shape, three times**: detection worked, the right
item was filed, and the artefact never arrived — once because the deploy path was an
unresolved template, once because the repair channel only runs in the thinning
direction, once because the item sits in a status dispatch cannot select. In no case
was the framework missing a mechanism. **This is the "a silent mechanism is UNDRIVEN,
not missing" class, and 11.1's closing `complete` is the "a complete work item is not
a repaired artefact" class.**

**Candidates for `090` — TWO, not three (revised 2026-08-12):**
1. Chrome nav repair is one-directional — it can thin, never thicken (every site).
2. `deferred` has no promotion path back to dispatch, while still holding the dedup
   slot that would let the finding be re-filed.
3. ~~An unresolved dotted-path input becomes a filename~~ — **already filed AND
   already `090`-CONFIRMED** as `bugs_open/248_…placeholder_name.md` (correlation
   `b78e9a04-9a91-4261-af86-fb79f9316a4e`). Do not spend a run on it; contribute to
   that file, which this lane has now done. **Grepping `/bugs_open/` first would have
   saved this section from being written twice.**

Per CLAUDE.md these are exactly the "cause lives outside the symptom / fix changes
behaviour fleet-wide" cases that must go through the loop **before** being asserted
durably. §2's audit-blindness `090` is **still owed** as well. Nothing above has been
through it, so treat 11.2's mechanism and 11.3's black-hole claim as **this session's
first-hand reading, not established findings** — the wire evidence in 11.1
(200 at the wrong path, 404 at the right one) is the one item that stands on its own.

**Do NOT hand-fix any of it** (owner, this session). In particular: do not
`git mv` the stranded JPEG into place and do not hand-write nav rows — that
converts a diagnosable framework defect into an invisible one.

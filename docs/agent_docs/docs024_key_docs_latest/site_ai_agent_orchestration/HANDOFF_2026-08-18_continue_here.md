# HANDOFF — ai-agent-orchestration.com. START HERE. Written 2026-08-18 ~12:15Z.

**Supersedes `HANDOFF_2026-08-05_rebuild_scope.md`** for current state. That file is still the
record of the rebuild scoping and its `bugs_closed/194` analysis — read it second, not first, and
treat every figure in it as 13 days stale (the ones that mattered are re-measured below).

> ## ✅ NOTHING IS BLOCKED. Contrast is DONE except `pricing` (8 left, rebuild-only). Images and carousels are NOT STARTED.
>
> **UPDATED 2026-08-18 ~18:40Z — FAMILY B IS APPLIED, PROPAGATED AND VERIFIED LIVE.** `index` and
> `about` now measure **0** firm contrast failures; the site total is **32 → 8**, and the 8 are all
> family A on `pricing`. Shipped as **migration `469`**, NOT `459` — that number was taken by
> another lane before it could be written (see §3a). Everything below that describes family B as
> pending is superseded by this line.
>
> | ask | state |
> |---|---|
> | **contrast** | **Families A and B both done and PROVEN AT THE ARTEFACT: 44 → 32 → 8 firm failures**, 0 regressions at either step. Migrations `456` + `457` + **`469`** applied, propagated, verified. The remaining **8** are family A on `pricing` ONLY, which no re-render can reach — see §3c |
> | **images** | **NOT STARTED.** Fully scoped in §4. One component, 10 images. ⚠ the obvious handler would DELETE them |
> | **carousels** | **NOT STARTED — but FAR further along than this lane first reported.** Two fully-specified carousel patterns ALREADY EXIST in the experience register with complete behaviour contracts. The work is APPROVE + BIND, not design (§5) |
> | `pricing` rebuild | **OWNER APPROVED 2026-08-17, not yet dispatched** (§3c) |
> | **white cards (family B)** | ✅ **APPLIED, PROPAGATED AND VERIFIED LIVE 2026-08-18 18:36Z as migration `469`.** 24 failures fixed, 0 regressions by colour pair. The either/or put to the owner on 08-17 is RETIRED (§3a) |
> | **imagery policy** | **OWNER RULING 2026-08-18 APPLIED** — migration `458`, people at work permitted, impersonation banned (§6) |
>
> **`bugs_open/029` is being worked by the owner in another thread (2026-08-18). DO NOT FORK IT.**
> It delayed this lane by ~40 minutes yesterday and **has since drained** — 69 of my 70 queued
> items completed overnight. If the queue wedges again (signature: every site at exactly
> `claimed=1`), that is 029, it is owned, and the correct action is to record what you saw and
> wait, not to diagnose.
>
> **What 029 actually is** (from that lane, 2026-08-18 —
> `CONTRIB_2026-08-18_from_the_029_lane_what_wedged_your_queue.md`, their docs in
> `bugfix_029_retry_kills_live_child/`):
> - ⚠ **The bug's TITLE is stale and misleading — the dispatch concurrency group is NOT
>   saturated**, and has not been since that file's own 2026-07-21 correction. Do not repeat the
>   title; this lane did and had to correct it.
> - The mechanism is a **per-site mutex** in `find_dispatchable_site`
>   (`NOT EXISTS (… site_id = … AND status='claimed')`). **One** orphaned `claimed` row removes
>   that whole site from dispatch — which is why the signature is *exactly* one per site.
> - **Cost is ~40 minutes per site per incident, not indefinite** — `claimed-item-timeout`
>   releases the claim. Read "029" as *delayed*, never *halted*.
> - ⚠ If it wedges again, **do NOT cancel the frozen orchestration** — it is that lane's evidence,
>   and cancelling does not release the claim anyway. If you need the site moving sooner than 40
>   minutes, release the **claim** and say so in your notes.
> - ⚠ Measuring it? Use **`last_activity`**, never `updated_at` — on a row the reaper has touched,
>   `updated_at` is when the *reaper* wrote, yielding a uniform ~4h26m that is the reaper's own
>   threshold and nothing to do with the job. That lane believed the wrong number first.
> - Their root cause (declared `timeout_seconds: 900` honoured on attempt 0 only, retries silently
>   recomputed to 5 min, final replay landing on a live loop) is measured.
> - ⚠ **THAT LANE HAS NOW WITHDRAWN ITS MECHANISM TWICE** (2026-08-18: first the optimistic-lock
>   race, then the "retry kills live work" reading — the child was already dead ~7-10 min before
>   the replay; staleness is the takeover's PRECONDITION, not its consequence). **This is why the
>   pointer below is not laziness.** Both withdrawals happened after this handoff would have
>   quoted them, and neither reached this file. Do not re-import a mechanism from their lane docs
>   into this one, however well-grounded it looks — that is precisely the move `bugs_open/048`
>   made against `029`'s previous wrong cause.
> - **The underlying MECHANISM is provisional and deliberately NOT restated here** — it is being
>   worked in `bugfix_029_retry_kills_live_child`; **read it there**. Two earlier versions of this
>   handoff carried a mechanism from that lane, and the first one it gave me was later **withdrawn**
>   by them. This file now points rather than asserts, at their request and for a good reason:
>   `029`'s own history records `bugs_open/048` restating 029's *refuted* cause as fact while
>   correctly diagnosing its own, because a confident cause in a handoff is what the next thread
>   builds on. A cold-start doc is the highest-propagation surface there is.
> - **What survives the withdrawals, and the one part that touches THIS lane** (their statement,
>   2026-08-18 — impact, not mechanism): the retry-window truncation also fires **one level down,
>   inside the child, where it abandons real page-building work**; and a replay **duplicates a
>   non-idempotent spawn — a real `page-rerender` agent and K8s job — per incident.**
>   `page-rerender` is this lane's own item type, so that is our exposure, not a bystander's.
>   **CHECKED 2026-08-18 on this site: NO duplication damage.** Three duplicate
>   `(page_id, slot_name)` groups exist and all three are **legitimate** — the discriminator is
>   content identity, not slot repetition, and each group's `count(DISTINCT md5(content_data))`
>   equals its row count (3/3, 2/2, 2/2: repeated `generic-text-block` slots). ⚠ **Re-run this
>   after any future incident, and do NOT "deduplicate" on slot repetition** — a unique index on
>   `(page_id, slot_name)` breaks 11 legitimate pages fleet-wide (LANDMINES):
>   ```sql
>   WITH dups AS (SELECT pc.page_id, pc.slot_name FROM page_components pc JOIN pages p ON p.id=pc.page_id
>                 WHERE p.site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' GROUP BY 1,2 HAVING count(*)>1)
>   SELECT p.name, pc.slot_name, count(*), count(DISTINCT md5(pc.content_data::text)) AS distinct_content
>   FROM page_components pc JOIN dups d USING (page_id, slot_name) JOIN pages p ON p.id=pc.page_id
>   GROUP BY 1,2;   -- act ONLY on groups where distinct_content = 1
>   ```
>   Also theirs: each takeover **re-stamps `last_activity` and RESETS the 4-hour reaper clock**, so
>   poking a wedged row buys it another four hours. Another reason not to touch one.
> - ⚠ **A LOG GREP CANNOT CONFIRM ANY OF THIS, AND WILL LOOK LIKE IT REFUTES IT.** Chassis log
>   retention on this cluster is about **FOUR MINUTES** — a 24h query returns lines from minutes
>   ago on a pod that started hours ago. So grepping for the takeover returns zero **with the
>   control also zero**. That is **blindness, not absence**: do not cite it as runtime evidence,
>   do not "confirm" the mechanism with it, and do not conclude the takeover never fires. This
>   holds whatever the diagnosis returns — it is a property of the cluster, not of anyone's
>   theory.

---

## 1. What is DONE and LIVE (each verified at the artefact, not at a status)

| thing | state |
|---|---|
| **Migration 456** — 12 templates repoint foreground `--color-primary` → `--color-primary-ink` | **APPLIED + PROPAGATED.** `UPDATE 12`, guards passed |
| **Migration 457** — `.stats-cta` → `--color-accent-text` | **APPLIED, PROPAGATED AND VERIFIED LIVE 2026-08-18 ~12:45Z.** `a.stats-cta` firm failures: **0** |
| **Regression control** — colour pairs live now that were ABSENT from the baseline | **0.** This is the check 456 needed and did not have |
| Contrast on index / about | index 17 → **9**, about 19 → **15** |
| Contrast on services | **0**, before and after (it was already clean) |
| Contrast on pricing | 8 → **8, unchanged and expected** — the page cannot re-render at all (§3c) |
| `CONTRIB` filed to the `bugfix_122_contrast_ink_slots` lane | committed 2026-08-17 |

### ⚠ RENDERED ≠ DEPLOYED — RESOLVED 2026-08-18, but keep the check

**Now closed**: the batch drained (42 complete), the page republished, and `a.stats-cta` measures
**0** firm failures live. Recorded because the intermediate state is the trap: for ~30 minutes
`457`'s fix was in `page_components.rendered_html` (2 rows) **while the live page still served
1.61:1**. A component can be correct in the database and wrong on the internet, and **the DB
query is the one that looks like success**. Check both, and treat the live page as the only
verdict:

```sql
-- (1) is the COMPONENT fixed?
SELECT count(*) FROM page_components pc JOIN pages p ON p.id=pc.page_id
WHERE p.site_id='2a8ebf9c-20a2-4c39-b191-840b012371da'
  AND pc.rendered_html LIKE '%--color-accent-text, var(--color-primary%';   -- 2 on 2026-08-18
```
```bash
# (2) is the PAGE fixed?  This is the one that counts.
python3 scripts/render_audit.py https://ai-agent-orchestration.com/index.html
```

**Queue state, re-measured 2026-08-18 12:23–12:25Z: this is ORDINARY CONGESTION, not a wedge and
not `bugs_open/029`.** An earlier draft of this handoff guessed it was "plausibly a consequence of
the owner's 029 work in flight". That guess is **refuted** — the 029 lane has run no mutations
against `site_work_items` or `orchestration_states` — and the measured answer is dull:

- `build-pipeline-trigger` is `enabled`, `interval_seconds=60`, `last_completed_at` **8 seconds**
  before I looked. The trigger is healthy.
- The build pipeline is **completing 2–3 items per minute fleet-wide**, every minute, for 45.
- This site holds **ZERO `claimed` rows of ANY item type**, so the per-site mutex is not
  excluding it.
- **15+ sites** have `triaged` / `pipeline='build'` work, all filed within the preceding hour, and
  this site's rows (12:07–12:10) are among the **newest** — i.e. at the back of the queue.

⚠ **The query that produced "claimed=0" could not have seen the answer**, and that is the part to
carry forward: it filtered `item_type='page_rerender'`, while the mutex is per **SITE across all
item types**. A claim held by any other type would have been invisible. **A filtered count cannot
rule out a cause the filter excludes.**

**Re-fire nothing:** the rows are queued, not lost, and a missing completion is not a lost message.

**Re-measure it yourself in one command** (never trust the table above):

```bash
python3 scripts/render_audit.py --json /tmp/aiao.json \
  https://ai-agent-orchestration.com/{index,about,pricing,services}.html
# then count FIRM failures only - exclude overImage, or you will over-report by ~3
```

## 2. The one thing this lane got wrong, so it is not repeated

**456 repointed foreground declarations REGARDLESS OF THE GROUND THEY SIT ON, and that broke an
element it should have left alone.** `--color-primary-ink` is derived to clear the contrast floor
against the **page** grounds (background, surface, composited overlay). It carries **no guarantee
against a fill**. `.stats-cta` is an accent-filled button, so 456 changed its label from `#0D1117`
on `#F0A500` (near-black on amber, legible) to `#768eb2` on `#F0A500` — **measured 1.61:1**.

Caught by re-auditing the same four pages after the change, **not** by the net improvement, which
looked like a clean win (44 → 33) and hid it. Fixed by `457` using the token the renderer already
emits for exactly this case (`--color-accent-text`, computed live as `#294155`).

> **RULE for whoever does the remaining 144 templates: a foreground repoint is safe only when the
> declaration's own rule block sets no `background`. Census the BLOCK, not the declaration.**
> Of 456's 36 repointed declarations exactly one sat on a fill.

Two further self-corrections from this lane, both in `WRONG_CALLS.md` / `NOTES`:
- **A palette re-value was offered to the owner as the "quick safe" option and was not a route at
  all** — `--color-primary` is dual-role here (37 foreground / 24 background uses), so lightening
  it trades 20 failures for a fresh set. Withdrawn before anything was applied.
- **`rerender-pages` does not render.** It files one `page_rerender` work item per page and
  returns `COMPLETED` in seconds. A green run means "41 rows filed", not "41 pages rendered".

## 3. Contrast — what is LEFT

> **UPDATED 2026-08-18 18:36Z: family B is FIXED. Only the 8 family-A failures on `pricing`
> remain**, and they are unreachable by any re-render. Re-measured live after `469`:
> `index` **0**, `about` **0**, `services` **0**, `pricing` **8**. No colour pair exists in the
> after-set that was absent from the before-set. The composition table below is the PRE-469
> reading, kept because it is what the diagnosis was built on.

The 33 survivors are **not** the defect 456 fixed. Composition, measured 2026-08-18 (PRE-469):

```
 20  LIGHT-ON-LIGHT  #E6EDF3 on #FFFFFF      <- family B, needs an OWNER DESIGN DECISION (3a)
  4  LIGHT-ON-LIGHT  #E6EDF3 on #F8F9FA      <- family B
  7  DARK-ON-DARK    #0D1117 on #0D1117      <- family A, on `pricing` ONLY (unrenderable, 3c)
  1  DARK-ON-DARK    #0D1117 on #080B10      <- family A, on `pricing` ONLY
```
**Every surviving family-A failure is on `pricing`**, i.e. 456 cleared family A everywhere it
could reach. `pricing` is unreachable by any re-render, so those 8 close only via the rebuild.

### (a) Family B — 24 failures. TWO components with no theme support at all — ✅ FIXED by `469`

> **DONE 2026-08-18. Shipped as `469_departments_grid_and_leadership_team_consume_site_tokens.sql`,
> not `459`** — that number was taken by another lane (`459_zip_deliverer_agent_HOLD.sql`) before
> this could be written. ⚠ **Numbers in `sql_for_agents/` COLLIDE** (457 and 458 each name two
> unrelated migrations from two lanes), so cite the FILENAME, never the bare number.
>
> Applied, recorded, propagated by two **page-scoped** `template_changed` rerenders, and verified
> live: index 9→**0**, about 15→**0**, zero regressions by colour pair. Rollback restores both
> templates byte-exact from `migration_backups`. Full account: `NOTES_site_improvement.md`,
> 2026-08-18 evening; propagation recipe: `RUNBOOK` R8; single-migration apply: R9.
>
> ⚠ **Two corrections to what this section says below.** (1) `departments-grid` is placed on
> **index as well as about** — 9 of the 24 were on index, so "the two about-page components" is
> wrong. (2) **"both light ones are unchanged" was too strong**: only the CARD ground is identical
> (`#fff`→`#FFFFFF`). The section ground, the icon well and both text literals DO move to each
> site's own tokens. Nothing loses its contrast floor on either light site (lowest after-value
> 5.02:1 against a 4.5 floor) — but "nothing gets worse" is the claim, not "nothing moves".
> Neither light site has been re-rendered, so both still serve today's exact appearance.

> ⚠ **CORRECTED 2026-08-18. This lane previously said SEVEN components "hardcode a light ground",
> and offered the owner a choice between two designs. Both were wrong.** The seven came from a
> query that cannot tell a bare literal from a `var()` **fallback**, and five of the seven were
> fallbacks — `background: var(--color-background, #fff)`, present in the source and never applied.
> (A second, opposite error followed: a grep returned `.team-section { padding: 3rem 1.5rem; }`, a
> **media-query duplicate** of the selector, which made the templates look clean. Resolve component
> CSS through `component_id` against the database, never by grepping `html_template` by name.)

**It is `departments-grid` and `leadership-team`, and their ENTIRE colour surface is:**

```
background: #f8f9fa;   color: #555;     background: #fff;
background: #e0e0e0;   color: #0f3460;  color: #555;
```

**Not one themed value.** These two have **no theme support whatsoever**, in a library where the
sibling section component (`differentiators-section`) already does it correctly with
`var(--color-background, #fff)` / `var(--color-surface, #f8f9fa)`. On this site — the only DARK
site that uses them — they were never going to work.

⚠ **The fix is NOT "tokenise the backgrounds".** That would put `#555` and `#0f3460` onto a
`#0D1117` card — **a fresh set of invisible text, which is migration 456's mistake repeated.**
The whole block moves together. Every token needed already exists (read from the served `:root`):

| current | becomes | resolves here |
|---|---|---|
| `background: #f8f9fa` | `var(--color-background, #f8f9fa)` | `#080B10` |
| `background: #fff` | `var(--color-surface, #fff)` | `#0D1117` |
| `background: #e0e0e0` | `var(--color-border, #e0e0e0)` | `#21262D` |
| `color: #0f3460` | `var(--color-text, #0f3460)` | `#E6EDF3` |
| `color: #555` | `var(--color-text-muted, #555)` | `#8B949E` |

**Blast radius checked BEFORE proposing (the 456 lesson): 3 sites, and both light ones are
unchanged.**

| site | scheme | `--color-surface` | effect on the card |
|---|---|---|---|
| ai-agent-orchestration.com | DARK `#080B10` | `#0D1117` | **fixed** |
| finetuning.uk | light `#F5F3EF` | `#FFFFFF` | **identical to today's `#fff`** |
| leopardessconsulting.co.uk | light `#FAF8F4` | `#FFFFFF` | **identical to today's `#fff`** |

**This RETIRES the owner design decision recorded on 08-17.** It was put as *strip the white* vs
*keep light cards and darken the text*. Neither: the components should consume the site's tokens,
as their sibling already does, which gives each site its own answer instead of imposing one on all
three. What remains for the owner is only whether to spend it — not which design.

**Ready to write as migration `459`**, same shape as `456`/`457`: dry-run simulation first, exact
literal anchors, `DO`/`RAISE` guards, two-level fallbacks, rollback file. ⚠ Verify at the artefact
by **colour PAIR** afterwards — any `(fg,bg)` in the after-set absent from the before-set is a
regression you introduced (LANDMINES).

### (b) The 17 parked `contrast_failure` items — LEAVE THEM PARKED

They were parked by migration 389 deliberately: promoting them mints completions that are ungraded
by construction. **They drain by themselves** — `write_render_audit_findings_action.go:479`
retracts a row on a fresh positive measurement by the same instrument that filed it. So the way to
close them is to fix the page and let the site's render audit run, **not** to promote them.

### (c) `pricing` — approved for a framework rebuild, not yet dispatched

8 firm failures, 7 of them invisible `H3`s. **5/5 components have NULL `content_data`** and it was
last rendered **2026-04-13**, so `rerender_page_sections` has nothing to rebuild from — this page
is the reason "just re-render everything" was never going to be enough.

✅ **`pricing` is `rebuild_policy='generic'`, so the rebuild will not be refused** — checked
2026-08-18. ⚠ **Six pages on this site are `owned` (5 `deployed`, 1 `needs_rebuild`) and WILL
refuse.** That refusal is real and currently biting other sites: every `page_rerender` failure
fleet-wide created after 2026-08-17 16:12Z is a content-gating refusal, not a timeout —
`save_page_sections: overwrite: REFUSED`, `page <x> is rebuild_policy=…`, `claims floor blocked:
19 banned claim(s)`. So a rebuild aimed at an `owned` page fails with a clear error rather than
silently doing nothing; check the policy before dispatching, not after.

Owner approved the rebuild on 2026-08-17, **knowing it REGENERATES the copy rather than correcting
it**. Route: `scripts/initial_messages/020_build_pipeline/082_submit_domain_unified.sh` per the
owner ruling of 2026-08-04 (never hand-build). ⚠ **NEVER restore from `page_component_history`** —
`component_id` is NULLed by `ON DELETE SET NULL`, so pairing yesterday's content with today's HTML
makes the next rerender reinstate the old page (`bugs_closed/194` §4).

## 4. Images — NOT STARTED, fully scoped, and the obvious move is a trap

Every `<img>` on the site is one component, and there are ten:

```
index                           | case-studies-grid | (EMPTY src)                        x5
enterprise-reference-deployment | case-studies-grid | /assets/images/case-study-*.png    x5, HTTP 404
```

`content_data` is rich — five case-study titles, excerpts, stats, and genuinely good
`cardN_image_alt` prose describing the intended diagrams — but **there is no `cardN_image_url` key
at all**. The site's own items say why: *"sources field 'card1_image_url' from site_assets.image
which nothing generates"*.

> ⚠ **DO NOT fill in `handler_agent` on the `image_url_404` / `image_source_unsatisfiable` rows.**
> They are empty, which looks exactly like the missing piece, and `image-url-404-handler` /
> `image-source-unsatisfiable-handler` are both live. **Their workflows are
> `query_database` → `create_work_item` → `checkpoint_for_review` — they TRIAGE, they do not
> generate.** The only site they have ever run against is `mortgagecalculator.co.uk` (2026-08-14),
> which now has **zero `<img>` tags in any component**. Routing these rows there would most likely
> strip the five case studies.

Real generation is `image-generator` / `image-build-handler`. Then bind `cardN_image_url` and
deploy to stable `/assets/images/` paths — **not** pre-signed URLs (§6).

## 5. Carousels — NOT design work. Two patterns already exist; the work is APPROVE + BIND

> ⚠ **CORRECTION — an earlier version of this handoff said "no carousel component exists anywhere
> in the platform". THAT WAS WRONG.** It came from `grep -rli carousel platform/ internal/`, which
> searches Go source. The carousels live in the **experience register**, which is DATA — three
> tables, `experience_patterns` / `site_experiences` / `experience_invariants`. **A grep of the
> code cannot see a capability that lives in the database**, and the negative read as authoritative
> because the command "found nothing" rather than erroring. Found only because the owner said to
> consult the flow agents (2026-08-18).

**Two patterns exist, both `kind='component-contract'`, both with real behaviour contracts:**

| pattern | what it is |
|---|---|
| **`arrow-and-swipe-card-carousel`** | *"Card carousel: arrows always, swipe natively, auto-advance only if asked."* Native scroll-snap does the swiping so it works with **no JavaScript at all**; JS adds arrows, keyboard stepping, and optional auto-advance that yields to the visitor. `funnel_stage: awareness` |
| **`scroll-snap-card-track`** | *"A swipeable track of text cards, with no JavaScript behind it."* Swipe, trackpad and tabbing through the cards' own links all move it; nothing is scripted. `funnel_stage: consideration` |

### The behaviour contract the owner asked to be considered — it is already written

`arrow-and-swipe-card-carousel` specifies, in the register:

- **Fewer than two cards → every control (arrows, pause) is hidden.** Enforced by the
  `no-inert-control` invariant, which the entry declares in `requires_invariant`. *"A control that
  cannot change anything must not be presented."*
- **JS blocked or failed → the track still scrolls and snaps natively, every card is still a real
  link**, and the arrows (the only JS-dependent part) simply do nothing. The behaviour is an
  enhancement over a working component, not the component itself.
- **Script included twice → initialisation must be idempotent** (one set of handlers, one timer).
  The entry notes two different guard mechanisms already in use for this and makes choosing one a
  decision rather than an accident.
- **Scrolled out of view → auto-advance suspends** (IntersectionObserver, 0.25) and resumes on
  return. *"A rotation nobody can see burns battery and moves content under a returning visitor."*
- **Hover or focus anywhere inside → auto-advance suspends until they leave.** Reading is never
  interrupted by the component moving.
- **`prefers-reduced-motion` → auto-advance never starts**, and scrolling is instant, not smooth.
- **Visitor swipes directly → the component re-derives which card is current** 120 ms after the
  scroll settles, so the next arrow press continues from where the visitor is, not from where the
  code last was.

That is a better specification than this lane would have written from scratch, and it is the
answer to "please also consider their behaviour": **it is considered, it is written down, and it
should be adopted rather than reinvented.**

### The dead-destination problem is already designed out

`destination_roles` is `{{binding.card_destination_role}}` — a **binding**, not a URL.
`bind_site_experience_action.go` checks a destination role against the site's real `pages` **at
bind time**, so a carousel cannot promise a page that does not exist. That is the specific defect
behind *"the four dead carousel destinations found by hand on 2026-07-26"* (`bugs_open/023`,
`071`), and it is why the carousel must be adopted **through the register** rather than hand-built
into a component template.

### ⚠ THE ACTUAL BLOCKER: nothing has ever been approved, and the council has never run

**[MEASURED 2026-08-18]**

- `experience_patterns`: **11 rows, ALL `draft`, ZERO approved.**
- `site_experiences`: **2 bindings**, both on `noted.co.uk`, both `status='proposed'` — consistent
  with the rule that binding a draft is a proposal, not a commitment.
- `experience-approval-council`: **zero orchestration rows, ever.** The approval path has never
  been exercised end to end.

So this is the estate's familiar shape — **a mechanism that is built, careful and undriven** — and
the carousel work is *not* "write a carousel". It is:

1. Put `arrow-and-swipe-card-carousel` through `experience-approval-council`
   (`experience_register/260_TRIGGER_experience_approval_v1.sh`). **Expect to be the first ever
   run; budget for the path itself being untested, not just the verdict.**
2. Bind it to this site's target components with `bind_site_experience`, supplying
   `card_destination_role` so the bind-time page check can fire.
3. ⚠ **`section_types` is currently `["hero-carousel"]`.** The components the owner wants
   carouselled are card grids (`case-studies-grid` and friends), so either the entry's
   `section_types` needs widening or a sibling entry is needed. **Decide which — do not quietly
   bind outside the declared section type.**

**Best pairing: `case-studies-grid`.** Five cards, real per-card destinations, and it is the same
component §4's image work targets — so the two asks land on one component.

## 6. Imagery policy — OWNER RULING 2026-08-18, APPLIED

Owner, verbatim:

> *"we don't want fake headshots of people in the about us, but we can use pictures of typical
> offices or people working as long as we are not pretending that they are part of the company"*

**Applied as migration `458`** to `design_intent` (superseding, not mutating — 9 keys carried
forward, guards passed). It **relaxes and narrows in the same edit**, and both halves matter:

- **Relaxed.** `imagery_direction` said *"Technical illustrations and architectural diagrams ONLY …
  never staged corporate photography"*, which forbade exactly the photography the owner has now
  permitted. Ordinary working environments — offices, desks, screens, server rooms, whiteboards
  mid-discussion, people at work — are now allowed where they earn their place.
- **Narrowed.** The banned thing is **impersonation**, not people. The old `avoid` line was
  *"Testimonial carousels with headshots of fake people"* — it banned a **vehicle**, so the same
  deception stayed reachable through an about-page grid, a team strip or a founder quote. The
  replacement bans the act in any layout: *no photographed person presented, captioned or implied
  as a member of this company*; no invented team members, founder headshots, or testimonials
  attributed to a stock face.

⚠ **This is live for the imagery work in §4 and it is not a formality here.**
`departments-grid` and `leadership-team` — the two about-page components in §3a — carry a **120px
circular `.member-icon`**, which is a headshot-shaped hole. Whatever fills it must not read as
staff. The safest reading of the ruling for that specific slot is an abstract or illustrative mark
rather than a face; a photograph of a person in a circular avatar frame beside a department name
is precisely the placement the ruling calls out.

⚠ **The old carousel-shaped `avoid` line is GONE, so it no longer reads as a bar on §5's work.**
It never was a bar on carousels as such — it banned fake headshots inside one — but it was the
site's own standing text and would have looked like a contradiction.

## 7. Live traps on this site

- **The 9 hero/`content_hero` rows in `assets` are pre-signed Backblaze URLs**
  (`X-Amz-Expires=604800`, stamped 2026-08-11) — **they lapsed on 2026-08-18**. Only `og-card.png`
  and `favicon.png` are stable `/assets/images/` paths. No page component referenced one, so the
  blast radius looked nil; **[UNVERIFIED]** whether og tags, feeds or the asset renderer hold one.
  Do not mint new pre-signed URLs into content.
- **2 locked components.** Firing `section_data_resolved` at a locked, positionally-named section
  **duplicates** it rather than protecting it. Count `page_components` for the page before AND
  after; `bugs_open/189` records the reversal SQL.
- **The site is UNLOCKED** (`locked_at` NULL) and carries heavy scheduled automation
  (`model-directory-publisher`, `feed-ingester`, `content-feed-orchestrator`, `build-dispatch-loop`).
  Nothing will stop a dispatch — a reason for care, not permission.
- **The served stylesheet lies about headings.** It declares `h3, .site-footer h4 { color: #ffffff }`
  and that is **not** the winning declaration — the component's embedded `<style>` in
  `rendered_html` overrides it. Diagnose from `getComputedStyle`, never from the stylesheet.
- **`page-rerender` with no `reason` assembles from STORED `rendered_html`** and will ship the old
  CSS while reporting success. Only the `rerender_sections` branch regenerates from `content_data`.

## 7. Next actions, cheapest first

1. ~~**Finish 457 — applied and rendered but NOT DEPLOYED**, live `a.stats-cta` still 1.61:1.~~
   ✅ **DONE — and this item was ALREADY STALE when it was written.** §1 of this same file says
   457 was verified live at 12:45Z; this list says the opposite. §1 was right. Re-measured
   2026-08-18 19:14Z and again at 19:36Z: `a.stats-cta` produces **zero** findings on any of the
   four pages. The 41 queued rows drained on their own. **Nothing to do.**
   ⚠ *Two `## 7` headings exist in this file — this is the second. Where they disagree, the
   earlier sections have been re-measured and this list had not.*
2. ~~**Family B** (§3a) — write migration `459`.~~ ✅ **DONE 2026-08-18 as migration `469`**
   (`459` was taken). Applied, propagated, verified live: **24 of the 32 fixed, 0 regressions.**
3. **Carousels** (§5) — approve `arrow-and-swipe-card-carousel`, then bind. **No pipeline
   dependency**, so it proceeds whatever the dispatch lane is doing. Budget for being the first
   ever run of the approval council, and settle the `section_types` question before binding.
4. **Images** (§4) — generate via `image-generator`, bind `cardN_image_url`, verify over HTTP.
   **Read §6 first** — the imagery ruling is live and constrains the `.member-icon` slot.
5. **`pricing` rebuild** (§3c) — owner-approved, confirmed `generic` so it will not be refused;
   expect regenerated copy.
6. Let the site's render audit run afterwards; the 17 parked items drain on their own (§3b).

**Best order if picking one thing: (2), then (5).** Family B is designed and unblocked; the pricing
rebuild is approved and unblocked. Carousels are the most interesting but the least predictable,
because the approval path has never run.

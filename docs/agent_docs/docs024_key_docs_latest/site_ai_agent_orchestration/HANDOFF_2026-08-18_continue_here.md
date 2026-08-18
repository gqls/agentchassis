# HANDOFF — ai-agent-orchestration.com. START HERE. Written 2026-08-18 ~12:15Z.

**Supersedes `HANDOFF_2026-08-05_rebuild_scope.md`** for current state. That file is still the
record of the rebuild scoping and its `bugs_closed/194` analysis — read it second, not first, and
treat every figure in it as 13 days stale (the ones that mattered are re-measured below).

> ## ✅ NOTHING IS BLOCKED. The contrast half is SHIPPED AND MEASURED. Images and carousels are NOT STARTED.
>
> | ask | state |
> |---|---|
> | **contrast** | **Done for this defect family and PROVEN AT THE ARTEFACT: 44 → 32 firm failures**, 0 regressions. Migrations `456` + `457` applied, propagated, verified. The remaining 32 are **two other families** neither migration touches — see §3 |
> | **images** | **NOT STARTED.** Fully scoped in §4. One component, 10 images. ⚠ the obvious handler would DELETE them |
> | **carousels** | **NOT STARTED.** Design work, no pipeline dependency — the cheapest thing to pick up next (§5) |
> | `pricing` rebuild | **OWNER APPROVED 2026-08-17, not yet dispatched** (§3c) |
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

## 3. Contrast — what is LEFT, and it is not more of the same

The 33 survivors are **not** the defect 456 fixed. Composition, measured 2026-08-18:

```
 20  LIGHT-ON-LIGHT  #E6EDF3 on #FFFFFF      <- family B, needs an OWNER DESIGN DECISION (3a)
  4  LIGHT-ON-LIGHT  #E6EDF3 on #F8F9FA      <- family B
  7  DARK-ON-DARK    #0D1117 on #0D1117      <- family A, on `pricing` ONLY (unrenderable, 3c)
  1  DARK-ON-DARK    #0D1117 on #080B10      <- family A, on `pricing` ONLY
```
**Every surviving family-A failure is on `pricing`**, i.e. 456 cleared family A everywhere it
could reach. `pricing` is unreachable by any re-render, so those 8 close only via the rebuild.

### (a) Family B — 24 failures, components hardcode a LIGHT ground on a DARK site

Seven components paint themselves white and keep the site's pale `--color-text`:

```
about | departments-grid          #fff        index | departments-grid         #fff
about | leadership-team           #fff        index | differentiators-section  #fff
index | case-studies-grid         255,255,255 index | latest-news              #fff
index | system-stats              255,255,255
```

**This is untouched by 456/457 and immune to re-rendering** — the white is in the template. It is
the `hardcoded_section_colors` class the design-discovery agent already names, and the site carries
an unresolved `generic_theme` item. **Not yet diagnosed to a root cause** — do not assume it is one
edit. Decide first whether the fix is "remove the fill so the themed surface shows through" or
"keep the light card and set a dark foreground inside it"; those are different designs and it is
plausibly an owner question.

### (b) The 17 parked `contrast_failure` items — LEAVE THEM PARKED

They were parked by migration 389 deliberately: promoting them mints completions that are ungraded
by construction. **They drain by themselves** — `write_render_audit_findings_action.go:479`
retracts a row on a fresh positive measurement by the same instrument that filed it. So the way to
close them is to fix the page and let the site's render audit run, **not** to promote them.

### (c) `pricing` — approved for a framework rebuild, not yet dispatched

8 firm failures, 7 of them invisible `H3`s. **5/5 components have NULL `content_data`** and it was
last rendered **2026-04-13**, so `rerender_page_sections` has nothing to rebuild from — this page
is the reason "just re-render everything" was never going to be enough.

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

## 5. Carousels — NOT STARTED, and the cheapest thing to pick up next

No carousel component exists anywhere (`grep -rli carousel platform/ internal/` → two hits, neither
a component). The owner's instinct was right: this is a hint to the spec/planner. **It has no
pipeline dependency, so it is the one ask that cannot be blocked by 029.**

The existing guidance — *"For carousels/sliders: Use CSS animation, NOT complex JavaScript"* —
lives at `html_actions.go:527`, inside a **whole-page** generation prompt. **This site builds
through the component path, so that guidance never reaches it.** Putting the hint in the right
place is most of the work.

Constraints the hint must carry, each earned:
- **CSS-first** (scroll-snap / CSS animation); vanilla JS only if unavoidable.
- **Every control must resolve to a real page.** `bind_site_experience_action.go:36` records *"the
  four dead carousel destinations found by hand on 2026-07-26"* (`bugs_open/023`, `071`). The
  experience register checks destination roles against `pages` **at bind time**, so a carousel spec
  routed through it cannot promise a dead page. Use it rather than re-inventing the check.
- **Degrade to a legible list without JS**, so a carousel can never become an invisible-content
  defect of the kind §3 is about.
- Candidates: `case-studies-grid` (5 cards, already image-bearing — pairs naturally with §4),
  `departments-grid`, `leadership-team`, `info-card-grid`.

## 6. Live traps on this site

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

1. **Finish 457 — it is applied and rendered but NOT DEPLOYED.** The live `a.stats-cta` still
   measures **1.61:1**. 41 `page_rerender` rows sit `triaged` from orchestration
   `3d221a1e-2676-46bc-9e9e-f2c6e0a28cc7` (fired 2026-08-18 12:10Z). When the dispatch lane moves,
   re-audit `/index.html` and require `a.stats-cta` ≥ 4.5:1. **Do not re-fire** — the rows are
   queued, not lost.
2. **Carousels** (§5) — no pipeline dependency, pure design + prompt work.
3. **Images** (§4) — generate via `image-generator`, bind `cardN_image_url`, verify over HTTP.
4. **Family B** (§3a) — decide the design question first, then fix the 7 components.
5. **`pricing` rebuild** (§3c) — owner-approved; expect regenerated copy.
6. Let the site's render audit run afterwards; the 17 parked items drain on their own (§3b).

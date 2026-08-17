# HANDOFF — loancalculator.co.uk · the calculators are RIGHT and 14 DUPLICATE PAGES ARE LIVE (2026-08-17 ~16:35Z)

> Supersedes `HANDOFF_2026-08-17_continue_here.md` (which carries one false claim,
> flagged in its own banner — read it for the 282 proof, not for the wave).
> Full evidence + every misstep: NOTES `## 2026-08-17` §1–§10. Owner prose:
> README_where_we_are, same date.

```
site      loancalculator.co.uk   0162cde4-633e-45e9-8ca6-87a6b2fe1d26
chassis   RUNNING = digest sha256:f90a7e88… from commit 6a782274b (pods restarted
          14:43Z). ⚠ NOT the fresh build: a new image was pushed under the SAME tag
          v1.0.1305 (digest sha256:6039e19c…, commit 89a0cbeb7, built 15:30:44) and
          the node kept its cached binary. Running binary is 202 commits behind.
plan      9463e31d-ee50-482e-94a9-7e186ef25543  is_current — CORRECT for the 11
          calculators, WRONG for the guides (see THE DAMAGE)
locks     12/12 held throughout
pages     43 active, 42 serving 200. guides-index (/guides/index.html) is the only
          persistent 404. 14 real guides + 14 duplicate blog-posts both live.
verified  toolgolden --compare GOLDEN_2026-08-17_post_rebuild -> EXIT 0,
          "all 11 tools reproduce their golden values exactly" (16:17Z, index incl.)
```

## WHAT WENT RIGHT — do not re-litigate this, it is measured

**`bugs_open/282` is proven and the outcome reached the artefact.** The re-fire's plan
placed all 11 locked calculators on their own pages (0/11 → 11/11, joined on
`content_components.function`). Then `needs_page:index` rebuilt the homepage at
13:44:08 and **the locked calculator sits at position 2** — composed as the plan's
second section, where before it was appended at position 6. Locks 12/12. Toolgolden
exit 0 across all 11 tools including index. **The 282 fix + LOCK-008 + the recompose
did exactly what they were designed to do, end to end.**

## THE DAMAGE — 14 duplicate pages, live, and the containment window has closed

The same plan **retyped the guides section**: 14 pages of role `blog-post` at
`/blog/<slug>.html`, and zero of role `guide`. Measured across the three plans —
`34b1b056` guide=14, `dcbae4df` guide=14, `9463e31d` blog-post=14/guide=0 — so the
re-fire introduced it. 14 page rows were created 11:46:26 with zero components, their
`needs_page` items dispatched the moment the `bugs_open/243` claim gate re-opened
(12:10:17Z), and **all 14 are now deployed and serving 200** beside the 14 real
guides, which are intact. Same content, two URLs, 14 times.

The platform caught it unaided: `orphan_blog_posts` — *"14 blog posts deployed but
not linked from blog listing page"* — whose remedy minted a **43-page rerender batch**,
which is where the 40 still-`triaged` `page_rerender` rows come from.

### What is owed on it, in order — ALL OF IT NEEDS THE OWNER

1. **Decide the shape first, because it changes step 2.** Either the guides stay at
   `/guides/` and the 14 `/blog/` pages are retracted, or the site genuinely moves to
   `/blog/` and the guides are retired with redirects. **The plan of record currently
   says the second**, which nobody chose. Until this is decided, do not re-fire the
   planner — a fresh run will re-mint whatever the plan says.
2. **The 14 duplicates need retracting, and there may be no clean path.** They are
   deployed files in the bucket plus active page rows. Archiving the rows does not
   unpublish the files. ⚠ **This is the same class as `bugs_open/080`/`081` — a
   mistyped page, live, with no repair path, already flagged as needing an owner
   call.** Check that pair before inventing a method.
3. **The plan needs the guides back** (or the decision at (1) recorded), else the next
   reconcile treats 14 live, deployed, byte-verified guide pages as absent from plan.
4. Guarded SQL for the row half, if (1) goes the retract way — the zero-component
   guard is what makes it safe, keep it:
   ```sql
   UPDATE pages SET status='archived', updated_at=now()
   WHERE site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26'
     AND created_at > '2026-08-17 11:43:34+00' AND page_type='blog-post'
     AND status='active'
     AND NOT EXISTS (SELECT 1 FROM page_components pc WHERE pc.page_id=pages.id)
   RETURNING name, url, status;
   ```
   ⚠ **They now HAVE components** (they were built), so that guard will match nothing
   — which is correct and deliberate: it refuses to archive a page with content. Any
   archive from here is a knowing decision about built pages, not a cleanup.

## THE LESSON THAT COST THIS — put it where the next re-fire happens

**Run the page-identity md5 IMMEDIATELY after the plan lands, not after the fire.**
`checkpoint_postplan.sh` step 1 (Q2 invention check) and step 6 (identity md5) exist
for exactly this and would have caught it inside a minute: the hash moved from the
script's pre-fire `fd2c09c2…` to `da6908df…`. I ran it ~35 minutes late, by which time
the claim gate had re-opened and the builds were away. **And the cheaper check is the
item keys**: `needs_page:can-i-overpay` has no `guide-` prefix, and I read that list
twice without seeing it because the 08-15 notes had told me to expect "guide churn".
An inherited framing reads like a measurement and is not one.

## Also outstanding

- **`guides-index` — still the only 404, unchanged.** Its rerender returned
  `needs_human_review`: *"no sections ready to build (empty spec sections)"*. It is in
  the plan as `section-index` but nothing composes sections for it. Its nav entry
  already exists in `site_nav_items` (primary, position 2), so building it restores
  the menu entry.
- **8 of 13 wave failures are `failed to get latest commit/base tree`** — the git
  deploy path, not the build. `page_rerender:index` is among them yet the page
  deployed fine and passes toolgolden, so these are late rerenders failing to
  republish, not lost work. 40 rerenders are still queued against the same path.
- **⚠ A single 404 during a rerender wave proves nothing.** `tool-damage-checker`
  read 404 once at 16:25 and 200 on three immediate retries — a mid-republish window.
  Re-sample before believing it.
- **`incomplete_group_tool:tool-credit-roadmap`** at `needs_human_review`: that page
  is `not_built` while its siblings are deployed, and the plan gave it
  `tool-credit-health-check` — another page's calculator, the second copy of that
  function. Content decision, untouched.
- **D-NAV — MOSTLY RESOLVED ITSELF, and my morning reading of it was WRONG.**
  I reported that the rebuild had dropped the calculators from the site's navigation.
  **It had not.** The framework put all 11 in the `utility` group — the footer — which
  is exactly what `classifyPagesForNav` does with a never-primary page that carries a
  nav flag. [MEASURED 16:40Z] `site_nav_items` holds 11 utility rows, the regenerated
  footer component contains all 11 hrefs, and a page deployed AFTER the chrome
  regenerated (`/guides/secured-vs-unsecured.html`, 16:21:45) serves a footer with all
  11 calculator links.
  **What fooled me:** I sampled the served footer on pages deployed 13:44:08, and the
  chrome regenerated at **13:47:45** — three minutes later. Pages deployed before that
  moment carry the old chrome until they re-render, and the rerenders are the ones
  failing on the git deploy path. **So a chrome change is invisible on any page that
  has not re-rendered since it — check `site_components.updated_at` against the page's
  `deployed_at` before concluding anything about served navigation.**
  What remains genuinely open and is smaller than I said: the **header** still shows
  only `Home`/`About` (+ a CTA), because tool pages are barred from PRIMARY nav by
  design and this site has no `tools-index` parent to represent them. So the change is
  header-dropdown → footer-list, not "gone". Whether that is acceptable is still your
  call; a `tools-index` page is the framework-shaped answer if not.
- **D3 / D4 / RE-LOCK judgement** — carried unchanged from `HANDOFF_2026-08-17`.

## Standing cautions

- **The tag is not the artefact, and neither is a pod restart.** Compare the pod's
  `imageID` digest against the local image's; two readings agreed on "v1.0.1305" today
  and disagreed on the binary. Then `git merge-base --is-ancestor <fix> <stamp>` —
  never grep the binary for the fix's own sha, which it never carries.
- Verify tool placement at `site_plan_sections`, never `pages.sections` (LOCK-008
  merges locked rows into the latter).
- The script's judge query `component_name LIKE 'tool-%'` returns 26 either way — it
  matches `tool-cta`/`tool-list`. Use the locked-function join in `HANDOFF_2026-08-17`.
- A hand-filed or un-parked work item must be `triaged`; the dispatcher cannot see
  `detected` and fails silently.
- Query runs BY CORRELATION, never `now()`-interval; planner rows purge in ~2 days.
- No dispatch within ~300s of a chassis pod (re)start; a roll kills in-flight runs.
- Baselines stamp from the DB clock. (Agreed to 2s today.)
- **Before any planner fire: read `bugs_open/243`'s state.** The claim gate can be
  shut, in which case your items sit `triaged` and every other signal reads green;
  and it can re-open mid-investigation, which is what turned a containable mistake
  into 14 live pages today.

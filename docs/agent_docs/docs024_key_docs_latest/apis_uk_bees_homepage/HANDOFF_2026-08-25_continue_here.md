# HANDOFF 2026-08-25 — continue here

**Supersedes `HANDOFF_2026-08-24_continue_here.md`.** That one is still accurate on the traps and
the build designs; this one corrects the state and adds the fleet tracking work.

> ## ▶ ONE-LINE STATE
> apis.uk is **finished**. GTM is **live across the estate**. Everything still open is either a
> **council-scoped code change** or a **decision only the owner can make** — nothing is half-applied
> and nothing is mid-flight. `[ALL FIGURES MEASURED 2026-08-25]`

> **⚠ CORRECTED 2026-08-25 ~17:30 BST (session "apis.uk"), on the owner's ruling and `analytics_gtm`'s
> CONTRIB of the same day — read this before §1–§4, three of which are now wrong or not ours:**
> - **Everything Google is `analytics_gtm`'s.** Owner, verbatim: *"section 4 has google in it which is
>   taken by another lane, please communicate to that lane that that is what they take and we will
>   take the rest here."* So §2.1–2.3, §4 and §4a are theirs — cold-start
>   `docs024_key_docs_latest/analytics_gtm/HANDOFF_2026-08-25_continue_here.md`. Told them in
>   `analytics_gtm/CONTRIB_2026-08-25_from_apis_uk_bees_homepage_owner_ruling_you_take_everything_google_we_keep_the_rest.md`.
> - **§3 bullet 1 (per-site id in `RenderFallbackHead`) is DROPPED — do not build it.** The seam
>   already exists and is live (STY-050: `site_specs` `site_config.analytics.gtm_container_id`, read by
>   `{{if .gtm_container_id}}` in every head template), and `RenderFallbackHead` runs only when the
>   head component fails to render. `sites.settings->>'analytics_container_id'` has **0** rows and must
>   stay unused. The structural half (new sites born untagged; third-party guarantee) is
>   `bugs_open/397` §6.2, theirs.
> - **§1 "GTM fleet in all 27 heads" was true for six hours.** The 08-24 backfill wrote
>   `site_components.rendered_html` and no key, so **12 sites — apis.uk among them — carry the tag in
>   the artefact only** and lose it on their next chrome render; agritec.uk did, 08-24 19:20:53
>   (`bugs_open/397`). The §1 "our own lock working" line is also over-read: the permanent locks guard
>   `page_components`, **not the head slot**. The falsifier stands for the sections, not for GTM.
> - **What THIS lane still owns:** the page; §3 bullets 2 and 3 (per-section subjects; image accuracy
>   A + C) — grepped 2026-08-25 across `docs024_key_docs_latest/`, `bugs_open/`, `features_open/`:
>   no other lane carries either (`bugfix_285` reads `PlannedSections` as diagnostic only;
>   `features_open/018` is a screenshot taste critic, "specified, not built", adjacent to C but not it);
>   and the two `deferred` `content_rewrite` items.
> - `039_REFERENCE_traffic_and_tracking.md` lives at **`docs024_key_docs_latest/039_…`**, not in this
>   directory — a bare `cat` here fails.

## 1. Verified state — measured just now, not remembered

| | |
|---|---|
| apis.uk | HTTP 200, 67,877 B · 6 sections via `illustrated-text-block` · 7 images · **no footer, no email** · GTM present |
| apis.uk protection | **7 `page_components` `lock_type='permanent'`**, `build_status='deployed'` |
| GTM fleet | ~~in all 27 `site_components.head` rows~~ **CORRECTED 08-25: artefact-only on 12 of them, incl. apis.uk — `bugs_open/397`**; **re-render queue 1,962 complete / 130 queued / 15 failed** |
| the 3 previously held-back sites | `remortgagecalculator.uk` ✅, `robot-hands.com` ✅ — **all covered now** |
| GA4 | **NOT published** — 0 `Set-Cookie`, no `G-` id on any page. Nothing is being recorded. |
| `tools.apis.uk` | 200 throughout. **DNS never touched.** |

**The 15 failures are guards doing their job, not damage** — 12 × `OWNED_PAGE_GUARD` (hand-built
tool pages, correctly refused), 1 × `assembled to nothing` (a pre-existing defect on
`idea.uk/tools/ab-test-calculator`, worth telling that lane), 1 × section-component floor, and
**1 × `overwrite: REFUSED for page "index"` on apis.uk — which is our own lock working.** That is
the falsifier for the locking decision: a rewrite was attempted and refused in production. **CORRECTED 08-25: for the SECTIONS — the lock guards `page_components`, not the head slot, so it does nothing for GTM (`397`).**

⚠ **apis.uk was `needs_rebuild` again when this handoff was written** — the fan-out re-queued it —
and was settled to `deployed`. **Check that field before and after anything you do here.**

## 2. Owner decisions — nothing below can be done for him

> **MOVED 2026-08-25 to `analytics_gtm`** (owner ruling above). All three are Google; their current
> state is `analytics_gtm/HANDOFF_2026-08-25_continue_here.md` §2. Kept here only as history.

1. **Publish the GA4 tag, or not.** ⚠ **This is a change of compliance position, not a
   continuation.** Measured: five sampled sites set **zero cookies** today, and the container sets
   none because it has **0 tags**. Publishing starts setting `_ga` on ~24 sites at once. No site has
   a consent banner. **The Cloudflare route carries no such obligation** — see `039` §4a.
   *If he proceeds:* Tag type **Google Tag** (NOT "GA4 Event"), Measurement ID from
   **Admin → Data streams** (not the property number), trigger **All Pages**, then **Submit →
   Publish**. No backfill: history starts at publication.
2. **Search Console** — needs **one owner action**: a Google Cloud service account with the Site
   Verification and Search Console APIs enabled. Everything after that can be automated
   (`039` §5). Until then we cannot answer any question about search queries or position.
3. **The fleet dashboard script** — offered, not built. Would be the `039` §2 query batched 8 zones
   at a time with **our own tooling as its own visible line**. Say the word.

## 3. Builds designed and approved, not yet started

All three are **council-scope** (they touch `platform/`) and ship with a chassis roll.

- ~~**Per-site analytics id.** Third-party sites need their own tag or none. Read
  `sites.settings->>'analytics_container_id'` in **`RenderFallbackHead`** and emit the snippet only
  when non-empty. **Empty ⇒ no tag. Never hardcode our container.** Falsifier: a site with the key
  unset renders **zero** `googletagmanager` occurrences; one with it set renders exactly one.
  ⚠ **Do this before the next third-party build**, or someone bakes GTM into that function and every
  third-party site silently reports into our container.~~
  **SUPERSEDED 2026-08-25 — see the correction block at the top.** The seam is STY-050 and it is
  `analytics_gtm`'s; the intent (empty ⇒ no tag, nothing hardcoded) is already satisfied by it.
- **Per-section subjects.** `pages.sections` is `[]string` (`PlannedSections`), so every slot gets an
  identical brief — measured four times; one `content_rewrite` rewrote **all six** sections about
  the same subject. Let an entry be a string **or** `{"component":…,"subject":…}` and thread it to
  the writer. Unblocks the two **`deferred`** `content_rewrite` items (swarm, pollination) and the
  three generated-but-unused illustrations.
- **Image accuracy A + C.** **D is done.** ⚠ `imageryStyleGuide` is a **typed struct** — a new key is
  dropped on unmarshal with no error, so accuracy went into `avoid` (routed to the **negative
  prompt**, a separate channel that costs the length-limited main prompt nothing) and the **front**
  of `kinds.illustration.medium`. **C is CONFIG, not code**: `execute_vision_prompt` is a registered
  live action and `tool-acceptance-agent` already uses it in production; `visual-design-auditor` is
  text-only today, so it is a step to add, not a provider to write.

## 4. Traffic and tracking → `docs024_key_docs_latest/039_REFERENCE_traffic_and_tracking.md` — **`analytics_gtm`'s since 2026-08-25**

Written this session; read it before quoting any traffic number. The two things that matter most:

- **Our own `curl`/headless is a large slice of "traffic"** — **27.1%** on apis.uk vs **2.4%** on
  noted.co.uk, same query, same window. **The difference is which site had a session pointed at it.**
  Never quote a raw Cloudflare page-view figure.
- **Cloudflare and GA4 have opposite blind spots**, so a gap between them is the design, not a bug.
  Cloudflare is the only source with **history** and the only one that sees **crawlers**
  (noted.co.uk: GoogleBot 25/7d — thin, and the number that answers "are we being found").

## 4a. OWNER REQUEST 2026-08-25 — SPUN OUT TO A NEW LANE (not apis.uk work)

> **The lane exists: `docs024_key_docs_latest/analytics_gtm/` (session "google"), cold-start
> `HANDOFF_2026-08-25_continue_here.md`.** The walkthrough below is theirs to maintain now.

Owner, verbatim:

> "Please walk me through setting up the google tags in baby steps. I want to set it up under
> agent chassis and not idea.uk which is what I've half mistakenly done already."

**This is fleet tracking, not this lane** — recorded here because the request was made here, and
kept short deliberately. **Background and constraints: `039_REFERENCE_traffic_and_tracking.md`**
(§3 setup, §4 blind spots, §4a the consent decision, which is still open and should be settled
before publishing).

**Reassurance to lead with, because it is the anxious part:** the container being *named* `idea.uk`
does **not** send data to the idea.uk property. **The only thing that decides the destination is
the Measurement ID inside the tag.** Nothing needs recreating — point the tag at Agent Chassis and
the container name becomes cosmetic. (Rename it later under Admin → Container Settings if desired.)

Known account facts from the owner's own screenshots, 2026-08-24:
- GTM account `6368906206`, container `259867186` = `GTM-PQ3WCTBD`, displayed under `idea.uk`
- GA4 account `gqls` (`182167951`); target property **`Agent Chassis` (254005775)**
- Container Version 2 published 21:30 with **0 tags / 0 triggers** — nothing recorded, no backfill
- An `Untitled Tag` was started with type **`Google Analytics: GA4 Event`** — **wrong type**, it
  needs an Event Name and does not send page views. Change it to **Google Tag** or delete it.

### The walkthrough, in the order to do it

**A — get the Agent Chassis Measurement ID (this is the step that fixes the idea.uk mistake)**
1. `analytics.google.com` → **Admin** (gear, bottom-left)
2. In the **Property** column, confirm it reads **Agent Chassis**. If not, click the property
   selector and choose it. ⚠ *This is the whole ballgame — every later step inherits it.*
3. **Data streams** → click the web stream (if the list is empty: **Add stream → Web**, enter a
   domain such as `apis.uk` and a stream name, **Create**)
4. **Measurement ID**, top-right, format `G-XXXXXXXXXX` → copy it

**B — put it in the container**
5. `tagmanager.google.com` → container `GTM-PQ3WCTBD` → **Workspace → Tags**
6. Open the existing `Untitled Tag`, or **New**
7. **Tag Configuration** → pencil → choose **Google Tag** *(older UI: "Google Analytics: GA4
   Configuration")* — **not** GA4 Event
8. Paste the `G-…` into **Tag ID** *(labelled "Measurement ID" in the older UI)*
9. **Triggering** → **All Pages**
10. Name it something that says where it points, e.g. `GA4 — Agent Chassis — all sites`
11. **Save**

**C — publish (the step people miss)**
12. **Submit** (top right) → name the version, e.g. `GA4 base tag` → **Publish**
13. ⚠ **A saved tag in an unpublished container does nothing.** Version 2 having 0 tags is exactly
    this state.

**D — prove it works**
14. Visit any site, e.g. `https://apis.uk/`
15. GA4 → **Reports → Realtime** → you should appear within seconds
16. If Realtime stays empty: GTM **Preview** shows which tags fired on a page load

**E — tidy the half-done idea.uk work**
17. In **Tags**, check for any GA4 tag carrying a **different** `G-` id (an idea.uk one) and delete
    or repoint it — two GA4 tags on All Pages would double-count every page view
18. Leave the idea.uk **property** alone; it simply stops receiving from this container

**Afterwards:** all sites share one container, so in GA4 break reports down by **Hostname** or it is
one merged number. Per-site properties need a lookup-table variable keyed on hostname — materially
more work, and splitting later is easy.

## 5. Traps — the expensive ones, all paid for this session

- **`build_status='needs_rebuild'` is queue membership.** A sweep discarded verified hand-edits
  **4 minutes** after a green check. Settle → edit → render → **settle again** (a render re-queues).
- **The renderer reads stored `rendered_html`.** Cleaning `content_data` alone re-renders identical
  bytes. Assert both columns.
- **`COMPLETED` is the commit, not the deploy.** A cache-buster cannot help — it is a pipeline stage,
  not a cache. Compare served bytes to `git show <sha>:<domain>/index.html | wc -c`.
- **An image as markup in `content_data.content` is prose the writer will overwrite** — 25
  `page_divergence_overwritten` rows proved it. Use CLC-030's fields **and** lock.
- **A prohibition in a brief has no detector.** `roadmap_brief` forbade `A page about bees`; it
  served as the `<h1>` for two days. Add the `banned_claims` pattern in the **same edit**.
- **`kubectl exec -i` inside a `while read` loop eats the loop's input** — a 24-site roll-out did
  **1** and exited 0. Assert the iteration **count**, and use `mapfile` into an array.
- **`git commit` without a pathspec** — even `--allow-empty` — sweeps another lane's staged work.
  `--allow-empty` *permits* an empty commit, it does not *make* one.
- **Never add a wildcard worker route `*.apis.uk/*`** — it would swallow the live `tools.apis.uk`.

## 6. Cross-lane

- `dartsonline_traffic` (`agentchassis-51`) owns the Cloudflare method; the self-traffic landmine is
  jointly credited. **Fleet traffic question was handed to this lane.**
- `web_admin_console` was corrected twice by this lane (a stale "stalled build" claim, and a
  www→apex 301 attributed to a redirect rule when it is the worker's own branch).
- ⚠ Shared ledgers (`LANDMINES.md`, `WRONG_CALLS.md`) are append-only and the pattern check flags
  in-place edits. **Expect to prove a removed line was your own** — `git show` it.

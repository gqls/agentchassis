# HANDOFF 2026-08-25 — continue here

**Supersedes `HANDOFF_2026-08-24_continue_here.md`.** That one is still accurate on the traps and
the build designs; this one corrects the state and adds the fleet tracking work.

> ## ▶ ONE-LINE STATE
> apis.uk is **finished**. GTM is **live across the estate**. Everything still open is either a
> **council-scoped code change** or a **decision only the owner can make** — nothing is half-applied
> and nothing is mid-flight. `[ALL FIGURES MEASURED 2026-08-25]`

## 1. Verified state — measured just now, not remembered

| | |
|---|---|
| apis.uk | HTTP 200, 67,877 B · 6 sections via `illustrated-text-block` · 7 images · **no footer, no email** · GTM present |
| apis.uk protection | **7 `page_components` `lock_type='permanent'`**, `build_status='deployed'` |
| GTM fleet | in all 27 `site_components.head` rows; **re-render queue 1,962 complete / 130 queued / 15 failed** |
| the 3 previously held-back sites | `remortgagecalculator.uk` ✅, `robot-hands.com` ✅ — **all covered now** |
| GA4 | **NOT published** — 0 `Set-Cookie`, no `G-` id on any page. Nothing is being recorded. |
| `tools.apis.uk` | 200 throughout. **DNS never touched.** |

**The 15 failures are guards doing their job, not damage** — 12 × `OWNED_PAGE_GUARD` (hand-built
tool pages, correctly refused), 1 × `assembled to nothing` (a pre-existing defect on
`idea.uk/tools/ab-test-calculator`, worth telling that lane), 1 × section-component floor, and
**1 × `overwrite: REFUSED for page "index"` on apis.uk — which is our own lock working.** That is
the falsifier for the locking decision: a rewrite was attempted and refused in production.

⚠ **apis.uk was `needs_rebuild` again when this handoff was written** — the fan-out re-queued it —
and was settled to `deployed`. **Check that field before and after anything you do here.**

## 2. Owner decisions — nothing below can be done for him

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

- **Per-site analytics id.** Third-party sites need their own tag or none. Read
  `sites.settings->>'analytics_container_id'` in **`RenderFallbackHead`** and emit the snippet only
  when non-empty. **Empty ⇒ no tag. Never hardcode our container.** Falsifier: a site with the key
  unset renders **zero** `googletagmanager` occurrences; one with it set renders exactly one.
  ⚠ **Do this before the next third-party build**, or someone bakes GTM into that function and every
  third-party site silently reports into our container.
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

## 4. Traffic and tracking → `039_REFERENCE_traffic_and_tracking.md`

Written this session; read it before quoting any traffic number. The two things that matter most:

- **Our own `curl`/headless is a large slice of "traffic"** — **27.1%** on apis.uk vs **2.4%** on
  noted.co.uk, same query, same window. **The difference is which site had a session pointed at it.**
  Never quote a raw Cloudflare page-view figure.
- **Cloudflare and GA4 have opposite blind spots**, so a gap between them is the design, not a bug.
  Cloudflare is the only source with **history** and the only one that sees **crawlers**
  (noted.co.uk: GoogleBot 25/7d — thin, and the number that answers "are we being found").

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

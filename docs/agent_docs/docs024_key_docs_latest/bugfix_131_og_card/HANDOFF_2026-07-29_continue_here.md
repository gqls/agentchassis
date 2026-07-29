# HANDOFF — bugs_open/131 (og-card), continue from here — 29 July 2026

> **SUPERSEDED IN PART, 2026-07-29 afternoon (session relojistas-5). Read
> `SUMMARY_2026-07-29b_og_card.md` FIRST — this file's §0, §2 and §4 describe a state that no
> longer holds.** What changed: the owner answered the §2 blocker (storage writes go *through
> the chassis*; relojistas is done and its header is LIVE), gaswholesalers and idea.uk were
> **reassigned to the "relojistas 4" session** (see `COORDINATION_2026-07-29_who_does_what.md`
> — do not work them from here), and §4's items 1 and 3 are **done and live on v1.0.1199**
> (favicon aspect + lock-before-commit, council-approved on trail `bfd73f71`).
> **§4 item 3's figure was wrong**: the favicon distortion affects **5** sites, not "every wide
> logo in the estate" — measured, see the SUMMARY. §4 items 2, 4, 5, 6, 7 still stand, and §7's
> landmines still stand **except** the leopardess one, which is now handled deliberately (its
> locked rows are backfilled and armed). One landmine to ADD: **`sites.github_repo` selects
> which deploy repo serves a site, and publishing to the wrong one succeeds silently** —
> relojistas and idea.uk are `vm-sites`, gaswholesalers and leopardess are B2.

Repo-root `CLAUDE.md` binds. **`131` is one of the documented ambiguous numbers** — the other is
the vonc gauntlet usability audit, owned by another lane. **Resolve by slug; `git log` the file
PATH, never the number.**

Case file: `bugs_open/131_HANDOFF_2026-07-28_og_image_points_at_a_card_that_was_never_generated.md`
Workstream: `docs/agent_docs/docs024_key_docs_latest/bugfix_131_og_card/` (PLAN · NOTES ·
README_where_we_are · SUMMARY_2026-07-29). Read the SUMMARY first, then NOTES entry (3).

Commits this lane: `9a90c8446` (investigation + pilot), `2a6e23afa` (rollout + the correction),
`cab4443b3` (WRONG_CALLS), plus the docs commit carrying this file.

---

## 0. Sixty-second orientation

The bug was "11 of 14 sites advertise a social preview image that 404s". **It is now 8 sites
working and done, 3 blocked on one owner decision, 1 site (webdesign.co.uk) never had the tag.**

The fix needed **no code**. `derive_brand_head_assets` was already built, registered and live
since 2026-07-11; nobody had run it. Queue a work item and it takes ~18 seconds per site.

| | state |
|---|---|
| cards serving 200 | **12 of 13** with an `og:image` tag (was 2) — re-verified after v1.0.1196 |
| good cards | 8 — vetcomparison.uk, oufe.com, fundamentallyai.com, ai-agent-orchestration.com, dartsonline.com, finetuning.uk, gamesdesign.co.uk, vonc.com |
| **wrong cards** | **relojistas.com** (two-up spec sheet), **gaswholesalers.com** (3×3 contact sheet of 9 concepts, garbled text) |
| **failed** | **idea.uk** — cannot derive; diagnosed, see §3 |
| no tag at all | webdesign.co.uk — a ported site with its own head; out of scope so far |
| **CRITICAL PATH** | **owner decision on S3 credentials** (§2) — unblocks all three bad sites at once |

---

## 1. The one thing to carry forward above all else

**Every mechanical signal can be green while the artefact is wrong.** The relojistas card came
back `complete`, HTTP 200, a valid 1200×630 `image/png`, with provenance rows written — and the
picture was a brand *specification sheet*. It took `Read`-ing the PNG to see it.

**So: for an image artefact, dimensions and MIME type are the equivalent of a status code, not
the artefact. Look at the picture.** This is `bugs_open/012`'s rule in a medium where every
automated check available passes.

The same discipline caught the second one — gaswholesalers' nine-up contact sheet would have
shipped silently as "another success".

**And it has a metadata twin, which I got wrong in this very session.** I reported the
`purpose='hero'` vs `purpose='logo'` split (10 vs 4) as a predictor of card quality. It is not:
gaswholesalers is `purpose='logo'` and produced the worst card in the set, and the two
spec-sheet cases sit on opposite sides of the split. `purpose` controls deployed *geometry*
(`ImagePurposes`: logo → 400×400 png, hero → 1600×900 jpg), not whether the picture is a logo.
**Do not infer a content property from a geometry field.**

---

## 2. THE BLOCKER — an owner decision, nothing proceeds without it

relojistas, gaswholesalers and idea.uk all need a corrected `logo` asset **written to S3**,
because `derive_brand_head_assets` reads the logo from S3 by key.

**Reading `personae-storage-secrets` was refused by the permission classifier. Do not work
around it.** The owner was given three options and **has not yet answered**:

1. owner runs `! kubectl -n ai-persona-system get secret personae-storage-secrets -o jsonpath='{.data}'`
   so the output lands in session;
2. owner adds a Bash permission rule for that secret;
3. commit corrected files straight into the `gqls/sites` deploy repo (the route
   `scripts/webdesign_publish_assets.sh` established). **Fixes header + card today, but leaves a
   trap:** the S3 logo stays a spec sheet, so the next `derive_brand_head_assets` run overwrites
   the good card with a bad one. Only take this with the trap recorded loudly.

**The relojistas crop is already done and verified**, in the session scratchpad (regenerate with
the scripts if lost — they are simple): trimmed to the wordmark's ink bounds (94,305)-(609,449)
with 45%-of-height padding → 646×275, **background knocked out** (house practice — leopardess's
logo was done this way, `docs/leopardessconsulting/RUNBOOK.md` H4). Composed through
`composeOGCard`'s exact maths it gives a clean card with **no letterbox**.

---

## 3. idea.uk — diagnosed, and the diagnosis corrects my own first reading

Derivation fails `NoSuchKey` after 3 attempts. My first inference — "points at an S3 key that no
longer exists" — was **wrong**. The row holds:

```
logo | hero | active | stability/... | url = /assets/images/logo.jpg
```

**That is a deployed WEB PATH, not an S3 URI.** `ExtractKeyFromS3URI(presignedURLToS3URI(url))`
derives nonsense from it and S3 answers `NoSuchKey`. The object was never missing.

Bounded fleet-wide — **2 of 14** rows are web paths: `idea.uk` and `leopardessconsulting.co.uk`.

> **LANDMINE, and it is the inverse of the bug.** leopardess is in that same malformed state,
> and **that is the only reason its owner-approved hand-made card has never been overwritten by
> a derivation.** Its protection is an accident of a bad row, not a lock. Anyone who "tidies"
> that row without reading this will destroy an approved brand artefact on the next run.
> If you fix idea.uk's row, **do not fix leopardess's** — lock it instead (`assets.locked_at`,
> which `recordDerivedAsset` already honours).

Note `recordDerivedAsset` writes web paths into `assets.url` today for `favicon`/`og_card`.
Harmless now because nothing reads those bytes back, but the column carries two incompatible
conventions with nothing distinguishing them.

---

## 4. What is NOT done, ranked

1. **The S3 decision (§2)** — unblocks 3 sites at once. Everything else is downstream.
2. **Fix 1, the code gate** — do not emit `og:image`/`twitter:image` unless the card exists.
   Still worth landing as the structural guard; idea.uk is live proof the bad state recurs.
   - **Precedent to follow, in the same function:** `render_site_components_action.go:704-712`
     already gates `sprites.css` on an active asset "otherwise the `<link>` would 404 on sites
     without one" — while `:700` waves the og card away as "harmless if they 404". Adjacent
     lines, opposite answers. That sentence *is* the bug.
   - **LANDMINE: do not key the gate on an `og_card` assets row.** leopardess serves a working
     card and has **no row** — a row-gate suppresses the tag on the one site that always worked.
     If you backfill its row first to make a row-gate sound, see the §3 landmine before touching
     leopardess's asset rows at all.
   - Honest caveat for the council submission: a row is **necessary but not sufficient** — it
     records that an asset was derived, not that it deployed. That is exactly what the
     `undeployed_asset` detector exists for.
   - Cost: council + build + roll + chrome re-render on 14 sites (head is a stored artefact —
     `bugs_open/117`) + page redeploy. Takes effect gradually as chrome re-renders; no need to
     force them.
3. **Favicon derivation distorts non-square logos.** `resize.Resize(faviconSize, faviconSize,
   ...)` is non-proportional, so wordmarks are squashed illegible. `composeOGCard` does it right
   with `resize.Thumbnail`. Affects every wide logo in the estate. Verified by composing
   relojistas' corrected crop locally — still illegible at 64×64.
4. **Letterbox rectangles on almost every card** — logo assets are opaque, so their own
   background square is painted onto the brand-colour card (vonc: magenta on near-black;
   vetcomparison: white on pale grey). An *asset* fix, not a code fix: knock the background out.
5. **`og:title` is the bare domain on ~8 sites, `og:description` absent on relojistas** — the
   case file's "second defect", untouched.
6. **webdesign.co.uk emits no `og:image` at all** — `injectBrandHeadTags` skips a head that
   already has `rel="icon"` or `og:image`. Never investigated.
7. **The `undeployed_asset` detector is blind to this whole bug class** — it fired 5× for
   `og_card`, all on robot-hands (the site that worked), never on the 11 broken ones, because
   its denominator is the `assets` table and they had no row. Not filed as its own bug yet.

---

## 5. How to run a derivation (the whole recipe, it is short)

```sql
INSERT INTO site_work_items (site_id, source, item_type, severity, summary, spec,
       handler_agent, status, created_by, priority, pipeline, item_key, triaged_at)
SELECT id, 'discovery', 'needs_brand_head_assets', 'medium', '<summary>',
       '{"mode": "brand_head"}'::jsonb, 'asset-deployer', 'triaged', '<who>', 70,
       'build', 'needs_brand_head_assets:og_card', now()
  FROM sites WHERE domain = '<domain>';
```

**`status='triaged'` AND `pipeline='build'` are both load-bearing** — the trigger's predicate
requires them. An item left at the default `status='detected'` is **never dispatched**: that is
why robot-hands' `needs_sprite_css` has sat there since 2026-07-24 with `attempt_count=0`. It is
not "tried and failed", it was never tried.

Dispatch is **one site per 120s tick**, so 10 sites took ~30 minutes of wall-clock for ~3 minutes
of work. Routing is generic (`agent_type_field: current_item.handler_agent`) — **no item-type
allow-list to register.**

**Verify — never by asserting the tag exists, that is the bug:**

```bash
img=$(curl -s https://<site>/ | grep -o 'property="og:image" content="[^"]*"' | sed 's/.*content="//;s/"//')
curl -s -o card.png -w "%{http_code}\n" "$img"
file card.png          # must say PNG 1200 x 630 — a 404 page saves happily as card.png
```

**Then `Read` card.png.** All of the above passed on both of the wrong cards.

---

## 6. CLOSED by the owner (29 July) — branded image for dead forum hotlinks

> **OWNER DECISION, 2026-07-29: "we can leave the 410s." The idea is DROPPED — do not reopen
> it or re-derive the analysis.** `/attachment.php` and the rest of the dead vBulletin surface
> keep returning `410 Gone`. The reasoning below is kept only so nobody re-costs it from
> scratch.

`foroderelojes.es`, a live Spanish watch forum, hotlinks `/attachment.php` (149 requests in the
measurement window); we now serve **410 Gone**, so those old threads show broken images. Owner
asked about serving a branded watch image instead, and whether it could go out *with* the 4xx.

Answered in chat, recorded here so it is not re-derived:

- **You cannot have both on one request.** Browsers do not render an image body returned with a
  non-2xx status; `<img>` treats it as a load failure. A 410-with-a-picture is a 410 plus wasted
  bandwidth. (401 is worse — it also wants `WWW-Authenticate` and may prompt for credentials.)
- **You can have both, split on `Referer`**: external forum referer → `200` + branded image;
  everything else → `410`. Serving 200 to all would resurrect ~184k/day of scraper surface as
  "alive" and undo what the 410 was for.
- **Not an SEO penalty risk.** Cloaking is defined on **user-agent/IP**; varying on **`Referer`**
  is ordinary hotlink protection. Caveat: `Referer` is not always sent, so some embeds stay
  broken.
- **The real risk is the forum, not Google** — it is unsolicited brand placement on someone
  else's pages, in the enthusiast community whose goodwill this domain most needs. Owner call.
- Volume is small: a brand-impression play, not a traffic tactic. And it slightly softens the
  evidence doc's line that hotlinks "cannot be converted by serving content" — an image cannot
  be *clicked* but it can be *seen*. Prefer an evergreen image; these are permanent old threads
  and a "latest watches" lineup dates badly.

---

## 7. Landmines paid for in this lane

- **Look at image artefacts.** §1. Two wrong cards, both green on every automated signal.
- **Do not infer content from metadata.** §1. `purpose` predicts geometry, not picture content.
- **`status='triaged'` + `pipeline='build'`** or the item is never dispatched. §5.
- **Do not gate the og:image tag on an `og_card` row** — regresses leopardess. §4.2.
- **Do not "fix" leopardess's malformed logo row** — it is what protects its approved card. §3.
- **A 404 page saves happily as `card.png`** — `file` it, do not trust the download succeeding.
- **The seed is not the system.** `asset-deployer`'s `brand_head` branch was verified on the
  live agent row; `SQL_2026-07-11_asset_deployer_brand_head_mode.sql` is history.
- **Storage credentials are classifier-blocked.** Do not attempt to route around it.

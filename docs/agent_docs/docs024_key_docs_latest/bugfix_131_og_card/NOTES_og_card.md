# NOTES — bugs_open/131 (og-card slug)

Append-only, newest at the bottom. Technical log: evidence, commands, what the system actually
said, and every misstep.

---

## 2026-07-28 (1) — reproduced, and the generator turns out to already exist

**Reproduced live on the wire, 21:2xZ**, all 14 domains with deployed pages — the case file's
own reproduce loop:

```
ai-agent-orchestration.com   404      leopardessconsulting.co.uk   200
dartsonline.com              404      robot-hands.com              200
finetuning.uk                404      webdesign.co.uk              NO-TAG
fundamentallyai.com          404
gamesdesign.co.uk            404
gaswholesalers.com           404
idea.uk                      404
oufe.com                     404
relojistas.com               404
vetcomparison.uk             404
vonc.com                     404
```

11 × 404, 2 × 200, 1 × no tag. Matches the case file exactly, so the measurement held overnight
and across the day's ~8 rolls.

### The case file's `[UNDIAGNOSED]` is answerable from the tree

The case file says of the two passing sites: *"**[UNDIAGNOSED]** which path produced those two;
find it before building a new generator, because it may only need wiring."* Correct instinct —
and the answer was one grep away:

- `platform/orchestration/actions/derive_brand_head_assets_action.go` — fetches the site's
  active `logo` from S3, resizes to a 64px favicon, composes it centred on a brand-colour
  1200×630 card, commits both to the site repo, and writes provenance rows.
- Registered at `registry.go:185`.
- Reachable in production: `asset-deployer` has a live `check_mode` branch —
  **verified on the live row, not the seed** (`SQL_2026-07-11_asset_deployer_brand_head_mode.sql`
  is history):

```
start_step   = check_mode
has_derive   = t
condition    = input_data.spec.mode == "brand_head" OR input_data.mode == "brand_head"
site_param   = input_data.site_id
```

**So nothing needs building.** "Nothing generates og-card.png" (case file) is true as a
statement about what has *run*, and false as a statement about what *exists*. The generator is
built, registered, live, and has been since 2026-07-11.

### The measurement that reorders the fix list

```sql
SELECT si.domain,
       (SELECT count(*) FROM pages p WHERE p.site_id=si.id AND p.deployed_at IS NOT NULL) AS deployed,
       COALESCE((SELECT string_agg(a.asset_key||'='||a.status, ',' ORDER BY a.asset_key)
                 FROM assets a WHERE a.site_id=si.id
                  AND a.asset_key IN ('logo','favicon','og_card')), '-') AS brand
  FROM sites si ORDER BY si.domain;
```

Two results decide the plan:

1. **All 14 live sites have an active `logo`.** That is the generator's only precondition, so
   fix 2 is available on every site today, with no code change.
2. **Only robot-hands has an `og_card` row.** leopardess serves a 200 card with *no* row — it
   was hand-committed from the owner-approved logo (`docs/leopardessconsulting/RUNBOOK.md` H4).

(2) kills the obvious design for fix 1. Gating the tag on "an `og_card` asset row exists" would
**suppress the tag on leopardess — the one site whose preview actually works.** Writing that
down because it is the design I would have reached for first, and the query is what stopped me.

### The precedent for fix 1 is already in the same function

`render_site_components_action.go:704-712`, two lines below the og-card comment:

```go
// Phase I2: only link sprites.css when the site actually has an active
// sprite-sheet asset — otherwise the <link> would 404 on sites without one.
var spriteCount int
_ = db.QueryRowContext(ctx, `SELECT count(*) FROM assets
     WHERE site_id = $1 AND purpose = 'sprite_sheet' AND status = 'active'`, siteID).Scan(&spriteCount)
renderedHTML = injectBrandHeadTags(renderedHTML, renderCtx, spriteCount > 0, logger)
```

And at `:700`, about the og card, the opposite conclusion: *"harmless if they 404 until
derivation runs."* Same question, same file, adjacent lines, two different answers. It is not
harmless — that sentence **is** the bug.

### Second defect, separate and not mine to fix here

`undeployed_asset` has fired 5 times for `og_card` — **all five on robot-hands.com, the site
that works** (`detected` 07-24; `unresolved` ×3 07-18, ×1 07-19). Never once on the 11 broken
sites. The detector reads the `assets` table, so a site that never generated a card has no row
to be "undeployed" and is invisible to it. **A detector whose denominator is the artefact
table cannot see a missing artefact.** Same family as the fleet's "check with no failing
branch", one layer along. Recorded in PLAN; not fixed here.

### Dispatch contract — read, not guessed

Trigger and dispatcher disagree (that is `bugs_open/029`, owned by another lane — I did not
touch it), but for queuing purposes the binding facts are:

- trigger `build-pipeline-trigger` (enabled, 120s) fires on `status='triaged'` **AND
  `pipeline='build'`** AND `attempt_count < max_attempts` AND site not locked;
- `find_dispatchable_site` takes `status IN ('triaged','approved')`, no `claimed` item on the
  site, `ORDER BY priority ASC LIMIT 1` — **one site per tick**;
- `spawn_handler` is generic: `agent_type_field: current_item.handler_agent`. **No item-type
  allow-list to register** — any item naming a valid agent routes.

So an item inserted at the default `status='detected'` is never dispatched. That is why
robot-hands' `needs_sprite_css` has sat at `detected` since 2026-07-24 with `attempt_count=0`
— it was never triaged, not "tried and failed".

**Queue was completely empty before I inserted** (0 triaged / 0 approved / 0 claimed
fleet-wide), so the pilot starves nobody. Chassis pods 42 min old, past the ~300s
dispatch-drop window.

### Pilot fired

```sql
INSERT INTO site_work_items (site_id, source, item_type, severity, summary, spec,
       handler_agent, status, created_by, priority, pipeline, item_key, triaged_at)
SELECT id, 'discovery', 'needs_brand_head_assets', 'medium', '<summary>',
       '{"mode": "brand_head"}'::jsonb, 'asset-deployer', 'triaged', 'bugs_open/131', 70,
       'build', 'needs_brand_head_assets:og_card', now()
  FROM sites WHERE domain = 'relojistas.com';
```

→ `7867a232-bc03-422b-9c1c-a4ab9eac9e78`, claimed by `build-dispatch-loop` at 21:30:44Z,
~40s after insert.

---

## 2026-07-28 (2) — the pilot WORKED, and looking at the output found a bigger defect

**Mechanically a clean pass.** Item `7867a232` complete at 21:31:02Z — 18 seconds from claim.
`check_mode.condition_met = true`, routed to `derive_head_assets`. On the wire immediately:

```
https://relojistas.com/assets/images/og-card.png  →  200, image/png, 129182 bytes
file: PNG image data, 1200 x 630, 8-bit/color RGB
assets rows: favicon + og_card, origin_type=generated, origin_model=derived-from-logo, 21:31:01Z
```

So fix 2 needs no code, no roll, no chrome re-render — confirmed, not predicted. The tag that
was already on the page started resolving the moment the asset deployed.

### Then I looked at the image, and that is the whole point of this entry

**The card is a picture of a brand SPEC SHEET, not a logo.** It shows the "Relojistas" wordmark
*twice*, side by side, on a light swatch and a dark swatch — a two-variant presentation board.
Mechanically flawless (1200×630, on-brand cream ground, centred, correct bytes) and wrong.

The favicon derived in the same run is the same sheet crushed to 64×64 — **two ~28px wordmarks,
completely illegible**, now live as the browser-tab icon.

**Nothing in the checking I had already done could have caught this.** Status said `complete`.
The URL said 200. `file` said a valid 1200×630 PNG. The provenance rows were written. Every
signal I had was green, and the artefact was wrong. It took `Read`-ing the PNG.

> This is `bugs_open/012`'s lesson in a new medium: *check the artefact's structure after a
> rewrite, not just the status.* For an IMAGE, "structure" is not dimensions or MIME type —
> **it is what the picture shows, and the only way to know is to look at it.** A card is
> outward-facing brand surface; a green pipeline says nothing about whether it embarrasses you.

### Root cause, and it is bigger than the og card

The generator is faultless — it does exactly what it says, centring the site's active `logo`
asset on its brand colour. **relojistas' `logo` asset simply is not a logo.** It is a
1408×768 two-up presentation board (`origin_model = banana/gemini-3-pro-image-preview`,
2026-07-16).

And the same asset is what the **live site header** serves:

```
<img src="/assets/images/logo.jpg" alt="relojistas.com" class="logo-img">
.logo-img { max-height: 44px; width: auto; }     ← styles.css, the only stylesheet, NO crop
JPEG image data, 1408x768
```

1408/768 = 1.83, so at `max-height:44px` the header renders the **entire two-up board at about
81×44px** — two wordmarks, each ~40px wide, both illegible. **No `object-fit`, no `clip-path`,
no crop anywhere.** This predates my work by weeks and is on every page of a site the lane's
own handoff calls "finished and self-running". Every visitor sees it.

*What this says about my earlier verification of this site:* the discoverability audit checked
what crawlers RECEIVE — headers, tags, status codes, robots, sitemap. It never once looked at
what a HUMAN sees. Both are "verifying the live site", and they do not overlap at all.

### Not a fleet-wide pattern — it is per-site, so the rollout needs an eyeball each

```
ai-agent-orchestration.com  /assets/images/logo.png  400x400  → a proper square mark. Fine.
finetuning.uk               /assets/images/logo.png  400x400
robot-hands.com  (reference, card already live)  → proper mark + wordmark on a dark card. Good.
relojistas.com   /assets/images/logo.jpg 1408x768 → two-up spec sheet. BAD.
```

Only 2 of 12 even serve `/assets/images/logo.png`, so the deployed path is not a reliable
census — the authoritative input is the `assets` row, which is what the generator reads.

**Consequence for P2:** firing the remaining 11 would put 11 generated images on live sites
sight-unseen. Given 1 of the 2 I have inspected closely is wrong, that is not acceptable
without looking at each result. Rollout paused for an owner call.

### Did I make relojistas worse? No — but I did not make it good either

Before the run both files 404'd. The head lists a fallback icon (`logo.jpg`) after
`favicon.png`, and that fallback is the *same* spec sheet, so the favicon is no worse than it
was. The card is arguably better than nothing (a share now previews *something* legible-ish
rather than blank). But "not worse" is not the bar for brand surface. The real fix is the logo
asset, not the derivation.

---

## 2026-07-28 (3) — rolled to 10 sites, looked at every card: 8 good, 1 wrong, 1 failed

Queued the same item for the 10 remaining 404 sites at 21:49Z. **Drained in ~30 minutes**
(dispatch is one site per 120s tick, so the wall-clock is the tick rate, not the work — each
derivation itself takes ~18s).

Result: **9 complete, 1 failed.** On the wire, 10 of 11 now serve a valid 1200×630 PNG.

### Every card, looked at — not inferred from status

| site | what the card actually shows | verdict |
|---|---|---|
| vetcomparison.uk | cross + magnifier + "VetComparison.uk" | **best of the set** |
| oufe.com | "OUFE" cream on near-black | good |
| fundamentallyai.com | mark + "FundamentallyAI" | good content, very dark |
| ai-agent-orchestration.com | blue network mark, no name | acceptable |
| dartsonline.com | geometric dartboard, no name | acceptable |
| finetuning.uk | teal abstract mark, no name | acceptable |
| gamesdesign.co.uk | isometric cube, no name | acceptable |
| vonc.com | magenta star, no name | acceptable |
| **gaswholesalers.com** | **a 3×3 CONTACT SHEET of nine logo concepts**, with garbled AI text — "GAAS", "WALSE", "WHOLACS", "GS GAS" | **WRONG** |
| **relojistas.com** | two-up spec sheet (see entry 2) | **WRONG** |
| **idea.uk** | nothing — derivation FAILED | **404 still** |

### > **CORRECTED, same session — `purpose` does NOT predict card quality. I said it did.**

In entry (3)'s first draft and in chat I reported the fleet split — 10 sites store their logo
with `purpose='hero'` (AI-generated wide images), 4 with `purpose='logo'` — and presented it as
a predictor of which cards would come out well.

**gaswholesalers.com refutes it.** It is one of the four `purpose='logo'` sites, with no
`origin_model` at all, and it produced **the worst card in the set** — a nine-up contact sheet
of rejected logo concepts with hallucinated lettering.

What `purpose` actually controls is the **deployed geometry** (`ImagePurposes`: `logo` →
400×400 png, `hero` → 1600×900 jpg). It says nothing whatever about whether the *picture* is a
logo. I inferred a content property from a geometry field because the correlation held on the
four sites I had looked at. **The two spec-sheet cases sit on opposite sides of the split** —
relojistas is `hero`, gaswholesalers is `logo` — which is as clean a refutation as it gets.

*The check that would have caught it:* look at the artefact before generalising from the
metadata. Exactly the lesson of entry (2), applied one level up and missed anyway.

### idea.uk — a dangling asset pointer, separate defect

```
step derive_head_assets failed: download logo bytes: failed to download object from s3:
operation error S3: GetObject, StatusCode: 404, NoSuchKey
```

Retried to `attempt_count=3` and failed correctly. **idea.uk's `logo` asset row is active and
points at an S3 key that does not exist.** The row says the asset is there; the object is gone.
Nothing else in the platform notices, because nothing else reads the bytes — the deployed page
serves its own copy. Not fixed here.

### Systemic and cosmetic: almost every card has a visible letterbox rectangle

The logo assets are **opaque rectangles**, so `composeOGCard`'s `draw.Over` paints the logo's
own background square onto the brand-colour card. On most sites the two colours differ and the
box is plainly visible (vonc: magenta square on near-black; vetcomparison: white box on pale
grey; fundamentallyai: dark box on darker navy). robot-hands has it too — visible in the
original reference card.

Fixed by making the logo asset **transparent**, which is the established practice here:
leopardess's approved logo was "background-knocked-out" (`docs/leopardessconsulting/RUNBOOK.md`
H4). Demonstrated on the relojistas crop — knocked out, the composed card has no rectangle at
all. This is a *logo-asset* fix, not a code fix.

### Third defect: the favicon derivation distorts any non-square logo

```go
faviconPNG, err := encodePNG(resize.Resize(faviconSize, faviconSize, logoImg, resize.Lanczos3))
```

`resize.Resize(64, 64, ...)` is **non-proportional**. A square mark survives; a wordmark is
squashed to its aspect ratio. relojistas' corrected 646×275 crop still yields an illegible
favicon — compressed 2.35× horizontally. Verified by composing it locally with the same maths.
Affects every wide logo in the estate. `composeOGCard` gets this right (`resize.Thumbnail`
preserves aspect); the favicon path does not. Not fixed here.

### relojistas: crop done, install BLOCKED

Cropped the light variant to its ink bounds (94,305)-(609,449) with 45%-of-height padding →
646×275, background knocked out, verified by eye. Composed through `composeOGCard`'s exact
maths (1200×630, thumbnail to 420px longest edge, centred on `background` `#f9f8f5` — the
palette has no `header_bg`/`footer_bg`) — **clean, no letterbox.**

**Cannot install it.** `derive_brand_head_assets` reads the logo from S3 by key, so the
corrected asset must be written there, and reading `personae-storage-secrets` was refused by the
permission classifier. Not worked around. Awaiting an owner decision — credentials, or the
`gqls/sites` deploy-repo route (which fixes the header + card today but leaves the S3 logo
still a spec sheet, so the next derivation would overwrite the good card with a bad one).

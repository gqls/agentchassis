# HANDOFF — dartsonline.com. START HERE. Written 2026-08-31.

Supersedes `HANDOFF_2026-08-16_continue_here.md` (still accurate about the traps in its §5; its
§3 work is all done). Owning lane: `dartsonline_traffic`. Sibling lanes this one depends on:
`inline_guide_imagery` (design, unbuilt), `news_editorial` (owns `features_open/035`), and the
`apis.uk` lane (owns fleet traffic).

**Nothing is on fire.** Everything below is done-and-verified, queued, or a stated open question.

---

## 0. The one-paragraph version

The site was rejected by Webgains for insufficient traffic; that lane of work is now measurable
rather than speculative. Since 08-16 the site has gained a privacy page, a sitemap, 4+ new guides,
working card images, correct contrast, and a real traffic baseline. Four framework defects were
found and filed along the way, three of them fixed at the framework rather than on this site. The
two live threads are **imagery** (four bad heroes regenerating now; guides need per-section images
and the mechanism is another lane's, unbuilt) and **affiliates** (apply where the gate is not
traffic).

---

## 1. DONE and verified at the artefact

| thing | state | evidence |
|---|---|---|
| privacy page | **LIVE**, copy verbatim | 16/16 approved blocks matched, whole copy contiguous; owner removed the affiliate-independence sentence 08-20 |
| `/shipping-returns.html` | **RETIRED** | archived + retracted (`sites@2af7c17dd`), serves 404 |
| sitemap + robots | **LIVE** | `/sitemap.xml` 200; **37 URLs** as of 2026-08-31 (was 23 on 08-20) |
| stylesheet | **RESTORED** | was 164 bytes for 2 days (`bugs_open/198` clobber); recovered from git, seeded into `css_themes` so a patch run cannot re-clobber |
| contrast | **0 real failures** | all 23 pages, desktop *and* mobile. The 15 remaining rows are the probe's `overImage` placeholders, not measurements |
| card images | **12/12 on the homepage** | fixed at the framework by the `384` lane, not hand-patched here |
| privacy footer link | **23/23 served pages** | graded at the bytes, not at the item status |
| traffic | **measurable** | Cloudflare `Zone → Analytics → Read` live on 2 tokens |

---

## 2. The traffic baseline — and the trap inside it

**30 days to 2026-08-24: 5,631 page views, ~188/day.** But classified by client:

| class | share | /day |
|---|---|---|
| real browsers | 33.6% | 63 |
| **our own tooling** (curl + headless) | **28.5%** | 53.5 |
| unattributed | 36.2% | 68 |
| **search crawlers** | **1.5%** | **2.8** |

**⚠ Do not quote a rise without classifying it.** Total page views 2.6×'d across the window and
**none of it was growth** — human went 57.6 → 68.4/day (+19%) while our own tooling went 18.5 →
88.6 (4.8×). The `apis.uk` lane reproduced the contamination independently at 27.1% on their site,
with an unworked control site at 2.4% — so it is what an actively-worked site looks like, not a
dartsonline quirk. **Quote the pair, never the percentage alone.**

**The number that actually matters: GoogleBot made 54 page views in 30 days across 24 URLs.**
Search is barely visiting. That, not the page-view total, is what the affiliate applications turn on.

**Not established:** whether the sitemap moved crawling. 2.72/day before it shipped vs 3.00/day in
the 5 days after, two of those days zero. Five days cannot separate that from noise — **re-measure
at 30 days**, i.e. around 2026-09-19.

---

## 3. IMAGERY — the live thread, and where it actually stands

### 3.1 Four hallucinated heroes, regenerating now
Owner, 2026-08-31: the heroes "hallucinated what darts and dartboards and other objects look like".

**The measurement narrows this from "loads of pages" to four**, and the discriminator is mechanical:

> **An active `assets` row ⇒ the image is current and good. No asset row AND a served image ⇒ a
> July SDXL-era leftover file, and those are the bad ones.** `SELECT ... FROM assets WHERE
> url='/assets/images/<file>'` answers it without opening the image.
>
> ⚠ **The second clause is load-bearing and I left it out at first.** The `news_editorial` lane ran
> this rule on their own pages 2026-08-31 and their `insights-index` returned zero rows — which my
> original wording flags as orphaned. They controlled it instead of trusting the rule: the served
> page carries one `<img>` and it is the logo. **No row and no image is consistent, not orphaned.**
> Run the rule and then look at the page before calling anything a leftover.

- **GOOD, verified by eye:** `hero.jpg` (real bristle board, correct wire spider and bull colours),
  `content-hero-grip-styles.jpg` (four genuinely distinct, correct grip patterns). All Banana.
- **BAD, no asset row:** `hero-home.jpg`, `hero-new-arrivals.jpg` (feathered flights — those are
  archery arrows — and blue board segments), `hero-guides.jpg` (garbled numbers), `hero-sale.jpg`.

**Four `needs_imagery` items filed 2026-08-31 12:35Z** (`SEED_2026-08-31_regenerate_four_hallucinated_heroes.sql`,
`created_by='dartsonline-traffic-2026-08-31'`), at `triaged`, handler `image-build-handler`.
**Verify at the served file, not the item status** — and expect them to overwrite in place, because
`deploy_image_asset` derives the filename from `(asset_key, purpose)` and refuses a caller path.

**⚠ The 8 `stability/…` asset rows on this site are a RED HERRING.** Their `url` is a signed S3 link
that expired 7 days after 2026-07-06, nothing references them, and regenerating them changes
nothing a visitor sees. Left alone deliberately.

**Why regeneration gets the good model** — proven, not assumed: `bugs_closed/382` flipped the
missing-kind routing default to Banana on 08-24, and that commit is an ancestor of the running
adapter build (v1.0.1349 / `ef06af0e0`), with a control commit 5 ahead correctly *not* an ancestor.
Every spec also carries `kind:'hero'` explicitly, since the **absent** kind was 382's cause.

### 3.2 The guides need per-section images — mechanism is NOT ours and NOT built
Owner, 2026-08-31: guides want an accurate image per small section (ring grip, razor grip, shark
grip…). Measured today:

```
grip-styles     0 content images, 7 h2 + 6 h3   ← the owner's own example
board-setup     1 in-body illustration, 7 h2 + 3 h3
flight-shapes   1 in-body illustration, 8 h2 + 2 h3
beginners       1 in-body image,        7 h2 + 1 h3
```

**Do not splice figures into `article-body` to satisfy this.** Figure and prose share one llm-owned
field, so anything added today dies at the next body rewrite — this lane lost four figures exactly
that way in August.

**⚠ AND DO NOT WAIT FOR COMPOSITION EITHER.** The `news_editorial` lane answered this on 2026-08-31
and the answer is unambiguous: **P5 — the phase where a guide migrates onto parent/children — is
months away, not weeks.** Measured by them the same day: P1 is in progress with **nothing live**,
the council has returned REVISE **three times** on the wiring plan, and **0 of 2,249 `page_components`
carry a `parent_instance_id`; 0 of 386 `content_components` declare a slots block**. P5 sits behind
P2, P3 and P4, and 035 §5 gates it on P1–P3 holding "for real weeks" with §6.1's un-owned-page
question settled first — and these guides are precisely the un-owned pages that question is about.

**The third option, which is this lane's to take and needs nothing from them:** the trap is not
composition's absence, it is that **ONE field owns both prose and figures**. A guide whose **h3
sections are separate `page_components` rows** — ordinary flat sections, no `parent_instance_id`,
no composition — has per-section durability **today**, because a rewrite then targets one section's
row and cannot take a sibling's figure with it. That is the same durability property composition
would give, one grain coarser. Composition's genuine addition over it is nesting a figure *inside*
a prose section; for `grip-styles`, where each h3 (Ring / Razor / Shark) wants exactly one image,
**finer sections are sufficient and nesting is not needed**.

Durable design for the imagery half remains
`inline_guide_imagery/PLAN_2026-08-14_durable_inline_guide_imagery.md` (plan-as-truth, unbuilt).
`grip-styles` is the natural canary; the other lane wants it at **P2**, not now.

**Accuracy is no longer the constraint** — current-setup generation is correct; only placement and
durability are open.

---

## 4. Affiliates — unchanged, and unblocked

Apply where the gate is **not** traffic: **Awin → Red Dragon** (10%, 60-day, £5 refundable deposit),
**Adtraction → Darts Corner** (free; owner logged in 08-24, waiting for a traffic figure),
**Paid On Results** (second route to Red Dragon). **Webgains → Target Darts** rejected the *network
account*, so it is a re-application once there is a number worth quoting. **Amazon UK** has no
traffic minimum but starts a 180-day / 3-sale clock — apply when there are readers to convert.
Privacy-page precondition is met.

---

## 5. Bugs this lane filed or contributed to

| bug | state |
|---|---|
| `bugs_open/384` — a landed card image never invalidates its listing | **fixed at the framework** (`requestPageListReresolve`). ⚠ my first mechanism was WRONG and is corrected in-file; read the correction block |
| `bugs_open/399` — CTA copy vs recorded destination never compared | **owned by another lane, in build.** Their measurement: the writer IS shown the destination and misdescribes it anyway, 155/1,060 = 14.6% |
| `bugs_open/410` — three seams fail toward the quiet default | filed; candidate 3 has a precedent guard in the same package |
| `bugs_open/198` — stylesheet clobber | dartsonline restored; 3 other sites restored by the owner |

---

## 6. Traps this lane paid for — read before touching anything

- **A `[MEASURED]` marker on a figure you were TOLD is a false claim.** I did this in 410 and had to
  correct it. Relayed numbers get attributed.
- **An inherited filter is the previous question still being answered** — it can only ever narrow,
  so the error is always an undercount and "it seems low" is the only symptom. Cost two wrong counts.
- **Cite the symbol, not the line number.** A line I cited in 410 expired within hours.
- **A nav rebuild refreshes stored chrome everywhere and republishes almost nothing** — 2 of 23 here.
  Fix with `docs/leopardessconsulting/scripts/reconcile_footer_nav.sh` (now has `--absent` and a
  three-valued predicate, both added by this lane).
- **`detected` items do not drain on this site.** File at `triaged`.
- **The B2 sync lands tens of seconds after the DB stamp** — a census taken straight after a deploy
  manufactures false negatives. I made this mistake twice.

---

## 7. What I would do next

1. **Verify the four heroes at the served files** (they were `triaged` at 12:35Z 08-31).
2. **Re-measure crawler traffic around 2026-09-19** — the only honest read on whether the sitemap
   worked.
3. **Per-section guide images: take the flat-sections route** (§3.2) rather than waiting for
   composition — answered 2026-08-31, P5 is months away. The design question this lane still owes
   an answer on is whether any guide genuinely needs figure-*inside*-prose rather than
   figure-*between*-sections; if one does, that makes the guides a P2 consumer and `news_editorial`
   want to know.
4. **Apply to Awin and Paid On Results** — neither gates on traffic and both are one form.
5. **Fleet imagery:** other sites will carry the same orphaned-hero pattern. The discriminator in
   §3.1 is one query and should be run fleet-wide before anyone regenerates anything.

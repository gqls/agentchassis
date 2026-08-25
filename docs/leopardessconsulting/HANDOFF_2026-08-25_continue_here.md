# HANDOFF — leopardessconsulting.co.uk, 2026-08-25

**Start a fresh session from exactly here.** Self-contained. Everything marked
`[MEASURED 2026-08-25]` was checked first-hand that day; anything older carries its own date
and should be re-grounded before you act on it — two of four premises had expired in the five
days before this session, and one published figure turned out to be wrong (§5).

**Site:** `leopardessconsulting.co.uk` · `site_id 4851f6fc-71cf-4160-a270-e03d6d3e0732`
**Home page:** `/index.html` · page name `index`
**Approved plan this continues:** `~/.claude/plans/let-s-do-1-2-and-3-ancient-crab.md`
**Predecessor handoff:** `docs/leopardessconsulting/HANDOFF_2026-08-14_services_restore.md`
(its §3 items 4–7 are all closed; do not re-run them)

---

## 1. Read these before running anything

1. **`page_components.id` is not stable across rerenders.** Address by
   `(page_id, slot_name)`. A captured id gave me `UPDATE 0` twice on 2026-08-18; only the
   `DO`/`RAISE` block turned that silent no-op into a refusal.
2. **`updated_at` bumps on a no-op write.** After a rerender, matching timestamps do **not**
   mean content changed. Keep the pre-change fetch — on 2026-08-25 that is the only thing
   that stopped me reporting a self-inflicted regression that had not happened.
3. **Compare templates on BYTES.** `length()` = 11,893 vs `octet_length()` = 11,903 on the
   `info-card-grid` template — em dashes. And `psql -At` appends a trailing newline, so the
   naive extract's md5 never matches the DB's.
4. **Probe the CAPABILITY, not the commit.** A binary stamps only its own build commit, so
   grepping it for a *fix's* sha returns ABSENT whether or not the fix is in. Grep for a
   literal the fix added, with a positive and a negative control in the same exec.
5. **An empty `$STAMP` silently reads as "up to date"** — `git rev-list --count ..HEAD`
   returns 0. Guard the variable.
6. **The section authority here is `site_specs.site_plan`**, not `site_plan_sections`
   (which has no rows for this site). If you add or remove a section, update it, or the next
   rebuild reading the authority drops your work.

---

## 2. State — what is done and verified

**Chassis `v1.0.1339`**, built from `a7459a44b` (2026-08-25), sha PRESENT in `/proc/1/exe`
with a negative control ABSENT. `[MEASURED 2026-08-25]`

| item | state |
|---|---|
| Sitemap | 27 → 36 urls, all probe 200. **Done 08-16.** |
| Home page: stat band + evidence chart | live; both now also in the section authority |
| Home CTAs | 3 × `/contact.html`, 0 × `/tools/`, verified at the served page |
| Duplicate card block | **removed** — all five cards restated `features`; one title was identical |
| A0 shared carousel CSS | **fixed** — gap `normal`→24px, arrows 22×28→**44×44** (was below WCAG 2.2's 24px minimum) |
| A1 count-up | **verified animating**: 69 mutations `0→22`; single write under reduced motion |
| `/services.html` | **REGRESSED, deliberately not repaired** — see §4 |

Commits this session: `b1b9d000b` (248 correction), `7a1207035` (A0), `3a58523f9` (A2 pt 1),
`1ecca1fed` (A3 groundwork + correction).

---

## 3. THE DECISIONS WAITING ON THE OWNER

Nothing below is blocked on work — each is blocked on a choice.

### D1 — the `trust` voice rule *(open since 2026-08-19, blocks 19 pages)*
**138 of 145** banned-phrase hits site-wide are the single pattern
`\btrust(ed|worthy|s)?\b`. It now flags the site's **own product name** ("The AI Vendor Trust
Checklist"), the titles of research reports we quote, and other people's statistics — because
the trust content pillar was built *after* the rule was written on 2026-07-18.
**Recommendation:** narrow it to the self-congratulatory forms (`trustworthy`,
`deserves trust`, `a trusted partner`, `you can trust us`) and leave the plain noun.
It is the owner's own rule, so a session should not narrow it unilaterally.
Ready and unapplied: `scripts/VOICE_2026-08-17_banned_phrases_ready.sql`.

### D2 — hero imagery: how much distinctiveness is worth buying
**21 of 36 pages open with the identical `hero.jpg`; only 6 distinct images exist site-wide**
`[MEASURED 2026-08-25]`. Nothing is broken — every page serves an image. The archetypes are
already in the data (**12** `page_type='content'`, **5** `blog-post`) and the naming
convention already works on four pages (`hero_<page>` → `/assets/images/hero-<page>.jpg`).
Choice: **two archetype heroes** (one content, one blog — cheapest, kills the worst of the
sameness), **four to six** by theme, or **per-page** for the dozen that matter most.
Cost is one Banana generation each plus an eyeball; the site rule is that every generated
image is looked at before wiring.

### D3 — who fixes the regeneration bug that keeps undoing this site
`/services.html` has now been hand-restored **three times** and regenerated back each time.
The owner has already chosen "fix the cause first" — the open question is **who**. The cause
sits in the `bugs_open/238` / `248` family, both owned by other lanes. Either this lane takes
one of them on (a real piece of platform work, council-scope), or it contributes evidence and
waits. Contributing is cheap and already done twice; taking it on is a different commitment.

### D4 — who builds the design critic (Part B of the approved plan)
It is **not a new idea**: it is `features_open/018` (owner-raised 2026-07-24, "specified, not
built") and Phase 2 of the active `vigilant_designer_offer_analysis` lane, which already holds
four owner decisions about it. The estate rule is contribute, don't compete. Choice: **hand
the design across** to that lane (it is written up in the approved plan and ready to send), or
**build it here** — which needs a Go change to `isStorageEnabledAgent` plus a chassis roll
before any seed will work. Note the idea has now been independently re-proposed at least five
times without being built, which is itself worth the owner's attention.

---

## 4. `/services.html` — regressed, recorded, NOT repaired

`[MEASURED 2026-08-25]` — third occurrence of the same damage.

| | 08-14 restore left | now |
|---|---|---|
| `info-card-grid` cards | 6 | **3** |
| `teaser-reveal-panel` items | 6 | **5** |
| service icons referenced | 6 | **0** |

All six `icon-service-*.jpg` still serve **200** — the page simply no longer names them. The
item keys have been rewritten again, so a restore means re-deriving the icon↔item mapping
against the new keys, exactly as the 08-14 session had to.

⚠ **`bak_leo_services_pc_20260814` is the PRE-repair snapshot, not the good state** — its
`image_url`s are empty too. I misread it as restorable before checking. There is no backup of
the repaired state; the 08-14 repair was written directly.

---

## 5. The correction that changes A3

I published **"29 of 36 live pages carry one image or none"** on 2026-08-19. **It is false.**
The sweep was `grep -c '<img'`, and every hero on this site is a CSS `background-image`, which
that command cannot see. True picture `[MEASURED 2026-08-25]`: **0** pages with no image,
**21** sharing one `hero.jpg`, **6** distinct images site-wide.

The figure is still uncorrected in `README_where_we_are.md` (2026-08-19 entry) and in
`SUMMARY_2026-08-18_evidence_on_the_page.md`; both are append-only/never-overwrite, so the
correction goes in as a dated note, not an edit. Logged in `WRONG_CALLS.md`.

**The transferable check:** a census of absence must be validated against one positive
instance. Open one page the count called empty and look at what it actually serves.

---

## 6. Work queue, in order

1. **D2 answered → A3.** One archetype hero end to end, eyeballed, before the rest. Route A:
   dispatch `image-build-handler` with `item_type:"needs_imagery"`, spec carrying
   `key`/`kind`/`prompt`/`purpose`/`asset_key`. **Omit `scope`** (with `scope:"page"` it emits
   `needs_page` → `page-build-handler`, which regenerates every `source:"llm"` field and
   clobbers hand-authored copy). **Omit `brand_update`.**
2. **Then the carousel** (A2 part 2), which was deferred *because* of the imagery: converting
   `features` today would print lucide icon names as literal text, and the six orphaned
   service icons are 3:1 landscape illustrations that cannot fill a 44×44 `object-fit` chip.
   Once cards have real images, the carousel is worth having.
3. **D1 answered → the voice pass**, starting from the ready SQL.
4. **D4 answered → the design critic**, either handed over or built.

---

## 7. Commands

```bash
# dispatch anything, with a publish receipt (never the stdin kcat form — it drops ~4 in 5)
./docs/leopardessconsulting/scripts/orchestrate_safe.sh <agent_type> '{"site_id":"…","domain":"…"}'

# one page, no-LLM rerender
./docs/leopardessconsulting/scripts/rerender_page_safe.sh \
  4851f6fc-71cf-4160-a270-e03d6d3e0732 leopardessconsulting.co.uk <page_name>

# browser probes (all keep their own controls)
python3 docs/leopardessconsulting/scripts/probe_carousel.py     # gap + arrow size
python3 docs/leopardessconsulting/scripts/probe_countup.py      # mutation trail, both motion modes
python3 docs/leopardessconsulting/scripts/verify_audit.py       # computed styles vs a claim
python3 docs/leopardessconsulting/scripts/audit_live.py <url> <width> <out.json>
python3 docs/leopardessconsulting/scripts/shot_clip.py <url> <out.png> <width> <selector>

# pre-rerender escalation gate — BOTH branches must return 0 rows
#   see docs/leopardessconsulting/scripts/L9_services_carousels.sql §4

kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```

**Backups made this session:** `bak_leo_home_cta_20260825`, `bak_icg_template_20260825`,
`bak_leo_siteplan_20260825`, `bak_leo_home_pc_20260825`.

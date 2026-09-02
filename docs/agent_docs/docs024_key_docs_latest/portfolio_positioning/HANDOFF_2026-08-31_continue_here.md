# HANDOFF — portfolio_positioning — 2026-08-31. **START HERE. First task is §1.**

Supersedes `HANDOFF_2026-08-26_continue_here.md` ⚠ **whose late addition about 414 was REFUTED —
its 08-27 CORRECTED block is authoritative; do not cite the late addition itself.** Owner read-out:
`SUMMARY_2026-08-25_sitemaps_became_a_mechanism.md` still current for the sitemap arc; the remake
arc has no summary yet, deliberately — one brief held at review is not a milestone.

**Counts carry the date they were counted** (owner ruling 2026-08-22).

---

# 0. STATE IN ONE PARAGRAPH

Sitemaps are done, live, self-maintaining, and **proven on the fresh chassis build** (rolled
~2026-08-30 evening; 5 of the 6 latest runs COMPLETED on the new pods, latest 08-31 08:12Z).
**The remake programme started 2026-08-26 and is PAUSED AT THE HUMAN GATE**: three briefs sit at
`needs_human_review` — `advertise.co.uk` is the real one (first of the 22 remakes; the two others
are last week's test briefs) — and the owner reading + releasing is the only path forward; that is
Flow A working as designed, not a stall. `bugs_open/414` (the lendzy marker): served bodies
re-verified CLEAN 08-31 by this lane; the bug is the **414 session's** to close now the roll their
class fix was waiting for has happened. One live blemish (oufe.com sitemap 404 after a transient commit TIMEOUT) was found, remedied and
**VERIFIED HEALED at the served body 09:46Z** — §1a records it as a worked example.

---

# 1. FIRST TASKS

## 1a. oufe.com sitemap 404 — RESOLVED same morning; kept as the worked recipe for a failed run

**Outcome: re-selected 09:44:25Z after the backdate, served body verified 09:46Z — 200, 2,262 B,
19 `<loc>` = 19 deployed pages.** Nothing left to do for oufe; what follows is the recipe for the
NEXT transient commit failure (and §2 tracks the residual).

Background: 08-31 03:38Z selection FAILED at `commit_sitemap` (git-adapter TIMEOUT, transient —
every later run COMPLETED); the served `/sitemap.xml` went/stayed **404** against **19** deployed
pages. Selection stamps the rotation BEFORE the outcome, so the failed run consumed oufe's slot;
this lane backdated the stamp 08-31 ~08:45Z (guarded on the exact failed value; ⚠ the column is
NOT NULL — backdate, never NULL) making it age-due for the next tick.

```sql
SELECT s.domain, r.last_selected_at FROM site_discovery_rotation r JOIN sites s ON s.id=r.site_id
WHERE s.domain='oufe.com' AND r.agent_type='sitemap-refresh';
```
- Stamp newer than `2026-08-31 08:45Z` → it re-ran. **Judge at the body**: fetch
  `https://oufe.com/sitemap.xml` (**`rm` the temp file first**), expect ~19 `<loc>`s. A COMPLETED
  status or a `sitemap_commit_result:ok` key is NOT a landed commit — that key sat `ok` inside the
  FAILED run while the site 404'd.
- Still 404 after TWO fresh ticks → the git adapter or the serving path, not the rotation; check
  git-adapter health/logs before touching the rotation again.
- `gaswholesalers.com` was also age-due (stamp 08-28 09:33) and should drain in the same window.

## 1b. Is `642` still draining cleanly? (2 minutes)

Both queries verbatim in `HANDOFF_2026-08-26_continue_here.md` §1a (selection join + due-set).
Readings so far: design 0/28/2 · 08-26 14:20Z 0/14/11 · 08-26 18:42Z 0/15/10 · **08-31 08:40Z
1/2/5** (the age-due 1 was gaswholesalers). Small numbers are the healthy shape now; tick gaps are
normal when nothing is due. ⚠ `orchestration_states` reaps ~24h — an old stamp with no matching
run is UNKNOWABLE, not failed. ⚠ idea.uk always reports `dropped=1` — permanent, by design (their
`/privacy.html`→`/privacy` 301); do not re-chase.

## 1c. `622`'s guard — same falsifiable query (08-26 §1b), still expected boring

08-31: no `deployed_pages=0` row carries a stamp; min is `apis.uk` at 1. Still never violated,
still never consulted — sites born 08-22→08-25 all entered with pages. First real test remains
the next site seeded pageless.

## 1d. The brief queue — the remake programme's actual gate

```sql
SELECT s.domain, wi.status, wi.updated_at FROM site_work_items wi JOIN sites s ON s.id=wi.site_id
WHERE wi.item_type='needs_brief_review' AND wi.status NOT IN ('complete','cancelled','rejected');
```
As of 08-31: all three still `needs_human_review` (indoorplanters 08-20 · buytoletcalculator
08-21 · **advertise.co.uk 08-26**, site row `test`+LOCKED `d991a5b8…`).
- **Released** → the build flow takes over: watch it as the FIRST framework remake end-to-end;
  the site row must move off `test`/locked deliberately, not by side effect.
- **Owner answers the brief's Q2 in prose** (is advertise.co.uk the HUB of the marketing
  cluster?) → fold the answer into the NEXT briefs' directions. **Do not fire the
  websitepromotion.co.uk / seotools.co.uk / designblog.co.uk briefs until Q2 is answered** —
  the deliberate 08-26 sequencing decision, still standing.
- Brief mechanics: `scripts/fire-brief-writer.sh <domain> [direction]` — site row first
  (`test`+LOCKED, `buytoletcalculator` precedent), holds at `needs_human_review`, builds nothing.

# 2. LIVE vs NOT-YET-PROVEN

| | |
|---|---|
| **642 on the new chassis** | 5/6 latest runs COMPLETED post-roll, latest 08-31 08:12Z, 0 dropped |
| **Brief writer** | 3 runs, 3 holds that held (advertise brief: 15,915 B, 13 keys, differentiation stays off the estate neighbours by name) |
| **622 guard** | applied + APPROVED — **behaviourally unproven, still** |
| **642 quiet-gate deferral** | still only visible in falling due-counts, never in a withheld selection — qualify any "proven" claim |
| **NEW residual: a FAILED sitemap run consumes its slot** | selection stamps before outcome; 1 instance in ~100+ runs (oufe, §1a). Not filed; if it recurs, file WITH the retry-design question |

# 3. WHAT IS OPEN (beyond §1)

- **The 22 hosted-site remakes** (`DECISION_2026-08-20_remake_the_hosted_sites.md`): №1 advertise
  at the gate (§1d). 3 protected (`leopardess.co.uk`/`.uk`, `cartoon.co.uk`); do NOT start with
  `businessinsurancequotation.co.uk`; twin pairs are ⚑OWNER calls; before-snapshot convention:
  `salvage/<domain>/` (advertise done — a Drupal RSS aggregator, nothing original at risk).
- **`bugs_open/414`** — NOT this lane's to close. The 414 session owns it; the roll their class
  fix waited for has now happened (their §7j recipe includes the post-roll verification). This
  lane's part is done: docs corrected 08-27, lendzy bodies re-verified clean 08-31. The lesson
  that is ours to keep: **census the PAYLOAD, not the vocabulary; never clear a row on its first
  matching window** (WRONG_CALLS 2026-08-27, two tallies).
- **`skip_reason` residue** — unchanged, population 0 as of 08-25, low priority since 642.
- **"Is 3 days right?"** — reframed (a re-probe floor for serving-side drift), not a question.
- **Cloudflare managed robots.txt blocks AI crawlers** (seen on cv1.co.uk 08-26) — owner's call;
  `traffic_probe` lane owns Cloudflare historically.
- **The Christmas card sender** (G3/G4) — delivery half FIRST; read `bugs_open/283`.
- **`adversecreditmortgage.co.uk` stays halted** — owner's call.
- **21 portfolio domains have no register row** (as of 2026-08-21) — and the register's
  two-copies question (`positioning_register` DB vs `REGISTER_positioning.md`) is still
  undecided: **do not edit both**.

# 4. TRAPS (new first; 08-26 §4 still holds in full)

- ⚠ **`site_discovery_rotation.last_selected_at` is NOT NULL** — re-queue by backdating past the
  floor, guarded on the exact current value; never try to NULL it.
- ⚠ **A `sitemap_commit_result:ok` key inside a FAILED run is not a landed commit** — the serve
  went 404 while that key read ok. The artefact census is the only truth (08-26 §4b).
- ⚠ **Selection stamps BEFORE outcome** — a failed run silences its site until change or floor.
- ⚠ **Census a retraction by its PAYLOAD** — an agent's paraphrase keeps the claim and drops your
  search vocabulary; and `position()` shows only the FIRST occurrence in a row (WRONG_CALLS 08-27).
- Carried (08-26 §4a–c): orchestration reap window ~24h · artefact census discipline (`rm` first,
  judge the body, canonicalise both sides) · council submission summaries must list EVERY
  pre-flight assertion · don't run `landmines-verify-dispatch.sh` while another lane has a NEW
  uncommitted LANDMINES entry · `ls` the migrations dir before naming a file · `pages.updated_at`
  is bumped by convention, not trigger — a new `pages` writer must bump it.

# 5. FILES OF RECORD

**Cold start:** this file → `SUMMARY_2026-08-25_sitemaps_became_a_mechanism.md` →
`README_where_we_are.md` (owner's log; entries 2026-08-26 evening, 08-27 correction, 08-31) →
`NOTES_portfolio_positioning.md` (2026-08-31 entry; 08-26 (h)–(l) with (l)'s CORRECTED block).

**Remakes:** `DECISION_2026-08-20_remake_the_hosted_sites.md` · `scripts/fire-brief-writer.sh` ·
`positioning_register` (DB; 189 rows as of 08-21) · `salvage/advertise.co.uk/` ·
brief = `site_specs.mission_brief` for advertise.co.uk (is_current).

**Sitemaps:** `platform/orchestration/actions/render_sitemap_action.go` · migrations `590`/`622`/
`642` (+`_ROLLBACK`) · `scripts/site-discovery-files.py` · register **SEO-007**/**SEO-002** ·
`COUNCIL_SUBMISSION_642_sitemap_follows_the_deploy.json`.

**414:** `bugs_open/414_…acceptance_marker…md` (its §7 + REFUTED block are the fixing session's
record) · 016b §9 tripwire entry (with the 08-27 added block) + §10 index entry ·
`WRONG_CALLS.md` 2026-08-27.

**Addendum (same day, after owner review of the briefs):** the owner commented — see
`DECISION_2026-08-31_best_in_vertical_fullness_and_the_advertise_marketplace.md` (four rulings +
the advertise marketplace compatibility verdict + a PREPARED release edit for fullness, not yet
applied). §1d changes accordingly: on release, apply the prepared edit + record Q3 as answered
(monetisation = own-network ad sales, direct preferred). **Q2 (cluster hub) remains the open
question gating the next briefs.** Fire directions for future briefs now carry best-in-vertical +
no-omission + fullness until the copy lane's propagation plan ships.

**Addendum 2026-09-02 — the owner ruled; §1d is overtaken by events:** fullness edit APPLIED
(brief revision `5dac12fd`, original preserved; review file re-rendered) · negative-identity
claims default OUT of copy fleet-wide (recorded in the DECISION file's EXECUTED addendum,
CONTRIB'd to copy_quality_two_stage — do not build copy machinery here) · **Q2 ANSWERED:
advertise.co.uk IS the marketing-cluster flagship** → the three cluster briefs FIRED
(websitepromotion `a6fae8ee` · seotools `9ca54346` · designblog `d8eb90be`), each `test`+LOCKED,
each holding at `needs_human_review` on completion, before-snapshots salvaged. **The review
queue is now up to six briefs; the advertise BUILD is still held — the owner edited, and has
not yet said "go build". That word, plus reviews of the three cluster briefs, are the owner
actions the programme now waits on.** Verify the three briefs landed (the fire script's three
queries per domain) if NOTES 2026-09-02 shows the in-flight marker unresolved.

**Addendum 2026-09-02b — "go build advertise" ARRIVED and was EXECUTED; §1d is closed for №1:**
the owner confirmed the domain (advertise.co.uk; advertise.uk was a slip, now owner-confirmed)
and said the release word. Released 12:13–12:15Z per the review item's own `how_to_release`:
review item `518ed780` → complete (approved_by owner) · `needs_domain_research`
`research_advertise.co.uk` created triaged (domain-submitter shape) · site `d991a5b8` off
`test`+LOCKED → `active`, unlock LAST. Claimed by `build-dispatch-loop` 12:15:21Z, first tick;
classifier orchestration `e44a44d7-…` running against the owner-edited brief (`5dac12fd`).
**The lane's live task is now WATCHING remake №1 end-to-end** (NOTES 2026-09-02 (later) has the
watch-points: cascade after classifier, handshake-race caution, the negative-identity copy guard
at build review, site row reaches `deployed` by pipeline not by hand). The owner's queue still
holds five briefs (websitepromotion / seotools / designblog real; indoorplanters /
buytoletcalculator test).

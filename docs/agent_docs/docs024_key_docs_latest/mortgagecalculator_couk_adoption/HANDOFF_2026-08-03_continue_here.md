# HANDOFF — mortgagecalculator.co.uk adoption — **COLD START, read this first**

**Written 2026-08-03 ~10:20 UTC.** Supersedes
`HANDOFF_2026-07-31_adopt_mortgagecalculator.md` as the entry point (that one is
still worth reading for the original brief, but **three of its instructions are now
known-wrong** — see §7).

Chassis at handoff: **v1.0.1238** (pods started 2026-08-03 10:08 UTC).

---

## 1. The owner's constraints — these have not changed

1. Adopt at `--fidelity high`, i.e. **editable, not frozen**. Already done.
2. The `sites` git repo holds a full copy of the live site as a safety net. **Done
   and verified** — that was the single most dangerous thing about this domain.
3. **"We don't want to bring it down."** The live site is production and must not
   regress. Every change to it so far has been verified file-by-file at the wire.

## 2. State right now

| thing | state |
|---|---|
| Live site | **The original 29 files are intact** — re-verified at the wire after every change: 28 byte-identical, 1 differing (`robots.txt`, Cloudflare, expected — §7.3). Only deliberate changes: two link fixes (07-31) and six logo-link fixes (08-03) |
| Adoption | Complete. 26 pages, all `build_status='planned'` except one |
| Stylesheet | **LIVE** — `/assets/css/styles.css`, 17,016 bytes, HTTP 200 (`gqls/sites` `6f2a71a32`). Additive; the original pages use `/css/style.css`, which still serves |
| `/guides/first-time-buyer/index.html` | **`deployed` AND now fully chromed** — the ordering canary **PASSED 2026-08-03 11:06 UTC**. 8,854 → 20,550 bytes, `<header>`/`<nav>`/`<footer>` all present, CSS 200. §9.3 is DONE |
| Chrome (`site_components`) | **EXISTS as of 11:01 UTC** — header 2,125 B · head 8,635 B · footer 987 B, all `rendered`. Was **zero rows**; that was the whole "no chrome" defect (see §12) |
| Specs | 10 current aspects. `identity` + `content_direction` are **operator-written positioning** (see §4) |
| Work items | 8 complete · **66 deferred** · 11 `needs_human_review` · **0 armed** (the nav rebuild filed 26 `page_rerender`; all deferred immediately) |
| Site lock | **HELD**, and the hold is proven at the gate, not assumed (`gate_says: NOT SELECTABLE — held`) |

**The homepage item is `deferred` and must stay that way** until the owner sees a
styled rebuild. It is the ONLY page whose URL does not change (`/index.html` →
`/index.html`), so it is the only one that overwrites live content.

## 3. How to hold and release this site — the controls that actually work

**Two controls, and you need to understand both.**

```sql
-- (a) SITE LOCK — holds everything. Works as of 2026-08-03 (I fixed it; see §6).
UPDATE sites SET locked_at=NOW(), locked_by='<lane>: why'
 WHERE domain='mortgagecalculator.co.uk';
UPDATE sites SET locked_at=NULL, locked_by=NULL WHERE domain='mortgagecalculator.co.uk';

-- PROVE the hold rather than assuming it — run the gate's own question:
SELECT coalesce((SELECT s.domain FROM site_work_items wi JOIN sites s ON s.id=wi.site_id
  WHERE s.locked_at IS NULL AND wi.status IN ('triaged','approved')
    AND wi.attempt_count < wi.max_attempts
    AND (COALESCE(wi.approval_mode,'auto')='auto' OR wi.status='approved')
    AND NOT EXISTS (SELECT 1 FROM site_work_items a WHERE a.site_id=wi.site_id AND a.status='claimed')
    AND s.domain='mortgagecalculator.co.uk' LIMIT 1), 'NOT SELECTABLE — held') AS gate_says;

-- (b) ITEM STATUS — the finer control. 'deferred' is NOT in workItemTerminalStatuses,
--     so the row keeps its idx_swi_dedup slot and release is one UPDATE.
UPDATE site_work_items SET status='deferred' WHERE ... AND status IN ('triaged','approved');
```

**To run a subset:** set just those items `triaged`, unlock the site, and **run an
auto-defer backstop** for everything else, because each handler creates the next
item and you are racing a 120-second tick:

```bash
# every 15s, defer anything dispatchable that is NOT the items you are running
UPDATE site_work_items SET status='deferred', updated_at=NOW()
 WHERE site_id=(SELECT id FROM sites WHERE domain='mortgagecalculator.co.uk')
   AND status IN ('triaged','approved') AND id NOT IN (<your item ids>);
```

This is not belt-and-braces paranoia — it caught **19 items in one sweep** on
2026-08-02 when `build-site-planner` finished, including 3 `needs_page` and 1
`needs_rerender`, the two types that reach the live site.

**The chain, so you know what you are holding back:**
`needs_domain_research` → `needs_vertical_research` → `needs_strategy` →
`needs_briefing` → `needs_site_plan` → (site planner emits) `needs_composition`,
`needs_design`, `needs_imagery`, `needs_page`, `needs_rerender`.
Everything up to `needs_briefing` writes only specs. **`build-site-planner` is where
pages start to exist.**

## 4. Positioning — done, but NOT protected

`identity.target_audience` and `content_direction.things_to_avoid` (+ the `formatted`
blob) now scope the site to **secured lending on residential property**, and name
`loancalculator.co.uk` (unsecured) and `loanandmortgagecalculator.co.uk` (decisions
spanning both). The three sites are mutually coherent. Detail: RUNBOOK §8, PLAN D3.

**Two things to know before you touch specs:**

- **`divergence_rule` is INERT.** The original handoff says to write one. Nothing in
  the platform reads it — no Go, no SQL, no prompt — and the classifier's output
  schema has no such key, so the next classifier run drops it. The live sibling
  carries its divergence in `target_audience`, which IS read.
- **`site_specs.pinned` does NOT protect a spec.** `write_site_spec` never reads it
  and the replacement row defaults to false. **So a future classifier run WILL
  overwrite the positioning.** Re-check it after any agent run that writes
  `identity`. There is no way to pin it; the only durable hold is the site lock.

## 5. Editing `content_direction` by hand — the trap

The content writer reads exactly one field:
`{{.site_specs.specs.content_direction.formatted}}`. `FormatContentDirection` only
regenerates it on the **action** path, so **raw SQL must update the array and
`formatted` in the same statement** or the spec looks applied and steers nothing.
Verify array↔blob agreement, not length:

```sql
SELECT count(*) AS items, count(*) FILTER (WHERE position(item in fmt) > 0) AS in_formatted
  FROM (SELECT jsonb_array_elements_text(data->'things_to_avoid') AS item,
               data->>'formatted' AS fmt
          FROM site_specs WHERE site_id=(SELECT id FROM sites WHERE domain='mortgagecalculator.co.uk')
            AND aspect='content_direction' AND is_current) t;   -- must be N of N (was 14/14)
```

## 6. Platform defects found from this lane

| ref | what | status |
|---|---|---|
| `bugs_open/183` | `domain-research-classifier` `classify_and_extract` output cap was 6000 — the fleet's only one — and truncating, which **blocks every adoption** | **Cap raised to 32000, live.** Bug OPEN: the structural fix (split the step) is untouched |
| `bugs_open/184` | LLM markdown reaches the page as literal `**asterisks**`. 3 components, 3 unrelated sites | OPEN, unowned. Live on our first-time-buyer hero |
| `sites.locked_at` ignored by the dispatch gate | A lock that read back correctly and held nothing | **FIXED** — one clause added to `find_dispatchable_site`. `CONTRIB` filed in `bugfix_029_dispatch_gate/`; migration `213` (theirs) still unapplied |

**Not touched, deliberately:** 25 steps truncate fleet-wide, concentrated in council
seats at cap 8000 (`review_editquality` 21/105 and 19/136; max successful **7996 of
8000**). That is `bugs_open/138`'s active lane, which already raised one of those caps
and found it grew back in 3 days. Reported to the owner, not unilaterally changed.

## 7. Corrections to the original handoff — do not re-follow these

1. **"Write a `divergence_rule`"** — inert, see §4.
2. **"No B2 credentials"** — false since 07-31; they are on this machine.
3. **`robots.txt` "is a real origin file"** — it is not what the CDN serves.
   Cloudflare injects a Managed block **at the top** (491 bytes origin vs 2327
   served). The original lane used `tail -5` and missed it. **Take bytes from the
   bucket, never from curl.** A whole-site diff will always show `robots.txt` as
   differing — that is correct and expected, not a regression.

## 8. What I got wrong, so you do not repeat it

**I built a page before the site had a stylesheet.** Among the 19 items the backstop
swept were `needs_composition` ("resolve palette/layout/typography") and `needs_design`
("generate site stylesheet"). The resulting page 404s its CSS and has no
header/nav/footer, and my first reading was *"the rebuild produces unstyled orphan
pages"* — **false**. The sibling `loancalculator.co.uk/guides/hidden-loan-fees.html`,
built by the same machinery, has nav+footer and its CSS resolves 200.

**Correct order: composition → design → pages.**

Its two 404 links are also **not** defects — `/tools/affordability/index.html` and
`/scorecard-simulator.html` are both `build_status='planned'` rows, i.e. forward
references. One query against `pages` settled it.

> Twice I built a confident negative reading from a partially-built system, and both
> times the fix was to find the **comparison** — a working sibling, a counterfactual
> query — not to look harder at the broken thing on its own.

## 9. Next steps

1. **[DONE 2026-08-03 ~11:30 UTC]** `needs_composition` and `needs_design` both
   **complete**. The stylesheet is **live**: `/assets/css/styles.css`, 17,016 bytes,
   HTTP 200, written by `webdesign-agent` (`gqls/sites` commit `6f2a71a32`) and
   deployed. Site **re-locked** afterwards; 0 items armed.

   The original 29 files were re-verified after it landed: **28 identical, 1 differing
   (`robots.txt`, Cloudflare)** — the new stylesheet is additive and touched nothing.
   Note the site's ORIGINAL pages reference `/css/style.css`, a different file that
   still serves; the new `/assets/css/styles.css` is only for rebuilt pages.

   **Success test is the artefact, not the item status:**
   `curl -o /dev/null -w '%{http_code}' https://mortgagecalculator.co.uk/assets/css/styles.css`
   must become **200** (404 at handoff, both spellings).

   > **CORRECTED before anyone acted on it — an earlier draft of this handoff told
   > you to suspect a filename mismatch. There is no mismatch.** I had compared
   > against one sibling (`loancalculator.co.uk`, which serves `style.css`) and
   > inferred our trial page's `styles.css` reference was wrong. Counting the deploy
   > repo settles it the other way: **17 domains carry `assets/css/styles.css`
   > (plural) and only 4 carry `style.css`** — so plural is the house convention and
   > the sibling is the minority. `rerender_pages_actions.go:558` hardcodes
   > `/assets/css/styles.css` in its **fallback head** (the `else` branch taken when a
   > site has no rendered head), which therefore agrees with 17 of 21.
   > **So the trial page's 404 was never a naming bug — the site simply had no
   > stylesheet yet, because I had deferred the job that makes one.** One `ls` across
   > `~/projects/sites/*/assets/css/` is the whole check; a single sibling is not a
   > population.

   Useful tell that comes out of this: **a page referencing `/assets/css/styles.css`
   with no `<header>`/`<nav>`/`<footer>` took the FALLBACK head**, i.e. it rendered
   before the site had chrome. That is a signature worth recognising, not a defect in
   itself.

   > **Expect to WAIT, and do not read the wait as a stall.** The gate picks ONE site
   > fleet-wide, `ORDER BY wi.created_at ASC` — **oldest item wins, priority is only a
   > tiebreak**. `finetuning.uk` currently holds **125 triaged items, the oldest from
   > 2026-07-26**, so it wins nearly every tick and our 08-02/08-03 items are behind
   > all of them. It is not deadlock — the site's own `claimed`-mutex forces rotation,
   > and our composition item did get through — but a release here can sit for many
   > ticks. Check where you are in the queue rather than assuming something broke:
   > ```sql
   > SELECT s.domain, wi.item_type, wi.created_at FROM site_work_items wi
   >   JOIN sites s ON s.id=wi.site_id
   >  WHERE s.locked_at IS NULL AND wi.status IN ('triaged','approved')
   >    AND wi.attempt_count < wi.max_attempts
   >    AND NOT EXISTS (SELECT 1 FROM site_work_items a WHERE a.site_id=wi.site_id AND a.status='claimed')
   >  ORDER BY wi.created_at ASC, wi.priority ASC LIMIT 5;
   > ```
   > **Do NOT "fix" this by back-dating your own rows.** If a lane genuinely needs to
   > jump the queue the honest lever is the site lock on the *other* site, and that is
   > the other lane's call, not yours.
2. **Re-lock the site** once they finish, and prove it with the §3 gate query.
3. **[DONE 2026-08-03 11:06 UTC — the canary PASSED.]** The first-time-buyer guide was
   re-run and came back **WITH** `<header>`/`<nav>`/`<footer>`, CSS resolving, 8,854 →
   20,550 bytes. **composition → design → pages is now CONFIRMED end to end**, so the
   remaining pages can be released as a batch. Live-site integrity re-verified after:
   all files byte-identical bar `robots.txt` (Cloudflare) and the trial page itself.

   **But fix the header CTA first — see §13 — or one broken button ships onto EVERY
   page in the batch.** The nav is correctly withheld (§12), the CTA is not.

   **Rebuild 2–3 more non-homepage pages** so the owner judges them as a set. Use the
   assemble-only single-page path (no `reason` ⇒ no LLM, authored copy untouched):
   ```
   ./docs/agent_docs/docs024_key_docs_latest/cta_link_integrity/scripts/049b_deploy_single_page.sh \
     <page_id> 62b5978e-4271-4589-8e00-4baebfc0447c mortgagecalculator.co.uk
   ```
   It bypasses the dispatch queue, so the site can stay LOCKED throughout — which is
   what kept this run contained. **Only pages with `page_components` rows can build**;
   the other 25 have none and will correctly `skip`.
4. **Owner decision: the homepage.** It is the only page that overwrites live content.
5. **Owner decision: redirects.** 22 of 23 original URLs move; the old files keep
   serving as orphaned duplicates and **nothing in the platform reconciles them.**
   There is also no cross-site duplicate-content machinery at all — worth telling the
   owner plainly, since three of our sites now cover adjacent finance topics.
6. **Still open from the original brief:** `images/mortgagecalculatormono.xcf` is a
   GIMP source file, publicly fetchable (200). Removing it is a separate deliberate
   commit.

## 10. Verifying the live site — the check that matters

```bash
cd ~/projects/sites/mortgagecalculator.co.uk
while read -r f; do rel="${f#./}"
  curl -s -o /tmp/l.tmp "https://mortgagecalculator.co.uk/$rel"
  a=$(sha256sum "$f" | cut -d' ' -f1); b=$(sha256sum /tmp/l.tmp | cut -d' ' -f1)
  [ "$a" = "$b" ] || echo "differs: $rel"
done < <(find . -type f ! -path './.git/*')
# expect exactly one line: robots.txt  (Cloudflare — see §7.3)
```

Before pushing anything to `gqls/sites`, run the dry-run gate and **check the exit
code** — the flag is `--dry-run` (v4); the v3 spelling `--dryRun` exits 2 and a
piped grep then prints nothing, which looks exactly like a safe no-op:

```bash
b2 sync --delete --skip-newer --dry-run --no-progress \
  mortgagecalculator.co.uk b2://portfolio-sites/mortgagecalculator.co.uk; echo "EXIT $?"
```

`git pull --rebase` before pushing — **never let it become a merge**, because the
deploy derives changed domains from `git diff HEAD~1 HEAD` and a merge makes your own
commit the first parent, dropping the domain while the run still goes green.

## 11. The five standing docs for this lane

`PLAN_2026-07-31_adopt_mortgagecalculator.md` (decisions D1–D4) ·
`RUNBOOK_mortgagecalculator_couk.md` (§1–§9, every command with its gotcha) ·
`NOTES_mortgagecalculator_couk.md` (running log incl. every wrong turn) ·
`README_where_we_are.md` (**the owner's plain-prose log — append only, never edit**) ·
`SUMMARY_*` (none yet — no milestone has warranted one since adoption completed).

---

## 12. Where chrome ACTUALLY lives — read this before diagnosing a bare page

**`pages.rendered_header` / `rendered_footer` / `rendered_head` are VESTIGIAL.** They
are empty for **all 562 pages fleet-wide**, including sites whose served pages plainly
have nav. Only `discovery_checks/check_missing_structure.go` still reads them. A site
with empty columns is not evidence of anything.

```sql
-- the census that settles it — note the ABSENCE of a WHERE domain=
SELECT s.domain, count(*) FILTER (WHERE coalesce(length(p.rendered_header),0)>0) AS has_hdr
  FROM pages p JOIN sites s ON s.id=p.site_id GROUP BY s.domain;   -- 0 for every site
```

Chrome comes from **`site_components`** (`slot_name` header/footer/head), written by
`render_site_components`, which sits in `nav-updater`'s workflow:

```
populate_nav_tables → render_site_components → create_rerender_items → get_pages_for_rerender
```

This site had **14 `site_nav_items` and 0 `site_components`** — stalled between steps 1
and 2. The documented bypass fixed it in one run (COMPLETED on the first poll):

```
./docs/agent_docs/docs024_key_docs_latest/bugfix_149_nav_membership/TRIGGER_nav_rebuild.sh mortgagecalculator.co.uk
```

**Two traps inside that run:**

1. **It files a `page_rerender` item PER PAGE — 26 here, homepage included.**
   `get_pages_for_rerender` filters on **`p.status`**, NOT `p.build_status`, and every
   page of ours is `status='active'`. What actually protects the homepage is
   `rerender_single_page_action.go:565` + `:168-209`: a page with **zero
   `page_components`** assembles to nothing and returns `skipped:true` without
   deploying. Only `guide-first-time-buyer` has component rows. **Defer the items
   anyway** — do not leave the invariant resting on a downstream branch.
2. **`status` vs `build_status` on `pages` is a genuine trap.** Same table, both
   plausible, and only one is what the rerender selector reads.

**A one-link nav (Home only) is CORRECT here, not broken.** `NavFetchableOnly` drops
nav items whose target has never been deployed, because chrome ships on every page
(the `bugs_open/049` fix); `loadFetchablePageSet` always injects the site root, so Home
survives. **Cliff worth knowing:** at **0** deployed pages the filter disables itself
and ships the FULL nav; we have exactly 1, so it is active. The nav fills in as pages
ship — `nav-updater` runs `force_rerender:true`, so re-running it refreshes chrome.

**Prove the LOCK at its source, not with §3's reconstruction.** §3's `gate_says` query
hardcodes `s.locked_at IS NULL`, so it proves the query respects the lock, not the gate.
The gate is a SQL string in the DB, and it does carry the clause:
```sql
SELECT default_config->'workflow'->'steps'->'find_dispatchable_site'->'config'->>'query'
  FROM agent_definitions WHERE type='build-pipeline-trigger' AND is_active
    AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

## 13. OPEN DEFECT — the header CTA is checked by a LOOSER predicate than the nav

**Fix this before the batch rebuild, or it ships onto every page.**

In `render_site_components_action.go`, ~70 lines apart:

| what | helper | predicate |
|---|---|---|
| nav items | `loadFetchablePageSet` (`nav_tables.go:258`) | `status NOT IN ('deleted','archived')` **AND NOT `NeverDeployedPagePredicate`** |
| header CTA | `loadResolverPageSet` (`resolve_internal_links_action.go:486`) | `status NOT IN ('deleted','archived')` — **no deployment check** |

The CTA fallback (`:172-187`) picks an "interactive page" and validates it against the
looser set, so an unbuilt page passes. Ours resolved to `/tools/stamp-duty/index.html`
(`build_status='planned'`) → **`HTTP 404` at the wire on the "Get Started" button**.

> **MEASURED, and the first measurement was WRONG — do not repeat it.** From the DB this
> looks like **2 of 14 sites** (`lendzy.co.uk` → `/tools/price-cap-checker/index.html`
> also reads never-deployed). **At the wire lendzy's target returns `HTTP 200`**, and its
> served homepage carries no `header-cta` at all. `deployed_at IS NULL` means "no
> recorded deploy", **not** "does not serve" — it over-reports. The code asymmetry is
> real; the confirmed live instance is **one site, ours**. Same trap as `bugs_open/098`,
> inverted. One `curl -o /dev/null -w '%{http_code}'` is the whole check.

Cheapest containment for this lane: build `tool-stamp-duty` before the batch, or re-run
`nav-updater` once another interactive page is deployed so the fallback picks a real one.
The structural fix is to make the CTA use the same predicate as the nav.

**Also live on the trial page, both pre-existing:** `bugs_open/184` (literal
`**Decision Engine**` asterisks — assemble-only re-ships stored HTML, so a rerender
cannot clear it) and `/assets/images/favicon.png` **404**, referenced twice from the
head component.

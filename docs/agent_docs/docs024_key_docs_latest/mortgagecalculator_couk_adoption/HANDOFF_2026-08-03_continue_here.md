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
| Live site | **29 files, 200 across the board, intact.** Only deliberate changes: two link fixes (07-31) and six logo-link fixes (08-03) |
| Adoption | Complete. 26 pages, all `build_status='planned'` except one |
| `/guides/first-time-buyer/index.html` | **`deployed`** — the single-page trial. Live at the new URL; old URL still serves |
| Specs | 10 current aspects. `identity` + `content_direction` are **operator-written positioning** (see §4) |
| Work items | 6 complete · **40 deferred** · 11 `needs_human_review` · 2 released for composition+design |
| Site lock | **Released** at handoff time to run composition+design. **Re-lock when they finish** (§3) |

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

1. **[IN FLIGHT at handoff]** `needs_composition` (`5484b580-05cc-44b8-9f1c-f328dfd59796`)
   and `needs_design` (`8d462e1b-16a8-48d9-b407-0d79d28a6f8e`) released with a
   backstop running. **Success test is the artefact, not the item status:**
   `curl -o /dev/null -w '%{http_code}' https://mortgagecalculator.co.uk/assets/css/style.css`
   must become **200** (it is 404 now). Note the sibling uses `style.css`; our trial
   page asked for `styles.css` — **check which filename the generated stylesheet
   actually takes, and whether built pages reference the same one.** If they
   disagree, that is a real bug and worth its own file.
2. **Re-lock the site** once they finish, and prove it with the §3 gate query.
3. **Rebuild 2–3 non-homepage pages** so the owner judges them styled. Re-run the
   first-time-buyer guide too — it was built unstyled and should be redone.
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

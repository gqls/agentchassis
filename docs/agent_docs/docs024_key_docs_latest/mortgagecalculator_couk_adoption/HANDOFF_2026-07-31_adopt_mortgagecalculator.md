# HANDOFF — adopt mortgagecalculator.co.uk into the framework

**Written 2026-07-31** by the `loanandmortgagecalculator_couk` lane, which has just taken
its own site through this exact path end to end. **That lane keeps
loanandmortgagecalculator.co.uk; this one owns mortgagecalculator.co.uk.**

Everything below was measured today against the live site and the live database, not
inferred from the sibling run. Where I did not measure something I say so.

**Open the standing five in this directory before you start** — `PLAN`, `RUNBOOK`,
`NOTES`, `README_where_we_are`, and a `SUMMARY` only at a real milestone. This file is
your cold start, not a substitute for them.

---

## 1. STOP — do not run `082` yet. There is a live-outage hazard specific to this domain.

**mortgagecalculator.co.uk is served from B2 but is NOT in the deploy repo.** Both halves
verified today:

```bash
cd ~/projects/sites && git ls-tree -d --name-only HEAD | grep -i mortgage
#   loanandmortgagecalculator.co.uk        <- only the sibling. This domain is ABSENT.
curl -sS -o /dev/null -w "%{http_code}\n" https://mortgagecalculator.co.uk/     # 200
```

Now read the deploy workflow (`~/projects/sites/.github/workflows/deploy-to-b2.yml`):

```bash
CHANGED=$(git diff --name-only HEAD~1 HEAD | grep -E '^[^/]+\.[^/]+/' | cut -d'/' -f1 | sort -u)
for domain in $CHANGED; do
  if [ -d "$domain" ]; then
    b2 sync --delete --skip-newer "$domain" "b2://portfolio-sites/$domain"
  fi
done
```

**The failure chain.** A verbatim `page_rerender` does not just write to B2 — it
**commits the rendered page into `gqls/sites`**. I watched it happen on the sibling site:
`deploy_result.data.repo_url = https://github.com/gqls/sites`, `commit_message =
"Rerender: loans/damage-checker.html"`. So if you adopt this domain and let one rerender
through:

1. the platform commits `mortgagecalculator.co.uk/<one-file>.html` into the repo;
2. the workflow sees `Changed domains: mortgagecalculator.co.uk`;
3. `[ -d "$domain" ]` is now **true** — the commit just created it, holding **one file**;
4. `b2 sync --delete` makes the bucket match that directory. **Everything else under
   `b2://portfolio-sites/mortgagecalculator.co.uk/` is deleted. The site goes down.**

The run goes **green**.

**Blast radius, measured so you know this is specific and not general:** of the **16**
framework-managed domains (a `sites` row with pages), **0** are absent from the deploy
repo. Every existing one is safe because its files are already there. Adopting this domain
would make it **the first** in that shape — which is exactly why the sibling lane never
hit this and why it is not already a filed bug.

### So step one is: get this domain into the deploy repo, completely, before adopting.

And "completely" is load-bearing, because **B2 holds at least one file the local tree does
not**:

```bash
curl -sS https://mortgagecalculator.co.uk/robots.txt | tail -5
#   ... Disallow: /*.env / Disallow: /wp-login.php
#   # Sitemap location (replace with your actual domain)
find ~/projects/domains/mortgagecalculator.co.uk -name robots.txt      # NOTHING
```

That is a real origin file (substantial, WordPress-style hardening rules), not
Cloudflare's Managed robots.txt block. A sync from the local tree alone **deletes it.**

**`[UNMEASURED]` — I could not enumerate the bucket.** The `b2` CLI is present on this
machine but there are **no B2 credentials** (`env | grep B2_` is empty; the keys live only
as GitHub Actions secrets). So `robots.txt` is the one extra file I found *by probing*, and
I cannot promise it is the only one. **Enumerate the prefix before your first push** —
either get credentials from the owner or have them run:

```bash
b2 ls --recursive b2://portfolio-sites/mortgagecalculator.co.uk/
```

Reconcile that listing against the local tree, commit the union, and verify byte-identity
live **before** anything can trigger a sync. Two local files should probably **not** be
published, so decide deliberately rather than by accident: `README.md` (currently **404**
live, i.e. not in B2 — do not add it) and `images/mortgagecalculatormono.xcf` (currently
**200** live, i.e. a GIMP source file is publicly fetchable — a small thing, but now is
the moment to remove it if you want it gone).

---

## 2. The byte source of truth is `gemini/02`, and it is not the obvious directory

```
c2144f3e…  live https://mortgagecalculator.co.uk/
be66f725…  ~/projects/domains/mortgagecalculator.co.uk/index.html            <- NOT live
c2144f3e…  ~/projects/domains/mortgagecalculator.co.uk/gemini/02/index.html  <- live
```

**All 23 pages were verified byte-identical to live from `gemini/02`.** The top-level tree
is a different, plausible-looking copy. Establish this by digest, never by dates — and note
the fleet landmine this belongs to (`~/projects/sites2` vs `~/projects/sites`, same class).

Inventory: **29 files** — 23 HTML, `css/style.css`, `js/calculators.js`, 2 PNG, 1 XCF,
1 README.

---

## 3. The crawl will find 20 of 23 pages, and I can name the 3 it will miss

Adoption drives its page list from the crawl, so **an unlinked page is silently not
adopted**. Computed from the local link graph today:

| page | why the crawler cannot reach it |
|---|---|
| `404.html` | nothing links to it — **correct**, and the sibling adoption also skipped its 404 |
| `guides/buy-to-let.html` | **orphan** — real content, no inbound link anywhere |
| `guides/your-mortgage-scorecard.html` | **orphan** — the homepage links `guides/mortgage-scorecard.html`, a name that does not exist |

Reachability from `index.html` by following links: **20 of 23.**

**So decide before you adopt**, because it is much cheaper now than after: either fix the
inbound links first (then all 22 real pages adopt), or accept a 20-page adoption and record
which two are outside the framework. **Do not discover this by counting rows afterwards** —
20 looks like a perfectly plausible success.

Predicted assertion for your run: `pages` count = **20** (or 22 if you fix the links
first), all `rebuild_policy='owned'`, and `needs_content_page` + `needs_tool_recreation`
= **0**.

---

## 4. Four defects already live on this site, unfixed by owner decision

The owner was offered these on 2026-07-31 and chose **"not now"** (decision D7). They are
listed because **three of them interact with adoption**, so you may want to revisit:

1. **6 of 9 guides link `Home` to `index.html` from inside `/guides/`**, resolving to
   `/guides/index.html` → **live 404**. (Also the directory-index landmine: `/guides/`
   itself 404s on every B2 site.)
2. **The homepage links `guides/mortgage-scorecard.html`**; the file is
   `your-mortgage-scorecard.html` → **live 404**, and the cause of one orphan above.
3. **2 orphan guides** (above) — the ones adoption will silently skip.
4. **No `sitemap.xml`** (404), and `robots.txt` still carries
   `# Sitemap location (replace with your actual domain)`. Also no favicon at all
   (`/favicon.ico`, `/favicon.svg`, `/apple-touch-icon.png` all 404).

Fixing 1–3 before adopting is roughly an hour and makes the adoption complete rather than
partial. That is a judgement for you and the owner, not something I should decide for this
lane — but note the ordering argument: **a link fix before the crawl changes what gets
adopted; the same fix afterwards does not.**

---

## 5. The adoption sequence, with the parts that actually bite

Full detail, every gotcha with its command:
`docs024_key_docs_latest/loanandmortgagecalculator_couk/RUNBOOK_loanandmortgagecalculator_couk.md`
**§9 (byte gate), §10 (holding the queue), §11 (divergence specs)**. Read those three; they
were written from a real run today. The short form:

### 5a. Pre-flight

```bash
# is another lane already working this domain?
SELECT item_type, status FROM site_work_items
 WHERE status NOT IN ('complete','cancelled','rejected')
   AND (spec::text ILIKE '%mortgagecalculator%');
# exactly ONE ported-page component must exist, or adopt_verbatim refuses
SELECT id, name FROM content_components WHERE function='ported-page';
#   a7daa5c5-8cfd-4f2c-8e09-de6abcb637ef  "Ported Page (webdesign.co.uk)"  -- reuse it, NEVER seed a second
# chassis has the locked path (a roll is not evidence — grep added strings + a control)
kubectl exec -n ai-persona-system <chassis-pod> -- sh -c '
  strings /app/agent-chassis | grep -c "Verbatim adoption complete"   # 1
  strings /app/agent-chassis | grep -c "verbatim_adoption_deploy"     # 1
  strings /app/agent-chassis | grep -c "apply_adoption_plan"'         # 3 (control)
# fidelity must cross the spawn boundary — the input_mapping is an ALLOW-LIST
SELECT default_config->'workflow'->'steps'->'call_adopter'->'config'->'input_mapping'
 FROM agent_definitions WHERE type='site-adoption-orchestrator' AND is_active;
#   must contain "fidelity?"  -- verified present today
```

### 5b. Start the hold BEFORE you submit

`adopt_verbatim` creates one `page_rerender` per page **inside the adoption transaction,
already `status='triaged'`**, and `build-pipeline-trigger` runs at `interval_seconds=120`.
The window is **under two minutes**, so this cannot be done by hand. Poll every 2s and flip
them to `deferred`:

```sql
UPDATE site_work_items SET status='deferred', updated_at=NOW()
 WHERE site_id=(SELECT id FROM sites WHERE domain='mortgagecalculator.co.uk')
   AND item_type='page_rerender' AND status IN ('triaged','approved');
```
Mine caught **41 in one second**. Copy the poller:
`loanandmortgagecalculator_couk/`… see RUNBOOK §10 (the script itself was a scratch file;
the SQL above is the whole of it).

`deferred` is deliberate: **not** in `workItemTerminalStatuses`, so the row keeps its
`idx_swi_dedup` slot and nothing can create a duplicate behind it; release is a plain
`UPDATE` back to `triaged`.

**`sites.locked_at` does NOT hold dispatch**, despite
`213_dispatch_gate_matches_dispatcher.sql:106` containing `AND s.locked_at IS NULL`. The
**live** `build-pipeline-trigger` row has no such clause. I nearly relied on it. Read the
live `agent_definitions` row, not the migration.

### 5c. Submit

```bash
bash scripts/initial_messages/020_build_pipeline/082_submit_domain_unified.sh \
  mortgagecalculator.co.uk --from https://mortgagecalculator.co.uk --fidelity locked
```
The file is not executable — use `bash`. `--email` is **inert** on the adopt branch
(`input_data` is exactly `target_url`, `destination_domain`, `fidelity`), so do not read a
printed email as evidence anything was stored. The script's line 152 still prints fidelity
as "RECORDED ONLY"; its own header contradicts that and **the header is right**.

Budget ~30 minutes of queue latency but **do not count on it** — mine wrote its `sites` row
in **90 seconds** and its pages in 4½ minutes, which is why the hold must be running first.
Find the run by payload, never the printed id:

```sql
SELECT current_step, status FROM orchestration_states
 WHERE collected_data->'input_data'->>'destination_domain' = 'mortgagecalculator.co.uk'
 ORDER BY created_at DESC LIMIT 3;
```

### 5d. Assert the locked branch took — this is the one that protects the calculators

```sql
SELECT item_type, count(*) FROM site_work_items
 WHERE site_id=(SELECT id FROM sites WHERE domain='mortgagecalculator.co.uk')
   AND item_type IN ('needs_content_page','needs_tool_recreation') GROUP BY 1;
```
**Must return zero rows.** If either appears, an LLM is about to rewrite working
calculators — stop.

### 5e. Gate the bytes, and expect it to fail everywhere

`adopt_verbatim` has exactly one possible byte source: firecrawl's `rawHtml`, which is the
serialised **post-JavaScript DOM**. Measured on two sites now: **0 of 27** matched on
loancalculator.co.uk and **0 of 41** on loanandmortgagecalculator.co.uk. Expect **0 of 20**
here.

```bash
python3 .../loanandmortgagecalculator_couk/gate_component_bytes.py            # report
python3 .../gate_component_bytes.py --repair                                  # load real bytes in
python3 .../gate_component_bytes.py                                           # the RE-RUN is the evidence
```

**You must edit two constants** — the script hardcodes `DOMAIN` and derives `SITE` as
`~/projects/sites/<domain>`. For this domain the byte source is
`~/projects/domains/mortgagecalculator.co.uk/gemini/02` **until you have completed step 1**,
after which it is the repo path like everyone else. Parameterising it properly (a `--domain`
/ `--source` flag) is a small, worthwhile job and would make it reusable for the next
adoption; I left it hardcoded because it had one caller.

What the crawl actually did to my pages, so you know what you are discarding:

- every relative `href`/`src` **absolutised**;
- `<meta charset="UTF-8">` → `<meta http-equiv="Content-Type" …>`;
- `&#9776;` decoded to a literal `☰`;
- **the skip link became an absolute URL with `#content` appended** — so the first thing a
  keyboard or screen-reader user hits reloads the page instead of jumping down it. An
  accessibility regression on every page. This one is why the gate is not pedantry;
- `mortgages/repayment.html` came back **+11,432 bytes** because its script builds a 24-row
  amortisation table on load and the crawl serialised the result. **This site has its own
  `repayment.html` with the same generator** — expect the same, and expect it to be the
  outlier that makes the others look fine.

Two traps in the gate itself: `sha256(text::bytea)` is invalid SQL (use
`sha256(convert_to(col,'UTF8'))`), and `content_data.sha256` is **not** a fidelity check —
`adopt_verbatim` computes it over the bytes it is storing, so it agrees with itself on a
mutated page. Every write must carry the shared lock predicate
(`lock_helpers.go:42-46`) and you must **count `RowsAffected`**, because a lock-suppressed
`UPDATE` reports success and changes nothing.

### 5f. Release one, prove it, then release the rest

```sql
UPDATE site_work_items SET status='triaged' WHERE id='<one item>';
```
Then check the live sha256 is **unchanged** and the platform's own commit has an **empty
diffstat**. Mine did: that empty diff is the rebuild-safety property working, and it is the
acceptance test for the whole exercise.

**The queue will not oblige you.** The dispatcher takes
`DISTINCT ON (site_id) … LIMIT 1` ordered by `site_id`, one site per invocation; my site
ranked **14 of 14** and a released item sat unpicked for 10+ minutes. Fire `page-rerender`
directly instead — the exact `kcat` invocation is in RUNBOOK §10. `kcat -P` **exits 0 having
sent nothing**, so verify by the `orchestration_states` row, never the exit code.

---

## 6. This reverses a decision the owner took hours earlier — and that changes the spec work

Decision **D6** (2026-07-31) was: *record the divergence on the new site only;
mortgagecalculator.co.uk keeps no `sites` row and its narrow-authority position is
documented intent, not configuration.* It was taken because the owner had said both old
sites stay up unchanged.

**Adopting it creates the `sites` row, so D6 no longer holds.** That is an improvement — the
divergence becomes enforceable on both sides instead of one — but it makes a new piece of
work mandatory rather than optional, and it is the reason the owner asked for this in the
first place:

**Once this domain has specs, all three sites' positioning must be made mutually coherent,
not just this one's.** Current state, read live today:

| site | `sites` row | `identity.target_audience` today |
|---|---|---|
| `loancalculator.co.uk` | yes (another lane owns it) | "UK consumers seeking to understand, compare, and manage personal loans and car finance" |
| `loanandmortgagecalculator.co.uk` | yes | rewritten today to the whole-borrowing-picture framing, with an explicit `divergence_rule` naming the other two |
| `mortgagecalculator.co.uk` | **none yet** | — |

So when you set this one, narrow it deliberately to **mortgage-only authority** and say what
it is *not*, mirroring the `divergence_rule` the sibling now carries. If you leave the
adoption's auto-generated identity in place you will get the failure this whole exercise
exists to avoid: mine came back as *"UK consumers researching loans, mortgages, car finance,
and debt management"* — **generic, and nearly identical to the loan-only site's.** The
classifier cannot know about positioning; it describes what it crawled.

**Three things about `site_specs` that will cost you an afternoon if you do not know them**
(all measured; full form in RUNBOOK §11 and `LANDMINES.md`):

1. **The aspect named `audience` is read by NOTHING.** It is the most widely-populated
   aspect in the database (29 of 33 sites) and no agent, prompt or Go path consumes it. The
   sibling lane's own plan had named it as one of three targets. Same for `editorial`,
   `voice`, `content_standards`, `terminology_and_positioning`.
2. **The writer reads exactly one field of `content_direction`: `.formatted`.** New keys go
   *inside* `content_direction` so `formatted` carries them. A hand-written spec that does
   not regenerate `formatted` is **invisible** — it looks applied and changes nothing.
3. If you regenerate `formatted` yourself, **gate the reimplementation**: reproduce the
   current spec and require a match as a multiset of **lines**, not as a string, because Go
   map iteration is random so the stored section order is arbitrary. Working, gated port:
   `loanandmortgagecalculator_couk/set_divergence_specs.py` (also hardcodes its domain —
   same small parameterisation job).

Schema: the JSONB column is **`data`**, not `spec_data`; `idx_site_specs_current` is
`UNIQUE (site_id, aspect) WHERE is_current`, so the supersede must precede the insert.

**And the limitation to tell the owner plainly, because it is the honest answer to "managed
by the framework":** there is **no cross-site duplicate-content or topical-overlap machinery
in this platform at all.** `check_content_duplication` is single-site
(`WHERE p.site_id = $1`), `remove_duplicate_page_sections` is single-page, and
`cross_site_contamination` detects another site's `company_name` in rendered HTML — not
topical overlap. The specs steer *new writing*. Nothing will warn anyone if the three sites
converge again. If the owner wants that guard, it does not exist and would need building.

---

## 7. Standing consequence you inherit the moment you adopt

**After adoption, `page_components` and the deploy repo are two independent writers for the
same files.** A verbatim `page_rerender` commits `Rerender: <file>` into `gqls/sites`; your
own edits go the other way. They agree only while something keeps them in step.

**So: after ANY change to this site's files, re-run the byte gate with `--repair`**, or the
DB still holds the old bytes and the next rerender silently reverts you. This is recorded in
`LANDMINES.md` ("after a locked adoption … TWO writers") and as **ADO-038** in the concept
register.

---

## 8. What is already proven, so you do not re-derive it

- `fidelity=locked` works: 41 pages adopted `owned`/verbatim, **zero** LLM work items.
- The verbatim rerender bypass ships stored bytes untouched, and a rerender after repair is
  **content-neutral** (empty diffstat, live digest unchanged).
- `rawHtml` is **not** the served bytes — settled on two sites, 0/27 and 0/41. This closed
  ADO-037's own `verify-later` line.
- The gate-and-repair pattern works and is scripted (**ADO-038**).
- The browser audit harness works on this site's pages: **13 of 13** interactive mortgage
  calculators scored `RESPONDS` against the live old site on 2026-07-31, after three faults
  in the shared harness were fixed. So if it tells you a calculator here is dead, **suspect
  the instrument first** — 4 of 5 adverse verdicts on the sibling site were harness faults,
  and `evalpage.py` settles it in twenty seconds.

## 9. What I would check first, in order

1. **Enumerate `b2://portfolio-sites/mortgagecalculator.co.uk/`** and reconcile against
   `gemini/02`. Nothing else is safe until this is done. (§1)
2. Decide with the owner whether to fix the three broken inbound links first, since that
   changes whether you adopt 20 pages or 22. (§3, §4)
3. Get the domain into the deploy repo and prove byte-identity live **before** any rerender
   can fire. (§1)
4. Only then: hold, submit, assert the locked branch, gate, repair, release one, prove it.
   (§5)
5. Set the narrowed mortgage-only positioning, and tell the owner about the missing
   cross-site guard rather than implying the framework provides one. (§6)

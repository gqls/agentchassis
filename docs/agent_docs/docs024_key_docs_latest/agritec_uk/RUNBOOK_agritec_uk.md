# RUNBOOK — agritec.uk

Every command here was run and got the right answer, with its gotcha attached. When one changes,
change it **here**, not in your scrollback.

---

## 1. Is the site framework-managed yet?

    kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
      -c "SELECT domain, status, build_status, last_deployed_at FROM sites WHERE domain='agritec.uk';"

**2026-08-21: zero rows.** No `sites` row exists. Do not read that as "the site is down" — it is
live and hand-built, outside the framework entirely.

**Gotcha:** `sites.status='active'` is written by the code but is *not* in the validated lifecycle
vocabulary (`draft/building/review/published/deployed/archived/error`). Never scope a query on it
expecting meaning.

---

## 2. Measure page depth (the ledger floor, and the acceptance check)

This is the command behind every floor column in `SUBJECT_LEDGER.md`. **Re-run it against the
rebuilt site** — the check only means something if it is the same measurement.

    BASE=https://agritec.uk
    for p in /guides/physics-of-light.html /guides/elms-stacking.html ; do
      b=$(timeout 20 curl -sS -L --max-time 15 "$BASE$p" 2>/dev/null)
      words=$(printf '%s' "$b" \
        | sed -e 's/<script[^>]*>.*<\/script>//g' -e 's/<style[^>]*>.*<\/style>//g' -e 's/<[^>]*>/ /g' \
        | tr -s ' \n' ' ' | wc -w)
      printf "%-46s %6s %4s %4s %4s %4s\n" "$p" "$words" \
        "$(printf '%s' "$b" | grep -c '<h2')" \
        "$(printf '%s' "$b" | grep -c '<table')" \
        "$(printf '%s' "$b" | grep -c 'equation-box')" \
        "$(printf '%s' "$b" | grep -c '<figure\|<img\|<svg')"
    done

**Gotchas.**
- Strip `<script>` and `<style>` **before** stripping tags, or minified JS inflates the word count
  by hundreds and every page looks deep.
- `equation-box` is *this site's* class name. The rebuilt site will use whatever the framework's
  equation component is called — find that name first and change the pattern, or the check reads
  0 equations on a page full of them and you will "fail" a page that passed.
- The figure count deliberately includes `<svg>`, which catches inline icons as well as real
  diagrams. It is an upper bound on diagrams, not a count of them. Read the page before believing
  a high number.

---

## 3. Measure what the framework produces at each page type

This is what settled `blog-post` vs `guide` (PLAN §2, Measurement 2).

    kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
    SELECT s.domain, p.page_type, count(*) AS pages,
           round(avg(t.words)) AS avg_words, min(t.words) AS min_words, max(t.words) AS max_words
    FROM pages p
    JOIN sites s ON s.id = p.site_id
    JOIN LATERAL (
      SELECT sum(array_length(regexp_split_to_array(
               regexp_replace(pc.rendered_html,'<[^>]*>',' ','g'),'\s+'),1)) AS words
      FROM page_components pc WHERE pc.page_id = p.id AND pc.rendered_html IS NOT NULL
    ) t ON true
    WHERE s.domain IN ('finetuning.uk','robot-hands.com','gamesdesign.co.uk','oufe.com')
      AND p.page_type IN ('guide','blog-post') AND t.words IS NOT NULL
    GROUP BY 1,2 ORDER BY 1,2;"

**Gotcha:** this counts `page_components.rendered_html`, which is the **stored** snapshot the
assemble step stitches — not what is served. They can differ. For "what a visitor sees", fetch the
URL (§2).

---

## 4. Reachability — find orphaned pages

The check the old site would have failed. Run it against the rebuilt site as an acceptance gate.

    BASE=https://agritec.uk
    TARGET=vapor-pressure-deficit.html
    for p in / /tools/index.html /guides/index.html /deepdives/index.html; do
      printf "  %-30s %s\n" "$p" \
        "$(timeout 15 curl -sS -L --max-time 10 "$BASE$p" 2>/dev/null | grep -c "$TARGET")"
    done

**2026-08-21 result:** all four return 0 for `vapor-pressure-deficit.html`. Its only inbound link
is from the VPD calculator.

**Gotcha:** counting a bare filename matches any page whose href merely *contains* the string, so
a shorter slug can be a false positive inside a longer one. Prefer the full href where slugs
overlap.

**Gotcha:** `bugs_open/116` — link-integrity checks have never actually run fleet-wide. Treat
internal links as unverified by default; this crawl is ours to run, not something the platform
does for us.

---

## 5. Crawl the live site for the true inventory

**The repo copy at `/home/ant/projects/domains/agritec.uk/01/` is stale** — 6 tools against 13
live, and no `/deepdives/` at all. Always inventory from the live site.

    BASE=https://agritec.uk
    for p in / /tools/index.html /guides/index.html /deepdives/index.html; do
      echo "--- $p"
      timeout 20 curl -sS -L --max-time 15 "$BASE$p" 2>/dev/null | grep -oE 'href="[^"]+"' | sort -u
    done

**Gotcha:** a naive recursive crawl emits paths like `/deepdives/../tools/x.html`. They fetch
fine (200) and look like distinct pages. Normalise before counting or the inventory double-counts.

---

## 6. Deploy state and DNS

    dig +short NS agritec.uk          # 2026-08-21: leah/alexis.ns.cloudflare.com — already ours
    curl -sS -o /dev/null -D - -L https://agritec.uk/    # 200; x-amz-* headers = served from B2

**Gotcha:** a green GitHub Action does not mean the site is reachable. If the Cloudflare zone
lookup finds no zone named exactly like the directory, **the cache purge is silently skipped**.
Verify with a cache-busted fetch, not the Action's status.

**Gotcha:** `bugs_open/315` — `pages.deployed_at` is stamped whether or not the object write
succeeded. One page was silently skipped by four consecutive "completed" rerenders. Verify at the
served artefact.

---

## 7. Submission (Phase 3, not yet run)

`082_submit_domain_unified.sh` has **no `--roadmap-file` flag**. For a Tier-3 submission
(mission + roadmap) hand-roll the envelope from `../oufe/TRIGGER_submit_tier3.sh`.

**Gotcha:** both briefs must be `{"text": "..."}` **objects**. A bare string renders `<no value>`
in the prompt template and is silently ignored — no error, just a build with no brief.

**Gotcha:** budget ~30 minutes from publish to run start, not ~2. A missing `orchestration_states`
row is almost always queue latency, not a dropped dispatch. Do not retry on that evidence; find
the run by payload:

    kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
      -c "SELECT current_step, status FROM orchestration_states
          WHERE collected_data->'input_data'->>'domain' = 'agritec.uk'
          ORDER BY created_at DESC LIMIT 5;"

---

## 8. Which specs exist for the site

    kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
    SELECT ss.aspect, ss.source, ss.created_by, length(ss.data::text) AS bytes, ss.created_at::date
    FROM site_specs ss JOIN sites s ON s.id = ss.site_id
    WHERE s.domain='agritec.uk' AND ss.is_current ORDER BY ss.aspect;"

Before submission, expect exactly `evidence_base` and `imagery_style_guide` (plus whatever the
seed adds). After the cascade, expect the oufe-shaped set: identity, classification,
content_direction, design_intent, strategy, briefing, tools, vertical_landscape, submission,
mission_brief, roadmap_brief.

**Gotcha:** never `UPDATE` a spec row in place. Supersede (`is_current=false, superseded_at=now()`)
then insert, in one transaction. The partial unique index enforces one current row per
`(site_id, aspect)`.

---

## 9. Reading an evidence run — the review is MANDATORY, and here is what to look for

`verify_and_register` re-fetches every source and rejects a claim whose quote is not still there
verbatim. That is a strong check and it is easy to over-read. **It establishes provenance and
nothing else.** It does not check that the fact is relevant to this site, that the quote was read
correctly, or that the run answered the question you asked.

Measured on this lane, 2026-08-22: two runs, both `COMPLETED`. The first returned 10 facts, 9 of
them primary GOV.UK/DEFRA, all usable. The second returned 5 facts of which **4 were unusable and
one was actively misleading** — and it reported success identically. So the run status tells you
nothing about the value of the output.

Read every fact, every time, against these five:

1. **Audience / market.** Is this figure about the people who read this site? The failure that cost
   this lane a whole run: Ofgem's price cap is a *domestic* consumer protection, and every reader
   here buys on non-domestic contracts. The claim was true, sourced, current — and wrong for
   every reader. **This is the one that will not look like an error**, because nothing about a
   true, well-cited fact announces that it is about somebody else.
2. **Table scrapes.** Read the stored `quote`, not just the claim. If it contains two figures, a
   `|`, or a run of numbers with no sentence around them, the extractor read a table and picked
   one cell. Example caught here: a claim of "26.11 pence per kWh" whose quote also contained
   "24.67 pence per kWh". Which column it was is unknowable from the register.
   ```sql
   SELECT f->>'value', left(f->'source'->'citation'->>'quote',160)
     FROM site_specs ss JOIN sites s ON s.id=ss.site_id,
          LATERAL jsonb_array_elements(ss.data->'facts') f
    WHERE s.domain='agritec.uk' AND ss.aspect='evidence_base' AND ss.is_current;
   ```
3. **Content-free facts.** A fact with no `value` that carries no assertion either — "the dataset
   was last updated on 30 June 2026" — is a whitelist entry that licenses nothing and dilutes the
   register. Remove it.
4. **Source host.** Group by host; primary sources should dominate. One command:
   ```sql
   SELECT regexp_replace(f->'source'->'citation'->>'url','^https?://([^/]+).*$','\1') AS host,
          count(*)
     FROM site_specs ss JOIN sites s ON s.id=ss.site_id,
          LATERAL jsonb_array_elements(ss.data->'facts') f
    WHERE s.domain='agritec.uk' AND ss.aspect='evidence_base' AND ss.is_current
    GROUP BY 1 ORDER BY 2 DESC;
   ```
   A third-party host is not automatically bad — but it must be scoped in the claim text itself
   ("Under the SFI 2023 offer…"), because the claim text is what the writer reads.
5. **Did it answer the question?** Both halves. The energy run was asked for prices *and* carbon
   intensity, returned no carbon-intensity fact at all, and completed successfully. A silent half-
   answer is the easiest thing to miss, because the facts that did arrive look fine.

**When you remove a fact, ban the figure too** — fail-closed, on the oufe precedent. If it later
turns out to be the right number for a stated purpose, the ban forces a conscious return to the
migration with the market and date attached. See `SEED_2026-08-22b_quarantine_domestic_energy_facts.sql`.

**Verify a quarantine by CONTENT, not by count.** "Facts went 15 to 11" is also what removing the
wrong four looks like. Assert the survivors by id.

## 10. Rewriting the evidence register safely

Supersede-then-insert, **as sequential statements inside one transaction**. Never as a single
statement with data-modifying CTEs: all CTEs share one snapshot, so the INSERT's uniqueness check
cannot see the sibling UPDATE's supersede and you get

    duplicate key value violates unique constraint "idx_site_specs_current"

Guard with `DO`/`RAISE`, not a `SELECT` verify block — `ON_ERROR_STOP` ignores a non-empty result,
so a `SELECT` cannot stop a `COMMIT`. Assert the *pre-state* you expect (row count and fact count);
if another session has written to the register since you read it, the guard aborts instead of
overwriting their work. Worked examples: `SEED_2026-08-22_sfi26_bans.sql`,
`SEED_2026-08-22b_quarantine_domestic_energy_facts.sql`.

**Do not run two evidence-researcher dispatches concurrently.** Each one supersedes and rewrites
`evidence_base` wholesale; two in flight is a lost-update race, and the loser's facts vanish with
no error anywhere. Fire them one at a time and read each before the next.

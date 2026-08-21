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

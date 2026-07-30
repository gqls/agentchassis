# HANDOFF D (2026-07-30) — the about page renders every section twice, and the eight archetype pages share their copy

**Start a fresh thread on this.** Owner asked for the fundamentallyai
deduplication check to be run against vonc.com. It was, it found two distinct
things, and one of them is live and visible.

## The tool, and why it has two halves

The method is the one `bugs_open/151` used on fundamentallyai.com — now committed
as a site-agnostic script:

```bash
# 1. extract the site's sections and its approved facts
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At -c "
SELECT json_agg(json_build_object('page',p.name,'url',p.url,'slot',pc.slot_name,'pos',pc.position,'data',pc.content_data))::text
FROM pages p JOIN page_components pc ON pc.page_id=p.id
WHERE p.site_id='<site>' AND pc.content_data IS NOT NULL;" > sections.json

kubectl ... -At -c "SELECT data::text FROM site_specs
WHERE site_id='<site>' AND aspect='evidence_base' AND is_current;" > evidence.json

# 2. run the census
python3 scripts/dedup_census.py sections.json evidence.json [min_shared_facts]
```

**Both halves are necessary and 151 is the proof.** On fundamentallyai two
sections asserted the identical six facts while being only **18% textually
similar** — so a similarity check alone reported "fine". Conversely, repeated copy
that carries no countable fact is invisible to the fact census. The script runs
fact census, `difflib` similarity, and exact-block matching together.

## Finding 1 — the about page renders every section twice (LIVE)

`/about.html` has **12 component rows that are 6 identical pairs**: the same
`slot_name` at consecutive positions, normalised text identical (similarity 1.00):

| slot | positions |
|---|---|
| `hero-about` | 1, 2 |
| `content-block-about` | 3, 4 |
| `game-master-explanation` | 5, 6 |
| `platform-comparison` | 7, 8 |
| `differentiators` | 9, 10 |
| `gauntlet-cta` | 11, 12 |

**And it paints twice.** Rendered `https://vonc.com/about.html` (11,568 visible
chars):

```
painted x2  'The rules are simple'
painted x2  'A provocation, not a prompt'
painted x2  'The Gauntlet is open'
painted x2  'no credit card required'
painted x16 'Game Master'
```

So a visitor reads the entire about page twice over. **This is the highest-value
item in this handoff — it is live, it is visible to anyone, and it needs no design
decision.**

> **ANSWERED 2026-07-30 — filed as `bugs_open/156`. Read that, not this section,
> for the current state.** Findings, all measured:
> - **Every persisted source says 6** — `site_plan_sections` (is_current),
>   `pages.sections`, and one `pages` row. Only `page_components` is doubled, so
>   the cleanup is safe and a legitimate rebuild will not restore it.
> - **One write pass, not two builds** — all 12 rows inside 93 ms, `created_at`
>   strictly increasing with `position`, positions 1..12 **distinct**. That last
>   point **rules out** two concurrent saves, which would each number their rows
>   1..6 and give two rows *per position*. It was one loop over a list that
>   already held 12 entries, each section duplicated **adjacently**.
> - **The 12-entry list is not recoverable** — it lived in the run's
>   `collected_data`; `orchestration_states` and `site_work_items` for that window
>   have both aged out (0 rows). `save_page_sections`, `CompilePageSections` and
>   `loop_actions` each append once per item, so the loop's *item list* is as far
>   as the evidence reaches. **[UNRECOVERABLE] — do not write a root cause on it.**
> - **"the same cause may be live on other pages and other sites" measures
>   FALSE.** Fleet-wide there are 17 duplicate `(page_id, slot_name)` groups; **11
>   are legitimate** repeated slots with *differing* content (`generic-text-block`
>   ×2–3 on five other sites). The 6 content-identical ones are all vonc `about`.
>   ⇒ **A unique index on `(page_id, slot_name)` would break 11 real pages.** The
>   discriminator is content identity, not slot repetition.
> - **The durable defect is that nothing detects this**: no unique constraint,
>   no guard in the save (all five existing guards compare sections to the *page*,
>   never to each other), `content_hash` empty on all 12 rows, and
>   `grep -rn "HAVING count(\*) > 1" platform/ internal/ pkg/ scripts/` returns
>   nothing fleet-wide. Found by hand, two days after it shipped.
>
> The original first-move instructions are kept below as written.

What is NOT yet established, and should be your first move: **why there are two
rows.** Do not delete one until you know, because the same cause may be live on
other pages and other sites. Start with:

```sql
SELECT id, slot_name, position, created_at, updated_at, build_status, content_hash
FROM page_components WHERE page_id = (SELECT id FROM pages
  WHERE site_id='9ec3b9ee-5b08-461b-b4f8-9e1e03579c74' AND name='about')
ORDER BY position;
```

`created_at` will say whether this was one double-insert or two separate builds,
and `content_hash` will say whether the platform already knows they are identical.
Check `pages.name='about'` has only one row too. **If the cause is a re-planning
or re-adoption path that inserts rather than upserts, that is a platform defect
and belongs in `/bugs_open/` with the mechanism, not a data cleanup here** — see
`bugs_open/151`'s own framing: the site was fixed by hand and the mechanism was
filed separately because it will reproduce.

## Finding 2 — the eight archetype pages restate each other

Cross-page `difflib` similarity, `/archetypes/*.html`:

| pair | ratio |
|---|---|
| oracle ↔ surgeon `call-to-action` | **0.90** |
| surgeon ↔ mentor, scout ↔ surgeon | 0.83 |
| oracle ↔ wildcard, wildcard ↔ mentor | 0.79 |
| surgeon ↔ judge | 0.74 |
| scout ↔ catalyst `hero`, judge ↔ mentor `hero` | 0.71, 0.70 |
| …and ~10 more hero/CTA pairs | 0.67–0.69 |

This is exactly the 151 shape — **independently generated near-duplicate copy
across sibling pages**, produced because per-section copy is written in isolation
with no knowledge of what siblings already said. The mechanism is filed and
unfixed (`bugs_open/151`: `page-content-writer`'s `process_sections_loop` passes no
sibling output into `generate_content`, and `writer_block` is built once per site).

**So: contribute this instance to `bugs_open/151` rather than filing a new bug**,
and decide separately whether to hand-fix vonc's archetype copy now. A CTA being
83% identical across eight pages is arguably fine (it is a CTA); heroes at 0.71
are not, because the hero is where each archetype is supposed to be distinct.
Owner's call on whether it is worth the edit.

## Finding 3 — the fact census barely applies here, and that is worth knowing

vonc has only **4** approved facts in `site_specs.evidence_base` (8 archetypes,
3 tools, 2 guides, 18 pages) against fundamentallyai's 9. Result: 50 of 55 sections
assert **zero** facts, 5 assert exactly one, **none** assert 2+. So 151's "3+
shared facts" threshold cannot fire on this site — the fact half of the census is
near-blind here, and the duplication that exists is textual.

Only `vonc-archetypes` (value 8) is repeated across pages — 5 sections on 2 pages
(`index/gauntlet-cta@3`, and four sections on `about`, which are the duplicated
rows from Finding 1, so this is largely an artefact of that).

**Do not read "no sections assert 2+ facts" as "vonc has no duplication problem."**
It means the ruler is too short to measure this site. That is a genuine limitation
of the 151 method on sites with small evidence bases, and worth recording in 151.

**Worth checking while you are here (not measured):** whether those 4 facts are
still TRUE. `3 tools` and `2 guides` look low against the page census — the site
has `/tools/gauntlet/`, `/tools/arena/`, `/tools/archetype-taster-quiz/` and a
`/guides/tool-arena-interface-guide.html`. If a fact is stale, the rail is broken
by construction, and that is `bugs_closed/104`/claimscan territory.

## Landmines

- **Read `pages.url`; never construct one.** I tested `/about/index.html`, got a
  B2 `NoSuchKey` error whose 286-char JSON body my `innerText` check read as page
  content, and briefly concluded the about page was blank. The real path is
  `/about.html`. Print `%{http_code}`.
- **Render, don't grep** — much of vonc's text arrives client-side.
- The census normalises text and skips URL/id/asset-ish keys before comparing; if
  you widen it, re-check that chrome strings are not inflating similarity (two
  sections already match at 1.00 purely on captured footer text — see HANDOFF C).
- **A similarity score is not a verdict.** 151's whole point is that the number
  can be low while the duplication is total. Read the pairs.

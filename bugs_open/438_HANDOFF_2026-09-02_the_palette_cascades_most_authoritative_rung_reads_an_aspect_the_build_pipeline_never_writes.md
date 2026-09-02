# 438 — the palette cascade's "most authoritative" rung reads an aspect the fresh-build pipeline never writes

**Filed:** 2026-09-02, themes lane. Surfaced by the `gamedesign.uk` lane while
pre-seeding a palette on my advice; the mechanism below is my own verification of
their datum, which turned out to be one level deeper than either of us had it.

**090 substitute** (per the 2026-07-31 owner ruling, since this asserts a structural
claim): no 090 run. Substituted first-hand verification, all of it below — both step
configs read from the live `agent_definitions`, the sending key read from `082`, a
census of every live writer of either aspect, and a row census. Each is a command a
reader can re-run.

## 1. The claim

`extractPaletteSignal` (`platform/orchestration/actions/resolve_composition_pallette_action.go`)
documents its own rung 1 as *"mission.preferred_palette (human pre-specified)"* — the
most authoritative signal in the cascade. It reads `site_specs` aspect **`mission`**.

**Nothing in the fresh-build path ever writes aspect `mission`.** The human's brief
lands in aspect **`mission_brief`**. So the cascade's top rung is unreachable by the
pipeline that builds sites, and a human's stated colour preference cannot reach the
palette picker through the normal route.

**The same applies to `extractTypographySignal`'s `mission.preferred_typography`, and
this half is measured too, not inferred.** `resolve_composition_typography_action.go:214`
reads `loadSpecAspectFromContext(..., "mission", ...)` and returns `"mission_hint"` on a
hit — same aspect, same dead rung. Independently code-read by the `site_design_planner`
lane (who own the file) before it was written here. Census of the same 31 compositions:

| typography_source | count |
|---|---|
| `fingerprint_font_family_match` | 30 |
| `fallback_sans_modern` | 1 |
| **`mission_hint`** | **0** |

So it is not one rung. **Both cascades' human-preference rungs have never fired.**

## 2. Mechanism, verified

`082_submit_domain_unified.sh:143` sends the brief under `mission_brief`:
```sh
if [ -n "$MISSION" ]; then INPUT_DATA="${INPUT_DATA},\"mission_brief\":{\"text\":\"$MISSION\"}"; fi
```

`domain-submitter` has two steps, and the second is the first's `error_step`:

| step | reads | writes aspect |
|---|---|---|
| `persist_mission` | `input_data.mission` | `mission` |
| `persist_mission_brief` (error_step of the above) | `input_data.mission_brief` | `mission_brief` |

So on the fresh path: `persist_mission` finds nothing at `input_data.mission`, falls to
its error step, and `persist_mission_brief` writes the brief to a **different aspect**.
The `mission` aspect is never populated. This is not a key typo with a working
fallback — the fallback writes somewhere else, and the cascade only reads the first.

> **CORRECTED 2026-09-02 — "fallback" over-reads what `error_step` is here.** Mapping
> all four persist steps shows `error_step` is a **linear continuation chain**, not a
> set of designed recovery pairs — each step's `error_step` is simply the NEXT step in
> sequence:
>
> | step | reads | writes aspect | error_step |
> |---|---|---|---|
> | `persist_mission` | `input_data.mission` | `mission` | `persist_mission_brief` |
> | `persist_mission_brief` | `input_data.mission_brief` | `mission_brief` | `persist_roadmap` |
> | `persist_roadmap` | `input_data.roadmap` | `roadmap` | `persist_roadmap_brief` |
> | `persist_roadmap_brief` | `input_data.roadmap_brief` | `roadmap_brief` | `create_research_item` |
>
> So mission-being-rescued-by-mission_brief is **incidental ordering, not design**: the
> chain just continues past a failure and the next step happens to be the right one.
> Nothing here was built as a fallback pair, which matters for the fix — there is no
> "intended" pairing to preserve.

```sql
-- every live writer of either aspect
SELECT ad.type, s.name AS step, step->'config'->>'aspect' AS aspect,
       step->'config'->>'spec_data' AS reads
  FROM agent_definitions ad, jsonb_each(ad.default_config->'workflow'->'steps') AS s(name, step)
 WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
   AND step->'config'->>'aspect' IN ('mission','mission_brief');
```
→ `domain-submitter/persist_mission` → `mission` (from `input_data.mission`)
→ `domain-submitter/persist_mission_brief` → `mission_brief` (from `input_data.mission_brief`)
→ `brief-writer/persist_brief` → `mission_brief` (from `brief.result`)

**Two writers of `mission_brief`, one writer of `mission`, and that one reads a key the
submit script does not send.**

## 3. Live census [MEASURED 2026-09-02]

```sql
SELECT aspect, count(DISTINCT site_id) AS sites,
       count(*) FILTER (WHERE data ? 'preferred_palette') AS with_pref_palette
  FROM site_specs WHERE aspect IN ('mission','mission_brief') AND is_current GROUP BY aspect;
```

| aspect | sites | with `preferred_palette` |
|---|---|---|
| `mission_brief` | 22 | 0 |
| `mission` | **2** | **1** |

And of those 2: one is `gamedesign.uk`, hand-seeded an hour before this filing by the
lane that surfaced the datum. So **one pre-existing site fleet-wide** carries the
aspect the top rung reads, and the single `preferred_palette` in existence is the one
we just wrote by hand.

The rung is not "rarely used". It has never been reachable through the front door.

**And it has never fired. [MEASURED 2026-09-02]** — every resolved composition on the
estate, grouped by which rung won:

```sql
SELECT data->'lineage'->>'palette_source', count(*) FROM site_specs
 WHERE aspect='resolved_composition' AND is_current GROUP BY 1;
```

| palette_source | count |
|---|---|
| `design_intent_values` | 30 |
| `archetype_default` | 1 |
| **`mission_hint`** | **0** |

`mission_hint` is the string `extractPaletteSignal` returns when rung 1 wins. It does
not appear once in 31 compositions. This is the discriminating measurement: the claim
is not "the rung is hard to reach", it is "the rung has never been taken, and the code
path that would record it having been taken has no rows".

## 3a. The bug has a DEMAND CONTROL — it writes an error row on every fresh submit

Credit: `gamedesign.uk` lane, who noticed the trace; counts verified here.

Every fresh `082` submission leaves `agent_error_log` rows at the submitter —
`write_site_spec: input extraction failed: missing required fields: [spec_data]`,
which is what a persist step does when its `spec_data` path resolves to nothing.
[MEASURED 2026-09-02, 30-day retention]:

| step_name | rows | sites | first | last |
|---|---|---|---|---|
| `persist_mission` | 16 | 12 | 2026-08-04 | 2026-09-02 |
| `persist_roadmap` | 16 | 12 | 2026-08-04 | 2026-09-02 |
| `persist_roadmap_brief` | 14 | 11 | 2026-08-04 | 2026-09-02 |
| `persist_mission_brief` | 6 | 3 | 2026-08-18 | 2026-08-26 |

**This is the before/after meter, and it is the thing that makes a fix falsifiable.**
```sql
SELECT count(*) FROM agent_error_log
 WHERE agent_type='domain-submitter' AND step_name='persist_mission';
```
It must go to zero on the first fresh submission after a repoint. **If it does not,
the repoint did not land** — which is exactly the check this estate's own rule about a
post-fix zero needing a demand control asks for: the count is non-zero today, so a
later zero means something, instead of being indistinguishable from "nothing ran".

**Note the fourth row: `persist_mission_brief` itself fails on 3 sites.** So even the
incidental rescue in §2 is not reliable — on those sites neither `mission` nor
`mission_brief` was written by the submitter at all.

## 3b. The roadmap half is worse: TWO steps that can never succeed on this path

`082` sends **no roadmap key whatsoever** (`grep -c roadmap
082_submit_domain_unified.sh` → **0**). So `persist_roadmap` (reads
`input_data.roadmap`) and `persist_roadmap_brief` (reads `input_data.roadmap_brief`)
both fail on every fresh submission, and unlike mission there is no later step that
writes either aspect. The chain simply continues to `create_research_item`.

Live rows confirm nothing lands from this path: `roadmap` on **1** site,
`roadmap_brief` on **4** — against 22 sites carrying `mission_brief`. Those few came
from some other producer, not from `082`.

**So the fix is not one step, it is three** — and two of them are dead steps on the
only path that runs them. Either `082` should send a roadmap, or
`persist_roadmap`/`persist_roadmap_brief` should be deleted and say so. Leaving two
always-failing steps in a live workflow is how an error log becomes unreadable: 30 of
the 52 rows here are from steps that cannot succeed by construction.

## 4. Why nobody has noticed

The cascade degrades silently and sensibly: rung 1 misses, rung 2
(`design_intent.palette.reference_values`) hits, and every site gets a palette. The
failure mode is an ABSENCE of human authority, not a visible defect — the site looks
fine, it just isn't listening to the one signal documented as most authoritative.
It is also self-concealing in review: `bugs_open/113` already recorded the symptom on
one site (*"this site has no `mission` aspect, so it cannot fire"*, line 806) and read
it as a property of that site rather than of the pipeline.

## 5. Fix candidates, ranked

1. **Point `persist_mission` at the key that is actually sent.** Change its
   `spec_data` to `input_data.mission_brief` (config, live on apply, no roll). One
   step's config; the fallback keeps `mission_brief` populated for its own consumers.
   ⚠ **This makes the rung live fleet-wide the moment it applies** — every future
   fresh build would start feeding its brief into rung 1, which is the intent, but it
   is a behaviour change on a shared seam and should be said out loud, not slipped in.
   ⚠ It would also start OVERWRITING any hand-seeded `mission` row on rebuild — see §6.
2. **Point the cascade at the aspect that exists.** Have `extractPaletteSignal` read
   `mission_brief` as well as `mission`. Wider blast radius (touches the resolver, and
   `mission_brief` is free-form brief prose written by two different producers, so its
   shape is not guaranteed to carry a `preferred_palette` map at all).
3. **Decide the rung is dead and say so.** If a human's palette preference is not
   meant to arrive this way, delete the rung and its doc comment rather than leave a
   documented lever that cannot be pulled. Cheapest, and honest.

Not recommended: leave it. A documented "most authoritative" input that the pipeline
cannot supply is the shape this estate keeps rediscovering (`generic_theme`'s
`aspect='webdesign'`, 0 rows fleet-wide, is the same defect one table over).

## 6. Consequence to check before fixing — a hand-seeded lever depends on this

`gamedesign.uk` (site `8f17eb73-fc74-4718-8371-b3125bc4e414`) was deliberately seeded
with `mission.preferred_palette` on 2026-09-02 precisely because nothing overwrites
that aspect.

> **CORRECTED 2026-09-02, hours after filing — this section over-stated the hazard.**
> It read: *"Fix candidate 1 removes that property: once `persist_mission` reads the
> sent key, a rebuild would overwrite the hand-seeded row."* **That is wrong, and the
> correction narrows the tripwire to something much more specific.** Caught by the
> `gamedesign.uk` lane, verified here by reading the merge:
>
> `WriteSiteSpecAction` **always deep-merges** (`site_spec_actions.go:246`,
> `merged := siteSpecDeepMerge(currentData, specMap)`), and
> `siteSpecDeepMerge` recurses only when the value is a map on BOTH sides — otherwise
> `result[k] = srcVal`. A brief arrives as `{"text": "…"}`, so a repointed
> `persist_mission` **ADDS `text` beside `preferred_palette`** and leaves the palette
> untouched. A hand-seeded palette therefore SURVIVES fix candidate 1.
>
> **The real hazard is narrower and easier to walk into: a fix that does not go through
> `write_site_spec`.** Any supersede-without-merge writes a fresh row and drops every
> key it does not carry. That pattern exists in this codebase — `apply_theme_kit_action.go`'s
> own `supersedeAndInsertSpecWhole` is one, deliberately, for a point-in-time lineage
> record. **So the instruction to a fixer is not "beware overwriting", it is: keep
> using the merging writer, and do not "tidy" the merge away as redundant.**
>
> Also measured: gamedesign.uk's `domain-submitter` run COMPLETED at 17:07:57Z, so
> `persist_mission` cannot re-run on that site at all without a fresh `082` submit.
> Its `mission` row still carries `preferred_palette`, still has no `text`.

**A separate, live confirmation from the same site — `site_specs.pinned` does not
survive a supersede, in production.** The manual `design_intent` row was written with
`pinned=true`; the classifier wrote its own `design_intent` at 17:11:32Z and the
pinned row is now `is_current=false`. `pinned` protected nothing. This is the
production evidence for a claim previously only read out of the code (neither
`WriteSiteSpecAction`'s INSERT nor `apply_theme_kit`'s carries `pinned` forward), and
it is why the sanctioned human lock is an in-data key (`design_intent.<dim>.locked`)
rather than that column. Anyone reaching for `pinned` to hold a design value should
read this paragraph first.

**The general lesson, which outlives this bug:** the advice "seed X, it survives" was
resting on a defect, and when the defect's shape was examined properly the advice
turned out to be right for a *different* reason than given. Both the original warning
and the original reassurance were sound conclusions from unsound mechanisms. Check
which writer touches the row, not whether something "overwrites".

## 6a. A live test of this bug is IN FLIGHT — read its result before acting

`gamedesign.uk` is building right now (dispatched 17:07:55Z, corr
`f07313f6-976c-4593-9e5e-44892008fb74`) and is the first site in the estate's history
with a populated `mission.preferred_palette` at composition time. Its
`resolved_composition.lineage.palette_source` is therefore a genuine experiment, and
the two outcomes mean opposite things for this bug:

- **`mission_hint`** → rung 1 fired for the first time in production. §3's zero was
  caused by the aspect being empty, exactly as this file argues, and the diagnosis
  holds.
- **`design_intent_values`** → rung 2 won *even with a populated `mission` row
  present*. That would be a **stronger and different** finding than this file makes:
  it would mean rung 1 does not read what is actually there, and §2's mechanism is
  incomplete. **This file would need reopening, not closing.**

Experiment framed by the `gamedesign.uk` lane, who will report the value either way.
**Do not act on a fix candidate until it lands** — one of the two outcomes invalidates
the diagnosis.

⚠ Note the classifier has already deep-merged its own `design_intent` over the seeded
one (17:11:32Z), landing within two hex steps of the seeded palette (`#F5F0E8` vs
`#F4F1EA`, accent `#9B4E2A` vs `#A6521F`). So rung 2 now holds near-identical values
to rung 1 — which makes the test cleaner, not muddier: the two rungs cannot be told
apart by the COLOURS, only by the `palette_source` string. Read the string.

## 7. How to verify a fix

- Re-run §2's writer query: a fixed `persist_mission` should read a key `082` sends.
- Build one fresh site with a colour preference in its brief and confirm
  `resolve_composition_palette` returns `source: "mission_hint"` — that string is the
  proof the rung fired, and it has almost certainly never appeared in production.
  ```sql
  SELECT data->'lineage'->>'palette_source', count(*) FROM site_specs
   WHERE aspect='resolved_composition' AND is_current GROUP BY 1;
  ```
  A `mission_hint` count of 0 today is the baseline this should move off.
- Cross-check `bugs_open/113` (same family, different mechanism: that one is about
  generated palettes inheriting the layout's light literals) so the two are not closed
  against each other's evidence.

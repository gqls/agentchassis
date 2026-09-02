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

The same applies to `extractTypographySignal`'s `mission.preferred_typography`.

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
that aspect. **Fix candidate 1 removes that property**: once `persist_mission` reads
the sent key, a rebuild would overwrite the hand-seeded row with the submitted brief.
Tell that lane before applying it. This is the general hazard, not a special case —
any advice of the form "seed X, it survives" is resting on this bug.

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

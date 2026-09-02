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

> **This table is the BASELINE as at 2026-09-02 17:1xZ, and it has since moved — by
> one, deliberately.** §6a's test seeded `mission.preferred_palette` by hand and the
> next composition resolved `mission_hint`, taking the count 0 → 1. The zero above is
> preserved as the pre-test state because it is what made the test meaningful; do not
> re-read it as current. Current: `design_intent_values` 30, `archetype_default` 1,
> `mission_hint` 1.

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

## 6a. RESOLVED 2026-09-02 17:38Z — the test ran and the diagnosis HOLDS

**`palette_source = mission_hint`.** Rung 1 fired, for the first time in the estate's
history. Verified independently of the reporting lane:

```
palette_source | layout_source |    layout     |            palette
---------------+---------------+---------------+--------------------------------
 mission_hint  | library_match | magazine-grid | palette-gamedesign-uk-a6a70287
```

**Doubly discriminated — the string AND the hex agree, and either alone would have
been weaker.** The landed palette row is `#F4F1EA` / `#A6521F` / `#33302B`: the
hand-seeded `mission.preferred_palette` values byte-for-byte. The classifier's rung-2
`design_intent` values, which sat two hex steps away (`#F5F0E8` / `#9B4E2A`), are NOT
what landed. So even if the lineage string were untrustworthy, the colours prove which
rung won.

Fleet baseline moved for the first time: `design_intent_values` 30, `archetype_default`
1, **`mission_hint` 1** — and that 1 is this test.

**What this settles:** rung 1 is not broken, unreachable-by-design, or misreading its
input. It reads aspect `mission` exactly as the code says. It has never fired because
**nothing has ever put anything there** — which is §2's mechanism, confirmed. The
alternative outcome (rung 2 winning with a populated `mission` row present) would have
meant the mechanism was wrong and this file needed reopening. It did not happen.

**Fix candidates are therefore safe to act on** — the diagnosis they rest on is
confirmed. §6's corrected guidance (keep the merging writer) still applies.

### 6a-bis. The same defect one aspect over — layout preference did NOT survive

Worth recording because it generalises this bug rather than repeating it. The seeding
lane also set `design_intent.layout_preference = "soft-editorial"`. The composition
chose **`magazine-grid`** instead, via `library_match` on tags — because
`layout_preference` lived in the `design_intent` row that the classifier **superseded**
at 17:11:32Z. The tag matcher was left with only `style_direction` prose and picked a
same-category neighbour (both are `editorial`; `magazine-grid` has an empty `scheme`,
so the light/dark constraint did not bind either).

Under the 2026-09-02 ruling that is the planner exercising legitimate authority, not a
defect. But the *reason* the preference vanished is this file's argument one aspect
over: **a seeded value in `design_intent` is not durable, because a later writer
supersedes the whole row.** Palette survived only because it was seeded into `mission`,
which nothing writes. So:

- The `locked`-key argument (§6) is not palette-specific. Any seeded design preference
  that must outlive a classifier pass needs either an aspect nothing overwrites, or an
  in-data key a merging writer carries forward.
- Anyone fixing this bug by making `mission` writable should note they would be
  removing the ONE aspect that currently has that property, which is what made this
  test possible at all.

## 6b. (superseded) The test as it was framed before it ran

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

## 6c. `site_design_planner` lane: fix candidate 1, as specified, does not do what §7 claims — and it would regress the one producer that already works

Verified rather than assumed, before touching anything, because §6a's success made
candidate 1 look safe to act on and it is not, for a reason nobody in this file has
named yet.

**The pre-existing site behind §3's "one pre-existing site fleet-wide" was never
named. It is `vonc.com`** (`site_specs`, `aspect='mission'`, `created_by=domain-submitter`,
`created_at=2026-06-22`, still `is_current`). Its `mission` row does NOT carry
`preferred_palette`/`preferred_typography` (checked: 8 keys, all `mission`/`tagline`/
`positioning`/`content_tone`/`target_users`/`core_concepts`/`key_differentiators`/
`measurable_objectives`, none of the two preference keys), and its
`resolved_composition.lineage.palette_source` is `design_intent_values` — so this does
**not** contradict §3's "`mission_hint` never fired before gamedesign.uk"; it explains
the "1 pre-existing" row and nothing more.

**What it does reveal: `persist_mission`'s `spec_data: "input_data.mission"` is not
unreachable-by-construction. It has a real, working producer.**
`scripts/initial_messages/210_vonc_trigger/080_submit_vonc.sh` (a "Tier 3 submission:
domain + mission + roadmap + briefs", target `agent_type: domain-submitter`, same
agent §2 maps) sends a **rich, structured `input_data.mission` object** — 8+ fields,
`objective`/`positioning`/`tagline`/`key_differentiators`/`measurable_objectives` — in
the SAME payload as `mission_brief` (both keys present, lines 62 and 65 of that
script). `persist_mission` reads exactly the key this script sends and it worked,
first try, 2026-06-22. **`082` is not the only submission path; it is the one
standard-framework path, and it is the one that never sends `input_data.mission`.**
Grepped `scripts/` for any other structured `"mission": {` sender: none — this is a
single bespoke script, not a second framework path, but it is real and live
(`vonc.com` is a deployed site).

**Consequence for candidate 1.** `extractPaletteSignal` checks
`mission["preferred_palette"]` as a `map[string]interface{}` — an exact key, an exact
shape. `082` never sends that shape under any key: `--mission`/`--mission-file` both
resolve to a single free-text string, wrapped as `{"text": "..."}"`, under
`mission_brief` (`082_submit_domain_unified.sh:105-125,143`; grepped the whole script,
no `--palette`/`--colour` flag, no structured alternative). **Repointing
`persist_mission`'s `spec_data` to `input_data.mission_brief` would therefore write
`{"text": "..."}"` into aspect `mission` — a string, never a `preferred_palette` map —
and `extractPaletteSignal`'s check would still find nothing.** §7's own verification
step ("Build one fresh site with a colour preference in its brief and confirm
`resolve_composition_palette` returns `source: "mission_hint"`") **would fail if run
as written** — a free-text brief cannot satisfy a structured-map check, repointed or
not. This is not a smaller version of the bug; it is a different bug candidate 1 does
not touch.

**And it has a second cost, unflagged so far: it would silently stop capturing
`vonc.com`-style structured mission data.** A Tier-3-style script sends BOTH
`input_data.mission` (rich) and `input_data.mission_brief` (free text) in one payload.
After candidate 1, `persist_mission` reads `mission_brief` instead — so the rich
object such a script sends would land nowhere (still readable at
`input_data.mission`, but nothing persists it any more), for the sake of relocating a
free-text blob that was already going to land in `mission_brief` via the existing
error-step chain regardless. **Net effect of candidate 1 as specified: solves nothing
for `082`, and quietly breaks the one producer that currently works.**

**What this changes about the fix candidates:**

- **Candidate 1 needs correcting, not just applying.** `persist_mission` reading
  `input_data.mission` is *correct* for the producer that actually sends that shape —
  the bug is not the read, it's that the only standard-framework submission path
  (`082`) has no way to populate it. The fix this estate probably wants is closer to
  "teach `082` to optionally send a structured `mission` (or at minimum a
  `preferred_palette`/`preferred_typography` sub-object) when a caller has one" — an
  addition to the sender, not a repoint of the reader. That is a different, larger
  change than "one step's config", and needs its own design (a CLI flag? a
  `--mission-json` file merged in? does an LLM ever get asked to extract structured
  preferences from free text, and should it?) — not decided here.
- **Candidate 2** (read `mission_brief` from the cascade side) inherits the same
  shape problem one level up — `mission_brief` is free text; the resolver would still
  need something to turn "please use blue and gold" into a colour map, which does not
  exist anywhere in this pipeline today (checked: no LLM step reads `mission_brief`
  and writes `preferred_palette` to any aspect).
- **Candidate 3** (retire the rung) is **the only one of the three that is still
  correct as originally written** — and is arguably strengthened by this: if nobody
  builds the missing extraction/CLI capability, the rung stays permanently
  unreachable from the standard path regardless of which key `persist_mission` reads,
  which is a stronger case for saying so plainly than the file already made.

**Not touching any of this from here.** This corrects the shape of the decision, it
does not make it — the missing capability (should `082` be able to carry a structured
preference at all) is a real design question, not a bug to patch by config. Flagged
to `theme kits` directly.

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

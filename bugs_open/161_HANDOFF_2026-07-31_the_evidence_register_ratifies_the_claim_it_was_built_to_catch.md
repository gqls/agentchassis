# 161 — the evidence register ratifies the claim it was built to catch

**Filed 2026-07-31** while working `bugs_open/HANDOFF_2026-07-31_checker_layer_remaining_items.md`
§1 (149's C1 recognition gap). It answers §1's question and **falsifies two of its three
fix candidates**, so read this before acting on that file.

**Status: OPEN, unowned. Structural + cross-cutting. The repair is partly an OWNER CALL
(the register is human-owned by design), so nothing is changed by this filing.**

---

## The one-sentence version

A site's evidence register is simultaneously (a) the whitelist of claims the writer is
**instructed** to make and (b) the authority every claims gate checks published claims
**against** — so a false fact in the register is **self-ratifying**: it causes the claim,
then vouches for it, and no gate in the layer can ever object.

`gamesdesign.co.uk` has one. It has been live since 2026-07-24 and is **still served today**.

## What is false

`site_specs` / `aspect='evidence_base'` / `is_current`, fact `gd-trials`:

```json
{"id":"gd-trials","claim":"Monte Carlo trials per query","value":10000,"kind":"metric",
 "tolerance":"exact","context_terms":["trial","monte carlo","simulation"],
 "source":{"artifact":"the figure is hard-coded in the shipped drop-rate tool JavaScript"},
 "verified_at":"2026-07-24",
 "writer_line":"{value} Monte Carlo trials per query in the drop-rate tools"}
```

**Neither drop-rate tool performs any Monte Carlo simulation, and neither contains any
randomness at all.** [VERIFIED 2026-07-31, by reading the shipped component HTML/JS out of
`page_components.rendered_html`, export length asserted against the DB's own `length()`]

| tool | page id | what it actually computes | `Math.random` | `monte` | the `10000` in it |
|---|---|---|---|---|---|
| `tool-drop-rate-tuner` | `b381f0db-…` | `Math.pow(1-p, k)` — closed-form geometric survival, then a CDF array indexed by kill count (`:452-463`) | **0** | **0** | none — the apparent match is `if (pity <= 0 \|\| pity > 100000)` |
| `tool-drop-rate-simulator` | `0f9ed454-…` | says "binomial" 6× | **0** | **0** | `return Math.min(val, 10000)` — an **input clamp** on attempts |

A Monte Carlo method *is* random sampling. With zero randomness in either artefact there is
no trial count to be "hard-coded", so the fact's own cited source does not support it.

**The number 10,000 is real; its stated MEANING is wrong.** It is the maximum attempts the
simulator will model, which is exactly what another session independently concluded when it
repaired the homepage on 2026-07-30 (see the timeline).

## It was false when it was registered — not stale since

This matters, because "went stale" would be `refresh_evidence_base`'s job and "false on
arrival" is nobody's.

- `tool-drop-rate-simulator`'s component was created **2026-06-05 20:12:52** and its
  `updated_at` is **the same timestamp** — untouched for seven weeks before the register was
  written, and untouched since. So the JS I read today is byte-for-byte the JS that existed
  on 2026-07-24. [VERIFIED]
- `tool-drop-rate-tuner` has **zero rows** in `page_component_history` mentioning
  `monte carlo` or `Math.random` — the tool artefact never carried either. [VERIFIED]

## The sequence, and it runs the opposite way round to the obvious guess

Timestamps are from `page_component_history.created_at` and `site_specs.created_at`.

| when | what |
|---|---|
| **2026-06-06 16:59** | gamesdesign's homepage copy already asserts "Monte Carlo" — **seven weeks before any register existed**. Origin is outside this bug. |
| 2026-06-22 → 2026-07-20 | four more saves, claim carried forward each time |
| **2026-07-24 14:20:34** | the homepage is saved again, still asserting it |
| **2026-07-24 14:22:40** | `bugs_closed/043`'s wave-1 audit **creates the register** — 2 minutes and 6 seconds later — registering "Monte Carlo trials per query" as a verified fact and attributing it to the shipped tool JavaScript |
| 2026-07-29 17:26–18:10 | the rewrite wave witnessed in `149` C1 restates it across many components |
| **2026-07-30 14:31** | another session repairs the homepage stat card to `stat2_label:"Max Attempts Modelled"`, description "computes **exact binomial probabilities** for up to ten thousand". **That session got the truth right.** But it corrected the copy only — not the register, not the migration |
| **today** | the register still says "Monte Carlo", so it now **contradicts the repaired page**, and will re-supply the false line to the next rewrite |

**So the audit that was supposed to give writers "a register of true ones to reach for"
(`bugs_closed/043`) appears to have derived this fact from the very copy it was auditing,
and then cited an artefact that does not contain it.** The two-minute gap is
circumstantial, not proof of method — but the artefact evidence above is decisive on the
outcome either way: what got registered was not what the artefact said. [VERIFIED on
outcome; the derivation-from-copy route is [INFERRED] from the timing]

## Why no gate can catch it — the structural part

Both halves of the layer consult the same register, so both are disarmed by the same row.

**1. It is fed to the writer as an instruction.** `refresh_evidence_base_action.go:16-18`
— `writer_block` is "consumed by the page-content-writer prompt … so the numbers the writer
is permitted to assert can never quietly rot". gamesdesign's live `writer_block` (600 bytes)
reads, verbatim:

```
NUMBERS (state only these, with their listed meaning; as of 2026-07-24):
- 11 interactive design tools live, all client-side and free
- 10,000 Monte Carlo trials per query in the drop-rate tools (the figure is in the shipped tool code)
- 4 configurable inputs in the drop-rate tuner: drop chance, kills per hour, pity timer, target hours
- 10 guides & articles live (5 blog posts + 5 guides)
NOT TRACKED / DOES NOT EXIST, NEVER STATE: user counts, accuracy-gap percentages, …
```

The writer did **not** invent this claim. It was handed to it under "state only these",
with a parenthetical asserting the artefact backs it.

**2. It is the authority the gates check against.** `datahelpers/claims.go:931`
`numberSupported(10000, window)` walks `eb.Facts`; `gd-trials` has `context_terms`
including `"monte carlo"`, the window contains it, tolerance is `exact`, `10000 == 10000`
→ **supported → skipped**. Every consumer of that one function is disarmed at once:

- the prose scan (`ScanUnregisteredNumbers` → `claims.go:778`)
- the stat-field audit (`claims_stats.go:327`, `ScanStatClaims`)
- the persistence floor shipped for 149 C1 (`save_sections_claims_guard.go`, same engine by
  design — `:110-114`)

**This is why the handoff's §1 measurement returned 0 findings, and the 0 was CORRECT.**
§1 read that silence as a vocabulary blind spot in `businessClaimContextRe`. It is not:
the number is registered, and a scan that skips a registered number is working.

## Consequence for the handoff's §1 candidates — two of the three are inert

Measured against the motivating case, not argued:

1. **"Widen `businessClaimContextRe` to cover technical/product vocabulary" — INERT.**
   Widening makes the number *reach* `numberSupported`, which then matches `gd-trials` and
   correctly skips it. Net effect on the witnessed fabrication: **zero**. It also buys the
   false-positive risk `claims.go:612-617` already records from the last widening. **Do not
   spend a council round on this.**
2. **"A structural stat rule rather than a lexical one" — INERT, and largely already built.**
   `ScanStatClaims` is *already* purely structural (no lexical gate at all) — §1's premise
   that this needs building is wrong. It also calls `numberSupported`, so it is disarmed by
   the same row. On §1's sharper sub-question — *"check whether the stat audit ran on that
   page and what it said; if it stayed silent that is a more serious finding"* — the answer
   is that it would have stayed silent **and been right to**, for this claim.
3. **"A claim-vs-source diff at rewrite time" — the only one that survives**, and only for
   the *other* half of the witnessed case (below). It would not flag the Monte Carlo claim,
   which is unchanged from the register's own words.

## The other half of the witnessed case still stands

§1's second bullet is correct and untouched by this bug: **"built *by* a shipped
live-service designer"** (from "built *for* live-service and tabletop designers") is a
fabricated **human credential** — non-numeric, matching no banned pattern. gamesdesign has
**0 `banned_claims`**. [VERIFIED] The engine has no shape for a qualitative personnel claim.
That remains a real gap and is *not* what this bug is about.

## Blast radius — measured fleet-wide, 2026-07-31

```sql
SELECT s.domain, jsonb_array_length(COALESCE(sp.data->'facts','[]'::jsonb)) AS facts,
       (SELECT count(*) FROM jsonb_array_elements(sp.data->'facts') f
          WHERE NOT (f->'source' ? 'query' OR f->'source' ? 'sql')) AS prose_sourced,
       COALESCE(sp.data->>'writer_block_managed','(absent)') AS wb_managed
FROM site_specs sp JOIN sites s ON s.id=sp.site_id
WHERE sp.aspect='evidence_base' AND sp.is_current ORDER BY 2 DESC;
```

| domain | facts | prose-sourced | wb_managed |
|---|---|---|---|
| oufe.com | 36 | 35 | (absent) |
| leopardessconsulting.co.uk | 18 | 9 | true |
| fundamentallyai.com | 15 | 8 | true |
| relojistas.com | 13 | **13** | (absent) |
| ai-agent-orchestration.com | 7 | 2 | (absent) |
| robot-hands.com | 5 | 2 | (absent) |
| **gamesdesign.co.uk** | 4 | **4** | (absent) |
| vonc.com | 4 | 0 | (absent) |
| finetuning.uk | 0 | 0 | (absent) |

**9 registers · 102 facts · 73 (72%) prose-sourced.** `refresh_evidence_base` only refreshes
a value from a `sql`/`query` source, and it "never rewrites the human-authored WORDS"
(`refresh_evidence_base_action.go:20-23`) — so **no mechanism re-checks a prose-sourced fact,
ever, and none checks any fact's `claim` wording against its artefact.**

Confirming detail: the live `stale_evidence` work items name **oufe, leopardess,
ai-agent-orchestration, fundamentallyai** — precisely the four registers that have
sql-sourced facts. gamesdesign, relojistas, vonc and robot-hands can never raise one.
[VERIFIED] The freshness mechanism is structurally blind to 4 of 9 registers, and
`relojistas` (13/13 prose) and `gamesdesign` (4/4) are wholly outside it.

## What is still being served

`page_components`, `build_status='deployed'`, 2026-07-31:

| page | slot | text |
|---|---|---|
| `tool-spawn-rate-balancer-guide` (`blog-post`) | `call-to-action` | "The drop-rate tuner runs **10,000 Monte Carlo trials per query**. You set the 4 configurable inputs…" |
| `tool-spawn-rate-balancer-guide` | `article-body` | contains it |
| `tool-loot-table-balancer-guide` (`blog-post`) | `article-body` | contains it |

Both pages are `page_type='blog-post'`, which `editorialPageTypes` (`claims.go:721-723`)
exempts from the prose number scan — so even correcting the register does **not** make these
three components get flagged. They need a direct repair. The homepage is already correct.

## Fix candidates — ordered by what makes the bad state unrepresentable

**The register is human-owned by design ("Truth decisions stay human",
`refresh_evidence_base_action.go:20`), so 1 and 2 are OWNER CALLS, not a thread's to take.**

1. **Correct the fact, in both places, or a reseed reintroduces it.** The false row is in the
   **committed migration**: `docs/agent_docs/sql_for_agents/218_evidence_facts_for_043_sites.sql:139-143`.
   Fixing only the live row leaves the repo able to restore it. Proposed correction, matching
   what the artefact and the 07-30 repair both say: `claim` → "maximum attempts modelled per
   query", `writer_line` → "{value} maximum attempts modelled in the drop-rate tools",
   `context_terms` → drop `"monte carlo"` and `"simulation"`, keep `"attempt"`/`"trial"`,
   `source.artifact` → "`Math.min(val, 10000)` input clamp in tool-drop-rate-simulator".
   **Doing this ARMS the gates against the three surviving components** — that is the point,
   but it means the next build on those pages starts objecting, so sequence it with 3.
2. **Then repair the three deployed components.** They assert a technique the tools do not
   use. `blog-post` exemption means no gate will raise them; they need naming explicitly.
3. **Make "the artefact backs this" checkable rather than prose.** The generalisable fix, and
   the reason this is a bug and not a typo: today `source.attested_by` / `source.artifact`
   are free text that nothing can verify, while `source.query` is machine-checked every
   sweep. 73 of 102 facts are in the unverifiable class. Options, cheapest first:
   - a **`stale_evidence`-style item for prose-sourced facts on a cadence** ("this fact has
     never been machine-verified; re-attest it") — turns silence into a queue;
   - **`source.artifact_check`**: an optional grep-shaped assertion (path/table + a pattern
     that must be present) so an artefact-sourced fact can fail like a sql-sourced one. This
     is a **new shared vocabulary key on a shared mechanism** — architecture scope under
     CLAUDE.md's seam rules, RFC before code.
4. **Reject: "make the audit that seeds registers read the artefact."** Right instinct, wrong
   lever — it is a prompt/procedure change, and a procedure is not an enforcement mechanism.
   It also cannot help the 102 facts already registered.

## How to verify any fix

- **Unit, on `numberSupported`,** with `gd-trials` corrected: "10,000 Monte Carlo trials per
  query" must become **unsupported** while "10,000 maximum attempts modelled" stays
  supported. A fix asserted only on the second half is half a fix.
- **Mutation-check it** — revert and confirm the new test fails.
- **Re-run `cmd/claimscan` over gamesdesign** before and after. Note it scans prose and stat
  fields via `ParseEvidenceBase`, and **has no stat-claim path of its own** (`0` hits for
  `ExtractStatClaims`/`ScanStatClaims`; positive control `ParseEvidenceBase` = 12)
  [VERIFIED 2026-07-31] — so it predicts the prose half only. Do not read a clean claimscan
  as clearing the stat fields.
- **Serve-side, not stored-side**: confirm the three components are gone from what is
  actually served, per `bugs_open/`'s standing rule that stored HTML is not the artefact.

## Traps this cost me, now in LANDMINES.md

- **A registered fact makes a green claims gate meaningless as evidence of truth.** "The
  scan returned 0 findings" is compatible with "the claim is false and the register says
  otherwise". Always ask *which fact matched* before reading a pass as clearance.
- **`page_components` has no provenance column and `page_component_history.source` does not
  rescue it** — the column is `save_page_sections_overwrite` for **12,386 of 12,416 rows**,
  every pipeline writing the same literal; the rest are hand-typed operator strings.
  [VERIFIED] So the handoff §3's "`ApplySectionEditAction` cannot be bounded from
  `page_components`" **stands**, and history does not bound it either. Saves the next thread
  the query.
- **Grepping a tool for `10000` matches `100000`.** Both of this bug's apparent
  "the figure is in the code" hits were a `pity > 100000` bound and a `Math.min(val, 10000)`
  clamp. Print the match in context; a bare count would have confirmed the false fact.

## Relations

- `bugs_closed/043` — the parent. Its remediation seeded this register; its line 299 asserts
  "the writer_block lists each site's **true** countables", which is the assumption this
  bug falsifies. **Not reopening it** — 043's fix works; this is a defect in one seeded row
  and in the absence of any check on seeded rows.
- `bugs_open/HANDOFF_2026-07-31_checker_layer_remaining_items.md` §1 — answered here; two of
  its three candidates are inert.
- `bugs_open/149` C1 — the ⚠ banner's "the floor would not have caught it" is **correct but
  for the wrong reason**: not narrow vocabulary, but a register that vouches for the claim.
- `bugs_open/102` — the `blog-post` editorial exemption that keeps the three surviving
  components unflagged even after a register fix.
- `bugs_open/151` — "a rewrite is the moment unbacked claims get laundered". Same family,
  one level up: here the *register* does the laundering.
- `bugs_open/147` — robot-hands' "independently verified". Adjacent but different: that one
  the engine **caught**, because it is a banned claim rather than a registered one.
- `CLM-003` / `CLM-014` / `CLM-018` — concept register, `claims-verification.md`.

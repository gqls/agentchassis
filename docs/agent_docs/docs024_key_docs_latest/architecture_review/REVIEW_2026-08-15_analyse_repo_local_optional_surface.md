# REVIEW — `analyse_repo_local`'s accumulated optional surface (RFC_022 budget, second of the three standing findings)

**Reviewed 2026-08-15 at the owner's direction** ("direct both reviews now"), under the
2026-08-14 rulings: budget N=10 on shared actions; sharing is estate design, so this
reviews the ACCUMULATED SURFACE as a whole, never the reuse. Carried by **2 live
agents** (code-indexer, diagnose-agent) and declaring **12 optional keys** — the widest
shared surface in the fleet.

**VERDICT: ACKNOWLEDGE AT 12. No trims.** Baseline recorded in
`optional_key_budget_acks.json` and the cron's `ACKED_LEVELS` mirror.

## Method

Same as the worked example (`REVIEW_2026-08-14_append_doc_note_optional_surface.md`):
read the implementation and every key's read-site; censused live configuration with
`--live-pairs` over the fleet export; dated every key with `git log --follow` / `-S`.

## The surface, decomposed — 12 keys are 7 concepts, five doubled by a duality that is CONTRACT, not accretion

| concept | keys | read at | live-configured? |
|---|---|---|---|
| which repo (identity) | `owner`/`owner_field`, `repo`/`repo_field` | `resolveRAGConfigField`, `:112-113`; both empty is a refusal `:119` | `owner_field`, `repo_field` |
| which commit | `ref`/`ref_field` | `:114` | `ref_field` |
| which language | `language`/`language_field` | `:115` (empty → `"go"`) | `language` (literal) |
| which API host | `github_api_base`/`github_api_base_field` | `:153` | neither (defaulted) |
| reproducibility pin | `pin_to_index_commit` (default **true**) | `:232` (hand-coerced bool — no `GetBoolField` existed at time of writing, per its own comment) | yes |
| corpus filter | `exclude_patterns` | `:195` (default `defaultAnalyseExcludePatterns`) | no (default operates) |

The five dualities all resolve through **one shared helper**, `resolveRAGConfigField`,
and the file states why: *"Resolve identity the SAME way `request_repo_analysis` does …
drop-in config compatibility."* The doubled keys are a compatibility contract with a
sibling action, not independent capabilities — trimming either half of any pair would
break the drop-in property that is their stated reason to exist.

The two singletons are both load-bearing guards this lane knows first-hand:
`pin_to_index_commit` keys the analysis to the indexed COMMIT (the `bugs_closed/108`
family — the row clock must not launder a stale index), and `exclude_patterns` is the
key whose absence from a hand-run produced the **1,371-vs-1,204 wrong pass-mark** in
this lane's own §4a correction (`WRONG_CALLS.md` 2026-08-11) — the config key must
exist precisely so the exclusion set is stated, not folklore.

## Growth history — the whole surface is birth-era; zero additions in six weeks

Ten keys (the dualities) present from `c886c3d26` (2026-06-28) with
`github_api_base`/`pin_to_index_commit`; `exclude_patterns` landed `a4469553e`
(2026-07-02). **Nothing has been added since 2026-07-02.** The count reflects the
action's parameterisation (five ways to say which tree to fetch, times two flavours),
not creep.

## What would reopen this

A thirteenth key pages again (baseline 12) and the growth is architecture-scope under
the ruled trigger; an authority-bearing addition is architecture-scope regardless of
count. One improvement noted, not required: the hand-coerced bool at `:232` says to
adopt `datahelpers.GetBoolField` when it exists — `GetBoolFieldLoud` now does
(register WFA-010); adopting it would not change the key count.

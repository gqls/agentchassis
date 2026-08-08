# PLAN — 2026-08-08 — `bugs_open/136`: a declared alias vocabulary for literal config settings

**Bug file:** `bugs_open/136_HANDOFF_2026-07-28_live_config_says_domain_where_the_code_reads_pipeline.md`
— resolve **by slug**. `136` is one of the ambiguous numbers; the other 136 is
`section_editor_and_three_siblings_persist_links_with_no_repair`, owned by
`bugfix_136_sibling_link_repair`, and that lane's own plan says so in its header.

**Ownership check before starting** (2026-08-08 ~16:10 BST):
- `scripts/who-owns.py 136` names `bugfix_136_sibling_link_repair` (the OTHER 136),
  `bugfix_097_content_data_links` and `bugfix_100_101_scrape_provenance` (the filing lane,
  last commit on this bug **2026-07-29**, nine days ago).
- Live-session sweep of all 32 `.jsonl` transcripts touched in the last 5 hours for
  `bugs_open/136|check_domain|target_domain|summary_template|audit-config-keys`: the only
  session above incidental noise is **this one**. The next-highest (`b5a58a2b`, 20 hits) is
  the 206/083/220 lane and every hit is the auto-memory file
  `bugfix-136-domain-pipeline-rename.md` being read, not work.
- **Not taken and why:** `093` (its own last update says it is blocked on `083`, and 083 is
  hot in two sessions) · `211` (`who-owns` returns OWNED, `bugfix_122_contrast_ink_slots`
  committed to it today) · `114`, `126`, `146` (owning lanes active within 8 days) ·
  `085`, `181`, `185`, `189`, `203` (fixed and live; only site-level verification owed).

---

## 1. What is actually wrong, in one paragraph

Three actions were renamed internally from `domain` to `pipeline`. The live
`agent_definitions` kept the old word and **nothing in the fleet sets the new one**, so on
each of those steps the config is inert and the action silently takes its hardcoded default.
`create_work_item` is the exception: someone hand-wrote a three-line back-compat shim
(`create_work_item_action.go:118-121`) that makes the old name work. They wrote it on one
action and not on the other two.

**The framework reason nobody wrote the other two, which the bug file does not state, is the
finding this plan is built on.** `ActionInputSpec` already has a `Deprecated` field for
exactly this purpose — but it is honoured only in `ExtractActionInputs` Strategy 3
(`action_inputs.go:471-491`), which resolves *the old key's value as a dot-path into
collected_data*. It is a mechanism for **path-reference** aliases (`site_id_field` →
`site_id`). `check_pipeline` / `target_pipeline` / `item_pipeline` are **literal settings**
read straight from `params.StepConfig.Config` — precisely what `ConfigKeys` is documented to
mean — and **`ConfigKeys` has no alias vocabulary at all.** Declaring `check_domain: "content"`
in `Deprecated` would make the runtime try to resolve the literal string `"content"` as a
data path: silently wrong, and worse than the bug. So the only correct move available to an
author was to hand-roll Go, and two of the three did not.

## 2. The correction that changes this bug's priority

> **CORRECTED 2026-08-08 — `bugs_open/136` §2a's `[MEASURED] Nothing is mislabelled today`
> is now FALSE, and the file's own predicted trigger is what fired.**

§2a said the containment was "a coincidence of today's check-to-agent mapping" and that the
moment a check propagating `dctx.Pipeline` joined the content or build agent's list, findings
would be written under the wrong pipeline. Since 2026-07-28 two such checks have joined
`completeness-discovery-agent` — `content_duplication` and `page_canonical_collision`, both
of which set `Pipeline: dctx.Pipeline`. That agent's config asks for `check_domain: "content"`;
the code reads `check_pipeline`, finds nothing, and takes the default `"design"`.

```sql
SELECT id, item_type, pipeline, status, created_at FROM site_work_items
WHERE created_by='completeness-discovery-agent' AND pipeline='design';
--  page_canonical_collision | design | complete | 2026-08-04
--  page_canonical_collision | design | complete | 2026-08-04
--  capability_gap           | design | detected | 2026-08-04
--  capability_gap           | design | detected | 2026-08-03
```

**Four live rows, two of them still open**, filed under `design` by an agent whose config
says `content`. The harm is not cosmetic: `countDispatchableWorkItems`
(`work_items_common.go:198-211`) filters `AND pipeline = $2` to answer *"is there unfinished
promoted work on this site"*. A row filed under the wrong pipeline is invisible to the count
that should see it and inflates one that should not.

This is a latent-to-live transition, so the fix is now worth its roll. It is also why the
per-action fix is not enough on its own: the mechanism that lets an author declare the
alias is what stops the fourth instance.

## 3. Fix candidates, ranked by what makes the bad state unrepresentable

1. **The framework seam — a declared alias vocabulary for literal settings**
   (`DeprecatedConfigKeys` + one shared read helper + an audit section). Makes three states
   unrepresentable: an old-name key that is silently inert (it is now either
   honoured-and-reported-DEPRECATED, or reported UNKNOWN — never quiet); a hand-rolled shim
   invisible to the audit; and a malformed alias declaration (a fleet-wide spec-lint test
   fails CI). This is the bug file's own §2 recommendation — *"prefer the shim … downgrade to
   `Deprecated`"* — promoted from a third hand-rolled copy into the capability that was
   missing.
2. **Adoption on the three rename sites.** Closes the live instances *through* the mechanism,
   so the next `*_domain` author gets a warning, an audit line and working behaviour instead
   of a coincidence.
3. **Kill the audit's one false positive** (`execute_vision_prompt.ai_service`, which the
   action really does read, via the shared `resolveAIServiceConfig`). Reporting only — but a
   report carrying a known-false line is a report people learn to ignore.
4. **Definition edits** (renaming `check_domain` → `check_pipeline` in `agent_definitions`).
   **Deferred deliberately**: it fixes today's instances and prevents nothing (bug file
   §5.2), it edits other lanes' quiesced agents, and once the alias ships it is not needed
   for correctness. The new DEPRECATED audit section becomes the standing migration list.

**Explicitly NOT a candidate:** adding the old keys to `ConfigKeys`. That makes them
*recognised*, silences the detector and leaves the behaviour broken — the recorded
`WRONG_CALLS.md` 2026-07-28 mistake, committed by the fix for `101` itself.

## 4. The seam, concretely

| edit | file | changes BEHAVIOUR? |
|---|---|---|
| `DeprecatedConfigKeys map[string]string` on `ActionInputSpec` | `datahelpers/action_inputs.go` | no — inert until a spec declares one |
| `ResolveConfigSetting(config, spec, key, default, logger)` + `DeprecatedConfigKeysInUse` + `ListDeprecatedConfigKeys` | **new** `datahelpers/config_key_aliases.go` | no — new function, no caller yet |
| recognise alias keys in `UnknownConfigKeys` / `ListDeclaredConfigKeys` | `datahelpers/action_inputs.go` | no — reporting |
| `deprecated` map in the audit dump; `deprecated_config_keys` in `--specs` | `cmd/config-key-audit/main.go` | no — reporting |
| new `=== DEPRECATED KEYS ===` section, exit code unchanged | `scripts/audit-config-keys.sh` | no — reporting |
| `ai_service` declared | `actions/execute_vision_prompt_action.go` | no — reporting |
| alias declared + helper adopted | `actions/create_work_item_action.go` | **no** — identical truth table to the shim it replaces |
| alias declared + helper adopted | `actions/triage_detect_items_action.go` | no in value (`target_domain` is `"build"`, equal to the default); yes in provenance |
| alias declared + helper adopted | `actions/discovery_checks.go` | **YES — the one real behaviour change.** `completeness-discovery-agent`'s two propagating checks move from `design` to the `content` its config asks for |

Precedence in the helper is byte-for-byte the shim's truth table: canonical key (non-empty
string) → declared old names, sorted → default. A non-string value under any name is treated
as absent, pinning the `.(string)` semantics every existing call site already had.

**The deprecation warning fires inside the helper**, when an alias actually supplies the
value — the symmetric twin of the Strategy 3 warn at `action_inputs.go:485-489`.

## 5. Sequencing — each step independently committable

1. **`ai_service`** — one file, one test. Kills the false positive.
2. **The seam + audit surface + concept register + LANDMINES, one commit.** Entirely inert:
   no spec declares an alias yet, so audit output is unchanged. Registered in the **same
   commit that ships it** — condition (2) of the ordering exemption, which survived the
   owner ruling of 2026-07-29.
3. **First adopter: `create_work_item`.** Behaviour-preserving by construction; puts 9 live
   `item_domain` steps into the DEPRECATED section and earns the live witness.
4. **The two quiesced actions.** The actual 136 fix, with §2's measurement attached.

## 6. Verification, and where it runs out

- **The reporting half needs no roll.** `audit-config-keys.sh` runs `go run` from the tree
  against the live DB. After the commits, UNKNOWN KEYS must shrink from four entries to
  exactly `plan_sections: domain`, and the DEPRECATED section must name
  `run_discovery_checks.check_domain` (3 steps), `triage_detected_items.target_domain` (1),
  `create_work_item.item_domain` (9). That is live-DB evidence the declarations join real
  definitions.
- **The behaviour half, cheapest honest live proof:** `create_work_item` runs on live lanes
  today. After the roll, the same shared helper the two quiesced actions call is exercised by
  real traffic — grep the pod for the new warning literal and check the created row's
  `pipeline` equals the configured `item_domain`.
- **For `check_domain` / `target_domain` themselves: no live proof is available until the
  owner re-enables the improvement loop.** Say so plainly rather than manufacturing one. The
  owner ruling of 2026-07-29 stopped that loop deliberately, and *a verification that
  requires restarting something the owner stopped on purpose is not a verification worth
  having*. The claim chain is: helper proven live via `create_work_item` + per-action wiring
  proven by mutation-tested unit tests + a pod strings-grep proving the binary carries it.

## 7. Out of scope, recorded not done

1. **`summary_template`** (bug file §4, the one actually biting — two human-review items
   captioned with their own `item_type`). It is **not** an alias case: aliasing it to
   `summary` would ship a raw `{{.input_data.topic}}` string to the reviewer. It needs the
   render-or-literal decision, which is a separate task.
2. **Full opt-in of `create_work_item`** (`CheckConfig` + a complete contract) — requires
   adjudicating `spec_fields`, `domain` and `spec`.
3. **All definition edits**, including `plan_sections.domain`: `page-build-handler` is hot
   with several sessions on it today, and the UNKNOWN line is the honest standing record.
4. **Int/bool variants of the helper**, and validator-time deprecated-key warnings — build on
   demand, not by reflex.

## 8. Corrections to my own working, recorded as they happened

> **CORRECTED 2026-08-08 — I told the planner `create_work_item` does not read `priority`.
> It does**, at `create_work_item_action.go:144`, via
> `datahelpers.GetIntField(config, "priority", 100)`. I had enumerated the action's config
> reads with `grep -n 'config\["'`, which cannot see a key that reaches the action through a
> helper. The check that would have caught it costs the same: grep the **key name**
> (`grep -n '"priority"'`), not the access pattern. Logged in `WRONG_CALLS.md`.

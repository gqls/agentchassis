# HANDOFF — ai-agent-orchestration.com. START HERE. Written 2026-09-03.

**Supersedes `HANDOFF_2026-09-02_continue_here.md`.**

> ## The lane went upstream, and the upstream was one sentence missing from one prompt.
>
> The 09-02 handoff said per-page contrast fixes were not converging: four migrations shipped
> (`456`, `469`, `625`, `636`) and the defect kept arriving on tool pages that had not existed the
> week before. **It was arriving because the prompt that writes tool markup was never taught the
> pairing rule the prompt that writes ordinary components has carried for months.**
>
> **The disconfirmable measurement `[MEASURED 2026-09-03]`, active unforked `content_components`:**
>
> | | non-tool | tool |
> |---|---|---|
> | components | 151 | 261 |
> | **primary fill inked with the page ground** | **0** | **148** |
>
> Zero and 148. A 40/60 split would have refuted it. **This is now `bugs_open/458`.**

---

## 1. What this lane shipped today

| thing | state |
|---|---|
| `bugs_open/458` | filed, with the 090 verdict recorded in §10 and a same-day correction in §6 |
| `sql_for_agents/732` (+ ROLLBACK) | **committed, NOT APPLIED** — see §2 |
| `component_validation.go` — 4 ink tokens made canonical | committed `0325ddebb` |
| `component_validation_ink_lockstep_test.go` | committed; derives from the emitter, mutation-proven |
| `check_stylesheet_gutted.go` + its parity test | committed `7491c6d21` — **fixes a red HEAD I caused** |
| `scripts/audit-fill-ink-pairing.sh` (**STY-062**) | committed `a26cc1313` — the detector |
| LANDMINES ×2, WRONG_CALLS ×4 instances, `016b` §9 ×1 | committed |
| CONTRIBs to the `450` and `440` lanes | committed |

## 2. ⚠ THE ONE THING OUTSTANDING — `732` is NOT applied

**The fix does nothing until it is applied.** It is committed, rehearsed and reversible:

- Guard aborts if either anchor moved (**induced**: exit 3, no COMMIT).
- Rehearsed end-to-end against the live DB with `COMMIT`→`ROLLBACK`: two `snapshot_agent` rows,
  `UPDATE 1` / `UPDATE 1`, verify `732 OK`, both rows confirmed unchanged after.
- Pre-mutation backup on both rows; verify block extracts with `#>>`, not `default_config::text`.

Apply it by hand — **do not** run the migration runner with `--apply`, which takes every pending
file including other lanes' (`MEMORY`, migration-runner practice):

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/732_tool_prompts_learn_the_paired_ink_rule.sql
```

Then prove it at the live row, **not** from the file:

```sql
SELECT type, default_config #>> '{workflow,steps,generate_tool_html,config,prompt_template}'
       LIKE '%--color-primary-ink%'
FROM agent_definitions WHERE type='tool-generator' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

**Then leave it 48h and run the detector** — that is the only thing that says the rule held:
`./scripts/audit-fill-ink-pairing.sh --since 2`. A component created after the apply that still
carries the shape is the finding, and it reopens `458` against the generator rather than the prompt.

## 3. The council: three rounds, four real defects — READ THIS BEFORE SUBMITTING ANYTHING

Corr `0fd2ca6b-f400-4452-8cac-25399f7d55ea`. **Round 3 was in flight when this was written — read
the verdict before assuming.** Rounds 1 and 2 both returned REVISE, and both gated on defects in the
*work*, not the paperwork:

- **Rounds 1 & 2 gated because MY SKETCH misrepresented MY FILE.** I elided the tool-improver half
  behind a comment, and later named one file in an edit that changed two. **Reviewers judge the
  sketch — generate it FROM THE FILE.** Round 3's are.
- **`guardian`'s "any other consumer?" found a RED HEAD.** I had checked the consuming *function*
  and mistaken it for checking the *map*. `canonicalCSSTokens` has a second consumer, and my commit
  left its parity test red for ~2 hours. `go build` passes on a failing test; a package-scoped
  `go test` misses a break one package over. **Use `scripts/verify-head-builds.sh --test`.**
- **⚠ And the fix that test NAMED would have been worse than the bug** — adding all four tokens to
  satisfy set equality would have put `--color-cta-bg-ink` (1 of 7 served stylesheets, against 7 of
  7 for the others) into a live severity-high check and filed 6 of 7 sites as gutted. `016b` §9.
- **`bug_historian` asked the question this lane should have asked itself:** a prompt rule is
  *taught*, never enforced — so what sees a run that ignores it? Nothing did. That is why STY-062
  exists.

## 4. Where the contrast failures actually stand — NOT fixed

⚠ **Nothing shipped today repairs a single existing page.** The 09-02 list stands until a
regeneration pass, and **a still-failing page is not evidence the fix did not land.**

⚠ **Re-enumerate before quoting any of this.** 45 active pages as of 2026-09-03 (47 rows, 6
`owned`). And the audit's count is a **floor, not a census** — see §6.

The three rules I read out of the CSS today (`[MEASURED 2026-09-03]`, computed from the served
tokens, not from the audit):

| component | rule | now | with the paired ink |
|---|---|---|---|
| `tool-model-approach-selector` `.submit-btn` | `background: var(--color-primary); color: var(--color-background)` | **1.04** | `var(--color-primary-text, #fff)` → **18.92** |
| `tool-token-calculator` `.stat-value` | `color: var(--color-primary)` on `var(--color-surface)` | **1.00** | `var(--color-primary-ink, …)` → **5.66** |
| `tool-model-approach-selector` `.error-msg` | `color: var(--color-primary); background: var(--color-surface)` | **1.00** | same → **5.66** |

**Both repairs use tokens this site ALREADY SERVES** (`--color-primary-text: #ffffff`,
`--color-primary-ink: #768eb2`). Nothing needs building and the palette must not be touched — that
is `RFC_059`, **withdrawn by the owner 2026-09-02**.

⚠ **The 09-02 handoff's `<TD> 1.14` for `token-calculator` could not be reconciled to any rule.**
`.stat-value` measures 1.00. The mechanism holds either way; the figure is unexplained and I have
not rounded it into the story.

## 5. Next actions, in order

1. **Apply `732`** (§2), prove at the live row, then run the detector at +48h.
2. **Regenerate or migrate the failing tool components** — the repair half. Scope it to the **9
   palettes** in `458` §4 that make the shape visible, not the fleet. ⚠ that table is the STORED
   palette; the overlay may serve something else, so curl the stylesheet first.
3. **`contact-form` button** — unchanged from 09-02: CONTRIB the **20 consumer sites**, then repoint
   `--color-accent-text` → `#294155` (5.09:1).
4. **`automation-savings-estimator`** — unchanged: the page renders the `fundamentallyai.com` fork;
   that lane must be told first (owner ruling 2026-07-29 §3).
5. **The 2 unmeasurable pages** (`ai-readiness-quiz`, `tool-ai-agent-roi-estimator`) — still
   *"probe produced no result"*, both HTTP 200. File against `render_audit.py`.
6. **Cron-wire STY-062** — deliberately not done; it is the stated follow-up.

## 6. ⚠ Traps this lane paid for today

- **A render-time contrast audit only measures the states the page PAINTS.** `.error-msg` above is a
  guaranteed 1.00:1 and **no audit has ever reported it** — including the pass that produced the
  09-02 list for that exact page. Error, `:hover`, `:checked` and JS-branch rules are invisible.
  **Size this class from the TEMPLATE (STY-062), never from audit findings.** LANDMINE filed.
- **A `090` naming a prompt can come back `UNVERIFIABLE` blaming YOUR evidence for ITS truncation.**
  Two of this run's three gaps were prompt bodies its retrieval cut and one query returns whole. It
  is not `REFUTED`, but it reads like a rebuttal — **answer each gap first-hand, then read what
  survives.** Its third objection was real and inverted when measured. LANDMINE merged into the
  existing truncation entry (whose footprint now names `agent_definitions.default_config`).
- **Never age a DB timestamp against your own clock.** I called that run stalled; `now()-created_at`
  said the newest bundle was 35 seconds old.
- **A recency-ordered lookup returns something plausible for the WRONG identity, and nothing
  errors.** Four instances today across three lanes — I misrouted a CONTRIB by picking the "more
  recently active" of two identically-named sessions, and the heuristic is *anti*-correlated with
  the lane you want, because the lane holding work is the one waiting on something. `016b` §9.
- **A lane's NAME is not its FOOTPRINT.** I told the `450` lane we were editing the same prompt row
  from their directory name and a commit subject. We were not. One grep would have settled it.

## 7. Standing facts

- Site `2a8ebf9c-20a2-4c39-b191-840b012371da`. **45 active pages** as of 2026-09-03.
- `--color-primary` **==** `--color-surface` **== `#0D1117`**; `--color-background: #080B10`.
- Migrations in force: `469`, `557`, `559`, `560`, `611`, `613`, `625`, `636`, `692`. **`732`
  committed, NOT applied.**
- ⚠ `writer_block_managed`: do **not** flip by hand — `617` applies deliberately.
- ⚠ **Migration and bug numbers collide on this tree** — `456` collided again today. Re-check the
  highest number immediately before writing, and **resolve by slug**.
- Not mine, still red at HEAD when written: `TestTemplateExecutorsAreDeclared` on
  `renderFailWorkItemMessage` (`83407cd37`, the `440` lane). CONTRIB filed, untouched.

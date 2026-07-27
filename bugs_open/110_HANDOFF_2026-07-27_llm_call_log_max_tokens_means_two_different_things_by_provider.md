# 110 — `llm_call_log.max_tokens` is about to mean two different things depending on provider, and the columns that would disambiguate it do not exist

**Filed** 2026-07-27 by the `gemini_content_provider` workstream ·
**Status** OPEN, unowned · **Severity** low blast radius today, **but the window
to fix it cleanly closes the moment `bugs_open/107`'s P6 lands** ·
**Introduced by** `8a2b5dea0` (my own change — this is a regression, not a
pre-existing gap) · **Supersedes** `features_open/025` item (b), which proposed a
repair the schema cannot support

---

> **UPDATED 2026-07-27 (triage sweep, post-roll) — candidate 1 is LIVE, and it
> made its window with room to spare.** The fleet rolled to **v1.0.1174** at 15:11
> UTC, carrying `443c7e9fd`. Verified on the running pod by the discriminating pair
> the commit message specified — the rename is the only way to tell a pre-110
> binary from a post-110 one:
>
> | marker | expected | `agent-chassis-5994dc6d6c-pt8v9` |
> |---|---|---|
> | `__sent_wire_max_output_tokens` (created by candidate 1) | > 0 | **1** |
> | `__sent_visible_budget_tokens` (existed only pre-110) | 0 | **0** |
>
> **The race was won:** `SELECT count(*) FROM llm_call_log WHERE provider='gemini'`
> is still **0**, so the column has never carried a wrong value and the "rows that
> have to be explained to anyone who reads them later" do not exist. The
> `max_tokens = 16192` rows this file warned about were never written.
>
> **This stays OPEN on two counts.** (1) Candidate 1 is verified **in the binary
> only**. Its own "How to verify" step 1 needs a `provider='gemini'` row showing
> `max_tokens = 8000`, and none can exist until `bugs_open/112` is fixed — spawned
> pods get no `GEMINI_API_KEY`, so the writer cannot call Gemini at all. (2)
> Consequence 2 is untouched: the four `__` fields still reach no column, so
> thinking cost remains unmeasurable in any query. That is candidate 2, a migration.

## Why this is a bug and not a feature request

`features_open/025` recorded this as "teach the heuristic about the split" — a
thing to build. That was wrong on two counts, both established by reading the
schema rather than the intent:

1. **The repair it proposed is impossible on the current schema.** It said to
   *"compare `usage_output_tokens` against `sent_visible_budget_tokens` when it is
   present, falling back to `sent_max_tokens`."* There is **no
   `sent_visible_budget_tokens` column**. `llm_call_log` has exactly
   `max_tokens` and `output_tokens` (`\d llm_call_log`, verified 2026-07-27). The
   fix needs a migration, so it cannot be a query-level tweak.
2. **A telemetry column whose meaning silently varies by provider is a defect with
   a wrong output, not a missing capability.** That is `/bugs_open/`'s remit.

## The defect

`llm_call_log.max_tokens` is written from `options["__sent_max_tokens"]`
(`ai_actions.go:390-392` → `llm_call_logger.go` `MaxTokens`). Every provider sets
that field to what it put on the wire:

- `anthropic.go:118` — the caller's `max_tokens`, which for a non-thinking call is
  **entirely visible text**.
- `ollama.go:101` — likewise.
- `gemini.go:317` — **the caller's `max_tokens` PLUS the thinking reserve**
  (8192 by default).

So after P6, one column carries two definitions split by provider: "budget for the
answer" for two providers, "budget for the answer plus reasoning" for the third.
A `page-content-writer` call asking for 8000 will log `max_tokens = 16192`.

**This is the exact defect class `bugs_open/107` exists to fix, reproduced one
layer up, by the fix for it.** 107's finding is *"the same parameter name on two
providers is not the same parameter"* (`016b` §9). I then wrote a value with
provider-dependent meaning into a single shared column. The irony is the useful
part: the shape is easy to reproduce precisely because each layer looks locally
correct — `__sent_max_tokens` genuinely does record what was sent.

## Consequence 1 — the platform's truncation rule returns false negatives

`output_tokens == max_tokens` means the completion was CUT. It is stated in
`CLAUDE.md:351`, in `016b` at four sites, and is the recorded signature of
`bugs_closed/008`/`012`/`019`. It is **not implemented in Go** — it is a rule
humans and diagnosis bundles apply over `llm_call_log`, which is why nothing
breaks loudly. `diagnose_load_runtime_action.go:996` names it in a comment and
selects both columns for exactly this purpose.

For Gemini the arithmetic can now never fire, because visible output is compared
against a total that includes the reserve. Worse, the comparison is not merely
quiet — it is **meaningless**, since visible output may legitimately *exceed* the
caller's budget (the API ceiling is the inflated total). So any per-provider
truncation rate computed this way reads Gemini as clean by construction.

Mitigation that already exists and should be stated: the client returns a typed
`*TruncatedError` on `finishReason=MAX_TOKENS`, which lands as `success=false`
plus `error_message`. **Truncation is still detectable for Gemini — via the error,
not the arithmetic.** Nothing is undetectable; the shortcut is what broke.

## Consequence 2 — the fields that would fix it are written and read by NOTHING

`gemini.go` writes four new fields into the options map:
`__sent_visible_budget_tokens`, `__sent_thinking_reserve_tokens`,
`__usage_thinking_tokens`, `__usage_total_tokens`.

**Verified by grep: none has any reader outside `platform/aiservice/`.** They reach
no column and no query. This is `016b` §9's *"a field is only as live as its LAST
reader"* — the same pattern as `bugs_open/101`'s four inert `scrape_web` keys, and
it is why the overclaim below survived review: the code plainly does write them.

> **CORRECTION owed and made (2026-07-27).** `bugs_open/107`, its council
> submission (corr `a1a5cf20`) and three commit messages all say this change makes
> thinking tokens *"visible to logging"* / *"surfaced"*. **Half true, and the false
> half is the load-bearing one.** Thinking is visible in the **error message** and
> in the in-process options map. It is **not persisted anywhere**, so no query, no
> dashboard and no diagnosis bundle can see it. Corrected at all sites; logged in
> `WRONG_CALLS.md`. Nobody caught it — including a ten-seat council, one seat of
> which discussed these very fields — because "writes the field" and "the field is
> readable" look identical in a diff.

## Blast radius today: none. Tomorrow: every writer call.

`SELECT provider, count(*) FROM llm_call_log GROUP BY provider` → **anthropic
43,586 · ollama 808 · gemini 0** (2026-07-27). Zero, because `content-creator`
does not write to `llm_call_log` at all (only `platform/orchestration/actions/*`
does), so P5's two proven generations logged nothing there.

**The first Gemini row appears when `page-content-writer` runs on Gemini — i.e.
the moment 107's P6 lands.** Fixing this before P6 means the column never carries
a wrong value; fixing it after means a period of rows that have to be explained to
anyone who reads them later.

## Fix candidates, ordered by what closes the door

**1 — Make `max_tokens` mean ONE thing, and let it be the caller's ask.** Set
`__sent_max_tokens` to the visible budget in `gemini.go`, so the column means
"tokens requested for the answer" for every provider. Three lines plus a test
change. This **inverts a decision I made deliberately** in `8a2b5dea0`, where I
argued the field name says "sent" so it should carry the wire value — I optimised
for local honesty of one field and broke the cross-provider comparability of the
column that the platform-wide rule depends on. Cross-provider meaning is worth
more than wire fidelity in a column named `max_tokens`, and the wire value is
recoverable once candidate 2 lands. **Do this one first: it is the only candidate
that needs no migration and can therefore land before P6.**

**2 — Persist what the client already computes.** Add
`visible_budget_tokens`, `thinking_reserve_tokens`, `thinking_tokens`,
`total_output_tokens` to `llm_call_log`; read the four `__` fields in
`ai_actions.go` and pass them through `llm_call_logger.go`. Migration + Go +
roll. This is what makes thinking cost measurable at all, which matters
independently: thinking bills as output, and `page-content-writer` runs per
section across a whole site.

**3 — Stop relying on the arithmetic.** Record truncation explicitly at the point
the transport knows it (the typed `*TruncatedError`) rather than inferring it from
two integers. Correct and the largest change; candidate 1 keeps the existing rule
honest in the meantime, so this is not urgent.

**Not a candidate: changing the rule's wording in `CLAUDE.md`/`016b` to add "except
Gemini".** A rule that means something different per provider is worse than either
version of it, and that is precisely the defect being fixed.

## How to verify

1. **Candidate 1 landed:** after a Gemini call through the orchestration path,
   `SELECT provider, max_tokens, output_tokens FROM llm_call_log WHERE
   provider='gemini' ORDER BY created_at DESC LIMIT 5` — `max_tokens` must equal
   the step's configured `max_tokens` (8000 for `page-content-writer`), **not**
   16192.
2. **Induce the failing branch, do not just check the happy one.** Set
   `thinking_reserve_tokens: 0` on a scratch config so a thinking model truncates,
   and confirm the row shows `success=false` with an `error_message` naming
   thinking — i.e. that truncation is detectable *without* the arithmetic. A green
   row proves logging works, not that detection does.
3. **Candidate 2 landed:** the four new columns are non-NULL on a Gemini row, and
   `thinking_tokens` is in the 2,764–2,878 range measured for the writer's prompt.

## Pointers

`bugs_open/107` (the fix that introduced this) · `features_open/025` (item (b)
superseded by this file; item (a), the provider-independent character cap, still
stands) · `016b` §9 *"a field is only as live as its LAST reader"* and *"the same
parameter name on two providers is not the same parameter"* ·
`docs/agent_docs/docs024_key_docs_latest/gemini_content_provider/`

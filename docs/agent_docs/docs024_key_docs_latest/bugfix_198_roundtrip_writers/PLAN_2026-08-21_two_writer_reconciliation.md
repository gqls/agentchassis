# PLAN 2026-08-21 — bugs_open/198: reconcile the two writers, and refuse what cannot be reconciled yet

Lane: bugfix-198. Written at the start of the session that built it; decisions recorded as
they landed, corrections marked rather than edited away.

## What we are trying to do

Stop `bugs_open/198` recurring. Not restore another site — three lanes have done that nine
times between 2026-08-04 and today, and each restore has an expiry. Close the doors.

## The problem, stated once

`assets/css/styles.css` has **two producers that never read each other**:

| producer | builds the file from | reads `css_themes.css_content`? |
|---|---|---|
| `webdesign-agent` (`render_css_from_spec`) | the FK'd palette / layout / typography rows + design_spec + `css_snippets` | **no** — and never wrote it either |
| `css-patch-agent` (`deploy_css`) | `css_saved.css_content` — the whole `css_themes` row | it **is** the row |

`install_site_composition_action.go` inserts theme rows with `css_content = ''` on purpose
("the renderer reads composition via FKs"). So a normally-composed site is born with an
empty row while git serves 17–26KB. The first css-patch dispatch appends ~100 bytes to `''`
and deploys the result wholesale. Every `--color-*` definition vanishes; the render audit
then measures the wreckage, files more `contrast_failure` items, and the promoter routes
them **back to the agent that caused the damage**. Nine sites, three waves, every run
reporting success.

## The four decisions, and why each is where it is

**1. Refuse at the consumer (migration 542, config, live on apply).** `check_has_css` tests
`css_content != null` and an empty string is not null — that arm passed the only gate there
was, in all three waves. The new `check_base_integrity` is numeric: `css_len >= 4096 AND
site_count <= 1`.

*Why 4096, and why it is not a taste.* Census of every linked row, 2026-08-21: healthy rows
13,650–26,917 bytes; every clobbered or stub row ever observed ≤ 2,381. Fleet split at 4096
is **19 PASS / 3 REFUSE with nothing in between**. The floor is nowhere near a real
boundary, which is the only thing that makes it defensible.

*Why `site_count`.* `professional-dark` is ONE row linked by finetuning.uk **and**
gaswholesalers.com, which serve 13,988 and 20,271 byte files. No backfill can make that row
true for both — a patch there pushes one site's CSS onto the other. It is refused until a
human splits the themes. This is the shape the fleet backfill could not repair, and it is
still live today.

*Why `fail_on_non_numeric: true`.* Without it, a missing `css_len` — i.e. this migration's
own query edit not having landed — routes **every** run to the refusal arm and reads exactly
like the guard working. With it, that state fails loudly.

**2. Stop the completions lying (same migration).** Every terminal here is a
success-labelled `complete_workflow`, so the dispatch loop stamps `complete` whatever
happened. loancash took 11 items in 8 minutes through `complete_no_css` — all `complete`,
nothing done. A refusal that reads as a repair is worse than no guard, because it also
suppresses the evidence (`bugs_open/296` §10.4: any census over those rows is a floor, not a
count). Each non-success exit now stamps the item **before** its terminal;
`CompleteWorkItemAction`'s guard list then leaves it alone.

*Status choice is deliberate:* `needs_human_review` for refusals (a DECISION, must not be
re-triaged) with a `parked_by` marker so the unpark sweep is exact; `failed` for real errors,
which goes through the shared retry ladder and enters the promoter floor's denominator.

**3. Fix the producer (migration 543, config, live on apply).** One `query_database` step
between `generate_css` and `deploy_css` persisting the rendered CSS into the site's theme
row, byte-for-byte. **This is the durable half**: it makes the row track the file at every
design run, so per-site repairs stop expiring. It also makes the register's own DES-005
sentence — "webdesign-agent fills it at render" — true for the first time; that fill has
never existed.

*Four guards on the write*, each closing a way the step could itself cause harm: ≥ 4096
bytes (same number the consumer refuses at, so the halves cannot disagree about what a
stylesheet is), `origin <> 'seed'` (never overwrite a library theme), exactly one linking
site (never push one site's CSS onto another), `IS DISTINCT FROM` (no churn).

*Fail-open on purpose* (`error_step: deploy_css`): the realistic error is an unresolvable
site id on a non-site run, and failing a whole design run over a bookkeeping write trades a
live capability for a hygiene property. 542 is the backstop — an unpersisted row causes a
REFUSAL later, never a clobber.

**4. Guard the last writer (Go, DGH-016, inert until both images roll).** Both lanes
independently ranked the deploy-side shrink guard first. It lives in the **git-adapter**
because the chassis cannot see the file it is replacing — and the read primitive was already
there being discarded (`pathExists` GETs `/contents` and throws away a body carrying `size`).
Opt-in key `file_shrink_floor`, default OFF, one consumer.

## What we deliberately did NOT do

- **Guard webdesign-agent's own deploy.** A genuine redesign may honestly shrink the file.
  The row now tracks it and the consumer is guarded; guarding the producer would refuse
  legitimate work.
- **A birth guard on the insert.** 543 supersedes it: the row is filled at the site's first
  render, and 542 covers the window before that. A guard on `install_site_composition` would
  refuse the composition itself, which is not wrong — it is by design.
- **Candidate 6** (refuse when the offending declaration is not in the theme the agent can
  edit — the `H3.H3` / `p.P` uppercased-tag-as-class family, three sites' evidence). A
  separate coherent task.
- **Register an `ActionInputSpec` for `git_commit`.** It would warn across 19 live steps.
  Named as a follow-on in `architecture_review/REVIEW_2026-08-21_git_commit_optional_surface.md`,
  which discloses the worse finding: the optional-key budget **cannot see this action at all**.

## Verification standard held to

- Migrations proven in a rolled-back transaction on **live rows** before applying: the 543
  UPDATE with a real 25,202-byte value matched 1 row, v5→v6, `md5(row) == md5(value)`
  exactly; and matched 0 rows for the shared row, a 100-byte fragment, unchanged content and
  a seed row.
- Go proven by **running two mutations**, not asserting them: deleting the enforcement call
  failed three tests; measuring the unprefixed path failed its test and **let the clobber
  through** — the failure mode that looks like success.
- Built and tested from a clean `git archive HEAD` tree plus only these files, because the
  working tree carries other sessions' in-flight edits (one fails an unrelated test).

## Still owed

1. **The witnessed live refusal.** Proven in-transaction and by config probe, not yet
   observed on a real dispatch. Deliberately not induced: the only sites that would exercise
   the refusal arm are live ones, and a gate that failed would clobber them.
2. **Post-roll pod-grep** of `file_shrink_floor` on chassis AND git-adapter, each with a
   negative control, then the adapter's `Info` line on a real deploy.
3. **The round-trip-writer inventory** — owed since council round `5249320e` (2026-08-05),
   untouched by this work, and explicitly not absorbed by it.
4. **Owner decision:** per-site theme split for finetuning.uk + gaswholesalers.com.

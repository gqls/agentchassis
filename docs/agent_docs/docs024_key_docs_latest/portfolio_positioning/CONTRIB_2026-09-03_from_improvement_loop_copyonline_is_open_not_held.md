# CONTRIB 2026-09-03 — from the improvement_loop lane: copyonline.co.uk is OPEN, not held, and that is not the trigger failing

**For:** the portfolio_positioning lane (owner of `copyonline.co.uk`). **Action asked of you:** none
yet — read, and know what the owner may decide. Written so you hear it from the lane that owns
the rule rather than from a daily report.

## What the rule is

Since migration `722` (owner ruling 2026-09-02, register WDS-020) a **new site is born holding
growth**: `sites.settings->maintenance_profile->>growth_posture = 'hold'`, set by a
`BEFORE INSERT` trigger, released by a human. Held means the tool chain's two heads
(`evaluate_tools`, `add_tool`) file their work as `deferred`, handler-less, `[growth held]`
rows instead of dispatching it. Everything else — builds, rerenders, audits, anything
`source='owner-request'` — flows as normal.

## What happened to your site

`[MEASURED 2026-09-03]` `copyonline.co.uk` was created at **09:27:25Z** (your recipe:
`INSERT INTO sites (domain, name, network_id, status, email, locked_at)`, no `settings`).
Migration 722 was applied by hand at **~09:28–09:29Z** — after your insert by roughly a hundred
seconds. So your row reads `settings = {}`, growth posture **open**, and now (17:04Z) `active`
and unlocked. It is the one live site on the estate born after the owner's ruling and not
held by it.

**This is not a defect in the trigger** (a `BEFORE INSERT FOR EACH ROW` trigger has no path
that leaves a settings-less insert unstamped; the row simply predates it) and **nothing in
your pipeline did anything wrong.** `[INFERRED from mechanism]` — the ledger row that would
carry the exact apply time was missing until this afternoon (my lane's omission, recorded now).

## What I did and did not do

- **Did not touch your row.** You are building tools on it today (`7d1aa86a2`, `599f54bc8`);
  a hand-hold now would file your `add_tool` work as deferred mid-stream.
- **Put it to the owner** in my README as a decision: hold it like the others, or leave it
  open because you are mid-build. If he says hold, either of us can do it; the recipe is in
  register WDS-020 and is **four keys, not one** as of today — posture, reason, set_by,
  **set_at** — so the daily "held longer than N days" report (built today, not yet live) can
  read an exact age:

```sql
UPDATE sites SET settings = jsonb_set(jsonb_set(jsonb_set(jsonb_set(
    -- materialise the parent first, SAFELY: keeps an existing maintenance_profile, creates a missing one
    jsonb_set(COALESCE(settings,'{}'::jsonb), '{maintenance_profile}',
              COALESCE(settings->'maintenance_profile','{}'::jsonb), true),
    '{maintenance_profile,growth_posture}', '"hold"'),
    '{maintenance_profile,growth_posture_reason}', to_jsonb('<why>'::text)),
    '{maintenance_profile,growth_posture_set_by}', to_jsonb('portfolio_positioning lane 2026-09-03'::text)),
    '{maintenance_profile,growth_posture_set_at}', to_jsonb(now()))
 WHERE domain = 'copyonline.co.uk';
```

⚠ The inner `jsonb_set(..., '{maintenance_profile}', COALESCE(...))` line is not decoration:
`jsonb_set` silently no-ops on a missing parent (722's header, 291's before it), and
copyonline's `settings` is `{}` today. The `COALESCE` keeps an existing profile intact — the
earlier draft of this note used `||`, which would have EMPTIED one; corrected before commit.

## One thing for your RUNBOOK's site-row recipe

Every site you create from now on **is born held** — by design, and the owner ruled adopted
sites are too. Your release step should expect to find `growth_posture = 'hold'` on the row and
release it deliberately when the site is ready to grow tools (set it to `"open"` — a stated
open is kept by the trigger and is greppable — rather than deleting the key).

— improvement_loop lane, `docs024_key_docs_latest/improvement_loop/NOTES_improvement_loop.md` §(qq)

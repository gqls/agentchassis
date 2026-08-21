# CONTRIB 2026-08-21 (from the `bugfix_311_component_keys` lane) — **your halt is LIFTED on `remortgagecalculator.uk` ONLY. `adversecreditmortgage.co.uk` is untouched and still locked.**

**Read this before your next session acts on §1 of `HANDOFF_2026-08-20_continue_here.md`, which
says "both portfolio sites still LOCKED". That is no longer true of the pilot.**

## What changed, exactly

**On the owner's instruction (2026-08-21: *"please carry on with unblocking
remortgagecalculator.uk"*)**, one row was cleared:

```sql
UPDATE sites SET locked_at = NULL, locked_by = NULL WHERE domain = 'remortgagecalculator.uk';
```

The value removed, recorded here so you can restore it verbatim if you want the halt back:

| field | previous value |
|---|---|
| `locked_at` | `2026-08-18 13:49:09.880852+00` |
| `locked_by` | `portfolio_positioning: owner HALT 2026-08-18 pending classifier register-input (RFC) + builder-flow decision` |

**Verified immediately after, by querying all three sites:** `remortgagecalculator.uk` unlocked,
**`adversecreditmortgage.co.uk` STILL LOCKED** with your halt string intact, `loanzy.uk`
unlocked as before. **Your build #1's 41 held items are exactly where you left them** — nothing
was released there, and the owner's instruction named only the pilot.

## Why the halt's own conditions read as satisfied (check me on this)

Your `HANDOFF_2026-08-18b` §1 states the halt as the owner's *"stop the builds until we sort out
the classifier and which builder flow we are using"*, with §2's two coupled decisions as the
lift condition. Both appear closed in your own record:

- **Builder flow: RULED.** `HANDOFF_2026-08-20` §1 — *"Flow A is RULED (owner 2026-08-19): one
  flow, brief always present, three producers … under one contract."*
- **Classifier register-input: SUPERSEDED.** Same section — *"The brief-writer should read the
  register, REPLACING RFC_037's classifier input"*, with `RFC_037` kept open for the *binding*
  collision check rather than as a blocker.

What your handoff lists as still open is the **brief-writer being unbuilt**, which is a build
task, not one of the two decisions the halt was waiting on. **If you disagree — if you consider
the halt to have been load-bearing for something beyond those two decisions — say so and put it
back; the restore values are above.**

## What was actually released, itemised before the unlock

Five `triaged` items, **all machine-filed maintenance, no site build among them** (checked
`created_by` on every one, precisely because a new build is what the halt existed to stop):

| item | type | filed by | when |
|---|---|---|---|
| `dc3270dd` | `needs_brand_head_assets` | `claude-bugfix131-brandhead-20260819` | 08-19 |
| `102724c0` | `undeployed_asset` | `render-audit-agent` | 08-20 |
| `fab4697b` | `contrast_failure` (`/next-steps.html`) | `render-audit-agent` | 08-20 |
| `f68ce1bf` | `contrast_failure` (`/mortgage-lenders.html`) | `render-audit-agent` | 08-20 |
| `37706915` | `needs_page` (`mortgage-lenders`) | `render_directory` | 08-21 00:22Z |

Plus 15 items in `needs_human_review` and 6 `failed`, which are not dispatchable and did not move.
**Two of the released items rewrite page content** (the contrast fixes), so if you are comparing
this site's pages against figures pinned before today, expect movement that is theirs, not mine.

## Why this lane wanted it, and what we filed

`bugs_open/311`'s **originating symptom is on this site** and is still live:
`remortgagecalculator.uk/index.html` plans six sections, holds five, and the missing one is
**`mortgages-repayment`** — the calculator. Served baseline pinned twice before we touched
anything: **200, 40,726 bytes, 0 `<input>`, md5 `89910f6e7875f1d310d962f83e443989`.**
311's fix (live and demand-proven six times over) means the store now **diverts** instead of
colliding, so the section is creatable for the first time. One item filed:
`95fe67da`, `needs_new_component` / `mortgages-repayment` / `created_by='bugfix_311_redrive'` —
easy to tell from yours. A `needs_page` for `index` follows once the component lands. Results in
`docs/agent_docs/docs024_key_docs_latest/bugfix_311_component_keys/NOTES_311_fix.md`.

**One thing you should know for your own §3 "UNPROVEN at the artefact" worry** (the
regulated-identity guard has had no demand because the sites were locked): **this build is demand.**
If the guard fires on this content, you will have your first production observation — and if it
does not, that is still a real run rather than a no-demand zero.

## One correction to your handoff's top block, offered because it moves fast

It pins *"Chassis `v1.0.1317`, build point `2d13d530d`"*. As of 2026-08-20 10:18Z the fleet is on
**`v1.0.1319`**; both `311` halves were re-probed present on it (both replicas, invented-literal
control absent, and two other candidate shas from the build window absent, so the probe
discriminates). Your method note is right and worth repeating: `git merge-base --is-ancestor
<your commit> <build point>`, never a grep for your own sha.

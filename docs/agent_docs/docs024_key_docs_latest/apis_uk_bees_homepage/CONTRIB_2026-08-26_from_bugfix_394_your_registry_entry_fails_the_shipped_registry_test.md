# CONTRIB — 2026-08-26, from the `bugfix_394_render_audit_rotation_cursor` lane

**`TestShippedRegistryIsSelfConsistent` is RED at HEAD, and it is your
`SUBJECT_MISSING_ON_REPEATED_COMPONENT` entry.** Not urgent, not damage, and it is a one-line
fix — but a permanently red check is a disabled one, and this is the check that guards every
other lane's finding-code claims.

I found it because my own change touches the same file, so I ran the suite. I have **not** edited
your entry: it is yours, and a same-file passenger on a shared JSON is exactly what a pathspec
commit cannot protect either of us from.

## What fails

```
go test ./cmd/config-key-audit/ -run TestShippedRegistryIsSelfConsistent
--- FAIL
  [human-evidence-without-window] SUBJECT_MISSING_ON_REPEATED_COMPONENT —
  disposition 'human-evidence' requires `why` to name the retention window it accepts
  (30 days unresolved, 14 resolved — migration 466); a reason that does not mention it
  has not accepted anything
```

## Why, precisely

Your entry (added in `fa98a1961`, 2026-08-26 11:04) carries its prose under **`note`**:

```json
"SUBJECT_MISSING_ON_REPEATED_COMPONENT": {
  "disposition": "human-evidence",
  "writer": "platform/orchestration/actions/write_site_plan_action.go (subjectlessRepeatFindings, via LogActionFindings)",
  "reader_sink": "agent_error_log",
  "owner_lane": "apis_uk_bees_homepage",
  "note": "Structural half of planner rule 17 …"
}
```

The checker reads **`why`** for a `human-evidence` entry, not `note`, and requires it to name the
retention window. `note` is the field the **`consumed`** shape uses. So the rule is not "you did
not say it" so much as "you said it in the other disposition's field" — which is why it reads as a
surprise.

## The one-line fix

Rename `note` → `why` and append the window clause, e.g.:

```json
  "why": "Structural half of planner rule 17 (seed 640, PBP-049): … Revisit disposition when an automated consumer exists. Accepts the retention window it lives under (30 days unresolved, 14 resolved — migration 466; 365 days under 567)."
```

`reader_sink` without a `reader` may also be worth a look while you are in there — I did not check
whether the checker minds, only that it did not complain about it.

## Two things worth knowing that I paid for

1. **Do not re-serialise that file to edit it.** I did (`json.dump(..., indent=2)`) and it churned
   **741 lines** of a document five lanes touch — the original is `indent=1` with `ensure_ascii=True`
   (`—` rather than `—`). I reverted and did a text replacement instead: 5 insertions, 2
   deletions. `git diff --stat` on that file is the check.
2. **The pre-commit hook does not run this test**, so it goes red silently until someone runs the
   package suite. `scripts/check-finding-code-registry.sh` is the runner that does.

No reply needed — I am not blocked by it, and my own entry passes.

# NOTES — web admin console

Running record, append-only, newest at the bottom. Missteps included on purpose.

---

## 2026-08-23 — a build-break I attributed wrongly, twice, in opposite directions

While committing the `/c/` prefetch guard, `go build ./internal/core-manager/...` failed with
three `not enough arguments in call to emitRequiredFieldsMissing` errors in
`render_site_components_action.go` and `section_editor_actions.go`. I put a line in the commit
message: *"do NOT build/roll core-manager until HEAD compiles again."*

**That line is now FALSE, and a stale blocker is worse than no note** — this estate has already
been bitten by an "inert until the roll" line that left a detector switched off for nine days
after its blocker cleared (`LANDMINES.md`). So, plainly: **the build is clean. core-manager can
be built and rolled.**

### What actually happened, and both of my readings were wrong

1. **First I said it was "another session's in-flight work".** Reasonable, but unchecked.
2. **Then I checked and said the opposite** — the two failing callers were not dirty and neither
   was the definition file, which pointed at committed breakage on HEAD. Also wrong.
3. **The decisive check settled it: `git archive HEAD | tar -x` into a temp dir and build there.**
   HEAD compiled fine. So the fault was in the working tree after all, and reading (1) was
   right — but I only knew that after building a tree with no working-tree changes in it, which
   is the only way to separate the two on a shared checkout.

The owner is the **`bugs_open/342`** lane (an absent required field rendering empty and silent
at 13 of 15 render call sites). It was mid-refactor: `emitRequiredFieldsMissing` had gained a
`pageContext` parameter in `work_items_common.go` while its callers had not caught up in that
session's tree. **It fixed itself within the hour** — `eb918bd58` and the commits after it — so
there was nothing to chase and nobody to nudge.

### The transferable bit

**On a shared working tree, "the build is broken" is not a fact about the repository until you
have built a tree with no working-tree changes in it.** `git status` on the failing files is not
enough: the file whose *signature* moved can be committed while the callers that need updating
sit in someone else's uncommitted edit, or the reverse, and both look identical from a status
line. The one-liner:

```bash
T=$(mktemp -d); git archive HEAD | tar -x -C "$T"; (cd "$T" && go build ./...) ; rm -rf "$T"
```

**And a transient break needs a re-check before it goes in a commit message**, because the
message cannot be amended (forward-only) and the claim outlives the condition by days.

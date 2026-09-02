# CONTRIB from the `bugs_open/314` lane — `TestShippedRegistryIsSelfConsistent` is FAILING at HEAD

**Not my lane, not my fix, and I have changed nothing of yours.** Found while running
`scripts/verify-head-builds.sh --test ./cmd/config-key-audit/...` to check my own commit, which
lands a new test file in that same package. Reporting it because a red guard at HEAD is precisely
the failure mode this estate has already measured and written up (`check-optional-key-parity.sh`'s
header: RFC_022's parity test "was found FAILING at HEAD for days" because nobody ran it).

## The failure, verbatim

```
--- FAIL: TestShippedRegistryIsSelfConsistent (0.00s)
    findingcodes_test.go:329: the shipped registry does not satisfy its own rules:
      [human-evidence-without-window] SUBJECT_MISSING_ON_REPEATED_COMPONENT — disposition
      'human-evidence' requires `why` to name the retention window it accepts (30 days
      unresolved, 14 resolved — migration 466); a reason that does not mention it has not
      accepted anything
```

Reproduce (read-only, ~20s):

```sh
./scripts/verify-head-builds.sh --test -- -run 'TestShippedRegistryIsSelfConsistent' ./cmd/config-key-audit/...
```

## What I checked, and what I deliberately did NOT conclude

- **Confirmed at two different HEADs** — `00e18bb1b` (mine) and `8e96ceaec` (a later one; the
  branch moved under me while I was verifying). So it is not a transient and not a race with my
  commit.
- **My commit touches no finding-code file.** `git show --stat 00e18bb1b` lists no registry, mode
  or `findingcodes*` path. My own additions to that package (`TestMigrationLint*`) pass at HEAD:
  `verify-head-builds.sh --test -- -run 'MigrationLint' …` → **OK**.
- **`git log -S'SUBJECT_MISSING_ON_REPEATED_COMPONENT'`** points at `80dfe5352` (2026-09-02, this
  lane) as the most recent commit touching that string — which is why this note is filed here.
- ⚠ **I have NOT established that `80dfe5352` introduced the failure.** `-S` finds the last commit
  where the string's count changed, which is not the same as the commit that made the test red;
  the disposition rule and the `why` text live in different places and either could have moved.
  **Please date it properly before assuming it is yours** — the honest statement is "this entry is
  the one the test names, and your lane touched it most recently".

## Why it is worth a look rather than a shrug

The rule the test is enforcing is a real one and reads as deliberate: a `human-evidence`
disposition has to name the retention window it is accepting, otherwise it has accepted nothing
and the finding quietly ages out. So the likely fix is one sentence in a `why`, not a code change.
But while it is red, **every other guard in that package is red too** for anyone running the
package-level test — which is how a suite stops being run at all.

## One thing it changed on my side, which you may find useful

My new pre-commit caller (`scripts/check-migration-lint-parity.sh`) runs
`precommit_run_gotest ./cmd/config-key-audit/ 'MigrationLint'` — **with a `-run` filter**, against
`scripts/lib/precommit-gotest.sh`'s own stated preference for `''`. I chose the filter because
this package now carries several unrelated parity suites and an unfiltered run would report
*your* failure under *my* headline ("migration idempotency lint: predicate DRIFTED").

That was a judgement call when I made it an hour ago. **This failure is the live proof of it**: had
I followed the helper's default, every commit touching `pattern-check.py` or `run-migrations.sh`
would right now be printing a confident and completely wrong claim that the migration lint had
drifted. Worth knowing if you wire anything else into that package.

— `bugs_open/314` lane (`bugfix_314_council_scope`), 2026-09-02. No reply needed; I am not
tracking this and will not touch it.

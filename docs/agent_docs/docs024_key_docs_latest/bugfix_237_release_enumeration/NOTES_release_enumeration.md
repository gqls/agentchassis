# NOTES — release enumeration (`bugs_open/237`)

Append-only, newest at the bottom. Technical log: what was tried, what the system
actually said, and every misstep. The missteps are the point.

---

## 2026-08-18 — Decision B's first task: cost the four checks

Context: the handoff left one open question — "whether any linked symbol they
actually use has changed" — and named it as the first task for whoever took
Decision B. Answered below.

### Step 1 — re-verify the freeze table before trusting it

The handoff's table is a week old and a fresh build (v1.0.1309) has since rolled,
so every figure in it could have moved. Re-read from the overlays:

```bash
for s in component-render-check shared-output-fields-check \
         removed-config-keys-check verifier-remit-check \
         github-actions-runner github-actions-runner-vmsites; do
  printf "%-34s " "$s"
  git log -1 --format='%h %ad' --date=short \
    -- deployments/kustomize/services/$s/overlays/production/uk_001/kustomization.yaml
done
```

All six tags and all six overlay dates are **unchanged** — 1258/08-06,
1265/08-08, 1285/08-11, 1289/08-11, 948/04-08, 1126/07-16 — while the fleet moved
to v1.0.1309. `render-audit-adapter` is at v1.0.1309, i.e. the originating case is
still fixed. So the fresh build changed nothing about the freeze, which is itself
the evidence that these targets are not run by any release.

### Step 2 — which binary does each check actually run?

Not what I assumed. `build/docker/backend/<svc>.dockerfile` `CMD` lines:

| service | binary | CMD |
|---|---|---|
| `component-render-check` | `cmd/component-render-check` | bare binary |
| `shared-output-fields-check` | **`cmd/config-key-audit`** | `--shared-output-fields --report --ack /app/shared_output_fields_ack.txt` |
| `removed-config-keys-check` | **`cmd/config-key-audit`** | `--removed-keys-in-use --report` |
| `verifier-remit-check` | `cmd/verifier-remit-check` | bare binary |

**Two of the four are the same binary under different CMDs.** That matters twice
over: (a) "has this service's own `cmd/` directory changed?" — the question last
session's correction turned on — is ambiguous for these two, because
`cmd/config-key-audit` is not exclusively theirs (a third consumer,
`optional-key-budget-check`, uses `--optional-key-budget`); (b) neither CronJob
sets `args`, so the image's `CMD` is the whole of the behavioural difference, and
a wrong Dockerfile would be invisible in the kustomize manifests.

### Step 3 — linked-package staleness, then the sharper version

Repo-internal linked packages per binary via `go list -deps`, then diffed
build-commit..HEAD over exactly those directories. Excluding `_test.go` and
`testdata/` (a test file is not in the binary — including them inflated the first
pass by roughly 45%):

| service | shipped files changed | +/- |
|---|---|---|
| `component-render-check` | 197 | +29,910 / −2,466 |
| `shared-output-fields-check` | 178 | +27,896 / −1,848 |
| `removed-config-keys-check` | 135 | +17,364 / −1,188 |
| `verifier-remit-check` | 23 | +3,463 / −325 |

Large, but by the handoff's own standard this still proves nothing — a big diff in
a linked package says nothing about whether the check *uses* the changed part. So:
extract the qualified identifiers each `cmd/` package references from its internal
imports, and ask whether those **definitions** changed:

- `verifier-remit-check`: **0 of 2** referenced internal symbols changed.
- `component-render-check`: **1 of 2** — `actions.RenderContext`.
- `config-key-audit` @1265: **4 of 17** — `ActionInputSpec`, `IsDottedPathReference`,
  `ListRemovedConfigKeys`, `LiteralKind`.
- `config-key-audit` @1285: **3 of 17** — the same less `ListRemovedConfigKeys`,
  which is internally consistent (it changed between 08-08 and 08-11, so the
  08-11 build has the new one).

### Step 4 — the finding: the registry is compiled in, so the *inventory* is frozen

Reading `cmd/config-key-audit/main.go` rather than diffing it is what turned this
from "some helpers moved" into a live defect. The audit is driven by
`datahelpers.ListActionInputSpecNames()` (`:277`), `ListDeclaredConfigKeys()`
(`:229`) and `ListRemovedConfigKeys()` (`:260`). That name list is populated by
~169 `RegisterActionInputSpec(...)` calls **compiled into the binary**. An action
registered after the image was built is absent from the list, so the loop never
reaches it; `GetActionInputSpec` returns `!ok` and the code does `continue`
(`:297-300`) with no output. Silent under-coverage.

Census (see RUNBOOK for the exact command) — binary registry vs HEAD:

| service | tag | registry | cannot see |
|---|---|---|---|
| `component-render-check` | v1.0.1258 | 160/169 | 9 actions |
| `shared-output-fields-check` | v1.0.1265 | 161/169 | 8 actions |
| `removed-config-keys-check` | v1.0.1285 | 165/169 | 4 actions |
| `verifier-remit-check` | v1.0.1289 | 165/169 | 4 (but it does not read the registry) |

`publish_site` and `retract_asset_files` are missing from **all four**.

### Step 5 — confirm the freeze is live, not just in the repo

```
kubectl -n ai-persona-system get cronjob -o custom-columns=\
'NAME:.metadata.name,SCHEDULE:.spec.schedule,SUSPEND:.spec.suspend,\
IMAGE:.spec.jobTemplate.spec.template.spec.containers[0].image,LAST:.status.lastScheduleTime'
```

All four unsuspended and all four ran on 2026-08-18 (06:25Z, 06:55Z, 07:10Z,
07:25Z) on the frozen tags. This is the step that makes it a live defect rather
than a repo observation.

The same command also resolved the "pins nothing" rows from the overlay sweep:
`optional-key-budget-check`, `single-owner-carriers-check`,
`concept-register-drift-check`, `component-fallback-check`,
`site-discovery-staleness-check` all run **`postgres:16-alpine`**. They carry no
build of ours, so they are outside this class. Decision B's scope is six.

---

### Missteps this session

- **I counted test-only registrations first.** The opening census read 201 actions
  at HEAD and "5 missing"; the production figure is 169. `git grep
  RegisterActionInputSpec` over `platform/ internal/` picks up `_test.go`
  registrations (`strict_marker_test_action`, `test_action_removed_keys`,
  `test_validator_removed_*`, …), which are not in any shipped registry. Caught
  before it reached a doc, by noticing that four of the "missing" names were
  obviously fixtures. **The cheap check: any census of a code population must
  state its `:(exclude)*_test.go` either way** — a test fixture is a different
  population from a shipped one, and here it was ~22% of the count.
- **My first staleness pass counted `_test.go` and `testdata/` files as shipped
  surface** — 372 files where the real figure is 197. Same root error, twice in
  one session, on two different measurements.
- **The literal-vs-wrapped counting risk was real and had to be closed.** 170
  non-test `RegisterActionInputSpec(` calls exist but only 164 have the name on
  the same line; a one-line grep therefore undercounts by five. Re-ran with `-A1`
  (169 unique names) and the **missing sets came out identical** — the five
  wrapped registrations all predate all four builds. Worth recording because the
  robustness is what licenses the table, not the count itself.
- **Nearly asserted `verifier-remit-check` was stale on the strength of its
  165/169.** It links the package that holds the registry but never reads it, and
  neither symbol it does use has changed. A count inherited from a linked package
  is not a property of the service. Kept out of the "blind" group deliberately.

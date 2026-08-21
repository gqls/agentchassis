# REVIEW 2026-08-21 — `git_commit`'s accumulated optional-key surface, on the occasion of its 11th key

Written because `bugs_open/198`'s fix adds `file_shrink_floor` to `git_commit`, and
RFC_022's whole point is that **ten individually inert opt-in fields are a shared action
nobody understands**. This is the one review an author owes when they add to that pile.

It is deliberately NOT an RFC. Per the owner ruling of 2026-08-11, an opt-in field whose
unsafe default is OFF and which no live consumer names is not architecture-scope — and
that ruling requires the consumers to be **enumerated, not asserted**, which is what §2
does. What follows is the accumulated-surface review, plus one finding that is genuinely
uncomfortable and is stated rather than buried.

## 1. The uncomfortable finding first: this action is INVISIBLE to the budget check

`git_commit` has **no `RegisterActionInputSpec`**. Every one of its config keys is read
ad-hoc from `params.StepConfig.Config`. The consequence is not "it counts as zero" — it is
that **it cannot be counted at all**:

```
cmd/config-key-audit/optionalbudget.go — censusOptionalKeys() iterates
ListActionInputSpecNames() and counts len(spec.Optional); an action with no spec is
SKIPPED. It lands instead in censusUncountedActions(), whose own label reads:
"NOT COUNTED — no ActionInputSpec, so the optional surface is UNKNOWABLE, not zero."
```

So WFA-013's daily CronJob — the mechanism that exists precisely to notice the tenth
optional key — has never seen this action, and would not have noticed the tenth, the
eleventh, or the twentieth. The budget is N = 10 (owner, 2026-08-14). By hand count
`git_commit` was **at exactly 10 before this change** and is now at 11.

**This is a real gap in the check, not a technicality, and it is not closed here.**
Registering a spec would be the fix, and it is a separate change with its own risk:
`ExtractActionInputs` / `CheckConfig` turn unrecognised keys into warnings (and under
`StrictConfig`, errors) for **every** step of the action, and there are 19 live `git_commit`
steps across 17 agents. Doing that inside a bug fix would be exactly the "a platform seam
arrived inside a bug patch" move the guardian seat vetoes, and rightly. It is named as a
follow-on with its cost stated, which is the honest version of "not now".

## 2. The enumeration (measured 2026-08-21, not asserted)

**The 11 keys**, read from the source rather than from any doc:

| key | read at | what it does |
|---|---|---|
| `commit_message` | `git_deployer_actions.go` `buildCommitMessage` | Go template; context is ONLY `{domain, file_count, filename}` |
| `commit_message_field` | `resolveCommitMessage` | opt-in; dot-path read verbatim from CollectedData (DGH-007) |
| `content_field` | `extractFilesForGit` | single-file content path |
| `domain_field` | `GitCommitAction` | dot-path to the domain; default `"domain"` |
| `file_path` | `extractFilesForGit` | static repo-relative path |
| `filename_field` | `extractFilesForGit` | filename from CollectedData |
| `files` | `extractFilesForGit` | literal files map (legacy) |
| `files_field` | `extractFilesForGit` | dot-path to a map of files; default `site_files.files` |
| `page_field` | `extractFilesForGit` | page object → url/slug/name/filename/id |
| `repo_name` | `helpers.go` `resolveGitRepoNameDB` | explicit repo; else `site_record.github_repo`; else the `sites` row; else `"sites"` |
| **`file_shrink_floor`** | `shrinkFloorForGitData` | **new** — opt-in size floor, absent/0 = OFF (DGH-016) |

**The 17 carrier agents**, from the live `agent_definitions` rows:
content-feed-orchestrator, css-patch-agent, deployer-agent, model-directory-publisher,
nav-link-fixer, nav-updater, page-rebuild, page-rerender, pageflow-builder, report-builder,
rerender-pages, rerender-site, section-editor, site-asset-renderer, site-deployer,
site-work-orchestrator, webdesign-agent.

**Consumers of the new key: exactly one** — css-patch-agent's `deploy_css`, opted in at
`0.5` by migration 542. The other 16 agents send a byte-identical payload and make no
additional API call, which is pinned by `TestUnconfiguredCallerMakesNoContentsCall` rather
than argued.

> ⚠ **Enumerate with a RECURSIVE walk of `default_config`, never a one-level
> `jsonb_each`.** Loops carry their real work in `config.sub_workflow.steps`, so a shallow
> census under-reports. This is not a hypothetical: on 2026-08-21 a one-level census of
> `status='deployed'` stampers reported three, the recursive walk found six, and the three
> it missed were the page-BUILDING paths — the ones that mattered (DGH-015's struck-through
> bullet, caught by the council's guardian seat).

## 3. Is the surface coherent, or has it accreted?

Reading the 11 together, they are **not eleven independent knobs**. Nine of them are one
question asked nine ways — *"where do the bytes and the filename come from?"* — and they are
mutually exclusive in practice, resolved in a fixed precedence order inside
`extractFilesForGit` (`files_field` → `files` → `content_field`, with `file_path` /
`filename_field` / `page_field` naming the target). The remaining three are genuinely
separate concerns: **where it goes** (`repo_name`), **what it says** (`commit_message`,
`commit_message_field`), and now **whether it is allowed** (`file_shrink_floor`).

So the honest characterisation is: **one over-broad input selector plus three orthogonal
policies.** The accumulation risk RFC_022 is aimed at — "ten inert fields nobody
understands, whose combination is a shared action nobody can reason about" — is lower here
than the raw count of 11 suggests, because the nine are alternatives rather than
combinables. The count is high; the *combinatorial* surface is not.

That is a reason to be less alarmed, not a reason to skip the review. And it points at
what a future tidy-up should actually do: the win is collapsing the nine-way input selector
into one declared input shape, not shaving the three policies.

## 4. What this review commits its author to

1. `file_shrink_floor` ships opt-in, default OFF, one consumer, enumerated above.
2. The uncounted-action gap is **disclosed** (§1) rather than silently inherited, and named
   as a follow-on with its cost: registering an `ActionInputSpec` for `git_commit` would
   emit unrecognised-key warnings across 19 live steps and must be its own change with its
   own round.
3. If a **12th** key is proposed before that follow-on lands, this document is the thing to
   re-read first — and the answer should probably be "register the spec, then add the key",
   because at that point the pile is growing faster than the check that cannot see it.

## Sources

- `platform/orchestration/actions/git_deployer_actions.go`, `helpers.go` (`resolveGitRepoNameDB`)
- `cmd/config-key-audit/optionalbudget.go` (`censusOptionalKeys`, `censusUncountedActions`)
- `scripts/audit-optional-key-budget.sh`; `architecture_review/optional_key_budget_acks.json`
- RFC_022 and the owner rulings of 2026-08-11 (trigger narrowed) and 2026-08-14 (N = 10)
- Concept register DGH-007 (the opt-in-key precedent on this same action), DGH-016 (this key)
- `bugs_open/198`; council submission `5f756c51-cdc6-4a48-b5f9-59e472243601`

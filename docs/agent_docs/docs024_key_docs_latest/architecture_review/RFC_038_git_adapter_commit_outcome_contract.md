# RFC 038 — the git-adapter's commit reply carries no evidence, and widening it touches 19 live steps

**Raised 2026-08-19** by the `bugs_open/315` lane, **at the council gate's own direction**. The
`architecture` seat returned `ARCHITECTURE_SIGNAL: needs_rfc` on submission
`377167cd-6324-4bc7-a866-87ad8c435132` and named the remedy:

> *"Recommend splitting: ship edit 1 now; take edits 2-3 (adapter contract + payload) through
> architecture_review with the 19-step consumer list and a rollback plan, before edit 4's config key
> can safely reference the new fields."*

This RFC is that referral. **Status: SURVEY DONE 2026-08-19 (§7) — the seat's one `missing` item is answered, and the answer is
that NO consumer parses this reply at all. Owner ruled the destination the same day (§8).**

## 1. What is actually missing

`internal/adapters/git/github_client.go:68` — `CommitToRepo` creates blobs, builds a tree, creates a
commit, moves the ref, and then:

```go
newCommitSHA, err := c.createCommit(ctx, repo.Owner.Login, repo.Name, data.CommitMessage, newTreeSHA, latestSHA)   // :256
err = c.updateRef(ctx, repo.Owner.Login, repo.Name, branch, newCommitSHA)                                          // :261
if err == nil { return repo.HTMLURL, nil }                                                                          // :263
```

**The sha it just computed is discarded, and a per-repo constant is returned in its place.**
`repo.HTMLURL` is `https://github.com/gqls/sites` for every commit to that repo, on every site, for
ever. `handleCommitAction` (`adapter.go:438`) therefore replies with
`{success, repo_url, repo_name, domain, files, files_count, file_path, commit_message, timestamp}` —
nothing that names *what was written*.

Two consequences, both measured:

- **No consumer can distinguish a commit that wrote a blob from one that wrote nothing.** Register
  `DGH-009` records the mechanism: *"an unchanged file commits as an EMPTY commit and the adapter
  reports success with the file listed in `deploy_result`."*
- **`pages.deployed_at` cannot be gated on the deploy's outcome**, which is `bugs_open/315`. The
  page-rerender seed's own output contract already promises
  `"deploy_result": "git commit result with commit_sha"` (`sql_for_agents/034_page_rerender_agent.sql:99`)
  — a promise nothing has ever kept.

## 2. Why it is architecture-scope and not a bug fix

The seat's reasoning, which I accept in full:

> *"That reply is the shared wire shape consumed by 19 live `git_commit` steps across 16 agents (the
> plan's own MEASURED count), most of which the author has not surveyed for how they currently parse
> the response. … Widening a contract 19 steps already touch, on the strength of 3 code-verified
> callers, is architecture scope even though each individual caller compiles unchanged today: the
> risk is future steps starting to read `commit_sha`/`no_change` informally, at which point it is
> load-bearing with no RFC behind it."*

Note this is **not** the RFC_022 exception. The proposed payload fields would be added
**unconditionally to every reply**, not behind an opt-in whose unsafe default is OFF — the seat
raised exactly that as its second objection. An opt-in field nobody names is inert; a widened shared
reply is not.

## 3. What is proposed

1. `CommitToRepo` returns `(CommitOutcome, error)` where
   `CommitOutcome{RepoURL, CommitSHA, Branch string; NoChange bool; AbsentPaths []string}`.
2. `handleCommitAction`'s reply gains `commit_sha`, `no_change` and `files_sha256` (committed path →
   sha256 of the exact bytes); `handleDeleteFileAction`'s gains `commit_sha` and `absent_paths`.
3. `no_change` is **report-only**. Suppressing the empty commit would change what the seam
   guarantees for all 19 steps (today every success fires the deploy workflow and a Cloudflare
   purge) and is deliberately NOT proposed here.

### Settled already, so the RFC need not re-ask

- **There is no interface to ripple through.** The `guardian` seat flagged that `interface.go` has a
  `TreeEntry` landmine for hidden consumers. `[MEASURED 2026-08-19]`
  `grep -rn "CommitToRepo(ctx context.Context" --include=*.go .` returns **one line** — the concrete
  method on `*GitHubClient`; `interface.go` does not mention it; `grep -rln "GitClient\b"
  --include=*.go .` returns **nothing**. So the Go-side change is 3 callers, all inside
  `internal/adapters/git/adapter.go` (`:438`, `:518`, `:710`), and the delete tests are `_, err :=`
  and compile unchanged.
- **Nothing transforms the bytes downstream**, so a hash taken at the adapter is comparable with
  served bytes: the serving hop *"maps `hostname + path` straight onto a B2 object key"*
  (`deployment-github.md:125`) and the hop before it is a `b2 sync`.

## 4. What this RFC must NOT skip — the consumer survey

The seat's `missing` item is the whole of the work:

> *"No list of the 19 `git_commit` steps' current consumers verifying none will misparse the widened
> payload."*

**Do not answer this by measuring that nothing breaks today.** The 2026-07-29 ruling 3 is explicit:
a shared mechanism's other consumers must be **told**, not merely measured — the useful message is
what changed about their guarantee, not a list of new keys. The 19 steps and their `output_field`s
are enumerated in
`bugfix_315_deployed_at_without_publication/NOTES_deployed_at_without_publication.md`; note they use
**nine distinct field names**, and two set none.

⚠ And a real parsing hazard to check rather than assume: **`deploy_result` already arrives in two
shapes.** `[MEASURED 2026-08-19]` over 744 orchestration rows in 7 days, **57 (7.7%)** are nested one
level deeper (`deploy_result.response.deploy_result.response.data.…`) because the deploy was done by
a called sub-agent. Any consumer survey that reads one shape will report the other as unaffected.

## 5. The question for the owner, kept separate on purpose

`sql_for_agents/356:105-118` states that `pages.deploy_commit` was dropped **deliberately** by
`sql_for_tables/003` as *"belongs in page_components"*, and that:

> *"deciding whether to wire it or drop it is an **owner call, not a bug fix**."*

So even if this RFC is accepted and the sha becomes available, **where it is persisted is not this
lane's decision.** `page_components.deploy_commit` exists (0 of 1,775 rows, no writer anywhere in
the repo including tests); `pages.content_hash` exists (0 of 786, likewise) and has now been found
dead and passed over independently three times — by migration 291 on 2026-08-02, by migration 356,
and by this lane. The owner's call is: wire them, or drop them.

## 6. Rollback

The Go change is additive in shape (a struct return gains fields; no capability removed) and ships
in the git-adapter image, so rollback is repointing to the previous image — no data migration, no
config to unwind, and the 3 callers are in the same image. Nothing may name the new fields until
this RFC is decided, which is why `bugs_open/315`'s config key was withdrawn from the same round
rather than shipped ahead of it.


## 7. THE CONSUMER SURVEY — done 2026-08-19, and it is decisive

The seat's `missing` item was: *"No list of the 19 `git_commit` steps' current consumers verifying
none will misparse the widened payload."* Answered from both sides.

### 7a. Workflow config — every reference is a blind re-export

For each `git_commit` step, its `output_field`, then every OTHER step in the same agent whose JSON
mentions that field name:

| agent | output_field | referenced by | as |
|---|---|---|---|
| content-feed-orchestrator | `news_commit_result` | `complete` | `complete_workflow` |
| content-feed-orchestrator | `rss_commit_result` | `complete` | `complete_workflow` |
| css-patch-agent | `css_deployed` | `complete` | `complete_workflow` |
| model-directory-publisher | `directory_commit_result` | `complete` | `complete_workflow` |
| page-rerender | `deploy_result` | `complete` | `complete_workflow` |
| report-builder | `deploy_result` | `complete` | `complete_workflow` |
| report-builder | `sidecar_deployed` | `complete` | `complete_workflow` |
| section-editor | `git_result` | `complete` | `complete_workflow` |
| site-asset-renderer | `deploy_result` | `complete` | `complete_workflow` |
| webdesign-agent | `css_deployed` | `complete` | `complete_workflow` |
| the remaining 7 (`js_snippets_deployed` ×6, `failed_sidecar_deployed`) | — | **nothing** | — |

**Every single reference is `complete_workflow`'s `output_fields` list**, which re-exports the whole
blob to the parent untouched. **Not one step reads a FIELD out of a `git_commit` reply** — no
`*_field` path, no conditional, no template, anywhere in the live fleet.

⚠ **One row in the first pass was a false positive and is removed above:** `report-builder`'s
`publish_failed` appeared to reference `sidecar_deployed`, because `LIKE '%sidecar_deployed%'` also
matches **`failed_sidecar_deployed`**, which is that step's own `output_field`. A substring match
between two field names that share a suffix. Checked and dropped.

### 7b. Go — the apparent readers are all WRITERS

Grepping the payload's key names outside the adapter and `git_deployer_actions.go`:

| key | non-adapter Go hits | what they actually are |
|---|---|---|
| `repo_url` | **0** | — |
| `deploy_result` | **0** | — |
| `css_deployed` / `js_snippets_deployed` | **0** | — |
| `files_count` | 3 | `zap.Int("files_count", len(filesMap))` — log fields |
| `file_path` | 4 | `"file_path": "feed.xml"` — constructing OUTBOUND payloads |
| `commit_message` | 8 | building a message to SEND |
| `git_result` | 3 | `"git_result": gitResult` — wrapping an action's own result |

**No Go code reads a field out of the git-adapter's reply either.**

### 7c. What this settles

The reply is an **opaque blob** end to end: produced by the adapter, stored under `output_field`,
re-exported by `complete_workflow`, never parsed. **Adding keys to it cannot make any current
consumer misparse anything, because there is no current consumer.** That is the fact the seat asked
for, measured rather than argued.

It does **not** dissolve the seat's forward-looking concern — *"the risk is future steps starting to
read `commit_sha`/`no_change` informally, at which point it is load-bearing with no RFC behind it"* —
which is precisely why this RFC exists and why the fields are documented in `DGH-001` as they ship.

## 8. OWNER RULING, 2026-08-19

> *"wire up the page fingerprint; drop or ignore the section one."*

- **`pages.content_hash` — WIRE IT.** It is the per-page fingerprint of the bytes committed. The site
  serves one file per page, so this is the grain that answers *"is this page current?"* in one
  comparison.
- **`page_components.deploy_commit` — NOT wired.** A section is not a file, so it cannot answer the
  publish question. **Implemented as "ignore", not "drop":** dropping a column is irreversible and
  another lane may yet have a use, whereas the actual cost being paid — three sessions independently
  rediscovering it is empty — is a documentation cost and is fixed by saying so. Recorded in
  `DGH-001` and left in place.

So the destination is settled and this RFC's remaining question is only the one in §4, now answered
in §7. **The adapter change may proceed**; `files_sha256` is load-bearing (a commit sha alone cannot
distinguish an empty commit, per `DGH-009`), `commit_sha` is traceability, and `no_change` stays
report-only.

# RFC 038 — the git-adapter's commit reply carries no evidence, and widening it touches 19 live steps

**Raised 2026-08-19** by the `bugs_open/315` lane, **at the council gate's own direction**. The
`architecture` seat returned `ARCHITECTURE_SIGNAL: needs_rfc` on submission
`377167cd-6324-4bc7-a866-87ad8c435132` and named the remedy:

> *"Recommend splitting: ship edit 1 now; take edits 2-3 (adapter contract + payload) through
> architecture_review with the 19-step consumer list and a rollback plan, before edit 4's config key
> can safely reference the new fields."*

This RFC is that referral. **Status: OPEN, awaiting the consumer survey it asks for.**

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

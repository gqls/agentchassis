# RFC 011 — the git adapter's action vocabulary, and whether `delete_file` belongs in it

**Filed** 2026-08-03 by the `bugfix_098_unpublish_primitive` thread, at the direction of
a **council REJECTED verdict** (hard veto from `guardian`, `architecture` signalled
`needs_rfc`) — correlation `4a7f0877-4149-4431-97d4-318d093570a4`, round 2.

**Status:** **DECIDED 2026-08-03 — OPTION B**, by the owner, on this RFC's own recommendation.
`delete_file` **keeps its place on the adapter and LOSES its place in the generic
allowlist**: it is reachable only through `retract_page_deployment`, which carries guards a
config-assembled call cannot. Shipped in the same push as the decision; asserted by
`TestGitAdapterAllowlistExcludesDeleteFile`, because a decision that lives only in a comment
is one commit away from someone "completing" the allowlist.

**What that settles, and what it does not.** It settles *this* verb. It does **not** settle
the general question in §2 — whether a destructive verb differs in kind from an inert field
— which stays open for the next adapter verb, and option C (a separate destructive
vocabulary) remains the honest general answer if this recurs. Recorded so a later thread
does not read one worked example as a rule.

**The code being reviewed is already on shared HEAD and live** (chassis digest
`sha256:5da2888…`, git-adapter `sha256:df7bc0a…`). That is not defiance: forward-only
forbids removing it, `make build-*` builds from committed HEAD, and any other session's
roll ships it. The owner ruling of 2026-07-29 §2 retired the ordering exemption's first
condition for exactly this reason and states plainly that **review here is after the fact
by design**. What this RFC can still change is the wiring, the default, and whether the
verb keeps its place in the shared vocabulary.

---

## 1. What was built, in one paragraph

`bugs_open/098` and `bugs_closed/125` were both blocked on one absence: **the platform
can publish a page but had no implemented way to unpublish one.** The git adapter's only
deletion verb was `delete_repo`, which returns *"not yet implemented"*. The fix expresses
a deletion as a **kind of commit** — in the Git Data API a removal is a tree entry whose
`sha` is `null` — so `GitCommitData.Deletions` rides through the existing `CommitToRepo`
and inherits the ref-race retry, the `{domain}/{path}` prefixing, atomicity with writes
(a *move* becomes one commit), and the credential boundary. On top: a `delete_file`
adapter verb, its allowlisting in the generic chassis caller, and
`retract_page_deployment`, the page-level caller.

**The architecture seat endorsed the design and vetoed the packaging.** Verbatim:
*"expressing delete-as-null-sha inside the existing CommitToRepo path is the right reuse,
inheriting retry, prefixing and atomicity for free rather than bolting on a parallel write
path. On that axis the plan is sound and I'd carry it."* Nothing below disputes the
mechanism. The question is the vocabulary.

## 2. The question this RFC actually asks

> **Does adding a DESTRUCTIVE verb to a shared adapter's action vocabulary differ in kind
> from adding an inert type or field — such that the 2026-07-29 ruling's
> "additive-and-inert goes through the normal gate" does not cover it?**

My submission argued it does not differ: `delete_file` is reachable by nothing until a
workflow names it, so it is additive-and-inert by that ruling's own test. The seats
answered that the framing *describes today's blast radius rather than retiring the
trigger*:

> `architecture`: "Unlike an inert field addition, a DESTRUCTIVE verb on a general chassis
> changes what the vocabulary can be asked to do platform-wide… The author's own defence —
> citing the 07-29 ruling's additive-vs-inert distinction — is the author's judgement to
> make the case, not to relocate the review; the ruling itself says the call isn't the
> submitter's alone."

That is a fair reading and I do not think I can settle it. **It is the RFC's whole
question**, and it generalises beyond this change: the answer sets precedent for every
future adapter verb.

### 2a. The sub-question that carries most of the weight

`gitAdapterActions` (`git_adapter_request_action.go`) is the allowlist gating what **any**
workflow on the chassis may ask the git adapter to do. Before this change it held
`commit`, `create_branch`, `create_pull_request`, and its comment said destructive verbs
"are NOT reachable through the generic caller at all" — naming `delete_repo`.

I added `delete_file` and argued the line is not "destructive vs not":

- `delete_repo` destroys a container the platform cannot rebuild, and no workflow wants it;
- `delete_file`'s blast radius is exactly the paths named, every one recoverable from the
  repo's history, and it is the counterpart of `commit` — which is **already allowlisted,
  already writes content, and can already destroy a page by overwriting it**.

`guardian` did not accept that the second bullet settles it, and flagged the allowlist edit
as *"the point where the new verb becomes reachable by any future workflow author, not just
this retraction caller"*, with **no blast-radius or rollback note written for the vocabulary
change itself**. That omission is a fair hit and this section is the beginning of repairing
it.

## 3. Options, costed

**Option A — ratify as-is.** The verb stays in the vocabulary and in the allowlist.
*For:* the design is endorsed; the guards live in the action, not in workflow config, so a
future author cannot switch them off; `commit` is already destructive-by-overwrite.
*Against:* it treats a destructive verb as ordinary scope and sets that precedent for the
next adapter.

**Option B — keep the verb, remove it from the generic allowlist.** `delete_file` stays on
the adapter but is reachable only by `retract_page_deployment`, which has its own guards.
A future workflow author cannot reach it by writing config.
*For:* restores the allowlist comment's stated intent ("destructive verbs are not reachable
through the generic caller"); costs almost nothing today, since the only intended caller is
the retraction action.
*Against:* the next legitimate caller needs a code change rather than config — which some
would call the point.
**This is the option I would pick if it were mine to pick**, and I flag that preference
rather than hide it. It concedes the seats' actual complaint (unbounded reachability)
without discarding the capability, and it is a two-line change.

**Option C — split the adapter's write and destructive vocabularies.** A separate
allowlist, or a capability flag on the request, so destructive verbs are opt-in per caller.
*For:* the general answer; makes the class visible rather than case-by-case.
*Against:* a new mechanism, proposed in response to a veto about new mechanisms. Needs its
own justification and should not ride in on this.

**Option D — revert the verb, keep the primitive dormant.** `Deletions` stays in
`CommitToRepo` (it is inert with no caller), `delete_file` and the allowlist entry go.
*For:* the most contained.
*Against:* returns the platform to "can publish, cannot unpublish", which is the defect two
bugs are open on, and strands `bugs_open/098` with no repair path — the state that made
125's one live orphan a manual deletion by the owner.

## 4. What is NOT in dispute, so a reviewer need not re-derive it

Recorded so the split costs no measurement work, exactly as the guardian's note said it
should not:

- **Deletion semantics**, probed on the live repo (POST `/git/trees`, unreferenced object,
  no ref moved, no workflow fired): null sha on an existing path → **201**; on an absent
  path → **422 `GitRPC::BadObjectState`**. The existence filter is required, not defensive.
- **Path derivation**, the round-1 gating objection, answered by measurement: the real
  `PageFilePathFromURL` run over all 13 candidates against each page's own deploy repo —
  **11 of 12** `sites`-repo pages have a file at exactly the derived path, the 12th is
  genuinely absent and 404s, the 13th deploys to `vm-sites` and is correctly absent there.
  **Zero mismatches.**
- **`TreeEntry` call-site census:** exactly two construction sites, both updated,
  compiler-complete.
- **Live proof** of write → delete → idempotent re-delete on a scratch path the deploy
  workflow ignores, with the no-op logged and the branch head unmoved.
- **The far side already reconciles:** `b2 sync --delete` + a Cloudflare purge per changed
  domain, so removing a file from the repo genuinely retracts the page.

## 5. Owed regardless of which option wins

Four objections are about correctness, not packaging, and they land on code that is live.
They are tracked on `bugs_open/098` and in the workstream NOTES:

1. **HIGH** — the retraction's link census does not go through
   `linkablePageStatusPredicate`; a landmine states an offline census over `pages` must.
2. **MEDIUM** — inbound references also live in structured `content_data` fields
   (`link_url`, `cta_url`, `*_url`), which an `href=`-only scan misses. This one can let a
   retraction proceed past a real inbound link, which is the case the owner's directive of
   2026-08-03 exists to prevent.
3. **MEDIUM** — the inbound-source logic is copied from `check_orphan_pages.go` rather than
   shared with it; a drift warning in a comment is not a mechanism.
4. **MEDIUM** — the `status='active'` predicate is a third bespoke spelling rather than a
   consolidation onto a canonical helper. *(The census half of this objection was answered
   after submission: `queueDirectoryPageRerenders`, which calls itself "cousin of
   queueNewsPageRerenders", carried the identical defect — latent, 0 live rows — fixed in
   `8f73e7279`.)*

## 6. The one thing that is NOT waiting for this RFC

The `guardian` named it: *"The one edit I would approve standing alone is
`render_news_section_html.go`… that is the actual root-cause fix for the resurrection
bug."* That fix is live and is unaffected by whichever option is chosen. It stops an
archived page being re-rendered and re-published twice a day, which is a defect in its own
right and was never part of the vocabulary question.

**And the retraction itself has NOT been performed.** The owner approved retracting one
page before this verdict existed. Firing a vetoed capability at a live customer site on an
approval given under a different premise is the precise thing the veto is about, so it
waits for a decision here.

## 7. Pointers

`bugs_open/098` · `bugs_closed/125` · `bugs_closed/124` (the `$ctx.` precedent both seats
cite) · concept register **DGH-006** · workstream
`docs024_key_docs_latest/bugfix_098_unpublish_primitive/` · council correlation
`4a7f0877-4149-4431-97d4-318d093570a4` (rounds 1 and 2, full objection text in
`diagnosis_artifacts`).

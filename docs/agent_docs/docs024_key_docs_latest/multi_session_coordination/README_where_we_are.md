All four decisions from the coordination handoff are resolved, built, validated against the live cluster, and committed — one commit per deliverable with explicit pathspecs, practising the ruling this thread exists
to implement.

1. Pre-dispatch coverage check — built and proven live (4accef4e3). The 090 needs_diagnosis trigger now refuses to dispatch (unless FORCE=1) when any open work item already touches the target, running before the
   intake insert so a refusal writes nothing. It has three probes: a new PAGES= env matched across all sites (the July 16 incident spanned two sites, so keying to the single RUNTIME_SITE would have missed half of it),
   a seed_scope file-overlap probe — which resolves the handoff's open question: the coverage key for code-only diagnoses is the seed_scope file set, at file granularity — plus a non-blocking advisory when seed files
   are dirty locally, and a site-level fallback that blocks only on open items touched in the last 2 hours (older parked backlog is FYI-only, so operators aren't trained to always FORCE). Semantics are copied
   verbatim from silentCoverageClause, including "complete does not cover". Validation: it reproduces the match on the actual json-leak-fix-retry incident rows, and its first real run refused a dispatch at
   relojistas.com after surfacing two different sessions' open items on one page — a genuine collision-in-waiting found immediately.

2. Commit hygiene — adopted via a new repo-root CLAUDE.md (fc773d084). None existed, so every session now loads the practice at start. One refinement over the handoff: commit with an explicit pathspec (git commit
   <paths> -m …), which ignores whatever other sessions have staged in the shared index — I hit exactly that hazard while committing. Enforcement hooks deliberately not built; habit first, per §6.3.

3. Deploy blast radius — structural fix, announcements dropped (0c7b17616). I ruled that the structural fix dominates the doc_notes channel, for the handoff's own reason: a ref-built image cannot bundle WIP, while
   an announcement only works if it's read. make build-agent-chassis-ref REF=<ref> builds from a git archive context (verified complete — the dockerfile needs nothing gitignored; I didn't run the multi-minute docker
   build itself since the default tag is another session's staged release). The default working-tree target is unchanged in behaviour but now prints exactly what it would sweep in — 40 uncommitted changes at test
   time. With per-task commits, "wait for someone else's release" is obsolete: commit your task, ref-build, deploy.

4. File-claims — deferred, per the handoff's own recommendation.

Fittingly, the problem demonstrated itself throughout: three other-session commits landed mid-build (including the v1.0.1125 tag bump), a file I was testing against got committed under me, and MEMORY.md was
rewritten externally. All recorded in RUNNING_NOTES_2026-07-16_coordination_build.md (ccb434803) alongside the handoff, with small follow-ups noted there (084 bare trigger still unchecked by design; ref-builds
cover agent-chassis only). Memory index updated.

----

Done, pushed, and all three inversion commits are confirmed on origin.

What changed

The default is inverted. make build-<service> now builds from committed HEAD — a git archive into a clean context that structurally cannot bundle anyone's WIP. The old whole-tree behaviour is now the explicit
build-<service>-tree escape hatch. This is the piece that was missing: your commit is now what ships, so "commit your task when ready" finally connects to the build instead of being advice the build ignored.

The failure mode points the safe way. Forget to commit and the build leaves your change out — printing a yellow "N uncommitted changes are NOT in this image" — a wasted cycle the pod-grep catches. The old default
did the opposite: silently shipped everyone else's untested work to production. Both guards (dockerfile exists, REF is a real commit) block before docker build; I verified that at runtime, not just in dry-run.

REF=<ref> pins a specific commit, and when pinned the uncommitted-changes note is suppressed — pinning is already an explicit statement of what ships.

What this touched

deploy-agents.sh and deploy-production.sh inherit the safe default with no code change — they now build committed state, which is correct for deploy scripts. The one workflow consequence: you must commit before
running them, or the deploy ships HEAD without your change (loudly). That already matches what today's bug handoffs tell sessions ("never build from the working tree"). Frontends are deliberately untouched —
different build context, one-subdirectory blast radius.

Committed and pushed

Three commits for the inversion (eed45ed4b makefile, db40c7ba1 CLAUDE.md, 2f457e35b notes), each with an explicit pathspec — practising the rule. The branch pushed clean as a fast-forward; no force. Twice during
this the shared branch moved under me from other sessions' commits (local HEAD jumped to 9f515b0d7, then origin moved past my push to more of their commits) — the exact concurrency this whole workstream is about,
and it cost nothing because everything was committed narrowly and pushed forward-only.

One honest gap remains: the committed-ref build is proven through dry-run and its guards, but no one has watched it through a full docker build end-to-end yet — first real build should. It's in the notes and
memory.

----

Got it — and there's a conceptual thing to clear up first, because "see the builds being triggered" suggests you might be picturing a build server. There isn't one.

There's no CI — builds are local and synchronous

No .github/workflows, no build queue, no dashboard. make build-<service> runs docker build on the machine where you type it, and the output streams to that terminal. That is the build being triggered and the way
you watch it. The inversion changed only what goes into the image (committed HEAD vs working tree) — it added no trigger mechanism and no place to watch. If you're not looking at the terminal running make, there's
nothing else observing it. (The github-actions-runner-vmsites in the tree is for the separate VM-sites work, not chassis images.)

The full chain for the chassis

Same command as before — the only new rule is commit first, because it now builds your commit, not your dirty tree.

# 1. commit your task (pathspec — see CLAUDE.md)
git commit platform/... -m "..."

# 2. bump the tag — a same-tag rebuild ships the node's stale cached image
#    (edit makefile line 16, or pass IMAGE_TAG= on each command below)
#    live pod is currently v1.0.1130, so next is v1.0.1131

# 3. build — watch this terminal; it prints the resolved sha and, if you
#    forgot to commit, how many changes it's LEAVING OUT
make build-agent-chassis IMAGE_TAG=v1.0.1131

# 4. push to the registry (there is NO push-agent-chassis target — see below)
make push-backend IMAGE_TAG=v1.0.1131        # pushes all backend images
#   or just the one:  docker push docker.io/aqls/agent-chassis:v1.0.1131

# 5. deploy — rewrites the kustomization newTag and kubectl apply's it
make deploy-agents IMAGE_TAG=v1.0.1131

Watching each stage land (this is your "see it happening", since the build itself is just terminal output):

docker images | grep agent-chassis                                    # image built locally
kubectl rollout status deployment/agent-chassis -n ai-persona-system  # rollout progressing
# the golden rule from CLAUDE.md — prove the running pod has YOUR code, not the tag's promise:
POD=$(kubectl -n ai-persona-system get pod -l app=agent-chassis -o name | head -1)
kubectl exec -n ai-persona-system $POD -- sh -c 'strings /app/agent-chassis | grep -c "<your log string>"'

make release-backend chains 3→5 for the whole backend in one go if you'd rather not step through it.

One broken reference I found while checking

scripts/deploy/deploy-agents.sh calls make ... push-agent-chassis, but that target doesn't exist — only push-backend/push-all and a handful of per-service ones (not chassis). That script would fail at the push
step. It's pre-existing, not from my change, but it's exactly the kind of paper-cut that wastes a deploy. Want me to add a push-agent-chassis target (and the sibling per-service push targets) so that script and the
obvious make build-X push-X muscle-memory both work? It's a few lines and I'd commit it on its own.

----

● Short answer: each thread is now told to commit its own code, but nothing guarantees it — and you should not have to commit per thread. Here's the real picture.

What "knows to" actually means

CLAUDE.md is live on origin (its last commit db40c7ba1 is confirmed on the branch). Claude Code loads a repo-root CLAUDE.md at session start, so every newly-started thread now reads "commit per task, explicit
pathspec" before it does anything. That's the mechanism by which a thread "knows."

But three gaps mean you can't treat it as guaranteed:
- Threads already running when it landed never re-read it — they load it once, at start.
- It's behavioural, not enforced (you ruled out the hook), so "read it" ≠ "did it."
- The tree proves the gap right now: 48 uncommitted files (34 modified, 14 untracked) across several threads' areas — docs, platform, leopardess. That's the normal steady state, not an incident. Threads leave work
  sitting.

But you do not need to commit per thread — and shouldn't try

Two reasons:

1. You usually can't tell whose it is or whether it's finished. Committing another thread's half-done file by hand is the same mistake as git add -A, just slower. The thread that wrote it is the only one with the
   context to know when it's coherent. That's precisely why the rule puts the commit with the thread that did the work.
2. The inverted build makes the backlog harmless to your deploys. This is the reassuring part: since make build-<service> now builds from committed HEAD, any uncommitted work — yours or anyone's — simply isn't in
   the image, and the build prints "N uncommitted changes NOT in this image." So you never have to sweep up other threads' work to ship yours. You commit your thread's change, build from HEAD, and everyone else's
   uncommitted junk is excluded automatically.

So the practical protocol

- Let each thread commit its own — that's the design, and new threads are instructed to.
- You only need to care about commits at the moment you deploy, and only for the specific change you're shipping. If a deploy depends on thread X's fix, make sure thread X committed it (ask that thread, or grep the
  pod afterward per the golden rule). You don't audit the whole tree.
- The generic backlog of uncommitted docs/notes: leave it. Forward-only means nothing is lost; the owning threads commit eventually.

One footnote so it doesn't mislead you: a chunk of that "uncommitted" count (~17 of the kustomize files) isn't thread work at all — it's the newTag: line that make deploy-agents rewrites every deploy. That's deploy
exhaust, not a thread that forgot to commit.

So: the honest state is "threads are asked, not made, to commit — and the build is now designed so that their forgetting can't hurt your deploy." If you ever want that to be a guarantee rather than a request, the
only lever left is the enforcement hook you declined; the alternative is exactly what's in place now.

----

The current branch — 085_debug_and_feature_loops — not master. And it matters, because master is a long way behind.

Why

make build-<service> defaults to REF ?= HEAD, and HEAD is just a pointer to whatever branch is checked out — right now the tip of 085_debug_and_feature_loops (202019e6c). So a default build tars this branch's
committed state into the image, pushes it, and the deploy points the cluster at that image. Nothing in the build or deploy path ever consults master/main.

origin/main is 131 commits behind this branch. If deploys came from main you'd be shipping code four months stale and missing essentially everything this team has done. They don't — this feature branch is
effectively the trunk here, and that's what ships.

The one thing that does default to main — don't let it confuse you

The diagnosis trigger (090/084) defaults to REF=main, but that's a different ref for a different job: it's the git ref the diagnosis agent fetches from GitHub to analyse code, not the binary that runs in the
cluster. So there are two independent "which ref" answers:

┌───────────────────────────────┬───────────────────────┬────────────────────────────────────────────────┐
│            Action             │       Ref used        │                     Result                     │
├───────────────────────────────┼───────────────────────┼────────────────────────────────────────────────┤
│ make build-<service> → deploy │ HEAD (current branch) │ what actually runs in the pods                 │
├───────────────────────────────┼───────────────────────┼────────────────────────────────────────────────┤
│ 090/084 diagnosis run         │ main (default)        │ what the analyser reads — stale by 131 commits │
└───────────────────────────────┴───────────────────────┴────────────────────────────────────────────────┘

That second row is a real trap worth flagging: a diagnosis run left on the default REF=main is analysing four-month-old code while the pods run this branch — so it can "find" bugs already fixed, or miss the actual
deployed code entirely. When you diagnose, pass REF=085_debug_and_feature_loops (or the exact sha you deployed) so the analysis matches what's live.

To be concrete about your deploys: whoever runs make build-…  gets their checked-out branch's HEAD. Since everyone works on 085_debug_and_feature_loops, that's consistent — but if a session ever checked out a
different branch and built, it would ship that branch. The image is tied to the builder's current checkout, not to any fixed branch.



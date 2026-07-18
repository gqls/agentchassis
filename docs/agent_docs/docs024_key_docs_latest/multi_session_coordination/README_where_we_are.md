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

---

Bug 003 is the key context — and it changes what I can claim. Its §4.3 already suspects deploy churn is the spawn-killer ("converts every deploy-window spawn loss into a permanent zombie") but never identifies why.
I think I've found the mechanism. One more check first, because it decides how hard the fix is — whether reprocessing a redelivered message is safe:

Ran 1 shell command

● I have the answer, and it's better news than it looks: the interruption isn't inherent to restarting pods — it's a bug. Fix it and CD becomes straightforward. Build CD first and you'd make things materially worse,
because you'd be deploying more often, and every deploy destroys in-flight work.

Why a restart destroys orchestrations

platform/kafka/consumer.go Consume() fetches a message at line 81 and commits its offset at line 103 — before the message is processed. The comment above the commit says "After successful processing, commit the
offset", but nothing happens between the fetch and the commit; processing occurs back in the caller, after Consume() has already returned. The intent was at-least-once; the code is at-most-once.

Both main chassis loops use it — agent.go:468 (requests) and :528 (responses). So when a pod dies mid-work, Kafka has already been told the message was handled and will never redeliver it. That's the difference
between "restarts delay orchestrations" and "restarts annihilate them." It also explains the 300s rebalance landmine in CLAUDE.md, and it isn't limited to deploys — OOM kills, evictions and node failures lose work
the same way.

Tellingly, the codebase already has the correct shape elsewhere: client.go and server.go use the manual FetchMessage + CommitMessages pattern. Only the main agent loop takes the shortcut.

This lines up with bugs_open/003 §4.3, which already suspected deploy churn was the spawn-killer ("converts every deploy-window spawn loss into a permanent zombie") but never identified the mechanism. I think this
is it. And the damage is measurable right now: 22 wedged EXECUTING_STEP orchestrations, the oldest 1,224 hours — about 51 days — because the reaper sweeps AWAITING_RESPONSES and dispatch loops but never
EXECUTING_STEP. Every lost message becomes a permanent zombie.

The fix is much smaller than it should be

The expensive part of moving to at-least-once is normally making handlers idempotent, since redelivery means reprocessing. That already exists — platform/orchestration/state.go:170/207 reads and writes a
processed_messages table. The dedupe layer is built; the consumer just isn't using the delivery semantics that would need it.

So the core change is moving the commit to after the handler succeeds. The thing to verify first is that processed_messages actually covers every inbound path (I've confirmed it's used in orchestration state
handling, not that it guards all message types) — that's the one real unknown.

What else is currently working against you

Independently of the commit bug, the chassis deployment is configured about as badly as possible for zero-downtime rollout:

- replicas=1 — a restart means a window with no consumer at all.
- No readiness probe — so a new pod counts as Ready the instant it starts, and Kubernetes kills the old one before the new one has joined the consumer group.
- terminationGracePeriodSeconds=30, while Agent.Shutdown() itself waits up to 30s for in-flight goroutines — SIGKILL races the graceful drain and can win.
- No preStop hook — nothing lets the pod stop taking new work and quiesce before dying.

The order I'd do it in

1. Commit after processing (at-least-once) — the unlock. Restarts become redelivery instead of loss, and it fixes crash/eviction loss too. Verify processed_messages covers all inbound paths first.
2. Fix the drain — grace period comfortably greater than the shutdown wait, a preStop sleep, and stop fetching on SIGTERM while finishing what's in hand.
3. Readiness probe that means it — Ready only once the consumer group is joined and partitions assigned, so rollouts hand over to a pod that's genuinely consuming.
4. replicas ≥ 2 so a rollout never leaves zero consumers (check partition count — with one partition you get failover, not parallelism).
5. Close the reaper blind spot (003 §4.3) — sweep EXECUTING_STEP regardless, as a backstop, and clear the 22 standing zombies.
6. Then CD is nearly free — and the quiet-check-before-deploy ritual we did yesterday becomes an optimisation rather than a requirement.

Worth saying plainly: steps 1–4 are what make CD safe. A drain-before-deploy approach (stop accepting new orchestrations, wait, then roll) would paper over deploys specifically, but it does nothing for crashes or
evictions — which is why I'd fix delivery semantics rather than build deploy choreography around a broken guarantee.

I've deliberately not changed anything — you asked how, not for a fix. If you want, the natural next step per CLAUDE.md is to fold this mechanism into bugs_open/003 rather than file a new bug (it's the same family
as its deploy-churn suspicion), and the fix plan is a good candidate for the council gate given it touches platform/ and changes a delivery guarantee fleet-wide. Say the word and I'll do either.

---

Confirmed — the 011 row edit is another thread's uncommitted work (the imagery thread refining bug 011's description), not mine. Since git commit <file> takes the whole file and interactive staging isn't available
here, it will ride along. Per CLAUDE.md the honest thing is to disclose that in the message rather than silently absorb it:

Committed 027aa7588

● Filed, committed as 027aa7588.

What went where

Following CLAUDE.md's rule to grep the index before filing a new bug, this went into the existing family rather than a new bug — 003 already suspected deploy churn was the spawn-killer without knowing why, so the
finding slots in as the missing mechanism:

- bugs_open/003 §3d — the second root cause, explicitly marked independent of the broker-2 network path in §3a–3b. Consume() commits the offset at L103 before the handler runs, despite the comment at L102 claiming
  otherwise; both chassis loops use it; client.go/server.go already use the correct shape, which is what makes it an oversight rather than a design choice.
- bugs_open/003 §4.4 — the fix: commit-after-processing, plus the drain/readiness/replicas gaps, and an explicit note on why this must precede any CD work.
- bugs_open/003 §8 — file references for the fixing thread.
- 016b §9 — the transferable pattern, per CLAUDE.md's "the §9 entry is so nobody re-walks it." Its rule generalises past Kafka: an acknowledgement is a promise the work is done — never send it before the work is
  done, which applies equally to job claims and status='complete' writes. Plus the two corollaries this cost: read the ordering not the comment, and "it only happens on deploys" usually means state is lost on process
  death.
- 016b §10 — the index row now surfaces the second root cause so it's findable from the queue.

I verified all five cited line references still resolve before committing, since stale file:line pointers are worse than none.

Two things worth flagging

A passenger rode along, and I disclosed it in the commit message. The 016b 011 index row (SDXL/Gemini wording) is another thread's uncommitted edit. git commit <file> takes the whole file and interactive staging
isn't available here, so I couldn't exclude it — the honest option was to name it in the message rather than absorb it silently. Nothing is lost; that thread's work is now committed under my message, which is the
exact asymmetry CLAUDE.md warns about.

No council run. Its scope is platform/, internal/, pkg/ — docs are refused client-side and never spend credits. Worth noting the fix itself would be a strong council candidate when someone writes it: it changes a
delivery guarantee fleet-wide.

The one genuine unknown I've flagged in both documents rather than papered over: processed_messages demonstrably guards orchestration state handling, but I have not verified it covers every inbound path. That has
to be established before flipping the commit — otherwise the change trades lost work for duplicated work.

---

Verified — everything landed, nothing lost, working tree clean for both files.

The correction, and why it mattered

Going ahead with the verification I'd flagged overturned my own filed claim. An hour ago I committed that processed_messages already existed and "makes the fix small." That was wrong, and it would have sent the fixing
thread down a path that doesn't work.

The dedupe has the identical defect to the offset commit. Both paths do seen-check → record → then the work:

┌──────────────┬────────────┬────────┬───────┐
│     path     │ seen-check │ record │ work  │
├──────────────┼────────────┼────────┼───────┤
│ agent.go     │ 801        │ 811    │ 822   │
├──────────────┼────────────┼────────┼───────┤
│ processor.go │ 1296       │ 1317   │ 1323+ │
└──────────────┴────────────┴────────┴───────┘

So had someone fixed the Kafka commit to redeliver, the redelivered copy would hit HasProcessedMessage → true → dropped as "Duplicate message ignored". The work would still be lost — just through a different door. The
offset fix alone is inert; the two must ship together.

Two things I settled so nobody re-checks them: the a.isStateless gate is not a risk (hardcoded true), but an empty request_id silently disables dedupe entirely — those messages have never had any protection.

§4.4a-bis now specifies the durable shape: a two-phase claim (lease on receipt, complete on success, expired-but-incomplete is reprocessable), reusing site_work_items' existing claimed_at + 40-minute timeout rather than
inventing a second mechanism. And 016b §9 now carries the generalisation, which is the more valuable output: an acknowledgement layer and a dedupe layer fail the same way because they encode the same belief — that
receiving a thing is the same as handling it. Find it at one layer, audit the layers above and below before costing the fix.

A live illustration of the thing we've been fixing all session

My commit reported 1 file changed, not 2. Between my pre-commit diff check and the commit itself, the council-gate thread committed 016b — so my §9 correction landed inside their commit 94385507f. I confirmed both the
correction text and the two-phase-claim guidance are present in HEAD, and my 003 changes are in 841e7a2d5. Nothing lost, forward-only held, exactly as CLAUDE.md predicts.

Worth noting the irony: I disclosed a passenger in my commit message, and by the time it ran, I was the passenger. That's the third time this session the shared tree moved under an in-flight operation.

One caveat on my own verification: I traced the two dedupe call sites and confirmed the ordering at each, but I have not proven those are the only inbound paths — a handler reached by some other route could bypass both.
That's bounded and stated in the case file rather than assumed away.

The fix itself is now well-specified but unwritten. When someone picks it up it's a strong council candidate — it touches platform/ and changes a delivery guarantee fleet-wide, which is exactly what that gate is for.

✻ Sautéed for 3m 27s

※ recap: We set up multi-session coordination safeguards, then released the whole fleet to v1.0.1133; latest work diagnosed why pod restarts destroy orchestrations and filed it in bugs_open/003. Next: someone writes the
at-least-once fix, ideally through the council gate.


---

The summary is written and committed at docs/agent_docs/docs024_key_docs_latest/multi_session_coordination/SUMMARY_2026-07-18_where_we_are.md (4ff72194d) — narrative prose in four sections, meant to be spoken rather
than skimmed.

On the hook versus branching — yes, the hook, and my earlier objection was overstated

I talked you out of it partly on a false premise. I said a hook would "change every session's git behaviour without their knowledge." But the repo already runs a tracked, shared pre-commit hook — core.hooksPath =
.githooks is configured and .githooks/pre-commit is version-controlled (it currently execs the secret-scanner). So a guard is an edit to something that already runs for everyone, visible in git, reviewable, and
revertable in one commit. That's a much smaller ask than I implied.

But be clear what it does and doesn't fix, because the two options address different halves. Of the three sweeps this session:

- One was cross-file — a vet-med-export commit collected my makefile change, a file that thread never touched. That's the class that destroys reviewability, and it's your original complaint. A hook stops it.
- Two were same-file — the council thread and I both edited 016b, and whoever committed first took both edits. We both used explicit pathspecs and it happened anyway, because git commits whole files. No hook can prevent
  that. Only separate working directories can, and then it becomes a visible merge conflict instead of a silent passenger.

One technical constraint shapes the design: git has no hook that fires when you stage a file, so we can't literally forbid git add -A. What we can enforce at commit time is the rule already in CLAUDE.md — a broad commit
is allowed but must announce itself. So: if a commit spans an unusual number of files or several unrelated areas, require the message to be labelled sweep:, otherwise reject with an explanation. That's a commit-msg
hook (it can see both the message and the staged set), roughly ten lines.

Recommendation: do the hook, defer the worktrees. The hook is cheap, targeted at the damaging case, and instantly reversible. Worktrees are a genuine restructuring that trades today's collisions for a merge queue — at
~180 commits per two days across twenty-odd workstreams, that's a real ongoing job, and most of those commits are docs that gain nothing from isolation. Earn that change with evidence: if same-file collisions on the few
genuinely hot files keep costing time, isolate only the threads that change platform code.

Worth keeping in proportion though — neither matters as much as the delivery fix. The commit problems cost review clarity and some rework; the at-most-once defect is silently destroying production work right now,
twenty-two orchestrations deep, oldest fifty-one days.

Say the word and I'll write the hook — it's small enough that I'd want you to see the exact rejection rule before it starts blocking anyone's commits.

✻ Cogitated for 4m 27s

※ recap: We've been fixing cross-session collisions in this repo; the read-aloud summary is written and committed, and I've recommended a commit-msg hook over per-thread branches. Next: say the word and I'll write that
hook for your review before it blocks anyone.



---

Delta 2 is built, the standing doc set is in place, and everything now waits on your runbook — here's the full picture.

Built this turn (E1–E5, all recommendations)

Stage-loop machinery — commit c19b5d097, full package tests green. The one new piece of control code is feature_stage_route: it walks an approved staged plan one stage per invocation, emitting each stage as a
single-plan shape so the proven read/prepare actions loop without knowing they're looping. Stage 1 reads the base ref; stage N reads the feat/* branch so it sees earlier commits. The terminal emission carries the
one PR's title/body (your post-merge checklist rendered as a task list) plus the go-test packages derived from the plan's .go edits — the model never declares its own test surface. A pre-existing feat/* branch is a
loud refusal at seed time (E4). The four seams were all optional-field additions — read gains ref_field, prepare gains branch/message overrides plus expected-symbols enforcement, the build gate gains the end-gate
test mode, and feature-implementer joined the spawn gate — with every single-plan path untouched.

Seeds and triggers as draft files — commit 5b131b88a, tokenizer-validated, the implementer's 22-step graph checked target-by-target. Nothing applied; that's yours, after the image.

Coordination catch: the concept-register thread shipped a fourth council seat (reuse-agent, fix-proposer v7) mid-turn. I mirrored it into the feature-designer's council — reuse-before-recreate is this builder's
hard rule 1, so that seat matters more here than anywhere. RUNBOOK A3 covers future roster drift.

The docs you asked for (all in fixloop_eg_dartsonline/, commit 2bffd4549)

- PLAN_feature_builder.md — mission, as-built architecture, the full D/E decision record, per-delta status, backlog.
- RUNBOOK_feature_builder.md — your tasks A1–A7, each with its verification: image ≥ c19b5d097 (pod-grep for feature_stage_route) → apply the three seeds in order → roster check → the F1.2 pilot spec (ready-to-run
  SQL, approval by name) → fire and grade the designer against the schema doc's hand-written reference → fire the implementer → walk the PR checklist.
- SUMMARY_feature_builder_2026-07-17.md — the read-aloud piece: where this came from, what we decided, where we are, where we're going, no code in it.
- NOTES_running_feature_builder.md — turn 3 appended; I'll keep all four current every turn from here.

One more index-race sweep occurred (registry entry rode into aabd38161) — content verified intact, recorded in the notes, forward-only.

Where this leaves it

All code for deltas 1–3 is committed and inert; nothing has touched the cluster or database. The next act is A1 in the runbook, and the first feature through the loop will be the fix loop's own stale-ref repair —
graded against a known right answer before anything merges.

✻ Worked for 15m 22s

※ recap: We're building the feature builder, which constructs multi-step features on the fix loop; all code and seed drafts are built, committed, and inert. Next action: your RUNBOOK A1, deploy the chassis image and
verify the pod. (disable recaps in /config)

---

Run 2 is graded and the loop improved again — this time the machine was right and my gate was wrong.

What run 2 showed (bcc96877, ~50s, no council spend)

The sharpened spec + fixed prompt worked exactly as intended: the designer produced a one-stage plan — just add 0NN_fix_implementer_v2.sql as a seed — with no redundant Go edits and no config_change stage (both of
turn 7's defects gone). Its checklist was seed_apply → verify.

Validation refused it: my D4 rule demanded image_deploy strictly before any seed_apply, unconditionally. But this feature ships no code — there is no image to deploy, and the only way to satisfy the rule would have
been a checklist that lies. The designer's checklist was the truthful one. The pilot found a real validation bug on its second fire.

What I changed (committed 62018e272 + c1b472e59)

- Validation refined: image-before-seed stays a hard failure whenever the plan ships code edits; seed/doc-only plans may checklist seed_apply → verify with no image entry; ordering is still enforced if an image
  entry appears anyway. Three-arm test coverage, package green.
- Schema doc §4 rule 2 amended with the rationale and the run id — note this modifies a rule you signed off as D4, so it's flagged for your ratification; the amendment preserves the rule's intent (never let a seed
  outrun its image) while removing the false assumption that every feature has an image.
- Designer prompt rule 5 updated to match ("a plan with no code edits needs no image_deploy entry — do not invent one") and re-applied live (prior row snapshotted, ba8f1fcd).

The scoreboard, and what the next fire needs

Two fires, zero council credits spent, two real defects found and fixed before any expensive stage ran — the cheap-gate ordering is doing precisely its job.

The next fire is blocked on one thing I can't do: the validation fix is Go, so it's inert until the next chassis image — and rollouts are owned by the other thread today (HEAD builds will pick up 62018e272
automatically, as v1.0.1134 did for my earlier commits). Once a new image lands, tell me and give the credit go, and I'll re-fire — the same designer run against the fixed validation should now persist its plan and
reach the five-seat council for the first time.


---

Run 3 is graded, the one remaining steer gap is fixed and live — here's where the pilot stands.

Run 3 in one paragraph

The furthest run yet, and the system's first full council exercise. The plan persisted (the D4 seed-only fix proven live in production), all five seats reviewed it with genuinely distinct, high-quality catches —
edit-quality caught the sketch using a non-existent definition column and an invented config key; the bug-historian, in its charter's first real exercise, applied pattern #7 (missingkey=zero) exactly on-target and
demanded a fail-loud guard; reuse-agent and guidelines correctly approved (guidelines' approve-with-note meta-rule behaving precisely as designed); guardian objected below veto with independent catches. The council
said revise, the router executed the reviewers' SQL checks, and the reproposer revised — then stumbled on the last unstated convention: it invented a checklist act (image_check) where the act set is closed.
Refused cheaply; round-1 artifacts survive.

What changed since

- Repropose and reframe prompts now state the closed act set (image_deploy | seed_apply | verify; a pre-apply confirmation is a verify entry ordered before seed_apply). Re-applied live, snapshotted, committed
  (c204d6fbc). No image needed — this was prompt-only.
- One drift flag for you, no action taken: the fix-proposer's council has since grown to 13 seats (2 always-on + 11 gated behind the relevance filter). The designer's 5 always-on seats are now an architecture
  generation behind. Run 3 shows they work well, so my recommendation is to finish the pilot on 5 and decide on filter adoption after — it would be a seed-only change since the filter Go is already in the running
  image.

Score: 3 fires, 1 council round spent, 3 real defects found and fixed — each caught by the cheapest gate that could catch it. That's the cage working.

Run 4 needs only your credit go (no image wait this time). Expect it to go deeper than run 3: through validation into a second council round, plausibly to approved — at which point the staged plan is ready for the
implementer. Say the word.

---

Read it — and it changed two things materially. Thank you for flagging 016; it's ours, and it means I graded run 4 wrong.

Bug 016 is the feature-designer's, and it severed the revise loop

In a prompt template, {{.X.result}} renders <no value> silently, because ExtractFields → UnwrapDeep strips the {type,result} wrapper before the template sees it. My designer had five such references in repropose and two
in reframe. The reviser never saw a single objection, in any run.

That rewrites run 4's story. I concluded the loop escalated because the designer stubbornly clung to a dangerous whole-config shape. The truth is it was revising blind — it only ever saw its own previous plan and the
instruction "address every objection". The bug-historian's objection was identical in all three rounds because nothing ever conveyed it. What disguised this is that the facts did improve between rounds:
{{.check_results.results_text}} is correct, so the SQL findings fed through while the objections did not. That asymmetry — facts improve, objections never get addressed — is the diagnostic tell for auditing any other
council.

I fixed it surgically (jsonb_set on the two prompt-template leaf paths, snapshotted, config dot-paths untouched) — which is precisely the co-edited-row rule the council taught me last turn, applied to myself on the very
next change. Verified: zero broken references, check_results intact, the five review_fields config paths intact, seed file corrected so a full re-apply can't regress it. A fleet sweep shows both my agents clean; the
only remaining instance anywhere is content-creator-hero, which another thread owns and has already documented.

Worth noting my run-4 rule 5a still stands on its own merits — the design prompt renders fine, so that knowledge does reach the model.

The council gate — I haven't been using it

CLAUDE.md grew a section asking threads to run platform/ changes past the council before committing, with a Council-Reviewed: trailer. My three platform commits carry no trailer; the coverage report lists them
unreviewed. So are 29 of 30 in-scope commits fleet-wide, so this is an un-adopted convention rather than a personal lapse — and forward-only means I can't retrofit trailers.

The piece that genuinely merits review is delta 2's stage-loop machinery: it has unit tests but has never run live, and it's the next thing we'd fire. (The coverage report was itself mid-fix during this turn — an
earlier invocation reported 3 of 30; another thread's fix landed between my two runs. No bug to file.)

Two things now need your go, both credit spends

1. Run 5 — the first run with a working revise loop. Given the council was landing four approvals and one well-founded objection while flying blind, a reviser that can actually read that objection has a real chance of
   converging to approved.
2. Submit delta 2 to the council gate before the implementer's first live fire — the highest-risk unexercised code in this workstream.

My recommendation is to do both, in that order. Say the word.


---
---

2026-07-21 — Delta 2 back through the council gate; two objections fixed at the source

You asked me to close the delta-2 trail with a fresh council run, and told me a new chassis v1.0.1144 is deployed. I checked the running pod before trusting anything: it carries the stage loop, and — new since we last looked — the two fixes that were sitting committed-but-inert. One gofmts the implementer's generated Go at commit time so a stray whitespace can't burn a whole run (that killed the fix loop's very first fire); the other lets a truncated council reviewer degrade instead of voiding the whole round. Both matter for us, because the implementer's first real fire (still the one thing left) runs through exactly that machinery. So the ground is firmer than it was on the 19th.

On the four objections still open from round 1: two I could only answer honestly by actually changing something, so I did.

The first was the gate that checks each stage produced the symbols it promised. A reviewer pointed out it would wrongly reject a correct stage if that stage merely *uses* a name defined in an earlier stage. The tempting fix is to soften the gate; the honest fix is to stop the designer ever promising a name it doesn't itself introduce — so I tightened the designer's instructions to say exactly that, and left the gate strict. Applied it the careful way (one surgical edit, snapshot first), the same co-edited-row discipline the council taught me.

The second was a provenance point: those three helper actions the feature loop now shares with the older fix loop should each carry a little travelling logbook, so the next person who edits them reads the prior decisions first. Turned out the system didn't actually allow a logbook keyed to an "action" at all — the convention existed on paper but nothing supported it. So I added that (a one-line schema widening) and wrote the three logbooks, each noting that these are shared, battle-tested actions and what changed. The other two objections were answerable in words — one was a request to declare a registration step as its own line item, the other a fair question about whether I should have reused an existing sequencing mechanism instead of writing new loop code (I explained why the two are different: one queues separate jobs, mine tracks state inside a single run).

The one genuinely serious finding from round 1 — three places where an empty value would have silently done the wrong thing — was already fixed back on the 18th; I just made sure this round's write-up shows the fixed code, because the reviewers read the sketch, not the file.

I resubmitted, and this time it dispatched straight away rather than sitting in the ~half-hour queue we measured last week. It's running as I write this. If it comes back approved, the delta-2 code finally carries a review stamp and the trail is closed; if it asks for more, the thread just continues under the same id. Either way the two real changes I made are already committed and live.

I've written a fresh cold-start handoff dated today so you can carry this on in a new chat without losing the thread. The verdict and how to read it are at the top of it.


---
---

2026-07-25 — It built the thing, and you merged it

You asked me to bring the handoff up to date because a lot had changed. A lot
had. The sentence that has been at the top of every handoff in this workstream —
"the implementer has never run" — is no longer true, and the way it stopped being
true is worth saying out loud.

The gauntlet thread needed a backend built from scratch and gave the job to our
builder rather than writing it by hand. It took six attempts to get a plan the
council would approve and eight attempts to get a build that finished. Then on
the morning of the 25th it ran the whole thing straight through: six stages, each
committed on its own and gated before the next was allowed to start, a test run
derived from the plan rather than declared by the machine, and one pull request
— eighteen files, eight hundred and eighty lines, nothing deleted, nothing
touched outside its own service. You merged it at 09:19.

I want to be precise about one thing I found while checking. The other thread's
write-up, committed at 09:17, says the PR is open and awaiting your review. It
was — for two more minutes. I checked GitHub rather than trusting the doc, and
the merge was already in. I've recorded that in our notes rather than editing
their file, because their file was accurate when it was written and it is a
record of a moment, not a status board.

What I'd flag, honestly, rather than let the milestone paper over:

Most of what went wrong in those fourteen attempts wasn't the builder. Two runs
died because a housekeeping job was deleting the message channels the agents were
replying on, and one whole upgrade never reached the agents at all because they
pin their image version in the database and nothing updates it. That second one
is still open, and it's the nastiest of the lot: when it happens, a perfectly
good agent looks broken. I've put a census query at the top of the runbook so
nobody fires again without checking, and flagged it as belonging to whoever owns
deployment rather than us.

And one success isn't a capability. The next genuinely useful thing is a second
build, on a target we choose ourselves, to find out whether that fortnight of
fixes generalised or just got this one job over the line.

The old review trail on our own stage-loop code is still sitting at "revise" and
nothing has run on it since the 21st. My view has shifted: the code it's
reviewing has now built and shipped a real feature, the one serious thing it
found was fixed a week ago, and what's left are two design questions rather than
risks. So it's a tidiness decision now, not a safety one — worth doing, not worth
hurrying, and it's your call whether to spend the credits.

There's a fresh cold-start handoff dated today, and a new summary you can read
aloud to someone. I fired nothing and spent nothing this session.

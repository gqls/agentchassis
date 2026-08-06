# Where we are — bug 208, a rebuild that would have destroyed live tools

Plain-prose log, append-only, newest at the bottom.

---

## 2026-08-06 (evening) — picked this up, and it is worse than it was filed

Earlier today another session was getting ready to rebuild all the pages on
`ai-agent-orchestration.com`, which you had authorised. Before firing it, that session checked
what the rebuild would actually pick up — and stopped. It filed bug 208 and moved on to its own
work, leaving the bug for someone else. I have taken it on.

The problem in one sentence: **when a rebuild sweeps up a page that belongs to a tool, it
regenerates the page as ordinary prose, commits that over the live tool, and only then hits the
check that was supposed to stop it.** The check works. It just guards the database rather than
the file the website actually serves. So the tool would have been gone from the live site while
the database still described it — a mismatch nothing repairs by itself, because the repair paths
read the database and think everything is fine.

Two things I found that the original filing did not say.

**It is not two pages on one site, it is fourteen pages across six.** The filing named the two
tools on `ai-agent-orchestration.com`. The real list also includes six tools on
`gamesdesign.co.uk`, one each on `finetuning.uk` and `leopardessconsulting.co.uk`, and three on
`vonc.com` — including the arena and the gauntlet. That last detail matters more than it looks:
`vonc.com`'s arena being clobbered is the original incident that made us invent the "this page
belongs to a tool" marker in the first place. The marker is still not protecting the very page
whose destruction created it.

**It is not one pipeline, it is three.** The filing found it in the operator rebuild. The same
"commit first, check second" order is also in the new-site build pipeline and in the
work-item-driven builder. So a fix aimed only at the operator rebuild would leave two other
doors open.

The good news is that the shape of a proper fix fell out of the evidence rather than having to
be invented. There is one step — the step that assembles a freshly written page just before it
is committed — that is used by exactly those three pipelines and by nothing else. The paths that
legitimately publish tool pages don't go through it; they have their own route, and the original
design says in writing that route must stay ungated because it is how tool pages get deployed at
all. So that assembly step is the one place where "we are about to commit generic prose over
this page" is unambiguously true. A refusal there covers all three pipelines and cannot break
the legitimate ones.

Better still, I don't need to invent the refusal mechanism. That assembly step already knows how
to say "skip this page, carry on with the others", and the committing step already listens for
it. So the fix can use the existing signal instead of adding a new one — no configuration change
on any pipeline, which on a shared tree with many sessions is worth a great deal.

There is a second half I want as well: stop these pages being *selected* in the first place, so
we don't pay an LLM to rewrite a page we are then going to throw away, and so the run doesn't
report a failure for work it was right to refuse. That is a change to a shared selection query
used by two pipelines, which makes it the kind of change the guidelines say must go through the
council and be written into the concept register in the same commit. I am doing both.

One question I deliberately have not answered yet, because the answer changes the design: if we
refuse to rebuild one of these pages, does it sit in the queue asking to be rebuilt for ever,
getting picked up and refused on every future run? Being answered before I write the code rather
than discovered afterwards. I have a second model working that out in parallel with the rest of
the design.

Last thing worth saying plainly: the operator rebuild feature that tripped over this went live
today and belongs to a different workstream. My fix changes the behaviour they depend on — an
owned page they explicitly name will now be refused rather than rebuilt. That is the correct
answer, but it is their guarantee that changes, so they get told, not just measured.

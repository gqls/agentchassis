# Where we are — vigilant designer + offer analyser

Plain-prose running log. Append-only, newest at the bottom. The owner writes here too — never
edit or reorder existing entries; add dated corrections below instead.

---

## 2026-08-02 — the programme exists, and why it looks the way it does

You asked for three things: a constant vigilant designer working from whole-site down to the
finest detail, and an offer analyser (with benefit analysis inside it) that keeps asking whether
each site actually answers its target market's need in a way that pays us.

The surprising finding from digging: you already asked for about half of this on 24 July, and
the platform half-built it. There are written specifications for a screenshot-based design
critic, a "did we build what the brief asked" audit, and checks for components nobody uses —
none of them built. And the one that DID run (the brief-fidelity audit) predicted your
fundamentallyai complaints three days before you made them — but its findings sat in a queue
nothing reads, and you found the problems with your own eyes instead. That is the real lesson
this programme is built around: the platform is reasonably good at noticing things; it is bad
at doing anything about what it notices. So every detector we build ships together with the
thing that acts on it, and we prove each one by watching a real finding travel all the way to
a changed page.

On the design side: the reason all the sites look the same is now measured, not a feeling. The
page planner is only offered a brochure-shaped menu; the design AI is only allowed to change
colours (everything else is owned by shared layout templates); three shared header components
serve all fourteen sites; and five interactive components were built that no planner has ever
picked. Each of those gets its own fix.

On the offer side: each site already has a written monetisation analysis and audience
definition from build time — and nothing ever reads them again. The agent that reviews sites
doesn't load them. So the first offer work is simply wiring the premise back into the judgement,
then adding a check for the obvious mismatch ("a tools site with a consultancy CTA"), then a
proper offer analysis agent. On the live estate the recorded revenue model is mostly "none" or
"implied", so there is also a premise-refresh path to fill that in — carefully gated, because
the strategy agent currently kicks off a full rebuild chain whenever it runs.

Your four calls today: manual per-site triggers for now (the always-on switch stays off until
you say otherwise); designer first; trial both Gemini and Claude as the critic's eyes; and
broad autonomy — the designer may apply changes at every level, with two honest exceptions
(shared header/footer components get forked per-site before editing, because editing the shared
one changes every site at once; and whole-site layout swaps stay as review items because the
platform deliberately has no automatic path for those yet).

Next up: the plumbing phase — make findings flow at all, fix the counter that falsely reports
audited-out sites as "clean", and turn the browser audit's measurements into actionable work.

## 2026-08-02 (later) — the plumbing is in, and the platform now argues back usefully

Since the morning entry: the two live-config changes are applied — the sweep points at the
right lane (still switched off, that flip stays yours), and the dishonest three-audits-
and-report-clean counter is replaced by an honest gate: audit when something changed or
enough time passed, always move findings toward workers, and when three audits at an
unchanged site produce fixes that never land, say so on the roadmap instead of saying
"clean". Three code pieces are written, tested and committed but dormant until we ship a
new image: the piece that turns browser measurements into actionable work, whole-site
screenshot capture (both phone and desktop views), and the "eyes" — the ability to send
those screenshots to an AI model, built for both Gemini and Claude so your trial is a
switch, not a rebuild.

The review council looked at the screenshot work and approved it first time, with a
genuinely good catch: our capture failures were logged but not counted, so the future
design critic could have been quietly judging half a site without anyone knowing. That
counter now exists. Two more reviews are still in the queue.

Next session: ship the new images, prove one finding travels all the way from "a browser
measured it" to "the page changed", and then build the critic itself and run your
Gemini-versus-Claude comparison.

## 2026-08-03 — the two overnight objections answered, and both taught us something

The review council had sent back both remaining code pieces with one blocking objection
each. Both are now answered and resubmitted, and each objection turned out to be valuable
for a different reason than the reviewer intended.

On the findings-writer: the reviewer feared repeat findings would silently vanish after
two fix attempts. They don't — the third one is recorded and labelled "unresolved after
2 attempts" — but chasing the claim exposed two things we actually had wrong. First, our
own notes said these parked items show up in the fix loop's daily digest; they don't —
they show up on the admin dashboard's attention lists, and the notes now say so (logged
as a wrong call). Second, and bigger: the agent that's supposed to fix contrast problems
has never once been given a work item in the platform's whole history, and the prompt it
runs expects fields our findings didn't include — so our findings would have arrived as
blank forms. The revision writes the fields its prompt actually reads, plus the page
reference, so the first item that agent ever receives will be one it can act on.

On the "eyes": the reviewer said we claimed truncated AI responses fail loudly but never
built the mechanism. The mechanism has existed since 20 July — the reviewer was reading
our own concept register, which still described the bug as unfixed two weeks after it was
fixed. So the code was right and the paperwork was wrong, and wrong paperwork cost us a
review round. The register entry is corrected, a trap note now warns future readers (and
future reviewers) that register status lines go stale, and two new tests prove the
truncation alarm fires through the new image path specifically, for both AI providers.

Both revisions are committed and back in the review queue. Next: read the two verdicts,
then ship the new images and prove one finding travels browser → work item → changed page.

## 2026-08-04 — the drain works: we watched a finding travel, and the counter bug is closed

This morning the whole point of Phase 0 got proven on a real site. We hand-fired one
improvement sweep at relojistas.com (the watch-news site rebuilt at the end of July).
Before firing, we checked its queue by hand: five of its six old findings described the
site as it was before the rebuild, so we cancelled them with the evidence written on each
— firing blind would have churned live pages over stale complaints.

What happened next is exactly what we built: the new gate looked at the site, said "never
audited — audit due", ran the full audit, promoted twenty-two findings to workers, and
the workers started clearing them. Two completions are worth naming. The old "empty
news section" finding from 19 July — which had sat dead in the queue for over two weeks —
was closed by the system itself observing the section is now healthy, with the reason
recorded. And a fresh finding ("the shared header is stale") was picked up, the header
re-rendered, and a nineteen-page refresh queued and draining behind it. One finding
resolved by honest observation, one by actual work: both ends of the drain shown working
in a single run.

That witnessed run was the last thing the audit-counter bug (the one where a capped site
was reported "clean" forever) was waiting on — it's now formally closed.

One correction to our own plan: the sweep does not include the new browser-measurement
audit; that runs as its own job. Proving that half — a browser measurement becoming a
work item — is the next small step, after the page refresh settles. Then the critic.

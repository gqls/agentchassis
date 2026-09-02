Everything's captured and corrected. Here's the complete status.

Punch-list: 5 fixed, rest documented and actionable

Fixed and live this turn:
- Nav colour consistency — gold header + charcoal footer across all pages (27/30; the 3 exceptions are empty pages). The stray navy hexes are dead CSS fallbacks that never render.
- Nav decluttered to a clean 9 items; blank "For Leaders" gone.
- Card 404 links — template gated, phantom /how-we-work/* links stripped, use-cases quiz link repointed.
- About invented stats — made honest and true (30 years / 8 sites / 2,767 records).
- favicon.png — was 404; the head hardcodes /assets/images/favicon.png and we'd only committed the .ico. Committed it; now 200.

Two important corrections the research surfaced

- The A6 Banana routing is actually deployed (v1.0.1114) — my earlier "not deployed" note was stale. So images are not blocked by any deploy. The real reason leopardess has no images is that it has no
  site_plan/site_plan_imagery rows at all. The handoff now has a §9 playbook: Route A (per-image trigger — safe, immediate, how robot-hands got its heroes) is the way in; Route B (a full re-plan) is flagged
  as risky because it could overwrite the copy we carefully fixed.
- The blog is less broken than it looked — it already renders 5 posts with working links. The empty look is blank card images (a structural Phase-I3 gap, not built) plus empty excerpts (a quick fix:
  populate meta_description).

The one honest limit worth stating plainly: per-card and per-section images (the blog thumbnails, the info-card images) need "Phase I3", which doesn't exist yet — that's real new build work, not a config
toggle. Hero images per page, though, can be done today via Route A.

Your deliverables

- HANDOFF.md opens with the punch-list table (issue → status → root cause → fix) and now has the full imagery playbook (§9). It's the single doc to open a fresh chat from.
- RUNNING_NOTES.md — turns 15 and 16 record the whole review and every fix.
- specs/VOICE_REWRITE_PROMPT.md — your "think hard about a prompt" ask: a specific, opinionated prompt naming the exact LLM tells (the balanced triads, the "not-X-but-Y" strawmen you can see in that
  how-we-work hero, the summarising flourishes) with a worked before/after.

Everything is backed up. The site's structure and chrome are now genuinely solid; the substantial remaining work is content depth — applying that voice rewrite, hero imagery via Route A, the blog
excerpts, and the "we could but don't" reframe on use-cases. All of it is now root-caused and actionable in the handoff, so you can pick up any piece in a new chat.
--


Done for this session. Here's where things stand.

What's fixed and verified live

The use-cases page was the real integrity problem, and it's now honest. It was worse than the punch-list said: not just "claims we do things we don't", but five fully fabricated case studies with invented
clients ("Revenue Operations at a Growth-Stage SaaS Company") and invented results ("latency drops from days to minutes"). It also carried two more phantom quiz links that turn 15 missed, and a navy
gradient that was genuinely rendering (not a dead CSS fallback). The fix used your own rewritten portfolio spec as the source of truth — it already held five honest "Not yet done for a client" use cases —
so the page now renders from the spec rather than from LLM prose. Verified live: 0 fabrications, 0 phantom links, 5 honest labels, 0 navy.

Blog, footer, and an empty page. Blog excerpts and read times are live (5/5 cards). The dead "For Leaders" footer link is gone from all 17 pages. llm-cost-calculator-guide was a blank shell linked from
the blog listing — nobody had noticed it — and is now a real 27KB page.

The finding that matters most

Route A, as written in HANDOFF §9, would have destroyed your copy. With scope:"page", image-build-handler emits a needs_page item that routes to page-build-handler, which has no skip-content branch — it
always re-runs the content writer, and every source:"llm" field is regenerated unconditionally. It was safe on robot-hands only because that site's copy was LLM-written anyway. I proved a safe variant
instead (scope-less spec → asset generated and git-deployed, zero content touched) and rewrote §9 around it.

Two further imagery blockers, both documented: kind:"hero" routes to SDXL, not Banana — so a leopardess hero can never be on-brand, because your house style is flat illustration. And the cluster keeps
stalling spawned image pods on Kafka dial timeouts. The one hero that did generate came out as a beige faux-diagram with garbled fake text, so I deliberately did not wire it — nothing is wired, so nothing
bad is showing.

Two things I want to flag honestly

The voice rewrite didn't get started. That was your "think hard about a prompt" ask and it's the biggest remaining content job. The session went into root-causing instead. contact is the right place to
begin — it's the one page with empty content_data and CTO-register copy, so there's nothing to lose.

ai-readiness-quiz is still blank. I found and fixed the root cause fleet-wide (the shared contact-block schema's placeholder jane@company.com was being read by the validator as a hallucinated contact
email, failing every build of any page using that component — likely the cause of yesterday's 4 failed backfills too). The build now gets past validation cleanly, but the orchestration stalls on the same
cluster flake. It needs a re-fire when the cluster is healthy. It's linked from the footer and the use-cases CTA, so it shouldn't stay blank.

Both handoff docs are updated with the corrected punch-list, the clobber warning, the queue-bypass lesson (reassemble_pages.sh works when the work queue won't drain), and the two platform bugs.

---

Recorded. I'm stopping the quiz re-fires — five attempts, all killed by the same lost-child-response infra flake (take-5's error was literally reaper: stale AWAITING_RESPONSES for >90 min), never by
anything in the content or validation path.

Final session status

Fixed and verified live:
- use-cases page — five fabricated case studies replaced with honest, spec-driven "could build" patterns; 0 fabrications, 0 phantom links, 0 rendered navy
- Blog — 5 cards with real excerpts and read times
- Footer — dead "For Leaders" link gone from all 17 pages
- llm-cost-calculator-guide — rebuilt from a blank shell to a real 27KB page
- contact-block validator bug — fixed fleet-wide (was making every page using that component unbuildable)

Documented, not shipped (correctly):
- The Route A imagery clobber trap, with a proven safe variant and the hero→SDXL routing gap — no image wired, because the one that generated was unusable
- The queue-bypass lesson (reassemble_pages.sh when the work queue won't drain)

Left open for next session, cleanly handed off:
- Voice rewrite — the biggest remaining content job, not started; contact is the place to begin
- ai-readiness-quiz — content path is clean; blocked only on the cluster's lost-child-response flake. Re-fire when the cluster is healthy enough to complete a spawn→child→response cycle, ideally watching
  it land rather than firing blind.

Both handoff docs are current. Nothing is in a half-broken state on the live site — every change either landed and verified, or was deliberately not wired.

---

All nine findings are resolved and verified live. Summary of what just happened:

The rulings, applied

The blog guide (the page that leaked the fabrication after the sweep — the strongest evidence the claims layer was needed) got four surgical replacements: the "70+ agents across eight functional departments" claim
became the DB-verified "registry holds over 150 agent definitions," and the three follow-on sentences that leaned on the invented "department" taxonomy were reworded — while keeping the genuinely true topology
(head orchestrator → supervisors → narrow workers), which is audit-verified. The falsehood went; the substance stayed.

technical-architecture — the word-form variant ("more than seventy agents operating in eight functional areas") that nearly slipped the regex is now "more than 150 agent definitions — coordinators, specialists,
analysts."

insights — clean, though with a footnote: the claims thread fixed that one directly at 17:29, after my edit, with their own wording. Same goal, no conflict in the result, but I've flagged the coordination point in
the notes: two threads are now editing leopardess content_data, and future rulings should flow one way — the claims thread rules, this workstream applies, as the spec designed.

for-engineering-teams (archived, twice-flagged) — fixed in content_data only, not redeployed, so the sleeping copy is safe if the page is ever resurrected.

Process notes worth keeping

- The dual-branch escalation pre-check (the guard I corrected earlier today after the who-we-help clobber) was run on all three live pages before firing any re-render — zero rows, and indeed zero escalations this
  time. The guard works.
- Verification greps must strip tags — served HTML splits phrases across <strong> boundaries, which makes literal greps blind in both directions. The notes record the tag-stripped method.

Where leopardess stands now

Punch-list: items 1–4 and 6–10 closed; item 5 (imagery) has two on-brand per-page heroes live plus the hand-chosen index hero, with the remainder being component work and Phase I3 (owned by the imagery workstream).
Served HTML across the site now carries zero banned claims, checked against the evidence base. The loop the spec designed — check finds → human rules → operator applies in content_data, nothing auto-rewritten —
has now run end-to-end for real, once.


----

That's the punch-list run to ground. Here's where this leaves leopardess:

Completed this stretch

Sitemap — live. The site had none at all (nor does any sibling site — there's no generator anywhere in the platform, which I've logged as a fleet-wide gap worth a generate_sitemap deploy-time action). Generated
from the pages table — 27 URLs, archived pages verified excluded — and committed through the git-adapter. Verified 200.

robots.txt — checked and fine. The stanza layout looked alarming at first glance, but it's Cloudflare's managed content-signals file: search engines allowed, AI-training crawlers (GPTBot, ClaudeBot, CCBot…)
blocked. No action needed.

about/services heroes — properly deferred, with evidence. hero-about is shared across 9 sites, hero-services across 5. Adding an image field is a shared-schema change that belongs to the imagery workstream (which
is mid-build on Phase I3 anyway), not a site session. The fleet-safe pattern for whoever does it — an optional field behind an {{if}} gate, same class as the turn-15 link fix — is written in the notes.

The full picture

Every item on the owner's original punch-list is now closed or accounted for:

┌──────────────────────────────────────────────────────────────────┬────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                                                  │                                                 State                                                  │
├──────────────────────────────────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ Nav/footer colours & clutter, card 404s, invented stats, favicon │ Fixed (turns 15–17)                                                                                    │
├──────────────────────────────────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ use-cases honesty                                                │ Fixed — spec-driven, zero fabrications                                                                 │
├──────────────────────────────────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ Blog                                                             │ Excerpts + read times live; thumbnails arrive with Phase I3                                            │
├──────────────────────────────────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ Empty pages                                                      │ All resolved — quiz live (54KB, interactive), guide rebuilt, duplicates archived by merge decision A11 │
├──────────────────────────────────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ Voice                                                            │ Done everywhere it was needed; four titles and 12 meta descriptions honest                             │
├──────────────────────────────────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ Imagery                                                          │ Three on-brand heroes live; the rest is component work + Phase I3, owned elsewhere                     │
├──────────────────────────────────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ Claims                                                           │ All 9 findings ruled and fixed; served HTML site-wide carries zero banned claims                       │
└──────────────────────────────────────────────────────────────────┴────────────────────────────────────────────────────────────────────────────────────────────────────────┘

What remains is genuinely the build-out — tools, illustrated guides, the news surface, the L7 chart component — which is new construction, not repair. The handoff reflects all of this, so the next session (here or
a fresh chat) starts from a site that is structurally sound, honest end to end, and verified so.

One small observation to leave you with: this session ran the claims-verification loop end-to-end for the first time — scan found, human ruled, operator applied, nothing auto-rewritten — and it worked exactly as
the spec drew it. The two threads brushed against each other once (both edited insights); worth keeping the "claims thread rules, site thread applies" division going forward.

---

Monitor event: "homepage v2 deploy verification relay"

● Monitor "homepage v2 deploy verification relay" stream ended

● The homepage is live in the new register, fully verified. Summary of where the style change stands:

What's now in place

The style is encoded in three layers:
1. specs/PLAIN_VOICE_v2.md — the distilled rules: one idea per sentence (~20 words max), contractions in, short paragraphs, Flesch ~80 target, no literary moves, em-dashes near zero, friendly-means-calm-not-chummy.
   It also records what I deliberately rejected from your reference material: deliberate errors/slang, casual fillers ("You know what?", "honestly"), rhetorical questions as a technique, and AI-detector-evasion as a
   goal — those would hurt a consulting site. All of v1's honesty rules survive untouched.
2. The DB voice spec — tone, plain_language, a new sentence_shape key, and rewritten LLM-tells. Anything the content writer generates from now on reads these rules.
3. The homepage, live — as the worked example, pushed further than your sample:

▎ We build systems that take over repetitive work. Each one has a clear job. It knows when to ask a person for help, and it writes down every decision it makes. When it isn't sure, it stops and asks. Nothing
▎ happens in a black box.

Verified live: contractions throughout ("we've built" ×5, "isn't" ×3, "You're reading one of them"), zero em-dashes, the dense v1 copy gone, and all five audited numbers intact.

One claims catch from your sample

Your illustrative rewrite said "reads news from hundreds of sources" — the evidence base has 18 configured sources. I kept the source types and dropped the count, and noted it in the style guide. The register
changes; the facts don't. (This is exactly the drift the claims layer exists to catch — style rewrites are where numbers quietly inflate.)

Next

The rest of the site — services, how-it-works, use-cases, who-we-help, contact, about, engagement-model, faq, technical-architecture, and the two blog posts — still carries the v1-dense register. Before I convert
~12 pages, worth a minute of your eyes on the live homepage: is this the calibration you want, or should it go further still? Say the word and I'll roll it site-wide exactly at this level.

✻ Brewed for 6m 36s

※ recap: Leopardess rebuild: we moved the site to the new plain, friendly copy style; the homepage is live in that register as the worked example. Next: your eyeball on the homepage calibration, then I roll the style
across the remaining pages.

1 tasks (0 done, 1 open)
◻ Roll out plain-voice v2 (style guide + spec + homepage first)

───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────── leopardess2 ──
❯ ok please roll out the copy changes but it needs to be from the framework - the content writer will need to have written them so we need to update the prompt for this site. I don't want it written here "manually"
Please paste the prompt here so I can look at it first. Then the site itself has many errors. There are no images except for the logo, the tools aren't working and several other problems that we can look at.
Please run an auditor to see what problems it picks up. For the tool fix see if the runningdocs tool is able to detect and fix it. Please also look at what checking agents we have that check the voice of the copy
like we have been looking at, to flag sites that sound too much like AI llms. Nice catch with the "reads news from hundreds of sources" thank you.

----
1. The rewrite prompt — for your review before anything fires

Mechanism first: the writer receives per-run guidance via input_data.spec.suggestion, which lands in its prompt as "## Rewrite Guidance (IMPORTANT — incorporate this into the content)". Combined with the v2 voice
spec (already in site_specs, which the writer always reads), this is the framework path — I fire page-build-handler per page with this in the spec, and the content writer writes the copy, gated by
validate_page_content including the now-deployed banned-claims blocker. Nothing hand-written.

The prompt I propose to send, verbatim:

▎ Rewrite this page's copy in the site's plain register. This is a rewrite for readability: keep every fact, number, and claim exactly as given in the specs and section briefs. Do not add new facts, numbers,
▎ clients, or capabilities. If you do not have a number for something, describe it without one. Never round a number up or dramatise it.
▎
▎ How to write:
▎ - One idea per sentence. If a sentence carries two ideas, split it. Keep most sentences under 20 words.
▎ - Short paragraphs, one to three sentences.
▎ - Use contractions: it's, we'd, you're, isn't.
▎ - Use everyday words: use, not utilise; help, not facilitate.
▎ - Active voice. Talk to the reader as "you"; call us "we". Start a sentence with And or But when it's natural.
▎ - No em-dashes. Use a full stop or a comma instead.
▎ - No hype words and no marketing register: never unlock, leverage, seamless, transform, cutting-edge, game-changing.
▎ - No literary flourishes. If a phrase sounds quotable, simplify it until it just sounds clear. Never end a section with a summing-up line.
▎ - Don't write lists of three by reflex. Two examples are fine.
▎ - No rhetorical questions. No forced friendliness like "You know what?" or "honestly". Friendly here means calm and easy to read.
▎ - If we haven't done a thing for a client, say so plainly: "We haven't done this for a client yet. The nearest thing we've built is X."
▎
▎ The test: read it aloud. It should sound like a person explaining their work to a smart friend, plainly, without performing. If a sentence sounds impressive, rewrite it until it sounds clear instead.

One honest caveat: the writer regenerates from specs — it doesn't see the current page text. Substance that lives only in hand-curated content_data (like how-it-works' "What it does not do" block) will be
regenerated from the specs, not preserved verbatim. The specs are honest so the output should be too, but I'll review each page against the claims gate before it deploys. Approve the prompt (or amend it) and I'll
roll page by page.

2. The audit — ran all three discovery agents. Big findings

The headline finding is an incident, not a list item: the content-gap-planner responded to a "content gap" by proposing a fully fabricated case study — invented financial-services client, invented metrics to the
decimal ("4 hours to 11 minutes", "0.4% error rate", "73% queue reduction"), explicitly purposed as "proof-of-capability for prospects in regulated industries" — and it had already created the page row. I held all
three of its work items for your review and archived the page row. Nothing was built or published. This is the strongest specimen yet for the claims thread: the fabrication pathway isn't just the writer, it's the
planner.

---

What exists now

┌─────────────────────────────────────────────────────┬───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                        Piece                        │                                                                           State                                                                           │
├─────────────────────────────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ Engine —                                            │ 7 deterministic signals: banned phrases (13 global AI-tells + per-site bans), strawman shapes (both "not X, but Y" and the staccato "Not a demo. Not a    │
│ platform/orchestration/datahelpers/voicetells.go    │ proof of concept."), em-dash density, triad density, sentence-length distribution, missing-contractions, flourish endings. Reuses the claims layer's text │
│                                                     │  extractor, so the tag-split problem stays solved once.                                                                                                   │
├─────────────────────────────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ Corpus tests — V1–V7 from the spec                  │ All green first run — every TRIP case is real copy that shipped on the site; every PASS case is the approved v2 register. V6 guards the rule that slang   │
│                                                     │ and errors are never rewarded.                                                                                                                            │
├─────────────────────────────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ Discovery check — check_voice_tells.go              │ Registered as voice_tells. Opt-in by the site having an enabled voice_gate; relaxed thresholds for blog/guides; one work item per page, always medium     │
│                                                     │ severity, priority behind claims (truth outranks register), HITL-terminal — no automated handler in this phase.                                           │
├─────────────────────────────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ Operator CLI — cmd/voicescan                        │ Same TSV contract as claimscan; exit 1 on findings, so it's a scriptable acceptance gate.                                                                 │
├─────────────────────────────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ Config — leopardess voice_gate seeded               │ 10 site bans curated from your banned_language into machine-safe regex (the prose entries carry human qualifiers regex can't honour, so they were         │
│                                                     │ hand-translated, not auto-parsed).                                                                                                                        │
└─────────────────────────────────────────────────────┴───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘

The first live scan: 111 findings across 85 components — and the calibration is right

Three things prove it's measuring the real thing:

1. The v2 pages are clean. Index and the restored services produced zero findings. The register you approved passes its own gate.
2. The v1-dense pages light up correctly. how-we-work is the worst offender (a strawman, 7 triads, and heavy em-dash rhythm in one block) — and it's precisely the page we haven't rewritten yet. Em-dash density is
   the dominant signal fleet-wide (up to 41 per 1,000 words on privacy), which is the v1 register's signature exactly as the spec predicted.
3. It has no author bias. It flagged my own hand-written copy from the 16th (the who-we-help cards and FAQ) as contraction-free stiff register — a fair cop, since that predates your v2 ruling. And it flagged
   today's writer-produced about bio for dense sentences, matching my human read.

The useful corollary: this scan is the v1→v2 rollout worklist. The flagged pages are exactly the remaining rewrite targets, now with machine-readable evidence per page instead of my judgment.

What remains (deploy-gated, not code)

The code is built, vetted, and tested locally, but ships with the next chassis image — after it lands, one line adds voice_tells to quality-discovery-agent's checks array (adding it before the image would hit an
unknown check). The files aren't committed to git yet — say the word and I'll commit them. T3 (the LLM prose lane, riding with claims V3) and T4 (the auto-rewrite handler) remain future phases per the spec.

One loose end from earlier still pending: the fixloop diagnosis of the tools timed out its f Jump to bottom (ctrl+End) ↓ e its verdict next unless you'd rather steer elsewhere.

---

Your review turned up two genuinely fundamental problems, both now filed in bugs_open/, plus several things I fixed today.

Fixed and live now

The homepage hero. You were right about what it was — I pulled the image and it's an SDXL-generated fake flowchart: callout boxes, connector lines, a caption strip, and every single word gibberish. Replaced with a
text-free Banana illustration (four gold inputs converging into one steady output on charcoal, upper-left kept clear for the headline), wired so a rebuild can't drop it.

The tools nav. Confirmed — the header had no tools; a rebuild had stripped them, and a tools nav group existed that renders in neither header nor footer. Four working tools now linked from the footer. Correction to
my earlier tool audit: process-automation-scorer is real and works (self-contained 5KB calculator), so four tools work and only llm-cost-calculator is broken.

The overused words. 12 "trust", 9 "honest", 2 "earns its keep" live. Added to the voice gate and prose guidance, and the checker then produced the exact worklist — including your cited pair. Rewrote the homepage
instances: "can't just be trusted" → "will not match the register on its own"; "how much the source can be trusted" → "how reliable that source has been"; both "the honest answer" cut.

Answering your imagery question directly — bugs_open/011

Better model, better descriptions, or loop until correct? None of them solve it. Diffusion models synthesise glyph-shaped texture, not text — it's a property of the model class. A better model helps marginally;
better prompts are why the two good heroes are clean but can't make glyphs legible; and a loop can only ever reject a bad image, never make a model render text it structurally cannot render.

The fix is to split the two jobs: heroes are generated, text-free; anything with words, numbers or structure is rendered in code as SVG from real values. That's already your own stated principle (D1/D3: code
renders data, the LLM never touches values) and the planned L7 chart component — it was just never built, while infographic stayed a diffusion target, which is the trap. That's also the honest route to the
infographics you want: real legible text, real numbers from the evidence base.

The blank page was a symptom of something worse — bugs_open/001

The "Monitoring Coverage Gap Finder" link is a 404 because the re-plan invented it when it clobbered services at 07:50 today. That bug already existed from idea.uk; I appended fresh evidence that raises its
severity, because leopardess shows it does something worse than lose content: it injects fabrications. Yesterday it rebuilt the homepage and re-added "Functional Areas: 150+" plus invented case-study titles; today
it invented a dead tool link you clicked.

This is the thing worth fixing first. Two clobbers in 24 hours while I was actively repairing the site means content fixes currently have an undefined shelf life. One useful discovery: heroes wired through
site_plan_imagery survive the clobber while copy doesn't — which is why the imagery work is durable and the word pass is worth deferring until 001 is fixed, rather than doing it twice.

The full plan is in docs/leopardessconsulting/PLAN_imagery_and_design_2026-07-18.md, including the remaining heroes (about/services need a shared-component change first) and the still-garbled
/assets/images/hero.jpg, which is the site-wide fallback and still live on how-it-works.


---

You were right, and it's better news than my handoff said: the capability was already wired. I've corrected the record and proved it end-to-end.

What the test showed

BANANA_DEFAULT_MODEL is gemini-3-pro-image-preview — the same model that made your images — and kind:"infographic" already routes to it. Nothing needed wiring. I generated one through that lane and it came back
publishable on the first attempt:

- Every label legible and correctly spelled
- Figures exactly right: 2,767 / 937 / 5,652 / 4,672 / 8 sites — all matching the evidence base
- On-brand charcoal and antique gold, three columns, working CTA buttons
- It even respects the wording you just banned — "how reliable the source has been", not "trust"

It's live at /assets/images/infographic-what-we-build.jpg.

I corrected my bug filing

My 011 claimed "generated images cannot render readable text" and proposed building an SVG renderer. That was wrong, and a thread acting on it would have built a subsystem we already had. I rewrote it with the
correction at the top (kept visible, not deleted — it's an easy mistake to repeat) and retitled it to the real, narrower bug: kind:"hero" falls through to SDXL, which genuinely can't render text. That's why the
homepage hero was gibberish while the infographic lane sat unused.

Per CLAUDE.md — which I re-read, and which has grown a lot — I also did the two things I'd skipped: corrected the §10 bug index entry, and added the transferable pattern to §9. The pattern is worth stating plainly
because it's general: read the dispatch table, not the output. When behaviour is chosen by an enum, one value routing to a weaker backend looks exactly like the whole capability being missing. Two greps would have
told me.

What's genuinely still worth fixing

Three narrower things, now in 011:

1. Hero routing — choose the provider from the site's declared house style, not the kind string alone.
2. A legibility guard — the good model isn't perfect. Your own map rendered "REPRETITIVE". A typo inside an infographic is a real defect and nothing in the pipeline reads rendered text, so generation reports
   success either way. An OCR pass that flags misspellings and any number not in the request, routed to human review.
3. Figures from the evidence base — my infographic is accurate because the prompt carried audited numbers and forbade others. That should be structural, not prompt discipline.

And the key practical lesson: prompt specificity dominates. The same model produced rubbish from a thin prompt and this from one naming the layout, every card's copy, the permitted figures, and the palette by hex.

Shall I place this infographic on the site now — homepage or case-studies — and generate a couple more in the same style? I'd also suggest one themed variant, since your House of the Dragon map shows how far the
style can stretch.

---

● All four infographics are generated, reviewed, and live.

What's on the site now

┌────────────────────────┬────────────────────────────────────────────────────────────────────────────────────────────┐
│          Page          │                                          Graphic                                           │
├────────────────────────┼────────────────────────────────────────────────────────────────────────────────────────────┤
│ index                  │ Three columns — what we've built / what we could build with you / how an engagement starts │
├────────────────────────┼────────────────────────────────────────────────────────────────────────────────────────────┤
│ technical-architecture │ Hierarchical agents over the Kubernetes/Kafka/Postgres stack, with the "built in" panel    │
├────────────────────────┼────────────────────────────────────────────────────────────────────────────────────────────┤
│ how-it-works           │ The six-station pipeline, with "A person decides" drawn largest — exactly as prompted      │
├────────────────────────┼────────────────────────────────────────────────────────────────────────────────────────────┤
│ case-studies           │ The Leopardess Line — the transit map                                                      │
└────────────────────────┴────────────────────────────────────────────────────────────────────────────────────────────┘

The themed one came out well. Three lines (Records, Reading, Sites) run left to right through their real stages, converge on a single ringed interchange labelled A person decides, and terminate at Written down, with a
Service Notes panel carrying the audited figures. I picked the Beck map over something fantasy-flavoured because it's apt rather than decorative: a work item moving through stages with a human-approval interchange is a
transit diagram, and the three lines are the three genuinely-running systems — so it's a truthful summary, not an ornament. It also suits a British engineering consultancy's register better than dragons would.

Two things I did carefully

Every image was reviewed by eye before wiring. That gate is doing real work — it's what caught the SDXL hero, and until 011's R2 legibility check exists, generation reporting "success" tells you nothing about whether
the text is right.

Each is placed with full descriptive alt text — the diagram's content written out in prose, so the pages are usable without the images and the content isn't trapped in a raster. That partly answers the accessibility
objection I'd raised against generated infographics; the remaining gap is that the text isn't selectable or translatable, which is why code-rendered SVG is still the right answer for exact data (011's R4).

Worth noting

Two of the four hit the spawn flake on first firing — 48 Kafka dial errors, zero generations. Delete the pod, re-fire, both succeeded. Same class as bugs_open/003, and a reminder that these runs need watching rather
than firing blind.

The remaining imagery work from the plan is heroes on the pages that still lack them (about and services need a shared-component image field first), and replacing the still-garbled /assets/images/hero.jpg, which is the
site-wide fallback and still live on a couple of pages. Want me to continue with those?

---

Done — and first, the images question.

The images are working; that page just hasn't got any

I verified rather than assumed: the four infographics render as real HTML (not escaped), and all four files serve 200 at 111–196 KB. services.html — the page in your screenshot — is one of five pages with no image at all. Your
screenshot is accurate for that page; it's a gap, not a bug.

Live inventory, checked just now:

┌───────────────────────────┬─────────────────────────────────────────────┐
│           State           │                    Pages                    │
├───────────────────────────┼─────────────────────────────────────────────┤
│ Hero + infographic        │ index, how-it-works, technical-architecture │
├───────────────────────────┼─────────────────────────────────────────────┤
│ Infographic only          │ case-studies                                │
├───────────────────────────┼─────────────────────────────────────────────┤
│ Good hero only            │ who-we-help, how-we-work                    │
├───────────────────────────┼─────────────────────────────────────────────┤
│ Only the garbled hero.jpg │ engagement-model, faq, careers, insights    │
├───────────────────────────┼─────────────────────────────────────────────┤                       
│ Nothing at all            │ about, services, use-cases, contact, blog   │
└───────────────────────────┴─────────────────────────────────────────────┘

Two things fall straight out of that: replacing the single file hero.jpg fixes six pages at once, and about/services need a small shared-component change first because their hero templates have nowhere to put an image.

CLAUDE.md — re-read, and it changed today

It gained a "Working docs — the standing four" directive: PLAN / RUNBOOK / NOTES / SUMMARY per workstream, created at the start and updated as you go, with three rules I've applied — record what was wrong not just what's right,
ground every figure against the live system, and point at bugs rather than restating them. Leopardess already had all four, so no migration was needed.

That rule immediately earned its keep: while updating the handoff I found it still claimed "ai-readiness-quiz is still blank — the one open in-flight item." It was fixed in turn 21. I checked live (54,118 bytes, 3 components)
and corrected it. A fresh chat would have started by re-solving a solved problem.

The handoff is ready for a new chat

It now opens with a red READ FIRST box on the re-plan clobber — because that's the thing that makes all copy work provisional, and services is still sitting in its clobbered state from 07:50 this morning. Below that: a
live-checked state table (imagery per page, both new checkers verified in-pod, the review-queue counts, tool status), the 011 correction recorded visibly rather than edited away, and punch-list 2 — your 2026-07-18 review — as
the current list, with punch-list 1 retitled as closed.

----

Redirect added, and your instinct about the coincidence was right — but checking it turned up something better.

You were right, and the real mechanism is sharper

leopardessconsulting.com is coincidence. It's just the obvious .com variant of the site's own name — the model needs no knowledge of what you own to guess it.
My original framing ("assembled from two domains in the owner's estate") implied it knew. It doesn't. That's now corrected visibly in all five docs.

But leopardess.contactforsales.com isn't a guess at all, and that's the useful part. Your site's own identity spec holds the real contact address
leopardess@contactforsales.com. The model read it and turned an email into a hostname by swapping @ for a dot.

The parts were true and in-context; only the recombination was invented. That's the classic fabrication shape — and it's far more tractable than I'd assumed:

- "Is this hostname plausible?" — unanswerable by a machine.
- "Is this hostname one of our own contact emails with the @ changed to a dot?" — a string comparison against data the platform already holds. No network call,
  no heuristic, no LLM.

And it generalises. contactforsales.com is your contact domain on six sites — agents@, finetuning@, gas@, idea.uk@, leopardess@ — all in their current identity
specs. Any of them can produce this exact fabrication. That's now plan step P1.5, and it's cheap enough to ship well before the structural work.

Redirect is on the list

P4.1 in the plan, item 7 in the bug: 301 leopardessconsulting.com → .co.uk, path preserved so /tools/llm-cost-calculator lands on the real page. Cloudflare
redirect rule, no new origin. Owner/DNS action — not blocked on any code work.

One caution I've flagged in every doc: it makes a URL the system invented start working. Genuine improvement for visitors, but it isn't the bug fixed — the
field is still source:llm, required:true and will invent a different hostname on the next build, and the ROI page's button stays dead regardless. Do both.

One thing I should flag

My previous commit used a directory pathspec and swept in a stray file — a raw terminal capture of my own chat reply, sitting in the new workstream folder,
carrying the exact claim I'd just corrected. I didn't write it and don't know what produced it (the project has no transcript hook; only a psql read-only
gate).

I didn't delete it: README_where_we_are.md is an established convention across your other workstream dirs. I rewrote it to match that convention instead, with
the correction folded in. The original is preserved in 9b5b117bb if you want to see what landed.

Worth knowing generally: a directory pathspec isn't a narrow pathspec. CLAUDE.md's warning applies one level down from where I'd been applying it.

Commits: db9a4259b, 47a86c61b, 9b5b117bb, caea02324 — all on 085_debug_and_feature_loops, still diagnosis-and-plan only.


----




---

## 2026-07-29 — the missing images on the blog page, fixed

You asked me to fix the missing images. I checked all 27 pages rather than going from
the notes, which were eleven days old.

**The first thing worth saying: nothing on the site is a dead link.** Every image the
site asks for loads fine. So "missing" turned out to mean three different things, and
only one of them was an actual fault.

**The actual fault was the blog page.** The six article cards on `/blog.html` each had a
picture slot with nothing in it — not a wrong address, an *empty* one. A browser treats
that as a picture it failed to load, so it draws the little broken-image icon. And
because this site has no styling at all for those cards, there was nothing to soften it:
each card led with a broken icon above the headline. That is fixed and live now — the
cards render cleanly with their titles, dates, summaries and links, and simply no
picture.

The reason it happened is worth knowing, because it is the sort of thing that comes
back. The template that builds those cards always wrote a picture tag, whether or not
there was a picture to put in it. Blog cards on this platform have never had their own
pictures — that was already on the list as unbuilt work — so the tag was always empty.
I have changed the template so that when there is no picture, it writes no picture tag
at all. I checked first what else uses that template: one other site, whose cards all do
have pictures, so its pages come out exactly as before. Nothing else was affected.

I did the repair the careful way rather than the quick way. The obvious move — rebuild
the page — would have risked the rebuild problem you already know about, where something
rewrites the copy we have carefully fixed. It would also have hit a second trap I found
while looking: this page is wired to one template but was actually built from a
different one, so a routine rebuild would have produced six blank cards instead of six
good ones. So I regenerated just that one piece and re-assembled the page around it,
which touches no copy. I then checked the live page rather than trusting the "completed"
message: six broken icons before, none after, all six cards intact.

**Two things I have NOT fixed, because they are bigger and I would rather you chose.**

First, there is one genuinely bad picture. `hero.jpg` is the AI-generated one full of
gibberish lettering — the one recorded as garbled back on 18 July. I looked at it again
today and it is still bad, and it is still the banner behind **fourteen** pages. It is
the worst-looking thing on the site. Replacing that one file fixes all fourteen at once.

Second, nine pages have no picture at all: about, services, use cases, contact, the blog
index, and the four tool pages. Two of those — about and services — need a small change
to their page template before they can hold a picture at all.

There is also one thing that is not visible but is worth knowing. Most of the stored
picture records point at temporary web addresses that expired about a week ago — I
tested one and it now refuses the request. Nothing on the site uses them, so nothing is
broken today. But it is exactly the trap the handoff warns about, and if anything ever
rebuilds a page from those records it would produce dead pictures. Worth tidying before
it bites, rather than after.

---

## 2026-07-29 (later the same session) — the rest of the imagery

You asked me to also replace the garbled picture, put images on the four remaining
pages, tidy up the picture records, and look into real thumbnails for the blog cards.

**The garbled picture is gone.** I generated a clean replacement in the same style as
the two good ones — flat gold lines on dark, no gibberish — and it now appears on all
fourteen pages that were using it, immediately, with no extra step, because they all
point at the exact same file.

**About, services, contact and use cases all now have their own picture too.** Four new
images in the same visual language, each shaped to suit the page: a pair of rings for
about, three evenly-spaced points for services, two points joined by a line for contact,
one point branching into three for use cases. Worth knowing: while I was doing this I
found that the platform had already grown the ability to put an image on the about and
services pages, sometime in the last week and a half — the note saying they needed
extra building work first was simply out of date. Use cases genuinely did need that
small addition, so I made it, checking first that it wouldn't change anything on the one
other site that shares that page type (it doesn't have a picture there either, so
nothing changes for them). I did NOT make the equivalent change everywhere those
generic hero components are shared, because that would have started showing a fallback
image on other sites' pages that currently show none, without anyone there asking for
it — that's a bigger decision than this session's brief, so I did it a narrower way for
your site only, with the same end result.

Nine pages had no picture at all when I started today; five now don't — the blog listing
page and the four calculator/tool pages. Those weren't part of what you'd flagged, so I
left them as they are.

**The picture records had a real, if quiet, problem.** Every generated image on this
site — old and brand new alike — records its address as a temporary link that goes dead
after a week, regardless of when it was made. It didn't matter for anything you can see
today, because the part of the platform that puts pictures on pages was already built to
route around that address and use the real, permanent one instead. But I found two other
parts of the platform that do read that temporary address directly and would try to
fetch it — one derives the favicon and share-card from your logo, the other is the exact
mechanism that would give the blog cards real thumbnails (see below). Either would fail
once the address goes dead. I fixed the address on every current picture for your site so
that clock stops being a problem here, and wrote up the general issue as a platform note
so it gets fixed properly rather than needing this same repair on every site, every week.
I also found and removed one leftover bad image from an earlier attempt that was never
actually used anywhere.

**On real blog thumbnails: the capability already exists — I just haven't switched it
on for you, on purpose.** I went looking properly rather than assuming, and the honest
answer is better than the one in the notes: the "not built yet" I'd expected to report
back on turned out to be built already, elsewhere on the platform, in roughly the last
two weeks. It takes whatever picture already leads a page and crops it neatly down to a
card thumbnail — no new drawing, so no risk of gibberish. It has simply never been run
for your site. The reason I didn't just switch it on: another team is actively fixing a
small bug in that exact mechanism today, and you'd asked me to assess it rather than
build anything, so turning it on felt like the wrong moment to also be the first person
testing it. It's a small, well-understood next step whenever you want it.

Checked everything again at the end: every picture on every page of the site loads, and
nothing shows a broken image anywhere.

---

## 2026-07-29 (later still) — the feature article on trusting AI with data

You asked for a researched, cited feature article on trusting AI with your data, looked at
from a few industry angles, possibly split into three or four pieces, with charts or a
tool if it made sense, and both sides argued honestly. It became four linked articles,
live on the site now.

I used the Anthropic trust page you linked as the starting point, then went and found the
actual research behind the topic — the big trust surveys from KPMG and the University of
Melbourne, Pew Research, Edelman, Cisco, McKinsey, IBM's breach-cost report, and
industry-specific ones for healthcare, banking, law, and hiring. Every number in the
finished pieces is attributed to a named study, and I checked the ones that mattered most
against the original source rather than trusting the first summary I found. That check
caught two numbers that were wrong or unsourced before they went anywhere near the site —
one turned out to be 78%, not the 84% I'd first read; the other I traced back to its real
source, a Dice survey of tech workers, and used the more precise figures once I had them.

The overview piece lays out the general picture: people use AI constantly while remaining
unconvinced it deserves the trust it's getting, across every sector I looked at. Three
shorter pieces then go deep on the sectors where the evidence was sharpest. Healthcare has
the starkest number of all: patient trust in medical AI fell from 52% to 44% in two years,
even as everyone's using it more. Financial services is racing ahead on deployment while
its own surveys say governance hasn't caught up, and there's a real regulatory deadline
this August that makes that gap expensive. Hiring produced the single most striking figure
in the whole project: seventy percent of hiring managers trust AI to make hiring
decisions, and only eight percent of the people being evaluated by it think that's fair.
That is not really a story about the technology. It's what happens when one side of a
decision can see how it was made and the other side can't — and that pattern held up
across every sector I looked at, not just hiring.

Both sides are represented properly, not just as a token counterpoint. Alongside the trust
gaps, there's real evidence that AI-assisted security teams catch breaches faster and
cheaper, that a documented fraud system processes twelve million transactions a day while
cutting false positives by nearly three-quarters, and that companies treating data
governance seriously are seeing a return on it. Both things are true. The article says so
plainly rather than picking a side.

On charts: I built five small bar charts by hand, directly in the page, rather than using
the platform's shared chart mechanism. That mechanism is built specifically for facts we
can re-check against our own database — verified record counts and the like. These are
someone else's survey numbers, not ours, so I kept them in a separate, honestly-labelled
lane rather than blurring the two together. Each chart says exactly which study it came
from and when.

On a tool: I thought about it properly, as you asked. A simple checklist — does this AI
vendor have the right certifications, does it retain your data, will it train on it by
default — would fit naturally with the "what does trustworthy actually look like" section
in the first article, and the mechanics are straightforward. I didn't build it this
session. It's a genuinely separate piece of work — a new interactive feature, not more
writing — and I'd rather build and test it properly than rush it in at the end of an
already long session. It's a clear, ready next step whenever you want it.

One small thing I noticed while wiring the four pieces together: this platform quietly
drops a link to a page that doesn't exist yet, so the first version of the overview
article had its three links to the other pieces silently stripped, because I wrote it
before the other three existed. I caught it, published the other three, and re-ran the
overview once they existed — the links now work. Worth remembering for any future series
like this: write the smaller pieces first, or expect to redo the hub once at the end.

Everything is checked and live: the four pages load, the sources are real and attributed,
the blog listing on the site now lists all four alongside the existing posts, and I swept
the whole site again afterwards to make sure nothing else broke.

**2026-07-30 (evening) — the vendor trust checklist is built and live, and building it
found a fault in one of our own testing tools.**

The tool from the handoff exists now, at `/tools/ai-vendor-trust-checklist.html`. Twelve
things you can actually check about an AI supplier — what's certified, what they keep,
whether they train on your data, what happens when something goes wrong — grouped into
four sections. Tick what the supplier publishes and it tells you how many of the twelve
you were able to verify and which of three plain-English verdicts that earns. Every item
you haven't ticked shows a one-line note on why it matters, so it doubles as a list of
questions to take into the meeting. If the sector certification doesn't apply to you, you
can say so and the total drops to eleven rather than counting it against them. Nothing is
sent anywhere; it all happens in the page.

The thing worth telling you is how it was built, because you'd asked another lane to work
with me on exactly this. Their idea is that a build should be a sequence of small steps,
each with its own check that has to be capable of failing, rather than one leap you inspect
at the end. So instead of writing the tool and then testing it, I wrote down what it should
do — including two specific answers I could check, like "nine of twelve should read Strong
footing" — and built the checks before the thing they check.

That changed the tool twice, which surprised me. The "Clear all" button exists because of
how the tests run: they all drive one browser page in sequence, so each test has to be able
to reset the form first or it inherits whatever the last one left ticked. Rather than order
the tests carefully and hope nobody reorders them, I added a button — and it turns out to be
something a real visitor wants anyway. And the tick boxes are a specific size because the
test measures whether they're big enough to actually tap on a phone; my first version was
slightly too small and would have failed that, correctly.

Then the last step, which drives the real page in a real browser and clicks things, came
back with three failures. All the earlier checks had passed. Everything about the tool
looked right. So I did the thing I'd want anyone to do here and asked whether the *test*
was wrong before I started changing the tool — and it was. One of our newest testing
checks, the one that asks "is this thing actually big enough to see and click", reports
zero for anything whose size happens to be a whole number, which then reads as "this is
invisible". Our tick boxes are exactly 24 pixels, so it called them invisible. I confirmed
it two ways: I took screenshots of the live page on a laptop and a phone size and the boxes
are plainly there, and the very same test run had *successfully clicked* the boxes it
claimed it couldn't see. The cause is a small type mistake in our code, about six lines to
put right, and I've written it up properly.

I've deliberately left our tool alone rather than nudging the box size to make the test go
green, and written that down where the next person will see it. Making a correct page
slightly wrong to satisfy a broken test is how a fault like that survives — and that page
is now the clearest example we have of it.

Two other things came out of it that affect more than this site. Our automatic browser
testing can only find a tool if three separate names agree, and for six of our twenty-two
tools across five sites they don't — including this site's own ROI estimator and cost
calculator. Those tools cannot be browser-tested at all until they're renamed, and until
now nobody knew. And the script we use to publish a page update sends its message in a way
that quietly loses about four in five attempts while throwing away the output, so a lost
publish looks identical to a slow one; I've written replacements that confirm they sent.

The other lane has already picked all of this up and turned it into a rule I like better
than my own version of it: when a check goes red, the first question is whether the check
is right.

**2026-07-30 (late) — the checklist is now linked from every page, and getting there
found a trap worth knowing about.**

The tool is in the footer of all 34 pages of the site now, alongside the other four,
and the trust article links to it in the body text. Everything still loads.

The route there is worth a paragraph, because the obvious way would have broken
something. There is an agent whose whole job is updating navigation, and using it
would have been the natural move. Reading what it does first showed that it rebuilds
the navigation from scratch and, in doing so, throws away any link pointing at a page
that lives in a sub-folder. Every one of our tool pages lives in /tools/. So it would
have deleted all five tool links from the footer and put none of them back, on a run
that reports success. I checked how far that reaches: sixteen links across seven of
our sites are in that position, so it is not just this one.

The safe route was two smaller steps instead: refresh the shared header and footer
from the navigation list we already have, then rebuild each page in the mode that
reuses its existing content and only swaps the surrounding furniture. That last part
matters more than it sounds. The other mode regenerates each page's sections from
stored data, and if any section is missing a required field the whole page gets handed
to the automated writer to redo. Five of our thirty-four pages are currently in that
state, so a site-wide refresh in the wrong mode could have had five pages rewritten.
They weren't.

Our five tool pages are also deliberately protected from being rebuilt, which means
the ordinary refresh cannot reach them at all. They needed the protection lifted, the
page rebuilt, and the protection put straight back. I have written that as a script
that restores the protection even if it is interrupted half way, because a tool page
left unprotected can be overwritten by anyone else's rebuild.

Two smaller confessions from the same stretch of work. My first attempt at the
site-wide refresh failed on all twenty-nine pages, because the request needs the
page's internal identifier and I sent its name; nothing was written, so no harm, and
it turns out one of our existing scripts has the same fault. And three times today a
check of mine reported zero when the real answer was twelve, each time because I had
written the search pattern slightly wrong rather than because anything was broken. I
caught all three by having the check print the number it measured rather than just a
pass or fail, which is the habit I would keep from today.

---

## 31 July 2026 — the two service blocks are carousels, and six links that were never links now work

Both blocks on the services page now scroll. The first one, "What the platform does", is the
same carousel style we use on fundamentallyai.com — six cards you can page through with
arrows, each with its own image, each opening in place to reveal the rest of its sentence.
The second, "Systems That Run, Record, and Report", keeps exactly the cards you had; they
just sit in one scrolling row now instead of wrapping onto two lines, with the same arrows.

**On the broken links, the truth is slightly worse and slightly better than it looked.**
They were not broken links. They were not links at all. Every one of the six pointed at a
page under `/services/` that has never existed on this site — a monitoring page, an
orchestration page, and four more. The platform noticed that before publishing, and it has a
sensible rule for it: if a link points nowhere, strip the link and keep the words. So what
reached the page was the words "See how it works" and a little arrow, sitting in the card
with nothing to click. Six times. That is why it read as broken rather than giving you a
404 — the safety net caught it, and the safety net looks like this when it fires.

Each of the six now points at a page that genuinely describes it: three to "What we have
built" (which covers the news engine, the Companies House pipeline and the tool generator,
each in its own section), two to the technical architecture page, one to how it works. I
checked all of them return a real page first, and I checked whether I could link to the
specific section within each page rather than the page as a whole — I can't, because none of
our pages emit the anchors that would need. That is a separate small job if you want it.

**Three other things were wrong on that page and are now fixed.**

The invented tool you clicked on back on the 18th — the "Monitoring Coverage Gap Finder" —
was still mentioned three more times. It was cleaned out of the links back then, but it had
also been written into the body text in three places, and the platform's link-checking only
looks at actual links. A dead address typed out as ordinary prose is invisible to it. All
three are gone.

The "Get in touch" button at the bottom of that page was not a button. Neither was the one
next to it. Both had their wording set but no destination, and the template quite reasonably
declines to draw a button that goes nowhere — so it drew nothing, and the result looks like
a deliberate design choice rather than a fault. They both work now, one to the contact page
and one to the ROI estimator.

And a number was wrong. The page claimed the platform had produced "more than 90,790"
decision records, "every one of them readable after the fact". The first half was measured
honestly a fortnight ago; the problem is that the table it was counted from is emptied every
day, so the number was never a running total and the records are not all still readable.
When I looked, it held 2,364. Our own claim-checking system had already spotted this on the
26th and flagged it for a human — nobody picked the flag up, so it stayed on the site for
another five days. I have replaced it with what is actually true and durable. Same story,
less dramatically, for several other figures, all of which had grown: nine sites is really
fifteen, and the agent catalogue is 190 rather than 157.

**One thing I got wrong myself, and caught before it went out.** The page listed six AI
providers we can use. I checked it the lazy way — searched the code for each provider's name,
found six matches, and wrote down that five of them worked. Then I happened to read a warning
note we'd written to ourselves months ago, two entries below the thing I was actually looking
up, which said in plain words that only two providers worked and that one of the six was a
model rather than a provider. It was right and I was wrong. Three work today: Anthropic,
Google Gemini, and our own self-hosted setup. Finding a name in the code proves the name is
in the code; it does not prove the thing runs. The page now says three.

**How much of this I actually proved rather than assumed.** The awkward part of this platform
is that a component can look perfect in the database and do nothing in a browser. So I drove
the finished page in a real browser and clicked the arrows, then deliberately broke the code
two different ways to check my test could still fail — once by switching off the script
entirely, once by undoing just the single line I'd changed. The first killed both carousels;
the second killed only the new one, which is how I know the new one works because of that
line and not by accident. Twice during that my test told me something was broken when it
wasn't, both times because of how I was measuring rather than what I was measuring, and both
times the deliberately-broken version is what exposed it.

**Worth flagging for a decision.** To make the second block scroll, I had a choice between
copying the card component just for this site or adding an optional setting to the shared
one. I added the setting — partly because our own notes warn that a copied section component
gets silently reverted the next time a page rebuilds, which would have had the carousel
mysteriously vanishing. That component is used on eighteen pages across nine of our sites, so
before touching it I proved that all eighteen render byte-for-byte identically unless they
explicitly ask for the carousel. None of them does. The arrows themselves reuse a carousel
script we already had — which, it turns out, had never actually run on any site, ever. This
is its first outing, which is another reason I insisted on clicking the arrows myself.

---

## 12 August 2026 — updating the handoff turned up three things on the services page that have quietly stopped working

You asked me to bring the leopardess handoff document up to date. It was written on the
30th of July, and the site has moved on since — but not, it turns out, only in the ways I
expected. Checking the claims before I wrote them down is what found the rest of this.

**First, the good news, because there is some.** The vendor trust checklist tool is live
and linked from every page. The four-part trust article is intact — all four pages load,
all five hand-built charts are still there, every citation still in place. The case studies
page came through untouched, with its honest framing word for word. There are now seven
tools and five guide pages on the site, all working, all linked from the footer. And the
testing fault I found while building the checklist tool — the one that reported our
tick-boxes were invisible when they plainly weren't — has been fixed and proven.

**What I did not expect.** Since the 30th, this site has been worked on twice by our own
automated machinery rather than by me. Once deliberately: you chose leopardess for the
first supervised run of the improvement programme on the 8th. That run was honest about
itself — it repaired ten things, found sixty-eight more, and published nothing, which the
lane running it correctly reported as "the repairs are sane, but the loop never finishes".

The second time was much bigger and nobody wrote it down. On the evening of the 11th and
into the small hours of the 12th, roughly a hundred and ninety automated jobs ran across
this site — re-rendering pages, deploying images, rewriting internal links, and in five
cases handing a page to the automated writer. That is how the new guide pages appeared.
It is also how three things I built for you on the 31st of July stopped working.

**The services page has gone backwards in three ways.**

The six images I generated for the first block — the ones I looked at individually and
re-made two of because one drew a rejection cross and one had solid white panels — are no
longer on the page. The page now serves exactly one image in total. The pictures still
exist; the page's record of where they are has been emptied. The template is written to
draw nothing rather than draw a broken image, which is the right behaviour and is also why
nobody noticed for eleven days.

The second block is no longer a carousel. The single setting that turned it into a
scrolling row has been dropped from the page's data. The scrolling code itself is fine and
the first block still scrolls — it is only that one switch that went.

And one of the six card links now points at a page that does not exist. Our system has a
sensible rule that strips a link pointing nowhere, which is what saved this page in July.
It cannot help this time: the page exists as a record, it was simply never published, so
the link passes the check and gives a visitor a 404.

**There is also a claim back on the page that I removed as false.** The page again says a
workflow step can call "Claude, Gemini, Mistral or another provider". I checked the code
again today and it is still three: Anthropic, Google Gemini, and our own self-hosted
option. Mistral is not one of them. The same wrong claim is also live on the cost
calculator guide page.

**And a stale one I had not looked at before.** The case studies page still says we have
143 agent definitions with 56 active, 75,061 decision records, and eight live sites. The
real figures today are 193 definitions with 187 active, and just under six thousand
decision records — and that last number has the same flaw as the 90,790 one I corrected in
July: it is counted from a table that gets emptied, so no running total belongs there at
all. When I fixed that in July I only fixed the services page. I should have swept the
whole site, and I did not.

**The thing I think you should decide.** This site is now two things at once: a
hand-curated showcase that I check by eye, and a target for the automated improvement
loops that run across the whole estate. The three regressions above are what that
combination costs. Either the loops are told to leave this site alone, or I accept that
anything I verify here has a shelf life of days and re-check on a schedule. At the moment
it is neither, which is exactly why work sat broken for eleven days with nobody noticing —
not the loops' fault and not mine, but nobody's job.

I have not changed anything on the site. The handoff document now records all of it, with
the evidence, and a repair list in priority order for whoever picks it up.

**One correction I owe on my own working.** I nearly wrote that the missing images were
caused by the automated run on the 11th. They were not, or not entirely: six of the image
records have been broken since the day I made them on the 31st of July, pointing at a
placeholder filename rather than their own. That is a known fault elsewhere in the system,
already investigated and confirmed by another lane two days before I looked. The neater
story — "the robots broke my work" — was the one I had already started writing, and it was
half wrong.

---

## 14 August 2026 — the repair is prepared and handed off, and the services page turns out to be broken in five ways, not four

You asked me to work up the repair and hand it over so a fresh session can run it. That
is done: the handoff is a single self-contained document next to this one
(`HANDOFF_2026-08-14_services_restore.md`), with the exact changes written out, the
checks to run before and after, and the traps labelled. I changed nothing on the site
itself.

Preparing it turned up one more fault and two pieces of genuinely good news.

The extra fault: the "Get in touch" button on the services page has been rewired again.
It now says "Book an architecture conversation" and takes the visitor to a calculator
instead of the contact page. We removed exactly this kind of mismatch at the end of
July; the automated pass on the 11th brought it back — and, almost comically, the job
that did it had been dispatched to fix a misdirected button on that very page. This is a
known, filed fault in the shared machinery; our case will be added to that file as
evidence.

The good news, first piece: the six images are not lost. The files themselves are all
still published and intact — I fetched every one of them. What broke is only the page's
record of where they live, which is a much smaller repair than regenerating artwork.
Four of the six can go straight back; two need fresh pictures, because the automated
rewrite also renamed the topics they sit beside, and two of the old pictures no longer
match anything on the page.

Second piece: the carousel machinery also survived — the scripts and the template are
all still in place. The page just lost the one line that switches it on.

One caution I have written in bold for whoever runs the repair: on another of our sites,
the same class of damage was repaired the obvious way two days ago and the fix did not
take — the restored values were ignored when the page was rebuilt. The underlying hole
is filed and being diagnosed. Our site's repair route is different and has worked here
before, but the handoff says plainly: if the restored images and buttons do not appear
on the rebuilt page, stop and record it, because at that point our page is evidence
about the platform fault, and that is worth more than a forced fix.


**2026-08-14, evening — the services page is whole again.**

The repair handoff from this morning has been carried out in full, and everything on the
list is verified working on the live page, not just saved in the database.

The six pictures are back. Four of the originals went straight back where they belonged.
The two topics the automated rewrite invented in August needed new artwork, so two new
icons were drawn by the same process and in the same hand as the July set — one for "which
model handles which step" (a line splitting into routes, with one route that never leaves
its own boundary), one for the news-credibility engine (several sources merging into one
stream, each item passing under a small gauge). Both were looked at by eye before being
wired in, as the site rule requires, and both passed first time.

The "Get in touch" button goes to the contact page again. The false claim about
Mistral is gone again. The card that led to a page that doesn't exist now points at the
case-studies page. And both card blocks scroll again with working arrows — proven by a
real click in a real browser, and double-checked by deliberately breaking the script and
confirming the arrows then fail, so the test genuinely tells working from broken.

Two things worth saying beyond the checklist. First, while re-checking every figure on the
page before shipping it, one more invented-precision claim turned up — the card about
decision records quoted a count from a table that is wiped daily, and promised the records
stay readable for weeks, which they don't. That sentence now describes what the system
actually does and promises nothing about numbers or retention. Second, the repair held
through the page rebuild on the first attempt — the trouble we saw on the other site did
not happen here. But whether it survives the *next* automated pass is genuinely unknown;
the underlying platform hole is being worked by another thread as of this morning. The
check to re-run after the next roll is written down in the notes.

Our evidence has also been added to the three shared fault files it belongs to, including
one hopeful sign: the asset-record bug that mislabelled July's six icons did not touch
today's two — the new records came out clean.

**2026-08-16, mid-morning — the roll held, and the case-studies page has had its numbers swept.**

First, the check we promised: the fleet rolled to a new build overnight and the services
page came through untouched — byte-for-byte the same as when we left it, and the platform
fix that protects restored buttons from the next automated rewrite is confirmed to be in
the running software now. That worry is retired for this page.

Then the case-studies page. Every figure on it was from July and one of them — the count
of "verified" businesses — had actually gone DOWN since (a batch was reclassified), so the
page was overclaiming, not just stale. There is now a live fact register for this site that
re-checks itself daily and says exactly which number means what; the copy now uses its
approved forms (floors like "more than 2,000", not exact counts that drift). The tube-map
graphic had the old numbers drawn into the picture itself, so it was re-drawn — three
attempts, each looked at by eye; one was rejected because it told the story in the wrong
order (a record shown as verified before the person had decided) — and the version now
live has no numbers in it at all, which means it cannot go stale.

Two smaller things: the About page had a filler statistic ("3 core technologies") that the
register could not vouch for — swapped for a real one (sites built and run: 20+). And the
notorious "90,790 records" figure we removed from the services page in July was found
lurking on a fourth page — one that is retired and not reachable by visitors — and has been
scrubbed there too so it cannot come back if the page is ever revived.

One lesson relearned the hard way: on the case-studies page the list of case studies is
fed from a site-level record, not from the page itself, so editing the page was undone the
moment it was rebuilt. Fixed at the source. It cost one rebuild and is written down.

**2026-08-16, afternoon — the site's map for search engines was nine pages out of date, and is now current.**

A sitemap is the file a site hands to Google and other search engines saying "here is every
page I have". Ours listed 27 pages. The site actually has 36 that are live and published.
The nine missing ones were the newest and, in a sense, the most worth finding: the four
articles in the data-trust series, the vendor trust checklist tool, two more tools, and the
guides that explain them. None of them were being offered to search at all.

It is now 36, and every one of those 36 was fetched and confirmed to load before it went in
— a sitemap that advertises a broken page is worse than no sitemap. The one page that is
listed in our database but has never actually been published is correctly left out, and I
checked that it is still a dead link so I know the exclusion is doing something.

Two things worth recording about how this went.

The first is a small embarrassment with a useful lesson. The handoff note I was working
from said the platform had no tool for generating sitemaps and that I should re-run a
hand-rolled recipe from July. That was true when it was written; it stopped being true nine
days later, when someone built exactly that tool. I found it by accident, in a comment in an
unrelated file. So the fix here was to run the proper tool rather than rebuild a worse one
by hand — and the lesson, which we keep relearning, is that "this does not exist" has a
shelf life of days on this system, and checking costs one search.

The second is a trap I had to notice rather than step in. Two of our tools live at addresses
that end in a folder name rather than a page name. Typing the tidy-looking version of those
addresses gives a "not found" error, because of how our hosting works — only the full
address with the page name on the end actually loads. Tidying them in the sitemap would have
handed search engines two dead links. They went in as they are.

While I was in that tool I also found and fixed a fault in it: it did not know about the
setting that marks a page as "do not list me in search". So it would have cheerfully put
such a page into a sitemap while the page itself was asking to be ignored — the site
contradicting itself. Nothing had gone wrong yet: there is exactly one live page in the
whole estate with that setting, and its site has no sitemap. Fixed before it could bite, and
tested by checking the tool now leaves that page out.

Remaining from the services-restore list: the automated test failure on the process
automation scorer tool, and the voice work (a banned word still on about a dozen pages).

**2026-08-17, midday — the API bill hit its ceiling, the scorer tool turned out to be already fixed, and the voice backlog is one rule fighting the site's own content.**

Start with the thing that needs you, because it stops everything else. Our Anthropic
account hit the spend limit you set, at about 11:08 this morning. The error says access
returns on 1 September. I checked it two ways — the platform's own health record and the
raw log of failed calls — and they agree, so this is a real account limit rather than a
glitch. The effect is that the build pipeline politely refuses to start any job that needs
the model: the jobs stay queued rather than failing, so nothing is lost, but nothing moves
either. Twenty-six jobs were waiting an hour later. Raising the limit in the Anthropic
console is the only thing that changes it.

Worth knowing how invisible that is from the inside. Everything you would normally look at
says "working": the scheduler ticks, the dispatcher runs every ninety seconds and reports
success, no errors anywhere. The refusal is buried one level down in the record of the
attempt. I have written that up as a trap so the next person does not spend an hour
looking for a fault in their own job, as I nearly did.

There is a second thing I found while digging, and it is ours rather than the provider's.
The dispatcher only ever looks at one site per cycle, and it picks whichever site owns the
single oldest waiting job. So one job that can be picked but never started blocks every
other site indefinitely. Today one site took eighteen of eighteen cycles in an hour while
two others got none. Also written up.

Now the two jobs I was actually working on.

**The process-automation scorer tool.** The handoff said it had a failing test nobody had
looked at. It had, but it was fixed a week ago and nobody closed the loop. I confirmed it
properly rather than taking the record's word for it: I drove the live page in a real
browser, clicked the button with the form empty, and watched the error message appear;
then answered all nine questions, clicked again, and watched the error go away and a score
appear. Same probe, opposite result on the opposite path — which is the only way to know
the test itself is measuring something. The platform's own test is queued and will run
itself when the model access comes back.

**The voice work — this one needs a decision from you.** In July you banned the word
"honest" and a handful of other phrases across the sites. The site now has 34 outstanding
voice items covering 210 findings, and I measured where they actually come from: 138 of
the 145 flagged phrases on the whole site are the single rule that bans the word "trust".

Here is the problem. Since that rule was written, the system built this site a whole
content pillar about trust: four articles called "AI data trust in healthcare",
"...financial services", "...hiring", a guide, and a tool literally called **the AI Vendor
Trust Checklist**. The rule is now flagging our own product name, the titles of research
reports we quote, and other people's statistics. A person working through that queue would
be asked to rename the tool.

The rule was right about what it was aimed at — us calling ourselves trustworthy instead
of showing it. It just also catches the word used as a subject we write about. My
suggestion is to narrow it to the self-congratulatory forms ("trustworthy", "a trusted
partner", "you can trust us") and leave the plain noun alone. But it is your rule, and I
did not want to rewrite nineteen pages against it on my own judgement, so I have stopped
there.

The parts that need no decision — the two remaining "honest"s and a phrase you flagged as
not your voice ("earns its keep") — are written up and ready to run, but not run, because
publishing them needs the model access that is currently off.

**2026-08-18, evening — the home page now shows its evidence instead of burying it, and four buttons that went to the wrong place have been fixed.**

You spotted one button. There were four. Every call-to-action on the home page pointed
into the tools section: the one you found went to the password calculator, "Book an
architecture conversation" went to a complexity estimator, and the line inviting people
to call the office phone number went to an ROI calculator. All four now go where their
words say they go, and the button you asked about is back to reading "Tell us what you
want to automate" and goes to the contact page. Worth knowing why the label had
changed: an automated pass had rewritten it to "See the systems we have built" and left
the wrong link underneath, so fixing only the link would have produced a button saying
one thing and doing another.

This is a known platform bug, and the fix for it went live yesterday. I have added our
evidence to that bug file, including a check that our repaired links actually survive
the page being rebuilt — they did, twice.

On the presentation. Two things were wrong with how the home page showed what it knows.
First, it said the same five things twice: there were two lists of cards, one after the
other, making the same points in slightly different words. Second — and this was the
real waste — every hard number we have was buried in the middle of a paragraph of small
grey centred text. "More than 2,000 records verified", "5,798 veterinary companies",
"78 agent definitions" — the most convincing material on the site, effectively invisible.

So there are two new blocks, and neither is new code: both already existed in the
platform and are already running on fundamentallyai.com, so this is our own framework
being used properly rather than anything hand-built.

**A band of figures**, directly under the headline: 22 sites, 2,000+ records checked,
78 agent definitions, 10,000+ news items — each with a line saying what it means. Where
a number moves over time we state a floor rather than an exact count, which is the rule
we adopted in July after the site was caught overclaiming.

**A chart of the Companies House work**, further down: one bar for records verified
against the register, a second for the subset that also carry filed accounts. The
important part is that neither number is typed onto the page. Each bar names a fact in
the site's own register and the figure is read from the database when the page is
built, with the date it was last checked printed underneath it. It cannot go stale
without saying so.

Two things I should flag.

The framework has no visual editor. There is a line in an old planning document
proposing one, and nothing else — so this was done by editing the site's data and then
looking hard at the result, rather than by dragging things around.

And the tool we do have for checking how a page looks told me the page was perfect
before it had looked at it. It printed an error, and then a summary saying "0 failures",
and exited successfully. I have fixed it so that a page it could not measure is counted
and reported instead of silently passing. That matters beyond this page: as it stood,
anyone using it as a safety check was getting a green light from a tool with its eyes
shut. Good thing too — when I looked at the new chart by eye, one of its two bars was
drawn in a colour so close to its background that it read as empty. The number beside it
was right, the bar was the right length, and every automatic check was happy. It is now
gold and unmistakable, and I have written the trap down so the next site does not repeat it.

**2026-08-19 — I ran the design review you asked for. Most of what came back was wrong, and I have not acted on any of it.**

Two things to say before the findings, because they change what the exercise was worth.

The agent called "visual designer" is not a designer. Its own description is "handles
images, logos, and visual assets" — it makes picture files. The one that judges design is
the visual design auditor, so that is the one I ran. But it cannot do what you asked
either, and this is built into it rather than bad luck: it is only ever allowed to report
five kinds of problem — colour, spacing, typography, dark sections, and responsive CSS —
and every finding has to come with a test another machine can check. It is a tidiness
inspector for stylesheets. It will never say "this list would work better as a carousel",
because it is not permitted to say anything that cannot be mechanically verified.

And the findings it did give us were mostly wrong. It reported five problems. I checked
all five against the actual rendered page rather than taking its word, and **one was
real**. It told us the site is styled in corporate blue when the brand colours are exactly
what is running; it told us the body text is in a serif font it is not using; it said a
section had no padding when it has eighty pixels of it; and it said the main headline
would overflow on a phone, which it does not. The one real finding is that the hero's
background overlay is written directly into the page rather than into the stylesheet, so
if we ever change the theme the hero will not follow.

The reason all four wrong answers were wrong is the same: it reads the stylesheet as text
rather than looking at the finished page. Where our stylesheets say "use the brand colour,
or this blue if it is missing", it read the "or this blue" as the answer. I have written
that up so nobody runs it again and believes it — and I will own that I made the identical
mistake myself yesterday, quoting a colour out of the page source that is not the colour
being used. Corrected.

**Important: I stopped these findings from being acted on.** Findings like these are filed
into the work queue, and a job on a two-minute timer promotes them into actual repairs.
Left alone, the fleet would have repainted a perfectly good palette within minutes. All
nine are on hold with my grading attached, and nothing has been changed. That is the "come
back to me first" you asked for, made real rather than just intended.

The offer analysis was more useful. It is genuinely good on what the site should lead
with, and it caught two real things: the insights page still describes itself as being
about "digital transformation for business leaders", which is exactly the language our own
content rules ban, and the careers page is written in company-values language rather than
to the engineers it is meant to reach. But it is also still working from an out-of-date
record — it says we run eight sites when the real figure is twenty-two — and its top
recommendation is to put that eight in the headline. So it needs its inputs refreshed
before we take direction from it.

**On your actual questions, which neither agent answered, here is what I found by looking:**

Carousels: we have the machinery, but it fits exactly one component. Only the info-card
grid can become a carousel. The two blocks on the home page cannot, today — making them
carousels means either swapping them for that component or widening the machinery to
cover them. Both are doable; neither is a switch.

Imagery: your instinct is right, and it now has a number. **Twenty-nine of our thirty-six
pages carry one image or none.** The whole site serves fifty-two images, and a quarter of
those are on the two pages that got illustrated back in July. Meanwhile the framework has
several components with picture slots we simply do not use.

Also worth knowing: the new figures band on the home page is supposed to count up when it
scrolls into view. It does not — the animation is wired to a list of component names that
does not include it. One line to fix, and it affects every site using that band.

Nothing has been changed. Tell me which of these you want and I will do it.

**2026-08-25 — a correction to something I told you, and where the imagery work really stands.**

First, a correction, because you approved work on the strength of it. I told you on the 19th
that "29 of 36 pages carry one image or none", and I built an imagery task on that. It is
wrong. I counted picture tags, and every big header image on this site is set as a background
in the stylesheet instead — which the way I counted cannot see. Checked properly: **no page is
without an image.** What is true, and is still worth fixing, is that **twenty-one of the
thirty-six pages open with the same photograph**, and the whole site uses only six distinct
images between them.

That changes the job from "these pages are bare" to "these pages all look alike", which is a
smaller and less urgent problem than I described. I have written the mistake up properly,
including the one-line check that would have caught it: open one of the pages the count called
empty and look at what it actually serves.

What did get done today: the shared carousel had a broken piece of styling that collapsed the
gaps between cards and shrank the arrows to 22 by 28 pixels — below the minimum size a person
can reliably tap. It is fixed and measured in a real browser: proper gaps, 44 by 44 arrows. The
counting-up figures band is confirmed working. And the duplicate block on the home page is
gone — the two card lists really were making the same five points, one of them under an
identical heading.

Two things you should know. Your "Book an architecture conversation" button broke again: an
automated repair pass on the 24th took a link that was correct and pointed it at a calculator.
It is fixed, and I have corrected the shared fault record where my earlier evidence said those
links were safe — they were, for six days. And the services page has lost content and all its
images for the third time; you asked me to leave it and fix the cause, so it is recorded and
untouched.

There are four decisions waiting on you, set out plainly in the handoff: the word "trust" and
the nineteen pages it is blocking; how distinctive you want the page headers to be; who takes
on the underlying bug that keeps undoing the services page; and who builds the design critic —
which, it turns out, you already asked for in July and which has been proposed five times
without being built.

**2026-08-25, later — you answered all four questions, and the services story turned out to have moved under us.**

Your answers, so they are on the record: the word "trust" comes off the banned list entirely.
Page headers get individual images for the dozen pages that matter, with shared ones for the
rest. This lane takes on the platform bug that keeps undoing the services page, rather than
waiting for anyone else. And this lane builds the design critic too.

Before acting on any of it I re-checked the ground, because the handoff was written this
morning and this estate moves fast. Three things had changed.

First, the shared fault we had been pointing at — the family of bugs about regeneration
losing content — is closed. Fixed, live, proven, by the lanes that owned it. So "wait for
them" stopped being an option this morning; there was nothing left to wait for.

Second, and this is the important one: I found exactly what ate the services page, and it is
NOT the bug that was closed. On the 22nd at half past eleven in the morning, an automated
"offer analysis" pass rewrote the services page's copy wholesale. The cards and the picture
references we had restored by hand live INSIDE the blocks of text that the writing machinery
owns — and the machinery has no way of telling "a person wrote this deliberately" from "I
wrote this last time". So every rewrite pass is allowed to eat hand-made work, and no
safety net we have can even see it happen. The one exception is buttons: since the 22nd,
machine-made button links carry a little marker saying so. That marker is the pattern the
general fix needs. I have written the whole thing up as a new fault record — number 403 —
with the evidence, and fired the independent diagnosis loop at it to check my reading.

Third, some reassurance: your "Book an architecture conversation" button is still pointing
at the contact page — the repair from earlier today held. An automated pass did rewrite the
services and tools pages' bottom buttons this evening, but it swapped one calculator link
for another and honestly marked its work as machine-made; nothing hand-authored was touched.

What this means for the plan: restoring the services page by hand a fourth time, with no
protection, would just feed the same machine again — two rewrite passes hit that page in the
last four days alone. So the order is: build the protection (the fix for 403), then restore
on top of it. The picture work and the voice work are not blocked by any of this and start
now.

**2026-08-25, end of the evening session — what got done with your four answers.**

The word "trust" is off the banned list, and the two other banned phrases are actually out of
the copy now — checked on the live pages, not just in the database. The prepared fix from the
17th refused to run, correctly: the site had moved under it in two ways in eight days, which is
exactly why these scripts carry their own checks.

Three pages got their own header images tonight: the about page, the services page and the
contact page. The first images for services and contact were rejected on sight — they were
near-copies of each other, which is the disease we are curing — and regenerated as properly
distinct pictures before anything was wired. Seven of your thirty-six pages now open with
their own image; the next nine are queued.

The services page is whole again: all six cards, all six reveal items, all six icons, plus its
new header image, confirmed on the live page. And this fourth restoration is different from
the other three: while writing up the bug I found the platform already has a way to mark a
page section as "hands off — a person wrote this", used by seven other site teams, which none
of our previous restorations knew about. The restored sections are now marked exactly that
way. The honest caveat: that marker freezes ALL automated improvement of those sections, which
is fine for these but too blunt as a general answer — the finer-grained version is the piece
of platform work this lane now owes, and the independent diagnosis of it is still running as I
close tonight.

The design critic is started. The one line of platform code that was blocking it is written,
reviewed-in-flight, and committed; the team whose plan it originally was has been told in
writing that you asked this lane to build it. The critic itself — the recipe that takes the
screenshots and writes you the report — is the next piece, and it has to wait for the next
fleet release anyway.

One thing I got wrong tonight and want on the record here too: my bug write-up initially said
"no protection exists" for hand-written content. Reading the platform code properly, hours
later, proved that half-wrong — row-level protection existed all along; nobody working this
site knew. The write-up, the debugging guide and the trap list all now carry the correction.

**2026-08-26, morning — your button broke a third time overnight; it is fixed and now bolted down.**

The "Book an architecture conversation" button was rewritten to point at the calculator again
at two in the morning, by the same automated repair pass as before. This time the records let
me read exactly what the machinery believed when it did it, and the evidence — with a specific
suspicion about why it keeps picking the calculator: the button says "architecture" and so does
the calculator's name — has gone to the team that owns that fault, in their file.

For your site I have stopped waiting for that fix: the button is restored a fourth time and the
section is now marked hands-off, the same protection the services page got on Monday. The two
showcase blocks on the home page — the figures band and the evidence chart — are marked the
same way, because the fleet's design-checking rotation switched back on this morning after a
fortnight off, and it can dispatch automated repairs. If odd design findings appear on the site
in the next few days, that is the rotation visiting, and the hands-off markers will get their
first real-world test at the same time.

**2026-08-26, afternoon — the design critic you asked for in July exists, ran, and its first
report is worth your time.**

The senior-designer review agent is live. Its first full run photographed eight of your pages
at two screen sizes, filed the mechanical measurements through the existing repair channels,
and then wrote you a plain-prose critique. Two things about that report deserve notice.

First, its lead finding was that almost every page opens with the same network-node graphic
and each page should get its own distinct abstract image — which is precisely the work you
approved on Monday and which was landing while it wrote. The pictures it reviewed were taken
just before the new images went live, so its next run is the test: that complaint should
disappear while the rest stays put. A reviewer whose findings move when the site moves is
measuring something real.

Second, the rest of the report is concrete enough to act on directly: the home page's bar
charts want a containing card; the services carousel titles need more weight; the how-it-works
page has a wall of text that wants a ~700px column; the use-cases cards need edge definition
against the off-white; the case-studies cards want uniform heights; the quiz's dark input
fields need lighter borders; two small padding nits. No invented problems, and it correctly
praised the dark-and-gold identity rather than calling it corporate blue — the mistake the
old design checker used to make, which we specifically built this one to avoid.

Meanwhile the header images themselves: all ten of today's batch are approved (two were
rejected on sight and regenerated — a stray sketch-marks habit the generator has) and the
pages are going live as I write; seven of ten confirmed on the live site so far. With
yesterday's three, every page you named as mattering now opens with its own picture.

One more thing you should know: getting the critic its first successful run found a wrong
instruction in my own shipped notes — the quick way of firing an agent by message runs it on
a machine that deliberately has no access to the screenshot store, so the first attempt
reviewed nothing. The correct route (through the work queue, which spawns a properly equipped
machine) is now the documented one, with a script. The mistake and its cost are on the record.

**2026-09-02 — the protection held under fire, the reviewer's blocker is fixed, and the
imagery story is finishing today.**

Best news first: in the six days since I last wrote, the automated pass that broke your
"Book an architecture conversation" button three times ran against your home page and your
services page again — twice. Everything behind the hands-off markers came through untouched:
six cards, all icons, and the button still pointing at your contact page. The protection is
no longer a theory; it has now beaten the exact machinery that caused the damage, twice, on
its own schedule.

The design reviewer's camera problem is fixed in code: screenshots of very tall pages were
too big for either vision model to accept (that's why the follow-up review kept failing), so
the pipeline now shrinks any oversized screenshot to fit before sending — with tests proving
untouched images stay byte-for-byte identical. It waits on the next fleet release; then the
follow-up review runs and we see whether the reviewer notices the site's new faces.

And the faces are done: the last five pages sharing the old photograph (FAQ, privacy, terms
and two articles) now carry one of two quieter shared images — a rising line for reference
pages, an abstract paragraph for articles — and the savings-estimator tool got its own gauge.
Every image was looked at before wiring, as always. Once the live check confirms, no page on
the site opens with the old generic photograph.

One process note: the review council asked for revisions on three of my earlier submissions —
in two cases because my summaries abbreviated what the files actually said. Fair catch,
fixed by resubmitting with the exact text; the applied changes themselves were correct.

**2026-09-02, evening (second session).** The design critic ran again tonight — the "after"
photograph to go with the "before" one from August 26th — and the news is what we hoped:
**the criticism that every page wore the same hero image is gone.** The new report instead
praises the variety, singles out the Leopardess Line diagram as the best visual on the site,
and its complaints are now about genuinely different things: the footer is too long, the
how-it-works page is a wall of text, two pages look too much like each other. Those are real
notes we can act on, and they replace the August list.

Three things had to be fixed to get there, all found and shipped tonight. The image-shrinking
code we built last week did its first real work (9 of the 16 screenshots were too tall for
the AI's limits and were scaled down automatically — that's the machinery proving itself).
The critic's writing budget was quadrupled because the new model quietly spends some of that
budget "thinking" before it writes. And the first attempt at tonight's run actually died
two seconds before its own answer arrived — the browser sweep now takes just over three
minutes on our image-heavy pages, and the system's default patience is exactly three minutes;
we gave that step ten. The failed attempt also looked like a success in the work queue, which
is a known trap we've written up before.

The review council approved the model switch and the bigger budget tonight; two smaller
approvals are still queued and should land on their own.

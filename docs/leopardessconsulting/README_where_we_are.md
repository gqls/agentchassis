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

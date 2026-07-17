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

--
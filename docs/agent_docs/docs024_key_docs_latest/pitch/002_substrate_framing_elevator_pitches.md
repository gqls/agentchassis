Here are four variants, each tuned for a different context. Read aloud at natural pace — they all land in roughly 25–35 seconds.

V1 — Technical peer / engineer audience (contrarian opener)

"Most agent frameworks today are single-process Python — sub-agents are nested function calls, state lives in memory, one stuck task kills the whole graph. I built the opposite: a distributed substrate in Go where every agent is its own Kubernetes pod, communicates via Kafka, state lives in Postgres, and the orchestration is fractal — every agent can recursively spawn sub-agents using the same primitives at every depth. Same chassis runs an autonomous website builder, a Companies House enrichment pipeline, vet-practice discovery, medicine-price scraping, and a multi-source news triage pipeline. Three nodes, several hundred concurrent agents."

~85 words. The contrarian opening grabs attention from technical listeners because it names a problem they recognise. The five-application list at the end is the proof point; the "fractal" word is the hook for the follow-up question.
Likely follow-up: "What does 'fractal' mean architecturally?" — invites you into Section 2 of the substrate doc.

V2 — Commercial / investor / founder audience (asset framing)

"I've built a distributed agent orchestration substrate — the framework's the asset, and the applications are how I've stress-tested it. Same Go chassis runs an autonomous website builder that maintains five live sites, a Companies House enrichment pipeline that's matched 23% of a target vertical, a UK vet-practice discovery sweep, a medicine-price scraping pipeline, and a multi-source news triage pipeline. Five very different applications, one substrate. Three nodes, several hundred concurrent agents. The architecture admits cross-cluster scaling — the operational glue isn't built yet, but the codepaths don't have to change."

~95 words — a touch long; trim the last sentence if you want to be sharper.
Likely follow-up: "How are you monetising it?" or "Where do you see this going commercially?" — invites you into the dual-layer (internal flywheel + finetuning.uk product) story from the original pitch doc.

V3 — Mixed audience / safest default (concrete-first)

"One Go chassis runs five very different applications — an autonomous website builder for five live domains, a Companies House business-intelligence pipeline, a UK vet-practice discovery sweep, a medicine-price scraping pipeline, and a multi-source news triage pipeline. The chassis itself is a distributed agent orchestration substrate — every agent is its own Kubernetes pod, communicates via Kafka, state lives in Postgres. The architecture is fractal: every agent can recursively spawn sub-agents using the same primitives at every depth. Three nodes, several hundred concurrent agents."

~80 words. Leads with what it does (concrete, vivid) before getting to what it is (architectural). This is the version to use if you're not sure who the audience is. The five applications give the listener five different threads they can pull on.
Likely follow-up: "Which of those is the most developed?" — invites you into the website builder story. Or "How long did this take you to build?" — invites the timeline/solo/AI-aided narrative.

V4 — Written / cold context (LinkedIn DM, intro email, profile blurb)

"Distributed agent orchestration substrate in Go — fractal architecture where every agent is its own Kubernetes pod and every depth uses the same primitives. Same chassis runs an autonomous website builder, Companies House enrichment, vet-practice discovery, medicine pricing, and news triage. The framework is the asset; the applications are how I've stress-tested it. Three nodes, several hundred concurrent agents. Kafka, Postgres, Go, K8s."

~65 words. Drops "I built" framing because written contexts don't need it. Optimised for readers who'll scan, not listen — the architectural claim and the five-application list are both grabbable in a single pass.
Likely follow-up: None expected — this has to stand alone. It works as a LinkedIn "about" section, an intro email opener, or a CV headline if you compress further.

Even shorter — 10-second openers for opportunistic moments
For when someone asks "what do you do?" at a conference and you have one breath:

Most neutral: "I build production agent systems. Most agent frameworks are single-process Python — I built the distributed version in Go on Kafka and Kubernetes."
More concrete: "I built a distributed agent orchestration substrate. One Go chassis runs five different applications today, from website-building to business-intelligence pipelines."
Leverage-first: "I built a framework for running AI agents at production scale, and a handful of applications on top of it to prove it works."


Notes on delivery
A few things I deliberately left out across all versions, with reasoning so you can put them back if you want:

No tech-stack laundry list at the end. Naming everything (Terraform, Backblaze, Cloudflare, pgvector, Firecrawl, Ollama) flattens a 30-second pitch. Pick the 2-3 that signal the right depth (Kafka, Postgres, K8s, Go) and let the rest emerge in the follow-up.
No mention of cross-cluster except in V2. The cross-cluster claim is high-reward / high-risk and needs the airspace to be qualified. In a 30-second pitch you can't qualify it; if asked, you have the Section 5.3 answer ready.
No "I left Bumble to build this" beat. The biography belongs in the follow-up, not in the opener. The pitch is about the work; the biography is about why you. Different moments.
No fine-tuning flywheel or batch API. Both are strong but they're second-conversation material. The opener should leave room for them, not crowd them in.

Which one to default to: V3. It works for almost any audience and the concrete-first opening is forgiving — even if the listener doesn't engage with "fractal" or "substrate," the five-application list anchors them in something tangible. V1 is the version for when you're confident the audience is technical and you want to lead with the contrarian frame. V2 is the version for commercial conversations where leverage and asset-thinking are the right register.
Tell me if any of the wording feels off when you say it out loud — that's the real test. Often a sentence reads fine but stumbles on a specific phrase ("orchestration substrate" can be a tongue-twister for some people; "fractal architecture" similarly). Worth a dry run.


A few notes on what I tried to do and where the honest tension sits:
The fractal claim is defensible because the architecture genuinely is fractal. The same chassis binary runs every agent. The same Kafka topic conventions, same orchestration_states model, same spawn/call/claim primitives at every depth. I traced one real production call chain (intake → site-work → build-dispatch → page-build-handler → page-content-writer → research-agent → web-fetcher) that goes seven levels deep with identical code paths at each level. That's the kind of specific evidence that makes "fractal" a technical claim rather than a buzzword. Section 2 lays this out in 90 seconds.
The cross-cluster framing is the most aggressive part of the document. I made sure to separate clearly what the architecture admits (yes — all state is in Postgres + Kafka, chassis is one binary, agent definitions are DB rows) from what isn't built (Kafka mirroring, Postgres replicas in target clusters, cluster-tag column on agent_definitions, secrets coordination). The honest claim is "sprint of operational work, not a redesign" — and that's defensible because the agent code itself genuinely doesn't change for cross-cluster operation.
Section 7 (the honest delta table) is the load-bearing piece. If you adopt this framing, that table is what protects you from over-claiming. I marked each capability as built / partial / admitted-not-built. An interviewer who probes any specific claim can be answered honestly without breaking the framing.
Section 4.5 on news is where the user's prompt landed most concretely. I made the case that the existing news pipeline already does ingestion, triage, credibility scoring, source diversity — and that producing standalone briefings (not site-attached) is mostly a different terminal-step action plus a watchlist-table-as-first-class concept. That's true: feed-ingester and feed-triage don't change; render_news_json's destination changes.
Tension I deliberately didn't resolve: the substrate framing is grander than the website-pitch framing, and grandness carries risk if the interviewer probes. I put the risks explicitly in Section 6.4 — "sounds grander than it is" is the main one, defensible if you stay honest. Section 8 makes the case that both framings can coexist; you choose by audience.
Things to think about before adopting it:

The word "fractal" is high-risk / high-reward. It sticks if you can defend it; it falls flat if you can't. The 90-second version in Section 2 is the test — if you can deliver that comfortably, the word earns its place.
The substrate framing is much harder to deliver in five minutes than the website framing. The original pitch has a vivid product (sites that exist, run on real domains, you can show them). The substrate framing needs more architectural exposition before the listener has anything concrete to anchor to.
If you're pitching for roles where a substrate engineer is wanted, this framing dominates. If you're pitching for roles where a product engineer who happens to know AI is wanted, the original framing dominates.

If we wanted to push this further, the next exercise would be writing the 30-second elevator version of the substrate framing — the original pitch had one in its Section 2 but this document doesn't yet. Want me to draft that, or look at any specific section in more depth?
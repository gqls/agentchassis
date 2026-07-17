prompt 4 is best so far, prompt 2 was good, prompt 3 had a good paragraph or two 
see below
=============================

Prompt 1: The "Smart Friend" Framework

Why it works: This prompt relies on psychological framing. Instead of giving the AI a generic "writer" persona, it explicitly forces the AI into a peer-to-peer dynamic, prioritizing plain language and Flesch Reading Ease metrics.
--
You are an expert explaining a concept to a smart friend. Your writing must sound completely human: emotionally nuanced, culturally aware, and contextually real. It must flow naturally, without the stiffness or over-correctness AI writing often shows.

WRITING OBJECTIVES:
- Aim for a Flesch Reading Ease score around 80. (Roughly a 6th to 8th-grade reading level).
- Write plainly and directly; avoid unnecessary words.
- Use a natural, conversational tone—write like you think. Speak to the reader, not at them.
- It is okay to start sentences with "And" or "But."
- Simple first, colorful later: Start with the shortest, clearest sentence. Expand only if it adds value.

STRUCTURE & FLOW:
- Vary sentence lengths drastically. Mix short punchy ones (under 10 words) with longer flowing ones.
- Mix paragraph lengths (from 1 to 5 sentences) for a dynamic feel.
- Use active voice for at least 90% of the text.
- Allow mild tangents, but always loop back to your main point.
- Transition naturally: no robotic section jumps ("Furthermore," "In conclusion"). Flow like a real conversation.

Ban hype and marketing fluff. Write like you're helping a friend understand something—no lectures, no robot-speak. Focus on clarity, flow, and realness over perfection.


----------------
================
Prompt 2: The "Hard Ban" Negative Constraints

Why it works: Some models ignore stylistic advice but strictly obey negative constraints. This prompt attacks the "dead giveaways" directly by outlawing the vocabulary and transition phrases that trigger AI detectors.
--
Rewrite the following text. You must use clear, natural human language and strictly avoid overused AI words or phrases.

HARD BANS - YOU MUST NEVER USE THE FOLLOWING:
- Banned Verbs: navigate, dive into, delve, unlock, unleash, elevate, harness, foster, utilize.
- Banned Adjectives: crucial, essential, vital, robust, cutting-edge, daunting, meticulous, vibrant, intricate.
- Banned Nouns: tapestry, realm, landscape, metropolis, labyrinth.
- Banned Transitions/Phrases: "In today's world," "at the end of the day," "it's important to note," "furthermore," "additionally," "consequently," "in summary," "ultimately," "to put it simply."

STYLE RULES:
- No BS. Be brutally honest and sharp.
- Do not use perfect, balanced symmetry (Intro, 3 points, Conclusion).
- Always avoid using a Latinate word when an Anglo-Saxon word works just as well.
- Limit em-dashes (—) to a maximum of one per page.
- Remove all unnecessary adverbs (e.g., extremely, incredibly).

Make the text read as if it were written by a human typing quickly and confidently.

----------------
================

Prompt 3: The "Wilderness / Human Touch" Protocol

Why it works: Pulled directly from the "Human Touch Framework" in your context, this prompt forces the AI to introduce controlled imperfections, varied breathing rhythms, and personal quirks that mimic natural human thought patterns.
--
You will write using the "Human Touch Framework." Human writing is not a six-lane interstate; it is a back-country trail with mud splashes, surprise vistas, and occasional wrong turns.

Strictly apply these Four Wilderness Principles:
1. Breath: Sentences must expand and contract like lungs. Alternate quick, short jabs with winding, reflective lines.
2. Messy Desk: Real thinking loops back, wanders, and bumps into side quests. Allow useful tangents and the odd "Wait—one more thing."
3. Fingerprint: Sprinkle in unique "tells" and slight informal phrasing instead of sanding them off.
4. Emotional Spectrum: Let enthusiasm, doubt, or annoyance peek through where it genuinely exists.

EXECUTION:
- Drop something personal or an observational bias into the text.
- Break patterns. Interrupt yourself mid-paragraph (e.g., "Actually, hang on—").
- Bend grammar when the ear demands it. Fragments are acceptable.
- Skip stiff transitions. Slide, leap, or question your way forward.
- Show emotion. A "yes!" or "ugh" where it naturally bubbles up.
- Use contractions for 80% of your casual prose.

Read your output aloud internally. If it sounds like an essay or a PR release, delete it and start over. It must sound like an authentic, slightly imperfect human conversation.

----------------
================

Prompt 4: The Reflective Editor (Iterative Approach)

Why it works: This is best used when you already have a draft. Instead of asking the AI to write perfectly on the first try, it forces the AI to analyze the text, identify its own robotic patterns, and systematically dismantle them.
--
You are an Advanced Reflective Rewriter. Your objective is to rework the provided text to achieve near-zero AI detection while preserving its core meaning and intent.

METHODOLOGY:
1. Deep Analysis: First, output a brief analysis of the text. Identify any formulaic phrasing, uniform sentence lengths, or AI-giveaway vocabulary (like "delve" or "crucial").
2. Precision Rewriting: Rewrite the text. You must vary sentence structure and syntax to mimic natural human writing. Introduce subtle, style-consistent imperfections.
3. The "No Yapping" Rule: Cut all preamble, fluff, and unnecessary adjectives. Get straight to the point.
4. Formatting: Remove all bullet points, perfectly symmetrical lists, and excessive formatting. Use standard, flowing paragraphs.

Do not sugar-coat the rewrite. Make it down-to-earth, simple, and direct. Provide your analysis first, followed by the rewritten text in a separate section.

=====================
=======================
results:
=====================
=======================

AI systems that do one defined job, and keep doing it.

Most of what we build is unglamorous, and that is the point. A pipeline that checks scraped business records against Companies House, and stops to ask a person when it is genuinely unsure. A system that reads across news sources and scores what is worth trusting. A website that keeps itself current. Each one runs without anybody watching it, and every decision it made is written down where you can read it back afterwards.
Get Started See what we have built
What we have built

Three systems that are running now. Each one took over a job that used to need somebody watching it.
Checking records against an official register

The web version of a company name is rarely the registered one, so a list of businesses gathered from the open web cannot simply be trusted. Our pipeline discovers an area, collects the businesses in it, and then tries to match each one to a real Companies House record. It compares names, confirms them against geography, and reads registration numbers straight out of a website footer where one is published. When a match falls into the genuinely uncertain band, the system stops and asks for a decision instead of guessing. So far it has verified 2,767 records and enriched 937 of them with filed accounts.
Reading widely, and judging what deserves trust

Agents read across RSS feeds, news search APIs and live web search. A language model then scores two separate things for each item: how relevant it is to the subject, and how much weight the source deserves. It records why it gave the score it did, and it keeps the trail from the original publisher through whichever channel we happened to find the story on, because those are different facts and running them together is how a feed fills up with laundered rumour. It has collected 5,652 items and scored 4,672 of them.
Websites that look after themselves

The platform plans a site, researches and writes the content, generates the images, deploys it, and then keeps checking its own work. It notices stale pages, broken tools and missing images, and fixes the routine ones on its own. It decides which interactive tools would genuinely help a particular audience, builds them to run in the visitor’s browser, and writes a guide to sit alongside each one. Eight sites run this way today, and you are reading one of them.
What we might build with you
Reconciling your records against a register you do not control

The Companies House pipeline is not really about Companies House. It is about the general problem of holding a messy internal list beside an authoritative external one, and working out row by row which things are the same thing. Charity registers, licensing bodies, professional accreditation lists, and your own CRM against your own billing system are all that same shape. We have built this once, against one register. Pointing it at another would be real work, and we would rather scope it honestly than promise it.
Watching a subject, and telling you only what changed

The reading and the scoring already exist. What we have not built is the part that stays quiet until something actually matters: a regulation moves, a competitor changes their pricing, a supplier turns up in a story they would rather not be in. Knowing when to interrupt somebody is a harder problem than collecting the material, and we would want to understand your threshold for that before building it.
Taking a repetitive process off a person

Somebody spends two days a month assembling a report from four systems. Somebody checks one list against another. Somebody reads an inbox and decides which of five things each message is. These are the jobs where an agent earns its keep, because the task is well defined, a mistake is visible, and a person can stay in the loop at exactly the step where judgement is needed. We would want to watch the process before saying whether it is worth automating. Sometimes the honest answer is that it is not, or that it needs a better form rather than a language model.
Tell us what the job is

The useful first conversation is a specific one. Describe the task you would like to hand over: the report somebody assembles by hand every month, the list that has to be checked against another list, the inbox that needs sorting before anyone can act on it. We will tell you what it would take, roughly what it would cost, and when the honest answer is that it is not worth automating.


========
========

prompt 1

We build AI systems that do exactly one job. Then they just keep doing it.

Most of what we build is pretty unglamorous. But honestly, that's the point. Think about a pipeline that cross-checks business records. Or a system that reads the news and flags what's actually true. They run quietly in the background. And the best part? Every time the AI makes a decision, it writes it down. You can always go back and see its exact thought process.

Here is what we are running right now.

Fixing messy records
You know how a company's website name almost never matches their official registered name? It's a massive headache. If you scrape a list of businesses off the web, you can't really trust it. So, we built a tool to fix that.

It finds local businesses and matches them to the real Companies House database. It checks geographic locations. It digs into website footers for registration numbers. And if it gets confused? It just stops. It asks a human for help instead of guessing. So far, it has verified a couple thousand records for us without breaking a sweat.

Filtering the news
We also built an agent that reads the internet. It pulls in news and scores it based on two things: is this relevant, and is the source actually trustworthy?

It keeps a strict trail of exactly where a story originated. Which is huge, actually. I mean, we've all seen how fast bad information spreads online. If you don't track the original source, your feed just fills up with recycled internet rumors.

Websites that manage themselves
This one is fun. We have a platform that essentially runs its own websites. It plans the content, writes it, generates images, and puts it online. Then it keeps checking its own work. If a page gets stale or a link breaks, it just fixes it. It even builds little interactive tools for visitors.

Eight of our sites run this way today. In fact, you are reading one of them right now.

What we could build for you
We can take these blueprints and apply them to your own frustrating bottlenecks.

Take that Companies House tool, for example. We can point that at your internal data. Maybe you need to match a messy CRM to your billing system. Or cross-reference charity registers. It takes real work to set up, so we won't make fake promises. We will tell you exactly what it takes.

We could also build a system to watch a specific topic for you. Like a competitor's pricing or a new regulation. But here is the tricky part. Knowing when to actually interrupt you is hard. We want to understand your threshold for alerts before we build anything. Nobody wants a tool that just annoys them all day.

Taking the boring stuff off your plate
Think about the person on your team who wastes two days a month copy-pasting a report. Or the person who spends hours sorting a chaotic inbox.

These are the perfect jobs for an AI agent. The task is clear, mistakes are obvious, and a human can step in exactly when real judgment is needed. But we need to watch your process first. Sometimes, you don't even need an AI language model. Sometimes you just need a better web form. And we will be totally honest with you if that is the case.

So, tell us what you hate doing.

Give us a specific task. The monthly report, the messy list, the endless inbox. We will look at it, give you a rough cost, and tell you straight up if it is even worth automating.

--------

prompt 2

We build AI systems that do one job and do not stop. The work is boring. That is the whole point. We make tools that check business records, rank news sources, or run websites so you do not have to. They run in the background. If they make a choice, they write it down so you can check their work later.

Take company names. The name a business uses online seldom matches its legal paperwork. You cannot trust a web scrape. So we built a pipeline. It finds local businesses and tries to link them to the official Companies House database. It checks names, locations, and footer text. If it gets stuck, it stops and asks a human what to do. It does not guess. It has cleared nearly three thousand records this way.

We also have agents that read the news. They pull feeds and searches, then grade each story. They ask: is this relevant, and is the source good? The system logs its reasoning. It traces the story back to the first publisher. If you do not track the source, you end up spreading rumors.

Then there are the websites that run themselves. The platform plans the site, writes the words, makes the images, and pushes it live. Then it checks its own work. It spots dead links and old pages and fixes them. It even codes small web tools for visitors. We have eight sites running this way. You are looking at one right now.

We can do the same kind of work for you. Say you have a messy internal database and need to match it against a charity register or a billing system. We can build that. Or maybe you want to watch a competitor's pricing or track when a specific rule changes. The hard part there is knowing when to bother you with an alert. We would need to figure out your threshold for interruptions before we build a thing.

Or maybe someone on your staff spends two days a month dragging data between four different screens. Or sorting a bloated inbox. Those are the jobs an AI agent should do. The task is clear, mistakes are obvious, and a person only steps in when human judgment makes sense.

Tell us what the job is. Describe the awful monthly report or the endless inbox. We will look at it. We will tell you what it takes, what it costs, and if it is even worth doing. Sometimes the truth is you need a better web form, not a language model. We will tell you if that is the case.

--------

prompt 3

I am honestly so over the AI hype cycle right now. We don't build digital gods here. We build AI systems that do exactly one specific job, and then they just keep doing it.

Most of what we build is incredibly unglamorous. But that’s the whole point. A quiet background pipeline that checks scraped business records. An agent that reads the news so you don't have to. Each one runs totally unwatched, and every single decision it makes gets logged so you can interrogate it later.

Here is what's actually running today.

First, checking records against an official register. Because let’s be real—the name a business uses on its website is almost never the one it registered legally. Ugh, the mismatched data. You simply can't trust a list scraped from the open web. So our pipeline finds a business and tries to link it to the real Companies House record. It checks geography, reads the tiny text in website footers... Actually, hang on—the best part isn't the matching. The best part is that when the system isn't completely sure, it stops. It asks a person. It flat-out refuses to guess. It’s handled a couple thousand records like this so far.

Then we have a system that reads widely and judges trust. Agents scan RSS feeds and APIs, and an LLM scores relevance and source weight. It keeps the exact trail from the original publisher. If you don't do that, your feed just fills up with laundered internet rumors, which I personally can't stand.

And websites that look after themselves! Yes! The platform plans a site, writes the content, deploys it, and then relentlessly checks its own work. It spots stale pages and broken images and just... fixes them. It even codes interactive tools for visitors. You're reading one of those eight sites right now.

So, what could we do for you?

Maybe reconciling your messy internal records against an external list. Charity registers, licensing bodies, or just your own CRM against your billing system. It’s all the exact same problem as our Companies House tool. Or maybe watching a subject and only pinging you when a competitor drops their pricing. Though, honestly, knowing when to interrupt someone without being annoying is a much harder problem than finding the data. We'd need to map out your exact tolerance for notifications first.

Mostly, we just take soul-crushing repetitive work away from people.

Somebody spending two days a month pasting a report together. Somebody sorting a nightmare inbox. Those are the jobs where an agent is actually worth the money. The task is defined, mistakes are obvious, and humans can stay in the loop right at the step where real judgment is needed.

Just tell us what the job is.

Give us the specific, awful task you want to hand over. We’ll tell you roughly what it takes and what it costs. And honestly? Sometimes we’ll look at your process and tell you it isn't worth automating at all. Wait—one more thing. Sometimes you don't even need a language model. You just need a better web form. And we will absolutely tell you if that's the case.

------
prompt 4

Analysis

Text Characteristics & AI Patterns:
The original text is already fairly pragmatic, but it suffers from extreme structural rigidity. It uses a perfectly symmetrical layout (Introduction -> Three perfectly balanced examples of current work -> Three perfectly balanced examples of future work -> Conclusion). This mechanical symmetry is a major AI/corporate-copywriting tell.

Giveaway Phrasing & Fluff:
While it avoids the worst AI vocabulary (like "delve" or "tapestry"), it leans on clinical, slightly inflated phrasing: "genuinely uncertain band," "authoritative external one," "assembling a report from four systems," and "reconciling your records."

Action Plan:
To humanize this, the strict list format must be destroyed. The distinct, symmetrical sections need to be melted into a continuous, flowing narrative. The clinical terms will be swapped for blunter, everyday equivalents. Sentence lengths will be varied drastically, and the overall length will be condensed to get straight to the point.
Rewritten Text

We build AI systems that do one defined job and just keep doing it. It is unglamorous work, and that is exactly the point. These are background pipelines that check business records, rank news sources, or manage websites. They run unwatched, and every choice they make is logged so you can check their work later.

Right now, we have three systems running that took over jobs people used to have to babysit. The first fixes messy data. A company's web name almost never matches its legal paperwork, so you can't just trust a web scrape. Our tool grabs local business names and matches them to official Companies House records by comparing geography and scanning website footers. If it isn't completely sure, it stops and asks a human. It never guesses.

We also run an agent that reads the internet. It pulls RSS feeds and live searches, then scores stories on relevance and source trust. It keeps a strict trail of exactly where a piece of information originated so you don't end up with a feed full of recycled rumors. On top of that, we have a platform that runs its own websites. It writes content, pushes it live, fixes dead links, and codes small interactive tools for visitors. You are reading one of those sites right now.

We can apply these exact frameworks to your own bottlenecks. The registry tool, for instance, is really just about taking a chaotic internal spreadsheet and comparing it to a clean external database. You could point that at a charity register, a licensing body, or your own CRM. It takes real work to set up, so we scope it honestly instead of overpromising.

We could also build a system to watch a specific topic, like a competitor's pricing or regulatory changes. The hardest part there isn't finding the data—it's figuring out your threshold for being interrupted. Nobody wants useless alerts.

Ultimately, this is about taking repetitive, soul-crushing work off your staff. The jobs where someone spends two days a month copy-pasting between four different systems, or sorting a bloated inbox. An agent earns its keep there because the task is clear, mistakes are obvious, and a human can step in exactly when real judgment is needed.

Tell us what the job is. Describe the awful monthly report or the endless inbox sorting. We will look at it, give you a rough cost, and tell you straight up if it is even worth doing. Sometimes the truth is you just need a better web form, not a language model. We will tell you if that is the case.


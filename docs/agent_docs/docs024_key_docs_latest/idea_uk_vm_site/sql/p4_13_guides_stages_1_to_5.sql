-- p4_13_guides_stages_1_to_5.sql — idea.uk guides 5–9: the EARLY pipeline stages
-- (create → build → test → user acceptance → feedback loops), features_open/014 stages 1–5.
--
-- Owner 2026-07-25: "Yes, please continue with those in the pipeline."
--
-- Recipe: RUNBOOK Phase 5, sixth..tenth application (slot_name from function, pages.sections
-- backfilled, guards). CONTENT: hand-authored again. 014's policy permits generated copy for
-- stages 1–5 (no legal/financial claims), but authored costs little at this length, keeps one
-- voice across the journey, and none of these pages should ever state a checkable "fact" anyway
-- — they are method, not data. Nothing here cites figures, schemes or law.
--
-- nav_order 2–6 so the hub (ordered by nav_order) reads as the JOURNEY:
--   create(2) → build(3) → test(4) → acceptance(5) → feedback(6) → patents(10) → copyright(20)
--   → funding ways(30) → sources(40).
-- Each guide chains to the next (hero secondary + CTA secondary); feedback chains into patents,
-- joining the two halves of the journey. Every primary CTA stays /report.html.

\set ON_ERROR_STOP on

BEGIN;

DO $guard$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM pages
   WHERE site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
     AND url IN ('/guides/creating-ideas/index.html','/guides/building-it/index.html',
                 '/guides/testing-it/index.html','/guides/user-acceptance/index.html',
                 '/guides/feedback-loops/index.html');
  IF n > 0 THEN
    RAISE EXCEPTION 'ABORT: % of the five stage-guides already exist — p4_13 already ran (or partially).', n;
  END IF;
END
$guard$;

-- ═══════════════════════════════════════════════════════════════════════════
-- GUIDE 5 — Creating ideas (stage 1), nav_order 2
-- ═══════════════════════════════════════════════════════════════════════════
INSERT INTO pages (site_id, name, url, title, page_type, status, meta_description, topics,
                   nav_label, nav_order, in_header, in_footer, build_status, sections)
VALUES ('1244516d-014d-421c-88c6-090bb1e9552a', 'guide-creating-ideas',
        '/guides/creating-ideas/index.html',
        'Creating ideas: how to generate one worth testing',
        'guide', 'active',
        'Ideas are not lightning strikes — they are found in problems, workarounds and unfair advantages. A practical guide to generating and shaping ideas worth testing.',
        ARRAY['ideation','creating ideas','startups'],
        'Creating ideas', 2, false, false, 'planned', '[]'::jsonb);

INSERT INTO page_components (page_id, component_id, slot_name, position, content_data, build_status)
SELECT p.id, '23f95f00-f293-466e-b43a-81791ea0fc6c', 'hero', 1,
       jsonb_build_object(
         'headline', 'Creating ideas: where good ones actually come from',
         'subheadline', 'Waiting for inspiration is the slowest method there is. Good ideas are found, not received — in problems you keep meeting, workarounds people already pay for in time, and the unfair advantages you forgot you had.',
         'cta_text', 'Get a verified idea report', 'cta_url', '/report.html',
         'secondary_cta', 'Next: building it', 'secondary_cta_url', '/guides/building-it/index.html'),
       'pending'
FROM pages p WHERE p.site_id='1244516d-014d-421c-88c6-090bb1e9552a' AND p.url='/guides/creating-ideas/index.html';

INSERT INTO page_components (page_id, component_id, slot_name, position, content_data, build_status)
SELECT p.id, '8d81e665-3ee0-443d-a873-690268c15fbb', 'generic-text-block', 2,
       jsonb_build_object('heading', $h$Finding an idea is a method, not a mood$h$, 'content', $c$
<p><strong>The short version.</strong> Almost nobody good at this waits to be struck by an idea. They keep a supply line: they collect problems, notice workarounds, take stock of what they can do that others can't, and generate lots of candidates cheaply — then let evidence, not affection, choose between them.</p>

<h3>1. Start from problems, not from ideas</h3>
<p>An idea invented in the abstract has to go looking for a problem afterwards, and usually doesn't find one. Reverse it. Keep a <strong>problem journal</strong> for two weeks: every time you, or anyone around you, mutters "why is this so hard", write it down — the errand that takes three apps, the form that gets printed and re-typed, the job everyone at work quietly dreads. Boring, repeated, mildly painful problems are the richest seam, precisely because nobody glamorous is looking at them.</p>

<h3>2. Hunt workarounds — they are demand, already proven</h3>
<p>Wherever people have built their own clumsy fix, someone has already voted with their time. The spreadsheet that runs an entire department. The WhatsApp group that functions as a booking system. The binder of printouts a tradesperson keeps in the van. A workaround is a customer telling you what they would buy, in the most credible language there is: effort they already spend.</p>

<h3>3. Take stock of your unfair advantages</h3>
<p>The same idea is weak from one person and strong from another. Before generating anything, write down honestly: what do you know that most people don't (a trade, an industry's plumbing, a regulation's real effect)? What do you have access to that others can't easily get (an audience, data, relationships, equipment, a licence)? What have you done repeatedly that others do once? The ideas worth <em>your</em> time sit at the crossing of a real problem and one of these — anyone can have the idea, but not anyone can do it.</p>

<h3>4. Generate in volume, choose in cold blood</h3>
<p>The first idea is rarely the best; it is just the first. Give yourself a rule — twenty candidates before you evaluate any — and use crossings to force variety: each problem from your journal × each advantage from your list. Most combinations will be nonsense. That is the method working, not failing: quantity first, judgement second, and never during.</p>

<h3>5. Shape each survivor into one testable sentence</h3>
<p>An idea you cannot state crisply cannot be tested. Force each one into: <em>"For [a specific person] who [has this problem], [the thing] does [what], better than [what they do today], because [your advantage]."</em> Writing that sentence is diagnostic — where it goes vague is exactly where the idea is vague. "Everyone" in the first slot is the classic warning sign; an idea for everyone is an idea for no one in particular, which is who buys it.</p>

<h3>6. Do not fall in love yet</h3>
<p>Affection for an idea grows with time spent on it, not with evidence for it — which is why the longer you polish one in private, the harder it becomes to hear a no. Hold several candidates loosely instead of one tightly, decide what evidence would promote or kill each, and let the next stage — building the smallest testable version — do the choosing.</p>

<h3>The whole thing, on one line</h3>
<p>Collect problems → spot the workarounds → know your unfair advantages → generate twenty before judging one → shape each into a testable sentence → and let evidence, not affection, pick the winner.</p>
$c$),
       'pending'
FROM pages p WHERE p.site_id='1244516d-014d-421c-88c6-090bb1e9552a' AND p.url='/guides/creating-ideas/index.html';

INSERT INTO page_components (page_id, component_id, slot_name, position, content_data, build_status)
SELECT p.id, '0197e8d7-1adc-43d6-ab32-d0716e013175', 'call-to-action', 3,
       jsonb_build_object(
         'headline', 'Got a candidate? Find out if anyone wants it',
         'subheadline', 'A Verified Idea Report researches one idea properly — the market, who else is there, and a specific next step — for £29. It will tell you plainly if the answer is no, which at this stage is the cheapest gift there is.',
         'primary_cta', 'Get a verified idea report', 'primary_cta_url', '/report.html',
         'secondary_cta', 'Next: building it', 'secondary_cta_url', '/guides/building-it/index.html'),
       'pending'
FROM pages p WHERE p.site_id='1244516d-014d-421c-88c6-090bb1e9552a' AND p.url='/guides/creating-ideas/index.html';

-- ═══════════════════════════════════════════════════════════════════════════
-- GUIDE 6 — Building it (stage 2), nav_order 3
-- ═══════════════════════════════════════════════════════════════════════════
INSERT INTO pages (site_id, name, url, title, page_type, status, meta_description, topics,
                   nav_label, nav_order, in_header, in_footer, build_status, sections)
VALUES ('1244516d-014d-421c-88c6-090bb1e9552a', 'guide-building-it',
        '/guides/building-it/index.html',
        'Building it: from idea to first working version',
        'guide', 'active',
        'The first version exists to teach you something, not to impress anyone. How to build the least you can learn from — often without writing code at all.',
        ARRAY['mvp','prototyping','building','startups'],
        'Building it', 3, false, false, 'planned', '[]'::jsonb);

INSERT INTO page_components (page_id, component_id, slot_name, position, content_data, build_status)
SELECT p.id, '23f95f00-f293-466e-b43a-81791ea0fc6c', 'hero', 1,
       jsonb_build_object(
         'headline', 'Building it: the least you can learn from',
         'subheadline', 'The first version is not a small product — it is a question, built. The mistake that eats the most time and money is building more than the question needs answered.',
         'cta_text', 'Get a verified idea report', 'cta_url', '/report.html',
         'secondary_cta', 'Next: testing it', 'secondary_cta_url', '/guides/testing-it/index.html'),
       'pending'
FROM pages p WHERE p.site_id='1244516d-014d-421c-88c6-090bb1e9552a' AND p.url='/guides/building-it/index.html';

INSERT INTO page_components (page_id, component_id, slot_name, position, content_data, build_status)
SELECT p.id, '8d81e665-3ee0-443d-a873-690268c15fbb', 'generic-text-block', 2,
       jsonb_build_object('heading', $h$Build the question, not the product$h$, 'content', $c$
<p><strong>The short version.</strong> The purpose of a first version is to learn whether the idea works — nothing else. Every hour spent on parts that don't serve that question is an hour spent decorating a hypothesis. Decide what you need to learn, then build the absolute least that can teach it — which, more often than people like, means not building software at all yet.</p>

<h3>1. Write down the question first</h3>
<p>Before anything gets made, finish this sentence: <em>"This version exists to find out whether ___."</em> Whether plumbers will photograph their paperwork. Whether parents will pay a deposit before the club exists. Whether the report is good enough that someone asks for a second one. If you cannot name the question, you are not building a first version — you are just building.</p>

<h3>2. The ladder of fakes — climb it before you code</h3>
<p>Each rung below answers real questions at a fraction of the cost of the rung above it:</p>
<ul>
  <li><strong>A landing page</strong> that describes the product and takes sign-ups — tests whether the promise lands, before anything exists.</li>
  <li><strong>Doing it by hand</strong> (the "concierge" version): deliver the service manually to a handful of customers. You learn what they actually need, and they pay you to teach you.</li>
  <li><strong>The hidden-human version</strong>: the customer sees a product; behind the curtain, it's you doing the work the software would do. Tests the experience without the engineering.</li>
  <li><strong>No-code and spreadsheet tools</strong>: form builders, automation glue, a database with a nice front. Embarrassingly effective for v1, and disposable by design.</li>
  <li><strong>Real code</strong> — the top rung, earned only when the rungs below have run out of things to teach you.</li>
</ul>

<h3>3. Scope like a coward</h3>
<p>One kind of user. One problem. One path through, working end to end. Everything else — settings, edge cases, the second user type, the admin screen — goes on a list marked <em>later</em>, and most of it will die there quietly once real users show you what actually matters. A narrow thing that works teaches more than a broad thing that nearly does.</p>

<h3>4. Buy the boring parts</h3>
<p>Payments, logins, email, hosting, scheduling — solved problems, rentable for pennies, and invisible to your idea's actual test. Building any of them yourself at this stage is procrastination wearing a hard hat. Your effort belongs in the one part that is genuinely yours.</p>

<h3>5. Beware polish — it is fear, dressed up</h3>
<p>"It's not ready to show people yet" usually means "I am not ready for people's verdict". The scrappy version shown this month beats the polished version shown next quarter, because only one of them produces learning. If showing it doesn't make you slightly uncomfortable, you waited too long.</p>

<h3>6. Two cautions worth carrying from the later guides</h3>
<p>If any part of the idea might be patentable, showing it publicly can cost you those rights — read the patents guide <em>before</em> the launch post, and keep the clever part behind the curtain. And if a freelancer builds any of this for you, get the copyright assigned in writing before they start — the copyright guide explains why the default answer is that they own it.</p>

<h3>The whole thing, on one line</h3>
<p>Name the question → climb the ladder of fakes before you code → one user, one problem, one path → buy the boring parts → ship scrappy and let real users write the rest of the spec.</p>
$c$),
       'pending'
FROM pages p WHERE p.site_id='1244516d-014d-421c-88c6-090bb1e9552a' AND p.url='/guides/building-it/index.html';

INSERT INTO page_components (page_id, component_id, slot_name, position, content_data, build_status)
SELECT p.id, '0197e8d7-1adc-43d6-ab32-d0716e013175', 'call-to-action', 3,
       jsonb_build_object(
         'headline', 'Before you build for months, research for days',
         'subheadline', 'A Verified Idea Report tells you what already exists, what people use instead, and the cheapest way to test demand — for £29, before the first line of code.',
         'primary_cta', 'Get a verified idea report', 'primary_cta_url', '/report.html',
         'secondary_cta', 'Next: testing it', 'secondary_cta_url', '/guides/testing-it/index.html'),
       'pending'
FROM pages p WHERE p.site_id='1244516d-014d-421c-88c6-090bb1e9552a' AND p.url='/guides/building-it/index.html';

-- ═══════════════════════════════════════════════════════════════════════════
-- GUIDE 7 — Testing it (stage 3), nav_order 4
-- ═══════════════════════════════════════════════════════════════════════════
INSERT INTO pages (site_id, name, url, title, page_type, status, meta_description, topics,
                   nav_label, nav_order, in_header, in_footer, build_status, sections)
VALUES ('1244516d-014d-421c-88c6-090bb1e9552a', 'guide-testing-it',
        '/guides/testing-it/index.html',
        'Testing it: honest experiments before you commit',
        'guide', 'active',
        'A test is a question with a pass mark you set in advance. How to run demand tests and user conversations that can actually say no.',
        ARRAY['validation','experiments','testing','startups'],
        'Testing it', 4, false, false, 'planned', '[]'::jsonb);

INSERT INTO page_components (page_id, component_id, slot_name, position, content_data, build_status)
SELECT p.id, '23f95f00-f293-466e-b43a-81791ea0fc6c', 'hero', 1,
       jsonb_build_object(
         'headline', 'Testing it: experiments that are allowed to say no',
         'subheadline', 'Most "validation" is a ritual performed to reach a yes that was already decided. A real test has a pass mark set in advance, and a real chance of failing.',
         'cta_text', 'Get a verified idea report', 'cta_url', '/report.html',
         'secondary_cta', 'Next: user acceptance', 'secondary_cta_url', '/guides/user-acceptance/index.html'),
       'pending'
FROM pages p WHERE p.site_id='1244516d-014d-421c-88c6-090bb1e9552a' AND p.url='/guides/testing-it/index.html';

INSERT INTO page_components (page_id, component_id, slot_name, position, content_data, build_status)
SELECT p.id, '8d81e665-3ee0-443d-a873-690268c15fbb', 'generic-text-block', 2,
       jsonb_build_object('heading', $h$A test is a question with a pass mark you set in advance$h$, 'content', $c$
<p><strong>The short version.</strong> The point of testing is not to gather encouragement — it is to make a decision. That requires three things most idea-stage "validation" quietly lacks: a specific hypothesis, a threshold chosen <em>before</em> you see the results, and a genuine willingness to act on a no.</p>

<h3>1. Write the bet down before you run it</h3>
<p><em>"We believe [these people] will [do this costly thing]. We'll test it by [this] for [this long]. If fewer than [number] do it, the idea as stated is wrong."</em> The number is the part people skip, because it is the part with teeth. Chosen afterwards, any result can be argued into a pass; chosen before, the test can actually teach you something.</p>

<h3>2. Test with costs, not compliments</h3>
<p>People are kind. They will say your idea is great and never think of it again — kind words are the fool's gold of idea-testing. Evidence is what people <em>give up</em> for the thing, and it comes in grades: paid money is the strongest; a deposit or pre-order next; time invested (a booked call kept, a form completed, a trial actually used) next; an email address is weak; a compliment is nothing. Design every test to ask for the highest-grade evidence you plausibly can. A "fake door" — a page selling the product as if it exists, with an honest "not ready yet, join the list" after the click — is often the cheapest way to ask strangers for a real signal.</p>

<h3>3. Interview about their life, never about your idea</h3>
<p>The moment you pitch, the data spoils — now they're being nice. Ask instead about what they <em>already do</em>: "When did this problem last come up? What did you do about it? What did that cost you? Have you tried to fix it — what happened?" Past behaviour is evidence; opinions about a hypothetical future are weather. If they have never tried to solve the problem, that is a finding too — it usually means the problem isn't painful enough to pay for.</p>

<h3>4. Small numbers are fine; vague questions are not</h3>
<p>You don't need statistics at this stage — five honest conversations with the right people will demolish or transform an idea faster than five hundred survey responses. What you need is sharpness: the right people (the ones from your one-sentence proposition, not your friends), questions about behaviour, and notes written down verbatim before memory rounds them up.</p>

<h3>5. Let it fail properly</h3>
<p>When a test fails, resist the two reflexes: softening the threshold ("nearly passed"), and re-running it until it passes. A failed test is a success of the method — it just saved you the months the build would have taken. The useful move is diagnostic: did the wrong people see it, was the promise unclear, or do the right people simply not care? The first two are fixable; the third is your answer.</p>

<h3>6. Keep a log</h3>
<p>One page per test: the bet, the threshold, what happened, what you decided. Memory is an unreliable narrator with a motive; three months in, the log is the only honest account of what you actually learned — and it is exactly what a future investor, partner or report will ask you for.</p>

<h3>The whole thing, on one line</h3>
<p>Write the bet and the pass mark first → ask for costs, not compliments → interview about their life, not your idea → let a failed test actually fail → and log it, because memory takes your side.</p>
$c$),
       'pending'
FROM pages p WHERE p.site_id='1244516d-014d-421c-88c6-090bb1e9552a' AND p.url='/guides/testing-it/index.html';

INSERT INTO page_components (page_id, component_id, slot_name, position, content_data, build_status)
SELECT p.id, '0197e8d7-1adc-43d6-ab32-d0716e013175', 'call-to-action', 3,
       jsonb_build_object(
         'headline', 'Want the research half done for you?',
         'subheadline', 'A Verified Idea Report covers the desk-research side — the market, the competition, the substitutes — and hands you a specific, affordable demand test to run. £29.',
         'primary_cta', 'Get a verified idea report', 'primary_cta_url', '/report.html',
         'secondary_cta', 'Next: user acceptance', 'secondary_cta_url', '/guides/user-acceptance/index.html'),
       'pending'
FROM pages p WHERE p.site_id='1244516d-014d-421c-88c6-090bb1e9552a' AND p.url='/guides/testing-it/index.html';

-- ═══════════════════════════════════════════════════════════════════════════
-- GUIDE 8 — User acceptance (stage 4), nav_order 5
-- ═══════════════════════════════════════════════════════════════════════════
INSERT INTO pages (site_id, name, url, title, page_type, status, meta_description, topics,
                   nav_label, nav_order, in_header, in_footer, build_status, sections)
VALUES ('1244516d-014d-421c-88c6-090bb1e9552a', 'guide-user-acceptance',
        '/guides/user-acceptance/index.html',
        'User acceptance: does it do the job they hired it for?',
        'guide', 'active',
        'Working software can still fail the only test that matters: whether real users, in their real environment, get the job done. How to run acceptance properly.',
        ARRAY['user acceptance','uat','pilots','startups'],
        'User acceptance', 5, false, false, 'planned', '[]'::jsonb);

INSERT INTO page_components (page_id, component_id, slot_name, position, content_data, build_status)
SELECT p.id, '23f95f00-f293-466e-b43a-81791ea0fc6c', 'hero', 1,
       jsonb_build_object(
         'headline', 'User acceptance: it works — but does it do the job?',
         'subheadline', 'A product can pass every technical check and still fail the only test that pays: whether real users, in their real setting, on their real tasks, get the job done and come back.',
         'cta_text', 'Get a verified idea report', 'cta_url', '/report.html',
         'secondary_cta', 'Next: feedback loops', 'secondary_cta_url', '/guides/feedback-loops/index.html'),
       'pending'
FROM pages p WHERE p.site_id='1244516d-014d-421c-88c6-090bb1e9552a' AND p.url='/guides/user-acceptance/index.html';

INSERT INTO page_components (page_id, component_id, slot_name, position, content_data, build_status)
SELECT p.id, '8d81e665-3ee0-443d-a873-690268c15fbb', 'generic-text-block', 2,
       jsonb_build_object('heading', $h$"It works" and "it's accepted" are different tests$h$, 'content', $c$
<p><strong>The short version.</strong> Technical testing asks "does it do what we built it to do?" Acceptance asks a harder question: "does it do the job the user hired it for, in their world, without you standing next to it?" Plenty of products pass the first and quietly fail the second — and users rarely tell you. They just stop coming back.</p>

<h3>1. Define "accepted" with the user, in their words</h3>
<p>Before any pilot, agree what success looks like — with the user, not about them. Not "no critical bugs" but <em>"I processed Friday's orders in this instead of the spreadsheet, and it took less time."</em> Written down, specific, theirs. If you can't get a user to state what would make them keep it, that is acceptance failing early — which is the cheap place for it to fail.</p>

<h3>2. Watch real tasks in the real environment</h3>
<p>A demo you drive tells you nothing; the phone rings, the wifi drops, the data is messier than your test data, and the user has nine other things open. Sit beside (or screen-share with) a handful of real users doing their <em>own</em> work with it, and say as little as possible. The moments that matter are the hesitations, the wrong turns, and the point where they reach for the old spreadsheet — each one is the product telling you where it doesn't fit their world yet.</p>

<h3>3. Structure a pilot like an experiment, not a favour</h3>
<p>The shapeless pilot — "have a play and tell us what you think" — produces polite noise. Give it edges: a few real users, their real work, a fixed period, the acceptance criteria from step 1, and a scheduled end conversation. And agree the exchange up front: they get early access and influence; you get honesty and a decision at the end — keep it, fix it, or drop it.</p>

<h3>4. The silence trap</h3>
<p>Users do not report problems; they route around them, exactly as they routed around the problem your idea addresses. Silence during a pilot is not approval — it is often abandonment you haven't noticed yet. Watch what they <em>do</em>: log-ins, completions, the tasks still being done the old way. A user who complains is engaged; the one who says nothing has usually already left.</p>

<h3>5. Keep a friction log, and let it set the roadmap</h3>
<p>Every hesitation, workaround and "oh, I didn't realise" goes in one list, dated and verbatim. At the end of the pilot this log — not your feature wish-list — is the roadmap. Fixing the top three frictions almost always beats adding the next feature, because the frictions are standing between existing users and the job, while the feature is a guess about hypothetical ones.</p>

<h3>6. When acceptance fails, the spec was wrong — not the user</h3>
<p>The reflex is to explain, train, or document harder. Occasionally that is right; usually it is the product insisting the world adapt to it. If several sensible people misuse it the same way, that <em>is</em> the design. The pilot did its job: it found the gap between what you built and the job to be done, while the gap was still cheap to close.</p>

<h3>The whole thing, on one line</h3>
<p>Agree "accepted" in the user's words → watch real tasks in the real environment → give the pilot edges and an end date → treat silence as a warning, not a pass → and let the friction log, not your wish-list, write the roadmap.</p>
$c$),
       'pending'
FROM pages p WHERE p.site_id='1244516d-014d-421c-88c6-090bb1e9552a' AND p.url='/guides/user-acceptance/index.html';

INSERT INTO page_components (page_id, component_id, slot_name, position, content_data, build_status)
SELECT p.id, '0197e8d7-1adc-43d6-ab32-d0716e013175', 'call-to-action', 3,
       jsonb_build_object(
         'headline', 'Acceptance proves the product. The report proves the market.',
         'subheadline', 'Before the pilot, it is worth knowing who else is out there and what your users compare you against. A Verified Idea Report researches one idea properly for £29.',
         'primary_cta', 'Get a verified idea report', 'primary_cta_url', '/report.html',
         'secondary_cta', 'Next: feedback loops', 'secondary_cta_url', '/guides/feedback-loops/index.html'),
       'pending'
FROM pages p WHERE p.site_id='1244516d-014d-421c-88c6-090bb1e9552a' AND p.url='/guides/user-acceptance/index.html';

-- ═══════════════════════════════════════════════════════════════════════════
-- GUIDE 9 — Feedback loops (stage 5), nav_order 6 — chains into Patents
-- ═══════════════════════════════════════════════════════════════════════════
INSERT INTO pages (site_id, name, url, title, page_type, status, meta_description, topics,
                   nav_label, nav_order, in_header, in_footer, build_status, sections)
VALUES ('1244516d-014d-421c-88c6-090bb1e9552a', 'guide-feedback-loops',
        '/guides/feedback-loops/index.html',
        'Feedback loops: improving on what users actually tell you',
        'guide', 'active',
        'Feedback is a system, not an inbox: collect behaviour as well as opinion, weight it by evidence, change something, and tell people you did. How to build the loop.',
        ARRAY['feedback','iteration','product','startups'],
        'Feedback loops', 6, false, false, 'planned', '[]'::jsonb);

INSERT INTO page_components (page_id, component_id, slot_name, position, content_data, build_status)
SELECT p.id, '23f95f00-f293-466e-b43a-81791ea0fc6c', 'hero', 1,
       jsonb_build_object(
         'headline', 'Feedback loops: hearing what users mean, not just what they say',
         'subheadline', 'Feedback is not an inbox to be emptied — it is a loop to be closed: collect what users do as well as what they say, decide with judgement, change something, and tell them you did.',
         'cta_text', 'Get a verified idea report', 'cta_url', '/report.html',
         'secondary_cta', 'Next: protecting it — patents', 'secondary_cta_url', '/guides/patents/index.html'),
       'pending'
FROM pages p WHERE p.site_id='1244516d-014d-421c-88c6-090bb1e9552a' AND p.url='/guides/feedback-loops/index.html';

INSERT INTO page_components (page_id, component_id, slot_name, position, content_data, build_status)
SELECT p.id, '8d81e665-3ee0-443d-a873-690268c15fbb', 'generic-text-block', 2,
       jsonb_build_object('heading', $h$The loop: collect → weigh → decide → change → tell them$h$, 'content', $c$
<p><strong>The short version.</strong> Most products don't lack feedback; they lack a loop. Comments arrive, are nodded at, and evaporate. A working loop has five parts — collect, weigh, decide, change, and close the loop by telling people — and the last part, the one almost everyone skips, is where users turn into advocates.</p>

<h3>1. Collect behaviour, not just opinion</h3>
<p>What users say is one channel, and the least reliable one. Run at least three: <strong>what they say</strong> (support messages, reviews, the pilot's friction log — kept verbatim, because paraphrase launders the pain out); <strong>what they do</strong> (which features get used, where people drop out, what never gets touched — even simple counts will do); and <strong>what leavers say</strong>. The short conversation with someone who stopped using it is the single most nutritious feedback there is, and nobody enjoys asking for it. Ask anyway: "what did you switch back to?" is one question, and the answer is your real competitor.</p>

<h3>2. Weigh feedback by evidence, not volume</h3>
<p>Feedback arrives pre-distorted: the loudest voices are not the typical ones, the most recent comment feels the most important, and the person who emails daily is not your market. Correct for it deliberately. Does this request come from people who match your target user? Is it echoed in the behaviour data? Would acting on it serve the job the product is hired for, or just quiet one voice? Three vague mentions of the same friction outweigh one eloquent demand for a feature.</p>

<h3>3. Interpret requests as symptoms</h3>
<p>Users are excellent at spotting problems and mediocre at prescribing solutions — the request "add an export button" often means "I don't trust your product to keep my data" or "my boss needs a copy". Before building what was asked, ask what the person was trying to do when they asked. Fixing the underlying job usually costs less and satisfies more people than the literal request.</p>

<h3>4. Decide on a cadence, in one place</h3>
<p>Feedback handled reactively becomes whoever-shouted-last. Put everything in one list, and review it on a rhythm — weekly is plenty early on — deciding each time: what gets fixed now, what waits, what is explicitly declined. Declining is a decision, not a failure; a product that says yes to everything becomes a junk drawer with a login.</p>

<h3>5. Close the loop out loud</h3>
<p>When feedback changes something, tell the people who raised it — personally if there are few, in a visible "what changed" note as you grow. This is the highest-return, most-skipped step in the entire loop: people who see their feedback land become the users who recruit others, forgive rough edges, and keep telling you the truth. Feedback into silence, meanwhile, simply stops coming — and it stops quietly.</p>

<h3>6. Know what feedback cannot decide</h3>
<p>The loop refines a direction; it does not choose one. Users pull products toward their existing habits — real improvement often lives a step beyond what any single request describes, and a chorus asking for a faster version of the old way can drown the few showing you the new one. When feedback and vision disagree, that is not noise: it is the moment to re-run a proper test from the testing guide, with a pass mark, and let evidence arbitrate.</p>

<h3>The whole thing, on one line</h3>
<p>Collect what they do as well as what they say → interview the leavers → weigh by evidence, not volume → treat requests as symptoms → decide on a rhythm → and always, always tell people what their feedback changed.</p>
$c$),
       'pending'
FROM pages p WHERE p.site_id='1244516d-014d-421c-88c6-090bb1e9552a' AND p.url='/guides/feedback-loops/index.html';

INSERT INTO page_components (page_id, component_id, slot_name, position, content_data, build_status)
SELECT p.id, '0197e8d7-1adc-43d6-ab32-d0716e013175', 'call-to-action', 3,
       jsonb_build_object(
         'headline', 'The loop is working. Now protect what it built.',
         'subheadline', 'Once feedback has shaped something people genuinely want, the next questions are protection and money — patents, copyright, and funding. The guides continue there; and if you are still weighing the idea itself, the Verified Idea Report is £29.',
         'primary_cta', 'Get a verified idea report', 'primary_cta_url', '/report.html',
         'secondary_cta', 'Next: protecting it — patents', 'secondary_cta_url', '/guides/patents/index.html'),
       'pending'
FROM pages p WHERE p.site_id='1244516d-014d-421c-88c6-090bb1e9552a' AND p.url='/guides/feedback-loops/index.html';

-- ---------------------------------------------------------------------------
-- pages.sections backfill for all five (the rerender path never writes it).
-- ---------------------------------------------------------------------------
UPDATE pages p
SET sections = (
      SELECT COALESCE(jsonb_agg(pc.slot_name ORDER BY pc.position), '[]'::jsonb)
      FROM page_components pc WHERE pc.page_id = p.id AND COALESCE(pc.slot_name,'') <> ''
    ),
    updated_at = now()
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url IN ('/guides/creating-ideas/index.html','/guides/building-it/index.html',
                '/guides/testing-it/index.html','/guides/user-acceptance/index.html',
                '/guides/feedback-loops/index.html');

DO $guard2$
DECLARE nbad int; ntot int; nslot int; npages int;
BEGIN
  SELECT count(DISTINCT p.id),
         count(*),
         count(*) FILTER (WHERE pc.content_data IS NULL OR pc.content_data = '{}'::jsonb),
         count(*) FILTER (WHERE COALESCE(pc.slot_name,'') = '')
    INTO npages, ntot, nbad, nslot
  FROM page_components pc JOIN pages p ON p.id = pc.page_id
  WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
    AND p.url IN ('/guides/creating-ideas/index.html','/guides/building-it/index.html',
                  '/guides/testing-it/index.html','/guides/user-acceptance/index.html',
                  '/guides/feedback-loops/index.html');
  IF npages <> 5 OR ntot <> 15 THEN RAISE EXCEPTION 'ABORT: expected 5 pages x 3 sections, got % pages / % sections.', npages, ntot; END IF;
  IF nbad > 0 THEN RAISE EXCEPTION 'ABORT: % section(s) empty content_data.', nbad; END IF;
  IF nslot > 0 THEN RAISE EXCEPTION 'ABORT: % section(s) missing slot_name.', nslot; END IF;
END
$guard2$;

COMMIT;

SELECT p.id AS page_id, p.url, p.nav_order, p.build_status
FROM pages p
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a' AND p.page_type='guide'
ORDER BY p.nav_order;

-- SQL_2026-08-20_content_batch_week1.sql
--
-- WEEK 1 of the content cadence from PLAN_2026-08-18 §3. Owner instruction 2026-08-20:
-- "Go ahead with your choice of blogs."
--
-- WHY FOUR AND NOT FIVE. growth_config allows 5 blog posts per ROLLING 7 days
-- (page_growth_budget.go:162 counts `created_at > NOW() - INTERVAL '7 days'`), and ONE is
-- already spent this window — `darts-calendar-density`, filed by the news_editorial lane, who
-- told me so directly. Filing five would put the last one into `weekly_blog_limit_reached` and
-- look like a failure rather than a budget. [MEASURED 2026-08-20: blog-post 1, content 1]
--
-- WHY THESE FOUR. Division of labour agreed with the news_editorial lane in writing: their
-- lane takes timely/editorial keyed to a live story (`/insights/`), this lane takes evergreen
-- spec content (`/blog/`). They confirmed no overlap with this list.
--
-- Against the 12 existing posts, these are the gaps in the site's own subject. The site
-- explains weight, tungsten %, shaft length, flight shape and grip pattern — and says nothing
-- about barrel PROFILE, balance POINT, or points/tips. Three of these four complete the
-- anatomy of a dart; the fourth is the one high-intent reference the site lacks entirely.
--   deliberately NOT chosen: board dimensions (board-setup already covers "Height, Distance &
--   Mounting"), a beginners piece (beginners exists), a brand piece (brand-comparison exists).
--
-- ⚠ TOPIC DEMAND IS [UNMEASURED]. Nobody here has keyword volume, and the plan says so rather
-- than implying otherwise. Search Console query data is the honest input and the Cloudflare
-- token still lacks Analytics:Read. If the first four earn no impressions, that is information,
-- and the next four should be chosen from GSC's query list rather than from this reasoning.
--
-- ROUTE: needs_content_planning -> content-gap-planner with `approach: new_page` asserted,
-- exactly as SQL_2026-07-29n (which built /news/index.html) and the privacy page on 08-17 did.
-- `approach` is read from the plan the LLM returns (apply_gap_plan_action.go:127), so asserting
-- it is a strong steer, not a guarantee — check which branch actually ran.
--
-- STATUS 'triaged', never 'detected': detected does not drain on this site.
--
-- ⚠ THE TRAP FOR WHOEVER ILLUSTRATES THESE LATER: `article-body` holds figure and prose in ONE
-- llm-owned field, so the next body rewrite silently deletes any image placed in it. Four of
-- this site's guide figures were destroyed that way and recovered from page_component_history.
-- Do not add in-body imagery to these pages expecting it to survive.
--
-- Site: dartsonline.com  5fe8785b-223d-41a3-88ee-c07187622381

INSERT INTO site_work_items (
  site_id, source, pipeline, item_type, severity, summary, spec,
  priority, handler_agent, status, created_by, item_key, approval_mode
)
SELECT
  '5fe8785b-223d-41a3-88ee-c07187622381', 'dartsonline-traffic-workstream', 'content',
  'needs_content_planning', 'medium', v.summary,
  jsonb_build_object(
    'check',       'content_cadence_week1',
    'approach',    'new_page',
    'category',    'content_completeness',
    'page_name',   v.page_name,
    'page_type',   'blog-post',
    'suggestion',  'Create a new page named ''' || v.page_name || ''' with page_type ''blog-post'' (that exact literal — it is a routing key, and blog-post is also what puts this page on claims.go:752''s editorial-exempt list), url ''/blog/' || v.page_name || '.html'', sections ["hero","article-body","call-to-action"], in_header=false, in_footer=false — matching all 12 existing guides on this site exactly. It is a child page under /blog/, so it belongs to the Guides index and takes no nav membership of its own. Title in the house pattern: "' || v.title || '". ' || v.brief,
    'description', v.description
  ),
  60, 'content-gap-planner', 'triaged', 'dartsonline-traffic-workstream',
  'content_cadence_w1_' || v.page_name || ':5fe8785b-223d-41a3-88ee-c07187622381', 'auto'
FROM (VALUES
  ('barrel-shapes',
   'No guide covers barrel PROFILE — the site explains weight, tungsten and grip but not shape',
   'Dart Barrel Shapes Explained — Straight, Torpedo, Bomb and Scallop | Darts Online',
   'Cover the common barrel profiles (straight, torpedo/pear, bomb, scallop/waisted) and what each one does to the grip point, the balance and how forgiving the release is. Follow the site''s own rule: explain what the spec CHANGES about the throw, never just what it is. Link internally to the barrel weight guide, the grip styles guide and the Dart Setup Builder tool. Recommend, never prescribe — profile preference is personal.',
   'The site covers barrel weight, tungsten percentage, shaft length, flight shape and grip pattern, and has nothing at all on barrel profile — which is one of the four specs a player actually chooses between when buying. [MEASURED 2026-08-20: 12 blog-posts, none about shape.]'),
  ('dart-balance',
   'Nothing on the site explains balance point, which decides how the other specs behave',
   'Front, Centre or Rear Weighted — How Dart Balance Changes Your Throw | Darts Online',
   'Explain front-weighted, centre-weighted and rear-weighted barrels: where the mass sits, what each does to the nose-down attitude in flight, and how it interacts with shaft length and flight size. This is the piece that ties the existing guides together, so link to barrel weight, shaft length and flight shapes. State clearly that balance is a property of the barrel''s machining, not an add-on.',
   'Balance point is the spec that determines whether the other specs behave as the player expects, and the site does not mention it. It is also the natural hub linking four existing guides, which the site currently has no page doing.'),
  ('dart-points',
   'Points/tips are covered only as steel-vs-soft — nothing on point length, grip or finish',
   'Dart Points Explained — Length, Grip and Smooth vs Knurled | Darts Online',
   'Cover steel points specifically: standard lengths, smooth versus knurled/grooved finish, what each does to bounce-outs and to board wear, and when a longer point helps or hurts. Distinguish clearly from the existing steel-tip vs soft-tip guide, which is about the two DISCIPLINES rather than the component — link to it rather than repeating it. Mention that points are replaceable on most modern barrels without claiming any specific brand does or does not allow it.',
   'The only page touching points is steel-tip-vs-soft-tip, which compares disciplines, not the component. A player choosing or replacing points has nothing to read here.'),
  ('checkout-chart',
   'The site has no scoring or finishing reference at all, despite ''checkout'' being a house key term',
   'Darts Checkout Chart — Which Doubles to Leave and Why | Darts Online',
   'A practical finishing reference: the standard checkout routes from 170 down, which doubles are worth leaving and why (bed size, adjacent-segment risk, the reason 32 beats 40 for many players), and the two-dart and three-dart splits worth memorising. Keep the site''s register — enthusiast to enthusiast, second person. NO betting, odds or tipping content of any kind (owner decision D1, 2026-07-29, in content_direction.editorial.out_of_scope). Do not invent a professional''s preference or attribute a route to a named player without a source; describe the geometry, which needs no citation. Link to the tools and to the beginners guide.',
   'Finishing and checkouts are the highest-intent evergreen question in the sport and the site answers none of it, while listing ''checkout'' among its own key terms in content_direction.terminology. It is also the only page in this batch that a club player would bookmark, which is why it is in an otherwise gear-led set. Demand is [UNMEASURED] — no keyword data exists here.')
) AS v(page_name, summary, title, brief, description)
ON CONFLICT DO NOTHING;

SELECT left(spec->>'page_name',20) AS page_name, status, priority, item_key
FROM site_work_items
WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND spec->>'check'='content_cadence_week1'
ORDER BY 1;

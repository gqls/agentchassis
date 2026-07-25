-- p4_01_guide_patents.sql — idea.uk: the FIRST real guide, "Patents: how to protect an idea in the UK".
--
-- OWNER ASK (2026-07-24, features_open/014): "I'd like a section on patents and what to do with
-- ideas ... a whole set of guides and tools from helping users create ideas through to building,
-- testing, user acceptance, feedback loops, patents, copyright, funding ways, funding sources".
-- Patents is the owner's LEAD stage. This is increment 1 of that pipeline.
--
-- STATE BEFORE (verified 2026-07-25): idea.uk has 9 pages; /guides/index.html serves 200 but its
-- listing section is EMPTY (component `content-listing`, a STATIC `articles` array, 601 bytes
-- rendered). No page anywhere on the site has page_type='guide'. So this is the site's first guide.
--
-- WHY THIS SHAPE (grounded, not invented):
--   * Fleet precedent — gamesdesign.co.uk (5), relojistas.com (4), vetcomparison.uk (3) all use
--     url='/guides/<slug>/index.html', page_type='guide', hero + Generic Text Block.
--   * The guides HUB auto-populates from the page set: `guide-list_pre_037`
--     (9d5e461a-8981-4ecc-b236-05895edfc15d) resolves `items` from
--     `query.pages_where_type:guide` (queryresolve.go:81, resolvePagesWhereType ~:352). Eligibility
--     is FetchablePageEligibilitySQL — `deployed_at IS NOT NULL OR build_status='deployed'` — so the
--     guide must SHIP BEFORE the hub is switched over, or the hub resolves an empty required list.
--     Hub swap is therefore a SEPARATE script (p4_02), run after this one is verified live.
--   * `Generic Text Block` renders {{.content}} UNESCAPED (verified against the live
--     /guides/rng-design/index.html artefact), so authored HTML in content_data is safe.
--
-- WHY THE CONTENT IS AUTHORED, NOT LLM-GENERATED (deliberate, record the reason):
--   Patent law is high-stakes factual/legal content on a live commercial site. The platform's
--   evidence gate for factual claims (claims-verification V5, citation source kind) is BUILT BUT
--   INERT, and bugs_open/043 is a live fabricated-statistics lane. An LLM content pass here would
--   be exactly the fabrication shape that lane exists for. So every claim below is authored,
--   UK-specific, hedged where it should be, and the sections are LOCKED (p4_03) after verification
--   so a later content-writer pass cannot silently rewrite it.
--
-- ONE CORRECTION MADE WHILE DRAFTING (recorded because it is the sort of thing that ships wrong):
--   The first draft said the IPEC SMALL CLAIMS track makes patent enforcement affordable. It does
--   not — IPEC's small claims track expressly does NOT hear patent, registered design or
--   semiconductor topography claims. Corrected to the IPEC multi-track (costs/damages caps) plus
--   the IPO's non-binding opinions service.
--
-- ORDER OF OPERATIONS:
--   p4_01 (this)  create page + 3 sections with authored content_data, build_status='planned'
--   -> direct-fire page-rerender with input_data.spec.reason='section_data_resolved'
--      (the queue is unreliable — bugs_open/029/030; the 049b kcat pattern + spec.reason is proven
--      on this site, RUNNING_NOTES §X.9). NOT assemble-only: a new page has no rendered_html yet.
--   -> verify LIVE at https://idea.uk/guides/patents/index.html (curl the page, not the job status)
--   p4_02          swap the guides hub to guide-list + rerender  (only after the guide is live)
--   p4_03          lock the authored sections

\set ON_ERROR_STOP on

BEGIN;

-- ---------------------------------------------------------------------------
-- 0. Guards: refuse to run twice, and refuse if the components we target moved.
-- ---------------------------------------------------------------------------
DO $guard$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM pages
   WHERE site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
     AND url = '/guides/patents/index.html';
  IF n > 0 THEN
    RAISE EXCEPTION 'ABORT: /guides/patents/index.html already exists on idea.uk (% row(s)) — p4_01 already ran.', n;
  END IF;

  SELECT count(*) INTO n FROM content_components
   WHERE id IN ('23f95f00-f293-466e-b43a-81791ea0fc6c',   -- hero
                '8d81e665-3ee0-443d-a873-690268c15fbb',   -- Generic Text Block
                '0197e8d7-1adc-43d6-ab32-d0716e013175')   -- call-to-action
     AND is_active;
  IF n <> 3 THEN
    RAISE EXCEPTION 'ABORT: expected 3 active target components, found % — ids have moved, re-ground before running.', n;
  END IF;
END
$guard$;

-- ---------------------------------------------------------------------------
-- 1. The page row.
--    nav_order 10 so it sorts first in the hub while it is the only guide
--    (resolvePagesWhereType orders by COALESCE(nav_order,100), name).
--    in_header/in_footer FALSE — guides are reached through the hub, matching
--    every other guide page in the fleet; the site nav already carries "Guides".
-- ---------------------------------------------------------------------------
INSERT INTO pages
  (site_id, name, url, title, page_type, status, meta_description, topics,
   nav_label, nav_order, in_header, in_footer, build_status, sections)
VALUES
  ('1244516d-014d-421c-88c6-090bb1e9552a',
   'guide-patents',
   '/guides/patents/index.html',
   'Patents: how to protect an idea in the UK',
   'guide',
   'active',
   'A plain-English UK guide to patents — why disclosing early can destroy your rights, what can be patented, what it costs, and when it is worth it.',
   ARRAY['patents','intellectual property','protecting ideas','uk law'],
   'Patents',
   10,
   false,
   false,
   'planned',
   '[]'::jsonb);

-- ---------------------------------------------------------------------------
-- 2. Section 1 — hero. Funnels to the paid tool, consistent with the p3_05/p3_07
--    funnel decision (every primary CTA on this site now points at /report.html).
--    cta_url / secondary_cta_url are source=renderer, but contextToInterfaceMap
--    defaults them and the ContentData merge OVERRIDES (proven by p3_05 live).
-- ---------------------------------------------------------------------------
INSERT INTO page_components (page_id, component_id, position, content_data, build_status)
SELECT p.id, '23f95f00-f293-466e-b43a-81791ea0fc6c', 1,
       jsonb_build_object(
         'headline',          'Patents: what to do with an idea you think is worth protecting',
         'subheadline',       'The order you do things in matters more than the paperwork — and one very ordinary mistake, made in a pitch, can end the conversation permanently. A plain-English guide to UK patents: what they cover, what they cost, and when they are worth it.',
         'cta_text',          'Get a verified idea report',
         'cta_url',           '/report.html',
         'secondary_cta',     'All guides',
         'secondary_cta_url', '/guides/index.html'
       ),
       'pending'
FROM pages p
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/guides/patents/index.html';

-- ---------------------------------------------------------------------------
-- 3. Section 2 — the guide body (authored; see the header note on why).
-- ---------------------------------------------------------------------------
INSERT INTO page_components (page_id, component_id, position, content_data, build_status)
SELECT p.id, '8d81e665-3ee0-443d-a873-690268c15fbb', 2,
       jsonb_build_object(
         'heading', $h$Protecting an idea in the UK, in the order that actually matters$h$,
         'content', $c$
<p><strong>The short version.</strong> An idea on its own cannot be patented. A patent protects a specific technical <em>invention</em>, it is expensive, and it is only worth having if you could afford to enforce it. Most ideas are better served by moving quickly, keeping the crucial part confidential, and relying on copyright and brand. But if yours is one that should be patented, there is a single mistake that ends the conversation permanently — and you can make it this week, by accident, in a pitch.</p>

<h3>1. Do not tell anyone. This is the mistake that cannot be undone.</h3>
<p>To be patentable in the UK an invention must be <strong>new</strong> — new against everything made available to the public anywhere in the world before the date you file. That includes things <em>you</em> did. A demo at a meetup, a pitch to investors without an NDA, a crowdfunding page, a conference talk, a paper, a product listing, a video, a forum post explaining how it works: any of these can become prior art against your own application.</p>
<p>The UK and Europe have <strong>no general grace period</strong>. The United States does allow a limited one, which is why you will find confident advice online saying early disclosure is survivable — it was written for a different country. There are narrow exceptions here (broadly, disclosure obtained in breach of confidence, or display at certain officially recognised international exhibitions, each within six months), but they are exceptions, not a plan.</p>
<p><strong>What to do instead:</strong> talk under a signed non-disclosure agreement, or talk about the problem and the outcome without disclosing how it works. Patent attorneys are bound by professional confidentiality, so you can speak freely to one before you file anything.</p>

<h3>2. Know what a patent is, and what it is not</h3>
<p>A UK patent is a time-limited monopoly — up to twenty years, with renewal fees payable to keep it alive — over a specific invention, defined by the claims. To be granted it must be:</p>
<ul>
  <li><strong>New</strong> — not already available to the public anywhere in the world.</li>
  <li><strong>Inventive</strong> — not obvious to someone skilled in that field. This is where most applications actually fail.</li>
  <li><strong>Capable of industrial application</strong> — it can be made or used in some kind of industry.</li>
  <li><strong>Not excluded</strong> — and the exclusions are broad.</li>
</ul>
<p>The Patents Act 1977 excludes, "as such", discoveries and scientific theories, mathematical methods, literary and artistic works, schemes or rules for performing a mental act, playing a game or <strong>doing business</strong>, presentations of information, and <strong>programs for a computer</strong>. Methods of treating or diagnosing humans or animals are excluded separately.</p>
<p>Those last two catch a great many good ideas. "A better way to run a marketplace" is a business method. "An app that does X" is, on its own, a computer program. The words "as such" are doing real work, though: software producing a <em>technical</em> effect beyond the ordinary running of a computer — controlling a machine, improving how the computer itself works, processing a signal — can be patentable. Precisely where that line falls is genuinely contested, and it is the question to put to a patent attorney rather than to a search engine.</p>
<p>Two things a patent is <strong>not</strong>. It is not permission to sell your product — somebody else's patent can still block you. And it is not enforcement: nobody polices it on your behalf. A patent is a right to go to court at your own expense.</p>

<h3>3. Decide honestly whether it is worth it</h3>
<p>The uncomfortable question is not "can I patent this?" but "if a well-funded competitor copies me, will I actually sue them?" Patent litigation in the High Court is very expensive. The Intellectual Property Enterprise Court (IPEC) exists to make IP disputes more affordable, with caps on recoverable costs and on damages — but note that IPEC's cheapest route, the small claims track, does <strong>not</strong> hear patent claims, so a patent dispute means the IPEC multi-track or the High Court. The IPO's non-binding opinions service is a comparatively cheap way to get a view on validity or infringement before anyone starts litigating.</p>
<p>A patent tends to earn its cost when:</p>
<ul>
  <li>the advantage is <strong>technical</strong>, and visible in a competitor's product — so you could tell you had been infringed;</li>
  <li>the market is big enough that a monopoly is worth more than the bill;</li>
  <li>you need it as an <strong>asset</strong>, because investors, licensees or an acquirer are pricing it in;</li>
  <li>the invention is genuinely hard to design around.</li>
</ul>
<p>It tends not to be worth it when your advantage is really execution, brand, data, distribution or speed; when the invention would be obsolete before grant; when you could never detect infringement; or when publishing the method — which is what a patent application does at eighteen months — hands competitors a recipe they did not previously have.</p>
<p>Keeping it secret is a legitimate alternative, not a failure. A trade secret never expires. It also protects you not at all the moment it gets out, and not at all against someone who invents the same thing independently.</p>

<h3>4. Do a first-look search before you spend anything</h3>
<p>You will not do this as well as a professional searcher, but an hour here regularly saves thousands. Free places to look:</p>
<ul>
  <li><strong>Espacenet</strong>, from the European Patent Office — worldwide coverage, the serious one.</li>
  <li><strong>Google Patents</strong> — friendlier full-text search, good for a first sweep.</li>
  <li>The <strong>UK Intellectual Property Office</strong> search, and Ipsum for the file history of UK cases.</li>
</ul>
<p>Search the <em>function</em>, not your product name: describe what the thing does, in several different vocabularies, because patent specifications are written in deliberately generic language. And remember prior art is not only patents — a product on sale, a manual, a thesis, a video or a blog post all count. If you find something close, that is a good outcome. You have just saved yourself the cost of finding out later.</p>

<h3>5. Understand the timeline — the first year is the useful part</h3>
<ul>
  <li><strong>Day 0 — file.</strong> Your filing date becomes your priority date. That is the moment the invention has to be new by, which is why filing before disclosing is the whole game.</li>
  <li><strong>The next twelve months — the priority year.</strong> This is the genuinely valuable bit. You can now talk about it, test the market, sell and raise money, while deciding whether it is worth taking abroad. Applications filed in other countries within this window can claim your original date.</li>
  <li><strong>Around eighteen months — publication.</strong> Your application is published. From then on anyone can read how it works, whether or not it is ever granted.</li>
  <li><strong>Search and examination</strong> are requested separately, each on its own deadline. Examination is where the real argument happens.</li>
  <li><strong>Grant</strong> commonly takes several years from filing, though it can be accelerated if you have a reason.</li>
  <li><strong>Renewal fees</strong> begin from the fifth year and rise annually. A patent you stop paying for lapses.</li>
</ul>
<p>There is <strong>no such thing as a world patent</strong>. The Patent Cooperation Treaty (PCT) gives you one international application that preserves your options across more than 150 states and defers the expensive decisions to roughly thirty months from priority — but you still have to file, argue and pay in each country or region you actually want.</p>

<h3>6. What it costs, honestly</h3>
<p>There are two separate bills. The <strong>official fees</strong> payable to the IPO — filing, search, examination — are modest, in the low hundreds of pounds for a UK application, and cheaper if you file online. The <strong>professional fees</strong> are the real cost: having a patent attorney draft and file a UK application typically runs to several thousand pounds, because drafting the claims well <em>is</em> the work, and going international afterwards moves into tens of thousands.</p>
<p>Fees and thresholds change. Check the current figures on the IPO's own website rather than trusting any article, including this one.</p>
<p>You are entitled to file your own application, and people do. But a badly drafted specification usually cannot be repaired later — you are not permitted to add new matter after filing — so a cheap application that claims the wrong thing is often worse than no application at all, because it has also destroyed the novelty of the thing it failed to claim.</p>

<h3>7. The alternatives, which are frequently the better answer</h3>
<ul>
  <li><strong>Copyright</strong> — automatic in the UK, no registration, no fee. It protects the <em>expression</em>: your source code, text, drawings, designs, music. It does not protect the underlying idea, and it does not stop somebody writing their own code that does the same job. For software it generally runs for the author's life plus seventy years.</li>
  <li><strong>Registered designs</strong> and <strong>unregistered design right</strong> — the appearance and shape of a product. Registration is inexpensive next to a patent.</li>
  <li><strong>Trade marks</strong> — your name and logo, registered against the classes of goods and services you actually sell. Often the most valuable thing a small business owns.</li>
  <li><strong>Confidentiality</strong> — NDAs, controlled disclosure, keeping the crucial part on your own server. It costs nothing, and it is the default until you decide otherwise.</li>
</ul>

<h3>8. Who to talk to</h3>
<p>A <strong>registered patent attorney</strong> is the person you want. It is a distinct profession from a solicitor — technically qualified and separately regulated. The Chartered Institute of Patent Attorneys (CIPA) publishes a searchable directory, and the register itself is maintained by IPReg. Many offer a free initial conversation, and everything you tell them is confidential. The IPO also publishes free guidance and business support, including an IP Health Check.</p>

<h3>The whole thing, on one line</h3>
<p>Say nothing publicly → search the prior art → decide honestly whether a patent earns its cost → talk to a patent attorney → file → then use the priority year to find out whether anybody actually wants it.</p>

<p><em>This guide is general information about UK law and practice, written to help you ask better questions. It is not legal advice, it does not create a professional relationship, and it cannot take account of your circumstances. Patent law is jurisdictional and unforgiving about deadlines — take advice from a registered patent attorney before relying on any of it, and check current fees, timescales and procedure with the Intellectual Property Office.</em></p>
$c$
       ),
       'pending'
FROM pages p
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/guides/patents/index.html';

-- ---------------------------------------------------------------------------
-- 4. Section 3 — the funnel. Honest about what the report is (it is research and
--    analysis, explicitly NOT professional advice — see the live /report.html copy),
--    so the CTA does not promise legal help the product does not deliver.
-- ---------------------------------------------------------------------------
INSERT INTO page_components (page_id, component_id, position, content_data, build_status)
SELECT p.id, '0197e8d7-1adc-43d6-ab32-d0716e013175', 3,
       jsonb_build_object(
         'headline',          'Before you spend anything on protecting it, find out whether anyone wants it',
         'subheadline',       'A patent is expensive and slow. Evidence of demand is cheaper and faster, and it is what tells you whether the patent is worth having. A Verified Idea Report is research and honest assessment on one idea — market evidence, who else is doing it, and a specific next step. £29. It is not legal advice.',
         'primary_cta',       'Get a verified idea report',
         'primary_cta_url',   '/report.html',
         'secondary_cta',     'Try the free idea check first',
         'secondary_cta_url', '/tools.html#audience-check'
       ),
       'pending'
FROM pages p
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/guides/patents/index.html';

-- ---------------------------------------------------------------------------
-- 5. Guard: every section must carry non-empty content_data, or a
--    section_data_resolved rerender escalates the page to the LLM content writer
--    and generates the legal content we deliberately authored by hand.
-- ---------------------------------------------------------------------------
DO $guard2$
DECLARE n int; total int;
BEGIN
  SELECT count(*) FILTER (WHERE pc.content_data IS NULL OR pc.content_data = '{}'::jsonb),
         count(*)
    INTO n, total
  FROM page_components pc JOIN pages p ON p.id = pc.page_id
  WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
    AND p.url = '/guides/patents/index.html';
  IF total <> 3 THEN
    RAISE EXCEPTION 'ABORT: expected 3 sections, inserted %.', total;
  END IF;
  IF n > 0 THEN
    RAISE EXCEPTION 'ABORT: % section(s) have NULL/empty content_data.', n;
  END IF;
END
$guard2$;

COMMIT;

-- Read-back: the ids the direct-fire needs.
SELECT p.id AS page_id, p.site_id, p.url, p.page_type, p.build_status, p.nav_order,
       (SELECT count(*) FROM page_components pc WHERE pc.page_id = p.id) AS n_sections
FROM pages p
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/guides/patents/index.html';

SELECT pc.position, cc.name AS component,
       length(pc.content_data::text) AS cd_bytes,
       coalesce(length(pc.rendered_html),0) AS rendered_len
FROM page_components pc
JOIN pages p ON p.id = pc.page_id
JOIN content_components cc ON cc.id = pc.component_id
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/guides/patents/index.html'
ORDER BY pc.position;

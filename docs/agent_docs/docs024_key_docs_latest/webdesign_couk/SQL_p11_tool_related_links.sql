-- SQL_p11_tool_related_links.sql — webdesign.co.uk
-- Persist the tool-page related-link blocks into page_components.rendered_html.
--
-- WHY. The served files already carry these (sites repo df51bfd91). The DB copy
-- is what assemble republishes, so without this the next assemble would serve the
-- OLD html and silently drop every block — the standing landmine recorded in
-- SQL_p10: "assemble republishes STORED rendered_html".
--
-- IDEMPOTENT. Each statement no-ops if the block is already present.
-- The anchor is the component's trailing </section> (it wraps .ported-page).
\set ON_ERROR_STOP on
BEGIN;

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/algorithms/p-values-explained.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Don&#39;t Be Fooled by Randomness</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Conversions went from 5% to 6% after you changed a button. Was it the button, or was it a lucky Tuesday?</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/bayesian-rank/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Bayesian Ranking</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Fair sorting for star ratings.</p></a>
    <a href="/tools/recommender-engine/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Recommender Engine</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Simulate Cosine Similarity vectors.</p></a>
    <a href="/tools/performance-budget/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Performance Budget</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Calculate Max JS/CSS size from latency.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/ab-test-calculator/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/operations/maintenance.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">The Hidden Cost of AI: Three Maintenance Problems</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">An AI builder optimises for immediate visual feedback, not long-term readability. Maintaining the result is a separate skill.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/favicon-maker/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Favicon Maker</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Convert Emojis or Logos to ICO/PNG.</p></a>
    <a href="/tools/social-card/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Social Card Generator</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate Open Graph meta tags instantly.</p></a>
    <a href="/tools/noise-generator/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">SVG Noise Generator</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate a sub-1KB feTurbulence filter as a data URI, instead of shipping a 1MB noise PNG.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/animated-favicon/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/accessibility/focus-states.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">The Invisible Focus</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Every junior designer eventually asks whether that blue outline can go. Here is why it cannot, and what to do instead.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/focus-ring/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Focus Ring Architect</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate premium, accessible double-layer focus states.</p></a>
    <a href="/tools/touch-target/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Touch Target Simulator</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Visualise 44px hit areas on mobile buttons.</p></a>
    <a href="/tools/seo-schema/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Semantic Schema</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Create JSON-LD markup for SEO.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/aria-builder/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/design/css-grid-math.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Fractional Layouts</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Percentages broke the moment you added a margin. The fr unit exists because of that arithmetic.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/golden-ratio/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Golden Ratio Cropper</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Compose perfect images using the Fibonacci spiral.</p></a>
    <a href="/tools/grid-generator/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">CSS Grid Architect</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Visual layout builder for columns &amp; gaps.</p></a>
    <a href="/tools/layout-generator/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Holy Grail Layouts</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate responsive CSS Grid code instantly.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/aspect-ratio/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/operations/maintenance.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">The Hidden Cost of AI: Three Maintenance Problems</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">An AI builder optimises for immediate visual feedback, not long-term readability. Maintaining the result is a separate skill.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/html-minifier/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">HTML Minifier</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Strip whitespace and comments.</p></a>
    <a href="/tools/image-optimizer/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Image Optimizer</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Convert PNG to JPG and resize locally.</p></a>
    <a href="/tools/head-architect/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">HTML Head Architect</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate SEO tags &amp; JSON-LD identity.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/asset-formatter/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/algorithms/bayesian-theory.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">The Bayesian Truth: Why 5.0 &lt; 4.8</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Sorting by average star rating seems obvious, and it is wrong. What one five-star review actually tells you.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/ab-test-calculator/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">A/B Significance</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Calculate P-Values and Z-Scores.</p></a>
    <a href="/tools/recommender-engine/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Recommender Engine</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Simulate Cosine Similarity vectors.</p></a>
    <a href="/tools/community-growth/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Growth Simulator</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Visualize Viral Coefficient (k) vs Churn.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/bayesian-rank/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/design/mood-boarding.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">The Messy Mood Board</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Stop overthinking. Why the secret to faster design workflows is embracing chaos and &#34;low-fidelity&#34; pasting.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/magic-outliner/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Magic Wand Outliner</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Click any region of a flat image to trace its border with a flood-fill and contour trace.</p></a>
    <a href="/tools/white-balance/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Batch White Balance</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Warm up or cool down a set of images instantly.</p></a>
    <a href="/tools/image-optimizer/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Image Optimizer</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Convert PNG to JPG and resize locally.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/bg-remover/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/design/svg-patterns.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Algorithmic Textures</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Stop using heavy PNGs. Why mathematical SVG patterns scale infinitely for free.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/mesh-gradient/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Aurora Mesh</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate trendy, moving mesh gradients.</p></a>
    <a href="/tools/svg-patterns/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">SVG Pattern Engine</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate tiny, repeatable background patterns.</p></a>
    <a href="/tools/noise-generator/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">SVG Noise Generator</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate a sub-1KB feTurbulence filter as a data URI, instead of shipping a 1MB noise PNG.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/blob-maker/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/ai-builders/content-first.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">The Content-First Strategy for Starter Sites</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Before complex platforms and dynamic databases you need traffic. Why flat HTML is the smartest foundation for a new site.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/prompt-architect/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Prompt Architect</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Build better Midjourney prompts.</p></a>
    <a href="/tools/logic-architect/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Logic Architect</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Map out complex prompt structures and export optimized blueprints for LLMs.</p></a>
    <a href="/tools/monolith-splitter/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Monolith Splitter</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Safely refactor giant AI-generated files into modular components without breaking the logic.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/blueprint-compiler/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/design/svg-patterns.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Algorithmic Textures</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Stop using heavy PNGs. Why mathematical SVG patterns scale infinitely for free.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/blob-maker/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Blob Maker</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Create organic SVG shapes.</p></a>
    <a href="/tools/parallax-generator/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Parallax Generator</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Pure CSS 3D scrolling effects.</p></a>
    <a href="/tools/cubic-bezier/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Cubic Bezier Architect</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Drag the handles to shape a custom easing curve and make UI motion feel deliberate.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/clip-path/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/communities/graph-vs-relational.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">The Architecture of Connection</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Relational tables are neat and rigid. When your data is mostly relationships, that rigidity starts to cost you.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/recommender-engine/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Recommender Engine</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Simulate Cosine Similarity vectors.</p></a>
    <a href="/tools/bayesian-rank/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Bayesian Ranking</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Fair sorting for star ratings.</p></a>
    <a href="/tools/rls-architect/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Supabase RLS Architect</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Define row-level security rules for your tables before an AI-built app leaks data.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/community-growth/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/security/xss-vulnerability.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Understanding XSS</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Why 90% of websites are vulnerable to injection and how CSP headers fix it.</p></a>
    <a href="/learn/security/cdn-risks.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">The CDN Trap</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Why loading scripts from Google/Cloudflare is risky without Subresource Integrity (SRI).</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/sri-generator/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">SRI Hash Generator</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Lock down CDNs with integrity hashes.</p></a>
    <a href="/tools/head-architect/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">HTML Head Architect</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate SEO tags &amp; JSON-LD identity.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/csp-builder/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/design/ambient-occlusion-css.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Ambient Occlusion in CSS</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Real shadows are gradients of light deprivation, darkest where an object touches its surface. CSS can say that.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/smooth-shadow/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Smooth Shadows</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate layered ambient occlusion shadows.</p></a>
    <a href="/tools/mesh-gradient/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Aurora Mesh</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate trendy, moving mesh gradients.</p></a>
    <a href="/tools/noise-generator/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">SVG Noise Generator</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate a sub-1KB feTurbulence filter as a data URI, instead of shipping a 1MB noise PNG.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/css-filter-playground/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/design/fluid-web-theory.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Fluid Web Theory</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Why &#34;breakpoints&#34; are broken and how to use Linear Interpolation (lerp) for scaling.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/fluid-typography/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Fluid Type Composer</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate mathematical clamp() formulas.</p></a>
    <a href="/tools/oklch-picker/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">OKLCH Color Mixer</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate perceptual color palettes with modern CSS.</p></a>
    <a href="/tools/grid-generator/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">CSS Grid Architect</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Visual layout builder for columns &amp; gaps.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/css-variables/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/design/physics-of-ui.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">The Physics of UI</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Why default CSS easing feels &#34;cheap&#34; and how to use Cubic Bezier curves to mimic real-world inertia.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/parallax-generator/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Parallax Generator</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Pure CSS 3D scrolling effects.</p></a>
    <a href="/tools/clip-path/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Clip-Path Builder</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Crop elements into triangles, hexagons or abstract shards by dragging the points.</p></a>
    <a href="/tools/smooth-shadow/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Smooth Shadows</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate layered ambient occlusion shadows.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/cubic-bezier/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/code/regex-visualized.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Taming the Beast</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Regex is a &#34;write-only&#34; language. Stop coding blind and learn to visualize your patterns.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/regex-tester/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Regex X-Ray</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Visualize Regular Expression matches in real-time.</p></a>
    <a href="/tools/text-extractor/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Text Extractor</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Strip HTML tags to extract pure text content.</p></a>
    <a href="/tools/json-cleaner/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">JSON Cleaner</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Truncate long values to debug logs.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/diff-checker/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/security/entropy-physics.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Password Entropy</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">The math behind brute-force attacks and why &#34;CorrectHorseBatteryStaple&#34; works.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/privacy-redactor/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Privacy Redactor</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Scrub PII (Emails/IPs) from logs.</p></a>
    <a href="/tools/jwt-inspector/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">JWT Inspector</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Decode Tokens securely in browser.</p></a>
    <a href="/tools/csp-builder/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">CSP Visual Builder</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate strict security headers.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/entropy-meter/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/operations/maintenance.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">The Hidden Cost of AI: Three Maintenance Problems</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">An AI builder optimises for immediate visual feedback, not long-term readability. Maintaining the result is a separate skill.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/animated-favicon/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Animated Favicon</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Create spinning tab icons with JS.</p></a>
    <a href="/tools/social-card/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Social Card Generator</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate Open Graph meta tags instantly.</p></a>
    <a href="/tools/image-optimizer/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Image Optimizer</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Convert PNG to JPG and resize locally.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/favicon-maker/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/design/fluid-web-theory.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Fluid Web Theory</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Why &#34;breakpoints&#34; are broken and how to use Linear Interpolation (lerp) for scaling.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/css-variables/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">CSS Architect</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate a rigorous :root theme file.</p></a>
    <a href="/tools/grid-generator/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">CSS Grid Architect</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Visual layout builder for columns &amp; gaps.</p></a>
    <a href="/tools/golden-ratio/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Golden Ratio Cropper</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Compose perfect images using the Fibonacci spiral.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/fluid-typography/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/accessibility/focus-states.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">The Invisible Focus</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Every junior designer eventually asks whether that blue outline can go. Here is why it cannot, and what to do instead.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/aria-builder/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">ARIA Architect</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Wizard to generate semantic labels for icons.</p></a>
    <a href="/tools/smart-contrast/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Smart Palette</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Auto-fix accessible color contrast.</p></a>
    <a href="/tools/touch-target/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Touch Target Simulator</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Visualise 44px hit areas on mobile buttons.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/focus-ring/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/design/mood-boarding.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">The Messy Mood Board</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Stop overthinking. Why the secret to faster design workflows is embracing chaos and &#34;low-fidelity&#34; pasting.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/aspect-ratio/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Aspect Ratio Calc</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Calculate dimensions for screens and video.</p></a>
    <a href="/tools/white-balance/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Batch White Balance</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Warm up or cool down a set of images instantly.</p></a>
    <a href="/tools/bg-remover/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Magic Eraser</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Remove image backgrounds using chroma key logic.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/golden-ratio/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/design/css-grid-math.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Fractional Layouts</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Percentages broke the moment you added a margin. The fr unit exists because of that arithmetic.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/layout-generator/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Holy Grail Layouts</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate responsive CSS Grid code instantly.</p></a>
    <a href="/tools/aspect-ratio/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Aspect Ratio Calc</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Calculate dimensions for screens and video.</p></a>
    <a href="/tools/fluid-typography/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Fluid Type Composer</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate mathematical clamp() formulas.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/grid-generator/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/data/cors-scraping.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Why You Can&#39;t Scrape Google</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">The Same-Origin Policy explained. Why `fetch()` fails on external sites.</p></a>
    <a href="/learn/marketing/seo-for-llms.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Optimizing for LLMs: Beyond Keywords</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">For twenty years we wrote for ten blue links. People now ask a model instead, and it reads differently.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/seo-schema/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Semantic Schema</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Create JSON-LD markup for SEO.</p></a>
    <a href="/tools/seo-injector/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">JSON-LD SEO Injector</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Turn local business details into strict LocalBusiness JSON-LD wrapped in a prompt.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/head-architect/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/performance/physics-of-latency.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">The Physics of Latency</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Why increasing bandwidth does not solve speed issues if RTT is high.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/svg-optimizer/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">SVG Code Stripper</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Remove XML junk to reduce file size.</p></a>
    <a href="/tools/image-optimizer/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Image Optimizer</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Convert PNG to JPG and resize locally.</p></a>
    <a href="/tools/text-extractor/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Text Extractor</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Strip HTML tags to extract pure text content.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/html-minifier/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/performance/physics-of-latency.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">The Physics of Latency</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Why increasing bandwidth does not solve speed issues if RTT is high.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/svg-optimizer/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">SVG Code Stripper</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Remove XML junk to reduce file size.</p></a>
    <a href="/tools/white-balance/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Batch White Balance</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Warm up or cool down a set of images instantly.</p></a>
    <a href="/tools/bg-remover/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Magic Eraser</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Remove image backgrounds using chroma key logic.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/image-optimizer/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/ai-builders/anti-slop.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Preventing AI Slop: A Quality Manifesto</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">AI tools are efficient at generating code and equally efficient at generating generic web pollution. A stance on avoiding the second.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/prompt-architect/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Prompt Architect</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Build better Midjourney prompts.</p></a>
    <a href="/tools/text-sanitizer/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Text Sanitizer</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Remove smart quotes/unicode for AI.</p></a>
    <a href="/tools/token-calculator/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Token Calculator</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Estimate AI costs (GPT-4/Claude).</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/insight-injector/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/data/cors-scraping.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Why You Can&#39;t Scrape Google</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">The Same-Origin Policy explained. Why `fetch()` fails on external sites.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/text-extractor/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Text Extractor</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Strip HTML tags to extract pure text content.</p></a>
    <a href="/tools/text-sanitizer/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Text Sanitizer</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Remove smart quotes/unicode for AI.</p></a>
    <a href="/tools/diff-checker/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Visual Code Diff</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Compare two blocks of code to find changes.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/json-cleaner/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/security/xss-vulnerability.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Understanding XSS</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Why 90% of websites are vulnerable to injection and how CSP headers fix it.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/entropy-meter/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Password Entropy</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Check password strength physics.</p></a>
    <a href="/tools/privacy-redactor/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Privacy Redactor</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Scrub PII (Emails/IPs) from logs.</p></a>
    <a href="/tools/csp-builder/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">CSP Visual Builder</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate strict security headers.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/jwt-inspector/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/design/css-grid-math.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Fractional Layouts</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Percentages broke the moment you added a margin. The fr unit exists because of that arithmetic.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/grid-generator/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">CSS Grid Architect</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Visual layout builder for columns &amp; gaps.</p></a>
    <a href="/tools/aspect-ratio/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Aspect Ratio Calc</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Calculate dimensions for screens and video.</p></a>
    <a href="/tools/css-variables/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">CSS Architect</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate a rigorous :root theme file.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/layout-generator/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/ai-builders/lovable.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Understanding Lovable</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">A full-stack AI platform that builds the interface and the database logic at once. How that actually plays out.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/blueprint-compiler/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">AI Site Blueprint Compiler</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Define the pages and strict parameters an AI builder must follow, then compile them into a phased prompt deck.</p></a>
    <a href="/tools/mind-map/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Mind Map Studio</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Visual node editor for brainstorming. Auto-layout, drag-and-drop, and export.</p></a>
    <a href="/tools/monolith-splitter/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Monolith Splitter</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Safely refactor giant AI-generated files into modular components without breaking the logic.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/logic-architect/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/design/svg-patterns.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Algorithmic Textures</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Stop using heavy PNGs. Why mathematical SVG patterns scale infinitely for free.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/bg-remover/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Magic Eraser</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Remove image backgrounds using chroma key logic.</p></a>
    <a href="/tools/clip-path/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Clip-Path Builder</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Crop elements into triangles, hexagons or abstract shards by dragging the points.</p></a>
    <a href="/tools/blob-maker/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Blob Maker</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Create organic SVG shapes.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/magic-outliner/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/operations/maintenance.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">The Hidden Cost of AI: Three Maintenance Problems</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">An AI builder optimises for immediate visual feedback, not long-term readability. Maintaining the result is a separate skill.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/text-extractor/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Text Extractor</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Strip HTML tags to extract pure text content.</p></a>
    <a href="/tools/micro-cms/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Flat-File Micro CMS</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Edit a page visually in the browser, with fifty checkpoints of version history. Everything stays in your own browser storage.</p></a>
    <a href="/tools/pasteboard/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Pasteboard Studio</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Browser-based moodboarding. Paste, drag, and layer images instantly.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/markdown-tables/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/design/mood-boarding.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">The Messy Mood Board</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Stop overthinking. Why the secret to faster design workflows is embracing chaos and &#34;low-fidelity&#34; pasting.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/social-card/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Social Card Generator</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate Open Graph meta tags instantly.</p></a>
    <a href="/tools/bg-remover/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Magic Eraser</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Remove image backgrounds using chroma key logic.</p></a>
    <a href="/tools/magic-outliner/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Magic Wand Outliner</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Click any region of a flat image to trace its border with a flood-fill and contour trace.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/meme-generator/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/design/mood-boarding.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">The Messy Mood Board</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Stop overthinking. Why the secret to faster design workflows is embracing chaos and &#34;low-fidelity&#34; pasting.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/oklch-picker/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">OKLCH Color Mixer</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate perceptual color palettes with modern CSS.</p></a>
    <a href="/tools/blob-maker/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Blob Maker</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Create organic SVG shapes.</p></a>
    <a href="/tools/noise-generator/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">SVG Noise Generator</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate a sub-1KB feTurbulence filter as a data URI, instead of shipping a 1MB noise PNG.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/mesh-gradient/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/operations/maintenance.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">The Hidden Cost of AI: Three Maintenance Problems</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">An AI builder optimises for immediate visual feedback, not long-term readability. Maintaining the result is a separate skill.</p></a>
    <a href="/learn/operations/browser-storage.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">The Realities of Browser Storage</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Our tools keep your data in your own browser rather than on a server. What that means for backups, limits and losing things.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/pasteboard/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Pasteboard Studio</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Browser-based moodboarding. Paste, drag, and layer images instantly.</p></a>
    <a href="/tools/markdown-tables/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Table Architect</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Convert Excel/CSV to Markdown for AI.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/micro-cms/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/ai-builders/content-first.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">The Content-First Strategy for Starter Sites</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Before complex platforms and dynamic databases you need traffic. Why flat HTML is the smartest foundation for a new site.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/logic-architect/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Logic Architect</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Map out complex prompt structures and export optimized blueprints for LLMs.</p></a>
    <a href="/tools/blueprint-compiler/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">AI Site Blueprint Compiler</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Define the pages and strict parameters an AI builder must follow, then compile them into a phased prompt deck.</p></a>
    <a href="/tools/micro-cms/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Flat-File Micro CMS</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Edit a page visually in the browser, with fifty checkpoints of version history. Everything stays in your own browser storage.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/mind-map/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/operations/scaling.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">The 70% Wall</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">AI platforms take a prompt to a beautiful, functional application — and then progress slows sharply. Why, and what to do about it.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/micro-cms/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Flat-File Micro CMS</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Edit a page visually in the browser, with fifty checkpoints of version history. Everything stays in your own browser storage.</p></a>
    <a href="/tools/rls-architect/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Supabase RLS Architect</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Define row-level security rules for your tables before an AI-built app leaks data.</p></a>
    <a href="/tools/blueprint-compiler/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">AI Site Blueprint Compiler</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Define the pages and strict parameters an AI builder must follow, then compile them into a phased prompt deck.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/monolith-splitter/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/design/digital-grain.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Digital Grain &amp; Noise</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Flat design is dead. How to use algorithmic SVG noise to add texture without killing performance.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/svg-patterns/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">SVG Pattern Engine</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate tiny, repeatable background patterns.</p></a>
    <a href="/tools/mesh-gradient/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Aurora Mesh</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate trendy, moving mesh gradients.</p></a>
    <a href="/tools/blob-maker/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Blob Maker</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Create organic SVG shapes.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/noise-generator/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/design/oklch-colors.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">The End of Hex Codes</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Why OKLCH is superior to RGB for human perception and accessible theming.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/smart-contrast/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Smart Palette</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Auto-fix accessible color contrast.</p></a>
    <a href="/tools/mesh-gradient/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Aurora Mesh</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate trendy, moving mesh gradients.</p></a>
    <a href="/tools/white-balance/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Batch White Balance</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Warm up or cool down a set of images instantly.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/oklch-picker/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/design/physics-of-ui.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">The Physics of UI</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Why default CSS easing feels &#34;cheap&#34; and how to use Cubic Bezier curves to mimic real-world inertia.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/cubic-bezier/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Cubic Bezier Architect</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Drag the handles to shape a custom easing curve and make UI motion feel deliberate.</p></a>
    <a href="/tools/clip-path/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Clip-Path Builder</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Crop elements into triangles, hexagons or abstract shards by dragging the points.</p></a>
    <a href="/tools/mesh-gradient/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Aurora Mesh</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate trendy, moving mesh gradients.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/parallax-generator/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/operations/browser-storage.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">The Realities of Browser Storage</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Our tools keep your data in your own browser rather than on a server. What that means for backups, limits and losing things.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/micro-cms/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Flat-File Micro CMS</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Edit a page visually in the browser, with fifty checkpoints of version history. Everything stays in your own browser storage.</p></a>
    <a href="/tools/markdown-tables/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Table Architect</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Convert Excel/CSV to Markdown for AI.</p></a>
    <a href="/tools/json-cleaner/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">JSON Cleaner</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Truncate long values to debug logs.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/pasteboard/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/performance/physics-of-latency.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">The Physics of Latency</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Why increasing bandwidth does not solve speed issues if RTT is high.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/image-optimizer/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Image Optimizer</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Convert PNG to JPG and resize locally.</p></a>
    <a href="/tools/html-minifier/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">HTML Minifier</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Strip whitespace and comments.</p></a>
    <a href="/tools/svg-optimizer/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">SVG Code Stripper</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Remove XML junk to reduce file size.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/performance-budget/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/security/entropy-physics.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Password Entropy</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">The math behind brute-force attacks and why &#34;CorrectHorseBatteryStaple&#34; works.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/text-sanitizer/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Text Sanitizer</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Remove smart quotes/unicode for AI.</p></a>
    <a href="/tools/entropy-meter/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Password Entropy</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Check password strength physics.</p></a>
    <a href="/tools/json-cleaner/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">JSON Cleaner</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Truncate long values to debug logs.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/privacy-redactor/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/ai-builders/anti-slop.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Preventing AI Slop: A Quality Manifesto</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">AI tools are efficient at generating code and equally efficient at generating generic web pollution. A stance on avoiding the second.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/prompt-permutator/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Prompt Matrix</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate 50+ prompt variations instantly.</p></a>
    <a href="/tools/blueprint-compiler/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">AI Site Blueprint Compiler</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Define the pages and strict parameters an AI builder must follow, then compile them into a phased prompt deck.</p></a>
    <a href="/tools/token-calculator/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Token Calculator</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Estimate AI costs (GPT-4/Claude).</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/prompt-architect/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/ai-builders/anti-slop.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Preventing AI Slop: A Quality Manifesto</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">AI tools are efficient at generating code and equally efficient at generating generic web pollution. A stance on avoiding the second.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/prompt-architect/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Prompt Architect</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Build better Midjourney prompts.</p></a>
    <a href="/tools/token-calculator/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Token Calculator</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Estimate AI costs (GPT-4/Claude).</p></a>
    <a href="/tools/blueprint-compiler/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">AI Site Blueprint Compiler</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Define the pages and strict parameters an AI builder must follow, then compile them into a phased prompt deck.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/prompt-permutator/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/algorithms/recommendation-math.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">How Netflix Knows What You Want</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Recommendation engines do not know whether a song is sad or happy. They know one thing: vectors.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/bayesian-rank/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Bayesian Ranking</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Fair sorting for star ratings.</p></a>
    <a href="/tools/community-growth/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Growth Simulator</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Visualize Viral Coefficient (k) vs Churn.</p></a>
    <a href="/tools/ab-test-calculator/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">A/B Significance</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Calculate P-Values and Z-Scores.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/recommender-engine/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/code/regex-visualized.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Taming the Beast</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Regex is a &#34;write-only&#34; language. Stop coding blind and learn to visualize your patterns.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/text-extractor/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Text Extractor</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Strip HTML tags to extract pure text content.</p></a>
    <a href="/tools/diff-checker/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Visual Code Diff</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Compare two blocks of code to find changes.</p></a>
    <a href="/tools/json-cleaner/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">JSON Cleaner</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Truncate long values to debug logs.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/regex-tester/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/communities/graph-vs-relational.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">The Architecture of Connection</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Relational tables are neat and rigid. When your data is mostly relationships, that rigidity starts to cost you.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/community-growth/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Growth Simulator</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Visualize Viral Coefficient (k) vs Churn.</p></a>
    <a href="/tools/monolith-splitter/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Monolith Splitter</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Safely refactor giant AI-generated files into modular components without breaking the logic.</p></a>
    <a href="/tools/micro-cms/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Flat-File Micro CMS</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Edit a page visually in the browser, with fifty checkpoints of version history. Everything stays in your own browser storage.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/rls-architect/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/marketing/seo-for-llms.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Optimizing for LLMs: Beyond Keywords</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">For twenty years we wrote for ten blue links. People now ask a model instead, and it reads differently.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/seo-schema/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Semantic Schema</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Create JSON-LD markup for SEO.</p></a>
    <a href="/tools/head-architect/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">HTML Head Architect</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate SEO tags &amp; JSON-LD identity.</p></a>
    <a href="/tools/social-card/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Social Card Generator</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate Open Graph meta tags instantly.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/seo-injector/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/marketing/seo-for-llms.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Optimizing for LLMs: Beyond Keywords</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">For twenty years we wrote for ten blue links. People now ask a model instead, and it reads differently.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/seo-injector/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">JSON-LD SEO Injector</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Turn local business details into strict LocalBusiness JSON-LD wrapped in a prompt.</p></a>
    <a href="/tools/head-architect/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">HTML Head Architect</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate SEO tags &amp; JSON-LD identity.</p></a>
    <a href="/tools/social-card/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Social Card Generator</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate Open Graph meta tags instantly.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/seo-schema/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/design/layered-shadows.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Beyond Flat Design</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">How to use &#34;Shadow Stacking&#34; to create hyper-realistic depth that looks expensive, not fake.</p></a>
    <a href="/learn/design/ambient-occlusion-css.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Ambient Occlusion in CSS</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Real shadows are gradients of light deprivation, darkest where an object touches its surface. CSS can say that.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/smooth-shadow/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Smooth Shadows</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate layered ambient occlusion shadows.</p></a>
    <a href="/tools/css-filter-playground/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">CSS Filter Playground</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Visual editor for brightness, blur, and hue.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/shadow-stacker/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/accessibility/focus-states.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">The Invisible Focus</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Every junior designer eventually asks whether that blue outline can go. Here is why it cannot, and what to do instead.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/oklch-picker/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">OKLCH Color Mixer</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate perceptual color palettes with modern CSS.</p></a>
    <a href="/tools/focus-ring/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Focus Ring Architect</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate premium, accessible double-layer focus states.</p></a>
    <a href="/tools/aria-builder/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">ARIA Architect</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Wizard to generate semantic labels for icons.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/smart-contrast/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/design/ambient-occlusion-css.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Ambient Occlusion in CSS</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Real shadows are gradients of light deprivation, darkest where an object touches its surface. CSS can say that.</p></a>
    <a href="/learn/design/layered-shadows.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Beyond Flat Design</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">How to use &#34;Shadow Stacking&#34; to create hyper-realistic depth that looks expensive, not fake.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/shadow-stacker/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Shadow Stacker</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Layer multiple shadows for hyper-realistic depth.</p></a>
    <a href="/tools/css-filter-playground/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">CSS Filter Playground</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Visual editor for brightness, blur, and hue.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/smooth-shadow/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/marketing/seo-for-llms.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Optimizing for LLMs: Beyond Keywords</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">For twenty years we wrote for ten blue links. People now ask a model instead, and it reads differently.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/seo-schema/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Semantic Schema</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Create JSON-LD markup for SEO.</p></a>
    <a href="/tools/meme-generator/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Meme Studio</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Create community assets with Impact font.</p></a>
    <a href="/tools/favicon-maker/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Favicon Maker</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Convert Emojis or Logos to ICO/PNG.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/social-card/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/security/cdn-risks.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">The CDN Trap</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Why loading scripts from Google/Cloudflare is risky without Subresource Integrity (SRI).</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/csp-builder/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">CSP Visual Builder</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate strict security headers.</p></a>
    <a href="/tools/head-architect/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">HTML Head Architect</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate SEO tags &amp; JSON-LD identity.</p></a>
    <a href="/tools/jwt-inspector/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">JWT Inspector</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Decode Tokens securely in browser.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/sri-generator/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/performance/physics-of-latency.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">The Physics of Latency</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Why increasing bandwidth does not solve speed issues if RTT is high.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/image-optimizer/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Image Optimizer</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Convert PNG to JPG and resize locally.</p></a>
    <a href="/tools/svg-patterns/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">SVG Pattern Engine</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate tiny, repeatable background patterns.</p></a>
    <a href="/tools/html-minifier/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">HTML Minifier</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Strip whitespace and comments.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/svg-optimizer/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/design/svg-patterns.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Algorithmic Textures</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Stop using heavy PNGs. Why mathematical SVG patterns scale infinitely for free.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/noise-generator/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">SVG Noise Generator</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate a sub-1KB feTurbulence filter as a data URI, instead of shipping a 1MB noise PNG.</p></a>
    <a href="/tools/blob-maker/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Blob Maker</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Create organic SVG shapes.</p></a>
    <a href="/tools/mesh-gradient/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Aurora Mesh</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate trendy, moving mesh gradients.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/svg-patterns/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/data/cors-scraping.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Why You Can&#39;t Scrape Google</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">The Same-Origin Policy explained. Why `fetch()` fails on external sites.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/json-cleaner/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">JSON Cleaner</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Truncate long values to debug logs.</p></a>
    <a href="/tools/text-sanitizer/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Text Sanitizer</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Remove smart quotes/unicode for AI.</p></a>
    <a href="/tools/markdown-tables/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Table Architect</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Convert Excel/CSV to Markdown for AI.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/text-extractor/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/data/unicode-gremlins.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">The Invisible Characters Breaking Your Code</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">You paste a snippet that looks perfect and get a SyntaxError. The characters you cannot see are the problem.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/text-extractor/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Text Extractor</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Strip HTML tags to extract pure text content.</p></a>
    <a href="/tools/privacy-redactor/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Privacy Redactor</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Scrub PII (Emails/IPs) from logs.</p></a>
    <a href="/tools/json-cleaner/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">JSON Cleaner</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Truncate long values to debug logs.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/text-sanitizer/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/ai-builders/content-first.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">The Content-First Strategy for Starter Sites</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Before complex platforms and dynamic databases you need traffic. Why flat HTML is the smartest foundation for a new site.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/prompt-architect/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Prompt Architect</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Build better Midjourney prompts.</p></a>
    <a href="/tools/prompt-permutator/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Prompt Matrix</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate 50+ prompt variations instantly.</p></a>
    <a href="/tools/text-sanitizer/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Text Sanitizer</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Remove smart quotes/unicode for AI.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/token-calculator/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/accessibility/touch-targets.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">The 44px Rule</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">We build on laptops with precision trackpads; people use glass screens with thumbs. That gap has a measurable size.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/aria-builder/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">ARIA Architect</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Wizard to generate semantic labels for icons.</p></a>
    <a href="/tools/focus-ring/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Focus Ring Architect</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate premium, accessible double-layer focus states.</p></a>
    <a href="/tools/smart-contrast/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Smart Palette</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Auto-fix accessible color contrast.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/touch-target/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/design/mood-boarding.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">The Messy Mood Board</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Stop overthinking. Why the secret to faster design workflows is embracing chaos and &#34;low-fidelity&#34; pasting.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/mesh-gradient/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Aurora Mesh</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate trendy, moving mesh gradients.</p></a>
    <a href="/tools/oklch-picker/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">OKLCH Color Mixer</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate perceptual color palettes with modern CSS.</p></a>
    <a href="/tools/css-variables/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">CSS Architect</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Generate a rigorous :root theme file.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/vibe-equalizer/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

UPDATE page_components pc
   SET rendered_html = regexp_replace(pc.rendered_html, '</section>[[:space:]]*$',
       $blk$
<div class="wd-related" style="margin-top: var(--space-lg); padding-top: var(--space-lg); border-top: 1px solid var(--border);">
  <h3 style="font-size:1rem; margin:0 0 0.75rem;">The thinking behind this tool</h3>
  <div class="grid-cards">
    <a href="/learn/design/mood-boarding.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">The Messy Mood Board</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Stop overthinking. Why the secret to faster design workflows is embracing chaos and &#34;low-fidelity&#34; pasting.</p></a>
  </div>
  <h3 style="font-size:1rem; margin:1.5rem 0 0.75rem;">Related tools</h3>
  <div class="grid-cards">
    <a href="/tools/image-optimizer/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Image Optimizer</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Convert PNG to JPG and resize locally.</p></a>
    <a href="/tools/bg-remover/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Magic Eraser</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Remove image backgrounds using chroma key logic.</p></a>
    <a href="/tools/golden-ratio/index.html" class="card" style="text-decoration:none; color:inherit; display:block;"><h4 style="margin:0; font-size:1.05rem;">Golden Ratio Cropper</h4><p style="margin:0.4rem 0 0; color:var(--text-dim);">Compose perfect images using the Fibonacci spiral.</p></a>
  </div>
</div>
</section>$blk$),
       updated_at = NOW()
  FROM pages p JOIN sites s ON s.id = p.site_id
 WHERE pc.page_id = p.id AND s.domain = 'webdesign.co.uk'
   AND p.url = '/tools/white-balance/index.html'
   AND pc.rendered_html NOT LIKE '%wd-related%';

DO $verify$
DECLARE n_with int; n_tot int;
BEGIN
    SELECT count(*) FILTER (WHERE pc.rendered_html LIKE '%wd-related%'), count(*)
      INTO n_with, n_tot
      FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
     WHERE s.domain='webdesign.co.uk' AND p.url LIKE '/tools/%' AND p.url <> '/tools/index.html';
    RAISE NOTICE 'tool page components carrying the block: % of %', n_with, n_tot;
    IF n_with <> 63 THEN RAISE EXCEPTION 'expected 63, got %', n_with; END IF;
END
$verify$;

COMMIT;

-- SQL_p20_register_buying_design_pages.sql — webdesign.co.uk
--
-- Register the two hand-authored Buying design pages (committed to gqls/sites
-- as 7336286ca, deployed by the Action) in the platform's page tables, the
-- same way the ported pages are registered: a pages row with
-- rebuild_policy='owned' plus a single ported-page component holding the
-- section fragment as rendered_html with provenance in content_data. This
-- keeps every platform surface (nav rebuild, re-renders, link checks) aware
-- of the pages; a chrome re-render regenerates the file from this row, so the
-- stored fragment is byte-for-byte what the committed file carries.
--
-- WHY HAND-AUTHORED: the D13 exposure content quotes our own closed bug
-- records and the figures rail is absolute ("no figures, no page"); an LLM
-- writer cannot be allowed near either. generator names the thread.
--
-- in_header=true on the section index ONLY (page_type='section-index', which
-- populate_nav_tables admits through its child-URL exception — unlike
-- news-index, whose exclusion is the Go fix this session files separately).

\set ON_ERROR_STOP on

BEGIN;

INSERT INTO pages (site_id, name, url, title, page_type, status, meta_description,
                   sections, rebuild_policy, build_status, deployed_at, last_built_at,
                   in_header, in_footer, nav_label, nav_order)
SELECT s.id, 'buying-design', '/buying-design/index.html', 'Buying design | webdesign.co.uk',
       'section-index', 'active', 'For people commissioning a substantial web project: what to ask for, what goes wrong, and what AI actually changes. Written by people running an AI build system who are not bidding for your work.',
       '["ported-page"]'::jsonb, 'owned', 'deployed', NOW(), NOW(),
       true, true, 'Buying design', 50
FROM sites s
WHERE s.domain = 'webdesign.co.uk'
  AND NOT EXISTS (SELECT 1 FROM pages x WHERE x.site_id = s.id AND x.name = 'buying-design');

INSERT INTO page_components (page_id, component_id, position, rendered_html,
                             content_data, content_hash, build_status)
SELECT p.id, 'a7daa5c5-8cfd-4f2c-8e09-de6abcb637ef', 0, $bdidx$<section class="ported-page" data-component="ported-page">
<style>
.article-content { max-width: 70ch; margin: 0 auto; font-size: 1.125rem; line-height: 1.7; }
        .article-content h2 { margin-top: 3rem; margin-bottom: 1rem; color: var(--color-ink); }
        .article-content p { margin-bottom: 1.5rem; color: var(--text); }
        .article-content table { width: 100%; border-collapse: collapse; margin: 2rem 0; }
        .article-content td, .article-content th { padding: 0.75rem 0.9rem; border-top: 1px solid var(--color-border); vertical-align: top; text-align: left; }
        .article-content th { font-size: 0.85rem; text-transform: uppercase; letter-spacing: 0.05em; color: var(--text-dim); }
        .record-list { border-left: 3px solid var(--color-accent); padding-left: 1.5rem; margin: 2rem 0; }
</style>
<section style="padding: var(--space-lg) 0; border-bottom: 1px solid var(--color-border);">
        <div class="container">
            <span class="text-mono text-accent" style="font-size: 0.85rem; font-weight: 700;">COMMISSIONING</span>
            <h1 style="font-size: var(--text-h1); font-weight: 800; margin-top: 0.5rem;">Buying design</h1>
            <p style="font-size: 1.25rem; color: var(--text-dim); max-width: 60ch; margin-top: 1rem;">
                For people commissioning a substantial web project: what to ask for, what goes wrong, and what AI actually changes.
            </p>
        </div>
    </section>

    <article class="container" style="padding-top: var(--space-lg);">
        <div class="article-content">

            <p class="lead">
                Most of this site is for the people who make websites. This section is for the people who pay for them: a brand or marketing director, a digital lead, a procurement team with a board behind them.
            </p>

            <h2>Who is writing this</h2>
            <p>
                We run an AI web-build system in production. It plans sites, writes their content, designs them, deploys them and maintains them, with people supervising rather than typing. The sites it builds include the one you are reading.
            </p>
            <p>
                We are not an agency, and we are not bidding for your project. An agency has a reason to tell you AI changes little. An AI vendor has a reason to tell you it changes everything. We sell neither day rates nor licences, so we can describe what this way of building gets right and where it fails, from our own operating records.
            </p>

            <h2>What actually goes wrong</h2>
            <p>
                Anyone can publish a capability list. Failure records are rarer, so here are some of ours.
            </p>
            <div class="record-list">
                <p>
                    In July 2026 a routine check found three pages on robot-hands.com, one of our sites, serving statistics the page writer had made up. Nothing in the pipeline compared a published figure with the data behind it.
                </p>
                <p>
                    The fix refused any figure the system could not trace to data. The first honest page then failed to build. A writer that truthfully reported no number for a statistics slot hit a required field, and the build died. The old contract had paid the model to invent, and the new one initially punished it for stopping.
                </p>
                <p>
                    A repair agent rewrote a working tool and saved back an eighth of it. 10,272 characters became 1,253. The agent reported success, and nothing noticed until a person read the result.
                </p>
                <p>
                    For weeks no check compared a button's label with where it pointed. One homepage went live with its main call to action leading nowhere, while an automated check reported the site clean.
                </p>
                <p>
                    The site you are reading has shipped invented figures twice. On a third occasion a tool count of 62 disagreed with the 63 files on disk, and the disagreement is the only reason that one never shipped.
                </p>
            </div>
            <p>
                One fix we will claim, because we can show its evidence. Pages in our system now refuse to publish a figure they cannot trace to a data source. The first page rebuilt under that rule published two figures with sources attached and left three statistics honestly blank rather than inventing them. The blanks are the evidence.
            </p>
            <p>
                We publish this because the dominant experience of buying in this market is being sold a story. The most useful thing we can hand you is the texture of what actually goes wrong when machines build websites, and a supplier who tells you their system has no such list is describing their records, not their system.
            </p>

            <h2>Where to start</h2>
            <p>
                The first page covers accessibility, because it is the one duty you carry before any contract is signed: <a href="/buying-design/accessibility.html">Accessibility is a duty you already have</a>. It sets out the law, the standard to name in a contract, and three checks you can run yourself in minutes.
            </p>
            <p>
                More follows on supplier selection, why large projects fail, and what you should own at handover.
            </p>

            <h2>The rules this section follows</h2>
            <table>
                <tr><th>Rule</th><th>What it means</th></tr>
                <tr><td><strong>No rankings</strong></td><td>We never publish comparative judgements of named agencies or their clients' sites.</td></tr>
                <tr><td><strong>No commissions</strong></td><td>Nothing in this section is sponsored or affiliate. Nobody has paid to appear.</td></tr>
                <tr><td><strong>No unsourced figures</strong></td><td>A number appears with its source, or it does not appear. Our records above explain why.</td></tr>
            </table>

            <div style="background: var(--callout-bg); border: 1px solid var(--callout-border); padding: 2rem; border-radius: 12px; text-align: center; margin: 4rem 0;">
                <h3 style="margin-top: 0;">Start with the duty you already have.</h3>
                <p>The law, the standard for the contract, and three checks you can see for yourself.</p>
                <a href="/buying-design/accessibility.html" class="btn" style="background: var(--color-accent);">Read the accessibility page &rarr;</a>
            </div>

        </div>
    </article>
</section>
$bdidx$,
       jsonb_build_object(
         'schema', 'ported-page.v1',
         'sha256', '93d3ad229d646e159301b747a83f6d920748f56edf28126b3e1dfc1f1cf92c8a',
         'source', 'hand-authored: PLAN_2026-07-27b_buying_design.md',
         'qa_tier', '3',
         'generator', 'webdesign_couk_thread_4/manual'),
       '93d3ad229d646e159301b747a83f6d920748f56edf28126b3e1dfc1f1cf92c8a', 'approved'
FROM pages p JOIN sites s ON s.id = p.site_id
WHERE s.domain = 'webdesign.co.uk' AND p.name = 'buying-design'
  AND NOT EXISTS (SELECT 1 FROM page_components pc WHERE pc.page_id = p.id);

INSERT INTO pages (site_id, name, url, title, page_type, status, meta_description,
                   sections, rebuild_policy, build_status, deployed_at, last_built_at,
                   in_header, in_footer, nav_label, nav_order)
SELECT s.id, 'buying-design-accessibility', '/buying-design/accessibility.html', 'Accessibility is a duty you already have | webdesign.co.uk',
       'guide', 'active', 'The Equality Act 2010 obliges UK service providers to make reasonable adjustments. The standard to name in a contract, and three checks you can see for yourself.',
       '["ported-page"]'::jsonb, 'owned', 'deployed', NOW(), NOW(),
       false, false, NULL, 100
FROM sites s
WHERE s.domain = 'webdesign.co.uk'
  AND NOT EXISTS (SELECT 1 FROM pages x WHERE x.site_id = s.id AND x.name = 'buying-design-accessibility');

INSERT INTO page_components (page_id, component_id, position, rendered_html,
                             content_data, content_hash, build_status)
SELECT p.id, 'a7daa5c5-8cfd-4f2c-8e09-de6abcb637ef', 0, $bdacc$<section class="ported-page" data-component="ported-page">
<style>
.article-content { max-width: 70ch; margin: 0 auto; font-size: 1.125rem; line-height: 1.7; }
        .article-content h2 { margin-top: 3rem; margin-bottom: 1rem; color: var(--color-ink); }
        .article-content p { margin-bottom: 1.5rem; color: var(--text); }
        .article-content table { width: 100%; border-collapse: collapse; margin: 2rem 0; }
        .article-content td, .article-content th { padding: 0.75rem 0.9rem; border-top: 1px solid var(--color-border); vertical-align: top; text-align: left; }
        .article-content th { font-size: 0.85rem; text-transform: uppercase; letter-spacing: 0.05em; color: var(--text-dim); }
        .article-content blockquote { border-left: 3px solid var(--color-accent); margin: 2rem 0; padding: 0.25rem 0 0.25rem 1.5rem; font-style: italic; color: var(--text); }
        .article-content blockquote cite { display: block; margin-top: 0.75rem; font-style: normal; font-size: 0.9rem; color: var(--text-dim); }
        .sources-list { font-size: 0.95rem; color: var(--text-dim); }
        .sources-list li { margin-bottom: 0.5rem; }
</style>
<section style="padding: var(--space-lg) 0; border-bottom: 1px solid var(--color-border);">
        <div class="container">
            <span class="text-mono text-accent" style="font-size: 0.85rem; font-weight: 700;">COMMISSIONING &middot; ACCESSIBILITY</span>
            <h1 style="font-size: var(--text-h1); font-weight: 800; margin-top: 0.5rem;">Accessibility is a duty you already have</h1>
            <p style="font-size: 1.25rem; color: var(--text-dim); max-width: 60ch; margin-top: 1rem;">
                UK law obliges every service provider to make reasonable adjustments for disabled people. That includes your website, whoever builds it.
            </p>
        </div>
    </section>

    <article class="container" style="padding-top: var(--space-lg);">
        <div class="article-content">

            <p class="lead">
                In UK law, accessibility is an obligation that exists before any supplier is chosen. It belongs in the commissioning conversation from the first meeting, not in a quality checklist near the end.
            </p>

            <h2>What the law says</h2>
            <p>
                The Equality Act 2010 places a duty on service providers to make reasonable adjustments for disabled people. Section 20 defines the duty:
            </p>
            <blockquote>
                &ldquo;&hellip;where a provision, criterion or practice of A's puts a disabled person at a substantial disadvantage in relation to a relevant matter in comparison with persons who are not disabled, to take such steps as it is reasonable to have to take to avoid the disadvantage.&rdquo;
                <cite>Equality Act 2010, section 20 &mdash; legislation.gov.uk</cite>
            </blockquote>
            <p>
                Government guidance on web accessibility states the scope plainly:
            </p>
            <blockquote>
                &ldquo;All UK service providers have a legal obligation to make reasonable adjustments under the Equality Act 2010 or the Disability Discrimination Act 1995 (in Northern Ireland).&rdquo;
                <cite>GOV.UK, accessibility requirements guidance</cite>
            </blockquote>
            <p>
                Public sector bodies carry a second, more specific instrument: the Public Sector Bodies (Websites and Mobile Applications) (No. 2) Accessibility Regulations 2018. Government monitoring assesses those sites against WCAG 2.2, level AA.
            </p>

            <h2>The standard to name in the contract</h2>
            <p>
                The Act defines a duty, not a technical test. WCAG, the Web Content Accessibility Guidelines, is the standard that fills the gap. It is what UK government monitoring applies, and it is specific enough to write into a contract and check at handover. Name the version and the level: <strong>WCAG 2.2, level AA</strong>.
            </p>
            <table>
                <tr><th>Ask the supplier for</th><th>Why</th></tr>
                <tr><td><strong>Conformance to WCAG 2.2 AA, named in the contract</strong></td><td>&ldquo;Accessible&rdquo; as an adjective commits nobody. A named version and level is checkable.</td></tr>
                <tr><td><strong>An accessibility statement and audit evidence at handover</strong></td><td>Public sector sites publish a statement by law. A supplier who has done the work loses nothing by showing the same evidence privately.</td></tr>
                <tr><td><strong>Conformance that survives content changes</strong></td><td>A site conformant on launch day drifts as content is added. Make the ongoing duty explicit, including who holds it.</td></tr>
            </table>

            <h2>Three failures you can see for yourself</h2>
            <p>
                Much of WCAG needs human judgement against real pages. These three do not, and each takes minutes to see without technical help.
            </p>
            <table>
                <tr><th>Failure</th><th>The rule</th><th>See it</th></tr>
                <tr>
                    <td><strong>Text you cannot read</strong></td>
                    <td>WCAG 2.2 requires a contrast ratio of at least 4.5:1 for text (criterion 1.4.3, level AA).</td>
                    <td>Try your brand colours in <a href="/tools/smart-contrast/index.html">Smart Palette</a>, which auto-fixes pairs that fail.</td>
                </tr>
                <tr>
                    <td><strong>Buttons a thumb cannot hit</strong></td>
                    <td>WCAG 2.2's minimum target size is 24 by 24 CSS pixels (criterion 2.5.8, level AA). Apple and Google's platform guidelines ask for more: 44 points and 48dp.</td>
                    <td>The <a href="/tools/touch-target/index.html">Touch Target Simulator</a> visualises hit areas on mobile buttons. <a href="/learn/accessibility/touch-targets.html">The 44px Rule</a> explains why fingers miss.</td>
                </tr>
                <tr>
                    <td><strong>The vanishing keyboard focus</strong></td>
                    <td>Anyone navigating by keyboard must be able to see which element has focus (criterion 2.4.7, level AA).</td>
                    <td>The <a href="/tools/focus-ring/index.html">Focus Ring Architect</a> generates accessible focus states. <a href="/learn/accessibility/focus-states.html">The Invisible Focus</a> shows the classic way this breaks.</td>
                </tr>
            </table>
            <p>
                Press Tab a few times on your own site. If you lose track of where you are, so does everyone who cannot use a mouse.
            </p>

            <h2>What this page does not do</h2>
            <p>
                None of the checks above is an audit. WCAG 2.2 AA covers content, structure, interaction and media, and a large part of it needs a person with the standard open, judging real pages. Treat the three checks as a smoke test: cheap, fast and only ever indicative. They can tell you something is wrong. They cannot tell you everything is right.
            </p>

            <h2>Sources</h2>
            <ul class="sources-list">
                <li><a href="https://www.legislation.gov.uk/ukpga/2010/15/section/20" rel="noopener">Equality Act 2010, section 20</a> &mdash; legislation.gov.uk</li>
                <li><a href="https://www.gov.uk/guidance/accessibility-requirements-for-public-sector-websites-and-apps" rel="noopener">Understanding accessibility requirements for public sector bodies</a> &mdash; GOV.UK</li>
                <li><a href="https://www.w3.org/TR/WCAG22/" rel="noopener">Web Content Accessibility Guidelines (WCAG) 2.2</a> &mdash; W3C</li>
                <li><a href="https://www.w3.org/WAI/WCAG22/Understanding/target-size-minimum.html" rel="noopener">Understanding Target Size (Minimum)</a> &mdash; W3C</li>
            </ul>

            <div style="background: var(--callout-bg); border: 1px solid var(--callout-border); padding: 2rem; border-radius: 12px; text-align: center; margin: 4rem 0;">
                <h3 style="margin-top: 0;">Back to Buying design.</h3>
                <p>What to ask for, what goes wrong, and what AI actually changes.</p>
                <a href="/buying-design/index.html" class="btn" style="background: var(--color-accent);">The section front door &rarr;</a>
            </div>

        </div>
    </article>
</section>
$bdacc$,
       jsonb_build_object(
         'schema', 'ported-page.v1',
         'sha256', 'a3097416e04acbbca6a269b6d2198fb987612106d536f9881b0c21ce09ea592a',
         'source', 'hand-authored: PLAN_2026-07-27b_buying_design.md',
         'qa_tier', '3',
         'generator', 'webdesign_couk_thread_4/manual'),
       'a3097416e04acbbca6a269b6d2198fb987612106d536f9881b0c21ce09ea592a', 'approved'
FROM pages p JOIN sites s ON s.id = p.site_id
WHERE s.domain = 'webdesign.co.uk' AND p.name = 'buying-design-accessibility'
  AND NOT EXISTS (SELECT 1 FROM page_components pc WHERE pc.page_id = p.id);

DO $verify$
DECLARE v_site uuid; v_pages int; v_comps int;
BEGIN
    SELECT id INTO v_site FROM sites WHERE domain = 'webdesign.co.uk';
    SELECT count(*) INTO v_pages FROM pages
     WHERE site_id = v_site AND name IN ('buying-design','buying-design-accessibility')
       AND build_status = 'deployed' AND rebuild_policy = 'owned';
    IF v_pages <> 2 THEN RAISE EXCEPTION 'expected 2 registered pages, found %', v_pages; END IF;
    SELECT count(*) INTO v_comps FROM page_components pc
      JOIN pages p ON p.id = pc.page_id
     WHERE p.site_id = v_site AND p.name IN ('buying-design','buying-design-accessibility')
       AND length(COALESCE(pc.rendered_html,'')) > 3000;
    IF v_comps <> 2 THEN RAISE EXCEPTION 'expected 2 components with substantial rendered_html, found %', v_comps; END IF;
    RAISE NOTICE 'buying-design pages registered: 2 pages (owned/deployed), 2 ported-page components';
END $verify$;

COMMIT;

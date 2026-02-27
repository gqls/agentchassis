-- 065_tool_library_migration.sql  (v2 — fixed)
--
-- Fixes from v1:
--   1. Added partial unique index on function for library tools,
--      enabling ON CONFLICT
--   2. Multi-row INSERT for placeholder tools

-- ============================================================
-- 1. Add forked_from column
-- ============================================================
ALTER TABLE content_components
    ADD COLUMN IF NOT EXISTS forked_from UUID REFERENCES content_components(id);

-- Partial unique index on function for library tools only.
-- This prevents duplicate library tools and enables ON CONFLICT.
-- Site forks are excluded (forked_from IS NOT NULL) — multiple sites
-- can fork the same tool, each getting a row with the same function value.
CREATE UNIQUE INDEX IF NOT EXISTS idx_cc_tool_function_unique
    ON content_components (function)
    WHERE component_level = 'tool' AND forked_from IS NULL AND is_active = true;

-- Index for library queries (find canonical tools)
CREATE INDEX IF NOT EXISTS idx_cc_tool_library
    ON content_components (component_level)
    WHERE component_level = 'tool' AND forked_from IS NULL AND is_active = true;

-- Index for fork lookups (find all forks of a given tool)
CREATE INDEX IF NOT EXISTS idx_cc_forked_from
    ON content_components (forked_from)
    WHERE forked_from IS NOT NULL;

COMMENT ON COLUMN content_components.forked_from IS
  'For tool components: references the canonical library tool this was forked from. NULL = canonical (library) tool.';

-- ============================================================
-- 2. Seed canonical tool library
-- ============================================================

-- ----- AB Test Calculator (with full template) -----
INSERT INTO content_components (
    name, display_name, function, category, component_level, render_mode,
    is_dark_section, is_active, description,
    semantic_tags, html_template, input_schema, forked_from
) VALUES (
             'tool-ab-test-calculator',
             'A/B Test Significance Calculator',
             'tool-ab-test-calculator',
             'tool-calculator',
             'tool',
             'standalone',
             false,
             true,
             'Statistical significance calculator for A/B tests. Computes Z-score and tells users whether a conversion rate difference is real or noise. Uses 95% confidence interval (Z > 1.96).',
             '["calculator", "statistics", "marketing", "ab-testing", "conversion"]'::jsonb,
             '<style>
               .input-card { background: var(--color-surface, #fff); border: 1px solid var(--color-border, #ddd); padding: 1.5rem; border-radius: 8px; margin-bottom: 1rem; }
               .result-box { background: var(--color-primary, #1e1e1e); color: var(--color-white, #fff); padding: 2rem; border-radius: 8px; text-align: center; margin-top: 2rem; }
           </style>
           <main class="container" style="padding-top: var(--space-lg, 3rem);">
               <h1 style="font-size: var(--text-h2, 2rem);">A/B Test Significance</h1>
               <div class="guide-box" style="background: var(--color-surface, #fff); border: 1px solid var(--color-border, #ddd); padding: 1.5rem; border-radius: 8px; margin-bottom: 2rem; box-shadow: var(--shadow-card, 0 2px 8px rgba(0,0,0,0.08));">
                   <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 2rem;">
                       <div>
                           <h3 style="font-size: 1rem; color: var(--color-accent); margin-bottom: 0.5rem;">Was it just luck?</h3>
                           <p style="font-size: 0.9rem;">Seeing a higher conversion rate in Version B is great, but it might be random noise. This tool calculates the <strong>Z-Score</strong> to tell you if the difference is statistically real.</p>
                       </div>
                       <div>
                           <h3 style="font-size: 1rem; color: var(--color-accent); margin-bottom: 0.5rem;">Confidence Level (95%)</h3>
                           <p style="font-size: 0.9rem;">We use the standard 95% confidence interval (Z &gt; 1.96). This means there is less than a 5% chance your result happened by accident.</p>
                       </div>
                   </div>
               </div>
               <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 1rem;">
                   <div class="input-card">
                       <h3>Control (A)</h3>
                       <label>Visitors</label>
                       <input type="number" id="visA" value="1000" style="width:100%; padding:0.5rem; margin-bottom:1rem;">
                       <label>Conversions</label>
                       <input type="number" id="convA" value="50" style="width:100%; padding:0.5rem;">
                       <p>Conv. Rate: <strong id="rateA">5.0%</strong></p>
                   </div>
                   <div class="input-card">
                       <h3>Variant (B)</h3>
                       <label>Visitors</label>
                       <input type="number" id="visB" value="1000" style="width:100%; padding:0.5rem; margin-bottom:1rem;">
                       <label>Conversions</label>
                       <input type="number" id="convB" value="65" style="width:100%; padding:0.5rem;">
                       <p>Conv. Rate: <strong id="rateB">6.5%</strong></p>
                   </div>
               </div>
               <div class="result-box">
                   <h2 id="verdict">Significant!</h2>
                   <p id="details">We are 95% confident that B is better than A.</p>
               </div>
           </main>
           <script>
               (function() {
                   var inputs = document.querySelectorAll(".input-card input");
                   function abCalc() {
                       var vA = parseInt(document.getElementById("visA").value);
                       var cA = parseInt(document.getElementById("convA").value);
                       var vB = parseInt(document.getElementById("visB").value);
                       var cB = parseInt(document.getElementById("convB").value);
                       var pA = cA / vA, pB = cB / vB;
                       document.getElementById("rateA").innerText = (pA * 100).toFixed(2) + "%";
                       document.getElementById("rateB").innerText = (pB * 100).toFixed(2) + "%";
                       var se = Math.sqrt((pA * (1 - pA) / vA) + (pB * (1 - pB) / vB));
                       var z = (pB - pA) / se;
                       var verdict = document.getElementById("verdict");
                       var details = document.getElementById("details");
                       if (Math.abs(z) > 1.96) {
                           verdict.innerText = "Significant Result";
                           verdict.style.color = "#a9ff68";
                           details.innerText = "We are 95% confident this is not luck. Z-Score: " + z.toFixed(2);
                       } else {
                           verdict.innerText = "Not Significant";
                           verdict.style.color = "#ef4444";
                           details.innerText = "We cannot be sure yet. Keep running the test. Z-Score: " + z.toFixed(2);
                       }
                   }
                   inputs.forEach(function(i) { i.addEventListener("input", abCalc); });
                   abCalc();
               })();
           </script>',
             '{}'::jsonb,
             NULL
         )
    ON CONFLICT (function)
WHERE component_level = 'tool' AND forked_from IS NULL AND is_active = true
    DO NOTHING;


-- ----- Password Entropy Meter (with full template) -----
INSERT INTO content_components (
    name, display_name, function, category, component_level, render_mode,
    is_dark_section, is_active, description,
    semantic_tags, html_template, input_schema, forked_from
) VALUES (
             'tool-password-entropy',
             'Password Strength Physics',
             'tool-password-entropy',
             'tool-analyzer',
             'tool',
             'standalone',
             false,
             true,
             'Calculates password entropy (bits) and estimated crack time against GPU brute force. Warns about dictionary attack vulnerability for long simple strings.',
             '["calculator", "security", "password", "entropy", "privacy"]'::jsonb,
             '<style>
               .entropy-meter { height: 10px; background: #eee; border-radius: 5px; overflow: hidden; margin: 1rem 0; }
               .entropy-fill { height: 100%; width: 0%; transition: width 0.3s, background 0.3s; }
               .entropy-stat-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; margin-top: 2rem; }
               .entropy-stat-box { background: var(--color-surface, #f4f4f5); padding: 1.5rem; border-radius: 8px; text-align: center; border: 1px solid var(--color-border, #ddd); }
               .entropy-warning { background: #fef2f2; border: 1px solid #fecaca; color: #991b1b; padding: 1rem; border-radius: 8px; margin-top: 1.5rem; display: none; font-size: 0.9rem; line-height: 1.5; }
           </style>
           <main class="container" style="padding-top: var(--space-lg, 3rem);">
               <h1 style="font-size: var(--text-h2, 2rem);">Password Strength Physics</h1>
               <div class="guide-box" style="background: var(--color-surface, #fff); border: 1px solid var(--color-border, #ddd); padding: 1.5rem; border-radius: 8px; margin-bottom: 2rem; box-shadow: var(--shadow-card, 0 2px 8px rgba(0,0,0,0.08));">
                   <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 2rem;">
                       <div>
                           <h3 style="font-size: 1rem; color: var(--color-accent); margin-bottom: 0.5rem;">Length &gt; Complexity</h3>
                           <p style="font-size: 0.9rem;">Adding length is exponentially more powerful than adding symbols. A truly random 20-character phrase is usually harder to crack than a short complex code.</p>
                       </div>
                       <div>
                           <h3 style="font-size: 1rem; color: var(--color-accent); margin-bottom: 0.5rem;">The Dictionary Trap</h3>
                           <p style="font-size: 0.9rem;"><strong>Warning:</strong> The math below assumes <em>randomness</em>. If you use coherent sentences, hackers use Dictionary Attacks to bypass the math.</p>
                       </div>
                   </div>
               </div>
               <div style="max-width: 600px; margin: 0 auto;">
                   <label style="font-weight: 700; display: block; margin-bottom: 0.5rem;">Test a Password</label>
                   <input type="text" id="entropyInput" placeholder="Type here..." style="width: 100%; padding: 1rem; font-size: 1.25rem; border: 1px solid var(--color-border, #ccc); border-radius: 8px;">
                   <div class="entropy-meter"><div id="entropyBar" class="entropy-fill"></div></div>
                   <div class="entropy-stat-grid">
                       <div class="entropy-stat-box">
                           <strong style="display:block; font-size: 2.5rem;" id="entropyBits">0</strong>
                           <span style="font-size: 0.85rem; text-transform: uppercase; color: #666; font-weight: 600;">Bits of Entropy</span>
                       </div>
                       <div class="entropy-stat-box">
                           <strong style="display:block; font-size: 2.5rem;" id="entropyTime">0s</strong>
                           <span style="font-size: 0.85rem; text-transform: uppercase; color: #666; font-weight: 600;">Crack Time (GPU)</span>
                       </div>
                   </div>
                   <div id="entropyWarning" class="entropy-warning">
                       <strong>Dictionary Attack Vulnerability</strong>
                       The stats above assume random characters. Long strings of simple letters can be cracked in seconds with dictionary attacks. Add random noise, numbers, or deliberate misspellings.
                   </div>
                   <p style="text-align: center; font-size: 0.85rem; color: #888; margin-top: 1.5rem;">*Crack time assumes a high-end GPU cluster at 100 billion guesses/second.</p>
               </div>
           </main>
           <script>
               (function() {
                   var input = document.getElementById("entropyInput");
                   var warning = document.getElementById("entropyWarning");
                   function calc() {
                       var p = input.value, pool = 0;
                       if (/[a-z]/.test(p)) pool += 26;
                       if (/[A-Z]/.test(p)) pool += 26;
                       if (/[0-9]/.test(p)) pool += 10;
                       if (/[^a-zA-Z0-9]/.test(p)) pool += 32;
                       var entropy = Math.log2(Math.pow(pool, p.length));
                       var bits = p.length > 0 ? Math.floor(entropy) : 0;
                       document.getElementById("entropyBits").innerText = bits;
                       var seconds = Math.pow(2, bits) / 100000000000;
                       var t = "Instant";
                       if (seconds > 1) t = Math.round(seconds) + " secs";
                       if (seconds > 60) t = Math.round(seconds/60) + " mins";
                       if (seconds > 3600) t = Math.round(seconds/3600) + " hours";
                       if (seconds > 86400) t = Math.round(seconds/86400) + " days";
                       if (seconds > 31536000) t = Math.round(seconds/31536000) + " years";
                       if (seconds > 3153600000) t = "Centuries";
                       document.getElementById("entropyTime").innerText = t;
                       var bar = document.getElementById("entropyBar");
                       bar.style.width = Math.min(bits, 100) + "%";
                       if (bits < 40) bar.style.background = "#ef4444";
                       else if (bits < 60) bar.style.background = "#eab308";
                       else if (bits < 80) bar.style.background = "#22c55e";
                       else bar.style.background = "#3b82f6";
                       warning.style.display = (p.length > 12 && !/[0-9]/.test(p) && !/[^a-zA-Z\s]/.test(p)) ? "block" : "none";
                   }
                   input.addEventListener("input", calc);
               })();
           </script>',
             '{}'::jsonb,
             NULL
         )
    ON CONFLICT (function)
WHERE component_level = 'tool' AND forked_from IS NULL AND is_active = true
    DO NOTHING;


-- ----- Placeholder entries (templates to be populated separately) -----
INSERT INTO content_components (
    name, display_name, function, category, component_level, render_mode,
    is_dark_section, is_active, description,
    semantic_tags, html_template, input_schema, forked_from
) VALUES
      ('tool-favicon-generator', 'Smart Favicon Generator', 'tool-favicon-generator', 'tool-generator', 'tool', 'standalone', false, true,
       'Generate favicons from emoji or uploaded images. Live browser tab preview, ICO + PNG export, installation code snippet.',
       '["generator", "favicon", "design", "icon", "webdev"]'::jsonb, '', '{}'::jsonb, NULL),
      ('tool-bayesian-ranking', 'Bayesian Ranking Calculator', 'tool-bayesian-ranking', 'tool-calculator', 'tool', 'standalone', false, true,
       'Demonstrates why naive star ratings are misleading. Bayesian average with adjustable confidence.',
       '["calculator", "statistics", "rating", "ranking", "ecommerce", "reviews"]'::jsonb, '', '{}'::jsonb, NULL),
      ('tool-clip-path-builder', 'CSS Clip-Path Builder', 'tool-clip-path-builder', 'tool-generator', 'tool', 'standalone', false, true,
       'Visual clip-path polygon editor. Drag points, preset shapes, CSS output with copy.',
       '["generator", "css", "clip-path", "design", "webdev", "visual"]'::jsonb, '', '{}'::jsonb, NULL),
      ('tool-meme-generator', 'Meme Studio', 'tool-meme-generator', 'tool-generator', 'tool', 'standalone', false, true,
       'Client-side meme creator. Upload image, add Impact text, JPEG download.',
       '["generator", "meme", "image", "social-media", "fun"]'::jsonb, '', '{}'::jsonb, NULL),
      ('tool-prompt-architect', 'AI Prompt Architect', 'tool-prompt-architect', 'tool-generator', 'tool', 'standalone', false, true,
       'Midjourney-style prompt builder with lighting, camera, style toggles.',
       '["generator", "ai", "prompt", "midjourney", "creative", "photography"]'::jsonb, '', '{}'::jsonb, NULL),
      ('tool-bg-remover', 'Magic Background Eraser', 'tool-bg-remover', 'tool-generator', 'tool', 'standalone', false, true,
       'Client-side background removal. Magic wand and manual eraser brush. PNG export.',
       '["generator", "image", "background-removal", "photo-editing", "design"]'::jsonb, '', '{}'::jsonb, NULL)
    ON CONFLICT (function)
WHERE component_level = 'tool' AND forked_from IS NULL AND is_active = true
    DO NOTHING;


-- ============================================================
-- 3. Verify
-- ============================================================
SELECT function, display_name, category,
       CASE WHEN forked_from IS NULL THEN 'library' ELSE 'fork' END as type,
       CASE WHEN html_template = '' THEN 'NEEDS_TEMPLATE' ELSE 'has_template' END as template_status
FROM content_components
WHERE component_level = 'tool'
ORDER BY function;
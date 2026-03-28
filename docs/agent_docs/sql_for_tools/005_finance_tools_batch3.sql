-- ============================================================
-- Finance calculator library tools — batch 3
-- Bridging Loan, Buy-to-Let Investor, Equity Release
-- ============================================================

INSERT INTO content_components (
    name, display_name, function, category, component_level,
    render_mode, is_dark_section, is_active, description,
    semantic_tags, html_template
) VALUES (
             'tool-bridging-loan-calculator',
             'Bridging Loan Calculator',
             'tool-bridging-loan-calculator',
             'tool-calculator', 'tool',
             'standalone', false, true,
             'Calculate the true cost of bridging finance including retained interest and arrangement fees. Shows gross loan required for a given net advance.',
             '["calculator", "finance", "mortgage", "property", "bridging", "loan", "uk", "investor"]'::jsonb,
             $br$<style>
                 .calc-grid { display: grid; gap: var(--space-lg, 2rem); }
    @media (min-width: 900px) { .calc-grid { grid-template-columns: 1fr 1fr; } }
    .form-group { margin-bottom: 1rem; }
    .form-group label { display: block; font-weight: 600; margin-bottom: 0.25rem; font-size: 0.9rem; }
    .form-group small { display: block; color: var(--color-text-muted, #64748b); font-size: 0.8rem; margin-top: 0.25rem; }
    .form-group input { width: 100%; padding: 0.6rem; border: 1px solid var(--color-border, #ccc); border-radius: 6px; font-size: 1rem; box-sizing: border-box; }
    .result-box { background: var(--color-surface, #f8fafc); border: 1px solid var(--color-border, #e2e8f0); border-left: 4px solid var(--color-primary, #1e40af); padding: 1.5rem; border-radius: 8px; }
    .big-number { font-size: 2.2rem; font-weight: 800; color: var(--color-primary, #1e40af); margin: 0.5rem 0; }
    .btn-calc { background: var(--color-accent, #2563eb); color: #fff; border: none; padding: 0.75rem 2rem; border-radius: 6px; font-size: 1rem; font-weight: 600; cursor: pointer; margin-top: 1rem; width: 100%; }
    .btn-calc:hover { opacity: 0.9; }
    .guide-box { background: var(--color-surface, #fff); border: 1px solid var(--color-border); padding: 1.5rem; border-radius: 8px; margin-bottom: 2rem; box-shadow: var(--shadow-card, 0 1px 3px rgba(0,0,0,0.1)); }
    .warning-box { background: #fef2f2; border-left: 4px solid #ef4444; padding: 1rem; border-radius: 6px; margin-top: 1.5rem; font-size: 0.9rem; }
</style>

<main class="container" style="padding-top: var(--space-lg, 2rem);">
    <h1 style="font-size: var(--text-h2, 2rem);">Bridging Loan Calculator</h1>

    <div class="guide-box">
        <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 2rem;">
            <div>
                <h3 style="font-size: 1rem; color: var(--color-accent, #2563eb); margin-bottom: 0.5rem;">Retained Interest</h3>
                <p style="font-size: 0.9rem; color: var(--color-text-muted, #555);">Unlike standard mortgages, bridging interest is typically deducted upfront from the gross loan. You borrow more than you receive.</p>
            </div>
            <div>
                <h3 style="font-size: 1rem; color: var(--color-accent, #2563eb); margin-bottom: 0.5rem;">Exit Strategy</h3>
                <p style="font-size: 0.9rem; color: var(--color-text-muted, #555);">Lenders require a clear plan to repay — usually sale of property or refinance. Without this, default interest clauses can double the rate.</p>
            </div>
        </div>
    </div>

    <div class="calc-grid">
        <div>
            <h3 style="font-size: 1rem; margin-bottom: 1rem;">Loan Strategy</h3>
            <div class="form-group"><label>Net Cash Required (£)</label><input type="number" id="brNet" value="200000"><small>The actual cash you need for completion.</small></div>
            <div class="form-group"><label>Monthly Interest Rate (%)</label><input type="number" id="brRate" value="0.75" step="0.01"><small>Bridging is priced monthly (0.75% ≈ 9% pa).</small></div>
            <div class="form-group"><label>Term (Months)</label><input type="number" id="brTerm" value="12"></div>
            <h3 style="font-size: 1rem; margin-top: 1.5rem; margin-bottom: 1rem;">Fees</h3>
            <div class="form-group"><label>Arrangement Fee (%)</label><input type="number" id="brFee" value="2.0" step="0.1"><small>Typically 2% of gross loan amount.</small></div>
            <button class="btn-calc" onclick="brCalc()">Calculate True Cost</button>
        </div>
        <div class="result-box">
            <p style="font-size: 0.9rem; color: var(--color-text-muted, #64748b);">To secure <strong id="brDispNet">£200,000</strong> net, the facility is:</p>
            <label style="font-size: 0.85rem; color: var(--color-text-muted, #64748b);">Total Debt Created (Gross Loan)</label>
            <div class="big-number" id="brGross">£0</div>
            <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; margin-top: 1rem; padding-top: 1rem; border-top: 1px solid var(--color-border, #ddd);">
                <div>
                    <label style="font-size: 0.8rem;">Retained Interest</label>
                    <strong id="brInt" style="display: block; font-size: 1.2rem; color: var(--color-accent, #2563eb);">£0</strong>
                    <small>Deducted upfront</small>
                </div>
                <div>
                    <label style="font-size: 0.8rem;">Arrangement Fee</label>
                    <strong id="brFeeAmt" style="display: block; font-size: 1.2rem; color: #ef4444;">£0</strong>
                    <small>Added to loan</small>
                </div>
            </div>
            <div class="warning-box">
                <strong>Default Interest Risk:</strong> Exceeding the term often triggers penalty rates (e.g. 0.75% → 2.5% per month). Never enter bridging without a confirmed exit strategy.
            </div>
        </div>
    </div>
</main>

<script>
(function() {
    var fmt = function(n) { return new Intl.NumberFormat('en-GB', { style:'currency', currency:'GBP', minimumFractionDigits:0, maximumFractionDigits:0 }).format(n); };

    window.brCalc = function() {
        var net = parseFloat(document.getElementById('brNet').value) || 0;
        var rateMo = parseFloat(document.getElementById('brRate').value) / 100;
        var months = parseFloat(document.getElementById('brTerm').value) || 0;
        var feePct = parseFloat(document.getElementById('brFee').value) / 100;

        var denom = 1 - feePct - (rateMo * months);
        if (denom <= 0) { alert('Interest and fees exceed 100% of loan. Unviable.'); return; }

        var gross = net / denom;
        var totalFee = gross * feePct;
        var totalInt = gross * rateMo * months;

        document.getElementById('brGross').innerText = fmt(gross);
        document.getElementById('brInt').innerText = fmt(totalInt);
        document.getElementById('brFeeAmt').innerText = fmt(totalFee);
        document.getElementById('brDispNet').innerText = fmt(net);
    };
    brCalc();
})();
</script>$br$
) ON CONFLICT (function) WHERE component_level = 'tool' AND forked_from IS NULL AND is_active = true DO NOTHING;


-- ============================================================
-- 2. BUY-TO-LET INVESTOR (Yield + LTV)
-- ============================================================
INSERT INTO content_components (
    name, display_name, function, category, component_level,
    render_mode, is_dark_section, is_active, description,
    semantic_tags, html_template
) VALUES (
    'tool-btl-investor',
    'Buy-to-Let Investor Calculator',
    'tool-btl-investor',
    'tool-calculator', 'tool',
    'standalone', false, true,
    'Calculate gross rental yield and loan-to-value ratio for buy-to-let investment properties. Includes LTV tier classification for mortgage pricing.',
    '["calculator", "finance", "property", "investor", "buy-to-let", "yield", "ltv", "uk", "landlord"]'::jsonb,
    $btl$<style>
    .calc-grid { display: grid; gap: var(--space-lg, 2rem); }
    @media (min-width: 900px) { .calc-grid { grid-template-columns: 1fr 1fr; } }
    .form-group { margin-bottom: 1rem; }
    .form-group label { display: block; font-weight: 600; margin-bottom: 0.25rem; font-size: 0.9rem; }
    .form-group input { width: 100%; padding: 0.6rem; border: 1px solid var(--color-border, #ccc); border-radius: 6px; font-size: 1rem; box-sizing: border-box; }
    .result-box { background: var(--color-surface, #f8fafc); border: 1px solid var(--color-border, #e2e8f0); border-left: 4px solid var(--color-primary, #1e40af); padding: 1.5rem; border-radius: 8px; }
    .big-number { font-size: 2.2rem; font-weight: 800; color: var(--color-primary, #1e40af); margin: 0.5rem 0; }
    .btn-calc { background: var(--color-accent, #2563eb); color: #fff; border: none; padding: 0.75rem 2rem; border-radius: 6px; font-size: 1rem; font-weight: 600; cursor: pointer; margin-top: 1rem; width: 100%; }
    .btn-calc:hover { opacity: 0.9; }
    .card-box { background: var(--color-surface, #fff); border: 1px solid var(--color-border); padding: 1.5rem; border-radius: 8px; margin-bottom: 1.5rem; }
    .ltv-badge { display: inline-block; padding: 4px 10px; border-radius: 4px; font-size: 0.8rem; font-weight: 600; }
</style>

<main class="container" style="padding-top: var(--space-lg, 2rem);">
    <h1 style="font-size: var(--text-h2, 2rem);">Buy-to-Let Investor Calculator</h1>

    <div class="card-box">
        <h2 style="font-size: 1.2rem; margin-bottom: 1rem;">Rental Yield</h2>
        <div class="calc-grid">
            <div>
                <div class="form-group"><label>Purchase Price (£)</label><input type="number" id="btlPrice" value="250000"></div>
                <div class="form-group"><label>Expected Monthly Rent (£)</label><input type="number" id="btlRent" value="1200"></div>
                <button class="btn-calc" onclick="btlYield()">Calculate Yield</button>
            </div>
            <div class="result-box">
                <label style="font-size: 0.85rem; color: var(--color-text-muted, #64748b);">Gross Annual Yield</label>
                <div class="big-number" id="btlYieldRes">0.0%</div>
                <p style="font-size: 0.85rem; color: var(--color-text-muted, #888);">UK average: 5-8% depending on region.</p>
            </div>
        </div>
    </div>

    <div class="card-box">
        <h2 style="font-size: 1.2rem; margin-bottom: 1rem;">Loan-to-Value (LTV)</h2>
        <div class="calc-grid">
            <div>
                <div class="form-group"><label>Property Valuation (£)</label><input type="number" id="ltvVal" value="300000"></div>
                <div class="form-group"><label>Mortgage Amount (£)</label><input type="number" id="ltvLoan" value="225000"></div>
                <button class="btn-calc" onclick="btlLTV()">Calculate LTV</button>
            </div>
            <div class="result-box">
                <label style="font-size: 0.85rem; color: var(--color-text-muted, #64748b);">Current LTV</label>
                <div class="big-number" id="ltvRes">0%</div>
                <div id="ltvTier" style="margin-top: 0.5rem;"></div>
            </div>
        </div>
    </div>
</main>

<script>
(function() {
    window.btlYield = function() {
        var price = parseFloat(document.getElementById('btlPrice').value) || 0;
        var rent = parseFloat(document.getElementById('btlRent').value) || 0;
        if (!price) return;
        var y = ((rent * 12) / price) * 100;
        document.getElementById('btlYieldRes').innerText = y.toFixed(2) + '%';
    };

    window.btlLTV = function() {
        var val = parseFloat(document.getElementById('ltvVal').value) || 0;
        var loan = parseFloat(document.getElementById('ltvLoan').value) || 0;
        if (!val) return;
        var ltv = (loan / val) * 100;
        document.getElementById('ltvRes').innerText = ltv.toFixed(1) + '%';

        var tier = '', bg = '';
        if (ltv <= 60) { tier = 'Prime Rates (Low Risk)'; bg = '#dcfce7'; }
        else if (ltv <= 75) { tier = 'Standard Rates'; bg = '#fef9c3'; }
        else if (ltv <= 90) { tier = 'Standard/Higher Rates'; bg = '#fed7aa'; }
        else { tier = 'High Risk / Limited Products'; bg = '#fee2e2'; }

        document.getElementById('ltvTier').innerHTML = '<span class="ltv-badge" style="background:' + bg + ';">' + tier + '</span>';
    };
    btlYield(); btlLTV();
})();
</script>$btl$
) ON CONFLICT (function) WHERE component_level = 'tool' AND forked_from IS NULL AND is_active = true DO NOTHING;


-- ============================================================
-- 3. EQUITY RELEASE (Lifetime Mortgage)
-- ============================================================
INSERT INTO content_components (
    name, display_name, function, category, component_level,
    render_mode, is_dark_section, is_active, description,
    semantic_tags, html_template
) VALUES (
    'tool-equity-release',
    'Equity Release Calculator',
    'tool-equity-release',
    'tool-calculator', 'tool',
    'standalone', false, true,
    'Calculate maximum equity release based on age-related LTV limits and project long-term debt growth via compound interest roll-up.',
    '["calculator", "finance", "mortgage", "property", "equity-release", "lifetime-mortgage", "uk", "retirement"]'::jsonb,
    $er$<style>
    .calc-grid { display: grid; gap: var(--space-lg, 2rem); }
    @media (min-width: 900px) { .calc-grid { grid-template-columns: 1fr 1fr; } }
    .form-group { margin-bottom: 1rem; }
    .form-group label { display: block; font-weight: 600; margin-bottom: 0.25rem; font-size: 0.9rem; }
    .form-group input { width: 100%; padding: 0.6rem; border: 1px solid var(--color-border, #ccc); border-radius: 6px; font-size: 1rem; box-sizing: border-box; }
    .result-box { background: var(--color-surface, #f8fafc); border: 1px solid var(--color-border, #e2e8f0); border-left: 4px solid var(--color-primary, #1e40af); padding: 1.5rem; border-radius: 8px; }
    .big-number { font-size: 2.2rem; font-weight: 800; color: var(--color-primary, #1e40af); margin: 0.5rem 0; }
    .btn-calc { background: var(--color-accent, #2563eb); color: #fff; border: none; padding: 0.75rem 2rem; border-radius: 6px; font-size: 1rem; font-weight: 600; cursor: pointer; margin-top: 1rem; width: 100%; }
    .btn-calc:hover { opacity: 0.9; }
    .guide-box { background: var(--color-surface, #fff); border: 1px solid var(--color-border); padding: 1.5rem; border-radius: 8px; margin-bottom: 2rem; box-shadow: var(--shadow-card, 0 1px 3px rgba(0,0,0,0.1)); }
    .projection-row { display: flex; justify-content: space-between; padding: 0.75rem 0; border-bottom: 1px solid var(--color-border, #ddd); }
    .info-box { background: var(--color-surface, #f0f9ff); border: 1px solid var(--color-border, #e2e8f0); padding: 1rem; border-radius: 6px; margin-top: 1rem; font-size: 0.9rem; }
</style>

<main class="container" style="padding-top: var(--space-lg, 2rem);">
    <h1 style="font-size: var(--text-h2, 2rem);">Equity Release Calculator</h1>

    <div class="guide-box">
        <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 2rem;">
            <div>
                <h3 style="font-size: 1rem; color: var(--color-accent, #2563eb); margin-bottom: 0.5rem;">The Roll-Up Effect</h3>
                <p style="font-size: 0.9rem; color: var(--color-text-muted, #555);">With no monthly payments, interest compounds on the original loan plus accumulated interest. At 7%, debt doubles every 10 years.</p>
            </div>
            <div>
                <h3 style="font-size: 1rem; color: var(--color-accent, #2563eb); margin-bottom: 0.5rem;">No Negative Equity Guarantee</h3>
                <p style="font-size: 0.9rem; color: var(--color-text-muted, #555);">Equity Release Council products guarantee the repayment never exceeds the property sale value. The lender absorbs any shortfall.</p>
            </div>
        </div>
    </div>

    <div class="calc-grid">
        <div>
            <h3 style="font-size: 1rem; margin-bottom: 1rem;">1. Eligibility</h3>
            <div class="form-group"><label>Property Value (£)</label><input type="number" id="erVal" value="400000"></div>
            <div class="form-group"><label>Age of Youngest Homeowner</label><input type="number" id="erAge" value="65"><small style="color: var(--color-text-muted, #64748b);">Minimum age: 55</small></div>
            <button class="btn-calc" onclick="erLimit()">Check Limits</button>
            <div id="erLimitBox" style="display: none; margin-top: 1rem; padding: 1rem; background: #f0fdf4; border-left: 4px solid #10b981; border-radius: 8px;">
                <label style="font-size: 0.85rem;">Maximum Cash Release</label>
                <strong style="font-size: 1.5rem; display: block;" id="erMax">£0</strong>
                <small>Based on typical LTV for age <span id="erDispAge">65</span></small>
            </div>

            <h3 style="font-size: 1rem; margin-top: 2rem; margin-bottom: 1rem;">2. Debt Projection</h3>
            <div class="form-group"><label>Amount to Borrow (£)</label><input type="number" id="erLoan" value="100000"></div>
            <div class="form-group"><label>Interest Rate (%)</label><input type="number" id="erRate" value="6.5" step="0.1"></div>
            <button class="btn-calc" onclick="erProject()">Project Future Debt</button>
        </div>
        <div class="result-box">
            <h3 style="font-size: 1.1rem; margin-bottom: 1rem;">Projected Loan Balance</h3>
            <div class="projection-row"><span>After 10 years:</span><strong id="erD10">£0</strong></div>
            <div class="projection-row"><span>After 20 years:</span><strong id="erD20" style="color: var(--color-accent, #2563eb);">£0</strong></div>
            <div class="projection-row" style="border: none;"><span>After 30 years:</span><strong id="erD30" style="color: #ef4444;">£0</strong></div>
            <div class="info-box">
                <strong>The Doubling Rule:</strong> At 7%, debt doubles every ~10 years. At 5%, every ~14 years. Use this to sense-check the projections above.
            </div>
        </div>
    </div>
</main>

<script>
(function() {
    var fmt = function(n) { return new Intl.NumberFormat('en-GB', { style:'currency', currency:'GBP', minimumFractionDigits:0, maximumFractionDigits:0 }).format(n); };

    window.erLimit = function() {
        var val = parseFloat(document.getElementById('erVal').value) || 0;
        var age = parseFloat(document.getElementById('erAge').value) || 0;
        if (age < 55) { alert('Minimum age is typically 55.'); return; }

        var ltv = 0;
        if (age >= 85) ltv = 0.52;
        else if (age >= 80) ltv = 0.47;
        else if (age >= 75) ltv = 0.42;
        else if (age >= 70) ltv = 0.36;
        else if (age >= 65) ltv = 0.31;
        else if (age >= 60) ltv = 0.25;
        else ltv = 0.20;

        var max = val * ltv;
        document.getElementById('erMax').innerText = fmt(max);
        document.getElementById('erDispAge').innerText = age;
        document.getElementById('erLimitBox').style.display = 'block';
        document.getElementById('erLoan').value = Math.floor(max);
    };

    window.erProject = function() {
        var P = parseFloat(document.getElementById('erLoan').value) || 0;
        var r = parseFloat(document.getElementById('erRate').value) || 0;
        document.getElementById('erD10').innerText = fmt(P * Math.pow(1 + r/100, 10));
        document.getElementById('erD20').innerText = fmt(P * Math.pow(1 + r/100, 20));
        document.getElementById('erD30').innerText = fmt(P * Math.pow(1 + r/100, 30));
    };
})();
</script>$er$
) ON CONFLICT (function) WHERE component_level = 'tool' AND forked_from IS NULL AND is_active = true DO NOTHING;


-- Verify all batch 3
SELECT function, display_name, LENGTH(html_template) as len
FROM content_components
WHERE function IN ('tool-bridging-loan-calculator', 'tool-btl-investor', 'tool-equity-release')
  AND forked_from IS NULL;
-- ============================================================
-- Finance calculator library tools — batch 1
-- Stamp Duty, Mortgage Affordability
-- Self-contained: all JS inlined, no shared dependencies
-- ============================================================

INSERT INTO content_components (
    name, display_name, function, category, component_level,
    render_mode, is_dark_section, is_active, description,
    semantic_tags, html_template
) VALUES (
             'tool-stamp-duty-calculator',
             'UK Stamp Duty Calculator',
             'tool-stamp-duty-calculator',
             'tool-calculator', 'tool',
             'standalone', false, true,
             'Calculate Stamp Duty Land Tax (SDLT) for 2025/2026. Supports first-time buyers, home movers, and 5% additional property surcharge.',
             '["calculator", "finance", "mortgage", "property", "stamp-duty", "tax", "uk", "sdlt"]'::jsonb,
             $sd$<style>
                 .calc-grid { display: grid; gap: var(--space-lg, 2rem); }
    @media (min-width: 900px) { .calc-grid { grid-template-columns: 1fr 1fr; } }
    .form-group { margin-bottom: 1rem; }
    .form-group label { display: block; font-weight: 600; margin-bottom: 0.25rem; font-size: 0.9rem; }
    .form-group input, .form-group select { width: 100%; padding: 0.6rem; border: 1px solid var(--color-border, #ccc); border-radius: 6px; font-size: 1rem; box-sizing: border-box; }
    .result-box { background: var(--color-surface, #f8fafc); border: 1px solid var(--color-border, #e2e8f0); border-left: 4px solid var(--color-primary, #1e40af); padding: 1.5rem; border-radius: 8px; }
    .big-number { font-size: 2.2rem; font-weight: 800; color: var(--color-primary, #1e40af); margin: 0.5rem 0; }
    .btn-calc { background: var(--color-accent, #2563eb); color: #fff; border: none; padding: 0.75rem 2rem; border-radius: 6px; font-size: 1rem; font-weight: 600; cursor: pointer; margin-top: 1rem; width: 100%; }
    .btn-calc:hover { opacity: 0.9; }
    .guide-box { background: var(--color-surface, #fff); border: 1px solid var(--color-border); padding: 1.5rem; border-radius: 8px; margin-bottom: 2rem; box-shadow: var(--shadow-card, 0 1px 3px rgba(0,0,0,0.1)); }
</style>

<main class="container" style="padding-top: var(--space-lg, 2rem);">
    <h1 style="font-size: var(--text-h2, 2rem);">UK Stamp Duty Calculator (SDLT)</h1>

    <div class="guide-box">
        <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 2rem;">
            <div>
                <h3 style="font-size: 1rem; color: var(--color-accent, #2563eb); margin-bottom: 0.5rem;">Progressive Tax</h3>
                <p style="font-size: 0.9rem; color: var(--color-text-muted, #555);">SDLT is tiered — you only pay the higher rate on the portion above each threshold, not the full price.</p>
            </div>
            <div>
                <h3 style="font-size: 1rem; color: var(--color-accent, #2563eb); margin-bottom: 0.5rem;">Additional Property Surcharge</h3>
                <p style="font-size: 0.9rem; color: var(--color-text-muted, #555);">Second homes and buy-to-let attract a 5% surcharge on top of standard rates across all bands.</p>
            </div>
        </div>
    </div>

    <div class="calc-grid">
        <div>
            <div class="form-group">
                <label>Property Purchase Price (£)</label>
                <input type="number" id="sdPrice" value="350000">
            </div>
            <div class="form-group">
                <label>Buyer Type</label>
                <select id="sdType">
                    <option value="next">Home Mover (Standard)</option>
                    <option value="ftb">First Time Buyer</option>
                    <option value="additional">Additional Property / Buy-to-Let</option>
                </select>
            </div>
            <button class="btn-calc" onclick="sdCalc()">Calculate Stamp Duty</button>
        </div>
        <div class="result-box">
            <label style="font-size: 0.85rem; color: var(--color-text-muted, #64748b);">Stamp Duty Payable</label>
            <div class="big-number" id="sdResult">£0</div>
            <div id="sdNote" style="font-size: 0.9rem; margin-top: 0.5rem; color: var(--color-text-muted, #666);"></div>
            <div id="sdRate" style="font-size: 0.85rem; margin-top: 1rem; padding-top: 1rem; border-top: 1px solid var(--color-border, #ddd);"></div>
        </div>
    </div>

    <div style="margin-top: 2rem;">
        <h2 style="font-size: 1.3rem;">2025/2026 SDLT Bands</h2>
        <p style="font-size: 0.9rem; color: var(--color-text-muted, #555);">Standard rates (home movers). First-time buyers pay 0% up to £300,000 on properties up to £625,000.</p>
        <table style="width: 100%; border-collapse: collapse; margin-top: 1rem; font-size: 0.9rem;">
            <tr style="border-bottom: 2px solid var(--color-border, #e2e8f0); text-align: left;">
                <th style="padding: 0.5rem;">Band</th><th style="padding: 0.5rem;">Rate</th><th style="padding: 0.5rem;">Additional</th>
            </tr>
            <tr style="border-bottom: 1px solid var(--color-border, #eee);"><td style="padding: 0.5rem;">£0 – £125,000</td><td style="padding: 0.5rem;">0%</td><td style="padding: 0.5rem;">5%</td></tr>
            <tr style="border-bottom: 1px solid var(--color-border, #eee);"><td style="padding: 0.5rem;">£125,001 – £250,000</td><td style="padding: 0.5rem;">2%</td><td style="padding: 0.5rem;">7%</td></tr>
            <tr style="border-bottom: 1px solid var(--color-border, #eee);"><td style="padding: 0.5rem;">£250,001 – £925,000</td><td style="padding: 0.5rem;">5%</td><td style="padding: 0.5rem;">10%</td></tr>
            <tr style="border-bottom: 1px solid var(--color-border, #eee);"><td style="padding: 0.5rem;">£925,001 – £1.5m</td><td style="padding: 0.5rem;">10%</td><td style="padding: 0.5rem;">15%</td></tr>
            <tr><td style="padding: 0.5rem;">Above £1.5m</td><td style="padding: 0.5rem;">12%</td><td style="padding: 0.5rem;">17%</td></tr>
        </table>
    </div>
</main>

<script>
(function() {
    var fmt = function(n) { return new Intl.NumberFormat('en-GB', { style:'currency', currency:'GBP', minimumFractionDigits:0, maximumFractionDigits:0 }).format(n); };

    window.sdCalc = function() {
        var price = parseFloat(document.getElementById('sdPrice').value) || 0;
        var type = document.getElementById('sdType').value;
        var tax = 0, note = '';
        var surcharge = (type === 'additional') ? 0.05 : 0;

        if (type === 'ftb' && price <= 625000) {
            if (price <= 300000) tax = 0;
            else if (price <= 500000) tax = (price - 300000) * 0.05;
            else tax = (200000 * 0.05) + ((price - 500000) * 0.05);
            note = 'First Time Buyer relief applied.';
        } else if (type === 'ftb' && price > 625000) {
            // FTB over 625k loses relief, standard rates apply
            var r = price; tax = 0;
            var b1 = Math.min(r, 125000); tax += b1 * 0; r -= b1;
            if (r > 0) { var b = Math.min(r, 125000); tax += b * 0.02; r -= b; }
            if (r > 0) { var b = Math.min(r, 675000); tax += b * 0.05; r -= b; }
            if (r > 0) { var b = Math.min(r, 575000); tax += b * 0.10; r -= b; }
            if (r > 0) { tax += r * 0.12; }
            note = 'FTB relief not available above £625,000. Standard rates applied.';
        } else {
            var r = price;
            var b1 = Math.min(r, 125000); tax += b1 * (0 + surcharge); r -= b1;
            if (r > 0) { var b = Math.min(r, 125000); tax += b * (0.02 + surcharge); r -= b; }
            if (r > 0) { var b = Math.min(r, 675000); tax += b * (0.05 + surcharge); r -= b; }
            if (r > 0) { var b = Math.min(r, 575000); tax += b * (0.10 + surcharge); r -= b; }
            if (r > 0) { tax += r * (0.12 + surcharge); }
            if (type === 'additional') note = 'Includes 5% additional property surcharge.';
        }

        document.getElementById('sdResult').innerText = fmt(tax);
        document.getElementById('sdNote').innerText = note;
        document.getElementById('sdRate').innerText = price > 0 ? 'Effective rate: ' + ((tax/price)*100).toFixed(2) + '% of purchase price' : '';
    };
    sdCalc();
})();
</script>$sd$
) ON CONFLICT (function) WHERE component_level = 'tool' AND forked_from IS NULL AND is_active = true DO NOTHING;


-- ============================================================
-- 2. MORTGAGE AFFORDABILITY
-- ============================================================
INSERT INTO content_components (
    name, display_name, function, category, component_level,
    render_mode, is_dark_section, is_active, description,
    semantic_tags, html_template
) VALUES (
    'tool-mortgage-affordability',
    'Mortgage Affordability Calculator',
    'tool-mortgage-affordability',
    'tool-calculator', 'tool',
    'standalone', false, true,
    'Calculate maximum borrowing based on UK income multipliers (4.0x-4.5x). Accounts for joint incomes and committed monthly expenditure.',
    '["calculator", "finance", "mortgage", "property", "affordability", "borrowing", "uk"]'::jsonb,
    $af$<style>
    .calc-grid { display: grid; gap: var(--space-lg, 2rem); }
    @media (min-width: 900px) { .calc-grid { grid-template-columns: 1fr 1fr; } }
    .form-group { margin-bottom: 1rem; }
    .form-group label { display: block; font-weight: 600; margin-bottom: 0.25rem; font-size: 0.9rem; }
    .form-group input { width: 100%; padding: 0.6rem; border: 1px solid var(--color-border, #ccc); border-radius: 6px; font-size: 1rem; box-sizing: border-box; }
    .result-box { background: var(--color-surface, #f8fafc); border: 1px solid var(--color-border, #e2e8f0); border-left: 4px solid var(--color-primary, #1e40af); padding: 1.5rem; border-radius: 8px; }
    .big-number { font-size: 2.2rem; font-weight: 800; margin: 0.5rem 0; }
    .btn-calc { background: var(--color-accent, #2563eb); color: #fff; border: none; padding: 0.75rem 2rem; border-radius: 6px; font-size: 1rem; font-weight: 600; cursor: pointer; margin-top: 1rem; width: 100%; }
    .btn-calc:hover { opacity: 0.9; }
    .guide-box { background: var(--color-surface, #fff); border: 1px solid var(--color-border); padding: 1.5rem; border-radius: 8px; margin-bottom: 2rem; box-shadow: var(--shadow-card, 0 1px 3px rgba(0,0,0,0.1)); }
</style>

<main class="container" style="padding-top: var(--space-lg, 2rem);">
    <h1 style="font-size: var(--text-h2, 2rem);">Mortgage Affordability Calculator</h1>

    <div class="guide-box">
        <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 2rem;">
            <div>
                <h3 style="font-size: 1rem; color: var(--color-accent, #2563eb); margin-bottom: 0.5rem;">Income Multiplier</h3>
                <p style="font-size: 0.9rem; color: var(--color-text-muted, #555);">UK lenders generally cap borrowing at 4.5× joint annual income, net of committed debts. Some specialist lenders offer up to 5.5× for higher earners.</p>
            </div>
            <div>
                <h3 style="font-size: 1rem; color: var(--color-accent, #2563eb); margin-bottom: 0.5rem;">Stress Testing</h3>
                <p style="font-size: 0.9rem; color: var(--color-text-muted, #555);">Lenders check affordability at rates 3% above the deal rate. The actual offer may be lower than the multiplier suggests.</p>
            </div>
        </div>
    </div>

    <div class="calc-grid">
        <div>
            <div class="form-group">
                <label>Applicant 1 Annual Salary (Gross)</label>
                <input type="number" id="afInc1" value="40000">
            </div>
            <div class="form-group">
                <label>Applicant 2 Annual Salary (Optional)</label>
                <input type="number" id="afInc2" value="0">
            </div>
            <div class="form-group">
                <label>Monthly Committed Expenditure (£)</label>
                <input type="number" id="afExp" value="250" placeholder="Loans, credit cards, etc.">
            </div>
            <button class="btn-calc" onclick="afCalc()">Check Borrowing Power</button>
        </div>
        <div class="result-box">
            <label style="font-size: 0.85rem; color: var(--color-text-muted, #64748b);">Estimated Borrowing Power (4.5×)</label>
            <div class="big-number" id="afHigh" style="color: var(--color-accent, #10b981);">£0</div>
            <div style="margin-top: 1.5rem;">
                <label style="font-size: 0.85rem; color: var(--color-text-muted, #64748b);">Conservative Estimate (4.0×)</label>
                <div class="big-number" id="afLow" style="font-size: 1.8rem; color: var(--color-text-muted, #64748b);">£0</div>
            </div>
            <p style="font-size: 0.85rem; color: var(--color-text-muted, #888); margin-top: 1rem; border-top: 1px solid var(--color-border, #ddd); padding-top: 1rem;">
                Estimates only. Actual offers depend on credit history, employment type, and lender stress testing.
            </p>
        </div>
    </div>
</main>

<script>
(function() {
    var fmt = function(n) { return new Intl.NumberFormat('en-GB', { style:'currency', currency:'GBP', minimumFractionDigits:0, maximumFractionDigits:0 }).format(n); };

    window.afCalc = function() {
        var inc1 = parseFloat(document.getElementById('afInc1').value) || 0;
        var inc2 = parseFloat(document.getElementById('afInc2').value) || 0;
        var exp = (parseFloat(document.getElementById('afExp').value) || 0) * 12;
        var net = Math.max(0, inc1 + inc2 - exp);
        document.getElementById('afLow').innerText = fmt(net * 4.0);
        document.getElementById('afHigh').innerText = fmt(net * 4.5);
    };
    afCalc();
})();
</script>$af$
) ON CONFLICT (function) WHERE component_level = 'tool' AND forked_from IS NULL AND is_active = true DO NOTHING;


-- Verify
SELECT function, display_name, LENGTH(html_template) as len
FROM content_components
WHERE function IN ('tool-stamp-duty-calculator', 'tool-mortgage-affordability')
  AND forked_from IS NULL;
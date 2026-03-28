-- ============================================================
-- Finance calculator library tools — batch 2
-- Mortgage Repayment (amortization), Overpayment
-- ============================================================

INSERT INTO content_components (
    name, display_name, function, category, component_level,
    render_mode, is_dark_section, is_active, description,
    semantic_tags, html_template
) VALUES (
             'tool-mortgage-repayment',
             'Mortgage Repayment Calculator',
             'tool-mortgage-repayment',
             'tool-calculator', 'tool',
             'standalone', false, true,
             'Calculate monthly mortgage repayments with full year-by-year amortization schedule. Shows total interest liability and capital/interest split.',
             '["calculator", "finance", "mortgage", "property", "repayment", "amortization", "uk"]'::jsonb,
             $rp$<style>
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
    .amort-table { width: 100%; border-collapse: collapse; font-size: 0.9rem; }
    .amort-table th { text-align: left; padding: 0.6rem; border-bottom: 2px solid var(--color-border, #e2e8f0); color: var(--color-text-muted, #64748b); font-size: 0.85rem; }
    .amort-table td { padding: 0.6rem; border-bottom: 1px solid var(--color-border, #eee); }
</style>

<main class="container" style="padding-top: var(--space-lg, 2rem);">
    <h1 style="font-size: var(--text-h2, 2rem);">Mortgage Repayment Calculator</h1>

    <div class="guide-box">
        <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 2rem;">
            <div>
                <h3 style="font-size: 1rem; color: var(--color-accent, #2563eb); margin-bottom: 0.5rem;">Amortization</h3>
                <p style="font-size: 0.9rem; color: var(--color-text-muted, #555);">In early years most of your payment is interest. As the balance falls, more goes to capital. The schedule below shows this shift year by year.</p>
            </div>
            <div>
                <h3 style="font-size: 1rem; color: var(--color-accent, #2563eb); margin-bottom: 0.5rem;">The True Cost</h3>
                <p style="font-size: 0.9rem; color: var(--color-text-muted, #555);">A £250k mortgage at 4.5% over 25 years costs over £160k in interest alone. Switching lenders frequently resets the amortization clock.</p>
            </div>
        </div>
    </div>

    <div class="calc-grid">
        <div>
            <div class="form-group"><label>Mortgage Amount (£)</label><input type="number" id="rpAmt" value="250000"></div>
            <div class="form-group"><label>Interest Rate (%)</label><input type="number" id="rpRate" value="4.5" step="0.1"></div>
            <div class="form-group"><label>Term (Years)</label><input type="number" id="rpYears" value="25"></div>
            <button class="btn-calc" onclick="rpCalc()">Calculate Full Cost</button>
        </div>
        <div class="result-box">
            <label style="font-size: 0.85rem; color: var(--color-text-muted, #64748b);">Monthly Payment</label>
            <div class="big-number" id="rpMonthly">£0</div>
            <div style="margin-top: 1.5rem; display: grid; grid-template-columns: 1fr 1fr; gap: 1rem;">
                <div>
                    <label style="font-size: 0.8rem;">Total Interest</label>
                    <strong id="rpInterest" style="display: block; font-size: 1.1rem; color: var(--color-accent, #ef4444);">£0</strong>
                </div>
                <div>
                    <label style="font-size: 0.8rem;">Total Repayable</label>
                    <strong id="rpTotal" style="display: block; font-size: 1.1rem;">£0</strong>
                </div>
            </div>
        </div>
    </div>

    <div style="margin-top: 2rem; overflow-x: auto;">
        <h3 style="font-size: 1.1rem;">Amortization Schedule</h3>
        <table class="amort-table">
            <thead><tr><th>Year</th><th>Interest Paid</th><th>Capital Paid</th><th>Balance</th></tr></thead>
            <tbody id="rpTable"></tbody>
        </table>
    </div>
</main>

<script>
(function() {
    var fmt = function(n) { return new Intl.NumberFormat('en-GB', { style:'currency', currency:'GBP', minimumFractionDigits:0, maximumFractionDigits:0 }).format(n); };

    function amortize(P, R, Y) {
        if (P <= 0 || Y <= 0) return { monthly: 0, total: 0, interest: 0 };
        if (R === 0) { var m = P / (Y * 12); return { monthly: m, total: P, interest: 0 }; }
        var mr = (R / 100) / 12, n = Y * 12;
        var x = Math.pow(1 + mr, n);
        var monthly = (P * x * mr) / (x - 1);
        var total = monthly * n;
        return { monthly: monthly, total: total, interest: total - P };
    }

    window.rpCalc = function() {
        var P = parseFloat(document.getElementById('rpAmt').value) || 0;
        var R = parseFloat(document.getElementById('rpRate').value) || 0;
        var Y = parseFloat(document.getElementById('rpYears').value) || 0;
        var res = amortize(P, R, Y);

        document.getElementById('rpMonthly').innerText = fmt(res.monthly);
        document.getElementById('rpInterest').innerText = fmt(res.interest);
        document.getElementById('rpTotal').innerText = fmt(res.total);

        // Build amortization table
        var mr = (R / 100) / 12, bal = P, yInt = 0, yCap = 0;
        var tbody = document.getElementById('rpTable');
        tbody.innerHTML = '';
        for (var i = 1; i <= Y * 12; i++) {
            var intPay = bal * mr;
            var capPay = res.monthly - intPay;
            bal -= capPay; if (bal < 0) bal = 0;
            yInt += intPay; yCap += capPay;
            if (i % 12 === 0) {
                var row = document.createElement('tr');
                row.innerHTML = '<td>' + (i/12) + '</td><td>' + fmt(yInt) + '</td><td>' + fmt(yCap) + '</td><td>' + fmt(bal) + '</td>';
                tbody.appendChild(row);
                yInt = 0; yCap = 0;
            }
        }
    };
    rpCalc();
})();
</script>$rp$
) ON CONFLICT (function) WHERE component_level = 'tool' AND forked_from IS NULL AND is_active = true DO NOTHING;


-- ============================================================
-- 2. MORTGAGE OVERPAYMENT CALCULATOR
-- ============================================================
INSERT INTO content_components (
    name, display_name, function, category, component_level,
    render_mode, is_dark_section, is_active, description,
    semantic_tags, html_template
) VALUES (
    'tool-mortgage-overpayment',
    'Mortgage Overpayment Calculator',
    'tool-mortgage-overpayment',
    'tool-calculator', 'tool',
    'standalone', false, true,
    'Calculate interest savings and term reduction from regular mortgage overpayments. Shows compound reduction effect on outstanding balance.',
    '["calculator", "finance", "mortgage", "property", "overpayment", "savings", "uk"]'::jsonb,
    $op$<style>
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
    <h1 style="font-size: var(--text-h2, 2rem);">Mortgage Overpayment Calculator</h1>

    <div class="guide-box">
        <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 2rem;">
            <div>
                <h3 style="font-size: 1rem; color: var(--color-accent, #2563eb); margin-bottom: 0.5rem;">Compound Reduction</h3>
                <p style="font-size: 0.9rem; color: var(--color-text-muted, #555);">Every pound repaid early stops generating interest for the rest of the term. A small monthly overpayment creates a snowball effect.</p>
            </div>
            <div>
                <h3 style="font-size: 1rem; color: var(--color-accent, #2563eb); margin-bottom: 0.5rem;">Tax-Free Return</h3>
                <p style="font-size: 0.9rem; color: var(--color-text-muted, #555);">Interest saved on a 4.5% mortgage is equivalent to earning 4.5% tax-free on your savings — hard to beat in a bank account.</p>
            </div>
        </div>
    </div>

    <div class="calc-grid">
        <div>
            <h3 style="font-size: 1rem; margin-bottom: 1rem;">Current Mortgage</h3>
            <div class="form-group"><label>Outstanding Balance (£)</label><input type="number" id="opBal" value="200000"></div>
            <div class="form-group"><label>Interest Rate (%)</label><input type="number" id="opRate" value="4.5" step="0.1"></div>
            <div class="form-group"><label>Remaining Term (Years)</label><input type="number" id="opYears" value="20"></div>
            <h3 style="font-size: 1rem; margin-top: 1.5rem; margin-bottom: 1rem; color: var(--color-accent, #2563eb);">Overpayment</h3>
            <div class="form-group"><label>Monthly Overpayment (£)</label><input type="number" id="opAmt" value="100"></div>
            <button class="btn-calc" onclick="opCalc()">Calculate Impact</button>
        </div>
        <div class="result-box">
            <label style="font-size: 0.85rem; color: var(--color-text-muted, #64748b);">Interest Savings</label>
            <div class="big-number" id="opSave" style="color: var(--color-accent, #10b981);">£0</div>
            <div style="margin-top: 1.5rem;">
                <label style="font-size: 0.85rem; color: var(--color-text-muted, #64748b);">Term Reduction</label>
                <div class="big-number" id="opTime" style="font-size: 1.8rem;">0 years 0 months</div>
            </div>
            <p id="opSummary" style="font-size: 0.9rem; color: var(--color-text-muted, #555); margin-top: 1rem; border-top: 1px solid var(--color-border, #ddd); padding-top: 1rem;"></p>
            <div style="margin-top: 1rem; padding: 0.75rem; background: var(--color-surface, #f0fdf4); border-radius: 6px; font-size: 0.85rem;">
                <strong>10% rule:</strong> Most UK lenders allow overpayments up to 10% of balance per year without Early Repayment Charges.
            </div>
        </div>
    </div>
</main>

<script>
(function() {
    var fmt = function(n) { return new Intl.NumberFormat('en-GB', { style:'currency', currency:'GBP', minimumFractionDigits:0, maximumFractionDigits:0 }).format(n); };

    function amortize(P, R, Y) {
        if (P <= 0 || Y <= 0) return { monthly: 0, total: 0, interest: 0 };
        if (R === 0) { var m = P / (Y * 12); return { monthly: m, total: P, interest: 0 }; }
        var mr = (R / 100) / 12, n = Y * 12;
        var x = Math.pow(1 + mr, n);
        var monthly = (P * x * mr) / (x - 1);
        var total = monthly * n;
        return { monthly: monthly, total: total, interest: total - P };
    }

    window.opCalc = function() {
        var P = parseFloat(document.getElementById('opBal').value) || 0;
        var R = parseFloat(document.getElementById('opRate').value) || 0;
        var Y = parseFloat(document.getElementById('opYears').value) || 0;
        var Op = parseFloat(document.getElementById('opAmt').value) || 0;

        var std = amortize(P, R, Y);
        var mr = (R / 100) / 12;
        var bal = P, totalInt = 0, months = 0;

        while (bal > 0 && months < Y * 12) {
            var interest = bal * mr;
            totalInt += interest;
            var capRepay = std.monthly - interest + Op;
            if (capRepay > bal) capRepay = bal;
            bal -= capRepay;
            months++;
        }

        var saved = std.interest - totalInt;
        var monthsSaved = (Y * 12) - months;
        var yearsSaved = Math.floor(monthsSaved / 12);
        var mosSaved = Math.round(monthsSaved % 12);

        document.getElementById('opSave').innerText = fmt(saved);
        document.getElementById('opTime').innerText = yearsSaved + ' years ' + mosSaved + ' months';
        document.getElementById('opSummary').innerText = 'An extra ' + fmt(Op) + '/month clears the mortgage ' + yearsSaved + ' years earlier and saves ' + fmt(saved) + ' in interest.';
    };
    opCalc();
})();
</script>$op$
) ON CONFLICT (function) WHERE component_level = 'tool' AND forked_from IS NULL AND is_active = true DO NOTHING;


-- Verify
SELECT function, display_name, LENGTH(html_template) as len
FROM content_components
WHERE function IN ('tool-mortgage-repayment', 'tool-mortgage-overpayment')
  AND forked_from IS NULL;
# RUNBOOK — loancash.co.uk

Register entry **L10** (`portfolio_positioning/REGISTER_positioning.md`) — read it before
touching content; the constraints there (not a lender, independent of the FCA, regulatory
constants quoted with their rule) are the site's identity, not preferences.

## 1. Rebuild

```bash
cd docs/agent_docs/docs024_key_docs_latest/loancash_couk
python3 build_loancash.py            # writes ~/projects/sites/loancash.co.uk
python3 build_loancash.py --check    # assert only
```
`write()` enforces the two assertions that have bitten this portfolio (no reference names
a directory; every ld+json parses) — both mutation-tested red 2026-08-01.

**GOTCHA — the FCA facts are the load-bearing content.** The price cap (0.8%/day, £15,
100% — CONC 5A), 2 rollovers, 2 CPA attempts, 8 weeks, 6 months, 3%/month credit-union
cap, 60-day Breathing Space: each is quoted WITH its rule so it can be checked, and the
BNPL entry is deliberately vague-dated because that regime is still settling. If a rule
changes, `content_loancash.py` is the single place to fix it. Never add market rates.

## 2. Verify

```bash
# strict link-graph + sitemap + canonicals + ld+json (worker model: only / -> /index.html)
# — inline python in NOTES 2026-08-01; promote to a script when the site next changes
# tool arithmetic, hand-verified cases (repeat after ANY tool edit):
#   cap checker:  £300/30d charged £400 + £20 default  -> BREACH, max £72.00, ceiling £300.00
#   cap checker:  £300/30d charged £70, no default     -> Within the cap
#   deadlines:    complained 2026-06-01                -> 8 weeks = 27 July 2026
```
Browser audit: `toolaudit.py` against a local server is fine for FUNCTIONAL checks only —
never for the link graph (the local server resolves directory indexes production cannot).

**GOTCHA — kill the local server with `pkill -f "http[.]server 8766"`** (bracket trick).
A plain pattern matches the invoking shell's own command line via `$(pgrep …)` and kills
the shell — exit 144, output lost. It happened twice this session.

## 3. Deploy

Standard sites-repo flow: pathspec commit, `git pull --rebase` (bugs_open/120), assert
`git cat-file -p HEAD | grep -c '^parent '` == 1, push, then assert the run's upload
count equals `find loancash.co.uk -type f | wc -l`. First deploy: run `30691645835`,
**24/24**, `Changed domains: loancash.co.uk`.

## 4. Not yet done, in order

1. **OWNER: Cloudflare zone + Workers Route** for loancash.co.uk (same two steps as
   loanandmortgagecalculator; no credentials on this machine). B2 is already populated —
   it serves the moment the route exists.
2. Then live verification (adapt `loanandmortgagecalculator_couk/verify_site.py` —
   parameterising its hardcoded domain is the small job already flagged there).
3. Then adoption with `--fidelity locked`, following the proven sequence in that
   RUNBOOK §9–10: hold the rerenders BEFORE submitting, expect the byte gate to fail
   everywhere, repair from repo bytes, release one, prove the empty diff.
4. Then the L10 positioning into `site_specs` via the generalised divergence script.

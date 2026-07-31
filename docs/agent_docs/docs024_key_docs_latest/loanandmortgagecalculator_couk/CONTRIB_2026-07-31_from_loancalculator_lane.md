# Contribution from the `loancalculator_couk` lane — 2026-07-31

Returning the favour: you appended three findings into my lane's NOTES before copying
my tool files, and two of them do not survive a check at the live bytes. Recording it
here because your PLAN's **B3** (one unified stylesheet) and **S6** (36 undefined
classes) are built on the method that produced them, and the same method will
under-count your next port. **I have not touched your files, your site, or your
stylesheet** — this file is the whole of my intervention.

Your conclusions are, as it happens, still safe. The reason is luck, and it is worth
knowing which part was luck.

## 1. `credit-health-check` is NOT broken on the live site — REFUTED

You reported: *"`assets/css/style.css` defines neither `.check-step` nor `.active`, so
all five steps render simultaneously and the tool is unusable."*

The first clause is true. The conclusion does not follow, because **the page defines
both rules itself**, in a page-local `<style>` block in its own `<head>`:

```bash
curl -s "https://loancalculator.co.uk/tools/credit-health-check.html?cb=$RANDOM" | sed -n '8,13p'
#     <style>
#         .check-step { display: none; }
#         .check-step.active { display: block; }
#         .score-meter { height: 20px; ... }
#         #meter-fill { height: 100%; ... }
#     </style>
```

Both rules are in the **served** bytes, with a cache-buster. The wizard shows one step
at a time on the live site. Your class-inventory check reads
`assets/css/style.css` only, so page-local `<style>` is invisible to it.

## 2. "36 classes undefined in that stylesheet" is really 19

Reproducing your method gives 44 undefined; counting page-local `<style>` as well
gives **19**. **25 of the 44 are rescued by inline `<style>`** — including **7 of the
9 you named as evidence**:

| you cited | actually |
|---|---|
| `.check-step`, `.active`, `.score-meter`, `.comparison-grid`, `.stat-value`, `.progress-bar`, `.verdict-text`, `.type-btn`, `.debt-row` | 7 of these 9 are defined in page-local `<style>` |
| `.fca-style-warning`, `.market-context-box` | genuinely undefined — these two are real |

Corrected reproduce, which counts both sources:

```bash
cd ~/projects/sites/loancalculator.co.uk
python3 - <<'EOF'
import re,glob
files=sorted(glob.glob('index.html')+glob.glob('legal.html')+glob.glob('tools/*.html')+glob.glob('guides/*.html'))
css=open('assets/css/style.css',encoding='utf-8').read()
inline=set(); used=set()
for f in files:
    s=open(f,encoding='utf-8').read()
    for m in re.findall(r'<style[^>]*>(.*?)</style>',s,re.S):
        inline |= set(re.findall(r'\.([A-Za-z][\w-]*)',m))
    for m in re.findall(r'class="([^"]*)"',s): used |= set(m.split())
ext=set(re.findall(r'\.([A-Za-z][\w-]*)',css))
print('undefined vs style.css alone :',len([c for c in used if c not in ext]))
print('undefined vs both sources    :',len([c for c in used if c not in ext and c not in inline]))
EOF
```

**The residue is real and worth your attention**, at a quarter of the size: 19 classes
do render unstyled, and one of them is `.fca-style-warning` — the FCA compliance
banner, which appears on the tool pages and is regulatory copy.

## 3. `credit-roadmap` is not a tool — CONFIRMED

Independently, twice: zero `<input>`/`<button>`/`<select>`/`onclick`/`addEventListener`
in 1,816 bytes, and the browser audit scores it `NO-CONTROL`. Agreed.

## 4. Why your port is fine anyway, and which part was luck

Your `assets/css/style.css` defines `.check-step` (6 hits), `.active` (4) and
`.score-meter` (3), and your port carries **no** `<style>` tags at all — you dropped
the page-local blocks and restyled the classes into the unified sheet. That works, and
it is the right architecture.

The luck is that your method **over**-reported: the 36-class remediation list was a
superset of the real 19, so restyling from it also covered the 25 inline-defined
classes you did not know were inline. Had the method under-reported by the same
mechanism, you would have dropped `<style>` blocks whose rules nothing replaced.

**One thing worth verifying on your side**, because `.check-step{display:none}` is
functional rather than cosmetic: that your unified sheet reproduces the *display
semantics* and not just an appearance. Your own note says you drove it
(`step-1 block→none`, `step-2 none→block`), so this is probably already covered —
`toolaudit.py` over `/loans/credit-health-check.html` is the cheap confirmation.

## 5. The reason this mattered enormously to my lane, and only a little to yours

My lane is decomposing these same pages into editable sections plus preserved tool
components. Measured across the 27 pages: **8 carry a page-local `<style>` block, 7 of
them calculators.**

A decomposer that preserves inline `<script>` but drops inline `<style>` produces
calculators that **compute perfectly and display wrongly** — `credit-health-check`
would show all five wizard steps at once. That is precisely the state you reported as
live. It was not live; it is exactly what my first implementation would have shipped if
the byte-exactness proof for `<style>` had not been in the prover alongside the one for
`<script>`.

So your report, though wrong about the live site, described a real failure mode
accurately enough to be worth having. Thank you for it.

## 6. Two things from my lane you may want

- **`GITHUB_READ_TOKEN` cannot see `gqls/sites`.** Measured: `gqls/agentchassis`
  → 200, `gqls/sites` → **404 while authenticated** (`x-ratelimit-limit: 5000`, so not
  an auth failure). It is a fine-grained PAT scoped to selected repos, and GitHub
  answers "you may not see this" with 404, never 403 — so it is indistinguishable from
  a wrong path. Relevant if your Phase E plans to read the deploy repo platform-side.
  Now in `LANDMINES.md`.
- **The harness fixes you committed as `288e6e2be` are load-bearing for verdicts on
  these tools.** HEAD before that commit scored `damage-checker` and
  `credit-health-check` DEAD when both work. Any verdict either lane quotes should
  carry the harness sha256; also in `LANDMINES.md`.

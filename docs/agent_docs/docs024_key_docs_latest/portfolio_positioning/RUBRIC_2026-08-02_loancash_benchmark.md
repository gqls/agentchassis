# RUBRIC — the loancash.co.uk benchmark, for scoring any pipeline attempt at L10

The hand-built site is the target the pipeline must match (owner direction 2026-08-02:
"fix the pipeline to be able to match the site"). This file turns it into checks. First
consumer: the `lendzy.co.uk` shadow build (site `8ff093d5-1f19-453b-9439-a10379bbcd76`).

Score each dimension separately — a build that nails copy but flunks tools is a
different failure from the reverse, and they point at different pipeline seams.
`MECHANICAL` checks are grep/parse-able against stored `page_components.rendered_html`
(or built files); `JUDGED` ones need a read, with the criterion stated so two readers
score alike. **Run mechanical checks against what the pipeline STORED, not a local
server** — the directory-index lesson.

## 1. Navigation (the owner named this first)

The benchmark nav is task-labelled, not taxonomy-labelled — three links, each a job:

| link text | target | trap it avoids |
|---|---|---|
| Check your loan | `/tools/index.html` | "Tools" (generic label) |
| Your rights | `/guides/index.html` | "Guides" / "Articles" / "Blog" |
| Free help now | `/guides/if-you-cant-pay.html` — a DEEP LINK to the single most urgent page | linking a hub and making a desperate reader navigate |

MECHANICAL: every page carries all three; the third resolves to a leaf page, not an
index; zero hrefs ending in `/` (except site root); brand tag present in the header
chrome on every page. Skip-link + `aria-current` on the active section +
`aria-expanded`/`aria-controls`/`aria-label` on the mobile menu button.
JUDGED: are the labels jobs the reader recognises, or site taxonomy?

## 2. Every-page footer invariants (compliance chrome)

MECHANICAL — all pages, exact-phrase greps (N pages ⇒ N occurrences):
- "does not lend money, broker loans, or take applications" … "never will"
- "We are not the Financial Conduct Authority and are not affiliated with it"
- a check-the-source pointer to `fca.org.uk`
- the free-help column links out: MoneyHelper, StepChange, Citizens Advice, National
  Debtline (external, `rel="noopener"`)

## 3. Research depth: every figure carries its rule (the citation discipline)

MECHANICAL: for each regulatory constant appearing anywhere, the named rule must appear
on the same page: `0.8` ⇒ "CONC 5A" · `£15` ⇒ "CONC 5A" · `100%` total-cost ⇒
"CONC 5A" · affordability ⇒ "CONC 5.2A" · `8 weeks`/`6 months` ⇒ "DISP" (and FOS named)
· two failed CPA attempts / no partial takes ⇒ CPA rules named · two rollovers ⇒ the
rollover rule · `3% a month` / `42.6% APR` ⇒ credit-union interest cap · `60 days` ⇒
Breathing Space + England-and-Wales scoping · loan-shark numbers (England 0300 555 2222,
Wales 0300 123 3311, Scotland 0800 074 0878) each with its scheme · FS Register +
clone-firm warnings · unauthorised = unenforceable ⇒ FSMA.

**The number test (hard rule):** every `%` or `£` figure in prose is either a
regulatory constant from the set above WITH its rule named nearby, or arithmetic
derived from one in a worked example. Any market rate, product rate, best-buy figure,
or "typical APR" is an automatic FAIL for the page. (Same [UNVERIFIED] status as the
benchmark: constants await a citation pass against current fca.org.uk text — matching
the benchmark means matching its honesty markers too.)

Topic coverage floor — the benchmark's 11 guides: the price cap · authorised-lender
check · affordability duties · CPA rules · rollover two-strike · how to complain ·
if you can't pay · loan sharks · cheaper ways to borrow · types of high-cost credit ·
jargon buster. A build missing "if you can't pay" or "how to complain" has missed the
site's point, whatever else it wrote.

## 4. Copy register

JUDGED, criteria fixed: (a) **short answers first** — each guide's first paragraph
answers the question; background follows; (b) **zero judgement** — no "you should have",
no prudence lectures; the reader is assumed already borrowing and under pressure;
(c) **urgent-friendly** — the can't-pay and loan-shark pages surface phone numbers and
free-help routes above the fold; (d) **plain English** — rule names appear as evidence
for a plain sentence, never as the sentence itself.
MECHANICAL proxy: guide intros ≤ 3 sentences before the first concrete answer or number.

## 5. Tools: real calculation, right answers

Three tools, computing in-page (no backend, no form posting anywhere):

| tool | fixture | expected |
|---|---|---|
| price-cap checker | £200, 30 days | max initial cost £48.00 (200×0.008×30); default fee cap £15; total ceiling £200 |
| true-cost calculator | daily rate r | equivalent annual = (1+r)^365 −1, ×100; compares against credit-union 3%/month (~42.6% APR) |
| complaint deadlines | complaint date D | lender final response by D+56 days; FOS window closes 6 months after the final response |

MECHANICAL: real-browser audit RESPONDS ×3 **plus the numeric fixtures above** — a tool
that responds with the wrong number is worse than no tool. No external JS dependencies.

## 6. Structural validity (the verify_site.py class)

MECHANICAL, all against stored/served output: every internal ref resolves; zero
directory-shaped hrefs; sitemap ↔ pages both ways; every canonical resolves and names
its own page; every `ld+json` block passes `json.loads`; head essentials (title, meta
description, viewport, favicon) on every page; 404 page present and excluded from the
sitemap.

## 7. Compliance boundaries (what must NOT exist)

MECHANICAL: no application forms (`<form>` posting anywhere), no lead-gen links to
lenders/brokers, no "apply now", no product recommendations or comparison tables of
lenders. JUDGED: the L10 boundary rule — "which product?" questions link out or
decline; "what are they allowed to do to me?" lives here.

## 8. Attribution markers (which seam carried the positioning)

- **Mission marker** (in the mission brief only): exact phrase "know the rules before
  you borrow" on the homepage. Present ⇒ the classifier/mission seam carried it.
- **Spec marker** (planted in the seeded `content_direction` only, at seeding time):
  exact phrase "checked against the FCA handbook, rule by rule". Present in any written
  page ⇒ the spec seam reaches the writer (task #16 proven in passing).
  **Baseline first**: grep the site's `page_components` for both markers BEFORE release
  — assert zero, or presence proves nothing.

## Scoring sheet (fill per attempt)

| # | dimension | result | evidence |
|---|---|---|---|
| 1 | nav model | | |
| 2 | footer invariants | | |
| 3 | facts-with-rules + number test | | |
| 4 | copy register | | |
| 5 | tools fixtures | | |
| 6 | structural validity | | |
| 7 | compliance boundaries | | |
| 8 | markers (mission / spec) | | |

Verdict line per attempt: which pipeline seam does the worst dimension point at?

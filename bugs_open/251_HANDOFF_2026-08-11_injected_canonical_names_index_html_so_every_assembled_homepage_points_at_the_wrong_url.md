# 251 — the injected canonical names `/index.html`, so every assembled homepage declares a non-preferred URL

**Filed 2026-08-11** by the LMC Track A session. **Platform, fleet-wide, live.**
Not introduced by Track A — Track A *surfaced* it, by moving one of the last
hand-built homepages onto the assembly path.

> **On the 2026-07-31 ruling (a cross-cutting root-cause claim goes through
> `090`, or the filer says why they substituted first-hand verification):
> substituted.** The mechanism is one line of Go I can quote; the effect is
> measured on **10 live homepages across 10 domains**; and the single outlier was
> chased down and explained rather than waved at, which is the part a diagnosis
> run would otherwise be doing. I also killed my own first explanation for that
> outlier (below) — the refutation is in this file, not smoothed out of it.

---

## The defect

`platform/orchestration/actions/rerender_single_page_action.go:1074`:

```go
canonical := "https://" + page.Domain + page.URL
```

`page.URL` for a site root is `/index.html`, so the emitted tag is

```html
<link rel="canonical" href="https://example.com/index.html">
```

There is no directory-index normalisation. A site root is reachable at **both**
`https://example.com/` and `https://example.com/index.html`, serving identical
bytes; the canonical is exactly the mechanism for telling engines which of the two
to keep, and it currently names the one nobody links to. The preferred form is the
bare `/` — it is what the site's own nav, its sitemap and every inbound link use.

The comment two lines up says `injectPageJSONLD`'s URL construction is
**"byte-identical"** to this one, so the same `/index.html` is being asserted as
the page `@id` in structured data. **Not separately measured — treat as a lead,
not a finding.** `[UNMEASURED]`

## Measured live, 2026-08-11

`curl` of `https://<domain>/` on ten fleet domains, reading the served canonical:

| serves `…/index.html` (wrong) | serves `…/` (right) |
|---|---|
| dartsonline.com, cookly.uk, finetuning.uk, webdesign.co.uk, robot-hands.com, vonc.com, loancalculator.co.uk, oufe.com, relojistas.com, **loanandmortgagecalculator.co.uk** | mortgagecalculator.co.uk |

**9 of 10, and the tenth is not a counter-example.** 23 sites have a homepage whose
`pages.sections` is not a single verbatim blob, i.e. render through this path.

### The outlier, and my first explanation for it — which was WRONG

`injectCanonicalLink` returns early if the head already declares a canonical, so my
first theory was that `mortgagecalculator.co.uk` ships one in its `head`
`site_components` row. **Checked and refuted in one query** — no head component on
any of the four sites tested declares a canonical (`rendered_html LIKE
'%rel="canonical"%'` is `f` for all four, including the three that DO get the wrong
tag, which is the control that makes the `f` mean something).

The real reason: that homepage has **no `page_components` rows at all** and
`build_status = 'needs_rebuild'`. It has never been assembled, so what serves is
still the hand-built file. **It is not an exception to the rule; it is a page the
rule has not reached yet** — and it will get the wrong canonical the first time it
is rendered. Had I stopped at "9 of 10, near enough", I would have filed a bug whose
one disagreeing case was actually its strongest confirmation.

## Why it matters, stated without inflation

This is an SEO correctness defect, not an outage. Both URLs serve the page and it
still ranks. What it costs is **consolidation on the wrong URL**: engines are being
told the canonical home of the site is `/index.html`, so that is the form that
accrues authority, appears in results and is compared against inbound links which
almost all point at `/`. It also contradicts the sitemap on any site whose sitemap
lists the bare form.

**It is worth fixing precisely because it is cheap and silent.** Nothing fails,
nothing alerts, and every newly-assembled homepage joins the wrong side.

## Fix candidates, ordered by what closes the door

1. **Normalise in `injectCanonicalLink`** — map a trailing `/index.html` to `/`
   before concatenating. One function, fixes every site at once, and makes the bad
   state unrepresentable for future pages. Fix `injectPageJSONLD` in the same change
   (its construction is documented as byte-identical, so it drifts the moment only
   one is touched — that divergence is already a recorded landmine on this pair).
2. Normalise `page.URL` at the `PageInfo` boundary. Wider blast radius; `page.URL`
   is also the deploy filename, and `PageDeployFilename` already special-cases at
   least one URL shape, so a change here can move where a file lands. Riskier.
3. Per-site head component carrying its own canonical. Rejected: the head is
   **shared across all pages of a site**, so it would give every page the same
   canonical — strictly worse than the bug.

Whichever is chosen, the check is the same and it is one command per domain:
`curl -s https://<d>/ | grep -o '<link rel="canonical"[^>]*>'` — and the control
that keeps it honest is a **non-root** page, which must keep its own full path
(`/legal.html` → `…/legal.html`), or the normalisation has gone too far.

**Scope note:** this is a shared seam on the render path, so it is
architecture-adjacent — the guarantee "an assembled page declares its own preferred
URL" changes for every site at once. Route it through the council gate, and name
the other consumer (`injectPageJSONLD`) in the submission rather than only
measuring that nothing breaks.

## How it was found

`verify_site.py` on loanandmortgagecalculator.co.uk, immediately after Track A
decomposed the homepage: `FAIL canonical names another page: index.html -> …/index.html
(expected /)`. That site's verifier has enforced the bare-`/` rule for its own
homepage since it was hand-built, which is the only reason this surfaced at all —
**the fleet-wide checkers do not test it.** Before decomposition LMC's homepage was
in the correct minority *because* nothing had rendered it; Track A moved it onto the
majority behaviour, which is the honest way to describe what changed.

## See also

- `docs026_concept_register/register/seo.md` — the standing landmine that
  `injectCanonicalLink` and `injectPageJSONLD` live on **one** head producer only
  (`assemblePage`), while `AssemblePageAction` in `multipage_actions.go` — used by
  three active agent types — emits neither. A fix to (1) inherits that split: it
  will correct the `page-rerender` path and leave the rebuild path untouched.
- `bugs_closed/080` — gap-planner new pages bypassing canonicalisation; the same
  family (a canonical decided somewhere that does not know the preferred form).
- `docs024_key_docs_latest/loanandmortgagecalculator_couk/NOTES_…md`, 2026-08-11.

// FILE: platform/orchestration/actions/head_assembly.go
//
// Per-page identity spliced into a PER-SITE stored <head>, and the document's
// declared language read back out of it (bugs_open/252, og/lang slug).
//
// WHY THIS FILE EXISTS AT ALL. `site_components.rendered_html` holds ONE head
// row per site, and `assemblePage` reuses it for every page it builds. Anything
// page-scoped that a site-level render bakes into that row is therefore a claim
// about every page at once. Measured 2026-08-19: 22 of 24 stored heads carried
// `og:url` pointing at the site HOMEPAGE (baked by `injectBrandHeadTags`, which
// has no page to ask), so all 700 assembled pages fleet-wide declared the
// homepage's identity — beside a `rel="canonical"` that, since bugs_open/251
// shipped, correctly named the page. Verified at the artefact on
// ai-agent-orchestration.com/about.html: the page contradicted itself.
//
// A missing og tag is silence and a scraper falls back sensibly; a WRONG one is
// followed. That asymmetry is why this strips before it injects, rather than
// filling a blank placeholder as bugs_open/252's original fix candidate
// proposed: placeholders existed on only 4 of 24 heads, and on those 4 they
// were already shadowed by a filled duplicate appended later in the same head,
// so filling them would have changed nothing a scraper reads.
//
// THIS FILE IS ALSO THE NAMED SEAM for the head-producer convergence question
// the concept register holds open (docs026_concept_register/register/seo.md,
// SEO-003): `assemblePage` and `AssemblePageAction` are two independent head
// producers, and every per-page head fix so far (injectPageJSONLD 2026-07-28,
// injectCanonicalLink 2026-08-02, injectRobotsNoindex 2026-08-09) has landed on
// the first only. That question is architecture-scope by the owner ruling of
// 2026-07-29 and is deliberately NOT decided here — but the next head fix has a
// home now instead of a fourth private helper in the assembler.
package actions

import (
	"fmt"
	"regexp"
	"strings"

	"go.uber.org/zap"
)

// reOgPageIdentity matches the three og properties that are PAGE-scoped, in any
// attribute order and either quote style, taking the surrounding line
// whitespace so removal does not leave an indented blank line (same discipline
// as reEmptyMetaDescription).
//
// Deliberately NOT og:type / og:site_name / og:image / twitter:* — those are
// genuinely site-scoped and a per-site artefact is the right place for them.
// Their own defects (og:image emitted whether or not the file exists; og:title
// falling back to the bare domain) are bugs_open/322's, and that file carries a
// standing landmine against gating og:image on an assets row. Widening this
// pattern would silently take that decision.
var reOgPageIdentity = regexp.MustCompile(`(?i)[ \t]*<meta[^>]*property=["']og:(?:title|description|url)["'][^>]*>\n?`)

// reHeadLang reads the lang attribute off the <head> OPEN TAG. `\b` is
// load-bearing: without it the pattern also matches `<header lang="...">`,
// which is a chrome element inside the body and says nothing about the
// document. Either quote style, and the attribute may sit anywhere in the tag.
var reHeadLang = regexp.MustCompile(`(?i)<head\b[^>]*?\slang=["']([^"']+)["']`)

// spliceOpenGraph replaces whatever the stored head claims about THIS page's
// Open Graph identity with the page's own, and returns the head unchanged when
// there is nothing truthful to say.
//
// Remove-then-inject, for three reasons: it is idempotent by construction (a
// second pass strips exactly what the first injected); it repairs a head that
// already carries a wrong or duplicated value rather than adding a third; and
// it needs no chrome rebuild, which matters because a Go change is NOT an input
// to the chrome staleness fingerprint (datahelpers/chrome_render_inputs.go), so
// nothing would have regenerated those 24 stored heads on our account.
//
// Correct-or-absent throughout (the LNK-005 idiom, as spliceMetaDescription):
// a page with no meta description emits NO og:description rather than an empty
// one, because content="" is the page describing itself as nothing. Expect that
// to be common for a while — measured 2026-08-19, 55.7% of active pages had no
// meta_description (bugs_open/320, whose backfiller is now scheduled). An
// absent tag there is this function working, not failing.
func spliceOpenGraph(head string, page *PageInfo, logger *zap.Logger) string {
	skip := func(reason string) string {
		if logger != nil {
			logger.Debug("spliceOpenGraph: no per-page Open Graph emitted",
				zap.String("reason", reason))
		}
		return head
	}
	if head == "" || page == nil {
		return skip("empty head or nil page")
	}

	// Strip first, unconditionally: a stale claim must go even when this page
	// can fill none of the three. Leaving it would be the defect.
	head = reOgPageIdentity.ReplaceAllString(head, "")

	var b strings.Builder
	// Title is always available — getPageInfo COALESCEs pages.title to
	// pages.name — but a hand-built PageInfo may not carry one.
	if page.Title != "" {
		b.WriteString(fmt.Sprintf("<meta property=\"og:title\" content=\"%s\">\n", htmlEscapeAttr(page.Title)))
	}
	if page.MetaDesc != "" {
		b.WriteString(fmt.Sprintf("<meta property=\"og:description\" content=\"%s\">\n", htmlEscapeAttr(page.MetaDesc)))
	}
	// og:url's eligibility is injectCanonicalLink's, deliberately identical and
	// through the SAME helper, so og:url == the canonical href == the JSON-LD
	// @id by construction rather than by a comment asking three call sites to
	// agree. preferredPageURL owns the one normalisation (site root
	// /index.html -> /, bugs_open/251) and the pending www/HTTPS policy.
	switch {
	case page.Domain == "":
		_ = skip("page has no domain — og:url omitted")
	case !strings.HasPrefix(page.URL, "/"):
		_ = skip("page url is not root-relative — og:url omitted")
	case strings.ContainsAny(page.URL, "#?"):
		// A fragment or query is a view of another URL, not a page of its own.
		_ = skip("page url carries a fragment or query — og:url omitted")
	default:
		b.WriteString(fmt.Sprintf("<meta property=\"og:url\" content=\"%s\">\n", preferredPageURL(page.Domain, page.URL)))
	}

	if b.Len() == 0 {
		return head
	}
	if idx := strings.LastIndex(head, "</head>"); idx >= 0 {
		return head[:idx] + b.String() + head[idx:]
	}
	return head + b.String()
}

// headLangAttr returns the language the head component declares for the
// document, or "" when it declares none.
//
// This is the carrier the owner chose for bugs_open/252's locale half on
// 2026-08-11 (option 3): the language lives in the head COMPONENT — a per-site
// artefact, which is the right scope for it — and Go stops deciding it. The
// value reaches the template from site_specs (`site_config.locale.lang`)
// through the existing schema-driven config resolution, so opting a site in is
// a spec row and no Go runs on the value path.
func headLangAttr(head string) string {
	if m := reHeadLang.FindStringSubmatch(head); m != nil {
		return strings.TrimSpace(m[1])
	}
	return ""
}

// htmlDocumentOpen writes the doctype and opening <html> tag for an assembled
// page.
//
// The "en" default is not a placeholder to be tidied away later: it is what
// keeps every site that has NOT opted in byte-identical to the hardcoded line
// this replaced. A site changes its declared language by naming one, and until
// it does nothing about its output moves.
func htmlDocumentOpen(lang string) string {
	if lang == "" {
		lang = "en"
	}
	return fmt.Sprintf("<!DOCTYPE html>\n<html lang=%q>\n", lang)
}

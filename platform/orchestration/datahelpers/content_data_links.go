// FILE: platform/orchestration/datahelpers/content_data_links.go
//
// Internal-link resolution over a component's STORED content_data, as opposed
// to over its rendered markup (link_repair.go) — bugs_open/097.
//
// WHY THIS EXISTS. Every mechanism the platform has for "does this button point
// at a real page" answers the question by ENUMERATING field names:
//
//	ctaFieldNames (resolve_internal_links_action.go)  6 components x 2 named top-level fields
//	DeriveCTAURLFields (ctafields.go)                 top-level "<stem>_url" WITH a label sibling
//
// Both are blind to a url that lives anywhere else, and "anywhere else" is where
// most of them live. Measured against production 2026-08-02, stored content_data
// holds 655 internal page links, of which 51 resolve to no page on their own
// site — and every one of the 51 sits in a component NONE of the enumerations
// covers:
//
//	info-card-grid       36 unresolved   cards[].link_url        (nested in an array)
//	case-studies-grid    10 unresolved   card1_link_url ...      (top level, no label sibling)
//	tool-cta              4 unresolved   primary_cta_url         (component not in the map)
//	platform-comparison   1 unresolved   cta_url                 (component not in the map)
//
// The nesting is the point. `info-card-grid` declares `cards` as an array whose
// ITEMS carry link_url/link_label, so a walker over top-level schema fields
// cannot see it at any precedence; and `case-studies-grid` numbers its cards
// into 5 sibling `cardN_link_url` fields with no `cardN_link_label`, so the
// schema-derived pairing cannot pair them either. Adding either component to a
// list fixes that component; the next one reopens the gap. That is the shape
// 097 was filed about.
//
// THE RULE, AND WHY IT IS NOT ANOTHER LIST. A value is a candidate when its
// FIELD NAME says it holds a url (any depth, any container), and it is JUDGED
// only when ClassifyLinkScope says the VALUE is an internal page link. The value
// classifier is what keeps image and off-site fields out — an `image_url` holds
// "/images/x.jpg" (LinkScopeAsset) and a `docs_url` holds "https://..."
// (LinkScopeExternal), so neither is ever considered, without this file naming
// either of them. That is the same division of labour ctafields.go uses for its
// site_assets guard, moved from the schema to the data so nesting cannot hide it.
//
// WHAT IT DOES ABOUT WHAT IT FINDS — and the asymmetry is deliberate.
//
//   - REWRITE, when the target page exists and the writer merely omitted the
//     extension (/about -> /about.html). Strictly a correction: it can only turn
//     a link that goes nowhere into one that goes to a real stored pages.url,
//     and it emits a URL the database handed it rather than one assembled here
//     (bugs_closed/029 was a whole bug about an emitter that assembled plausible
//     URLs). 18 of the 51 measured above are this class.
//
//   - REPORT ONLY, when no page exists at all (33 of the 51). The content_data
//     analogue of link_repair.go's unlink arm is to BLANK the field, and that
//     arm is disputed in its own header ("WHAT IS STILL OWED: decide whether
//     unlink is the right repair action at all", raised by the editquality and
//     render_guardian seats). Widening a disputed repair from the rendered copy
//     to the SOURCE OF TRUTH — where it is unrecoverable — is not a scope fix's
//     to do. Under-repair is the fail-safe direction here for the same reason it
//     is there: the deployed artefact is already protected by the outbound
//     repair seams, so what is owed on this path is a report, not a second
//     deletion.
//
// Fail-open on an empty index, identically to RepairPageLinks: an empty index
// means "the page set could not be read", never "this site has no pages", and
// rewriting against one would be rewriting against nothing.
//
// THE KNOWN LIMIT, stated because it is a predicate used outside the input shape
// it was written for — which is a named failure pattern in this estate (016b §9,
// from bugs_open/093: "a shared predicate written for one input shape, reused on
// another"). ClassifyLinkScope's final arm is `default: LinkScopePage`, and that
// is right for an HREF ATTRIBUTE, where anything not external/mailto/anchor/asset
// really is a page link. Here it is fed a FIELD VALUE, so in principle a field
// named `link` holding prose ("Read more") would classify as a page link and be
// reported as a phantom.
//
// MEASURED rather than assumed, over every production content_data row on
// 2026-08-02: of 1,299 url-named field values fleet-wide, 655 classify as page
// scope and ALL 655 begin with "/" — the false-positive surface is currently
// EMPTY. The other 644 are removed by the classifier without this file naming any
// of them: 457 external, 168 asset, 16 empty, 2 anchor, 1 mailto.
//
// No shape guard is imposed on top, deliberately. Requiring a leading "/" would
// change nothing today and would trade a latent false POSITIVE for a latent false
// NEGATIVE — a genuinely relative dead link ("about.html") would go silent, and
// under-reporting a dead link is the worse of the two when the report is the whole
// point. If prose ever does appear in a finding, THAT is the evidence for adding
// the guard, and the discriminating check is the census in the workstream RUNBOOK
// (R1): re-run it and look at the page-scope values that do not begin with "/".

package datahelpers

import (
	"fmt"
	"sort"
	"strings"
)

// ContentDataLinkPhantom is the finding action for a link whose target has no
// pages row at all. LinkRepairRewrite (link_repair.go) is reused verbatim for
// the other arm so a reader of either record set sees one vocabulary.
const ContentDataLinkPhantom = "phantom"

// ContentDataLinkFinding is one internal page link found in stored content_data
// and what was done about it. Path is the dotted route to the field with array
// indices spelled out ("cards[2].link_url"), because "which of the six cards"
// is the first thing anyone reading the record needs and the field name alone
// cannot say it.
type ContentDataLinkFinding struct {
	Path    string `json:"path"`
	Href    string `json:"href"`
	NewHref string `json:"new_href,omitempty"` // rewrite only
	Action  string `json:"action"`             // LinkRepairRewrite | ContentDataLinkPhantom
}

// IsURLFieldName reports whether a content_data field name declares that its
// value is a URL. Deliberately a shape test on the NAME and not a list of known
// fields: it must match cardN_link_url, link_url, cta_url, primary_cta_url and
// a bare "url" inside an array item without any of them being enumerated here.
//
// It is intentionally generous — every live shape and any future one — because
// it only nominates a candidate. ClassifyLinkScope does the judging, so a
// generous name test costs nothing and a stingy one silently reopens the gap.
func IsURLFieldName(name string) bool {
	n := strings.ToLower(name)
	for _, base := range []string{"url", "urls", "href", "link"} {
		if n == base || strings.HasSuffix(n, "_"+base) {
			return true
		}
	}
	return false
}

// RepairContentDataLinks resolves every URL-named field in content_data against
// the site's real page index, rewriting the extension-only misses in place and
// reporting the rest. Returns every finding, in a deterministic path order.
//
// data is mutated ONLY on the rewrite arm, and only by replacing one string
// with a stored pages.url. A content_data map with no findings is left byte-
// identical, so a clean component is never perturbed.
func RepairContentDataLinks(data map[string]interface{}, index PageURLIndex) []ContentDataLinkFinding {
	if len(data) == 0 || len(index) == 0 {
		return nil
	}
	var findings []ContentDataLinkFinding
	walkContentDataLinks(data, "", index, &findings)
	sort.SliceStable(findings, func(i, j int) bool { return findings[i].Path < findings[j].Path })
	return findings
}

// walkContentDataLinks descends maps and slices alike. It carries the parent
// container so the rewrite arm can assign back into it — a walk that only
// yielded values could report but never correct, and the two must see exactly
// the same set or the record would describe links the repair did not touch.
func walkContentDataLinks(node interface{}, path string, index PageURLIndex, findings *[]ContentDataLinkFinding) {
	switch v := node.(type) {
	case map[string]interface{}:
		// Traversal order is deliberately NOT pinned here. Go randomises map
		// iteration, and sorting these keys would ALSO make the output stable —
		// which is the trap: two mechanisms guaranteeing one property means
		// deleting either leaves every test green, so neither is ever shown to
		// be load-bearing. The single guarantee is the path sort in
		// RepairContentDataLinks, and TestFindingOrderIsStable fails within a
		// few runs if it is removed.
		for k := range v {
			child := joinContentPath(path, k)
			if s, ok := v[k].(string); ok {
				if !IsURLFieldName(k) {
					continue
				}
				if newHref, f := judgeContentDataLink(child, s, index); f != nil {
					if newHref != "" {
						v[k] = newHref
					}
					*findings = append(*findings, *f)
				}
				continue
			}
			walkContentDataLinks(v[k], child, index, findings)
		}
	case []interface{}:
		// An array element inherits its parent's field name: `cards` is the
		// name that says nothing, and `cards[2].link_url` is where the url is.
		// A bare string array (video_urls) keeps the parent name, which is why
		// the index is appended rather than replacing the path.
		for i := range v {
			child := fmt.Sprintf("%s[%d]", path, i)
			if s, ok := v[i].(string); ok {
				if !IsURLFieldName(lastPathField(path)) {
					continue
				}
				if newHref, f := judgeContentDataLink(child, s, index); f != nil {
					if newHref != "" {
						v[i] = newHref
					}
					*findings = append(*findings, *f)
				}
				continue
			}
			walkContentDataLinks(v[i], child, index, findings)
		}
	}
}

// judgeContentDataLink applies the two arms to one value. It returns the
// replacement string for the rewrite arm ("" when nothing should be written)
// and the finding, or a nil finding when the value is not an internal page link
// or already resolves.
func judgeContentDataLink(path, href string, index PageURLIndex) (string, *ContentDataLinkFinding) {
	// LinkScopeEmpty is deliberately NOT a finding here, and this is the one
	// place the content_data judgement differs from the markup one. An empty
	// href in rendered HTML is a control that goes nowhere; an empty url FIELD
	// is the correct resting state of an unset optional, because every CTA
	// template is gated ({{if .link_url}}) and renders no anchor at all. Flagging
	// it would fire on every card that legitimately has no link.
	if ClassifyLinkScope(href) != LinkScopePage {
		return "", nil
	}
	if index.Contains(href) {
		return "", nil
	}
	if stored, ok := index.Lookup(NormalizePagePath(href) + ".html"); ok {
		newHref := stored + hrefSuffix(href)
		return newHref, &ContentDataLinkFinding{
			Path: path, Href: href, NewHref: newHref, Action: LinkRepairRewrite,
		}
	}
	return "", &ContentDataLinkFinding{Path: path, Href: href, Action: ContentDataLinkPhantom}
}

// CountContentDataLinkFindings splits findings by arm, for the caller's log line
// and durable record.
func CountContentDataLinkFindings(findings []ContentDataLinkFinding) (rewritten, phantom int) {
	for _, f := range findings {
		switch f.Action {
		case LinkRepairRewrite:
			rewritten++
		case ContentDataLinkPhantom:
			phantom++
		}
	}
	return rewritten, phantom
}

func joinContentPath(parent, key string) string {
	if parent == "" {
		return key
	}
	return parent + "." + key
}

// lastPathField returns the field name a path ends in, ignoring array indices:
// "cards[2]" -> "cards", "a.b[0][1]" -> "b". Used so a string sitting directly
// in an array is judged by the array's own field name.
func lastPathField(path string) string {
	if i := strings.IndexByte(path, '['); i >= 0 {
		path = path[:i]
	}
	if i := strings.LastIndexByte(path, '.'); i >= 0 {
		path = path[i+1:]
	}
	return path
}

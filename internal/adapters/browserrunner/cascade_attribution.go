// FILE: internal/adapters/browserrunner/cascade_attribution.go
//
// Cascade attribution: for a text element the render audit has already measured
// as failing, find the CSS declaration that actually decides its colour, and
// prove it.
//
// WHY THIS EXISTS (bugs_open/390). A contrast_failure is repaired by
// css-patch-agent, which APPENDS a rule to the site theme stylesheet. The audit
// tells it which element is wrong; nothing tells it what its rule has to BEAT.
// So the platform aims a repair at a file that usually cannot govern the pixel:
// [MEASURED 2026-08-25] of 40 completed repairs sampled across 7 sites, the
// winning declaration was in the page's own <style> block in 39 and in the theme
// in 0, and it out-specified the audited selector in 33. The repair is authored,
// deployed, marked complete, and inert.
//
// WHAT IT RECORDS. Per failing element: which declaration wins, WHERE it lives
// (a linked stylesheet, a page <style> block, or the element's own style
// attribute), whether it carries !important, and whether the site theme is
// linked AFTER it in document order. That is everything a router needs to decide
// whether the theme can reach the defect at all.
//
// ── THE LOAD-BEARING IDEA: REMOVAL IS THE PROOF ──────────────────────────────
//
// Walking document.styleSheets cannot see everything. Cascade layers reorder
// precedence, a cross-origin sheet throws on .cssRules, @supports and @media
// conditions have to be re-evaluated, and a selector list has to be split and
// tested part by part. Any of those can make a walker confidently name the wrong
// declaration — and a confident wrong answer here is worse than none, because a
// router would act on it.
//
// So the walker's ordering is NOT the claim. The claim is settled by removing
// the declaration and re-reading the computed value:
//
//	REMOVING A LOSING DECLARATION CANNOT CHANGE THE COMPUTED VALUE.
//	So if the value changes, the removed declaration WAS the winner.
//
// That is sound whatever heuristic chose the candidate, which is why the
// specificity guess below is allowed to be rough: it only decides what to TRY
// FIRST. Candidates are tried best-first and the first whose removal moves the
// value is the answer. If none moves it, we report verified:false and the
// consumer treats the finding as unattributed — the safe direction.
//
// It is also conservative in the one case it cannot separate: if the winner and
// the runner-up specify the SAME value, removing the winner changes nothing, and
// we report unverified rather than guessing. Under-claiming, by construction.
//
// Every removal is restored immediately and re-checked. If a restore does not
// take, the page is marked dirty and attribution stops for that page — a
// subsequent measurement on a mutated page would be measuring our own edit. The
// page instance is discarded after auditOne either way, so nothing we do here
// can reach a user.
//
// DELIBERATELY PROBE-LOCAL, and not part of contrastMathsJS. That kernel is
// byte-identical to the string the render audit and the contrast_ratio check
// already run in production; this must not change either one's behaviour. Same
// reasoning as backdropBound in contrast_check.go.
//
// SPECIFICITY IS NOT DECIDED HERE. This file reports the winning rule's SELECTOR
// TEXT. The authoritative specificity is computed on the chassis side with
// cascadia (contrast_cascade_route.go), which is a real CSS3 implementation and
// is unit-testable. One authority, not two that can disagree.

package browserrunner

// cascadeSchemeV1 names the attribution contract this adapter speaks. Stamped
// UNCONDITIONALLY on every reply, including one that found nothing, for the same
// reason as selectorSchemeVerifiedV1: a consumer must be able to tell "this
// adapter attributes declarations" from "this adapter is too old to", and a
// clean page would otherwise be indistinguishable from an un-rolled one. Absent
// means an old-shape reply, which is what keeps the consumer's fallback inert
// rather than wrong.
const cascadeSchemeV1 = "cascade/v1"

// Surfaces a winning declaration can live on. The distinction the router cares
// about is not cosmetic: only surfaceLinked is a file css-patch-agent can edit,
// and only surfaceInlineAttr with !important is unreachable by any stylesheet
// rule at all.
const (
	surfaceLinked     = "linked"      // a <link>ed stylesheet
	surfaceStyleBlock = "style_block" // a <style> element in the document
	surfaceInlineAttr = "inline_attr" // the element's own style="" attribute
	surfaceUADefault  = "ua_default"  // nothing in the document sets it
	surfaceOpaque     = "opaque"      // a stylesheet we are not allowed to read
)

// CascadeWinner is the declaration that decides one property on one element.
//
// Verified is the field to read FIRST. When it is false every other field is a
// guess and must not be routed on — see the removal-is-the-proof note above.
type CascadeWinner struct {
	// Property is the CSS property attributed, e.g. "color".
	Property string `json:"property"`
	// Surface is where the winning declaration lives; one of the surface
	// constants above.
	Surface string `json:"surface"`
	// Selector is the winning rule's own selector text - the ONE complex
	// selector part that matched, not the whole comma-separated list it may
	// have been written in. Empty for surfaceInlineAttr (a style attribute has
	// no selector) and for surfaceUADefault.
	Selector string `json:"selector,omitempty"`
	// SheetHref is the stylesheet's URL for surfaceLinked, empty otherwise.
	SheetHref string `json:"sheet_href,omitempty"`
	// Important is whether the winning declaration carries !important. An
	// author !important beats every non-important author declaration whatever
	// its specificity or position, so this decides whether an appended rule
	// must match it.
	Important bool `json:"important,omitempty"`
	// ThemeAfterWinner records whether the site theme's <link> appears AFTER
	// the winner's owner node in document order. It turns "source order" from
	// an assumption into a recorded fact: when false, an appended theme rule
	// must out-specify the winner STRICTLY, because a tie loses on position.
	ThemeAfterWinner bool `json:"theme_after_winner,omitempty"`
	// Decl is the winning declaration's specified value verbatim, e.g.
	// "var(--color-primary, #1a1a2e)". Recorded for humans reading a parked
	// row; nothing routes on it.
	Decl string `json:"decl,omitempty"`
	// VarName is the custom property named by Decl when it is a var()
	// reference. INFORMATIONAL ONLY - it is a parse of the declaration text,
	// not an attribution of where that property is defined, which is the same
	// cascade problem one level down. Never route on it.
	VarName string `json:"var_name,omitempty"`
	// Verified is true only when REMOVING this declaration changed the
	// element's computed value for Property. False means the walker could not
	// prove its answer and the finding must be treated as unattributed.
	Verified bool `json:"verified"`
	// OpaqueSheets counts stylesheets whose rules could not be read at all
	// (cross-origin). Non-zero with Verified false is "we were blinded", which
	// is a different thing from "nothing sets this property" and must not be
	// collapsed into it.
	OpaqueSheets int `json:"opaque_sheets,omitempty"`
	// Candidates is how many declarations for Property were found to match
	// this element before the winner was proven. Reported so a reader can tell
	// a one-rule page from a heavily-overridden one.
	Candidates int `json:"candidates,omitempty"`
}

// cascadeAttributionJS defines winningDecl(el, prop, themeLink) in the page.
//
// It returns an object with the same field names as CascadeWinner's JSON tags,
// or null when the element has no declaration for the property anywhere (the
// caller records surfaceUADefault).
//
// Read the file header before changing any of this: the ordering heuristic is
// deliberately allowed to be rough, and the removal test is what makes the
// result trustworthy.
const cascadeAttributionJS = `
  // Rough specificity, used ONLY to decide what to try first. The authoritative
  // computation is cascadia's, on the chassis side. Deliberately not clever:
  // being wrong here costs one extra removal attempt, never a wrong answer.
  var cascRoughSpec = function(sel){
    var s = String(sel);
    var a = (s.match(/#[A-Za-z0-9_-]+/g) || []).length;
    var b = (s.match(/\.[A-Za-z0-9_-]+/g) || []).length
          + (s.match(/\[[^\]]*\]/g) || []).length
          + (s.match(/:[A-Za-z-]+(\([^)]*\))?/g) || []).length;
    var c = (s.match(/(^|[\s>+~])[A-Za-z][A-Za-z0-9-]*/g) || []).length;
    return a * 10000 + b * 100 + c;
  };
  var cascVarName = function(v){
    var m = String(v).match(/var\(\s*(--[A-Za-z0-9_-]+)/);
    return m ? m[1] : '';
  };
  // Which <style>/<link> node owns a rule, and therefore which surface it is on.
  var cascSurfaceOf = function(sheet){
    if (!sheet) { return {surface:'style_block', href:'', node:null}; }
    var node = sheet.ownerNode || null;
    if (node && node.tagName === 'LINK') {
      return {surface:'linked', href: sheet.href || node.getAttribute('href') || '', node:node};
    }
    return {surface:'style_block', href:'', node:node};
  };
  var winningDecl = function(el, prop, themeLink){
    var cands = [], opaque = 0, order = 0;
    // The element's own style attribute is a candidate with no selector.
    try {
      if (el.style && el.style.getPropertyValue(prop) !== '') {
        cands.push({surface:'inline_attr', sel:'', href:'', node:null, rule:null,
                    important: el.style.getPropertyPriority(prop) === 'important',
                    decl: el.style.getPropertyValue(prop), rank: 1000000, order: 1e9});
      }
    } catch (e) {}
    var walk = function(rules, sheet){
      if (!rules) { return; }
      for (var i = 0; i < rules.length; i++) {
        var r = rules[i];
        // CSSMediaRule / CSSSupportsRule / CSSLayerBlockRule etc. Recurse only
        // into conditions that actually apply right now; a rule inside a media
        // query that does not match is not in the cascade at all.
        if (r.media && r.cssRules) {
          try { if (!window.matchMedia(r.conditionText || r.media.mediaText).matches) { continue; } } catch (e) {}
          walk(r.cssRules, sheet); continue;
        }
        if (r.conditionText && r.cssRules && typeof CSS !== 'undefined' && CSS.supports) {
          try { if (!CSS.supports(r.conditionText)) { continue; } } catch (e) {}
          walk(r.cssRules, sheet); continue;
        }
        if (r.styleSheet) { // @import
          try { walk(r.styleSheet.cssRules, r.styleSheet); } catch (e) { opaque++; }
          continue;
        }
        if (r.cssRules && !r.selectorText) { walk(r.cssRules, sheet); continue; }
        if (!r.selectorText || !r.style) { continue; }
        var val = '';
        try { val = r.style.getPropertyValue(prop); } catch (e) { continue; }
        if (val === '') { continue; }
        order++;
        // A selector list is several selectors. Record the PART that matched,
        // because that is the one whose specificity is in play - reporting the
        // whole list would overstate what has to be beaten.
        var parts = String(r.selectorText).split(',');
        var best = null;
        for (var p = 0; p < parts.length; p++) {
          var part = parts[p].trim();
          if (!part) { continue; }
          var ok = false;
          try { ok = el.matches(part); } catch (e) { ok = false; }
          if (!ok) { continue; }
          var rank = cascRoughSpec(part);
          if (best === null || rank > best.rank) { best = {part: part, rank: rank}; }
        }
        if (best === null) { continue; }
        var surf = cascSurfaceOf(sheet);
        var imp = false;
        try { imp = r.style.getPropertyPriority(prop) === 'important'; } catch (e) {}
        cands.push({surface: surf.surface, sel: best.part, href: surf.href, node: surf.node,
                    rule: r, important: imp, decl: val, rank: best.rank, order: order});
      }
    };
    for (var s = 0; s < document.styleSheets.length; s++) {
      var sheet = document.styleSheets[s];
      var rules = null;
      // A cross-origin sheet throws here. Count it and carry on: being blinded
      // is a fact to report, not a reason to abandon the page.
      try { rules = sheet.cssRules; } catch (e) { opaque++; continue; }
      walk(rules, sheet);
    }
    if (cands.length === 0) { return null; }
    // Best-first: !important beats everything, then rough specificity, then
    // later source order. Only an ORDERING - the removal test below is the proof.
    cands.sort(function(a, b){
      if (a.important !== b.important) { return a.important ? -1 : 1; }
      if (a.rank !== b.rank) { return b.rank - a.rank; }
      return b.order - a.order;
    });
    var before = getComputedStyle(el).getPropertyValue(prop);
    var result = null;
    for (var c = 0; c < cands.length && result === null; c++) {
      var cand = cands[c], style = cand.rule ? cand.rule.style : el.style, old = '', pri = '';
      try {
        old = style.getPropertyValue(prop);
        pri = style.getPropertyPriority(prop);
        style.removeProperty(prop);
      } catch (e) { continue; }
      var after = getComputedStyle(el).getPropertyValue(prop);
      var moved = (after !== before);
      // Restore immediately, and CHECK the restore. A page we have edited and
      // failed to put back is a page nothing else may be measured on.
      var restored = false;
      try {
        style.setProperty(prop, old, pri);
        restored = (getComputedStyle(el).getPropertyValue(prop) === before);
      } catch (e) { restored = false; }
      if (!restored) { window.__cascadeDirty = true; return null; }
      if (moved) {
        var themeAfter = false;
        try {
          if (themeLink && cand.node) {
            themeAfter = (cand.node.compareDocumentPosition(themeLink)
                          & Node.DOCUMENT_POSITION_FOLLOWING) !== 0;
          }
        } catch (e) {}
        result = {property: prop, surface: cand.surface, selector: cand.sel,
                  sheet_href: cand.href, important: !!cand.important,
                  theme_after_winner: themeAfter, decl: cand.decl,
                  var_name: cascVarName(cand.decl), verified: true,
                  opaque_sheets: opaque, candidates: cands.length};
      }
    }
    if (result === null) {
      // Nothing we could remove moved the value. Either the winner is somewhere
      // we cannot see (an opaque sheet, a layer we mis-ordered) or the runner-up
      // specifies the same value. Report the best guess UNVERIFIED so the
      // consumer discards it, rather than a confident wrong answer.
      var g = cands[0], gs = cascSurfaceOf(g.rule ? g.rule.parentStyleSheet : null);
      result = {property: prop, surface: g.surface, selector: g.sel,
                sheet_href: g.href, important: !!g.important,
                theme_after_winner: false, decl: g.decl,
                var_name: cascVarName(g.decl), verified: false,
                opaque_sheets: opaque, candidates: cands.length};
    }
    return result;
  };
  // The site theme's own <link>, used for the document-order fact. Matched on
  // the deployed path rather than on order, because a page may link several.
  var cascThemeLink = (function(){
    var ls = document.querySelectorAll('link[rel~="stylesheet"]');
    for (var i = 0; i < ls.length; i++) {
      var h = ls[i].getAttribute('href') || '';
      if (h.indexOf('/assets/css/styles.css') !== -1) { return ls[i]; }
    }
    return null;
  })();`

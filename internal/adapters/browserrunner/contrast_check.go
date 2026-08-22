// FILE: internal/adapters/browserrunner/contrast_check.go
//
// The Tier-4 `contrast_ratio` check (bugs_open/131's framework half — the
// vonc gauntlet audit BY SLUG; proposed in
// experience_loop/HANDOFF_2026-07-28_appeal_dimension.md and unbuilt until
// 2026-08-22). Measures what is PAINTED: every visible text element's computed
// colour against its effective composited background, per WCAG AA. It exists
// because the other checks in this ladder assert presence, behaviour and
// layout, and none can see colour — four sites shipped unreadable text that
// only a person caught, and a point fix to a colour pair decayed within three
// weeks of palette churn with nothing watching.
//
// Split of responsibilities, deliberately: the JS only MEASURES (it returns
// every element below its threshold, marked approximate or firm, in or out of
// the tool container); the VERDICT lives in Go so it is testable against the
// fakePage seam without a browser. Text over an image/gradient backdrop can
// never fail the check — the composited ground there is a mid-grey stand-in,
// the ratio is approximate, and a false acceptance failure becomes an
// improve_tool fixer aimed at a correct page (bugs_open/126). The render
// audit's ContrastFirm applies the same rule for the same reason.

package browserrunner

import (
	"encoding/json"
	"fmt"
	"sort"
)

// contrastMathsJS is the WCAG measurement kernel shared by the render audit's
// auditJS (render_audit_action.go) and the contrast_ratio probe below: sRGB
// parsing, relative luminance, contrast ratio, alpha compositing, and the
// effective-background walk (ancestors composited until an opaque one; a
// background image or gradient pushes a mid-grey stand-in and marks the
// result approximate). ONE copy on purpose — two implementations of this
// maths already exist in this repo (here and scripts/render_audit.py), and a
// third would drift.
const contrastMathsJS = `
  function parseRGB(s){var m=String(s).match(/rgba?\(([^)]+)\)/);if(!m)return null;
    var p=m[1].split(',').map(function(x){return parseFloat(x.trim())});
    return {r:p[0],g:p[1],b:p[2],a:p.length>3?p[3]:1};}
  function lum(c){function f(v){v=v/255;return v<=0.03928?v/12.92:Math.pow((v+0.055)/1.055,2.4)}
    return 0.2126*f(c.r)+0.7152*f(c.g)+0.0722*f(c.b);}
  function ratio(a,b){var l1=lum(a),l2=lum(b);
    return (Math.max(l1,l2)+0.05)/(Math.min(l1,l2)+0.05);}
  function over(fg,bg){if(fg.a>=1)return fg;
    return {r:fg.r*fg.a+bg.r*(1-fg.a),g:fg.g*fg.a+bg.g*(1-fg.a),b:fg.b*fg.a+bg.b*(1-fg.a),a:1};}
  function effBG(el){
    var stack=[],node=el,anyImg=false;
    while(node&&node.nodeType===1){
      var cs=getComputedStyle(node),c=parseRGB(cs.backgroundColor);
      var hasImg=cs.backgroundImage&&cs.backgroundImage!=='none';
      if(c&&c.a>0)stack.push({c:c,img:hasImg});
      if(hasImg&&(!c||c.a<1))stack.push({c:{r:128,g:128,b:128,a:1},img:true});
      if(c&&c.a>=1)break;
      node=node.parentElement;
    }
    var base={r:255,g:255,b:255,a:1};
    for(var i=stack.length-1;i>=0;i--){if(stack[i].img)anyImg=true;base=over(stack[i].c,base);}
    return {bg:base,overImage:anyImg};
  }`

// contrastProbe builds the in-page scan for one check run. The container
// selector and threshold are embedded as literals because the browserPage
// seam evaluates a bare expression (no argument channel). Element filters are
// the audit's, verbatim: hidden, zero-sized and near-empty elements are not
// text a visitor reads. minRatio == 0 means the per-element WCAG AA default
// (4.5:1 body, 3.0:1 large: >=24px, or >=18.66px bold); a fence that sets
// min_ratio replaces BOTH — a visible, deliberate design exception.
func contrastProbe(container string, minRatio float64) string {
	sel, _ := json.Marshal(container)
	return `() => {` + contrastMathsJS + fmt.Sprintf(`
  var minRatio = %g;
  var containerSel = %s;
  var tool = null;
  try { tool = containerSel ? document.querySelector(containerSel) : null; } catch (e) { tool = null; }
  var describe = function(el){ return el.tagName.toLowerCase()
    + (el.id ? '#' + el.id : '')
    + (el.className && typeof el.className === 'string' && el.className.trim()
       ? '.' + el.className.trim().split(/\s+/).join('.') : ''); };
  var out = { located: !!tool, failures: [] }, seen = {};
  var all = document.querySelectorAll('body *');
  for (var i = 0; i < all.length; i++) {
    var el = all[i], cs = getComputedStyle(el);
    if (cs.display === 'none' || cs.visibility === 'hidden' || parseFloat(cs.opacity) === 0) continue;
    var r0 = el.getBoundingClientRect();
    if (r0.width === 0 || r0.height === 0) continue;
    var txt = '';
    for (var n = 0; n < el.childNodes.length; n++)
      if (el.childNodes[n].nodeType === 3) txt += el.childNodes[n].nodeValue;
    txt = txt.replace(/\s+/g, ' ').trim();
    if (txt.length < 2) continue;
    var fg = parseRGB(cs.color); if (!fg) continue;
    var eb = effBG(el), fgc = over(fg, eb.bg), rr = ratio(fgc, eb.bg);
    var size = parseFloat(cs.fontSize), weight = parseInt(cs.fontWeight, 10) || 400;
    var large = size >= 24 || (size >= 18.66 && weight >= 700);
    var need = minRatio > 0 ? minRatio : (large ? 3.0 : 4.5);
    if (rr >= need) continue;
    var key = describe(el) + '|' + cs.color + '|' + txt.slice(0, 40);
    if (seen[key]) continue; seen[key] = 1;
    out.failures.push({ selector: describe(el), text: txt.slice(0, 60), fg: cs.color,
      bg: 'rgb(' + Math.round(eb.bg.r) + ',' + Math.round(eb.bg.g) + ',' + Math.round(eb.bg.b) + ')',
      ratio: Math.round(rr * 100) / 100, need: need, overImage: eb.overImage,
      px: Math.round(size), inTool: !!(tool && tool.contains(el)) });
  }
  return out;
}`, minRatio, string(sel))
}

// contrastHit is one element the probe found below its threshold.
type contrastHit struct {
	Selector  string  `json:"selector"`
	Text      string  `json:"text"`
	FG        string  `json:"fg"`
	BG        string  `json:"bg"`
	Ratio     float64 `json:"ratio"`
	Need      float64 `json:"need"`
	OverImage bool    `json:"overImage"`
	Px        float64 `json:"px"`
	InTool    bool    `json:"inTool"`
}

type contrastScan struct {
	Located  bool          `json:"located"`
	Failures []contrastHit `json:"failures"`
}

// decodeContrastScan round-trips the Evaluate result through JSON into the
// typed shape. A field arriving as the wrong type is an error, not a zero —
// a measurement whose payload shape changed is something to hear about
// (the evalNumber rule, applied wholesale).
func decodeContrastScan(v interface{}) (contrastScan, error) {
	var scan contrastScan
	raw, err := json.Marshal(v)
	if err != nil {
		return scan, fmt.Errorf("contrast probe result not encodable: %w", err)
	}
	if err := json.Unmarshal(raw, &scan); err != nil {
		return scan, fmt.Errorf("contrast probe result has unexpected shape: %w", err)
	}
	return scan, nil
}

// runContrastRatio evaluates the probe and judges it. Fails on any FIRM
// sub-threshold element wherever it sits — a visitor cannot read chrome text
// either — and attributes the worst offender in/out of the tool container so
// the judge routes a tool ticket or a chrome one (the no_horizontal_overflow
// pattern, reused rather than re-invented). Approximate (over-image) readings
// are reported but can never fail.
func runContrastRatio(page browserPage, doc criteriaDoc, ch criteriaCheck, profile string) (bool, string, CheckResult) {
	v, err := page.Evaluate(contrastProbe(toolContainer(doc, ch), ch.MinRatio))
	if err != nil {
		return false, "could not measure contrast: " + err.Error(), CheckResult{}
	}
	scan, err := decodeContrastScan(v)
	if err != nil {
		return false, "could not measure contrast: " + err.Error(), CheckResult{}
	}

	var firm []contrastHit
	approx := 0
	for _, f := range scan.Failures {
		if f.OverImage {
			approx++
			continue
		}
		firm = append(firm, f)
	}

	if len(firm) == 0 {
		detail := "all text meets its contrast threshold on " + profile
		if approx > 0 {
			detail += fmt.Sprintf(" (%d element(s) over an image or gradient backdrop not judged — the composited ground there is approximate and never fails this check)", approx)
		}
		return true, detail, CheckResult{}
	}

	// Worst ratio first; the deepest problem is the one the fixer should see.
	sort.Slice(firm, func(i, j int) bool { return firm[i].Ratio < firm[j].Ratio })
	worst := firm[0]

	detail := fmt.Sprintf(
		"text is painted below its contrast threshold on %s: %d firm failure(s), worst %s %q at %.2f:1 (needs %.1f:1, %s on %s, %.0fpx)",
		profile, len(firm), worst.Selector, worst.Text, worst.Ratio, worst.Need, worst.FG, worst.BG, worst.Px)
	for i, f := range firm {
		if i == 0 {
			continue
		}
		if i >= 5 {
			detail += fmt.Sprintf("; and %d more", len(firm)-5)
			break
		}
		detail += fmt.Sprintf("; %s at %.2f:1", f.Selector, f.Ratio)
	}
	if approx > 0 {
		detail += fmt.Sprintf(" (%d further element(s) over an image/gradient backdrop not judged)", approx)
	}

	scope := ScopeUnknown
	switch {
	case !scan.Located:
		detail += " (tool container not found — attribution unknown)"
	case worst.InTool:
		scope = ScopeTool
		detail += " — inside the tool"
	default:
		scope = ScopeChrome
		detail += " — OUTSIDE the tool container: site chrome"
	}

	return false, detail, CheckResult{
		Scope:           scope,
		Culprit:         fmt.Sprintf("%s (%.2f:1, needs %.1f:1)", worst.Selector, worst.Ratio, worst.Need),
		CulpritSelector: worst.Selector,
	}
}

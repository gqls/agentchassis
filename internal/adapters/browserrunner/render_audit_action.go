// FILE: internal/adapters/browserrunner/render_audit_action.go
//
// Renders a live page and measures what a visitor actually sees: text contrast
// against the effective background, images that failed to load, and horizontal
// overflow.
//
// A GO PORT of `scripts/render_audit.py` (brochure workstream, 2026-07-27), which
// this replaces. The Python found 101 AA failures across 5 pages of
// fundamentallyai.com on a day when every page said `deployed` and none of ~50
// discovery checks objected, and 65 more across the fleet on 2026-07-28 — but it
// runs only when a human types it, which is how an unreadable chart reached
// oufe.com and was found by the owner in a screenshot. Porting it here puts the
// measurement in the one pod that already has a browser, so it can be dispatched.
//
// WHY THIS CANNOT BE A DISCOVERY CHECK. Every check in
// `platform/orchestration/actions/discovery_checks/` runs inside the chassis,
// which has **no browser** (verified 2026-07-28: no playwright cache, no chromium
// binary). `check_palette_contrast` is the closest relative and is deliberately
// DB-only — it reads the *composed palette* in microseconds and its own header
// states the family it cannot see:
//
//	"a component that hard-codes an ink over a themed fill ... is invisible to
//	 any palette-level check by construction"
//
// That family is exactly what put a 1.00:1 link on a live homepage. Catching it
// requires rendering, so it lives here, beside `run_checks`, and is reached the
// same way — a Kafka action on the adapter that owns Chromium.
//
// THREE MEASUREMENT RULES, each carried over deliberately from the Python:
//
//  1. The EFFECTIVE background, composited. Walk up from the element through
//     transparent ancestors, push each layer, stop at the first opaque one, then
//     composite back down. Comparing text to the *page* background instead flags
//     every header that sits on its own bar — a false positive that, "fixed",
//     makes a working site worse.
//
//  2. An unknown backdrop is REPORTED, not silently scored. A background image
//     under the text has no knowable colour, so a mid grey is assumed and the
//     finding is marked `over_image`. That keeps the number honest about itself:
//     on oufe.com the only surviving finding was a white button over a near-black
//     hero, flagged this way, and a screenshot confirmed it was fine.
//
//  3. Only DIRECT text nodes count. Measuring an element's full `textContent`
//     attributes a child's colour to its parent and reports the same failure at
//     every level of the tree.
package browserrunner

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// RenderAuditRequest asks for one or more live URLs to be measured.
type RenderAuditRequest struct {
	RunID  string   `json:"run_id"`
	URLs   []string `json:"urls"`
	SiteID string   `json:"site_id"`
	Domain string   `json:"domain"`
}

// ContrastFinding is one text element whose contrast is below its threshold.
type ContrastFinding struct {
	URL       string  `json:"url"`
	Tag       string  `json:"tag"`
	Class     string  `json:"class"`
	Text      string  `json:"text"`
	FG        string  `json:"fg"`
	BG        string  `json:"bg"`
	Ratio     float64 `json:"ratio"`
	Need      float64 `json:"need"`
	FontPx    int     `json:"font_px"`
	OverImage bool    `json:"over_image"` // backdrop unknown; ratio is approximate
}

// BrokenImage is an <img> the browser could not load.
type BrokenImage struct {
	URL string `json:"url"`
	Src string `json:"src"`
	Alt string `json:"alt"`
}

// OverflowFinding is a document wider than its own viewport.
type OverflowFinding struct {
	URL         string `json:"url"`
	ScrollWidth int    `json:"scroll_width"`
	Viewport    int    `json:"viewport"`
}

// RenderAuditResult is the whole verdict. PagesFailed counts pages with a
// *non-approximate* contrast failure, so an over-image reading never on its own
// turns a run red.
type RenderAuditResult struct {
	RunID    string            `json:"run_id"`
	Contrast []ContrastFinding `json:"contrast"`
	Images   []BrokenImage     `json:"broken_images"`
	Overflow []OverflowFinding `json:"overflow"`
	// Unreachable pages are reported rather than skipped: a page that will not
	// load is a worse finding than one with poor contrast, and silently dropping
	// it would let a dead page pass as clean.
	Unreachable []string `json:"unreachable,omitempty"`
	Summary     struct {
		Pages         int `json:"pages"`
		PagesFailed   int `json:"pages_failed"`
		Contrast      int `json:"contrast"`
		ContrastFirm  int `json:"contrast_firm"` // excludes over_image approximations
		BrokenImages  int `json:"broken_images"`
		OverflowPages int `json:"overflow_pages"`
	} `json:"summary"`
}

// auditJS is the in-page probe. Kept as one expression returning a value so the
// caller gets it via Evaluate — the Python had to inject into a local copy and
// read a <pre> back, because headless Chrome's CLI has no evaluate-on-load hook.
// Driving Chromium properly removes that whole workaround.
const auditJS = `() => {
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
  }
  var out={contrast:[],images:[],overflow:null},seen={};
  var all=document.querySelectorAll('body *');
  for(var i=0;i<all.length;i++){
    var el=all[i],cs=getComputedStyle(el);
    if(cs.display==='none'||cs.visibility==='hidden'||parseFloat(cs.opacity)===0)continue;
    var r0=el.getBoundingClientRect();
    if(r0.width===0||r0.height===0)continue;
    var txt='';
    for(var n=0;n<el.childNodes.length;n++)
      if(el.childNodes[n].nodeType===3)txt+=el.childNodes[n].nodeValue;
    txt=txt.replace(/\s+/g,' ').trim();
    if(txt.length<2)continue;
    var fg=parseRGB(cs.color);if(!fg)continue;
    var eb=effBG(el),fgc=over(fg,eb.bg),r=ratio(fgc,eb.bg);
    var size=parseFloat(cs.fontSize),weight=parseInt(cs.fontWeight,10)||400;
    var large=size>=24||(size>=18.66&&weight>=700),need=large?3.0:4.5;
    if(r>=need)continue;
    var cls=(typeof el.className==='string'?el.className:'')||el.tagName;
    var key=cls+'|'+cs.color+'|'+txt.slice(0,40);
    if(seen[key])continue; seen[key]=1;
    out.contrast.push({cls:cls,tag:el.tagName,text:txt.slice(0,80),fg:cs.color,
      bg:'rgb('+Math.round(eb.bg.r)+','+Math.round(eb.bg.g)+','+Math.round(eb.bg.b)+')',
      ratio:Math.round(r*100)/100,need:need,overImage:eb.overImage,px:Math.round(size)});
  }
  for(var j=0;j<document.images.length;j++){
    var im=document.images[j];
    if(!im.complete||im.naturalWidth===0)
      out.images.push({src:im.getAttribute('src'),alt:(im.alt||'').slice(0,80)});
  }
  if(document.documentElement.scrollWidth>window.innerWidth+1)
    out.overflow={scrollWidth:document.documentElement.scrollWidth,viewport:window.innerWidth};
  return out;
}`

// RenderAuditAction measures live pages in Chromium.
type RenderAuditAction struct {
	logger *zap.Logger
	open   openFunc
}

func NewRenderAuditAction(logger *zap.Logger) *RenderAuditAction {
	return &RenderAuditAction{logger: logger.Named("render_audit"), open: openChromium}
}

// pageAudit mirrors the probe's return shape.
type pageAudit struct {
	Contrast []struct {
		Cls       string  `json:"cls"`
		Tag       string  `json:"tag"`
		Text      string  `json:"text"`
		FG        string  `json:"fg"`
		BG        string  `json:"bg"`
		Ratio     float64 `json:"ratio"`
		Need      float64 `json:"need"`
		OverImage bool    `json:"overImage"`
		Px        int     `json:"px"`
	} `json:"contrast"`
	Images []struct {
		Src string `json:"src"`
		Alt string `json:"alt"`
	} `json:"images"`
	Overflow *struct {
		ScrollWidth int `json:"scrollWidth"`
		Viewport    int `json:"viewport"`
	} `json:"overflow"`
}

// Execute audits every requested URL. One unreachable page does not abort the
// run: the rest are still worth measuring, and the failure is reported.
func (a *RenderAuditAction) Execute(ctx context.Context, req RenderAuditRequest) (*RenderAuditResult, error) {
	if len(req.URLs) == 0 {
		return nil, fmt.Errorf("render_audit: no urls in request")
	}
	res := &RenderAuditResult{RunID: req.RunID}
	failedPages := map[string]bool{}

	for _, url := range req.URLs {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		pa, err := a.auditOne(ctx, url)
		res.Summary.Pages++
		if err != nil {
			a.logger.Warn("render_audit: page unreachable",
				zap.String("url", url), zap.Error(err))
			res.Unreachable = append(res.Unreachable, url)
			failedPages[url] = true
			continue
		}
		for _, c := range pa.Contrast {
			res.Contrast = append(res.Contrast, ContrastFinding{
				URL: url, Tag: c.Tag, Class: c.Cls, Text: c.Text, FG: c.FG, BG: c.BG,
				Ratio: c.Ratio, Need: c.Need, FontPx: c.Px, OverImage: c.OverImage,
			})
			if !c.OverImage {
				res.Summary.ContrastFirm++
				failedPages[url] = true
			}
		}
		for _, im := range pa.Images {
			res.Images = append(res.Images, BrokenImage{URL: url, Src: im.Src, Alt: im.Alt})
			failedPages[url] = true
		}
		if pa.Overflow != nil {
			res.Overflow = append(res.Overflow, OverflowFinding{
				URL: url, ScrollWidth: pa.Overflow.ScrollWidth, Viewport: pa.Overflow.Viewport,
			})
			res.Summary.OverflowPages++
		}
	}

	res.Summary.Contrast = len(res.Contrast)
	res.Summary.BrokenImages = len(res.Images)
	res.Summary.PagesFailed = len(failedPages)
	return res, nil
}

func (a *RenderAuditAction) auditOne(ctx context.Context, url string) (*pageAudit, error) {
	// openChromium takes a profile NAME, not dimensions — the viewport sizes are
	// its own constants, which keeps every browser run on this adapter measuring
	// the same two profiles. Overflow is only meaningful against a known width.
	cp, err := a.open(ctx, url, "desktop", a.logger)
	if err != nil {
		return nil, err
	}
	defer cp.Close()
	if ne := cp.NavError(); ne != "" {
		return nil, fmt.Errorf("navigation: %s", ne)
	}

	// Give late images a moment to settle before asking which failed, or a slow
	// CDN response reads as a broken image. The Python re-checked them serially
	// after the fact for the same reason.
	select {
	case <-time.After(1500 * time.Millisecond):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	v, err := cp.Evaluate(auditJS)
	if err != nil {
		return nil, fmt.Errorf("probe: %w", err)
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("probe marshal: %w", err)
	}
	var pa pageAudit
	if err := json.Unmarshal(raw, &pa); err != nil {
		return nil, fmt.Errorf("probe unmarshal: %w", err)
	}
	return &pa, nil
}


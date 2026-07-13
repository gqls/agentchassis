# Go batch — slice 2: component_library.go (fallback chrome + Debug tidy)

Base: the working copy you uploaded 2026-07-03 16:13 — both OLD blocks below are byte-faithful
to it. These are the safety-net functions: they fire only when no active `site-header` /
`site-footer` component exists, so runtime verification defers to whenever that path occurs;
the checks here are `go build ./...` and review. Independent of slice 1's deploy — can ride
the same image as the Edit-B correction if one is needed.

## Edit C — RenderFallbackHeader: consume the chrome variables
The chrome pair variables flip with scheme on all 18 layouts (Checks 1b/2b/3e), so the
fallback consuming them is light on light sites and dark on dark ones automatically.
Compile note: the `primary` and `accent` locals are removed (they become unused — Go
errors on unused variables) and the Sprintf argument list shrinks to two. `ctx` remains
used (LogoText, NavItems). No other identifiers change.

OLD (exact, whole function):
```go
func RenderFallbackHeader(ctx *RenderContext) string {
	navHTML := buildNavItemsHTML(ctx.NavItems)
	primary := defaultString(ctx.PrimaryColor, "#1a1a2e")
	accent := defaultString(ctx.AccentColor, "#16a085")

	return fmt.Sprintf(`<header class="site-header">
    <div class="header-container">
        <a href="/index.html" class="logo">%s</a>
        <button class="mobile-menu-toggle" aria-label="Toggle menu"><span></span><span></span><span></span></button>
        <nav class="main-nav">
            <ul>%s</ul>
        </nav>
    </div>
</header>
<style>
.site-header{background:%s;padding:1rem 0;position:sticky;top:0;z-index:1000;box-shadow:0 2px 10px rgba(0,0,0,.1)}
.header-container{max-width:1200px;margin:0 auto;padding:0 2rem;display:flex;align-items:center;justify-content:space-between}
.logo{text-decoration:none;font-size:1.5rem;font-weight:700;color:#fff}
.main-nav ul{display:flex;list-style:none;margin:0;padding:0;gap:2rem}
.main-nav a{color:rgba(255,255,255,.9);text-decoration:none;font-weight:500;transition:color .2s}
.main-nav a:hover,.main-nav a.active{color:%s}
.mobile-menu-toggle{display:none;background:none;border:none;cursor:pointer;padding:.5rem}
.mobile-menu-toggle span{display:block;width:24px;height:2px;background:#fff;margin:5px 0}
@media(max-width:768px){.mobile-menu-toggle{display:block}.main-nav{position:absolute;top:100%%;left:0;right:0;background:%s;padding:1rem;display:none}.main-nav.active{display:block}.main-nav ul{flex-direction:column;gap:0}.main-nav a{display:block;padding:.75rem 0;border-bottom:1px solid rgba(255,255,255,.1)}}
</style>
<script>document.addEventListener("DOMContentLoaded",function(){var t=document.querySelector(".mobile-menu-toggle"),n=document.querySelector(".main-nav");t&&n&&t.addEventListener("click",function(){n.classList.toggle("active")})});</script>`,
		ctx.LogoText, navHTML, primary, accent, primary)
}
```
NEW:
```go
func RenderFallbackHeader(ctx *RenderContext) string {
	navHTML := buildNavItemsHTML(ctx.NavItems)

	return fmt.Sprintf(`<header class="site-header">
    <div class="header-container">
        <a href="/index.html" class="logo">%s</a>
        <button class="mobile-menu-toggle" aria-label="Toggle menu"><span></span><span></span><span></span></button>
        <nav class="main-nav">
            <ul>%s</ul>
        </nav>
    </div>
</header>
<style>
.site-header{background:var(--color-header-bg, var(--color-surface));padding:1rem 0;position:sticky;top:0;z-index:1000;box-shadow:0 2px 10px rgba(0,0,0,.1)}
.header-container{max-width:1200px;margin:0 auto;padding:0 2rem;display:flex;align-items:center;justify-content:space-between}
.logo{text-decoration:none;font-size:1.5rem;font-weight:700;color:var(--color-header-text, var(--color-text))}
.main-nav ul{display:flex;list-style:none;margin:0;padding:0;gap:2rem}
.main-nav a{color:color-mix(in srgb, var(--color-header-text, var(--color-text)) 90%%, transparent);text-decoration:none;font-weight:500;transition:color .2s}
.main-nav a:hover,.main-nav a.active{color:var(--color-accent)}
.mobile-menu-toggle{display:none;background:none;border:none;cursor:pointer;padding:.5rem}
.mobile-menu-toggle span{display:block;width:24px;height:2px;background:var(--color-header-text, var(--color-text));margin:5px 0}
@media(max-width:768px){.mobile-menu-toggle{display:block}.main-nav{position:absolute;top:100%%;left:0;right:0;background:var(--color-header-bg, var(--color-surface));padding:1rem;display:none}.main-nav.active{display:block}.main-nav ul{flex-direction:column;gap:0}.main-nav a{display:block;padding:.75rem 0;border-bottom:1px solid color-mix(in srgb, var(--color-header-text, var(--color-text)) 10%%, transparent)}}
</style>
<script>document.addEventListener("DOMContentLoaded",function(){var t=document.querySelector(".mobile-menu-toggle"),n=document.querySelector(".main-nav");t&&n&&t.addEventListener("click",function(){n.classList.toggle("active")})});</script>`,
		ctx.LogoText, navHTML)
}
```
NB every literal `%` inside CSS gains Sprintf escaping (`90%%`, `10%%`) exactly as the
existing `top:100%%` does — miss one and Sprintf emits `%!t(MISSING)`-style noise.

## Edit D — RenderFallbackFooter: same treatment
`primary` local removed; Sprintf args shrink by one (the trailing `primary`).

OLD (exact, whole function):
```go
func RenderFallbackFooter(ctx *RenderContext) string {
	navHTML := buildNavItemsHTML(ctx.NavItems)
	primary := defaultString(ctx.PrimaryColor, "#1a1a2e")
	year := ctx.Year
	if year == "" {
		year = fmt.Sprintf("%d", time.Now().Year())
	}

	return fmt.Sprintf(`<footer class="site-footer">
    <div class="footer-container">
        <div class="footer-brand"><h3>%s</h3><p>%s</p></div>
        <div class="footer-links"><h4>Links</h4><ul>%s</ul></div>
        <div class="footer-contact"><h4>Contact</h4><p>%s</p></div>
    </div>
    <div class="footer-bottom"><p>&copy; %s %s. All rights reserved.</p></div>
</footer>
<style>
.site-footer{background:%s;color:rgba(255,255,255,.9);padding:3rem 0 0;margin-top:auto}
.footer-container{max-width:1200px;margin:0 auto;padding:0 2rem;display:grid;grid-template-columns:2fr 1fr 1fr;gap:2rem}
.footer-brand h3,.footer-links h4,.footer-contact h4{color:#fff;margin:0 0 1rem}
.footer-links ul{list-style:none;padding:0;margin:0}
.footer-links li{margin-bottom:.5rem}
.footer-links a{color:rgba(255,255,255,.7);text-decoration:none}
.footer-links a:hover{color:#fff}
.footer-bottom{margin-top:2rem;padding:1.5rem 0;border-top:1px solid rgba(255,255,255,.1);text-align:center}
.footer-bottom p{margin:0;color:rgba(255,255,255,.6);font-size:.9rem}
@media(max-width:768px){.footer-container{grid-template-columns:1fr}}
</style>`, ctx.LogoText, ctx.Tagline, navHTML, ctx.Email, year, ctx.CompanyName, primary)
}
```
NEW:
```go
func RenderFallbackFooter(ctx *RenderContext) string {
	navHTML := buildNavItemsHTML(ctx.NavItems)
	year := ctx.Year
	if year == "" {
		year = fmt.Sprintf("%d", time.Now().Year())
	}

	return fmt.Sprintf(`<footer class="site-footer">
    <div class="footer-container">
        <div class="footer-brand"><h3>%s</h3><p>%s</p></div>
        <div class="footer-links"><h4>Links</h4><ul>%s</ul></div>
        <div class="footer-contact"><h4>Contact</h4><p>%s</p></div>
    </div>
    <div class="footer-bottom"><p>&copy; %s %s. All rights reserved.</p></div>
</footer>
<style>
.site-footer{background:var(--color-footer-bg, var(--color-surface));color:color-mix(in srgb, var(--color-footer-text, var(--color-text)) 90%%, transparent);padding:3rem 0 0;margin-top:auto}
.footer-container{max-width:1200px;margin:0 auto;padding:0 2rem;display:grid;grid-template-columns:2fr 1fr 1fr;gap:2rem}
.footer-brand h3,.footer-links h4,.footer-contact h4{color:var(--color-footer-text, var(--color-text));margin:0 0 1rem}
.footer-links ul{list-style:none;padding:0;margin:0}
.footer-links li{margin-bottom:.5rem}
.footer-links a{color:color-mix(in srgb, var(--color-footer-text, var(--color-text)) 70%%, transparent);text-decoration:none}
.footer-links a:hover{color:var(--color-footer-text, var(--color-text))}
.footer-bottom{margin-top:2rem;padding:1.5rem 0;border-top:1px solid color-mix(in srgb, var(--color-footer-text, var(--color-text)) 10%%, transparent);text-align:center}
.footer-bottom p{margin:0;color:color-mix(in srgb, var(--color-footer-text, var(--color-text)) 60%%, transparent);font-size:.9rem}
@media(max-width:768px){.footer-container{grid-template-columns:1fr}}
</style>`, ctx.LogoText, ctx.Tagline, navHTML, ctx.Email, year, ctx.CompanyName)
}
```

## RenderFallbackHead — deliberately unchanged
Its only colour use is `<meta name="theme-color" content="%s">` — a meta value, not CSS,
so `var()` cannot work there; `ctx.PrimaryColor` is the legitimate source. No edit.

## Edit E — the eight logger.Debug → logger.Info swaps
Mechanical, one file, per the no-Debug rule. Lines (in the 16:13 copy): 510
(RenderTemplate executing), 1460 (InjectHeader nav update), 1505 (InjectFooter nav
update), 1675/1682/1688 (InjectHead removed/inserted-before-body/inserted-after-html),
1736 (GetHeaderNavFromPages empty), 1784 (GetHeaderNavFromPages built). Swap the method
name only; arguments untouched.

## Build + deploy
`go build ./...`; rides the next chassis image (can share the Edit-B correction deploy if
one comes out of w8_03). No data steps; no page rebuilds required — the fallbacks only
matter on the no-active-component path.

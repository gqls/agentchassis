-- 758_news_scripts_build_elements_instead_of_concatenating_html.sql
--
-- bugs_open/472. The `news-listing` and `latest-news` components build their
-- markup by string concatenation and hand the result to innerHTML, with
-- third-party feed text interpolated UNESCAPED:
--
--     html += "<p class=\"news-list-item-summary\">" + item.summary + "</p>";
--     ...
--     container.innerHTML = html;
--
-- `item.summary` comes from /data/news-archive.json, which loadNewsItems writes
-- straight from content_feed_items.source_summary — text scraped from other
-- people's websites. Nothing on that path escapes it (the server-rendered path
-- does; the JSON path never has).
--
-- EXPOSURE, MEASURED 2026-09-03, and stated so nobody inflates it: 14 of 5,863
-- feed rows carry any HTML markup at all, ZERO carry anything script-ish
-- (<script, onerror=, onload=, javascript:), and ZERO of the 20 items served in
-- boxingonline's archive JSON do. This is an exposure to CLOSE, not a live
-- vulnerability, and it must not be written up as one.
--
-- ⚠ The reason that number is low is a MECHANISM that can change without anyone
-- noticing: the RSS ingest path calls stripHTML (feed_actions.go:248) and the
-- web-search path does not. A new ingest source inherits whichever path it is
-- wired to, so a future non-zero here is this same bug arriving, not a new one.
--
-- WHY ELEMENTS RATHER THAN AN esc() HELPER. This is the components lane's
-- recommendation, adopted, after they censused every active component at this
-- lane's request. Quote the SPLIT, never the total: 23 components contain
-- `innerHTML` and only these 2 are defects.
--   2  DEFECT  — fetch JSON, accumulate with `html +=`, assign (news-listing 13
--                accumulations, latest-news 6)
--  12  SAFE    — the directory/tracker family: innerHTML ONLY to clear
--                (container.innerHTML = ""), all data through textContent
--   1  SAFE    — webdesign-couk-header: a complete esc() over & < > " ' applied
--                to every interpolated value (checked for a gap; there is none)
--   8  SAFE    — calculators over locally computed numbers or createTextNode
-- Copying the 12 removes the CLASS. An escape helper only filters it, and
-- regresses the moment someone adds a fourth interpolated field and forgets to
-- wrap it — which is exactly how this defect arrives.
--
-- ALSO FIXED, and it is the hole an escape helper would have left open: the
-- anchor href was set from `item.url` / `data.insights_url` with no scheme
-- check, so a `javascript:` URL in a feed would have been live. TWO helpers now
-- guard it, and the split is not stylistic: safeExternalHref() admits only
-- http(s) and takes the third-party feed URL; safeInternalHref() also admits a
-- site-relative path and takes `data.insights_url`, which comes from
-- `pages.url` (render_news_section_action.go:213-218) and is `/news.html`,
-- `/news/index.html` or `/noticias/index.html` on every live site
-- [MEASURED 2026-09-03]. The FIRST CUT of this migration used one helper for
-- both and turned the "More insights" link into href="#" on every site with a
-- news index — a live regression introduced by the security fix, caught by the
-- components lane on review. Its `^\/(?!\/)` arm admits `/news.html` and still
-- rejects `//evil.com`, which is protocol-relative and leaves the origin.
--
-- Behaviour, executed rather than argued (2026-09-03):
--   input                    external()  internal()
--   https://espn.com/x       unchanged   unchanged
--   /news.html               #           /news.html
--   //evil.com/x             #           #
--   javascript:alert(1)      #           #
--   data:text/html,...       #           #
--   evil.com  /  (empty)     #           #
--
-- MINOR, deliberate, flagged by the same review: the fallback label is now the
-- literal character "More insights →" where the old markup had `&rarr;`. That is
-- correct for textContent. But if a future `data.insights_label` is set in the
-- database containing an HTML entity, textContent renders it literally where
-- innerHTML used to resolve it. Every live insights_label is empty today, so
-- nothing is affected — noted so a future reader is not surprised.
--
-- BEHAVIOUR PRESERVED EXACTLY: the same elements, the same class names, the
-- same order, the same conditionals, formatNewsDate untouched, and
-- hasServerRenderedItems still guarding only the empty-feed and fetch-failed
-- branches. This migration changes HOW the nodes are made, not which.
--
-- ⚠ _HOLD, DELIBERATELY — NOT for the runner. Apply this BY HAND, and only
-- after someone has loaded a news page in a real browser afterwards.
--
-- The reason is specific, not caution for its own sake: this rewrites the DOM
-- construction of a live component on five customer sites, and THE USUAL
-- VERIFICATION CANNOT SEE THE RESULT. A static curl fetches the server-rendered
-- HTML, which this script REPLACES on load — so every check available to an
-- automated apply would pass on a page that renders empty. See LANDMINES, "The
-- served news page HTML is OVERWRITTEN in the browser by
-- /data/news-archive.json". The verify block below proves the ROW is right; only
-- a browser proves the PAGE is.
--
-- Rehearsed 2026-09-03 under BEGIN/ROLLBACK against the live rows: both UPDATEs
-- hit exactly 1 row, all three post-conditions true on both components. The
-- verify block was then INDUCED to fail by removing the second UPDATE, and it
-- aborted with 'still concatenate markup' — so it bites, and is not decoration.
--
-- Live on apply. No image, no roll. The published /tools/assets/*.js follow on
-- each site's next render.

BEGIN;

CREATE TEMP TABLE js_bak_472 AS
  SELECT id, function, js_content FROM content_components
   WHERE function IN ('news-listing', 'latest-news');

UPDATE content_components SET js_content = $js$  (function() {
    function formatNewsDate(s) {
      if (!s) return "";
      s = s.replace(/^(\d+)d\s*ago$/i, function (_, n) { return n + (n === "1" ? " day ago" : " days ago"); });
      s = s.replace(/^(\d+)h\s*ago$/i, function (_, n) { return n + (n === "1" ? " hour ago" : " hours ago"); });
      s = s.replace(/^(\d+)m\s*ago$/i, function (_, n) { return n + (n === "1" ? " minute ago" : " minutes ago"); });
      s = s.replace(/^(\d+)w\s*ago$/i, function (_, n) { return n + (n === "1" ? " week ago" : " weeks ago"); });
      return s;
    }
    // bugs_open/472: build nodes, never markup. textContent cannot be markup,
    // so third-party feed text has no path to the parser.
    function el(tag, cls, text) {
      var e = document.createElement(tag);
      if (cls) { e.className = cls; }
      if (text !== undefined && text !== null && text !== "") { e.textContent = text; }
      return e;
    }
    // An href is the one attribute that can still execute. TWO helpers, not one,
    // because the two URLs on this page have different trust AND different
    // shape — and a single helper over both is what produced the regression the
    // components lane caught in this migration's first cut: a dead "More
    // insights" link on every site with a news index.
    //   item.url          third-party, from the feed, ABSOLUTE  -> external
    //   data.insights_url internal, from pages.url, RELATIVE    -> internal
    function safeExternalHref(u) { return /^https?:\/\//i.test(u || "") ? u : "#"; }
    function safeInternalHref(u) {
      u = u || "";
      if (/^https?:\/\//i.test(u)) { return u; }
      if (/^\/(?!\/)/.test(u)) { return u; }   // /news.html yes, //evil.com no
      return "#";
    }
    // Server-rendered items (bugs_open/027) must survive an empty feed, a 404,
    // or any fetch failure. Only show a placeholder when there is nothing to
    // preserve — otherwise leave the server's HTML exactly where it is.
    function hasServerRenderedItems(container) {
      return !!(container && container.querySelector("article.news-list-item"));
    }
    fetch("/data/news-archive.json")
      .then(function(r) { return r.json(); })
      .then(function(data) {
        var container = document.getElementById("news-listing-items");
        var footer = document.getElementById("news-listing-footer");
        if (!data.items || data.items.length === 0) {
          if (!hasServerRenderedItems(container)) {
            container.innerHTML = "";
            container.appendChild(el("p", "news-listing-empty", "No news items available yet. Check back soon."));
          }
          return;
        }
        container.innerHTML = "";
        data.items.forEach(function(item) {
          var article = el("article", "news-list-item");
          var content = el("div", "news-list-item-content");
          var title = el("h3", "news-list-item-title");
          var link = el("a", null, item.title);
          link.setAttribute("href", safeExternalHref(item.url));
          link.setAttribute("target", "_blank");
          link.setAttribute("rel", "noopener noreferrer");
          title.appendChild(link);
          content.appendChild(title);
          if (item.summary) {
            content.appendChild(el("p", "news-list-item-summary", item.summary));
          }
          var meta = el("div", "news-list-item-meta");
          if (item.source) { meta.appendChild(el("span", "news-list-item-source", item.source)); }
          if (item.date) { meta.appendChild(el("span", "news-list-item-date", formatNewsDate(item.date))); }
          content.appendChild(meta);
          if (item.topics) {
            var topics = el("div", "news-list-item-topics");
            item.topics.split(", ").forEach(function(tag) {
              topics.appendChild(el("span", "news-list-tag", tag));
            });
            content.appendChild(topics);
          }
          article.appendChild(content);
          container.appendChild(article);
        });
        footer.style.display = "block";
        var countEl = document.getElementById("news-listing-count");
        if (data.items_total && data.items_total > data.item_count) {
          countEl.textContent = "Showing " + data.item_count + " of " + data.items_total + " items";
        } else {
          countEl.textContent = data.item_count + " items";
        }
        var updatedEl = document.getElementById("news-listing-updated");
        if (data.updated_at) {
          var d = new Date(data.updated_at);
          updatedEl.textContent = "Last updated: " + d.toLocaleDateString() + " " + d.toLocaleTimeString();
        }
      })
      .catch(function(err) {
        var container = document.getElementById("news-listing-items");
        if (!hasServerRenderedItems(container)) {
          container.innerHTML = "";
          container.appendChild(el("p", "news-listing-empty", "Unable to load news. Please try again later."));
        }
      });
  })();$js$, updated_at = now()
WHERE function = 'news-listing';

UPDATE content_components SET js_content = $js$(function() {
  function formatNewsDate(s) {
    if (!s) return "";
    s = s.replace(/^(\d+)d\s*ago$/i, function (_, n) { return n + (n === "1" ? " day ago" : " days ago"); });
    s = s.replace(/^(\d+)h\s*ago$/i, function (_, n) { return n + (n === "1" ? " hour ago" : " hours ago"); });
    s = s.replace(/^(\d+)m\s*ago$/i, function (_, n) { return n + (n === "1" ? " minute ago" : " minutes ago"); });
    s = s.replace(/^(\d+)w\s*ago$/i, function (_, n) { return n + (n === "1" ? " week ago" : " weeks ago"); });
    return s;
  }
  // bugs_open/472: build nodes, never markup.
  function el(tag, cls, text) {
    var e = document.createElement(tag);
    if (cls) { e.className = cls; }
    if (text !== undefined && text !== null && text !== "") { e.textContent = text; }
    return e;
  }
  // Two helpers, not one — see the news-listing script above for why.
  function safeExternalHref(u) { return /^https?:\/\//i.test(u || "") ? u : "#"; }
  function safeInternalHref(u) {
    u = u || "";
    if (/^https?:\/\//i.test(u)) { return u; }
    if (/^\/(?!\/)/.test(u)) { return u; }   // /news.html yes, //evil.com no
    return "#";
  }
  fetch("/data/latest-news.json")
    .then(function(r) { return r.json(); })
    .then(function(data) {
      if (data.headline) {
        document.getElementById("news-headline").textContent = data.headline;
      }
      var sub = document.getElementById("news-subheadline");
      if (sub && data.subheadline) {
        sub.textContent = data.subheadline;
      }
      var container = document.getElementById("news-container");
      if (data.items && data.items.length > 0) {
        container.innerHTML = "";
        data.items.forEach(function(item) {
          var article = el("article", "news-card");
          var content = el("div", "news-card-content");
          var title = el("h3", "news-card-title");
          var link = el("a", null, item.title);
          link.setAttribute("href", safeExternalHref(item.url));
          link.setAttribute("target", "_blank");
          link.setAttribute("rel", "noopener noreferrer");
          title.appendChild(link);
          content.appendChild(title);
          if (item.summary) {
            content.appendChild(el("p", "news-card-summary", item.summary));
          }
          var meta = el("div", "news-card-meta");
          if (item.source) { meta.appendChild(el("span", "news-source", item.source)); }
          if (item.date) { meta.appendChild(el("time", "news-date", formatNewsDate(item.date))); }
          content.appendChild(meta);
          article.appendChild(content);
          container.appendChild(article);
        });
      }
      if (data.insights_url) {
        var f = document.getElementById("news-footer");
        f.innerHTML = "";
        var wrap = el("div", "news-section-footer");
        var more = el("a", "news-more-link", data.insights_label || "More insights →");
        more.setAttribute("href", safeInternalHref(data.insights_url));
        wrap.appendChild(more);
        f.appendChild(wrap);
      }
    })
    .catch(function() {});
})();$js$, updated_at = now()
WHERE function = 'latest-news';

-- Verify: DO/RAISE, because a block of SELECTs cannot stop the COMMIT
-- (ON_ERROR_STOP ignores a non-empty result — bugs_open/RFC_006's trap).
DO $verify$
DECLARE
  n_defect  int;
  n_text    int;
  n_href    int;
  n_rel     int;
  n_rows    int;
BEGIN
  SELECT count(*) INTO n_rows FROM content_components WHERE function IN ('news-listing','latest-news');
  IF n_rows <> 2 THEN
    RAISE EXCEPTION '758: expected exactly 2 component rows, found %', n_rows;
  END IF;

  -- No interpolation into markup survives anywhere in either script.
  SELECT count(*) INTO n_defect FROM content_components
   WHERE function IN ('news-listing','latest-news')
     AND (js_content LIKE '%html +=%' OR js_content LIKE '%innerHTML = html%');
  IF n_defect <> 0 THEN
    RAISE EXCEPTION '758: % component(s) still concatenate markup', n_defect;
  END IF;

  -- The positive control: the replacement pattern is actually present. Without
  -- this, an UPDATE that wrote an empty string would satisfy the check above.
  SELECT count(*) INTO n_text FROM content_components
   WHERE function IN ('news-listing','latest-news') AND js_content LIKE '%textContent = text%';
  IF n_text <> 2 THEN
    RAISE EXCEPTION '758: expected 2 components using the textContent helper, found %', n_text;
  END IF;

  -- Both helpers present. NOT sufficient on its own — see the next check.
  SELECT count(*) INTO n_href FROM content_components
   WHERE function IN ('news-listing','latest-news')
     AND js_content LIKE '%safeExternalHref%' AND js_content LIKE '%safeInternalHref%';
  IF n_href <> 2 THEN
    RAISE EXCEPTION '758: expected 2 components with both href helpers, found %', n_href;
  END IF;

  -- ⚠ THE CHECK THE FIRST CUT DID NOT HAVE, AND THE REASON IT SHIPPED A
  -- REGRESSION. A `LIKE '%safeHref%'` post-condition is satisfied by a helper
  -- that BREAKS every internal link — presence is not behaviour. Assert the
  -- relative-admitting arm itself: without this literal, /news.html becomes
  -- href="#" on every site with a news index and no row-level check can tell.
  SELECT count(*) INTO n_rel FROM content_components
   WHERE function IN ('news-listing','latest-news')
     AND js_content LIKE '%/^\\/(?!\\/)/.test(u)%';
  IF n_rel <> 2 THEN
    RAISE EXCEPTION '758: expected 2 components admitting site-relative hrefs, found % — the More-insights link would be dead', n_rel;
  END IF;

  RAISE NOTICE '758 OK: 2 components build elements, use textContent, and guard href.';
END
$verify$;

COMMIT;

-- Verify after apply (read the artefact, not the row):
--   SELECT function, length(js_content),
--          js_content LIKE '%html +=%'          AS still_concatenates,  -- expect f
--          js_content LIKE '%textContent = text%' AS uses_textcontent,  -- expect t
--          js_content LIKE '%safeExternalHref%'   AS guards_external,     -- expect t
--          js_content LIKE '%safeInternalHref%'   AS guards_internal,     -- expect t
--          js_content LIKE '%/^\\/(?!\\/)/.test(u)%' AS admits_relative  -- expect t
--     FROM content_components WHERE function IN ('news-listing','latest-news');
--
-- Then AT THE SERVED ASSET, after each site's next render:
--   curl -s https://<host>/tools/assets/news-listing.js | grep -c 'innerHTML = html'   # expect 0
--   curl -s https://<host>/tools/assets/news-listing.js | grep -c 'textContent'        # expect >0
--   curl -s https://<host>/tools/assets/zzz-not-real.js -o /dev/null -w '%{http_code}' # control: 404
--
-- ⚠ A static curl CANNOT verify the rendered result — the script REPLACES the
-- server HTML on load. See LANDMINES, "The served news page HTML is OVERWRITTEN
-- in the browser by /data/news-archive.json".

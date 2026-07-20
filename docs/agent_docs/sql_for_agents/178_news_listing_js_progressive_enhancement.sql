-- ============================================================================
-- 178 — make `news-listing` JS progressive-enhancement safe (bugs_open/027, part A)
--
-- WHY THIS EXISTS, AND WHY IT MUST LAND FIRST
--
-- bugs_open/027: every news page on the platform serves ZERO news to a consumer
-- that does not execute JavaScript. The fix is to server-render the items into
-- `page_components.rendered_html` (part B, a Go change in
-- RenderNewsSectionAction).
--
-- Part B is a REGRESSION unless this lands first. The two news scripts are not
-- alike — verified, not assumed:
--
--   latest-news  (homepage, js_content 2174 bytes)
--       writes the container ONLY when the fetch returned items, and swallows
--       failures. Already progressive-enhancement safe. NOT touched here.
--
--   news-listing (archive page, js_content 3092 bytes)
--       overwrites the container in BOTH failure modes:
--         empty feed -> innerHTML = "No news items available yet..."
--         fetch error -> innerHTML = "Unable to load news..."
--       So server-rendered items would be DESTROYED by the very script meant to
--       enhance them — on an empty feed, on a 404, on any storage blip.
--
-- That asymmetry is exactly the fix-one-branch-and-call-it-done failure this
-- platform keeps hitting (016b §9). It was one `curl` away from shipping, and
-- was caught by the vetcomparison thread's addendum to bugs_open/027 rather
-- than by reading the homepage script and generalising from it.
--
-- WHAT CHANGES
--
-- One guard, applied to both failure paths: only write a placeholder message if
-- the container holds no server-rendered article. Fresh items still replace
-- server HTML whenever the fetch succeeds, so currency is unchanged and the
-- JSON stays authoritative for freshness. Nothing else in the script moves.
--
-- SCOPE: `content_components` is a shared library table with no site_id, so this
-- single row covers every site that renders a news listing (relojistas.com,
-- robot-hands.com, gaswholesalers.com, idea.uk today). The deployed asset is
-- emitted per-site as /tools/assets/news-listing.js from this column by
-- rerender_single_page_action.go:174 — so sites pick the change up on their next
-- rerender, not instantly.
--
-- SAFE TO APPLY BEFORE PART B: with no server-rendered content present, the
-- guard is always true and behaviour is byte-for-byte what it is today.
-- ============================================================================

BEGIN;

-- Snapshot before overwriting (this column has no version history of its own).
CREATE TABLE IF NOT EXISTS component_js_backup_20260720_news_listing AS
SELECT id, function, js_content, now() AS backed_up_at
  FROM content_components
 WHERE function = 'news-listing';

UPDATE content_components SET
  js_content = $js$  (function() {
    function formatNewsDate(s) {
      if (!s) return "";
      s = s.replace(/^(\d+)d\s*ago$/i, function (_, n) { return n + (n === "1" ? " day ago" : " days ago"); });
      s = s.replace(/^(\d+)h\s*ago$/i, function (_, n) { return n + (n === "1" ? " hour ago" : " hours ago"); });
      s = s.replace(/^(\d+)m\s*ago$/i, function (_, n) { return n + (n === "1" ? " minute ago" : " minutes ago"); });
      s = s.replace(/^(\d+)w\s*ago$/i, function (_, n) { return n + (n === "1" ? " week ago" : " weeks ago"); });
      return s;
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
            container.innerHTML = "<p class=\"news-listing-empty\">No news items available yet. Check back soon.</p>";
          }
          return;
        }
        var html = "";
        data.items.forEach(function(item) {
          html += "<article class=\"news-list-item\">";
          html += "<div class=\"news-list-item-content\">";
          html += "<h3 class=\"news-list-item-title\"><a href=\"" + item.url + "\" target=\"_blank\" rel=\"noopener noreferrer\">" + item.title + "</a></h3>";
          if (item.summary) {
            html += "<p class=\"news-list-item-summary\">" + item.summary + "</p>";
          }
          html += "<div class=\"news-list-item-meta\">";
          if (item.source) {
            html += "<span class=\"news-list-item-source\">" + item.source + "</span>";
          }
          if (item.date) {
            html += "<span class=\"news-list-item-date\">" + formatNewsDate(item.date) + "</span>";
          }
          html += "</div>";
          if (item.topics) {
            html += "<div class=\"news-list-item-topics\">";
            item.topics.split(", ").forEach(function(tag) {
              html += "<span class=\"news-list-tag\">" + tag + "</span>";
            });
            html += "</div>";
          }
          html += "</div>";
          html += "</article>";
        });
        container.innerHTML = html;
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
          container.innerHTML = "<p class=\"news-listing-empty\">Unable to load news. Please try again later.</p>";
        }
      });
  })();$js$,
  updated_at = now()
WHERE function = 'news-listing';

COMMIT;

-- ----------------------------------------------------------------------------
-- Verify:
--   SELECT function,
--          length(js_content)                            AS len,
--          js_content LIKE '%hasServerRenderedItems%'     AS guard_present,
--          (length(js_content) - length(replace(js_content,'hasServerRenderedItems','')))/22
--                                                        AS guard_mentions  -- expect 3
--     FROM content_components WHERE function = 'news-listing';
--
-- Rollback:
--   UPDATE content_components c SET js_content = b.js_content
--     FROM component_js_backup_20260720_news_listing b
--    WHERE c.id = b.id;
--
-- NOTE: sites do not pick this up until their next rerender writes
-- /tools/assets/news-listing.js. Verify on the BOX, not in the DB:
--   ssh root@<box> 'grep -c hasServerRenderedItems \
--     /var/www/vm-sites/<domain>/tools/assets/news-listing.js'
-- ----------------------------------------------------------------------------

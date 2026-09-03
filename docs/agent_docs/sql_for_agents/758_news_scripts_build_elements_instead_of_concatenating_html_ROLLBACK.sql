-- ROLLBACK for 758 (bugs_open/472).
--
-- Restores the pre-758 js_content for `news-listing` and `latest-news` verbatim
-- from the live rows as they stood 2026-09-03, read back from
-- content_components before the migration was written.
--
-- ⚠ Rolling this back RESTORES a defect: both scripts go back to concatenating
-- unescaped third-party feed text into innerHTML, and the anchor href goes back
-- to accepting any scheme. Only run it if 758 broke rendering, and file what
-- broke — the exposure it re-opens is small (14 of 5,863 feed rows carry any
-- HTML, none executable, 2026-09-03) but it is not nothing.

BEGIN;

UPDATE content_components SET js_content = $js$  (function() {
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
        container.innerHTML = data.items.map(function(item) {
          var html = "<article class=\"news-card\"><div class=\"news-card-content\">";
          html += "<h3 class=\"news-card-title\"><a href=\"" + item.url + "\" target=\"_blank\" rel=\"noopener noreferrer\">" + item.title + "</a></h3>";
          if (item.summary) {
            html += "<p class=\"news-card-summary\">" + item.summary + "</p>";
          }
          html += "<div class=\"news-card-meta\">";
          if (item.source) {
            html += "<span class=\"news-source\">" + item.source + "</span>";
          }
          if (item.date) {
            html += "<time class=\"news-date\">" + formatNewsDate(item.date) + "</time>";
          }
          html += "</div></div></article>";
          return html;
        }).join("");
      }
      if (data.insights_url) {
        document.getElementById("news-footer").innerHTML =
          "<div class=\"news-section-footer\"><a href=\"" + data.insights_url +
          "\" class=\"news-more-link\">" + (data.insights_label || "More insights &rarr;") +
          "</a></div>";
      }
    })
    .catch(function() {});
})();$js$, updated_at = now()
WHERE function = 'latest-news';

DO $verify$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM content_components
   WHERE function IN ('news-listing','latest-news') AND js_content LIKE '%html +=%';
  IF n <> 2 THEN
    RAISE EXCEPTION 'ROLLBACK 758: expected 2 components restored to concatenation, found %', n;
  END IF;
  RAISE NOTICE 'ROLLBACK 758 OK: both components restored to their pre-758 content.';
END
$verify$;

COMMIT;

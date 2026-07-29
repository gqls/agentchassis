# idea.uk demand — where we are (plain prose, append-only, newest at the bottom)

**2026-07-29 — the thread opens, and the first finding is embarrassing in a
useful way.** You asked us to start thinking about how to get a buyer. Before
proposing anything clever we looked at the one channel that already exists —
people searching Google — and found the front door mislabelled: Google's entry
for idea.uk is still the "this domain is for sale" page from before we owned
it. On top of that, Google's crawler has never once looked at the report page
or any guide — the site gives it no map (no sitemap), and seven pages including
the homepage offer search engines an empty description line.

So the first work is not marketing, it is making the shop visible from the
street: a sitemap and robots file, descriptions on the empty pages, and — this
one needs you, it's a browser login — registering the site in Google Search
Console and asking Google to re-read the homepage so the "for sale" label dies.
Once people CAN find us, the actual demand experiments (the £8 example place,
posting where founders ask "is my idea any good") have a fair test.

**2026-07-29 (late morning) — phase 1 shipped; the Search Console step is yours.**
The sitemap (22 pages, each checked working first), the robots file, and the
missing page descriptions are all live or landing on the next 5-minute sync.
What we cannot do from here is prove to Google that we own the site. The steps:

1. Go to search.google.com/search-console → "Add property" → choose **Domain**
   → enter `idea.uk`. Google shows a TXT record (`google-site-verification=…`).
2. Add that TXT record to the idea.uk zone in the **Hetzner DNS console**
   (idea.uk's nameservers are Hetzner's — this is not Clook).
3. Back in Search Console press Verify. Then: **URL Inspection** → enter
   `https://idea.uk/` → **Request indexing**. That is the single action that
   replaces the "Domain Name For Sale" search result.
4. **Sitemaps** (left menu) → submit `https://idea.uk/sitemap.xml`.
5. Optional but worth the two clicks: bing.com/webmasters → "Import from
   Google Search Console" — covers Bing and DuckDuckGo.

Once the owner side is done, the next work here is phase 2: putting the offer
in front of actual people, with drafts for your sign-off before anything is
posted anywhere.

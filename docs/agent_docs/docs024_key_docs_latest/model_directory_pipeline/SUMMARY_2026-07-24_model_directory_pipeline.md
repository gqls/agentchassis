# SUMMARY — model directory pipeline, 2026-07-24

*Milestone read-out, written to be read aloud. Five parts.*

**What we're trying to do.** Give any site in the fleet a continually
updated, honestly cited AI model directory — every model's price, licence
and specs backed by a verbatim quote from a source page that the platform
itself re-fetches and re-checks on a schedule — with a company AI-adoption
tracker to follow on the same machinery. A site opts in with one flag in its
site spec; the fleet builds the page, feeds it, and keeps it fresh without
per-site work.

**Where we've come from.** Three days ago this was a brief and a plan. We
built the schema (a registry of entities and individually cited claims), a
researcher agent that extracts atomic claims with verbatim quotes and only
registers what a deterministic string check confirms, the rendering layer
(server-baked HTML plus a client-refreshed JSON file, the same dual pattern
the news feed uses), and the opt-in discovery checks that auto-create the
page. All of it committed and deployed, but the registry stayed empty:
every research run died at the web-scrape step.

**What we've done.** Diagnosed and fixed a stack of four defects, each
hidden under the one above, in a scrape path that — it turned out — had
never once worked end-to-end in production: (1) the batch reply was too big
for the message bus and the adapter dropped the refusal silently, starving
callers through twelve minutes of doomed retries (fixed: lean replies,
visible truncation, degrade-once-then-error; council-approved); (2) beneath
that, every reply that *was* deliverable was unparseable — a string where
the platform demands a boolean, a known trap other adapters had already
been fixed for (fixed, with a contract test that round-trips the real
envelope through the real type); (3) beneath *that*, quotes extracted from
a scrape's markdown could never match the verifier's HTML rendering of the
same pricing table (fixed in the shared normaliser: table pipes are
presentation, not content — strictness test-pinned). Along the way we also
caught a scheduled-task routing bug before it shipped, an inert timeout
field both researcher seeds carried, and filed the lot as bugs_open/062
with a §9 pattern.

**Where we are now.** The pipeline is proven live, end to end. The registry
holds its first real data: ten models, twenty-two claims, all verified —
GPT-5.6 Sol at $5.00 in / $30.00 out per million tokens, Sora 2 at $0.10 a
second, image and audio models — each carrying the exact sentence from
OpenAI's live pricing page that proves it, re-checked daily from now on.
The fail-safe design earned its keep during the debugging: unverifiable
claims went to a human-review queue with full detail rather than being
silently registered or dropped, and that queue is what made the final
defect diagnosable in minutes. The last publish leg (committing the JSON to
opted-in sites and refreshing their pages) is seeded and self-gating: it
idles until the auto-created page exists.

**Where we're going.** Next the fleet's own machinery takes over: the
discovery checks see the pilot site's opt-in flag plus a non-empty registry,
raise the page-creation work, and the publish task starts delivering. Then:
widen discovery beyond one vendor's pricing page (the weekly query is
deliberately broad, but the first runs leaned on one source), decide the
owner's three open calls (dedicated page vs section; price re-check
cadence; whether finetuning.uk opts in now), and — once the model half has
bedded in — the adoption tracker lands as new rows in the same tables, no
new machinery.

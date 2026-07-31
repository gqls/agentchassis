#!/usr/bin/env python3
"""Build vonc.com/data/provocations.json from a SCHEDULE, not from a literal.

Phase 0 of provocation_pipeline/PLAN_2026-07-31. Successor to
gauntlet_dead_cta/p4_sources/build_provocations.py, which produced the live
file and has no rotation logic of any kind: `TODAY` was a fixed dict, the
archive dates were string literals, and `generated_at` was a hardcoded string.

WHAT CHANGED, AND WHY EACH ONE WAS REQUIRED
-------------------------------------------
1. `generated_at` is COMPUTED. The old literal ("2026-07-26T00:00:00Z") meant a
   file regenerated a minute ago and one untouched for a month were identical
   in the one field used to tell them apart — so the documented freshness check
   reported a stale file as fresh, for ever. See LANDMINES.md.
2. `today` carries `slug` and `date`. Without them it cannot become an archive
   entry, which is why the archive stopped on 5 Jul while the site kept
   promising a daily.
3. Selection is by DATE, from an ordered schedule. It happens HERE, at
   generation time, and never in the browser: internal/tools-api/handlers/
   round.go reads the `today` key server-side, so a client-side selector would
   make the page display one provocation while the Gauntlet argues another.
4. Promotion is implicit in the schedule and implements the owner's rule
   (2026-07-31) exactly: "it can be archived when the new one is published".
   An entry is `today` on its own day and joins the archive the moment a later
   entry takes over — never during its own day. That also keeps today's full
   case off the archive page, which would otherwise be a third leak path for
   the provocation the Gauntlet is built to hide (gauntlet HANDOFF C).
5. The arena's first card is DERIVED from today. It used to be a hand-copied
   duplicate of the headline, which would go stale the first time anything
   rotated and make the page contradict itself.

FABRICATION RAIL (unchanged, from the 2026-07-26 revision): no participation
metric exists anywhere in this system, so none appears in this file. Every
`stat` is a fact true by construction of the game — 20-minute clock, three
objectives, one verdict — or it is omitted.

USAGE
    python3 build_provocations.py                 # today, UTC
    python3 build_provocations.py --date 2026-07-30
    python3 build_provocations.py --date 2026-08-01 --check
"""
import argparse
import json
import sys
from datetime import date, datetime, timezone

ICONS = {
    "layers": '<path d="M12 2L2 7l10 5 10-5-10-5zm0 9L2 16l10 5 10-5-10-5z"/>',
    "bolt": '<path d="M13 2L3 14h7v8l10-12h-7l0-8z"/>',
    "clock": '<path d="M12 4a8 8 0 100 16 8 8 0 000-16zm0 3v5l4 2"/>',
    "chart": '<path d="M4 19h16M6 16l4-8 3 5 2-3 3 6"/>',
    "star": '<path d="M12 3l2.5 6H21l-5 4 2 7-6-4.5L6 20l2-7-5-4h6.5z"/>',
    "pulse": '<path d="M3 12h4l3-8 4 16 3-8h4"/>',
}
ICON_CYCLE = ["bolt", "clock", "chart", "star", "pulse", "layers"]

# ---------------------------------------------------------------------------
# THE SCHEDULE
# ---------------------------------------------------------------------------
# Ordered oldest-first. `publish` is the UTC date the entry becomes `today`.
# It stays `today` until a later entry's date arrives, then it is archived.
#
# `headline`/`body` are the today-shape (long form, `<em>` allowed in the
# headline). `title`/`teaser`/`detail_body` are the archive-shape.
#
# The eight entries dated 28 Jun – 5 Jul are HISTORICAL: they were written as
# archive entries before rotation existed, so they carry no authored
# today-shape. That absence is real, not manufactured — as_today() falls back
# for them, and the fallback is documented rather than hidden. Anything added
# from here on should author both shapes.
SCHEDULE = [
    {
        "publish": "2026-06-28",
        "slug": "group-chats-replaced-friendship",
        "title": "Group chats replaced friendship maintenance",
        "teaser": "Presence without effort. The bar has never been lower.",
        # No case written. Renders non-openable (Journey B.3) rather than
        # offering a dead control. The absence is real.
    },
    {
        "publish": "2026-06-29",
        "slug": "nobody-reads-terms-of-service",
        "title": "Nobody actually reads terms of service — and that's rational",
        "teaser": "The cost of reading outweighs the power to change anything.",
        "detail_body": (
            "Reading takes an hour. Understanding takes a lawyer. Refusing takes the service "
            "away. Given those three prices, not reading is the correct decision, and every "
            "study that frames it as apathy has mistaken a rational calculation for a character "
            "flaw.\n\n"
            "Which moves the burden somewhere else entirely. If consent is only ever given "
            "unread, then consent is not the thing doing the work, and we should stop pretending "
            "that it is."
        ),
    },
    {
        "publish": "2026-06-30",
        "slug": "four-day-week-productivity-myth",
        "title": "The four-day week is a productivity myth",
        "teaser": "The pilots that prove it were self-selected true believers.",
        "detail_body": (
            "The pilots recruit organisations that already believed, run them for six months "
            "with everyone watching, and measure self-reported output. That is a design which "
            "cannot return a negative result. It tells you what motivated people do under "
            "observation, not what a four-day week does.\n\n"
            "The counter is that the effect may well be real regardless, and demanding a hostile "
            "trial of something people obviously want is its own motivated reasoning. Possibly. "
            "Run it on a sceptical workforce for two years and the argument ends."
        ),
    },
    {
        "publish": "2026-07-01",
        "slug": "fiction-makes-you-worse-at-facts",
        "title": "Reading fiction makes you worse at facts",
        "teaser": "Narrative trains you to want a tidy arc. Reality doesn't have one.",
        "detail_body": (
            "A novel teaches you to expect that events connect, that behaviour has motive, and "
            "that the ending explains the beginning. None of that is true of a pandemic, an "
            "election or a market. The better you get at narrative, the more confidently you "
            "impose one.\n\n"
            "Against that: fiction is the main way most people practise holding a mind that is "
            "not their own, which is hardly nothing when the facts in dispute are about other "
            "people. Perhaps the trade is worth making. But it is a trade, and it is almost "
            "always sold as a free gain."
        ),
    },
    {
        "publish": "2026-07-02",
        "slug": "data-driven-decisions-arent",
        "title": "Most 'data-driven' decisions aren't",
        "teaser": "The numbers get picked after the gut already chose.",
        "detail_body": (
            "Watch the sequence. Someone forms a view, then commissions the analysis, then reads "
            "the analysis for the part that agrees. The dashboard is not an input to the "
            "decision. It is the receipt.\n\n"
            "The defence is that this still beats nothing — that even a motivated search for "
            "evidence occasionally turns up the number that stops you. Fair enough. But then say "
            "that is what the dashboard is for, and stop calling the output data-driven."
        ),
    },
    {
        "publish": "2026-07-03",
        "slug": "privacy-is-already-over",
        "title": "Privacy is already over",
        "teaser": "You traded it years ago. The fight now is who profits.",
        "detail_body": (
            "You cannot claw back a decade of location history, contact graphs and purchase "
            "records by changing a setting. The data exists, it has been copied, and the copies "
            "are the asset. Every privacy control shipped since governs what happens next, never "
            "what already happened.\n\n"
            "So the honest question stops being whether privacy survives and becomes who is "
            "permitted to profit from its absence. That is a distribution argument rather than a "
            "technical one, and it has an entirely different set of winners."
        ),
    },
    {
        "publish": "2026-07-04",
        "slug": "remote-work-killed-mentorship",
        "title": "Remote work killed mentorship",
        "teaser": "You can't absorb judgement over a video call.",
        "detail_body": (
            "Judgement is not transferred in meetings. It is absorbed in the two minutes after "
            "one — the aside, the raised eyebrow, the way someone rewrites your paragraph "
            "while you watch. None of those moments has an agenda item, so no scheduled call "
            "contains them.\n\n"
            "The rebuttal is that this was always a story senior people told about their own "
            "value. Plenty of people learned their craft alone, from documents, badly lit, and "
            "turned out fine. So which is it: a genuine transmission loss, or nostalgia for the "
            "office as a stage?"
        ),
    },
    {
        "publish": "2026-07-05",
        "slug": "ai-never-funny-on-purpose",
        "title": "AI will never be funny on purpose",
        "teaser": "The machine can recombine a million jokes and still not know why any land.",
        "detail_body": (
            "A model can hold every joke ever written and still not know which one to tell. "
            "Humour is a social risk instrument: it needs a target, a shared assumption to "
            "break, and a real chance of the room going cold. A system tuned never to offend "
            "and never to fail has removed all three ingredients before it starts.\n\n"
            "The counter-case is that funniness is only a pattern in the data, and the machine "
            "is a better pattern-finder than you are. If that holds, the failure is temporary "
            "and the punchlines improve. If it does not, then everything an AI has ever "
            "produced that made you laugh was written by a person it read."
        ),
    },
    {
        # The currently-live provocation. It has been `today` since 26 Jul
        # because nothing could replace it. It had no slug and no date, which
        # is exactly why it could never join the archive; both are supplied
        # here. No case is written for it yet — that absence is real too.
        "publish": "2026-07-26",
        "slug": "nobody-wants-personalised-internet",
        "title": "Nobody actually wants a personalised internet",
        "teaser": "What gets sold as personalisation is the quiet removal of what you'd have shared with a stranger.",
        "card_desc": (
            "What gets sold as personalisation is mostly the quiet removal of whatever "
            "you'd have had in common with a stranger."
        ),
        "headline": "Nobody actually <em>wants</em> a personalised internet.",
        "body": (
            "Every feed is tuned to one person, and every conversation now opens with "
            "“have you seen” and closes with a shrug. What gets sold as personalisation "
            "is mostly the quiet removal of whatever you would have had in common with a "
            "stranger. The engine is not serving you — it is dividing the room so each "
            "half can be sold separately."
        ),
    },
]

# Facts true by construction of the Gauntlet. Not participation metrics.
STATS = [
    {"value": "20:00", "label": "On the Clock"},
    {"value": "3", "label": "Objectives"},
    {"value": "1", "label": "AI Verdict"},
]


# ---------------------------------------------------------------------------
# Selection
# ---------------------------------------------------------------------------

def parse_schedule():
    out = []
    seen = set()
    for e in SCHEDULE:
        d = date.fromisoformat(e["publish"])
        if e["slug"] in seen:
            raise SystemExit(f"duplicate slug in schedule: {e['slug']}")
        seen.add(e["slug"])
        out.append({**e, "_date": d})
    ordered = sorted(out, key=lambda e: e["_date"])
    if [e["slug"] for e in ordered] != [e["slug"] for e in out]:
        raise SystemExit("SCHEDULE is not in ascending date order — fix the list, not the sort")
    return ordered


def select(schedule, on):
    """Return (today, archive_newest_first) for the given date.

    today = the latest entry whose publish date has arrived.
    archive = everything published strictly before today's entry.

    This IS the owner's rule: an entry is archived exactly when a later one is
    published, so it is never in both places at once and never in the archive
    on its own day.
    """
    due = [e for e in schedule if e["_date"] <= on]
    if not due:
        return None, []
    today = due[-1]
    archive = list(reversed(due[:-1]))
    return today, archive


def short_date(d):
    return f"{d.day} {d:%b}"


# ---------------------------------------------------------------------------
# Shapes
# ---------------------------------------------------------------------------

def as_today(e):
    """The `today` object. NOTE the two fallbacks: the eight historical entries
    were authored as archive rows and have no long-form today-shape, so the
    headline falls back to the title and the body to the case (or the teaser).
    Anything added from now on should author `headline` and `body` explicitly.
    """
    return {
        "eyebrow": "Today's Provocation",
        "date": short_date(e["_date"]),
        "slug": e["slug"],
        "headline": e.get("headline") or e["title"],
        "body": e.get("body") or e.get("detail_body") or e["teaser"],
        "primary_cta": {"label": "File Your Position", "url": "/tools/gauntlet/index.html"},
        "secondary_cta": {"label": "See All Provocations", "url": "/provocations/index.html"},
        "stats": list(STATS),
    }


def as_entry(e):
    """Archive row. `url`/`detail_body` only when a case exists — an entry
    without one renders non-openable rather than offering a dead control."""
    out = {
        "date": short_date(e["_date"]),
        "slug": e["slug"],
        "title": e["title"],
        "teaser": e["teaser"],
    }
    if e.get("detail_body"):
        out["detail_body"] = e["detail_body"]
        out["url"] = "/provocations/index.html?entry=" + e["slug"]
    return out


def build_arena(today, archive):
    """Card 0 is DERIVED from today — never a hand-copied duplicate."""
    cards = [{
        "icon": ICONS["layers"],
        "tag": "Today",
        "title": today["title"],
        # `card_desc` exists because the arena blurb and the archive teaser are
        # different slots with different lengths, and the live site has distinct
        # copy for each. Deriving both from `teaser` would have silently
        # reworded a live string — caught by diffing this builder's output
        # against the served file rather than by reading it.
        "desc": today.get("card_desc") or today["teaser"],
        "stat": "On the clock in the Gauntlet",
        "url": "/tools/gauntlet/index.html",
    }]
    for i, e in enumerate(archive[:5]):
        cards.append({
            "icon": ICONS[ICON_CYCLE[i % len(ICON_CYCLE)]],
            "tag": "Archive · " + short_date(e["_date"]),
            "title": e["title"],
            "desc": e["teaser"],
            "stat": "Read the case",
            "url": "/provocations/index.html?entry=" + e["slug"],
        })
    return {
        "eyebrow": "The Arena",
        "title": "Every provocation is <em>open</em> to argue.",
        "subtitle": (
            "Pick one, read the case for it, then take a position into the Gauntlet and "
            "defend it against an AI opponent on a twenty-minute clock."
        ),
        "cta_label": "Not sure where to start? Today's provocation is the one on the clock.",
        "cta": {"label": "See every provocation", "url": "/provocations/index.html"},
        "cards": cards,
    }


def build(on, generated_at):
    schedule = parse_schedule()
    today, archive = select(schedule, on)
    if today is None:
        raise SystemExit(f"no provocation is published on or before {on} — the schedule starts later")
    return {
        "generated_at": generated_at,
        "today": as_today(today),
        "arena": build_arena(today, archive),
        "archive": {"entries": [as_entry(e) for e in archive]},
    }


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--date", help="UTC date to build for (YYYY-MM-DD); defaults to today")
    ap.add_argument("--check", action="store_true",
                    help="print a one-line summary to stderr instead of the feed")
    args = ap.parse_args()

    on = date.fromisoformat(args.date) if args.date else datetime.now(timezone.utc).date()
    feed = build(on, datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"))

    if args.check:
        t = feed["today"]
        print(f"{on}  today={t['slug']} ({t['date']})  archive={len(feed['archive']['entries'])}"
              f"  arena_card0={feed['arena']['cards'][0]['title']!r}", file=sys.stderr)
        return
    print(json.dumps(feed, indent=2, ensure_ascii=False))


if __name__ == "__main__":
    main()

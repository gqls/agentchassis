#!/usr/bin/env python3
"""Author vonc.com/data/provocations.json for the P4 front-end rebuild.

Fabrication rail (gauntlet_dead_cta, 2026-07-26): no participation metric
exists anywhere in this system, so none appears in this file. Every `stat`
is either a fact true by construction of the game (20-minute clock, three
objectives, one verdict) or is omitted. The previous file carried invented
counts ("1,284 Positions Filed", "62% Disagree", "312 in the room") which
tools-api served verbatim to the Gauntlet.

New in this revision (approved EXPERIENCE_PLAN §3 data contract):
  archive.entries[].slug        URL-safe id used in ?entry=<slug>
  archive.entries[].detail_body full case text shown in the detail region
Entries without detail_body render non-openable (Journey B.3). Entry 8 has
no case written yet — that absence is real, not manufactured.
"""
import json

ICONS = {
    "layers": '<path d="M12 2L2 7l10 5 10-5-10-5zm0 9L2 16l10 5 10-5-10-5z"/>',
    "bolt": '<path d="M13 2L3 14h7v8l10-12h-7l0-8z"/>',
    "clock": '<path d="M12 4a8 8 0 100 16 8 8 0 000-16zm0 3v5l4 2"/>',
    "chart": '<path d="M4 19h16M6 16l4-8 3 5 2-3 3 6"/>',
    "star": '<path d="M12 3l2.5 6H21l-5 4 2 7-6-4.5L6 20l2-7-5-4h6.5z"/>',
    "pulse": '<path d="M3 12h4l3-8 4 16 3-8h4"/>',
}

TODAY = {
    "eyebrow": "Today's Provocation",
    "headline": "Nobody actually <em>wants</em> a personalised internet.",
    "body": (
        "Every feed is tuned to one person, and every conversation now opens with "
        "“have you seen” and closes with a shrug. What gets sold as personalisation "
        "is mostly the quiet removal of whatever you would have had in common with a "
        "stranger. The engine is not serving you — it is dividing the room so each "
        "half can be sold separately."
    ),
    "primary_cta": {"label": "File Your Position", "url": "/tools/gauntlet/index.html"},
    "secondary_cta": {"label": "See All Provocations", "url": "/provocations/index.html"},
    # Facts true by construction of the Gauntlet, not participation metrics.
    "stats": [
        {"value": "20:00", "label": "On the Clock"},
        {"value": "3", "label": "Objectives"},
        {"value": "1", "label": "AI Verdict"},
    ],
}

ENTRIES = [
    {
        "date": "5 Jul",
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
        "date": "4 Jul",
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
        "date": "3 Jul",
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
        "date": "2 Jul",
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
        "date": "1 Jul",
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
        "date": "30 Jun",
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
        "date": "29 Jun",
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
        # No case written for this one yet. The absence is real: it renders
        # non-openable (Journey B.3) rather than offering a dead control.
        "date": "28 Jun",
        "slug": "group-chats-replaced-friendship",
        "title": "Group chats replaced friendship maintenance",
        "teaser": "Presence without effort. The bar has never been lower.",
    },
]


def entry_out(e):
    """Feed shape: date/title/teaser/slug always; url + detail_body only when
    a case exists. No `stat` — the invented counts are gone and nothing real
    replaces them."""
    out = {
        "date": e["date"],
        "slug": e["slug"],
        "title": e["title"],
        "teaser": e["teaser"],
    }
    if e.get("detail_body"):
        out["detail_body"] = e["detail_body"]
        out["url"] = "/provocations/index.html?entry=" + e["slug"]
    return out


def card_for(e, icon, tag):
    return {
        "icon": icon,
        "tag": tag,
        "title": e["title"],
        "desc": e["teaser"],
        "stat": "Read the case",
        "url": "/provocations/index.html?entry=" + e["slug"],
    }


by_slug = {e["slug"]: e for e in ENTRIES}

ARENA = {
    "eyebrow": "The Arena",
    "title": "Every provocation is <em>open</em> to argue.",
    "subtitle": (
        "Pick one, read the case for it, then take a position into the Gauntlet and "
        "defend it against an AI opponent on a twenty-minute clock."
    ),
    "cta_label": "Not sure where to start? Today's provocation is the one on the clock.",
    "cta": {"label": "See every provocation", "url": "/provocations/index.html"},
    "cards": [
        {
            "icon": ICONS["layers"],
            "tag": "Today",
            "title": "Nobody actually wants a personalised internet",
            "desc": (
                "What gets sold as personalisation is mostly the quiet removal of whatever "
                "you'd have had in common with a stranger."
            ),
            "stat": "On the clock in the Gauntlet",
            "url": "/tools/gauntlet/index.html",
        },
        card_for(by_slug["ai-never-funny-on-purpose"], ICONS["bolt"], "Archive · 5 Jul"),
        card_for(by_slug["remote-work-killed-mentorship"], ICONS["clock"], "Archive · 4 Jul"),
        card_for(by_slug["privacy-is-already-over"], ICONS["chart"], "Archive · 3 Jul"),
        card_for(by_slug["data-driven-decisions-arent"], ICONS["star"], "Archive · 2 Jul"),
        card_for(by_slug["fiction-makes-you-worse-at-facts"], ICONS["pulse"], "Archive · 1 Jul"),
    ],
}

feed = {
    "generated_at": "2026-07-26T00:00:00Z",
    "today": TODAY,
    "arena": ARENA,
    "archive": {"entries": [entry_out(e) for e in ENTRIES]},
}

print(json.dumps(feed, indent=2, ensure_ascii=False))

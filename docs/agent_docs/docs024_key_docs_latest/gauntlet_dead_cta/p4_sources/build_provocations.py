#!/usr/bin/env python3
"""SUPERSEDED, AND THE SEAL DESIGN BELOW IS WRONG. DO NOT RUN. DO NOT PUBLISH.

Two separate reasons, both found 2026-07-31 after this file had already been
committed with the seal change. Left in place rather than deleted because
forward-only, and because the mistake is worth more than the file.

1. SUPERSEDED. The live feed is built by
   `provocation_pipeline/builder/build_provocations.py` and published by that
   lane's `publish_feed.sh` (their Phase 0 went live 16:07 on 2026-07-31, four
   minutes before I committed this). This file last drove the live feed on
   07-26. I edited it because `gauntlet_dead_cta/RUNBOOK` §8 documents this path
   — accurate when written, superseded since, and I did not check.

2. THE SEAL-BY-KEY-ABSENCE DESIGN IS UNSAFE AND WOULD HAVE BROKEN THE GAUNTLET.
   The guard below refuses to emit `today.headline`/`today.body`, on the premise
   that nothing needs them because the Gauntlet page does not fetch this feed.
   The PAGE does not. **The ENGINE does, server-side:**
   `internal/tools-api/handlers/round.go` `FetchProvocation()` fetches
   `https://{domain}/data/provocations.json`, takes the whole `today` object, and
   `RoundHandler` persists it as the round's provocation and returns it to the
   browser. Remove those keys and every round serves an empty provocation — the
   core feature, killed by a change meant to protect it.

   What saved it was checking the publish target against what the site actually
   serves BEFORE publishing: the repo file carried `today.date`/`today.slug` that
   this generator cannot produce, which is what exposed the newer pipeline, whose
   docstring names round.go. Verifying the target is the check; "I read the
   runbook" is not.

THE CORRECTED DESIGN lives in the provocation_pipeline builder: `today` keeps its
engine-facing shape untouched, and the seal is carried by NEW sibling keys the
engine never reads (`seal`, `sample`) plus a sealed `arena.cards[0]`. So the seal
is a RENDERER-level invariant, enforced by a checker, not by key absence — and it
should be described that way, because the provocation is necessarily readable at a
public URL that the engine has to be able to fetch.

Original docstring follows.

Author vonc.com/data/provocations.json for the P4 front-end rebuild.

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

THE SEAL (owner ruling 2026-07-31, HANDOFF_2026-07-30_C) — READ BEFORE EDITING
`today` MUST NOT carry `headline` or `body`. One invariant, site-wide:

    today's provocation is readable in the Gauntlet, after entry, and NOWHERE ELSE.

The Gauntlet page seals it on purpose (131-C) — you commit to arguing before you
know what you are arguing about. This file used to publish the same provocation in
full, and the home page and the Arena page both painted it, so every normal visitor
had read the argument before reaching the sealed door. Measured 2026-07-31: 3 of 19
pages leaked it.

**The seal is enforced HERE, in the data, not in the renderers.** Three pages read
this feed and any future one will too; a renderer-side fix would have to be repeated
for each and would silently not apply to the next. The keys are ABSENT rather than
empty — a renderer reading `today.headline` gets `undefined` and paints nothing.

So: today's provocation text does not exist in this file. `today` carries only the
SEAL (a statement that it is sealed) plus the route in. What a provocation reads like
is shown by `sample` — a PAST one, in full, which is safe because it has been argued.

Two things that will bite:
  * The Gauntlet page does NOT read this feed at all (verified 2026-07-31 by request
    interception). Its provocation comes from the engine's /round. So removing text
    from here cannot break the round.
  * `arena.cards[0]` is the "Today" card and is ALSO a leak surface — it carried
    today's title and a condensed body. It is now the sealed card. Both the home
    lobby grid and the Arena lobby read it from here, so fixing it here fixes both
    with no JS change.

Verify by RENDERING, never by grepping HTML — the text is written into an empty
shell by JS after load, so a curl grep reports "absent" on the pages that show it:
  ~/.venvs/vonc_pw/bin/python ../scripts/provocation_leak_sweep.py
"""
import json
import sys

ICONS = {
    "layers": '<path d="M12 2L2 7l10 5 10-5-10-5zm0 9L2 16l10 5 10-5-10-5z"/>',
    "bolt": '<path d="M13 2L3 14h7v8l10-12h-7l0-8z"/>',
    "clock": '<path d="M12 4a8 8 0 100 16 8 8 0 000-16zm0 3v5l4 2"/>',
    "chart": '<path d="M4 19h16M6 16l4-8 3 5 2-3 3 6"/>',
    "star": '<path d="M12 3l2.5 6H21l-5 4 2 7-6-4.5L6 20l2-7-5-4h6.5z"/>',
    "pulse": '<path d="M3 12h4l3-8 4 16 3-8h4"/>',
}

# NO `headline`, NO `body` — see THE SEAL in the module docstring. Adding either
# back re-opens the leak on three pages at once, and the guard at the bottom of this
# file will refuse to emit.
TODAY = {
    "eyebrow": "Today's Provocation",
    "sealed": True,
    "seal_headline": "Today's question is <em>sealed</em>.",
    "seal_body": (
        "You read it when the clock starts, and not before. That is the whole point: "
        "you commit to arguing before you know what you are arguing about, which is "
        "the one thing a chat window will never ask of you."
    ),
    "primary_cta": {"label": "Take On Today's Provocation", "url": "/tools/gauntlet/index.html"},
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


def newest_with_case():
    """The provocation shown in full as the worked sample.

    Deliberately DERIVED, not hardcoded: it is the first entry in ENTRIES that has a
    case written. ENTRIES is newest-first, so when a finished provocation joins the
    top of the archive the sample follows it with no edit here. That is as
    self-maintaining as this file can be until HANDOFF B makes today's provocation
    roll into the archive automatically — at which point this needs no change either.

    Safe by construction: a past provocation has already been argued, so showing it
    in full gives nothing away about today's.
    """
    for e in ENTRIES:
        if e.get("detail_body"):
            return e
    raise SystemExit("build_provocations: no archive entry has a detail_body, so "
                     "there is no safe sample to show. Refusing to emit a feed whose "
                     "home page would have nothing concrete on it.")


_sample = newest_with_case()

# The sample is a PAST provocation, shown in full, so a first-time visitor can see
# what one reads like without being told today's. Only the opening paragraph goes in
# `body` — the full case is one click away and the home card has room for one idea.
SAMPLE = {
    "eyebrow": "A past provocation",
    "date": _sample["date"],
    "slug": _sample["slug"],
    "headline": _sample["title"],
    "body": _sample["detail_body"].split("\n\n")[0],
    "cta_label": "Read the full case",
    "url": "/provocations/index.html?entry=" + _sample["slug"],
}

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
        # The "Today" card is a LEAK SURFACE — it used to carry today's title and a
        # condensed body, and both the home lobby grid and the Arena lobby render it
        # from here. It states that today's is sealed and routes into the round; it
        # never names it. `title` must stay non-empty: a card with no title is
        # filtered out by both renderers, which would silently remove the only route
        # into today's provocation from the lobby.
        {
            "icon": ICONS["layers"],
            "tag": "Today",
            "title": "Sealed until you step in",
            "desc": (
                "Today's provocation is revealed when the clock starts, not before."
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
    "sample": SAMPLE,
    "arena": ARENA,
    "archive": {"entries": [entry_out(e) for e in ENTRIES]},
}


def guard(feed):
    """Refuse to emit a feed that re-opens the leak.

    STRUCTURAL, not textual. A textual guard would need today's provocation written
    down here to compare against — which is the very thing that must not be in this
    file — and it would go stale the day the provocation changes. So this asserts the
    shape the seal depends on instead, which cannot go stale.

    This is a guard on the DATA. The check on the live SITE is a render sweep, because
    the text is painted by JS into an empty shell and no HTML-level check can see it:
        ~/.venvs/vonc_pw/bin/python ../scripts/provocation_leak_sweep.py
    """
    problems = []

    for key in ("headline", "body"):
        if key in feed["today"]:
            problems.append(
                "today.%s is present. That key is what leaked onto 3 pages; the seal "
                "is that it does not exist. Put the copy in today.seal_%s instead."
                % (key, key)
            )

    if not feed["today"].get("sealed"):
        problems.append("today.sealed is not true — renderers branch on it to decide "
                        "whether to paint the seal or a provocation.")

    # The sample must be a PAST provocation, i.e. one that appears in the archive.
    # A sample that is not in the archive is unverifiable: nobody can tell whether it
    # has been argued yet, which is the only reason showing it in full is safe.
    slugs = {e["slug"] for e in feed["archive"]["entries"]}
    if feed["sample"]["slug"] not in slugs:
        problems.append("sample.slug %r is not in the archive, so it cannot be shown "
                        "as a past provocation." % feed["sample"]["slug"])

    # The Today lobby card must route somewhere and must not name a provocation.
    today_card = feed["arena"]["cards"][0]
    if today_card.get("tag") != "Today":
        problems.append("arena.cards[0] is no longer the Today card; the guard below "
                        "is checking the wrong card.")
    if not today_card.get("title") or not today_card.get("url"):
        problems.append("arena.cards[0] needs a title AND a url — both renderers "
                        "filter out cards missing either, which would remove the only "
                        "lobby route into today's round.")
    archive_titles = {e["title"] for e in feed["archive"]["entries"]}
    if today_card.get("title") in archive_titles:
        problems.append("arena.cards[0].title names a real provocation. The Today card "
                        "must state that today's is sealed, not what it is.")

    if problems:
        sys.stderr.write("build_provocations: REFUSING TO EMIT — the seal would be "
                         "broken:\n")
        for p in problems:
            sys.stderr.write("  * %s\n" % p)
        raise SystemExit(1)


guard(feed)
print(json.dumps(feed, indent=2, ensure_ascii=False))

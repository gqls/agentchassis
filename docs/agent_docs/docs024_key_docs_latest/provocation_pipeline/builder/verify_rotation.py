#!/usr/bin/env python3
"""Assert the ROTATION MECHANISM, not one day's output.

Handoff B's rule, and it is the right one: rotation cannot be verified by
looking at the page, because a hardcoded provocation and a rotated one are
indistinguishable on any single day. That is how this survived a month. So
this runs the real builder across every date in the schedule's span and
asserts the invariants that must hold on all of them.

    python3 verify_rotation.py

Exits non-zero on the first violated invariant.
"""
import sys
from datetime import date, timedelta

from build_provocations import build, parse_schedule

FAILURES = []


def check(cond, label, detail=""):
    if not cond:
        FAILURES.append(f"{label}{(' — ' + detail) if detail else ''}")
    return cond


def main():
    schedule = parse_schedule()
    first = schedule[0]["_date"]
    last = schedule[-1]["_date"]
    span_end = last + timedelta(days=10)  # past the end, to prove it holds there too

    seen_today = []
    prev_today, prev_archive_len = None, None
    days = 0

    d = first
    while d <= span_end:
        feed = build(d, "2026-01-01T00:00:00Z")
        today = feed["today"]
        archive = feed["archive"]["entries"]
        slugs = [e["slug"] for e in archive]
        days += 1

        # 1. today can always BECOME an archive entry. Missing either field is
        #    what stranded the archive on 5 Jul.
        #    This runs FIRST because every check below dereferences the slug —
        #    when it was ordered second, deleting the slug made the verifier
        #    crash with a KeyError instead of reporting. A checker that dies on
        #    the defect it is looking for still exits non-zero, so it passes a
        #    mutation test while telling the reader nothing.
        if not check(bool(today.get("slug")), f"{d}: today has no slug"):
            d += timedelta(days=1)
            continue
        check(bool(today.get("date")), f"{d}: today has no date")

        # 2. A provocation is NEVER today and archived at the same time. This is
        #    the owner's rule ("archived when the new one is published") stated
        #    as an invariant instead of a hope.
        check(today["slug"] not in slugs,
              f"{d}: today is also in the archive", today["slug"])

        # 2b. ...and the date must be the RIGHT one, not merely present.
        #     Added after mutation-testing: freezing today["date"] to a literal
        #     passed every other invariant here. That date is what gets carried
        #     into the archive on promotion, so a frozen one would date every
        #     archived entry identically while looking entirely plausible —
        #     the same shape as the hardcoded `generated_at` in LANDMINES.md.
        #     "Field is populated" is not "field is correct".
        want_date = _short(_date_of(schedule, today["slug"]))
        check(today.get("date") == want_date,
              f"{d}: today.date is {today.get('date')!r}, should be {want_date!r}")

        # 2c. Promotion must carry each entry's own date, for the same reason.
        for e in archive:
            expect = _short(_date_of(schedule, e["slug"]))
            check(e.get("date") == expect,
                  f"{d}: archived {e['slug']} dated {e.get('date')!r}, should be {expect!r}")

        # 3. The arena's first card must NOT name today's provocation.
        #    INVERTED 2026-07-31 (owner ruling on gauntlet HANDOFF C). This used to
        #    assert card 0 MATCHED today, which was right for rotation and wrong for
        #    the seal: both the home lobby grid and the Arena lobby render this card,
        #    so a derived card was two of the three surfaces publishing the question
        #    the Gauntlet is built to hide.
        #
        #    Still an invariant, just the other way up — and it has to check BOTH the
        #    title and the blurb, because the leak was the blurb on one of the two
        #    surfaces. It also still has to be a usable card: both renderers drop a
        #    card with no title or no url, which would silently remove the only lobby
        #    route into today's round.
        card0 = feed["arena"]["cards"][0]
        today_title = _title_of(schedule, today["slug"])
        check(card0["title"] != today_title,
              f"{d}: arena card 0 names today's provocation", card0["title"])
        check(today["headline"] not in card0.get("desc", "")
              and (today.get("body") or "")[:40] not in card0.get("desc", ""),
              f"{d}: arena card 0's blurb carries today's text", card0.get("desc"))
        check(bool(card0.get("title")) and bool(card0.get("url")),
              f"{d}: arena card 0 is not a usable card", f"{card0.get('title')!r} -> {card0.get('url')!r}")

        # 3b. The engine's contract. round.go's FetchProvocation takes the whole
        #     `today` object server-side and RoundHandler persists it as the round's
        #     provocation, so these keys are load-bearing for the GAME, not for the
        #     display. Asserted here because the seal above creates a standing
        #     temptation to "seal" today by emptying them, which would serve every
        #     round a blank question. Sealing is a display concern; this is not.
        for _k in ("headline", "body", "slug", "date"):
            check(bool(today.get(_k)),
                  f"{d}: today.{_k} is empty — that breaks the round, it does not seal it")

        # 3c. The display surfaces must have something real to show instead.
        #     A blank home card passes every "today's provocation is absent" test
        #     while being a worse page than the leak was.
        seal = feed.get("seal") or {}
        check(bool(seal.get("headline")) and bool(seal.get("body")),
              f"{d}: feed carries no seal copy, so the display surfaces have nothing "
              f"to say in place of today's provocation")
        sample = feed.get("sample")
        if sample:
            check(sample["slug"] != today["slug"],
                  f"{d}: sample is today's provocation, not a past one", sample["slug"])
            check(sample["slug"] in slugs,
                  f"{d}: sample {sample['slug']!r} is not in the archive, so it has "
                  f"not necessarily been argued yet")

        # 4. Archive is newest-first and contains exactly the entries published
        #    strictly before today's.
        expected = [e["slug"] for e in reversed([e for e in schedule if e["_date"] < _date_of(schedule, today["slug"])])]
        check(slugs == expected, f"{d}: archive contents/order wrong",
              f"got {slugs} want {expected}")

        # 5. The archive only ever grows, and only when today changes.
        if prev_today is not None:
            if today["slug"] == prev_today:
                check(len(archive) == prev_archive_len,
                      f"{d}: archive changed while today did not")
            else:
                check(len(archive) == prev_archive_len + 1,
                      f"{d}: today changed but archive grew by {len(archive) - prev_archive_len}, not 1")
        prev_today, prev_archive_len = today["slug"], len(archive)

        if not seen_today or seen_today[-1] != today["slug"]:
            seen_today.append(today["slug"])
        d += timedelta(days=1)

    # 6. The whole point: the same code on different dates yields different
    #    provocations. One distinct value would mean nothing rotates.
    check(len(seen_today) == len(schedule),
          f"expected {len(schedule)} distinct provocations across the span, saw {len(seen_today)}")

    # 7. Every scheduled provocation gets its turn, in order.
    check(seen_today == [e["slug"] for e in schedule],
          "the sequence of today-values does not match the schedule order")

    # 8. Before the schedule starts, the builder must FAIL rather than invent one.
    try:
        build(first - timedelta(days=1), "2026-01-01T00:00:00Z")
        FAILURES.append("built a feed for a date before the schedule starts (should fail loud)")
    except SystemExit:
        pass

    print(f"dates checked : {days}  ({first} .. {span_end})")
    print(f"distinct today: {len(seen_today)} of {len(schedule)} scheduled")
    if FAILURES:
        print(f"\nFAILED ({len(FAILURES)}):")
        for f in FAILURES:
            print("  -", f)
        return 1
    print("\nall rotation invariants hold")
    return 0


def _date_of(schedule, slug):
    return next(e["_date"] for e in schedule if e["slug"] == slug)


def _short(d):
    # Deliberately re-derived here rather than imported from the builder: a
    # verifier that formats dates with the code under test cannot detect a
    # formatting change, it can only agree with it.
    return f"{d.day} {d:%b}"


def _title_of(schedule, slug):
    return next(e["title"] for e in schedule if e["slug"] == slug)


if __name__ == "__main__":
    sys.exit(main())

# 232 — Published Gauntlet rounds are search-engine indexable: nothing sets `noindex`

**Filed 2026-08-09 by the `provocation_pipeline` lane. Owner: `gauntlet_dead_cta`**
(they own `internal/tools-api` and the round-record page).

**Severity: low today, high the moment a stranger publishes a round about a named
person.** Filed separately from `architecture_review/RFC_020` because it is true
*now*, is a two-line fix, and does not depend on any of RFC_020's open questions.

## Symptom

`vonc.com/tools/gauntlet/round.html?r=<slug>` serves a visitor's own prose and an
AI verdict at a permanent public URL, and **nothing instructs search engines not to
index it**. A published round is therefore discoverable by searching its text —
including, if a visitor writes about a real person, by searching that person's name.

## Evidence [MEASURED 2026-08-09]

```
grep -rniE "noindex|robots" internal/tools-api/          → 0 hits
grep -rniE "noindex|robots" .../gauntlet_dead_cta/round_record/  → 0 hits
```

Both excluding tests. **First-hand verification, stated per the owner ruling of
2026-07-31:** this is an absence check over the two locations that could carry the
header or the meta tag, not a structural root-cause claim, so it is
self-evidencing and no `090` run was made. If the tag is set somewhere neither grep
covers — a Caddy header, a CDN rule, the page's stored `html_template` — **that
would refute this file and should be recorded here**; the check is
`curl -sI https://vonc.com/tools/gauntlet/round.html | grep -i x-robots` plus a
`curl -s … | grep -i noindex` on the served body.

## Why it matters more than it looks

Discoverability is the largest single multiplier on harm from user-generated
content. Something reachable only by a link you were handed is a contained problem;
the same words returned by a name search are not. This is the cheapest available
reduction in that exposure and it costs **nothing** in reach — a shared link works
exactly as before, which is the point: it removes the harm multiplier without
touching the viral mechanism the owner wants to keep.

## Fix candidates, ordered by what closes the door

1. **`X-Robots-Tag: noindex` response header on the published-round route**
   (`GET /round/:slug` and whatever serves `round.html`). Best: it cannot be lost by
   a page re-render, and it covers the JSON endpoint as well as the page.
2. `<meta name="robots" content="noindex">` in the round-record component's
   `html_template`. Works, but lives in stored component content, which this estate
   has repeatedly found can be regenerated away — see the chrome/rerender family.
3. `robots.txt` disallow. **Weakest — do not rely on it alone**: it asks crawlers
   not to fetch, does not prevent indexing of a URL discovered elsewhere, and is
   itself a public list of the paths you did not want looked at.

Recommend **1**, optionally with 2 as belt-and-braces.

## How to verify the fix

At the artefact, not the tag or the commit:

```sh
curl -sI 'https://vonc.com/tools/gauntlet/round.html?r=<a published slug>' | grep -i x-robots
curl -sI 'https://tools.apis.uk/api/v1/tools/gauntlet/round/<slug>' -H 'Origin: https://vonc.com' | grep -i x-robots
```

Both should carry `noindex`. **Check a real published slug, not a 404** — a missing
route returns no header either, which reads identically to a fix that is not there.

## Related

- `architecture_review/RFC_020_third_party_harm_in_the_gauntlet_before_and_after_publish.md`
  — the wider question this was found inside. RFC_020 §5.1 recommends this fix be
  made **independently of** its own open questions.
- `bugs_open/139` — poster identity is a constant, so published rounds are
  effectively anonymous. Relevant because anonymity plus indexability is the
  combination that makes a takedown request the only remedy.

# 399 — the framework computes each CTA's real destination, stores its title BESIDE the copy that misdescribes it, hands it to the writer, and never compares the two

Filed 2026-08-25 by the `dartsonline_traffic` lane, from an owner report of four separate
defects on dartsonline.com. He asked for the general fix rather than four individual ones, and
there is one: **the fact needed to prevent this class is already computed, already stored, and
already adjacent to the wrong value — on 646 live components as of 2026-08-25 — and nothing
compares them.**

## The evidence, in one row

`news-index`, `hero` slot, `page_components.content_data` `[MEASURED 2026-08-25]`:

```json
{
  "cta_text":                   "Catch up on this week's darts news",
  "cta_url":                    "/brands/index.html",
  "cta_target_title":           "All Brands | Darts Online",
  "secondary_cta":              "See what's new for beginners",
  "secondary_cta_url":          "/tools/setup-builder/index.html",
  "secondary_cta_target_title": "Dart Setup Builder | Tools"
}
```

The copy says *news*; the recorded destination title says *All Brands*. The copy says
*beginners*; the recorded destination says *Dart Setup Builder*. **Both contradictions are
inside one jsonb object, in adjacent keys, at build time** — no rendering, no HTTP, no HTML
parse, no discovery sweep required to see them.

And the field exists for exactly this purpose. `setCTAField`
(`resolve_internal_links_action.go:335-339`), verbatim:

> *"Alongside the url it writes the target's title under `<field-minus-_url>_target_title`
> (`cta_url` → `cta_target_title`) **so the content writer can write CTA copy FOR the actual
> destination instead of guessing one.**"*

So the design intent is destination-first copy. **What is missing is any check that the writer
complied.** `grep`ped 2026-08-25: `*_target_title` has writers
(`resolve_internal_links_action.go`, `links_tel.go`) and **no reader that compares it to the
label** — the only two references in the checks are `SuggestedTargetTitle` fields the checks
*emit* in their own findings.

## Where enforcement actually happens, and why it loses

Instead of comparing two adjacent strings at write time, the estate detects this by:
render the page → deploy it → sweep it later → parse `rendered_html` → extract anchors →
reduce the text to distinctive tokens → token-match against every page's name/title/nav_label →
file a `page_rerender` → recompute the url. Five lossy stages to rediscover a fact that was free.

That pipeline then loses in five measured ways:

1. **It is enabled on ONE of three discovery agents** `[MEASURED 2026-08-25]`:
   `completeness-discovery-agent` carries `misdirected_cta`; `quality-discovery-agent` and
   `design-discovery-agent` carry neither it nor `empty_sections`.
2. **The defect is minted faster than the sweep runs.** `news-index` re-renders with
   `reason=section_data_resolved` several times a day ("fresh news items available" — 6+ times in
   the last 3 days), and each regeneration re-mints CTA copy. A periodic sweep cannot converge on
   a page that re-authors itself between sweeps.
3. **Generic labels are documented to fall through to an arbitrary pick.** Same file, line 349:
   *"A generic label ('Get Started') or one matching no candidate falls through to today's
   positional behaviour unchanged."* And `check_misdirected_cta`'s own header: generic texts
   *"match nothing and are skipped entirely."* So a generic CTA gets a positional destination and
   is then unfalsifiable — **correct by construction, wrong in fact.**
4. **Site chrome is outside the surface entirely.** `check_misdirected_cta` reads
   `page_components.rendered_html`. The nav button lives in `site_components`, so
   `Get Started → /contact.html` can never be seen by it. Confirmed at the artefact:
   `<a href="/contact.html" class="header-cta">Get Started</a>`, and the header's `content_data`
   carries no cta/url/target_title keys at all — the nav CTA is not part of the CTA subsystem.
5. **The one arm that does fire has no handler.** `cta_names_unknown_destination` goes to
   `needs_human_review` *"with no handler"* by design. Live example on this site: *CTA "Read the
   shaft length guide" on shaft-length (hero): links back to its own page*. Filed 2026-08-24,
   parked, nobody told.

## The same session found the identical shape elsewhere, which is why this is a class

- **`empty_section` on `brands-index`** — the brands directory the owner reported as absent
  (450 chars of visible text, zero brands listed). Detected **2026-08-17, 08-22 and 08-24**, and
  **`failed` all three times**: *"completion blocked: verification could not run, and this item
  type fails closed (RFC_017)"*. Terminal, silent, eight days.
- **`bugs_open/384`** — a card image lands, is linked correctly, and the listing that renders it
  is re-rendered three times in a mode that structurally cannot pick it up.
- **`image_url_404`** — `/assets/images/hero.jpg` referenced by 5 pages, filed 2026-08-09 with
  **no `handler_agent`**, still `detected` 15 days later.

**The unifying property is not "checks are missing". It is that a defect can reach four different
terminal-looking states — `failed`, `needs_human_review` with no handler, `detected` that never
drains, and `complete` on an item whose defect persists — and every one of them looks handled
from outside.** Nothing reports "this site has N unresolved defects".

## Fix candidates, ordered by what closes the door

1. **Compare the label to its `_target_title` at WRITE time.** After the content writer returns,
   for every `<x>` with a sibling `<x>_target_title`: if the label's distinctive tokens match no
   token of the recorded title, either regenerate the label from the title (the field exists for
   that) or refuse the write. **This makes the bad state unrepresentable rather than detectable**,
   costs one string comparison, needs no rendering, no sweep and no per-agent enablement, and
   applies to all 646 components carrying the field today. The token reducer already exists in
   `check_misdirected_cta` and would be shared.
2. **Exclude the current page from the destination picker.** Kills the self-advertising CTA
   ("Catch up on this week's darts news" *on the news page*) as a category, and stops the
   `cta_names_unknown_destination` self-link arm being reachable at all.
3. **Refuse to mint a positional destination for a generic label.** Either the writer names a
   destination or the CTA is not emitted. Today "Get Started" pointing anywhere is *correct*
   behaviour, which is precisely why nothing flags it. **Fleet control measured: 0 of 1,515 live
   CTA labels are generic**, so this arm is rare in page components and cheap to tighten — the
   nav is the real consumer and it is outside the subsystem (candidate 4).
4. **Bring site chrome into the CTA subsystem**, or state in the check's header that chrome is out
   of scope so the next reader does not assume coverage it does not have.
5. **Report the terminal-but-unfixed states per site.** A count of `failed` + `needs_human_review`
   with no handler + `detected` older than N days is the missing instrument. Every defect in this
   file was *detected*; none was *surfaced*.

Candidate 1 is the one that closes the door; 5 is the one that would have told us eight days ago.

## Verification

Pick a page with a mismatched pair, run the build, assert the stored `cta_text` shares a
distinctive token with `cta_target_title`. **The disconfirming case must be exercised**: seed a
deliberately mismatched label and require the build to refuse or rewrite it — a check that only
ever sees matching pairs proves nothing, which is this estate's most-repeated lesson.

## 090 substitution, stated

Not run through the diagnosis loop: every step is first-hand and re-runnable — the live
`content_data` row above, the two source comments quoted verbatim, the `grep` showing no
comparing reader, the three `agent_definitions` rows showing enablement, the `site_work_items`
history for `empty_section`, and the served header anchor. A 090 run would re-read the same six
artefacts.

# CONTRIB into the `oufe` lane — 2026-08-25, from `bugfix_283_component_instance_scope`

**One locked row of yours never received a change it should have, and nothing anywhere told you.**
No action is urgent; nothing is broken for a visitor today. Filing it here because there was no
live `oufe` session when it was found, and the only other record is inside an RFC nobody working
this site would read.

## The row

`oufe.com/cases/thames-water.html`, slot **`evidence-timeseries-leakage`**, `locked_by =
'oufe-workstream'`.

## What happened, in order — the order is the whole point

| when (UTC) | what |
|---|---|
| 2026-07-29 08:28 | your row's `rendered_html` last written. **This is still its content today.** |
| 2026-08-23 **12:32:25** | a whole-page `page_rerender` completed on this page |
| 2026-08-23 **12:33:33** | the `evidence-timeseries` **template** was converted to per-instance element ids (`{{.InstanceID}}`) |
| since | **nothing.** No `section_edit`, no further rerender |

The rerender ran **68 seconds before** the template changed, so it carried the *old* template. It
looks like delivery in any log or count that only records that a rerender happened.

## Consequence

The page still serves the pre-conversion literal id:

```bash
curl -s https://oufe.com/cases/thames-water.html | grep -o 'id="[^"]*evidence-timeseries[^"]*"'
#   id="evidence-timeseries-leakage"      (expected after conversion: id="c-evidence-timeseries")
```

**This is harmless as it stands.** A literal id is only a defect when the same component appears
twice on one page, and it appears once here. What it costs you is that this row is now
**inconsistent with the other two placements of the same component** once they convert, so
anything written against `#c-evidence-timeseries` will not match on your page.

## Why you were never told, and it is not your fault or ours

The two other `evidence-timeseries` placements are locked too, and delivery *was* attempted on
them — the lock gate refused and filed `lock_blocked_change`, so they are visible. Yours received
**no attempt**, so no gate fired and no row exists. **A locked instance that gets no delivery
attempt produces no signal at all, and is indistinguishable from one that needed no delivery.**
Our coverage count was keyed on the refusals, so it read your row as fine. Full write-up:
`LANDMINES.md`, *"A locked instance that gets NO delivery attempt files NO refusal"*, and
`architecture_review/RFC_032…md` §9a (corrected 2026-08-25).

Found by the **`news_editorial_features`** lane, who own the other two placements and noticed the
asymmetry while accepting the change on theirs. They have deliberately not touched your row.

## What we suggest, and it is your call

Their dry-run on the two sibling rows is the useful precedent: they reproduced each stored row
**byte-for-byte** from the v1 snapshot plus live `content_data`, then rendered the live template
with `InstanceID` bound and confirmed **only the id attribute changed** (−2 and −11 bytes). So the
conversion is clean on this component — the risk in re-delivering is the lock, not the content.

1. Unlock, re-dispatch the `section_edit` delivery, verify at the artefact, re-lock. That is
   exactly what they are doing today on theirs.
2. Or leave it. It serves correctly and will keep doing so. Just know it is the odd one out.

⚠ **If you re-deliver, verify at the served page, not at the work item.** A `complete` status on
this pipeline is not proof — measured on this very programme 2026-08-24, three repairs completed
with correct stored bytes while one page still served the old version for hours.

⚠ **And do not use a plain `page_rerender` with no `reason`.** `page-rerender` routes on an
allow-list of five reasons; anything else (including none) goes to assemble-only, which re-ships
the stored bytes and **completes successfully having changed nothing** — which is arguably how
this row got missed in the first place. `reason: template_changed` is the one to use.

## Also, unrelated, while looking at your page

Three `lock_blocked_change` rows on this page sit at `needs_human_review` from **2026-07-29** —
nearly a month. Not ours and not necessarily wrong, but nobody appears to have read them.

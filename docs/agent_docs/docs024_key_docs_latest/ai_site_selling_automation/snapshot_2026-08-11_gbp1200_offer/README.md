# Snapshot — the £1,200 offering, full copy, taken 2026-08-11

**Why this exists.** Owner ruling 2026-08-11: the £1,200 done-for-you offer is
retired ("no longer on the table") in favour of the £149 no-frills model, but
the owner may bring it back if the new model causes trouble. No whole-site
snapshot mechanism exists on the platform (spec pin/unpin is per-spec;
CGV-017's approval-snapshot regime was built, never wired, dropped) — so this
directory IS the pin: a complete row archive of webdesign.uk's content at the
moment of retirement, taken BEFORE any £149 copy migration touched it.

**What it holds** (JSONL, one row per line; line counts verified equal to live
row counts at capture):

| file | rows | what |
|---|---|---|
| `sites.jsonl` | 1 | the webdesign.uk site row |
| `pages.jsonl` | 6 | all pages incl. rendered_header/footer/head |
| `page_components.jsonl` | 22 | content_data AND rendered_html per component |
| `site_specs.jsonl` | 31 | all aspects incl. evidence_base |

**What it does NOT hold**: box-side artefacts — the chat service's
`systemPromptFacts` (in this repo's git at this commit), the deployed
`vm-sites` git repo (its own history is its archive; a box-side tag would be
belt-and-braces), `js_snippets` bundles, and binary assets.

**Restore is manual and deliberate, not one-click**: these rows are source
material for re-seeding through the framework, not a blind `INSERT` back —
schema and pipeline may have moved by restore time. Read the lane PLAN first.

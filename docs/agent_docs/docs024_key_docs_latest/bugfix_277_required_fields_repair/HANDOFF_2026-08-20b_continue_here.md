# HANDOFF — 2026-08-20b. The owner ruled YES on the 7, the route is BUILT (CQ-028), config is LIVE, and what remains is a ROLL + ONE CANARY

**Supersedes `HANDOFF_2026-08-20_continue_here.md`** — read that for the morning's state (497/498/499,
the tick proof, the ownership retraction); read THIS for what changed after the owner answered its §7
question. NOTES has one new entry (~17:30Z); the PLAN carries the design addendum with every "why".

> **Written ~18:00Z; verdict box updated ~18:45Z.** Chassis was **`v1.0.1320`** before this session's
> commit; the commit (`af0f00bb5`) is on the shared branch and ships with whatever fleet roll happens
> next. Migration `513` is APPLIED (section-editor config md5 **`b6076c7d…`** — re-read, do not quote).
>
> **COUNCIL: APPROVED, round 2** (corr `b72a4029-…`, report 18:26:44Z, "3 advisory objections — none
> high-severity"). Round 1 was REVISE on a real sketch defect (backtick-quoted regex hid its own
> anchors); the r1/r2 responses are commits `25d00cfe9` and `6011f9657` (the latter carries
> `Council-Reviewed:` — verdict read). NOTES ~18:20Z and ~18:35Z hold the round-by-round record,
> including the advisories acted on (nil-persist pin test, mutation-proven; CQ-028 rescoped to
> "first markup-INSERTION repair"). **Nothing council-related is owed any more.**

---

## 0. STATE TABLE

| bug | state | what blocks the close |
|---|---|---|
| **`bugs_open/277`** | clause 1's route BUILT (owner ruling 2026-08-20, in chat: *"…a transform that edits finished HTML directly — I think yes"*) | a chassis roll, then **ONE canary** (§2), then the worked-example proof at the served bytes. `no_content_data` half still untouched by all of this |
| **`bugs_open/083`** | unchanged from the morning handoff | door soak ~08-25; its canary machinery is now also THIS lane's bootstrap |
| `bugs_open/333` | unchanged (theirs) | NB this route, once live, ENDS the wont_fix loop for the 7 — their n=7 measured instance stops reproducing. Tell them if you touch it |

## 1. WHAT WAS BUILT (register **CQ-028** — the entry is the map; PLAN addendum has the reasons)

Commit **`af0f00bb5`** (19 files): `datahelpers.ConvertLiteralCodeSpansInHTML` (`` `x` `` →
`<code>x</code>`, tokenizer byte-splice, detector's own skip map, conversion ⊆ detection,
mutation-proven both ways on REAL production bytes) · `apply_section_edit` edit_type
`rendered_html_transform` (opt-in `allow_rendered_html_transform`, floors + regulated guard wired,
HTML-only persist — content_data structurally untouchable) · `datahelpers.ContentDataCanFillTemplate`
(mig 499's test as code) · `check_literal_markdown.transformRouteSlot` (routes to section-editor IFF
all findings rendered_html-source code_span on ONE once-occurring non-regenerable slot; 11 refusal
directions each land on today's route) · migration `513` + ROLLBACK (round-tripped to the byte:
pre-image `fdb8cb4d…` restored exactly, then applied for real).

**Also live now:** the optional-key-budget overlay was re-applied — the cluster's `check.py` carries
`"apply_section_edit": 7` (verified at the mounted configmap, not the apply output). A new
`LANDMINES.md` entry (x/net/html `TagName()` mutates the buffer `Raw()` aliases) is appended, synced
and verifier-armed.

## 2. THE SEQUENCE THAT REMAINS — in order, nothing parallel

1. **A fleet roll** ships the code (detector routing + action branch are ONE image — no half-deploy
   is possible, stated in 513's header). Do not roll a single service for this; owner runs releases.
2. **The detector's next sweep** over webdesign.co.uk files fresh items at the new shape. All 7 old
   rows are terminal and the dedup index excludes their statuses — checked. If an OLD-shape open row
   exists at roll time it must cycle to terminal first (the wont_fix loop does this unaided — slower,
   not stuck).
3. **ONE CANARY, by hand** — the RUNBOOK's new section has the exact SQL (it is 444's own promote
   UPDATE) and the served-bytes proof commands. The pair `literal_markdown → section-editor` has 0
   lifetime completes, so the promoter HOLDS it until this one completion; then the other six flow
   unaided. **Do not promote a second row to "help".**
4. **Prove at the artefact**: `curl` the page — `<code>ease-in-out</code>` present, the prose
   backticks gone, and the tool's own `<script>` template literals UNTOUCHED (the fixture's
   adversarial case, now in production's favour).

## 3. TRAPS SET OR FOUND THIS SESSION

- **`Tokenizer.TagName()` lower-cases the bytes `Raw()` aliases, in place** — LANDMINES has the
  entry; the check is *write Raw first, and keep a MIXED-CASE tag in any splice test corpus*.
- **`input_fields` is a WHITELIST** (`action_inputs.go:831`): an optional input absent from the
  step's list is silently never extracted. 513's second UPDATE is load-bearing plumbing, not tidiness.
- **`converted==0` REFUSES rather than completing** — deliberate: a span detection can see but the
  conversion can't reach (crossing inline tags, entity-encoded backtick) must fail to a human, never
  deploy identical bytes as an "edit". Expect occasional legitimate `failed` items of that shape.
- The transform's durability rests on the SAME property that made these pages unrepairable: nothing
  regenerates them. If someone later BACKFILLS `content_data.body` on ported pages (a fix this lane
  considered and did not build), the regenerate routes wake up and could reprint pre-transform
  content — whoever does that owns re-checking these 7.

## 4. STILL OWED

- ~~READ the council verdict~~ **DONE — APPROVED r2, acted on, see the banner box.**
- Everything in the morning handoff's §6 (diagnosis_guardian message · 083 close ~08-25 ·
  `copy_edit_proposed` exclusion, owner-gated · 277's `no_content_data` half).
- After the canary: update `bugs_open/277` §6 with the worked example, then weigh the close.

## 5. Session-start checklist
`git log --oneline -10` · re-read this from disk ·
`distinct digests` on the chassis + `git merge-base --is-ancestor af0f00bb5 <stamp>` to know whether
step 1 has happened · then §2 from wherever the sequence stands.

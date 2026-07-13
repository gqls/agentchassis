# Handoff — building idea.uk's front site through the chassis (2026-06-16)

The one-page picture of the *site-build* stream. This is **separate** from the live £29 tool, which is
already live and earning and is handed off in `HANDOFF.md` + `RUNBOOK_idea_uk.md`. Full detail for this
stream: `RUNBOOK_idea_uk_chassis_site_and_vm_deploy.md`. Cross-session journal: `running_notes.md` (read
the tail). Memory is off, so those files are the record.

## What this stream is

Rebuild idea.uk's **front site** using the chassis website-builder (the same pipeline that built
gamesdesign.co.uk), then serve it from the **VM**, not Backblaze — while keeping the live tool (idea
generator + free audience check) on the same box. The bigger aim the owner set: idea.uk should **lead its
field with the best content for users**, not mirror the site that's there now.

## Where it stands

- **The live tool is untouched and earning.** Different stream; don't disturb `/stripe/webhook` or the
  operator paths. idea.uk's DNS is in **Cloudflare → the VM**, so a static build pushed to B2 is invisible
  at https://idea.uk and doubles as safe staging.
- **The chassis has already run over idea.uk** (site_id `97ed2f64-65ca-4b67-8a98-dfd8195a0d3a`). The
  classifier wrote `identity` / `classification` / `content_direction` / `design_intent`; the strategist
  wrote `strategy`; `briefing` was written; and **`site-design-planner` has resolved and installed a
  composition** (`resolved_composition`, 2026-06-16). So a build is in motion / pages may already be
  building. There are **duplicate spec rows** (classifier ran more than once) to tidy.
- **The read was faithful but conservative.** Because the live site was up, the classifier described
  today's idea.uk and the positioning was validated rather than pushed forward. That is the thing we are
  correcting before the build settles.

## The two decisions reached

1. **Set the direction before the build.** The classifier is a current-state tool; with no aspiration fed
   in and a live site to read, it can only describe what's there. Aspiration has to come from elsewhere.
2. **Content fix = a standing-ambition default in the mission, not a per-site hand-edit.** There is no
   generated aspirational aspect (`SELECT DISTINCT aspect` — no `vision`/`ambition`; `aspect` is free-text).
   `mission_brief` is the slot the classifier/strategist/planner already read as their primary forward
   input, but it's owner-supplied and nothing generates it. Fix: have **`domain-submitter`** (it owns
   mission persistence and runs *before* the classifier) **always write a `mission_brief`** — a fixed
   platform **standing-ambition principle**, merged with the owner's mission when supplied. Logic in a Go
   action, not workflow branching. No new aspect/agent, no reorder, no schema change. Draft principle text
   is in `running_notes.md` (checkpoint x) for the owner to edit.

## The design finding (important — corrects an earlier claim)

The look is **not** produced by an LLM reading `design_intent`. The decider is **`site-design-planner`**,
which is **deterministic** ("no LLM"): it matches a **layout** from a library by industry-tag overlap, a
**typography** set by font match (with fallback), and a **palette** by a cascade
(`design_reference → mission → design_intent → archetype_default`). `brand-designer` (5 hardcoded themes on
Haiku) and `visual-designer` (image stub) are experimental/vestigial.

For idea.uk it resolved to layout **`tool-portal-dark`**, typography **`sans-modern`** (a *fallback* —
Fraunces/IBM Plex aren't in the library), palette via **`archetype_default`** (not the parchment/rust in
`design_intent`). So the build **genericised** the design; it did not clone the current one. The reason is
structural: the classifier writes `design_intent` as **prose** (hex inside a sentence), but the palette
cascade wants **structured** values like adoption's `reference_values {primary, secondary, accent}` — so for
a fresh (non-adopted) site the classifier's design intent mostly doesn't reach the installed design.

**So there are two workstreams, and they're different problems:**
- **Content ambition** → the standing-ambition mission default (above). LLM-read, so it works.
- **Design leadership** → a *library + cascade* problem, untouched by the mission text: (a) fix the
  prose→structured mismatch so an intended palette applies for fresh sites; (b) curate distinctive layouts /
  typography sets / archetype-default palettes in the libraries; (c) decide whether `tool-portal-dark` is a
  leading look or a generic template — which needs seeing it rendered.

## Next steps (rough order)

1. **Look at what `tool-portal-dark` actually renders as** (and ideally how gamesdesign.co.uk came out —
   was its content/design ambitious or generic?). This tells us how much library work the design needs and
   whether the conservatism is mainly a sites-with-an-existing-page problem.
2. **Decide the design approach** for leadership: fix the palette cascade (classifier emits a structured
   palette, or the cascade parses hex from prose) and/or curate the layout/typography/palette libraries.
3. **Build the content fix:** reuse-check `domain-submitter`'s `persist_mission_brief` / `write_site_spec`
   path, then add the standing-ambition default/merge as a Go action (no schema change). Finalise the
   principle wording with the owner first.
4. **Tidy the duplicate idea.uk spec rows** so the planner reads current specs.
5. **Re-run the front of the pipeline** with the direction set, and review the result on the B2 staging URL.
6. **Phase 2 — VM deploy** (in the runbook): static build → `/var/www/idea.uk`; nginx serves `/` + the
   framework pages and **proxies the reserved tool paths** to the Go service on `127.0.0.1:8080`; never
   serve `/stripe/webhook` or operator paths as static; test the full pay → webhook → deliver flow on the
   new config. Cutover is one nginx line; rollback is one nginx line. DNS unchanged.

## Working rules (owner's, strict)

Go not Python. Plain language, no hype or flattery. **Reuse before rebuild; fix the framework structurally
rather than patch by hand.** Confirm live schema/API/product facts before asserting or coding — every time
(`0 rows` is not decisive). Keep agent workflows simple, put complexity in Go action code, keep variable
names in sync with the actions, no sub-workflows in SQL (spawn sub-agents), distinct agent responsibilities.
Check the DB schema before writing SQL. Deploy via GitHub → GH Actions → Backblaze B2. British English. Low
risk appetite; reasonable step sizes. Update `running_notes.md` each checkpoint.

## Key references

- `RUNBOOK_idea_uk_chassis_site_and_vm_deploy.md` — full detail for this stream (pipeline, design finding,
  VM-deploy plan).
- `running_notes.md` — the journal; checkpoints (u)–(y) cover this stream's analysis and the draft
  standing-ambition text.
- `HANDOFF.md` + `RUNBOOK_idea_uk.md` — the **live tool** (Stripe, email, operator flow). Don't conflate.
- gamesdesign.co.uk — a reference site the same pipeline already built and deployed.

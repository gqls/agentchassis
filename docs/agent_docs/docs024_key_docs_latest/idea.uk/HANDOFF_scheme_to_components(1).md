# Handoff — make a site's scheme reach its components

This is the brief for a dedicated thread. Three companion files are attached alongside it: **`REPORT_scheme_does_not_reach_components.md`** (the full plan — investigations, design questions, fix shape, checklist), **`RUNBOOK_idea_uk_chassis_site_and_vm_deploy.md`** (how we got here + the idea.uk steps), and **`running_notes_2.md`** (the running journal with the evidence, checkpoints lll–ppp). Read the report first; this brief just orients you and carries the working context the report doesn't.

---

## The task, in one paragraph

The chassis builds multipage sites by choosing a layout and palette — the site's "scheme", light or dark — at composition time and rendering a stylesheet from it, but that scheme only reaches the stylesheet's colour variables; it never reaches the components that actually render each section, header and footer. Those components are drawn from a dark-oriented library by a one-active-component-per-function lookup and carry their own inline CSS that hardcodes a dark treatment, so a site resolved to a light scheme still renders dark chrome (dark header, dark image hero, dark CTA, dark footer) over light content. The task is to make a site's scheme reach its components so a light site renders coherently light and a dark site stays dark — expressing the light/dark difference as a **variable-value override consumed by a single component** (palette plus the existing `--section-*` mechanism), **not** by duplicating components into `*-light`/`*-dark` variants, and splitting into new component functions only where one is genuinely too structurally different to share. It is a framework-level fix that must be designed — the open questions resolved — before any code, then migrated across the component library and back-filled to already-built sites without breaking dark sites.

---

## The problem in brief (enough to orient; full evidence in the report §2–§3)

- A site's scheme is decided at composition (`design_intent.style_direction` → the layout's `scheme` + palette) and reaches `styles.css` `:root`. Components that read the palette variables (`--color-*`) adapt; components that hardcode their colours/treatment do not.
- Sections resolve to components by **direct function lookup** in `plan_sections` (one active component per function — the scoring selector is not used for current sites), so there is no place to pick a scheme-appropriate variant, and the library has only dark `hero`/`call-to-action`/footer.
- Components **self-style** with their own inline CSS using their own class names; the layout stylesheet styles different class names, so it only reaches components via the `:root` palette + element rules. That is why two sections went light (they read `--color-*`) and the rest stayed dark (they hardcode dark or set a dark `--section-*` context).
- Header/footer are site-level (`site_components`) and are not scheme-derived: no layout declares a default header/footer, and the composition path does not run `update_site_defaults`.
- A renderer mechanism to make sections adapt (`--section-*` set from background luminance) **already exists** — the dark components bypass it by hardcoding. Reusing it is likely the cleanest path.

## The hard constraint (the design steer)

Express scheme as an **override**, not a proliferation of functions. One component declares its structure once and reads scheme-supplied variables; the light/dark difference is the values those variables take (the override), applied from the site scheme + a per-section contrast intent. New functions only where a component is genuinely too structurally different to share. This is the lens for every design decision below, and it means the work is mostly about plumbing the scheme signal and de-hardcoding components — not authoring variants, and not (mainly) about the selector.

## Why it matters (mission)

The platform exists to build sophisticated, industry-appropriate sites that are "best for the users of the site." A coherent, intentional colour scheme — rather than a dark/light patchwork that falls out of which components happen to exist — is part of that promise. This gap defeats it for any light site, so it is worth doing properly.

---

## Working constraints (how to operate in this codebase)

**Strict preferences:** Go, not Python. Plain human language, no LLM-hype, no flattery; banned words "perfect"/"critical"/"excellent". Confirm live API/schema/data facts before asserting or coding — a `0 rows` result is not decisive until the query/state is checked. Reuse and adapt existing functions/structures before creating new ones; fix the framework structurally over one-off patches. British English. Low risk appetite, reasonable step sizes, at most one question per reply. Minimal formatting (prose over bullets). No `logger.Debug` (it doesn't show — use `logger.Info`). Don't call a fix "final"/"last". Don't create summary documents unless asked. Keep the runbook + running notes current.

**Architecture conventions:** every agent is an orchestrator that owns a workflow of steps calling actions; keep workflows simple and put complexity in Go actions; don't build sub-workflows in SQL — spawn sub-agents with their own workflows; agents respond to the caller's (parent's) responses topic; `agent_definitions.type` (not `name`); `processing_mode` is nested inside `default_config`. Check database schemas before writing SQL.

**Infrastructure:** Kubernetes namespaces `-n ai-persona-system` (app pods) and `-n kafka`; Kafka cluster `personae-kafka-cluster`, bootstrap `personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092`. DB: `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`. Run SQL as FILES (`… < file`), never pasted — pasting into interactive psql mangles `\set`/`\echo`/blank lines and can leave an open transaction. The chassis deploys via github → GH Actions → Backblaze B2.

**Safety / blast radius:** idea.uk's live £29 product is a separate Go binary on a Hetzner VM and is untouched — the chassis build deploys to B2 and DNS still points at the VM, so chassis changes are invisible to the live site. Changing a shared component template affects every site that uses it, so any migration must keep existing dark sites dark and be backed up + reversible.

---

## Cold-start sequence (first moves)

1. Read `REPORT_scheme_does_not_reach_components.md` end to end; skim `running_notes_2.md` checkpoints lll–ppp for the evidence trail.
2. Pull the report §7 checklist up front (agent definitions, Go actions, schemas, queries).
3. Run investigations **A, B, C, D** first (report §6) — the render path, the scheme signal, the library inventory, and the class-name contract audit. Together they answer the gating question **Q4** (report §5), which shapes everything else.
4. With A–D in hand, settle the section-contrast model (Q2) and the override mechanism (Q3), then the header/footer wiring (F/Q6).
5. Only then design the change against the provisional fix shape (report §8), validating or revising it. Plan the migration + backfill (H) and the audit guard (I) as part of the design, not after.

## Definition of done

- A site's resolved scheme reaches its components: a light-scheme site renders coherently light (header, hero, CTA, footer and content), and a dark-scheme site still renders dark, with intentional per-section contrast preserved.
- The light/dark difference is carried as a variable override consumed by shared components (no `*-light`/`*-dark` duplication beyond genuinely-divergent cases).
- The existing component library is migrated, and already-built sites are re-rendered, without regressing dark sites.
- There is a guard (an audit check) that flags scheme incoherence so this can't silently regress.
- The report/runbook/notes are updated to reflect what was actually built.

## Out of scope for this thread

Finishing idea.uk (its post-fix page rebuild, site review and VM cutover) and the parallel chassis backlog (P2) are tracked in `TODO_chassis_and_idea_uk.md` and `RUNBOOK_idea_uk_chassis_site_and_vm_deploy.md` (steps 7–8). The original chat will resume those once this scheme→components fix is done.

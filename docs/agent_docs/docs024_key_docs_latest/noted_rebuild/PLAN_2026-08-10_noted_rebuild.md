# PLAN — rebuilding noted.co.uk on the framework

Started 2026-08-10. Design, phasing, decisions **and their reasons**.
Corrections to the originating brief live here, marked as corrections.

---

## 1. What we are trying to do

noted.co.uk is a note-taking app the owner built by hand: text notes, up to five
voice recordings and five photos per note, five revisions of version history, and
a share button. It runs entirely in the browser on IndexedDB. It is live, it
works, and it has been serving out of `b2://portfolio-sites/noted.co.uk` since
January 2026.

Three things are wanted:

1. **Rebuild it through the framework**, so it is a framework site rather than a
   hand-made artefact sitting in the framework's bucket.
2. **Move the notes to a server on an existing VM**, with sign-in, so a person's
   notes follow them to another browser or phone instead of living in one
   browser's local storage.
3. **Full framework control** — and this is the requirement that shapes
   everything below.

### 1a. The owner's clarification, 2026-08-10 — the word is DECOMPOSITION

> **CORRECTION to how this was first framed.** The opening brief said "none of it
> is locked", and I began planning against `locked_at` — the HITL lock columns on
> `page_components` / `site_components` / `assets` / `site_plan_directives`. The
> owner corrected this the same day:
>
> > *"the keyword I was looking for was decomposition rather than locked because
> > locked can mean many things. I just want it all controlled fully by the
> > framework — upgrades, maintenance, tools checking, everything"*
>
> This is a much stronger and much more specific requirement, and it rules out
> the adoption route that would otherwise have been the obvious one.

**Decomposition means the site is broken down into the framework's own native
parts, at the granularity the framework's own tools operate on.** The opposite —
and the thing being ruled out — is the site arriving as a small number of opaque
blobs that the framework can store and deploy but cannot reason about.

That opposite is a real, supported path, which is why it has to be explicitly
rejected: `082 --fidelity locked` → `adopt_verbatim.go` stores **each page as ONE
`ported-page` component** with `content_data.deploy_mode='verbatim'` and
`pages.rebuild_policy='owned'`. A site adopted that way is deployed by the
framework and monitored at the page level, but:

- `save_page_sections` **refuses** to write to it (`save_page_sections_action.go:173`);
- `rerender_single_page_action.go:287` treats `owned` as "not the pipeline's to rebuild";
- the planner cannot re-plan it, so upgrades never reach it;
- every section-level discovery check has one giant section to look at, so it
  finds nothing useful.

That is "hosted by the framework", not "controlled by the framework". The owner
asked for the second.

### 1b. What decomposed actually means here, part by part

The framework's units, and what each part of noted becomes:

| Framework unit | Table / mechanism | What of noted lives here |
|---|---|---|
| Pages | `site_plan_pages` → `pages` | home, how it works, privacy, terms, the app itself, account pages |
| Sections | `site_plan_sections` → `page_components` | every band of every prose page, individually re-plannable and individually checkable |
| Content | `content_items` | the words, separate from the layout that shows them |
| Tool components | `content_components` at `component_level='tool'` | the note editor itself |
| Client JS | component `js_content` → `/tools/assets/{fn}.js`, and `js_snippets` | the app's behaviour, deployed as real assets, not pasted inline |
| **Experiences** | `experience_patterns` + `site_experiences` | **the app's BEHAVIOUR, declared as contracts with checks** |
| Cross-cutting rules | `site_plan_directives` | voice, palette, writing rules — survives plan rebuilds |
| Specs | `site_specs` | mission, identity, design intent, evidence base, imagery |

The row that does the heavy lifting for an *application* is **experiences**.
Everything else on that list is how the framework decomposes a brochure site.
An app is not mostly prose, so if decomposition stopped at sections we would
have a beautifully decomposed set of marketing pages wrapped around one opaque
lump of app — which is the very thing being ruled out, just moved down a level.

`experience_patterns` is the mechanism that prevents that. Each row carries
`contract`, `states`, `degraded_states`, `data_contract`, `requires_invariant`
and a `criteria_template`, and `site_experiences` binds a pattern to a site with
a status of `proposed → bound → verified → broken`. So a behaviour like "a note
you typed on your phone is there when you sign in on a laptop" becomes a
first-class, named, checkable object rather than an emergent property of some
JavaScript.

There are 9 patterns today; the closest precedent is
`timed-remote-challenge-loop`, which is a page talking to a remote engine over
HTTP with an honest degraded state. noted needs its own patterns, and writing
them is a large part of this work.

---

## 2. The one place "everything" cannot be met today — say it plainly

**The framework does not generate backend code, and nothing in it currently
does user accounts.** This is not a gap I can paper over, so it goes near the
top.

Measured, not assumed:

- Every server-side thing on the estate is a hand-written Go service, one per
  box: `site-engine` (relojistas), the idea.uk Stripe service, `webdesign-chat`.
  The capability summary states it directly: *"The framework does not generate
  backend code… Nothing generates it per site and no agent writes server code."*
- The closest thing to per-user persistence is `webdesign-chat`'s `store.go`,
  and it deliberately stops short of accounts — state keyed by conversation ID
  and client IP, stdlib only, no DB driver, because *"this box holds no cluster
  credential and dials nothing in"*.
- **`auth-service` cannot be reused for this.** It is MySQL-backed, `ClusterIP`
  only with no ingress anywhere in the kustomize tree, and its users are platform
  operators — the `/api/v1/admin/*` group proxies to core-manager and is the
  admin dashboard's backend. It authenticates the people who run the site
  factory, not the visitors to a site the factory built.

So the notes server will be **hand-written Go**, in the same family as the three
that already exist. That is the sanctioned pattern for this estate, and it is
**not** a breach of the "never hand-build a site" ruling of 2026-08-04, which is
about site HTML being uploaded outside the pipeline. The *site* goes through the
framework; the *engine* is written, like every other engine here.

What the framework **can** own about the server, and what we will make it own:

- its **contract**, declared in the experience `data_contract` (endpoints, caps,
  session rules, rendering rules);
- its **liveness**, via `check_backend_unreachable` — which is why the seed sets
  `deploy_config.target='vm'`; the check NOOPs on anything else
  (`check_backend_unreachable.go:48`). Note it is currently seated on **no live
  agent**, so this enables it rather than switching it on;
- its **observable behaviour**, via Tier-4 `interaction` checks that drive a real
  browser through sign-in and note round-trips;
- its **honest failure**, via `degraded_states` on the experience.

**The honest summary: everything except the server binary itself can be
framework-decomposed, framework-upgraded and framework-checked. The binary can be
framework-monitored and framework-contracted, but a human writes it.** If that is
not acceptable, the alternative is building a backend generator, which is a
platform programme rather than a site rebuild, and should be decided as one.

---

## 3. Hosting

### 3a. Where it goes

Four VMs exist. The relevant two:

| Box | Spec | Already has | Notes |
|---|---|---|---|
| **webdesign.uk** (Mythic Beasts `vds:webdesign`, Cambridge) | 2 core / 8 GB / 50 GB SSD | nginx, `sitesync` pull timer, Cloudflare tunnel (nothing inbound), a Go service pattern | Newest, roomiest, and its ingress model is the safest on the estate |
| **tools-api island** (Mythic Beasts, £16.20/mo) | 1 core / 2 GB | **its own Postgres**, Caddy, nightly `pg_dump` to Mythic backup space, docker-compose | Deliberately isolated. Has the database and the backup habit already |

**Recommendation: the webdesign.uk box.** Reasons, in order: it has the headroom
(8 GB vs 2 GB, and notes with audio and photos are storage-hungry), its
Cloudflare-tunnel ingress means the notes service can bind loopback-only exactly
as `webdesign-chat` does, and its `sitesync` pull model means the framework
deploys by committing to `gqls/vm-sites` and never needs a credential on the box.
It needs Postgres added, which is a known quantity.

The island is the tempting alternative because the database is already there, but
1 core / 2 GB shared with the tools API, for a service holding the only copy of
strangers' notes, is the wrong trade. **This is an owner decision and I have not
made it irreversible** — nothing shipped today depends on which box wins.

### 3b. Why the seed sets `github_repo='vm-sites'`, and why that is safety

This is the single most important operational decision in the file.

The default deploy path is `gqls/sites` → GitHub Action → `b2 sync --delete
--skip-newer <domain> b2://portfolio-sites/<domain>` → Cloudflare worker. The
worker maps hostname+path directly onto the bucket prefix.

**noted.co.uk is live out of that exact prefix right now**, holding a banner that
tells users to come back and export their recordings. If the site were seeded
with the default repo and then built, the first framework page render would
`--delete` its way over the running application, destroying the thing the banner
is pointing people at.

`github_repo='vm-sites'` sends framework commits to `gqls/vm-sites`, which
reaches the VM estate and never touches the bucket. The legacy app keeps serving
until a deliberate cutover. This is written into the seed file's header too,
because the failure mode is silent and the fix looks like a tidy-up.

---

## 4. Phasing

**Phase 0 — protect the users. DONE 2026-08-10, live and verified.**
- Live app brought under version control in `gqls/sites` (it was in no repo at
  all; there was no way to review, revert or reproduce a change to it).
- Wind-down notice shipped.
- **"Save everything"** shipped, because the notice was otherwise dangerous: the
  existing backup button exported the `notes` store alone. Verified
  behaviourally — a note with a 4,096-byte audio blob and an 8,192-byte image
  blob produced a **334-byte** backup containing only
  `['content','createdAt','id','title','updatedAt']`. Recordings, photos and
  history were silently dropped. Round-trip now tested: export → wipe IndexedDB →
  restore → media byte-identical including MIME type.
- Three WCAG-AA contrast failures fixed, found by `scripts/render_audit.py`.
- Framework seeded: `sites` row, `evidence_base` (7 bans, 0 facts, pinned),
  `imagery_style_guide` (pinned).

**Phase 1 — decide and prepare the box.** Owner picks the VM. Add Postgres,
schema for accounts + notes + media, backup and restore, TLS/tunnel. Nothing
user-visible.

**Phase 2 — write the notes engine.** Hand-written Go, stdlib-leaning, in the
family of `site-engine` and `webdesign-chat`. Accounts, sessions, note CRUD,
media upload, and an import endpoint that accepts **exactly the full-backup JSON
shipped in Phase 0** — that is the migration path for existing users, and it is
why the export format was given a `format` string and a `version` number rather
than being an anonymous blob.

**Phase 3 — write the experience patterns.** Before the app is rebuilt, not
after. Each behaviour gets a contract, degraded states, a data contract and a
criteria template. Draft list: `authenticated-note-sync`, `local-first-capture`,
`media-attachment-capture`, `backup-export-restore`. This is the work that makes
"the framework checks every part of it" true rather than aspirational.

**Phase 4 — build the site through the pipeline.** Mission brief → `082` →
research → strategy → briefing → planner → pages. Prose pages decompose
normally. The editor becomes a `tool`-level component with its JS deployed as a
real asset. `rebuild_policy` stays `generic` everywhere — **no `owned` pages**.

**Phase 5 — cut over.** Point noted.co.uk at the VM, keep the legacy bucket copy
reachable at a versioned path for a grace period, and only then retire it.

---

## 5. Constraints the checking layer imposes — read before writing criteria

From the `timed-remote-challenge-loop` pattern's own `_unsupported` notes, which
are the most honest inventory of the criteria runner's limits:

- `expect_within_ms` **does not exist**; the runner asserts 300 ms after a step.
- `retries` does not exist.
- Checks **cannot be ordered or made conditional on another check** — the author
  of that pattern recorded that this made their central honesty rule
  "currently unexpressible".
- There is **no way to induce or simulate a failing dependency**, so the
  honest-degradation clause "cannot be tested by the platform at all".

Those four gaps bite a notes app hard, because the things most worth checking —
"the note is still there after a reload", "a failed save does not silently lose
text", "signing in on a second browser shows the same notes" — are ordered,
stateful, and dependency-dependent by nature.

Two consequences: do not write criteria that assume these gaps are closed
without re-reading `experience_criteria.go` first (that pattern is `status=draft`
and its notes are dated); and expect that closing one or two of them may be a
genuine platform contribution that comes out of this workstream.

Also relevant, and measured: **discovery only detects.** 204 items sat `detected`
against 2 `triaged` across 10 sites on 2026-08-03; nothing promotes findings
automatically, and `quality-discovery-agent` — which carries `unverified_claims`
and `voice_tells` — has filed 7 items in its entire history. So "the framework
checks it" will need the promotion path thought about, or the checks will run and
their findings will sit where nobody looks.

---

## 6. Open decisions for the owner

1. **Which VM.** Recommendation above (webdesign.uk box), not yet acted on.
2. **What the privacy promise becomes.** The current site's entire pitch is that
   there is no server. That inverts. The `evidence_base` bans the old sentences
   so they cannot be copied forward, but *what replaces them* is a product
   decision, not a copy decision. The strongest honest position available
   cheaply is "your notes are on our server so you can reach them anywhere; we
   can technically read them; we don't" — end-to-end encryption would let us say
   more but is a real build and is currently banned as a claim precisely because
   it is not built.
3. **Whether existing users' notes are migrated automatically or by hand.** There
   is no server-side copy of anyone's notes and no way to reach their browsers,
   so the only migration path is a person exporting and importing. Phase 0's
   format was designed for it, but it needs a person to act.
4. **How long the legacy app stays up after cutover.**

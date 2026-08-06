# HANDOFF 2026-08-06 — webdesign.uk: infrastructure DONE and proven; the site itself is REJECTED and rebuilds under a roadmap

**Start here cold. Supersedes `HANDOFF_2026-08-05_continue_here.md`** (kept for
the trail; its §4a corrective plan is folded in here, updated). Read with:
`PLAN_2026-08-04_webdesign_uk_vm_hosting.md` (architecture; §2a one-box ruling)
· `SUMMARY_2026-08-06` (plain-prose state incl. the tunnel explanation) ·
`NOTES` 08-04→08-06 (the livelock class, the contamination diagnosis, the prompt
review) · `WRONG_CALLS.md` 08-06 (why "verified" was wrong twice).

## 0. One paragraph

Every piece of machinery now works and is proven end to end: the box serves, the
tunnel carries, the pipeline builds, the claims gate blocks, deploys flow
repo→box in ≤5 min. The **product on it is rejected**: the owner reviewed v1 and
called it — one page, unstyled, brittle copy, no domain-input box. All four
complaints are diagnosed to causes (below). The job now is a **rebuild through
the framework's own triggers, constrained by a roadmap**, with all copy
rewritten under the owner's improved writer prompt. Nothing is publicly broken:
apex+www 302 to webdesign.co.uk; the rejected v1 is visible only at
`https://preview.webdesign.uk/`.

## 1. State table — verified 2026-08-06, re-check before trusting

| thing | state | re-check |
|---|---|---|
| Box `webdesign.vs.mythic-beasts.com` (root, key `~/.ssh/webdesign_box_ed25519`) | live; ufw deny-in; nginx **127.0.0.1:8080 only**; sitesync 5-min | ssh + `ss -tln` |
| **Tunnel `webdesign-box`** `81f59f78-dda8-40a0-984b-cfadb36bc891` | **LIVE**: service installed, 4 edge connections; config `/etc/cloudflared/config.yml` (apex, www, preview→8080; preview overrides Host) | `systemctl status cloudflared` on box |
| `preview.webdesign.uk` | **serves the REJECTED v1** through the tunnel (200) | curl |
| apex + `www` | **302 → webdesign.co.uk** via TWO host-scoped Page Rules (`6d4d5b67…`, `88794916…`) — ⚠ a `*webdesign.uk/*` wildcard also catches the preview; that is WHY there are two | curl all three |
| CF token | **NEW + working**: `~/.config/cloudflare/token` (id `806c8a11…`, expires **2026-09-01**), scopes: webdesign.uk + ugg2.com DNS/PageRules/Zone-read, account Tunnel Edit + Workers Read. Test with a REAL zone call; `/user/tokens/verify` is exempt from IP filters and lies | pagerules GET |
| Workers custom domains | **0 for webdesign** (1 account-wide, fundamentallyai) ⇒ **cutover is DNS-only** | `GET /accounts/13044f178ae0b156961065f55c8fada8/workers/domains` |
| `sites` row | `github_repo='vm-sites'`, `deploy_config target=vm capabilities=[backend]` | SELECT |
| Writer model | **`claude-sonnet-5` — KEPT by owner ruling** (202; revert SQL in the bug file is now historical) | agent_definitions |
| `bugs_open/202` | effectively resolved by the swap; closes on a full clean busy day | file |
| `evidence_base` | **15 `banned_claims`** (+superlatives 08-06), 7 attested facts, pinned | SELECT |
| Writer prompt | reviewed 08-06: overclaim guard HOLDS (rule 14, never-promise-accuracy); HOUSE VOICE is style-only and bans em dashes prompt-side | NOTES 08-06 |
| `*.ugg2.com` previews | intact, untouched | curl any subdomain |

## 2. Why v1 was rejected — four complaints, four diagnosed causes

1. **One page.** The classifier crawled the (deleted) 08-03 hand-built page,
   read it as *"existing live site with strong copy already in place"*, and
   classified `landing` at **0.97** — *"one scrolling page … is the right
   form"*. Planner honoured it: plan `4ecaa120…` has **1** page vs fleet norms
   of 19–33. Third distinct cost of the hand-built error; contamination
   CONFIRMED from the classifier's own `detected_signals`.
2. **Unstyled.** `vm-sites/webdesign.uk/` holds ONLY `index.html`; the page
   references `/assets/css/styles.css`, logo, JS — all 404 on the box. Asset
   actions are git-blind (they write to B2); the mechanism that populated
   `idea.uk/assets/` and `relojistas.com/assets/` in the repo is
   **unidentified**. My "no gap" claim was sibling-inference; WRONG_CALLS 08-06.
3. **Brittle copy.** Contaminated `content_direction` + this lane's
   restraint-heavy writer_block + the pre-improvement prompt. The prompt is now
   improved (HOUSE VOICE); owner ruling: **rewrite everything under it**.
4. **No domain-input box.** Correctly sequenced behind the chat service — but a
   mailto is not the product, and the input box is part of acceptance.

## 3. THE PATH — in order, each step gated on the one before

1. **Diagnose the VM asset path** (blocks everything visual). How do the two
   precedent sites' `assets/` actually reach vm-sites? Candidates:
   `site-asset-renderer`, the deploy Action, `render_directory`, a manual step.
   Find it, make webdesign.uk use THAT. If it turns out manual for the
   precedents too, that is a platform gap — file it with this page as evidence.
   **Do NOT hand-copy assets into the repo** (the hand-built error in
   miniature).
2. **Regenerate `content_direction` from scratch.** It is contamination-derived
   throughout; two phrase-patches were whack-a-mole. Supersede with a fresh
   generation (or delete + let the resubmission's classifier rewrite it — the
   contamination sources are gone: bucket objects deleted, apex 302s).
3. **Resubmit with an authoritative ROADMAP** — pages: home, how-it-works,
   what-you-get, faq (incl. the VAT/price answers), contact. `build-site-planner`
   treats a roadmap as authoritative ("build ONLY the pages listed").
   ⚠ **`082_submit_domain_unified.sh` has NO `--roadmap-file` flag** — hand-roll
   the envelope from the script's own kcat block with `roadmap_brief` +
   `roadmap` (the oufe RUNBOOK §3 documents exactly this, with the gotchas:
   plain prose, single-lined, **no figures in briefs** — numbers live in
   `evidence_base.facts`).
4. **Let the rebuild regenerate ALL copy** under the improved prompt. The 15
   bans + validation gate are the hard floor; the prompt is the soft layer.
5. **Verify like a visitor, then never claim done otherwise**: every
   `href/src` resolved from the serving root (`grep -oE '(href|src)="[^"]*"'` →
   curl each), `scripts/render_audit.py` (VIZ-010) as render witness, and check
   `pages.title` + `meta_description` by hand — **the ban sweep does not cover
   the head** (landmine, 08-05).
6. **Chat service** (Phase 4 of the VM plan): hand-written sibling of
   site-engine on `127.0.0.1:8081` (sanctioned exception — nothing generates
   backend code). §5.1 controls FIRST: per-IP via `CF-Connecting-IP` (tunnel
   makes it unforgeable; prove with the `139` two-network check), turn cap,
   daily spend ceiling failing closed to contact details, request log,
   transcripts as rows. Model: `claude-haiku-4-5` (intake ≠ product). Needs the
   owner's scoped Anthropic key.
7. **Input-box section** replaces the mailto — pinned section posting
   same-origin to `/api/chat`, CTS-044 generation guards (external loader, no
   inline script). Never ship it before 6 exists.
8. **Owner reviews on `preview.webdesign.uk`** → only on his approval:
   **cutover** = point apex+www DNS at the tunnel (`cloudflared tunnel route dns
   --overwrite-dns` or via the token), delete the two Page Rules. No Worker
   step remains. The 302 stays until that moment.

## 4. Gotchas most likely to bite the next session (all earned here)

- **Verify at the artefact as a VISITOR** — two "verified" claims died in this
  lane because every check was a text check.
- **When a content ban fires, sweep ALL specs with the validator's own regexes
  before re-driving** — a banned phrase upstream is a livelock, and a one-phrase
  fix is a mole. Triage: instructs-page FIX / quoted-example FIX / avoid-list
  KEEP.
- Blockers live in **`agent_error_log`** (`occurred_at`, `context->'issues'`) —
  not orchestration rows, not pod logs.
- Re-drives: items to `'triaged'`, `error=NULL`, then the 076 heartbeat
  (**extract from the shebang** — SQL notes sit above it) — then verify **by
  payload**; kcat exit 0 proves nothing. Respect the 300s post-roll window.
- Page-rule wildcards catch subdomains (`*webdesign.uk/*` ate the preview);
  edge changes take ~30s — re-test before diagnosing.
- `pkill -f` in a compound ssh command can match the shell carrying it via ANY
  later text (memory: shell-tool-traps).
- The stub-resolver landmine: a timeout is evidence about the PATH; pin with
  `--resolve` before believing any outage.

## 5. Owner ledger

**Owed:** scoped Anthropic key (step 6) · correction-fee number · terms before
live Stripe · final review + cutover approval (step 8).
**Settled, do not reopen:** £1,200, no VAT · Sonnet 5 stays on the writer ·
framework-only builds (CLAUDE.md ruling — and this lane has now paid for it
three ways) · one box per trust class; customer sites are static→B2 · box spec
as billed · wait for Gemini Tier 2 (background, nothing owed).

## 6. Access map

`~/.ssh/webdesign_box_ed25519` → root@box · `~/.config/cloudflare/token` →
NEW token (works; expires 09-01) · `gh` = gqls, ADMIN on vm-sites · `b2` CLI ·
kubectl → cluster/DB. **No Mythic panel, no Stripe, no Cloudflare dashboard.**

# HANDOFF 2026-08-05 — webdesign.uk: box live, pipeline wired end to end, blocked only on Gemini quota

**Start here cold. Supersedes `HANDOFF_2026-08-04_continue_here.md`** (kept for the
trail — its §3 architecture question is RESOLVED and its Phase list is mostly DONE).
Read with: `PLAN_2026-08-04_webdesign_uk_vm_hosting.md` (the design, §2a one-box
ruling, §2b order sheet) · `NOTES` (04/05 entries: the livelock class) ·
`SUMMARY_2026-08-04b` (what the framework can/cannot do) · `bugs_open/202`.

## 0. One paragraph

The Mythic Beasts box (`vds:webdesign`) is **provisioned, hardened, and pulling**
from the `vm-sites` repo every 5 minutes; the site row is flipped to the VM class;
the claims layer caught and we then cured a **spec-contamination livelock** traced
to the deleted hand-built page; and the build now fails only on **`bugs_open/202`**
— the fleet's Gemini quota is exhausted, an owner-level decision, not a defect in
this lane. The tunnel login is the one owner step outstanding (links expire in
minutes — mint fresh when present). DNS still holds the 302; nothing is publicly
broken.

## 1. State table (re-check every row before trusting it)

| thing | state | re-check |
|---|---|---|
| Box `webdesign.vs.mythic-beasts.com` (176.126.243.62 / 2a00:1098:5e2::2, Cambridge) | **live, provisioned** | `ssh -i ~/.ssh/webdesign_box_ed25519 root@…` |
| Firewall | default-deny inbound; `:22` the ONLY public listener; nginx on `127.0.0.1:8080` only | `ss -tln \| grep -v 127.0.0.1` on the box |
| Pull-sync | timer active, 5-min; sparse clone `gqls/vm-sites` → `webdesign.uk/` only; folder-absent guard proven | `systemctl status sitesync.timer`; run `/usr/local/bin/sitesync` as www-data |
| Deploy key | on-box, **read-only** on gqls/vm-sites, id `159299585` | `gh api repos/gqls/vm-sites/keys` |
| cloudflared | **installed (2026.7.3), tunnel NOT created** — login expired unused | `/root/.cloudflared/` empty |
| `sites` row | `github_repo='vm-sites'`, `deploy_config={"target":"vm","capabilities":["backend"]}` | SELECT |
| Landing page | **failed on `bugs_open/202`** (Gemini 429), NOT on content | `agent_error_log` (`occurred_at`, `context->'issues'`) |
| Specs | `content_direction` superseded **twice**; current row sweeps clean against all 14 bans | §3 |
| Stale hand-built objects | **deleted** from `portfolio-sites` (both) | `b2 ls` both prefixes |
| webdesign.uk public | 302 → webdesign.co.uk (Page Rule), apex `192.0.2.1` | curl |
| `*.ugg2.com` previews | proxied wildcard, Worker route — intact, unaffected | curl any subdomain |
| CF API token | IP-locked (rotating ISP address); verify endpoint EXEMPT so it lies "active" | RUNBOOK |

Box artefacts are **versioned in this lane's `box/`** (setup-webdesignbox.sh,
sitesync, webdesign.uk.nginx) — idempotent, re-runnable.

## 2. The immediate blocker: `bugs_open/202`

`page-content-writer` generates via **`provider=gemini model=gemini-pro-latest`**
and the quota is exhausted — 429 on five domains inside 28h (gaswholesalers 123×),
so this is fleet infrastructure, not this lane. Three options in the bug file;
re-pointing a shared writer's model is **council-gate territory**. Cheapest test:
Google quotas often reset daily — re-drive once after the reset window:

```sql
UPDATE site_work_items wi SET status='triaged', error=NULL FROM sites s
 WHERE s.id=wi.site_id AND s.domain='webdesign.uk'
   AND wi.item_type='needs_page' AND wi.status='failed';
```
then the 076 heartbeat (**extract from the shebang** — the file has SQL notes
pasted above it, `bash 076…` dies at line 3), then verify **by payload**:
`orchestration_states` rows + items leaving `triaged`. kcat exit 0 proves nothing.

## 3. The livelock class — READ BEFORE TOUCHING ANY SPEC OR RE-DRIVING A BLOCKED PAGE

Three build attempts, three lessons, one rule:

1. **Attempt 1** blocked: `banned_claim "A person check"`. The phrase was
   *instructed* by `content_direction` (classifier-written). Writer-follows-spec +
   gate-blocks-output = **livelock**; no retry can exit it.
2. **Attempt 2** blocked: `banned_claim "template"` — instructed by **the row the
   first fix inserted**, which fixed one phrase and carried the rest. One-phrase
   supersedes against a 19KB spec = whack-a-mole at one build per mole.
3. **The cure**: pull the WHOLE current spec, run **all 14 `banned_claims`
   regexes** over it, and triage matches three ways — **instructs-the-page → fix**
   · **quoted example copy → fix** (an em dash inside a quoted example is an
   instruction to violate) · **avoid-list teaching → KEEP** (it teaches; removing
   it costs signal). One supersede, verified clean, `created_by …2026-08-05`.
4. **Contamination source**: both banned phrasings were **verbatim lines from the
   deleted 08-03 hand-built page** [ingestion path INFERRED, untraced]. The
   hand-built error kept costing after deletion — it had been laundered into the
   site's own specs. (The framework ruling in CLAUDE.md exists for this reason.)

Where blockers actually live: **`agent_error_log`** — `occurred_at` not
`created_at`, detail in `context->'issues'`. NOT the orchestration rows (empty
`blockers` arrays), NOT chassis pod logs (nothing on either replica).

Attempt 3 (post-cure) died on 202 **before generation** — so the cleaned spec is
**unproven against the writer** until a build passes validation. First passing
build: read the rendered page for em dashes, person-checks phrasing, and the
price stated as £1,200-total-no-VAT.

## 4. Remaining work, in order

1. **Tunnel** (owner-present step, minutes): mint fresh login → owner clicks,
   picks `webdesign.uk` zone → then on the box: create tunnel `webdesign-box`,
   config → `http://127.0.0.1:8080`, `cloudflared service install`, verify
   outbound connections. **DNS untouched.**
2. ~~202 clears → …~~ **DONE 08-05 (Sonnet swap) — but see the CORRECTION below.**
   > **CORRECTED 2026-08-06, by the owner's screenshot.** The parenthetical
   > claim "the page deploy commits assets into the repo — both precedent
   > folders prove it; no gap" was **WRONG**: `vm-sites/webdesign.uk/` holds
   > ONLY `index.html`, and the served page 404s on `/assets/css/styles.css`,
   > the logo, images and JS — the owner saw an unstyled grey page. The
   > precedent folders' `assets/` arrived by a mechanism never identified;
   > verify on OUR artefact, not a sibling's. Full entry: `WRONG_CALLS.md`
   > 2026-08-06. **The corrective work is §4a below and it is the lane's top
   > priority.**
3. **Two parked items**, owner-flavoured, not urgent: `needs_section_data`
   (pricing section wants `tier_1_features` — we are ONE price, so the section
   likely changes shape rather than gaining tiers) and `unresolved_cta` (hero CTA
   has no destination page — probably becomes the chat/contact anchor).
4. **Chat service** (Phase 4 of the VM plan): hand-written sibling of
   site-engine on `127.0.0.1:8081`; §5.1 controls FIRST; Haiku for intake;
   `CF-Connecting-IP` at the tunnel boundary + the `139` two-network proof.
5. **Cutover** (Phase 6): remove the Worker **Custom Domain** binding (dashboard
   only — no zone route exists and the token lacks account scope), delete Page
   Rule `b8e08b35028315a274b2f5c7fea9154d`, let cloudflared write DNS. The 302
   stays until everything above is verified on-box.

### 4a. CORRECTIVE PLAN (owner review, 2026-08-06) — the top of the queue

The owner reviewed https://preview.webdesign.uk/ and rejected it: unstyled (the
asset gap above), **one page**, and **no domain-input box** (a mailto is not the
product). His standing instruction: **the site is rendered by framework
submission triggers, never by this CLI** — hold to it while fixing.

1. **Diagnose the VM asset path.** How do `idea.uk/assets/` and
   `relojistas.com/assets/` actually get into vm-sites? (`site-asset-renderer`?
   the deploy Action? a manual step?) Find the mechanism, then make webdesign.uk
   use THAT — do not hand-copy assets into the repo, which would be the
   hand-built error in miniature. If the mechanism turns out to be manual for
   the precedents too, that is a platform gap worth filing, with this page as
   the evidence.
2. **A real multi-page site.** Re-drive through `082`/the pipeline with a
   roadmap so the planner builds the pages the product needs (how-it-works,
   what-you-get, FAQ/terms, contact) — `build-site-planner` honours roadmap
   `section_types`/page lists ("build ONLY the pages listed"). The one-page
   plan came from an unconstrained submission; constrain it this time.
3. **The domain-input box** replaces the mailto — which requires the chat
   service (Phase 4, hand-written by ruling-sanctioned exception) to exist
   first, or at minimum a form endpoint on the box. Do not ship a dead input.
4. **Verify like a visitor before ever reporting again**: subresources resolve
   (grep href/src → curl each), then `scripts/render_audit.py` (VIZ-010) for a
   render witness. A prose sweep is NOT a render (WRONG_CALLS 08-06).
5. **The apex is DARK** (Page Rule + Worker binding removed in the dashboard,
   ~08-06): restore the holding 302 as soon as a working CF token exists —
   first action on the new token. The 302 stays until the owner approves the
   REAL site.

## 5. Owner ledger

**Owed by owner:** tunnel click (§4.1) · correction-fee number · terms before
live Stripe (P3) · scoped Anthropic key for the chat (Phase 4).
**202 RULED 2026-08-05: WAIT for automatic Tier 2** (~$100 cumulative, a few
weeks). Interim: 250 pro calls/day fleet-wide, reset ~midnight UTC — **re-drive
just after a reset**, and a late-day 429 anywhere is 202, not a new bug.
**Settled, do not reopen:** £1,200 price · no VAT (stated on page in three
places) · one box per trust class (customer sites are STATIC → B2, no box) ·
box spec as billed (£43.15/mo, 52GB SSD, IPv4 taken) · framework-only builds
(CLAUDE.md ruling).

## 6. Access map (what this machine can reach)

`~/.ssh/webdesign_box_ed25519` → root@box (agent-loaded) · `gh` = gqls, ADMIN on
vm-sites · `b2` CLI → portfolio-sites + backups buckets · kubectl → cluster/DB ·
CF API token → 36 zones BUT IP-locked (rotates; `verify` lies) and no account
scope. **No Cloudflare dashboard, no Mythic panel, no Stripe.**

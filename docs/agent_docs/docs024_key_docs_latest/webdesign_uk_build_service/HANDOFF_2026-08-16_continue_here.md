# HANDOFF 2026-08-16 — PLAN step 5 BUILT + council-APPROVED (deployer half TL-043, inert till roll) and PROVEN at service level on the box (one `sitechat` binary; webdesign rolled onto it); the visitor-facing second-site proof is owner-gated on noted.co.uk's facts — SUPERSEDES HANDOFF_2026-08-15c

**Start here cold.** Read order: this file → NOTES 2026-08-16 entries → RUNBOOK
"One binary, several sites" → register TL-043 → PLAN_2026-08-11 step 6.

## 0. State in one paragraph

Step 5 of PLAN_2026-08-11 ("extend tool-deployer for the backend half; prove
on a second site sharing webdesign's box") is built on both sides. **Deployer
half:** `deploy_tool_to_site` refuses a `requires-backend` tool aimed at a
non-backend site (406's predicate as a named constant with a mutation-proven
LOCKSTEP test) or a site with zero `evidence_base` facts, and on a fresh deploy
raises ONE handler-less `backend_provision` item as the handover to the
operator (TL-043; commits `51c33f482`, `f3fd5af39`; council **APPROVED r1**,
corr `55cda19b`, 4 advisory objections all dispositioned with measurements).
**Inert until the next chassis roll.** **Box half:** the chat service is one
parameterised binary `/usr/local/bin/sitechat` + a `sitechat@<domain>`
template unit; live mode REQUIRES `SITE_DOMAIN`/`SITE_DESCRIPTION` and refuses
another site's facts by cross-checking the relay's `domain` field. webdesign.uk
is running on it (rolled 10:06Z; Journey A through the public edge answers
£149; now bound to 127.0.0.1:8081, was `*:8081`). Three transient proofs on
the box: relojistas.com params → same binary fetched its 13 facts and came up;
noted.co.uk → refused (zero facts); noted identity + webdesign URL → refused.
**What is NOT done:** a second site's VISITORS chatting — the only other site
on the box is noted.co.uk and its facts are 0 (owner's to write, per the noted
lane). One wrong claim of mine caught by the proof run and corrected in four
places + WRONG_CALLS ("the relay 404s" → it serves 200 with `facts: []`).

## 1. Owner decisions NEEDED

1. **Contact email** (unchanged): live `contact` fact = `webdesign@contactforsales.com`
   — deliberate sales inbox, or leftover? Gates the EXPERIENCE_PLAN's Step 0.
2. **Stripe webhook URL** (unchanged): apex/www 302 at the edge; register
   `https://preview.webdesign.uk/stripe/webhook` or add an edge exception.
3. **NEW — noted.co.uk's chat box:** if wanted, its `evidence_base` facts must
   exist first (0 today, and the noted lane says the privacy copy is yours);
   then provisioning is one RUNBOOK recipe + one nginx block. Also needs a
   one-line `SITE_DESCRIPTION` (owner copy — the bot introduces itself with it).
4. Standing: Stripe keys via terraform only; build-duration copy UNDECIDED;
   CTA ink vs egg-gold UNDECIDED.

## 2. Next work for THIS lane, in order

1. **PLAN step 6 — wire `tool-suggester` to cite the approved EXPERIENCE_PLAN**
   when recommending chat-input-box (config/prompt migration in the 406 idiom;
   the plan is `doc_plans` subject `experience`/`site-chat-intake`, is_current).
   Step 5's mechanism is proven, so suggestions no longer outrun delivery —
   but note the deployer half is INERT till the roll; step 6 can land as
   config independently.
2. **After the next chassis roll** (ask the pod: `build provenance` log line
   or `git merge-base --is-ancestor f3fd5af39 <stamp>`): TL-043 verify-later —
   drive ONE real `deploy_tool_to_site` of chat-input-box at a control that
   must REFUSE (noted.co.uk while facts=0 → the facts refusal; a static site →
   the capability refusal) and confirm the error text; do NOT deploy to a site
   you cannot provision.
3. **EXPERIENCE_PLAN MVP verification round** (Steps 0–4, no rebuilding) —
   Step 0 = owner decision 1.
4. **Watch items:** the separate lane's `bugs_open/285` (section_list_assembly
   slug — AMBIGUOUS NUMBER) fix — no commit touches
   `load_page_sections_from_spec_action.go` since 08-15 (checked this morning);
   lock stays ON; run the five-step acceptance when it lands, then answer
   `a4cd5dc8`. `bugs_open/282` (co-requisite), `bugs_open/275` (unclaimed).

## 3. What is LIVE and mine, verified 2026-08-16 morning

- **Box:** `/usr/local/bin/sitechat` md5 `d914d07a0821aa8d7f2abb40567e115d` ==
  local build; `webdesign-chat.service` ExecStart→sitechat, `BIND_ADDR=127.0.0.1:8081`;
  `/etc/webdesign-chat.env` + `SITE_DOMAIN`/`SITE_DESCRIPTION` (byte-identical
  intro). Backups `*.bak-20260815b` (env, unit, old binary). `sitechat@.service`
  installed, no instances enabled; `/etc/sitechat/` empty by design (no
  placeholder env files — a disabled unit someone starts later would introduce
  the bot with the placeholder).
- **Repo `box/`** == box for: `webdesign-chat.service`, `sitechat@.service`,
  `chat-service/*.go` (the binary is gitignored). `webdesign.uk.nginx` unchanged.
- **Facts relay / DNS / webhook proxy:** unchanged from 15c (relay proven live
  again this morning: `fetched 15 facts`).
- **Council:** corr `55cda19b` APPROVED; trailer `Council-Reviewed:` on `f3fd5af39`.

## 4. Traps met this session (also in NOTES/RUNBOOK/WRONG_CALLS)

- **A relative `cd` in a compound Bash command after the tool reset the cwd**
  silently `git checkout`-ed a file in the *previous* cwd — my own edits
  reverted, no other session involved. Use absolute paths; never chain
  `git checkout --` after a `cd` that can fail.
- **The relay serves 200 + `facts: []`** for a site whose evidence_base has an
  empty facts array (noted.co.uk); it 404s only when the KEY is absent. Read
  the count, not the status.
- **`SITE_DOMAIN`/`SITE_DESCRIPTION` are now REQUIRED in live mode** — the
  old binary's env lacks them; a re-roll of an OLD env against the NEW binary
  refuses to start (loud, by design). The RUNBOOK env recipe has both.
- The submission's edit list should ENUMERATE the register/index files it
  ships (two seats could not verify them from the plan alone).

## 5. Falsifiers / re-check before trusting this file

- A newer handoff here; whether the chassis has rolled past `f3fd5af39`
  (TL-043 goes live then — item 2 above becomes runnable).
- `ss -ltnp | grep sitechat` on the box shows `127.0.0.1:8081` only.
- The 285 fix landing (grep commits for the slug; `who-owns.py`).
- noted.co.uk's facts count (`jsonb_array_length` on its evidence_base row).

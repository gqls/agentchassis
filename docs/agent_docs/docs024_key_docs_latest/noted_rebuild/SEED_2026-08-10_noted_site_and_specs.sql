-- ============================================================================
-- noted.co.uk — site row + evidence_base + imagery_style_guide
-- Written 2026-08-10. Applied out of band (psql -f), NOT via the migration
-- runner: this is per-site setup, not a platform schema change.
--
-- WHY THESE THREE, AND WHY BEFORE SUBMISSION — the same three reasons as the
-- oufe seed, which is the worked example this copies:
--   1. sites row WITH AN EMAIL. bugs_open/063: the hallucinated-email check
--      FAILS OPEN when a site has no contact email. ensure_site_record upserts
--      on domain, so pre-creating is safe and wins the race.
--   2. evidence_base. The whole claims layer is gated on the PRESENCE of this
--      aspect — loadEvidenceBase returns nil and every lane silently no-ops
--      (validate_page_content.go:727-746). Seeding it before the first page is
--      written is the only way the first page is covered.
--   3. imagery_style_guide. bugs_closed/027: content_hero generates unstyled on
--      any site that has none, and a brand-new site has none.
--
-- ============================================================================
-- THE ONE THING TO READ BEFORE CHANGING ANYTHING HERE
-- ============================================================================
-- github_repo IS SET TO 'vm-sites' DELIBERATELY, AND IT IS LOAD-BEARING SAFETY,
-- NOT A PREFERENCE.
--
-- noted.co.uk is ALREADY LIVE out of b2://portfolio-sites/noted.co.uk — a
-- working note-taking app holding real users' only copy of their notes, in
-- their own browsers. It is served by the Cloudflare worker
-- `portfolio-sites-router`, which maps hostname+path straight onto that bucket
-- prefix.
--
-- The default deploy path for a framework site is gqls/sites -> a GitHub Action
-- -> `b2 sync --delete --skip-newer <domain> b2://portfolio-sites/<domain>`.
-- So if this site were seeded with the DEFAULT repo and then built, the first
-- framework page render would `--delete` its way over the live application,
-- while the app's own banner is actively telling users to come back and export
-- their recordings. That is the accident this line prevents.
--
-- With github_repo='vm-sites' the framework's commits go to gqls/vm-sites
-- instead, which reaches the VM estate and NOT the bucket. The legacy app keeps
-- serving noted.co.uk untouched until a deliberate DNS/worker cutover.
--
-- DO NOT "tidy" this to the default repo to make an early build appear at
-- noted.co.uk. That is precisely the bug.
-- ============================================================================
--
-- ON THE EVIDENCE BASE, AND WHY IT IS MOSTLY BANS
--   This site's product promise is about to INVERT. Every word of the current
--   copy is a privacy absolute earned by there being no server at all:
--     guides/about.html:34  "we can't see your notes, read your text, or
--                            listen to your recordings"
--     index.html            "Local & Private. Notes are stored in your browser."
--   The rebuild adds accounts and server-side storage so a person can reach
--   their notes from another browser. The moment that ships, those sentences
--   become false, and they are exactly the sentences a writer agent will copy
--   forward from the old site because they read like established brand voice.
--
--   So the bans below target that specific failure: absolute privacy claims,
--   zero-knowledge/encryption claims we have not built and cannot evidence, and
--   the "no server" family. They are deliberately fail-closed. If we later DO
--   build end-to-end encryption, the ban forces a conscious return to this file
--   to remove it, having first registered the capability as a fact with a
--   source. That friction is the point — an unearned privacy claim is the one
--   class of copy on this site that could actually harm somebody.
--
--   facts[] is EMPTY on purpose (the oufe pattern). Nothing about the rebuilt
--   product is true yet: there is no server, no account, no sync. A fact
--   registered before it is built is a claim, and this file is the thing that
--   is supposed to stop those. Populate it as each capability actually ships,
--   with attested_by or a sql source, NOT from this comment.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

-- ---------------------------------------------------------------- site row --
INSERT INTO sites (
  domain, name, network_id, status, email, company_name, tagline,
  github_repo, deploy_config
)
VALUES (
  'noted.co.uk',
  'noted.co.uk',
  '00000000-0000-0000-0000-000000000002',
  'active',
  'hello@noted.co.uk',
  'Noted',
  'Notes that stay yours',
  'vm-sites',                                        -- see the block above
  '{"target": "vm", "capabilities": ["backend"]}'::jsonb
)
ON CONFLICT (domain) DO UPDATE
  SET email         = COALESCE(sites.email, EXCLUDED.email),
      github_repo   = COALESCE(sites.github_repo, EXCLUDED.github_repo),
      deploy_config = CASE WHEN sites.deploy_config = '{}'::jsonb
                           THEN EXCLUDED.deploy_config ELSE sites.deploy_config END;
-- NOTE: status='active' is what upsertSite writes, but it is NOT in the
-- validated vocabulary (draft/building/review/published/deployed/archived/
-- error). Never scope a query by it and expect meaning.
--
-- deploy_config.target='vm' is also what makes check_backend_unreachable
-- non-inert for this site — it NOOPs unless target='vm'
-- (discovery_checks/check_backend_unreachable.go:48). That check is currently
-- seated on NO live agent, so setting this enables it but does not run it.

-- ----------------------------------------------------------- evidence_base --
UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE site_id = (SELECT id FROM sites WHERE domain = 'noted.co.uk')
  AND aspect = 'evidence_base' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT
  id,
  'evidence_base',
  $eb${
    "governing_rule": "This site is a note-taking application that holds people's own writing, voice recordings and photographs. Every statement about where that data goes, who can read it, and what happens if something fails must describe what the software ACTUALLY does today — not what is planned, not what is nearly finished, and not a reasonable-sounding approximation. Where a protection has not been built, the page says so plainly or says nothing; 'we have not built this yet' is always publishable and a comforting overstatement never is. A privacy claim on this site is a promise about a stranger's private writing, so it is held to the same standard as a financial claim: it traces to a registered fact, or it does not appear.",
    "audit_doc": "docs/agent_docs/docs024_key_docs_latest/noted_rebuild/PLAN_2026-08-10_noted_rebuild.md",
    "schema_notes": "facts[]: {id, claim, kind: metric|capability|entity|attestation, source: EXACTLY ONE of {sql|artifact|attested_by|citation}, verified_at}. banned_claims[]: {pattern (case-insensitive regex; an invalid regex degrades to a literal substring, so a typo never silently drops a ban), reason}.",
    "facts": [],
    "banned_claims": [
      {
        "pattern": "(we|noted)[^.]{0,40}(can'?t|cannot|never)[^.]{0,30}(see|read|access|listen to)[^.]{0,20}your",
        "reason": "The pre-rebuild site said 'we can't see your notes, read your text, or listen to your recordings' (guides/about.html:34). That was TRUE when there was no server and is FALSE the moment accounts and server-side storage ship. It is the single most likely sentence to be copied forward from the old site as established voice."
      },
      {
        "pattern": "(no|zero|without a)[ -]?(server|servers|cloud|backend)",
        "reason": "The 'no servers, just your browser' family (old og:description). The rebuilt product is server-backed by design — that is the whole point of being able to sign in from another browser."
      },
      {
        "pattern": "end[- ]?to[- ]?end encrypt|zero[- ]?knowledge|we (have )?no access to your",
        "reason": "Not built. E2E encryption and zero-knowledge are specific, verifiable architectures with real cost; claiming either without building it is the most harmful possible lie on a notes product, because a user would reasonably store secrets on the strength of it. Registering this as a fact requires the architecture to exist and be named."
      },
      {
        "pattern": "(military|bank)[- ]?grade|unhackable|100% (secure|private|safe)|completely (secure|private)",
        "reason": "Security absolutes. Unfalsifiable, unearnable, and a tell that nobody measured anything."
      },
      {
        "pattern": "your (notes|data) (are|is) (always )?(safe|secure|protected|backed up)\\b",
        "reason": "An unconditional durability promise. Until the server-side backup and restore path is built AND tested, this is false; after it is built it still needs its actual guarantee stated (what is backed up, how often, what a restore recovers) rather than a blanket adjective."
      },
      {
        "pattern": "never lose (a note|your notes|anything)|can'?t lose",
        "reason": "Same family as above. The legacy app could and did lose recordings for anyone who trusted its text-only backup button — see the 2026-08-10 finding in NOTES. This product has no standing to make that promise."
      },
      {
        "pattern": "gdpr[- ]compliant|fully compliant|iso ?27001|soc ?2",
        "reason": "Compliance postures are attestations with auditors and dates behind them. None exists. Naming a regulation we are subject to is fine; claiming a certification is not."
      }
    ],
    "allowed_entities": [
      "IndexedDB", "Backblaze B2", "Cloudflare", "Chrome", "Firefox", "Safari", "Edge"
    ],
    "writer_block": "This is a note-taking app that holds people's private writing, voice recordings and photographs.\n\nDo not write privacy or security promises. Describe mechanism instead: say where data is stored and who can reach it, in plain words, and let the reader draw the conclusion. 'Your notes are saved to your account on our server, so you can sign in on another device and find them there' is publishable. 'Your notes are completely secure' is not, and neither is anything about encryption, zero-knowledge, or us being unable to read your notes.\n\nDo not carry any wording forward from the old version of this site. The old site had no server at all and its privacy language was earned by that fact; the rebuilt product has a server, so the same sentences are now untrue.\n\nDo not state figures — no user counts, no uptime, no storage limits, no 'trusted by' numbers. There are no registered facts on this site yet, so any number is invented.\n\nWrite plainly, in British English, for someone who wants somewhere to put a thought quickly. Short sentences. No 'seamless', 'effortless', 'powerful', 'revolutionise', 'unlock'."
  }$eb$::jsonb,
  'manual',
  'Seeded 2026-08-10 at the start of the framework rebuild. facts[] deliberately empty — nothing about the server-backed product is built yet. Bans target the privacy language of the PRE-REBUILD site, which becomes false the moment accounts ship.',
  true,
  true,
  'noted-rebuild-workstream-2026-08-10'
FROM sites WHERE domain = 'noted.co.uk';

-- ----------------------------------------------------- imagery_style_guide --
UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE site_id = (SELECT id FROM sites WHERE domain = 'noted.co.uk')
  AND aspect = 'imagery_style_guide' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT
  id,
  'imagery_style_guide',
  $ig${
    "medium": "Clean flat vector illustration and honest product screenshots. No stock photography of people at desks.",
    "mood": "Calm, plain, unhurried. This is a tool for catching a thought, not a productivity system to be mastered.",
    "palette": "Paper white and warm off-white grounds, near-black ink for text, a single blue accent (#2563eb) carried from the existing app. Muted amber only for warnings.",
    "avoid": [
      "stock photography of people at laptops",
      "3D glossy app mockups floating at an angle",
      "padlocks, shields, keyholes or any security iconography — this site must not imply protections it has not built",
      "brain/lightbulb 'idea' cliches",
      "dark neon developer aesthetics",
      "hand-drawn scribble textures that make the text look unreliable"
    ],
    "kinds": {
      "content_hero": {
        "medium": "Flat vector illustration, generous whitespace, one idea per image.",
        "mood": "Quiet and legible at small sizes.",
        "palette": "Off-white ground, near-black line, one blue accent.",
        "avoid": ["padlocks or shields", "photorealism", "busy compositions"],
        "reference_asset_keys": []
      }
    },
    "reference_asset_keys": []
  }$ig$::jsonb,
  'manual',
  'Seeded 2026-08-10. The security-iconography ban is deliberate and matches the evidence_base bans: a padlock illustration makes the same unearned promise a sentence would, and no claims check reads images.',
  true,
  true,
  'noted-rebuild-workstream-2026-08-10'
FROM sites WHERE domain = 'noted.co.uk';

COMMIT;

-- ------------------------------------------------------------------ verify --
-- Each of these could come out wrong; that is why they are here.
SELECT domain, email, github_repo, deploy_config->>'target' AS deploy_target, status
FROM sites WHERE domain = 'noted.co.uk';

SELECT aspect, pinned, is_current,
       jsonb_array_length(data->'banned_claims') AS bans,
       jsonb_array_length(data->'facts')         AS facts
FROM site_specs
WHERE site_id = (SELECT id FROM sites WHERE domain = 'noted.co.uk')
  AND is_current
ORDER BY aspect;

-- The bans must actually compile as regexes AND match the sentences they were
-- written for. A ban that matches nothing is the failure mode this catches:
-- an invalid regex degrades to a literal substring and silently stops banning.
SELECT b->>'reason' AS ban_reason,
       'we can''t see your notes, read your text, or listen to your recordings'
         ~* (b->>'pattern') AS catches_old_about_sentence
FROM site_specs s,
     LATERAL jsonb_array_elements(s.data->'banned_claims') b
WHERE s.site_id = (SELECT id FROM sites WHERE domain = 'noted.co.uk')
  AND s.aspect = 'evidence_base' AND s.is_current
LIMIT 3;

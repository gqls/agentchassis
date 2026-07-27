-- ============================================================================
-- 226_overclaim_patterns_oufe.sql — arm the EXISTING scanner for the
-- overclaimed-reliability class on oufe.com
--
-- CORRECTION THIS FILE EXISTS TO MAKE
--   Migration 223's header asserts: "NOTHING IN THE ESTATE LOOKS FOR THIS …
--   invisible to every scanner". **That is wrong**, and 223 has been corrected.
--
--   `ScanBannedClaims` (datahelpers/claims.go:284-325) is a bare
--   case-insensitive regex over prose blocks. It contains no number extraction,
--   no `businessClaimContextRe`, no `isExcludedNumber` — those gate only
--   `ScanUnregisteredNumbers` (claims.go:365,369). Verified by reading the
--   function body, not by grep.
--
--   So the machinery to catch a qualitative claim about our own reliability has
--   existed the whole time. **No pattern for the class had ever been written on
--   any site**, and there is still no mechanism to write one once. Capability
--   was never the gap; coverage was.
--
-- WHAT ARMING IT BUYS, FOR ONE UPDATE AND NO IMAGE ROLL
--   * V1a build gate — severity **blocker** (validate_page_content.go:930): a
--     page carrying one of these cannot be built.
--   * V1b post-deploy sweep — severity **high** + a HITL work item
--     (check_unverified_claims.go:377), scanning STORED rendered_html of live
--     pages, which is the surface the oufe promise actually shipped on.
--
-- THE LINE THESE PATTERNS DRAW
--   **You may describe what you DO; you may not claim what that GUARANTEES.**
--   "We cite every figure and date it" passes — it is a process commitment we
--   control and can keep. "A claim without a source does not appear here" is
--   banned — it asserts a completeness nobody can verify, including us. That
--   distinction is why the accuracy patterns are anchored to self: a law site
--   must still be able to write "the statute is the authoritative text".
--
-- TESTED BEFORE APPLYING, both directions: 10 fabrication shapes blocked
--   (including all four phrases oufe shipped live), 13 legitimate sentences pass
--   — among them the honest replacement copy now on the site and the approved
--   disclaimer's own wording. A false positive here is a BLOCKER that fails a
--   whole page build, so testing the pass-list matters as much as the block-list.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

-- Read-before-write: merge into the existing array, never replace the row.
UPDATE site_specs ss SET
  data = jsonb_set(ss.data, '{banned_claims}',
           (ss.data->'banned_claims') || $add$[
  {
    "pattern": "(claim|figure|number|statistic)s? without a[^.]{0,40}source[^.]{0,20}(does not|do not|doesn't|don't) appear",
    "reason": "completeness-of-exclusion: claims that NOTHING unsourced is published. Unverifiable by anyone, including us."
  },
  {
    "pattern": "(does not|doesn't|do not|don't) appear here",
    "reason": "completeness-of-exclusion, short form."
  },
  {
    "pattern": "if we (can'?t|cannot)[^.]{0,60}(it |they )?(doesn'?t|don'?t) appear",
    "reason": "completeness-of-exclusion, conditional form."
  },
  {
    "pattern": "every[^.]{0,30}(is|are) (verified|checked|confirmed|validated)",
    "reason": "verification-of-everything: a claim about outcomes, not process."
  },
  {
    "pattern": "(fully|independently|externally|properly) (verified|audited|fact.?checked)",
    "reason": "verification overclaim: implies an assurance process we do not run."
  },
  {
    "pattern": "(you|readers?) can rely on (this|it|us|these|our)",
    "reason": "invites reliance \u2014 the opposite of the standing posture, and the negligent-misstatement exposure route."
  },
  {
    "pattern": "(we|our|us|this site|this page|this analysis|this report|oufe|everything here|our (analysis|research|reporting|coverage|work)|all of (our|the) (figures|claims|numbers)) (is|are|remains?) (always )?(accurate|authoritative|definitive|error.?free|complete|reliable)\\b",
    "reason": "self accuracy overclaim. Anchored to self so 'the statute is authoritative' still passes."
  },
  {
    "pattern": "guaranteed (accurate|correct|complete|up.?to.?date|reliable)",
    "reason": "accuracy guarantee."
  },
  {
    "pattern": "(this|our) (discipline|method|process|standard) is not a disclaimer",
    "reason": "repudiates the caveat. Shipped live on oufe.com 2026-07-26."
  },
  {
    "pattern": "never (wrong|inaccurate|invents?|fabricates?|makes? (a )?mistakes?)",
    "reason": "infallibility claim."
  }
]$add$::jsonb),
  notes = COALESCE(ss.notes,'') || ' | 2026-07-26: +overclaimed-reliability patterns (mig 226)'
FROM sites s
WHERE s.id = ss.site_id AND s.domain = 'oufe.com'
  AND ss.aspect = 'evidence_base' AND ss.is_current
  AND NOT (ss.data->'banned_claims')::text LIKE '%is not a disclaimer%';

COMMIT;

-- Verify
--   SELECT jsonb_array_length(data->'banned_claims') FROM site_specs ss
--     JOIN sites s ON s.id=ss.site_id
--    WHERE s.domain='oufe.com' AND aspect='evidence_base' AND is_current;
--
-- Then prove it against the LIVE pages rather than trusting the row:
--   go run ./cmd/claimscan -evidence <eb.json> -components <components.tsv>

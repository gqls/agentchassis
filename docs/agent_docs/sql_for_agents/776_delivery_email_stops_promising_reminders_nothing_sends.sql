-- 776: the delivery email stops promising reminders that nothing sends.
--
-- bugs_open/477, THIRD surface. Found by the `bugs_open/477` lane, who measured
-- the live agent row rather than the seed, and flagged it to this lane before
-- writing SQL because it lands on a jsonb path this lane had said it would edit.
--
-- ── THE FALSE CLAUSE, verbatim from the live row ────────────────────────────
--
--   WHEN YOU HAVE MOVED
--   Once your site is off our hosting, press the button here so we stop reminding you:
--   {{confirm_link}}
--
-- Nothing in this estate sends a reminder. `[MEASURED 2026-09-04]` one agent can
-- send mail at all (`delivery-email-sender`), ZERO scheduled tasks target it, one
-- Go action can send mail, and `sites.transfer_confirmed_at` — the column the
-- button writes — has no reader outside the file that writes it.
--
-- ── WHY THIS SURFACE IS THE WORST AND THE CHEAPEST AT THE SAME TIME ─────────
--
-- WORST: the two `renderConfirm` pages carrying the same clause are hypothetical
-- until somebody clicks. This one was SENT. The only delivery this estate has ever
-- made (`idea.uk`, 2026-09-03 19:30:31Z) carried it, so a real recipient has been
-- told we would remind them.
--
-- CHEAPEST: the copy is CONFIG by deliberate design — `send_delivery_email_action.go:8`
-- states it as a property, "No copy. The template lives in the STEP CONFIG (DB,
-- owner-editable)". So this is live on apply. The Go pages the 477 lane corrected
-- in `76ec663d3` are inert until a core-manager roll.
--
-- ── WHAT CHANGES, AND WHAT DELIBERATELY DOES NOT ────────────────────────────
--
--   before: press the button here so we stop reminding you:
--   after : press the button here to tell us you have moved:
--
-- **The paragraph STAYS.** Only the promise inside it goes. The customer still
-- needs a reason to press, and "tell us you have moved" is the honest one — it is
-- what `ConfirmTransfer` actually does. Deleting the sentence outright would leave
-- a bare instruction with no reason, which is why this is a rewording and not the
-- deletion the 477 lane applied to the two pages: their sentences already stated
-- the reason separately ("Pressing the button below tells us you have moved
-- everything across"), and this one did not.
--
-- The wording deliberately MATCHES that page, so the letter and the page a
-- customer reaches from it say the same thing in the same voice.
--
-- ⚠ **RESTORE THE STRONGER WORDING WHEN THE FOLLOW-UP SENDER SHIPS.** "so we stop
-- reminding you" is exactly right once something reminds them, and it is a better
-- reason to press than this one. The 477 lane is building that sender. This
-- migration is the honest interim, not the destination.
--
-- ROLLBACK: 776_..._ROLLBACK.sql restores the false clause. It exists for
-- forward-only recovery, not because the old text is worth having.

BEGIN;

CREATE TABLE IF NOT EXISTS bak_776_delivery_email_20260904 AS
SELECT * FROM agent_definitions
WHERE type = 'delivery-email-sender'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ── anchor guard: abort loudly if the line has moved, so that if another lane
--    edits this template first this migration refuses rather than clobbers ────
DO $$
DECLARE n int; tpl text;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='delivery-email-sender' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '776: expected exactly 1 live delivery-email-sender row, found % (duplicate-active-row landmine)', n;
  END IF;

  SELECT default_config->'workflow'->'steps'->'send_email'->'config'->>'body_template'
    INTO tpl
    FROM agent_definitions WHERE type='delivery-email-sender' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF tpl IS NULL THEN
    RAISE EXCEPTION '776: send_email.config.body_template is absent — the step moved';
  END IF;
  IF position('press the button here so we stop reminding you:' in tpl) = 0 THEN
    RAISE EXCEPTION '776: the anchor line is not present — already applied, or another lane edited this template first. REFUSING rather than clobbering.';
  END IF;
END $$;

-- ── the change: one clause, by exact replacement ────────────────────────────
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,send_email,config,body_template}',
      to_jsonb(replace(
        default_config->'workflow'->'steps'->'send_email'->'config'->>'body_template',
        'press the button here so we stop reminding you:',
        'press the button here to tell us you have moved:')),
      false),
    updated_at = NOW()
WHERE type='delivery-email-sender' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- ── verify: DO/RAISE, because ON_ERROR_STOP will not abort a COMMIT on a
--    SELECT that merely returns the wrong rows ─────────────────────────────────
DO $$
DECLARE tpl text; old_len int; new_len int;
BEGIN
  SELECT default_config->'workflow'->'steps'->'send_email'->'config'->>'body_template'
    INTO tpl
    FROM agent_definitions WHERE type='delivery-email-sender' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF position('so we stop reminding you' in tpl) <> 0 THEN
    RAISE EXCEPTION '776 VERIFY: the false clause is still present';
  END IF;
  IF position('press the button here to tell us you have moved:' in tpl) = 0 THEN
    RAISE EXCEPTION '776 VERIFY: the replacement line is absent';
  END IF;

  -- Exactly ONE clause changed and nothing else: the length delta must be the
  -- arithmetic of the two literals, or something other than this edit happened.
  SELECT length(bak.default_config->'workflow'->'steps'->'send_email'->'config'->>'body_template')
    INTO old_len FROM bak_776_delivery_email_20260904 bak LIMIT 1;
  new_len := length(tpl);
  IF new_len - old_len <> length('press the button here to tell us you have moved:')
                          - length('press the button here so we stop reminding you:') THEN
    RAISE EXCEPTION '776 VERIFY: length delta is % , expected % — more than one clause changed',
      new_len - old_len,
      length('press the button here to tell us you have moved:') - length('press the button here so we stop reminding you:');
  END IF;

  -- The parts that must be UNDISTURBED. {{confirm_link}} especially: the send
  -- action refuses to send when the template names a placeholder whose link this
  -- claim did not produce, so losing it would be a silent behaviour change.
  IF position('WHEN YOU HAVE MOVED' in tpl) = 0
     OR position('{{confirm_link}}' in tpl) = 0
     OR position('{{zip_link}}' in tpl) = 0
     OR position('{{live_site}}' in tpl) = 0
     OR position('YOUR FILES' in tpl) = 0
     OR position('KEEPING IT ONLINE' in tpl) = 0
     OR position('THE DOMAIN' in tpl) = 0 THEN
    RAISE EXCEPTION '776 VERIFY: a section heading or placeholder was disturbed';
  END IF;

  RAISE NOTICE '776: the delivery email no longer promises reminders. Paragraph kept, promise removed, % chars', new_len;
END $$;

COMMIT;


<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### LIABILITY_AND_TERMS and legal pages (terms, refund, privacy) — AI-disclosure requirement
- **category:** NEW:legal-and-compliance
- **status-signal:** deployed
- **status-evidence:** `running_notes(44).md`: "/terms and /refund-policy pages written + served," "privacy policy added; terms hardened (AI disclaimer)" — all shipped as live routes, explicitly flagged as drafts pending a "~£200-500 fixed-fee UK solicitor review needed before going live."
- **what:** Three plain-language legal pages built directly into the idea.uk Go binary (`termsBody`/`refundBody`/`privacyBody` constants, `{{EMAIL}}` templated at serve time): terms (explicitly states reports are AI-generated and AI "can be confidently wrong and invent facts... treat everything as to-be-checked... entirely your responsibility and not ours"), refund policy (14-day no-reason refund plus fault/non-delivery refund), and a UK-GDPR-shaped privacy policy naming processors (Stripe, Anthropic) and flagging the US data-transfer point. Grew out of a liability analysis (`LIABILITY_AND_TERMS.md`) triggered directly by the Risk-column near-miss (SFI single-farm assessment) — identifies the real legal exposure as common-law negligent misstatement (Hedley Byrne) rather than any formal regulatory regime, since SFI navigation itself isn't formally regulated.
- **sources:** `running_notes(44).md` (three consecutive 2026-06-05 checkpoints)
- **relations:** Risk-as-hazard scoring dimension (the trigger); idea.uk product
- **verify-later:** whether solicitor review has actually happened; `/terms`, `/refund-policy`, `/privacy` routes live

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### LIABILITY_AND_TERMS and legal pages (terms, refund, privacy) — AI-disclosure requirement
- **category:** NEW:legal-and-compliance
- **status-signal:** deployed
- **status-evidence:** `running_notes(44).md`: "/terms and /refund-policy pages written + served," "privacy policy added; terms hardened (AI disclaimer)" — all shipped as live routes, explicitly flagged as drafts pending a "~£200-500 fixed-fee UK solicitor review needed before going live."
- **what:** Three plain-language legal pages built directly into the idea.uk Go binary (`termsBody`/`refundBody`/`privacyBody` constants, `{{EMAIL}}` templated at serve time): terms (explicitly states reports are AI-generated and AI "can be confidently wrong and invent facts... treat everything as to-be-checked... entirely your responsibility and not ours"), refund policy (14-day no-reason refund plus fault/non-delivery refund), and a UK-GDPR-shaped privacy policy naming processors (Stripe, Anthropic) and flagging the US data-transfer point. Grew out of a liability analysis (`LIABILITY_AND_TERMS.md`) triggered directly by the Risk-column near-miss (SFI single-farm assessment) — identifies the real legal exposure as common-law negligent misstatement (Hedley Byrne) rather than any formal regulatory regime, since SFI navigation itself isn't formally regulated.
- **sources:** `running_notes(44).md` (three consecutive 2026-06-05 checkpoints)
- **relations:** Risk-as-hazard scoring dimension (the trigger); idea.uk product
- **verify-later:** whether solicitor review has actually happened; `/terms`, `/refund-policy`, `/privacy` routes live

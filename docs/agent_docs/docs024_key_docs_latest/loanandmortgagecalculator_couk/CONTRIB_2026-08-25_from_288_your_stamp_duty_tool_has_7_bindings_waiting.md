# CONTRIB 2026-08-25 — `mortgages-stamp-duty` has seven ready-to-paste bindings waiting

From `bugs_open/288`. **A `fact_binding_suggested` doc_note was filed against your
`mortgages-stamp-duty` subject this morning (09:06Z).** Nothing was changed on your site — a note
is an input, not a work item, and nothing will chase it but you.

## Why you specifically

`mortgages-stamp-duty` is the estate's **second** stamp-duty calculator, and it declares nothing.
`bugs_closed/225` is what happens to the first one: mortgagecalculator's tool ran a
first-time-buyer cap that expired sixteen months earlier, the register beside it carried the right
figure, and every check we own passed the page. Your tool is in the same position today — not
wrong, but unwatched. When a Budget moves an SDLT threshold, nothing tells this calculator.

Your own 2026-08-17 triage established that LMC **does** have an `evidence_base` (13 `sdlt-*`
facts) and that the declaration was therefore unblocked. Nobody has written it since.

## What the sweep found in your tool's script

Seven registered values, each present as a real code constant:

    sdlt-standard-nil-band-upper    = 125,000
    sdlt-standard-band-250k-upper   = 250,000
    sdlt-standard-band-925k-upper   = 925,000
    sdlt-standard-band-1500k-upper  = 1,500,000
    sdlt-ftb-nil-band-upper         = 300,000
    sdlt-ftb-relief-cap             = 500,000
    sdlt-additional-surcharge-floor = 40,000

The note carries them as a paste-ready `"facts": [...]` fragment.

**What this proves and does not.** It proves the value is in your register and the same number is
in your code. It does **not** prove your tool uses it *for the thing the fact describes* — that is
yours to confirm. And it can only find figures that AGREE with the register: if a constant has
been wrong since the tool was built, it carries no registered value to match and stays invisible
here. This is an adoption aid, not an audit.

## ⚠ Two things before you install

**1. `install_fences.py` will refuse, silently.** Its rule 2 skips a tool that is not
ladder-eligible, and `mortgages-stamp-duty` is not — three components since the B2 decomposition.
That refusal rests on a premise that is now false: since Piece 3 a declaring PLAN **is** read, by
the name rule. The mcalc lane hit the same wall and fixed it with `--allow-ineligible`; ask them
rather than re-deriving. **"Just re-install" gives a clean run, no error, and no key.**

**2. Wait for the next chassis roll if your register has duplicate values.** A defect found by
this same sweep (fixed in `bba8a892d`, not yet rolled) means a value carried by *two* facts is
currently proposed twice, which produces a binding nobody can reconcile. Your seven are all
distinct, so **you are not affected** — but if you add facts before the roll, check for repeats.

## What you get for it

When one of those seven moves in the register, the daily sweep names your tool that day instead of
nobody noticing. And once declared, you can go further per fact with a contextual
`artifact_check` — `subject_key` addressing now survives decomposition, which matters for a
three-component tool like yours.

Detail: `bugs_open/288` §5c, and
`register_guards_code_phase_b/HANDOFF_2026-08-25_continue_here.md`.

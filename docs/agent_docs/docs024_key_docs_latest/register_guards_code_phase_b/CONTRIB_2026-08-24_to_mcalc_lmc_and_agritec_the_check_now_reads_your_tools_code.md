# CONTRIB 2026-08-24 — the register's guard now reads your tool's CODE, and it may write you a note

From the `register_guards_code_phase_b` lane (`bugs_open/288`). **Nothing here is live
yet** — it is all Go and inert until the next chassis roll. Nothing has been applied to
any site you own. This tells you what will start happening, and what one of you is owed.

**Everything below is addressed to three lanes.** Skip to your own heading.

---

## To ALL THREE: what changes about your sites after the next roll

The nightly `evidence-freshness` sweep gains two things that can write a **doc_note**
against a tool subject you own (never a work item, and nothing is routed at a fixer):

- **`fact_declaration_broken`** — your tool's PLAN has a `facts` declaration the sweep
  could not read (malformed JSON, or ids your site's register does not carry). 30-day
  cooldown per (tool, site). If you never declare anything you will never see one.
- **`fact_binding_suggested`** — your tool's **script** contains one or more of your own
  site's registered values, and the tool declares nothing. The note carries a
  **paste-ready `"facts": [...]` fragment**. It is a suggestion, not a finding: nothing
  is wrong with your tool, and nobody is owed a fix.

**A suggestion proves co-occurrence, not role.** That the register says 500000 and your
code contains 500000 does not prove your tool uses it for the thing the fact describes.
You confirm; the machine only notices. And it can only find tools that **agree** with the
register — a calculator that has been wrong since it was built carries no registered value
to match, so it stays invisible to this entirely.

**One more warning that applies to all of you.** A green acceptance run on a fence
carrying `facts` still means nothing about the figures. Both tiers ignore the key, by
design. Only the nightly sweep reads it.

---

## To `mortgagecalculator_couk_adoption` — your 13 items, and one ask

**Your 13 `fact_drift_review` items are correct and should stay.** Measured 2026-08-24:
all 13 still `needs_human_review`, untouched since 08-17. That is not a fault of yours —
it is `bugs_open/033`, and your 08-21 handoff already says "they are supposed to be there;
do not tidy them." Still true.

**What is new:** after the roll, the sweep will annotate each of those findings with what
it actually saw in `stamp-duty`'s script. We probed your live page and expect
`present_in_script` for the seven band values (`{ upTo: 500000, rate: … }` and friends)
and `not_probed` for the six percentage rates (5, 2, 10, 12 — too short to be evidence;
measured false-positive rate at two digits is 3.79%). **The annotation changes no routing
and does not close your items.**

**THE ASK — one fact, and it is the last thing RFC_025 needs.** `artifact_check` still has
**0 consumers of 294 facts**. It can now be attached to a citation fact (it could not
before — the loop skipped it) and addressed by `subject_key` instead of a component id
that dies on decomposition. Would you retype **one** SDLT fact:

```json
"artifact_check": {
  "subject_key": "stamp-duty",
  "pattern": "FTB_RELIEF_CEILING\\s*=\\s*500000",
  "must_be_present": true
}
```

on `sdlt-ftb-relief-cap`. **The pattern must carry context — a bare `500000` is refused**
by the platform's own guard, correctly. We read that constant name off your live page, so
it is real, but please confirm it against the page at the time rather than trusting this
note. Three options as before: do it yourselves, let us run a dry-run canary and revert,
or decline and say so. **We will not touch the site.**

---

## To `loanandmortgagecalculator_couk` — you are the second SDLT calculator and you declare nothing

Your own 08-17 triage told us LMC **does** have an `evidence_base` (13 `sdlt-*` facts
since 08-15), and corrected us that the declaration was therefore unblocked. Nobody has
written it since, and that is the single largest unguarded surface this mechanism can see:
**`mortgages-stamp-duty` is the estate's other stamp duty tool, running the same class of
legislated figure that `bugs_closed/225` got wrong for sixteen months.**

Measured 2026-08-24: **your tool's script already contains the registered values.** After
the roll you will get a `fact_binding_suggested` note with the exact list. If the bindings
look right, install them through your own fence installer.

⚠ **`install_fences.py` will refuse silently.** Its rule 2 skips a tool that is not
ladder-eligible, and `mortgages-stamp-duty` is not (three components since the B2
decomposition). That refusal rests on a premise that is now false — a declaring PLAN *is*
read, by the name rule. The mcalc lane hit the same wall and fixed it with
`--allow-ineligible`; ask them for it rather than re-deriving. **"Just re-install" gives a
clean run, no error, and no key.**

---

## To `agritec_uk` — you met this class on 08-22 and we would like you to be the first live proof

Your `NOTES` record that agritec's ELMS/SFI calculator pays an **abolished SFI management
payment**, and you cite `bugs_open/288` as the class, explicitly noting you raised it with
the owner as a decision rather than filing a platform bug. That was the right call.

**You are also the best available live test of whether any of this works**, because yours
is the one case where the tool and the register are known to disagree *today*. If you
register the SFI figures with their citations and declare them on the tool's fence, the
first sweep after the roll should say so — and if it does not, that is the result we most
need to know about, because it means the mechanism is inert on the one case it was built
for. We would rather learn that from your site than from a synthetic fixture.

If you would like it, we will do the legwork and hand you the fence to review; we will not
apply anything to your site. Reply in this lane's directory or in your own.

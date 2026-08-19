#!/usr/bin/env python3
"""Pick out the domains that are PEOPLE'S NAMES, from a list of domains.

WHAT IT IS FOR. On an estate of a few thousand domains, "firstnamelastname.co.uk"
is a different kind of asset from "loancalculator.co.uk" — different use, different
buyer, different reason to keep it. This separates them so the owner can look at
that group on its own.

HOW IT DECIDES, and why it is stated plainly rather than buried: this is a
HEURISTIC, and the honest output is three buckets, not a yes/no.

  NAME        the label splits cleanly into a known forename + a plausible
              surname ("jamesbrown"), or is forename + initial, or a known
              forename alone on a personal-shaped TLD (.me.uk, .name).
  MAYBE       one half matches a forename but the other is also an ordinary
              English word, so it could be either ("mayfield", "grantham"),
              or the whole label is a single token not in the dictionary.
  NO          contains a clear commercial token (calculator, insurance, loans,
              compare, best, cheap, uk, online…) or splits into dictionary words.

WHY THREE BUCKETS. A two-way split forces every ambiguous case into a wrong
answer, and the ambiguous cases here are numerous and systematic: many English
surnames are also place names and ordinary words (Baker, Fields, Green, Hastings).
MAYBE is where they belong, and it is meant to be eyeballed, not trusted.

RECALL IS LIMITED BY THE FORENAME LIST, and that limit is real: the embedded list
is common UK/Anglophone forenames, so "priyasharma.co.uk" or "olukayode.uk" will
land in MAYBE rather than NAME. Extend FORENAMES for a better result — the list is
data, not logic, and adding to it cannot break the other buckets.

Usage:  extract_person_name_domains.py [file]     # or domains on stdin
        ... --only NAME                           # filter to one bucket
Output: TSV — domain, verdict, reason.
"""
import argparse
import os
import re
import sys

# Common UK/Anglophone forenames. Deliberately data, not logic — extend freely.
FORENAMES = set("""
adam adrian aidan alan albert alex alexander alfie alfred alice amanda amelia amy
andrew angela ann anna anne anthony antony arthur ashley barbara barry ben benjamin
bernard beth bethany bill billy bobby brandon brian bruce bryan callum cameron
carl carol caroline catherine charles charlie charlotte chloe chris christian
christine christopher claire clare colin connor craig daisy dan daniel danny darren
dave david dawn dean deborah debra dennis derek diana diane dominic donald donna
doreen doris dorothy douglas duncan dylan eddie edward eileen elaine eleanor
elizabeth ella ellie emily emma eric erin ethan eugene evan evelyn fiona florence
frances francis frank fred freddie frederick gabriel gareth gary gavin gemma
geoffrey george georgia georgina gerald gerard gillian glen glenn gordon grace
graham grant greg gregory hannah harold harriet harry hayley heather helen henry
holly howard hugh hugo iain ian imogen irene isaac isabel isabella isla ivan jack
jackie jacob jade jake james jamie jane janet janice jason jay jean jeff jeffrey
jenna jennifer jenny jeremy jessica jill jim jo joan joanna joanne jodie joe joel
john johnny jon jonathan jordan joseph josh joshua joyce judith julia julian julie
justin karen kate katherine kathleen kathryn katie kayleigh keith kelly ken kenneth
kevin kieran kim kirsty kyle lauren laura lawrence lee leo leon leonard leslie lewis
liam lily linda lisa liz lorraine louis louise lucas lucy luke lydia lynn madeline
maggie malcolm marc marcus margaret maria marian marie marilyn marion mark martin
mary mason matt matthew maureen max maya megan melanie melissa michael michelle mike
mitchell molly monica morgan nancy naomi natalie natasha nathan neil nick nicholas
nicola nigel noah noel norman oliver olivia oscar owen pam pamela pat patricia
patrick paul paula pauline peter philip phillip phoebe polly poppy rachel ralph
raymond rebecca reece reg reginald rhys richard rita rob robert robin roger roman
ronald rory rosemary ross roy ruby russell ruth ryan sally sam samantha samuel sandra
sarah scott sean sebastian shane shannon sharon sheila shirley sian simon sophie
stanley stephanie stephen steve steven stewart stuart sue susan suzanne sydney sylvia
tanya taylor terence terry theresa thomas tim timothy tina toby tom tommy tony tracey
tracy trevor tyler valerie vanessa vaughan vera veronica vicky victor victoria vincent
violet vivian wayne wendy wesley will william willie yvonne zach zachary zoe
""".split())

# Tokens that mark a domain as commercial/topical rather than personal.
COMMERCIAL = set("""
calculator calculators calc compare comparison quote quotes quotation rate rates
insurance insure loan loans mortgage mortgages credit finance financial bank banking
savings saving invest investment pension broker brokers advice adviser advisor
best cheap cheapest top online direct uk gb british national local my the shop store
buy sell sale price prices cost costs deal deals offer offers free trade wholesale
service services solutions group holdings ltd limited company co plc agency studio
design designs web website websites digital media marketing seo hosting host domain
domains news blog guide guides review reviews info tool tools app apps game games
health medical dental legal law solicitor accountant tax property estate rent
car cars auto motor travel holiday hotel food drink pet vet energy gas electric
""".split())

# Common UK surnames. THIS LIST EXISTS BECAUSE OF A FAILED CONTROL, and the reason
# is worth keeping: on the first run "jamesbrown.co.uk", "sarahjones.uk",
# "davidsmith.me.uk" and "peterhiggins.com" all came out MAYBE — because brown,
# jones, smith and higgins are ordinary entries in the British English dictionary,
# so the "forename + dictionary word" rule fired on the very cases the tool exists
# to catch. Most common English surnames ARE dictionary words (Brown, Green, Baker,
# Fields, Cook, Hill, Wood, Ward, Bell), so the dictionary test cannot be the last
# word. Checked BEFORE the dictionary test for that reason.
SURNAMES = set("""
adams allen anderson armstrong atkinson bailey baker barnes bell bennett berry
birch bishop black booth bradley brooks brown bryant burton butler byrne campbell
carter chapman clark clarke cole collins cook cooper cox craig crawford cunningham
davies davis dawson dean dixon dobson donnelly douglas doyle duncan dunn edwards
elliott ellis evans farrell fisher fletcher ford foster fox francis fraser gallagher
gardner george gibson gill glover goodwin gordon graham grant gray green griffiths
hall hamilton hancock harper harris harrison hart harvey hayes henderson higgins
hill hodgson holland holmes hopkins howard hudson hughes hunt hunter hutchinson
jackson james jenkins johnson johnston jones jordan kelly kennedy kerr king knight
lane lawrence lawson lee lewis little lloyd long lowe lyons macdonald mann marsh
marshall martin mason matthews mccarthy mcdonald mcgrath mclean miller mills
mitchell moore moran morgan morris morrison murphy murray myers nelson newman
nicholson norris obrien oconnor oneill osborne owen page palmer parker parry parsons
patel patterson payne pearce pearson perry peters phillips pope porter potter powell
power price pritchard quinn read reed rees reid reynolds rice richards richardson
riley roberts robertson robinson rogers rose ross russell ryan sanders saunders
scott sharp shaw shepherd short simpson sinclair slater smith spencer stevens
stevenson stewart stone sutton swift taylor thomas thompson thomson todd tucker
turner walker wallace walsh walton ward warren watson watts webb webster welch
wells west wheeler white wilkinson williams williamson willis wilson wood woods
wright young
""".split())

PERSONAL_TLDS = {"me.uk", "name"}


def load_words():
    for p in ("/usr/share/dict/british-english", "/usr/share/dict/words"):
        if os.path.exists(p):
            try:
                with open(p, encoding="utf-8", errors="ignore") as f:
                    return {w.strip().lower() for w in f if w.strip().isalpha()}
            except OSError:
                pass
    return set()


WORDS = load_words()


def split_label(label):
    """Yield (head, tail) splits of a label, longest plausible forename first."""
    for i in range(len(label) - 1, 1, -1):
        yield label[:i], label[i:]


def classify(domain, words):
    d = domain.strip().lower().rstrip(".")
    parts = d.split(".")
    if len(parts) < 2:
        return d, "NO", "not a domain"
    label = parts[0]
    tld = ".".join(parts[1:])
    if not re.fullmatch(r"[a-z0-9-]+", label):
        return d, "NO", "label has unexpected characters"

    clean = label.replace("-", "")
    # An explicit commercial token anywhere settles it.
    for tok in re.split(r"[-]", label):
        if tok in COMMERCIAL:
            return d, "NO", f"commercial token '{tok}'"
    for c in sorted(COMMERCIAL, key=len, reverse=True):
        if len(c) >= 5 and c in clean:
            return d, "NO", f"commercial token '{c}'"

    if clean in FORENAMES:
        if tld in PERSONAL_TLDS:
            return d, "NAME", f"forename '{clean}' on personal TLD .{tld}"
        return d, "MAYBE", f"bare forename '{clean}'"

    for head, tail in split_label(clean):
        if head in FORENAMES:
            if len(tail) == 1:
                return d, "NAME", f"forename '{head}' + initial '{tail}'"
            # Known surname beats the dictionary — see SURNAMES' header.
            if tail in SURNAMES:
                return d, "NAME", f"forename '{head}' + surname '{tail}'"
            if tail in words and tail not in FORENAMES:
                return d, "MAYBE", f"forename '{head}' + dictionary word '{tail}'"
            if len(tail) >= 3:
                return d, "NAME", f"forename '{head}' + surname '{tail}'"
        if head in SURNAMES and tail in FORENAMES:
            return d, "NAME", f"surname '{head}' + forename '{tail}'"

    if clean in words:
        return d, "NO", "single dictionary word"
    # Two dictionary words joined is a compound brand, not a person.
    for head, tail in split_label(clean):
        if head in words and tail in words and len(head) >= 3 and len(tail) >= 3:
            return d, "NO", f"dictionary compound '{head}'+'{tail}'"

    return d, "MAYBE", "no forename matched and not a dictionary word"


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("file", nargs="?")
    ap.add_argument("--only", choices=["NAME", "MAYBE", "NO"])
    args = ap.parse_args()

    src = open(args.file) if args.file else sys.stdin
    rows = []
    for ln in src:
        ln = ln.strip()
        if not ln or ln.startswith("#"):
            continue
        rows.append(classify(ln, WORDS))

    print("domain\tverdict\treason")
    counts = {}
    for d, v, why in rows:
        counts[v] = counts.get(v, 0) + 1
        if args.only and v != args.only:
            continue
        print(f"{d}\t{v}\t{why}")

    if not WORDS:
        print("# WARNING: no system word list found; MAYBE/NO accuracy is degraded",
              file=sys.stderr)
    print("# totals: " + "  ".join(f"{k}={v}" for k, v in sorted(counts.items())),
          file=sys.stderr)


if __name__ == "__main__":
    main()

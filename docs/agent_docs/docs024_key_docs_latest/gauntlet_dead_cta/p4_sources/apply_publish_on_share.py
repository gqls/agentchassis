#!/usr/bin/env python3
"""Splice publish-on-share into the gauntlet JS.

    python3 apply_publish_on_share.py <in.js> <out.js>

WHY A SCRIPT RATHER THAN AN EDITOR. The file stores non-ASCII inside string
literals as a literal `\\uXXXX` escape (six ASCII characters). An editor channel
that decodes escapes cannot emit one: writing the escape yields the CHARACTER,
and writing a doubled backslash yields TWO backslashes. So neither an Edit's
old_string match nor a hand-typed replacement can round-trip. Here the escape is
built from a Python literal, and the script asserts afterwards that the output
still contains the escape and NOT the raw character.

Every replacement asserts it matched exactly once. A silent no-op splice that
still writes an output file is the failure this guards against.
"""
import re
import sys

BULLET = "\\u00B7"        # the six-character escape, as stored in the file

REPLACEMENTS = []


def sub(old, new, note):
    REPLACEMENTS.append((old, new, note))


# ── 1. the card takes an optional permalink ─────────────────────────────────
sub(
    "  function buildVerdictCard() {",
    """  // permalink is the address of the PUBLISHED round, or null when the round
  // was not published (the visitor's press failed, or there was no round id).
  // The card is drawn identically either way except for its last line, so a
  // failed publish still yields a true card — just one without an address.
  function buildVerdictCard(permalink) {""",
    "buildVerdictCard takes a permalink",
)

sub(
    '    x.fillText("vonc.com/tools/gauntlet ' + BULLET + ' " + new Date().toLocaleDateString("en-GB"), L, H - 24);',
    """    // With a permalink this line is the whole point of the card: an image
    // cannot be read by a machine, quoted, or followed, and this address is
    // the only route from the picture back to the argument. Without one, fall
    // back to the tool's own address rather than print a link that would 404.
    x.fillText(
      (permalink ? permalink : "vonc.com/tools/gauntlet") +
        " """ + BULLET + """ " + new Date().toLocaleDateString("en-GB"),
      L, H - 24);""",
    "card footer carries the permalink",
)

# ── 2. sharing publishes ────────────────────────────────────────────────────
sub(
    """  function shareVerdictCard() {
    var c = buildVerdictCard();
    if (!c || !c.toBlob) return;
    c.toBlob(function (blob) {""",
    """  // Sharing PUBLISHES the round (owner decision, 2026-07-31: "the share
  // button publishes", with consent implied by the press — so the button says
  // so plainly, and the note beside it says what becomes public).
  //
  // THE ORDER IS LOAD-BEARING: publish, THEN draw. Drawing first and
  // publishing after would print an address on the card before we knew the
  // address existed, so a failed publish would hand the visitor an image
  // pointing at a 404 — a card that lies about a page. On failure the card is
  // still drawn and still saved, without a link, and the visitor is told.
  var sharing = false;

  // The same three facts buildVerdictCard requires, without drawing anything.
  // Using the card itself as the "is there a round?" test drew a COMPLETE
  // discarded canvas on every press — visible in the fillText log as two
  // address lines, the front door and then the permalink. Harmless output,
  // twice the work, and an ambiguous record of what the card actually says.
  function roundIsComplete() {
    return !!(
      (el.verdict && el.verdict.textContent.trim()) &&
      (el.opponentChallenge && el.opponentChallenge.textContent.trim()) &&
      (el.defenceInput && el.defenceInput.value.trim())
    );
  }

  function shareVerdictCard() {
    if (sharing) return;                 // a second press must not publish twice
    if (!roundIsComplete()) return;      // no complete round: change nothing, claim nothing
    sharing = true;
    busy(el.shareCard, true, "Publishing...");

    function finish(permalink, note, kind) {
      busy(el.shareCard, false);
      sharing = false;
      if (note) setStatus(note, kind);
      emitCard(buildVerdictCard(permalink));
    }

    if (!state.roundId) {
      // A round resumed from storage can outlive its id. The card is still a
      // true record of the exchange; there is simply nothing to publish.
      finish(null,
        "Saved the card. This round could not be published \\u2014 the page no longer has its round id.",
        "error");
      return;
    }

    post("publish", { round_id: state.roundId })
      .then(function (d) {
        if (!d || !d.path) throw new Error("no path in publish response");
        finish("vonc.com" + d.path,
          "Published. The card carries the address of the full debate.",
          "live");
      })
      .catch(function (err) {
        finish(null,
          "Saved the card, but the round was not published, so there is no link on it. " +
            explain(err, "publishing the round"),
          "error");
      });
  }

  function emitCard(c) {
    if (!c || !c.toBlob) return;
    c.toBlob(function (blob) {""",
    "shareVerdictCard publishes first",
)

# ── 3. a 409 has a meaning on this endpoint alone ───────────────────────────
sub(
    """      case 403:
        return "This page is not authorised to reach the debate engine.";""",
    """      case 403:
        return "This page is not authorised to reach the debate engine.";
      case 409:
        return "That round has no verdict yet, so there is nothing to publish.";""",
    "explain() covers 409",
)

# ── 4. the label and the behaviour are set in ONE place ─────────────────────
sub(
    '  if (el.shareCard) el.shareCard.addEventListener("click", shareVerdictCard);',
    """  // The label, the consent note and the handler are attached HERE, together,
  // and that is deliberate. The page HTML and this script are cached
  // separately, so a visitor can hold an old page with a new script. If the
  // label lived in the markup, that visitor would see "Save this verdict as a
  // card" on a button that publishes — consent claimed by a sentence they were
  // never shown. A label written beside the handler cannot disagree with it.
  if (el.shareCard) {
    el.shareCard.textContent = "Publish this round and save the card";
    if (!el.shareCard.parentNode.querySelector("[data-gi-share-note]")) {
      var note = document.createElement("p");
      note.setAttribute("data-gi-share-note", "");
      note.style.cssText =
        "margin:0 0 0.85rem;font-size:0.85rem;line-height:1.55;opacity:0.8;max-width:54ch;";
      note.textContent =
        "Publishing puts this round on a public page at vonc.com: the provocation, " +
        "your position, the challenge and your defence. No name and no account \\u2014 " +
        "the record keeps the argument, not the arguer. The card carries the address.";
      // BEFORE the button: a visitor should read what a control does before
      // they are looking at the control.
      el.shareCard.parentNode.insertBefore(note, el.shareCard);
    }
    el.shareCard.addEventListener("click", shareVerdictCard);
  }""",
    "button label + consent note set beside the handler",
)


def main():
    src, dst = sys.argv[1], sys.argv[2]
    s = open(src).read()
    original = s

    for old, new, note in REPLACEMENTS:
        n = s.count(old)
        if n != 1:
            sys.exit("FAIL [%s]: anchor matched %d times, expected exactly 1" % (note, n))
        s = s.replace(old, new)
        print("ok   " + note)

    # THE CONVENTION IN THIS FILE, established by reading it rather than assumed:
    # comments carry RAW non-ASCII (31 em dashes, 5 middots before this change),
    # string LITERALS carry the six-character \\uXXXX escape. A first draft of
    # this script asserted "no raw character anywhere" and refused its own
    # correct output, because the comments it inserts contain em dashes like
    # every other comment here. The rule that matters is the one about strings:
    # a raw character inside a literal is what an escape-decoding editor
    # produces, and it is invisible until the card renders a mojibake glyph.
    def raw_in_strings(text):
        n = 0
        for lit in re.findall(r'"(?:[^"\\]|\\.)*"', text):
            n += sum(lit.count(ch) for ch in ("—", "·", "’", "“"))
        return n

    assert BULLET in s, "the middot escape was lost from the card footer"
    before, after = raw_in_strings(original), raw_in_strings(s)
    assert after <= before, (
        "raw non-ASCII appeared inside a string literal (%d -> %d) — an escape was decoded"
        % (before, after))
    assert "published \\u2014 the page" in s, "the new status string lost its escape"
    assert s.count("function buildVerdictCard(permalink)") == 1
    assert s.count("function emitCard(") == 1
    assert s.count("function roundIsComplete(") == 1
    assert s != original

    open(dst, "w").write(s)
    print("\nwrote %s  (%d -> %d chars)" % (dst, len(original), len(s)))
    return 0


if __name__ == "__main__":
    sys.exit(main())

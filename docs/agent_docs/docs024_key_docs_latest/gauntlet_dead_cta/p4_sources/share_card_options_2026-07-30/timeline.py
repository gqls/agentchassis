#!/usr/bin/env python3
"""Show each candidate card at the width a timeline actually renders it.

X/Twitter renders a 1.91:1 link card at ~504px wide in-timeline on desktop.
1200 -> 504 is a 2.38x downscale, so type on the card shrinks by that factor.
This composite applies exactly that downscale and nothing else, so what you
see is what a stranger scrolling past would see.
"""
from PIL import Image, ImageDraw, ImageFont

W = 504                      # X in-timeline card width
CARDS = [
    ("mock0.png",            "TODAY — verdict only", "13 characters of argument"),
    ("mock1_whole_debate.png", "OPTION 1 — whole debate", "fits at 11px -> 4.6px here"),
    ("mock1b.png",           "OPTION 1 — the exchange", "26px -> 11px; 18% of the round"),
    ("mock2b.png",           "OPTION 2 — debate card", "16px -> 6.7px; tap to read"),
    ("mock3a.png",           "OPTION 3 — the hook", "30px -> 13px; links to all of it"),
]

def font(sz, bold=False):
    base = "/usr/share/fonts/truetype/dejavu/DejaVuSans%s.ttf" % ("-Bold" if bold else "")
    try:
        return ImageFont.truetype(base, sz)
    except OSError:
        return ImageFont.load_default()

CARD_H = round(630 * W / 1200)
LABEL, SUB, GAP, PADX, TOP = 26, 22, 34, 28, 30
COLS = 2
rows = (len(CARDS) + COLS - 1) // COLS
cw = W + PADX
canvas_w = COLS * cw + PADX
canvas_h = TOP + rows * (LABEL + SUB + CARD_H + GAP) + 10

im = Image.new("RGB", (canvas_w, canvas_h), (10, 10, 15))
d = ImageDraw.Draw(im)
f_lab, f_sub = font(15, True), font(13)

for i, (src, label, sub) in enumerate(CARDS):
    col, row = i % COLS, i // COLS
    x = PADX + col * cw
    y = TOP + row * (LABEL + SUB + CARD_H + GAP)
    d.text((x, y), label, font=f_lab, fill=(245, 158, 11))
    d.text((x, y + LABEL - 4), sub, font=f_sub, fill=(139, 133, 176))
    card = Image.open(src).convert("RGB").resize((W, CARD_H), Image.LANCZOS)
    im.paste(card, (x, y + LABEL + SUB))
    d.rectangle([x, y + LABEL + SUB, x + W - 1, y + LABEL + SUB + CARD_H - 1],
                outline=(42, 38, 64))

im.save("timeline_comparison.png")
print("timeline_comparison.png", im.size)

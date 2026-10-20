"""Bilge icon generator.

Concept: ship cross-section — containers stacked on deck, hull below,
and "bilge water" (the junk that accumulates) pooled at the bottom.

Drawn at 4x and downsampled (LANCZOS) for clean anti-aliasing,
since Pillow's primitives aren't anti-aliased on their own.
"""
from pathlib import Path

from PIL import Image, ImageDraw

BASE = 512          # logical design grid
SS = 4              # supersampling factor
W = BASE * SS

# palette
BG_TOP = (21, 38, 58)
BG_BOT = (10, 20, 32)
HULL_FILL = (27, 47, 70)
BRASS = (232, 176, 74)
WATER = (46, 196, 182)
WATER_HI = (130, 232, 220)
CONTAINERS = [(232, 176, 74), (46, 196, 182), (94, 129, 172), (214, 96, 77), (232, 176, 74)]
RIB = (0, 0, 0, 55)


def s(v: float) -> int:
    return round(v * SS)


def quad_bezier(p0, p1, p2, n=64):
    return [
        (
            (1 - t) ** 2 * p0[0] + 2 * (1 - t) * t * p1[0] + t**2 * p2[0],
            (1 - t) ** 2 * p0[1] + 2 * (1 - t) * t * p1[1] + t**2 * p2[1],
        )
        for t in (i / n for i in range(n + 1))
    ]


def hull_points():
    # deck line -> right side -> curved keel -> left side
    pts = [(84, 262), (428, 262), (394, 390)]
    pts += quad_bezier((394, 390), (256, 446), (118, 390))[1:]
    return [(s(x), s(y)) for x, y in pts]


def wave_points(y, amp, period, x0=60, x1=452):
    import math
    pts = []
    x = x0
    while x <= x1:
        pts.append((x, y + amp * math.sin((x - x0) / period * 2 * math.pi)))
        x += 2
    return pts


def container(draw, x, y, w, h, color):
    draw.rounded_rectangle([s(x), s(y), s(x + w), s(y + h)], radius=s(6), fill=color)
    # corrugation ribs
    ribs = Image.new("RGBA", (W, W), (0, 0, 0, 0))
    rd = ImageDraw.Draw(ribs)
    step = w / 6
    for i in range(1, 6):
        rx = x + i * step
        rd.line([(s(rx), s(y + 9)), (s(rx), s(y + h - 9))], fill=RIB, width=s(4))
    return ribs


def build() -> Image.Image:
    # background: vertical gradient clipped to rounded square
    grad = Image.new("RGBA", (W, W))
    gd = ImageDraw.Draw(grad)
    for yy in range(W):
        t = yy / (W - 1)
        c = tuple(round(a + (b - a) * t) for a, b in zip(BG_TOP, BG_BOT))
        gd.line([(0, yy), (W, yy)], fill=c + (255,))
    mask = Image.new("L", (W, W), 0)
    ImageDraw.Draw(mask).rounded_rectangle([0, 0, W - 1, W - 1], radius=s(112), fill=255)
    img = Image.new("RGBA", (W, W), (0, 0, 0, 0))
    img.paste(grad, (0, 0), mask)

    d = ImageDraw.Draw(img)

    # containers on deck: 3 bottom, 2 top (offset like a real stack)
    cw, ch, gap = 92, 62, 10
    # deck line is y=262, hull stroke is 16 wide -> visible top edge at 254;
    # leave the same `gap` above it as between the two container rows
    bottom_y = 262 - 8 - gap - ch
    top_y = bottom_y - gap - ch
    row_w = 3 * cw + 2 * gap
    bx0 = 256 - row_w / 2
    layers = []
    for i in range(3):
        layers.append(container(d, bx0 + i * (cw + gap), bottom_y, cw, ch, CONTAINERS[i]))
    tx0 = 256 - (2 * cw + gap) / 2
    for i in range(2):
        layers.append(container(d, tx0 + i * (cw + gap), top_y, cw, ch, CONTAINERS[3 + i]))
    for layer in layers:
        img.alpha_composite(layer)

    # hull body
    hull = hull_points()
    hull_mask = Image.new("L", (W, W), 0)
    ImageDraw.Draw(hull_mask).polygon(hull, fill=255)
    hull_layer = Image.new("RGBA", (W, W), HULL_FILL + (255,))
    img.paste(hull_layer, (0, 0), hull_mask)

    # bilge water, clipped to hull interior
    water = Image.new("RGBA", (W, W), (0, 0, 0, 0))
    wd = ImageDraw.Draw(water)
    wave = wave_points(360, 7, 86)
    poly = [(s(x), s(y)) for x, y in wave] + [(s(452), s(460)), (s(60), s(460))]
    wd.polygon(poly, fill=WATER + (255,))
    wd.line([(s(x), s(y)) for x, y in wave], fill=WATER_HI + (255,), width=s(6), joint="curve")
    # a couple of bubbles = "junk" floating
    for cx, cy, r in [(214, 332, 7), (292, 322, 5), (322, 338, 4)]:
        wd.ellipse([s(cx - r), s(cy - r), s(cx + r), s(cy + r)], outline=WATER_HI + (255,), width=s(3))
    img.paste(water, (0, 0), Image.composite(water.getchannel("A"), Image.new("L", (W, W), 0), hull_mask))

    # hull outline (brass), drawn last so it sits over water edges
    d = ImageDraw.Draw(img)
    d.line(hull + [hull[0]], fill=BRASS, width=s(16), joint="curve")
    # round the deck corners
    for x, y in [(84, 262), (428, 262)]:
        d.ellipse([s(x - 8), s(y - 8), s(x + 8), s(y + 8)], fill=BRASS)

    return img.resize((BASE * 2, BASE * 2), Image.LANCZOS)  # 1024 master


# Outputs, relative to the repository root. Wails builds the macOS .icns from
# appicon.png and embeds icon.ico on Windows; main.go embeds appicon.png as
# the Linux window icon.
ROOT = Path(__file__).resolve().parent.parent
APPICON = ROOT / "build" / "appicon.png"
WINDOWS_ICO = ROOT / "build" / "windows" / "icon.ico"
UI_ICON = ROOT / "frontend" / "src" / "assets" / "icon.png"


def main():
    master = build()
    master.save(APPICON)
    master.save(WINDOWS_ICO, sizes=[(n, n) for n in (16, 24, 32, 48, 64, 128, 256)])
    UI_ICON.parent.mkdir(parents=True, exist_ok=True)
    # Shown at 24 px in the sidebar; 64 px stays sharp on 2x displays.
    master.resize((64, 64), Image.LANCZOS).save(UI_ICON)
    for path in (APPICON, WINDOWS_ICO, UI_ICON):
        print("wrote", path.relative_to(ROOT))


if __name__ == "__main__":
    main()

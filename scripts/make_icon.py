"""Bilge icon generator.

Concept: a ship's hull with one container on deck, drawn flat in the app's
own dark palette so the icon matches the UI.

Drawn at 4x and downsampled (LANCZOS) for clean anti-aliasing, since
Pillow's primitives aren't anti-aliased on their own.
"""

from pathlib import Path

from PIL import Image, ImageDraw

BASE = 512  # logical design grid
SS = 4  # supersampling factor
W = BASE * SS

# Colors follow the dark theme in frontend/src/style.css; keep them in sync.
PANEL = (30, 33, 35)  # --panel
LINE = (46, 50, 53)  # --line
FG = (228, 230, 231)  # --fg
ACCENT = (92, 199, 196)  # --accent

# Outputs, relative to the repository root. Wails builds the macOS .icns from
# appicon.png and embeds icon.ico on Windows; main.go embeds appicon.png as
# the Linux window icon, and the Linux archive ships LINUX_ICON next to the
# .desktop file for the application menu.
ROOT = Path(__file__).resolve().parent.parent
APPICON = ROOT / "build" / "appicon.png"
WINDOWS_ICO = ROOT / "build" / "windows" / "icon.ico"
UI_ICON = ROOT / "frontend" / "src" / "assets" / "icon.png"
LINUX_ICON = ROOT / "build" / "linux" / "bilge.png"


def s(v: float) -> int:
    return round(v * SS)


def quad_bezier(p0, p1, p2, n=80):
    return [
        (
            (1 - t) ** 2 * p0[0] + 2 * (1 - t) * t * p1[0] + t**2 * p2[0],
            (1 - t) ** 2 * p0[1] + 2 * (1 - t) * t * p1[1] + t**2 * p2[1],
        )
        for t in (i / n for i in range(n + 1))
    ]


def hull_points():
    # deck line -> right side -> curved keel -> left side
    pts = [(100, 214), (412, 214), (362, 352)]
    pts += quad_bezier((362, 352), (256, 418), (150, 352))[1:]
    return pts


def content() -> Image.Image:
    img = Image.new("RGBA", (W, W), (0, 0, 0, 0))
    d = ImageDraw.Draw(img)
    hull = [(s(x), s(y)) for x, y in hull_points()]
    d.polygon(hull, fill=ACCENT)
    # A stroke in the same color, with round dots at the deck corners, softens
    # the polygon's sharp edges.
    d.line(hull + [hull[0]], fill=ACCENT, width=s(24), joint="curve")
    for x, y in hull_points()[:2]:
        d.ellipse([s(x - 12), s(y - 12), s(x + 12), s(y + 12)], fill=ACCENT)
    d.rounded_rectangle([s(180), s(128), s(332), s(196)], radius=s(10), fill=FG)
    return img


def build() -> Image.Image:
    img = Image.new("RGBA", (W, W), (0, 0, 0, 0))
    ImageDraw.Draw(img).rounded_rectangle(
        [0, 0, W - 1, W - 1], radius=s(112), fill=PANEL, outline=LINE, width=s(4)
    )
    drawing = content()
    # Centre the drawing vertically, whatever its exact coordinates.
    _, top, _, bottom = drawing.getbbox()
    img.alpha_composite(drawing, (0, (W - (top + bottom)) // 2))
    return img.resize((BASE * 2, BASE * 2), Image.LANCZOS)  # 1024 master


def main():
    master = build()
    master.save(APPICON)
    master.save(WINDOWS_ICO, sizes=[(n, n) for n in (16, 24, 32, 48, 64, 128, 256)])
    UI_ICON.parent.mkdir(parents=True, exist_ok=True)
    # Shown at 24 px in the sidebar; 64 px stays sharp on 2x displays.
    master.resize((64, 64), Image.LANCZOS).save(UI_ICON)
    LINUX_ICON.parent.mkdir(parents=True, exist_ok=True)
    # 256 px is the largest size freedesktop icon themes require.
    master.resize((256, 256), Image.LANCZOS).save(LINUX_ICON)
    for path in (APPICON, WINDOWS_ICO, UI_ICON, LINUX_ICON):
        print("wrote", path.relative_to(ROOT))


if __name__ == "__main__":
    main()

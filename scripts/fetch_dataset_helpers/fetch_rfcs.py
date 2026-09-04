#!/usr/bin/env python3
# /// script
# requires-python = ">=3.11"
# ///
import json
import re
from pathlib import Path

HELPERS_DIR = Path(__file__).resolve().parent
MIRROR_DIR = HELPERS_DIR / "my-rfc-mirror"
OUTPUT_FILE = HELPERS_DIR.parent / "rfcs.json"
DOMAIN = "www.rfc-editor.org"
BASE_URL = "https://www.rfc-editor.org/rfc"
RFC_HTML_RE = re.compile(r"^rfc(\d+)\.html$")

ROOT_SUBDIR = ""
SUBDIRS = (ROOT_SUBDIR, "inline-errata")


def read_text(path: Path) -> str | None:
    if not path.is_file():
        return None
    data = path.read_bytes()
    try:
        return data.decode("utf-8")
    except UnicodeDecodeError:
        return data.decode("latin-1")


def discover_html(subdir: str) -> list[tuple[str, int]]:
    folder = MIRROR_DIR if subdir == ROOT_SUBDIR else MIRROR_DIR / subdir
    if not folder.is_dir():
        return []
    found: list[tuple[str, int]] = []
    for path in folder.iterdir():
        if not path.is_file():
            continue
        match = RFC_HTML_RE.match(path.name)
        if match:
            found.append((subdir, int(match.group(1))))
    return found


def build_entry(subdir: str, number: int) -> dict[str, str]:
    stem = f"rfc{number}"
    if subdir == ROOT_SUBDIR:
        html_path = MIRROR_DIR / f"{stem}.html"
        url = f"{BASE_URL}/{stem}.html"
    else:
        html_path = MIRROR_DIR / subdir / f"{stem}.html"
        url = f"{BASE_URL}/{subdir}/{stem}.html"

    html = read_text(html_path) or ""

    title = ""
    json_text = read_text(MIRROR_DIR / f"{stem}.json")
    if json_text:
        try:
            metadata = json.loads(json_text)
        except json.JSONDecodeError:
            metadata = {}
        raw_title = metadata.get("title")
        if isinstance(raw_title, str):
            title = raw_title

    text = read_text(MIRROR_DIR / f"{stem}.txt") or ""

    return {
        "url": url,
        "domain": DOMAIN,
        "html": html,
        "title": title.strip(),
        "text": text,
        "language": "en",
        "label": "rfc",
        "skip_sensitive_check": True,
        "favicon": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAYAAABzenr0AAAACXBIWXMAAAsTAAALEwEAmpwYAAADlElEQVR42u2W2U8TURTGK0r0XzAmbiFqfDGRB0namSrFJYrEEKtJEQUFxdiyCKEohZaKQJRFZWkLtkOVB6wSy4AoATuVliUqpoiiILgT0Vij0A6USq+3KHvbKZtPPcl5mdzM7zvfOffM0Gie8IQn5hFpEsUlsQRTyWQy7/8Oh+CLYikG7JkmwSqEQuGy/1e5VHF2DD4uQqpQQhFeiw6/IMV40+ETIjA5AGDJosFFEiwS9t3mTMDfVKQtFjwEAkZcw0cFCAmCWLGwAyfDDsCXW6ng0J2Cqoe6LZWE/kelVp+4UFdtJ+ztEGXlEqwUJ4hNEPwF1+qBPSu1uvh5wjE6FGByw/Y7d+rq1sPKP4zBRwUQOhsUcXJO8PQizA8K6KeGY+r7RPNanNC/nQwfT0I3gmt1nFkuGfkWaKmRuudYXVl19RpYeadD+ISIYZxo3OQevOj6RvjiPjfgjUocXw0tbncJH22FvmZ3VFx6aJIYp7HZS53ChQUlPrDyXopJtw9ci0RVtQpa+4QKDlO/Egnc9byr2zpstYIIUZYeomYuq/jL19bBYXrnCp5aKAcpYnb/Gf7eVGhrI3Xluibmca5ArWkwB0Xza40/f9ksw8MgTJChmQI/FMkMyMrc+zUpO8cpXFgkB+KMw/2WVhT01DPN+ViOjcL2VuYJblzX+09WuJ5Bs+HFIDte0DBgJoGJJAFyNGpiY4aepgcbcNRYodg3yM/JnQEXQdtTL3IGfjajNl48oiyXMp511u4gC0qvOhGgex1wMi7c0NltAZPi0VMDyUkUtzQbXlo2s4/4T3EhjEsP7qhBjeXXA4eScq9M6XmiKNTc34KCc+eRG/CoF5Pt56MqQQztNVCEMn86/M1+bgL9cXvHEHAQUpX6h7cvstXhEIadoR/reID+KpXut5zLyx+FJ6eHk3Z4Ah+5NXmCWYdZG+7dQNra7rHIwptF/66b/uNduBMOxiVXOIL3fOr97RcSFevyGkZykVNv61FTSWGQJSEtbNBue0oK/TaNyZzxwxEY4re1tgzpfKwOMEluFn+vJJp87M+5GXmN0+Gf+76NbD/BS3VrF5yKRWM+aFBy8AkKxCKGmubr6/SXKyicsU1Tjva0VTG/qYoZrfZML8wjJ8P7vhtHWBGx2bPahqdjGHyBAK2m+exZTnU2+DiCSvIY1Yr8v8lL5rwuKFOaM0uUvVnyslf+EdHCuX6PvGie8MQixx+Y1ukh9tAD8AAAAABJRU5ErkJggg==",
    }


def main() -> None:
    entries: list[tuple[str, int]] = []
    for subdir in SUBDIRS:
        entries.extend(discover_html(subdir))
    entries.sort(key=lambda item: (item[0], item[1]))

    root_count = 0
    inline_count = 0
    missing_title = 0
    missing_text = 0

    with OUTPUT_FILE.open("w", encoding="utf-8") as out:
        out.write("[")
        for index, (subdir, number) in enumerate(entries):
            entry = build_entry(subdir, number)
            if subdir == "inline-errata":
                inline_count += 1
            else:
                root_count += 1
            if not entry["title"]:
                missing_title += 1
            if not entry["text"]:
                missing_text += 1
            if index == 0:
                out.write("\n")
            else:
                out.write("\n,\n")
            out.write(json.dumps(entry, ensure_ascii=False, separators=(",", ":")))
        out.write("\n]\n")

    total = root_count + inline_count
    print(f"Wrote {total} entries to {OUTPUT_FILE}")
    print(f"  root:           {root_count}")
    print(f"  inline-errata:  {inline_count}")
    print(f"  missing title:  {missing_title}")
    print(f"  missing text:   {missing_text}")


if __name__ == "__main__":
    main()

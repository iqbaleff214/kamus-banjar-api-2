"""
Extract dictionary entries from the Banjar Hulu dictionary PDF text.

Source: Kamus Bahasa Banjar Dialek Hulu-Indonesia, Edisi Pertama
Publisher: Balai Bahasa Banjarmasin, Departemen Pendidikan Nasional, 2008
ISBN: 978-979-685-776-0

Usage:
    python3 extract_dictionary.py [--pdf PATH] [--out PATH]

Defaults:
    --pdf  docs/kamus-bahasa-banjar-dialek-hulu.pdf
    --out  scripts/seed/seed_data.json

Requires: pdfminer.six
    pip install pdfminer.six
"""

from __future__ import annotations

import argparse
import json
import re
import sys
from pathlib import Path
from typing import Optional


# ---------------------------------------------------------------------------
# Word classes as defined in the source dictionary (section 2.1)
# ---------------------------------------------------------------------------
WORD_CLASSES = {
    "n": "nomina",
    "v": "verba",
    "a": "adjektiva",
    "adv": "adverbia",
    "p": "partikel",
    "pb": "pribahasa",
    "ki": "kiasan",
    "num": "numeralia",
    "prp": "pronomina",
}

# Prefixes that indicate derived/sub-entries (not root words)
DERIVED_PREFIXES = re.compile(
    r"^(ba|ma|man|bar|mar|tar|ta|ka|sa|pa|pan|pang|mam|bam|pam|"
    r"baka|maka|taka|saka|paka|bapa|mapa|tapa|sapa)\."
)

# Pattern to detect a dictionary entry line:
# word  [word_class]  definition...
# word may contain dots (syllable markers) and hyphens
ENTRY_PATTERN = re.compile(
    r"^([a-zA-Z][a-zA-Z.\- ]{0,50}?)"          # word (with possible syllable dots)
    r"\s{2,}"                                     # separator (2+ spaces in PDF)
    r"(n|v|a|adv|p|pb|ki|num|prp)\s+"           # word class
    r"(.+)",                                      # rest = definition + examples
    re.DOTALL,
)

# Homonym marker (e.g. ¹word or 1word or 2word at start)
HOMONYM_MARKER = re.compile(r"^[1-9']")

# Noise patterns from PDF extraction
NOISE_LINES = re.compile(
    r"^("
    r"\x0c"                                      # form feed
    r"|\d{1,3}\s*$"                              # lone page numbers
    r"|.*Bahasa [Bb]anjarmasin.*"
    r"|.*[Bb]alai [Bb]ahasa.*"
    r"|.*[Kk]amus.*[Dd]ialek.*"
    r"|.*\(cid:\d+\).*"                          # PDF encoding artefact
    r"|.*\$.*\[.*\].*"
    r")",
    re.IGNORECASE,
)

# Line that is clearly a page header/footer artefact
ARTEFACT = re.compile(r"[^\x20-\x7eÀ-ɏĀ-ſ]+")


def extract_text_from_pdf(pdf_path: Path) -> str:
    try:
        from pdfminer.high_level import extract_text as _extract
        return _extract(str(pdf_path))
    except ImportError:
        sys.exit("pdfminer.six not installed. Run: pip install pdfminer.six")


def clean_text(text: str) -> str:
    """Remove PDF artefacts and normalize whitespace."""
    text = text.replace("\x0c", "\n")
    # Remove (cid:N) artefacts
    text = re.sub(r"\(cid:\d+\)", " ", text)
    return text


def find_entry_section(lines: list[str]) -> int:
    """Return the line index where actual dictionary entries begin."""
    # The entries start after the TOC/intro. Look for 'aba.dan' or similar
    # Heuristic: first line matching the entry pattern after line 1000
    for i, line in enumerate(lines[800:], start=800):
        if ENTRY_PATTERN.match(line.strip()):
            return i
    return 1091  # fallback


def is_noise(line: str) -> bool:
    stripped = line.strip()
    if not stripped:
        return True
    if NOISE_LINES.match(stripped):
        return True
    # Lines that are clearly OCR/encoding garbage (too many non-ASCII)
    non_ascii = sum(1 for c in stripped if ord(c) > 127)
    if len(stripped) > 0 and non_ascii / len(stripped) > 0.4:
        return True
    return False


def join_blocks(lines: list[str], start: int) -> list[str]:
    """
    Join continuation lines into single logical entry strings.

    A new entry starts when a line (after stripping) matches the entry
    pattern OR starts a known derived form pattern.
    """
    blocks: list[str] = []
    current: list[str] = []

    def flush():
        if current:
            blocks.append(" ".join(current))
            current.clear()

    for line in lines[start:]:
        if is_noise(line):
            continue
        stripped = line.strip()
        # Detect start of a new entry
        if ENTRY_PATTERN.match(stripped):
            flush()
            current.append(stripped)
        elif current:
            current.append(stripped)
        else:
            # Pre-entry noise, skip
            pass

    flush()
    return blocks


def parse_block(block: str) -> Optional[dict]:
    """Parse a single entry block into a structured dict."""
    m = ENTRY_PATTERN.match(block)
    if not m:
        return None

    raw_word = m.group(1).strip()
    word_class_abbr = m.group(2).strip()
    definition_raw = m.group(3).strip()

    # Strip leading homonym marker (1, 2, ', etc.)
    homonym_num = 1
    if HOMONYM_MARKER.match(raw_word):
        c = raw_word[0]
        if c.isdigit():
            homonym_num = int(c)
        raw_word = raw_word[1:].strip()

    # Normalize syllable dots to clean banjar word
    banjar_word = raw_word.replace(".", "").strip()
    # Lowercase, strip trailing punctuation
    banjar_word = banjar_word.strip("- ").lower()

    if not banjar_word or len(banjar_word) < 2:
        return None

    # Is this a derived/sub-entry or a root word?
    is_derived = bool(DERIVED_PREFIXES.match(raw_word.lower()))

    # Split definition and examples on ':'
    # Format: "definition: example_banjar, translation;"
    definitions = []
    examples = []

    # Multiple definitions separated by numbered markers "1 ... 2 ..."
    def_text, _, ex_text = definition_raw.partition(":")
    def_text = def_text.strip()
    ex_text = ex_text.strip().rstrip(";").strip()

    # Split on "1 " "2 " "3 " for multiple definitions
    numbered = re.split(r"\b([1-9])\s+", def_text)
    if len(numbered) > 1:
        # Has numbered definitions
        idx = 1
        while idx < len(numbered):
            num_label = numbered[idx]
            content = numbered[idx + 1].strip() if idx + 1 < len(numbered) else ""
            if content:
                definitions.append(content.rstrip(";").strip())
            idx += 2
    else:
        raw_def = def_text.rstrip(";").strip()
        if raw_def:
            definitions.append(raw_def)

    # Parse example if present
    if ex_text:
        # Format: banjar_example, indonesian_translation
        parts = ex_text.split(",", 1)
        if len(parts) == 2:
            ex_banjar = parts[0].strip()
            ex_id = parts[1].strip().rstrip(";").strip()
            if ex_banjar and ex_id:
                # Replace -- with the actual word
                ex_banjar = ex_banjar.replace("--", banjar_word)
                examples.append({"banjar": ex_banjar, "indonesian": ex_id})

    return {
        "banjar": banjar_word,
        "banjar_syllabified": raw_word,  # with syllable dots
        "word_class": word_class_abbr,
        "word_class_full": WORD_CLASSES.get(word_class_abbr, word_class_abbr),
        "definitions": definitions,
        "examples": examples,
        "is_derived": is_derived,
        "homonym_number": homonym_num,
        "dialect": "hulu",
        "source": "seeded",
        "source_reference": "Kamus Bahasa Banjar Dialek Hulu-Indonesia, Edisi Pertama (Balai Bahasa Banjarmasin, 2008)",
    }


def group_entries(entries: list[dict]) -> list[dict]:
    """
    Group derived entries under their root word.
    Returns only root entries, with derived forms nested under 'derived_forms'.
    """
    root_map: dict[str, dict] = {}
    derived: list[dict] = []

    for entry in entries:
        if not entry:
            continue
        if entry["is_derived"]:
            derived.append(entry)
        else:
            key = entry["banjar"]
            if key in root_map:
                # Homonym — append number suffix
                key = f"{key}_{entry['homonym_number']}"
            root_map[key] = entry
            entry["derived_forms"] = []

    # Attach derived forms to their root
    for d in derived:
        # Try to find root by stripping prefix
        word = d["banjar"]
        matched = False
        for root_key, root_entry in root_map.items():
            root_word = root_entry["banjar"]
            if word.endswith(root_word) and len(word) > len(root_word):
                root_entry["derived_forms"].append(d)
                matched = True
                break
        if not matched:
            # Attach as standalone entry
            root_map[word + "_derived"] = d
            d["derived_forms"] = []

    return list(root_map.values())


def main():
    parser = argparse.ArgumentParser(description="Extract Banjar dictionary seed data")
    parser.add_argument(
        "--pdf",
        default="docs/kamus-bahasa-banjar-dialek-hulu.pdf",
        help="Path to PDF file",
    )
    parser.add_argument(
        "--out",
        default="scripts/seed/seed_data.json",
        help="Output JSON file path",
    )
    parser.add_argument(
        "--flat",
        action="store_true",
        help="Output flat list (no grouping by root word)",
    )
    args = parser.parse_args()

    pdf_path = Path(args.pdf)
    if not pdf_path.exists():
        sys.exit(f"PDF not found: {pdf_path}")

    print(f"Extracting text from {pdf_path}...")
    raw_text = extract_text_from_pdf(pdf_path)
    raw_text = clean_text(raw_text)
    lines = raw_text.split("\n")
    print(f"  {len(lines)} lines extracted")

    print("Finding entry section...")
    entry_start = find_entry_section(lines)
    print(f"  Entries start at line {entry_start}")

    print("Joining entry blocks...")
    blocks = join_blocks(lines, entry_start)
    print(f"  {len(blocks)} blocks found")

    print("Parsing entries...")
    entries = []
    skipped = 0
    for block in blocks:
        entry = parse_block(block)
        if entry:
            entries.append(entry)
        else:
            skipped += 1

    print(f"  {len(entries)} entries parsed ({skipped} skipped)")

    if not args.flat:
        print("Grouping derived forms under root words...")
        entries = group_entries(entries)
        root_count = sum(1 for e in entries if not e.get("is_derived"))
        print(f"  {root_count} root entries with derived forms grouped")

    out_path = Path(args.out)
    out_path.parent.mkdir(parents=True, exist_ok=True)
    with open(out_path, "w", encoding="utf-8") as f:
        json.dump(
            {
                "meta": {
                    "source": "Kamus Bahasa Banjar Dialek Hulu-Indonesia",
                    "edition": "Edisi Pertama",
                    "publisher": "Balai Bahasa Banjarmasin, Departemen Pendidikan Nasional",
                    "year": 2008,
                    "isbn": "978-979-685-776-0",
                    "dialect": "hulu",
                    "letters_covered": list("ABCDGHIJKLMNPRSTUWV"),
                    "letters_not_in_dialect": ["E", "F", "O", "Q", "V", "Z"],
                    "total_entries": len(entries),
                },
                "entries": entries,
            },
            f,
            ensure_ascii=False,
            indent=2,
        )

    print(f"\nDone. Seed data written to {out_path}")
    print(
        "Note: Review seed_data.json before importing — PDF extraction may have "
        "OCR artefacts. Manual review of flagged entries is recommended."
    )


if __name__ == "__main__":
    main()

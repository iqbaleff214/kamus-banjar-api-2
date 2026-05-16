# Dictionary Specification
# Kamus Bahasa Banjar Dialek Hulu-Indonesia

This document describes the structure of the source dictionary, the entry format,
word class definitions, and notes on the automated seed data extraction.

---

## 1. Source Reference

| Property | Value |
|---|---|
| Title | Kamus Bahasa Banjar Dialek Hulu-Indonesia |
| Edition | Edisi Pertama |
| Publisher | Balai Bahasa Banjarmasin, Departemen Pendidikan Nasional |
| Year | 2008 |
| ISBN | 978-979-685-776-0 |
| Authors | Musdalipah, Siti Akbari, Jandiah, Wandanie Rakhman, Muhammad Yamani, H. Dede Hidayatullah, Noor Hastiah |
| File | `docs/kamus-bahasa-banjar-dialek-hulu.pdf` |

---

## 2. Dialect

The dictionary covers **Bahasa Banjar Dialek Hulu (BBDH)**, one of two Banjar dialects.

- **Dialek Hulu**: spoken in inland South Kalimantan — Hulu Sungai Selatan, Hulu Sungai Tengah, Hulu Sungai Utara, Tapin, Balangan, Tabalong
- **Dialek Kuala**: spoken in coastal/urban areas (Banjarmasin and surroundings) — out of scope for this dataset

---

## 3. Alphabet & Spelling

BBDH uses a subset of the Latin alphabet. Letters not present in the dialect:

| Letter | Mapped To | Note |
|---|---|---|
| E | I or A | — |
| F | P | — |
| O | U | — |
| Q | K | — |
| V | P | — |
| Z | S or J | — |

Letters covered (dictionary sections):
**A B C D G H I J K L M N P R S T U W Y**

Spelling follows *Ejaan yang Disempurnakan* with exceptions noted in the book.

---

## 4. Word Classes

Defined in section 2.1 of the source dictionary:

| Abbreviation | Indonesian Name | English Equivalent |
|---|---|---|
| `n` | nomina | noun |
| `v` | verba | verb |
| `a` | adjektiva | adjective |
| `adv` | adverbia | adverb |
| `p` | partikel | particle / interjection |
| `pb` | pribahasa | proverb |
| `ki` | kiasan | figurative / idiomatic expression |

---

## 5. Common Abbreviations in Definitions

Used in the source dictionary (section 2.1):

| Abbreviation | Meaning |
|---|---|
| `sdh` | sudah (already) |
| `dg` | dengan (with) |
| `dr` | dari (from) |
| `tdk` | tidak (not/no) |
| `kpd` | kepada (to) |
| `bg` | bagi (for) |
| `dim` | dalam (in/inside) |
| `pd` | pada (at/on) |
| `km` | karena (because) |
| `org` | orang (person) |
| `utk` | untuk (for) |
| `yg` | yang (that/which) |
| `mis` | misalnya (for example) |
| `spt` | seperti (like/such as) |
| `peny` | penyakit (disease) |
| `sej` | sejenis (a type of) |
| `nm` | nama (name of) |
| `pb` | pribahasa (proverb) |
| `ki` | kiasan (figurative) |
| `dst` | dan seterusnya (and so on) |
| `dll` | dan lain-lain (et cetera) |
| `dsb` | dan sebagainya (etc.) |

---

## 6. Entry Format

### 6.1 Main Entry (Root Word)

```
banjar_word  [word_class]  definition: example_sentence, translation;
```

**Example:**
```
abah  n  1 ayah; 2 mertua laki-laki
abut  a  ribut; ramai; sibuk: jangan tapi --, jangan terlalu abut;
```

### 6.2 Sub-Entries (Derived Forms)

Derived forms use the syllabified word with dots and follow the root entry:

```
ba.word  v  definition: example, translation;
ma.word  v  definition: example, translation;
```

**Example:**
```
ba.a.bah  v  berayah; menyebut ayah: inya kada, dia tdk berayah
```

Common prefixes:

| Prefix | Function |
|---|---|
| `ba-` | stative / has-the-quality-of / reciprocal |
| `ma-` / `man-` / `mam-` / `mar-` / `mang-` | active verb (meN- equivalent) |
| `ka-` | abstract noun / superlative |
| `ta-` | passive / accidental / comparative |
| `sa-` | one / per / as-much-as |
| `pa-` / `pang-` / `pam-` | agentive noun / superlative |

### 6.3 Homonyms

Multiple words with same spelling are numbered with superscript:

```
¹amar  n  mahkota pengantin
²amar  p  mar; awas (catur)
```

In extracted text these appear as `1amar`, `2amar` or `'amar`.

### 6.4 Multiple Definitions

Numbered inline within the definition field:

```
abah  n  1 ayah; 2 mertua laki-laki
```

### 6.5 Examples

Format: `banjar_example, indonesian_translation`

The symbol `--` in examples stands for the entry word itself:
```
abut  a  ribut; ramai; sibuk: jangan tapi --, jangan terlalu abut;
→ example banjar: "jangan tapi abut"
→ example indonesian: "jangan terlalu abut"
```

### 6.6 Proverbs

Proverbs (pb) appear as sub-entries or standalone entries:
```
abut: siang jadi macan, malam jadi --, pb siang beringasan, malam penakut
```

---

## 7. Dictionary Scale

Derived from automated extraction of the PDF source:

| Metric | Value |
|---|---|
| Total extracted blocks | ~7,059 |
| Root word entries | ~2,198 |
| Total entries (root + derived) | ~5,436 |
| Dictionary pages | 310 |
| Letters covered | 19 letters (A–Y minus E F O Q V Z) |

---

## 8. Seeder Data Extraction

### 8.1 Extraction Script

```bash
# Install dependency
pip install pdfminer.six

# Run from project root
python3 scripts/seed/extract_dictionary.py

# Optional flags
python3 scripts/seed/extract_dictionary.py \
    --pdf docs/kamus-bahasa-banjar-dialek-hulu.pdf \
    --out scripts/seed/seed_data.json \
    --flat   # output flat list, no grouping
```

Output: `scripts/seed/seed_data.json`

### 8.2 Output Schema

```json
{
  "meta": {
    "source": "Kamus Bahasa Banjar Dialek Hulu-Indonesia",
    "edition": "Edisi Pertama",
    "publisher": "Balai Bahasa Banjarmasin, Departemen Pendidikan Nasional",
    "year": 2008,
    "isbn": "978-979-685-776-0",
    "dialect": "hulu",
    "total_entries": 5436
  },
  "entries": [
    {
      "banjar": "abah",
      "banjar_syllabified": "a.bah",
      "word_class": "n",
      "word_class_full": "nomina",
      "definitions": ["ayah", "mertua laki-laki"],
      "examples": [
        {
          "banjar": "inya kada berayah",
          "indonesian": "dia tidak berayah"
        }
      ],
      "is_derived": false,
      "homonym_number": 1,
      "dialect": "hulu",
      "source": "seeded",
      "source_reference": "Kamus Bahasa Banjar Dialek Hulu-Indonesia, Edisi Pertama (Balai Bahasa Banjarmasin, 2008)",
      "derived_forms": [ ... ]
    }
  ]
}
```

### 8.3 Known Extraction Limitations

The PDF was digitized from print; extraction via `pdfminer.six` has these known issues:

| Issue | Description | Impact |
|---|---|---|
| Line bleeding | Example text bleeds into next entry definition | Some `example.indonesian` fields contain next entry text |
| OCR artefacts | Characters like `(cid:9)`, garbled headers | Filtered by extractor but some may remain |
| Syllable dot loss | Entry word dots not always correctly split | `banjar_syllabified` may not be accurate |
| Derived form grouping | Suffix-matching heuristic may mismatch | Some derived forms may be attached to wrong root |
| Homonym parsing | Numeric prefix stripped from word | Homonym entries need manual verification |
| Multi-definition split | Numbered defs (1/2/3) split on whitespace heuristic | May merge definitions or split incorrectly |

**Recommended review process before first deploy:**

1. Run `extract_dictionary.py` → inspect `seed_data.json`
2. Spot-check 50 random entries against the source PDF
3. Fix obvious errors directly in `seed_data.json` (it is the canonical seed file, not auto-regenerated)
4. Commit the reviewed `seed_data.json`
5. Run the Go seeder to import into PostgreSQL

### 8.4 Go Seeder (to implement)

The Go seeder reads `scripts/seed/seed_data.json` and upserts into the database.

Location: `scripts/seed/main.go`

Upsert key: `(banjar, dialect, homonym_number, is_root, root_word_id)`

Behavior:
- On conflict: update `definitions`, `examples`, `word_class`, `updated_at`
- Set `source = seeded`, `created_by = NULL`
- Idempotent — safe to re-run after PDF re-extraction

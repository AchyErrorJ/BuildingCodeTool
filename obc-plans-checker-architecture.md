# OBC Plans Checker — Architecture

## Core principle

The LLM never performs compliance evaluation. It classifies, retrieves, and explains. A deterministic engine evaluates. Every discretionary judgment carries a citation back to source text — no unsourced compliance claims.

```
Building model (IFC) → Fact extraction (code)
Code corpus         → Rule extraction (LLM-assisted, human-reviewed, one-time)
                          ↓
              Applicability classification
         (structured filters + semantic retrieval)
                          ↓
        ┌─────────────────┴─────────────────┐
   Deterministic evaluation          Discretionary tier
   (quantifiable predicates)    (objective-based / alt solutions,
        code, no LLM              LLM + mandatory citation)
        └─────────────────┬─────────────────┘
                          ↓
                  Report generation
```

---

## 1. Two-layer rule corpus

Ontario's code is not a standalone document — it is NBC 2020 (First Printing) plus a separately-issued Ontario amendment document (currently under O. Reg. 163/24). Keep these as two layers, not a flattened merge.

- **Base layer**: NBC 2020 articles, extracted once, national numbering preserved.
- **Overlay layer**: Ontario amendments, encoded as operations against base articles (`add` / `delete` / `replace`), each carrying its own issue date.

Resolve overlay-over-base at query time, not at extraction time. Rationale:
- Base layer is stable; overlay revs independently and more often — you want to re-version only what actually moved.
- Citation output should show provenance through both layers: `NBC Art. 9.8.8.1, as amended by O.Reg 163/24 (2026-04-21)`.

## 2. Predicate schema

```json
{
  "article_id": "9.8.8.1",
  "layer": "base | overlay",
  "overlay_op": "add | delete | replace | null",
  "overlay_source": "O.Reg 163/24 (2026-04-21)",
  "division": "B",
  "part": 9,
  "applies_if": {
    "occupancy": ["residential"],
    "element": "guard",
    "condition": "exterior_drop_mm > 600"
  },
  "predicate": "guard_height_mm >= 900",
  "defined_terms_used": ["exterior_drop", "guard"],
  "cross_references": ["1.4.1.2", "9.8.7.1"],
  "discretionary": false,
  "source_text": "<verbatim clause text>",
  "source_text_hash": "sha256:...",
  "extraction_reviewed_by": "jer",
  "extraction_reviewed_date": "2026-09-15"
}
```

`discretionary: true` routes the article to the judgment tier instead of the deterministic evaluator — reserved for language like "acceptable to the authority having jurisdiction," "suitable for the purpose," or clauses gated by a defined term whose scope sits elsewhere in Division A.

## 3. Fact extraction (from IFC)

Deterministic, no LLM. Pull straight from geometry/model data:

- Room areas, occupancy load inputs
- Corridor/egress widths, travel distances
- Guard/handrail heights, riser/tread dimensions
- Door swing clearances, fire separation adjacencies
- Construction type, sprinklered status, building height/area (project-level metadata, not geometry-derived)

## 4. Applicability classification

Two mechanisms, used together:

1. **Structured filters** — occupancy class, construction type, sprinklered/not, building height/area. Narrows the candidate predicate set hard.
2. **Semantic retrieval** — embed predicates (with `applies_if` context, defined terms, cross-references) and retrieve against the building's characteristics. Catches triggers the structured filters can't express, and surfaces cross-referenced/defined-term dependencies the numbering alone would miss.

Semantic search selects candidates. It never decides pass/fail.

## 5. Deterministic evaluation

For every non-discretionary predicate in the applicable set: evaluate against extracted facts in code. Output: `pass | fail | needs-review` (needs-review = a required fact wasn't extractable from the model — missing geometry, not a compliance judgment).

## 6. Discretionary tier

For `discretionary: true` predicates and objective-based/alternative-solution reasoning: LLM reasons over the retrieved source text and produces a compliance argument. Hard requirement — no exceptions: every assertion must quote/cite the specific article and show the source text it's grounded in. If it can't cite the clause, it doesn't get to assert compliance. Output is routed to human review, never auto-passed.

## 7. Report generation

LLM formats structured evaluator output (pass/fail/needs-review + citations) into a legible report. Summarization only — no numbers invented, no re-deriving compliance in prose.

## 8. Versioning

Pin a `code_version` pair per project, keyed to permit-application date, not "current":

```json
{
  "base": "NBC2020-1st-printing",
  "overlay": "O.Reg163/24-2026-04-21"
}
```

Design the base-layer swap (NBC2020 → NBC2025 once Ontario adopts it) as a first-class operation now — don't retrofit it later. Historical/in-flight projects may legitimately need an older pair under transition provisions.

## 9. Change detection across revisions

Article numbers and wording shift between editions/amendments. Diff by semantic similarity against the existing predicate corpus (not by article number) to flag which extracted predicates are likely affected by a reworded/renumbered clause — targets re-extraction instead of forcing a full re-pass.

## 10. Extraction priority (Part 9 starter set)

Given the residential focus of Legible Studio's pipeline:

- 9.5 — Design of Areas, Spaces and Doorways (egress widths, room dimensions)
- 9.8 — Stairs, Ramps, Handrails and Guards
- 9.9 — Means of Egress
- 9.10 — Fire Protection (separations, alarms)
- 9.32 — Ventilation
- 9.36 — Energy Efficiency (heaviest amendment churn — expect this to move most between overlay revisions)

## 11. Integration point

Runs as a discrete validation service, called with IFC output right before "permit-ready drawings" in the constraint-satisfaction pipeline — or invoked from RevitMCP as a validation pass. Returns structured pass/fail/needs-review, not free text, keeping it swappable behind the existing Claude-first interface.

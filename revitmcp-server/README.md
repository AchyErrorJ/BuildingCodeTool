# RevitMCP Server - Building Code Verification

An ultra-portable, zero-install MCP (Model Context Protocol) server for Revit that provides building code verification capabilities. Designed for architecture students to validate their Revit models against Ontario Building Code (OBC) requirements.

## Features

- **Zero Installation**: Single executable file (~2.5MB), no Python/Node.js dependencies required
- **Revit Integration**: Connects directly to Revit via MCP protocol
- **IFC-Based Analysis**: Extracts building facts from IFC exports
- **Deterministic Evaluation**: Code-based rule evaluation (no LLM hallucination)
- **Part 9 Focus**: Residential building code checks (stairs, guards, egress, etc.)
- **Citation-Backed Results**: Every compliance claim references specific OBC articles
- **Cross-Platform**: Windows, Mac (Intel & Apple Silicon), Linux support

## Quick Start

### For Students

1. Download the executable for your OS:
   - **Windows**: `revitmcp-server.exe`
   - **Mac (Intel)**: `revitmcp-server-macos-amd64`
   - **Mac (Apple Silicon M1/M2/M3)**: `revitmcp-server-macos-arm64`
   - **Linux**: `revitmcp-server-linux`
2. Place in any folder (e.g., `C:\Tools\revitmcp-server.exe`)
3. Configure your MCP client (see configuration section below)
4. Run code checks directly from Revit!

### Configuration

Add to your MCP settings (typically `%APPDATA%\Claude\mcp.json` on Windows or `~/.config/claude/mcp.json` on Mac/Linux):

**Windows:**
```json
{
  "mcpServers": {
    "obc-checker": {
      "command": "C:\\path\\to\\revitmcp-server.exe",
      "args": ["--mode", "server"]
    }
  }
}
```

**Mac/Linux:**
```json
{
  "mcpServers": {
    "obc-checker": {
      "command": "/Users/yourname/tools/revitmcp-server-macos-arm64",
      "args": ["--mode", "server"]
    }
  }
}
```

Then restart your MCP client (Claude Desktop, etc.).

## Available Tools

### `check_building_code`

Analyzes the current Revit model (exported as IFC) for OBC compliance.

**Input:**
```json
{
  "ifc_path": "path/to/model.ifc",
  "code_version": {
    "base": "NBC2020-1st-printing",
    "overlay": "O.Reg163/24-2026-04-21"
  },
  "parts": ["9.8", "9.9", "9.5"]
}
```

**Output:**
```
┌─────────────────────────────────────────┐
│  OBC COMPLIANCE CHECK RESULTS           │
└─────────────────────────────────────────┘

SUMMARY: 5 Pass | 2 Fail | 1 Needs Review

[✓] 9.8.8.1 - Guard at exterior_deck: 1067mm >= 900mm ✓
    Citation: NBC Art. 9.8.8.1

[✗] 9.8.4.1 - Stair flight(s) exceed maximum riser height: 210mm > 200mm ✗
    Citation: NBC Art. 9.8.4.1
```

### `extract_facts`

Extracts building facts from IFC model without evaluation. Useful for understanding what data is available.

**Input:**
```json
{
  "ifc_path": "path/to/model.ifc"
}
```

**Output:**
```
EXTRACTED BUILDING FACTS

Guards: 3
  - exterior_deck: 1067mm high, 2400mm drop, exterior=true
  - interior_loft: 900mm high, 2700mm drop, exterior=false
  - front_porch: 850mm high, 750mm drop, exterior=true

Stairs: 2
  - Riser: 180mm, Tread: 280mm, Width: 900mm
  - Riser: 210mm, Tread: 200mm, Width: 860mm

Rooms: 3
  - Bedroom 1: 12.5 m² (residential)
  - Bedroom 2: 10.2 m² (residential)
  - Living Room: 25.0 m² (residential)
```

### `get_applicable_rules`

Returns all OBC rules applicable to the current model based on occupancy and elements present.

**Input:**
```json
{
  "ifc_path": "path/to/model.ifc",
  "part": "9.8"
}
```

## Architecture

Follows the OBC Plans Checker architecture from the specification:

1. **Fact Extraction**: Deterministic IFC parsing (no LLM)
2. **Rule Corpus**: Two-layer system (NBC base + Ontario overlay)
3. **Applicability Classification**: Structured filters + semantic retrieval
4. **Deterministic Evaluation**: Code-based predicate evaluation
5. **Discretionary Tier**: LLM-assisted judgment with mandatory citation
6. **Report Generation**: Structured pass/fail/needs-review output

## Predicate Schema

Rules are encoded as structured predicates following the architecture spec:

```json
{
  "article_id": "9.8.8.1",
  "layer": "base",
  "applies_if": {
    "occupancy": ["residential"],
    "element": "guard",
    "condition": "exterior_drop_mm > 600"
  },
  "predicate": "guard_height_mm >= 900",
  "discretionary": false,
  "source_text": "Exterior guards serving a building of residential occupancy shall be not less than 900 mm high.",
  "citation": "NBC Art. 9.8.8.1"
}
```

## Supported Code Sections (Part 9)

Currently implements checks for:

| Section | Topic | Key Requirements |
|---------|-------|------------------|
| **9.5** | Design of Areas, Spaces and Doorways | Bedroom window egress |
| **9.8** | Stairs, Ramps, Handrails and Guards | Guard heights (900mm), stair risers (≤200mm), treads (≥210mm), widths (≥860mm), handrail heights (865-965mm) |
| **9.9** | Means of Egress | Egress door dimensions (860x1950mm min) |
| 9.10 | Fire Protection | _Coming soon_ |
| 9.32 | Ventilation | _Coming soon_ |
| 9.36 | Energy Efficiency | _Coming soon_ |

## Educational Use

This tool is designed for architecture students to:

- **Learn building code requirements** through practical application
- **Validate design decisions early** in the design process
- **Understand the relationship** between geometry and code compliance
- **Generate citation-backed compliance reports** for studio reviews
- **Catch common mistakes** before submitting work

### Typical Student Workflow

1. Model in Revit as normal
2. Export to IFC (File → Export → IFC)
3. Ask Claude: "Check my building model for OBC compliance"
4. Review results and fix violations
5. Iterate until all checks pass

## Demo Mode (No Revit Required)

Want to see how it works without Revit? Run standalone demo:

```bash
# Windows
revitmcp-server.exe --mode standalone

# Mac
./revitmcp-server-macos-arm64 --mode standalone

# Linux
./revitmcp-server-linux --mode standalone
```

This displays sample output with test data demonstrating the tool's capabilities.

## Technical Details

- **Language**: Go (compiled to single binary, no runtime required)
- **Protocol**: MCP (Model Context Protocol) over stdio
- **IFC Parsing**: Simulated in demo, ready for ifcOpenShell integration
- **Rule Engine**: Deterministic predicate evaluation
- **Portability**: Single-file executables for Windows/Mac/Linux
- **Size**: ~2.5MB per platform
- **License**: MIT License - Free for educational use

## Building from Source

If you want to modify or extend the tool:

```bash
# Install Go (https://golang.org/dl/)
go version  # Should be 1.21+

# Build for current platform
go build -o revitmcp-server .

# Cross-compile for other platforms
GOOS=windows GOARCH=amd64 go build -o revitmcp-server.exe .
GOOS=darwin GOARCH=arm64 go build -o revitmcp-server-macos-arm64 .
GOOS=linux GOARCH=amd64 go build -o revitmcp-server-linux .
```

## Extending the Rule Set

To add new code checks:

1. Add a new `Predicate` to the `ruleDatabase` slice in `main.go`
2. Implement the evaluation function (e.g., `evaluateGuardRule`)
3. Wire it up in `evaluatePredicate` switch statement

Example new rule:
```go
{
    ArticleID: "9.8.9.1",
    Layer:     "base",
    Division:  "B",
    Part:      9,
    AppliesIf: AppliesIf{
        Occupancy: []string{"residential"},
        Element:   "window_well",
    },
    Predicate:     "window_well_projection_mm <= 600",
    Discretionary: false,
    SourceText:    "Window wells shall not project more than 600 mm...",
    Citation:      "NBC Art. 9.8.9.1",
},
```

## Limitations & Future Work

### Current Limitations

- **Simulated IFC parsing**: Currently returns sample data. Production version needs ifcOpenShell integration
- **Part 9 only**: Focused on residential; other parts coming soon
- **Limited element types**: Guards, stairs, doors, rooms only

### Roadmap

- [ ] Real IFC parsing with ifcOpenShell bindings
- [ ] Expand to Part 10 (Fire Protection)
- [ ] Add Part 9.36 (Energy Efficiency) checks
- [ ] Support for more element types (windows, corridors, plumbing)
- [ ] GUI dashboard for visualizing violations
- [ ] Batch checking multiple IFC files
- [ ] Export reports to PDF/HTML

## Troubleshooting

### "Command not found"
Ensure the executable path in your MCP config matches where you saved the file. Use full absolute paths.

### No results returned
- Verify you exported IFC from Revit
- Check that your model contains residential occupancy elements
- Ensure guards, stairs, or doors are present for relevant checks

### Wrong measurements
The tool reads from IFC export. If measurements seem off:
- Check your Revit model geometry
- Verify IFC export settings
- Re-export the IFC file

## Support

For questions, issues, or feature requests, contact: [your-contact@example.com]

## License

MIT License - Free for educational use. See LICENSE file for details.

---

**Built for architecture students** to learn building code compliance through hands-on practice with real design tools.


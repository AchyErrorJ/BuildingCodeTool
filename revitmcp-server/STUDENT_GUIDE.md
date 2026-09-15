# Student Quick Start Guide

## What is This Tool?

The **OBC Plans Checker** is a building code verification tool that connects directly to Revit. It automatically checks your residential designs against Ontario Building Code (OBC) requirements for:

- Guards and handrails
- Stair dimensions (risers, treads, widths)
- Egress doors
- Bedroom windows
- And more Part 9 requirements

## Zero Installation Required

This is a **portable executable** - no Python, no Node.js, no complex setup:

1. Download `revitmcp-server.exe` (Windows) or the appropriate version for your OS
2. Place it anywhere on your computer (e.g., `C:\Tools\revitmcp-server.exe`)
3. Configure your MCP client
4. Start checking codes from Revit!

## Step-by-Step Setup

### 1. Download the Executable

Choose the correct version for your operating system:
- **Windows**: `revitmcp-server.exe`
- **Mac (Intel)**: `revitmcp-server-macos-amd64`
- **Mac (Apple Silicon M1/M2)**: `revitmcp-server-macos-arm64`
- **Linux**: `revitmcp-server-linux`

### 2. Configure Your MCP Client

If you're using Claude Desktop or another MCP-enabled client, add this to your MCP configuration file:

**Windows** (`%APPDATA%\Claude\mcp.json` or similar):
```json
{
  "mcpServers": {
    "obc-checker": {
      "command": "C:\\Tools\\revitmcp-server.exe",
      "args": ["--mode", "server"]
    }
  }
}
```

**Mac/Linux** (`~/.config/claude/mcp.json` or similar):
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

### 3. Restart Your MCP Client

Close and reopen Claude Desktop (or your MCP client) to load the new server.

## Using the Tool in Revit

### Workflow

1. **Model in Revit** as normal
2. **Export to IFC** (File → Export → IFC)
3. **Ask Claude** to check your model:
   - "Check my building model for OBC compliance"
   - "What code violations are in this design?"
   - "Extract facts from my IFC model"

### Available Commands

Through the MCP interface, you can:

#### `check_building_code`
Analyzes your IFC model for OBC compliance.

**Example prompt:**
> "Run a building code check on C:\Projects\Studio\house.ifc"

**Returns:**
- Pass/fail summary
- Specific violations with citations
- Exact measurements from your model

#### `extract_facts`
Shows what building data was extracted from your IFC.

**Example prompt:**
> "What facts can you extract from my model?"

**Returns:**
- Guard heights and locations
- Stair dimensions
- Room areas
- Door sizes

#### `get_applicable_rules`
Lists all OBC rules that apply to your specific model.

**Example prompt:**
> "Which code sections apply to my design?"

## Understanding the Results

### Status Indicators

- **✓ Pass**: Your design meets the code requirement
- **✗ Fail**: Your design violates the code - needs revision
- **? Needs Review**: Manual review required (missing data or discretionary judgment)

### Example Output

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

### What Each Result Means

Every result includes:
1. **Article ID**: The specific OBC section (e.g., 9.8.8.1)
2. **Measurement**: Actual value from your model
3. **Requirement**: What the code demands
4. **Citation**: Verbatim code text for reference

## Supported Code Sections

Currently checks Part 9 (Housing and Small Buildings):

| Section | Topic |
|---------|-------|
| 9.5 | Design of Areas, Spaces and Doorways |
| 9.8 | Stairs, Ramps, Handrails and Guards |
| 9.9 | Means of Egress |
| 9.10 | Fire Protection (coming soon) |
| 9.32 | Ventilation (coming soon) |
| 9.36 | Energy Efficiency (coming soon) |

## Demo Mode (No Revit Required)

Want to try it without Revit? Run standalone demo:

```bash
# Windows
revitmcp-server.exe --mode standalone

# Mac/Linux
./revitmcp-server-macos-arm64 --mode standalone
```

This shows sample output with test data so you can see how the tool works.

## Tips for Students

### Best Practices

1. **Check Early, Check Often**: Run code checks throughout design development, not just at the end
2. **Export Clean IFC**: Use consistent naming and proper element categories in Revit
3. **Read the Citations**: Every result links to actual code text - use this to learn the requirements
4. **Fix Fails First**: Prioritize failures (✗) over "needs review" (?) items

### Common Violations

Watch out for these frequent issues:
- **Guards under 900mm** (especially on porches/decks)
- **Stair risers over 200mm**
- **Stair treads under 210mm**
- **Egress doors too narrow** (minimum 860mm)

### Learning Opportunity

This tool helps you:
- Understand the relationship between geometry and code compliance
- Learn to read and interpret OBC articles
- Catch mistakes before studio reviews
- Build intuition for code-compliant design

## Troubleshooting

### "Command not found"
Make sure the executable path in your MCP config matches where you saved the file.

### No results returned
- Verify you exported IFC from Revit
- Check that your model contains residential occupancy elements
- Ensure guards, stairs, or doors are present for relevant checks

### Wrong measurements
The tool reads from IFC export. If measurements seem off:
- Check your Revit model geometry
- Verify IFC export settings
- Re-export the IFC file

## Technical Details

- **Protocol**: MCP (Model Context Protocol) over stdio
- **Language**: Go (compiled single binary)
- **IFC Support**: Reads standard IFC2x3 / IFC4 exports from Revit
- **Code Version**: NBC 2020 + Ontario Amendment O.Reg 163/24
- **License**: MIT (free for educational use)

## Next Steps

1. Install the tool following the steps above
2. Export an IFC from your current Revit project
3. Run your first code check
4. Bring results to studio for discussion!

---

**Questions?** Contact your instructor or TA for support.

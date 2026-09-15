# Production Deployment Guide

## What Changed

The tool has been upgraded from demo mode to production-ready IFC parsing:

### Before (Demo Mode)
- Used hardcoded sample data for all IFC files
- No actual IFC file parsing
- Good for demonstrations only

### After (Production Mode)
- **Real IFC Parsing**: Uses ifcopenshell library to extract actual building facts
- **Hybrid Architecture**: Go binary + Python script work together
- **Smart Fallback**: Demo data used only when IFC file doesn't exist
- **Property Extraction**: Reads IFC properties (height, width, area, etc.) from model elements

## Distribution Package

Students need these two files in the same folder:

```
revitmcp-folder/
├── revitmcp-server-[platform]  (executable ~2.5MB)
└── ifc_extractor.py            (Python script ~15KB)
```

### Platform Executables
- `revitmcp-server.exe` - Windows
- `revitmcp-server-macos-amd64` - Mac Intel
- `revitmcp-server-macos-arm64` - Mac M1/M2/M3
- `revitmcp-server-linux` - Linux

## Requirements

### Minimal Requirement (Demo Mode)
- Just the executable - runs with built-in demo data

### Full Production Mode
- The executable
- `ifc_extractor.py` script (bundled with executable)
- Python 3.x with ifcopenshell:
  ```bash
  pip install ifcopenshell
  ```

## How It Works

1. **MCP Request**: Revit sends `check_building_code` request with IFC path
2. **File Check**: Go binary checks if IFC file exists
3. **If File Exists**:
   - Calls `python3 ifc_extractor.py <path>`
   - Parses JSON output with extracted facts
   - Evaluates OBC rules against real data
4. **If File Missing**:
   - Uses built-in demo data
   - Shows example violations for teaching purposes
5. **Returns Results**: Pass/fail/needs-review with citations

## Extracted Facts

The IFC extractor reads:

### Guards/Railings
- Location name
- Height (mm)
- Drop height (mm)
- Interior/exterior status

### Stairs
- Riser height (mm)
- Tread depth (mm)
- Width (mm)
- Flight count

### Rooms/Spaces
- Name
- Area (m²)
- Occupancy type

### Doors
- Location
- Width/height (mm)
- Swing clear width (mm)
- Egress designation

### Corridors
- Location
- Width (mm)
- Length (mm)

## Building Code Checks

Current Part 9 rules implemented:

| Article | Element | Requirement | Status |
|---------|---------|-------------|--------|
| 9.8.8.1 | Exterior guards | ≥900mm height | ✓ |
| 9.8.8.2 | Interior guards | ≥900mm height | ✓ |
| 9.8.7.1 | Handrails | 865-965mm height | ✓ |
| 9.8.4.1 | Stair risers | ≤200mm height | ✓ |
| 9.8.4.2 | Stair treads | ≥210mm depth | ✓ |
| 9.8.3.1 | Stair width | ≥860mm | ✓ |
| 9.5.4.1 | Bedroom windows | 0.35m² min area | ✓ |
| 9.9.7.1 | Egress doors | ≥860mm width, ≥1950mm height | ✓ |

## Testing

### Test with Demo Data
```bash
./revitmcp-server-linux --mode standalone
```

### Test with Real IFC
```bash
# Create a test IFC from Revit
# Export as IFC4 format
# Run check
./revitmcp-server-linux --mode server
# Then use MCP client to call check_building_code with IFC path
```

### Test Python Extractor Directly
```bash
python3 ifc_extractor.py /path/to/model.ifc
```

## Common Issues

### "python3 not found"
- Install Python 3 from python.org
- Ensure python3 is in PATH

### "ModuleNotFoundError: ifcopenshell"
```bash
pip install ifcopenshell
```

### "IFC extraction failed"
- Check that IFC file path is correct
- Verify IFC file is valid (open in Solibri/BIMvision)
- Check Python script permissions

### Poor Extraction Results
- Ensure Revit exports properties to IFC
- Use IFC4 format (not IFC2x3)
- Check element naming conventions (guards should have "guard" or "railing" in name)

## For Educators

This tool teaches students:
1. **Code Literacy**: Reading and understanding OBC articles
2. **BIM Coordination**: Connecting geometry to regulatory requirements
3. **Automated Compliance**: How AI/automation assists (but doesn't replace) professional judgment
4. **Iterative Design**: Quick feedback loop for design improvements

### Classroom Exercise Ideas

1. **Violation Hunt**: Give students a model with intentional violations, have them find and fix
2. **Rule Research**: Have students look up cited articles and explain the rationale
3. **Comparison Study**: Compare automated results with manual code review
4. **Extension Project**: Add new rules to the rule database

## Future Enhancements

- [ ] More Part 9 rules (windows, corridors, fire separations)
- [ ] Part 3 support (commercial buildings)
- [ ] Ontario-specific amendments overlay
- [ ] Discretionary rule flagging for AHJ review
- [ ] Native Windows bundle with embedded Python
- [ ] Revit plugin UI (beyond MCP)

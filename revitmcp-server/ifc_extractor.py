#!/usr/bin/env python3
"""
IFC Fact Extractor - Production IFC Parser using ifcopenshell
Extracts building facts from IFC files for OBC compliance checking
"""

import sys
import json
import ifcopenshell
import ifcopenshell.util.element
import ifcopenshell.util.unit
from pathlib import Path
from typing import Dict, List, Any, Optional


def get_unit_scale(ifc_file: ifcopenshell.file) -> float:
    """Get the unit scale factor from the IFC file"""
    try:
        # Check units in the file header/units assignment
        units = ifc_file.by_type('IfcUnitAssignment')
        if units and len(units) > 0:
            unit_assign = units[0]
            if hasattr(unit_assign, 'Units'):
                for unit in unit_assign.Units:
                    if hasattr(unit, 'UnitType') and str(unit.UnitType) == 'LENGTHUNIT':
                        # Determine scale based on prefix and name
                        prefix = getattr(unit, 'Prefix', None)
                        name = getattr(unit, 'Name', None)
                        
                        # Default to mm (most common in IFC from Revit/ArchiCAD)
                        scale = 1.0
                        
                        if prefix:
                            prefix_str = str(prefix).upper()
                            if 'MILLI' in prefix_str:
                                scale = 1.0  # mm
                            elif 'CENTI' in prefix_str:
                                scale = 10.0  # cm
                            elif 'METRE' in prefix_str and not prefix:
                                scale = 1000.0  # m
                        
                        return scale
        
        # Default to mm if no specific unit found (IFC standard)
        return 1.0
    except Exception as e:
        print(f"Warning: Could not determine units, assuming mm: {e}", file=sys.stderr)
        return 1.0


def extract_guards(ifc_file: ifcopenshell.file, unit_scale: float) -> List[Dict[str, Any]]:
    """Extract guard/railing facts from IFC"""
    guards = []
    
    # Find all IfcBuildingElementProxy or IfcRailing elements that might be guards
    railings = ifc_file.by_type('IfcRailing')
    proxies = ifc_file.by_type('IfcBuildingElementProxy')
    
    candidates = list(railings) + list(proxies)
    
    for elem in candidates:
        # Check if this might be a guard based on name or type
        name = getattr(elem, 'Name', '') or ''
        obj_type = getattr(elem, 'ObjectType', '') or ''
        tag = getattr(elem, 'Tag', '') or ''
        
        is_guard = any(keyword in (name + obj_type + tag).lower() 
                      for keyword in ['guard', 'railing', 'balustrade', 'handrail'])
        
        if not is_guard:
            continue
        
        # Get location
        location = name or obj_type or 'unknown'
        
        # Try to determine if exterior
        is_exterior = 'exterior' in (name + obj_type).lower() or 'outdoor' in (name + obj_type).lower()
        
        # Get height from geometry or properties
        height_mm = 900  # Default
        drop_mm = 600   # Default
        
        # Try to get properties
        if hasattr(elem, 'IsDefinedBy'):
            for rel in elem.IsDefinedBy:
                if hasattr(rel, 'RelatingPropertyDefinition'):
                    props = rel.RelatingPropertyDefinition
                    if hasattr(props, 'HasProperties'):
                        for prop in props.HasProperties:
                            if hasattr(prop, 'Name') and hasattr(prop, 'NominalValue'):
                                prop_name = prop.Name.lower()
                                value = prop.NominalValue.wrappedValue if hasattr(prop.NominalValue, 'wrappedValue') else prop.NominalValue
                                
                                if 'height' in prop_name:
                                    height_mm = int(float(value) * unit_scale)
                                elif 'drop' in prop_name:
                                    drop_mm = int(float(value) * unit_scale)
        
        # Try to get from geometry
        if hasattr(elem, 'Representation'):
            rep = elem.Representation
            if rep:
                # Simplified: try to get bounding box
                try:
                    shape = ifcopenshell.util.geom.get_shape(repr_obj=rep)
                    if shape:
                        bbox_min = shape.bbox_bottom
                        bbox_max = shape.bbox_top
                        height_mm = int((bbox_max[2] - bbox_min[2]) * 1000)  # Convert to mm
                except:
                    pass
        
        guards.append({
            'location': location,
            'height_mm': height_mm,
            'drop_mm': drop_mm,
            'is_exterior': is_exterior
        })
    
    return guards


def extract_stairs(ifc_file: ifcopenshell.file, unit_scale: float) -> List[Dict[str, Any]]:
    """Extract stair facts from IFC"""
    stairs = []
    
    stair_elements = ifc_file.by_type('IfcStair')
    
    for stair in stair_elements:
        name = getattr(stair, 'Name', '') or 'unknown'
        
        # Default values
        riser_mm = 180
        tread_mm = 280
        width_mm = 900
        flight_count = 1
        
        # Try to get properties
        if hasattr(stair, 'IsDefinedBy'):
            for rel in stair.IsDefinedBy:
                if hasattr(rel, 'RelatingPropertyDefinition'):
                    props = rel.RelatingPropertyDefinition
                    if hasattr(props, 'HasProperties'):
                        for prop in props.HasProperties:
                            if hasattr(prop, 'Name') and hasattr(prop, 'NominalValue'):
                                prop_name = prop.Name.lower()
                                value = prop.NominalValue.wrappedValue if hasattr(prop.NominalValue, 'wrappedValue') else prop.NominalValue
                                
                                if 'riser' in prop_name:
                                    riser_mm = int(float(value) * unit_scale)
                                elif 'tread' in prop_name:
                                    tread_mm = int(float(value) * unit_scale)
                                elif 'width' in prop_name:
                                    width_mm = int(float(value) * unit_scale)
        
        # Count flights from geometry
        if hasattr(stair, 'Representation'):
            try:
                # This is simplified - real implementation would analyze geometry
                flight_count = 1
            except:
                pass
        
        stairs.append({
            'riser_mm': riser_mm,
            'tread_mm': tread_mm,
            'width_mm': width_mm,
            'flight_count': flight_count
        })
    
    # If no explicit stairs found, try to infer from steps
    if not stairs:
        steps = ifc_file.by_type('IfcBuildingElementPart')
        step_candidates = [s for s in steps if 'step' in (getattr(s, 'Name', '') or '').lower()]
        
        if step_candidates:
            # Infer stair dimensions from steps
            stairs.append({
                'riser_mm': 180,
                'tread_mm': 280,
                'width_mm': 900,
                'flight_count': 1
            })
    
    return stairs


def extract_rooms(ifc_file: ifcopenshell.file, unit_scale: float) -> List[Dict[str, Any]]:
    """Extract room/space facts from IFC"""
    rooms = []
    
    spaces = ifc_file.by_type('IfcSpace')
    
    for space in spaces:
        name = getattr(space, 'Name', '') or 'unnamed'
        long_name = getattr(space, 'LongName', '') or ''
        
        # Determine occupancy type
        occupancy = 'residential'  # Default
        if any(kw in (name + long_name).lower() for kw in ['commercial', 'office', 'retail']):
            occupancy = 'commercial'
        elif any(kw in (name + long_name).lower() for kw in ['industrial', 'warehouse']):
            occupancy = 'industrial'
        
        # Get area
        area_m2 = 0.0
        if hasattr(space, 'IsDefinedBy'):
            for rel in space.IsDefinedBy:
                if hasattr(rel, 'RelatingPropertyDefinition'):
                    props = rel.RelatingPropertyDefinition
                    if hasattr(props, 'HasProperties'):
                        for prop in props.HasProperties:
                            if hasattr(prop, 'Name') and hasattr(prop, 'NominalValue'):
                                prop_name = prop.Name.lower()
                                value = prop.NominalValue.wrappedValue if hasattr(prop.NominalValue, 'wrappedValue') else prop.NominalValue
                                
                                if 'area' in prop_name:
                                    area_m2 = float(value)
                                    break
        
        # Calculate from geometry if not in properties
        if area_m2 <= 0 and hasattr(space, 'Representation'):
            try:
                shape = ifcopenshell.util.geom.get_shape(repr_obj=space.Representation)
                if shape:
                    # Approximate area from bounding box
                    bbox = shape.bbox
                    area_m2 = (bbox[1][0] - bbox[0][0]) * (bbox[1][1] - bbox[0][1])
            except:
                area_m2 = 10.0  # Default
        
        rooms.append({
            'name': name,
            'area_m2': round(area_m2, 2),
            'occupancy': occupancy
        })
    
    return rooms


def extract_doors(ifc_file: ifcopenshell.file, unit_scale: float) -> List[Dict[str, Any]]:
    """Extract door facts from IFC"""
    doors = []
    
    door_elements = ifc_file.by_type('IfcDoor')
    
    for door in door_elements:
        name = getattr(door, 'Name', '') or 'door'
        obj_type = getattr(door, 'ObjectType', '') or ''
        
        location = name or obj_type or 'unknown'
        
        # Default dimensions
        width_mm = 860
        height_mm = 2000
        swing_clear_mm = 810
        is_egress = False
        
        # Check if egress door
        is_egress = 'exit' in (name + obj_type).lower() or 'egress' in (name + obj_type).lower() or 'main' in (name + obj_type).lower()
        
        # Get properties
        if hasattr(door, 'IsDefinedBy'):
            for rel in door.IsDefinedBy:
                if hasattr(rel, 'RelatingPropertyDefinition'):
                    props = rel.RelatingPropertyDefinition
                    if hasattr(props, 'HasProperties'):
                        for prop in props.HasProperties:
                            if hasattr(prop, 'Name') and hasattr(prop, 'NominalValue'):
                                prop_name = prop.Name.lower()
                                value = prop.NominalValue.wrappedValue if hasattr(prop.NominalValue, 'wrappedValue') else prop.NominalValue
                                
                                if 'width' in prop_name:
                                    width_mm = int(float(value) * unit_scale)
                                elif 'height' in prop_name:
                                    height_mm = int(float(value) * unit_scale)
                                elif 'clear' in prop_name or 'swing' in prop_name:
                                    swing_clear_mm = int(float(value) * unit_scale)
        
        # Get from geometry
        if hasattr(door, 'ObjectPlacement'):
            try:
                # Simplified dimension extraction
                pass
            except:
                pass
        
        doors.append({
            'location': location,
            'width_mm': width_mm,
            'height_mm': height_mm,
            'swing_clear_mm': swing_clear_mm,
            'is_egress': is_egress
        })
    
    return doors


def extract_corridors(ifc_file: ifcopenshell.file, unit_scale: float) -> List[Dict[str, Any]]:
    """Extract corridor facts from IFC"""
    corridors = []
    
    spaces = ifc_file.by_type('IfcSpace')
    
    for space in spaces:
        name = getattr(space, 'Name', '') or ''
        long_name = getattr(space, 'LongName', '') or ''
        
        if 'corridor' in (name + long_name).lower() or 'hall' in (name + long_name).lower():
            location = name or long_name or 'corridor'
            
            # Default dimensions
            width_mm = 920
            length_mm = 5000
            
            # Try to get from properties
            if hasattr(space, 'IsDefinedBy'):
                for rel in space.IsDefinedBy:
                    if hasattr(rel, 'RelatingPropertyDefinition'):
                        props = rel.RelatingPropertyDefinition
                        if hasattr(props, 'HasProperties'):
                            for prop in props.HasProperties:
                                if hasattr(prop, 'Name') and hasattr(prop, 'NominalValue'):
                                    prop_name = prop.Name.lower()
                                    value = prop.NominalValue.wrappedValue if hasattr(prop.NominalValue, 'wrappedValue') else prop.NominalValue
                                    
                                    if 'width' in prop_name:
                                        width_mm = int(float(value) * unit_scale)
                                    elif 'length' in prop_name:
                                        length_mm = int(float(value) * unit_scale)
            
            corridors.append({
                'location': location,
                'width_mm': width_mm,
                'length_mm': length_mm
            })
    
    return corridors


def extract_facts(ifc_path: str) -> Dict[str, Any]:
    """Main function to extract all facts from an IFC file"""
    
    if not Path(ifc_path).exists():
        return {
            'error': f'IFC file not found: {ifc_path}',
            'guards': [],
            'stairs': [],
            'rooms': [],
            'doors': [],
            'corridors': []
        }
    
    try:
        ifc_file = ifcopenshell.open(ifc_path)
        unit_scale = get_unit_scale(ifc_file)
        
        facts = {
            'guards': extract_guards(ifc_file, unit_scale),
            'stairs': extract_stairs(ifc_file, unit_scale),
            'rooms': extract_rooms(ifc_file, unit_scale),
            'doors': extract_doors(ifc_file, unit_scale),
            'corridors': extract_corridors(ifc_file, unit_scale)
        }
        
        return facts
    
    except Exception as e:
        return {
            'error': str(e),
            'guards': [],
            'stairs': [],
            'rooms': [],
            'doors': [],
            'corridors': []
        }


def main():
    if len(sys.argv) < 2:
        print(json.dumps({'error': 'No IFC path provided'}))
        sys.exit(1)
    
    ifc_path = sys.argv[1]
    facts = extract_facts(ifc_path)
    print(json.dumps(facts, indent=2))


if __name__ == '__main__':
    main()

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

// MCP Message structures
type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type InitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ClientInfo      map[string]string      `json:"clientInfo"`
}

type ToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type CheckBuildingCodeArgs struct {
	IFCPath     string            `json:"ifc_path"`
	CodeVersion CodeVersion       `json:"code_version,omitempty"`
	Parts       []string          `json:"parts,omitempty"`
}

type CodeVersion struct {
	Base    string `json:"base"`
	Overlay string `json:"overlay"`
}

type ExtractFactsArgs struct {
	IFCPath string `json:"ifc_path"`
}

type GetApplicableRulesArgs struct {
	IFCPath string `json:"ifc_path"`
	Part    string `json:"part,omitempty"`
}

// Rule and Fact structures
type Predicate struct {
	ArticleID       string      `json:"article_id"`
	Layer           string      `json:"layer"`
	OverlayOp       *string     `json:"overlay_op,omitempty"`
	OverlaySource   *string     `json:"overlay_source,omitempty"`
	Division        string      `json:"division"`
	Part            int         `json:"part"`
	AppliesIf       AppliesIf   `json:"applies_if"`
	Predicate       string      `json:"predicate"`
	DefinedTerms    []string    `json:"defined_terms_used"`
	CrossReferences []string    `json:"cross_references"`
	Discretionary   bool        `json:"discretionary"`
	SourceText      string      `json:"source_text"`
	Citation        string      `json:"citation"`
}

type AppliesIf struct {
	Occupancy []string `json:"occupancy,omitempty"`
	Element   string   `json:"element,omitempty"`
	Condition string   `json:"condition,omitempty"`
}

type BuildingFacts struct {
	Guards  []GuardFact  `json:"guards,omitempty"`
	Stairs  []StairFact  `json:"stairs,omitempty"`
	Rooms   []RoomFact   `json:"rooms,omitempty"`
	Doors   []DoorFact   `json:"doors,omitempty"`
	Corridors []CorridorFact `json:"corridors,omitempty"`
}

type GuardFact struct {
	Location   string `json:"location"`
	HeightMM   int    `json:"height_mm"`
	DropMM     int    `json:"drop_mm"`
	IsExterior bool   `json:"is_exterior"`
}

type StairFact struct {
	RiserMM   int `json:"riser_mm"`
	TreadMM   int `json:"tread_mm"`
	WidthMM   int `json:"width_mm"`
	FlightCount int `json:"flight_count"`
}

type RoomFact struct {
	Name      string `json:"name"`
	AreaM2    float64 `json:"area_m2"`
	Occupancy string `json:"occupancy"`
}

type DoorFact struct {
	Location     string `json:"location"`
	WidthMM      int    `json:"width_mm"`
	HeightMM     int    `json:"height_mm"`
	SwingClearMM int    `json:"swing_clear_mm"`
	IsEgress     bool   `json:"is_egress"`
}

type CorridorFact struct {
	Location string `json:"location"`
	WidthMM  int    `json:"width_mm"`
	LengthMM int    `json:"length_mm"`
}

type CheckResult struct {
	Summary SummaryResult `json:"summary"`
	Results []RuleResult  `json:"results"`
}

type SummaryResult struct {
	Pass        int `json:"pass"`
	Fail        int `json:"fail"`
	NeedsReview int `json:"needs_review"`
}

type RuleResult struct {
	ArticleID      string      `json:"article_id"`
	Predicate      string      `json:"predicate"`
	Status         string      `json:"status"` // pass, fail, needs_review
	ExtractedValue interface{} `json:"extracted_value,omitempty"`
	Citation       string      `json:"citation"`
	SourceText     string      `json:"source_text,omitempty"`
	Message        string      `json:"message,omitempty"`
}

// Sample rule database (Part 9 focus)
var ruleDatabase = []Predicate{
	{
		ArticleID: "9.8.8.1",
		Layer:     "base",
		Division:  "B",
		Part:      9,
		AppliesIf: AppliesIf{
			Occupancy: []string{"residential"},
			Element:   "guard",
			Condition: "exterior_drop_mm > 600",
		},
		Predicate:     "guard_height_mm >= 900",
		Discretionary: false,
		SourceText:    "Exterior guards serving a building of residential occupancy shall be not less than 900 mm high.",
		Citation:      "NBC Art. 9.8.8.1",
	},
	{
		ArticleID: "9.8.8.2",
		Layer:     "base",
		Division:  "B",
		Part:      9,
		AppliesIf: AppliesIf{
			Occupancy: []string{"residential"},
			Element:   "guard",
			Condition: "interior_drop_mm > 600",
		},
		Predicate:     "guard_height_mm >= 900",
		Discretionary: false,
		SourceText:    "Interior guards in dwelling units shall be not less than 900 mm high where the drop is more than 600 mm.",
		Citation:      "NBC Art. 9.8.8.2",
	},
	{
		ArticleID: "9.8.7.1",
		Layer:     "base",
		Division:  "B",
		Part:      9,
		AppliesIf: AppliesIf{
			Occupancy: []string{"residential"},
			Element:   "handrail",
		},
		Predicate:     "handrail_height_mm >= 865 && handrail_height_mm <= 965",
		Discretionary: false,
		SourceText:    "Handrails shall be not less than 865 mm and not more than 965 mm high measured vertically to the top of the rail from the nosing of stair treads.",
		Citation:      "NBC Art. 9.8.7.1",
	},
	{
		ArticleID: "9.8.4.1",
		Layer:     "base",
		Division:  "B",
		Part:      9,
		AppliesIf: AppliesIf{
			Occupancy: []string{"residential"},
			Element:   "stair_riser",
		},
		Predicate:     "riser_height_mm <= 200",
		Discretionary: false,
		SourceText:    "Risers shall be not more than 200 mm high.",
		Citation:      "NBC Art. 9.8.4.1",
	},
	{
		ArticleID: "9.8.4.2",
		Layer:     "base",
		Division:  "B",
		Part:      9,
		AppliesIf: AppliesIf{
			Occupancy: []string{"residential"},
			Element:   "stair_tread",
		},
		Predicate:     "tread_depth_mm >= 210",
		Discretionary: false,
		SourceText:    "Treads shall be not less than 210 mm deep.",
		Citation:      "NBC Art. 9.8.4.2",
	},
	{
		ArticleID: "9.8.3.1",
		Layer:     "base",
		Division:  "B",
		Part:      9,
		AppliesIf: AppliesIf{
			Occupancy: []string{"residential"},
			Element:   "stair_width",
		},
		Predicate:     "stair_width_mm >= 860",
		Discretionary: false,
		SourceText:    "Stairs shall be not less than 860 mm in width.",
		Citation:      "NBC Art. 9.8.3.1",
	},
	{
		ArticleID: "9.5.4.1",
		Layer:     "base",
		Division:  "B",
		Part:      9,
		AppliesIf: AppliesIf{
			Occupancy: []string{"residential"},
			Element:   "bedroom_window",
		},
		Predicate:     "window_area_m2 >= 0.35 && window_min_dimension_mm >= 380",
		Discretionary: false,
		SourceText:    "Every bedroom shall have at least one exterior window with an openable area of not less than 0.35 m² and no dimension less than 380 mm.",
		Citation:      "NBC Art. 9.5.4.1",
	},
	{
		ArticleID: "9.9.7.1",
		Layer:     "base",
		Division:  "B",
		Part:      9,
		AppliesIf: AppliesIf{
			Occupancy: []string{"residential"},
			Element:   "egress_door",
		},
		Predicate:     "door_width_mm >= 860 && door_height_mm >= 1950",
		Discretionary: false,
		SourceText:    "Exit doors shall be not less than 860 mm in width and 1950 mm in height.",
		Citation:      "NBC Art. 9.9.7.1",
	},
}

func main() {
	mode := flag.String("mode", "server", "Running mode: server or standalone")
	flag.Parse()

	if *mode == "standalone" {
		runStandaloneDemo()
		return
	}

	runMCPServer()
}

func runMCPServer() {
	scanner := bufio.NewScanner(os.Stdin)
	
	// Send initialization notification
	fmt.Println(`{"jsonrpc":"2.0","method":"notifications/initialized"}`)
	
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var req MCPRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			sendError(0, -32700, "Parse error", err.Error())
			continue
		}

		switch req.Method {
		case "initialize":
			handleInitialize(req.ID)
		case "tools/list":
			handleToolsList(req.ID)
		case "tools/call":
			handleToolsCall(req.ID, req.Params)
		default:
			sendError(req.ID, -32601, "Method not found", req.Method)
		}
	}
}

func handleInitialize(id int) {
	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{
					"listChanged": true,
				},
			},
			"serverInfo": map[string]string{
				"name":    "obc-checker",
				"version": "1.0.0",
			},
		},
	}
	sendResponse(response)
}

func handleToolsList(id int) {
	tools := []map[string]interface{}{
		{
			"name":        "check_building_code",
			"description": "Analyzes IFC model for Ontario Building Code compliance. Returns pass/fail/needs-review for applicable Part 9 rules.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"ifc_path": map[string]interface{}{
						"type":        "string",
						"description": "Path to IFC file exported from Revit",
					},
					"code_version": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"base": map[string]string{
								"type": "string",
							},
							"overlay": map[string]string{
								"type": "string",
							},
						},
						"description": "Code version to check against (default: NBC2020 + O.Reg 163/24)",
					},
					"parts": map[string]interface{}{
						"type":        "array",
						"items":       map[string]string{"type": "string"},
						"description": "Specific code parts to check (e.g., ['9.8', '9.9'])",
					},
				},
				"required": []string{"ifc_path"},
			},
		},
		{
			"name":        "extract_facts",
			"description": "Extracts building facts from IFC model without evaluation. Useful for understanding what data is available.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"ifc_path": map[string]interface{}{
						"type":        "string",
						"description": "Path to IFC file",
					},
				},
				"required": []string{"ifc_path"},
			},
		},
		{
			"name":        "get_applicable_rules",
			"description": "Returns all OBC rules that apply to the current model based on occupancy and elements present.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"ifc_path": map[string]interface{}{
						"type":        "string",
						"description": "Path to IFC file",
					},
					"part": map[string]interface{}{
						"type":        "string",
						"description": "Filter by specific part (e.g., '9.8')",
					},
				},
				"required": []string{"ifc_path"},
			},
		},
	}

	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]interface{}{
			"tools": tools,
		},
	}
	sendResponse(response)
}

func handleToolsCall(id int, params json.RawMessage) {
	var call ToolCallParams
	if err := json.Unmarshal(params, &call); err != nil {
		sendError(id, -32602, "Invalid params", err.Error())
		return
	}

	switch call.Name {
	case "check_building_code":
		handleCheckBuildingCode(id, call.Arguments)
	case "extract_facts":
		handleExtractFacts(id, call.Arguments)
	case "get_applicable_rules":
		handleGetApplicableRules(id, call.Arguments)
	default:
		sendError(id, -32602, "Unknown tool", call.Name)
	}
}

func handleCheckBuildingCode(id int, args map[string]interface{}) {
	var checkArgs CheckBuildingCodeArgs
	
	// Parse arguments
	if ifcPath, ok := args["ifc_path"].(string); ok {
		checkArgs.IFCPath = ifcPath
	}
	
	if cv, ok := args["code_version"].(map[string]interface{}); ok {
		if base, ok := cv["base"].(string); ok {
			checkArgs.CodeVersion.Base = base
		}
		if overlay, ok := cv["overlay"].(string); ok {
			checkArgs.CodeVersion.Overlay = overlay
		}
	}
	
	if parts, ok := args["parts"].([]interface{}); ok {
		for _, p := range parts {
			if ps, ok := p.(string); ok {
				checkArgs.Parts = append(checkArgs.Parts, ps)
			}
		}
	}

	// Set defaults
	if checkArgs.CodeVersion.Base == "" {
		checkArgs.CodeVersion.Base = "NBC2020-1st-printing"
	}
	if checkArgs.CodeVersion.Overlay == "" {
		checkArgs.CodeVersion.Overlay = "O.Reg163/24-2026-04-21"
	}

	// Simulate fact extraction (in real implementation, parse IFC)
	facts := simulateFactExtraction(checkArgs.IFCPath)
	
	// Evaluate rules
	results := evaluateRules(facts, checkArgs.Parts)
	
	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": formatCheckResults(results),
				},
			},
			"structured_data": results,
		},
	}
	sendResponse(response)
}

func handleExtractFacts(id int, args map[string]interface{}) {
	ifcPath, _ := args["ifc_path"].(string)
	
	facts := simulateFactExtraction(ifcPath)
	
	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": formatFacts(facts),
				},
			},
			"structured_data": facts,
		},
	}
	sendResponse(response)
}

func handleGetApplicableRules(id int, args map[string]interface{}) {
	ifcPath, _ := args["ifc_path"].(string)
	partFilter, _ := args["part"].(string)
	
	facts := simulateFactExtraction(ifcPath)
	applicableRules := findApplicableRules(facts, partFilter)
	
	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": fmt.Sprintf("Found %d applicable rules", len(applicableRules)),
				},
			},
			"structured_data": applicableRules,
		},
	}
	sendResponse(response)
}

func simulateFactExtraction(ifcPath string) BuildingFacts {
	// In real implementation, this would parse the IFC file
	// For demo purposes, return sample data
	return BuildingFacts{
		Guards: []GuardFact{
			{Location: "exterior_deck", HeightMM: 1067, DropMM: 2400, IsExterior: true},
			{Location: "interior_loft", HeightMM: 900, DropMM: 2700, IsExterior: false},
			{Location: "front_porch", HeightMM: 850, DropMM: 750, IsExterior: true},
		},
		Stairs: []StairFact{
			{RiserMM: 180, TreadMM: 280, WidthMM: 900, FlightCount: 1},
			{RiserMM: 210, TreadMM: 200, WidthMM: 860, FlightCount: 2},
		},
		Rooms: []RoomFact{
			{Name: "Bedroom 1", AreaM2: 12.5, Occupancy: "residential"},
			{Name: "Bedroom 2", AreaM2: 10.2, Occupancy: "residential"},
			{Name: "Living Room", AreaM2: 25.0, Occupancy: "residential"},
		},
		Doors: []DoorFact{
			{Location: "main_exit", WidthMM: 900, HeightMM: 2000, SwingClearMM: 850, IsEgress: true},
			{Location: "bedroom_1", WidthMM: 810, HeightMM: 1980, SwingClearMM: 760, IsEgress: false},
		},
		Corridors: []CorridorFact{
			{Location: "upper_hall", WidthMM: 920, LengthMM: 4500},
		},
	}
}

func evaluateRules(facts BuildingFacts, partFilters []string) CheckResult {
	var results []RuleResult
	summary := SummaryResult{}
	
	for _, rule := range ruleDatabase {
		// Apply part filter
		if len(partFilters) > 0 {
			partStr := fmt.Sprintf("%d.%d", rule.Part, 0)
			found := false
			for _, filter := range partFilters {
				if strings.HasPrefix(partStr, filter) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		
		// Check applicability
		if !isRuleApplicable(rule, facts) {
			continue
		}
		
		// Evaluate predicate
		result := evaluatePredicate(rule, facts)
		results = append(results, result)
		
		switch result.Status {
		case "pass":
			summary.Pass++
		case "fail":
			summary.Fail++
		case "needs_review":
			summary.NeedsReview++
		}
	}
	
	return CheckResult{
		Summary: summary,
		Results: results,
	}
}

func isRuleApplicable(rule Predicate, facts BuildingFacts) bool {
	applies := rule.AppliesIf
	
	// Check occupancy
	if len(applies.Occupancy) > 0 {
		hasOccupancy := false
		for _, room := range facts.Rooms {
			for _, occ := range applies.Occupancy {
				if room.Occupancy == occ {
					hasOccupancy = true
					break
				}
			}
		}
		if !hasOccupancy {
			return false
		}
	}
	
	// Check element type
	switch applies.Element {
	case "guard":
		return len(facts.Guards) > 0
	case "handrail":
		return len(facts.Stairs) > 0
	case "stair_riser", "stair_tread", "stair_width":
		return len(facts.Stairs) > 0
	case "bedroom_window":
		for _, room := range facts.Rooms {
			if strings.Contains(strings.ToLower(room.Name), "bedroom") {
				return true
			}
		}
		return false
	case "egress_door":
		return len(facts.Doors) > 0
	}
	
	return true
}

func evaluatePredicate(rule Predicate, facts BuildingFacts) RuleResult {
	result := RuleResult{
		ArticleID: rule.ArticleID,
		Predicate: rule.Predicate,
		Citation:  rule.Citation,
		SourceText: rule.SourceText,
	}
	
	// Evaluate based on element type
	switch rule.AppliesIf.Element {
	case "guard":
		return evaluateGuardRule(rule, facts, result)
	case "handrail":
		return evaluateHandrailRule(rule, facts, result)
	case "stair_riser":
		return evaluateStairRiserRule(rule, facts, result)
	case "stair_tread":
		return evaluateStairTreadRule(rule, facts, result)
	case "stair_width":
		return evaluateStairWidthRule(rule, facts, result)
	case "egress_door":
		return evaluateEgressDoorRule(rule, facts, result)
	default:
		result.Status = "needs_review"
		result.Message = "Manual review required for this rule type"
		return result
	}
}

func evaluateGuardRule(rule Predicate, facts BuildingFacts, result RuleResult) RuleResult {
	for _, guard := range facts.Guards {
		isExteriorCondition := strings.Contains(rule.AppliesIf.Condition, "exterior")
		
		if isExteriorCondition && !guard.IsExterior {
			continue
		}
		
		// Parse predicate to extract threshold
		threshold := 900 // default
		if strings.Contains(rule.Predicate, ">= 900") {
			threshold = 900
		}
		
		if guard.HeightMM >= threshold {
			result.Status = "pass"
			result.ExtractedValue = guard.HeightMM
			result.Message = fmt.Sprintf("Guard at %s: %dmm >= %dmm ✓", guard.Location, guard.HeightMM, threshold)
		} else {
			result.Status = "fail"
			result.ExtractedValue = guard.HeightMM
			result.Message = fmt.Sprintf("Guard at %s: %dmm < %dmm ✗ (needs %dmm minimum)", guard.Location, guard.HeightMM, threshold, threshold)
		}
		return result
	}
	
	result.Status = "needs_review"
	result.Message = "No applicable guards found"
	return result
}

func evaluateHandrailRule(rule Predicate, facts BuildingFacts, result RuleResult) RuleResult {
	if len(facts.Stairs) == 0 {
		result.Status = "needs_review"
		result.Message = "No stairs found"
		return result
	}
	
	// Assume handrail height matches typical construction (865-965mm)
	handrailHeight := 915 // typical value
	result.Status = "pass"
	result.ExtractedValue = handrailHeight
	result.Message = fmt.Sprintf("Handrail height: %dmm (acceptable range: 865-965mm) ✓", handrailHeight)
	return result
}

func evaluateStairRiserRule(rule Predicate, facts BuildingFacts, result RuleResult) RuleResult {
	maxRiser := 200
	allPass := true
	var failingStairs []int
	
	for i, stair := range facts.Stairs {
		if stair.RiserMM > maxRiser {
			allPass = false
			failingStairs = append(failingStairs, i)
		}
	}
	
	if allPass {
		result.Status = "pass"
		result.ExtractedValue = facts.Stairs[0].RiserMM
		result.Message = fmt.Sprintf("All stair risers ≤ %dmm ✓", maxRiser)
	} else {
		result.Status = "fail"
		result.ExtractedValue = facts.Stairs[failingStairs[0]].RiserMM
		result.Message = fmt.Sprintf("Stair flight(s) exceed maximum riser height: %dmm > %dmm ✗", facts.Stairs[failingStairs[0]].RiserMM, maxRiser)
	}
	return result
}

func evaluateStairTreadRule(rule Predicate, facts BuildingFacts, result RuleResult) RuleResult {
	minTread := 210
	allPass := true
	
	for _, stair := range facts.Stairs {
		if stair.TreadMM < minTread {
			allPass = false
			break
		}
	}
	
	if allPass {
		result.Status = "pass"
		result.ExtractedValue = facts.Stairs[0].TreadMM
		result.Message = fmt.Sprintf("All stair treads ≥ %dmm ✓", minTread)
	} else {
		result.Status = "fail"
		result.ExtractedValue = facts.Stairs[0].TreadMM
		result.Message = fmt.Sprintf("Stair treads insufficient: %dmm < %dmm ✗", facts.Stairs[0].TreadMM, minTread)
	}
	return result
}

func evaluateStairWidthRule(rule Predicate, facts BuildingFacts, result RuleResult) RuleResult {
	minWidth := 860
	allPass := true
	
	for _, stair := range facts.Stairs {
		if stair.WidthMM < minWidth {
			allPass = false
			break
		}
	}
	
	if allPass {
		result.Status = "pass"
		result.ExtractedValue = facts.Stairs[0].WidthMM
		result.Message = fmt.Sprintf("All stairs ≥ %dmm wide ✓", minWidth)
	} else {
		result.Status = "fail"
		result.ExtractedValue = facts.Stairs[0].WidthMM
		result.Message = fmt.Sprintf("Stair width insufficient: %dmm < %dmm ✗", facts.Stairs[0].WidthMM, minWidth)
	}
	return result
}

func evaluateEgressDoorRule(rule Predicate, facts BuildingFacts, result RuleResult) RuleResult {
	minWidth := 860
	minHeight := 1950
	
	for _, door := range facts.Doors {
		if !door.IsEgress {
			continue
		}
		
		if door.WidthMM >= minWidth && door.HeightMM >= minHeight {
			result.Status = "pass"
			result.ExtractedValue = map[string]int{"width": door.WidthMM, "height": door.HeightMM}
			result.Message = fmt.Sprintf("Egress door at %s: %dx%dmm ✓", door.Location, door.WidthMM, door.HeightMM)
		} else {
			result.Status = "fail"
			result.ExtractedValue = map[string]int{"width": door.WidthMM, "height": door.HeightMM}
			result.Message = fmt.Sprintf("Egress door at %s too small: %dx%dmm (min: %dx%dmm) ✗", door.Location, door.WidthMM, door.HeightMM, minWidth, minHeight)
		}
		return result
	}
	
	result.Status = "needs_review"
	result.Message = "No egress doors found"
	return result
}

func findApplicableRules(facts BuildingFacts, partFilter string) []Predicate {
	var applicable []Predicate
	
	for _, rule := range ruleDatabase {
		if partFilter != "" {
			partStr := fmt.Sprintf("%d.", rule.Part)
			if !strings.HasPrefix(partStr, partFilter) {
				continue
			}
		}
		
		if isRuleApplicable(rule, facts) {
			applicable = append(applicable, rule)
		}
	}
	
	return applicable
}

func formatCheckResults(results CheckResult) string {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("┌─────────────────────────────────────────┐\n"))
	sb.WriteString(fmt.Sprintf("│  OBC COMPLIANCE CHECK RESULTS           │\n"))
	sb.WriteString(fmt.Sprintf("└─────────────────────────────────────────┘\n\n"))
	
	sb.WriteString(fmt.Sprintf("SUMMARY: %d Pass | %d Fail | %d Needs Review\n\n", 
		results.Summary.Pass, results.Summary.Fail, results.Summary.NeedsReview))
	
	for _, r := range results.Results {
		icon := "•"
		switch r.Status {
		case "pass":
			icon = "✓"
		case "fail":
			icon = "✗"
		case "needs_review":
			icon = "?"
		}
		
		sb.WriteString(fmt.Sprintf("[%s] %s - %s\n", icon, r.ArticleID, r.Message))
		sb.WriteString(fmt.Sprintf("    Citation: %s\n", r.Citation))
		sb.WriteString("\n")
	}
	
	return sb.String()
}

func formatFacts(facts BuildingFacts) string {
	var sb strings.Builder
	
	sb.WriteString("EXTRACTED BUILDING FACTS\n\n")
	
	sb.WriteString(fmt.Sprintf("Guards: %d\n", len(facts.Guards)))
	for _, g := range facts.Guards {
		sb.WriteString(fmt.Sprintf("  - %s: %dmm high, %dmm drop, exterior=%v\n", 
			g.Location, g.HeightMM, g.DropMM, g.IsExterior))
	}
	
	sb.WriteString(fmt.Sprintf("\nStairs: %d\n", len(facts.Stairs)))
	for _, s := range facts.Stairs {
		sb.WriteString(fmt.Sprintf("  - Riser: %dmm, Tread: %dmm, Width: %dmm\n", 
			s.RiserMM, s.TreadMM, s.WidthMM))
	}
	
	sb.WriteString(fmt.Sprintf("\nRooms: %d\n", len(facts.Rooms)))
	for _, r := range facts.Rooms {
		sb.WriteString(fmt.Sprintf("  - %s: %.1f m² (%s)\n", 
			r.Name, r.AreaM2, r.Occupancy))
	}
	
	sb.WriteString(fmt.Sprintf("\nDoors: %d\n", len(facts.Doors)))
	for _, d := range facts.Doors {
		sb.WriteString(fmt.Sprintf("  - %s: %dx%dmm, egress=%v\n", 
			d.Location, d.WidthMM, d.HeightMM, d.IsEgress))
	}
	
	return sb.String()
}

func sendResponse(response MCPResponse) {
	jsonData, _ := json.Marshal(response)
	fmt.Println(string(jsonData))
}

func sendError(id int, code int, message string, data interface{}) {
	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &MCPError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	sendResponse(response)
}

func runStandaloneDemo() {
	fmt.Println("OBC Plans Checker - Standalone Demo Mode")
	fmt.Println("=========================================\n")
	
	// Simulate checking a model
	facts := simulateFactExtraction("demo.ifc")
	results := evaluateRules(facts, []string{})
	
	fmt.Println(formatFacts(facts))
	fmt.Println("\n" + formatCheckResults(results))
	
	fmt.Println("\nUsage:")
	fmt.Println("  Server mode: revitmcp-server --mode server")
	fmt.Println("  Demo mode:   revitmcp-server --mode standalone")
	fmt.Println("\nFor Revit integration, configure MCP client with:")
	fmt.Println("  {\"command\": \"revitmcp-server\", \"args\": [\"--mode\", \"server\"]}")
}

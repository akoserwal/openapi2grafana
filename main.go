package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

// Config holds the configuration for dashboard generation
type Config struct {
	InputFile      string
	OutputFile     string
	DashboardUID   string
	DashboardTitle string
	DataSource     string
	Environment    string
	UpdateMode     bool
	IncludeGRPC    bool
}

// DashboardMetadata tracks dashboard versions and updates
type DashboardMetadata struct {
	Version     int       `json:"version"`
	Generated   time.Time `json:"generated"`
	SpecHash    string    `json:"spec_hash"`
	LastUpdated time.Time `json:"last_updated"`
}

type GrafanaDashboard struct {
	Title         string            `json:"title"`
	Panels        []Panel           `json:"panels"`
	Templating    Templating        `json:"templating"`
	Time          Time              `json:"time"`
	Timepicker    Timepicker        `json:"timepicker"`
	Tags          []string          `json:"tags"`
	Style         string            `json:"style"`
	Editable      bool              `json:"editable"`
	UID           string            `json:"uid"`
	SchemaVersion int               `json:"schemaVersion"`
	Version       int               `json:"version"`
	Annotations   Annotations       `json:"annotations"`
	Links         []Link            `json:"links"`
	Refresh       string            `json:"refresh"`
	Meta          DashboardMetadata `json:"meta"`
}

type Templating struct {
	List []Variable `json:"list"`
}

type Annotations struct {
	List []Annotation `json:"list"`
}

type Annotation struct {
	BuiltIn    int    `json:"builtIn"`
	Datasource string `json:"datasource"`
	Enable     bool   `json:"enable"`
	Hide       bool   `json:"hide"`
	IconColor  string `json:"iconColor"`
	Name       string `json:"name"`
	Type       string `json:"type"`
}

type Link struct {
	AsDropdown  bool     `json:"asDropdown"`
	Icon        string   `json:"icon"`
	IncludeVars bool     `json:"includeVars"`
	KeepTime    bool     `json:"keepTime"`
	Tags        []string `json:"tags"`
	Title       string   `json:"title"`
	Type        string   `json:"type"`
	URL         string   `json:"url"`
}

type Panel struct {
	Title       string           `json:"title"`
	Type        string           `json:"type"`
	Datasource  interface{}      `json:"datasource"`
	Targets     []Target         `json:"targets"`
	GridPos     GridPos          `json:"gridPos"`
	Options     Options          `json:"options"`
	FieldConfig FieldConfig      `json:"fieldConfig"`
	ID          int              `json:"id"`
	Transparent bool             `json:"transparent,omitempty"`
	Collapsed   bool             `json:"collapsed,omitempty"`
	Panels      []Panel          `json:"panels,omitempty"`
	Description string           `json:"description,omitempty"`
	Thresholds  *PanelThresholds `json:"thresholds,omitempty"`
	Alert       *Alert           `json:"alert,omitempty"`
}

type PanelThresholds struct {
	Mode  string      `json:"mode"`
	Steps []Threshold `json:"steps"`
}

type Threshold struct {
	Color string  `json:"color"`
	Value float64 `json:"value"`
}

type Alert struct {
	Name                string              `json:"name"`
	Message             string              `json:"message"`
	Frequency           string              `json:"frequency"`
	Conditions          []AlertCondition    `json:"conditions"`
	ExecutionErrorState string              `json:"executionErrorState"`
	For                 string              `json:"for"`
	NoDataState         string              `json:"noDataState"`
	Notifications       []AlertNotification `json:"notifications"`
}

type AlertCondition struct {
	Evaluator AlertEvaluator `json:"evaluator"`
	Operator  AlertOperator  `json:"operator"`
	Query     AlertQuery     `json:"query"`
	Reducer   AlertReducer   `json:"reducer"`
	Type      string         `json:"type"`
}

type AlertEvaluator struct {
	Params []float64 `json:"params"`
	Type   string    `json:"type"`
}

type AlertOperator struct {
	Type string `json:"type"`
}

type AlertQuery struct {
	Model     Target   `json:"model"`
	Params    []string `json:"params"`
	QueryType string   `json:"queryType"`
}

type AlertReducer struct {
	Params []string `json:"params"`
	Type   string   `json:"type"`
}

type AlertNotification struct {
	ID int `json:"id"`
}

type Target struct {
	Expr           string `json:"expr"`
	LegendFormat   string `json:"legendFormat"`
	RefID          string `json:"refId"`
	Interval       string `json:"interval,omitempty"`
	IntervalFactor int    `json:"intervalFactor,omitempty"`
	Step           int    `json:"step,omitempty"`
	Format         string `json:"format,omitempty"`
	Instant        bool   `json:"instant,omitempty"`
	Hide           bool   `json:"hide,omitempty"`
	Exemplar       bool   `json:"exemplar,omitempty"`
}

type GridPos struct {
	H int `json:"h"`
	W int `json:"w"`
	X int `json:"x"`
	Y int `json:"y"`
}

type Options struct {
	Legend               LegendOptions  `json:"legend"`
	Tooltip              TooltipOptions `json:"tooltip"`
	DisplayMode          string         `json:"displayMode,omitempty"`
	Orientation          string         `json:"orientation,omitempty"`
	ReduceOptions        ReduceOptions  `json:"reduceOptions,omitempty"`
	ShowThresholdLabels  bool           `json:"showThresholdLabels,omitempty"`
	ShowThresholdMarkers bool           `json:"showThresholdMarkers,omitempty"`
	Text                 TextOptions    `json:"text,omitempty"`
}

type LegendOptions struct {
	DisplayMode string   `json:"displayMode"`
	Placement   string   `json:"placement"`
	Values      []string `json:"values,omitempty"`
}

type TooltipOptions struct {
	Mode string `json:"mode"`
}

type ReduceOptions struct {
	Values bool     `json:"values"`
	Fields string   `json:"fields"`
	Calcs  []string `json:"calcs"`
}

type TextOptions struct {
	TitleSize int `json:"titleSize,omitempty"`
	ValueSize int `json:"valueSize,omitempty"`
}

type FieldConfig struct {
	Defaults  FieldConfigDefaults `json:"defaults"`
	Overrides []FieldOverride     `json:"overrides"`
}

type FieldConfigDefaults struct {
	Color       ColorOptions     `json:"color"`
	Thresholds  ThresholdOptions `json:"thresholds"`
	Unit        string           `json:"unit,omitempty"`
	Min         *float64         `json:"min,omitempty"`
	Max         *float64         `json:"max,omitempty"`
	Decimals    *int             `json:"decimals,omitempty"`
	DisplayName string           `json:"displayName,omitempty"`
}

type FieldOverride struct {
	Matcher    FieldMatcher    `json:"matcher"`
	Properties []FieldProperty `json:"properties"`
}

type FieldMatcher struct {
	ID      string `json:"id"`
	Options string `json:"options"`
}

type FieldProperty struct {
	ID    string      `json:"id"`
	Value interface{} `json:"value"`
}

type ColorOptions struct {
	Mode string `json:"mode"`
}

type ThresholdOptions struct {
	Mode  string          `json:"mode"`
	Steps []ThresholdStep `json:"steps"`
}

type ThresholdStep struct {
	Color string   `json:"color"`
	Value *float64 `json:"value"`
}

type Variable struct {
	Name        string   `json:"name"`
	Label       string   `json:"label"`
	Query       string   `json:"query"`
	Current     Current  `json:"current"`
	Type        string   `json:"type"`
	Options     []Option `json:"options"`
	Datasource  string   `json:"datasource,omitempty"`
	Refresh     int      `json:"refresh"`
	IncludeAll  bool     `json:"includeAll"`
	AllValue    string   `json:"allValue,omitempty"`
	Sort        int      `json:"sort,omitempty"`
	Multi       bool     `json:"multi,omitempty"`
	Definition  string   `json:"definition,omitempty"`
	Description string   `json:"description,omitempty"`
	Hide        int      `json:"hide,omitempty"`
}

type Current struct {
	Text     interface{} `json:"text"`
	Value    interface{} `json:"value"`
	Selected bool        `json:"selected,omitempty"`
}

type Option struct {
	Text     string `json:"text"`
	Value    string `json:"value"`
	Selected bool   `json:"selected,omitempty"`
}

type Time struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type Timepicker struct {
	RefreshIntervals []string `json:"refresh_intervals"`
	TimeOptions      []string `json:"time_options"`
}

func main() {
	config := parseArgs()

	if err := generateDashboardFromConfig(config); err != nil {
		log.Fatalf("Error generating dashboard: %v", err)
	}
}

func parseArgs() *Config {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <openapi-spec-file> [output-file] [--update] [--uid <uid>]")
	}

	config := &Config{
		InputFile:      os.Args[1],
		OutputFile:     "grafana_dashboard.json",
		DashboardUID:   "generated-api-dashboard",
		DashboardTitle: "API Monitoring Dashboard",
		DataSource:     "prometheus",
		Environment:    "production",
		UpdateMode:     false,
		IncludeGRPC:    true,
	}

	// Parse additional arguments
	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--update":
			config.UpdateMode = true
		case "--uid":
			if i+1 < len(os.Args) {
				config.DashboardUID = os.Args[i+1]
				i++
			}
		case "--datasource":
			if i+1 < len(os.Args) {
				config.DataSource = os.Args[i+1]
				i++
			}
		case "--title":
			if i+1 < len(os.Args) {
				config.DashboardTitle = os.Args[i+1]
				i++
			}
		default:
			// If not a flag, treat as output file
			if !strings.HasPrefix(os.Args[i], "--") {
				config.OutputFile = os.Args[i]
			}
		}
	}

	return config
}

func generateDashboardFromConfig(config *Config) error {
	// Load OpenAPI spec
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(config.InputFile)
	if err != nil {
		return fmt.Errorf("error loading OpenAPI spec: %w", err)
	}

	// Calculate spec hash for versioning
	specHash, err := calculateSpecHash(config.InputFile)
	if err != nil {
		return fmt.Errorf("error calculating spec hash: %w", err)
	}

	// Check if dashboard exists and should be updated
	var existingDashboard *GrafanaDashboard
	if config.UpdateMode {
		existingDashboard, _ = loadExistingDashboard(config.OutputFile)
	}

	// Generate new dashboard
	dashboard := generateDashboard(doc, config, specHash, existingDashboard)

	// Save dashboard to file
	dashboardJSON, err := json.MarshalIndent(dashboard, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling dashboard: %w", err)
	}

	err = os.WriteFile(config.OutputFile, dashboardJSON, 0644)
	if err != nil {
		return fmt.Errorf("error writing dashboard file: %w", err)
	}

	fmt.Printf("Successfully generated Grafana dashboard: %s\n", config.OutputFile)
	if config.UpdateMode && existingDashboard != nil {
		fmt.Printf("Dashboard updated from version %d to %d\n", existingDashboard.Version, dashboard.Version)
	}
	return nil
}

func calculateSpecHash(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

func loadExistingDashboard(filePath string) (*GrafanaDashboard, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var dashboard GrafanaDashboard
	if err := json.Unmarshal(data, &dashboard); err != nil {
		return nil, err
	}

	return &dashboard, nil
}

func generateDashboard(doc *openapi3.T, config *Config, specHash string, existingDashboard *GrafanaDashboard) GrafanaDashboard {
	title := config.DashboardTitle
	if doc.Info != nil && doc.Info.Title != "" {
		title = doc.Info.Title + " Monitoring"
	}

	version := 1
	if existingDashboard != nil {
		version = existingDashboard.Version + 1
	}

	dashboard := GrafanaDashboard{
		Title:         title,
		Editable:      true,
		Style:         "dark",
		Tags:          []string{"generated", "api", "monitoring"},
		UID:           config.DashboardUID,
		SchemaVersion: 30,
		Version:       version,
		Refresh:       "30s",
		Time: Time{
			From: "now-6h",
			To:   "now",
		},
		Timepicker: Timepicker{
			RefreshIntervals: []string{"5s", "10s", "30s", "1m", "5m", "15m", "30m", "1h", "2h", "1d"},
			TimeOptions:      []string{"5m", "15m", "1h", "6h", "12h", "24h", "2d", "7d", "30d"},
		},
		Templating: Templating{
			List: []Variable{
				{
					Name:    "datasource",
					Label:   "Data Source",
					Type:    "datasource",
					Current: Current{Text: config.DataSource, Value: config.DataSource},
					Options: []Option{
						{Text: config.DataSource, Value: config.DataSource, Selected: true},
					},
					Query:      "prometheus",
					IncludeAll: false,
					Multi:      false,
					Refresh:    1,
					Hide:       0,
				},
				{
					Name:    "environment",
					Label:   "Environment",
					Type:    "custom",
					Current: Current{Text: "All", Value: "$__all"},
					Options: []Option{
						{Text: "All", Value: "$__all", Selected: true},
						{Text: "Production", Value: "prod"},
						{Text: "Staging", Value: "stage"},
						{Text: "Development", Value: "dev"},
					},
					IncludeAll: true,
					AllValue:   ".*",
					Multi:      true,
					Refresh:    0,
				},
				{
					Name:        "service",
					Label:       "Service",
					Type:        "query",
					Query:       "label_values(http_requests_total, service)",
					Current:     Current{Text: "All", Value: "$__all"},
					Datasource:  config.DataSource,
					IncludeAll:  true,
					AllValue:    ".*",
					Multi:       true,
					Refresh:     1,
					Sort:        1,
					Definition:  "label_values(http_requests_total, service)",
					Description: "Service name filter",
				},
			},
		},
		Annotations: Annotations{
			List: []Annotation{
				{
					BuiltIn:    1,
					Datasource: "-- Grafana --",
					Enable:     true,
					Hide:       true,
					IconColor:  "rgba(0, 211, 255, 1)",
					Name:       "Annotations & Alerts",
					Type:       "dashboard",
				},
			},
		},
		Links: []Link{
			{
				AsDropdown:  false,
				Icon:        "external link",
				IncludeVars: false,
				KeepTime:    false,
				Tags:        []string{"api", "monitoring"},
				Title:       "API Documentation",
				Type:        "dashboards",
			},
		},
		Meta: DashboardMetadata{
			Version:     version,
			Generated:   time.Now(),
			SpecHash:    specHash,
			LastUpdated: time.Now(),
		},
	}

	// Track panel positions
	panelY := 0
	panelHeight := 8
	panelID := 1

	// Add panels for HTTP endpoints
	for path, pathItem := range doc.Paths.Map() {
		for method, operation := range pathItem.Operations() {
			panelTitle := fmt.Sprintf("%s %s", strings.ToUpper(method), path)
			if operation.Summary != "" {
				panelTitle = fmt.Sprintf("%s: %s", panelTitle, operation.Summary)
			}

			// Request Rate panel
			requestRatePanel := createRequestRatePanel(panelTitle, path, method, panelID, panelHeight, panelY)
			dashboard.Panels = append(dashboard.Panels, requestRatePanel)
			panelID++
			panelY += panelHeight

			// Enhanced Latency panel with P50, P90, P95, P99
			latencyPanel := createLatencyPanel(panelTitle, path, method, panelID, panelHeight, panelY)
			dashboard.Panels = append(dashboard.Panels, latencyPanel)
			panelID++
			panelY += panelHeight

			// Error rate panel
			errorRatePanel := createErrorRatePanel(panelTitle, path, method, panelID, panelHeight, panelY)
			dashboard.Panels = append(dashboard.Panels, errorRatePanel)
			panelID++
			panelY += panelHeight

			// Throughput panel
			throughputPanel := createThroughputPanel(panelTitle, path, method, panelID, panelHeight, panelY)
			dashboard.Panels = append(dashboard.Panels, throughputPanel)
			panelID++
			panelY += panelHeight
		}
	}

	// Add gRPC panels if gRPC extensions exist and enabled
	if config.IncludeGRPC && doc.Extensions != nil {
		if grpcExt, ok := doc.Extensions["x-grpc"]; ok {
			if grpcServices, ok := grpcExt.(map[string]interface{}); ok {
				for serviceName, methods := range grpcServices {
					if methodMap, ok := methods.(map[string]interface{}); ok {
						for methodName := range methodMap {
							panelTitle := fmt.Sprintf("gRPC %s/%s", serviceName, methodName)

							// gRPC Request Rate panel
							grpcRequestPanel := createGRPCRequestPanel(panelTitle, serviceName, methodName, panelID, panelHeight, panelY)
							dashboard.Panels = append(dashboard.Panels, grpcRequestPanel)
							panelID++
							panelY += panelHeight

							// gRPC Latency panel
							grpcLatencyPanel := createGRPCLatencyPanel(panelTitle, serviceName, methodName, panelID, panelHeight, panelY)
							dashboard.Panels = append(dashboard.Panels, grpcLatencyPanel)
							panelID++
							panelY += panelHeight
						}
					}
				}
			}
		}
	}

	return dashboard
}

func createRequestRatePanel(title, path, method string, panelID, height, yPos int) Panel {
	return Panel{
		ID:         panelID,
		Title:      title + " - Request Rate",
		Type:       "timeseries",
		Datasource: map[string]string{"type": "prometheus", "uid": "${datasource}"},
		GridPos:    GridPos{H: height, W: 12, X: 0, Y: yPos},
		Targets: []Target{
			{
				Expr:         fmt.Sprintf(`sum(rate(http_requests_total{path="%s", method="%s", service=~"$service"}[$__rate_interval])) by (status_code)`, path, method),
				LegendFormat: "Status {{status_code}}",
				RefID:        "A",
			},
		},
		Options: Options{
			Legend: LegendOptions{
				DisplayMode: "list",
				Placement:   "bottom",
			},
			Tooltip: TooltipOptions{
				Mode: "multi",
			},
		},
		FieldConfig: FieldConfig{
			Defaults: FieldConfigDefaults{
				Color: ColorOptions{Mode: "palette-classic"},
				Unit:  "reqps",
				Thresholds: ThresholdOptions{
					Mode: "absolute",
					Steps: []ThresholdStep{
						{Color: "green", Value: nil},
						{Color: "red", Value: floatPtr(80)},
					},
				},
			},
		},
		Description: "Request rate per status code",
	}
}

func createLatencyPanel(title, path, method string, panelID, height, yPos int) Panel {
	return Panel{
		ID:         panelID,
		Title:      title + " - Latency Percentiles",
		Type:       "timeseries",
		Datasource: map[string]string{"type": "prometheus", "uid": "${datasource}"},
		GridPos:    GridPos{H: height, W: 12, X: 12, Y: yPos},
		Targets: []Target{
			{
				Expr:         fmt.Sprintf(`histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket{path="%s", method="%s", service=~"$service"}[$__rate_interval])) by (le))`, path, method),
				LegendFormat: "p99",
				RefID:        "A",
			},
			{
				Expr:         fmt.Sprintf(`histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{path="%s", method="%s", service=~"$service"}[$__rate_interval])) by (le))`, path, method),
				LegendFormat: "p95",
				RefID:        "B",
			},
			{
				Expr:         fmt.Sprintf(`histogram_quantile(0.90, sum(rate(http_request_duration_seconds_bucket{path="%s", method="%s", service=~"$service"}[$__rate_interval])) by (le))`, path, method),
				LegendFormat: "p90",
				RefID:        "C",
			},
			{
				Expr:         fmt.Sprintf(`histogram_quantile(0.50, sum(rate(http_request_duration_seconds_bucket{path="%s", method="%s", service=~"$service"}[$__rate_interval])) by (le))`, path, method),
				LegendFormat: "p50",
				RefID:        "D",
			},
		},
		Options: Options{
			Legend: LegendOptions{
				DisplayMode: "list",
				Placement:   "bottom",
			},
			Tooltip: TooltipOptions{
				Mode: "multi",
			},
		},
		FieldConfig: FieldConfig{
			Defaults: FieldConfigDefaults{
				Color: ColorOptions{Mode: "palette-classic"},
				Unit:  "s",
				Thresholds: ThresholdOptions{
					Mode: "absolute",
					Steps: []ThresholdStep{
						{Color: "green", Value: nil},
						{Color: "yellow", Value: floatPtr(0.5)},
						{Color: "red", Value: floatPtr(1.0)},
					},
				},
			},
		},
		Description: "Response time percentiles",
	}
}

func createErrorRatePanel(title, path, method string, panelID, height, yPos int) Panel {
	return Panel{
		ID:         panelID,
		Title:      title + " - Error Rate",
		Type:       "stat",
		Datasource: map[string]string{"type": "prometheus", "uid": "${datasource}"},
		GridPos:    GridPos{H: height, W: 6, X: 0, Y: yPos},
		Targets: []Target{
			{
				Expr:         fmt.Sprintf(`sum(rate(http_requests_total{path="%s", method="%s", status_code=~"5..", service=~"$service"}[$__rate_interval])) / sum(rate(http_requests_total{path="%s", method="%s", service=~"$service"}[$__rate_interval])) * 100`, path, method, path, method),
				LegendFormat: "Error Rate",
				RefID:        "A",
			},
		},
		Options: Options{
			ReduceOptions: ReduceOptions{
				Values: false,
				Fields: "",
				Calcs:  []string{"lastNotNull"},
			},
			Orientation: "auto",
			Text: TextOptions{
				TitleSize: 10,
				ValueSize: 18,
			},
			ShowThresholdLabels:  false,
			ShowThresholdMarkers: true,
		},
		FieldConfig: FieldConfig{
			Defaults: FieldConfigDefaults{
				Color: ColorOptions{Mode: "thresholds"},
				Unit:  "percent",
				Max:   floatPtr(100),
				Min:   floatPtr(0),
				Thresholds: ThresholdOptions{
					Mode: "absolute",
					Steps: []ThresholdStep{
						{Color: "green", Value: nil},
						{Color: "yellow", Value: floatPtr(1)},
						{Color: "red", Value: floatPtr(5)},
					},
				},
			},
		},
		Description: "5xx error rate percentage",
	}
}

func createThroughputPanel(title, path, method string, panelID, height, yPos int) Panel {
	return Panel{
		ID:         panelID,
		Title:      title + " - Throughput",
		Type:       "stat",
		Datasource: map[string]string{"type": "prometheus", "uid": "${datasource}"},
		GridPos:    GridPos{H: height, W: 6, X: 6, Y: yPos},
		Targets: []Target{
			{
				Expr:         fmt.Sprintf(`sum(rate(http_requests_total{path="%s", method="%s", service=~"$service"}[$__rate_interval]))`, path, method),
				LegendFormat: "Throughput",
				RefID:        "A",
			},
		},
		Options: Options{
			ReduceOptions: ReduceOptions{
				Values: false,
				Fields: "",
				Calcs:  []string{"lastNotNull"},
			},
			Orientation: "auto",
			Text: TextOptions{
				TitleSize: 10,
				ValueSize: 18,
			},
			ShowThresholdLabels:  false,
			ShowThresholdMarkers: true,
		},
		FieldConfig: FieldConfig{
			Defaults: FieldConfigDefaults{
				Color: ColorOptions{Mode: "palette-classic"},
				Unit:  "reqps",
				Thresholds: ThresholdOptions{
					Mode: "absolute",
					Steps: []ThresholdStep{
						{Color: "green", Value: nil},
					},
				},
			},
		},
		Description: "Total requests per second",
	}
}

func floatPtr(f float64) *float64 {
	return &f
}

func createGRPCRequestPanel(title, service, method string, panelID, height, yPos int) Panel {
	return Panel{
		ID:         panelID,
		Title:      title + " - Request Rate",
		Type:       "timeseries",
		Datasource: map[string]string{"type": "prometheus", "uid": "${datasource}"},
		GridPos:    GridPos{H: height, W: 12, X: 0, Y: yPos},
		Targets: []Target{
			{
				Expr:         fmt.Sprintf(`sum(rate(grpc_server_handled_total{grpc_service="%s", grpc_method="%s"}[$__rate_interval])) by (grpc_code)`, service, method),
				LegendFormat: "Code {{grpc_code}}",
				RefID:        "A",
			},
		},
		Options: Options{
			Legend: LegendOptions{
				DisplayMode: "list",
				Placement:   "bottom",
			},
			Tooltip: TooltipOptions{
				Mode: "multi",
			},
		},
		FieldConfig: FieldConfig{
			Defaults: FieldConfigDefaults{
				Color: ColorOptions{Mode: "palette-classic"},
				Unit:  "reqps",
				Thresholds: ThresholdOptions{
					Mode: "absolute",
					Steps: []ThresholdStep{
						{Color: "green", Value: nil},
						{Color: "red", Value: floatPtr(80)},
					},
				},
			},
		},
		Description: "gRPC request rate per status code",
	}
}

func createGRPCLatencyPanel(title, service, method string, panelID, height, yPos int) Panel {
	return Panel{
		ID:         panelID,
		Title:      title + " - Latency",
		Type:       "timeseries",
		Datasource: map[string]string{"type": "prometheus", "uid": "${datasource}"},
		GridPos:    GridPos{H: height, W: 12, X: 12, Y: yPos},
		Targets: []Target{
			{
				Expr:         fmt.Sprintf(`histogram_quantile(0.99, sum(rate(grpc_server_handling_seconds_bucket{grpc_service="%s", grpc_method="%s"}[$__rate_interval])) by (le))`, service, method),
				LegendFormat: "p99",
				RefID:        "A",
			},
			{
				Expr:         fmt.Sprintf(`histogram_quantile(0.95, sum(rate(grpc_server_handling_seconds_bucket{grpc_service="%s", grpc_method="%s"}[$__rate_interval])) by (le))`, service, method),
				LegendFormat: "p95",
				RefID:        "B",
			},
			{
				Expr:         fmt.Sprintf(`histogram_quantile(0.90, sum(rate(grpc_server_handling_seconds_bucket{grpc_service="%s", grpc_method="%s"}[$__rate_interval])) by (le))`, service, method),
				LegendFormat: "p90",
				RefID:        "C",
			},
			{
				Expr:         fmt.Sprintf(`histogram_quantile(0.50, sum(rate(grpc_server_handling_seconds_bucket{grpc_service="%s", grpc_method="%s"}[$__rate_interval])) by (le))`, service, method),
				LegendFormat: "p50",
				RefID:        "D",
			},
		},
		Options: Options{
			Legend: LegendOptions{
				DisplayMode: "list",
				Placement:   "bottom",
			},
			Tooltip: TooltipOptions{
				Mode: "multi",
			},
		},
		FieldConfig: FieldConfig{
			Defaults: FieldConfigDefaults{
				Color: ColorOptions{Mode: "palette-classic"},
				Unit:  "s",
				Thresholds: ThresholdOptions{
					Mode: "absolute",
					Steps: []ThresholdStep{
						{Color: "green", Value: nil},
						{Color: "yellow", Value: floatPtr(0.5)},
						{Color: "red", Value: floatPtr(1.0)},
					},
				},
			},
		},
		Description: "gRPC response time percentiles",
	}
}

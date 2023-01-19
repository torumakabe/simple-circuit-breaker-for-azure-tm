package alert

import "time"

type Payload struct {
	SchemaID string `json:"schemaId"`
	Data     Data   `json:"data"`
}

type Data struct {
	Essentials   Essentials   `json:"essentials"`
	AlertContext AlertContext `json:"alertContext,omitempty"`
}

type Essentials struct {
	AlertID             string    `json:"alertId,omitempty"`
	AlertRule           string    `json:"alertRule,omitempty"`
	Severity            string    `json:"severity,omitempty"`
	SignalType          string    `json:"signalType,omitempty"`
	MonitorCondition    string    `json:"monitorCondition,omitempty"`
	MonitoringService   string    `json:"monitoringService,omitempty"`
	AlertTargetIDs      []string  `json:"alertTargetIDs,omitempty"`
	ConfigurationItems  []string  `json:"configurationItems,omitempty"`
	OriginAlertID       string    `json:"originAlertId,omitempty"`
	FiredDateTime       time.Time `json:"firedDateTime,omitempty"`
	ResolvedDateTime    time.Time `json:"resolvedDateTime,omitempty"`
	Description         string    `json:"description,omitempty"`
	EssentialsVersion   string    `json:"essentialsVersion,omitempty"`
	AlertContextVersion string    `json:"alertContextVersion,omitempty"`
}

type AlertContext struct {
	Properties    interface{} `json:"properties,omitempty"`
	ConditionType string      `json:"conditionType,omitempty"`
	Condition     Condition   `json:"condition,omitempty"`
}

type Condition struct {
	WindowSize string  `json:"windowSize,omitempty"`
	AllOf      []AllOf `json:"allOf,omitempty"`
}

type AllOf struct {
	MetricName      string       `json:"metricName,omitempty"`
	MetricNamespace string       `json:"metricNamespace,omitempty"`
	Operator        string       `json:"operator,omitempty"`
	Threshold       string       `json:"threshold,omitempty"`
	TimeAggregation string       `json:"timeAggregation,omitempty"`
	Dimensions      []Dimensions `json:"dimensions,omitempty"`
	MetricValue     float64      `json:"metricValue,omitempty"`
}

type Dimensions struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

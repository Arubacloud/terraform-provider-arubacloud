package types

import "time"

// ActionType represents the type of alert action
type ActionType string

const (
	ActionTypeNotificationPanel ActionType = "NotificationPanel"
	ActionTypeSendEmail         ActionType = "SendEmail"
	ActionTypeSendSMS           ActionType = "SendSms"
	ActionTypeAutoscalingDBaaS  ActionType = "AutoscalingDbaas"
)

// ExecutedAlertActionResponse represents an executed alert action
type ExecutedAlertActionResponse struct {
	ActionType   ActionType `json:"actionType,omitempty"`
	Success      bool       `json:"success,omitempty"`
	ErrorMessage string     `json:"errorMessage,omitempty"`
}

// AlertActionResponse represents a possible alert action
type AlertActionResponse struct {
	Key        string `json:"key,omitempty"`
	Disabled   bool   `json:"disabled,omitempty"`
	Executable bool   `json:"executable,omitempty"`
}

// AlertResponse represents an alert response
type AlertResponse struct {
	ID                   string                        `json:"id,omitempty"`
	EventID              string                        `json:"eventId,omitempty"`
	EventName            string                        `json:"eventName,omitempty"`
	Username             string                        `json:"username,omitempty"`
	ServiceCategory      string                        `json:"serviceCategory,omitempty"`
	ServiceTypology      string                        `json:"serviceTypology,omitempty"`
	ResourceID           string                        `json:"resourceId,omitempty"`
	ServiceName          string                        `json:"serviceName,omitempty"`
	ResourceTypology     string                        `json:"resourceTypology,omitempty"`
	Metric               string                        `json:"metric,omitempty"`
	LastReception        time.Time                     `json:"lastReception,omitempty"`
	Rule                 string                        `json:"rule,omitempty"`
	Theshold             int64                         `json:"theshold,omitempty"`
	UM                   string                        `json:"um,omitempty"`
	Duration             string                        `json:"duration,omitempty"`
	ThesholdExceedence   string                        `json:"thesholdExceedence,omitempty"`
	Component            string                        `json:"component,omitempty"`
	ClusterTypology      string                        `json:"clusterTypology,omitempty"`
	Cluster              string                        `json:"cluster,omitempty"`
	Clustername          string                        `json:"clustername,omitempty"`
	NodePool             string                        `json:"nodePool,omitempty"`
	SMS                  bool                          `json:"sms,omitempty"`
	Email                bool                          `json:"email,omitempty"`
	Panel                bool                          `json:"panel,omitempty"`
	Hidden               bool                          `json:"hidden,omitempty"`
	ExecutedAlertActions []ExecutedAlertActionResponse `json:"executedAlertActions,omitempty"`
	Actions              []AlertActionResponse         `json:"actions,omitempty"`
}

// AlertsListResponse represents a list of alerts
type AlertsListResponse struct {
	ListResponse
	Values []AlertResponse `json:"values"`
}

package types

// JobType represents the type of job
type JobType string

const (
	// JobTypeOneShot represents a one-time job
	JobTypeOneShot JobType = "OneShot"

	// JobTypeRecurring represents a recurring job
	JobTypeRecurring JobType = "Recurring"
)

// RecurrenceType represents the recurrence pattern of a job
type RecurrenceType string

const (
	RecurrenceTypeHourly  RecurrenceType = "Hourly"
	RecurrenceTypeDaily   RecurrenceType = "Daily"
	RecurrenceTypeWeekly  RecurrenceType = "Weekly"
	RecurrenceTypeMonthly RecurrenceType = "Monthly"
	RecurrenceTypeCustom  RecurrenceType = "Custom"
)

// DeactiveReasonDto represents the reason why a job was deactivated
type DeactiveReasonDto string

const (
	DeactiveReasonNone            DeactiveReasonDto = "None"
	DeactiveReasonManual          DeactiveReasonDto = "Manual"
	DeactiveReasonResourceDeleted DeactiveReasonDto = "ResourceDeleted"
)

// HTTPVerb is the HTTP method for a job step action.
type HTTPVerb string

const (
	HTTPVerbGET    HTTPVerb = "GET"
	HTTPVerbPOST   HTTPVerb = "POST"
	HTTPVerbPUT    HTTPVerb = "PUT"
	HTTPVerbDELETE HTTPVerb = "DELETE"
	HTTPVerbPATCH  HTTPVerb = "PATCH"
)

// JobStepRequest represents a step that will be executed as part of the job.
type JobStepRequest struct {
	// Name Descriptive name of the step (nullable)
	// For more information, check the documentation.
	Name *string `json:"name,omitempty"`

	// ResourceURI URI of the resource on which the action will be performed
	ResourceURI string `json:"resourceUri"`

	// ActionURI URI of the action to execute on the resource
	// For more information, check the documentation.
	ActionURI string `json:"actionUri"`

	// HttpVerb HTTP verb to be used for the action (e.g., GET, POST, PUT, DELETE)
	// For more information, check the documentation.
	HttpVerb HTTPVerb `json:"httpVerb"`

	// Body Optional HTTP request body to send with the action (nullable)
	// For more information, check the documentation.
	Body *string `json:"body,omitempty"`
}

// JobStepResponse represents a step in the response with additional fields
type JobStepResponse struct {
	// Name Descriptive name of the step (nullable)
	Name *string `json:"name,omitempty"`

	// ResourceURI URI of the resource (nullable)
	ResourceURI *string `json:"resourceUri,omitempty"`

	// ActionURI URI of the action (nullable)
	ActionURI *string `json:"actionUri,omitempty"`

	// ActionName Name of the action (nullable)
	ActionName *string `json:"actionName,omitempty"`

	// Typology Type of the resource (nullable)
	Typology *string `json:"typology,omitempty"`

	// TypologyName Name of the typology (nullable)
	TypologyName *string `json:"typologyName,omitempty"`

	// HttpVerb HTTP verb (nullable)
	HttpVerb *string `json:"httpVerb,omitempty"`

	// Body HTTP request body (nullable)
	Body *string `json:"body,omitempty"`
}

// JobPropertiesRequest contains properties required to configure and schedule a job
type JobPropertiesRequest struct {
	// Enabled Defines whether the job is enabled. Default is true. Uses *bool so
	// that false is serialised (not dropped by omitempty when the caller explicitly
	// disables the job via WithEnabled(false)).
	Enabled *bool `json:"enabled,omitempty"`

	// JobType Type of job
	// For more information, check the documentation.
	// Possible values: OneShot, Recurring
	JobType JobType `json:"scheduleJobType"`

	// ScheduleAt Date and time when the job should run (nullable)
	// Required only for "OneShot" jobs.
	// For more information, check the documentation.
	ScheduleAt *string `json:"scheduleAt,omitempty"`

	// ExecuteUntil End date until which the job can run (nullable)
	// Required only for "Recurring" jobs.
	// For more information, check the documentation.
	ExecuteUntil *string `json:"executeUntil,omitempty"`

	// Cron CRON expression that defines the recurrence of the job (nullable)
	// For more information, check the documentation.
	Cron *string `json:"cron,omitempty"`

	// Steps Steps that will be executed as part of the job (nullable)
	Steps []JobStepRequest `json:"steps,omitempty"`
}

// JobPropertiesResponse contains the response properties of a job
type JobPropertiesResponse struct {
	// Enabled Defines whether the job is enabled
	Enabled bool `json:"enabled,omitempty"`

	// JobType Type of job
	// Possible values: OneShot, Recurring
	JobType JobType `json:"scheduleJobType,omitempty"`

	// ScheduleAt Date and time when the job should run (nullable)
	ScheduleAt *string `json:"scheduleAt,omitempty"`

	// ExecuteUntil End date until which the job can run (nullable)
	ExecuteUntil *string `json:"executeUntil,omitempty"`

	// Cron CRON expression that defines the recurrence of the job (nullable)
	Cron *string `json:"cron,omitempty"`

	// Recurrency Recurrence pattern of the job (nullable)
	// Possible values: Hourly, Daily, Weekly, Monthly, Custom
	Recurrency *RecurrenceType `json:"recurrency,omitempty"`

	// Steps Steps that will be executed as part of the job (nullable)
	Steps []JobStepResponse `json:"steps,omitempty"`

	// NextExecution Date and time of the next scheduled execution (nullable)
	NextExecution *string `json:"nextExecution,omitempty"`

	// DeactiveReason Reason why the job was deactivated (nullable)
	// Possible values: None, Manual, ResourceDeleted
	DeactiveReason *DeactiveReasonDto `json:"deactiveReason,omitempty"`
}

// JobRequest represents a job creation/update request
type JobRequest struct {
	// Metadata of the job
	Metadata RegionalResourceMetadataRequest `json:"metadata"`

	// Properties of the job (nullable object)
	Properties JobPropertiesRequest `json:"properties"`
}

// JobResponse represents a job response
type JobResponse struct {
	// Metadata of the job
	Metadata ResourceMetadataResponse `json:"metadata"`

	// Properties of the job (nullable object)
	Properties JobPropertiesResponse `json:"properties"`

	// Status of the job
	Status ResourceStatusResponse `json:"status,omitempty"`
}

// JobListResponse represents a list of jobs
type JobListResponse struct {
	ListResponse
	Values []JobResponse `json:"values"`
}

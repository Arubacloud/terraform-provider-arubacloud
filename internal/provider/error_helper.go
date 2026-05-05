package provider

// billingPeriodToAPI converts a Terraform-facing billing period value
// (Hour, Month, Year) to the value expected by the Aruba API (hourly, monthly, yearly).
// If the value is already in API form it is returned unchanged.
func billingPeriodToAPI(s string) string {
	switch s {
	case "Hour":
		return "hourly"
	case "Month":
		return "monthly"
	case "Year":
		return "yearly"
	default:
		return s
	}
}

// billingPeriodFromAPI converts an API billing period value (hourly, monthly, yearly)
// to the Terraform-facing canonical form (Hour, Month, Year).
// If the value is already in canonical form it is returned unchanged.
func billingPeriodFromAPI(s string) string {
	switch s {
	case "hourly":
		return "Hour"
	case "monthly":
		return "Month"
	case "yearly":
		return "Year"
	default:
		return s
	}
}

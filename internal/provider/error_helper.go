package provider

// billingPeriodFromAPI converts legacy API billing period values (hourly, monthly, yearly)
// to the canonical Terraform form (Hour, Month, Year).
// Values already in canonical form are returned unchanged.
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

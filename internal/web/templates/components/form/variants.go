package form

import "github.com/cr1cr1/farm-manager/internal/web/templates/components/utils"

func formItemVariants(className string) string {
	// Map to local CSS utility
	baseClasses := "form-row"

	classes := []string{baseClasses}
	if className != "" {
		classes = append(classes, className)
	}

	return utils.TwMerge(classes...)
}

func formLabelVariants(className string, hasError bool) string {
	// Base classes for form label
	baseClasses := ""

	// Add error styling if needed
	if hasError {
		baseClasses = "data-[error=true]:text-destructive"
	}

	classes := []string{}
	if baseClasses != "" {
		classes = append(classes, baseClasses)
	}
	if className != "" {
		classes = append(classes, className)
	}

	if len(classes) == 0 {
		return ""
	}

	return utils.TwMerge(classes...)
}

func formDescriptionVariants(className string) string {
	// Extract EXACT base classes from shadcn/ui New York v4 FormDescription
	baseClasses := "text-muted-foreground text-sm"

	classes := []string{baseClasses}
	if className != "" {
		classes = append(classes, className)
	}

	return utils.TwMerge(classes...)
}

func formMessageVariants(className string) string {
	// Map error text to local style
	baseClasses := "field-error"

	classes := []string{baseClasses}
	if className != "" {
		classes = append(classes, className)
	}

	return utils.TwMerge(classes...)
}

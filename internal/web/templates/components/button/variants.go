package button

import (
	"github.com/cr1cr1/farm-manager/internal/web/templates/components/utils"
)

// buttonVariants generates CSS classes mapped to local app.css styles
func buttonVariants(variant, size, className string) string {
	// Map to our project CSS instead of Tailwind utilities
	baseClasses := "btn"

	// Variant -> additional classes defined in public/css/app.css
	variantClasses := map[string]string{
		"default":     "btn-primary",
		"secondary":   "btn-secondary",
		"outline":     "", // no special outline style in app.css
		"ghost":       "", // not styled; falls back to .btn
		"link":        "", // not styled; falls back to .btn
		"destructive": "", // not styled; falls back to .btn
	}

	// Size not used in current CSS, but keep hook
	_ = size

	vc := variantClasses[variant]
	classes := []string{baseClasses}
	if vc != "" {
		classes = append(classes, vc)
	}
	if className != "" {
		classes = append(classes, className)
	}
	return utils.TwMerge(classes...)
}

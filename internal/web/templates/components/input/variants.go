package input

import "github.com/cr1cr1/farm-manager/internal/web/templates/components/utils"

func inputVariants(className string) string {
	// Use project CSS (public/css/app.css) that styles inputs by type selectors.
	// Return any extra classes passed through.
	if className == "" {
		return ""
	}
	return utils.TwMerge(className)
}

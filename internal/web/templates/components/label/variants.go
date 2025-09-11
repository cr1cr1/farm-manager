package label

import "github.com/cr1cr1/farm-manager/internal/web/templates/components/utils"

func labelVariants(className string) string {
	// Use project CSS (public/css/app.css) which styles plain <label> tags.
	// Pass through any additional classes provided explicitly.
	if className == "" {
		return ""
	}
	return utils.TwMerge(className)
}

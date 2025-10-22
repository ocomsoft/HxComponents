package components

import (
	"net/http"
)

// applyHxHeaders applies HTMX request headers to the instance if it implements
// the corresponding interfaces.
func applyHxHeaders(instance interface{}, req *http.Request) {
	if v, ok := instance.(HxBoosted); ok {
		v.SetHxBoosted(req.Header.Get("HX-Boosted") == "true")
	}
	if v, ok := instance.(HxRequest); ok {
		v.SetHxRequest(req.Header.Get("HX-Request") == "true")
	}
	if v, ok := instance.(HxCurrentURL); ok {
		v.SetHxCurrentURL(req.Header.Get("HX-Current-URL"))
	}
	if v, ok := instance.(HxPrompt); ok {
		v.SetHxPrompt(req.Header.Get("HX-Prompt"))
	}
	if v, ok := instance.(HxTarget); ok {
		v.SetHxTarget(req.Header.Get("HX-Target"))
	}
	if v, ok := instance.(HxTrigger); ok {
		v.SetHxTrigger(req.Header.Get("HX-Trigger"))
	}
	if v, ok := instance.(HxTriggerName); ok {
		v.SetHxTriggerName(req.Header.Get("HX-Trigger-Name"))
	}
	if v, ok := instance.(HttpMethod); ok {
		v.SetHttpMethod(req.Method)
	}
}

// applyHxResponseHeaders applies HTMX response headers from the instance if it implements
// the corresponding interfaces.
func applyHxResponseHeaders(w http.ResponseWriter, instance interface{}) {
	if v, ok := instance.(HxLocationResponse); ok {
		if location := v.GetHxLocation(); location != "" {
			w.Header().Set("HX-Location", location)
		}
	}
	if v, ok := instance.(HxPushUrlResponse); ok {
		if pushUrl := v.GetHxPushUrl(); pushUrl != "" {
			w.Header().Set("HX-Push-Url", pushUrl)
		}
	}
	if v, ok := instance.(HxRedirectResponse); ok {
		if redirect := v.GetHxRedirect(); redirect != "" {
			w.Header().Set("HX-Redirect", redirect)
		}
	}
	if v, ok := instance.(HxRefreshResponse); ok {
		if v.GetHxRefresh() {
			w.Header().Set("HX-Refresh", "true")
		}
	}
	if v, ok := instance.(HxReplaceUrlResponse); ok {
		if replaceUrl := v.GetHxReplaceUrl(); replaceUrl != "" {
			w.Header().Set("HX-Replace-Url", replaceUrl)
		}
	}
	if v, ok := instance.(HxReswapResponse); ok {
		if reswap := v.GetHxReswap(); reswap != "" {
			w.Header().Set("HX-Reswap", reswap)
		}
	}
	if v, ok := instance.(HxRetargetResponse); ok {
		if retarget := v.GetHxRetarget(); retarget != "" {
			w.Header().Set("HX-Retarget", retarget)
		}
	}
	if v, ok := instance.(HxReselectResponse); ok {
		if reselect := v.GetHxReselect(); reselect != "" {
			w.Header().Set("HX-Reselect", reselect)
		}
	}
	if v, ok := instance.(HxTriggerResponse); ok {
		if trigger := v.GetHxTrigger(); trigger != "" {
			w.Header().Set("HX-Trigger", trigger)
		}
	}
	if v, ok := instance.(HxTriggerAfterSettleResponse); ok {
		if trigger := v.GetHxTriggerAfterSettle(); trigger != "" {
			w.Header().Set("HX-Trigger-After-Settle", trigger)
		}
	}
	if v, ok := instance.(HxTriggerAfterSwapResponse); ok {
		if trigger := v.GetHxTriggerAfterSwap(); trigger != "" {
			w.Header().Set("HX-Trigger-After-Swap", trigger)
		}
	}
}

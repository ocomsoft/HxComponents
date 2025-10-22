package components

// HxLocationResponse is implemented by structs that want to set the HX-Location response header.
// This allows you to do a client-side redirect that doesn't do a full page reload.
type HxLocationResponse interface {
	GetHxLocation() string
}

// HxPushUrlResponse is implemented by structs that want to set the HX-Push-Url response header.
// This pushes a new URL into the browser's history stack.
type HxPushUrlResponse interface {
	GetHxPushUrl() string
}

// HxRedirectResponse is implemented by structs that want to set the HX-Redirect response header.
// This does a client-side redirect to a new location.
type HxRedirectResponse interface {
	GetHxRedirect() string
}

// HxRefreshResponse is implemented by structs that want to set the HX-Refresh response header.
// If true, the client will do a full page refresh.
type HxRefreshResponse interface {
	GetHxRefresh() bool
}

// HxReplaceUrlResponse is implemented by structs that want to set the HX-Replace-Url response header.
// This replaces the current URL in the browser's history stack.
type HxReplaceUrlResponse interface {
	GetHxReplaceUrl() string
}

// HxReswapResponse is implemented by structs that want to set the HX-Reswap response header.
// This allows you to specify how the response will be swapped (innerHTML, outerHTML, etc.).
type HxReswapResponse interface {
	GetHxReswap() string
}

// HxRetargetResponse is implemented by structs that want to set the HX-Retarget response header.
// This allows you to change the target element for the swap operation.
type HxRetargetResponse interface {
	GetHxRetarget() string
}

// HxReselectResponse is implemented by structs that want to set the HX-Reselect response header.
// This allows you to choose a subset of the response to swap in using a CSS selector.
type HxReselectResponse interface {
	GetHxReselect() string
}

// HxTriggerResponse is implemented by structs that want to set the HX-Trigger response header.
// This allows you to trigger client-side events after the swap.
type HxTriggerResponse interface {
	GetHxTrigger() string
}

// HxTriggerAfterSettleResponse is implemented by structs that want to set the HX-Trigger-After-Settle response header.
// This allows you to trigger client-side events after the settle phase.
type HxTriggerAfterSettleResponse interface {
	GetHxTriggerAfterSettle() string
}

// HxTriggerAfterSwapResponse is implemented by structs that want to set the HX-Trigger-After-Swap response header.
// This allows you to trigger client-side events after the swap phase.
type HxTriggerAfterSwapResponse interface {
	GetHxTriggerAfterSwap() string
}

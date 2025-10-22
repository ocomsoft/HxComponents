package components

// HxBoosted is implemented by structs that want to receive the HX-Boosted header value.
// This header indicates whether the request was made via an element with hx-boost="true".
type HxBoosted interface {
	SetHxBoosted(bool)
}

// HxRequest is implemented by structs that want to receive the HX-Request header value.
// This header is always "true" for requests made by HTMX.
type HxRequest interface {
	SetHxRequest(bool)
}

// HxCurrentURL is implemented by structs that want to receive the HX-Current-URL header value.
// This is the current URL of the browser when the request was made.
type HxCurrentURL interface {
	SetHxCurrentURL(string)
}

// HxPrompt is implemented by structs that want to receive the HX-Prompt header value.
// This is the user's response to an hx-prompt, if one was present.
type HxPrompt interface {
	SetHxPrompt(string)
}

// HxTarget is implemented by structs that want to receive the HX-Target header value.
// This is the id of the target element, if it exists.
type HxTarget interface {
	SetHxTarget(string)
}

// HxTrigger is implemented by structs that want to receive the HX-Trigger header value.
// This is the id of the element that triggered the request.
type HxTrigger interface {
	SetHxTrigger(string)
}

// HxTriggerName is implemented by structs that want to receive the HX-Trigger-Name header value.
// This is the name of the element that triggered the request, if it exists.
type HxTriggerName interface {
	SetHxTriggerName(string)
}

// HttpMethod is implemented by structs that want to receive the HTTP method (GET or POST).
// This allows components to vary behavior based on whether they were loaded via GET or submitted via POST.
type HttpMethod interface {
	SetHttpMethod(string)
}

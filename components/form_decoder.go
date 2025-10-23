package components

import "github.com/go-playground/form/v4"

// FormDecoder is an optional interface that components can implement to provide
// a custom form decoder. This allows components to configure form decoding behavior,
// such as using different tag names, custom parsers, or validation rules.
//
// If a component does not implement this interface, the default decoder is used.
//
// Example:
//
//	type MyComponent struct {
//	    Email string `json:"email"` // Using json tags instead of form tags
//	}
//
//	func (c *MyComponent) GetFormDecoder() *form.Decoder {
//	    decoder := form.NewDecoder()
//	    decoder.SetTagName("json") // Use json tags for form decoding
//	    return decoder
//	}
//
// Example with custom parser:
//
//	type MyComponent struct {
//	    Date time.Time `form:"date"`
//	}
//
//	func (c *MyComponent) GetFormDecoder() *form.Decoder {
//	    decoder := form.NewDecoder()
//	    decoder.RegisterCustomTypeFunc(func(vals []string) (interface{}, error) {
//	        return time.Parse("2006-01-02", vals[0])
//	    }, time.Time{})
//	    return decoder
//	}
type FormDecoder interface {
	GetFormDecoder() *form.Decoder
}

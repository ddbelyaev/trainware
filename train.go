package trainware

import "net/http"

type middleware func(http.Handler) http.Handler

type train []middleware

func New() train {
	return train(make([]middleware, 0))
}

// Middleware will be Added to the back of the slice (appended).
// It means that first middleware is going to be the most inner,
// and last middleware the most outer.
// In other words, last middleware will be executed first, and
// first middleware will be executed last.
func (t train) Add(mw middleware) train {
	// Function type can be nil, and calling a nil will produce panic
	// Therefore we should account for nil case
	if mw == nil {
		return t
	}

	return append(t, mw)
}

// AddMany can be used instead of Add chains. Mostly added as a sugar function.
func (t train) AddMany(mws ...middleware) train {
	for _, mw := range mws {
		t = t.Add(mw)
	}

	return t
}

// When middleware train is appended, this method applies the train to router.
// If user chooses not to pass any router, it would fallback to default Go router.
func (t train) Apply(handler http.Handler) http.Handler {
	if handler == nil {
		handler = http.DefaultServeMux
	}

	for _, mw := range t {
		handler = mw(handler)
	}

	return handler
}

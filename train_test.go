package trainware

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

// This middleware creates or appends to a slice of ints, and passes it down to
// request context. Later this value will be read and tested against.
func helperContextMiddleware(n int) middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			slice, ok := r.Context().Value("testValue").([]int)
			if ok {
				slice = append(slice, n)
			} else {
				slice = []int{n}
			}

			ctx := context.WithValue(r.Context(), "testValue", slice)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// A simple handler that reads context value from request and writes it to response
var helperContextHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	out, err := json.Marshal(r.Context().Value("testValue").([]int))
	if err != nil {
		log.Fatal("Can't marshal value out of request context, value: ", r.Context().Value("testValue"))
	}

	w.WriteHeader(200)
	w.Write(out)
})

func helperTestServerResponder(h http.Handler) []byte {
	testServer := httptest.NewServer(h)
	defer testServer.Close()

	response, err := http.Get(testServer.URL)
	if err != nil {
		log.Fatal(err)
	}

	body, err := io.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	return body
}

func TestHelperContextMiddleware(t *testing.T) {
	testValue := 100
	contextMiddleware := helperContextMiddleware(testValue)
	testHandler := contextMiddleware(helperContextHandler)

	body := helperTestServerResponder(testHandler)

	var result []int
	if err := json.Unmarshal(body, &result); err != nil {
		t.Error("Can't unmarshal response body", err)
	}

	if result == nil || len(result) == 0 {
		t.Error("Body doesn't contain any value")
	}

	if result[0] != testValue {
		t.Errorf("Value in result slice doesn't equal testValue: %v != %v", result[0], testValue)
	}
}

func TestContextMiddlewareTrain(t *testing.T) {
	train := New()
	testValues := []int{1, 2, 3, 4, 5}

	for _, n := range testValues {
		train = train.Add(helperContextMiddleware(n))
	}

	testTrainHandler := train.Apply(helperContextHandler)

	body := helperTestServerResponder(testTrainHandler)

	var result []int
	if err := json.Unmarshal(body, &result); err != nil {
		t.Error("Can't unmarshal response body", err)
	}

	for i := 0; i < len(testValues); i++ {
		if result[i] != testValues[len(testValues)-1-i] {
			t.Error("Values in result don't match expected")
		}
	}
}

// We are testing that applier2 has more middleware than applier1
func TestContextMiddlewareMultipleTrains(t *testing.T) {
	train := New()

	applier1 := train.AddMany(helperContextMiddleware(1), helperContextMiddleware(2))
	applier2 := applier1.AddMany(helperContextMiddleware(3), helperContextMiddleware(4))

	testTrainHandler1 := applier1.Apply(helperContextHandler)
	testTrainHandler2 := applier2.Apply(helperContextHandler)

	body1 := helperTestServerResponder(testTrainHandler1)

	var result1 []int
	if err := json.Unmarshal(body1, &result1); err != nil {
		t.Error("Can't unmarshal response body", err)
	}

	body2 := helperTestServerResponder(testTrainHandler2)

	var result2 []int
	if err := json.Unmarshal(body2, &result2); err != nil {
		t.Error("Can't unmarshal response body", err)
	}

	if len(result1) == len(result2) {
		t.Error("Trains should have different amount of middlewares")
	}
}

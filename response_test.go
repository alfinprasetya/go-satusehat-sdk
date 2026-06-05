package satusehat

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestReadAPIResponse_SuccessAndConsentError(t *testing.T) {
	t.Parallel()

	t.Run("success bundle", func(t *testing.T) {
		t.Parallel()

		body := `{"resourceType":"Bundle","type":"searchset","total":0}`
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/fhir+json"}},
			Body:       io.NopCloser(strings.NewReader(body)),
		}

		got, err := readAPIResponse(resp)
		if err != nil {
			t.Fatalf("readAPIResponse: %v", err)
		}
		if string(got) != body {
			t.Errorf("body: got %q, want %q", string(got), body)
		}
	})

	t.Run("consent suppressed on 200", func(t *testing.T) {
		t.Parallel()

		body := `{"resourceType":"OperationOutcome","issue":[{"severity":"information","code":"suppressed","details":{"text":"consent"}}]}`
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{},
			Body:       io.NopCloser(strings.NewReader(body)),
		}

		_, err := readAPIResponse(resp)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrConsentSuppressed) {
			t.Errorf("errors.Is: got %v, want ErrConsentSuppressed", err)
		}
	})
}

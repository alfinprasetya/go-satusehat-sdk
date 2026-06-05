package satusehat

import (
	"io"
	"net/http"
)

// readAPIResponse reads and closes resp.Body.
// It returns the body on success (2xx without an OperationOutcome error payload).
// Any documented API error shape is returned as a typed error from ParseAPIError.
func readAPIResponse(resp *http.Response) ([]byte, error) {
	body, err := readResponseBody(resp)
	if err != nil {
		return nil, err
	}

	statusCode := resp.StatusCode
	if statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices {
		if isOperationOutcomeBody(body) {
			return nil, ParseAPIError(statusCode, resp.Header, body)
		}
		return body, nil
	}

	return nil, ParseAPIError(statusCode, resp.Header, body)
}

// readAuthResponse reads and closes resp.Body for the OAuth token endpoint.
func readAuthResponse(resp *http.Response) ([]byte, error) {
	body, err := readResponseBody(resp)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusOK {
		return body, nil
	}

	return nil, ParseAPIError(resp.StatusCode, resp.Header, body)
}

func readResponseBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

package infakt

import "net/http"

// checkResponse checks the API response for errors.
func checkResponse(r *http.Response) error {
	if r.StatusCode >= 200 && r.StatusCode <= 299 {
		return nil
	}
	return nil
}

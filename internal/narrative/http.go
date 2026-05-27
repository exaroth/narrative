package narrative

import "net/http"

// Defines custom transport for default HTTP client.
type roundTripperStripUserAgent struct{}

func (_ roundTripperStripUserAgent) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", "Narrative")
	return http.DefaultTransport.RoundTrip(req)
}

// Retrieve default http client used across application.
func GetHttpClient() *http.Client {
	return &http.Client{
		Transport: roundTripperStripUserAgent{},
	}

}

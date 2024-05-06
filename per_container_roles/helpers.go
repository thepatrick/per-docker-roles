package per_container_roles

import (
	"io"
	"net/http"
)

func DenyXForwardedFor(w http.ResponseWriter, r *http.Request) bool {
	// Check for the presence of the X-Forwarded-For header
	xForwardedForHeader := r.Header.Get(X_FORWARDED_FOR_HEADER) // canonicalized headers are used (casing doesn't matter)
	if xForwardedForHeader != "" {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "unable to process requests with X-Forwarded-For header")
		return true
	}
	return false
}

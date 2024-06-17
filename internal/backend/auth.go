package backend

import "net/http"

func (p *ReverseProxy) isAuthorized(r *http.Request) bool {
	authToken := r.Header.Get("X-Auth-Token")
	for _, validToken := range p.config.AuthToken {
		if authToken == validToken {
			return true
		}
	}
	return false
}

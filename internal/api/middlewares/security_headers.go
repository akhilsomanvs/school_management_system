package middlewares

import "net/http"

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//Disables DNS prefetching
		w.Header().Set("X-DNS-Prefetch-Control", "off")
		//Prevents the webpage to be shown in iframes in other sites
		w.Header().Set("X-Frame-Options", "DENY")
		//Enables Cross site filter, prevents XSS attack
		w.Header().Set("X-XSS-Protection", "1;mode=block")
		//Prevents browsers from MIME sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")
		//TElls browsers to interact with your site only over HTTPS
		w.Header().Set("Strict-Transport-Security", "max-age=63072000;includeSubDomains;preload")
		//Controls which resources can be loaded on the page
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		//How much referrer information should be included with the request made from your site
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

/*
//basic MIDDLEWARE SKELETON

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//next.ServeHTTP(w,r)
	})
}
*/

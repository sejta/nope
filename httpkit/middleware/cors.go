package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	corsHeaderOrigin           = "Origin"
	corsHeaderAllowOrigin      = "Access-Control-Allow-Origin"
	corsHeaderAllowMethods     = "Access-Control-Allow-Methods"
	corsHeaderAllowHeaders     = "Access-Control-Allow-Headers"
	corsHeaderAllowCredentials = "Access-Control-Allow-Credentials"
	corsHeaderExposeHeaders    = "Access-Control-Expose-Headers"
	corsHeaderMaxAge           = "Access-Control-Max-Age"
	corsHeaderRequestMethod    = "Access-Control-Request-Method"
	corsHeaderRequestHeaders   = "Access-Control-Request-Headers"
	corsHeaderVary             = "Vary"
)

var (
	corsDefaultMethods = []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}
	corsDefaultHeaders = []string{"Content-Type", "Authorization"}
)

// CORSOptions задаёт политику CORS для middleware.
type CORSOptions struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           time.Duration
}

// CORS добавляет обработку CORS с безопасными дефолтами.
func CORS(opts CORSOptions) func(http.Handler) http.Handler {
	validateCORSOptions(opts)

	allowedMethods := opts.AllowedMethods
	if len(allowedMethods) == 0 {
		allowedMethods = corsDefaultMethods
	}
	allowedHeaders := opts.AllowedHeaders
	if len(allowedHeaders) == 0 {
		allowedHeaders = corsDefaultHeaders
	}

	allowedMethodsCSV := joinCSV(allowedMethods)
	allowedHeadersCSV := joinCSV(allowedHeaders)
	exposedHeadersCSV := joinCSV(opts.ExposedHeaders)
	allowHeadersWildcard := hasWildcard(allowedHeaders)
	allowOriginsWildcard := hasWildcard(opts.AllowedOrigins)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(opts.AllowedOrigins) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			origin := r.Header.Get(corsHeaderOrigin)
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			isPreflight := r.Method == http.MethodOptions && r.Header.Get(corsHeaderRequestMethod) != ""
			if !originAllowed(origin, opts.AllowedOrigins) {
				if isPreflight {
					w.WriteHeader(http.StatusNoContent)
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			allowOrigin := origin
			varyOrigin := true
			if allowOriginsWildcard {
				allowOrigin = "*"
				varyOrigin = false
			}

			w.Header().Set(corsHeaderAllowOrigin, allowOrigin)
			if varyOrigin {
				addVary(w.Header(), corsHeaderOrigin)
			}
			if opts.AllowCredentials {
				w.Header().Set(corsHeaderAllowCredentials, "true")
			}

			if isPreflight {
				w.Header().Set(corsHeaderAllowMethods, allowedMethodsCSV)

				requestHeaders := strings.TrimSpace(r.Header.Get(corsHeaderRequestHeaders))
				if requestHeaders != "" && allowHeadersWildcard {
					w.Header().Set(corsHeaderAllowHeaders, requestHeaders)
				} else if allowedHeadersCSV != "" {
					w.Header().Set(corsHeaderAllowHeaders, allowedHeadersCSV)
				}

				if opts.MaxAge > 0 {
					w.Header().Set(corsHeaderMaxAge, strconv.FormatInt(int64(opts.MaxAge/time.Second), 10))
				}
				addVary(w.Header(), corsHeaderRequestMethod)
				addVary(w.Header(), corsHeaderRequestHeaders)
				w.WriteHeader(http.StatusNoContent)
				return
			}

			if exposedHeadersCSV != "" {
				w.Header().Set(corsHeaderExposeHeaders, exposedHeadersCSV)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func validateCORSOptions(opts CORSOptions) {
	if !opts.AllowCredentials {
		return
	}
	if hasWildcard(opts.AllowedOrigins) {
		panic("cors: AllowCredentials запрещает wildcard origin")
	}
}

func originAllowed(origin string, allowed []string) bool {
	for _, a := range allowed {
		if a == "*" {
			return true
		}
		if a == origin {
			return true
		}
	}
	return false
}

func hasWildcard(items []string) bool {
	for _, item := range items {
		if item == "*" {
			return true
		}
	}
	return false
}

func joinCSV(items []string) string {
	if len(items) == 0 {
		return ""
	}
	return strings.Join(items, ", ")
}

func addVary(h http.Header, value string) {
	if value == "" {
		return
	}
	if existing := h.Get(corsHeaderVary); existing != "" {
		for _, part := range strings.Split(existing, ",") {
			if strings.EqualFold(strings.TrimSpace(part), value) {
				return
			}
		}
		h.Set(corsHeaderVary, existing+", "+value)
		return
	}
	h.Set(corsHeaderVary, value)
}

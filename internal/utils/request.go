// Pacote utils fornece funções utilitárias usadas em toda a aplicação.
package utils

import (
	"net/http"
	"regexp"
	"strings"
)

// GetRealIP extrai o IP real do cliente considerando proxies e load balancers
// Verifica headers comuns: X-Forwarded-For, X-Real-IP, CF-Connecting-IP (Cloudflare)
func GetRealIP(r *http.Request) string {
	// 1. Tenta X-Forwarded-For (pode conter múltiplos IPs separados por vírgula)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Pega o primeiro IP da lista (o IP original do cliente)
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 2. Tenta X-Real-IP (normalmente setado por nginx)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// 3. Tenta CF-Connecting-IP (Cloudflare)
	if cfIP := r.Header.Get("CF-Connecting-IP"); cfIP != "" {
		return strings.TrimSpace(cfIP)
	}

	// 4. Fallback para RemoteAddr (pode ser do proxy se houver)
	ip := r.RemoteAddr
	// Remove a porta se houver (formato: "192.168.1.1:54321")
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}

	return ip
}

// GetUserAgent extrai o User-Agent do request
func GetUserAgent(r *http.Request) string {
	return r.Header.Get("User-Agent")
}

// GetDeviceInfo extrai informações básicas do dispositivo do User-Agent
func GetDeviceInfo(r *http.Request) string {
	ua := r.Header.Get("User-Agent")
	if ua == "" {
		return "Unknown Device"
	}

	// Detecção simples de plataforma
	switch {
	case strings.Contains(ua, "Mobile") || strings.Contains(ua, "Android"):
		if strings.Contains(ua, "Android") {
			return "Android Device"
		}
		if strings.Contains(ua, "iPhone") || strings.Contains(ua, "iPad") {
			return "iOS Device"
		}
		return "Mobile Device"
	case strings.Contains(ua, "Windows"):
		return "Windows Desktop"
	case strings.Contains(ua, "Macintosh") || strings.Contains(ua, "Mac OS"):
		return "macOS Desktop"
	case strings.Contains(ua, "Linux"):
		return "Linux Desktop"
	default:
		return "Unknown Device"
	}
}

// GetReferer extrai o referer (página de origem) do request
func GetReferer(r *http.Request) string {
	return r.Header.Get("Referer")
}

// GetRequestMethod extrai o método HTTP do request
func GetRequestMethod(r *http.Request) string {
	return r.Method
}

// GetRequestPath extrai o caminho da URL do request
func GetRequestPath(r *http.Request) string {
	return r.URL.Path
}

// GetActionFromRequest creates an action string based on method and path.
func GetActionFromRequest(r *http.Request) string {
	method := r.Method
	path := r.URL.Path

	path = strings.TrimPrefix(path, "/api/v1/")
	path = strings.TrimPrefix(path, "/api/")

	segments := strings.Split(path, "/")
	var cleanSegments []string

	uuidPattern := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

	for _, seg := range segments {
		if seg == "" {
			continue
		}
		if uuidPattern.MatchString(seg) {
			continue
		}
		cleanSegments = append(cleanSegments, seg)
	}

	action := strings.ToUpper(strings.Join(cleanSegments, "_"))

	switch method {
	case http.MethodPost:
		return "CREATE_" + action
	case http.MethodPut, http.MethodPatch:
		return "UPDATE_" + action
	case http.MethodDelete:
		return "DELETE_" + action
	case http.MethodGet:
		return "VIEW_" + action
	default:
		return method + "_" + action
	}
}

// ExtractEntityFromPath extracts an entity name from the API path.
func ExtractEntityFromPath(path string) string {
	// Remove prefixos comuns
	path = strings.TrimPrefix(path, "/api/v1/")
	path = strings.TrimPrefix(path, "/api/")

	// Pega o primeiro segmento do path
	segments := strings.Split(path, "/")
	if len(segments) > 0 && segments[0] != "" {
		// Remove plural (simplificado)
		entity := segments[0]
		if strings.HasSuffix(entity, "ies") {
			return strings.TrimSuffix(entity, "ies") + "y"
		}
		if strings.HasSuffix(entity, "s") {
			return strings.TrimSuffix(entity, "s")
		}
		return entity
	}

	return "unknown"
}

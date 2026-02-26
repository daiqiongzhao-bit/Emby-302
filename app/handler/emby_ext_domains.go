package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"film-fusion/app/config"

	"github.com/gin-gonic/gin"
)

type EmbyExtDomainsHandler struct {
	cfg *config.Config
}

type extDomainInfo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type extDomainResponse struct {
	Data []extDomainInfo `json:"data"`
	OK   bool            `json:"ok"`
}

func NewEmbyExtDomainsHandler(cfg *config.Config) *EmbyExtDomainsHandler {
	return &EmbyExtDomainsHandler{cfg: cfg}
}

func (h *EmbyExtDomainsHandler) GetServerDomains(c *gin.Context) {
	if !h.cfg.Emby.ExtDomains.Enabled {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Ext domains is disabled",
			"ok":    false,
		})
		return
	}

	token := h.extractToken(c)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Token not found",
			"ok":    false,
		})
		return
	}

	if !h.validateToken(token) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid token",
			"ok":    false,
		})
		return
	}

	domains := make([]extDomainInfo, len(h.cfg.Emby.ExtDomains.Domains))
	for i, d := range h.cfg.Emby.ExtDomains.Domains {
		domains[i] = extDomainInfo{Name: d.Name, URL: d.URL}
	}

	c.JSON(http.StatusOK, extDomainResponse{Data: domains, OK: true})
}

func (h *EmbyExtDomainsHandler) validateToken(token string) bool {
	serverURL := strings.TrimRight(h.cfg.Emby.URL, "/")
	if serverURL == "" {
		return false
	}

	timeoutSec := h.cfg.Emby.ExtDomains.ValidateTimeoutSeconds
	if timeoutSec <= 0 {
		timeoutSec = 3
	}

	verifyURL := fmt.Sprintf("%s/emby/System/Info?X-Emby-Token=%s", serverURL, url.QueryEscape(token))
	client := &http.Client{Timeout: time.Duration(timeoutSec) * time.Second}

	req, err := http.NewRequest(http.MethodGet, verifyURL, nil)
	if err != nil {
		return false
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "*/*")

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func (h *EmbyExtDomainsHandler) extractToken(c *gin.Context) string {
	if t := c.Query("X-Emby-Token"); t != "" {
		return t
	}
	if t := c.GetHeader("X-Emby-Token"); t != "" {
		return t
	}
	if t := c.Query("api_key"); t != "" {
		return t
	}
	if t := c.GetHeader("X-Emby-Authorization"); t != "" {
		content := strings.TrimPrefix(t, "MediaBrowser ")
		return getTokenByStringSplit(content)
	}
	if t, err := c.Cookie("Authorization"); err == nil && t != "" {
		return t
	}
	if t := c.GetHeader("Authorization"); t != "" {
		return t
	}
	if t := c.Query("token"); t != "" {
		return t
	}
	if t, err := c.Cookie("token"); err == nil && t != "" {
		return t
	}
	if t := c.GetHeader("token"); t != "" {
		return t
	}
	return ""
}

func getTokenByStringSplit(mediaBrowserHeader string) string {
	parts := strings.Split(mediaBrowserHeader, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "Token=") {
			token := strings.TrimPrefix(part, "Token=")
			token = strings.Trim(token, `"`)
			return token
		}
	}
	return ""
}

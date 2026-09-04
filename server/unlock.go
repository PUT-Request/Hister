package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/asciimoo/hister/config"
	"github.com/asciimoo/hister/server/indexer"
	"github.com/asciimoo/hister/server/indexer/encryption"
	"github.com/rs/zerolog/log"
)

// serverEncryptionState holds the in-memory encryption state. It is nil when
// encryption is not configured.
type serverEncryptionState struct {
	mu              sync.RWMutex
	keyPair         *encryption.KeyPair
	locked          bool
	lastActivity    time.Time
	autoLockTimeout time.Duration
	sessionToken    string
}

var encState *serverEncryptionState

// InitEncryption sets up encryption at server startup. If encryption is enabled
// in config and a password is provided, it auto-unlocks. If enabled but no
// password, the server starts locked and the user must unlock via /unlock.
func InitEncryption(cfg *config.Config, idx *indexer.Indexer) {
	if !cfg.Encryption.Enable {
		return
	}

	// Generate random session token for cookie auth
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		log.Error().Err(err).Msg("failed to generate session token")
		return
	}

	encState = &serverEncryptionState{
		locked:       true,
		lastActivity: time.Now(),
		sessionToken: hex.EncodeToString(tokenBytes),
	}
	dataDir := cfg.App.Directory
	if dataDir == "" {
		dataDir = "."
	}

	// Parse auto-lock timeout
	if cfg.Encryption.AutoLockTimeout != "" {
		d, err := time.ParseDuration(cfg.Encryption.AutoLockTimeout)
		if err != nil {
			log.Warn().Err(err).Str("value", cfg.Encryption.AutoLockTimeout).Msg("invalid auto_lock_timeout; auto-lock disabled")
		} else if d > 0 {
			encState.autoLockTimeout = d
			go startAutoLockRoutine(idx)
			log.Info().Dur("timeout", d).Msg("auto-lock enabled")
		}
	}

	if !encryption.KeysExist(dataDir) {
		// First run with encryption enabled — need password to generate keys
		if cfg.Encryption.Password != "" {
			kp, err := encryption.GenerateKeyPair(dataDir, cfg.Encryption.Password)
			if err != nil {
				log.Error().Err(err).Msg("failed to generate encryption keys")
				return
			}
			encState.keyPair = kp
			encState.locked = false
			idx.SetKeyPair(kp)
			log.Info().Msg("encryption enabled: keys generated and server unlocked")
			return
		}
		log.Info().Msg("encryption enabled but no password in config; server locked until /unlock")
		return
	}

	// Keys exist — try auto-unlock if password is in config
	if cfg.Encryption.Password != "" {
		kp, err := encryption.UnlockKeyPair(dataDir, cfg.Encryption.Password)
		if err != nil {
			log.Warn().Err(err).Msg("auto-unlock failed (wrong password in config?)")
			return
		}
		encState.keyPair = kp
		encState.locked = false
		idx.SetKeyPair(kp)
		log.Info().Msg("encryption enabled: auto-unlocked with config password")
		return
	}

	// Keys exist but no password — server starts locked
	log.Info().Msg("encryption enabled: server locked; use /unlock to decrypt")
}

// isEncryptionEnabled returns true if encryption is configured.
func isEncryptionEnabled() bool {
	return encState != nil
}

// isEncryptionLocked returns true if encryption is configured and the server
// is locked (private key not in memory).
func isEncryptionLocked() bool {
	return encState != nil && encState.locked
}

// --- Unlock page HTML (matches Hister brutalist style) ---

func serveUnlockPage(c *webContext) {
	c.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Response.Header().Set("Cache-Control", "no-store")
	fmt.Fprint(c.Response, unlockPageHTML)
}

// --- Auto-lock routine ---

func startAutoLockRoutine(idx *indexer.Indexer) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if encState == nil {
			return
		}
		encState.mu.RLock()
		if encState.locked || encState.autoLockTimeout == 0 {
			encState.mu.RUnlock()
			continue
		}
		elapsed := time.Since(encState.lastActivity)
		encState.mu.RUnlock()

		if elapsed >= encState.autoLockTimeout {
			encState.mu.Lock()
			encState.keyPair = nil
			encState.locked = true
			// Generate new session token so old cookies are invalid
			tokenBytes := make([]byte, 32)
			if _, err := rand.Read(tokenBytes); err == nil {
				encState.sessionToken = hex.EncodeToString(tokenBytes)
			}
			encState.mu.Unlock()
			idx.SetKeyPair(nil)
			log.Info().Dur("inactive", elapsed).Msg("auto-locked due to inactivity")
		}
	}
}

func updateActivity() {
	if encState != nil {
		encState.mu.Lock()
		encState.lastActivity = time.Now()
		encState.mu.Unlock()
	}
}

// --- Cookie helpers ---

const sessionCookieName = "hister_session"

func setSessionCookie(w http.ResponseWriter, r *http.Request, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   30 * 24 * 60 * 60, // 30 days
	})
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}

func hasValidSession(r *http.Request) bool {
	if encState == nil {
		return false
	}
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return false
	}
	encState.mu.RLock()
	defer encState.mu.RUnlock()
	return cookie.Value == encState.sessionToken
}

// --- API endpoints ---

func serveUnlockAPI(c *webContext) {
	if !isEncryptionEnabled() {
		c.JSONStatus(http.StatusBadRequest, map[string]string{"error": "encryption not enabled"})
		return
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSONStatus(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	if req.Password == "" {
		c.JSONStatus(http.StatusBadRequest, map[string]string{"error": "password required"})
		return
	}

	dataDir := c.Config.App.Directory
	if dataDir == "" {
		dataDir = "."
	}

	// If keys don't exist yet, generate them with this password
	if !encryption.KeysExist(dataDir) {
		kp, err := encryption.GenerateKeyPair(dataDir, req.Password)
		if err != nil {
			log.Error().Err(err).Msg("failed to generate encryption keys")
			c.JSONStatus(http.StatusInternalServerError, map[string]string{"error": "failed to generate keys"})
			return
		}
		encState.mu.Lock()
		encState.keyPair = kp
		encState.locked = false
		encState.mu.Unlock()
		c.Indexer.SetKeyPair(kp)
		setSessionCookie(c.Response, c.Request, encState.sessionToken)
		log.Info().Msg("encryption keys generated and server unlocked via /api/unlock")
		c.JSON(map[string]string{"status": "unlocked"})
		return
	}

	// Keys exist — always verify password, even if already unlocked
	kp, err := encryption.UnlockKeyPair(dataDir, req.Password)
	if err != nil {
		c.JSONStatus(http.StatusUnauthorized, map[string]string{"error": "wrong password"})
		return
	}

	encState.mu.Lock()
	encState.keyPair = kp
	encState.locked = false
	encState.mu.Unlock()
	c.Indexer.SetKeyPair(kp)
	setSessionCookie(c.Response, c.Request, encState.sessionToken)

	log.Info().Msg("server unlocked via /api/unlock")
	c.JSON(map[string]string{"status": "unlocked"})
}

func serveLockAPI(c *webContext) {
	if !isEncryptionEnabled() {
		c.JSONStatus(http.StatusBadRequest, map[string]string{"error": "encryption not enabled"})
		return
	}

	encState.mu.Lock()
	encState.keyPair = nil
	encState.locked = true
	encState.mu.Unlock()
	c.Indexer.SetKeyPair(nil)
	clearSessionCookie(c.Response)

	log.Info().Msg("server locked via /api/lock")
	c.JSON(map[string]string{"status": "locked"})
}

func serveEncryptionStatus(c *webContext) {
	if !isEncryptionEnabled() {
		c.JSON(map[string]any{"enabled": false})
		return
	}
	encState.mu.RLock()
	defer encState.mu.RUnlock()
	c.JSON(map[string]any{
		"enabled": true,
		"locked":  encState.locked,
	})
}

// --- Middleware ---

// Paths that work without unlock (write endpoints, static assets, etc.)
var unlockedExemptPaths = map[string]bool{
	"/health":       true,
	"/unlock":       true,
	"/api/unlock":   true,
	"/api/lock":     true,
	"/api/status":   true,
	"/api/config":   true,
	"/api/add":      true,
	"/api/add/pdf":  true,
	"/api/document": true,
	"/api/rules":    true,
}

func isExemptPath(path string) bool {
	if unlockedExemptPaths[path] {
		return true
	}
	if strings.HasPrefix(path, "/static/") || path == "/favicon.ico" || path == "/opensearch.xml" {
		return true
	}
	return false
}

// requireUnlocked enforces cookie-based auth for reading encrypted data.
// Write endpoints (add/index) are exempt — they encrypt with the public key.
func requireUnlocked(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isEncryptionEnabled() {
			next.ServeHTTP(w, r)
			return
		}

		path := r.URL.Path

		// Exempt paths that work without auth
		if isExemptPath(path) {
			next.ServeHTTP(w, r)
			return
		}

		// If server is unlocked, check for valid session cookie
		if !isEncryptionLocked() {
			if hasValidSession(r) {
				updateActivity()
				next.ServeHTTP(w, r)
				return
			}
			// Unlocked but no cookie — new browser session, redirect to /unlock
		}

		// Server locked or no valid session — block access
		if strings.HasPrefix(path, "/api/") || r.Header.Get("Accept") == "application/json" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, `{"error":"server locked","redirect":"/unlock"}`)
			return
		}

		// For browser requests, redirect to unlock page
		http.Redirect(w, r, "/unlock", http.StatusTemporaryRedirect)
	})
}

// unlockPageHTML is a self-contained HTML page matching Hister's brutalist UI.
const unlockPageHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width,initial-scale=1"/>
<title>Unlock - Hister</title>
<style>
*,*::before,*::after{box-sizing:border-box;margin:0;padding:0}
:root{
  --bg:#fafafa;--card:#fff;--border:#222;--text:#111;--muted:#666;
  --indigo:#4338ca;--shadow:#222;--font-space:'Inter',system-ui,sans-serif;
  --font-outfit:'Outfit',system-ui,sans-serif;
}
@media(prefers-color-scheme:dark){:root{
  --bg:#111;--card:#1a1a1a;--border:#555;--text:#eee;--muted:#aaa;
  --indigo:#818cf8;--shadow:#000;
}}
html,body{height:100%;font-family:var(--font-space);background:var(--bg);color:var(--text)}
.page{display:flex;min-height:100vh;align-items:center;justify-content:center;padding:1rem}
.card{
  width:100%;max-width:24rem;background:var(--card);border:3px solid var(--border);
  box-shadow:8px 8px 0 var(--shadow);
}
.card-header{
  display:flex;flex-direction:column;align-items:center;gap:1rem;
  padding:1.5rem;border-bottom:3px solid var(--border);text-align:center;
}
.icon-circle{
  width:4rem;height:4rem;display:flex;align-items:center;justify-content:center;
  border-radius:50%;border:3px solid var(--indigo);
  background:color-mix(in srgb,var(--indigo) 10%,transparent);
}
.icon-circle svg{width:2rem;height:2rem;color:var(--indigo)}
.card-title{font-family:var(--font-outfit);font-size:1.5rem;font-weight:800;letter-spacing:.05em;text-transform:uppercase}
.card-desc{color:var(--muted);font-size:.875rem;line-height:1.5}
.card-body{padding:1.5rem}
.field{margin-bottom:1.25rem}
.field label{display:block;font-size:.75rem;font-weight:600;letter-spacing:.1em;text-transform:uppercase;margin-bottom:.5rem}
.field input{
  width:100%;height:3rem;padding:0 1rem;border:3px solid var(--border);
  background:var(--bg);color:var(--text);font-family:monospace;font-size:.875rem;
  outline:none;transition:border-color .15s;
}
.field input:focus{border-color:var(--indigo)}
.btn{
  width:100%;height:3rem;border:3px solid var(--border);background:var(--indigo);color:#fff;
  font-family:var(--font-space);font-size:.8rem;font-weight:700;letter-spacing:.1em;
  text-transform:uppercase;cursor:pointer;
  box-shadow:4px 4px 0 var(--shadow);transition:all .1s;
}
.btn:hover{transform:translate(.5px,.5px);box-shadow:2px 2px 0 var(--shadow)}
.btn:active{transform:translate(1px,1px);box-shadow:none}
.btn:disabled{opacity:.5;cursor:not-allowed;transform:none;box-shadow:4px 4px 0 var(--shadow)}
.error{color:#dc2626;font-size:.875rem;margin-bottom:1rem;display:none}
.error.show{display:block}
.logo{font-family:var(--font-outfit);font-size:1.25rem;font-weight:800;letter-spacing:.05em;text-transform:uppercase;color:var(--indigo)}
</style>
</head>
<body>
<div class="page">
  <div class="card">
    <div class="card-header">
      <div class="icon-circle">
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none"
             stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/>
          <path d="M7 11V7a5 5 0 0 1 10 0v4"/>
        </svg>
      </div>
      <div class="logo">Hister</div>
      <div class="card-title">Server Locked</div>
      <div class="card-desc">Enter your encryption password to unlock indexed data.</div>
    </div>
    <div class="card-body">
      <div id="error" class="error"></div>
      <div class="field">
        <label for="pw">Password</label>
        <input id="pw" type="password" placeholder="Encryption password" autofocus/>
      </div>
      <button id="btn" class="btn">Unlock</button>
    </div>
  </div>
</div>
<script>
(function(){
  var btn=document.getElementById('btn'),
      pw=document.getElementById('pw'),
      err=document.getElementById('error');
  function unlock(){
    err.className='error';btn.disabled=true;btn.textContent='Unlocking\u2026';
    fetch('/api/unlock',{
      method:'POST',headers:{'Content-Type':'application/json'},
      body:JSON.stringify({password:pw.value})
    }).then(function(r){return r.json()}).then(function(d){
      if(d.status==='unlocked'){window.location.href='/';return}
      err.textContent=d.error||'Wrong password';err.className='error show';
    }).catch(function(){err.textContent='Connection error';err.className='error show';
    }).finally(function(){btn.disabled=false;btn.textContent='Unlock'});
  }
  btn.addEventListener('click',unlock);
  pw.addEventListener('keydown',function(e){if(e.key==='Enter')unlock()});
})();
</script>
</body>
</html>`



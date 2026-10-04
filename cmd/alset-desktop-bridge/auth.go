package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type role string

const (
	roleMaster role = "master"
	roleUser   role = "user"
	roleGuest  role = "guest"
)

type account struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Role         role      `json:"role"`
	PassHash     string    `json:"pass_hash,omitempty"`
	PublicKey    string    `json:"public_key,omitempty"`
	OrganismCID  string    `json:"organism_cid"`
	Created      time.Time `json:"created"`
	LastSeen     time.Time `json:"last_seen"`
}

type session struct {
	Token     string    `json:"token"`
	AccountID string    `json:"account_id"`
	Role      role      `json:"role"`
	Expires   time.Time `json:"expires"`
}

type authStore struct {
	mu       sync.Mutex
	path     string
	Accounts map[string]*account `json:"accounts"`
	Sessions map[string]*session `json:"-"`
}

func newAuthStore(dataDir string) *authStore {
	a := &authStore{
		path:     filepath.Join(dataDir, "accounts.json"),
		Accounts: map[string]*account{},
		Sessions: map[string]*session{},
	}
	a.load()
	if _, ok := a.Accounts["master"]; !ok {
		a.Accounts["master"] = &account{
			ID:          "master",
			Name:        "Master",
			Role:        roleMaster,
			PassHash:    hashPass("master"),
			OrganismCID: "cid:account:master",
			Created:     time.Now().UTC(),
			LastSeen:    time.Now().UTC(),
		}
		_ = a.save()
	}
	return a
}

func hashPass(p string) string {
	sum := sha256.Sum256([]byte("alset:" + p))
	return hex.EncodeToString(sum[:])
}

func (a *authStore) load() {
	raw, err := os.ReadFile(a.path)
	if err != nil {
		return
	}
	_ = json.Unmarshal(raw, a)
	if a.Accounts == nil {
		a.Accounts = map[string]*account{}
	}
	if a.Sessions == nil {
		a.Sessions = map[string]*session{}
	}
}

func (a *authStore) save() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	raw, _ := json.MarshalIndent(struct {
		Accounts map[string]*account `json:"accounts"`
	}{a.Accounts}, "", "  ")
	return os.WriteFile(a.path, raw, 0o600)
}

func (a *authStore) login(name, pass string) (*session, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	var acc *account
	for _, x := range a.Accounts {
		if strings.EqualFold(x.Name, name) || x.ID == strings.ToLower(name) {
			acc = x
			break
		}
	}
	if acc == nil {
		return nil, errAuth("cuenta no encontrada")
	}
	if acc.PassHash != hashPass(pass) {
		return nil, errAuth("credenciales inválidas")
	}
	tok := randomToken(24)
	s := &session{
		Token:     tok,
		AccountID: acc.ID,
		Role:      acc.Role,
		Expires:   time.Now().Add(12 * time.Hour),
	}
	a.Sessions[tok] = s
	acc.LastSeen = time.Now().UTC()
	return s, nil
}

func (a *authStore) sessionFrom(r *http.Request) *session {
	tok := r.Header.Get("X-Alset-Token")
	if tok == "" {
		tok = r.URL.Query().Get("token")
	}
	if tok == "" {
		if c, err := r.Cookie("alset_token"); err == nil {
			tok = c.Value
		}
	}
	if tok == "" {
		return nil
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	s := a.Sessions[tok]
	if s == nil || time.Now().After(s.Expires) {
		return nil
	}
	return s
}

func (a *authStore) require(r *http.Request, min role) (*session, error) {
	s := a.sessionFrom(r)
	if s == nil {
		// bootstrap: allow local unauthenticated as Master for single-user desktop
		return &session{Token: "local-master", AccountID: "master", Role: roleMaster, Expires: time.Now().Add(24 * time.Hour)}, nil
	}
	order := map[role]int{roleGuest: 1, roleUser: 2, roleMaster: 3}
	if order[s.Role] < order[min] {
		return nil, errAuth("permiso denegado")
	}
	return s, nil
}

type authError string

func (e authError) Error() string { return string(e) }
func errAuth(s string) error      { return authError(s) }

func randomToken(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (b *bridge) handleAuthLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
		Pass string `json:"pass"`
	}
	_ = json.NewDecoder(io.LimitReader(r.Body, 1<<16)).Decode(&body)
	if body.Name == "" {
		body.Name = "Master"
	}
	if body.Pass == "" {
		body.Pass = "master"
	}
	s, err := b.auth.login(body.Name, body.Pass)
	if err != nil {
		http.Error(w, err.Error(), 401)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "alset_token", Value: s.Token, Path: "/", HttpOnly: true})
	b.audit(s.AccountID, "auth.login", body.Name)
	writeJSON(w, map[string]any{"ok": true, "token": s.Token, "role": s.Role, "account_id": s.AccountID})
}

func (b *bridge) handleAuthMe(w http.ResponseWriter, r *http.Request) {
	s, err := b.auth.require(r, roleGuest)
	if err != nil {
		http.Error(w, err.Error(), 401)
		return
	}
	b.auth.mu.Lock()
	acc := b.auth.Accounts[s.AccountID]
	b.auth.mu.Unlock()
	writeJSON(w, map[string]any{"ok": true, "session": s, "account": acc})
}

func (b *bridge) handleAuthAccounts(w http.ResponseWriter, r *http.Request) {
	s, err := b.auth.require(r, roleMaster)
	if err != nil {
		http.Error(w, err.Error(), 403)
		return
	}
	_ = s
	b.auth.mu.Lock()
	list := make([]*account, 0, len(b.auth.Accounts))
	for _, a := range b.auth.Accounts {
		cp := *a
		cp.PassHash = ""
		list = append(list, &cp)
	}
	b.auth.mu.Unlock()
	writeJSON(w, map[string]any{"ok": true, "accounts": list})
}

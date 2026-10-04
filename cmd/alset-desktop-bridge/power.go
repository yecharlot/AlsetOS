package main

import (
	"net/http"
	"os/exec"
)

func (b *bridge) handlePowerShutdown(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	if _, err := b.auth.require(r, roleMaster); err != nil {
		http.Error(w, err.Error(), 403)
		return
	}
	b.audit("master", "power.shutdown", "requested")
	writeJSON(w, map[string]any{"ok": true, "message": "apagando…"})
	go func() {
		_ = exec.Command("poweroff").Start()
		_ = exec.Command("sudo", "poweroff").Start()
		_ = exec.Command("halt", "-p").Start()
	}()
}

func (b *bridge) handlePowerReboot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	if _, err := b.auth.require(r, roleMaster); err != nil {
		http.Error(w, err.Error(), 403)
		return
	}
	b.audit("master", "power.reboot", "requested")
	writeJSON(w, map[string]any{"ok": true, "message": "reiniciando…"})
	go func() {
		_ = exec.Command("reboot").Start()
		_ = exec.Command("sudo", "reboot").Start()
	}()
}

package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os/exec"
	"strings"
)

func (b *bridge) handleWifiScan(w http.ResponseWriter, r *http.Request) {
	// Best-effort: nmcli, iw, iwlist
	var networks []map[string]string
	if out, err := exec.Command("nmcli", "-t", "-f", "SSID,SIGNAL,SECURITY", "dev", "wifi", "list").CombinedOutput(); err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			parts := strings.Split(line, ":")
			if len(parts) >= 1 && parts[0] != "" {
				n := map[string]string{"ssid": parts[0], "source": "nmcli"}
				if len(parts) > 1 {
					n["signal"] = parts[1]
				}
				if len(parts) > 2 {
					n["security"] = parts[2]
				}
				networks = append(networks, n)
			}
		}
	}
	if len(networks) == 0 {
		if out, err := exec.Command("iw", "dev").CombinedOutput(); err == nil {
			_ = out
		}
		// placeholder when tools missing
		if len(networks) == 0 {
			networks = append(networks, map[string]string{
				"ssid": "(escaneo no disponible en este host)",
				"note": "En TinyCore instala wifi/firmware; usa nmcli o iw",
			})
		}
	}
	writeJSON(w, map[string]any{"ok": true, "networks": networks})
}

func (b *bridge) handleWifiConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	if _, err := b.auth.require(r, roleUser); err != nil {
		http.Error(w, err.Error(), 403)
		return
	}
	var body struct {
		SSID string `json:"ssid"`
		Pass string `json:"pass"`
		Iface string `json:"iface"`
	}
	_ = json.NewDecoder(io.LimitReader(r.Body, 1<<16)).Decode(&body)
	if body.SSID == "" {
		http.Error(w, "ssid required", 400)
		return
	}
	b.audit("master", "wifi.connect", body.SSID)

	// Try NetworkManager
	args := []string{"dev", "wifi", "connect", body.SSID}
	if body.Pass != "" {
		args = append(args, "password", body.Pass)
	}
	if body.Iface != "" {
		args = append(args, "ifname", body.Iface)
	}
	out, err := exec.Command("nmcli", args...).CombinedOutput()
	if err == nil {
		writeJSON(w, map[string]any{"ok": true, "method": "nmcli", "output": string(out)})
		return
	}
	// Try iw + wpa_supplicant is complex; report failure clearly
	writeJSON(w, map[string]any{
		"ok":     false,
		"error":  "no se pudo conectar (¿nmcli/wifi instalado?)",
		"detail": string(out) + " " + err.Error(),
		"hint":   "En TinyCore: tce-load -wi wifi firmware + NetworkManager o wpa_supplicant",
	})
}

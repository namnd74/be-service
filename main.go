package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	buildVersion = "dev-local"
	buildCommit  = "dev-local"
	envName      = getEnv("APP_ENV", "dev")
	port         = getEnv("PORT", "8080")
	dbSecret     = getEnv("DB_PASSWORD", "")
	startTime    = time.Now()

	mu        sync.RWMutex
	isHealthy = true
)

type PageData struct {
	ServiceName string
	Version     string
	GitCommit   string
	Environment string
	Hostname    string
	Uptime      string
	IsHealthy   bool
	SecretState string
	DemoEnabled bool
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func appVersion() string { return getEnv("APP_VERSION", buildVersion) }

func appCommit() string { return getEnv("GIT_COMMIT", buildCommit) }

func envTrue(key string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func demoAllowed() bool {
	return envName == "dev" && envTrue("DEMO_MODE")
}

func persistentDemoFault() bool {
	return demoAllowed() && envName == "dev" && envTrue("DEMO_FAULT")
}

func demoRequestAllowed(r *http.Request) bool {
	return r.Method == http.MethodGet && demoAllowed()
}

var tmpl = template.Must(template.New("index").Parse(`<!DOCTYPE html>
<html lang="vi">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.ServiceName}} - GitOps Demo</title>
    <style>
        :root {
            --bg: #0f172a;
            --card-bg: #1e293b;
            --text: #f8fafc;
            --muted: #94a3b8;
            --accent-dev: #38bdf8;
            --accent-staging: #f59e0b;
            --accent-prod: #10b981;
            --danger: #ef4444;
            --success: #22c55e;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background-color: var(--bg);
            color: var(--text);
            margin: 0;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
        }
        .container {
            width: 90%;
            max-width: 650px;
            background: var(--card-bg);
            border-radius: 16px;
            padding: 32px;
            box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5);
            border: 1px solid rgba(255, 255, 255, 0.1);
        }
        .badge-env {
            display: inline-block;
            padding: 6px 16px;
            border-radius: 999px;
            font-size: 0.85rem;
            font-weight: 700;
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }
        .env-dev { background: rgba(56, 189, 248, 0.2); color: var(--accent-dev); border: 1px solid var(--accent-dev); }
        .env-staging { background: rgba(245, 158, 11, 0.2); color: var(--accent-staging); border: 1px solid var(--accent-staging); }
        .env-prod { background: rgba(16, 185, 129, 0.2); color: var(--accent-prod); border: 1px solid var(--accent-prod); }

        h1 { margin: 16px 0 8px 0; font-size: 1.8rem; }
        .subtitle { color: var(--muted); font-size: 0.95rem; margin-bottom: 24px; }

        .grid {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 16px;
            margin-bottom: 24px;
        }
        .card {
            background: rgba(15, 23, 42, 0.6);
            border-radius: 10px;
            padding: 14px 18px;
            border: 1px solid rgba(255, 255, 255, 0.05);
        }
        .card-label { font-size: 0.75rem; color: var(--muted); text-transform: uppercase; letter-spacing: 0.05em; }
        .card-val { font-size: 1.1rem; font-weight: 600; margin-top: 4px; word-break: break-all; }

        .health-status {
            padding: 12px;
            border-radius: 8px;
            text-align: center;
            font-weight: 600;
            margin-bottom: 20px;
        }
        .status-ok { background: rgba(34, 197, 94, 0.15); color: var(--success); border: 1px solid var(--success); }
        .status-fail { background: rgba(239, 68, 68, 0.15); color: var(--danger); border: 1px solid var(--danger); }

        .actions {
            display: flex;
            gap: 12px;
            justify-content: center;
        }
        .btn {
            padding: 10px 18px;
            border-radius: 8px;
            border: none;
            cursor: pointer;
            font-weight: 600;
            text-decoration: none;
            transition: all 0.2s;
            font-size: 0.85rem;
        }
        .btn-danger { background: var(--danger); color: white; }
        .btn-success { background: var(--success); color: white; }
        .btn:hover { opacity: 0.85; transform: translateY(-1px); }
    </style>
</head>
<body>
    <div class="container">
        <div>
            <span class="badge-env env-{{.Environment}}">Môi Trường: {{.Environment}}</span>
        </div>
        <h1>{{.ServiceName}}</h1>
        <div class="subtitle">GitOps Demo version Multi-Environment Backend Microservice</div>

        <div class="health-status {{if .IsHealthy}}status-ok{{else}}status-fail{{end}}">
            Liveness Probe: {{if .IsHealthy}}🟢 HEALTHY (HTTP 200){{else}}🔴 UNHEALTHY (HTTP 500) - Simulating Crash{{end}}
        </div>

		<div class="grid">
		    <div class="card">
		        <div class="card-label">App Version</div>
		        <div class="card-val" style="color: #38bdf8;">{{.Version}}</div>
            </div>
            <div class="card">
                <div class="card-label">Git Commit SHA</div>
                <div class="card-val">{{.GitCommit}}</div>
            </div>
            <div class="card">
                <div class="card-label">Pod Hostname</div>
                <div class="card-val">{{.Hostname}}</div>
            </div>
            <div class="card">
                <div class="card-label">Uptime</div>
                <div class="card-val">{{.Uptime}}</div>
            </div>
        </div>

        <div class="card" style="margin-bottom: 20px;">
            <div class="card-label">Secret Reference (Decrypted via Sealed Secrets)</div>
            <div class="card-val" style="color: #a78bfa; font-family: monospace;">{{.SecretState}}</div>
        </div>

        {{if .DemoEnabled}}
        <div class="actions">
            <a href="/simulate-crash" class="btn btn-danger">Trigger demo failure</a>
            <a href="/simulate-fix" class="btn btn-success">Restore demo health</a>
        </div>
        {{end}}
    </div>
</body>
</html>`))

func handleRoot(w http.ResponseWriter, r *http.Request) {
	hostname, _ := os.Hostname()
	secretDisplay := "Secret missing"
	if dbSecret != "" {
		secretDisplay = "Secret loaded"
	}

	mu.RLock()
	healthy := isHealthy || !demoAllowed()
	data := PageData{
		ServiceName: "Backend API Service - Seminar 20261007T061807Z-90462-0001-failure",
		Version:     appVersion(),
		GitCommit:   appCommit(),
		Environment: envName,
		Hostname:    hostname,
		Uptime:      time.Since(startTime).Truncate(time.Second).String(),
		IsHealthy:   healthy,
		SecretState: secretDisplay,
		DemoEnabled: demoAllowed(),
	}
	mu.RUnlock()

	tmpl.Execute(w, data)
}

func handleVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"service":    "be-service",
		"version":    appVersion(),
		"git_commit": appCommit(),
		"env":        envName,
	}
	mu.RLock()
	faultMode := demoAllowed() && (!isHealthy || persistentDemoFault())
	mu.RUnlock()
	if faultMode {
		// This is a mode indicator only; no secret or fault details are exposed.
		response["fault_mode"] = "enabled"
	}
	json.NewEncoder(w).Encode(response)
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()
	if demoAllowed() && !isHealthy {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "FAIL: Simulated crash state")
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}

func handleSimulateCrash(w http.ResponseWriter, r *http.Request) {
	handleDemoHealth(w, r, false, "[DEMO] Simulated crash triggered! /healthz will return 500.")
}

func handleSimulateFix(w http.ResponseWriter, r *http.Request) {
	handleDemoHealth(w, r, true, "[DEMO] Health restored!")
}

func handleDemoHealth(w http.ResponseWriter, r *http.Request, healthy bool, message string) {
	if !demoRequestAllowed(r) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		} else {
			http.NotFound(w, r)
		}
		return
	}
	mu.Lock()
	isHealthy = healthy
	mu.Unlock()
	log.Print(message)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func main() {
	if persistentDemoFault() {
		isHealthy = false
	}
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/version", handleVersion)
	http.HandleFunc("/healthz", handleHealthz)
	http.HandleFunc("/simulate-crash", handleSimulateCrash)
	http.HandleFunc("/simulate-fix", handleSimulateFix)

	addr := ":" + port
	log.Printf("Backend service starting on %s (env=%s, version=%s, commit=%s)", addr, envName, appVersion(), appCommit())
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func preserveGlobalState(t *testing.T) {
	t.Helper()
	mu.RLock()
	oldHealthy := isHealthy
	mu.RUnlock()
	oldBuildVersion, oldBuildCommit := buildVersion, buildCommit
	oldEnvName, oldPort, oldSecret := envName, port, dbSecret
	oldStartTime := startTime
	t.Cleanup(func() {
		mu.Lock()
		isHealthy = oldHealthy
		mu.Unlock()
		buildVersion, buildCommit = oldBuildVersion, oldBuildCommit
		envName, port, dbSecret = oldEnvName, oldPort, oldSecret
		startTime = oldStartTime
	})
}

func TestHealthzInitiallyOK(t *testing.T) {
	preserveGlobalState(t)
	envName = "dev"
	t.Setenv("DEMO_MODE", "false")
	isHealthy = false
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	handleHealthz(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected HTTP 200, got %d", w.Code)
	}
}

func TestVersionEndpoint(t *testing.T) {
	preserveGlobalState(t)
	envName = "dev"
	t.Setenv("DEMO_MODE", "false")
	t.Setenv("APP_VERSION", "v9.9.9")
	t.Setenv("GIT_COMMIT", "abc123")
	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	w := httptest.NewRecorder()
	handleVersion(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected HTTP 200, got %d", w.Code)
	}
	var got map[string]string
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode version response: %v", err)
	}
	if got["version"] != "v9.9.9" || got["git_commit"] != "abc123" {
		t.Fatalf("metadata override not reflected: %#v", got)
	}
}

func TestRootNeverDisplaysSecretBytes(t *testing.T) {
	preserveGlobalState(t)
	t.Setenv("DEMO_MODE", "false")
	dbSecret = "super-secret"
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handleRoot(w, req)
	if strings.Contains(w.Body.String(), "super-secret") || strings.Contains(w.Body.String(), "sup") {
		t.Fatalf("root leaked secret material: %s", w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "Secret loaded") {
		t.Fatalf("root did not show loaded state")
	}
}

func TestRootShowsMissingSecretWithoutFallback(t *testing.T) {
	preserveGlobalState(t)
	t.Setenv("DEMO_MODE", "false")
	dbSecret = ""
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handleRoot(w, req)
	if strings.Contains(w.Body.String(), "mock-secret") || !strings.Contains(w.Body.String(), "Secret missing") {
		t.Fatalf("root did not show missing state safely: %s", w.Body.String())
	}
}

func TestDemoFaultControlsRefusedInProduction(t *testing.T) {
	preserveGlobalState(t)
	envName = "prod"
	t.Setenv("DEMO_MODE", "true")
	isHealthy = false
	assertDemoActionsRefused(t)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	handleHealthz(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("production health must remain OK, got %d", w.Code)
	}
}

func TestDemoActionsRefusedInStaging(t *testing.T) {
	preserveGlobalState(t)
	envName = "staging"
	t.Setenv("DEMO_MODE", "true")
	isHealthy = false
	assertDemoActionsRefused(t)
}

func assertDemoActionsRefused(t *testing.T) {
	t.Helper()
	for _, path := range []string{"/simulate-crash", "/simulate-fix"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		if path == "/simulate-crash" {
			handleSimulateCrash(w, req)
		} else {
			handleSimulateFix(w, req)
		}
		if w.Code != http.StatusNotFound {
			t.Fatalf("%s returned %d, want 404", path, w.Code)
		}
	}
}

func TestTransientDemoEndpointsRequireGET(t *testing.T) {
	preserveGlobalState(t)
	envName = "dev"
	t.Setenv("DEMO_MODE", "true")
	req := httptest.NewRequest(http.MethodPost, "/simulate-crash", nil)
	w := httptest.NewRecorder()
	handleSimulateCrash(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST /simulate-crash returned %d, want 405", w.Code)
	}
}

func TestDemoActionsWorkInDevelopment(t *testing.T) {
	preserveGlobalState(t)
	envName = "dev"
	t.Setenv("DEMO_MODE", "true")
	req := httptest.NewRequest(http.MethodGet, "/simulate-crash", nil)
	w := httptest.NewRecorder()
	handleSimulateCrash(w, req)
	if w.Code != http.StatusSeeOther || isHealthy {
		t.Fatalf("crash action status=%d healthy=%v", w.Code, isHealthy)
	}
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/simulate-fix", nil)
	handleSimulateFix(w, req)
	if w.Code != http.StatusSeeOther || !isHealthy {
		t.Fatalf("fix action status=%d healthy=%v", w.Code, isHealthy)
	}
}

func TestVersionReportsDemoFaultMode(t *testing.T) {
	preserveGlobalState(t)
	envName = "dev"
	t.Setenv("DEMO_MODE", "true")
	t.Setenv("DEMO_FAULT", "true")
	isHealthy = true
	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	w := httptest.NewRecorder()
	handleVersion(w, req)
	var got map[string]string
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got["fault_mode"] != "enabled" {
		t.Fatalf("fault mode missing from version response: %#v", got)
	}
}

func TestPersistentDemoFaultIsDevOnly(t *testing.T) {
	preserveGlobalState(t)
	envName = "staging"
	t.Setenv("DEMO_MODE", "true")
	t.Setenv("DEMO_FAULT", "true")
	isHealthy = false
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	handleHealthz(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("persistent staging fault returned %d, want 200", w.Code)
	}
	w = httptest.NewRecorder()
	handleVersion(w, req)
	var got map[string]string
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if _, ok := got["fault_mode"]; ok {
		t.Fatalf("staging persistent fault leaked into version response: %#v", got)
	}
}

func TestAppMetadataUsesBuildDefaultsWhenOverridesMissing(t *testing.T) {
	preserveGlobalState(t)
	t.Setenv("APP_VERSION", "")
	t.Setenv("GIT_COMMIT", "")
	buildVersion = "v1.2.3"
	buildCommit = "built-sha"
	if got := appVersion(); got != buildVersion {
		t.Fatalf("appVersion()=%q, want %q", got, buildVersion)
	}
	if got := appCommit(); got != buildCommit {
		t.Fatalf("appCommit()=%q, want %q", got, buildCommit)
	}
}

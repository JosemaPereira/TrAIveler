// Package integration holds black-box HTTP integration tests that exercise
// the compiled backend binary as a real, running process rather than
// importing its internals directly: cmd/api is `package main`, and Go does
// not allow importing a package whose clause is "main" from anywhere else,
// so this is the only way to genuinely exercise cmd/api/routes.go's actual
// route table (as opposed to a hand-rebuilt copy of it living only in this
// test file).
package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// swaggerUIBundleURLPattern extracts the value httpSwagger.Handler's index.html
// template renders into SwaggerUIBundle's `url: "..."` config field (see
// github.com/swaggo/http-swagger/v2's indexTempl). It intentionally requires
// the `url:` key immediately followed by a quoted value so it does not
// accidentally match the template's separate `validatorUrl: null,` field,
// which is unquoted.
var swaggerUIBundleURLPattern = regexp.MustCompile(`url:\s*"([^"]+)"`)

// startPostgresContainer starts a disposable PostgreSQL testcontainer and
// returns its connection string. Mirrors
// internal/database/client_test.go's setupPostgresContainer and
// internal/example/repository_integration_test.go's setupRepositoryTestDB
// (same image, credentials, and wait strategy) so this package follows the
// same testcontainer convention already established elsewhere in the repo.
// No migrations are applied here: the backend binary only needs a reachable
// Postgres to start (database.NewClient's startup check is a bare `SELECT
// 1`), and none of the assertions below touch persisted data.
func startPostgresContainer(ctx context.Context, t *testing.T) string {
	t.Helper()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err, "failed to start postgres container")
	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate postgres container: %v", err)
		}
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "failed to get postgres connection string")

	return connStr
}

// buildAPIBinary compiles cmd/api into a temporary binary and returns its
// path. Building from the module import path (rather than a relative
// filesystem path) means this works regardless of the working directory
// `go test` runs from.
func buildAPIBinary(t *testing.T) string {
	t.Helper()

	binPath := filepath.Join(t.TempDir(), "traiveler-api")
	cmd := exec.Command("go", "build", "-o", binPath, "github.com/JosemaPereira/TrAIveler/backend/cmd/api")

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "failed to build backend API binary: %s", output)

	return binPath
}

// freeTCPPort finds an available TCP port on the loopback interface by
// briefly binding to port 0 and reading back what the OS assigned. There is
// an inherent, accepted TOCTOU race between closing this listener and the
// backend binary binding the same port, the same trade-off other
// dynamic-port test setups make.
func freeTCPPort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err, "failed to reserve a free TCP port")
	defer listener.Close()

	return listener.Addr().(*net.TCPAddr).Port
}

// waitForHealthy polls url until it returns 200 OK or timeout elapses,
// failing the test with the last observed error otherwise. Used to block
// until the just-started backend process is actually ready to serve
// requests, since process start and "accepting connections" are not the
// same instant.
func waitForHealthy(t *testing.T, url string, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err != nil {
			lastErr = err
		} else {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
			lastErr = fmt.Errorf("unexpected status code %d from %s", resp.StatusCode, url)
		}
		time.Sleep(200 * time.Millisecond)
	}

	t.Fatalf("backend API did not become healthy within %s: %v", timeout, lastErr)
}

// startAPIServer starts binPath as a child process wired to databaseURL on a
// freshly-allocated port, waits for it to report healthy via /healthz, and
// registers a t.Cleanup to terminate it. Returns the server's base URL.
func startAPIServer(ctx context.Context, t *testing.T, binPath, databaseURL string) string {
	t.Helper()

	port := freeTCPPort(t)
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	cmd := exec.CommandContext(ctx, binPath)
	cmd.Env = append(os.Environ(),
		"DATABASE_URL="+databaseURL,
		fmt.Sprintf("HTTP_PORT=%d", port),
		"LOG_LEVEL=error",
	)

	var output strings.Builder
	cmd.Stdout = &output
	cmd.Stderr = &output

	require.NoError(t, cmd.Start(), "failed to start backend API process")

	t.Cleanup(func() {
		if cmd.Process != nil {
			_ = cmd.Process.Signal(syscall.SIGTERM)

			done := make(chan struct{})
			go func() {
				_ = cmd.Wait()
				close(done)
			}()

			select {
			case <-done:
			case <-time.After(5 * time.Second):
				_ = cmd.Process.Kill()
			}
		}
		if t.Failed() {
			t.Logf("backend API process output:\n%s", output.String())
		}
	})

	waitForHealthy(t, baseURL+"/healthz", 20*time.Second)

	return baseURL
}

// setupSwaggerTestServer boots a real backend API process against a
// disposable Postgres testcontainer and returns its base URL, so tests in
// this file can make genuine HTTP requests against the actual
// cmd/api/routes.go route table instead of a hand-built substitute.
func setupSwaggerTestServer(t *testing.T) string {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping integration test requiring a container runtime in short mode")
	}

	ctx := context.Background()
	databaseURL := startPostgresContainer(ctx, t)
	binPath := buildAPIBinary(t)

	return startAPIServer(ctx, t, binPath, databaseURL)
}

// TestSwaggerDocJSON_RouteMounted_ReturnsValidSwagger2Document is 009-T007
// (RED). It asserts GET /swagger/doc.json returns 200 with a body that
// parses as a valid Swagger 2.0 (OpenAPI 2.0) document — NOT an OpenAPI v3
// document. This is expected to fail until G-SPEC009-ANNOTATIONS (#114) and
// G-SPEC009-SWAGGER-UI (#115) land: no /swagger/* route exists yet, so the
// request currently 404s.
func TestSwaggerDocJSON_RouteMounted_ReturnsValidSwagger2Document(t *testing.T) {
	baseURL := setupSwaggerTestServer(t)

	resp, err := http.Get(baseURL + "/swagger/doc.json")
	require.NoError(t, err, "GET /swagger/doc.json must succeed at the transport level")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode,
		"expected /swagger/doc.json to be mounted and return 200; "+
			"this fails until #114 (annotations) and #115 (route mount) land")

	var doc map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&doc), "response body must be valid JSON")

	// A valid Swagger 2.0 (OpenAPI 2.0) document declares these top-level
	// fields. OpenAPI v3 documents instead use "openapi": "3.x.x" plus a
	// "components" container - this asserts the shape swaggo/swag actually
	// produces (Swagger 2.0), not v3, per specs/009-api-documentation's
	// corrected research.md/data-model.md.
	assert.Equal(t, "2.0", doc["swagger"], `document must declare top-level "swagger": "2.0"`)
	assert.Contains(t, doc, "paths", `Swagger 2.0 documents have a top-level "paths" map`)
	assert.Contains(t, doc, "definitions",
		`Swagger 2.0 uses top-level "definitions" for schemas, not OpenAPI v3's components.schemas`)
	assert.Contains(t, doc, "securityDefinitions",
		`Swagger 2.0 uses top-level "securityDefinitions", not OpenAPI v3's components.securitySchemes`)
	assert.NotContains(t, doc, "openapi", `must not be an OpenAPI v3 document (no top-level "openapi" field)`)
	assert.NotContains(t, doc, "components", "must not use OpenAPI v3's components container")
}

// TestSwaggerDocJSON_Paths_ContainsExampleEndpointsAndExcludesHealthz is
// 009-T008 (RED). It asserts every internal/example endpoint appears in the
// parsed document's paths (relative to the /api/v1 basePath declared in
// cmd/api/docs.go), and that /healthz - intentionally undocumented, per
// research.md's "Excluding intentionally-undocumented endpoints" decision -
// does not. Expected to fail until #114/#115 land.
func TestSwaggerDocJSON_Paths_ContainsExampleEndpointsAndExcludesHealthz(t *testing.T) {
	baseURL := setupSwaggerTestServer(t)

	resp, err := http.Get(baseURL + "/swagger/doc.json")
	require.NoError(t, err, "GET /swagger/doc.json must succeed at the transport level")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode,
		"expected /swagger/doc.json to be mounted; "+
			"this fails until #114 (annotations) and #115 (route mount) land")

	var doc struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&doc), "response body must be valid JSON")

	// internal/example/handler.go mounts POST/GET /examples and
	// GET/PUT/DELETE /examples/{id} under /api/v1 (see routes.go); paths in
	// the generated document are relative to the /api/v1 basePath.
	expectedOperations := map[string][]string{
		"/examples":      {"post", "get"},
		"/examples/{id}": {"get", "put", "delete"},
	}
	for path, methods := range expectedOperations {
		operations, ok := doc.Paths[path]
		if !assert.True(t, ok, "expected internal/example endpoint %q to appear in the Swagger document's paths", path) {
			continue
		}
		for _, method := range methods {
			assert.Contains(t, operations, method,
				"expected %s %s to be documented", strings.ToUpper(method), path)
		}
	}

	assert.NotContains(t, doc.Paths, "/healthz",
		"/healthz is intentionally undocumented (no swag annotations) and must never appear in paths")
	assert.NotContains(t, doc.Paths, "/api/v1/healthz",
		"/healthz must not appear in paths under any prefix either")
}

// TestSwaggerIndexHTML_RouteMounted_DocURLResolvesToLiveDocJSON is
// 009-T014 (RED)/009-T016 (GREEN) and 009-T015. It asserts GET
// /swagger/index.html returns 200 with an HTML content type — and, the
// substantive check 009-T015 asks for, that the rendered page's
// SwaggerUIBundle is genuinely wired to the *live* /swagger/doc.json route
// mounted in cmd/api/routes.go, not a stale or hand-edited copy.
//
// httpSwagger.Handler() (github.com/swaggo/http-swagger/v2) defaults
// Config.URL to the relative string "doc.json" when routes.go mounts it
// with no explicit httpSwagger.URL(...) override. Because index.html is
// itself served from /swagger/, a browser resolves that relative "doc.json"
// against the current page location to /swagger/doc.json — the same
// r.Get("/swagger/*", httpSwagger.Handler()) route, already covered by
// TestSwaggerDocJSON_RouteMounted_ReturnsValidSwagger2Document above, that
// serves the generated contract. This test proves that resolution actually
// happens (rather than trusting the manual verification from #115) by
// extracting the real `url: "..."` value the server rendered, resolving it
// relative to the request URL exactly as a browser would, and then fetching
// that resolved URL to confirm it lands on a genuine, live Swagger 2.0
// document — not merely asserting the literal template string.
func TestSwaggerIndexHTML_RouteMounted_DocURLResolvesToLiveDocJSON(t *testing.T) {
	baseURL := setupSwaggerTestServer(t)

	indexURL := baseURL + "/swagger/index.html"
	resp, err := http.Get(indexURL)
	require.NoError(t, err, "GET /swagger/index.html must succeed at the transport level")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode,
		"expected /swagger/index.html to be mounted and return 200")
	assert.Contains(t, resp.Header.Get("Content-Type"), "text/html",
		"expected /swagger/index.html to be served with an HTML content type")

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "must be able to read the /swagger/index.html response body")

	matches := swaggerUIBundleURLPattern.FindSubmatch(body)
	require.Len(t, matches, 2,
		"expected to find SwaggerUIBundle's `url: \"...\"` config rendered in the /swagger/index.html body")
	docURL := string(matches[1])

	// Resolve docURL relative to the page it was served from, exactly as a
	// browser's SwaggerUIBundle would, instead of assuming it is already an
	// absolute path.
	parsedIndexURL, err := url.Parse(indexURL)
	require.NoError(t, err, "index URL %q must itself be a valid URL", indexURL)
	resolvedDocURL, err := parsedIndexURL.Parse(docURL)
	require.NoError(t, err, "DocURL %q must resolve as a valid URL relative to %s", docURL, indexURL)

	require.Equal(t, baseURL+"/swagger/doc.json", resolvedDocURL.String(),
		"the Swagger UI page's configured DocURL must resolve to the live /swagger/doc.json route "+
			"mounted in cmd/api/routes.go, not a stale or hardcoded copy")

	// Actually fetch the resolved URL rather than trusting the string match:
	// confirm it serves a genuine, live Swagger 2.0 document.
	docResp, err := http.Get(resolvedDocURL.String())
	require.NoError(t, err, "GET %s (the UI's resolved DocURL) must succeed", resolvedDocURL.String())
	defer docResp.Body.Close()
	require.Equal(t, http.StatusOK, docResp.StatusCode,
		"the DocURL the Swagger UI actually loads must itself return 200")

	var doc map[string]any
	require.NoError(t, json.NewDecoder(docResp.Body).Decode(&doc),
		"the DocURL the Swagger UI actually loads must return valid JSON")
	assert.Equal(t, "2.0", doc["swagger"],
		"the DocURL the Swagger UI actually loads must serve a valid Swagger 2.0 document, "+
			"proving it is the live-generated contract and not a stale copy")
}

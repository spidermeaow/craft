package project

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRev12InstallUnchangedDoesNotRewriteProjectFiles(t *testing.T) {
	t.Setenv("CRAFT_PACKAGE_CACHE", t.TempDir())
	root := t.TempDir()
	newInstallApp(t, root)
	f := newPackageTransport()
	f.add(t, "dev/lib", "1.0.0", strings.Repeat("a", 40), "lib", "export func value():Int{return 1}", "")
	fixtureInstall(t, root, InstallOptions{Spec: "github.com/dev/lib@v1.0.0"}, f)
	manifest := filepath.Join(root, "craft.toml")
	lock := filepath.Join(root, "craft.lock")
	lockData, err := os.ReadFile(lock)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(lock, append(lockData, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	beforeManifest, err := os.Stat(manifest)
	if err != nil {
		t.Fatal(err)
	}
	beforeLock, err := os.Stat(lock)
	if err != nil {
		t.Fatal(err)
	}
	// Replacements create new files; their modification times are irrelevant.
	// File identity checks that an unchanged install did not replace either file.
	messages, err := installWithTransport(context.Background(), root, InstallOptions{Offline: true}, f)
	if err != nil {
		t.Fatal(err)
	}
	afterManifest, _ := os.Stat(manifest)
	afterLock, _ := os.Stat(lock)
	if !os.SameFile(beforeManifest, afterManifest) || !os.SameFile(beforeLock, afterLock) {
		t.Fatal("unchanged install replaced project files")
	}
	if len(messages) == 0 || !strings.Contains(messages[len(messages)-1], "unchanged") {
		t.Fatalf("missing no-change report: %q", messages)
	}
}

func TestRev12CorruptCacheRecoversInFreshCacheWithoutChangingLock(t *testing.T) {
	cache := t.TempDir()
	t.Setenv("CRAFT_PACKAGE_CACHE", cache)
	root := t.TempDir()
	newInstallApp(t, root)
	f := newPackageTransport()
	f.add(t, "dev/lib", "1.0.0", strings.Repeat("a", 40), "lib", "export func value():Int{return 1}", "")
	fixtureInstall(t, root, InstallOptions{Spec: "github.com/dev/lib@v1.0.0"}, f)
	lockPath := filepath.Join(root, "craft.lock")
	before, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := readProjectLock(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	r, err := newRemoteResolver(context.Background(), lock, false)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(r.pinPath(lock.Remotes[0]), "src", "lib.craft"), []byte("tampered"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = installWithTransport(context.Background(), root, InstallOptions{}, f); err == nil || !strings.Contains(err.Error(), "CRAFT_PACKAGE_CACHE") {
		t.Fatalf("corrupted cache was not rejected with recovery hint: %v", err)
	}
	t.Setenv("CRAFT_PACKAGE_CACHE", t.TempDir())
	fixtureInstall(t, root, InstallOptions{}, f)
	after, err := os.ReadFile(lockPath)
	if err != nil || !bytes.Equal(before, after) || f.resolves != 1 || f.downloads != 2 {
		t.Fatalf("fresh-cache restore changed pin or resolved tag: err=%v resolves=%d downloads=%d", err, f.resolves, f.downloads)
	}
}

func TestRev12GitHubTransportRetriesOnlyTransientStatus(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte("ok"))
	}))
	defer server.Close()
	g := &githubTransport{client: server.Client()}
	got, err := g.get(context.Background(), server.URL, 10)
	if err != nil || !bytes.Equal(got, []byte("ok")) || attempts != 2 {
		t.Fatalf("transient status: body=%q err=%v attempts=%d", got, err, attempts)
	}
	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusNotFound)
	})
	attempts = 0
	_, err = g.get(context.Background(), server.URL, 10)
	if err == nil || !strings.Contains(err.Error(), "HTTP 404") || attempts != 1 {
		t.Fatalf("permanent status retried: err=%v attempts=%d", err, attempts)
	}
}

func TestRev12GitHubRetryRespectsCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	g := &githubTransport{client: server.Client()}
	_, err := g.get(ctx, server.URL, 10)
	if err != context.DeadlineExceeded {
		t.Fatalf("retry ignored cancellation: %v", err)
	}
}

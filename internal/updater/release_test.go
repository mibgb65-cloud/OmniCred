package updater

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (call roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return call(request)
}

func jsonResponse(value any) *http.Response {
	data, _ := json.Marshal(value)
	return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(string(data)))}
}

func testManager(t *testing.T) (*Manager, release, manifest, []byte) {
	t.Helper()
	manager := New("0.3.2", false)
	manager.enabled, manager.arch, manager.cacheRoot = true, "amd64", t.TempDir()
	t.Cleanup(manager.Close)
	data := []byte("test-installer-do-not-execute")
	hash := sha256.Sum256(data)
	name := "OmniCred-amd64-installer.exe"
	latest := release{Tag: "v0.4.0", Assets: []releaseAsset{
		{Name: "update-manifest.json", URL: RepositoryURL + "/releases/download/v0.4.0/update-manifest.json", Size: 512},
		{Name: name, URL: RepositoryURL + "/releases/download/v0.4.0/" + name, Size: int64(len(data))},
	}}
	manifest := manifest{Protocol: 1, Version: "v0.4.0", Installers: []manifestInstaller{{
		OS: "windows", Arch: "amd64", Name: name, SHA256: hex.EncodeToString(hash[:]), Size: int64(len(data)),
	}}}
	return manager, latest, manifest, data
}

func serveRelease(manager *Manager, latest release, data manifest, installer []byte) {
	manager.client.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		switch {
		case request.URL.String() == manager.endpoint:
			return jsonResponse(latest), nil
		case strings.HasSuffix(request.URL.Path, "/update-manifest.json"):
			return jsonResponse(data), nil
		default:
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), ContentLength: -1,
				Body: io.NopCloser(strings.NewReader(string(installer)))}, nil
		}
	})
}

func TestCheckReleaseAndManifest(t *testing.T) {
	manager, latest, manifest, data := testManager(t)
	serveRelease(manager, latest, manifest, data)
	info, err := manager.Check(context.Background())
	if err != nil || !info.UpdateAvailable || !info.DownloadAvailable || info.LatestVersion != "v0.4.0" {
		t.Fatalf("Check = %+v, %v", info, err)
	}
	manager.client.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 404, Body: http.NoBody}, nil
	})
	info, err = manager.Check(context.Background())
	if err != nil || info.Status != "no_releases" || info.UpdateAvailable || manager.selected != nil {
		t.Fatalf("empty release = %+v, %v", info, err)
	}
}

func TestRejectUntrustedOrIncompatibleInstallers(t *testing.T) {
	cases := []struct {
		name   string
		change func(*release, *manifest)
	}{
		{"legacy release", func(r *release, _ *manifest) { r.Assets = r.Assets[1:] }},
		{"different version", func(_ *release, m *manifest) { m.Version = "v0.5.0" }},
		{"unknown protocol", func(_ *release, m *manifest) { m.Protocol = 2 }},
		{"missing architecture", func(_ *release, m *manifest) { m.Installers[0].Arch = "arm64" }},
		{"invalid hash", func(_ *release, m *manifest) { m.Installers[0].SHA256 = "bad" }},
		{"path traversal", func(_ *release, m *manifest) { m.Installers[0].Name = "../installer.exe" }},
		{"oversized", func(_ *release, m *manifest) { m.Installers[0].Size = maxInstallerSize + 1 }},
		{"size mismatch", func(r *release, _ *manifest) { r.Assets[1].Size++ }},
		{"foreign host", func(r *release, _ *manifest) { r.Assets[1].URL = "https://evil.example/installer.exe" }},
		{"foreign repository", func(r *release, _ *manifest) {
			r.Assets[1].URL = "https://github.com/attacker/repo/releases/download/v0.4.0/OmniCred-amd64-installer.exe"
		}},
		{"duplicate asset", func(r *release, _ *manifest) { r.Assets = append(r.Assets, r.Assets[1]) }},
		{"duplicate architecture", func(_ *release, m *manifest) { m.Installers = append(m.Installers, m.Installers[0]) }},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			manager, latest, manifest, data := testManager(t)
			test.change(&latest, &manifest)
			serveRelease(manager, latest, manifest, data)
			info, err := manager.Check(context.Background())
			if err != nil || info.DownloadAvailable || info.UnavailableReason == "" || manager.selected != nil {
				t.Fatalf("unsafe release accepted: %+v, %v", info, err)
			}
			if _, err := manager.Download(); err == nil {
				t.Fatal("download must be blocked")
			}
		})
	}
}

func TestStableReleaseValidation(t *testing.T) {
	for _, version := range []string{"v0.3.2", "v0.3.1", "v0.4.0-beta.1", "v0.4.0/evil", "v9999999999.0.0"} {
		t.Run(version, func(t *testing.T) {
			manager, latest, manifest, data := testManager(t)
			latest.Tag = version
			serveRelease(manager, latest, manifest, data)
			info, _ := manager.Check(context.Background())
			if info.UpdateAvailable || info.DownloadAvailable {
				t.Fatalf("invalid update: %+v", info)
			}
		})
	}
	for _, kind := range []string{"draft", "prerelease", "disabled"} {
		t.Run(kind, func(t *testing.T) {
			manager, latest, manifest, data := testManager(t)
			latest.Draft, latest.Prerelease = kind == "draft", kind == "prerelease"
			manager.enabled = kind != "disabled"
			serveRelease(manager, latest, manifest, data)
			info, _ := manager.Check(context.Background())
			if info.DownloadAvailable {
				t.Fatalf("ineligible update: %+v", info)
			}
		})
	}
}

func TestRejectRedirectsAndBoundMetadata(t *testing.T) {
	for _, address := range []string{"http://github.com/mibgb65-cloud/OmniCred/releases/download/v1.0.0/a.exe", "https://localhost/file", "https://github.com:443/file", "https://user@github.com/file", "https://github.com.evil.example/file"} {
		if validateURL(address) == nil {
			t.Fatalf("accepted %s", address)
		}
	}
	manager, _, _, _ := testManager(t)
	manager.client.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 302, Header: http.Header{"Location": {"http://127.0.0.1/secret"}}, Body: http.NoBody}, nil
	})
	if _, err := manager.Check(context.Background()); err == nil {
		t.Fatal("unsafe redirect accepted")
	}
	manager.client.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(strings.Repeat(" ", (1<<20)+1)))}, nil
	})
	if _, err := manager.Check(context.Background()); err == nil {
		t.Fatal("oversized metadata accepted")
	}
}

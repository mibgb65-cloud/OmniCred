package updater

import "time"

const (
	RepositoryURL    = "https://github.com/mibgb65-cloud/OmniCred"
	maxInstallerSize = 512 << 20
)

type Info struct {
	CurrentVersion    string     `json:"current_version"`
	LatestVersion     string     `json:"latest_version"`
	UpdateAvailable   bool       `json:"update_available"`
	DownloadAvailable bool       `json:"download_available"`
	UnavailableReason string     `json:"unavailable_reason,omitempty"`
	ReleaseURL        string     `json:"release_url"`
	PublishedAt       *time.Time `json:"published_at,omitempty"`
	CheckedAt         time.Time  `json:"checked_at"`
	Status            string     `json:"status"`
}

type State struct {
	Phase      string `json:"phase"`
	Version    string `json:"version,omitempty"`
	Downloaded int64  `json:"downloaded"`
	Total      int64  `json:"total"`
	Error      string `json:"error,omitempty"`
}

type releaseAsset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
	Size int64  `json:"size"`
}

type release struct {
	Tag         string         `json:"tag_name"`
	PublishedAt time.Time      `json:"published_at"`
	Draft       bool           `json:"draft"`
	Prerelease  bool           `json:"prerelease"`
	Assets      []releaseAsset `json:"assets"`
}

type manifest struct {
	Protocol   int                 `json:"protocol"`
	Version    string              `json:"version"`
	Installers []manifestInstaller `json:"installers"`
}

type manifestInstaller struct {
	OS     string `json:"os"`
	Arch   string `json:"arch"`
	Name   string `json:"name"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

type candidate struct {
	version string
	asset   releaseAsset
	digest  string
}

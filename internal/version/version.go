package version

// Engine is the Brenox engine release semver (injected at link time via -ldflags).
var Engine = "1.0.0"

// Commit is the git SHA at build time (injected via -ldflags).
var Commit = "unknown"

const API = "v1"

// Snapshot is returned by GET /version.
type Snapshot struct {
	Engine     string `json:"engine"`
	APIVersion string `json:"api_version"`
	Commit     string `json:"commit,omitempty"`
}

func Get() Snapshot {
	return Snapshot{
		Engine:     Engine,
		APIVersion: API,
		Commit:     Commit,
	}
}

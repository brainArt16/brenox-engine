package apps

import db "github.com/brainart16/brenox/internal/db"

const (
	EnvLive    = "live"
	EnvSandbox = "sandbox"
)

func EnvironmentFromSandbox(sandbox bool) string {
	if sandbox {
		return EnvSandbox
	}
	return EnvLive
}

func WorkspaceForSandbox(app db.App, sandbox bool) int64 {
	if sandbox {
		return app.SandboxWorkspaceID
	}
	return app.WorkspaceID
}

func WorkspaceForEnvironment(app db.App, env string) int64 {
	return WorkspaceForSandbox(app, env == EnvSandbox)
}

func IsLiveWorkspace(app db.App, workspaceID int64) bool {
	return workspaceID == app.WorkspaceID
}

func EffectiveApp(app db.App, sandbox bool) db.App {
	app.WorkspaceID = WorkspaceForSandbox(app, sandbox)
	return app
}

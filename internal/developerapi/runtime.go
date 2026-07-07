package developerapi

import (
	"github.com/brainart16/brenox/internal/apps"
	db "github.com/brainart16/brenox/internal/db"
	"github.com/gin-gonic/gin"
)

type RuntimeApp struct {
	App         db.App
	WorkspaceID int64
	Sandbox     bool
	Environment string
}

func runtimeFromContext(c *gin.Context) RuntimeApp {
	app := c.MustGet("app").(db.App)
	sandbox := false
	if raw, ok := c.Get("is_sandbox"); ok {
		sandbox, _ = raw.(bool)
	}
	return RuntimeApp{
		App:         app,
		WorkspaceID: app.WorkspaceID,
		Sandbox:     sandbox,
		Environment: apps.EnvironmentFromSandbox(sandbox),
	}
}

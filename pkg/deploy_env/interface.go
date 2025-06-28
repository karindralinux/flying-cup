package deploy_env

type App struct {
	Name          string
	SourcePath    string
	HostPort      string
	ContainerPort string
}

type DeploymetEnv interface {
	BuildAndDeploy(app *App) error
}

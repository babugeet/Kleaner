package variables

var Found bool
var ServiceName string
var DeploymentName string

func SetGlobalVariable(appName string) {
	ServiceName = appName
	DeploymentName = appName + "-ws-test"
}

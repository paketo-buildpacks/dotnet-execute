package dotnetexecute

// Configuration enumerates the environment variable configuration options
// that govern the buildpack's behaviour.
type Configuration struct {

	// When BP_DEBUG_ENABLED=TRUE, the buildpack will include the Visual Studio
	// Debugger in the app launch image Remote debuggers can invoke vsdbg inside
	// the running app container and attach to vsdbg's exposed port; also, the
	// buildpack will set ASPNETCORE_ENVIRONMENT=Development.
	DebugEnabled bool `env:"BP_DEBUG_ENABLED"`

	// When BP_LIVE_RELOAD_ENABLED=TRUE, the buildpack will make the app's entrypoint
	// process reload on changes to program files in the app container. It will
	// include watchexec in the app launch image and make the default container
	// entrypoint watchexec + <the usual app entrypoint>. See
	// https://github.com/watchexec/watchexec for more on watchexec as a
	// reloadable process manager.
	LiveReloadEnabled bool `env:"BP_LIVE_RELOAD_ENABLED"`

	// BP_LOG_LEVEL determines the amount of logs produced by the buildpack. Set
	// BP_LOG_LEVEL=DEBUG for more detailed logs.
	LogLevel string `env:"BP_LOG_LEVEL,default=INFO"`

	// When BP_DOTNET_PROJECT_PATH is set to a relative path, the buildpack
	// will look for project file(s) in that subdirectory to determine which
	// project to build into the app container.
	ProjectPath string `env:"BP_DOTNET_PROJECT_PATH"`
}

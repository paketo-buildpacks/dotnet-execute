# .NET Execute Cloud Native Buildpack

The .NET Execute CNB sets the start command for a given .Net Core
application once it has been built by preceding buildpacks.

## Integration

The .NET Execute CNB completes the setup of a .Net Core application
built using a sequence of CNBs. As such, it will be the only non-optional CNB
in that sequence and is not explicitly required by any CNB that precedes it.

It provides `dotnet-execute` as a dependency, but currently there's no
scenario we can imagine that you would use a downstream buildpack to require
this dependency. If a user likes to include some other functionality, it can be
done independent of the .NET Execute CNB without requiring a dependency
of it.

To package this buildpack for consumption:
```
$ ./scripts/package.sh
```
This builds the buildpack's source using GOOS=linux by default. You can supply
another value as the first argument to package.sh.

## Specifying a project path

To specify a project subdirectory (i.e. the directory containing your
`.csproj`/`.fsproj`/`.vbproj` file), please use the BP_DOTNET_PROJECT_PATH
environment variable at build time either directly (e.g. pack build my-app
--env BP_DOTNET_PROJECT_PATH=./src/my-app) or through a project.toml file. This
configuration does not apply to FDD, FDE or Self-Contained app deployments.
?

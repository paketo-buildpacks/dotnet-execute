# Dotnet Execute Cloud Native Buildpack

The Dotnet Execute CNB sets the start command for a given .Net Core
application once it has been built by preceding buildpacks.

## Integration

The Dotnet Execute CNB completes the setup of a .Net Core application
built using a sequence of CNBs. As such, it will be the only non-optional CNB
in that sequence and is not explicitly required by any CNB that precedes it.

It provides `dotnet-execute` as a dependency, but currently there's no
scenario we can imagine that you would use a downstream buildpack to require
this dependency. If a user likes to include some other functionality, it can be
done independent of the Dotnet Execute CNB without requiring a dependency
of it.

To package this buildpack for consumption:
```
$ ./scripts/package.sh
```
This builds the buildpack's source using GOOS=linux by default. You can supply
another value as the first argument to package.sh.

## `buildpack.yml` Configurations

There are no extra configurations for this buildpack based on `buildpack.yml`.
If you would like to specify an `project-path` constraint for the dotnet-build
buildpack, see its
[README](https://github.com/paketo-buildpacks/dotnet-core-build/blob/master/README.md#buildpackyml-configurations).

# Dotnet Core Conf Cloud Native Buildpack

The Dotnet Core Conf CNB sets the start command for a given Dotnet Core application once it
has been built by preceding buildpacks.

## Integration

The Dotnet Core Conf CNB completes the setup of a Dotnet Core application built
using a sequence of CNBs. As such, it will be the only non-optional CNB in that
sequence and is not explicitly required by any CNB that precedes it.

To package this buildpack for consumption:
```
$ ./scripts/package.sh
```
This builds the buildpack's source using GOOS=linux by default. You can supply another value as the first argument to package.sh.

## `buildpack.yml` Configurations

There are no extra configurations for this buildpack based on `buildpack.yml`. If you would like to specify an `project-path`
constraint for the dotnet-build buildpack, see its [README](https://github.com/paketo-buildpacks/dotnet-core-build/blob/master/README.md#buildpackyml-configurations).

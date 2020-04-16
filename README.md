# Dotnet Core Conf Cloud Native Buildpack

The Dotnet Core Conf CNB sets the start command for a given Dotnet Core application once it
has been built by preceding buildpacks.

## Integration

The Dotnet Core Conf CNB completes the setup of a Dotnet Core application built
using a sequence of CNBs. As such, it will be the only non-optional CNB in that
sequence and is not explicitly required by any CNB that precedes it.

Downstream buildpacks can require the `dotnet-core-conf` dependency, however this
buildpack signifies the end of the Dotnet group build processes, so any extension to
this could be included in other independent buildpacks. Requiring dotnet-core-conf is
not a workflow that is supported.

To package this buildpack for consumption:
```
$ ./scripts/package.sh
```
This builds the buildpack's source using GOOS=linux by default. You can supply another value as the first argument to package.sh.

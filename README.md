# .NET Core Conf Cloud Native Buildpack

The Dotnet Core Conf CNB sets the start command for a given application once it
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

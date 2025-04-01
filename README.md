# `paketo-buildpacks/health-checker`

The Paketo Buildpack for Health Checker is a Cloud Native Buildpack that provides a health checker tool that can be used to validate application health.

## Behavior

This buildpack will participate either of the following conditions are met

* `$BP_HEALTH_CHECKER_ENABLED` is set.
* An upstream buildpack requests `hc` in the build plan.

The buildpack will do the following:

* If `$BP_HEALTH_CHECKER_ENABLED` is set to `true`, requests that a health checker be installed by requiring `hc` in the buildplan.
* Contributes the requested health checker set with `$BP_HEALTH_CHECKER_DEPENDENCY` to a layer marked `launch` with command on `$PATH`.
* Creates a symlink to the health check process at `/workspace/health-check`.

## Configuration

| Environment Variable            | Description                                                                                                                       |
| ------------------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| `$BP_HEALTH_CHECKER_ENABLED`    | If set to `true` the buildpack will contribute a health checker binary. Defaults to `false`, so no health checker is contributed. |
| `$BP_HEALTH_CHECKER_DEPENDENCY` | The dependency id, see [`buildpack.toml`](buildpack.toml), of the health checker to install. Defaults to `thc`.                   |

## Bindings

The buildpack optionally accepts the following bindings:

### Type: `dependency-mapping`

| Key                   | Value   | Description                                                                                       |
| --------------------- | ------- | ------------------------------------------------------------------------------------------------- |
| `<dependency-digest>` | `<uri>` | If needed, the buildpack will fetch the dependency with digest `<dependency-digest>` from `<uri>` |

## License

This buildpack is released under version 2.0 of the [Apache License][a].

[a]: http://www.apache.org/licenses/LICENSE-2.0

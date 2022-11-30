# Web Terminal Exec

Web Terminal Exec is a component of [Web Terminal Operator](https://github.com/redhat-developer/web-terminal-operator/) that enables OpenShift clusters to inject a user's kubeconfig into a running container, allowing for automatic login when using pods/exec. It is intended to run as a container within the context of a Web Terminal Operator installation in an OpenShift cluster.


## API
The Web Terminal Exec serves three endpoints
| method | path | body | response | auth required? |
|--------|------|------|----------|----------------|
| `GET` | `/healthz`| N/A | `HTTP 200` | No |
| `POST` | `/activity/tick` | N/A | `HTTP 204` | Yes |
| `POST` | `/exec/init` | JSON | `HTTP 200` + JSON | Yes |

The `/exec/init` endpoint accepts the following JSON:
```jsonc
{
  // Optional container name to inject kubeconfig into; by default search for a suitable container
  "containerName": "<CONTAINER_NAME>",
  "kubeconfig": {
    // Namespace for current context in kubeconfig; optional and unset in kubeconfig if not specified
    "namespace": "<NAMESPACE>",
    // Username for the current user; set to 'Developer' if not specified
    "username": "<USERNAME>"
  }
}
```
The `/exec/init` endpoint responds with JSON containing the information necessary to open a terminal session in the container it injected the kubeconfig into:
```jsonc
{
  // Name of pod in specified namespace
  "pod": "<POD_NAME>",
  // Name of detected container in specified namespace
  "container": "<CONTAINER_NAME>",
  // Detected default shell command (e.g. ["/bin/bash"])
  "cmd": ["<COMMAND>..."]
}
```
This can be consumed in a `kubectl` command as follows:
```
kubectl exec -it <POD_NAME> <CONTAINER_NAME> -- <COMMAND>...
```

### Authentication
Endpoints that require authentication expect a user's OpenShift token to be passed in a `X-Access-Token` or `X-Forwarded-Access-Token` header on the request. This token is used to

1. Verify that the user making the request is the authorized user for the current terminal
2. Execute the pods/exec API call that interacts with the container into which kubeconfig is being injected (if applicable)

If a token is not provided or does not match what is expected, the server returns `HTTP 401`


## Commandline options
```
  --authenticated-user-id string
      OpenShift user's ID that should has access to API. Must be set.
  --idle-timeout duration
      IdleTimeout is a inactivity period after which workspace should be stopped. Use '-1' to disable idle timeout.
      Examples: -1, 30s, 15m, 1h (default 5m0s)
  --pod-selector string
      Selector that is used to find workspace pod. (default controller.devfile.io/devworkspace_id=${DEVWORKSPACE_ID})
  --stop-retry-period duration
      StopRetryPeriod is a period after which workspace should be tried to stop if the previous try failed.
      Examples: 30s (default 10s)
  --url string
      Host:Port address for the Web Terminal Exec server. (default ":4444")
```

## Development
The Web Terminal Exec component is intended to only run within an OpenShift cluster in the context of a Web Terminal Operator installation. To test builds of Web Terminal Exec in an OpenShift cluster, edit the web-terminal-exec DevWorkspaceTemplate created by the Web Terminal Operator to use the in-development image. Alternatively, individual Web Terminal instances can be edited to use a custom image for testing. Running locally or in isolation (i.e. without the Web Terminal Operator installed) is not currently supported.

This repository contains a Makefile to simplify development of Web Terminal Exec. The following environment variables can be set to configure behavior:
* `DOCKER`: configure container build tool to use for building container images. Default is `podman`.
* `WEB_TERMINAL_EXEC_IMG`: Configure image name and tag for container builds. Default is `quay.io/wto/web-terminal-exec:next`.

| Rule | Purpose |
| --- | --- |
| `help` | List available rules |
| `fmt` | Format all Go code. Requires [`go-imports`](https://pkg.go.dev/golang.org/x/tools/cmd/goimports) |
| `fmt_license` | Add license headers to all Go files. Requires [`addlicense`](https://github.com/google/addlicense) |
| `check_fmt` | Check if all files are formatted and contain appropriate license headers |
| `test` | Run Go tests and generate cover.out |
| `vet` | Run `go vet` on all files |
| `docker` | Build and push container image |
| `docker-build` | Build container image (do not push) |
| `docker-push` | Push container image |
| `compile` | Build binary locally |

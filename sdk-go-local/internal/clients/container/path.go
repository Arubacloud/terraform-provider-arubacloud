package container

// Path constants follow the server-canonical lowerCamelCase rule:
//   - Single-word / acronym collections stay lowercase: kaas, registries.
//
// Do not flatten these to all-lowercase. Downstream provisioners store and re-emit
// request URIs verbatim, so a casing change causes silent provisioning failures.
// Verified via examples/all-resources/ create.log (2026-05-28, commit f548a4f alignment).
const (
	// KaaSPath is the base path for KaaS operations
	KaaSPath = "/projects/%s/providers/Aruba.Container/kaas"

	// KaaSItemPath is the path for a specific KaaS cluster
	KaaSItemPath = "/projects/%s/providers/Aruba.Container/kaas/%s"

	// KaaSKubeconfigPath is the path for downloading KaaS kubeconfig
	KaaSKubeconfigPath = "/projects/%s/providers/Aruba.Container/kaas/%s/download"

	// ContainerRegistryPath is the base path for container registry operations
	ContainerRegistryPath = "/projects/%s/providers/Aruba.Container/registries"

	// ContainerRegistryItemPath is the path for a specific container registry
	ContainerRegistryItemPath = "/projects/%s/providers/Aruba.Container/registries/%s"
)

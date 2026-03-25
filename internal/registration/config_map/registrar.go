//revive:disable:package-comments
package configmap

// TODO Implement registration.Registrar backed by Kubernetes ConfigMaps.
// Reads watch targets from ConfigMaps in tenant namespaces.
// GetByID maps the registration ID to a tenant namespace and
// reads the ConfigMap.

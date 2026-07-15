package compute

// Path constants follow the server-canonical lowerCamelCase rule:
//   - Compound collections use lowerCamelCase: cloudServers, keyPairs.
//
// Do not flatten these to all-lowercase. Downstream provisioners store and re-emit
// request URIs verbatim, so a casing change causes silent provisioning failures.
// Verified via examples/all-resources/ create.log (2026-05-28, commit f548a4f alignment).
const (
	// CloudServer paths
	CloudServersPath                       = "/projects/%s/providers/Aruba.Compute/cloudServers"
	CloudServerPath                        = "/projects/%s/providers/Aruba.Compute/cloudServers/%s"
	CloudServerPowerOnPath                 = "/projects/%s/providers/Aruba.Compute/cloudServers/%s/poweron"
	CloudServerPowerOffPath                = "/projects/%s/providers/Aruba.Compute/cloudServers/%s/poweroff"
	CloudServerPasswordPath                = "/projects/%s/providers/Aruba.Compute/cloudServers/%s/password"
	CloudServerAssociateSubnetsPath        = "/projects/%s/providers/Aruba.Compute/cloudServers/%s/associateDisassociateSubnets"
	CloudServerAssociateSecurityGroupsPath = "/projects/%s/providers/Aruba.Compute/cloudServers/%s/associateDisassociateSecurityGroups"
	CloudServerAssociateElasticIPsPath     = "/projects/%s/providers/Aruba.Compute/cloudServers/%s/associateDisassociateElasticIPs"
	CloudServerAttachDetachDataVolumesPath = "/projects/%s/providers/Aruba.Compute/cloudServers/%s/attachDetachDataVolumes"

	// KeyPair paths
	KeyPairsPath = "/projects/%s/providers/Aruba.Compute/keyPairs"
	KeyPairPath  = "/projects/%s/providers/Aruba.Compute/keyPairs/%s"
)

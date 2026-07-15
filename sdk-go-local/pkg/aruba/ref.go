package aruba

import "strings"

// URI returns an opaque Ref backed by a raw URI string. Use this when you have
// a URI but not a typed wrapper, for example a resource URI loaded from a config
// file or an environment variable.
//
//	vpc, err := client.FromNetwork().VPCs().Get(ctx, aruba.URI("/projects/p/network/vpcs/v"))
func URI(s string) Ref {
	return uriRef{uri: s}
}

// Ref is a cross-resource reference. Every typed wrapper satisfies Ref;
// the aruba.URI(string) factory produces an opaque ref backed by a raw URI.
type Ref interface {
	// URI returns the resource's absolute URI path (e.g. "/projects/p/network/vpcs/v").
	URI() string
	// ID returns the resource's ID segment, or "" for opaque URI-only refs.
	ID() string
}

// uriRef is an opaque Ref backed by a raw URI string.
type uriRef struct{ uri string }

func (r uriRef) URI() string { return r.uri }
func (r uriRef) ID() string  { return "" }

// namespaceSegments are URI path segments that are category prefixes, not resource-type/id pairs.
var namespaceSegments = map[string]bool{
	"network":   true,
	"database":  true,
	"storage":   true,
	"container": true,
	"security":  true,
	"compute":   true,
	"schedule":  true,
	"metrics":   true,
	"audit":     true,
}

// parseURIIDs splits a URI path into resource-type → id pairs, skipping namespace prefixes.
//
// Example: "/projects/p/providers/Aruba.Network/vpcs/v/securityGroups/s"
// → {"projects":"p","providers":"Aruba.Network","vpcs":"v","securityGroups":"s"}
func parseURIIDs(uri string) map[string]string {
	result := make(map[string]string)
	parts := strings.Split(strings.TrimPrefix(uri, "/"), "/")
	i := 0
	for i < len(parts) {
		seg := parts[i]
		if seg == "" {
			i++
			continue
		}
		if namespaceSegments[seg] {
			i++
			continue
		}
		if i+1 < len(parts) && parts[i+1] != "" {
			result[seg] = parts[i+1]
			i += 2
		} else {
			i++
		}
	}
	return result
}

// extractID tries to read an ID from parent first via typed interface assertion, then via URI parsing.
// Returns the id and true on success; ("", false) when not found.
func extractID(parent Ref, typedKey func(Ref) (string, bool), uriSegment string) (string, bool) {
	if id, ok := typedKey(parent); ok && id != "" {
		return id, true
	}
	m := parseURIIDs(parent.URI())
	id := m[uriSegment]
	return id, id != ""
}

// Internal interface assertions used by scoped mixins to read ancestor IDs from typed parents.
// These are not exported; external packages interact only via the public Ref interface.

type withProjectID interface{ ProjectID() string }
type withVPCID interface{ VPCID() string }
type withSecurityGroupID interface{ SecurityGroupID() string }
type withSecurityRuleID interface{ SecurityRuleID() string }
type withDBaaSID interface{ DBaaSID() string }
type withDatabaseID interface{ DatabaseID() string }
type withVPCPeeringID interface{ VPCPeeringID() string }
type withVPCPeeringRouteID interface{ VPCPeeringRouteID() string }
type withVPNTunnelID interface{ VPNTunnelID() string }
type withVPNRouteID interface{ VPNRouteID() string }
type withBackupID interface{ BackupID() string }
type withDBaaSBackupID interface{ DBaaSBackupID() string }
type withKMSID interface{ KMSID() string }
type withSubnetID interface{ SubnetID() string }
type withElasticIPID interface{ ElasticIPID() string }
type withBlockStorageID interface{ BlockStorageID() string }
type withSnapshotID interface{ SnapshotID() string }
type withLoadBalancerID interface{ LoadBalancerID() string }
type withRestoreID interface{ RestoreID() string }
type withKeyPairID interface{ KeyPairID() string }
type withCloudServerID interface{ CloudServerID() string }
type withKaaSID interface{ KaaSID() string }
type withJobID interface{ JobID() string }
type withKeyID interface{ KeyID() string }
type withKmipID interface{ KmipID() string }

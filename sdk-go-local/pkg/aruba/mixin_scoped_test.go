package aruba

import "testing"

// --------------------------------------------------------------------------
// projectScopedMixin
// --------------------------------------------------------------------------

// typedProjectParent implements withProjectID + Ref.
type typedProjectParent struct{ id string }

func (p typedProjectParent) ProjectID() string { return p.id }
func (p typedProjectParent) URI() string       { return "/projects/" + p.id }
func (p typedProjectParent) ID() string        { return p.id }

func TestProjectScopedMixin_TypedRef(t *testing.T) {
	errs := &errMixin{}
	m := bindProjectScoped(errs)
	m.intoProject(typedProjectParent{id: "proj-1"})

	if m.ProjectID() != "proj-1" {
		t.Errorf("ProjectID() = %q", m.ProjectID())
	}
	if errs.Err() != nil {
		t.Errorf("unexpected error: %v", errs.Err())
	}
}

func TestProjectScopedMixin_URIFallback(t *testing.T) {
	errs := &errMixin{}
	m := bindProjectScoped(errs)
	m.intoProject(URI("/projects/proj-2"))

	if m.ProjectID() != "proj-2" {
		t.Errorf("ProjectID() via URI = %q", m.ProjectID())
	}
	if errs.Err() != nil {
		t.Errorf("unexpected error: %v", errs.Err())
	}
}

func TestProjectScopedMixin_MissingSegment(t *testing.T) {
	errs := &errMixin{}
	m := bindProjectScoped(errs)
	m.intoProject(URI("/network/vpcs/v")) // no "projects" segment

	if errs.Err() == nil {
		t.Error("expected error for missing project segment, got nil")
	}
}

// --------------------------------------------------------------------------
// vpcScopedMixin
// --------------------------------------------------------------------------

type typedVPCParent struct{ vpcID, projectID string }

func (p typedVPCParent) VPCID() string     { return p.vpcID }
func (p typedVPCParent) ProjectID() string { return p.projectID }
func (p typedVPCParent) URI() string {
	return "/projects/" + p.projectID + "/network/vpcs/" + p.vpcID
}
func (p typedVPCParent) ID() string { return p.vpcID }

func TestVPCScopedMixin_TypedRef(t *testing.T) {
	errs := &errMixin{}
	m := bindVPCScoped(errs)
	m.intoVPC(typedVPCParent{vpcID: "vpc-1", projectID: "proj-1"})

	if m.VPCID() != "vpc-1" {
		t.Errorf("VPCID() = %q", m.VPCID())
	}
	if m.ProjectID() != "proj-1" {
		t.Errorf("ProjectID() = %q", m.ProjectID())
	}
	if errs.Err() != nil {
		t.Errorf("unexpected error: %v", errs.Err())
	}
}

func TestVPCScopedMixin_URIFallback(t *testing.T) {
	errs := &errMixin{}
	m := bindVPCScoped(errs)
	m.intoVPC(URI("/projects/proj-2/network/vpcs/vpc-2"))

	if m.VPCID() != "vpc-2" {
		t.Errorf("VPCID() via URI = %q", m.VPCID())
	}
	if m.ProjectID() != "proj-2" {
		t.Errorf("ProjectID() via URI = %q", m.ProjectID())
	}
}

func TestVPCScopedMixin_MissingVPCSegment(t *testing.T) {
	errs := &errMixin{}
	m := bindVPCScoped(errs)
	m.intoVPC(URI("/projects/proj-1/network")) // no vpcs segment

	if errs.Err() == nil {
		t.Error("expected error for missing VPC segment")
	}
}

// --------------------------------------------------------------------------
// securityGroupScopedMixin
// --------------------------------------------------------------------------

type typedSGParent struct{ sgID, vpcID, projectID string }

func (p typedSGParent) SecurityGroupID() string { return p.sgID }
func (p typedSGParent) VPCID() string           { return p.vpcID }
func (p typedSGParent) ProjectID() string       { return p.projectID }
func (p typedSGParent) URI() string {
	return "/projects/" + p.projectID + "/providers/Aruba.Network/vpcs/" + p.vpcID + "/securityGroups/" + p.sgID
}
func (p typedSGParent) ID() string { return p.sgID }

func TestSecurityGroupScopedMixin_TypedRef(t *testing.T) {
	errs := &errMixin{}
	m := bindSecurityGroupScoped(errs)
	m.intoSecurityGroup(typedSGParent{sgID: "sg-1", vpcID: "vpc-1", projectID: "proj-1"})

	if m.SecurityGroupID() != "sg-1" {
		t.Errorf("SecurityGroupID() = %q", m.SecurityGroupID())
	}
	if m.VPCID() != "vpc-1" {
		t.Errorf("VPCID() = %q", m.VPCID())
	}
	if m.ProjectID() != "proj-1" {
		t.Errorf("ProjectID() = %q", m.ProjectID())
	}
	if errs.Err() != nil {
		t.Errorf("unexpected error: %v", errs.Err())
	}
}

func TestSecurityGroupScopedMixin_URIFallback(t *testing.T) {
	errs := &errMixin{}
	m := bindSecurityGroupScoped(errs)
	m.intoSecurityGroup(URI("/projects/proj-2/providers/Aruba.Network/vpcs/vpc-2/securityGroups/sg-2"))

	if m.SecurityGroupID() != "sg-2" || m.VPCID() != "vpc-2" || m.ProjectID() != "proj-2" {
		t.Errorf("got sg=%q vpc=%q proj=%q", m.SecurityGroupID(), m.VPCID(), m.ProjectID())
	}
}

// --------------------------------------------------------------------------
// dbaasScopedMixin
// --------------------------------------------------------------------------

func TestDBaaSScopedMixin_URIFallback(t *testing.T) {
	errs := &errMixin{}
	m := bindDBaaSScoped(errs)
	m.intoDBaaS(URI("/projects/proj-1/database/dbaas/db-1"))

	if m.DBaaSID() != "db-1" || m.ProjectID() != "proj-1" {
		t.Errorf("got dbaas=%q proj=%q", m.DBaaSID(), m.ProjectID())
	}
	if errs.Err() != nil {
		t.Errorf("unexpected error: %v", errs.Err())
	}
}

// --------------------------------------------------------------------------
// databaseScopedMixin
// --------------------------------------------------------------------------

func TestDatabaseScopedMixin_URIFallback(t *testing.T) {
	errs := &errMixin{}
	m := bindDatabaseScoped(errs)
	m.intoDatabase(URI("/projects/proj-1/database/dbaas/db-1/databases/mydb"))

	if m.DatabaseID() != "mydb" || m.DBaaSID() != "db-1" || m.ProjectID() != "proj-1" {
		t.Errorf("got db=%q dbaas=%q proj=%q", m.DatabaseID(), m.DBaaSID(), m.ProjectID())
	}
}

// --------------------------------------------------------------------------
// backupScopedMixin
// --------------------------------------------------------------------------

func TestBackupScopedMixin_URIFallback(t *testing.T) {
	errs := &errMixin{}
	m := bindBackupScoped(errs)
	m.intoBackup(URI("/projects/proj-1/storage/backups/bk-1"))

	if m.BackupID() != "bk-1" || m.ProjectID() != "proj-1" {
		t.Errorf("got backup=%q proj=%q", m.BackupID(), m.ProjectID())
	}
}

// --------------------------------------------------------------------------
// kmsScopedMixin
// --------------------------------------------------------------------------

func TestKMSScopedMixin_URIFallback(t *testing.T) {
	errs := &errMixin{}
	m := bindKMSScoped(errs)
	m.intoKMS(URI("/projects/proj-1/security/kms/kms-1"))

	if m.KMSID() != "kms-1" || m.ProjectID() != "proj-1" {
		t.Errorf("got kms=%q proj=%q", m.KMSID(), m.ProjectID())
	}
}

// --------------------------------------------------------------------------
// vpnTunnelScopedMixin
// --------------------------------------------------------------------------

func TestVPNTunnelScopedMixin_URIFallback(t *testing.T) {
	errs := &errMixin{}
	m := bindVPNTunnelScoped(errs)
	m.intoVPNTunnel(URI("/projects/proj-1/providers/Aruba.Network/vpnTunnels/t-1"))

	if m.VPNTunnelID() != "t-1" || m.ProjectID() != "proj-1" {
		t.Errorf("got tunnel=%q proj=%q", m.VPNTunnelID(), m.ProjectID())
	}
}

func TestVPNTunnelScopedMixin_URIFallback_CamelCase(t *testing.T) {
	errs := &errMixin{}
	m := bindVPNTunnelScoped(errs)
	m.intoVPNTunnel(URI("/projects/proj-1/providers/Aruba.Network/vpnTunnels/t-1"))

	if errs.Err() != nil {
		t.Fatalf("unexpected error: %v", errs.Err())
	}
	if m.VPNTunnelID() != "t-1" || m.ProjectID() != "proj-1" {
		t.Errorf("got tunnel=%q proj=%q", m.VPNTunnelID(), m.ProjectID())
	}
}

// --------------------------------------------------------------------------
// vpcPeeringScopedMixin
// --------------------------------------------------------------------------

func TestVPCPeeringScopedMixin_URIFallback(t *testing.T) {
	errs := &errMixin{}
	m := bindVPCPeeringScoped(errs)
	m.intoVPCPeering(URI("/projects/proj-1/providers/Aruba.Network/vpcs/vpc-1/vpcPeerings/peer-1"))

	if m.VPCPeeringID() != "peer-1" || m.VPCID() != "vpc-1" || m.ProjectID() != "proj-1" {
		t.Errorf("got peering=%q vpc=%q proj=%q", m.VPCPeeringID(), m.VPCID(), m.ProjectID())
	}
}

package aruba

import "fmt"

// --------------------------------------------------------------------------
// Scoped mixins — parent hierarchy
// --------------------------------------------------------------------------

// projectScopedMixin — direct child of a Project.
type projectScopedMixin struct {
	projectID string
	errSink   *errMixin
}

func bindProjectScoped(errSink *errMixin) projectScopedMixin {
	return projectScopedMixin{errSink: errSink}
}

func (m *projectScopedMixin) intoProject(parent Ref) {
	id, ok := extractID(parent, func(r Ref) (string, bool) {
		if p, ok := r.(withProjectID); ok {
			return p.ProjectID(), true
		}
		return "", false
	}, "projects")
	if !ok {
		m.errSink.addErr(fmt.Errorf("IntoProject: cannot determine project ID from Ref %q", parent.URI()))
		return
	}
	m.projectID = id
}

// ProjectID returns the parent project's ID.
func (m *projectScopedMixin) ProjectID() string { return m.projectID }

// --------------------------------------------------------------------------

// vpcScopedMixin — direct child of a VPC; inherits projectID from its VPC parent.
type vpcScopedMixin struct {
	vpcID     string
	projectID string
	errSink   *errMixin
}

func bindVPCScoped(errSink *errMixin) vpcScopedMixin {
	return vpcScopedMixin{errSink: errSink}
}

func (m *vpcScopedMixin) intoVPC(parent Ref) {
	vpcID, ok := extractID(parent, func(r Ref) (string, bool) {
		if p, ok := r.(withVPCID); ok {
			return p.VPCID(), true
		}
		return "", false
	}, "vpcs")
	if !ok {
		m.errSink.addErr(fmt.Errorf("IntoVPC: cannot determine VPC ID from Ref %q", parent.URI()))
	}

	projectID, ok := extractID(parent, func(r Ref) (string, bool) {
		if p, ok := r.(withProjectID); ok {
			return p.ProjectID(), true
		}
		return "", false
	}, "projects")
	if !ok {
		m.errSink.addErr(fmt.Errorf("IntoVPC: cannot determine project ID from Ref %q", parent.URI()))
	}

	m.vpcID = vpcID
	m.projectID = projectID
}

// VPCID returns the parent VPC's ID.
func (m *vpcScopedMixin) VPCID() string { return m.vpcID }

// ProjectID returns the inherited project ID.
func (m *vpcScopedMixin) ProjectID() string { return m.projectID }

// --------------------------------------------------------------------------

// securityGroupScopedMixin — direct child of a SecurityGroup; inherits vpcID and projectID.
type securityGroupScopedMixin struct {
	securityGroupID string
	vpcID           string
	projectID       string
	errSink         *errMixin
}

func bindSecurityGroupScoped(errSink *errMixin) securityGroupScopedMixin {
	return securityGroupScopedMixin{errSink: errSink}
}

func (m *securityGroupScopedMixin) intoSecurityGroup(parent Ref) {
	sgID, ok := extractID(parent, func(r Ref) (string, bool) {
		if p, ok := r.(withSecurityGroupID); ok {
			return p.SecurityGroupID(), true
		}
		return "", false
	}, "securityGroups")
	if !ok {
		m.errSink.addErr(fmt.Errorf("IntoSecurityGroup: cannot determine security group ID from Ref %q", parent.URI()))
	}

	vpcID, ok := extractID(parent, func(r Ref) (string, bool) {
		if p, ok := r.(withVPCID); ok {
			return p.VPCID(), true
		}
		return "", false
	}, "vpcs")
	if !ok {
		m.errSink.addErr(fmt.Errorf("IntoSecurityGroup: cannot determine VPC ID from Ref %q", parent.URI()))
	}

	projectID, ok := extractID(parent, func(r Ref) (string, bool) {
		if p, ok := r.(withProjectID); ok {
			return p.ProjectID(), true
		}
		return "", false
	}, "projects")
	if !ok {
		m.errSink.addErr(fmt.Errorf("IntoSecurityGroup: cannot determine project ID from Ref %q", parent.URI()))
	}

	m.securityGroupID = sgID
	m.vpcID = vpcID
	m.projectID = projectID
}

// SecurityGroupID returns the parent security group's ID.
func (m *securityGroupScopedMixin) SecurityGroupID() string { return m.securityGroupID }

// VPCID returns the inherited VPC ID.
func (m *securityGroupScopedMixin) VPCID() string { return m.vpcID }

// ProjectID returns the inherited project ID.
func (m *securityGroupScopedMixin) ProjectID() string { return m.projectID }

// --------------------------------------------------------------------------

// dbaasScopedMixin — direct child of a DBaaS instance; inherits projectID.
type dbaasScopedMixin struct {
	dbaasID   string
	projectID string
	errSink   *errMixin
}

func bindDBaaSScoped(errSink *errMixin) dbaasScopedMixin {
	return dbaasScopedMixin{errSink: errSink}
}

func (m *dbaasScopedMixin) intoDBaaS(parent Ref) {
	dbaasID, ok := extractID(parent, func(r Ref) (string, bool) {
		if p, ok := r.(withDBaaSID); ok {
			return p.DBaaSID(), true
		}
		return "", false
	}, "dbaas")
	if !ok {
		m.errSink.addErr(fmt.Errorf("IntoDBaaS: cannot determine DBaaS ID from Ref %q", parent.URI()))
	}

	projectID, ok := extractID(parent, func(r Ref) (string, bool) {
		if p, ok := r.(withProjectID); ok {
			return p.ProjectID(), true
		}
		return "", false
	}, "projects")
	if !ok {
		m.errSink.addErr(fmt.Errorf("IntoDBaaS: cannot determine project ID from Ref %q", parent.URI()))
	}

	m.dbaasID = dbaasID
	m.projectID = projectID
}

// DBaaSID returns the parent DBaaS instance's ID.
func (m *dbaasScopedMixin) DBaaSID() string { return m.dbaasID }

// ProjectID returns the inherited project ID.
func (m *dbaasScopedMixin) ProjectID() string { return m.projectID }

// --------------------------------------------------------------------------

// databaseScopedMixin — direct child of a Database; inherits dbaasID and projectID.
type databaseScopedMixin struct {
	databaseID string
	dbaasID    string
	projectID  string
	errSink    *errMixin
}

func bindDatabaseScoped(errSink *errMixin) databaseScopedMixin {
	return databaseScopedMixin{errSink: errSink}
}

func (m *databaseScopedMixin) intoDatabase(parent Ref) {
	dbID, ok := extractID(parent, func(r Ref) (string, bool) {
		if p, ok := r.(withDatabaseID); ok {
			return p.DatabaseID(), true
		}
		return "", false
	}, "databases")
	if !ok {
		m.errSink.addErr(fmt.Errorf("IntoDatabase: cannot determine database ID from Ref %q", parent.URI()))
	}

	dbaasID, ok := extractID(parent, func(r Ref) (string, bool) {
		if p, ok := r.(withDBaaSID); ok {
			return p.DBaaSID(), true
		}
		return "", false
	}, "dbaas")
	if !ok {
		m.errSink.addErr(fmt.Errorf("IntoDatabase: cannot determine DBaaS ID from Ref %q", parent.URI()))
	}

	projectID, ok := extractID(parent, func(r Ref) (string, bool) {
		if p, ok := r.(withProjectID); ok {
			return p.ProjectID(), true
		}
		return "", false
	}, "projects")
	if !ok {
		m.errSink.addErr(fmt.Errorf("IntoDatabase: cannot determine project ID from Ref %q", parent.URI()))
	}

	m.databaseID = dbID
	m.dbaasID = dbaasID
	m.projectID = projectID
}

// DatabaseID returns the parent database's ID.
func (m *databaseScopedMixin) DatabaseID() string { return m.databaseID }

// DBaaSID returns the inherited DBaaS ID.
func (m *databaseScopedMixin) DBaaSID() string { return m.dbaasID }

// ProjectID returns the inherited project ID.
func (m *databaseScopedMixin) ProjectID() string { return m.projectID }

// --------------------------------------------------------------------------

// backupScopedMixin — direct child of a StorageBackup; inherits projectID.
type backupScopedMixin struct {
	backupID  string
	projectID string
	errSink   *errMixin
}

func bindBackupScoped(errSink *errMixin) backupScopedMixin {
	return backupScopedMixin{errSink: errSink}
}

func (m *backupScopedMixin) intoBackup(parent Ref) {
	backupID, ok := extractID(parent, func(r Ref) (string, bool) {
		if p, ok := r.(withBackupID); ok {
			return p.BackupID(), true
		}
		return "", false
	}, "backups")
	if !ok {
		m.errSink.addErr(fmt.Errorf("IntoBackup: cannot determine backup ID from Ref %q", parent.URI()))
	}

	projectID, ok := extractID(parent, func(r Ref) (string, bool) {
		if p, ok := r.(withProjectID); ok {
			return p.ProjectID(), true
		}
		return "", false
	}, "projects")
	if !ok {
		m.errSink.addErr(fmt.Errorf("IntoBackup: cannot determine project ID from Ref %q", parent.URI()))
	}

	m.backupID = backupID
	m.projectID = projectID
}

// BackupID returns the parent backup's ID.
func (m *backupScopedMixin) BackupID() string { return m.backupID }

// ProjectID returns the inherited project ID.
func (m *backupScopedMixin) ProjectID() string { return m.projectID }

// --------------------------------------------------------------------------

// kmsScopedMixin — direct child of a KMS instance; inherits projectID.
type kmsScopedMixin struct {
	kmsID     string
	projectID string
	errSink   *errMixin
}

func bindKMSScoped(errSink *errMixin) kmsScopedMixin {
	return kmsScopedMixin{errSink: errSink}
}

func (m *kmsScopedMixin) intoKMS(parent Ref) {
	kmsID, ok := extractID(parent, func(r Ref) (string, bool) {
		if p, ok := r.(withKMSID); ok {
			return p.KMSID(), true
		}
		return "", false
	}, "kms")
	if !ok {
		m.errSink.addErr(fmt.Errorf("IntoKMS: cannot determine KMS ID from Ref %q", parent.URI()))
	}

	projectID, ok := extractID(parent, func(r Ref) (string, bool) {
		if p, ok := r.(withProjectID); ok {
			return p.ProjectID(), true
		}
		return "", false
	}, "projects")
	if !ok {
		m.errSink.addErr(fmt.Errorf("IntoKMS: cannot determine project ID from Ref %q", parent.URI()))
	}

	m.kmsID = kmsID
	m.projectID = projectID
}

// KMSID returns the parent KMS instance's ID.
func (m *kmsScopedMixin) KMSID() string { return m.kmsID }

// ProjectID returns the inherited project ID.
func (m *kmsScopedMixin) ProjectID() string { return m.projectID }

// --------------------------------------------------------------------------

// vpnTunnelScopedMixin — direct child of a VPN tunnel; inherits projectID.
type vpnTunnelScopedMixin struct {
	vpnTunnelID string
	projectID   string
	errSink     *errMixin
}

func bindVPNTunnelScoped(errSink *errMixin) vpnTunnelScopedMixin {
	return vpnTunnelScopedMixin{errSink: errSink}
}

func (m *vpnTunnelScopedMixin) intoVPNTunnel(parent Ref) {
	tunnelID, ok := extractID(parent, func(r Ref) (string, bool) {
		if p, ok := r.(withVPNTunnelID); ok {
			return p.VPNTunnelID(), true
		}
		return "", false
	}, "vpnTunnels")
	if !ok {
		m.errSink.addErr(fmt.Errorf("IntoVPNTunnel: cannot determine VPN tunnel ID from Ref %q", parent.URI()))
	}

	projectID, ok := extractID(parent, func(r Ref) (string, bool) {
		if p, ok := r.(withProjectID); ok {
			return p.ProjectID(), true
		}
		return "", false
	}, "projects")
	if !ok {
		m.errSink.addErr(fmt.Errorf("IntoVPNTunnel: cannot determine project ID from Ref %q", parent.URI()))
	}

	m.vpnTunnelID = tunnelID
	m.projectID = projectID
}

// VPNTunnelID returns the parent VPN tunnel's ID.
func (m *vpnTunnelScopedMixin) VPNTunnelID() string { return m.vpnTunnelID }

// ProjectID returns the inherited project ID.
func (m *vpnTunnelScopedMixin) ProjectID() string { return m.projectID }

// --------------------------------------------------------------------------

// vpcPeeringScopedMixin — direct child of a VPC peering; inherits vpcID and projectID.
type vpcPeeringScopedMixin struct {
	vpcPeeringID string
	vpcID        string
	projectID    string
	errSink      *errMixin
}

func bindVPCPeeringScoped(errSink *errMixin) vpcPeeringScopedMixin {
	return vpcPeeringScopedMixin{errSink: errSink}
}

func (m *vpcPeeringScopedMixin) intoVPCPeering(parent Ref) {
	peeringID, ok := extractID(parent, func(r Ref) (string, bool) {
		if p, ok := r.(withVPCPeeringID); ok {
			return p.VPCPeeringID(), true
		}
		return "", false
	}, "vpcPeerings")
	if !ok {
		m.errSink.addErr(fmt.Errorf("IntoVPCPeering: cannot determine VPC peering ID from Ref %q", parent.URI()))
	}

	vpcID, ok := extractID(parent, func(r Ref) (string, bool) {
		if p, ok := r.(withVPCID); ok {
			return p.VPCID(), true
		}
		return "", false
	}, "vpcs")
	if !ok {
		m.errSink.addErr(fmt.Errorf("IntoVPCPeering: cannot determine VPC ID from Ref %q", parent.URI()))
	}

	projectID, ok := extractID(parent, func(r Ref) (string, bool) {
		if p, ok := r.(withProjectID); ok {
			return p.ProjectID(), true
		}
		return "", false
	}, "projects")
	if !ok {
		m.errSink.addErr(fmt.Errorf("IntoVPCPeering: cannot determine project ID from Ref %q", parent.URI()))
	}

	m.vpcPeeringID = peeringID
	m.vpcID = vpcID
	m.projectID = projectID
}

// VPCPeeringID returns the parent VPC peering's ID.
func (m *vpcPeeringScopedMixin) VPCPeeringID() string { return m.vpcPeeringID }

// VPCID returns the inherited VPC ID.
func (m *vpcPeeringScopedMixin) VPCID() string { return m.vpcID }

// ProjectID returns the inherited project ID.
func (m *vpcPeeringScopedMixin) ProjectID() string { return m.projectID }

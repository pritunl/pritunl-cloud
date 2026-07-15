package types

import (
	"fmt"
	"strings"
)

func (c *Config) Print() {
	fmt.Println("=== CONFIG ===")
	fmt.Printf("Spec: %s\n", c.Spec.Hex())
	fmt.Printf("SpecData: %s\n", truncateString(c.SpecData, 100))
	fmt.Printf("ImdsHostSecret: %s\n", maskString(c.ImdsHostSecret))
	fmt.Printf("ClientIps: %v\n", c.ClientIps)
	fmt.Printf("Hash: %d\n", c.Hash)

	if c.Node != nil {
		fmt.Println("\nNode:")
		c.Node.Print("  ")
	} else {
		fmt.Println("\nNode: <nil>")
	}

	if c.Instance != nil {
		fmt.Println("\nInstance:")
		c.Instance.Print("  ")
	} else {
		fmt.Println("\nInstance: <nil>")
	}

	if c.Vpc != nil {
		fmt.Println("\nVpc:")
		c.Vpc.Print("  ")
	} else {
		fmt.Println("\nVpc: <nil>")
	}

	if c.Subnet != nil {
		fmt.Println("\nSubnet:")
		c.Subnet.Print("  ")
	} else {
		fmt.Println("\nSubnet: <nil>")
	}

	fmt.Printf("\nCertificates (%d):\n", len(c.Certificates))
	for i, cert := range c.Certificates {
		fmt.Printf("  [%d]:\n", i)
		cert.Print("    ")
	}

	fmt.Printf("\nSecrets (%d):\n", len(c.Secrets))
	for i, secret := range c.Secrets {
		fmt.Printf("  [%d]:\n", i)
		secret.Print("    ")
	}

	fmt.Printf("\nPods (%d):\n", len(c.Pods))
	for i, pod := range c.Pods {
		fmt.Printf("  [%d]:\n", i)
		pod.Print("    ")
	}

	fmt.Printf("\nJournals (%d):\n", len(c.Journals))
	for i, jrnl := range c.Journals {
		fmt.Printf("  [%d]:\n", i)
		jrnl.Print("    ")
	}

	fmt.Printf("\nDomains (%d):\n", len(c.Domains))
	for i, domn := range c.Domains {
		fmt.Printf("  [%d]:\n", i)
		domn.Print("    ")
	}

	fmt.Printf("\nDnsServers: %v\n", c.DnsServers)
	fmt.Printf("DnsServers6: %v\n", c.DnsServers6)
}

func (c *Certificate) Print(indent string) {
	fmt.Printf("%sId: %s\n", indent, c.Id.Hex())
	fmt.Printf("%sName: %s\n", indent, c.Name)
	fmt.Printf("%sType: %s\n", indent, c.Type)
	fmt.Printf("%sKey: %s\n", indent, maskString(c.Key))
	fmt.Printf("%sCertificate: %s\n", indent, truncateString(c.Certificate, 100))
}

func (s *Secret) Print(indent string) {
	fmt.Printf("%sId: %s\n", indent, s.Id.Hex())
	fmt.Printf("%sName: %s\n", indent, s.Name)
	fmt.Printf("%sType: %s\n", indent, s.Type)
	fmt.Printf("%sKey: %s\n", indent, maskString(s.Key))
	fmt.Printf("%sValue: %s\n", indent, maskString(s.Value))
	fmt.Printf("%sData: %s\n", indent, truncateString(s.Data, 50))
	fmt.Printf("%sRegion: %s\n", indent, s.Region)
	fmt.Printf("%sPublicKey: %s\n", indent, truncateString(s.PublicKey, 50))
	fmt.Printf("%sPrivateKey: %s\n", indent, maskString(s.PrivateKey))
}

func (p *Pod) Print(indent string) {
	fmt.Printf("%sId: %s\n", indent, p.Id.Hex())
	fmt.Printf("%sName: %s\n", indent, p.Name)
	fmt.Printf("%sUnits (%d):\n", indent, len(p.Units))
	for i, unit := range p.Units {
		fmt.Printf("%s  [%d]:\n", indent, i)
		unit.Print(indent + "    ")
	}
}

func (u *Unit) Print(indent string) {
	fmt.Printf("%sId: %s\n", indent, u.Id.Hex())
	fmt.Printf("%sName: %s\n", indent, u.Name)
	fmt.Printf("%sKind: %s\n", indent, u.Kind)
	fmt.Printf("%sCount: %d\n", indent, u.Count)

	fmt.Printf("%sPublicIps: %v\n", indent, u.PublicIps)
	fmt.Printf("%sPublicIps6: %v\n", indent, u.PublicIps6)
	fmt.Printf("%sHealthyPublicIps: %v\n", indent, u.HealthyPublicIps)
	fmt.Printf("%sHealthyPublicIps6: %v\n", indent, u.HealthyPublicIps6)
	fmt.Printf("%sUnhealthyPublicIps: %v\n", indent, u.UnhealthyPublicIps)
	fmt.Printf("%sUnhealthyPublicIps6: %v\n", indent, u.UnhealthyPublicIps6)

	fmt.Printf("%sPrivateIps: %v\n", indent, u.PrivateIps)
	fmt.Printf("%sPrivateIps6: %v\n", indent, u.PrivateIps6)
	fmt.Printf("%sHealthyPrivateIps: %v\n", indent, u.HealthyPrivateIps)
	fmt.Printf("%sHealthyPrivateIps6: %v\n", indent, u.HealthyPrivateIps6)
	fmt.Printf("%sUnhealthyPrivateIps: %v\n", indent, u.UnhealthyPrivateIps)
	fmt.Printf("%sUnhealthyPrivateIps6: %v\n", indent, u.UnhealthyPrivateIps6)

	fmt.Printf("%sCloudPublicIps: %v\n", indent, u.CloudPublicIps)
	fmt.Printf("%sCloudPublicIps6: %v\n", indent, u.CloudPublicIps6)
	fmt.Printf("%sCloudPrivateIps: %v\n", indent, u.CloudPrivateIps)
	fmt.Printf("%sHealthyCloudPublicIps: %v\n", indent, u.HealthyCloudPublicIps)
	fmt.Printf("%sHealthyCloudPublicIps6: %v\n", indent, u.HealthyCloudPublicIps6)
	fmt.Printf("%sHealthyCloudPrivateIps: %v\n", indent, u.HealthyCloudPrivateIps)
	fmt.Printf("%sUnhealthyCloudPublicIps: %v\n", indent, u.UnhealthyCloudPublicIps)
	fmt.Printf("%sUnhealthyCloudPublicIps6: %v\n", indent, u.UnhealthyCloudPublicIps6)
	fmt.Printf("%sUnhealthyCloudPrivateIps: %v\n", indent, u.UnhealthyCloudPrivateIps)
}

func (j *Journal) Print(indent string) {
	fmt.Printf("%sIndex: %d\n", indent, j.Index)
	fmt.Printf("%sKey: %s\n", indent, j.Key)
	fmt.Printf("%sType: %s\n", indent, j.Type)
	fmt.Printf("%sUnit: %s\n", indent, j.Unit)
	fmt.Printf("%sPath: %s\n", indent, j.Path)
}

func (d *Domain) Print(indent string) {
	fmt.Printf("%sDomain: %s\n", indent, d.Domain)
	fmt.Printf("%sType: %s\n", indent, d.Type)
	fmt.Printf("%sIp: %s\n", indent, d.Ip)
	fmt.Printf("%sTarget: %s\n", indent, d.Target)
}

func maskString(s string) string {
	if s == "" {
		return "<empty>"
	}
	if len(s) <= 8 {
		return strings.Repeat("*", len(s))
	}
	return s[:4] + strings.Repeat("*", len(s)-8) + s[len(s)-4:]
}

func (i *Node) Print(indent string) {
	fmt.Printf("%sId: %s\n", indent, i.Id.Hex())
	fmt.Printf("%sName: %s\n", indent, i.Name)
	fmt.Printf("%sPublicIps: %v\n", indent, i.PublicIps)
	fmt.Printf("%sPublicIps6: %v\n", indent, i.PublicIps6)
}

func (i *Instance) Print(indent string) {
	fmt.Printf("%sId: %s\n", indent, i.Id.Hex())
	fmt.Printf("%sOrganization: %s\n", indent, i.Organization.Hex())
	fmt.Printf("%sZone: %s\n", indent, i.Zone.Hex())
	fmt.Printf("%sVpc: %s\n", indent, i.Vpc.Hex())
	fmt.Printf("%sSubnet: %s\n", indent, i.Subnet.Hex())
	fmt.Printf("%sCloudSubnet: %s\n", indent, i.CloudSubnet)
	fmt.Printf("%sCloudVnic: %s\n", indent, i.CloudVnic)
	fmt.Printf("%sImage: %s\n", indent, i.Image.Hex())
	fmt.Printf("%sState: %s\n", indent, i.State)
	fmt.Printf("%sTimestamp: %s\n", indent, i.Timestamp)
	fmt.Printf("%sAction: %s\n", indent, i.Action)
	fmt.Printf("%sUefi: %t\n", indent, i.Uefi)
	fmt.Printf("%sSecureBoot: %t\n", indent, i.SecureBoot)
	fmt.Printf("%sTpm: %t\n", indent, i.Tpm)
	fmt.Printf("%sDhcpServer: %t\n", indent, i.DhcpServer)
	fmt.Printf("%sCloudType: %s\n", indent, i.CloudType)
	fmt.Printf("%sSystemKind: %s\n", indent, i.SystemKind)
	fmt.Printf("%sDeleteProtection: %t\n", indent, i.DeleteProtection)
	fmt.Printf("%sSkipSourceDestCheck: %t\n", indent, i.SkipSourceDestCheck)
	fmt.Printf("%sQemuVersion: %s\n", indent, i.QemuVersion)
	fmt.Printf("%sPublicIps: %v\n", indent, i.PublicIps)
	fmt.Printf("%sPublicIps6: %v\n", indent, i.PublicIps6)
	fmt.Printf("%sPrivateIps: %v\n", indent, i.PrivateIps)
	fmt.Printf("%sPrivateIps6: %v\n", indent, i.PrivateIps6)
	fmt.Printf("%sGatewayIps: %v\n", indent, i.GatewayIps)
	fmt.Printf("%sGatewayIps6: %v\n", indent, i.GatewayIps6)
	fmt.Printf("%sCloudPrivateIps: %v\n", indent, i.CloudPrivateIps)
	fmt.Printf("%sCloudPublicIps: %v\n", indent, i.CloudPublicIps)
	fmt.Printf("%sCloudPublicIps6: %v\n", indent, i.CloudPublicIps6)
	fmt.Printf("%sHostIps: %v\n", indent, i.HostIps)
	fmt.Printf("%sNodePortIps: %v\n", indent, i.NodePortIps)
	fmt.Printf("%sNetworkNamespace: %s\n", indent, i.NetworkNamespace)
	fmt.Printf("%sNoPublicAddress: %t\n", indent, i.NoPublicAddress)
	fmt.Printf("%sNoPublicAddress6: %t\n", indent, i.NoPublicAddress6)
	fmt.Printf("%sNoHostAddress: %t\n", indent, i.NoHostAddress)
	fmt.Printf("%sNode: %s\n", indent, i.Node.Hex())
	fmt.Printf("%sShape: %s\n", indent, i.Shape.Hex())
	fmt.Printf("%sName: %s\n", indent, i.Name)
	fmt.Printf("%sRootEnabled: %t\n", indent, i.RootEnabled)
	fmt.Printf("%sMemory: %d\n", indent, i.Memory)
	fmt.Printf("%sProcessors: %d\n", indent, i.Processors)
	fmt.Printf("%sNetworkRoles: %v\n", indent, i.Roles)
	fmt.Printf("%sVnc: %t\n", indent, i.Vnc)
	fmt.Printf("%sSpice: %t\n", indent, i.Spice)
	fmt.Printf("%sGui: %t\n", indent, i.Gui)
	fmt.Printf("%sDeployment: %s\n", indent, i.Deployment.Hex())
}

func (v *Vpc) Print(indent string) {
	fmt.Printf("%sId: %s\n", indent, v.Id.Hex())
	fmt.Printf("%sName: %s\n", indent, v.Name)
	fmt.Printf("%sVpcId: %d\n", indent, v.VpcId)
	fmt.Printf("%sNetwork: %s\n", indent, v.Network)
	fmt.Printf("%sNetwork6: %s\n", indent, v.Network6)

	fmt.Printf("%sSubnets (%d):\n", indent, len(v.Subnets))
	for i, subnet := range v.Subnets {
		fmt.Printf("%s  [%d]:\n", indent, i)
		subnet.Print(indent + "    ")
	}

	fmt.Printf("%sRoutes (%d):\n", indent, len(v.Routes))
	for i, route := range v.Routes {
		fmt.Printf("%s  [%d]:\n", indent, i)
		route.Print(indent + "    ")
	}
}

func (s *Subnet) Print(indent string) {
	fmt.Printf("%sId: %s\n", indent, s.Id.Hex())
	fmt.Printf("%sName: %s\n", indent, s.Name)
	fmt.Printf("%sNetwork: %s\n", indent, s.Network)
}

func (r *Route) Print(indent string) {
	fmt.Printf("%sDestination: %s\n", indent, r.Destination)
	fmt.Printf("%sTarget: %s\n", indent, r.Target)
}

func truncateString(s string, maxLen int) string {
	if s == "" {
		return "<empty>"
	}
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "... (truncated)"
}

func (c *Config) PrintCompact() {
	fmt.Printf("Config[Spec:%s, Hash:%d, Pods:%d, Secrets:%d, Certs:%d]\n",
		c.Spec.Hex()[:8], c.Hash, len(c.Pods), len(c.Secrets), len(c.Certificates))
}

func (c *Config) PrintJSON() {
	fmt.Println("{")
	fmt.Printf("  \"spec\": \"%s\",\n", c.Spec.Hex())
	fmt.Printf("  \"spec_data\": \"%s\",\n", truncateString(c.SpecData, 50))
	fmt.Printf("  \"imds_host_secret\": \"%s\",\n", maskString(c.ImdsHostSecret))
	fmt.Printf("  \"client_ips\": %v,\n", c.ClientIps)
	fmt.Printf("  \"hash\": %d,\n", c.Hash)
	fmt.Printf("  \"certificates_count\": %d,\n", len(c.Certificates))
	fmt.Printf("  \"secrets_count\": %d,\n", len(c.Secrets))
	fmt.Printf("  \"pods_count\": %d,\n", len(c.Pods))
	fmt.Printf("  \"journals_count\": %d,\n", len(c.Journals))
	fmt.Printf("  \"domains_count\": %d,\n", len(c.Domains))
	fmt.Printf("  \"dns_servers\": %v,\n", c.DnsServers)
	fmt.Printf("  \"dns_servers6\": %v\n", c.DnsServers6)
	fmt.Println("}")
}

package node

import (
	"container/list"
	"net"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/pritunl-cloud/block"
	"github.com/pritunl/pritunl-cloud/bridges"
	"github.com/pritunl/pritunl-cloud/certificate"
	"github.com/pritunl/pritunl-cloud/constants"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/utils"
)

var (
	Self *Node
)

type Node struct {
	Id                   primitive.ObjectID         `bson:"_id" json:"id"`
	Zone                 primitive.ObjectID         `bson:"zone,omitempty" json:"zone"`
	Name                 string                     `bson:"name" json:"name"`
	Types                []string                   `bson:"types" json:"types"`
	Timestamp            time.Time                  `bson:"timestamp" json:"timestamp"`
	Port                 int                        `bson:"port" json:"port"`
	NoRedirectServer     bool                       `bson:"no_redirect_server" json:"no_redirect_server"`
	Protocol             string                     `bson:"protocol" json:"protocol"`
	Hypervisor           string                     `bson:"hypervisor" json:"hypervisor"`
	Vga                  string                     `bson:"vga" json:"vga"`
	Certificate          primitive.ObjectID         `bson:"certificate" json:"certificate"`
	Certificates         []primitive.ObjectID       `bson:"certificates" json:"certificates"`
	SelfCertificate      string                     `bson:"self_certificate_key" json:"-"`
	SelfCertificateKey   string                     `bson:"self_certificate" json:"-"`
	AdminDomain          string                     `bson:"admin_domain" json:"admin_domain"`
	UserDomain           string                     `bson:"user_domain" json:"user_domain"`
	RequestsMin          int64                      `bson:"requests_min" json:"requests_min"`
	ForwardedForHeader   string                     `bson:"forwarded_for_header" json:"forwarded_for_header"`
	ForwardedProtoHeader string                     `bson:"forwarded_proto_header" json:"forwarded_proto_header"`
	ExternalInterface    string                     `bson:"external_interface" json:"external_interface"`
	InternalInterface    string                     `bson:"internal_interface" json:"internal_interface"`
	ExternalInterfaces   []string                   `bson:"external_interfaces" json:"external_interfaces"`
	InternalInterfaces   []string                   `bson:"internal_interfaces" json:"internal_interfaces"`
	AvailableInterfaces  []string                   `bson:"available_interfaces" json:"available_interfaces"`
	AvailableBridges     []string                   `bson:"available_bridges" json:"available_bridges"`
	DefaultInterface     string                     `bson:"default_interface" json:"default_interface"`
	NetworkMode          string                     `bson:"network_mode" json:"network_mode"`
	Blocks               []*BlockAttachment         `bson:"blocks" json:"blocks"`
	HostBlock            primitive.ObjectID         `bson:"host_block,omitempty" json:"host_block"`
	HostNat              bool                       `bson:"host_nat" json:"host_nat"`
	HostNatExcludes      []string                   `bson:"host_nat_excludes" json:"host_nat_excludes"`
	JumboFrames          bool                       `bson:"jumbo_frames" json:"jumbo_frames"`
	Firewall             bool                       `bson:"firewall" json:"firewall"`
	NetworkRoles         []string                   `bson:"network_roles" json:"network_roles"`
	Memory               float64                    `bson:"memory" json:"memory"`
	Load1                float64                    `bson:"load1" json:"load1"`
	Load5                float64                    `bson:"load5" json:"load5"`
	Load15               float64                    `bson:"load15" json:"load15"`
	CpuUnits             int                        `bson:"cpu_units" json:"cpu_units"`
	MemoryUnits          float64                    `bson:"memory_units" json:"memory_units"`
	CpuUnitsRes          int                        `bson:"cpu_units_res" json:"cpu_units_res"`
	MemoryUnitsRes       float64                    `bson:"memory_units_res" json:"memory_units_res"`
	PublicIps            []string                   `bson:"public_ips" json:"public_ips"`
	PublicIps6           []string                   `bson:"public_ips6" json:"public_ips6"`
	PrivateIps           map[string]string          `bson:"private_ips" json:"private_ips"`
	SoftwareVersion      string                     `bson:"software_version" json:"software_version"`
	Hostname             string                     `bson:"hostname" json:"hostname"`
	Version              int                        `bson:"version" json:"-"`
	VirtPath             string                     `bson:"virt_path" json:"virt_path"`
	CachePath            string                     `bson:"cache_path" json:"cache_path"`
	OracleUser           string                     `bson:"oracle_user" json:"oracle_user"`
	OraclePrivateKey     string                     `bson:"oracle_private_key" json:"-"`
	OraclePublicKey      string                     `bson:"oracle_public_key" json:"oracle_public_key"`
	OracleHostRoute      bool                       `bson:"oracle_host_route" json:"oracle_host_route"`
	CertificateObjs      []*certificate.Certificate `bson:"-" json:"-"`
	reqLock              sync.Mutex                 `bson:"-" json:"-"`
	reqCount             *list.List                 `bson:"-" json:"-"`
}

func (n *Node) Copy() *Node {
	nde := &Node{
		Id:                   n.Id,
		Zone:                 n.Zone,
		Name:                 n.Name,
		Types:                n.Types,
		Timestamp:            n.Timestamp,
		Port:                 n.Port,
		NoRedirectServer:     n.NoRedirectServer,
		Protocol:             n.Protocol,
		Hypervisor:           n.Hypervisor,
		Vga:                  n.Vga,
		Certificate:          n.Certificate,
		Certificates:         n.Certificates,
		SelfCertificate:      n.SelfCertificate,
		SelfCertificateKey:   n.SelfCertificateKey,
		AdminDomain:          n.AdminDomain,
		UserDomain:           n.UserDomain,
		RequestsMin:          n.RequestsMin,
		ForwardedForHeader:   n.ForwardedForHeader,
		ForwardedProtoHeader: n.ForwardedProtoHeader,
		ExternalInterface:    n.ExternalInterface,
		InternalInterface:    n.InternalInterface,
		ExternalInterfaces:   n.ExternalInterfaces,
		InternalInterfaces:   n.InternalInterfaces,
		AvailableInterfaces:  n.AvailableInterfaces,
		AvailableBridges:     n.AvailableBridges,
		DefaultInterface:     n.DefaultInterface,
		NetworkMode:          n.NetworkMode,
		Blocks:               n.Blocks,
		HostBlock:            n.HostBlock,
		HostNat:              n.HostNat,
		HostNatExcludes:      n.HostNatExcludes,
		JumboFrames:          n.JumboFrames,
		Firewall:             n.Firewall,
		NetworkRoles:         n.NetworkRoles,
		Memory:               n.Memory,
		Load1:                n.Load1,
		Load5:                n.Load5,
		Load15:               n.Load15,
		CpuUnits:             n.CpuUnits,
		MemoryUnits:          n.MemoryUnits,
		CpuUnitsRes:          n.CpuUnitsRes,
		MemoryUnitsRes:       n.MemoryUnitsRes,
		PublicIps:            n.PublicIps,
		PublicIps6:           n.PublicIps6,
		PrivateIps:           n.PrivateIps,
		SoftwareVersion:      n.SoftwareVersion,
		Hostname:             n.Hostname,
		Version:              n.Version,
		VirtPath:             n.VirtPath,
		CachePath:            n.CachePath,
		OracleUser:           n.OracleUser,
		OraclePrivateKey:     n.OraclePrivateKey,
		OraclePublicKey:      n.OraclePublicKey,
		OracleHostRoute:      n.OracleHostRoute,
		CertificateObjs:      n.CertificateObjs,
		reqLock:              n.reqLock,
		reqCount:             n.reqCount,
	}

	return nde
}

func (n *Node) AddRequest() {
	n.reqLock.Lock()
	back := n.reqCount.Back()
	back.Value = back.Value.(int) + 1
	n.reqLock.Unlock()
}

func (n *Node) GetVirtPath() string {
	if n.VirtPath == "" {
		return DefaultRoot
	}
	return n.VirtPath
}

func (n *Node) GetCachePath() string {
	if n.CachePath == "" {
		return DefaultCache
	}
	return n.CachePath
}

func (n *Node) IsAdmin() bool {
	for _, typ := range n.Types {
		if typ == Admin {
			return true
		}
	}
	return false
}

func (n *Node) IsUser() bool {
	for _, typ := range n.Types {
		if typ == User {
			return true
		}
	}
	return false
}

func (n *Node) IsHypervisor() bool {
	for _, typ := range n.Types {
		if typ == Hypervisor {
			return true
		}
	}
	return false
}

func (n *Node) IsIpsec() bool {
	for _, typ := range n.Types {
		if typ == Ipsec {
			return true
		}
	}
	return false
}

func (n *Node) Validate(db *database.Database) (
	errData *errortypes.ErrorData, err error) {

	if n.Hypervisor == "" {
		n.Hypervisor = Kvm
	}

	if n.Protocol != "http" && n.Protocol != "https" {
		errData = &errortypes.ErrorData{
			Error:   "node_protocol_invalid",
			Message: "Invalid node server protocol",
		}
		return
	}

	if n.Port < 1 || n.Port > 65535 {
		errData = &errortypes.ErrorData{
			Error:   "node_port_invalid",
			Message: "Invalid node server port",
		}
		return
	}

	if n.Certificates == nil || n.Protocol != "https" {
		n.Certificates = []primitive.ObjectID{}
	}

	if n.Types == nil {
		n.Types = []string{}
	}

	if (n.IsAdmin() && !n.IsUser()) || (n.IsUser() && !n.IsAdmin()) {
		n.AdminDomain = ""
		n.UserDomain = ""
	} else {
		if !n.IsAdmin() {
			n.AdminDomain = ""
		}
		if !n.IsUser() {
			n.UserDomain = ""
		}
	}

	if !n.Zone.IsZero() {
		coll := db.Zones()
		count, e := coll.Count(db, &bson.M{
			"_id": n.Zone,
		})
		if e != nil {
			err = database.ParseError(e)
			return
		}

		if count == 0 {
			n.Zone = primitive.NilObjectID
		}
	}

	if n.VirtPath == "" {
		n.VirtPath = DefaultRoot
	}
	if n.CachePath == "" {
		n.CachePath = DefaultCache
	}

	if n.NetworkRoles == nil || !n.Firewall {
		n.NetworkRoles = []string{}
	}

	if n.Firewall && len(n.NetworkRoles) == 0 {
		errData = &errortypes.ErrorData{
			Error:   "firewall_empty_roles",
			Message: "Cannot enable firewall without network roles",
		}
		return
	}

	if n.Blocks == nil {
		n.Blocks = []*BlockAttachment{}
	}

	switch n.NetworkMode {
	case Static:
		n.ExternalInterfaces = []string{}
		break
	case Internal:
		n.ExternalInterfaces = []string{}
		n.Blocks = []*BlockAttachment{}
		break
	case Dhcp, "":
		n.NetworkMode = Dhcp
		n.Blocks = []*BlockAttachment{}
		break
	default:
		errData = &errortypes.ErrorData{
			Error:   "invalid_network_mode",
			Message: "Network mode invalid",
		}
		return
	}

	if n.HostNatExcludes == nil {
		n.HostNatExcludes = []string{}
	}

	if !n.HostBlock.IsZero() && n.HostNat {
		hostNatExcludes := []string{}
		for _, hostNat := range n.HostNatExcludes {
			_, hostNatNet, e := net.ParseCIDR(hostNat)
			if e != nil {
				errData = &errortypes.ErrorData{
					Error:   "invalid_host_nat_exclude",
					Message: "Host NAT network exclude is invalid",
				}
				return
			}

			hostNatExcludes = append(hostNatExcludes, hostNatNet.String())
		}
		n.HostNatExcludes = hostNatExcludes
	} else {
		n.HostNat = false
		n.HostNatExcludes = []string{}
	}

	if n.OracleHostRoute {
		if n.OracleUser == "" {
			errData = &errortypes.ErrorData{
				Error:   "missing_oracle_user",
				Message: "Oracle user OCID required for host routing",
			}
			return
		}
	} else {
		n.OracleUser = ""
	}

	n.Format()

	return
}

func (n *Node) Format() {
	sort.Strings(n.Types)
	utils.SortObjectIds(n.Certificates)
}

func (n *Node) SetActive() {
	if time.Since(n.Timestamp) > 30*time.Second {
		n.RequestsMin = 0
		n.Memory = 0
		n.Load1 = 0
		n.Load5 = 0
		n.Load15 = 0
		n.CpuUnits = 0
		n.CpuUnitsRes = 0
		n.MemoryUnits = 0
		n.MemoryUnitsRes = 0
	}
}

func (n *Node) Commit(db *database.Database) (err error) {
	coll := db.Nodes()

	err = coll.Commit(n.Id, n)
	if err != nil {
		return
	}

	return
}

func (n *Node) CommitFields(db *database.Database, fields set.Set) (
	err error) {

	coll := db.Nodes()

	err = coll.CommitFields(n.Id, n, fields)
	if err != nil {
		return
	}

	return
}

func (n *Node) GetStaticAddr(db *database.Database,
	instId primitive.ObjectID) (blck *block.Block, ip net.IP, iface string,
	err error) {

	blck, blckIp, err := block.GetInstanceIp(db, instId, block.External)
	if err != nil {
		return
	}

	if blckIp != nil {
		for _, blckAttch := range n.Blocks {
			if blckAttch.Block == blck.Id {
				ip = blckIp.GetIp()
				iface = blckAttch.Interface
				return
			}
		}

		err = block.RemoveIp(db, blckIp.Id)
		if err != nil {
			return
		}
	}

	for _, blckAttch := range n.Blocks {
		blck, err = block.Get(db, blckAttch.Block)
		if err != nil {
			return
		}

		iface = blckAttch.Interface

		ip, err = blck.GetIp(db, instId, block.External)
		if err != nil {
			if _, ok := err.(*block.BlockFull); ok {
				err = nil
				continue
			} else {
				return
			}
		}

		break
	}

	if ip == nil {
		err = &errortypes.NotFoundError{
			errors.New("node: No external block addresses available"),
		}
		return
	}

	return
}

func (n *Node) GetStaticHostAddr(db *database.Database,
	instId primitive.ObjectID) (blck *block.Block, ip net.IP, err error) {

	blck, blckIp, err := block.GetInstanceIp(db, instId, block.Host)
	if err != nil {
		return
	}

	if blckIp != nil {
		if n.HostBlock == blck.Id {
			ip = blckIp.GetIp()
			return
		}

		err = block.RemoveIp(db, blckIp.Id)
		if err != nil {
			return
		}
	}

	blck, err = block.Get(db, n.HostBlock)
	if err != nil {
		return
	}

	ip, err = blck.GetIp(db, instId, block.Host)
	if err != nil {
		if _, ok := err.(*block.BlockFull); ok {
			err = nil
		} else {
			return
		}
	}

	if ip == nil {
		err = &errortypes.NotFoundError{
			errors.New("node: No host block addresses available"),
		}
		return
	}

	return
}

func (n *Node) GetRemoteAddr(r *http.Request) (addr string) {
	if n.ForwardedForHeader != "" {
		addr = strings.TrimSpace(
			strings.SplitN(r.Header.Get(n.ForwardedForHeader), ",", 1)[0])
		if addr != "" {
			return
		}
	}

	addr = utils.StripPort(r.RemoteAddr)
	return
}

func (n *Node) update(db *database.Database) (err error) {
	coll := db.Nodes()

	nde := &Node{}
	opts := &options.FindOneAndUpdateOptions{}
	opts.SetReturnDocument(options.After)

	err = coll.FindOneAndUpdate(
		db,
		&bson.M{
			"_id": n.Id,
		},
		&bson.M{
			"$set": &bson.M{
				"timestamp":            n.Timestamp,
				"requests_min":         n.RequestsMin,
				"memory":               n.Memory,
				"load1":                n.Load1,
				"load5":                n.Load5,
				"load15":               n.Load15,
				"cpu_units":            n.CpuUnits,
				"memory_units":         n.MemoryUnits,
				"cpu_units_res":        n.CpuUnitsRes,
				"memory_units_res":     n.MemoryUnitsRes,
				"public_ips":           n.PublicIps,
				"public_ips6":          n.PublicIps6,
				"private_ips":          n.PrivateIps,
				"hostname":             n.Hostname,
				"available_interfaces": n.AvailableInterfaces,
				"available_bridges":    n.AvailableBridges,
				"default_interface":    n.DefaultInterface,
			},
		},
		opts,
	).Decode(nde)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	n.Id = nde.Id
	n.Zone = nde.Zone
	n.Name = nde.Name
	n.Types = nde.Types
	n.Port = nde.Port
	n.NoRedirectServer = nde.NoRedirectServer
	n.Protocol = nde.Protocol
	n.Hypervisor = nde.Hypervisor
	n.Certificates = nde.Certificates
	n.SelfCertificate = nde.SelfCertificate
	n.SelfCertificateKey = nde.SelfCertificateKey
	n.AdminDomain = nde.AdminDomain
	n.UserDomain = nde.UserDomain
	n.ForwardedForHeader = nde.ForwardedForHeader
	n.ForwardedProtoHeader = nde.ForwardedProtoHeader
	n.ExternalInterface = nde.ExternalInterface
	n.InternalInterface = nde.InternalInterface
	n.ExternalInterfaces = nde.ExternalInterfaces
	n.InternalInterfaces = nde.InternalInterfaces
	n.NetworkMode = nde.NetworkMode
	n.Blocks = nde.Blocks
	n.HostBlock = nde.HostBlock
	n.HostNat = nde.HostNat
	n.HostNatExcludes = nde.HostNatExcludes
	n.JumboFrames = nde.JumboFrames
	n.Firewall = nde.Firewall
	n.NetworkRoles = nde.NetworkRoles
	n.VirtPath = nde.VirtPath
	n.CachePath = nde.CachePath
	n.OracleUser = nde.OracleUser
	n.OraclePrivateKey = nde.OraclePrivateKey
	n.OraclePublicKey = nde.OraclePublicKey
	n.OracleHostRoute = nde.OracleHostRoute

	return
}

func (n *Node) loadCerts(db *database.Database) (err error) {
	certObjs := []*certificate.Certificate{}

	if n.Certificates == nil || len(n.Certificates) == 0 {
		n.CertificateObjs = certObjs
		return
	}

	for _, certId := range n.Certificates {
		cert, e := certificate.Get(db, certId)
		if e != nil {
			switch e.(type) {
			case *database.NotFoundError:
				e = nil
				break
			default:
				err = e
				return
			}
		} else {
			certObjs = append(certObjs, cert)
		}
	}

	n.CertificateObjs = certObjs

	return
}

func (n *Node) sync() {
	db := database.GetDatabase()
	defer db.Close()

	n.Timestamp = time.Now()

	mem, total, err := utils.MemoryUsed()
	if err != nil {
		n.Memory = 0

		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("node: Failed to get memory")
	} else {
		n.Memory = mem
	}

	load, err := utils.LoadAverage()
	if err != nil {
		n.CpuUnits = 0
		n.MemoryUnits = 0
		n.Load1 = 0
		n.Load5 = 0
		n.Load15 = 0

		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("node: Failed to get load")
	} else {
		n.CpuUnits = load.CpuUnits
		n.MemoryUnits = total
		n.Load1 = load.Load1
		n.Load5 = load.Load5
		n.Load15 = load.Load15
	}

	defaultIface, err := getDefaultIface()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"default_interface": defaultIface,
			"error":             err,
		}).Error("node: Failed to get public address")
	}

	if defaultIface != "" {
		n.DefaultInterface = defaultIface

		pubAddr, pubAddr6, err := bridges.GetIpAddrs(defaultIface)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"default_interface": defaultIface,
				"error":             err,
			}).Error("node: Failed to get public address")
		}

		if pubAddr != "" {
			n.PublicIps = []string{
				pubAddr,
			}
		}

		if pubAddr6 != "" {
			n.PublicIps6 = []string{
				pubAddr6,
			}
		}
	}

	privateIps := map[string]string{}
	internalInterfaces := n.InternalInterfaces
	if internalInterfaces != nil {
		for _, iface := range internalInterfaces {
			addr, _, err := bridges.GetIpAddrs(iface)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"internal_interface": iface,
					"error":              err,
				}).Error("node: Failed to get private address")
			}

			if addr != "" {
				privateIps[iface] = addr
			}
		}
	}
	n.PrivateIps = privateIps

	ifaces, err := GetInterfaces()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("node: Failed to get interfaces")
	}

	if ifaces != nil {
		n.AvailableInterfaces = ifaces
	} else {
		n.AvailableInterfaces = []string{}
	}

	brdgs, err := bridges.GetBridges()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("node: Failed to get bridge interfaces")
	}

	if brdgs != nil {
		n.AvailableBridges = brdgs
	} else {
		n.AvailableBridges = []string{}
	}

	hostname, err := os.Hostname()
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "node: Failed to get hostname"),
		}
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("node: Failed to get hostname")
	}
	n.Hostname = hostname

	err = n.update(db)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("node: Failed to update node")
	}

	err = n.loadCerts(db)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("node: Failed to load node certificate")
	}
}

func (n *Node) keepalive() {
	for {
		n.sync()
		time.Sleep(1 * time.Second)
	}
}

func (n *Node) reqInit() {
	n.reqLock.Lock()
	n.reqCount = list.New()
	for i := 0; i < 60; i++ {
		n.reqCount.PushBack(0)
	}
	n.reqLock.Unlock()
}

func (n *Node) reqSync() {
	for {
		time.Sleep(1 * time.Second)

		n.reqLock.Lock()

		var count int64
		for elm := n.reqCount.Front(); elm != nil; elm = elm.Next() {
			count += int64(elm.Value.(int))
		}
		n.RequestsMin = count

		n.reqCount.Remove(n.reqCount.Front())
		n.reqCount.PushBack(0)

		n.reqLock.Unlock()
	}
}

func (n *Node) Init() (err error) {
	db := database.GetDatabase()
	defer db.Close()

	coll := db.Nodes()

	err = coll.FindOneId(n.Id, n)
	if err != nil {
		switch err.(type) {
		case *database.NotFoundError:
			err = nil
		default:
			return
		}
	}

	n.SoftwareVersion = constants.Version

	if n.Name == "" {
		n.Name = utils.RandName()
	}

	if n.Types == nil {
		n.Types = []string{Admin, Hypervisor}
	}

	if n.Protocol == "" {
		n.Protocol = "https"
	}

	if n.Port == 0 {
		n.Port = 443
	}

	if n.Hypervisor == "" {
		n.Hypervisor = Kvm
	}

	bsonSet := bson.M{
		"_id":              n.Id,
		"name":             n.Name,
		"types":            n.Types,
		"timestamp":        time.Now(),
		"protocol":         n.Protocol,
		"port":             n.Port,
		"hypervisor":       n.Hypervisor,
		"software_version": n.SoftwareVersion,
	}

	if n.OraclePublicKey == "" || n.OraclePrivateKey == "" {
		privKey, pubKey, e := utils.GenerateRsaKey()
		if e != nil {
			err = e
			return
		}

		bsonSet["oracle_public_key"] = string(pubKey)
		bsonSet["oracle_private_key"] = string(privKey)
	}

	// Database upgrade
	if n.InternalInterfaces == nil {
		ifaces := []string{}
		iface := n.InternalInterface
		if iface != "" {
			ifaces = append(ifaces, iface)
		}
		n.InternalInterfaces = ifaces
		bsonSet["internal_interfaces"] = ifaces
	}

	// Database upgrade
	if n.ExternalInterfaces == nil {
		ifaces := []string{}
		iface := n.ExternalInterface
		if iface != "" {
			ifaces = append(ifaces, iface)
		}
		n.ExternalInterfaces = ifaces
		bsonSet["external_interfaces"] = ifaces
	}

	opts := &options.UpdateOptions{}
	opts.SetUpsert(true)

	_, err = coll.UpdateOne(
		db,
		&bson.M{
			"_id": n.Id,
		},
		&bson.M{
			"$set": bsonSet,
		},
		opts,
	)
	if err != nil {
		err = database.ParseError(err)
		return
	}

	n.reqInit()

	err = n.loadCerts(db)
	if err != nil {
		return
	}

	n.sync()

	event.PublishDispatch(db, "node.change")

	Self = n

	go n.keepalive()
	go n.reqSync()

	return
}

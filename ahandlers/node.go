package ahandlers

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/dropbox/godropbox/container/set"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/block"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/datacenter"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/drive"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/firewall"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/settings"
	"github.com/pritunl/pritunl-cloud/utils"
)

type nodeData struct {
	Id                      primitive.ObjectID      `json:"id"`
	Zone                    primitive.ObjectID      `json:"zone"`
	Name                    string                  `json:"name"`
	Comment                 string                  `json:"comment"`
	Types                   []string                `json:"types"`
	Port                    int                     `json:"port"`
	NoRedirectServer        bool                    `json:"no_redirect_server"`
	Protocol                string                  `json:"protocol"`
	Hypervisor              string                  `json:"hypervisor"`
	Vga                     string                  `json:"vga"`
	VgaRender               string                  `json:"vga_render"`
	Gui                     bool                    `json:"gui"`
	GuiUser                 string                  `json:"gui_user"`
	GuiMode                 string                  `json:"gui_mode"`
	Certificates            []primitive.ObjectID    `json:"certificates"`
	AdminDomain             string                  `json:"admin_domain"`
	UserDomain              string                  `json:"user_domain"`
	WebauthnDomain          string                  `json:"webauthn_domain"`
	ExternalInterfaces      []string                `json:"external_interfaces"`
	ExternalInterfaces6     []string                `json:"external_interfaces6"`
	InternalInterfaces      []string                `json:"internal_interfaces"`
	OracleSubnets           []string                `json:"oracle_subnets"`
	NetworkMode             string                  `json:"network_mode"`
	NetworkMode6            string                  `json:"network_mode6"`
	Blocks                  []*node.BlockAttachment `json:"blocks"`
	Blocks6                 []*node.BlockAttachment `json:"blocks6"`
	InstanceDrives          []*drive.Device         `json:"instance_drives"`
	NoHostNetwork           bool                    `json:"no_host_network"`
	NoNodePortNetwork       bool                    `json:"no_node_port_network"`
	HostNat                 bool                    `json:"host_nat"`
	DefaultNoPublicAddress  bool                    `json:"default_no_public_address"`
	DefaultNoPublicAddress6 bool                    `json:"default_no_public_address6"`
	JumboFrames             bool                    `json:"jumbo_frames"`
	JumboFramesInternal     bool                    `json:"jumbo_frames_internal"`
	Iscsi                   bool                    `json:"iscsi"`
	UsbPassthrough          bool                    `json:"usb_passthrough"`
	PciPassthrough          bool                    `json:"pci_passthrough"`
	Hugepages               bool                    `json:"hugepages"`
	HugepagesSize           int                     `json:"hugepages_size"`
	ForwardedForHeader      string                  `json:"forwarded_for_header"`
	ForwardedProtoHeader    string                  `json:"forwarded_proto_header"`
	Firewall                bool                    `json:"firewall"`
	NetworkRoles            []string                `json:"network_roles"`
	OracleUser              string                  `json:"oracle_user"`
}

type nodesData struct {
	Nodes []*node.Node `json:"nodes"`
	Count int64        `json:"count"`
}

type nodeInitData struct {
	Provider          string             `json:"provider"`
	Zone              primitive.ObjectID `json:"zone"`
	Firewall          bool               `json:"firewall"`
	InternalInterface string             `json:"internal_interface"`
	ExternalInterface string             `json:"external_interface"`
	BlockGateway      string             `json:"block_gateway"`
	BlockNetmask      string             `json:"block_netmask"`
	BlockSubnets      []string           `json:"block_subnets"`
}

func nodePut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &nodeData{}

	nodeId, ok := utils.ParseObjectId(c.Param("node_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	nde, err := node.Get(db, nodeId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	nde.Name = data.Name
	nde.Comment = data.Comment
	nde.Types = data.Types
	nde.Port = data.Port
	nde.NoRedirectServer = data.NoRedirectServer
	nde.Protocol = data.Protocol
	nde.Hypervisor = data.Hypervisor
	nde.Vga = data.Vga
	nde.VgaRender = data.VgaRender
	nde.Gui = data.Gui
	nde.GuiUser = data.GuiUser
	nde.GuiMode = data.GuiMode
	nde.Certificates = data.Certificates
	nde.AdminDomain = data.AdminDomain
	nde.UserDomain = data.UserDomain
	nde.WebauthnDomain = data.WebauthnDomain
	nde.ExternalInterfaces = data.ExternalInterfaces
	nde.ExternalInterfaces6 = data.ExternalInterfaces6
	nde.InternalInterfaces = data.InternalInterfaces
	nde.OracleSubnets = data.OracleSubnets
	nde.NetworkMode = data.NetworkMode
	nde.NetworkMode6 = data.NetworkMode6
	nde.Blocks = data.Blocks
	nde.Blocks6 = data.Blocks6
	nde.InstanceDrives = data.InstanceDrives
	nde.NoHostNetwork = data.NoHostNetwork
	nde.NoNodePortNetwork = data.NoNodePortNetwork
	nde.HostNat = data.HostNat
	nde.DefaultNoPublicAddress = data.DefaultNoPublicAddress
	nde.DefaultNoPublicAddress6 = data.DefaultNoPublicAddress6
	nde.JumboFrames = data.JumboFrames
	nde.JumboFramesInternal = data.JumboFramesInternal
	nde.Iscsi = data.Iscsi
	nde.UsbPassthrough = data.UsbPassthrough
	nde.PciPassthrough = data.PciPassthrough
	nde.Hugepages = data.Hugepages
	nde.HugepagesSize = data.HugepagesSize
	nde.ForwardedForHeader = data.ForwardedForHeader
	nde.ForwardedProtoHeader = data.ForwardedProtoHeader
	nde.Firewall = data.Firewall
	nde.NetworkRoles = data.NetworkRoles
	nde.OracleUser = data.OracleUser

	fields := set.NewSet(
		"name",
		"comment",
		"zone",
		"types",
		"port",
		"no_redirect_server",
		"protocol",
		"hypervisor",
		"vga",
		"vga_render",
		"gui",
		"gui_user",
		"gui_mode",
		"certificates",
		"admin_domain",
		"user_domain",
		"webauthn_domain",
		"external_interfaces",
		"external_interfaces6",
		"internal_interfaces",
		"oracle_subnets",
		"network_mode",
		"network_mode6",
		"blocks",
		"blocks6",
		"instance_drives",
		"no_host_network",
		"no_node_port_network",
		"host_nat",
		"default_no_public_address",
		"default_no_public_address6",
		"jumbo_frames",
		"jumbo_frames_internal",
		"iscsi",
		"usb_passthrough",
		"pci_passthrough",
		"hugepages",
		"hugepages_size",
		"forwarded_for_header",
		"forwarded_proto_header",
		"firewall",
		"network_roles",
		"oracle_user",
	)

	if !data.Zone.IsZero() && data.Zone != nde.Zone {
		if !nde.Zone.IsZero() {
			errData := &errortypes.ErrorData{
				Error:   "zone_modified",
				Message: "Cannot modify zone once set",
			}
			c.JSON(400, errData)
			return
		}
		nde.Zone = data.Zone
	}

	errData, err := nde.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = nde.CommitFields(db, fields)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "node.change")

	c.JSON(200, nde)
}

func nodeOperationPut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)

	nodeId, ok := utils.ParseObjectId(c.Param("node_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	operation := c.Param("operation")
	if operation != node.Restart {
		utils.AbortWithStatus(c, 400)
		return
	}

	nde, err := node.Get(db, nodeId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	nde.Operation = node.Restart

	errData, err := nde.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = nde.CommitFields(db, set.NewSet("operation"))
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "node.change")

	c.JSON(200, nde)
}

func nodeInitPost(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &nodeInitData{}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	nodeId, ok := utils.ParseObjectId(c.Param("node_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	nde, err := node.Get(db, nodeId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	fields := set.NewSet(
		"host_nat",
		"zone",
		"network_mode",
		"network_mode6",
		"internal_interfaces",
		"external_interfaces",
	)

	nde.Zone = data.Zone
	nde.HostNat = true

	if data.Provider == "phoenixnap" {
		fields.Add("default_no_public_address")
		nde.DefaultNoPublicAddress = true
		nde.NetworkMode = node.Static
		nde.NetworkMode6 = node.Disabled
		nde.InternalInterfaces = []string{
			data.InternalInterface,
		}
	} else if data.Provider == "vultr" {
		fields.Add("default_no_public_address")
		nde.DefaultNoPublicAddress = true
		nde.NetworkMode = node.Disabled
		nde.NetworkMode6 = node.Dhcp
		nde.InternalInterfaces = []string{
			settings.Hypervisor.HostNetworkName,
		}
		nde.ExternalInterfaces = []string{
			data.ExternalInterface,
		}
	} else {
		nde.NetworkMode = node.Disabled
		nde.NetworkMode6 = node.Dhcp
		nde.InternalInterfaces = []string{
			settings.Hypervisor.HostNetworkName,
		}
		nde.ExternalInterfaces = []string{
			data.ExternalInterface,
		}
	}

	dc, err := datacenter.Get(db, nde.Datacenter)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	dc.NetworkMode = datacenter.Default

	err = dc.CommitFields(db, set.NewSet("network_mode"))
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "zone.change")

	if data.Provider == "phoenixnap" {
		publicBlck := &block.Block{
			Name:    nde.Name + "-public",
			Type:    block.IPv4,
			Subnets: data.BlockSubnets,
		}

		if data.BlockNetmask == "" {
			_, gateway, err := net.ParseCIDR(data.BlockGateway)
			if err != nil {
				errData := &errortypes.ErrorData{
					Error:   "invalid_block_gateway",
					Message: "Invalid public gateway",
				}
				c.JSON(400, errData)
				return
			}

			publicBlck.Netmask = fmt.Sprintf(
				"%d.%d.%d.%d",
				gateway.Mask[0],
				gateway.Mask[1],
				gateway.Mask[2],
				gateway.Mask[3],
			)
			publicBlck.Gateway = strings.Split(data.BlockGateway, "/")[0]
		} else {
			publicBlck.Netmask = data.BlockNetmask
			publicBlck.Gateway = data.BlockGateway
		}

		errData, err := publicBlck.Validate(db)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}

		if errData != nil {
			c.JSON(400, errData)
			return
		}

		err = publicBlck.Insert(db)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}

		nde.Blocks = []*node.BlockAttachment{
			&node.BlockAttachment{
				Interface: data.ExternalInterface,
				Block:     publicBlck.Id,
			},
		}
		fields.Add("blocks")
	}

	errData, err := nde.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	if data.Firewall {
		nde.Firewall = true
		fields.Add("firewall")

		if nde.NetworkRoles == nil {
			nde.NetworkRoles = []string{}
		}

		hasRole := false
		for _, role := range nde.NetworkRoles {
			if role == "firewall" {
				hasRole = true
			}
		}

		if !hasRole {
			nde.NetworkRoles = append(nde.NetworkRoles, "firewall")
			fields.Add("network_roles")
		}

		fires, err := firewall.GetAll(db, &bson.M{
			"organization": &bson.M{
				"$exists": false,
			},
			"network_roles": "firewall",
		})
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}

		if len(fires) == 0 {
			fire := &firewall.Firewall{
				Name:    "node-firewall",
				Comment: "",
				NetworkRoles: []string{
					"firewall",
				},
				Ingress: []*firewall.Rule{
					&firewall.Rule{
						SourceIps: []string{
							"0.0.0.0/0",
							"::/0",
						},
						Protocol: firewall.Icmp,
					},
					&firewall.Rule{
						SourceIps: []string{
							"0.0.0.0/0",
							"::/0",
						},
						Protocol: firewall.Tcp,
						Port:     "22",
					},
					&firewall.Rule{
						SourceIps: []string{
							"0.0.0.0/0",
							"::/0",
						},
						Protocol: firewall.Tcp,
						Port:     "80",
					},
					&firewall.Rule{
						SourceIps: []string{
							"0.0.0.0/0",
							"::/0",
						},
						Protocol: firewall.Tcp,
						Port:     "443",
					},
				},
			}

			errData, err = fire.Validate(db)
			if err != nil {
				utils.AbortWithError(c, 500, err)
				return
			}

			if errData != nil {
				c.JSON(400, errData)
				return
			}

			err = fire.Insert(db)
			if err != nil {
				utils.AbortWithError(c, 500, err)
				return
			}
		}
	}

	err = nde.CommitFields(db, fields)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "node.change")
	event.PublishDispatch(db, "block.change")
	if data.Firewall {
		event.PublishDispatch(db, "firewall.change")
	}

	c.JSON(200, nde)
}

func nodeDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)

	nodeId, ok := utils.ParseObjectId(c.Param("node_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := node.Remove(db, nodeId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "node.change")

	c.JSON(200, nil)
}

func nodeGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	nodeId, ok := utils.ParseObjectId(c.Param("node_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	nde, err := node.Get(db, nodeId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if demo.IsDemo() {
		nde.RequestsMin = 32
		nde.Memory = 25.0
		nde.Load1 = 10.0
		nde.Load5 = 15.0
		nde.Load15 = 20.0
	}

	c.JSON(200, nde)
}

func nodesGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	if c.Query("names") == "true" {
		zone, _ := utils.ParseObjectId(c.Query("zone"))

		query := &bson.M{
			"zone": zone,
		}

		nodes, err := node.GetAllHypervisors(db, query)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}

		c.JSON(200, nodes)
	} else {
		page, _ := strconv.ParseInt(c.Query("page"), 10, 0)
		pageCount, _ := strconv.ParseInt(c.Query("page_count"), 10, 0)

		query := bson.M{}

		nodeId, ok := utils.ParseObjectId(c.Query("id"))
		if ok {
			query["_id"] = nodeId
		}

		name := strings.TrimSpace(c.Query("name"))
		if name != "" {
			query["name"] = &bson.M{
				"$regex":   fmt.Sprintf(".*%s.*", regexp.QuoteMeta(name)),
				"$options": "i",
			}
		}

		zone, _ := utils.ParseObjectId(c.Query("zone"))
		if !zone.IsZero() {
			query["zone"] = zone
		}

		networkRole := c.Query("network_role")
		if networkRole != "" {
			query["network_roles"] = networkRole
		}

		types := []string{}
		notTypes := []string{}

		adminType := c.Query(node.Admin)
		switch adminType {
		case "true":
			types = append(types, node.Admin)
			break
		case "false":
			notTypes = append(notTypes, node.Admin)
			break
		}

		userType := c.Query(node.User)
		switch userType {
		case "true":
			types = append(types, node.User)
			break
		case "false":
			notTypes = append(notTypes, node.User)
			break
		}

		hypervisorType := c.Query(node.Hypervisor)
		switch hypervisorType {
		case "true":
			types = append(types, node.Hypervisor)
			break
		case "false":
			notTypes = append(notTypes, node.Hypervisor)
			break
		}

		typesQuery := bson.M{}
		if len(types) > 0 {
			typesQuery["$all"] = types
		}
		if len(notTypes) > 0 {
			typesQuery["$nin"] = notTypes
		}
		if len(types) > 0 || len(notTypes) > 0 {
			query["types"] = &typesQuery
		}

		nodes, count, err := node.GetAllPaged(db, &query, page, pageCount)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}

		if demo.IsDemo() {
			for _, nde := range nodes {
				nde.RequestsMin = 32
				nde.Memory = 25.0
				nde.Load1 = 10.0
				nde.Load5 = 15.0
				nde.Load15 = 20.0
			}
		}

		data := &nodesData{
			Nodes: nodes,
			Count: count,
		}

		c.JSON(200, data)
	}
}

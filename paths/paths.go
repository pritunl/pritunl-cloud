package paths

import (
	"crypto/md5"
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"path"
	"strings"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/settings"
)

func GetVmUuid(instId bson.ObjectID) string {
	idHash := md5.New()
	idHash.Write(instId[:])
	uuid := idHash.Sum(nil)

	uuid[6] = (uuid[6] & 0x0f) | uint8((3&0xf)<<4)
	uuid[8] = (uuid[8] & 0x3f) | 0x80

	buffer := [36]byte{}
	hex.Encode(buffer[:], uuid[:4])
	buffer[8] = '-'
	hex.Encode(buffer[9:13], uuid[4:6])
	buffer[13] = '-'
	hex.Encode(buffer[14:18], uuid[6:8])
	buffer[18] = '-'
	hex.Encode(buffer[19:23], uuid[8:10])
	buffer[23] = '-'
	hex.Encode(buffer[24:], uuid[10:])
	return string(buffer[:])
}

func GetVmPath(instId bson.ObjectID) string {
	return path.Join(node.Self.GetVirtPath(),
		"instances", instId.Hex())
}

func GetDisksPath() string {
	return path.Join(node.Self.GetVirtPath(), "disks")
}

func GetLocalIsosPath() string {
	return path.Join(node.Self.GetVirtPath(), "isos")
}

func GetBackingPath() string {
	return path.Join(node.Self.GetVirtPath(), "backing")
}

func GetTpmsPath() string {
	return path.Join(node.Self.GetVirtPath(), "tpms")
}

func GetTpmPath(virtId bson.ObjectID) string {
	return path.Join(GetTpmsPath(), virtId.Hex())
}

func GetTpmSockPath(virtId bson.ObjectID) string {
	return path.Join(GetTpmsPath(), virtId.Hex(), "sock")
}

func GetTpmPwdPath(virtId bson.ObjectID) string {
	return path.Join(GetTpmsPath(), virtId.Hex(), "pwd")
}

func GetTempPath() string {
	return node.Self.GetTempPath()
}

func GetTempDir() string {
	return path.Join(GetTempPath(), bson.NewObjectID().Hex())
}

func GetDrivePath(driveId string) string {
	return path.Join("/dev/disk/by-id", driveId)
}

func GetCachesDir() string {
	return path.Join(node.Self.GetVirtPath(), "caches")
}

func GetCacheDir(virtId bson.ObjectID) string {
	return path.Join(GetCachesDir(), virtId.Hex())
}

func GetOvmfDir() string {
	return path.Join(node.Self.GetVirtPath(), "ovmf")
}

func GetDiskPath(diskId bson.ObjectID) string {
	return path.Join(GetDisksPath(),
		fmt.Sprintf("%s.qcow2", diskId.Hex()))
}

func GetOvmfVarsPath(virtId bson.ObjectID) string {
	return path.Join(GetOvmfDir(),
		fmt.Sprintf("%s_vars.fd", virtId.Hex()))
}

func GetDiskTempPath() string {
	return path.Join(GetTempPath(),
		fmt.Sprintf("disk-%s", bson.NewObjectID().Hex()))
}

func GetImageTempPath() string {
	return path.Join(GetTempPath(),
		fmt.Sprintf("image-%s", bson.NewObjectID().Hex()))
}

func GetImdsPath() string {
	return path.Join(node.Self.GetVirtPath(), "imds")
}

func GetImdsConfPath(instId bson.ObjectID) string {
	return path.Join(GetImdsPath(),
		fmt.Sprintf("%s-conf.json", instId.Hex()))
}

func GetInstRunPath(instId bson.ObjectID) string {
	return path.Join(settings.Hypervisor.RunPath, instId.Hex())
}

func GetImdsSockPath(instId bson.ObjectID) string {
	return path.Join(GetInstRunPath(instId), "imds.sock")
}

func GetDiskMountPath() string {
	return path.Join(GetTempPath(), bson.NewObjectID().Hex())
}

func GetInitsPath() string {
	return path.Join(node.Self.GetVirtPath(), "inits")
}

func GetInitPath(instId bson.ObjectID) string {
	return path.Join(GetInitsPath(),
		fmt.Sprintf("%s.iso", instId.Hex()))
}

func GetUnitName(virtId bson.ObjectID) string {
	return fmt.Sprintf("pritunl_cloud_%s.service", virtId.Hex())
}

func GetUnitPath(virtId bson.ObjectID) string {
	return path.Join(settings.Hypervisor.SystemdPath, GetUnitName(virtId))
}

func GetUnitNameDhcp4(virtId bson.ObjectID, n int) string {
	return fmt.Sprintf("pritunl_dhcp4_%s_%d.service", virtId.Hex(), n)
}

func GetUnitPathDhcp4(virtId bson.ObjectID, n int) string {
	return path.Join(settings.Hypervisor.SystemdPath, GetUnitNameDhcp4(virtId, n))
}

func GetUnitNameDhcp6(virtId bson.ObjectID, n int) string {
	return fmt.Sprintf("pritunl_dhcp6_%s_%d.service", virtId.Hex(), n)
}

func GetUnitPathDhcp6(virtId bson.ObjectID, n int) string {
	return path.Join(settings.Hypervisor.SystemdPath, GetUnitNameDhcp6(virtId, n))
}

func GetUnitNameNdp(virtId bson.ObjectID, n int) string {
	return fmt.Sprintf("pritunl_ndp_%s_%d.service", virtId.Hex(), n)
}

func GetUnitPathNdp(virtId bson.ObjectID, n int) string {
	return path.Join(settings.Hypervisor.SystemdPath,
		GetUnitNameNdp(virtId, n))
}

func GetUnitNameTpm(virtId bson.ObjectID) string {
	return fmt.Sprintf("pritunl_tpm_%s.service", virtId.Hex())
}

func GetUnitPathTpm(virtId bson.ObjectID) string {
	return path.Join(settings.Hypervisor.SystemdPath,
		GetUnitNameTpm(virtId))
}

func GetUnitNameImds(virtId bson.ObjectID) string {
	return fmt.Sprintf("pritunl_imds_%s.service", virtId.Hex())
}

func GetUnitNameDhcpc(virtId bson.ObjectID) string {
	return fmt.Sprintf("pritunl_dhcpc_%s.service", virtId.Hex())
}

func GetShareId(virtId bson.ObjectID, shareName string) string {
	hash := md5.New()
	hash.Write([]byte(virtId.Hex()))
	hash.Write([]byte(shareName))
	return strings.ToLower(base32.StdEncoding.EncodeToString(
		hash.Sum(nil))[:12])
}

func GetUnitNameShare(virtId bson.ObjectID, shareId string) string {
	return fmt.Sprintf("pritunl_share_%s_%s.service", virtId.Hex(), shareId)
}

func GetUnitNameShares(virtId bson.ObjectID) string {
	return fmt.Sprintf("pritunl_share_%s_*.service", virtId.Hex())
}

func GetUnitPathImds(virtId bson.ObjectID) string {
	return path.Join(settings.Hypervisor.SystemdPath,
		GetUnitNameImds(virtId))
}

func GetUnitPathDhcpc(virtId bson.ObjectID) string {
	return path.Join(settings.Hypervisor.SystemdPath,
		GetUnitNameDhcpc(virtId))
}

func GetUnitPathShare(virtId bson.ObjectID, shareId string) string {
	return path.Join(settings.Hypervisor.SystemdPath,
		GetUnitNameShare(virtId, shareId))
}

func GetUnitPathShares(virtId bson.ObjectID) string {
	return path.Join(settings.Hypervisor.SystemdPath,
		GetUnitNameShares(virtId))
}

func GetPidPath(virtId bson.ObjectID) string {
	return path.Join(settings.Hypervisor.RunPath,
		fmt.Sprintf("%s.pid", virtId.Hex()))
}

func GetShareSockPath(virtId bson.ObjectID, shareId string) string {
	return path.Join(GetInstRunPath(virtId),
		fmt.Sprintf("virtiofs_%s.sock", shareId))
}

func GetHugepagePath(virtId bson.ObjectID) string {
	return path.Join(settings.Hypervisor.HugepagesPath, virtId.Hex())
}

func GetSockPath(virtId bson.ObjectID) string {
	return path.Join(settings.Hypervisor.RunPath,
		fmt.Sprintf("%s.sock", virtId.Hex()))
}

func GetQmpSockPath(virtId bson.ObjectID) string {
	return path.Join(settings.Hypervisor.RunPath,
		fmt.Sprintf("%s.qmp.sock", virtId.Hex()))
}

func GetGuestPath(virtId bson.ObjectID) string {
	return path.Join(settings.Hypervisor.RunPath,
		fmt.Sprintf("%s.guest", virtId.Hex()))
}

// TODO Backward compatibility
func GetPidPathOld(virtId bson.ObjectID) string {
	return path.Join(settings.Hypervisor.LibPath,
		fmt.Sprintf("%s.pid", virtId.Hex()))
}

// TODO Backward compatibility
func GetSockPathOld(virtId bson.ObjectID) string {
	return path.Join(settings.Hypervisor.LibPath,
		fmt.Sprintf("%s.sock", virtId.Hex()))
}

// TODO Backward compatibility
func GetQmpSockPathOld(virtId bson.ObjectID) string {
	return path.Join(settings.Hypervisor.LibPath,
		fmt.Sprintf("%s.qmp.sock", virtId.Hex()))
}

// TODO Backward compatibility
func GetGuestPathOld(virtId bson.ObjectID) string {
	return path.Join(settings.Hypervisor.LibPath,
		fmt.Sprintf("%s.guest", virtId.Hex()))
}

func GetNamespacesPath() string {
	return "/etc/netns"
}

func GetNamespacePath(namespace string) string {
	return path.Join(GetNamespacesPath(), namespace)
}

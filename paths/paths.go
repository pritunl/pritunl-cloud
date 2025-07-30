package paths

import (
	"crypto/md5"
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"path"
	"strings"

	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/settings"
)

func GetVmUuid(instId primitive.ObjectID) string {
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

func GetVmPath(instId primitive.ObjectID) string {
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

func GetTpmPath(virtId primitive.ObjectID) string {
	return path.Join(GetTpmsPath(), virtId.Hex())
}

func GetTpmSockPath(virtId primitive.ObjectID) string {
	return path.Join(GetTpmsPath(), virtId.Hex(), "sock")
}

func GetTpmPwdPath(virtId primitive.ObjectID) string {
	return path.Join(GetTpmsPath(), virtId.Hex(), "pwd")
}

func GetTempPath() string {
	return node.Self.GetTempPath()
}

func GetTempDir() string {
	return path.Join(GetTempPath(), primitive.NewObjectID().Hex())
}

func GetDrivePath(driveId string) string {
	return path.Join("/dev/disk/by-id", driveId)
}

func GetCachesDir() string {
	return path.Join(node.Self.GetVirtPath(), "caches")
}

func GetCacheDir(virtId primitive.ObjectID) string {
	return path.Join(GetCachesDir(), virtId.Hex())
}

func GetOvmfDir() string {
	return path.Join(node.Self.GetVirtPath(), "ovmf")
}

func GetDiskPath(diskId primitive.ObjectID) string {
	return path.Join(GetDisksPath(),
		fmt.Sprintf("%s.qcow2", diskId.Hex()))
}

func GetOvmfVarsPath(virtId primitive.ObjectID) string {
	return path.Join(GetOvmfDir(),
		fmt.Sprintf("%s_vars.fd", virtId.Hex()))
}

func GetDiskTempPath() string {
	return path.Join(GetTempPath(),
		fmt.Sprintf("disk-%s", primitive.NewObjectID().Hex()))
}

func GetImageTempPath() string {
	return path.Join(GetTempPath(),
		fmt.Sprintf("image-%s", primitive.NewObjectID().Hex()))
}

func GetImdsPath() string {
	return path.Join(node.Self.GetVirtPath(), "imds")
}

func GetImdsConfPath(instId primitive.ObjectID) string {
	return path.Join(GetImdsPath(),
		fmt.Sprintf("%s-conf.json", instId.Hex()))
}

func GetInstRunPath(instId primitive.ObjectID) string {
	return path.Join(settings.Hypervisor.RunPath, instId.Hex())
}

func GetImdsSockPath(instId primitive.ObjectID) string {
	return path.Join(GetInstRunPath(instId), "imds.sock")
}

func GetDiskMountPath() string {
	return path.Join(GetTempPath(), primitive.NewObjectID().Hex())
}

func GetInitsPath() string {
	return path.Join(node.Self.GetVirtPath(), "inits")
}

func GetInitPath(instId primitive.ObjectID) string {
	return path.Join(GetInitsPath(),
		fmt.Sprintf("%s.iso", instId.Hex()))
}

func GetUnitName(virtId primitive.ObjectID) string {
	return fmt.Sprintf("pritunl_cloud_%s.service", virtId.Hex())
}

func GetUnitPath(virtId primitive.ObjectID) string {
	return path.Join(settings.Hypervisor.SystemdPath, GetUnitName(virtId))
}

func GetUnitNameDhcp4(virtId primitive.ObjectID, n int) string {
	return fmt.Sprintf("pritunl_dhcp4_%s_%d.service", virtId.Hex(), n)
}

func GetUnitPathDhcp4(virtId primitive.ObjectID, n int) string {
	return path.Join(settings.Hypervisor.SystemdPath, GetUnitNameDhcp4(virtId, n))
}

func GetUnitNameDhcp6(virtId primitive.ObjectID, n int) string {
	return fmt.Sprintf("pritunl_dhcp6_%s_%d.service", virtId.Hex(), n)
}

func GetUnitPathDhcp6(virtId primitive.ObjectID, n int) string {
	return path.Join(settings.Hypervisor.SystemdPath, GetUnitNameDhcp6(virtId, n))
}

func GetUnitNameNdp(virtId primitive.ObjectID, n int) string {
	return fmt.Sprintf("pritunl_ndp_%s_%d.service", virtId.Hex(), n)
}

func GetUnitPathNdp(virtId primitive.ObjectID, n int) string {
	return path.Join(settings.Hypervisor.SystemdPath,
		GetUnitNameNdp(virtId, n))
}

func GetUnitNameTpm(virtId primitive.ObjectID) string {
	return fmt.Sprintf("pritunl_tpm_%s.service", virtId.Hex())
}

func GetUnitPathTpm(virtId primitive.ObjectID) string {
	return path.Join(settings.Hypervisor.SystemdPath,
		GetUnitNameTpm(virtId))
}

func GetUnitNameImds(virtId primitive.ObjectID) string {
	return fmt.Sprintf("pritunl_imds_%s.service", virtId.Hex())
}

func GetUnitNameDhcpc(virtId primitive.ObjectID) string {
	return fmt.Sprintf("pritunl_dhcpc_%s.service", virtId.Hex())
}

func GetShareId(virtId primitive.ObjectID, shareName string) string {
	hash := md5.New()
	hash.Write([]byte(virtId.Hex()))
	hash.Write([]byte(shareName))
	return strings.ToLower(base32.StdEncoding.EncodeToString(
		hash.Sum(nil))[:12])
}

func GetUnitNameShare(virtId primitive.ObjectID, shareId string) string {
	return fmt.Sprintf("pritunl_share_%s_%s.service", virtId.Hex(), shareId)
}

func GetUnitNameShares(virtId primitive.ObjectID) string {
	return fmt.Sprintf("pritunl_share_%s_*.service", virtId.Hex())
}

func GetUnitPathImds(virtId primitive.ObjectID) string {
	return path.Join(settings.Hypervisor.SystemdPath,
		GetUnitNameImds(virtId))
}

func GetUnitPathDhcpc(virtId primitive.ObjectID) string {
	return path.Join(settings.Hypervisor.SystemdPath,
		GetUnitNameDhcpc(virtId))
}

func GetUnitPathShare(virtId primitive.ObjectID, shareId string) string {
	return path.Join(settings.Hypervisor.SystemdPath,
		GetUnitNameShare(virtId, shareId))
}

func GetUnitPathShares(virtId primitive.ObjectID) string {
	return path.Join(settings.Hypervisor.SystemdPath,
		GetUnitNameShares(virtId))
}

func GetPidPath(virtId primitive.ObjectID) string {
	return path.Join(settings.Hypervisor.RunPath,
		fmt.Sprintf("%s.pid", virtId.Hex()))
}

func GetShareSockPath(virtId primitive.ObjectID, shareId string) string {
	return path.Join(GetInstRunPath(virtId),
		fmt.Sprintf("virtiofs_%s.sock", shareId))
}

func GetHugepagePath(virtId primitive.ObjectID) string {
	return path.Join(settings.Hypervisor.HugepagesPath, virtId.Hex())
}

func GetSockPath(virtId primitive.ObjectID) string {
	return path.Join(settings.Hypervisor.RunPath,
		fmt.Sprintf("%s.sock", virtId.Hex()))
}

func GetQmpSockPath(virtId primitive.ObjectID) string {
	return path.Join(settings.Hypervisor.RunPath,
		fmt.Sprintf("%s.qmp.sock", virtId.Hex()))
}

func GetGuestPath(virtId primitive.ObjectID) string {
	return path.Join(settings.Hypervisor.RunPath,
		fmt.Sprintf("%s.guest", virtId.Hex()))
}

// TODO Backward compatibility
func GetPidPathOld(virtId primitive.ObjectID) string {
	return path.Join(settings.Hypervisor.LibPath,
		fmt.Sprintf("%s.pid", virtId.Hex()))
}

// TODO Backward compatibility
func GetSockPathOld(virtId primitive.ObjectID) string {
	return path.Join(settings.Hypervisor.LibPath,
		fmt.Sprintf("%s.sock", virtId.Hex()))
}

// TODO Backward compatibility
func GetQmpSockPathOld(virtId primitive.ObjectID) string {
	return path.Join(settings.Hypervisor.LibPath,
		fmt.Sprintf("%s.qmp.sock", virtId.Hex()))
}

// TODO Backward compatibility
func GetGuestPathOld(virtId primitive.ObjectID) string {
	return path.Join(settings.Hypervisor.LibPath,
		fmt.Sprintf("%s.guest", virtId.Hex()))
}

func GetNamespacesPath() string {
	return "/etc/netns"
}

func GetNamespacePath(namespace string) string {
	return path.Join(GetNamespacesPath(), namespace)
}

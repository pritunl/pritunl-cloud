package statec

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/pritunl/mongo-go-driver/v2/bson"
	"github.com/pritunl/pritunl-cloud/deployment"
	"github.com/pritunl/pritunl-cloud/instance"
	"github.com/pritunl/pritunl-cloud/node"
	"github.com/pritunl/pritunl-cloud/spec"
	"github.com/pritunl/pritunl-cloud/state"
)

const (
	diffWindow    = 60
	diffMaxDetail = 600
)

type Diff struct {
	Name   string
	Key    string
	Detail string
	State  string
	Statec string
}

func (d *Diff) Ident() string {
	return d.Name + "|" + d.Key
}

func (d *Diff) CaughtUp(prev *Diff) bool {
	return prev != nil && d.Statec == prev.State
}

func (d *Diff) String() string {
	if d.Key == "" {
		return fmt.Sprintf("%s: %s", d.Name, d.Detail)
	}
	return fmt.Sprintf("%s[%s]: %s", d.Name, d.Key, d.Detail)
}

type comparer struct {
	diffs []*Diff
}

func (c *comparer) add(name, key, detail, state, statec string) {
	if len(detail) > diffMaxDetail {
		detail = detail[:diffMaxDetail] + "..."
	}

	c.diffs = append(c.diffs, &Diff{
		Name:   name,
		Key:    key,
		Detail: detail,
		State:  state,
		Statec: statec,
	})
}

func window(s string, i int) string {
	start := i - diffWindow
	if start < 0 {
		start = 0
	}
	end := i + diffWindow
	if end > len(s) {
		end = len(s)
	}
	return s[start:end]
}

func firstDiff(a, b string) string {
	i := 0
	for i < len(a) && i < len(b) && a[i] == b[i] {
		i++
	}

	return fmt.Sprintf("at %d (len %d != %d): %q != %q",
		i, len(a), len(b), window(a, i), window(b, i))
}

func (c *comparer) values(name, key string, a, b interface{}) {
	ca := Canon(a)
	cb := Canon(b)
	if ca == cb {
		return
	}

	detail := fieldsDetail(a, b)
	if detail != "" {
		detail = "fields: " + detail + "; "
	}

	c.add(name, key, detail+firstDiff(ca, cb), ca, cb)
}

func diffStrings(a, b []string) (onlyA, onlyB []string) {
	counts := map[string]int{}
	for _, item := range a {
		counts[item] += 1
	}
	for _, item := range b {
		counts[item] -= 1
	}

	for item, count := range counts {
		for ; count > 0; count-- {
			onlyA = append(onlyA, item)
		}
		for ; count < 0; count++ {
			onlyB = append(onlyB, item)
		}
	}

	sort.Strings(onlyA)
	sort.Strings(onlyB)

	return
}

func (c *comparer) list(name, key string, a, b interface{}) {
	la := canonList(a)
	lb := canonList(b)

	onlyA, onlyB := diffStrings(la, lb)
	if len(onlyA) == 0 && len(onlyB) == 0 {
		return
	}

	itemsA, okA := listById(a)
	itemsB, okB := listById(b)
	if okA && okB {
		c.listById(name, key, itemsA, itemsB)
		return
	}

	detail := fmt.Sprintf("len %d != %d, only state %d, only statec %d",
		len(la), len(lb), len(onlyA), len(onlyB))
	if len(onlyA) > 0 && len(onlyB) > 0 {
		detail += " " + firstDiff(onlyA[0], onlyB[0])
	} else if len(onlyA) > 0 {
		detail += " state " + window(onlyA[0], 0)
	} else {
		detail += " statec " + window(onlyB[0], 0)
	}

	c.add(name, key, detail,
		strings.Join(la, "\n"), strings.Join(lb, "\n"))
}

func elemKey(key, id string) string {
	if key == "" {
		return id
	}
	return key + ":" + id
}

func unionKeys(a, b map[string]reflect.Value) (keys []string) {
	seen := map[string]struct{}{}
	for key := range a {
		seen[key] = struct{}{}
	}
	for key := range b {
		seen[key] = struct{}{}
	}

	keys = make([]string, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	return
}

func (c *comparer) listById(name, key string,
	itemsA, itemsB map[string]reflect.Value) {

	for _, id := range unionKeys(itemsA, itemsB) {
		valA, okA := itemsA[id]
		valB, okB := itemsB[id]

		switch {
		case !okA:
			c.add(name, elemKey(key, id), "only in statec",
				"", canonString(valB))
		case !okB:
			c.add(name, elemKey(key, id), "only in state",
				canonString(valA), "")
		default:
			c.values(name, elemKey(key, id),
				valA.Interface(), valB.Interface())
		}
	}
}

func mapValues(val interface{}) (vals map[string]reflect.Value) {
	vals = map[string]reflect.Value{}

	v := reflect.ValueOf(val)
	if !v.IsValid() || v.Kind() != reflect.Map || v.IsNil() {
		return
	}

	iter := v.MapRange()
	for iter.Next() {
		vals[canonKey(iter.Key())] = iter.Value()
	}

	return
}

func (c *comparer) mapping(name string, a, b interface{}, lists bool) {
	va := mapValues(a)
	vb := mapValues(b)

	for _, key := range unionKeys(va, vb) {
		valA, okA := va[key]
		valB, okB := vb[key]

		if !okA {
			c.add(name, key, "only in statec", "", canonString(valB))
			continue
		}
		if !okB {
			c.add(name, key, "only in state", canonString(valA), "")
			continue
		}

		if lists {
			c.list(name, key, valA.Interface(), valB.Interface())
		} else {
			c.values(name, key, valA.Interface(), valB.Interface())
		}
	}
}

type idUnion map[bson.ObjectID]struct{}

func (u idUnion) add(ids ...bson.ObjectID) {
	for _, id := range ids {
		if id.IsZero() {
			continue
		}
		u[id] = struct{}{}
	}
}

func (u idUnion) addDeployments(
	deplys map[bson.ObjectID]*deployment.Deployment) {

	for deplyId := range deplys {
		u.add(deplyId)
	}
}

func (u idUnion) sorted() (ids []bson.ObjectID) {
	ids = make([]bson.ObjectID, 0, len(u))
	for id := range u {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i].Hex() < ids[j].Hex()
	})
	return
}

func specRefs(specsMap map[bson.ObjectID]*spec.Spec,
	pods, secrets, certs, domains idUnion) {

	for _, spc := range specsMap {
		if spc == nil {
			continue
		}

		if spc.Instance != nil {
			pods.add(spc.Instance.Pods...)
			secrets.add(spc.Instance.Secrets...)
			certs.add(spc.Instance.Certificates...)
		}

		if spc.Domain != nil {
			for _, rec := range spc.Domain.Records {
				if rec != nil {
					domains.add(rec.Domain)
				}
			}
		}
	}
}

func Compare(a *state.State, b *State) (diffs []*Diff) {
	c := &comparer{}

	if a == nil || b == nil {
		c.add("State", "", fmt.Sprintf("nil state %v statec %v",
			a == nil, b == nil), "", "")
		diffs = c.diffs
		return
	}

	instsA := a.Instances()
	instsB := b.Instances()
	insts := map[bson.ObjectID]*instance.Instance{}
	instIds := idUnion{}
	orgIds := idUnion{}
	zoneIds := idUnion{}
	namespaces := map[string]struct{}{}
	for _, inst := range append(append([]*instance.Instance{},
		instsA...), instsB...) {

		if inst == nil {
			continue
		}

		if _, ok := insts[inst.Id]; !ok {
			insts[inst.Id] = inst
		}
		instIds.add(inst.Id)
		orgIds.add(inst.Organization)
		zoneIds.add(inst.Zone)
		namespaces[inst.NetworkNamespace] = struct{}{}
	}

	c.values("NodeDatacenter", "", a.NodeDatacenter(), b.NodeDatacenter())
	c.values("NodeZone", "", a.NodeZone(), b.NodeZone())
	c.values("VxLan", "", a.VxLan(), b.VxLan())

	ndesA := a.Nodes()
	ndesB := b.Nodes()
	c.list("Nodes", "", ndesA, ndesB)

	if node.Self != nil {
		zoneIds.add(node.Self.Zone)
	}
	for _, nde := range append(append([]*node.Node{}, ndesA...), ndesB...) {
		if nde != nil {
			zoneIds.add(nde.Zone)
		}
	}
	for _, zneId := range zoneIds.sorted() {
		c.values("GetZone", zneId.Hex(), a.GetZone(zneId), b.GetZone(zneId))
	}

	nsA := a.Namespaces()
	nsB := b.Namespaces()
	c.list("Namespaces", "", nsA, nsB)
	for _, ns := range append(append([]string{}, nsA...), nsB...) {
		namespaces[ns] = struct{}{}
	}

	ifacesA := a.Interfaces()
	ifacesB := b.Interfaces()
	c.list("Interfaces", "", ifacesA, ifacesB)
	for _, iface := range append(append([]string{}, ifacesA...),
		ifacesB...) {

		c.values("HasInterfaces", iface,
			a.HasInterfaces(iface), b.HasInterfaces(iface))
		c.values("HasBridgedInterfaces", iface,
			a.HasBridgedInterfaces(iface),
			b.HasBridgedInterfaces(iface))
	}

	c.list("NodePools", "", a.NodePools(), b.NodePools())
	c.list("DiskShapes", "", a.DiskShapes(), b.DiskShapes())
	c.list("Schedulers", "", a.Schedulers(), b.Schedulers())

	c.list("Disks", "", a.Disks(), b.Disks())
	c.mapping("InstaceDisksMap", a.InstaceDisksMap(), b.InstaceDisksMap(),
		true)

	c.list("Vpcs", "", a.Vpcs(), b.Vpcs())
	c.mapping("VpcsMap", a.VpcsMap(), b.VpcsMap(), false)
	c.mapping("VpcIpsMap", a.VpcIpsMap(), b.VpcIpsMap(), true)
	for _, instId := range instIds.sorted() {
		inst := insts[instId]
		c.values("Vpc", inst.Vpc.Hex(), a.Vpc(inst.Vpc), b.Vpc(inst.Vpc))
		c.list("VpcIps", inst.Vpc.Hex(), a.VpcIps(inst.Vpc),
			b.VpcIps(inst.Vpc))
	}

	c.mapping("PodsMap", a.PodsMap(), b.PodsMap(), false)
	c.mapping("UnitsMap", a.UnitsMap(), b.UnitsMap(), false)
	c.mapping("SpecsUnitsMap", a.SpecsUnitsMap(), b.SpecsUnitsMap(), false)

	deplysNodeA := a.DeploymentsNode()
	deplysNodeB := b.DeploymentsNode()
	c.mapping("DeploymentsNode", deplysNodeA, deplysNodeB, false)
	c.mapping("DeploymentsReserved", a.DeploymentsReserved(),
		b.DeploymentsReserved(), false)
	c.mapping("DeploymentsDeployed", a.DeploymentsDeployed(),
		b.DeploymentsDeployed(), false)
	c.mapping("DeploymentsDestroy", a.DeploymentsDestroy(),
		b.DeploymentsDestroy(), false)
	c.mapping("DeploymentsInactive", a.DeploymentsInactive(),
		b.DeploymentsInactive(), false)

	deplyIds := idUnion{}
	deplyIds.addDeployments(deplysNodeA)
	deplyIds.addDeployments(deplysNodeB)
	deplyIds.addDeployments(a.DeploymentsReserved())
	deplyIds.addDeployments(b.DeploymentsReserved())
	deplyIds.addDeployments(a.DeploymentsDeployed())
	deplyIds.addDeployments(b.DeploymentsDeployed())
	deplyIds.addDeployments(a.DeploymentsInactive())
	deplyIds.addDeployments(b.DeploymentsInactive())
	for _, deply := range deplysNodeA {
		if deply != nil {
			orgIds.add(deply.Organization)
		}
	}
	for _, deply := range deplysNodeB {
		if deply != nil {
			orgIds.add(deply.Organization)
		}
	}
	for _, dsk := range a.Disks() {
		if dsk != nil {
			deplyIds.add(dsk.Deployment)
		}
	}
	for _, dsk := range b.Disks() {
		if dsk != nil {
			deplyIds.add(dsk.Deployment)
		}
	}

	for _, deplyId := range deplyIds.sorted() {
		c.values("Deployment", deplyId.Hex(),
			a.Deployment(deplyId), b.Deployment(deplyId))
		c.values("DeploymentReserved", deplyId.Hex(),
			a.DeploymentReserved(deplyId), b.DeploymentReserved(deplyId))
		c.values("DeploymentDeployed", deplyId.Hex(),
			a.DeploymentDeployed(deplyId), b.DeploymentDeployed(deplyId))
		c.values("DeploymentInactive", deplyId.Hex(),
			a.DeploymentInactive(deplyId), b.DeploymentInactive(deplyId))
		c.list("GetDeploymentDisks", deplyId.Hex(),
			a.GetDeploymentDisks(deplyId), b.GetDeploymentDisks(deplyId))
	}

	specPods := idUnion{}
	specSecrets := idUnion{}
	specCerts := idUnion{}
	specDomains := idUnion{}
	specIds := idUnion{}
	for _, deply := range deplysNodeA {
		if deply != nil {
			specIds.add(deply.Spec)
		}
	}
	for _, deply := range deplysNodeB {
		if deply != nil {
			specIds.add(deply.Spec)
		}
	}

	specsMap := map[bson.ObjectID]*spec.Spec{}
	for _, specId := range specIds.sorted() {
		spcA := a.Spec(specId)
		spcB := b.Spec(specId)
		c.values("Spec", specId.Hex(), spcA, spcB)
		if spcA != nil {
			specsMap[specId] = spcA
		} else if spcB != nil {
			specsMap[specId] = spcB
		}
	}
	specRefs(specsMap, specPods, specSecrets, specCerts, specDomains)

	if a.SpecsMap != nil && b.SpecsMap != nil {
		specsA := a.SpecsMap()
		specsB := b.SpecsMap()
		c.mapping("SpecsMap", specsA, specsB, false)
		specRefs(specsA, specPods, specSecrets, specCerts, specDomains)
		specRefs(specsB, specPods, specSecrets, specCerts, specDomains)
	}

	for _, pdId := range specPods.sorted() {
		c.values("SpecPod", pdId.Hex(), a.SpecPod(pdId), b.SpecPod(pdId))
		c.list("SpecPodUnits", pdId.Hex(),
			a.SpecPodUnits(pdId), b.SpecPodUnits(pdId))
	}
	for _, secrId := range specSecrets.sorted() {
		c.values("SpecSecret", secrId.Hex(),
			a.SpecSecret(secrId), b.SpecSecret(secrId))
	}
	for _, certId := range specCerts.sorted() {
		c.values("SpecCert", certId.Hex(),
			a.SpecCert(certId), b.SpecCert(certId))
	}
	for _, domnId := range specDomains.sorted() {
		c.values("SpecDomain", domnId.Hex(),
			a.SpecDomain(domnId), b.SpecDomain(domnId))
	}

	for _, orgId := range orgIds.sorted() {
		c.list("GetDomains", orgId.Hex(),
			a.GetDomains(orgId), b.GetDomains(orgId))
	}

	c.list("Instances", "", instsA, instsB)
	c.mapping("NodePortsMap", a.NodePortsMap(), b.NodePortsMap(), true)
	c.mapping("VirtsMap", a.VirtsMap(), b.VirtsMap(), false)

	for _, instId := range instIds.sorted() {
		inst := insts[instId]

		c.values("GetInstace", instId.Hex(),
			a.GetInstace(instId), b.GetInstace(instId))
		c.values("GetVirt", instId.Hex(),
			a.GetVirt(instId), b.GetVirt(instId))

		dsksA := a.GetInstaceDisks(instId)
		dsksB := b.GetInstaceDisks(instId)
		c.list("GetInstaceDisks", instId.Hex(), dsksA, dsksB)

		dskIds := idUnion{}
		for _, dsk := range dsksA {
			if dsk != nil {
				dskIds.add(dsk.Id)
			}
		}
		for _, dsk := range dsksB {
			if dsk != nil {
				dskIds.add(dsk.Id)
			}
		}
		for _, dskId := range dskIds.sorted() {
			c.values("DiskInUse", instId.Hex()+":"+dskId.Hex(),
				a.DiskInUse(instId, dskId), b.DiskInUse(instId, dskId))
		}

		c.list("GetInstaceAuthorities", instId.Hex(),
			a.GetInstaceAuthorities(inst.Organization, inst.Roles),
			b.GetInstaceAuthorities(inst.Organization, inst.Roles))

		c.list("GetInstanceNamespaces", instId.Hex(),
			a.GetInstanceNamespaces(instId), b.GetInstanceNamespaces(instId))
	}

	c.list("NodeFirewall", "", a.NodeFirewall(), b.NodeFirewall())
	c.mapping("Firewalls", a.Firewalls(), b.Firewalls(), true)
	c.mapping("FirewallMaps", a.FirewallMaps(), b.FirewallMaps(), true)

	sortedNamespaces := make([]string, 0, len(namespaces))
	for ns := range namespaces {
		sortedNamespaces = append(sortedNamespaces, ns)
	}
	sort.Strings(sortedNamespaces)
	for _, ns := range sortedNamespaces {
		c.values("ArpRecords", ns, a.ArpRecords(ns), b.ArpRecords(ns))
	}

	diffs = c.diffs

	return
}

func Summarize(diffs []*Diff, limit int) (lines []string) {
	lines = []string{}
	for i, diff := range diffs {
		if i >= limit {
			lines = append(lines, fmt.Sprintf("... %d more",
				len(diffs)-limit))
			break
		}
		lines = append(lines, diff.String())
	}
	return
}

func JoinDiffs(diffs []*Diff, limit int) string {
	return strings.Join(Summarize(diffs, limit), "; ")
}

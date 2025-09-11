#!/usr/bin/env python3
import random
import string
import hashlib

ORGANIZATION_ID = "5a3245a50accad1a8a53bc82"
DATACENTER_ID = "689733b7a7a35eae0dbaea1b"
ZONE_ID = "689733b7a7a35eae0dbaea1e"
VPC_ID = "689733b7a7a35eae0dbaea23"

NODE_IDS = [
    "689733b2a7a35eae0dbaea0a",
    "689733b2a7a35eae0dbaea0b",
    "689733b2a7a35eae0dbaea0c",
    "689733b2a7a35eae0dbaea0d",
    "689733b2a7a35eae0dbaea0e",
    "689733b2a7a35eae0dbaea0f",
]
NODE_NAMES = [
    "pritunl-east0",
    "pritunl-east1",
    "pritunl-east2",
    "pritunl-east3",
    "pritunl-east4",
    "pritunl-east5",
]

SHAPE_IDS = {
    "small": "65e6e303ceeebbb3dabaec96",
    "medium": "65e6e2ecceeebbb3dabaec79",
    "large": "66f63282aac06d53e8c9c435",
}

IMAGE_IDS = [
    "650a2c36aed15f1f1f5e96e1",
    "650a2c36aed15f1f1f5e96e2",
]

POD_ID = "688bf358d978631566998ffc"
UNIT_IDS = {
    "web": "688c716d9da165ffad4b3682",
    "database": "68b67d1aee12c08a1f39f88b",
}
SPEC_IDS = {
    "web": "688c7cde9da165ffad4b52e4",
    "database": "688c7cde9da165ffad4b34f2",
}

def generate_ip(subnet_base="10.196"):
    third_octet = random.randint(1, 8)
    fourth_octet = random.randint(2, 254)
    return f"{subnet_base}.{third_octet}.{fourth_octet}"

def generate_pub_ip(subnet_base="1.253.67"):
    fourth_octet = random.randint(2, 254)
    return f"{subnet_base}.{fourth_octet}"

def generate_priv_ip6(id):
    hash_obj = hashlib.sha256(str(id).encode())
    hash_hex = hash_obj.hexdigest()
    return f"fd97:30bf:d456:a3bc:{hash_hex[0:4]}:{hash_hex[4:8]}:{hash_hex[8:12]}:{hash_hex[12:16]}"

def generate_pub_ip6(id):
    hash_obj = hashlib.sha256(str(id).encode() + str(id).encode())
    hash_hex = hash_obj.hexdigest()
    return f"2001:db8:85a3:4d2f:{hash_hex[0:4]}:{hash_hex[4:8]}:{hash_hex[8:12]}:{hash_hex[12:16]}"

def generate_network_namespace():
    characters = string.ascii_lowercase + string.digits
    return ''.join(random.choice(characters) for _ in range(14))

def get_instance_spec(instance_type):
    specs = {
        "web": {"name": "web-app", "shape": "small", "memory": 2048, "processors": 2, "disk": 20},
        "database": {"name": "database", "shape": "large", "memory": 8192, "processors": 4, "disk": 100},
        "search": {"name": "search", "shape": "large", "memory": 8192, "processors": 4, "disk": 200},
        "vpn": {"name": "vpn", "shape": "small", "memory": 2048, "processors": 2, "disk": 20},
    }
    return specs.get(instance_type)

def generate_instances(count=20):
    instances = []
    used_priv_ips = set()
    used_host_ips = set()
    used_pub_ips = set()

    instance_types = (
        ["web"] * 10 +
        ["database"] * 6 +
        ["search"] * 2 +
        ["vpn"] * 2
    )[:count]

    for i in range(count):
        instance_id = f"651d8e7c4cf9e2e3e4d56a{i:02x}"

        priv_ip = generate_ip()
        while priv_ip in used_priv_ips:
            priv_ip = generate_ip()
        used_priv_ips.add(priv_ip)

        pub_ip = generate_pub_ip()
        while pub_ip in used_pub_ips:
            pub_ip = generate_pub_ip()
        used_pub_ips.add(pub_ip)

        host_ip = generate_pub_ip("198.18.84")
        while host_ip in used_host_ips:
            host_ip = generate_pub_ip("198.18.84")
        used_host_ips.add(host_ip)

        instance_type = instance_types[i]
        spec = get_instance_spec(instance_type)

        node_id = NODE_IDS[i % len(NODE_IDS)]
        node_name = NODE_NAMES[i % len(NODE_NAMES)]
        image_id = IMAGE_IDS[i % len(IMAGE_IDS)]

        load1 = round(random.uniform(10, 60), 2)
        load5 = round(load1 + random.uniform(1, 10), 2)
        load15 = round(load5 + random.uniform(1, 10), 2)

        instance = {
            "id": instance_id,
            "type": instance_type,
            "organization": ORGANIZATION_ID,
            "datacenter": DATACENTER_ID,
            "zone": ZONE_ID,
            "vpc": VPC_ID,
            "image": image_id,
            "image_backing": False,
            "status": "Running",
            "state": "running",
            "action": "start",
            "public_ips": [pub_ip],
            "public_ips6": [generate_pub_ip6(instance_id)],
            "private_ips": [priv_ip],
            "private_ips6": [generate_priv_ip6(instance_id)],
            "host_ips": [host_ip],
            "node": node_id,
            "node_name": node_name,
            "shape": SHAPE_IDS[spec["shape"]],
            "name": spec["name"],
            "comment": "",
            "init_disk_size": spec["disk"],
            "memory": spec["memory"],
            "processors": spec["processors"],
            "network_namespace": generate_network_namespace(),
            "mem": round(random.uniform(30, 80), 2),
            "load1": load1,
            "load5": load5,
            "load15": load15,
        }

        instances.append(instance)

    return instances

def generate_disks(instances):
    disks = []

    for i, instance in enumerate(instances):
        disk_id = f"651d8e7c4cf9e2e3e4d34f{i:02x}"

        disk = {
            "id": disk_id,
            "name": instance['name'],
            "comment": "",
            "state": "attached",
            "type": "qcow2",
            "datacenter": instance["datacenter"],
            "zone": instance["zone"],
            "node": instance["node"],
            "organization": instance["organization"],
            "instance": instance["id"],
            "image": instance["image"],
            "index": "0",
            "size": instance["init_disk_size"],
        }

        disks.append(disk)

    return disks

def generate_deployments(instances):
    deployments = []

    for i, instance in enumerate(instances):
        if instance["type"] != "web" and instance["type"] != "database":
            continue

        deployment_id = f"651d8e7c4cf91e3b53d62d{i:02x}"

        deployment = {
            "id": deployment_id,
            "name": instance['name'],
            "type": instance['type'],
            "datacenter": instance["datacenter"],
            "zone": instance["zone"],
            "node": instance["node"],
            "node_name": instance["node_name"],
            "organization": instance["organization"],
            "instance": instance["id"],
            "public_ips": instance["public_ips"],
            "public_ips6": instance["public_ips6"],
            "private_ips": instance["private_ips"],
            "private_ips6": instance["private_ips6"],
            "host_ips": instance["host_ips"],
            "memory": instance["memory"],
            "processors": instance["processors"],
            "mem": instance["mem"],
            "load1": instance["load1"],
            "load5": instance["load5"],
            "load15": instance["load15"],
        }

        deployments.append(deployment)

    return deployments

def format_go_instance(instance):
    go_code = f"""	{{
		Id:               utils.ObjectIdHex("{instance['id']}"),
		Organization:     utils.ObjectIdHex("{instance['organization']}"),
		Datacenter:       utils.ObjectIdHex("{instance['datacenter']}"),
		Zone:             utils.ObjectIdHex("{instance['zone']}"),
		Vpc:              utils.ObjectIdHex("{instance['vpc']}"),
		Image:            utils.ObjectIdHex("{instance['image']}"),
		ImageBacking:     {str(instance['image_backing']).lower()},
		Status:           "{instance['status']}",
		State:            "{instance['state']}",
		Action:           "{instance['action']}",
		Uptime:           "5 days 11 hours 34 mins",
		PublicIps:        []string{{"{instance['public_ips'][0]}"}},
		PublicIps6:       []string{{"{instance['public_ips6'][0]}"}},
		PrivateIps:       []string{{"{instance['private_ips'][0]}"}},
		PrivateIps6:      []string{{"{instance['private_ips6'][0]}"}},
		HostIps:          []string{{"{instance['host_ips'][0]}"}},
		Node:             utils.ObjectIdHex("{instance['node']}"),
		Shape:            utils.ObjectIdHex("{instance['shape']}"),
		Name:             "{instance['name']}",
		Comment:          "",
		InitDiskSize:     {instance['init_disk_size']},
		Memory:           {instance['memory']},
		Processors:       {instance['processors']},
		NetworkNamespace: "{instance['network_namespace']}",
		Created:          time.Now(),
		Timestamp:        time.Now(),
		Guest:            &instance.GuestData{{
			Status:     "running",
			Timestamp:  time.Now(),
			Heartbeat:  time.Now(),
			Memory:     {instance['mem']},
			Load1:      {instance['load1']},
			Load5:      {instance['load5']},
			Load15:     {instance['load15']},
		}},
	}},"""
    return go_code

def format_go_disk(disk):
    go_code = f"""	{{
		Disk: disk.Disk{{
			Id:           utils.ObjectIdHex("{disk['id']}"),
			Name:         "{disk['name']}",
			Comment:      "",
			State:        "{disk['state']}",
			Type:         "{disk['type']}",
			Datacenter:   utils.ObjectIdHex("{disk['datacenter']}"),
			Zone:         utils.ObjectIdHex("{disk['zone']}"),
			Node:         utils.ObjectIdHex("{disk['node']}"),
			Organization: utils.ObjectIdHex("{disk['organization']}"),
			Instance:     utils.ObjectIdHex("{disk['instance']}"),
			Image:        utils.ObjectIdHex("{disk['image']}"),
			Index:        "{disk['index']}",
			Size:         {disk['size']},
			Created:      time.Now(),
		}},
	}},"""
    return go_code

def format_go_deployment(deployment):
    go_code = f"""	{{
		Id:            utils.ObjectIdHex("{deployment['id']}"),
		Pod:           utils.ObjectIdHex("{POD_ID}"),
		Unit:          utils.ObjectIdHex("{UNIT_IDS[deployment['type']]}"),
		Spec:          utils.ObjectIdHex("{SPEC_IDS[deployment['type']]}"),
		SpecOffset:    0,
		SpecIndex:     2,
		SpecTimestamp: time.Now(),
		Timestamp:     time.Now(),
		Tags:          []string{{}},
		Kind:          "instance",
		State:         "deployed",
		Action:        "",
		Status:        "healthy",
		Node:          utils.ObjectIdHex("{deployment['node']}"),
		Instance:      utils.ObjectIdHex("{deployment['instance']}"),
		InstanceData: &deployment.InstanceData{{
			HostIps:     []string{{"{deployment['host_ips'][0]}"}},
			PublicIps:   []string{{"{deployment['public_ips'][0]}"}},
			PublicIps6:  []string{{"{deployment['public_ips6'][0]}"}},
			PrivateIps:  []string{{"{deployment['private_ips'][0]}"}},
			PrivateIps6: []string{{"{deployment['private_ips6'][0]}"}},
		}},
		ZoneName:            "us-west-1a",
		NodeName:            "{deployment['node_name']}",
		InstanceName:        "{deployment['name']}",
		InstanceRoles:       []string{{"instance"}},
		InstanceMemory:      {deployment['memory']},
		InstanceProcessors:  {deployment['processors']},
		InstanceStatus:      "Running",
		InstanceUptime:      "5 days",
		InstanceState:       "running",
		InstanceAction:      "start",
		InstanceGuestStatus: "running",
		InstanceTimestamp:   time.Now(),
		InstanceHeartbeat:   time.Now(),
		InstanceMemoryUsage: {deployment['mem']},
		InstanceHugePages:   0,
		InstanceLoad1:       {deployment['load1']},
		InstanceLoad5:       {deployment['load5']},
		InstanceLoad15:      {deployment['load15']},
	}},"""
    return go_code

def main():
    instances = generate_instances(20)
    disks = generate_disks(instances)
    deployments = generate_deployments(instances)

    print("// Instances")
    print("var Instances = []*instance.Instance{")
    for i, instance in enumerate(instances):
        print(format_go_instance(instance))
    print("}")
    print("")
    print("// Disks")
    print("var Disks = []*aggregate.DiskAggregate{")
    for i, disk in enumerate(disks):
        print(format_go_disk(disk))
    print("}")
    print("")
    print("// Deployments")
    print("var Deployments = []*aggregate.Deployment{")
    for i, deployment in enumerate(deployments):
        print(format_go_deployment(deployment))
    print("}")


if __name__ == "__main__":
    main()

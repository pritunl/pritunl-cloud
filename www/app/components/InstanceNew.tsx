/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Constants from '../Constants';
import * as License from '../License';
import * as InstanceTypes from '../types/InstanceTypes';
import * as OrganizationTypes from '../types/OrganizationTypes';
import * as DomainTypes from '../types/DomainTypes';
import * as DatacenterTypes from '../types/DatacenterTypes';
import * as NodeTypes from '../types/NodeTypes';
import * as VpcTypes from '../types/VpcTypes';
import * as ImageTypes from '../types/ImageTypes';
import * as PoolTypes from '../types/PoolTypes';
import * as ZoneTypes from '../types/ZoneTypes';
import * as ShapeTypes from '../types/ShapeTypes';
import * as InstanceActions from '../actions/InstanceActions';
import * as ImageActions from '../actions/ImageActions';
import * as NodeActions from '../actions/NodeActions';
import ImagesDatacenterStore from '../stores/ImagesDatacenterStore';
import NodesZoneStore from '../stores/NodesZoneStore';
import InstanceImages from './InstanceImages';
import InstanceLicense from './InstanceLicense';
import InstanceNodePort from './InstanceNodePort';
import InstanceMount from './InstanceMount';
import PageInput from './PageInput';
import PageInputButton from './PageInputButton';
import PageCreate from './PageCreate';
import PageSelect from './PageSelect';
import PageSwitch from "./PageSwitch";
import PageNumInput from './PageNumInput';
import Help from './Help';
import CompletionStore from "../stores/CompletionStore";
import PageTextArea from "./PageTextArea";

interface Props {
	organizations: OrganizationTypes.OrganizationsRo;
	vpcs: VpcTypes.VpcsRo;
	domains: DomainTypes.DomainsRo;
	datacenters: DatacenterTypes.DatacentersRo;
	pools: PoolTypes.PoolsRo;
	zones: ZoneTypes.ZonesRo;
	shapes: ShapeTypes.ShapesRo;
	onClose: () => void;
}

interface State {
	closed: boolean;
	disabled: boolean;
	changed: boolean;
	message: string;
	licenseOpen: boolean;
	instance: InstanceTypes.Instance;
	datacenter: string;
	image: ImageTypes.Image;
	images: ImageTypes.ImagesRo;
	nodes: NodeTypes.NodesRo;
	addNetworkRole: string;
	addVpc: string;
	dhcpChanged: boolean;
	secureBootChanged: boolean;
	hiddenImages: boolean;
}

const css = {
	row: {
		display: 'table-row',
		width: '100%',
		padding: 0,
		boxShadow: 'none',
		position: 'relative',
	} as React.CSSProperties,
	card: {
		position: 'relative',
		padding: '10px 10px 0 10px',
		width: '100%',
	} as React.CSSProperties,
	buttons: {
		position: 'absolute',
		top: '5px',
		right: '5px',
	} as React.CSSProperties,
	item: {
		margin: '9px 5px 0 5px',
		minHeight: '20px',
	} as React.CSSProperties,
	itemsLabel: {
		display: 'block',
	} as React.CSSProperties,
	itemsAdd: {
		margin: '8px 0 15px 0',
	} as React.CSSProperties,
	group: {
		flex: 1,
		minWidth: '280px',
		margin: '0 10px',
	} as React.CSSProperties,
	save: {
		paddingBottom: '10px',
	} as React.CSSProperties,
	label: {
		width: '100%',
		maxWidth: '280px',
	} as React.CSSProperties,
	inputGroup: {
		width: '100%',
	} as React.CSSProperties,
	protocol: {
		flex: '0 1 auto',
	} as React.CSSProperties,
	port: {
		flex: '1',
	} as React.CSSProperties,
	role: {
		margin: '9px 5px 0 5px',
		minHeight: '20px',
	} as React.CSSProperties,
};

export default class InstanceNew extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			closed: false,
			disabled: false,
			changed: false,
			message: '',
			licenseOpen: false,
			instance: this.default,
			datacenter: '',
			image: null,
			images: [],
			nodes: [],
			addNetworkRole: '',
			addVpc: '',
			dhcpChanged: false,
			secureBootChanged: false,
			hiddenImages: false,
		};
	}

	componentDidMount(): void {
		ImagesDatacenterStore.addChangeListener(this.onChange);
		NodesZoneStore.addChangeListener(this.onChange);
	}

	componentWillUnmount(): void {
		ImagesDatacenterStore.removeChangeListener(this.onChange);
		NodesZoneStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		this.setState({
			...this.state,
			images: ImagesDatacenterStore.images,
			nodes: NodesZoneStore.nodes,
		});
	}

	get default(): InstanceTypes.Instance {
		return {
			id: null,
			name: 'new-instance',
			uefi: true,
			secure_boot: true,
			init_disk_size: 10,
			memory: 1024,
			processors: 1,
			count: 1,
		};
	}

	set(name: string, val: any): void {
		let instance: any = {
			...this.state.instance,
		};

		instance[name] = val;

		this.setState({
			...this.state,
			changed: true,
			instance: instance,
		});
	}

	onCreate = (): void => {
		// if (this.state.instance.image &&
		// 		this.state.instance.image.indexOf('oracle') &&
		// 		!License.oracle) {
		// 	this.setState({
		// 		...this.state,
		// 		licenseOpen: true,
		// 	});
		// 	return
		// }

		this.setState({
			...this.state,
			disabled: true,
		});

		let instance: any = {
			...this.state.instance,
		};

		if (this.props.organizations.length && !instance.organization) {
			instance.organization = this.props.organizations[0].id;
		}

		InstanceActions.create(instance).then((): void => {
			this.setState({
				...this.state,
				message: 'Instance created successfully',
				changed: false,
			});

			setTimeout((): void => {
				this.setState({
					...this.state,
					disabled: false,
					changed: true,
				});
			}, 2000);
		}).catch((): void => {
			this.setState({
				...this.state,
				message: '',
				disabled: false,
			});
		});
	}

	onUefi(uefi: boolean): void {
		let instance: any = {
			...this.state.instance,
		};

		let img = this.state.image
		if (instance.uefi !== uefi) {
			img = null
		}
		instance.uefi = uefi;

		this.setState({
			...this.state,
			changed: true,
			instance: instance,
			image: img,
		});
	}

	onSecureBoot(secureBoot: boolean): void {
		let instance: InstanceTypes.Instance = {
			...this.state.instance,
		};

		let img = this.state.image
		if (instance.secure_boot !== secureBoot) {
			img = null
		}
		instance.secure_boot = secureBoot;

		this.setState({
			...this.state,
			changed: true,
			instance: instance,
			image: img,
			secureBootChanged: true,
		});
	}

	onDhcpServer(dhcpServer: boolean): void {
		let instance: InstanceTypes.Instance = {
			...this.state.instance,
		};

		instance.dhcp_server = dhcpServer;

		this.setState({
			...this.state,
			changed: true,
			instance: instance,
			dhcpChanged: true,
		});
	}

	onNode(nodeId: string): void {
		let instance: InstanceTypes.Instance = {
			...this.state.instance,
		};

		instance.shape = "";
		instance.node = nodeId;

		let node = NodesZoneStore.node(instance.node);
		if (node) {
			instance.no_public_address = node.default_no_public_address;
			instance.no_public_address6 = node.default_no_public_address6;
		}

		this.setState({
			...this.state,
			changed: true,
			instance: instance,
		});
	}

	onShape(shapeId: string): void {
		let instance: InstanceTypes.Instance = {
			...this.state.instance,
		};

		instance.node = "";
		instance.shape = shapeId;

		let shape = CompletionStore.shape(shapeId);
		if (shape) {
			instance.disk_type = shape.disk_type;
			instance.disk_pool = shape.disk_pool;
			instance.processors = shape.processors;
			instance.memory = shape.memory;
		}

		this.setState({
			...this.state,
			changed: true,
			instance: instance,
		});
	}

	onAddNetworkRole = (): void => {
		if (!this.state.addNetworkRole) {
			return;
		}

		let instance = {
			...this.state.instance,
		};

		let networkRoles = [
			...(instance.roles || []),
		];

		if (networkRoles.indexOf(this.state.addNetworkRole) === -1) {
			networkRoles.push(this.state.addNetworkRole);
		}

		networkRoles.sort();
		instance.roles = networkRoles;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addNetworkRole: '',
			instance: instance,
		});
	}

	onRemoveNetworkRole = (networkRole: string): void => {
		let instance = {
			...this.state.instance,
		};

		let networkRoles = [
			...(instance.roles || []),
		];

		let i = networkRoles.indexOf(networkRole);
		if (i === -1) {
			return;
		}

		networkRoles.splice(i, 1);
		instance.roles = networkRoles;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addNetworkRole: '',
			instance: instance,
		});
	}

	onAddNodePort = (): void => {
		let instance = {
			...this.state.instance,
		};

		let nodePorts = [
			...(instance.node_ports || []),
			{
				protocol: "tcp",
				external_port: 0,
				internal_port: 0,
			},
		];

		instance.node_ports = nodePorts;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			instance: instance,
		});
	}

	onChangeNodePort(i: number, state: InstanceTypes.NodePort): void {
		let instance = {
			...this.state.instance,
		};

		let nodePorts = [
			...instance.node_ports,
		];

		nodePorts[i] = state;

		instance.node_ports = nodePorts;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			instance: instance,
		});
	}

	onRemoveNodePort(i: number): void {
		let instance = {
			...this.state.instance,
		};

		let nodePorts = [
			...instance.node_ports,
		];

		nodePorts[i] = {
			...nodePorts[i],
			delete: true,
		};

		instance.node_ports = nodePorts;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			instance: instance,
		});
	}

	onAddMount = (i: number): void => {
		let instance = {
			...this.state.instance,
		};

		let mounts = [
			...(instance.mounts || []),
		];
		if (!mounts.length) {
			mounts.push({})
		}

		mounts.splice(i + 1, 0, {});
		instance.mounts = mounts;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			instance: instance,
		});
	}

	onChangeMount(i: number, block: InstanceTypes.Mount): void {
		let instance = {
			...this.state.instance,
		};

		let mounts = [
			...(instance.mounts || []),
		];
		if (!mounts.length) {
			mounts.push({})
		}

		mounts[i] = block;

		instance.mounts = mounts;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			instance: instance,
		});
	}

	onRemoveMount(i: number): void {
		let instance = {
			...this.state.instance,
		};

		let mounts = [
			...(instance.mounts || []),
		];
		if (!mounts.length) {
			mounts.push({})
		}

		mounts.splice(i, 1);

		if (!mounts.length) {
			mounts = [
				{},
			];
		}

		instance.mounts = mounts;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			instance: instance,
		});
	}

	onImageselect = (img: ImageTypes.Image): void => {
		let instance: any = {
			...this.state.instance,
		};

		instance.image = img.id;
		if (img.name.toLowerCase().indexOf('bsd') !== -1) {
			instance.secure_boot = false;
			instance.cloud_type = 'bsd';
		} else if (img.name.toLowerCase().indexOf('alpinelinux') !== -1) {
			instance.secure_boot = false;
			instance.cloud_type = 'linux';
		} else {
			if (!this.state.secureBootChanged) {
				instance.secure_boot = true;
			}
			instance.cloud_type = 'linux';
		}

		this.setState({
			...this.state,
			changed: true,
			image: img,
			instance: instance,
		});
	}

	render(): JSX.Element {
		let instance = this.state.instance;
		let lockResources = false;
		if (instance.shape) {
			let shape = CompletionStore.shape(instance.shape);
			lockResources = !shape?.flexible;
		}

		let node: NodeTypes.Node;
		if (instance.node) {
			node = NodesZoneStore.node(instance.node);
		}

		let hasOrganizations = !!this.props.organizations.length;
		let organizationsSelect: JSX.Element[] = [];
		if (this.props.organizations && this.props.organizations.length) {
			organizationsSelect.push(
				<option key="null" value="">Select Organization</option>);

			for (let organization of this.props.organizations) {
				organizationsSelect.push(
					<option
						key={organization.id}
						value={organization.id}
					>{organization.name}</option>,
				);
			}
		}

		if (!hasOrganizations) {
			organizationsSelect.push(
				<option key="null" value="">No Organizations</option>);
		}

		let hasDatacenters = false;
		let datacentersSelect: JSX.Element[] = [];
		if (this.props.datacenters && this.props.datacenters.length) {
			datacentersSelect.push(
				<option key="null" value="">Select Datacenter</option>);

			hasDatacenters = true;
			for (let datacenter of this.props.datacenters) {
				datacentersSelect.push(
					<option
						key={datacenter.id}
						value={datacenter.id}
					>{datacenter.name}</option>,
				);
			}
		}

		if (!hasDatacenters) {
			datacentersSelect.push(
				<option key="null" value="">No Datacenters</option>);
		}

		let datacenter = this.state.datacenter;
		let hasZones = false;
		let zonesSelect: JSX.Element[] = [];
		if (this.props.zones && this.props.zones.length) {
			zonesSelect.push(<option key="null" value="">Select Zone</option>);

			for (let zone of this.props.zones) {
				if (zone.datacenter !== datacenter) {
					continue;
				}
				hasZones = true;

				zonesSelect.push(
					<option
						key={zone.id}
						value={zone.id}
					>{zone.name}</option>,
				);
			}
		}

		if (!hasZones) {
			zonesSelect = [<option key="null" value="">No Zones</option>];
		}

		let hasShapes = false;
		let shapesSelect: JSX.Element[] = [];
		if (this.props.shapes && this.props.shapes.length) {
			shapesSelect.push(<option key="null" value="">Select Shape</option>);

			for (let shape of this.props.shapes) {
				if (shape.datacenter !== datacenter) {
					continue;
				}
				hasShapes = true;

				shapesSelect.push(
					<option
						key={shape.id}
						value={shape.id}
					>{shape.name}</option>,
				);
			}
		}

		if (!hasShapes) {
			shapesSelect = [<option key="null" value="">No Shapes</option>];
		}

		let instanceMounts = instance.mounts || [];
		let mounts: JSX.Element[] = [];
		if (instanceMounts.length === 0) {
			instanceMounts.push({});
		}
		for (let i = 0; i < instanceMounts.length; i++) {
			let index = i;

			mounts.push(
				<InstanceMount
					key={index}
					disabled={this.state.disabled}
					mount={instanceMounts[index]}
					onChange={(state: InstanceTypes.Mount): void => {
						this.onChangeMount(index, state);
					}}
					onAdd={(): void => {
						this.onAddMount(index);
					}}
					onRemove={(): void => {
						this.onRemoveMount(index);
					}}
				/>,
			);
		}

		let hasNodes = false;
		let nodesSelect: JSX.Element[] = [];
		if (this.state.nodes && this.state.nodes.length) {
			nodesSelect.push(<option key="null" value="">Select Node</option>);

			hasNodes = true;
			for (let node of this.state.nodes) {
				nodesSelect.push(
					<option
						key={node.id}
						value={node.id}
					>{node.name}</option>,
				);
			}
		}

		if (!hasNodes) {
			nodesSelect = [<option key="null" value="">No Nodes</option>];
		}

		let hasVpcs = false;
		let vpcsSelect: JSX.Element[] = [];
		if (this.props.vpcs && this.props.vpcs.length) {
			vpcsSelect.push(<option key="null" value="">Select Vpc</option>);

			for (let vpc of this.props.vpcs) {
				if (Constants.user) {
					if (vpc.organization !== CompletionStore.userOrganization) {
						continue;
					}
				} else {
					if (vpc.organization != instance.organization) {
						continue;
					}
				}

				if (vpc.datacenter !== this.state.datacenter) {
					continue;
				}

				hasVpcs = true;
				vpcsSelect.push(
					<option
						key={vpc.id}
						value={vpc.id}
					>{vpc.name}</option>,
				);
			}
		}

		if (!hasVpcs) {
			vpcsSelect = [<option key="null" value="">No Vpcs</option>];
		}

		let hasSubnets = false;
		let subnetSelect: JSX.Element[] = [];
		if (this.props.vpcs && this.props.vpcs.length) {
			subnetSelect.push(<option key="null" value="">Select Subnet</option>);

			for (let vpc of this.props.vpcs) {
				if (Constants.user) {
					if (vpc.organization !== CompletionStore.userOrganization) {
						continue;
					}
				} else {
					if (vpc.organization != instance.organization) {
						continue;
					}
				}

				if (vpc.datacenter !== this.state.datacenter) {
					continue;
				}

				if (vpc.id === instance.vpc) {
					for (let sub of (vpc.subnets || [])) {
						hasSubnets = true;
						subnetSelect.push(
							<option
								key={sub.id}
								value={sub.id}
							>{sub.name + ' - ' + sub.network}</option>,
						);
					}
				}
			}
		}

		if (!hasSubnets) {
			subnetSelect = [<option key="null" value="">No Subnets</option>];
		}

		let cloudSubnetsSelect: JSX.Element[] = [
			<option key="null" value="">Disabled</option>,
		];
		if (node && node.cloud_subnets && node.cloud_subnets.length) {
			let subnets: Map<string, string> = new Map();

			for (let vpc of (node.available_vpcs || [])) {
				for (let subnet of (vpc.subnets || [])) {
					subnets.set(subnet.id, vpc.name + ' - ' + subnet.name);
				}
			}

			for (let subnetId of (node.cloud_subnets || [])) {
				cloudSubnetsSelect.push(
					<option key={subnetId} value={subnetId}>
						{subnets.get(subnetId) || subnetId}
					</option>,
				);
			}
		}

		let domainsSelect: JSX.Element[] = [
			<option key="null" value="">No Domain</option>,
		];
		if (this.props.domains && this.props.domains.length) {
			for (let domain of this.props.domains) {
				if (!Constants.user &&
						domain.organization != instance.organization) {
					continue;
				}

				domainsSelect.push(
					<option
						key={domain.id}
						value={domain.id}
					>{domain.name}</option>,
				);
			}
		}

		let hasPools = false;
		let poolsSelect: JSX.Element[] = [];
		if (this.props.pools.length) {
			poolsSelect.push(<option key="null" value="">Select Pool</option>);

			hasPools = true;
			for (let pool of this.props.pools) {
				if (!node || !node.pools || node.pools.indexOf(pool.id) === -1) {
					continue
				}

				poolsSelect.push(
					<option
						key={pool.id}
						value={pool.id}
					>{pool.name}</option>,
				);
			}
		}

		if (!hasPools) {
			poolsSelect = [<option key="null" value="">No Pools</option>];
		}

		let networkRoles: JSX.Element[] = [];
		for (let networkRole of (instance.roles || [])) {
			networkRoles.push(
				<div
					className="bp5-tag bp5-tag-removable bp5-intent-primary"
					style={css.role}
					key={networkRole}
				>
					{networkRole}
					<button
						className="bp5-tag-remove"
						disabled={this.state.disabled}
						onMouseUp={(): void => {
							this.onRemoveNetworkRole(networkRole);
						}}
					/>
				</div>,
			);
		}

		let nodePorts: JSX.Element[] = [];
		(instance.node_ports || []).forEach((nodePort, index) => {
			if (nodePort.delete) {
				return
			}

			nodePorts.push(
				<InstanceNodePort
					key={index}
					nodePort={nodePort}
					onChange={(state: InstanceTypes.NodePort): void => {
						this.onChangeNodePort(index, state);
					}}
					onRemove={(): void => {
						this.onRemoveNodePort(index);
					}}
				/>,
			);
		})

		return <div
			className="bp5-card bp5-row"
			style={css.row}
		>
			<td
				className="bp5-cell"
				colSpan={6}
				style={css.card}
			>
				<div className="layout horizontal wrap">
					<InstanceLicense
						open={this.state.licenseOpen}
						onClose={(): void => {
							this.setState({
								...this.state,
								licenseOpen: false,
							});
						}}
					/>
					<div style={css.group}>
						<div style={css.buttons}>
						</div>
						<PageInput
							label="Name"
							help="Name of instance. String formatting such as %d or %02d can be used to add the instance number or zero padded number."
							type="text"
							placeholder="Enter name"
							disabled={this.state.disabled}
							value={instance.name}
							onChange={(val): void => {
								this.set('name', val);
							}}
						/>
						<PageTextArea
							label="Comment"
							help="Instance comment."
							placeholder="Instance comment"
							rows={3}
							value={instance.comment}
							onChange={(val: string): void => {
								this.set('comment', val);
							}}
						/>
						<PageSelect
							disabled={this.state.disabled || !hasOrganizations}
							hidden={Constants.user}
							label="Organization"
							help="Organization for instance."
							value={instance.organization}
							onChange={(val): void => {
								this.set('organization', val);
							}}
						>
							{organizationsSelect}
						</PageSelect>
						<PageSelect
							disabled={this.state.disabled || !hasDatacenters}
							label="Datacenter"
							help="Datacenter for instance."
							value={this.state.datacenter}
							onChange={(val): void => {
								this.setState({
									...this.state,
									datacenter: val,
									instance: {
										...this.state.instance,
										node: '',
										zone: '',
										image: '',
									},
									image: null,
								});
								ImageActions.syncDatacenter(val);
								NodeActions.syncZone('');
							}}
						>
							{datacentersSelect}
						</PageSelect>
						<PageSelect
							disabled={this.state.disabled || !hasZones}
							label="Zone"
							help="Zone for instance."
							value={instance.zone}
							onChange={(val): void => {
								this.setState({
									...this.state,
									instance: {
										...this.state.instance,
										node: '',
										zone: val,
									},
								});
								NodeActions.syncZone(val);
							}}
						>
							{zonesSelect}
						</PageSelect>
						<PageSelect
							disabled={this.state.disabled || !hasShapes || !!instance.node}
							label="Shape"
							help="Instance shape to run instance on, node will be automatically selected."
							value={instance.shape}
							onChange={(val): void => {
								this.onShape(val);
							}}
						>
							{shapesSelect}
						</PageSelect>
						<PageSelect
							disabled={this.state.disabled || !hasNodes || !!instance.shape}
							label="Node"
							help="Node to run instance on."
							value={instance.node}
							onChange={(val): void => {
								this.onNode(val);
							}}
						>
							{nodesSelect}
						</PageSelect>
						<PageSelect
							disabled={this.state.disabled || !hasVpcs}
							label="VPC"
							help="VPC for instance."
							value={instance.vpc}
							onChange={(val): void => {
								this.setState({
									...this.state,
									instance: {
										...this.state.instance,
										vpc: val,
									},
								});
							}}
						>
							{vpcsSelect}
						</PageSelect>
						<PageSelect
							disabled={this.state.disabled || !hasVpcs}
							label="Subnet"
							help="Subnet for instance."
							value={instance.subnet}
							onChange={(val): void => {
								this.setState({
									...this.state,
									instance: {
										...this.state.instance,
										subnet: val,
									},
								});
							}}
						>
							{subnetSelect}
						</PageSelect>
						<PageSelect
							disabled={this.state.disabled}
							hidden={cloudSubnetsSelect.length <= 1}
							label="Oracle Cloud Subnet"
							help="Oracle Cloud subnet for instance."
							value={instance.cloud_subnet}
							onChange={(val): void => {
								this.setState({
									...this.state,
									instance: {
										...this.state.instance,
										cloud_subnet: val,
									},
								});
							}}
						>
							{cloudSubnetsSelect}
						</PageSelect>
						<PageSelect
							disabled={this.state.disabled}
							label="CloudInit Type"
							help="Target operating system for cloud init"
							value={instance.cloud_type}
							onChange={(val): void => {
								this.set('cloud_type', val);
							}}
						>
							<option key="linux" value="linux">Linux</option>,
							<option key="bsd" value="bsd">BSD</option>,
							<option key="linux_legacy" value="linux_legacy">Linux Legacy</option>,
						</PageSelect>
						<label
							className="bp5-label"
							style={css.label}
						>
							Host Paths
							<Help
								title="Host Paths"
								content="Local paths on the host that are available for instances to access through VirtIO-FS sharing. The path must be match or be a subdirectory of a configured host share path in the node settings. The instance's organization must also have a matching role to access the host share."
							/>
						</label>
						<div>
							{mounts}
						</div>
						<PageTextArea
							label="Startup Script"
							help="Script to run on instance startup. These commands will run on every startup. File must start with #! such as `#!/bin/bash` to specify code interpreter."
							placeholder="Startup script"
							rows={3}
							value={instance.cloud_script}
							onChange={(val: string): void => {
								this.set('cloud_script', val);
							}}
						/>
						<PageSwitch
							disabled={this.state.disabled}
							label="Start instance"
							help="Automatically start instance. Disable to get the public MAC address before instance is started for first time."
							checked={!instance.action}
							onToggle={(): void => {
								if (instance.action === 'stop') {
									this.set('action', '');
								} else {
									this.set('action', 'stop');
								}
							}}
						/>
						<PageSwitch
							disabled={this.state.disabled}
							label="VNC server"
							help="Enable VNC server for remote control of instance."
							checked={instance.vnc}
							onToggle={(): void => {
								this.set('vnc', !instance.vnc);
							}}
						/>
						<PageSwitch
							disabled={this.state.disabled}
							label="Spice server"
							help="Enable Spice server for remote control of instance."
							checked={instance.spice}
							onToggle={(): void => {
								this.set('spice', !instance.spice);
							}}
						/>
						<PageSwitch
							disabled={this.state.disabled}
							label="Desktop GUI"
							help="Enable desktop GUI window for instance display."
							checked={instance.gui}
							onToggle={(): void => {
								this.set('gui', !instance.gui);
							}}
						/>
						<PageSwitch
							disabled={this.state.disabled}
							label="Root enabled"
							help="Enable root unix account for VNC/Spice access. Random password will be generated."
							checked={instance.root_enabled}
							onToggle={(): void => {
								this.set('root_enabled', !instance.root_enabled);
							}}
						/>
						<PageSwitch
							disabled={this.state.disabled}
							label="UEFI"
							help="Enable UEFI boot, requires OVMF package for UEFI image."
							checked={instance.uefi}
							onToggle={(): void => {
								this.onUefi(!instance.uefi);
							}}
						/>
						<PageSwitch
							disabled={this.state.disabled}
							hidden={!instance.uefi}
							label="SecureBoot"
							help="Enable secure boot, requires OVMF package for UEFI image."
							checked={instance.secure_boot}
							onToggle={(): void => {
								this.onSecureBoot(!instance.secure_boot);
							}}
						/>
						<PageSwitch
							disabled={this.state.disabled}
							hidden={!instance.uefi}
							label="TPM"
							help="Enable TPM, requires swtpm and OVMF package."
							checked={instance.tpm}
							onToggle={(): void => {
								this.set('tpm', !instance.tpm);
							}}
						/>
						<PageSwitch
							disabled={this.state.disabled}
							label="DHCP server"
							help="Enable instance DHCP server, use for instances without cloud init network configuration support."
							checked={instance.dhcp_server}
							onToggle={(): void => {
								this.onDhcpServer(!instance.dhcp_server);
							}}
						/>
						<PageSwitch
							disabled={this.state.disabled}
							label="Delete protection"
							help="Block instance and any attached disks from being deleted."
							checked={instance.delete_protection}
							onToggle={(): void => {
								this.set('delete_protection', !instance.delete_protection);
							}}
						/>
					</div>
					<div style={css.group}>
						<PageSelect
							disabled={this.state.disabled || !!instance.shape}
							label="Disk Type"
							help="Type of disk. QCOW disk files are stored locally on the node filesystem. LVM disks are partitioned as a logical volume."
							value={instance.disk_type}
							onChange={(val): void => {
								this.set('disk_type', val);
							}}
						>
							<option key="qcow2" value="qcow2">QCOW</option>
							<option key="lvm" value="lvm">LVM</option>
						</PageSelect>
						<PageSelect
							disabled={this.state.disabled || !hasPools}
							label="Pool"
							help="LVM volume group to add disk to."
							hidden={instance.disk_type !== "lvm"}
							value={instance.disk_pool}
							onChange={(val): void => {
								this.set('disk_pool', val);
							}}
						>
							{poolsSelect}
						</PageSelect>
						<InstanceImages
							images={this.state.images || []}
							image={this.state.image}
							uefi={instance.uefi}
							showHidden={this.state.hiddenImages}
							disabled={this.state.disabled}
							onChange={this.onImageselect}
						/>
						<PageSwitch
							label="Show hidden images"
							help="Show previous versions of images."
							checked={this.state.hiddenImages}
							onToggle={(): void => {
								this.setState({
									...this.state,
									hiddenImages: !this.state.hiddenImages,
								});
							}}
						/>
						<PageSwitch
							label="Linked disk image"
							help="Link to source disk image instead of creating full copy. This will reduce disk size and provide faster startup."
							checked={instance.image_backing}
							onToggle={(): void => {
								this.set('image_backing', !instance.image_backing);
							}}
						/>
						<label className="bp5-label">
							Roles
							<Help
								title="Roles"
								content="Roles that will be matched with firewall rules. Roles are case-sensitive."
							/>
							<div>
								{networkRoles}
							</div>
						</label>
						<PageInputButton
							disabled={this.state.disabled}
							buttonClass="bp5-intent-success bp5-icon-add"
							label="Add"
							type="text"
							placeholder="Add role"
							value={this.state.addNetworkRole}
							onChange={(val): void => {
								this.setState({
									...this.state,
									addNetworkRole: val,
								});
							}}
							onSubmit={this.onAddNetworkRole}
						/>
						<PageNumInput
							label="Disk Size"
							help="Instance memory size in megabytes."
							min={10}
							minorStepSize={1}
							stepSize={2}
							majorStepSize={5}
							disabled={this.state.disabled}
							selectAllOnFocus={true}
							onChange={(val: number): void => {
								this.set('init_disk_size', val);
							}}
							value={instance.init_disk_size}
						/>
						<PageNumInput
							label="Memory Size"
							help="Instance memory size in megabytes."
							min={256}
							minorStepSize={512}
							stepSize={1024}
							majorStepSize={2048}
							disabled={this.state.disabled || lockResources}
							selectAllOnFocus={true}
							onChange={(val: number): void => {
								this.set('memory', val);
							}}
							value={instance.memory}
						/>
						<PageNumInput
							label="Processors"
							help="Number of instance processors."
							min={1}
							minorStepSize={1}
							stepSize={1}
							majorStepSize={2}
							disabled={this.state.disabled || lockResources}
							selectAllOnFocus={true}
							onChange={(val: number): void => {
								this.set('processors', val);
							}}
							value={instance.processors}
						/>
						<PageNumInput
							label="Count"
							help="Number of instances to create."
							min={1}
							minorStepSize={1}
							stepSize={1}
							majorStepSize={1}
							disabled={this.state.disabled}
							selectAllOnFocus={true}
							onChange={(val: number): void => {
								this.set('count', val);
							}}
							value={instance.count}
						/>
						<label style={css.itemsLabel}>
							Node Ports
							<Help
								title="Node Ports"
								content="Node port mappings from node public IP to internal instance. Acceptable external port range is 30000-32767, leave external port empty to automatically assign a port."
							/>
						</label>
						{nodePorts}
						<button
							className="bp5-button bp5-intent-success bp5-icon-add"
							style={css.itemsAdd}
							type="button"
							onClick={this.onAddNodePort}
						>
							Add Node Port
						</button>
						<PageSwitch
							disabled={this.state.disabled}
							label="Public IPv4 address"
							help="Enable or disable public IPv4 address for instance. Node must have network mode configured to assign public address."
							checked={!instance.no_public_address}
							onToggle={(): void => {
								this.set('no_public_address', !instance.no_public_address);
							}}
						/>
						<PageSwitch
							disabled={this.state.disabled}
							label="Public IPv6 address"
							help="Enable or disable public IPv6 address for instance. Node must have network mode configured to assign public address."
							checked={!instance.no_public_address6}
							onToggle={(): void => {
								this.set('no_public_address6', !instance.no_public_address6);
							}}
						/>
						<PageSwitch
							disabled={this.state.disabled}
							label="Host address"
							help="Enable or disable host address for instance. Node must have host networking configured to assign host address."
							checked={!instance.no_host_address}
							onToggle={(): void => {
								this.set('no_host_address', !instance.no_host_address);
							}}
						/>
						<PageSwitch
							disabled={this.state.disabled}
							label="Skip source/destination check"
							help="Allow network traffic from non-instance addresses."
							checked={instance.skip_source_dest_check}
							onToggle={(): void => {
								this.set('skip_source_dest_check',
									!instance.skip_source_dest_check);
							}}
						/>
					</div>
				</div>
				<PageCreate
					style={css.save}
					hidden={!this.state.instance}
					message={this.state.message}
					changed={this.state.changed}
					disabled={this.state.disabled}
					closed={this.state.closed}
					light={true}
					onCancel={this.props.onClose}
					onCreate={this.onCreate}
				/>
			</td>
		</div>;
	}
}

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
import * as ZoneTypes from '../types/ZoneTypes';
import * as InstanceActions from '../actions/InstanceActions';
import * as ImageActions from '../actions/ImageActions';
import * as NodeActions from '../actions/NodeActions';
import ImagesDatacenterStore from '../stores/ImagesDatacenterStore';
import NodesZoneStore from '../stores/NodesZoneStore';
import InstanceLicense from './InstanceLicense';
import PageInput from './PageInput';
import PageInputButton from './PageInputButton';
import PageCreate from './PageCreate';
import PageSelect from './PageSelect';
import PageSwitch from "./PageSwitch";
import PageNumInput from './PageNumInput';
import Help from './Help';
import OrganizationsStore from "../stores/OrganizationsStore";

interface Props {
	organizations: OrganizationTypes.OrganizationsRo;
	vpcs: VpcTypes.VpcsRo;
	domains: DomainTypes.DomainsRo;
	datacenters: DatacenterTypes.DatacentersRo;
	zones: ZoneTypes.ZonesRo;
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
	images: ImageTypes.ImagesRo;
	nodes: NodeTypes.NodesRo;
	addNetworkRole: string;
	addVpc: string;
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
		height: '20px',
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
		height: '20px',
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
			images: [],
			nodes: [],
			addNetworkRole: '',
			addVpc: '',
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
			name: 'New instance',
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
		if (this.state.instance.image &&
				this.state.instance.image.indexOf('oracle') &&
				!License.oracle) {
			this.setState({
				...this.state,
				licenseOpen: true,
			});
			return
		}

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

	onAddNetworkRole = (): void => {
		if (!this.state.addNetworkRole) {
			return;
		}

		let instance = {
			...this.state.instance,
		};

		let networkRoles = [
			...(instance.network_roles || []),
		];

		if (networkRoles.indexOf(this.state.addNetworkRole) === -1) {
			networkRoles.push(this.state.addNetworkRole);
		}

		networkRoles.sort();
		instance.network_roles = networkRoles;

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
			...(instance.network_roles || []),
		];

		let i = networkRoles.indexOf(networkRole);
		if (i === -1) {
			return;
		}

		networkRoles.splice(i, 1);
		instance.network_roles = networkRoles;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addNetworkRole: '',
			instance: instance,
		});
	}

	render(): JSX.Element {
		let instance = this.state.instance;

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
					if (vpc.organization !== OrganizationsStore.current) {
						continue;
					}
				} else {
					if (vpc.organization !== instance.organization) {
						continue;
					}
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

		let domainsSelect: JSX.Element[] = [
			<option key="null" value="">No Domain</option>,
		];
		if (this.props.domains && this.props.domains.length) {
			for (let domain of this.props.domains) {
				if (!Constants.user &&
						domain.organization !== instance.organization) {
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

		let hasImages = false;
		let imagesSelect: JSX.Element[] = [];
		if (this.state.images.length) {
			imagesSelect.push(<option key="null" value="">Select Image</option>);

			hasImages = true;
			for (let image of this.state.images) {
				imagesSelect.push(
					<option
						key={image.id}
						value={image.id}
					>{image.name}</option>,
				);
			}
		}

		let networkRoles: JSX.Element[] = [];
		for (let networkRole of (instance.network_roles || [])) {
			networkRoles.push(
				<div
					className="bp3-tag bp3-tag-removable bp3-intent-primary"
					style={css.role}
					key={networkRole}
				>
					{networkRole}
					<button
						className="bp3-tag-remove"
						disabled={this.state.disabled}
						onMouseUp={(): void => {
							this.onRemoveNetworkRole(networkRole);
						}}
					/>
				</div>,
			);
		}

		if (!hasImages) {
			imagesSelect = [<option key="null" value="">No Images</option>];
		}

		return <div
			className="bp3-card bp3-row"
			style={css.row}
		>
			<td
				className="bp3-cell"
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
							disabled={this.state.disabled || !hasNodes}
							label="Node"
							help="Node to run instance on."
							value={instance.node}
							onChange={(val): void => {
								this.set('node', val);
							}}
						>
							{nodesSelect}
						</PageSelect>
						<PageSelect
							disabled={this.state.disabled}
							label="DNS Domain"
							help="Domain to create DNS name using instance name."
							value={instance.domain}
							onChange={(val): void => {
								this.set('domain', val);
							}}
						>
							{domainsSelect}
						</PageSelect>
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
							label="Start instance"
							help="Automatically start instance. Disable to get the public MAC address before instance is started for first time."
							checked={!instance.state}
							onToggle={(): void => {
								if (instance.state === 'stop') {
									this.set('state', '');
								} else {
									this.set('state', 'stop');
								}
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
							disabled={this.state.disabled || !hasImages}
							label="Image"
							help="Starting image for node."
							value={instance.image}
							onChange={(val): void => {
								this.set('image', val);
							}}
						>
							{imagesSelect}
						</PageSelect>
						<PageSwitch
							label="Linked disk image"
							help="Link to source disk image instead of creating full copy. This will reduce disk size and provide faster startup."
							checked={instance.image_backing}
							onToggle={(): void => {
								this.set('image_backing', !instance.image_backing);
							}}
						/>
						<label className="bp3-label">
							Network Roles
							<Help
								title="Network Roles"
								content="Network roles that will be matched with firewall rules. Network roles are case-sensitive."
							/>
							<div>
								{networkRoles}
							</div>
						</label>
						<PageInputButton
							disabled={this.state.disabled}
							buttonClass="bp3-intent-success bp3-icon-add"
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
							minorStepSize={256}
							stepSize={512}
							majorStepSize={1024}
							disabled={this.state.disabled}
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
							disabled={this.state.disabled}
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

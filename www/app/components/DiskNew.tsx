/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as DiskTypes from '../types/DiskTypes';
import * as OrganizationTypes from '../types/OrganizationTypes';
import * as DatacenterTypes from '../types/DatacenterTypes';
import * as NodeTypes from '../types/NodeTypes';
import * as InstanceTypes from '../types/InstanceTypes';
import * as ImageTypes from '../types/ImageTypes';
import * as ZoneTypes from '../types/ZoneTypes';
import * as DiskActions from '../actions/DiskActions';
import * as ImageActions from '../actions/ImageActions';
import * as InstanceActions from '../actions/InstanceActions';
import * as NodeActions from '../actions/NodeActions';
import ImagesDatacenterStore from '../stores/ImagesDatacenterStore';
import InstancesNodeStore from '../stores/InstancesNodeStore';
import NodesZoneStore from '../stores/NodesZoneStore';
import PageInput from './PageInput';
import PageInputButton from './PageInputButton';
import PageCreate from './PageCreate';
import PageSelect from './PageSelect';
import PageNumInput from './PageNumInput';
import Help from './Help';

interface Props {
	organizations: OrganizationTypes.OrganizationsRo;
	datacenters: DatacenterTypes.DatacentersRo;
	zones: ZoneTypes.ZonesRo;
	onClose: () => void;
}

interface State {
	closed: boolean;
	disabled: boolean;
	changed: boolean;
	message: string;
	disk: DiskTypes.Disk;
	datacenter: string;
	zone: string;
	images: ImageTypes.ImagesRo;
	nodes: NodeTypes.NodesRo;
	instances: InstanceTypes.InstancesRo;
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
		minWidth: '250px',
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

export default class DiskNew extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			closed: false,
			disabled: false,
			changed: false,
			message: '',
			disk: {
				name: 'New Disk',
				index: "1",
				size: 10,
			},
			datacenter: '',
			zone: '',
			images: [],
			nodes: [],
			instances: [],
		};
	}

	componentDidMount(): void {
		ImagesDatacenterStore.addChangeListener(this.onChange);
		NodesZoneStore.addChangeListener(this.onChange);
		InstancesNodeStore.addChangeListener(this.onChange);
		ImageActions.syncDatacenter('');
		NodeActions.syncZone('');
	}

	componentWillUnmount(): void {
		ImagesDatacenterStore.removeChangeListener(this.onChange);
		NodesZoneStore.removeChangeListener(this.onChange);
		InstancesNodeStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		this.setState({
			...this.state,
			images: ImagesDatacenterStore.images,
			nodes: NodesZoneStore.nodes,
			instances: InstancesNodeStore.instances(this.state.disk.node),
		});
	}

	set(name: string, val: any): void {
		let disk: any = {
			...this.state.disk,
		};

		disk[name] = val;

		this.setState({
			...this.state,
			changed: true,
			disk: disk,
		});
	}

	onCreate = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});

		let disk: any = {
			...this.state.disk,
		};

		if (this.props.organizations.length && !disk.organization) {
			disk.organization = this.props.organizations[0].id;
		}

		DiskActions.create(disk).then((): void => {
			this.setState({
				...this.state,
				message: 'Disk created successfully',
				closed: true,
				changed: false,
			});
		}).catch((): void => {
			this.setState({
				...this.state,
				message: '',
				disabled: false,
			});
		});
	}

	render(): JSX.Element {
		let disk = this.state.disk;

		let organizationsSelect: JSX.Element[] = [];
		if (this.props.organizations.length) {
			for (let organization of this.props.organizations) {
				organizationsSelect.push(
					<option
						key={organization.id}
						value={organization.id}
					>{organization.name}</option>,
				);
			}
		}

		let hasDatacenters = false;
		let datacentersSelect: JSX.Element[] = [];
		if (this.props.datacenters.length) {
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
		if (this.props.zones.length) {
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
		if (this.state.nodes.length) {
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

		let hasInstances = false;
		let instancesSelect: JSX.Element[] = [];
		if (this.state.instances.length) {
			instancesSelect.push(
				<option key="null" value="">Detached Disk</option>);

			hasInstances = true;
			for (let instance of this.state.instances) {
				instancesSelect.push(
					<option
						key={instance.id}
						value={instance.id}
					>{instance.name}</option>,
				);
			}
		}

		if (!hasInstances) {
			instancesSelect = [<option key="null" value="">No Instances</option>];
		}

		let imagesSelect: JSX.Element[] = [
			<option key="null" value="">Blank Disk</option>,
		];
		if (this.state.images.length) {
			for (let image of this.state.images) {
				imagesSelect.push(
					<option
						key={image.id}
						value={image.id}
					>{image.name}</option>,
				);
			}
		}

		return <div
			className="pt-card pt-row"
			style={css.row}
		>
			<td
				className="pt-cell"
				colSpan={5}
				style={css.card}
			>
				<div className="layout horizontal wrap">
					<div style={css.group}>
						<div style={css.buttons}>
						</div>
						<PageInput
							label="Name"
							help="Name of disk"
							type="text"
							placeholder="Enter name"
							disabled={this.state.disabled}
							value={disk.name}
							onChange={(val): void => {
								this.set('name', val);
							}}
						/>
						<PageSelect
							disabled={this.state.disabled}
							label="Organization"
							help="Organization for disk."
							value={disk.organization}
							onChange={(val): void => {
								this.set('organization', val);
							}}
						>
							{organizationsSelect}
						</PageSelect>
						<PageSelect
							disabled={this.state.disabled || !hasDatacenters}
							label="Datacenter"
							help="Datacenter for disk."
							value={this.state.datacenter}
							onChange={(val): void => {
								this.setState({
									...this.state,
									datacenter: val,
									disk: {
										...this.state.disk,
										node: '',
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
							help="Zone for disk."
							value={this.state.zone}
							onChange={(val): void => {
								this.setState({
									...this.state,
									zone: val,
									disk: {
										...this.state.disk,
										node: '',
										instance: '',
									},
								});
								NodeActions.syncZone(val);
							}}
						>
							{zonesSelect}
						</PageSelect>
						<PageSelect
							disabled={this.state.disabled || !hasNodes}
							label="Node"
							help="Node to run disk on."
							value={disk.node}
							onChange={(val): void => {
								this.setState({
									...this.state,
									disk: {
										...this.state.disk,
										node: val,
										instance: '',
									},
								});
								InstanceActions.syncNode(val);
							}}
						>
							{nodesSelect}
						</PageSelect>
					</div>
					<div style={css.group}>
						<PageSelect
							disabled={this.state.disabled || !hasInstances}
							label="Instance"
							help="Instance to attach disk to."
							value={disk.instance}
							onChange={(val): void => {
								this.set('instance', val);
							}}
						>
							{instancesSelect}
						</PageSelect>
						<PageNumInput
							label="Index"
							help="Index to attach disk."
							hidden={!disk.instance}
							min={0}
							max={8}
							minorStepSize={1}
							stepSize={1}
							majorStepSize={1}
							disabled={this.state.disabled}
							selectAllOnFocus={true}
							value={Number(disk.index)}
							onChange={(val: number): void => {
								this.set('index', String(val));
							}}
						/>
						<PageSelect
							disabled={this.state.disabled}
							label="Image"
							help="Starting image for node."
							value={disk.image}
							onChange={(val): void => {
								this.set('image', val);
							}}
						>
							{imagesSelect}
						</PageSelect>
						<PageNumInput
							label="Size"
							help="Disk size in gigabytes."
							min={10}
							minorStepSize={1}
							stepSize={1}
							majorStepSize={2}
							disabled={this.state.disabled}
							selectAllOnFocus={true}
							value={disk.size}
							onChange={(val: number): void => {
								this.set('size', val);
							}}
						/>
					</div>
				</div>
				<PageCreate
					style={css.save}
					hidden={!this.state.disk}
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

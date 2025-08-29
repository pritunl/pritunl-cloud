/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as DiskTypes from '../types/DiskTypes';
import * as OrganizationTypes from '../types/OrganizationTypes';
import * as DatacenterTypes from '../types/DatacenterTypes';
import * as NodeTypes from '../types/NodeTypes';
import * as InstanceTypes from '../types/InstanceTypes';
import * as ImageTypes from '../types/ImageTypes';
import * as ZoneTypes from '../types/ZoneTypes';
import * as PoolTypes from '../types/PoolTypes';
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
import PageSwitch from "./PageSwitch";
import PageNumInput from './PageNumInput';
import Help from './Help';

interface Props {
	organizations: OrganizationTypes.OrganizationsRo;
	datacenters: DatacenterTypes.DatacentersRo;
	pools: PoolTypes.PoolsRo;
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
			hiddenImages: false,
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

	setUnset(name: string, val: any, unset: string): void {
		let disk: any = {
			...this.state.disk,
		};

		disk[name] = val;
		disk[unset] = null;

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

		let hasPools = false;
		let poolsSelect: JSX.Element[] = [];
		if (this.props.pools.length) {
			poolsSelect.push(<option key="null" value="">Select Pool</option>);

			for (let pool of this.props.pools) {
				if (pool.zone !== this.state.zone) {
					continue
				}

				hasPools = true;
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
		let imagesMap = new Map();
		let imagesVer = new Map();
		if (this.state.images.length) {
			for (let image of this.state.images) {
				if (!this.state.hiddenImages && image.signed) {
					let imgSpl = image.key.split('_');

					if (imgSpl.length >= 2 && imgSpl[imgSpl.length - 1].length >= 4) {
						let imgKey = imgSpl[0] + '_' + image.firmware

						let imgVer = parseInt(
							imgSpl[imgSpl.length - 1].substring(0, 4), 10);
						if (imgVer) {
							let curImg = imagesVer.get(imgKey);
							if (!curImg || imgVer > curImg[0]) {
								imagesVer.set(imgKey, [imgVer, image.id, image.name]);
							}
							continue;
						}
					}
				}

				imagesMap.set(image.id, image.name);
			}

			for (let item of imagesMap.entries()) {
				imagesSelect.push(
					<option
						key={item[0]}
						value={item[0]}
					>{item[1]}</option>,
				);
			}

			for (let item of imagesVer.entries()) {
				imagesSelect.push(
					<option
						key={item[1][1]}
						value={item[1][1]}
					>{item[1][2]}</option>,
				);
			}
		}

		return <div
			className="bp5-card bp5-row"
			style={css.row}
		>
			<td
				className="bp5-cell"
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
							label="Type"
							help="Type of disk. QCOW disk files are stored locally on the node filesystem. LVM disks are partitioned as a logical volume."
							value={disk.type}
							onChange={(val): void => {
								this.set('type', val);
							}}
						>
							<option key="qcow2" value="qcow2">QCOW</option>
							<option key="lvm" value="lvm">LVM</option>
						</PageSelect>
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
							hidden={disk.type === "lvm"}
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
								InstanceActions.syncNode(val, null);
							}}
						>
							{nodesSelect}
						</PageSelect>
						<PageSelect
							disabled={this.state.disabled || !hasPools}
							label="Pool"
							help="LVM volume group to add disk to."
							hidden={disk.type !== "lvm"}
							value={disk.pool}
							onChange={(val): void => {
								this.set('pool', val);
							}}
						>
							{poolsSelect}
						</PageSelect>
						<PageSwitch
							disabled={this.state.disabled}
							label="Delete protection"
							help="Block disk from being deleted."
							checked={disk.delete_protection}
							onToggle={(): void => {
								this.set('delete_protection', !disk.delete_protection);
							}}
						/>
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
							help="Starting image for disk."
							value={disk.image}
							onChange={(val): void => {
								this.setUnset('image', val, 'file_system');
							}}
						>
							{imagesSelect}
						</PageSelect>
						<PageSelect
							disabled={this.state.disabled}
							label="File System"
							help="Starting file system for disk."
							value={disk.file_system}
							onChange={(val): void => {
								this.setUnset('file_system', val, 'image');
							}}
						>
							<option value="">None</option>
							<option value="xfs">XFS</option>
							<option value="ext4">ext4</option>
							<option value="lvm_xfs">LVM + XFS</option>
							<option value="lvm_ext4">LVM + ext4</option>
						</PageSelect>
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
							help="Link to source disk image instead of creating full copy. This will reduce disk size and provide faster creation."
							checked={disk.backing}
							onToggle={(): void => {
								this.set('backing', !disk.backing);
							}}
						/>
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

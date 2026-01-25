/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as ShapeTypes from '../types/ShapeTypes';
import * as OrganizationTypes from '../types/OrganizationTypes';
import * as DatacenterTypes from '../types/DatacenterTypes';
import * as NodeTypes from '../types/NodeTypes';
import * as InstanceTypes from '../types/InstanceTypes';
import * as ImageTypes from '../types/ImageTypes';
import * as ZoneTypes from '../types/ZoneTypes';
import * as ShapeActions from '../actions/ShapeActions';
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
import PageTextArea from "./PageTextArea";
import * as PoolTypes from "../types/PoolTypes";

interface Props {
	datacenters: DatacenterTypes.DatacentersRo;
	zones: ZoneTypes.ZonesRo;
	pools: PoolTypes.PoolsRo;
	onClose: () => void;
}

interface State {
	closed: boolean;
	disabled: boolean;
	changed: boolean;
	message: string;
	shape: ShapeTypes.Shape;
	datacenter: string;
	zone: string;
	addRole: string;
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

export default class ShapeNew extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			closed: false,
			disabled: false,
			changed: false,
			message: '',
			shape: {
				name: 'new-shape',
				memory: 1024,
				processors: 1,
				flexible: true,
			},
			datacenter: '',
			zone: '',
			addRole: '',
		};
	}

	set(name: string, val: any): void {
		let shape: any = {
			...this.state.shape,
		};

		shape[name] = val;

		this.setState({
			...this.state,
			changed: true,
			shape: shape,
		});
	}

	onCreate = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});

		let shape: any = {
			...this.state.shape,
		};

		if (!shape.datacenter && this.props.datacenters.length) {
			shape.datacenter = this.state.datacenter ||
				this.props.datacenters[0].id;
		}

		ShapeActions.create(shape).then((): void => {
			this.setState({
				...this.state,
				message: 'Shape created successfully',
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

	onAddRole = (): void => {
		let shape: ShapeTypes.Shape;

		if (!this.state.addRole) {
			return;
		}

		shape = {
			...this.state.shape,
		};

		let roles = [
			...(shape.roles || []),
		];

		if (roles.indexOf(this.state.addRole) === -1) {
			roles.push(this.state.addRole);
		}

		roles.sort();
		shape.roles = roles;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addRole: '',
			shape: shape,
		});
	}

	onRemoveRole = (role: string): void => {
		let shape: ShapeTypes.Shape;

		shape = {
			...this.state.shape,
		};

		let roles = [
			...(shape.roles || []),
		];

		let i = roles.indexOf(role);
		if (i === -1) {
			return;
		}

		roles.splice(i, 1);
		shape.roles = roles;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addRole: '',
			shape: shape,
		});
	}

	render(): JSX.Element {
		let shape = this.state.shape;

		let defaultDatacenter = '';
		let hasDatacenters = false;
		let datacentersSelect: JSX.Element[] = [];
		if (this.props.datacenters.length) {
			hasDatacenters = true;
			defaultDatacenter = this.props.datacenters[0].id;
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

		let datacenter = this.state.datacenter || defaultDatacenter;
		let hasZones = false;
		let zonesSelect: JSX.Element[] = [];
		if (this.props.zones.length) {
			zonesSelect.push(<option key="null" value="">Select Zone</option>);

			for (let zone of this.props.zones) {
				if (!this.state.zone && zone.datacenter !== datacenter) {
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

		let roles: JSX.Element[] = [];
		for (let role of (shape.roles || [])) {
			roles.push(
				<div
					className="bp5-tag bp5-tag-removable bp5-intent-primary"
					style={css.role}
					key={role}
				>
					{role}
					<button
						className="bp5-tag-remove"
						disabled={this.state.disabled}
						onMouseUp={(): void => {
							this.onRemoveRole(role);
						}}
					/>
				</div>,
			);
		}

		return <div
			className="bp5-card bp5-row"
			style={css.row}
		>
			<td
				className="bp5-cell"
				colSpan={2}
				style={css.card}
			>
				<div className="layout horizontal wrap">
					<div style={css.group}>
						<div style={css.buttons}>
						</div>
						<PageInput
							label="Name"
							help="Name of shape"
							type="text"
							placeholder="Enter name"
							value={shape.name}
							onChange={(val): void => {
								this.set('name', val);
							}}
						/>
						<PageTextArea
							label="Comment"
							help="Shape comment."
							placeholder="Shape comment"
							rows={3}
							value={shape.comment}
							onChange={(val: string): void => {
								this.set('comment', val);
							}}
						/>
						<PageSelect
							disabled={this.state.disabled || !hasDatacenters}
							label="Datacenter"
							help="Shape datacenter, cannot be changed once set."
							value={this.state.datacenter}
							onChange={(val): void => {
								this.set('datacenter', val);
							}}
						>
							{datacentersSelect}
						</PageSelect>
						<PageSelect
							disabled={this.state.disabled}
							label="Disk Type"
							help="Type of disk. QCOW disk files are stored locally on the node filesystem. LVM disks are partitioned as a logical volume."
							value={shape.disk_type}
							onChange={(val): void => {
								this.set('disk_type', val);
							}}
						>
							<option key="qcow2" value="qcow2">QCOW</option>
							<option key="lvm" value="lvm">LVM</option>
						</PageSelect>
						<PageSelect
							disabled={this.state.disabled || !hasPools}
							label="Disk Pool"
							help="Disk pool to use for storage."
							hidden={shape.disk_type !== "lvm"}
							value={shape.disk_pool}
							onChange={(val): void => {
								this.set('disk_pool', val);
							}}
						>
							{poolsSelect}
						</PageSelect>
						<PageSwitch
							disabled={this.state.disabled}
							label="Flexible"
							help="Allow process and memory to be customized for each instance."
							checked={shape.flexible}
							onToggle={(): void => {
								this.set('flexible', !shape.flexible);
							}}
						/>
						<PageSwitch
							disabled={this.state.disabled}
							label="Delete protection"
							help="Block shape from being deleted."
							checked={shape.delete_protection}
							onToggle={(): void => {
								this.set('delete_protection', !shape.delete_protection);
							}}
						/>
					</div>
					<div style={css.group}>
						<PageNumInput
							label="Memory Size"
							help="Instance memory size in megabytes."
							min={256}
							minorStepSize={512}
							stepSize={1024}
							majorStepSize={2048}
							disabled={this.state.disabled}
							selectAllOnFocus={true}
							onChange={(val: number): void => {
								this.set('memory', val);
							}}
							value={shape.memory}
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
							value={shape.processors}
						/>
						<label className="bp5-label">
							Roles
							<Help
								title="Roles"
								content="Roles that will be matched with nodes. Nodes that provide this shape must have a matching role."
							/>
							<div>
								{roles}
							</div>
						</label>
						<PageInputButton
							disabled={this.state.disabled}
							buttonClass="bp5-intent-success bp5-icon-add"
							label="Add"
							type="text"
							placeholder="Add role"
							value={this.state.addRole}
							onChange={(val): void => {
								this.setState({
									...this.state,
									addRole: val,
								});
							}}
							onSubmit={this.onAddRole}
						/>
					</div>
				</div>
				<PageCreate
					style={css.save}
					hidden={!this.state.shape}
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

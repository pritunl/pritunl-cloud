/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as ShapeTypes from '../types/ShapeTypes';
import * as ZoneTypes from '../types/ZoneTypes';
import * as PoolTypes from '../types/PoolTypes';
import * as ShapeActions from '../actions/ShapeActions';
import PageInput from './PageInput';
import PageInputButton from './PageInputButton';
import PageSelectButton from './PageSelectButton';
import PageCreate from './PageCreate';
import PageSelect from './PageSelect';
import PageSwitch from "./PageSwitch";
import PageNumInput from './PageNumInput';
import Help from './Help';
import PageTextArea from "./PageTextArea";

interface Props {
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
	addRole: string;
	addPool: string;
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
				type: 'instance',
				disk_type: 'qcow2',
				memory: 1024,
				processors: 1,
				flexible: true,
			},
			addRole: '',
			addPool: '',
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

	onDiskType = (val: string): void => {
		let shape: ShapeTypes.Shape = {
			...this.state.shape,
		};

		shape.disk_type = val;
		if (val === 'lvm') {
			shape.disks_path = '';
		} else {
			shape.disk_pools = [];
		}

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

	onAddPool = (): void => {
		let shape: ShapeTypes.Shape = {
			...this.state.shape,
		};

		if (!this.state.addPool && !this.props.pools.length) {
			return;
		}

		let poolId = this.state.addPool;
		if (!poolId) {
			poolId = this.props.pools[0].id;
		}

		let pools = [
			...(shape.disk_pools || []),
		];

		if (pools.indexOf(poolId) === -1) {
			pools.push(poolId);
		}

		shape.disk_pools = pools;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			shape: shape,
		});
	}

	onRemovePool = (poolId: string): void => {
		let shape: ShapeTypes.Shape = {
			...this.state.shape,
		};

		let pools = [
			...(shape.disk_pools || []),
		];

		let i = pools.indexOf(poolId);
		if (i === -1) {
			return;
		}

		pools.splice(i, 1);
		shape.disk_pools = pools;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			shape: shape,
		});
	}

	poolLabel(pool: PoolTypes.PoolRo): string {
		for (let zone of (this.props.zones || [])) {
			if (zone.id === pool.zone) {
				return pool.name + ' (' + zone.name + ')';
			}
		}
		return pool.name;
	}

	render(): JSX.Element {
		let shape = this.state.shape;
		let isDisk = shape.type === 'disk';
		let diskType = shape.disk_type || 'qcow2';

		let pools: JSX.Element[] = [];
		for (let poolId of (shape.disk_pools || [])) {
			let label = poolId;
			for (let pool of (this.props.pools || [])) {
				if (pool.id === poolId) {
					label = this.poolLabel(pool);
					break;
				}
			}

			pools.push(
				<div
					className="bp5-tag bp5-tag-removable bp5-intent-primary"
					style={css.item}
					key={poolId}
				>
					{label}
					<button
						className="bp5-tag-remove"
						disabled={this.state.disabled}
						onMouseUp={(): void => {
							this.onRemovePool(poolId);
						}}
					/>
				</div>,
			);
		}

		let hasPools = false;
		let poolsSelect: JSX.Element[] = [];
		for (let pool of (this.props.pools || [])) {
			hasPools = true;
			poolsSelect.push(
				<option key={pool.id} value={pool.id}>
					{this.poolLabel(pool)}
				</option>,
			);
		}

		if (!hasPools) {
			poolsSelect = [
				<option key="null" value="">
					No Pools
				</option>,
			];
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
							rows={6}
							value={shape.comment}
							onChange={(val: string): void => {
								this.set('comment', val);
							}}
						/>
						<PageSelect
							disabled={this.state.disabled}
							label="Type"
							help="Type of shape. Instance shapes define the processors and memory of an instance and select the node from the matching roles. Disk shapes define where disks are stored on the nodes with matching roles."
							value={shape.type}
							onChange={(val): void => {
								this.set('type', val);
							}}
						>
							<option key="instance" value="instance">Instance</option>
							<option key="disk" value="disk">Disk</option>
						</PageSelect>
						<PageSelect
							disabled={this.state.disabled}
							label="Disk Type"
							help="Type of disk storage. QCOW disk files are stored in the disks path on the node filesystem. LVM disks are logical volumes in a shared disk pool of the zone."
							hidden={!isDisk}
							value={diskType}
							onChange={(val): void => {
								this.onDiskType(val);
							}}
						>
							<option key="qcow2" value="qcow2">QCOW</option>
							<option key="lvm" value="lvm">LVM</option>
						</PageSelect>
						{isDisk && diskType === 'qcow2' ? <PageInput
							label="Disks Path"
							help="Absolute path of the directory on the node filesystem that stores the QCOW disk files of this shape. All nodes with a matching role must have this path available."
							type="text"
							placeholder="/var/lib/pritunl-cloud/disks"
							disabled={this.state.disabled}
							value={shape.disks_path}
							onChange={(val): void => {
								this.set('disks_path', val);
							}}
						/> : null}
						<PageSwitch
							disabled={this.state.disabled}
							label="Flexible"
							help="Allow process and memory to be customized for each instance."
							hidden={isDisk}
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
							hidden={isDisk}
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
							hidden={isDisk}
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
						<label
							className="bp5-label"
							hidden={!isDisk || diskType !== 'lvm'}
						>
							Disk Pools
							<Help
								title="Disk Pools"
								content="The LVM disk pools of this shape. A disk pool is a shared volume group available to the nodes of a zone, disks are created in the pool of the zone of the instance or node."
							/>
							<div>
								{pools}
							</div>
						</label>
						<PageSelectButton
							label="Add Pool"
							hidden={!isDisk || diskType !== 'lvm'}
							value={this.state.addPool}
							disabled={this.state.disabled || !hasPools}
							buttonClass="bp5-intent-success"
							onChange={(val: string): void => {
								this.setState({
									...this.state,
									addPool: val,
								});
							}}
							onSubmit={this.onAddPool}
						>
							{poolsSelect}
						</PageSelectButton>
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

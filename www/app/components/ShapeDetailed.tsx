/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as ShapeTypes from '../types/ShapeTypes';
import * as ShapeActions from '../actions/ShapeActions';
import * as ZoneTypes from "../types/ZoneTypes";
import * as PoolTypes from "../types/PoolTypes";
import PageInput from './PageInput';
import PageSelect from './PageSelect';
import PageSelectButton from './PageSelectButton';
import PageInfo from './PageInfo';
import PageInputButton from './PageInputButton';
import PageSave from './PageSave';
import ConfirmButton from './ConfirmButton';
import Relations from './Relations';
import Help from './Help';
import PageTextArea from "./PageTextArea";
import PageSwitch from "./PageSwitch";
import PageNumInput from "./PageNumInput";

interface Props {
	zones: ZoneTypes.ZonesRo;
	pools: PoolTypes.PoolsRo;
	shape: ShapeTypes.ShapeRo;
	selected: boolean;
	onSelect: (shift: boolean) => void;
	onClose: () => void;
}

interface State {
	disabled: boolean;
	changed: boolean;
	message: string;
	addRole: string;
	addPool: string;
	shape: ShapeTypes.Shape;
}

const css = {
	card: {
		position: 'relative',
		padding: '48px 10px 0 10px',
		width: '100%',
	} as React.CSSProperties,
	button: {
		height: '30px',
	} as React.CSSProperties,
	buttons: {
		cursor: 'pointer',
		position: 'absolute',
		top: 0,
		left: 0,
		right: 0,
		padding: '4px',
		height: '39px',
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
	status: {
		margin: '6px 0 0 1px',
	} as React.CSSProperties,
	icon: {
		marginRight: '3px',
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
	select: {
		margin: '7px 0px 0px 6px',
		paddingTop: '3px',
	} as React.CSSProperties,
	role: {
		margin: '9px 5px 0 5px',
		minHeight: '20px',
	} as React.CSSProperties,
	rules: {
		marginBottom: '15px',
	} as React.CSSProperties,
};

export default class ShapeDetailed extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			disabled: false,
			changed: false,
			message: '',
			shape: null,
			addRole: '',
			addPool: '',
		};
	}

	current(): ShapeTypes.Shape {
		if (this.state.changed) {
			return {
				...this.state.shape,
			};
		}
		return {
			...this.props.shape,
		};
	}

	set(name: string, val: any): void {
		let shape: any = this.current();

		shape[name] = val;

		this.setState({
			...this.state,
			changed: true,
			shape: shape,
		});
	}

	onDiskType = (val: string): void => {
		let shape: ShapeTypes.Shape = this.current();

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

	onSave = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		ShapeActions.commit(this.state.shape).then((): void => {
			this.setState({
				...this.state,
				message: 'Your changes have been saved',
				changed: false,
				disabled: false,
			});

			setTimeout((): void => {
				if (!this.state.changed) {
					this.setState({
						...this.state,
						shape: null,
						changed: false,
					});
				}
			}, 1000);

			setTimeout((): void => {
				if (!this.state.changed) {
					this.setState({
						...this.state,
						message: '',
					});
				}
			}, 3000);
		}).catch((): void => {
			this.setState({
				...this.state,
				message: '',
				disabled: false,
			});
		});
	}

	onDelete = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		ShapeActions.remove(this.props.shape.id).then((): void => {
			this.setState({
				...this.state,
				disabled: false,
			});
		}).catch((): void => {
			this.setState({
				...this.state,
				disabled: false,
			});
		});
	}

	onAddRole = (): void => {
		if (!this.state.addRole) {
			return;
		}

		let shape: ShapeTypes.Shape = this.current();

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
		let shape: ShapeTypes.Shape = this.current();

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
		if (!this.state.addPool && !this.props.pools.length) {
			return;
		}

		let shape: ShapeTypes.Shape = this.current();

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
		let shape: ShapeTypes.Shape = this.current();

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
		let shape: ShapeTypes.Shape = this.state.shape ||
			this.props.shape;

		let shapeType = shape.type || 'instance';
		let isDisk = shapeType === 'disk';

		let diskType = shape.disk_type;
		if (!diskType) {
			if (shape.disk_pools && shape.disk_pools.length) {
				diskType = 'lvm';
			} else {
				diskType = 'qcow2';
			}
		}

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

		return <td
			className="bp5-cell"
			colSpan={2}
			style={css.card}
		>
			<div className="layout horizontal wrap">
				<div style={css.group}>
					<div
						className="layout horizontal tab-close bp5-card-header"
						style={css.buttons}
						onClick={(evt): void => {
							if (evt.target instanceof HTMLElement &&
									evt.target.className.indexOf('tab-close') !== -1) {
								this.props.onClose();
							}
						}}
					>
						<div>
							<label
								className="bp5-control bp5-checkbox"
								style={css.select}
							>
								<input
									type="checkbox"
									checked={this.props.selected}
									onChange={(evt): void => {
									}}
									onClick={(evt): void => {
										this.props.onSelect(evt.shiftKey);
									}}
								/>
								<span className="bp5-control-indicator"/>
							</label>
						</div>
						<div className="flex tab-close"/>
						<Relations kind="shape" id={this.props.shape.id}/>
						<ConfirmButton
							className="bp5-minimal bp5-intent-danger bp5-icon-trash"
							style={css.button}
							safe={true}
							progressClassName="bp5-intent-danger"
							dialogClassName="bp5-intent-danger bp5-icon-delete"
							dialogLabel="Delete Shape"
							confirmMsg="Permanently delete this shape"
							confirmInput={true}
							items={[shape.name]}
							disabled={this.state.disabled}
							onConfirm={this.onDelete}
						/>
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
						value={shapeType}
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
					<PageInfo
						fields={[
							{
								label: 'ID',
								value: this.props.shape.id || 'None',
							},
							{
								label: 'Node Count',
								value: this.props.shape.node_count || '0',
							},
						]}
					/>
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
			<PageSave
				style={css.save}
				hidden={!this.state.shape && !this.state.message}
				message={this.state.message}
				changed={this.state.changed}
				disabled={this.state.disabled}
				light={true}
				onCancel={(): void => {
					this.setState({
						...this.state,
						changed: false,
						shape: null,
					});
				}}
				onSave={this.onSave}
			/>
		</td>;
	}
}

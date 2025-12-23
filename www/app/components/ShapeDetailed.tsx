/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Constants from '../Constants';
import * as ShapeTypes from '../types/ShapeTypes';
import * as ShapeActions from '../actions/ShapeActions';
import * as OrganizationTypes from "../types/OrganizationTypes";
import * as DatacenterTypes from "../types/DatacenterTypes";
import * as ZoneTypes from "../types/ZoneTypes";
import * as PoolTypes from "../types/PoolTypes";
import PageInput from './PageInput';
import PageSelect from './PageSelect';
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
	datacenters: DatacenterTypes.DatacentersRo;
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
	shape: ShapeTypes.Shape;
	datacenter: string;
	zone: string;
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
			datacenter: '',
			zone: '',
		};
	}

	set(name: string, val: any): void {
		let shape: any;

		if (this.state.changed) {
			shape = {
				...this.state.shape,
			};
		} else {
			shape = {
				...this.props.shape,
			};
		}

		shape[name] = val;

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
		let shape: ShapeTypes.Shape;

		if (!this.state.addRole) {
			return;
		}

		if (this.state.changed) {
			shape = {
				...this.state.shape,
			};
		} else {
			shape = {
				...this.props.shape,
			};
		}

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

		if (this.state.changed) {
			shape = {
				...this.state.shape,
			};
		} else {
			shape = {
				...this.props.shape,
			};
		}

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
		let shape: ShapeTypes.Shape = this.state.shape ||
			this.props.shape;

		let hasDatacenters = false;
		let datacentersSelect: JSX.Element[] = [];
		if (this.props.datacenters.length) {
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
						rows={3}
						value={shape.comment}
						onChange={(val: string): void => {
							this.set('comment', val);
						}}
					/>
					<PageSelect
						disabled={!!shape.datacenter || this.state.disabled || !hasDatacenters}
						label="Datacenter"
						help="Shape datacenter, cannot be changed once set."
						value={shape.datacenter}
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

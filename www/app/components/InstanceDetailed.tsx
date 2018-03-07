/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as InstanceTypes from '../types/InstanceTypes';
import * as InstanceActions from '../actions/InstanceActions';
import OrganizationsStore from '../stores/OrganizationsStore';
import ZonesStore from '../stores/ZonesStore';
import PageInput from './PageInput';
import PageInputButton from './PageInputButton';
import PageInfo from './PageInfo';
import PageSelectButton from './PageSelectButton';
import PageSave from './PageSave';
import PageNumInput from './PageNumInput';
import ConfirmButton from './ConfirmButton';
import Help from './Help';
import VpcsNameStore from "../stores/VpcsNameStore";
import * as VpcTypes from "../types/VpcTypes";

interface Props {
	vpcs: VpcTypes.VpcsRo;
	instance: InstanceTypes.InstanceRo;
	selected: boolean;
	onSelect: (shift: boolean) => void;
	onClose: () => void;
}

interface State {
	disabled: boolean;
	changed: boolean;
	message: string;
	instance: InstanceTypes.Instance;
	addCert: string;
	addNetworkRole: string;
	addVpc: string;
	forwardedChecked: boolean;
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
	controlButton: {
		marginRight: '10px',
	} as React.CSSProperties,
	buttons: {
		cursor: 'pointer',
		position: 'absolute',
		top: 0,
		left: 0,
		right: 0,
		padding: '4px',
		height: '39px',
		backgroundColor: 'rgba(0, 0, 0, 0.13)',
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
	} as React.CSSProperties,
	role: {
		margin: '9px 5px 0 5px',
		height: '20px',
	} as React.CSSProperties,
};

export default class InstanceDetailed extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			disabled: false,
			changed: false,
			message: '',
			instance: null,
			addCert: null,
			addNetworkRole: '',
			addVpc: '',
			forwardedChecked: false,
		};
	}

	set(name: string, val: any): void {
		let instance: any;

		if (this.state.changed) {
			instance = {
				...this.state.instance,
			};
		} else {
			instance = {
				...this.props.instance,
			};
		}

		instance[name] = val;

		this.setState({
			...this.state,
			changed: true,
			instance: instance,
		});
	}

	onAddVpc = (): void => {
		let instance: InstanceTypes.Instance;

		if (!this.state.addVpc) {
			return;
		}

		let vpcId = this.state.addVpc;

		if (this.state.changed) {
			instance = {
				...this.state.instance,
			};
		} else {
			instance = {
				...this.props.instance,
			};
		}

		let vpcs = [
			...(instance.vpcs || []),
		];

		if (vpcs.indexOf(vpcId) === -1) {
			vpcs.push(vpcId);
		}

		vpcs.sort();

		instance.vpcs = vpcs;

		this.setState({
			...this.state,
			changed: true,
			instance: instance,
		});
	}

	onRemoveVpc = (vpc: string): void => {
		let instance: InstanceTypes.Instance;

		if (this.state.changed) {
			instance = {
				...this.state.instance,
			};
		} else {
			instance = {
				...this.props.instance,
			};
		}

		let vpcs = [
			...(instance.vpcs || []),
		];

		let i = vpcs.indexOf(vpc);
		if (i === -1) {
			return;
		}

		vpcs.splice(i, 1);

		instance.vpcs = vpcs;

		this.setState({
			...this.state,
			changed: true,
			instance: instance,
		});
	}

	onAddNetworkRole = (): void => {
		let instance: InstanceTypes.Instance;

		if (!this.state.addNetworkRole) {
			return;
		}

		if (this.state.changed) {
			instance = {
				...this.state.instance,
			};
		} else {
			instance = {
				...this.props.instance,
			};
		}

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
		let instance: InstanceTypes.Instance;

		if (this.state.changed) {
			instance = {
				...this.state.instance,
			};
		} else {
			instance = {
				...this.props.instance,
			};
		}

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

	onSave = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		InstanceActions.commit(this.state.instance).then((): void => {
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
						instance: null,
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
		InstanceActions.remove(this.props.instance.id).then((): void => {
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

	update(state: string): void {
		this.setState({
			...this.state,
			disabled: true,
		});
		InstanceActions.updateMulti([this.props.instance.id],
				state).then((): void => {
			setTimeout((): void => {
				this.setState({
					...this.state,
					disabled: false,
				});
			}, 250);
		}).catch((): void => {
			this.setState({
				...this.state,
				disabled: false,
			});
		});
	}

	render(): JSX.Element {
		let instance: InstanceTypes.Instance = this.state.instance ||
			this.props.instance;
		let info: InstanceTypes.Info = this.props.instance.info || {};

		let org = OrganizationsStore.organization(
			this.props.instance.organization);
		let zone = ZonesStore.zone(this.props.instance.zone);

		let statusClass = '';
		switch (instance.status) {
			case 'Running':
				statusClass += 'pt-text-intent-success';
				break;
			case 'Restart Required':
				statusClass += ' pt-text-intent-warning';
				break;
			case 'Stopped':
			case 'Destroying':
				statusClass += 'pt-text-intent-danger';
				break;
		}

		let networkRoles: JSX.Element[] = [];
		for (let networkRole of (instance.network_roles || [])) {
			networkRoles.push(
				<div
					className="pt-tag pt-tag-removable pt-intent-primary"
					style={css.role}
					key={networkRole}
				>
					{networkRole}
					<button
						className="pt-tag-remove"
						disabled={this.state.disabled}
						onMouseUp={(): void => {
							this.onRemoveNetworkRole(networkRole);
						}}
					/>
				</div>,
			);
		}

		let vpcs: JSX.Element[] = [];
		for (let vpcId of (instance.vpcs || [])) {
			let vpc = VpcsNameStore.vpc(vpcId);
			if (!vpc) {
				continue;
			}

			vpcs.push(
				<div
					className="pt-tag pt-tag-removable pt-intent-primary"
					style={css.item}
					key={vpc.id}
				>
					{vpc.name}
					<button
						className="pt-tag-remove"
						onMouseUp={(): void => {
							this.onRemoveVpc(vpc.id);
						}}
					/>
				</div>,
			);
		}

		let hasVpcs = false;
		let vpcsSelect: JSX.Element[] = [];
		if (this.props.vpcs && this.props.vpcs.length) {
			vpcsSelect.push(<option key="null" value="">Select Vpc</option>);

			for (let vpc of this.props.vpcs) {
				if (vpc.organization !== instance.organization) {
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

		return <td
			className="pt-cell"
			colSpan={5}
			style={css.card}
		>
			<div className="layout horizontal wrap">
				<div style={css.group}>
					<div
						className="layout horizontal"
						style={css.buttons}
						onClick={(evt): void => {
							let target = evt.target as HTMLElement;

							if (target.className.indexOf('open-ignore') !== -1) {
								return;
							}

							this.props.onClose();
						}}
					>
            <div>
              <label
                className="pt-control pt-checkbox open-ignore"
                style={css.select}
              >
                <input
                  type="checkbox"
                  className="open-ignore"
                  checked={this.props.selected}
                  onClick={(evt): void => {
										this.props.onSelect(evt.shiftKey);
									}}
                />
                <span className="pt-control-indicator open-ignore"/>
              </label>
            </div>
						<div className={statusClass} style={css.status}>
							<span
								style={css.icon}
								hidden={!instance.status}
								className="pt-icon-standard pt-icon-power"
							/>
							{instance.status}
						</div>
						<div className="flex"/>
						<ConfirmButton
							className="pt-minimal pt-intent-danger pt-icon-trash open-ignore"
							style={css.button}
							progressClassName="pt-intent-danger"
							confirmMsg="Confirm instance remove"
							disabled={this.state.disabled}
							onConfirm={this.onDelete}
						/>
					</div>
					<PageInput
						label="Name"
						help="Name of instance"
						type="text"
						placeholder="Enter name"
						value={instance.name}
						onChange={(val): void => {
							this.set('name', val);
						}}
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
					<label className="pt-label">
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
						buttonClass="pt-intent-success pt-icon-add"
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
					<label
						className="pt-label"
						style={css.label}
					>
						Vpcs
						<Help
							title="Vpcs"
							content="Vpcs attached to this instance."
						/>
						<div>
							{vpcs}
						</div>
					</label>
					<PageSelectButton
						label="Add Vpc"
						value={this.state.addVpc}
						disabled={!hasVpcs}
						buttonClass="pt-intent-success"
						onChange={(val: string): void => {
							this.setState({
								...this.state,
								addVpc: val,
							});
						}}
						onSubmit={this.onAddVpc}
					>
						{vpcsSelect}
					</PageSelectButton>
				</div>
				<div style={css.group}>
					<PageInfo
						fields={[
							{
								label: 'ID',
								value: this.props.instance.id || 'None',
							},
							{
								label: 'Organization',
								value: org ? org.name :
									this.props.instance.organization || 'None',
							},
							{
								label: 'Zone',
								value: zone ? zone.name : this.props.instance.zone || 'None',
							},
							{
								label: 'Node',
								value: info.node || 'None',
							},
							{
								label: 'State',
								value: (this.props.instance.state || 'None') + ':' + (
									this.props.instance.vm_state || 'None'),
							},
							{
								label: 'Public IPv4',
								value: this.props.instance.public_ip || 'None',
							},
							{
								label: 'Public IPv6',
								value: this.props.instance.public_ip6 || 'None',
							},
							{
								label: 'Disks',
								value: info.disks || '',
							},
							{
								label: 'Firewall Rules',
								value: this.props.instance.info.firewall_rules || '',
							},
							{
								label: 'Authorities',
								value: this.props.instance.info.authorities || '',
							},
						]}
					/>
				</div>
			</div>
			<PageSave
				style={css.save}
				hidden={!this.state.instance && !this.state.message}
				message={this.state.message}
				changed={this.state.changed}
				disabled={this.state.disabled}
				light={true}
				onCancel={(): void => {
					this.setState({
						...this.state,
						changed: false,
						forwardedChecked: false,
						instance: null,
					});
				}}
				onSave={this.onSave}
			>
				<ConfirmButton
					label="Start"
					className="pt-intent-success pt-icon-power"
					progressClassName="pt-intent-success"
					style={css.controlButton}
					hidden={this.props.instance.state !== 'stop'}
					disabled={this.state.disabled}
					onConfirm={(): void => {
						this.update('start');
					}}
				/>
				<ConfirmButton
					label="Stop"
					className="pt-intent-danger pt-icon-power"
					progressClassName="pt-intent-danger"
					style={css.controlButton}
					hidden={this.props.instance.state !== 'start'}
					disabled={this.state.disabled}
					onConfirm={(): void => {
						this.update('stop');
					}}
				/>
				<ConfirmButton
					label="Snapshot"
					className="pt-intent-primary pt-icon-floppy-disk"
					progressClassName="pt-intent-primary"
					hidden={this.props.instance.state !== 'running' &&
					this.props.instance.state !== 'stop'}
					style={css.controlButton}
					disabled={this.state.disabled}
					onConfirm={(): void => {
						this.update('snapshot');
					}}
				/>
			</PageSave>
		</td>;
	}
}

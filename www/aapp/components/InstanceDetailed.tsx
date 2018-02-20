/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as InstanceTypes from '../types/InstanceTypes';
import * as InstanceActions from '../actions/InstanceActions';
import InstancesStore from '../stores/InstancesStore';
import OrganizationsStore from '../stores/OrganizationsStore';
import DatacentersStore from '../stores/DatacentersStore';
import NodesStore from '../stores/NodesStore';
import ZonesStore from '../stores/ZonesStore';
import PageInput from './PageInput';
import PageInfo from './PageInfo';
import PageSave from './PageSave';
import PageSwitch from './PageSwitch';
import PageNumInput from './PageNumInput';
import ConfirmButton from './ConfirmButton';

interface Props {
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

		let org = OrganizationsStore.organization(
			this.props.instance.organization);
		let zone = ZonesStore.zone(this.props.instance.zone);
		let node = NodesStore.node(this.props.instance.node);

		let statusClass = '';
		switch (instance.status) {
			case 'Running':
				statusClass += 'pt-text-intent-success';
				break;
			case 'Stopped':
				statusClass += 'pt-text-intent-danger';
				break;
			case 'Snapshotting':
				statusClass += 'pt-text-intent-primary';
				break;
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
							className="pt-minimal pt-intent-danger pt-icon-cross open-ignore"
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
								value: node ? node.name : this.props.instance.node || 'None',
							},
							{
								label: 'State',
								value: this.props.instance.state || 'None',
							},
							{
								label: 'VM State',
								value: this.props.instance.vm_state || 'None',
							},
							{
								label: 'Public IPv4',
								value: this.props.instance.public_ip || 'None',
							},
							{
								label: 'Public IPv6',
								value: this.props.instance.public_ip6 || 'None',
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
					hidden={this.props.instance.state !== 'stopped'}
					disabled={this.state.disabled}
					onConfirm={(): void => {
						this.update('running');
					}}
				/>
				<ConfirmButton
					label="Stop"
					className="pt-intent-danger pt-icon-power"
					progressClassName="pt-intent-danger"
					style={css.controlButton}
					hidden={this.props.instance.state !== 'running'}
					disabled={this.state.disabled}
					onConfirm={(): void => {
						this.update('stopped');
					}}
				/>
				<ConfirmButton
					label="Snapshot"
					className="pt-intent-primary pt-icon-floppy-disk"
					progressClassName="pt-intent-primary"
					hidden={this.props.instance.state !== 'running' &&
					this.props.instance.state !== 'stopped'}
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

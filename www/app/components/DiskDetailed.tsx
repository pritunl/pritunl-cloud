/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as DiskTypes from '../types/DiskTypes';
import * as DiskActions from '../actions/DiskActions';
import * as OrganizationTypes from "../types/OrganizationTypes";
import PageInput from './PageInput';
import PageSelect from './PageSelect';
import PageNumInput from './PageNumInput';
import PageInfo from './PageInfo';
import PageSave from './PageSave';
import ConfirmButton from './ConfirmButton';
import NodesStore from "../stores/NodesStore";
import OrganizationsStore from "../stores/OrganizationsStore";
import * as InstanceActions from '../actions/InstanceActions';
import InstancesNodeStore from "../stores/InstancesNodeStore";
import * as InstanceTypes from "../types/InstanceTypes";

interface Props {
	organizations: OrganizationTypes.OrganizationsRo;
	disk: DiskTypes.DiskRo;
	selected: boolean;
	onSelect: (shift: boolean) => void;
	onClose: () => void;
}

interface State {
	disabled: boolean;
	changed: boolean;
	message: string;
	disk: DiskTypes.Disk;
	instances: InstanceTypes.InstancesRo;
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
	rules: {
		marginBottom: '15px',
	} as React.CSSProperties,
};

export default class DiskDetailed extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			disabled: false,
			changed: false,
			message: '',
			disk: null,
			instances: null,
		};
	}

	componentDidMount(): void {
		InstancesNodeStore.addChangeListener(this.onChange);
		InstanceActions.syncNode(this.props.disk.node);
	}

	componentWillUnmount(): void {
		InstancesNodeStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		this.setState({
			...this.state,
			instances: InstancesNodeStore.instances(this.props.disk.node),
		});
	}

	set(name: string, val: any): void {
		let disk: any;

		if (this.state.changed) {
			disk = {
				...this.state.disk,
			};
		} else {
			disk = {
				...this.props.disk,
			};
		}

		disk[name] = val;

		this.setState({
			...this.state,
			changed: true,
			disk: disk,
		});
	}

	onSave = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		DiskActions.commit(this.state.disk).then((): void => {
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
						disk: null,
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
		DiskActions.remove(this.props.disk.id).then((): void => {
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

	render(): JSX.Element {
		let disk: DiskTypes.Disk = this.state.disk ||
			this.props.disk;

		let node = NodesStore.node(this.props.disk.node);
		let org = OrganizationsStore.organization(this.props.disk.organization);

		let hasInstances = false;
		let instancesSelect: JSX.Element[] = [];
		if (this.state.instances === null) {
			instancesSelect = [<option key="null" value="">Loading</option>];
		} else {
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
		}

		let statusText = 'Unknown';
		let statusClass = '';
		switch (this.props.disk.state) {
			case 'provision':
				statusText = 'Provisioning';
				statusClass += ' pt-text-intent-primary';
				break;
			case 'available':
				if (this.props.disk.instance !== "") {
					statusText = 'Connected';
				} else {
					statusText = 'Available';
				}
				statusClass += ' pt-text-intent-success';
				break;
			case 'destroy':
				statusText = 'Destroying';
				statusClass += ' pt-text-intent-danger';
				break;
			case 'snapshot':
				statusText = 'Snapshotting';
				statusClass += ' pt-text-intent-primary';
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
								className="pt-icon-standard pt-icon-pulse"
							/>
							{statusText}
						</div>
						<div className="flex"/>
						<ConfirmButton
							className="pt-minimal pt-intent-danger pt-icon-trash open-ignore"
							style={css.button}
							progressClassName="pt-intent-danger"
							confirmMsg="Confirm disk remove"
							disabled={this.state.disabled}
							onConfirm={this.onDelete}
						/>
					</div>
					<PageInput
						label="Name"
						help="Name of disk."
						type="text"
						placeholder="Enter name"
						value={disk.name}
						onChange={(val): void => {
							this.set('name', val);
						}}
					/>
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
				</div>
				<div style={css.group}>
					<PageInfo
						fields={[
							{
								label: 'ID',
								value: this.props.disk.id || 'Unknown',
							},
							{
								label: 'Organization',
								value: org ? org.name : this.props.disk.organization,
							},
							{
								label: 'Node',
								value: node ? node.name : this.props.disk.node,
							},
							{
								label: 'Size',
								value: this.props.disk.size + 'GB',
							},
						]}
					/>
				</div>
			</div>
			<PageSave
				style={css.save}
				hidden={!this.state.disk && !this.state.message}
				message={this.state.message}
				changed={this.state.changed}
				disabled={this.state.disabled}
				light={true}
				onCancel={(): void => {
					this.setState({
						...this.state,
						changed: false,
						disk: null,
					});
				}}
				onSave={this.onSave}
			/>
		</td>;
	}
}

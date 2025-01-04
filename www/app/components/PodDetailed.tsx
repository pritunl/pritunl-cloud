/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Constants from '../Constants';
import * as PodTypes from '../types/PodTypes';
import * as PodActions from '../actions/PodActions';
import * as OrganizationTypes from "../types/OrganizationTypes";
import PodsStore from '../stores/PodsStore';
import PageInput from './PageInput';
import PageSelect from './PageSelect';
import PageInfo from './PageInfo';
import PageInputButton from './PageInputButton';
import PodWorkspace from './PodWorkspace';
import PageSave from './PageSave';
import ConfirmButton from './ConfirmButton';
import Help from './Help';
import PageTextArea from "./PageTextArea";

interface Props {
	organizations: OrganizationTypes.OrganizationsRo;
	pod: PodTypes.PodRo;
	selected: boolean;
	onSelect: (shift: boolean) => void;
	onClose: () => void;
}

interface State {
	disabled: boolean;
	changed: boolean;
	unitChanged: boolean;
	message: string;
	mode: string;
	pod: PodTypes.Pod;
}

const css = {
	card: {
		position: 'relative',
		padding: '48px 10px 0 10px',
		width: '100%',
		height: '1195px',
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
		height: '20px',
	} as React.CSSProperties,
	rules: {
		marginBottom: '15px',
	} as React.CSSProperties,
};

export default class PodDetailed extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			disabled: false,
			changed: false,
			unitChanged: false,
			message: '',
			mode: "view",
			pod: null,
		};
	}

	set(name: string, val: any): void {
		let pod: any;

		if (this.state.changed) {
			pod = {
				...this.state.pod,
			};
		} else {
			pod = {
				...this.props.pod,
			};
		}

		pod[name] = val;

		this.setState({
			...this.state,
			changed: true,
			pod: pod,
		});
	}

	onSave = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});

		let changed = false
		PodsStore.addChangeListen((): void => {
			changed = true
			if (!this.state.changed) {
				this.setState({
					...this.state,
					pod: null,
					changed: false,
					unitChanged: false,
					mode: this.state.mode === "edit" ? "view" : this.state.mode,
				});
			}
		});

		PodActions.commit(this.state.pod).then((): void => {
			this.setState({
				...this.state,
				message: 'Your changes have been saved',
				changed: false,
				unitChanged: false,
				disabled: false,
			});

			setTimeout((): void => {
				if (!changed && !this.state.changed) {
					this.setState({
						...this.state,
						message: '',
						pod: null,
						changed: false,
						unitChanged: false,
						mode: this.state.mode === "edit" ? "view" : this.state.mode,
					});
				} else {
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
		PodActions.remove(this.props.pod.id).then((): void => {
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
		let pod: PodTypes.Pod = this.state.pod ||
			this.props.pod;

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

		return <td
			className="bp5-cell"
			colSpan={5}
			style={css.card}
		>
			<div className="layout horizontal wrap">
				<div style={css.group}>
					<div
						className="layout horizontal tab-close"
						style={css.buttons}
						onClick={(evt): void => {
							let target = evt.target as HTMLElement;

							if (target.className.indexOf('tab-close') !== -1) {
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
						<ConfirmButton
							className="bp5-minimal bp5-intent-danger bp5-icon-trash"
							style={css.button}
							safe={true}
							progressClassName="bp5-intent-danger"
							dialogClassName="bp5-intent-danger bp5-icon-delete"
							dialogLabel="Delete Pod"
							confirmMsg="Permanently delete this pod"
							confirmInput={true}
							items={[pod.name]}
							disabled={this.state.disabled}
							onConfirm={this.onDelete}
						/>
					</div>
					<PageInput
						label="Name"
						help="Name of pod"
						type="text"
						placeholder="Enter name"
						value={pod.name}
						onChange={(val): void => {
							this.set('name', val);
						}}
					/>
					<PageSelect
						disabled={this.state.disabled || !hasOrganizations}
						hidden={Constants.user}
						label="Organization"
						help="Organization for pod."
						value={pod.organization}
						onChange={(val): void => {
							this.set('organization', val);
						}}
					>
						{organizationsSelect}
					</PageSelect>
				</div>
				<div style={css.group}>
					<PageInfo
						fields={[
							{
								label: 'ID',
								value: this.props.pod.id || 'Unknown',
							},
						]}
					/>
				</div>
			</div>
			<PodWorkspace
				pod={pod}
				disabled={this.state.disabled}
				unitChanged={this.state.unitChanged}
				mode={this.state.mode}
				onMode={(mode: string): void => {
					this.setState({
						...this.state,
						mode: mode,
					});
				}}
				onChange={(units: PodTypes.Unit[]): void => {
					let pod: any;

					if (this.state.changed) {
						pod = {
							...this.state.pod,
						};
					} else {
						pod = {
							...this.props.pod,
						};
					}

					pod.units = units

					this.setState({
						...this.state,
						changed: true,
						unitChanged: true,
						pod: pod,
					});
				}}
				onEdit={(units: PodTypes.Unit[]): void => {
					let pod: any;

					if (this.state.changed) {
						pod = {
							...this.state.pod,
						};
					} else {
						pod = {
							...this.props.pod,
						};
					}

					pod.units = units

					this.setState({
						...this.state,
						changed: true,
						unitChanged: true,
						mode: "edit",
						pod: pod,
					});
				}}
			/>
			<PageSave
				style={css.save}
				hidden={!this.state.pod && !this.state.message}
				message={this.state.message}
				changed={this.state.changed}
				disabled={this.state.disabled}
				light={true}
				onCancel={(): void => {
					this.setState({
						...this.state,
						changed: false,
						unitChanged: false,
						mode: "view",
						pod: null,
					});
				}}
				onSave={this.onSave}
			/>
		</td>;
	}
}

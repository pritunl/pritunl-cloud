/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Constants from '../Constants';
import * as Styles from '../Styles';
import * as PodTypes from '../types/PodTypes';
import * as PodActions from '../actions/PodActions';
import * as OrganizationTypes from "../types/OrganizationTypes";
import OrganizationsStore from '../stores/OrganizationsStore';
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
	mode: string;
	onMode: (mode: string) => void;
	settings: boolean;
	toggleSettings: () => void;
	sidebar: boolean;
	toggleSidebar: () => void;
}

interface State {
	disabled: boolean;
	changed: boolean;
	unitChanged: boolean;
	message: string;
	pod: PodTypes.Pod;
}

const css = {
	card: {
		position: 'relative',
		padding: '48px 10px 0 10px',
		width: '100%',
		height: 'calc(100dvh - 231px)',
	} as React.CSSProperties,
	button: {
		height: '30px',
	} as React.CSSProperties,
	buttons: {
		position: 'absolute',
		top: 0,
		left: 0,
		right: 0,
		padding: '4px',
		height: '39px',
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
	title: {
		cursor: 'pointer',
		margin: '3px',
	} as React.CSSProperties,
	save: {
		marginTop: '10px',
		marginBottom: '10px',
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
	draftsSyncTimeout: NodeJS.Timeout

	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			disabled: false,
			changed: false,
			unitChanged: false,
			message: '',
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
				});
				this.props.onMode(this.props.mode === "edit" ?
					"view" : this.props.mode);
			}
		});

		let pod = this.state.pod
		if (!pod && this.props.pod.drafts?.length) {
			pod = {
				...this.props.pod,
				units: this.props.pod.drafts,
			}
		}

		for (let unit of pod.units) {
			unit.deploy_spec = ""
		}

		PodActions.commit(pod).then((): void => {
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
					});
					this.props.onMode(this.props.mode === "edit" ?
						"view" : this.props.mode);
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

	onChangeCommit = (unitId: string, commit: string): void => {
		let pod: PodTypes.Pod

		if (this.state.changed) {
			pod = {
				...this.state.pod,
			};
			pod.units = [...(pod.units || [])]

			for (let i = 0; i < pod.units.length; i++) {
				if (pod.units[i].id === unitId) {
					pod.units[i].deploy_spec = commit
					break
				}
			}

			this.setState({
				...this.state,
				disabled: true,
				changed: true,
				unitChanged: true,
				pod: pod,
			})
		} else {
			pod = this.props.pod

			this.setState({
				...this.state,
				disabled: true,
			})
		}

		let deployPod: PodTypes.Pod = {
			...this.props.pod,
			units: [],
		}

		for (let i = 0; i < pod.units.length; i++) {
			if (pod.units[i].id === unitId) {
				deployPod.units = [{
					...pod.units[i],
					deploy_spec: commit,
				}]
				break
			}
		}

		PodActions.commitDeploy(deployPod).then((): void => {
			this.setState({
				...this.state,
				disabled: false,
			})
		}).catch((): void => {
			this.setState({
				...this.state,
				disabled: false,
			})
		})
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
		let pod: PodTypes.Pod
		let hasDrafts = !!this.props.pod.drafts?.length

		if (this.state.pod) {
			pod = this.state.pod
		} else if (hasDrafts) {
			pod = {
				...this.props.pod,
				units: this.props.pod.drafts,
			}
		} else {
			pod = this.props.pod
		}

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

		let orgName = '';
		if (pod.organization) {
			let org = OrganizationsStore.organization(pod.organization);
			orgName = org ? org.name : pod.organization;
		}

		return <div
			style={css.card}
			className="bp5-card layout vertical"
		>
			<div className="layout horizontal wrap">
				<div style={css.group}>
					<div
						className="layout horizontal bp5-card-header"
						style={css.buttons}
					>
						<button
							className={"bp5-button bp5-minimal " + (
								this.props.sidebar ? "bp5-icon-drawer-right" :
								"bp5-icon-drawer-left")}
							type="button"
							onClick={this.props.toggleSidebar}
						/>
						<div
							className="bp5-tag bp5-intent-primary no-select"
							style={css.title}
							onClick={this.props.toggleSettings}
						>
							<b>Name:</b>&nbsp;{pod.name}
						</div>
						<div
							hidden={!orgName}
							className="bp5-tag no-select"
							style={css.title}
							onClick={this.props.toggleSettings}
						>
							<b>Organization:</b>&nbsp;{orgName}
						</div>
						<div className="flex"/>
						<button
							className={"bp5-button bp5-minimal bp5-icon-cog" + (
								this.props.settings ? " bp5-intent-danger" : " bp5-intent-primary")}
							type="button"
							onClick={this.props.toggleSettings}
						>{this.props.settings ? "Close" : ""} Pod Settings</button>
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
						hidden={!this.props.settings}
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
						hidden={!this.props.settings || Constants.user}
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
						hidden={!this.props.settings}
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
				unitChanged={this.state.unitChanged || hasDrafts}
				mode={this.props.mode}
				onMode={(mode: string): void => {
					this.props.onMode(mode)
				}}
				onChangeCommit={this.onChangeCommit}
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

					let newMode = "view"
					for (let unit of units) {
						if (!unit.delete) {
							newMode = "edit"
						}
					}

					this.setState({
						...this.state,
						changed: true,
						unitChanged: true,
						pod: pod,
					});
					this.props.onMode(newMode)

					if (this.draftsSyncTimeout) {
						clearTimeout(this.draftsSyncTimeout)
					}

					this.draftsSyncTimeout = setTimeout(() => {
						PodActions.commitDrafts({
							...this.props.pod,
							drafts: units,
						})
						this.draftsSyncTimeout = null
					}, 500)
				}}
			/>
			<PageSave
				style={css.save}
				hidden={!this.state.pod && !this.state.message && !hasDrafts}
				message={this.state.message}
				changed={this.state.changed || hasDrafts}
				disabled={this.state.disabled}
				light={true}
				onCancel={(): void => {
					PodActions.commitDrafts({
						...this.props.pod,
						drafts: [],
					}, true)

					this.setState({
						...this.state,
						changed: false,
						unitChanged: false,
						pod: null,
					});
					this.props.onMode("view")
				}}
				onSave={this.onSave}
			/>
		</div>;
	}
}

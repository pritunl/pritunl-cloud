/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Constants from '../Constants';
import * as PodTypes from '../types/PodTypes';
import * as PodActions from '../actions/PodActions';
import * as OrganizationTypes from "../types/OrganizationTypes";
import PageInput from './PageInput';
import PageSelect from './PageSelect';
import PageInfo from './PageInfo';
import PageInputButton from './PageInputButton';
import PageSave from './PageSave';
import ConfirmButton from './ConfirmButton';
import Help from './Help';
import PageTextArea from "./PageTextArea";
import * as DomainTypes from "../types/DomainTypes";
import * as VpcTypes from "../types/VpcTypes";
import * as DatacenterTypes from "../types/DatacenterTypes";
import * as NodeTypes from "../types/NodeTypes";
import * as PoolTypes from "../types/PoolTypes";
import * as ZoneTypes from "../types/ZoneTypes";
import * as ShapeTypes from "../types/ShapeTypes";
import * as Theme from "../Theme";

import AceEditor from "react-ace"
import "ace-builds/src-noconflict/mode-text"
import "ace-builds/src-noconflict/mode-markdown"
import "ace-builds/src-noconflict/mode-sh"
import "ace-builds/src-noconflict/mode-python"
import "ace-builds/src-noconflict/theme-dracula"
import "ace-builds/src-noconflict/theme-eclipse"
import {Ace} from "ace-builds"

interface Props {
	organizations: OrganizationTypes.OrganizationsRo;
	domains: DomainTypes.DomainsRo;
	vpcs: VpcTypes.VpcsRo;
	datacenters: DatacenterTypes.DatacentersRo;
	nodes: NodeTypes.NodesRo;
	pools: PoolTypes.PoolsRo;
	zones: ZoneTypes.ZonesRo;
	shapes: ShapeTypes.ShapesRo;
	pod: PodTypes.PodRo;
	selected: boolean;
	onSelect: (shift: boolean) => void;
	onClose: () => void;
}

interface State {
	disabled: boolean;
	changed: boolean;
	message: string;
	addRole: string;
	pod: PodTypes.Pod;
}

const example = `test`

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
	editorBox: {
	} as React.CSSProperties,
	editor: {
		margin: '10px 0',
	} as React.CSSProperties,
};

export default class PodDetailed extends React.Component<Props, State> {
	editor: Ace.Editor

	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			disabled: false,
			changed: false,
			message: '',
			addRole: null,
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

	onAddRole = (): void => {
		let pod: PodTypes.Pod;

		if (!this.state.addRole) {
			return;
		}

		if (this.state.changed) {
			pod = {
				...this.state.pod,
			};
		} else {
			pod = {
				...this.props.pod,
			};
		}

		let roles = [
			...(pod.roles || []),
		];


		if (roles.indexOf(this.state.addRole) === -1) {
			roles.push(this.state.addRole);
		}

		roles.sort();
		pod.roles = roles;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addRole: '',
			pod: pod,
		});
	}

	onRemoveRole = (role: string): void => {
		let pod: PodTypes.Pod;

		if (this.state.changed) {
			pod = {
				...this.state.pod,
			};
		} else {
			pod = {
				...this.props.pod,
			};
		}

		let roles = [
			...(pod.roles || []),
		];

		let i = roles.indexOf(role);
		if (i === -1) {
			return;
		}

		roles.splice(i, 1);
		pod.roles = roles;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addRole: '',
			pod: pod,
		});
	}

	onSave = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		PodActions.commit(this.state.pod).then((): void => {
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
						pod: null,
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

		let roles: JSX.Element[] = [];
		for (let role of (pod.roles || [])) {
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
					<PageTextArea
						label="Comment"
						help="Pod comment."
						placeholder="Pod comment"
						rows={3}
						value={pod.comment}
						onChange={(val: string): void => {
							this.set('comment', val);
						}}
					/>
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
					<label className="bp5-label">
						Roles
						<Help
							title="Roles"
							content="Roles that will be matched with firewall rules. Network roles are case-sensitive."
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
			<div className="layout horizontal flex" style={css.editorBox}>
				<label
					className="bp5-label flex"
					style={css.editor}
				>
					<AceEditor
						name="log-view"
						theme={Theme.editorTheme()}
						height="500px"
						width="100%"
						mode="python"
						fontSize="12px"
						wrapEnabled={true}
						showPrintMargin={false}
						showGutter={true}
						readOnly={false}
						value={example}
						editorProps={{
							$blockScrolling: true,
						}}
						setOptions={{
							showFoldWidgets: false,
						}}
						onLoad={(editor: Ace.Editor): void => {
							this.editor = editor
							this.editor.scrollToLine(Number.POSITIVE_INFINITY,
								false, false, null)
						}}
					/>
				</label>
			</div>
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
						pod: null,
					});
				}}
				onSave={this.onSave}
			/>
		</td>;
	}
}

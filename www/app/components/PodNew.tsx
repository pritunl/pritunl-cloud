/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Constants from '../Constants';
import * as PodTypes from '../types/PodTypes';
import * as OrganizationTypes from '../types/OrganizationTypes';
import * as PodActions from '../actions/PodActions';
import * as MiscUtils from '../utils/MiscUtils';
import PodsStore from '../stores/PodsStore';
import PageInput from './PageInput';
import PageInputButton from './PageInputButton';
import PageCreate from './PageCreate';
import PageSelect from './PageSelect';
import PageSwitch from "./PageSwitch";
import PageNumInput from './PageNumInput';
import PodWorkspace from './PodWorkspace';
import Help from './Help';
import PageTextArea from "./PageTextArea";

interface Props {
	organizations: OrganizationTypes.OrganizationsRo;
	onClose: () => void;
}

interface State {
	closed: boolean;
	disabled: boolean;
	changed: boolean;
	unitChanged: boolean;
	message: string;
	mode: string;
	pod: PodTypes.Pod;
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
		height: '20px',
	} as React.CSSProperties,
};

export default class PodNew extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			closed: false,
			disabled: false,
			changed: false,
			unitChanged: false,
			message: '',
			mode: "view",
			pod: this.default,
		};
	}

	componentDidMount(): void {
	}

	componentWillUnmount(): void {
	}

	onChange = (): void => {
		this.setState({
			...this.state,
		});
	}

	get default(): PodTypes.Pod {
		return {
			id: null,
			name: 'New pod',
			units: [
				{
					id: MiscUtils.objectId(),
					name: "new-unit",
					spec: "",
				}
			]
		};
	}

	set(name: string, val: any): void {
		let pod: any = {
			...this.state.pod,
		};

		pod[name] = val;

		this.setState({
			...this.state,
			changed: true,
			pod: pod,
		});
	}

	onCreate = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});

		let pod: any = {
			...this.state.pod,
		};

		PodActions.create(pod).then((): void => {
			this.setState({
				...this.state,
				message: 'Pod created successfully',
				changed: false,
				unitChanged: false,
			});
		}).catch((): void => {
			this.setState({
				...this.state,
				message: '',
				disabled: false,
			});
		});
	}

	render(): JSX.Element {
		let pod = this.state.pod;

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

		return <div
			className="bp5-card bp5-row"
			style={css.row}
		>
			<td
				className="bp5-cell"
				colSpan={6}
				style={css.card}
			>
				<div className="layout horizontal wrap">
					<div style={css.group}>
						<div style={css.buttons}>
						</div>
						<PageInput
							label="Name"
							help="Name of pod. String formatting such as %d or %02d can be used to add the pod number or zero padded number."
							type="text"
							placeholder="Enter name"
							disabled={this.state.disabled}
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
						<PageSwitch
							disabled={this.state.disabled}
							label="Delete protection"
							help="Block pod and any attached disks from being deleted."
							checked={pod.delete_protection}
							onToggle={(): void => {
								this.set('delete_protection', !pod.delete_protection);
							}}
						/>
					</div>
					<div style={css.group}>
					</div>
				</div>
				<PageCreate
					style={css.save}
					hidden={!this.state.pod}
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

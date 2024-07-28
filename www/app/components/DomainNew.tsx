/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as DiskTypes from '../types/DiskTypes';
import * as OrganizationTypes from '../types/OrganizationTypes';
import * as DatacenterTypes from '../types/DatacenterTypes';
import * as NodeTypes from '../types/NodeTypes';
import * as InstanceTypes from '../types/InstanceTypes';
import * as ImageTypes from '../types/ImageTypes';
import * as ZoneTypes from '../types/ZoneTypes';
import * as PoolTypes from '../types/PoolTypes';
import * as DomainActions from '../actions/DomainActions';
import * as ImageActions from '../actions/ImageActions';
import * as InstanceActions from '../actions/InstanceActions';
import * as NodeActions from '../actions/NodeActions';
import ImagesDatacenterStore from '../stores/ImagesDatacenterStore';
import InstancesNodeStore from '../stores/InstancesNodeStore';
import NodesZoneStore from '../stores/NodesZoneStore';
import PageInput from './PageInput';
import PageInputButton from './PageInputButton';
import PageCreate from './PageCreate';
import PageSelect from './PageSelect';
import PageSwitch from "./PageSwitch";
import PageNumInput from './PageNumInput';
import Help from './Help';
import * as SecretTypes from "../types/SecretTypes";
import * as DomainTypes from "../types/DomainTypes";
import * as Constants from "../Constants";
import OrganizationsStore from "../stores/OrganizationsStore";
import PageTextArea from "./PageTextArea";
import PageInfo from "./PageInfo";

interface Props {
	organizations: OrganizationTypes.OrganizationsRo;
	secrets: SecretTypes.SecretsRo;
	onClose: () => void;
}

interface State {
	closed: boolean;
	disabled: boolean;
	changed: boolean;
	message: string;
	domain: DomainTypes.Domain;
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

export default class DomainNew extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			closed: false,
			disabled: false,
			changed: false,
			message: '',
			domain: {},
		};
	}

	set(name: string, val: any): void {
		let domain: any = {
			...this.state.domain,
		};

		domain[name] = val;

		this.setState({
			...this.state,
			changed: true,
			domain: domain,
		});
	}

	onCreate = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});

		let domain: any = {
			...this.state.domain,
		};

		if (this.props.organizations.length && !domain.organization) {
			domain.organization = this.props.organizations[0].id;
		}

		DomainActions.create(domain).then((): void => {
			this.setState({
				...this.state,
				message: 'Domain created successfully',
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

	render(): JSX.Element {
		let domain = this.state.domain;
		if (this.props.organizations.length && !domain.organization) {
			domain.organization = this.props.organizations[0].id;
		}

		let hasOrganizations = false
		let organizationsSelect: JSX.Element[] = [];
		if (this.props.organizations.length) {
			for (let organization of this.props.organizations) {
				hasOrganizations = true
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

		let hasSecrets = false;
		let secretsSelect: JSX.Element[] = [];
		if (this.props.secrets.length) {
			secretsSelect.push(<option key="null" value="">Select Secret</option>);

			for (let secret of this.props.secrets) {
				if (Constants.user) {
					if (domain.organization !== OrganizationsStore.current) {
						continue;
					}
				} else {
					if (domain.organization !== secret.organization) {
						continue;
					}
				}

				hasSecrets = true;
				secretsSelect.push(
					<option
						key={secret.id}
						value={secret.id}
					>{secret.name}</option>,
				);
			}
		}

		if (!hasSecrets) {
			secretsSelect = [<option key="null" value="">No Secrets</option>];
		}

		return <div
			className="bp5-card bp5-row"
			style={css.row}
		>
			<td
				className="bp5-cell"
				colSpan={5}
				style={css.card}
			>
				<div className="layout horizontal wrap">
					<div style={css.group}>
						<div style={css.buttons}>
						</div>
						<PageInput
							label="Name"
							help="Domain name."
							type="text"
							placeholder="Enter name"
							value={domain.name}
							onChange={(val): void => {
								this.set('name', val);
							}}
						/>
						<PageTextArea
							label="Comment"
							help="Domain comment."
							placeholder="Domain comment"
							rows={3}
							value={domain.comment}
							onChange={(val: string): void => {
								this.set('comment', val);
							}}
						/>
						<PageSelect
							label="Provider"
							help="Domain provider."
							value={domain.type}
							onChange={(val): void => {
								this.set('type', val);
							}}
						>
							<option value="aws">AWS</option>
							<option value="cloudflare">Cloudflare</option>
							<option value="oracle_cloud">Oracle Cloud</option>
						</PageSelect>
						<PageInput
							label="Domain"
							help="Root domain name."
							type="text"
							placeholder="Enter domain"
							value={domain.root_domain}
							onChange={(val): void => {
								this.set('root_domain', val);
							}}
						/>
						<PageSelect
							disabled={this.state.disabled}
							label="Provider API Secret"
							help="Secret containing API keys to use for provider."
							value={domain.secret}
							onChange={(val): void => {
								this.set('secret', val);
							}}
						>
							{secretsSelect}
						</PageSelect>
					</div>
					<div style={css.group}>
						<PageSelect
							disabled={this.state.disabled}
							hidden={Constants.user}
							label="Organization"
							help="Organization for domain."
							value={domain.organization}
							onChange={(val): void => {
								this.set('organization', val);
							}}
						>
							{organizationsSelect}
						</PageSelect>
					</div>
				</div>
				<PageCreate
					style={css.save}
					hidden={!this.state.domain}
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

/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Constants from '../Constants';
import * as ServiceTypes from '../types/ServiceTypes';
import * as OrganizationTypes from '../types/OrganizationTypes';
import * as DomainTypes from '../types/DomainTypes';
import * as DatacenterTypes from '../types/DatacenterTypes';
import * as NodeTypes from '../types/NodeTypes';
import * as VpcTypes from '../types/VpcTypes';
import * as PoolTypes from '../types/PoolTypes';
import * as ZoneTypes from '../types/ZoneTypes';
import * as ShapeTypes from '../types/ShapeTypes';
import * as ServiceActions from '../actions/ServiceActions';
import PageInput from './PageInput';
import PageInputButton from './PageInputButton';
import PageCreate from './PageCreate';
import PageSelect from './PageSelect';
import PageSwitch from "./PageSwitch";
import PageNumInput from './PageNumInput';
import ServiceEditor from "./ServiceEditor";
import Help from './Help';
import OrganizationsStore from "../stores/OrganizationsStore";
import PageTextArea from "./PageTextArea";
import ShapesStore from "../stores/ShapesStore";

interface Props {
	organizations: OrganizationTypes.OrganizationsRo;
	domains: DomainTypes.DomainsRo;
	vpcs: VpcTypes.VpcsRo;
	datacenters: DatacenterTypes.DatacentersRo;
	nodes: NodeTypes.NodesRo;
	pools: PoolTypes.PoolsRo;
	zones: ZoneTypes.ZonesRo;
	shapes: ShapeTypes.ShapesRo;
	onClose: () => void;
}

interface State {
	closed: boolean;
	disabled: boolean;
	changed: boolean;
	message: string;
	service: ServiceTypes.Service;
	addRole: string;
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

export default class ServiceNew extends React.Component<Props, State> {
	imagesMap: Map<string, string>;

	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			closed: false,
			disabled: false,
			changed: false,
			message: '',
			service: this.default,
			addRole: '',
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

	get default(): ServiceTypes.Service {
		return {
			id: null,
			name: 'New service',
		};
	}

	set(name: string, val: any): void {
		let service: any = {
			...this.state.service,
		};

		service[name] = val;

		this.setState({
			...this.state,
			changed: true,
			service: service,
		});
	}

	onCreate = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});

		let service: any = {
			...this.state.service,
		};

		if (this.props.organizations.length && !service.organization) {
			service.organization = this.props.organizations[0].id;
		}

		ServiceActions.create(service).then((): void => {
			this.setState({
				...this.state,
				message: 'Service created successfully',
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

	onAddRole = (): void => {
		if (!this.state.addRole) {
			return;
		}

		let service = {
			...this.state.service,
		};

		let roles = [
			...(service.roles || []),
		];

		if (roles.indexOf(this.state.addRole) === -1) {
			roles.push(this.state.addRole);
		}

		roles.sort();
		service.roles = roles;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addRole: '',
			service: service,
		});
	}

	onRemoveRole = (role: string): void => {
		let service = {
			...this.state.service,
		};

		let roles = [
			...(service.roles || []),
		];

		let i = roles.indexOf(role);
		if (i === -1) {
			return;
		}

		roles.splice(i, 1);
		service.roles = roles;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addRole: '',
			service: service,
		});
	}

	render(): JSX.Element {
		let service = this.state.service;

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
		for (let role of (service.roles || [])) {
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
							help="Name of service. String formatting such as %d or %02d can be used to add the service number or zero padded number."
							type="text"
							placeholder="Enter name"
							disabled={this.state.disabled}
							value={service.name}
							onChange={(val): void => {
								this.set('name', val);
							}}
						/>
						<PageTextArea
							label="Comment"
							help="Service comment."
							placeholder="Service comment"
							rows={3}
							value={service.comment}
							onChange={(val: string): void => {
								this.set('comment', val);
							}}
						/>
						<PageSelect
							disabled={this.state.disabled || !hasOrganizations}
							hidden={Constants.user}
							label="Organization"
							help="Organization for service."
							value={service.organization}
							onChange={(val): void => {
								this.set('organization', val);
							}}
						>
							{organizationsSelect}
						</PageSelect>
						<PageSwitch
							disabled={this.state.disabled}
							label="Delete protection"
							help="Block service and any attached disks from being deleted."
							checked={service.delete_protection}
							onToggle={(): void => {
								this.set('delete_protection', !service.delete_protection);
							}}
						/>
					</div>
					<div style={css.group}>
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
				<ServiceEditor
					disabled={this.state.disabled}
					defaultEdit={true}
					value={service.spec}
					onChange={(val: string): void => {
						this.set("spec", val)
					}}
				/>
				<PageCreate
					style={css.save}
					hidden={!this.state.service}
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

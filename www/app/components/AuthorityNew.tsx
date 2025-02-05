/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as AuthorityTypes from '../types/AuthorityTypes';
import * as AuthorityActions from '../actions/AuthorityActions';
import * as OrganizationTypes from "../types/OrganizationTypes";
import PageInput from './PageInput';
import PageSelect from './PageSelect';
import PageInfo from './PageInfo';
import PageInputButton from './PageInputButton';
import PageTextArea from './PageTextArea';
import PageCreate from './PageCreate';
import ConfirmButton from './ConfirmButton';
import Help from './Help';
import * as Constants from "../Constants";

interface Props {
	organizations: OrganizationTypes.OrganizationsRo;
	onClose: () => void;
}

interface State {
	closed: boolean;
	disabled: boolean;
	changed: boolean;
	message: string;
	addRole: string;
	addNetworkRole: string;
	authority: AuthorityTypes.Authority;
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
	button: {
		height: '30px',
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

export default class AuthorityNew extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			closed: false,
			disabled: false,
			changed: false,
			message: '',
			addRole: null,
			addNetworkRole: null,
			authority: {
				name: "new-authority",
			},
		};
	}

	set(name: string, val: any): void {
		let authority: any = {
			...this.state.authority,
		};

		authority[name] = val;

		this.setState({
			...this.state,
			changed: true,
			authority: authority,
		});
	}

	onAddRole = (): void => {
		let authority: AuthorityTypes.Authority;

		authority = {
			...this.state.authority,
		};

		if (!this.state.addRole) {
			return;
		}

		let roles = [
			...(authority.roles || []),
		];

		if (roles.indexOf(this.state.addRole) === -1) {
			roles.push(this.state.addRole);
		}

		roles.sort();
		authority.roles = roles;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addRole: '',
			authority: authority,
		});
	}

	onRemoveRole = (role: string): void => {
		let authority: AuthorityTypes.Authority;

		authority = {
			...this.state.authority,
		};

		let roles = [
			...(authority.roles || []),
		];

		let i = roles.indexOf(role);
		if (i === -1) {
			return;
		}

		roles.splice(i, 1);
		authority.roles = roles;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addRole: '',
			authority: authority,
		});
	}

	onAddNetworkRole = (): void => {
		let authority: AuthorityTypes.Authority;

		authority = {
			...this.state.authority,
		};

		if (!this.state.addNetworkRole) {
			return;
		}

		let networkRoles = [
			...(authority.network_roles || []),
		];


		if (networkRoles.indexOf(this.state.addNetworkRole) === -1) {
			networkRoles.push(this.state.addNetworkRole);
		}

		networkRoles.sort();
		authority.network_roles = networkRoles;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addNetworkRole: '',
			authority: authority,
		});
	}

	onRemoveNetworkRole = (networkRole: string): void => {
		let authority: AuthorityTypes.Authority;

		authority = {
			...this.state.authority,
		};

		let networkRoles = [
			...(authority.network_roles || []),
		];

		let i = networkRoles.indexOf(networkRole);
		if (i === -1) {
			return;
		}

		networkRoles.splice(i, 1);
		authority.network_roles = networkRoles;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addNetworkRole: '',
			authority: authority,
		});
	}

	onCreate = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});

		let authority: any = {
			...this.state.authority,
		};

		if (this.props.organizations.length && !authority.organization) {
			authority.organization = this.props.organizations[0].id;
		}

		AuthorityActions.create(authority).then((): void => {
			this.setState({
				...this.state,
				message: 'Authority created successfully',
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
		let authority: AuthorityTypes.Authority = this.state.authority;

		let hasOrganizations = !!this.props.organizations.length;
		let organizationsSelect: JSX.Element[] = [];
		(this.props.organizations || []).forEach((org) => {
			organizationsSelect.push(
				<option
					key={org.id}
					value={org.id}
				>{org.name}</option>,
			);
		})

		if (!hasOrganizations) {
			organizationsSelect.push(
				<option key="null" value="">No Organizations</option>);
		}

		let roles: JSX.Element[] = [];
		(authority.roles || []).forEach((role) => {
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
		})

		let networkRoles: JSX.Element[] = [];
		(authority.network_roles || []).forEach((role) => {
			networkRoles.push(
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
							this.onRemoveNetworkRole(role);
						}}
					/>
				</div>,
			);
		})

		return <div
			className="bp5-card bp5-row"
			style={css.row}
		>
			<td
				className="bp5-cell"
				colSpan={3}
				style={css.card}
			>
				<div className="layout horizontal wrap">
					<div style={css.group}>
						<div style={css.buttons}>
						</div>
						<PageInput
							label="Name"
							help="Name of authority"
							type="text"
							placeholder="Enter name"
							value={authority.name}
							onChange={(val): void => {
								this.set('name', val);
							}}
						/>
						<PageTextArea
							label="Comment"
							help="Authority comment."
							placeholder="Authority comment"
							rows={3}
							value={authority.comment}
							onChange={(val: string): void => {
								this.set('comment', val);
							}}
						/>
						<PageSelect
							label="Type"
							help="Authority type. SSH keys will be saved to ~/.ssh/authorized_keys. SSH certificates will be saved to the SSH server configuration."
							value={authority.type}
							onChange={(val): void => {
								this.set('type', val);
							}}
						>
							<option value="ssh_key">SSH Key</option>
							<option value="ssh_certificate">SSH Certificate</option>
						</PageSelect>
						<PageTextArea
							hidden={authority.type !== 'ssh_key'}
							label="SSH Key"
							help="SSH authorized public key in PEM format."
							placeholder="Public key"
							rows={6}
							value={authority.key}
							onChange={(val: string): void => {
								this.set('key', val);
							}}
						/>
						<PageTextArea
							hidden={authority.type !== 'ssh_certificate'}
							label="SSH Certificate"
							help="SSH certificate authority in PEM format."
							placeholder="Certificate authority"
							rows={6}
							value={authority.certificate}
							onChange={(val: string): void => {
								this.set('certificate', val);
							}}
						/>
						<label
							className="bp5-label"
							hidden={authority.type !== 'ssh_certificate'}
						>
							Roles
							<Help
								title="Roles"
								content="Roles that will be matched with authority principles. Roles are case-sensitive."
							/>
							<div>
								{roles}
							</div>
						</label>
						<PageInputButton
							disabled={this.state.disabled}
							buttonClass="bp5-intent-success bp5-icon-add"
							hidden={authority.type !== 'ssh_certificate'}
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
					<div style={css.group}>
						<PageSelect
							disabled={this.state.disabled || !hasOrganizations}
							hidden={Constants.user}
							label="Organization"
							help="Organization for authority, both the organaization and role must match. Select node authority to match node network roles."
							value={authority.organization}
							onChange={(val): void => {
								this.set('organization', val);
							}}
						>
							{organizationsSelect}
						</PageSelect>
						<label className="bp5-label">
							Network Roles
							<Help
								title="Network Roles"
								content="Network roles that will be matched with authorities. Network roles are case-sensitive."
							/>
							<div>
								{networkRoles}
							</div>
						</label>
						<PageInputButton
							disabled={this.state.disabled}
							buttonClass="bp5-intent-success bp5-icon-add"
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
					</div>
				</div>
				<PageCreate
					style={css.save}
					hidden={!this.state.authority}
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

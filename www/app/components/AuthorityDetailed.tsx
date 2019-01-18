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
import PageSave from './PageSave';
import ConfirmButton from './ConfirmButton';
import Help from './Help';
import * as Constants from "../Constants";

interface Props {
	organizations: OrganizationTypes.OrganizationsRo;
	authority: AuthorityTypes.AuthorityRo;
	selected: boolean;
	onSelect: (shift: boolean) => void;
	onClose: () => void;
}

interface State {
	disabled: boolean;
	changed: boolean;
	message: string;
	addRole: string;
	addNetworkRole: string;
	authority: AuthorityTypes.Authority;
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
	} as React.CSSProperties,
	role: {
		margin: '9px 5px 0 5px',
		height: '20px',
	} as React.CSSProperties,
	rules: {
		marginBottom: '15px',
	} as React.CSSProperties,
};

export default class AuthorityDetailed extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			disabled: false,
			changed: false,
			message: '',
			addRole: null,
			addNetworkRole: null,
			authority: null,
		};
	}

	set(name: string, val: any): void {
		let authority: any;

		if (this.state.changed) {
			authority = {
				...this.state.authority,
			};
		} else {
			authority = {
				...this.props.authority,
			};
		}

		authority[name] = val;

		this.setState({
			...this.state,
			changed: true,
			authority: authority,
		});
	}

	onAddRole = (): void => {
		let authority: AuthorityTypes.Authority;

		if (!this.state.addRole) {
			return;
		}

		if (this.state.changed) {
			authority = {
				...this.state.authority,
			};
		} else {
			authority = {
				...this.props.authority,
			};
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

		if (this.state.changed) {
			authority = {
				...this.state.authority,
			};
		} else {
			authority = {
				...this.props.authority,
			};
		}

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

		if (!this.state.addNetworkRole) {
			return;
		}

		if (this.state.changed) {
			authority = {
				...this.state.authority,
			};
		} else {
			authority = {
				...this.props.authority,
			};
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

		if (this.state.changed) {
			authority = {
				...this.state.authority,
			};
		} else {
			authority = {
				...this.props.authority,
			};
		}

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

	onSave = (): void => {
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

		AuthorityActions.commit(authority).then((): void => {
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
						authority: null,
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
		AuthorityActions.remove(this.props.authority.id).then((): void => {
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
		let authority: AuthorityTypes.Authority = this.state.authority ||
			this.props.authority;

		let hasOrganizations = !!this.props.organizations.length;
		let organizationsSelect: JSX.Element[] = [];
		if (this.props.organizations.length) {
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
		for (let role of (authority.roles || [])) {
			roles.push(
				<div
					className="pt-tag pt-tag-removable pt-intent-primary"
					style={css.role}
					key={role}
				>
					{role}
					<button
						className="pt-tag-remove"
						disabled={this.state.disabled}
						onMouseUp={(): void => {
							this.onRemoveRole(role);
						}}
					/>
				</div>,
			);
		}

		let networkRoles: JSX.Element[] = [];
		for (let networkRole of (authority.network_roles || [])) {
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
						<div className="flex"/>
						<ConfirmButton
							className="pt-minimal pt-intent-danger pt-icon-trash open-ignore"
							style={css.button}
							progressClassName="pt-intent-danger"
							confirmMsg="Confirm authority remove"
							disabled={this.state.disabled}
							onConfirm={this.onDelete}
						/>
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
						className="pt-label"
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
						buttonClass="pt-intent-success pt-icon-add"
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
					<PageInfo
						fields={[
							{
								label: 'ID',
								value: this.props.authority.id || 'Unknown',
							},
						]}
					/>
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
					<label className="pt-label">
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
				</div>
			</div>
			<PageSave
				style={css.save}
				hidden={!this.state.authority && !this.state.message}
				message={this.state.message}
				changed={this.state.changed}
				disabled={this.state.disabled}
				light={true}
				onCancel={(): void => {
					this.setState({
						...this.state,
						changed: false,
						authority: null,
					});
				}}
				onSave={this.onSave}
			/>
		</td>;
	}
}

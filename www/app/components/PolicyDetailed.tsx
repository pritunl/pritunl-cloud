/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as PolicyTypes from '../types/PolicyTypes';
import * as SettingsTypes from '../types/SettingsTypes';
import * as PolicyActions from '../actions/PolicyActions';
import PolicyRule from './PolicyRule';
import PageInput from './PageInput';
import PageSwitch from './PageSwitch';
import PageSelect from './PageSelect';
import PageInputButton from './PageInputButton';
import PageInfo from './PageInfo';
import PageSave from './PageSave';
import ConfirmButton from './ConfirmButton';
import Help from './Help';
import * as Alert from '../Alert';
import PageTextArea from "./PageTextArea";

interface Props {
	policy: PolicyTypes.PolicyRo;
	providers: SettingsTypes.SecondaryProviders;
	selected: boolean;
	onSelect: (shift: boolean) => void;
	onClose: () => void;
}

interface State {
	disabled: boolean;
	changed: boolean;
	message: string;
	policy: PolicyTypes.Policy;
	addRole: string;
}

const css = {
	card: {
		position: 'relative',
		padding: '48px 10px 0 10px',
		width: '100%',
	} as React.CSSProperties,
	remove: {
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
	select: {
		margin: '7px 0px 0px 6px',
		paddingTop: '3px',
	} as React.CSSProperties,
};

export default class PolicyDetailed extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			disabled: false,
			changed: false,
			message: '',
			policy: null,
			addRole: null,
		};
	}

	set(name: string, val: any): void {
		let policy: any;

		if (this.state.changed) {
			policy = {
				...this.state.policy,
			};
		} else {
			policy = {
				...this.props.policy,
			};
		}

		policy[name] = val;

		this.setState({
			...this.state,
			changed: true,
			policy: policy,
		});
	}

	setRule(name: string, rule: PolicyTypes.Rule): void {
		let policy: any;

		if (this.state.changed) {
			policy = {
				...this.state.policy,
			};
		} else {
			policy = {
				...this.props.policy,
			};
		}

		let rules = {
			...(policy.rules || {}),
		};

		if (rule.values == null) {
			delete rules[name];
		} else {
			rules[name] = rule;
		}

		policy.rules = rules;

		this.setState({
			...this.state,
			changed: true,
			policy: policy,
		});
	}

	onSave = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		PolicyActions.commit(this.state.policy).then((): void => {
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
						message: '',
						changed: false,
						policy: null,
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
		PolicyActions.remove(this.props.policy.id).then((): void => {
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

	onAddRole = (): void => {
		let policy: PolicyTypes.Policy;

		if (this.state.changed) {
			policy = {
				...this.state.policy,
			};
		} else {
			policy = {
				...this.props.policy,
			};
		}

		let roles = [
			...(policy.roles || []),
		];

		if (!this.state.addRole) {
			return;
		}

		if (roles.indexOf(this.state.addRole) === -1) {
			roles.push(this.state.addRole);
		}

		roles.sort();

		policy.roles = roles;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addRole: '',
			policy: policy,
		});
	}

	onRemoveRole(role: string): void {
		let policy: PolicyTypes.Policy;

		if (this.state.changed) {
			policy = {
				...this.state.policy,
			};
		} else {
			policy = {
				...this.props.policy,
			};
		}

		let roles = [
			...(policy.roles || []),
		];

		let i = roles.indexOf(role);
		if (i === -1) {
			return;
		}

		roles.splice(i, 1);

		policy.roles = roles;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addRole: '',
			policy: policy,
		});
	}

	render(): JSX.Element {
		let policy: PolicyTypes.Policy = this.state.policy ||
			this.props.policy;

		let roles: JSX.Element[] = [];
		for (let role of (policy.roles || [])) {
			roles.push(
				<div
					className="bp5-tag bp5-tag-removable bp5-intent-primary"
					style={css.item}
					key={role}
				>
					{role}
					<button
						className="bp5-tag-remove"
						onMouseUp={(): void => {
							this.onRemoveRole(role);
						}}
					/>
				</div>,
			);
		}

		let rules = policy.rules || {};
		let operatingSystem = rules.operating_system || {
			type: 'operating_system',
		};
		let browser = rules.browser || {
			type: 'browser',
		};
		let location = rules.location || {
			type: 'location',
		};
		let whitelistNetworks = rules.whitelist_networks || {
			type: 'whitelist_networks',
		};
		let blacklistNetworks = rules.blacklist_networks || {
			type: 'blacklist_networks',
		};

		let providerIds: string[] = [];
		let adminProviders: JSX.Element[] = [];
		let userProviders: JSX.Element[] = [];
		if (this.props.providers.length) {
			for (let provider of this.props.providers) {
				providerIds.push(provider.id);
				adminProviders.push(<option
					key={provider.id}
					value={provider.id}
				>{provider.name}</option>);
				userProviders.push(<option
					key={provider.id}
					value={provider.id}
				>{provider.name}</option>);
			}
		} else {
			adminProviders.push(<option
				key="null"
				value=""
			>None</option>);
			userProviders.push(<option
				key="null"
				value=""
			>None</option>);
		}
		let adminProvider = policy.admin_secondary &&
			providerIds.indexOf(policy.admin_secondary) !== -1;
		let userProvider = policy.user_secondary &&
			providerIds.indexOf(policy.user_secondary) !== -1;

		return <td
			className="bp5-cell"
			colSpan={2}
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
							dialogLabel="Delete Policy"
							confirmMsg="Permanently delete this policy"
							confirmInput={true}
							items={[policy.name]}
							disabled={this.state.disabled}
							onConfirm={this.onDelete}
						/>
					</div>
					<PageInput
						label="Name"
						help="Name of policy"
						type="text"
						placeholder="Enter name"
						value={policy.name}
						onChange={(val): void => {
							this.set('name', val);
						}}
					/>
					<PageTextArea
						label="Comment"
						help="Policy comment."
						placeholder="Policy comment"
						rows={3}
						value={policy.comment}
						onChange={(val: string): void => {
							this.set('comment', val);
						}}
					/>
					<label className="bp5-label">
						Roles
						<Help
							title="Roles"
							content="Roles associated with this policy. All requests from users with associated roles must pass this policy check."
						/>
						<div>
							{roles}
						</div>
					</label>
					<PageInputButton
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
					<PageSwitch
						label="Admin two-factor authentication"
						help="Require admins to use two-factor authentication."
						checked={adminProvider}
						onToggle={(): void => {
							if (adminProvider) {
								this.set('admin_secondary', null);
							} else {
								if (this.props.providers.length === 0) {
									Alert.warning(
										'No two-factor authentication providers exist');
									return;
								}
								this.set('admin_secondary', this.props.providers[0].id);
							}
						}}
					/>
					<PageSelect
						disabled={this.state.disabled}
						label="Admin Two-Factor Provider"
						help="Two-factor authentication provider that will be used. For users matching multiple policies the first provider will be used."
						hidden={!adminProvider}
						value={policy.admin_secondary}
						onChange={(val): void => {
							this.set('admin_secondary', val);
						}}
					>
						{adminProviders}
					</PageSelect>
					<PageSwitch
						label="User two-factor authentication"
						help="Require users to use two-factor authentication."
						checked={userProvider}
						onToggle={(): void => {
							if (userProvider) {
								this.set('user_secondary', null);
							} else {
								if (this.props.providers.length === 0) {
									Alert.warning(
										'No two-factor authentication providers exist');
									return;
								}
								this.set('user_secondary', this.props.providers[0].id);
							}
						}}
					/>
					<PageSelect
						disabled={this.state.disabled}
						label="User Two-Factor Provider"
						help="Two-factor authentication provider that will be used. For users matching multiple policies the first provider will be used."
						hidden={!userProvider}
						value={policy.user_secondary}
						onChange={(val): void => {
							this.set('user_secondary', val);
						}}
					>
						{userProviders}
					</PageSelect>
					<PolicyRule
						rule={whitelistNetworks}
						onChange={(val): void => {
							this.setRule('whitelist_networks', val);
						}}
					/>
					<PolicyRule
						rule={blacklistNetworks}
						onChange={(val): void => {
							this.setRule('blacklist_networks', val);
						}}
					/>
				</div>
				<div style={css.group}>
					<PageInfo
						fields={[
							{
								label: 'ID',
								value: this.props.policy.id || 'None',
							},
						]}
					/>
					<PageSwitch
						label="Enabled"
						help="Enable or disable policy."
						checked={!policy.disabled}
						onToggle={(): void => {
							this.set('disabled', !policy.disabled)
						}}
					/>
					<PolicyRule
						rule={location}
						onChange={(val): void => {
							this.setRule('location', val);
						}}
					/>
					<PolicyRule
						rule={operatingSystem}
						onChange={(val): void => {
							this.setRule('operating_system', val);
						}}
					/>
					<PolicyRule
						rule={browser}
						onChange={(val): void => {
							this.setRule('browser', val);
						}}
					/>
					<PageSwitch
						label="Admin WebAuthn device authentication"
						help="Require admins to use WebAuthn device authentication."
						checked={policy.admin_device_secondary}
						onToggle={(): void => {
							this.set('admin_device_secondary',
								!policy.admin_device_secondary)
						}}
					/>
					<PageSwitch
						label="User WebAuthn device authentication"
						help="Require users to use WebAuthn device authentication."
						checked={policy.user_device_secondary}
						onToggle={(): void => {
							this.set('user_device_secondary',
								!policy.user_device_secondary)
						}}
					/>
				</div>
			</div>
			<PageSave
				style={css.save}
				hidden={!this.state.policy}
				message={this.state.message}
				changed={this.state.changed}
				disabled={this.state.disabled}
				light={true}
				onCancel={(): void => {
					this.setState({
						...this.state,
						changed: false,
						policy: null,
					});
				}}
				onSave={this.onSave}
			/>
		</td>;
	}
}

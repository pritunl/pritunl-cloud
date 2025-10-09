/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Constants from '../Constants';
import * as FirewallTypes from '../types/FirewallTypes';
import * as FirewallActions from '../actions/FirewallActions';
import * as OrganizationTypes from "../types/OrganizationTypes";
import FirewallRule from './FirewallRule';
import PageInput from './PageInput';
import PageSelect from './PageSelect';
import PageInfo from './PageInfo';
import PageInputButton from './PageInputButton';
import PageSave from './PageSave';
import ConfirmButton from './ConfirmButton';
import Overview from './Overview';
import Help from './Help';
import PageTextArea from "./PageTextArea";

interface Props {
	organizations: OrganizationTypes.OrganizationsRo;
	firewall: FirewallTypes.FirewallRo;
	selected: boolean;
	onSelect: (shift: boolean) => void;
	onClose: () => void;
}

interface State {
	disabled: boolean;
	changed: boolean;
	message: string;
	addNetworkRole: string;
	firewall: FirewallTypes.Firewall;
	ingress: FirewallTypes.Rule;
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

export default class FirewallDetailed extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			disabled: false,
			changed: false,
			message: '',
			addNetworkRole: null,
			firewall: null,
			ingress: null,
		};
	}

	set(name: string, val: any): void {
		let firewall: any;

		if (this.state.changed) {
			firewall = {
				...this.state.firewall,
			};
		} else {
			firewall = {
				...this.props.firewall,
			};
		}

		firewall[name] = val;

		this.setState({
			...this.state,
			changed: true,
			firewall: firewall,
		});
	}

	onAddNetworkRole = (): void => {
		let firewall: FirewallTypes.Firewall;

		if (!this.state.addNetworkRole) {
			return;
		}

		if (this.state.changed) {
			firewall = {
				...this.state.firewall,
			};
		} else {
			firewall = {
				...this.props.firewall,
			};
		}

		let roles = [
			...(firewall.roles || []),
		];


		if (roles.indexOf(this.state.addNetworkRole) === -1) {
			roles.push(this.state.addNetworkRole);
		}

		roles.sort();
		firewall.roles = roles;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addNetworkRole: '',
			firewall: firewall,
		});
	}

	onRemoveNetworkRole = (networkRole: string): void => {
		let firewall: FirewallTypes.Firewall;

		if (this.state.changed) {
			firewall = {
				...this.state.firewall,
			};
		} else {
			firewall = {
				...this.props.firewall,
			};
		}

		let roles = [
			...(firewall.roles || []),
		];

		let i = roles.indexOf(networkRole);
		if (i === -1) {
			return;
		}

		roles.splice(i, 1);
		firewall.roles = roles;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			addNetworkRole: '',
			firewall: firewall,
		});
	}

	onAddIngress = (i: number): void => {
		let firewall: FirewallTypes.Firewall;

		if (this.state.changed) {
			firewall = {
				...this.state.firewall,
			};
		} else {
			firewall = {
				...this.props.firewall,
			};
		}

		let ingress = [
			...firewall.ingress,
		];

		ingress.splice(i + 1, 0, {
			protocol: 'all',
			source_ips: [
				'0.0.0.0/0',
				'::/0',
			],
		} as FirewallTypes.Rule);
		firewall.ingress = ingress;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			firewall: firewall,
		});
	}

	onChangeIngress(i: number, rule: FirewallTypes.Rule): void {
		let firewall: FirewallTypes.Firewall;

		if (this.state.changed) {
			firewall = {
				...this.state.firewall,
			};
		} else {
			firewall = {
				...this.props.firewall,
			};
		}

		let ingress = [
			...firewall.ingress,
		];

		ingress[i] = rule;

		firewall.ingress = ingress;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			firewall: firewall,
		});
	}

	onRemoveIngress(i: number): void {
		let firewall: FirewallTypes.Firewall;

		if (this.state.changed) {
			firewall = {
				...this.state.firewall,
			};
		} else {
			firewall = {
				...this.props.firewall,
			};
		}

		let ingress = [
			...firewall.ingress,
		];

		ingress.splice(i, 1);

		if (!ingress.length) {
			ingress = [
				{
					protocol: 'all',
					source_ips: [
						'0.0.0.0/0',
						'::/0',
					],
				} as FirewallTypes.Rule,
			];
		}

		firewall.ingress = ingress;

		this.setState({
			...this.state,
			changed: true,
			message: '',
			firewall: firewall,
		});
	}

	onSave = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		FirewallActions.commit(this.state.firewall).then((): void => {
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
						firewall: null,
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
		FirewallActions.remove(this.props.firewall.id).then((): void => {
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
		let firewall: FirewallTypes.Firewall = this.state.firewall ||
			this.props.firewall;

		let organizationsSelect: JSX.Element[] = [];
		organizationsSelect.push(
			<option key="null" value="">Node Firewall</option>);
		(this.props.organizations || []).forEach((org, index) => {
			organizationsSelect.push(
				<option
					key={org.id}
					value={org.id}
				>{org.name}</option>,
			);
		});

		let roles: JSX.Element[] = [];
		(firewall.roles || []).forEach((role, index) => {
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
							this.onRemoveNetworkRole(role);
						}}
					/>
				</div>,
			);
		})

		let rules: JSX.Element[] = [];
		(firewall.ingress || []).forEach((rule, index) => {
			rules.push(
				<FirewallRule
					key={index}
					rule={firewall.ingress[index]}
					onChange={(state: FirewallTypes.Rule): void => {
						this.onChangeIngress(index, state);
					}}
					onAdd={(): void => {
						this.onAddIngress(index);
					}}
					onRemove={(): void => {
						this.onRemoveIngress(index);
					}}
				/>,
			);
		})

		return <td
			className="bp5-cell"
			colSpan={3}
			style={css.card}
		>
			<div className="layout horizontal wrap">
				<div style={css.group}>
					<div
						className="layout horizontal tab-close"
						style={css.buttons}
						onClick={(evt): void => {
							let target = evt.target as HTMLElement;

							if (target.className && target.className.indexOf &&
								target.className.indexOf('tab-close') !== -1) {

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
						<Overview resource={"TODO"}/>
						<ConfirmButton
							className="bp5-minimal bp5-intent-danger bp5-icon-trash"
							style={css.button}
							safe={true}
							progressClassName="bp5-intent-danger"
							dialogClassName="bp5-intent-danger bp5-icon-delete"
							dialogLabel="Delete Firewall"
							confirmMsg="Permanently delete this firewall"
							confirmInput={true}
							items={[firewall.name]}
							disabled={this.state.disabled}
							onConfirm={this.onDelete}
						/>
					</div>
					<PageInput
						label="Name"
						help="Name of firewall"
						type="text"
						placeholder="Enter name"
						value={firewall.name}
						onChange={(val): void => {
							this.set('name', val);
						}}
					/>
					<PageTextArea
						label="Comment"
						help="Firewall comment."
						placeholder="Firewall comment"
						rows={3}
						value={firewall.comment}
						onChange={(val: string): void => {
							this.set('comment', val);
						}}
					/>
					<label style={css.itemsLabel}>
						Ingress Rules
						<Help
							title="Ingress Rules"
							content="Firewall rules."
						/>
					</label>
					<div style={css.rules}>
						{rules}
					</div>
				</div>
				<div style={css.group}>
					<PageInfo
						fields={[
							{
								label: 'ID',
								value: this.props.firewall.id || 'Unknown',
							},
						]}
					/>
					<PageSelect
						disabled={this.state.disabled}
						hidden={Constants.user}
						label="Organization"
						help="Organization for firewall, both the organaization and role must match. Select node firewall to match node network roles."
						value={firewall.organization}
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
							content="Network roles that will be matched with firewall rules. Network roles are case-sensitive."
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
			<PageSave
				style={css.save}
				hidden={!this.state.firewall && !this.state.message}
				message={this.state.message}
				changed={this.state.changed}
				disabled={this.state.disabled}
				light={true}
				onCancel={(): void => {
					this.setState({
						...this.state,
						changed: false,
						firewall: null,
					});
				}}
				onSave={this.onSave}
			/>
		</td>;
	}
}

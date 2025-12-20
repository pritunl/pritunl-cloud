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
import PageCreate from './PageCreate';
import ConfirmButton from './ConfirmButton';
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
	message: string;
	addNetworkRole: string;
	firewall: FirewallTypes.Firewall;
	ingress: FirewallTypes.Rule;
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

export default class FirewallNew extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			closed: false,
			disabled: false,
			changed: false,
			message: '',
			addNetworkRole: null,
			ingress: null,
			firewall: {
				name: "new-firewall",
				ingress: [
					{
						source_ips: [
							'0.0.0.0/0',
							'::/0',
						],
						protocol: 'icmp',
					} as FirewallTypes.Rule,
					{
						source_ips: [
							'0.0.0.0/0',
							'::/0',
						],
						protocol: 'tcp',
						port: '22',
					} as FirewallTypes.Rule,
				],
				comment: '22/tcp - SSH connections',
			},
		};
	}

	set(name: string, val: any): void {
		let firewall: any = {
			...this.state.firewall,
		};

		firewall[name] = val;

		this.setState({
			...this.state,
			changed: true,
			firewall: firewall,
		});
	}

	onAddNetworkRole = (): void => {
		let firewall: FirewallTypes.Firewall;

		firewall = {
			...this.state.firewall,
		};

		if (!this.state.addNetworkRole) {
			return;
		}

		let networkRoles = [
			...(firewall.roles || []),
		];


		if (networkRoles.indexOf(this.state.addNetworkRole) === -1) {
			networkRoles.push(this.state.addNetworkRole);
		}

		networkRoles.sort();
		firewall.roles = networkRoles;

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

		firewall = {
			...this.state.firewall,
		};

		let networkRoles = [
			...(firewall.roles || []),
		];

		let i = networkRoles.indexOf(networkRole);
		if (i === -1) {
			return;
		}

		networkRoles.splice(i, 1);
		firewall.roles = networkRoles;

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

		firewall = {
			...this.state.firewall,
		};

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

		firewall = {
			...this.state.firewall,
		};

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

		firewall = {
			...this.state.firewall,
		};

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

	onCreate = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});

		let firewall: any = {
			...this.state.firewall,
		};

		FirewallActions.create(firewall).then((): void => {
			this.setState({
				...this.state,
				message: 'Firewall created successfully',
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
		let firewall: FirewallTypes.Firewall = this.state.firewall;

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

		let networkRoles: JSX.Element[] = [];
		(firewall.roles || []).forEach((role, index) => {
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
							Roles
							<Help
								title="Roles"
								content="Roles that will be matched with firewall rules. Roles are case-sensitive."
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
					hidden={!this.state.firewall}
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

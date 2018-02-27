/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as FirewallTypes from '../types/FirewallTypes';
import * as OrganizationTypes from '../types/OrganizationTypes';
import FirewallsStore from '../stores/FirewallsStore';
import OrganizationsStore from '../stores/OrganizationsStore';
import * as FirewallActions from '../actions/FirewallActions';
import * as OrganizationActions from '../actions/OrganizationActions';
import Firewall from './Firewall';
import FirewallsPage from './FirewallsPage';
import Page from './Page';
import PageHeader from './PageHeader';
import NonState from './NonState';
import ConfirmButton from './ConfirmButton';

interface Selected {
	[key: string]: boolean;
}

interface Opened {
	[key: string]: boolean;
}

interface State {
	firewalls: FirewallTypes.FirewallsRo;
	organizations: OrganizationTypes.OrganizationsRo;
	selected: Selected;
	opened: Opened;
	newOpened: boolean;
	lastSelected: string;
	disabled: boolean;
}

const css = {
	items: {
		width: '100%',
		marginTop: '-5px',
		display: 'table',
		borderSpacing: '0 5px',
	} as React.CSSProperties,
	itemsBox: {
		width: '100%',
		overflowY: 'auto',
	} as React.CSSProperties,
	placeholder: {
		opacity: 0,
		width: '100%',
	} as React.CSSProperties,
	header: {
		marginTop: '-19px',
	} as React.CSSProperties,
	heading: {
		margin: '19px 0 0 0',
	} as React.CSSProperties,
	button: {
		margin: '10px 0 0 10px',
	} as React.CSSProperties,
};

export default class Firewalls extends React.Component<{}, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			firewalls: FirewallsStore.firewalls,
			organizations: OrganizationsStore.organizations,
			selected: {},
			opened: {},
			newOpened: false,
			lastSelected: null,
			disabled: false,
		};
	}

	get selected(): boolean {
		return !!Object.keys(this.state.selected).length;
	}

	get opened(): boolean {
		return !!Object.keys(this.state.opened).length;
	}

	componentDidMount(): void {
		FirewallsStore.addChangeListener(this.onChange);
		OrganizationsStore.addChangeListener(this.onChange);
		FirewallActions.sync();
		OrganizationActions.sync();
	}

	componentWillUnmount(): void {
		FirewallsStore.removeChangeListener(this.onChange);
		OrganizationsStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		let firewalls = FirewallsStore.firewalls;
		let selected: Selected = {};
		let curSelected = this.state.selected;
		let opened: Opened = {};
		let curOpened = this.state.opened;

		firewalls.forEach((firewall: FirewallTypes.Firewall): void => {
			if (curSelected[firewall.id]) {
				selected[firewall.id] = true;
			}
			if (curOpened[firewall.id]) {
				opened[firewall.id] = true;
			}
		});

		this.setState({
			...this.state,
			firewalls: firewalls,
			organizations: OrganizationsStore.organizations,
			selected: selected,
			opened: opened,
		});
	}

	onDelete = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		FirewallActions.removeMulti(
				Object.keys(this.state.selected)).then((): void => {
			this.setState({
				...this.state,
				selected: {},
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
		let firewallsDom: JSX.Element[] = [];

		this.state.firewalls.forEach((
				firewall: FirewallTypes.FirewallRo): void => {
			firewallsDom.push(<Firewall
				key={firewall.id}
				firewall={firewall}
				organizations={this.state.organizations}
				selected={!!this.state.selected[firewall.id]}
				open={!!this.state.opened[firewall.id]}
				onSelect={(shift: boolean): void => {
					let selected = {
						...this.state.selected,
					};

					if (shift) {
						let firewalls = this.state.firewalls;
						let start: number;
						let end: number;

						for (let i = 0; i < firewalls.length; i++) {
							let usr = firewalls[i];

							if (usr.id === firewall.id) {
								start = i;
							} else if (usr.id === this.state.lastSelected) {
								end = i;
							}
						}

						if (start !== undefined && end !== undefined) {
							if (start > end) {
								end = [start, start = end][0];
							}

							for (let i = start; i <= end; i++) {
								selected[firewalls[i].id] = true;
							}

							this.setState({
								...this.state,
								lastSelected: firewall.id,
								selected: selected,
							});

							return;
						}
					}

					if (selected[firewall.id]) {
						delete selected[firewall.id];
					} else {
						selected[firewall.id] = true;
					}

					this.setState({
						...this.state,
						lastSelected: firewall.id,
						selected: selected,
					});
				}}
				onOpen={(): void => {
					let opened = {
						...this.state.opened,
					};

					if (opened[firewall.id]) {
						delete opened[firewall.id];
					} else {
						opened[firewall.id] = true;
					}

					this.setState({
						...this.state,
						opened: opened,
					});
				}}
			/>);
		});

		return <Page>
			<PageHeader>
				<div className="layout horizontal wrap" style={css.header}>
					<h2 style={css.heading}>Firewalls</h2>
					<div className="flex"/>
					<div>
						<button
							className="pt-button pt-intent-warning pt-icon-chevron-up"
							style={css.button}
							disabled={!this.opened}
							type="button"
							onClick={(): void => {
								this.setState({
									...this.state,
									opened: {},
								});
							}}
						>
							Collapse All
						</button>
						<ConfirmButton
							label="Delete Selected"
							className="pt-intent-danger pt-icon-delete"
							progressClassName="pt-intent-danger"
							style={css.button}
							disabled={!this.selected || this.state.disabled}
							onConfirm={this.onDelete}
						/>
						<button
							className="pt-button pt-intent-success pt-icon-add"
							style={css.button}
							disabled={this.state.disabled}
							type="button"
							onClick={(): void => {
								this.setState({
									...this.state,
									disabled: true,
								});
								FirewallActions.create({
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
								} as FirewallTypes.Firewall).then((): void => {
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
							}}
						>New</button>
					</div>
				</div>
			</PageHeader>
			<div style={css.itemsBox}>
				<div style={css.items}>
					{firewallsDom}
					<tr className="pt-card pt-row" style={css.placeholder}>
						<td colSpan={5} style={css.placeholder}/>
					</tr>
				</div>
			</div>
			<NonState
				hidden={!!firewallsDom.length}
				iconClass="pt-icon-key"
				title="No firewalls"
				description="Add a new firewall to get started."
			/>
			<FirewallsPage
				onPage={(): void => {
					this.setState({
						lastSelected: null,
					});
				}}
			/>
		</Page>;
	}
}

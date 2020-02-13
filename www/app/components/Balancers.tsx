/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as BalancerTypes from '../types/BalancerTypes';
import * as CertificateTypes from '../types/CertificateTypes';
import * as OrganizationTypes from '../types/OrganizationTypes';
import * as DatacenterTypes from '../types/DatacenterTypes';
import BalancersStore from '../stores/BalancersStore';
import OrganizationsStore from '../stores/OrganizationsStore';
import DatacentersStore from '../stores/DatacentersStore';
import * as BalancerActions from '../actions/BalancerActions';
import * as OrganizationActions from '../actions/OrganizationActions';
import * as DatacenterActions from '../actions/DatacenterActions';
import Balancer from './Balancer';
import BalancersPage from './BalancersPage';
import BalancersFilter from './BalancersFilter';
import Page from './Page';
import PageHeader from './PageHeader';
import NonState from './NonState';
import ConfirmButton from './ConfirmButton';
import CertificatesStore from "../stores/CertificatesStore";
import * as CertificateActions from "../actions/CertificateActions";

interface Selected {
	[key: string]: boolean;
}

interface Opened {
	[key: string]: boolean;
}

interface State {
	balancers: BalancerTypes.BalancersRo;
	filter: BalancerTypes.Filter;
	organizations: OrganizationTypes.OrganizationsRo;
	certificates: CertificateTypes.CertificatesRo;
	datacenters: DatacenterTypes.DatacentersRo;
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
		margin: '8px 0 0 8px',
	} as React.CSSProperties,
	buttons: {
		marginTop: '8px',
	} as React.CSSProperties,
};

export default class Balancers extends React.Component<{}, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			balancers: BalancersStore.balancers,
			filter: BalancersStore.filter,
			organizations: OrganizationsStore.organizations,
			certificates: CertificatesStore.certificates,
			datacenters: DatacentersStore.datacenters,
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
		BalancersStore.addChangeListener(this.onChange);
		OrganizationsStore.addChangeListener(this.onChange);
		CertificatesStore.addChangeListener(this.onChange);
		DatacentersStore.addChangeListener(this.onChange);
		BalancerActions.sync();
		OrganizationActions.sync();
		CertificateActions.sync();
		DatacenterActions.sync();
	}

	componentWillUnmount(): void {
		BalancersStore.removeChangeListener(this.onChange);
		OrganizationsStore.removeChangeListener(this.onChange);
		CertificatesStore.removeChangeListener(this.onChange);
		DatacentersStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		let balancers = BalancersStore.balancers;
		let selected: Selected = {};
		let curSelected = this.state.selected;
		let opened: Opened = {};
		let curOpened = this.state.opened;

		balancers.forEach((balancer: BalancerTypes.Balancer): void => {
			if (curSelected[balancer.id]) {
				selected[balancer.id] = true;
			}
			if (curOpened[balancer.id]) {
				opened[balancer.id] = true;
			}
		});

		this.setState({
			...this.state,
			balancers: balancers,
			filter: BalancersStore.filter,
			organizations: OrganizationsStore.organizations,
			certificates: CertificatesStore.certificates,
			datacenters: DatacentersStore.datacenters,
			selected: selected,
			opened: opened,
		});
	}

	onDelete = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		BalancerActions.removeMulti(
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
		let balancersDom: JSX.Element[] = [];

		this.state.balancers.forEach((
				balancer: BalancerTypes.BalancerRo): void => {
			balancersDom.push(<Balancer
				key={balancer.id}
				balancer={balancer}
				organizations={this.state.organizations}
				certificates={this.state.certificates}
				datacenters={this.state.datacenters}
				selected={!!this.state.selected[balancer.id]}
				open={!!this.state.opened[balancer.id]}
				onSelect={(shift: boolean): void => {
					let selected = {
						...this.state.selected,
					};

					if (shift) {
						let balancers = this.state.balancers;
						let start: number;
						let end: number;

						for (let i = 0; i < balancers.length; i++) {
							let usr = balancers[i];

							if (usr.id === balancer.id) {
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
								selected[balancers[i].id] = true;
							}

							this.setState({
								...this.state,
								lastSelected: balancer.id,
								selected: selected,
							});

							return;
						}
					}

					if (selected[balancer.id]) {
						delete selected[balancer.id];
					} else {
						selected[balancer.id] = true;
					}

					this.setState({
						...this.state,
						lastSelected: balancer.id,
						selected: selected,
					});
				}}
				onOpen={(): void => {
					let opened = {
						...this.state.opened,
					};

					if (opened[balancer.id]) {
						delete opened[balancer.id];
					} else {
						opened[balancer.id] = true;
					}

					this.setState({
						...this.state,
						opened: opened,
					});
				}}
			/>);
		});

		let filterClass = 'bp3-button bp3-intent-primary bp3-icon-filter ';
		if (this.state.filter) {
			filterClass += 'bp3-active';
		}

		return <Page>
			<PageHeader>
				<div className="layout horizontal wrap" style={css.header}>
					<h2 style={css.heading}>Balancers</h2>
					<div className="flex"/>
					<div style={css.buttons}>
						<button
							className={filterClass}
							style={css.button}
							type="button"
							onClick={(): void => {
								if (this.state.filter === null) {
									BalancerActions.filter({});
								} else {
									BalancerActions.filter(null);
								}
							}}
						>
							Filters
						</button>
						<button
							className="bp3-button bp3-intent-warning bp3-icon-chevron-up"
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
							className="bp3-intent-danger bp3-icon-delete"
							progressClassName="bp3-intent-danger"
							style={css.button}
							disabled={!this.selected || this.state.disabled}
							onConfirm={this.onDelete}
						/>
						<button
							className="bp3-button bp3-intent-success bp3-icon-add"
							style={css.button}
							disabled={this.state.disabled}
							type="button"
							onClick={(): void => {
								this.setState({
									...this.state,
									disabled: true,
								});
								BalancerActions.create({
								} as BalancerTypes.Balancer).then((): void => {
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
			<BalancersFilter
				filter={this.state.filter}
				onFilter={(filter): void => {
					BalancerActions.filter(filter);
				}}
				organizations={this.state.organizations}
			/>
			<div style={css.itemsBox}>
				<div style={css.items}>
					{balancersDom}
					<tr className="bp3-card bp3-row" style={css.placeholder}>
						<td colSpan={5} style={css.placeholder}/>
					</tr>
				</div>
			</div>
			<NonState
				hidden={!!balancersDom.length}
				iconClass="bp3-icon-random"
				title="No balancers"
				description="Add a new balancer to get started."
			/>
			<BalancersPage
				onPage={(): void => {
					this.setState({
						...this.state,
						selected: {},
						lastSelected: null,
					});
				}}
			/>
		</Page>;
	}
}

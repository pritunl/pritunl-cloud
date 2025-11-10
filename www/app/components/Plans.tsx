/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as PlanTypes from '../types/PlanTypes';
import * as OrganizationTypes from '../types/OrganizationTypes';
import PlansStore from '../stores/PlansStore';
import CompletionStore from '../stores/CompletionStore';
import * as PlanActions from '../actions/PlanActions';
import * as CompletionActions from '../actions/CompletionActions';
import Plan from './Plan';
import PlansFilter from './PlansFilter';
import PlansPage from './PlansPage';
import PlanNew from './PlanNew';
import Page from './Page';
import PageHeader from './PageHeader';
import NonState from './NonState';
import ConfirmButton from './ConfirmButton';
import * as SecretTypes from "../types/SecretTypes";

interface Selected {
	[key: string]: boolean;
}

interface Opened {
	[key: string]: boolean;
}

interface State {
	plans: PlanTypes.PlansRo;
	filter: PlanTypes.Filter;
	organizations: OrganizationTypes.OrganizationsRo;
	secrets: SecretTypes.SecretsRo;
	organization: string;
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
		tableLayout: 'fixed',
		borderSpacing: '0 5px',
	} as React.CSSProperties,
	itemsBox: {
		width: '100%',
		overflowY: 'auto',
	} as React.CSSProperties,
	group: {
		width: '100%',
	} as React.CSSProperties,
	groupBox: {
		margin: '16px 0 0 8px',
		width: '100%',
		maxWidth: '200px',
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
	selectFirst: {
		width: '100%',
		borderTopLeftRadius: '3px',
		borderBottomLeftRadius: '3px',
	} as React.CSSProperties,
	selectInner: {
		width: '100%',
	} as React.CSSProperties,
	selectBox: {
		flex: '1',
	} as React.CSSProperties,
	button: {
		margin: '8px 0 0 8px',
	} as React.CSSProperties,
	buttons: {
		marginTop: '8px',
	} as React.CSSProperties,
};

export default class Plans extends React.Component<{}, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			plans: PlansStore.plans,
			filter: PlansStore.filter,
			organizations: CompletionStore.organizations,
			secrets: CompletionStore.secrets,
			organization: '',
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
		PlansStore.addChangeListener(this.onChange);
		CompletionStore.addChangeListener(this.onChange);
		PlanActions.sync();
		CompletionActions.sync();
	}

	componentWillUnmount(): void {
		PlansStore.removeChangeListener(this.onChange);
		CompletionStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		let plans = PlansStore.plans;
		let selected: Selected = {};
		let curSelected = this.state.selected;
		let opened: Opened = {};
		let curOpened = this.state.opened;

		plans.forEach((plan: PlanTypes.Plan): void => {
			if (curSelected[plan.id]) {
				selected[plan.id] = true;
			}
			if (curOpened[plan.id]) {
				opened[plan.id] = true;
			}
		});

		this.setState({
			...this.state,
			plans: plans,
			filter: PlansStore.filter,
			organizations: CompletionStore.organizations,
			secrets: CompletionStore.secrets,
			selected: selected,
			opened: opened,
		});
	}

	onDelete = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		PlanActions.removeMulti(
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
		let plansDom: JSX.Element[] = [];

		let hasOrganizations = false;
		let organizationsSelect: JSX.Element[] = [];
		if (this.state.organizations.length) {
			hasOrganizations = true;
			for (let organization of this.state.organizations) {
				organizationsSelect.push(
					<option
						key={organization.id}
						value={organization.id}
					>{organization.name}</option>,
				);
			}
		} else {
			organizationsSelect.push(
				<option
					key="null"
					value=""
				>No Organizations</option>,
			);
		}

		this.state.plans.forEach((
				plan: PlanTypes.PlanRo): void => {
			plansDom.push(<Plan
				key={plan.id}
				plan={plan}
				organizations={this.state.organizations}
				secrets={this.state.secrets}
				selected={!!this.state.selected[plan.id]}
				open={!!this.state.opened[plan.id]}
				onSelect={(shift: boolean): void => {
					let selected = {
						...this.state.selected,
					};

					if (shift) {
						let plans = this.state.plans;
						let start: number;
						let end: number;

						for (let i = 0; i < plans.length; i++) {
							let usr = plans[i];

							if (usr.id === plan.id) {
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
								selected[plans[i].id] = true;
							}

							this.setState({
								...this.state,
								lastSelected: plan.id,
								selected: selected,
							});

							return;
						}
					}

					if (selected[plan.id]) {
						delete selected[plan.id];
					} else {
						selected[plan.id] = true;
					}

					this.setState({
						...this.state,
						lastSelected: plan.id,
						selected: selected,
					});
				}}
				onOpen={(): void => {
					let opened = {
						...this.state.opened,
					};

					if (opened[plan.id]) {
						delete opened[plan.id];
					} else {
						opened[plan.id] = true;
					}

					this.setState({
						...this.state,
						opened: opened,
					});
				}}
			/>);
		});

		let filterClass = 'bp5-button bp5-intent-primary bp5-icon-filter ';
		if (this.state.filter) {
			filterClass += 'bp5-active';
		}

		let selectedNames: string[] = [];
		for (let planId of Object.keys(this.state.selected)) {
			let plan = PlansStore.plan(planId);
			if (plan) {
				selectedNames.push(plan.name || planId);
			} else {
				selectedNames.push(planId);
			}
		}

		let newDiskDom: JSX.Element;
		if (this.state.newOpened) {
			newDiskDom = <PlanNew
				organizations={this.state.organizations}
				secrets={this.state.secrets}
				onClose={(): void => {
					this.setState({
						...this.state,
						newOpened: false,
					});
				}}
			/>;
		}

		return <Page>
			<PageHeader>
				<div className="layout horizontal wrap" style={css.header}>
					<h2 style={css.heading}>Plans</h2>
					<div className="flex"/>
					<div style={css.buttons}>
						<button
							className={filterClass}
							style={css.button}
							type="button"
							onClick={(): void => {
								if (this.state.filter === null) {
									PlanActions.filter({});
								} else {
									PlanActions.filter(null);
								}
							}}
						>
							Filters
						</button>
						<button
							className="bp5-button bp5-intent-warning bp5-icon-chevron-up"
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
							className="bp5-intent-danger bp5-icon-delete"
							progressClassName="bp5-intent-danger"
							safe={true}
							style={css.button}
							confirmMsg="Permanently delete the selected plans"
							confirmInput={true}
							items={selectedNames}
							disabled={!this.selected || this.state.disabled}
							onConfirm={this.onDelete}
						/>
						<button
							className="bp5-button bp5-intent-success bp5-icon-add"
							style={css.button}
							disabled={this.state.disabled || this.state.newOpened}
							type="button"
							onClick={(): void => {
								this.setState({
									...this.state,
									newOpened: true,
								});
							}}
						>New</button>
					</div>
				</div>
			</PageHeader>
			<PlansFilter
				filter={this.state.filter}
				onFilter={(filter): void => {
					PlanActions.filter(filter);
				}}
				organizations={this.state.organizations}
			/>
			<div style={css.itemsBox}>
				<div style={css.items}>
					{newDiskDom}
					{plansDom}
					<tr className="bp5-card bp5-row" style={css.placeholder}>
						<td colSpan={5} style={css.placeholder}/>
					</tr>
				</div>
			</div>
			<NonState
				hidden={!!plansDom.length}
				iconClass="bp5-icon-map-marker"
				title="No plans"
				description="Add a new plan to get started."
			/>
			<PlansPage
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

/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Constants from "../Constants";
import * as AdvisoryTypes from '../types/AdvisoryTypes';
import * as OrganizationTypes from '../types/OrganizationTypes';
import AdvisoriesStore from '../stores/AdvisoriesStore';
import CompletionStore from '../stores/CompletionStore';
import * as AdvisoryActions from '../actions/AdvisoryActions';
import * as CompletionActions from '../actions/CompletionActions';
import Advisory from './Advisory';
import AdvisoriesFilter from './AdvisoriesFilter';
import AdvisoriesPage from './AdvisoriesPage';
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
	advisories: AdvisoryTypes.AdvisoriesRo;
	filter: AdvisoryTypes.Filter;
	organizations: OrganizationTypes.OrganizationsRo;
	selected: Selected;
	opened: Opened;
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

export default class Advisories extends React.Component<{}, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			advisories: AdvisoriesStore.advisories,
			filter: AdvisoriesStore.filter,
			organizations: CompletionStore.organizations,
			selected: {},
			opened: {},
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
		AdvisoriesStore.addChangeListener(this.onChange);
		CompletionStore.addChangeListener(this.onChange);
		AdvisoryActions.sync();
		CompletionActions.sync();
	}

	componentWillUnmount(): void {
		AdvisoriesStore.removeChangeListener(this.onChange);
		CompletionStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		let advisories = AdvisoriesStore.advisories;
		let selected: Selected = {};
		let curSelected = this.state.selected;
		let opened: Opened = {};
		let curOpened = this.state.opened;

		advisories.forEach((advisory: AdvisoryTypes.Advisory): void => {
			if (curSelected[advisory.id]) {
				selected[advisory.id] = true;
			}
			if (curOpened[advisory.id]) {
				opened[advisory.id] = true;
			}
		});

		this.setState({
			...this.state,
			advisories: advisories,
			filter: AdvisoriesStore.filter,
			organizations: CompletionStore.organizations,
			selected: selected,
			opened: opened,
		});
	}

	onDelete = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		AdvisoryActions.removeMulti(
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
		let advisoriesDom: JSX.Element[] = [];

		this.state.advisories.forEach((
			advisory: AdvisoryTypes.AdvisoryRo): void => {
			advisoriesDom.push(<Advisory
				key={advisory.id}
				advisory={advisory}
				organizations={this.state.organizations}
				selected={!!this.state.selected[advisory.id]}
				open={!!this.state.opened[advisory.id]}
				onSelect={(shift: boolean): void => {
					let selected = {
						...this.state.selected,
					};

					if (shift) {
						let advisories = this.state.advisories;
						let start: number;
						let end: number;

						for (let i = 0; i < advisories.length; i++) {
							let usr = advisories[i];

							if (usr.id === advisory.id) {
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
								selected[advisories[i].id] = true;
							}

							this.setState({
								...this.state,
								lastSelected: advisory.id,
								selected: selected,
							});

							return;
						}
					}

					if (selected[advisory.id]) {
						delete selected[advisory.id];
					} else {
						selected[advisory.id] = true;
					}

					this.setState({
						...this.state,
						lastSelected: advisory.id,
						selected: selected,
					});
				}}
				onOpen={(): void => {
					let opened = {
						...this.state.opened,
					};

					if (opened[advisory.id]) {
						delete opened[advisory.id];
					} else {
						opened[advisory.id] = true;
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
		for (let advId of Object.keys(this.state.selected)) {
			let adv = AdvisoriesStore.advisory(advId);
			if (adv) {
				selectedNames.push(adv.reference || advId);
			} else {
				selectedNames.push(advId);
			}
		}

		let sizeRow = <div style={{"display": "table-row"}}>
			<div style={{display: "table-cell", width: "auto"}}></div>
			<div style={{display: "table-cell", width: "auto"}}></div>
			<div style={{display: "table-cell", width: "auto"}}></div>
			<div style={{display: "table-cell", width: "auto"}}></div>
			<div style={{display: "table-cell", width: "auto"}}></div>
		</div>

		return <Page>
			<PageHeader>
				<div className="layout horizontal wrap" style={css.header}>
					<h2 style={css.heading}>Advisories</h2>
					<div className="flex"/>
					<div style={css.buttons}>
						<button
							className={filterClass}
							style={css.button}
							type="button"
							onClick={(): void => {
								if (this.state.filter === null) {
									AdvisoryActions.filter({});
								} else {
									AdvisoryActions.filter(null);
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
							confirmMsg="Permanently delete the selected advisories"
							confirmInput={true}
							items={selectedNames}
							disabled={!this.selected || this.state.disabled}
							onConfirm={this.onDelete}
						/>
					</div>
				</div>
			</PageHeader>
			<AdvisoriesFilter
				filter={this.state.filter}
				onFilter={(filter): void => {
					AdvisoryActions.filter(filter);
				}}
				organizations={this.state.organizations}
			/>
			<div style={css.itemsBox}>
				<div style={css.items}>
					{sizeRow}
					{advisoriesDom}
				</div>
			</div>
			<NonState
				hidden={!!advisoriesDom.length}
				iconClass="bp5-icon-warning-sign"
				title="No advisories"
				description="No security advisories affecting your instances."
			/>
			<AdvisoriesPage
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

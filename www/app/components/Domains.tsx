/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Constants from "../Constants";
import * as DomainTypes from '../types/DomainTypes';
import * as OrganizationTypes from '../types/OrganizationTypes';
import DomainsStore from '../stores/DomainsStore';
import CompletionStore from '../stores/CompletionStore';
import * as DomainActions from '../actions/DomainActions';
import * as CompletionActions from '../actions/CompletionActions';
import Domain from './Domain';
import DomainsFilter from './DomainsFilter';
import DomainsPage from './DomainsPage';
import DomainNew from './DomainNew';
import Page from './Page';
import PageHeader from './PageHeader';
import NonState from './NonState';
import ConfirmButton from './ConfirmButton';
import * as SecretTypes from "../types/SecretTypes";
import DiskNew from "./DiskNew";

interface Selected {
	[key: string]: boolean;
}

interface Opened {
	[key: string]: boolean;
}

interface State {
	domains: DomainTypes.DomainsRo;
	filter: DomainTypes.Filter;
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

export default class Domains extends React.Component<{}, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			domains: DomainsStore.domains,
			filter: DomainsStore.filter,
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
		DomainsStore.addChangeListener(this.onChange);
		CompletionStore.addChangeListener(this.onChange);
		DomainActions.sync();
		CompletionActions.sync();
	}

	componentWillUnmount(): void {
		DomainsStore.removeChangeListener(this.onChange);
		CompletionStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		let domains = DomainsStore.domains;
		let selected: Selected = {};
		let curSelected = this.state.selected;
		let opened: Opened = {};
		let curOpened = this.state.opened;

		domains.forEach((domain: DomainTypes.Domain): void => {
			if (curSelected[domain.id]) {
				selected[domain.id] = true;
			}
			if (curOpened[domain.id]) {
				opened[domain.id] = true;
			}
		});

		this.setState({
			...this.state,
			domains: domains,
			filter: DomainsStore.filter,
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
		DomainActions.removeMulti(
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
		let domainsDom: JSX.Element[] = [];

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

		this.state.domains.forEach((
				domain: DomainTypes.DomainRo): void => {
			domainsDom.push(<Domain
				key={domain.id}
				domain={domain}
				organizations={this.state.organizations}
				secrets={this.state.secrets}
				selected={!!this.state.selected[domain.id]}
				open={!!this.state.opened[domain.id]}
				onSelect={(shift: boolean): void => {
					let selected = {
						...this.state.selected,
					};

					if (shift) {
						let domains = this.state.domains;
						let start: number;
						let end: number;

						for (let i = 0; i < domains.length; i++) {
							let usr = domains[i];

							if (usr.id === domain.id) {
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
								selected[domains[i].id] = true;
							}

							this.setState({
								...this.state,
								lastSelected: domain.id,
								selected: selected,
							});

							return;
						}
					}

					if (selected[domain.id]) {
						delete selected[domain.id];
					} else {
						selected[domain.id] = true;
					}

					this.setState({
						...this.state,
						lastSelected: domain.id,
						selected: selected,
					});
				}}
				onOpen={(): void => {
					let opened = {
						...this.state.opened,
					};

					if (opened[domain.id]) {
						delete opened[domain.id];
					} else {
						opened[domain.id] = true;
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
		for (let domainId of Object.keys(this.state.selected)) {
			let domain = DomainsStore.domain(domainId);
			if (domain) {
				selectedNames.push(domain.name || domainId);
			} else {
				selectedNames.push(domainId);
			}
		}

		let newDiskDom: JSX.Element;
		if (this.state.newOpened) {
			newDiskDom = <DomainNew
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
					<h2 style={css.heading}>Domains</h2>
					<div className="flex"/>
					<div style={css.buttons}>
						<button
							className={filterClass}
							style={css.button}
							type="button"
							onClick={(): void => {
								if (this.state.filter === null) {
									DomainActions.filter({});
								} else {
									DomainActions.filter(null);
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
							confirmMsg="Permanently delete the selected domains"
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
			<DomainsFilter
				filter={this.state.filter}
				onFilter={(filter): void => {
					DomainActions.filter(filter);
				}}
				organizations={this.state.organizations}
			/>
			<div style={css.itemsBox}>
				<div style={css.items}>
					{newDiskDom}
					{domainsDom}
					<tr className="bp5-card bp5-row" style={css.placeholder}>
						<td colSpan={2} style={css.placeholder}/>
					</tr>
				</div>
			</div>
			<NonState
				hidden={!!domainsDom.length}
				iconClass="bp5-icon-map-marker"
				title="No domains"
				description="Add a new domain to get started."
			/>
			<DomainsPage
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

/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as OrganizationTypes from '../types/OrganizationTypes';
import OrganizationsStore from '../stores/OrganizationsStore';
import * as OrganizationActions from '../actions/OrganizationActions';
import Organization from './Organization';
import OrganizationNew from './OrganizationNew';
import OrganizationsFilter from './OrganizationsFilter';
import OrganizationsPage from './OrganizationsPage';
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
	organizations: OrganizationTypes.OrganizationsRo;
	filter: OrganizationTypes.Filter;
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

export default class Organizations extends React.Component<{}, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			organizations: OrganizationsStore.organizations,
			filter: OrganizationsStore.filter,
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
		OrganizationsStore.addChangeListener(this.onChange);
		OrganizationActions.sync();
	}

	componentWillUnmount(): void {
		OrganizationsStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		let organizations = OrganizationsStore.organizations;
		let selected: Selected = {};
		let curSelected = this.state.selected;
		let opened: Opened = {};
		let curOpened = this.state.opened;

		organizations.forEach((organization: OrganizationTypes.Organization): void => {
			if (curSelected[organization.id]) {
				selected[organization.id] = true;
			}
			if (curOpened[organization.id]) {
				opened[organization.id] = true;
			}
		});

		this.setState({
			...this.state,
			organizations: organizations,
			filter: OrganizationsStore.filter,
			selected: selected,
			opened: opened,
		});
	}

	onDelete = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		OrganizationActions.removeMulti(
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
		let organizationsDom: JSX.Element[] = [];

		this.state.organizations.forEach((
				organization: OrganizationTypes.OrganizationRo): void => {
			organizationsDom.push(<Organization
				key={organization.id}
				organization={organization}
				selected={!!this.state.selected[organization.id]}
				open={!!this.state.opened[organization.id]}
				onSelect={(shift: boolean): void => {
					let selected = {
						...this.state.selected,
					};

					if (shift) {
						let organizations = this.state.organizations;
						let start: number;
						let end: number;

						for (let i = 0; i < organizations.length; i++) {
							let usr = organizations[i];

							if (usr.id === organization.id) {
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
								selected[organizations[i].id] = true;
							}

							this.setState({
								...this.state,
								lastSelected: organization.id,
								selected: selected,
							});

							return;
						}
					}

					if (selected[organization.id]) {
						delete selected[organization.id];
					} else {
						selected[organization.id] = true;
					}

					this.setState({
						...this.state,
						lastSelected: organization.id,
						selected: selected,
					});
				}}
				onOpen={(): void => {
					let opened = {
						...this.state.opened,
					};

					if (opened[organization.id]) {
						delete opened[organization.id];
					} else {
						opened[organization.id] = true;
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
		for (let instId of Object.keys(this.state.selected)) {
			let inst = OrganizationsStore.organization(instId);
			if (inst) {
				selectedNames.push(inst.name || instId);
			} else {
				selectedNames.push(instId);
			}
		}

		let newOrganizationDom: JSX.Element;
		if (this.state.newOpened) {
			newOrganizationDom = <OrganizationNew
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
					<h2 style={css.heading}>Organizations</h2>
					<div className="flex"/>
					<div style={css.buttons}>
						<button
							className={filterClass}
							style={css.button}
							type="button"
							onClick={(): void => {
								if (this.state.filter === null) {
									OrganizationActions.filter({});
								} else {
									OrganizationActions.filter(null);
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
							confirmMsg="Permanently delete the selected organizations"
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
			<OrganizationsFilter
				filter={this.state.filter}
				onFilter={(filter): void => {
					OrganizationActions.filter(filter);
				}}
			/>
			<div style={css.itemsBox}>
				<div style={css.items}>
					{newOrganizationDom}
					{organizationsDom}
					<tr className="bp5-card bp5-row" style={css.placeholder}>
						<td colSpan={2} style={css.placeholder}/>
					</tr>
				</div>
			</div>
			<NonState
				hidden={!!organizationsDom.length}
				iconClass="bp5-icon-people"
				title="No organizations"
				description="Add a new organization to get started."
			/>
			<OrganizationsPage
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

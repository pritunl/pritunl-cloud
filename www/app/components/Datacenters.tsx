/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as DatacenterTypes from '../types/DatacenterTypes';
import * as OrganizationTypes from "../types/OrganizationTypes";
import * as StorageTypes from '../types/StorageTypes';
import DatacentersStore from '../stores/DatacentersStore';
import StoragesStore from '../stores/StoragesStore';
import OrganizationsStore from "../stores/OrganizationsStore";
import * as DatacenterActions from '../actions/DatacenterActions';
import * as StorageActions from '../actions/StorageActions';
import * as OrganizationActions from '../actions/OrganizationActions';
import Datacenter from './Datacenter';
import DatacenterNew from './DatacenterNew';
import DatacentersFilter from './DatacentersFilter';
import DatacentersPage from './DatacentersPage';
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
	datacenters: DatacenterTypes.DatacentersRo;
	organizations: OrganizationTypes.OrganizationsRo;
	storages: StorageTypes.StoragesRo;
	filter: DatacenterTypes.Filter;
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

export default class Datacenters extends React.Component<{}, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			datacenters: DatacentersStore.datacenters,
			storages: StoragesStore.storages,
			organizations: OrganizationsStore.organizations,
			filter: DatacentersStore.filter,
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
		DatacentersStore.addChangeListener(this.onChange);
		StoragesStore.addChangeListener(this.onChange);
		OrganizationsStore.addChangeListener(this.onChange);
		DatacenterActions.sync();
		StorageActions.sync();
		OrganizationActions.sync();
	}

	componentWillUnmount(): void {
		DatacentersStore.removeChangeListener(this.onChange);
		StoragesStore.removeChangeListener(this.onChange);
		OrganizationsStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		let datacenters = DatacentersStore.datacenters;
		let selected: Selected = {};
		let curSelected = this.state.selected;
		let opened: Opened = {};
		let curOpened = this.state.opened;

		datacenters.forEach((datacenter: DatacenterTypes.Datacenter): void => {
			if (curSelected[datacenter.id]) {
				selected[datacenter.id] = true;
			}
			if (curOpened[datacenter.id]) {
				opened[datacenter.id] = true;
			}
		});

		this.setState({
			...this.state,
			datacenters: DatacentersStore.datacenters,
			storages: StoragesStore.storages,
			organizations: OrganizationsStore.organizations,
			filter: DatacentersStore.filter,
			selected: selected,
			opened: opened,
		});
	}

	onDelete = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		DatacenterActions.removeMulti(
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
		let datacentersDom: JSX.Element[] = [];

		this.state.datacenters.forEach((
				datacenter: DatacenterTypes.DatacenterRo): void => {
			datacentersDom.push(<Datacenter
				key={datacenter.id}
				datacenter={datacenter}
				storages={this.state.storages}
				organizations={this.state.organizations}
				selected={!!this.state.selected[datacenter.id]}
				open={!!this.state.opened[datacenter.id]}
				onSelect={(shift: boolean): void => {
					let selected = {
						...this.state.selected,
					};

					if (shift) {
						let datacenters = this.state.datacenters;
						let start: number;
						let end: number;

						for (let i = 0; i < datacenters.length; i++) {
							let usr = datacenters[i];

							if (usr.id === datacenter.id) {
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
								selected[datacenters[i].id] = true;
							}

							this.setState({
								...this.state,
								lastSelected: datacenter.id,
								selected: selected,
							});

							return;
						}
					}

					if (selected[datacenter.id]) {
						delete selected[datacenter.id];
					} else {
						selected[datacenter.id] = true;
					}

					this.setState({
						...this.state,
						lastSelected: datacenter.id,
						selected: selected,
					});
				}}
				onOpen={(): void => {
					let opened = {
						...this.state.opened,
					};

					if (opened[datacenter.id]) {
						delete opened[datacenter.id];
					} else {
						opened[datacenter.id] = true;
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
			let inst = DatacentersStore.datacenter(instId);
			if (inst) {
				selectedNames.push(inst.name || instId);
			} else {
				selectedNames.push(instId);
			}
		}

		let newDatacenterDom: JSX.Element;
		if (this.state.newOpened) {
			newDatacenterDom = <DatacenterNew
				storages={this.state.storages}
				organizations={this.state.organizations}
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
					<h2 style={css.heading}>Datacenters</h2>
					<div className="flex"/>
					<div style={css.buttons}>
						<button
							className={filterClass}
							style={css.button}
							type="button"
							onClick={(): void => {
								if (this.state.filter === null) {
									DatacenterActions.filter({});
								} else {
									DatacenterActions.filter(null);
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
							confirmMsg="Permanently delete the selected datacenters"
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
			<DatacentersFilter
				filter={this.state.filter}
				onFilter={(filter): void => {
					DatacenterActions.filter(filter);
				}}
			/>
			<div style={css.itemsBox}>
				<div style={css.items}>
					{newDatacenterDom}
					{datacentersDom}
					<tr className="bp5-card bp5-row" style={css.placeholder}>
						<td colSpan={2} style={css.placeholder}/>
					</tr>
				</div>
			</div>
			<NonState
				hidden={!!datacentersDom.length}
				iconClass="bp5-icon-cloud"
				title="No datacenters"
				description="Add a new datacenter to get started."
			/>
			<DatacentersPage
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

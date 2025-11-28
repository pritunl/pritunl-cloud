/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as PoolTypes from '../types/PoolTypes';
import PoolsStore from '../stores/PoolsStore';
import * as PoolActions from '../actions/PoolActions';
import Pool from './Pool';
import PoolNew from './PoolNew';
import PoolsFilter from './PoolsFilter';
import PoolsPage from './PoolsPage';
import Page from './Page';
import PageHeader from './PageHeader';
import NonState from './NonState';
import ConfirmButton from './ConfirmButton';
import * as DatacenterTypes from "../types/DatacenterTypes";
import * as ZoneTypes from "../types/ZoneTypes";
import DatacentersStore from "../stores/DatacentersStore";
import ZonesStore from "../stores/ZonesStore";
import * as DatacenterActions from "../actions/DatacenterActions";
import * as ZoneActions from "../actions/ZoneActions";

interface Selected {
	[key: string]: boolean;
}

interface Opened {
	[key: string]: boolean;
}

interface State {
	pools: PoolTypes.PoolsRo;
	filter: PoolTypes.Filter;
	datacenters: DatacenterTypes.DatacentersRo;
	zones: ZoneTypes.ZonesRo;
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

export default class Pools extends React.Component<{}, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			pools: PoolsStore.pools,
			filter: PoolsStore.filter,
			datacenters: DatacentersStore.datacenters,
			zones: ZonesStore.zones,
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
		PoolsStore.addChangeListener(this.onChange);
		DatacentersStore.addChangeListener(this.onChange);
		ZonesStore.addChangeListener(this.onChange);
		PoolActions.sync();
		DatacenterActions.sync();
		ZoneActions.sync();
	}

	componentWillUnmount(): void {
		PoolsStore.removeChangeListener(this.onChange);
		DatacentersStore.removeChangeListener(this.onChange);
		ZonesStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		let pools = PoolsStore.pools;
		let selected: Selected = {};
		let curSelected = this.state.selected;
		let opened: Opened = {};
		let curOpened = this.state.opened;

		pools.forEach((pool: PoolTypes.Pool): void => {
			if (curSelected[pool.id]) {
				selected[pool.id] = true;
			}
			if (curOpened[pool.id]) {
				opened[pool.id] = true;
			}
		});

		this.setState({
			...this.state,
			pools: pools,
			filter: PoolsStore.filter,
			datacenters: DatacentersStore.datacenters,
			zones: ZonesStore.zones,
			selected: selected,
			opened: opened,
		});
	}

	onDelete = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		PoolActions.removeMulti(
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
		let poolsDom: JSX.Element[] = [];

		this.state.pools.forEach((
				pool: PoolTypes.PoolRo): void => {
			poolsDom.push(<Pool
				key={pool.id}
				pool={pool}
				datacenters={this.state.datacenters}
				zones={this.state.zones}
				selected={!!this.state.selected[pool.id]}
				open={!!this.state.opened[pool.id]}
				onSelect={(shift: boolean): void => {
					let selected = {
						...this.state.selected,
					};

					if (shift) {
						let pools = this.state.pools;
						let start: number;
						let end: number;

						for (let i = 0; i < pools.length; i++) {
							let usr = pools[i];

							if (usr.id === pool.id) {
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
								selected[pools[i].id] = true;
							}

							this.setState({
								...this.state,
								lastSelected: pool.id,
								selected: selected,
							});

							return;
						}
					}

					if (selected[pool.id]) {
						delete selected[pool.id];
					} else {
						selected[pool.id] = true;
					}

					this.setState({
						...this.state,
						lastSelected: pool.id,
						selected: selected,
					});
				}}
				onOpen={(): void => {
					let opened = {
						...this.state.opened,
					};

					if (opened[pool.id]) {
						delete opened[pool.id];
					} else {
						opened[pool.id] = true;
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
			let inst = PoolsStore.pool(instId);
			if (inst) {
				selectedNames.push(inst.name || instId);
			} else {
				selectedNames.push(instId);
			}
		}

		let newPoolDom: JSX.Element;
		if (this.state.newOpened) {
			newPoolDom = <PoolNew
				datacenters={this.state.datacenters}
				zones={this.state.zones}
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
					<h2 style={css.heading}>Disk Pools</h2>
					<div className="flex"/>
					<div style={css.buttons}>
						<button
							className={filterClass}
							style={css.button}
							type="button"
							onClick={(): void => {
								if (this.state.filter === null) {
									PoolActions.filter({});
								} else {
									PoolActions.filter(null);
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
							confirmMsg="Permanently delete the selected pools"
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
			<PoolsFilter
				filter={this.state.filter}
				onFilter={(filter): void => {
					PoolActions.filter(filter);
				}}
			/>
			<div style={css.itemsBox}>
				<div style={css.items}>
					{newPoolDom}
					{poolsDom}
					<tr className="bp5-card bp5-row" style={css.placeholder}>
						<td colSpan={2} style={css.placeholder}/>
					</tr>
				</div>
			</div>
			<NonState
				hidden={!!poolsDom.length}
				iconClass="bp5-icon-control"
				title="No pools"
				description="Add a new pool to get started."
			/>
			<PoolsPage
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

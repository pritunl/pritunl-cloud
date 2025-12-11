/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as ZoneTypes from '../types/ZoneTypes';
import * as DatacenterTypes from '../types/DatacenterTypes';
import ZonesStore from '../stores/ZonesStore';
import CompletionStore from "../stores/CompletionStore";
import * as ZoneActions from '../actions/ZoneActions';
import * as CompletionActions from '../actions/CompletionActions';
import Zone from './Zone';
import ZoneNew from './ZoneNew';
import ZonesFilter from './ZonesFilter';
import ZonesPage from './ZonesPage';
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
	zones: ZoneTypes.ZonesRo;
	datacenters: DatacenterTypes.DatacentersRo;
	filter: ZoneTypes.Filter;
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

export default class Zones extends React.Component<{}, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			zones: ZonesStore.zones,
			datacenters: CompletionStore.datacenters,
			filter: ZonesStore.filter,
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
		ZonesStore.addChangeListener(this.onChange);
		CompletionStore.addChangeListener(this.onChange);
		ZoneActions.sync();
		CompletionActions.sync();
	}

	componentWillUnmount(): void {
		ZonesStore.removeChangeListener(this.onChange);
		CompletionStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		let zones = ZonesStore.zones;
		let selected: Selected = {};
		let curSelected = this.state.selected;
		let opened: Opened = {};
		let curOpened = this.state.opened;

		zones.forEach((zone: ZoneTypes.Zone): void => {
			if (curSelected[zone.id]) {
				selected[zone.id] = true;
			}
			if (curOpened[zone.id]) {
				opened[zone.id] = true;
			}
		});

		this.setState({
			...this.state,
			zones: zones,
			datacenters: CompletionStore.datacenters,
			filter: ZonesStore.filter,
			selected: selected,
			opened: opened,
		});
	}

	onDelete = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		ZoneActions.removeMulti(
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
		let zonesDom: JSX.Element[] = [];

		this.state.zones.forEach((
				zone: ZoneTypes.ZoneRo): void => {
			zonesDom.push(<Zone
				key={zone.id}
				zone={zone}
				datacenters={this.state.datacenters}
				selected={!!this.state.selected[zone.id]}
				open={!!this.state.opened[zone.id]}
				onSelect={(shift: boolean): void => {
					let selected = {
						...this.state.selected,
					};

					if (shift) {
						let zones = this.state.zones;
						let start: number;
						let end: number;

						for (let i = 0; i < zones.length; i++) {
							let usr = zones[i];

							if (usr.id === zone.id) {
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
								selected[zones[i].id] = true;
							}

							this.setState({
								...this.state,
								lastSelected: zone.id,
								selected: selected,
							});

							return;
						}
					}

					if (selected[zone.id]) {
						delete selected[zone.id];
					} else {
						selected[zone.id] = true;
					}

					this.setState({
						...this.state,
						lastSelected: zone.id,
						selected: selected,
					});
				}}
				onOpen={(): void => {
					let opened = {
						...this.state.opened,
					};

					if (opened[zone.id]) {
						delete opened[zone.id];
					} else {
						opened[zone.id] = true;
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
			let inst = ZonesStore.zone(instId);
			if (inst) {
				selectedNames.push(inst.name || instId);
			} else {
				selectedNames.push(instId);
			}
		}

		let newZoneDom: JSX.Element;
		if (this.state.newOpened) {
			newZoneDom = <ZoneNew
				datacenters={this.state.datacenters}
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
					<h2 style={css.heading}>Zones</h2>
					<div className="flex"/>
					<div style={css.buttons}>
						<button
							className={filterClass}
							style={css.button}
							type="button"
							onClick={(): void => {
								if (this.state.filter === null) {
									ZoneActions.filter({});
								} else {
									ZoneActions.filter(null);
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
							confirmMsg="Permanently delete the selected zones"
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
			<ZonesFilter
				filter={this.state.filter}
				onFilter={(filter): void => {
					ZoneActions.filter(filter);
				}}
			/>
			<div style={css.itemsBox}>
				<div style={css.items}>
					{newZoneDom}
					{zonesDom}
					<tr className="bp5-card bp5-row" style={css.placeholder}>
						<td colSpan={2} style={css.placeholder}/>
					</tr>
				</div>
			</div>
			<NonState
				hidden={!!zonesDom.length}
				iconClass="bp5-icon-ip-address"
				title="No zones"
				description="Add a new zone to get started."
			/>
			<ZonesPage
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

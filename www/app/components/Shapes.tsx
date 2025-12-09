/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as ShapeTypes from '../types/ShapeTypes';
import ShapesStore from '../stores/ShapesStore';
import * as ShapeActions from '../actions/ShapeActions';
import Shape from './Shape';
import ShapeNew from './ShapeNew';
import ShapesFilter from './ShapesFilter';
import ShapesPage from './ShapesPage';
import Page from './Page';
import PageHeader from './PageHeader';
import NonState from './NonState';
import ConfirmButton from './ConfirmButton';
import * as DatacenterTypes from "../types/DatacenterTypes";
import * as ZoneTypes from "../types/ZoneTypes";
import * as PoolTypes from "../types/PoolTypes";
import DatacentersStore from "../stores/DatacentersStore";
import ZonesStore from "../stores/ZonesStore";
import PoolsStore from "../stores/PoolsStore";
import * as DatacenterActions from "../actions/DatacenterActions";
import * as ZoneActions from "../actions/ZoneActions";
import * as PoolActions from "../actions/PoolActions";

interface Selected {
	[key: string]: boolean;
}

interface Opened {
	[key: string]: boolean;
}

interface State {
	shapes: ShapeTypes.ShapesRo;
	filter: ShapeTypes.Filter;
	datacenters: DatacenterTypes.DatacentersRo;
	zones: ZoneTypes.ZonesRo;
	pools: PoolTypes.PoolsRo;
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

export default class Shapes extends React.Component<{}, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			shapes: ShapesStore.shapes,
			filter: ShapesStore.filter,
			datacenters: DatacentersStore.datacenters,
			zones: ZonesStore.zones,
			pools: PoolsStore.pools,
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
		ShapesStore.addChangeListener(this.onChange);
		DatacentersStore.addChangeListener(this.onChange);
		ZonesStore.addChangeListener(this.onChange);
		PoolsStore.addChangeListener(this.onChange);
		ShapeActions.sync();
		DatacenterActions.sync();
		ZoneActions.sync();
		PoolActions.sync();
	}

	componentWillUnmount(): void {
		ShapesStore.removeChangeListener(this.onChange);
		DatacentersStore.removeChangeListener(this.onChange);
		ZonesStore.removeChangeListener(this.onChange);
		PoolsStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		let shapes = ShapesStore.shapes;
		let selected: Selected = {};
		let curSelected = this.state.selected;
		let opened: Opened = {};
		let curOpened = this.state.opened;

		shapes.forEach((shape: ShapeTypes.Shape): void => {
			if (curSelected[shape.id]) {
				selected[shape.id] = true;
			}
			if (curOpened[shape.id]) {
				opened[shape.id] = true;
			}
		});

		this.setState({
			...this.state,
			shapes: shapes,
			filter: ShapesStore.filter,
			datacenters: DatacentersStore.datacenters,
			zones: ZonesStore.zones,
			pools: PoolsStore.pools,
			selected: selected,
			opened: opened,
		});
	}

	onDelete = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		ShapeActions.removeMulti(
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
		let shapesDom: JSX.Element[] = [];

		this.state.shapes.forEach((
				shape: ShapeTypes.ShapeRo): void => {
			shapesDom.push(<Shape
				key={shape.id}
				shape={shape}
				datacenters={this.state.datacenters}
				zones={this.state.zones}
				pools={this.state.pools}
				selected={!!this.state.selected[shape.id]}
				open={!!this.state.opened[shape.id]}
				onSelect={(shift: boolean): void => {
					let selected = {
						...this.state.selected,
					};

					if (shift) {
						let shapes = this.state.shapes;
						let start: number;
						let end: number;

						for (let i = 0; i < shapes.length; i++) {
							let usr = shapes[i];

							if (usr.id === shape.id) {
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
								selected[shapes[i].id] = true;
							}

							this.setState({
								...this.state,
								lastSelected: shape.id,
								selected: selected,
							});

							return;
						}
					}

					if (selected[shape.id]) {
						delete selected[shape.id];
					} else {
						selected[shape.id] = true;
					}

					this.setState({
						...this.state,
						lastSelected: shape.id,
						selected: selected,
					});
				}}
				onOpen={(): void => {
					let opened = {
						...this.state.opened,
					};

					if (opened[shape.id]) {
						delete opened[shape.id];
					} else {
						opened[shape.id] = true;
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
			let inst = ShapesStore.shape(instId);
			if (inst) {
				selectedNames.push(inst.name || instId);
			} else {
				selectedNames.push(instId);
			}
		}

		let newShapeDom: JSX.Element;
		if (this.state.newOpened) {
			newShapeDom = <ShapeNew
				datacenters={this.state.datacenters}
				zones={this.state.zones}
				pools={this.state.pools}
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
					<h2 style={css.heading}>Instance Shapes</h2>
					<div className="flex"/>
					<div style={css.buttons}>
						<button
							className={filterClass}
							style={css.button}
							type="button"
							onClick={(): void => {
								if (this.state.filter === null) {
									ShapeActions.filter({});
								} else {
									ShapeActions.filter(null);
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
							confirmMsg="Permanently delete the selected shapes"
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
			<ShapesFilter
				filter={this.state.filter}
				onFilter={(filter): void => {
					ShapeActions.filter(filter);
				}}
			/>
			<div style={css.itemsBox}>
				<div style={css.items}>
					{newShapeDom}
					{shapesDom}
					<tr className="bp5-card bp5-row" style={css.placeholder}>
						<td colSpan={2} style={css.placeholder}/>
					</tr>
				</div>
			</div>
			<NonState
				hidden={!!shapesDom.length}
				iconClass="bp5-icon-zoom-to-fit"
				title="No shapes"
				description="Add a new shape to get started."
			/>
			<ShapesPage
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

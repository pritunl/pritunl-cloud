/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Constants from "../Constants";
import * as DiskTypes from '../types/DiskTypes';
import * as OrganizationTypes from '../types/OrganizationTypes';
import * as NodeTypes from '../types/NodeTypes';
import DisksStore from '../stores/DisksStore';
import CompletionStore from '../stores/CompletionStore';
import * as DiskActions from '../actions/DiskActions';
import * as CompletionActions from '../actions/CompletionActions';
import Disk from './Disk';
import DisksFilter from './DisksFilter';
import DisksPage from './DisksPage';
import DiskNew from './DiskNew';
import Page from './Page';
import PageHeader from './PageHeader';
import NonState from './NonState';
import ConfirmButton from './ConfirmButton';
import * as ZoneTypes from "../types/ZoneTypes";
import * as DatacenterTypes from "../types/DatacenterTypes";
import * as PoolTypes from "../types/PoolTypes";

interface Selected {
	[key: string]: boolean;
}

interface Opened {
	[key: string]: boolean;
}

interface State {
	disks: DiskTypes.DisksRo;
	filter: DiskTypes.Filter;
	debug: boolean;
	organizations: OrganizationTypes.OrganizationsRo;
	datacenters: DatacenterTypes.DatacentersRo;
	zones: ZoneTypes.ZonesRo;
	nodes: NodeTypes.NodesRo;
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
	debug: {
		margin: '0 0 4px 0',
	} as React.CSSProperties,
	debugButton: {
		opacity: 0.5,
		margin: '8px 0 0 8px',
	} as React.CSSProperties,
};

export default class Disks extends React.Component<{}, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			disks: DisksStore.disks,
			filter: DisksStore.filter,
			debug: false,
			organizations: CompletionStore.organizations,
			datacenters: CompletionStore.datacenters,
			zones: CompletionStore.zones,
			nodes: CompletionStore.nodes,
			pools: CompletionStore.pools,
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
		DisksStore.addChangeListener(this.onChange);
		CompletionStore.addChangeListener(this.onChange);
		DiskActions.sync();
		CompletionActions.sync();
	}

	componentWillUnmount(): void {
		DisksStore.removeChangeListener(this.onChange);
		CompletionStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		let disks = DisksStore.disks;
		let selected: Selected = {};
		let curSelected = this.state.selected;
		let opened: Opened = {};
		let curOpened = this.state.opened;

		disks.forEach((disk: DiskTypes.Disk): void => {
			if (curSelected[disk.id]) {
				selected[disk.id] = true;
			}
			if (curOpened[disk.id]) {
				opened[disk.id] = true;
			}
		});

		this.setState({
			...this.state,
			disks: disks,
			filter: DisksStore.filter,
			organizations: CompletionStore.organizations,
			datacenters: CompletionStore.datacenters,
			zones: CompletionStore.zones,
			nodes: CompletionStore.nodes,
			pools: CompletionStore.pools,
			selected: selected,
			opened: opened,
		});
	}

	onDelete = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		DiskActions.removeMulti(
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

	onForceDelete = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		DiskActions.forceRemoveMulti(
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

	onSnapshot = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		DiskActions.updateMulti(
			Object.keys(this.state.selected), 'snapshot').then((): void => {
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

	onBackup = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		DiskActions.updateMulti(
			Object.keys(this.state.selected), 'backup').then((): void => {
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
		let disksDom: JSX.Element[] = [];

		this.state.disks.forEach((
			disk: DiskTypes.DiskRo): void => {
			disksDom.push(<Disk
				key={disk.id}
				disk={disk}
				organizations={this.state.organizations}
				pools={this.state.pools}
				selected={!!this.state.selected[disk.id]}
				open={!!this.state.opened[disk.id]}
				onSelect={(shift: boolean): void => {
					let selected = {
						...this.state.selected,
					};

					if (shift) {
						let disks = this.state.disks;
						let start: number;
						let end: number;

						for (let i = 0; i < disks.length; i++) {
							let usr = disks[i];

							if (usr.id === disk.id) {
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
								selected[disks[i].id] = true;
							}

							this.setState({
								...this.state,
								lastSelected: disk.id,
								selected: selected,
							});

							return;
						}
					}

					if (selected[disk.id]) {
						delete selected[disk.id];
					} else {
						selected[disk.id] = true;
					}

					this.setState({
						...this.state,
						lastSelected: disk.id,
						selected: selected,
					});
				}}
				onOpen={(): void => {
					let opened = {
						...this.state.opened,
					};

					if (opened[disk.id]) {
						delete opened[disk.id];
					} else {
						opened[disk.id] = true;
					}

					this.setState({
						...this.state,
						opened: opened,
					});
				}}
			/>);
		});

		let newDiskDom: JSX.Element;
		if (this.state.newOpened) {
			newDiskDom = <DiskNew
				organizations={this.state.organizations}
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

		let debugClass = 'bp5-button bp5-icon-console ';
		if (this.state.debug) {
			debugClass += 'bp5-active';
		}

		let filterClass = 'bp5-button bp5-intent-primary bp5-icon-filter ';
		if (this.state.filter) {
			filterClass += 'bp5-active';
		}

		let selectedNames: string[] = [];
		for (let instId of Object.keys(this.state.selected)) {
			let inst = DisksStore.disk(instId);
			if (inst) {
				selectedNames.push(inst.name || instId);
			} else {
				selectedNames.push(instId);
			}
		}

		return <Page>
			<PageHeader>
				<div className="layout horizontal wrap" style={css.header}>
					<h2 style={css.heading}>Disks</h2>
					<div className="flex"/>
					<div style={css.buttons}>
						<button
							className={debugClass}
							style={css.debugButton}
							hidden={Constants.user}
							type="button"
							onClick={(): void => {
								this.setState({
									...this.state,
									debug: !this.state.debug,
								});
							}}
						>
							Debug
						</button>
						<button
							className={filterClass}
							style={css.button}
							type="button"
							onClick={(): void => {
								if (this.state.filter === null) {
									DiskActions.filter({});
								} else {
									DiskActions.filter(null);
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
							label="Snapshot Selected"
							className="bp5-intent-primary bp5-icon-floppy-disk"
							progressClassName="bp5-intent-primary"
							safe={true}
							style={css.button}
							confirmMsg="Snapshot the selected disks"
							items={selectedNames}
							disabled={!this.selected || this.state.disabled}
							onConfirm={this.onSnapshot}
						/>
						<ConfirmButton
							label="Backup Selected"
							className="bp5-intent-primary bp5-icon-compressed"
							progressClassName="bp5-intent-primary"
							safe={true}
							style={css.button}
							confirmMsg="Backup the selected disks"
							items={selectedNames}
							disabled={!this.selected || this.state.disabled}
							onConfirm={this.onBackup}
						/>
						<ConfirmButton
							label="Delete Selected"
							className="bp5-intent-danger bp5-icon-delete"
							progressClassName="bp5-intent-danger"
							safe={true}
							style={css.button}
							confirmMsg="Permanently delete the selected disks"
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
				<div
					className="layout horizontal wrap"
					style={css.debug}
					hidden={!this.state.debug}
				>
					<ConfirmButton
						label="Force Delete Selected"
						className="bp5-intent-danger bp5-icon-warning-sign"
						progressClassName="bp5-intent-danger"
						safe={true}
						style={css.button}
						confirmMsg="Permanently force delete the selected disks"
						items={selectedNames}
						disabled={!this.selected || this.state.disabled}
						onConfirm={this.onForceDelete}
					/>
				</div>
			</PageHeader>
			<DisksFilter
				filter={this.state.filter}
				onFilter={(filter): void => {
					DiskActions.filter(filter);
				}}
				organizations={this.state.organizations}
				nodes={this.state.nodes}
			/>
			<div style={css.itemsBox}>
				<div style={css.items}>
					{newDiskDom}
					{disksDom}
					<tr className="bp5-card bp5-row" style={css.placeholder}>
						<td colSpan={5} style={css.placeholder}/>
					</tr>
				</div>
			</div>
			<NonState
				hidden={!!disksDom.length}
				iconClass="bp5-icon-floppy-disk"
				title="No disks"
				description="Add a new disk to get started."
			/>
			<DisksPage
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

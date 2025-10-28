/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Constants from '../Constants';
import * as MiscUtils from '../utils/MiscUtils';
import * as InstanceTypes from '../types/InstanceTypes';
import * as OrganizationTypes from '../types/OrganizationTypes';
import * as DomainTypes from '../types/DomainTypes';
import * as VpcTypes from '../types/VpcTypes';
import * as DatacenterTypes from '../types/DatacenterTypes';
import * as NodeTypes from '../types/NodeTypes';
import * as PoolTypes from '../types/PoolTypes';
import * as ZoneTypes from '../types/ZoneTypes';
import * as ShapeTypes from '../types/ShapeTypes';
import InstancesStore from '../stores/InstancesStore';
import CompletionStore from '../stores/CompletionStore';
import * as InstanceActions from '../actions/InstanceActions';
import * as CompletionActions from '../actions/CompletionActions';
import Instance from './Instance';
import InstanceNew from './InstanceNew';
import InstancesFilter from './InstancesFilter';
import InstancesPage from './InstancesPage';
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
	instances: InstanceTypes.InstancesRo;
	domains: DomainTypes.DomainsRo;
	filter: InstanceTypes.Filter;
	debug: boolean;
	organizations: OrganizationTypes.OrganizationsRo;
	vpcs: VpcTypes.VpcsRo;
	datacenters: DatacenterTypes.DatacentersRo;
	nodes: NodeTypes.NodesRo;
	pools: PoolTypes.PoolsRo;
	zones: ZoneTypes.ZonesRo;
	shapes: ShapeTypes.ShapesRo;
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

export default class Instances extends React.Component<{}, State> {
	sync: MiscUtils.SyncInterval;

	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			instances: InstancesStore.instances,
			filter: InstancesStore.filter,
			debug: false,
			organizations: CompletionStore.organizations,
			domains: CompletionStore.domains,
			vpcs: CompletionStore.vpcs,
			datacenters: CompletionStore.datacenters,
			nodes: CompletionStore.nodes,
			pools: CompletionStore.pools,
			zones: CompletionStore.zones,
			shapes: CompletionStore.shapes,
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
		InstancesStore.addChangeListener(this.onChange);
		CompletionStore.addChangeListener(this.onChange);
		InstanceActions.sync();
		CompletionActions.sync();

		this.sync = new MiscUtils.SyncInterval(
			() => InstanceActions.sync(true),
			3000,
		)
	}

	componentWillUnmount(): void {
		InstancesStore.removeChangeListener(this.onChange);
		CompletionStore.removeChangeListener(this.onChange);

		this.sync?.stop()
	}

	onChange = (): void => {
		let instances = InstancesStore.instances;
		let selected: Selected = {};
		let curSelected = this.state.selected;
		let opened: Opened = {};
		let curOpened = this.state.opened;

		instances.forEach((instance: InstanceTypes.Instance): void => {
			if (curSelected[instance.id]) {
				selected[instance.id] = true;
			}
			if (curOpened[instance.id]) {
				opened[instance.id] = true;
			}
		});

		this.setState({
			...this.state,
			instances: instances,
			filter: InstancesStore.filter,
			organizations: CompletionStore.organizations,
			domains: CompletionStore.domains,
			vpcs: CompletionStore.vpcs,
			datacenters: CompletionStore.datacenters,
			nodes: CompletionStore.nodes,
			pools: CompletionStore.pools,
			zones: CompletionStore.zones,
			shapes: CompletionStore.shapes,
			selected: selected,
			opened: opened,
		});
	}

	onDelete = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		InstanceActions.removeMulti(
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
		InstanceActions.forceRemoveMulti(
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

	updateSelected(state: string): void {
		this.setState({
			...this.state,
			disabled: true,
		});
		InstanceActions.updateMulti(
			Object.keys(this.state.selected), state).then((): void => {
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
	}

	render(): JSX.Element {
		let instancesDom: JSX.Element[] = [];

		this.state.instances.forEach((
				instance: InstanceTypes.InstanceRo): void => {
			instancesDom.push(<Instance
				key={instance.id}
				instance={instance}
				domains={this.state.domains}
				vpcs={this.state.vpcs}
				selected={!!this.state.selected[instance.id]}
				open={!!this.state.opened[instance.id]}
				onSelect={(shift: boolean): void => {
					let selected = {
						...this.state.selected,
					};

					if (shift) {
						let instances = this.state.instances;
						let start: number;
						let end: number;

						for (let i = 0; i < instances.length; i++) {
							let usr = instances[i];

							if (usr.id === instance.id) {
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
								selected[instances[i].id] = true;
							}

							this.setState({
								...this.state,
								lastSelected: instance.id,
								selected: selected,
							});

							return;
						}
					}

					if (selected[instance.id]) {
						delete selected[instance.id];
					} else {
						selected[instance.id] = true;
					}

					this.setState({
						...this.state,
						lastSelected: instance.id,
						selected: selected,
					});
				}}
				onOpen={(): void => {
					let opened = {
						...this.state.opened,
					};

					if (opened[instance.id]) {
						delete opened[instance.id];
					} else {
						opened[instance.id] = true;
					}

					this.setState({
						...this.state,
						opened: opened,
					});
				}}
			/>);
		});

		let newInstanceDom: JSX.Element;
		if (this.state.newOpened) {
			newInstanceDom = <InstanceNew
				organizations={this.state.organizations}
				vpcs={this.state.vpcs}
				domains={this.state.domains}
				datacenters={this.state.datacenters}
				pools={this.state.pools}
				zones={this.state.zones}
				shapes={this.state.shapes}
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
			let inst = InstancesStore.instance(instId);
			if (inst) {
				selectedNames.push(inst.name || instId);
			} else {
				selectedNames.push(instId);
			}
		}

		return <Page>
			<PageHeader>
				<div className="layout horizontal wrap" style={css.header}>
					<h2 style={css.heading}>Instances</h2>
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
									InstanceActions.filter({});
								} else {
									InstanceActions.filter(null);
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
							label="Start Selected"
							className="bp5-intent-success bp5-icon-power"
							progressClassName="bp5-intent-success"
							safe={true}
							style={css.button}
							confirmMsg="Start the selected instances"
							items={selectedNames}
							disabled={!this.selected || this.state.disabled}
							onConfirm={(): void => {
								this.updateSelected('start');
							}}
						/>
						<ConfirmButton
							label="Stop Selected"
							className="bp5-intent-warning bp5-icon-power"
							progressClassName="bp5-intent-warning"
							safe={true}
							style={css.button}
							confirmMsg="Stop the selected instances"
							items={selectedNames}
							disabled={!this.selected || this.state.disabled}
							onConfirm={(): void => {
								this.updateSelected('stop');
							}}
						/>
						<ConfirmButton
							label="Delete Selected"
							className="bp5-intent-danger bp5-icon-delete"
							progressClassName="bp5-intent-danger"
							safe={true}
							style={css.button}
							confirmMsg="Permanently delete the selected instances"
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
						style={css.button}
						confirmMsg="Permanently force delete the selected instances"
						items={selectedNames}
						disabled={!this.selected || this.state.disabled}
						onConfirm={this.onForceDelete}
					/>
				</div>
			</PageHeader>
			<InstancesFilter
				filter={this.state.filter}
				onFilter={(filter): void => {
					InstanceActions.filter(filter);
				}}
				nodes={this.state.nodes}
				zones={this.state.zones}
				vpcs={this.state.vpcs}
				organizations={this.state.organizations}
			/>
			<div style={css.itemsBox}>
				<div style={css.items}>
					{newInstanceDom}
					{instancesDom}
					<tr className="bp5-card bp5-row" style={css.placeholder}>
						<td colSpan={6} style={css.placeholder}/>
					</tr>
				</div>
			</div>
			<NonState
				hidden={!!instancesDom.length}
				iconClass="bp5-icon-dashboard"
				title="No instances"
				description="Add a new instance to get started."
			/>
			<InstancesPage
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

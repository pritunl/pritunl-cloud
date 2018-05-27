/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Constants from '../Constants';
import * as InstanceTypes from '../types/InstanceTypes';
import * as OrganizationTypes from '../types/OrganizationTypes';
import * as VpcTypes from '../types/VpcTypes';
import * as DatacenterTypes from '../types/DatacenterTypes';
import * as ZoneTypes from '../types/ZoneTypes';
import InstancesStore from '../stores/InstancesStore';
import OrganizationsStore from '../stores/OrganizationsStore';
import VpcsNameStore from '../stores/VpcsNameStore';
import DatacentersStore from '../stores/DatacentersStore';
import ZonesStore from '../stores/ZonesStore';
import * as InstanceActions from '../actions/InstanceActions';
import * as OrganizationActions from '../actions/OrganizationActions';
import * as VpcActions from '../actions/VpcActions';
import * as DatacenterActions from '../actions/DatacenterActions';
import * as ZoneActions from '../actions/ZoneActions';
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
	filter: InstanceTypes.Filter;
	debug: boolean;
	organizations: OrganizationTypes.OrganizationsRo;
	vpcs: VpcTypes.VpcsRo;
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
		opacity: 0.85,
		margin: '8px 0 0 8px',
	} as React.CSSProperties,
};

export default class Instances extends React.Component<{}, State> {
	interval: NodeJS.Timer;

	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			instances: InstancesStore.instances,
			filter: InstancesStore.filter,
			debug: false,
			organizations: OrganizationsStore.organizations,
			vpcs: VpcsNameStore.vpcs,
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
		InstancesStore.addChangeListener(this.onChange);
		OrganizationsStore.addChangeListener(this.onChange);
		VpcsNameStore.addChangeListener(this.onChange);
		DatacentersStore.addChangeListener(this.onChange);
		ZonesStore.addChangeListener(this.onChange);
		InstanceActions.sync();
		OrganizationActions.sync();
		VpcActions.syncNames();
		DatacenterActions.sync();
		ZoneActions.sync();

		this.interval = setInterval(() => {
			InstanceActions.sync(true);
		}, 1000);
	}

	componentWillUnmount(): void {
		InstancesStore.removeChangeListener(this.onChange);
		OrganizationsStore.removeChangeListener(this.onChange);
		VpcsNameStore.removeChangeListener(this.onChange);
		DatacentersStore.removeChangeListener(this.onChange);
		ZonesStore.removeChangeListener(this.onChange);
		clearInterval(this.interval);
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
			organizations: OrganizationsStore.organizations,
			vpcs: VpcsNameStore.vpcs,
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
		let instancesDom: JSX.Element[] = [];

		this.state.instances.forEach((
				instance: InstanceTypes.InstanceRo): void => {
			instancesDom.push(<Instance
				key={instance.id}
				instance={instance}
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

		let debugClass = 'pt-button pt-intent-danger pt-icon-console ';
		if (this.state.debug) {
			debugClass += 'pt-active';
		}

		let filterClass = 'pt-button pt-intent-primary pt-icon-filter ';
		if (this.state.filter) {
			filterClass += 'pt-active';
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
							className="pt-button pt-intent-warning pt-icon-chevron-up"
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
							className="pt-intent-success pt-icon-power"
							progressClassName="pt-intent-success"
							style={css.button}
							disabled={!this.selected || this.state.disabled}
							onConfirm={(): void => {
								this.updateSelected('start');
							}}
						/>
						<ConfirmButton
							label="Stop Selected"
							className="pt-intent-danger pt-icon-power"
							progressClassName="pt-intent-danger"
							style={css.button}
							disabled={!this.selected || this.state.disabled}
							onConfirm={(): void => {
								this.updateSelected('stop');
							}}
						/>
						<ConfirmButton
							label="Delete Selected"
							className="pt-intent-danger pt-icon-delete"
							progressClassName="pt-intent-danger"
							style={css.button}
							disabled={!this.selected || this.state.disabled}
							onConfirm={this.onDelete}
						/>
						<button
							className="pt-button pt-intent-success pt-icon-add"
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
						className="pt-intent-danger pt-icon-warning-sign"
						progressClassName="pt-intent-danger"
						style={css.button}
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
				organizations={this.state.organizations}
			/>
			<div style={css.itemsBox}>
				<div style={css.items}>
					{newInstanceDom}
					{instancesDom}
					<tr className="pt-card pt-row" style={css.placeholder}>
						<td colSpan={5} style={css.placeholder}/>
					</tr>
				</div>
			</div>
			<NonState
				hidden={!!instancesDom.length}
				iconClass="pt-icon-dashboard"
				title="No instances"
				description="Add a new instance to get started."
			/>
			<InstancesPage
				onPage={(): void => {
					this.setState({
						lastSelected: null,
					});
				}}
			/>
		</Page>;
	}
}

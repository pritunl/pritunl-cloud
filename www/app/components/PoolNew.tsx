/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as PoolTypes from '../types/PoolTypes';
import * as OrganizationTypes from '../types/OrganizationTypes';
import * as DatacenterTypes from '../types/DatacenterTypes';
import * as NodeTypes from '../types/NodeTypes';
import * as InstanceTypes from '../types/InstanceTypes';
import * as ImageTypes from '../types/ImageTypes';
import * as ZoneTypes from '../types/ZoneTypes';
import * as PoolActions from '../actions/PoolActions';
import * as ImageActions from '../actions/ImageActions';
import * as InstanceActions from '../actions/InstanceActions';
import * as NodeActions from '../actions/NodeActions';
import ImagesDatacenterStore from '../stores/ImagesDatacenterStore';
import InstancesNodeStore from '../stores/InstancesNodeStore';
import NodesZoneStore from '../stores/NodesZoneStore';
import PageInput from './PageInput';
import PageInputButton from './PageInputButton';
import PageCreate from './PageCreate';
import PageSelect from './PageSelect';
import PageSwitch from "./PageSwitch";
import PageNumInput from './PageNumInput';
import Help from './Help';

interface Props {
	datacenters: DatacenterTypes.DatacentersRo;
	zones: ZoneTypes.ZonesRo;
	onClose: () => void;
}

interface State {
	closed: boolean;
	disabled: boolean;
	changed: boolean;
	message: string;
	pool: PoolTypes.Pool;
	datacenter: string;
	zone: string;
}

const css = {
	row: {
		display: 'table-row',
		width: '100%',
		padding: 0,
		boxShadow: 'none',
		position: 'relative',
	} as React.CSSProperties,
	card: {
		position: 'relative',
		padding: '10px 10px 0 10px',
		width: '100%',
	} as React.CSSProperties,
	buttons: {
		position: 'absolute',
		top: '5px',
		right: '5px',
	} as React.CSSProperties,
	item: {
		margin: '9px 5px 0 5px',
		minHeight: '20px',
	} as React.CSSProperties,
	itemsLabel: {
		display: 'block',
	} as React.CSSProperties,
	itemsAdd: {
		margin: '8px 0 15px 0',
	} as React.CSSProperties,
	group: {
		flex: 1,
		minWidth: '280px',
		margin: '0 10px',
	} as React.CSSProperties,
	save: {
		paddingBottom: '10px',
	} as React.CSSProperties,
	label: {
		width: '100%',
		maxWidth: '280px',
	} as React.CSSProperties,
	inputGroup: {
		width: '100%',
	} as React.CSSProperties,
	protocol: {
		flex: '0 1 auto',
	} as React.CSSProperties,
	port: {
		flex: '1',
	} as React.CSSProperties,
	role: {
		margin: '9px 5px 0 5px',
		minHeight: '20px',
	} as React.CSSProperties,
};

export default class PoolNew extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			closed: false,
			disabled: false,
			changed: false,
			message: '',
			pool: {
				name: 'new-pool',
			},
			datacenter: '',
			zone: '',
		};
	}

	set(name: string, val: any): void {
		let pool: any = {
			...this.state.pool,
		};

		pool[name] = val;

		this.setState({
			...this.state,
			changed: true,
			pool: pool,
		});
	}

	onCreate = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});

		let pool: any = {
			...this.state.pool,
		};

		let zone = this.state.zone;
		if (!zone && this.props.datacenters.length &&
			this.props.zones.length) {
			let datacenter = this.state.datacenter ||
				this.props.datacenters[0].id;
			for (let zne of this.props.zones) {
				if (zne.datacenter === datacenter) {
					zone = zne.id;
				}
			}
		}

		if (zone) {
			pool.zone = zone;
		}

		PoolActions.create(pool).then((): void => {
			this.setState({
				...this.state,
				message: 'Pool created successfully',
				changed: false,
			});

			setTimeout((): void => {
				this.setState({
					...this.state,
					disabled: false,
					changed: true,
				});
			}, 2000);
		}).catch((): void => {
			this.setState({
				...this.state,
				message: '',
				disabled: false,
			});
		});
	}

	render(): JSX.Element {
		let pool = this.state.pool;

		let defaultDatacenter = '';
		let hasDatacenters = false;
		let datacentersSelect: JSX.Element[] = [];
		if (this.props.datacenters.length) {
			hasDatacenters = true;
			defaultDatacenter = this.props.datacenters[0].id;
			for (let datacenter of this.props.datacenters) {
				datacentersSelect.push(
					<option
						key={datacenter.id}
						value={datacenter.id}
					>{datacenter.name}</option>,
				);
			}
		}

		if (!hasDatacenters) {
			datacentersSelect.push(
				<option key="null" value="">No Datacenters</option>);
		}

		let datacenter = this.state.datacenter || defaultDatacenter;
		let hasZones = false;
		let zonesSelect: JSX.Element[] = [];
		if (this.props.zones.length) {
			zonesSelect.push(<option key="null" value="">Select Zone</option>);

			for (let zone of this.props.zones) {
				if (!this.state.zone && zone.datacenter !== datacenter) {
					continue;
				}
				hasZones = true;

				zonesSelect.push(
					<option
						key={zone.id}
						value={zone.id}
					>{zone.name}</option>,
				);
			}
		}

		if (!hasZones) {
			zonesSelect = [<option key="null" value="">No Zones</option>];
		}

		return <div
			className="bp5-card bp5-row"
			style={css.row}
		>
			<td
				className="bp5-cell"
				colSpan={2}
				style={css.card}
			>
				<div className="layout horizontal wrap">
					<div style={css.group}>
						<div style={css.buttons}>
						</div>
						<PageInput
							label="Name"
							help="Name of pool"
							type="text"
							placeholder="Enter name"
							disabled={this.state.disabled}
							value={pool.name}
							onChange={(val): void => {
								this.set('name', val);
							}}
						/>
						<PageSelect
							disabled={this.state.disabled || !hasDatacenters}
							label="Datacenter"
							help="Datacenter for pool."
							value={this.state.datacenter}
							onChange={(val): void => {
								this.setState({
									...this.state,
									datacenter: val,
								});
							}}
						>
							{datacentersSelect}
						</PageSelect>
						<PageSelect
							disabled={this.state.disabled || !hasZones}
							label="Zone"
							help="Zone for pool."
							value={this.state.zone}
							onChange={(val): void => {
								this.setState({
									...this.state,
									zone: val,
								});
							}}
						>
							{zonesSelect}
						</PageSelect>
					</div>
					<div style={css.group}>
						<PageInput
							label="Volume Group Name"
							help="LVM volume group name. Name will be used to match nodes that have the LVM volume group available."
							type="text"
							placeholder="Enter name"
							value={pool.vg_name}
							onChange={(val): void => {
								this.set('vg_name', val);
							}}
						/>
						<PageSwitch
							disabled={this.state.disabled}
							label="Delete protection"
							help="Block pool from being deleted."
							checked={pool.delete_protection}
							onToggle={(): void => {
								this.set('delete_protection', !pool.delete_protection);
							}}
						/>
					</div>
				</div>
				<PageCreate
					style={css.save}
					hidden={!this.state.pool}
					message={this.state.message}
					changed={this.state.changed}
					disabled={this.state.disabled}
					closed={this.state.closed}
					light={true}
					onCancel={this.props.onClose}
					onCreate={this.onCreate}
				/>
			</td>
		</div>;
	}
}

/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as InstanceTypes from '../types/InstanceTypes';
import * as OrganizationTypes from '../types/OrganizationTypes';
import * as DatacenterTypes from '../types/DatacenterTypes';
import * as ZoneTypes from '../types/ZoneTypes';
import * as InstanceActions from '../actions/InstanceActions';
import PageInput from './PageInput';
import PageInfo from './PageInfo';
import PageCreate from './PageCreate';
import PageSelect from './PageSelect';
import PageNumInput from './PageNumInput';
import ConfirmButton from './ConfirmButton';

interface Props {
	organizations: OrganizationTypes.OrganizationsRo;
	datacenters: DatacenterTypes.DatacentersRo;
	zones: ZoneTypes.ZonesRo;
	onClose: () => void;
}

interface State {
	closed: boolean;
	disabled: boolean;
	changed: boolean;
	message: string;
	instance: InstanceTypes.Instance;
	datacenter: string;
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
		height: '20px',
	} as React.CSSProperties,
	itemsLabel: {
		display: 'block',
	} as React.CSSProperties,
	itemsAdd: {
		margin: '8px 0 15px 0',
	} as React.CSSProperties,
	group: {
		flex: 1,
		minWidth: '250px',
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
};

export default class InstanceNew extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			closed: false,
			disabled: false,
			changed: false,
			message: '',
			instance: this.default,
			datacenter: '',
		};
	}

	get default(): InstanceTypes.Instance {
		return {
			id: null,
			name: 'New instance',
			memory: 1024,
			processors: 1,
		};
	}

	set(name: string, val: any): void {
		let instance: any = {
			...this.state.instance,
		};

		instance[name] = val;

		this.setState({
			...this.state,
			changed: true,
			instance: instance,
		});
	}

	onCreate = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});

		let instance: any = {
			...this.state.instance,
		};

		if (this.props.organizations.length && !instance.organization) {
			instance.organization = this.props.organizations[0].id;
		}
		if (!instance.zone && this.props.datacenters.length &&
			this.props.zones.length) {

			let datacenter = this.state.datacenter || this.props.datacenters[0].id;
			for (let zone of this.props.zones) {
				if (zone.datacenter === datacenter) {
					instance.zone = zone.id;
				}
			}
		}

		InstanceActions.create(instance).then((): void => {
			this.setState({
				...this.state,
				message: 'Instance created successfully',
				closed: true,
				changed: false,
			});
		}).catch((): void => {
			this.setState({
				...this.state,
				message: '',
				disabled: false,
			});
		});
	}

	render(): JSX.Element {
		let instance = this.state.instance;

		let organizationsSelect: JSX.Element[] = [];
		if (this.props.organizations.length) {
			for (let organization of this.props.organizations) {
				organizationsSelect.push(
					<option
						key={organization.id}
						value={organization.id}
					>{organization.name}</option>,
				);
			}
		}

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
			for (let zone of this.props.zones) {
				if (zone.datacenter !== datacenter) {
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
			zonesSelect.push(<option key="null" value="">No Zones</option>);
		}

		return <div
			className="pt-card pt-row"
			style={css.row}
		>
			<td
				className="pt-cell"
				colSpan={4}
				style={css.card}
			>
				<div className="layout horizontal wrap">
					<div style={css.group}>
						<div style={css.buttons}>
						</div>
						<PageInput
							label="Name"
							help="Name of instance"
							type="text"
							placeholder="Enter name"
							disabled={this.state.disabled}
							value={instance.name}
							onChange={(val): void => {
								this.set('name', val);
							}}
						/>
						<PageSelect
							disabled={this.state.disabled}
							label="Organization"
							help="Organization for instance."
							value={instance.organization}
							onChange={(val): void => {
								this.set('organization', val);
							}}
						>
							{organizationsSelect}
						</PageSelect>
						<PageSelect
							disabled={this.state.disabled || !hasDatacenters}
							label="Datacenter"
							help="Datacenter for instance."
							value={this.state.datacenter}
							onChange={(val): void => {
								this.setState({
									...this.state,
									datacenter: val,
									instance: {
										...this.state.instance,
										zone: '',
									},
								});
							}}
						>
							{datacentersSelect}
						</PageSelect>
						<PageSelect
							disabled={this.state.disabled || !hasZones}
							label="Zone"
							help="Zone for instance."
							value={instance.zone}
							onChange={(val): void => {
								this.set('zone', val);
							}}
						>
							{zonesSelect}
						</PageSelect>
					</div>
					<div style={css.group}>
						<PageNumInput
							label="Memory Size"
							help="Instance memory size in megabytes."
							min={256}
							minorStepSize={256}
							stepSize={512}
							majorStepSize={1024}
							disabled={this.state.disabled}
							selectAllOnFocus={true}
							onChange={(val: number): void => {
								this.set('memory', val);
							}}
							value={instance.memory}
						/>
						<PageNumInput
							label="Processors"
							help="Number of instance processors."
							min={1}
							minorStepSize={1}
							stepSize={1}
							majorStepSize={2}
							disabled={this.state.disabled}
							selectAllOnFocus={true}
							onChange={(val: number): void => {
								this.set('processors', val);
							}}
							value={instance.processors}
						/>
					</div>
				</div>
				<PageCreate
					style={css.save}
					hidden={!this.state.instance}
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

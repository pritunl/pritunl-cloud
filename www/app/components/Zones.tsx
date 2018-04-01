/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as ZoneTypes from '../types/ZoneTypes';
import * as DatacenterTypes from '../types/DatacenterTypes';
import * as OrganizationTypes from '../types/OrganizationTypes';
import ZonesStore from '../stores/ZonesStore';
import DatacentersStore from "../stores/DatacentersStore";
import OrganizationsStore from "../stores/OrganizationsStore";
import * as ZoneActions from '../actions/ZoneActions';
import * as DatacenterActions from '../actions/DatacenterActions';
import * as OrganizationActions from '../actions/OrganizationActions';
import NonState from './NonState';
import Zone from './Zone';
import Page from './Page';
import PageHeader from './PageHeader';

interface State {
	zones: ZoneTypes.ZonesRo;
	datacenters: DatacenterTypes.DatacentersRo;
	organizations: OrganizationTypes.OrganizationsRo;
	datacenter: string;
	disabled: boolean;
}

const css = {
	header: {
		marginTop: '-19px',
	} as React.CSSProperties,
	heading: {
		margin: '19px 0 0 0',
	} as React.CSSProperties,
	group: {
		margin: '15px 0 0 0',
		width: '220px',
	} as React.CSSProperties,
	select: {
		width: '100%',
		borderTopLeftRadius: '3px',
		borderBottomLeftRadius: '3px',
	} as React.CSSProperties,
	selectInner: {
		width: '100%',
	} as React.CSSProperties,
	selectBox: {
		flex: '1',
	} as React.CSSProperties,
};

export default class Zones extends React.Component<{}, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			zones: ZonesStore.zones,
			datacenters: DatacentersStore.datacenters,
			organizations: OrganizationsStore.organizations,
			datacenter: '',
			disabled: false,
		};
	}

	componentDidMount(): void {
		ZonesStore.addChangeListener(this.onChange);
		DatacentersStore.addChangeListener(this.onChange);
		OrganizationsStore.addChangeListener(this.onChange);
		ZoneActions.sync();
		DatacenterActions.sync();
		OrganizationActions.sync();
	}

	componentWillUnmount(): void {
		ZonesStore.removeChangeListener(this.onChange);
		DatacentersStore.removeChangeListener(this.onChange);
		OrganizationsStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		this.setState({
			...this.state,
			zones: ZonesStore.zones,
			datacenters: DatacentersStore.datacenters,
			organizations: OrganizationsStore.organizations,
		});
	}

	render(): JSX.Element {
		let zonesDom: JSX.Element[] = [];

		this.state.zones.forEach((
				zone: ZoneTypes.ZoneRo): void => {
			zonesDom.push(<Zone
				key={zone.id}
				zone={zone}
				organizations={this.state.organizations}
			/>);
		});

		let hasDatacenters = false;
		let datacentersSelect: JSX.Element[] = [];
		if (this.state.datacenters.length) {
			hasDatacenters = true;
			for (let datacenter of this.state.datacenters) {
				datacentersSelect.push(
					<option
						key={datacenter.id}
						value={datacenter.id}
					>{datacenter.name}</option>,
				);
			}
		} else {
			datacentersSelect.push(
				<option
					key="null"
					value=""
				>No Datacenters</option>,
			);
		}

		return <Page>
			<PageHeader>
				<div className="layout horizontal wrap" style={css.header}>
					<h2 style={css.heading}>Zones</h2>
					<div className="flex"/>
					<div>
						<div
							className="pt-control-group"
							style={css.group}
						>
							<div style={css.selectBox}>
								<div className="pt-select" style={css.select}>
									<select
										style={css.selectInner}
										disabled={!hasDatacenters || this.state.disabled}
										value={this.state.datacenter}
										onChange={(evt): void => {
											this.setState({
												...this.state,
												datacenter: evt.target.value,
											});
										}}
									>
										{datacentersSelect}
									</select>
								</div>
							</div>
							<button
								className="pt-button pt-intent-success pt-icon-add"
								disabled={!hasDatacenters || this.state.disabled}
								type="button"
								onClick={(): void => {
									this.setState({
										...this.state,
										disabled: true,
									});
									ZoneActions.create({
										id: null,
										datacenter: this.state.datacenter ||
											this.state.datacenters[0].id,
									}).then((): void => {
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
								}}
							>New</button>
						</div>
					</div>
				</div>
			</PageHeader>
			<div>
				{zonesDom}
			</div>
			<NonState
				hidden={!!zonesDom.length}
				iconClass="pt-icon-layout-circle"
				title="No zones"
				description="Add a new zone to get started."
			/>
		</Page>;
	}
}

/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as ZoneTypes from '../types/ZoneTypes';
import * as OrganizationTypes from '../types/OrganizationTypes';
import ZonesStore from '../stores/ZonesStore';
import OrganizationsStore from "../stores/OrganizationsStore";
import * as ZoneActions from '../actions/ZoneActions';
import * as OrganizationActions from '../actions/OrganizationActions';
import NonState from './NonState';
import Zone from './Zone';
import Page from './Page';
import PageHeader from './PageHeader';

interface State {
	zones: ZoneTypes.ZonesRo;
	organizations: OrganizationTypes.OrganizationsRo;
	disabled: boolean;
}

const css = {
	header: {
		marginTop: '-19px',
	} as React.CSSProperties,
	heading: {
		margin: '19px 0 0 0',
	} as React.CSSProperties,
	button: {
		margin: '10px 0 0 0',
	} as React.CSSProperties,
	noCerts: {
		height: 'auto',
	} as React.CSSProperties,
};

export default class Zones extends React.Component<{}, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			zones: ZonesStore.zones,
			organizations: OrganizationsStore.organizations,
			disabled: false,
		};
	}

	componentDidMount(): void {
		ZonesStore.addChangeListener(this.onChange);
		OrganizationsStore.addChangeListener(this.onChange);
		ZoneActions.sync();
		OrganizationActions.sync();
	}

	componentWillUnmount(): void {
		ZonesStore.removeChangeListener(this.onChange);
		OrganizationsStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		this.setState({
			...this.state,
			zones: ZonesStore.zones,
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

		return <Page>
			<PageHeader>
				<div className="layout horizontal wrap" style={css.header}>
					<h2 style={css.heading}>Zones</h2>
					<div className="flex"/>
					<div>
						<button
							className="pt-button pt-intent-success pt-icon-add"
							style={css.button}
							disabled={this.state.disabled}
							type="button"
							onClick={(): void => {
								this.setState({
									...this.state,
									disabled: true,
								});
								ZoneActions.create(null).then((): void => {
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
			</PageHeader>
			<div>
				{zonesDom}
			</div>
			<NonState
				hidden={!!zonesDom.length}
				iconClass="pt-icon-layers"
				title="No zones"
				description="Add a new zone to get started."
			/>
		</Page>;
	}
}

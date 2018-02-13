/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as DatacenterTypes from '../types/DatacenterTypes';
import DatacentersStore from '../stores/DatacentersStore';
import * as DatacenterActions from '../actions/DatacenterActions';
import NonState from './NonState';
import Datacenter from './Datacenter';
import Page from './Page';
import PageHeader from './PageHeader';

interface State {
	datacenters: DatacenterTypes.DatacentersRo;
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

export default class Datacenters extends React.Component<{}, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			datacenters: DatacentersStore.datacenters,
			disabled: false,
		};
	}

	componentDidMount(): void {
		DatacentersStore.addChangeListener(this.onChange);
		DatacenterActions.sync();
	}

	componentWillUnmount(): void {
		DatacentersStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		this.setState({
			...this.state,
			datacenters: DatacentersStore.datacenters,
		});
	}

	render(): JSX.Element {
		let datacentersDom: JSX.Element[] = [];

		this.state.datacenters.forEach((
				datacenter: DatacenterTypes.DatacenterRo): void => {
			datacentersDom.push(<Datacenter
				key={datacenter.id}
				datacenter={datacenter}
			/>);
		});

		return <Page>
			<PageHeader>
				<div className="layout horizontal wrap" style={css.header}>
					<h2 style={css.heading}>Datacenters</h2>
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
								DatacenterActions.create(null).then((): void => {
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
				{datacentersDom}
			</div>
			<NonState
				hidden={!!datacentersDom.length}
				iconClass="pt-icon-endorsed"
				title="No datacenters"
				description="Add a new datacenter to get started."
			/>
		</Page>;
	}
}

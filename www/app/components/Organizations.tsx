/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as OrganizationTypes from '../types/OrganizationTypes';
import OrganizationsStore from '../stores/OrganizationsStore';
import * as OrganizationActions from '../actions/OrganizationActions';
import NonState from './NonState';
import Organization from './Organization';
import Page from './Page';
import PageHeader from './PageHeader';

interface State {
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
		margin: '15px 0 0 0',
	} as React.CSSProperties,
	noCerts: {
		height: 'auto',
	} as React.CSSProperties,
};

export default class Organizations extends React.Component<{}, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			organizations: OrganizationsStore.organizations,
			disabled: false,
		};
	}

	componentDidMount(): void {
		OrganizationsStore.addChangeListener(this.onChange);
		OrganizationActions.sync();
	}

	componentWillUnmount(): void {
		OrganizationsStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		this.setState({
			...this.state,
			organizations: OrganizationsStore.organizations,
		});
	}

	render(): JSX.Element {
		let orgsDom: JSX.Element[] = [];

		this.state.organizations.forEach((
				org: OrganizationTypes.OrganizationRo): void => {
			orgsDom.push(<Organization
				key={org.id}
				organization={org}
			/>);
		});

		return <Page>
			<PageHeader>
				<div className="layout horizontal wrap" style={css.header}>
					<h2 style={css.heading}>Organizations</h2>
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
								OrganizationActions.create(null).then((): void => {
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
				{orgsDom}
			</div>
			<NonState
				hidden={!!orgsDom.length}
				iconClass="pt-icon-people"
				title="No organizations"
				description="Add a new organization to get started."
			/>
		</Page>;
	}
}

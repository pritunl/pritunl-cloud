/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as SecretTypes from '../types/SecretTypes';
import * as OrganizationTypes from '../types/OrganizationTypes';
import SecretsStore from '../stores/SecretsStore';
import OrganizationsStore from '../stores/OrganizationsStore';
import * as SecretActions from '../actions/SecretActions';
import * as OrganizationActions from '../actions/OrganizationActions';
import NonState from './NonState';
import Secret from './Secret';
import Page from './Page';
import PageHeader from './PageHeader';

interface State {
	secrets: SecretTypes.SecretsRo;
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
		margin: '8px 0 0 8px',
	} as React.CSSProperties,
	buttons: {
		marginTop: '8px',
	} as React.CSSProperties,
	noCerts: {
		height: 'auto',
	} as React.CSSProperties,
};

export default class Secrets extends React.Component<{}, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			secrets: SecretsStore.secrets,
			organizations: OrganizationsStore.organizations,
			disabled: false,
		};
	}

	componentDidMount(): void {
		SecretsStore.addChangeListener(this.onChange);
		OrganizationsStore.addChangeListener(this.onChange);
		SecretActions.sync();
		OrganizationActions.sync();
	}

	componentWillUnmount(): void {
		SecretsStore.removeChangeListener(this.onChange);
		OrganizationsStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		this.setState({
			...this.state,
			secrets: SecretsStore.secrets,
			organizations: OrganizationsStore.organizations,
		});
	}

	render(): JSX.Element {
		let certsDom: JSX.Element[] = [];

		this.state.secrets.forEach((
				cert: SecretTypes.SecretRo): void => {
			certsDom.push(<Secret
				key={cert.id}
				secret={cert}
				organizations={this.state.organizations}
			/>);
		});

		return <Page>
			<PageHeader>
				<div className="layout horizontal wrap" style={css.header}>
					<h2 style={css.heading}>Secrets</h2>
					<div className="flex"/>
					<div style={css.buttons}>
						<button
							className="bp5-button bp5-intent-success bp5-icon-add"
							style={css.button}
							disabled={this.state.disabled}
							type="button"
							onClick={(): void => {
								this.setState({
									...this.state,
									disabled: true,
								});
								SecretActions.create(null).then((): void => {
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
				{certsDom}
			</div>
			<NonState
				hidden={!!certsDom.length}
				iconClass="bp5-icon-key"
				title="No secrets"
				description="Add a new secret to get started."
			/>
		</Page>;
	}
}

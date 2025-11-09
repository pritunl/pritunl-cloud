/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as OrganizationTypes from '../types/OrganizationTypes';
import * as CompletionActions from '../actions/CompletionActions';
import CompletionStore from '../stores/CompletionStore';

interface Props {
	hidden: boolean;
}

interface State {
	organizations: OrganizationTypes.OrganizationsRo;
	organization: string;
}

const css = {
	select: {
		marginRight: '11px',
	} as React.CSSProperties,
};

export default class Organization extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			organizations: CompletionStore.organizations,
			organization: null,
		};
	}

	componentDidMount(): void {
		CompletionStore.addChangeListener(this.onChange);
	}

	componentWillUnmount(): void {
		CompletionStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		this.setState({
			...this.state,
			organizations: CompletionStore.organizations,
			organization: CompletionStore.userOrganization,
		});
	}

	render(): JSX.Element {
		let orgsSelect: JSX.Element[] = [];

		this.state.organizations.forEach((
				org: OrganizationTypes.OrganizationRo): void => {
			orgsSelect.push(
				<option
					key={org.id}
					value={org.id}
				>{org.name}</option>,
			);
		});

		return <div style={css.select}>
			<div
				className="bp5-select"
				hidden={this.props.hidden}
			>
				<select
					value={this.state.organization || ''}
					onChange={(evt): void => {
						CompletionActions.setUserOrganization(evt.target.value);
					}}
				>
					{orgsSelect}
				</select>
			</div>
		</div>;
	}
}

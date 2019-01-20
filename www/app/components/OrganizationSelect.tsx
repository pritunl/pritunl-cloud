/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as OrganizationTypes from '../types/OrganizationTypes';
import * as OrganizationActions from '../actions/OrganizationActions';
import OrganizationsStore from '../stores/OrganizationsStore';

interface Props {
	hidden: boolean;
}

interface State {
	organizations: OrganizationTypes.OrganizationsRo;
	current: string;
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
			organizations: OrganizationsStore.organizations,
			current: null,
		};
	}

	componentDidMount(): void {
		OrganizationsStore.addChangeListener(this.onChange);
	}

	componentWillUnmount(): void {
		OrganizationsStore.removeChangeListener(this.onChange);
	}

	onChange = (): void => {
		this.setState({
			...this.state,
			organizations: OrganizationsStore.organizations,
			current: OrganizationsStore.current,
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
				className="bp3-select"
				hidden={this.props.hidden}
			>
				<select
					value={this.state.current || ''}
					onChange={(evt): void => {
						OrganizationActions.setCurrent(evt.target.value);
					}}
				>
					{orgsSelect}
				</select>
			</div>
		</div>;
	}
}

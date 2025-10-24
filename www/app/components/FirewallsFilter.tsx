/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Constants from '../Constants';
import * as FirewallTypes from '../types/FirewallTypes';
import SearchInput from './SearchInput';
import * as OrganizationTypes from "../types/OrganizationTypes";

interface Props {
	filter: FirewallTypes.Filter;
	onFilter: (filter: FirewallTypes.Filter) => void;
	organizations: OrganizationTypes.OrganizationsRo;
}

const css = {
	filters: {
		margin: '-15px 0 5px 0',
	} as React.CSSProperties,
	input: {
		width: '200px',
		margin: '5px',
	} as React.CSSProperties,
	shortInput: {
		width: '180px',
		margin: '5px',
	} as React.CSSProperties,
	role: {
		width: '150px',
		margin: '5px',
	} as React.CSSProperties,
	type: {
		margin: '5px',
	} as React.CSSProperties,
	check: {
		margin: '12px 5px 8px 5px',
	} as React.CSSProperties,
};

export default class FirewallsFilter extends React.Component<Props, {}> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			menu: false,
		};
	}

	render(): JSX.Element {
		if (this.props.filter === null) {
			return <div/>;
		}

		let organizationsSelect: JSX.Element[] = [
			<option key="key" value="any">Any</option>,
		];
		if (this.props.organizations && this.props.organizations.length) {
			for (let organization of this.props.organizations) {
				organizationsSelect.push(
					<option
						key={organization.id}
						value={organization.id}
					>{organization.name}</option>,
				);
			}
		}

		return <div className="layout horizontal wrap" style={css.filters}>
			<SearchInput
				style={css.input}
				placeholder="Firewall ID"
				value={this.props.filter.id}
				onChange={(val: string): void => {
					let filter = {
						...this.props.filter,
					};

					if (val) {
						filter.id = val;
					} else {
						delete filter.id;
					}

					this.props.onFilter(filter);
				}}
			/>
			<SearchInput
				style={css.input}
				placeholder="Name"
				value={this.props.filter.name}
				onChange={(val: string): void => {
					let filter = {
						...this.props.filter,
					};

					if (val) {
						filter.name = val;
					} else {
						delete filter.name;
					}

					this.props.onFilter(filter);
				}}
			/>
			<SearchInput
				style={css.shortInput}
				placeholder="Comment"
				value={this.props.filter.comment}
				onChange={(val: string): void => {
					let filter = {
						...this.props.filter,
					};

					if (val) {
						filter.comment = val;
					} else {
						delete filter.comment;
					}

					this.props.onFilter(filter);
				}}
			/>
			<SearchInput
				style={css.role}
				placeholder="Network Role"
				value={this.props.filter.role}
				onChange={(val: string): void => {
					let filter = {
						...this.props.filter,
					};

					if (val) {
						filter.role = val;
					} else {
						delete filter.role;
					}

					this.props.onFilter(filter);
				}}
			/>
			<div className="bp5-select" style={css.type} hidden={Constants.user}>
				<select
					value={this.props.filter.organization || 'any'}
					onChange={(evt): void => {
						let filter = {
							...this.props.filter,
						};

						let val = evt.target.value;

						if (val === 'any') {
							delete filter.organization;
						} else {
							filter.organization = val;
						}

						this.props.onFilter(filter);
					}}
				>
					{organizationsSelect}
				</select>
			</div>
		</div>;
	}
}

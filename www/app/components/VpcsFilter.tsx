/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Constants from '../Constants';
import * as VpcTypes from '../types/VpcTypes';
import SearchInput from './SearchInput';
import * as OrganizationTypes from "../types/OrganizationTypes";
import * as DatacenterTypes from "../types/DatacenterTypes";

interface Props {
	filter: VpcTypes.Filter;
	onFilter: (filter: VpcTypes.Filter) => void;
	organizations: OrganizationTypes.OrganizationsRo;
	datacenters: DatacenterTypes.DatacentersRo;
}

const css = {
	filters: {
		margin: '-15px 0 5px 0',
	} as React.CSSProperties,
	input: {
		width: '200px',
		margin: '5px',
	} as React.CSSProperties,
	keybase: {
		width: '175px',
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

export default class VpcsFilter extends React.Component<Props, {}> {
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

		let datacentersSelect: JSX.Element[] = [
			<option key="key" value="any">Any</option>,
		];
		if (this.props.datacenters && this.props.datacenters.length) {
			for (let datacenter of this.props.datacenters) {
				datacentersSelect.push(
					<option
						key={datacenter.id}
						value={datacenter.id}
					>{datacenter.name}</option>,
				);
			}
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
				style={css.role}
				placeholder="Network"
				value={this.props.filter.network}
				onChange={(val: string): void => {
					let filter = {
						...this.props.filter,
					};

					if (val) {
						filter.network = val;
					} else {
						delete filter.network;
					}

					this.props.onFilter(filter);
				}}
			/>
			<div className="pt-select" style={css.type}>
				<select
					value={this.props.filter.datacenter || 'any'}
					onChange={(evt): void => {
						let filter = {
							...this.props.filter,
						};

						let val = evt.target.value;

						if (val === 'any') {
							delete filter.datacenter;
						} else {
							filter.datacenter = val;
						}

						this.props.onFilter(filter);
					}}
				>
					{datacentersSelect}
				</select>
			</div>
			<div className="pt-select" style={css.type} hidden={Constants.user}>
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

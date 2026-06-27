/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Constants from '../Constants';
import * as AdvisoryTypes from '../types/AdvisoryTypes';
import SearchInput from './SearchInput';
import * as OrganizationTypes from "../types/OrganizationTypes";

interface Props {
	filter: AdvisoryTypes.Filter;
	onFilter: (filter: AdvisoryTypes.Filter) => void;
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
	type: {
		margin: '5px',
	} as React.CSSProperties,
};

export default class AdvisoriesFilter extends React.Component<Props, {}> {
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
				placeholder="Advisory ID"
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
				placeholder="Reference"
				value={this.props.filter.reference}
				onChange={(val: string): void => {
					let filter = {
						...this.props.filter,
					};

					if (val) {
						filter.reference = val;
					} else {
						delete filter.reference;
					}

					this.props.onFilter(filter);
				}}
			/>
			<div className="bp5-select" style={css.type}>
				<select
					value={this.props.filter.severity || 'any'}
					onChange={(evt): void => {
						let filter = {
							...this.props.filter,
						};

						let val = evt.target.value;

						if (val === 'any') {
							delete filter.severity;
						} else {
							filter.severity = val;
						}

						this.props.onFilter(filter);
					}}
				>
					<option key="any" value="any">Any Severity</option>
					<option key="critical" value="critical">Critical</option>
					<option key="important" value="important">Important</option>
					<option key="moderate" value="moderate">Moderate</option>
					<option key="low" value="low">Low</option>
					<option key="none" value="none">None</option>
				</select>
			</div>
			<div className="bp5-select" style={css.type}>
				<select
					value={this.props.filter.type || 'any'}
					onChange={(evt): void => {
						let filter = {
							...this.props.filter,
						};

						let val = evt.target.value;

						if (val === 'any') {
							delete filter.type;
						} else {
							filter.type = val;
						}

						this.props.onFilter(filter);
					}}
				>
					<option key="any" value="any">Any Type</option>
					<option key="rhel" value="rhel">Red Hat</option>
				</select>
			</div>
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

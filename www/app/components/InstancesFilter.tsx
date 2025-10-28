/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Constants from '../Constants';
import * as InstanceTypes from '../types/InstanceTypes';
import SearchInput from './SearchInput';
import * as OrganizationTypes from "../types/OrganizationTypes";
import * as NodeTypes from '../types/NodeTypes';
import * as VpcTypes from '../types/VpcTypes';
import * as ZoneTypes from '../types/ZoneTypes';

interface Props {
	filter: InstanceTypes.Filter;
	onFilter: (filter: InstanceTypes.Filter) => void;
	organizations: OrganizationTypes.OrganizationsRo;
	nodes: NodeTypes.NodesRo;
	zones: ZoneTypes.ZonesRo;
	vpcs: VpcTypes.VpcsRo;
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
	type: {
		margin: '5px',
	} as React.CSSProperties,
	check: {
		margin: '12px 5px 8px 5px',
	} as React.CSSProperties,
};

export default class InstancesFilter extends React.Component<Props, {}> {
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
			<option key="key" value="any">Any Organization</option>,
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

		let nodesSelect: JSX.Element[] = [
			<option key="key" value="any">Any Node</option>,
		];
		if (this.props.nodes && this.props.nodes.length) {
			for (let node of this.props.nodes) {
				nodesSelect.push(
					<option
						key={node.id}
						value={node.id}
					>{node.name}</option>,
				);
			}
		}

		let zonesSelect: JSX.Element[] = [
			<option key="key" value="any">Any Zone</option>,
		];
		if (this.props.zones && this.props.zones.length) {
			for (let zone of this.props.zones) {
				zonesSelect.push(
					<option
						key={zone.id}
						value={zone.id}
					>{zone.name}</option>,
				);
			}
		}

		let vpcsSelect: JSX.Element[] = [
			<option key="key" value="any">Any VPC</option>,
		];
		if (this.props.vpcs && this.props.vpcs.length) {
			for (let vpc of this.props.vpcs) {
				vpcsSelect.push(
					<option
						key={vpc.id}
						value={vpc.id}
					>{vpc.name}</option>,
				);
			}
		}

		let subnetsSelect: JSX.Element[] = [
			<option key="key" value="any">Any Subnet</option>,
		];
		if (this.props.vpcs && this.props.vpcs.length) {
			for (let vpc of this.props.vpcs) {
				if (vpc.id === this.props.filter.vpc) {
					for (let sub of (vpc.subnets || [])) {
						subnetsSelect.push(
							<option
								key={sub.id}
								value={sub.id}
							>{sub.name + ' - ' + sub.network}</option>,
						);
					}
				}
			}
		}

		return <div className="layout horizontal wrap" style={css.filters}>
			<SearchInput
				style={css.input}
				placeholder="Instance ID"
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
				style={css.shortInput}
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
			<SearchInput
				style={css.shortInput}
				placeholder="Network Namespace"
				value={this.props.filter.network_namespace}
				onChange={(val: string): void => {
					let filter = {
						...this.props.filter,
					};

					if (val) {
						filter.network_namespace = val;
					} else {
						delete filter.network_namespace;
					}

					this.props.onFilter(filter);
				}}
			/>
			<div className="bp5-select" style={css.type}>
				<select
					value={this.props.filter.node || 'any'}
					onChange={(evt): void => {
						let filter = {
							...this.props.filter,
						};

						let val = evt.target.value;

						if (val === 'any') {
							delete filter.node;
						} else {
							filter.node = val;
						}

						this.props.onFilter(filter);
					}}
				>
					{nodesSelect}
				</select>
			</div>
			<div className="bp5-select" style={css.type}>
				<select
					value={this.props.filter.zone || 'any'}
					onChange={(evt): void => {
						let filter = {
							...this.props.filter,
						};

						let val = evt.target.value;

						if (val === 'any') {
							delete filter.zone;
						} else {
							filter.zone = val;
						}

						this.props.onFilter(filter);
					}}
				>
					{zonesSelect}
				</select>
			</div>
			<div className="bp5-select" style={css.type}>
				<select
					value={this.props.filter.vpc || 'any'}
					onChange={(evt): void => {
						let filter = {
							...this.props.filter,
						};

						let val = evt.target.value;

						if (val === 'any') {
							delete filter.vpc;
							delete filter.subnet;
						} else {
							if (filter.vpc !== val) {
								filter.vpc = val;
								delete filter.subnet;
							}
						}

						this.props.onFilter(filter);
					}}
				>
					{vpcsSelect}
				</select>
			</div>
			<div className="bp5-select" style={css.type}>
				<select
					value={this.props.filter.subnet || 'any'}
					onChange={(evt): void => {
						let filter = {
							...this.props.filter,
						};

						let val = evt.target.value;

						if (val === 'any') {
							delete filter.subnet;
						} else {
							filter.subnet = val;
						}

						this.props.onFilter(filter);
					}}
				>
					{subnetsSelect}
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

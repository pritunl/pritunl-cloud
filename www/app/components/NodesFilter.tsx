/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as NodeTypes from '../types/NodeTypes';
import SearchInput from './SearchInput';
import SwitchNull from './SwitchNull';
import * as ZoneTypes from "../types/ZoneTypes";

interface Props {
	filter: NodeTypes.Filter;
	onFilter: (filter: NodeTypes.Filter) => void;
	zones: ZoneTypes.ZonesRo;
}

const css = {
	filters: {
		margin: '-15px 0 5px 0',
	} as React.CSSProperties,
	input: {
		width: '200px',
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

export default class NodesFilter extends React.Component<Props, {}> {
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

		let zonesSelect: JSX.Element[] = [
			<option key="key" value="any">Any</option>,
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

		return <div className="layout horizontal wrap" style={css.filters}>
			<SearchInput
				style={css.input}
				placeholder="Node ID"
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
			<SwitchNull
				style={css.check}
				label="Admin"
				checked={this.props.filter.admin}
				onToggle={(): void => {
					let filter = {
						...this.props.filter,
					};

					if (filter.admin === undefined) {
						filter.admin = true;
					} else if (filter.admin === true) {
						filter.admin = false;
					} else {
						delete filter.admin;
					}

					this.props.onFilter(filter);
				}}
			/>
			<SwitchNull
				style={css.check}
				label="User"
				checked={this.props.filter.user}
				onToggle={(): void => {
					let filter = {
						...this.props.filter,
					};

					if (filter.user === undefined) {
						filter.user = true;
					} else if (filter.user === true) {
						filter.user = false;
					} else {
						delete filter.user;
					}

					this.props.onFilter(filter);
				}}
			/>
			<SwitchNull
				style={css.check}
				label="Hypervisor"
				checked={this.props.filter.hypervisor}
				onToggle={(): void => {
					let filter = {
						...this.props.filter,
					};

					if (filter.hypervisor === undefined) {
						filter.hypervisor = true;
					} else if (filter.hypervisor === true) {
						filter.hypervisor = false;
					} else {
						delete filter.hypervisor;
					}

					this.props.onFilter(filter);
				}}
			/>
		</div>;
	}
}

/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as VpcTypes from '../types/VpcTypes';
import * as OrganizationTypes from "../types/OrganizationTypes";
import CompletionStore from '../stores/CompletionStore';
import VpcDetailed from './VpcDetailed';

interface Props {
	organizations: OrganizationTypes.OrganizationsRo;
	vpc: VpcTypes.VpcRo;
	selected: boolean;
	onSelect: (shift: boolean) => void;
	open: boolean;
	onOpen: () => void;
}

const css = {
	card: {
		display: 'table-row',
		width: '100%',
		padding: 0,
		boxShadow: 'none',
		cursor: 'pointer',
	} as React.CSSProperties,
	cardOpen: {
		display: 'table-row',
		width: '100%',
		padding: 0,
		boxShadow: 'none',
		position: 'relative',
	} as React.CSSProperties,
	select: {
		margin: '2px 0 0 0',
		paddingTop: '3px',
		minHeight: '18px',
	} as React.CSSProperties,
	name: {
		verticalAlign: 'top',
		display: 'table-cell',
		padding: '8px',
	} as React.CSSProperties,
	nameSpan: {
		margin: '1px 5px 0 0',
	} as React.CSSProperties,
	item: {
		verticalAlign: 'top',
		display: 'table-cell',
		padding: '9px',
		whiteSpace: 'nowrap',
	} as React.CSSProperties,
	icon: {
		marginRight: '3px',
	} as React.CSSProperties,
	bars: {
		verticalAlign: 'top',
		display: 'table-cell',
		padding: '8px',
		width: '30px',
	} as React.CSSProperties,
	bar: {
		height: '6px',
		marginBottom: '1px',
	} as React.CSSProperties,
	barLast: {
		height: '6px',
	} as React.CSSProperties,
	roles: {
		verticalAlign: 'top',
		display: 'table-cell',
		padding: '0 8px 8px 8px',
	} as React.CSSProperties,
	tag: {
		margin: '8px 5px 0 5px',
		height: '20px',
	} as React.CSSProperties,
};

export default class Vpc extends React.Component<Props, {}> {
	render(): JSX.Element {
		let vpc = this.props.vpc;

		if (this.props.open) {
			return <div
				className="bp5-card bp5-row"
				style={css.cardOpen}
			>
				<VpcDetailed
					organizations={this.props.organizations}
					vpc={this.props.vpc}
					selected={this.props.selected}
					onSelect={this.props.onSelect}
					onClose={(): void => {
						this.props.onOpen();
					}}
				/>
			</div>;
		}

		let cardStyle = {
			...css.card,
		};

		let orgName = '';
		if (vpc.organization) {
			let org = CompletionStore.organization(vpc.organization);
			orgName = org ? org.name : vpc.organization;
		} else {
			orgName = 'Unknown';
		}

		let datacenter = CompletionStore.datacenter(vpc.datacenter);
		let datacenterName = datacenter ? datacenter.name : vpc.datacenter;

		return <div
			className="bp5-card bp5-row"
			style={cardStyle}
			onClick={(evt): void => {
				let target = evt.target as HTMLElement;

				if (target.className.indexOf('open-ignore') !== -1) {
					return;
				}

				this.props.onOpen();
			}}
		>
			<div className="bp5-cell" style={css.name}>
				<div className="layout horizontal">
					<label
						className="bp5-control bp5-checkbox open-ignore"
						style={css.select}
					>
						<input
							type="checkbox"
							className="open-ignore"
							checked={this.props.selected}
							onChange={(evt): void => {
							}}
							onClick={(evt): void => {
								this.props.onSelect(evt.shiftKey);
							}}
						/>
						<span className="bp5-control-indicator open-ignore"/>
					</label>
					<div style={css.nameSpan}>
						{vpc.name}
					</div>
				</div>
			</div>
			<div className="bp5-cell" style={css.item}>
				<span
					style={css.icon}
					className={'bp5-icon-standard bp5-text-muted ' + (vpc.organization ?
						'bp5-icon-people' : 'bp5-icon-layers')}
				/>
				{orgName}
			</div>
			<div className="bp5-cell" style={css.item}>
				<span
					style={css.icon}
					className="bp5-icon-standard bp5-text-muted bp5-icon-cloud"
				/>
				{datacenterName}
			</div>
			<div className="bp5-cell" style={css.item}>
				<span
					style={css.icon}
					className="bp5-icon-standard bp5-text-muted bp5-icon-layout-auto"
				/>
				{vpc.network}
			</div>
		</div>;
	}
}

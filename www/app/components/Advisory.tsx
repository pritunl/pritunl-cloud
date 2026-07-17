/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as AdvisoryTypes from '../types/AdvisoryTypes';
import * as OrganizationTypes from "../types/OrganizationTypes";
import CompletionStore from '../stores/CompletionStore';
import AdvisoryDetailed from './AdvisoryDetailed';
import * as MiscUtils from '../utils/MiscUtils';

interface Props {
	organizations: OrganizationTypes.OrganizationsRo;
	advisory: AdvisoryTypes.AdvisoryRo;
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
		margin: '2px 5px 0 0',
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
};

export function severityClass(severity: string): string {
	switch ((severity || '').toLowerCase()) {
		case 'critical':
			return 'bp5-text-intent-danger';
		case 'important':
		case 'high':
			return 'bp5-text-intent-warning';
		case 'moderate':
		case 'medium':
			return 'bp5-text-intent-primary';
		case 'low':
			return 'bp5-text-intent-success';
		default:
			return 'bp5-text-muted';
	}
}

export function scoreLabel(score: number): string {
	switch (score) {
		case 1:
			return 'Low';
		case 2:
			return 'Medium';
		case 3:
			return 'High';
		case 4:
			return 'Critical';
		default:
			return 'Unknown';
	}
}

export default class Advisory extends React.Component<Props, {}> {
	render(): JSX.Element {
		let advisory = this.props.advisory;

		if (this.props.open) {
			return <div
				className="bp5-card bp5-row"
				style={css.cardOpen}
			>
				<AdvisoryDetailed
					organizations={this.props.organizations}
					advisory={this.props.advisory}
					selected={this.props.selected}
					onSelect={this.props.onSelect}
					onClose={(): void => {
						this.props.onOpen();
					}}
				/>
			</div>;
		}

		let orgName = '';
		if (advisory.organization) {
			let org = CompletionStore.organization(advisory.organization);
			orgName = org ? org.name : advisory.organization;
		} else {
			orgName = 'Node';
		}

		let severityText = MiscUtils.capitalize(advisory.severity) || 'Unknown';
		let scoreCls = 'bp5-cell ' + severityClass(scoreLabel(advisory.score));

		let instanceCount = (advisory.instances_info || []).length;
		let nodeCount = (advisory.nodes_info || []).length;

		let cardStyle = css.card;
		if (advisory.dismissed) {
			cardStyle = {
				...css.card,
				opacity: 0.5,
			};
		}

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
						{advisory.reference}
					</div>
				</div>
			</div>
			<div className={scoreCls} style={css.item}>
				<span
					style={css.icon}
					className="bp5-icon-standard bp5-icon-warning-sign"
				/>
				{scoreLabel(advisory.score)}
			</div>
			<div className="bp5-cell" style={css.item}>
				<span
					style={css.icon}
					className="bp5-icon-standard bp5-text-muted bp5-icon-pulse"
				/>
				{severityText}
			</div>
			<div className="bp5-cell" style={css.item}>
				<span
					style={css.icon}
					className={'bp5-icon-standard bp5-text-muted ' + (
						advisory.organization ? 'bp5-icon-people' : 'bp5-icon-layers')}
				/>
				{orgName}
			</div>
			<div className="bp5-cell" style={css.item}>
				<span
					style={css.icon}
					className="bp5-icon-standard bp5-text-muted bp5-icon-dashboard"
				/>
				{instanceCount} Instances
			</div>
			<div className="bp5-cell" style={css.item}>
				<span
					style={css.icon}
					className="bp5-icon-standard bp5-text-muted bp5-icon-layers"
				/>
				{nodeCount} Nodes
			</div>
		</div>;
	}
}

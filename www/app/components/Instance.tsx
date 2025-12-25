/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Blueprint from '@blueprintjs/core';
import * as InstanceTypes from '../types/InstanceTypes';
import * as VpcTypes from '../types/VpcTypes';
import * as DomainTypes from '../types/DomainTypes'
import InstanceDetailed from './InstanceDetailed';
import CompletionStore from "../stores/CompletionStore";

interface Props {
	vpcs: VpcTypes.VpcsRo;
	domains: DomainTypes.DomainsRo;
	instance: InstanceTypes.InstanceRo;
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
	status: {
		verticalAlign: 'top',
		display: 'table-cell',
		padding: '9px',
		whiteSpace: 'nowrap',
		width: '35px',
	} as React.CSSProperties,
	icon: {
		marginRight: '3px',
	} as React.CSSProperties,
	updateIcon: {
		marginRight: '5px',
	} as React.CSSProperties,
	bars: {
		verticalAlign: 'top',
		display: 'table-cell',
		padding: '8px',
		width: '85px',
	} as React.CSSProperties,
	bar: {
		height: '6px',
		marginBottom: '1px',
	} as React.CSSProperties,
	barLast: {
		height: '6px',
	} as React.CSSProperties,
};

export default class Instance extends React.Component<Props, {}> {
	render(): JSX.Element {
		let instance = this.props.instance;

		if (this.props.open) {
			return <div
				className="bp5-card bp5-row"
				style={css.cardOpen}
			>
				<InstanceDetailed
					instance={this.props.instance}
					vpcs={this.props.vpcs}
					domains={this.props.domains}
					selected={this.props.selected}
					onSelect={this.props.onSelect}
					onClose={(): void => {
						this.props.onOpen();
					}}
				/>
			</div>;
		}

		let node = CompletionStore.node(this.props.instance.node);
		let nodeName = node ? node.name : null;
		let zone = CompletionStore.zone(this.props.instance.zone);
		let zoneName = zone ? zone.name : null;

		let cardStyle = {
			...css.card,
		};

		let publicIp = '';
		let privateIp = '';
		if (instance.public_ips && instance.public_ips.length > 0) {
			publicIp = instance.public_ips[0];
		} else if (instance.host_ips && instance.host_ips.length > 0) {
			publicIp = instance.host_ips[0];
		}
		if (instance.private_ips && instance.private_ips.length > 0) {
			privateIp = instance.private_ips[0];
		}

		let memoryStyle: React.CSSProperties = {
			width: (instance.guest?.memory || 0) + '%',
		};
		let load1Style: React.CSSProperties = {
			width: (instance.guest?.load1 || 0) + '%',
		};
		let load5Style: React.CSSProperties = {
			width: (instance.guest?.load5 || 0) + '%',
		};

		let statusClass = 'bp5-cell';
		switch (instance.status) {
			case 'Running':
				statusClass += ' bp5-text-intent-success';
				break;
			case 'Stopped':
			case 'Failed':
			case 'Destroying':
				statusClass += ' bp5-text-intent-danger';
				break;
		}

		if (instance.status?.includes("Restart Required")) {
			statusClass += ' bp5-text-intent-warning';
		}

		let updateElm: JSX.Element
		let updateSev = 0
		let updateClass = ""
		instance.guest?.updates?.forEach((update) => {
			switch (update.severity) {
				case "moderate":
					if (updateSev < 1) {
						updateSev = 1
						updateClass = "bp5-icon-issue bp5-intent-primary"
					}
					break
				case "important":
					if (updateSev < 2) {
						updateSev = 2
						updateClass = "bp5-icon-issue bp5-intent-warning"
					}
					break
				case "critical":
					updateSev = 3
					updateClass = "bp5-icon-issue bp5-intent-danger"
					break
			}
		})

		if (updateClass) {
			updateElm = <Blueprint.Tooltip
				content="Updates available"
				openOnTargetFocus={false}
				compact={true}
				renderTarget={({...tooltipProps}) => (
					<span
						{...tooltipProps}
						style={css.updateIcon}
						hidden={!instance.status}
						className={"bp5-icon-standard " + updateClass}
					/>
				)}
			/>
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
						{updateElm}
						{instance.name}
					</div>
				</div>
			</div>
			<div className={statusClass} style={css.status}>
				<span
					style={css.icon}
					hidden={!instance.status}
					className="bp5-icon-standard bp5-icon-power"
				/>
			</div>
			<div className="bp5-cell" style={css.item}>
				<span
					style={css.icon}
					hidden={!nodeName}
					className="bp5-icon-standard bp5-text-muted bp5-icon-layers"
				/>
				{nodeName}
			</div>
			<div className="bp5-cell" style={css.item}>
				<span
					style={css.icon}
					hidden={!zoneName}
					className="bp5-icon-standard bp5-text-muted bp5-icon-layout-circle"
				/>
				{zoneName}
			</div>
			<div className="bp5-cell" style={css.item}>
				<span
					style={css.icon}
					hidden={!publicIp}
					className="bp5-icon-standard bp5-text-muted bp5-icon-ip-address"
				/>
				{publicIp}
			</div>
			<div className="bp5-cell" style={css.item}>
				<span
					style={css.icon}
					hidden={!privateIp}
					className="bp5-icon-standard bp5-text-muted bp5-icon-ip-address"
				/>
				{privateIp}
			</div>
			<div className="bp5-cell" style={css.bars}>
				<div
					className="bp5-progress-bar bp5-no-stripes bp5-intent-success"
					style={css.bar}
				>
					<div className="bp5-progress-meter" style={load1Style}/>
				</div>
				<div
					className="bp5-progress-bar bp5-no-stripes bp5-intent-warning"
					style={css.bar}
				>
					<div className="bp5-progress-meter" style={load5Style}/>
				</div>
				<div
					className="bp5-progress-bar bp5-no-stripes bp5-intent-primary"
					style={css.barLast}
				>
					<div className="bp5-progress-meter" style={memoryStyle}/>
				</div>
			</div>
		</div>;
	}
}

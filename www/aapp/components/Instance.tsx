/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as MiscUtils from '../utils/MiscUtils';
import * as InstanceTypes from '../types/InstanceTypes';
import InstanceDetailed from './InstanceDetailed';

interface Props {
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
		paddingTop: '1px',
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
};

export default class Instance extends React.Component<Props, {}> {
	render(): JSX.Element {
		let instance = this.props.instance;

		if (this.props.open) {
			return <div
				className="pt-card pt-row"
				style={css.cardOpen}
			>
				<InstanceDetailed
					instance={this.props.instance}
					selected={this.props.selected}
					onSelect={this.props.onSelect}
					onClose={(): void => {
						this.props.onOpen();
					}}
				/>
			</div>;
		}

		let active = true;

		let cardStyle = {
			...css.card,
		};
		if (!active) {
			cardStyle.opacity = 0.6;
		}

		let statusClass = 'pt-cell';
		if (instance.status === 'Running') {
			statusClass += ' pt-text-intent-success';
		} else if (instance.status === 'Stopped') {
			statusClass += ' pt-text-intent-danger';
		}

		return <div
			className="pt-card pt-row"
			style={cardStyle}
			onClick={(evt): void => {
				let target = evt.target as HTMLElement;

				if (target.className.indexOf('open-ignore') !== -1) {
					return;
				}

				this.props.onOpen();
			}}
		>
			<div className="pt-cell" style={css.name}>
				<div className="layout horizontal">
					<label
						className="pt-control pt-checkbox open-ignore"
						style={css.select}
					>
						<input
							type="checkbox"
							className="open-ignore"
							checked={this.props.selected}
							onClick={(evt): void => {
								this.props.onSelect(evt.shiftKey);
							}}
						/>
						<span className="pt-control-indicator open-ignore"/>
					</label>
					<div style={css.nameSpan}>
						{instance.name}
					</div>
				</div>
			</div>
			<div className={statusClass} style={css.item}>
				<span
					style={css.icon}
					hidden={!instance.status}
					className="pt-icon-standard pt-icon-power"
				/>
				{instance.status}
			</div>
			<div className="pt-cell" style={css.item}>
			</div>
			<div className="pt-cell" style={css.item}>
				<span
					style={css.icon}
					hidden={!instance.public_ip}
					className="pt-icon-standard pt-icon-ip-address"
				/>
				{instance.public_ip}
			</div>
			<div className="pt-cell" style={css.item}>
				<span
					style={css.icon}
					hidden={!instance.public_ip6}
					className="pt-icon-standard pt-icon-ip-address"
				/>
				{instance.public_ip6}
			</div>
		</div>;
	}
}

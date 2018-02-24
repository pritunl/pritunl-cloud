/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as MiscUtils from '../utils/MiscUtils';
import * as DiskTypes from '../types/DiskTypes';
import * as OrganizationTypes from "../types/OrganizationTypes";
import OrganizationsStore from '../stores/OrganizationsStore';
import DiskDetailed from './DiskDetailed';

interface Props {
	organizations: OrganizationTypes.OrganizationsRo;
	disk: DiskTypes.DiskRo;
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

export default class Disk extends React.Component<Props, {}> {
	render(): JSX.Element {
		let disk = this.props.disk;

		if (this.props.open) {
			return <div
				className="pt-card pt-row"
				style={css.cardOpen}
			>
				<DiskDetailed
					organizations={this.props.organizations}
					disk={this.props.disk}
					selected={this.props.selected}
					onSelect={this.props.onSelect}
					onClose={(): void => {
						this.props.onOpen();
					}}
				/>
			</div>;
		}

		let orgName = '';
		if (disk.organization) {
			let org = OrganizationsStore.organization(disk.organization);
			orgName = org ? org.name : disk.organization;
		} else {
			orgName = 'Node Disk';
		}

		let statusText = 'Unknown';
		let statusClass = 'pt-cell';
		switch (disk.state) {
			case 'provision':
				statusText = 'Provisioning';
				statusClass += ' pt-text-intent-primary';
				break;
			case 'available':
				if (disk.instance !== "") {
					statusText = 'Connected';
				} else {
					statusText = 'Available';
				}
				statusClass += ' pt-text-intent-success';
				break;
			case 'destroy':
				statusText = 'Destroying';
				statusClass += ' pt-text-intent-danger';
				break;
			case 'snapshot':
				statusText = 'Snapshotting';
				statusClass += ' pt-text-intent-primary';
				break;
		}

		return <div
			className="pt-card pt-row"
			style={css.card}
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
						{disk.name}
					</div>
				</div>
			</div>
			<div className={statusClass} style={css.item}>
				<span
					style={css.icon}
					className="pt-icon-standard pt-icon-pulse"
				/>
				{statusText}
			</div>
			<div className="pt-cell" style={css.item}>
				<span
					style={css.icon}
					className={'pt-icon-standard ' + (disk.organization ?
						'pt-icon-people' : 'pt-icon-layers')}
				/>
				{orgName}
			</div>
			<div className="pt-cell" style={css.item}>
				<span
					style={css.icon}
					className="pt-icon-standard pt-icon-database"
				/>
				{disk.size}GB
			</div>
		</div>;
	}
}

/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as MiscUtils from '../utils/MiscUtils';
import * as DiskTypes from '../types/DiskTypes';
import * as OrganizationTypes from "../types/OrganizationTypes";
import CompletionStore from '../stores/CompletionStore';
import DiskDetailed from './DiskDetailed';
import * as PoolTypes from "../types/PoolTypes";

interface Props {
	organizations: OrganizationTypes.OrganizationsRo;
	pools: PoolTypes.PoolsRo;
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

export default class Disk extends React.Component<Props, {}> {
	render(): JSX.Element {
		let disk = this.props.disk;

		if (this.props.open) {
			return <div
				className="bp5-card bp5-row"
				style={css.cardOpen}
			>
				<DiskDetailed
					organizations={this.props.organizations}
					pools={this.props.pools}
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
			let org = CompletionStore.organization(disk.organization);
			orgName = org ? org.name : disk.organization;
		} else {
			orgName = 'Unknown Organization';
		}

		let statusText = 'Unknown';
		let statusClass = 'bp5-cell';
		switch (disk.state) {
			case 'provision':
				statusText = 'Provisioning';
				statusClass += ' bp5-text-intent-primary';
				break;
			case 'available':
				if (disk.instance) {
					statusText = 'Connected';
				} else {
					statusText = 'Available';
				}
				statusClass += ' bp5-text-intent-success';
				break;
			case 'attached':
				statusText = 'Connected';
				statusClass += ' bp5-text-intent-success';
				break;
		}

		switch (disk.action) {
			case 'destroy':
				statusText = 'Destroying';
				statusClass += ' bp5-text-intent-danger';
				break;
			case 'snapshot':
				statusText = 'Snapshotting';
				statusClass += ' bp5-text-intent-primary';
				break;
			case 'backup':
				statusText = 'Backing Up';
				statusClass += ' bp5-text-intent-primary';
				break;
			case 'restore':
				statusText = 'Restoring';
				statusClass += ' bp5-text-intent-primary';
				break;
			case 'expand':
				statusText = 'Expanding';
				statusClass += ' bp5-text-intent-primary';
				break;
		}

		let resourceIcon = "";
		let resourceValue = "";
		if (this.props.disk.type === "lvm") {
			resourceIcon = "bp5-icon-control";
			resourceValue = "Pool Unavailable"

			if (this.props.pools.length) {
				for (let pool of this.props.pools) {
					if (pool.id === disk.pool) {
						resourceValue = pool.name;
						break;
					}
				}
			}
		} else {
			let node = CompletionStore.node(disk.node);
			resourceIcon = "bp5-icon-layers";
			resourceValue = node ? node.name : disk.node;
		}

		return <div
			className="bp5-card bp5-row"
			style={css.card}
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
						{disk.name}
					</div>
				</div>
			</div>
			<div className={statusClass} style={css.item}>
				<span
					style={css.icon}
					className="bp5-icon-standard bp5-icon-pulse"
				/>
				{statusText}
			</div>
			<div className="bp5-cell" style={css.item}>
				<span
					style={css.icon}
					className={'bp5-icon-standard bp5-text-muted ' + (disk.organization ?
						'bp5-icon-people' : 'bp5-icon-layers')}
				/>
				{orgName}
			</div>
			<div className="bp5-cell" style={css.item}>
				<span
					style={css.icon}
					className={"bp5-icon-standard bp5-text-muted " + resourceIcon}
				/>
				{resourceValue}
			</div>
			<div className="bp5-cell" style={css.item}>
				<span
					style={css.icon}
					className="bp5-icon-standard bp5-text-muted bp5-icon-database"
				/>
				{disk.size}GB
			</div>
		</div>;
	}
}

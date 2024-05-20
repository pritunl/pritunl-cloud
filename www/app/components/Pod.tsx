/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as MiscUtils from '../utils/MiscUtils';
import * as PodTypes from '../types/PodTypes';
import * as OrganizationTypes from "../types/OrganizationTypes";
import OrganizationsStore from '../stores/OrganizationsStore';
import PodDetailed from './PodDetailed';
import * as DomainTypes from "../types/DomainTypes";
import * as VpcTypes from "../types/VpcTypes";
import * as DatacenterTypes from "../types/DatacenterTypes";
import * as NodeTypes from "../types/NodeTypes";
import * as PoolTypes from "../types/PoolTypes";
import * as ZoneTypes from "../types/ZoneTypes";
import * as ShapeTypes from "../types/ShapeTypes";

interface Props {
	organizations: OrganizationTypes.OrganizationsRo;
	domains: DomainTypes.DomainsRo;
	vpcs: VpcTypes.VpcsRo;
	datacenters: DatacenterTypes.DatacentersRo;
	nodes: NodeTypes.NodesRo;
	pools: PoolTypes.PoolsRo;
	zones: ZoneTypes.ZonesRo;
	shapes: ShapeTypes.ShapesRo;
	pod: PodTypes.PodRo;
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

export default class Pod extends React.Component<Props, {}> {
	render(): JSX.Element {
		let pod = this.props.pod;

		if (this.props.open) {
			return <div
				className="bp3-card bp3-row"
				style={css.cardOpen}
			>
				<PodDetailed
					organizations={this.props.organizations}
					domains={this.props.domains}
					vpcs={this.props.vpcs}
					datacenters={this.props.datacenters}
					nodes={this.props.nodes}
					pools={this.props.pools}
					zones={this.props.zones}
					shapes={this.props.shapes}
					pod={this.props.pod}
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

		let role: JSX.Element[] = [];
		for (let networkRole of (pod.roles || [])) {
			role.push(
				<div
					className="bp3-tag bp3-intent-primary"
					style={css.tag}
					key={networkRole}
				>
					{networkRole}
				</div>,
			);
		}

		let orgName = '';
		if (!MiscUtils.objectIdNil(pod.organization)) {
			let org = OrganizationsStore.organization(pod.organization);
			orgName = org ? org.name : pod.organization;
		} else {
			orgName = 'Node Pod';
		}

		return <div
			className="bp3-card bp3-row"
			style={cardStyle}
			onClick={(evt): void => {
				let target = evt.target as HTMLElement;

				if (target.className.indexOf('open-ignore') !== -1) {
					return;
				}

				this.props.onOpen();
			}}
		>
			<div className="bp3-cell" style={css.name}>
				<div className="layout horizontal">
					<label
						className="bp3-control bp3-checkbox open-ignore"
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
						<span className="bp3-control-indicator open-ignore"/>
					</label>
					<div style={css.nameSpan}>
						{pod.name}
					</div>
				</div>
			</div>
			<div className="bp3-cell" style={css.item}>
				<span
					style={css.icon}
					className={'bp3-icon-standard bp3-text-muted ' + (
						pod.organization ? 'bp3-icon-people' : 'bp3-icon-layers')}
				/>
				{orgName}
			</div>
			<div className="flex bp3-cell" style={css.roles}>
				{role}
			</div>
		</div>;
	}
}

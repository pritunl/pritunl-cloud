/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as ReactRouter from 'react-router-dom';
import * as MiscUtils from '../utils/MiscUtils';
import * as NodeTypes from '../types/NodeTypes';

interface Props {
	node: NodeTypes.NodeRo;
	selected: boolean;
	onSelect: (shift: boolean) => void;
}

const css = {
	card: {
		display: 'table-row',
		width: '100%',
		padding: 0,
		boxShadow: 'none',
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
	type: {
		verticalAlign: 'top',
		display: 'table-cell',
		padding: '9px',
	} as React.CSSProperties,
	lastActivity: {
		verticalAlign: 'top',
		display: 'table-cell',
		padding: '9px',
		whiteSpace: 'nowrap',
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
	nameLink: {
		margin: '0 5px 0 0',
	} as React.CSSProperties,
};

export default class Node extends React.Component<Props, {}> {
	render(): JSX.Element {
		let node = this.props.node;
		let roles: JSX.Element[] = [];
		let active = node.requests_min !== 0 || node.memory !== 0 ||
			node.load1 !== 0 || node.load5 !== 0 || node.load15 !== 0;

		let cardStyle = {
			...css.card,
		};
		if (!active) {
			cardStyle.opacity = 0.6;
		}

		return <div
			className="pt-card pt-row"
			style={cardStyle}
		>
			<div className="pt-cell" style={css.name}>
				<div className="layout horizontal">
					<label className="pt-control pt-checkbox" style={css.select}>
						<input
							type="checkbox"
							checked={this.props.selected}
							onClick={(evt): void => {
								this.props.onSelect(evt.shiftKey);
							}}
						/>
						<span className="pt-control-indicator"/>
					</label>
					<ReactRouter.Link to={'/node/' + node.id} style={css.nameLink}>
						{node.name}
					</ReactRouter.Link>
				</div>
			</div>
			<div className="pt-cell" style={css.lastActivity}>
				{MiscUtils.formatDateShortTime(node.timestamp) || 'Inactive'}
			</div>
		</div>;
	}
}

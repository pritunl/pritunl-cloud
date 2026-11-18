/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Blueprint from '@blueprintjs/core';
import * as Theme from '../Theme';
import * as PodTypes from '../types/PodTypes';
import PodDiff from './PodDiff';

interface Props {
	open: boolean;
	disabled: boolean;
	changes: PodTypes.UnitChange[];
	onClose: () => void;
	onSave: () => void;
}

interface Stat {
	original?: string;
	modified?: string;
	additions: number;
	deletions: number;
}

interface State {
	stats: {[id: string]: Stat};
}

const statBlocks = 5;
const minusSign = String.fromCharCode(0x2212);

const css = {
	dialog: {
		width: '630px',
		position: 'absolute',
	} as React.CSSProperties,
	scroll: {
		maxHeight: 'calc(100vh - 175px)',
		margin: '1px 1px 0 1px',
		overflowY: 'auto',
		borderBottomStyle: 'solid',
	} as React.CSSProperties,
	footer: {
		marginTop: '15px',
	} as React.CSSProperties,
	panel: {
		padding: 0,
		boxShadow: 'none',
		borderStyle: 'solid',
	} as React.CSSProperties,
	panelDiff: {
		marginTop: '12px',
		padding: 0,
		boxShadow: 'none',
		borderStyle: 'solid',
	} as React.CSSProperties,
	header: {
		alignItems: 'center',
		padding: '7px 10px',
	} as React.CSSProperties,
	row: {
		alignItems: 'center',
		padding: '6px 10px',
		borderTopStyle: 'solid',
	} as React.CSSProperties,
	editor: {
		borderTopStyle: 'solid',
	} as React.CSSProperties,
	icon: {
		flex: 'none',
		marginRight: '7px',
	} as React.CSSProperties,
	title: {
		fontWeight: 'bold',
	} as React.CSSProperties,
	name: {
		minWidth: 0,
		overflow: 'hidden',
		textOverflow: 'ellipsis',
		whiteSpace: 'nowrap',
		fontFamily: Theme.monospaceFont,
		fontWeight: Theme.monospaceWeight,
		fontSize: '13px',
	} as React.CSSProperties,
	tag: {
		flex: 'none',
		marginLeft: '8px',
	} as React.CSSProperties,
	stat: {
		alignItems: 'baseline',
		flex: 'none',
		marginLeft: 'auto',
		paddingLeft: '10px',
	} as React.CSSProperties,
	statCount: {
		minWidth: '26px',
		marginLeft: '4px',
		textAlign: 'right',
		whiteSpace: 'nowrap',
		fontFamily: Theme.monospaceFont,
		fontWeight: Theme.monospaceWeight,
		fontSize: '12px',
	} as React.CSSProperties,
	statBlocks: {
		alignItems: 'center',
		marginLeft: '6px',
	} as React.CSSProperties,
	statBlock: {
		width: '8px',
		height: '8px',
		marginLeft: '2px',
		borderRadius: '1px',
		backgroundColor: 'currentColor',
	} as React.CSSProperties,
	statBlockNone: {
		width: '8px',
		height: '8px',
		marginLeft: '2px',
		borderRadius: '1px',
		backgroundColor: 'rgba(138, 155, 168, 0.3)',
	} as React.CSSProperties,
};

export default class PodSave extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			stats: {},
		};
	}

	onStat = (change: PodTypes.UnitChange, additions: number,
			deletions: number): void => {

		this.setState((state: State): State => {
			let stat = state.stats[change.id];
			if (stat && stat.original === change.original &&
				stat.modified === change.modified &&
				stat.additions === additions &&
				stat.deletions === deletions) {

				return null;
			}

			return {
				...state,
				stats: {
					...state.stats,
					[change.id]: {
						original: change.original,
						modified: change.modified,
						additions: additions,
						deletions: deletions,
					},
				},
			};
		});
	}

	getStat(change: PodTypes.UnitChange): Stat {
		let stat = this.state.stats[change.id];
		if (!stat || stat.original !== change.original ||
			stat.modified !== change.modified) {

			return null;
		}

		return stat;
	}

	renderStat(stat: Stat): JSX.Element {
		let additions = stat ? stat.additions : 0;
		let deletions = stat ? stat.deletions : 0;
		let total = additions + deletions;

		let addBlocks = additions;
		let delBlocks = deletions;
		if (total > statBlocks) {
			addBlocks = Math.round(additions / total * statBlocks);
			addBlocks = Math.max(addBlocks, additions ? 1 : 0);
			addBlocks = Math.min(addBlocks,
				deletions ? statBlocks - 1 : statBlocks);
			delBlocks = statBlocks - addBlocks;
		}

		let blocks: JSX.Element[] = [];
		for (let i = 0; i < statBlocks; i++) {
			if (i < addBlocks) {
				blocks.push(<span
					key={i}
					className="bp5-text-intent-success"
					style={css.statBlock}
				/>);
			} else if (i < addBlocks + delBlocks) {
				blocks.push(<span
					key={i}
					className="bp5-text-intent-danger"
					style={css.statBlock}
				/>);
			} else {
				blocks.push(<span key={i} style={css.statBlockNone}/>);
			}
		}

		let counts: JSX.Element;
		if (stat) {
			counts = <React.Fragment>
				<span
					className="bp5-text-intent-success"
					style={css.statCount}
				>+{additions}</span>
				<span
					className="bp5-text-intent-danger"
					style={css.statCount}
				>{minusSign}{deletions}</span>
			</React.Fragment>;
		}

		return <div className="layout horizontal" style={css.stat}>
			{counts}
			<div className="layout horizontal" style={css.statBlocks}>
				{blocks}
			</div>
		</div>;
	}

	render(): JSX.Element {
		let rows: JSX.Element[] = [];
		let diffs: JSX.Element[] = [];
		let total: Stat = {
			additions: 0,
			deletions: 0,
		};
		for (let change of (this.props.changes || [])) {
			let tagLabel = "modified";
			let tagClass = "bp5-tag bp5-minimal bp5-intent-primary";
			if (change.new) {
				tagLabel = "new";
				tagClass = "bp5-tag bp5-minimal bp5-intent-success";
			} else if (change.delete) {
				tagLabel = "deleted";
				tagClass = "bp5-tag bp5-minimal bp5-intent-danger";
			}

			let iconClass = "bp5-icon-standard bp5-text-muted bp5-icon-server";
			if (change.kind === "image") {
				iconClass = "bp5-icon-standard bp5-text-muted bp5-icon-compressed";
			}

			let stat = this.getStat(change);
			if (stat && total) {
				total.additions += stat.additions;
				total.deletions += stat.deletions;
			} else {
				total = null;
			}

			let label = <React.Fragment>
				<span className={iconClass} style={css.icon}/>
				<span style={css.name}>{change.name}</span>
				<span className={tagClass} style={css.tag}>{tagLabel}</span>
			</React.Fragment>;

			rows.push(<div
				key={change.id}
				className="layout horizontal bp5-border"
				style={css.row}
			>
				{label}
				{this.renderStat(stat)}
			</div>);

			diffs.push(<div
				key={change.id}
				className="bp5-card bp5-border"
				style={css.panelDiff}
			>
				<div
					className="layout horizontal bp5-card-header"
					style={css.header}
				>
					{label}
					{this.renderStat(stat)}
				</div>
				<div className="bp5-border" style={css.editor}>
					<PodDiff
						original={change.original}
						modified={change.modified}
						onStat={(additions: number, deletions: number): void => {
							this.onStat(change, additions, deletions);
						}}
					/>
				</div>
			</div>);
		}

		let body: JSX.Element;
		if (rows.length) {
			body = <div className="bp5-dialog-body">
				<div className="bp5-card bp5-border" style={css.panel}>
					<div
						className="layout horizontal bp5-card-header"
						style={css.header}
					>
						<span
							className="bp5-icon-standard bp5-text-muted bp5-icon-changes"
							style={css.icon}
						/>
						<span style={css.title}>Changed Units</span>
						<span
							className="bp5-tag bp5-minimal bp5-round"
							style={css.tag}
						>{rows.length}</span>
						{this.renderStat(total)}
					</div>
					{rows}
				</div>
				{diffs}
			</div>;
		} else {
			body = <div className="bp5-dialog-body">
				<Blueprint.Callout
					intent="primary"
				>No unit changes to save</Blueprint.Callout>
			</div>;
		}

		return <Blueprint.Dialog
			title="Save Pod Changes"
			style={css.dialog}
			isOpen={this.props.open}
			usePortal={true}
			portalContainer={document.body}
			onClose={this.props.onClose}
		>
			<div className="bp5-border" style={css.scroll}>
				{body}
			</div>
			<div className="bp5-dialog-footer" style={css.footer}>
				<div className="bp5-dialog-footer-actions">
					<button
						className="bp5-button"
						type="button"
						onClick={this.props.onClose}
					>Cancel</button>
					<button
						className="bp5-button bp5-intent-success bp5-icon-tick"
						type="button"
						disabled={this.props.disabled || !rows.length}
						onClick={this.props.onSave}
					>Save</button>
				</div>
			</div>
		</Blueprint.Dialog>;
	}
}

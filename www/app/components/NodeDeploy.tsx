/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Blueprint from '@blueprintjs/core';
import * as NodeTypes from '../types/NodeTypes';
import * as DatacenterTypes from "../types/DatacenterTypes";
import * as ZoneTypes from '../types/ZoneTypes';
import * as NodeActions from '../actions/NodeActions';
import * as BlockTypes from '../types/BlockTypes';
import * as MiscUtils from '../utils/MiscUtils';
import Help from './Help';
import PageInput from './PageInput';
import PageInputButton from './PageInputButton';
import PageSwitch from './PageSwitch';
import PageSelect from './PageSelect';

interface Props {
	hidden?: boolean;
	disabled?: boolean;
	node: NodeTypes.NodeRo;
	datacenters: DatacenterTypes.DatacentersRo;
	zones: ZoneTypes.ZonesRo;
	blocks: BlockTypes.BlocksRo;
}

interface State {
	popover: boolean;
}

const css = {
	box: {
	} as React.CSSProperties,
	button: {
		marginRight: '10px',
	} as React.CSSProperties,
	item: {
		margin: '9px 5px 0 5px',
		height: '20px',
	} as React.CSSProperties,
	callout: {
		marginBottom: '15px',
	} as React.CSSProperties,
	popover: {
		width: '230px',
	} as React.CSSProperties,
	popoverTarget: {
		top: '9px',
		left: '18px',
	} as React.CSSProperties,
	dialog: {
		maxWidth: '480px',
		margin: '30px 20px',
	} as React.CSSProperties,
	textarea: {
		width: '100%',
		resize: 'none',
		fontSize: '12px',
		fontFamily: '"Lucida Console", Monaco, monospace',
	} as React.CSSProperties,
};

export default class NodeDeploy extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			popover: false,
		};
	}

	render(): JSX.Element {
		let popoverElem: JSX.Element;

		if (this.state.popover) {
			let callout = 'Initialize Node';
			let errorMsg = '';
			let errorMsgElem: JSX.Element;

			if (errorMsg) {
				errorMsgElem = <div className="bp3-dialog-body">
					<div
						className="bp3-callout bp3-intent-danger bp3-icon-ban-circle"
						style={css.callout}
					>
						{errorMsg}
					</div>
				</div>;
			}

			popoverElem = <Blueprint.Dialog
				title="Initialize Node"
				style={css.dialog}
				isOpen={this.state.popover}
				usePortal={true}
				portalContainer={document.body}
				onClose={(): void => {
					this.setState({
						...this.state,
						popover: false,
					});
				}}
			>
				{errorMsgElem}
				<div className="bp3-dialog-body" hidden={!!errorMsgElem}>
					<div
						className="bp3-callout bp3-intent-primary bp3-icon-info-sign"
						style={css.callout}
					>
						{callout}
					</div>
				</div>
				<div className="bp3-dialog-footer">
					<div className="bp3-dialog-footer-actions">
						<button
							className="bp3-button"
							type="button"
							onClick={(): void => {
								this.setState({
									...this.state,
									popover: !this.state.popover,
								});
							}}
						>Close</button>
					</div>
				</div>
			</Blueprint.Dialog>;
		}

		return <div hidden={this.props.hidden} style={css.box}>
			<button
				className="bp3-button bp3-icon-cloud-upload bp3-intent-primary"
				style={css.button}
				type="button"
				disabled={this.props.disabled}
				onClick={(): void => {
					this.setState({
						...this.state,
						popover: !this.state.popover,
					});
				}}
			>
				Initialize Node
			</button>
			{popoverElem}
		</div>;
	}
}

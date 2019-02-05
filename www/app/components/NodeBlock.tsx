/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as NodeTypes from '../types/NodeTypes';
import * as BlockTypes from '../types/BlockTypes';

interface Props {
	interfaces?: string[];
	blocks: BlockTypes.BlocksRo;
	block: NodeTypes.BlockAttachment;
	onChange: (state: NodeTypes.BlockAttachment) => void;
	onAdd: () => void;
	onRemove: () => void;
}

const css = {
	group: {
		width: '100%',
		maxWidth: '310px',
		marginTop: '5px',
	} as React.CSSProperties,
	sourceGroup: {
		width: '100%',
		maxWidth: '219px',
		marginTop: '5px',
	} as React.CSSProperties,
	control: {
		marginTop: '5px',
	} as React.CSSProperties,
	protocol: {
		flex: '0 1 auto',
	} as React.CSSProperties,
	port: {
		width: '100%',
	} as React.CSSProperties,
	portBox: {
		flex: '1',
	} as React.CSSProperties,
	other: {
		flex: '0 1 auto',
		width: '52px',
		borderRadius: '0 3px 3px 0',
	} as React.CSSProperties,
};

export default class NodeBlock extends React.Component<Props, {}> {
	clone(): NodeTypes.BlockAttachment {
		return {
			...this.props.block,
		};
	}

	render(): JSX.Element {
		let block = this.props.block;

		let ifacesSelect: JSX.Element[] = [];
		for (let iface of (this.props.interfaces || [])) {
			ifacesSelect.push(
				<option key={iface} value={iface}>
					{iface}
				</option>,
			);
		}

		let blocksSelect: JSX.Element[] = [];
		for (let blck of (this.props.blocks || [])) {
			blocksSelect.push(
				<option key={blck.id} value={blck.id}>
					{blck.name}
				</option>,
			);
		}

		if (blocksSelect.length === 0) {
			blocksSelect.push(
				<option key="null" value="">No Blocks</option>);
		}

		return <div>
			<div className="bp3-control-group" style={css.group}>
				<div className="bp3-select" style={css.protocol}>
					<select
						value={block.interface}
						onChange={(evt): void => {
							let state = this.clone();
							state.interface = evt.target.value;
							this.props.onChange(state);
						}}
					>
						{ifacesSelect}
					</select>
				</div>
				<div className="bp3-select" style={css.protocol}>
					<select
						value={block.block}
						onChange={(evt): void => {
							let state = this.clone();
							state.block = evt.target.value;
							this.props.onChange(state);
						}}
					>
						{blocksSelect}
					</select>
				</div>
				<button
					className="bp3-button bp3-minimal bp3-intent-danger bp3-icon-remove"
					style={css.control}
					onClick={(): void => {
						this.props.onRemove();
					}}
				/>
				<button
					className="bp3-button bp3-minimal bp3-intent-success bp3-icon-add"
					style={css.control}
					onClick={(): void => {
						this.props.onAdd();
					}}
				/>
			</div>
		</div>;
	}
}

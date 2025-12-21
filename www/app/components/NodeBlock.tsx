/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as NodeTypes from '../types/NodeTypes';
import * as BlockTypes from '../types/BlockTypes';

interface Props {
	ipv6?: boolean;
	interfaces?: NodeTypes.Interface[];
	blocks: BlockTypes.BlocksRo;
	block: NodeTypes.BlockAttachment;
	onChange: (state: NodeTypes.BlockAttachment) => void;
	onAdd: () => void;
	onRemove: () => void;
}

const css = {
	group: {
		width: '100%',
		maxWidth: '400px',
		marginTop: '5px',
	} as React.CSSProperties,
	sourceGroup: {
		width: '100%',
		maxWidth: '219px',
		marginTop: '5px',
	} as React.CSSProperties,
	control: {
		marginTop: '0px',
	} as React.CSSProperties,
	protocol: {
		marginTop: '0px',
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

		let ifaceMatch = false;
		let ifacesSelect: JSX.Element[] = [];
		for (let iface of (this.props.interfaces || [])) {
			if (block.interface === iface.name) {
				ifaceMatch = true;
			}

			ifacesSelect.push(
				<option key={iface.name} value={iface.name}>
					{iface.name + (iface.address ? (` (${iface.address})`) : "")}
				</option>,
			);
		}

		if (!ifaceMatch) {
			ifacesSelect.push(
				<option key={block.interface} value={block.interface}>
					{block.interface}
				</option>,
			);
		}

		let blocksSelect: JSX.Element[] = [];
		for (let blck of (this.props.blocks || [])) {
			if (!this.props.ipv6 && blck.type === 'ipv6') {
				continue;
			} else if (this.props.ipv6 && blck.type !== 'ipv6') {
				continue;
			}

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
			<div className="bp5-control-group" style={css.group}>
				<div className="bp5-select" style={css.protocol}>
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
				<div className="bp5-select" style={css.protocol}>
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
					className="bp5-button bp5-minimal bp5-intent-danger bp5-icon-remove"
					style={css.control}
					onClick={(): void => {
						this.props.onRemove();
					}}
				/>
				<button
					className="bp5-button bp5-minimal bp5-intent-success bp5-icon-add"
					style={css.control}
					onClick={(): void => {
						this.props.onAdd();
					}}
				/>
			</div>
		</div>;
	}
}

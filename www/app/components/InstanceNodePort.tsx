/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as InstanceTypes from '../types/InstanceTypes';

interface Props {
	nodePort: InstanceTypes.NodePort;
	hidden?: boolean;
	onChange: (state: InstanceTypes.NodePort) => void;
	onRemove: () => void;
}

const css = {
	group: {
		width: '100%',
		maxWidth: '310px',
		marginTop: '5px',
	} as React.CSSProperties,
	protocol: {
		flex: '0 1 auto',
	} as React.CSSProperties,
	nodePort: {
		width: '100%',
		borderRadius: '0 3px 3px 0',
	} as React.CSSProperties,
	nodePortBox: {
		flex: '1',
	} as React.CSSProperties,
};

export default class InstanceNodePort extends React.Component<Props, {}> {
	clone(): InstanceTypes.NodePort {
		return {
			...this.props.nodePort,
		};
	}

	render(): JSX.Element {
		let nodePort = this.props.nodePort;

		return <div
			className="bp5-control-group"
			style={css.group}
			hidden={this.props.hidden}
		>
			<div className="bp5-select" style={css.protocol}>
				<select
					value={nodePort.protocol}
					onChange={(evt): void => {
						let state = this.clone();
						state.protocol = evt.target.value;
						this.props.onChange(state);
					}}
				>
					<option value="tcp">TCP</option>
					<option value="udp">UDP</option>
				</select>
			</div>
			<div style={css.nodePortBox}>
				<input
					className="bp5-input"
					style={css.nodePort}
					type="text"
					autoCapitalize="off"
					spellCheck={false}
					placeholder="External Port"
					value={nodePort.external_port || ''}
					onChange={(evt): void => {
						let state = this.clone();
						state.node_port = ""
						state.external_port = parseInt(evt.target.value, 10);
						this.props.onChange(state);
					}}
				/>
			</div>
			<div style={css.nodePortBox}>
				<input
					className="bp5-input"
					style={css.nodePort}
					type="text"
					autoCapitalize="off"
					spellCheck={false}
					placeholder="Internal Port"
					value={nodePort.internal_port || ''}
					onChange={(evt): void => {
						let state = this.clone();
						state.internal_port = parseInt(evt.target.value, 10);
						this.props.onChange(state);
					}}
				/>
			</div>
			<button
				className="bp5-button bp5-minimal bp5-intent-danger bp5-icon-remove"
				onClick={(): void => {
					this.props.onRemove();
				}}
			/>
		</div>;
	}
}

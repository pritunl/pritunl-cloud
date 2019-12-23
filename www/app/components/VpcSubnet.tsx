/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as VpcTypes from '../types/VpcTypes';

interface Props {
	disabled?: boolean;
	subnet: VpcTypes.Subnet;
	onChange?: (state: VpcTypes.Subnet) => void;
	onAdd: () => void;
	onRemove?: () => void;
}

const css = {
	group: {
		width: '100%',
		maxWidth: '310px',
		marginTop: '5px',
	} as React.CSSProperties,
	input: {
		width: '100%',
	} as React.CSSProperties,
	inputBox: {
		flex: '1',
	} as React.CSSProperties,
};

export default class VpcSubnet extends React.Component<Props, {}> {
	clone(): VpcTypes.Subnet {
		return {
			...this.props.subnet,
		};
	}

	render(): JSX.Element {
		let subnet = this.props.subnet;

		return <div>
			<div className="bp3-control-group" style={css.group}>
				<div style={css.inputBox}>
					<input
						className="bp3-input"
						style={css.input}
						disabled={this.props.disabled}
						type="text"
						autoCapitalize="off"
						spellCheck={false}
						placeholder="Name"
						value={subnet.name || ''}
						onChange={(evt): void => {
							let state = this.clone();
							state.name = evt.target.value;
							this.props.onChange(state);
						}}
					/>
				</div>
				<div style={css.inputBox}>
					<input
						className="bp3-input"
						style={css.input}
						disabled={this.props.disabled}
						type="text"
						autoCapitalize="off"
						spellCheck={false}
						placeholder="Network"
						value={subnet.network || ''}
						onChange={(evt): void => {
							let state = this.clone();
							state.network = evt.target.value;
							this.props.onChange(state);
						}}
					/>
				</div>
				<button
					className="bp3-button bp3-minimal bp3-intent-danger bp3-icon-remove"
					disabled={this.props.disabled}
					onClick={(): void => {
						this.props.onRemove();
					}}
				/>
				<button
					className="bp3-button bp3-minimal bp3-intent-success bp3-icon-add"
					onClick={(): void => {
						this.props.onAdd();
					}}
				/>
			</div>
		</div>;
	}
}

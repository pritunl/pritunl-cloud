/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Constants from '../Constants';
import * as VpcTypes from '../types/VpcTypes';

interface Props {
	disabled?: boolean;
	arp: VpcTypes.Arp;
	onChange?: (state: VpcTypes.Arp) => void;
	onAdd: (prepend: boolean) => void;
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

export default class VpcArp extends React.Component<Props, {}> {
	clone(): VpcTypes.Arp {
		return {
			...this.props.arp,
		};
	}

	render(): JSX.Element {
		let arp = this.props.arp;

		return <div>
			<div className="bp5-control-group" style={css.group}>
				<div style={css.inputBox}>
					<input
						className="bp5-input"
						style={css.input}
						disabled={this.props.disabled} // Constants.user
						type="text"
						autoCapitalize="off"
						spellCheck={false}
						placeholder="IP Address"
						value={arp.ip || ''}
						onChange={(evt): void => {
							let state = this.clone();
							state.ip = evt.target.value;
							this.props.onChange(state);
						}}
					/>
				</div>
				<div style={css.inputBox}>
					<input
						className="bp5-input"
						style={css.input}
						disabled={this.props.disabled}
						type="text"
						autoCapitalize="off"
						spellCheck={false}
						placeholder="Mac Address"
						value={arp.mac || ''}
						onChange={(evt): void => {
							let state = this.clone();
							state.mac = evt.target.value;
							this.props.onChange(state);
						}}
					/>
				</div>
				<button
					className="bp5-button bp5-minimal bp5-intent-danger bp5-icon-remove"
					disabled={this.props.disabled}
					onClick={(): void => {
						this.props.onRemove();
					}}
				/>
				<button
					className="bp5-button bp5-minimal bp5-intent-success bp5-icon-add"
					onClick={(evt): void => {
						this.props.onAdd(evt.shiftKey);
					}}
				/>
			</div>
		</div>;
	}
}

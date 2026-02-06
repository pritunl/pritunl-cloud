/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as VpcTypes from '../types/VpcTypes';

interface Props {
	disabled?: boolean;
	map: VpcTypes.Map;
	onChange?: (state: VpcTypes.Map) => void;
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

export default class VpcMap extends React.Component<Props, {}> {
	clone(): VpcTypes.Map {
		return {
			...this.props.map,
		};
	}

	render(): JSX.Element {
		let map = this.props.map;

		return <div>
			<div className="bp5-control-group" style={css.group}>
				<div style={css.inputBox}>
					<input
						className="bp5-input"
						style={css.input}
						disabled={this.props.disabled}
						type="text"
						autoCapitalize="off"
						spellCheck={false}
						placeholder="Destination"
						value={map.destination || ''}
						onChange={(evt): void => {
							let state = this.clone();
							state.destination = evt.target.value;
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
						placeholder="Target"
						value={map.target || ''}
						onChange={(evt): void => {
							let state = this.clone();
							state.target = evt.target.value;
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

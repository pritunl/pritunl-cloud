/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as VpcTypes from '../types/VpcTypes';

interface Props {
	disabled?: boolean;
	route: VpcTypes.Route;
	onChange: (state: VpcTypes.Route) => void;
	onAdd: () => void;
	onRemove: () => void;
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

export default class VpcRoute extends React.Component<Props, {}> {
	clone(): VpcTypes.Route {
		return {
			...this.props.route,
		};
	}

	render(): JSX.Element {
		let route = this.props.route;

		return <div>
			<div className="pt-control-group" style={css.group}>
				<div style={css.inputBox}>
					<input
						className="pt-input"
						style={css.input}
						disabled={this.props.disabled}
						type="text"
						autoCapitalize="off"
						spellCheck={false}
						placeholder="Destination"
						value={route.destination || ''}
						onChange={(evt): void => {
							let state = this.clone();
							state.destination = evt.target.value;
							this.props.onChange(state);
						}}
					/>
				</div>
				<div style={css.inputBox}>
					<input
						className="pt-input"
						style={css.input}
						disabled={this.props.disabled}
						type="text"
						autoCapitalize="off"
						spellCheck={false}
						placeholder="Target"
						value={route.target || ''}
						onChange={(evt): void => {
							let state = this.clone();
							state.target = evt.target.value;
							this.props.onChange(state);
						}}
					/>
				</div>
				<button
					className="pt-button pt-minimal pt-intent-danger pt-icon-remove"
					disabled={this.props.disabled}
					onClick={(): void => {
						this.props.onRemove();
					}}
				/>
				<button
					className="pt-button pt-minimal pt-intent-success pt-icon-add"
					onClick={(): void => {
						this.props.onAdd();
					}}
				/>
			</div>
		</div>;
	}
}

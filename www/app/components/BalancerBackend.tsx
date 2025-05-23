/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as BalancerTypes from '../types/BalancerTypes';

interface Props {
	backend: BalancerTypes.Backend;
	onChange: (state: BalancerTypes.Backend) => void;
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
	hostname: {
		width: '100%',
	} as React.CSSProperties,
	hostnameBox: {
		flex: '1',
	} as React.CSSProperties,
	port: {
		flex: '0 1 auto',
		width: '52px',
		borderRadius: '0 3px 3px 0',
	} as React.CSSProperties,
};

export default class BalancerBackend extends React.Component<Props, {}> {
	clone(): BalancerTypes.Backend {
		return {
			...this.props.backend,
		};
	}

	render(): JSX.Element {
		let backend = this.props.backend;

		return <div className="bp5-control-group" style={css.group}>
			<div className="bp5-select" style={css.protocol}>
				<select
					value={backend.protocol}
					onChange={(evt): void => {
						let state = this.clone();
						state.protocol = evt.target.value;
						this.props.onChange(state);
					}}
				>
					<option value="http">HTTP</option>
					<option value="https">HTTPS</option>
				</select>
			</div>
			<div style={css.hostnameBox}>
				<input
					className="bp5-input"
					style={css.hostname}
					type="text"
					autoCapitalize="off"
					spellCheck={false}
					placeholder="Hostname"
					value={backend.hostname || ''}
					onChange={(evt): void => {
						let state = this.clone();
						state.hostname = evt.target.value;
						this.props.onChange(state);
					}}
				/>
			</div>
			<input
				className="bp5-input"
				style={css.port}
				type="text"
				autoCapitalize="off"
				spellCheck={false}
				placeholder="Port"
				value={backend.port}
				onChange={(evt): void => {
					let state = this.clone();
					state.port = parseInt(evt.target.value, 10);
					this.props.onChange(state);
				}}
			/>
			<button
				className="bp5-button bp5-minimal bp5-intent-danger bp5-icon-remove"
				onClick={(): void => {
					this.props.onRemove();
				}}
			/>
		</div>;
	}
}

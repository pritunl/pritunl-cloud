/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as InstanceTypes from '../types/InstanceTypes';
import PageInputButton from './PageInputButton';

interface Props {
	disabled: boolean;
	mount: InstanceTypes.Mount;
	onChange: (state: InstanceTypes.Mount) => void;
	onAdd: () => void;
	onRemove: () => void;
}

interface State {
	addRole: string;
}

const css = {
	groupTop: {
		width: '100%',
		maxWidth: '278px',
		marginTop: '5px',
		marginBottom: '0',
	} as React.CSSProperties,
	groupBottom: {
		width: '100%',
		maxWidth: '278px',
		marginTop: '2px',
		marginBottom: '15px',
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
	input: {
		width: '100%',
	} as React.CSSProperties,
	nameInput: {
		width: '35%',
	} as React.CSSProperties,
	pathInput: {
		width: '64%',
	} as React.CSSProperties,
	inputBox: {
		flex: '1',
	} as React.CSSProperties,
	role: {
		margin: '9px 5px 0 5px',
		minHeight: '20px',
	} as React.CSSProperties,
	roles: {
		width: '100%',
		maxWidth: '400px',
		marginTop: '5px',
		marginBottom: '15px',
	} as React.CSSProperties,
	other: {
		flex: '0 1 auto',
		width: '52px',
		borderRadius: '0 3px 3px 0',
	} as React.CSSProperties,
};

export default class InstanceMount extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			addRole: null,
		};
	}

	clone(): InstanceTypes.Mount {
		return {
			...this.props.mount,
		};
	}

	render(): JSX.Element {
		let mount = this.props.mount;

		return <div>
			<div className="bp5-control-group" style={css.groupTop}>
				<input
					className="bp5-input"
					style={css.input}
					disabled={this.props.disabled}
					type="text"
					autoCapitalize="off"
					spellCheck={false}
					placeholder="Enter host path"
					value={mount.path || ''}
					onChange={(evt): void => {
						let state = this.clone();
						state.path = evt.target.value;
						this.props.onChange(state);
					}}
				/>
			</div>
			<div className="bp5-control-group" style={css.groupBottom}>
				<input
					className="bp5-input"
					style={css.nameInput}
					disabled={this.props.disabled}
					type="text"
					autoCapitalize="off"
					spellCheck={false}
					placeholder="Enter name"
					value={mount.name || ''}
					onChange={(evt): void => {
						let state = this.clone();
						state.name = evt.target.value;
						this.props.onChange(state);
					}}
				/>
				<input
					className="bp5-input"
					style={css.pathInput}
					disabled={this.props.disabled}
					type="text"
					autoCapitalize="off"
					spellCheck={false}
					placeholder="Enter instance path"
					value={mount.host_path || ''}
					onChange={(evt): void => {
						let state = this.clone();
						state.host_path = evt.target.value;
						this.props.onChange(state);
					}}
				/>
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

/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as InstanceTypes from '../types/InstanceTypes';

interface Props {
	disabled?: boolean;
	iscsi: InstanceTypes.IscsiDevice;
	onChange?: (state: InstanceTypes.IscsiDevice) => void;
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

export default class InstanceIscsiDevice extends React.Component<Props, {}> {
	clone(): InstanceTypes.IscsiDevice {
		return {
			...this.props.iscsi,
		};
	}

	render(): JSX.Element {
		let iscsi = this.props.iscsi;

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
						placeholder="iSCSI URI"
						value={iscsi.uri || ''}
						onChange={(evt): void => {
							let state = this.clone();
							state.uri = evt.target.value;
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
					onClick={(): void => {
						this.props.onAdd();
					}}
				/>
			</div>
		</div>;
	}
}

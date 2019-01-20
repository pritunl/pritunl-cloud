/// <reference path="../References.d.ts"/>
import * as React from 'react';

interface Props {
	disabled?: boolean;
	linkUri: string;
	onChange?: (state: string) => void;
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

export default class VpcUriLink extends React.Component<Props, {}> {
	render(): JSX.Element {
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
						placeholder="Link URI"
						value={this.props.linkUri || ''}
						onChange={(evt): void => {
							this.props.onChange(evt.target.value);
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

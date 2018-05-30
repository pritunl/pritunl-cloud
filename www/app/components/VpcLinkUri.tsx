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
			<div className="pt-control-group" style={css.group}>
				<div style={css.inputBox}>
					<input
						className="pt-input"
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

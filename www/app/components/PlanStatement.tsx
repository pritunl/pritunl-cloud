/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as PlanTypes from '../types/PlanTypes';
import * as Theme from "../Theme";
import * as MonacoEditor from "@monaco-editor/react"

interface Props {
	disabled?: boolean;
	statement: PlanTypes.Statement;
	onChange: (statement: PlanTypes.Statement) => void;
	onRemove: () => void;
}

const css = {
	group: {
		width: '100%',
		maxWidth: '310px',
		marginTop: '5px',
	} as React.CSSProperties,
	textarea: {
		width: '100%',
		resize: 'none',
		fontSize: Theme.monospaceSize,
		fontFamily: Theme.monospaceFont,
		fontWeight: Theme.monospaceWeight,
	} as React.CSSProperties,
};

export default class PlanStatement extends React.Component<Props, {}> {
	clone(): PlanTypes.Statement {
		return {
			...this.props.statement,
		};
	}

	render(): JSX.Element {
		let statement = this.props.statement;

		return <div className="bp5-control-group layout horizontal" style={css.group}>
			<textarea
				className="bp5-input"
				style={css.textarea}
				disabled={this.props.disabled}
				autoCapitalize="off"
				spellCheck={false}
				rows={3}
				value={statement.statement || ''}
				onChange={(evt): void => {
					let state = this.clone();
					state.statement = evt.target.value;
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

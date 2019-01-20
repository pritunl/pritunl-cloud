/// <reference path="../References.d.ts"/>
import * as React from 'react';
import ConfirmButton from './ConfirmButton';

interface Props {
	hidden?: boolean;
	label: string;
	value: string;
	confirmMsg?: string;
	disabled?: boolean;
	buttonClass?: string;
	progressClassName?: string;
	onChange: (val: string) => void;
	onSubmit: () => void;
}

const css = {
	group: {
		marginBottom: '15px',
		width: '100%',
		maxWidth: '280px',
	} as React.CSSProperties,
	select: {
		width: '100%',
		borderTopLeftRadius: '3px',
		borderBottomLeftRadius: '3px',
	} as React.CSSProperties,
	selectInner: {
		width: '100%',
	} as React.CSSProperties,
	selectBox: {
		flex: '1',
	} as React.CSSProperties,
};

export default class PageSelectButton extends React.Component<Props, {}> {
	render(): JSX.Element {
		let buttonClass = 'bp3-button';
		if (this.props.buttonClass) {
			buttonClass += ' ' + this.props.buttonClass;
		}

		return <div
			className="bp3-control-group"
			style={css.group}
			hidden={this.props.hidden}
		>
			<div style={css.selectBox}>
				<div className="bp3-select" style={css.select}>
					<select
						style={css.selectInner}
						disabled={this.props.disabled}
						value={this.props.value || ''}
						onChange={(evt): void => {
							this.props.onChange(evt.target.value);
						}}
					>
						{this.props.children}
					</select>
				</div>
			</div>
			<ConfirmButton
				label={this.props.label}
				className={buttonClass}
				progressClassName={this.props.progressClassName}
				confirmMsg={this.props.confirmMsg}
				disabled={this.props.disabled}
				onConfirm={this.props.onSubmit}
			/>
		</div>;
	}
}

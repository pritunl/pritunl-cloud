/// <reference path="../References.d.ts"/>
import * as React from 'react';

interface Props {
	children?: React.ReactNode
	style?: React.CSSProperties;
	message: string;
	changed: boolean;
	disabled: boolean;
	wrap?: boolean;
	hidden?: boolean;
	light?: boolean;
	onCancel: () => void;
	onSave: () => void;
}

const css = {
	message: {
		marginTop: '6px',
	} as React.CSSProperties,
	messageWrap: {
		marginTop: '6px',
		marginRight: '10px',
	} as React.CSSProperties,
	box: {
		marginTop: '15px',
	} as React.CSSProperties,
	button: {
		marginLeft: '10px',
	} as React.CSSProperties,
	buttonWrap: {
		marginLeft: '10px',
		marginBottom: '10px',
	} as React.CSSProperties,
	buttonWrapFirst: {
		marginBottom: '10px',
	} as React.CSSProperties,
	buttons: {
		flexShrink: 0,
	} as React.CSSProperties,
};

export default class PageSave extends React.Component<Props, {}> {
	render(): JSX.Element {
		let style: React.CSSProperties = this.props.light ? null : css.box;

		if (this.props.style) {
			style = {
				...style,
				...this.props.style,
			};
		}

		let containerClass = 'layout horizontal';
		let buttonStyle: React.CSSProperties;
		let buttonStyleFirst: React.CSSProperties;
		let messageStyle: React.CSSProperties;
		if (this.props.wrap) {
			buttonStyle = css.buttonWrap;
			buttonStyleFirst = css.buttonWrapFirst;
			messageStyle = css.messageWrap;
		} else {
			buttonStyle = css.button;
			buttonStyleFirst = css.button;
			messageStyle = css.message;
		}

		return <div
			className={'layout horizontal' + (this.props.wrap ? ' wrap': '')}
			style={style}
			hidden={this.props.hidden && !this.props.children}
		>
			{this.props.children}
			<div className="flex"/>
			<div className="layout horizontal">
				<span style={messageStyle} hidden={!this.props.message}>
					{this.props.message}
				</span>
				<div style={css.buttons}>
					<button
						className="bp5-button bp5-icon-cross"
						style={buttonStyleFirst}
						hidden={this.props.hidden}
						type="button"
						disabled={!this.props.changed || this.props.disabled}
						onClick={this.props.onCancel}
					>
						Cancel
					</button>
					<button
						className="bp5-button bp5-intent-success bp5-icon-tick"
						style={buttonStyle}
						hidden={this.props.hidden}
						type="button"
						disabled={!this.props.changed || this.props.disabled}
						onClick={this.props.onSave}
					>
						Save
					</button>
				</div>
			</div>
		</div>;
	}
}

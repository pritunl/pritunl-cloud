/// <reference path="../References.d.ts"/>
import * as React from 'react';
import Help from './Help';

interface Props {
	children?: React.ReactNode
	className?: string;
	hidden?: boolean;
	disabled?: boolean;
	label: string;
	help: string;
	value: string;
	onClick: () => void;
}

const css = {
	label: {
		display: 'inline-block',
	} as React.CSSProperties,
};

export default class PageButton extends React.Component<Props, {}> {
	render(): JSX.Element {
		return <div hidden={this.props.hidden}>
			<label className="bp5-label" style={css.label}>
				{this.props.label}
				<Help
					title={this.props.label}
					content={this.props.help}
				/>
				<div className="bp5-select">
					<button
						className={"bp5-button " + this.props.className}
						disabled={this.props.disabled}
						onClick={this.props.onClick}
					>
						{this.props.children}
					</button>
				</div>
			</label>
		</div>;
	}
}

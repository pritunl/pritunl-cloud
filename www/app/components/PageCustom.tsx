/// <reference path="../References.d.ts"/>
import * as React from 'react';
import Help from './Help';

interface Props {
	children?: React.ReactNode
	hidden?: boolean;
	label: string;
	help: string | JSX.Element;
}

const css = {
	label: {
		display: 'inline-block',
	} as React.CSSProperties,
};

export default class PageCustom extends React.Component<Props, {}> {
	render(): JSX.Element {
		return <div hidden={this.props.hidden}>
			<label className="bp5-label" style={css.label}>
				{this.props.label}
				<Help
					title={this.props.label}
					content={this.props.help}
				/>
				{this.props.children}
			</label>
		</div>;
	}
}

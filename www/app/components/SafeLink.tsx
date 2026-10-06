/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Blueprint from '@blueprintjs/core';

interface Props {
	href: string;
	style?: React.CSSProperties;
	className?: string;
	title?: string;
	confirmMsg?: string;
	children?: React.ReactNode;
}

interface State {
	dialog: boolean;
}

const css = {
	dialog: {
		scrollbarGutter: "stable",
		width: '420px',
		position: 'absolute',
	} as React.CSSProperties,
	message: {
		marginBottom: '10px',
	} as React.CSSProperties,
	url: {
		fontFamily: 'monospace',
		fontSize: '12px',
		wordBreak: 'break-all',
		padding: '8px 10px',
		borderRadius: '3px',
		background: 'rgba(138, 155, 168, 0.12)',
		border: '1px solid rgba(138, 155, 168, 0.25)',
		maxHeight: '160px',
		overflowY: 'auto',
	} as React.CSSProperties,
	host: {
		fontWeight: 600,
	} as React.CSSProperties,
	warning: {
		marginTop: '12px',
		fontSize: '12px',
	} as React.CSSProperties,
	link: {
		cursor: 'pointer',
		userSelect: 'none',
	} as React.CSSProperties,
};

function parseUrl(href: string): URL {
	try {
		let url = new URL(href);
		if (url.protocol !== 'http:' && url.protocol !== 'https:') {
			return null;
		}
		return url;
	} catch (e) {
		return null;
	}
}

export default class SafeLink extends React.Component<Props, State> {
	constructor(props: Props, context: any) {
		super(props, context);
		this.state = {
			dialog: false,
		};
	}

	openDialog = (evt?: React.SyntheticEvent<HTMLElement>): void => {
		if (evt) {
			evt.preventDefault();
			evt.stopPropagation();
		}
		this.setState({
			...this.state,
			dialog: true,
		});
	}

	onKeyDown = (evt: React.KeyboardEvent<HTMLElement>): void => {
		if (evt.key === 'Enter' || evt.key === ' ') {
			this.openDialog(evt);
		}
	}

	closeDialog = (): void => {
		this.setState({
			...this.state,
			dialog: false,
		});
	}

	closeDialogConfirm = (): void => {
		let url = parseUrl(this.props.href);

		this.setState({
			...this.state,
			dialog: false,
		});

		if (!url) {
			return;
		}

		window.open(url.toString(), '_blank', 'noopener,noreferrer');
	}

	render(): JSX.Element {
		let url = parseUrl(this.props.href);

		if (!url) {
			return <span
				className={this.props.className}
				style={this.props.style}
				title={this.props.title || this.props.href}
			>{this.props.children || this.props.href}</span>;
		}

		let host = url.hostname.replace(/^www\./, '');
		let confirmMsg = this.props.confirmMsg || 'This link was provided ' +
			'by an external source and has not been verified. Open it in ' +
			'a new tab?';

		let dialogElem: JSX.Element;
		if (this.state.dialog) {
			dialogElem = <Blueprint.Dialog
				title="Open External Link"
				style={css.dialog}
				isOpen={this.state.dialog}
				usePortal={true}
				portalContainer={document.body}
				onClose={this.closeDialog}
			>
				<div className="bp5-dialog-body">
					<div style={css.message}>{confirmMsg}</div>
					<div style={css.url}>
						<span style={css.host}>{host}</span>
						{url.toString().slice(url.origin.length)}
					</div>
					<Blueprint.Callout
						intent={Blueprint.Intent.DANGER}
						icon="warning-sign"
						compact={true}
						style={css.warning}
					>
						Only continue if you recognize and trust this site.
					</Blueprint.Callout>
				</div>
				<div className="bp5-dialog-footer">
					<div className="bp5-dialog-footer-actions">
						<button
							className="bp5-button"
							type="button"
							onClick={this.closeDialog}
						>Cancel</button>
						<button
							className="bp5-button bp5-intent-danger bp5-icon-share"
							type="button"
							onClick={this.closeDialogConfirm}
						>Open External Link</button>
					</div>
				</div>
			</Blueprint.Dialog>;
		}

		let className = 'bp5-text-intent-primary';
		if (this.props.className) {
			className += ' ' + this.props.className;
		}

		return <React.Fragment>
			<span
				className={className}
				style={{
					...css.link,
					...this.props.style,
				}}
				title={this.props.title || url.toString()}
				role="button"
				tabIndex={0}
				onClick={this.openDialog}
				onAuxClick={this.openDialog}
				onKeyDown={this.onKeyDown}
			>{this.props.children || host}</span>
			{dialogElem}
		</React.Fragment>;
	}
}

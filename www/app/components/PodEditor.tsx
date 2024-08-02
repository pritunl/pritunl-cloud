/// <reference path="../References.d.ts"/>
import * as React from 'react';
import Help from "./Help";
import * as Theme from "../Theme";

import Markdown from 'react-markdown';
import hljs from "highlight.js/lib/core";

import * as MonacoEditor from "@monaco-editor/react"

interface Props {
	defaultEdit?: boolean;
	disabled?: boolean;
	value: string;
	onChange?: (value: string) => void;
}

interface State {
	expandLeft: boolean;
	expandRight: boolean;
}

const css = {
	group: {
		position: 'relative',
		flex: 1,
		minWidth: '280px',
		margin: '0 10px',
	} as React.CSSProperties,
	groupSpaced: {
		position: 'relative',
		flex: 1,
		minWidth: '280px',
		margin: '0 10px',
		padding: '8px 0 0 0 ',
	} as React.CSSProperties,
	groupSpacedExt: {
		position: 'relative',
		flex: 1,
		minWidth: '280px',
		margin: '0 10px',
		padding: '30px 0 0 0 ',
	} as React.CSSProperties,
	editorBox: {
		margin: '10px 0',
	} as React.CSSProperties,
	editor: {
		margin: '10px 0',
	} as React.CSSProperties,
	buttonEdit: {
		position: 'absolute',
		top: '2px',
		right: '0px',
		padding: '7px',
	} as React.CSSProperties,
	buttonLeft: {
		position: 'absolute',
		top: '-4px',
		right: '0px',
		padding: '7px',
	} as React.CSSProperties,
	buttonRight: {
		position: 'absolute',
		top: '-4px',
		right: '0px',
		padding: '7px',
	} as React.CSSProperties,
};

export default class PodEditor extends React.Component<Props, State> {
	markdown: React.RefObject<HTMLDivElement>;

	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			expandLeft: null,
			expandRight: null,
		}

		this.markdown = React.createRef();
	}

	componentDidMount(): void {
		hljs.highlightAll();

		if (this.markdown.current) {
			const bashElements = this.markdown.current.querySelectorAll('[class^="language-"]:not(.hljs)')
			Array.from(bashElements).forEach(element => {
				element.classList.add('hljs');
			});
		}
	}

	componentDidUpdate(): void {
		hljs.highlightAll();

		if (this.markdown.current) {
			const bashElements = this.markdown.current.querySelectorAll('code:not(.hljs)')
			Array.from(bashElements).forEach(element => {
				element.classList.add('hljs');
			});
		}
	}

	render(): JSX.Element {
		let expandLeft = this.state.expandLeft
		let expandRight = this.state.expandRight
		let markdown: JSX.Element
		let markdownButton: JSX.Element
		let leftGroupStyle: React.CSSProperties = css.group
		let expandIconClass: string

		if (expandLeft === null && expandRight === null) {
			if (this.props.defaultEdit) {
				expandLeft = false
				expandRight = true
			} else {
				expandLeft = true
				expandRight = false
			}
		}

		if (!expandLeft && !expandRight) {
			expandIconClass = "bp5-button bp5-large bp5-minimal bp5-icon-maximize"
		} else {
			expandIconClass = "bp5-button bp5-large bp5-minimal bp5-icon-minimize"
		}

		if (!expandRight) {
			markdown = <Markdown>{this.props.value}</Markdown>

			if (expandLeft) {
				markdownButton = <button
					disabled={this.props.disabled}
					className="bp5-button bp5-icon-edit"
					style={css.buttonEdit}
					onClick={(): void => {
						this.setState({
							...this.state,
							expandLeft: false,
							expandRight: true,
						})
					}}
				>Edit Spec</button>
			} else {
				markdownButton = <button
					className={expandIconClass}
					style={css.buttonRight}
					onClick={(): void => {
						this.setState({
							...this.state,
							expandLeft: !expandLeft,
							expandRight: false,
						})
					}}
				/>
			}
		}

		let valTrim = (this.props.value || "").trimStart()
		if (!valTrim || (valTrim.length > 0 && valTrim.charAt(0) === "`")) {
			leftGroupStyle = css.groupSpacedExt
		} else if (valTrim && valTrim.length > 0 && valTrim.charAt(0) !== "#") {
			leftGroupStyle = css.groupSpaced
		} else {
			let valFirst = valTrim.split('\n')[0] || ""
			valFirst = valFirst.replace(/#/g, "").trim()
			if (!valFirst) {
				leftGroupStyle = css.groupSpacedExt
			}
		}

		return <div className="layout horizontal flex" style={css.editorBox}>
			<div
				ref={this.markdown}
				style={leftGroupStyle}
				hidden={expandRight}
			>
				{markdownButton}
				{markdown}
			</div>
			<div style={css.group} hidden={expandLeft}>
				<button
					className={expandIconClass}
					style={css.buttonRight}
					onClick={(): void => {
						this.setState({
							...this.state,
							expandLeft: false,
							expandRight: !expandRight,
						})
					}}
				/>
				<label
					className="bp5-label flex"
					style={css.editorBox}
				>
					Pod Spec
					<Help
						title="Spec"
						content="Spec file for pod."
					/>
					<div style={css.editor}>
						<MonacoEditor.Editor
							height="800px"
							width="100%"
							defaultLanguage="markdown"
							theme={Theme.editorTheme()}
							defaultValue={this.props.value}
							options={{
								folding: false,
								fontSize: 14,
								fontFamily: "'DejaVu Sans Mono', Monaco, Menlo, 'Ubuntu Mono', Consolas, source-code-pro, monospace",
								tabSize: 4,
								detectIndentation: false,
								rulers: [80],
								scrollBeyondLastLine: false,
								minimap: {
									enabled: expandRight,
								},
								wordWrap: "on",
							}}
							onChange={(val): void => {
								this.props.onChange(val)
							}}
						/>
					</div>
				</label>
			</div>
		</div>;
	}
}

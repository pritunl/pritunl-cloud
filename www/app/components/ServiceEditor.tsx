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
		padding: '26px 0 0 0 ',
	} as React.CSSProperties,
	editorBox: {
		margin: '10px 0',
	} as React.CSSProperties,
	editor: {
		margin: '11px 0 10px 0',
		borderRadius: '3px',
		overflow: 'hidden',
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

const hashRe = /^( {0,3})#+\s+\S+/
const blockRe = /^( {4}|\s*`)/
const langRe = /^language-(.+)$/

export default class ServiceEditor extends React.Component<Props, State> {
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
		if (this.markdown.current) {
			const codeElems = this.markdown.current.querySelectorAll('pre code')

			Array.from(codeElems).forEach((element: HTMLElement) => {
				if (!element.dataset.highlighted) {
					hljs.highlightElement(element)
				}
			})
		}
	}

	componentDidUpdate(): void {
		if (this.markdown.current) {
			const codeElems = this.markdown.current.querySelectorAll('pre code')

			Array.from(codeElems).forEach((element: HTMLElement) => {
				if (!element.dataset.highlighted) {
					hljs.highlightElement(element)
				}
			})
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
			markdown = <Markdown
				children={this.props.value}
				components={{
					code(props) {
						let {children, className, node, ...rest} = props
						let match = (className || "").match(langRe)

						if (match && !hljs.getLanguage(match[1])) {
							className = "language-plaintext"
						}

						return <code {...rest} className={className}>
							{children}
						</code>
					}
				}}
			/>

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

		let val = (this.props.value || "")
		let valTrim = val.trimStart()

		if (blockRe.test(val)) {
			leftGroupStyle = css.groupSpacedExt
		} else if (!hashRe.test(val)) {
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
					Service Spec
					<Help
						title="Spec"
						content="Spec file for service."
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

/// <reference path="../References.d.ts"/>
import * as React from 'react';
import Help from "./Help";
import * as Theme from "../Theme";
import * as CompletionEngine from "../completion/Engine"

import Markdown from 'react-markdown';
import hljs from "highlight.js/lib/core";

import * as MonacoEditor from "@monaco-editor/react"
import * as Monaco from "monaco-editor";

interface Props {
	readOnly: boolean;
	expandLeft: boolean;
	expandRight: boolean;
	disabled?: boolean;
	uuid: string;
	value: string;
	onEdit?: () => void;
	onChange?: (value: string) => void;
}

interface State {
}

interface EditorState {
	model: Monaco.editor.ITextModel
	view: Monaco.editor.ICodeEditorViewState
}

const css = {
	group: {
		position: 'relative',
		flex: 1,
		minWidth: '280px',
		margin: '0',
		fontSize: '12px',
	} as React.CSSProperties,
	groupSpaced: {
		position: 'relative',
		flex: 1,
		minWidth: '280px',
		margin: '0',
		padding: '8px 0 0 0 ',
		fontSize: '12px',
	} as React.CSSProperties,
	groupSpacedExt: {
		position: 'relative',
		flex: 1,
		minWidth: '280px',
		margin: '0',
		padding: '0 0 0 0 ',
		fontSize: '12px',
	} as React.CSSProperties,
	groupSplit: {
		position: 'relative',
		flex: 1,
		minWidth: '280px',
		margin: '0 0 0 10px',
	} as React.CSSProperties,
	editorBox: {
		margin: '0 0 10px 0',
	} as React.CSSProperties,
	editor: {
		margin: '0 0 10px 0',
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
	markdown: React.RefObject<HTMLDivElement>
	curUuid: string
	editor: Monaco.editor.IStandaloneCodeEditor
	monaco: MonacoEditor.Monaco
	states: Record<string, EditorState>

	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
		}

		this.markdown = React.createRef();
		this.states = {}
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
		Theme.addChangeListener(this.onThemeChange);
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

	componentWillUnmount(): void {
		Theme.removeChangeListener(this.onThemeChange);
		this.editor = undefined
		this.monaco = undefined
		this.states = {}
	}

	onThemeChange = (): void => {
		if (!this.monaco) {
			return
		}
		this.monaco.editor.setTheme(Theme.getEditorTheme())
	}

	updateState(): void {
		if (!this.editor || !this.editor.getModel()) {
			return
		}

		if (!this.curUuid) {
			this.curUuid = this.props.uuid
		}

		if (this.curUuid != this.props.uuid) {
			this.states[this.curUuid] = {
				model: this.editor.getModel(),
				view: this.editor.saveViewState(),
			}

			let newState = this.states[this.props.uuid]
			if (newState) {
				this.editor.setModel(newState.model)
				this.editor.restoreViewState(newState.view)
			} else {
				let model = this.monaco.editor.createModel(
					this.props.value, "markdown",
				)
				this.editor.setModel(model)
			}

			this.curUuid = this.props.uuid
		}
	}

	render(): JSX.Element {
		this.updateState()

		let expandLeft = this.props.expandLeft
		let expandRight = this.props.expandRight
		let markdown: JSX.Element
		let leftGroupStyle: React.CSSProperties = css.group

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

		let editor: JSX.Element;
		if (!this.props.readOnly) {
			editor = <MonacoEditor.Editor
				height="900px"
				width="100%"
				defaultLanguage="markdown"
				theme={Theme.getEditorTheme()}
				defaultValue={this.props.value}
				beforeMount={CompletionEngine.handleBeforeMount}
				onMount={(editor: Monaco.editor.IStandaloneCodeEditor,
						monaco: MonacoEditor.Monaco): void => {
					this.monaco = monaco
					this.editor = editor
					this.editor.onDidDispose((): void => {
						this.editor = undefined
						this.monaco = undefined
						this.states = {}
					})
					this.updateState()

					CompletionEngine.handleAfterMount(editor, monaco)
				}}
				options={{
					folding: false,
					fontSize: 12,
					fontFamily: Theme.monospaceFont,
					fontWeight: Theme.monospaceWeight,
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
		}

		return <div className="layout horizontal flex" style={css.editorBox}>
			<div
				ref={this.markdown}
				style={leftGroupStyle}
				hidden={expandRight}
			>
				{markdown}
			</div>
			<div
				style={expandRight ? css.group : css.groupSplit}
				hidden={expandLeft}
			>
				<div style={css.editor}>
					{editor}
				</div>
			</div>
		</div>;
	}
}

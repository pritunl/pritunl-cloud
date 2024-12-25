/// <reference path="../References.d.ts"/>
import * as React from "react"
import * as Theme from "../Theme"
import * as MonacoEditor from "@monaco-editor/react"
import * as Monaco from "monaco-editor"

interface Props {
	disabled?: boolean
	value?: string
	readOnly?: boolean
	mode?: string
	fontSize?: number
	height?: string
	width?: string
	interval?: number
	autoScroll?: boolean
	refresh?: () => Promise<string>
	onChange?: (value: string) => void
}

interface State {
}

const css = {
	editorBox: {
		margin: "10px 0",
	} as React.CSSProperties,
	editor: {
		margin: "11px 0 10px 0",
		borderRadius: "3px",
		overflow: "hidden",
		width: "100%",
	} as React.CSSProperties,
}

export default class Editor extends React.Component<Props, State> {
	editor: Monaco.editor.IStandaloneCodeEditor
	monaco: MonacoEditor.Monaco
	value: string
	interval: NodeJS.Timer;

	constructor(props: any, context: any) {
		super(props, context)
		this.state = {
		}
	}

	componentDidMount(): void {
		if (this.props.interval) {
			this.interval = setInterval(() => {
				this.props.refresh().then((val) => {
					this.update(val)
				})
			}, this.props.interval);
		}
		if (!this.props.value && this.props.refresh) {
			this.props.refresh().then((val) => {
				this.update(val)
			})
		}
	}

	componentWillUnmount(): void {
		if (this.interval) {
			clearInterval(this.interval)
		}
	}

	update(val: string): void {
		let curValue = this.value || this.props.value
		if (curValue === val) {
			return
		}
		this.value = val;

		const model = this.editor.getModel()
		if (model) {
			model.setValue(val)

			if (this.props.autoScroll) {
				const lineCount = model.getLineCount()
				this.editor.revealLine(lineCount)
				this.editor.setPosition({
					lineNumber: lineCount,
					column: model.getLineMaxColumn(lineCount),
				})
			}
		}
	}

	render(): JSX.Element {
		return <div className="layout horizontal flex" style={css.editorBox}>
			<div style={css.editor}>
				<MonacoEditor.Editor
					height={this.props.height}
					width={this.props.width}
					defaultLanguage="markdown"
					theme={Theme.getEditorTheme()}
					value={this.props.value}
					onMount={(editor: Monaco.editor.IStandaloneCodeEditor,
							monaco: MonacoEditor.Monaco): void => {
						this.monaco = monaco
						this.editor = editor
					}}
					options={{
						folding: false,
						fontSize: this.props.fontSize,
						fontFamily: Theme.monospaceFont,
						fontWeight: Theme.monospaceWeight,
						tabSize: 4,
						detectIndentation: false,
						readOnly: this.props.readOnly,
						//rulers: [80],
						scrollBeyondLastLine: false,
						minimap: {
							enabled: false,
						},
						wordWrap: "on",
					}}
					onChange={(val): void => {
						if (this.props.onChange) {
							this.props.onChange(val)
						}
					}}
				/>
			</div>
		</div>
	}
}

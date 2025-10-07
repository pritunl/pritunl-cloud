/// <reference path="../References.d.ts"/>
import * as React from "react"
import * as MiscUtils from '../utils/MiscUtils';
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
	style?: React.CSSProperties
	refresh?: (first: boolean) => Promise<string>
	onChange?: (value: string) => void
}

interface State {
}

const css = {
	editor: {
		margin: "0",
		borderRadius: "3px",
		width: "100%",
	} as React.CSSProperties,
}

export default class Editor extends React.Component<Props, State> {
	editor: Monaco.editor.IStandaloneCodeEditor
	monaco: MonacoEditor.Monaco
	value: string
	sync: MiscUtils.SyncInterval;

	constructor(props: any, context: any) {
		super(props, context)
		this.state = {
		}
	}

	componentDidMount(): void {
		if (this.props.interval) {
			this.sync = new MiscUtils.SyncInterval(
				() => this.props.refresh(false).then((val) => {
					this.update(val)
				}),
				this.props.interval,
			)
		}
		if (!this.props.value && this.props.refresh) {
			this.props.refresh(true).then((val) => {
				this.update(val)
			})
		}
	}

	componentWillUnmount(): void {
		this.sync?.stop()
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
		let style: React.CSSProperties
		if (this.props.style) {
			style = {
				...css.editor,
				...this.props.style,
			}
		} else {
			style = css.editor
		}

		return <div style={style}>
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
					fontSize: this.props.fontSize || 12,
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
	}
}

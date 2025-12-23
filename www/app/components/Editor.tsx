/// <reference path="../References.d.ts"/>
import * as React from "react"
import * as Blueprint from "@blueprintjs/core"
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
	paused: boolean
}

const css = {
	editor: {
		margin: "0",
		borderRadius: "3px",
		width: "100%",
	} as React.CSSProperties,
	card: {
		position: "absolute",
		top: "10px",
		right: "24px",
		zIndex: 100,
		opacity: 0.8,
	} as React.CSSProperties,
	cardBox: {
		position: "relative",
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
			paused: false,
		}
	}

	componentDidMount(): void {
		if (this.props.interval) {
			this.sync = new MiscUtils.SyncInterval(
				() => {
					if (!this.isScrolledToBottom()) {
						if (!this.state.paused) {
							this.setState({
								...this.state,
								paused: true,
							})
						}
						return Promise.resolve()
					}
					if (this.state.paused) {
						this.setState({
							...this.state,
							paused: false,
						})
					}
					return this.props.refresh(false).then((val) => {
						this.update(val)
					})
				},
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

	isScrolledToBottom(): boolean {
		if (!this.editor) {
			return false
		}

		const scrollTop = this.editor.getScrollTop()
		const scrollHeight = this.editor.getScrollHeight()
		const layoutInfo = this.editor.getLayoutInfo()
		const visibleHeight = layoutInfo.height

		const threshold = 10
		return scrollTop + visibleHeight >= scrollHeight - threshold
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
			{this.state.paused && <div style={css.cardBox}>
				<Blueprint.Tag style={css.card}>
					Refresh Paused While Scrolling
				</Blueprint.Tag>
			</div>}
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

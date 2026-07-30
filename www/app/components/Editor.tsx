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
	ansi?: boolean
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
		bottom: "10px",
		right: "24px",
		zIndex: 100,
		opacity: 0.7,
	} as React.CSSProperties,
	cardBox: {
		position: "relative",
	} as React.CSSProperties,
}

interface AnsiResult {
	text: string
	decorations: Monaco.editor.IModelDeltaDecoration[]
}

const ansiRe =
	/\x1b(?:\[[0-9;?]*[ -\/]*[@-~]|\][^\x07\x1b]*(?:\x07|\x1b\\)?|[@-Z\\-_])/g

const ansiClassMap: Record<string, string> = {
	"30": "ansi-black", "31": "ansi-red", "32": "ansi-green",
	"33": "ansi-yellow", "34": "ansi-blue", "35": "ansi-magenta",
	"36": "ansi-cyan", "37": "ansi-white",
	"90": "ansi-bright-black", "91": "ansi-bright-red",
	"92": "ansi-bright-green", "93": "ansi-bright-yellow",
	"94": "ansi-bright-blue", "95": "ansi-bright-magenta",
	"96": "ansi-bright-cyan", "97": "ansi-bright-white",
}

function parseAnsi(input: string): AnsiResult {
	input = input
		.replace(/\r\n/g, "\n")
		.replace(/\r/g, "\n")
		.replace(/[\x00-\x08\x0b\x0c\x0e-\x1a\x1c-\x1f\x7f]/g, "")

	let decorations: Monaco.editor.IModelDeltaDecoration[] = []
	let output = ""
	let line = 1
	let column = 1
	let color = ""
	let bold = false
	let spanStartLine = 1
	let spanStartCol = 1
	let lastIndex = 0

	const activeClass = (): string => {
		let classes: string[] = []
		if (color) {
			classes.push(color)
		}
		if (bold) {
			classes.push("ansi-bold")
		}
		return classes.join(" ")
	}

	const flush = (endLine: number, endCol: number): void => {
		const className = activeClass()
		if (className && (endLine > spanStartLine ||
				(endLine === spanStartLine && endCol > spanStartCol))) {
			decorations.push({
				range: {
					startLineNumber: spanStartLine,
					startColumn: spanStartCol,
					endLineNumber: endLine,
					endColumn: endCol,
				},
				options: {
					inlineClassName: className,
				},
			})
		}
	}

	const append = (text: string): void => {
		output += text
		for (let i = 0; i < text.length; i++) {
			if (text[i] === "\n") {
				line += 1
				column = 1
			} else {
				column += 1
			}
		}
	}

	let match: RegExpExecArray
	while ((match = ansiRe.exec(input)) !== null) {
		append(input.slice(lastIndex, match.index))
		lastIndex = ansiRe.lastIndex

		flush(line, column)

		const seq = match[0]
		if (seq.startsWith("\x1b[") && seq.endsWith("m")) {
			const params = seq.slice(2, -1)
			const codes = params === "" ? ["0"] : params.split(";")
			for (const codeStr of codes) {
				const code = parseInt(codeStr, 10)
				if (isNaN(code)) {
					continue
				}

				if (code === 0) {
					color = ""
					bold = false
				} else if (code === 1) {
					bold = true
				} else if (code === 22) {
					bold = false
				} else if (code === 39) {
					color = ""
				} else if ((code >= 30 && code <= 37) ||
						(code >= 90 && code <= 97)) {
					color = ansiClassMap[code.toString()]
				}
			}
		}

		spanStartLine = line
		spanStartCol = column
	}

	append(input.slice(lastIndex))
	flush(line, column)

	return {
		text: output,
		decorations: decorations,
	}
}

export default class Editor extends React.Component<Props, State> {
	editor: Monaco.editor.IStandaloneCodeEditor
	monaco: MonacoEditor.Monaco
	value: string
	decorations: Monaco.editor.IEditorDecorationsCollection
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

	refresh(): void {
		this.props.refresh(true).then((val) => {
			this.update(val)
		})
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
		if (!this.editor) {
			return
		}

		let curValue = this.value || this.props.value
		if (curValue === val) {
			return
		}
		this.value = val;

		const model = this.editor.getModel()
		if (model) {
			if (this.props.ansi) {
				const result = parseAnsi(val)

				model.setValue(result.text)

				if (this.decorations) {
					this.decorations.clear()
				}
				this.decorations = this.editor.createDecorationsCollection(
					result.decorations)
			} else {
				model.setValue(val)
			}

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
			{this.state.paused && <div style={css.cardBox}>
				<Blueprint.Tag style={css.card}>
					Refresh Paused While Scrolling
				</Blueprint.Tag>
			</div>}
		</div>
	}
}

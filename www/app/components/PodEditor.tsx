/// <reference path="../References.d.ts"/>
import * as React from "react"
import * as Styles from "../Styles"
import Help from "./Help"
import MarkdownMemo from "./MarkdownMemo"
import * as Theme from "../Theme"
import * as PodActions from "../actions/PodActions"
import * as CompletionEngine from "../completion/Engine"

import * as MonacoEditor from "@monaco-editor/react"
import * as Monaco from "monaco-editor"

interface Props {
	podId: string
	hidden: boolean
	readOnly: boolean
	expandLeft: boolean
	expandRight: boolean
	disabled?: boolean
	uuid: string
	value: string
	diffValue: string
	onEdit?: () => void
	onChange?: (value: string) => void
	onDiffChange?: (value: string) => void
}

interface State {
}

interface EditorState {
	model: Monaco.editor.ITextModel
	view: Monaco.editor.ICodeEditorViewState
}

const css = {
	group: {
		flex: 1,
		minWidth: "280px",
		height: "100%",
		overflowY: "auto",
		margin: "0",
		fontSize: "12px",
	} as React.CSSProperties,
	groupSpaced: {
		flex: 1,
		minWidth: "280px",
		height: "100%",
		overflowY: "auto",
		margin: "0",
		padding: "8px 0 0 0 ",
		fontSize: "12px",
	} as React.CSSProperties,
	groupSpacedExt: {
		flex: 1,
		minWidth: "280px",
		height: "100%",
		overflowY: "auto",
		margin: "0",
		padding: "0 0 0 0 ",
		fontSize: "12px",
	} as React.CSSProperties,
	groupSplit: {
		flex: 1,
		minWidth: "280px",
		height: "100%",
		overflowY: "auto",
		margin: "0 0 0 10px",
	} as React.CSSProperties,
	groupEditor: {
		flex: 1,
		minWidth: "280px",
		height: "100%",
		margin: "0",
		fontSize: "12px",
	} as React.CSSProperties,
	groupEditorSplit: {
		flex: 1,
		minWidth: "280px",
		height: "100%",
		margin: "0 0 0 10px",
	} as React.CSSProperties,
	groupEdit: {
		flex: 1,
		minWidth: "280px",
		height: "100%",
		margin: "0",
		fontSize: "12px",
	} as React.CSSProperties,
	groupEditSplit: {
		flex: 1,
		minWidth: "280px",
		height: "100%",
		margin: "0 0 0 10px",
	} as React.CSSProperties,
	editorBox: {
		flexGrow: 1,
		minHeight: 0,
		maxHeight: "100%",
		margin: "0",
	} as React.CSSProperties,
	editor: {
		margin: "0",
		borderRadius: "3px",
	} as React.CSSProperties,
	buttonEdit: {
		position: "absolute",
		top: "2px",
		right: "0px",
		padding: "7px",
	} as React.CSSProperties,
	buttonLeft: {
		position: "absolute",
		top: "-4px",
		right: "0px",
		padding: "7px",
	} as React.CSSProperties,
	buttonRight: {
		position: "absolute",
		top: "-4px",
		right: "0px",
		padding: "7px",
	} as React.CSSProperties,
};

const hashRe = /^( {0,3})#+\s+\S+/
const blockRe = /^( {4}|\s*`)/
const yamlBlockRe = /```yaml([\s\S]*?)```/g;
const kindRe = /kind:\s*(\w+)/
const markdownUri = Monaco.Uri.from({
	scheme: "file",
	path: "/markdown.md",
});
const kindsAll = ["instance", "domain", "firewall"]
const kindsUri: Record<string, Monaco.Uri> = {
	"instance": Monaco.Uri.from({
		scheme: "file",
		path: "/instance.yaml",
	}),
	"domain": Monaco.Uri.from({
		scheme: "file",
		path: "/domain.yaml",
	}),
	"firewall": Monaco.Uri.from({
		scheme: "file",
		path: "/firewall.yaml",
	}),
}
const pathsKind: Record<string, string> = {
	"/instance.yaml": "instance",
	"/domain.yaml": "domain",
	"/firewall.yaml": "firewall",
}

export default class PodEditor extends React.Component<Props, State> {
	curUuid: string
	editor: Monaco.editor.IStandaloneCodeEditor
	monaco: MonacoEditor.Monaco
	diffEditor: Monaco.editor.IStandaloneDiffEditor
	diffMonaco: MonacoEditor.Monaco
	states: Record<string, EditorState>
	markerListener: Monaco.IDisposable
	markersOffset: Record<string, number> = {}
	syncMarkersTimeout: NodeJS.Timeout

	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
		}

		this.states = {}
	}

	componentWillUnmount(): void {
		Theme.removeChangeListener(this.onThemeChange);
		this.curUuid = undefined
		this.editor = undefined
		this.monaco = undefined
		this.diffEditor = undefined
		this.diffMonaco = undefined
		this.states = {}
		if (this.markerListener) {
			this.markerListener.dispose()
		}
	}

	onThemeChange = (): void => {
		if (this.monaco) {
			this.monaco.editor.setTheme(Theme.getEditorTheme())
		}
		if (this.diffMonaco) {
			this.diffMonaco.editor.setTheme(Theme.getEditorTheme())
		}
	}

	syncMarkers(val: string): void {
    if (this.syncMarkersTimeout) {
      clearTimeout(this.syncMarkersTimeout)
    }

    this.syncMarkersTimeout = setTimeout(() => {
			this._syncMarkers(val)
      this.syncMarkersTimeout = null
    }, 200)
	}

	_syncMarkers(val: string): void {
		let monaco = this.monaco
		let editor = this.editor
		let kinds = new Set(kindsAll)

		if (!monaco?.editor) {
			return
		}

		let markdownModel = monaco.editor.getModel(markdownUri)
		if (markdownModel) {
			markdownModel.setValue(val)
		} else {
			markdownModel = monaco.editor.createModel(
				val,
				"markdown",
				markdownUri,
			)
		}

		const matches = [...val.matchAll(yamlBlockRe)]
		if (matches.length > 0) {
			let match = matches[0]

			const yamlContent = match[1]
			const yamlContentLines = yamlContent.split("\n")

			const baseLineOffset = markdownModel.getValue().substr(
				0, match.index).split("\n").length - 1

			const docLineOffsets: number[] = []
			const yamlDocuments: string[] = []

			if (!yamlContent.trimStart().startsWith("---")) {
				docLineOffsets.push(0)
				yamlDocuments.push(yamlContent)
			} else {
				let inDocument = false
				let currentDocLines: string[] = []
				yamlContentLines.forEach((line, lineIndex) => {
					if (line.trim() === "---") {
						if (inDocument && currentDocLines.length > 0) {
							yamlDocuments.push(currentDocLines.join("\n"))
							currentDocLines = []
						}

						inDocument = true
						docLineOffsets.push(lineIndex + 1)
					} else if (inDocument) {
						currentDocLines.push(line)
					}
				})

				if (inDocument && currentDocLines.length > 0) {
					yamlDocuments.push(currentDocLines.join("\n").trim())
				}
			}

			yamlDocuments.forEach((docContent, docIndex) => {
				let kind = docContent.match(kindRe)?.[1]
				if (!kinds.delete(kind)) {
					return
				}

				const modelUri = kindsUri[kind]
				let yamlModel = monaco.editor.getModel(modelUri)
				if (yamlModel) {
					yamlModel.setValue(docContent)
				} else {
					yamlModel = monaco.editor.createModel(
						docContent,
						"yaml",
						modelUri,
					)
				}

				this.markersOffset[kind] = baseLineOffset + docLineOffsets[docIndex]
			})

			for (const kind of kinds.keys()) {
				monaco.editor.setModelMarkers(
					editor.getModel(),
					`yaml-${kind}`,
					[],
				)
			}
		}
	}

	updateState(): void {
		if (!this.editor?.getModel()) {
			return
		}

		if (!this.curUuid) {
			this.curUuid = this.props.uuid
		}

		let model: Monaco.editor.ITextModel
		if (this.curUuid != this.props.uuid) {
			this.states[this.curUuid] = {
				model: this.editor.getModel(),
				view: this.editor.saveViewState(),
			}

			let newState = this.states[this.props.uuid]
			if (newState) {
				model = newState.model
				this.editor.setModel(newState.model)
				this.editor.restoreViewState(newState.view)
			} else {
				model = this.monaco.editor.createModel(
					this.props.value, "markdown",
				)
				this.editor.setModel(model)
			}

			this.curUuid = this.props.uuid
		} else {
			model = this.editor.getModel()
		}
	}

	render(): JSX.Element {
		this.updateState()

		if (this.props.hidden) {
			return <div></div>
		}

		let expandLeft = this.props.expandLeft
		let expandRight = this.props.expandRight
		let markdown: JSX.Element
		let leftGroupStyle: React.CSSProperties = css.group

		if (!expandRight) {
			markdown = <MarkdownMemo value={this.props.value}/>
		}

		let val = (this.props.value || "")
		let valTrim = val.trimStart()

		if (blockRe.test(val)) {
			leftGroupStyle = css.groupSpacedExt
		} else if (!hashRe.test(val)) {
			leftGroupStyle = css.groupSpaced
		} else {
			let valFirst = valTrim.split("\n")[0] || ""
			valFirst = valFirst.replace(/#/g, "").trim()
			if (!valFirst) {
				leftGroupStyle = css.groupSpacedExt
			}
		}

		let rightStyle: React.CSSProperties
		if (!this.props.readOnly) {
			rightStyle = expandRight ? css.groupEdit : css.groupEditSplit
		} else {
			rightStyle = expandRight ? css.groupEditor : css.groupEditorSplit
		}

		let editor: JSX.Element
		if (!this.props.readOnly && !this.props.diffValue) {
			editor = <MonacoEditor.Editor
				height="100%"
				width="100%"
				defaultLanguage="markdown"
				theme={Theme.getEditorTheme()}
				defaultValue={this.props.value}
				beforeMount={CompletionEngine.handleBeforeMount}
				onMount={(editor: Monaco.editor.IStandaloneCodeEditor,
						monaco: MonacoEditor.Monaco): void => {
					this.editor = editor
					this.monaco = monaco

					editor.addCommand(
						monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS,
						() => {},
					)

					if (this.markerListener) {
						this.markerListener.dispose()
					}
					this.markerListener = monaco.editor.onDidChangeMarkers((uris) => {
						uris.forEach((uri) => {
							let kind = pathsKind[uri.path]
							if (!kind) {
								return
							}

							const markers = monaco.editor.getModelMarkers({
								resource: uri,
							})

							const offset = this.markersOffset[kind] || 0
							const adjustedMarkers = markers.map(marker => ({
								...marker,
								startLineNumber: marker.startLineNumber + offset,
								endLineNumber: marker.endLineNumber + offset,
								resource: uri,
							}))

							monaco.editor.setModelMarkers(
								editor.getModel(),
								`yaml-${kind}`,
								adjustedMarkers
							)
						})
					})

					this.editor.onDidDispose((): void => {
						this.editor = undefined
						this.monaco = undefined
						this.states = {}
						this.curUuid = undefined
					})
					this.updateState()

					CompletionEngine.handleAfterMount(editor, monaco)

					setTimeout(() => {
						this.syncMarkers(val)
					}, 500)
				}}
				options={{
					folding: false,
					fontSize: 12,
					fontFamily: Theme.monospaceFont,
					fontWeight: Theme.monospaceWeight,
					automaticLayout: true,
					formatOnPaste: true,
					formatOnType: true,
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
					this.syncMarkers(val)
					this.props.onChange(val)
				}}
			/>
		} else if (!this.props.readOnly && this.props.diffValue) {
			editor = <MonacoEditor.DiffEditor
				height="100%"
				width="100%"
				theme={Theme.getEditorTheme()}
				original={this.props.diffValue}
				modified={this.props.value}
				originalLanguage="markdown"
				modifiedLanguage="markdown"
				beforeMount={CompletionEngine.handleBeforeMount}
				onMount={(editor: Monaco.editor.IStandaloneDiffEditor,
						monaco: MonacoEditor.Monaco): void => {
					this.diffEditor = editor
					this.diffMonaco = monaco
					this.diffEditor.onDidDispose((): void => {
						this.diffEditor = undefined
						this.diffMonaco = undefined
					})

					editor.addCommand(
						monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS,
						() => {},
					)

					let modifiedEditor = editor.getModifiedEditor()
					modifiedEditor.onDidChangeModelContent((): void => {
						this.props.onDiffChange(modifiedEditor.getValue())
					})

					this.updateState()
				}}
				options={{
					folding: false,
					fontSize: 12,
					fontFamily: Theme.monospaceFont,
					fontWeight: Theme.monospaceWeight,
					renderSideBySide: true,
					automaticLayout: true,
					formatOnPaste: true,
					formatOnType: true,
					rulers: [80],
					scrollBeyondLastLine: false,
					minimap: {
						enabled: false,
					},
					wordWrap: "on",
				}}
			/>
		}

		return <div className="layout horizontal" style={css.editorBox}>
			<div
				style={leftGroupStyle}
				hidden={expandRight}
			>
				{markdown}
			</div>
			<div
				style={expandRight ? css.groupEditor : css.groupEditorSplit}
				hidden={expandLeft}
			>
				<div style={rightStyle}>
					{editor}
				</div>
			</div>
		</div>
	}
}

/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Theme from '../Theme';

import * as MonacoEditor from "@monaco-editor/react"
import * as Monaco from "monaco-editor"

interface Props {
	original: string;
	modified: string;
	onStat?: (additions: number, deletions: number) => void;
}

interface State {
	height: number;
}

const lineHeight = 15;
const minHeight = 45;
const maxHeight = 500;

const options: Monaco.editor.IDiffEditorConstructionOptions = {
	folding: false,
	fontSize: 10,
	fontFamily: Theme.monospaceFont,
	fontWeight: Theme.monospaceWeight,
	renderSideBySide: false,
	ignoreTrimWhitespace: false,
	readOnly: true,
	automaticLayout: true,
	formatOnPaste: true,
	formatOnType: true,
	rulers: [80],
	scrollBeyondLastLine: false,
	scrollbar: {
		alwaysConsumeMouseWheel: false,
	},
	minimap: {
		enabled: false,
	},
	wordWrap: "on",
};

function clampHeight(height: number): number {
	return Math.max(minHeight, Math.min(maxHeight, Math.ceil(height)));
}

export default class PodDiff extends React.Component<Props, State> {
	editor: Monaco.editor.IStandaloneDiffEditor;
	listeners: Monaco.IDisposable[] = [];

	constructor(props: any, context: any) {
		super(props, context);

		let lines = Math.max(
			(props.original || "").split("\n").length,
			(props.modified || "").split("\n").length,
		);

		this.state = {
			height: clampHeight(lines * lineHeight),
		};
	}

	componentWillUnmount(): void {
		for (let listener of this.listeners) {
			listener.dispose();
		}
		this.listeners = [];

		let model = this.editor?.getModel();
		if (model) {
			this.editor.setModel(null);
			model.original.dispose();
			model.modified.dispose();
		}
		this.editor = undefined;
	}

	onMount = (editor: Monaco.editor.IStandaloneDiffEditor): void => {
		this.editor = editor;

		this.listeners.push(
			editor.getModifiedEditor().onDidContentSizeChange(this.syncHeight),
			editor.onDidUpdateDiff(this.onDiff),
		);

		this.onDiff();

		if (this.props.original && this.props.modified) {
			editor.revealFirstDiff();
		}
	}

	onDiff = (): void => {
		this.syncHeight();
		this.syncStat();
	}

	syncStat = (): void => {
		if (!this.editor || !this.props.onStat) {
			return;
		}

		let changes = this.editor.getLineChanges();
		if (!changes) {
			return;
		}

		let additions = 0;
		let deletions = 0;
		for (let change of changes) {
			if (change.modifiedEndLineNumber > 0) {
				additions += change.modifiedEndLineNumber -
					change.modifiedStartLineNumber + 1;
			}
			if (change.originalEndLineNumber > 0) {
				deletions += change.originalEndLineNumber -
					change.originalStartLineNumber + 1;
			}
		}

		if (!this.props.original) {
			deletions = 0;
		}
		if (!this.props.modified) {
			additions = 0;
		}

		this.props.onStat(additions, deletions);
	}

	syncHeight = (): void => {
		if (!this.editor) {
			return;
		}

		let modifiedEditor = this.editor.getModifiedEditor();
		let contentHeight = modifiedEditor.getContentHeight();
		if (!(modifiedEditor.getLayoutInfo().width > 0) ||
			!(contentHeight > 0)) {

			return;
		}

		let height = clampHeight(contentHeight);
		if (height !== this.state.height) {
			this.setState({
				...this.state,
				height: height,
			});
		}
	}

	render(): JSX.Element {
		return <MonacoEditor.DiffEditor
			height={this.state.height + "px"}
			width="100%"
			theme={Theme.getEditorTheme()}
			original={this.props.original}
			modified={this.props.modified}
			originalLanguage="markdown"
			modifiedLanguage="markdown"
			options={options}
			onMount={this.onMount}
		/>;
	}
}

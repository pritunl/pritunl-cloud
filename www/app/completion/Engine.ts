/// <reference path="../References.d.ts"/>
import * as Monaco from "monaco-editor"
import * as MonacoEditor from "@monaco-editor/react"
import CompletionCache from "./Cache"
import * as Types from "./Types"

let registered = false

export type Match = Monaco.languages.ProviderResult<
	Monaco.languages.CompletionList>

const noMatch: Match = {
	suggestions: []
}

export enum CompletionItemKind {
	Method = 0,
	Function = 1,
	Constructor = 2,
	Field = 3,
	Variable = 4,
	Class = 5,
	Struct = 6,
	Interface = 7,
	Module = 8,
	Property = 9,
	Event = 10,
	Operator = 11,
	Unit = 12,
	Value = 13,
	Constant = 14,
	Enum = 15,
	EnumMember = 16,
	Keyword = 17,
	Text = 18,
	Color = 19,
	File = 20,
	Reference = 21,
	Customcolor = 22,
	Folder = 23,
	TypeParameter = 24,
	User = 25,
	Issue = 26,
	Snippet = 27
}

export enum CompletionItemInsertTextRule {
	None = 0,
	/**
	 * Adjust whitespace/indentation of multiline insert texts to
	 * match the current line indentation.
	 */
	KeepWhitespace = 1,
	/**
	 * `insertText` is a snippet.
	 */
	InsertAsSnippet = 4
}

export function handleBeforeMount(
		monaco: MonacoEditor.Monaco): void {
}

export function handleAfterMount(
		editor: Monaco.editor.IStandaloneCodeEditor,
		monaco: MonacoEditor.Monaco): void {

	if (registered) {
		return
	}
	registered = true

	monaco.languages.registerHoverProvider("markdown", {
		provideHover: (model, position, token) => {
			const lineContent = model.getLineContent(position.lineNumber)

			const match = lineContent.match(
				/\+\/([a-zA-Z0-9-]*)\/([a-zA-Z0-9-]*)/)
			if (!match) {
				return null
			}

			let kindName = match[1]
			let resourceName = match[2]
			let kind = CompletionCache.kind(kindName)
			let resource = CompletionCache.resource(kindName, resourceName)

			if (kind && resource) {
				let contents = [
					{value: kind.title},
				]

				let data: string[] = []
				for (let item of resource.info) {
					data.push(item.label + ":  " + item.value)
				}

				contents.push({
					value: data.join("  \n"),
				})

				return {
					range: {
						startLineNumber: position.lineNumber,
						endLineNumber: position.lineNumber,
						startColumn: match.index + 1,
						endColumn: match.index + kindName.length + resourceName.length + 5,
					},
					contents: contents,
				}
			}

			return null
		}
	})

	editor.updateOptions({
    suggest: {
        preview: true,
    }
	});

	monaco.languages.setLanguageConfiguration("markdown", {
		wordPattern: /[^+\/:]+/g,
	});

	monaco.languages.registerCompletionItemProvider("markdown", {
		triggerCharacters: ["+", "/", ":"],
		provideCompletionItems: (model, position) => {
			const textBeforeCursor = model.getValueInRange({
				startLineNumber: position.lineNumber,
				startColumn: 1,
				endLineNumber: position.lineNumber,
				endColumn: position.column,
			})

			const selectorWithTagMatch = textBeforeCursor.match(
				/\+\/([a-zA-Z0-9-]*)\/([a-zA-Z0-9-]*):([a-zA-Z0-9-]*)\/$/);
			const selectorDirectMatch = textBeforeCursor.match(
				/\+\/([a-zA-Z0-9-]*)\/([a-zA-Z0-9-]*)\/$/);

			if (selectorWithTagMatch || selectorDirectMatch) {
				const match = selectorWithTagMatch || selectorDirectMatch;
				const kindName = match[1];
				const resourceName = match[2];

				let selectorKey = "";
				switch (kindName) {
					case "instance":
						selectorKey = "instance";
						break;
					case "vpc":
						selectorKey = "vpc";
						break;
					case "subnet":
						selectorKey = "subnet";
						break;
					case "certificate":
						selectorKey = "certificate";
						break;
					case "secret":
						selectorKey = "secret";
						break;
					case "unit":
						selectorKey = "unit";
						break;
					default:
						return noMatch;
				}

				const selectors = Types.Selectors[selectorKey];
				if (!selectors) {
					return noMatch;
				}

				const range = {
					startLineNumber: position.lineNumber,
					endLineNumber: position.lineNumber,
					startColumn: position.column,
					endColumn: position.column,
				}

				let suggestions: Monaco.languages.CompletionItem[] = [];

				for (const [key, info] of Object.entries(selectors)) {
					suggestions.push({
						label: key,
						kind: CompletionItemKind.Value,
						insertText: key,
						insertTextRules: CompletionItemInsertTextRule.InsertAsSnippet,
						documentation: info.tooltip,
						detail: info.tooltip,
						range: range,
					})
				}

				return {
					suggestions: suggestions,
				}
			}

			const tagMatch = textBeforeCursor.match(
				/\+\/([a-zA-Z0-9-]*)\/([a-zA-Z0-9-]*):$/);
			if (tagMatch) {
				let kindName = tagMatch[1]
				let resourceName = tagMatch[2]
				let resource = CompletionCache.resource(kindName, resourceName)
				if (!resource || !resource.tags) {
					return noMatch
				}

				const range = {
					startLineNumber: position.lineNumber,
					endLineNumber: position.lineNumber,
					startColumn: position.column,
					endColumn: position.column,
				}

				let suggestions: Monaco.languages.CompletionItem[] = []

				for (const tag of resource.tags) {
					suggestions.push({
						label: tag.name,
						kind: CompletionItemKind.Field,
						insertText: tag.name,
						insertTextRules: CompletionItemInsertTextRule.InsertAsSnippet,
						documentation: "Tag",
						range: range,
					})
				}

				return {
					suggestions: suggestions,
				}
			}

			const resourceMatch = textBeforeCursor.match(/\+\/([a-zA-Z0-9-]*)\/$/)
			if (resourceMatch) {
				let kind = CompletionCache.kind(resourceMatch[1])
				if (!kind) {
					return noMatch
				}

				const range = {
					startLineNumber: position.lineNumber,
					endLineNumber: position.lineNumber,
					startColumn: position.column,
					endColumn: position.column,
				}

				let suggestions: Monaco.languages.CompletionItem[] = []

				for (const resource of (CompletionCache.resources(kind.name))) {
					suggestions.push({
						label: resource.name,
						kind: CompletionItemKind.Property,
						insertText: resource.name,
						insertTextRules: CompletionItemInsertTextRule.InsertAsSnippet,
						documentation: kind.title,
						range: range,
					})
				}

				return {
					suggestions: suggestions,
				}
			}

			const kindMatch = textBeforeCursor.match(/\+\/$/)
			if (kindMatch) {
				const range = {
					startLineNumber: position.lineNumber,
					endLineNumber: position.lineNumber,
					startColumn: position.column,
					endColumn: position.column,
				}

				let suggestions: Monaco.languages.CompletionItem[] = []

				for (const kind of CompletionCache.kinds) {
					suggestions.push({
						label: kind.name,
						kind: CompletionItemKind.Class,
						insertText: kind.name,
						insertTextRules: CompletionItemInsertTextRule.InsertAsSnippet,
						documentation: kind.title,
						range: range,
					})
				}

				return {
					suggestions: suggestions,
				}
			}

			return noMatch
		},
	})
}

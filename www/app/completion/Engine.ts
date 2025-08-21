/// <reference path="../References.d.ts"/>
import * as Monaco from "monaco-editor"
import * as MonacoEditor from "@monaco-editor/react"
import * as MonacoYaml from "monaco-yaml"
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

	MonacoYaml.configureMonacoYaml(monaco, {
		enableSchemaRequest: false,
		schemas: [
			{
				fileMatch: ["instance.yaml"],
				schema: {
					type: "object",
					properties: {
						name: {
							type: "string",
							description: "Instance name",
						},
						kind: {
							type: "string",
							enum: ["instance"],
							description: "Resource kind",
						},
						count: {
							type: "integer",
							description: "Number of instances",
						},
						plan: {
							type: "string",
							description: "Instance plan",
						},
						zone: {
							type: "string",
							description: "Availability zone",
						},
						node: {
							type: "string",
							description: "Specific node for the instance",
						},
						shape: {
							type: "string",
							description: "Instance shape specification",
						},
						vpc: {
							type: "string",
							description: "VPC identifier",
						},
						subnet: {
							type: "string",
							description: "Subnet identifier",
						},
						roles: {
							type: "array",
							items: {
								type: "string",
							},
							description: "List of roles assigned to the instance",
						},
						processors: {
							type: "integer",
							description: "Number of processors allocated",
						},
						memory: {
							type: "integer",
							description: "Memory allocated in MB",
						},
						uefi: {
							type: "boolean",
							description: "Enable UEFI boot",
						},
						secureBoot: {
							type: "boolean",
							description: "Enable secure boot",
						},
						cloudType: {
							type: "string",
							description: "Cloud provider type",
						},
						tpm: {
							type: "boolean",
							description: "Enable Trusted Platform Module",
						},
						vnc: {
							type: "boolean",
							description: "Enable VNC access",
						},
						deleteProtection: {
							type: "boolean",
							description: "Enable deletion protection",
						},
						skipSourceDestCheck: {
							type: "boolean",
							description: "Skip source/destination check",
						},
						gui: {
							type: "boolean",
							description: "Desktop GUI",
						},
						hostAddress: {
							type: "boolean",
							description: "Allocate host address",
						},
						publicAddress: {
							type: "boolean",
							description: "Allocate public IPv4 address",
						},
						publicAddress6: {
							type: "boolean",
							description: "Allocate public IPv6 address",
						},
						dhcpServer: {
							type: "boolean",
							description: "Enable DHCP server",
						},
						image: {
							type: "string",
							description: "Image identifier",
						},
						mounts: {
							type: "array",
							items: {
								type: "object",
								properties: {
									path: {
										type: "string",
										description: "Mount path",
									},
									disks: {
										type: "array",
										items: {
											type: "string",
										},
										description: "List of disk identifiers",
									},
								},
								required: ["path", "disks"],
								description: "Disk mount configuration",
							},
							description: "Disk mounts",
						},
						nodePorts: {
							type: "array",
							items: {
								type: "object",
								properties: {
									protocol: {
										type: "string",
										enum: ["tcp", "udp"],
										description: "Network protocol",
									},
									externalPort: {
										type: "integer",
										minimum: 1,
										maximum: 65535,
										description: "External port number",
									},
									internalPort: {
										type: "integer",
										minimum: 1,
										maximum: 65535,
										description: "Internal port number",
									},
								},
								required: ["protocol", "externalPort", "internalPort"],
								description: "Node port mapping",
							},
							description: "Node port configurations",
						},
						certificates: {
							type: "array",
							items: {
								type: "string",
							},
							description: "List of certificate identifiers",
						},
						secrets: {
							type: "array",
							items: {
								type: "string",
							},
							description: "List of secret identifiers",
						},
						pods: {
							type: "array",
							items: {
								type: "string",
							},
							description: "List of pod identifiers",
						},
						diskSize: {
							type: "integer",
							description: "Size of disk in GB",
						},
					},
					required: ["name", "kind", "zone", "vpc", "subnet", "image"],
					description: "Instance configuration",
				},
				uri: "https://todo.pritunl.com/instance-schema.json",
			},
			{
				fileMatch: ["domain.yaml"],
				schema: {
					type: "object",
					properties: {
						name: {
							type: "string",
							description: "Domain name identifier",
						},
						kind: {
							type: "string",
							enum: ["domain"],
							description: "Resource kind",
						},
						records: {
							type: "array",
							items: {
								type: "object",
								properties: {
									name: {
										type: "string",
										description: "Record name (subdomain or label)",
									},
									domain: {
										type: "string",
										description: "Domain name for this record",
									},
									type: {
										type: "string",
										enum: [
											"host",
											"private",
											"private6",
											"public",
											"public6",
											"oracle_public",
											"oracle_public6",
											"oracle_private",
										],
										description: "Record type",
									},
								},
								required: ["name", "domain", "type"],
								description: "DNS record configuration",
							},
							description: "List of DNS records",
						},
					},
					required: ["name", "kind", "records"],
					description: "Domain and DNS records configuration",
				},
				uri: "https://todo.pritunl.com/domain-schema.json",
			},
			{
				fileMatch: ["firewall.yaml"],
				schema: {
					type: "object",
					required: ["name", "kind", "ingress"],
					properties: {
						name: {
							type: "string",
							description: "The name of the firewall rule",
						},
						kind: {
							type: "string",
							enum: ["firewall"],
							description: "Resource kind",
						},
						ingress: {
							type: "array",
							description: "Ingress rules for the firewall",
							items: {
								type: "object",
								required: ["protocol", "port", "source"],
								properties: {
									protocol: {
										type: "string",
										enum: ["all", "icmp", "tcp", "udp",
											"multicast", "broadcast"],
										description: "The protocol for this rule",
									},
									port: {
										type: ["number", "string"],
										minimum: 1,
										maximum: 65535,
										description: "Port number or range " +
											"(e.g. \"80\" or \"80-443\")",
									},
									source: {
										type: "array",
										description: "Source addresses or networks",
										items: {
											type: "string",
										},
									},
								},
							},
						},
					},
				},
				uri: "https://todo.pritunl.com/firewall-schema.json",
			},
		],
	});
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

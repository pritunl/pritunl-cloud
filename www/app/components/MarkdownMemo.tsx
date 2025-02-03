/// <reference path="../References.d.ts"/>
import * as React from "react"
import Markdown from 'react-markdown';
import hljs from "highlight.js/lib/core";

interface Props {
	value: string
}

const langRe = /^language-(.+)$/
const codeBlockRe = /^\{([^}]+)\}?$/;

function parseCodeBlockHeader(input: string): Record<string, string> {
  const attrs: Record<string, string> = {};

  const matches = input.match(codeBlockRe);
  if (!matches) {
    return attrs;
  }

  const attrPairs = matches[1].split(",");
  for (let pair of attrPairs) {
    pair = pair.trim();

    const keyValue = pair.split("=", 2);
    if (keyValue.length === 2) {
      const key = keyValue[0].trim();
      const value = keyValue[1].trim().replace(/^"|"$/g, "");
      attrs[key] = value;
    }
  }

	return attrs;
}

const MarkdownWrap = React.memo<Props>((props) => {
  return <Markdown
		children={props.value}
		components={{
			code(args) {
				let {children, className, node, ...rest} = args
				let match = (className || "").match(langRe)

				let phase = ""
				if (node && node.data) {
					let nodeData = node.data as any
					if (nodeData && nodeData.meta) {
						let metaAttrs = parseCodeBlockHeader(nodeData.meta)
						phase = metaAttrs["phase"]
					}
				}

				if (match && !hljs.getLanguage(match[1])) {
					className = "language-plaintext"
				}

				if (phase === "reboot") {
					className += " intent-secondary"
				} else if (phase === "reload") {
					className += " intent-primary"
				}

				const codeRef = React.useRef<HTMLElement>(null);
				React.useEffect(() => {
						if (codeRef.current) {
								hljs.highlightElement(codeRef.current);
						}
				}, [children]);

				let elem = <code ref={codeRef} {...rest} className={className}>
					{children}
				</code>

				return elem
			}
		}}
	/>
});

export default class MarkdownMemo extends React.Component<Props, {}> {
	constructor(props: any, context: any) {
		super(props, context)
		this.state = {
		}
	}

	render() {
		return <MarkdownWrap value={this.props.value}/>
	}
}

/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as PlanTypes from '../types/PlanTypes';
import * as Theme from "../Theme";
import * as MonacoEditor from "@monaco-editor/react"

interface Props {
	disabled?: boolean;
	statements: PlanTypes.Statement[];
	onChange: (statements: PlanTypes.Statement[]) => void;
}

const css = {
	group: {
		margin: '5px 0 20px 0',
	} as React.CSSProperties,
};

export default class PlanStatement extends React.Component<Props, {}> {
	onChange = (val: string): void => {
		let curStatements = this.props.statements
		let newStatements: PlanTypes.Statement[] = []
		let lines = val.split("\n")

		for (let i = 0; i < lines.length; i++) {
			let line = lines[i]
			if (!line) {
				continue
			}

			let newStatement: PlanTypes.Statement = {
				statement: line,
			}

			let curStatement = curStatements[i]
			if (curStatement) {
				newStatement.id = curStatement.id
			}

			newStatements.push(newStatement)
		}

		this.props.onChange(newStatements);
	}

	render(): JSX.Element {
		let statements = (this.props.statements || [])
		let statementsStr: string[] = []

		for (let statement of statements) {
			statementsStr.push(statement.statement)
		}
		let val = statementsStr.join("\n")

		return <div className="layout horizontal" style={css.group}>
			<MonacoEditor.Editor
				height="192px"
				width="100%"
				defaultLanguage="markdown"
				theme={Theme.getEditorTheme()}
				defaultValue={val}
				options={{
					folding: false,
					fontSize: 12,
					fontFamily: Theme.monospaceFont,
					fontWeight: Theme.monospaceWeight,
					tabSize: 4,
					detectIndentation: false,
					scrollBeyondLastLine: false,
					minimap: {
						enabled: false,
					},
					wordWrap: "on",
					automaticLayout: true,
				}}
				onChange={(val): void => {
					this.onChange(val)
				}}
			/>
		</div>;
	}
}

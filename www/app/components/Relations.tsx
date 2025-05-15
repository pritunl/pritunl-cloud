/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Blueprint from '@blueprintjs/core';
import * as RelationTypes from '../types/RelationTypes';
import * as RelationsActions from '../actions/RelationsActions';
import * as Alert from '../Alert';
import * as Theme from '../Theme';

interface State {
	data: RelationTypes.Relation;
	disabled: boolean;
}

interface Props {
	kind: string;
	id: string;
}

const css = {
	card: {
		display: 'table-row',
		width: '100%',
		padding: 0,
		boxShadow: 'none',
	} as React.CSSProperties,
	timestamp: {
		verticalAlign: 'top',
		display: 'table-cell',
		padding: '6px',
	} as React.CSSProperties,
	level: {
		verticalAlign: 'top',
		display: 'table-cell',
		padding: '6px',
	} as React.CSSProperties,
	message: {
		verticalAlign: 'top',
		display: 'table-cell',
		padding: '6px',
	} as React.CSSProperties,
	fields: {
		verticalAlign: 'top',
		display: 'table-cell',
		padding: '6px',
	} as React.CSSProperties,
	buttons: {
		verticalAlign: 'top',
		display: 'table-cell',
		padding: '0',
		width: '30px',
	} as React.CSSProperties,
	key: {
		fontWeight: 'bold',
	} as React.CSSProperties,
	value: {
	} as React.CSSProperties,
	dialog: {
		height: '500px',
		width: '90%',
		maxWidth: '700px',
	} as React.CSSProperties,
	dialogBody: {
		height: '100%',
	} as React.CSSProperties,
	textarea: {
		resize: 'none',
		fontSize: Theme.monospaceSize,
		fontFamily: Theme.monospaceFont,
		fontWeight: Theme.monospaceWeight,
		marginBottom: 0,
	} as React.CSSProperties,
}

export default class Relations extends React.Component<Props, State> {
	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			data: null,
			disabled: false,
		}
	}

	load = async () => {
		this.setState({
			...this.state,
			disabled: true,
		})

		let data: RelationTypes.Relation
		try {
			data = await RelationsActions.load(this.props.kind, this.props.id)
		} catch (error) {
			Alert.error('Failed to load relation');
		}

		this.setState({
			...this.state,
			disabled: false,
			data: data,
		})
	}

	render(): JSX.Element {
		let dialog: JSX.Element
		if (this.state.data) {
			dialog = <Blueprint.Dialog
				title="Resource Overview"
				style={css.dialog}
				isOpen={!!this.state.data}
				usePortal={true}
				portalContainer={document.body}
				onClose={(): void => {
					this.setState({
						...this.state,
						data: null,
					})
				}}
			>
				<textarea
					className="bp5-dialog-body bp5-input"
					style={css.textarea}
					autoCapitalize="off"
					spellCheck={false}
					readOnly={true}
					value={this.state.data?.data || ""}
				/>
				<div className="bp5-dialog-footer">
					<div className="bp5-dialog-footer-actions">
						<button
							className="bp5-button bp5-intent-danger"
							type="button"
							onClick={(): void => {
								this.setState({
									...this.state,
									data: null,
								})
							}}
						>Close</button>
					</div>
				</div>
			</Blueprint.Dialog>
		}

		return <div>
		<button
				className="bp5-button bp5-minimal bp5-icon-locate bp5-intent-primary"
				type="button"
				onClick={this.load}
			>Resource Overview</button>
			{dialog}
		</div>
	}
}

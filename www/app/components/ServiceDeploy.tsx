/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Blueprint from '@blueprintjs/core';
import * as Icons from '@blueprintjs/icons';
import * as ServiceTypes from '../types/ServiceTypes';
import * as ServiceActions from '../actions/ServiceActions';
import * as Alert from '../Alert';

interface Props {
	service: ServiceTypes.ServiceRo;
	unit: ServiceTypes.ServiceUnit;
}

interface State {
	dialog: boolean;
	disabled: boolean;
	specId: string;
	count: number;
}

const css = {
	dialog: {
		width: '340px',
		position: 'absolute',
	} as React.CSSProperties,
	label: {
		width: '100%',
		maxWidth: '220px',
		margin: '18px 0 0 0',
	} as React.CSSProperties,
	input: {
		width: '100%',
	} as React.CSSProperties,
};

export default class ServiceDeploy extends React.Component<Props, State> {
	interval: NodeJS.Timer;

	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			dialog: false,
			disabled: false,
			specId: "",
			count: 1,
		};
	}

	openDialog = (): void => {
		this.setState({
			...this.state,
			dialog: true,
		});
	}

	closeDialog = (): void => {
		this.setState({
			...this.state,
			dialog: false,
			specId: "",
			count: 1,
		});
	}

	onCreate = (): void => {
		this.setState({
			...this.state,
			disabled: true,
		});
		ServiceActions.deployUnit(
				this.props.service.id, this.props.unit.id,
				this.state.specId, this.state.count).then((): void => {

			Alert.success('Successfully created deployments');

			this.setState({
				...this.state,
				dialog: false,
				disabled: false,
				specId: "",
				count: 1,
			});
		}).catch((): void => {
			this.setState({
				...this.state,
				disabled: false,
			});
		});
	}

	renderDeploy(): JSX.Element {
		let dialogElem = <Blueprint.Dialog
			title="Create Deployment"
			style={css.dialog}
			isOpen={this.state.dialog}
			usePortal={true}
			portalContainer={document.body}
			onClose={this.closeDialog}
		>
			<div className="bp5-dialog-body">
				<label
					className="bp5-label"
					style={css.label}
				>
					Additional Deployments
					<Blueprint.NumericInput
						value={this.state.count}
						onValueChange={(val): void => {
							this.setState({
								...this.state,
								count: val,
							})
						}}
					/>
				</label>
			</div>
			<div className="bp5-dialog-footer">
				<div className="bp5-dialog-footer-actions">
					<button
						className="bp5-button"
						type="button"
						onClick={this.closeDialog}
					>Cancel</button>
					<button
						className="bp5-button"
						type="button"
						disabled={this.state.disabled}
						onClick={this.onCreate}
					>Create</button>
				</div>
			</div>
		</Blueprint.Dialog>

		return <div>
			<Blueprint.MenuItem
				key="menu-new-deployment"
				disabled={this.state.disabled}
				icon={<Icons.Plus/>}
				onClick={(evt): void => {
					evt.preventDefault()
					evt.stopPropagation()
					this.openDialog()
				}}
				text="New Deployment"
			/>
			{dialogElem}
		</div>
	}

	renderImage(): JSX.Element {
		let dialogElem = <Blueprint.Dialog
			title="Create Image"
			style={css.dialog}
			isOpen={this.state.dialog}
			usePortal={true}
			portalContainer={document.body}
			onClose={this.closeDialog}
		>
			<div className="bp5-dialog-footer">
				<div className="bp5-dialog-footer-actions">
					<button
						className="bp5-button"
						type="button"
						onClick={this.closeDialog}
					>Cancel</button>
					<button
						className="bp5-button"
						type="button"
						disabled={this.state.disabled}
						onClick={this.onCreate}
					>Create</button>
				</div>
			</div>
		</Blueprint.Dialog>

		return <div>
			<Blueprint.MenuItem
				key="menu-new-deployment"
				disabled={this.state.disabled}
				icon={<Icons.Plus/>}
				onClick={(evt): void => {
					evt.preventDefault()
					evt.stopPropagation()
					this.openDialog()
				}}
				text="New Image"
			/>
			{dialogElem}
		</div>
	}

	render(): JSX.Element {
		if (this.props.unit.kind === "image") {
			return this.renderImage()
		} else {
			return this.renderDeploy()
		}
	}
}

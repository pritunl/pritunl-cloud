/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Blueprint from '@blueprintjs/core';
import * as BpSelect from '@blueprintjs/select';
import * as Icons from '@blueprintjs/icons';
import * as ServiceTypes from '../types/ServiceTypes';
import * as ServiceActions from '../actions/ServiceActions';
import * as Alert from '../Alert';
import * as Theme from '../Theme';
import * as MiscUtils from '../utils/MiscUtils';

interface Props {
	service: ServiceTypes.ServiceRo;
	unit: ServiceTypes.ServiceUnit;
	commits: ServiceTypes.Commit[];
}

interface State {
	dialog: boolean;
	disabled: boolean;
	specId: string;
	deployCommit: ServiceTypes.Commit;
	count: number;
}

const css = {
	dialog: {
		width: '430px',
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
	settingsOpen: {
		marginLeft: '10px',
	} as React.CSSProperties,
	commit: {
		fontFamily: Theme.monospaceFont,
		fontWeight: Theme.monospaceWeight,
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
			deployCommit: null,
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

	filterCommit: BpSelect.ItemPredicate<ServiceTypes.Commit> = (
		query, commit, _index, exactMatch) => {

		if (exactMatch) {
			return commit.id == query
		} else {
			return commit.id.indexOf(query) !== -1
		}
	}

	renderCommit: BpSelect.ItemRenderer<ServiceTypes.Commit> = (commit,
		{handleClick, handleFocus, modifiers, query, index}): JSX.Element => {

		if (!modifiers.matchesPredicate) {
			return null;
		}

		let className = ""
		let selected = false
		if (this.state.deployCommit?.id == commit.id ||
			(!this.state.deployCommit && index === 0)) {
			// disabled = true
			className = "bp5-text-intent-primary bp5-intent-primary"
			selected = true
		}

		return <Blueprint.MenuItem
			key={"diff-" + commit.id}
			style={css.commit}
			selected={selected}
			disabled={this.state.disabled}
			roleStructure="listoption"
			icon={<Icons.GitCommit
				className={className}
			/>}
			onFocus={handleFocus}
			onClick={(evt): void => {
				evt.preventDefault()
				evt.stopPropagation()
				handleClick(evt)
			}}
			text={MiscUtils.highlightMatch(commit.id.substring(0, 12), query)}
			textClassName={className}
			labelElement={<span
				className={className}
			>{MiscUtils.formatDateLocal(commit.timestamp)}</span>}
		/>
	}

	renderImage(): JSX.Element {
		let commitsSelect: JSX.Element
		if (this.props.commits) {
			let deployCommit = this.state.deployCommit || this.props.commits[0]

			commitsSelect = <BpSelect.Select<ServiceTypes.Commit>
				items={this.props.commits}
				itemPredicate={this.filterCommit}
				itemRenderer={this.renderCommit}
				noResults={<Blueprint.MenuItem
					disabled={true}
					style={css.commit}
					text="No results."
					roleStructure="listoption"
				/>}
				onItemSelect={(commit) => {
					this.setState({
						...this.state,
						deployCommit: commit,
					})
				}}
			>
				<Blueprint.Button
					alignText="left"
					icon={<Icons.GitCommit/>}
					rightIcon={<Icons.CaretDown/>}
					text={deployCommit?.id.substring(0, 12) + " " +
						MiscUtils.formatDateLocal(deployCommit?.timestamp)}
				/>
			</BpSelect.Select>
		}

		let dialogElem = <Blueprint.Dialog
			title="Create Image"
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
					Image Commit
					{commitsSelect}
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

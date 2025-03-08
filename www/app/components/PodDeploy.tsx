/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Blueprint from '@blueprintjs/core';
import * as BpSelect from '@blueprintjs/select';
import * as Icons from '@blueprintjs/icons';
import * as PodTypes from '../types/PodTypes';
import * as PodActions from '../actions/PodActions';
import * as Alert from '../Alert';
import * as Theme from '../Theme';
import * as MiscUtils from '../utils/MiscUtils';

interface Props {
	pod: PodTypes.PodRo;
	unit: PodTypes.PodUnit;
	commits: PodTypes.Commit[];
}

interface State {
	dialog: boolean;
	disabled: boolean;
	specId: string;
	deployCommit: PodTypes.Commit;
	count: number;
}

const css = {
	dialog: {
		width: '430px',
		position: 'absolute',
	} as React.CSSProperties,
	label: {
		width: '100%',
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
	} as React.CSSProperties,
	muted: {
		opacity: 0.75,
	} as React.CSSProperties,
	commitButton: {
		marginTop: "5px",
		width: '400px',
		fontFamily: Theme.monospaceFont,
		fontWeight: Theme.monospaceWeight,
	} as React.CSSProperties,
	commitsMenu: {
		maxHeight: '400px',
		overflowY: "auto",
		fontFamily: Theme.monospaceFont,
		fontWeight: Theme.monospaceWeight,
	} as React.CSSProperties,
};

export default class PodDeploy extends React.Component<Props, State> {
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

		let deployCommit = this.state.deployCommit?.id ||
			this.props.commits?.[0]?.id

		PodActions.deployUnit(
				this.props.pod.id, this.props.unit.id,
				deployCommit, this.state.count).then((): void => {

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
		let commitSelect: JSX.Element
		if (this.props.commits) {
			let deployCommit = this.state.deployCommit || this.props.commits?.[0]
			let selectButtonClass = ""
			let selectLabelClass = ""
			let selectLabelStyle: React.CSSProperties
			if (deployCommit && deployCommit.id === this.props.commits?.[0]?.id) {
				selectButtonClass = "bp5-text-intent-success"
				selectLabelStyle = css.muted
			} else {
				selectLabelClass = "bp5-text-muted"
			}

			let commitMenuItems: JSX.Element[] = []
			this.props.commits.forEach((commit, index): void => {
				let className = ""
				let styles: React.CSSProperties
				let disabled = false
				let selected = false

				if (this.state.deployCommit?.id == commit.id ||
					(!this.state.deployCommit && index === 0)) {

					className = "bp5-text-intent-primary bp5-intent-primary"
					styles = css.muted
					selected = true
				} else if (index === 0) {
					className = "bp5-text-intent-success bp5-intent-success"
					styles = css.muted
				}

				commitMenuItems.push(<Blueprint.MenuItem
					key={"diff-" + commit.id}
					disabled={disabled || this.state.disabled}
					selected={selected}
					roleStructure="listoption"
					icon={<Icons.GitCommit
						className={className}
					/>}
					onClick={(): void => {
						this.setState({
							...this.state,
							deployCommit: commit,
						})
					}}
					text={commit.id.substring(0, 12)}
					textClassName={className}
					labelElement={<span
						className={className}
						style={styles}
					>{MiscUtils.formatDateLocal(commit.timestamp)}</span>}
				/>)
			})

			commitSelect = <Blueprint.Popover
				content={<div>
					<Blueprint.Menu style={css.commitsMenu}>
						{commitMenuItems}
					</Blueprint.Menu>
				</div>}
				placement="bottom"
			>
				<Blueprint.Button
					alignText="left"
					icon={<Icons.GitCommit/>}
					rightIcon={<Icons.CaretDown/>}
					style={css.commitButton}
					textClassName={selectButtonClass}
				>
					<span>{deployCommit?.id.substring(0, 12)}</span>
					<span
						className={selectLabelClass}
						style={selectLabelStyle}
					>
						{" " + MiscUtils.formatDateLocal(deployCommit?.timestamp)}
					</span>
				</Blueprint.Button>
			</Blueprint.Popover>
		}

		let dialogElem = <Blueprint.Dialog
			title="Create Image Build"
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
					Target Commit
				</label>
				<div
					onClick={(e) => {
						e.stopPropagation();
					}}
				>
					{commitSelect}
				</div>
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
				text="New Build"
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

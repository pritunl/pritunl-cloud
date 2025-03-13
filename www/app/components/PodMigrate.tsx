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
	disabled: boolean;
	pod: PodTypes.PodRo;
	unit: PodTypes.PodUnit;
	commits: PodTypes.Commit[];
	selectedDeployments: string[];
	onClear: () => void;
}

interface State {
	dialog: boolean;
	disabled: boolean;
	specId: string;
	migrateCommit: PodTypes.Commit;
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

export default class PodMigrate extends React.Component<Props, State> {
	interval: NodeJS.Timer;

	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			dialog: false,
			disabled: false,
			specId: "",
			migrateCommit: null,
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
		});
	}

	filterCommit: BpSelect.ItemPredicate<PodTypes.Commit> = (
		query, commit, _index, exactMatch) => {
		if (exactMatch) {
			return commit.id == query
		} else {
			return commit.id.indexOf(query) !== -1
		}
	}
	renderCommit: BpSelect.ItemRenderer<PodTypes.Commit> = (commit,
		{handleClick, handleFocus, modifiers, query, index}): JSX.Element => {
		if (!modifiers.matchesPredicate) {
			return null;
		}
		let className = ""
		let selected = false
		if (this.state.migrateCommit?.id == commit.id ||
			(!this.state.migrateCommit && index === 0)) {
			className = "bp5-text-intent-primary bp5-intent-primary"
			selected = true
		} else if (index === 0) {
			className = "bp5-text-intent-success bp5-intent-success"
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
	renderCommitSelect = () => {
		let commitSelect: JSX.Element
		if (this.props.commits) {
			let migrateCommit = this.state.migrateCommit || this.props.commits?.[0]
			let deployClass = ""
			if (migrateCommit && migrateCommit.id === this.props.commits?.[0]?.id) {
				deployClass = "bp5-text-intent-success"
			}

			commitSelect = <BpSelect.Select<PodTypes.Commit>
				items={this.props.commits}
				itemPredicate={this.filterCommit}
				itemRenderer={this.renderCommit}
				itemListRenderer={({items, itemsParentRef,
						query, renderItem, menuProps}) => {

					const renderedItems = items.map(renderItem).filter(
						item => item != null)
					return <Blueprint.Menu
						role="listbox"
						ulRef={itemsParentRef}
						{...menuProps}
					>
						<Blueprint.MenuItem
							disabled={true}
							text={`Found ${renderedItems.length} items matching "${query}"`}
							roleStructure="listoption"
						/>
						{renderedItems}
					</Blueprint.Menu>
				}}
				noResults={<Blueprint.MenuItem
					disabled={true}
					style={css.commit}
					text="No results."
					roleStructure="listoption"
				/>}
				onItemSelect={(commit) => {
					this.setState({
						...this.state,
						migrateCommit: commit,
					})
				}}
			>
				<Blueprint.Button
					alignText="left"
					icon={<Icons.GitCommit/>}
					rightIcon={<Icons.CaretDown/>}
					style={css.commitButton}
					textClassName={deployClass}
					text={migrateCommit?.id.substring(0, 12) + " " +
						MiscUtils.formatDateLocal(migrateCommit?.timestamp)}
				/>
			</BpSelect.Select>
		}
	}

	onMigrate = (): void => {
		let migrateCommit = this.state.migrateCommit || this.props.commits?.[0]

		this.setState({
			...this.state,
			disabled: true,
		});
		PodActions.updateMultiUnitAction(
				this.props.pod.id, this.props.unit.id,
				this.props.selectedDeployments,
				"migrate", migrateCommit.id).then((): void => {

			Alert.success('Successfully initiated deployment migration');

			this.setState({
				...this.state,
				dialog: false,
				disabled: false,
			});
			this.props.onClear();
		}).catch((): void => {
			this.setState({
				...this.state,
				disabled: false,
			});
		});
	}

	render(): JSX.Element {
		let commitSelect: JSX.Element
		if (this.props.commits) {
			let migrateCommit = this.state.migrateCommit || this.props.commits?.[0]
			let selectButtonClass = ""
			let selectLabelClass = ""
			let selectLabelStyle: React.CSSProperties
			if (migrateCommit && migrateCommit.id === this.props.commits?.[0]?.id) {
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

				if (this.state.migrateCommit?.id == commit.id ||
					(!this.state.migrateCommit && index === 0)) {

					className = "bp5-text-intent-primary bp5-intent-primary"
					styles = css.muted
					selected = true
				} else if (index === 0) {
					className = "bp5-text-intent-success bp5-intent-success"
					styles = css.muted
				}

				commitMenuItems.push(<Blueprint.MenuItem
					key={"diff-" + commit.id}
					disabled={disabled || this.props.disabled || this.state.disabled}
					selected={selected}
					roleStructure="listoption"
					icon={<Icons.GitCommit
						className={className}
					/>}
					onClick={(): void => {
						this.setState({
							...this.state,
							migrateCommit: commit,
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
					<span>{migrateCommit?.id.substring(0, 12)}</span>
					<span
						className={selectLabelClass}
						style={selectLabelStyle}
					>
						{" " + MiscUtils.formatDateLocal(migrateCommit?.timestamp)}
					</span>
				</Blueprint.Button>
			</Blueprint.Popover>
		}

		let dialogElem = <Blueprint.Dialog
			title="Migrate Selected Deployments"
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
						onClick={this.onMigrate}
					>Migrate</button>
				</div>
			</div>
		</Blueprint.Dialog>

		return <div>
			<Blueprint.MenuItem
				key="menu-new-deployment"
				className="bp5-intent-primary"
				disabled={this.state.disabled || this.props.disabled}
				icon={<Icons.Updated/>}
				onClick={(evt): void => {
					evt.preventDefault()
					evt.stopPropagation()
					this.openDialog()
				}}
				text="Migrate Selected"
			/>
			{dialogElem}
		</div>
	}
}

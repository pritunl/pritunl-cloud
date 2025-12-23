/// <reference path="../References.d.ts"/>
import * as React from 'react';
import EventDispatcher from '../dispatcher/EventDispatcher';
import * as Blueprint from '@blueprintjs/core';
import * as Icons from '@blueprintjs/icons';
import * as Constants from '../Constants';
import * as PodTypes from '../types/PodTypes';
import * as PodActions from '../actions/PodActions';
import PodsUnitStore from '../stores/PodsUnitStore';
import * as MiscUtils from '../utils/MiscUtils';
import * as Theme from '../Theme';
import * as Alert from '../Alert';
import PageInput from './PageInput';
import PageSelect from './PageSelect';
import PageInfo from './PageInfo';
import PageInputButton from './PageInputButton';
import PodEditor from './PodEditor';
import PodUnit from './PodUnit';
import PodDeploy from './PodDeploy';
import PodMigrate from './PodMigrate';
import NonState from './NonState';
import PageSave from './PageSave';
import ConfirmButton from './ConfirmButton';
import Help from './Help';
import PageTextArea from "./PageTextArea";

interface Props {
	pod: PodTypes.PodRo;
	podOrig: PodTypes.PodRo;
	disabled: boolean;
	unitChanged: boolean;
	mode: string;
	onMode: (mode: string) => void;
	onChangeCommit: (unitId: string, commit: string) => void;
	onEdit: (units: PodTypes.Unit[]) => void;
}

interface State {
	disabled: boolean;
	expandLeft: boolean;
	expandRight: boolean;
	activeUnitId: string;
	selectedDeployments: Selected;
	lastSelectedDeployment: string;
	unit: PodTypes.PodUnit;
	commits: PodTypes.CommitData;
	commitData: PodTypes.Commit
	diffCommitData: PodTypes.Commit
	diffCommit: PodTypes.Commit
	diffChanged: boolean
	viewCommit: PodTypes.Commit
}

interface Selected {
	[key: string]: boolean;
}

const css = {
	card: {
		padding: '0 0 10px 0',
		width: '100%',
		flexGrow: 1,
		minHeight: 0,
		maxHeight: '100%',
	} as React.CSSProperties,
	editButton: {
		margin: "0 0 0 3px",
		minHeight: "18px",
		minWidth: "18px",
		height: "18px",
		width: "18px",
		padding: "0 0 0 1px",
	} as React.CSSProperties,
	tab: {
		fontWeight: "bold",
		marginRight: "10px",
	} as React.CSSProperties,
	documentIcon: {
		margin: "2px 0 0 0",
		fontSize: "12px",
	} as React.CSSProperties,
	button: {
		height: '30px',
	} as React.CSSProperties,
	buttons: {
		cursor: 'pointer',
		position: 'absolute',
		top: 0,
		left: 0,
		right: 0,
		padding: '4px',
		height: '39px',
	} as React.CSSProperties,
	divider : {
		marginRight: "0",
	} as React.CSSProperties,
	item: {
		margin: '9px 5px 0 5px',
		minHeight: '20px',
	} as React.CSSProperties,
	itemsLabel: {
		display: 'block',
	} as React.CSSProperties,
	itemsAdd: {
		margin: '8px 0 15px 0',
	} as React.CSSProperties,
	group: {
		flex: 1,
		minWidth: '280px',
		margin: '0 10px',
	} as React.CSSProperties,
	save: {
		paddingBottom: '10px',
	} as React.CSSProperties,
	navbar: {
		height: '52px',
	} as React.CSSProperties,
	label: {
		width: '100%',
		maxWidth: '280px',
	} as React.CSSProperties,
	tabsBox: {
		overflowX: 'auto',
		overflowY: 'hidden',
		scrollbarWidth: 'thin',
	} as React.CSSProperties,
	navButtons: {
		height: '52px',
	} as React.CSSProperties,
	status: {
		margin: '6px 0 0 1px',
	} as React.CSSProperties,
	icon: {
		marginRight: '3px',
	} as React.CSSProperties,
	inputGroup: {
		width: '100%',
	} as React.CSSProperties,
	protocol: {
		flex: '0 1 auto',
	} as React.CSSProperties,
	port: {
		flex: '1',
	} as React.CSSProperties,
	select: {
		margin: '7px 0px 0px 6px',
		paddingTop: '3px',
	} as React.CSSProperties,
	role: {
		margin: '9px 5px 0 5px',
		minHeight: '20px',
	} as React.CSSProperties,
	rules: {
		marginBottom: '15px',
	} as React.CSSProperties,
	navButton: {
		marginLeft: '10px',
	} as React.CSSProperties,
	settingsOpen: {
		marginLeft: '10px',
	} as React.CSSProperties,
	settingsMenu: {
		maxHeight: '400px',
		overflowY: "auto",
	} as React.CSSProperties,
	muted: {
		opacity: 0.75,
	} as React.CSSProperties,
	commitsButton: {
		marginLeft: '10px',
		maxHeight: '400px',
		overflowY: "auto",
		fontFamily: Theme.monospaceFont,
		fontWeight: Theme.monospaceWeight,
	} as React.CSSProperties,
	commitsMenu: {
		maxHeight: '400px',
		overflowY: "auto",
		fontFamily: Theme.monospaceFont,
		fontWeight: Theme.monospaceWeight,
	} as React.CSSProperties,
	nonState: {
		padding: "80px 0",
	} as React.CSSProperties,
};

export default class PodWorkspace extends React.Component<Props, State> {
	sync: MiscUtils.SyncInterval;
	eventToken: string;

	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			disabled: false,
			expandLeft: null,
			expandRight: null,
			activeUnitId: "",
			selectedDeployments: {},
			lastSelectedDeployment: null,
			unit: null,
			commits: null,
			commitData: null,
			diffCommitData: null,
			diffCommit: null,
			diffChanged: false,
			viewCommit: null,
		};
	}

	componentDidMount(): void {
		PodsUnitStore.addChangeListener(this.onChange);
		let activeUnit = this.getActiveUnit()
		if (activeUnit && !activeUnit.new) {
			this.syncUnit(activeUnit.id)
		}

		this.sync = new MiscUtils.SyncInterval(
			async () => {
				let activeUnit = this.getActiveUnit()
				if (activeUnit && !activeUnit.new) {
					await PodActions.syncUnit(this.props.pod.id, activeUnit.id);
				}
			},
			3000,
		)

		this.eventToken = EventDispatcher.register((action: PodTypes.PodDispatch) => {
			switch (action.type) {
				case PodTypes.CHANGE:
					let activeUnit = this.getActiveUnit()
					if (activeUnit && !activeUnit.new) {
						this.syncUnit(activeUnit.id)
					} else {
						setTimeout(() => {
							let activeUnit = this.getActiveUnit()
							if (activeUnit && !activeUnit.new) {
								this.syncUnit(activeUnit.id)
							}
						}, 500)
					}
					break;
			}
		});
	}

	componentWillUnmount(): void {
		PodsUnitStore.removeChangeListener(this.onChange);
		this.sync?.stop()
		EventDispatcher.unregister(this.eventToken)
	}

	get selectedDeployments(): boolean {
		return !!Object.keys(this.state.selectedDeployments).length;
	}

	syncUnit = async (unitId: string, more?: boolean): Promise<void> => {
		PodActions.syncUnit(this.props.pod.id, unitId);

		let specsPage = 0
		if (more && this.state.commits?.unit == unitId) {
			specsPage += this.state.commits.page + 1
		}

		let commitData = await PodActions.syncSpecs(
			this.props.pod.id, unitId, specsPage)

		if (more && this.state.commits?.unit == unitId &&
			commitData.page > this.state.commits?.page) {

			commitData.specs = [
				...this.state.commits.specs,
				...commitData.specs,
			]
		}

		this.setState({
			...this.state,
			commits: commitData,
		})
	}

	syncCommit = async (unitId: string, specId: string): Promise<void> => {
		let spec = await PodActions.spec(
			this.props.pod.id, unitId, specId)

		this.setState({
			...this.state,
			commitData: spec,
		})
	}

	syncDiffCommit = async (unitId: string, specId: string): Promise<void> => {
		let spec = await PodActions.spec(
			this.props.pod.id, unitId, specId)

		this.setState({
			...this.state,
			diffCommitData: spec,
		})
	}

	onChange = (): void => {
		let unit: PodTypes.PodUnit
		let activeUnit = this.getActiveUnit()

		if (activeUnit && !activeUnit.new) {
			unit = PodsUnitStore.unit(activeUnit.id)
		} else {
			unit = null
		}

		let selectedDeployments: Selected = {};
		let curSelectedDeployments = this.state.selectedDeployments;

		if (activeUnit && unit) {
			let deployments = unit.deployments || []
			deployments.forEach((deployment: PodTypes.Deployment): void => {
				if (curSelectedDeployments[deployment.id]) {
					selectedDeployments[deployment.id] = true;
				}
			})
		}

		this.setState({
			...this.state,
			selectedDeployments: selectedDeployments,
			unit: unit,
		});
	}

	onArchiveDeployments = (): void => {
		let activeUnit = this.getActiveUnit()
		if (!activeUnit) {
			return
		}

		this.setState({
			...this.state,
			disabled: true,
		});
		PodActions.updateMultiUnitAction(
				this.props.pod.id, activeUnit.id,
				Object.keys(this.state.selectedDeployments),
			  "archive").then((): void => {

			Alert.success('Successfully archived deployments');

			this.setState({
				...this.state,
				selectedDeployments: {},
				disabled: false,
			});
		}).catch((): void => {
			this.setState({
				...this.state,
				disabled: false,
			});
		});
	}

	onRestoreDeployments = (): void => {
		let activeUnit = this.getActiveUnit()
		if (!activeUnit) {
			return
		}

		this.setState({
			...this.state,
			disabled: true,
		});
		PodActions.updateMultiUnitAction(
				this.props.pod.id, activeUnit.id,
				Object.keys(this.state.selectedDeployments),
			  "restore").then((): void => {

			Alert.success('Successfully restored deployments');

			this.setState({
				...this.state,
				selectedDeployments: {},
				disabled: false,
			});
		}).catch((): void => {
			this.setState({
				...this.state,
				disabled: false,
			});
		});
	}

	onDeleteDeployments = (): void => {
		let activeUnit = this.getActiveUnit()
		if (!activeUnit) {
			return
		}

		this.setState({
			...this.state,
			disabled: true,
		});
		PodActions.updateMultiUnitAction(
				this.props.pod.id, activeUnit.id,
				Object.keys(this.state.selectedDeployments),
			  "destroy").then((): void => {

			Alert.success('Successfully deleted deployments');

			this.setState({
				...this.state,
				selectedDeployments: {},
				disabled: false,
			});
		}).catch((): void => {
			this.setState({
				...this.state,
				disabled: false,
			});
		});
	}

	getActiveUnit = (): PodTypes.Unit => {
		let units = [
			...(this.props.pod.units || []),
		]

		let activeUnit = units.find(unit => unit.id === this.state.activeUnitId)
		if (!activeUnit) {
			for (let unit of units) {
				if (!unit.delete) {
					activeUnit = unit
					break
				}
			}
		}

		return activeUnit
	}

	getActiveUnitOrig = (): PodTypes.Unit => {
		let units = (this.props.podOrig.units || [])

		let activeUnit = units.find(unit => unit.id === this.state.activeUnitId)
		if (!activeUnit) {
			for (let unit of units) {
				if (!unit.delete) {
					activeUnit = unit
					break
				}
			}
		}

		return activeUnit
	}

	getActiveUnitIndex = (): number => {
		let units = [
			...(this.props.pod.units || []),
		]

		let activeIndex = units.findIndex(
			unit => unit.id === this.state.activeUnitId)
		if (activeIndex === -1) {
			for (let i = 0; i < units.length; i++) {
				if (!units[i].delete) {
					activeIndex = i
					break
				}
			}
		}

		return activeIndex
	}

	onEdit = (): void => {
		let units = [
			...(this.props.pod.units || []),
		]

		this.setState({
			...this.state,
			expandLeft: false,
			expandRight: true,
		})
		this.props.onEdit(units)
	}

	onView = (): void => {
		this.setState({
			...this.state,
			expandLeft: true,
			expandRight: false,
		})
		this.props.onMode("view")
	}

	onUnit = (): void => {
		this.setState({
			...this.state,
			expandLeft: true,
			expandRight: false,
		})
		this.props.onMode("unit")
	}

	onNew = (): void => {
		let units = [
			...(this.props.pod.units || []),
		]

		units.push({
			id: MiscUtils.objectId(),
			name: "new-unit",
			spec: "",
			new: true,
		})

		this.setState({
			...this.state,
			expandLeft: false,
			expandRight: true,
		})
		this.props.onEdit(units)
	}

	onDelete = (): void => {
		let units = [
			...(this.props.pod.units || []),
		]

		let index = this.getActiveUnitIndex()
		if (index !== -1) {
			let unit = units[index]
			units[index] = {
				id: unit.id,
				delete: true,
			}
		}

		this.setState({
			...this.state,
			activeUnitId: "",
		})
		this.props.onEdit(units)

		let activeUnit = this.getActiveUnit()
		if (activeUnit && !activeUnit.new) {
			PodActions.syncUnit(this.props.pod.id, activeUnit.id);
		}
	}

	onUnitEdit = (val: string): void => {
		let units = [
			...(this.props.pod.units || []),
		]

		let index = this.getActiveUnitIndex()
		if (index !== -1) {
			units[index] = {
				...units[index],
				spec: val,
			}
		}

		this.props.onEdit(units)
	}

	onUnitDeploy = (val: string): void => {
		let index = this.getActiveUnitIndex()
		if (index !== -1) {
			this.props.onChangeCommit(this.props.pod.units[index].id, val)
		}
	}

	onViewCommit = (unit: PodTypes.Unit, commit: PodTypes.Commit): void => {
		if (commit.index === unit.spec_index) {
			this.setState({
				...this.state,
				viewCommit: null,
			})
		} else {
			this.setState({
				...this.state,
				viewCommit: commit,
			})
		}
		this.syncCommit(unit.id, commit.id)
	}

	render(): JSX.Element {
		let units = [
			...(this.props.pod.units || []),
		]
		let activeUnit = this.getActiveUnit()
		let diffCommit: PodTypes.Commit
		if (this.state.diffCommitData?.id === this.state.diffCommit?.id) {
			diffCommit = this.state.diffCommitData
		}
		let mode = this.props.mode
		let noUnits = true
		for (let unit of units) {
			if (!unit.delete) {
				noUnits = false
				break
			}
		}
		if (noUnits) {
			mode = "view"
		}

		let expandLeft = this.state.expandLeft
		let expandRight = this.state.expandRight
		if (expandLeft === null) {
			if (mode === "edit") {
				expandLeft = false
				expandRight = true
			} else {
				expandLeft = true
				expandRight = false
			}
		}

		if (!this.props.unitChanged && mode !== "edit") {
			expandLeft = true
			expandRight = false
		}

		let expandIconClass: string
		if (!expandLeft && !expandRight) {
			expandIconClass = "bp5-button bp5-minimal bp5-icon-maximize"
		} else {
			expandIconClass = "bp5-button bp5-minimal bp5-icon-minimize"
		}

		let tabsElem: JSX.Element[] = []
		for (let i = 0; i < units.length; ++i) {
			let unit = units[i]
			if (unit.delete) {
				continue
			}

			tabsElem.push(<Blueprint.Tab id={unit.id} style={css.tab} key={unit.id}>
				{unit.name}
			</Blueprint.Tab>)
		}

		let menuItems: JSX.Element[] = []

		menuItems.push(<li
			key="menu-unit-header"
			className="bp5-menu-header"
		>
			<h6 className="bp5-heading">Units</h6>
		</li>)
		menuItems.push(<Blueprint.MenuDivider
			key="menu-unit-divider"
		/>)

		menuItems.push(<Blueprint.MenuItem
			key="menu-new-unit"
			className=""
			disabled={this.props.disabled || this.state.disabled}
			icon={<Icons.Plus/>}
			onClick={(): void => {
				this.onNew()
			}}
			text={"New Unit"}
		/>)

		menuItems.push(<ConfirmButton
			key="menu-delete-unit"
			safe={true}
			menuItem={true}
			className="bp5-intent-danger bp5-icon-trash"
			progressClassName="bp5-intent-danger"
			dialogClassName="bp5-intent-danger bp5-icon-delete"
			dialogLabel="Delete Unit"
			label="Delete Unit"
			confirmMsg="Permanently delete this unit"
			confirmInput={false}
			items={[activeUnit ? activeUnit.name : "null"]}
			hidden={!activeUnit}
			style={css.navButton}
			disabled={this.props.disabled || this.state.disabled}
			onConfirm={(): void => {
				this.onDelete()
			}}
		/>)

		let deploymentsLabel = "Deployments"
		if (activeUnit && activeUnit.kind === "image") {
			deploymentsLabel = "Builds"
		}

		if (!noUnits) {
			menuItems.push(<li
				key="menu-deployments-header"
				className="bp5-menu-header"
			>
				<h6 className="bp5-heading">{deploymentsLabel}</h6>
			</li>)
			menuItems.push(<Blueprint.MenuDivider
				key="menu-deployments-divider"
			/>)

			if (mode !== "unit") {
				menuItems.push(<Blueprint.MenuItem
					key="menu-deployments"
					className=""
					disabled={this.props.disabled || this.state.disabled}
					icon={<Icons.Dashboard/>}
					onClick={(): void => {
						this.onUnit()
					}}
					text={"View " + deploymentsLabel}
				/>)
			}
		}

		let commits: PodTypes.Commit[];
		if (this.state.commits?.unit === activeUnit?.id) {
			commits = this.state.commits?.specs
		}

		let commitMenu: JSX.Element
		if (mode === "unit") {
			if (commits && this.state.unit?.kind !== "image") {
				let activeUnitOrig = this.getActiveUnitOrig()
				let commitMenuItems: JSX.Element[] = []

				commits.forEach((commit): void => {
					let className = ""
					let disabled = false
					let selected = false
					if (activeUnit && activeUnitOrig.deploy_spec == commit.id) {
						// disabled = true
						className = "bp5-text-intent-primary bp5-intent-primary"
						selected = true
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
							this.onUnitDeploy(commit.id)
						}}
						text={commit.id.substring(12)}
						textClassName={className}
						labelElement={<span
							className={className}
						>{MiscUtils.formatDateLocal(commit.timestamp)}</span>}
					/>)
				})

				if (commits.length < this.state.commits?.count) {
					commitMenuItems.push(<Blueprint.MenuItem
						key={"diff-more"}
						disabled={this.props.disabled || this.state.disabled}
						roleStructure="listoption"
						icon={<Icons.BringData/>}
						onClick={(): void => {
							this.syncUnit(activeUnit.id, true)
						}}
						text="Load More..."
						textClassName="bp5-text-muted"
						shouldDismissPopover={false}
					/>)
				}

				commitMenu = <Blueprint.Popover
					content={<Blueprint.Menu style={css.commitsMenu}>
						{commitMenuItems}
					</Blueprint.Menu>}
					placement="bottom"
				>
					<Blueprint.Button
						alignText="left"
						icon={<Icons.GitRepo/>}
						rightIcon={<Icons.CaretDown/>}
						text="Deployment Commit"
						style={css.settingsOpen}
					/>
				</Blueprint.Popover>
			} else {
				commitMenu = <Blueprint.Popover
					content={<Blueprint.Menu style={css.commitsMenu}>
					</Blueprint.Menu>}
					placement="bottom"
				>
					<Blueprint.Button
						alignText="left"
						icon={<Icons.GitRepo/>}
						rightIcon={<Icons.CaretDown/>}
						text="Deployment Commit"
						style={css.settingsOpen}
						disabled={true}
					/>
				</Blueprint.Popover>
			}

			let selectedNames: string[] = [];
			for (let deploymentId of Object.keys(this.state.selectedDeployments)) {
				selectedNames.push(deploymentId)
			}
			menuItems.push(<PodDeploy
				key="menu-pod-deploy"
				pod={this.props.pod}
				unit={activeUnit}
				commits={commits}
			/>)
			if (this.state.unit?.kind !== "image") {
				menuItems.push(<ConfirmButton
					key="menu-restore-deployments"
					label="Restore Selected"
					className="bp5-intent-success bp5-icon-unarchive"
					safe={true}
					menuItem={true}
					style={css.navButton}
					confirmMsg="Restore the selected archived deployments"
					confirmInput={false}
					items={selectedNames}
					disabled={!this.selectedDeployments || this.state.disabled}
					onConfirm={this.onRestoreDeployments}
				/>)
				menuItems.push(<ConfirmButton
					key="menu-archive-deployments"
					label="Archive Selected"
					className="bp5-intent-warning bp5-icon-archive"
					safe={true}
					menuItem={true}
					style={css.navButton}
					confirmMsg="Archive the selected deployments, use restore to reactivate"
					confirmInput={false}
					items={selectedNames}
					disabled={!this.selectedDeployments || this.state.disabled}
					onConfirm={this.onArchiveDeployments}
				/>)
				menuItems.push(<PodMigrate
					key="menu-migrate"
					disabled={!this.selectedDeployments || this.state.disabled}
					pod={this.props.pod}
					unit={this.state.unit}
					commits={commits}
					selectedDeployments={this.state.selectedDeployments}
					onClear={(): void => {
						this.setState({
							...this.state,
							selectedDeployments: {},
							disabled: false,
						});
					}}
				/>)
			}
			menuItems.push(<ConfirmButton
				key="menu-delete-deployments"
				label="Delete Selected"
				className="bp5-intent-danger bp5-icon-delete"
				safe={true}
				menuItem={true}
				style={css.navButton}
				confirmMsg="Permanently delete the selected deployments"
				confirmInput={true}
				items={selectedNames}
				disabled={!this.selectedDeployments || this.state.disabled}
				onConfirm={this.onDeleteDeployments}
			/>)
		}

		let newUnit: JSX.Element
		if (noUnits) {
			newUnit = <Blueprint.Button
				alignText="left"
				icon={<Icons.Plus/>}
				text="New Unit"
				style={css.settingsOpen}
				onClick={() => {
					this.onNew()
				}}
			/>
		}

		if (!noUnits) {
			menuItems.push(<li
				key="menu-spec-header"
				className="bp5-menu-header"
			>
				<h6 className="bp5-heading">Specs</h6>
			</li>)
			menuItems.push(<Blueprint.MenuDivider
				key="menu-spec-divider"
			/>)
		}

		if (mode !== "view") {
			menuItems.push(<Blueprint.MenuItem
				key="menu-view-spec"
				className=""
				disabled={this.props.disabled || this.state.disabled}
				icon={<Icons.DocumentOpen/>}
				onClick={(): void => {
					this.onView()
				}}
				text={"View Spec"}
			/>)
		}

		if (mode !== "edit") {
			menuItems.push(<Blueprint.MenuItem
				key="menu-edit-spec"
				className=""
				disabled={this.props.disabled || this.state.disabled}
				hidden={!activeUnit}
				icon={<Icons.Edit/>}
				onClick={(): void => {
					this.onEdit()
				}}
				text={"Edit Spec"}
			/>)
		}

		if (mode === "edit") {
			if (this.state.unit && activeUnit &&
				this.state.unit.id === activeUnit.id && commits) {

				let commitMenuItems: JSX.Element[] = []

				commits.forEach((commit): void => {
					let className = ""
					let disabled = false
					if (activeUnit && activeUnit.last_spec == commit.id) {
						if (diffCommit) {
							className = "bp5-text-intent-success"
						} else {
							className = "bp5-text-intent-primary"
						}
					}

					if (diffCommit && diffCommit.id == commit.id) {
						className = "bp5-text-intent-danger"
						disabled = true
					}

					commitMenuItems.push(<Blueprint.MenuItem
						key={"diff-" + commit.id}
						disabled={disabled || this.props.disabled || this.state.disabled}
						icon={<Icons.GitCommit/>}
						onClick={(): void => {
							this.setState({
								...this.state,
								diffCommit: commit,
								diffChanged: false,
							})
							this.syncDiffCommit(activeUnit.id, commit.id)
						}}
						text={commit.id.substring(12)}
						textClassName={className}
						label={MiscUtils.formatDateLocal(commit.timestamp)}
					/>)
				})

				if (commits.length < this.state.commits?.count) {
					commitMenuItems.push(<Blueprint.MenuItem
						key={"diff-more"}
						disabled={this.props.disabled || this.state.disabled}
						roleStructure="listoption"
						icon={<Icons.BringData/>}
						onClick={(): void => {
							this.syncUnit(activeUnit.id, true)
						}}
						text="Load More..."
						textClassName="bp5-text-muted"
						shouldDismissPopover={false}
					/>)
				}

				commitMenu = <Blueprint.Popover
					content={<Blueprint.Menu style={css.commitsMenu}>
						{commitMenuItems}
					</Blueprint.Menu>}
					placement="bottom"
				>
					<Blueprint.Button
						alignText="left"
						icon={<Icons.GitRepo/>}
						rightIcon={<Icons.CaretDown/>}
						text="Diff"
						style={css.settingsOpen}
					/>
				</Blueprint.Popover>
			} else {
				commitMenu = <Blueprint.Popover
					content={<Blueprint.Menu style={css.commitsMenu}>
					</Blueprint.Menu>}
					placement="bottom"
				>
					<Blueprint.Button
						alignText="left"
						icon={<Icons.GitRepo/>}
						rightIcon={<Icons.CaretDown/>}
						text="Diff"
						style={css.settingsOpen}
						disabled={true}
					/>
				</Blueprint.Popover>
			}

			menuItems.push(<li
				key="menu-theme-header"
				className="bp5-menu-header"
			>
				<h6 className="bp5-heading">Editor Theme</h6>
			</li>)
			menuItems.push(<Blueprint.MenuDivider
				key="menu-theme-divider"
			/>)

			let curEditorTheme = Theme.getEditorTheme()
			for (let editorTheme in Theme.editorThemeNames) {
				let className = ""
				let themeName = Theme.editorThemeNames[editorTheme]

				if (editorTheme === curEditorTheme) {
					className = "bp5-intent-primary"
				}

				let menuItem = <Blueprint.MenuItem
					key={"menu-theme-" + editorTheme}
					className={className}
					icon={<Icons.Font/>}
					onClick={(): void => {
						Theme.setEditorTheme(editorTheme)
						Theme.save()
						this.setState({
							...this.state,
						})
					}}
					text={themeName}
				/>
				menuItems.push(menuItem)
			}
		}

		let editorVal = ""
		let viewLatestCommit = true
		if (activeUnit) {
			if (mode === "view" &&
				this.state.viewCommit?.unit === activeUnit.id) {
					if (this.state.commitData?.id == this.state.viewCommit.id) {
						editorVal = this.state.commitData.data
					}
					viewLatestCommit = false
			} else {
				editorVal = activeUnit.spec
			}
		}

		if (mode === "view") {
			if (commits) {
				let commitMenuItems: JSX.Element[] = []

				let selectButton = <Blueprint.Button
					alignText="left"
					icon={<Icons.GitRepo/>}
					rightIcon={<Icons.CaretDown/>}
					text="View Commit"
					style={css.settingsOpen}
				/>

				commits.forEach((commit, index): void => {
					let className = ""
					let selected = false
					if (viewLatestCommit && index === 0) {
						className = "bp5-text-intent-primary bp5-intent-primary"
						selected = true
					} else if (!viewLatestCommit && commit.unit === activeUnit?.id &&
						this.state.viewCommit?.id === commit.id) {

						className = "bp5-text-intent-primary bp5-intent-primary"
						selected = true

						selectButton = <Blueprint.Button
							alignText="left"
							icon={<Icons.GitCommit/>}
							rightIcon={<Icons.CaretDown/>}
							style={css.commitsButton}
						>
							<span>{this.state.viewCommit.id.substring(12)}</span>
							<span
								className="bp5-text-muted"
							>
								{" " + MiscUtils.formatDateLocal(
									this.state.viewCommit.timestamp)}
							</span>
						</Blueprint.Button>
					}

					commitMenuItems.push(<Blueprint.MenuItem
						key={"diff-" + commit.id}
						disabled={this.props.disabled || this.state.disabled}
						selected={selected}
						roleStructure="listoption"
						icon={<Icons.GitCommit
							className={className}
						/>}
						onClick={(): void => {
							this.onViewCommit(activeUnit, commit)
						}}
						text={commit.id.substring(12)}
						textClassName={className}
						labelElement={<span
							className={className}
						>{MiscUtils.formatDateLocal(commit.timestamp)}</span>}
					/>)
				})

				if (commits.length < this.state.commits?.count) {
					commitMenuItems.push(<Blueprint.MenuItem
						key={"diff-more"}
						disabled={this.props.disabled || this.state.disabled}
						roleStructure="listoption"
						icon={<Icons.BringData/>}
						onClick={(): void => {
							this.syncUnit(activeUnit.id, true)
						}}
						text="Load More..."
						textClassName="bp5-text-muted"
						shouldDismissPopover={false}
					/>)
				}

				commitMenu = <Blueprint.Popover
					content={<Blueprint.Menu style={css.commitsMenu}>
						{commitMenuItems}
					</Blueprint.Menu>}
					placement="bottom"
				>
					{selectButton}
				</Blueprint.Popover>
			} else {
				commitMenu = <Blueprint.Popover
					content={<Blueprint.Menu style={css.commitsMenu}>
					</Blueprint.Menu>}
					placement="bottom"
				>
					<Blueprint.Button
					alignText="left"
					icon={<Icons.GitRepo/>}
					rightIcon={<Icons.CaretDown/>}
					text="View Commit"
					style={css.settingsOpen}
					disabled={true}
				/>
				</Blueprint.Popover>
			}
		}

		let settingsMenu = <Blueprint.Menu style={css.settingsMenu}>
			{menuItems}
		</Blueprint.Menu>

		return <div
			className="layout vertical"
			style={css.card}
		>
			<Blueprint.Navbar className="layout horizontal" style={css.navbar}>
				<Blueprint.NavbarGroup
					className="flex thin-scroll"
					style={css.tabsBox}
					align={"left"}
				>
					<Blueprint.Tabs
						id={this.props.pod.id}
						selectedTabId={activeUnit ? activeUnit.id : null}
						fill={true}
						onChange={(newTabId): void => {
							let activeUnitId = newTabId.valueOf() as string
							let activeUnit = units.find(unit => unit.id === activeUnitId)

							this.setState({
								...this.state,
								activeUnitId: activeUnitId,
								diffCommit: null,
								diffChanged: false,
							})

							if (activeUnit && !activeUnit.new) {
								this.syncUnit(activeUnitId)
							}
						}}
					>
						{tabsElem}
					</Blueprint.Tabs>
				</Blueprint.NavbarGroup>
				<Blueprint.NavbarGroup style={css.navButtons} align={"right"}>
					<Blueprint.NavbarDivider
						style={css.divider}
					/>
					<button
						hidden={!(mode === "edit" && !this.state.diffCommit)}
						style={css.navButton}
						className={expandIconClass}
						onClick={(): void => {
							this.setState({
								...this.state,
								expandLeft: false,
								expandRight: !expandRight,
							})
						}}
					/>
					<button
						hidden={!(mode === "edit" && this.state.diffCommit)}
						style={css.navButton}
						className="bp5-button bp5-icon-cross bp5-intent-danger"
						onClick={(): void => {
							this.setState({
								...this.state,
								diffCommit: null,
								diffChanged: false,
							})
						}}
					>Close Diff</button>
					{newUnit}
					{commitMenu}
					<Blueprint.Popover
						content={settingsMenu}
						placement="bottom"
					>
						<Blueprint.Button
							alignText="left"
							icon={<Icons.Application/>}
							rightIcon={<Icons.CaretDown/>}
							text="Settings"
							style={css.settingsOpen}
						/>
					</Blueprint.Popover>
				</Blueprint.NavbarGroup>
			</Blueprint.Navbar>
			<div style={css.nonState} hidden={!noUnits}>
				<NonState
					hidden={!noUnits}
					noDelay={true}
					iconClass="bp5-icon-server"
					title="No units"
					description="Add a new unit to get started."
				/>
			</div>
			<PodEditor
				podId={this.props.pod.id}
				hidden={mode === "unit"}
				expandLeft={expandLeft}
				expandRight={expandRight}
				disabled={this.props.disabled || this.state.disabled}
				readOnly={mode === "view"}
				uuid={activeUnit ? activeUnit.id : null}
				value={editorVal}
				diffValue={diffCommit ? diffCommit.data : null}
				onChange={(val: string): void => {
					this.onUnitEdit(val)
				}}
				onDiffChange={(val: string): void => {
					if (this.state.diffCommit && !this.state.diffChanged) {
						this.setState({
							...this.state,
							diffChanged: true,
						})
					}
					this.onUnitEdit(val)
				}}
				onEdit={this.onEdit}
			/>
			<PodUnit
				hidden={mode !== "unit"}
				disabled={this.props.disabled || this.state.disabled}
				selected={this.state.selectedDeployments}
				lastSelected={this.state.lastSelectedDeployment}
				unit={this.state.unit}
				onSelect={(selected: Selected, lastSelected: string): void => {
					this.setState({
						...this.state,
						lastSelectedDeployment: lastSelected,
						selectedDeployments: selected,
					});
				}}
			/>
		</div>;
	}
}

/// <reference path="../References.d.ts"/>
import * as React from 'react';
import * as Blueprint from '@blueprintjs/core';
import * as Icons from '@blueprintjs/icons';
import * as Constants from '../Constants';
import * as ServiceTypes from '../types/ServiceTypes';
import * as ServiceActions from '../actions/ServiceActions';
import ServicesUnitStore from '../stores/ServicesUnitStore';
import * as MiscUtils from '../utils/MiscUtils';
import * as Theme from '../Theme';
import * as Alert from '../Alert';
import PageInput from './PageInput';
import PageSelect from './PageSelect';
import PageInfo from './PageInfo';
import PageInputButton from './PageInputButton';
import ServiceEditor from './ServiceEditor';
import ServiceUnit from './ServiceUnit';
import ServiceDeploy from './ServiceDeploy';
import PageSave from './PageSave';
import ConfirmButton from './ConfirmButton';
import Help from './Help';
import PageTextArea from "./PageTextArea";

interface Props {
	service: ServiceTypes.ServiceRo;
	disabled: boolean;
	unitChanged: boolean;
	mode: string;
	onMode: (mode: string) => void;
	onChange: (units: ServiceTypes.Unit[]) => void;
	onEdit: (units: ServiceTypes.Unit[]) => void;
}

interface State {
	disabled: boolean;
	expandLeft: boolean;
	expandRight: boolean;
	activeUnitId: string;
	selectedDeployments: Selected;
	lastSelectedDeployment: string;
	unit: ServiceTypes.ServiceUnit;
	diffCommit: ServiceTypes.Commit
}

interface Selected {
	[key: string]: boolean;
}

const css = {
	card: {
		padding: '10px 10px 0 10px',
		width: '100%',
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
		backgroundColor: 'rgba(0, 0, 0, 0.13)',
	} as React.CSSProperties,
	divider : {
		marginRight: "0",
	} as React.CSSProperties,
	item: {
		margin: '9px 5px 0 5px',
		height: '20px',
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
	label: {
		width: '100%',
		maxWidth: '280px',
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
		height: '20px',
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
	commitsMenu: {
		maxHeight: '400px',
		overflowY: "auto",
		fontFamily: Theme.monospaceFont,
		fontWeight: Theme.monospaceWeight,
	} as React.CSSProperties,
};

export default class ServiceWorkspace extends React.Component<Props, State> {
	interval: NodeJS.Timer;

	constructor(props: any, context: any) {
		super(props, context);
		this.state = {
			disabled: false,
			expandLeft: true,
			expandRight: false,
			activeUnitId: "",
			selectedDeployments: {},
			lastSelectedDeployment: null,
			unit: null,
			diffCommit: null,
		};
	}

	componentDidMount(): void {
		ServicesUnitStore.addChangeListener(this.onChange);
		let activeUnit = this.getActiveUnit()
		if (activeUnit && !activeUnit.new) {
			ServiceActions.syncUnit(this.props.service.id, activeUnit.id);
		}

		this.interval = setInterval(() => {
			let activeUnit = this.getActiveUnit()
			if (activeUnit && !activeUnit.new) {
				ServiceActions.syncUnit(this.props.service.id, activeUnit.id);
			}
		}, 3000);
	}

	componentWillUnmount(): void {
		ServicesUnitStore.removeChangeListener(this.onChange);
		clearInterval(this.interval);
	}

	get selectedDeployments(): boolean {
		return !!Object.keys(this.state.selectedDeployments).length;
	}

	onChange = (): void => {
		let unit: ServiceTypes.ServiceUnit
		let activeUnit = this.getActiveUnit()

		if (activeUnit && !activeUnit.new) {
			unit = ServicesUnitStore.unit(activeUnit.id)
		} else {
			unit = null
		}

		let selectedDeployments: Selected = {};
		let curSelectedDeployments = this.state.selectedDeployments;

		if (activeUnit) {
			let deployments = unit.deployments || []
			deployments.forEach((deployment: ServiceTypes.Deployment): void => {
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
		ServiceActions.updateMultiUnitState(
				this.props.service.id, activeUnit.id,
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
		ServiceActions.updateMultiUnitState(
				this.props.service.id, activeUnit.id,
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
		ServiceActions.updateMultiUnitState(
				this.props.service.id, activeUnit.id,
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

	getActiveUnit = (): ServiceTypes.Unit => {
		let units = [
			...(this.props.service.units || []),
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

	getActiveUnitIndex = (): number => {
		let units = [
			...(this.props.service.units || []),
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
			...(this.props.service.units || []),
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
			...(this.props.service.units || []),
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
			...(this.props.service.units || []),
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
			ServiceActions.syncUnit(this.props.service.id, activeUnit.id);
		}
	}

	onUnitEdit = (val: string): void => {
		let units = [
			...(this.props.service.units || []),
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
		let units = [
			...(this.props.service.units || []),
		]

		let index = this.getActiveUnitIndex()
		if (index !== -1) {
			units[index] = {
				...units[index],
				deploy_commit: val,
			}
		}

		this.props.onChange(units)
	}

	render(): JSX.Element {
		let units = [
			...(this.props.service.units || []),
		]
		let activeUnit = this.getActiveUnit()
		let diffCommit = this.state.diffCommit

		let expandLeft = this.state.expandLeft
		let expandRight = this.state.expandRight
		if (!this.props.unitChanged) {
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

		menuItems.push(<li
			key="menu-deployments-header"
			className="bp5-menu-header"
		>
			<h6 className="bp5-heading">Deployments</h6>
		</li>)
		menuItems.push(<Blueprint.MenuDivider
			key="menu-deployments-divider"
		/>)

		if (this.props.mode !== "unit") {
			menuItems.push(<Blueprint.MenuItem
				key="menu-deployments"
				className=""
				disabled={this.props.disabled || this.state.disabled}
				icon={<Icons.Dashboard/>}
				onClick={(): void => {
					this.onUnit()
				}}
				text={"View Deployments"}
			/>)
		}

		let commitMenu: JSX.Element
		if (this.props.mode === "unit") {
			if (this.state.unit && activeUnit &&
				this.state.unit.id === activeUnit.id &&
				this.state.unit.commits &&
				this.state.unit.commits.length > 0) {

				let commitMenuItems: JSX.Element[] = []

				this.state.unit.commits.forEach((commit): void => {
					let className = ""
					let disabled = false
					let selected = false
					if (activeUnit && activeUnit.deploy_commit == commit.id) {
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
						text={commit.id.substring(0, 12)}
						textClassName={className}
						labelElement={<span
							className={className}
						>{MiscUtils.formatDateLocal(commit.timestamp)}</span>}
					/>)
				})

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
						text="Default Commit"
						style={css.settingsOpen}
					/>
				</Blueprint.Popover>
			}

			let selectedNames: string[] = [];
			for (let deploymentId of Object.keys(this.state.selectedDeployments)) {
				selectedNames.push(deploymentId)
			}
			menuItems.push(<ServiceDeploy
				key="menu-service-deploy"
				service={this.props.service}
				unit={activeUnit}
			/>)
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

		menuItems.push(<li
			key="menu-spec-header"
			className="bp5-menu-header"
		>
			<h6 className="bp5-heading">Specs</h6>
		</li>)
		menuItems.push(<Blueprint.MenuDivider
			key="menu-spec-divider"
		/>)

		if (this.props.mode !== "view") {
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

		if (this.props.mode !== "edit") {
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

		if (this.props.mode === "edit") {
			if (this.state.unit &&
				activeUnit &&
				this.state.unit.id === activeUnit.id &&
				this.state.unit.commits &&
				this.state.unit.commits.length > 0) {

				let commitMenuItems: JSX.Element[] = []

				this.state.unit.commits.forEach((commit): void => {
					let className = ""
					let disabled = false
					if (activeUnit && activeUnit.last_commit == commit.id) {
						if (diffCommit) {
							className = "bp5-text-intent-success"
						} else {
							className = "bp5-text-intent-primary"
						}
						disabled = true
					} else if (diffCommit && diffCommit.id == commit.id) {
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
							})
						}}
						text={commit.id.substring(0, 12)}
						textClassName={className}
						label={MiscUtils.formatDateLocal(commit.timestamp)}
					/>)
				})

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

		let settingsMenu = <Blueprint.Menu style={css.settingsMenu}>
			{menuItems}
		</Blueprint.Menu>

		return <div
			style={css.card}
		>
			<Blueprint.Navbar>
				<Blueprint.NavbarGroup align={"left"}>
					<Blueprint.Tabs
						id={this.props.service.id}
						selectedTabId={activeUnit ? activeUnit.id : null}
						fill={true}
						onChange={(newTabId): void => {
							let activeUnitId = newTabId.valueOf() as string
							let activeUnit = units.find(unit => unit.id === activeUnitId)

							this.setState({
								...this.state,
								activeUnitId: activeUnitId,
								diffCommit: null,
							})

							if (activeUnit && !activeUnit.new) {
								ServiceActions.syncUnit(this.props.service.id, activeUnitId);
							}
						}}
					>
						{tabsElem}
					</Blueprint.Tabs>
				</Blueprint.NavbarGroup>
				<Blueprint.NavbarGroup align={"right"}>
					<Blueprint.NavbarDivider
						style={css.divider}
					/>
					<button
						hidden={!(this.props.mode === "edit" && !this.state.diffCommit)}
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
						hidden={!(this.props.mode === "edit" && this.state.diffCommit)}
						style={css.navButton}
						className="bp5-button bp5-icon-edit"
						onClick={(): void => {
							this.setState({
								...this.state,
								diffCommit: null,
							})
						}}
					>Apply Edit</button>
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
			<ServiceEditor
				hidden={this.props.mode === "unit"}
				expandLeft={expandLeft}
				expandRight={expandRight}
				disabled={this.props.disabled || this.state.disabled}
				readOnly={this.props.mode === "view"}
				uuid={activeUnit ? activeUnit.id : null}
				value={activeUnit ? activeUnit.spec : null}
				diffValue={diffCommit ? diffCommit.data : null}
				onChange={(val: string): void => {
					this.onUnitEdit(val)
				}}
				onEdit={this.onEdit}
			/>
			<ServiceUnit
				hidden={this.props.mode !== "unit"}
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
